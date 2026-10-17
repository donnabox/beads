package memorytypes

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Violation identifies the first invalid field (a JSON Pointer) and expectation.
// Inputs and offending values are never retained or rewritten.
type Violation struct {
	Path     string
	Expected string
}

func (v *Violation) Error() string { return fmt.Sprintf("%s: expected %s", v.Path, v.Expected) }

func invalid(path, expected string) error { return &Violation{path, expected} }

// ValidateProperties checks the complete prospective properties object, including
// inherited Memory requirements and applicability interval order. It does not
// merge patches, infer a type, insert defaults, or mutate the input.
func ValidateProperties(kind BeadKind, data json.RawMessage) error {
	if !kind.IsMemory() {
		return invalid("/type", "a fixed Memory family contract")
	}
	obj, err := object(data, "/properties")
	if err != nil {
		return err
	}
	for _, field := range []string{"title", "body"} {
		if err := textField(obj, field, "/properties", true, field == "title"); err != nil {
			return err
		}
	}
	if _, exists := obj["kind"]; exists {
		return invalid("/properties/kind", "no kind discriminator")
	}
	if raw, ok := obj["inception"]; ok {
		if err := inception(raw); err != nil {
			return err
		}
	}
	if raw, ok := obj["applicability"]; ok {
		if err := applicability(raw); err != nil {
			return err
		}
	}
	switch kind {
	case Observation:
		if err := textField(obj, "statement", "/properties", true, true); err != nil {
			return err
		}
		if raw, ok := obj["observedAt"]; ok {
			_, err := instant(raw, "/properties/observedAt")
			return err
		}
	case Decision:
		if err := textField(obj, "choice", "/properties", true, true); err != nil {
			return err
		}
		if raw, ok := obj["alternatives"]; ok {
			return textArray(raw, "/properties/alternatives", false)
		}
	case Procedure:
		return textArray(obj["steps"], "/properties/steps", true)
	case Policy:
		if err := textField(obj, "rule", "/properties", true, true); err != nil {
			return err
		}
		if raw, ok := obj["authority"]; ok {
			return referenceValue(raw, "/properties/authority")
		}
	case Lesson:
		return textField(obj, "takeaway", "/properties", true, true)
	}
	return nil
}

// ValidateLinkProperties checks the shared open properties object. Endpoint
// checks are separate. It does not accept an unknown or Dependency relationship.
func ValidateLinkProperties(kind LinkKind, data json.RawMessage) error {
	if !kind.valid() {
		return invalid("/type", "a fixed Memory relationship")
	}
	obj, err := object(data, "/properties")
	if err != nil {
		return err
	}
	if _, exists := obj["kind"]; exists {
		return invalid("/properties/kind", "no kind discriminator")
	}
	return textField(obj, "note", "/properties", false, true)
}

func object(raw json.RawMessage, path string) (map[string]json.RawMessage, error) {
	if !utf8.Valid(raw) {
		return nil, invalid(path, "a UTF-8 JSON object")
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil || obj == nil {
		return nil, invalid(path, "an object")
	}
	return obj, nil
}

func textValue(raw json.RawMessage, path string, nonblank bool) (string, error) {
	var text *string
	if err := json.Unmarshal(raw, &text); err != nil || text == nil {
		return "", invalid(path, "a string")
	}
	// The pinned source checker uses Python's Unicode-aware non-whitespace rule.
	if nonblank && strings.TrimFunc(*text, func(r rune) bool { return unicode.IsSpace(r) || (r >= '\x1c' && r <= '\x1f') }) == "" {
		return "", invalid(path, "a nonblank string")
	}
	return *text, nil
}

func textField(obj map[string]json.RawMessage, key, path string, required, nonblank bool) error {
	raw, ok := obj[key]
	if !ok && !required {
		return nil
	}
	_, err := textValue(raw, path+"/"+key, nonblank)
	return err
}

func textArray(raw json.RawMessage, path string, nonempty bool) error {
	var values []json.RawMessage
	if err := json.Unmarshal(raw, &values); err != nil || values == nil {
		return invalid(path, "an array of nonblank strings")
	}
	if nonempty && len(values) == 0 {
		return invalid(path, "at least one step")
	}
	for i, value := range values {
		if _, err := textValue(value, fmt.Sprintf("%s/%d", path, i), true); err != nil {
			return err
		}
	}
	return nil
}

func closed(obj map[string]json.RawMessage, path string, allowed ...string) error {
	keys := make([]string, 0, len(obj))
	for key := range obj {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		found := false
		for _, a := range allowed {
			if a == key {
				found = true
				break
			}
		}
		if !found {
			escaped := strings.ReplaceAll(strings.ReplaceAll(key, "~", "~0"), "/", "~1")
			return invalid(path+"/"+escaped, "no additional member")
		}
	}
	return nil
}

func inception(raw json.RawMessage) error {
	const path = "/properties/inception"
	obj, err := object(raw, path)
	if err != nil {
		return err
	}
	if err := closed(obj, path, "method", "origin"); err != nil {
		return err
	}
	method, err := textValue(obj["method"], path+"/method", false)
	if err != nil {
		return err
	}
	switch method {
	case "native", "converted", "imported", "unknown":
	default:
		return invalid(path+"/method", "native, converted, imported or unknown")
	}
	if origin, ok := obj["origin"]; ok {
		return referenceValue(origin, path+"/origin")
	}
	return nil
}
