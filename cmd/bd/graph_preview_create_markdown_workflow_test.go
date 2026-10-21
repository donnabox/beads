//go:build cgo

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
)

// A markdown file is one native Issue batch and one graph catalog transaction.
// Both installed backends must expose every created heading, while a file with
// an unrepresentable dependency must leave no earlier heading behind.
func TestGraphPreviewCreateMarkdownBatchWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			args := []string{"init", "--graph-mode", "link", "--prefix", "md", "--scope-url", "https://example.invalid/markdown/", "--skip-hooks", "--skip-agents", "--non-interactive"}
			if engine == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for ordinary shared-server qualification")
				}
				args = append(args, "--server", "--external", "--server-host", "127.0.0.1", "--server-port", port, "--server-user", "root")
			}
			call := func(code string, args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, code, append(args, "--json")...)
			}
			call("", args...)
			plan := filepath.Join(work, "issues.md")
			if err := os.WriteFile(plan, []byte("## First heading\n\n### Description\nFirst body.\n\n### Priority\n1\n\n### Type\nbug\n\n## Second heading\n\n### Description\nSecond body.\n"), 0600); err != nil {
				t.Fatal(err)
			}
			var issues []*types.Issue
			raw := call("", "create", "--file", plan)
			if err := json.Unmarshal([]byte(raw), &issues); err != nil {
				t.Fatalf("markdown --json must be the ordinary bare Issue array: %v\n%s", err, raw)
			}
			if len(issues) != 2 || issues[0].Title != "First heading" || issues[0].Priority != 1 || issues[0].Description != "First body." || issues[1].Title != "Second heading" {
				t.Fatalf("markdown batch fields: %+v", issues)
			}
			for _, issue := range issues {
				shown := graphMixedResult[graphstore.IssueRecord](t, call("", "show", issue.ID, "--format", "graph-json"))
				if shown.Properties.Title != issue.Title || shown.Properties.ID == issue.ID || shown.ID != "https://example.invalid/markdown/"+issue.ID {
					t.Fatalf("markdown Issue mapping differs: created=%+v shown=%+v", issue, shown)
				}
			}
			prereq := graphMixedResult[graphstore.IssueRecord](t, call("", "create", "Prerequisite", "--id", "md-prereq"))
			if err := os.WriteFile(plan, []byte("## Plain batch item\n\n## Dependent batch item\n\n### Dependencies\nmd-prereq\n"), 0600); err != nil {
				t.Fatal(err)
			}
			raw = call("", "create", "--file", plan)
			if err := json.Unmarshal([]byte(raw), &issues); err != nil || len(issues) != 2 {
				t.Fatalf("dependency batch Issue array: %v\n%s", err, raw)
			}
			dependent := graphMixedResult[graphstore.IssueRecord](t, call("", "show", issues[1].ID, "--format", "graph-json"))
			if len(dependent.Owned) != 1 {
				t.Fatalf("dependent Issue lacks its canonical Link: %+v", dependent)
			}
			var link graphstore.LinkRecord
			if err := json.Unmarshal(dependent.Owned[0], &link); err != nil {
				t.Fatal(err)
			}
			if link.Source != dependent.ID || link.Target != prereq.ID || link.Properties["dep_type"] != string(types.DepBlocks) {
				t.Fatalf("markdown dependency Link = %+v", link)
			}
			if err := os.WriteFile(plan, []byte("## Must not land\n\n## Has a dependency\n\n### Dependencies\nmissing-issue\n"), 0600); err != nil {
				t.Fatal(err)
			}
			call("not_found", "create", "--file", plan)
			if err := os.WriteFile(plan, []byte("## External must not land\n\n### Dependencies\nexternal:other#1\n"), 0600); err != nil {
				t.Fatal(err)
			}
			call("capability_unavailable", "create", "--file", plan)
			if err := os.WriteFile(plan, []byte("## Typed external must not land\n\n### Dependencies\nblocked-by: external:other#1\n"), 0600); err != nil {
				t.Fatal(err)
			}
			call("capability_unavailable", "create", "--file", plan)
			if err := os.WriteFile(plan, []byte("No level-two Issue headings here.\n"), 0600); err != nil {
				t.Fatal(err)
			}
			call("invalid_properties", "create", "--file", plan)
			if err := os.WriteFile(plan, []byte("## Needs validation\n\n### Type\nbug\n\n### Description\nNo reproduction steps.\n"), 0600); err != nil {
				t.Fatal(err)
			}
			call("invalid_properties", "create", "--file", plan, "--validate")
			page := graphMixedResult[graphstore.IssueListPage](t, call("", "list", "--all", "--format=records-json"))
			if len(page.Items) != 5 {
				t.Fatalf("refused markdown file or lint left a partial Issue: %+v", page.Items)
			}
		})
	}
}
