//go:build cgo

package main

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func graphMixedResult[T any](t *testing.T, output string) T {
	t.Helper()
	var envelope struct {
		Result T `json:"result"`
	}
	if err := json.Unmarshal([]byte(output), &envelope); err != nil {
		t.Fatalf("decode command output: %v\n%s", err, output)
	}
	return envelope.Result
}

// One complete installed-command sequence keeps the transferred adapters honest.
// Every command runs in a new process; the required CI lane supplies the server.
func TestGraphPreviewMixedCoreWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			const scope = "https://example.invalid/mixed/"
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
			memory := graphMixedResult[graphstore.Record](t, call("remember", "Original body", "--id", "beads/plan", "--title", "Plan"))
			other := call("remember", "Independent body", "--id", "beads/context", "--title", "Context")
			call("create", "Ship the release", "--id", "beads/work", "--description", "Original Issue", "--priority", "1")
			gate := call("create", "Prerequisite", "--id", "beads/gate")

			selected := graphMixedResult[graphstore.MemoryMutationResult](t, call("remember", "--update", "beads/plan", "--title", "Selected title", "--if-revision", memory.Revision))
			if !selected.Changed || selected.Memory.Properties.Body != memory.Properties.Body || selected.Replaced != nil {
				t.Fatal("selected guarded update lost omitted body or claimed unconditional overwrite")
			}
			replaced := graphMixedResult[graphstore.MemoryMutationResult](t, call("update", "beads/plan", "--properties", `{"title":"Final plan","body":"Replaced body"}`, "--unconditional"))
			if !replaced.Changed || replaced.Replaced == nil || replaced.Memory.Revision == selected.Memory.Revision {
				t.Fatal("unconditional Memory replacement omitted predecessor disclosure")
			}
			before := call("show", "beads/plan")
			refuse("revision_conflict", "remember", "--update", "beads/plan", "--title", "Final plan", "--if-revision", memory.Revision)
			if call("show", "beads/plan") != before || call("show", "beads/context") != other || call("show", "beads/gate") != gate {
				t.Fatal("stale Memory update changed state")
			}

			related := scope + "types/preview-related-v2"
			first := graphMixedResult[graphstore.LinkMutationResult](t, call("link", "beads/plan", "beads/context", "--resource-type", related, "--id", "links/context", "--properties", `{"note":"first"}`, "--if-source-revision", replaced.Memory.Revision))
			current := graphMixedResult[graphstore.Record](t, call("show", "beads/plan"))
			call("link", "beads/plan", "beads/work", "--resource-type", related, "--id", "links/work", "--if-source-revision", current.Revision)
			issueBefore := call("show", "beads/work")
			call("link", "beads/work", "beads/context", "--resource-type", related, "--id", "links/unowned")
			if call("show", "beads/work") != issueBefore || call("show", "beads/context") != other {
				t.Fatal("unowned informational Link versioned an endpoint")
			}
			incident := graphMixedResult[[]graphstore.LinkRecord](t, call("links", "beads/plan"))
			if len(incident) != 2 {
				t.Fatalf("Memory incident Links = %d, want 2", len(incident))
			}
			current = graphMixedResult[graphstore.Record](t, call("show", "beads/plan"))
			changedLink := graphMixedResult[graphstore.LinkMutationResult](t, call("update", "links/context", "--properties", `{"note":"replaced"}`, "--if-revision", first.Link.Revision, "--if-source-revision", current.Revision))
			if !changedLink.Changed || changedLink.Link.Properties["note"] != "replaced" {
				t.Fatal("Link properties were not replaced")
			}
			before = call("show", "beads/plan")
			linkBefore := call("show", "links/context")
			refuse("revision_conflict", "update", "links/context", "--properties", `{"note":"stale"}`, "--if-revision", first.Link.Revision, "--unconditional-source")
			if call("show", "beads/plan") != before || call("show", "links/context") != linkBefore {
				t.Fatal("stale Link update changed Link or owning Memory")
			}

			dependency := graphMixedResult[graphstore.DependencyResult](t, call("dep", "add", "beads/work", "beads/gate"))
			readyContains := func(id string) bool {
				t.Helper()
				for _, issue := range graphMixedResult[[]graphstore.IssueRecord](t, call("ready")) {
					if issue.ID == scope+id {
						return true
					}
				}
				return false
			}
			if readyContains("beads/work") {
				t.Fatal("blocked Issue appeared ready")
			}
			call("close", "beads/gate", "--reason", "Done")
			if !readyContains("beads/work") {
				t.Fatal("closing prerequisite did not unblock Issue")
			}
			workIssue := graphMixedResult[graphstore.IssueRecord](t, call("show", "beads/work"))
			call("unlink", dependency.Link.ID, "--if-revision", dependency.Link.Revision, "--if-source-revision", workIssue.Revision)
			call("close", "beads/work", "--reason", "Done")
			call("reopen", "beads/work", "--reason", "One more change")
			workIssue = graphMixedResult[graphstore.IssueRecord](t, call("show", "beads/work"))
			edited := graphMixedResult[graphstore.IssueMutationResult](t, call("update", "beads/work", "--description", "Revised Issue", "--if-revision", workIssue.Revision))
			if !edited.Changed || edited.Issue.Properties.Description != "Revised Issue" {
				t.Fatal("Issue text edit was not retained")
			}
			refuse("revision_conflict", "update", "beads/work", "--description", "Revised Issue", "--if-revision", workIssue.Revision)
			if got := graphMixedResult[graphstore.IssueRecord](t, call("show", "beads/work")); !reflect.DeepEqual(got, edited.Issue) {
				t.Fatal("stale same-value Issue update changed state")
			}

			for _, linkPath := range []string{"links/context", "links/work", "links/unowned"} {
				link := graphMixedResult[graphstore.LinkRecord](t, call("show", linkPath))
				args := []string{"unlink", linkPath, "--if-revision", link.Revision}
				if linkPath != "links/unowned" {
					owner := graphMixedResult[graphstore.Record](t, call("show", "beads/plan"))
					args = append(args, "--if-source-revision", owner.Revision)
				}
				call(args...)
				refuse("gone", "show", linkPath)
			}
			if links := graphMixedResult[[]graphstore.LinkRecord](t, call("links", "beads/plan")); len(links) != 0 {
				t.Fatal("unlinked Memory retained current incident Links")
			}
			final := graphMixedResult[graphstore.Record](t, call("show", scope+"beads/plan"))
			if final.Properties != replaced.Memory.Properties || len(final.Owned) != 0 || call("show", "beads/context") != other {
				t.Fatal("final fresh-process Memory state differs from the mixed workflow")
			}
		})
	}
}
