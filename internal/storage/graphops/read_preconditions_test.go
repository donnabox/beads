package graphops

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage/authority"
)

func readFactFixture(t *testing.T) (context.Context, expectedReadFacts, preconditionObservations) {
	t.Helper()
	start := time.Now()
	ctx, cancel := context.WithDeadline(t.Context(), start.Add(5*time.Second))
	t.Cleanup(cancel)
	w, err := authority.NewWitness(authority.WitnessFields{
		ScopeURL: "https://example.com/scope/", AuthorityID: strings.Repeat("a", 32), Epoch: 9,
		LedgerSeq: 3, LedgerHash: strings.Repeat("b", 64), StateVersion: strings.Repeat("c", 64),
		StateCommit: strings.Repeat("v", 32), GrantedAt: "2026-09-20T12:00:00.123456Z",
	})
	if err != nil {
		t.Fatal(err)
	}
	key := strings.Repeat("d", 64)
	want := expectedReadFacts{witness: &w, installationKey: key, witnessInstallationKey: key, database: "graph_db", branch: "main"}
	seq := uint64(3)
	got := preconditionObservations{
		scope: scopeObservation{presence: observationPresent, url: w.Fields().ScopeURL, authorityID: w.Fields().AuthorityID, epoch: 9},
		ledger: ledgerObservation{requested: &seq,
			recorded: ledgerPoint{observationPresent, 3, strings.Repeat("b", 64)},
			tip:      ledgerPoint{observationPresent, 8, strings.Repeat("e", 64)}, head: strings.Repeat("u", 32)},
		lease: leaseObservation{presence: observationPresent, scopeURL: w.Fields().ScopeURL,
			authorityID: w.Fields().AuthorityID, holder: key, renewer: strings.Repeat("f", 32), epoch: 9,
			grantedAt: "2026-09-20 12:00:00.123456", heartbeatAt: "2026-09-20 12:00:01.000000",
			clock: "2026-09-20 12:00:02.000000", expiresAt: "2026-09-20 12:00:20.000000",
			fence: strings.Repeat("0", 32), zone: "+00:00", queryStarted: start, queryFinished: start},
		state: stateObservation{database: "graph_db", branch: "main", version: strings.Repeat("c", 64)},
	}
	return ctx, want, got
}

func replaceReadWitness(t *testing.T, want *expectedReadFacts, change func(*authority.WitnessFields)) {
	t.Helper()
	fields := want.witness.Fields()
	change(&fields)
	w, err := authority.NewWitness(fields)
	if err != nil {
		t.Fatal(err)
	}
	want.witness = &w
}

func TestReadFactsMatchingValuesDoNotRequireUnchangedHeadOrRenewal(t *testing.T) {
	ctx, want, got := readFactFixture(t)
	// These independently supplied pure facts deliberately permit a later tip
	// with unchanged state.version. Real observers hash the ledger, so that
	// combination cannot occur in an actual unchanged database. A later HEAD
	// alone is a raw fact, not a failed ancestry check.
	// A renewal may change these informational/CAS cells without stealing.
	got.lease.renewer = strings.Repeat("1", 32)
	got.lease.fence = strings.Repeat("2", 32)
	if err := checkReadFacts(ctx, want, got); err != nil {
		t.Fatal(err)
	}
	got.ledger.tip = got.ledger.recorded
	got.ledger.head = want.witness.Fields().StateCommit
	if err := checkReadFacts(ctx, want, got); err != nil {
		t.Fatal(err)
	}
}

