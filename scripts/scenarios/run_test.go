package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Steps and scenarios the runner tests are built from. The commands they call
// (token, echo-args, ...) belong to the fake bd in fakebd_test.go.
const (
	stepToken = `{"name":"tok","argv":["token","--json"],
		"capture":{"V":{"path":"$.result.version","kind":"token"}},
		"expect":[{"op":"exit","equals":0},{"op":"json_path","path":"$.result.version","equals_capture":"V"}]}`
	stepEcho = `{"name":"echo","argv":["echo-args","${V}"],
		"expect":[{"op":"exit","equals":0},{"op":"contains","in":"stdout","marker":"echo-args"}]}`
)

func scenarioJSON(id, state, extra string, steps ...string) string {
	return fmt.Sprintf(`{"id":%q,"title":"t","spec":["#5877 S3"],"state":%q%s,"steps":[%s]}`,
		id, state, extra, strings.Join(steps, ","))
}

func passScenario(id string) string { return scenarioJSON(id, "pass", "", stepToken, stepEcho) }

func scenarioDir(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, body := range files {
		writeFile(t, dir, name, body)
	}
	return dir
}

// scratchTmp points the OS temp dir at a fresh directory so a test can see
// exactly which workspaces the driver created and removed.
func scratchTmp(t *testing.T) string {
	t.Helper()
	d := t.TempDir()
	t.Setenv("TMPDIR", d)
	return d
}

func workspaceDirs(t *testing.T, scratch string) []string {
	t.Helper()
	entries, err := os.ReadDir(scratch)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "scenarios-") {
			names = append(names, e.Name())
		}
	}
	return names
}

type driverRun struct {
	code   int
	stdout string
	stderr string
	out    string
}

func runDriver(t *testing.T, bd, scenarios string, extra ...string) driverRun {
	t.Helper()
	out := filepath.Join(t.TempDir(), "out")
	args := append([]string{"--bd", bd, "--out", out, "--scenarios", scenarios}, extra...)
	var so, se bytes.Buffer
	code := run(args, &so, &se)
	return driverRun{code: code, stdout: so.String(), stderr: se.String(), out: out}
}

type receiptScenario struct {
	ID                string            `json:"id"`
	State             string            `json:"state"`
	Outcome           string            `json:"outcome"`
	Engine            string            `json:"engine"`
	Steps             int               `json:"steps"`
	File              string            `json:"file"`
	FileSHA256        string            `json:"file_sha256"`
	TranscriptSHA256A string            `json:"transcript_sha256_A"`
	TranscriptSHA256B string            `json:"transcript_sha256_B"`
	Equal             bool              `json:"equal"`
	DurationS         float64           `json:"duration_s"`
	Finding           string            `json:"finding"`
	Gap               string            `json:"gap"`
	Captures          map[string]string `json:"captures"`
}

type receipts struct {
	DriverVersion string `json:"driver_version"`
	Source        struct {
		Commit string `json:"commit"`
		Dirty  *bool  `json:"dirty"`
	} `json:"source"`
	BD struct {
		Path    string `json:"path"`
		SHA256  string `json:"sha256"`
		Version string `json:"version"`
	} `json:"bd"`
	Env          map[string]string `json:"env"`
	ScenariosDir string            `json:"scenarios_dir"`
	Engine       string            `json:"engine"`
	Discovered   int               `json:"discovered"`
	Executed     int               `json:"executed"`
	Scenarios    []receiptScenario `json:"scenarios"`
}

func readReceipts(t *testing.T, out string) receipts {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(out, "receipts.json"))
	if err != nil {
		t.Fatalf("receipts.json: %v", err)
	}
	var r receipts
	if err := json.Unmarshal(data, &r); err != nil {
		t.Fatalf("receipts.json is not valid JSON: %v\n%s", err, data)
	}
	return r
}

func (r receipts) scenario(t *testing.T, id string) receiptScenario {
	t.Helper()
	for _, s := range r.Scenarios {
		if s.ID == id {
			return s
		}
	}
	t.Fatalf("no receipt for scenario %s in %+v", id, r.Scenarios)
	return receiptScenario{}
}

