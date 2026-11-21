//go:build unix

package graphsession

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

const compositionChildKey = "BEADS_TEST_LOCAL_INFILE_CASE"
const compositionArtifactsKey = "BEADS_TEST_LOCAL_INFILE_ARTIFACTS"

type compositionOutput struct {
	mu        sync.Mutex
	data      bytes.Buffer
	truncated bool
}

func (b *compositionOutput) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	// Retain a bounded prefix but continue draining so output cannot block exit.
	n := min(len(p), (1<<20)-b.data.Len())
	_, err := b.data.Write(p[:n])
	if n != len(p) {
		b.truncated = true
	}
	return len(p), err
}
func (b *compositionOutput) String() string { b.mu.Lock(); defer b.mu.Unlock(); return b.data.String() }
func writeCompositionReceipt(t *testing.T, dir, name string, value any) {
	t.Helper()
	b, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile(filepath.Join(dir, name), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.Write(append(b, '\n'))
	err = errors.Join(err, f.Close())
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("composition_owner_%s=%s", name, b)
}

// compositionOwner has exactly one waiter and no concurrent signal path. The
// fixed child does not fork, close its output descriptors, install an auto-reap
// handler, or call Wait on other children. Ordinary Go executable hosting is a
// prerequisite; this fixture does not qualify arbitrary embedded hosts.
//
// In particular, Darwin's os.Process does not by itself fix concurrent Signal /
// Wait PID reuse. Every signal below completes before the sole Wait starts.
// Natural output EOF is an observation, not a process-status or engine proof.
type compositionOwner struct {
	cmd      *exec.Cmd
	release  *os.File
	reader   *os.File
	output   *compositionOutput
	readDone chan struct{}
	readErr  error // published by closing readDone
	waitDone chan error
	events   []string // only the parent coordinator writes these
	finished bool
	result   compositionOutcome
}

type compositionOutcome struct {
	Forced          bool     `json:"forced"`
	Reaped          bool     `json:"reaped"`
	OutputJoined    bool     `json:"output_joined"`
	OutputTruncated bool     `json:"output_truncated"`
	WaitError       string   `json:"wait_error"`
	WaitErr         error    `json:"-"`
	CleanupErr      error    `json:"-"`
	Events          []string `json:"events"`
	Err             error    `json:"-"`
}

func startCompositionChild(t *testing.T, name string, bound time.Duration) (*compositionOwner, string) {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if root := os.Getenv(compositionArtifactsKey); root != "" {
		dir, err = os.MkdirTemp(root, "case-"+strings.ReplaceAll(name, "/", "-")+"-")
		if err != nil {
			t.Fatal(err)
		}
	}
	input, release, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(errors.Join(err, input.Close(), release.Close()))
	}
	cmd := exec.Command(exe, "-test.run=^TestLocalInfileCompositionChild$", "-test.v", "-test.timeout="+(bound-5*time.Second).String())
	cmd.Env = append(os.Environ(), compositionChildKey+"="+name)
	// Actual files avoid os/exec copy workers and Wait-owned output closure.
	cmd.Stdin, cmd.Stdout, cmd.Stderr = input, writer, writer
	o := &compositionOwner{cmd: cmd, release: release, reader: reader, output: &compositionOutput{}, readDone: make(chan struct{}), waitDone: make(chan error, 1)}
	started := false
	t.Cleanup(func() {
		if started && !o.finished {
			r := o.finish(true)
			writeCompositionReceipt(t, dir, "rescue.json", map[string]any{"case": name, "pid": cmd.Process.Pid, "result": r, "error": fmt.Sprint(r.Err), "output": o.output.String()})
			if !r.Reaped || !r.OutputJoined || r.Err != nil {
				t.Errorf("pre-release child cleanup: %v", r.Err)
			}
		}
	})
	if err = cmd.Start(); err != nil {
		t.Fatal(errors.Join(err, input.Close(), release.Close(), reader.Close(), writer.Close()))
	}
	started = true
	go func() {
		_, o.readErr = io.Copy(o.output, reader)
		close(o.readDone)
	}()
	if err = errors.Join(input.Close(), writer.Close()); err != nil {
		t.Fatal(err)
	}
	writeCompositionReceipt(t, dir, "before-release.json", map[string]any{"case": name, "pid": cmd.Process.Pid, "argv": cmd.Args, "output_directory": dir, "bound_seconds": bound.Seconds(), "release": "pending", "wait_started": false, "process_groups": false})
	return o, dir
}

func compositionClose(f *os.File) error {
	err := f.Close()
	if errors.Is(err, os.ErrClosed) {
		return nil
	}
	return err
}

