//go:build cgo

package main

import (
	"os"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// A graph path need not equal an Issue's native ID. The ordinary blocked
// --parent filter must use the native ID after resolving a graph selector.
func TestGraphPreviewBlockedParentWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			const scope = "https://example.invalid/blocked-parent/"
			args := []string{"init", "--graph-mode", "link", "--scope-url", scope, "--prefix", "bp", "--skip-hooks", "--skip-agents", "--non-interactive"}
			if engine == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for shared-server qualification")
				}
				args = append(args, "--server", "--external", "--server-host", "127.0.0.1", "--server-port", port, "--server-user", "root")
			}
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			call(args...)
			parent := graphMixedResult[graphstore.IssueRecord](t, call("create", "Parent", "--graph-path", "beads/selected-parent", "--type", "epic"))
			child := graphMixedResult[graphstore.IssueRecord](t, call("create", "Child", "--parent", parent.ID))
			blocker := graphMixedResult[graphstore.IssueRecord](t, call("create", "Prerequisite"))
			unrelated := graphMixedResult[graphstore.IssueRecord](t, call("create", "Unrelated"))
			call("dep", "add", child.ID, blocker.ID)
			call("dep", "add", unrelated.ID, blocker.ID)
			for _, selector := range []string{parent.Properties.ID, parent.ID} {
				got := graphMixedResult[[]graphstore.BlockedIssue](t, call("blocked", "--parent", selector))
				if len(got) != 1 || got[0].Issue.ID != child.ID {
					t.Fatalf("blocked --parent %q selected %+v", selector, got)
				}
			}
			if got := graphMixedResult[[]graphstore.BlockedIssue](t, call("blocked", "--parent", scope+"beads/missing")); len(got) != 0 {
				t.Fatalf("missing explicit parent selected %+v", got)
			}
			call("remember", "Distinct public ID", "--id", "beads/"+parent.Properties.ID)
			if got := graphMixedResult[[]graphstore.BlockedIssue](t, call("blocked", "--parent", parent.Properties.ID)); len(got) != 1 || got[0].Issue.ID != child.ID {
				t.Fatalf("Memory path shorthand displaced the native parent ID: %+v", got)
			}
			if got := graphMixedResult[[]graphstore.BlockedIssue](t, call("blocked", "--parent", parent.ID)); len(got) != 1 || got[0].Issue.ID != child.ID {
				t.Fatalf("explicit parent stopped resolving after bare-ID collision: %+v", got)
			}
		})
	}
}
