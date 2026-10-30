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

func TestCLIProfileUserGlobalFreshNestedUnsetRoundTrip(t *testing.T) {
	path := isolateCLIProfileConfig(t)
	for cycle := 0; cycle < 2; cycle++ {
		if err := SetUserYamlConfig(CLIProfileKey, "canonical"); err != nil {
			t.Fatal(err)
		}
		if value, present, err := UserYamlValueStrict(CLIProfileKey); err != nil || !present || value != "canonical" {
			t.Fatalf("cycle %d saved = %q, %v, %v", cycle, value, present, err)
		}
		if changed, err := UnsetUserYamlConfig(CLIProfileKey); err != nil || !changed {
			t.Fatalf("cycle %d unset = %v, %v", cycle, changed, err)
		}
		if value, present, err := UserYamlValueStrict(CLIProfileKey); err != nil || present || value != "" {
			t.Fatalf("cycle %d cleared = %q, %v, %v", cycle, value, present, err)
		}
		before, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if changed, err := UnsetUserYamlConfig(CLIProfileKey); err != nil || changed {
			t.Fatalf("cycle %d repeated unset = %v, %v", cycle, changed, err)
		}
		after, err := os.ReadFile(path)
		if err != nil || string(after) != string(before) {
			t.Fatalf("cycle %d no-op changed bytes: %v", cycle, err)
		}
	}
}

func TestCLIProfileUserGlobalUnsetPreservesConfiguration(t *testing.T) {
	for _, tc := range []struct {
		name    string
		seed    string
		want    string
		changed bool
	}{
		{name: "last child and comments", seed: "# user settings\nactor: keep\ncli: # namespace\n  # note\n  profile: canonical # selected\n# end\n", want: "# user settings\nactor: keep\n# cli: # namespace\n  # note\n  # profile: canonical # selected\n# end\n", changed: true},
		{name: "active sibling", seed: "actor: keep\ncli: # namespace\n  option: keep\n  profile: canonical\n", want: "actor: keep\ncli: # namespace\n  option: keep\n  # profile: canonical\n", changed: true},
		{name: "other nested profile", seed: "other:\n  cli:\n    profile: keep\ncli:\n  profile: canonical\n", want: "other:\n  cli:\n    profile: keep\n# cli:\n  # profile: canonical\n", changed: true},
		{name: "commented sibling", seed: "cli:\n  # option: keep\n  profile: canonical\n\n", want: "# cli:\n  # option: keep\n  # profile: canonical\n\n", changed: true},
		{name: "unreferenced anchor", seed: "actor: keep\ncli: &prefs\n  profile: canonical\n", want: "actor: keep\n# cli: &prefs\n  # profile: canonical\n", changed: true},
		{name: "CRLF write keeps established normalization", seed: "actor: keep\r\ncli:\r\n  profile: canonical\r\n", want: "actor: keep\n# cli:\n  # profile: canonical\n", changed: true},
		{name: "absent CRLF stays byte identical", seed: "cli:\r\n  option: keep\r\n", want: "cli:\r\n  option: keep\r\n"},
		{name: "no final newline", seed: "cli:\n  profile: canonical", want: "# cli:\n  # profile: canonical", changed: true},
		{name: "unrelated NaN with present profile", seed: "score: .nan\ncli:\n  profile: canonical\n", want: "score: .nan\n# cli:\n  # profile: canonical\n", changed: true},
		{name: "unrelated NaN with absent profile", seed: "score: .nan\nactor: keep\n", want: "score: .nan\nactor: keep\n"},
		{name: "unrelated alias", seed: "defaults: &prefs\n  option: keep\ncopy: *prefs\ncli:\n  profile: canonical\n", want: "defaults: &prefs\n  option: keep\ncopy: *prefs\n# cli:\n  # profile: canonical\n", changed: true},
		{name: "unrelated repeated anchor", seed: "first: &saved 1.0\nsecond: &saved 1\ncopy: *saved\ncli:\n  profile: canonical\n", want: "first: &saved 1.0\nsecond: &saved 1\ncopy: *saved\n# cli:\n  # profile: canonical\n", changed: true},
		{name: "repair unaliased collection profile", seed: "cli:\n  profile: {old: true}\nactor: keep\n", want: "# cli:\n  # profile: {old: true}\nactor: keep\n", changed: true},
		{name: "flat and nested spellings", seed: "cli.profile: compat\ncli:\n  profile: canonical\n", want: "# cli.profile: compat\n# cli:\n  # profile: canonical\n", changed: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := isolateCLIProfileConfig(t)
			writeCLIProfileConfig(t, path, tc.seed)
			if changed, err := UnsetUserYamlConfig(CLIProfileKey); err != nil || changed != tc.changed {
				t.Fatalf("unset = %v, %v; want changed %v", changed, err, tc.changed)
			}
			got, err := os.ReadFile(path)
			if err != nil || string(got) != tc.want {
				t.Fatalf("configuration = %q, %v; want %q", got, err, tc.want)
			}
			if value, present, err := UserYamlValueStrict(CLIProfileKey); err != nil || present || value != "" {
				t.Fatalf("cleared = %q, %v, %v", value, present, err)
			}
		})
	}
}

