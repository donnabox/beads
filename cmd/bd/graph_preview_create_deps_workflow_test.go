//go:build cgo

package main

import (
	"encoding/json"
	"os"
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
			graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "create", "Child", "--parent", "dep-prereq", "--deps", "dep-consumer", "--json")
		})
	}
}
