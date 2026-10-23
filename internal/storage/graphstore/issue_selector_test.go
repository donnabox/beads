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
	"github.com/steveyegge/beads/internal/workapi"
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
			mixed, err := s.ReadBeadSelected(ctx, selector)
			if err != nil || !reflect.DeepEqual(mixed, created) {
				t.Fatalf("abbreviated mixed read selected wrong Issue: %+v %v", mixed, err)
			}
			detail, err := s.ShowIssueDetailsSelected(ctx, selector, workapi.DetailOptions{})
			if err != nil || detail == nil || detail.ID != "beads/source" || detail.Revision != created.Revision || detail.Title != created.Properties.Title {
				t.Fatalf("abbreviated detail read selected wrong Issue: %+v %v", detail, err)
			}
			combined, combinedDetail, err := s.ReadBeadWithIssueDetailsSelected(ctx, selector, workapi.DetailOptions{})
			if err != nil || !reflect.DeepEqual(combined, created) || !reflect.DeepEqual(combinedDetail, detail) {
				t.Fatalf("combined abbreviated read changed either projection: %+v %+v %v", combined, combinedDetail, err)
			}
			memory, err := s.Create(ctx, CreateRequest{Path: "beads/notes", Title: "Notes"})
			if err != nil {
				t.Fatal(err)
			}
			mixed, err = s.ReadBeadSelected(ctx, IssueSelector{Path: "beads/notes", NativeID: "notes"})
			if err != nil || !reflect.DeepEqual(mixed, memory) {
				t.Fatalf("bare Memory fallback changed selection: %+v %v", mixed, err)
			}
			if _, err := s.ShowIssueDetailsSelected(ctx, IssueSelector{Path: "beads/notes"}, workapi.DetailOptions{}); !errors.Is(err, ErrCapabilityUnavailable) {
				t.Fatalf("explicit Memory detail read should refuse as non-Issue: %v", err)
			}
			combined, combinedDetail, err = s.ReadBeadWithIssueDetailsSelected(ctx, IssueSelector{Path: "beads/notes", NativeID: "notes"}, workapi.DetailOptions{})
			if err != nil || !reflect.DeepEqual(combined, memory) || combinedDetail != nil {
				t.Fatalf("combined Memory read fabricated Issue details: %+v %+v %v", combined, combinedDetail, err)
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
			if _, err := s.ReadBeadSelected(ctx, selector); !errors.Is(err, storage.ErrValidation) || !reflect.DeepEqual(before, workflowState(t, ctx, s)) {
				t.Fatalf("bare mixed read hid an Issue/Memory collision: %v", err)
			}
			if _, err := s.ShowIssueDetailsSelected(ctx, selector, workapi.DetailOptions{}); !errors.Is(err, storage.ErrValidation) || !reflect.DeepEqual(before, workflowState(t, ctx, s)) {
				t.Fatalf("bare detail read hid an Issue/Memory collision: %v", err)
			}
			if _, _, err := s.ReadBeadWithIssueDetailsSelected(ctx, selector, workapi.DetailOptions{}); !errors.Is(err, storage.ErrValidation) || !reflect.DeepEqual(before, workflowState(t, ctx, s)) {
				t.Fatalf("combined read hid an Issue/Memory collision: %v", err)
			}
			explicit, err := s.ShowIssueSelected(ctx, IssueSelector{Path: "beads/source"})
			if err != nil || explicit.ID != created.ID || explicit.Properties.Title != title {
				t.Fatalf("explicit path did not stay exact: %+v %v", explicit, err)
			}
			detail, err = s.ShowIssueDetailsSelected(ctx, IssueSelector{Path: "beads/source"}, workapi.DetailOptions{})
			if err != nil || detail == nil || detail.ID != "beads/source" || detail.Title != title {
				t.Fatalf("explicit detail path did not stay exact: %+v %v", detail, err)
			}
			mixed, err = s.ReadBeadSelected(ctx, IssueSelector{Path: "beads/a3f8"})
			if err != nil {
				t.Fatalf("explicit Memory path did not stay exact: %v", err)
			}
			if _, ok := mixed.(Record); !ok {
				t.Fatalf("explicit Memory path returned %T", mixed)
			}
		})
	}
}
func TestOpaqueResourceReadSelectorKeepsPublicIDsSeparateFromGraphPaths(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, options := issueExperimentOptions(t, backend)
			s, err := OpenExisting(ctx, options)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = s.Close() })
			literal := plainIssue("Literal")
			literal.Issue.ID = "exp-beads/literal"
			_, err = s.CreateIssue(ctx, "beads/allocation", literal)
			if err != nil {
				t.Fatal(err)
			}
			path, record, details, err := s.ReadResourceWithIssueDetailsSelected(ctx, "beads/exp-beads/literal", "exp-beads/literal", workapi.DetailOptions{})
			if read, ok := record.(IssueRecord); err != nil || path != "beads/allocation" || !ok || read.Properties.ID != literal.Issue.ID || details == nil || details.ID != "beads/allocation" {
				t.Fatalf("literal public ID read: path=%q record=%+v details=%+v err=%v", path, record, details, err)
			}
			pathOnly := plainIssue("Path only")
			pathOnly.Issue.ID = "exp-other-id"
			_, err = s.CreateIssue(ctx, "beads/path-only", pathOnly)
			if err != nil {
				t.Fatal(err)
			}
			if _, _, _, err := s.ReadResourceWithIssueDetailsSelected(ctx, "beads/beads/path-only", "beads/path-only", workapi.DetailOptions{}); !errors.Is(err, storage.ErrValidation) {
				t.Fatalf("local path spelling silently selected an unrelated Issue: %v", err)
			}
			path, record, details, err = s.ReadResourceWithIssueDetailsSelected(ctx, "beads/path-only", "", workapi.DetailOptions{})
			if read, ok := record.(IssueRecord); err != nil || path != "beads/path-only" || !ok || read.Properties.ID != pathOnly.Issue.ID || details == nil || details.ID != "beads/path-only" {
				t.Fatalf("canonical path read: path=%q record=%+v details=%+v err=%v", path, record, details, err)
			}
			memory, err := s.Create(ctx, CreateRequest{Path: "beads/notes", Title: "Notes"})
			if err != nil {
				t.Fatal(err)
			}
			path, record, details, err = s.ReadResourceWithIssueDetailsSelected(ctx, "beads/notes", "notes", workapi.DetailOptions{})
			if read, ok := record.(Record); err != nil || path != "beads/notes" || !ok || read.ID != memory.ID || details != nil {
				t.Fatalf("unambiguous Memory fallback: path=%q record=%+v details=%+v err=%v", path, record, details, err)
			}
			path, record, details, err = s.ReadResourceWithIssueDetailsSelected(ctx, "beads/beads/notes", "beads/notes", workapi.DetailOptions{})
			if read, ok := record.(Record); err != nil || path != "beads/notes" || !ok || read.ID != memory.ID || details != nil {
				t.Fatalf("exact Memory public spelling: path=%q record=%+v details=%+v err=%v", path, record, details, err)
			}
			link, err := s.AddDependency(ctx, DependencyRequest{Path: "links/edge", SourcePath: "beads/allocation", TargetPath: "beads/path-only", Actor: "author"})
			if err != nil {
				t.Fatal(err)
			}
			path, record, details, err = s.ReadResourceWithIssueDetailsSelected(ctx, "beads/edge", "edge", workapi.DetailOptions{})
			if read, ok := record.(LinkRecord); err != nil || path != "links/edge" || !ok || read.ID != link.Link.ID || details != nil {
				t.Fatalf("local Link ID read: path=%q record=%+v details=%+v err=%v", path, record, details, err)
			}
			path, record, details, err = s.ReadResourceWithIssueDetailsSelected(ctx, "beads/links/edge", "links/edge", workapi.DetailOptions{})
			if read, ok := record.(LinkRecord); err != nil || path != "links/edge" || !ok || read.ID != link.Link.ID || details != nil {
				t.Fatalf("exact Link public spelling: path=%q record=%+v details=%+v err=%v", path, record, details, err)
			}
			if _, err := s.Create(ctx, CreateRequest{Path: "beads/edge", Title: "Same public spelling"}); err != nil {
				t.Fatal(err)
			}
			if _, _, _, err := s.ReadResourceWithIssueDetailsSelected(ctx, "beads/edge", "edge", workapi.DetailOptions{}); !errors.Is(err, storage.ErrValidation) {
				t.Fatalf("Memory/Link collision silently chose one backing: %v", err)
			}
			if _, err := s.Unlink(ctx, LinkDeleteRequest{Path: "links/edge", Actor: "remover", ExpectedRevision: link.Link.Revision, ExpectedSourceRevision: link.Source.Revision}); err != nil {
				t.Fatal(err)
			}
			path, record, details, err = s.ReadResourceWithIssueDetailsSelected(ctx, "beads/edge", "edge", workapi.DetailOptions{})
			if read, ok := record.(Record); err != nil || path != "beads/edge" || !ok || read.Properties.Title != "Same public spelling" || details != nil {
				t.Fatalf("removed Link still shadowed live Memory: path=%q record=%+v details=%+v err=%v", path, record, details, err)
			}
		})
	}
}
