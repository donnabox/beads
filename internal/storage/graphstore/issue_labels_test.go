//go:build cgo

package graphstore

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/types"
	publicops "github.com/steveyegge/beads/issueops"
)

func issueLabels(values ...string) *[]string { return &values }

func assertIssueLabels(t *testing.T, got, want []string) {
	t.Helper()
	got, want = slices.Clone(got), slices.Clone(want)
	slices.Sort(got)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Fatalf("labels = %q, want %q", got, want)
	}
}

// reopenState covers payloads, versions, audit, coordination and owned state;
// include the separate authoritative label rows to detect partial label writes.
func issueLabelsState(t *testing.T, ctx context.Context, s *Store) map[string]string {
	t.Helper()
	state := reopenState(t, ctx, s)
	rows, err := s.db.QueryContext(ctx, "SELECT issue_id, label FROM labels ORDER BY issue_id, label")
	if err != nil {
		t.Fatal(err)
	}
	var labels [][2]string
	for rows.Next() {
		var row [2]string
		if err := rows.Scan(&row[0], &row[1]); err != nil {
			_ = rows.Close()
			t.Fatal(err)
		}
		labels = append(labels, row)
	}
	if err := errors.Join(rows.Err(), rows.Close()); err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(labels)
	if err != nil {
		t.Fatal(err)
	}
	state["labels"] = string(raw)
	return state
}

