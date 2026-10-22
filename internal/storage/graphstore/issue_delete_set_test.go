//go:build cgo

package graphstore

import (
	"database/sql"
	"errors"
	"reflect"
	"testing"

	"github.com/steveyegge/beads/internal/storage"
	publicops "github.com/steveyegge/beads/issueops"
)

func TestIssueDeleteSetPlanUsesOneBoundedGraphSnapshot(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			f := setupDependencyUnlink(t, backend, 2)
			if _, err := f.store.Create(f.ctx, CreateRequest{Path: "beads/memory", Title: "Not an Issue", Body: "Memory"}); err != nil {
				t.Fatal(err)
			}
			before := workflowState(t, f.ctx, f.store)
			plan := func(paths []string, cascade bool) (issueDeleteSetPlan, error) {
				var got issueDeleteSetPlan
				err := f.store.withTx(f.ctx, false, func(tx *sql.Tx) error {
					var err error
					got, err = f.store.planIssueDeleteSetInTx(f.ctx, tx, paths, cascade)
					return err
				})
				return got, err
			}
			flat, err := plan([]string{"beads/first", "beads/first"}, false)
			if err != nil || len(flat.Issues) != 1 || flat.Issues[0].ID != f.targets[0].ID ||
				len(flat.Links) != 1 || flat.Links[0].ID != f.links[0].ID || len(flat.Native.All) != 1 {
				t.Fatalf("flat plan: %+v %v", flat, err)
			}
			cascade, err := plan([]string{"beads/first"}, true)
			if err != nil || len(cascade.Issues) != 2 || cascade.Issues[0].ID != f.targets[0].ID ||
				cascade.Issues[1].ID != f.source.ID || len(cascade.Links) != 2 ||
				cascade.Links[0].ID != f.links[0].ID || cascade.Links[1].ID != f.links[1].ID ||
				len(cascade.Native.All) != 2 {
				t.Fatalf("cascade plan: %+v %v", cascade, err)
			}
			if _, err := plan([]string{"beads/first", "beads/missing"}, true); !errors.Is(err, ErrNotFound) {
				t.Fatalf("missing named Issue: %v", err)
			}
			if _, err := plan([]string{"beads/first", "beads/memory"}, false); !errors.Is(err, storage.ErrValidation) {
				t.Fatalf("non-Issue member was not refused before writing: %v", err)
			}
			oversize := make([]string, PreviewSnapshotLimit+1)
			for i := range oversize {
				oversize[i] = "beads/first"
			}
			if _, err := plan(oversize, false); !errors.Is(err, ErrLimitExceeded) {
				t.Fatalf("oversize named set was not refused: %v", err)
			}
			if !reflect.DeepEqual(before, workflowState(t, f.ctx, f.store)) {
				t.Fatal("read-only plan changed graph or native state")
			}
		})
	}
}

