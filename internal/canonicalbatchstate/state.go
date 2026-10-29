// Package canonicalbatchstate stages canonical CRUD against a detached working
// graph. It performs no I/O, allocation persistence, public revision minting,
// authorization, Type validation, ownership publication, or event emission.
// A checked publisher must admit and commit the complete returned plan in the
// same transaction that acquired its baseline.
package canonicalbatchstate

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/canonicalbatch"
	"github.com/steveyegge/beads/internal/graphpatch"
)

var (
	ErrInvalid     = errors.New("invalid canonical batch state")
	ErrUnsupported = errors.New("unsupported canonical batch state operation")
	ErrNotFound    = errors.New("batch resource is not live")
	ErrKind        = errors.New("batch resource kind mismatch")
	ErrRevision    = errors.New("batch baseline revision mismatch")
	ErrAllocated   = errors.New("batch identity is already allocated")
	ErrLimit       = errors.New("batch staged selection limit exceeded")
)

// Resource contains covered state and its original committed envelope. Revision
// is a baseline token only: a changed staged record keeps that token, and a new
// record has none. Neither is a published postimage until the publisher assigns
// its actual final revision. Envelope is preserved verbatim for unchanged state.
type Resource struct {
	ID         string
	Kind       graphops.ResourceKind
	Revision   string
	Type       canonicalbatch.Reference
	Properties json.RawMessage
	Metadata   json.RawMessage
	Source     canonicalbatch.Reference
	Target     canonicalbatch.Reference
	Envelope   json.RawMessage
}

// Input is caller-owned. Allocate must only choose an identity, without durable
// effects; the publisher commits reservations with the decision. ResolveType
// resolves an authored affiliation once under the publisher's checked authority.
// CheckPin checks an exact retained committed Bead version, never a staged one.
// Without it only a matching baseline Bead revision can satisfy an endpoint pin.
type Input struct {
	Scope       string
	Current     map[string]Resource
	ReservedIDs []string
	Members     []canonicalbatch.Member
	Allocate    func(index int, kind graphops.ResourceKind) (string, error)
	ResolveType func(canonicalbatch.Reference) (canonicalbatch.Reference, error)
	CheckPin    func(canonicalbatch.Reference) error
	// These are complete checked Type authority, not ancestry inferred from
	// names or the currently populated records. Nil maps mean unavailable.
	TypeKinds    map[string]graphops.ResourceKind
	TypeFamilies map[string][]string
}

type Mapping struct {
	OperationIndex int
	Operation      string
	ID             string
	OperationName  string
}

// SelectionCapture freezes identity membership once. Later uses inspect the
// then-current staged bodies of those same identities, without re-running the
// predicate or silently omitting a captured Resource that was subsequently deleted.
type SelectionCapture struct {
	OperationIndex int
	Name           string
	Kind           graphops.ResourceKind
	IDs            []string
}

type ProvenanceInput struct {
	OperationIndex int
	Value          json.RawMessage
}

// Outcome has exactly one final meaning for a directly targeted identity.
// Before and After are detached copies; allocated has neither. Deleted retains
// the original last-live Before, including its revision, never a staged version.
type Outcome struct {
	ID         string
	Kind       graphops.ResourceKind
	Outcome    string
	Changed    bool
	Before     *Resource
	After      *Resource
	Provenance []ProvenanceInput
}

// Plan is detached from all input. Final includes untouched live records as well
// as staged records; Outcomes only directly targeted identities. The publisher
// must add boundary ownership effects, validate final graph constraints, assign
// actual revisions, and durably retain the bounded result. ConsumedIDs includes
// every new allocation, including creation followed by deletion.
type Plan struct {
	Mappings    []Mapping
	Outcomes    []Outcome
	Final       map[string]Resource
	ConsumedIDs []string
	Provenance  map[string][]ProvenanceInput
	Selections  []SelectionCapture
}

type MemberError struct {
	OperationIndex int
	Err            error
}

