//go:build cgo

package graphops

import (
	"context"
	"crypto/sha256"
	"database/sql"
	_ "embed"
	"errors"
	"os"
	"path/filepath"
	"runtime/debug"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/storage/embeddeddolt"
)

//go:embed testdata/state_fixture.sql
var stateFixtureSQL string

// This opt-in crosses the embedded driver's transaction/snapshot boundary.
// The fixture is not a minted Scope, lawful ledger, authority or managed-server
// qualification. It must run under the separately reviewed private environment.
func TestGraphStateEngineObservations(t *testing.T) {
	if os.Getenv("BEADS_TEST_GRAPH_STATE_FIXTURE") != "1" {
		t.Skip("opt-in embedded eight-table state observation fixture")
	}
	if value, present := os.LookupEnv("DOLT_METRICS_DISABLED"); !present || value != "1" {
		t.Fatal("launch process with DOLT_METRICS_DISABLED=1 before driver initialization")
	}
	if !filepath.IsAbs(os.Getenv("HOME")) || !filepath.IsAbs(os.Getenv("TMPDIR")) {
		t.Fatal("explicit absolute private HOME and TMPDIR required")
	}
	t.Logf("supervisor environment: HOME=%q TMPDIR=%q DOLT_METRICS_DISABLED=%q", os.Getenv("HOME"), os.Getenv("TMPDIR"), os.Getenv("DOLT_METRICS_DISABLED"))
	ctx, cancel := context.WithTimeout(t.Context(), 3*time.Minute)
	defer cancel()
	data := filepath.Join(t.TempDir(), "data")
	if err := os.Mkdir(data, 0o700); err != nil {
		t.Fatal(err)
	}
	t.Logf("private fixture data=%q", data)
	boot, closeBoot, err := embeddeddolt.OpenSQL(ctx, data, "", "")
	if err != nil {
		t.Fatal(err)
	}
	_, createErr := boot.ExecContext(ctx, "CREATE DATABASE graph_state_fixture")
	if err := errors.Join(createErr, closeBoot()); err != nil {
		t.Fatal(err)
	}
	db, closeEngine, err := embeddeddolt.OpenSQL(ctx, data, "graph_state_fixture", "main")
	if err != nil {
		t.Fatal(err)
	}
	closed := false
	defer func() {
		if !closed {
			if err := closeEngine(); err != nil {
				t.Error(err)
			}
		}
	}()
	if info, ok := debug.ReadBuildInfo(); ok {
		t.Logf("effective build info:\n%s", info.String())
	} else {
		t.Fatal("build info unavailable")
	}
	t.Logf("base schema SHA256=%x state schema SHA256=%x", sha256.Sum256([]byte(engineFixtureSQL)), sha256.Sum256([]byte(stateFixtureSQL)))
	seedEngineFixture(t, ctx, db, data)
	stateExec(t, ctx, db, stateFixtureSQL)
	baseline := stateRead(t, ctx, db)
	if baseline.database != "graph_state_fixture" || baseline.branch != "main" {
		t.Fatalf("unexpected binding: %+v", baseline)
	}
	t.Logf("initial eight-table observation: %+v", baseline)
	stateTableMutationControls(t, ctx, db, baseline)
	stateExclusionAndCommitControls(t, ctx, db, baseline)
	after := stateSnapshotControl(t, ctx, db)
	stateBranchAndMissingControl(t, ctx, db, after)
	closed = true
	if err := closeEngine(); err != nil {
		t.Fatal(err)
	}
	if t.Failed() {
		t.Fatal("state observation controls failed")
	}
	t.Log("GRAPH_STATE_FIXTURE_OK: embedded engine and connector closed")
}

type stateExecer interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func stateExec(t *testing.T, ctx context.Context, db stateExecer, query string, args ...any) {
	t.Helper()
	if _, err := db.ExecContext(ctx, query, args...); err != nil {
		t.Fatalf("fixture SQL %q: %v", query, err)
	}
}

