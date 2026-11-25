package bdp

import (
	"context"
	"encoding/json"

	"github.com/steveyegge/beads/beadserrors"
)

// Primitive names a BDP Core primitive. A scope advertises the set it offers
// through [Scope.Capabilities]; a primitive it does not offer faults with a
// *[beadserrors.ErrUnsupported] naming it.
type Primitive string

// The Core primitive set (core/core.lumen §3.3). get and list and the
// changefeed pull are the minimal tier; the rest are the full tier, each
// advertised, not assumed.
const (
	PrimitiveGet       Primitive = "get"       // safe point read                (minimal)
	PrimitiveList      Primitive = "list"      // a scope's current resources    (minimal)
	PrimitivePull      Primitive = "pull"      // read a channel / the changefeed (minimal)
	PrimitiveRevisions Primitive = "revisions" // one resource's history         (full)
	PrimitiveExport    Primitive = "export"    // bulk read as JSONL             (full)
	PrimitiveCreate    Primitive = "create"    // admit a new bead               (full)
	PrimitivePut       Primitive = "put"       // idempotent full-state write    (full)
	PrimitivePatch     Primitive = "patch"     // JSON Patch /body + /metadata   (full)
	PrimitiveDelete    Primitive = "delete"    // retire, minting a tombstone    (full)
	PrimitivePush      Primitive = "push"      // write a channel                (full)
	PrimitiveImport    Primitive = "import"    // bulk write path                (full)
	PrimitiveCall      Primitive = "call"      // invoke an Operation            (full)
	PrimitiveCancel    Primitive = "cancel"    // stop a running async call      (full)
)

// Capabilities is what [Scope.Capabilities] advertises — read on the wire by a
// get of the well-known /sys/capabilities resource, not a distinct primitive
// (core.lumen §Capabilities @ bdp-core-0.2.3-rc2). Discovery is part of the
// contract — the spec requires it so a caller never guesses — which is why it is
// a mandatory member alongside Get.
//
// rc2 partitions the advertisement à la carte: the base core package's per-op
// capabilities (a primitive is offered iff its field is present), the scope-wide
// traits that are neither per-primitive nor pack-specific, and other packages'
// capabilities keyed by package path. This mirrors that shape faithfully; Packs
// is left opaque (this is a core client).
type Capabilities struct {
	// Core is the base package's advertised primitives — a primitive is offered
	// iff its field is present. Get is always present (the floor).
	Core CoreCapabilities `json:"core"`
	// Scope is the scope-wide traits (storage classes, revision model, id model).
	Scope ScopeCapabilities `json:"scope"`
	// Packs is other packages' capabilities, keyed by package path; opaque here.
	Packs map[string]json.RawMessage `json:"packs,omitempty"`
}

// CoreCapabilities is the core package's à-la-carte primitive advertisement: a
// primitive is offered iff its field is non-nil, and that field carries the
// primitive's limits. Get is always present — point reads are the floor
// (core.lumen §CoreCapabilities @ bdp-core-0.2.3-rc2).
type CoreCapabilities struct {
	Get       GetCaps        `json:"get"`
	List      *ListCaps      `json:"list,omitempty"`
	Pull      *PullCaps      `json:"pull,omitempty"`
	Create    *CreateCaps    `json:"create,omitempty"`
	Put       *PutCaps       `json:"put,omitempty"`
	Patch     *PatchCaps     `json:"patch,omitempty"`
	Delete    *DeleteCaps    `json:"delete,omitempty"`
	Revisions *RevisionsCaps `json:"revisions,omitempty"`
	Export    *ExportCaps    `json:"export,omitempty"`
	Import    *ImportCaps    `json:"import,omitempty"`
	Push      *PushCaps      `json:"push,omitempty"`
}