func readFileString(t *testing.T, path ...string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(path...))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestRunAllPassExitsZero(t *testing.T) {
	bd := installFakeBD(t, fakeConfig{})
	dir := scenarioDir(t, map[string]string{"P1.json": passScenario("P1"), "P2.json": passScenario("P2")})
	r := runDriver(t, bd, dir)
	if r.code != 0 {
		t.Fatalf("exit = %d, want 0\nstdout:\n%s\nstderr:\n%s", r.code, r.stdout, r.stderr)
	}
	rc := readReceipts(t, r.out)
	if rc.Discovered != 2 || rc.Executed != 2 {
		t.Errorf("discovered=%d executed=%d, want 2 and 2", rc.Discovered, rc.Executed)
	}
	for _, id := range []string{"P1", "P2"} {
		s := rc.scenario(t, id)
		if s.Outcome != "pass" || s.State != "pass" || !s.Equal || s.Steps != 2 || s.Engine != "embedded" {
			t.Errorf("receipt for %s = %+v", id, s)
		}
		if s.TranscriptSHA256A == "" || s.TranscriptSHA256A != s.TranscriptSHA256B {
			t.Errorf("%s transcript digests A=%q B=%q, want equal and non-empty", id, s.TranscriptSHA256A, s.TranscriptSHA256B)
		}
		// transcripts/<scenario>/<A|B>/NN-<step>/{argv,exit,stdout,stderr,sha256}
		for _, ws := range []string{"A", "B"} {
			for _, f := range []string{"argv", "exit", "stdout", "stderr", "sha256"} {
				if _, err := os.Stat(filepath.Join(r.out, "transcripts", id, ws, "01-tok", f)); err != nil {
					t.Errorf("missing transcript file: %v", err)
				}
			}
		}
		a := readFileString(t, r.out, "transcripts", id, "A", "01-tok", "stdout")
		b := readFileString(t, r.out, "transcripts", id, "B", "01-tok", "stdout")
		if a != b || !strings.Contains(a, "<TOKEN#1>") {
			t.Errorf("normalized stdout differs between workspaces or is not normalized:\nA=%s\nB=%s", a, b)
		}
	}
	if !strings.Contains(r.stdout, "P1") || !strings.Contains(r.stdout, "pass") {
		t.Errorf("stdout should summarize each scenario, got:\n%s", r.stdout)
	}
}

func TestRunFailingExpectationExitsOne(t *testing.T) {
	bd := installFakeBD(t, fakeConfig{})
	failing := scenarioJSON("F1", "pass", "", `{"name":"echo","argv":["echo-args"],"expect":[{"op":"exit","equals":5}]}`)
	r := runDriver(t, bd, scenarioDir(t, map[string]string{"F1.json": failing}))
	if r.code != 1 {
		t.Fatalf("exit = %d, want 1\n%s", r.code, r.stderr)
	}
	if !strings.Contains(r.stderr, "F1") || !strings.Contains(r.stderr, "exit") {
		t.Errorf("stderr should name the scenario and the failed assertion:\n%s", r.stderr)
	}
	if got := readReceipts(t, r.out).scenario(t, "F1").Outcome; got != "fail" {
		t.Errorf("outcome = %q, want fail", got)
	}
}

func TestRunXfailThatFailsExitsZero(t *testing.T) {
	bd := installFakeBD(t, fakeConfig{})
	body := scenarioJSON("X1", "xfail", `,"finding":"be-abc12"`, `{"name":"echo","argv":["echo-args"],"expect":[{"op":"exit","equals":5}]}`)
	r := runDriver(t, bd, scenarioDir(t, map[string]string{"X1.json": body}))
	if r.code != 0 {
		t.Fatalf("exit = %d, want 0: an xfail that fails is the expected outcome\n%s", r.code, r.stderr)
	}
	s := readReceipts(t, r.out).scenario(t, "X1")
	if s.Outcome != "xfail" || s.Finding != "be-abc12" {
		t.Errorf("receipt = %+v, want outcome xfail citing be-abc12", s)
	}
}

func TestRunXpassExitsOne(t *testing.T) {
	bd := installFakeBD(t, fakeConfig{})
	body := scenarioJSON("X2", "xfail", `,"finding":"be-abc12"`, stepToken)
	r := runDriver(t, bd, scenarioDir(t, map[string]string{"X2.json": body}))
	if r.code != 1 {
		t.Fatalf("exit = %d, want 1: an xfail that passes is an unexpected pass\n%s", r.code, r.stderr)
	}
	if !strings.Contains(r.stderr, "xpass") && !strings.Contains(r.stderr, "unexpectedly passed") {
		t.Errorf("stderr should say the xfail passed:\n%s", r.stderr)
	}
	if got := readReceipts(t, r.out).scenario(t, "X2").Outcome; got != "xpass" {
		t.Errorf("outcome = %q, want xpass", got)
	}
}

