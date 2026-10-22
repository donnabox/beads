//go:build cgo

package main

import (
	"os"
	"reflect"
	"strings"
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
			issue := graphMixedResult[graphstore.IssueRecord](t, call("create", "Work", "--id", "beads/work"))
			prefix, _, ok := strings.Cut(issue.Properties.ID, "-")
			if !ok {
				t.Fatalf("fixture needs a configured Issue prefix: %q", issue.Properties.ID)
			}
			abbreviable := graphMixedResult[graphstore.IssueRecord](t, call("create", "Abbreviable", "--id", prefix+"-a3f8e9"))
			abbreviatedPreview := graphMixedResult[graphstore.IssueDeleteResult](t, call("delete", "a3f8"))
			if !abbreviatedPreview.Preview || abbreviatedPreview.Issue.ID != abbreviable.ID {
				t.Fatalf("abbreviated Issue preview: %+v", abbreviatedPreview)
			}
			refuse("revision_conflict", "delete", "a3f8", "--force", "--if-revision", "stale")
			abbreviatedDelete := graphMixedResult[graphstore.IssueDeleteResult](t, call("delete", "a3f8", "--force"))
			if !abbreviatedDelete.Deleted || abbreviatedDelete.Issue.ID != abbreviable.ID {
				t.Fatalf("abbreviated Issue delete: %+v", abbreviatedDelete)
			}
			refuse("gone", "show", abbreviable.Properties.ID)
			preview := graphMixedResult[graphstore.IssueDeleteResult](t, call("delete", "work"))
			if !preview.Preview || preview.Deleted || !reflect.DeepEqual(preview.Issue, issue) {
				t.Fatal("default preview did not disclose unchanged Issue")
			}
			unguarded := graphMixedResult[graphstore.IssueRecord](t, call("create", "Unguarded deletion", "--id", "beads/unguarded"))
			applied := graphMixedResult[graphstore.IssueDeleteResult](t, call("delete", "unguarded", "--force"))
			if !applied.Deleted || applied.Preview || !reflect.DeepEqual(applied.Issue, unguarded) {
				t.Fatal("unguarded --force did not delete the current Issue")
			}
			refuse("gone", "show", "unguarded")
			refuse("revision_conflict", "delete", "work", "--force", "--if-revision", "stale")
			refuse("capability_unavailable", "forget", "work")
			call("remember", "Other", "--id", "beads/other")
			if memory := graphMixedResult[graphstore.MemoryDeleteResult](t, call("delete", "other")); !memory.Preview || memory.Deleted {
				t.Fatalf("bare Memory delete preview stopped dispatching: %+v", memory)
			}
			linked := graphMixedResult[graphstore.IssueRecord](t, call("create", "Linked", "--id", "beads/linked"))
			call("link", "linked", "other", "--type", "types/references", "--id", "links/relation")
			linkedPreview := graphMixedResult[graphstore.IssueDeleteResult](t, call("delete", "linked"))
			if !linkedPreview.Preview || linkedPreview.Deleted || !reflect.DeepEqual(linkedPreview.Issue, linked) || len(linkedPreview.IncidentLinkIDs) != 1 {
				t.Fatal("informational Link preview did not disclose its cleanup")
			}
			refuse("revision_conflict", "delete", "linked", "--force", "--if-revision", "stale")
			linkedDelete := graphMixedResult[graphstore.IssueDeleteResult](t, call("delete", "linked", "--force", "--if-revision", linked.Revision))
			if !linkedDelete.Deleted || len(linkedDelete.IncidentLinkIDs) != 1 {
				t.Fatal("linked Issue deletion omitted Link cleanup")
			}
			refuse("gone", "show", "links/relation")
			call("show", "other")

			blocker := graphMixedResult[graphstore.IssueRecord](t, call("create", "Blocker", "--id", "beads/blocker"))
			call("create", "Dependent", "--id", "beads/dependent")
			dependency := graphMixedResult[graphstore.DependencyResult](t, call("dep", "add", "dependent", "blocker"))
			refuse("constraint_violation", "delete", "blocker")
			if got := graphMixedResult[graphstore.IssueRecord](t, call("show", "--format", "graph-json", "blocker")); !reflect.DeepEqual(got, blocker) {
				t.Fatal("dependent refusal changed blocker")
			}
			forced := graphMixedResult[graphstore.IssueDeleteResult](t, call("delete", "blocker", "--force", "--if-revision", blocker.Revision))
			if !forced.Deleted || len(forced.IncidentLinkIDs) != 1 || forced.IncidentLinkIDs[0] != dependency.Link.ID {
				t.Fatal("force deletion omitted orphan Link cleanup")
			}
			call("show", "dependent")
			refuse("gone", "show", dependency.Link.ID)

			final := graphMixedResult[graphstore.IssueRecord](t, call("show", "--format", "graph-json", "work"))
			before := call("versions", "work")
			deleted := graphMixedResult[graphstore.IssueDeleteResult](t, call("delete", "work", "--force", "--if-revision", final.Revision))
			if !deleted.Deleted || deleted.Preview || !reflect.DeepEqual(deleted.Issue, final) {
				t.Fatal("apply did not disclose final live Issue")
			}
			refuse("gone", "show", "work")
			refuse("gone", "delete", "work", "--force")
			refuse("identity_reserved", "create", "Replacement", "--id", "beads/work")
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

			if old := graphMixedResult[graphstore.IssueRecord](t, call("show", "work", "--version", final.Revision)); old.ID != final.ID || old.Revision != final.Revision {
				t.Fatal("final live snapshot unavailable")
			}
			call("show", "other")
		})
	}
}