func (e *MemberError) Error() string {
	return fmt.Sprintf("batch operation %d: %v", e.OperationIndex, e.Err)
}
func (e *MemberError) Unwrap() error { return e.Err }

type admitted struct {
	member     canonicalbatch.Member
	kind       graphops.ResourceKind
	properties *graphpatch.Patch
	metadata   *graphpatch.Patch
	selection  *selection
}

type capturedBinding struct {
	kind graphops.ResourceKind
	ids  []string
	set  bool
}

type staging struct {
	input            Input
	initial          map[string]Resource
	live             map[string]Resource
	reserved         map[string]bool
	created          map[string]graphops.ResourceKind
	bindings         map[string]capturedBinding
	targets          map[string]graphops.ResourceKind
	provenance       map[string][]ProvenanceInput
	mappings         []Mapping
	selections       []SelectionCapture
	selectionWork    int
	selectionBytes   int
	metadataReplaced map[string]bool
}

// Stage validates the complete member grammar before allocation or staging, then
// applies members in order. On any error it returns a zero Plan and never changes
// the caller's baseline, member byte slices, or retained envelopes.
func Stage(input Input) (Plan, error) {
	if err := graphops.ValidateScopeURL(input.Scope); err != nil {
		return Plan{}, fmt.Errorf("%w: scope: %v", ErrInvalid, err)
	}
	if len(input.Members) == 0 || len(input.Members) > canonicalbatch.MaxOperations {
		return Plan{}, fmt.Errorf("%w: member count", ErrInvalid)
	}
	members := make([]admitted, len(input.Members))
	names := map[string]graphops.ResourceKind{}
	for i, m := range input.Members {
		a, err := admit(m)
		if err != nil {
			return Plan{}, &MemberError{i, err}
		}
		if a.selection != nil {
			if err := a.selection.checkBindings(names); err != nil {
				return Plan{}, &MemberError{i, err}
			}
			if err := a.selection.checkTypes(input); err != nil {
				return Plan{}, &MemberError{i, err}
			}
		}
		for _, operand := range []struct {
			value canonicalbatch.Operand
			kind  graphops.ResourceKind
		}{{m.Bead, graphops.KindBead}, {m.Link, graphops.KindLink}, {m.Source, graphops.KindBead}, {m.Target, graphops.KindBead}} {
			if operand.value.Binding == "" {
				continue
			}
			kind, found := names[operand.value.Binding]
			if !found {
				return Plan{}, &MemberError{i, fmt.Errorf("%w: unknown or forward binding %q", ErrInvalid, operand.value.Binding)}
			}
			if kind != operand.kind {
				return Plan{}, &MemberError{i, fmt.Errorf("%w: binding %q", ErrKind, operand.value.Binding)}
			}
		}
		if m.Name != "" {
			if _, duplicate := names[m.Name]; duplicate {
				return Plan{}, &MemberError{i, fmt.Errorf("%w: duplicate binding %q", ErrInvalid, m.Name)}
			}
			names[m.Name] = a.kind
		}
		members[i] = a
	}
	s := staging{input: input, initial: map[string]Resource{}, live: map[string]Resource{}, reserved: map[string]bool{}, created: map[string]graphops.ResourceKind{}, bindings: map[string]capturedBinding{}, targets: map[string]graphops.ResourceKind{}, provenance: map[string][]ProvenanceInput{}, metadataReplaced: map[string]bool{}}
	for id, r := range input.Current {
		if id != r.ID || !r.Kind.Valid() || r.Revision == "" || r.Type.URI == "" {
			return Plan{}, fmt.Errorf("%w: incomplete baseline resource %q", ErrInvalid, id)
		}
		if err := s.localID(id); err != nil {
			return Plan{}, err
		}
		var err error
		r.Properties, err = object(r.Properties, false)
		if err != nil {
			return Plan{}, fmt.Errorf("%w: baseline properties for %s: %v", ErrInvalid, id, err)
		}
		r.Metadata, err = object(r.Metadata, true)
		if err != nil {
			return Plan{}, fmt.Errorf("%w: baseline metadata for %s: %v", ErrInvalid, id, err)
		}
		s.initial[id] = clone(r)
		s.live[id] = clone(r)
		s.reserved[id] = true
	}
	for _, id := range input.ReservedIDs {
		if err := s.localID(id); err != nil {
			return Plan{}, err
		}
		s.reserved[id] = true
	}
	for i, a := range members {
		if err := s.apply(i, a); err != nil {
			return Plan{}, &MemberError{i, err}
		}
	}
	return s.finish()
}

