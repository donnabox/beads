//go:build cgo

package main

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func TestGraphPreviewUsabilityWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			const scope = "https://example.invalid/usability/"
			args := []string{"init", "--graph-mode", "link", "--scope-url", scope, "--skip-hooks", "--skip-agents", "--non-interactive"}
			if engine == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for ordinary shared-server CLI qualification")
				}
				args = append(args, "--server", "--external", "--server-host", "127.0.0.1", "--server-port", port, "--server-user", "root")
			}
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			refuse := func(code string, args ...string) {
				t.Helper()
				graphPolicyCLI(t, bd, work, home, nil, code, append(args, "--json")...)
			}
			call(args...)
			memory := graphMixedResult[graphstore.Record](t, call("create", "--bead-type", "types/preview-memory-v2", "--id", "beads/policy", "--body", "Code flow policy\nTarget integration."))
			other := graphMixedResult[graphstore.Record](t, call("create", "Old code flow policy", "--bead-type", graphstore.MemoryTypeURL(scope)))
			issue := graphMixedResult[graphstore.IssueRecord](t, call("create", "Move the branch", "--bead-type", "types/preview-issue-v2", "--type", "task"))
			if memory.ID != scope+"beads/policy" || memory.Properties.Title != "Code flow policy" || other.ID == memory.ID || other.Properties.Body != "" || issue.Type != graphstore.IssueTypeURL(scope) {
				t.Fatal("typed creation lost identity, body or Type")
			}
			refuse("identity_reserved", "create", "No overwrite", "--bead-type", "types/preview-memory-v2", "--id", "beads/policy")
			refuse("capability_unavailable", "create", "Bad", "--bead-type", "types/preview-memory-v2", "--priority", "2")
			refuse("invalid_selector", "link", memory.ID, other.ID, "--link-type", "types/example-follows", "--resource-type", "types/example-follows")
			refuse("invalid_selector", "link", memory.ID, other.ID, "--link-type", "https://foreign.invalid/types/example-follows")
			refuse("capability_unavailable", "link", memory.ID, other.ID, "--link-type", "types/missing")
			refuse("invalid_properties", "link", memory.ID, other.ID, "--link-type", "types/preview-blocks-v1", "--unconditional-source")
			human := graphPolicyCLI(t, bd, work, home, nil, "", "memories")
			if !strings.Contains(human, "beads/policy") || !strings.Contains(human, "Code flow policy") || strings.Contains(human, scope) || strings.Contains(human, memory.Version) {
				t.Fatalf("noisy Memory list: %s", human)
			}
			details := graphPolicyCLI(t, bd, work, home, nil, "", "memories", "--details")
			if !strings.Contains(details, memory.Version) {
				t.Fatal("details omitted exact version")
			}
			for _, spec := range []struct{ typ, target string }{{"types/example-follows", other.ID}, {"types/example-cites", issue.ID}} {
				link := graphMixedResult[graphstore.LinkMutationResult](t, call("link", memory.ID, spec.target, "--link-type", spec.typ, "--properties", `{"note":"code flow"}`))
				if link.Link.Type != scope+spec.typ || link.Link.Source != memory.ID || link.Link.Target != spec.target || link.ReplacedSource == nil {
					t.Fatal("informational Link lost Type/endpoints/ownership")
				}
				owner := graphMixedResult[graphstore.Record](t, call("show", memory.ID))
				var owned graphstore.LinkRecord
				if len(owner.Owned) != 1 || json.Unmarshal(owner.Owned[0], &owned) != nil || owned.Type != scope+spec.typ {
					t.Fatal("selected Type not retained as owned")
				}
				incident := graphMixedResult[[]graphstore.LinkRecord](t, call("links", memory.ID, "--link-type", spec.typ))
				if len(incident) != 1 || incident[0].Type != scope+spec.typ {
					t.Fatal("local Type filter failed")
				}
				edited := graphMixedResult[graphstore.LinkMutationResult](t, call("update", link.Link.ID, "--properties", `{"note":"updated"}`, "--if-revision", link.Link.Revision))
				if edited.Link.Type != link.Link.Type {
					t.Fatal("update changed Type")
				}
				old := graphMixedResult[graphstore.LinkRecord](t, call("show", link.Link.ID, "--version", link.Link.Version))
				if !reflect.DeepEqual(old, link.Link) {
					t.Fatal("retained Link changed")
				}
				call("unlink", memory.ID, spec.target, "--link-type", scope+spec.typ, "--if-revision", edited.Link.Revision)
				retained := graphMixedResult[graphstore.Record](t, call("show", memory.ID, "--version", owner.Version))
				if !reflect.DeepEqual(retained, owner) {
					t.Fatal("unlink rewrote old Memory-owned snapshot")
				}
			}
			// Legacy spelling remains accepted for existing scripts.
			call("link", memory.ID, other.ID, "--resource-type", "types/preview-related-v2")
			ordinaryWork, ordinaryHome := t.TempDir(), t.TempDir()
			graphPolicyCLI(t, bd, ordinaryWork, ordinaryHome, nil, "capability_unavailable", "create", "--bead-type", "types/preview-memory-v2", "--body", "No store", "--json")
		})
	}
}
