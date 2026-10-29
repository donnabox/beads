package graphstore

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	mysql "github.com/go-sql-driver/mysql"
)

// These tests exercise database/sql transaction finalization with strict SQL
// expectations. They model no graph rows, engine state or persistence guarantee.
func TestTransactionPolicyFinalization(t *testing.T) {
	callbackFault := errors.New("validation failed after staged writes")
	rollbackFault := errors.New("rollback response unavailable")
	commitFault := errors.New("commit response unavailable")
	conflict := &mysql.MySQLError{Number: 1213, SQLState: [5]byte{'4', '0', '0', '0', '1'}}
	for _, tc := range []struct {
		name             string
		mode             string
		callbackErr      error
		rollbackErr      error
		commitErr        error
		callbackRollback bool
		want             []error
		unknown          bool
		conflict         bool
	}{
		{name: "dry-run-success", mode: "dry-run"},
		{name: "dry-run-rollback-failure", mode: "dry-run", rollbackErr: rollbackFault, want: []error{rollbackFault}, unknown: true},
		{name: "dry-run-validation-failure", mode: "dry-run", callbackErr: callbackFault, want: []error{callbackFault}},
		{name: "dry-run-validation-and-rollback-failure", mode: "dry-run", callbackErr: callbackFault, rollbackErr: rollbackFault, want: []error{callbackFault, rollbackFault}, unknown: true},
		{name: "dry-run-cancellation-and-rollback-failure", mode: "dry-run", callbackErr: context.Canceled, rollbackErr: rollbackFault, want: []error{context.Canceled, rollbackFault}, unknown: true},
		{name: "dry-run-already-rolled-back", mode: "dry-run", callbackErr: context.Canceled, callbackRollback: true, want: []error{context.Canceled, sql.ErrTxDone}, unknown: true},
		{name: "reader-success", mode: "read"},
		{name: "reader-rollback-failure", mode: "read", rollbackErr: rollbackFault, want: []error{rollbackFault}},
		{name: "reader-callback-and-rollback-failure", mode: "read", callbackErr: callbackFault, rollbackErr: rollbackFault, want: []error{callbackFault, rollbackFault}},
		{name: "writer-success", mode: "write"},
		{name: "writer-validation-failure", mode: "write", callbackErr: callbackFault, want: []error{callbackFault}},
		{name: "writer-validation-and-rollback-failure", mode: "write", callbackErr: callbackFault, rollbackErr: rollbackFault, want: []error{callbackFault, rollbackFault}, unknown: true},
		{name: "writer-commit-failure", mode: "write", commitErr: commitFault, want: []error{commitFault}, unknown: true},
		{name: "writer-confirmed-conflict", mode: "write", commitErr: conflict, want: []error{conflict}, conflict: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, mock := transactionPolicyMock(t)
			mock.ExpectBegin()
			if tc.mode == "write" && tc.callbackErr == nil {
				mock.ExpectCommit().WillReturnError(tc.commitErr)
			} else {
				mock.ExpectRollback().WillReturnError(tc.rollbackErr)
			}
			attempts := 0
			callback := func(tx *sql.Tx) error {
				attempts++
				if tc.callbackRollback {
					// Deterministic ErrTxDone at finalization, without racing the
					// database/sql context-cancellation watcher or starting an engine.
					if err := tx.Rollback(); err != nil {
						t.Fatalf("callback rollback: %v", err)
					}
				}
				return tc.callbackErr
			}
			var err error
			switch tc.mode {
			case "read":
				err = s.withTx(context.Background(), false, callback)
			case "write":
				err = s.withTx(context.Background(), true, callback)
			case "dry-run":
				err = s.withTxPolicy(context.Background(), txPolicy{mayWrite: true, commit: false}, callback)
			default:
				t.Fatalf("unknown test mode %q", tc.mode)
			}
			if attempts != 1 || s.db.Stats().InUse != 0 {
				t.Fatalf("callback replay or connection leak: attempts=%d in-use=%d", attempts, s.db.Stats().InUse)
			}
			if len(tc.want) == 0 && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, cause := range tc.want {
				if !errors.Is(err, cause) {
					t.Errorf("lost cause %v: %v", cause, err)
				}
			}
			if errors.Is(err, ErrOutcomeUnknown) != tc.unknown || errors.Is(err, ErrConflict) != tc.conflict {
				t.Fatalf("wrong outcome classification: unknown=%t conflict=%t err=%v", tc.unknown, tc.conflict, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestTransactionPolicyContextBeforeOpen(t *testing.T) {
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	for _, tc := range []struct {
		name string
		ctx  context.Context
		want error
	}{
		{name: "nil"},
		{name: "canceled", ctx: canceled, want: context.Canceled},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// No database is configured: admission must return before opening one.
			s := &Store{}
			called := false
			err := s.withTxPolicy(tc.ctx, txPolicy{mayWrite: true}, func(*sql.Tx) error {
				called = true
				return nil
			})
			if called || err == nil || (tc.want != nil && !errors.Is(err, tc.want)) {
				t.Fatalf("context admission: called=%t err=%v", called, err)
			}
		})
	}
}

func TestSelectedTypeTransactionValidation(t *testing.T) {
	validationFault := errors.New("selected graph validation unavailable")
	const relevantQuery = "SELECT EXISTS(SELECT 1 FROM graph_preview_catalog WHERE resource_kind='bead' AND allocation_state='live' AND backing='generic') OR EXISTS(SELECT 1 FROM graph_preview_links)"
	for _, mode := range []string{"native", "selected", "selected-refusal", "import-dry-run"} {
		t.Run(mode, func(t *testing.T) {
			s, mock := transactionPolicyMock(t)
			mock.ExpectBegin()
			// Strict ordering proves that final validation follows the callback.
			mock.ExpectExec("staged operation").WillReturnResult(sqlmock.NewResult(0, 1))
			if mode != "native" {
				if mode == "selected-refusal" {
					mock.ExpectQuery(relevantQuery).WillReturnError(validationFault)
				} else {
					mock.ExpectQuery(relevantQuery).WillReturnRows(sqlmock.NewRows([]string{"relevant"}).AddRow(false))
				}
			}
			if mode == "selected-refusal" || mode == "import-dry-run" {
				mock.ExpectRollback()
			} else {
				mock.ExpectCommit()
			}
			callback := func(tx *sql.Tx) error {
				_, err := tx.ExecContext(context.Background(), "staged operation")
				return err
			}
			var err error
			switch mode {
			case "native":
				err = s.withTx(context.Background(), true, callback)
			case "selected", "selected-refusal":
				err = s.withTypeValidationTx(context.Background(), callback)
			case "import-dry-run":
				err = s.withTxPolicy(context.Background(), txPolicy{mayWrite: true, commit: false, validateTypes: true}, callback)
			}
			if mode == "selected-refusal" {
				if !errors.Is(err, validationFault) || errors.Is(err, ErrOutcomeUnknown) {
					t.Fatalf("validation refusal: %v", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if s.db.Stats().InUse != 0 {
				t.Fatal("transaction leaked a connection")
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestTransactionPolicyAdmissionBeforeBegin(t *testing.T) {
	setFault := errors.New("session setting unavailable")
	for _, tc := range []struct {
		name   string
		setErr error
		branch string
		want   error
	}{
		{name: "session-setting-failure", setErr: setFault, want: setFault},
		{name: "branch-mismatch", branch: "feature", want: ErrInvalidStore},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, mock := transactionPolicyConnection(t)
			for i, statement := range []string{"SET @@dolt_transaction_commit = 0", "SET @@dolt_force_transaction_commit = 0", "SET @@dolt_allow_commit_conflicts = 0", "SET @@time_zone = '+00:00'"} {
				if tc.setErr != nil && i == 2 {
					mock.ExpectExec(statement).WillReturnError(tc.setErr)
					break
				}
				mock.ExpectExec(statement).WillReturnResult(sqlmock.NewResult(0, 0))
			}
			if tc.setErr == nil {
				mock.ExpectQuery("SELECT active_branch()").WillReturnRows(sqlmock.NewRows([]string{"branch"}).AddRow(tc.branch))
			}
			// No BEGIN, COMMIT or ROLLBACK is expected: admission must release
			// the checked-out connection without starting a transaction.
			attempts := 0
			err := s.withTxPolicy(context.Background(), txPolicy{mayWrite: true}, func(*sql.Tx) error {
				attempts++
				return nil
			})
			if attempts != 0 || s.db.Stats().InUse != 0 {
				t.Fatalf("pre-BEGIN admission leaked execution or connection: attempts=%d in-use=%d", attempts, s.db.Stats().InUse)
			}
			if !errors.Is(err, tc.want) || errors.Is(err, ErrOutcomeUnknown) || errors.Is(err, ErrConflict) {
				t.Fatalf("pre-BEGIN admission lost cause or changed outcome classification: %v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func transactionPolicyMock(t *testing.T) (*Store, sqlmock.Sqlmock) {
	t.Helper()
	s, mock := transactionPolicyConnection(t)
	for _, statement := range []string{"SET @@dolt_transaction_commit = 0", "SET @@dolt_force_transaction_commit = 0", "SET @@dolt_allow_commit_conflicts = 0", "SET @@time_zone = '+00:00'"} {
		mock.ExpectExec(statement).WillReturnResult(sqlmock.NewResult(0, 0))
	}
	mock.ExpectQuery("SELECT active_branch()").WillReturnRows(sqlmock.NewRows([]string{"branch"}).AddRow("main"))
	return s, mock
}

func transactionPolicyConnection(t *testing.T) (*Store, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		mock.ExpectClose()
		if err := db.Close(); err != nil {
			t.Error(err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})
	return &Store{db: db}, mock
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

func TestEmbeddedCommitErrorClassification(t *testing.T) {
	for _, tc := range []struct {
		name, backend string
		input         error
		conflict      bool
	}{
		{"embedded-state-omitted", "embedded", &mysql.MySQLError{Number: 1213}, true},
		{"wrapped-embedded-state-omitted", "embedded", errors.Join(errors.New("retained wrapper"), &mysql.MySQLError{Number: 1213}), true},
		{"embedded-confirmed-state", "embedded", &mysql.MySQLError{Number: 1213, SQLState: [5]byte{'4', '0', '0', '0', '1'}}, true},
		{"server-confirmed-state", "server", &mysql.MySQLError{Number: 1213, SQLState: [5]byte{'4', '0', '0', '0', '1'}}, true},
		{"server-state-omitted", "server", &mysql.MySQLError{Number: 1213}, false},
		{"default-state-omitted", "", &mysql.MySQLError{Number: 1213}, false},
		{"unknown-backend-state-omitted", "unrecognized", &mysql.MySQLError{Number: 1213}, false},
		{"embedded-wrong-state", "embedded", &mysql.MySQLError{Number: 1213, SQLState: [5]byte{'H', 'Y', '0', '0', '0'}}, false},
		{"server-wrong-state", "server", &mysql.MySQLError{Number: 1213, SQLState: [5]byte{'H', 'Y', '0', '0', '0'}}, false},
		{"lock-timeout", "embedded", &mysql.MySQLError{Number: 1205}, false},
		{"same-message-wrong-number", "embedded", &mysql.MySQLError{Number: 1105, Message: "Error 1213 serialization failure"}, false},
		{"text-is-not-typed", "embedded", errors.New("Error 1213 serialization failure 40001"), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := classifyCommitErrorForBackend(tc.input, tc.backend)
			if !errors.Is(err, tc.input) || errors.Is(err, ErrConflict) != tc.conflict || errors.Is(err, ErrOutcomeUnknown) == tc.conflict {
				t.Fatalf("wrong backend/cause classification: %v", err)
			}
		})
	}
	if err := classifyCommitErrorForBackend(nil, "embedded"); err != nil {
		t.Fatal(err)
	}
}