// A declared skip is visible, not silent: it is recorded with its gap and
// counted as executed, so a dropped scenario still shows as discovered != executed.
func TestRunDeclaredSkipIsRecorded(t *testing.T) {
	bd := installFakeBD(t, fakeConfig{})
	body := scenarioJSON("K1", "skip", `,"gap":"be-tatfn#3"`, stepToken)
	dir := scenarioDir(t, map[string]string{"K1.json": body, "P1.json": passScenario("P1")})
	r := runDriver(t, bd, dir)
	if r.code != 0 {
		t.Fatalf("exit = %d, want 0\n%s", r.code, r.stderr)
	}
	rc := readReceipts(t, r.out)
	if rc.Discovered != 2 || rc.Executed != 2 {
		t.Errorf("discovered=%d executed=%d, want 2 and 2", rc.Discovered, rc.Executed)
	}
	s := rc.scenario(t, "K1")
	if s.Outcome != "skip" || s.Gap != "be-tatfn#3" {
		t.Errorf("receipt = %+v, want outcome skip citing be-tatfn#3", s)
	}
	if _, err := os.Stat(filepath.Join(r.out, "transcripts", "K1")); err == nil {
		t.Error("a skipped scenario must not run, but it has transcripts")
	}
}

func TestRunDivergenceExitsOne(t *testing.T) {
	bd := installFakeBD(t, fakeConfig{})
	// rand prints a fresh value each call and is not a token: nothing may hide it.
	body := scenarioJSON("D1", "pass", "", `{"name":"r","argv":["rand"],"expect":[{"op":"exit","equals":0}]}`)
	r := runDriver(t, bd, scenarioDir(t, map[string]string{"D1.json": body}))
	if r.code != 1 {
		t.Fatalf("exit = %d, want 1\n%s", r.code, r.stderr)
	}
	if !strings.Contains(r.stderr, "diverged") {
		t.Errorf("stderr should say the workspaces diverged:\n%s", r.stderr)
	}
	s := readReceipts(t, r.out).scenario(t, "D1")
	if s.Equal || s.Outcome != "diverged" || s.TranscriptSHA256A == s.TranscriptSHA256B {
		t.Errorf("receipt = %+v, want equal=false outcome=diverged and different digests", s)
	}
}

func TestRunRecordedAtOrderViolationFailsScenario(t *testing.T) {
	bd := installFakeBD(t, fakeConfig{})
	body := scenarioJSON("O1", "pass", "",
		`{"name":"late","argv":["token-at","2026-09-29T23:00:05Z"],"expect":[{"op":"exit","equals":0}]}`,
		`{"name":"early","argv":["token-at","2026-09-29T23:00:01Z"],"expect":[{"op":"exit","equals":0}]}`)
	r := runDriver(t, bd, scenarioDir(t, map[string]string{"O1.json": body}))
	if r.code != 1 || !strings.Contains(r.stderr, "recordedAt") {
		t.Errorf("exit = %d, want 1 naming recordedAt\n%s", r.code, r.stderr)
	}
}

func TestRunBDMissingExitsTwo(t *testing.T) {
	dir := scenarioDir(t, map[string]string{"P1.json": passScenario("P1")})
	for _, bd := range []string{filepath.Join(t.TempDir(), "nope"), "bd"} {
		r := runDriver(t, bd, dir)
		if r.code != 2 || !strings.Contains(r.stderr, "bd") {
			t.Errorf("--bd %q: exit = %d, want 2 naming bd\n%s", bd, r.code, r.stderr)
		}
	}
}

func TestRunGraphModeInactiveExitsTwo(t *testing.T) {
	bd := installFakeBD(t, fakeConfig{Legacy: true})
	scratch := scratchTmp(t)
	r := runDriver(t, bd, scenarioDir(t, map[string]string{"P1.json": passScenario("P1")}))
	if r.code != 2 || !strings.Contains(r.stderr, "graph mode") {
		t.Errorf("exit = %d, want 2 naming graph mode\n%s", r.code, r.stderr)
	}
	if left := workspaceDirs(t, scratch); len(left) != 0 {
		t.Errorf("workspaces left behind after a guard trip: %v", left)
	}
}

func TestRunInitFailureExitsTwo(t *testing.T) {
	bd := installFakeBD(t, fakeConfig{InitExit: 1})
	r := runDriver(t, bd, scenarioDir(t, map[string]string{"P1.json": passScenario("P1")}))
	if r.code != 2 {
		t.Errorf("exit = %d, want 2\n%s", r.code, r.stderr)
	}
}

