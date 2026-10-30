package schema

import (
	"github.com/steveyegge/beads/internal/testutil"
	"strings"
	"testing"
)

// Upstream's final survival-shape/behavior controls are preserved at the
// forward extension. The signed-off fork0071 CREATE remains byte-for-byte
// frozen; no premise that its old schema was never applied is used here.
func TestMigration0075ExtendsSurvivalWithoutRewritingFrozenBase(t *testing.T) {
	upSQL, err := MigrationSQL("0075_extend_epoch_survival.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"ALTER TABLE epoch_minted_addresses ADD COLUMN carried_from VARCHAR(255) NULL",
		"ALTER TABLE epoch_minted_addresses ADD COLUMN gone_at_epoch INT NULL",
		"CREATE INDEX idx_epoch_minted_addresses_carried_from ON epoch_minted_addresses (carried_from)",
		"ALTER TABLE epoch_minted_addresses ADD CONSTRAINT ck_epoch_minted_addresses_gone_not_before_mint CHECK (gone_at_epoch IS NULL OR gone_at_epoch >= minted_epoch)",
		"COLUMN_NAME = 'carried_from'", "COLUMN_NAME = 'gone_at_epoch'",
		"INDEX_NAME = 'idx_epoch_minted_addresses_carried_from'",
		"CONSTRAINT_NAME = 'ck_epoch_minted_addresses_gone_not_before_mint'",
	} {
		if !strings.Contains(upSQL, want) {
			t.Errorf("forward extension missing %q", want)
		}
	}
	statements := strings.ToUpper(stripSQLLineComments(upSQL))
	for _, forbidden := range []string{"DROP TABLE", "UPDATE ", "DELETE ", "INSERT ", "CREATE TABLE"} {
		if strings.Contains(statements, forbidden) {
			t.Errorf("forward extension contains destructive/backfill statement %q", forbidden)
		}
	}
	if got := strings.Count(statements, "PREPARE STMT FROM @SQL"); got != 4 {
		t.Errorf("got %d independently guarded DDL steps,want4", got)
	}
	base, err := MigrationSQL("0071_add_epoch_minted_addresses.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"carried_from", "gone_at_epoch"} {
		if strings.Contains(stripSQLLineComments(base), name) {
			t.Errorf("frozen base unexpectedly contains %s", name)
		}
	}
	direct := cliCompatibleMigrationSQL("0075_extend_epoch_survival.up.sql", upSQL)
	if strings.Contains(strings.ToUpper(direct), "PREPARE") || direct == upSQL {
		t.Error("fresh CLI bundle needs direct-DDL override")
	}
}

func TestLatestVersionIncludesMigration0075(t *testing.T) {
	if got := LatestVersion(); got != 75 {
		t.Fatalf("LatestVersion=%d,want75 forward epoch survival extension", got)
	}
}

