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
			call("create", "Authentication work", "--id", "beads/work", "--priority", "1", "--labels", "backend", "--metadata", `{"team":"backend"}`)
			call("create", "Other work", "--id", "beads/other", "--priority", "2")
			call("close", "beads/other")
			call("create", "Child work", "--id", "beads/work.child", "--priority", "3", "--type", "bug")

			var found []*types.IssueWithCounts
			if out := call("search", "Authentication"); json.Unmarshal([]byte(out), &found) != nil || len(found) != 1 || found[0].ID != "beads/work" || found[0].Title != "Authentication work" {
				t.Fatalf("search title: %s", out)
			}
			if out := call("search", "work", "--status", "closed"); json.Unmarshal([]byte(out), &found) != nil || len(found) != 1 || found[0].ID != "beads/other" {
				t.Fatalf("search closed filter: %s", out)
			}
			if out := call("search", "work", "--metadata-field", "team=backend", "--has-metadata-key", "team", "--created-after", "2020-01-01"); json.Unmarshal([]byte(out), &found) != nil || len(found) != 1 || found[0].ID != "beads/work" {
				t.Fatalf("search metadata/date filters: %s", out)
			}
			if out := graphPolicyCLI(t, bd, work, home, nil, "", "search", "work", "--metadata-field", "team=backend", "--format", "json"); json.Unmarshal([]byte(out), &found) != nil || len(found) != 1 || found[0].ID != "beads/work" {
				t.Fatalf("search metadata with JSON alias: %s", out)
			}
			if out := call("search", "work", "--created-before", "2020-01-01"); json.Unmarshal([]byte(out), &found) != nil || len(found) != 0 {
				t.Fatalf("search created-before: %s", out)
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
			if out := call("count", "--by-priority"); json.Unmarshal([]byte(out), &grouped) != nil || grouped.Total != 3 || len(grouped.Groups) != 3 || grouped.Groups[0].Group != "P1" || grouped.Groups[1].Group != "P2" || grouped.Groups[2].Group != "P3" {
				t.Fatalf("grouped count: %s", out)
			}
			for _, tc := range []struct {
				name string
				args []string
				want int
			}{
				{"canonical ID set", []string{"--id", "beads/work,beads/other"}, 2},
				{"bare ID", []string{"--id", "work"}, 1},
				{"parent dotted fallback", []string{"--parent", "beads/work"}, 1},
				{"edge-less dotted child has no parent", []string{"--no-parent"}, 3},
				{"exclude status", []string{"--exclude-status", "closed"}, 2},
				{"exclude type", []string{"--exclude-type", "bug"}, 2},
				{"ID and exclusion", []string{"--id", "beads/work,beads/other", "--exclude-status", "closed"}, 1},
				{"parent and exclusion", []string{"--parent", "beads/work", "--exclude-type", "bug"}, 0},
			} {
				out := call(append([]string{"count"}, tc.args...)...)
				if json.Unmarshal([]byte(out), &count) != nil || count.Count != tc.want {
					t.Fatalf("count %s: %s", tc.name, out)
				}
			}
			if out := call("query", "title=Authentication"); json.Unmarshal([]byte(out), &found) != nil || len(found) != 1 || found[0].ID != "beads/work" {
				t.Fatalf("query default closed exclusion: %s", out)
			}
			if out := call("query", "id=beads/other OR priority=1", "--all"); json.Unmarshal([]byte(out), &found) != nil || len(found) != 2 {
				t.Fatalf("query canonical ID and OR: %s", out)
			}
			if out := call("query", "parent=beads/work AND type=bug"); json.Unmarshal([]byte(out), &found) != nil || len(found) != 1 || found[0].ID != "beads/work.child" {
				t.Fatalf("query canonical parent: %s", out)
			}
			for _, command := range [][]string{{"search", "work"}, {"count"}, {"query", "type=task"}} {
				out := graphPolicyCLI(t, bd, work, home, nil, "", append(command, "--format", "json")...)
				if !json.Valid([]byte(out)) {
					t.Fatalf("%v --format json did not use the native JSON alias: %s", command, out)
				}
				graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", append(command, "--format", "records-json")...)
			}
			graphPolicyCLI(t, bd, work, home, nil, "invalid_selector", "search", "work", "--metadata-field", "invalid", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "invalid_selector", "search", "work", "--created-after", "not-a-date", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "count", "--include-infra", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "invalid_selector", "count", "--parent", "work", "--no-parent", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "invalid_selector", "count", "--exclude-status", "bogus", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "invalid_selector", "count", "--id", "links/wrong-kind", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "invalid_selector", "count", "--by-status", "--by-type", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "invalid_selector", "query", "parent!=beads/work", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "invalid_selector", "query", "type=task", "--offset", "1", "--sort", "priority", "--json")
		})
	}
}
