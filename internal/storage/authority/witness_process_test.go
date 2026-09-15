//go:build unix

package authority

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/atomicfile"
	"golang.org/x/sys/unix"
)

type witnessBarrier struct {
	PID                int
	Stage, OperationID string
}

func childBarrier(t *testing.T, stage, id string) {
	t.Helper()
	if err := json.NewEncoder(os.Stdout).Encode(witnessBarrier{os.Getpid(), stage, id}); err != nil {
		t.Fatal(err)
	}
	var b [1]byte
	if _, err := os.Stdin.Read(b[:]); err != nil {
		t.Fatal(err)
	}
	t.Fatal("crash barrier released without parent kill")
}

// Inert in normal runs; only the owning parent supplies all three fixture inputs.
func TestWitnessProcessHelper(t *testing.T) {
	if os.Getenv("WITNESS_FIXTURE_HELPER") != "owned-crash-control" {
		t.Skip("owned subprocess helper")
	}
	dir, stage := os.Getenv("WITNESS_FIXTURE_DIR"), os.Getenv("WITNESS_FIXTURE_STAGE")
	if dir == "" || stage == "" {
		t.Fatal("incomplete helper fixture")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	m, err := New(ctx, dir)
	if err != nil {
		t.Fatal(err)
	}
	g, err := m.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := g.Close(); err != nil {
			t.Error(err)
		}
	}()
	if stage == "lock" {
		childBarrier(t, stage, "")
	}
	point := strings.TrimPrefix(stage, "foreign ")
	foreign := point != stage
	kind := Mutation
	mode := PromoteMode("")
	if stage == "self operation" {
		point = "operation"
		kind = Promote
		mode = SelfRegrant
	}
	if foreign {
		kind = Promote
		mode = Steal
	}
	if !foreign && (stage == "config after" || stage == "config_written") {
		kind = Rotate
	}
	if stage == "final remove" {
		kind = Adopt
	}
	e := recording()
	tr, err := g.Begin(ctx, intentFor(kind, mode, Ordinary), e)
	if err != nil {
		t.Fatal(err)
	}
	s, err := g.read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	id := s.OperationID()
	if point == "begun" {
		childBarrier(t, stage, id)
	}
	// This is a named fixture fact file, never a simulated engine result receipt.
	fact := completionFor(s)
	b, err := json.Marshal(fact)
	if err != nil {
		t.Fatal(err)
	}
	if err = atomicfile.WriteFile(filepath.Join(dir, "fixture-facts.json"), b, 0600); err != nil {
		t.Fatal(err)
	}
	if err = syncWitnessAncestors(ctx, dir); err != nil {
		t.Fatal(err)
	}
	if point == "operation" {
		childBarrier(t, stage, id)
	}
	c := &fileConfiguration{}
	if stage == "config after" {
		c.afterWrite = func() error { childBarrier(t, stage, id); return ErrEvidence }
	}
	for _, phase := range []Phase{LocalCommitted, Published, ConfigWritten} {
		if err = tr.SetPhase(ctx, phase, e, c); err != nil {
			t.Fatal(err)
		}
		if point == string(phase) {
			childBarrier(t, stage, id)
		}
	}
	if point == "before final" {
		replace := m.io.replace
		m.io.replace = func(p string, b []byte, mode os.FileMode) error {
			next, err := decodeEnvelope(b)
			if err != nil {
				return err
			}
			if !next.Pending() {
				childBarrier(t, stage, id)
			}
			return replace(p, b, mode)
		}
		if err = tr.Finalize(ctx, e, c); err != nil {
			t.Fatal(err)
		}
	}
	if point == "final replace" || point == "final remove" {
		flush := m.io.syncDirectory
		m.io.syncDirectory = func(ctx context.Context, d string) error {
			visible, err := readWitnessFile(m.path())
			if err != nil {
				return err
			}
			if !visible.Pending() {
				childBarrier(t, stage, id)
			}
			return flush(ctx, d)
		}
		if err = tr.Finalize(ctx, e, c); err != nil {
			t.Fatal(err)
		}
	}
	t.Fatal("unknown crash stage")
}

type ownedWitnessChild struct {
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
	cancel  context.CancelFunc
	waited  bool
	barrier witnessBarrier
	stderr  *os.File
}

