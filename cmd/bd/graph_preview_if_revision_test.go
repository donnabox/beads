//go:build cgo

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// --if-revision is one flag with two meanings, and which one applies depends on
// the workspace. In a graph_mode link workspace it is an opaque observed graph
// token. Everywhere else on update, delete, close and assign it is the
// compare-and-swap upstream added in #7203: a decimal bead revision whose
// mismatch exits 13, which gascity's bdstore_conditional.go classifier reads.
//
// Upstream registers the flag on those four verbs, so the preview registers it
// only where upstream has none (remember, forget, graph unlink) and its link
// mode handlers read upstream's registration. Registering it twice panics at
// init, so these tests also keep the whole cmd/bd test binary startable.

// graphIfRevisionGuardMismatchExit is a literal on purpose: the contract with
// gascity is the number, not whatever name the constant for it happens to have.
const graphIfRevisionGuardMismatchExit = 13

// graphIfRevisionRun runs one disposable CLI process and returns both streams
// with its exit status. graphPolicyCLI cannot stand in: it demands a typed
// refusal on stderr, and a refused upstream --if-revision write leads with a
// human line there, so the status itself is part of what these tests pin.
func graphIfRevisionRun(t *testing.T, bd, work, home string, args ...string) (stdout, stderr string, exit int) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
	defer cancel()
	cmd := graphMemoryReadCommand(ctx, bd, work, home, args...)
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	err := cmd.Run()
	if ctx.Err() != nil {
		t.Fatalf("CLI did not complete within deadline: %v args=%v\n%s", ctx.Err(), args, errOut.String())
	}
	var exitErr *exec.ExitError
	switch {
	case err == nil:
	case errors.As(err, &exitErr):
		exit = exitErr.ExitCode()
	default:
		t.Fatalf("could not run bd %v: %v", args, err)
	}
	return out.String(), errOut.String(), exit
}

// graphIfRevisionRefused asserts a typed graph refusal: exactly this exit
// status, nothing on stdout, and a stderr that is one {"code": ...} document.
func graphIfRevisionRefused(t *testing.T, bd, work, home string, wantExit int, wantCode string, args ...string) {
	t.Helper()
	stdout, stderr, exit := graphIfRevisionRun(t, bd, work, home, append(append([]string(nil), args...), "--json")...)
	var diagnostic struct {
		Code      string `json:"code"`
		Retryable bool   `json:"retryable"`
	}
	if exit != wantExit || stdout != "" || json.Unmarshal([]byte(stderr), &diagnostic) != nil || diagnostic.Code != wantCode || diagnostic.Retryable {
		t.Fatalf("bd %v: want exit %d with a non-retryable typed %s refusal, got exit %d\nstdout: %s\nstderr: %s", args, wantExit, wantCode, exit, stdout, stderr)
	}
}

// Who owns --if-revision on each verb is part of the CLI contract, visible in
// --help: upstream's definition on the four verbs it owns, the preview's own
// on the three it does not. Only destructive preview verbs retain an explicit
// --unconditional choice.
func TestGraphPreviewIfRevisionFlagOwnership(t *testing.T) {
	for _, cmd := range []*cobra.Command{updateCmd, deleteCmd, closeCmd, assignCmd} {
		flag := cmd.Flags().Lookup("if-revision")
		if flag == nil {
			t.Errorf("bd %s has no --if-revision", cmd.Name())
			continue
		}
		if !strings.HasPrefix(flag.Usage, ifRevisionFlagHelp) {
			t.Errorf("bd %s --if-revision is not upstream's compare-and-swap flag; usage = %q", cmd.Name(), flag.Usage)
		}
	}
	for _, cmd := range []*cobra.Command{rememberCmd, forgetCmd, graphUnlinkCmd} {
		flag := cmd.Flags().Lookup("if-revision")
		if flag == nil {
			t.Errorf("bd %s lost its graph preview --if-revision", cmd.Name())
			continue
		}
		if strings.HasPrefix(flag.Usage, ifRevisionFlagHelp) {
			t.Errorf("bd %s --if-revision claims upstream's decimal compare-and-swap meaning; usage = %q", cmd.Name(), flag.Usage)
		}
	}
	for _, tc := range []struct {
		cmd  *cobra.Command
		want bool
	}{
		{updateCmd, false}, {deleteCmd, true}, {forgetCmd, true}, {rememberCmd, false}, {graphUnlinkCmd, true},
		{closeCmd, false}, {assignCmd, false},
	} {
		if has := tc.cmd.Flags().Lookup("unconditional") != nil; has != tc.want {
			t.Errorf("bd %s has --unconditional = %t, want %t (fork-only flag)", tc.cmd.Name(), has, tc.want)
		}
	}
}

