//go:build cgo

package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/configfile"
	"github.com/steveyegge/beads/internal/storage/doltutil"
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
	var db *sql.DB
	var closeDB func() error
	if cfg.DoltMode == configfile.DoltModeServer {
		user := cfg.DoltServerUser
		if user == "" {
			user = configfile.DefaultDoltServerUser
		}
		db, err = sql.Open("mysql", doltutil.ServerDSN{Host: cfg.DoltServerHost, Port: cfg.DoltServerPort, User: user, Database: cfg.DoltDatabase, Socket: cfg.DoltServerSocket, TLS: cfg.DoltServerTLS}.String())
		if err == nil {
			closeDB = db.Close
		}
	} else {
		db, closeDB, err = embeddeddolt.OpenSQL(ctx, filepath.Join(cfg.GraphWorkspace, "embeddeddolt"), cfg.DoltDatabase, "main")
	}
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
	linkPublicID := strings.TrimPrefix(link.ID, scope)
	if linkPublicID == link.ID {
		t.Fatalf("unexpected Link Scope: %s", link.ID)
	}
	var linkPath, dependencyID string
	if err := tx.QueryRowContext(ctx, `SELECT path, backing_key FROM graph_preview_catalog WHERE public_id=? AND backing='dependency'`, linkPublicID).Scan(&linkPath, &dependencyID); err != nil {
		t.Fatalf("resolve public Dependency %q: %v", linkPublicID, err)
	}
	changed, err := tx.ExecContext(ctx, `UPDATE dependencies SET depends_on_issue_id=? WHERE id=? AND issue_id=?`, targetID, dependencyID, sourceID)
	if err != nil {
		t.Fatal(err)
	}
	if count, err := changed.RowsAffected(); err != nil || count != 1 {
		t.Fatalf("retarget dependency: rows=%d err=%v", count, err)
	}
	// The CLI carries the public revision token; read the exact stored
	// snapshot so this synthetic cycle preserves every retained Link field.
	var retained []byte
	if err := tx.QueryRowContext(ctx, `SELECT snapshot FROM graph_preview_versions WHERE path=? AND version=(SELECT revision FROM graph_preview_catalog WHERE path=?)`, linkPath, linkPath).Scan(&retained); err != nil {
		t.Fatal(err)
	}
	var stored graphstore.LinkRecord
	if err := json.Unmarshal(retained, &stored); err != nil {
		t.Fatal(err)
	}
	if stored.ID != link.ID {
		t.Fatalf("retained Link mismatch: %s != %s", stored.ID, link.ID)
	}
	stored.Target = targetURL
	linkBytes := canonical(stored)
	updated, err := tx.ExecContext(ctx, `UPDATE graph_preview_versions SET snapshot=? WHERE path=? AND version=(SELECT revision FROM graph_preview_catalog WHERE path=?)`, linkBytes, linkPath, linkPath)
	if err != nil {
		t.Fatal(err)
	}
	if count, err := updated.RowsAffected(); err != nil || count != 1 {
		t.Fatalf("retarget retained Link: rows=%d err=%v", count, err)
	}
	var sourcePath string
	if err := tx.QueryRowContext(ctx, `SELECT path FROM graph_preview_catalog WHERE backing_key=? AND backing='issue'`, sourceID).Scan(&sourcePath); err != nil {
		t.Fatalf("resolve native source %q: %v", sourceID, err)
	}
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
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			const scope = "https://example.invalid/check-cycles/"
			graphPolicyCLI(t, bd, work, home, nil, "", graphSetupInitArgs(t, engine, scope, "--prefix", "demo", "--skip-hooks", "--skip-agents")...)
			a := graphMixedResult[graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "create", "A", "--id", "beads/a", "--force", "--json"))
			b := graphMixedResult[graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "create", "B", "--id", "beads/b", "--force", "--json"))
			graphPolicyCLI(t, bd, work, home, nil, "", "create", "C", "--id", "beads/c", "--force", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "", "dep", "add", scope+"beads/a", scope+"beads/b", "--json")
			bc := graphMixedResult[graphstore.DependencyResult](t, graphPolicyCLI(t, bd, work, home, nil, "", "dep", "add", scope+"beads/b", scope+"beads/c", "--json"))
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
		})
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
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			const scope = "https://example.invalid/tree-directions/"
			graphPolicyCLI(t, bd, work, home, nil, "", graphSetupInitArgs(t, engine, scope, "--prefix", "demo", "--skip-hooks", "--skip-agents")...)
			for _, issue := range []struct{ id, title string }{
				{"beads/root", "Root"},
				{"beads/blocker", "Blocker"},
				{"beads/dependent", "Dependent"},
			} {
				graphPolicyCLI(t, bd, work, home, nil, "", "create", issue.title, "--id", issue.id, "--force", "--json")
			}
			graphPolicyCLI(t, bd, work, home, nil, "", "dep", "add", scope+"beads/root", scope+"beads/blocker", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "", "dep", "add", scope+"beads/dependent", scope+"beads/root", "--json")

			got := strings.TrimSpace(graphPolicyCLI(t, bd, work, home, nil, "", "dep", "tree", "beads/root", "--direction", "both"))
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
		})
	}
}

