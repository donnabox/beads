package graphmanaged

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"
	"unicode/utf8"
)

var (
	errUnsupported = errors.New("managed platform unsupported")
	errBudget      = errors.New("managed input budget exceeded")
	errInput       = errors.New("managed input refused")
	errProtocol    = errors.New("managed control protocol refused")
	errCleanup     = errors.New("managed cleanup incomplete")
)

const (
	maxCandidates       = 64
	maxEntries          = 4096
	maxDirectoryDepth   = 8
	maxComponents       = 256
	maxFiles            = 256
	maxFileBytes        = 1 << 20
	maxSourceBytes      = 32 << 20
	maxExecutableBytes  = 2 << 30
	maxNodes            = 32768
	maxJSONDepth        = 32
	maxStringBytes      = 256 << 10
	maxStringsBytes     = 4 << 20
	maxPathBytes        = 4096
	maxPathsBytes       = 1 << 20
	maxRetainedBytes    = 8 << 20
	maxEnvironment      = 64
	maxEnvironmentBytes = 16 << 10
	admissionDuration   = 10 * time.Second
)

type budget struct{ candidates, entries, files, sources, nodes, strings, paths, retained, environment, environmentBytes int64 }

func charge(used *int64, n, limit int64, unit string) error {
	if n < 0 || *used < 0 || n > limit || *used > limit-n {
		return fmt.Errorf("%w: %s", errBudget, unit)
	}
	*used += n
	return nil
}

// All paths use base64 native bytes; JSON decoding never replaces Unix names.
type sourceInput struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}
type databaseInput struct {
	Name   string `json:"name"`
	Root   string `json:"root"`
	Branch string `json:"default_branch"`
}
type directoryInput struct {
	Path    string   `json:"path"`
	Entries []string `json:"entries"`
}
type environmentInput struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
type description struct {
	Version          int                `json:"version"`
	Executable       string             `json:"executable"`
	ExecutableSHA256 string             `json:"executable_sha256"`
	ProfileSHA256    string             `json:"profile_sha256"`
	Cwd              string             `json:"cwd"`
	SecurityRoots    []string           `json:"security_roots"`
	Sources          []sourceInput      `json:"sources"`
	Databases        []databaseInput    `json:"databases"`
	Directories      []directoryInput   `json:"directories"`
	Environment      []environmentInput `json:"environment"`
}

func decodeDescription(raw []byte, b *budget) (description, error) {
	var d description
	if len(raw) > maxFileBytes {
		return d, fmt.Errorf("%w: description bytes", errBudget)
	}
	if err := validateJSON(raw, b); err != nil {
		return d, err
	}
	if err := validateDescriptionFields(raw); err != nil {
		return d, err
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&d); err != nil {
		return d, errInput
	}
	if d.Version != 1 || !digestValid(d.ExecutableSHA256) || !digestValid(d.ProfileSHA256) {
		return d, errInput
	}
	return d, nil
}

