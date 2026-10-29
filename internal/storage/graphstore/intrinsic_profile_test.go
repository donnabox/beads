package graphstore

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// No engine/server is used. Strict database/sql mocks prove metadata-only
// acquisition/control flow and cleanup, not valid SHOW/native schema facts.
func intrinsicProfileTransaction(t *testing.T) (*sql.Tx, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectBegin()
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		mock.ExpectRollback()
		if err := tx.Rollback(); err != nil {
			t.Error(err)
		}
		mock.ExpectClose()
		if err := db.Close(); err != nil {
			t.Error(err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})
	return tx, mock
}

func intrinsicProfileOptions() Options {
	return Options{Database: "types_test", Binding: Binding{WorkspaceID: "/private/fixture", ScopeURL: "https://example.org/scope", AuthorityID: strings.Repeat("a", 32), SchemaVersion: TypesFreshSchemaVersion}}
}

func expectIntrinsicBinding(mock sqlmock.Sqlmock, o Options, actualWorkspace string) {
	mock.ExpectQuery(intrinsicBindingQuery).WithArgs(len(o.Binding.WorkspaceID), len(o.Binding.ScopeURL)).WillReturnRows(
		sqlmock.NewRows([]string{"workspace", "scope_url", "authority_id", "schema_version", "writer_token"}).AddRow(actualWorkspace, o.Binding.ScopeURL, o.Binding.AuthorityID, int64(14), strings.Repeat("b", 32))).RowsWillBeClosed()
}

func expectIntrinsicMetadataCandidate(t *testing.T, mock sqlmock.Sqlmock, o Options) {
	expectIntrinsicMetadataFixture(t, mock, o, false)
}