func TestGraphPreviewStructureReads(t *testing.T) {
	bd := buildBDUnderTest(t)
	work, home := t.TempDir(), t.TempDir()
	const scope = "https://example.invalid/structure/"
	graphPolicyCLI(t, bd, work, home, nil, "", "init", "--graph-mode", "link", "--prefix", "demo", "--scope-url", scope, "--skip-hooks", "--skip-agents", "--non-interactive", "--json")
	graphPolicyCLI(t, bd, work, home, nil, "", "create", "Parent", "--graph-path", "beads/parent", "--id", "beads/parent", "--force", "--json")
	graphPolicyCLI(t, bd, work, home, nil, "", "create", "Work", "--graph-path", "beads/work", "--id", "beads/work", "--force", "--json")
	graphPolicyCLI(t, bd, work, home, nil, "", "create", "Isolated", "--graph-path", "beads/isolated", "--id", "beads/isolated", "--force", "--json")

	// Only native parent-child edges count. A blocking dependency does not
	// make an Issue a child.
	if got := graphPolicyCLI(t, bd, work, home, nil, "", "children", scope+"beads/parent", "--json"); strings.TrimSpace(got) != "[]" {
		t.Fatalf("empty native children must retain the ordinary JSON array shape: %q", got)
	}
	graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "children", "parent", "--pretty", "--json")
	graphPolicyCLI(t, bd, work, home, nil, "not_found", "children", "missing", "--json")

	created := graphMixedResult[graphstore.DependencyResult](t, graphPolicyCLI(t, bd, work, home, nil, "", "dep", "add", scope+"beads/work", scope+"beads/parent", "--json"))
	if created.Link.Source != scope+"beads/work" || created.Link.Target != scope+"beads/parent" {
		t.Fatalf("dependency endpoints lost canonical IDs: %+v", created.Link)
	}
	// The native positional and flag forms share one blocking writer. A repeated
	// pair is a no-op and must not append a second source History version.
	flagged := graphMixedResult[graphstore.DependencyResult](t, graphPolicyCLI(t, bd, work, home, nil, "", "dep", "add", scope+"beads/work", "--blocked-by", scope+"beads/parent", "--no-cycle-check", "--json"))
	if flagged.Changed || flagged.Link.ID != created.Link.ID || flagged.Source.Revision != created.Source.Revision {
		t.Fatalf("--blocked-by changed a repeated blocking pair: %+v", flagged)
	}
	alias := graphMixedResult[graphstore.DependencyResult](t, graphPolicyCLI(t, bd, work, home, nil, "", "dep", "add", scope+"beads/work", "--depends-on", scope+"beads/parent", "--json"))
	if alias.Changed || alias.Link.ID != created.Link.ID || alias.Source.Revision != created.Source.Revision {
		t.Fatalf("--depends-on changed a repeated blocking pair: %+v", alias)
	}
	workHistory := graphPolicyCLI(t, bd, work, home, nil, "", "versions", "work", "--json")
	parentHistory := graphPolicyCLI(t, bd, work, home, nil, "", "versions", "parent", "--json")
	// Native dep add treats an empty alias as absent, and gives --blocked-by
	// precedence when both nonempty aliases are present. These repeat the
	// existing edge, so neither form may append source History.
	for _, flags := range [][]string{
		{scope + "beads/work", scope + "beads/parent", "--blocked-by=", "--json"},
		{scope + "beads/work", "--blocked-by", scope + "beads/parent", "--depends-on", scope + "beads/isolated", "--json"},
	} {
		result := graphMixedResult[graphstore.DependencyResult](t, graphPolicyCLI(t, bd, work, home, nil, "", append([]string{"dep", "add"}, flags...)...))
		if result.Changed || result.Link.ID != created.Link.ID {
			t.Fatalf("native alias precedence changed the blocking pair for %v: %+v", flags, result)
		}
	}
	for _, flags := range [][]string{
		{"work", "parent", "--blocked-by", "isolated", "--json"},
		{"work", "parent", "--depends-on", "isolated", "--json"},
		{"work", "parent", "--blocked-by", "parent", "--depends-on", "isolated", "--json"},
	} {
		graphPolicyCLI(t, bd, work, home, nil, "stderr:cannot use both positional depends-on-id and --blocked-by/--depends-on flag", append([]string{"dep", "add"}, flags...)...)
	}
	graphPolicyCLI(t, bd, work, home, nil, "stderr:requires 2 arg(s)", "dep", "add", "work", "--blocked-by=", "--json")
	if after := graphPolicyCLI(t, bd, work, home, nil, "", "versions", "work", "--json"); after != workHistory {
		t.Fatalf("refused dependency arguments or alias no-op appended source History\nbefore: %s\nafter: %s", workHistory, after)
	}
	if after := graphPolicyCLI(t, bd, work, home, nil, "", "versions", "parent", "--json"); after != parentHistory {
		t.Fatalf("refused dependency arguments or alias no-op appended target History\nbefore: %s\nafter: %s", parentHistory, after)
	}
	for _, tc := range []struct {
		selector, direction string
	}{
		{scope + "beads/work", "down"},
		{scope + "beads/parent", "up"},
	} {
		got := graphPolicyCLI(t, bd, work, home, nil, "", "dep", "list", tc.selector, "--direction", tc.direction)
		if !strings.Contains(got, "beads/work depends on beads/parent") {
			t.Fatalf("%s list lost blocking dependency: %q", tc.direction, got)
		}
	}
	if got := graphPolicyCLI(t, bd, work, home, nil, "", "children", scope+"beads/parent", "--json"); strings.TrimSpace(got) != "[]" {
		t.Fatalf("blocking dependency leaked into native children: %q", got)
	}
	graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "dep", "list", "work", "--type", "parent-child", "--json")
	graphPolicyCLI(t, bd, work, home, nil, "invalid_selector", "dep", "list", "work", "--direction", "both", "--json")
	graphPolicyCLI(t, bd, work, home, nil, "invalid_selector", "dep", "list", "https://other.invalid/beads/work", "--json")
	if got := graphPolicyCLI(t, bd, work, home, nil, "", "dep", "list", scope+"beads/work"); !strings.Contains(got, "beads/work depends on beads/parent") {
		t.Fatalf("missing human dependency summary: %q", got)
	}
	taskType, _ := graphstore.NativeIssueTypeURL(scope, "task")
	blocksType, _ := graphstore.NativeDependencyTypeURL(scope, "blocks")
	if view := graphPolicyCLI(t, bd, work, home, nil, "", "graph", scope+"beads/work"); !strings.Contains(view, "beads/work ["+strings.TrimPrefix(taskType, scope)+"]") || !strings.Contains(view, "2 Beads, 1 Links") || !strings.Contains(view, "└── ["+strings.TrimPrefix(blocksType, scope)+"]") {
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
	if tree := graphPolicyCLI(t, bd, work, home, nil, "", "dep", "tree", scope+"beads/work"); !strings.Contains(tree, "beads/work") || !strings.Contains(tree, "beads/parent") {
		t.Fatalf("blocking dependency tree lost its nodes: %q", tree)
	}
	if tree := graphPolicyCLI(t, bd, work, home, nil, "", "dep", "tree", scope+"beads/isolated"); !strings.Contains(tree, "has no dependencies") {
		t.Fatalf("isolated Issue falsely reported a dependency tree: %q", tree)
	}
	if tree := graphPolicyCLI(t, bd, work, home, nil, "", "dep", "tree", scope+"beads/isolated", "--direction", "up"); !strings.Contains(tree, "has no dependents") {
		t.Fatalf("isolated Issue falsely reported dependents: %q", tree)
	}
	if tree := graphPolicyCLI(t, bd, work, home, nil, "", "dep", "tree", scope+"beads/work", "--direction", "both"); !strings.Contains(tree, "beads/work") || !strings.Contains(tree, "beads/parent") || strings.Index(tree, "beads/parent") < strings.Index(tree, "beads/work") {
		t.Fatalf("both-direction dependency tree must start at its root: %q", tree)
	}
	if cycles := graphPolicyCLI(t, bd, work, home, nil, "", "dep", "cycles"); !strings.Contains(cycles, "No blocking dependency cycles") {
		t.Fatalf("acyclic blocking graph reported cycles: %q", cycles)
	}
	generic := graphSelectedTestSummary(t, graphPolicyCLI(t, bd, work, home, nil, "", "graph", scope+"beads/work", "--json"))
	if generic.Root != scope+"beads/work" || len(generic.Links) != 1 {
		t.Fatalf("generic traversal changed: %+v", generic)
	}
	for _, args := range [][]string{
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
	if got := graphPolicyCLI(t, bd, work, home, []string{"BEADS_MAX_ROWS=1"}, "", "graph", scope+"beads/work", "--max-rows=0"); !strings.Contains(got, "2 Beads, 1 Links") {
		t.Fatalf("explicit zero did not override graph row cap: %q", got)
	}
	if got := graphPolicyCLI(t, bd, work, home, []string{"BEADS_MAX_ROWS=1"}, "", "dep", "tree", scope+"beads/work", "--max-rows=0"); !strings.Contains(got, "beads/parent") {
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

func TestGraphPreviewChildrenNative(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			const scope = "https://example.invalid/native-children/"
			graphPolicyCLI(t, bd, work, home, nil, "", graphSetupInitArgs(t, engine, scope, "--prefix", "demo", "--skip-hooks", "--skip-agents")...)
			var childID, parentID string
			for _, id := range []string{"beads/parent", "beads/child", "beads/parent.lookalike"} {
				created := graphMixedResult[graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "create", id, "--id", id, "--force", "--json"))
				switch id {
				case "beads/parent":
					parentID = created.Properties.ID
				case "beads/child":
					childID = created.Properties.ID
				}
			}
			graphPolicyCLI(t, bd, work, home, nil, "", "dep", "add", scope+"beads/child", scope+"beads/parent", "--type", "parent-child", "--json")
			var children []struct {
				ID     string  `json:"id"`
				Parent *string `json:"parent"`
				Title  string  `json:"title"`
			}
			got := graphPolicyCLI(t, bd, work, home, nil, "", "children", "beads/parent", "--json")
			if err := json.Unmarshal([]byte(got), &children); err != nil {
				t.Fatalf("ordinary children JSON shape: %v: %s", err, got)
			}
			if len(children) != 1 || children[0].ID != childID || children[0].Parent == nil || *children[0].Parent != parentID || children[0].Title != "beads/child" {
				t.Fatalf("native edge membership or ordinary row shape: %+v", children)
			}
			if text := graphPolicyCLI(t, bd, work, home, nil, "", "children", scope+"beads/parent"); !strings.Contains(text, "beads/child") || strings.Contains(text, "parent.lookalike") {
				t.Fatalf("human children omitted native child or inferred dotted ID: %q", text)
			}
		})
	}
}