func admit(m canonicalbatch.Member) (admitted, error) {
	a := admitted{member: m}
	if m.Operation == "select" {
		var err error
		a.selection, err = admitSelection(m)
		if err != nil {
			return a, err
		}
		a.kind = a.selection.kind
		return a, nil
	}
	switch m.Operation {
	case "createBead", "updateBead", "deleteBead":
		a.kind = graphops.KindBead
	case "createLink", "updateLink", "deleteLink":
		a.kind = graphops.KindLink
	default:
		return a, fmt.Errorf("%w: %s", ErrUnsupported, m.Operation)
	}
	if m.Set || m.Query != nil {
		return a, fmt.Errorf("%w: set/read members", ErrUnsupported)
	}
	if m.Kind != "" && m.Kind != a.kind {
		return a, fmt.Errorf("%w: declared member kind", ErrKind)
	}
	create := strings.HasPrefix(m.Operation, "create")
	deleteMember := strings.HasPrefix(m.Operation, "delete")
	if !create && (m.Name != "" || m.ID != "") {
		return a, fmt.Errorf("%w: create-only member", ErrInvalid)
	}
	if create && (m.Type.URI == "" || m.ExpectedRevision != "" || m.Bead.Present() || m.Link.Present()) {
		return a, fmt.Errorf("%w: creation requires a Type and no mutation target/guard", ErrInvalid)
	}
	if !create {
		target := m.Bead
		if a.kind == graphops.KindLink {
			target = m.Link
		}
		if !target.Present() || target.Revision != "" {
			return a, fmt.Errorf("%w: mutation target must select a current identity", ErrInvalid)
		}
		if a.kind == graphops.KindBead && m.Link.Present() || a.kind == graphops.KindLink && m.Bead.Present() {
			return a, fmt.Errorf("%w: wrong target field", ErrInvalid)
		}
	}
	if a.kind == graphops.KindBead && (m.Source.Present() || m.Target.Present()) {
		return a, fmt.Errorf("%w: Bead endpoints", ErrInvalid)
	}
	if create && a.kind == graphops.KindLink && (!m.Source.Present() || !m.Target.Present()) {
		return a, fmt.Errorf("%w: Link endpoints required", ErrInvalid)
	}
	if deleteMember && (m.Type.URI != "" || m.Source.Present() || m.Target.Present() || m.Properties != nil || m.Metadata != nil || m.PropertiesChange != nil || m.MetadataChange != nil) {
		return a, fmt.Errorf("%w: delete has mutation fields", ErrInvalid)
	}
	if m.Properties != nil && m.PropertiesChange != nil || m.Metadata != nil && m.MetadataChange != nil {
		return a, fmt.Errorf("%w: replacement and change are exclusive", ErrInvalid)
	}
	if create && (m.PropertiesChange != nil || m.MetadataChange != nil) {
		return a, fmt.Errorf("%w: create cannot patch missing state", ErrInvalid)
	}
	var err error
	if m.Properties != nil {
		a.member.Properties, err = object(m.Properties, false)
		if err != nil {
			return a, err
		}
	}
	if m.Metadata != nil {
		a.member.Metadata, err = object(m.Metadata, false)
		if err != nil {
			return a, err
		}
	}
	if m.PropertiesChange != nil {
		a.properties, err = graphpatch.Parse(m.PropertiesChange)
		if err != nil {
			return a, err
		}
	}
	if m.MetadataChange != nil {
		a.metadata, err = graphpatch.Parse(m.MetadataChange)
		if err != nil {
			return a, err
		}
	}
	if m.RevisionProvenance != nil {
		a.member.RevisionProvenance, err = object(m.RevisionProvenance, false)
		if err != nil {
			return a, err
		}
	}
	for _, op := range []canonicalbatch.Operand{m.Bead, m.Link, m.Source, m.Target} {
		if op.Binding != "" && op.URI != "" || op.Revision != "" && !op.Present() {
			return a, fmt.Errorf("%w: malformed Reference operand", ErrInvalid)
		}
	}
	return a, nil
}

