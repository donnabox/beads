//go:build cgo

package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/configfile"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
	publicops "github.com/steveyegge/beads/issueops"
)

// graphListWorkspace initializes a graph workspace on one engine through normal
// initialization, so every later command is a separate installed process.
func graphListWorkspace(t *testing.T, bd, engine, scope string) (work, home string) {
	t.Helper()
	work, home = t.TempDir(), t.TempDir()
	args := []string{"init", "--graph-mode", "link", "--scope-url", scope, "--skip-hooks", "--skip-agents", "--non-interactive", "--json"}
	if engine == "server" {
		port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
		if port == "" {
			t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for ordinary shared-server CLI qualification")
		}
		args = append(args, "--server", "--external", "--server-host", "127.0.0.1", "--server-port", port, "--server-user", "root")
	}
	graphPolicyCLI(t, bd, work, home, nil, "", args...)
	return work, home
}

// graphListConfig appends workspace configuration. Initialization writes none in
// a graph workspace, and the graph CLI has no command that sets any.
func graphListConfig(t *testing.T, work, yaml string) {
	t.Helper()
	path := filepath.Join(work, ".beads", "config.yaml")
	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	body := string(existing)
	if body != "" && !strings.HasSuffix(body, "\n") {
		body += "\n"
	}
	if err := os.WriteFile(path, []byte(body+yaml), 0o600); err != nil {
		t.Fatal(err)
	}
}

// graphListSeedIssue creates an Issue through the controlled writer between two
// installed processes, for a state no admitted command can author. Close it
// before the next process opens the workspace, embedded mode included.
func graphListSeedIssue(t *testing.T, work, path string, request publicops.CreateRequest) {
	t.Helper()
	cfg, err := configfile.LoadForDiscovery(filepath.Join(work, ".beads"))
	if err != nil || cfg == nil {
		t.Fatalf("read initialized binding: %v", err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
	defer cancel()
	s, err := graphstore.OpenExisting(ctx, graphstore.Options{
		Backend: cfg.DoltMode, DataDir: filepath.Join(cfg.GraphWorkspace, "embeddeddolt"), Database: cfg.DoltDatabase, Branch: "main",
		Binding:    graphstore.Binding{WorkspaceID: cfg.GraphWorkspace, ScopeURL: cfg.GraphScopeURL, AuthorityID: cfg.GraphAuthorityID, SchemaVersion: cfg.GraphSchemaVersion},
		ServerHost: cfg.DoltServerHost, ServerPort: cfg.DoltServerPort, ServerUser: cfg.DoltServerUser, ServerSocket: cfg.DoltServerSocket, ServerTLS: cfg.DoltServerTLS,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := s.Close(); err != nil {
			t.Error(err)
		}
	}()
	request.Issue.ID = path
	request.ForceIDPrefix = true
	if _, err := s.CreateIssue(ctx, path, request); err != nil {
		t.Fatal(err)
	}
}

// graphListBeads runs the structured Bead list and returns its page: the
// canonical Bead paths, without the scope, in the order listed.
func graphListBeads(t *testing.T, bd, work, home, scope string, extra ...string) (paths []string, more bool) {
	t.Helper()
	beads, more := graphListDecode(t, graphPolicyCLI(t, bd, work, home, nil, "", append([]string{"list", "--format", "records-json"}, extra...)...))
	for _, bead := range beads {
		paths = append(paths, strings.TrimPrefix(bead.ID, scope))
	}
	return paths, more
}

func graphListSorted(paths []string) []string {
	sorted := slices.Clone(paths)
	slices.Sort(sorted)
	return sorted
}

func TestGraphPreviewListReadyWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			const scope = "https://example.invalid/list-ready/"
			work, home := graphListWorkspace(t, bd, engine, scope)
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", args...)
			}
			call("create", "Blocked", "--id", "beads/blocked", "--force")
			call("create", "Blocker", "--id", "beads/blocker", "--force")
			call("create", "Other", "--id", "beads/other", "--force")
			call("remember", "context", "--id", "beads/memory")
			call("dep", "add", "beads/blocked", "beads/blocker")

			paths, more := graphListBeads(t, bd, work, home, scope, "--ready", "--sort=id")
			if !slices.Equal(paths, []string{"beads/blocker", "beads/other"}) || more {
				t.Fatalf("blocker-aware ready paths=%v more=%t", paths, more)
			}
			paths, more = graphListBeads(t, bd, work, home, scope, "--ready", "--sort=id", "--limit=1")
			if !slices.Equal(paths, []string{"beads/blocker"}) || !more {
				t.Fatalf("ready page paths=%v more=%t", paths, more)
			}
			var legacy []types.IssueWithCounts
			if err := json.Unmarshal([]byte(call("list", "--json", "--ready", "--sort=id")), &legacy); err != nil {
				t.Fatal(err)
			}
			if len(legacy) != 2 || legacy[0].ID != "beads/blocker" || legacy[1].ID != "beads/other" {
				t.Fatalf("legacy ready Issue IDs: %+v", legacy)
			}
			graphPatchProcess(t, bd, work, home, nil, "invalid_selector", 90*time.Second, "list", "--ready", "--title", "Blocked", "--json")
			if got := strings.TrimSpace(call("list", "--ready", "--parent", "beads/blocker", "--json")); got != "[]" {
				t.Fatalf("ready parent without descendants = %q, want []", got)
			}
		})
	}
}

func TestGraphPreviewListReadyParentWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			const scope = "https://example.invalid/list-ready-parent/"
			work, home := graphListWorkspace(t, bd, engine, scope)
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", args...)
			}
			for _, path := range []string{"beads/root", "beads/mid", "beads/grand", "beads/other"} {
				call("create", path, "--id", path, "--force")
			}
			call("dep", "add", "beads/mid", "beads/root", "--type", "parent-child")
			call("dep", "add", "beads/grand", "beads/mid", "--type", "parent-child")
			call("close", "beads/mid", "--force")
			paths, more := graphListBeads(t, bd, work, home, scope, "--ready", "--parent", "beads/root", "--sort=id")
			if !slices.Equal(paths, []string{"beads/grand"}) || more {
				t.Fatalf("ready transitive parent paths=%v more=%t", paths, more)
			}
			aliasPaths, aliasMore := graphListBeads(t, bd, work, home, scope, "--ready", "--filter-parent", "beads/root", "--sort=id")
			if !slices.Equal(aliasPaths, paths) || aliasMore != more {
				t.Fatalf("ready parent alias paths=%v more=%t", aliasPaths, aliasMore)
			}
			var legacy []types.IssueWithCounts
			if err := json.Unmarshal([]byte(call("list", "--json", "--ready", "--parent", "beads/root")), &legacy); err != nil {
				t.Fatal(err)
			}
			if len(legacy) != 1 || legacy[0].ID != "beads/grand" {
				t.Fatalf("legacy ready parent rows: %+v", legacy)
			}
			text := call("list", "--ready", "--parent", "beads/root")
			if !strings.Contains(text, "beads/grand") || strings.Contains(text, "beads/mid") || strings.Contains(text, "beads/other") {
				t.Fatalf("human ready parent row = %q", text)
			}
			if explicit := call("list", "--ready", "--parent", "beads/root", "--tree=true"); explicit != text {
				t.Fatalf("explicit tree flag must preserve shipping flat ready rows: %q versus %q", explicit, text)
			}
		})
	}
}

func TestGraphPreviewListIDWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			const scope = "https://example.invalid/list-ids/"
			work, home := graphListWorkspace(t, bd, engine, scope)
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", args...)
			}
			first := graphMixedResult[graphstore.IssueRecord](t, call("create", "First", "--id", "beads/first", "--json", "--force"))
			second := graphMixedResult[graphstore.IssueRecord](t, call("create", "Second", "--id", "beads/second", "--json", "--force"))
			rootless := graphMixedResult[graphstore.IssueRecord](t, call("create", "Rootless", "--graph-path", "beads/private-rootless", "--json"))
			call("remember", "context", "--id", "beads/memory", "--json")
			paths, more := graphListBeads(t, bd, work, home, scope, "--id", first.Properties.ID)
			if !slices.Equal(paths, []string{"beads/first"}) || more {
				t.Fatalf("native --id selected %v, more=%t", paths, more)
			}
			paths, more = graphListBeads(t, bd, work, home, scope, "--id", rootless.ID)
			if !slices.Equal(paths, []string{rootless.Properties.ID}) || more {
				t.Fatalf("canonical public ID selected %v, more=%t; want %q", paths, more, rootless.Properties.ID)
			}
			paths, more = graphListBeads(t, bd, work, home, scope, "--id", scope+"beads/private-rootless")
			if len(paths) != 0 || more {
				t.Fatalf("private path URL must not select rootless public ID: %v, more=%t", paths, more)
			}
			paths, more = graphListBeads(t, bd, work, home, scope, "--id", "beads/first,beads/second", "--sort=id", "--limit=1")
			if !slices.Equal(paths, []string{"beads/first"}) || !more {
				t.Fatalf("ID set must filter before paging: %v, more=%t", paths, more)
			}
			paths, more = graphListBeads(t, bd, work, home, scope, "--id", second.Properties.ID, "--id", first.Properties.ID, "--sort=id")
			if !slices.Equal(paths, []string{"beads/first", "beads/second"}) || more {
				t.Fatalf("repeated public IDs selected %v, more=%t", paths, more)
			}
			for _, missing := range []string{"beads/missing", "beads/memory"} {
				paths, more = graphListBeads(t, bd, work, home, scope, "--id", missing)
				if len(paths) != 0 || more {
					t.Fatalf("non-Issue or missing ID %q selected %v, more=%t", missing, paths, more)
				}
			}
			var rows []struct {
				ID string `json:"id"`
			}
			output := call("list", "--id", "beads/second", "--json")
			if err := json.Unmarshal([]byte(output), &rows); err != nil || len(rows) != 1 || rows[0].ID != "beads/second" {
				t.Fatalf("legacy JSON ID selection: %s (%v)", output, err)
			}
			rows = nil
			output = call("list", "--id", "beads/missing", "--json")
			if err := json.Unmarshal([]byte(output), &rows); err != nil || rows == nil || len(rows) != 0 {
				t.Fatalf("missing legacy JSON ID selection: %s (%v)", output, err)
			}
			human := call("list", "--id", "beads/first", "--flat")
			if !strings.Contains(human, "First") || strings.Contains(human, "Second") {
				t.Fatalf("human ID selection: %s", human)
			}
			long := call("list", "--id", "beads/first", "--long")
			if !strings.Contains(long, "First") || strings.Contains(long, "Second") {
				t.Fatalf("long ID selection: %s", long)
			}
		})
	}
}

func TestGraphPreviewListNoParentWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			const scope = "https://example.invalid/no-parent/"
			work, home := graphListWorkspace(t, bd, engine, scope)
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", args...)
			}
			parent := graphMixedResult[graphstore.IssueRecord](t, call("create", "Parent", "--id", "beads/parent", "--json", "--force"))
			child := graphMixedResult[graphstore.IssueRecord](t, call("create", "Child", "--parent", "beads/parent", "--force", "--json"))
			call("create", "Grandchild", "--parent", child.Properties.ID, "--force", "--json")
			call("create", "Dotted without an edge", "--id", "beads/parent.9", "--json", "--force")
			call("remember", "memory body", "--id", "beads/memory", "--title", "Memory", "--json")

			var rows []struct {
				ID string `json:"id"`
			}
			output := call("list", "--no-parent", "--json", "--sort=id")
			if err := json.Unmarshal([]byte(output), &rows); err != nil {
				t.Fatalf("decode no-parent Issue list: %v: %s", err, output)
			}
			ids := make([]string, 0, len(rows))
			for _, row := range rows {
				ids = append(ids, row.ID)
			}
			if !slices.Equal(ids, []string{"beads/parent", "beads/parent.9"}) {
				t.Fatalf("no-parent must select only Issues without a parent edge, including an edge-less dotted ID: %v", ids)
			}
			paths, more := graphListBeads(t, bd, work, home, scope, "--no-parent", "--sort=id")
			if !slices.Equal(paths, ids) || more {
				t.Fatalf("structured no-parent Issue list = %v, more=%t; want %v", paths, more, ids)
			}
			human := call("list", "--no-parent", "--flat")
			if !strings.Contains(human, "beads/parent") || !strings.Contains(human, "beads/parent.9") || strings.Contains(human, "Child") || strings.Contains(human, "beads/memory") {
				t.Fatalf("no-parent human list: %s", human)
			}
			wantChildren := graphListSorted([]string{strings.TrimPrefix(child.ID, scope), "beads/parent.9"})
			paths, more = graphListBeads(t, bd, work, home, scope, "--parent", "beads/parent", "--sort=id")
			if !slices.Equal(paths, wantChildren) || more {
				t.Fatalf("parent records list = %v, more=%t; want %v", paths, more, wantChildren)
			}
			first, more := graphListBeads(t, bd, work, home, scope, "--parent", "beads/parent", "--sort=id", "--limit=1")
			if !slices.Equal(first, wantChildren[:1]) || !more {
				t.Fatalf("parent limited page = %v, more=%t; want %v and more", first, more, wantChildren[:1])
			}
			graphPolicyCLI(t, bd, work, home, []string{"BEADS_MAX_ROWS=2"}, "", "list", "--parent", "beads/parent", "--format=records-json")
			graphPolicyCLI(t, bd, work, home, []string{"BEADS_MAX_ROWS=1"}, "", "list", "--parent", "beads/parent", "--format=records-json", "--limit=1")
			graphPolicyCLI(t, bd, work, home, []string{"BEADS_MAX_ROWS=1"}, "capability_unavailable", "list", "--parent", "beads/parent", "--format=records-json")
			output = call("list", "--parent", parent.Properties.ID, "--json", "--sort=id")
			rows = nil
			if err := json.Unmarshal([]byte(output), &rows); err != nil {
				t.Fatalf("decode parent Issue list: %v: %s", err, output)
			}
			ids = ids[:0]
			for _, row := range rows {
				ids = append(ids, row.ID)
			}
			if !slices.Equal(ids, wantChildren) {
				t.Fatalf("bare native parent list = %v, want %v", ids, wantChildren)
			}
			if alias := call("list", "--filter-parent", "beads/parent", "--json", "--sort=id"); alias != call("list", "--parent", "beads/parent", "--json", "--sort=id") {
				t.Fatalf("filter-parent alias differs from parent:\n%s\n%s", alias, output)
			}
			human = call("list", "--parent", "beads/parent", "--flat", "--sort=id")
			for _, id := range wantChildren {
				if !strings.Contains(human, id) {
					t.Fatalf("parent flat list missing %s: %s", id, human)
				}
			}
			if strings.Contains(human, "Grandchild") || strings.Contains(human, "Memory") {
				t.Fatalf("parent flat list leaked other resources: %s", human)
			}
			long := call("list", "--parent", "beads/parent", "--flat", "--long")
			if !strings.Contains(long, "Child") || !strings.Contains(long, "Dotted without an edge") || strings.Contains(long, "Grandchild") {
				t.Fatalf("parent long list: %s", long)
			}
			tree := call("list", "--parent", "beads/parent", "--limit=1")
			if !strings.Contains(tree, "Parent") || !strings.Contains(tree, "Child") || !strings.Contains(tree, "Grandchild") || !strings.Contains(tree, "Dotted without an edge") || strings.Contains(tree, "Memory") || strings.Index(tree, "Child") > strings.Index(tree, "Grandchild") || !strings.Contains(tree, "Total: 4 issues") {
				t.Fatalf("unpaged parent tree lost the hierarchy: %s", tree)
			}
			if explicit := call("list", "--parent", "beads/parent", "--tree=true", "--limit=1"); explicit != tree {
				t.Fatalf("explicit parent tree differs from default:\n%s\n%s", explicit, tree)
			}
			graphListRefusal(t, bd, work, home, []string{"BEADS_MAX_ROWS=2"}, "list", "--parent", "beads/parent")
			call("create", "Edge-less orphan", "--id", "beads/orphan.3", "--json", "--force")
			paths, more = graphListBeads(t, bd, work, home, scope, "--parent", "beads/orphan")
			if !slices.Equal(paths, []string{"beads/orphan.3"}) || more {
				t.Fatalf("missing bare parent dotted fallback = %v, more=%t", paths, more)
			}
			paths, more = graphListBeads(t, bd, work, home, scope, "--parent", "beads/memory")
			if len(paths) != 0 || more {
				t.Fatalf("non-Issue parent unexpectedly selected Issues: %v, more=%t", paths, more)
			}
		})
	}
}

