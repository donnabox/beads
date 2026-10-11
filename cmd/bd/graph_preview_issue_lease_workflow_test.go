//go:build cgo

package main

import (
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
			refuse := func(code string, args ...string) {
				t.Helper()
				graphPolicyCLI(t, bd, work, home, nil, code, append(args, "--json")...)
			}
			call(args...)
			call("create", "Leased work", "--id", "beads/work")
			claim := graphMixedResult[graphstore.IssueMutationResult](t, call("update", "work", "--claim", "--actor", "holder"))
			before := graphMemoryReadSnapshot(t, work)
			refuse("constraint_violation", "heartbeat", "work", "--actor", "stranger")
			refuse("permission_denied", "heartbeat", "work", "--actor", "holder", "--readonly")
			refuse("invalid_selector", "heartbeat", "links/work", "--actor", "holder")
			if !reflect.DeepEqual(before, graphMemoryReadSnapshot(t, work)) {
				t.Fatal("heartbeat refusal changed graph state")
			}
			call("heartbeat", "work", "--actor", "holder")
			current := graphMixedResult[graphstore.IssueRecord](t, call("show", "work", "--format", "graph-json"))
			if current.Revision != claim.Issue.Revision || current.Properties.LeaseExpiresAt == nil || current.Properties.HeartbeatAt == nil {
				t.Fatalf("heartbeat changed retained Issue or lost lease: %+v", current)
			}
			call("show", "work", "--version", claim.Issue.Revision)
			refuse("invalid_properties", "reclaim", "--older-than=-1s")
			refuse("invalid_properties", "reclaim", "--id=")
			refuse("permission_denied", "reclaim", "--readonly")
			live := graphMemoryReadSnapshot(t, work)
			call("reclaim", "--older-than=0s", "--id=work", "--actor", "reaper")
			if !reflect.DeepEqual(live, graphMemoryReadSnapshot(t, work)) {
				t.Fatal("live-lease reclaim changed graph state")
			}
			call("create", "Stateful work", "--id", "beads/stateful", "--labels", "mode:degraded")
			initialStatus := graphMixedResult[map[string]any](t, call("status"))
			if summary, ok := initialStatus["summary"].(map[string]any); !ok || summary["total_issues"] != float64(2) {
				t.Fatalf("bare status did not count native Issues: %+v", initialStatus)
			}
			stateBefore := graphMemoryReadSnapshot(t, work)
			refuse("capability_unavailable", "status", "--assigned")
			refuse("capability_unavailable", "set-state", "stateful", "mode=healthy")
			if !reflect.DeepEqual(stateBefore, graphMemoryReadSnapshot(t, work)) {
				t.Fatal("state refusal changed graph state")
			}
			read := graphMixedResult[map[string]any](t, call("state", "stateful", "mode"))
			if read["value"] != "degraded" {
				t.Fatalf("state read: %+v", read)
			}
			listed := graphMixedResult[map[string]any](t, call("state", "list", "stateful"))
			if states, ok := listed["states"].(map[string]any); !ok || states["mode"] != "degraded" {
				t.Fatalf("state list: %+v", listed)
			}
		})
	}
}
