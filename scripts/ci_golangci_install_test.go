package scripts_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestInstallGolangciLintFailsClosedOnChecksumMismatch is the regression net
// for install-golangci-lint.sh's supply-chain gate (SF-3). Without it, nothing
// exercises the sha256 comparison's behavior -- TestGolangciLintInstallScriptPinned
// above only pins the hash strings as text, so e.g. replacing the comparison
// with an always-false condition passes every other test in this package. A
// fake curl hands the script a tarball whose content never matches either
// pinned hash; a fake tar makes extraction succeed anyway (so a bypassed check
// is exposed cleanly as "installed the wrong thing", not masked by tar
// choking on bad bytes), which is exactly the gap a real supply-chain attack
// (a compromised release asset, a MITM'd download) would exploit if the
// comparison were missing or broken.
func TestInstallGolangciLintFailsClosedOnChecksumMismatch(t *testing.T) {
	bash := requireHostTool(t, "bash")

	bin := t.TempDir()
	stateDir := t.TempDir()
	runnerTemp := t.TempDir()
	curlLog := filepath.Join(stateDir, "curl-calls")
	tarLog := filepath.Join(stateDir, "tar-calls")
	githubPath := filepath.Join(stateDir, "github-path")
	for _, path := range []string{curlLog, tarLog, githubPath} {
		if err := os.WriteFile(path, nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}

	// Writes fixed, wrong-checksum bytes to curl's -o target. Any download
	// (first attempt or a retry) "succeeds" the same way, so the script
	// reaches the sha256 comparison deterministically.
	writeExecutable(t, filepath.Join(bin, "curl"), `#!/bin/sh
set -eu
printf '%s\n' "$*" >>"$GOLANGCI_INSTALL_CURL_LOG"
out=""
prev=""
for arg in "$@"; do
  if [ "$prev" = "-o" ]; then
    out="$arg"
  fi
  prev="$arg"
done
if [ -n "$out" ]; then
  printf 'not the real golangci-lint release asset\n' >"$out"
fi
`)
	// Materializes a plausible-looking extracted binary unconditionally, the
	// way a real tarball's contents would after a successful `tar -xzf`. The
	// script's sha256 check -- not tar's own format validation -- must be
	// what stands between this fake binary and $RUNNER_TEMP/golangci-lint.
	writeExecutable(t, filepath.Join(bin, "tar"), `#!/bin/sh
set -eu
printf '%s\n' "$*" >>"$GOLANGCI_INSTALL_TAR_LOG"
outdir=""
prev=""
for arg in "$@"; do
  if [ "$prev" = "-C" ]; then
    outdir="$arg"
  fi
  prev="$arg"
done
mkdir -p "$outdir/golangci-lint-2.10.1-linux-amd64"
cat >"$outdir/golangci-lint-2.10.1-linux-amd64/golangci-lint" <<'SCRIPT'
#!/bin/sh
printf 'golangci-lint has version 2.10.1 built from unknown\n'
SCRIPT
chmod +x "$outdir/golangci-lint-2.10.1-linux-amd64/golangci-lint"
`)
	// Pins the platform so the asset name, arch lookup and extracted layout
	// are fixed regardless of the machine running the test.
	writeExecutable(t, filepath.Join(bin, "uname"), `#!/bin/sh
set -eu
case "${1:-}" in
  -m) printf 'x86_64\n' ;;
  *) printf 'Linux\n' ;;
esac
`)

	pathEnv := shellPathEnv()
	binPath := shellPathUnderEnv(t, bash, bin, pathEnv)
	statePath := shellPathUnderEnv(t, bash, stateDir, pathEnv)
	commandPath := binPath + ":" + os.Getenv("PATH") + ":/usr/bin:/bin"
	if runtime.GOOS == "windows" {
		commandPath = binPath + ":/usr/bin:/bin"
	}
	root := sourceRepoRoot(t)
	env := []string{
		"PATH=" + os.Getenv("PATH"),
		"BEADS_TEST_COMMAND_PATH=" + commandPath,
		"LC_ALL=C",
		"LANG=C",
		"BASH_ENV=",
		"ENV=",
		"RUNNER_TEMP=" + shellPathUnderEnv(t, bash, runnerTemp, pathEnv),
		"GITHUB_PATH=" + statePath + "/github-path",
		"GOLANGCI_INSTALL_CURL_LOG=" + statePath + "/curl-calls",
		"GOLANGCI_INSTALL_TAR_LOG=" + statePath + "/tar-calls",
	}
	for _, name := range []string{"curl", "tar", "uname"} {
		requireShellCommandPath(t, bash, root, env, name, binPath+"/"+name)
	}

	cmd := bashScriptCommand(bash, "scripts/ci/install-golangci-lint.sh")
	cmd.Dir = root
	cmd.Env = env
	output, runErr := cmd.CombinedOutput()

	if code := pullDoltExitCode(runErr); code == 0 {
		t.Fatalf("exit = 0, want non-zero; output:\n%s", output)
	}
	if !strings.Contains(string(output), "sha256 mismatch") {
		t.Errorf("output does not contain %q:\n%s", "sha256 mismatch", output)
	}
	if calls := readCallLines(t, tarLog); len(calls) != 0 {
		t.Errorf("tar was invoked %d time(s), want 0 (the sha256 check must reject the download before extraction): %q", len(calls), calls)
	}
	if calls := readCallLines(t, curlLog); len(calls) == 0 {
		t.Fatalf("curl was never invoked; test did not exercise the script's download path")
	}
	if entries, err := os.ReadDir(runnerTemp); err != nil {
		t.Fatal(err)
	} else if len(entries) != 0 {
		t.Errorf("RUNNER_TEMP = %v, want empty (nothing installed on a checksum mismatch)", entries)
	}
	if data, err := os.ReadFile(githubPath); err != nil {
		t.Fatal(err)
	} else if len(data) != 0 {
		t.Errorf("GITHUB_PATH = %q, want empty (nothing appended on a checksum mismatch)", data)
	}
}
