//go:build cgo

package main

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func TestGraphPreviewDependencyRowCapGate(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Int(maxRowsFlagName, 0, "")
	t.Setenv(maxRowsEnvVar, "1")
	var refusal *exitError
	if err := graphPreviewRejectDependencyRowCap(cmd); !errors.As(err, &refusal) || refusal.Code != 5 {
		t.Fatalf("positive environment cap was not refused before a snapshot: %v", err)
	}
	if err := cmd.Flags().Set(maxRowsFlagName, "0"); err != nil {
		t.Fatal(err)
	}
	if err := graphPreviewRejectDependencyRowCap(cmd); err != nil {
		t.Fatalf("explicit zero must override the environment cap: %v", err)
	}
}

func TestGraphPreviewDepTreeBothDistinguishesDirections(t *testing.T) {
	bd := buildBDUnderTest(t)
	work, home := t.TempDir(), t.TempDir()
	const scope = "https://example.invalid/tree-directions/"
	graphPolicyCLI(t, bd, work, home, nil, "", "init", "--graph-mode", "link", "--prefix", "demo", "--scope-url", scope, "--skip-hooks", "--skip-agents", "--non-interactive", "--json")
	for _, issue := range []struct{ id, title string }{
		{"beads/root", "Root"},
		{"beads/blocker", "Blocker"},
		{"beads/dependent", "Dependent"},
	} {
		graphPolicyCLI(t, bd, work, home, nil, "", "create", issue.title, "--id", issue.id, "--json")
	}
	graphPolicyCLI(t, bd, work, home, nil, "", "dep", "add", "beads/root", "beads/blocker", "--json")
	graphPolicyCLI(t, bd, work, home, nil, "", "dep", "add", "beads/dependent", "beads/root", "--json")

	got := strings.TrimSpace(graphPolicyCLI(t, bd, work, home, nil, "", "dep", "tree", "root", "--direction", "both"))
	want := strings.Join([]string{
		"Blocking dependency tree for beads/root (both; graph preview)",
		"beads/root  Root",
		"Dependencies (what beads/root depends on):",
		"  beads/blocker  Blocker",
		"Dependents (what depends on beads/root):",
		"  beads/dependent  Dependent",
	}, "\n")
	if got != want {
		t.Fatalf("both-direction tree must separate blockers from dependents\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestGraphPreviewStructureReads(t *testing.T) {
	bd := buildBDUnderTest(t)
	work, home := t.TempDir(), t.TempDir()
	const scope = "https://example.invalid/structure/"
	graphPolicyCLI(t, bd, work, home, nil, "", "init", "--graph-mode", "link", "--prefix", "demo", "--scope-url", scope, "--skip-hooks", "--skip-agents", "--non-interactive", "--json")
	graphPolicyCLI(t, bd, work, home, nil, "", "create", "Parent", "--id", "beads/parent", "--json")
	graphPolicyCLI(t, bd, work, home, nil, "", "create", "Work", "--id", "beads/work", "--json")
	graphPolicyCLI(t, bd, work, home, nil, "", "create", "Isolated", "--id", "beads/isolated", "--json")

	// A blocking-only workspace has no canonical parent-child Link mapping.
	// An empty children result would falsely look like a complete family read.
	graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "children", scope+"beads/parent", "--json")
	graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "children", "parent", "--pretty", "--json")
	graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "children", "missing", "--json")

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
		got := graphPolicyCLI(t, bd, work, home, nil, "", "dep", "list", tc.selector, "--direction", tc.direction)
		if !strings.Contains(got, "beads/work depends on beads/parent") {
			t.Fatalf("%s list lost blocking dependency: %q", tc.direction, got)
		}
	}
	graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "children", "parent", "--json")
	graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "dep", "list", "work", "--type", "parent-child", "--json")
	graphPolicyCLI(t, bd, work, home, nil, "invalid_selector", "dep", "list", "work", "--direction", "both", "--json")
	graphPolicyCLI(t, bd, work, home, nil, "invalid_selector", "dep", "list", "https://other.invalid/beads/work", "--json")
	if got := graphPolicyCLI(t, bd, work, home, nil, "", "dep", "list", "work"); !strings.Contains(got, "beads/work depends on beads/parent") {
		t.Fatalf("missing human dependency summary: %q", got)
	}
	if view := graphPolicyCLI(t, bd, work, home, nil, "", "graph", "work"); !strings.Contains(view, "beads/work depends on beads/parent") || !strings.Contains(view, "2 Issues, 1 Dependencies") {
		t.Fatalf("blocking graph lost its component: %q", view)
	}
	var check graphDependencyCheck
	checkJSON := graphPolicyCLI(t, bd, work, home, nil, "", "graph", "check", "--json")
	if err := json.Unmarshal([]byte(checkJSON), &check); err != nil {
		t.Fatalf("graph check must use native top-level JSON: %v: %s", err, checkJSON)
	}
	if !check.Clean || len(check.Cycles) != 0 {
		t.Fatalf("acyclic native blocking graph failed check: %+v", check)
	}
	if tree := graphPolicyCLI(t, bd, work, home, nil, "", "dep", "tree", "work"); !strings.Contains(tree, "beads/work") || !strings.Contains(tree, "beads/parent") {
		t.Fatalf("blocking dependency tree lost its nodes: %q", tree)
	}
	if tree := graphPolicyCLI(t, bd, work, home, nil, "", "dep", "tree", "isolated"); !strings.Contains(tree, "has no dependencies") {
		t.Fatalf("isolated Issue falsely reported a dependency tree: %q", tree)
	}
	if tree := graphPolicyCLI(t, bd, work, home, nil, "", "dep", "tree", "isolated", "--direction", "up"); !strings.Contains(tree, "has no dependents") {
		t.Fatalf("isolated Issue falsely reported dependents: %q", tree)
	}
	if tree := graphPolicyCLI(t, bd, work, home, nil, "", "dep", "tree", "work", "--direction", "both"); !strings.Contains(tree, "beads/work") || !strings.Contains(tree, "beads/parent") || strings.Index(tree, "beads/parent") < strings.Index(tree, "beads/work") {
		t.Fatalf("both-direction dependency tree must start at its root: %q", tree)
	}
	if cycles := graphPolicyCLI(t, bd, work, home, nil, "", "dep", "cycles"); !strings.Contains(cycles, "No blocking dependency cycles") {
		t.Fatalf("acyclic blocking graph reported cycles: %q", cycles)
	}
	generic := graphMixedResult[graphGenericResult](t, graphPolicyCLI(t, bd, work, home, nil, "", "graph", "work", "--view", "generic", "--json"))
	if generic.Root != scope+"beads/work" || generic.Projection != "summary" {
		t.Fatalf("generic traversal changed: %+v", generic)
	}
	graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "graph", "work", "--dot", "--json")
	for _, args := range [][]string{
		{"graph", "work", "--json"},
		{"dep", "list", "work", "--json"},
		{"dep", "list", "work", "parent", "--json"},
		{"dep", "tree", "work", "--json"},
		{"dep", "cycles", "--json"},
	} {
		graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", args...)
	}
	graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "graph", "work", "--max-rows=1", "--json")
	graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "dep", "tree", "work", "--max-rows=1", "--json")
	graphPolicyCLI(t, bd, work, home, []string{"BEADS_MAX_ROWS=1"}, "capability_unavailable", "graph", "work", "--json")
	graphPolicyCLI(t, bd, work, home, []string{"BEADS_MAX_ROWS=1"}, "capability_unavailable", "dep", "tree", "work", "--json")
	graphPolicyCLI(t, bd, work, home, []string{"BEADS_MAX_ROWS=bad"}, "invalid_properties", "graph", "work", "--json")
	graphPolicyCLI(t, bd, work, home, nil, "invalid_selector", "dep", "tree", "work", "--max-rows=-1", "--json")
	if got := graphPolicyCLI(t, bd, work, home, []string{"BEADS_MAX_ROWS=1"}, "", "graph", "work", "--max-rows=0"); !strings.Contains(got, "2 Issues, 1 Dependencies") {
		t.Fatalf("explicit zero did not override graph row cap: %q", got)
	}
	if got := graphPolicyCLI(t, bd, work, home, []string{"BEADS_MAX_ROWS=1"}, "", "dep", "tree", "work", "--max-rows=0"); !strings.Contains(got, "beads/parent") {
		t.Fatalf("explicit zero did not override tree row cap: %q", got)
	}
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

