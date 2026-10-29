package resourcequery

import (
	"context"
	"encoding/json"
	"sort"
	"strings"

	graph "github.com/steveyegge/beads/graphops"
)

// ListCandidate is a closed, detached revisionless view containing id, type,
// properties and optional metadata. It is not a complete public Resource.
// Type references retain the same string or pinned-object representation as
// Query. Candidate kind belongs to ListInventory, including for Link views.
type ListCandidate struct{ Data json.RawMessage }

// ListInventory asserts the complete authorized candidate inventory of Kind at
// one staged step, with complete Type authority from that same staged context.
// Candidates must be nonnil; an explicitly empty slice is a valid empty
// inventory. Both authority maps may be nil when family tests are not used.
// Supplying maps asserts complete kind-correct reflexive ancestry for all known
// Types; supplying just one map is an incomplete assertion. Callers must prevent concurrent mutation
// and recapture after any staged mutation. Select opens no storage or provider.
type ListInventory struct {
	Kind         graph.ResourceKind
	Candidates   []ListCandidate
	TypeKinds    map[string]graph.ResourceKind
	TypeFamilies map[string][]string
}

// ListIDSelector is an immutable admitted selection, safe to reuse with fresh
// detached inventories. It carries no query result controls or external bindings.
type ListIDSelector struct {
	request *Request
	limits  Limits
}

// ResourceKind is the inventory kind required by the admitted scan.
func (s *ListIDSelector) ResourceKind() graph.ResourceKind {
	if s == nil || s.request == nil {
		return ""
	}
	return s.request.kind
}

// CompileListIDs admits the existing scan or filter(scan) List selection
// fragment, never a full Query. The fixed List/resources envelope is included in
// RequestBytes; all expanded literals count toward the usual AST bounds.
func CompileListIDs(selection json.RawMessage, limits Limits) (*ListIDSelector, error) {
	l, err := normalizeLimits(limits)
	if err != nil {
		return nil, err
	}
	const prefix = `{"ir":"bdp-query-design-1","profile":"list","selection":`
	const suffix = `,"result":{"form":"resources"}}`
	if len(selection) > l.RequestBytes-len(prefix)-len(suffix) {
		return nil, fail("request-limit", "/query/selection", "selection exceeds admitted limits", ErrLimit)
	}
	raw := make([]byte, 0, len(prefix)+len(selection)+len(suffix))
	raw = append(raw, prefix...)
	raw = append(raw, selection...)
	raw = append(raw, suffix...)
	r, err := Decode(raw)
	if err != nil {
		return nil, err
	}
	if err = admit(r, l); err != nil {
		return nil, err
	}
	if err = stagedLocal(r.root); err != nil {
		return nil, err
	}
	return &ListIDSelector{request: r, limits: l}, nil
}

// stagedLocal checks every admitted node, including unreachable branches. It
// narrows List's local surface without defining another parser or public AST.
func stagedLocal(n *node) error {
	switch n.op {
	case "scan", "filter", "binding", "literal", "bool", "and", "or", "not", "present", "compare", "contains", "member", "typeIs", "id":
	case "field":
		if len(n.path) == 0 || n.path[0].indexed {
			return stagedUnsupported(n)
		}
		switch n.path[0].key {
		case "id", "type", "properties", "metadata":
		default:
			return stagedUnsupported(n)
		}
	default:
		return stagedUnsupported(n)
	}
	for _, child := range []*node{n.input, n.roots, n.of, n.predicate, n.left, n.right, n.text, n.substring, n.value, n.array} {
		if child != nil {
			if err := stagedLocal(child); err != nil {
				return err
			}
		}
	}
	for _, child := range n.operands {
		if err := stagedLocal(child); err != nil {
			return err
		}
	}
	return nil
}
func stagedUnsupported(n *node) error {
	return fail("staged-list-unsupported", n.at, "selection requires more than the revisionless local view", ErrUnsupported)
}

type listView struct {
	id, typ string
	body    map[string]any
}

// readListView parses and copies the candidate independently of readRecord.
// No empty or fabricated revision enters the complete Resource boundary.
func readListView(raw []byte) (listView, error) {
	bad := func() (listView, error) {
		return listView{}, fail("invalid-list-candidate", "", "invalid closed revisionless candidate", ErrDataset)
	}
	canonical, err := graph.CanonicalizeJSON(raw)
	if err != nil {
		return bad()
	}
	v, err := jsonValue(canonical)
	if err != nil {
		return bad()
	}
	m, err := object(v, "", []string{"id", "type", "properties"}, []string{"metadata"})
	if err != nil {
		return bad()
	}
	id, ok := m["id"].(string)
	if !ok || graph.ValidateTypeURL(id) != nil || graph.ValidateReferenceURI(id) != nil || strings.Contains(id, "?") {
		return bad()
	}
	typ, err := parseReference(m["type"], "")
	if err != nil || graph.ValidateTypeURL(typ.uri) != nil {
		return bad()
	}
	if _, ok = m["properties"].(map[string]any); !ok {
		return bad()
	}
	if meta, present := m["metadata"]; present {
		if _, ok = meta.(map[string]any); !ok {
			return bad()
		}
	}
	return listView{id: id, typ: typ.uri, body: m}, nil
}

