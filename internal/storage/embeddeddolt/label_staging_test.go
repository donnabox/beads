//go:build cgo

package embeddeddolt_test

import (
	"context"
	"testing"

	"github.com/steveyegge/beads/internal/storage/embeddeddolt"
	"github.com/steveyegge/beads/internal/storage/testhelpers/labelstaging"
)

func embeddedLabelStagingFixture(t *testing.T) labelstaging.LabelStagingFixture {
	t.Helper()
	te := newTestEnv(t, "labelstage")
	ops, err := embeddeddolt.NewIssueOperations(te.store)
	if err != nil {
		t.Fatal(err)
	}
	return labelstaging.LabelStagingFixture{
		IssuePrefix: "labelstage", Operations: ops, CreateIssue: te.store.CreateIssue, Commit: te.store.Commit,
		Exec: func(ctx context.Context, q string, args ...any) error {
			db, closeDB, err := embeddeddolt.OpenSQL(ctx, te.dataDir, te.database, "main")
			if err != nil {
				return err
			}
			defer closeDB()
			_, err = db.ExecContext(ctx, q, args...)
			return err
		},
		QueryScalar: func(ctx context.Context, q string, args []any, dest ...any) error {
			db, closeDB, err := embeddeddolt.OpenSQL(ctx, te.dataDir, te.database, "main")
			if err != nil {
				return err
			}
			defer closeDB()
			return db.QueryRowContext(ctx, q, args...).Scan(dest...)
		},
		GetIssue: te.store.GetIssue, Transaction: te.store.RunInTransaction,
	}
}

func TestLabelStagingCommittedState(t *testing.T) {
	skipUnlessEmbeddedDolt(t)
	for _, route := range []string{"lifecycle", "transaction"} {
		t.Run(route, func(t *testing.T) {
			f := embeddedLabelStagingFixture(t)
			labelstaging.RunLabelStagingMutations(t, t.Context(), f, route)
		})
	}
}
func TestLabelStagingNoopIsolation(t *testing.T) {
	skipUnlessEmbeddedDolt(t)
	for _, route := range []string{"lifecycle", "transaction"} {
		t.Run(route, func(t *testing.T) {
			labelstaging.RunLabelStagingNoops(t, t.Context(), embeddedLabelStagingFixture(t), route)
		})
	}
}
func TestLabelStagingPriorTransactionMutation(t *testing.T) {
	skipUnlessEmbeddedDolt(t)
	labelstaging.RunLabelStagingTransactionKeepsPriorDirty(t, t.Context(), embeddedLabelStagingFixture(t))
}
func TestLabelStagingWispIsolation(t *testing.T) {
	skipUnlessEmbeddedDolt(t)
	for _, route := range []string{"lifecycle", "transaction"} {
		t.Run(route, func(t *testing.T) {
			labelstaging.RunLabelStagingWisp(t, t.Context(), embeddedLabelStagingFixture(t), route)
		})
	}
}
func TestLabelOrdinaryCloseGuard(t *testing.T) {
	skipUnlessEmbeddedDolt(t)
	labelstaging.RunLabelOrdinaryCloseGuard(t, t.Context(), embeddedLabelStagingFixture(t))
}

func TestLabelStagingFailureIsolation(t *testing.T) {
	skipUnlessEmbeddedDolt(t)
	labelstaging.RunLabelStagingFailureIsolation(t, t.Context(), embeddedLabelStagingFixture(t))
}
