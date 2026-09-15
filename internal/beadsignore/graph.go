// Package beadsignore owns the shared graph runtime names and local Git hygiene.
package beadsignore

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/steveyegge/beads/internal/atomicfile"
)

const WitnessName = "graph-authority.local.json"
const LockName = "graph-authority.lock"
const TemporaryPattern = ".~graph-authority.local.json.*"
const GraphTemplate = "\n# Graph runtime (local-only)\nconfig.local.yaml\ngraph-authority.local.json\ngraph-authority.lock\n.~graph-authority.local.json.*\n"

func GraphPatterns() []string {
	return []string{"config.local.yaml", WitnessName, LockName, TemporaryPattern}
}
func Sensitive(name string) bool {
	base := filepath.Base(name)
	return base == WitnessName || strings.HasPrefix(base, ".~graph-authority.local.json.")
}

var ErrHygiene = errors.New("graph runtime Git hygiene refused")

// Prepare appends missing rules without changing existing bytes. The caller owns
// application serialization and directory durability; sync is required even on retry.
func Prepare(ctx context.Context, dir string, syncDirectory func() error) error {
	if ctx == nil || syncDirectory == nil {
		return ErrHygiene
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	p := filepath.Join(dir, ".gitignore")
	var b []byte
	info, err := os.Lstat(p)
	if err == nil {
		if !info.Mode().IsRegular() {
			return ErrHygiene
		}
		f, e := os.Open(p) // #nosec G304 -- fixed .gitignore name in the resolved directory; inode checked before/after.
		if e != nil {
			return e
		}
		opened, e := f.Stat()
		if e != nil || !os.SameFile(info, opened) {
			return errors.Join(ErrHygiene, e, f.Close())
		}
		b, e = io.ReadAll(io.LimitReader(f, 1024*1024+1))
		closeErr := f.Close()
		if e != nil || closeErr != nil {
			return errors.Join(e, closeErr)
		}
		if len(b) > 1024*1024 {
			return ErrHygiene
		}
		now, e := os.Lstat(p)
		if e != nil || !os.SameFile(info, now) {
			return errors.Join(ErrHygiene, e)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	have := map[string]bool{}
	for _, line := range strings.Split(string(b), "\n") {
		have[strings.TrimSpace(line)] = true
	}
	var missing []string
	for _, name := range GraphPatterns() {
		if !have[name] {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		if len(b) > 0 && b[len(b)-1] != '\n' {
			b = append(b, '\n')
		}
		b = append(b, []byte("\n# Graph runtime (local-only)\n"+strings.Join(missing, "\n")+"\n")...)
		if err := atomicfile.WriteFile(p, b, 0600); err != nil {
			return err
		}
	}
	if err := syncDirectory(); err != nil {
		return err
	}
	for _, name := range []string{WitnessName, LockName, "config.local.yaml", ".~graph-authority.local.json.probe"} {
		if err := Check(ctx, dir, name, true); err != nil {
			return err
		}
	}
	return nil
}

// Check is read-only: a tracked path never becomes safe through ignore rules.
func Check(ctx context.Context, dir, name string, requireIgnored bool) error {
	root, code, err := gitRead(ctx, dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return err
	}
	if code != 0 {
		// Git's status alone is insufficient: independently establish no .git entry.
		for p := dir; ; p = filepath.Dir(p) {
			if _, e := os.Lstat(filepath.Join(p, ".git")); e == nil {
				return ErrHygiene
			} else if !errors.Is(e, os.ErrNotExist) {
				return e
			}
			if filepath.Dir(p) == p {
				break
			}
		}
		if code == 128 && strings.Contains(root, "not a git repository") {
			return nil
		}
		return ErrHygiene
	}
	_, code, err = gitRead(ctx, dir, "ls-files", "--error-unmatch", "--", name)
	if err != nil {
		return err
	}
	if code == 0 {
		return ErrHygiene
	}
	if code != 1 {
		return ErrHygiene
	}
	if requireIgnored {
		_, code, err = gitRead(ctx, dir, "check-ignore", "--no-index", "--", name)
		if err != nil {
			return err
		}
		if code != 0 {
			return ErrHygiene
		}
	}
	return nil
}

type boundedOutput struct{ bytes.Buffer }

func (b *boundedOutput) Write(p []byte) (int, error) {
	if b.Len()+len(p) > 256*1024 {
		return 0, ErrHygiene
	}
	return b.Buffer.Write(p)
}
func gitRead(ctx context.Context, dir string, args ...string) (string, int, error) {
	if ctx == nil {
		return "", -1, ErrHygiene
	}
	wait, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	prefix := []string{"-c", "core.fsmonitor=false", "-C", dir}
	// check-ignore consumes literal filenames itself and rejects pathspec magic.
	if len(args) == 0 || args[0] != "check-ignore" {
		prefix = append([]string{"--literal-pathspecs"}, prefix...)
	}
	argv := append(prefix, args...)
	cmd := exec.CommandContext(wait, "git", argv...)
	cmd.WaitDelay = time.Second
	for _, v := range os.Environ() {
		k, _, _ := strings.Cut(v, "=")
		if strings.HasPrefix(k, "GIT_") || k == "LC_ALL" {
			continue
		}
		cmd.Env = append(cmd.Env, v)
	}
	cmd.Env = append(cmd.Env, "LC_ALL=C", "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=/dev/null", "GIT_TERMINAL_PROMPT=0")
	var out boundedOutput
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if wait.Err() != nil {
		return "", -1, wait.Err()
	}
	if err == nil {
		return out.String(), 0, nil
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return out.String(), exit.ExitCode(), nil
	}
	return "", -1, fmt.Errorf("read graph Git hygiene: %w", err)
}
