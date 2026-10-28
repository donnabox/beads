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
			issue := graphMixedResult[graphstore.IssueRecord](t, call("create", "Work", "--id", "beads/work", "--force"))
			preview := graphMixedResult[graphstore.IssueDeleteResult](t, call("delete", issue.Properties.ID))
			if !preview.Preview || preview.Deleted || !reflect.DeepEqual(preview.Issue, issue) {
				t.Fatal("default preview did not disclose unchanged Issue")
			}
			unguarded := graphMixedResult[graphstore.IssueRecord](t, call("create", "Unguarded deletion", "--id", "beads/unguarded", "--force"))
			applied := graphMixedResult[graphstore.IssueDeleteResult](t, call("delete", unguarded.Properties.ID, "--force"))
			if !applied.Deleted || applied.Preview || !reflect.DeepEqual(applied.Issue, unguarded) {
				t.Fatal("unguarded --force did not delete the current Issue")
			}
			refuse("gone", "show", unguarded.ID)
			refuse("revision_conflict", "delete", issue.Properties.ID, "--force", "--if-revision", "=stale")
			refuse("capability_unavailable", "forget", issue.Properties.ID)
			call("remember", "Other", "--id", "beads/other")
			linked := graphMixedResult[graphstore.IssueRecord](t, call("create", "Linked", "--id", "beads/linked", "--force"))
			call("link", linked.ID, scope+"beads/other", "--type", "references", "--id", "links/relation")
			linkedPreview := graphMixedResult[graphstore.IssueDeleteResult](t, call("delete", linked.Properties.ID))
			if !linkedPreview.Preview || linkedPreview.Deleted || !reflect.DeepEqual(linkedPreview.Issue, linked) || len(linkedPreview.IncidentLinkIDs) != 1 {
				t.Fatal("informational Link preview did not disclose its cleanup")
			}
			refuse("revision_conflict", "delete", linked.Properties.ID, "--force", "--if-revision", "=stale")
			linkedDelete := graphMixedResult[graphstore.IssueDeleteResult](t, call("delete", linked.Properties.ID, "--force", "--if-revision", "="+linked.Revision))
			if !linkedDelete.Deleted || len(linkedDelete.IncidentLinkIDs) != 1 {
				t.Fatal("linked Issue deletion omitted Link cleanup")
			}
			refuse("gone", "show", "links/relation")
			call("show", "beads/other")

			blocker := graphMixedResult[graphstore.IssueRecord](t, call("create", "Blocker", "--id", "beads/blocker", "--force"))
			call("create", "Dependent", "--id", "beads/dependent", "--force")
			dependency := graphMixedResult[graphstore.DependencyResult](t, call("dep", "add", "beads/dependent", blocker.Properties.ID))
			refuse("constraint_violation", "delete", blocker.Properties.ID)
			if got := graphMixedResult[graphstore.IssueRecord](t, call("show", "--format", "graph-json", blocker.Properties.ID)); !reflect.DeepEqual(got, blocker) {
				t.Fatal("dependent refusal changed blocker")
			}
			forced := graphMixedResult[graphstore.IssueDeleteResult](t, call("delete", blocker.Properties.ID, "--force", "--if-revision", "="+blocker.Revision))
			if !forced.Deleted || len(forced.IncidentLinkIDs) != 1 || forced.IncidentLinkIDs[0] != dependency.Link.ID {
				t.Fatal("force deletion omitted orphan Link cleanup")
			}
			call("show", "beads/dependent")
			refuse("gone", "show", dependency.Link.ID)

			final := graphMixedResult[graphstore.IssueRecord](t, call("show", "--format", "graph-json", issue.ID))
			before := call("versions", issue.ID)
			deleted := graphMixedResult[graphstore.IssueDeleteResult](t, call("delete", issue.Properties.ID, "--force", "--if-revision", "="+final.Revision))
			if !deleted.Deleted || deleted.Preview || !reflect.DeepEqual(deleted.Issue, final) {
				t.Fatal("apply did not disclose final live Issue")
			}
			refuse("gone", "show", issue.ID)
			refuse("gone", "delete", issue.Properties.ID, "--force")
			refuse("identity_reserved", "create", "Replacement", "--id", "beads/work", "--force")
			type terminalHistory struct {
				CurrentState *graphstore.BeadTerminalState `json:"currentState"`
				Versions     []map[string]any              `json:"versions"`
				Terminal     []map[string]any              `json:"terminal_versions"`
			}
			prior := graphMixedResult[terminalHistory](t, before)
			after := graphMixedResult[terminalHistory](t, call("versions", issue.ID))
			if deleted.Terminal == nil || !reflect.DeepEqual(after.CurrentState, deleted.Terminal) || !reflect.DeepEqual(after.Versions, prior.Versions) || len(after.Terminal) != 1 {
				t.Fatal("native deletion lost prior snapshots or distinct D")
			}

			if old := graphMixedResult[graphstore.IssueRecord](t, call("show", issue.ID, "--revision", "="+final.Revision)); old.ID != final.ID || old.Revision != final.Revision {
				t.Fatal("final live snapshot unavailable")
			}
			call("show", "beads/other")
		})
	}
}
