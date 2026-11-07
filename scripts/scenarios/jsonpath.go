package main

import (
	"fmt"
	"strconv"
	"strings"
)

// jsonPath is the minimal path subset scenario files may use: $.a.b[0].c and
// the array projection $.items[*].id. Nothing else is accepted, so a scenario
// cannot smuggle in filters, recursion or slices whose meaning differs between
// JSONPath implementations.
type jsonPath struct {
	raw  string
	segs []pathSeg
}

type segKind int

const (
	segKey segKind = iota
	segIndex
	segWildcard
)

type pathSeg struct {
	kind  segKind
	key   string
	index int
}

func isKeyChar(c byte, first bool) bool {
	switch {
	case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c == '_':
		return true
	case c >= '0' && c <= '9', c == '-':
		return !first
	}
	return false
}

func parseJSONPath(s string) (*jsonPath, error) {
	if !strings.HasPrefix(s, "$") {
		return nil, fmt.Errorf("json path %q must start with $", s)
	}
	rest := s[1:]
	if rest == "" {
		return nil, fmt.Errorf("json path %q selects the whole document; name a field", s)
	}
	var segs []pathSeg
	wildcards := 0
	for rest != "" {
		switch rest[0] {
		case '.':
			rest = rest[1:]
			end := 0
			for end < len(rest) && isKeyChar(rest[end], end == 0) {
				end++
			}
			if end == 0 {
				return nil, fmt.Errorf("json path %q: expected a field name after '.'", s)
			}
			segs = append(segs, pathSeg{kind: segKey, key: rest[:end]})
			rest = rest[end:]
		case '[':
			closing := strings.IndexByte(rest, ']')
			if closing < 0 {
				return nil, fmt.Errorf("json path %q: unterminated '['", s)
			}
			inner := rest[1:closing]
			rest = rest[closing+1:]
			if inner == "*" {
				wildcards++
				segs = append(segs, pathSeg{kind: segWildcard})
				continue
			}
			n, err := strconv.Atoi(inner)
			if err != nil || n < 0 || inner != strconv.Itoa(n) {
				return nil, fmt.Errorf("json path %q: only [N] with a plain non-negative index and [*] are supported, got [%s]", s, inner)
			}
			segs = append(segs, pathSeg{kind: segIndex, index: n})
		default:
			return nil, fmt.Errorf("json path %q: unsupported syntax at %q (only $.a.b[0].c and $.items[*].id are supported)", s, rest)
		}
	}
	if wildcards > 1 {
		return nil, fmt.Errorf("json path %q: at most one [*] is supported", s)
	}
	return &jsonPath{raw: s, segs: segs}, nil
}

func (p *jsonPath) hasWildcard() bool {
	for _, seg := range p.segs {
		if seg.kind == segWildcard {
			return true
		}
	}
	return false
}

// eval returns the value at the path: one element, or one per array element
// for a [*] projection. A missing field, wrong container type or short array is
// an error; a projection never quietly skips the elements it cannot read.
func (p *jsonPath) eval(doc any) ([]any, error) {
	cur := []any{doc}
	for _, seg := range p.segs {
		var next []any
		for _, v := range cur {
			switch seg.kind {
			case segKey:
				obj, ok := v.(map[string]any)
				if !ok {
					return nil, fmt.Errorf("json path %s: cannot read field %q of %s", p.raw, seg.key, describeJSON(v))
				}
				child, ok := obj[seg.key]
				if !ok {
					return nil, fmt.Errorf("json path %s: no field %q", p.raw, seg.key)
				}
				next = append(next, child)
			case segIndex:
				arr, ok := v.([]any)
				if !ok {
					return nil, fmt.Errorf("json path %s: cannot index [%d] into %s", p.raw, seg.index, describeJSON(v))
				}
				if seg.index >= len(arr) {
					return nil, fmt.Errorf("json path %s: index %d out of range (length %d)", p.raw, seg.index, len(arr))
				}
				next = append(next, arr[seg.index])
			case segWildcard:
				arr, ok := v.([]any)
				if !ok {
					return nil, fmt.Errorf("json path %s: cannot project [*] over %s", p.raw, describeJSON(v))
				}
				next = append(next, arr...)
			}
		}
		cur = next
	}
	return cur, nil
}

func describeJSON(v any) string {
	switch v.(type) {
	case map[string]any:
		return "an object"
	case []any:
		return "an array"
	case string:
		return "a string"
	case nil:
		return "null"
	case bool:
		return "a boolean"
	default:
		return "a number"
	}
}
