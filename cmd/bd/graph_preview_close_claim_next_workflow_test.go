//go:build cgo

package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
)

// An installed graph workspace must close and claim in one durable act. A
// missing sibling must prevent both the close and the claim, while retrying
// a no-op close never claims a second Issue or creates another retained version.
func TestGraphPreviewCloseClaimNextWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			const scope = "https://example.invalid/close-claim-next/"
			initArgs := []string{"init", "--graph-mode", "link", "--scope-url", scope, "--skip-hooks", "--skip-agents", "--non-interactive"}
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
			call(initArgs...)
			call("create", "Blocking prerequisite", "--id", "blocker", "--priority", "3")
			call("create", "Next work", "--id", "dependent", "--priority", "0")
			call("create", "Spare work", "--id", "spare", "--priority", "1")
			call("dep", "add", "dependent", "blocker")
			dependentBefore := graphMixedResult[graphstore.IssueRecord](t, call("show", "--format", "graph-json", "dependent"))
			spareBefore := graphMixedResult[graphstore.IssueRecord](t, call("show", "--format", "graph-json", "spare"))
			blockerBefore := call("versions", "blocker")
			type claimResult struct {
				Closed  []graphstore.IssueRecord `json:"closed"`
				Claimed graphstore.IssueRecord   `json:"claimed"`
			}
			first := graphMixedResult[claimResult](t, call("close", "blocker", "--claim-next", "--session", "claim-session", "--reason", "Done"))
			if len(first.Closed) != 1 || first.Closed[0].Properties.Status != types.StatusClosed ||
				first.Closed[0].Properties.ClosedBySession != "claim-session" || first.Claimed.ID != scope+"beads/dependent" ||
				first.Claimed.Properties.Status != types.StatusInProgress || first.Claimed.Properties.Assignee == "" ||
				first.Claimed.Properties.LeaseExpiresAt == nil || first.Claimed.Properties.HeartbeatAt == nil ||
				first.Claimed.Properties.LeaseExpiresAt.Sub(*first.Claimed.Properties.HeartbeatAt) != 5*time.Minute ||
				first.Claimed.Revision == dependentBefore.Revision {
				t.Fatalf("close did not atomically claim newly ready Issue: %+v", first)
			}
			if call("versions", "blocker") == blockerBefore {
				t.Fatal("close did not retain the native Issue successor")
			}
			closedVersions, dependentVersions := call("versions", "blocker"), call("versions", "dependent")
			if retry := graphMixedResult[[]graphstore.IssueRecord](t, call("close", "blocker", "--claim-next", "--reason", "Ignored")); len(retry) != 1 || !reflect.DeepEqual(retry[0], first.Closed[0]) {
				t.Fatalf("already-closed retry changed Issue or claimed work: %+v", retry)
			}
			if call("versions", "blocker") != closedVersions || call("versions", "dependent") != dependentVersions ||
				!reflect.DeepEqual(graphMixedResult[graphstore.IssueRecord](t, call("show", "--format", "graph-json", "spare")), spareBefore) {
				t.Fatal("no-op retry minted a version or claimed spare work")
			}
			call("create", "Higher priority spare", "--id", "next", "--priority", "0")
			missingClose := func(args ...string) string {
				t.Helper()
				ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
				defer cancel()
				cmd := exec.CommandContext(ctx, bd, append(args, "--json")...)
				cmd.Dir = work
				cmd.Env = []string{
					"PATH=" + os.Getenv("PATH"), "HOME=" + home,
					"TMPDIR=" + os.TempDir(), "TMP=" + os.TempDir(), "TEMP=" + os.TempDir(),
					"XDG_CONFIG_HOME=" + filepath.Join(home, ".config"), "XDG_CACHE_HOME=" + filepath.Join(home, ".cache"),
					"GIT_CONFIG_GLOBAL=" + filepath.Join(home, "missing-gitconfig"), "GIT_CONFIG_NOSYSTEM=1",
					"BD_DISABLE_METRICS=1", "BD_DISABLE_EVENT_FLUSH=1", "BD_NON_INTERACTIVE=1",
					"BEADS_DOLT_AUTO_START=0", "DOLT_METRICS_DISABLED=1", "NO_COLOR=1",
				}
				var stdout, stderr bytes.Buffer
				cmd.Stdout, cmd.Stderr = &stdout, &stderr
				err := cmd.Run()
				if ctx.Err() != nil {
					t.Fatalf("missing-selector close timed out: %v", ctx.Err())
				}
				if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 3 || !strings.Contains(stderr.String(), `"code":"not_found"`) {
					t.Fatalf("missing close did not report absent sibling: %v stdout=%s stderr=%s", err, stdout.String(), stderr.String())
				}
				return stdout.String()
			}
			spareVersionsBefore, nextVersionsBefore := call("versions", "spare"), call("versions", "next")
			if output := missingClose("close", "spare", "missing", "--claim-next", "--reason", "Done"); output != "" {
				t.Fatalf("missing selector emitted a committed close or claim: %s", output)
			}
			if call("versions", "spare") != spareVersionsBefore || call("versions", "next") != nextVersionsBefore {
				t.Fatal("missing selector changed the sibling Issue or claimed next work")
			}
			landed := graphMixedResult[claimResult](t, call("close", "spare", "--claim-next", "--reason", "Done"))
			if len(landed.Closed) != 1 || landed.Closed[0].ID != scope+"beads/spare" || landed.Claimed.ID != scope+"beads/next" {
				t.Fatalf("valid retry did not close and claim: %+v", landed)
			}
			call("create", "Last unassigned work", "--id", "terminal")
			nothingReady := graphMixedResult[[]graphstore.IssueRecord](t, call("close", "terminal", "--claim-next"))
			if len(nothingReady) != 1 || nothingReady[0].Properties.Status != types.StatusClosed {
				t.Fatalf("no-ready close did not commit without a claim: %+v", nothingReady)
			}
		})
	}
}
