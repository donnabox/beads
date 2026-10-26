package scripts_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/steveyegge/beads/scripts/internal/testbash"
)

func TestUpgradeSmokeMultiVersionDispatch(t *testing.T) {
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, candidate, failVersion string
		wantExit                     int
	}{
		{"prebuilt", "candidate with spaces/bd", "", 0},
		{"automatic", "", "", 0},
		{"continues after failure", "candidate with spaces/bd", "v0.61.0", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			const scriptPath = "scripts/upgrade-smoke-test.sh"
			const candidateHelper = "scripts/lib/smoke-candidate.sh"
			script, err := os.ReadFile(filepath.Join(sourceRepoRoot(t), scriptPath))
			if err != nil {
				t.Fatal(err)
			}
			shebang, body, ok := strings.Cut(string(script), "\n")
			if !ok || !strings.HasPrefix(shebang, "#!") {
				t.Fatal("upgrade smoke entrypoint must start with a shebang")
			}
			// Observe the child's script body, where Bash 3.2 has assigned positional arguments.
			// The actual dispatch body below this insertion stays byte-for-byte unchanged.
			const observer = `if [ -z "${BEADS_LOOP_ROOT:-}" ]; then
    export BEADS_LOOP_ROOT=1
else
    if [ -z "${1:-}" ]; then
        printf 'child version is empty (BASH_VERSION=%s)\n' "${BASH_VERSION:-unknown}" >&2
        exit 43
    fi
    printf '%s\t%s\t%s\n' "$1" "${CANDIDATE_BIN:-}" "${SMOKE_VERSIONS:-}" >> calls
    if [ -n "${SMOKE_VERSIONS:-}" ]; then
        printf 'child inherited loop control (BASH_VERSION=%s)\n' "${BASH_VERSION:-unknown}" >&2
        exit 42
    fi
    [ "$1" != "${BEADS_LOOP_FAIL:-}" ] || exit 19
    exit 0
fi
`
			script = []byte(shebang + "\n" + observer + body)
			for _, path := range []string{scriptPath, ".buildflags", candidateHelper} {
				data := script
				mode := os.FileMode(0644)
				if path == scriptPath {
					mode = 0755
				} else {
					data, err = os.ReadFile(filepath.Join(sourceRepoRoot(t), path))
					if err != nil {
						// Standalone sources predate the helper; a composed reference requires it.
						if path == candidateHelper && os.IsNotExist(err) && !strings.Contains(body, "smoke-candidate.sh") {
							continue
						}
						t.Fatal(err)
					}
				}
				destination := filepath.Join(dir, path)
				if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(destination, data, mode); err != nil {
					t.Fatal(err)
				}
			}
			path := os.Getenv("PATH")
			if runtime.GOOS == "windows" {
				path = "/usr/bin:/bin"
			}
			// Relative fixture paths use the same mount under both Bash invocations.
			// HOME stays absent to stop a child missing the observer before smoke setup.
			env := []string{"PATH=" + path, "LC_ALL=C", "LANG=C", "ENV=", "BASH_ENV=",
				"BEADS_LOOP_FAIL=" + tc.failVersion, "CANDIDATE_BIN=" + tc.candidate,
				"SMOKE_VERSIONS=v0.62.0 v0.61.0 v0.60.0"}
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, bash, "--noprofile", "--norc", "scripts/upgrade-smoke-test.sh")
			cmd.Dir, cmd.Env = dir, env
			out, err := cmd.CombinedOutput()
			if ctx.Err() != nil || cmd.ProcessState == nil || cmd.ProcessState.ExitCode() != tc.wantExit {
				t.Fatalf("dispatch exit = %v, want %d: %s", err, tc.wantExit, out)
			}
			got, err := os.ReadFile(filepath.Join(dir, "calls"))
			if err != nil {
				t.Fatalf("read child dispatch: %v; output: %s", err, out)
			}
			var want strings.Builder
			for _, version := range []string{"v0.62.0", "v0.61.0", "v0.60.0"} {
				want.WriteString(version + "\t" + tc.candidate + "\t\n")
			}
			if string(got) != want.String() {
				t.Fatalf("child dispatch = %q, want %q; output: %s", got, want.String(), out)
			}
		})
	}
}

