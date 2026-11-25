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
	// IfAbsent, when true, applies the write only if no resource exists at the id
	// (HTTP If-None-Match: *) — create-if-absent; a present id faults
	// ErrAlreadyExists (409, the same status as a Create collision). A put without
	// it targets the resource, so a retired target faults 410. (core.lumen §put @
	// bdp-core-0.2.3-rc2; advertised by PutCaps.IfAbsent.)
	IfAbsent bool `json:"-"`
	// Precondition, when set, applies the write only if the current resource
	// satisfies this flat-AND state filter (HTTP ?precondition=); else faults
	// precondition-failed (412). It rides the request line, never the body.
	// (core.lumen §put @ bdp-core-0.2.3-rc2; advertised by PutCaps.Precondition.)
	Precondition ListFilter `json:"-"`
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

// PushEvent is one event offered to Push. Ordinal, when set, is the
// pusher-assigned position that makes the append idempotent (keyed); absent, the
// channel assigns one. Body is the channel's pushSchema payload — any value, not
// only a record (core.lumen §PushEvent @ bdp-core-0.2.3-rc2).
type PushEvent struct {
	Ordinal *Ordinal        `json:"ordinal,omitempty"`
	Body    json.RawMessage `json:"body"`
}

// PushAck is what Push returns — the high-water ordinal durably appended, a
// per-event result, and the backpressure signal. Push is non-blocking: the
// channel accepts a prefix up to its admission Credit and defers the rest; the
// pusher holds the deferred tail and re-offers from AcceptedThrough+1 after
// RetryAfter. A short Results with Credit 0 is normal backpressure, not an
// error — distinct from limit-exceeded (422), the hard cap (core.lumen §PushAck
// @ bdp-core-0.2.3-rc2).
type PushAck struct {
	AcceptedThrough Ordinal      `json:"acceptedThrough"`
	Results         []PushResult `json:"results"`
	Credit          int          `json:"credit"`
	RetryAfter      int          `json:"retryAfter,omitempty"`
}

// PushResultKind is the fate of one offered event.
type PushResultKind string

const (
	PushAccepted PushResultKind = "accepted" // appended at Ordinal
	PushDeduped  PushResultKind = "deduped"  // already present (idempotent re-push)
	PushDeferred PushResultKind = "deferred" // backpressure; re-offer after RetryAfter
)

// PushResult is the fate of one offered event — the rc2 open union
// Accepted / Deduped / Deferred; Ordinal is set for Accepted and Deduped
// (core.lumen §PushResult @ bdp-core-0.2.3-rc2).
//
// NOTE: push is not exercised by the clean-room smoke test, so the exact wire
// tagging/casing of this union is UNVERIFIED against a live server; it is
// modeled here for contract completeness and verified when push is first built
// out (Track A).
type PushResult struct {
	Kind    PushResultKind `json:"kind"`
	Ordinal *Ordinal       `json:"ordinal,omitempty"`
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

// OutcomeKind is the discriminator on the settled result of a Call.
type OutcomeKind string

const (
	OutcomeSuccess  OutcomeKind = "success"  // the Result's T, carried under "value"
	OutcomeFailure  OutcomeKind = "failure"  // the Result's E, carried under "error"
	OutcomeFault    OutcomeKind = "fault"    // a Problem, carried under "problem"
	OutcomeCanceled OutcomeKind = "canceled" // killed; optional "reason"
)

// Outcome is what happened to an invocation — the settled result of a Call, and
// the one place in Core with a rich result. On the wire it is an INTERNALLY
// TAGGED object: a "kind" discriminator plus that arm's own payload key —
//
//	{"kind":"success","value":T} · {"kind":"failure","error":E} ·
//	{"kind":"fault","problem":P}  · {"kind":"canceled","reason"?:string}
//
// so a reader never infers the arm from the HTTP status alone (core.lumen
// §Outcome @ bdp-core-0.2.3-rc2). The non-domain set (fault·canceled) is an open
// union, so an unrecognized kind is preserved in Unknown rather than dropped.
//
// Only Success means the operation went as intended. Failure is the declared
// domain "no" — the Result's E, settling a sync Call as HTTP 409. Fault is an
// undeclared, non-domain failure carrying a Problem — the universal fault
// channel, never a Result's E.
type Outcome struct {
	// Kind is the settled arm; empty only on a zero Outcome.
	Kind OutcomeKind
	// Value is the declared success value (Kind == OutcomeSuccess).
	Value json.RawMessage
	// Error is the declared failure value, the Result's E (Kind == OutcomeFailure).
	Error json.RawMessage
	// Fault is the Problem of a non-domain failure (Kind == OutcomeFault).
	Fault *Problem
	// Reason is the optional cancel reason (Kind == OutcomeCanceled).
	Reason string
	// Unknown preserves the raw body of an unrecognized kind, so an evolving
	// outcome set is tolerated rather than lost.
	Unknown json.RawMessage
}

// wireOutcome is the internally-tagged on-wire form of an [Outcome].
type wireOutcome struct {
	Kind    OutcomeKind     `json:"kind"`
	Value   json.RawMessage `json:"value,omitempty"`
	Error   json.RawMessage `json:"error,omitempty"`
	Problem *Problem        `json:"problem,omitempty"`
	Reason  string          `json:"reason,omitempty"`
}

// UnmarshalJSON decodes the internally-tagged Outcome, preserving an
// unrecognized kind verbatim in Unknown.
func (o *Outcome) UnmarshalJSON(b []byte) error {
	var w wireOutcome
	if err := json.Unmarshal(b, &w); err != nil {
		return err
	}
	*o = Outcome{Kind: w.Kind}
	switch w.Kind {
	case OutcomeSuccess:
		o.Value = w.Value
	case OutcomeFailure:
		o.Error = w.Error
	case OutcomeFault:
		o.Fault = w.Problem
	case OutcomeCanceled:
		o.Reason = w.Reason
	default:
		o.Unknown = append(json.RawMessage(nil), b...)
	}
	return nil
}

// MarshalJSON writes the internally-tagged Outcome.
func (o Outcome) MarshalJSON() ([]byte, error) {
	if o.Unknown != nil && o.Kind == "" {
		return o.Unknown, nil
	}
	w := wireOutcome{Kind: o.Kind}
	switch o.Kind {
	case OutcomeSuccess:
		w.Value = o.Value
	case OutcomeFailure:
		w.Error = o.Error
	case OutcomeFault:
		w.Problem = o.Fault
	case OutcomeCanceled:
		w.Reason = o.Reason
	}
	return json.Marshal(w)
}
