package bdp

import (
	"context"

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

// Capabilities is what [Scope.Capabilities] advertises: which primitives a
// scope offers and the limits on each. Discovery is part of the contract — the
// spec requires it so a caller never guesses — which is why it is a mandatory
// member alongside Get rather than a defaulted one.
type Capabilities struct {
	// Primitives maps each offered primitive to its advertised limits. Absence
	// from the map means the primitive is not offered (and faults with
	// ErrUnsupported). Get is always present.
	Primitives map[Primitive]PrimitiveCapability
	// Unversioned reports the scope keeps no revisions, so its writes return
	// unpinned references (its StorageClass is not Durable{Versioned:true}).
	Unversioned bool
}

// Offers reports whether the scope advertises primitive p.
func (c Capabilities) Offers(p Primitive) bool {
	_, ok := c.Primitives[p]
	return ok
}

// PrimitiveCapability carries the advertised knobs for one primitive. A field
// is meaningful only for the primitives it applies to (e.g. IfMatch for the
// conditional writes, Async for call); it is false/zero otherwise. The spec's
// per-primitive limit records are richer — this condenses the knobs a Core
// caller branches on.
type PrimitiveCapability struct {
	// Selector reports the read honors a projection selector.
	Selector bool
	// IfMatch reports the write honors the IfRevision (If-Match) precondition.
	IfMatch bool
	// JSONL reports the streaming read offers the bulk JSONL transport.
	JSONL bool
	// SSE reports the streaming read offers the live SSE transport.
	SSE bool
	// Revisions reports export can emit full revision history.
	Revisions bool
	// Async reports call accepts async disposition.
	Async bool
	// MaxBatch bounds a batched op (0 = unspecified / unbounded).
	MaxBatch int
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
	// pinned reference. A set ifRevision is the If-Match precondition.
	Patch(ctx context.Context, ref ResourceID, patch JSONPatch, ifRevision Rev) (ResourceID, error)

	// Delete retires a resource, minting a tombstone revision which it returns
	// (not a representation). Idempotent. A set ifRevision is the If-Match
	// precondition.
	Delete(ctx context.Context, ref ResourceID, ifRevision Rev) (Rev, error)

	// Push appends events to a channel in order — the dual of Pull and the
	// channel analog of Import.
	Push(ctx context.Context, channel ResourceID, events []ChangeEvent, strategy *PushStrategy) (PushReport, error)

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

func (u UnimplementedScope) Patch(context.Context, ResourceID, JSONPatch, Rev) (ResourceID, error) {
	return ResourceID{}, u.unsupported("patch")
}

func (u UnimplementedScope) Delete(context.Context, ResourceID, Rev) (Rev, error) {
	return "", u.unsupported("delete")
}

func (u UnimplementedScope) Push(context.Context, ResourceID, []ChangeEvent, *PushStrategy) (PushReport, error) {
	return PushReport{}, u.unsupported("push")
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