func TestIssueLabelsLifecycle(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, _, s, original, target, dependency := reopenFixture(t, backend)
			memory, err := s.Create(ctx, CreateRequest{Path: "beads/context", Body: "Memory context"})
			if err != nil {
				t.Fatal(err)
			}
			info, err := s.AddInformationalLink(ctx, LinkCreateRequest{Path: "links/context", SourcePath: "beads/work", TargetPath: "beads/context", Actor: "author"})
			if err != nil {
				t.Fatal(err)
			}
			saved := []IssueRecord{original}
			wanted := []string{"Alpha", "alpha", "cafe", "café", "É", "é"}
			supplied := []string{"", "Alpha", "alpha", "cafe", "café", "É", "é", "Alpha", ""}
			coordinated := false
			s.afterWrite = func(stage string) error {
				if stage == "coordination" {
					coordinated = true
					supplied[1] = "caller mutated backing array"
					supplied = []string{"caller replaced slice"}
				}
				return nil
			}
			edited, err := s.UpdateIssue(ctx, UpdateIssueRequest{Path: "beads/work", Actor: "label-editor", ExpectedRevision: original.Revision, Labels: &supplied})
			s.afterWrite = nil
			if err != nil || !coordinated || !edited.Changed || edited.Issue.Revision == original.Revision || edited.Issue.Attribution.Actor != "label-editor" {
				t.Fatalf("replacement: %+v coordinated=%v err=%v", edited, coordinated, err)
			}
			assertIssueLabels(t, edited.Issue.Properties.Labels, wanted)
			properties := *edited.Issue.Properties
			properties.Labels = original.Properties.Labels
			properties.UpdatedAt = original.Properties.UpdatedAt
			properties.ContentHash = original.Properties.ContentHash
			properties.RowVersion = original.Properties.RowVersion
			if !reflect.DeepEqual(properties, *original.Properties) || !reflect.DeepEqual(edited.Issue.Owned, original.Owned) {
				t.Fatal("label replacement changed unrelated properties or owned Dependency")
			}
			assertIssueEditCounts(t, ctx, s, original.Properties.ID, 3)
			saved = append(saved, edited.Issue)
			for _, label := range wanted {
				page, err := s.ListIssues(ctx, publicops.ListRequest{Labels: []string{label}})
				if err != nil || page.HasMore || len(page.Items) != 1 || !reflect.DeepEqual(page.Items[0], edited.Issue) {
					t.Fatalf("list label %q: %+v %v", label, page, err)
				}
			}

			state := issueLabelsState(t, ctx, s)
			for _, unconditional := range []bool{false, true} {
				request := UpdateIssueRequest{Path: "beads/work", Actor: "different-actor", Unconditional: unconditional, Labels: issueLabels("é", "É", "café", "cafe", "alpha", "Alpha", "Alpha", "")}
				if !unconditional {
					request.ExpectedRevision = edited.Issue.Revision
				}
				s.afterWrite = func(stage string) error { return errors.New("no-op reached write stage " + stage) }
				got, err := s.UpdateIssue(ctx, request)
				s.afterWrite = nil
				if err != nil || got.Changed || !reflect.DeepEqual(got.Issue, edited.Issue) {
					t.Fatalf("equivalent set: %+v %v", got, err)
				}
			}
			got, err := s.UpdateIssue(ctx, UpdateIssueRequest{Path: "beads/work", Actor: "editor", ExpectedRevision: original.Revision, Labels: &wanted})
			if !errors.Is(err, ErrConflict) || !reflect.DeepEqual(got, IssueMutationResult{}) {
				t.Fatalf("stale equivalent set: %+v %v", got, err)
			}
			if !reflect.DeepEqual(state, issueLabelsState(t, ctx, s)) {
				t.Fatal("equivalent set/refusal changed labels, attribution, events, coordination or versions")
			}

			omitted, err := s.UpdateIssue(ctx, UpdateIssueRequest{Path: "beads/work", Actor: "text-editor", ExpectedRevision: edited.Issue.Revision, Title: issueEditString("Renamed")})
			if err != nil || !omitted.Changed {
				t.Fatalf("omitted labels: %+v %v", omitted, err)
			}
			assertIssueLabels(t, omitted.Issue.Properties.Labels, wanted)
			saved = append(saved, omitted.Issue)
			assertIssueEditCounts(t, ctx, s, original.Properties.ID, 4)
			mixed, err := s.UpdateIssue(ctx, UpdateIssueRequest{Path: "beads/work", Actor: "mixed-editor", ExpectedRevision: omitted.Issue.Revision, Labels: issueLabels("new", "Alpha", "new", ""), Priority: issuePriority(0), Description: issueEditString("Atomic description")})
			if err != nil || !mixed.Changed || mixed.Issue.Properties.Priority != 0 || mixed.Issue.Properties.Description != "Atomic description" || mixed.Issue.Properties.Title != "Renamed" || mixed.Issue.Attribution.Actor != "mixed-editor" || !reflect.DeepEqual(mixed.Issue.Owned, original.Owned) {
				t.Fatalf("mixed edit: %+v %v", mixed, err)
			}
			assertIssueLabels(t, mixed.Issue.Properties.Labels, []string{"new", "Alpha"})
			assertIssueEditCounts(t, ctx, s, original.Properties.ID, 5)
			saved = append(saved, mixed.Issue)
			page, err := s.ListIssues(ctx, publicops.ListRequest{Labels: []string{"alpha"}})
			if err != nil || len(page.Items) != 0 {
				t.Fatalf("removed case-distinct label still selected: %+v %v", page, err)
			}
			cleared, err := s.UpdateIssue(ctx, UpdateIssueRequest{Path: "beads/work", Actor: "clearer", Unconditional: true, Labels: issueLabels()})
			if err != nil || !cleared.Changed || len(cleared.Issue.Properties.Labels) != 0 || !reflect.DeepEqual(cleared.Issue.Owned, original.Owned) {
				t.Fatalf("explicit empty clear: %+v %v", cleared, err)
			}
			assertIssueEditCounts(t, ctx, s, original.Properties.ID, 6)
			saved = append(saved, cleared.Issue)
			state = issueLabelsState(t, ctx, s)
			for _, labels := range [][]string{nil, {}, {"", ""}} {
				got, err := s.UpdateIssue(ctx, UpdateIssueRequest{Path: "beads/work", Actor: "noop-clearer", ExpectedRevision: cleared.Issue.Revision, Labels: &labels})
				if err != nil || got.Changed || !reflect.DeepEqual(got.Issue, cleared.Issue) {
					t.Fatalf("empty-set no-op: %+v %v", got, err)
				}
			}
			if !reflect.DeepEqual(state, issueLabelsState(t, ctx, s)) {
				t.Fatal("empty-set no-op changed stored state")
			}
			for _, record := range saved {
				assertIssueListRetained(t, ctx, s, "beads/work", record)
			}
			for path, want := range map[string]LinkRecord{"links/block": dependency, "links/context": info.Link} {
				if got, err := s.ShowLink(ctx, path); err != nil || !reflect.DeepEqual(got, want) {
					t.Fatalf("Link %s changed: %+v %v", path, got, err)
				}
			}
			if got, err := s.ShowIssue(ctx, "beads/prereq"); err != nil || !reflect.DeepEqual(got, target) {
				t.Fatalf("dependency target changed: %+v %v", got, err)
			}
			if got, err := s.Show(ctx, "beads/context"); err != nil || !reflect.DeepEqual(got, memory) {
				t.Fatalf("Memory changed: %+v %v", got, err)
			}
			assertReadyIDs(t, ctx, s, target.ID)
		})
	}
}

