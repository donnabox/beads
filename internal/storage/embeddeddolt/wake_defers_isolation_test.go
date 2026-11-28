//go:build cgo

package embeddeddolt_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/storage/embeddeddolt"
	"github.com/steveyegge/beads/internal/storage/issueops"
	"github.com/steveyegge/beads/internal/types"
)

// The expired-defer sweep returns every dated defer whose date has passed to open, in one
// transaction. With versioned history on, each wake mints a version of the issue, and the mint
// refuses what it cannot record faithfully: a number outside the I-JSON exact-integer range in
// the issue's metadata or in a gate's timeout. A row that holds such a value reaches a store with
// history on off the usual path (written while history was off, by `bd sql`, or pulled from a
// clone that had history off), and before the sweep learned to skip it, the refusal aborted the
// whole transaction: nothing woke, on every read, for as long as that one row stayed deferred.
//
// These tests pin the replacement behavior at the layer where the transaction is real. The sweep
// decides per row, before it writes anything, whether the mint would refuse the row, and skips
// and reports such a row instead of waking it. The decision has to be the mint's own, which is
// why the first test compares the two on every kind of row instead of asserting what the sweep
// is thought to do.

// wakeRun numbers each environment. The sweep warns once per process per issue id, so an id that
// repeats across runs of one test binary would silence the warning a later run asserts on.
var wakeRun atomic.Int64

// wakeEnv is a store plus the run number that keeps its issue ids unique within the process.
type wakeEnv struct {
	*testEnv
	run int64
}

func newWakeEnv(t *testing.T) *wakeEnv {
	t.Helper()
	te := newTestEnv(t, "wk")
	configurer, ok := any(te.store).(storage.VersionedHistoryConfigurer)
	if !ok {
		t.Fatalf("%T does not implement storage.VersionedHistoryConfigurer", te.store)
	}
	// Every fixture is written with history off: that is how a row comes to hold a value the
	// mint refuses, and it keeps the fixtures' own version rows out of the counts below.
	configurer.SetVersionedHistoryEnabled(false)
	return &wakeEnv{testEnv: te, run: wakeRun.Add(1)}
}

// id names an issue for this run. The tags used for the batch sort in the order the sweep reads
// its rows, so the first tag is the lowest id.
func (e *wakeEnv) id(tag string) string { return fmt.Sprintf("wk-r%d-%s", e.run, tag) }

// wakeKind is the kind of row an expired defer sits in, by what decides whether the mint versions
// it.
type wakeKind int

const (
	// kindDurable is an issues-table row that participates in history: participation_generation
	// is set, so the mint runs its admission gate over it.
	kindDurable wakeKind = iota
	// kindLegacy is an issues-table row that never participated: participation_generation is
	// NULL, and the mint's write fence returns before the admission gate, so it is skipped whole.
	kindLegacy
	// kindNoHistory is a promoted no-history row: a durable issues-table row with NoHistory set.
	// It carries a participation_generation on purpose, so that only the rule that no-history
	// rows are never versioned, and not the fence, keeps the mint from refusing it.
	kindNoHistory
	// kindWisp is a row in the wisps table.
	kindWisp
)

func (k wakeKind) String() string {
	switch k {
	case kindDurable:
		return "durable"
	case kindLegacy:
		return "legacy"
	case kindNoHistory:
		return "no-history"
	case kindWisp:
		return "wisp"
	}
	return fmt.Sprintf("kind(%d)", int(k))
}

const (
	// A metadata number past 2^53-1.
	poisonedMetadata = `{"n":1727000000000000000}`
	cleanMetadata    = `{"n":1}`
	// A gate timeout past 2^53-1 nanoseconds (2600 hours is 9.36e15).
	poisonedTimeoutNs = int64(2600 * time.Hour)
	cleanTimeoutNs    = int64(time.Hour)
)

