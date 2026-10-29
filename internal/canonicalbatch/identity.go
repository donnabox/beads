package canonicalbatch

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/steveyegge/beads/graphops"
)

// DecisionJSON returns canonical request identity for the selected experimental
// grammar. Capture labels are presentation: names are omitted and structural
// binding operands identify their defining operation's zero-based slot instead.
// The returned bytes are comparison material, not an executable batch request.
// NativeJSON and Members retain the original labels for execution and output.
//
// Only grammar-owned binding operands are rewritten. Ordinary JSON values,
// including query literals and query-local variable names, remain unchanged.
func (b *Batch) DecisionJSON() ([]byte, error) {
	if b == nil || len(b.members) == 0 {
		return nil, invalid("decision identity requires a compiled batch")
	}
	definitions := make(map[string]int)
	for i, m := range b.members {
		if m.Name != "" {
			definitions[m.Name] = i
		}
	}
	operations := make([]json.RawMessage, len(b.members))
	for i, original := range b.members {
		m := cloneMember(original)
		m.Name = ""
		normalize := func(operand *Operand, _ graphops.ResourceKind, _ bool) error {
			if operand.Binding == "" {
				return nil
			}
			index, exists := definitions[operand.Binding]
			if !exists || index >= i {
				return invalid("decision binding %q does not identify an earlier operation", operand.Binding)
			}
			operand.Binding = "operation-" + strconv.Itoa(index)
			return nil
		}
		for _, operand := range []*Operand{&m.Bead, &m.Link, &m.Source, &m.Target} {
			if err := normalize(operand, "", false); err != nil {
				return nil, err
			}
		}
		if m.Query != nil {
			var err error
			// Reuse the compiler's grammar traversal: refs.references is a batch
			// operand, while literal.value and row bindings are not.
			m.Query, _, err = resolveQuery(m.Query, "", false, normalize)
			if err != nil {
				return nil, fmt.Errorf("decision operation %d: %w", i, err)
			}
		}
		encoded, err := marshalMember(m)
		if err != nil {
			return nil, err
		}
		operations[i] = encoded
	}
	request, err := json.Marshal(struct {
		Version    string            `json:"capabilityVersion"`
		Operations []json.RawMessage `json:"operations"`
	}{Version, operations})
	if err != nil {
		return nil, err
	}
	return graphops.CanonicalizeJSON(request)
}