// ScopeCapabilities is a scope's scope-wide traits — neither per-primitive nor
// pack-specific (core.lumen §ScopeCapabilities @ bdp-core-0.2.3-rc2). The traits
// do not combine freely: versioned REQUIRES permanent (a revision chain may not
// span two resources under one reused id), equivalently reusable REQUIRES
// unversioned.
type ScopeCapabilities struct {
	// StorageClasses is which persistence classes the scope can honor — a set
	// over Durable/Ephemeral (Intrinsic is universal and never listed).
	StorageClasses []StorageClass `json:"storageClasses"`
	// RevisionModel is whether the scope mints revisions, scope-wide.
	RevisionModel RevisionModel `json:"revisionModel"`
	// IdModel is whether a retired id is permanently spent, scope-wide.
	IdModel IdModel `json:"idModel"`
	// Aliases, when present, signals the scope supports alias resources and the
	// scope-root overlay. Opaque here.
	Aliases *AliasCaps `json:"aliases,omitempty"`
	// BodyFormSpill reports the scope accepts the optional body-form spill for an
	// over-budget read/write guard; absent/false ⇒ over-budget faults request-too-large.
	BodyFormSpill bool `json:"bodyFormSpill,omitempty"`
}

// Offers reports whether the scope advertises primitive p — presence of its
// core capability. Get is always offered.
func (c Capabilities) Offers(p Primitive) bool {
	switch p {
	case PrimitiveGet:
		return true
	case PrimitiveList:
		return c.Core.List != nil
	case PrimitivePull:
		return c.Core.Pull != nil
	case PrimitiveCreate:
		return c.Core.Create != nil
	case PrimitivePut:
		return c.Core.Put != nil
	case PrimitivePatch:
		return c.Core.Patch != nil
	case PrimitiveDelete:
		return c.Core.Delete != nil
	case PrimitiveRevisions:
		return c.Core.Revisions != nil
	case PrimitiveExport:
		return c.Core.Export != nil
	case PrimitiveImport:
		return c.Core.Import != nil
	case PrimitivePush:
		return c.Core.Push != nil
	default:
		// call/cancel are not core data primitives advertised here; a caller
		// learns an operation's availability by resolving it.
		return false
	}
}

// Unversioned reports whether the scope keeps no revisions (its RevisionModel is
// unversioned) — the convenience predicate for the common branch.
func (c Capabilities) Unversioned() bool {
	return c.Scope.RevisionModel == RevisionModelUnversioned
}

// RevisionModel is the scope-wide versioning trait (core.lumen §RevisionModel @
// bdp-core-0.2.3-rc2). versioned ⇒ the scope mints revisions and write refs are
// pinned; unversioned ⇒ it mints none, rev is absent everywhere, and write refs
// relax to unpinned scope-wide. The empty value is unspecified — a scope that
// advertises discovery always states it.
type RevisionModel string

const (
	RevisionModelVersioned   RevisionModel = "versioned"
	RevisionModelUnversioned RevisionModel = "unversioned"
)

// IdModel is the scope-wide id model (core.lumen §IdModel @ bdp-core-0.2.3-rc2):
// whether an id retired by delete is permanently spent or may be minted again.
// permanent ⇒ a retired id is never reused (an unpinned inbound reference faults
// 410 Gone). reusable ⇒ a retired id may be re-minted, REQUIRES unversioned, and
// obliges delete to advertise a precondition guard (no revision exists for
// If-Match), so a client retrying a delete under reusable has no revision to
// pin and must guard by state.
type IdModel string

const (
	IdModelPermanent IdModel = "permanent"
	IdModelReusable  IdModel = "reusable"
)

// GetCaps are the limits advertised for get.
type GetCaps struct {
	// Selector reports a select projection is honored on a read.
	Selector bool `json:"selector"`
}

// ListCaps are the limits advertised for list.
type ListCaps struct {
	// Conditions is the low-bar filter operators the scope supports.
	Conditions []ConditionKind `json:"conditions"`
	// Selector reports a select projection is honored.
	Selector bool `json:"selector"`
	// Sort reports server-side sort is offered.
	Sort bool `json:"sort"`
	// CompactFilter reports the compact ?where= dialect is accepted (on top of
	// the URL-encoded JSON Filter floor every scope accepts).
	CompactFilter bool `json:"compactFilter"`
	// MaxLimit is the maximum page size when capped (0 = uncapped).
	MaxLimit int `json:"maxLimit,omitempty"`
}

// ConditionKind is one operator in the low-bar list filter; the richer query
// language is a pack capability, not core (core.lumen §ConditionKind @ rc2).
type ConditionKind string

const (
	ConditionCompare     ConditionKind = "compare"
	ConditionMatches     ConditionKind = "matches"
	ConditionHas         ConditionKind = "has"
	ConditionMissing     ConditionKind = "missing"
	ConditionContains    ConditionKind = "contains"
	ConditionContainsAny ConditionKind = "containsAny"
	ConditionExcludes    ConditionKind = "excludes"
	ConditionIsType      ConditionKind = "isType"
)