func TestRunTimeoutExitsTwo(t *testing.T) {
	old := stepTimeout
	stepTimeout = 300 * time.Millisecond
	defer func() { stepTimeout = old }()
	bd := installFakeBD(t, fakeConfig{})
	body := scenarioJSON("T1", "pass", "", `{"name":"hang","argv":["sleep","30"],"expect":[{"op":"exit","equals":0}]}`)
	start := time.Now()
	r := runDriver(t, bd, scenarioDir(t, map[string]string{"T1.json": body}))
	if r.code != 2 || !strings.Contains(r.stderr, "timeout") {
		t.Errorf("exit = %d, want 2 naming a timeout\n%s", r.code, r.stderr)
	}
	if elapsed := time.Since(start); elapsed > 20*time.Second {
		t.Errorf("run took %s: the hung bd was not killed at the cap", elapsed)
	}
}

func TestRunManifestErrorExitsThreeBeforeTouchingBD(t *testing.T) {
	bad := mod(t, baseScenario, `"state": "pass"`, `"state": "maybe"`)
	dir := scenarioDir(t, map[string]string{"S1.json": bad})
	// bd does not exist: if the manifest were checked after bd, this would exit 2.
	r := runDriver(t, filepath.Join(t.TempDir(), "nope"), dir)
	if r.code != 3 {
		t.Errorf("exit = %d, want 3\n%s", r.code, r.stderr)
	}
	if _, err := os.Stat(r.out); err == nil {
		t.Error("--out was created although the manifest is invalid")
	}
}

func TestRunEmptyScenariosDirectoryExitsThree(t *testing.T) {
	bd := installFakeBD(t, fakeConfig{})
	r := runDriver(t, bd, t.TempDir())
	if r.code != 3 {
		t.Errorf("exit = %d, want 3: an empty scenarios directory is never a silent zero\n%s", r.code, r.stderr)
	}
}

func TestRunList(t *testing.T) {
	dir := scenarioDir(t, map[string]string{"P2.json": passScenario("P2"), "P1.json": passScenario("P1")})
	var so, se bytes.Buffer
	if code := run([]string{"--list", "--scenarios", dir}, &so, &se); code != 0 {
		t.Fatalf("exit = %d\n%s", code, se.String())
	}
	if so.String() != "P1\nP2\n" {
		t.Errorf("--list printed %q, want one sorted id per line", so.String())
	}

	so.Reset()
	se.Reset()
	bad := scenarioDir(t, map[string]string{"S1.json": mod(t, baseScenario, `"state": "pass"`, `"state": "maybe"`)})
	if code := run([]string{"--list", "--scenarios", bad}, &so, &se); code != 3 || so.Len() != 0 {
		t.Errorf("--list on a bad manifest: exit %d stdout %q, want 3 and no ids", code, so.String())
	}
	if code := run([]string{"--list", "--scenarios", t.TempDir()}, &so, &se); code != 3 {
		t.Errorf("--list on an empty directory: exit %d, want 3", code)
	}
}

func TestRunScenarioSelection(t *testing.T) {
	bd := installFakeBD(t, fakeConfig{})
	dir := scenarioDir(t, map[string]string{"P1.json": passScenario("P1"), "P2.json": passScenario("P2")})

	r := runDriver(t, bd, dir, "--scenario", "P1")
	if r.code != 0 {
		t.Fatalf("exit = %d\n%s", r.code, r.stderr)
	}
	rc := readReceipts(t, r.out)
	if rc.Discovered != 2 || rc.Executed != 1 || len(rc.Scenarios) != 1 {
		t.Errorf("--scenario P1: discovered=%d executed=%d, want 2 and 1 (discovery ignores the filter)", rc.Discovered, rc.Executed)
	}

	r = runDriver(t, bd, dir, "--scenario", "P1", "--scenario", "P2")
	if rc := readReceipts(t, r.out); r.code != 0 || rc.Executed != 2 {
		t.Errorf("repeated --scenario: exit %d executed %d, want 0 and 2", r.code, rc.Executed)
	}

	if r := runDriver(t, bd, dir, "--scenario", "ZZ"); r.code != 3 {
		t.Errorf("unknown --scenario: exit %d, want 3", r.code)
	}
}