func (s *staging) localID(id string) error {
	if _, err := graphops.ScopeLocalPublicResourceID(s.input.Scope, id); err != nil {
		return fmt.Errorf("%w: scope-local identity %q: %v", ErrInvalid, id, err)
	}
	return nil
}

func (s *staging) resolve(op canonicalbatch.Operand) (canonicalbatch.Reference, error) {
	id := op.URI
	if op.Binding != "" {
		capture, ok := s.bindings[op.Binding]
		if !ok {
			return canonicalbatch.Reference{}, fmt.Errorf("%w: unknown or forward binding %q", ErrInvalid, op.Binding)
		}
		if capture.set || len(capture.ids) != 1 {
			return canonicalbatch.Reference{}, fmt.Errorf("%w: a set cannot fill a scalar Reference", ErrKind)
		}
		id = capture.ids[0]
	}
	if err := s.localID(id); err != nil {
		return canonicalbatch.Reference{}, err
	}
	return canonicalbatch.Reference{URI: id, Revision: op.Revision}, nil
}

func (s *staging) endpoint(op canonicalbatch.Operand) (canonicalbatch.Reference, error) {
	r, err := s.resolve(op)
	if err != nil {
		return r, err
	}
	if r.Revision != "" {
		if s.input.CheckPin != nil {
			if err := s.input.CheckPin(r); err != nil {
				return r, err
			}
			return r, nil
		}
		base, ok := s.initial[r.URI]
		if !ok || base.Kind != graphops.KindBead || base.Revision != r.Revision {
			return r, fmt.Errorf("%w: exact endpoint is not an acquired committed Bead version", ErrRevision)
		}
		return r, nil
	}
	current, ok := s.live[r.URI]
	if !ok {
		return r, fmt.Errorf("%w: endpoint %s", ErrNotFound, r.URI)
	}
	if current.Kind != graphops.KindBead {
		return r, fmt.Errorf("%w: endpoint %s", ErrKind, r.URI)
	}
	return r, nil
}

func (s *staging) apply(index int, a admitted) error {
	if a.selection != nil {
		return s.selectIDs(index, a)
	}
	if !strings.HasPrefix(a.member.Operation, "create") {
		target := a.member.Bead
		if a.kind == graphops.KindLink {
			target = a.member.Link
		}
		if target.Binding != "" {
			capture, ok := s.bindings[target.Binding]
			if !ok {
				return fmt.Errorf("%w: unknown binding %q", ErrInvalid, target.Binding)
			}
			if capture.kind != a.kind {
				return ErrKind
			}
			if capture.set {
				if a.member.ExpectedRevision != "" && len(capture.ids) != 1 {
					return fmt.Errorf("%w: a guard requires a singleton captured set", ErrRevision)
				}
				for _, id := range capture.ids {
					if err := s.selectionTick(); err != nil {
						return err
					}
					if r, live := s.live[id]; live {
						if err := s.chargeSelectionResource(r); err != nil {
							return err
						}
					}
					single := a
					if a.kind == graphops.KindBead {
						single.member.Bead = canonicalbatch.Operand{URI: id}
					} else {
						single.member.Link = canonicalbatch.Operand{URI: id}
					}
					if err := s.applyOne(index, single); err != nil {
						return err
					}
				}
				return nil
			}
		}
	}
	return s.applyOne(index, a)
}

