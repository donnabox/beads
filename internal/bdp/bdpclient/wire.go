package bdpclient

import (
	"encoding/json"

	"github.com/steveyegge/beads/internal/bdp"
)

// This file is the impedance layer between the spec-faithful bdp.Scope types and
// the clean-room server's wire JSON (pinned at bdp.lumen b555931c), the dual of
// the Rust clean-room's wire.rs. The wire is a SPIKE that diverges from ratified
// core-v0.2 in a few places the translations below absorb and document:
//
//   - A representation is EXTERNALLY tagged ({"Bead":{…}} / {"Generic":{…}}).
//     core-v0.2 discriminates a bead structurally, with no envelope tag — the
//     clean-room's own wire.rs admits "a conformant decoder wouldn't see this
//     wrapper". The spike chose it so serde could round-trip raw-JSON bodies.
//   - A bead carries a top-level `kind` string. In core-v0.2 kind is not stored;
//     it is the type path's id. The wire `kind` is the type's leaf ("task").
//   - The firewalled payload is `properties` on the wire; core-v0.2 calls it
//     `body`. Same bytes, different name.
//   - A write's minted rev rides in a quoted `ETag` response header, not the body.
//   - The Scope uri arm is carried as the token "web" below. That is an
//     UNVERIFIED spike assumption read off the clean-room source during M2 — the
//     green smoke test only exercises local-scope (absent) ids, so the uri arm
//     has never round-tripped. The ratified decoder ring (core.lumen §1.1) names
//     the literal "uri"; verify/retarget this token when the uri arm is first
//     exercised or when targeting a ratified-wire server.
//
// Keeping this translation OUT of package bdp is the point: the contract tracks
// the spec, and a client speaks whatever wire its server speaks.

// wireResourceID mirrors the server's { scope?, id, rev? }.
type wireResourceID struct {
	Scope *string `json:"scope,omitempty"`
	ID    string  `json:"id"`
	Rev   *string `json:"rev,omitempty"`
}

func toWireID(r bdp.ResourceID) wireResourceID {
	w := wireResourceID{ID: r.ID}
	switch {
	case r.Namespace.URI:
		web := "web" // spike token for the uri arm (unverified; ratified §1.1 literal is "uri") — see file header
		w.Scope = &web
	case r.Namespace.Store != "":
		s := r.Namespace.Store
		w.Scope = &s
	}
	if r.Rev != "" {
		rev := string(r.Rev)
		w.Rev = &rev
	}
	return w
}

func (w wireResourceID) toBDP() bdp.ResourceID {
	r := bdp.ResourceID{ID: w.ID}
	if w.Scope != nil {
		switch *w.Scope {
		case "web": // spike token for the uri arm (unverified; ratified §1.1 literal is "uri")
			r.Namespace = bdp.Namespace{URI: true}
		default:
			r.Namespace = bdp.Namespace{Store: *w.Scope}
		}
	}
	if w.Rev != nil {
		r.Rev = bdp.Rev(*w.Rev)
	}
	return r
}

// wireCreateRequest is the body of create and put: { kind, type, id?, properties?, metadata? }.
type wireCreateRequest struct {
	Kind       string          `json:"kind"`
	Type       wireResourceID  `json:"type"`
	ID         *wireResourceID `json:"id,omitempty"`
	Properties json.RawMessage `json:"properties,omitempty"`
	Metadata   json.RawMessage `json:"metadata,omitempty"`
}

// wireProvenance mirrors the server's provenance block.
type wireProvenance struct {
	Principals       []json.RawMessage `json:"principals"`
	Inception        string            `json:"inception"`
	CommittedAt      string            `json:"committedAt"`
	PreviousRevision *string           `json:"previousRevision,omitempty"`
	Message          *string           `json:"message,omitempty"`
}

func (w wireProvenance) toBDP() *bdp.Provenance {
	pr := &bdp.Provenance{
		Inception:   bdp.Inception(w.Inception),
		CommittedAt: w.CommittedAt,
		Message:     deref(w.Message),
	}
	for _, p := range w.Principals {
		pr.Principals = append(pr.Principals, p)
	}
	if w.PreviousRevision != nil {
		pr.PreviousRevision = bdp.Rev(*w.PreviousRevision)
	}
	return pr
}

// wireRepresentation is the externally-tagged get response.
type wireRepresentation struct {
	Bead    *wireBead    `json:"Bead,omitempty"`
	Generic *wireGeneric `json:"Generic,omitempty"`
}

type wireBead struct {
	Kind         string          `json:"kind"`
	ID           wireResourceID  `json:"id"`
	Type         wireResourceID  `json:"type"`
	StorageClass json.RawMessage `json:"storageClass"`
	Properties   json.RawMessage `json:"properties"`
	Metadata     json.RawMessage `json:"metadata"`
	Provenance   wireProvenance  `json:"provenance"`
	CreatedAt    string          `json:"createdAt"`
}

type wireGeneric struct {
	Kind        string          `json:"kind"`
	ID          wireResourceID  `json:"id"`
	ContentType string          `json:"contentType"`
	Body        json.RawMessage `json:"body"`
	Provenance  wireProvenance  `json:"provenance"`
}

// toBDP folds a wire representation back into the spec-faithful Resource.
// Lossy by design — the friction is the finding (see file header):
//   - Bead: wire `kind` ("task") is the closest thing to the spec type path, and
//     the wire `type` ref (the type bead) is carried as Authority; `properties`
//     becomes Body. storageClass is synthesized by the server and dropped.
//   - Generic: content type and raw body map straight across.
func (w wireRepresentation) toBDP() bdp.Representation {
	switch {
	case w.Bead != nil:
		b := w.Bead
		typeRef := b.Type.toBDP()
		return bdp.Resource{
			ID:         b.ID.toBDP(),
			Authority:  &typeRef,
			Type:       b.Kind,
			Body:       nonEmpty(b.Properties),
			Metadata:   nonEmpty(b.Metadata),
			Provenance: b.Provenance.toBDP(),
			CreatedAt:  b.CreatedAt,
		}
	case w.Generic != nil:
		g := w.Generic
		return bdp.Resource{
			ID:         g.ID.toBDP(),
			Type:       g.ContentType,
			Body:       nonEmpty(g.Body),
			Provenance: g.Provenance.toBDP(),
		}
	default:
		return bdp.Resource{}
	}
}

// wireProblem is the RFC 9457 error body.
type wireProblem struct {
	Type   string  `json:"type"`
	Title  string  `json:"title"`
	Status int     `json:"status"`
	Code   string  `json:"code"`
	Retry  string  `json:"retry"`
	Detail *string `json:"detail,omitempty"`
}

func (w wireProblem) toProblem() bdp.Problem {
	return bdp.Problem{
		Type:   w.Type,
		Title:  w.Title,
		Status: w.Status,
		Code:   w.Code,
		Retry:  w.Retry,
		Detail: deref(w.Detail),
	}
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func nonEmpty(r json.RawMessage) json.RawMessage {
	if len(r) == 0 {
		return nil
	}
	return r
}
