package resourceget

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/steveyegge/beads/graphops"
)

// TypeReference is an exact affiliation carried by a complete Resource. Its
// presence proves representation, not authorization or a checked definition.
type TypeReference struct {
	URI      string `json:"uri"`
	Revision string `json:"revision"`
}

// ResourceDocument carries a complete selected common Resource without
// projecting a native Issue into Memory or losing its exact Type edition.
// Storage must check allocation, backing, definition and retained correspondence
// before construction. The zero value is invalid.
type ResourceDocument struct {
	scope       string
	kind        graphops.ResourceKind
	id          string
	revision    string
	affiliation TypeReference
	native      []byte
	definition  *DefinitionDocument
}

// NewResourceDocument validates representation only. It never infers a
// definition category, installed status or origin from an ID or body shape.
func NewResourceDocument(scope string, kind graphops.ResourceKind, native []byte) (ResourceDocument, error) {
	if kind != graphops.KindBead && kind != graphops.KindLink {
		return ResourceDocument{}, fmt.Errorf("invalid Resource kind %q", kind)
	}
	if err := graphops.ValidateScopeURL(scope); err != nil {
		return ResourceDocument{}, fmt.Errorf("Resource Scope: %w", err)
	}
	canonical, err := graphops.CanonicalizeJSON(native)
	if err != nil {
		return ResourceDocument{}, err
	}
	fields, err := definitionObject(canonical)
	if err != nil {
		return ResourceDocument{}, err
	}
	// Representation only: storage's checked exact Type determines ownership.
	// Absence stays absent; this constructor never manufactures an owned bucket.
	if err := validateCompleteResourceRecord(scope, DefinitionKind(kind), fields, false); err != nil {
		return ResourceDocument{}, err
	}
	id, _ := definitionString(fields["id"])
	revision, _ := definitionString(fields["revision"])
	var affiliation TypeReference
	if err := json.Unmarshal(fields["type"], &affiliation); err != nil {
		return ResourceDocument{}, err
	}
	return ResourceDocument{scope: scope, kind: kind, id: id, revision: revision, affiliation: affiliation, native: bytes.Clone(canonical)}, nil
}

func (d ResourceDocument) ID() string                  { return d.id }
func (d ResourceDocument) Revision() string            { return d.revision }
func (d ResourceDocument) ScopeURL() string            { return d.scope }
func (d ResourceDocument) Kind() graphops.ResourceKind { return d.kind }
func (d ResourceDocument) Type() TypeReference         { return d.affiliation }
func (d ResourceDocument) NativeJSON() []byte          { return bytes.Clone(d.native) }

func (d ResourceDocument) MarshalJSON() ([]byte, error) {
	if len(d.native) == 0 {
		return nil, fmt.Errorf("uninitialized Resource document")
	}
	return d.NativeJSON(), nil
}

// Resource preserves a definition reader's already-carried display provenance
// while exposing the same complete Resource bytes to the ordinary inventory.
func (d DefinitionDocument) Resource() (ResourceDocument, error) {
	r, err := NewResourceDocument(d.scope, graphops.ResourceKind(d.kind), d.native)
	if err != nil {
		return ResourceDocument{}, err
	}
	copy := d
	copy.native = bytes.Clone(d.native)
	if d.origin != nil {
		origin := *d.origin
		copy.origin = &origin
	}
	r.definition = &copy
	return r, nil
}

// Definition reports an explicit definition-reader conversion, never a shape
// test. Callers must not interpret this display information as authorization.
func (d ResourceDocument) Definition() (DefinitionDocument, bool) {
	if d.definition == nil {
		return DefinitionDocument{}, false
	}
	copy := *d.definition
	copy.native = bytes.Clone(copy.native)
	if copy.origin != nil {
		origin := *copy.origin
		copy.origin = &origin
	}
	return copy, true
}
