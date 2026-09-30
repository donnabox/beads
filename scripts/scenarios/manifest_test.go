package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// baseScenario is a minimal valid scenario. The manifest tests derive their
// broken variants from it with mod, so each case states only what it breaks.
const baseScenario = `{
  "id": "S1",
  "title": "minimal",
  "spec": ["#5877 S3"],
  "state": "pass",
  "steps": [
    {"name": "one", "argv": ["token", "--json"],
     "capture": {"V": {"path": "$.result.version", "kind": "token"}},
     "expect": [{"op": "exit", "equals": 0}]},
    {"name": "two", "argv": ["echo-args", "${V}"],
     "expect": [{"op": "exit", "equals": 0}, {"op": "contains", "in": "stdout", "marker": "echo-args"}]}
  ]
}`

func mod(t *testing.T, s, old, repl string) string {
	t.Helper()
	if strings.Count(s, old) != 1 {
		t.Fatalf("mod: %q occurs %d times in the base scenario, want exactly 1", old, strings.Count(s, old))
	}
	return strings.Replace(s, old, repl, 1)
}

func writeFile(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func loadOne(t *testing.T, name, body string) ([]*Scenario, error) {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, dir, name, body)
	return loadScenarios(dir)
}

func wantManifestError(t *testing.T, err error, substr string) {
	t.Helper()
	if err == nil {
		t.Fatalf("loadScenarios succeeded, want a manifest error containing %q", substr)
	}
	var me *manifestError
	if !errors.As(err, &me) {
		t.Fatalf("error %v (%T) is not a *manifestError, so run() would not exit 3", err, err)
	}
	if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(substr)) {
		t.Fatalf("manifest error %q does not mention %q", err, substr)
	}
}

func TestLoadScenariosValid(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "S1.json", baseScenario)
	writeFile(t, dir, "README.md", "not a scenario")
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(dir, "sub"), "S9.json", baseScenario)

	got, err := loadScenarios(dir)
	if err != nil {
		t.Fatalf("loadScenarios: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("discovered %d scenarios, want 1: only top-level *.json files count", len(got))
	}
	sc := got[0]
	if sc.ID != "S1" || sc.State != "pass" || len(sc.Steps) != 2 || sc.Steps[0].Name != "one" {
		t.Errorf("scenario decoded wrong: %+v", sc)
	}
	sum := sha256.Sum256([]byte(baseScenario))
	if sc.SHA256 != hex.EncodeToString(sum[:]) {
		t.Errorf("SHA256 = %s, want the sha256 of the file bytes", sc.SHA256)
	}
	if filepath.Base(sc.File) != "S1.json" {
		t.Errorf("File = %q, want the scenario file path", sc.File)
	}
}

func TestLoadScenariosSortedByID(t *testing.T) {
	dir := t.TempDir()
	for _, id := range []string{"R2", "R1b", "R1"} {
		writeFile(t, dir, id+".json", strings.Replace(baseScenario, `"id": "S1"`, `"id": "`+id+`"`, 1))
	}
	got, err := loadScenarios(dir)
	if err != nil {
		t.Fatalf("loadScenarios: %v", err)
	}
	var ids []string
	for _, sc := range got {
		ids = append(ids, sc.ID)
	}
	if strings.Join(ids, ",") != "R1,R1b,R2" {
		t.Errorf("ids = %v, want R1,R1b,R2", ids)
	}
}

func TestLoadScenariosMissingOrEmptyDirectory(t *testing.T) {
	_, err := loadScenarios(filepath.Join(t.TempDir(), "does-not-exist"))
	wantManifestError(t, err, "scenarios directory")

	empty := t.TempDir()
	_, err = loadScenarios(empty)
	wantManifestError(t, err, "no scenario files")

	writeFile(t, empty, "notes.md", "only prose")
	_, err = loadScenarios(empty)
	wantManifestError(t, err, "no scenario files")
}

func TestLoadScenariosIDMustEqualFileStem(t *testing.T) {
	_, err := loadOne(t, "S2.json", baseScenario)
	wantManifestError(t, err, "must equal")
}

func TestLoadScenariosInvalidID(t *testing.T) {
	for _, id := range []string{"bad id", "-lead", ".hidden", "a/b", strings.Repeat("x", 65)} {
		body := strings.Replace(baseScenario, `"id": "S1"`, `"id": "`+id+`"`, 1)
		if strings.Contains(id, "/") {
			// a slash cannot be a file stem; it must fail on the id, not on writing the file
			dir := t.TempDir()
			writeFile(t, dir, "x.json", body)
			_, err := loadScenarios(dir)
			wantManifestError(t, err, "id")
			continue
		}
		_, err := loadOne(t, id+".json", body)
		wantManifestError(t, err, "invalid id")
	}
}