func TestGraphPreviewEpicNativeProgressAndClose(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			const scope = "https://example.invalid/native-epic/"
			call := func(code string, args ...string) string {
				return graphPolicyCLI(t, bd, work, home, nil, code, args...)
			}
			call("", graphSetupInitArgs(t, engine, scope, "--prefix", "demo", "--skip-hooks", "--skip-agents")...)
			call("", "create", "Release epic", "--id", "beads/epic", "--force", "--type", "epic", "--json")
			call("", "create", "Child work", "--id", "beads/child", "--force", "--json")
			call("", "create", "Dotted lookalike", "--id", "beads/epic.lookalike", "--force", "--json")
			call("", "dep", "add", scope+"beads/child", scope+"beads/epic", "--type", "parent-child", "--json")
			decode := func(raw string) []struct {
				Epic struct {
					ID string `json:"id"`
				} `json:"epic"`
				TotalChildren    int  `json:"total_children"`
				ClosedChildren   int  `json:"closed_children"`
				EligibleForClose bool `json:"eligible_for_close"`
			} {
				t.Helper()
				var statuses []struct {
					Epic struct {
						ID string `json:"id"`
					} `json:"epic"`
					TotalChildren    int  `json:"total_children"`
					ClosedChildren   int  `json:"closed_children"`
					EligibleForClose bool `json:"eligible_for_close"`
				}
				if err := json.Unmarshal([]byte(raw), &statuses); err != nil {
					t.Fatalf("epic status JSON: %v: %s", err, raw)
				}
				return statuses
			}
			statuses := decode(call("", "epic", "status", "--json"))
			if len(statuses) != 1 || statuses[0].Epic.ID != "beads/epic" || statuses[0].TotalChildren != 1 || statuses[0].ClosedChildren != 0 || statuses[0].EligibleForClose {
				t.Fatalf("native child membership: %+v", statuses)
			}
			if got := call("", "epic", "status", "--eligible-only", "--json"); strings.TrimSpace(got) != "[]" {
				t.Fatalf("open child was eligible: %s", got)
			}
			if got := call("", "epic", "close-eligible", "--json"); !strings.Contains(got, `"count": 0`) {
				t.Fatalf("open child permitted epic close: %s", got)
			}
			call("", "close", scope+"beads/child", "--reason", "Done", "--json")
			statuses = decode(call("", "epic", "status", "--eligible-only", "--json"))
			if len(statuses) != 1 || statuses[0].ClosedChildren != 1 || !statuses[0].EligibleForClose {
				t.Fatalf("completed child did not qualify epic: %+v", statuses)
			}
			before := call("", "versions", "beads/epic", "--json")
			if got := call("", "epic", "close-eligible", "--dry-run", "--json"); !strings.Contains(got, `"eligible_for_close": true`) || call("", "versions", "beads/epic", "--json") != before {
				t.Fatalf("dry-run changed epic or lost eligibility: %s", got)
			}
			if got := call("", "epic", "close-eligible", "--json"); !strings.Contains(got, `"count": 1`) || !strings.Contains(got, `"beads/epic"`) {
				t.Fatalf("eligible epic did not close: %s", got)
			}
			if got := call("", "epic", "status", "--json"); strings.TrimSpace(got) != "[]" {
				t.Fatalf("closed epic still open: %s", got)
			}
		})
	}
}

