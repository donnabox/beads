//go:build cgo

package main

import (
	"os"
	"strings"
	"testing"
)

// Installed CLI processes must list every native dependency kind in both
// directions without mistaking the one nominal Link Type for a blocks edge.
func TestGraphPreviewDepListKindsWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			args := []string{"init", "--graph-mode", "link", "--prefix", "demo", "--scope-url", "https://example.invalid/dep-list-kinds/", "--skip-hooks", "--skip-agents", "--non-interactive", "--json"}
			if engine == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for shared-server CLI qualification")
				}
				args = append(args, "--server", "--external", "--server-host", "127.0.0.1", "--server-port", port, "--server-user", "root")
			}
			graphPolicyCLI(t, bd, work, home, nil, "", args...)
			for _, id := range []string{"root", "blocker", "tracked", "related"} {
				graphPolicyCLI(t, bd, work, home, nil, "", "create", id, "--id", "beads/"+id, "--json")
			}
			for _, edge := range []struct{ target, kind string }{
				{"blocker", "blocks"}, {"tracked", "tracks"}, {"related", "related"},
			} {
				graphPolicyCLI(t, bd, work, home, nil, "", "dep", "add", "root", edge.target, "--type", edge.kind, "--json")
			}
			all := graphPolicyCLI(t, bd, work, home, nil, "", "dep", "list", "root")
			for _, edge := range []struct{ target, kind string }{
				{"blocker", "blocks"}, {"tracked", "tracks"}, {"related", "related"},
			} {
				line := "beads/root depends on beads/" + edge.target + " via " + edge.kind
				if !strings.Contains(all, line) {
					t.Fatalf("default listing omitted native %s edge: %q", edge.kind, all)
				}
				for _, tc := range []struct{ anchor, direction string }{
					{"root", "down"}, {edge.target, "up"},
				} {
					got := graphPolicyCLI(t, bd, work, home, nil, "", "dep", "list", tc.anchor, "--direction", tc.direction, "--type", edge.kind)
					if !strings.Contains(got, line) || strings.Count(got, " depends on ") != 1 {
						t.Fatalf("%s %s filter must return only its %s edge: %q", tc.anchor, tc.direction, edge.kind, got)
					}
				}
			}
			if got := graphPolicyCLI(t, bd, work, home, nil, "", "dep", "list", "root", "--type", "workspace-custom"); !strings.Contains(got, "No Dependencies.") {
				t.Fatalf("valid absent custom kind should yield an empty listing: %q", got)
			}
			graphPolicyCLI(t, bd, work, home, nil, "invalid_properties", "dep", "list", "root", "--type", strings.Repeat("x", 33), "--json")
			graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "dep", "list", "root", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "dep", "list", "root", "tracked", "--json")
		})
	}
}