func TestReadFactsRefuseDistinctWitnessAndScopeStates(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*expectedReadFacts, *preconditionObservations)
		want   error
	}{
		{"no Scope", func(w *expectedReadFacts, o *preconditionObservations) {
			o.scope = scopeObservation{presence: observationAbsent}
			w.witness = nil
		}, graph.ErrNoScope},
		{"zero Scope observation", func(_ *expectedReadFacts, o *preconditionObservations) { o.scope = scopeObservation{} }, errCorrupt},
		{"absent witness", func(w *expectedReadFacts, _ *preconditionObservations) { w.witness = nil }, graph.ErrNotAuthority},
		{"pending active witness", func(w *expectedReadFacts, _ *preconditionObservations) { w.pending = true }, authority.ErrWitnessPending},
		{"pending without active witness", func(w *expectedReadFacts, _ *preconditionObservations) { w.pending = true; w.witness = nil }, authority.ErrWitnessPending},
		{"zero witness value", func(w *expectedReadFacts, _ *preconditionObservations) { w.witness = new(authority.Witness) }, authority.ErrInvalidRecord},
		{"unverified", func(w *expectedReadFacts, _ *preconditionObservations) {
			replaceReadWitness(t, w, func(f *authority.WitnessFields) { f.Unverified = true })
		}, authority.ErrWitnessUnverified},
		{"foreign installation", func(w *expectedReadFacts, _ *preconditionObservations) {
			w.witnessInstallationKey = strings.Repeat("e", 64)
		}, authority.ErrWrongInstallation},
		{"missing local installation", func(w *expectedReadFacts, _ *preconditionObservations) { w.installationKey = "" }, errObservationOperand},
		{"scope URL", func(_ *expectedReadFacts, o *preconditionObservations) { o.scope.url = "https://example.com/other/" }, graph.ErrNotAuthority},
		{"scope authority", func(_ *expectedReadFacts, o *preconditionObservations) { o.scope.authorityID = strings.Repeat("b", 32) }, graph.ErrNotAuthority},
		{"scope epoch", func(_ *expectedReadFacts, o *preconditionObservations) { o.scope.epoch++ }, graph.ErrNotAuthority},
		{"database", func(_ *expectedReadFacts, o *preconditionObservations) { o.state.database = "other" }, graph.ErrNotAuthority},
		{"branch", func(_ *expectedReadFacts, o *preconditionObservations) { o.state.branch = "feature" }, graph.ErrNotAuthority},
		{"unbound database", func(w *expectedReadFacts, _ *preconditionObservations) { w.database = "" }, errObservationOperand},
		{"unbound branch", func(w *expectedReadFacts, _ *preconditionObservations) { w.branch = "" }, errObservationOperand},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, want, got := readFactFixture(t)
			tc.change(&want, &got)
			got.state.version = strings.Repeat("0", 64) // earlier refusal must win.
			if err := checkReadFacts(ctx, want, got); !errors.Is(err, tc.want) || errors.Is(err, graph.ErrStateChanged) {
				t.Fatalf("got %v want %v", err, tc.want)
			}
		})
	}
	for _, e := range []error{authority.ErrWitnessPending, authority.ErrWitnessUnverified, authority.ErrWrongInstallation} {
		if !errors.Is(e, graph.ErrNotAuthority) {
			t.Fatalf("lost public refusal classification: %v", e)
		}
	}
}

func TestReadFactsExactLedgerPrefix(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*ledgerObservation)
		want   error
	}{
		{"not requested", func(l *ledgerObservation) { l.requested = nil }, errObservationOperand},
		{"wrong operand", func(l *ledgerObservation) { l.requested = new(uint64(4)) }, errObservationOperand},
		{"recorded absent with later tip", func(l *ledgerObservation) { l.recorded = ledgerPoint{presence: observationAbsent} }, graph.ErrStateRewound},
		{"different recorded seq", func(l *ledgerObservation) { l.recorded.seq++ }, graph.ErrStateRewound},
		{"different recorded hash", func(l *ledgerObservation) { l.recorded.hash = strings.Repeat("0", 64) }, graph.ErrStateRewound},
		{"absent tip", func(l *ledgerObservation) { l.tip = ledgerPoint{presence: observationAbsent} }, graph.ErrStateRewound},
		{"tip behind", func(l *ledgerObservation) { l.tip.seq = 2 }, graph.ErrStateRewound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, want, got := readFactFixture(t)
			tc.change(&got.ledger)
			got.lease.presence = observationAbsent
			got.state.version = strings.Repeat("0", 64)
			if err := checkReadFacts(ctx, want, got); !errors.Is(err, tc.want) {
				t.Fatalf("got %v want %v", err, tc.want)
			}
		})
	}
}