// Closed and pinned Issues are hidden from the Bead list unless --all, which
// also lifts the default row limit; an explicit --limit still wins. The CLI
// cannot author a pinned Issue, so the test seeds one between processes.
func TestGraphPreviewListHiddenIssuesWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			const scope = "https://example.invalid/hidden/"
			work, home := graphListWorkspace(t, bd, engine, scope)
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			list := func(extra ...string) ([]string, bool) {
				t.Helper()
				if !slices.Contains(extra, "--type") && !slices.Contains(extra, "--exact-type") {
					extra = append(extra, "--type", "any")
				}
				return graphListBeads(t, bd, work, home, scope, extra...)
			}
			human := func(extra ...string) string {
				t.Helper()
				if !slices.Contains(extra, "--type") && !slices.Contains(extra, "--exact-type") {
					extra = append(extra, "--type", "any")
				}
				return graphPolicyCLI(t, bd, work, home, nil, "", append([]string{"list"}, extra...)...)
			}
			// A piped process has no default row limit; a configured one is how the
			// default limit becomes visible to it.
			graphListConfig(t, work, "list:\n  limit: 2\n")
			call("remember", "private body", "--id", "beads/plan", "--title", "Plan")
			call("create", "Open work", "--id", "beads/open", "--force", "--priority", "1")
			call("create", "Finished work", "--id", "beads/done", "--force", "--priority", "2")
			call("close", scope+"beads/done")
			graphListSeedIssue(t, work, "beads/pin", publicops.CreateRequest{Actor: "test-author", Issue: &publicops.Issue{
				Title: "Pinned note", Description: "kept", Status: types.StatusPinned, IssueType: types.TypeTask, Priority: 3,
			}})
			formatted := graphPolicyCLI(t, bd, work, home, nil, "", "list", "--format=JSON", "--all")
			ordinary := graphPolicyCLI(t, bd, work, home, nil, "", "list", "--json", "--all")
			if formatted != ordinary {
				t.Fatalf("--format=JSON must retain the shipping --json alias:\nformat: %s\njson: %s", formatted, ordinary)
			}
			graphPolicyCLI(t, bd, work, home, nil, "", "list", "--format=json", "--include-comments", "--all")
			if treeFalse, flat := human("--tree=false", "--all"), human("--flat", "--all"); treeFalse != flat {
				t.Fatalf("--tree=false must retain the shipping flat-output alias:\ntree=false: %s\nflat: %s", treeFalse, flat)
			}
			listIDs := func(args ...string) []string {
				t.Helper()
				output := graphPolicyCLI(t, bd, work, home, nil, "", append([]string{"list", "--json"}, args...)...)
				var rows []struct {
					ID string `json:"id"`
				}
				if err := json.Unmarshal([]byte(output), &rows); err != nil {
					t.Fatalf("decode sorted Issue list: %v: %s", err, output)
				}
				ids := make([]string, 0, len(rows))
				for _, row := range rows {
					ids = append(ids, row.ID)
				}
				return ids
			}
			if got := listIDs("--all", "--sort=id", "--limit=2"); !slices.Equal(got, []string{"beads/done", "beads/open"}) {
				t.Fatalf("id sort must precede page trim: %v", got)
			}
			if got := listIDs("--long", "--all", "--sort=id", "--limit=2"); !slices.Equal(got, []string{"beads/done", "beads/open"}) {
				t.Fatalf("--json must keep its Issue array when --long is present: %v", got)
			}
			if got := listIDs("--all", "--sort=closed", "--limit=1"); !slices.Equal(got, []string{"beads/done"}) {
				t.Fatalf("closed sort must precede page trim: %v", got)
			}
			long := human("--long", "--all", "--sort=id", "--limit=2")
			if !strings.Contains(long, "Found 2 issues:") || !strings.Contains(long, "beads/done") || !strings.Contains(long, "beads/open") ||
				!strings.Contains(long, "More matching Issues exist") || strings.Contains(long, "beads/plan") || strings.Contains(long, "beads/pin") {
				t.Fatalf("--long must render the checked Issue page without Memories or off-page Issues: %s", long)
			}

			expect := func(what string, got []string, more bool, want []string, wantMore bool) {
				t.Helper()
				if !slices.Equal(graphListSorted(got), graphListSorted(want)) || more != wantMore {
					t.Fatalf("%s: listed %v more=%t, want %v more=%t", what, got, more, want, wantMore)
				}
			}
			everything := []string{"beads/done", "beads/open", "beads/pin", "beads/plan"}
			issues := []string{"beads/done", "beads/open", "beads/pin"}

			defaultIDs, defaultMore := graphListBeads(t, bd, work, home, scope)
			expect("default Issue family excludes Memories", defaultIDs, defaultMore, []string{"beads/open"}, false)
			got, more := list()
			expect("all-Type list hides the closed and the pinned Issue", got, more, []string{"beads/open", "beads/plan"}, false)
			got, more = list("--type", "types/issue")
			expect("Issue Type hides them too", got, more, []string{"beads/open"}, false)
			got, more = list("--type", "types/memory")
			expect("a Memory is never hidden", got, more, []string{"beads/plan"}, false)

			got, more = list("--all")
			expect("--all shows them and lifts the configured default limit", got, more, everything, false)
			got, more = list("--all", "--type", "types/issue")
			expect("--all with the Issue Type", got, more, issues, false)
			got, more = list("--all", "--limit", "3")
			if len(got) != 3 || !more {
				t.Fatalf("an explicit --limit must win over --all: listed %v more=%t", got, more)
			}
			got, more = list("--limit", "1")
			if len(got) != 1 || !more {
				t.Fatalf("a limit under the shown Beads must report more: listed %v more=%t", got, more)
			}

			// The human rows carry an Issue's status and priority, and the hint
			// offers --all only while it would still help.
			rows := human("--all")
			for _, want := range []string{"beads/done [task] closed P2 \"Finished work\"\n", "beads/open [task] open P1 \"Open work\"\n", "beads/pin [task] pinned P3 \"Pinned note\"\n", "beads/plan [types/memory] \"Plan\"\n"} {
				if !strings.Contains(rows, want) {
					t.Fatalf("human rows lack %q: %s", want, rows)
				}
			}
			if out := human("--all", "--limit", "3"); !strings.Contains(out, "More Beads exist; increase --limit within preview bounds.") || strings.Contains(out, "use --all") {
				t.Fatalf("hint under --all must not offer --all: %s", out)
			}
			if out := human("--limit", "1"); !strings.Contains(out, "More Beads exist; increase --limit or use --all within preview bounds.") {
				t.Fatalf("hint without --all must offer it: %s", out)
			}
			if out := human(); !strings.HasPrefix(out, "Beads (2; more: false; graph preview)\n") || strings.Contains(out, "beads/done") || strings.Contains(out, "beads/pin") || !strings.Contains(out, "beads/plan [types/memory] \"Plan\"\n") {
				t.Fatalf("default human list: %s", out)
			}
		})
	}
}