// PullCaps are the limits advertised for pull (the changefeed and streams).
type PullCaps struct {
	// SSE reports live sse transport is offered.
	SSE bool `json:"sse"`
	// JSONL reports bulk jsonl transport is offered.
	JSONL bool `json:"jsonl"`
	// Filter reports a where filter is honored on the stream.
	Filter bool `json:"filter"`
	// Selector reports a select projection is honored on each item.
	Selector bool `json:"selector"`
}

// CreateCaps are the limits advertised for create.
type CreateCaps struct {
	// ServerAllocatesID reports the server will allocate the id (no requestedId).
	ServerAllocatesID bool `json:"serverAllocatesId"`
	// ClientProposedID reports a caller-proposed id (requestedId) is honored.
	ClientProposedID bool `json:"clientProposedId"`
}

// PutCaps are the write guards put honors (core.lumen §PutCaps @ rc2).
type PutCaps struct {
	// IfRevision reports the If-Match optimistic-concurrency guard is honored.
	IfRevision bool `json:"ifRevision"`
	// IfAbsent reports the If-None-Match: * create-if-absent guard is honored.
	IfAbsent bool `json:"ifAbsent"`
	// Precondition reports the ?precondition= state guard is honored.
	Precondition bool `json:"precondition"`
}

// PatchCaps are the write guards patch honors.
type PatchCaps struct {
	IfRevision   bool `json:"ifRevision"`
	Precondition bool `json:"precondition"`
}

// DeleteCaps are the write guards delete honors. A reusable scope offering
// delete MUST set Precondition (it has no revision for If-Match to carry).
type DeleteCaps struct {
	IfRevision   bool `json:"ifRevision"`
	Precondition bool `json:"precondition"`
}

// RevisionsCaps, present, advertises that prior revisions are addressable
// (revisions) — offered only by a versioned scope. It carries no limits.
type RevisionsCaps struct{}

// ExportCaps are the limits advertised for export.
type ExportCaps struct {
	// Revisions reports the export can include prior revisions (history), not
	// only the current state.
	Revisions bool `json:"revisions"`
}

// ImportCaps are the limits advertised for import.
type ImportCaps struct {
	// HistoryRestore reports importing prior revisions (history restore) is honored.
	HistoryRestore bool `json:"historyRestore"`
	// Window is the admission window before backpressure, when bounded; its
	// presence advertises resumable/backpressured import, absent ⇒ one-shot.
	Window *int `json:"window,omitempty"`
}

// PushCaps are the limits advertised for push.
type PushCaps struct {
	// MaxBatch is the maximum events accepted in one batch, when capped.
	MaxBatch *int `json:"maxBatch,omitempty"`
	// Window is the admission window before backpressure, when bounded.
	Window *int `json:"window,omitempty"`
}

// AliasCaps is advertised for aliases — its presence signals alias resources
// under sys/aliases/ and the scope-root overlay are active. Opaque here.
type AliasCaps struct{}

// ServiceDescriptor is the bounded discovery signpost returned by GET / (a
// scope's base URL) — a thin envelope that names the protocol and links to the
// rest. With get and /sys/capabilities it is the mandatory floor; a client MAY
// hardcode GET /, GET /sys/capabilities and GET /bdp/openapi (core.lumen
// §ServiceDescriptor @ bdp-core-0.2.3-rc2).
type ServiceDescriptor struct {
	// Version is the BDP protocol version this scope speaks, a full SemVer.
	Version string `json:"version"`
	// Revision is the exact spec commit/tag it was built against; diagnostic only.
	Revision Rev `json:"revision,omitempty"`
	// Scope is the scope's identity, echoed; absent on a base URL fronting no
	// single scope.
	Scope string `json:"scope,omitempty"`
	// Base is the scope's base URL, echoed so a client can confirm the origin.
	Base string `json:"base,omitempty"`
	// Links are typed links to the rest of the service.
	Links ServiceLinks `json:"links"`
}