// Outside link mode the preview must stay out of the way: a stale decimal token
// on update or delete is upstream's compare-and-swap (exit 13, nothing written),
// not a graph-only flag refused with capability_unavailable (exit 5). What stays
// graph-only is --unconditional and, on forget, --if-revision.
func TestGraphPreviewIfRevisionOutsideLinkModeIsUpstreamCAS(t *testing.T) {
	bd := buildBDUnderTest(t)
	work, home := t.TempDir(), t.TempDir()
	call := func(args ...string) string {
		t.Helper()
		return graphPolicyCLI(t, bd, work, home, nil, "", args...)
	}
	call("init", "--prefix", "ifrev", "--skip-hooks", "--skip-agents", "--non-interactive", "--json")
	var created struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(call("create", "Stale guard", "--json")), &created); err != nil || created.ID == "" {
		t.Fatalf("create did not report an id: %v", err)
	}
	id := created.ID
	// bd show --json is an array whose elements carry the revision as a decimal
	// string, the same value --if-revision takes back.
	revision := func() int64 {
		t.Helper()
		var shown []struct {
			Revision string `json:"revision"`
		}
		out := call("show", id, "--json")
		if err := json.NewDecoder(strings.NewReader(out)).Decode(&shown); err != nil || len(shown) != 1 {
			t.Fatalf("show %s --json: %v\n%s", id, err, out)
		}
		value, err := strconv.ParseInt(shown[0].Revision, 10, 64)
		if err != nil {
			t.Fatalf("show %s reported revision %q, want a decimal int64: %v", id, shown[0].Revision, err)
		}
		return value
	}
	stale := revision()
	call("update", id, "--notes", "bump", "--json")
	current := revision()
	if current == stale {
		t.Fatal("an unguarded update did not advance the revision")
	}

	for _, tc := range []struct {
		name string
		args []string
	}{
		{"update", []string{"update", id, "--priority", "3", "--if-revision", strconv.FormatInt(stale, 10)}},
		{"delete preview", []string{"delete", id, "--if-revision", strconv.FormatInt(stale, 10)}},
		{"delete force", []string{"delete", id, "--force", "--if-revision", strconv.FormatInt(stale, 10)}},
	} {
		_, stderr, exit := graphIfRevisionRun(t, bd, work, home, append(tc.args, "--json")...)
		if exit != graphIfRevisionGuardMismatchExit {
			t.Fatalf("stale --if-revision %s exited %d, want %d (precondition_failed, not the graph preview's refusal)\n%s", tc.name, exit, graphIfRevisionGuardMismatchExit, stderr)
		}
		body := decodeIfRevisionBody(t, lastJSONLine(t, stderr), false)
		if body["code"] != ifRevisionCodePreconditionFailed || body["id"] != id || body["expected_revision"] != float64(stale) {
			t.Fatalf("stale --if-revision %s reported %v, want precondition_failed for %s expecting revision %d", tc.name, body, id, stale)
		}
		if got := revision(); got != current {
			t.Fatalf("stale --if-revision %s wrote: revision %d -> %d", tc.name, current, got)
		}
	}

	for _, args := range [][]string{
		{"delete", id, "--force", "--unconditional"},
		{"forget", "ifrev-key", "--unconditional"},
		{"forget", "ifrev-key", "--if-revision", "1"},
	} {
		graphIfRevisionRefused(t, bd, work, home, 5, "capability_unavailable", args...)
	}

	// A token that is not a decimal is upstream's typo refusal here, never a
	// graph token: those only exist in a link workspace.
	for _, args := range [][]string{
		{"update", id, "--priority", "3", "--if-revision", "observed"},
		{"delete", id, "--force", "--if-revision", "observed"},
	} {
		stdout, stderr, exit := graphIfRevisionRun(t, bd, work, home, args...)
		if exit != 1 || !strings.Contains(stdout+stderr, `invalid --if-revision "observed"`) {
			t.Fatalf("bd %v exited %d, want upstream's exit 1 refusal of a non-decimal token\nstdout: %s\nstderr: %s", args, exit, stdout, stderr)
		}
	}
	if got := revision(); got != current {
		t.Fatalf("a refused write changed the revision: %d -> %d", current, got)
	}

	call("update", id, "--priority", "3", "--if-revision", strconv.FormatInt(current, 10), "--json")
	updated := revision()
	if updated == current {
		t.Fatal("a matching --if-revision update did not apply")
	}
	call("delete", id, "--force", "--if-revision", strconv.FormatInt(updated, 10), "--json")
	if _, _, exit := graphIfRevisionRun(t, bd, work, home, "show", id, "--json"); exit == 0 {
		t.Fatal("a matching --if-revision delete left the issue in place")
	}
}

