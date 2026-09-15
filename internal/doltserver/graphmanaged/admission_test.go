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
	"syscall"
	"testing"
	"time"
	"unicode/utf8"
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
	d := description{Version: 1, Executable: encoded(exe), ExecutableSHA256: digest(exe), ProfileSHA256: strings.Repeat("a", 64), Cwd: encoded(root), Argv: []string{"--config=explicit"}, SecurityRoots: []string{encoded(root)}, Sources: []sourceInput{{Path: encoded(source), SHA256: digest(source)}}, Databases: []databaseInput{{Name: "graph", Root: encoded(root), Branch: "main"}}, Directories: []directoryInput{{Path: encoded(root), Entries: []string{encoded("adapter"), encoded("config")}}}, Environment: []environmentInput{{Name: "TMPDIR", Path: encoded(root)}}}
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
	d.Environment[0].Path = "changed"
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
			d.Environment = append(d.Environment, environmentInput{Name: "HOME", Path: encoded("/tmp")})
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
	for name, limit := range map[string]int64{"candidates": maxCandidates, "entries": maxEntries, "files": maxFiles, "source-bytes": maxSourceBytes, "nodes": maxNodes, "strings": maxStringsBytes, "paths": maxPathsBytes, "retained": maxRetainedBytes, "environment-count": maxEnvironment, "environment-bytes": maxEnvironmentBytes, "security-roots": maxSecurityRoots, "inventory-anchors": maxInventoryAnchors, "arguments": maxArguments, "argument-bytes": maxArgumentBytes} {
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
		t.Fatalf("host prerequisite: in-memory 2 GiB hashing throughput must allow completion within 10s; hashed=%d max_read=%d: %v", n, reader.maxRead, err)
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
	for _, key := range []string{"version", "executable", "executable_sha256", "profile_sha256", "cwd", "security_roots", "sources", "databases", "directories", "environment", "path", "sha256", "name", "root", "default_branch", "entries", "argv"} {
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

func inspectDescription(t *testing.T, d description) (admitted, error) {
	t.Helper()
	raw, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	return inspect(context.Background(), raw)
}

func TestCanonicalComponentLimit(t *testing.T) {
	if !platformSupported() {
		t.Skip("platform intentionally unsupported")
	}
	p := testRoot(t)
	count := len(strings.Split(strings.TrimPrefix(p, string(filepath.Separator)), string(filepath.Separator)))
	for count < 257 {
		p = filepath.Join(p, "x")
		if err := os.Mkdir(p, 0700); err != nil {
			t.Fatal(err)
		}
		count++
		if count == 256 || count == 257 {
			got, err := decodeCanonical(context.Background(), time.Now().Add(time.Second), encoded(p), &budget{})
			if count == 256 && (err != nil || got != p) || count == 257 && !errors.Is(err, errBudget) {
				t.Fatal(count, err)
			}
		}
	}
}

func TestInventoryOutsideCwd(t *testing.T) {
	if !platformSupported() {
		t.Skip("platform intentionally unsupported")
	}
	for _, kind := range []string{"security", "database", "unanchored"} {
		t.Run(kind, func(t *testing.T) {
			d, _ := descriptionFixture(t)
			outside := testRoot(t)
			if kind == "security" {
				d.SecurityRoots = append(d.SecurityRoots, encoded(outside))
			}
			if kind == "database" {
				d.Databases = append(d.Databases, databaseInput{Name: "outside", Root: encoded(outside), Branch: "main"})
			}
			p := outside
			for i := 1; i <= 9; i++ {
				p = filepath.Join(p, "x")
				if err := os.Mkdir(p, 0700); err != nil {
					t.Fatal(err)
				}
				if i != 8 && i != 9 {
					continue
				}
				d.Directories = []directoryInput{{Path: encoded(p)}}
				_, err := inspectDescription(t, d)
				if kind == "unanchored" {
					if !errors.Is(err, errInput) {
						t.Fatal(err)
					}
				} else if i == 8 && err != nil || i == 9 && !errors.Is(err, errBudget) {
					t.Fatal(i, err)
				}
			}
		})
	}
}

func TestRootAndInventoryOccurrenceLimits(t *testing.T) {
	if !platformSupported() {
		t.Skip("platform intentionally unsupported")
	}
	for _, kind := range []string{"security", "inventory"} {
		t.Run(kind, func(t *testing.T) {
			d, _ := descriptionFixture(t)
			rootBytes, _ := base64.StdEncoding.DecodeString(d.Cwd)
			root := string(rootBytes)
			d.Directories = nil
			limit := 64
			if kind == "security" {
				d.SecurityRoots = nil
			} else {
				limit = 256
			}
			for i := 0; i < limit; i++ {
				p := filepath.Join(root, "anchor"+strconv.Itoa(i))
				if err := os.Mkdir(p, 0700); err != nil {
					t.Fatal(err)
				}
				if kind == "security" {
					d.SecurityRoots = append(d.SecurityRoots, encoded(p))
				} else {
					d.Directories = append(d.Directories, directoryInput{Path: encoded(p)})
				}
			}
			if _, err := inspectDescription(t, d); err != nil {
				t.Fatal("at limit", err)
			}
			// The rejected duplicate occurrence must hit its budget before alias validation.
			if kind == "security" {
				d.SecurityRoots = append(d.SecurityRoots, d.SecurityRoots[0])
			} else {
				d.Directories = append(d.Directories, d.Directories[0])
			}
			if _, err := inspectDescription(t, d); !errors.Is(err, errBudget) {
				t.Fatal("over limit duplicate", err)
			}
		})
	}
}

func TestDescriptionArgumentEnvelope(t *testing.T) {
	if !platformSupported() {
		t.Skip("platform intentionally unsupported")
	}
	_, raw := descriptionFixture(t)
	var object map[string]any
	if err := json.Unmarshal(raw, &object); err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		name string
		args []string
		want error
	}{
		{"healthy", []string{"--config=/explicit/path", "positional", "generation", "--generation-other=1", "-generation-other=1"}, nil},
		{"count-at", make([]string, 64), nil},
		{"count-over", make([]string, 65), errBudget},
		{"bytes-at", []string{strings.Repeat("x", (16<<10)-1)}, nil},
		{"bytes-over", []string{strings.Repeat("x", 16<<10)}, errBudget},
		{"generation-equals", []string{"--generation=99"}, errInput},
		{"generation-bare", []string{"--generation", "99"}, errInput},
		{"protocol-equals", []string{"--managed-protocol=2"}, errInput},
		{"protocol-bare", []string{"--managed-protocol", "2"}, errInput},
		{"single-generation-equals", []string{"-generation=99"}, errInput},
		{"single-generation-bare", []string{"-generation", "99"}, errInput},
		{"single-protocol-equals", []string{"-managed-protocol=2"}, errInput},
		{"single-protocol-bare", []string{"-managed-protocol", "2"}, errInput},
		{"separator", []string{"--"}, errInput},
		{"nul", []string{"x\x00y"}, errInput},
	} {
		t.Run(tt.name, func(t *testing.T) {
			object["argv"] = tt.args
			data, err := json.Marshal(object)
			if err != nil {
				t.Fatal(err)
			}
			_, err = inspect(context.Background(), data)
			if tt.want == nil && err != nil || tt.want != nil && !errors.Is(err, tt.want) {
				t.Fatal(err)
			}
		})
	}
}

func TestTemporaryPathAdmissionAndRecheck(t *testing.T) {
	if !platformSupported() {
		t.Skip("platform intentionally unsupported")
	}
	for _, native := range []string{"private", "native-\xff"} {
		t.Run(encoded(native), func(t *testing.T) {
			d, _ := descriptionFixture(t)
			parent := testRoot(t)
			p := filepath.Join(parent, native)
			if err := os.Mkdir(p, 0700); err != nil {
				if !utf8.ValidString(native) && errors.Is(err, syscall.EILSEQ) {
					t.Skip("host filesystem rejects native-byte name; representation covered separately")
				}
				t.Fatal(err)
			}
			d.Environment = []environmentInput{{Name: "TMPDIR", Path: encoded(p)}}
			a, err := inspectDescription(t, d)
			if err != nil || len(a.environment) != 1 || a.environment[0] != "TMPDIR="+p {
				t.Fatal(a.environment, err)
			}
			if err = a.recheck(context.Background()); err != nil {
				t.Fatal(err)
			}
			if err = os.Rename(p, p+"-old"); err != nil {
				t.Fatal(err)
			}
			if err = os.Mkdir(p, 0700); err != nil {
				t.Fatal(err)
			}
			if err = a.recheck(context.Background()); !errors.Is(err, errInput) {
				t.Fatal("replaced TMPDIR admitted", err)
			}
		})
	}
	t.Run("control-before-filesystem", func(t *testing.T) {
		for _, bad := range []string{"/missing\ncontrol", "/missing\rcontrol", "/missing\x00control", "/missing/../unclean"} {
			d, _ := descriptionFixture(t)
			d.Cwd = encoded("/missing-must-not-be-inspected")
			d.Environment = []environmentInput{{Name: "TMPDIR", Path: encoded(bad)}}
			if _, err := inspectDescription(t, d); !errors.Is(err, errInput) {
				t.Fatal("filesystem preceded path validation", err)
			}
		}
	})
	t.Run("symlink-and-not-an-inventory-anchor", func(t *testing.T) {
		d, _ := descriptionFixture(t)
		parent := filepath.Join(testRoot(t), "parent")
		if err := os.Mkdir(parent, 0700); err != nil {
			t.Fatal(err)
		}
		p := filepath.Join(parent, "private")
		if err := os.Mkdir(p, 0700); err != nil {
			t.Fatal(err)
		}
		alias := filepath.Join(parent, "alias")
		if err := os.Symlink(parent, alias); err != nil {
			t.Fatal(err)
		}
		d.Environment = []environmentInput{{Name: "TMPDIR", Path: encoded(filepath.Join(alias, "private"))}}
		if _, err := inspectDescription(t, d); !errors.Is(err, errInput) {
			t.Fatal("intermediate symlink admitted", err)
		}
		d.Environment[0].Path = encoded(p)
		d.Directories = []directoryInput{{Path: encoded(p)}}
		if _, err := inspectDescription(t, d); !errors.Is(err, errInput) {
			t.Fatal("TMPDIR became an inventory anchor", err)
		}
		d.Directories = nil
		a, err := inspectDescription(t, d)
		if err != nil {
			t.Fatal(err)
		}
		if err = os.Rename(parent, parent+"-old"); err != nil {
			t.Fatal(err)
		}
		if err = os.Symlink(parent+"-old", parent); err != nil {
			t.Fatal(err)
		}
		if err = a.recheck(context.Background()); !errors.Is(err, errInput) {
			t.Fatal("ancestor alias passed TMPDIR recheck", err)
		}
	})
}

func TestInventoryUsesNearestAnchor(t *testing.T) {
	if !platformSupported() {
		t.Skip("platform intentionally unsupported")
	}
	d, _ := descriptionFixture(t)
	pBytes, _ := base64.StdEncoding.DecodeString(d.Cwd)
	p := string(pBytes)
	for range 9 {
		p = filepath.Join(p, "nested")
		if err := os.Mkdir(p, 0700); err != nil {
			t.Fatal(err)
		}
	}
	d.SecurityRoots = append(d.SecurityRoots, encoded(p))
	p = filepath.Join(p, "inventory")
	if err := os.Mkdir(p, 0700); err != nil {
		t.Fatal(err)
	}
	d.Directories = []directoryInput{{Path: encoded(p)}}
	if _, err := inspectDescription(t, d); err != nil {
		t.Fatal("nearer anchor ignored", err)
	}
	alias := filepath.Join(filepath.Dir(p), "alias")
	if err := os.Symlink(p, alias); err != nil {
		t.Fatal(err)
	}
	d.Directories = append(d.Directories, directoryInput{Path: encoded(alias)})
	if _, err := inspectDescription(t, d); !errors.Is(err, errInput) {
		t.Fatal("inventory alias admitted", err)
	}
}

func TestArgumentRetentionAndRuntimeNamespace(t *testing.T) {
	if !platformSupported() {
		t.Skip("platform intentionally unsupported")
	}
	d, _ := descriptionFixture(t)
	a, err := inspectDescription(t, d)
	if err != nil {
		t.Fatal(err)
	}
	original := a.argv[0]
	d.Argv[0] = "changed"
	if a.argv[0] != original {
		t.Fatal("argument aliases caller slice")
	}
	next, err := inspectDescription(t, d)
	if err != nil || next.config == a.config {
		t.Fatal("argv missing from description digest", err)
	}
	args := childArguments(a, math.MaxUint64)
	if args[len(args)-2] != "--managed-protocol=1" || args[len(args)-1] != "--generation=18446744073709551615" {
		t.Fatal(args)
	}
	if len(args[len(args)-2])+len(args[len(args)-1])+2 > 64 {
		t.Fatal("reserved argument envelope exceeded")
	}
	b := budget{retained: maxRetainedBytes}
	if _, err = admitArguments([]string{"x"}, &b); !errors.Is(err, errBudget) {
		t.Fatal("unaccounted retained argv", err)
	}
	b = budget{}
	if _, err = admitArguments([]string{"--generation=1"}, &b); !errors.Is(err, errInput) || b.arguments != 1 || b.argumentBytes != 15 {
		t.Fatal("rejected argument not charged", b, err)
	}
}

func TestEnvironmentNativeByteRepresentation(t *testing.T) {
	if !platformSupported() {
		t.Skip("platform intentionally unsupported")
	}
	p := "/explicit/native-\xff"
	b := budget{}
	env, err := admitEnvironment([]environmentInput{{Name: "TMPDIR", Path: encoded(p)}}, &b)
	if err != nil || len(env) != 1 || env[0] != "TMPDIR="+p || b.paths != int64(len(p)) || b.environmentBytes != int64(len(p)+7) {
		t.Fatal("native environment bytes replaced or uncharged", env, b, err)
	}
}

// Target the object shape, not the first shared key spelling in the document.
func TestDescriptionNestedCaseFoldedFields(t *testing.T) {
	_, raw := descriptionFixture(t)
	for _, target := range []struct{ shape, key string }{
		{"environment", "name"}, {"environment", "path"}, {"directories", "path"},
		{"sources", "path"}, {"sources", "sha256"},
		{"databases", "name"}, {"databases", "root"}, {"databases", "default_branch"},
		{"directories", "entries"},
	} {
		for _, alias := range []bool{false, true} {
			t.Run(target.shape+"/"+target.key+"/alias="+strconv.FormatBool(alias), func(t *testing.T) {
				var object map[string]any
				if err := json.Unmarshal(raw, &object); err != nil {
					t.Fatal(err)
				}
				entry := object[target.shape].([]any)[0].(map[string]any)
				value, present := entry[target.key]
				if !present {
					t.Fatal("fixture key absent")
				}
				entry[strings.ToUpper(target.key)] = value
				if !alias {
					delete(entry, target.key)
				}
				variant, err := json.Marshal(object)
				if err != nil {
					t.Fatal(err)
				}
				if _, err = decodeDescription(variant, &budget{}); err == nil {
					t.Fatal("nested case-folded field accepted")
				}
			})
		}
	}
	if _, err := decodeDescription(raw, &budget{}); err != nil {
		t.Fatal("exact field neighbor", err)
	}
}

func TestEnvironmentBudgetGuardsRejectedNames(t *testing.T) {
	for _, over := range []bool{false, true} {
		t.Run("over="+strconv.FormatBool(over), func(t *testing.T) {
			nameBytes := maxEnvironmentBytes - len("/x") - 1
			want := errInput
			if over {
				nameBytes++
				want = errBudget
			}
			b := budget{}
			_, err := admitEnvironment([]environmentInput{{Name: strings.Repeat("n", nameBytes), Path: encoded("/x")}}, &b)
			if !errors.Is(err, want) || b.environment != 1 {
				t.Fatal("rejected name budget", b, err)
			}
			if over && !strings.Contains(err.Error(), "environment bytes") {
				t.Fatal("wrong budget refused", err)
			}
			if !over && b.environmentBytes != maxEnvironmentBytes {
				t.Fatal("exact byte charge", b)
			}
		})
	}
	entries := make([]environmentInput, maxEnvironment+1)
	for i := range entries {
		entries[i] = environmentInput{Name: "TMPDIR", Path: encoded("/x")}
	}
	b := budget{}
	if _, err := admitEnvironment(entries, &b); !errors.Is(err, errInput) || b.environment != 2 {
		t.Fatal("single-key whitelist should dominate the entry ceiling", b, err)
	}
}

// Exercise direct append boundaries that the 32KiB production drain cannot
// reach in one read, including the discarded counter's saturation boundary.
func TestLogTailAppendBoundaries(t *testing.T) {
	const limit = 64 << 10
	pattern := bytes.Repeat([]byte("0123456789abcdef"), limit/16)
	full := bytes.Repeat([]byte("a"), limit)
	for _, tt := range []struct {
		name                     string
		initial, payload, want   []byte
		discarded, wantDiscarded uint64
	}{
		{"payload-at-limit", []byte("prior"), pattern, pattern, 7, 12},
		{"payload-over-limit", []byte("prior"), append([]byte("drop"), pattern...), pattern, 7, 16},
		{"discarded-below-max", full, []byte("x"), append(bytes.Clone(full[1:]), 'x'), math.MaxUint64 - 2, math.MaxUint64 - 1},
		{"discarded-at-max", full, []byte("xy"), append(bytes.Clone(full[2:]), 'x', 'y'), math.MaxUint64 - 2, math.MaxUint64},
		{"discarded-saturates", full, []byte("xyz"), append(bytes.Clone(full[3:]), 'x', 'y', 'z'), math.MaxUint64 - 2, math.MaxUint64},
	} {
		t.Run(tt.name, func(t *testing.T) {
			tail := logTail{bytes: bytes.Clone(tt.initial), discarded: tt.discarded}
			tail.append(tt.payload)
			if len(tail.bytes) != limit || !bytes.Equal(tail.bytes, tt.want) || tail.discarded != tt.wantDiscarded {
				t.Fatalf("retained=%d discarded=%d want_discarded=%d content_match=%v", len(tail.bytes), tail.discarded, tt.wantDiscarded, bytes.Equal(tail.bytes, tt.want))
			}
		})
	}
}
