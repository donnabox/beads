package graphops

import (
	"context"
	"errors"
	"time"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage/authority"
)

// expectedReadFacts contains copied administrative facts, not a LeaseClaim.
// A caller must separately own a qualified provider claim, load the witness for
// every call, and obtain all observations in the body transaction. These values
// alone cannot establish those obligations, even when checkReadFacts returns nil.
// The current slice has no provider accessor or production caller.
type expectedReadFacts struct {
	witness                                 *authority.Witness
	pending                                 bool
	installationKey, witnessInstallationKey string
	database, branch                        string
}

// checkReadFacts compares already-decoded observations. It performs no I/O,
// creates no grant, repairs nothing and never advances a witness. Refusal order
// keeps an absent Scope distinct, then checks witness/binding, ledger prefix,
// lease and finally state change. State validation belongs outside the held tx.
func checkReadFacts(ctx context.Context, want expectedReadFacts, got preconditionObservations) error {
	if err := observationContext(ctx); err != nil {
		return err
	}
	if got.scope.presence == observationAbsent {
		return graph.ErrNoScope
	}
	if got.scope.presence != observationPresent {
		return corrupt(errors.New("missing decoded Scope observation"))
	}
	if want.pending {
		return authority.ErrWitnessPending
	}
	if want.witness == nil {
		return graph.ErrNotAuthority
	}
	w := want.witness.Fields()
	// NewWitness checks only value shape, never provenance or permission.
	if _, err := authority.NewWitness(w); err != nil {
		return err
	}
	if w.Unverified {
		return authority.ErrWitnessUnverified
	}
	if _, err := observationLabel(want.installationKey, 64, 'f'); err != nil {
		return errObservationOperand
	}
	if want.witnessInstallationKey != want.installationKey {
		return authority.ErrWrongInstallation
	}
	if want.database == "" || want.branch == "" {
		return errObservationOperand
	}
	if got.state.database != want.database || got.state.branch != want.branch ||
		got.scope.url != w.ScopeURL || got.scope.authorityID != w.AuthorityID || got.scope.epoch != w.Epoch {
		return graph.ErrNotAuthority
	}
	if got.ledger.requested == nil || *got.ledger.requested != w.LedgerSeq {
		return errObservationOperand
	}
	if got.ledger.recorded.presence != observationPresent || got.ledger.recorded.seq != w.LedgerSeq ||
		got.ledger.recorded.hash != w.LedgerHash || got.ledger.tip.presence != observationPresent ||
		got.ledger.tip.seq < w.LedgerSeq {
		return graph.ErrStateRewound
	}
	l := got.lease
	if l.presence != observationPresent || l.scopeURL != w.ScopeURL || l.authorityID != w.AuthorityID ||
		l.epoch != w.Epoch || l.holder != want.installationKey || l.zone != "+00:00" {
		return graph.ErrNotAuthority
	}
	grant, err := readLeaseInstant(l.grantedAt)
	if err != nil {
		return err
	}
	if grant.Format(time.RFC3339Nano) != w.GrantedAt {
		return graph.ErrNotAuthority
	}
	deadline, _ := ctx.Deadline() // observationContext requires it.
	if err := checkReadLeaseBudget(l, grant, deadline, time.Now()); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if got.state.version != w.StateVersion {
		return graph.ErrStateChanged
	}
	return nil
}

// The observer already validated calendar/precision and retained a civil value.
// Only call after observing exactly +00:00 on a separately qualified UTC writer
// leg. This conversion cannot detect a historical non-UTC writer. Authority's
// time domain is strictly after the Unix epoch, even though raw observations may
// retain earlier dates for diagnostics.
func readLeaseInstant(value civilTimestamp) (time.Time, error) {
	t, err := time.Parse("2006-01-02 15:04:05.000000", string(value))
	if err != nil || !t.After(time.Unix(0, 0)) {
		return time.Time{}, corrupt(errors.New("lease timestamp outside authority time domain"))
	}
	return t, nil
}

func hasMonotonic(t time.Time) bool { return t != t.Round(0) }

func checkReadLeaseBudget(l leaseObservation, grant, deadline, now time.Time) error {
	clock, clockErr := readLeaseInstant(l.clock)
	expiry, expiryErr := readLeaseInstant(l.expiresAt)
	heartbeat, heartbeatErr := readLeaseInstant(l.heartbeatAt)
	if err := errors.Join(clockErr, expiryErr, heartbeatErr); err != nil {
		return err
	}
	// Renewers are informational and the fence changes on every lease write;
	// neither is a stable equality binding for a protected read.
	if grant.After(heartbeat) || heartbeat.After(clock) || !expiry.After(clock) {
		return graph.ErrNotAuthority
	}
	// Refuse timing without monotonic provenance. Round(0) strips only that
	// component; this tests presence without reading Go's private time layout.
	// No application wall clock is compared with the database civil clock.
	if !hasMonotonic(l.queryStarted) || !hasMonotonic(l.queryFinished) || !hasMonotonic(deadline) || !hasMonotonic(now) ||
		l.queryFinished.Before(l.queryStarted) || now.Before(l.queryFinished) || !deadline.After(now) {
		return graph.ErrNotAuthority
	}
	// NOW(6) is sampled after queryStarted. Charging deadline-queryStarted
	// covers the whole query, scanning/close, the later state query, comparison
	// and all remaining work. Subtract one microsecond for clock quantization.
	// Strict inequality refuses equality at the lease boundary. A serving
	// provider must also cap the per-row deadline below one third of its TTL.
	remaining := expiry.Sub(clock) - time.Microsecond
	if remaining <= 0 || deadline.Sub(l.queryStarted) >= remaining {
		return graph.ErrNotAuthority
	}
	return nil
}
