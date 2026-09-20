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
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"
	"testing"
	"time"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage/embeddeddolt"
)

//go:embed testdata/precondition_fixture.sql
var preconditionFixtureSQL string

// Disposable projection only: not a lawful ledger, minted Scope, public Reader
// or authority claim. Managed SQL-server parity is a separate qualification.
func TestGraphPreconditionEngineObservations(t *testing.T) {
	if os.Getenv("BEADS_TEST_GRAPH_PRECONDITION_FIXTURE") != "1" {
		t.Skip("opt-in embedded raw precondition observation fixture")
	}
	if os.Getenv("DOLT_METRICS_DISABLED") != "1" {
		t.Fatal("DOLT_METRICS_DISABLED=1 required before driver initialization")
	}
	if !filepath.IsAbs(os.Getenv("HOME")) || !filepath.IsAbs(os.Getenv("TMPDIR")) {
		t.Fatal("absolute private HOME and TMPDIR required")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 3*time.Minute)
	defer cancel()
	data := filepath.Join(t.TempDir(), "data")
	if err := os.Mkdir(data, 0o700); err != nil {
		t.Fatal(err)
	}
	t.Logf("private HOME=%q TMPDIR=%q data=%q", os.Getenv("HOME"), os.Getenv("TMPDIR"), data)
	boot, closeBoot, err := embeddeddolt.OpenSQL(ctx, data, "", "")
	if err != nil {
		t.Fatal(err)
	}
	_, createErr := boot.ExecContext(ctx, "CREATE DATABASE graph_precondition_fixture")
	if err := errors.Join(createErr, closeBoot()); err != nil {
		t.Fatal(err)
	}
	db, closeEngine, err := embeddeddolt.OpenSQL(ctx, data, "graph_precondition_fixture", "main")
	if err != nil {
		t.Fatal(err)
	}
	closed := false
	t.Cleanup(func() {
		if !closed {
			if err := closeEngine(); err != nil {
				t.Error(err)
			}
		}
	})
	info, ok := debug.ReadBuildInfo()
	if !ok {
		t.Fatal("build info unavailable")
	}
	t.Logf("effective build info:\n%s", info.String())
	t.Logf("base schema SHA256=%x precondition schema SHA256=%x", sha256.Sum256([]byte(engineFixtureSQL)), sha256.Sum256([]byte(preconditionFixtureSQL)))
	seedEngineFixture(t, ctx, db, data)
	stateExec(t, ctx, db, preconditionFixtureSQL)
	reader := preconditionConnection(t, ctx, db)
	writer := preconditionConnection(t, ctx, db)
	// SET is connection setup, outside every observed transaction/body budget.
	stateExec(t, ctx, reader, "SET time_zone = '+00:00'")
	stateExec(t, ctx, writer, "SET time_zone = '+00:00'")
	baseline := readPreconditionFixture(t, ctx, reader, nil)
	if baseline.scope.epoch != 1 || baseline.scope.url != fixtureScope || baseline.ledger.tip.presence != observationAbsent || baseline.lease.presence != observationPresent || baseline.lease.zone != "+00:00" || baseline.state.branch != "main" {
		t.Fatalf("unexpected baseline: %+v", baseline)
	}
	wantLease := leaseObservation{
		presence: observationPresent, scopeURL: fixtureScope,
		authorityID: strings.Repeat("a", 32), holder: strings.Repeat("b", 64),
		renewer: strings.Repeat("c", 32), epoch: 1, fence: strings.Repeat("d", 32),
		grantedAt: "2026-09-17 00:00:00.000000", expiresAt: "2026-09-17 01:00:00.000000",
		heartbeatAt: "2026-09-17 00:00:00.000000", zone: "+00:00",
	}
	if preconditionStableFacts(baseline).lease != wantLease || baseline.lease.clock == "" {
		t.Fatalf("lease projection lost or reordered physical fields: %+v", baseline.lease)
	}
	repeated := readPreconditionFixture(t, ctx, reader, nil)
	if !reflect.DeepEqual(preconditionStableFacts(baseline), preconditionStableFacts(repeated)) {
		t.Fatal("read-only observation changed stable table or lease facts")
	}
	preconditionAbsenceAndUnsigned(t, ctx, reader, baseline)
	preconditionUTCBinding(t, ctx, reader)
	preconditionExplain(t, ctx, reader)
	preconditionSnapshot(t, ctx, reader, writer)
	readPreconditionComposition(t, ctx, reader, data)
	preconditionCorruptionAndMissing(t, ctx, reader)
	// Release pinned connections before checking engine closure.
	if err := errors.Join(reader.Close(), writer.Close()); err != nil {
		t.Fatal(err)
	}
	closed = true
	if err := closeEngine(); err != nil {
		t.Fatal(err)
	}
	t.Log("GRAPH_PRECONDITION_FIXTURE_OK: observations, UTC binding and two-session snapshot; connections and engine closed")
}

