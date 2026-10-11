//go:build cgo

package main

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

// Status discovery uses the selected graph workspace and preserves the
// ordinary statuses JSON shape in separate installed CLI processes.
func TestGraphPreviewStatusesWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			args := []string{"init", "--graph-mode", "link", "--scope-url", "https://example.invalid/statuses/", "--skip-hooks", "--skip-agents", "--non-interactive"}
			if engine == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for ordinary shared-server qualification")
				}
				args = append(args, "--server", "--external", "--server-host", "127.0.0.1", "--server-port", port, "--server-user", "root")
			}
			graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			before := graphMemoryReadSnapshot(t, work)
			var result struct {
				BuiltInStatuses []statusInfo `json:"built_in_statuses"`
				CustomStatuses  []struct {
					Name     string `json:"name"`
					Category string `json:"category"`
				} `json:"custom_statuses"`
			}
			output := graphPolicyCLI(t, bd, work, home, nil, "", "statuses", "--json")
			if err := json.Unmarshal([]byte(output), &result); err != nil {
				t.Fatal(err)
			}
			if len(result.BuiltInStatuses) != len(builtInStatuses) || len(result.CustomStatuses) != 0 {
				t.Fatalf("unexpected status inventory: %+v", result)
			}
			for i, item := range result.BuiltInStatuses {
				if item.Name != string(builtInStatuses[i].Status) || item.Category != string(builtInStatuses[i].Category) || item.Description != builtInStatuses[i].Description {
					t.Fatalf("built-in status %d changed: %+v", i, item)
				}
			}
			human := graphPolicyCLI(t, bd, work, home, nil, "", "statuses")
			if !strings.Contains(human, "Built-in statuses:") || !strings.Contains(human, "No custom statuses configured.") {
				t.Fatalf("status inventory missing human details: %q", human)
			}
			graphPolicyCLI(t, bd, work, home, nil, "invalid_selector", "statuses", "unexpected", "--json")
			var again any
			if err := json.Unmarshal([]byte(graphPolicyCLI(t, bd, work, home, nil, "", "statuses", "--json")), &again); err != nil {
				t.Fatal(err)
			}
			var first any
			if err := json.Unmarshal([]byte(output), &first); err != nil || !reflect.DeepEqual(first, again) {
				t.Fatalf("status discovery changed on repeat: %v", err)
			}
			if after := graphMemoryReadSnapshot(t, work); !reflect.DeepEqual(before, after) {
				t.Fatal("status discovery changed graph state")
			}
			graphListConfig(t, work, "status:\n  custom: [\"triaged:active\"]\n")
			customOutput := graphPolicyCLI(t, bd, work, home, nil, "", "statuses", "--json")
			if err := json.Unmarshal([]byte(customOutput), &result); err != nil || len(result.CustomStatuses) != 1 || result.CustomStatuses[0].Name != "triaged" || result.CustomStatuses[0].Category != "active" {
				t.Fatalf("custom status discovery: %s (%v)", customOutput, err)
			}
		})
	}
}