// wakeSeed is one expired dated defer to write.
type wakeSeed struct {
	id        string
	kind      wakeKind
	metadata  string
	timeoutNs int64
	// expiredFor is how long ago the defer date passed. Rows are given distinct dates so that the
	// primary-key order and the defer_until order agree, but the sweep's unordered SELECT is served
	// in neither: the engine returns rows by updated_at, at one-second granularity, and then by id,
	// and seeding a row refreshes updated_at. A test that depends on the order rows are read in
	// therefore seeds them in that order.
	expiredFor time.Duration
}

func (s wakeSeed) table() string {
	if s.kind == kindWisp {
		return "wisps"
	}
	return "issues"
}

// seed writes s as an expired dated defer, with versioned history off. The row is created through
// the store and then set to the exact shape under test by SQL, which is the path a row takes to
// hold a value the mint would refuse.
func (e *wakeEnv) seed(t *testing.T, s wakeSeed) {
	t.Helper()
	ctx := t.Context()
	issue := &types.Issue{
		ID: s.id, Title: "expired defer " + s.id, IssueType: types.TypeTask, Status: types.StatusOpen,
		Ephemeral: s.kind == kindWisp,
	}
	if err := e.store.CreateIssue(ctx, issue, "actor"); err != nil {
		t.Fatalf("creating %s: %v", s.id, err)
	}
	var participation any
	noHistory := 0
	switch s.kind {
	case kindDurable:
		participation = 1
	case kindNoHistory:
		participation = 1
		noHistory = 1
	}
	e.exec(t, ctx, fmt.Sprintf(`UPDATE %s SET status = 'deferred', defer_until = ?, metadata = ?, timeout_ns = ?,
		participation_generation = ?, no_history = ? WHERE id = ?`, s.table()),
		time.Now().UTC().Add(-s.expiredFor), s.metadata, s.timeoutNs, participation, noHistory, s.id)
}

// forget removes a fixture so the next check sweeps only its own rows.
func (e *wakeEnv) forget(t *testing.T, ids ...string) {
	t.Helper()
	for _, id := range ids {
		e.exec(t, t.Context(), `DELETE FROM issues WHERE id = ?`, id)
		e.exec(t, t.Context(), `DELETE FROM wisps WHERE id = ?`, id)
	}
}

// inTx runs fn in one transaction on a fresh SQL connection with versioned history scoped to
// historyOn, the way a store scopes the transaction it hands the sweep, and rolls the transaction
// back afterwards, so nothing a check does to the database outlives the check.
func (e *wakeEnv) inTx(t *testing.T, historyOn bool, fn func(ctx context.Context, tx *sql.Tx)) {
	t.Helper()
	ctx := t.Context()
	db, cleanup, err := embeddeddolt.OpenSQL(ctx, e.dataDir, e.database, "main")
	if err != nil {
		t.Fatalf("OpenSQL: %v", err)
	}
	defer func() { _ = cleanup() }()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}
	defer func() { _ = tx.Rollback() }()
	defer issueops.ScopeVersionedHistoryTransaction(tx, historyOn)()
	fn(ctx, tx)
}

func scalar(t *testing.T, ctx context.Context, tx *sql.Tx, query string, args []any, dest ...any) {
	t.Helper()
	if err := tx.QueryRowContext(ctx, query, args...).Scan(dest...); err != nil {
		t.Fatalf("%q: %v", query, err)
	}
}

// versionRows counts the version rows minted for id.
func versionRows(t *testing.T, ctx context.Context, tx *sql.Tx, id string) int {
	t.Helper()
	var n int
	scalar(t, ctx, tx, `SELECT COUNT(*) FROM issue_versions WHERE issue_id = ?`, []any{id}, &n)
	return n
}

