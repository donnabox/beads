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
	"github.com/steveyegge/beads/internal/types"
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

// Capture exact migration/ignore/dirty metadata plus commit identity to prove
// the explicit ahead-schema issue-read route performs no migration or repair.
func aheadIssueReadState(t *testing.T, te *testEnv) string {
	t.Helper()
	ctx := t.Context()
	db, cleanup, err := embeddeddolt.OpenSQL(ctx, te.dataDir, te.database, "main")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := cleanup(); err != nil {
			t.Error(err)
		}
	}()
	var state [][][]sql.NullString
	for _, query := range []string{
		"SELECT DOLT_HASHOF('HEAD')",
		"SELECT version, content_hash FROM schema_migrations ORDER BY version",
		"SELECT * FROM dolt_ignore ORDER BY pattern",
		"SELECT * FROM dolt_status ORDER BY table_name",
	} {
		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			t.Fatal(err)
		}
		cols, err := rows.Columns()
		if err != nil {
			t.Fatal(err)
		}
		var data [][]sql.NullString
		for rows.Next() {
			cells := make([]sql.NullString, len(cols))
			dest := make([]any, len(cols))
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
		state = append(state, data)
	}
	encoded, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	return string(encoded)
}

func TestEmbeddedAheadIssueReadOverrideDoesNotMigrateOrPermitWrites(t *testing.T) {
	requireEmbedded(t)
	for _, survival := range []bool{false, true} {
		name := "released epoch"
		if survival {
			name = "deferred survival epoch"
		}
		t.Run(name, func(t *testing.T) {
			ctx := t.Context()
			te := newTestEnv(t, "ahead")
			if survival {
				if err := embeddeddolt.PrepareSurvivalTestStore(ctx, te.store); err != nil {
					t.Fatal(err)
				}
			}
			issue := &types.Issue{ID: "ahead-1", Title: "readable ahead", Status: types.StatusOpen, Priority: 1, IssueType: types.TypeTask}
			if err := te.store.CreateIssue(ctx, issue, "tester"); err != nil {
				t.Fatal(err)
			}
			if err := te.store.Commit(ctx, "seed ahead issue fixture"); err != nil {
				t.Fatal(err)
			}
			if err := te.store.Close(); err != nil {
				t.Fatal(err)
			}
			te.exec(t, ctx, "INSERT INTO schema_migrations (version, content_hash) VALUES (?, ?)", schema.LatestVersion()+1, "unknown-future-fixture")
			before := aheadIssueReadState(t, te)
			t.Setenv("BD_IGNORE_SCHEMA_SKEW", "")
			if store, err := embeddeddolt.OpenForReadOnlyCommand(ctx, filepath.Dir(te.dataDir), te.database, "main"); !schema.IsSchemaSkewError(err) {
				if store != nil {
					_ = store.Close()
				}
				t.Fatalf("default ahead read must refuse: %v", err)
			}
			t.Setenv("BD_IGNORE_SCHEMA_SKEW", "1")
			ro, err := embeddeddolt.OpenForReadOnlyCommand(ctx, filepath.Dir(te.dataDir), te.database, "main")
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := ro.Close(); err != nil {
					t.Error(err)
				}
			}()
			got, err := ro.GetIssue(ctx, issue.ID)
			if err != nil || got == nil || got.Title != issue.Title {
				t.Fatalf("issue read=%+v: %v", got, err)
			}
			for name, call := range map[string]func() error{
				"config": func() error { return ro.SetConfig(ctx, "issue_prefix", "changed") },
				"create": func() error {
					return ro.CreateIssue(ctx, &types.Issue{ID: "ahead-2", Title: "must refuse", Status: types.StatusOpen, Priority: 1, IssueType: types.TypeTask}, "tester")
				},
				"commit":  func() error { return ro.Commit(ctx, "must refuse") },
				"migrate": func() error { _, err := ro.ApplySchemaMigrations(ctx); return err },
			} {
				if err := call(); !errors.Is(err, embeddeddolt.ErrReadOnly) {
					t.Fatalf("%s mutation must be read-only refused: %v", name, err)
				}
			}
			if survival {
				if live, err := ro.StillServes(ctx, "store", "address"); live || !errors.Is(err, storeops.ErrEpochFormatNotReleased) {
					t.Fatalf("survival epoch still-serves=%v: %v", live, err)
				}
				if result, err := ro.Resolve(ctx, "store", "address"); !errors.Is(err, storeops.ErrEpochFormatNotReleased) {
					t.Fatalf("survival epoch resolve=%+v: %v", result, err)
				}
			}
			for name, open := range map[string]func() (*embeddeddolt.EmbeddedDoltStore, error){
				"strict": func() (*embeddeddolt.EmbeddedDoltStore, error) {
					return embeddeddolt.Open(ctx, filepath.Dir(te.dataDir), te.database, "main")
				},
				"reconcile": func() (*embeddeddolt.EmbeddedDoltStore, error) {
					return embeddeddolt.OpenForWorkingSetReconcile(ctx, filepath.Dir(te.dataDir), te.database, "main")
				},
			} {
				store, err := open()
				if store != nil {
					_ = store.Close()
				}
				var incompatible *schema.EpochMigrationCompatibilityError
				if !errors.As(err, &incompatible) {
					t.Fatalf("%s must retain migration admission refusal: %v", name, err)
				}
			}
			if after := aheadIssueReadState(t, te); after != before {
				t.Fatal("ahead issue read or refused mutation changed HEAD/cursor hashes/ignore/dirty metadata")
			}
		})
	}
}

func TestEmbeddedAheadIssueReadOverrideRefusesMissingProjection(t *testing.T) {
	requireEmbedded(t)
	ctx := t.Context()
	te := newTestEnv(t, "missing")
	if err := te.store.Close(); err != nil {
		t.Fatal(err)
	}
	te.exec(t, ctx, "ALTER TABLE issues DROP COLUMN title")
	te.exec(t, ctx, "INSERT INTO schema_migrations (version, content_hash) VALUES (?, ?)", schema.LatestVersion()+1, "unknown-future-fixture")
	before := aheadIssueReadState(t, te)
	t.Setenv("BD_IGNORE_SCHEMA_SKEW", "1")
	store, err := embeddeddolt.OpenForReadOnlyCommand(ctx, filepath.Dir(te.dataDir), te.database, "main")
	if store != nil {
		_ = store.Close()
	}
	if err == nil || !strings.Contains(err.Error(), "title") {
		t.Fatalf("missing ahead projection must refuse: %v", err)
	}
	if after := aheadIssueReadState(t, te); after != before {
		t.Fatal("missing projection refusal changed fixture")
	}
}
