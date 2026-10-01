//go:build cgo

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/configfile"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/templates/agents"
)

func TestGraphPreviewClaudeStopAutoInit(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			args := []string{"init", "--graph-mode", "link", "--scope-url", "https://example.invalid/auto-stop/", "--non-interactive", "--json"}
			if engine == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for ordinary shared-server CLI qualification")
				}
				args = append(args, "--server", "--external", "--server-host", "127.0.0.1", "--server-port", port, "--server-user", "root")
			}
			graphPolicyCLI(t, bd, work, home, nil, "", args...)
			graphPolicyCLI(t, bd, work, home, nil, "", "setup", "claude", "--check")
			settings, err := os.ReadFile(filepath.Join(work, ".claude", "settings.json"))
			if err != nil || bytes.Count(settings, []byte("bd claude-hook stop")) != 1 {
				t.Fatalf("init did not register one Stop hook: %v %s", err, settings)
			}
			claude, err := os.ReadFile(filepath.Join(work, "CLAUDE.md"))
			if err != nil || string(claude) != "@AGENTS.md\n" {
				t.Fatalf("init did not import graph guidance: %v %q", err, claude)
			}
		})
	}
}

func TestGraphPreviewClaudeStopAutoInitPreflight(t *testing.T) {
	bd := buildBDUnderTest(t)
	work, home := t.TempDir(), t.TempDir()
	settings := filepath.Join(work, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(settings), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(settings, []byte(`{"hooks":{"Stop":"not an array"}}`), 0600); err != nil {
		t.Fatal(err)
	}
	graphPolicyCLI(t, bd, work, home, nil, "graph_not_initialized", "init", "--graph-mode", "link", "--scope-url", "https://example.invalid/auto-stop/", "--non-interactive", "--json")
	if _, err := os.Stat(filepath.Join(work, ".beads")); !os.IsNotExist(err) {
		t.Fatalf("preflight created graph storage: %v", err)
	}
	graphPolicyCLI(t, bd, work, home, nil, "", "init", "--graph-mode", "link", "--scope-url", "https://example.invalid/auto-stop/", "--skip-hooks", "--non-interactive", "--json")
	raw, err := os.ReadFile(settings)
	if err != nil || string(raw) != `{"hooks":{"Stop":"not an array"}}` {
		t.Fatalf("skip-hooks changed existing settings: %v %q", err, raw)
	}
}

// This invokes the actual installed Stop registration and Steph's handler; it
// proves delivery and execution, not that an agent chooses useful facts to save.
func TestGraphPreviewClaudeStopWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			write := func(path, content string) {
				t.Helper()
				if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(content), 0600); err != nil {
					t.Fatal(err)
				}
			}
			read := func(path string) []byte {
				t.Helper()
				raw, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				return raw
			}
			write(filepath.Join(work, "AGENTS.md"), "User instructions remain.\n")
			args := []string{"init", "--graph-mode", "link", "--scope-url", "https://example.invalid/stop/", "--skip-hooks", "--non-interactive", "--json"}
			if engine == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for ordinary shared-server CLI qualification")
				}
				args = append(args, "--server", "--external", "--server-host", "127.0.0.1", "--server-port", port, "--server-user", "root")
			}
			graphPolicyCLI(t, bd, work, home, nil, "", args...)
			settings := filepath.Join(work, ".claude", "settings.json")
			if _, err := os.Stat(settings); !os.IsNotExist(err) {
				t.Fatal("init implicitly installed hook")
			}
			beforeAgents := read(filepath.Join(work, "AGENTS.md"))
			if _, err := os.Stat(filepath.Join(work, "CLAUDE.md")); !os.IsNotExist(err) {
				t.Fatal("init implicitly installed Claude import")
			}
			write(settings, `{"permissions":{"allow":["Read"]},"hooks":{"SessionStart":[{"hooks":[{"type":"command","command":"user-start"}]}],"Stop":[{"hooks":[{"type":"command","command":"user-stop"}]}]}}`)
			call := func(args ...string) string { t.Helper(); return graphPolicyCLI(t, bd, work, home, nil, "", args...) }
			call("setup", "claude")
			beforeClaude := read(filepath.Join(work, "CLAUDE.md"))
			if string(beforeClaude) != "@AGENTS.md\n" {
				t.Fatalf("fresh setup did not load graph guidance: %q", beforeClaude)
			}
			installed := read(settings)
			if info, err := os.Stat(settings); err != nil || info.Mode().Perm() != 0600 {
				t.Fatalf("setup widened settings permissions: %v %v", info, err)
			}
			call("setup", "Claude", "--project")
			call("setup", "claude", "--check")
			if !bytes.Equal(installed, read(settings)) || !bytes.Equal(beforeAgents, read(filepath.Join(work, "AGENTS.md"))) || !bytes.Equal(beforeClaude, read(filepath.Join(work, "CLAUDE.md"))) {
				t.Fatal("setup changed user instructions or was not idempotent")
			}
			if bytes.Contains(installed, []byte("bd prime")) || !bytes.Contains(installed, []byte("user-start")) || !bytes.Contains(installed, []byte("user-stop")) || !bytes.Contains(installed, []byte("permissions")) {
				t.Fatal("setup lost siblings or installed prime")
			}
			var configured struct {
				Hooks map[string][]struct {
					Hooks []struct{ Type, Command string }
				}
			}
			if err := json.Unmarshal(installed, &configured); err != nil {
				t.Fatal(err)
			}
			var command []string
			for _, entry := range configured.Hooks["Stop"] {
				for _, hook := range entry.Hooks {
					if hook.Command == "bd claude-hook stop" && hook.Type == "command" {
						if command != nil {
							t.Fatal("duplicate Stop registration")
						}
						command = strings.Fields(hook.Command)
					}
				}
			}
			if len(command) != 3 {
				t.Fatal("missing actual Stop command")
			}
			// Seed a record, then make the database unavailable during the hook.
			original := call("remember", "Existing fact", "--json")
			memory := graphMixedResult[graphstore.Record](t, original)
			transcript := filepath.Join(work, "session.jsonl")
			write(transcript, "{\"type\":\"assistant\",\"message\":{\"content\":[{\"type\":\"tool_use\",\"id\":\"tool-1\",\"name\":\"Read\",\"input\":{}}]}}\n")
			beadsDir := filepath.Join(work, ".beads")
			metadata := filepath.Join(beadsDir, "metadata.json")
			originalMetadata := read(metadata)
			var hookEnv []string
			if engine == "server" {
				cfg, err := configfile.Load(beadsDir)
				if err != nil {
					t.Fatal(err)
				}
				cfg.DoltServerPort = 1
				if err := cfg.Save(beadsDir); err != nil {
					t.Fatal(err)
				}
				hookEnv = []string{"BEADS_DOLT_SERVER_PORT=1"}
			} else {
				if err := os.Rename(filepath.Join(beadsDir, "embeddeddolt"), filepath.Join(beadsDir, "embeddeddolt-unavailable")); err != nil {
					t.Fatal(err)
				}
			}
			restored := false
			restore := func() {
				if restored {
					return
				}
				restored = true
				if engine == "embedded" {
					if err := os.Rename(filepath.Join(beadsDir, "embeddeddolt-unavailable"), filepath.Join(beadsDir, "embeddeddolt")); err != nil {
						t.Fatal(err)
					}
				}
				if err := os.WriteFile(metadata, originalMetadata, 0600); err != nil {
					t.Fatal(err)
				}
			}
			t.Cleanup(restore)
			runStop := func(session string, active bool) string {
				t.Helper()
				input, _ := json.Marshal(claudeHookInput{SessionID: session, TranscriptPath: transcript, CWD: work, HookEventName: "Stop", StopHookActive: active})
				ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
				defer cancel()
				process := graphMemoryReadCommand(ctx, bd, work, home, command[1:]...)
				process.Env = append(process.Env, hookEnv...)
				process.Stdin = bytes.NewReader(input)
				var stdout, stderr bytes.Buffer
				process.Stdout, process.Stderr = &stdout, &stderr
				if err := process.Run(); err != nil || ctx.Err() != nil || stderr.Len() != 0 {
					t.Fatalf("store-free Stop failed: %v timeout=%v stderr=%s", err, ctx.Err(), stderr.String())
				}
				return stdout.String()
			}
			var decision claudeStopDecision
			if err := json.Unmarshal([]byte(runStop("first", false)), &decision); err != nil || decision.Decision != "block" || decision.Reason != claudeStopMemoryReminder {
				t.Fatalf("missing original reminder: %+v %v", decision, err)
			}
			if got := runStop("first", false); got != "" {
				t.Fatalf("repeated Stop reminded again: %s", got)
			}
			if got := runStop("reentrant", true); got != "" {
				t.Fatalf("reentrant Stop reminded: %s", got)
			}
			restore()
			if call("show", memory.ID, "--json") != original {
				t.Fatal("hook changed stored Memory")
			}
			remembered := graphMixedResult[graphstore.Record](t, call("remember", "Keep integration work on integration", "--json"))
			graphMemoryReadRaw(t, bd, work, home, "Keep integration work on integration", "", "recall", remembered.ID)
			for _, flags := range [][]string{{"--global"}, {"--stealth"}, {"--print"}, {"--check", "--remove"}, {"--project=false"}} {
				code := "capability_unavailable"
				if flags[0] == "--check" || flags[0] == "--project=false" {
					code = "invalid_selector"
				}
				graphPolicyCLI(t, bd, work, home, nil, code, append([]string{"setup", "claude", "--json"}, flags...)...)
			}
			graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "setup", "claude", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "setup", "codex", "--json")
			call("setup", "claude", "--remove")
			removed := read(settings)
			graphSetupRefusal(t, bd, work, home, "not installed", "setup", "claude", "--check")
			call("setup", "claude", "--remove")
			if !bytes.Equal(beforeClaude, read(filepath.Join(work, "CLAUDE.md"))) {
				t.Fatal("remove changed guidance import")
			}
			if bytes.Contains(removed, []byte("bd claude-hook stop")) || !bytes.Equal(removed, read(settings)) || !bytes.Contains(removed, []byte("user-stop")) || !bytes.Equal(beforeAgents, read(filepath.Join(work, "AGENTS.md"))) {
				t.Fatal("remove lost user content or was not idempotent")
			}
			// Existing user content gets one import, then survives install/remove.
			write(filepath.Join(work, "CLAUDE.md"), "User Claude instructions remain.")
			if err := os.Chmod(filepath.Join(work, "CLAUDE.md"), 0600); err != nil {
				t.Fatal(err)
			}
			call("setup", "claude")
			if info, err := os.Stat(filepath.Join(work, "CLAUDE.md")); err != nil || info.Mode().Perm() != 0600 {
				t.Fatalf("setup widened CLAUDE permissions: %v %v", info, err)
			}
			wantClaude := "User Claude instructions remain.\n\n@AGENTS.md\n"
			if got := string(read(filepath.Join(work, "CLAUDE.md"))); got != wantClaude {
				t.Fatalf("existing instructions %q", got)
			}
			call("setup", "claude")
			call("setup", "claude", "--check")
			call("setup", "claude", "--remove")
			if got := string(read(filepath.Join(work, "CLAUDE.md"))); got != wantClaude {
				t.Fatal("repeat install/remove changed user instructions")
			}
		})
	}
}

