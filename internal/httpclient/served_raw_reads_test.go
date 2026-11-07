//go:build cgo

package httpclient

import (
	"context"
	"errors"
	"slices"
	"sort"
	"testing"

	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/types"
)

// The listing shape of SearchIssues, run as a dual run: the same method asked of
// the reference store and of the client, and the answers compared, because it
// used to refuse and a command reading through it printed a blank instead.

func rawReadIDs(rows []*types.Issue) []string {
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	sort.Strings(ids)
	return ids
}

// seedRawReads plants a small graph: an epic with one closed and one open child,
// a blocker pair, a template, a gate, a wisp, and a closed issue.
func seedRawReads(t *testing.T, ctx context.Context, e *servedEnv) {
	t.Helper()
	for _, row := range []*types.Issue{
		{ID: "rr-epic", Title: "epic", Status: types.StatusOpen, IssueType: types.IssueType("epic"), Labels: []string{"rr-red"}},
		{ID: "rr-done", Title: "done child", Status: types.StatusOpen, IssueType: types.TypeTask},
		{ID: "rr-open", Title: "open child", Status: types.StatusOpen, IssueType: types.TypeTask},
		{ID: "rr-blocker", Title: "blocker", Description: "body of the blocker", Status: types.StatusOpen, IssueType: types.TypeTask, Priority: 1, Labels: []string{"rr-blue"}},
		{ID: "rr-blocked", Title: "blocked", Status: types.StatusOpen, IssueType: types.TypeTask},
		{ID: "rr-tmpl", Title: "template", Status: types.StatusOpen, IssueType: types.TypeTask, IsTemplate: true},
		{ID: "rr-closed", Title: "closed", Status: types.StatusOpen, IssueType: types.TypeTask},
	} {
		if err := e.createIssue(ctx, row, "seed"); err != nil {
			t.Fatalf("seed %s: %v", row.ID, err)
		}
	}
	if err := e.createWisp(ctx, &types.Issue{ID: "rr-wisp", Title: "wisp", Status: types.StatusOpen, IssueType: types.TypeTask}, "seed"); err != nil {
		t.Fatalf("seed wisp: %v", err)
	}
	for _, edge := range []*types.Dependency{
		{IssueID: "rr-done", DependsOnID: "rr-epic", Type: types.DepParentChild},
		{IssueID: "rr-open", DependsOnID: "rr-epic", Type: types.DepParentChild},
		{IssueID: "rr-blocked", DependsOnID: "rr-blocker", Type: types.DepBlocks},
	} {
		if err := e.addDependency(ctx, edge, "seed"); err != nil {
			t.Fatalf("seed edge %s: %v", edge.IssueID, err)
		}
	}
	for _, id := range []string{"rr-done", "rr-closed"} {
		if err := e.reference.CloseIssue(ctx, id, "shipped", "seed", ""); err != nil {
			t.Fatalf("close %s: %v", id, err)
		}
	}
}

// TestServedSearchIssuesListingShape: a filter naming neither ids nor a parent is
// the listing shape, and an unset member restricts nothing.
func TestServedSearchIssuesListingShape(t *testing.T) {
	skipUnlessEmbeddedDolt(t)
	ctx := context.Background()
	env := newServedEnv(t, "rr")
	seedRawReads(t, ctx, env)

	open := types.StatusOpen
	taskType := types.TypeTask
	label := "rr-blue"
	no := false
	for name, filter := range map[string]types.IssueFilter{
		"the empty filter is every row on both planes": {},
		"a status":           {Status: &open},
		"several statuses":   {Statuses: []types.Status{types.StatusOpen, types.StatusClosed}},
		"a type":             {IssueType: &taskType},
		"a label":            {Labels: []string{label}},
		"an excluded status": {ExcludeStatus: []types.Status{types.StatusClosed}},
		"no templates":       {IsTemplate: &no},
		"a title":            {TitleContains: "child"},
	} {
		t.Run(name, func(t *testing.T) {
			got, err := env.subject.SearchIssues(ctx, "", filter)
			if err != nil {
				t.Fatal(err)
			}
			want, err := env.reference.SearchIssues(ctx, "", filter)
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(rawReadIDs(got), rawReadIDs(want)) {
				t.Fatalf("got %v, want %v", rawReadIDs(got), rawReadIDs(want))
			}
		})
	}

	t.Run("the limit and the row cap hold", func(t *testing.T) {
		got, err := env.subject.SearchIssues(ctx, "", types.IssueFilter{Limit: 2})
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 2 {
			t.Fatalf("got %d rows, want 2", len(got))
		}
		_, err = env.subject.SearchIssues(ctx, "", types.IssueFilter{MaxRows: 1})
		if err == nil {
			t.Fatal("a cap below the row count answered")
		}
	})

	t.Run("an inexpressible member refuses and sends nothing wider", func(t *testing.T) {
		prefix := "rr-"
		_, err := env.subject.SearchIssues(ctx, "", types.IssueFilter{IDPrefix: prefix})
		var unsupported *storage.ErrUnsupported
		if !errors.As(err, &unsupported) {
			t.Fatalf("IDPrefix: %v", err)
		}
	})
}
