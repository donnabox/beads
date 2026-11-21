package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	// markerName is written into every workspace root by newWorkspace. No step
	// runs unless the marker is there and names that root, so a bd can only ever
	// be pointed at a directory this driver made for the purpose.
	markerName = ".scenarios-driver-workspace"

	// hermeticPath is the whole PATH a bd child gets: enough to find git, and
	// nothing from the caller.
	hermeticPath = "/usr/local/bin:/usr/bin:/bin"

	scenarioActor = "scenario-actor"
	scopeURL      = "https://example.invalid/scenarios/"
	scopePrefix   = "scn"
)

// stepTimeout caps every bd call. A variable only so a test can shrink it.
var stepTimeout = 90 * time.Second

// harnessError is a failure of the driver's own machinery, not of a scenario:
// bd missing, an unsafe workspace, graph mode not active, a timeout. run maps
// it to exit 2.
type harnessError struct{ msg string }

func (e *harnessError) Error() string { return e.msg }

func harnessf(format string, args ...any) error {
	return &harnessError{fmt.Sprintf(format, args...)}
}

// workspace is one disposable bd project: root/work is the project directory
// (cwd of every bd call) and root/home is HOME and the XDG dirs.
type workspace struct {
	Root, Work, Home string
}

func newWorkspace() (*workspace, error) {
	root, err := os.MkdirTemp("", "scenarios-")
	if err != nil {
		return nil, err
	}
	w := &workspace{Root: root, Work: filepath.Join(root, "work"), Home: filepath.Join(root, "home")}
	fail := func(err error) (*workspace, error) {
		_ = os.RemoveAll(root)
		return nil, err
	}
	for _, dir := range []string{w.Work, w.Home} {
		if err := os.Mkdir(dir, 0o700); err != nil {
			return fail(err)
		}
	}
	// GIT_CONFIG_GLOBAL points here: an empty file, so no global git config leaks in.
	if err := os.WriteFile(filepath.Join(w.Home, "empty-gitconfig"), nil, 0o600); err != nil {
		return fail(err)
	}
	if err := os.WriteFile(filepath.Join(root, markerName), []byte(root+"\n"), 0o600); err != nil {
		return fail(err)
	}
	return w, nil
}

// close removes the workspace. It only ever removes a directory newWorkspace
// could have made.
func (w *workspace) close() error {
	if !filepath.IsAbs(w.Root) || !strings.HasPrefix(filepath.Base(w.Root), "scenarios-") {
		return fmt.Errorf("refusing to remove %q: not a workspace this driver made", w.Root)
	}
	return os.RemoveAll(w.Root)
}

// guard refuses a workspace that is not safe to point bd at: the root must be
// under the OS temp dir, and carry the driver's marker naming that very root.
func (w *workspace) guard() error {
	if !filepath.IsAbs(w.Root) {
		return fmt.Errorf("workspace root %q is not absolute", w.Root)
	}
	realRoot, err := filepath.EvalSymlinks(w.Root)
	if err != nil {
		return fmt.Errorf("workspace root: %w", err)
	}
	realTmp, err := filepath.EvalSymlinks(os.TempDir())
	if err != nil {
		return fmt.Errorf("OS temp dir: %w", err)
	}
	if rel, err := filepath.Rel(realTmp, realRoot); err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("workspace root %s is not under the OS temp dir %s", realRoot, realTmp)
	}
	marker, err := os.ReadFile(filepath.Join(w.Root, markerName)) //nolint:gosec // G304: the marker inside a directory this driver created
	if err != nil {
		return fmt.Errorf("workspace marker missing: %w", err)
	}
	if got := strings.TrimSpace(string(marker)); got != w.Root {
		return fmt.Errorf("workspace marker names %q, not this root %q", got, w.Root)
	}
	if filepath.Dir(w.Work) != w.Root || filepath.Dir(w.Home) != w.Root {
		return fmt.Errorf("work and home must sit directly under the workspace root")
	}
	return nil
}

