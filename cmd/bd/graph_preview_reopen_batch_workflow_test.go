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
			const scope = "https://example.invalid/reopen-batch/"
			args := []string{"init", "--graph-mode", "link", "--scope-url", scope, "--skip-hooks", "--skip-agents", "--non-interactive"}
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
			uri := func(id string) string { return scope + "beads/" + id }
			call(args...)
			call("create", "First", "--id", "beads/first", "--force")
			call("create", "Second", "--id", "beads/second", "--force")
			closed := graphMixedResult[[]graphstore.IssueMutationResult](t, call("close", uri("first"), uri("second"), "--reason", "First done", "--reason", "Second done"))
			if len(closed) != 2 || !closed[0].Changed || !closed[1].Changed ||
				closed[0].Issue.Properties.Status != types.StatusClosed || closed[1].Issue.Properties.Status != types.StatusClosed {
				t.Fatalf("batch did not close both Issues in input order: %+v", closed)
			}
			firstBefore, secondBefore := call("versions", "beads/first"), call("versions", "beads/second")

			reopened := graphMixedResult[[]graphstore.IssueMutationResult](t, call("reopen", uri("first"), uri("second"), "--reason", "More work"))
			if len(reopened) != 2 || !reopened[0].Changed || !reopened[1].Changed ||
				reopened[0].Issue.Properties.Status != types.StatusOpen || reopened[1].Issue.Properties.Status != types.StatusOpen {
				t.Fatalf("batch did not reopen both Issues in input order: %+v", reopened)
			}
			if call("versions", "beads/first") == firstBefore || call("versions", "beads/second") == secondBefore {
				t.Fatal("batch reopen did not retain both Issue successors")
			}

			runMixedBatch := func(wantExit int, wantStderr string, args ...string) string {
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
				gotExit := 0
				if exit, ok := err.(*exec.ExitError); ok {
					gotExit = exit.ExitCode()
				} else if err != nil {
					t.Fatalf("mixed batch could not run: %v", err)
				}
				if gotExit != wantExit || !strings.Contains(stderr.String(), wantStderr) {
					t.Fatalf("mixed batch exit=%d want=%d; stdout=%s stderr=%s", gotExit, wantExit, stdout.String(), stderr.String())
				}
				if wantExit == 0 && strings.Contains(stderr.String(), `"code":`) {
					t.Fatalf("proxied-style mixed defer warning became a structured graph error: %s", stderr.String())
				}
				return stdout.String()
			}

			call("close", uri("first"))
			changed := graphMixedResult[[]graphstore.IssueMutationResult](t, runMixedBatch(1, `"code":"not_found"`, "reopen", uri("first"), "missing"))
			if len(changed) != 1 || !changed[0].Changed || changed[0].Issue.Properties.Status != types.StatusOpen {
				t.Fatalf("partial batch hid the successful reopen: %+v", changed)
			}
			call("create", "Third", "--id", "beads/third", "--force")
			thirdBefore := call("versions", "beads/third")
			if output := runMixedBatch(3, `"code":"not_found"`, "close", uri("third"), "missing"); output != "" {
				t.Fatalf("missing close selector emitted a successful sibling: %s", output)
			}
			if got := call("versions", "beads/third"); got != thirdBefore {
				t.Fatal("missing close selector changed a sibling Issue")
			}
			// Ordinary direct defer/undefer pre-resolves all IDs and rejects a
			// missing sibling before writing, while the ordinary proxied route
			// commits survivors and returns 0 with a plain stderr warning.
			// Preserve the graph route's existing proxied-style behavior until
			// that native split receives a release-contract decision.
			deferredPartial := graphMixedResult[[]graphstore.IssueMutationResult](t, runMixedBatch(0, "Error changing deferral for missing:", "defer", uri("third"), "missing"))
			if len(deferredPartial) != 1 || !deferredPartial[0].Changed || deferredPartial[0].Issue.Properties.Status != types.StatusDeferred {
				t.Fatalf("mixed defer hid the successful Issue: %+v", deferredPartial)
			}
			undeferredPartial := graphMixedResult[[]graphstore.IssueMutationResult](t, runMixedBatch(0, "Error changing deferral for missing:", "undefer", uri("third"), "missing"))
			if len(undeferredPartial) != 1 || !undeferredPartial[0].Changed || undeferredPartial[0].Issue.Properties.Status != types.StatusOpen {
				t.Fatalf("mixed undefer hid the successful Issue: %+v", undeferredPartial)
			}
			positional := graphMixedResult[graphstore.IssueMutationResult](t, call("done", uri("third"), "Positional reason"))
			if !positional.Changed || positional.Issue.Properties.CloseReason != "Positional reason" {
				t.Fatalf("done alias lost its positional reason: %+v", positional)
			}
			call("create", "Fourth", "--id", "beads/fourth", "--force")
			reasonFile := filepath.Join(work, "close-reason.txt")
			if err := os.WriteFile(reasonFile, []byte("Reason from file\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			filed := graphMixedResult[graphstore.IssueMutationResult](t, call("close", uri("fourth"), "--reason-file", reasonFile))
			if !filed.Changed || filed.Issue.Properties.CloseReason != "Reason from file\n" {
				t.Fatalf("close lost its reason file: %+v", filed)
			}
		})
	}
}

