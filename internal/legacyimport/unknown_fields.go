package legacyimport

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/steveyegge/beads/internal/types"
)

// Options applies only to ordinary export conversion, never graph packages.
type Options struct {
	IgnoreUnknownFields bool
}

// IgnoredField names a typed legacy path, without record IDs or source values.
// Path is a JSON array of escaped components; [] denotes any array element.
type IgnoredField struct {
	Path       string
	Count      int
	Truncated  bool
	PathDigest string // Stable digest of the full path, present only after truncation.
}

// Report bounds displayed paths while accounting for every discarded field.
// Discarding an unknown object counts its field once, not its opaque children.
type Report struct {
	Fields             []IgnoredField
	Total              int
	Distinct           int
	OmittedDistinct    int
	OmittedOccurrences int
}

const (
	maxReportedPaths      = 32
	maxPathComponentBytes = 64
)

type ignoredFields map[string]int

func (fields ignoredFields) add(path []string) {
	encoded, _ := json.Marshal(path) // Strings from validated JSON are always encodable.
	fields[string(encoded)]++
}

func (fields ignoredFields) report() Report {
	report := Report{Distinct: len(fields)}
	paths := make([]string, 0, len(fields))
	for path, count := range fields {
		paths = append(paths, path)
		report.Total += count
	}
	sort.Strings(paths)
	for i, path := range paths {
		count := fields[path]
		if i >= maxReportedPaths {
			report.OmittedDistinct++
			report.OmittedOccurrences += count
			continue
		}
		var components []string
		_ = json.Unmarshal([]byte(path), &components) // add produced this JSON array.
		truncated := false
		for j, component := range components {
			if len(component) > maxPathComponentBytes {
				end := maxPathComponentBytes
				for !utf8.RuneStart(component[end]) {
					end--
				}
				components[j], truncated = component[:end], true
			}
		}
		entry := IgnoredField{Path: asciiJSONPath(components), Count: count, Truncated: truncated}
		if truncated {
			digest := sha256.Sum256([]byte(path))
			entry.PathDigest = fmt.Sprintf("%x", digest[:8])
		}
		report.Fields = append(report.Fields, entry)
	}
	return report
}

// Only typed struct fields have an unknown vocabulary. Maps, RawMessage and
// arbitrary metadata/property dictionaries remain untouched. Validation of the
// complete original JSON precedes this walk; strictDecode still checks all known
// values after it, including custom unmarshallers and legacy aliases.
func discardUnknownFields(raw []byte, typ reflect.Type, path []string, ignored ignoredFields) ([]byte, error) {
	raw = bytes.TrimSpace(raw)
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	if typ == reflect.TypeOf(json.RawMessage{}) || bytes.Equal(raw, []byte("null")) {
		return raw, nil
	}
	if typ.Kind() == reflect.Struct && len(raw) != 0 && raw[0] == '{' {
		known := map[string]reflect.Type{}
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			name := strings.Split(field.Tag.Get("json"), ",")[0]
			if field.PkgPath == "" && name != "" && name != "-" {
				known[name] = field.Type
			}
		}
		if typ == reflect.TypeOf(types.BondRef{}) {
			known["proto_id"] = reflect.TypeOf("")
		}
		var values map[string]json.RawMessage
		if err := json.Unmarshal(raw, &values); err != nil {
			return nil, err
		}
		keys := make([]string, 0, len(values))
		for key := range values {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			child, ok := known[key]
			if !ok {
				for name := range known {
					if strings.EqualFold(name, key) {
						return nil, fmt.Errorf("legacy fields require exact case-sensitive names")
					}
				}
				if reservedField(key) {
					return nil, fmt.Errorf("graph or provenance fields are unavailable in legacy conversion")
				}
				ignored.add(appendPath(path, key))
				delete(values, key)
				continue
			}
			value, err := discardUnknownFields(values[key], child, appendPath(path, key), ignored)
			if err != nil {
				return nil, err
			}
			values[key] = value
		}
		return json.Marshal(values)
	}
	if (typ.Kind() == reflect.Slice || typ.Kind() == reflect.Array) && len(raw) != 0 && raw[0] == '[' {
		var values []json.RawMessage
		if err := json.Unmarshal(raw, &values); err != nil {
			return nil, err
		}
		for i, value := range values {
			clean, err := discardUnknownFields(value, typ.Elem(), appendPath(path, "[]"), ignored)
			if err != nil {
				return nil, err
			}
			values[i] = clean
		}
		return json.Marshal(values)
	}
	return raw, nil
}

func appendPath(path []string, component string) []string {
	return append(append([]string(nil), path...), component)
}

// This is a refusal vocabulary, not a graph wire schema. Never erase graph
// identity, retained data or provenance to reinterpret a Resource as an Issue.
// Underscore-prefixed names are record/header discriminators and provenance.
func reservedField(name string) bool {
	if strings.HasPrefix(name, "_") {
		return true
	}
	switch strings.ToLower(name) {
	case "type", "revision", "version", "properties", "owned", "attribution", "source", "target",
		"kind", "resource", "resources", "record", "records", "manifest", "format", "format_version",
		"scope", "scope_url", "package_id", "provenance", "history", "versions", "types", "links",
		"schemaversion", "preview", "result", "code", "retryable",
		"wisp", "wisp_plane", "dependency_count", "dependent_count", "comment_count", "comments_omitted":
		return true
	}
	return false
}

// Marshal first, then escape non-ASCII code points without re-quoting existing
// JSON escapes. This preserves exact string round-trips while making C1, DEL,
// bidi controls and all other non-ASCII text inert on a terminal. Go's quoted
// string syntax alone is not JSON (for example, \x7f and \U0001f600).
func asciiJSONPath(components []string) string {
	encoded, _ := json.Marshal(components)
	var out strings.Builder
	for _, r := range string(encoded) {
		if r < 0x7f {
			out.WriteRune(r)
		} else if r <= 0xffff {
			fmt.Fprintf(&out, "\\u%04x", r)
		} else {
			high, low := utf16.EncodeRune(r)
			fmt.Fprintf(&out, "\\u%04x\\u%04x", high, low)
		}
	}
	return out.String()
}
