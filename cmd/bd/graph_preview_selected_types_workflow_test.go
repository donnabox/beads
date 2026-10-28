//go:build cgo

package main

import (
	"encoding/json"
	"fmt"
	graph "github.com/steveyegge/beads/graphops"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/configfile"
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
			graphAssessmentSwitch(t, work, true)
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			refuse := func(code string, args ...string) {
				t.Helper()
				graphPolicyCLI(t, bd, work, home, nil, code, append(args, "--json")...)
			}
			inventory := graphMixedResult[graphTypeInventory](t, call("types"))
			// Native concrete declarations extend the inventory; the fixed
			// selected catalog must still contain its selected eight Bead and ten Link declarations.
			baseCount := func(descriptors []json.RawMessage, nativeBase string) int {
				t.Helper()
				count := 0
				for _, raw := range descriptors {
					var descriptor struct {
						ID, Name, Describes string
						ConformsTo          []string
					}
					if err := json.Unmarshal(raw, &descriptor); err != nil {
						t.Fatal(err)
					}
					base, classification, native := graphstore.NativeTypeClassification(scope, descriptor.ID, graph.ResourceKind(descriptor.Describes))
					if !native {
						count++
						continue
					}
					if base != nativeBase || descriptor.Name != classification || len(descriptor.ConformsTo) != 1 || descriptor.ConformsTo[0] != nativeBase {
						t.Fatalf("native descriptor lost its identity or family: %s", raw)
					}
				}
				return count
			}
			if beads, links := baseCount(inventory.BeadTypes, graphstore.IssueTypeURL(scope)), baseCount(inventory.LinkTypes, graphstore.DependencyTypeURL(scope)); beads != 8 || links != 10 {
				t.Fatalf("selected catalog not installed: BeadTypes=%d LinkTypes=%d", beads, links)
			}
			for _, descriptor := range append(inventory.BeadTypes, inventory.LinkTypes...) {
				for _, absent := range []string{"types/preview-", "types/example-", "types/Type", "types/BeadType", "types/LinkType", "types/ConformsTo", "types/refines"} {
					if strings.Contains(string(descriptor), absent) {
						t.Fatalf("private/old Type exposed: %s", descriptor)
					}
				}
			}
			policy := graphMixedResult[graphstore.Record](t, call("create", "--type", "policy", "--id", "policy-note", "--properties", `{"title":"Policy","body":"Keep the body","extra":{"kept":[1,true,null]}}`))
			if policy.Type != scope+"policy" || !strings.Contains(call("show", policy.ID), `"kept"`) {
				t.Fatal("selected creation/show lost Type or authored extras")
			}
			memory := graphMixedResult[graphstore.Record](t, call("create", "--type", "memory", "--id", "memory-note", "--properties", `{"title":"Memory","body":"Original","extra":{"kept":true}}`))
			edited := graphMixedResult[graphstore.MemoryMutationResult](t, call("update", memory.ID, "--type", "requirement", "--properties", `{"body":"Changed","new":42}`, "--if-revision", "="+memory.Revision))
			if !edited.Changed || edited.Memory.Type != scope+"requirement" || !strings.Contains(call("show", memory.ID), `"kept"`) {
				t.Fatal("combined retype/merge lost inherited complete properties")
			}
			retained := graphMixedResult[graphstore.Record](t, call("show", "["+memory.ID+"]@="+memory.Revision))
			if !reflect.DeepEqual(retained, memory) {
				t.Fatal("retype rewrote the original Type/payload/history")
			}
			refuse("revision_conflict", "update", memory.ID, "--type", "policy", "--if-revision", "="+memory.Revision)
			patched := graphMixedResult[graphstore.MemoryMutationResult](t, call("update", memory.ID, "--type", "scenario", "--patch", `[{"op":"add","path":"/ordered","value":["one","two"]}]`, "--if-revision", "="+edited.Memory.Revision))
			if patched.Memory.Type != scope+"scenario" || !strings.Contains(call("show", memory.ID), `"ordered"`) {
				t.Fatal("ordered properties and Type did not commit together")
			}
			typeOnly := graphMixedResult[graphstore.MemoryMutationResult](t, call("update", memory.ID, "--type", "requirement"))
			if !typeOnly.Changed || typeOnly.Memory.Type != scope+"requirement" || !strings.Contains(call("show", memory.ID), `"ordered"`) {
				t.Fatal("unguarded type-only update did not preserve complete properties")
			}
			var selected struct {
				Result struct{ Items []struct{ ID, Type string } }
			}
			if err := json.Unmarshal([]byte(call("list", "--type", "requirement", "--format", "records-json", "--all")), &selected); err != nil || len(selected.Result.Items) != 1 || selected.Result.Items[0].ID != memory.ID {
				t.Fatalf("selected Bead filter failed: %+v %v", selected, err)
			}
			constraint := graphMixedResult[graphstore.Record](t, call("create", "Constraint", "--type", "constraint", "--id", "constraint-note"))
			owned := graphMixedResult[graphstore.LinkMutationResult](t, call("link", "create", "policy-note", "constraint-note", "--id", "subject", "--type", "subject-to"))
			if owned.Link.Type != scope+"subject-to" || owned.ReplacedSource == nil {
				t.Fatal("selected Link creation did not use source ownership")
			}
			refuse("invalid_properties", "update", constraint.ID, "--type", "memory", "--if-revision", "="+constraint.Revision)
			if got := graphMixedResult[graphstore.Record](t, call("show", constraint.ID)); !reflect.DeepEqual(got, constraint) {
				t.Fatal("incident-Link refusal changed the Bead")
			}
			links := graphMixedResult[[]graphstore.LinkRecord](t, call("link", "list", "--type", "subject-to"))
			if len(links) != 1 || links[0].ID != owned.Link.ID {
				t.Fatal("selected Link filter failed")
			}
			currentPolicy := graphMixedResult[graphstore.Record](t, call("show", policy.ID))
			subjects, err := json.Marshal(map[string]any{"title": "Evaluation", "subjects": []map[string]string{{"uri": policy.ID, "revision": currentPolicy.Revision}}})
			if err != nil {
				t.Fatal(err)
			}
			evaluation := graphMixedResult[graphstore.Record](t, call("create", "--id", "evaluation-note", "--type", "evaluation", "--properties", string(subjects)))
			claim := graphMixedResult[graphstore.LinkMutationResult](t, call("link", "create", "evaluation-note", "constraint-note", "--type", "satisfies", "--if-source-revision", evaluation.Revision))
			owner := graphMixedResult[graphstore.Record](t, call("show", evaluation.ID))
			if claim.Link.Type != scope+"satisfies" || owner.Revision == evaluation.Revision || len(owner.Owned) != 1 {
				t.Fatal("Evaluation claim was not owned by the Evaluation")
			}
			var retainedClaim graphstore.LinkRecord
			if err := json.Unmarshal(owner.Owned[0], &retainedClaim); err != nil || retainedClaim.ID != claim.Link.ID {
				t.Fatalf("Evaluation owned state omitted its claim: %v", err)
			}
			if after := graphMixedResult[graphstore.Record](t, call("show", policy.ID)); !reflect.DeepEqual(after, currentPolicy) {
				t.Fatal("Evaluation claim mutated its subject")
			}
			refuse("invalid_properties", "link", "create", "evaluation-note", "constraint-note", "--type", "conformance")
			if got := graphMixedResult[[]graphstore.LinkRecord](t, call("link", "list", "--exact-type", "conformance")); len(got) != 0 {
				t.Fatal("abstract Type was instantiated")
			}
			for _, intrinsic := range []string{"Type", "BeadType", "LinkType", "ConformsTo", "Refines"} {
				refuse("capability_unavailable", "list", "--type", intrinsic, "--format", "records-json")
			}
			for _, removed := range []string{"preview-memory", "example-cites", "example-follows"} {
				// Whole literal IDs have no deprecated-prefix alias. The native
				// admission path refuses these unconfigured classifications.
				refuse("invalid_properties", "create", "Removed", "--type", "types/"+removed)
			}
			decision := graphMixedResult[graphstore.IssueRecord](t, call("create", "Native decision", "--graph-path", "beads/decision", "--type", "decision"))
			nativeDecisionType, err := graphstore.NativeIssueTypeURL(scope, "decision")
			if err != nil || decision.Type != nativeDecisionType {
				t.Fatalf("bare native decision lost its concrete native Type: got=%s want=%s err=%v", decision.Type, nativeDecisionType, err)
			}
			epic := graphMixedResult[graphstore.IssueRecord](t, call("create", "Epic", "--graph-path", "beads/epic", "--type", "epic"))
			call("create", "Child", "--type", "task", "--parent", epic.ID)
			call("children", epic.ID)
			call("epic", "status")
			currentEpic := graphMixedResult[graphstore.IssueRecord](t, call("show", epic.ID, "--format", "graph-json"))
			call("link", "create", epic.ID, decision.ID, "--type", "blocks", "--if-source-revision", currentEpic.Revision)
		})
	}
}

