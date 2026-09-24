//go:build cgo

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// These controls use the real command boundary: in-process flag checks miss
// configuration loading and the early graph dispatch that caused the bypass.
func TestGraphPreviewCLIWritePolicy(t *testing.T) {
	bd := buildBDUnderTest(t)
	initArgs := []string{"init", "--graph-mode", "link", "--scope-url", "https://example.invalid/policy/", "--skip-hooks", "--skip-agents", "--non-interactive", "--json"}
	type policyCase struct {
		name  string
		flags []string
		env   []string
		setup func(*testing.T, string, string) func()
		code  string
	}
	configFile := func(workspace bool) func(*testing.T, string, string) func() {
		return func(t *testing.T, work, home string) func() {
			path := filepath.Join(home, ".config", "bd", "config.yaml")
			if workspace {
				path = filepath.Join(work, ".beads", "config.yaml")
			}
			writeFile(t, path, []byte("readonly: true\n"))
			return func() {
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
	freeze := func(t *testing.T, work, _ string) func() {
		writeFile(t, filepath.Join(work, "mayor", "town.json"), []byte("{}\n"))
		path := filepath.Join(work, "MIGRATION-FREEZE")
		writeFile(t, path, []byte("policy-test\t2026-09-24T00:00:00Z\tgraph write control\n"))
		return func() {
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
		}
	}
	policies := []policyCase{
		{name: "readonly-flag", flags: []string{"--readonly"}, code: "permission_denied"},
		{name: "readonly-env", env: []string{"BD_READONLY=true"}, code: "permission_denied"},
		{name: "readonly-config", setup: configFile(false), code: "permission_denied"},
		{name: "migration-freeze", setup: freeze, code: "permission_denied"},
	}
	initCases := append([]policyCase(nil), policies...)
	for _, flag := range []struct{ name, value string }{
		{"server-host", "127.0.0.1"}, {"server-port", "1"}, {"server-user", "root"},
		{"server-socket", "/missing/policy.sock"}, {"server-tls", "true"},
	} {
		initCases = append(initCases, policyCase{name: "orphan-" + flag.name, flags: []string{"--" + flag.name + "=" + flag.value}, code: "invalid_selector"})
	}
	for _, tc := range initCases {
		t.Run("init/"+tc.name, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			if tc.setup != nil {
				defer tc.setup(t, work, home)()
			}
			before := legacyUpgradeTreeDigest(t, work)
			graphPolicyCLI(t, bd, work, home, tc.env, tc.code, append(append([]string(nil), initArgs...), tc.flags...)...)
			if after := legacyUpgradeTreeDigest(t, work); after != before {
				t.Fatal("refused init changed workspace")
			}
			if _, err := os.Lstat(filepath.Join(work, ".beads")); !os.IsNotExist(err) {
				t.Fatalf("refused init created .beads: %v", err)
			}
		})
	}

	// One normal embedded init supplies the store. No SQL seeding or fixture
	// schema is used; each refused write is checked from a new CLI process.
	t.Run("remember", func(t *testing.T) {
		work, home := t.TempDir(), t.TempDir()
		graphPolicyCLI(t, bd, work, home, nil, "", initArgs...)
		policies[2].setup = configFile(true)
		for _, tc := range policies {
			t.Run(tc.name, func(t *testing.T) {
				cleanup := func() {}
				if tc.setup != nil {
					cleanup = tc.setup(t, work, home)
				}
				defer cleanup()
				path := "beads/" + tc.name
				args := []string{"remember", "must not persist", "--id", path, "--title", "Refused", "--json"}
				before := legacyUpgradeTreeDigest(t, work)
				graphPolicyCLI(t, bd, work, home, tc.env, tc.code, append(args, tc.flags...)...)
				if after := legacyUpgradeTreeDigest(t, work); after != before {
					t.Fatal("refused remember changed workspace")
				}
				graphPolicyCLI(t, bd, work, home, nil, "not_found", "show", path, "--json")
			})
		}
	})
}

func TestGraphPreviewCorruptMetadataRefusesBeforeLegacyOpening(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, marker := range []bool{false, true} {
		name := "explicit-assertion"
		if marker {
			name = "persisted-marker"
		}
		t.Run(name, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			beadsDir := filepath.Join(work, ".beads")
			writeFile(t, filepath.Join(beadsDir, "metadata.json"), []byte("{\n"))
			args := []string{"show", "beads/plan", "--json"}
			if marker {
				writeFile(t, filepath.Join(beadsDir, graphPreviewMarker), []byte(graphPreviewGeneration))
			} else {
				args = append(args, "--graph-mode", "link")
			}
			before := legacyUpgradeTreeDigest(t, work)
			graphPolicyCLI(t, bd, work, home, nil, "graph_not_initialized", args...)
			if after := legacyUpgradeTreeDigest(t, work); after != before {
				t.Fatal("corrupt graph metadata refusal changed workspace")
			}
		})
	}
}

func TestGraphPreviewInitQuiet(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, jsonMode := range []bool{false, true} {
		name := "human"
		if jsonMode {
			name = "json"
		}
		t.Run(name, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			args := []string{"init", "--graph-mode", "link", "--scope-url", "https://example.invalid/quiet/", "--skip-hooks", "--skip-agents", "--non-interactive", "--quiet"}
			if jsonMode {
				args = append(args, "--json")
			}
			out := graphPolicyCLI(t, bd, work, home, nil, "", args...)
			if jsonMode {
				if !json.Valid([]byte(out)) {
					t.Fatalf("--quiet suppressed structured output: %q", out)
				}
			} else if out != "" {
				t.Fatalf("--quiet emitted human output: %q", out)
			}
			graphPolicyCLI(t, bd, work, home, nil, "", "remember", "Persisted after quiet init", "--id", "beads/quiet", "--title", "Quiet", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "", "show", "beads/quiet", "--json")
		})
	}
}