func TestReadFactsLeaseBindingAndTimeDomain(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*leaseObservation)
		want   error
	}{
		{"absent", func(l *leaseObservation) { l.presence = observationAbsent }, graph.ErrNotAuthority},
		{"holder", func(l *leaseObservation) { l.holder = strings.Repeat("e", 64) }, graph.ErrNotAuthority},
		{"scope", func(l *leaseObservation) { l.scopeURL = "https://example.com/other/" }, graph.ErrNotAuthority},
		{"authority", func(l *leaseObservation) { l.authorityID = strings.Repeat("b", 32) }, graph.ErrNotAuthority},
		{"epoch", func(l *leaseObservation) { l.epoch++ }, graph.ErrNotAuthority},
		{"grant", func(l *leaseObservation) { l.grantedAt = "2026-09-20 12:00:00.123457" }, graph.ErrNotAuthority},
		{"SYSTEM is not qualified UTC", func(l *leaseObservation) { l.zone = "SYSTEM" }, graph.ErrNotAuthority},
		{"offset", func(l *leaseObservation) { l.zone = "+05:00" }, graph.ErrNotAuthority},
		{"UTC alias not exact pin", func(l *leaseObservation) { l.zone = "UTC" }, graph.ErrNotAuthority},
		{"expired", func(l *leaseObservation) { l.expiresAt = "2026-09-20 12:00:01.999999" }, graph.ErrNotAuthority},
		{"equal clock", func(l *leaseObservation) { l.expiresAt = l.clock }, graph.ErrNotAuthority},
		{"pre-epoch clock", func(l *leaseObservation) { l.clock = "1960-01-01 00:00:00.000000" }, errCorrupt},
		{"epoch expiry", func(l *leaseObservation) { l.expiresAt = "1970-01-01 00:00:00.000000" }, errCorrupt},
		{"invalid heartbeat", func(l *leaseObservation) { l.heartbeatAt = "invalid" }, errCorrupt},
		{"invalid grant", func(l *leaseObservation) { l.grantedAt = "invalid" }, errCorrupt},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, want, got := readFactFixture(t)
			tc.change(&got.lease)
			got.state.version = strings.Repeat("0", 64)
			if err := checkReadFacts(ctx, want, got); !errors.Is(err, tc.want) || errors.Is(err, graph.ErrStateChanged) {
				t.Fatalf("got %v want %v", err, tc.want)
			}
		})
	}
}

func TestReadLeaseBudgetAccountsForQueryAndPostQueryElapsed(t *testing.T) {
	_, _, got := readFactFixture(t)
	start := time.Now()
	l := got.lease
	l.queryStarted = start
	l.queryFinished = start.Add(2 * time.Second)
	// DB expiry-clock is 18s; quantization leaves 17.999999s. Elapsed query
	// and later work must be charged, even when less time remains at check.
	for _, tc := range []struct {
		name        string
		deadline    time.Duration
		now         time.Duration
		wantRefusal bool
	}{
		{"inside", 17*time.Second + 999998*time.Microsecond, 3 * time.Second, false},
		{"equal conservative boundary", 17*time.Second + 999999*time.Microsecond, 3 * time.Second, true},
		{"equal raw expiry", 18 * time.Second, 3 * time.Second, true},
		{"query elapsed cannot be refunded", 19 * time.Second, 2 * time.Second, true},
		{"post-query elapsed cannot be refunded", 20 * time.Second, 10 * time.Second, true},
		{"deadline already reached before arithmetic", 3 * time.Second, 3 * time.Second, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := checkReadLeaseBudget(l, start.Add(tc.deadline), start.Add(tc.now))
			if tc.wantRefusal && !errors.Is(err, graph.ErrNotAuthority) || !tc.wantRefusal && err != nil {
				t.Fatalf("err=%v refusal=%t", err, tc.wantRefusal)
			}
		})
	}
}

func TestReadLeaseBudgetRequiresOrderedMonotonicOperands(t *testing.T) {
	_, _, got := readFactFixture(t)
	start := time.Now()
	for _, mode := range []string{"start wall-only", "finish wall-only", "deadline wall-only", "now wall-only", "negative query", "future finish"} {
		t.Run(mode, func(t *testing.T) {
			l := got.lease
			l.queryStarted, l.queryFinished = start, start.Add(time.Second)
			now, deadline := start.Add(2*time.Second), start.Add(5*time.Second)
			switch mode {
			case "start wall-only":
				l.queryStarted = l.queryStarted.Round(0)
			case "finish wall-only":
				l.queryFinished = l.queryFinished.Round(0)
			case "deadline wall-only":
				deadline = deadline.Round(0)
			case "now wall-only":
				now = now.Round(0)
			case "negative query":
				l.queryFinished = start.Add(-time.Second)
			case "future finish":
				l.queryFinished = now.Add(time.Second)
			}
			if err := checkReadLeaseBudget(l, deadline, now); !errors.Is(err, graph.ErrNotAuthority) {
				t.Fatalf("accepted %s: %v", mode, err)
			}
		})
	}
}

