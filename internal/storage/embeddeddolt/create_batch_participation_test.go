//go:build cgo

package embeddeddolt_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/storage/createbatchequiv"
	"github.com/steveyegge/beads/internal/storage/issueops"
	"github.com/steveyegge/beads/internal/types"
)

// The participation_generation write fence (design §16.2b) lives at the
// version-mint seam: RecordVersionForCreateInTx stamps a brand-new row's
// participation_generation from store_epoch.epoch in the same mint that
// writes its first issue_versions row, while RecordVersionInTx is
// update-shaped and mints nothing at all for a row whose generation is NULL.
// A batch create defers every mint to one final pass in
// CreateIssuesInTxWithContext, so that pass must still tell the two apart:
// a brand-new row that reached it through the update-shaped entry point would
// silently get no stamp, no version row and no current_revision. The
// create-batch equivalence harness (createbatchequiv) cannot see that, because
// it never reads participation_generation; these tests do.
//
// DisableCreateFastPathsForTest is process-global, so nothing here runs in
// parallel and its calls never overlap.

const participationActor = "importer"

var participationAt = time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

func participationIssue(suffix, title string) *types.Issue {
	return &types.Issue{
		ID: createbatchequiv.Prefix + "-" + suffix, Title: title, Status: types.StatusOpen, Priority: 2,
		IssueType: types.TypeTask, CreatedAt: participationAt, UpdatedAt: participationAt,
	}
}

// brandNewParticipationIssues returns n issues with explicit ids that no
// store holds yet, so the batch's create cache defers every one of them to
// its multi-row INSERT.
func brandNewParticipationIssues(n int) []*types.Issue {
	out := make([]*types.Issue, 0, n)
	for i := 1; i <= n; i++ {
		out = append(out, participationIssue(fmt.Sprintf("pg%03d", i), fmt.Sprintf("brand new %d", i)))
	}
	return out
}

func participationIDs(issues []*types.Issue) []string {
	ids := make([]string, 0, len(issues))
	for _, issue := range issues {
		ids = append(ids, issue.ID)
	}
	return ids
}

// createParticipationBatch runs one CreateIssuesInTxWithResult call in its own
// transaction, with versioned history (and the events journal) on or off for
// that transaction, the way the equivalence harness scopes them.
func createParticipationBatch(t *testing.T, db *sql.DB, issues []*types.Issue, history bool) {
	t.Helper()
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	clearJournal := issueops.ScopeEventsJournalTransaction(tx, history)
	clearVersions := issueops.ScopeVersionedHistoryTransaction(tx, history)
	_, err = issueops.CreateIssuesInTxWithResult(ctx, tx, issues, participationActor,
		storage.BatchCreateOptions{SkipPrefixValidation: true})
	clearVersions()
	clearJournal()
	if err != nil {
		_ = tx.Rollback()
		t.Fatalf("CreateIssuesInTxWithResult (history=%v): %v", history, err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}
}

// participationGenerations reads every issues row's participation_generation.
func participationGenerations(t *testing.T, db *sql.DB) map[string]sql.NullInt64 {
	t.Helper()
	rows, err := db.Query("SELECT id, participation_generation FROM issues")
	if err != nil {
		t.Fatalf("read participation_generation: %v", err)
	}
	defer rows.Close()
	out := map[string]sql.NullInt64{}
	for rows.Next() {
		var id string
		var gen sql.NullInt64
		if err := rows.Scan(&id, &gen); err != nil {
			t.Fatalf("read participation_generation: %v", err)
		}
		out[id] = gen
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("read participation_generation: %v", err)
	}
	return out
}

// participationStoreEpoch reads store_epoch.epoch; ok is false when the row
// has not been seeded yet (no version was ever minted).
func participationStoreEpoch(t *testing.T, db *sql.DB) (epoch int64, ok bool) {
	t.Helper()
	err := db.QueryRow("SELECT epoch FROM store_epoch WHERE id = 1").Scan(&epoch)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false
	}
	if err != nil {
		t.Fatalf("read store_epoch: %v", err)
	}
	return epoch, true
}

// participationVersionRows lists issue_versions as "issue_id revision actor",
// sorted, the columns the equivalence harness compares.
func participationVersionRows(t *testing.T, db *sql.DB) []string {
	t.Helper()
	rows, err := db.Query("SELECT issue_id, revision, change_actor FROM issue_versions")
	if err != nil {
		t.Fatalf("read issue_versions: %v", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var issueID string
		var revision int64
		var actor sql.NullString
		if err := rows.Scan(&issueID, &revision, &actor); err != nil {
			t.Fatalf("read issue_versions: %v", err)
		}
		out = append(out, fmt.Sprintf("%s %d %s", issueID, revision, actor.String))
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("read issue_versions: %v", err)
	}
	slices.Sort(out)
	return out
}

// firstVersionRows is what one create-shaped mint per id writes.
func firstVersionRows(ids []string) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, fmt.Sprintf("%s 1 %s", id, participationActor))
	}
	slices.Sort(out)
	return out
}