// Select returns sorted distinct actual identities, or nil IDs on any failure.
// Missing capture, malformed candidates, incomplete authority and exhausted
// limits never become successful empty or partial selections.
func (s *ListIDSelector) Select(ctx context.Context, input ListInventory) ([]string, error) {
	if ctx == nil {
		return nil, fail("invalid-context", "", "a context is required", ErrInvalid)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if s == nil || s.request == nil {
		return nil, fail("invalid-selector", "", "CompileListIDs must construct the selector", ErrInvalid)
	}
	if !input.Kind.Valid() || input.Kind != s.request.kind {
		return nil, fail("invalid-inventory-kind", "", "inventory kind differs from selection", ErrDataset)
	}
	if input.Candidates == nil {
		return nil, fail("list-inventory-unavailable", "", "complete staged inventory is required", ErrUnavailable)
	}
	work := 0
	tick := func() error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if work >= s.limits.MaxWork {
			return fail("work-limit", "", "query work bound exhausted", ErrLimit)
		}
		work++
		return nil
	}
	families, err := listFamilies(input, s.request.families, tick)
	if err != nil {
		return nil, err
	}
	if len(input.Candidates) > s.limits.MaxWork-work {
		return nil, fail("work-limit", "", "query work bound exhausted", ErrLimit)
	}
	views := make([]listView, 0, len(input.Candidates))
	seen := map[string]bool{}
	for _, candidate := range input.Candidates {
		if err := tick(); err != nil {
			return nil, err
		}
		view, err := readListView(candidate.Data)
		if err != nil {
			return nil, err
		}
		if seen[view.id] {
			return nil, fail("duplicate-current-resource", "", "inventory contains duplicate identities", ErrDataset)
		}
		seen[view.id] = true
		if kind, ok := input.TypeKinds[view.typ]; input.TypeKinds != nil && (!ok || kind != input.Kind) {
			return nil, fail("invalid-family-authority", "", "candidate Type lacks kind-correct authority", ErrDataset)
		}
		views = append(views, view)
	}
	result := make([]string, 0)
	root := s.request.root
	for _, view := range views {
		if err := tick(); err != nil {
			return nil, err
		}
		yes := true
		if root.op == "filter" {
			k := predicateKernel{
				tick: tick,
				val: func(n *node) (value, error) {
					if err := tick(); err != nil {
						return value{}, err
					}
					if n.op == "literal" {
						return value{data: n.literal, present: true}, nil
					}
					if err := tick(); err != nil {
						return value{}, err
					}
					if n.op == "id" {
						return value{data: view.id, present: true}, nil
					}
					return fieldValue(view.body, n.path, tick)
				},
				typeIs: func(n *node) (bool, error) {
					if err := tick(); err != nil {
						return false, err
					}
					return matchesType(view.typ, n, families), nil
				},
			}
			var err error
			yes, err = k.predicate(root.predicate)
			if err != nil {
				return nil, err
			}
		}
		if yes {
			if len(result) >= s.limits.MaxResults {
				return nil, fail("result-limit", "", "query collection bound exhausted", ErrLimit)
			}
			result = append(result, view.id)
		}
	}
	sort.Strings(result)
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

// listFamilies validates supplied authority without inventing a Types capability
// for Core-only callers. Query's exact Type test remains nominal equality;
// family tests require the same honest authority even when never evaluated.
func listFamilies(input ListInventory, uses []familyUse, tick func() error) (map[string]map[string]bool, error) {
	if len(uses) != 0 && (input.TypeKinds == nil || input.TypeFamilies == nil) {
		return nil, fail("family-mechanism-unavailable", uses[0].at, "authoritative Type family interpretation is unsupported", ErrUnsupported)
	}
	if input.TypeKinds == nil && input.TypeFamilies == nil {
		return nil, nil
	}
	if input.TypeKinds == nil || input.TypeFamilies == nil {
		return nil, fail("invalid-family-authority", "", "staged Type authority is incomplete", ErrDataset)
	}
	families := map[string]map[string]bool{}
	// Validate the asserted complete Type authority even for scan-only requests,
	// empty inventories and branches whose value evaluation will short-circuit.
	for uri, kind := range input.TypeKinds {
		if err := tick(); err != nil {
			return nil, err
		}
		if graph.ValidateTypeURL(uri) != nil || !kind.Valid() {
			return nil, fail("invalid-family-authority", "", "invalid Type kind authority", ErrDataset)
		}
		if err := loadFamilyAuthority(uri, kind, input.TypeKinds, input.TypeFamilies, families, tick); err != nil {
			return nil, err
		}
	}
	for uri := range input.TypeFamilies {
		if err := tick(); err != nil {
			return nil, err
		}
		if _, ok := input.TypeKinds[uri]; !ok {
			return nil, fail("invalid-family-authority", "", "ancestry has no authoritative Type kind", ErrDataset)
		}
	}
	// A complete ancestry must also include every ancestor's ancestry.
	for _, family := range families {
		for ancestor := range family {
			for inherited := range families[ancestor] {
				if err := tick(); err != nil {
					return nil, err
				}
				if !family[inherited] {
					return nil, fail("invalid-family-authority", "", "Type ancestry is not transitively complete", ErrDataset)
				}
			}
		}
	}

	for _, use := range uses {
		if err := tick(); err != nil {
			return nil, err
		}
		actual, ok := input.TypeKinds[use.uri]
		if !ok {
			return nil, fail("family-authority-unavailable", use.at, "queried Type has no authoritative kind", ErrUnavailable)
		}
		if actual != use.kind {
			return nil, fail("invalid-type-kind", use.at, "queried Type describes another Resource kind", ErrInvalid)
		}
	}
	return families, nil
}
