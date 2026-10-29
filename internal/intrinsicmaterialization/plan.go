// Package intrinsicmaterialization assembles uncommitted local candidates from
// one checked intrinsic publication. It performs no allocation or installation.
package intrinsicmaterialization

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/intrinsictypecatalog"
)

// ErrPlan identifies invalid or inconsistent candidate planning data.
var ErrPlan = errors.New("invalid intrinsic materialization plan")

// Reference is an exact edition; revision tokens remain opaque.
type Reference = intrinsictypecatalog.Reference

// Role identifies a member of the fixed four-definition/two-declaration set.
type Role string

const (
	Type                Role = "type"
	BeadType            Role = "bead-type"
	LinkType            Role = "link-type"
	Refines             Role = "refines"
	BeadTypeRefinesType Role = "bead-type-refines-type"
	LinkTypeRefinesType Role = "link-type-refines-type"
	namespace                = "https://github.com/gastownhall/bdp/types/type-system/"
	semanticSource           = "c6dfb3c05fe707cf605030c5661123e8dda0d8f8"
	contextEncoding          = "intrinsic-materialization-plan/1"
	// MaxContextBytes is a private planning budget, not a Resource ID grammar cap.
	MaxContextBytes = 1 << 20
)

// Scope explicitly binds candidates to a native authority and opaque context.
// Neither this binding nor a successful Build proves that authority was granted.
type Scope struct {
	URI         string `json:"uri"`
	AuthorityID string `json:"authorityID"`
	ContextID   string `json:"contextID"`
}

// Allocation is a candidate from a future controlled allocator, not a receipt.
// PublicID is Scope-local; PrivatePath is an independent native rooted locator.
type Allocation struct {
	Role        Role               `json:"role"`
	PublicID    string             `json:"publicID"`
	PrivatePath string             `json:"privatePath"`
	Kind        graph.ResourceKind `json:"kind"`
	Local       Reference          `json:"local"`
	AuthorityID string             `json:"authorityID"`
}

// Binding retains the unmodified shared record separately from the local body.
// Returned records are independent JSON values; they carry no installed status.
type Binding struct {
	Allocation Allocation      `json:"allocation"`
	Global     Reference       `json:"global"`
	Original   json.RawMessage `json:"original"`
	Record     json.RawMessage `json:"record"`
}

type contextPayload struct {
	Encoding       string                      `json:"encoding"`
	SemanticSource string                      `json:"semanticSource"`
	ArtifactSource intrinsictypecatalog.Source `json:"artifactSource"`
	Scope          Scope                       `json:"scope"`
	Bindings       []Binding                   `json:"bindings"`
}
type envelope struct {
	Payload contextPayload `json:"payload"`
	SHA256  string         `json:"sha256"`
}

// Plan is an immutable, uncommitted candidate context. Its zero value is invalid.
type Plan struct{ payload contextPayload }

type member struct {
	role     Role
	revision string
	kind     graph.ResourceKind
}

func members() [6]member {
	return [6]member{{Type, "type-1", graph.KindBead}, {BeadType, "bead-type-1", graph.KindBead},
		{LinkType, "link-type-1", graph.KindBead}, {Refines, "refines-1", graph.KindBead},
		{BeadTypeRefinesType, "declaration-1", graph.KindLink}, {LinkTypeRefinesType, "declaration-1", graph.KindLink}}
}
func invalid(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrPlan, fmt.Sprintf(format, args...))
}
func digest(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }

// Build assembles six complete candidate bodies from the fixed checked catalog.
// No caller-supplied digest authenticates a catalog. Shared semantic identities
// and schema selectors remain global; mirror endpoints alone become local.
func Build(c *intrinsictypecatalog.Catalog, scope Scope, allocations []Allocation) (*Plan, error) {
	if c == nil {
		return nil, invalid("missing checked catalog")
	}
	source := intrinsictypecatalog.SelectedSource()
	raw := c.Bytes()
	if len(raw) != source.Bytes || digest(raw) != source.SHA256 {
		return nil, invalid("unselected catalog")
	}
	if err := graph.ValidatePersistedScopeURL(scope.URI); err != nil {
		return nil, invalid("Scope: %v", err)
	}
	if len(scope.AuthorityID) != 32 || strings.Trim(scope.AuthorityID, "0123456789abcdef") != "" {
		return nil, invalid("authority must be 32 lowercase hexadecimal characters")
	}
	if _, err := graph.NewRevision(scope.ContextID); err != nil {
		return nil, invalid("context: %v", err)
	}
	if len(allocations) != 6 {
		return nil, invalid("require exactly six candidate allocations")
	}
	// Bound aggregate input before producing duplicated full record captures.
	budget := len(scope.URI) + len(scope.AuthorityID) + len(scope.ContextID)
	byRole := make(map[Role]Allocation, 6)
	identities := make(map[string]bool, 6)
	locators := make(map[string]bool, 6)
	for _, a := range allocations {
		budget += len(a.Role) + len(a.PublicID) + len(a.PrivatePath) + len(a.Kind) + len(a.Local.URI) + len(a.Local.Revision) + len(a.AuthorityID)
		if budget > MaxContextBytes {
			return nil, invalid("candidate context exceeds planning budget")
		}
		var selected *member
		for _, m := range members() {
			if m.role == a.Role {
				copy := m
				selected = &copy
				break
			}
		}
		if selected == nil || byRole[a.Role].Role != "" {
			return nil, invalid("unknown or duplicate role %q", a.Role)
		}
		if a.Kind != selected.kind {
			return nil, invalid("role %q has wrong Resource kind", a.Role)
		}
		if a.AuthorityID != scope.AuthorityID {
			return nil, invalid("role %q has inconsistent authority", a.Role)
		}
		uri, err := graph.CanonicalPublicResourceURL(scope.URI, a.PublicID)
		if err != nil || uri != a.Local.URI {
			return nil, invalid("role %q public identity does not match Scope and public ID", a.Role)
		}
		if err := graph.ValidatePath(a.PrivatePath, a.Kind); err != nil {
			return nil, invalid("role %q private locator: %v", a.Role, err)
		}
		if _, err := graph.NewRevision(a.Local.Revision); err != nil {
			return nil, invalid("role %q revision: %v", a.Role, err)
		}
		for _, m := range members() {
			if a.Local.URI == namespace+string(m.role) {
				return nil, invalid("candidate would rebind shared publication identity")
			}
		}
		if identities[a.Local.URI] || locators[a.PrivatePath] {
			return nil, invalid("duplicate candidate identity or private locator")
		}
		identities[a.Local.URI] = true
		locators[a.PrivatePath] = true
		byRole[a.Role] = a
	}
	originals := make(map[string]json.RawMessage, 6)
	for _, r := range append(c.Definitions(), c.Declarations()...) {
		var id struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(r, &id); err != nil {
			return nil, invalid("checked source record: %v", err)
		}
		originals[id.ID] = r
	}
	p := &Plan{payload: contextPayload{contextEncoding, semanticSource, source, scope, make([]Binding, 0, 6)}}
	records := make(map[Role]json.RawMessage, 6)
	// Real local Links are constructed first, then captured under final owners.
	for _, role := range []Role{BeadTypeRefinesType, LinkTypeRefinesType} {
		fields, err := recordFields(originals[namespace+string(role)])
		if err != nil {
			return nil, err
		}
		a := byRole[role]
		put(fields, "id", a.Local.URI)
		put(fields, "revision", a.Local.Revision)
		owner := BeadType
		if role == LinkTypeRefinesType {
			owner = LinkType
		}
		put(fields, "source", byRole[owner].Local.URI)
		put(fields, "target", byRole[Type].Local)
		records[role], err = json.Marshal(fields)
		if err != nil {
			return nil, invalid("local Link: %v", err)
		}
	}
	for _, role := range []Role{Type, BeadType, LinkType, Refines} {
		fields, err := recordFields(originals[namespace+string(role)])
		if err != nil {
			return nil, err
		}
		a := byRole[role]
		put(fields, "id", a.Local.URI)
		put(fields, "revision", a.Local.Revision)
		owned := map[string][]json.RawMessage{}
		if role == BeadType {
			owned[namespace+string(Refines)] = []json.RawMessage{records[BeadTypeRefinesType]}
		}
		if role == LinkType {
			owned[namespace+string(Refines)] = []json.RawMessage{records[LinkTypeRefinesType]}
		}
		put(fields, "ownedLinks", owned)
		records[role], err = json.Marshal(fields)
		if err != nil {
			return nil, invalid("local Bead: %v", err)
		}
	}
	for _, m := range members() {
		p.payload.Bindings = append(p.payload.Bindings, Binding{byRole[m.role], Reference{URI: namespace + string(m.role), Revision: m.revision}, originals[namespace+string(m.role)], records[m.role]})
	}
	if _, err := p.EncodeContext(); err != nil {
		return nil, err
	}
	return p, nil
}