// ServiceLinks is the discovery link set — a relation's absence means the
// service does not offer it (core.lumen §ServiceLinks @ bdp-core-0.2.3-rc2).
type ServiceLinks struct {
	// ServiceDesc is the machine contract (/bdp/openapi), RFC 8631 service-desc.
	ServiceDesc string `json:"serviceDesc,omitempty"`
	// ServiceMeta is the scope's live Capabilities (/sys/capabilities), service-meta.
	ServiceMeta string `json:"serviceMeta,omitempty"`
	// ServiceDoc is human-readable documentation, service-doc.
	ServiceDoc string `json:"serviceDoc,omitempty"`
	// Changes is the scope's changefeed Channel (/sys/changes), BDP extension relation.
	Changes string `json:"changes,omitempty"`
}

// Scope is one opened BDP store — the v0.2 Core conformance boundary. A client
// (talking to a remote store over HTTP) and a storage backend both satisfy it.
//
// The shape is à la carte, not profiles: Get and Capabilities are the only
// members a scope must implement, and every other primitive is advertised
// through Capabilities and faults with a *[beadserrors.ErrUnsupported] when a
// scope does not offer it. Embed [UnimplementedScope] to inherit that refusing
// default for every optional primitive and write only Get and Capabilities.
//
// Error model: a data primitive returns its value directly and faults with an
// error ([FaultError], carrying a [Problem], and mapped onto a shared sentinel
// where one fits). The rich Result/Outcome split is Call's alone and rides in
// the returned [Reply].
type Scope interface {
	// Capabilities advertises the primitives this scope offers and their limits.
	// Mandatory: a caller relies on it to know what is safe to invoke.
	Capabilities(ctx context.Context) (Capabilities, error)

	// Get is the one mandatory primitive: a safe point read resolving ref to its
	// representation — the current one, or the pinned revision when ref pins Rev.
	// An optional selector projects the result. It faults ErrNotFound when the
	// resource is absent or erased (a hidden resource reads identically). Get is
	// safe and has no declared failure.
	Get(ctx context.Context, ref ResourceID, selector *Selector) (Representation, error)

	// --- Optional reads ---

	// List returns a window of a namespace's current resources, filtered by the
	// low-bar where and projected by selector. Snapshot: json/jsonl, no SSE.
	List(ctx context.Context, ns Namespace, where ListFilter, selector *Selector, req *ReadRequest) (Page[Representation], error)

	// Revisions returns a window of one resource's history — id-targeted like Get,
	// windowed like List. Snapshot.
	Revisions(ctx context.Context, ref ResourceID, selector *Selector, req *ReadRequest) (Page[Representation], error)

	// Pull reads a channel as a window of its events — canonically the scope's
	// changefeed at /sys/changes, whose events are ChangeEvents. One cursor spans
	// the windowed, bulk, and live transports.
	Pull(ctx context.Context, channel ResourceID, req ReadRequest) (Page[ChangeEvent], error)

	// Export is the bulk read of a namespace as a window of export records —
	// current state, or every revision in order when revisions is true (a true
	// backup). Snapshot, cut at a cursor so the changefeed can resume from it.
	Export(ctx context.Context, ns Namespace, where ListFilter, selector *Selector, revisions bool, req *ReadRequest) (Page[ExportRecord], error)

	// --- Optional writes ---

	// Create admits a new bead (fail-if-exists) and returns its pinned reference.
	// The input is closed: an undefined envelope member is refused, not ignored.
	Create(ctx context.Context, in CreateInput) (ResourceID, error)

	// Put is the idempotent full-state write (create-or-replace) and returns the
	// written pinned reference. A write that changes nothing keeps the revision.
	Put(ctx context.Context, in PutInput) (ResourceID, error)

	// Patch applies a JSON Patch to /body and /metadata and returns the patched
	// pinned reference. Guards ride the request line, never the body: a set
	// ifRevision is the If-Match guard, a non-empty precondition is the
	// ?precondition= state guard.
	Patch(ctx context.Context, ref ResourceID, patch JSONPatch, ifRevision Rev, precondition ListFilter) (ResourceID, error)

	// Delete retires a resource, minting a tombstone revision which it returns
	// (not a representation). Idempotent. A set ifRevision is the If-Match guard;
	// a non-empty precondition is the ?precondition= state guard (the only guard a
	// reusable scope can offer, since it keeps no revision to pin).
	Delete(ctx context.Context, ref ResourceID, ifRevision Rev, precondition ListFilter) (Rev, error)

	// Push appends one or more events to a channel and returns a PushAck — the
	// high-water ordinal, a per-event result, and the backpressure signal. Push
	// is non-blocking: the channel accepts a prefix and defers the rest, and the
	// pusher re-offers the deferred tail. A keyed event (carrying an Ordinal) is
	// idempotent. The dual of Pull and the channel analog of Import.
	Push(ctx context.Context, channel ResourceID, events []PushEvent) (PushAck, error)

	// Import is the bulk write path: it replays export records into a namespace
	// under the given strategy.
	Import(ctx context.Context, ns Namespace, records []ExportRecord, strategy *ImportStrategy) (ImportReport, error)

	// --- Optional invoke ---

	// Call invokes an Operation and returns its Reply — the only primitive with
	// the rich Outcome (declared success, declared failure, or fault). An async
	// request settles later via the returned handle.
	Call(ctx context.Context, operation ResourceID, req Request) (Reply, error)

	// Cancel stops a running async call, returning the settled async-result
	// reference.
	Cancel(ctx context.Context, asyncResult ResourceID) (ResourceID, error)
}

