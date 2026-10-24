//go:build cgo

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// Capture both channels through the same private command environment as the
// other graph workflow helpers. Notices must not corrupt canonical JSON or
// accompany an acquisition/traversal refusal.
func graphSelectedWorkflowCapture(t *testing.T, bd, work, home, code string, args ...string) (string, string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
	defer cancel()
	cmd := graphMemoryReadCommand(ctx, bd, work, home, args...)
	var out, diagnostics bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &diagnostics
	err := cmd.Run()
	if ctx.Err() != nil {
		t.Fatalf("graph command exceeded deadline: %v args=%v stderr=%s", ctx.Err(), args, diagnostics.String())
	}
	if code == "" {
		if err != nil || !json.Valid(out.Bytes()) {
			t.Fatalf("graph command did not return structured success: args=%v err=%v stdout=%s stderr=%s", args, err, out.String(), diagnostics.String())
		}
	} else {
		var failure struct{ Code string }
		if err == nil || out.Len() != 0 || json.Unmarshal(diagnostics.Bytes(), &failure) != nil || failure.Code != code {
			t.Fatalf("graph refusal lost atomic output: expected=%s err=%v stdout=%s stderr=%s", code, err, out.String(), diagnostics.String())
		}
	}
	return out.String(), diagnostics.String()
}