func TestLoadScenariosDuplicateIDs(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "S1.json", baseScenario)
	writeFile(t, dir, "s1.json", strings.Replace(baseScenario, `"id": "S1"`, `"id": "s1"`, 1))
	_, err := loadScenarios(dir)
	entries, rerr := os.ReadDir(dir)
	if rerr != nil {
		t.Fatal(rerr)
	}
	if len(entries) == 1 {
		// A case-insensitive filesystem (macOS APFS, Windows) folded the second
		// write into the first: there is one file, S1.json, and its body now says
		// "s1". No pair of files exists to collide, so the duplicate cannot be built
		// here (the next test builds it another way); the loader must still fail
		// closed on what is left.
		wantManifestError(t, err, "must equal the file stem")
		return
	}
	wantManifestError(t, err, "duplicate")
}

// The duplicate check compares ids, not file names, so it needs no two files whose
// names differ only by case. This is the same check on a fixture every filesystem
// can hold: the second id is a case-variant of the first, in a file of another name.
func TestLoadScenariosDuplicateIDsDifferingOnlyByCase(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "S1.json", baseScenario)
	writeFile(t, dir, "other.json", strings.Replace(baseScenario, `"id": "S1"`, `"id": "s1"`, 1))
	_, err := loadScenarios(dir)
	wantManifestError(t, err, "duplicate scenario id")
}

func TestLoadScenariosManifestErrors(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{"unknown state", mod(t, baseScenario, `"state": "pass"`, `"state": "maybe"`), "unknown state"},
		{"xfail without finding", mod(t, baseScenario, `"state": "pass"`, `"state": "xfail"`), "finding"},
		{"xfail with malformed finding", mod(t, baseScenario, `"state": "pass"`, `"state": "xfail", "finding": "not a bead"`), "finding"},
		{"skip without gap", mod(t, baseScenario, `"state": "pass"`, `"state": "skip"`), "gap"},
		{"skip with malformed gap", mod(t, baseScenario, `"state": "pass"`, `"state": "skip", "gap": "item 3"`), "gap"},
		{"finding on a passing scenario", mod(t, baseScenario, `"state": "pass"`, `"state": "pass", "finding": "be-abc12"`), "only valid"},
		{"gap on a passing scenario", mod(t, baseScenario, `"state": "pass"`, `"state": "pass", "gap": "be-tatfn#3"`), "only valid"},
		{"empty spec citation", mod(t, baseScenario, `"spec": ["#5877 S3"]`, `"spec": []`), "spec"},
		{"no steps", `{"id": "S1", "title": "t", "spec": ["#5877 S3"], "state": "pass", "steps": []}`, "steps"},
		{"unknown top-level field", mod(t, baseScenario, `"title": "minimal"`, `"title": "minimal", "extra": 1`), "unknown field"},
		{"unknown step field", mod(t, baseScenario, `"name": "one"`, `"name": "one", "shell": true`), "unknown field"},
		{"unknown assertion primitive", mod(t, baseScenario, `{"op": "contains", "in": "stdout", "marker": "echo-args"}`, `{"op": "exec", "cmd": "rm -rf /"}`), "unknown op"},
		{"unknown assertion field", mod(t, baseScenario, `{"op": "contains", "in": "stdout", "marker": "echo-args"}`, `{"op": "contains", "in": "stdout", "marker": "x", "regex": ".*"}`), "unknown field"},
		{"actor uppercase", mod(t, baseScenario, `"name": "one"`, `"name": "one", "actor": "Alice"`), "actor"},
		{"actor leading digit", mod(t, baseScenario, `"name": "one"`, `"name": "one", "actor": "1abc"`), "actor"},
		{"actor with space", mod(t, baseScenario, `"name": "one"`, `"name": "one", "actor": "a b"`), "actor"},
		{"actor too long", mod(t, baseScenario, `"name": "one"`, `"name": "one", "actor": "a`+strings.Repeat("b", 32)+`"`), "actor"},
		{"unsupported json path in an assertion", mod(t, baseScenario, `"path": "$.result.version"`, `"path": "$..version"`), "json path"},
		{"step without an exit expectation", mod(t, baseScenario, `"expect": [{"op": "exit", "equals": 0}]},`, `"expect": []},`), "exit"},
		{"step with two exit expectations", mod(t, baseScenario, `{"op": "contains", "in": "stdout", "marker": "echo-args"}`, `{"op": "exit", "nonzero": true}`), "exit"},
		{"undefined variable", mod(t, baseScenario, `"${V}"`, `"${NOPE}"`), "undefined"},
		{"variable used before its capture", mod(t, baseScenario, `["token", "--json"]`, `["token", "${V}"]`), "undefined"},
		{"unterminated variable", mod(t, baseScenario, `"${V}"`, `"${V"`), "variable"},
		{"duplicate step names", mod(t, baseScenario, `"name": "two"`, `"name": "one"`), "duplicate step"},
		{"invalid step name", mod(t, baseScenario, `"name": "two"`, `"name": "Bad Name"`), "step name"},
		{"compare against an unknown step", mod(t, baseScenario, `{"op": "contains", "in": "stdout", "marker": "echo-args"}`,
			`{"op": "compare", "left": {"step": "ghost", "field": "stdout"}, "right": {"step": "one", "field": "stdout"}, "relation": "equal"}`), "unknown step"},
		{"compare against a later step", mod(t, baseScenario, `"expect": [{"op": "exit", "equals": 0}]},`,
			`"expect": [{"op": "exit", "equals": 0}, {"op": "compare", "left": {"step": "one", "field": "stdout"}, "right": {"step": "two", "field": "stdout"}, "relation": "differ"}]},`), "unknown step"},
		{"stdout equals an undeclared input", mod(t, baseScenario, `{"op": "contains", "in": "stdout", "marker": "echo-args"}`, `{"op": "stdout", "equals_input": "NOPE"}`), "input"},
		{"generated stdin with no size", mod(t, baseScenario, `"name": "one"`, `"name": "one", "stdin": {"generate": {"repeat": "x", "bytes": 0}}`), "generate"},
		{"stdin with two forms", mod(t, baseScenario, `"name": "one"`, `"name": "one", "stdin": {"text": "a", "b64": "YQ=="}`), "stdin"},
		{"stdin with bad base64", mod(t, baseScenario, `"name": "one"`, `"name": "one", "stdin": {"b64": "!!!"}`), "stdin"},
		{"capture with unknown kind", mod(t, baseScenario, `"kind": "token"`, `"kind": "secret"`), "kind"},
		{"capture with two sources", mod(t, baseScenario, `"path": "$.result.version", "kind": "token"`, `"path": "$.result.version", "stdout": true`), "capture"},
		{"capture with no source", mod(t, baseScenario, `"path": "$.result.version", "kind": "token"`, `"kind": "token"`), "capture"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := loadOne(t, "S1.json", tc.body)
			wantManifestError(t, err, tc.want)
		})
	}
}

