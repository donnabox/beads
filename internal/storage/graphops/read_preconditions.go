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

var (
	errReadLeaseExpired = errors.Join(errors.New("own graph lease expired"), graph.ErrNotAuthority)
	errReadLeaseBudget  = errors.Join(errors.New("graph lease cannot cover read deadline"), graph.ErrNotAuthority)
)

// checkReadFacts compares already-decoded observations. It performs no I/O,
// creates no grant, repairs nothing and never advances a witness. Expected
// database/branch and local installation are always required, including on a
// diagnostic call without an active witness. A recorded sequence is required
// only with an active witness; nil is legitimate for an absent/pending witness.
// Binding precedes Scope absence. State validation is last and belongs outside
// the held tx; private expired/budget errors never authorize a regrant here.
func checkReadFacts(ctx context.Context, want expectedReadFacts, got preconditionObservations) error {
	if err := observationContext(ctx); err != nil {
		return err
	}
	if err := checkReadOperands(want); err != nil {
		return err
	}
	_, databaseErr := observationText(got.state.database, 1024)
	_, branchErr := observationText(got.state.branch, 1024)
	if errors.Join(databaseErr, branchErr) != nil {
		return corrupt(errors.New("missing decoded database or branch"))
	}
	if got.state.database != want.database || got.state.branch != want.branch {
		return graph.ErrNotAuthority
	}
	if err := checkReadObservationShapes(got); err != nil {
		return err
	}
	if got.scope.presence == observationAbsent {
		return graph.ErrNoScope
	}
	if (want.pending || want.witness != nil) && want.witnessInstallationKey != want.installationKey {
		return authority.ErrWrongInstallation
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
	grantWant, err := time.Parse(time.RFC3339Nano, w.GrantedAt)
	// DATETIME(6) requires the provider to persist witness and lease at the
	// same microsecond precision; do not normalize a malformed expectation.
	if err != nil || grantWant.Nanosecond()%1000 != 0 {
		return authority.ErrInvalidRecord
	}
	if w.Unverified {
		return authority.ErrWitnessUnverified
	}
	if got.scope.url != w.ScopeURL || got.scope.authorityID != w.AuthorityID || got.scope.epoch > w.Epoch {
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
	// An older same-authority epoch may be a restored state. Diagnose its
	// missing prefix above before refusing an otherwise inconsistent epoch.
	if got.scope.epoch != w.Epoch {
		return graph.ErrNotAuthority
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
	if !grant.Equal(grantWant) {
		return graph.ErrNotAuthority
	}
	deadline, _ := ctx.Deadline() // observationContext requires it.
	if err := checkReadLeaseBudget(l, deadline, time.Now()); err != nil {
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

func checkReadOperands(want expectedReadFacts) error {
	_, databaseErr := observationText(want.database, 1024)
	_, branchErr := observationText(want.branch, 1024)
	_, keyErr := observationLabel(want.installationKey, 64, 'f')
	var witnessKeyErr error
	if want.witness != nil || want.pending || want.witnessInstallationKey != "" {
		_, witnessKeyErr = observationLabel(want.witnessInstallationKey, 64, 'f')
	}
	if errors.Join(databaseErr, branchErr, keyErr, witnessKeyErr) != nil {
		return errObservationOperand
	}
	return nil
}

// Observers own scalar parsing; these checks detect missing/zero observations
// or a bad operand supplied by composition, without turning them into a public
// authority or rewind diagnosis. No complete schema/ledger census is implied.
func checkReadObservationShapes(got preconditionObservations) error {
	for _, p := range []observationPresence{got.scope.presence, got.ledger.recorded.presence, got.ledger.tip.presence, got.lease.presence} {
		if p != observationPresent && p != observationAbsent {
			return corrupt(errors.New("missing decoded precondition observation"))
		}
	}
	if got.ledger.requested != nil && (*got.ledger.requested == 0 || *got.ledger.requested > graph.MaxLedgerSeq) {
		return errObservationOperand
	}
	_, versionErr := observationLabel(got.state.version, 64, 'f')
	_, headErr := observationLabel(got.ledger.head, 32, 'v')
	if errors.Join(versionErr, headErr) != nil {
		return corrupt(errors.New("missing decoded state or HEAD label"))
	}
	for _, point := range []ledgerPoint{got.ledger.recorded, got.ledger.tip} {
		if point.presence == observationAbsent {
			if point.seq != 0 || point.hash != "" {
				return corrupt(errors.New("payload on absent ledger observation"))
			}
		} else if _, err := decodeLedgerPoint(point.seq, point.hash); err != nil {
			return err
		}
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

func checkReadLeaseBudget(l leaseObservation, deadline, now time.Time) error {
	clock, clockErr := readLeaseInstant(l.clock)
	expiry, expiryErr := readLeaseInstant(l.expiresAt)
	_, heartbeatErr := readLeaseInstant(l.heartbeatAt)
	if err := errors.Join(clockErr, expiryErr, heartbeatErr); err != nil {
		return err
	}
	// Renewers are informational and the fence changes on every lease write;
	// neither is a stable equality binding for a protected read.
	// Heartbeat is an application-clock diagnostic; its relation to grant or
	// database NOW is not a read precondition. Only its time domain is checked.
	if !expiry.After(clock) {
		return errReadLeaseExpired
	}
	// Refuse timing without monotonic provenance. Round(0) strips only that
	// component; this tests presence without reading Go's private time layout.
	// No application wall clock is compared with the database civil clock.
	if !hasMonotonic(l.queryStarted) || !hasMonotonic(l.queryFinished) || !hasMonotonic(deadline) || !hasMonotonic(now) ||
		l.queryFinished.Before(l.queryStarted) || now.Before(l.queryFinished) || !deadline.After(now) {
		return graph.ErrNotAuthority
	}
	// Qualified legs must sample NOW(6) at statement start, not transaction
	// start. Charging deadline-queryStarted
	// covers the whole query, scanning/close, the later state query, comparison
	// and all remaining work. Subtract one microsecond for clock quantization.
	// Strict inequality refuses equality at the lease boundary. A serving
	// provider must also cap the per-row deadline below one third of its TTL.
	// Suspend/clock discontinuities are not qualified by this arithmetic; the
	// provider must separately establish its clock/deadline behavior.
	remaining := expiry.Sub(clock) - time.Microsecond
	if remaining <= 0 || deadline.Sub(l.queryStarted) >= remaining {
		return errReadLeaseBudget
	}
	return nil
}
