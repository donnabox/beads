//go:build cgo

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/configfile"
	"github.com/steveyegge/beads/internal/httpapi/bdpwire"
	"github.com/steveyegge/beads/internal/httpapi/graphread"
	"github.com/steveyegge/beads/internal/resourcequery"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// Reuse normal CLI initialization/authorship and the existing reserved workflow
// binary/environment, then the actual checked Reader. This is not a new engine
// harness, HTTP activation, discovery/ownership proof or concurrent-write test.
func TestCanonicalQueryAndGraphCheckedLocalWorkflow(t *testing.T) {
	if os.Getenv("BEADS_TEST_EMBEDDED_DOLT") != "1" {
		t.Skip("set BEADS_TEST_EMBEDDED_DOLT=1 for embedded Query/Graph local qualification")
	}
	isolateBeadsDirForTest(t)
	bd := buildBDUnderTest(t)
	work, home := t.TempDir(), t.TempDir()
	const scope = "https://example.invalid/invocation-local/"
	graphPolicyCLI(t, bd, work, home, nil, "", "init", "--graph-mode", "link", "--graph-profile", "types", "--scope-url", scope, "--prefix", "invoke", "--skip-hooks", "--skip-agents", "--non-interactive", "--json")
	graphPolicyCLI(t, bd, work, home, nil, "", "create", "Note", "--id", "note", "--type", "memory", "--json")
	cfg, err := configfile.LoadForDiscovery(filepath.Join(work, ".beads"))
	if err != nil || cfg == nil || cfg.GraphSchemaVersion != graphstore.TypesFreshSchemaVersion || cfg.GraphScopeURL != scope {
		t.Fatalf("normal Types14 binding absent: %+v %v", cfg, err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
	defer cancel()
	store, err := graphstore.OpenExisting(ctx, graphstore.Options{
		Backend: cfg.DoltMode, DataDir: filepath.Join(cfg.GraphWorkspace, "embeddeddolt"), Database: cfg.DoltDatabase, Branch: "main",
		Binding: graphstore.Binding{WorkspaceID: cfg.GraphWorkspace, ScopeURL: cfg.GraphScopeURL, AuthorityID: cfg.GraphAuthorityID, SchemaVersion: cfg.GraphSchemaVersion},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	})
	reader := graphread.New(store)
	adapter, err := newCanonicalLocalQueryAdapter(reader)
	if err != nil {
		t.Fatal(err)
	}
	invocation, err := admitCanonicalQueryInvocation([]byte(queryASTCountFixture))
	if err != nil {
		t.Fatal(err)
	}
	query, err := invokeCanonicalQueryLocal(ctx, scope, invocation, adapter)
	if err != nil {
		t.Fatal(err)
	}
	oracle, err := reader.ExecuteResourceQuery(ctx, invocation.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	oracleBytes, err := json.Marshal(oracle)
	if err != nil || !bytes.Equal(bytes.TrimSpace(query.Bytes()), oracleBytes) {
		t.Fatalf("shared local Query changed actual Reader body: %s/%s %v", query.Bytes(), oracleBytes, err)
	}
	graph, err := invokeCanonicalGraphLocal(ctx, scope, reader, canonicalGraphRequestOptions{Roots: []bdpwire.Reference{{URI: scope + "note"}}, Direction: "outgoing", Depth: json.RawMessage(`0`)})
	if err != nil {
		t.Fatal(err)
	}
	if graph.ScopeURL() != scope || graph.StatusCode() != 200 || graph.Headers() != nil || graph.ResultForm() != "graph" || query.ResultForm() != "count" {
		t.Fatal("lost local owning Scope/form or invented HTTP headers")
	}
	var actual struct {
		Result     resourcequery.GraphResult
		Evaluation resourcequery.Evaluation
	}
	if err := json.Unmarshal(graph.Bytes(), &actual); err != nil || len(actual.Result.Roots) != 1 || actual.Result.Roots[0].ID != scope+"note" || len(actual.Result.Beads) != 1 || len(actual.Result.Links) != 0 {
		t.Fatalf("Graph adapter did not execute actual checked Reader: %s %v", graph.Bytes(), err)
	}
	if wrong, err := invokeCanonicalQueryLocal(ctx, "https://other.example/", invocation, adapter); !errors.Is(err, resourcequery.ErrDataset) || wrong.StatusCode() != 0 {
		t.Fatalf("local Scope mismatch executed/fell back: %+v %v", wrong, err)
	}
	// The adapter borrows Store. Both Query and Graph succeeded through it;
	// now finish ordinary cleanup before inspecting/delivering detached data.
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	if refused, err := invokeCanonicalQueryLocal(ctx, scope, invocation, adapter); err == nil || refused.StatusCode() != 0 {
		t.Fatalf("closed borrowed store was reopened or yielded success: %+v %v", refused, err)
	}
	var local struct{ Result resourcequery.CountResult }
	if err := json.Unmarshal(query.Value(), &local); err != nil {
		t.Fatal(err)
	}
	if local.Result.Count < 1 {
		t.Fatal("actual complete native capture has no authored record")
	}
	// A normal fresh CLI process exercises the narrow runQueryAST function
	// delta and its untouched close-before-deliver/error/output wrapper.
	cli := queryASTWorkflowRead(t, bd, work, home, "", "aggregate", "query", "--ast", queryASTCountFixture)
	var count resourcequery.CountResult
	if err := json.Unmarshal(cli.Result, &count); err != nil || count != local.Result {
		t.Fatalf("registered Query AST path changed: %s %v", cli.Result, err)
	}
}
