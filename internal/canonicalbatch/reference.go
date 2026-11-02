package canonicalbatch

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/steveyegge/beads/internal/clirevision"
)

// ParseReference uses the existing CLI boundary (graph_preview_reference.go):
// unwrapped URLs are literal, local IDs use ID@=TOKEN, and absolute URL pins
// use [URL]@=TOKEN. Ordinal/current selectors require live acquisition and are
// unsupported by this offline compiler. Colon-only local IDs stay local.
func ParseReference(input, scope string) (Reference, error) {
	if input == "" || !utf8.ValidString(input) {
		return Reference{}, invalid("Reference must be nonempty UTF-8")
	}
	if strings.HasPrefix(input, "batch-ref::") {
		return Reference{}, invalid("Type/Reference input does not admit a batch binding")
	}
	identity, suffix := input, ""
	if input[0] == '[' {
		depth, end := 0, -1
		for i, c := range input {
			if c == '[' {
				depth++
			}
			if c == ']' {
				depth--
				if depth == 0 {
					end = i
					break
				}
				if depth < 0 {
					return Reference{}, invalid("unbalanced URL wrapper")
				}
			}
		}
		if end < 0 || end+2 >= len(input) || input[end+1] != '@' {
			return Reference{}, invalid("wrapped URL requires @revision")
		}
		identity, suffix = input[1:end], input[end+2:]
		parsed, err := url.Parse(identity)
		if err != nil || !strings.Contains(identity, "://") || parsed.Host == "" || parsed.Opaque != "" || (!strings.EqualFold(parsed.Scheme, "http") && !strings.EqualFold(parsed.Scheme, "https")) || strings.ContainsAny(parsed.EscapedPath(), "[]") || strings.ContainsAny(parsed.RawQuery, "[]") {
			return Reference{}, invalid("invalid bracketed HTTP(S) URL")
		}
	} else if !strings.Contains(input, "://") {
		if at := strings.IndexByte(input, '@'); at >= 0 {
			identity, suffix = input[:at], input[at+1:]
			if identity == "" || suffix == "" {
				return Reference{}, invalid("local Reference requires ID@revision")
			}
		}
	}
	var revision string
	if suffix != "" {
		selector, err := clirevision.Parse(suffix, 4096)
		if err != nil {
			return Reference{}, invalid("revision selector: %v", err)
		}
		if selector.Current || selector.Ordinal != 0 {
			return Reference{}, fmt.Errorf("%w: revision selector needs online resolution", ErrUnsupported)
		}
		revision = selector.Token
	}
	uri, err := resolveURI(identity, scope, true)
	if err != nil {
		return Reference{}, err
	}
	return Reference{URI: uri, Revision: revision}, nil
}

// ParseOperand additionally recognizes the reserved CLI batch-ref::NAME spelling
// in an admitted operand position. Property strings never pass through it.
func ParseOperand(input, scope string) (Operand, error) {
	if name, ok := strings.CutPrefix(input, "batch-ref::"); ok {
		if !validName(name) {
			return Operand{}, invalid("invalid binding name")
		}
		return Operand{Binding: name}, nil
	}
	ref, err := ParseReference(input, scope)
	if err != nil {
		return Operand{}, err
	}
	return Operand{URI: ref.URI, Revision: ref.Revision}, nil
}

// CompileMember admits one independent emitter record. Internal and external
// names remain unresolved; whole-batch compilation checks their definitions,
// ordering and supplied arguments before any execution.
func CompileMember(raw []byte, scope string) (Member, []byte, error) {
	if len(raw) > MaxInputBytes {
		return Member{}, nil, fmt.Errorf("%w: bytes", ErrLimit)
	}
	m, err := parseMember(raw, scope, true)
	if err != nil {
		return Member{}, nil, err
	}
	encoded, err := marshalMember(m)
	if err != nil {
		return Member{}, nil, err
	}
	canonical, err := canonicalObject(encoded)
	if err != nil {
		return Member{}, nil, err
	}
	return cloneMember(m), append(bytes.Clone(canonical), '\n'), nil
}