// --force and --session take the native Issue close route on both installed
// engines; retrying a closed Issue cannot replace its first close reason or
// session, and a refused close cannot create a graph version.
func TestGraphPreviewCloseForceAndSessionWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			initArgs := []string{"init", "--graph-mode", "link", "--scope-url", "https://example.invalid/close-policy/", "--skip-hooks", "--skip-agents", "--non-interactive"}
			if engine == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for ordinary shared-server CLI qualification")
				}
				initArgs = append(initArgs, "--server", "--external", "--server-host", "127.0.0.1", "--server-port", port, "--server-user", "root")
			}
			call := func(extraEnv []string, code string, args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, extraEnv, code, append(args, "--json")...)
			}
			call(nil, "", initArgs...)
			call(nil, "", "create", "Prerequisite", "--id", "beads/prerequisite", "--force")
			call(nil, "", "create", "Blocked work", "--id", "beads/work", "--force")
			call(nil, "", "dep", "add", "https://example.invalid/close-policy/beads/work", "https://example.invalid/close-policy/beads/prerequisite")
			before := call(nil, "", "versions", "beads/work")
			call(nil, "constraint_violation", "close", "https://example.invalid/close-policy/beads/work", "--reason", "Not yet")
			if got := call(nil, "", "versions", "beads/work"); got != before {
				t.Fatal("refused blocked close recorded a version")
			}
			closed := graphMixedResult[graphstore.IssueMutationResult](t, call(nil, "", "close", "https://example.invalid/close-policy/beads/work", "--force", "--reason", "Override", "--session", "session-one"))
			if !closed.Changed || closed.Issue.Properties.Status != types.StatusClosed || closed.Issue.Properties.CloseReason != "Override" || closed.Issue.Properties.ClosedBySession != "session-one" {
				t.Fatalf("forced close lost native state or session: %+v", closed)
			}
			closedVersions := call(nil, "", "versions", "beads/work")
			if closedVersions == before {
				t.Fatal("forced close did not retain its Issue successor")
			}
			repeated := graphMixedResult[graphstore.IssueMutationResult](t, call(nil, "", "close", "https://example.invalid/close-policy/beads/work", "--force", "--reason", "Must not replace", "--session", "session-two"))
			if repeated.Changed || repeated.Issue.Properties.CloseReason != "Override" || repeated.Issue.Properties.ClosedBySession != "session-one" || call(nil, "", "versions", "beads/work") != closedVersions {
				t.Fatalf("re-close changed the first close or retained a successor: %+v", repeated)
			}

			call(nil, "", "create", "Assigned work", "--id", "beads/assigned", "--force", "--assignee", "foreign-holder")
			assignedBefore := call(nil, "", "versions", "beads/assigned")
			call(nil, "invalid_properties", "close", "https://example.invalid/close-policy/beads/assigned")
			if call(nil, "", "versions", "beads/assigned") != assignedBefore {
				t.Fatal("refused foreign-holder close recorded a version")
			}
			forced := graphMixedResult[graphstore.IssueMutationResult](t, call([]string{"CLAUDE_SESSION_ID=session-from-env"}, "", "close", "https://example.invalid/close-policy/beads/assigned", "--force"))
			if !forced.Changed || forced.Issue.Properties.ClosedBySession != "session-from-env" || forced.Issue.Properties.Status != types.StatusClosed {
				t.Fatalf("forced holder close lost session fallback: %+v", forced)
			}
		})
	}
}

