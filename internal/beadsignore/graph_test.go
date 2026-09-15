package beadsignore

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func gitFixture(t *testing.T, dir string, args ...string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", dir}, args...)...)
	cmd.WaitDelay = time.Second
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("fixture git %v: %v %s", args, err, output)
	}
}
func TestGraphIgnoreGitAndNonGit(t *testing.T) {
	for _, isGit := range []bool{false, true} {
		t.Run(map[bool]string{false: "non git", true: "git"}[isGit], func(t *testing.T) {
			dir := t.TempDir()
			if isGit {
				gitFixture(t, dir, "init", "-q")
				gitFixture(t, dir, "config", "core.hooksPath", filepath.Join(dir, "no-hooks"))
			}
			custom := "# retained bytes\ncustom-rule"
			if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(custom), 0600); err != nil {
				t.Fatal(err)
			}
			syncs := 0
			flush := func() error { syncs++; return nil }
			if err := Prepare(context.Background(), dir, flush); err != nil {
				t.Fatal(err)
			}
			b, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
			if err != nil || !strings.HasPrefix(string(b), custom) {
				t.Fatal("changed existing bytes")
			}
			for _, name := range GraphPatterns() {
				if !strings.Contains(string(b), name) {
					t.Fatal(name)
				}
			}
			if err = Prepare(context.Background(), dir, flush); err != nil {
				t.Fatal(err)
			}
			again, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
			if err != nil || string(again) != string(b) || syncs != 2 {
				t.Fatal("retry not byte stable/durable")
			}
			names := GraphPatterns()
			names[0] = "bad"
			if GraphPatterns()[0] != "config.local.yaml" {
				t.Fatal("mutable names")
			}
		})
	}
}
func TestGraphIgnoreRefusals(t *testing.T) {
	for _, kind := range []string{"tracked", "negated", "bad repo", "symlink", "sync error", "canceled"} {
		t.Run(kind, func(t *testing.T) {
			dir := t.TempDir()
			gitFixture(t, dir, "init", "-q")
			ctx := context.Background()
			flush := func() error { return nil }
			switch kind {
			case "tracked":
				if err := os.WriteFile(filepath.Join(dir, WitnessName), []byte("fixture"), 0600); err != nil {
					t.Fatal(err)
				}
				gitFixture(t, dir, "add", "--", WitnessName)
			case "negated":
				if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(GraphTemplate+"!"+WitnessName+"\n"), 0600); err != nil {
					t.Fatal(err)
				}
			case "bad repo":
				if err := os.RemoveAll(filepath.Join(dir, ".git")); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, ".git"), []byte("gitdir: /does/not/exist\n"), 0600); err != nil {
					t.Fatal(err)
				}
			case "symlink":
				if err := os.Symlink("missing", filepath.Join(dir, ".gitignore")); err != nil {
					t.Fatal(err)
				}
			case "sync error":
				flush = func() error { return errors.New("fixture sync") }
			case "canceled":
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}
			if err := Prepare(ctx, dir, flush); err == nil {
				t.Fatal("hygiene admitted")
			}
		})
	}
}
func TestGraphSensitiveNames(t *testing.T) {
	for _, name := range []string{WitnessName, "nested/" + WitnessName, ".~graph-authority.local.json.orphan"} {
		if !Sensitive(name) {
			t.Fatal(name)
		}
	}
	for _, name := range []string{LockName, "graph-authority.local.json.backup", "unrelated"} {
		if Sensitive(name) {
			t.Fatal(name)
		}
	}
}
