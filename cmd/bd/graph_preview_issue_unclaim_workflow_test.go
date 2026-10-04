//go:build cgo

package main

import (
	"os"
	"reflect"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
)

// Every operation is a fresh installed bd process against real embedded or
// ordinary shared-server Dolt. The server lane is mandatory in graph CI.
func TestGraphPreviewIssueUnclaimWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			args := []string{"init", "--graph-mode", "link", "--scope-url", "https://example.invalid/unclaim/", "--skip-hooks", "--skip-agents", "--non-interactive"}
			if engine == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for ordinary shared-server unclaim qualification")
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
			call("create", "Claim work", "--id", "beads/work", "--labels", "demo,graph")
			assigned := graphMixedResult[graphstore.IssueRecord](t, call("create", "Assigned work", "--id", "beads/assigned", "--assignee", "rig.agent"))
			refuse("constraint_violation", "unclaim", "assigned", "--actor", "rig.agent")
			if got := graphMixedResult[graphstore.IssueRecord](t, call("show", "assigned")); !reflect.DeepEqual(got, assigned) {
				t.Fatal("unclaim erased an initially assigned open Issue")
			}
			call("remember", "Context", "--id", "beads/context")
			before := graphMixedResult[graphstore.IssueRecord](t, call("show", "work"))
			claimed := graphMixedResult[graphstore.IssueMutationResult](t, call("update", "work", "--claim", "--actor", "rig.agent"))
			if !claimed.Changed || claimed.Issue.Properties.Status != types.StatusInProgress {
				t.Fatalf("claim did not start work: %+v", claimed)
			}
			refuse("constraint_violation", "unclaim", "work", "--actor", "other.agent")
			refuse("capability_unavailable", "unclaim", "work", "--actor", "other.agent", "--force")
			refuse("capability_unavailable", "unclaim", "work", "--if-assignee", "rig.agent")
			refuse("capability_unavailable", "unclaim", "work", "--reason", "leaving")
			refuse("permission_denied", "unclaim", "work", "--actor", "rig.agent", "--readonly")
			refuse("invalid_properties", "unclaim", "context", "--actor", "rig.agent")
			refuse("invalid_selector", "unclaim", "links/context", "--actor", "rig.agent")
			if got := graphMixedResult[graphstore.IssueRecord](t, call("show", "work")); !reflect.DeepEqual(got, claimed.Issue) {
				t.Fatal("refused releases changed live claim")
			}
			released := graphMixedResult[graphstore.IssueMutationResult](t, call("unclaim", "work", "--actor", "rig_agent"))
			if !released.Changed || released.Issue.Properties.Status != types.StatusOpen || released.Issue.Properties.Assignee != "" || released.Issue.Properties.LeaseExpiresAt != nil || released.Issue.Revision == claimed.Issue.Revision || released.Issue.Attribution.Actor != "rig_agent" {
				t.Fatalf("incomplete release: %+v", released)
			}
			for _, record := range []graphstore.IssueRecord{before, claimed.Issue, released.Issue} {
				got := graphMixedResult[graphstore.IssueRecord](t, call("show", "work", "--version", record.Version))
				properties := *record.Properties
				properties.ContentHash, properties.RowVersion = "", 0
				record.Properties = &properties
				if !reflect.DeepEqual(got, record) {
					t.Fatalf("retained version changed: got=%+v want=%+v", got, record)
				}
			}
			refuse("constraint_violation", "unclaim", "work", "--actor", "rig.agent")
			reclaimed := graphMixedResult[graphstore.IssueMutationResult](t, call("update", "work", "--claim", "--actor", "next.agent"))
			if !reclaimed.Changed || reclaimed.Issue.Properties.Assignee != "next.agent" {
				t.Fatalf("reclaim after release: %+v", reclaimed)
			}
		})
	}
}