// A close can hand the user newly unblocked work without claiming or
// versioning that work. The suggestion uses the native dependency query but
// exposes only canonical graph Issues.
func TestGraphPreviewCloseSuggestNextWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			initArgs := []string{"init", "--graph-mode", "link", "--scope-url", "https://example.invalid/close-suggest/", "--skip-hooks", "--skip-agents", "--non-interactive"}
			if engine == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for ordinary shared-server CLI qualification")
				}
				initArgs = append(initArgs, "--server", "--external", "--server-host", "127.0.0.1", "--server-port", port, "--server-user", "root")
			}
			call := func(code string, args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, code, append(args, "--json")...)
			}
			call("", initArgs...)
			call("", "create", "Prerequisite", "--id", "beads/prerequisite", "--force")
			call("", "create", "Dependent", "--id", "beads/dependent", "--force")
			call("", "dep", "add", "https://example.invalid/close-suggest/beads/dependent", "https://example.invalid/close-suggest/beads/prerequisite")
			dependentBefore := call("", "versions", "beads/dependent")
			prerequisiteBefore := call("", "versions", "beads/prerequisite")
			call("", "create", "Unrelated", "--id", "beads/unrelated", "--force")
			unrelatedBefore := call("", "versions", "beads/unrelated")
			call("invalid_selector", "close", "https://example.invalid/close-suggest/beads/prerequisite", "https://example.invalid/close-suggest/beads/unrelated", "--suggest-next")
			if call("", "versions", "beads/prerequisite") != prerequisiteBefore || call("", "versions", "beads/unrelated") != unrelatedBefore {
				t.Fatal("multi-target suggestion refusal changed an Issue")
			}
			type suggestion struct {
				Closed    []graphstore.IssueRecord `json:"closed"`
				Unblocked []graphstore.IssueRecord `json:"unblocked"`
			}
			first := graphMixedResult[suggestion](t, call("", "close", "https://example.invalid/close-suggest/beads/prerequisite", "--suggest-next", "--reason", "Done"))
			if len(first.Closed) != 1 || first.Closed[0].Properties.Status != types.StatusClosed ||
				len(first.Unblocked) != 1 || first.Unblocked[0].ID != "https://example.invalid/close-suggest/beads/dependent" {
				t.Fatalf("close did not name newly unblocked Issue: %+v", first)
			}
			if call("", "versions", "beads/dependent") != dependentBefore {
				t.Fatal("suggestion changed or claimed the dependent Issue")
			}
			closedVersions := call("", "versions", "beads/prerequisite")
			if closedVersions == prerequisiteBefore {
				t.Fatal("close did not retain its Issue successor")
			}
			again := graphMixedResult[suggestion](t, call("", "close", "https://example.invalid/close-suggest/beads/prerequisite", "--suggest-next", "--reason", "Different"))
			if len(again.Unblocked) != 1 || call("", "versions", "beads/prerequisite") != closedVersions {
				t.Fatalf("idempotent close lost suggestion or minted a version: %+v", again)
			}
			alone := graphMixedResult[graphstore.IssueMutationResult](t, call("", "close", "https://example.invalid/close-suggest/beads/unrelated", "--suggest-next"))
			if !alone.Changed || alone.Issue.Properties.Status != types.StatusClosed {
				t.Fatalf("empty suggestion did not preserve close result: %+v", alone)
			}
		})
	}
}

// Selector resolution belongs to each partial-batch member, while command-wide
// syntax and revision-guard admission must still refuse before any writes.
func TestGraphPreviewDeferralPartialBatchWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			const scope = "https://example.invalid/defer-partial/"
			work, home := graphListWorkspace(t, bd, engine, scope)
			call := func(code string, args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, code, append(args, "--json")...)
			}
			issue := graphMixedResult[graphstore.IssueRecord](t, call("", "create", "Work", "--graph-path", "beads/private-work"))
			memory := graphMixedResult[graphstore.Record](t, call("", "remember", "Context", "--id", "context"))
			memoryBefore := call("", "show", memory.ID)
			for _, command := range []string{"defer", "undefer"} {
				before := call("", "versions", issue.ID)
				call("invalid_selector", command, issue.ID, "missing", "--if-revision", issue.Revision)
				call("invalid_selector", command, issue.Properties.ID+"@current", "missing")
				if got := call("", "versions", issue.ID); got != before {
					t.Fatal("command-wide refusal changed Issue History")
				}
				stdout, stderr, exit := graphGenerationRun(t, bd, work, home, command, "missing", issue.ID, memory.ID, "--json")
				if exit != 0 || !strings.Contains(stderr, "Error changing deferral for missing:") || !strings.Contains(stderr, "Error changing deferral for "+memory.ID+":") || strings.Contains(stderr, `"code":`) {
					t.Fatalf("%s partial batch lost per-member warnings: exit=%d stdout=%s stderr=%s", command, exit, stdout, stderr)
				}
				results := graphMixedResult[[]graphstore.IssueMutationResult](t, stdout)
				want := types.StatusDeferred
				if command == "undefer" {
					want = types.StatusOpen
				}
				if len(results) != 1 || !results[0].Changed || results[0].Issue.ID != issue.ID || results[0].Issue.Properties.Status != want {
					t.Fatalf("%s lost successful sibling: %+v", command, results)
				}
				_, _, oldRows, _ := graphVersionsListed(t, before)
				after := call("", "versions", issue.ID)
				_, _, newRows, _ := graphVersionsListed(t, after)
				if len(newRows) != len(oldRows)+1 || call("", "show", memory.ID) != memoryBefore {
					t.Fatal("partial batch lost exact Issue successor or changed wrong-Kind Memory")
				}
				call("revision_conflict", command, issue.ID, "--if-revision", issue.Revision)
				if call("", "versions", issue.ID) != after {
					t.Fatal("stale single-member guard changed History")
				}
			}
		})
	}
}
