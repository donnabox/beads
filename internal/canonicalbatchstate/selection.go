package canonicalbatchstate

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/canonicalbatch"
	"github.com/steveyegge/beads/internal/resourcequery"
)

// These fail-closed cumulative adapter budgets complement the shared selector's
// per-evaluation work/result limits. None authorizes truncated success.
const (
	MaxSelectionWork  = 1_000_000
	MaxSelectionBytes = 32 << 20
	MaxSelectedIDs    = 10_000
)

// Batch owns structural input resolution and fixed identity capture; the shared
// resourcequery kernel owns List admission, local predicates and value semantics.
// No staged record is promoted to a public revision-bearing Resource.
type selection struct {
	kind          graphops.ResourceKind
	scan          bool
	references    []string
	sourceBinding string
	selector      *resourcequery.ListIDSelector
}

func admitSelection(m canonicalbatch.Member, limits resourcequery.Limits) (*selection, error) {
	// The batch compiler validates the original grammar, including references,
	// before the only structural transformation: replacing this collection
	// source with a scan over its materialized, checked candidate inventory.
	raw, err := json.Marshal(struct {
		Operation string          `json:"operation"`
		Name      string          `json:"name"`
		Query     json.RawMessage `json:"query"`
	}{"select", m.Name, m.Query})
	if err != nil {
		return nil, fmt.Errorf("%w: selection envelope", ErrInvalid)
	}
	checked, _, err := canonicalbatch.CompileMember(raw, "")
	if err != nil {
		if errors.Is(err, canonicalbatch.ErrUnsupported) {
			return nil, fmt.Errorf("%w: %w", ErrUnsupported, err)
		}
		return nil, fmt.Errorf("%w: %w", ErrInvalid, err)
	}
	if len(checked.Query) > 262144 {
		return nil, fmt.Errorf("%w: staged selection envelope bytes", ErrLimit)
	}
	f := nodeFields(checked.Query)
	if text(f, "profile") != "list" {
		return nil, fmt.Errorf("%w: staged receiver requires the local List profile", ErrUnsupported)
	}
	root := nodeFields(f["selection"])
	source := root
	filtered := text(root, "op") == "filter"
	if filtered {
		source = nodeFields(root["input"])
	}
	collection := text(source, "collection")
	s := &selection{kind: graphops.KindBead}
	if collection == "links" {
		s.kind = graphops.KindLink
	}
	switch text(source, "op") {
	case "scan":
		s.scan = true
	case "refs":
		refs := bytes.TrimSpace(source["references"])
		if len(refs) > 0 && refs[0] == '{' {
			s.sourceBinding = text(nodeFields(refs), "binding")
		} else {
			var entries []json.RawMessage
			if json.Unmarshal(refs, &entries) != nil {
				return nil, ErrInvalid
			}
			for _, entry := range entries {
				var uri string
				if json.Unmarshal(entry, &uri) != nil {
					return nil, fmt.Errorf("%w: staged selection has no historical acquisition", ErrUnsupported)
				}
				s.references = append(s.references, uri)
			}
		}
	default:
		return nil, fmt.Errorf("%w: List source requires scan or current refs", ErrUnsupported)
	}
	if m.Kind != "" && m.Kind != s.kind {
		return nil, ErrKind
	}
	scan, err := json.Marshal(struct {
		Operation  string `json:"op"`
		Collection string `json:"collection"`
	}{"scan", collection})
	if err != nil {
		return nil, err
	}
	fragment := scan
	if filtered {
		// Only input is replaced. In particular predicate.literal.value is
		// opaque JSON, even if its keys spell op, refs, binding or input.
		root["input"] = scan
		fragment, err = json.Marshal(root)
		if err != nil {
			return nil, err
		}
	}
	s.selector, err = resourcequery.CompileListIDs(fragment, limits)
	if err != nil {
		var failure *resourcequery.Failure
		if errors.As(err, &failure) && failure.Code == "profile-mismatch" {
			return nil, fmt.Errorf("%w: %w", ErrUnsupported, err)
		}
		return nil, selectionError(err)
	}
	if s.selector.ResourceKind() != s.kind {
		return nil, ErrKind
	}
	return s, nil
}

func nodeFields(raw json.RawMessage) map[string]json.RawMessage {
	var fields map[string]json.RawMessage
	_ = json.Unmarshal(raw, &fields)
	return fields
}

func text(fields map[string]json.RawMessage, key string) string {
	var value string
	_ = json.Unmarshal(fields[key], &value)
	return value
}

func selectionError(err error) error {
	switch {
	case errors.Is(err, resourcequery.ErrUnsupported):
		return fmt.Errorf("%w: %w", ErrUnsupported, err)
	case errors.Is(err, resourcequery.ErrLimit):
		return fmt.Errorf("%w: %w", ErrLimit, err)
	case errors.Is(err, resourcequery.ErrUnavailable):
		return fmt.Errorf("%w: %w", ErrNotFound, err)
	case errors.Is(err, resourcequery.ErrInvalid), errors.Is(err, resourcequery.ErrDataset):
		return fmt.Errorf("%w: %w", ErrInvalid, err)
	default:
		return err
	}
}