// Real subprocess authoring + coherent provider acquisition is the boundary:
// pure projection tests cannot prove mixed native/generic CLI routing and pins.
func TestGraphPreviewSelectedGraphWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			const scope = "https://example.invalid/selected-graph/"
			work, home := graphListWorkspace(t, bd, engine, scope)
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			for _, id := range []string{"i1", "i2", "isolated-issue"} {
				call("create", id, "--id", id, "--type", "task", "--force")
			}
			for _, id := range []string{"m1", "m2", "isolated-memory"} {
				call("create", id, "--id", id, "--type", "types/memory")
			}
			for _, edge := range [][3]string{{"i1", "m1", "a"}, {"m1", "i2", "b"}, {"i2", "m2", "c"}, {"i1", "m1", "parallel"}, {"m1", "i1", "cycle"}} {
				call("link", "create", edge[0], edge[1], "--id", edge[2], "--type", "types/references")
			}
			call("dep", "add", "i1", "i2")
			call("dep", "add", "i2", "isolated-issue")
			pin := graphMixedResult[graphstore.Record](t, call("show", "m2"))
			call("link", "create", "m1", "m2", "--id", "pin", "--type", "types/references", "--target-revision", pin.Revision)
			call("update", "m2", "--properties", `{"title":"New current title"}`)
			implicit := call("graph", "i1")
			explicit := call("graph", "i1", "--link-type", "Dependency")
			if implicit != explicit {
				t.Fatalf("omitted Dependency differs: %s\n%s", implicit, explicit)
			}
			currentIssue := graphMixedResult[graphstore.IssueRecord](t, call("show", "i1", "--format", "graph-json"))
			historical := []string{"i1@1", "i1@=" + currentIssue.Revision}
			for _, withLiteralSiblings := range []bool{false, true} {
				if withLiteralSiblings {
					for _, id := range historical {
						// --id is literal data; its returned absolute URI selects
						// that identity without parsing a local revision suffix.
						literal := graphMixedResult[graphstore.Record](t, call("create", "Literal reference-shaped ID", "--id", id, "--type", "types/memory"))
						selected := graphSelectedTestSummary(t, call("graph", literal.ID, "--link-type=any"))
						if len(selected.Nodes) != 1 || selected.Nodes[0].ID != literal.ID || len(selected.Links) != 0 || !selected.Complete {
							t.Fatalf("canonical literal identity became a revision selector: %s %+v", literal.ID, selected)
						}
					}
				}
				if bare, current := call("graph", "i1"), call("graph", "i1@current"); bare != current {
					t.Fatalf("current reference changed graph with literal siblings=%t: %s\n%s", withLiteralSiblings, bare, current)
				}
				for _, selector := range historical {
					graphSelectedWorkflowCapture(t, bd, work, home, "capability_unavailable", "graph", selector, "--json")
				}
			}
			implicitJSON, implicitNotice := graphSelectedWorkflowCapture(t, bd, work, home, "", "graph", "i1", "--json")
			explicitJSON, explicitNotice := graphSelectedWorkflowCapture(t, bd, work, home, "", "graph", "i1", "--link-type=Dependency", "--json")
			if implicitJSON != implicit || explicitJSON != implicitJSON || !strings.Contains(implicitNotice, scope+"types/references") || !strings.Contains(implicitNotice, "--link-type any") || strings.Contains(explicitNotice, "omitted Link Types") {
				t.Fatalf("default disclosure changed stdout or explicit selection: implicit=%s explicit=%s implicit stderr=%s explicit stderr=%s", implicitJSON, explicitJSON, implicitNotice, explicitNotice)
			}
			quietJSON, quietNotice := graphSelectedWorkflowCapture(t, bd, work, home, "", "graph", "i1", "--quiet", "--json")
			if quietJSON != implicitJSON || quietNotice != "" {
				t.Fatalf("quiet graph changed JSON or emitted an advisory: stdout=%s stderr=%s", quietJSON, quietNotice)
			}
			_, refusedNotice := graphSelectedWorkflowCapture(t, bd, work, home, "capability_unavailable", "graph", "i1", "--max-nodes=1", "--json")
			if strings.Contains(refusedNotice, "omitted Link Types") {
				t.Fatalf("refusal emitted success disclosure: %s", refusedNotice)
			}
			generic := graphSelectedTestSummary(t, call("graph", "i1", "--link-type", "References", "--direction", "out", "--depth", "10"))
			if len(generic.Nodes) != 4 || len(generic.Links) != 6 || !generic.Complete {
				t.Fatalf("mixed path or parallel/cycle lost: %+v", generic)
			}
			found := false
			for _, l := range generic.Links {
				if l.ID == scope+"pin" {
					found = l.TargetRevision == pin.Revision
				}
			}
			if !found {
				t.Fatal("actual persisted pin erased")
			}
			shallow := graphSelectedTestSummary(t, call("graph", "i1", "--link-type", "references", "--direction", "out", "--depth", "1"))
			if len(shallow.Nodes) != 2 || len(shallow.Links) != 3 || shallow.Complete {
				t.Fatalf("induced cycle/parallel or frontier lost: %+v", shallow)
			}
			all := graphSelectedTestSummary(t, call("graph", "--all", "--link-type", "references"))
			if len(all.Nodes) != 4 || len(all.Links) != 6 {
				t.Fatalf("nondependency all selected unrelated isolates: %+v", all)
			}
			union := graphSelectedTestSummary(t, call("graph", "i1", "--link-type", "references", "--link-type", "dependency"))
			if len(union.Nodes) != 5 || len(union.Links) != 8 {
				t.Fatalf("repeatable union: %+v", union)
			}
			any := graphSelectedTestSummary(t, call("graph", "i1", "--link-type=any"))
			mixed := graphSelectedTestSummary(t, call("graph", "i1", "--link-type=dependency", "--exact-link-type=references", "--exact-link-type=references"))
			for name, selected := range map[string]graphGenericResult{"any": any, "family-exact union": mixed} {
				if !reflect.DeepEqual(selected.Nodes, union.Nodes) || !reflect.DeepEqual(selected.Links, union.Links) {
					t.Fatalf("%s changed selected population or duplicated Links: %+v expected=%+v", name, selected, union)
				}
			}
			exact := graphSelectedTestSummary(t, call("graph", "i1", "--exact-link-type=references", "--direction=out", "--depth=10"))
			if !reflect.DeepEqual(exact.Nodes, generic.Nodes) || !reflect.DeepEqual(exact.Links, generic.Links) {
				t.Fatalf("exact References changed the six stored Reference Links: %+v expected=%+v", exact, generic)
			}
			exactIDs := []string{}
			for _, link := range exact.Links {
				if link.Type != scope+"types/references" {
					t.Fatalf("exact References included another Type: %+v", link)
				}
				exactIDs = append(exactIDs, strings.TrimPrefix(link.ID, scope))
			}
			if !reflect.DeepEqual(exactIDs, []string{"a", "b", "c", "cycle", "parallel", "pin"}) {
				t.Fatalf("exact References lost direct stored identities: %v", exactIDs)
			}
			family := graphSelectedTestSummary(t, call("graph", "i1", "--link-type", "references"))
			if !reflect.DeepEqual(family.Nodes, generic.Nodes) || !reflect.DeepEqual(family.Links, generic.Links) {
				t.Fatalf("Association refinements: %+v %+v", family, union)
			}
			isolated := graphSelectedTestSummary(t, call("graph", "isolated-memory", "--link-type", "references"))
			if len(isolated.Nodes) != 1 || len(isolated.Links) != 0 {
				t.Fatal("explicit isolated root lost")
			}
			if call("graph", "--all") != call("graph", "--all", "--link-type", "dependency") {
				t.Fatal("all Dependency defaults differ")
			}
			allDeps := graphSelectedTestSummary(t, call("graph", "--all"))
			if len(allDeps.Nodes) != 3 || len(allDeps.Links) != 2 {
				t.Fatalf("Dependency all lost familiar isolated open Issue: %+v", allDeps)
			}
			for _, flag := range []string{"", "--dot", "--html", "--compact", "--box"} {
				args := []string{"graph", "i1", "--link-type", "references"}
				if flag != "" {
					args = append(args, flag)
				}
				text := graphPolicyCLI(t, bd, work, home, nil, "", args...)
				for _, id := range []string{"a", "parallel", "cycle", "pin"} {
					want := "(" + id + ")"
					if flag == "--dot" || flag == "--html" {
						want = scope + id // export data keeps the exact canonical ID
					}
					if !strings.Contains(text, want) {
						t.Fatalf("%s lost %s", flag, id)
					}
				}
				if flag == "" || flag == "--compact" {
					if !strings.Contains(text, "└──") || !strings.Contains(text, "4 Beads, 6 Links") || !strings.Contains(text, "(cycle)") || !strings.Contains(text, "target revision=\""+pin.Revision+"\"") {
						t.Fatalf("installed compact relationship tree lost selection/pins:\n%s", text)
					}
				}
			}
			graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "graph", "i1", "--link-type", "references", "--open", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "graph", "i1", "--link-type", "references", "--max-nodes=2", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "graph", "i1", "--link-type", "types/memory", "--json")
		})
	}
}
