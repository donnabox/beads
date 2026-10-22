//go:build cgo

package graphstore

import (
	"database/sql"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/types"
)

func TestIssueNativeSelectorHeartbeatStateAndReclaim(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, options := issueExperimentOptions(t, backend)
			s, err := OpenExisting(ctx, options)
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			created, err := s.CreateIssue(ctx, "beads/custom-path", plainIssue("Custom"))
			if err != nil {
				t.Fatal(err)
			}
			if created.Properties.ID == "custom-path" {
				t.Fatal("fixture needs a native ID distinct from canonical path")
			}
			claimed, err := s.ClaimIssue(ctx, "beads/custom-path", "holder")
			if err != nil || !claimed.Changed {
				t.Fatalf("claim: %+v %v", claimed, err)
			}
			selector := IssueSelector{Path: "beads/" + created.Properties.ID, NativeID: created.Properties.ID}
			shown, err := s.ShowIssueSelected(ctx, selector)
			if err != nil || !reflect.DeepEqual(shown, claimed.Issue) {
				t.Fatalf("selected read: %+v %v", shown, err)
			}
			renewed, err := s.HeartbeatIssueSelected(ctx, selector, "holder")
			if err != nil || renewed.Revision != claimed.Issue.Revision || renewed.Properties.LeaseExpiresAt == nil {
				t.Fatalf("selected heartbeat changed retained History: %+v %v", renewed, err)
			}
			assertIssueEditCounts(t, ctx, s, created.Properties.ID, 2)
			assertIssueEditVersion(t, ctx, s, "beads/custom-path", claimed.Issue)
			if err := s.withTx(ctx, true, func(tx *sql.Tx) error {
				_, err := tx.ExecContext(ctx, `UPDATE leases SET lease_expires_at=? WHERE issue_id=?`, time.Now().UTC().Add(-time.Hour), created.Properties.ID)
				return err
			}); err != nil {
				t.Fatal(err)
			}
			reclaimed, err := s.ReclaimExpiredIssueLeasesSelected(ctx, 0, types.ReclaimFilter{}, []IssueSelector{selector}, "reaper")
			if err != nil || len(reclaimed) != 1 || reclaimed[0].ID != created.Properties.ID {
				t.Fatalf("selected recovery: %+v %v", reclaimed, err)
			}
			after, err := s.ShowIssue(ctx, "beads/custom-path")
			if err != nil || after.Revision == renewed.Revision || after.Properties.Status != types.StatusOpen {
				t.Fatalf("selected recovery lost Issue writer: %+v %v", after, err)
			}
			assertIssueEditCounts(t, ctx, s, created.Properties.ID, 3)
			stable := workflowState(t, ctx, s)
			if _, err := s.HeartbeatIssueSelected(ctx, selector, "holder"); !errors.Is(err, storage.ErrNotClaimable) || !reflect.DeepEqual(stable, workflowState(t, ctx, s)) {
				t.Fatalf("released holder heartbeat changed state: %v", err)
			}
		})
	}
}

