//go:build cgo

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/resourcequery"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// Reuse the graph workflow's private subprocess environment and normal binary.
// The test never opens a provider or seeds schema behind the CLI's back.
func queryASTWorkflowRun(t *testing.T, bd, work, home, stdin string, args ...string) (string, string, int) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
	defer cancel()
	cmd := graphMemoryReadCommand(ctx, bd, work, home, args...)
	cmd.Stdin = strings.NewReader(stdin)
	var out, diagnostic bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &diagnostic
	err := cmd.Run()
	if ctx.Err() != nil {
		t.Fatalf("Query workflow deadline: %v args=%v stderr=%s", ctx.Err(), args, diagnostic.String())
	}
	if err == nil {
		return out.String(), diagnostic.String(), 0
	}
	var exit *exec.ExitError
	if !errors.As(err, &exit) {
		t.Fatalf("Query workflow could not run CLI: %v", err)
	}
	return out.String(), diagnostic.String(), exit.ExitCode()
}

func queryASTWorkflowRequest(t *testing.T, profile, form string, selection map[string]any) string {
	t.Helper()
	raw, err := json.Marshal(map[string]any{"ir": resourcequery.IR, "profile": profile, "selection": selection, "result": map[string]any{"form": form}})
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

type queryASTWorkflowResponse struct {
	Result     json.RawMessage          `json:"result"`
	Evaluation resourcequery.Evaluation `json:"evaluation"`
}

func queryASTWorkflowRead(t *testing.T, bd, work, home, stdin, kind string, args ...string) queryASTWorkflowResponse {
	t.Helper()
	out, diagnostic, exit := queryASTWorkflowRun(t, bd, work, home, stdin, args...)
	var envelope struct {
		SchemaVersion int                      `json:"schema_version"`
		BDPSpec       string                   `json:"bdp_spec"`
		ResultKind    string                   `json:"result_kind"`
		Outcome       string                   `json:"outcome"`
		Data          queryASTWorkflowResponse `json:"data"`
	}
	if exit != 0 || diagnostic != "" || json.Unmarshal([]byte(out), &envelope) != nil || envelope.SchemaVersion != 2 || envelope.BDPSpec != queryASTContract || envelope.ResultKind != kind || envelope.Outcome != "success" || !envelope.Data.Evaluation.Complete || envelope.Data.Evaluation.Scope != "response" {
		t.Fatalf("Query CLI carrier: exit=%d stdout=%s stderr=%s", exit, out, diagnostic)
	}
	return envelope.Data
}

func queryASTWorkflowIDs(t *testing.T, items []json.RawMessage) []string {
	t.Helper()
	ids := make([]string, 0, len(items))
	for _, item := range items {
		var record struct{ ID, Revision, Type string }
		if err := json.Unmarshal(item, &record); err != nil || record.ID == "" || record.Revision == "" || record.Type == "" {
			t.Fatalf("incomplete selected Resource: %s (%v)", item, err)
		}
		ids = append(ids, record.ID)
	}
	slices.Sort(ids)
	return ids
}

// Three semantic cases share one normally authored fixture per engine. Every
// query is a fresh process, so successful results also exercise store reopen.
func TestQueryASTInstalledWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("requires reserved embedded/shared-server Query workflow qualification")
	}
	isolateBeadsDirForTest(t)
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			const scope = "https://example.invalid/query-ast/"
			work, home := graphListWorkspace(t, bd, engine, scope, "qast")
			call := func(t *testing.T, args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			i1 := graphMixedResult[graphstore.IssueRecord](t, call(t, "create", "Work", "--id", "i1", "--type", "task", "--force"))
			m1 := graphMixedResult[graphstore.Record](t, call(t, "create", "Plan", "--id", "m1", "--type", "memory"))
			i2 := graphMixedResult[graphstore.IssueRecord](t, call(t, "create", "Gate", "--id", "i2", "--type", "task", "--force"))
			m2 := graphMixedResult[graphstore.Record](t, call(t, "create", "Context", "--id", "m2", "--type", "memory"))
			var links []string
			for _, edge := range [][3]string{{i1.ID, m1.ID, "a"}, {m1.ID, i2.ID, "b"}, {i2.ID, m2.ID, "c"}} {
				created := graphMixedResult[graphstore.LinkMutationResult](t, call(t, "link", "create", edge[0], edge[1], "--id", edge[2], "--type", "references"))
				links = append(links, created.Link.ID)
			}
			dep := graphMixedResult[graphstore.DependencyResult](t, call(t, "dep", "add", i1.ID, i2.ID))
			slices.Sort(links)
			binding := func(name string) map[string]any { return map[string]any{"op": "binding", "name": name} }
			family := func(name, typ string) map[string]any {
				return map[string]any{"op": "typeIs", "of": binding(name), "type": scope + typ, "match": "family"}
			}
			refs := map[string]any{"op": "refs", "collection": "beads", "references": []string{i1.ID}}
			walk := map[string]any{"op": "walk", "roots": refs, "direction": "outgoing", "depth": "unbounded", "bind": "l", "through": family("l", "references")}
			read := func(t *testing.T, ast, kind string) queryASTWorkflowResponse {
				t.Helper()
				return queryASTWorkflowRead(t, bd, work, home, "", kind, "query", "--ast", ast)
			}
			count := func(t *testing.T, collection string, predicate map[string]any) int {
				t.Helper()
				selection := map[string]any{"op": "filter", "input": map[string]any{"op": "scan", "collection": collection}, "bind": "b", "predicate": predicate}
				response := read(t, queryASTWorkflowRequest(t, "list", "count", selection), "aggregate")
				var result resourcequery.CountResult
				wantKind := "bead"
				if collection == "links" {
					wantKind = "link"
				}
				if err := json.Unmarshal(response.Result, &result); err != nil || result.Form != "count" || string(result.ResourceKind) != wantKind {
					t.Fatalf("count carrier: %s (%v)", response.Result, err)
				}
				return result.Count
			}
			var status struct {
				Capabilities map[string]bool
				Limits       map[string]int
			}
			if raw := call(t, "status", "--graph"); json.Unmarshal([]byte(raw), &status) != nil || !status.Capabilities["queryAST"] || status.Limits["queryAllocatedResources"] != graphstore.QueryMaxResources || status.Limits["queryRequestBytes"] != resourcequery.DefaultLimits().RequestBytes || status.Limits["queryDatasetBytes"] != 16*1024*1024 {
				t.Fatalf("Query capability/budgets missing from local graph status: %s", raw)
			}
			t.Run("JSON-file-stdin", func(t *testing.T) {
				ast := queryASTWorkflowRequest(t, "query", "resources", walk)
				path := filepath.Join(t.TempDir(), "query.json")
				if err := os.WriteFile(path, []byte(ast), 0o600); err != nil {
					t.Fatal(err)
				}
				var first json.RawMessage
				for _, transport := range []struct{ argument, stdin string }{{ast, ""}, {"@" + path, ""}, {"-", ast}} {
					response := queryASTWorkflowRead(t, bd, work, home, transport.stdin, "resource_collection", "query", "--ast", transport.argument)
					var result resourcequery.ResourcesResult
					if err := json.Unmarshal(response.Result, &result); err != nil || result.Form != "resources" || result.ResourceKind != "link" || result.Next != nil || !slices.Equal(queryASTWorkflowIDs(t, result.Items), links) {
						t.Fatalf("mixed related traversal: %s (%v)", response.Result, err)
					}
					if first != nil && !bytes.Equal(first, response.Result) {
						t.Fatal("transport changed selected persisted records")
					}
					first = response.Result
					if len(response.Evaluation.Traversals) != 1 || response.Evaluation.Traversals[0].Completion != "exhausted" || response.Evaluation.Traversals[0].Closure != "complete" {
						t.Fatalf("closure: %+v", response.Evaluation)
					}
				}
			})
			t.Run("families-local-count-and-graph", func(t *testing.T) {
				if got := count(t, "beads", family("b", "memory")); got != 2 {
					t.Fatalf("Memory family count = %d", got)
				}
				if got := count(t, "beads", family("b", "issue")); got != 2 {
					t.Fatalf("Issue refinement family count = %d", got)
				}
				local := map[string]any{"op": "compare", "operator": "eq", "left": map[string]any{"op": "field", "of": binding("b"), "path": []string{"properties", "title"}}, "right": map[string]any{"op": "literal", "value": "Plan"}}
				if got := count(t, "beads", local); got != 1 {
					t.Fatalf("local title count = %d", got)
				}
				if got := count(t, "links", family("b", "dependency")); got != 1 {
					t.Fatalf("native Dependency family count = %d", got)
				}
				selected := map[string]any{"op": "refs", "collection": "links", "references": []string{dep.Link.ID}}
				dependency := read(t, queryASTWorkflowRequest(t, "query", "resources", selected), "resource_collection")
				var native resourcequery.ResourcesResult
				if err := json.Unmarshal(dependency.Result, &native); err != nil || !slices.Equal(queryASTWorkflowIDs(t, native.Items), []string{dep.Link.ID}) {
					t.Fatalf("native Dependency absent: %s (%v)", dependency.Result, err)
				}
				var storedDependency struct{ Type, Source, Target string }
				if err := json.Unmarshal(native.Items[0], &storedDependency); err != nil || storedDependency.Type != dep.Link.Type || storedDependency.Source != i1.ID || storedDependency.Target != i2.ID {
					t.Fatalf("native Dependency body changed: %s (%v)", native.Items[0], err)
				}
				response := read(t, queryASTWorkflowRequest(t, "graph", "graph", walk), "graph")
				var result resourcequery.GraphResult
				if err := json.Unmarshal(response.Result, &result); err != nil {
					t.Fatal(err)
				}
				wantBeads := []string{i1.ID, m1.ID, i2.ID, m2.ID}
				slices.Sort(wantBeads)
				if result.Form != "graph" || len(result.Roots) != 1 || result.Roots[0].ID != i1.ID || !slices.Equal(queryASTWorkflowIDs(t, result.Beads), wantBeads) || !slices.Equal(queryASTWorkflowIDs(t, result.Links), links) || len(result.Endpoints) != 3 {
					t.Fatalf("mixed graph: %s", response.Result)
				}
				editions := map[resourcequery.Edition]bool{}
				for _, bead := range result.Beads {
					var edition resourcequery.Edition
					if err := json.Unmarshal(bead, &edition); err != nil {
						t.Fatal(err)
					}
					editions[edition] = true
				}
				if !editions[result.Roots[0]] {
					t.Fatal("graph root has no represented edition")
				}
				mapped := map[resourcequery.Edition]bool{}
				for _, mapping := range result.Endpoints {
					if mapped[mapping.Link] {
						t.Fatalf("duplicate endpoint mapping: %+v", mapping)
					}
					mapped[mapping.Link] = true
					if !editions[mapping.Source] || !editions[mapping.Target] {
						t.Fatalf("phantom endpoint: %+v", mapping)
					}
					found := false
					for _, raw := range result.Links {
						var link struct{ ID, Revision, Source, Target string }
						if err := json.Unmarshal(raw, &link); err != nil {
							t.Fatal(err)
						}
						if link.ID == mapping.Link.ID {
							found = link.Revision == mapping.Link.Revision && link.Source == mapping.Source.ID && link.Target == mapping.Target.ID
						}
					}
					if !found {
						t.Fatalf("endpoint mapping rewrote authored Link: %+v", mapping)
					}
				}
			})
			t.Run("refusals-and-close-reopen", func(t *testing.T) {
				graphAST := queryASTWorkflowRequest(t, "graph", "graph", walk)
				before := read(t, graphAST, "graph")
				pin := map[string]any{"op": "refs", "collection": "beads", "references": []any{map[string]any{"uri": i1.ID, "revision": dep.Source.Revision}}}
				missing := map[string]any{"op": "refs", "collection": "beads", "references": []string{scope + "missing"}}
				projection := strings.TrimSuffix(queryASTCountFixture, "}") + `,"projection":{"omit":["/properties"]}}`
				for _, tc := range []struct {
					ast, code string
					exit      int
					raw       bool
				}{
					{queryASTWorkflowRequest(t, "query", "resources", pin), "capability_unavailable", 5, false},
					{projection, "capability_unavailable", 5, true},
					{queryASTWorkflowRequest(t, "query", "resources", missing), "not_found", 3, false},
					{`{"ir":"bdp-query-design-1"}`, "invalid_selector", 2, false},
				} {
					args := []string{"query", "--ast", tc.ast}
					if tc.raw {
						args = append(args, "--omit", "envelope")
					}
					out, diagnostic, exit := queryASTWorkflowRun(t, bd, work, home, "", args...)
					if exit != tc.exit || !strings.HasPrefix(diagnostic, tc.code+": ") || strings.Contains(diagnostic, scope+"missing") {
						t.Fatalf("Query refusal: exit=%d stdout=%s stderr=%s", exit, out, diagnostic)
					}
					if tc.raw {
						if out != "" {
							t.Fatal("raw refusal emitted stdout")
						}
						continue
					}
					var failure struct {
						Outcome string
						Data    json.RawMessage
						Error   struct {
							Origin  string
							Problem struct{ Code string }
						}
					}
					if json.Unmarshal([]byte(out), &failure) != nil || failure.Outcome != "failure" || failure.Data != nil || failure.Error.Origin != "cli" || failure.Error.Problem.Code != tc.code {
						t.Fatalf("canonical refusal: %s", out)
					}
				}
				if after := read(t, graphAST, "graph"); !bytes.Equal(before.Result, after.Result) {
					t.Fatal("failed reads mutated represented graph")
				}
				open := map[string]any{"op": "and", "operands": []any{family("b", "issue"), map[string]any{"op": "compare", "operator": "eq", "left": map[string]any{"op": "field", "of": binding("b"), "path": []string{"properties", "status"}}, "right": map[string]any{"op": "literal", "value": "open"}}}}
				if got := count(t, "beads", open); got != 2 {
					t.Fatalf("initial open count = %d", got)
				}
				call(t, "close", i2.ID)
				if got := count(t, "beads", open); got != 1 {
					t.Fatalf("fresh-process closed count = %d", got)
				}
				call(t, "reopen", i2.ID)
				if got := count(t, "beads", open); got != 2 {
					t.Fatalf("fresh-process reopened count = %d", got)
				}
			})
		})
	}
}

