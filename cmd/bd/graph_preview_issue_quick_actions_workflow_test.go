//go:build cgo

package main

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
)

func TestGraphPreviewIssueQuickActionsWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			args := []string{"init", "--graph-mode", "link", "--scope-url", "https://example.invalid/quick/", "--skip-hooks", "--skip-agents", "--non-interactive"}
			if engine == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for shared-server qualification")
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
			read := func() graphstore.IssueRecord {
				t.Helper()
				return graphMixedResult[graphstore.IssueRecord](t, call("show", "beads/work", "--format", "graph-json"))
			}
			issue := func(output string) types.Issue {
				t.Helper()
				var result types.Issue
				if err := json.Unmarshal([]byte(output), &result); err != nil {
					t.Fatalf("decode native-shaped Issue: %v\n%s", err, output)
				}
				return result
			}
			call(args...)
			call("create", "Quick actions", "--id", "beads/work", "--description", "Keep this")
			call("remember", "Memory text", "--id", "beads/context", "--title", "Context")
			before := read()
			priority := issue(call("priority", "beads/work", "0"))
			current := read()
			if priority.ID != "beads/work" || priority.Priority != 0 || current.Properties.Priority != 0 || current.Revision == before.Revision || current.Properties.Description != "Keep this" {
				t.Fatalf("priority write did not preserve Issue shape and fields: %+v %+v", priority, current)
			}
			state := graphMemoryReadSnapshot(t, work)
			call("priority", "beads/work", "P0")
			if got := read(); got.Revision != current.Revision || !reflect.DeepEqual(state, graphMemoryReadSnapshot(t, work)) {
				t.Fatal("same priority changed Resource version or native state")
			}
			note := issue(call("note", "beads/work", "Progress", "continues"))
			after := read()
			if note.ID != "beads/work" || note.Notes != "Progress continues" || after.Properties.Notes != note.Notes || after.Revision == current.Revision {
				t.Fatalf("note did not append once: %+v %+v", note, after)
			}
			call("note", "beads/work", "Second")
			if got := read(); got.Properties.Notes != "Progress continues\nSecond" {
				t.Fatalf("second note overwrote prior text: %q", got.Properties.Notes)
			}
			preAssign := read()
			assigned := issue(call("assign", "beads/work", "alice", "--if-revision", preAssign.Revision, "--actor", "alice"))
			if assigned.ID != "beads/work" || assigned.Assignee != "alice" || read().Revision == preAssign.Revision {
				t.Fatalf("assign lost canonical identity or failed to version: %+v", assigned)
			}
			refuse("revision_conflict", "assign", "beads/work", "bob", "--if-revision", preAssign.Revision)
			state = graphMemoryReadSnapshot(t, work)
			call("assign", "beads/work", "alice")
			if !reflect.DeepEqual(state, graphMemoryReadSnapshot(t, work)) {
				t.Fatal("same assignee changed native state")
			}
			claimed := graphMixedResult[graphstore.IssueMutationResult](t, call("update", "beads/work", "--claim", "--actor", "alice"))
			if !claimed.Changed || claimed.Issue.Properties.Assignee != "alice" {
				t.Fatalf("claim setup failed: %+v", claimed)
			}
			refuse("constraint_violation", "assign", "beads/work", "bob", "--actor", "bob")
			transferred := issue(call("assign", "beads/work", "bob", "--actor", "bob", "--force"))
			if transferred.Assignee != "bob" || read().Properties.Assignee != "bob" {
				t.Fatalf("explicit forced transfer failed: %+v", transferred)
			}
			refuse("invalid_properties", "priority", "beads/work", "P9")
			refuse("permission_denied", "assign", "beads/work", "alice", "--readonly")
			refuse("permission_denied", "priority", "beads/work", "2", "--readonly")
			refuse("permission_denied", "note", "beads/work", "No", "--readonly")
			refuse("capability_unavailable", "priority", "beads/context", "2")
			refuse("capability_unavailable", "note", "beads/context", "No")
			refuse("capability_unavailable", "assign", "beads/context", "alice")
		})
	}
}
