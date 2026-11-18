package bdp

import (
	"errors"

	"github.com/steveyegge/beads/beadserrors"
)

// The BDP Core fault model (core.lumen §Problem, normative): every primitive
// returns its declared value on success or faults with a [Problem]. A fault is
// delivered out of band — the HTTP non-2xx response (RFC 9457) — and is never
// part of the declared return type; it is orthogonal to the value. A data
// primitive has no declared failure at all: a read returns a rep or faults.
//
// Only Call additionally carries a declared failure — the operation's own "no",
// which rides in [Outcome].Failure (the Result's E) and settles a sync call as
// HTTP 409. That 409 is a declared failure in the Reply body, not a fault; it is
// distinct from the 409 a data primitive raises (e.g. Create's identity clash),
// which is an out-of-band Problem like any other fault. A Problem is the
// universal fault channel, never the E of a Result.
//
// So that a caller holding only this contract can classify a fault without
// parsing a Problem, the common faults are also mapped onto sentinels: a
// transport returns (or wraps) the matching one, and errors.Is finds it.
//
// Two of those sentinels are the domain-neutral ones that already sit beneath
// every bd leaf ([beadserrors.ErrNotFound], [beadserrors.ErrValidation]). The
// other two — a revision precondition miss and an identity collision — have no
// domain-neutral form in beadserrors (issueops owns issue-flavored twins), so
// BDP defines its own protocol-level sentinels here rather than couple this
// boundary to the issue domain.
var (
	// ErrNotFound is the 404/410 fault: the resource is absent or erased (or
	// hidden, which reads identically — a service never reveals a hidden
	// resource).
	ErrNotFound = beadserrors.ErrNotFound
	// ErrValidation is the 400/422 fault: the request was malformed or
	// unprocessable.
	ErrValidation = beadserrors.ErrValidation
	// ErrRevisionMismatch is the 412 fault: an If-Match / IfRevision precondition
	// did not match the current revision, so the write changed nothing. (The
	// spec's wire code is "revision-mismatch".)
	ErrRevisionMismatch = errors.New("bdp: revision mismatch")
	// ErrAlreadyExists is the 409 fault from Create against an id already in use,
	// or used and retired — an existence conflict, not a declared constraint. It
	// is an out-of-band Problem (a data primitive has no declared failure), not
	// to be confused with the 409 a sync Call returns for a declared failure,
	// which rides in Outcome.Failure.
	ErrAlreadyExists = errors.New("bdp: identity taken")
)

// Problem is an RFC 9457 problem document — what a fault carries, and the body
// of a failing response from a data primitive. It is the universal fault
// channel, never the E of a Result (a declared failure is [Outcome].Failure).
// A transport attaches it to the returned error (via [FaultError]) and also maps
// the fault onto a sentinel above where one fits.
type Problem struct {
	// Type is a URI naming the problem kind.
	Type string `json:"type"`
	// Title is a short human summary.
	Title string `json:"title"`
	// Status is the HTTP status code.
	Status int `json:"status"`
	// Code is a stable, machine-readable problem code.
	Code string `json:"code"`
	// Retry says whether a retry can help: "never", "after-state-change",
	// or "after-delay".
	Retry string `json:"retry"`
	// Detail is an optional human detail.
	Detail string `json:"detail,omitempty"`
	// Instance is an optional URI for the specific occurrence.
	Instance string `json:"instance,omitempty"`
}

// FaultError wraps a [Problem] as a Go error. A transport returns one for a
// fault that no sentinel above classifies, and wraps the matching sentinel for
// one that does, so errors.Is(err, ErrNotFound) and errors.As(err, &fe) both
// work off the one returned value.
type FaultError struct {
	Problem Problem
	// wrapped is the sentinel this fault also matches (via Unwrap), when one fits.
	wrapped error
}

// NewFaultError builds a FaultError for p, chaining sentinel so errors.Is finds
// it. sentinel may be nil when no shared sentinel fits the fault.
func NewFaultError(p Problem, sentinel error) *FaultError {
	return &FaultError{Problem: p, wrapped: sentinel}
}

func (e *FaultError) Error() string {
	if e.Problem.Code != "" {
		return "bdp fault: " + e.Problem.Code + " (" + e.Problem.Title + ")"
	}
	return "bdp fault: " + e.Problem.Title
}

// Unwrap exposes the shared sentinel this fault maps onto, so errors.Is matches
// it. It returns nil when the fault maps to no sentinel.
func (e *FaultError) Unwrap() error { return e.wrapped }