func (s *selection) checkBindings(names map[string]graphops.ResourceKind) error {
	if s.sourceBinding == "" {
		return nil
	}
	kind, ok := names[s.sourceBinding]
	if !ok {
		return fmt.Errorf("%w: unknown or forward query input binding %q", ErrInvalid, s.sourceBinding)
	}
	if kind != s.kind {
		return fmt.Errorf("%w: query input binding %q", ErrKind, s.sourceBinding)
	}
	return nil
}

func (s *selection) checkTypes(input Input) error {
	for _, uri := range s.references {
		if _, err := graphops.ScopeLocalPublicResourceID(input.Scope, uri); err != nil {
			return fmt.Errorf("%w: cross-Scope staged selection", ErrUnsupported)
		}
	}
	// Validate authority even for an empty inventory or a branch that later
	// short-circuits. The shared selector deliberately supports Core-only
	// nominal tests with both maps nil, while family tests require authority.
	_, err := s.selector.Select(input.Context, resourcequery.ListInventory{Kind: s.kind, Candidates: []resourcequery.ListCandidate{}, TypeKinds: input.TypeKinds, TypeFamilies: input.TypeFamilies})
	return selectionError(err)
}

func (s *staging) selectionTick() error {
	if err := s.input.Context.Err(); err != nil {
		return err
	}
	s.selectionWork++
	if s.selectionWork > MaxSelectionWork {
		return fmt.Errorf("%w: cumulative adapter work", ErrLimit)
	}
	return nil
}

func (s *staging) selectIDs(index int, a admitted) error {
	query := a.selection
	ids := []string{}
	if query.scan {
		for id, r := range s.live {
			if err := s.selectionTick(); err != nil {
				return err
			}
			if r.Kind == query.kind {
				ids = append(ids, id)
			}
		}
	} else if query.sourceBinding != "" {
		capture, ok := s.bindings[query.sourceBinding]
		if !ok {
			return ErrInvalid
		}
		ids = append(ids, capture.ids...)
	} else {
		ids = append(ids, query.references...)
	}
	seen := map[string]bool{}
	unique := make([]string, 0, len(ids))
	for _, id := range ids {
		if err := s.selectionTick(); err != nil {
			return err
		}
		if !seen[id] {
			seen[id] = true
			unique = append(unique, id)
		}
	}
	sort.Slice(unique, func(i, j int) bool { return graphops.CompareCodeUnits(unique[i], unique[j]) < 0 })
	candidates := make([]resourcequery.ListCandidate, 0, len(unique))
	for _, id := range unique {
		if err := s.selectionTick(); err != nil {
			return err
		}
		r, ok := s.live[id]
		if !ok {
			return fmt.Errorf("%w: captured/query Resource", ErrNotFound)
		}
		if r.Kind != query.kind {
			return ErrKind
		}
		view, err := revisionlessCandidate(r)
		if err != nil {
			return err
		}
		if err := s.chargeSelectionBytes(len(view.Data)); err != nil {
			return err
		}
		candidates = append(candidates, view)
	}
	selected, err := query.selector.Select(s.input.Context, resourcequery.ListInventory{Kind: query.kind, Candidates: candidates, TypeKinds: s.input.TypeKinds, TypeFamilies: s.input.TypeFamilies})
	if err != nil {
		return selectionError(err)
	}
	if len(selected) > MaxSelectedIDs {
		return fmt.Errorf("%w: complete captured set", ErrLimit)
	}
	// Selection meaning is unordered; the Batch's canonical URI ordering makes
	// subsequent set mutation expansion and outcome diagnostics deterministic.
	sort.Slice(selected, func(i, j int) bool { return graphops.CompareCodeUnits(selected[i], selected[j]) < 0 })
	s.bindings[a.member.Name] = capturedBinding{kind: query.kind, ids: append([]string{}, selected...), set: true}
	s.selections = append(s.selections, SelectionCapture{OperationIndex: index, Name: a.member.Name, Kind: query.kind, IDs: append([]string{}, selected...)})
	return nil
}

func revisionlessCandidate(r Resource) (resourcequery.ListCandidate, error) {
	// This closed view contains no source/target, provenance, owned captures,
	// stale baseline revision or invented empty revision, even for Link sets.
	view := struct {
		ID         string                   `json:"id"`
		Type       canonicalbatch.Reference `json:"type"`
		Properties json.RawMessage          `json:"properties"`
		Metadata   json.RawMessage          `json:"metadata,omitempty"`
	}{r.ID, r.Type, r.Properties, r.Metadata}
	data, err := json.Marshal(view)
	if err != nil {
		return resourcequery.ListCandidate{}, fmt.Errorf("%w: revisionless candidate: %v", ErrInvalid, err)
	}
	return resourcequery.ListCandidate{Data: data}, nil
}

func (s *staging) chargeSelectionResource(r Resource) error {
	view, err := revisionlessCandidate(r)
	if err != nil {
		return err
	}
	return s.chargeSelectionBytes(len(view.Data))
}

func (s *staging) chargeSelectionBytes(size int) error {
	if err := s.input.Context.Err(); err != nil {
		return err
	}
	if size > MaxSelectionBytes-s.selectionBytes {
		return fmt.Errorf("%w: cumulative serialized candidate work", ErrLimit)
	}
	s.selectionBytes += size
	return nil
}