// Exercise the production consent predicate directly, without starting a
// database or replacing a bd binary. Guidance may include arbitrarily long
// diagnostics; finding it must not depend on an early-exiting pipeline.
func TestUpgradeSmokeConsentGuidance(t *testing.T) {
	bash, err := testbash.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	script, err := os.ReadFile(filepath.Join(sourceRepoRoot(t), "scripts/upgrade-smoke-test.sh"))
	if err != nil {
		t.Fatal(err)
	}
	entry := strings.SplitN(string(script), "\n", 3)
	if len(entry) < 3 || entry[0] != "#!/bin/bash" || entry[1] != "set -euo pipefail" {
		t.Fatal("consent predicate harness requires the production Bash/pipefail entrypoint")
	}
	before, _, ok := strings.Cut(string(script), "; then\n        pass \"Candidate refuses un-migrated database with consent guidance\"")
	start := strings.LastIndex(before, "\n    elif ")
	if !ok || start < 0 {
		t.Fatal("cannot locate production migration-consent predicate")
	}
	predicate := before[start+len("\n    elif "):]
	// Verbatim guidance captured by the failed PR262 hosted smoke test.
	const capturedGuidance = `This database is at schema v23; this bd release uses schema v72.
bd no longer upgrades a database automatically (see the release notes).

  To upgrade this database (one-way; older bd releases cannot read it):
      bd migrate schema

  To keep the current schema, keep using the bd release that created it.

  Scripted/CI consent: BD_ALLOW_MIGRATE=1 bd <command>
`
	path := os.Getenv("PATH")
	if runtime.GOOS == "windows" {
		path = "/usr/bin:/bin"
	}
	for _, tc := range []struct {
		name, output string
		wantExit     int
		withoutTools bool
		missingFile  bool
	}{
		{"PR262 captured guidance", capturedGuidance, 0, false, false},
		{"guidance before long diagnostics", "bd migrate schema\n" + strings.Repeat("diagnostic context\n", 100000), 0, false, false},
		{"guidance after long diagnostics", strings.Repeat("diagnostic context\n", 100000) + "bd migrate schema\n", 0, false, false},
		{"empty", "", 3, false, false},
		{"unrelated failure", "Error: database unavailable", 3, false, false},
		{"consent variable alone", "BD_ALLOW_MIGRATE=1", 3, false, false},
		{"incomplete command", "bd migrate", 3, false, false},
		{"literal backslash before guidance", "diagnostic \\c\nbd migrate schema\n", 0, false, false},
		{"guidance without grep on PATH", capturedGuidance, 0, true, false},
		{"missing fixture is not a no-match", "", 90, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := filepath.Join(t.TempDir(), "output")
			if !tc.missingFile {
				if err := os.WriteFile(fixture, []byte(tc.output), 0600); err != nil {
					t.Fatal(err)
				}
			}
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			body := entry[1] + "\n[[ -f \"$1\" && -r \"$1\" ]] || exit 90\nGATE_OUT=$(<\"$1\") || exit 91\nif " + predicate + "; then exit 0; else exit 3; fi\n"
			cmd := exec.CommandContext(ctx, bash, "--noprofile", "--norc", "-c", body, "consent-predicate", msysPath(fixture))
			testPath := path
			if tc.withoutTools {
				testPath = msysPath(t.TempDir())
			}
			cmd.Env = []string{"PATH=" + testPath, "LC_ALL=C", "LANG=C", "ENV=", "BASH_ENV="}
			out, err := cmd.CombinedOutput()
			code := -1
			if cmd.ProcessState != nil {
				code = cmd.ProcessState.ExitCode()
			}
			if ctx.Err() != nil || code != tc.wantExit {
				t.Fatalf("consent predicate exit = %d (%v), want %d: %s", code, err, tc.wantExit, out)
			}
		})
	}
}