// One bad scenario must not hide another: a scenario author fixes everything
// in one pass rather than one error per run.
func TestLoadScenariosReportsEveryProblem(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "S1.json", mod(t, baseScenario, `"state": "pass"`, `"state": "maybe"`))
	writeFile(t, dir, "S2.json", strings.Replace(mod(t, baseScenario, `"${V}"`, `"${NOPE}"`), `"id": "S1"`, `"id": "S2"`, 1))
	_, err := loadScenarios(dir)
	wantManifestError(t, err, "unknown state")
	wantManifestError(t, err, "undefined")
}

func TestLoadScenariosStateFieldsAccepted(t *testing.T) {
	for name, body := range map[string]string{
		"xfail": mod(t, baseScenario, `"state": "pass"`, `"state": "xfail", "finding": "be-abc12"`),
		"skip":  mod(t, baseScenario, `"state": "pass"`, `"state": "skip", "gap": "be-tatfn#3"`),
	} {
		got, err := loadOne(t, "S1.json", body)
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		if got[0].State != name {
			t.Errorf("%s: State = %q", name, got[0].State)
		}
	}
}

func TestLoadScenariosAcceptsActorAndGeneratedStdin(t *testing.T) {
	body := mod(t, baseScenario, `"name": "one"`,
		`"name": "one", "actor": "alice-2", "stdin": {"name": "BIG", "generate": {"repeat": "ab", "bytes": 1048577}}`)
	got, err := loadOne(t, "S1.json", body)
	if err != nil {
		t.Fatalf("loadScenarios: %v", err)
	}
	if got[0].Steps[0].Actor != "alice-2" {
		t.Errorf("Actor = %q", got[0].Steps[0].Actor)
	}
}
