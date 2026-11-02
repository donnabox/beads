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
	if os.Getenv("BEADS_TEST_EMBEDDED_DOLT") != "1" && os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT") == "" {
		t.Skip("set BEADS_TEST_EMBEDDED_DOLT=1 or BEADS_GRAPH_TEST_SERVER_PORT for List AST qualification")
	}
	isolateBeadsDirForTest(t)
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			if engine == "embedded" && os.Getenv("BEADS_TEST_EMBEDDED_DOLT") != "1" {
				t.Skip("set BEADS_TEST_EMBEDDED_DOLT=1 for embedded List AST qualification")
			}
			const scope = "https://example.invalid/list-ast/"
			work, home := graphListWorkspace(t, bd, engine, scope, "last")
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			issue := graphMixedResult[graphstore.IssueRecord](t, call("create", "Shared Work", "--id", "i1", "--type", "task", "--force", "--metadata", `{"n":1,"marker":null,"series":[1,2],"payload":{"aa":1,"b":2}}`))
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
			for _, tc := range []struct {
				args  []string
				count int
			}{
				{[]string{"list", "--type", "memory", "--property-field", "title=Shared Plan", "--has-property-key", "body"}, 1},
				{[]string{"list", "--exact-type", "memory", "--property-field", "title=Shared Plan"}, 1},
				{[]string{"list", "--type", "memory", "--property-field", "title=Shared Plan", "--property-field", "title=Other"}, 0},
				{[]string{"list", "--type", "memory", "--ast", shared}, 1},
				{[]string{"link", "list", "--from", issue.ID, "--property-field", "title=Evidence"}, 1},
				{[]string{"link", "list", "--from", memory.ID, "--property-field", "title=Evidence"}, 0},
				{[]string{"list", "--type", "memory", "--page-size", "1", "--limit", "1"}, 2},
			} {
				args := append(tc.args, "--count-only", "--output-format", "json", "--omit", "envelope")
				out, diagnostic, exit := queryASTWorkflowRun(t, bd, work, home, "", args...)
				var counted struct {
					Count int `json:"count"`
				}
				if exit != 0 || diagnostic != "" || json.Unmarshal([]byte(out), &counted) != nil || counted.Count != tc.count {
					t.Fatalf("field/Type/AST/endpoint count intersection: args%v exit%d stderr%s out%s", args, exit, diagnostic, out)
				}
			}
			// Compare the new captured-graph field adapter with the existing
			// SQL-backed metadata selector on the same Issue and engine. Search
			// remains an Issue-only compatibility command; no code there changes.
			for _, field := range []string{"n=1", "n=1.0", "marker=null", "marker=", "series=[1, 2]", "series=[1,2]", `payload={"b": 2, "aa": 1}`} {
				legacyOut := call("search", "Shared", "--metadata-field", field)
				var legacy []graphstore.IssueRecord
				if err := json.Unmarshal([]byte(legacyOut), &legacy); err != nil {
					t.Fatalf("legacy metadata oracle: %s %v", legacyOut, err)
				}
				out, diagnostic, exit := queryASTWorkflowRun(t, bd, work, home, "", "list", "--metadata-field", field, "--count-only", "--output-format", "json", "--omit", "envelope")
				var counted struct {
					Count int `json:"count"`
				}
				if exit != 0 || diagnostic != "" || json.Unmarshal([]byte(out), &counted) != nil || counted.Count != len(legacy) {
					t.Fatalf("metadata compatibility %q: legacy%s count%s stderr%s exit%d", field, legacyOut, out, diagnostic, exit)
				}
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
