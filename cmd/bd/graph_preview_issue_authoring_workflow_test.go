//go:build cgo

package main

import (
	"os"
	"reflect"
	"testing"
	"time"

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
			past := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
			future := time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)
			selected := func(want bool, filters ...string) {
				t.Helper()
				page := graphMixedResult[graphstore.IssueListPage](t, graphPolicyCLI(t, bd, work, home, nil, "", append([]string{"list", "--format=records-json", "--all", "--limit=10"}, filters...)...))
				if page.HasMore || (want && (len(page.Items) != 1 || page.Items[0].ID != scope+"beads/work")) || (!want && len(page.Items) != 0) {
					t.Fatalf("due selection %v: %+v", filters, page)
				}
			}
			call(args...)
			memory := call("remember", "Context", "--id", "beads/context", "--title", "Context")
			initial := graphMixedResult[graphstore.IssueRecord](t, call("create", "Authored", "--id", "beads/work", "--design", "Design — 雪", "--acceptance", "Done\r\n", "--assignee", "author", "--estimate=0", "--external-ref", " tracker #1 ", "--spec-id", " spec ", "--notes", " Initial\r\n雪 ", "--due=2000-01-01T00:00:00Z", "--actor", "author"))
			p := initial.Properties
			if p.Design != "Design — 雪" || p.AcceptanceCriteria != "Done\r\n" || p.Assignee != "author" || p.EstimatedMinutes == nil || *p.EstimatedMinutes != 0 || p.ExternalRef == nil || *p.ExternalRef != " tracker #1 " || p.SpecID != " spec " || p.Notes != " Initial\r\n雪 " || p.StartedAt != nil || p.LeaseExpiresAt != nil || p.DueAt == nil || !p.DueAt.Equal(past) {
				t.Fatalf("initial fields/presence/lease: %+v", p)
			}
			if got := graphMixedResult[graphstore.IssueRecord](t, call("show", initial.ID)); !reflect.DeepEqual(got, initial) {
				t.Fatal("fresh initial read differs")
			}
			exact(initial)
			selected(true, "--due-after=1999-01-01T00:00:00Z", "--due-before=2001-01-01T00:00:00Z", "--overdue")
			selected(false, "--due-before=2000-01-01T00:00:00Z")
			selected(false, "--due-after=2000-01-01T00:00:00Z")
			target := call("create", "Gate", "--id", "beads/gate")
			dep := graphMixedResult[graphstore.DependencyResult](t, call("dep", "add", "beads/work", "beads/gate"))
			info := graphMixedResult[graphstore.LinkMutationResult](t, call("link", "beads/work", "beads/context", "--id", "links/context", "--resource-type", scope+"types/preview-related-v2"))
			current := graphMixedResult[graphstore.IssueRecord](t, call("show", "beads/work"))
			before := current
			edit := []string{"update", "beads/work", "--estimate=45", "--external-ref=next", "--spec-id=next spec", "--due=2100-01-01T00:00:00Z", "--design=Revised", "--acceptance=Accepted", "--append-notes=Progress", "--if-revision", current.Revision, "--actor", "author"}
			changed := graphMixedResult[graphstore.IssueMutationResult](t, call(edit...))
			current = changed.Issue
			if !changed.Changed || current.Revision == before.Revision || !reflect.DeepEqual(current.Owned, before.Owned) || *current.Properties.EstimatedMinutes != 45 || *current.Properties.ExternalRef != "next" || current.Properties.SpecID != "next spec" || current.Properties.Design != "Revised" || current.Properties.AcceptanceCriteria != "Accepted" || current.Properties.Notes != initial.Properties.Notes+"\nProgress" || current.Properties.DueAt == nil || !current.Properties.DueAt.Equal(future) {
				t.Fatalf("combined authoring lost fields/ownership: %+v", changed)
			}
			if got := graphMixedResult[graphstore.IssueRecord](t, call("show", current.ID)); !reflect.DeepEqual(got, current) {
				t.Fatal("fresh edited read differs")
			}
			exact(before)
			exact(current)
			selected(true, "--due-after=2099-01-01T00:00:00Z", "--due-before=2101-01-01T00:00:00Z")
			selected(false, "--overdue")
			state := graphMemoryReadSnapshot(t, work)
			noop := graphMixedResult[graphstore.IssueMutationResult](t, call("update", "beads/work", "--estimate=45", "--external-ref=next", "--spec-id=next spec", "--due=2100-01-01T00:00:00Z", "--if-revision", current.Revision))
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
			cleared := graphMixedResult[graphstore.IssueMutationResult](t, call("update", "beads/work", "--estimate=0", "--external-ref=", "--spec-id=", "--due=", "--if-revision", current.Revision))
			current = cleared.Issue
			if !cleared.Changed || current.Properties.EstimatedMinutes == nil || *current.Properties.EstimatedMinutes != 0 || current.Properties.ExternalRef != nil || current.Properties.SpecID != "" || current.Properties.Notes != before.Properties.Notes || current.Properties.DueAt != nil {
				t.Fatal("nullable clear/zero or notes preservation lost")
			}
			exact(before)
			exact(current)
			selected(false, "--due-before=2101-01-01T00:00:00Z")
			selected(false, "--due-after=1999-01-01T00:00:00Z")
			selected(false, "--overdue")
			state = graphMemoryReadSnapshot(t, work)
			clearAgain := graphMixedResult[graphstore.IssueMutationResult](t, call("update", "beads/work", "--due=", "--if-revision", current.Revision))
			if clearAgain.Changed || !reflect.DeepEqual(clearAgain.Issue, current) || !reflect.DeepEqual(state, graphMemoryReadSnapshot(t, work)) {
				t.Fatal("repeated due clear mutated complete state")
			}
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

func TestGraphPreviewIssueDatelessDeferralWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			args := []string{"init", "--graph-mode", "link", "--scope-url", "https://example.invalid/deferral/", "--skip-hooks", "--skip-agents", "--non-interactive"}
			if engine == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for ordinary shared-server deferral qualification")
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
			original := graphMixedResult[graphstore.IssueRecord](t, call("create", "Work", "--id", "work"))
			memory := call("remember", "Context", "--id", "context", "--title", "Context")
			refuse("invalid_selector", "defer", original.ID)
			refuse("capability_unavailable", "defer", original.ID, "--until", "tomorrow", "--unconditional")
			refuse("capability_unavailable", "defer", original.ID, "--reason", "later", "--unconditional")
			refuse("invalid_selector", "defer", original.ID, "context", "--unconditional")
			refuse("invalid_properties", "defer", "context", "--unconditional")
			if call("show", original.ID) == memory {
				t.Fatal("Issue and Memory identity collided")
			}
			deferred := graphMixedResult[graphstore.IssueMutationResult](t, call("defer", "work", "--if-revision", original.Revision))
			if !deferred.Changed || string(deferred.Issue.Properties.Status) != "deferred" || deferred.Issue.Properties.DeferUntil != nil || deferred.Issue.Version == original.Version {
				t.Fatalf("defer result: %+v", deferred)
			}
			refuse("revision_conflict", "defer", "work", "--if-revision", original.Revision)
			refuse("invalid_selector", "undefer", "work")
			if got := graphMixedResult[[]graphstore.IssueRecord](t, call("ready")); len(got) != 0 {
				t.Fatalf("deferred Issue remained ready: %+v", got)
			}
			noop := graphMixedResult[graphstore.IssueMutationResult](t, call("defer", "work", "--if-revision", deferred.Issue.Revision))
			if noop.Changed || !reflect.DeepEqual(noop.Issue, deferred.Issue) {
				t.Fatalf("repeat defer changed Issue: %+v", noop)
			}
			opened := graphMixedResult[graphstore.IssueMutationResult](t, call("undefer", "work", "--if-revision", deferred.Issue.Revision))
			if !opened.Changed || string(opened.Issue.Properties.Status) != "open" || opened.Issue.Version == deferred.Issue.Version {
				t.Fatalf("undefer result: %+v", opened)
			}
			if got := graphMixedResult[[]graphstore.IssueRecord](t, call("ready")); len(got) != 1 || got[0].ID != original.ID {
				t.Fatalf("undeferred Issue remained hidden: %+v", got)
			}
			if call("show", "context") != memory {
				t.Fatal("deferral changed unrelated Memory")
			}
		})
	}
}
