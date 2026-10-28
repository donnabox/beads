//go:build cgo

package graphstore

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage/issueops"
	"github.com/steveyegge/beads/internal/types"
	publicops "github.com/steveyegge/beads/issueops"
)

func TestReparentIssueAtomicNativeAndGraphLinks(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			_, options := issueExperimentOptions(t, backend)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			t.Cleanup(cancel)
			ctx = issueops.WithEventsJournal(ctx, true)
			store, err := OpenExisting(ctx, options)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = store.Close() })
			create := func(path, title string) IssueRecord {
				t.Helper()
				issue, err := store.CreateIssue(ctx, path, publicops.CreateRequest{Actor: "author", Issue: &types.Issue{
					Title: title, Status: types.StatusOpen, Priority: 2, IssueType: types.TypeTask,
				}})
				if err != nil {
					t.Fatal(err)
				}
				return issue
			}
			create("beads/old-parent", "Old parent")
			newParent := create("beads/new-parent", "New parent")
			child := create("beads/child", "Child")
			old, err := store.AddDependency(ctx, DependencyRequest{Path: "links/old-parent", Type: types.DepParentChild,
				SourcePath: "beads/child", TargetPath: "beads/old-parent", Actor: "author"})
			if err != nil {
				t.Fatal(err)
			}
			child = old.Source
			newTitle, newPriority := "Child after move", 1
			request := IssueReparentRequest{Path: "beads/child", ParentPath: "beads/new-parent", Actor: "editor", ExpectedRevision: child.Revision,
				Edit: UpdateIssueRequest{Title: &newTitle, Priority: &newPriority}}
			before := workflowState(t, ctx, store)
			stale := request
			stale.ExpectedRevision = "00000000000000000000000000000000"
			if _, err := store.ReparentIssue(ctx, stale); !errors.Is(err, ErrConflict) || !reflect.DeepEqual(workflowState(t, ctx, store), before) {
				t.Fatalf("stale revision changed native or graph state: %v", err)
			}
			fault := errors.New("injected reparent failure")
			for _, stage := range []string{"issue-reparent", "reparent-old-links", "reparent-new-link", "reparent-link-retained", "reparent-issue-retained"} {
				store.afterWrite = func(at string) error {
					if at == stage {
						return fault
					}
					return nil
				}
				_, err := store.ReparentIssue(ctx, request)
				store.afterWrite = nil
				if !errors.Is(err, fault) || !reflect.DeepEqual(workflowState(t, ctx, store), before) {
					t.Fatalf("failure at %s left native or graph effects: %v", stage, err)
				}
			}
			result, err := store.ReparentIssue(ctx, request)
			if err != nil {
				t.Fatal(err)
			}
			if !result.Changed || len(result.Removed) != 1 || result.Added == nil || result.Issue.Revision == child.Revision ||
				result.Removed[0].ID != old.Link.ID || result.Added.Source != child.ID || result.Added.Target != newParent.ID ||
				len(result.Issue.Owned) != 1 || result.Issue.Properties.Title != newTitle || result.Issue.Properties.Priority != newPriority {
				t.Fatalf("incoherent reparent result: %+v", result)
			}
			if _, err := store.ShowLink(ctx, "links/old-parent"); !errors.Is(err, ErrGone) {
				t.Fatalf("old Link remained live: %v", err)
			}
			localLinkID := strings.TrimPrefix(result.Added.ID, options.Binding.ScopeURL)
			if localLinkID == result.Added.ID || len(localLinkID) != graph.GeneratedResourceIDInitialLength || strings.ContainsAny(localLinkID, "/-.") {
				t.Fatalf("replacement Link ID is not a short rootless allocation: %q", result.Added.ID)
			}
			allocated, err := store.ResolveCanonicalPublicResourceIdentity(ctx, result.Added.ID)
			if err != nil {
				t.Fatal(err)
			}
			newPath := allocated.PrivatePath
			if got, err := store.ShowLink(ctx, newPath); err != nil || !reflect.DeepEqual(got, *result.Added) {
				t.Fatalf("new Link is not current: %+v %v", got, err)
			}
			reserved := workflowState(t, ctx, store)
			if _, err := store.Create(ctx, CreateRequest{PublicID: localLinkID, Title: "Must not reuse Link ID"}); !errors.Is(err, ErrAlreadyExists) ||
				!reflect.DeepEqual(workflowState(t, ctx, store), reserved) {
				t.Fatalf("replacement Link did not reserve its public ID across kinds: %v", err)
			}
			var target string
			if err := store.db.QueryRowContext(ctx, `SELECT depends_on_issue_id FROM dependencies WHERE issue_id=? AND type=?`, child.Properties.ID, string(types.DepParentChild)).Scan(&target); err != nil || target != newParent.Properties.ID {
				t.Fatalf("native parent edge = %q, %v", target, err)
			}
			if kids, err := store.ListChildren(ctx, "beads/old-parent"); err != nil || len(kids) != 0 {
				t.Fatalf("old parent still has child: %+v %v", kids, err)
			}
			if kids, err := store.ListChildren(ctx, "beads/new-parent"); err != nil || len(kids) != 1 || kids[0].ID != child.ID {
				t.Fatalf("new parent has wrong children: %+v %v", kids, err)
			}
			noop := request
			noop.Path, noop.NativeID = "beads/"+child.Properties.ID, child.Properties.ID
			noop.ParentPath, noop.ParentNativeID = "beads/"+newParent.Properties.ID, newParent.Properties.ID
			noop.ExpectedRevision = result.Issue.Revision
			after := workflowState(t, ctx, store)
			if again, err := store.ReparentIssue(ctx, noop); err != nil || again.Changed || !reflect.DeepEqual(workflowState(t, ctx, store), after) {
				t.Fatalf("same-parent update was not a no-op: %+v %v", again, err)
			}
			// An edit against an already selected parent still goes through the
			// same checked Issue writer, without replacing the parent Link.
			description := "Edit after move"
			sameParentEdit := noop
			sameParentEdit.Edit = UpdateIssueRequest{Description: &description}
			edited, err := store.ReparentIssue(ctx, sameParentEdit)
			if err != nil || !edited.Changed || len(edited.Removed) != 0 || edited.Added != nil ||
				edited.Issue.Properties.Description != description || edited.Issue.Revision == result.Issue.Revision ||
				len(edited.Issue.Owned) != 1 {
				t.Fatalf("same-parent scalar edit changed Link identity or failed: %+v %v", edited, err)
			}
			if current, err := store.ShowLink(ctx, newPath); err != nil || current.ID != result.Added.ID {
				t.Fatalf("same-parent scalar edit replaced the parent Link: %+v %v", current, err)
			}
			cycleState := workflowState(t, ctx, store)
			if _, err := store.ReparentIssue(ctx, IssueReparentRequest{
				Path: "beads/new-parent", ParentPath: "beads/child", Actor: "editor", ExpectedRevision: newParent.Revision,
			}); err == nil || !reflect.DeepEqual(workflowState(t, ctx, store), cycleState) {
				t.Fatalf("parent cycle was admitted or changed state: %v", err)
			}
			orphan := create("beads/orphan", "Orphan")
			attached, err := store.ReparentIssue(ctx, IssueReparentRequest{
				Path: "beads/orphan", ParentPath: "beads/new-parent", Actor: "editor", ExpectedRevision: orphan.Revision,
			})
			if err != nil || !attached.Changed || len(attached.Removed) != 0 || attached.Added == nil ||
				attached.Added.Source != orphan.ID || attached.Added.Target != newParent.ID {
				t.Fatalf("unparented Issue could not attach atomically: %+v %v", attached, err)
			}
			movedBack, err := store.ReparentIssue(ctx, IssueReparentRequest{
				Path: "beads/child", ParentPath: "beads/old-parent", Actor: "editor", ExpectedRevision: edited.Issue.Revision,
			})
			if err != nil || !movedBack.Changed || len(movedBack.Removed) != 1 || movedBack.Removed[0].ID != result.Added.ID {
				t.Fatalf("replacement Link was not tombstoned by next reparent: %+v %v", movedBack, err)
			}
			reserved = workflowState(t, ctx, store)
			if _, err := store.Create(ctx, CreateRequest{PublicID: localLinkID, Title: "Must not reuse tombstoned Link ID"}); !errors.Is(err, ErrAlreadyExists) ||
				!reflect.DeepEqual(workflowState(t, ctx, store), reserved) {
				t.Fatalf("tombstoned replacement Link lost lifetime public ID: %v", err)
			}
		})
	}
}