func TestGraphPreviewEpicCloseEligibleActorAndPartialFailure(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			const scope = "https://example.invalid/epic-actors/"
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			call(graphSetupInitArgs(t, engine, scope, "--prefix", "demo", "--skip-hooks", "--skip-agents")...)
			for _, owner := range []string{"alice", "bob"} {
				epic := "beads/" + owner + "-epic"
				child := "beads/" + owner + "-child"
				call("create", owner+" epic", "--id", epic, "--force", "--type", "epic", "--assignee", owner, "--actor", owner)
				call("create", owner+" child", "--id", child, "--force")
				call("dep", "add", scope+child, scope+epic, "--type", "parent-child")
				call("close", scope+child, "--reason", "Done")
			}
			beforeAlice := call("versions", "beads/alice-epic")
			beforeBob := call("versions", "beads/bob-epic")
			checkFailure := func(actor string, wantClosed int) {
				t.Helper()
				stdout, stderr, exit := graphGenerationRun(t, bd, work, home, "epic", "close-eligible", "--actor", actor, "--json")
				if exit == 0 || stdout != "" {
					t.Fatalf("actor %s: close failures looked successful: exit=%d stdout=%q stderr=%q", actor, exit, stdout, stderr)
				}
				var diagnostic struct {
					Code, Message string
				}
				if err := json.Unmarshal([]byte(stderr), &diagnostic); err != nil || diagnostic.Code != "invalid_properties" ||
					!strings.Contains(diagnostic.Message, fmt.Sprintf("closed %d of 2 eligible epics", wantClosed)) {
					t.Fatalf("actor %s: missing typed partial-failure receipt: %+v, %v", actor, diagnostic, err)
				}
				if wantClosed == 1 && !strings.Contains(diagnostic.Message, "beads/alice-epic") {
					t.Fatalf("partial failure omitted successfully closed ID: %+v", diagnostic)
				}
			}
			checkFailure("stranger", 0)
			if call("versions", "beads/alice-epic") != beforeAlice || call("versions", "beads/bob-epic") != beforeBob {
				t.Fatal("all-refused epic close appended History")
			}
			checkFailure("alice", 1)
			_, _, aliceBeforeRows, _ := graphVersionsListed(t, beforeAlice)
			_, _, aliceAfterRows, _ := graphVersionsListed(t, call("versions", "beads/alice-epic"))
			if len(aliceAfterRows) != len(aliceBeforeRows)+1 {
				t.Fatalf("successful sibling close must append one Issue History version: %d -> %d", len(aliceBeforeRows), len(aliceAfterRows))
			}
			if got := graphMixedResult[graphstore.IssueRecord](t, call("show", scope+"beads/alice-epic", "--format", "graph-json")); got.Properties.Status != "closed" || got.Attribution.Actor != "alice" {
				t.Fatalf("explicit actor lost authority or provenance: %+v", got)
			}
			if got := graphMixedResult[graphstore.IssueRecord](t, call("show", scope+"beads/bob-epic", "--format", "graph-json")); got.Properties.Status == "closed" {
				t.Fatalf("another actor closed Bob's epic: %+v", got)
			}
			if call("versions", "beads/bob-epic") != beforeBob {
				t.Fatal("refused sibling appended History")
			}
			if got := call("epic", "close-eligible", "--actor", "bob"); !strings.Contains(got, `"count": 1`) {
				t.Fatalf("assigned actor could not finish remaining epic: %s", got)
			}
			if got := graphMixedResult[graphstore.IssueRecord](t, call("show", scope+"beads/bob-epic", "--format", "graph-json")); got.Properties.Status != "closed" || got.Attribution.Actor != "bob" {
				t.Fatalf("second epic close lost actor provenance: %+v", got)
			}
		})
	}
}