// UnimplementedScope supplies the refusing default — a *[beadserrors.ErrUnsupported]
// — for every OPTIONAL primitive. Embed it to build a scope à la carte:
//
//	type myScope struct{ bdp.UnimplementedScope }
//	func (s *myScope) Capabilities(context.Context) (bdp.Capabilities, error) { ... }
//	func (s *myScope) Get(ctx context.Context, ref bdp.ResourceID, sel *bdp.Selector) (bdp.Representation, error) { ... }
//
// It deliberately does NOT provide Get or Capabilities, so the compiler makes
// every implementer write the two mandatory members itself — the go-ification of
// the spec's "get is the only required primitive" (plus the mandatory discovery
// member). Set Backend to name this scope in the refusals it returns.
type UnimplementedScope struct {
	// Backend names this scope in the ErrUnsupported refusals it returns, e.g.
	// "bdp remote client". Optional.
	Backend string
}

func (u UnimplementedScope) unsupported(op string) error {
	return &beadserrors.ErrUnsupported{Op: op, Backend: u.Backend}
}

func (u UnimplementedScope) List(context.Context, Namespace, ListFilter, *Selector, *ReadRequest) (Page[Representation], error) {
	return Page[Representation]{}, u.unsupported("list")
}

func (u UnimplementedScope) Revisions(context.Context, ResourceID, *Selector, *ReadRequest) (Page[Representation], error) {
	return Page[Representation]{}, u.unsupported("revisions")
}

func (u UnimplementedScope) Pull(context.Context, ResourceID, ReadRequest) (Page[ChangeEvent], error) {
	return Page[ChangeEvent]{}, u.unsupported("pull")
}

func (u UnimplementedScope) Export(context.Context, Namespace, ListFilter, *Selector, bool, *ReadRequest) (Page[ExportRecord], error) {
	return Page[ExportRecord]{}, u.unsupported("export")
}

func (u UnimplementedScope) Create(context.Context, CreateInput) (ResourceID, error) {
	return ResourceID{}, u.unsupported("create")
}

func (u UnimplementedScope) Put(context.Context, PutInput) (ResourceID, error) {
	return ResourceID{}, u.unsupported("put")
}

func (u UnimplementedScope) Patch(context.Context, ResourceID, JSONPatch, Rev, ListFilter) (ResourceID, error) {
	return ResourceID{}, u.unsupported("patch")
}

func (u UnimplementedScope) Delete(context.Context, ResourceID, Rev, ListFilter) (Rev, error) {
	return "", u.unsupported("delete")
}

func (u UnimplementedScope) Push(context.Context, ResourceID, []PushEvent) (PushAck, error) {
	return PushAck{}, u.unsupported("push")
}

func (u UnimplementedScope) Import(context.Context, Namespace, []ExportRecord, *ImportStrategy) (ImportReport, error) {
	return ImportReport{}, u.unsupported("import")
}

func (u UnimplementedScope) Call(context.Context, ResourceID, Request) (Reply, error) {
	return Reply{}, u.unsupported("call")
}

func (u UnimplementedScope) Cancel(context.Context, ResourceID) (ResourceID, error) {
	return ResourceID{}, u.unsupported("cancel")
}