func stateRead(t *testing.T, ctx context.Context, db *sql.DB) stateObservation {
	t.Helper()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	result, readErr := observeStateInTx(ctx, tx)
	if err := errors.Join(readErr, tx.Rollback()); err != nil {
		t.Fatal(err)
	}
	return result
}

func stateTableMutationControls(t *testing.T, ctx context.Context, db *sql.DB, baseline stateObservation) {
	t.Helper()
	// Deliberately one independent table change at a time, then checked rollback.
	changes := []string{
		"UPDATE graph_scope SET epoch = 2 WHERE id = 1",
		"INSERT INTO graph_scope_history VALUES ('https://old.example/scope/', 1, '2026-09-17 00:00:00', 'fixture')",
		"UPDATE graph_type_descriptors SET installed_seq = 2",
		"UPDATE graph_beads SET updated_at = '2026-09-17 00:00:01' WHERE path = 'beads/plan'",
		"UPDATE graph_links SET updated_at = '2026-09-17 00:00:01' WHERE path = 'links/plan-decision'",
		"UPDATE graph_ledger_seq SET next_seq = 2 WHERE id = 0",
		"INSERT INTO graph_ledger_events (seq, op_id, kind, scope_url, authority_id, epoch, at, prev_hash, hash) VALUES (1, REPEAT('a', 32), 'mint', 'https://graph.example/fixture/', REPEAT('b', 32), 1, '2026-09-17 00:00:00', REPEAT('0', 64), REPEAT('1', 64))",
		"INSERT INTO graph_allocations VALUES ('beads/plan', 'bead', 1, REPEAT('a', 32), 1, 'live', NULL, REPEAT('a', 32), 1)",
	}
	for slot, query := range changes {
		t.Run("table-"+stateColumns[slot+2], func(t *testing.T) {
			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback() }()
			stateExec(t, ctx, tx, query)
			got, err := observeStateInTx(ctx, tx)
			if err != nil {
				t.Fatal(err)
			}
			for i := range got.hashes {
				if (got.hashes[i] != baseline.hashes[i]) != (i == slot) {
					t.Fatalf("change slot %d affected slot %d: %+v", slot, i, got)
				}
			}
			if got.version == baseline.version {
				t.Fatal("table change did not change composite")
			}
			if err := tx.Rollback(); err != nil {
				t.Fatal(err)
			}
			if restored := stateRead(t, ctx, db); restored != baseline {
				t.Fatalf("rollback changed observation: %+v", restored)
			}
		})
	}
}

func stateExclusionAndCommitControls(t *testing.T, ctx context.Context, db *sql.DB, baseline stateObservation) {
	t.Helper()
	for _, query := range []string{
		"INSERT INTO issues VALUES ('issue-control', 'unrelated issue')",
		"UPDATE graph_authority_lease SET expires_at = '2026-09-17 02:00:00', heartbeat_at = '2026-09-17 00:30:00', fence = REPEAT('e', 32) WHERE id = 1",
	} {
		stateExec(t, ctx, db, query)
		if got := stateRead(t, ctx, db); got != baseline {
			t.Fatalf("excluded table activity changed graph observation: %+v", got)
		}
	}
	stateExec(t, ctx, db, "CALL DOLT_ADD('graph_scope', 'graph_scope_history', 'graph_type_descriptors', 'graph_beads', 'graph_links', 'graph_ledger_seq', 'graph_ledger_events', 'graph_allocations')")
	stateExec(t, ctx, db, "CALL DOLT_COMMIT('-m', 'state observation fixture', '--author', 'Fixture <fixture@example.invalid>')")
	if got := stateRead(t, ctx, db); got != baseline {
		t.Fatalf("Dolt commit changed content-addressed graph state: %+v", got)
	}
	t.Log("unrelated issue/lease activity excluded; DOLT_ADD and DOLT_COMMIT preserve all eight hashes")
}

