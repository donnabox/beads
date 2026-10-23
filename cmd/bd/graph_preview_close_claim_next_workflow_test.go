//go:build cgo

package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
)

// An installed graph workspace must close and claim in one durable act. A
// missing sibling preflight leaves every Issue unchanged, while retrying a
// no-op close never claims a second Issue or creates another retained version.
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
			uri := func(id string) string { return scope + "beads/" + id }
			call(initArgs...)
			call("create", "Blocking prerequisite", "--graph-path", "beads/blocker", "--priority", "3")
			call("create", "Next work", "--graph-path", "beads/dependent", "--priority", "0")
			call("create", "Spare work", "--graph-path", "beads/spare", "--priority", "1")
			call("dep", "add", uri("dependent"), uri("blocker"))
			dependentBefore := graphMixedResult[graphstore.IssueRecord](t, call("show", "--format", "graph-json", uri("dependent")))
			spareBefore := graphMixedResult[graphstore.IssueRecord](t, call("show", "--format", "graph-json", uri("spare")))
			blockerBefore := call("versions", "blocker")
			type claimResult struct {
				Closed  []graphstore.IssueRecord `json:"closed"`
				Claimed graphstore.IssueRecord   `json:"claimed"`
			}
			first := graphMixedResult[claimResult](t, call("close", uri("blocker"), "--claim-next", "--session", "claim-session", "--reason", "Done"))
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
			if retry := graphMixedResult[[]graphstore.IssueRecord](t, call("close", uri("blocker"), "--claim-next", "--reason", "Ignored")); len(retry) != 1 || !reflect.DeepEqual(retry[0], first.Closed[0]) {
				t.Fatalf("already-closed retry changed Issue or claimed work: %+v", retry)
			}
			if call("versions", "blocker") != closedVersions || call("versions", "dependent") != dependentVersions ||
				!reflect.DeepEqual(graphMixedResult[graphstore.IssueRecord](t, call("show", "--format", "graph-json", uri("spare"))), spareBefore) {
				t.Fatal("no-op retry minted a version or claimed spare work")
			}
			call("create", "Higher priority spare", "--graph-path", "beads/next", "--priority", "0")
			nextBefore := graphMixedResult[graphstore.IssueRecord](t, call("show", "--format", "graph-json", uri("next")))
			spareVersions, nextVersions := call("versions", "spare"), call("versions", "next")
			missingClose := func() {
				t.Helper()
				ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
				defer cancel()
				cmd := graphMemoryReadCommand(ctx, bd, work, home, "close", uri("spare"), "missing", "--claim-next", "--reason", "Done", "--json")
				var stdout, stderr bytes.Buffer
				cmd.Stdout, cmd.Stderr = &stdout, &stderr
				err := cmd.Run()
				var exit *exec.ExitError
				if ctx.Err() != nil || !errors.As(err, &exit) || exit.ExitCode() != 3 || stdout.Len() != 0 || !strings.Contains(stderr.String(), `"code":"not_found"`) {
					t.Fatalf("missing close must refuse without output: err=%v stdout=%s stderr=%s", err, stdout.String(), stderr.String())
				}
			}
			for i := 0; i < 2; i++ {
				missingClose()
				if call("versions", "spare") != spareVersions || call("versions", "next") != nextVersions ||
					!reflect.DeepEqual(graphMixedResult[graphstore.IssueRecord](t, call("show", "--format", "graph-json", uri("spare"))), spareBefore) ||
					!reflect.DeepEqual(graphMixedResult[graphstore.IssueRecord](t, call("show", "--format", "graph-json", uri("next"))), nextBefore) {
					t.Fatal("missing close selector changed an Issue or claimed the next one")
				}
			}
			landed := graphMixedResult[claimResult](t, call("close", uri("spare"), "--claim-next", "--reason", "Done"))
			if len(landed.Closed) != 1 || landed.Closed[0].ID != uri("spare") || landed.Claimed.ID != uri("next") || landed.Claimed.Revision == nextBefore.Revision {
				t.Fatalf("valid close did not claim next work after refused batch: %+v", landed)
			}
			call("create", "Last unassigned work", "--graph-path", "beads/terminal")
			nothingReady := graphMixedResult[[]graphstore.IssueRecord](t, call("close", uri("terminal"), "--claim-next"))
			if len(nothingReady) != 1 || nothingReady[0].Properties.Status != types.StatusClosed {
				t.Fatalf("no-ready close did not commit without a claim: %+v", nothingReady)
			}
		})
	}
}