func TestRunRejectsUnsupportedEngine(t *testing.T) {
	bd := installFakeBD(t, fakeConfig{})
	dir := scenarioDir(t, map[string]string{"P1.json": passScenario("P1")})
	if r := runDriver(t, bd, dir, "--engine", "server"); r.code != 2 {
		t.Errorf("--engine server: exit %d, want 2 (embedded only)\n%s", r.code, r.stderr)
	}
	if r := runDriver(t, bd, dir, "--engine", "embedded"); r.code != 0 {
		t.Errorf("--engine embedded: exit %d, want 0\n%s", r.code, r.stderr)
	}
}

func TestRunOutDirMustBeFresh(t *testing.T) {
	bd := installFakeBD(t, fakeConfig{})
	dir := scenarioDir(t, map[string]string{"P1.json": passScenario("P1")})
	out := t.TempDir()
	writeFile(t, out, "stale.txt", "left over from another run")
	var so, se bytes.Buffer
	if code := run([]string{"--bd", bd, "--out", out, "--scenarios", dir}, &so, &se); code != 2 {
		t.Errorf("non-empty --out: exit %d, want 2\n%s", code, se.String())
	}
	so.Reset()
	se.Reset()
	if code := run([]string{"--bd", bd, "--out", t.TempDir(), "--scenarios", dir}, &so, &se); code != 0 {
		t.Errorf("empty existing --out: exit %d, want 0\n%s", code, se.String())
	}
}

func TestRunUsageErrorsExitTwo(t *testing.T) {
	dir := scenarioDir(t, map[string]string{"P1.json": passScenario("P1")})
	for name, args := range map[string][]string{
		"unknown flag": {"--bogus"},
		"no --bd":      {"--out", t.TempDir(), "--scenarios", dir},
		"no --out":     {"--bd", "/x/bd", "--scenarios", dir},
	} {
		var so, se bytes.Buffer
		if code := run(args, &so, &se); code != 2 {
			t.Errorf("%s: exit %d, want 2", name, code)
		}
	}
}

func TestRunActorReachesArgvAsFlag(t *testing.T) {
	bd := installFakeBD(t, fakeConfig{})
	body := scenarioJSON("A1", "pass", "",
		`{"name":"as-alice","argv":["echo-args","x"],"actor":"alice","expect":[{"op":"exit","equals":0},{"op":"contains","in":"stdout","marker":"--actor"},{"op":"contains","in":"stdout","marker":"alice"}]}`,
		`{"name":"default","argv":["echo-args","x"],"expect":[{"op":"exit","equals":0},{"op":"not_contains","in":"stdout","marker":"--actor"}]}`)
	r := runDriver(t, bd, scenarioDir(t, map[string]string{"A1.json": body}))
	if r.code != 0 {
		t.Fatalf("exit = %d\n%s", r.code, r.stderr)
	}
	argv := readFileString(t, r.out, "transcripts", "A1", "A", "01-as-alice", "argv")
	if !strings.Contains(argv, `"--actor","alice"`) {
		t.Errorf("transcript argv %q should show the appended --actor alice", argv)
	}
	if strings.Contains(readFileString(t, r.out, "transcripts", "A1", "A", "02-default", "argv"), "--actor") {
		t.Error("a step without an actor must not get --actor")
	}
}

func TestRunChildEnvIsHermeticEndToEnd(t *testing.T) {
	t.Setenv("BEADS_DIR", "/leak/from/parent")
	t.Setenv("BEADS_DB", "/leak/from/parent/db")
	bd := installFakeBD(t, fakeConfig{})
	body := scenarioJSON("E1", "pass", "", `{"name":"env","argv":["echo-env"],"expect":[
		{"op":"exit","equals":0},
		{"op":"not_contains","in":"stdout","marker":"BEADS_DIR"},
		{"op":"not_contains","in":"stdout","marker":"BEADS_DB="},
		{"op":"not_contains","in":"stdout","marker":"/leak"},
		{"op":"contains","in":"stdout","marker":"BEADS_ACTOR=scenario-actor"},
		{"op":"contains","in":"stdout","marker":"BD_DISABLE_METRICS=1"},
		{"op":"contains","in":"stdout","marker":"TZ=UTC"}]}`)
	r := runDriver(t, bd, scenarioDir(t, map[string]string{"E1.json": body}))
	if r.code != 0 {
		t.Fatalf("exit = %d\n%s", r.code, r.stderr)
	}
}

