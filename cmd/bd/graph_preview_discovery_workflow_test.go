//go:build cgo

package main

import (
	"encoding/json"
	"testing"

	"github.com/steveyegge/beads/internal/types"
)

func TestGraphPreviewIssueDiscoveryWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := graphListWorkspace(t, bd, engine, "https://example.invalid/discovery/")
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			call("create", "Authentication work", "--id", "beads/work", "--priority", "1", "--labels", "backend")
			call("create", "Other work", "--id", "beads/other", "--priority", "2")
			call("close", "beads/other")

			var found []*types.IssueWithCounts
			if out := call("search", "beads/work"); json.Unmarshal([]byte(out), &found) != nil || len(found) != 1 || found[0].ID != "beads/work" || found[0].Title != "Authentication work" {
				t.Fatalf("search canonical ID: %s", out)
			}
			if out := call("search", "work", "--status", "closed"); json.Unmarshal([]byte(out), &found) != nil || len(found) != 1 || found[0].ID != "beads/other" {
				t.Fatalf("search closed filter: %s", out)
			}
			if out := call("search", "missing"); json.Unmarshal([]byte(out), &found) != nil || len(found) != 0 || out == "null\n" {
				t.Fatalf("empty search must be a JSON array: %s", out)
			}
			var count struct {
				Count         int `json:"count"`
				SchemaVersion int `json:"schema_version"`
			}
			if out := call("count", "--status", "open", "--label", "backend"); json.Unmarshal([]byte(out), &count) != nil || count.Count != 1 || count.SchemaVersion != JSONSchemaVersion {
				t.Fatalf("filtered count: %s", out)
			}
			var grouped struct {
				Total  int `json:"total"`
				Groups []struct {
					Group string `json:"group"`
					Count int    `json:"count"`
				} `json:"groups"`
			}
			if out := call("count", "--by-priority"); json.Unmarshal([]byte(out), &grouped) != nil || grouped.Total != 2 || len(grouped.Groups) != 2 || grouped.Groups[0].Group != "P1" || grouped.Groups[1].Group != "P2" {
				t.Fatalf("grouped count: %s", out)
			}
			if out := call("query", "title=work"); json.Unmarshal([]byte(out), &found) != nil || len(found) != 1 || found[0].ID != "beads/work" {
				t.Fatalf("query default closed exclusion: %s", out)
			}
			if out := call("query", "id=beads/other OR priority=1", "--all"); json.Unmarshal([]byte(out), &found) != nil || len(found) != 2 {
				t.Fatalf("query canonical ID and OR: %s", out)
			}
			for _, command := range [][]string{{"search", "work"}, {"count"}, {"query", "type=task"}} {
				out := graphPolicyCLI(t, bd, work, home, nil, "", append(command, "--format", "json")...)
				if !json.Valid([]byte(out)) {
					t.Fatalf("%v --format json did not use the native JSON alias: %s", command, out)
				}
				graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", append(command, "--format", "records-json")...)
			}
			graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "search", "work", "--metadata-field", "team=backend", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "search", "work", "--metadata-field", "team=backend", "--format", "json")
			graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "search", "work", "--created-after", "2026-01-01", "--status", "open", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "count", "--include-infra", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "count", "--id", "beads/work", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "invalid_selector", "count", "--by-status", "--by-type", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "query", "parent=beads/work", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "invalid_selector", "query", "type=task", "--offset", "1", "--sort", "priority", "--json")
		})
	}
}
