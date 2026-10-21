//go:build cgo

package main

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// Run each installed journey on its own disposable workspace. The server phase
// uses only the explicit test server port, never a user's workspace or service.
func TestGraphPreviewSelectedTypesWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			const scope = "https://example.invalid/selected-types/"
			work, home := graphListWorkspace(t, bd, engine, scope)
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			refuse := func(code string, args ...string) {
				t.Helper()
				graphPolicyCLI(t, bd, work, home, nil, code, append(args, "--json")...)
			}
			inventory := graphMixedResult[graphTypeInventory](t, call("types"))
			if len(inventory.BeadTypes) != 10 || len(inventory.LinkTypes) != 10 {
				t.Fatalf("selected catalog not installed: %+v", inventory)
			}
			for _, descriptor := range append(inventory.BeadTypes, inventory.LinkTypes...) {
				for _, absent := range []string{"types/preview-", "types/example-", "types/Type", "types/BeadType", "types/LinkType", "types/ConformsTo", "types/refines"} {
					if strings.Contains(string(descriptor), absent) {
						t.Fatalf("private/old Type exposed: %s", descriptor)
					}
				}
			}
			policy := graphMixedResult[graphstore.Record](t, call("create", "--type", "types/policy", "--id", "policy", "--properties", `{"title":"Policy","body":"Keep the body","extra":{"kept":[1,true,null]}}`))
			if policy.Type != scope+"types/policy" || !strings.Contains(call("show", policy.ID), `"kept"`) {
				t.Fatal("selected creation/show lost Type or authored extras")
			}
			memory := graphMixedResult[graphstore.Record](t, call("create", "--type", "types/memory", "--id", "memory", "--properties", `{"title":"Memory","body":"Original","extra":{"kept":true}}`))
			edited := graphMixedResult[graphstore.MemoryMutationResult](t, call("update", memory.ID, "--type", "types/requirement", "--properties", `{"body":"Changed","new":42}`, "--if-revision", memory.Revision))
			if !edited.Changed || edited.Memory.Type != scope+"types/requirement" || !strings.Contains(call("show", memory.ID), `"kept"`) {
				t.Fatal("combined retype/merge lost inherited complete properties")
			}
			retained := graphMixedResult[graphstore.Record](t, call("show", "memory@"+memory.Revision))
			if !reflect.DeepEqual(retained, memory) {
				t.Fatal("retype rewrote the original Type/payload/history")
			}
			refuse("revision_conflict", "update", memory.ID, "--type", "types/policy", "--if-revision", memory.Revision)
			patched := graphMixedResult[graphstore.MemoryMutationResult](t, call("update", memory.ID, "--type", "types/procedure", "--patch", `[{"op":"add","path":"/ordered","value":["one","two"]}]`, "--if-revision", edited.Memory.Revision))
			if patched.Memory.Type != scope+"types/procedure" || !strings.Contains(call("show", memory.ID), `"ordered"`) {
				t.Fatal("ordered properties and Type did not commit together")
			}
			typeOnly := graphMixedResult[graphstore.MemoryMutationResult](t, call("update", memory.ID, "--type", "types/requirement"))
			if !typeOnly.Changed || typeOnly.Memory.Type != scope+"types/requirement" || !strings.Contains(call("show", memory.ID), `"ordered"`) {
				t.Fatal("unguarded type-only update did not preserve complete properties")
			}
			var selected struct {
				Result struct{ Items []struct{ ID, Type string } }
			}
			if err := json.Unmarshal([]byte(call("list", "--type", "types/requirement", "--format", "records-json", "--all")), &selected); err != nil || len(selected.Result.Items) != 1 || selected.Result.Items[0].ID != memory.ID {
				t.Fatalf("selected Bead filter failed: %+v %v", selected, err)
			}
			constraint := graphMixedResult[graphstore.Record](t, call("create", "Constraint", "--type", "types/constraint", "--id", "constraint"))
			owned := graphMixedResult[graphstore.LinkMutationResult](t, call("link", "create", "policy", "constraint", "--id", "subject", "--type", "types/subject-to"))
			if owned.Link.Type != scope+"types/subject-to" || owned.ReplacedSource == nil {
				t.Fatal("selected Link creation did not use source ownership")
			}
			refuse("invalid_properties", "update", constraint.ID, "--type", "types/memory", "--if-revision", constraint.Revision)
			if got := graphMixedResult[graphstore.Record](t, call("show", constraint.ID)); !reflect.DeepEqual(got, constraint) {
				t.Fatal("incident-Link refusal changed the Bead")
			}
			links := graphMixedResult[[]graphstore.LinkRecord](t, call("link", "list", "--type", "types/subject-to"))
			if len(links) != 1 || links[0].ID != owned.Link.ID {
				t.Fatal("selected Link filter failed")
			}
			currentPolicy := graphMixedResult[graphstore.Record](t, call("show", policy.ID))
			subjects, err := json.Marshal(map[string]any{"title": "Evaluation", "subjects": []map[string]string{{"uri": policy.ID, "revision": currentPolicy.Revision}}})
			if err != nil {
				t.Fatal(err)
			}
			evaluation := graphMixedResult[graphstore.Record](t, call("create", "--id", "evaluation", "--type", "types/evaluation", "--properties", string(subjects)))
			claim := graphMixedResult[graphstore.LinkMutationResult](t, call("link", "create", "evaluation", "constraint", "--type", "types/satisfies", "--if-source-revision", evaluation.Revision))
			owner := graphMixedResult[graphstore.Record](t, call("show", evaluation.ID))
			if claim.Link.Type != scope+"types/satisfies" || owner.Revision == evaluation.Revision || len(owner.Owned) != 1 {
				t.Fatal("Evaluation claim was not owned by the Evaluation")
			}
			var retainedClaim graphstore.LinkRecord
			if err := json.Unmarshal(owner.Owned[0], &retainedClaim); err != nil || retainedClaim.ID != claim.Link.ID {
				t.Fatalf("Evaluation owned state omitted its claim: %v", err)
			}
			if after := graphMixedResult[graphstore.Record](t, call("show", policy.ID)); !reflect.DeepEqual(after, currentPolicy) {
				t.Fatal("Evaluation claim mutated its subject")
			}
			refuse("invalid_properties", "link", "create", "evaluation", "constraint", "--type", "types/conformance")
			if got := graphMixedResult[[]graphstore.LinkRecord](t, call("link", "list", "--type", "types/conformance")); len(got) != 0 {
				t.Fatal("abstract Type was instantiated")
			}
			for _, intrinsic := range []string{"Type", "BeadType", "LinkType", "ConformsTo", "Refines"} {
				refuse("invalid_selector", "list", "--type", "types/"+intrinsic, "--format", "records-json")
			}
			for _, removed := range []string{"preview-memory", "example-cites", "example-follows"} {
				refuse("capability_unavailable", "create", "Removed", "--type", "types/"+removed)
			}
			decision := graphMixedResult[graphstore.IssueRecord](t, call("create", "Native decision", "--id", "beads/decision", "--type", "decision"))
			if decision.Type != graphstore.IssueTypeURL(scope) {
				t.Fatal("bare native decision became a nominal Memory Type")
			}
			epic := graphMixedResult[graphstore.IssueRecord](t, call("create", "Epic", "--id", "beads/epic", "--type", "epic"))
			call("create", "Child", "--type", "task", "--parent", epic.ID)
			call("children", epic.ID)
			call("epic", "status")
			currentEpic := graphMixedResult[graphstore.IssueRecord](t, call("show", epic.ID, "--format", "graph-json"))
			call("link", "create", epic.ID, decision.ID, "--type", "blocks", "--if-source-revision", currentEpic.Revision)
		})
	}
}
