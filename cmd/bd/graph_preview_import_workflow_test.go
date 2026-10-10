//go:build cgo

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// Installed process admission, input selection and persisted graph projections
// are distinct from the lower transaction fault-injection tests.
func TestGraphPreviewLegacyImportInstalledWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	fixture, err := os.ReadFile(filepath.Join("testdata", "legacy-import", "ordinary.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			isolateBeadsDirForTest(t)
			work, home := t.TempDir(), t.TempDir()
			const scope = "https://example.invalid/legacy-import/"
			call := func(input string, code string, args ...string) string {
				t.Helper()
				if args[0] != "list" {
					args = append(args, "--json")
				}
				return graphPatchProcess(t, bd, work, home, strings.NewReader(input), code, 90*time.Second, args...)
			}
			init := []string{"init", "--graph-mode", "link", "--scope-url", scope, "--skip-hooks", "--skip-agents", "--non-interactive"}
			if engine == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for required shared-server import proof")
				}
				init = append(init, "--server", "--external", "--server-host", "127.0.0.1", "--server-port", port, "--server-user", "root")
			}
			call("", "", init...)
			before := call("", "", "list", "--all", "--format", "records-json")
			for _, input := range []string{
				string(fixture) + "\n{broken}",
				"",
				`{"_schema":"beads-jsonl/1"}`,
				`{"id":"one-a","title":"a"}` + "\n" + `{"id":"two-b","title":"b"}`,
				strings.Replace(string(fixture), `"depends_on_id":"legacy-a"`, `"depends_on_id":"absent"`, 1),
				strings.Replace(string(fixture), `"type":"blocks"`, `"type":"parent-child"`, 1),
				string(fixture) + `{"_type":"issue","id":"bad","title":"bad","history":[]}`,
				`{"id":"old-a","title":"a","dependencies":[{"depends_on_id":"old-b","type":"blocks"}]}
{"id":"old-b","title":"b","dependencies":[{"depends_on_id":"old-a","type":"blocks"}]}`,
			} {
				call(input, "invalid_properties", "import", "-")
				if got := call("", "", "list", "--all", "--format", "records-json"); got != before {
					t.Fatal("refused import changed inventory")
				}
			}
			call(string(fixture), "capability_unavailable", "import", "-", "--allow-stale")
			call(string(fixture), "capability_unavailable", "import", "-", "--dedup")
			call(string(fixture), "capability_unavailable", "import", "-", "--global")
			call("", "invalid_selector", "import", "one", "two")
			call("", "invalid_properties", "import", "one", "--input", "two")
			call(strings.Repeat(" ", 16*1024*1024+1), "capability_unavailable", "import", "-")
			call(string(fixture), "permission_denied", "import", "-", "--readonly")
			call(string(fixture), "invalid_properties", "import")
			preview := graphMixedResult[graphstore.LegacyImportResult](t, call(string(fixture), "", "import", "-", "--dry-run"))
			if !preview.DryRun || preview.Issues != 2 || preview.Dependencies != 1 || preview.Comments != 1 || preview.Memories != 1 {
				t.Fatalf("preview=%+v", preview)
			}
			if call("", "", "list", "--all", "--format", "records-json") != before {
				t.Fatal("dry run reserved data")
			}
			file := filepath.Join(work, "ordinary.jsonl")
			if err := os.WriteFile(file, fixture, 0600); err != nil {
				t.Fatal(err)
			}
			result := graphMixedResult[graphstore.LegacyImportResult](t, call("", "", "import", "--input", file))
			if result.DryRun || result.IDs["legacy-b"] != scope+"beads/legacy-b" || result.MemoryIDs["fixture-key"] != scope+"beads/fixture-key" {
				t.Fatalf("result=%+v", result)
			}
			issue := graphMixedResult[graphstore.IssueRecord](t, call("", "", "show", "legacy-b"))
			if len(issue.Owned) != 1 || issue.Properties.Notes != "A note" || issue.Properties.ID != "legacy-b" {
				t.Fatalf("issue=%+v", issue)
			}
			memory := graphMixedResult[graphstore.Record](t, call("", "", "show", "fixture-key"))
			if memory.Properties.Title != "fixture-key" || memory.Properties.Body != "Remembered context" {
				t.Fatalf("memory=%+v", memory)
			}
			comments := call("", "", "comments", "legacy-b")
			if !strings.Contains(comments, "fixture-commenter") || !strings.Contains(comments, "A preserved comment") {
				t.Fatalf("comments=%s", comments)
			}
			retained := call("", "", "show", issue.ID, "--version", issue.Revision)
			if !strings.Contains(retained, "A note") {
				t.Fatal("missing retained initial Issue")
			}
			inventory := call("", "", "list", "--all", "--format", "records-json")
			versions := call("", "", "versions", issue.ID)
			for _, args := range [][]string{{"import", file}, {"import", file, "--dry-run"}} {
				call("", "invalid_properties", args...)
				if call("", "", "list", "--all", "--format", "records-json") != inventory || call("", "", "versions", issue.ID) != versions {
					t.Fatal("repeat import changed data or History")
				}
			}
			fresh := graphMixedResult[graphstore.IssueRecord](t, call("", "", "create", "After import"))
			if !strings.HasPrefix(fresh.Properties.ID, "legacy-") {
				t.Fatalf("new Issue prefix: %s", fresh.Properties.ID)
			}
			call("", "", "dep", "add", fresh.ID, "legacy-a")
			call("", "", "dep", "add", "legacy-b", fresh.ID)
		})
	}
}
