package bdp

import "encoding/json"

// This file renders the request/response shapes the primitives take and return:
// read-shaping and streaming, the write inputs, the change feed, export/import,
// and the call protocol.

// # Read-shaping and streaming

// Selector is a read projection over a representation: keep or drop subtrees by
// JSON Pointer, and expand named references inline. It travels in the query
// string, so a projected read has its own cacheable identity; Type and ID are
// never dropped.
type Selector struct {
	// Include keeps only these subtrees (JSON Pointers); everything else is dropped.
	Include []string `json:"include,omitempty"`
	// Omit drops these subtrees; everything else is kept.
	Omit []string `json:"omit,omitempty"`
	// Expand expands these named references inline instead of leaving them as refs.
	Expand []string `json:"expand,omitempty"`
}

// ListFilter is the low-bar filter every service supports: a flat conjunction
// (AND) of simple field conditions, exactly the query-string-expressible set.
// The full boolean/traversal query is a corelib capability, not Core.
type ListFilter []ListCondition

// ListCondition is one flat field condition. The spec's listCondition is a
// union of a handful of kinds (compare, is-kind by prefix/exact, has/missing);
// this renders the shape the Core low bar needs, kept deliberately small.
type ListCondition struct {
	// Field is the field path the condition tests, e.g. "body/status". Empty
	// when Kind targets the resource kind rather than a body field.
	Field string `json:"field,omitempty"`
	// Op is the comparison, e.g. "eq", "ne", "lt", "in". For a kind test it is
	// "kind-prefix" (is-it-an-issue?) or "kind-exact".
	Op string `json:"op"`
	// Value is the value compared against (opaque; shape depends on Op).
	Value json.RawMessage `json:"value,omitempty"`
}

// Ordinal is an event's position in its stream, monotone within one channel.
type Ordinal uint64

// Cursor is an opaque, transport-independent resume token, bound to one stream
// and the selection it came from and refused elsewhere. It is also the point an
// export is cut at, so a consumer can resume the change feed where a backup ends.
type Cursor string

// Transport is the wire projection a streaming read selects. Windowed JSON is
// the floor every service offers; JSONL is the bulk form; SSE is live and valid
// only on a live source (pull), refused by a snapshot (list/export).
type Transport string

const (
	TransportJSON  Transport = "json"
	TransportJSONL Transport = "jsonl"
	TransportSSE   Transport = "sse"
)

// ReadRequest bundles where to resume a streaming read, how much, the
// transport, and per-item shaping. After and From are mutually exclusive.
type ReadRequest struct {
	// After resumes at an opaque cursor (exact). Mutually exclusive with From.
	After Cursor `json:"after,omitempty"`
	// From starts at a known ordinal. Mutually exclusive with After.
	From *Ordinal `json:"from,omitempty"`
	// Limit caps the window size (0 = service default).
	Limit int `json:"limit,omitempty"`
	// Format is the transport form.
	Format Transport `json:"format,omitempty"`
	// Filter filters the stream.
	Filter ListFilter `json:"filter,omitempty"`
	// Select projects each item.
	Select *Selector `json:"select,omitempty"`
}

// Page is one finite window of a stream — the windowed floor every service
// offers: one request, one page, nothing held open. It is generic over the item
// type so each read states what it returns (representations, raw events, export
// records).
type Page[T any] struct {
	// Items are the items in this window.
	Items []T `json:"items"`
	// Cursor is where to resume; opaque, bound to this stream and selection.
	Cursor Cursor `json:"cursor"`
	// Head reports whether the window has caught up to the live edge.
	Head bool `json:"head,omitempty"`
}

// # Write inputs

// IDProposal is a caller's proposed identity on create. Exactly one of Exact or
// Prefix is set: Exact proposes a full id (fail-if-exists); Prefix roots an
// auto-id the server mints the leaf under. Both empty means the server assigns
// the whole id.
type IDProposal struct {
	// Exact is a full caller-proposed id.
	Exact string `json:"exact,omitempty"`
	// Prefix is an id prefix the server mints the leaf under.
	Prefix string `json:"prefix,omitempty"`
}