func (s *staging) applyOne(index int, a admitted) error {
	m := a.member
	var r Resource
	if strings.HasPrefix(m.Operation, "create") {
		id := m.ID
		if id != "" && graphops.ValidateCanonicalResourceID(id) == nil {
			var err error
			id, err = graphops.CanonicalPublicResourceURL(s.input.Scope, id)
			if err != nil {
				return err
			}
		}
		if id == "" {
			if s.input.Allocate == nil {
				return fmt.Errorf("%w: no identity allocator", ErrUnsupported)
			}
			var err error
			id, err = s.input.Allocate(index, a.kind)
			if err != nil {
				return err
			}
		}
		if err := s.localID(id); err != nil {
			return err
		}
		if s.reserved[id] {
			return fmt.Errorf("%w: %s", ErrAllocated, id)
		}
		if m.Name != "" {
			if _, found := s.bindings[m.Name]; found {
				return fmt.Errorf("%w: duplicate binding %s", ErrInvalid, m.Name)
			}
		}
		r = Resource{ID: id, Kind: a.kind, Properties: json.RawMessage(`{}`), Metadata: json.RawMessage(`{}`)}
		s.reserved[id] = true
		s.created[id] = a.kind
	} else {
		target := m.Bead
		if a.kind == graphops.KindLink {
			target = m.Link
		}
		ref, err := s.resolve(target)
		if err != nil {
			return err
		}
		var ok bool
		r, ok = s.live[ref.URI]
		if !ok {
			return fmt.Errorf("%w: %s", ErrNotFound, ref.URI)
		}
		if r.Kind != a.kind {
			return fmt.Errorf("%w: %s", ErrKind, ref.URI)
		}
		if m.ExpectedRevision != "" {
			baseline, exists := s.initial[r.ID]
			if !exists || baseline.Revision != m.ExpectedRevision {
				return fmt.Errorf("%w: %s", ErrRevision, r.ID)
			}
		}
	}
	s.targets[r.ID] = a.kind
	if m.RevisionProvenance != nil {
		s.provenance[r.ID] = append(s.provenance[r.ID], ProvenanceInput{index, bytes.Clone(m.RevisionProvenance)})
	}
	if strings.HasPrefix(m.Operation, "delete") {
		delete(s.live, r.ID)
	} else {
		if m.Type.URI != "" {
			r.Type = m.Type
			if s.input.ResolveType != nil {
				var err error
				r.Type, err = s.input.ResolveType(m.Type)
				if err != nil {
					return err
				}
				if r.Type.URI == "" {
					return fmt.Errorf("%w: empty resolved Type", ErrInvalid)
				}
			}
		}
		var err error
		if m.Source.Present() {
			r.Source, err = s.endpoint(m.Source)
			if err != nil {
				return err
			}
		}
		if m.Target.Present() {
			r.Target, err = s.endpoint(m.Target)
			if err != nil {
				return err
			}
		}
		r.Properties, err = edit(r.Properties, m.Properties, a.properties)
		if err != nil {
			return err
		}
		r.Metadata, err = edit(r.Metadata, m.Metadata, a.metadata)
		if err != nil {
			return err
		}
		if m.Metadata != nil {
			s.metadataReplaced[r.ID] = true
		}
		// An absent legacy object supplies the empty working object for
		// patches, not an instruction to publish a newly present object. This
		// also preserves absence when separate members add then remove values.
		// Explicit object replacement deliberately supplies presence instead.
		if baseline, exists := s.initial[r.ID]; exists && baseline.Metadata == nil && !s.metadataReplaced[r.ID] && bytes.Equal(r.Metadata, []byte(`{}`)) {
			r.Metadata = nil
		}
		s.live[r.ID] = clone(r)
	}
	if m.Name != "" {
		s.bindings[m.Name] = capturedBinding{kind: a.kind, ids: []string{r.ID}}
	}
	s.mappings = append(s.mappings, Mapping{index, m.Operation, r.ID, m.Name})
	return nil
}

