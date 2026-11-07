package dolt

import (
	"context"
	"errors"
	"testing"

	"github.com/steveyegge/beads/internal/types"
	publicops "github.com/steveyegge/beads/issueops"
)

// TestGuardHooksOnTheStoreLeg holds the store-backed bodies to what the unit-of-work
// leg's TestGuardComparedInsideTheWrite holds that leg to, short of the races, which
// bd serve meets on the other leg. The hooks are one contract with two bodies.
func TestGuardHooksOnTheStoreLeg(t *testing.T) {
	ctx := context.Background()
	store, cleanup := setupTestStore(t)
	defer cleanup()
	lifecycle, err := store.IssueLifecycle()
	if err != nil {
		t.Fatal(err)
	}
	deleter, err := store.Deleter()
	if err != nil {
		t.Fatal(err)
	}
	editor, err := store.DependencyEditor()
	if err != nil {
		t.Fatal(err)
	}
	seed := func(t *testing.T) *types.Issue {
		t.Helper()
		created, err := lifecycle.Create(ctx, publicops.CreateRequest{
			Actor: "seed", StoredTimestamps: true,
			Issue: &types.Issue{Title: "before", IssueType: types.TypeTask, Status: types.StatusOpen, Priority: 2},
		})
		if err != nil {
			t.Fatalf("seed: %v", err)
		}
		return created.Issue
	}
	read := func(t *testing.T, id string) *types.Issue {
		t.Helper()
		var row *types.Issue
		if _, err := lifecycle.Update(ctx, publicops.UpdateRequest{
			Actor: "reader", IssueID: id,
			InTx: func(current *types.Issue, _ *publicops.UpdateRequest) error { row = current; return nil },
		}); err != nil {
			t.Fatalf("read: %v", err)
		}
		if row == nil {
			t.Fatal("this leg ignored InTx")
		}
		return row
	}
	stale := errors.New("stale")

	t.Run("update", func(t *testing.T) {
		issue := seed(t)
		if _, err := lifecycle.Update(ctx, publicops.UpdateRequest{
			Actor: "x", IssueID: issue.ID, Patch: publicops.IssuePatch{Title: publicops.Field[string]{Set: true, Value: "no"}},
			InTx: func(*types.Issue, *publicops.UpdateRequest) error { return stale },
		}); !errors.Is(err, stale) {
			t.Fatalf("refusal = %v", err)
		}
		before := read(t, issue.ID)
		if before.Title != "before" || before.RowVersion != issue.RowVersion || !before.CreatedAt.Equal(issue.CreatedAt) {
			t.Fatalf("after a refusal, and against the stored create: %+v vs %+v", before, issue)
		}
		noop, err := lifecycle.Update(ctx, publicops.UpdateRequest{
			Actor: "x", IssueID: issue.ID,
			InTx: func(*types.Issue, *publicops.UpdateRequest) error { return nil },
		})
		if err != nil || noop.Changed || noop.Issue.RowVersion != before.RowVersion {
			t.Fatalf("no-op = %+v, %v", noop, err)
		}
		labeled, err := lifecycle.Update(ctx, publicops.UpdateRequest{
			Actor: "x", IssueID: issue.ID,
			InTx: func(current *types.Issue, attempt *publicops.UpdateRequest) error {
				attempt.Patch = publicops.IssuePatch{Labels: publicops.LabelPatch{Add: []string{current.Title}}}
				return nil
			},
		})
		if err != nil || !labeled.Changed || len(labeled.Issue.Labels) != 1 || labeled.Issue.Labels[0] != "before" {
			t.Fatalf("derived label = %+v, %v", labeled, err)
		}
		if labeled.Issue.RowVersion == before.RowVersion {
			t.Error("a guarded label-only write left row_lock alone")
		}
	})

	t.Run("close and reopen", func(t *testing.T) {
		issue := seed(t)
		refuse := &publicops.RowCheck{Run: func(*types.Issue) error { return stale }}
		pass := &publicops.RowCheck{Run: func(*types.Issue) error { return nil }}
		if _, err := lifecycle.Close(ctx, publicops.CloseRequest{Actor: "x", IssueID: issue.ID, InTx: refuse}); !errors.Is(err, stale) {
			t.Fatalf("close refusal = %v", err)
		}
		closed, err := lifecycle.Close(ctx, publicops.CloseRequest{Actor: "x", IssueID: issue.ID, InTx: pass})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := lifecycle.Reopen(ctx, publicops.ReopenRequest{Actor: "x", IssueID: issue.ID, InTx: refuse}); !errors.Is(err, stale) {
			t.Fatalf("reopen refusal = %v", err)
		}
		reopened, err := lifecycle.Reopen(ctx, publicops.ReopenRequest{Actor: "x", IssueID: issue.ID, InTx: pass})
		if err != nil {
			t.Fatal(err)
		}
		if reopened.Issue.RowVersion == closed.Issue.RowVersion && read(t, issue.ID).RowVersion == closed.Issue.RowVersion {
			t.Error("a guarded reopen left row_lock alone")
		}
	})

	t.Run("incident edges", func(t *testing.T) {
		a, b := seed(t), seed(t)
		if _, err := editor.AddDependencies(ctx, publicops.AddDependenciesRequest{
			Actor: "linker",
			Edges: []publicops.DependencyEdge{{IssueID: a.ID, DependsOnID: b.ID, Type: types.DepBlocks}},
		}); err != nil {
			t.Fatal(err)
		}
		for _, id := range []string{a.ID, b.ID} {
			var incident *publicops.IncidentEdgesError
			if _, err := deleter.Delete(ctx, publicops.DeleteRequest{Actor: "x", IDs: []string{id}, RefuseIncidentEdges: true, Force: true}); !errors.As(err, &incident) {
				t.Errorf("delete %s = %v", id, err)
			}
		}
	})

	t.Run("delete", func(t *testing.T) {
		issue := seed(t)
		if _, err := deleter.Delete(ctx, publicops.DeleteRequest{
			Actor: "x", IDs: []string{issue.ID},
			InTx: func(*types.Issue) error { return stale },
		}); !errors.Is(err, stale) {
			t.Fatalf("refusal = %v", err)
		}
		if _, err := deleter.Delete(ctx, publicops.DeleteRequest{
			Actor: "x", IDs: []string{issue.ID}, RefuseIncidentEdges: true,
			InTx: func(*types.Issue) error { return nil },
		}); err != nil {
			t.Fatalf("delete = %v", err)
		}
		if _, err := deleter.Delete(ctx, publicops.DeleteRequest{
			Actor: "x", IDs: []string{"a", "b"},
			InTx: func(*types.Issue) error { return nil },
		}); !errors.Is(err, publicops.ErrValidation) {
			t.Fatalf("two ids = %v", err)
		}
	})
}
