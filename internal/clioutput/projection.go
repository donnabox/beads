package clioutput

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
)

// Protection describes semantic context, not a Resource schema. Required names
// retain complete member values whenever this object remains represented. Use
// them for identity, Type, revision, endpoints, or a collective subject group.
// Members and Items apply recursively only at explicitly described locations;
// arbitrary objects with an "id" field receive no automatic protection.
// An optional containing member/array element can still be omitted in full.
// Descriptors must not expose graph membership as optional representational data.
type Protection struct {
	Required []string
	Members  map[string]*Protection
	Items    *Protection
}

func cloneProtection(p *Protection, active map[*Protection]bool) (*Protection, error) {
	if p == nil {
		return nil, nil
	}
	if active[p] {
		return nil, fmt.Errorf("cyclic projection protection descriptor")
	}
	active[p] = true
	defer delete(active, p)
	if p.Items != nil && (len(p.Required) != 0 || len(p.Members) != 0) {
		return nil, fmt.Errorf("projection protection cannot describe both object and array")
	}
	out := &Protection{Required: append([]string(nil), p.Required...)}
	sort.Strings(out.Required)
	out.Members = make(map[string]*Protection, len(p.Members))
	for name, member := range p.Members {
		child, err := cloneProtection(member, active)
		if err != nil {
			return nil, err
		}
		out.Members[name] = child
	}
	var err error
	out.Items, err = cloneProtection(p.Items, active)
	return out, err
}

func project(raw json.RawMessage, context *Protection, include, omit *mask, full, root bool, path string) (json.RawMessage, bool, error) {
	if omit.selectsAll() {
		if root {
			return nil, false, fmt.Errorf("cannot omit the Resource root")
		}
		return nil, false, nil
	}
	full = full || include.selectsAll()
	raw = bytes.TrimSpace(raw)
	switch raw[0] {
	case '{':
		return projectObject(raw, context, include, omit, full, root, path)
	case '[':
		return projectArray(raw, context, include, omit, full, path)
	default:
		if !full {
			return nil, false, nil
		}
		if !bytes.Equal(raw, []byte("null")) && context != nil && (len(context.Required) != 0 || len(context.Members) != 0 || context.Items != nil) {
			return nil, false, fmt.Errorf("projection context at %q has the wrong value shape", path)
		}
		return raw, true, nil
	}
}

func projectObject(raw json.RawMessage, context *Protection, include, omit *mask, full, root bool, path string) (json.RawMessage, bool, error) {
	var source map[string]json.RawMessage
	if err := json.Unmarshal(raw, &source); err != nil {
		return nil, false, err
	}
	out := make(map[string]json.RawMessage)
	// Sorting also makes the first reported descriptor/mask error deterministic.
	keys := make([]string, 0, len(source))
	for key := range source {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if !full && include.child(key) == nil {
			continue
		}
		var childContext *Protection
		if context != nil {
			childContext = context.Members[key]
		}
		value, keep, err := project(source[key], childContext, include.child(key), omit.child(key), full, false, pointerChild(path, key))
		if err != nil {
			return nil, false, err
		}
		if keep {
			out[key] = value
		}
	}
	if !full && !root && len(out) == 0 {
		return nil, false, nil
	}
	if context != nil {
		if context.Items != nil {
			return nil, false, fmt.Errorf("expected array context at %q", path)
		}
		for _, key := range context.Required {
			value, exists := source[key]
			if !exists {
				return nil, false, fmt.Errorf("missing mandatory context %q", pointerChild(path, key))
			}
			if removesValue(value, omit.child(key)) {
				return nil, false, fmt.Errorf("cannot omit mandatory context %q", pointerChild(path, key))
			}
			out[key] = value
		}
	}
	encoded, err := json.Marshal(out)
	return encoded, true, err
}

func projectArray(raw json.RawMessage, context *Protection, include, omit *mask, full bool, path string) (json.RawMessage, bool, error) {
	var source []json.RawMessage
	if err := json.Unmarshal(raw, &source); err != nil {
		return nil, false, err
	}
	out := make([]json.RawMessage, 0, len(source))
	var itemContext *Protection
	if context != nil {
		if len(context.Required) != 0 || len(context.Members) != 0 {
			return nil, false, fmt.Errorf("expected object context at %q", path)
		}
		itemContext = context.Items
	}
	for i, value := range source {
		key := strconv.Itoa(i)
		if !full && include.child(key) == nil {
			continue
		}
		projected, keep, err := project(value, itemContext, include.child(key), omit.child(key), full, false, pointerChild(path, key))
		if err != nil {
			return nil, false, err
		}
		if keep {
			out = append(out, projected)
		}
	}
	if !full && len(out) == 0 {
		return nil, false, nil
	}
	encoded, err := json.Marshal(out)
	return encoded, true, err
}

// removesValue detects only paths present in the original value. An absent path
// remains a no-op, including inside an otherwise protected member.
func removesValue(raw json.RawMessage, omit *mask) bool {
	if omit == nil {
		return false
	}
	if omit.terminal {
		return true
	}
	raw = bytes.TrimSpace(raw)
	switch raw[0] {
	case '{':
		var members map[string]json.RawMessage
		if json.Unmarshal(raw, &members) != nil {
			return false // Input JSON was validated before projection.
		}
		for key, value := range members {
			if removesValue(value, omit.child(key)) {
				return true
			}
		}
	case '[':
		var items []json.RawMessage
		if json.Unmarshal(raw, &items) != nil {
			return false
		}
		for i, value := range items {
			if removesValue(value, omit.child(strconv.Itoa(i))) {
				return true
			}
		}
	}
	return false
}