func TestIssueDeleteSetComposesNativeAndGraphWrites(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			f := setupDependencyUnlink(t, backend, 2)
			before := workflowState(t, f.ctx, f.store)
			request := IssueDeleteSetRequest{Paths: []string{"beads/first"}, Unconditional: true, Actor: "deleter"}
			unnamedGuard := request
			unnamedGuard.ExpectedRevisions = map[string]string{"beads/second": f.targets[1].Revision}
			if got, err := f.store.DeleteIssueSetWithIncidentLinks(f.ctx, unnamedGuard); !errors.Is(err, storage.ErrValidation) || !reflect.ValueOf(got).IsZero() {
				t.Fatalf("unnamed revision guard was ignored: %+v %v", got, err)
			}
			missing := request
			missing.Paths = []string{"beads/first", "beads/missing"}
			if got, err := f.store.DeleteIssueSetWithIncidentLinks(f.ctx, missing); !errors.Is(err, ErrNotFound) || !reflect.ValueOf(got).IsZero() ||
				!reflect.DeepEqual(before, workflowState(t, f.ctx, f.store)) {
				t.Fatalf("missing selector changed earlier Issue: %+v %v", got, err)
			}
			if got, err := f.store.DeleteIssueSetWithIncidentLinks(f.ctx, request); !errors.Is(err, publicops.ErrDependentsOutsideRequest) || !reflect.ValueOf(got).IsZero() {
				t.Fatalf("unforced outside dependent: %+v %v", got, err)
			}
			request.Force, request.Preview = true, true
			preview, err := f.store.DeleteIssueSetWithIncidentLinks(f.ctx, request)
			if err != nil || !preview.Preview || preview.Deleted || len(preview.Issues) != 1 || len(preview.IncidentLinkIDs) != 1 ||
				!reflect.DeepEqual(before, workflowState(t, f.ctx, f.store)) {
				t.Fatalf("force preview: %+v %v", preview, err)
			}
			request.Preview = false
			fault := errors.New("injected set delete failure")
			f.store.afterWrite = func(stage string) error {
				if stage == "issue-delete-native" {
					return fault
				}
				return nil
			}
			if got, err := f.store.DeleteIssueSetWithIncidentLinks(f.ctx, request); !errors.Is(err, fault) || !reflect.ValueOf(got).IsZero() {
				t.Fatalf("native-stage rollback: %+v %v", got, err)
			}
			f.store.afterWrite = nil
			if !reflect.DeepEqual(before, workflowState(t, f.ctx, f.store)) {
				t.Fatal("native-stage failure changed state")
			}
			deleted, err := f.store.DeleteIssueSetWithIncidentLinks(f.ctx, request)
			if err != nil || !deleted.Deleted || len(deleted.Issues) != 1 || deleted.Issues[0].Terminal == nil ||
				deleted.Issues[0].Terminal.PreviousVersion != deleted.Issues[0].Issue.Revision {
				t.Fatalf("forced set delete: %+v %v", deleted, err)
			}
			if _, err := f.store.ShowIssue(f.ctx, "beads/first"); !errors.Is(err, ErrGone) {
				t.Fatalf("deleted target: %v", err)
			}
			if current, err := f.store.ShowIssue(f.ctx, "beads/source"); err != nil || current.Revision == f.source.Revision {
				t.Fatalf("surviving dependent: %+v %v", current, err)
			}
		})
	}
}