func TestIssueLabelsRefusalAndRollback(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, _, s, original, _, _ := reopenFixture(t, backend)
			seeded, err := s.UpdateIssue(ctx, UpdateIssueRequest{Path: "beads/work", Actor: "seed", ExpectedRevision: original.Revision, Labels: issueLabels("keep", "remove")})
			if err != nil || !seeded.Changed {
				t.Fatalf("seed labels: %+v %v", seeded, err)
			}
			state := issueLabelsState(t, ctx, s)
			request := UpdateIssueRequest{Path: "beads/work", Actor: "editor", ExpectedRevision: seeded.Issue.Revision, Title: issueEditString("Atomic title"), Priority: issuePriority(0), Labels: issueLabels("keep", "new")}
			for _, tc := range []struct {
				name  string
				label string
				want  error
			}{
				{"invalid UTF-8", string([]byte{255}), storage.ErrValidation},
				{"overlength", strings.Repeat("x", types.MaxFieldLen+1), types.ErrFieldTooLong},
			} {
				t.Run(tc.name, func(t *testing.T) {
					r := request
					r.Labels = issueLabels("valid", tc.label)
					touched := false
					s.afterWrite = func(string) error { touched = true; return nil }
					got, err := s.UpdateIssue(ctx, r)
					s.afterWrite = nil
					if !errors.Is(err, tc.want) || touched || !reflect.DeepEqual(got, IssueMutationResult{}) {
						t.Fatalf("refusal: %+v touched=%v err=%v", got, touched, err)
					}
					if !reflect.DeepEqual(state, issueLabelsState(t, ctx, s)) {
						t.Fatal("invalid request changed stored state")
					}
				})
			}
			for _, stage := range []string{"issue-update", "issue-retained", "source-catalog", "source-retained"} {
				t.Run(stage, func(t *testing.T) {
					fault := errors.New("label edit rollback at " + stage)
					s.afterWrite = func(at string) error {
						if at == stage {
							return fault
						}
						return nil
					}
					got, err := s.UpdateIssue(ctx, request)
					s.afterWrite = nil
					if !errors.Is(err, fault) || !reflect.DeepEqual(got, IssueMutationResult{}) {
						t.Fatalf("rollback: %+v %v", got, err)
					}
					if !reflect.DeepEqual(state, issueLabelsState(t, ctx, s)) {
						t.Fatal("failed mixed edit leaked labels, scalar or retained state")
					}
				})
			}
			assertIssueEditCounts(t, ctx, s, original.Properties.ID, 3)
			assertIssueListRetained(t, ctx, s, "beads/work", original)
			assertIssueListRetained(t, ctx, s, "beads/work", seeded.Issue)
		})
	}
}