func TestReadFactsDeadlineAndStateChange(t *testing.T) {
	_, want, got := readFactFixture(t)
	for _, ctx := range []context.Context{nil, context.Background()} {
		if err := checkReadFacts(ctx, want, got); !errors.Is(err, errObservationDeadline) {
			t.Fatalf("missing deadline: %v", err)
		}
	}
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	cancel()
	if err := checkReadFacts(ctx, want, got); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation lost: %v", err)
	}
	ctx, want, got = readFactFixture(t)
	got.state.version = strings.Repeat("0", 64)
	if err := checkReadFacts(ctx, want, got); !errors.Is(err, graph.ErrStateChanged) || errors.Is(err, graph.ErrNotAuthority) {
		t.Fatalf("state classification: %v", err)
	}
	// Civil DB time may be far from application wall time; only the DB's
	// own interval and local monotonic duration are compared.
	ctx, want, got = readFactFixture(t)
	replaceReadWitness(t, &want, func(w *authority.WitnessFields) { w.GrantedAt = "1999-01-01T00:00:00Z" })
	got.lease.grantedAt, got.lease.heartbeatAt = "1999-01-01 00:00:00.000000", "1999-01-01 00:00:01.000000"
	got.lease.clock, got.lease.expiresAt = "1999-01-01 00:00:02.000000", "1999-01-01 00:00:20.000000"
	if err := checkReadFacts(ctx, want, got); err != nil {
		t.Fatalf("application wall clock leaked into lease comparison: %v", err)
	}
}

func TestReadFactsBindingPrecedesNoScope(t *testing.T) {
	for _, mode := range []string{"active", "pending", "absent"} {
		t.Run(mode, func(t *testing.T) {
			ctx, want, got := readFactFixture(t)
			got.scope = scopeObservation{presence: observationAbsent}
			if mode == "pending" {
				want.pending = true
			}
			if mode == "absent" {
				want.witness = nil
				want.witnessInstallationKey = ""
				got.ledger.requested = nil
			}
			if err := checkReadFacts(ctx, want, got); !errors.Is(err, graph.ErrNoScope) {
				t.Fatalf("bound absent Scope: %v", err)
			}
			for _, field := range []string{"database", "branch"} {
				wrong := got
				if field == "database" {
					wrong.state.database = "other"
				} else {
					wrong.state.branch = "feature"
				}
				if err := checkReadFacts(ctx, want, wrong); !errors.Is(err, graph.ErrNotAuthority) || errors.Is(err, graph.ErrNoScope) {
					t.Fatalf("absent on wrong %s: %v", field, err)
				}
			}
		})
	}
}

func TestReadFactsDistinguishesRestoredAndSupersedingEpochs(t *testing.T) {
	for _, tc := range []struct {
		name    string
		epoch   uint64
		missing bool
		want    error
	}{
		{"restore lost prefix", 8, true, graph.ErrStateRewound},
		{"older epoch retained prefix is inconsistent", 8, false, graph.ErrNotAuthority},
		{"superseding epoch lost prefix", 10, true, graph.ErrNotAuthority},
		{"superseding epoch retained prefix", 10, false, graph.ErrNotAuthority},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, want, got := readFactFixture(t)
			got.scope.epoch = tc.epoch
			if tc.missing {
				got.ledger.recorded = ledgerPoint{presence: observationAbsent}
				got.ledger.tip = ledgerPoint{observationPresent, 2, strings.Repeat("e", 64)}
			}
			if err := checkReadFacts(ctx, want, got); !errors.Is(err, tc.want) {
				t.Fatalf("got %v want %v", err, tc.want)
			}
		})
	}
}

func TestReadFactsForeignInstallationPrecedesWitnessState(t *testing.T) {
	for _, mode := range []string{"pending", "pending without active", "unverified"} {
		t.Run(mode, func(t *testing.T) {
			ctx, want, got := readFactFixture(t)
			want.witnessInstallationKey = strings.Repeat("e", 64)
			if mode == "unverified" {
				replaceReadWitness(t, &want, func(w *authority.WitnessFields) { w.Unverified = true })
			} else {
				want.pending = true
				if mode == "pending without active" {
					want.witness = nil
					got.ledger.requested = nil
				}
			}
			if err := checkReadFacts(ctx, want, got); !errors.Is(err, authority.ErrWrongInstallation) {
				t.Fatalf("foreign %s: %v", mode, err)
			}
		})
	}
}

