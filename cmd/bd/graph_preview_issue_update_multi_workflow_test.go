//go:build cgo

package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func TestGraphPreviewIssueMultiUpdateWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			const scope = "https://example.invalid/multi-update/"
			initArgs := []string{"init", "--graph-mode", "link", "--scope-url", scope, "--skip-hooks", "--skip-agents", "--non-interactive", "--json"}
			if engine == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for shared-server qualification")
				}
				initArgs = append(initArgs, "--server", "--external", "--server-host", "127.0.0.1", "--server-port", port, "--server-user", "root")
			}
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			call(initArgs...)
			for _, id := range []string{"first", "second", "third"} {
				call("create", id, "--id", "beads/"+id)
			}
			before := graphMixedResult[graphstore.IssueRecord](t, call("show", "first", "--format", "graph-json"))
			multi := func(args ...string) (string, string, int) {
				t.Helper()
				ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
				defer cancel()
				cmd := exec.CommandContext(ctx, bd, append(args, "--json")...)
				cmd.Dir = work
				cmd.Env = []string{
					"PATH=" + os.Getenv("PATH"), "HOME=" + home,
					"TMPDIR=" + os.TempDir(), "TMP=" + os.TempDir(), "TEMP=" + os.TempDir(),
					"XDG_CONFIG_HOME=" + filepath.Join(home, ".config"),
					"XDG_CACHE_HOME=" + filepath.Join(home, ".cache"),
					"GIT_CONFIG_GLOBAL=" + filepath.Join(home, "missing-gitconfig"), "GIT_CONFIG_NOSYSTEM=1",
					"BD_DISABLE_METRICS=1", "BD_DISABLE_EVENT_FLUSH=1", "BD_NON_INTERACTIVE=1",
					"BEADS_DOLT_AUTO_START=0", "DOLT_METRICS_DISABLED=1", "NO_COLOR=1",
				}
				var stdout, stderr bytes.Buffer
				cmd.Stdout, cmd.Stderr = &stdout, &stderr
				err := cmd.Run()
				if ctx.Err() != nil {
					t.Fatalf("multi update timed out: %v", ctx.Err())
				}
				if err == nil {
					return stdout.String(), stderr.String(), 0
				}
				var exit *exec.ExitError
				if !errors.As(err, &exit) {
					t.Fatalf("multi update could not run: %v", err)
				}
				return stdout.String(), stderr.String(), exit.ExitCode()
			}
			stdout, stderr, code := multi("update", "first", "missing", "second", "--priority=0", "--assignee=team")
			if code != 1 || !strings.Contains(stderr, `"code":"not_found"`) || !strings.Contains(stderr, "missing") {
				t.Fatalf("partial update exit=%d stdout=%s stderr=%s", code, stdout, stderr)
			}
			results := graphMixedResult[[]graphstore.IssueMutationResult](t, stdout)
			if len(results) != 2 || results[0].Issue.ID != scope+"beads/first" || results[1].Issue.ID != scope+"beads/second" {
				t.Fatalf("successful IDs lost input order: %+v", results)
			}
			for _, id := range []string{"first", "second"} {
				issue := graphMixedResult[graphstore.IssueRecord](t, call("show", id, "--format", "graph-json"))
				if issue.Properties.Priority != 0 || issue.Properties.Assignee != "team" {
					t.Fatalf("%s was not edited: %+v", id, issue.Properties)
				}
			}
			if issue := graphMixedResult[graphstore.IssueRecord](t, call("show", "third", "--format", "graph-json")); issue.Properties.Priority != 2 || issue.Properties.Assignee != "" {
				t.Fatalf("unselected Issue changed: %+v", issue.Properties)
			}
			if prior := graphMixedResult[graphstore.IssueRecord](t, call("show", "first", "--version", before.Revision)); prior.Properties.Priority != 2 || prior.Properties.Assignee != "" {
				t.Fatalf("successful edit did not retain predecessor: %+v", prior.Properties)
			}
			stdout, stderr, code = multi("update", "first", "second", "--priority=1", "--if-revision", before.Revision)
			if code != 2 || stdout != "" || !strings.Contains(stderr, `"code":"invalid_selector"`) {
				t.Fatalf("shared revision guard was admitted: exit=%d stdout=%s stderr=%s", code, stdout, stderr)
			}
			stdout, stderr, code = multi("update", "first", "second", "--parent=third", "--priority=1")
			if code != 5 || stdout != "" || !strings.Contains(stderr, `"code":"capability_unavailable"`) {
				t.Fatalf("multi reparent was admitted: exit=%d stdout=%s stderr=%s", code, stdout, stderr)
			}
			stdout, stderr, code = multi("update", "first", "second", "--priority=1")
			if code != 0 || stderr != "" || len(graphMixedResult[[]graphstore.IssueMutationResult](t, stdout)) != 2 {
				t.Fatalf("all-good update: exit=%d stdout=%s stderr=%s", code, stdout, stderr)
			}
		})
	}
}
