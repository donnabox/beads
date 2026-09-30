package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	tokX = "0123456789abcdef0123456789abcdef"
	tokY = "fedcba9876543210fedcba9876543210"
	tokZ = "aaaaaaaaaaaaaaaabbbbbbbbbbbbbbbb"
)

func stepOf(name string, argv []string, exit int, stdout, stderr string) *stepResult {
	return &stepResult{Name: name, Argv: argv, Exit: exit, Stdout: []byte(stdout), Stderr: []byte(stderr)}
}

func createResult(tok string) string {
	return `{"preview":true,"result":{"attribution":{"actor":"a","recordedAt":"2026-09-29T23:52:48.472631238Z","status":"claimed"},"id":"https://example.invalid/scenarios/beads/m1","revision":"` + tok + `","version":"` + tok + `"},"schemaVersion":1}`
}

func updateResult(newTok, oldTok string) string {
	return `{"preview":true,"result":{"changed":true,"memory":{"id":"https://example.invalid/scenarios/beads/m1","revision":"` + newTok + `","version":"` + newTok + `"},"replaced":{"id":"https://example.invalid/scenarios/beads/m1","version":"` + oldTok + `"}},"schemaVersion":1}`
}

// normalizeAll runs steps through one fresh normalizer, in order, the way a
// workspace's transcript is normalized.
func normalizeAll(roots []string, steps ...*stepResult) []*stepResult {
	n := newNormalizer(roots...)
	out := make([]*stepResult, len(steps))
	for i, s := range steps {
		out[i] = n.normalize(s, nil)
	}
	return out
}

func TestNormalizeTokenOrdinalsByFirstAppearance(t *testing.T) {
	out := normalizeAll(nil,
		stepOf("create", []string{"remember", "--id", "beads/m1"}, 0, createResult(tokX), ""),
		stepOf("update", []string{"remember", "--update", "beads/m1", "--if-revision", tokX}, 0, updateResult(tokY, tokX), ""),
		stepOf("conflict", []string{"remember", "--update", "beads/m1", "--if-revision", tokX, "--json"}, 4, "",
			`{"code":"revision_conflict","message":"graph transaction conflicted: Memory revision changed (current `+tokY+`)","retryable":false}`),
	)
	create, update, conflict := out[0], out[1], out[2]

	if got := strings.Count(string(create.Stdout), "<TOKEN#1>"); got != 2 {
		t.Errorf("create stdout has %d <TOKEN#1>, want 2 (revision and version):\n%s", got, create.Stdout)
	}
	if got := update.Argv[len(update.Argv)-1]; got != "<TOKEN#1>" {
		t.Errorf("update --if-revision argv = %q, want <TOKEN#1>: the same token keeps the same placeholder in argv", got)
	}
	if !strings.Contains(string(update.Stdout), "<TOKEN#2>") || strings.Count(string(update.Stdout), "<TOKEN#1>") != 1 {
		t.Errorf("update stdout should hold <TOKEN#2> for the new version and <TOKEN#1> for the replaced one:\n%s", update.Stdout)
	}
	if !strings.Contains(string(conflict.Stderr), "(current <TOKEN#2>)") {
		t.Errorf("conflict stderr = %q, want the embedded current revision mapped to <TOKEN#2>", conflict.Stderr)
	}
	for _, s := range out {
		all := strings.Join(s.Argv, " ") + string(s.Stdout) + string(s.Stderr)
		for _, raw := range []string{tokX, tokY} {
			if strings.Contains(all, raw) {
				t.Errorf("step %s still contains raw token %s", s.Name, raw)
			}
		}
	}
}

// The reason this normalizer is not a blanket mask: two distinct versions
// collapsing into one is exactly the bug a scenario driver exists to catch.
func TestNormalizeStructureCollapseNegative(t *testing.T) {
	distinct := normalizeAll(nil,
		stepOf("a", nil, 0, createResult(tokX), ""),
		stepOf("b", nil, 0, createResult(tokY), ""),
	)
	collapsed := normalizeAll(nil,
		stepOf("a", nil, 0, createResult(tokX), ""),
		stepOf("b", nil, 0, createResult(tokX), ""),
	)
	if transcriptDigest(distinct) == transcriptDigest(collapsed) {
		t.Fatal("a transcript with two distinct tokens compares equal to one where they collapsed into one")
	}
	b := string(distinct[1].Stdout)
	if !strings.Contains(b, "<TOKEN#2>") {
		t.Errorf("second distinct token should be <TOKEN#2>, got:\n%s", b)
	}
	for _, s := range append(distinct, collapsed...) {
		for _, blanket := range []string{"<VERSION>", "<REVISION>"} {
			if strings.Contains(string(s.Stdout), blanket) {
				t.Errorf("step %s carries the blanket placeholder %s, which hides collapsed versions:\n%s", s.Name, blanket, s.Stdout)
			}
		}
	}
	// Equal tokens keep equal placeholders.
	if strings.Count(string(collapsed[1].Stdout), "<TOKEN#1>") != 2 {
		t.Errorf("a repeated token must keep its placeholder:\n%s", collapsed[1].Stdout)
	}
}

