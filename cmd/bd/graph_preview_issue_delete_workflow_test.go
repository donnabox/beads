//go:build cgo

package main

import (
	"os"
	"reflect"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func TestGraphPreviewIssueDeleteWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			const scope = "https://example.invalid/issue-delete/"
			initArgs := []string{"init", "--graph-mode", "link", "--scope-url", scope, "--skip-hooks", "--skip-agents", "--non-interactive"}
			if engine == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for ordinary shared-server CLI qualification")
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
			issue := graphMixedResult[graphstore.IssueRecord](t, call("create", "Work", "--graph-path", "beads/work"))
			preview := graphMixedResult[graphstore.IssueDeleteResult](t, call("delete", "work"))
			if !preview.Preview || preview.Deleted || !reflect.DeepEqual(preview.Issue, issue) {
				t.Fatal("default preview did not disclose unchanged Issue")
			}
			unguarded := graphMixedResult[graphstore.IssueRecord](t, call("create", "Unguarded deletion", "--graph-path", "beads/unguarded"))
			applied := graphMixedResult[graphstore.IssueDeleteResult](t, call("delete", "unguarded", "--force"))
			if !applied.Deleted || applied.Preview || !reflect.DeepEqual(applied.Issue, unguarded) {
				t.Fatal("unguarded --force did not delete the current Issue")
			}
			refuse("gone", "show", unguarded.ID)
			refuse("revision_conflict", "delete", "work", "--force", "--if-revision", "stale")
			refuse("capability_unavailable", "forget", "work")
			call("remember", "Other", "--id", "beads/other")
			call("link", "work", "other", "--type", "types/references", "--id", "links/relation")
			refuse("constraint_violation", "delete", "work")
			refuse("constraint_violation", "delete", "work", "--force", "--if-revision", issue.Revision)
			if got := graphMixedResult[graphstore.IssueRecord](t, call("show", "--format", "graph-json", issue.ID)); !reflect.DeepEqual(got, issue) {
				t.Fatal("incident Link refusal changed Issue")
			}
			link := graphMixedResult[graphstore.LinkRecord](t, call("show", scope+"links/relation"))
			call("link", "delete", "--force", "links/relation", "--if-revision", link.Revision)
			final := graphMixedResult[graphstore.IssueRecord](t, call("show", "--format", "graph-json", issue.ID))
			before := call("versions", "work")
			deleted := graphMixedResult[graphstore.IssueDeleteResult](t, call("delete", "work", "--force", "--if-revision", final.Revision))
			if !deleted.Deleted || deleted.Preview || !reflect.DeepEqual(deleted.Issue, final) {
				t.Fatal("apply did not disclose final live Issue")
			}
			refuse("gone", "show", issue.ID)
			refuse("gone", "delete", "work", "--force")
			refuse("identity_reserved", "create", "Replacement", "--graph-path", "beads/work")
			type terminalHistory struct {
				CurrentState *graphstore.BeadTerminalState `json:"currentState"`
				Versions     []map[string]any              `json:"versions"`
				Terminal     []map[string]any              `json:"terminal_versions"`
			}
			prior := graphMixedResult[terminalHistory](t, before)
			after := graphMixedResult[terminalHistory](t, call("versions", "work"))
			if deleted.Terminal == nil || !reflect.DeepEqual(after.CurrentState, deleted.Terminal) || !reflect.DeepEqual(after.Versions, prior.Versions) || len(after.Terminal) != 1 {
				t.Fatal("native deletion lost prior snapshots or distinct D")
			}

			if old := graphMixedResult[graphstore.IssueRecord](t, call("show", issue.ID, "--version", final.Revision)); old.ID != final.ID || old.Revision != final.Revision {
				t.Fatal("final live snapshot unavailable")
			}
			call("show", "other")
		})
	}
}
