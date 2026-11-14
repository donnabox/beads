package clioutput

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// mask is a union of decoded RFC 6901 paths. A terminal selects its complete
// value. Children remain useful for deterministic original-index evaluation.
type mask struct {
	terminal bool
	children map[string]*mask
}

func parseMask(paths []string) (*mask, error) {
	root := &mask{}
	for _, path := range paths {
		parts, err := pointerParts(path)
		if err != nil {
			return nil, err
		}
		node := root
		for _, part := range parts {
			if node.children == nil {
				node.children = make(map[string]*mask)
			}
			if node.children[part] == nil {
				node.children[part] = &mask{}
			}
			node = node.children[part]
		}
		node.terminal = true
	}
	return root, nil
}

func pointerParts(path string) ([]string, error) {
	if !utf8.ValidString(path) {
		return nil, fmt.Errorf("JSON Pointer must be valid UTF-8")
	}
	if path == "" {
		return nil, nil
	}
	if !strings.HasPrefix(path, "/") {
		return nil, fmt.Errorf("JSON Pointer %q must be empty or begin with /", path)
	}
	parts := strings.Split(path[1:], "/")
	for i, part := range parts {
		var decoded strings.Builder
		for j := 0; j < len(part); j++ {
			if part[j] != '~' {
				decoded.WriteByte(part[j])
				continue
			}
			j++
			if j == len(part) || (part[j] != '0' && part[j] != '1') {
				return nil, fmt.Errorf("invalid JSON Pointer escape in %q", path)
			}
			if part[j] == '0' {
				decoded.WriteByte('~')
			} else {
				decoded.WriteByte('/')
			}
		}
		parts[i] = decoded.String()
	}
	return parts, nil
}

func (m *mask) child(key string) *mask {
	if m == nil {
		return nil
	}
	return m.children[key]
}

func (m *mask) selectsAll() bool {
	return m != nil && m.terminal
}

func pointerChild(path, key string) string {
	key = strings.ReplaceAll(strings.ReplaceAll(key, "~", "~0"), "/", "~1")
	return path + "/" + key
}

// validateDepth admits whole top-level sections and the explicitly declared
// nested projection subtrees. This keeps an optional protocol plane opaque
// unless its adapter deliberately exposes finer representation masks.
func validateDepth(paths []string, expandable *mask) error {
	for _, path := range paths {
		parts, err := pointerParts(path)
		if err != nil {
			return err
		}
		if len(parts) <= 1 {
			continue
		}
		node := expandable
		allowed := node.selectsAll()
		for _, part := range parts {
			node = node.child(part)
			allowed = allowed || node.selectsAll()
		}
		if !allowed {
			return fmt.Errorf("nested projection is not admitted at %q", path)
		}
	}
	return nil
}