func TestNormalizeMasksTimestampsAndSortsObjectArraysByID(t *testing.T) {
	stdout := `{"result":{"items":[{"id":"b","attribution":{"recordedAt":"2026-09-29T23:52:48.472631238Z"}},{"id":"a","attribution":{"recordedAt":"2026-09-30T00:00:00Z"}}]}}`
	out := normalizeAll(nil, stepOf("list", nil, 0, stdout, ""))[0]
	text := string(out.Stdout)
	if strings.Count(text, "<TS>") != 2 || strings.Contains(text, "2026-09-") {
		t.Errorf("timestamps not masked:\n%s", text)
	}
	if strings.Index(text, `"a"`) > strings.Index(text, `"b"`) {
		t.Errorf("object array not id-sorted:\n%s", text)
	}
	var v any
	if err := json.Unmarshal(out.Stdout, &v); err != nil {
		t.Errorf("normalized stdout is not valid JSON: %v\n%s", err, text)
	}
}

func TestNormalizeDropsCommit(t *testing.T) {
	out := normalizeAll(nil, stepOf("v", nil, 0, `{"commit":"abc123","x":1}`, ""))[0]
	if strings.Contains(string(out.Stdout), "commit") {
		t.Errorf("commit should be dropped like protocol.CanonicalizeJSON does:\n%s", out.Stdout)
	}
}

