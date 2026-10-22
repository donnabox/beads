//go:build cgo

package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func TestGraphPreviewIssueDeleteSetWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			const scope = "https://example.invalid/issue-delete-set/"
			initArgs := []string{"init", "--graph-mode", "link", "--scope-url", scope, "--skip-hooks", "--skip-agents", "--non-interactive"}
			if engine == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for shared-server CLI qualification")
				}
				initArgs = append(initArgs, "--server", "--external", "--server-host", "127.0.0.1", "--server-port", port, "--server-user", "root")
			}
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			refuse := func(code string, args ...string) {
				t.Helper()
				graphPolicyCLI(t, bd, work, home, nil, code, append(args, "--json")...)
			}
			call(initArgs...)
			call("create", "Blocker", "--id", "beads/blocker")
			call("create", "Dependent", "--id", "beads/dependent")
			call("dep", "add", "dependent", "blocker")
			blockerBefore, dependentBefore := call("versions", "blocker"), call("versions", "dependent")
			refuse("not_found", "delete", "blocker", "dependent", "missing", "--force")
			if call("versions", "blocker") != blockerBefore || call("versions", "dependent") != dependentBefore {
				t.Fatal("missing named selector changed an earlier Issue")
			}
			preview := graphMixedResult[graphstore.IssueDeleteSetResult](t, call("delete", "blocker", "dependent"))
			if !preview.Preview || preview.Deleted || len(preview.Issues) != 2 || len(preview.IncidentLinkIDs) != 1 {
				t.Fatalf("named pair preview: %+v", preview)
			}
			if call("versions", "blocker") != blockerBefore || call("versions", "dependent") != dependentBefore {
				t.Fatal("named pair preview wrote state")
			}
			refuse("invalid_selector", "delete", "blocker", "dependent", "--if-revision", "stale", "--force")
			applied := graphMixedResult[graphstore.IssueDeleteSetResult](t, call("delete", "blocker", "dependent", "--force"))
			if !applied.Deleted || len(applied.Issues) != 2 || len(applied.IncidentLinkIDs) != 1 {
				t.Fatalf("named pair deletion: %+v", applied)
			}
			refuse("gone", "show", "blocker")
			refuse("gone", "show", "dependent")

			call("create", "Cascade blocker", "--id", "beads/cascade-blocker")
			call("create", "Cascade dependent", "--id", "beads/cascade-dependent")
			call("dep", "add", "cascade-dependent", "cascade-blocker")
			cascadePreview := graphMixedResult[graphstore.IssueDeleteSetResult](t, call("delete", "cascade-blocker", "--cascade"))
			if !cascadePreview.Preview || len(cascadePreview.Issues) != 2 || len(cascadePreview.IncidentLinkIDs) != 1 {
				t.Fatalf("cascade preview: %+v", cascadePreview)
			}
			cascade := graphMixedResult[graphstore.IssueDeleteSetResult](t, call("delete", "cascade-blocker", "--cascade", "--force"))
			if !cascade.Deleted || len(cascade.Issues) != 2 {
				t.Fatalf("cascade deletion: %+v", cascade)
			}
			refuse("gone", "show", "cascade-blocker")
			refuse("gone", "show", "cascade-dependent")

			call("create", "File first", "--id", "beads/file-first")
			call("create", "File second", "--id", "beads/file-second")
			file := filepath.Join(work, "delete-ids.txt")
			if err := os.WriteFile(file, []byte("# native Issue IDs\nfile-first\n\nfile-second\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			fromFile := graphMixedResult[graphstore.IssueDeleteSetResult](t, call("delete", "--from-file", file, "--force"))
			if !fromFile.Deleted || len(fromFile.Issues) != 2 {
				t.Fatalf("file-selected deletion: %+v", fromFile)
			}
			refuse("gone", "show", "file-first")
			refuse("gone", "show", "file-second")
		})
	}
}
