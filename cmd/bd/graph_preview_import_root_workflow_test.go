//go:build cgo

package main

import (
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func TestGraphPreviewLegacyImportRootInstalledWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	const fixture = `{"id":"old-a","title":"Target"}
{"id":"old-b","title":"Source","dependencies":[{"depends_on_id":"old-a","type":"relates-to"}],"comments":[{"id":"original-comment","author":"original","text":"rooted comment","created_at":"2026-10-20T01:02:03Z"}]}
{"_type":"memory","key":"beads/note","value":""}`
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			isolateBeadsDirForTest(t)
			work, home := t.TempDir(), t.TempDir()
			const scope = "https://example.invalid/rooted-import/"
			call := func(input, code string, args ...string) string {
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
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for required rooted import proof")
				}
				init = append(init, "--server", "--external", "--server-host", "127.0.0.1", "--server-port", port, "--server-user", "root")
			}
			call("", "", init...)
			before := call("", "", "list", "--all", "--type", "any", "--format", "records-json")
			for _, root := range []string{"", "..", "external:x"} {
				call(fixture, "invalid_properties", "import", "-", "--id-root", root, "--dry-run")
				if got := call("", "", "list", "--all", "--type", "any", "--format", "records-json"); got != before {
					t.Fatal("root refusal changed inventory")
				}
			}
			preview := graphMixedResult[graphstore.LegacyImportResult](t, call(fixture, "", "import", "-", "--id-root", "foo/bar", "--dry-run"))
			if !preview.DryRun || !preview.GeneratedIDsProvisional || preview.SourceIssuePrefix != "old" || preview.IDRoot != "foo/bar" || preview.IssuePrefix == "" {
				t.Fatalf("rooted preview=%+v", preview)
			}
			if got := call("", "", "list", "--all", "--type", "any", "--format", "records-json"); got != before {
				t.Fatal("rooted dry-run reserved data")
			}
			human := graphPatchProcess(t, bd, work, home, strings.NewReader(fixture), "", 90*time.Second, "import", "-", "--id-root", "foo/bar", "--dry-run")
			if !strings.Contains(human, "ID root: foo/bar") || !strings.Contains(human, "provisional") {
				t.Fatalf("human dry-run hides root/provisional IDs: %s", human)
			}
			result := graphMixedResult[graphstore.LegacyImportResult](t, call(fixture, "", "import", "-", "--id-root", "foo/bar"))
			if result.DryRun || result.GeneratedIDsProvisional || result.IssuePrefix != preview.IssuePrefix || !reflect.DeepEqual(result.IDs, preview.IDs) || !reflect.DeepEqual(result.MemoryIDs, preview.MemoryIDs) {
				t.Fatalf("rooted apply=%+v", result)
			}
			if result.IDs["old-a"] != scope+"foo/bar/old-a" || result.MemoryIDs["beads/note"] != scope+"foo/bar/beads/note" {
				t.Fatalf("literal/root mapping changed: %+v", result)
			}
			memory := graphMixedResult[graphstore.Record](t, call("", "", "show", result.MemoryIDs["beads/note"]))
			if memory.Properties.Body != "" || memory.Properties.Title != "Untitled memory" {
				t.Fatalf("empty legacy Memory changed: %+v", memory)
			}
			issue := graphMixedResult[graphstore.IssueRecord](t, call("", "", "show", result.IDs["old-b"], "--format", "graph-json"))
			retained := graphMixedResult[graphstore.IssueRecord](t, call("", "", "show", issue.ID, "--revision", issue.Revision))
			if issue.Properties.ID != "foo/bar/old-b" || !reflect.DeepEqual(issue.Properties, retained.Properties) || !reflect.DeepEqual(issue.Owned, retained.Owned) {
				t.Fatalf("rooted retained Issue mismatch: %+v %+v", issue, retained)
			}
			link := graphMixedResult[graphstore.LinkRecord](t, call("", "", "link", "show", result.DependencyIDs["old-b -> old-a"]))
			if !strings.HasPrefix(link.ID, scope+"foo/bar/") || strings.Contains(strings.TrimPrefix(link.ID, scope+"foo/bar/"), "/") || link.Source != result.IDs["old-b"] || link.Target != result.IDs["old-a"] || link.Properties["dep_type"] != "relates-to" {
				t.Fatalf("rooted Link/endpoints=%+v", link)
			}
			comments := call("", "", "comments", issue.ID)
			if !strings.Contains(comments, `"issue_id":"foo/bar/old-b"`) && !strings.Contains(comments, `"issue_id": "foo/bar/old-b"`) {
				t.Fatalf("rooted comment ownership missing: %s", comments)
			}
			fresh := graphMixedResult[graphstore.IssueRecord](t, call("", "", "create", "After rooted import"))
			if !strings.HasPrefix(fresh.Properties.ID, result.IssuePrefix+"-") {
				t.Fatalf("root changed future create prefix: %+v", fresh)
			}
			call("", "", "dep", "add", fresh.ID, result.IDs["old-a"])
			call("", "", "dep", "add", result.IDs["old-b"], fresh.ID)
		})
	}
}
