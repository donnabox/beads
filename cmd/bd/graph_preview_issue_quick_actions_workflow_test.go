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
			const scope = "https://example.invalid/quick/"
			args := []string{"init", "--graph-mode", "link", "--scope-url", scope, "--skip-hooks", "--skip-agents", "--non-interactive"}
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
				return graphMixedResult[graphstore.IssueRecord](t, call("show", scope+"beads/work", "--format", "graph-json"))
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
			call("create", "Quick actions", "--id", "beads/work", "--force", "--description", "Keep this")
			call("remember", "Memory text", "--id", "beads/context", "--title", "Context")
			beforeTag := read()
			tagged := issue(call("tag", "beads/work", " release "))
			if tagged.ID != "beads/work" || !reflect.DeepEqual(tagged.Labels, []string{"release"}) || read().Revision == beforeTag.Revision {
				t.Fatalf("tag did not update the native Issue and its graph revision: %+v", tagged)
			}
			state := graphMemoryReadSnapshot(t, work)
			call("tag", "beads/work", "release")
			if !reflect.DeepEqual(state, graphMemoryReadSnapshot(t, work)) {
				t.Fatal("repeated tag changed native state")
			}
			if human := graphPolicyCLI(t, bd, work, home, nil, "", "tag", "beads/work", "review"); !strings.Contains(human, `Added label "review" to`) {
				t.Fatalf("tag human confirmation missing: %q", human)
			}
			before := read()
			priority := issue(call("priority", scope+"beads/work", "0"))
			current := read()
			if priority.ID != current.Properties.ID || priority.Priority != 0 || current.Properties.Priority != 0 || current.Revision == before.Revision || current.Properties.Description != "Keep this" {
				t.Fatalf("priority write did not preserve Issue shape and fields: %+v %+v", priority, current)
			}
			state = graphMemoryReadSnapshot(t, work)
			call("priority", scope+"beads/work", "P0")
			if got := read(); got.Revision != current.Revision || !reflect.DeepEqual(state, graphMemoryReadSnapshot(t, work)) {
				t.Fatal("same priority changed Resource version or native state")
			}
			note := issue(call("note", scope+"beads/work", "Progress", "continues"))
			after := read()
			if note.ID != after.Properties.ID || note.Notes != "Progress continues" || after.Properties.Notes != note.Notes || after.Revision == current.Revision {
				t.Fatalf("note did not append once: %+v %+v", note, after)
			}
			call("note", scope+"beads/work", "Second")
			if got := read(); got.Properties.Notes != "Progress continues\nSecond" {
				t.Fatalf("second note overwrote prior text: %q", got.Properties.Notes)
			}
			preAssign := read()
			assigned := issue(call("assign", scope+"beads/work", "alice", "--if-revision", "="+preAssign.Revision, "--actor", "alice"))
			if assigned.ID != preAssign.Properties.ID || assigned.Assignee != "alice" || read().Revision == preAssign.Revision {
				t.Fatalf("assign lost canonical identity or failed to version: %+v", assigned)
			}
			refuse("revision_conflict", "assign", scope+"beads/work", "bob", "--if-revision", "="+preAssign.Revision)
			state = graphMemoryReadSnapshot(t, work)
			call("assign", scope+"beads/work", "alice")
			if !reflect.DeepEqual(state, graphMemoryReadSnapshot(t, work)) {
				t.Fatal("same assignee changed native state")
			}
			claimed := graphMixedResult[graphstore.IssueMutationResult](t, call("update", scope+"beads/work", "--claim", "--actor", "alice"))
			if !claimed.Changed || claimed.Issue.Properties.Assignee != "alice" {
				t.Fatalf("claim setup failed: %+v", claimed)
			}
			refuse("constraint_violation", "assign", scope+"beads/work", "bob", "--actor", "bob")
			transferred := issue(call("assign", scope+"beads/work", "bob", "--actor", "bob", "--force"))
			if transferred.Assignee != "bob" || read().Properties.Assignee != "bob" {
				t.Fatalf("explicit forced transfer failed: %+v", transferred)
			}
			refuse("invalid_properties", "priority", scope+"beads/work", "P9")
			refuse("permission_denied", "assign", scope+"beads/work", "alice", "--readonly")
			refuse("permission_denied", "priority", scope+"beads/work", "2", "--readonly")
			refuse("permission_denied", "note", scope+"beads/work", "No", "--readonly")
			refuse("permission_denied", "tag", scope+"beads/work", "no", "--readonly")
			refuse("invalid_properties", "tag", scope+"beads/work", " ")
			refuse("capability_unavailable", "priority", scope+"beads/context", "2")
			refuse("capability_unavailable", "note", scope+"beads/context", "No")
			refuse("capability_unavailable", "assign", scope+"beads/context", "alice")
			refuse("capability_unavailable", "tag", scope+"beads/context", "x")

			custom := graphMixedResult[graphstore.IssueRecord](t, call("create", "Custom path", "--graph-path", "beads/custom-path"))
			if custom.Properties.ID == "custom-path" || custom.ID != scope+custom.Properties.ID {
				t.Fatal("fixture needs a native ID distinct from its graph path")
			}
			if got := issue(call("priority", custom.Properties.ID, "1")); got.ID != custom.Properties.ID || got.Priority != 1 {
				t.Fatalf("native-ID priority selected the wrong Issue: %+v", got)
			}
			if got := issue(call("note", custom.Properties.ID, "Native ID note")); got.ID != custom.Properties.ID || got.Notes != "Native ID note" {
				t.Fatalf("native-ID note selected the wrong Issue: %+v", got)
			}
			if got := issue(call("assign", custom.Properties.ID, "carol")); got.ID != custom.Properties.ID || got.Assignee != "carol" {
				t.Fatalf("native-ID assign selected the wrong Issue: %+v", got)
			}
			if got := issue(call("tag", custom.Properties.ID, "native-label")); got.ID != custom.Properties.ID || !reflect.DeepEqual(got.Labels, []string{"native-label"}) {
				t.Fatalf("native-ID tag selected the wrong Issue: %+v", got)
			}
			parts := strings.SplitN(custom.Properties.ID, "-", 2)
			if len(parts) != 2 {
				t.Fatalf("fixture needs a configured Issue prefix: %q", custom.Properties.ID)
			}
			abbreviable := graphMixedResult[graphstore.IssueRecord](t, call("create", "Abbreviable", "--id", parts[0]+"-a3f8e9"))
			partialID := "a3f8"
			if got := issue(call("priority", partialID, "2")); got.ID != abbreviable.Properties.ID || got.Priority != 2 {
				t.Fatalf("abbreviated priority selected the wrong Issue: %+v", got)
			}
			if got := issue(call("note", partialID, "Partial ID note")); got.ID != abbreviable.Properties.ID || got.Notes != "Partial ID note" {
				t.Fatalf("abbreviated note selected the wrong Issue: %+v", got)
			}
			if got := issue(call("assign", partialID, "dana")); got.ID != abbreviable.Properties.ID || got.Assignee != "dana" {
				t.Fatalf("abbreviated assign selected the wrong Issue: %+v", got)
			}
			var stateResult struct {
				IssueID string `json:"issue_id"`
			}
			if err := json.Unmarshal([]byte(call("state", "list", partialID)), &stateResult); err != nil || stateResult.IssueID != abbreviable.Properties.ID {
				t.Fatalf("abbreviated state selected the wrong Issue: %+v %v", stateResult, err)
			}
			call("remember", "Colliding Memory", "--id", "beads/"+custom.Properties.ID)
			memoryBefore := call("show", scope+"beads/"+custom.Properties.ID)
			if got := issue(call("priority", custom.Properties.ID, "0")); got.ID != custom.Properties.ID || got.Priority != 0 {
				t.Fatalf("Memory path shadowed public Issue ID: %+v", got)
			}
			issue(call("note", custom.Properties.ID, "Public ID note"))
			issue(call("assign", custom.Properties.ID, "dana"))
			issue(call("tag", custom.Properties.ID, "public-label"))
			if call("show", scope+"beads/"+custom.Properties.ID) != memoryBefore {
				t.Fatal("public Issue actions changed colliding Memory path")
			}
			if got := issue(call("priority", custom.ID, "0")); got.Priority != 0 {
				t.Fatalf("canonical identity did not select custom Issue: %+v", got)
			}
		})
	}
}