// CreateInput is the closed input to Create: an envelope member Core does not
// define is refused, not ignored (this does not reach inside Body/Metadata,
// which Core does not interpret).
type CreateInput struct {
	// Namespace is the scope to create the resource in.
	Namespace Namespace `json:"-"`
	// ID is the caller-proposed identity; zero value means the server assigns it.
	ID IDProposal `json:"requestedId,omitempty"`
	// Type is the bead's type path (e.g. "issue/task").
	Type string `json:"type"`
	// Authority is the package that names Type; nil means the server resolves it.
	Authority *Authority `json:"authority,omitempty"`
	// Body is the bead's body — the firewalled domain payload.
	Body json.RawMessage `json:"body"`
	// Metadata is an optional open sidecar the protocol does not interpret.
	Metadata json.RawMessage `json:"metadata,omitempty"`
}

// PutInput is the input to Put — the idempotent full-state write. Target is the
// resource's id (its Rev is ignored; Put targets the current resource). Use
// IfRevision for the optimistic-concurrency precondition.
type PutInput struct {
	// Target is the resource to write (its id).
	Target ResourceID `json:"-"`
	// Type is the bead's type path.
	Type string `json:"type"`
	// Authority is the package that names Type; nil means the server resolves it.
	Authority *Authority `json:"authority,omitempty"`
	// Body is the full new body — Put replaces state wholesale.
	Body json.RawMessage `json:"body"`
	// Metadata is an optional open sidecar the protocol does not interpret.
	Metadata json.RawMessage `json:"metadata,omitempty"`
	// IfRevision, when set, applies the write only if it matches the current
	// revision (the HTTP If-Match precondition); a stale value faults with a
	// revision mismatch and changes nothing.
	IfRevision Rev `json:"-"`
}

// JSONPatch is an RFC 6902 JSON Patch document. On patch, its paths are limited
// to /body and /metadata. JSON Patch is the base format; a richer patch format
// is an advertised capability.
type JSONPatch = json.RawMessage

// # Change events

// ChangeEvent reports one committed change to one resource, carried on the
// scope's built-in change feed at /sys/changes. It is a lifecycle union
// (created, updated, deleted); the shape is opaque here and decoded by the
// reader. One commit produces exactly one event per resource changed; a failed
// or no-change request produces none.
type ChangeEvent = json.RawMessage

// # Export and import

// ExportRecord is one line of an export: a begin record, a resource line, or an
// end record (told apart by a discriminator). A reader refuses a file that ends
// without the end record, so a truncated backup is never taken for whole.
type ExportRecord = json.RawMessage

// ImportStrategy controls how Import merges records.
type ImportStrategy struct {
	// Dedup skips a record already present instead of faulting (idempotent re-import).
	Dedup bool `json:"dedup,omitempty"`
	// AllowStale admits a record whose revision is behind the store's current one.
	AllowStale bool `json:"allowStale,omitempty"`
	// DryRun validates and reports, but writes nothing.
	DryRun bool `json:"dryRun,omitempty"`
	// IDRoot remaps incoming ids under this id-root prefix.
	IDRoot string `json:"idRoot,omitempty"`
}

// ImportReport is what Import reports — counts and any skipped/conflicted records.
type ImportReport struct {
	Imported  int              `json:"imported"`
	Skipped   int              `json:"skipped"`
	Conflicts []RecordConflict `json:"conflicts"`
}

// RecordConflict names one rejected import record by id and reason.
type RecordConflict struct {
	ID     ResourceID `json:"id"`
	Reason string     `json:"reason"`
}