func TestRunStepsRunInTheWorkspaceAndRootIsMasked(t *testing.T) {
	bd := installFakeBD(t, fakeConfig{})
	body := scenarioJSON("W1", "pass", "", `{"name":"where","argv":["cwd"],"expect":[{"op":"exit","equals":0}]}`)
	r := runDriver(t, bd, scenarioDir(t, map[string]string{"W1.json": body}))
	if r.code != 0 {
		t.Fatalf("exit = %d\n%s", r.code, r.stderr)
	}
	for _, ws := range []string{"A", "B"} {
		if got := readFileString(t, r.out, "transcripts", "W1", ws, "01-where", "stdout"); got != "<WS>/work\n" {
			t.Errorf("workspace %s: cwd normalized to %q, want <WS>/work", ws, got)
		}
	}
}

// A temp dir that is itself a symlink (macOS: /var -> /private/var) gives the
// workspace root two spellings. The driver hands the child paths in the spelling
// it was given (HOME and the XDG dirs); a child that asks the kernel where it is
// gets the resolved one (cwd). Both must become <WS>, or two workspaces stop being
// byte-comparable on any host whose temp dir is a symlink.
func TestRunMasksBothSpellingsOfTheRootWhenTheTempDirIsASymlink(t *testing.T) {
	link, real := symlinkedDir(t)
	t.Setenv("TMPDIR", link)
	bd := installFakeBD(t, fakeConfig{})
	body := scenarioJSON("W2", "pass", "",
		`{"name":"where","argv":["cwd"],"expect":[{"op":"exit","equals":0}]}`,
		`{"name":"env","argv":["echo-env"],"expect":[{"op":"exit","equals":0}]}`)
	r := runDriver(t, bd, scenarioDir(t, map[string]string{"W2.json": body}))
	if r.code != 0 {
		t.Fatalf("exit = %d\n%s", r.code, r.stderr)
	}
	for _, ws := range []string{"A", "B"} {
		if got := readFileString(t, r.out, "transcripts", "W2", ws, "01-where", "stdout"); got != "<WS>/work\n" {
			t.Errorf("workspace %s: cwd (the resolved spelling) normalized to %q, want <WS>/work", ws, got)
		}
		env := readFileString(t, r.out, "transcripts", "W2", ws, "02-env", "stdout")
		if !strings.Contains(env, "HOME=<WS>/home\n") {
			t.Errorf("workspace %s: HOME (the spelling the driver was given) not masked:\n%s", ws, env)
		}
		for _, spelling := range []string{link, real} {
			if strings.Contains(env, spelling) {
				t.Errorf("workspace %s: %s survived normalization:\n%s", ws, spelling, env)
			}
		}
	}
}

func TestRunKeepWorkspaces(t *testing.T) {
	bd := installFakeBD(t, fakeConfig{})
	dir := scenarioDir(t, map[string]string{"P1.json": passScenario("P1")})

	scratch := scratchTmp(t)
	if r := runDriver(t, bd, dir); r.code != 0 {
		t.Fatalf("exit = %d\n%s", r.code, r.stderr)
	}
	if left := workspaceDirs(t, scratch); len(left) != 0 {
		t.Errorf("workspaces not removed after the run: %v", left)
	}

	r := runDriver(t, bd, dir, "--keep-workspaces")
	if r.code != 0 {
		t.Fatalf("exit = %d\n%s", r.code, r.stderr)
	}
	kept := workspaceDirs(t, scratch)
	if len(kept) != 2 {
		t.Fatalf("--keep-workspaces kept %v, want the A and B workspaces", kept)
	}
	for _, name := range kept {
		if !strings.Contains(r.stderr, name) {
			t.Errorf("stderr should report the kept workspace %s:\n%s", name, r.stderr)
		}
	}
}

func TestRunInitCommandLine(t *testing.T) {
	bd := installFakeBD(t, fakeConfig{})
	markers := []string{"init", "--graph-mode", "link", "--scope-url", "https://example.invalid/scenarios/",
		"--prefix", "scn", "--skip-hooks", "--skip-agents", "--non-interactive"}
	var asserts []string
	for _, m := range markers {
		asserts = append(asserts, fmt.Sprintf(`{"op":"contains","in":"stdout","marker":%q}`, m))
	}
	body := scenarioJSON("I1", "pass", "", `{"name":"init-argv","argv":["show-init-argv"],"expect":[{"op":"exit","equals":0},`+strings.Join(asserts, ",")+`]}`)
	if r := runDriver(t, bd, scenarioDir(t, map[string]string{"I1.json": body})); r.code != 0 {
		t.Fatalf("exit = %d\n%s", r.code, r.stderr)
	}
}