func TestIssueDeleteSetResolvesEverySelectorBeforePlanning(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, _, s := memoryDeleteStore(t, backend)
			firstRequest := plainIssue("First")
			firstRequest.Issue.ID = "exp-a3f8e9"
			_, err := s.CreateIssue(ctx, "beads/first", firstRequest)
			if err != nil {
				t.Fatal(err)
			}
			secondRequest := plainIssue("Second")
			secondRequest.Issue.ID = "exp-b7c4d2"
			if _, err := s.CreateIssue(ctx, "beads/second", secondRequest); err != nil {
				t.Fatal(err)
			}
			if _, err := s.Create(ctx, CreateRequest{Path: "beads/context", Title: "Context"}); err != nil {
				t.Fatal(err)
			}
			link, err := s.AddInformationalLink(ctx, LinkCreateRequest{
				Path: "links/context", SourcePath: "beads/first", TargetPath: "beads/context", TypeURL: RelatedTypeURL(s.ScopeURL()),
			})
			if err != nil {
				t.Fatal(err)
			}
			request := IssueDeleteSetRequest{
				Selectors:     []IssueSelector{{Path: "beads/a3f8", NativeID: "a3f8"}, {Path: "beads/second"}},
				Unconditional: true, Actor: "deleter", Preview: true,
			}
			before := workflowState(t, ctx, s)
			preview, err := s.DeleteIssueSetWithIncidentLinks(ctx, request)
			if err != nil || !preview.Preview || preview.Deleted || len(preview.Issues) != 2 ||
				!reflect.DeepEqual(preview.IncidentLinkIDs, []string{link.Link.ID}) || !reflect.DeepEqual(before, workflowState(t, ctx, s)) {
				t.Fatalf("resolved preview: %+v %v", preview, err)
			}
			missing := request
			missing.Selectors = append(append([]IssueSelector{}, request.Selectors...), IssueSelector{Path: "beads/missing", NativeID: "missing"})
			missing.Preview = false
			if got, err := s.DeleteIssueSetWithIncidentLinks(ctx, missing); !errors.Is(err, ErrNotFound) || !reflect.ValueOf(got).IsZero() ||
				!reflect.DeepEqual(before, workflowState(t, ctx, s)) {
				t.Fatalf("missing set member changed state: %+v %v", got, err)
			}
			stale := request
			stale.Preview = false
			stale.Unconditional = false
			second, err := s.ShowIssue(ctx, "beads/second")
			if err != nil {
				t.Fatal(err)
			}
			stale.ExpectedRevisions = map[string]string{"beads/a3f8": "stale", "beads/second": second.Revision}
			if got, err := s.DeleteIssueSetWithIncidentLinks(ctx, stale); !errors.Is(err, ErrConflict) || !reflect.ValueOf(got).IsZero() ||
				!reflect.DeepEqual(before, workflowState(t, ctx, s)) {
				t.Fatalf("stale set member changed state: %+v %v", got, err)
			}
			otherRequest := plainIssue("Other")
			otherRequest.Issue.ID = "exp-a3f8f0"
			if _, err := s.CreateIssue(ctx, "beads/other", otherRequest); err != nil {
				t.Fatal(err)
			}
			before = workflowState(t, ctx, s)
			if got, err := s.DeleteIssueSetWithIncidentLinks(ctx, missing); !errors.Is(err, storage.ErrValidation) || !reflect.ValueOf(got).IsZero() ||
				!reflect.DeepEqual(before, workflowState(t, ctx, s)) {
				t.Fatalf("ambiguous set member changed state: %+v %v", got, err)
			}
			explicit := request
			explicit.Selectors = []IssueSelector{{Path: "beads/a3f8e"}}
			if got, err := s.DeleteIssueSetWithIncidentLinks(ctx, explicit); !errors.Is(err, ErrNotFound) || !reflect.ValueOf(got).IsZero() ||
				!reflect.DeepEqual(before, workflowState(t, ctx, s)) {
				t.Fatalf("explicit path resolved an abbreviation or changed state: %+v %v", got, err)
			}
			aliases := request
			aliases.Selectors = []IssueSelector{{Path: "beads/a3f8e", NativeID: "a3f8e"}, {Path: "beads/first"}}
			aliases.ExpectedRevisions = map[string]string{"beads/a3f8e": "first-guard", "beads/first": "second-guard"}
			if got, err := s.DeleteIssueSetWithIncidentLinks(ctx, aliases); !errors.Is(err, storage.ErrValidation) || !reflect.ValueOf(got).IsZero() ||
				!reflect.DeepEqual(before, workflowState(t, ctx, s)) {
				t.Fatalf("conflicting guards for aliases changed state: %+v %v", got, err)
			}
			request.Selectors[0] = IssueSelector{Path: "beads/a3f8e", NativeID: "a3f8e"}
			request.Preview = false
			deleted, err := s.DeleteIssueSetWithIncidentLinks(ctx, request)
			if err != nil || !deleted.Deleted || len(deleted.Issues) != 2 || len(deleted.IncidentLinkIDs) != 1 {
				t.Fatalf("resolved set delete: %+v %v", deleted, err)
			}
			for _, path := range []string{"beads/first", "beads/second"} {
				if _, err := s.ShowIssue(ctx, path); !errors.Is(err, ErrGone) {
					t.Fatalf("%s was not deleted: %v", path, err)
				}
			}
		})
	}
}

