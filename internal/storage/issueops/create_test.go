package issueops

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/storage/depid"
	"github.com/steveyegge/beads/internal/types"
)

func TestValidateCreateIssuesMixedBucketDependenciesRejectsCrossBucketEdges(t *testing.T) {
	regularA := &types.Issue{ID: "test-regular-a", IssueType: types.TypeTask}
	regularB := &types.Issue{ID: "test-regular-b", IssueType: types.TypeTask}
	wispA := &types.Issue{ID: "test-wisp-a", IssueType: types.TypeTask, Ephemeral: true}
	wispB := &types.Issue{ID: "test-wisp-b", IssueType: types.TypeTask, Ephemeral: true}

	tests := []struct {
		name      string
		issues    []*types.Issue
		wantError bool
	}{
		{
			name: "regular to wisp",
			issues: []*types.Issue{
				{
					ID:        regularA.ID,
					IssueType: types.TypeTask,
					Dependencies: []*types.Dependency{{
						DependsOnID: wispA.ID,
						Type:        types.DepBlocks,
					}},
				},
				wispA,
			},
			wantError: true,
		},
		{
			name: "wisp to regular",
			issues: []*types.Issue{
				regularA,
				{
					ID:        wispA.ID,
					IssueType: types.TypeTask,
					Ephemeral: true,
					Dependencies: []*types.Dependency{{
						DependsOnID: regularA.ID,
						Type:        types.DepBlocks,
					}},
				},
			},
			wantError: true,
		},
		{
			name: "same bucket dependencies",
			issues: []*types.Issue{
				regularB,
				{
					ID:        regularA.ID,
					IssueType: types.TypeTask,
					Dependencies: []*types.Dependency{{
						DependsOnID: regularB.ID,
						Type:        types.DepBlocks,
					}},
				},
				wispB,
				{
					ID:        wispA.ID,
					IssueType: types.TypeTask,
					Ephemeral: true,
					Dependencies: []*types.Dependency{{
						DependsOnID: wispB.ID,
						Type:        types.DepBlocks,
					}},
				},
			},
		},
		{
			name: "out of batch target",
			issues: []*types.Issue{
				{
					ID:        regularA.ID,
					IssueType: types.TypeTask,
					Dependencies: []*types.Dependency{{
						DependsOnID: "test-external-wisp",
						Type:        types.DepBlocks,
					}},
				},
				wispA,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCreateIssuesMixedBucketDependencies(tt.issues)
			if tt.wantError {
				if err == nil || !strings.Contains(err.Error(), "cross-bucket dependency") {
					t.Fatalf("error = %v, want cross-bucket dependency", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("error = %v, want nil", err)
			}
		})
	}
}

func TestFilterCreateIssuesMixedBucketDependenciesSkipsWhenConfigured(t *testing.T) {
	regular := &types.Issue{
		ID:        "test-regular-source",
		IssueType: types.TypeTask,
		Dependencies: []*types.Dependency{{
			DependsOnID: "test-wisp-target",
			Type:        types.DepBlocks,
		}},
	}
	wisp := &types.Issue{
		ID:        "test-wisp-target",
		IssueType: types.TypeTask,
		Ephemeral: true,
	}
	var skipped []string

	filtered, err := filterCreateIssuesMixedBucketDependencies([]*types.Issue{regular, wisp}, storage.BatchCreateOptions{
		SkipDependencyValidationErrors: true,
		OnSkippedDependency: func(issueID, dependsOnID, reason string) {
			skipped = append(skipped, issueID+" -> "+dependsOnID+": "+reason)
		},
	})
	if err != nil {
		t.Fatalf("filterCreateIssuesMixedBucketDependencies error = %v, want nil", err)
	}
	if len(filtered) != 2 {
		t.Fatalf("len(filtered) = %d, want 2", len(filtered))
	}
	if len(filtered[0].Dependencies) != 0 {
		t.Fatalf("filtered[0].Dependencies = %#v, want none", filtered[0].Dependencies)
	}
	if len(regular.Dependencies) != 1 {
		t.Fatalf("regular.Dependencies was mutated to %#v, want original dependency preserved", regular.Dependencies)
	}
	if len(skipped) != 1 || !strings.Contains(skipped[0], "test-regular-source -> test-wisp-target") ||
		!strings.Contains(skipped[0], "cross-bucket dependency") {
		t.Fatalf("skipped = %#v, want cross-bucket dependency detail", skipped)
	}
}

func TestPersistDependenciesHonorsImportedCreatedBy(t *testing.T) {
	ctx := context.Background()
	db, mock, tx := beginMockTx(t)
	defer db.Close()

	target := &types.Issue{ID: "target", IssueType: types.TypeTask}
	source := &types.Issue{
		ID:        "source",
		IssueType: types.TypeTask,
		Dependencies: []*types.Dependency{{
			DependsOnID: "target",
			Type:        types.DepRelated,
			CreatedBy:   "someone.else",
		}},
	}

	mock.ExpectQuery("SELECT 1 FROM wisps WHERE id = \\? LIMIT 1").
		WithArgs("target").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("SELECT 1 FROM issues WHERE id = \\?").
		WithArgs("target").
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectExec("INSERT INTO dependencies").
		WithArgs(depid.New("source", "target"), "source", "target", types.DepRelated, "someone.else", sqlmock.AnyArg(), "{}", "").
		WillReturnResult(sqlmock.NewResult(0, 1))

	result, err := PersistDependenciesWithOptionsResult(ctx, tx, []*types.Issue{target, source}, "current.user", storage.BatchCreateOptions{})
	if err != nil {
		t.Fatalf("PersistDependenciesWithOptionsResult error = %v, want nil", err)
	}
	if !result.ChangedTables["dependencies"] {
		t.Fatalf("ChangedTables = %#v, want dependencies changed", result.ChangedTables)
	}

	mock.ExpectRollback()
	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPersistDependenciesDefaultsCreatedByToActor(t *testing.T) {
	ctx := context.Background()
	db, mock, tx := beginMockTx(t)
	defer db.Close()

	target := &types.Issue{ID: "target", IssueType: types.TypeTask}
	source := &types.Issue{
		ID:        "source",
		IssueType: types.TypeTask,
		Dependencies: []*types.Dependency{{
			DependsOnID: "target",
			Type:        types.DepRelated,
		}},
	}

	mock.ExpectQuery("SELECT 1 FROM wisps WHERE id = \\? LIMIT 1").
		WithArgs("target").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("SELECT 1 FROM issues WHERE id = \\?").
		WithArgs("target").
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectExec("INSERT INTO dependencies").
		WithArgs(depid.New("source", "target"), "source", "target", types.DepRelated, "current.user", sqlmock.AnyArg(), "{}", "").
		WillReturnResult(sqlmock.NewResult(0, 1))

	_, err := PersistDependenciesWithOptionsResult(ctx, tx, []*types.Issue{target, source}, "current.user", storage.BatchCreateOptions{})
	if err != nil {
		t.Fatalf("PersistDependenciesWithOptionsResult error = %v, want nil", err)
	}

	mock.ExpectRollback()
	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPersistDependenciesClassifiesBareCrossPrefixTargetAsExternal(t *testing.T) {
	ctx := context.Background()
	db, mock, tx := beginMockTx(t)
	defer db.Close()

	source := &types.Issue{
		ID:        "sym-3su",
		IssueType: types.TypeTask,
		Dependencies: []*types.Dependency{{
			DependsOnID: "mkt-456",
			Type:        types.DepRelated,
		}},
	}
	var skipped []string

	// A bare target with a different issue prefix is external. In particular,
	// persistence must not probe either local target table before this insert.
	mock.ExpectExec("INSERT INTO dependencies \\(id, issue_id, depends_on_external").
		WithArgs(depid.New("sym-3su", "mkt-456"), "sym-3su", "mkt-456", types.DepRelated, "tester", sqlmock.AnyArg(), "{}", "").
		WillReturnResult(sqlmock.NewResult(0, 1))

	result, err := PersistDependenciesWithOptionsResult(ctx, tx, []*types.Issue{source}, "tester", storage.BatchCreateOptions{
		OnSkippedDependency: func(issueID, dependsOnID, reason string) {
			skipped = append(skipped, issueID+" -> "+dependsOnID+": "+reason)
		},
	})
	if err != nil {
		t.Fatalf("PersistDependenciesWithOptionsResult error = %v, want nil", err)
	}
	if len(skipped) != 0 {
		t.Fatalf("skipped = %#v, want none", skipped)
	}
	if !result.ChangedTables["dependencies"] {
		t.Fatalf("ChangedTables = %#v, want dependencies changed", result.ChangedTables)
	}

	mock.ExpectRollback()
	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPersistDependenciesReturnsTargetLookupErrors(t *testing.T) {
	ctx := context.Background()
	db, mock, tx := beginMockTx(t)
	defer db.Close()
	targetErr := errors.New("target lookup failed")
	issue := &types.Issue{
		ID:        "source",
		IssueType: types.TypeTask,
		Dependencies: []*types.Dependency{{
			DependsOnID: "target",
			Type:        types.DepBlocks,
		}},
	}

	mock.ExpectQuery("SELECT 1 FROM wisps WHERE id = \\? LIMIT 1").
		WithArgs("target").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("SELECT 1 FROM issues WHERE id = \\?").
		WithArgs("target").
		WillReturnError(targetErr)

	err := PersistDependencies(ctx, tx, []*types.Issue{issue}, "tester")
	if err == nil || !strings.Contains(err.Error(), "failed to check dependency target target for source") {
		t.Fatalf("error = %v, want dependency target lookup error", err)
	}

	mock.ExpectRollback()
	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPersistDependenciesSkipsValidationErrorsWhenConfigured(t *testing.T) {
	ctx := context.Background()
	db, mock, tx := beginMockTx(t)
	defer db.Close()
	issue := &types.Issue{
		ID:        "source",
		IssueType: types.TypeTask,
		Dependencies: []*types.Dependency{{
			DependsOnID: "source",
			Type:        types.DepBlocks,
		}},
	}
	var skipped []string

	mock.ExpectQuery("SELECT 1 FROM wisps WHERE id = \\? LIMIT 1").
		WithArgs("source").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("SELECT 1 FROM issues WHERE id = \\?").
		WithArgs("source").
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	result, err := PersistDependenciesWithOptionsResult(ctx, tx, []*types.Issue{issue}, "tester", storage.BatchCreateOptions{
		SkipDependencyValidationErrors: true,
		OnSkippedDependency: func(issueID, dependsOnID, reason string) {
			skipped = append(skipped, issueID+" -> "+dependsOnID+": "+reason)
		},
	})
	if err != nil {
		t.Fatalf("PersistDependenciesWithOptionsResult error = %v, want nil", err)
	}
	if len(result.ChangedTables) != 0 {
		t.Fatalf("ChangedTables = %#v, want none", result.ChangedTables)
	}
	if len(skipped) != 1 || !strings.Contains(skipped[0], "source -> source") ||
		!strings.Contains(skipped[0], "cannot depend on itself") {
		t.Fatalf("skipped = %#v, want self-dependency detail", skipped)
	}

	mock.ExpectRollback()
	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPersistDependenciesRejectsHierarchyBlocking(t *testing.T) {
	ctx := context.Background()
	db, mock, tx := beginMockTx(t)
	defer db.Close()
	issue := &types.Issue{
		ID:        "child",
		IssueType: types.TypeTask,
		Dependencies: []*types.Dependency{{
			DependsOnID: "parent",
			Type:        types.DepConditionalBlocks,
		}},
	}

	mock.ExpectQuery("SELECT 1 FROM wisps WHERE id = \\? LIMIT 1").
		WithArgs("parent").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("SELECT 1 FROM issues WHERE id = \\?").
		WithArgs("parent").
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectQuery("WITH RECURSIVE ancestors").
		WithArgs("child", "parent").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	_, err := PersistDependenciesWithOptionsResult(ctx, tx, []*types.Issue{issue}, "tester", storage.BatchCreateOptions{})
	if err == nil || !strings.Contains(err.Error(), "cannot be blocked by its ancestor") {
		t.Fatalf("error = %v, want ancestor hierarchy rejection", err)
	}

	mock.ExpectRollback()
	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPersistDependenciesValidatesPlannedHierarchyBeforeBlocking(t *testing.T) {
	ctx := context.Background()
	db, mock, tx := beginMockTx(t)
	defer db.Close()
	child := &types.Issue{
		ID:        "bd-child",
		IssueType: types.TypeTask,
		Dependencies: []*types.Dependency{
			{DependsOnID: "bd-grand", Type: types.DepBlocks}, // Deliberately first.
			{DependsOnID: "bd-parent", Type: types.DepParentChild},
		},
	}
	parent := &types.Issue{
		ID:        "bd-parent",
		IssueType: types.TypeTask,
		Dependencies: []*types.Dependency{{
			DependsOnID: "bd-grand",
			Type:        types.DepParentChild,
		}},
	}

	for _, pair := range [][2]string{{"bd-child", "bd-parent"}, {"bd-parent", "bd-grand"}} {
		mock.ExpectQuery("SELECT 1 FROM wisps WHERE id = \\? LIMIT 1").
			WithArgs(pair[1]).
			WillReturnError(sql.ErrNoRows)
		mock.ExpectQuery("SELECT 1 FROM issues WHERE id = \\?").
			WithArgs(pair[1]).
			WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
		mock.ExpectQuery("WITH RECURSIVE reachable").
			WithArgs(pair[1], pair[0]).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		mock.ExpectExec("INSERT INTO dependencies").
			WithArgs(depid.New(pair[0], pair[1]), pair[0], pair[1], types.DepParentChild, "tester", sqlmock.AnyArg(), "{}", "").
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec("REPLACE INTO local_metadata").
			WithArgs(dependencyCoordinationKey(pair[1], dependencyCoordinationDurableTier), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}
	mock.ExpectQuery("SELECT 1 FROM wisps WHERE id = \\? LIMIT 1").
		WithArgs("bd-grand").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("SELECT 1 FROM issues WHERE id = \\?").
		WithArgs("bd-grand").
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectQuery("WITH RECURSIVE ancestors").
		WithArgs("bd-child", "bd-grand").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	_, err := PersistDependenciesWithOptionsResult(ctx, tx, []*types.Issue{child, parent}, "tester", storage.BatchCreateOptions{})
	if err == nil || !strings.Contains(err.Error(), "cannot be blocked by its ancestor") {
		t.Fatalf("error = %v, want planned-ancestor rejection", err)
	}

	mock.ExpectRollback()
	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPersistDependenciesSkipsHierarchyValidationAcrossPrefixes(t *testing.T) {
	ctx := context.Background()
	db, mock, tx := beginMockTx(t)
	defer db.Close()
	issue := &types.Issue{
		ID:        "aa-source",
		IssueType: types.TypeTask,
		Dependencies: []*types.Dependency{{
			DependsOnID: "bb-target",
			Type:        types.DepBlocks,
		}},
	}

	// No target or ancestors query: target existence and hierarchy cannot be
	// validated locally across rig prefixes.
	mock.ExpectQuery("WITH RECURSIVE reachable").
		WithArgs("bb-target", "aa-source").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec("INSERT INTO dependencies \\(id, issue_id, depends_on_external").
		WithArgs(depid.New("aa-source", "bb-target"), "aa-source", "bb-target", types.DepBlocks, "tester", sqlmock.AnyArg(), "{}", "").
		WillReturnResult(sqlmock.NewResult(0, 1))

	_, err := PersistDependenciesWithOptionsResult(ctx, tx, []*types.Issue{issue}, "tester", storage.BatchCreateOptions{})
	if err != nil {
		t.Fatalf("cross-prefix blocking dependency: %v", err)
	}

	mock.ExpectRollback()
	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// The mock half of the missing-parent skip: it pins the STATEMENTS (no counter
// read, no upsert). Its real-backend twin is the conformance case
// ReconcileSkipsMissingParentCounter (backend/conformance/portable.go), which
// pins what a live engine does with the same skip on both Dolt legs.
func TestReconcileChildCountersSkipsMissingParent(t *testing.T) {
	ctx := context.Background()
	db, mock, tx := beginMockTx(t)
	defer db.Close()

	mock.ExpectQuery("SELECT 1 FROM wisps LIMIT 1").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("SELECT 1 FROM issues WHERE id = \\?").
		WithArgs("test-deleted-parent").
		WillReturnError(sql.ErrNoRows)

	changed, err := ReconcileChildCounters(ctx, tx, []*types.Issue{{
		ID:        "test-deleted-parent.7",
		IssueType: types.TypeTask,
	}})
	if err != nil {
		t.Fatalf("ReconcileChildCounters error = %v, want nil", err)
	}
	if len(changed) != 0 {
		t.Fatalf("changed tables = %#v, want none", changed)
	}

	// No counter SELECT or upsert is expected after the missing-parent lookup.
	mock.ExpectRollback()
	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestReconcileChildCountersReturnsParentLookupError(t *testing.T) {
	ctx := context.Background()
	db, mock, tx := beginMockTx(t)
	defer db.Close()
	lookupErr := errors.New("parent lookup failed")

	mock.ExpectQuery("SELECT 1 FROM wisps LIMIT 1").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("SELECT 1 FROM issues WHERE id = \\?").
		WithArgs("test-parent").
		WillReturnError(lookupErr)

	_, err := ReconcileChildCounters(ctx, tx, []*types.Issue{{
		ID:        "test-parent.1",
		IssueType: types.TypeTask,
	}})
	if err == nil || !strings.Contains(err.Error(), "failed to check child counter parent test-parent") || !errors.Is(err, lookupErr) {
		t.Fatalf("error = %v, want contextual parent lookup error", err)
	}

	// A lookup failure must not be mistaken for an absent parent or reach the
	// counter table.
	mock.ExpectRollback()
	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestReconcileChildCountersReturnsWispLookupError(t *testing.T) {
	ctx := context.Background()
	db, mock, tx := beginMockTx(t)
	defer db.Close()
	lookupErr := errors.New("wisp lookup failed")

	mock.ExpectQuery("SELECT 1 FROM wisps LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectQuery("SELECT id FROM wisps WHERE id IN \\(\\?\\)").
		WithArgs("test-parent").
		WillReturnError(lookupErr)

	_, err := ReconcileChildCounters(ctx, tx, []*types.Issue{{
		ID:        "test-parent.1",
		IssueType: types.TypeTask,
	}})
	if err == nil || !strings.Contains(err.Error(), "failed to route child counter parents") || !errors.Is(err, lookupErr) {
		t.Fatalf("error = %v, want contextual wisp lookup error", err)
	}

	// A failed wisp lookup must stop routing before any issues or counter query.
	mock.ExpectRollback()
	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// TestSplitCrossPlaneBatchEdges pins the partition the proxied import relies on
// (wy-zdfs6r): the rows keep every edge the engine can write in one batch, the
// cross-plane in-batch edges come back as edge-only copies, and the deferred
// batches are SINGLE-PLANE so handing one back to the engine cannot re-trigger
// filterCreateIssuesMixedBucketDependencies.
func TestSplitCrossPlaneBatchEdges(t *testing.T) {
	mk := func(id string, wisp bool, deps ...*types.Dependency) *types.Issue {
		return &types.Issue{ID: id, IssueType: types.TypeTask, Ephemeral: wisp, Dependencies: deps}
	}
	dep := func(source, target string) *types.Dependency {
		return &types.Dependency{IssueID: source, DependsOnID: target, Type: types.DepBlocks}
	}

	t.Run("SinglePlaneBatchIsUntouched", func(t *testing.T) {
		issues := []*types.Issue{mk("r1", false, dep("r1", "r2")), mk("r2", false)}
		rows, planes := SplitCrossPlaneBatchEdges(issues)
		if len(planes) != 0 {
			t.Fatalf("deferredPlanes = %#v, want none: no plane boundary is crossed", planes)
		}
		if &rows[0] != &issues[0] {
			t.Fatalf("rows was copied for a batch with nothing to defer")
		}
	})

	t.Run("EdgeOutOfTheBatchIsWiredInline", func(t *testing.T) {
		// The target is NOT a row of this batch, so it is an ordinary
		// committed row the engine's existence check finds: deferring it would
		// be work for nothing.
		issues := []*types.Issue{mk("r1", false, dep("r1", "w-elsewhere")), mk("w1", true)}
		rows, planes := SplitCrossPlaneBatchEdges(issues)
		if len(planes) != 0 {
			t.Fatalf("deferredPlanes = %#v, want none for an out-of-batch target", planes)
		}
		if len(rows[0].Dependencies) != 1 {
			t.Fatalf("rows[0].Dependencies = %#v, want the edge kept inline", rows[0].Dependencies)
		}
	})

	t.Run("CrossPlaneInBatchEdgesDeferOntoTheirOwnPlanes", func(t *testing.T) {
		// r1 -> w1 crosses; r1 -> r2 does not and must stay inline. w1 -> r2
		// crosses the other way, so the two deferred sources sit on opposite
		// planes and MUST NOT share a batch: w1 is r1's deferred target, and a
		// mixed second pass would skip-report r1 -> w1 all over again.
		issues := []*types.Issue{
			mk("r1", false, dep("r1", "w1"), dep("r1", "r2")),
			mk("r2", false),
			mk("w1", true, dep("w1", "r2")),
		}
		rows, planes := SplitCrossPlaneBatchEdges(issues)

		if len(rows) != 3 {
			t.Fatalf("len(rows) = %d, want 3", len(rows))
		}
		if len(rows[0].Dependencies) != 1 || rows[0].Dependencies[0].DependsOnID != "r2" {
			t.Fatalf("rows[0].Dependencies = %#v, want only the same-plane r1 -> r2", rows[0].Dependencies)
		}
		if len(rows[2].Dependencies) != 0 {
			t.Fatalf("rows[2].Dependencies = %#v, want none", rows[2].Dependencies)
		}
		// The caller's own issues are never mutated.
		if len(issues[0].Dependencies) != 2 || len(issues[2].Dependencies) != 1 {
			t.Fatalf("caller's issues were mutated: %#v / %#v", issues[0].Dependencies, issues[2].Dependencies)
		}

		if len(planes) != 2 {
			t.Fatalf("deferredPlanes = %d batches, want 2 (one per plane)", len(planes))
		}
		for i, plane := range planes {
			wisp := IsWisp(plane[0])
			for _, row := range plane {
				if IsWisp(row) != wisp {
					t.Fatalf("deferredPlanes[%d] mixes planes: %#v", i, plane)
				}
			}
		}
		if len(planes[0]) != 1 || planes[0][0].ID != "r1" || IsWisp(planes[0][0]) {
			t.Fatalf("deferredPlanes[0] = %#v, want the regular source r1", planes[0])
		}
		if got := planes[0][0].Dependencies; len(got) != 1 || got[0].DependsOnID != "w1" {
			t.Fatalf("r1's deferred edges = %#v, want only r1 -> w1", got)
		}
		if len(planes[1]) != 1 || planes[1][0].ID != "w1" || !IsWisp(planes[1][0]) {
			t.Fatalf("deferredPlanes[1] = %#v, want the wisp source w1", planes[1])
		}
	})

	t.Run("EdgeOnlyCopiesCarryNoAuxData", func(t *testing.T) {
		// The row's labels and comments merge with the row in the first pass;
		// a second merge would re-run their inserts for nothing.
		source := mk("r1", false, dep("r1", "w1"))
		source.Labels = []string{"lane:test"}
		source.Comments = []*types.Comment{{ID: "c1", Text: "carried"}}
		_, planes := SplitCrossPlaneBatchEdges([]*types.Issue{source, mk("w1", true)})
		if len(planes) != 1 || len(planes[0]) != 1 {
			t.Fatalf("deferredPlanes = %#v, want one row on one plane", planes)
		}
		if planes[0][0].Labels != nil || planes[0][0].Comments != nil {
			t.Fatalf("edge-only copy carries aux data: labels=%#v comments=%#v", planes[0][0].Labels, planes[0][0].Comments)
		}
		if len(source.Labels) != 1 || len(source.Comments) != 1 {
			t.Fatalf("caller's row lost its aux data: %#v / %#v", source.Labels, source.Comments)
		}
	})
}