// The real CLI reads the existing workspace config seam on every invocation.
func graphAssessmentSwitch(t *testing.T, work string, enabled bool) {
	t.Helper()
	dir := filepath.Join(work, ".beads")
	cfg, err := configfile.LoadForDiscovery(dir)
	if err != nil || cfg == nil {
		t.Fatalf("load assessment fixture config: %v", err)
	}
	cfg.GraphAssessmentsEnabled = enabled
	if err := cfg.Save(dir); err != nil {
		t.Fatal(err)
	}
}

func TestGraphPreviewAssessmentAvailabilityWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			const scope = "https://example.invalid/assessment-gate/"
			work, home := graphListWorkspace(t, bd, engine, scope)
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			refuse := func(args ...string) {
				t.Helper()
				graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", append(args, "--json")...)
			}
			inventory := graphMixedResult[graphTypeInventory](t, call("types"))
			if len(inventory.BeadTypes) != 20 || len(inventory.LinkTypes) != 25 {
				t.Fatal("default assessment inventory is not20 Bead/25 Link Types")
			}
			for _, args := range [][]string{{"create", "E", "--type", "evaluation", "--id", "denied"}, {"create", "E", "--type", scope + "evaluation", "--id", "denied"}} {
				refuse(args...)
			}
			subject := graphMixedResult[graphstore.Record](t, call("create", "Subject", "--id", "subject", "--type", "memory"))
			constraint := graphMixedResult[graphstore.Record](t, call("create", "Rule", "--id", "rule", "--type", "constraint"))
			props, err := json.Marshal(map[string]any{"title": "Evaluation", "subjects": []map[string]string{{"uri": subject.ID, "revision": subject.Revision}}})
			if err != nil {
				t.Fatal(err)
			}
			graphAssessmentSwitch(t, work, true)
			all := graphMixedResult[graphTypeInventory](t, call("types", "--details"))
			if len(all.BeadTypes) != 21 || len(all.LinkTypes) != 29 {
				t.Fatal("enabled CLI inventory lost definitions")
			}
			eval := graphMixedResult[graphstore.Record](t, call("create", "--type", "evaluation", "--id", "eval", "--properties", string(props)))
			claim := graphMixedResult[graphstore.LinkMutationResult](t, call("link", "create", eval.ID, constraint.ID, "--type", "satisfies", "--id", "claim", "--if-source-revision", eval.Revision))
			graphAssessmentSwitch(t, work, false)
			before, history, link := call("show", eval.ID), call("versions", eval.ID), call("link", "show", claim.Link.ID)
			refuse("update", eval.ID, "--properties", `{"body":"refused"}`)
			refuse("update", eval.ID, "--from-revision", "="+eval.Revision)
			refuse("link", "update", claim.Link.ID, "--properties", `{"evidence":"refused"}`)
			refuse("update", subject.ID, "--type", "evaluation", "--properties", string(props))
			if call("show", eval.ID) != before || call("versions", eval.ID) != history || call("link", "show", claim.Link.ID) != link {
				t.Fatal("CLI refusal changed current/history/owned claim")
			}
			removed := graphMixedResult[graphstore.LinkDeleteResult](t, call("link", "delete", claim.Link.ID, "--force"))
			if !removed.Changed || removed.Link.State != "deleted" {
				t.Fatal("disabled assessment cleanup did not remove the claim")
			}
			call("update", eval.ID, "--type", "memory")
			refuse("update", eval.ID, "--type", "evaluation")
			call("delete", eval.ID, "--force")
		})
	}
}

