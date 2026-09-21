package graphops

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/beadsignore"
	"github.com/steveyegge/beads/internal/storage/authority"
)

// Cooperative mechanical fixture only. The sqlmock transaction is independently
// closed by mockTx's cleanup; these hooks do not qualify physical engine cleanup.
type attemptResourceFixture struct {
	queryer
	workCtx     context.Context
	events      []string
	beforeQuery func(int)
	queries     int
	rollback    func(context.Context) error
	connection  func(context.Context) error
}

func (r *attemptResourceFixture) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if ctx != r.workCtx {
		return nil, errors.New("changed exact work context")
	}
	r.queries++
	r.events = append(r.events, "query")
	if r.beforeQuery != nil {
		r.beforeQuery(r.queries)
	}
	return r.queryer.QueryContext(ctx, query, args...)
}
func (r *attemptResourceFixture) rollbackRead(ctx context.Context) error {
	r.events = append(r.events, "rollback")
	if r.rollback != nil {
		return r.rollback(ctx)
	}
	return nil
}
func (r *attemptResourceFixture) closeReadConnection(ctx context.Context) error {
	r.events = append(r.events, "connection")
	if r.connection != nil {
		return r.connection(ctx)
	}
	return nil
}
func attemptFixture(t *testing.T) (context.Context, expectedReadFacts, *attemptResourceFixture, sqlmock.Sqlmock) {
	t.Helper()
	ctx, want, _ := readFactFixture(t)
	version, err := composeStateVersion(stateLabels())
	if err != nil {
		t.Fatal(err)
	}
	replaceReadWitness(t, &want, func(w *authority.WitnessFields) { w.StateVersion = version })
	tx, m := mockTx(t)
	return ctx, want, &attemptResourceFixture{queryer: tx, workCtx: ctx}, m
}
func expectAttemptFacts(m sqlmock.Sqlmock, want expectedReadFacts, noScope bool) {
	var seq driver.Value
	if want.witness != nil {
		seq = strconv.FormatUint(want.witness.Fields().LedgerSeq, 10)
	}
	scope := sqlmock.NewRows(scopeObservationColumns)
	if !noScope {
		scope.AddRow(int64(1), "https://example.com/scope/", strings.Repeat("a", 32), int64(9), observationTime)
	}
	m.ExpectQuery(regexp.QuoteMeta(scopeObservationQuery)).WillReturnRows(scope).RowsWillBeClosed()
	var recorded, recordedHash driver.Value
	if seq != nil {
		recorded = int64(3)
		recordedHash = strings.Repeat("b", 64)
	}
	m.ExpectQuery(regexp.QuoteMeta(ledgerObservationQuery)).WithArgs(seq).WillReturnRows(sqlmock.NewRows(ledgerObservationColumns).AddRow(recorded, recordedHash, int64(8), strings.Repeat("e", 64), strings.Repeat("u", 32))).RowsWillBeClosed()
	m.ExpectQuery(regexp.QuoteMeta(leaseObservationQuery)).WillReturnRows(sqlmock.NewRows(leaseObservationColumns).AddRow(int64(1), int64(1), "https://example.com/scope/", strings.Repeat("a", 32), want.installationKey, strings.Repeat("f", 32), int64(9), "2026-09-20 12:00:00.123456", "2026-09-20 12:00:20.000000", "2026-09-20 12:00:01.000000", strings.Repeat("0", 32), "2026-09-20 12:00:02.000000", "+00:00")).RowsWillBeClosed()
	values := stateValues()
	values[0] = "graph_db"
	m.ExpectQuery(regexp.QuoteMeta(stateObservationQuery)).WillReturnRows(sqlmock.NewRows(stateColumns).AddRow(values...)).RowsWillBeClosed()
}
func attemptBody(r *attemptResourceFixture, result int, err error) func(context.Context, queryer) (int, error) {
	return func(ctx context.Context, q queryer) (int, error) {
		r.events = append(r.events, "body")
		if ctx != r.workCtx || q != r {
			return 0, errors.New("body resource/context mismatch")
		}
		return result, err
	}
}
func TestReadAttemptStageOrdering(t *testing.T) {
	for _, mode := range []string{"success", "observation", "state", "body-state", "pending", "absent", "no-scope", "pending-no-scope", "wrong-binding-no-scope", "invalid-facts"} {
		t.Run(mode, func(t *testing.T) {
			ctx, want, r, m := attemptFixture(t)
			var wantErr error
			var bodyErr error
			wantStage := readAttemptCompare
			bodyRuns := false
			noScope := strings.Contains(mode, "no-scope")
			switch mode {
			case "success":
				bodyRuns = true
				wantStage = readAttemptBody
			case "observation":
				wantErr = errors.New("observation failed")
				wantStage = readAttemptObserve
			case "state":
				replaceReadWitness(t, &want, func(w *authority.WitnessFields) { w.StateVersion = strings.Repeat("e", 64) })
				wantErr = graph.ErrStateChanged
			case "body-state":
				bodyRuns = true
				wantStage = readAttemptBody
				bodyErr = graph.ErrStateChanged
				wantErr = bodyErr
			case "pending", "pending-no-scope":
				want.pending = true
				wantErr = authority.ErrWitnessPending
			case "absent":
				want.witness = nil
				wantErr = graph.ErrNotAuthority
			case "no-scope":
				want.witness = nil
			case "invalid-facts":
				want.database = ""
				wantErr = errObservationOperand
			case "wrong-binding-no-scope":
				want.database = "other"
				wantErr = graph.ErrNotAuthority
			}
			if noScope && mode != "wrong-binding-no-scope" {
				wantErr = graph.ErrNoScope
			}
			if mode == "observation" {
				m.ExpectQuery(regexp.QuoteMeta(scopeObservationQuery)).WillReturnError(wantErr)
			} else {
				expectAttemptFacts(m, want, noScope)
			}
			result, c, err := runReadAttempt(ctx, ownReadAttempt(r), want, nil, attemptBody(r, 42, bodyErr))
			if !errors.Is(err, wantErr) || c.stage != wantStage || !c.cleanupOK || !c.workLive || c.retryEligible() != (mode == "state") {
				t.Fatalf("result=%d completion=%+v err=%v", result, c, err)
			}
			if (result == 42) != (mode == "success") {
				t.Fatal("partial result", result)
			}
			queries := 4
			if mode == "observation" {
				queries = 1
			}
			events := make([]string, queries)
			for i := range events {
				events[i] = "query"
			}
			if bodyRuns {
				events = append(events, "body")
			}
			events = append(events, "rollback", "connection")
			if !reflect.DeepEqual(r.events, events) {
				t.Fatalf("events %v want %v", r.events, events)
			}
		})
	}
}
func TestReadAttemptInvalidEntryStillCleans(t *testing.T) {
	boom := errors.New("partial acquisition failed")
	for _, mode := range []string{"nil-context", "unbounded", "canceled", "nil-body", "entry-error"} {
		t.Run(mode, func(t *testing.T) {
			ctx, want, r, _ := attemptFixture(t)
			r.rollback = func(c context.Context) error {
				if c.Err() != nil {
					t.Error("cleanup inherited cancellation")
				}
				if _, ok := c.Deadline(); !ok {
					t.Error("unbounded cleanup")
				}
				return nil
			}
			var entryErr error
			body := attemptBody(r, 1, nil)
			switch mode {
			case "nil-context":
				ctx = nil
			case "unbounded":
				ctx = context.Background()
			case "canceled":
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			case "nil-body":
				body = nil
			case "entry-error":
				entryErr = boom
			}
			result, c, err := runReadAttempt(ctx, ownReadAttempt(r), want, entryErr, body)
			if result != 0 || err == nil || c.stage != readAttemptEntry || !c.cleanupOK || !reflect.DeepEqual(r.events, []string{"rollback", "connection"}) {
				t.Fatalf("%d %+v %v events=%v", result, c, err, r.events)
			}
			if mode == "entry-error" && !errors.Is(err, boom) {
				t.Fatal(err)
			}
		})
	}
}
func TestReadAttemptOwnerTransfer(t *testing.T) {
	ctx, want, r, _ := attemptFixture(t)
	owner := ownReadAttempt(r)
	// Build a copied value without copying atomic noCopy internals; self still
	// refers to the original token and must refuse before touching the resource.
	copied := &readAttemptOwner{self: owner, resource: r}
	var typedNil *attemptResourceFixture
	entry := errors.Join(errors.New("acquisition failed"), graph.ErrStateChanged)
	for _, bad := range []*readAttemptOwner{nil, {}, copied, ownReadAttempt(nil), ownReadAttempt(typedNil)} {
		_, completion, err := runReadAttempt(ctx, bad, want, entry, attemptBody(r, 1, nil))
		if err == nil || !errors.Is(completion.diagnostic, entry) || !errors.Is(completion.diagnostic, err) || errors.Is(err, graph.ErrStateChanged) || len(r.events) != 0 {
			t.Fatal("invalid owner consumed resources", err, r.events)
		}
	}
	if _, _, err := runReadAttempt(ctx, owner, want, entry, attemptBody(r, 1, nil)); !errors.Is(err, entry) {
		t.Fatal(err)
	}
	if _, completion, err := runReadAttempt(ctx, owner, want, entry, attemptBody(r, 1, nil)); err == nil || !errors.Is(completion.diagnostic, entry) || errors.Is(err, graph.ErrStateChanged) {
		t.Fatal("double consume")
	}
	if !reflect.DeepEqual(r.events, []string{"rollback", "connection"}) {
		t.Fatal(r.events)
	}
}
func TestReadAttemptSuppressesDiscardedErrorChains(t *testing.T) {
	cleanup := errors.New("cleanup failed")
	absence := &exactAbsence{state: graph.AllocationPruned}
	gone := &graph.GoneError{}
	joined := errors.Join(graph.ErrStateChanged, graph.ErrNoScope, graph.ErrNotFound, absence, gone)
	for _, mode := range []string{"rollback", "connection", "both", "cancel", "success-rollback", "success-connection", "success-both", "success-cancel"} {
		t.Run(mode, func(t *testing.T) {
			ctx, want, r, m := attemptFixture(t)
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()
			r.workCtx = ctx
			failure := strings.TrimPrefix(mode, "success-")
			bodyErr := joined
			if strings.HasPrefix(mode, "success-") {
				bodyErr = nil
			}
			expectAttemptFacts(m, want, false)
			r.rollback = func(c context.Context) error {
				if c == ctx || c.Err() != nil {
					t.Error("cleanup context not independent")
				}
				deadline, ok := c.Deadline()
				if !ok || time.Until(deadline) > readAttemptCleanupTimeout {
					t.Error("cleanup unbounded")
				}
				if failure == "rollback" || failure == "both" {
					return cleanup
				}
				return nil
			}
			r.connection = func(context.Context) error {
				if failure == "cancel" {
					cancel()
				}
				if failure == "connection" || failure == "both" {
					return cleanup
				}
				return nil
			}
			result, c, err := runReadAttempt(ctx, ownReadAttempt(r), want, nil, attemptBody(r, 42, bodyErr))
			var gotAbsence *exactAbsence
			var gotGone *graph.GoneError
			if result != 0 || err == nil || errors.Is(err, graph.ErrStateChanged) || errors.Is(err, graph.ErrNoScope) || errors.Is(err, graph.ErrNotFound) || errors.As(err, &gotAbsence) || errors.As(err, &gotGone) || !errors.Is(c.diagnostic, bodyErr) || c.stage != readAttemptBody || c.retryEligible() {
				t.Fatalf("leaked result/cause: %d %+v %v", result, c, err)
			}
			if failure == "cancel" {
				if !errors.Is(err, context.Canceled) || c.workLive {
					t.Fatal(err, c)
				}
			} else if !errors.Is(err, cleanup) || c.cleanupOK {
				t.Fatal(err, c)
			}
			if r.events[len(r.events)-1] != "connection" {
				t.Fatal(r.events)
			}
		})
	}
}
func TestReadAttemptStateChangeRequiresCleanCompletion(t *testing.T) {
	for _, mode := range []string{"connection", "cancel"} {
		t.Run(mode, func(t *testing.T) {
			ctx, want, r, m := attemptFixture(t)
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()
			r.workCtx = ctx
			replaceReadWitness(t, &want, func(w *authority.WitnessFields) { w.StateVersion = strings.Repeat("e", 64) })
			expectAttemptFacts(m, want, false)
			r.connection = func(context.Context) error {
				if mode == "cancel" {
					cancel()
					return nil
				}
				return errors.New("close failed")
			}
			result, c, err := runReadAttempt(ctx, ownReadAttempt(r), want, nil, attemptBody(r, 42, nil))
			if result != 0 || err == nil || errors.Is(err, graph.ErrStateChanged) || !c.stateChanged || c.retryEligible() {
				t.Fatal(result, c, err)
			}
		})
	}
}
func TestReadAttemptCleanupAfterPanic(t *testing.T) {
	for _, mode := range []string{"body", "rollback", "connection", "body-and-rollback", "nil-body-panic"} {
		t.Run(mode, func(t *testing.T) {
			ctx, want, r, m := attemptFixture(t)
			expectAttemptFacts(m, want, false)
			workPanic := errors.New("work panic")
			cleanupPanic := errors.New("cleanup panic")
			if mode == "rollback" || mode == "body-and-rollback" {
				r.rollback = func(context.Context) error { panic(cleanupPanic) }
			}
			if mode == "connection" {
				r.connection = func(context.Context) error { panic(cleanupPanic) }
			}
			var got any
			returned := false
			func() {
				defer func() { got = recover() }()
				_, _, _ = runReadAttempt(ctx, ownReadAttempt(r), want, nil, func(context.Context, queryer) (int, error) {
					r.events = append(r.events, "body")
					if mode == "body" || mode == "body-and-rollback" {
						panic(workPanic)
					}
					if mode == "nil-body-panic" {
						panic(nil)
					}
					return 1, nil
				})
				returned = true
			}()
			if returned || got == nil || r.events[len(r.events)-1] != "connection" {
				t.Fatal("panic/cleanup lost", returned, got, r.events)
			}
			if mode == "body" || mode == "body-and-rollback" {
				if got != workPanic {
					t.Fatal(got)
				}
			} else if mode != "nil-body-panic" && got != cleanupPanic {
				t.Fatal(got)
			}
		})
	}
}
func TestReadAttemptCleanupBudgetAndSecondObligation(t *testing.T) {
	ctx, want, r, _ := attemptFixture(t)
	r.rollback = func(c context.Context) error { <-c.Done(); return nil }
	r.connection = func(c context.Context) error {
		if !errors.Is(c.Err(), context.DeadlineExceeded) {
			t.Error("not same expired cleanup context")
		}
		return nil
	}
	_, c, err := runReadAttempt(ctx, ownReadAttempt(r), want, errors.New("entry failed"), attemptBody(r, 1, nil))
	if c.cleanupOK || c.retryEligible() || !errors.Is(err, context.DeadlineExceeded) || !reflect.DeepEqual(r.events, []string{"rollback", "connection"}) {
		t.Fatal(c, err, r.events)
	}
}
func TestReadAttemptUsesCopiedLoadedSequence(t *testing.T) {
	ctx, want, r, m := attemptFixture(t)
	expectAttemptFacts(m, want, false)
	old := want.witness
	r.beforeQuery = func(n int) {
		if n == 1 {
			f := old.Fields()
			f.LedgerSeq = 99
			next, err := authority.NewWitness(f)
			if err != nil {
				t.Fatal(err)
			}
			*old = next
		}
	}
	result, c, err := runReadAttempt(ctx, ownReadAttempt(r), want, nil, attemptBody(r, 42, nil))
	if result != 42 || err != nil || c.retryEligible() {
		t.Fatal(result, c, err)
	}
}