func preconditionConnection(t *testing.T, ctx context.Context, db *sql.DB) *sql.Conn {
	t.Helper()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := conn.Close(); err != nil && !errors.Is(err, sql.ErrConnDone) {
			t.Error(err)
		}
	})
	return conn
}
func preconditionTx(t *testing.T, ctx context.Context, conn *sql.Conn) *sql.Tx {
	t.Helper()
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			t.Error(err)
		}
	})
	return tx
}
func readPreconditionFixture(t *testing.T, ctx context.Context, conn *sql.Conn, seq *uint64) preconditionObservations {
	t.Helper()
	tx := preconditionTx(t, ctx, conn)
	got, readErr := observePreconditionsInTx(ctx, tx, seq)
	if err := errors.Join(readErr, tx.Rollback()); err != nil {
		t.Fatal(err)
	}
	return got
}
func preconditionStableFacts(value preconditionObservations) preconditionObservations {
	value.lease.clock = ""
	value.lease.queryStarted = time.Time{}
	value.lease.queryFinished = time.Time{}
	return value
}

func preconditionAbsenceAndUnsigned(t *testing.T, ctx context.Context, conn *sql.Conn, baseline preconditionObservations) {
	t.Helper()
	tx := preconditionTx(t, ctx, conn)
	stateExec(t, ctx, tx, "DELETE FROM graph_scope")
	stateExec(t, ctx, tx, "DELETE FROM graph_authority_lease")
	missing := uint64(3)
	absent, err := observePreconditionsInTx(ctx, tx, &missing)
	if err != nil || absent.scope.presence != observationAbsent || absent.lease.presence != observationAbsent || absent.lease.clock == "" || absent.lease.zone != "+00:00" || absent.ledger.recorded.presence != observationAbsent || absent.ledger.tip.presence != observationAbsent || absent.ledger.head == "" {
		t.Fatalf("absent observations: %+v %v", absent, err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if restored := readPreconditionFixture(t, ctx, conn, nil); !reflect.DeepEqual(preconditionStableFacts(restored), preconditionStableFacts(baseline)) {
		t.Fatal("absence rollback changed retained facts")
	}
	tx = preconditionTx(t, ctx, conn)
	for i, seq := range []uint64{1 << 63, graph.MaxLedgerSeq - 1, graph.MaxLedgerSeq} {
		preconditionInsertLedger(t, ctx, tx, seq, strings.Repeat(strconv.Itoa(i+1), 64))
	}
	for i, seq := range []uint64{1 << 63, graph.MaxLedgerSeq - 1, graph.MaxLedgerSeq} {
		got, err := observeLedgerInTx(ctx, tx, &seq)
		if err != nil || got.recorded.seq != seq || got.recorded.hash != strings.Repeat(strconv.Itoa(i+1), 64) || got.tip.seq != graph.MaxLedgerSeq {
			t.Fatalf("unsigned exact lookup: %+v %v", got, err)
		}
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	t.Log("Scope/lease absence retains clock/zone/HEAD; exact high-bit and adjacent MaxLedgerSeq lookups pass")
}
func preconditionInsertLedger(t *testing.T, ctx context.Context, tx stateExecer, seq uint64, hash string) {
	t.Helper()
	stateExec(t, ctx, tx, `INSERT INTO graph_ledger_events
 (seq, op_id, kind, scope_url, authority_id, authority_epoch, at, prev_hash, hash)
 VALUES (CAST(? AS UNSIGNED), REPEAT('a',32), 'mint', ?, REPEAT('a',32), 1, '2026-09-20 00:00:00', REPEAT('0',64), ?)`, strconv.FormatUint(seq, 10), fixtureScope, hash)
}

func preconditionUTCBinding(t *testing.T, ctx context.Context, conn *sql.Conn) {
	t.Helper()
	tx := preconditionTx(t, ctx, conn)
	input := time.Date(2026, 9, 20, 12, 34, 56, 123456000, time.FixedZone("fixture +05", 5*60*60))
	// The storage leg normalizes parameters: this driver formats civil fields
	// without preserving the Go value's offset or converting it to UTC.
	utc := input.UTC()
	stateExec(t, ctx, tx, "UPDATE graph_authority_lease SET granted_at = ?, heartbeat_at = ?, expires_at = ? WHERE id = 1", utc, utc, utc.Add(time.Hour))
	got, err := observeLeaseInTx(ctx, tx)
	if err != nil || got.grantedAt != "2026-09-20 07:34:56.123456" || got.heartbeatAt != got.grantedAt || got.expiresAt != "2026-09-20 08:34:56.123456" || got.zone != "+00:00" {
		t.Fatalf("explicit UTC-normalized Go parameter rendering: %+v %v", got, err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	// Characterize an unnormalized caller separately: even with a UTC session,
	// the driver binds the original civil fields. This is not a compliant write.
	tx = preconditionTx(t, ctx, conn)
	stateExec(t, ctx, tx, "UPDATE graph_authority_lease SET granted_at = ?, heartbeat_at = ?, expires_at = ? WHERE id = 1", input, input, input.Add(time.Hour))
	raw, err := observeLeaseInTx(ctx, tx)
	if err != nil || raw.grantedAt != "2026-09-20 12:34:56.123456" || raw.heartbeatAt != raw.grantedAt || raw.expiresAt != "2026-09-20 13:34:56.123456" || raw.zone != "+00:00" {
		t.Fatalf("raw non-UTC parameter civil-field characterization: %+v %v", raw, err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	// A non-UTC reader exposes its own zone, not the zone of a historical writer.
	stateExec(t, ctx, conn, "SET time_zone = '+05:00'")
	tx = preconditionTx(t, ctx, conn)
	other, err := observeLeaseInTx(ctx, tx)
	if err != nil || other.zone != "+05:00" || other.grantedAt != "2026-09-17 00:00:00.000000" {
		t.Fatalf("civil data rewritten under reader zone: %+v %v", other, err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	stateExec(t, ctx, conn, "SET time_zone = '+00:00'")
	t.Log("explicit UTC-normalized Go binding renders UTC; raw non-UTC binding retains civil fields despite UTC session; non-UTC reader retains zone without rewriting stored civil timestamp")
}

func preconditionSnapshot(t *testing.T, ctx context.Context, reader, writer *sql.Conn) {
	t.Helper()
	stateExec(t, ctx, writer, "CALL DOLT_ADD('graph_scope', 'graph_scope_history', 'graph_type_descriptors', 'graph_beads', 'graph_links', 'graph_ledger_seq', 'graph_ledger_events', 'graph_allocations')")
	stateExec(t, ctx, writer, "CALL DOLT_COMMIT('-m', 'precondition baseline', '--author', 'Fixture <fixture@example.invalid>')")
	read := preconditionTx(t, ctx, reader)
	seq := uint64(1)
	before, err := observePreconditionsInTx(ctx, read, &seq)
	if err != nil {
		t.Fatal(err)
	}
	bead, err := readBeadInTx(ctx, read, fixtureScope, "beads/plan", fixtureLimits)
	if err != nil {
		t.Fatal(err)
	}
	write := preconditionTx(t, ctx, writer)
	stateExec(t, ctx, write, "UPDATE graph_scope SET authority_epoch = 2 WHERE id = 1")
	stateExec(t, ctx, write, "UPDATE graph_authority_lease SET authority_epoch = 2, fence = REPEAT('e',32), expires_at = NOW(6) + INTERVAL 1 HOUR, heartbeat_at = NOW(6) WHERE id = 1")
	preconditionInsertLedger(t, ctx, write, seq, strings.Repeat("f", 64))
	stateExec(t, ctx, write, "UPDATE graph_beads SET properties = ? WHERE path = 'beads/plan'", []byte(`{"text":"precondition successor"}`))
	if err := write.Commit(); err != nil {
		t.Fatal(err)
	}
	// A real Dolt commit advances HEAD while A still holds its read transaction.
	stateExec(t, ctx, writer, "CALL DOLT_ADD('graph_scope', 'graph_ledger_events', 'graph_beads')")
	stateExec(t, ctx, writer, "CALL DOLT_COMMIT('-m', 'precondition successor', '--author', 'Fixture <fixture@example.invalid>')")
	during, err := observePreconditionsInTx(ctx, read, &seq)
	if err != nil || !reflect.DeepEqual(preconditionStableFacts(before), preconditionStableFacts(during)) {
		t.Fatalf("snapshot facts changed after B commit: before=%+v after=%+v err=%v", before, during, err)
	}
	if during.lease.clock <= before.lease.clock {
		t.Fatalf("NOW6 did not advance across actual intervening work: %q <= %q", during.lease.clock, before.lease.clock)
	}
	still, err := readBeadInTx(ctx, read, fixtureScope, "beads/plan", fixtureLimits)
	if err != nil || still.Bead.Properties().String() != bead.Bead.Properties().String() {
		t.Fatalf("row snapshot changed: %v", err)
	}
	if err := read.Rollback(); err != nil {
		t.Fatal(err)
	}
	fresh := preconditionTx(t, ctx, reader)
	after, err := observePreconditionsInTx(ctx, fresh, &seq)
	if err != nil || after.scope.epoch != 2 || after.ledger.recorded.seq != 1 || after.ledger.tip.seq != 1 || after.lease.epoch != 2 || after.lease.fence != strings.Repeat("e", 32) || after.ledger.head == before.ledger.head || after.state.version == before.state.version {
		t.Fatalf("fresh facts did not advance: %+v %v", after, err)
	}
	updated, err := readBeadInTx(ctx, fresh, fixtureScope, "beads/plan", fixtureLimits)
	if err != nil || updated.Bead.Properties().String() != `{"text":"precondition successor"}` {
		t.Fatalf("fresh row did not advance: %v", err)
	}
	repeated, err := observePreconditionsInTx(ctx, fresh, &seq)
	if err != nil || !reflect.DeepEqual(preconditionStableFacts(after), preconditionStableFacts(repeated)) {
		t.Fatalf("fresh observation changed data: %v", err)
	}
	if err := fresh.Rollback(); err != nil {
		t.Fatal(err)
	}
	t.Log("two-session Scope/ledger/ignored-lease/HEAD/hash/body snapshot; live NOW6 advances; fresh transaction observes B commits")
}

func preconditionExplain(t *testing.T, ctx context.Context, conn *sql.Conn) {
	t.Helper()
	rows, err := conn.QueryContext(ctx, "EXPLAIN FORMAT=tree "+ledgerObservationQuery, "3")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			t.Error(err)
		}
	}()
	columns, err := rows.Columns()
	if err != nil {
		t.Fatal(err)
	}
	if len(columns) != 1 || columns[0] != "plan" {
		t.Fatalf("unexpected EXPLAIN tree columns: %v", columns)
	}
	count := 0
	for rows.Next() {
		count++
		if count > 64 {
			t.Fatal("EXPLAIN row bound")
		}
		values := make([]any, len(columns))
		targets := make([]any, len(columns))
		for i := range values {
			targets[i] = &values[i]
		}
		if err := rows.Scan(targets...); err != nil {
			t.Fatal(err)
		}
		for i, value := range values {
			text, err := observationText(value, 4096)
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("ledger EXPLAIN %s: %s", columns[i], text)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if count == 0 {
		t.Fatal("missing EXPLAIN plan")
	}
	// A reviewer must inspect this actual plan before claiming indexed access.
}

func preconditionCorruptionAndMissing(t *testing.T, ctx context.Context, conn *sql.Conn) {
	t.Helper()
	tx := preconditionTx(t, ctx, conn)
	stateExec(t, ctx, tx, "UPDATE graph_ledger_events SET hash = REPEAT('z',64) WHERE seq = 1")
	seq := uint64(1)
	got, err := observePreconditionsInTx(ctx, tx, &seq)
	if !errors.Is(err, errCorrupt) || !reflect.DeepEqual(got, preconditionObservations{}) {
		t.Fatalf("malformed actual ledger produced facts: %+v %v", got, err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	// Disposable schema negative control, after all successful snapshot checks.
	stateExec(t, ctx, conn, "DROP TABLE graph_authority_lease")
	tx = preconditionTx(t, ctx, conn)
	got, err = observePreconditionsInTx(ctx, tx, &seq)
	if err == nil || !reflect.DeepEqual(got, preconditionObservations{}) {
		t.Fatalf("missing lease table became absent: %+v %v", got, err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	t.Log("actual malformed ledger and missing lease table refuse without partial facts")
}