func TestIssueSelectorSharesNativeAbbreviationAcrossReadAndUpdate(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, options := issueExperimentOptions(t, backend)
			s, err := OpenExisting(ctx, options)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = s.Close() })
			request := plainIssue("Source")
			request.Issue.ID = "exp-a3f8e9"
			created, err := s.CreateIssue(ctx, "beads/source", request)
			if err != nil {
				t.Fatal(err)
			}
			selector := IssueSelector{Path: "beads/a3f8", NativeID: "a3f8"}
			shown, err := s.ShowIssueSelected(ctx, selector)
			if err != nil || !reflect.DeepEqual(shown, created) {
				t.Fatalf("abbreviated read selected wrong Issue: %+v %v", shown, err)
			}
			paths, err := s.ResolveIssuePaths(ctx, []IssueSelector{selector})
			if err != nil || !reflect.DeepEqual(paths, []string{"beads/source"}) {
				t.Fatalf("abbreviated count filter selected wrong path: %v %v", paths, err)
			}
			title := "Updated"
			updated, err := s.UpdateIssue(ctx, UpdateIssueRequest{Path: selector.Path, NativeID: selector.NativeID, Actor: "editor", ExpectedRevision: created.Revision, Title: &title})
			if err != nil || !updated.Changed || updated.Issue.ID != created.ID || updated.Issue.Properties.Title != title {
				t.Fatalf("abbreviated update selected wrong Issue: %+v %v", updated, err)
			}
			second := plainIssue("Other")
			second.Issue.ID = "exp-a3f8f0"
			if _, err := s.CreateIssue(ctx, "beads/other", second); err != nil {
				t.Fatal(err)
			}
			before := workflowState(t, ctx, s)
			if _, err := s.UpdateIssue(ctx, UpdateIssueRequest{Path: selector.Path, NativeID: selector.NativeID, Actor: "editor", ExpectedRevision: updated.Issue.Revision, Title: &title}); !errors.Is(err, storage.ErrValidation) || !reflect.DeepEqual(before, workflowState(t, ctx, s)) {
				t.Fatalf("ambiguous abbreviation changed Issue: %v", err)
			}
			if _, err := s.Create(ctx, CreateRequest{Path: "beads/a3f8", Title: "Colliding Memory"}); err != nil {
				t.Fatal(err)
			}
			before = workflowState(t, ctx, s)
			if _, err := s.ShowIssueSelected(ctx, selector); !errors.Is(err, storage.ErrValidation) || !reflect.DeepEqual(before, workflowState(t, ctx, s)) {
				t.Fatalf("bare graph-path collision selected an Issue: %v", err)
			}
			explicit, err := s.ShowIssueSelected(ctx, IssueSelector{Path: "beads/source"})
			if err != nil || explicit.ID != created.ID || explicit.Properties.Title != title {
				t.Fatalf("explicit path did not stay exact: %+v %v", explicit, err)
			}
		})
	}
}

func TestPositionalDependencyResolvesNativeAbbreviationsInWrite(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, options := issueExperimentOptions(t, backend)
			s, err := OpenExisting(ctx, options)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = s.Close() })
			create := func(path, id string) IssueRecord {
				t.Helper()
				request := plainIssue(id)
				request.Issue.ID = id
				issue, err := s.CreateIssue(ctx, path, request)
				if err != nil {
					t.Fatal(err)
				}
				return issue
			}
			source := create("beads/work", "exp-a3f8e9")
			target := create("beads/need", "exp-b7c4d2")
			request := DependencyRequest{SourcePath: "beads/a3f8", SourceNativeID: "a3f8", TargetPath: "beads/b7c4", TargetNativeID: "b7c4", Actor: "tester"}
			linked, err := s.AddDependency(ctx, request)
			if err != nil || !linked.Changed || linked.Source.Properties.ID != source.Properties.ID || linked.Target.Properties.ID != target.Properties.ID {
				t.Fatalf("abbreviated dependency: %+v %v", linked, err)
			}
			if linked.Link.Source != source.ID || linked.Link.Target != target.ID {
				t.Fatalf("wrong endpoint allocation: %+v", linked.Link)
			}
			repeat, err := s.AddDependency(ctx, request)
			if err != nil || repeat.Changed || repeat.Link.ID != linked.Link.ID {
				t.Fatalf("abbreviation repeat changed link: %+v %v", repeat, err)
			}
			if _, err := s.Create(ctx, CreateRequest{Path: "beads/a3f8", Title: "Colliding Memory", Body: "not an Issue"}); err != nil {
				t.Fatal(err)
			}
			if _, err := s.AddDependency(ctx, request); !errors.Is(err, storage.ErrValidation) {
				t.Fatalf("bare graph-path collision: %v", err)
			}
			request.SourcePath, request.SourceNativeID = "beads/work", ""
			create("beads/other", "exp-b7c4e1")
			if _, err := s.AddDependency(ctx, request); !errors.Is(err, storage.ErrValidation) {
				t.Fatalf("ambiguous target abbreviation: %v", err)
			}
			exact := create("beads/exact", "exp-b7c4")
			resolved, err := s.AddDependency(ctx, request)
			if err != nil || resolved.Link.Target != exact.ID {
				t.Fatalf("exact hash must win over abbreviations: %+v %v", resolved, err)
			}
			if _, err := s.AddDependency(ctx, DependencyRequest{SourcePath: "beads/work", TargetPath: "beads/need", Actor: "tester"}); err != nil {
				t.Fatalf("explicit paths must remain exact: %v", err)
			}
		})
	}
}
