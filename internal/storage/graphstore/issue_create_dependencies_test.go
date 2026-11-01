//go:build cgo

package graphstore

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/types"
)

func TestCreateIssueWithDependenciesMapsBothDirectionsAtomically(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, options := issueExperimentOptions(t, backend)
			store, err := OpenExisting(ctx, options)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = store.Close() })
			prereq, err := store.CreateIssue(ctx, "beads/prereq", plainIssue("Prerequisite"))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := store.CreateIssue(ctx, "beads/consumer", plainIssue("Consumer")); err != nil {
				t.Fatal(err)
			}
			create := plainIssue("New")
			created, err := store.CreateIssueWithDependencies(ctx, "beads/new", create, []IssueCreateDependency{
				{Target: IssueSelector{Path: "beads/prereq"}, Type: types.DepBlocks},
				{Target: IssueSelector{Path: "beads/consumer"}, Type: types.DepBlocks, Reverse: true},
			})
			if err != nil {
				t.Fatal(err)
			}
			consumer, err := store.ShowIssue(ctx, "beads/consumer")
			if err != nil {
				t.Fatal(err)
			}
			if len(created.Owned) != 1 || len(consumer.Owned) != 1 || created.Attribution.Inception != "create" {
				t.Fatalf("missing source-owned Links or create History: new=%+v consumer=%+v", created, consumer)
			}
			for _, check := range []struct {
				owned  json.RawMessage
				source string
				target string
			}{
				{created.Owned[0], created.ID, prereq.ID},
				{consumer.Owned[0], consumer.ID, created.ID},
			} {
				var link LinkRecord
				if err := json.Unmarshal(check.owned, &link); err != nil {
					t.Fatal(err)
				}
				if link.Source != check.source || link.Target != check.target || link.Properties["dep_type"] != string(types.DepBlocks) {
					t.Fatalf("incorrect canonical dependency Link: %+v", link)
				}
			}
			before := workflowState(t, ctx, store)
			fault := errors.New("injected Link retention failure")
			store.afterWrite = func(stage string) error {
				if stage == "create-dependency-retained" {
					return fault
				}
				return nil
			}
			_, err = store.CreateIssueWithDependencies(ctx, "beads/rolled-back", plainIssue("Rollback"), []IssueCreateDependency{
				{Target: IssueSelector{Path: "beads/prereq"}, Type: types.DepBlocks},
			})
			store.afterWrite = nil
			if !errors.Is(err, fault) || !reflect.DeepEqual(before, workflowState(t, ctx, store)) {
				t.Fatalf("failed graph Link mapping left native create state: %v", err)
			}
			_, err = store.CreateIssueWithDependencies(ctx, "beads/cycle", plainIssue("Cycle"), []IssueCreateDependency{
				{Target: IssueSelector{Path: "beads/prereq"}, Type: types.DepBlocks},
				{Target: IssueSelector{Path: "beads/prereq"}, Type: types.DepBlocks, Reverse: true},
			})
			if !errors.Is(err, storage.ErrValidation) || !reflect.DeepEqual(before, workflowState(t, ctx, store)) {
				t.Fatalf("cycle refusal left native or graph create state: %v", err)
			}

			createNative := func(path, id string) IssueRecord {
				t.Helper()
				request := plainIssue(id)
				request.Issue.ID = id
				record, err := store.CreateIssue(ctx, path, request)
				if err != nil {
					t.Fatal(err)
				}
				return record
			}
			selected := createNative("beads/abbreviated-target", "exp-a3f8e9")
			for _, tc := range []struct {
				name, operand string
			}{
				{"bare", "a3f8"},
				{"configured-prefix", "exp-a3f8"},
			} {
				request := plainIssue(tc.name)
				request.Issue.ID = "exp-created-" + tc.name
				created, err := store.CreateIssueWithDependencies(ctx, "beads/created-"+tc.name, request, []IssueCreateDependency{
					{Target: IssueSelector{Path: "beads/" + tc.operand, NativeID: tc.operand}, Type: types.DepBlocks},
				})
				if err != nil || len(created.Owned) != 1 || created.Attribution.Inception != "create" {
					t.Fatalf("%s abbreviation create: %+v, %v", tc.name, created, err)
				}
				var link LinkRecord
				if err := json.Unmarshal(created.Owned[0], &link); err != nil {
					t.Fatal(err)
				}
				if link.Source != created.ID || link.Target != selected.ID || link.Properties["dep_type"] != string(types.DepBlocks) {
					t.Fatalf("%s selected wrong dependency: %+v", tc.name, link)
				}
			}
			createNative("beads/ambiguous-target", "exp-a3f8e1")
			refuse := func(name string, selector IssueSelector, want error) {
				t.Helper()
				before := workflowState(t, ctx, store)
				request := plainIssue(name)
				request.Issue.ID = "exp-refused-" + name
				_, err := store.CreateIssueWithDependencies(ctx, "beads/refused-"+name, request, []IssueCreateDependency{
					// A valid reverse edge precedes the failing target. No new Issue,
					// existing owner revision, native edge or retained row may survive.
					{Target: IssueSelector{Path: "beads/prereq"}, Type: types.DepBlocks, Reverse: true},
					{Target: selector, Type: types.DepBlocks},
				})
				if !errors.Is(err, want) || !reflect.DeepEqual(before, workflowState(t, ctx, store)) {
					t.Fatalf("%s refusal changed complete create state or error: %v", name, err)
				}
			}
			refuse("ambiguous", IssueSelector{Path: "beads/a3f8", NativeID: "a3f8"}, storage.ErrValidation)
			// This is the exact selector that a canonical URL produces. It must
			// not turn the shortened URL into an abbreviation after the CLI fix.
			refuse("exact-short", IssueSelector{Path: "beads/a3f8", NativeID: "a3f8", Exact: true}, ErrNotFound)
			exact := createNative("beads/exact-target", "exp-a3f8")
			request := plainIssue("Exact native ID")
			request.Issue.ID = "exp-exact-source"
			exactCreated, err := store.CreateIssueWithDependencies(ctx, "beads/exact-source", request, []IssueCreateDependency{
				{Target: IssueSelector{Path: "beads/exp-a3f8", NativeID: "exp-a3f8"}, Type: types.DepBlocks},
			})
			if err != nil || len(exactCreated.Owned) != 1 {
				t.Fatalf("exact ID did not beat ambiguous longer IDs: %+v, %v", exactCreated, err)
			}
			var exactLink LinkRecord
			if err := json.Unmarshal(exactCreated.Owned[0], &exactLink); err != nil {
				t.Fatal(err)
			}
			if exactLink.Target != exact.ID || exactLink.Source != exactCreated.ID {
				t.Fatalf("exact ID selected another Issue: %+v", exactLink)
			}
			memory, err := store.Create(ctx, CreateRequest{Path: "beads/fence-memory", PublicID: "a3f8", Title: "Exact Memory", Actor: "tester"})
			if err != nil {
				t.Fatal(err)
			}
			refuse("memory", IssueSelector{Path: "beads/a3f8", NativeID: "a3f8"}, storage.ErrValidation)
			createNative("beads/link-decoy", "exp-b7c4d2")
			if _, err := store.AddInformationalLink(ctx, LinkCreateRequest{
				PublicID: "b7c4", Path: "links/fence", TypeURL: store.ScopeURL() + "references",
				SourcePath: "beads/fence-memory", TargetPath: "beads/prereq", Actor: "tester", ExpectedSourceRevision: memory.Revision,
			}); err != nil {
				t.Fatal(err)
			}
			refuse("link", IssueSelector{Path: "beads/b7c4", NativeID: "b7c4"}, storage.ErrValidation)
			dead := createNative("beads/dead-target", "exp-d3ad")
			createNative("beads/dead-decoy", "exp-d3adbeef")
			if _, err := store.DeleteIssue(ctx, IssueDeleteRequest{Path: "beads/dead-target", Actor: "tester", ExpectedRevision: dead.Revision}); err != nil {
				t.Fatal(err)
			}
			refuse("tombstone", IssueSelector{Path: "beads/exp-d3ad", NativeID: "exp-d3ad"}, ErrGone)
		})
	}
}
