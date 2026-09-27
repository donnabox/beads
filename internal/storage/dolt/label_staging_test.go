package dolt

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/storage/testhelpers/labelstaging"
)

// An explicit graph test port opts into a caller-owned disposable server.
// With this opt-in, fresh database provisioning is serialized and never falls
// back to production. Without it, the existing test-container fixture may run
// tests in parallel on isolated branches of its already-provisioned database.
func doltLabelStagingFixture(t *testing.T) labelstaging.LabelStagingFixture {
	t.Helper()
	var store *DoltStore
	if text := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT"); text != "" {
		port, err := strconv.Atoi(text)
		if err != nil || port < 1024 || port > 65535 || port == 3306 || port == 3307 {
			t.Fatalf("invalid dedicated test port %q", text)
		}
		token := make([]byte, 12)
		if _, err := rand.Read(token); err != nil {
			t.Fatal(err)
		}
		database := "test_label_staging_" + hex.EncodeToString(token)
		ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
		defer cancel()
		admin, err := sql.Open("mysql", fmt.Sprintf("root@tcp(127.0.0.1:%d)/?parseTime=true", port))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if _, err := admin.ExecContext(ctx, "DROP DATABASE IF EXISTS `"+database+"`"); err != nil {
				t.Error(err)
			}
			if err := admin.Close(); err != nil {
				t.Error(err)
			}
		})
		store, err = New(ctx, &Config{Path: t.TempDir(), Database: database, ServerMode: true, ServerHost: "127.0.0.1", ServerPort: port, ServerUser: "root", CreateIfMissing: true, MaxOpenConns: 1, CommitterName: "label staging test", CommitterEmail: "label-staging@example.invalid"})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := store.Close(); err != nil {
				t.Error(err)
			}
		})
	} else {
		var cleanup func()
		store, cleanup = setupTestStore(t)
		t.Cleanup(cleanup)
	}
	if err := store.SetConfig(t.Context(), "issue_prefix", "labelstage"); err != nil {
		t.Fatal(err)
	}
	ops, err := NewIssueOperations(store)
	if err != nil {
		t.Fatal(err)
	}
	return labelstaging.LabelStagingFixture{
		IssuePrefix: "labelstage", Operations: ops, CreateIssue: store.CreateIssue, Commit: store.Commit,
		Exec: func(ctx context.Context, q string, args ...any) error {
			_, err := store.db.ExecContext(ctx, q, args...)
			return err
		},
		QueryScalar: func(ctx context.Context, q string, args []any, dest ...any) error {
			return store.db.QueryRowContext(ctx, q, args...).Scan(dest...)
		},
		GetIssue: store.GetIssue, Transaction: store.RunInTransaction, DirectAdd: store.AddLabel, DirectRemove: store.RemoveLabel,
	}
}

func TestLabelStagingCommittedState(t *testing.T) {
	for _, route := range []string{"lifecycle", "transaction", "direct"} {
		t.Run(route, func(t *testing.T) {
			labelstaging.RunLabelStagingMutations(t, t.Context(), doltLabelStagingFixture(t), route)
		})
	}
}
func TestLabelStagingNoopIsolation(t *testing.T) {
	for _, route := range []string{"lifecycle", "transaction", "direct"} {
		t.Run(route, func(t *testing.T) {
			labelstaging.RunLabelStagingNoops(t, t.Context(), doltLabelStagingFixture(t), route)
		})
	}
}
func TestLabelStagingPriorTransactionMutation(t *testing.T) {
	labelstaging.RunLabelStagingTransactionKeepsPriorDirty(t, t.Context(), doltLabelStagingFixture(t))
}
func TestLabelStagingWispIsolation(t *testing.T) {
	for _, route := range []string{"lifecycle", "transaction", "direct"} {
		t.Run(route, func(t *testing.T) {
			labelstaging.RunLabelStagingWisp(t, t.Context(), doltLabelStagingFixture(t), route)
		})
	}
}
func TestLabelOrdinaryCloseGuard(t *testing.T) {
	labelstaging.RunLabelOrdinaryCloseGuard(t, t.Context(), doltLabelStagingFixture(t))
}

func TestLabelStagingFailureIsolation(t *testing.T) {
	labelstaging.RunLabelStagingFailureIsolation(t, t.Context(), doltLabelStagingFixture(t))
}
