//go:build cgo

package main

import (
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func TestGraphPreviewStructureReads(t *testing.T) {
	bd := buildBDUnderTest(t)
	work, home := t.TempDir(), t.TempDir()
	const scope = "https://example.invalid/structure/"
	graphPolicyCLI(t, bd, work, home, nil, "", "init", "--graph-mode", "link", "--prefix", "demo", "--scope-url", scope, "--skip-hooks", "--skip-agents", "--non-interactive", "--json")
	graphPolicyCLI(t, bd, work, home, nil, "", "create", "Parent", "--id", "beads/parent", "--json")
	graphPolicyCLI(t, bd, work, home, nil, "", "create", "Work", "--id", "beads/work", "--json")

	// The read command is admitted, but it must not invent children from names
	// or treat a blocking dependency as parentage.
	children := graphMixedResult[[]graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "children", scope+"beads/parent", "--json"))
	if len(children) != 0 {
		t.Fatalf("unexpected children before a native parent-child edge: %+v", children)
	}
	graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "children", "parent", "--pretty", "--json")
	graphPolicyCLI(t, bd, work, home, nil, "not_found", "children", "missing", "--json")
	graphPolicyCLI(t, bd, work, home, nil, "invalid_selector", "children", "https://other.invalid/beads/parent", "--json")

	created := graphMixedResult[graphstore.DependencyResult](t, graphPolicyCLI(t, bd, work, home, nil, "", "dep", "add", "beads/work", "beads/parent", "--json"))
	if created.Link.Source != scope+"beads/work" || created.Link.Target != scope+"beads/parent" {
		t.Fatalf("dependency endpoints lost canonical IDs: %+v", created.Link)
	}
	workHistory := graphPolicyCLI(t, bd, work, home, nil, "", "versions", "work", "--json")
	parentHistory := graphPolicyCLI(t, bd, work, home, nil, "", "versions", "parent", "--json")
	for _, tc := range []struct {
		selector, direction string
	}{
		{"work", "down"},
		{scope + "beads/parent", "up"},
	} {
		links := graphMixedResult[[]graphstore.LinkRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "dep", "list", tc.selector, "--direction", tc.direction, "--json"))
		if len(links) != 1 || links[0].ID != created.Link.ID || links[0].Source != created.Link.Source || links[0].Target != created.Link.Target {
			t.Fatalf("%s list did not preserve canonical Link: %+v", tc.direction, links)
		}
	}
	children = graphMixedResult[[]graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "children", "parent", "--json"))
	if len(children) != 0 {
		t.Fatalf("blocking dependency became parentage: %+v", children)
	}
	graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "dep", "list", "work", "--type", "parent-child", "--json")
	graphPolicyCLI(t, bd, work, home, nil, "invalid_selector", "dep", "list", "work", "--direction", "both", "--json")
	graphPolicyCLI(t, bd, work, home, nil, "invalid_selector", "dep", "list", "https://other.invalid/beads/work", "--json")
	if got := graphPolicyCLI(t, bd, work, home, nil, "", "dep", "list", "work"); !strings.Contains(got, "beads/work depends on beads/parent") {
		t.Fatalf("missing human dependency summary: %q", got)
	}
	view := graphMixedResult[graphDependencyView](t, graphPolicyCLI(t, bd, work, home, nil, "", "graph", "work", "--json"))
	if view.Root != scope+"beads/work" || len(view.Issues) != 2 || len(view.Dependencies) != 1 || view.Dependencies[0].ID != created.Link.ID {
		t.Fatalf("blocking graph lost its component or canonical Link: %+v", view)
	}
	check := graphMixedResult[graphDependencyCheck](t, graphPolicyCLI(t, bd, work, home, nil, "", "graph", "check", "--json"))
	if !check.Clean || len(check.Cycles) != 0 {
		t.Fatalf("acyclic native blocking graph failed check: %+v", check)
	}
	tree := graphMixedResult[[]graphDependencyTreeNode](t, graphPolicyCLI(t, bd, work, home, nil, "", "dep", "tree", "work", "--json"))
	if len(tree) != 2 || tree[0].Issue.ID != scope+"beads/work" || tree[1].Issue.ID != scope+"beads/parent" || tree[1].ParentID != scope+"beads/work" {
		t.Fatalf("blocking dependency tree lost canonical parentage: %+v", tree)
	}
	if cycles := graphMixedResult[[][]string](t, graphPolicyCLI(t, bd, work, home, nil, "", "dep", "cycles", "--json")); len(cycles) != 0 {
		t.Fatalf("acyclic blocking graph reported cycles: %+v", cycles)
	}
	generic := graphMixedResult[graphGenericResult](t, graphPolicyCLI(t, bd, work, home, nil, "", "graph", "work", "--view", "generic", "--json"))
	if generic.Root != scope+"beads/work" || generic.Projection != "summary" {
		t.Fatalf("generic traversal changed: %+v", generic)
	}
	graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "graph", "work", "--dot", "--json")
	graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "dep", "tree", "work", "--status", "open", "--json")
	graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "dep", "cycles", "--include-tracks", "--json")
	graphPolicyCLI(t, bd, work, home, nil, "invalid_selector", "graph", "check", "work", "--json")
	if got := graphPolicyCLI(t, bd, work, home, nil, "", "versions", "work", "--json"); got != workHistory {
		t.Fatalf("dependency reads changed source History\nbefore: %s\nafter: %s", workHistory, got)
	}
	if got := graphPolicyCLI(t, bd, work, home, nil, "", "versions", "parent", "--json"); got != parentHistory {
		t.Fatalf("child and dependency reads changed parent History\nbefore: %s\nafter: %s", parentHistory, got)
	}
}