func TestIssueDeleteSetNamedDependentPair(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			f := setupDependencyUnlink(t, backend, 2)
			result, err := f.store.DeleteIssueSetWithIncidentLinks(f.ctx, IssueDeleteSetRequest{
				Paths: []string{"beads/source", "beads/first"}, Unconditional: true, Actor: "deleter",
			})
			if err != nil || !result.Deleted || len(result.Issues) != 2 || len(result.IncidentLinkIDs) != 2 {
				t.Fatalf("named pair: %+v %v", result, err)
			}
			for _, path := range []string{"beads/first", "beads/source"} {
				if _, err := f.store.ShowIssue(f.ctx, path); !errors.Is(err, ErrGone) {
					t.Fatalf("named pair left %s live: %v", path, err)
				}
			}
		})
	}
}

func TestIssueDeleteSetCascadeAndSurvivingReference(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			f := setupDependencyUnlink(t, backend, 2)
			before := workflowState(t, f.ctx, f.store)
			request := IssueDeleteSetRequest{Paths: []string{"beads/first"}, Cascade: true, Unconditional: true, Actor: "deleter"}
			fault := errors.New("injected terminal failure")
			f.store.afterWrite = func(stage string) error {
				if stage == "issue-delete-retained" {
					return fault
				}
				return nil
			}
			if got, err := f.store.DeleteIssueSetWithIncidentLinks(f.ctx, request); !errors.Is(err, fault) || !reflect.ValueOf(got).IsZero() {
				t.Fatalf("terminal rollback: %+v %v", got, err)
			}
			f.store.afterWrite = nil
			if !reflect.DeepEqual(before, workflowState(t, f.ctx, f.store)) {
				t.Fatal("cascade rollback changed state")
			}
			deleted, err := f.store.DeleteIssueSetWithIncidentLinks(f.ctx, request)
			if err != nil || !deleted.Deleted || len(deleted.Issues) != 2 || len(deleted.IncidentLinkIDs) != 2 {
				t.Fatalf("cascade: %+v %v", deleted, err)
			}
			for _, path := range []string{"beads/first", "beads/source"} {
				if _, err := f.store.ShowIssue(f.ctx, path); !errors.Is(err, ErrGone) {
					t.Fatalf("cascade left %s live: %v", path, err)
				}
			}
			if got, err := f.store.ShowIssue(f.ctx, "beads/second"); err != nil || got.ID != f.targets[1].ID {
				t.Fatalf("cascade removed independent target: %+v %v", got, err)
			}
		})
	}
}

func TestIssueDeleteSetVersionsRewrittenNeighbor(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			f := setupDependencyUnlink(t, backend, 1)
			citation := "see " + f.targets[0].Properties.ID + " for context"
			edited, err := f.store.UpdateIssue(f.ctx, UpdateIssueRequest{Path: "beads/source", Unconditional: true, Actor: "writer", Description: issueEditString(citation)})
			if err != nil {
				t.Fatal(err)
			}
			beforeKind, beforeVersions, err := f.store.Versions(f.ctx, "beads/source")
			if err != nil || beforeKind != KindIssue {
				t.Fatalf("before versions: %s %+v %v", beforeKind, beforeVersions, err)
			}
			result, err := f.store.DeleteIssueSetWithIncidentLinks(f.ctx, IssueDeleteSetRequest{Paths: []string{"beads/first"}, Force: true, Unconditional: true, Actor: "deleter"})
			if err != nil || !result.Deleted {
				t.Fatalf("delete with citation: %+v %v", result, err)
			}
			current, err := f.store.ShowIssue(f.ctx, "beads/source")
			if err != nil || current.Properties.Description != "see [deleted:"+f.targets[0].Properties.ID+"] for context" || current.Revision == edited.Issue.Revision {
				t.Fatalf("rewritten neighbor: %+v %v", current, err)
			}
			kind, versions, err := f.store.Versions(f.ctx, "beads/source")
			if err != nil || kind != KindIssue || len(versions) <= len(beforeVersions) || versions[0].Version != current.Revision {
				t.Fatalf("retained neighbor: %s %+v %v", kind, versions, err)
			}
		})
	}
}
