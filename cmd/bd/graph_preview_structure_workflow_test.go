//go:build cgo

package main

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/configfile"
	"github.com/steveyegge/beads/internal/storage/embeddeddolt"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/storage/issueops"
)

// A native graph writer rejects cycles. For the read-only check's cyclic case,
// retarget one mapped Dependency in a disposable workspace and update its
// retained Link and owning Issue bytes together. This is fixture setup only.
func graphPreviewRetargetDependencyForCycle(t *testing.T, work, scope string, link graphstore.LinkRecord, sourceID, targetID, targetURL string) {
	t.Helper()
	cfg, err := configfile.LoadForDiscovery(filepath.Join(work, ".beads"))
	if err != nil || cfg == nil {
		t.Fatalf("load fixture workspace: config=%v err=%v", cfg, err)
	}
	ctx := t.Context()
	db, closeDB, err := embeddeddolt.OpenSQL(ctx, filepath.Join(cfg.GraphWorkspace, "embeddeddolt"), cfg.DoltDatabase, "main")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := closeDB(); err != nil {
			t.Error(err)
		}
	}()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback() //nolint:errcheck // Commit below owns the successful path.
	canonical := func(value any) []byte {
		t.Helper()
		raw, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		result, err := graph.CanonicalizeJSON(raw)
		if err != nil {
			t.Fatal(err)
		}
		return result
	}
	linkPath := strings.TrimPrefix(link.ID, scope)
	if linkPath == link.ID {
		t.Fatalf("unexpected Link Scope: %s", link.ID)
	}
	var dependencyID string
	if err := tx.QueryRowContext(ctx, `SELECT backing_key FROM graph_preview_catalog WHERE path=?`, linkPath).Scan(&dependencyID); err != nil {
		t.Fatal(err)
	}
	changed, err := tx.ExecContext(ctx, `UPDATE dependencies SET depends_on_issue_id=? WHERE id=? AND issue_id=?`, targetID, dependencyID, sourceID)
	if err != nil {
		t.Fatal(err)
	}
	if count, err := changed.RowsAffected(); err != nil || count != 1 {
		t.Fatalf("retarget dependency: rows=%d err=%v", count, err)
	}
	link.Target = targetURL
	linkBytes := canonical(link)
	if _, err := tx.ExecContext(ctx, `UPDATE graph_preview_versions SET snapshot=? WHERE path=? AND version=?`, linkBytes, linkPath, link.Version); err != nil {
		t.Fatal(err)
	}
	sourcePath := strings.TrimPrefix(link.Source, scope)
	if _, err := tx.ExecContext(ctx, `UPDATE graph_preview_issue_versions SET owned=? WHERE path=? AND version=(SELECT revision FROM graph_preview_catalog WHERE path=?)`, canonical([]json.RawMessage{linkBytes}), sourcePath, sourcePath); err != nil {
		t.Fatal(err)
	}
	issue, err := issueops.HydrateIssueOperationResult(ctx, tx, sourceID, false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE issue_versions SET durable_state=? WHERE issue_id=? AND revision=(SELECT current_revision FROM issues WHERE id=?)`, canonical(issue), sourceID, sourceID); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
}

func TestGraphPreviewCheckCycleJSONUsesNativeIssueIDs(t *testing.T) {
	bd := buildBDUnderTest(t)
	work, home := t.TempDir(), t.TempDir()
	const scope = "https://example.invalid/check-cycles/"
	graphPolicyCLI(t, bd, work, home, nil, "", "init", "--graph-mode", "link", "--prefix", "demo", "--scope-url", scope, "--skip-hooks", "--skip-agents", "--non-interactive", "--json")
	a := graphMixedResult[graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "create", "A", "--id", "beads/a", "--json"))
	b := graphMixedResult[graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "create", "B", "--id", "beads/b", "--json"))
	graphPolicyCLI(t, bd, work, home, nil, "", "create", "C", "--id", "beads/c", "--json")
	graphPolicyCLI(t, bd, work, home, nil, "", "dep", "add", "beads/a", "beads/b", "--json")
	bc := graphMixedResult[graphstore.DependencyResult](t, graphPolicyCLI(t, bd, work, home, nil, "", "dep", "add", "beads/b", "beads/c", "--json"))
	graphPreviewRetargetDependencyForCycle(t, work, scope, bc.Link, b.Properties.ID, a.Properties.ID, a.ID)

	stdout, stderr, exit := graphGenerationRun(t, bd, work, home, "graph", "check", "--json")
	if exit != 1 || stderr != "" {
		t.Fatalf("cyclic graph check: exit=%d stderr=%q stdout=%q", exit, stderr, stdout)
	}
	var check graphDependencyCheck
	if err := json.Unmarshal([]byte(stdout), &check); err != nil {
		t.Fatalf("decode cyclic graph check: %v: %s", err, stdout)
	}
	if check.Clean || check.Summary.CycleCount != 1 || !reflect.DeepEqual(check.Cycles, [][]string{{a.Properties.ID, b.Properties.ID}}) {
		t.Fatalf("check JSON must identify native Issue cycle members: %+v", check)
	}
	if specific := graphPolicyCLI(t, bd, work, home, nil, "", "dep", "cycles"); !strings.Contains(specific, a.ID+" → "+b.ID) {
		t.Fatalf("graph-specific cycle projection lost canonical URLs: %q", specific)
	}
}

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
	// The native positional and flag forms share one blocking writer. A repeated
	// pair is a no-op and must not append a second source History version.
	flagged := graphMixedResult[graphstore.DependencyResult](t, graphPolicyCLI(t, bd, work, home, nil, "", "dep", "add", "work", "--blocked-by", "parent", "--no-cycle-check", "--json"))
	if flagged.Changed || flagged.Link.ID != created.Link.ID || flagged.Source.Revision != created.Source.Revision {
		t.Fatalf("--blocked-by changed a repeated blocking pair: %+v", flagged)
	}
	alias := graphMixedResult[graphstore.DependencyResult](t, graphPolicyCLI(t, bd, work, home, nil, "", "dep", "add", "work", "--depends-on", "parent", "--json"))
	if alias.Changed || alias.Link.ID != created.Link.ID || alias.Source.Revision != created.Source.Revision {
		t.Fatalf("--depends-on changed a repeated blocking pair: %+v", alias)
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
		{"create", "Child", "--parent", "original", "--json"},
		{"dep", "add", "original", "replacement", "--type", "parent-child", "--json"},
		{"dep", "add", "original", "replacement", "--type", "duplicates", "--json"},
		{"dep", "add", "original", "replacement", "--type", "supersedes", "--json"},
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
