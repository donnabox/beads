//go:build cgo

package main

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
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

			custom := graphMixedResult[graphstore.IssueRecord](t, call("create", "Custom path", "--id", "beads/custom-path"))
			if custom.Properties.ID == "custom-path" {
				t.Fatal("fixture needs a native ID distinct from its graph path")
			}
			if got := issue(call("priority", custom.Properties.ID, "1")); got.ID != "beads/custom-path" || got.Priority != 1 {
				t.Fatalf("native-ID priority selected the wrong Issue: %+v", got)
			}
			if got := issue(call("note", custom.Properties.ID, "Native ID note")); got.ID != "beads/custom-path" || got.Notes != "Native ID note" {
				t.Fatalf("native-ID note selected the wrong Issue: %+v", got)
			}
			if got := issue(call("assign", custom.Properties.ID, "carol")); got.ID != "beads/custom-path" || got.Assignee != "carol" {
				t.Fatalf("native-ID assign selected the wrong Issue: %+v", got)
			}
			parts := strings.SplitN(custom.Properties.ID, "-", 2)
			if len(parts) != 2 {
				t.Fatalf("fixture needs a configured Issue prefix: %q", custom.Properties.ID)
			}
			abbreviable := graphMixedResult[graphstore.IssueRecord](t, call("create", "Abbreviable", "--id", parts[0]+"-a3f8e9"))
			partialID := "a3f8"
			if got := issue(call("priority", partialID, "2")); got.ID != "beads/"+abbreviable.Properties.ID || got.Priority != 2 {
				t.Fatalf("abbreviated priority selected the wrong Issue: %+v", got)
			}
			if got := issue(call("note", partialID, "Partial ID note")); got.ID != "beads/"+abbreviable.Properties.ID || got.Notes != "Partial ID note" {
				t.Fatalf("abbreviated note selected the wrong Issue: %+v", got)
			}
			if got := issue(call("assign", partialID, "dana")); got.ID != "beads/"+abbreviable.Properties.ID || got.Assignee != "dana" {
				t.Fatalf("abbreviated assign selected the wrong Issue: %+v", got)
			}
			var stateResult struct {
				IssueID string `json:"issue_id"`
			}
			if err := json.Unmarshal([]byte(call("state", "list", partialID)), &stateResult); err != nil || stateResult.IssueID != abbreviable.Properties.ID {
				t.Fatalf("abbreviated state selected the wrong Issue: %+v %v", stateResult, err)
			}
			call("remember", "Colliding Memory", "--id", "beads/"+custom.Properties.ID)
			beforeCollision := graphMemoryReadSnapshot(t, work)
			refuse("invalid_properties", "priority", custom.Properties.ID, "0")
			refuse("invalid_properties", "note", custom.Properties.ID, "Wrong target")
			refuse("invalid_properties", "assign", custom.Properties.ID, "wrong")
			if !reflect.DeepEqual(beforeCollision, graphMemoryReadSnapshot(t, work)) {
				t.Fatal("ambiguous native ID changed Issue or Memory state")
			}
			if got := issue(call("priority", "beads/custom-path", "0")); got.Priority != 0 {
				t.Fatalf("explicit path did not select custom Issue: %+v", got)
			}
		})
	}
}
