//go:build linux

package dolt

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/doltserver"
	"github.com/steveyegge/beads/internal/testutil"
	"github.com/steveyegge/beads/internal/testutil/bazeltest"
)

// This process boundary exercises the real TestMain allocation and exit policy,
// not just the selector: an ordinary t.TempDir fixture must belong to the suite.
// A shell blocked on stdin supplies a real process without requiring Dolt.
func TestSuiteFixtureLeakFailsOwningSuite(t *testing.T) {
	if os.Getenv("BEADS_SUITE_LEAK_PROBE") == "1" {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		stdin := os.NewFile(3, "leak-probe-stdin")
		t.Cleanup(func() { _ = stdin.Close() })
		actor := startSuiteLeakActor(t, ctx, t.TempDir(), stdin, false)
		fmt.Printf("LEAK_PROBE_ID=%d:%s\n%s\n", actor.cmd.Process.Pid, actor.startTime, actor.diagnostic())
		// Intentionally omit shutdown. The parent holds stdin open through
		// this suite's sweep, then closes it even if the suite fails.
		return
	}

	// A foreign suite's deleted-cwd process must survive the child's sweep.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	foreignRead, foreignWrite, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer foreignRead.Close()
	defer foreignWrite.Close()
	foreign := startSuiteLeakActor(t, ctx, t.TempDir(), foreignRead, true)
	if err := os.Remove(foreign.cmd.Dir); err != nil {
		t.Fatal(err)
	}

	readEnd, writeEnd, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer readEnd.Close()
	defer writeEnd.Close()
	t.Setenv("BEADS_SUITE_LEAK_PROBE", "1")
	t.Setenv("BEADS_TEST_SKIP", "dolt")
	// The child skips Dolt on purpose, so a lane's fail-closed switch (set
	// when this suite runs against a required server) must not turn that
	// skip into a FATAL before the child reaches its leak sweep.
	t.Setenv(testutil.EnvRequireDoltContainer, "")
	t.Setenv(doltserver.AllowLeakEnv, "")
	// This child is a complete suite, not one of the schema-init helpers
	// which deliberately bypass suite ownership and shutdown.
	for _, key := range helperSubprocessSentinels {
		t.Setenv(key, "")
	}
	child := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestSuiteFixtureLeakFailsOwningSuite$", "-test.count=1")
	child.Env = bazeltest.ShardFreeEnv(os.Environ())
	child.ExtraFiles = []*os.File{readEnd}
	out, err := child.CombinedOutput()
	// Release the intentional stand-in on every outcome, including a child
	// that unexpectedly passed without reaping it. Only observe the PID
	// reported by this child; never sweep or signal a foreign root here.
	_ = writeEnd.Close()
	assertSuiteLeakProbeStopped(t, ctx, out)
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
		t.Fatalf("leaking suite returned %v; want exit 1\n%s", err, out)
	}
	if !strings.Contains(string(out), "FAIL: internal/storage/dolt leaked 1 dolt sql-server") {
		t.Fatalf("suite failed without attributing its fixture leak:\n%s", out)
	}
	select {
	case <-foreign.done:
		t.Fatal("child suite consumed the foreign suite's leak evidence")
	default:
	}
}

// suiteLeakActor keeps fixture startup and exit observable. Start alone does
// not establish that the shell reached its blocking read before cwd removal.
type suiteLeakActor struct {
	cmd       *exec.Cmd
	done      chan struct{}
	waitErr   error // published by closing done
	startTime string
	stderr    suiteLeakStderr
}

type suiteLeakStderr struct {
	mu   sync.Mutex
	text string
}

func (b *suiteLeakStderr) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	n := len(p)
	if remaining := 4096 - len(b.text); remaining > 0 {
		if len(p) > remaining {
			p = p[:remaining]
		}
		b.text += string(p)
	}
	return n, nil
}
func (b *suiteLeakStderr) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.text
}

