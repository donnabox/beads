package doctor

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/beadsignore"
	"github.com/steveyegge/beads/internal/storage/domain"
	domainfs "github.com/steveyegge/beads/internal/storage/domain/fs"
)

func graphFixtureGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", dir}, args...)...)
	cmd.WaitDelay = time.Second
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("fixture git %v: %v %s", args, err, b)
	}
}
func TestGraphHygieneSharedConsumers(t *testing.T) {
	for _, name := range beadsignore.GraphPatterns() {
		if !containsGitignorePattern(GitignoreTemplate, name) {
			t.Fatal("canonical template", name)
		}
		found := false
		for _, p := range requiredPatterns {
			if p == name {
				found = true
			}
		}
		if !found {
			t.Fatal("doctor required pattern", name)
		}
	}
	root := t.TempDir()
	dir := filepath.Join(root, "external [literal]")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("BEADS_DIR", dir)
	repo := domainfs.NewBeadsDirFSRepository(root, domain.BeadsDirTemplates{BeadsGitignore: GitignoreTemplate})
	if err := repo.WriteBeadsGitignore(context.Background()); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil || string(b) != GitignoreTemplate {
		t.Fatal("domain template parity", err)
	}
}
func TestGraphHygieneTrackedSensitive(t *testing.T) {
	for _, external := range []bool{false, true} {
		for _, name := range []string{beadsignore.WitnessName, ".~graph-authority.local.json.orphan"} {
			t.Run(name+map[bool]string{false: " local", true: " external"}[external], func(t *testing.T) {
				root := t.TempDir()
				graphFixtureGit(t, root, "init", "-q")
				nameDir := ".beads"
				if external {
					nameDir = "external [literal]"
				}
				dir := filepath.Join(root, nameDir)
				if err := os.Mkdir(dir, 0700); err != nil {
					t.Fatal(err)
				}
				t.Setenv("BEADS_DIR", dir)
				clearResolveBeadsDirCache()
				t.Cleanup(clearResolveBeadsDirCache)
				if err := os.WriteFile(filepath.Join(dir, name), []byte("disposable fixture"), 0600); err != nil {
					t.Fatal(err)
				}
				graphFixtureGit(t, root, "--literal-pathspecs", "add", "-f", "--", filepath.Join(nameDir, name))
				check := CheckTrackedRuntimeFiles(root)
				if check.Status != StatusError || !strings.Contains(check.Detail, name) {
					t.Fatalf("sensitive classification: %#v", check)
				}
				if err := FixTrackedRuntimeFiles(root); err != nil {
					t.Fatal(err)
				}
				if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
					t.Fatal("local bytes removed", err)
				}
				if check = CheckTrackedRuntimeFiles(root); check.Status != StatusOK {
					t.Fatal("untrack result", check)
				}
			})
		}
	}
}