func graphSetupRefusal(t *testing.T, bd, work, home, fragment string, args ...string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	cmd := graphMemoryReadCommand(ctx, bd, work, home, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	var exit *exec.ExitError
	if ctx.Err() != nil || !errors.As(err, &exit) || exit.ExitCode() != 5 || stdout.Len() != 0 || !strings.HasPrefix(stderr.String(), "capability_unavailable:") || !strings.Contains(stderr.String(), fragment) {
		t.Fatalf("setup refusal %v: err=%v timeout=%v stdout=%q stderr=%q", args, err, ctx.Err(), stdout.String(), stderr.String())
	}
}

func TestGraphPreviewClaudeStopProfileRefusal(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			args := []string{"init", "--graph-mode", "link", "--scope-url", "https://example.invalid/stop-profile/", "--skip-hooks", "--skip-agents", "--non-interactive", "--json"}
			if engine == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for ordinary shared-server CLI qualification")
				}
				args = append(args, "--server", "--external", "--server-host", "127.0.0.1", "--server-port", port, "--server-user", "root")
			}
			graphPolicyCLI(t, bd, work, home, nil, "", args...)
			settings, claude := filepath.Join(work, ".claude", "settings.json"), filepath.Join(work, "CLAUDE.md")
			if err := os.MkdirAll(filepath.Dir(settings), 0755); err != nil {
				t.Fatal(err)
			}
			const userSettings = `{"user_setting":true}`
			const userClaude = "User instructions\n@AGENTS.md\n"
			if err := os.WriteFile(settings, []byte(userSettings), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(claude, []byte(userClaude), 0600); err != nil {
				t.Fatal(err)
			}
			stale := strings.Replace(agents.RenderSection(agents.ProfileGraphPreview), "hash:"+agents.CurrentHash(agents.ProfileGraphPreview), "hash:00000000", 1)
			for _, tc := range []struct{ name, content, fragment string }{
				{"missing", "", "requires existing graph-preview"},
				{"minimal", agents.RenderSection(agents.ProfileMinimal), "requires existing graph-preview"},
				{"stale", stale, "stale managed hash"},
			} {
				t.Run(tc.name, func(t *testing.T) {
					path := filepath.Join(work, "AGENTS.md")
					if tc.content != "" {
						if err := os.WriteFile(path, []byte(tc.content), 0600); err != nil {
							t.Fatal(err)
						}
					}
					for _, suffix := range [][]string{nil, {"--check"}} {
						graphSetupRefusal(t, bd, work, home, tc.fragment, append([]string{"setup", "claude"}, suffix...)...)
						for path, want := range map[string]string{settings: userSettings, claude: userClaude} {
							got, err := os.ReadFile(path)
							if err != nil || string(got) != want {
								t.Fatalf("refusal changed %s: %q %v", path, got, err)
							}
						}
						got, err := os.ReadFile(path)
						if tc.content == "" {
							if !os.IsNotExist(err) {
								t.Fatalf("refusal created missing instructions: %v", err)
							}
						} else if err != nil || string(got) != tc.content {
							t.Fatal("refusal refreshed or changed instructions")
						}
					}
				})
			}
		})
	}
}