func startSuiteLeakActor(t *testing.T, ctx context.Context, cwd string, stdin *os.File, cleanup bool) *suiteLeakActor {
	t.Helper()
	readyRead, readyWrite, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = readyRead.Close(); _ = readyWrite.Close() })
	actor := &suiteLeakActor{cmd: exec.Command("/bin/sh", "-c", "printf R; read line", "dolt", "sql-server"), done: make(chan struct{})}
	actor.cmd.Dir, actor.cmd.Stdin = cwd, stdin
	actor.cmd.Stdout, actor.cmd.Stderr = readyWrite, &actor.stderr
	if err := actor.cmd.Start(); err != nil {
		t.Fatal(err)
	}
	_ = readyWrite.Close()
	go func() { actor.waitErr = actor.cmd.Wait(); close(actor.done) }()
	t.Cleanup(func() {
		if cleanup || t.Failed() {
			// This exact exec.Cmd is ours. Never signal a PID discovered by
			// a broad scan or expand the suite sweeper's ownership boundary.
			select {
			case <-actor.done:
			default:
				_ = actor.cmd.Process.Kill()
				<-actor.done
			}
		}
	})
	ready := make(chan error, 1)
	go func() {
		var b [1]byte
		_, err := io.ReadFull(readyRead, b[:])
		if err == nil && b[0] != 'R' {
			err = fmt.Errorf("unexpected ready byte %q", b[0])
		}
		ready <- err
	}()
	select {
	case err := <-ready:
		if err != nil {
			t.Fatalf("leak actor readiness: %v; %s", err, actor.diagnostic())
		}
	case <-actor.done:
		t.Fatalf("leak actor exited before readiness: %s", actor.diagnostic())
	case <-ctx.Done():
		t.Fatalf("leak actor readiness: %v; %s", ctx.Err(), actor.diagnostic())
	}
	cmdline, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", actor.cmd.Process.Pid))
	actualCwd, cwdErr := os.Readlink(fmt.Sprintf("/proc/%d/cwd", actor.cmd.Process.Pid))
	expectedCwd, canonicalErr := filepath.EvalSymlinks(cwd)
	var startErr error
	actor.startTime, startErr = suiteLeakStartTime(actor.cmd.Process.Pid)
	if startErr != nil {
		t.Fatalf("leak actor start identity: %v; %s", startErr, actor.diagnostic())
	}
	if err != nil || cwdErr != nil || !strings.Contains(string(cmdline), "dolt") || !strings.Contains(string(cmdline), "sql-server") || canonicalErr != nil || actualCwd != expectedCwd {
		t.Fatalf("leak actor identity/cwd not established: %s", actor.diagnostic())
	}
	return actor
}

func (a *suiteLeakActor) diagnostic() string {
	pid := a.cmd.Process.Pid
	cmdline, cmdErr := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
	cwd, cwdErr := os.Readlink(fmt.Sprintf("/proc/%d/cwd", pid))
	state := "running"
	select {
	case <-a.done:
		state = fmt.Sprintf("exited: %v", a.waitErr)
	default:
	}
	return fmt.Sprintf("actor pid=%d requestedCwd=%q cwd=%q cwdErr=%v cmdline=%q cmdErr=%v state=%s stderr=%q", pid, a.cmd.Dir, cwd, cwdErr, cmdline, cmdErr, state, a.stderr.String())
}

func assertSuiteLeakProbeStopped(t *testing.T, ctx context.Context, out []byte) {
	t.Helper()
	for _, line := range strings.Split(string(out), "\n") {
		if !strings.HasPrefix(line, "LEAK_PROBE_ID=") {
			continue
		}
		pidText, startTime, ok := strings.Cut(strings.TrimPrefix(line, "LEAK_PROBE_ID="), ":")
		pid, err := strconv.Atoi(pidText)
		if err != nil || pid <= 0 || !ok || startTime == "" {
			t.Fatalf("invalid own probe PID: %q", line)
		}
		// An orphan may remain briefly as a zombie awaiting its reaper. An
		// absent/empty cmdline means it is no longer an executing actor.
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()
		for {
			currentStart, err := suiteLeakStartTime(pid)
			if os.IsNotExist(err) || (err == nil && currentStart != startTime) {
				return
			}
			if err != nil {
				t.Fatalf("inspect own leak probe identity %d: %v", pid, err)
			}
			cmdline, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "cmdline"))
			if os.IsNotExist(err) || (err == nil && len(cmdline) == 0) {
				return
			}
			if err != nil {
				t.Fatalf("inspect own leak probe %d: %v", pid, err)
			}
			select {
			case <-ctx.Done():
				t.Fatalf("own leak probe %d still executing after stdin close: %v\n%s", pid, ctx.Err(), out)
			case <-ticker.C:
			}
		}
	}
	t.Fatalf("child produced no owned probe identity:\n%s", out)
}

// Linux stat field22 is process starttime. The command name can contain
// spaces/parentheses, so fields after its final closing parenthesis start at3.
func suiteLeakStartTime(pid int) (string, error) {
	raw, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "stat"))
	if err != nil {
		return "", err
	}
	end := strings.LastIndex(string(raw), ") ")
	if end < 0 {
		return "", fmt.Errorf("invalid proc stat for %d", pid)
	}
	fields := strings.Fields(string(raw)[end+2:])
	if len(fields) <= 19 {
		return "", fmt.Errorf("short proc stat for %d", pid)
	}
	return fields[19], nil
}
