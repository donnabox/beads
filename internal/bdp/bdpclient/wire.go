package bdpclient

import (
	"encoding/json"
	"strings"

	"github.com/steveyegge/beads/internal/bdp"
)

// This file is the impedance layer between the spec-faithful bdp.Scope types and
// the clean-room server's wire JSON, the dual of the Rust clean-room's wire.rs.
// It is aligned to and LIVE-VALIDATED against bdp-cleanroom-v023 @ 5fba7d23 (the
// rc2 server): the read/write DTOs here are the rc2 §rep / §create shapes, no
// longer the earlier spike envelope. Notes the layer still carries:
//
//   - A write's minted rev rides in a quoted `ETag` response header, not the body
//     (delete's body is also the self-describing { "rev": … }).
//   - The Scope uri arm is carried as the token "web" below. That is an
//     UNVERIFIED assumption read off the clean-room source — the smoke test only
//     exercises local-scope (absent) ids, so the uri arm has never round-tripped.
//     The ratified decoder ring (core.lumen §1.1) names the literal "uri";
//     verify/retarget when the uri arm is first exercised.
//
// (The discovery DTOs — ServiceDescriptor and /sys/capabilities — live in
// discovery.go, which documents the clean-room-vs-ratified SHAPE GAP it absorbs.)
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

// wireCreateBody is the body of both create and put (core.lumen §create/§put at
// bdp-core-0.2.3-rc2). The clean-room server decodes it with
// deny_unknown_fields, so it carries EXACTLY these members and nothing else:
//
//   - `type` is the type PATH string ("issue/task"), not a ref object.
//   - `body` is the representation body (the rc2 name; the old spike called it
//     "properties").
//   - `authority` is the optional type-naming package ref.
//
// The proposed id (create) and the write guards (If-Match / If-None-Match /
// ?precondition=) ride the request line, NEVER the body — so the body stays
// exactly the representation. `retired` is service-managed and refused on create,
// so it is never sent.
type wireCreateBody struct {
	Authority *wireResourceID `json:"authority,omitempty"`
	Type      string          `json:"type"`
	Body      json.RawMessage `json:"body"`
	Metadata  json.RawMessage `json:"metadata,omitempty"`
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

// wireRep is the GET response — the rc2 §rep envelope, FLAT (no Bead/Generic
// wrapper; the old spike wrapped it, which the clean-room's own wire.rs noted "a
// conformant decoder wouldn't see"). It maps almost 1:1 onto bdp.Resource:
// `type` is the type PATH string, `body` is the representation body,
// `storageClass` is a bare lowercase string, and `retired` marks a tombstone.
type wireRep struct {
	ID           wireResourceID  `json:"id"`
	Authority    *wireResourceID `json:"authority,omitempty"`
	Type         string          `json:"type"`
	Body         json.RawMessage `json:"body"`
	Metadata     json.RawMessage `json:"metadata,omitempty"`
	Provenance   *wireProvenance `json:"provenance,omitempty"`
	CreatedAt    string          `json:"createdAt,omitempty"`
	StorageClass string          `json:"storageClass,omitempty"`
	Retired      bool            `json:"retired,omitempty"`
}

func (w wireRep) toBDP() bdp.Representation {
	r := bdp.Resource{
		ID:           w.ID.toBDP(),
		Type:         w.Type,
		Body:         nonEmpty(w.Body),
		Metadata:     nonEmpty(w.Metadata),
		CreatedAt:    w.CreatedAt,
		StorageClass: bdp.StorageClass(w.StorageClass),
		Retired:      w.Retired,
	}
	if w.Authority != nil {
		a := w.Authority.toBDP()
		r.Authority = &a
	}
	if w.Provenance != nil {
		r.Provenance = w.Provenance.toBDP()
	}
	return r
}

// wirePage is the clean-room's Page envelope for a streaming/snapshot read
// (core.lumen §Page): items plus the resume cursor (serialized null, not
// omitted, when there is nothing to resume) and the live-edge marker. For list
// the item type is wireResourceID — the clean-room's list returns IDS, not full
// representations.
type wirePage[T any] struct {
	Items  []T     `json:"items"`
	Cursor *string `json:"cursor"`
	Head   bool    `json:"head"`
}

// wireCondition is the clean-room's kind-tagged list filter condition (the
// `where=` array). This client emits the Compare kind (the low-bar case bd
// uses): a root-addressed Field path (core.lumen §1243, e.g. "body/status" →
// ["body","status"]), a compare Op ("eq"/"ne"/"lt"/"le"/"gt"/"ge"), and the
// value. The server AND-conjoins the conditions.
type wireCondition struct {
	Kind  string          `json:"kind"`
	Field []string        `json:"field"`
	Op    string          `json:"op"`
	Value json.RawMessage `json:"value"`
}

// toWireConditions lowers a bdp.ListFilter to the clean-room `where=` array.
func toWireConditions(where bdp.ListFilter) []wireCondition {
	out := make([]wireCondition, 0, len(where))
	for _, c := range where {
		out = append(out, wireCondition{
			Kind:  "compare",
			Field: splitFieldPath(c.Field),
			Op:    c.Op,
			Value: c.Value,
		})
	}
	return out
}

func splitFieldPath(p string) []string {
	p = strings.Trim(p, "/")
	if p == "" {
		return nil
	}
	return strings.Split(p, "/")
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
