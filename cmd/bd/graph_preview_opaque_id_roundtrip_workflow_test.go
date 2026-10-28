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
				var removed struct {
					Removed bool   `json:"removed"`
					IssueID string `json:"issue_id"`
				}
				if err := json.Unmarshal([]byte(call("dep", "remove", "beads/x", "target")), &removed); err != nil || !removed.Removed || removed.IssueID != "beads/x" {
					t.Fatalf("dependency removal selected path owner instead of literal public ID: %+v, %v", removed, err)
				}
				if shown := graphMixedResult[graphstore.IssueRecord](t, call("show", "x", "--format", "graph-json")); len(shown.Owned) != 0 {
					t.Fatalf("dependency removal changed other public Issue ID: %+v", shown)
				}
				call("dep", "add", "beads/x", "target")
				memory := graphMixedResult[graphstore.Record](t, call("remember", "Context", "--id", "beads/context"))
				beforeMemory := call("show", "beads/context", "--format", "graph-json")
				graphPolicyCLI(t, bd, work, home, nil, "not_found", "update", "context", "--properties", `{"title":"Must refuse"}`, "--json")
				if afterMemory := call("show", "beads/context", "--format", "graph-json"); afterMemory != beforeMemory {
					t.Fatalf("shorthand write changed Memory despite pre-transaction selector refusal: %s", afterMemory)
				}
				call("update", "beads/context", "--properties", `{"title":"Exact public ID"}`)
				info := graphMixedResult[graphstore.LinkMutationResult](t, call("link", "beads/x", "beads/context", "--type", "references"))
				if info.Link.Source != created["beads/x"].ID || info.Link.Source == created["x"].ID {
					t.Fatalf("informational Link selected path owner instead of literal public ID: %+v", info.Link)
				}
				unlinked := graphMixedResult[graphstore.LinkDeleteResult](t, call("link", "delete", info.Link.ID, "--force"))
				if !unlinked.Changed || unlinked.Link.ID != info.Link.ID {
					t.Fatalf("pair unlink selected the wrong public-ID endpoints: %+v", unlinked)
				}
				incoming := graphMixedResult[graphstore.LinkMutationResult](t, call("link", "beads/context", "beads/x", "--type", "references"))
				if incoming.Link.Source != memory.ID || incoming.Link.Target != created["beads/x"].ID {
					t.Fatalf("informational target selected path owner instead of literal public ID: %+v", incoming.Link)
				}
				unlinked = graphMixedResult[graphstore.LinkDeleteResult](t, call("link", "delete", incoming.Link.ID, "--force"))
				if !unlinked.Changed || unlinked.Link.ID != incoming.Link.ID {
					t.Fatalf("pair unlink selected the wrong public-ID target: %+v", unlinked)
				}
				closed := graphMixedResult[graphstore.IssueMutationResult](t, call("close", "x", "--reason", "done"))
				if closed.Issue.Properties.ID != "x" || closed.Issue.Properties.Status != types.StatusClosed {
					t.Fatalf("close targeted the wrong Issue: %+v", closed)
				}
				if shown := graphMixedResult[graphstore.IssueRecord](t, call("show", "beads/x", "--format", "graph-json")); shown.Properties.Status != types.StatusOpen {
					t.Fatalf("close of x changed beads/x: %+v", shown)
				}
				preview := graphMixedResult[graphstore.IssueDeleteSetResult](t, call("delete", "beads/x", "target"))
				if !preview.Preview || len(preview.Issues) != 2 || len(preview.IncidentLinkIDs) != 1 {
					t.Fatalf("opaque set preview selected the wrong Issues or Links: %+v", preview)
				}
				deleted := graphMixedResult[graphstore.IssueDeleteSetResult](t, call("delete", "beads/x", "target", "--force"))
				if !deleted.Deleted || len(deleted.Issues) != 2 || len(deleted.IncidentLinkIDs) != 1 {
					t.Fatalf("opaque set deletion selected the wrong Issues or Links: %+v", deleted)
				}
				if shown := graphMixedResult[graphstore.IssueRecord](t, call("show", "x", "--format", "graph-json")); shown.Properties.Status != types.StatusClosed {
					t.Fatalf("deleting beads/x also deleted x: %+v", shown)
				}
			})
		}
	}
}