func stateSnapshotControl(t *testing.T, ctx context.Context, db *sql.DB) stateObservation {
	t.Helper()
	reader, err := db.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := reader.Close(); err != nil {
			t.Error(err)
		}
	}()
	writer, err := db.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := writer.Close(); err != nil {
			t.Error(err)
		}
	}()
	read, err := reader.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = read.Rollback() }()
	before, err := observeStateInTx(ctx, read)
	if err != nil {
		t.Fatal(err)
	}
	bead, err := readBeadInTx(ctx, read, fixtureScope, "beads/plan", fixtureLimits)
	if err != nil {
		t.Fatal(err)
	}
	write, err := writer.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = write.Rollback() }()
	stateExec(t, ctx, write, "UPDATE graph_beads SET properties = ? WHERE path = 'beads/plan'", []byte(`{"text":"snapshot successor"}`))
	if err := write.Commit(); err != nil {
		t.Fatal(err)
	}
	during, err := observeStateInTx(ctx, read)
	if err != nil || during != before {
		t.Fatalf("hash snapshot changed after other session commit: %+v %v", during, err)
	}
	still, err := readBeadInTx(ctx, read, fixtureScope, "beads/plan", fixtureLimits)
	if err != nil || still.Bead.Properties().String() != bead.Bead.Properties().String() {
		t.Fatalf("row snapshot changed after other session commit: %v", err)
	}
	if err := read.Rollback(); err != nil {
		t.Fatal(err)
	}
	fresh, err := reader.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = fresh.Rollback() }()
	after, err := observeStateInTx(ctx, fresh)
	if err != nil || after.version == before.version || after.hashes[3] == before.hashes[3] {
		t.Fatalf("fresh hash snapshot did not advance: %+v %v", after, err)
	}
	for i := range after.hashes {
		if (after.hashes[i] != before.hashes[i]) != (i == 3) {
			t.Fatalf("snapshot Bead mutation affected hash slot %d", i)
		}
	}
	updated, err := readBeadInTx(ctx, fresh, fixtureScope, "beads/plan", fixtureLimits)
	if err != nil || updated.Bead.Properties().String() != `{"text":"snapshot successor"}` {
		t.Fatalf("fresh row snapshot did not advance: %v", err)
	}
	if err := fresh.Rollback(); err != nil {
		t.Fatal(err)
	}
	t.Log("two reserved SQL sessions: hashes and row body retain old snapshot, then advance together")
	return after
}

func stateBranchAndMissingControl(t *testing.T, ctx context.Context, db *sql.DB, main stateObservation) {
	t.Helper()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	closed := false
	defer func() {
		if !closed {
			if err := conn.Close(); err != nil {
				t.Error(err)
			}
		}
	}()
	// A second branch starts at the committed baseline, before snapshot mutation.
	stateExec(t, ctx, conn, "CALL DOLT_BRANCH('state_other')")
	stateExec(t, ctx, conn, "USE `graph_state_fixture/state_other`")
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	branch, readErr := observeStateInTx(ctx, tx)
	if err := errors.Join(readErr, tx.Rollback()); err != nil {
		t.Fatal(err)
	}
	if branch.branch != "state_other" || branch.database != "graph_state_fixture/state_other" {
		t.Fatalf("branch binding lost: %+v", branch)
	}
	stateExec(t, ctx, conn, "DROP TABLE graph_allocations")
	tx, err = conn.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	missing, readErr := observeStateInTx(ctx, tx)
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if readErr == nil || missing != (stateObservation{}) {
		t.Fatalf("missing table produced partial/default observation: %+v %v", missing, readErr)
	}
	t.Logf("second branch binding retained; missing table refused: %v", readErr)
	closed = true
	if err := conn.Close(); err != nil {
		t.Fatal(err)
	}
	if got := stateRead(t, ctx, db); got != main {
		t.Fatalf("branch-local DROP changed main: %+v", got)
	}
}
