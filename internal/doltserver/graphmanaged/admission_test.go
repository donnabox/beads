package graphmanaged

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func encoded(p string) string { return base64.StdEncoding.EncodeToString([]byte(p)) }
func testRoot(t *testing.T) string {
	t.Helper()
	p, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return p
}
func mustWrite(t *testing.T, p string, contents []byte, mode os.FileMode) {
	t.Helper()
	if err := os.WriteFile(p, contents, mode); err != nil {
		t.Fatal(err)
	}
}
func descriptionFixture(t *testing.T) (description, []byte) {
	t.Helper()
	root := testRoot(t)
	source := filepath.Join(root, "config")
	exe := filepath.Join(root, "adapter")
	mustWrite(t, source, []byte("opaque engine owned bytes"), 0600)
	mustWrite(t, exe, []byte("inert executable bytes"), 0700)
	digest := func(p string) string {
		data, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		s := sha256.Sum256(data)
		return hex.EncodeToString(s[:])
	}
	d := description{Version: 1, Executable: encoded(exe), ExecutableSHA256: digest(exe), ProfileSHA256: strings.Repeat("a", 64), Cwd: encoded(root), SecurityRoots: []string{encoded(root)}, Sources: []sourceInput{{Path: encoded(source), SHA256: digest(source)}}, Databases: []databaseInput{{Name: "graph", Root: encoded(root), Branch: "main"}}, Directories: []directoryInput{{Path: encoded(root), Entries: []string{encoded("adapter"), encoded("config")}}}, Environment: []environmentInput{{Name: "TMPDIR", Value: root}}}
	raw, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	return d, raw
}
func TestAdmissionOwnedSnapshot(t *testing.T) {
	if !platformSupported() {
		t.Skip("platform intentionally unsupported")
	}
	d, raw := descriptionFixture(t)
	a, err := inspect(context.Background(), raw)
	if err != nil {
		t.Fatal(err)
	}
	for i := range raw {
		raw[i] = 'x'
	}
	d.Environment[0].Value = "changed"
	if err = a.recheck(context.Background()); err != nil {
		t.Fatal(err)
	}
	mustWrite(t, a.pins[0].path, []byte("changed"), 0600)
	if !errors.Is(a.recheck(context.Background()), errInput) {
		t.Fatal("changed source admitted")
	}
}
func TestAdmissionRefusals(t *testing.T) {
	if !platformSupported() {
		t.Skip("platform intentionally unsupported")
	}
	cases := map[string]func(*description){
		"unknown-version":   func(d *description) { d.Version = 2 },
		"profile":           func(d *description) { d.ProfileSHA256 = "wrong" },
		"digest":            func(d *description) { d.Sources[0].SHA256 = strings.Repeat("b", 64) },
		"duplicate-root":    func(d *description) { d.Databases = append(d.Databases, d.Databases[0]) },
		"physical-alias":    func(d *description) { d.Databases = append(d.Databases, d.Databases[0]); d.Databases[1].Name = "other" },
		"missing-security":  func(d *description) { d.SecurityRoots = nil },
		"missing-candidate": func(d *description) { d.Databases = nil },
		"ambient-HOME": func(d *description) {
			d.Environment = append(d.Environment, environmentInput{Name: "HOME", Value: "/tmp"})
		},
		"extra-entry":    func(d *description) { d.Directories[0].Entries = append(d.Directories[0].Entries, encoded("unknown")) },
		"relative":       func(d *description) { d.Executable = encoded("adapter") },
		"empty-branch":   func(d *description) { d.Databases[0].Branch = "" },
		"uppercase":      func(d *description) { d.Databases[0].Name = "Graph" },
		"duplicate-file": func(d *description) { d.Sources = append(d.Sources, d.Sources[0]) },
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			d, _ := descriptionFixture(t)
			change(&d)
			raw, _ := json.Marshal(d)
			if _, err := inspect(context.Background(), raw); err == nil {
				t.Fatal("accepted")
			}
		})
	}
}
func TestLexicalBoundaries(t *testing.T) {
	tests := []struct {
		name  string
		raw   string
		valid bool
	}{
		{"string-at", `"` + strings.Repeat("x", maxStringBytes) + `"`, true},
		{"string-over", `"` + strings.Repeat("x", maxStringBytes+1) + `"`, false},
		{"escaped-at", `"` + strings.Repeat(`\u0061`, maxStringBytes) + `"`, true},
		{"escaped-over", `"` + strings.Repeat(`\u0061`, maxStringBytes+1) + `"`, false},
		{"depth-at", strings.Repeat("[", maxJSONDepth) + "0" + strings.Repeat("]", maxJSONDepth), true},
		{"depth-over", strings.Repeat("[", maxJSONDepth+1) + "0" + strings.Repeat("]", maxJSONDepth+1), false},
		{"nodes-at", "[" + strings.Repeat("0,", maxNodes-2) + "0]", true},
		{"nodes-over", "[" + strings.Repeat("0,", maxNodes-1) + "0]", false},
		{"duplicate", `{"a":0,"\u0061":1}`, false}, {"trailing", `{} {}`, false},
		{"surrogate", `"\ud800"`, false}, {"low-surrogate", `"\udc00"`, false}, {"pair", `"\ud83d\ude00"`, true},
		{"invalid-utf8", "\"\xff\"", false}, {"invalid-escape", `"\q"`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := budget{}
			err := validateJSON([]byte(tt.raw), &b)
			if (err == nil) != tt.valid {
				t.Fatalf("valid=%v error=%v", tt.valid, err)
			}
		})
	}
}
func TestBudgetCheckedBoundaries(t *testing.T) {
	for name, limit := range map[string]int64{"candidates": maxCandidates, "entries": maxEntries, "files": maxFiles, "source-bytes": maxSourceBytes, "nodes": maxNodes, "strings": maxStringsBytes, "paths": maxPathsBytes, "retained": maxRetainedBytes, "environment-count": maxEnvironment, "environment-bytes": maxEnvironmentBytes} {
		t.Run(name, func(t *testing.T) {
			used := limit - 1
			if err := charge(&used, 1, limit, name); err != nil {
				t.Fatal(err)
			}
			if !errors.Is(charge(&used, 1, limit, name), errBudget) || used != limit {
				t.Fatal("one-over mutated or admitted")
			}
		})
	}
	for _, tt := range []struct{ used, n, limit int64 }{{math.MaxInt64, 1, math.MaxInt64}, {0, -1, 10}, {-1, 1, 10}, {0, 11, 10}} {
		v := tt.used
		if !errors.Is(charge(&v, tt.n, tt.limit, "overflow"), errBudget) || v != tt.used {
			t.Fatal(tt)
		}
	}
}
func TestNativePathBoundaries(t *testing.T) {
	for _, n := range []int{maxPathBytes, maxPathBytes + 1} {
		b := budget{}
		p, err := nativePath(encoded(strings.Repeat("x", n)), &b)
		if n == maxPathBytes && len(p) != n || n > maxPathBytes && !errors.Is(err, errBudget) {
			t.Fatalf("%d %v", n, err)
		}
	}
	b := budget{}
	p, err := nativePath(encoded("/native-\xff"), &b)
	if err != nil || p != "/native-\xff" {
		t.Fatal(p, err)
	}
	for _, p := range []string{"", encoded("bad\x00path"), "@@@"} {
		if _, err := nativePath(p, &budget{}); err == nil {
			t.Fatal("accepted invalid path")
		}
	}
}
func TestStreamLimits(t *testing.T) {
	for _, n := range []int{maxFileBytes, maxFileBytes + 1} {
		b := budget{}
		_, _, err := hashStream(context.Background(), time.Now().Add(time.Second), bytes.NewReader(make([]byte, n)), maxFileBytes, &b, false)
		if (err == nil) != (n == maxFileBytes) {
			t.Fatalf("%d %v", n, err)
		}
	}
	b := budget{sources: maxSourceBytes - 3}
	if _, _, err := hashStream(context.Background(), time.Now().Add(time.Second), strings.NewReader("1234"), maxFileBytes, &b, false); !errors.Is(err, errBudget) {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := hashStream(ctx, time.Now().Add(time.Second), strings.NewReader("x"), 100, &budget{}, false); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, _, err := hashStream(context.Background(), time.Now().Add(-time.Second), strings.NewReader("x"), 100, &budget{}, false); !errors.Is(err, errBudget) {
		t.Fatal(err)
	}
}
func TestFileInventoryRefusesSpecialAndAliases(t *testing.T) {
	if !platformSupported() {
		t.Skip("platform intentionally unsupported")
	}
	root := testRoot(t)
	target := filepath.Join(root, "source")
	mustWrite(t, target, []byte("x"), 0600)
	alias := filepath.Join(root, "alias")
	if err := os.Symlink(target, alias); err != nil {
		t.Fatal(err)
	}
	if _, err := decodeCanonical(context.Background(), time.Now().Add(time.Second), encoded(alias), &budget{}); err == nil {
		t.Fatal("symlink admitted")
	}
	if _, err := hashFile(context.Background(), time.Now().Add(time.Second), alias, false, &budget{}); err == nil {
		t.Fatal("nofollow failed")
	}
	if _, err := hashFile(context.Background(), time.Now().Add(time.Second), root, false, &budget{}); err == nil {
		t.Fatal("directory as regular file")
	}
	if err := os.Chmod(target, 0666); err != nil {
		t.Fatal(err)
	}
	if _, err := hashFile(context.Background(), time.Now().Add(time.Second), target, false, &budget{}); err == nil {
		t.Fatal("mutable source admitted")
	}
}
func TestDirectoryEntryLimit(t *testing.T) {
	if !platformSupported() {
		t.Skip("platform intentionally unsupported")
	}
	root := testRoot(t)
	for i := 0; i < maxEntries; i++ {
		name := filepath.Join(root, strconvName(i))
		mustWrite(t, name, nil, 0600)
	}
	got, err := inventoryDirectory(context.Background(), time.Now().Add(10*time.Second), root, &budget{})
	if err != nil || len(got.entries) != maxEntries {
		t.Fatal(len(got.entries), err)
	}
	mustWrite(t, filepath.Join(root, "one-more"), nil, 0600)
	if _, err = inventoryDirectory(context.Background(), time.Now().Add(10*time.Second), root, &budget{}); !errors.Is(err, errBudget) {
		t.Fatal(err)
	}
}
func strconvName(n int) string { return strconv.Itoa(n) }

func TestAdmissionOccurrenceAndDepthLimits(t *testing.T) {
	if !platformSupported() {
		t.Skip("platform intentionally unsupported")
	}
	t.Run("candidate-at-and-over", func(t *testing.T) {
		d, _ := descriptionFixture(t)
		rootBytes, _ := base64.StdEncoding.DecodeString(d.Cwd)
		root := string(rootBytes)
		d.Directories = nil
		d.Databases = nil
		for i := 0; i < maxCandidates+1; i++ {
			p := filepath.Join(root, "db"+strconv.Itoa(i))
			if err := os.Mkdir(p, 0700); err != nil {
				t.Fatal(err)
			}
			d.Databases = append(d.Databases, databaseInput{Name: "db" + strconv.Itoa(i), Root: encoded(p), Branch: "main"})
			if i == maxCandidates-1 || i == maxCandidates {
				raw, _ := json.Marshal(d)
				_, err := inspect(context.Background(), raw)
				if i == maxCandidates-1 && err != nil || i == maxCandidates && !errors.Is(err, errBudget) {
					t.Fatal(i, err)
				}
			}
		}
	})
	t.Run("files-at-and-over", func(t *testing.T) {
		d, _ := descriptionFixture(t)
		rootBytes, _ := base64.StdEncoding.DecodeString(d.Cwd)
		root := string(rootBytes)
		d.Directories = nil
		d.Sources = nil
		digest := sha256.Sum256([]byte("x"))
		for i := 0; i < maxFiles+1; i++ {
			p := filepath.Join(root, "file"+strconv.Itoa(i))
			mustWrite(t, p, []byte("x"), 0600)
			d.Sources = append(d.Sources, sourceInput{Path: encoded(p), SHA256: hex.EncodeToString(digest[:])})
			if i == maxFiles-1 || i == maxFiles {
				raw, _ := json.Marshal(d)
				_, err := inspect(context.Background(), raw)
				if i == maxFiles-1 && err != nil || i == maxFiles && !errors.Is(err, errBudget) {
					t.Fatal(i, err)
				}
			}
		}
	})
	t.Run("directory-depth-at-and-over", func(t *testing.T) {
		d, _ := descriptionFixture(t)
		rootBytes, _ := base64.StdEncoding.DecodeString(d.Cwd)
		p := string(rootBytes)
		for i := 1; i <= maxDirectoryDepth+1; i++ {
			p = filepath.Join(p, "level")
			if err := os.Mkdir(p, 0700); err != nil {
				t.Fatal(err)
			}
			if i == maxDirectoryDepth || i == maxDirectoryDepth+1 {
				d.Directories = []directoryInput{{Path: encoded(p)}}
				raw, _ := json.Marshal(d)
				_, err := inspect(context.Background(), raw)
				if i == maxDirectoryDepth && err != nil || i > maxDirectoryDepth && !errors.Is(err, errBudget) {
					t.Fatal(i, err)
				}
			}
		}
	})
}
func TestJSONCumulativeStringsAndRawLimit(t *testing.T) {
	raw := "[" + strings.Repeat(`"`+strings.Repeat("x", maxStringBytes)+`",`, 15) + `"` + strings.Repeat("x", maxStringBytes) + `"]`
	if err := validateJSON([]byte(raw), &budget{}); err != nil {
		t.Fatal(err)
	}
	over := raw[:len(raw)-1] + `,"x"]`
	if err := validateJSON([]byte(over), &budget{}); !errors.Is(err, errBudget) {
		t.Fatal(err)
	}
	_, data := descriptionFixture(t)
	padded := append(append([]byte(nil), data...), bytes.Repeat([]byte{' '}, maxFileBytes-len(data))...)
	if _, err := decodeDescription(padded, &budget{}); err != nil {
		t.Fatal(err)
	}
	if _, err := decodeDescription(append(padded, ' '), &budget{}); !errors.Is(err, errBudget) {
		t.Fatal(err)
	}
}
func TestExecutableStreamingBound(t *testing.T) {
	// The real stream helper uses a counted zero reader and one64KiB buffer.
	// This intentionally hashes the actual2GiB boundary once; the rejection
	// control below uses a small limit to avoid repeating expensive hash work.
	reader := &countedZeroReader{remaining: maxExecutableBytes}
	_, n, err := hashStream(context.Background(), time.Now().Add(10*time.Second), reader, maxExecutableBytes, &budget{}, true)
	if err != nil || n != maxExecutableBytes || reader.maxRead > 64<<10 {
		t.Fatal(n, reader.maxRead, err)
	}
	reader = &countedZeroReader{remaining: 18}
	if _, _, err = hashStream(context.Background(), time.Now().Add(time.Second), reader, 17, &budget{}, true); !errors.Is(err, errBudget) {
		t.Fatal(err)
	}
}

type countedZeroReader struct {
	remaining int64
	maxRead   int
}

func (r *countedZeroReader) Read(p []byte) (int, error) {
	r.maxRead = max(r.maxRead, len(p))
	if r.remaining == 0 {
		return 0, io.EOF
	}
	n := int(min(int64(len(p)), r.remaining))
	clear(p[:n])
	r.remaining -= int64(n)
	return n, nil
}

func TestDescriptionRejectsCaseFoldedFields(t *testing.T) {
	_, raw := descriptionFixture(t)
	for _, key := range []string{"version", "executable", "executable_sha256", "profile_sha256", "cwd", "security_roots", "sources", "databases", "directories", "environment", "path", "sha256", "name", "root", "default_branch", "entries", "value"} {
		t.Run(key, func(t *testing.T) {
			variant := bytes.Replace(raw, []byte(`"`+key+`":`), []byte(`"`+strings.ToUpper(key)+`":`), 1)
			if bytes.Equal(variant, raw) {
				t.Fatal("fixture key absent")
			}
			if _, err := decodeDescription(variant, &budget{}); err == nil {
				t.Fatal("case-folded field accepted")
			}
		})
	}
	alias := bytes.Replace(raw, []byte(`"version":1`), []byte(`"version":2,"VERSION":1`), 1)
	if _, err := decodeDescription(alias, &budget{}); err == nil {
		t.Fatal("case alias overwrote original field")
	}
	if _, err := decodeDescription(raw, &budget{}); err != nil {
		t.Fatal("exact field neighbor", err)
	}
}

func TestCancelledAdmissionStopsBeforeFilesystem(t *testing.T) {
	if !platformSupported() {
		t.Skip("platform intentionally unsupported")
	}
	d, _ := descriptionFixture(t)
	d.Cwd = encoded("/missing-managed-cancelled-input")
	raw, _ := json.Marshal(d)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := inspect(ctx, raw); !errors.Is(err, context.Canceled) {
		t.Fatal("filesystem inspection happened after cancellation", err)
	}
	a := admitted{roots: []rootPin{{path: "/missing-managed-cancelled-root"}}}
	if err := a.recheck(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal("root recheck happened after cancellation", err)
	}
}

func TestCanonicalWalkChecksWorkBetweenComponents(t *testing.T) {
	if !platformSupported() {
		t.Skip("platform intentionally unsupported")
	}
	// Err changes after three checks. The fourth check must happen before a
	// second path component is inspected, rather than after the whole walk.
	ctx := &cancelAfterChecks{Context: context.Background(), remaining: 3}
	_, err := decodeCanonical(ctx, time.Now().Add(time.Second), encoded(filepath.Join(testRoot(t), "missing-cancelled-component")), &budget{})
	if !errors.Is(err, context.Canceled) || ctx.calls != 4 {
		t.Fatal(ctx.calls, err)
	}
	_, err = decodeCanonical(context.Background(), time.Now().Add(-time.Second), encoded("/missing-expired-component"), &budget{})
	if !errors.Is(err, errBudget) {
		t.Fatal("expired admission inspected filesystem", err)
	}
	root := testRoot(t)
	a := admitted{roots: []rootPin{{path: root}, {path: "/missing-cancelled-next-root"}}}
	info, err := trustedDirectory(root)
	if err != nil {
		t.Fatal(err)
	}
	a.roots[0].identity = info
	ctx = &cancelAfterChecks{Context: context.Background(), remaining: 3}
	if err = a.recheck(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal("root loop omitted checkpoint", err)
	}
}

func TestRecheckCompletionCheckpoint(t *testing.T) {
	ctx := &cancelAfterChecks{Context: context.Background(), remaining: 1}
	if err := (admitted{}).recheck(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal("missing completion checkpoint", err)
	}
}

type cancelAfterChecks struct {
	context.Context
	remaining, calls int
}

func (c *cancelAfterChecks) Err() error {
	c.calls++
	if c.remaining == 0 {
		return context.Canceled
	}
	c.remaining--
	return nil
}
