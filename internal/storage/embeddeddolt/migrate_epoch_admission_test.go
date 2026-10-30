//go:build cgo

package embeddeddolt_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/storage/embeddeddolt"
	storeops "github.com/steveyegge/beads/internal/storage/issueops"
	"github.com/steveyegge/beads/internal/storage/schema"
)

// These are new disposable embedded fixtures. They exercise the ordinary
// production migration entry point, not a replay of the fresh CLI bundle.
func TestEmbeddedPreEpochSelectedLineageMigrates(t *testing.T) {
	requireEmbedded(t)
	ctx := t.Context()
	dir := seedMainSchemaAt(t, ctx, 70)
	conn, closeConn := openPinnedConn(t, ctx, dir)
	defer closeConn()
	if _, err := schema.MigrateUp(ctx, conn); err != nil {
		t.Fatalf("pre-epoch selected upgrade: %v", err)
	}
	if got := currentMainVersion(t, ctx, conn); got != 74 {
		t.Fatalf("cursor%d, want74", got)
	}
	// A second ordinary pass must accept the completed actual engine metadata
	// through the same enforced-shape guard and perform no numbered migration.
	if applied, err := schema.MigrateUp(ctx, conn); err != nil || applied != 0 {
		t.Fatalf("completed epoch reopen applied%d: %v", applied, err)
	}
}

func TestEmbeddedEpochEmptyOldLineageUpgradesWithoutReinterpretation(t *testing.T) {
	requireEmbedded(t)
	ctx := t.Context()
	dir := seedMainSchemaAt(t, ctx, 72)
	conn, closeConn := openPinnedConn(t, ctx, dir)
	defer closeConn()

	before := epochFixtureState(t, ctx, conn)
	if _, err := schema.MigrateUp(ctx, conn); err != nil {
		t.Fatalf("released epoch upgrade: %v", err)
	}
	if got := currentMainVersion(t, ctx, conn); got != 74 {
		t.Fatalf("cursor%d,want74", got)
	}
	if after := epochFixtureState(t, ctx, conn); after != before {
		t.Fatal("upgrade reinterpreted or rewrote released epoch rows/DDL")
	}

	if applied, err := schema.MigrateUp(ctx, conn); err != nil || applied != 0 {
		t.Fatalf("ordinary reopen applied%d: %v", applied, err)
	}
}

type epochRefusalConn struct {
	*sql.Conn
	writes int
}

func (d *epochRefusalConn) ExecContext(context.Context, string, ...any) (sql.Result, error) {
	d.writes++
	return nil, errors.New("unexpected migration write after epoch refusal")
}

func (d *epochRefusalConn) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	trimmed := strings.ToUpper(strings.TrimSpace(query))
	if !strings.HasPrefix(trimmed, "SELECT ") && !strings.HasPrefix(trimmed, "SHOW ") {
		d.writes++
		return nil, errors.New("unexpected procedure/write query after epoch refusal")
	}
	return d.Conn.QueryContext(ctx, query, args...)
}

// Capture the complete epoch fixture data and DDL, with SQL NULL distinct
// from an empty string. This is logical fixture evidence, not a file-hash claim.
func epochFixtureState(t *testing.T, ctx context.Context, conn *sql.Conn) string {
	t.Helper()
	var table, ddl string
	if err := conn.QueryRowContext(ctx, "SHOW CREATE TABLE epoch_minted_addresses").Scan(&table, &ddl); err != nil {
		t.Fatal(err)
	}
	rows, err := conn.QueryContext(ctx, "SELECT * FROM epoch_minted_addresses ORDER BY address")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		t.Fatal(err)
	}
	var data [][]sql.NullString
	for rows.Next() {
		cells := make([]sql.NullString, len(columns))
		dest := make([]any, len(columns))
		for i := range cells {
			dest[i] = &cells[i]
		}
		if err := rows.Scan(dest...); err != nil {
			t.Fatal(err)
		}
		data = append(data, cells)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if err := rows.Close(); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(struct {
		Table, DDL string
		Columns    []string
		Data       [][]sql.NullString
	}{table, ddl, columns, data})
	if err != nil {
		t.Fatal(err)
	}
	return string(encoded)
}

