//go:build cgo

package main

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
)

// A graph Issue's opaque revision is a complete-record guard, including its
// owned Link projection. Guarded close uses the native single-Issue writer and
// refuses stale retries before the already-closed no-op can mask a conflict.
func TestGraphPreviewCloseGuardWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			const scope = "https://example.invalid/close-guard/"
			initArgs := []string{"init", "--graph-mode", "link", "--scope-url", scope,
				"--skip-hooks", "--skip-agents", "--non-interactive"}
			if engine == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for ordinary shared-server qualification")
				}
				initArgs = append(initArgs, "--server", "--external", "--server-host", "127.0.0.1",
					"--server-port", port, "--server-user", "root")
			}
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			refuse := func(wantCode int, want string, args ...string) {
				t.Helper()
				stdout, stderr, code := graphVersionsProcess(t, bd, work, home, append(args, "--json")...)
				if code != wantCode || stdout != "" || !strings.Contains(stderr, want) {
					t.Fatalf("bd %v: exit=%d stdout=%q stderr=%q; want exit=%d and %q", args, code, stdout, stderr, wantCode, want)
				}
			}
			call(initArgs...)
			call("create", "First", "--graph-path", "beads/first")
			call("create", "Second", "--graph-path", "beads/second")
			before := graphMixedResult[graphstore.IssueRecord](t, call("show", "--format", "graph-json", scope+"beads/first"))
			versionsBefore := call("versions", "first")
			refuse(4, "revision_conflict", "close", scope+"beads/first", "--if-revision", "stale", "--reason", "done")
			refuse(2, "invalid_selector", "close", scope+"beads/first", scope+"beads/second", "--if-revision", before.Revision)
			refuse(2, "invalid_selector", "close", scope+"beads/first", "--if-revision", before.Revision, "--claim-next")
			refuse(2, "invalid_selector", "close", scope+"beads/first", "--if-revision", before.Revision, "--suggest-next")
			if got := call("versions", "first"); got != versionsBefore {
				t.Fatal("a refused guard or combination changed the Issue")
			}
			closed := graphMixedResult[graphstore.IssueMutationResult](t,
				call("close", scope+"beads/first", "--if-revision", before.Revision, "--reason", "done"))
			if !closed.Changed || closed.Issue.Properties.Status != types.StatusClosed ||
				closed.Issue.Revision == before.Revision || closed.Issue.Properties.CloseReason != "done" {
				t.Fatalf("guarded close did not use the native Issue lifecycle: %+v", closed)
			}
			versionsClosed := call("versions", "first")
			if versionsClosed == versionsBefore {
				t.Fatal("guarded close did not retain a successor")
			}
			refuse(4, "revision_conflict", "close", scope+"beads/first", "--if-revision", before.Revision, "--reason", "retry")
			retry := graphMixedResult[graphstore.IssueMutationResult](t,
				call("close", scope+"beads/first", "--if-revision", closed.Issue.Revision, "--reason", "retry"))
			if retry.Changed || !reflect.DeepEqual(retry.Issue, closed.Issue) || call("versions", "first") != versionsClosed {
				t.Fatalf("accepted already-closed retry minted another version or changed the first reason: %+v", retry)
			}
		})
	}
}

// A complete public URI names the reserved Issue allocation, even when its
// path-looking ID is not the private graph locator accepted by the writer.
// The same selection must reach both single and atomic batch close.
func TestGraphPreviewCloseCanonicalPublicIssue(t *testing.T) {
	isolateBeadsDirForTest(t)
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			const scope = "https://example.invalid/close-public-id/"
			work, home := graphListWorkspace(t, bd, engine, scope)
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			create := func(id string) graphstore.IssueRecord {
				t.Helper()
				return graphMixedResult[graphstore.IssueRecord](t, call("create", id, "--id", id, "--force"))
			}
			first, second, third := create("beads/first"), create("beads/second"), create("beads/third")
			missing := scope + "beads/missing"
			graphPolicyCLI(t, bd, work, home, nil, "not_found", "close", missing, "--reason", "done", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "not_found", "close", first.ID, missing, "--reason", "done", "--json")
			unchanged := graphMixedResult[graphstore.IssueRecord](t, call("show", first.ID, "--format", "graph-json"))
			if unchanged.Properties.Status == types.StatusClosed {
				t.Fatalf("failed batch close changed its valid Issue: %+v", unchanged)
			}
			closed := graphMixedResult[graphstore.IssueMutationResult](t, call("close", first.ID, "--reason", "done"))
			if !closed.Changed || closed.Issue.ID != first.ID || closed.Issue.Properties.Status != types.StatusClosed {
				t.Fatalf("single canonical close selected the wrong Issue: %+v", closed)
			}
			batch := graphMixedResult[[]graphstore.IssueMutationResult](t, call("close", second.ID, third.ID, "--reason", "done"))
			if len(batch) != 2 || !batch[0].Changed || !batch[1].Changed || batch[0].Issue.ID != second.ID || batch[1].Issue.ID != third.ID {
				t.Fatalf("batch canonical close lost exact public selections: %+v", batch)
			}
		})
	}
}

func TestGraphPreviewReopenCanonicalBatchPartial(t *testing.T) {
	isolateBeadsDirForTest(t)
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			const scope = "https://example.invalid/reopen-public-batch/"
			work, home := graphListWorkspace(t, bd, engine, scope)
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			created := graphMixedResult[graphstore.IssueRecord](t, call("create", "First", "--id", "beads/first", "--force"))
			call("close", created.ID, "--reason", "done")
			stdout, stderr, exit := graphGenerationRun(t, bd, work, home, "reopen", created.ID, scope+"beads/missing", "--json")
			if exit != 1 || !strings.Contains(stderr, "not_found") {
				t.Fatalf("batch reopen lost partial-failure status: exit=%d stdout=%s stderr=%s", exit, stdout, stderr)
			}
			results := graphMixedResult[[]graphstore.IssueMutationResult](t, stdout)
			if len(results) != 1 || !results[0].Changed || results[0].Issue.ID != created.ID {
				t.Fatalf("batch reopen lost successful sibling: %+v", results)
			}
			shown := graphMixedResult[graphstore.IssueRecord](t, call("show", created.ID, "--format", "graph-json"))
			if shown.Properties.Status == types.StatusClosed {
				t.Fatalf("successful reopen did not persist: %+v", shown)
			}
		})
	}
}