// In a graph_mode link workspace the same flag still carries a graph token. Both
// of these tokens would go wrong in upstream's parser (a typo, and a stale
// decimal), so a typed revision_conflict is the proof that the graph handler,
// not upstream's compare-and-swap, received them.
func TestGraphPreviewIfRevisionInLinkModeStaysGraphToken(t *testing.T) {
	bd := buildBDUnderTest(t)
	work, home := t.TempDir(), t.TempDir()
	call := func(args ...string) string {
		t.Helper()
		return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
	}
	refuse := func(code string, args ...string) {
		t.Helper()
		graphPolicyCLI(t, bd, work, home, nil, code, append(args, "--json")...)
	}
	call("init", "--graph-mode", "link", "--scope-url", "https://example.invalid/if-revision/", "--skip-hooks", "--skip-agents", "--non-interactive")
	memory := graphMixedResult[graphstore.Record](t, call("remember", "Guarded body", "--id", "beads/plan", "--title", "Plan"))
	const replacement = `{"title":"Plan","body":"Replaced body"}`
	before := call("show", "beads/plan")
	for _, token := range []string{"observed", "7"} {
		refuse("revision_conflict", "update", "beads/plan", "--properties", replacement, "--if-revision", token)
		refuse("revision_conflict", "delete", "beads/plan", "--if-revision", token)
		refuse("revision_conflict", "delete", "beads/plan", "--force", "--if-revision", token)
	}
	if call("show", "beads/plan") != before {
		t.Fatal("a refused graph token changed the Memory")
	}

	replaced := graphMixedResult[graphstore.MemoryMutationResult](t, call("update", "beads/plan", "--properties", replacement, "--if-revision", memory.Revision))
	if !replaced.Changed || replaced.Memory.Revision == memory.Revision {
		t.Fatal("the observed graph token did not guard-and-apply the update")
	}
	deleted := graphMixedResult[graphstore.MemoryDeleteResult](t, call("delete", "beads/plan", "--force", "--if-revision", replaced.Memory.Revision))
	if !deleted.Deleted || deleted.Preview {
		t.Fatal("the observed graph token did not guard-and-apply the deletion")
	}
}

// In graph mode close now interprets --if-revision as the opaque observed graph
// Issue token. Assign has no graph route and still refuses the flag rather than
// quietly dropping its upstream numeric guard.
func TestGraphPreviewCloseIfRevisionUsesGraphToken(t *testing.T) {
	bd := buildBDUnderTest(t)
	work, home := t.TempDir(), t.TempDir()
	call := func(args ...string) string {
		t.Helper()
		return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
	}
	call("init", "--graph-mode", "link", "--scope-url", "https://example.invalid/close-if-revision/", "--skip-hooks", "--skip-agents", "--non-interactive")
	call("create", "Close guard", "--id", "beads/close-guard")
	before := call("show", "beads/close-guard")
	for _, args := range [][]string{
		{"close", "beads/close-guard", "--if-revision", "1"},
		{"close", "beads/close-guard", "--reason", "Done", "--if-revision", "1"},
	} {
		graphIfRevisionRefused(t, bd, work, home, 4, "revision_conflict", args...)
	}
	graphIfRevisionRefused(t, bd, work, home, 5, "capability_unavailable", "assign", "beads/close-guard", "alice", "--if-revision", "1")
	if call("show", "beads/close-guard") != before {
		t.Fatal("a refused guarded close changed the Issue")
	}
	observed := graphMixedResult[graphstore.IssueRecord](t, before)
	if closed := graphMixedResult[graphstore.IssueMutationResult](t, call("close", "beads/close-guard", "--reason", "Done", "--if-revision", observed.Revision)); !closed.Changed {
		t.Fatal("a close with the observed graph revision did not apply")
	}
}
