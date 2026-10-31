// Package canonicaledit lowers ordinary object edits without reading or mutating Resources.
package canonicaledit

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/graphpatch"
)

type operation struct {
	Op    string          `json:"op"`
	Path  string          `json:"path"`
	Value json.RawMessage `json:"value,omitempty"`
}

// MergeOperations lowers a shallow merge to deterministic top-level adds. A
// nested object is one value; it is never recursively merged with old state.
func MergeOperations(raw json.RawMessage) ([]byte, error) {
	canonical, err := graph.CanonicalizeJSON(raw)
	if err != nil {
		return nil, err
	}
	if len(canonical) == 0 || canonical[0] != '{' {
		return nil, fmt.Errorf("merge input must be a JSON object")
	}
	var values map[string]json.RawMessage
	if err := json.Unmarshal(canonical, &values); err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return graph.CompareCodeUnits(keys[i], keys[j]) < 0 })
	operations := make([]operation, 0, len(keys))
	for _, key := range keys {
		operations = append(operations, operation{Op: "add", Path: pointer(key), Value: values[key]})
	}
	return json.Marshal(operations)
}

func pointer(key string) string {
	return "/" + strings.ReplaceAll(strings.ReplaceAll(key, "~", "~0"), "/", "~1")
}

// KeyedValues parses keyed scalar edits with last occurrence winning.
func KeyedValues(assignments []string) (map[string]json.RawMessage, error) {
	values := make(map[string]json.RawMessage, len(assignments))
	for _, assignment := range assignments {
		key, value, ok := strings.Cut(assignment, "=")
		if !ok || !utf8.ValidString(key) {
			return nil, fmt.Errorf("set requires a UTF-8 KEY=VALUE")
		}
		parsed, err := Scalar(value)
		if err != nil {
			return nil, err
		}
		values[key] = parsed
	}
	return values, nil
}

// Lower keyed operations without reading the Resource. Set order is first
// occurrence, values are last occurrence, and deduplicated unsets run last.
// The writer checks the full submitted edit and resulting Type policy in its
// guarded transaction; add-null/remove makes an absent writable unset succeed.
func KeyedOperations(sets, unsets []string) ([]byte, error) {
	values := make(map[string]json.RawMessage)
	order := make([]string, 0, len(sets))
	for _, assignment := range sets {
		key, value, ok := strings.Cut(assignment, "=")
		if !ok || !utf8.ValidString(key) {
			return nil, fmt.Errorf("property set requires a UTF-8 KEY=VALUE")
		}
		parsed, err := Scalar(value)
		if err != nil {
			return nil, err
		}
		if _, exists := values[key]; !exists {
			order = append(order, key)
		}
		values[key] = parsed
	}
	operations := make([]operation, 0, len(order)+2*len(unsets))
	for _, key := range order {
		operations = append(operations, operation{Op: "add", Path: pointer(key), Value: values[key]})
	}
	seen := make(map[string]bool)
	for _, key := range unsets {
		if !utf8.ValidString(key) {
			return nil, fmt.Errorf("property unset requires a UTF-8 key")
		}
		if !seen[key] {
			seen[key] = true
			operations = append(operations, operation{Op: "add", Path: pointer(key), Value: json.RawMessage("null")}, operation{Op: "remove", Path: pointer(key)})
		}
	}
	raw, err := json.Marshal(operations)
	if err != nil {
		return nil, err
	}
	if _, err := graphpatch.Parse(raw); err != nil {
		return nil, err
	}
	return raw, nil
}

// Numeric-looking JSON values must pass the same exact-number admission as
// object inputs. Overflow/precision errors never turn a number into a string.
func Scalar(value string) (json.RawMessage, error) {
	if !utf8.ValidString(value) {
		return nil, fmt.Errorf("set value must be valid UTF-8")
	}
	if value == "null" || value == "true" || value == "false" || (len(value) > 0 && (value[0] == '-' || value[0] >= '0' && value[0] <= '9') && json.Valid([]byte(value))) {
		canonical, err := graph.CanonicalizeJSON([]byte(value))
		if err != nil {
			return nil, err
		}
		return canonical, nil
	}
	encoded, err := json.Marshal(value)
	return encoded, err
}