func TestGraphPreviewUnmappedIssueStructureRefusesWithoutMutation(t *testing.T) {
	bd := buildBDUnderTest(t)
	work, home := t.TempDir(), t.TempDir()
	const scope = "https://example.invalid/unmapped-structure/"
	graphPolicyCLI(t, bd, work, home, nil, "", "init", "--graph-mode", "link", "--prefix", "demo", "--scope-url", scope, "--skip-hooks", "--skip-agents", "--non-interactive", "--json")
	graphPolicyCLI(t, bd, work, home, nil, "", "create", "Original", "--id", "beads/original", "--json")
	graphPolicyCLI(t, bd, work, home, nil, "", "create", "Replacement", "--id", "beads/replacement", "--json")
	originalHistory := graphPolicyCLI(t, bd, work, home, nil, "", "versions", "original", "--json")
	replacementHistory := graphPolicyCLI(t, bd, work, home, nil, "", "versions", "replacement", "--json")
	before := legacyUpgradeTreeDigest(t, work)

	for _, args := range [][]string{
		{"duplicate", "original", "--of", "replacement", "--json"},
		{"supersede", "original", "--with", "replacement", "--json"},
		{"epic", "status", "--json"},
		{"epic", "status", "--eligible-only", "--json"},
		{"epic", "close-eligible", "--dry-run", "--json"},
		{"epic", "close-eligible", "--json"},
	} {
		graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", args...)
		if after := legacyUpgradeTreeDigest(t, work); after != before {
			t.Fatalf("unmapped Issue relationship command %v changed workspace", args)
		}
	}
	if got := graphPolicyCLI(t, bd, work, home, nil, "", "versions", "original", "--json"); got != originalHistory {
		t.Fatalf("refused relationship commands changed original History\nbefore: %s\nafter: %s", originalHistory, got)
	}
	if got := graphPolicyCLI(t, bd, work, home, nil, "", "versions", "replacement", "--json"); got != replacementHistory {
		t.Fatalf("refused relationship commands changed replacement History\nbefore: %s\nafter: %s", replacementHistory, got)
	}
}