func TestRunStdinForms(t *testing.T) {
	bd := installFakeBD(t, fakeConfig{})
	body := scenarioJSON("N1", "pass", "",
		`{"name":"text","argv":["cat"],"stdin":{"name":"B1","text":"first body\n"},"expect":[{"op":"exit","equals":0},{"op":"stdout","equals_input":"B1"}]}`,
		`{"name":"b64","argv":["cat"],"stdin":{"name":"B2","b64":"AAEC/w=="},"expect":[{"op":"exit","equals":0},{"op":"stdout","equals_input":"B2"}]}`,
		`{"name":"big","argv":["cat"],"stdin":{"name":"B3","generate":{"repeat":"ab","bytes":1048577}},"expect":[{"op":"exit","equals":0},{"op":"stdout","equals_input":"B3"},{"op":"not_contains","in":"stdout","marker":"c"}]}`)
	r := runDriver(t, bd, scenarioDir(t, map[string]string{"N1.json": body}))
	if r.code != 0 {
		t.Fatalf("exit = %d\n%s", r.code, r.stderr)
	}
	// generated stdin is deterministic and exactly N bytes: "ab" repeated, cut at 1 MiB + 1
	if got := readFileString(t, r.out, "transcripts", "N1", "A", "03-big", "stdout"); len(got) != 1048577 || !strings.HasPrefix(got, "abab") || !strings.HasSuffix(got, "aba") {
		t.Errorf("generated stdin has %d bytes, want exactly 1048577 of repeating ab", len(got))
	}
}

func TestRunStopsAtFirstFailingStep(t *testing.T) {
	bd := installFakeBD(t, fakeConfig{})
	body := scenarioJSON("S1", "pass", "",
		stepToken,
		`{"name":"bad","argv":["echo-args"],"expect":[{"op":"exit","equals":9}]}`,
		`{"name":"never","argv":["echo-args"],"expect":[{"op":"exit","equals":0}]}`)
	r := runDriver(t, bd, scenarioDir(t, map[string]string{"S1.json": body}))
	if r.code != 1 {
		t.Fatalf("exit = %d, want 1\n%s", r.code, r.stderr)
	}
	for _, ws := range []string{"A", "B"} {
		if _, err := os.Stat(filepath.Join(r.out, "transcripts", "S1", ws, "02-bad")); err != nil {
			t.Errorf("workspace %s should have run the failing step: %v", ws, err)
		}
		if _, err := os.Stat(filepath.Join(r.out, "transcripts", "S1", ws, "03-never")); err == nil {
			t.Errorf("workspace %s ran a step after the failure", ws)
		}
	}
	if s := readReceipts(t, r.out).scenario(t, "S1"); s.Steps != 2 || !s.Equal {
		t.Errorf("receipt = %+v, want 2 executed steps and equal workspaces", s)
	}
}

// Acceptance 3: two runs give byte-identical normalized transcripts.
func TestRunTranscriptsAreByteIdenticalAcrossRuns(t *testing.T) {
	bd := installFakeBD(t, fakeConfig{})
	dir := scenarioDir(t, map[string]string{"P1.json": passScenario("P1"), "P2.json": passScenario("P2")})
	snapshot := func(out string) map[string]string {
		files := map[string]string{}
		root := filepath.Join(out, "transcripts")
		err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			rel, _ := filepath.Rel(root, p)
			data, err := os.ReadFile(p)
			files[rel] = string(data)
			return err
		})
		if err != nil {
			t.Fatal(err)
		}
		return files
	}
	first := runDriver(t, bd, dir)
	second := runDriver(t, bd, dir)
	if first.code != 0 || second.code != 0 {
		t.Fatalf("exits %d and %d\n%s\n%s", first.code, second.code, first.stderr, second.stderr)
	}
	a, b := snapshot(first.out), snapshot(second.out)
	if len(a) == 0 || len(a) != len(b) {
		t.Fatalf("transcript file counts differ or are empty: %d vs %d", len(a), len(b))
	}
	for name, content := range a {
		if b[name] != content {
			t.Errorf("transcript file %s differs between two runs:\n%q\nvs\n%q", name, content, b[name])
		}
	}
}

