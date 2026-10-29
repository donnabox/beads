package translate

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// standInBd writes an executable sh script that stands in for bd and returns its
// path. The test runs it through ExecuteWith, so what matters is how it ends.
func standInBd(t *testing.T, body string) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "bd")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\n"+body+"\n"), 0o755); err != nil { // #nosec G306 -- a test stand-in must be executable
		t.Fatalf("writing the stand-in: %v", err)
	}
	return bin
}

// A bd that ran and exited non-zero refused the action. ExecuteWith says so with
// a typed error that carries the invocation, the exit code and what bd printed,
// so a caller can record the refusal without parsing text. The exit error stays
// reachable underneath, and the message keeps its form.
func TestExecuteWith_NonZeroExitIsAnExecError(t *testing.T) {
	bin := standInBd(t, "echo 'refusing to update' >&2\necho 'and a line on stdout'\nexit 3")
	action := Action{Kind: KindUpdate, Issue: "x-1", Argv: []string{"update", "x-1", "--title", "A title"}}

	err := ExecuteWith(context.Background(), bin, t.TempDir(), action)
	var refused *ExecError
	if !errors.As(err, &refused) {
		t.Fatalf("err = %v (%T), want an *ExecError", err, err)
	}
	if refused.ExitCode != 3 {
		t.Errorf("ExitCode = %d, want 3", refused.ExitCode)
	}
	if !reflect.DeepEqual(refused.Argv, action.Argv) {
		t.Errorf("Argv = %v, want %v", refused.Argv, action.Argv)
	}
	if !strings.Contains(refused.Output, "refusing to update") || !strings.Contains(refused.Output, "and a line on stdout") {
		t.Errorf("Output = %q, want bd's stderr and stdout", refused.Output)
	}
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 3 {
		t.Errorf("the *exec.ExitError is not reachable under the ExecError: %v", err)
	}
	for _, want := range []string{"bd update x-1 --title A title", "exit status 3", "refusing to update"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("message %q does not contain %q", err.Error(), want)
		}
	}
}

// Only a refusal is an ExecError. A bd killed by a signal, and one that cannot
// be started at all, are failures of the harness's own machinery, and a caller
// must not read either as bd saying no.
func TestExecuteWith_OnlyARefusalIsAnExecError(t *testing.T) {
	cases := map[string]string{
		"killed by a signal": standInBd(t, "kill -9 $$"),
		"cannot be started":  filepath.Join(t.TempDir(), "no-such-bd"),
	}
	for name, bin := range cases {
		t.Run(name, func(t *testing.T) {
			err := ExecuteWith(context.Background(), bin, t.TempDir(), Action{Kind: KindUpdate, Argv: []string{"update", "x-1"}})
			if err == nil {
				t.Fatal("ExecuteWith succeeded")
			}
			var refused *ExecError
			if errors.As(err, &refused) {
				t.Errorf("a bd that %s was read as a refusal: %+v", name, refused)
			}
		})
	}
}