// wakeEvents counts the status-change events the sweep recorded for id.
func wakeEvents(t *testing.T, ctx context.Context, tx *sql.Tx, kind wakeKind, id string) int {
	t.Helper()
	table := "events"
	if kind == kindWisp {
		table = "wisp_events"
	}
	var n int
	scalar(t, ctx, tx, fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE issue_id = ? AND event_type = ? AND actor = ?`, table),
		[]any{id, string(types.EventStatusChanged), issueops.DeferWakeActor}, &n)
	return n
}

// deferState reads whether id is still a deferred row with a date, and its status.
func deferState(t *testing.T, ctx context.Context, tx *sql.Tx, kind wakeKind, id string) (status string, dated bool) {
	t.Helper()
	table := "issues"
	if kind == kindWisp {
		table = "wisps"
	}
	var until sql.NullTime
	scalar(t, ctx, tx, fmt.Sprintf(`SELECT status, defer_until FROM %s WHERE id = ?`, table), []any{id}, &status, &until)
	return status, until.Valid
}

// linesNaming returns the lines of out that mention id.
func linesNaming(out, id string) []string {
	var lines []string
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, id) {
			lines = append(lines, line)
		}
	}
	return lines
}

// sweepVerdict says what one sweep did with the single row id: woke it, skipped it, or failed.
func sweepVerdict(res issueops.WakeDefersResult, err error, id string) string {
	switch {
	case err != nil:
		return "the sweep failed: " + err.Error()
	case slices.ContainsFunc(res.Skipped, func(u issueops.UnversionableIssue) bool { return u.ID == id }):
		return "skipped"
	case slices.Contains(res.Issues, id) || slices.Contains(res.Wisps, id):
		return "woken"
	}
	return "neither woken nor skipped"
}

// TestWakeSweepDecidesAsTheMintDoesOnEveryKindOfRow compares the sweep's verdict on a row with the
// verdict the mint gives the same row, rather than with an expectation written by hand. A
// pre-flight that disagrees with the mint is worse than none: it skips a row the mint would have
// accepted, so the row never wakes, or it wakes a row the mint then refuses, which aborts the
// batch again. The mint's verdict is read by calling it on the row in a transaction that rolls
// back. The expected cells are stated as well, so the two cannot agree on the wrong answer: only a
// durable row that participates in history and holds the value is refused, for both fields. A
// legacy row is skipped by the mint's write fence whatever it holds, a no-history row and a wisp
// are never versioned, and every clean twin is accepted.
func TestWakeSweepDecidesAsTheMintDoesOnEveryKindOfRow(t *testing.T) {
	skipUnlessEmbeddedDolt(t)
	e := newWakeEnv(t)

	fields := []struct {
		name         string
		held, clean  wakeSeed
		refusedField string
	}{
		{"metadata number", wakeSeed{metadata: poisonedMetadata, timeoutNs: cleanTimeoutNs}, wakeSeed{metadata: cleanMetadata, timeoutNs: cleanTimeoutNs}, "metadata"},
		{"gate timeout", wakeSeed{metadata: cleanMetadata, timeoutNs: poisonedTimeoutNs}, wakeSeed{metadata: cleanMetadata, timeoutNs: cleanTimeoutNs}, "timeout"},
	}
	for _, kind := range []wakeKind{kindDurable, kindLegacy, kindNoHistory, kindWisp} {
		for _, field := range fields {
			for _, holds := range []bool{true, false} {
				name := fmt.Sprintf("%s row, %s, clean twin", kind, field.name)
				seed := field.clean
				if holds {
					name = fmt.Sprintf("%s row, %s, holds the value", kind, field.name)
					seed = field.held
				}
				t.Run(name, func(t *testing.T) {
					seed.id, seed.kind, seed.expiredFor = e.id("cell"), kind, time.Hour
					e.seed(t, seed)
					defer e.forget(t, seed.id)

					wantRefused := holds && kind == kindDurable

					var mintErr error
					e.inTx(t, true, func(ctx context.Context, tx *sql.Tx) {
						mintErr = issueops.RecordVersionInTx(ctx, tx, seed.id, "actor")
					})
					if (mintErr != nil) != wantRefused {
						t.Fatalf("the mint's verdict on this row = %v, want refused: %v; the fixture is not the cell it names", mintErr, wantRefused)
					}

					var res issueops.WakeDefersResult
					var sweepErr error
					e.inTx(t, true, func(ctx context.Context, tx *sql.Tx) {
						res, sweepErr = issueops.WakeExpiredDefersInTx(ctx, tx)
					})
					got := sweepVerdict(res, sweepErr, seed.id)
					wantVerdict := "woken"
					if mintErr != nil {
						wantVerdict = "skipped"
					}
					if got != wantVerdict {
						t.Fatalf("the sweep's verdict = %q, the mint's verdict was %v, so the sweep should have %q", got, mintErr, wantVerdict)
					}
					if wantRefused {
						if len(res.Skipped) != 1 || !errors.Is(res.Skipped[0].Err, issueops.ErrIntegerNotRepresentable) || res.Skipped[0].Field != field.refusedField {
							t.Fatalf("skipped = %+v, want the one row, refused with ErrIntegerNotRepresentable in field %q", res.Skipped, field.refusedField)
						}
					} else if len(res.Skipped) != 0 {
						t.Fatalf("skipped = %+v, want none: the mint accepts this row", res.Skipped)
					}
				})
			}
		}
	}
}

// TestWakeSweepSkipsTheRefusedRowAndWakesTheRest is the batch the sweep used to abort on. The
// poisoned row has the lowest id and the earliest date, so a sweep that stops at the first
// refusal stops before it reaches any other row. Beside it sit rows that must still wake: two
// clean durable rows, a promoted no-history row and a wisp that hold the same value (which the
// mint accepts as a no-op), and a legacy row that holds it (which the mint's write fence skips).
func TestWakeSweepSkipsTheRefusedRowAndWakesTheRest(t *testing.T) {
	skipUnlessEmbeddedDolt(t)
	e := newWakeEnv(t)

	seeds := map[string]wakeSeed{
		"poisoned": {id: e.id("a-poisoned"), kind: kindDurable, metadata: poisonedMetadata, timeoutNs: cleanTimeoutNs, expiredFor: 6 * time.Hour},
		"clean1":   {id: e.id("b-clean1"), kind: kindDurable, metadata: cleanMetadata, timeoutNs: cleanTimeoutNs, expiredFor: 5 * time.Hour},
		"clean2":   {id: e.id("c-clean2"), kind: kindDurable, metadata: cleanMetadata, timeoutNs: cleanTimeoutNs, expiredFor: 4 * time.Hour},
		"nohist":   {id: e.id("d-nohistory"), kind: kindNoHistory, metadata: poisonedMetadata, timeoutNs: cleanTimeoutNs, expiredFor: 3 * time.Hour},
		"wisp":     {id: e.id("e-wisp"), kind: kindWisp, metadata: poisonedMetadata, timeoutNs: cleanTimeoutNs, expiredFor: 2 * time.Hour},
		"legacy":   {id: e.id("f-legacy"), kind: kindLegacy, metadata: poisonedMetadata, timeoutNs: cleanTimeoutNs, expiredFor: time.Hour},
	}
	// Seed in a fixed order with the poisoned row first, so that the sweep reads it first (see
	// wakeSeed.expiredFor). Ranging over the map would seed in a random order.
	for _, name := range []string{"poisoned", "clean1", "clean2", "nohist", "wisp", "legacy"} {
		e.seed(t, seeds[name])
	}
	poisoned := seeds["poisoned"].id

	e.inTx(t, true, func(ctx context.Context, tx *sql.Tx) {
		// The sweep's SELECT has no ORDER BY. Read it the way the sweep does and require the
		// poisoned row first, so that a failure below is about the batch and not about the order.
		rows, err := tx.QueryContext(ctx, `SELECT id FROM issues WHERE status = 'deferred' AND defer_until IS NOT NULL AND defer_until <= UTC_TIMESTAMP()`)
		if err != nil {
			t.Fatalf("reading the order the sweep reads its rows in: %v", err)
		}
		var order []string
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				t.Fatalf("scanning: %v", err)
			}
			order = append(order, id)
		}
		_ = rows.Close()
		if len(order) == 0 || order[0] != poisoned {
			t.Fatalf("the sweep reads %v first, want the poisoned row %s; the batch below would not exercise an abort at the first refusal", order, poisoned)
		}

		before := map[string]int{}
		for name, s := range seeds {
			before[name] = versionRows(t, ctx, tx, s.id)
		}

		res, err := issueops.WakeExpiredDefersInTx(ctx, tx)
		if err != nil {
			t.Fatalf("the sweep returned %v; one row the mint refuses must not stop the rest of the batch", err)
		}

		if got, want := sortedIDs(res.Issues), sortedIDs([]string{seeds["clean1"].id, seeds["clean2"].id, seeds["nohist"].id, seeds["legacy"].id}); !slices.Equal(got, want) {
			t.Errorf("woken issues = %v, want %v", got, want)
		}
		if got, want := res.Wisps, []string{seeds["wisp"].id}; !slices.Equal(got, want) {
			t.Errorf("woken wisps = %v, want %v", got, want)
		}
		if len(res.Skipped) != 1 || res.Skipped[0].ID != poisoned {
			t.Fatalf("skipped = %+v, want exactly the poisoned row %s", res.Skipped, poisoned)
		}
		if skip := res.Skipped[0]; !errors.Is(skip.Err, issueops.ErrIntegerNotRepresentable) || skip.Field != "metadata" {
			t.Errorf("the skip names field %q with error %v, want field %q and ErrIntegerNotRepresentable", skip.Field, skip.Err, "metadata")
		}

		for name, s := range seeds {
			status, dated := deferState(t, ctx, tx, s.kind, s.id)
			woke := wakeEvents(t, ctx, tx, s.kind, s.id)
			minted := versionRows(t, ctx, tx, s.id) - before[name]
			if name == "poisoned" {
				if status != "deferred" || !dated || woke != 0 || minted != 0 {
					t.Errorf("poisoned row: status %q, dated %v, wake events %d, version rows minted %d; want it left deferred and dated with no wake event and no version row", status, dated, woke, minted)
				}
				continue
			}
			wantMinted := 0
			if name == "clean1" || name == "clean2" {
				wantMinted = 1
			}
			if status != "open" || dated || woke != 1 || minted != wantMinted {
				t.Errorf("%s row: status %q, dated %v, wake events %d, version rows minted %d; want open, undated, one wake event and %d version rows", name, status, dated, woke, minted, wantMinted)
			}
		}
	})
}

// TestWakeSweepWritesNothingWhenEveryExpiredRowIsSkipped covers the batch with only the poisoned
// row: the sweep reports it and returns a nil error with nothing woken, and the row, its events
// and its versions are exactly as they were, so a caller that commits only when something woke has
// nothing to commit.
func TestWakeSweepWritesNothingWhenEveryExpiredRowIsSkipped(t *testing.T) {
	skipUnlessEmbeddedDolt(t)
	e := newWakeEnv(t)
	p := wakeSeed{id: e.id("only-poisoned"), kind: kindDurable, metadata: poisonedMetadata, timeoutNs: cleanTimeoutNs, expiredFor: time.Hour}
	e.seed(t, p)

	e.inTx(t, true, func(ctx context.Context, tx *sql.Tx) {
		snapshot := func() string {
			var status string
			var until, updated sql.NullTime
			var lock sql.NullInt64
			scalar(t, ctx, tx, `SELECT status, defer_until, updated_at, row_lock FROM issues WHERE id = ?`, []any{p.id}, &status, &until, &updated, &lock)
			return fmt.Sprintf("status=%s defer_until=%v updated_at=%v row_lock=%v events=%d versions=%d",
				status, until, updated, lock, wakeEvents(t, ctx, tx, p.kind, p.id), versionRows(t, ctx, tx, p.id))
		}
		before := snapshot()

		res, err := issueops.WakeExpiredDefersInTx(ctx, tx)
		if err != nil {
			t.Fatalf("the sweep returned %v, want a nil error: the only expired row is one the mint refuses", err)
		}
		if len(res.Issues) != 0 || len(res.Wisps) != 0 {
			t.Errorf("woken = issues %v, wisps %v, want nothing", res.Issues, res.Wisps)
		}
		if len(res.Skipped) != 1 || res.Skipped[0].ID != p.id {
			t.Errorf("skipped = %+v, want exactly %s", res.Skipped, p.id)
		}
		if after := snapshot(); after != before {
			t.Errorf("the sweep wrote to a row it skipped:\n  before: %s\n  after:  %s", before, after)
		}
	})
}

// TestWakeSweepWithHistoryOffWakesTheRowItWouldOtherwiseSkip pins the other side: the mint is a
// no-op with history off, so nothing can refuse a row, and the sweep wakes it, with no version row
// and nothing skipped. (The statement-level half, that it loads nothing to find that out, is in
// the issueops package tests, where the statements can be scripted.)
func TestWakeSweepWithHistoryOffWakesTheRowItWouldOtherwiseSkip(t *testing.T) {
	skipUnlessEmbeddedDolt(t)
	e := newWakeEnv(t)
	p := wakeSeed{id: e.id("history-off"), kind: kindDurable, metadata: poisonedMetadata, timeoutNs: cleanTimeoutNs, expiredFor: time.Hour}
	e.seed(t, p)

	e.inTx(t, false, func(ctx context.Context, tx *sql.Tx) {
		res, err := issueops.WakeExpiredDefersInTx(ctx, tx)
		if err != nil {
			t.Fatalf("the sweep with history off returned %v", err)
		}
		if !slices.Equal(res.Issues, []string{p.id}) || len(res.Skipped) != 0 {
			t.Errorf("woken issues = %v, skipped = %+v, want %s woken and nothing skipped", res.Issues, res.Skipped, p.id)
		}
		if status, dated := deferState(t, ctx, tx, p.kind, p.id); status != "open" || dated {
			t.Errorf("status %q, dated %v, want open and undated", status, dated)
		}
		if n := versionRows(t, ctx, tx, p.id); n != 0 {
			t.Errorf("%d version rows, want none: history is off", n)
		}
	})
}

// TestWakeSweepNamesASkippedRowOncePerProcess covers the cost of a skip: the row stays deferred
// and expired, so every sweep meets it again and reports it again in the result, but the warning
// that names it is printed once per process, not on every ready-work read.
func TestWakeSweepNamesASkippedRowOncePerProcess(t *testing.T) {
	skipUnlessEmbeddedDolt(t)
	e := newWakeEnv(t)
	p := wakeSeed{id: e.id("warned"), kind: kindDurable, metadata: poisonedMetadata, timeoutNs: cleanTimeoutNs, expiredFor: time.Hour}
	e.seed(t, p)

	sweep := func() (issueops.WakeDefersResult, string) {
		var res issueops.WakeDefersResult
		var err error
		out := captureStderr(t, func() {
			e.inTx(t, true, func(ctx context.Context, tx *sql.Tx) {
				res, err = issueops.WakeExpiredDefersInTx(ctx, tx)
			})
		})
		if err != nil {
			t.Fatalf("the sweep returned %v", err)
		}
		return res, out
	}

	res, out := sweep()
	if len(res.Skipped) != 1 || res.Skipped[0].ID != p.id {
		t.Fatalf("first sweep skipped = %+v, want exactly %s", res.Skipped, p.id)
	}
	lines := linesNaming(out, p.id)
	if len(lines) != 1 {
		t.Fatalf("first sweep printed %d lines naming %s, want one:\n%s", len(lines), p.id, out)
	}
	for _, want := range []string{"metadata", issueops.ErrIntegerNotRepresentable.Error()} {
		if !strings.Contains(lines[0], want) {
			t.Errorf("the warning %q does not name %q: it should name the issue, the field and the refusal", lines[0], want)
		}
	}

	res, out = sweep()
	if len(res.Skipped) != 1 || res.Skipped[0].ID != p.id {
		t.Errorf("second sweep skipped = %+v, want %s again: the row is still deferred and expired", res.Skipped, p.id)
	}
	if again := linesNaming(out, p.id); len(again) != 0 {
		t.Errorf("second sweep printed a second warning:\n%s", strings.Join(again, "\n"))
	}
}
