package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/testutil/bazeltest"
)

// TestB1MainsCharacterised_TranslatorDryRun drives the real command line with
// --dry-run against a stand-in dolt that returns canned row diffs, and checks
// stdout, exit status and every query the command sends to dolt against a
// golden transcript captured before the command's logic moved into a library.
// The stand-in keeps the test free of external tools, so it runs everywhere
// the command can be built.

var updateGolden = flag.Bool("update", false, "rewrite golden transcripts from the current output")

const (
	cliFromRef = "0123456789abcdefghijklmnopqrstuv"
	cliToRef   = "vutsrqponmlkjihgfedcba9876543210"
)

func buildCLI(t *testing.T) string {
	t.Helper()
	if bazeltest.IsBazel() {
		t.Skip("the command is built from source, which the Bazel sandbox does not hold; the go test lanes run this test")
	}
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Fatalf("go toolchain not found: %v", err)
	}
	out := filepath.Join(t.TempDir(), "mutation-translator")
	if b, err := exec.Command(goBin, "build", "-o", out, ".").CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, b)
	}
	return out
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// translatorShim is a stand-in dolt that answers the two queries the command
// issues from canned CSV files and records every argument it is given.
type translatorShim struct {
	dir     string
	argvLog string
}

func newTranslatorShim(t *testing.T, issuesCSV, depsCSV string) translatorShim {
	t.Helper()
	dir := t.TempDir()
	s := translatorShim{dir: dir, argvLog: filepath.Join(dir, "argv.log")}
	issues, deps := filepath.Join(dir, "issues.csv"), filepath.Join(dir, "deps.csv")
	for path, content := range map[string]string{issues: issuesCSV, deps: depsCSV} {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("writing %s: %v", path, err)
		}
	}
	script := fmt.Sprintf("#!/bin/sh\nprintf '%%s\\n' \"$@\" >> %s\ncase \"$3\" in\n*dolt_commit_diff_issues*) cat %s ;;\n*dolt_commit_diff_dependencies*) cat %s ;;\nesac\n",
		shellQuote(s.argvLog), shellQuote(issues), shellQuote(deps))
	if err := os.WriteFile(filepath.Join(dir, "dolt"), []byte(script), 0o755); err != nil {
		t.Fatalf("writing the dolt stand-in: %v", err)
	}
	return s
}

func runTranslator(t *testing.T, bin string, shim translatorShim, args ...string) (exit int, stdout, stderr string) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Env = []string{"PATH=" + shim.dir + ":/usr/bin:/bin", "HOME=" + t.TempDir()}
	var so, se bytes.Buffer
	cmd.Stdout, cmd.Stderr = &so, &se
	if err := cmd.Run(); err != nil {
		var ee *exec.ExitError
		if !errors.As(err, &ee) {
			t.Fatalf("running %s: %v", bin, err)
		}
		exit = ee.ExitCode()
	}
	return exit, so.String(), se.String()
}

func readShimLog(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ""
		}
		t.Fatalf("reading %s: %v", path, err)
	}
	return string(b)
}

func checkGolden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name)
	if *updateGolden {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatalf("creating testdata: %v", err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatalf("writing %s: %v", path, err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading golden %s (run with -update to create it): %v", path, err)
	}
	if string(want) != got {
		t.Errorf("output differs from %s\n--- want\n%s\n--- got\n%s", path, want, got)
	}
}

func TestB1MainsCharacterised_TranslatorDryRun(t *testing.T) {
	bin := buildCLI(t)

	const (
		issuesHeaderOnly = "diff_type,to_id,from_id\n"
		depsHeaderOnly   = "diff_type,to_issue_id,to_depends_on_issue_id,to_type,from_issue_id,from_depends_on_issue_id\n"
	)
	cases := []struct {
		name, issuesCSV, depsCSV string
	}{
		{
			name:      "issue added",
			issuesCSV: "diff_type,to_id,to_title,to_description,to_priority,to_issue_type,to_status,from_id\nadded,x-1,\"Widget, deluxe\",Does things,2,task,open,\n",
			depsCSV:   depsHeaderOnly,
		},
		{
			name:      "issue modified",
			issuesCSV: "diff_type,to_id,to_title,to_description,to_status,from_id,from_title,from_description,from_status\nmodified,x-1,New title,New desc,open,x-1,Old title,Old desc,open\n",
			depsCSV:   depsHeaderOnly,
		},
		{
			name:      "issue closed",
			issuesCSV: "diff_type,to_id,to_status,to_close_reason,to_closed_at,from_id,from_status,from_close_reason,from_closed_at\nmodified,x-1,closed,done,2026-01-01 00:00:00,x-1,open,,\n",
			depsCSV:   depsHeaderOnly,
		},
		{
			name:      "issue removed",
			issuesCSV: "diff_type,to_id,from_id\nremoved,,x-1\n",
			depsCSV:   depsHeaderOnly,
		},
		{
			name:      "dependency added",
			issuesCSV: issuesHeaderOnly,
			depsCSV:   depsHeaderOnly + "added,x-1,x-2,blocks,,\n",
		},
		{
			name:      "dependency removed",
			issuesCSV: issuesHeaderOnly,
			depsCSV:   depsHeaderOnly + "removed,,,,x-1,x-2\n",
		},
		{
			name:      "nothing changed",
			issuesCSV: issuesHeaderOnly,
			depsCSV:   depsHeaderOnly,
		},
	}

	var transcript strings.Builder
	for _, tc := range cases {
		shim := newTranslatorShim(t, tc.issuesCSV, tc.depsCSV)
		exit, stdout, stderr := runTranslator(t, bin, shim,
			"--data-dir", t.TempDir(), "--work-dir", t.TempDir(),
			"--from", cliFromRef, "--to", cliToRef, "--issue", "x-1", "--dry-run")
		if exit != 0 {
			t.Fatalf("%s: exited %d\nstderr: %s", tc.name, exit, stderr)
		}
		fmt.Fprintf(&transcript, "== %s\nexit: %d\nstdout:\n%sdolt argv:\n%s\n", tc.name, exit, stdout, readShimLog(t, shim.argvLog))
	}

	shim := newTranslatorShim(t, "", "")
	exit, stdout, stderr := runTranslator(t, bin, shim)
	fmt.Fprintf(&transcript, "== missing flags\nexit: %d\nstdout:\n%sstderr:\n%sdolt argv:\n%s\n", exit, stdout, stderr, readShimLog(t, shim.argvLog))

	checkGolden(t, "cli_dry_run.golden", transcript.String())
}
