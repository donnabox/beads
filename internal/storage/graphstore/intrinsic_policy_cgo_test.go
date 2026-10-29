//go:build cgo

package graphstore

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	mysql "github.com/go-sql-driver/mysql"
)

// These unchanged positive-law and transaction controls require the selected
// native implementation; parser/scalar and ordinary policy controls do not.
func TestIntrinsicNativeIgnoreLaw(t *testing.T) {
	m, err := loadIntrinsicExpected()
	if err != nil {
		t.Fatal(err)
	}
	var base [][]any
	for _, object := range m.Objects {
		if object.Ignore == "Ignore" {
			base = append(base, []any{object.Name, int64(1)})
		}
	}
	if len(base) != 13 {
		t.Fatal("base allowances differ")
	}
	if err = checkIntrinsicIgnorePolicy(m.Objects, base); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		rows [][]any
	}{
		{"broad", append(append([][]any(nil), base...), []any{"*", true})},
		{"adjunct", append(append([][]any(nil), base...), []any{"graph_preview_intrinsic_members", true})},
		{"duplicate", append(append([][]any(nil), base...), base[0])},
		{"null", [][]any{{"events", nil}}},
		{"loose bool", [][]any{{"events", "true"}}},
		{"invalid bool", [][]any{{"events", int64(2)}}},
		{"oversize", [][]any{{strings.Repeat("a", 1025), true}}},
		{"UTF8", [][]any{{string([]byte{255}), true}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := checkIntrinsicIgnorePolicy(m.Objects, tc.rows); err == nil {
				t.Fatal("policy mutation accepted")
			}
		})
	}
	// Specific negative beats a broad positive under the pinned native law.
	only := []intrinsicExpectedObject{{Name: "graph_preview_intrinsic_members", Kind: "BASE TABLE", Ignore: "DontIgnore"}}
	if err := checkIntrinsicIgnorePolicy(only, [][]any{{"*", true}, {"graph_preview_intrinsic_members", false}}); err != nil {
		t.Fatal(err)
	}
	// Native underscore is literal, not SQL LIKE's one-character wildcard.
	if err := checkIntrinsicIgnorePolicy([]intrinsicExpectedObject{{Name: "axb", Kind: "BASE TABLE", Ignore: "DontIgnore"}}, [][]any{{"a_b", true}}); err != nil {
		t.Fatal(err)
	}
}

// Full selected metadata is still acquired before closed-purpose admission.
// SQL mocks here exercise dispatch/finalization, not native schema validity.
func TestTypesTransactionPolicyFinalization(t *testing.T) {
	fault := errors.New("Types transaction finalization fault")
	for _, mode := range []string{"open", "read", "list", "create", "callback", "rollback-failure", "commit-failure", "conflict", "generic", "unknown", "write-read", "uncommitted-create", "validate-types"} {
		t.Run(mode, func(t *testing.T) {
			s, mock := transactionPolicyMock(t)
			s.options = intrinsicProfileOptions()
			mock.ExpectBegin()
			// Current wrapper checks binding, then the metadata fact independently
			// rechecks the actual binding in that SAME transaction.
			mock.ExpectQuery(`SELECT workspace, scope_url, authority_id,
        schema_version, writer_token FROM graph_preview_scope WHERE singleton = 1
 AND OCTET_LENGTH(workspace)=? AND OCTET_LENGTH(scope_url)=?`).WithArgs(len(s.options.Binding.WorkspaceID), len(s.options.Binding.ScopeURL)).WillReturnRows(sqlmock.NewRows([]string{"workspace", "scope", "authority", "schema", "token"}).AddRow(s.options.Binding.WorkspaceID, s.options.Binding.ScopeURL, s.options.Binding.AuthorityID, int64(TypesFreshSchemaVersion), "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"))
			expectIntrinsicBinding(mock, s.options, s.options.Binding.WorkspaceID)
			expectIntrinsicMetadataFixture(t, mock, s.options, true)
			p := txPolicy{purpose: txDefinitionRead}
			if mode == "create" || mode == "callback" || mode == "rollback-failure" || mode == "commit-failure" || mode == "conflict" || mode == "uncommitted-create" || mode == "validate-types" {
				p = txPolicy{purpose: txDefinitionCreate, mayWrite: true, commit: true}
			}
			unsupported := mode == "generic" || mode == "unknown" || mode == "write-read" || mode == "uncommitted-create" || mode == "validate-types"
			switch mode {
			case "open":
				p.purpose = txDefinitionOpen
			case "list":
				p.purpose = txDefinitionList
			case "generic":
				p.purpose = txOrdinary
			case "unknown":
				p.purpose = txPurpose(255)
			case "write-read":
				p.mayWrite = true
			case "uncommitted-create":
				p.commit = false
			case "validate-types":
				p.validateTypes = true
			}
			if !unsupported && p.commit && mode != "callback" && mode != "rollback-failure" {
				e := mock.ExpectCommit()
				if mode == "commit-failure" {
					e.WillReturnError(fault)
				}
				if mode == "conflict" {
					e.WillReturnError(&mysql.MySQLError{Number: 1213, SQLState: [5]byte{'4', '0', '0', '0', '1'}})
				}
			} else {
				e := mock.ExpectRollback()
				if mode == "rollback-failure" {
					e.WillReturnError(fault)
				}
			}
			calls := 0
			err := s.withTxPolicy(context.Background(), p, func(*sql.Tx) error {
				calls++
				if mode == "callback" || mode == "rollback-failure" {
					return fault
				}
				return nil
			})
			wantCalls := 1
			if unsupported {
				wantCalls = 0
			}
			if calls != wantCalls || s.db.Stats().InUse != 0 {
				t.Fatalf("callback replay/leak: %d %v", calls, err)
			}
			if unsupported && !errors.Is(err, ErrCapabilityUnavailable) {
				t.Fatalf("unsupported purpose admitted: %v", err)
			}
			if (mode == "open" || mode == "read" || mode == "list" || mode == "create") && err != nil {
				t.Fatal(err)
			}
			if (mode == "callback" || mode == "rollback-failure" || mode == "commit-failure") && !errors.Is(err, fault) {
				t.Fatalf("lost cause: %v", err)
			}
			if errors.Is(err, ErrOutcomeUnknown) != (mode == "commit-failure" || mode == "rollback-failure") {
				t.Fatalf("unknown classification: %v", err)
			}
			if errors.Is(err, ErrConflict) != (mode == "conflict") {
				t.Fatalf("conflict classification: %v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
