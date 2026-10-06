package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func envMap(env []string) map[string]string {
	m := map[string]string{}
	for _, kv := range env {
		k, v, _ := strings.Cut(kv, "=")
		m[k] = v
	}
	return m
}

func TestNewWorkspaceLayout(t *testing.T) {
	w, err := newWorkspace()
	if err != nil {
		t.Fatalf("newWorkspace: %v", err)
	}
	root := w.Root
	realRoot, _ := filepath.EvalSymlinks(root)
	realTmp, _ := filepath.EvalSymlinks(os.TempDir())
	if !strings.HasPrefix(realRoot+string(filepath.Separator), realTmp+string(filepath.Separator)) {
		t.Errorf("root %s is not under the OS temp dir %s", realRoot, realTmp)
	}
	for _, dir := range []string{w.Work, w.Home} {
		if fi, err := os.Stat(dir); err != nil || !fi.IsDir() {
			t.Errorf("%s is not a directory: %v", dir, err)
		}
		if filepath.Dir(dir) != root {
			t.Errorf("%s is not directly under the root %s", dir, root)
		}
	}
	if _, err := os.Stat(filepath.Join(root, markerName)); err != nil {
		t.Errorf("no driver marker in a fresh workspace: %v", err)
	}
	if err := w.guard(); err != nil {
		t.Errorf("guard on a fresh workspace: %v", err)
	}
	if err := w.close(); err != nil {
		t.Errorf("close: %v", err)
	}
	if _, err := os.Stat(root); !os.IsNotExist(err) {
		t.Errorf("close left %s behind (stat err: %v)", root, err)
	}
}

func TestGuardRejectsRootOutsideOSTempDir(t *testing.T) {
	w, err := newWorkspace()
	if err != nil {
		t.Fatalf("newWorkspace: %v", err)
	}
	defer w.close()
	// The workspace is under the real temp dir; point the OS temp dir elsewhere
	// and the same workspace is now "unsafe".
	t.Setenv("TMPDIR", t.TempDir())
	err = w.guard()
	if err == nil || !strings.Contains(err.Error(), "temp") {
		t.Errorf("guard = %v, want a refusal because the root is not under the OS temp dir", err)
	}
}

func TestGuardRejectsMissingMarker(t *testing.T) {
	w, err := newWorkspace()
	if err != nil {
		t.Fatalf("newWorkspace: %v", err)
	}
	defer w.close()
	if err := os.Remove(filepath.Join(w.Root, markerName)); err != nil {
		t.Fatal(err)
	}
	err = w.guard()
	if err == nil || !strings.Contains(err.Error(), "marker") {
		t.Errorf("guard = %v, want a refusal because the marker is gone", err)
	}
}

func TestGuardRejectsMarkerForAnotherRoot(t *testing.T) {
	w, err := newWorkspace()
	if err != nil {
		t.Fatalf("newWorkspace: %v", err)
	}
	defer w.close()
	// A marker copied from some other workspace must not authorize this one.
	if err := os.WriteFile(filepath.Join(w.Root, markerName), []byte("/some/other/root\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	err = w.guard()
	if err == nil || !strings.Contains(err.Error(), "marker") {
		t.Errorf("guard = %v, want a refusal because the marker names another root", err)
	}
}

func TestGuardRejectsMissingRoot(t *testing.T) {
	w, err := newWorkspace()
	if err != nil {
		t.Fatalf("newWorkspace: %v", err)
	}
	if err := os.RemoveAll(w.Root); err != nil {
		t.Fatal(err)
	}
	if err := w.guard(); err == nil {
		t.Error("guard accepted a workspace whose root no longer exists")
	}
}

func TestHermeticEnvIsAWhitelist(t *testing.T) {
	w := &workspace{Root: "/tmp/scn-x", Work: "/tmp/scn-x/work", Home: "/tmp/scn-x/home"}
	// Everything a parent might plausibly hand down. None of it may arrive.
	for _, kv := range []string{
		"PATH=/weird/bin", "HOME=/root", "TZ=America/Los_Angeles",
		"BEADS_DIR=/leak", "BEADS_DB=/leak/db", "BD_FOO=1", "BD_DISABLE_METRICS=0",
		"BEADS_ACTOR=evil", "BEADS_DOLT_AUTO_START=1", "BEADS_DOLT_SERVER_PORT=3307",
		"GC_RIG=beads", "GIT_DIR=/x", "AWS_SECRET_ACCESS_KEY=hunter2",
	} {
		k, v, _ := strings.Cut(kv, "=")
		t.Setenv(k, v)
	}
	env := hermeticEnv(w)
	want := map[string]string{
		"HOME":                          "/tmp/scn-x/home",
		"XDG_CONFIG_HOME":               "/tmp/scn-x/home/.config",
		"XDG_CACHE_HOME":                "/tmp/scn-x/home/.cache",
		"XDG_DATA_HOME":                 "/tmp/scn-x/home/.local/share",
		"XDG_STATE_HOME":                "/tmp/scn-x/home/.local/state",
		"PATH":                          hermeticPath,
		"TZ":                            "UTC",
		"NO_COLOR":                      "1",
		"GIT_CONFIG_NOSYSTEM":           "1",
		"GIT_CONFIG_GLOBAL":             "/tmp/scn-x/home/empty-gitconfig",
		"BEADS_DOLT_AUTO_START":         "0",
		"BD_DISABLE_METRICS":            "1",
		"BD_DISABLE_EVENT_FLUSH":        "1",
		"DOLT_METRICS_DISABLED":         "1",
		"BEADS_ACTOR":                   "scenario-actor",
		"BEADS_TEST_IGNORE_REPO_CONFIG": "1",
	}
	if got := envMap(env); !reflect.DeepEqual(got, want) {
		t.Errorf("hermetic env is not exactly the whitelist:\n got %v\nwant %v", got, want)
	}
	if err := assertHermeticEnv(env); err != nil {
		t.Errorf("assertHermeticEnv on a whitelist env: %v", err)
	}
}

func TestHermeticEnvDropsParentBeadsDir(t *testing.T) {
	t.Setenv("BEADS_DIR", "/leak/from/parent")
	t.Setenv("BEADS_DB", "/leak/from/parent/db")
	t.Setenv("BD_ANYTHING", "1")
	w := &workspace{Root: "/tmp/scn-x", Work: "/tmp/scn-x/work", Home: "/tmp/scn-x/home"}
	env := hermeticEnv(w)
	for _, k := range []string{"BEADS_DIR", "BEADS_DB", "BD_ANYTHING"} {
		if _, ok := envMap(env)[k]; ok {
			t.Errorf("parent %s reached the child environment", k)
		}
	}
	if err := assertHermeticEnv(env); err != nil {
		t.Errorf("assertHermeticEnv: %v", err)
	}
}

func TestAssertHermeticEnvTripsOnLeaks(t *testing.T) {
	w := &workspace{Root: "/tmp/scn-x", Work: "/tmp/scn-x/work", Home: "/tmp/scn-x/home"}
	clean := hermeticEnv(w)
	for _, leak := range []string{"BEADS_DIR=/x", "BEADS_DB=/x", "BD_ACTOR=x", "BEADS_DOLT_SERVER_PORT=1", "GC_RIG=beads"} {
		err := assertHermeticEnv(append(append([]string{}, clean...), leak))
		if err == nil {
			t.Errorf("assertHermeticEnv accepted %s", leak)
			continue
		}
		key, _, _ := strings.Cut(leak, "=")
		if !strings.Contains(err.Error(), key) {
			t.Errorf("error %q does not name the leaked variable %s", err, key)
		}
	}
}
