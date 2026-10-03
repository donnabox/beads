package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestMain lets the test binary double as a fake `bd`: installFakeBD symlinks
// the binary as <dir>/bd, and a process started under that name runs
// fakeBDMain instead of the tests. What the fake does is chosen by fakebd.json
// next to the symlink, never by environment, because the driver hands its
// children a whitelisted environment and would strip anything else.
func TestMain(m *testing.M) {
	if filepath.Base(os.Args[0]) == "bd" {
		os.Exit(fakeBDMain(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
	}
	os.Exit(m.Run())
}

type fakeConfig struct {
	Legacy   bool `json:"legacy"`    // init succeeds but leaves no graph-preview-format marker
	InitExit int  `json:"init_exit"` // nonzero: init fails with this exit code
}

// installFakeBD returns the path of a fake bd configured by cfg.
func installFakeBD(t *testing.T, cfg fakeConfig) string {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable: %v", err)
	}
	dir := t.TempDir()
	bd := filepath.Join(dir, "bd")
	if err := os.Symlink(exe, bd); err != nil {
		t.Fatalf("symlink fake bd: %v", err)
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "fakebd.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	return bd
}

func loadFakeConfig() fakeConfig {
	var cfg fakeConfig
	if data, err := os.ReadFile(filepath.Join(filepath.Dir(os.Args[0]), "fakebd.json")); err == nil {
		_ = json.Unmarshal(data, &cfg)
	}
	return cfg
}

func fakeToken() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

// fakeBDMain implements just enough commands to drive the runner:
//
//	init, status --graph, version        lifecycle the driver itself performs
//	token, token-update                  write results carrying fresh random tokens
//	echo-args, echo-env, cwd, cat        observe how the driver invoked us
//	show-init-argv                       what init was called with
//	refuse CODE EXIT, refuse-text CODE EXIT  typed refusals, JSON and plain-text
//	sleep SECONDS, rand                  hang / produce a value normalization must NOT hide
func fakeBDMain(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	cfg := loadFakeConfig()
	if len(args) == 0 {
		fmt.Fprintln(stderr, "fake bd: no command")
		return 64
	}
	switch args[0] {
	case "init":
		if cfg.InitExit != 0 {
			fmt.Fprintln(stderr, "init_failed: fake init failure")
			return cfg.InitExit
		}
		if err := os.MkdirAll(".beads", 0o755); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		data, _ := json.Marshal(args)
		_ = os.WriteFile(filepath.Join(".beads", "init-argv"), data, 0o644)
		if !cfg.Legacy {
			_ = os.WriteFile(filepath.Join(".beads", "graph-preview-format"), []byte("link-preview-v5\n"), 0o644)
		}
		fmt.Fprintln(stdout, "Initialized fake workspace")
		return 0
	case "status":
		if _, err := os.Stat(filepath.Join(".beads", "graph-preview-format")); err != nil {
			fmt.Fprintln(stderr, "capability_unavailable: not a graph workspace")
			return 5
		}
		fmt.Fprintln(stdout, "Mixed graph preview (fake)")
		return 0
	case "version":
		fmt.Fprintln(stdout, "fake-bd 0.0.0")
		return 0
	case "show-init-argv":
		data, err := os.ReadFile(filepath.Join(".beads", "init-argv"))
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintln(stdout, string(data))
		return 0
	case "token":
		tok := fakeToken()
		fmt.Fprintf(stdout, `{"preview":true,"result":{"attribution":{"actor":"a","recordedAt":%q,"status":"claimed"},"id":"https://example.invalid/scenarios/beads/x","revision":%q,"version":%q},"schemaVersion":1}`+"\n",
			time.Now().UTC().Format(time.RFC3339Nano), tok, tok)
		return 0
	case "token-at":
		tok := fakeToken()
		fmt.Fprintf(stdout, `{"preview":true,"result":{"attribution":{"actor":"a","recordedAt":%q,"status":"claimed"},"id":"https://example.invalid/scenarios/beads/x","revision":%q,"version":%q},"schemaVersion":1}`+"\n",
			args[1], tok, tok)
		return 0
	case "token-update":
		tok := fakeToken()
		fmt.Fprintf(stdout, `{"preview":true,"result":{"changed":true,"memory":{"attribution":{"actor":"a","recordedAt":%q,"status":"claimed"},"id":"https://example.invalid/scenarios/beads/x","revision":%q,"version":%q},"replaced":{"id":"https://example.invalid/scenarios/beads/x","version":%q}},"schemaVersion":1}`+"\n",
			time.Now().UTC().Format(time.RFC3339Nano), tok, tok, fakeToken())
		return 0
	case "echo-args":
		enc := json.NewEncoder(stdout)
		_ = enc.Encode(args)
		return 0
	case "echo-env":
		env := os.Environ()
		sort.Strings(env)
		fmt.Fprintln(stdout, strings.Join(env, "\n"))
		return 0
	case "cwd":
		wd, _ := os.Getwd()
		fmt.Fprintln(stdout, wd)
		return 0
	case "cat":
		_, _ = io.Copy(stdout, stdin)
		return 0
	case "refuse", "refuse-text":
		if len(args) < 3 {
			return 64
		}
		exit, _ := strconv.Atoi(args[2])
		if args[0] == "refuse" {
			fmt.Fprintf(stderr, `{"code":%q,"message":"refused","retryable":false}`+"\n", args[1])
		} else {
			fmt.Fprintf(stderr, "%s: refused\n", args[1])
		}
		return exit
	case "sleep":
		secs, _ := strconv.ParseFloat(args[1], 64)
		time.Sleep(time.Duration(secs * float64(time.Second)))
		return 0
	case "rand":
		fmt.Fprintln(stdout, fakeToken()[:8])
		return 0
	}
	fmt.Fprintf(stderr, "fake bd: unknown command %q\n", args[0])
	return 64
}