func TestGraphPreviewDepRemovePairAndNoop(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			const scope = "https://example.invalid/dep-remove/"
			call := func(code string, args ...string) string {
				return graphPolicyCLI(t, bd, work, home, nil, code, args...)
			}
			call("", graphSetupInitArgs(t, engine, scope, "--prefix", "demo", "--skip-hooks", "--skip-agents")...)
			created := map[string]graphstore.IssueRecord{}
			for _, id := range []string{"work", "first", "second"} {
				created[id] = graphMixedResult[graphstore.IssueRecord](t, call("", "create", id, "--id", "beads/"+id, "--force", "--json"))
			}
			added := graphMixedResult[graphstore.DependencyResult](t, call("", "dep", "add", scope+"beads/work", scope+"beads/first", "--id", "links/first", "--json"))
			firstHistory := call("", "versions", "beads/first", "--json")
			workHistory := call("", "versions", "beads/work", "--json")
			type removalJSON struct {
				Status      string `json:"status"`
				Removed     bool   `json:"removed"`
				IssueID     string `json:"issue_id"`
				DependsOnID string `json:"depends_on_id"`
			}
			decode := func(raw string) removalJSON {
				t.Helper()
				var result removalJSON
				if err := json.Unmarshal([]byte(raw), &result); err != nil {
					t.Fatalf("dep remove did not preserve native flat JSON: %v: %s", err, raw)
				}
				return result
			}
			missing := decode(call("", "dep", "remove", scope+"beads/work", scope+"beads/second", "--json"))
			if missing.Status != "not_found" || missing.Removed || missing.IssueID != added.Source.Properties.ID || missing.DependsOnID == "" || call("", "versions", "beads/work", "--json") != workHistory {
				t.Fatalf("absent pair was not a native no-op: %+v", missing)
			}
			removed := decode(call("", "dep", "remove", scope+"beads/work", scope+"beads/first", "--json"))
			if removed.Status != "removed" || !removed.Removed || removed.IssueID != added.Source.Properties.ID || removed.DependsOnID != created["first"].Properties.ID {
				t.Fatalf("pair removal lost native JSON fields: %+v", removed)
			}
			if _, _, before, _ := graphVersionsListed(t, workHistory); len(before) == 0 {
				t.Fatal("source has no creation/add History")
			} else if _, _, after, _ := graphVersionsListed(t, call("", "versions", "beads/work", "--json")); len(after) != len(before)+1 {
				t.Fatalf("pair removal should append exactly one source History version: %d → %d", len(before), len(after))
			}
			call("gone", "show", scope+"links/first", "--json")
			afterRemoval := call("", "versions", "beads/work", "--json")
			again := decode(call("", "dep", "rm", scope+"beads/work", scope+"beads/first", "--json"))
			if again.Status != "not_found" || again.Removed || call("", "versions", "beads/work", "--json") != afterRemoval {
				t.Fatalf("repeated pair removal changed source History: %+v", again)
			}
			if got := call("", "versions", "beads/first", "--json"); got != firstHistory {
				t.Fatalf("pair removal changed target History\nbefore: %s\nafter: %s", firstHistory, got)
			}
			// Explicit private graph paths differ from native public IDs. The native
			// IDs printed by familiar dep output must resolve to these mapped Issues,
			// while a colliding graph Bead must make the bare spelling ambiguous.
			source := graphMixedResult[graphstore.IssueRecord](t, call("", "create", "Custom source", "--graph-path", "beads/custom-source-path", "--json"))
			target := graphMixedResult[graphstore.IssueRecord](t, call("", "create", "Custom target", "--graph-path", "beads/custom-target-path", "--json"))
			if source.Properties.ID == "custom-source-path" || target.Properties.ID == "custom-target-path" {
				t.Fatalf("fixture must distinguish canonical path from native backing ID: %s %s", source.Properties.ID, target.Properties.ID)
			}
			call("", "dep", "add", source.ID, target.ID, "--id", "links/custom-edge", "--json")
			native := decode(call("", "dep", "remove", source.Properties.ID, target.Properties.ID, "--json"))
			if !native.Removed || native.IssueID != source.Properties.ID || native.DependsOnID != target.Properties.ID {
				t.Fatalf("bare native IDs did not resolve the checked Issue pair: %+v", native)
			}
			call("", "dep", "add", source.ID, target.ID, "--id", "links/custom-readded", "--json")
			call("", "remember", "Colliding Memory", "--id", "beads/"+source.Properties.ID, "--json")
			memoryURL := scope + "beads/" + source.Properties.ID
			beforeMemory := call("", "show", memoryURL, "--json")
			bare := decode(call("", "dep", "remove", source.Properties.ID, target.Properties.ID, "--json"))
			if !bare.Removed || bare.IssueID != source.Properties.ID || bare.DependsOnID != target.Properties.ID || call("", "show", memoryURL, "--json") != beforeMemory {
				t.Fatalf("opaque public Issue ID did not win over another Bead's graph path: %+v", bare)
			}
			call("", "dep", "add", source.ID, target.ID, "--id", "links/custom-explicit", "--json")
			explicit := decode(call("", "dep", "remove", source.ID, target.ID, "--json"))
			if !explicit.Removed || explicit.IssueID != source.Properties.ID || explicit.DependsOnID != target.Properties.ID {
				t.Fatalf("explicit canonical selectors lost authority beside the public-ID collision: %+v", explicit)
			}
		})
	}
}

