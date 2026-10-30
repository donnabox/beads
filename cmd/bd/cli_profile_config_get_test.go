package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/config"
)

func TestConfigGetCLIProfileDistinguishesSavedEmpty(t *testing.T) {
	base := t.TempDir()
	t.Setenv("HOME", filepath.Join(base, "home"))
	t.Setenv("USERPROFILE", filepath.Join(base, "home"))
	t.Setenv("home", filepath.Join(base, "home"))
	t.Setenv("APPDATA", filepath.Join(base, "native"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(base, "native"))
	t.Setenv("BEADS_DIR", "")
	t.Chdir(t.TempDir())
	path, err := config.UserConfigYamlPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	oldJSON := jsonOutput
	t.Cleanup(func() { jsonOutput = oldJSON })
	for _, saved := range []bool{false, true} {
		contents := "actor: keep\n"
		if saved {
			contents += "cli.profile: \"\"\n"
		}
		if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
		jsonOutput = false
		out := captureStdout(t, func() error {
			return configGetCmd.RunE(configGetCmd, []string{config.CLIProfileKey})
		})
		if saved {
			if strings.Contains(out, "not set") || !strings.Contains(out, "saved empty value") || !strings.Contains(out, `= ""`) {
				t.Fatalf("saved empty inspection = %q", out)
			}
		} else if !strings.Contains(out, "not set") {
			t.Fatalf("absent inspection = %q", out)
		}
		jsonOutput = true
		out = captureStdout(t, func() error {
			return configGetCmd.RunE(configGetCmd, []string{config.CLIProfileKey})
		})
		var payload map[string]interface{}
		if err := json.Unmarshal([]byte(out), &payload); err != nil {
			t.Fatal(err)
		}
		if present, ok := payload["present"].(bool); !ok || present != saved || payload["value"] != "" {
			t.Fatalf("JSON presence = %#v; want present=%v, empty value", payload, saved)
		}
	}
	// Ordinary user-global keys retain their existing output shape.
	jsonOutput = true
	out := captureStdout(t, func() error {
		return configGetCmd.RunE(configGetCmd, []string{"metrics.endpoint"})
	})
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatal(err)
	}
	if _, added := payload["present"]; added {
		t.Fatal("profile-specific presence field leaked to another setting")
	}
}