func TestNormalizeWorkspaceRoot(t *testing.T) {
	real := t.TempDir()
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(real, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	n := newNormalizer(link)
	out := n.normalize(stepOf("s", []string{"cwd", link + "/work/x"}, 0,
		`{"path":"`+real+`/work/y"}`, real+"/home is not writable\n"), nil)
	if out.Argv[1] != "<WS>/work/x" {
		t.Errorf("argv = %q, want the symlinked spelling of the root masked", out.Argv[1])
	}
	if !strings.Contains(string(out.Stdout), "<WS>/work/y") {
		t.Errorf("stdout = %s, want the resolved spelling of the root masked", out.Stdout)
	}
	if string(out.Stderr) != "<WS>/home is not writable\n" {
		t.Errorf("stderr = %q", out.Stderr)
	}
}

func TestNormalizeKeepsRawStdoutBytes(t *testing.T) {
	for _, body := range []string{"first body\nsecond line\n", "no trailing newline", "  leading space\n", ""} {
		out := normalizeAll(nil, stepOf("recall", nil, 0, body, ""))[0]
		if string(out.Stdout) != body {
			t.Errorf("raw stdout %q changed to %q", body, out.Stdout)
		}
	}
}

func TestNormalizeReplacesKnownTokensInRawText(t *testing.T) {
	n := newNormalizer()
	n.normalize(stepOf("create", nil, 0, createResult(tokX), ""), nil)
	out := n.normalize(stepOf("recall", nil, 0, "body mentions "+tokX+" verbatim\n", ""), nil)
	if string(out.Stdout) != "body mentions <TOKEN#1> verbatim\n" {
		t.Errorf("stdout = %q", out.Stdout)
	}
}

func TestNormalizeCapturedValuesByKind(t *testing.T) {
	n := newNormalizer()
	out := n.normalize(stepOf("dep", []string{"dep", "add"}, 0, `{"result":{"linkId":"lnk-9f3a","ref":"`+tokZ+`"}}`, ""),
		[]capturedValue{{Kind: kindID, Value: "lnk-9f3a"}, {Kind: kindToken, Value: tokZ}})
	text := string(out.Stdout)
	if !strings.Contains(text, "<ID#1>") || !strings.Contains(text, "<TOKEN#1>") {
		t.Errorf("captured id and token should get independent ordinals:\n%s", text)
	}
	later := n.normalize(stepOf("use", []string{"dep", "rm", "lnk-9f3a"}, 0, "", ""), nil)
	if later.Argv[2] != "<ID#1>" {
		t.Errorf("captured id in a later argv = %q, want <ID#1>", later.Argv[2])
	}
}

// Two workspaces get different tokens, timestamps and roots; they must
// normalize to the same bytes, and a real behavioural difference must not.
func TestNormalizeTwoWorkspacesAgree(t *testing.T) {
	build := func(root, t1, t2, ts string) []*stepResult {
		return normalizeAll([]string{root},
			stepOf("create", []string{"remember", "--id", "beads/m1"}, 0,
				strings.Replace(createResult(t1), "2026-09-29T23:52:48.472631238Z", ts, 1), ""),
			stepOf("update", []string{"remember", "--if-revision", t1}, 0, updateResult(t2, t1), ""),
			stepOf("cwd", []string{"cwd"}, 0, root+"/work\n", ""),
		)
	}
	a := build(t.TempDir(), tokX, tokY, "2026-09-29T23:52:48.1Z")
	b := build(t.TempDir(), tokZ, "11111111111111112222222222222222", "2027-01-01T00:00:00Z")
	if transcriptDigest(a) != transcriptDigest(b) {
		t.Fatalf("structurally identical workspaces normalize differently:\nA=%s\nB=%s", a[1].Stdout, b[1].Stdout)
	}
	c := normalizeAll([]string{t.TempDir()},
		stepOf("create", []string{"remember", "--id", "beads/m1"}, 0, createResult(tokX), ""),
		stepOf("update", []string{"remember", "--if-revision", tokX}, 1, updateResult(tokY, tokX), ""), // different exit
		stepOf("cwd", []string{"cwd"}, 0, "elsewhere\n", ""),
	)
	if transcriptDigest(a) == transcriptDigest(c) {
		t.Error("a real difference (exit code, stdout) was normalized away")
	}
}

func TestTranscriptDigestFraming(t *testing.T) {
	base := stepOf("s", []string{"a", "b"}, 0, "out", "err")
	same := stepOf("s", []string{"a", "b"}, 0, "out", "err")
	if stepDigest(base) != stepDigest(same) {
		t.Error("identical steps digest differently")
	}
	for name, other := range map[string]*stepResult{
		"argv boundary":          stepOf("s", []string{"a b"}, 0, "out", "err"),
		"exit code":              stepOf("s", []string{"a", "b"}, 1, "out", "err"),
		"stdout/stderr boundary": stepOf("s", []string{"a", "b"}, 0, "ou", "terr"),
		"stdout vs stderr":       stepOf("s", []string{"a", "b"}, 0, "err", "out"),
	} {
		if stepDigest(base) == stepDigest(other) {
			t.Errorf("digest ignores a change in %s", name)
		}
	}
	if transcriptDigest([]*stepResult{base}) == transcriptDigest([]*stepResult{stepOf("t", []string{"a", "b"}, 0, "out", "err")}) {
		t.Error("transcript digest ignores the step name")
	}
	if len(stepDigest(base)) != 64 {
		t.Errorf("digest %q is not a hex sha256", stepDigest(base))
	}
}

func recordedAtStep(name, stdout string) *stepResult {
	return stepOf(name, nil, 0, stdout, "")
}

func TestCheckRecordedAtOrder(t *testing.T) {
	const (
		t0 = "2026-09-29T23:00:00Z"
		t1 = "2026-09-29T23:00:01.5Z"
		t2 = "2026-09-29T23:00:02Z"
	)
	create := func(ts string) *stepResult {
		return recordedAtStep("create", `{"result":{"attribution":{"recordedAt":"`+ts+`"}}}`)
	}
	update := func(ts, replacedTS string) *stepResult {
		return recordedAtStep("update", `{"result":{"changed":true,"memory":{"attribution":{"recordedAt":"`+ts+`"}},"replaced":{"attribution":{"recordedAt":"`+replacedTS+`"}}}}`)
	}
	// forget and list report an OLD recordedAt (the record's last write), so they
	// must not take part in the ordering.
	forget := recordedAtStep("forget", `{"result":{"deleted":true,"memory":{"attribution":{"recordedAt":"`+t0+`"}}}}`)
	list := recordedAtStep("list", `{"result":{"items":[{"attribution":{"recordedAt":"1999-01-01T00:00:00Z"}}]}}`)
	noop := recordedAtStep("noop", `{"result":{"changed":false,"memory":{"attribution":{"recordedAt":"`+t0+`"}}}}`)
	raw := recordedAtStep("recall", "first body\n")

	if err := checkRecordedAtOrder([]*stepResult{create(t1), update(t2, t1), forget, list, noop, raw}); err != nil {
		t.Errorf("non-decreasing writes flagged: %v", err)
	}
	if err := checkRecordedAtOrder([]*stepResult{create(t1), update(t1, t0)}); err != nil {
		t.Errorf("equal timestamps are non-decreasing but were flagged: %v", err)
	}
	err := checkRecordedAtOrder([]*stepResult{create(t2), update(t1, t0)})
	if err == nil || !strings.Contains(err.Error(), "recordedAt") {
		t.Errorf("decreasing recordedAt not flagged: %v", err)
	}
	if err := checkRecordedAtOrder(nil); err != nil {
		t.Errorf("no steps: %v", err)
	}
}