// requireStamped fails unless every id carries the store's epoch.
func requireStamped(t *testing.T, what string, gens map[string]sql.NullInt64, epoch int64, ids []string) {
	t.Helper()
	var bad []string
	for _, id := range ids {
		if gen, ok := gens[id]; !ok || !gen.Valid || gen.Int64 != epoch {
			bad = append(bad, fmt.Sprintf("%s=%v", id, gen))
		}
	}
	if len(bad) > 0 {
		t.Errorf("%s: %d of %d brand-new rows lack participation_generation = store_epoch.epoch (%d): %v",
			what, len(bad), len(ids), epoch, bad[:min(len(bad), 10)])
	}
}

// (a) A batch of brand-new explicit-id issues, large enough that its rows go
// out in more than one multi-row INSERT (issueInsertRowsPerStatement is 100):
// every row is stamped with store_epoch.epoch and gets exactly one version
// row, and the batch stores the same issue_versions rows as the same input run
// through the per-row bodies (DisableCreateFastPathsForTest) in a fresh store.
func TestCreateBatchStampsParticipationGenerationOnBrandNewRows(t *testing.T) {
	skipUnlessEmbeddedDolt(t)
	const n = 150
	ids := participationIDs(brandNewParticipationIssues(n))
	want := firstVersionRows(ids)

	run := func(t *testing.T, perRow bool) []string {
		t.Helper()
		db := openEquivalenceDB(t)
		if perRow {
			restore := issueops.DisableCreateFastPathsForTest()
			defer restore()
		}
		createParticipationBatch(t, db, brandNewParticipationIssues(n), true)
		epoch, ok := participationStoreEpoch(t, db)
		if !ok {
			t.Fatal("store_epoch has no row after a versioned batch: nothing was minted")
		}
		requireStamped(t, fmt.Sprintf("perRow=%v", perRow), participationGenerations(t, db), epoch, ids)
		return participationVersionRows(t, db)
	}

	fast := run(t, false)
	if !slices.Equal(fast, want) {
		t.Errorf("batch fast path stored %d version rows, want one first version per brand-new issue (%d)\n got: %v",
			len(fast), len(want), fast[:min(len(fast), 10)])
	}
	perRow := run(t, true)
	if !slices.Equal(fast, perRow) {
		t.Errorf("issue_versions differs between the batch fast path and the per-row bodies:\nfast:    %v\nper-row: %v",
			fast[:min(len(fast), 10)], perRow[:min(len(perRow), 10)])
	}
}

// (b) A batch row that resolves to an EXISTING legacy row (created while
// versioned history was off, so its generation is NULL) is an update-shaped
// write: it stays NULL and gets no version row (a skip, not a refusal), while
// its brand-new siblings in the same batch are stamped and versioned.
func TestCreateBatchLeavesExistingLegacyRowUnstamped(t *testing.T) {
	skipUnlessEmbeddedDolt(t)
	db := openEquivalenceDB(t)
	legacy := participationIssue("legacy", "legacy row")
	createParticipationBatch(t, db, []*types.Issue{legacy}, false)
	if gen := participationGenerations(t, db)[legacy.ID]; gen.Valid {
		t.Fatalf("precondition: a row created with history off must be legacy (NULL), got %v", gen)
	}

	siblings := brandNewParticipationIssues(3)
	batch := []*types.Issue{
		siblings[0],
		participationIssue("legacy", "legacy row, updated by the batch"),
		siblings[1],
		siblings[2],
	}
	createParticipationBatch(t, db, batch, true)

	epoch, ok := participationStoreEpoch(t, db)
	if !ok {
		t.Fatal("store_epoch has no row after a versioned batch: nothing was minted")
	}
	gens := participationGenerations(t, db)
	if gen := gens[legacy.ID]; gen.Valid {
		t.Errorf("existing legacy row %s: participation_generation = %d, want NULL (an update never promotes a legacy row)", legacy.ID, gen.Int64)
	}
	siblingIDs := participationIDs(siblings)
	requireStamped(t, "brand-new siblings of a legacy row", gens, epoch, siblingIDs)
	if got, want := participationVersionRows(t, db), firstVersionRows(siblingIDs); !slices.Equal(got, want) {
		t.Errorf("issue_versions = %v, want only the siblings' first versions %v (none for the legacy row)", got, want)
	}
	var title string
	if err := db.QueryRow("SELECT title FROM issues WHERE id = ?", legacy.ID).Scan(&title); err != nil {
		t.Fatal(err)
	}
	if title != "legacy row, updated by the batch" {
		t.Errorf("legacy row title = %q: the batch did not reach the existing row", title)
	}
}

// (c) With versioned history off, a batch of brand-new rows stamps nothing and
// mints nothing: every participation_generation stays NULL and issue_versions
// stays empty (FR-7: indistinguishable from a true legacy row).
func TestCreateBatchWithHistoryOffStampsNothing(t *testing.T) {
	skipUnlessEmbeddedDolt(t)
	db := openEquivalenceDB(t)
	issues := brandNewParticipationIssues(20)
	createParticipationBatch(t, db, issues, false)

	gens := participationGenerations(t, db)
	for _, id := range participationIDs(issues) {
		gen, ok := gens[id]
		if !ok {
			t.Fatalf("%s was not created", id)
		}
		if gen.Valid {
			t.Errorf("%s: participation_generation = %d with history off, want NULL", id, gen.Int64)
		}
	}
	if got := participationVersionRows(t, db); len(got) != 0 {
		t.Errorf("history off minted %d version rows, want none: %v", len(got), got[:min(len(got), 10)])
	}
}