// TestMigration0075RecordsSurvivalInTheAddressTableThroughDoltCLI applies the
// full migration bundle through a real `dolt` binary (skipped without one, see
// testutil.RequireDoltBinary) and checks what a pure-Go SQL-text check cannot:
// the column types and nullability Dolt reports, the index it actually built,
// and that the CHECK is enforced on both INSERT and UPDATE, with the boundary
// it is meant to have.
func TestMigration0075RecordsSurvivalInTheAddressTableThroughDoltCLI(t *testing.T) {
	testutil.RequireDoltBinary(t)

	dir := t.TempDir()
	runDoltCommand(t, dir, "init", "--name", "test", "--email", "test@example.com")
	runDoltSQL(t, dir, AllMigrationsSQL())

	requireDoltColumnShape(t, dir, "epoch_minted_addresses", "carried_from", "varchar(255)", "YES")
	requireDoltDataType(t, dir, "epoch_minted_addresses", "gone_at_epoch", "int", "YES")
	// The five columns the mint writes keep their shape, so the mint's
	// five-column INSERT is unchanged.
	requireDoltColumnShape(t, dir, "epoch_minted_addresses", "address", "varchar(255)", "NO")
	requireDoltColumnShape(t, dir, "epoch_minted_addresses", "minted_epoch", "int", "NO")
	requireDoltNoRows(t, dir, "SELECT address FROM epoch_minted_addresses", "epoch_minted_addresses")
	requireDoltCount(t, dir,
		`SELECT COUNT(*) AS c FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'epoch_minted_addresses' AND INDEX_NAME = 'idx_epoch_minted_addresses_carried_from' AND COLUMN_NAME = 'carried_from'`,
		"1")

	// A fresh mint writes the five original columns and is a root row: both
	// survival columns come out NULL with no change to the INSERT.
	runDoltSQL(t, dir, `INSERT INTO epoch_minted_addresses (address, store_id, minted_id, minted_epoch, minted_at) VALUES ('root-a', 'store', 'record-a', 3, '2026-09-01 00:00:00')`)
	requireDoltCount(t, dir, `SELECT COUNT(*) AS c FROM epoch_minted_addresses WHERE address = 'root-a' AND carried_from IS NULL AND gone_at_epoch IS NULL`, "1")

	// A carry row names the lineage root.
	runDoltSQL(t, dir, `INSERT INTO epoch_minted_addresses (address, store_id, minted_id, minted_epoch, minted_at, carried_from) VALUES ('carry-a', 'store', 'record-a', 5, '2026-09-01 00:00:00', 'root-a')`)
	requireDoltCount(t, dir, `SELECT COUNT(*) AS c FROM epoch_minted_addresses WHERE carried_from = 'root-a'`, "1")

	// The CHECK: an epoch lost in may not precede the epoch minted in.
	if err := runDoltSQLExpectingError(t, dir, `UPDATE epoch_minted_addresses SET gone_at_epoch = 2 WHERE address = 'root-a'`); err == nil {
		t.Error("UPDATE gone_at_epoch below minted_epoch succeeded, want the CHECK constraint to reject it")
	}
	if err := runDoltSQLExpectingError(t, dir, `INSERT INTO epoch_minted_addresses (address, store_id, minted_id, minted_epoch, minted_at, gone_at_epoch) VALUES ('bad-b', 'store', 'record-b', 4, '2026-09-01 00:00:00', 3)`); err == nil {
		t.Error("INSERT with gone_at_epoch below minted_epoch succeeded, want the CHECK constraint to reject it")
	}
	// Equal is allowed: a carry row is born at the new epoch and can be lost in
	// that same epoch.
	runDoltSQL(t, dir, `UPDATE epoch_minted_addresses SET gone_at_epoch = 3 WHERE address = 'root-a'`)
	requireDoltCount(t, dir, `SELECT COUNT(*) AS c FROM epoch_minted_addresses WHERE address = 'root-a' AND gone_at_epoch = 3`, "1")
	// A later epoch is allowed too.
	runDoltSQL(t, dir, `UPDATE epoch_minted_addresses SET gone_at_epoch = 9 WHERE address = 'carry-a'`)
	requireDoltCount(t, dir, `SELECT COUNT(*) AS c FROM epoch_minted_addresses WHERE address = 'carry-a' AND gone_at_epoch = 9`, "1")

	// The address stays the primary key.
	if err := runDoltSQLExpectingError(t, dir, `INSERT INTO epoch_minted_addresses (address, store_id, minted_id, minted_epoch, minted_at) VALUES ('root-a', 'store', 'record-a', 3, '2026-09-01 00:00:00')`); err == nil {
		t.Error("duplicate address insert succeeded, want primary key violation")
	}
}

// stripSQLLineComments drops "--" comment lines so a check over a migration's
// statements is not tripped by the prose in its header.
func stripSQLLineComments(sqlText string) string {
	var kept []string
	for _, line := range strings.Split(sqlText, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "--") {
			continue
		}
		kept = append(kept, line)
	}
	return strings.Join(kept, "\n")
}
