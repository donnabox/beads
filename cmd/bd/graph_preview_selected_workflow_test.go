//go:build cgo

package main

import (
	"reflect"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

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
				call("create", id, "--id", id, "--type", "task")
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
			generic := graphSelectedTestSummary(t, call("graph", "i1", "--link-type", "References", "--direction", "out", "--depth", "10"))
			if len(generic.Nodes) != 4 || len(generic.Links) != 6 || !generic.Complete {
				t.Fatalf("mixed path or parallel/cycle lost: %+v", generic)
			}
			found := false
			for _, l := range generic.Links {
				if l.ID == scope+"links/pin" {
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
			for _, flag := range []string{"--dot", "--html", "--compact", "--box"} {
				text := graphPolicyCLI(t, bd, work, home, nil, "", "graph", "i1", "--link-type", "references", flag)
				for _, id := range []string{"links/a", "links/parallel", "links/cycle", "links/pin"} {
					if !strings.Contains(text, id) {
						t.Fatalf("%s lost %s", flag, id)
					}
				}
			}
			graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "graph", "i1", "--link-type", "references", "--open", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "graph", "i1", "--link-type", "references", "--max-nodes=2", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "graph", "i1", "--link-type", "types/memory", "--json")
		})
	}
}