// Test-only administrative facts drive the actual witness file lifecycle. They
// are synthetic and establish neither provider authority nor engine evidence.
type attemptWitnessEvidence struct{}

func (attemptWitnessEvidence) ObserveOwnership(context.Context) (authority.Ownership, error) {
	return authority.Ownership{Database: "fixture", Branch: "main", SessionGeneration: "synthetic", ObservationVersion: 1, Topology: authority.Solo, Workspace: authority.ExclusiveWorkspace}, nil
}
func (attemptWitnessEvidence) CompareAdvance(_ context.Context, r authority.CompareRequest) (authority.Comparison, error) {
	return authority.NewComparison(r, authority.CurrentContained)
}
func (attemptWitnessEvidence) VerifyClearUnverified(context.Context, authority.ClearRequest) (authority.ClearProof, error) {
	return authority.ClearProof{}, authority.ErrEvidence
}
func (attemptWitnessEvidence) PrepareTransition(_ context.Context, r authority.PrepareRequest) (authority.Preparation, error) {
	w := authority.WitnessFields{ScopeURL: "https://example.com/scope/", AuthorityID: strings.Repeat("a", 32), Epoch: 1, LedgerSeq: 3, LedgerHash: strings.Repeat("b", 64), StateVersion: strings.Repeat("c", 64), GrantedAt: "2026-09-20T12:00:00.123456Z"}
	p := authority.PreparationFields{PreStateVersion: authority.OptionalToken{Presence: authority.Absent}, PreHead: authority.OptionalToken{Presence: authority.Absent}, PreLedger: authority.LedgerState{Presence: authority.Absent}, PreScope: authority.ScopeState{Presence: authority.Absent}, PreLease: authority.LeaseState{Presence: authority.Absent}}
	for i := range p.WorkingTables {
		p.WorkingTables[i] = authority.TableState{Presence: authority.Absent}
		p.ResultTables[i] = authority.TableState{Presence: authority.Present, Schema: "fixture-schema", Content: "fixture-content"}
	}
	if prior, ok := r.Prior().Witness(); ok {
		w = prior.Fields()
		p.PreStateVersion = authority.OptionalToken{Presence: authority.Present, Value: w.StateVersion}
		p.PreHead = authority.OptionalToken{Presence: authority.Present, Value: w.StateCommit}
		p.PreLineage = strings.Repeat("4", 64)
		p.PreLedger = authority.LedgerState{Presence: authority.Present, Seq: w.LedgerSeq, Hash: w.LedgerHash}
		p.PreScope = authority.ScopeState{Presence: authority.Present, URL: w.ScopeURL, AuthorityID: w.AuthorityID, Epoch: w.Epoch}
		p.PreLease = attemptWitnessLease(r.InstallationKey(), w)
		p.WorkingTables = p.ResultTables
		w.LedgerSeq++
		w.LedgerHash = strings.Repeat("e", 64)
		w.StateVersion = strings.Repeat("f", 64)
		p.Install = &authority.InstallPreparation{ArtifactDigest: r.Intent().Fields().ArtifactDigest, Validated: authority.Verified}
	} else {
		p.Mint = &authority.MintPreparation{CatalogDigest: r.Intent().Fields().ArtifactDigest, LawfulTarget: authority.Verified}
	}
	w.StateCommit = ""
	p.ResultWitness = &w
	p.ResultLease = attemptWitnessLease(r.InstallationKey(), w)
	p.ResultLease.Fence = strings.Repeat("8", 32)
	return authority.NewPreparation(r, p)
}
func attemptWitnessLease(key string, w authority.WitnessFields) authority.LeaseState {
	return authority.LeaseState{Presence: authority.Present, Holder: key, ScopeURL: w.ScopeURL, AuthorityID: w.AuthorityID, Epoch: w.Epoch, Fence: strings.Repeat("1", 32), Renewer: strings.Repeat("2", 32), GrantedAt: w.GrantedAt, HeartbeatAt: w.GrantedAt, ExpiresAt: "2026-09-20T12:01:00.123456Z"}
}
func (attemptWitnessEvidence) InspectTransition(_ context.Context, r authority.InspectRequest) (authority.Inspection, error) {
	s := r.Snapshot()
	p := s.Preparation()
	w := p.ResultWitness
	return authority.NewInspection(r, authority.InspectionFields{State: authority.CompleteEvent, StateVersion: w.StateVersion, Tables: p.ResultTables, Scope: authority.ScopeState{Presence: authority.Present, URL: w.ScopeURL, AuthorityID: w.AuthorityID, Epoch: w.Epoch}, Ledger: authority.LedgerState{Presence: authority.Present, Seq: w.LedgerSeq, Hash: w.LedgerHash}, Lease: p.ResultLease, Head: authority.OptionalToken{Presence: authority.Present, Value: strings.Repeat("v", 32)}, OperationID: s.OperationID(), OperationCommit: strings.Repeat("v", 32), Complete: authority.Verified})
}
func TestReadAttemptPassiveIntakeLifecycle(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("durable witness operations require the existing Unix authority implementation")
	}
	root := t.TempDir()
	dir := filepath.Join(root, ".beads")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("BEADS_INSTALLATION_ID_FILE", filepath.Join(root, "installation-id"))
	t.Setenv("BDP_SCOPE_URL", "")
	t.Setenv("BD_BDP_SCOPE_URL", "")
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	manager, err := authority.New(ctx, dir)
	if err != nil {
		t.Fatal(err)
	}
	key, err := authority.InstallationKey(ctx, dir)
	if err != nil {
		t.Fatal(err)
	}
	binding := readAttemptBinding{key, "fixture", "main"}
	absent, err := passiveAttemptLoad(t, dir, ctx, manager, binding)
	if err != nil || absent.witness != nil || absent.pending || absent.witnessInstallationKey != "" {
		t.Fatal(absent, err)
	}
	guard, err := manager.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := guard.Close(); err != nil {
			t.Error(err)
		}
	}()
	evidence := attemptWitnessEvidence{}
	for _, kind := range []authority.Kind{authority.Mint, authority.Install} {
		fields := authority.IntentFields{Kind: kind, ArtifactDigest: strings.Repeat("a", 64)}
		if kind == authority.Mint {
			fields.ScopeURL = "https://example.com/scope/"
		}
		intent, err := authority.NewIntent(fields)
		if err != nil {
			t.Fatal(err)
		}
		transition, err := guard.Begin(ctx, intent, evidence)
		if err != nil {
			t.Fatal(kind, err)
		}
		pending, err := passiveAttemptLoad(t, dir, ctx, manager, binding)
		if err != nil || !pending.pending || pending.witnessInstallationKey != key || (pending.witness != nil) != (kind == authority.Install) {
			t.Fatal("pending mapping", pending, err)
		}
		for _, phase := range []authority.Phase{authority.LocalCommitted, authority.Published, authority.ConfigWritten} {
			if err := transition.SetPhase(ctx, phase, evidence, nil); err != nil {
				t.Fatal(err)
			}
		}
		if err := transition.Finalize(ctx, evidence, nil); err != nil {
			t.Fatal(err)
		}
		active, err := passiveAttemptLoad(t, dir, ctx, manager, binding)
		if err != nil || active.pending || active.witness == nil || active.witnessInstallationKey != key {
			t.Fatal("active mapping", active, err)
		}
		seq := uint64(3)
		if kind == authority.Install {
			seq = 4
		}
		if active.witness.Fields().LedgerSeq != seq {
			t.Fatal("cached witness", active.witness.Fields())
		}
	}
	for _, mode := range []string{"nil-manager", "nil-context", "bad-binding", "canceled"} {
		t.Run(mode, func(t *testing.T) {
			c := ctx
			m := manager
			b := binding
			switch mode {
			case "nil-manager":
				m = nil
			case "nil-context":
				c = nil
			case "bad-binding":
				b.database = ""
			case "canceled":
				var stop context.CancelFunc
				c, stop = context.WithCancel(ctx)
				stop()
			}
			got, err := loadReadAttemptFacts(c, m, b)
			if err == nil || got.witness != nil || got.installationKey != "" {
				t.Fatal("failed intake retained facts", got, err)
			}
		})
	}
	saved, err := os.ReadFile(filepath.Join(dir, beadsignore.WitnessName))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, beadsignore.WitnessName), []byte("{"), 0600); err != nil {
		t.Fatal(err)
	}
	got, err := loadReadAttemptFacts(ctx, manager, binding)
	if err == nil || got.witness != nil || got.installationKey != "" {
		t.Fatal("corrupt file admitted", got, err)
	}
	if err := os.WriteFile(filepath.Join(dir, beadsignore.WitnessName), saved, 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("BEADS_INSTALLATION_ID_FILE", filepath.Join(root, "other-installation-id"))
	other, err := authority.New(ctx, dir)
	if err != nil {
		t.Fatal(err)
	}
	got, err = loadReadAttemptFacts(ctx, other, binding)
	if !errors.Is(err, authority.ErrWrongInstallation) || got.witness != nil || got.installationKey != "" {
		t.Fatal("foreign manager admitted", got, err)
	}

}