func TestRunReceiptsRecordProvenance(t *testing.T) {
	bd := installFakeBD(t, fakeConfig{})
	stable := `{"name":"refuse","argv":["refuse-text","gone","3"],
		"capture":{"CODE":{"stderr_code":true}},
		"expect":[{"op":"exit","nonzero":true},{"op":"stdout","empty":true}]}`
	dir := scenarioDir(t, map[string]string{"P1.json": scenarioJSON("P1", "pass", "", stepToken, stable)})
	r := runDriver(t, bd, dir)
	if r.code != 0 {
		t.Fatalf("exit = %d\n%s", r.code, r.stderr)
	}
	rc := readReceipts(t, r.out)

	if rc.DriverVersion == "" || rc.Engine != "embedded" {
		t.Errorf("driver_version=%q engine=%q", rc.DriverVersion, rc.Engine)
	}
	if rc.BD.Path != bd || rc.BD.Version != "fake-bd 0.0.0" {
		t.Errorf("bd = %+v, want path %s and version %q", rc.BD, bd, "fake-bd 0.0.0")
	}
	exe, _ := os.ReadFile(bd) // follows the symlink to the test binary
	sum := sha256.Sum256(exe)
	if rc.BD.SHA256 != hex.EncodeToString(sum[:]) {
		t.Errorf("bd.sha256 = %s, want the digest of the binary", rc.BD.SHA256)
	}
	if filepath.Clean(rc.ScenariosDir) != filepath.Clean(dir) {
		t.Errorf("scenarios_dir = %q, want %q", rc.ScenariosDir, dir)
	}
	if rc.Env["TZ"] != "UTC" || rc.Env["HOME"] != "<WS>/home" || rc.Env["BEADS_ACTOR"] != "scenario-actor" {
		t.Errorf("env whitelist not recorded (workspace path masked): %v", rc.Env)
	}
	if _, leaked := rc.Env["BEADS_DIR"]; leaked {
		t.Error("BEADS_DIR appears in the recorded env")
	}
	if rc.Source.Dirty == nil && rc.Source.Commit != "" {
		t.Errorf("source = %+v: a known commit needs a known dirty flag", rc.Source)
	}
	fileSum := sha256.Sum256([]byte(scenarioJSON("P1", "pass", "", stepToken, stable)))
	s := rc.scenario(t, "P1")
	if s.FileSHA256 != hex.EncodeToString(fileSum[:]) || s.File != "P1.json" {
		t.Errorf("scenario file receipt = %s %s, want P1.json and its sha256", s.File, s.FileSHA256)
	}
	if s.DurationS < 0 {
		t.Errorf("duration_s = %v", s.DurationS)
	}
	// stable captures are receipted; per-workspace random tokens are not
	if s.Captures["CODE"] != "gone" {
		t.Errorf("captures = %v, want CODE=gone", s.Captures)
	}
	if _, ok := s.Captures["V"]; ok {
		t.Errorf("a token capture must not be receipted as a value: %v", s.Captures)
	}
}

// R1b in miniature: two refusals that share a code must fail a "differ"
// comparison, and two that do not must pass it. Plain-text refusals count.
func TestRunCompareStderrCodes(t *testing.T) {
	bd := installFakeBD(t, fakeConfig{})
	mk := func(id, second string) string {
		return scenarioJSON(id, "pass", "",
			`{"name":"first","argv":["refuse-text","gone","3"],"expect":[{"op":"exit","nonzero":true},{"op":"stdout","empty":true}]}`,
			`{"name":"second","argv":["refuse-text","`+second+`","3"],"expect":[{"op":"exit","nonzero":true},{"op":"stdout","empty":true},
				{"op":"compare","left":{"step":"first","field":"stderr_code"},"right":{"step":"second","field":"stderr_code"},"relation":"differ"}]}`)
	}
	if r := runDriver(t, bd, scenarioDir(t, map[string]string{"C1.json": mk("C1", "not_found")})); r.code != 0 {
		t.Errorf("different codes: exit %d, want 0\n%s", r.code, r.stderr)
	}
	same := runDriver(t, bd, scenarioDir(t, map[string]string{"C2.json": mk("C2", "gone")}))
	if same.code != 1 {
		t.Errorf("same code on both refusals: exit %d, want 1\n%s", same.code, same.stderr)
	}
	// declared xfail with its finding: the same refusal is then the expected outcome
	xf := scenarioJSON("C3", "xfail", `,"finding":"be-abc12"`,
		`{"name":"first","argv":["refuse-text","gone","3"],"expect":[{"op":"exit","nonzero":true}]}`,
		`{"name":"second","argv":["refuse-text","gone","3"],"expect":[{"op":"exit","nonzero":true},
			{"op":"compare","left":{"step":"first","field":"stderr_code"},"right":{"step":"second","field":"stderr_code"},"relation":"differ"}]}`)
	if r := runDriver(t, bd, scenarioDir(t, map[string]string{"C3.json": xf})); r.code != 0 {
		t.Errorf("xfail on an indistinguishable pair: exit %d, want 0\n%s", r.code, r.stderr)
	}
}