// PushStrategy controls how Push handles a rejected or re-pushed event — the
// event analog of ImportStrategy.
type PushStrategy struct {
	// Dedup skips an event already appended (idempotent re-push) instead of faulting.
	Dedup bool `json:"dedup,omitempty"`
	// DryRun validates and reports, but appends nothing.
	DryRun bool `json:"dryRun,omitempty"`
}

// PushReport is what Push reports — counts and any skipped/rejected events.
type PushReport struct {
	Pushed    int             `json:"pushed"`
	Skipped   int             `json:"skipped"`
	Conflicts []EventConflict `json:"conflicts"`
}

// EventConflict names one rejected pushed event by its ordinal and reason.
type EventConflict struct {
	Ordinal Ordinal `json:"ordinal"`
	Reason  string  `json:"reason"`
}

// # The call protocol

// Request is what a caller sends to Call. Arguments is a name/value list whose
// shape the target Operation's parameters dictate.
type Request struct {
	// Arguments are the positional-by-name arguments, shaped by the operation.
	Arguments []Argument `json:"arguments"`
	// Async requests async disposition (default sync).
	Async bool `json:"async,omitempty"`
	// CausedBy is the cause that triggered this call (correlation).
	CausedBy string `json:"causedBy,omitempty"`
	// CalledBy is the calling frame (correlation, diagnostic-tier).
	CalledBy string `json:"calledBy,omitempty"`
	// Provenance is the caller's part of provenance only (never server fields).
	Provenance *Provenance `json:"provenance,omitempty"`
	// IdempotencyKey collapses a retried call onto the first effect.
	IdempotencyKey string `json:"idempotencyKey,omitempty"`
	// Metadata is an opaque per-call sidecar.
	Metadata json.RawMessage `json:"metadata,omitempty"`
}

// Argument is one name/value call argument.
type Argument struct {
	Name  string          `json:"name"`
	Value json.RawMessage `json:"value"`
}

// Reply is what Call returns — flat, no context wrapper. Outcome keeps the
// domain Result nested; Cause/Frame/Transaction are correlation.
type Reply struct {
	// Outcome is what happened — the domain Result or a non-domain fate.
	Outcome Outcome `json:"outcome"`
	// Cause is the cause this call ran under (correlation).
	Cause string `json:"cause"`
	// Frame is the frame this call ran under (correlation).
	Frame string `json:"frame"`
	// Transaction is the commit this call belonged to, if any.
	Transaction string `json:"transaction,omitempty"`
	// Metadata is an opaque per-reply sidecar.
	Metadata json.RawMessage `json:"metadata,omitempty"`
}

// Outcome is what happened to an invocation — the settled result of a Call, and
// the one place in Core with a rich result. Exactly one field is set. Only
// Success means the operation went as intended. The non-domain set
// (Fault/Canceled) is open: a reader must tolerate an unrecognized fate, so an
// unknown arm is reported through Unknown rather than dropped.
type Outcome struct {
	// Success carries the operation's declared success value (the Result's T).
	Success json.RawMessage `json:"success,omitempty"`
	// Failure carries the operation's declared failure — the Result's E, an
	// expected domain "no". It settles a sync Call as HTTP 409. Call is the only
	// primitive with a declared failure; a data primitive has none.
	Failure json.RawMessage `json:"failure,omitempty"`
	// Fault carries a Problem — an undeclared, non-domain failure (NOT the E):
	// the call itself went wrong. A Problem is the universal fault channel here
	// exactly as it is for the data primitives, never a Result's E.
	Fault *Problem `json:"fault,omitempty"`
	// Canceled reports the invocation was killed; Reason is optional.
	Canceled *CanceledOutcome `json:"canceled,omitempty"`
	// Unknown holds a non-domain fate this translation does not model, so an
	// evolving outcome set is tolerated rather than lost.
	Unknown json.RawMessage `json:"-"`
}

// CanceledOutcome is the payload of a canceled invocation.
type CanceledOutcome struct {
	Reason string `json:"reason,omitempty"`
}
