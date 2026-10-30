package issueops

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func releasedEpochColumnRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"COLUMN_NAME", "COLUMN_TYPE", "IS_NULLABLE", "COLUMN_DEFAULT", "EXTRA"}).
		AddRow("address", "varchar(255)", "NO", nil, "").
		AddRow("store_id", "varchar(255)", "NO", nil, "").
		AddRow("minted_id", "varchar(255)", "NO", nil, "").
		AddRow("minted_epoch", "int", "NO", nil, "").
		AddRow("minted_at", "datetime", "NO", nil, "")
}

func expectReleasedEpochColumns(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(regexp.QuoteMeta(releasedEpochColumnsQuery)).WillReturnRows(releasedEpochColumnRows()).RowsWillBeClosed()
}

// These callers include raw DBTX/UOW and nonmigrating gateway/read-only paths.
// With no other SQL expectations, any epoch row read or write fails the test.
func TestReleasedEpochEntryPointsRefuseOtherFormatsBeforeRowAccess(t *testing.T) {
	ctx := context.Background()
	calls := map[string]func(DBTX) error{
		"current": func(tx DBTX) error { _, err := CurrentEpochInTx(ctx, tx, "store"); return err },
		"bump":    func(tx DBTX) error { _, err := BumpEpochInTx(ctx, tx, "store", "restore"); return err },
		"mint":    func(tx DBTX) error { _, err := MintUnderEpochInTx(ctx, tx, "store", "id"); return err },
		"serves":  func(tx DBTX) error { _, err := StillServesInTx(ctx, tx, "store", "lost-address"); return err },
		"resolve": func(tx DBTX) error { _, err := ResolveEpochInTx(ctx, tx, "store", "lost-address"); return err },
		"remint":  func(tx DBTX) error { _, err := CurrentAddressForInTx(ctx, tx, "store", "lost-address"); return err },
	}
	for name, call := range calls {
		for _, format := range []string{"survival", "extra", "missing", "wrong-type"} {
			t.Run(name+"/"+format, func(t *testing.T) {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatal(err)
				}
				defer func() { _ = db.Close() }()
				rows := releasedEpochColumnRows()
				switch format {
				case "survival":
					rows.AddRow("carried_from", "varchar(255)", "YES", nil, "").AddRow("gone_at_epoch", "int", "YES", nil, "")
				case "extra":
					rows.AddRow("extension", "int", "YES", nil, "")
				case "missing":
					rows = sqlmock.NewRows([]string{"COLUMN_NAME", "COLUMN_TYPE", "IS_NULLABLE", "COLUMN_DEFAULT", "EXTRA"})
				case "wrong-type":
					rows = sqlmock.NewRows([]string{"COLUMN_NAME", "COLUMN_TYPE", "IS_NULLABLE", "COLUMN_DEFAULT", "EXTRA"}).AddRow("address", "text", "NO", nil, "")
				}
				mock.ExpectQuery(regexp.QuoteMeta(releasedEpochColumnsQuery)).WillReturnRows(rows).RowsWillBeClosed()
				if err := call(db); !errors.Is(err, ErrEpochFormatNotReleased) {
					t.Fatalf("want incompatible format, got %v", err)
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestReleasedEpochMetadataErrorsAreNotFreshDefaults(t *testing.T) {
	cause := errors.New("metadata unavailable")
	for _, failure := range []string{"query", "scan", "rows", "close"} {
		t.Run(failure, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = db.Close() }()
			query := mock.ExpectQuery(regexp.QuoteMeta(releasedEpochColumnsQuery))
			switch failure {
			case "query":
				query.WillReturnError(cause)
			case "scan":
				query.WillReturnRows(sqlmock.NewRows([]string{"COLUMN_NAME", "COLUMN_TYPE", "IS_NULLABLE", "COLUMN_DEFAULT", "EXTRA"}).AddRow(nil, "int", "NO", nil, ""))
			case "rows":
				query.WillReturnRows(releasedEpochColumnRows().RowError(1, cause))
			case "close":
				query.WillReturnRows(releasedEpochColumnRows().CloseError(cause))
			}
			_, err = CurrentEpochInTx(context.Background(), db, "store")
			if err == nil {
				t.Fatal("metadata failure admitted as released format")
			}
			if failure != "scan" && !errors.Is(err, cause) {
				t.Fatalf("lost metadata cause: %v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestReleasedEpochInvalidMintRefusesBeforeMetadata(t *testing.T) {
	for _, args := range [][2]string{{"bad:store", "id"}, {"store", "bad:id"}} {
		if _, err := MintUnderEpochInTx(context.Background(), nil, args[0], args[1]); !errors.Is(err, ErrEpochAddressSeparator) {
			t.Fatalf("want separator refusal before SQL, got %v", err)
		}
	}
}
