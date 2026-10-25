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
	isolateBeadsDirForTest(t)
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			const scope = "https://example.invalid/usability/"
			taskType, err := graphstore.NativeIssueTypeURL(scope, "task")
			if err != nil {
				t.Fatal(err)
			}
			blocksType, err := graphstore.NativeDependencyTypeURL(scope, "blocks")
			if err != nil {
				t.Fatal(err)
			}
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
			var installed struct {
				Result struct {
					BeadTypes []json.RawMessage `json:"beadTypes"`
					LinkTypes []json.RawMessage `json:"linkTypes"`
				} `json:"result"`
			}
			if err := json.Unmarshal([]byte(call("types")), &installed); err != nil || len(installed.Result.BeadTypes) != 22 || len(installed.Result.LinkTypes) != 25 {
				t.Fatalf("installed graph Types unavailable: %+v %v", installed, err)
			}
			for _, item := range append(installed.Result.BeadTypes, installed.Result.LinkTypes...) {
				var descriptor struct{ ID, Describes string }
				if err := json.Unmarshal(item, &descriptor); err != nil || !strings.HasPrefix(descriptor.ID, scope+"types/") {
					t.Fatalf("invalid installed descriptor: %s %v", item, err)
				}
			}
			visible := graphPolicyCLI(t, bd, work, home, nil, "", "types")
			if !strings.Contains(visible, "Bead Types (22):") || !strings.Contains(visible, strings.TrimPrefix(taskType, scope)) ||
				!strings.Contains(visible, "Link Types (25):") || !strings.Contains(visible, strings.TrimPrefix(blocksType, scope)) ||
				!strings.Contains(visible, "task") || strings.Contains(visible, "Issue classifications (") ||
				!strings.Contains(visible, "blocks") || strings.Contains(visible, "Dependency kinds (") ||
				!strings.Contains(visible, "types/subject-to") ||
				strings.Contains(visible, "preview-issue") || strings.Contains(visible, "preview-dependency") {
				t.Fatalf("Type selection help missing: %s", visible)
			}
			detailsText := graphPolicyCLI(t, bd, work, home, nil, "", "types", "--details")
			if !strings.Contains(detailsText, `"describes": "link"`) || !strings.Contains(detailsText, `"source":`) || !strings.Contains(detailsText, `"target":`) {
				t.Fatalf("full Link descriptor unavailable: %s", detailsText)
			}
			refuse("capability_unavailable", "types", "--sections")
			memory := graphMixedResult[graphstore.Record](t, call("create", "--type", "types/memory", "--id", "policy", "--body", "Code flow policy\nTarget integration."))
			other := graphMixedResult[graphstore.Record](t, call("create", "Old code flow policy", "--type", graphstore.MemoryTypeURL(scope)))
			issue := graphMixedResult[graphstore.IssueRecord](t, call("create", "Move the branch", "--graph-path", "beads/work", "--type", "task"))
			if memory.ID != scope+"policy" || memory.Properties.Title != "Code flow policy" || other.ID == memory.ID || other.Properties.Body != "" || issue.ID != scope+issue.Properties.ID || issue.Type != taskType {
				t.Fatal("typed creation lost identity, body or Type")
			}
			if shown := graphMixedResult[graphstore.Record](t, call("show", "policy")); shown.ID != memory.ID {
				t.Fatal("short Bead selector did not resolve to its canonical identity")
			}
			refuse("identity_reserved", "create", "No overwrite", "--type", "types/memory", "--id", "policy")
			refuse("capability_unavailable", "create", "Bad", "--type", "types/memory", "--priority", "2")
			refuse("invalid_properties", "link", memory.ID, other.ID, "--type", "unregistered-dependency")
			refuse("invalid_selector", "link", memory.ID, other.ID, "--type", "https://foreign.invalid/types/example-follows")
			refuse("capability_unavailable", "link", memory.ID, other.ID, "--type", "types/missing")
			refuse("capability_unavailable", "create", "--type", "types/memory-preview", "--properties", `{"title":"One","body":"One is the loneliest number."}`)
			refuse("capability_unavailable", "link", "create", "policy", "work", "--type", "types/missing", "--properties", `{"note":"test"}`)
			refuse("invalid_selector", "link", "create", "policy", "work", "--from", "policy", "--type", "example-cites")
			beforeLinks := call("link", "list", "--incident-to", "policy")
			for _, flags := range [][]string{
				{"--from=policy", "--from=policy", "--to=work"},
				{"--from=policy", "--from=rule", "--to=work"},
				{"--from=policy", "--to=work", "--to=work"},
				{"--from=policy", "--to=work", "--to=rule"},
			} {
				refuse("invalid_selector", append([]string{"link", "create", "--type", "types/references"}, flags...)...)
			}
			if call("link", "list", "--incident-to", "policy") != beforeLinks {
				t.Fatal("duplicate endpoint flags created a Link")
			}
			refuse("invalid_properties", "link", memory.ID, other.ID, "--type", blocksType)
			human := graphPolicyCLI(t, bd, work, home, nil, "", "memories")
			identityRow := false
			for _, row := range strings.Split(human, "\n") {
				columns := strings.Fields(row)
				if len(columns) > 1 && columns[0] == strings.TrimPrefix(memory.ID, scope) && strings.Join(columns[1:], " ") == memory.Properties.Title {
					identityRow = true
				}
			}
			if !identityRow || !strings.Contains(human, "Code flow policy") || strings.Contains(human, scope) || strings.Contains(human, memory.Revision) {
				t.Fatalf("noisy Memory list: %s", human)
			}
			details := graphPolicyCLI(t, bd, work, home, nil, "", "memories", "--details")
			if !strings.Contains(details, memory.Revision) {
				t.Fatal("details omitted exact version")
			}
			constraint := graphMixedResult[graphstore.Record](t, call("create", "Rule", "--type", "types/constraint", "--id", "rule"))
			for index, spec := range []struct{ typ, target string }{{"types/superseded-by", other.ID}, {"types/subject-to", constraint.ID}} {
				args := []string{"link", "create", "policy", strings.TrimPrefix(spec.target, scope), "--type", spec.typ, "--properties", `{"note":"code flow"}`}
				if index == 1 {
					args = []string{"link", "create", "--from", "policy", "--to", strings.TrimPrefix(spec.target, scope), "--type", spec.typ, "--properties", `{"note":"code flow"}`}
				}
				link := graphMixedResult[graphstore.LinkMutationResult](t, call(args...))
				if link.Link.Type != scope+spec.typ || link.Link.Source != memory.ID || link.Link.Target != spec.target || link.ReplacedSource == nil {
					t.Fatal("informational Link lost Type/endpoints/ownership")
				}
				owner := graphMixedResult[graphstore.Record](t, call("show", memory.ID))
				var owned graphstore.LinkRecord
				if len(owner.Owned) != 1 || json.Unmarshal(owner.Owned[0], &owned) != nil || owned.Type != scope+spec.typ {
					t.Fatal("selected Type not retained as owned")
				}
				incident := graphMixedResult[[]graphstore.LinkRecord](t, call("link", "list", "--incident-to", "policy", "--type", spec.typ))
				if len(incident) != 1 || incident[0].Type != scope+spec.typ {
					t.Fatal("local Type filter failed")
				}
				count := graphMixedResult[map[string]int](t, call("link", "count", "--type", spec.typ))
				if count["count"] != 1 {
					t.Fatalf("Link collection count = %d, want 1", count["count"])
				}
				edited := graphMixedResult[graphstore.LinkMutationResult](t, call("link", "update", link.Link.ID, "--properties", `{"note":"updated"}`, "--if-revision", "="+link.Link.Revision))
				if edited.Link.Type != link.Link.Type {
					t.Fatal("update changed Type")
				}
				// Every installed informational Type lists its history, not only Related.
				resource, kind, versions, members := graphVersionsListed(t, call("link", "versions", link.Link.ID))
				if resource != strings.TrimPrefix(link.Link.ID, scope) || kind != "link" || len(versions) != 2 ||
					versions[0].Version != edited.Link.Revision || versions[1].Version != link.Link.Revision {
					t.Fatalf("%s Link versions: resource=%q kind=%q rows=%+v", spec.typ, resource, kind, versions)
				}
				graphVersionsAssertKeys(t, spec.typ, members)
				old := graphMixedResult[graphstore.LinkRecord](t, call("link", "show", link.Link.ID, "--revision", "="+link.Link.Revision))
				if !reflect.DeepEqual(old, link.Link) {
					t.Fatal("retained Link changed")
				}
				if comparison := call("link", "compare", link.Link.ID, "--from", link.Link.Revision, "--to", edited.Link.Revision); !strings.Contains(comparison, "note") {
					t.Fatalf("Link comparison omitted changed property: %s", comparison)
				}
				call("link", "delete", "--force", link.Link.ID, "--if-revision", "="+edited.Link.Revision)
				retained := graphMixedResult[graphstore.Record](t, call("show", memory.ID, "--revision", "="+owner.Revision))
				if !reflect.DeepEqual(retained, owner) {
					t.Fatal("unlink rewrote old Memory-owned snapshot")
				}
			}
			// Removed catalog aliases refuse; current References remain selectable.
			refuse("capability_unavailable", "link", memory.ID, other.ID, "--type", "types/preview-related")
			refuse("invalid_properties", "link", other.ID, memory.ID, "--type", "preview-related-memory")
			call("link", memory.ID, issue.ID, "--type", "types/references")
			// Equal basenames remain separate exact identities only when supplied explicitly.
			sameIDBead := graphMixedResult[graphstore.Record](t, call("create", "Same local basename", "--type", "types/memory", "--id", "beads/edge"))
			shortLink := graphMixedResult[graphstore.LinkMutationResult](t, call("link", memory.ID, other.ID, "--type", "types/references", "--id", "links/edge"))
			if sameIDBead.ID != scope+"beads/edge" || shortLink.Link.ID != scope+"links/edge" {
				t.Fatal("explicit Bead and Link identities were rewritten")
			}
			refuse("identity_reserved", "link", memory.ID, other.ID, "--type", "types/references", "--id", "beads/edge")
			call("link", "delete", "--force", "links/edge", "--if-revision", "="+shortLink.Link.Revision)
			refuse("gone", "show", scope+"links/edge")
			if shown := graphMixedResult[graphstore.Record](t, call("show", "beads/edge")); shown.ID != sameIDBead.ID {
				t.Fatal("Link-only unlink selected the Bead with the same basename")
			}
			call("remember", "Updated code flow policy", "--update", "policy")
			current := graphMixedResult[graphstore.Record](t, call("show", "policy"))
			if current.ID != memory.ID || current.Properties.Body != "Updated code flow policy" {
				t.Fatal("short Memory update changed identity or lost its body")
			}
			call("compare", "policy", "--from", memory.Revision, "--to", current.Revision)
			ordinaryWork, ordinaryHome := t.TempDir(), t.TempDir()
			graphPolicyCLI(t, bd, ordinaryWork, ordinaryHome, nil, "capability_unavailable", "create", "--type", "types/memory", "--body", "No store", "--json")
			graphPolicyCLI(t, bd, ordinaryWork, ordinaryHome, nil, "capability_unavailable", "list", "--type", "types/memory", "--format", "records-json")
		})
	}
}