func recordFields(raw json.RawMessage) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage
	if len(raw) == 0 {
		return nil, invalid("missing selected original record")
	}
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, invalid("original record: %v", err)
	}
	return fields, nil
}

// Values passed to put are closed assembler types with valid UTF-8. JSON errors
// are impossible here; full record/payload marshaling remains checked by callers.
func put(fields map[string]json.RawMessage, key string, value any) {
	raw, _ := json.Marshal(value)
	fields[key] = raw
}
func clone(b Binding) Binding {
	b.Original = append(json.RawMessage(nil), b.Original...)
	b.Record = append(json.RawMessage(nil), b.Record...)
	return b
}

// Scope returns the explicitly bound candidate context without authentication.
func (p *Plan) Scope() Scope {
	if p == nil {
		return Scope{}
	}
	return p.payload.Scope
}

// Bindings returns all six roles in fixed publication order with independent bytes.
func (p *Plan) Bindings() []Binding {
	if p == nil {
		return nil
	}
	out := make([]Binding, len(p.payload.Bindings))
	for i, b := range p.payload.Bindings {
		out[i] = clone(b)
	}
	return out
}

// LookupLocal selects only an exact local edition in this immutable context.
func (p *Plan) LookupLocal(ref Reference) (Binding, bool) {
	if p != nil {
		for _, b := range p.payload.Bindings {
			if b.Allocation.Local == ref {
				return clone(b), true
			}
		}
	}
	return Binding{}, false
}

// LocalFor selects the exact local candidate for a shared pair in this context.
// It is not a generic Type alias or current/default lookup.
func (p *Plan) LocalFor(ref Reference) (Reference, bool) {
	if p != nil {
		for _, b := range p.payload.Bindings {
			if b.Global == ref {
				return b.Allocation.Local, true
			}
		}
	}
	return Reference{}, false
}

// EncodeContext encodes private candidate planning data, never installed state.
func (p *Plan) EncodeContext() ([]byte, error) {
	if p == nil || p.payload.Encoding != contextEncoding || len(p.payload.Bindings) != 6 {
		return nil, invalid("uninitialized plan")
	}
	payload, err := canonicalPayload(p.payload)
	if err != nil {
		return nil, invalid("encode payload: %v", err)
	}
	out, err := json.Marshal(envelope{p.payload, digest(payload)})
	if err != nil {
		return nil, invalid("encode context: %v", err)
	}
	if len(out) > MaxContextBytes {
		return nil, invalid("encoded context exceeds planning budget")
	}
	return out, nil
}

// DecodeContext strictly reconstructs an uncommitted plan using checked source.
// A checksum is only corruption detection: every record and binding must also
// equal the assembler's result. A coherent new candidate is not an allocation.
func DecodeContext(c *intrinsictypecatalog.Catalog, raw []byte) (*Plan, error) {
	if len(raw) > MaxContextBytes {
		return nil, invalid("encoded context exceeds planning budget")
	}
	canonical, err := graph.CanonicalizeJSON(raw)
	if err != nil {
		return nil, invalid("context JSON: %v", err)
	}
	var env envelope
	dec := json.NewDecoder(bytes.NewReader(canonical))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&env); err != nil {
		return nil, invalid("context shape: %v", err)
	}
	payload, err := canonicalPayload(env.Payload)
	if err != nil {
		return nil, invalid("context payload: %v", err)
	}
	if env.SHA256 != digest(payload) {
		return nil, invalid("context checksum mismatch")
	}
	if len(env.Payload.Bindings) != 6 {
		return nil, invalid("require exactly six candidate allocations")
	}
	allocations := make([]Allocation, len(env.Payload.Bindings))
	for i, b := range env.Payload.Bindings {
		allocations[i] = b.Allocation
	}
	p, err := Build(c, env.Payload.Scope, allocations)
	if err != nil {
		return nil, err
	}
	expected, err := p.EncodeContext()
	if err != nil {
		return nil, err
	}
	expected, err = graph.CanonicalizeJSON(expected)
	if err != nil {
		return nil, invalid("assembled context JSON: %v", err)
	}
	// This also refuses case-insensitive struct matches, omitted fields, reordered
	// bindings and recomputed checksums on altered original/local/source records.
	if !bytes.Equal(canonical, expected) {
		return nil, invalid("context differs from checked source and assembled bindings")
	}
	return p, nil
}

func canonicalPayload(payload contextPayload) ([]byte, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return graph.CanonicalizeJSON(raw)
}
