//go:build cgo

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func TestListASTInstalledWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("requires reserved embedded/shared-server List AST qualification")
	}
	isolateBeadsDirForTest(t)
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			const scope = "https://example.invalid/list-ast/"
			work, home := graphListWorkspace(t, bd, engine, scope, "last")
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			issue := graphMixedResult[graphstore.IssueRecord](t, call("create", "Shared Work", "--id", "i1", "--type", "task", "--force"))
			memory := graphMixedResult[graphstore.Record](t, call("create", "Shared Plan", "--id", "m1", "--type", "memory", "--body", "safe rails"))
			call("create", "Other", "--id", "m2", "--type", "memory")
			link := graphMixedResult[graphstore.LinkMutationResult](t, call("link", "create", issue.ID, memory.ID, "--id", "a", "--type", "references", "--properties", `{"title":"Evidence"}`))
			binding := map[string]any{"op": "binding", "name": "item"}
			contains := func(field, text string) map[string]any {
				return map[string]any{"op": "contains", "text": map[string]any{"op": "field", "of": binding, "path": []string{"properties", field}}, "substring": map[string]any{"op": "literal", "value": text}}
			}
			ast := func(collection string, predicate map[string]any) string {
				return queryASTWorkflowRequest(t, "list", "resources", map[string]any{"op": "filter", "input": map[string]any{"op": "scan", "collection": collection}, "bind": "item", "predicate": predicate})
			}
			read := func(input, stdin string, args ...string) []json.RawMessage {
				t.Helper()
				args = append(args, "--ast", input, "--output-format", "json", "--unpaged", "--omit", "envelope")
				out, diagnostic, exit := queryASTWorkflowRun(t, bd, work, home, stdin, args...)
				var page struct {
					Items []json.RawMessage `json:"items"`
					Next  *string           `json:"next"`
				}
				if exit != 0 || diagnostic != "" || json.Unmarshal([]byte(out), &page) != nil || page.Items == nil || page.Next != nil {
					t.Fatalf("List AST execution: exit%d stderr%s out%s", exit, diagnostic, out)
				}
				return page.Items
			}
			ids := func(rows []json.RawMessage) []string {
				t.Helper()
				result := make([]string, 0, len(rows))
				for _, row := range rows {
					var record struct {
						ID string `json:"id"`
					}
					if err := json.Unmarshal(row, &record); err != nil {
						t.Fatal(err)
					}
					result = append(result, record.ID)
				}
				return result
			}
			shared := ast("beads", contains("title", "Shared"))
			if got := ids(read(shared, "", "list")); len(got) != 1 || got[0] != issue.ID {
				t.Fatalf("default Issue population changed: %v", got)
			}
			if got := ids(read(shared, "", "list", "--type", "any")); len(got) != 2 {
				t.Fatalf("all-Type predicate intersection: %v", got)
			}
			if got := ids(read(shared, "", "list", "--type", "memory")); len(got) != 1 || got[0] != memory.ID {
				t.Fatalf("Memory family predicate intersection: %v", got)
			}
			path := filepath.Join(work, "list-ast.json")
			body := ast("beads", contains("body", "rails"))
			if err := os.WriteFile(path, []byte(body), 0600); err != nil {
				t.Fatal(err)
			}
			if got := ids(read("@"+path, "", "list", "--type", "memory")); len(got) != 1 || got[0] != memory.ID {
				t.Fatalf("file body predicate: %v", got)
			}
			if got := ids(read("-", body, "list", "--type", "memory", "--include", "/properties/title")); len(got) != 1 || got[0] != memory.ID {
				t.Fatalf("stdin/projection predicate: %v", got)
			}
			links := ast("links", contains("title", "Evidence"))
			if got := ids(read(links, "", "link", "list", "--from", issue.ID)); len(got) != 1 || got[0] != link.Link.ID {
				t.Fatalf("Link local predicate/endpoint intersection: %v", got)
			}
			if got := read(links, "", "link", "list", "--from", memory.ID); len(got) != 0 {
				t.Fatalf("Link endpoint filter was ignored: %s", got)
			}
			for _, args := range [][]string{
				{"list", "--ast", links},
				{"list", "--ast", shared, "--ast", shared},
				{"list", "--ast", shared, "--status", "open"},
				{"list", "--type", "any", "--ast", shared, "--page-size", "1"},
			} {
				out, _, exit := queryASTWorkflowRun(t, bd, work, home, "", append(args, "--output-format", "json")...)
				if exit == 0 || strings.Contains(out, `"items"`) {
					t.Fatalf("List refusal returned partial membership: exit%d %s", exit, out)
				}
			}
		})
	}
}