func TestIssueLabelsClosedIssue(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, _, s, _, _, dependency := reopenFixture(t, backend)
			target, err := s.CloseIssue(ctx, "beads/prereq", "prerequisite complete", "closer")
			if err != nil {
				t.Fatal(err)
			}
			closed, err := s.CloseIssue(ctx, "beads/work", "work complete", "closer")
			if err != nil {
				t.Fatal(err)
			}
			spectator, err := s.CreateIssue(ctx, "beads/unrelated", plainIssue("Unrelated ready work"))
			if err != nil {
				t.Fatal(err)
			}
			readyBefore, err := s.ReadyIssues(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if len(readyBefore) != 1 || readyBefore[0].ID != spectator.ID {
				t.Fatalf("closed fixture readiness: %+v", readyBefore)
			}
			before := closed.Issue
			if before.Properties.Status != types.StatusClosed || before.Properties.ClosedAt == nil || before.Properties.CloseReason != "work complete" || len(before.Owned) != 1 {
				t.Fatal("fixture must be a closed Issue with owned Dependency")
			}
			counts := make(map[string]int)
			for _, table := range []string{"issue_versions", "graph_preview_issue_versions"} {
				var count int
				if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table+" WHERE issue_id=?", before.Properties.ID).Scan(&count); err != nil {
					t.Fatal(err)
				}
				counts[table] = count
			}
			eventCounts := func() map[string]int {
				t.Helper()
				counts := make(map[string]int)
				for _, kind := range []string{"label_added", "label_removed", "updated", "reopened", "closed", "all"} {
					query := "SELECT COUNT(*) FROM events WHERE issue_id=?"
					args := []any{before.Properties.ID}
					if kind != "all" {
						query += " AND event_type=?"
						args = append(args, kind)
					}
					var count int
					if err := s.db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
						t.Fatal(err)
					}
					counts[kind] = count
				}
				return counts
			}
			beforeEvents := eventCounts()
			edited, err := s.UpdateIssue(ctx, UpdateIssueRequest{Path: "beads/work", Actor: "closed-label-editor", ExpectedRevision: before.Revision, Labels: issueLabels("graph", "closed-label")})
			if err != nil || !edited.Changed || edited.Issue.Revision == before.Revision || edited.Issue.Attribution.Actor != "closed-label-editor" {
				t.Fatalf("closed label edit: %+v %v", edited, err)
			}
			assertIssueLabels(t, edited.Issue.Properties.Labels, []string{"graph", "closed-label"})
			properties := *edited.Issue.Properties
			properties.Labels = before.Properties.Labels
			properties.UpdatedAt = before.Properties.UpdatedAt
			properties.RowVersion = before.Properties.RowVersion
			properties.ContentHash = before.Properties.ContentHash
			if !reflect.DeepEqual(properties, *before.Properties) || !reflect.DeepEqual(edited.Issue.Owned, before.Owned) {
				t.Fatal("label edit changed closure, unrelated properties, or owned Dependency")
			}
			current, err := s.ShowIssue(ctx, "beads/work")
			if err != nil || !reflect.DeepEqual(current, edited.Issue) {
				t.Fatalf("closed current read: %+v %v", current, err)
			}
			for table, count := range counts {
				var got int
				if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table+" WHERE issue_id=?", before.Properties.ID).Scan(&got); err != nil || got != count+1 {
					t.Fatalf("%s count%d want%d err%v", table, got, count+1, err)
				}
			}
			afterEvents := eventCounts()
			for kind, count := range beforeEvents {
				delta := 0
				if kind == "label_added" || kind == "label_removed" {
					delta = 1
				}
				if kind == "all" {
					delta = 2
				}
				if afterEvents[kind] != count+delta {
					t.Fatalf("closed edit %s events%d want%d", kind, afterEvents[kind], count+delta)
				}
			}
			var authorEvents int
			if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM events WHERE issue_id=? AND actor='closed-label-editor' AND event_type IN ('label_added','label_removed')", before.Properties.ID).Scan(&authorEvents); err != nil || authorEvents != 2 {
				t.Fatalf("label event authorship%d err%v", authorEvents, err)
			}
			if readyAfter, err := s.ReadyIssues(ctx); err != nil || !reflect.DeepEqual(readyAfter, readyBefore) {
				t.Fatalf("labels reopened/changed readiness: %+v %v", readyAfter, err)
			}
			if got, err := s.ShowLink(ctx, "links/block"); err != nil || !reflect.DeepEqual(got, dependency) {
				t.Fatalf("Dependency changed: %+v %v", got, err)
			}
			if got, err := s.ShowIssue(ctx, "beads/prereq"); err != nil || !reflect.DeepEqual(got, target.Issue) {
				t.Fatalf("closed target changed: %+v %v", got, err)
			}
			assertIssueListRetained(t, ctx, s, "beads/work", before)
			assertIssueListRetained(t, ctx, s, "beads/work", edited.Issue)
		})
	}
}