func TestCLIProfileUserGlobalUnsetRefusesUnsafeBindings(t *testing.T) {
	for _, tc := range []struct {
		name string
		seed string
	}{
		{name: "last-child parent alias", seed: "cli: &prefs\n  profile: canonical\ncopy: *prefs\n"},
		{name: "parent alias with sibling", seed: "cli: &prefs\n  profile: canonical\n  option: keep\ncopy: *prefs\n"},
		{name: "profile scalar alias", seed: "cli:\n  profile: &saved canonical\ncopy: *saved\n"},
		{name: "aliased parent", seed: "defaults: &prefs\n  profile: canonical\ncli: *prefs\n"},
		{name: "flow parent", seed: "cli: {profile: canonical}\n"},
		{name: "quoted parent not edited by line walker", seed: "\"cli\":\n  profile: canonical\n"},
		{name: "rebound nested scalar alias", seed: "seed: &saved 1.0\ncli:\n  profile: &saved 1\ncopy: *saved\n"},
		{name: "rebound flat scalar alias", seed: "seed: &saved 1.0\ncli.profile: &saved 1\ncopy: *saved\n"},
		{name: "rebound parent alias", seed: "seed: &prefs\n  profile: 1.0\ncli: &prefs\n  profile: 1\ncopy: *prefs\n"},
		{name: "rebound nested descendant alias", seed: "seed: &saved 1.0\ncli:\n  profile: {old: &saved 1}\ncopy: *saved\n"},
		{name: "rebound flat descendant alias", seed: "seed: &saved 1.0\ncli.profile: {old: &saved 1}\ncopy: *saved\n"},
		{name: "malformed with matching leaf", seed: "cli:\n  profile: canonical\nbroken: [\n"},
		{name: "duplicate parent", seed: "cli:\n  profile: canonical\ncli:\n  option: keep\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := isolateCLIProfileConfig(t)
			writeCLIProfileConfig(t, path, tc.seed)
			if changed, err := UnsetUserYamlConfig(CLIProfileKey); err == nil || changed {
				t.Fatalf("unsafe unset = %v, %v", changed, err)
			}
			got, err := os.ReadFile(path)
			if err != nil || string(got) != tc.seed {
				t.Fatalf("refused unset changed configuration: %q, %v", got, err)
			}
		})
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
	for _, contents := range []string{"cli: [", "- not-a-mapping\n", "cli.profile: null\n", "cli.profile: [canonical]\n", "cli: scalar\n", "cli: null\n", "cli:\n  # profile: canonical\n", "cli: [canonical]\n"} {
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