func TestReadFactsCompositionOperandsAndZeroObservations(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*expectedReadFacts, *preconditionObservations)
		want   error
	}{
		{"missing database before absent Scope", func(w *expectedReadFacts, o *preconditionObservations) {
			w.database = ""
			o.scope = scopeObservation{presence: observationAbsent}
		}, errObservationOperand},
		{"invalid local key before pending", func(w *expectedReadFacts, _ *preconditionObservations) { w.installationKey = "bad"; w.pending = true }, errObservationOperand},
		{"invalid witness key before pending", func(w *expectedReadFacts, _ *preconditionObservations) {
			w.witnessInstallationKey = ""
			w.pending = true
		}, errObservationOperand},
		{"zero requested sequence", func(_ *expectedReadFacts, o *preconditionObservations) { o.ledger.requested = new(uint64) }, errObservationOperand},
		{"zero recorded presence", func(_ *expectedReadFacts, o *preconditionObservations) { o.ledger.recorded = ledgerPoint{} }, errCorrupt},
		{"zero tip presence", func(_ *expectedReadFacts, o *preconditionObservations) { o.ledger.tip = ledgerPoint{} }, errCorrupt},
		{"zero lease presence", func(_ *expectedReadFacts, o *preconditionObservations) { o.lease = leaseObservation{} }, errCorrupt},
		{"zero state version", func(_ *expectedReadFacts, o *preconditionObservations) { o.state.version = "" }, errCorrupt},
		{"zero observed database", func(_ *expectedReadFacts, o *preconditionObservations) { o.state.database = "" }, errCorrupt},
		{"zero HEAD", func(_ *expectedReadFacts, o *preconditionObservations) { o.ledger.head = "" }, errCorrupt},
		{"absent ledger payload", func(_ *expectedReadFacts, o *preconditionObservations) {
			o.ledger.recorded.presence = observationAbsent
		}, errCorrupt},
		{"zero present ledger seq", func(_ *expectedReadFacts, o *preconditionObservations) { o.ledger.recorded.seq = 0 }, errCorrupt},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, want, got := readFactFixture(t)
			tc.change(&want, &got)
			if err := checkReadFacts(ctx, want, got); !errors.Is(err, tc.want) {
				t.Fatalf("got %v want %v", err, tc.want)
			}
		})
	}
	for _, pending := range []bool{false, true} {
		ctx, want, got := readFactFixture(t)
		want.witness = nil
		want.pending = pending
		got.ledger.requested = nil // diagnostic observation, not a missing active-witness lookup
		wantErr := graph.ErrNotAuthority
		if pending {
			wantErr = authority.ErrWitnessPending
		}
		if err := checkReadFacts(ctx, want, got); !errors.Is(err, wantErr) || errors.Is(err, errObservationOperand) {
			t.Fatalf("legitimate nil diagnostic operand: %v", err)
		}
	}
}

func TestReadFactsGrantPrecisionAndInformationalHeartbeat(t *testing.T) {
	ctx, want, got := readFactFixture(t)
	replaceReadWitness(t, &want, func(w *authority.WitnessFields) { w.GrantedAt = "2026-09-20T12:00:00.123456789Z" })
	if err := checkReadFacts(ctx, want, got); !errors.Is(err, authority.ErrInvalidRecord) {
		t.Fatalf("submicrosecond expectation: %v", err)
	}
	for _, heartbeat := range []civilTimestamp{"2026-09-20 11:59:50.000000", "2026-09-20 12:00:30.000000"} {
		ctx, want, got := readFactFixture(t)
		// Simulate renewal with independent application-clock heartbeat. The
		// DB owns expiry/clock; the unchanged witness still names this grant.
		got.lease.heartbeatAt = heartbeat
		got.lease.expiresAt = "2026-09-20 12:01:00.000000"
		got.lease.fence = strings.Repeat("1", 32)
		got.lease.renewer = strings.Repeat("2", 32)
		if err := checkReadFacts(ctx, want, got); err != nil {
			t.Fatalf("informational heartbeat %s blocked renewal: %v", heartbeat, err)
		}
	}
}

func TestReadFactsPrivateLeaseRefusalsAndDeadlinePlumbing(t *testing.T) {
	ctx, want, got := readFactFixture(t)
	got.lease.expiresAt = got.lease.clock
	if err := checkReadFacts(ctx, want, got); !errors.Is(err, errReadLeaseExpired) || !errors.Is(err, graph.ErrNotAuthority) {
		t.Fatalf("own expired classification: %v", err)
	}
	got.lease.holder = strings.Repeat("e", 64)
	if err := checkReadFacts(ctx, want, got); !errors.Is(err, graph.ErrNotAuthority) || errors.Is(err, errReadLeaseExpired) {
		t.Fatalf("foreign expired must not suggest self-regrant: %v", err)
	}
	_, want, got = readFactFixture(t)
	ctx, cancel := context.WithDeadline(t.Context(), got.lease.queryStarted.Add(30*time.Second))
	defer cancel()
	if err := checkReadFacts(ctx, want, got); !errors.Is(err, errReadLeaseBudget) || !errors.Is(err, graph.ErrNotAuthority) || errors.Is(err, errReadLeaseExpired) {
		t.Fatalf("actual context deadline exceeds live lease: %v", err)
	}
}