// hermeticVars is the whole environment a bd child sees. It is built from
// nothing: the caller's environment is never read, so BEADS_DIR, BEADS_DB, any
// BD_* or anything else ambient cannot reach bd whatever the caller exported.
func hermeticVars(w *workspace) map[string]string {
	return map[string]string{
		"HOME":                          w.Home,
		"XDG_CONFIG_HOME":               filepath.Join(w.Home, ".config"),
		"XDG_CACHE_HOME":                filepath.Join(w.Home, ".cache"),
		"XDG_DATA_HOME":                 filepath.Join(w.Home, ".local", "share"),
		"XDG_STATE_HOME":                filepath.Join(w.Home, ".local", "state"),
		"PATH":                          hermeticPath,
		"TZ":                            "UTC",
		"NO_COLOR":                      "1",
		"GIT_CONFIG_NOSYSTEM":           "1",
		"GIT_CONFIG_GLOBAL":             filepath.Join(w.Home, "empty-gitconfig"),
		"BEADS_DOLT_AUTO_START":         "0",
		"BD_DISABLE_METRICS":            "1",
		"BD_DISABLE_EVENT_FLUSH":        "1",
		"DOLT_METRICS_DISABLED":         "1",
		"BEADS_ACTOR":                   scenarioActor,
		"BEADS_TEST_IGNORE_REPO_CONFIG": "1",
	}
}

func hermeticEnv(w *workspace) []string {
	vars := hermeticVars(w)
	env := make([]string, 0, len(vars))
	for k, v := range vars {
		env = append(env, k+"="+v)
	}
	sort.Strings(env)
	return env
}

// assertHermeticEnv fails if the child environment holds anything off the
// whitelist. It is the check behind "never inherit BEADS_DIR, BEADS_DB or any
// BD_* from the parent": it runs before every bd call, so a future edit that
// starts passing something through trips it instead of silently pointing bd
// at a live workspace.
func assertHermeticEnv(env []string) error {
	allowed := hermeticVars(&workspace{})
	for _, kv := range env {
		key, _, _ := strings.Cut(kv, "=")
		if _, ok := allowed[key]; !ok {
			return fmt.Errorf("child environment holds %s, which is not on the whitelist (ambient BEADS_*, BD_* and GC_* must never reach bd)", key)
		}
	}
	return nil
}

// exec runs bd with argv in the workspace, capped at stepTimeout, and returns
// what it printed. bd is always the explicit binary the caller resolved, never
// looked up on PATH.
func (w *workspace) exec(ctx context.Context, bd string, argv []string, stdin []byte) (*stepResult, error) {
	if err := w.guard(); err != nil {
		return nil, harnessf("unsafe workspace: %v", err)
	}
	env := hermeticEnv(w)
	if err := assertHermeticEnv(env); err != nil {
		return nil, harnessf("%v", err)
	}
	ctx, cancel := context.WithTimeout(ctx, stepTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, bd, argv...) //nolint:gosec // G204: bd is the explicit binary under test resolved from --bd and argv is scenario data; running exactly that is the driver's job
	cmd.Dir, cmd.Env = w.Work, env
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	// If bd leaves a grandchild holding the pipes, Wait must still return.
	cmd.WaitDelay = 2 * time.Second
	err := cmd.Run()

	res := &stepResult{Argv: argv, Stdout: stdout.Bytes(), Stderr: stderr.Bytes()}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return nil, harnessf("timeout: bd %s did not finish within %s", strings.Join(argv, " "), stepTimeout)
	}
	var exitErr *exec.ExitError
	switch {
	case err == nil:
	case errors.As(err, &exitErr):
		if exitErr.ExitCode() < 0 {
			return nil, harnessf("bd %s was killed by a signal: %v", strings.Join(argv, " "), err)
		}
		res.Exit = exitErr.ExitCode()
	default:
		return nil, harnessf("cannot run bd %s: %v", strings.Join(argv, " "), err)
	}
	return res, nil
}

var initArgv = []string{
	"init", "--graph-mode", "link", "--scope-url", scopeURL, "--prefix", scopePrefix,
	"--skip-hooks", "--skip-agents", "--non-interactive",
}

// initGraph creates the graph-mode project and proves graph mode is active.
// BD_GRAPH_MODE is only an assertion after init; init is what enables it.
func (w *workspace) initGraph(ctx context.Context, bd string) error {
	res, err := w.exec(ctx, bd, initArgv, nil)
	if err != nil {
		return err
	}
	if res.Exit != 0 {
		return harnessf("bd init failed (exit %d): %s", res.Exit, strings.TrimSpace(string(res.Stderr)))
	}
	res, err = w.exec(ctx, bd, []string{"status", "--graph"}, nil)
	if err != nil {
		return err
	}
	if res.Exit != 0 {
		return harnessf("graph mode is not active: bd status --graph exited %d: %s", res.Exit, strings.TrimSpace(string(res.Stderr)))
	}
	if _, err := os.Stat(filepath.Join(w.Work, ".beads", "graph-preview-format")); err != nil {
		return harnessf("graph mode is not active: .beads/graph-preview-format is missing (%v)", err)
	}
	return nil
}