func (o *compositionOwner) releaseChild() error {
	if err := o.release.SetWriteDeadline(time.Now().Add(time.Second)); err != nil {
		return err
	}
	_, err := o.release.Write([]byte{1})
	return errors.Join(err, compositionClose(o.release))
}

func (o *compositionOwner) finish(force bool) compositionOutcome {
	if o.finished {
		return o.result
	}
	// Mark before any I/O: repeated cleanup cannot signal, wait, or close twice.
	o.finished = true
	r := compositionOutcome{Forced: force}
	r.Err = compositionClose(o.release)
	if force {
		o.events = append(o.events, "signal-start")
		err := o.cmd.Process.Kill()
		o.events = append(o.events, "signal-finished")
		if err != nil && !errors.Is(err, os.ErrProcessDone) && !errors.Is(err, syscall.ESRCH) {
			r.Err = errors.Join(r.Err, err)
		}
	}
	// There are no signals after this point, even when Wait misses its bound.
	o.events = append(o.events, "wait-start")
	go func() { o.waitDone <- o.cmd.Wait() }()
	deadline := time.Now().Add(2 * time.Second)
	timer := time.NewTimer(time.Until(deadline))
	select {
	case err := <-o.waitDone:
		r.Reaped = true
		r.WaitError = fmt.Sprint(err)
		r.WaitErr = err
	case <-timer.C:
		// The retained owner and its sole waiter remain unresolved. Never attempt
		// a late signal, another waiter, or claim successful process cleanup.
		r.Err = errors.Join(r.Err, errors.New("owned child Wait unresolved; sole waiter retained, no late rescue"))
	}
	timer.Stop()
	// Interrupt our read worker independently of Cmd.Wait. Closing a pollable
	// os.Pipe read end wakes a pending Read; preserve partial output on failure.
	select {
	case <-o.readDone:
	default:
		r.Err = errors.Join(r.Err, compositionClose(o.reader))
	}
	join := time.NewTimer(max(time.Until(deadline), time.Millisecond))
	select {
	case <-o.readDone:
		r.OutputJoined = true
		if o.readErr != nil && !errors.Is(o.readErr, os.ErrClosed) {
			r.Err = errors.Join(r.Err, o.readErr)
		}
	case <-join.C:
		r.Err = errors.Join(r.Err, errors.New("owned output worker not joined"))
	}
	join.Stop()
	r.Err = errors.Join(r.Err, compositionClose(o.reader))
	o.output.mu.Lock()
	r.OutputTruncated = o.output.truncated
	o.output.mu.Unlock()
	if r.OutputTruncated {
		r.Err = errors.Join(r.Err, errors.New("composition child output limit"))
	}
	r.CleanupErr = r.Err
	r.Err = errors.Join(r.CleanupErr, r.WaitErr)
	r.Events = append([]string(nil), o.events...)
	o.result = r
	return r
}

func runCompositionChild(t *testing.T, name string, bound time.Duration) {
	t.Helper()
	start := time.Now()
	o, dir := startCompositionChild(t, name, bound)
	if err := o.releaseChild(); err != nil {
		t.Fatal(err)
	}
	forced := false
	timer := time.NewTimer(time.Until(start.Add(bound - 5*time.Second)))
	select {
	case <-o.readDone:
	case <-timer.C:
		forced = true
	}
	timer.Stop()
	r := o.finish(forced)
	text := o.output.String()
	writeCompositionReceipt(t, dir, "after-wait.json", map[string]any{"case": name, "pid": o.cmd.Process.Pid, "elapsed_seconds": time.Since(start).Seconds(), "forced": r.Forced, "wait_error": r.WaitError, "reaped": r.Reaped, "output_joined": r.OutputJoined, "events": r.Events, "error": fmt.Sprint(r.Err), "output": text})
	if r.Err != nil {
		t.Fatalf("child failed: %v\n%s", r.Err, text)
	}
	if forced {
		t.Fatal("child exceeded natural close budget; forced rescue is failure")
	}
	if !strings.Contains(text, "composition_protocol=") || !strings.Contains(text, "--- PASS: TestLocalInfileCompositionChild") {
		t.Fatalf("missing executed protocol/terminal test receipt: %s", text)
	}
	if time.Since(start) > bound {
		t.Errorf("case exceeded %s envelope", bound)
	}
}