func TestReadAttemptRealBodyBudgets(t *testing.T) {
	for _, method := range []string{"bead", "link", "incident"} {
		t.Run(method, func(t *testing.T) {
			ctx, want, r, m := attemptFixture(t)
			expectAttemptFacts(m, want, false)
			count := 5
			scope := want.witness.Fields().ScopeURL
			var body func(context.Context, queryer) (int, error)
			switch method {
			case "bead":
				row := validRow()
				expectBead(m, row)
				decl, err := graph.NewOwnedLinkDecl(relationType, "Explains", 2)
				if err != nil {
					t.Fatal(err)
				}
				expectDescriptor(m, memoryDescriptor(t, decl))
				expectLinks(m, "source_kind = 'in' AND source_path = ? AND type_url IN (?)", []driver.Value{row.path, relationType}, sqlmock.NewRows(allLinkColumns), 2)
				count = 7
				body = func(c context.Context, q queryer) (int, error) {
					record, err := readBeadInTx(c, q, scope, row.path, fixtureLimits)
					if err == nil && record.Bead.Path() != row.path {
						t.Fatal("wrong Bead")
					}
					return 1, err
				}
			case "link":
				row := validLinkRow()
				d := relationDescriptor(t)
				values := exactValues(row.path, graph.KindLink, append(linkValues(row), d.ID(), d.CanonicalJSON(), blobLength(d.CanonicalJSON()), d.Fingerprint()))
				m.ExpectQuery(regexp.QuoteMeta(exactLinkQuery)).WithArgs(fixtureLimits.valueBytes, fixtureLimits.valueBytes, row.path, row.path, row.path).WillReturnRows(sqlmock.NewRows(exactLinkColumns).AddRow(values...)).RowsWillBeClosed()
				body = func(c context.Context, q queryer) (int, error) {
					link, err := readLinkInTx(c, q, scope, row.path, fixtureLimits)
					if err == nil && link.Path() != row.path {
						t.Fatal("wrong Link")
					}
					return 1, err
				}
			case "incident":
				expectIncident(m, graph.DirectionBoth, fixtureLimits, incidentRows(validLinkRow()))
				body = func(c context.Context, q queryer) (int, error) {
					links, err := readIncidentLinksInTx(c, q, scope, "beads/plan", graph.DirectionBoth, fixtureLimits)
					if err == nil && len(links) != 1 {
						t.Fatal("wrong incident result")
					}
					return 1, err
				}
			}
			result, completion, err := runReadAttempt(ctx, ownReadAttempt(r), want, nil, body)
			if err != nil || result != 1 || !completion.cleanupOK || r.queries != count {
				t.Fatal(result, completion, err, r.queries, count)
			}
		})
	}
}

