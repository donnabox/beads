//go:build cgo

package main

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
)

func TestGraphPreviewOpaqueIssueIDShowUpdateCloseRoundtrip(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, order := range [][]string{{"x", "beads/x"}, {"beads/x", "x"}} {
		for _, engine := range []string{"embedded", "server"} {
			t.Run(order[0]+"-first/"+engine, func(t *testing.T) {
				work, home := t.TempDir(), t.TempDir()
				const scope = "https://example.invalid/opaque-roundtrip/"
				init := []string{"init", "--graph-mode", "link", "--scope-url", scope, "--skip-hooks", "--skip-agents", "--non-interactive"}
				if engine == "server" {
					port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
					if port == "" {
						t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for shared-server qualification")
					}
					init = append(init, "--server", "--external", "--server-host", "127.0.0.1", "--server-port", port, "--server-user", "root")
				}
				call := func(args ...string) string {
					t.Helper()
					return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
				}
				call(init...)
				created := map[string]graphstore.IssueRecord{}
				for _, id := range order {
					created[id] = graphMixedResult[graphstore.IssueRecord](t, call("create", id, "--id", id, "--force"))
					if created[id].Properties.ID != id {
						t.Fatalf("--id %q changed the native public ID: %+v", id, created[id])
					}
				}
				if created[order[0]].ID == created[order[1]].ID {
					t.Fatal("distinct public IDs shared one graph address")
				}
				for _, id := range order {
					shown := graphMixedResult[graphstore.IssueRecord](t, call("show", id, "--format", "graph-json"))
					if shown.ID != created[id].ID || shown.Properties.ID != id {
						t.Fatalf("show %q changed resource: %+v", id, shown)
					}
					var details []types.IssueDetails
					if err := json.Unmarshal([]byte(call("show", id)), &details); err != nil {
						t.Fatal(err)
					}
					if len(details) != 1 || details[0].ID != id {
						t.Fatalf("familiar show %q did not emit its public ID: %+v", id, details)
					}
				}
				updated := graphMixedResult[graphstore.IssueMutationResult](t, call("update", "beads/x", "--title", "Only literal slash ID"))
				if updated.Issue.Properties.ID != "beads/x" || updated.Issue.Properties.Title != "Only literal slash ID" {
					t.Fatalf("update targeted the wrong Issue: %+v", updated)
				}
				if shown := graphMixedResult[graphstore.IssueRecord](t, call("show", "x", "--format", "graph-json")); shown.Properties.Title != "x" {
					t.Fatalf("update of beads/x changed x: %+v", shown)
				}
				call("create", "Target", "--id", "target", "--force")
				dependency := graphMixedResult[graphstore.DependencyResult](t, call("dep", "add", "beads/x", "target"))
				if !dependency.Changed || dependency.Source.Properties.ID != "beads/x" || dependency.Link.Source != created["beads/x"].ID {
					t.Fatalf("dependency selected path owner instead of literal public ID: %+v", dependency)
				}
				if shown := graphMixedResult[graphstore.IssueRecord](t, call("show", "x", "--format", "graph-json")); len(shown.Owned) != 0 {
					t.Fatalf("dependency changed other public Issue ID: %+v", shown)
				}
				closed := graphMixedResult[graphstore.IssueMutationResult](t, call("close", "x", "--reason", "done"))
				if closed.Issue.Properties.ID != "x" || closed.Issue.Properties.Status != types.StatusClosed {
					t.Fatalf("close targeted the wrong Issue: %+v", closed)
				}
				if shown := graphMixedResult[graphstore.IssueRecord](t, call("show", "beads/x", "--format", "graph-json")); shown.Properties.Status != types.StatusOpen {
					t.Fatalf("close of x changed beads/x: %+v", shown)
				}
			})
		}
	}
}