// Selected=true uses the complete checked-in projection only to exercise
// transaction control flow. It is NOT independent native schema qualification.
func expectIntrinsicMetadataFixture(t *testing.T, mock sqlmock.Sqlmock, o Options, selected bool) {
	t.Helper()
	m, err := loadIntrinsicExpected()
	if err != nil {
		t.Fatal(err)
	}
	const mode = "NO_ENGINE_SUBSTITUTION,ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES"
	mock.ExpectQuery("SELECT @@sql_mode, @@dolt_show_system_tables").WillReturnRows(sqlmock.NewRows([]string{"mode", "system"}).AddRow(mode, int64(0))).RowsWillBeClosed()
	census := sqlmock.NewRows([]string{"name", "kind"})
	for _, object := range m.Objects {
		census.AddRow(object.Name, object.Kind)
	}
	db := intrinsicQuote(o.Database)
	mock.ExpectQuery("SHOW FULL TABLES FROM " + db).WillReturnRows(census).RowsWillBeClosed()
	for _, object := range m.Objects {
		if object.Kind != "BASE TABLE" {
			continue
		}
		// Deliberately NOT the correct table structures: syntax+identity can
		// never authenticate schema equivalence or emit checked14.
		ddl := "CREATE TABLE " + intrinsicQuote(object.Name) + " (id BIGINT NOT NULL PRIMARY KEY)"
		if selected {
			ddl = m.Native.Tables[object.Name]
		}
		mock.ExpectQuery("SHOW CREATE TABLE " + db + "." + intrinsicQuote(object.Name)).WillReturnRows(sqlmock.NewRows([]string{"Table", "Create Table"}).AddRow(object.Name, ddl)).RowsWillBeClosed()
		if intrinsicNativeSelected(m) && !selected {
			// The first decoy table must refuse before any later metadata/body query.
			return
		}
	}
	columns := []string{"Field", "Type", "Collation", "Null", "Key", "Default", "Extra", "Privileges", "Comment"}
	envelope := sqlmock.NewRows(columns)
	if selected {
		for _, row := range m.Native.SchemaEnvelope {
			values := make([]driver.Value, len(row))
			for i, value := range row {
				values[i] = value
			}
			envelope.AddRow(values...)
		}
	} else {
		for _, name := range []string{"type", "name", "fragment", "extra", "sql_mode"} {
			envelope.AddRow(name, "unqualified", nil, "YES", "", nil, "", "select", "")
		}
	}
	mock.ExpectQuery("SHOW FULL COLUMNS FROM " + db + ".`dolt_schemas`").WillReturnRows(envelope).RowsWillBeClosed()
	mock.ExpectQuery("SELECT COUNT(*),CAST(LEAST(COALESCE(SUM(COALESCE(OCTET_LENGTH(type),0)+COALESCE(OCTET_LENGTH(name),0)+COALESCE(OCTET_LENGTH(fragment),0)+COALESCE(OCTET_LENGTH(extra),0)+COALESCE(OCTET_LENGTH(sql_mode),0)),0),?) AS UNSIGNED) FROM dolt_schemas").WithArgs(uint64(2*intrinsicStatementLimit + 1)).WillReturnRows(sqlmock.NewRows([]string{"count", "size"}).AddRow(int64(2), int64(100))).RowsWillBeClosed()
	views := sqlmock.NewRows([]string{"type", "name", "fragment", "extra", "sql_mode"})
	for _, object := range m.Objects {
		if object.Kind == "VIEW" {
			if selected {
				view := m.Native.Views[object.Name]
				views.AddRow("view", object.Name, view.Fragment, view.Extra, view.SQLMode)
			} else {
				views.AddRow("view", object.Name, "CREATE VIEW "+intrinsicQuote(object.Name)+" AS SELECT id FROM issues", "{}", mode)
			}
		}
	}
	mock.ExpectQuery("SELECT type,name,fragment,extra,sql_mode FROM " + db + ".`dolt_schemas` ORDER BY BINARY type,BINARY name").WillReturnRows(views).RowsWillBeClosed()
	for _, object := range m.Objects {
		if object.Kind != "VIEW" {
			continue
		}
		resolved := sqlmock.NewRows(columns)
		if selected {
			for _, row := range m.Native.Views[object.Name].Columns {
				values := make([]driver.Value, len(row))
				for i, value := range row {
					values[i] = value
				}
				resolved.AddRow(values...)
			}
		} else {
			for _, name := range object.ResolvedNames {
				resolved.AddRow(name, "unqualified", nil, "YES", "", nil, "", "select", "")
			}
		}
		mock.ExpectQuery("SHOW FULL COLUMNS FROM " + db + "." + intrinsicQuote(object.Name)).WillReturnRows(resolved).RowsWillBeClosed()
	}
	ignore := sqlmock.NewRows([]string{"pattern", "ignored"})
	for _, object := range m.Objects {
		if object.Ignore == "Ignore" {
			ignore.AddRow(object.Name, true)
		}
	}
	mock.ExpectQuery("SELECT COUNT(*),CAST(LEAST(COALESCE(SUM(COALESCE(OCTET_LENGTH(pattern),0)+COALESCE(OCTET_LENGTH(ignored),0)),0),?) AS UNSIGNED) FROM dolt_ignore").WithArgs(uint64((64 << 10) + 1)).WillReturnRows(sqlmock.NewRows([]string{"count", "size"}).AddRow(int64(13), int64(100))).RowsWillBeClosed()
	mock.ExpectQuery("SELECT pattern,ignored FROM " + db + ".`dolt_ignore` ORDER BY BINARY pattern").WillReturnRows(ignore).RowsWillBeClosed()
}

