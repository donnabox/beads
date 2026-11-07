package uow

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/steveyegge/beads/internal/types"
	publicops "github.com/steveyegge/beads/issueops"
)

// guardFingerprint is what a guard compares: everything a label-only write and a
// title-only write can each move, which the row lock alone does not cover.
func guardFingerprint(issue *types.Issue) string {
	return issue.Title + "\x00" + strings.Join(issue.Labels, ",")
}

// staleGuard refuses unless the row still reads as want.
func staleGuard(want string, current *types.Issue) error {
	if guardFingerprint(current) != want {
		return &publicops.VersionMismatchError{Expected: current.RowVersion, Current: current.RowVersion}
	}
	return nil
}

// TestGuardComparedInsideTheWrite proves the two halves of a guard that lives in
// the write's own transaction, on real Dolt, by landing the competing write at the
// one moment an outside comparison cannot see: after this request's guard passed
// and before its commit.
//
// Dolt has no row locks and merges concurrent commits cell by cell, so a comparison
// in the transaction is not enough by itself: a label write and a title write share
// no cell and would both land. The row_lock rewrite on every guarded write is what
// makes them collide, and the replay is what reports the mismatch.
func TestGuardComparedInsideTheWrite(t *testing.T) {
	ctx := context.Background()
	provider := newUOWRoleFixtureProvider(t, ctx, "grd")
	lifecycle, err := NewIssueOperations(provider)
	if err != nil {
		t.Fatal(err)
	}
	seed := func(t *testing.T, id string) string {
		t.Helper()
		issue := &types.Issue{ID: id, Title: "before", IssueType: types.TypeTask, Status: types.StatusOpen, Priority: 2}
		created, err := lifecycle.Create(ctx, publicops.CreateRequest{Actor: "seed", Issue: issue})
		if err != nil {
			t.Fatalf("seed: %v", err)
		}
		return guardFingerprint(created.Issue)
	}
	read := func(t *testing.T, id string) *types.Issue {
		t.Helper()
		var current *types.Issue
		_, err := lifecycle.Update(ctx, publicops.UpdateRequest{
			Actor: "reader", IssueID: id,
			InTx: func(row *types.Issue, _ *publicops.UpdateRequest) error { current = row; return nil },
		})
		if err != nil {
			t.Fatalf("read back: %v", err)
		}
		if current == nil {
			t.Fatal("the guard did not run: this leg ignored InTx")
		}
		return current
	}
	title := func(value string) publicops.IssuePatch {
		return publicops.IssuePatch{Title: publicops.Field[string]{Set: true, Value: value}}
	}
	label := func(value string) publicops.IssuePatch {
		return publicops.IssuePatch{Labels: publicops.LabelPatch{Add: []string{value}}}
	}

	// race runs first with a guard on want, landing second to completion between
	// first's guard and first's commit, on the first attempt only.
	race := func(t *testing.T, id, want string, first, second publicops.IssuePatch) (firstErr, secondErr error, firstGuards int32) {
		t.Helper()
		var attempts atomic.Int32
		_, firstErr = lifecycle.Update(ctx, publicops.UpdateRequest{
			Actor: "first", IssueID: id, Patch: first,
			InTx: func(current *types.Issue, _ *publicops.UpdateRequest) error {
				if err := staleGuard(want, current); err != nil {
					return err
				}
				if attempts.Add(1) == 1 {
					_, secondErr = lifecycle.Update(ctx, publicops.UpdateRequest{
						Actor: "second", IssueID: id, Patch: second,
						InTx: func(current *types.Issue, _ *publicops.UpdateRequest) error { return staleGuard(want, current) },
					})
				}
				return nil
			},
		})
		return firstErr, secondErr, attempts.Load()
	}

	t.Run("a title write and a label write on one token: exactly one lands", func(t *testing.T) {
		id := "grd-race1"
		want := seed(t, id)
		firstErr, secondErr, _ := race(t, id, want, title("after"), label("x"))
		if secondErr != nil {
			t.Fatalf("the write that landed first was refused: %v", secondErr)
		}
		if !errors.Is(firstErr, publicops.ErrVersionMismatch) {
			t.Fatalf("the stale write came back %v, want a version mismatch: both landed on one token", firstErr)
		}
		current := read(t, id)
		if current.Title != "before" || !strings.Contains(strings.Join(current.Labels, ","), "x") {
			t.Fatalf("row = %q %v, want only the label write", current.Title, current.Labels)
		}
	})

	t.Run("a label-only write is a guarded writer too", func(t *testing.T) {
		id := "grd-race2"
		want := seed(t, id)
		firstErr, secondErr, _ := race(t, id, want, label("x"), title("after"))
		if secondErr != nil {
			t.Fatalf("the write that landed first was refused: %v", secondErr)
		}
		if !errors.Is(firstErr, publicops.ErrVersionMismatch) {
			t.Fatalf("the stale label write came back %v, want a version mismatch: both landed on one token", firstErr)
		}
		current := read(t, id)
		if current.Title != "after" || len(current.Labels) != 0 {
			t.Fatalf("row = %q %v, want only the title write", current.Title, current.Labels)
		}
	})

	t.Run("a delete loses to a guarded label write that landed first", func(t *testing.T) {
		id := "grd-race3"
		want := seed(t, id)
		deleter, err := NewDeleter(provider)
		if err != nil {
			t.Fatal(err)
		}
		var attempts atomic.Int32
		var labelErr error
		_, deleteErr := deleter.Delete(ctx, publicops.DeleteRequest{
			Actor: "first", IDs: []string{id},
			InTx: func(current *types.Issue) error {
				if err := staleGuard(want, current); err != nil {
					return err
				}
				if attempts.Add(1) == 1 {
					_, labelErr = lifecycle.Update(ctx, publicops.UpdateRequest{
						Actor: "second", IssueID: id, Patch: label("x"),
						InTx: func(current *types.Issue, _ *publicops.UpdateRequest) error { return staleGuard(want, current) },
					})
				}
				return nil
			},
		})
		if labelErr != nil {
			t.Fatalf("the write that landed first was refused: %v", labelErr)
		}
		if !errors.Is(deleteErr, publicops.ErrVersionMismatch) {
			t.Fatalf("the stale delete came back %v, want a version mismatch", deleteErr)
		}
		if current := read(t, id); len(current.Labels) != 1 {
			t.Fatalf("labels = %v", current.Labels)
		}
	})

	t.Run("a reopen writes row_lock, so it loses to a guarded label write", func(t *testing.T) {
		id := "grd-race4"
		seed(t, id)
		if _, err := lifecycle.Close(ctx, publicops.CloseRequest{Actor: "seed", IssueID: id, Reason: "done"}); err != nil {
			t.Fatal(err)
		}
		want := guardFingerprint(read(t, id))
		var attempts atomic.Int32
		var labelErr error
		_, reopenErr := lifecycle.Reopen(ctx, publicops.ReopenRequest{
			Actor: "first", IssueID: id,
			InTx: &publicops.RowCheck{Run: func(current *types.Issue) error {
				if err := staleGuard(want, current); err != nil {
					return err
				}
				if attempts.Add(1) == 1 {
					_, labelErr = lifecycle.Update(ctx, publicops.UpdateRequest{
						Actor: "second", IssueID: id, Patch: label("x"),
						InTx: func(current *types.Issue, _ *publicops.UpdateRequest) error { return staleGuard(want, current) },
					})
				}
				return nil
			}},
		})
		if labelErr != nil {
			t.Fatalf("the write that landed first was refused: %v", labelErr)
		}
		if !errors.Is(reopenErr, publicops.ErrVersionMismatch) {
			t.Fatalf("the stale reopen came back %v, want a version mismatch", reopenErr)
		}
	})

	t.Run("a close loses to a guarded label write", func(t *testing.T) {
		id := "grd-race5"
		want := seed(t, id)
		var attempts atomic.Int32
		var labelErr error
		_, closeErr := lifecycle.Close(ctx, publicops.CloseRequest{
			Actor: "first", IssueID: id, Reason: "done",
			InTx: &publicops.RowCheck{Run: func(current *types.Issue) error {
				if err := staleGuard(want, current); err != nil {
					return err
				}
				if attempts.Add(1) == 1 {
					_, labelErr = lifecycle.Update(ctx, publicops.UpdateRequest{
						Actor: "second", IssueID: id, Patch: label("x"),
						InTx: func(current *types.Issue, _ *publicops.UpdateRequest) error { return staleGuard(want, current) },
					})
				}
				return nil
			}},
		})
		if labelErr != nil {
			t.Fatalf("the write that landed first was refused: %v", labelErr)
		}
		if !errors.Is(closeErr, publicops.ErrVersionMismatch) {
			t.Fatalf("the stale close came back %v, want a version mismatch", closeErr)
		}
	})

	t.Run("a guard can compute the patch from the state the write lands on", func(t *testing.T) {
		id := "grd-derive"
		seed(t, id)
		result, err := lifecycle.Update(ctx, publicops.UpdateRequest{
			Actor: "derive", IssueID: id,
			InTx: func(current *types.Issue, attempt *publicops.UpdateRequest) error {
				attempt.Patch = title(current.Title + " and after")
				return nil
			},
		})
		if err != nil || !result.Changed || result.Issue.Title != "before and after" {
			t.Fatalf("update = %+v, %v", result, err)
		}
	})

	t.Run("an empty patch is a no-op that moves nothing", func(t *testing.T) {
		id := "grd-noop"
		seed(t, id)
		before := read(t, id)
		result, err := lifecycle.Update(ctx, publicops.UpdateRequest{
			Actor: "noop", IssueID: id,
			InTx: func(*types.Issue, *publicops.UpdateRequest) error { return nil },
		})
		if err != nil || result.Changed {
			t.Fatalf("update = %+v, %v", result, err)
		}
		if after := read(t, id); after.RowVersion != before.RowVersion || !after.UpdatedAt.Equal(before.UpdatedAt) {
			t.Fatalf("a no-op moved the row: %d %v -> %d %v", before.RowVersion, before.UpdatedAt, after.RowVersion, after.UpdatedAt)
		}
	})
}
