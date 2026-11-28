//go:build cgo

package doctor

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/storage/embeddeddolt"
	storageissueops "github.com/steveyegge/beads/internal/storage/issueops"
	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/internal/versionedhistory"
)

// mintInRolledBackTx runs the mint on the stored row id and returns what it answered.
// It runs RecordVersionAtInTx, which is recordVersionAtInTx without the switch and
// without the write fence, so it refuses a row for exactly the reason the production
// mint would and does not depend on versioned history being on. The transaction is
// rolled back, so the store is not changed by asking.
func mintInRolledBackTx(ctx context.Context, db *sql.DB, id string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	return storageissueops.RecordVersionAtInTx(ctx, tx, id, "doctor-test", time.Now().UTC())
}

// The check is run over real rows, read back through the store's own reader, beside
// the real mint run over the same rows: the ids the check names must be the ids the
// mint refuses, and no others. That is the agreement the pre-flight rests on, taken on
// a store and not on constructed values, and it holds the brief projection the check
// reads through to carrying what can be refused (a gate's timeout is in a column, not
// in metadata). Every row is written with versioned history off, which is the state a
// store is in when someone wants to know whether it is safe to turn it on; the first
// assertion is that it was.
//
// Wisps and no-history rows live in their own table, are never versioned, and are the
// control: they hold the same oversize values, and neither the check nor the mint may
// object.
func TestCheckVersionableIssues_EmbeddedStoreAgreesWithTheMint(t *testing.T) {
	if os.Getenv("BEADS_TEST_EMBEDDED_DOLT") != "1" {
		t.Skip("set BEADS_TEST_EMBEDDED_DOLT=1 to run embedded dolt tests")
	}

	ctx := t.Context()
	beadsDir := filepath.Join(t.TempDir(), ".beads")
	const database = "beads"

	store, err := embeddeddolt.Open(ctx, beadsDir, database, "main")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("close store: %v", err)
		}
	})
	if err := store.SetConfig(ctx, "issue_prefix", "vi"); err != nil {
		t.Fatalf("SetConfig issue_prefix: %v", err)
	}
	if err := store.Commit(ctx, "init"); err != nil {
		t.Fatalf("Commit: %v", err)
	}
	if on, err := store.GetConfig(ctx, versionedhistory.ConfigKey); err != nil || on != "" {
		t.Fatalf("versioned history is not off in a fresh store: %q, %v", on, err)
	}

	oversize := json.RawMessage(`{"ts":1727000000000000000}`)
	base := func(title string) *types.Issue {
		return &types.Issue{Title: title, Status: types.StatusOpen, Priority: 2, IssueType: types.TypeTask}
	}
	gate := func(title string, timeout time.Duration) *types.Issue {
		issue := base(title)
		issue.IssueType = types.TypeGate
		issue.AwaitType = "timer"
		issue.Timeout = timeout
		return issue
	}
	withMetadata := func(issue *types.Issue, metadata json.RawMessage) *types.Issue {
		issue.Metadata = metadata
		return issue
	}
	ephemeral := withMetadata(base("wisp with an oversize number"), oversize)
	ephemeral.Ephemeral = true
	noHistory := withMetadata(base("no-history row with an oversize number"), oversize)
	noHistory.NoHistory = true

	rows := []struct {
		label   string
		issue   *types.Issue
		refused bool
	}{
		{"metadata that is clean", withMetadata(base("clean"), json.RawMessage(`{"ok":1}`)), false},
		{"metadata number past the range", withMetadata(base("oversize number"), oversize), true},
		{"gate timeout 2500h", gate("gate 2500h", 2500*time.Hour), false},
		{"gate timeout 2600h", gate("gate 2600h", 2600*time.Hour), true},
		{"gate timeout 8760h", gate("gate 8760h", 8760*time.Hour), true},
		{"gate timeout -2600h", gate("gate minus 2600h", -2600*time.Hour), true},
		{"gate timeout of exactly 2^53-1 ns", gate("gate at the bound", time.Duration(maxExactNanoseconds)), false},
		{"gate timeout of exactly 2^53 ns", gate("gate past the bound", time.Duration(maxExactNanoseconds+1)), true},
		{"wisp holding an oversize number", ephemeral, false},
		{"no-history row holding an oversize number", noHistory, false},
	}
	for _, r := range rows {
		if err := store.CreateIssue(ctx, r.issue, "tester"); err != nil {
			t.Fatalf("create %q: %v", r.label, err)
		}
		if r.issue.ID == "" {
			t.Fatalf("create %q: the store assigned no id", r.label)
		}
	}

	reader, err := store.IssueReader()
	if err != nil {
		t.Fatalf("IssueReader: %v", err)
	}
	check := checkVersionableIssues(ctx, reader)
	if check.Status == StatusOK {
		t.Fatalf("a store holding a 2600h gate and an oversize number read clean: %s", check.Message)
	}
	checkNames := map[string]bool{}
	for _, id := range offenderIDs(check.Detail) {
		checkNames[id] = true
	}

	db, cleanup, err := embeddeddolt.OpenSQL(ctx, filepath.Join(beadsDir, "embeddeddolt"), database, "main")
	if err != nil {
		t.Fatalf("OpenSQL: %v", err)
	}
	t.Cleanup(func() {
		if err := cleanup(); err != nil {
			t.Errorf("close sql: %v", err)
		}
	})

	for _, r := range rows {
		t.Run(r.label, func(t *testing.T) {
			id := r.issue.ID
			mintErr := mintInRolledBackTx(ctx, db, id)
			mintRefused := false
			switch {
			case mintErr == nil:
			case errors.Is(mintErr, storageissueops.ErrIntegerNotRepresentable):
				mintRefused = true
			default:
				t.Fatalf("the mint failed on %s for a reason that is not a refusal: %v", id, mintErr)
			}

			if mintRefused != r.refused {
				t.Errorf("the mint refused=%t on %s, the corpus says refused=%t (mint answered %v)", mintRefused, id, r.refused, mintErr)
			}
			if checkNames[id] != mintRefused {
				t.Errorf("the check names %s = %t but the mint refuses it = %t: the pre-flight and the mint disagree", id, checkNames[id], mintRefused)
			}
		})
	}
}
