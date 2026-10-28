// Package resourceget exposes a current Resource read role for the shipping
// bad780/schema-v0 profile. It does not advertise the newer BDP v0.1 contract,
// History, authorization, or completeness beyond the source's checked read.
package resourceget

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/httpapi/bdpwire"
)

// Kind is a requested or observed Resource kind. Any is only a request value;
// a successfully constructed NativeDocument always has a concrete kind.
type Kind uint8

const (
	Any Kind = iota
	Bead
	Link
)

// Request selects a current Resource using a Scope-local shipping selector or
// an exact canonical URL in the reader's Scope. Local selection preserves the
// source's exact-ID-first legacy shorthand rules; the result reports the actual
// canonical identity. This operation never parses CLI revision syntax.
type Request struct {
	ID         string
	AssertKind Kind
}

// Reader is an optional current read role. Its acquisition and lifetime belong
// to its provider. A successful read includes the source's complete owned state.
type Reader interface {
	Get(context.Context, Request) (NativeDocument, error)
}

var (
	ErrInvalidRequest  = errors.New("invalid current Resource request")
	ErrInvalidDocument = errors.New("invalid current native Resource document")
	ErrInvalidSource   = errors.New("invalid current Resource source")
	ErrNotFound        = errors.New("current Resource not found")
	ErrGone            = errors.New("current Resource was deleted")
	ErrKindMismatch    = errors.New("current Resource kind does not match request")
	ErrClosed          = errors.New("current Resource source is closed")
)

// Contract identifies the representation actually carried, not a claim about
// the provider's conformance, authority, or the caller's access rights.
type Contract struct {
	SpecCommit string
	SchemaID   string
	BDPVersion string
}

// NativeDocument is an immutable validated current native record. It contains
// no transport envelope. The zero value is invalid and cannot be marshaled.
type NativeDocument struct {
	scope, id, revision string
	kind                Kind
	native              string
}

func (d NativeDocument) Contract() Contract {
	if d.native == "" {
		return Contract{}
	}
	return Contract{SpecCommit: bdpwire.Pin, SchemaID: bdpwire.SchemaID, BDPVersion: bdpwire.BDPVersion}
}

func (d NativeDocument) ScopeURL() string { return d.scope }
func (d NativeDocument) Kind() Kind       { return d.kind }
func (d NativeDocument) ID() string       { return d.id }
func (d NativeDocument) Revision() string { return d.revision }

// NativeJSON returns an independent byte copy on every call.
func (d NativeDocument) NativeJSON() json.RawMessage { return json.RawMessage(d.native) }

// MarshalJSON emits the native record itself, preserving whole References.
func (d NativeDocument) MarshalJSON() ([]byte, error) {
	if d.native == "" {
		return nil, ErrInvalidDocument
	}
	return []byte(d.native), nil
}

// UnmarshalJSON refuses decoding without the trusted Scope and expected kind.
// Use ParseLegacyCurrent explicitly; generic decoding must not silently produce
// an invalid zero document or overwrite an already checked document.
func (*NativeDocument) UnmarshalJSON([]byte) error {
	return fmt.Errorf("%w: use ParseLegacyCurrent with explicit Scope and kind", ErrInvalidDocument)
}

// ParseLegacyCurrent validates a complete current Bead or Link representation
// under the shipping profile, including required metadata. It checks identity
// against scope but performs no source read, Type lookup or authorization.
// Aggregate links are not part of this full-current Get carrier. Explicit empty
// ownership buckets from the shipping profile are retained byte-for-byte.
// Callers must bound untrusted input before parsing. NativeJSON preserves input
// formatting; encoding/json may compact or escape MarshalJSON's native output.
func ParseLegacyCurrent(scope string, kind Kind, native []byte) (NativeDocument, error) {
	fail := func(err error) (NativeDocument, error) {
		return NativeDocument{}, fmt.Errorf("%w: %w", ErrInvalidDocument, err)
	}
	if err := graph.ValidateScopeURL(scope); err != nil {
		return fail(err)
	}
	var id, revision string
	switch kind {
	case Bead:
		var v bdpwire.CurrentBeadRecord
		if err := bdpwire.Unmarshal(native, &v); err != nil {
			return fail(err)
		}
		if err := validateRecord(scope, v.ID, v.Type, v.Revision, v.Attribution, v.ChangeContext); err != nil {
			return fail(err)
		}
		if v.Links != nil {
			return fail(errors.New("aggregate links are not a current Get record"))
		}
		seen := map[string]bool{}
		for typ, links := range v.OwnedLinks {
			if err := graph.ValidateTypeURL(typ); err != nil {
				return fail(err)
			}
			for _, link := range links {
				if err := validateLink(scope, link); err != nil {
					return fail(err)
				}
				if link.Type != typ || link.Source.URI != v.ID || seen[link.ID] {
					return fail(errors.New("owned Link grouping, source or identity is inconsistent"))
				}
				seen[link.ID] = true
			}
		}
		id, revision = v.ID, v.Revision
	case Link:
		var v bdpwire.CurrentLinkRecord
		if err := bdpwire.Unmarshal(native, &v); err != nil {
			return fail(err)
		}
		if err := validateLink(scope, v); err != nil {
			return fail(err)
		}
		id, revision = v.ID, v.Revision
	default:
		return fail(errors.New("a concrete Bead or Link kind is required"))
	}
	return NativeDocument{scope: scope, id: id, revision: revision, kind: kind, native: string(native)}, nil
}

func validateRecord(scope, id, typ, revision string, attribution *bdpwire.Attribution, change *bdpwire.ChangeContext) error {
	if _, err := graph.ScopeLocalPublicResourceID(scope, id); err != nil {
		return err
	}
	if err := graph.ValidateTypeURL(typ); err != nil {
		return err
	}
	if revision == "" {
		return errors.New("revision must be nonempty")
	}
	if attribution != nil && (attribution.Principal == "" || !attribution.Basis.Valid()) {
		return errors.New("invalid attribution")
	}
	if change != nil {
		return change.Validate()
	}
	return nil
}

func validateLink(scope string, v bdpwire.CurrentLinkRecord) error {
	if err := validateRecord(scope, v.ID, v.Type, v.Revision, v.Attribution, v.ChangeContext); err != nil {
		return err
	}
	for _, ref := range []bdpwire.Reference{v.Source, v.Target} {
		if err := graph.ValidateReferenceURI(ref.URI); err != nil {
			return err
		}
		if strings.HasPrefix(ref.URI, scope) {
			if _, err := graph.ScopeLocalPublicResourceID(scope, ref.URI); err != nil {
				return err
			}
		}
	}
	return nil
}