func digestValid(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, c := range s {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}

func nativePath(encoded string, b *budget) (string, error) {
	if base64.StdEncoding.DecodedLen(len(encoded)) > maxPathBytes+2 {
		return "", fmt.Errorf("%w: path bytes", errBudget)
	}
	raw, err := base64.StdEncoding.Strict().DecodeString(encoded)
	if err != nil || len(raw) == 0 || bytes.IndexByte(raw, 0) >= 0 {
		return "", errInput
	}
	if len(raw) > maxPathBytes {
		return "", fmt.Errorf("%w: path bytes", errBudget)
	}
	if err = charge(&b.paths, int64(len(raw)), maxPathsBytes, "path bytes"); err != nil {
		return "", err
	}
	if err = charge(&b.retained, int64(len(raw)), maxRetainedBytes, "retained bytes"); err != nil {
		return "", err
	}
	return string(raw), nil
}

func checkWork(ctx context.Context, deadline time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if !time.Now().Before(deadline) {
		return fmt.Errorf("%w: admission time", errBudget)
	}
	return nil
}

// scanJSON bounds decoded allocation before encoding/json sees input. Object
// names are decoded only after their own and cumulative string charges pass.
// It rejects duplicate names, invalid UTF-8, lone surrogates and trailing data.
type jsonScan struct {
	raw    []byte
	pos    int
	budget *budget
}

func validateJSON(raw []byte, b *budget) error {
	if !utf8.Valid(raw) {
		return errInput
	}
	s := jsonScan{raw: raw, budget: b}
	if err := s.value(0); err != nil {
		return err
	}
	s.space()
	if s.pos != len(raw) {
		return errInput
	}
	return nil
}
func (s *jsonScan) space() {
	for s.pos < len(s.raw) && strings.ContainsRune(" \r\n\t", rune(s.raw[s.pos])) {
		s.pos++
	}
}
func (s *jsonScan) value(depth int) error {
	s.space()
	if s.pos == len(s.raw) {
		return errInput
	}
	if err := charge(&s.budget.nodes, 1, maxNodes, "JSON nodes"); err != nil {
		return err
	}
	switch s.raw[s.pos] {
	case '{', '[':
		if depth >= maxJSONDepth {
			return fmt.Errorf("%w: JSON depth", errBudget)
		}
		object := s.raw[s.pos] == '{'
		s.pos++
		s.space()
		end := byte(']')
		if object {
			end = '}'
		}
		if s.pos < len(s.raw) && s.raw[s.pos] == end {
			s.pos++
			return nil
		}
		names := map[string]struct{}{}
		for {
			if object {
				s.space()
				start := s.pos
				if err := s.stringValue(); err != nil {
					return err
				}
				var name string
				if json.Unmarshal(s.raw[start:s.pos], &name) != nil {
					return errInput
				}
				if _, ok := names[name]; ok {
					return errInput
				}
				names[name] = struct{}{}
				s.space()
				if s.pos == len(s.raw) || s.raw[s.pos] != ':' {
					return errInput
				}
				s.pos++
			}
			if err := s.value(depth + 1); err != nil {
				return err
			}
			s.space()
			if s.pos == len(s.raw) {
				return errInput
			}
			c := s.raw[s.pos]
			s.pos++
			if c == end {
				return nil
			}
			if c != ',' {
				return errInput
			}
		}
	case '"':
		return s.stringValue()
	default:
		start := s.pos
		for s.pos < len(s.raw) && !strings.ContainsRune(" \r\n\t,]}", rune(s.raw[s.pos])) {
			s.pos++
		}
		if start == s.pos || !json.Valid(s.raw[start:s.pos]) {
			return errInput
		}
		return nil
	}
}
func (s *jsonScan) stringValue() error {
	if s.pos == len(s.raw) || s.raw[s.pos] != '"' {
		return errInput
	}
	s.pos++
	var size int64
	for s.pos < len(s.raw) {
		c := s.raw[s.pos]
		s.pos++
		if c == '"' {
			return charge(&s.budget.strings, size, maxStringsBytes, "decoded strings")
		}
		n := 1
		switch {
		case c < 0x20:
			return errInput
		case c == '\\':
			if s.pos == len(s.raw) {
				return errInput
			}
			e := s.raw[s.pos]
			s.pos++
			switch e {
			case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
			case 'u':
				r, err := s.hexRune()
				if err != nil {
					return err
				}
				if utf16.IsSurrogate(r) {
					if r < 0xD800 || r > 0xDBFF || s.pos+2 > len(s.raw) || string(s.raw[s.pos:s.pos+2]) != "\\u" {
						return errInput
					}
					s.pos += 2
					low, err := s.hexRune()
					if err != nil || low < 0xDC00 || low > 0xDFFF {
						return errInput
					}
					r = utf16.DecodeRune(r, low)
				}
				n = utf8.RuneLen(r)
			default:
				return errInput
			}
		case c >= utf8.RuneSelf:
			_, n = utf8.DecodeRune(s.raw[s.pos-1:])
			s.pos += n - 1
		}
		if err := charge(&size, int64(n), maxStringBytes, "decoded string"); err != nil {
			return err
		}
	}
	return errInput
}
func (s *jsonScan) hexRune() (rune, error) {
	if s.pos+4 > len(s.raw) {
		return 0, errInput
	}
	n, err := strconv.ParseUint(string(s.raw[s.pos:s.pos+4]), 16, 16)
	if err != nil {
		return 0, errInput
	}
	s.pos += 4
	return rune(n), nil
}

// encoding/json accepts case-insensitive struct names. The wire description
// does not: check every object against exact tagged names before typed decode.
// The lexical prepass has already bounded all payloads and rejected duplicates.
func exactObject(raw []byte, allowed ...string) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil || fields == nil {
		return nil, errInput
	}
	for name := range fields {
		found := false
		for _, key := range allowed {
			if name == key {
				found = true
				break
			}
		}
		if !found {
			return nil, errInput
		}
	}
	return fields, nil
}
func validateDescriptionFields(raw []byte) error {
	fields, err := exactObject(raw, "version", "executable", "executable_sha256", "profile_sha256", "cwd", "security_roots", "sources", "databases", "directories", "environment")
	if err != nil {
		return err
	}
	for _, shape := range []struct {
		name   string
		fields []string
	}{
		{"sources", []string{"path", "sha256"}},
		{"databases", []string{"name", "root", "default_branch"}},
		{"directories", []string{"path", "entries"}},
		{"environment", []string{"name", "value"}},
	} {
		value, ok := fields[shape.name]
		if !ok {
			continue
		}
		var entries []json.RawMessage
		if err := json.Unmarshal(value, &entries); err != nil {
			return errInput
		}
		for _, entry := range entries {
			if _, err := exactObject(entry, shape.fields...); err != nil {
				return err
			}
		}
	}
	return nil
}
