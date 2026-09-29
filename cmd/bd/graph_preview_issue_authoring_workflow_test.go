//go:build cgo

package main

import (
	"os"
	"reflect"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func TestGraphPreviewIssueAuthoringWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			const scope = "https://example.invalid/authoring/"
			args := []string{"init", "--graph-mode", "link", "--scope-url", scope, "--skip-hooks", "--skip-agents", "--non-interactive"}
			if engine == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for ordinary shared-server authoring qualification")
				}
				args = append(args, "--server", "--external", "--server-host", "127.0.0.1", "--server-port", port, "--server-user", "root")
			}
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			exact := func(record graphstore.IssueRecord) {
				t.Helper()
				got := graphMixedResult[graphstore.IssueRecord](t, call("show", record.ID, "--version", record.Version))
				p := *record.Properties
				p.ContentHash, p.RowVersion = "", 0
				record.Properties = &p
				if !reflect.DeepEqual(got, record) {
					t.Fatal("exact predecessor/current changed")
				}
			}
			call(args...)
			memory := call("remember", "Context", "--id", "beads/context", "--title", "Context")
			initial := graphMixedResult[graphstore.IssueRecord](t, call("create", "Authored", "--id", "beads/work", "--design", "Design — 雪", "--acceptance", "Done\r\n", "--assignee", "author", "--estimate=0", "--external-ref", " tracker #1 ", "--spec-id", " spec ", "--notes", " Initial\r\n雪 ", "--actor", "author"))
			p := initial.Properties
			if p.Design != "Design — 雪" || p.AcceptanceCriteria != "Done\r\n" || p.Assignee != "author" || p.EstimatedMinutes == nil || *p.EstimatedMinutes != 0 || p.ExternalRef == nil || *p.ExternalRef != " tracker #1 " || p.SpecID != " spec " || p.Notes != " Initial\r\n雪 " || p.StartedAt != nil || p.LeaseExpiresAt != nil {
				t.Fatalf("initial fields/presence/lease: %+v", p)
			}
			if got := graphMixedResult[graphstore.IssueRecord](t, call("show", initial.ID)); !reflect.DeepEqual(got, initial) {
				t.Fatal("fresh initial read differs")
			}
			exact(initial)
			target := call("create", "Gate", "--id", "beads/gate")
			dep := graphMixedResult[graphstore.DependencyResult](t, call("dep", "add", "beads/work", "beads/gate"))
			info := graphMixedResult[graphstore.LinkMutationResult](t, call("link", "beads/work", "beads/context", "--id", "links/context", "--resource-type", scope+"types/preview-related-v2"))
			current := graphMixedResult[graphstore.IssueRecord](t, call("show", "beads/work"))
			before := current
			edit := []string{"update", "beads/work", "--estimate=45", "--external-ref=next", "--spec-id=next spec", "--design=Revised", "--acceptance=Accepted", "--append-notes=Progress", "--if-revision", current.Revision, "--actor", "author"}
			changed := graphMixedResult[graphstore.IssueMutationResult](t, call(edit...))
			current = changed.Issue
			if !changed.Changed || current.Revision == before.Revision || !reflect.DeepEqual(current.Owned, before.Owned) || *current.Properties.EstimatedMinutes != 45 || *current.Properties.ExternalRef != "next" || current.Properties.SpecID != "next spec" || current.Properties.Design != "Revised" || current.Properties.AcceptanceCriteria != "Accepted" || current.Properties.Notes != initial.Properties.Notes+"\nProgress" {
				t.Fatalf("combined authoring lost fields/ownership: %+v", changed)
			}
			if got := graphMixedResult[graphstore.IssueRecord](t, call("show", current.ID)); !reflect.DeepEqual(got, current) {
				t.Fatal("fresh edited read differs")
			}
			exact(before)
			exact(current)
			state := graphMemoryReadSnapshot(t, work)
			noop := graphMixedResult[graphstore.IssueMutationResult](t, call("update", "beads/work", "--estimate=45", "--external-ref=next", "--spec-id=next spec", "--if-revision", current.Revision))
			if noop.Changed || !reflect.DeepEqual(noop.Issue, current) || !reflect.DeepEqual(state, graphMemoryReadSnapshot(t, work)) {
				t.Fatal("authoring noop mutated complete state")
			}
			graphPolicyCLI(t, bd, work, home, nil, "revision_conflict", append(edit, "--json")...)
			for _, refusal := range []struct {
				code string
				args []string
			}{
				{"invalid_properties", []string{"update", "beads/work", "--estimate=-1", "--unconditional"}},
				{"capability_unavailable", []string{"update", "beads/work", "--notes=Replace", "--unconditional"}},
				{"capability_unavailable", []string{"update", "beads/work", "--estimate=1", "--claim"}},
				{"permission_denied", []string{"update", "beads/work", "--spec-id=No", "--unconditional", "--readonly"}},
			} {
				graphPolicyCLI(t, bd, work, home, nil, refusal.code, append(refusal.args, "--json")...)
			}
			if !reflect.DeepEqual(state, graphMemoryReadSnapshot(t, work)) {
				t.Fatal("refusal mutated complete state")
			}
			before = current
			cleared := graphMixedResult[graphstore.IssueMutationResult](t, call("update", "beads/work", "--estimate=0", "--external-ref=", "--spec-id=", "--if-revision", current.Revision))
			current = cleared.Issue
			if !cleared.Changed || current.Properties.EstimatedMinutes == nil || *current.Properties.EstimatedMinutes != 0 || current.Properties.ExternalRef != nil || current.Properties.SpecID != "" || current.Properties.Notes != before.Properties.Notes {
				t.Fatal("nullable clear/zero or notes preservation lost")
			}
			exact(before)
			exact(current)
			if call("show", "beads/context") != memory || call("show", "beads/gate") != target {
				t.Fatal("authoring changed neighboring Bead")
			}
			if got := graphMixedResult[graphstore.LinkRecord](t, call("show", dep.Link.ID)); !reflect.DeepEqual(got, dep.Link) {
				t.Fatal("authoring changed Dependency")
			}
			if got := graphMixedResult[graphstore.LinkRecord](t, call("show", info.Link.ID)); !reflect.DeepEqual(got, info.Link) {
				t.Fatal("authoring changed informational Link")
			}
		})
	}
}
