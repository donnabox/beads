package schema

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestEpochCheckMeaningPreservesGrouping(t *testing.T) {
	for _, text := range []string{
		"gone_at_epoch IS NULL OR gone_at_epoch >= minted_epoch",
		"((`gone_at_epoch` is null) or (`gone_at_epoch` >= `minted_epoch`))",
		"(((gone_at_epoch IS NULL) OR (gone_at_epoch >= minted_epoch)))",
	} {
		if !epochCheckMatches(text) {
			t.Errorf("expected real equivalent CHECK rejected: %s", text)
		}
	}
	for _, text := range []string{
		"(gone_at_epoch IS NULL OR gone_at_epoch) >= minted_epoch",
		"gone_at_epoch IS NULL OR (gone_at_epoch >= minted_epoch) AND false",
		"gone_at_epoch IS NULL OR gone_at_epoch > minted_epoch",
		"gone_at_epoch IS NOT NULL OR gone_at_epoch >= minted_epoch",
		"gone_at_epoch IS NULL AND gone_at_epoch >= minted_epoch",
		"gone_at_epoch IS NULL OR gone_at_epoch >= minted_epoch OR true",
		"(gone_at_epoch IS NULL OR gone_at_epoch >= minted_epoch",
		"gone_at_epoch IS NULL OR gone_at_epoch >= minted_epoch)",
		"", "true",
	} {
		if epochCheckMatches(text) {
			t.Errorf("incompatible CHECK accepted: %s", text)
		}
	}
}

type admissionNoWrites struct {
	*sql.DB
	writes int
}

func (d *admissionNoWrites) ExecContext(context.Context, string, ...any) (sql.Result, error) {
	d.writes++
	return nil, errors.New("admission attempted a write")
}

func expectAdmissionTables(mock sqlmock.Sqlmock, cursor, epoch bool) {
	for _, table := range []struct {
		name    string
		present bool
	}{{"schema_migrations", cursor}, {"epoch_minted_addresses", epoch}} {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM information_schema.schemata WHERE schema_name = DATABASE()")).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		rows := sqlmock.NewRows([]string{"table"})
		if table.present {
			rows.AddRow(table.name)
		}
		mock.ExpectQuery(regexp.QuoteMeta("SHOW TABLES")).WillReturnRows(rows)
	}
}