func TestReadAttemptBodyGoexitCleansAndJoins(t *testing.T) {
	ctx, want, r, m := attemptFixture(t)
	expectAttemptFacts(m, want, false)
	done := make(chan struct{})
	var recovered any
	returned := false
	go func() {
		defer close(done)
		defer func() { recovered = recover() }()
		_, _, _ = runReadAttempt(ctx, ownReadAttempt(r), want, nil, func(context.Context, queryer) (int, error) {
			r.events = append(r.events, "body")
			runtime.Goexit()
			return 0, nil
		})
		returned = true
	}()
	select {
	case <-done:
	case <-ctx.Done():
		t.Fatal("owned Goexit worker failed to join")
	}
	if recovered != nil || returned || !reflect.DeepEqual(r.events, []string{"query", "query", "query", "query", "body", "rollback", "connection"}) {
		t.Fatalf("Goexit changed: panic=%v returned=%v events=%v", recovered, returned, r.events)
	}
}

type attemptFileState struct {
	name     string
	mode     os.FileMode
	size     int64
	modified time.Time
	contents string
}

// This compares retained filesystem state, not transient syscalls or atime.
func attemptDirectoryState(t *testing.T, dir string) []attemptFileState {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	result := make([]attemptFileState, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			t.Fatal(err)
		}
		state := attemptFileState{name: entry.Name(), mode: info.Mode(), size: info.Size(), modified: info.ModTime()}
		if info.Mode().IsRegular() {
			raw, err := os.ReadFile(filepath.Join(dir, entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			state.contents = string(raw)
		}
		result = append(result, state)
	}
	return result
}
func passiveAttemptLoad(t *testing.T, dir string, ctx context.Context, manager *authority.Manager, binding readAttemptBinding) (expectedReadFacts, error) {
	t.Helper()
	before := attemptDirectoryState(t, dir)
	facts, err := loadReadAttemptFacts(ctx, manager, binding)
	if after := attemptDirectoryState(t, dir); !reflect.DeepEqual(before, after) {
		t.Fatal("passive Load changed retained directory state")
	}
	return facts, err
}
