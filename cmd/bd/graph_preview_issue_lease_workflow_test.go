//go:build cgo

package main

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func TestGraphPreviewIssueLeaseCommands(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			const scope = "https://example.invalid/issue-lease/"
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
			nativeJSON := func(args ...string) map[string]any {
				t.Helper()
				var result map[string]any
				if err := json.Unmarshal([]byte(call(args...)), &result); err != nil {
					t.Fatalf("decode native command %v: %v", args, err)
				}
				return result
			}
			refuse := func(code string, args ...string) {
				t.Helper()
				graphPolicyCLI(t, bd, work, home, nil, code, append(args, "--json")...)
			}
			call(args...)
			call("create", "Leased work", "--id", "work", "--force")
			claim := graphMixedResult[graphstore.IssueMutationResult](t, call("update", "work", "--claim", "--actor", "holder"))
			before := graphMemoryReadSnapshot(t, work)
			refuse("constraint_violation", "heartbeat", claim.Issue.ID, "--actor", "stranger")
			refuse("permission_denied", "heartbeat", claim.Issue.ID, "--actor", "holder", "--readonly")
			refuse("not_found", "heartbeat", "links/work", "--actor", "holder")
			if !reflect.DeepEqual(before, graphMemoryReadSnapshot(t, work)) {
				t.Fatal("heartbeat refusal changed graph state")
			}
			call("heartbeat", claim.Issue.ID, "--actor", "holder")
			current := graphMixedResult[graphstore.IssueRecord](t, call("show", claim.Issue.ID, "--format", "graph-json"))
			if current.Revision != claim.Issue.Revision || current.Properties.LeaseExpiresAt == nil || current.Properties.HeartbeatAt == nil {
				t.Fatalf("heartbeat changed retained Issue or lost lease: %+v", current)
			}
			call("show", claim.Issue.ID, "--revision", claim.Issue.Revision)
			refuse("invalid_properties", "reclaim", "--older-than=-1s")
			refuse("invalid_properties", "reclaim", "--id=")
			refuse("permission_denied", "reclaim", "--readonly")
			live := graphMemoryReadSnapshot(t, work)
			call("reclaim", "--older-than=0s", "--id="+claim.Issue.ID, "--actor", "reaper")
			if !reflect.DeepEqual(live, graphMemoryReadSnapshot(t, work)) {
				t.Fatal("live-lease reclaim changed graph state")
			}
			call("create", "Stateful work", "--id", "beads/stateful", "--force", "--labels", "mode:degraded")
			// Bare status and state retain their native JSON shapes.
			initialStatus := nativeJSON("status")
			if summary, ok := initialStatus["summary"].(map[string]any); !ok || summary["total_issues"] != float64(2) {
				t.Fatalf("bare status did not count native Issues: %+v", initialStatus)
			}
			stateBefore := graphMemoryReadSnapshot(t, work)
			refuse("capability_unavailable", "status", "--assigned")
			refuse("capability_unavailable", "set-state", "stateful", "mode=healthy")
			if !reflect.DeepEqual(stateBefore, graphMemoryReadSnapshot(t, work)) {
				t.Fatal("state refusal changed graph state")
			}
			read := nativeJSON("state", scope+"beads/stateful", "mode")
			if read["value"] != "degraded" {
				t.Fatalf("state read: %+v", read)
			}
			listed := nativeJSON("state", "list", scope+"beads/stateful")
			if states, ok := listed["states"].(map[string]any); !ok || states["mode"] != "degraded" {
				t.Fatalf("state list: %+v", listed)
			}
			call("remember", "Ordinary Memory", "--id", "beads/plain-memory")
			memoryBefore := graphMemoryReadSnapshot(t, work)
			for _, tc := range []struct{ selector, code string }{
				{"plain-memory", "capability_unavailable"},
				{"beads/plain-memory", "capability_unavailable"},
				{scope + "beads/plain-memory", "capability_unavailable"},
			} {
				refuse(tc.code, "state", tc.selector, "mode")
				refuse(tc.code, "state", "list", tc.selector)
			}
			if !reflect.DeepEqual(memoryBefore, graphMemoryReadSnapshot(t, work)) {
				t.Fatal("state refusal on Memory changed graph state")
			}

			// A native public ID differs from its private graph path.
			// Every Issue-only command must resolve it through the checked catalog.
			custom := graphMixedResult[graphstore.IssueRecord](t, call("create", "Custom path", "--graph-path", "beads/custom-path", "--labels", "mode:ready"))
			if custom.Properties.ID == "custom-path" || custom.ID != scope+custom.Properties.ID {
				t.Fatalf("fixture did not separate canonical path and native ID: %+v", custom)
			}
			customClaim := graphMixedResult[graphstore.IssueMutationResult](t, call("update", custom.ID, "--claim", "--actor", "holder"))
			customBefore := call("versions", custom.ID)
			call("heartbeat", custom.Properties.ID, "--actor", "holder")
			if got := nativeJSON("state", custom.Properties.ID, "mode"); got["value"] != "ready" || got["issue_id"] != custom.Properties.ID {
				t.Fatalf("native-ID state read selected wrong Issue: %+v", got)
			}
			if got := nativeJSON("state", "list", custom.Properties.ID); got["issue_id"] != custom.Properties.ID {
				t.Fatalf("native-ID state list selected wrong Issue: %+v", got)
			}
			call("reclaim", "--older-than=0s", "--id="+custom.Properties.ID, "--actor", "reaper")
			if got := call("versions", custom.ID); got != customBefore {
				t.Fatal("native-ID heartbeat/state/live-lease reclaim changed retained History")
			}
			if got := graphMixedResult[graphstore.IssueRecord](t, call("show", custom.ID, "--format", "graph-json")); got.Revision != customClaim.Issue.Revision || got.Properties.LeaseExpiresAt == nil {
				t.Fatalf("native-ID commands lost live lease: %+v", got)
			}
			call("remember", "Colliding Memory", "--id", "beads/"+custom.Properties.ID)
			// A different Resource's graph path cannot shadow the opaque public
			// Issue ID, even when its suffix has the same spelling.
			call("heartbeat", custom.Properties.ID, "--actor", "holder")
			call("reclaim", "--older-than=0s", "--id="+custom.Properties.ID, "--actor", "reaper")
			if got := nativeJSON("state", custom.Properties.ID, "mode"); got["value"] != "ready" {
				t.Fatalf("public Issue ID was shadowed by a Memory path: %+v", got)
			}
			call("heartbeat", custom.ID, "--actor", "holder")
			call("reclaim", "--older-than=0s", "--id="+custom.ID, "--actor", "reaper")
			if got := nativeJSON("state", custom.ID, "mode"); got["value"] != "ready" {
				t.Fatalf("canonical identity lost precedence: %+v", got)
			}
		})
	}
}