func TestEpochAdmissionRefusesAmbiguousLineageBeforeWrites(t *testing.T) {
	for _, test := range []struct {
		name       string
		version    int
		hashColumn bool
		hash       any
		hasRow     bool
	}{
		{"future cursor", 76, false, nil, false},
		{"missing hash column", 70, false, nil, false},
		{"NULL hash", 70, true, nil, true},
		{"missing collision row", 70, true, nil, false},
		{"foreign hash", 70, true, "foreign-upstream-lineage", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			guard := &admissionNoWrites{DB: db}
			expectAdmissionTables(mock, true, false)
			mock.ExpectQuery(regexp.QuoteMeta("SELECT COALESCE(MAX(version), 0) FROM schema_migrations")).WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(test.version))
			if test.version <= forkEpochExtensionVersion {
				columns := sqlmock.NewRows([]string{"Field", "Type", "Null", "Key", "Default", "Extra"})
				if test.hashColumn {
					columns.AddRow("content_hash", "char(64)", "YES", "", nil, "")
				}
				mock.ExpectQuery(regexp.QuoteMeta("SHOW COLUMNS FROM schema_migrations LIKE 'content_hash'")).WillReturnRows(columns)
				if test.hashColumn {
					rows := sqlmock.NewRows([]string{"version", "content_hash"})
					if test.hasRow {
						rows.AddRow(70, test.hash)
					}
					mock.ExpectQuery(regexp.QuoteMeta("SELECT version, content_hash FROM schema_migrations WHERE version BETWEEN 70 AND 75 ORDER BY version")).WillReturnRows(rows)
				}
			}
			err = checkForkEpochAdmission(context.Background(), guard, forkEpochExtensionVersion)
			var refused *EpochMigrationCompatibilityError
			if !errors.As(err, &refused) {
				t.Fatalf("want concrete compatibility refusal, got %v", err)
			}
			if guard.writes != 0 {
				t.Errorf("admission performed %d writes", guard.writes)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// These fixtures distinguish old/partial/completed metadata. In particular,
// an index name and a CHECK name alone must never establish safe semantics.
func expectEpochMetadata(mock sqlmock.Sqlmock, full, prefix bool, enforced string) {
	columns := sqlmock.NewRows([]string{"Field", "Type", "Null", "Key", "Default", "Extra"}).
		AddRow("address", "varchar(255)", "NO", "PRI", nil, "").
		AddRow("store_id", "varchar(255)", "NO", "", nil, "").
		AddRow("minted_id", "varchar(255)", "NO", "", nil, "").
		AddRow("minted_epoch", "int", "NO", "", nil, "").
		AddRow("minted_at", "datetime", "NO", "", nil, "")
	if full {
		columns.AddRow("carried_from", "varchar(255)", "YES", "MUL", nil, "").AddRow("gone_at_epoch", "int", "YES", "", nil, "")
	}
	mock.ExpectQuery(regexp.QuoteMeta("SHOW COLUMNS FROM epoch_minted_addresses")).WillReturnRows(columns)
	indices := sqlmock.NewRows([]string{"INDEX_NAME", "NON_UNIQUE", "SEQ_IN_INDEX", "COLUMN_NAME", "SUB_PART"})
	if prefix {
		indices.AddRow("PRIMARY", 0, 1, "address", 10)
	} else {
		indices.AddRow("PRIMARY", 0, 1, "address", nil)
	}
	if full {
		indices.AddRow("idx_epoch_minted_addresses_carried_from", 1, 1, "carried_from", nil)
	}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT INDEX_NAME, NON_UNIQUE, SEQ_IN_INDEX, COLUMN_NAME, SUB_PART FROM information_schema.statistics WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'epoch_minted_addresses' ORDER BY INDEX_NAME, SEQ_IN_INDEX")).WillReturnRows(indices)
	if prefix {
		return
	} // Refusal must precede the next metadata probe.
	constraints := sqlmock.NewRows([]string{"CONSTRAINT_NAME", "CONSTRAINT_TYPE", "CHECK_CLAUSE", "ENFORCED"}).AddRow("PRIMARY", "PRIMARY KEY", nil, "YES")
	if full {
		constraints.AddRow("ck_epoch_minted_addresses_gone_not_before_mint", "CHECK", "gone_at_epoch IS NULL OR gone_at_epoch >= minted_epoch", enforced)
	}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT tc.CONSTRAINT_NAME, tc.CONSTRAINT_TYPE, cc.CHECK_CLAUSE, tc.ENFORCED FROM information_schema.table_constraints tc LEFT JOIN information_schema.check_constraints cc ON cc.CONSTRAINT_SCHEMA = tc.CONSTRAINT_SCHEMA AND cc.CONSTRAINT_NAME = tc.CONSTRAINT_NAME WHERE tc.TABLE_SCHEMA = DATABASE() AND tc.TABLE_NAME = 'epoch_minted_addresses'")).WillReturnRows(constraints)
}

func expectSelectedEpochLineage(t *testing.T, mock sqlmock.Sqlmock, version int) {
	t.Helper()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COALESCE(MAX(version), 0) FROM schema_migrations")).WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(version))
	mock.ExpectQuery(regexp.QuoteMeta("SHOW COLUMNS FROM schema_migrations LIKE 'content_hash'")).WillReturnRows(sqlmock.NewRows([]string{"Field"}).AddRow("content_hash"))
	rows := sqlmock.NewRows([]string{"version", "content_hash"})
	for v := 70; v <= version; v++ {
		data, err := mainSource.files.ReadFile("migrations/" + forkEpochLineageFiles[v])
		if err != nil {
			t.Fatal(err)
		}
		rows.AddRow(v, fmt.Sprintf("%x", sha256.Sum256(data)))
	}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT version, content_hash FROM schema_migrations WHERE version BETWEEN 70 AND 75 ORDER BY version")).WillReturnRows(rows)
}

func TestEpochAdmissionDistinguishesOldAndCompletedData(t *testing.T) {
	for _, test := range []struct {
		name                   string
		version, target        int
		full, occupied, refuse bool
	}{
		{"empty old table without writer fence", 72, 75, false, false, true},
		{"occupied old table", 72, 75, false, true, true},
		{"occupied pending extension", 73, 75, true, true, true},
		{"empty pending extension without writer fence", 73, 75, true, false, true},
		{"completed extension retains rows", 75, 75, true, true, false},
		{"completed cursor with old shape", 75, 75, false, false, true},
		{"default target refuses occupied old", 72, 0, false, true, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			guard := &admissionNoWrites{DB: db}
			expectAdmissionTables(mock, true, true)
			expectSelectedEpochLineage(t, mock, test.version)
			if test.version >= forkEpochExtensionVersion {
				expectEpochMetadata(mock, test.full, false, "YES")
			}
			err = checkForkEpochAdmission(context.Background(), guard, test.target)
			if test.refuse {
				var refused *EpochMigrationCompatibilityError
				if !errors.As(err, &refused) {
					t.Fatalf("want compatibility refusal, got %v", err)
				}
			} else if err != nil {
				t.Fatalf("recognized fixture refused: %v", err)
			}
			if guard.writes != 0 {
				t.Fatalf("admission wrote %d times", guard.writes)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestEpochAdmissionRejectsChangedIndexOrDisabledCheck(t *testing.T) {
	for _, test := range []struct {
		name     string
		prefix   bool
		enforced string
	}{
		{"prefix primary key", true, "YES"},
		{"unenforced loss check", false, "NO"},
	} {
		t.Run(test.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			guard := &admissionNoWrites{DB: db}
			expectAdmissionTables(mock, true, true)
			expectSelectedEpochLineage(t, mock, 75)
			expectEpochMetadata(mock, true, test.prefix, test.enforced)
			err = checkForkEpochAdmission(context.Background(), guard, 75)
			var refused *EpochMigrationCompatibilityError
			if !errors.As(err, &refused) {
				t.Fatalf("unsafe metadata admitted: %v", err)
			}
			if guard.writes != 0 {
				t.Fatalf("admission wrote %d times", guard.writes)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestEpochAdmissionTableDiscoveryFailureIsNotFresh(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	guard := &admissionNoWrites{DB: db}
	failed := errors.New("catalog unavailable")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM information_schema.schemata WHERE schema_name = DATABASE()")).WillReturnError(failed)
	err = checkForkEpochAdmission(context.Background(), guard, 75)
	if !errors.Is(err, failed) {
		t.Fatalf("catalog failure treated as fresh or lost: %v", err)
	}
	if guard.writes != 0 {
		t.Fatalf("admission wrote %d times", guard.writes)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

// expectCurrentEpochAdmission is separate from migrationWorkNeeded: a cursor
// at latest does not by itself establish compatible retained epoch semantics.
func expectCurrentEpochAdmission(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	expectAdmissionTables(mock, true, true)
	expectSelectedEpochLineage(t, mock, LatestVersion())
	expectEpochMetadata(mock, true, false, "YES")
}

// Public migration consent and the schema-skew read escape hatch must not
// become authority to reinterpret retained pre-extension addresses.
func TestMigrationAuthorityDoesNotBypassOccupiedEpochRefusal(t *testing.T) {
	for _, route := range []struct {
		name    string
		grant   func()
		migrate func(context.Context, DBConn) (int, error)
	}{
		{"local consent", func() { SetLocalMigrateConsent(true) }, MigrateUp},
		{"force consent", func() { SetForceAllowRemoteMigrate(true) }, MigrateUp},
		{"historical API at latest", func() {}, func(ctx context.Context, db DBConn) (int, error) { return MigrateUpTo(ctx, db, 75) }},
		{"bootstrap consent seam", func() {}, migrateUpConsented},
	} {
		t.Run(route.name, func(t *testing.T) {
			resetConsentState(t)
			t.Setenv("BD_IGNORE_SCHEMA_SKEW", "1")
			route.grant()
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			guard := &admissionNoWrites{DB: db}
			expectAdmissionTables(mock, true, true)
			expectSelectedEpochLineage(t, mock, 72)
			applied, err := route.migrate(context.Background(), guard)
			var refused *EpochMigrationCompatibilityError
			if !errors.As(err, &refused) {
				t.Fatalf("want compatibility refusal, got %v", err)
			}
			if applied != 0 || guard.writes != 0 {
				t.Fatalf("refused migration applied%d/wrote%d", applied, guard.writes)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// Pending fixtures model a genuine selected common/fork lineage; they never
// invent an epoch table before the migration that created it.
func expectPreExtensionEpochAdmission(t *testing.T, mock sqlmock.Sqlmock, version int) {
	t.Helper()
	if version < 0 || version >= forkEpochExtensionVersion {
		t.Fatalf("not a pre-extension fixture: %d", version)
	}
	epoch := version >= 71
	expectAdmissionTables(mock, true, epoch)
	if version >= 70 {
		expectSelectedEpochLineage(t, mock, version)
	} else {
		expectScalar(mock, "SELECT COALESCE(MAX(version), 0) FROM schema_migrations", "version", version)
	}
}
