//go:build cgo

package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
)

// Each call is a separate installed process against a fresh graph workspace.
// A shared Dolt server is supplied by the qualification runner.
func TestGraphPreviewCloseReopenBatchWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			args := []string{"init", "--graph-mode", "link", "--scope-url", "https://example.invalid/reopen-batch/", "--skip-hooks", "--skip-agents", "--non-interactive"}
			if engine == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for ordinary shared-server CLI qualification")
				}
				args = append(args, "--server", "--external", "--server-host", "127.0.0.1", "--server-port", port, "--server-user", "root")
			}
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			call(args...)
			call("create", "First", "--id", "beads/first")
			call("create", "Second", "--id", "beads/second")
			closed := graphMixedResult[[]graphstore.IssueMutationResult](t, call("close", "first", "second", "--reason", "First done", "--reason", "Second done"))
			if len(closed) != 2 || !closed[0].Changed || !closed[1].Changed ||
				closed[0].Issue.Properties.Status != types.StatusClosed || closed[1].Issue.Properties.Status != types.StatusClosed {
				t.Fatalf("batch did not close both Issues in input order: %+v", closed)
			}
			firstBefore, secondBefore := call("versions", "first"), call("versions", "second")

			reopened := graphMixedResult[[]graphstore.IssueMutationResult](t, call("reopen", "first", "second", "--reason", "More work"))
			if len(reopened) != 2 || !reopened[0].Changed || !reopened[1].Changed ||
				reopened[0].Issue.Properties.Status != types.StatusOpen || reopened[1].Issue.Properties.Status != types.StatusOpen {
				t.Fatalf("batch did not reopen both Issues in input order: %+v", reopened)
			}
			if call("versions", "first") == firstBefore || call("versions", "second") == secondBefore {
				t.Fatal("batch reopen did not retain both Issue successors")
			}

			runPartial := func(args ...string) string {
				t.Helper()
				ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
				defer cancel()
				partial := exec.CommandContext(ctx, bd, append(args, "--json")...)
				partial.Dir = work
				partial.Env = []string{
					"PATH=" + os.Getenv("PATH"), "HOME=" + home,
					"TMPDIR=" + os.TempDir(), "TMP=" + os.TempDir(), "TEMP=" + os.TempDir(),
					"XDG_CONFIG_HOME=" + filepath.Join(home, ".config"),
					"XDG_CACHE_HOME=" + filepath.Join(home, ".cache"),
					"GIT_CONFIG_GLOBAL=" + filepath.Join(home, "missing-gitconfig"), "GIT_CONFIG_NOSYSTEM=1",
					"BD_DISABLE_METRICS=1", "BD_DISABLE_EVENT_FLUSH=1", "BD_NON_INTERACTIVE=1",
					"BEADS_DOLT_AUTO_START=0", "DOLT_METRICS_DISABLED=1", "NO_COLOR=1",
				}
				var stdout, stderr bytes.Buffer
				partial.Stdout, partial.Stderr = &stdout, &stderr
				err := partial.Run()
				if ctx.Err() != nil {
					t.Fatalf("batch hung: %v", ctx.Err())
				}
				if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 1 {
					t.Fatalf("partial batch exit = %v; stdout=%s stderr=%s", err, stdout.String(), stderr.String())
				}
				if !strings.Contains(stderr.String(), `"code":"not_found"`) {
					t.Fatalf("missing target was not reported on stderr: %s", stderr.String())
				}
				return stdout.String()
			}

			call("close", "first")
			changed := graphMixedResult[[]graphstore.IssueMutationResult](t, runPartial("reopen", "first", "missing"))
			if len(changed) != 1 || !changed[0].Changed || changed[0].Issue.Properties.Status != types.StatusOpen {
				t.Fatalf("partial batch hid the successful reopen: %+v", changed)
			}
			call("create", "Third", "--id", "beads/third")
			closedPartial := graphMixedResult[[]graphstore.IssueMutationResult](t, runPartial("close", "third", "missing"))
			if len(closedPartial) != 1 || !closedPartial[0].Changed || closedPartial[0].Issue.Properties.Status != types.StatusClosed {
				t.Fatalf("partial batch hid the successful close: %+v", closedPartial)
			}
			call("reopen", "third")
			positional := graphMixedResult[graphstore.IssueMutationResult](t, call("done", "third", "Positional reason"))
			if !positional.Changed || positional.Issue.Properties.CloseReason != "Positional reason" {
				t.Fatalf("done alias lost its positional reason: %+v", positional)
			}
			call("create", "Fourth", "--id", "beads/fourth")
			reasonFile := filepath.Join(work, "close-reason.txt")
			if err := os.WriteFile(reasonFile, []byte("Reason from file\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			filed := graphMixedResult[graphstore.IssueMutationResult](t, call("close", "fourth", "--reason-file", reasonFile))
			if !filed.Changed || filed.Issue.Properties.CloseReason != "Reason from file\n" {
				t.Fatalf("close lost its reason file: %+v", filed)
			}
		})
	}
}
