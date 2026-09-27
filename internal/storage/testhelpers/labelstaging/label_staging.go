// Package labelstaging provides adapter-specific committed-state regression
// fixtures. These checks are internal storage tests, not public role contracts.
package labelstaging

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/types"
	publicops "github.com/steveyegge/beads/issueops"
)

// LabelStagingFixture uses production adapter entry points and read-only HEAD
// observations. Commit is used only to establish the seed before the operation
// under test; never to repair staging after that operation.
type LabelStagingFixture struct {
	IssuePrefix  string
	Operations   publicops.Lifecycle
	CreateIssue  func(context.Context, *types.Issue, string) error
	Commit       func(context.Context, string) error
	Exec         func(context.Context, string, ...any) error
	QueryScalar  func(context.Context, string, []any, ...any) error
	GetIssue     func(context.Context, string) (*types.Issue, error)
	Transaction  func(context.Context, string, func(storage.Transaction) error) error
	DirectAdd    func(context.Context, string, string, string) error
	DirectRemove func(context.Context, string, string, string) error
}

func seedLabelStaging(t *testing.T, ctx context.Context, f LabelStagingFixture, suffix string, wisp bool) (*types.Issue, *types.Issue) {
	t.Helper()
	past := time.Date(2002, 3, 4, 5, 6, 7, 0, time.UTC)
	issue := &types.Issue{ID: f.IssuePrefix + "-" + suffix, Title: "label staging", Status: types.StatusOpen, Priority: 2, IssueType: types.TypeTask, CreatedAt: past, UpdatedAt: past, Labels: []string{"keep", "remove"}, Ephemeral: wisp}
	dirty := &types.Issue{ID: issue.ID + "-unrelated", Title: "committed title", Status: types.StatusOpen, Priority: 2, IssueType: types.TypeTask}
	for _, row := range []*types.Issue{issue, dirty} {
		if err := f.CreateIssue(ctx, row, "seed"); err != nil {
			t.Fatal(err)
		}
	}
	if err := f.Commit(ctx, "seed label staging only"); err != nil {
		t.Fatal(err)
	}
	var err error
	issue, err = f.GetIssue(ctx, issue.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !issue.UpdatedAt.Equal(past) {
		t.Fatalf("seed constituent labels changed explicit timestamp: %v", issue.UpdatedAt)
	}
	assertScalar(t, ctx, f, "transaction auto-commit disabled", 0, "SELECT @@dolt_transaction_commit", nil)
	return issue, dirty
}

func labelStagingHead(t *testing.T, ctx context.Context, f LabelStagingFixture) (string, int) {
	t.Helper()
	var hash string
	var count int
	if err := f.QueryScalar(ctx, "SELECT DOLT_HASHOF('HEAD')", nil, &hash); err != nil {
		t.Fatal(err)
	}
	if err := f.QueryScalar(ctx, "SELECT COUNT(*) FROM dolt_log", nil, &count); err != nil {
		t.Fatal(err)
	}
	return hash, count
}

func runLabelStagingChange(ctx context.Context, f LabelStagingFixture, route, id string, remove, mixed bool) error {
	switch route {
	case "lifecycle":
		patch := publicops.IssuePatch{}
		if remove {
			patch.Labels.Remove = []string{"remove"}
		} else {
			patch.Labels.Add = []string{"added"}
		}
		if mixed {
			patch.Title = publicops.Field[string]{Set: true, Value: "mixed title"}
		}
		result, err := f.Operations.Update(ctx, publicops.UpdateRequest{Actor: "label-writer", IssueID: id, Patch: patch})
		if err == nil && !result.Changed {
			return fmt.Errorf("actual label mutation returned unchanged")
		}
		return err
	case "transaction":
		return f.Transaction(ctx, "selective label mutation", func(tx storage.Transaction) error {
			if remove {
				return tx.RemoveLabel(ctx, id, "remove", "label-writer")
			}
			return tx.AddLabel(ctx, id, "added", "label-writer")
		})
	case "direct":
		if remove {
			return f.DirectRemove(ctx, id, "remove", "label-writer")
		}
		return f.DirectAdd(ctx, id, "added", "label-writer")
	}
	return fmt.Errorf("unknown label test route %q", route)
}

func assertLabelStagingCommitted(t *testing.T, ctx context.Context, f LabelStagingFixture, before *types.Issue, remove bool) {
	t.Helper()
	current, err := f.GetIssue(ctx, before.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !current.UpdatedAt.After(before.UpdatedAt) || current.RowVersion == before.RowVersion {
		t.Fatal("real label mutation did not touch working Issue timestamp/row token")
	}
	var at time.Time
	var token int64
	if err := f.QueryScalar(ctx, "SELECT updated_at, row_lock FROM issues AS OF 'HEAD' WHERE id = ?", []any{before.ID}, &at, &token); err != nil {
		t.Fatal(err)
	}
	if !at.Equal(current.UpdatedAt) || token != current.RowVersion {
		t.Fatalf("F1: Issue label touch omitted from HEAD: HEAD=(%v,%d) working=(%v,%d)", at, token, current.UpdatedAt, current.RowVersion)
	}
	wanted := map[string]bool{"keep": true, "remove": !remove, "added": !remove}
	wantCount := 3
	if remove {
		wantCount = 1
	}
	assertScalar(t, ctx, f, "HEAD full label count", wantCount, "SELECT COUNT(*) FROM labels AS OF 'HEAD' WHERE issue_id = ?", []any{before.ID})
	for label, present := range wanted {
		want := 0
		if present {
			want = 1
		}
		assertScalar(t, ctx, f, "HEAD label "+label, want, "SELECT COUNT(*) FROM labels AS OF 'HEAD' WHERE issue_id = ? AND label = ?", []any{before.ID, label})
	}
	assertScalar(t, ctx, f, "no leftover Issue staging", 0, "SELECT COUNT(*) FROM dolt_status WHERE table_name = 'issues'", nil)
}

// RunLabelStagingMutations checks label-only, removal and mixed edits. Each
// subtest gets old authored timestamps; no sleeps or timestamp SQL are needed.
func RunLabelStagingMutations(t *testing.T, ctx context.Context, f LabelStagingFixture, route string) {
	t.Helper()
	for _, op := range []string{"add", "remove", "mixed"} {
		if op == "mixed" && route != "lifecycle" {
			continue
		}
		t.Run(op, func(t *testing.T) {
			before, _ := seedLabelStaging(t, ctx, f, route+"-"+op, false)
			head, count := labelStagingHead(t, ctx, f)
			var eventCount int
			if err := f.QueryScalar(ctx, "SELECT COUNT(*) FROM events WHERE issue_id = ?", []any{before.ID}, &eventCount); err != nil {
				t.Fatal(err)
			}
			if err := runLabelStagingChange(ctx, f, route, before.ID, op == "remove", op == "mixed"); err != nil {
				t.Fatal(err)
			}
			assertLabelStagingCommitted(t, ctx, f, before, op == "remove")
			var afterEvents int
			if err := f.QueryScalar(ctx, "SELECT COUNT(*) FROM events WHERE issue_id = ?", []any{before.ID}, &afterEvents); err != nil {
				t.Fatal(err)
			}
			if afterEvents <= eventCount {
				t.Fatal("real label mutation omitted working audit event")
			}
			assertLabelEventsIgnored(t, ctx, f)
			after, n := labelStagingHead(t, ctx, f)
			if after == head || n != count+1 {
				t.Fatalf("expected one selective commit, hashes %s/%s counts %d/%d", head, after, count, n)
			}
			if op == "mixed" {
				assertScalar(t, ctx, f, "mixed scalar committed", "mixed title", "SELECT title FROM issues AS OF 'HEAD' WHERE id = ?", []any{before.ID})
			}
		})
	}
}

// RunLabelStagingNoops proves adding issues to a staging list unconditionally
// would be wrong: real change is necessary before staging unrelated dirty rows.
func RunLabelStagingNoops(t *testing.T, ctx context.Context, f LabelStagingFixture, route string) {
	t.Helper()
	before, dirty := seedLabelStaging(t, ctx, f, route+"-noops", false)
	if err := f.Exec(ctx, "UPDATE issues SET title = ? WHERE id = ?", "uncommitted title", dirty.ID); err != nil {
		t.Fatal(err)
	}
	head, count := labelStagingHead(t, ctx, f)
	var events int
	if err := f.QueryScalar(ctx, "SELECT COUNT(*) FROM events WHERE issue_id = ?", []any{before.ID}, &events); err != nil {
		t.Fatal(err)
	}
	for _, remove := range []bool{false, true} {
		var err error
		switch route {
		case "lifecycle":
			patch := publicops.LabelPatch{Add: []string{"keep"}}
			if remove {
				patch = publicops.LabelPatch{Remove: []string{"missing"}}
			}
			var result publicops.UpdateResult
			result, err = f.Operations.Update(ctx, publicops.UpdateRequest{Actor: "noop", IssueID: before.ID, Patch: publicops.IssuePatch{Labels: patch}})
			if err == nil && result.Changed {
				t.Fatal("label no-op reported changed")
			}
		case "transaction":
			err = f.Transaction(ctx, "label noop must not commit unrelated Issue", func(tx storage.Transaction) error {
				if remove {
					return tx.RemoveLabel(ctx, before.ID, "missing", "noop")
				}
				return tx.AddLabel(ctx, before.ID, "keep", "noop")
			})
		case "direct":
			if remove {
				err = f.DirectRemove(ctx, before.ID, "missing", "noop")
			} else {
				err = f.DirectAdd(ctx, before.ID, "keep", "noop")
			}
		}
		if err != nil {
			t.Fatal(err)
		}
		after, n := labelStagingHead(t, ctx, f)
		if after != head || n != count {
			t.Fatal("no-op committed unrelated dirty state or minted commit")
		}
		got, err := f.GetIssue(ctx, before.ID)
		if err != nil || !reflect.DeepEqual(got, before) {
			t.Fatalf("no-op changed Issue: %+v %v", got, err)
		}
		assertScalar(t, ctx, f, "no-op event count", events, "SELECT COUNT(*) FROM events WHERE issue_id = ?", []any{before.ID})
		assertScalar(t, ctx, f, "unrelated working edit", "uncommitted title", "SELECT title FROM issues WHERE id = ?", []any{dirty.ID})
		assertScalar(t, ctx, f, "unrelated HEAD unchanged", "committed title", "SELECT title FROM issues AS OF 'HEAD' WHERE id = ?", []any{dirty.ID})
	}
}

func RunLabelStagingTransactionKeepsPriorDirty(t *testing.T, ctx context.Context, f LabelStagingFixture) {
	t.Helper()
	before, _ := seedLabelStaging(t, ctx, f, "transaction-keeps-prior", false)
	if err := f.Transaction(ctx, "real labels followed by noops", func(tx storage.Transaction) error {
		if err := tx.AddLabel(ctx, before.ID, "added", "writer"); err != nil {
			return err
		}
		if err := tx.AddLabel(ctx, before.ID, "added", "writer"); err != nil {
			return err
		}
		return tx.RemoveLabel(ctx, before.ID, "missing", "writer")
	}); err != nil {
		t.Fatal(err)
	}
	assertLabelStagingCommitted(t, ctx, f, before, false)
}

func RunLabelStagingWisp(t *testing.T, ctx context.Context, f LabelStagingFixture, route string) {
	t.Helper()
	before, dirty := seedLabelStaging(t, ctx, f, route+"-wisp", true)
	if err := f.Exec(ctx, "UPDATE issues SET title = ? WHERE id = ?", "uncommitted title", dirty.ID); err != nil {
		t.Fatal(err)
	}
	head, count := labelStagingHead(t, ctx, f)
	if err := runLabelStagingChange(ctx, f, route, before.ID, false, false); err != nil {
		t.Fatal(err)
	}
	current, err := f.GetIssue(ctx, before.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !current.UpdatedAt.After(before.UpdatedAt) || current.RowVersion == before.RowVersion {
		t.Fatal("wisp label mutation omitted working timestamp/row token")
	}
	after, n := labelStagingHead(t, ctx, f)
	if after != head || n != count {
		t.Fatal("wisp label edit staged regular Issue table or minted durable commit")
	}
	assertScalar(t, ctx, f, "wisp label persisted", 1, "SELECT COUNT(*) FROM wisp_labels WHERE issue_id = ? AND label = 'added'", []any{before.ID})
	assertScalar(t, ctx, f, "wisp leaves dirty Issue working", "uncommitted title", "SELECT title FROM issues WHERE id = ?", []any{dirty.ID})
	assertScalar(t, ctx, f, "wisp leaves dirty Issue uncommitted", "committed title", "SELECT title FROM issues AS OF 'HEAD' WHERE id = ?", []any{dirty.ID})
}

// RunLabelOrdinaryCloseGuard pins the proposed widened ordinary RowVersion
// coverage. This regression evidence is not human acceptance of that contract.
func RunLabelOrdinaryCloseGuard(t *testing.T, ctx context.Context, f LabelStagingFixture) {
	t.Helper()
	for _, op := range []string{"add", "remove", "noop"} {
		t.Run(op, func(t *testing.T) {
			before, _ := seedLabelStaging(t, ctx, f, "ordinary-guard-"+op, false)
			if op == "noop" {
				if err := f.Transaction(ctx, "no-op preserves guard", func(tx storage.Transaction) error {
					if err := tx.AddLabel(ctx, before.ID, "keep", "noop"); err != nil {
						return err
					}
					return tx.RemoveLabel(ctx, before.ID, "missing", "noop")
				}); err != nil {
					t.Fatal(err)
				}
			} else if err := runLabelStagingChange(ctx, f, "lifecycle", before.ID, op == "remove", false); err != nil {
				t.Fatal(err)
			}
			current, err := f.GetIssue(ctx, before.ID)
			if err != nil {
				t.Fatal(err)
			}
			head, count := labelStagingHead(t, ctx, f)
			result, err := f.Operations.Close(ctx, publicops.CloseRequest{IssueID: before.ID, Actor: "closer", Reason: "guard probe", ExpectedVersion: &before.RowVersion})
			if op == "noop" {
				if err != nil || !result.Changed || result.Issue.Status != types.StatusClosed {
					t.Fatalf("unchanged label guard refused: %+v %v", result, err)
				}
				return
			}
			if !errors.Is(err, storage.ErrVersionMismatch) || !reflect.DeepEqual(result, publicops.CloseResult{}) {
				t.Fatalf("old ordinary row token did not refuse: %+v %v", result, err)
			}
			after, n := labelStagingHead(t, ctx, f)
			if after != head || n != count {
				t.Fatal("stale close changed HEAD")
			}
			got, err := f.GetIssue(ctx, before.ID)
			if err != nil || !reflect.DeepEqual(got, current) {
				t.Fatalf("stale close changed working Issue: %+v %v", got, err)
			}
		})
	}
}

// Events intentionally stay in the ignored working plane (migration 0062).
// Match the existing adapter staging tests: HEAD may have no events table or
// an empty table, but may never acquire the newly authored audit rows.
func assertLabelEventsIgnored(t *testing.T, ctx context.Context, f LabelStagingFixture) {
	t.Helper()
	assertScalar(t, ctx, f, "audit plane ignored", 1, "SELECT COUNT(*) FROM dolt_ignore WHERE pattern = 'events' AND ignored = true", nil)
	var count int
	if err := f.QueryScalar(ctx, "SELECT COUNT(*) FROM events AS OF 'HEAD'", nil, &count); err == nil {
		if count != 0 {
			t.Fatalf("ignored events have %d HEAD rows", count)
		}
	} else {
		t.Logf("HEAD has no queryable ignored events table: %v", err)
	}
}

func assertScalar[T comparable](t *testing.T, ctx context.Context, fixture LabelStagingFixture, name string, want T, query string, args []any) {
	t.Helper()
	var got T
	if err := fixture.QueryScalar(ctx, query, args, &got); err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	if got != want {
		t.Fatalf("%s = %v, want %v", name, got, want)
	}
}
