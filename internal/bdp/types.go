package bdp

import "encoding/json"

// This file renders the BDP Core value vocabulary the primitive signatures
// reach. It is a mechanical translation of core/core.lumen (spec/core-v0.2),
// go-ified where the Rust clean-room's choices were Rust-specific:
//
//   - The phantom type parameter on ref<T>/rep<T> is dropped. The spec marks T
//     "checker-only, erased on the wire", so a Go ResourceID carries no T.
//   - Opaque bodies/metadata are json.RawMessage (the analog of the clean-room's
//     zero-copy RawJson), not decoded structs — Core does not interpret them.
//   - Interned Arc<str> handles become plain string.
//
// # Identity — revisions, scope, ResourceId

// Rev is an opaque resource-version token. It is compared for exact equality
// only, never parsed or ordered: the revision changes exactly when the
// resource's state changes, and a write that would leave the state unchanged
// mints no new Rev (and produces no change event).
type Rev string

// Digest is a content-addressed Rev (a sha256 or commit hash). The spec rule is
// "Digest is Rev": a digest satisfies any Rev, and there is no Rev-vs-Digest
// split on the wire, so this is an alias, not a distinct type.
type Digest = Rev

// Namespace selects where an id resolves — the spec's Scope union. It is
// exactly one of three cases, and the zero value is the common one:
//
//   - zero value: the local (current) BDP store the caller is talking to.
//   - Store set: another BDP store, named by its URL; the id is relative to it.
//   - URI true: NOT a BDP store — the id is itself an absolute URI whose scheme
//     resolves it (a resource outside any store).
//
// Store and URI are mutually exclusive; a non-empty Store with URI true is
// malformed.
type Namespace struct {
	// Store is the URL of another BDP store the id is relative to. Empty for a
	// local id or a URI id.
	Store string
	// URI reports that the id is an absolute URI resolved outside any store
	// (the spec's literal "uri" scope).
	URI bool
}

// Local reports whether the namespace is the local store (the zero value).
func (n Namespace) Local() bool { return n.Store == "" && !n.URI }

// ResourceID is the identity of a resource: the namespace it lives in, an id
// within that namespace, and an optional revision. On the wire this projects to
// a URL (the store URL joined with the id); in Go it stays structured.
//
// A ResourceID is also a reference (the spec's ref<T> is just a ResourceId). A
// ResourceID with Rev set is pinned to that exact, immutable revision; with Rev
// empty it is unpinned and floats to the resource's current revision. The
// distinction matters after a delete: a pinned reference still resolves from
// history, while an unpinned one follows the resource to its tombstone.
type ResourceID struct {
	Namespace Namespace
	// ID is the id relative to Namespace (or the absolute URI when
	// Namespace.URI is true).
	ID string
	// Rev pins an exact revision; empty means the current one.
	Rev Rev
}

// Pinned reports whether the reference pins an exact revision.
func (r ResourceID) Pinned() bool { return r.Rev != "" }

// Ref is a by-reference pointer to a resource. The spec's ref<T> is a
// ResourceId with a checker-only, wire-erased T, so in Go a reference is just a
// ResourceID.
type Ref = ResourceID

// # The resource shape

// Resource is the universal shape of a versioned resource — bead or bare
// content. Only ID, Type and Body are always present; the rest is optional.
// This is also what a read returns (the spec's rep<Resource> is a Resource as
// of one revision, possibly projected by a selector).
//
// A bead is a Resource whose Body follows a declared schema and whose Authority,
// when present, is a package; bare content (a text/plain file, an image) is a
// Resource whose Type is a media type and whose Body is raw bytes. The envelope
// is identical for both — domain fields live under Body, never beside ID and
// Type. That firewall is what lets one set of primitives serve every type.
type Resource struct {
	ID ResourceID `json:"id"`
	// Authority names who names Type — a package (bead) or a registry such as
	// IANA (content). Nil means ad-hoc / duck-typed.
	Authority *Authority `json:"authority,omitempty"`
	// Type is the type path, e.g. "issue/task" (a bead kind) or "text/plain"
	// (a media type), resolved within Authority's namespace.
	Type string `json:"type"`
	// Body is the content — opaque here; narrowed to a schema by a bead's type.
	Body json.RawMessage `json:"body"`
	// Metadata is an opaque per-resource sidecar the protocol does not interpret.
	Metadata json.RawMessage `json:"metadata,omitempty"`
	// Provenance records who caused this revision, how its period began, and when.
	Provenance *Provenance `json:"provenance,omitempty"`
	// CreatedAt is the resource's birth time (RFC 3339), denormalized onto
	// every revision.
	CreatedAt string `json:"createdAt,omitempty"`
	// StorageClass is how the resource is kept (durable, ephemeral, intrinsic).
	StorageClass *StorageClass `json:"storageClass,omitempty"`
}

// Representation is the by-value form a read returns — the resolved,
// point-in-time value of a resource as of one revision, possibly projected by a
// Selector. The spec's rep<Resource> is a Resource, so this is an alias.
type Representation = Resource

// Authority is the resource that names a Type — a package (for a bead) or a
// registry like IANA (for content). It is a reference, so its pin lives in Rev:
// the pair (Authority, Type) fully qualifies a type, frozen at the resource's
// inception so the type replays identically.
type Authority = ResourceID

// StorageClass is how a resource is kept. Exactly one field is set.
//
// A scope whose resources are not Durable with Versioned true has no revisions
// to pin, and so relaxes its write references from pinned to unpinned —
// advertised through discovery (see UnversionedCapability).
type StorageClass struct {
	// Durable: replicated and exported. Non-nil carries whether prior revisions
	// are retained and addressable.
	Durable *DurableClass `json:"durable,omitempty"`
	// Ephemeral: local and never exported.
	Ephemeral bool `json:"ephemeral,omitempty"`
	// Intrinsic: computed by the service, not stored.
	Intrinsic bool `json:"intrinsic,omitempty"`
}

// DurableClass carries the parameters of a Durable storage class.
type DurableClass struct {
	// Versioned reports whether prior revisions are retained and addressable.
	Versioned bool `json:"versioned"`
}

// # Provenance

// Provenance records who took part in the change that made a revision, how the
// resource's present period began, and when. It is data, not evidence: Core
// records it and does not claim it is true.
type Provenance struct {
	// Principals are the identities that took part in the change; may be empty
	// for observed or baseline revisions.
	Principals []Principal `json:"principals"`
	// Inception is how the resource's present period began.
	Inception Inception `json:"inception"`
	// CommittedAt is when this revision committed (RFC 3339); server-supplied,
	// never caller input.
	CommittedAt string `json:"committedAt"`
	// PreviousRevision is the before-image pointer — the basis for
	// compensation. Empty on create.
	PreviousRevision Rev `json:"previousRevision,omitempty"`
	// Message is an optional free-text note describing the change.
	Message string `json:"message,omitempty"`
	// Metadata is per-revision opaque data.
	Metadata json.RawMessage `json:"metadata,omitempty"`
}

// Inception is how a resource's present period began. It is a small closed set
// in Core; a reader should still tolerate an unrecognized value.
type Inception string

const (
	InceptionCreate  Inception = "create"
	InceptionImport  Inception = "import"
	InceptionRevive  Inception = "revive"
	InceptionUnknown Inception = "unknown"
)

// Principal is an identity that took part in a change. A role grants no
// permission; authorization is a separate decision. The principal is opaque at
// this layer (the spec's fields — id, displayName, roles, basis, authority —
// ride inside).
type Principal = json.RawMessage
