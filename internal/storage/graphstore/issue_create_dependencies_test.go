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
			abbreviatedTarget := plainIssue("Abbreviated target")
			abbreviatedTarget.Issue.ID = "exp-a3f8e9"
			target, err := store.CreateIssue(ctx, "beads/abbreviated-target", abbreviatedTarget)
			if err != nil {
				t.Fatal(err)
			}
			partial := IssueCreateDependency{Target: IssueSelector{Path: "beads/a3f8", NativeID: "a3f8"}, Type: types.DepBlocks}
			createdByPrefix, err := store.CreateIssueWithDependencies(ctx, "beads/abbreviated-child", plainIssue("Abbreviated child"), []IssueCreateDependency{partial})
			if err != nil || len(createdByPrefix.Owned) != 1 {
				t.Fatalf("partial dependency target was not resolved in create transaction: %+v %v", createdByPrefix, err)
			}
			var linked LinkRecord
			if err := json.Unmarshal(createdByPrefix.Owned[0], &linked); err != nil || linked.Target != target.ID {
				t.Fatalf("partial dependency selected wrong target: %+v %v", linked, err)
			}
			ambiguousTarget := plainIssue("Second abbreviated target")
			ambiguousTarget.Issue.ID = "exp-a3f8f0"
			if _, err := store.CreateIssue(ctx, "beads/second-abbreviated-target", ambiguousTarget); err != nil {
				t.Fatal(err)
			}
			beforeAmbiguous := workflowState(t, ctx, store)
			if _, err := store.CreateIssueWithDependencies(ctx, "beads/ambiguous-child", plainIssue("Ambiguous child"), []IssueCreateDependency{partial}); !errors.Is(err, storage.ErrValidation) || !reflect.DeepEqual(beforeAmbiguous, workflowState(t, ctx, store)) {
				t.Fatalf("ambiguous abbreviation changed create state: %v", err)
			}
			explicit := IssueCreateDependency{Target: IssueSelector{Path: "beads/abbreviated-target"}, Type: types.DepBlocks}
			createdExplicit, err := store.CreateIssueWithDependencies(ctx, "beads/explicit-child", plainIssue("Explicit child"), []IssueCreateDependency{explicit})
			if err != nil || len(createdExplicit.Owned) != 1 {
				t.Fatalf("explicit target path should stay exact despite native abbreviation ambiguity: %+v %v", createdExplicit, err)
			}
		})
	}
}