// Exercise the reported create/list disagreement through the actual binary,
// including storage, inventory JSON, nominal/family filters and shared flags.
func TestGraphPreviewFlatTypeIdentityWorkflow(t *testing.T) {
	isolateBeadsDirForTest(t)
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			const scope = "https://example.invalid/flat-types/"
			work, home := graphListWorkspace(t, bd, engine, scope)
			defaultInventory := graphMixedResult[graphTypeInventory](t, graphPolicyCLI(t, bd, work, home, nil, "", "types", "--json"))
			if len(defaultInventory.BeadTypes) != 20 || len(defaultInventory.LinkTypes) != 25 {
				t.Fatal("default catalog is not20Bead25Link", defaultInventory)
			}
			graphAssessmentSwitch(t, work, true)
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			for _, tc := range []struct{ id, name string }{{"note", "memory"}, {"scenario-note", "scenario"}, {"constraint-note", "constraint"}} {
				record := graphMixedResult[graphstore.Record](t, call("create", "Title "+tc.id, "--id", tc.id, "--type", tc.name, "--body", "Body"))
				if record.ID != scope+tc.id || record.Type != scope+tc.name {
					t.Fatalf("actual create identity %s: %+v", tc.name, record)
				}
				shown := graphMixedResult[graphstore.Record](t, call("show", record.ID))
				if shown.Type != record.Type {
					t.Fatal("show changed declared Type")
				}
			}
			inventory := graphMixedResult[graphTypeInventory](t, call("types", "--details"))
			if len(inventory.BeadTypes)+len(inventory.LinkTypes) != 50 {
				t.Fatalf("fresh inventory is not exact50: %+v", inventory)
			}
			ids := map[string]bool{}
			for _, raw := range append(inventory.BeadTypes, inventory.LinkTypes...) {
				var d struct{ ID string }
				if err := json.Unmarshal(raw, &d); err != nil {
					t.Fatal(err)
				}
				ids[d.ID] = true
			}
			for _, name := range []string{"memory", "scenario", "decision", "task", "blocks", "references"} {
				if !ids[scope+name] {
					t.Fatalf("missing exact installed Type %s", name)
				}
			}
			if ids[scope+"procedure"] || ids[scope+"types/memory"] {
				t.Fatal("removed/hidden generated identity installed")
			}
			human := graphPolicyCLI(t, bd, work, home, nil, "", "types")
			if !strings.Contains(human, "  memory ") || strings.Contains(human, "native-issue/") || strings.Contains(human, "native-dependency/") {
				t.Fatalf("human catalog disagrees: %s", human)
			}
			for _, tc := range []struct {
				flag, name string
				want       int
			}{{"--type", "memory", 3}, {"--exact-type", "memory", 1}, {"--type", scope + "memory", 3}, {"--exact-type", scope + "scenario", 1}} {
				var page struct{ Items []json.RawMessage }
				page = graphMixedResult[struct{ Items []json.RawMessage }](t, call("list", tc.flag, tc.name, "--format", "records-json", "--all"))
				if len(page.Items) != tc.want {
					t.Fatalf("list %s %s = %d want%d", tc.flag, tc.name, len(page.Items), tc.want)
				}
				var count struct{ Count int }
				if err := json.Unmarshal([]byte(call("count", tc.flag, tc.name)), &count); err != nil || count.Count != tc.want {
					t.Fatalf("count %s %s = %d %v", tc.flag, tc.name, count.Count, err)
				}
			}
			linked := graphMixedResult[struct {
				Link   graphstore.LinkRecord
				Source struct{ Revision string }
			}](t, call("link", "create", "note", "constraint-note", "--id", "context", "--type", "subject-to"))
			if linked.Link.Type != scope+"subject-to" || linked.Source.Revision == "" || linked.Link.Revision == linked.Source.Revision {
				t.Fatal("Link/owned source independent identities or revisions lost", linked)
			}
			graphOut := graphMixedResult[struct{ Links []struct{ Type string } }](t, call("graph", "note", "--link-type", "subject-to", "--depth", "1"))
			if len(graphOut.Links) != 1 || graphOut.Links[0].Type != linked.Link.Type {
				t.Fatalf("graph --link-type differs: %+v", graphOut)
			}
			changed := graphMixedResult[graphstore.MemoryMutationResult](t, call("update", "note", "--type", "policy"))
			if changed.Memory.Type != scope+"policy" || changed.Memory.ID != scope+"note" {
				t.Fatal("retype changed identity", changed)
			}
			// Exact raw native labels resolve identically through mutation, Link
			// collection filtering and both graph Type flags, without graph-only folding.
			source := graphMixedResult[graphstore.IssueRecord](t, call("create", "Work"))
			for index, raw := range []string{"é", "a b", "p%q", "BLOCKS", "unfamiliar/é"} {
				want, err := graphstore.NativeDependencyTypeURL(scope, raw)
				if err != nil {
					t.Fatal(err)
				}
				local, err := graph.ScopeLocalPublicResourceID(scope, want)
				if err != nil {
					t.Fatal(err)
				}
				for variant, operand := range []string{raw, local, want} {
					target := graphMixedResult[graphstore.IssueRecord](t, call("create", fmt.Sprintf("Target%d-%d", index, variant)))
					result := graphMixedResult[struct{ Link graphstore.LinkRecord }](t, call("link", source.ID, target.ID, "--type", operand, "--id", fmt.Sprintf("edge-%d-%d", index, variant)))
					if result.Link.Type != want {
						t.Fatalf("native operand %q mutation %q want %q", operand, result.Link.Type, want)
					}
				}
				for _, operand := range []string{raw, local, want} {
					page := graphMixedResult[[]graphstore.LinkRecord](t, call("link", "list", "--exact-type", operand))
					if len(page) != 3 {
						t.Fatalf("native operand %q Link list count%d", operand, len(page))
					}
					for _, flag := range []string{"--link-type", "--exact-link-type"} {
						selected := graphMixedResult[struct{ Links []struct{ Type string } }](t, call("graph", source.ID, flag, operand, "--depth", "1"))
						if len(selected.Links) != 3 {
							t.Fatalf("native operand %q graph%s count%d", operand, flag, len(selected.Links))
						}
						for _, link := range selected.Links {
							if link.Type != want {
								t.Fatalf("native operand %q graph%s changedType: %s", operand, flag, link.Type)
							}
						}
					}
				}
			}

			before := call("types")
			graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "config", "set", "types.custom", "memory", "--json")
			if after := call("types"); after != before {
				t.Fatal("refused normal config command changed inventory")
			}
			for _, args := range [][]string{{"create", "Impossible", "--bead-type", "memory"}, {"list", "--bead-type", "memory"}} {
				_, stderr, exit := graphGenerationRun(t, bd, work, home, args...)
				if exit == 0 || !strings.Contains(stderr, "unknown flag: --bead-type") {
					t.Fatalf("nonexistent flag admitted: %v exit%d %s", args, exit, stderr)
				}
			}
		})
	}
}

