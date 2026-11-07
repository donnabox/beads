//go:build cgo

package httpclient

import (
	"context"
	"errors"
	"slices"
	"sort"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/types"
)

// The raw reads a composition of v0 operations answers (raw_reads.go) and the
// listing shape of SearchIssues, run as a dual run: the same method asked of the
// reference store and of the client, and the answers compared, because each of
// these used to refuse and a command reading through one printed a blank instead.

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

func TestServedRawReadsAgreeWithTheReference(t *testing.T) {
	skipUnlessEmbeddedDolt(t)
	ctx := context.Background()
	env := newServedEnv(t, "rr")
	seedRawReads(t, ctx, env)

	t.Run("GetDependents hydrates what the wire projects away", func(t *testing.T) {
		got, err := env.subject.GetDependents(ctx, "rr-blocker")
		if err != nil {
			t.Fatal(err)
		}
		want, err := env.reference.GetDependents(ctx, "rr-blocker")
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(rawReadIDs(got), rawReadIDs(want)) || len(got) != 1 {
			t.Fatalf("got %v, want %v", rawReadIDs(got), rawReadIDs(want))
		}
		if got[0].CreatedAt.IsZero() || got[0].UpdatedAt.IsZero() {
			t.Errorf("the dependent came back shallow: %+v", got[0])
		}
	})

	t.Run("GetDependencies", func(t *testing.T) {
		got, err := env.subject.GetDependencies(ctx, "rr-blocked")
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 || got[0].ID != "rr-blocker" || got[0].Description != "body of the blocker" {
			t.Fatalf("got %+v", got)
		}
	})

	t.Run("GetIssuesByIDs drops a miss and carries labels", func(t *testing.T) {
		got, err := env.subject.GetIssuesByIDs(ctx, []string{"rr-blocker", "rr-none", "rr-wisp"})
		if err != nil {
			t.Fatal(err)
		}
		if ids := rawReadIDs(got); !slices.Equal(ids, []string{"rr-blocker", "rr-wisp"}) {
			t.Fatalf("got %v", ids)
		}
	})

	t.Run("GetDependencyRecordsForIssues keys only ids with edges", func(t *testing.T) {
		got, err := env.subject.GetDependencyRecordsForIssues(ctx, []string{"rr-blocked", "rr-tmpl", "rr-none"})
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 || len(got["rr-blocked"]) != 1 || got["rr-blocked"][0].DependsOnID != "rr-blocker" {
			t.Fatalf("got %+v", got)
		}
	})

	t.Run("GetDependencyCounts is the blocks-only pair", func(t *testing.T) {
		ids := []string{"rr-blocker", "rr-blocked", "rr-epic"}
		got, err := env.subject.GetDependencyCounts(ctx, ids)
		if err != nil {
			t.Fatal(err)
		}
		want, err := env.reference.GetDependencyCounts(ctx, ids)
		if err != nil {
			t.Fatal(err)
		}
		for _, id := range ids {
			if *got[id] != *want[id] {
				t.Errorf("%s: got %+v, want %+v", id, *got[id], *want[id])
			}
		}
		if got["rr-blocker"].DependentCount != 1 || got["rr-blocked"].DependencyCount != 1 {
			t.Errorf("the seeded edge did not count: %+v %+v", *got["rr-blocker"], *got["rr-blocked"])
		}
	})

	t.Run("DetectCycles on an acyclic graph", func(t *testing.T) {
		got, err := env.subject.DetectCycles(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 0 {
			t.Fatalf("got %d cycles", len(got))
		}
	})

	t.Run("GetEpicsEligibleForClosure", func(t *testing.T) {
		got, err := env.subject.GetEpicsEligibleForClosure(ctx)
		if err != nil {
			t.Fatal(err)
		}
		want, err := env.reference.GetEpicsEligibleForClosure(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != len(want) || len(got) != 1 {
			t.Fatalf("got %d, want %d", len(got), len(want))
		}
		if got[0].Epic.ID != "rr-epic" || got[0].TotalChildren != want[0].TotalChildren ||
			got[0].ClosedChildren != want[0].ClosedChildren || got[0].EligibleForClose != want[0].EligibleForClose {
			t.Fatalf("got %+v, want %+v", *got[0], *want[0])
		}
	})

	t.Run("GetStaleIssues honors the status, the plane and the limit", func(t *testing.T) {
		// A day in the future makes every row older than the cutoff.
		for _, filter := range []types.StaleFilter{
			{Days: -1},
			{Days: -1, Status: "closed"},
			{Days: -1, Limit: 1},
		} {
			got, err := env.subject.GetStaleIssues(ctx, filter)
			if err != nil {
				t.Fatal(err)
			}
			want, err := env.reference.GetStaleIssues(ctx, filter)
			if err != nil {
				t.Fatal(err)
			}
			if filter.Limit > 0 {
				if len(got) != len(want) {
					t.Errorf("%+v: got %d rows, want %d", filter, len(got), len(want))
				}
				continue
			}
			if !slices.Equal(rawReadIDs(got), rawReadIDs(want)) {
				t.Errorf("%+v: got %v, want %v", filter, rawReadIDs(got), rawReadIDs(want))
			}
		}
	})
}

// TestServedStaleExemptsALiveHeartbeat: the local query leaves out an issue whose
// lease heartbeat is newer than the cutoff, so an actively worked claim is not
// stale whatever its updated_at says. The listing has no predicate for it.
func TestServedStaleExemptsALiveHeartbeat(t *testing.T) {
	skipUnlessEmbeddedDolt(t)
	ctx := context.Background()
	env := newServedEnv(t, "sb")
	if err := env.createIssue(ctx, &types.Issue{ID: "sb-1", Title: "claimed", Status: types.StatusOpen, IssueType: types.TypeTask}, "seed"); err != nil {
		t.Fatal(err)
	}
	if err := env.reference.ClaimIssue(ctx, "sb-1", "ada"); err != nil {
		t.Fatal(err)
	}
	if err := env.reference.HeartbeatIssue(ctx, "sb-1", "ada"); err != nil {
		t.Fatal(err)
	}
	// Heartbeats no longer stamp updated_at, so this is what a long-running claim
	// looks like: an old row with a fresh lease.
	old := time.Now().UTC().AddDate(0, 0, -30)
	if err := env.execRaw(ctx, "UPDATE issues SET updated_at = ? WHERE id = ?", old, "sb-1"); err != nil {
		t.Fatal(err)
	}
	want, err := env.reference.GetStaleIssues(ctx, types.StaleFilter{Days: 7})
	if err != nil {
		t.Fatal(err)
	}
	got, err := env.subject.GetStaleIssues(ctx, types.StaleFilter{Days: 7})
	if err != nil {
		t.Fatal(err)
	}
	if len(want) != 0 {
		t.Fatalf("the reference reports %v stale; the fixture is not the case this pins", rawReadIDs(want))
	}
	if len(got) != 0 {
		t.Fatalf("got %v, want a live heartbeat to exempt the row", rawReadIDs(got))
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
