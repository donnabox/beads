package epochfixture

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestPrivateFixtureRefusesExistingAddressesOrWrongCursorBeforeDDL(t *testing.T) {
	for _, tc := range []struct {
		name          string
		version, rows int
	}{{"future", 75, 0}, {"older", 72, 0}, {"retained", 74, 1}} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			mock.ExpectQuery(regexp.QuoteMeta("SELECT COALESCE(MAX(version), 0) FROM schema_migrations")).WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow(tc.version))
			if tc.version == 74 {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM epoch_minted_addresses")).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(tc.rows))
			}
			if err := Prepare(context.Background(), db); err == nil {
				t.Fatal("unsafe private fixture admitted")
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestPrivateFixturePreservesAdmissionQueryFailure(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	boom := errors.New("lost selected database")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COALESCE(MAX(version), 0) FROM schema_migrations")).WillReturnError(boom)
	if err := Prepare(context.Background(), db); !errors.Is(err, boom) {
		t.Fatalf("error=%v,want original admission failure", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