func TestGraphPreviewNativeIssueAliasParity(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			const scope = "https://example.invalid/alias-parity/"
			work, home := graphListWorkspace(t, bd, engine, scope)
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			for _, alias := range []struct{ raw, name string }{{"feat", "feature"}, {"adr", "decision"}} {
				created := graphMixedResult[graphstore.IssueRecord](t, call("create", "Alias parity", "--type", alias.raw))
				if created.Type != scope+alias.name {
					t.Fatalf("alias %s create Type %s", alias.raw, created.Type)
				}
				shown := graphMixedResult[graphstore.IssueRecord](t, call("show", created.ID, "--format", "graph-json"))
				if shown.Type != created.Type {
					t.Fatalf("alias %s show Type %s", alias.raw, shown.Type)
				}
				for _, literal := range []string{scope + alias.raw, "types/" + alias.raw} {
					graphPolicyCLI(t, bd, work, home, nil, "invalid_selector", "list", "--type", literal, "--format", "records-json", "--json")
				}
				for _, flag := range []string{"--type", "--exact-type"} {
					page := graphMixedResult[struct{ Items []graphstore.IssueRecord }](t, call("list", flag, alias.raw, "--format", "records-json", "--all"))
					found := false
					for _, item := range page.Items {
						if item.ID == created.ID && item.Type == created.Type {
							found = true
						}
					}
					if !found {
						t.Fatalf("alias %s %s list lost created Issue", alias.raw, flag)
					}
					var count struct{ Count int }
					if err := json.Unmarshal([]byte(call("count", flag, alias.raw)), &count); err != nil || count.Count != len(page.Items) {
						t.Fatalf("alias %s count/list disagreement: %+v %v", alias.raw, count, err)
					}
				}
				updated := graphMixedResult[graphstore.IssueMutationResult](t, call("update", created.ID, "--type", alias.raw))
				if updated.Issue.Type != created.Type {
					t.Fatalf("alias %s update Type %s", alias.raw, updated.Issue.Type)
				}
			}

		})
	}
}