// These three process-lifecycle controls are separate from the twelve protocol
// cases. Their 3*10s envelope raises the total to 230s; focused cap is 270s.
func TestLocalInfileOwnerOrder(t *testing.T) {
	for _, name := range []string{"exited-before-rescue", "timeout", "pre-release"} {
		t.Run(name, func(t *testing.T) {
			start := time.Now()
			child := "lifecycle/block"
			if name == "exited-before-rescue" {
				child = "lifecycle/exit"
			}
			o, dir := startCompositionChild(t, child, 10*time.Second)
			if name != "pre-release" {
				if err := o.releaseChild(); err != nil {
					t.Fatal(err)
				}
				timer := time.NewTimer(2 * time.Second)
				if name == "exited-before-rescue" {
					select {
					case <-o.readDone:
					case <-timer.C:
						t.Fatal("child output did not reach EOF")
					}
				} else {
					select {
					case <-o.readDone:
						t.Fatal("blocked child exited before timeout")
					case <-timer.C:
					}
				}
				timer.Stop()
			}
			r := o.finish(true)
			// This recreates the timer/exit ordering hazard without PID reuse:
			// even already-observed EOF does not start Wait before rescue ends.
			if strings.Join(r.Events, ",") != "signal-start,signal-finished,wait-start" {
				t.Errorf("unsafe order: %v", r.Events)
			}
			var exitError *exec.ExitError
			if !r.Reaped || !r.OutputJoined || r.CleanupErr != nil {
				t.Errorf("cleanup: %+v error=%v", r, r.Err)
			}
			if name == "exited-before-rescue" && r.WaitErr != nil {
				t.Errorf("naturally exited child status: %v", r.WaitErr)
			}
			if name == "timeout" && (!strings.Contains(o.output.String(), "composition_lifecycle=block") || !errors.As(r.WaitErr, &exitError)) {
				t.Errorf("missing blocked-child execution / forced exit evidence: %v %s", r.WaitErr, o.output.String())
			}
			if name == "exited-before-rescue" && (!strings.Contains(o.output.String(), "composition_lifecycle=exit") || !strings.Contains(o.output.String(), "--- PASS: TestLocalInfileCompositionChild")) {
				t.Error("missing natural child completion output")
			}
			again := o.finish(true)
			if len(again.Events) != 3 || len(o.events) != 3 {
				t.Errorf("repeated finish signaled or waited again: %v", o.events)
			}
			writeCompositionReceipt(t, dir, "after-wait.json", map[string]any{"case": "owner/" + name, "pid": o.cmd.Process.Pid, "elapsed_seconds": time.Since(start).Seconds(), "forced": r.Forced, "wait_error": r.WaitError, "reaped": r.Reaped, "output_joined": r.OutputJoined, "events": r.Events, "error": fmt.Sprint(r.Err), "output": o.output.String(), "expected_rescue_control": true})
			if time.Since(start) > 10*time.Second {
				t.Error("lifecycle case exceeded 10s envelope")
			}
		})
	}
}
func TestLocalInfileComposition(t *testing.T) {
	// Protocol envelope: 8*20s + 4*10s = 200s. The separate lifecycle
	// family adds 30s, within the corrected 270s focused package cap.
	for _, stage := range []string{"initialize", "commit"} {
		for _, kind := range []string{"reader", "file"} {
			for _, offer := range []string{"offered", "omitted"} {
				name := stage + "/" + kind + "/" + offer
				t.Run(name, func(t *testing.T) { runCompositionChild(t, name, 20*time.Second) })
			}
		}
	}
	for _, name := range []string{"normal/offered", "normal/omitted", "row/value", "cleanup/error"} {
		t.Run(name, func(t *testing.T) { runCompositionChild(t, name, 10*time.Second) })
	}
}
func TestLocalInfileCompositionChild(t *testing.T) {
	name := os.Getenv(compositionChildKey)
	if name == "" {
		t.Skip("inert owned subprocess entry point")
	}
	var release [1]byte
	if _, err := io.ReadFull(os.Stdin, release[:]); err != nil || release[0] != 1 {
		t.Fatalf("owner release missing: %v", err)
	}
	switch name {
	case "lifecycle/exit":
		fmt.Println("composition_lifecycle=exit")
	case "lifecycle/block":
		fmt.Println("composition_lifecycle=block")
		// A timer keeps the runtime alive until the owner terminates this child.
		<-time.NewTimer(time.Hour).C
	case "normal/offered":
		compositionNormal(t, true, false)
	case "normal/omitted":
		compositionNormal(t, false, false)
	case "row/value":
		compositionNormal(t, true, true)
	case "cleanup/error":
		compositionCleanup(t)
	default:
		parts := strings.Split(name, "/")
		if len(parts) != 3 || (parts[0] != "initialize" && parts[0] != "commit") || (parts[1] != "reader" && parts[1] != "file") || (parts[2] != "offered" && parts[2] != "omitted") {
			t.Fatalf("invalid child selection %q", name)
		}
		compositionAdversary(t, parts[0], parts[1], parts[2] == "offered")
	}
}