func (s *staging) finish() (Plan, error) {
	p := Plan{Mappings: s.mappings, Final: map[string]Resource{}, Provenance: map[string][]ProvenanceInput{}, Selections: s.selections}
	ids := make([]string, 0, len(s.targets))
	for id := range s.targets {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return graphops.CompareCodeUnits(ids[i], ids[j]) < 0 })
	for _, id := range ids {
		before, existed := s.initial[id]
		after, survives := s.live[id]
		o := Outcome{ID: id, Kind: s.targets[id], Provenance: cloneProvenance(s.provenance[id])}
		if existed {
			v := clone(before)
			o.Before = &v
		}
		if survives {
			v := clone(after)
			o.After = &v
		}
		switch {
		case !existed && !survives:
			o.Outcome = "allocated"
			o.Changed = true
		case !existed:
			o.Outcome = "created"
			o.Changed = true
		case !survives:
			o.Outcome = "deleted"
			o.Changed = true
		case equalCovered(before, after):
			o.Outcome = "unchanged"
			v := clone(before)
			o.After = &v
			s.live[id] = clone(before)
		default:
			o.Outcome = "updated"
			o.Changed = true
		}
		if o.Changed && o.After != nil && len(o.Provenance) > 1 {
			for _, supplied := range o.Provenance[1:] {
				if !bytes.Equal(o.Provenance[0].Value, supplied.Value) {
					return Plan{}, &MemberError{supplied.OperationIndex, fmt.Errorf("%w: conflicting revision provenance for %s", ErrInvalid, id)}
				}
			}
		}
		p.Outcomes = append(p.Outcomes, o)
	}
	for id, r := range s.live {
		p.Final[id] = clone(r)
	}
	for id := range s.created {
		p.ConsumedIDs = append(p.ConsumedIDs, id)
	}
	sort.Slice(p.ConsumedIDs, func(i, j int) bool { return graphops.CompareCodeUnits(p.ConsumedIDs[i], p.ConsumedIDs[j]) < 0 })
	for id, inputs := range s.provenance {
		p.Provenance[id] = cloneProvenance(inputs)
	}
	return p, nil
}

func object(raw json.RawMessage, optional bool) (json.RawMessage, error) {
	if raw == nil && optional {
		return nil, nil
	}
	if raw == nil {
		return nil, fmt.Errorf("%w: missing object", ErrInvalid)
	}
	v, err := graphops.CanonicalizeJSON(raw)
	if err != nil {
		return nil, fmt.Errorf("%w: object: %v", ErrInvalid, err)
	}
	if len(v) == 0 || v[0] != '{' {
		return nil, fmt.Errorf("%w: expected object", ErrInvalid)
	}
	return v, nil
}

func edit(current, replacement json.RawMessage, patch *graphpatch.Patch) (json.RawMessage, error) {
	if replacement != nil {
		return bytes.Clone(replacement), nil
	}
	if patch == nil {
		return bytes.Clone(current), nil
	}
	absent := current == nil
	if absent {
		current = json.RawMessage(`{}`)
	}
	properties, err := graphops.NewProperties(current)
	if err != nil {
		return nil, err
	}
	result, err := patch.Apply(properties)
	if err != nil {
		return nil, err
	}
	next := result.Bytes()
	if absent && bytes.Equal(next, []byte(`{}`)) {
		return nil, nil
	}
	return next, nil
}

func equalCovered(a, b Resource) bool {
	return a.ID == b.ID && a.Kind == b.Kind && a.Type == b.Type && bytes.Equal(a.Properties, b.Properties) && bytes.Equal(a.Metadata, b.Metadata) && a.Source == b.Source && a.Target == b.Target
}

func clone(r Resource) Resource {
	r.Properties = bytes.Clone(r.Properties)
	r.Metadata = bytes.Clone(r.Metadata)
	r.Envelope = bytes.Clone(r.Envelope)
	return r
}

func cloneProvenance(in []ProvenanceInput) []ProvenanceInput {
	if in == nil {
		return nil
	}
	out := make([]ProvenanceInput, len(in))
	for i, v := range in {
		out[i] = ProvenanceInput{v.OperationIndex, bytes.Clone(v.Value)}
	}
	return out
}
