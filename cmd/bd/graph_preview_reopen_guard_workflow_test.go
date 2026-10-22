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

func TestGraphPreviewReopenGuardWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			initArgs := []string{"init", "--graph-mode", "link", "--scope-url", "https://example.invalid/reopen-guard/", "--skip-hooks", "--skip-agents", "--non-interactive"}
			if engine == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for ordinary shared-server qualification")
				}
				initArgs = append(initArgs, "--server", "--external", "--server-host", "127.0.0.1", "--server-port", port, "--server-user", "root")
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
			call("create", "Subject", "--id", "subject")
			call("create", "Other", "--id", "other")
			original := graphMixedResult[graphstore.IssueRecord](t, call("show", "subject", "--format", "graph-json"))
			closed := graphMixedResult[graphstore.IssueMutationResult](t, call("close", "subject", "--reason", "done"))
			versionsClosed := call("versions", "subject")
			refuse(4, "revision_conflict", "reopen", "subject", "--if-revision", original.Revision)
			refuse(2, "invalid_selector", "reopen", "subject", "--if-revision=")
			refuse(2, "invalid_selector", "reopen", "subject", "other", "--if-revision", closed.Issue.Revision)
			if call("versions", "subject") != versionsClosed {
				t.Fatal("refused guard changed Issue versions")
			}
			opened := graphMixedResult[graphstore.IssueMutationResult](t, call("reopen", "subject", "--if-revision", closed.Issue.Revision, "--reason", "retry"))
			if !opened.Changed || opened.Issue.Properties.Status != types.StatusOpen || opened.Issue.Revision == closed.Issue.Revision {
				t.Fatalf("matching guard did not reopen Issue: %+v", opened)
			}
			versionsOpen := call("versions", "subject")
			refuse(4, "revision_conflict", "reopen", "subject", "--if-revision", closed.Issue.Revision)
			noop := graphMixedResult[graphstore.IssueMutationResult](t, call("reopen", "subject", "--if-revision", opened.Issue.Revision))
			if noop.Changed || !reflect.DeepEqual(noop.Issue, opened.Issue) || call("versions", "subject") != versionsOpen {
				t.Fatalf("guarded already-open retry changed Issue: %+v", noop)
			}
		})
	}
}