// This control deliberately uses normal init rather than bdInit's schema seed.
func TestQueryASTDependencyModeLegacyWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("requires reserved normal embedded CLI qualification")
	}
	isolateBeadsDirForTest(t)
	bd := buildBDUnderTest(t)
	work, home := t.TempDir(), t.TempDir()
	graphPolicyCLI(t, bd, work, home, nil, "", "init", "--graph-mode", "dependency", "--prefix", "qlegacy", "--skip-hooks", "--skip-agents", "--non-interactive")
	created := parseIssueJSON(t, []byte(graphPolicyCLI(t, bd, work, home, nil, "", "create", "Legacy", "--id", "qlegacy-one", "--json")))
	for attempt := 0; attempt < 2; attempt++ {
		out, diagnostic, exit := queryASTWorkflowRun(t, bd, work, home, "", "query", "id="+created.ID, "--json")
		var items []struct{ ID, Title string }
		if exit != 0 || json.Unmarshal([]byte(out), &items) != nil || len(items) != 1 || items[0].ID != created.ID || items[0].Title != "Legacy" || strings.Contains(diagnostic, "capability_unavailable") {
			t.Fatalf("legacy no-AST query changed: exit=%d stdout=%s stderr=%s", exit, out, diagnostic)
		}
		if attempt == 0 {
			refused, stderr, status := queryASTWorkflowRun(t, bd, work, home, "", "query", "--ast", queryASTCountFixture, "--omit", "envelope")
			if status != 5 || refused != "" || !strings.HasPrefix(stderr, "capability_unavailable: ") {
				t.Fatalf("dependency AST fell through to legacy query: %d %s / %s", status, refused, stderr)
			}
		}
	}
}