func TestEmbeddedEpochOccupiedOldLineageUpgradesWithoutReinterpretation(t *testing.T) {
	requireEmbedded(t)
	ctx := t.Context()
	dir := seedMainSchemaAt(t, ctx, 72)
	conn, closeConn := openPinnedConn(t, ctx, dir)
	defer closeConn()

	execFrozenGuard(t, ctx, conn, "INSERT INTO store_epoch (id,epoch) VALUES (1,2)")
	for _, row := range []struct {
		address, id string
		epoch       int
	}{{"live0", "live", 0}, {"live1", "live", 1}, {"live2", "live", 2}, {"lapsed0", "lapsed", 0}, {"lapsed2", "lapsed", 2}} {
		if _, err := conn.ExecContext(ctx, "INSERT INTO epoch_minted_addresses (address,store_id,minted_id,minted_epoch,minted_at) VALUES (?,?,?,?,?)", row.address, "legacy-store", row.id, row.epoch, "2026-10-27 00:00:00"); err != nil {
			t.Fatal(err)
		}
	}
	mustDrain(t, ctx, conn, "CALL DOLT_ADD('-A')")
	mustDrain(t, ctx, conn, "CALL DOLT_COMMIT('-m', 'test: released epoch records')")
	checkAnswers := func() {
		t.Helper()
		for address, want := range map[string]bool{"live0": true, "lapsed0": false} {
			got, err := storeops.StillServesInTx(ctx, conn, "legacy-store", address)
			if err != nil || got != want {
				t.Fatalf("%s served=%v,error%v,want%v", address, got, err, want)
			}
		}
	}
	checkAnswers()

	before := epochFixtureState(t, ctx, conn)
	if _, err := schema.MigrateUp(ctx, conn); err != nil {
		t.Fatalf("released epoch upgrade: %v", err)
	}
	if got := currentMainVersion(t, ctx, conn); got != 74 {
		t.Fatalf("cursor%d,want74", got)
	}
	if after := epochFixtureState(t, ctx, conn); after != before {
		t.Fatal("upgrade reinterpreted or rewrote released epoch rows/DDL")
	}
	checkAnswers()
	if applied, err := schema.MigrateUp(ctx, conn); err != nil || applied != 0 {
		t.Fatalf("ordinary reopen applied%d: %v", applied, err)
	}
}

func TestEmbeddedEpochPartialDDLIsSafelyRefusedByOrdinaryMigration(t *testing.T) {
	requireEmbedded(t)
	ctx := t.Context()
	dir := seedMainSchemaAt(t, ctx, 73)
	conn, closeConn := openPinnedConn(t, ctx, dir)
	defer closeConn()
	// A durable first statement without a cursor is the killed-DDL shape.
	// Do not finish, commit, reset or replay the body behind the production guard.
	execFrozenGuard(t, ctx, conn, "ALTER TABLE epoch_minted_addresses ADD COLUMN carried_from VARCHAR(255) NULL")
	before := epochFixtureState(t, ctx, conn)
	applied, err := schema.MigrateUp(ctx, conn)
	var refused *schema.EpochMigrationCompatibilityError
	if !errors.As(err, &refused) || !strings.Contains(refused.Reason, "survival or unknown epoch columns") {
		t.Fatalf("want compatibility refusal before partial DDL repair, got %v", err)
	}
	if applied != 0 {
		t.Fatalf("partial epoch refusal applied%d", applied)
	}
	if got := currentMainVersion(t, ctx, conn); got != 73 {
		t.Fatalf("partial epoch cursor%d, want73", got)
	}
	if after := epochFixtureState(t, ctx, conn); after != before {
		t.Fatal("partial epoch state was rewritten")
	}
	found := false
	for _, table := range dirtyTableNames(t, ctx, conn) {
		found = found || table == "epoch_minted_addresses"
	}
	if !found {
		t.Fatal("partial epoch debris was silently committed or removed")
	}
}

// A nonmigrating read-only open may read unrelated data, but must not answer a
// terminally lost survival-format row through the released epoch algorithm.
func TestEmbeddedEpochReadOnlyRefusesSurvivalInterpretation(t *testing.T) {
	requireEmbedded(t)
	ctx := t.Context()
	te := newSurvivalTestEnv(t, "epochro")
	incoming := survivalStore(te.store)
	address, err := incoming.MintUnderEpoch(ctx, "store", "retained")
	if err != nil {
		t.Fatal(err)
	}
	if err := incoming.LoseVersion(ctx, "store", address); err != nil {
		t.Fatal(err)
	}
	if live, err := incoming.StillServes(ctx, "store", address); err != nil || live {
		t.Fatalf("private survival fixture live=%v: %v", live, err)
	}
	if err := te.store.Close(); err != nil {
		t.Fatal(err)
	}
	ro, err := embeddeddolt.OpenReadOnly(ctx, filepath.Dir(te.dataDir), te.database, "main")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := ro.Close(); err != nil {
			t.Error(err)
		}
	}()
	if prefix, err := ro.GetConfig(ctx, "issue_prefix"); err != nil || prefix != "epochro" {
		t.Fatalf("unrelated read prefix=%q: %v", prefix, err)
	}
	if live, err := ro.StillServes(ctx, "store", address); live || !errors.Is(err, storeops.ErrEpochFormatNotReleased) {
		t.Fatalf("released read-only live=%v: %v", live, err)
	}
	if result, err := ro.Resolve(ctx, "store", address); !errors.Is(err, storeops.ErrEpochFormatNotReleased) {
		t.Fatalf("released read-only resolve=%+v: %v", result, err)
	}
}