func TestGraphPreviewIncompleteIssueStructureRefusesWithoutMutation(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			const scope = "https://example.invalid/unmapped-structure/"
			graphPolicyCLI(t, bd, work, home, nil, "", graphSetupInitArgs(t, engine, scope, "--prefix", "demo", "--skip-hooks", "--skip-agents")...)
			graphPolicyCLI(t, bd, work, home, nil, "", "create", "Original", "--id", "beads/original", "--force", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "", "create", "Replacement", "--id", "beads/replacement", "--force", "--json")
			originalHistory := graphPolicyCLI(t, bd, work, home, nil, "", "versions", "beads/original", "--json")
			replacementHistory := graphPolicyCLI(t, bd, work, home, nil, "", "versions", "beads/replacement", "--json")
			before := legacyUpgradeTreeDigest(t, work)

			for _, args := range [][]string{
				{"duplicate", "beads/original", "--of", "beads/replacement", "--json"},
				{"supersede", "beads/original", "--with", "beads/replacement", "--json"},
			} {
				graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", args...)
				if after := legacyUpgradeTreeDigest(t, work); after != before {
					t.Fatalf("incomplete Issue relationship command %v changed workspace", args)
				}
			}
			if got := graphPolicyCLI(t, bd, work, home, nil, "", "versions", "beads/original", "--json"); got != originalHistory {
				t.Fatalf("refused relationship commands changed original History\nbefore: %s\nafter: %s", originalHistory, got)
			}
			if got := graphPolicyCLI(t, bd, work, home, nil, "", "versions", "beads/replacement", "--json"); got != replacementHistory {
				t.Fatalf("refused relationship commands changed replacement History\nbefore: %s\nafter: %s", replacementHistory, got)
			}
		})
	}
}