func TestIntrinsicProfileNeverEmitsUnqualifiedFact(t *testing.T) {
	tx, mock := intrinsicProfileTransaction(t)
	o := intrinsicProfileOptions()
	expectIntrinsicBinding(mock, o, o.Binding.WorkspaceID)
	expectIntrinsicMetadataCandidate(t, mock, o)
	budget := newIntrinsicMetadataBudget()
	fact, err := checkTypesFreshProfileInTx(context.Background(), tx, o, budget)
	m, loadErr := loadIntrinsicExpected()
	if loadErr != nil {
		t.Fatal(loadErr)
	}
	want := "canonical schema/view correspondence remains unqualified"
	if intrinsicNativeSelected(m) {
		want = "fixed native table correspondence"
	}
	if !errors.Is(err, ErrInvalidStore) || !strings.Contains(err.Error(), want) || fact != (checkedStoreProfile{}) {
		t.Fatalf("unqualified fact: %+v %v", fact, err)
	}
	// All expected metadata up to the selected refusal must have been read. Any intrinsic source,
	// member, snapshot, record or owned body query is unexpected and cannot
	// reach this exact terminal diagnostic. No result/body query is whitelisted.
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestIntrinsicProfileWrongBindingStopsBeforeMetadata(t *testing.T) {
	tx, mock := intrinsicProfileTransaction(t)
	o := intrinsicProfileOptions()
	expectIntrinsicBinding(mock, o, "/different")
	budget := newIntrinsicMetadataBudget()
	fact, err := checkTypesFreshProfileInTx(context.Background(), tx, o, budget)
	if !errors.Is(err, ErrInvalidStore) || !strings.Contains(err.Error(), "actual persisted intrinsic binding") || fact != (checkedStoreProfile{}) {
		t.Fatalf("wrong diagnosis: %+v %v", fact, err)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestIntrinsicMetadataRowsErrorsAndBounds(t *testing.T) {
	fault := errors.New("metadata driver failure")
	for _, mode := range []string{"success", "query", "row", "close", "extra-row", "extra-result", "bytes", "budget", "UTF8", "scalar"} {
		t.Run(mode, func(t *testing.T) {
			tx, mock := intrinsicProfileTransaction(t)
			expect := mock.ExpectQuery("SELECT fixed_metadata")
			rows := sqlmock.NewRows([]string{"value"}).AddRow("ok")
			switch mode {
			case "query":
				expect.WillReturnError(fault)
			case "row":
				expect.WillReturnRows(rows.RowError(0, fault)).RowsWillBeClosed()
			case "close":
				expect.WillReturnRows(rows.CloseError(fault)).RowsWillBeClosed()
			case "extra-row":
				expect.WillReturnRows(rows.AddRow("extra")).RowsWillBeClosed()
			case "extra-result":
				expect.WillReturnRows(rows, sqlmock.NewRows([]string{"value"}).AddRow("extra")).RowsWillBeClosed()
			case "bytes":
				expect.WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow(strings.Repeat("x", intrinsicStatementLimit+1))).RowsWillBeClosed()
			case "UTF8":
				expect.WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow([]byte{255})).RowsWillBeClosed()
			case "scalar":
				expect.WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow(float64(1))).RowsWillBeClosed()
			default:
				expect.WillReturnRows(rows).RowsWillBeClosed()
			}
			limit := PreviewCurrentReadByteLimit
			if mode == "budget" {
				limit = 1
			}
			budget := &intrinsicMetadataBudget{remaining: limit}
			got, err := readIntrinsicMetadataRows(context.Background(), tx, budget, "SELECT fixed_metadata", 1, 1)
			if mode == "success" {
				if err != nil || !reflect.DeepEqual(got, [][]any{{"ok"}}) {
					t.Fatalf("positive acquisition: %#v %v", got, err)
				}
			} else {
				if err == nil || got != nil {
					t.Fatalf("failure leaked rows: %#v %v", got, err)
				}
				if (mode == "query" || mode == "row" || mode == "close") && !errors.Is(err, fault) {
					t.Fatalf("driver cause lost: %v", err)
				}
			}
		})
	}
}

// An exhausted metadata allowance must stop later record/context acquisition
// before ANY SQL body query. The strict mock has no query expectations.
func TestIntrinsicOperationBudgetStopsBeforeBodyAcquisition(t *testing.T) {
	for _, which := range []string{"record", "context"} {
		t.Run(which, func(t *testing.T) {
			tx, _ := intrinsicProfileTransaction(t)
			budget := &intrinsicMetadataBudget{remaining: 64}
			if err := budget.charge(64); err != nil {
				t.Fatal(err)
			}
			o := intrinsicProfileOptions()
			if which == "context" {
				plan, e := loadIntrinsicContextInTx(context.Background(), tx, o, strings.Repeat("c", 32), budget)
				if !errors.Is(e, ErrLimitExceeded) || plan != nil {
					t.Fatalf("context: plan=%v err=%v", plan, e)
				}
			} else {
				var doc IntrinsicDocument
				e := (&Store{options: o}).readIntrinsicDocumentInTx(context.Background(), tx, "beads/type-for-bead-type", "", &doc, budget)
				if !errors.Is(e, ErrLimitExceeded) || len(doc.Body) != 0 {
					t.Fatalf("record: doc=%v err=%v", doc, e)
				}
			}
		})
	}
}
