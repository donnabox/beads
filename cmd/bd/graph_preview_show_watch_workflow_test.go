//go:build cgo && !windows

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func TestGraphPreviewShowWatchWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			initArgs := []string{"init", "--graph-mode", "link", "--prefix", "watch", "--scope-url", "https://example.invalid/watch/", "--skip-hooks", "--skip-agents", "--non-interactive", "--json"}
			if engine == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for ordinary shared-server watch qualification")
				}
				initArgs = append(initArgs, "--server", "--external", "--server-host", "127.0.0.1", "--server-port", port, "--server-user", "root")
			}
			graphPolicyCLI(t, bd, work, home, nil, "", initArgs...)
			created := graphMixedResult[graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "create", "Initial watch title", "--id", "beads/work", "--json"))
			nativeID := created.Properties.ID
			if !strings.HasPrefix(nativeID, "watch-") || len(nativeID) < len("watch-")+3 {
				t.Fatalf("fixture lacks a native Issue hash: %q", nativeID)
			}
			hash := strings.TrimPrefix(nativeID, "watch-")
			abbreviated := hash[:len(hash)-1]
			graphPolicyCLI(t, bd, work, home, nil, "", "remember", "Memory body", "--id", "beads/memory", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "not_found", "show", "beads/missing", "--watch", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "show", "beads/memory", "--watch", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "show", "beads/work", "--watch", "--version", "old", "--json")

			logDir := t.TempDir()
			outPath, errPath := filepath.Join(logDir, "stdout"), filepath.Join(logDir, "stderr")
			out, err := os.Create(outPath)
			if err != nil {
				t.Fatal(err)
			}
			defer out.Close()
			stderr, err := os.Create(errPath)
			if err != nil {
				t.Fatal(err)
			}
			defer stderr.Close()
			watch := exec.Command(bd, "show", abbreviated, "--watch")
			watch.Dir = work
			watch.Env = []string{
				"PATH=" + os.Getenv("PATH"), "HOME=" + home,
				"TMPDIR=" + os.TempDir(), "TMP=" + os.TempDir(), "TEMP=" + os.TempDir(),
				"XDG_CONFIG_HOME=" + filepath.Join(home, ".config"),
				"XDG_CACHE_HOME=" + filepath.Join(home, ".cache"),
				"GIT_CONFIG_GLOBAL=" + filepath.Join(home, "missing-gitconfig"), "GIT_CONFIG_NOSYSTEM=1",
				"BD_DISABLE_METRICS=1", "BD_DISABLE_EVENT_FLUSH=1", "BD_NON_INTERACTIVE=1",
				"BEADS_DOLT_AUTO_START=0", "DOLT_METRICS_DISABLED=1", "NO_COLOR=1",
			}
			if developerDir := os.Getenv("DEVELOPER_DIR"); developerDir != "" {
				watch.Env = append(watch.Env, "DEVELOPER_DIR="+developerDir)
			}
			watch.Stdout, watch.Stderr = out, stderr
			if err := watch.Start(); err != nil {
				t.Fatal(err)
			}
			finished := false
			t.Cleanup(func() {
				if !finished {
					_ = watch.Process.Kill()
					_ = watch.Wait()
				}
			})
			waitFor := func(path, want string) {
				t.Helper()
				deadline := time.Now().Add(20 * time.Second)
				for time.Now().Before(deadline) {
					data, err := os.ReadFile(path)
					if err == nil && strings.Contains(string(data), want) {
						return
					}
					time.Sleep(100 * time.Millisecond)
				}
				stdout, _ := os.ReadFile(outPath)
				stderr, _ := os.ReadFile(errPath)
				t.Fatalf("watch did not emit %q; stdout=%s stderr=%s", want, stdout, stderr)
			}
			waitFor(outPath, "Initial watch title")
			waitFor(errPath, "Watching for changes")
			graphPolicyCLI(t, bd, work, home, nil, "", "update", "beads/work", "--title", "Revised watch title", "--json")
			waitFor(outPath, "Revised watch title")
			if err := watch.Process.Signal(os.Interrupt); err != nil {
				t.Fatal(err)
			}
			if err := watch.Wait(); err != nil {
				t.Fatalf("watch did not stop cleanly: %v", err)
			}
			finished = true
		})
	}
}
