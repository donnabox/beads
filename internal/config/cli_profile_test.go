package config

import (
	"os"
	"path/filepath"
	"testing"
)

func isolateCLIProfileConfig(t *testing.T) string {
	t.Helper()
	base := t.TempDir()
	t.Setenv("HOME", filepath.Join(base, "home"))
	t.Setenv("USERPROFILE", filepath.Join(base, "home"))
	t.Setenv("home", filepath.Join(base, "home"))
	t.Setenv("APPDATA", filepath.Join(base, "native"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(base, "native"))
	t.Setenv("BEADS_DIR", "")
	path, err := UserConfigYamlPath()
	if err != nil {
		t.Fatal(err)
	}
	return path
}

func writeCLIProfileConfig(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestCLIProfileUserGlobalPersistenceAndRepair(t *testing.T) {
	path := isolateCLIProfileConfig(t)
	t.Chdir(t.TempDir()) // Deliberately outside an initialized workspace.
	if !IsYamlOnlyKey(CLIProfileKey) || !IsUserGlobalKey(CLIProfileKey) {
		t.Fatal("profile would route to workspace/database configuration")
	}
	if value, present, err := UserYamlValueStrict(CLIProfileKey); err != nil || present || value != "" {
		t.Fatalf("missing = %q, %v, %v", value, present, err)
	}
	if err := SetUserYamlConfig(CLIProfileKey, "canonical"); err != nil {
		t.Fatal(err)
	}
	if value, present, err := UserYamlValueStrict(CLIProfileKey); err != nil || !present || value != "canonical" {
		t.Fatalf("saved = %q, %v, %v", value, present, err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := SetUserYamlConfig(CLIProfileKey, "unknown"); err == nil {
		t.Fatal("invalid profile was written")
	}
	after, err := os.ReadFile(path)
	if err != nil || string(before) != string(after) {
		t.Fatalf("invalid write changed settings: %v", err)
	}
	writeCLIProfileConfig(t, path, "actor: keep\ncli.profile: unknown\n")
	if value, present, err := UserYamlValueStrict(CLIProfileKey); err != nil || !present || value != "unknown" {
		t.Fatalf("raw invalid value cannot be inspected: %q, %v, %v", value, present, err)
	}
	if err := SetUserYamlConfig(CLIProfileKey, "compat"); err != nil {
		t.Fatal(err)
	}
	if value, present, err := UserYamlValueStrict(CLIProfileKey); err != nil || !present || value != "compat" {
		t.Fatalf("repaired = %q, %v, %v", value, present, err)
	}
	if changed, err := UnsetUserYamlConfig(CLIProfileKey); err != nil || !changed {
		t.Fatalf("unset = %v, %v", changed, err)
	}
	if value, present, err := UserYamlValueStrict(CLIProfileKey); err != nil || present || value != "" {
		t.Fatalf("cleared = %q, %v, %v", value, present, err)
	}
	if actor, present, err := UserYamlValueStrict("actor"); err != nil || !present || actor != "keep" {
		t.Fatalf("unrelated config lost: %q, %v, %v", actor, present, err)
	}
}

func TestCLIProfileIgnoresWorkspaceConfiguration(t *testing.T) {
	path := isolateCLIProfileConfig(t)
	writeCLIProfileConfig(t, path, "cli:\n  profile: canonical\n")
	for _, value := range []string{"compat", "unknown"} {
		workspace := t.TempDir()
		beadsDir := filepath.Join(workspace, ".beads")
		writeCLIProfileConfig(t, filepath.Join(beadsDir, "config.yaml"), "cli:\n  profile: "+value+"\n")
		writeCLIProfileConfig(t, filepath.Join(beadsDir, "config.local.yaml"), "cli.profile: "+value+"\n")
		t.Setenv("BEADS_DIR", beadsDir)
		t.Chdir(workspace)
		if got, present, err := UserYamlValueStrict(CLIProfileKey); err != nil || !present || got != "canonical" {
			t.Fatalf("workspace changed global preference: %q, %v, %v", got, present, err)
		}
	}
}

func TestCLIProfileStrictReadErrorsAreNotAbsence(t *testing.T) {
	path := isolateCLIProfileConfig(t)
	for _, contents := range []string{"cli: [", "- not-a-mapping\n", "cli.profile: null\n", "cli.profile: [canonical]\n", "cli: scalar\n"} {
		writeCLIProfileConfig(t, path, contents)
		if _, _, err := UserYamlValueStrict(CLIProfileKey); err == nil {
			t.Fatalf("invalid YAML/shape silently treated as absent: %q", contents)
		}
		before, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if contents == "cli: [" {
			if err := SetUserYamlConfig(CLIProfileKey, "compat"); err == nil {
				t.Fatal("malformed YAML was silently overwritten")
			}
			after, err := os.ReadFile(path)
			if err != nil || string(before) != string(after) {
				t.Fatalf("failed repair changed malformed YAML: %v", err)
			}
		}
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(path, 0o700); err != nil {
		t.Fatal(err)
	}
	if _, _, err := UserYamlValueStrict(CLIProfileKey); err == nil {
		t.Fatal("unreadable config target silently treated as absent")
	}
}
