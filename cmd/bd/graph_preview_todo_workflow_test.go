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

func TestGraphPreviewTodoIssueWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			args := []string{"init", "--graph-mode", "link", "--scope-url", "https://example.invalid/todo/", "--skip-hooks", "--skip-agents", "--non-interactive"}
			if engine == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for ordinary shared-server qualification")
				}
				args = append(args, "--server", "--external", "--server-host", "127.0.0.1", "--server-port", port, "--server-user", "root")
			}
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			issue := func(raw string) types.Issue {
				t.Helper()
				var result types.Issue
				if err := json.Unmarshal([]byte(raw), &result); err != nil {
					t.Fatalf("decode Issue: %v: %s", err, raw)
				}
				return result
			}
			list := func(args ...string) []*types.Issue {
				t.Helper()
				var result []*types.Issue
				if raw := call(args...); json.Unmarshal([]byte(raw), &result) != nil {
					t.Fatalf("decode TODO list: %s", raw)
				}
				return result
			}
			call(args...)
			created := issue(call("todo", "add", "Write tests", "--priority", "1", "--description", "Keep context"))
			if created.ID == "" || created.IssueType != types.TypeTask || created.Status != types.StatusOpen || created.Priority != 1 || created.Description != "Keep context" {
				t.Fatalf("TODO add lost native task defaults: %+v", created)
			}
			other := graphMixedResult[graphstore.IssueRecord](t, call("create", "A bug", "--graph-path", "beads/bug", "--type", "bug"))
			call("remember", "Memory body", "--id", "memo")
			for _, args := range [][]string{{"todo"}, {"todo", "list"}} {
				got := list(args...)
				if len(got) != 1 || got[0].ID != created.ID {
					t.Fatalf("%v listed %+v; expected the open task only", args, got)
				}
			}
			before := graphMixedResult[graphstore.IssueRecord](t, call("show", created.ID, "--format", "graph-json"))
			_, _, initialVersions, _ := graphVersionsListed(t, call("versions", created.ID))
			graphPolicyCLI(t, bd, work, home, nil, "revision_conflict", "todo", "done", created.ID, "--if-revision", "=stale", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "invalid_selector", "todo", "done", created.ID, other.ID, "--if-revision", "="+before.Revision, "--json")
			graphPolicyCLI(t, bd, work, home, nil, "permission_denied", "todo", "add", "No", "--readonly", "--json")
			if len(initialVersions) != 1 {
				t.Fatalf("TODO creation should retain one native version; got %d", len(initialVersions))
			}
			var done struct {
				Closed []string `json:"closed"`
				Reason string   `json:"reason"`
			}
			if raw := call("todo", "done", created.ID, "--if-revision", "="+before.Revision); json.Unmarshal([]byte(raw), &done) != nil || len(done.Closed) != 1 || done.Closed[0] != created.ID || done.Reason != "Completed" {
				t.Fatalf("guarded TODO completion: %s", raw)
			}
			closed := graphMixedResult[graphstore.IssueRecord](t, call("show", created.ID, "--format", "graph-json"))
			_, _, versions, _ := graphVersionsListed(t, call("versions", created.ID))
			if closed.Properties.Status != types.StatusClosed || closed.Properties.CloseReason != "Completed" || len(versions) != 2 || versions[0].Version != closed.Revision {
				t.Fatalf("TODO close did not retain exactly one native successor: %+v %+v", closed, versions)
			}
			if got := list("todo"); len(got) != 0 {
				t.Fatalf("closed task remained in default list: %+v", got)
			}
			if got := list("todo", "list", "--all"); len(got) != 1 || got[0].ID != created.ID {
				t.Fatalf("--all did not include closed task: %+v", got)
			}
			state := graphMemoryReadSnapshot(t, work)
			closedVersions := call("versions", created.ID)
			call("todo", "done", created.ID, "--reason", "do not replace")
			if got := call("versions", created.ID); got != closedVersions || !reflect.DeepEqual(state, graphMemoryReadSnapshot(t, work)) {
				t.Fatal("repeated TODO completion changed native or retained state")
			}
			call("todo", "done", other.ID, "--reason", "fixed bug")
			if got := graphMixedResult[graphstore.IssueRecord](t, call("show", other.ID, "--format", "graph-json")); got.Properties.Status != types.StatusClosed || got.Properties.CloseReason != "fixed bug" {
				t.Fatalf("TODO done must accept any Issue, as native does: %+v", got)
			}
			if got := list("todo", "list", "--all"); len(got) != 1 || got[0].ID != created.ID {
				t.Fatalf("non-task Issue leaked into task list: %+v", got)
			}
			defaultPriority := issue(call("todo", "add", "Default priority"))
			if defaultPriority.Priority != 2 || defaultPriority.Status != types.StatusOpen || defaultPriority.IssueType != types.TypeTask {
				t.Fatalf("TODO add did not use open task P2 defaults: %+v", defaultPriority)
			}
			if got := list("todo"); len(got) != 1 || got[0].ID != defaultPriority.ID {
				t.Fatalf("bare TODO omitted the default-priority open task: %+v", got)
			}
			memoryBefore := graphMemoryReadSnapshot(t, work)
			out, stderr, code := graphVersionsProcess(t, bd, work, home, "todo", "done", "memo", "--json")
			var nonIssue struct {
				Closed []string `json:"closed"`
				Reason string   `json:"reason"`
			}
			if err := json.Unmarshal([]byte(out), &nonIssue); code != 0 || err != nil || len(nonIssue.Closed) != 0 || nonIssue.Reason != "Completed" || !strings.Contains(stderr, "Error: failed to ") || !strings.Contains(stderr, " memo:") {
				t.Fatalf("non-Issue TODO target: exit=%d stdout=%s stderr=%s decode=%v", code, out, stderr, err)
			}
			if !reflect.DeepEqual(memoryBefore, graphMemoryReadSnapshot(t, work)) {
				t.Fatal("non-Issue TODO target changed graph state")
			}

			// Native todo done tries every target and exits successfully even
			// when an individual ID is malformed or missing. Both valid writes
			// must commit through the native Issue History writer exactly once.
			batch := func(bad string) {
				t.Helper()
				first := issue(call("todo", "add", "First batch task"))
				last := issue(call("todo", "add", "Last batch task"))
				out, stderr, code := graphVersionsProcess(t, bd, work, home, "todo", "done", first.ID, bad, last.ID, "--json")
				if code != 0 || !strings.Contains(stderr, "Error: failed to get issue "+bad+":") {
					t.Fatalf("batch %q: exit=%d stdout=%s stderr=%s", bad, code, out, stderr)
				}
				var result struct {
					Closed []string `json:"closed"`
					Reason string   `json:"reason"`
				}
				if err := json.Unmarshal([]byte(out), &result); err != nil || !reflect.DeepEqual(result.Closed, []string{first.ID, last.ID}) || result.Reason != "Completed" {
					t.Fatalf("batch %q lost successful IDs: %s (%v)", bad, out, err)
				}
				for _, id := range result.Closed {
					record := graphMixedResult[graphstore.IssueRecord](t, call("show", id, "--format", "graph-json"))
					_, _, rows, _ := graphVersionsListed(t, call("versions", id))
					if record.Properties.Status != types.StatusClosed || len(rows) != 2 || rows[0].Version != record.Revision {
						t.Fatalf("batch target %s did not close with one retained successor: %+v %+v", id, record, rows)
					}
				}
			}
			batch("links/not-an-issue")
			batch("beads/missing")
		})
	}
}
