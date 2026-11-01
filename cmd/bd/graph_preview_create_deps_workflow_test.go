//go:build cgo

package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
)

func TestGraphPreviewCreateLocalDependenciesWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			initArgs := []string{"init", "--graph-mode", "link", "--prefix", "dep", "--scope-url", "https://example.invalid/create-deps/", "--skip-hooks", "--skip-agents", "--non-interactive", "--json"}
			if engine == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for ordinary shared-server qualification")
				}
				initArgs = append(initArgs, "--server", "--external", "--server-host", "127.0.0.1", "--server-port", port, "--server-user", "root")
			}
			graphPolicyCLI(t, bd, work, home, nil, "", initArgs...)
			prereq := graphMixedResult[graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "create", "Prerequisite", "--id", "dep-prereq", "--json"))
			consumer := graphMixedResult[graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "create", "Consumer", "--id", "dep-consumer", "--json"))
			created := graphMixedResult[graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "create", "New", "--id", "dep-new", "--deps", "blocked-by:dep-prereq,blocks:dep-consumer", "--json"))
			consumer = graphMixedResult[graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "show", "dep-consumer", "--format", "graph-json", "--json"))
			if created.Properties.IssueType != types.TypeTask || len(created.Owned) != 1 || len(consumer.Owned) != 1 || created.Attribution.Inception != "create" {
				t.Fatalf("create did not publish source-owned canonical Links: new=%+v consumer=%+v", created, consumer)
			}
			for _, check := range []struct {
				owned  json.RawMessage
				source string
				target string
			}{
				{created.Owned[0], created.ID, prereq.ID},
				{consumer.Owned[0], consumer.ID, created.ID},
			} {
				var link graphstore.LinkRecord
				if err := json.Unmarshal(check.owned, &link); err != nil {
					t.Fatal(err)
				}
				if link.Source != check.source || link.Target != check.target || link.Properties["dep_type"] != string(types.DepBlocks) {
					t.Fatalf("wrong canonical Link for native dependency: %+v", link)
				}
			}
			graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "create", "External", "--id", "dep-external", "--deps", "blocked-by:external:other#1", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "not_found", "show", "dep-external", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "not_found", "create", "Missing", "--id", "dep-missing-source", "--deps", "dep-absent", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "not_found", "show", "dep-missing-source", "--json")
			parent := graphMixedResult[graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "create", "Parent", "--id", "dep-parent", "--json"))
			child := graphMixedResult[graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "create", "Child", "--parent", "dep-parent", "--deps", "blocked-by:dep-prereq,blocks:dep-consumer", "--json"))
			consumer = graphMixedResult[graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "show", "dep-consumer", "--format", "graph-json", "--json"))
			if child.Properties.ID != "dep-parent.1" || len(child.Owned) != 2 || len(consumer.Owned) != 2 {
				t.Fatalf("child and dependency Links did not commit together: parent=%+v child=%+v consumer=%+v", parent, child, consumer)
			}
			rooted := graphMixedResult[graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "create", "Rooted", "--id-root", "notes/team", "--deps", "blocked-by:dep-prereq", "--json"))
			if !strings.HasPrefix(rooted.Properties.ID, "notes/team/") || len(rooted.Owned) != 1 {
				t.Fatalf("rooted Issue lost its local dependency: %+v", rooted)
			}
			found := map[string]bool{}
			for _, owned := range child.Owned {
				var link graphstore.LinkRecord
				if err := json.Unmarshal(owned, &link); err != nil {
					t.Fatal(err)
				}
				found[link.Target+"|"+link.Properties["dep_type"].(string)] = true
			}
			if !found[parent.ID+"|parent-child"] || !found[prereq.ID+"|blocks"] {
				t.Fatalf("child lacks parent and prerequisite Links: %+v", child.Owned)
			}

			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			selected := graphMixedResult[graphstore.IssueRecord](t, call("create", "Abbreviation target", "--id", "dep-a3f8e9"))
			for _, tc := range []struct {
				name, operand string
			}{
				{"bare", "a3f8"},
				{"configured-prefix", "dep-a3f8"},
				{"canonical-url", selected.ID},
			} {
				got := graphMixedResult[graphstore.IssueRecord](t, call("create", tc.name, "--id", "dep-selected-"+tc.name, "--deps", "blocked-by:"+tc.operand))
				if len(got.Owned) != 1 || got.Attribution.Inception != "create" {
					t.Fatalf("%s did not atomically create its owned Link: %+v", tc.name, got)
				}
				var link graphstore.LinkRecord
				if err := json.Unmarshal(got.Owned[0], &link); err != nil {
					t.Fatal(err)
				}
				if link.Source != got.ID || link.Target != selected.ID || link.Properties["dep_type"] != string(types.DepBlocks) {
					t.Fatalf("%s resolved the wrong target or direction: %+v", tc.name, link)
				}
			}
			abbreviatedChild := graphMixedResult[graphstore.IssueRecord](t, call("create", "Abbreviated child", "--parent", "dep-parent", "--deps", "blocked-by:a3f8"))
			if len(abbreviatedChild.Owned) != 2 {
				t.Fatalf("abbreviated child lost its parent/dependency Link: %+v", abbreviatedChild)
			}
			childTargets := map[string]bool{}
			for _, raw := range abbreviatedChild.Owned {
				var link graphstore.LinkRecord
				if err := json.Unmarshal(raw, &link); err != nil {
					t.Fatal(err)
				}
				if link.Source != abbreviatedChild.ID {
					t.Fatalf("abbreviated child Link has wrong source: %+v", link)
				}
				childTargets[link.Target+"|"+link.Properties["dep_type"].(string)] = true
			}
			if !childTargets[parent.ID+"|parent-child"] || !childTargets[selected.ID+"|blocks"] {
				t.Fatalf("abbreviated child selected wrong parent/target: %+v", abbreviatedChild.Owned)
			}
			refuse := func(name, operand, code string) {
				t.Helper()
				beforeOwner := call("show", prereq.ID, "--format", "graph-json")
				beforeTarget := call("show", selected.ID, "--format", "graph-json")
				beforeLinks := call("link", "list")
				id := "dep-refused-" + name
				// The valid reverse dependency must not publish an existing
				// owner's revision or Link when the next target refuses.
				graphPolicyCLI(t, bd, work, home, nil, code, "create", name, "--id", id, "--deps", "blocks:dep-prereq,blocked-by:"+operand, "--json")
				graphPolicyCLI(t, bd, work, home, nil, "not_found", "show", id, "--json")
				if after := call("show", prereq.ID, "--format", "graph-json"); after != beforeOwner {
					t.Fatalf("%s refusal changed the preceding dependency owner", name)
				}
				if after := call("show", selected.ID, "--format", "graph-json"); after != beforeTarget {
					t.Fatalf("%s refusal changed the abbreviation target", name)
				}
				if after := call("link", "list"); after != beforeLinks {
					t.Fatalf("%s refusal changed the Link inventory", name)
				}
			}
			refuse("exact-short-url", "https://example.invalid/create-deps/a3f8", "not_found")
			call("create", "Ambiguous target", "--id", "dep-a3f8e1")
			refuse("ambiguous", "a3f8", "invalid_properties")
			exact := graphMixedResult[graphstore.IssueRecord](t, call("create", "Exact target", "--id", "dep-a3f8"))
			exactCreated := graphMixedResult[graphstore.IssueRecord](t, call("create", "Exact precedence", "--id", "dep-exact-source", "--deps", "blocked-by:dep-a3f8"))
			if len(exactCreated.Owned) != 1 {
				t.Fatalf("exact ID lost its dependency: %+v", exactCreated)
			}
			var exactLink graphstore.LinkRecord
			if err := json.Unmarshal(exactCreated.Owned[0], &exactLink); err != nil {
				t.Fatal(err)
			}
			if exactLink.Source != exactCreated.ID || exactLink.Target != exact.ID {
				t.Fatalf("exact ID fell through to a longer Issue: %+v", exactLink)
			}
			memory := graphMixedResult[graphstore.Record](t, call("remember", "Exact Memory", "--id", "a3f8", "--create-only"))
			refuse("memory", "a3f8", "invalid_properties")
			call("create", "Link decoy", "--id", "dep-b7c4d2")
			call("link", "create", "--type", "references", "--id", "b7c4", "--from", memory.ID, "--to", prereq.ID, "--if-source-revision", memory.Revision)
			refuse("link", "b7c4", "invalid_properties")
			dead := graphMixedResult[graphstore.IssueRecord](t, call("create", "Deleted exact target", "--id", "dep-d3ad"))
			call("create", "Deleted target decoy", "--id", "dep-d3adbeef")
			deleted := graphMixedResult[graphstore.IssueDeleteResult](t, call("delete", "--force", dead.ID, "--if-revision", "="+dead.Revision))
			if !deleted.Deleted || deleted.Terminal == nil {
				t.Fatalf("tombstone fixture did not delete its exact Issue: %+v", deleted)
			}
			graphPolicyCLI(t, bd, work, home, nil, "gone", "show", dead.ID, "--json")
			refuse("tombstone", "dep-d3ad", "gone")
		})
	}
}