func launchWitnessChild(t *testing.T, m *Manager, stage string) *ownedWitnessChild {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	cmd := exec.CommandContext(ctx, exe, "-test.run=^TestWitnessProcessHelper$", "-test.timeout=18s")
	cmd.WaitDelay = time.Second
	cmd.Env = append(os.Environ(), "WITNESS_FIXTURE_HELPER=owned-crash-control", "WITNESS_FIXTURE_DIR="+m.dir, "WITNESS_FIXTURE_STAGE="+stage)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		if e := stdin.Close(); e != nil {
			t.Error(e)
		}
		t.Fatal(err)
	}
	stderr, err := os.Create(filepath.Join(m.dir, "fixture-child-stderr"))
	if err != nil {
		cancel()
		if e := stdin.Close(); e != nil {
			t.Error(e)
		}
		if e := stdout.Close(); e != nil {
			t.Error(e)
		}
		t.Fatal(err)
	}
	cmd.Stderr = stderr
	if err = cmd.Start(); err != nil {
		cancel()
		if e := stdin.Close(); e != nil {
			t.Error(e)
		}
		if e := stdout.Close(); e != nil {
			t.Error(e)
		}
		if e := stderr.Close(); e != nil {
			t.Error(e)
		}
		t.Fatal(err)
	}
	c := &ownedWitnessChild{cmd: cmd, stdin: stdin, stdout: stdout, cancel: cancel, stderr: stderr}
	t.Cleanup(func() {
		if !c.waited {
			c.killAndReap(t)
		}
	})
	type result struct {
		barrier witnessBarrier
		err     error
	}
	done := make(chan result, 1)
	go func() {
		var b witnessBarrier
		err := json.NewDecoder(bufio.NewReader(stdout)).Decode(&b)
		done <- result{b, err}
	}()
	select {
	case r := <-done:
		if r.err != nil {
			c.killAndReap(t)
			b, e := os.ReadFile(stderr.Name())
			t.Fatalf("child barrier: %v; log %q (%v)", r.err, b, e)
		}
		c.barrier = r.barrier
	case <-ctx.Done():
		c.killAndReap(t)
		r := <-done
		t.Fatalf("child barrier timeout: %v (%v)", ctx.Err(), r.err)
	}
	if c.barrier.PID != cmd.Process.Pid || c.barrier.Stage != stage {
		t.Fatal("unowned barrier")
	}
	return c
}
func (c *ownedWitnessChild) killAndReap(t *testing.T) {
	t.Helper()
	if c.waited {
		t.Fatal("child waited twice")
	}
	err := c.cmd.Process.Kill()
	if err != nil && !errors.Is(err, os.ErrProcessDone) {
		t.Error(err)
	}
	waitErr := c.cmd.Wait()
	c.waited = true
	c.cancel()
	if err = c.stdin.Close(); err != nil && !errors.Is(err, os.ErrClosed) {
		t.Error(err)
	}
	if err = c.stderr.Close(); err != nil {
		t.Error(err)
	}
	var exit *exec.ExitError
	if waitErr != nil && !errors.As(waitErr, &exit) {
		t.Error(waitErr)
	}
	if err = unix.Kill(c.cmd.Process.Pid, 0); !errors.Is(err, unix.ESRCH) {
		t.Errorf("child %d not absent: %v", c.cmd.Process.Pid, err)
	}
	t.Logf("owned child pid=%d stage=%s reaped; signal-zero=ESRCH; wait=%v", c.cmd.Process.Pid, c.barrier.Stage, waitErr)
}
func TestWitnessProcessCrashRecovery(t *testing.T) {
	for _, stage := range []string{"begun", "operation", "self operation", "foreign operation", "local_committed", "published", "config after", "config_written", "final replace", "final remove", "foreign begun", "foreign local_committed", "foreign published", "foreign config_written", "foreign before final", "foreign final replace"} {
		t.Run(stage, func(t *testing.T) {
			m := testManager(t)
			foreign := strings.HasPrefix(stage, "foreign ")
			point := strings.TrimPrefix(stage, "foreign ")
			saveActive(t, m, foreign, foreign)
			configFixture(t, m)
			c := launchWitnessChild(t, m, stage)
			id := c.barrier.OperationID
			if id == "" {
				t.Fatal("missing durable operation")
			}
			c.killAndReap(t)
			restarted, err := New(context.Background(), m.dir)
			if err != nil {
				t.Fatal(err)
			}
			g := testGuard(t, restarted)
			s, err := g.read(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if point == "final replace" || point == "final remove" {
				if s.Pending() {
					t.Fatal("final name not visible")
				}
				if err = g.Recover(context.Background(), nil, nil); err != nil {
					t.Fatal("retained durability flush", err)
				}
				if (point == "final remove") != s.Absent() {
					t.Fatal("wrong final effect")
				}
				return
			}
			if !s.Pending() || s.OperationID() != id {
				t.Fatal("crash lost operation")
			}
			e := recording()
			e.owner.SessionGeneration = "fresh-after-crash"
			e.inspect = func(r InspectRequest) (Inspection, error) {
				if point == "begun" {
					return NewInspection(r, preFor(r.Snapshot()))
				}
				b, err := os.ReadFile(filepath.Join(m.dir, "fixture-facts.json"))
				if err != nil {
					return Inspection{}, err
				}
				var f InspectionFields
				if err = json.Unmarshal(b, &f); err != nil {
					return Inspection{}, err
				}
				return NewInspection(r, f)
			}
			if err = g.Recover(context.Background(), e, &fileConfiguration{}); err != nil {
				t.Fatal(err)
			}
			after, err := g.read(context.Background())
			if err != nil || after.Pending() {
				t.Fatal("recovery incomplete", err)
			}
			if point == "begun" {
				w, ok := after.Witness()
				if !ok || w.fields.StateCommit != "before" || w.fields.Unverified != foreign {
					t.Fatal("no-effect recovery changed prior")
				}
			}
			if e.prepareCount != 0 {
				t.Fatal("recovery prepared second operation")
			}
		})
	}
}
func TestWitnessProcessLockCapAndReacquire(t *testing.T) {
	m := testManager(t)
	c := launchWitnessChild(t, m, "lock")
	before, err := os.Lstat(filepath.Join(m.dir, "graph-authority.lock"))
	if err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	if _, err = m.Acquire(context.Background()); !errors.Is(err, ErrWitnessBusy) {
		t.Fatal("internal lock cap", err)
	}
	if time.Since(start) > 4*time.Second {
		t.Fatal("unbounded lock cap")
	}
	c.killAndReap(t)
	g := testGuard(t, m)
	now, err := g.file.Stat()
	if err != nil || !os.SameFile(before, now) {
		t.Fatal("lock inode changed")
	}
	t.Log(fmt.Sprintf("parent reacquired permanent lock after pid %d was reaped", c.cmd.Process.Pid))
}