func TestGraphPreviewRejectsUnsupportedBackend(t *testing.T) {
	bd := buildBDUnderTest(t)
	work, home := t.TempDir(), t.TempDir()
	graphPolicyCLI(t, bd, work, home, nil, "", "init", "--graph-mode", "link", "--scope-url", "https://example.invalid/backend/", "--skip-hooks", "--skip-agents", "--non-interactive", "--json")
	metadataPath := filepath.Join(work, ".beads", "metadata.json")
	original, err := os.ReadFile(metadataPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, backend := range []string{"postgres", "mysql", "sqlite", "unsupported-preview"} {
		t.Run(backend, func(t *testing.T) {
			var metadata map[string]any
			if err := json.Unmarshal(original, &metadata); err != nil {
				t.Fatal(err)
			}
			metadata["backend"] = backend
			modified, err := json.Marshal(metadata)
			if err != nil {
				t.Fatal(err)
			}
			writeFile(t, metadataPath, modified)
			before := legacyUpgradeTreeDigest(t, work)
			path := "beads/refused-" + backend
			for _, args := range [][]string{
				{"remember", "must not persist", "--id", path, "--title", "Refused", "--json"},
				{"show", path, "--json"}, {"status", "--graph", "--json"},
			} {
				graphPolicyCLI(t, bd, work, home, nil, "graph_not_initialized", args...)
			}
			if after := legacyUpgradeTreeDigest(t, work); after != before {
				t.Fatal("unsupported backend refusal changed workspace")
			}
			writeFile(t, metadataPath, original)
			graphPolicyCLI(t, bd, work, home, nil, "not_found", "show", path, "--json")
		})
	}
}

func graphPolicyCLI(t *testing.T, bd, work, home string, extraEnv []string, code string, args ...string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bd, args...)
	cmd.Dir = work
	// A whitelist prevents the operator's selected workspace, routing, policy,
	// credentials and telemetry settings from entering these disposable runs.
	cmd.Env = []string{
		"PATH=" + os.Getenv("PATH"), "HOME=" + home,
		"TMPDIR=" + os.TempDir(), "TMP=" + os.TempDir(), "TEMP=" + os.TempDir(),
		"XDG_CONFIG_HOME=" + filepath.Join(home, ".config"),
		"XDG_CACHE_HOME=" + filepath.Join(home, ".cache"),
		"GIT_CONFIG_GLOBAL=" + filepath.Join(home, "missing-gitconfig"), "GIT_CONFIG_NOSYSTEM=1",
		"BD_DISABLE_METRICS=1", "BD_DISABLE_EVENT_FLUSH=1", "BD_NON_INTERACTIVE=1",
		"BEADS_DOLT_AUTO_START=0", "DOLT_METRICS_DISABLED=1", "NO_COLOR=1",
	}
	if developerDir := os.Getenv("DEVELOPER_DIR"); developerDir != "" {
		cmd.Env = append(cmd.Env, "DEVELOPER_DIR="+developerDir)
	}
	cmd.Env = append(cmd.Env, extraEnv...)
	var out, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &stderr
	err := cmd.Run()
	if ctx.Err() != nil {
		t.Fatalf("CLI did not complete within deadline: %v\n%s", ctx.Err(), stderr.String())
	}
	if code == "" {
		if err != nil {
			t.Fatalf("graph command %v failed: %v\n%s", args, err, stderr.String())
		}
		return out.String()
	}
	if err == nil {
		t.Fatalf("expected %s refusal; stdout=%s", code, out.String())
	}
	if out.Len() != 0 {
		t.Fatalf("refusal emitted success stdout: %s", out.String())
	}
	var diagnostic struct {
		Code      string `json:"code"`
		Retryable bool   `json:"retryable"`
	}
	if err := json.Unmarshal(stderr.Bytes(), &diagnostic); err != nil {
		t.Fatalf("expected typed refusal: %v\n%s", err, stderr.String())
	}
	if diagnostic.Code != code || diagnostic.Retryable {
		t.Fatalf("expected non-retryable %s; stderr=%s", code, stderr.String())
	}
	return out.String()
}
