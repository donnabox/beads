//go:build darwin || linux

package graphmanaged

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

// The fixture runs this test binary only, imports no engine, never launches
// descendants, and accepts only its private test switches after a positional fixture marker.
func TestManagedFixture(t *testing.T) {
	mode := ""
	options := map[string]string{}
	for _, arg := range os.Args {
		if key, value, ok := strings.Cut(arg, "="); ok {
			options[key] = value
		}
	}
	mode = options["--managed-fixture"]
	if mode == "" {
		return
	}
	if err := runFixture(mode, options); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	os.Exit(0)
}
func inheritedPipe(fd uintptr) (*os.File, error) {
	if err := syscall.SetNonblock(int(fd), true); err != nil {
		return nil, err
	}
	syscall.CloseOnExec(int(fd))
	f := os.NewFile(fd, "managed fixture pipe")
	info, err := f.Stat()
	if err != nil {
		return nil, errors.Join(err, f.Close())
	}
	if info.Mode()&os.ModeNamedPipe == 0 {
		return nil, errors.Join(errInput, f.Close())
	}
	if err = f.SetDeadline(time.Now().Add(10 * time.Second)); err != nil {
		return nil, errors.Join(err, f.Close())
	}
	return f, nil
}
func runFixture(mode string, options map[string]string) error {
	if mode == "exit-error" {
		ready, err := inheritedPipe(3)
		if err != nil {
			return err
		}
		_, err = ready.Write([]byte{1})
		closeErr := ready.Close()
		if err != nil || closeErr != nil {
			return errors.Join(err, closeErr)
		}
		return errors.New("deliberate fixture exit failure")
	}
	if mode == "exit-before" {
		return nil
	}
	control, err := inheritedPipe(4)
	if err != nil {
		return err
	}
	defer control.Close()
	report, err := inheritedPipe(5)
	if err != nil {
		return err
	}
	defer report.Close()
	f := os.NewFile(3, "managed inherited listener")
	syscall.CloseOnExec(3)
	listener, err := net.FileListener(f)
	closeErr := f.Close()
	if err != nil || closeErr != nil {
		return errors.Join(err, closeErr)
	}
	defer listener.Close()
	addr, ok := listener.Addr().(*net.TCPAddr)
	if !ok || !addr.IP.Equal(net.IPv4(127, 0, 0, 1)) {
		return errInput
	}
	generation, err := strconv.ParseUint(options["--generation"], 10, 64)
	if err != nil {
		return err
	}
	if mode == "ignore-term" || mode == "graceful-eof" || mode == "ignore-eof" {
		signal.Ignore(syscall.SIGTERM)
	}
	if mode == "wait-prepared" || mode == "ignore-term" {
		time.Sleep(time.Hour)
		return nil
	}
	prepared := frame{Protocol: 1, Phase: "prepared", Generation: generation, Profile: options["--profile"], Config: options["--config"], Registration: options["--registration"], Socket: listener.Addr().String()}
	switch mode {
	case "wrong-phase":
		prepared.Phase = "activated"
	case "wrong-socket":
		prepared.Socket = "127.0.0.1:1"
	case "oversize":
		var b [4]byte
		binary.BigEndian.PutUint32(b[:], maxFrame+1)
		_, err = report.Write(b[:])
		if err != nil {
			return err
		}
		time.Sleep(time.Hour)
	case "malformed", "duplicate-json", "truncated":
		payload := []byte(`{"protocol":`)
		if mode == "duplicate-json" {
			payload = []byte(`{"protocol":1,"protocol":1}`)
		}
		var b [4]byte
		binary.BigEndian.PutUint32(b[:], uint32(len(payload)))
		if mode == "truncated" {
			binary.BigEndian.PutUint32(b[:], uint32(len(payload)+100))
		}
		if _, err = report.Write(append(b[:], payload...)); err != nil {
			return err
		}
		return nil
	}
	bytes, err := frameBytes(prepared)
	if err != nil {
		return err
	}
	if err = writeFrame(report, bytes, time.Second); err != nil {
		return err
	}
	if mode == "duplicate-prepared" {
		if err = writeFrame(report, bytes, time.Second); err != nil {
			return err
		}
	}
	if mode == "premature-ack" || mode == "premature-eof" || mode == "premature-extra" {
		ack := prepared
		ack.Phase = "activated"
		bytes, err := frameBytes(ack)
		if err != nil {
			return err
		}
		if err = writeFrame(report, bytes, time.Second); err != nil {
			return err
		}
		if mode == "premature-eof" {
			if err = report.Close(); err != nil {
				return err
			}
		}
		if mode == "premature-extra" {
			_, err = report.Write([]byte{1})
			if err != nil {
				return err
			}
		}
		time.Sleep(time.Hour)
		return nil
	}
	if mode == "never-control" {
		time.Sleep(time.Hour)
		return nil
	}
	if mode == "blocked-report" {
		_, err = report.Write(make([]byte, 1<<20))
		return err
	}
	activation, err := readFrame(context.Background(), control, time.Second)
	if err != nil {
		return err
	}
	if activation != (frame{Protocol: 1, Phase: "activate", Generation: generation}) {
		return errProtocol
	}
	if mode == "no-ack" {
		time.Sleep(time.Hour)
		return nil
	}
	ack := prepared
	ack.Phase = "activated"
	if mode == "wrong-sequence" {
		ack.Generation++
	}
	if mode == "wrong-digest" {
		ack.Config = strings.Repeat("f", 64)
	}
	// SQL-free accepting transition: the inherited socket is already listening;
	// a dedicated accept worker exists before ack and is joined before exit.
	acceptDone := make(chan struct{})
	go func() {
		defer close(acceptDone)
		for {
			conn, e := listener.Accept()
			if e != nil {
				return
			}
			_, _ = conn.Write([]byte("fixture\n"))
			_ = conn.Close()
		}
	}()
	defer func() { _ = listener.Close(); <-acceptDone }()
	bytes, err = frameBytes(ack)
	if err != nil {
		return err
	}
	if err = writeFrame(report, bytes, time.Second); err != nil {
		return err
	}
	if mode == "exit-after" {
		return nil
	}
	if mode == "extra-byte" {
		_, err = report.Write([]byte{1})
		if err != nil {
			return err
		}
	}
	if mode == "noisy" {
		for range 16 {
			if _, err = os.Stdout.Write(make([]byte, 32<<10)); err != nil {
				return err
			}
			if _, err = os.Stderr.Write(make([]byte, 32<<10)); err != nil {
				return err
			}
		}
	}
	if mode == "ignore-eof" {
		time.Sleep(time.Hour)
		return nil
	}
	var extra [1]byte
	err = readExact(context.Background(), control, extra[:], time.Second)
	if errors.Is(err, io.EOF) {
		return nil
	}
	return err
}

var fixtureTiming = timing{startup: 3 * time.Second, cleanup: 5 * time.Second, grace: 50 * time.Millisecond, pipe: 100 * time.Millisecond}

func fixtureAdmission(t *testing.T, mode string) admitted {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	exe, err = filepathCanonical(exe)
	if err != nil {
		t.Fatal(err)
	}
	root := testRoot(t)
	a := admitted{executable: exe, cwd: root, profile: strings.Repeat("a", 64), config: strings.Repeat("b", 64), registration: strings.Repeat("c", 64)}
	pin, err := hashFile(context.Background(), time.Now().Add(10*time.Second), exe, true, &budget{})
	if err != nil {
		t.Fatal(err)
	}
	a.pins = []filePin{pin}
	a.argv, err = admitArguments([]string{"-test.run=^TestManagedFixture$", "managed-fixture", "--profile=" + a.profile, "--config=" + a.config, "--registration=" + a.registration, "--managed-fixture=" + mode}, &budget{})
	if err != nil {
		t.Fatal(err)
	}
	return a
}
func filepathCanonical(p string) (string, error) { return filepath.EvalSymlinks(p) }
func launchFixture(t *testing.T, ctx context.Context, mode string, tm timing) (*controller, *generation, error) {
	t.Helper()
	a := fixtureAdmission(t, mode)
	c := &controller{observe: recordOwner}
	g, err := c.startWithTiming(ctx, a, tm)
	if g != nil {
		t.Logf("owned child pid=%d mode=%s", g.cmd.Process.Pid, mode)
		t.Cleanup(func() {
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
			defer cancel()
			if err := g.close(cleanupCtx); errors.Is(err, errCleanup) {
				t.Error("fixture cleanup incomplete", err)
			}
			select {
			case <-g.waitDone:
			default:
				t.Errorf("child %d not reaped", g.cmd.Process.Pid)
			}
		})
	}
	return c, g, err
}

// This decorator records entry and return of actual process methods. It is
// private to each fixture and delegates every call to that fixture's sole Cmd.
type recordedProcess struct {
	inner  processOwner
	mu     sync.Mutex
	events []string
}

func recordOwner(inner processOwner) processOwner { return &recordedProcess{inner: inner} }
func (r *recordedProcess) record(event string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, event)
}
func (r *recordedProcess) Signal(s os.Signal) error {
	r.record("signal-start")
	err := r.inner.Signal(s)
	r.record("signal-end")
	return err
}
func (r *recordedProcess) Kill() error {
	r.record("kill-start")
	err := r.inner.Kill()
	r.record("kill-end")
	return err
}
func (r *recordedProcess) Wait() error {
	r.record("wait-start")
	err := r.inner.Wait()
	r.record("wait-end")
	return err
}
func assertReaped(t *testing.T, g *generation) error {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()
	err := g.close(ctx)
	if errors.Is(err, errCleanup) {
		t.Fatal("cleanup did not complete", err)
	}
	select {
	case <-g.done:
	default:
		t.Fatal("close returned before coordinator completion", err)
	}
	select {
	case <-g.waitDone:
	default:
		t.Fatal("not reaped")
	}
	for _, done := range g.workers {
		select {
		case <-done:
		default:
			t.Fatal("worker not joined")
		}
	}
	recorder, ok := g.process.(*recordedProcess)
	if !ok {
		t.Fatal("missing actual process recorder")
	}
	if !recordsCommand(recorder, g.cmd) || g.cmd.ProcessState == nil {
		t.Fatal("recorder does not identify the waited child")
	}
	recorder.mu.Lock()
	events := slices.Clone(recorder.events)
	recorder.mu.Unlock()
	if !slices.Equal(events, []string{"signal-start", "signal-end", "kill-start", "kill-end", "wait-start", "wait-end"}) {
		t.Fatal(events)
	}
	if g.valid() {
		t.Fatal("still valid")
	}
	return err
}
func TestManagedProcessManifest(t *testing.T) {
	cases := []struct {
		name, mode string
		healthy    bool
	}{
		{"01_healthy", "healthy", true},
		{"03_exit_before_prepared", "exit-before", false},
		{"04_exit_after_ack", "exit-after", true},
		{"09_never_reads_activation", "never-control", false},
		{"10_no_activation_ack", "no-ack", false},
		{"11_wrong_phase", "wrong-phase", false},
		{"12_wrong_sequence", "wrong-sequence", false},
		{"12_wrong_digest", "wrong-digest", false},
		{"13_duplicate_prepared", "duplicate-prepared", false},
		{"14_extra_serving_byte", "extra-byte", true},
		{"15_oversized_header", "oversize", false},
		{"16_malformed_json", "malformed", false},
		{"16_duplicate_json", "duplicate-json", false},
		{"16_truncated_payload", "truncated", false},
		{"17_hung_startup", "wait-prepared", false},
		{"18_control_eof", "graceful-eof", true},
		{"18_ignored_control_eof", "ignore-eof", true},
		{"19_ignored_term", "ignore-term", false},
		{"21_natural_exit_ordering", "exit-after", true},
		{"25_noisy_logs", "noisy", true},
		{"26_blocked_report_writer", "blocked-report", false},
		{"28_wrong_socket", "wrong-socket", false},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			tm := fixtureTiming
			if tt.mode == "graceful-eof" || tt.mode == "ignore-eof" {
				tm.grace = 2 * time.Second
			}
			_, g, err := launchFixture(t, ctx, tt.mode, tm)
			loss := tt.mode == "exit-after" || tt.mode == "extra-byte"
			if !loss && (err == nil) != tt.healthy {
				t.Fatalf("healthy=%v err=%v", tt.healthy, err)
			}
			if g == nil {
				t.Fatal("no child")
			}
			if tt.mode == "blocked-report" && !errors.Is(err, errProtocol) {
				t.Fatal("blocked report did not fail protocol admission", err)
			}
			if loss && err != nil {
				want := io.EOF
				if tt.mode == "extra-byte" {
					want = errProtocol
				}
				if !errors.Is(err, want) || g.valid() {
					t.Fatal("invalid prepublication refusal", err)
				}
				conn, dialErr := net.DialTimeout("tcp4", g.expected.Socket, 100*time.Millisecond)
				if dialErr == nil {
					_ = conn.Close()
					t.Fatal("refused endpoint still accepting")
				}
			}
			if tt.mode == "healthy" || tt.mode == "graceful-eof" || tt.mode == "ignore-eof" {
				if !g.valid() {
					t.Fatal("healthy start returned an invalid generation")
				}
				conn, err := net.DialTimeout("tcp4", g.expected.Socket, time.Second)
				if err != nil {
					t.Fatal(err)
				}
				if err = conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
					t.Fatal(err)
				}
				var buf [8]byte
				_, err = io.ReadFull(conn, buf[:])
				_ = conn.Close()
				if err != nil || string(buf[:]) != "fixture\n" {
					t.Fatal(err)
				}
			}
			if tt.mode == "noisy" {
				deadline := time.Now().Add(time.Second)
				for time.Now().Before(deadline) {
					g.stdout.mu.Lock()
					d := g.stdout.discarded
					g.stdout.mu.Unlock()
					g.stderr.mu.Lock()
					e := g.stderr.discarded
					g.stderr.mu.Unlock()
					if d > 0 && e > 0 {
						break
					}
					time.Sleep(time.Millisecond)
				}
			}
			if tt.mode == "exit-after" || tt.mode == "extra-byte" {
				deadline := time.Now().Add(time.Second)
				for g.valid() && time.Now().Before(deadline) {
					time.Sleep(time.Millisecond)
				}
				if g.valid() {
					t.Fatal("loss did not promptly invalidate")
				}
			}
			closeErr := assertReaped(t, g)
			if tt.mode == "graceful-eof" {
				if !onlyProcessGone(closeErr) || !g.cmd.ProcessState.Success() {
					t.Fatal("control EOF did not exit cleanly", closeErr, g.cmd.ProcessState)
				}
			}
			if tt.mode == "ignore-eof" {
				var exitError *exec.ExitError
				if !errors.As(closeErr, &exitError) || g.cmd.ProcessState.Success() {
					t.Fatal("ignored EOF escaped forced-exit discriminator", closeErr)
				}
			}
			if tt.mode == "noisy" {
				for _, tail := range []*logTail{&g.stdout, &g.stderr} {
					tail.mu.Lock()
					n, d := len(tail.bytes), tail.discarded
					tail.mu.Unlock()
					if n > 64<<10 || d == 0 {
						t.Fatal(n, d)
					}
				}
			}
		})
	}
}

// Linux pidfd signalling may report an already-exited unreaped child. That
// expected signal result is compatible with clean EOF; no other error is.
func onlyProcessGone(err error) bool {
	if err == nil {
		return true
	}
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		for _, part := range joined.Unwrap() {
			if !onlyProcessGone(part) {
				return false
			}
		}
		return true
	}
	return errors.Is(err, os.ErrProcessDone)
}
func TestManagedCancellationAndClose(t *testing.T) {
	for _, tt := range []struct {
		name, mode string
		delay      time.Duration
	}{{"06_awaiting_prepared", "wait-prepared", 100 * time.Millisecond}, {"08_between_write_and_ack", "no-ack", 200 * time.Millisecond}, {"22_exit_races_cancel", "exit-after", 100 * time.Millisecond}} {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			timer := time.AfterFunc(tt.delay, cancel)
			defer timer.Stop()
			defer cancel()
			_, g, _ := launchFixture(t, ctx, tt.mode, fixtureTiming)
			if g == nil {
				t.Fatal("not started")
			}
			assertReaped(t, g)
		})
	}
	t.Run("20_concurrent_close", func(t *testing.T) {
		_, g, err := launchFixture(t, context.Background(), "healthy", fixtureTiming)
		if err != nil {
			t.Fatal(err)
		}
		var wg sync.WaitGroup
		results := make([]error, 8)
		for i := range results {
			wg.Add(1)
			go func() { defer wg.Done(); results[i] = g.close(context.Background()) }()
		}
		wg.Wait()
		for _, got := range results {
			if fmt.Sprint(got) != fmt.Sprint(results[0]) {
				t.Fatal("outcome changed")
			}
		}
		assertReaped(t, g)
	})
	t.Run("05_cancel_before_spawn", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		c := &controller{}
		g, err := c.start(ctx, admitted{})
		if g != nil || !errors.Is(err, context.Canceled) {
			t.Fatal(g, err)
		}
	})
	t.Run("02_spawn_failure", func(t *testing.T) {
		a := fixtureAdmission(t, "healthy")
		a.executable = "/not/a/managed/executable"
		a.pins = nil
		c := &controller{}
		g, err := c.startWithTiming(context.Background(), a, fixtureTiming)
		if g != nil || err == nil {
			t.Fatal(g, err)
		}
	})
	t.Run("29_generation_not_revived", func(t *testing.T) {
		c, g, err := launchFixture(t, context.Background(), "healthy", fixtureTiming)
		if err != nil {
			t.Fatal(err)
		}
		assertReaped(t, g)
		if next, err := c.start(context.Background(), admitted{}); err == nil || next != nil {
			t.Fatal("owner restarted")
		}
		_, next, err := launchFixture(t, context.Background(), "healthy", fixtureTiming)
		if err != nil {
			t.Fatal(err)
		}
		if g.valid() {
			t.Fatal("old generation revived")
		}
		assertReaped(t, next)
	})
}
func TestFrameAndPipeLimits(t *testing.T) {
	for _, raw := range []string{`{"protocol":1,"phase":"activate","generation":1}`, `{"protocol":1,"phase":"activate","generation":18446744073709551615}`} {
		if _, err := parseFrame([]byte(raw)); err != nil {
			t.Fatal(err)
		}
	}
	for _, raw := range []string{`{"protocol":1,"phase":"activate","generation":1.0}`, `{"protocol":1,"phase":"activate","generation":1e0}`, `{"protocol":1,"phase":"activate","generation":18446744073709551616}`, `{"protocol":1,"phase":"activate","generation":null}`, `{"protocol":1,"phase":"activate","generation":1,"other":0}`, `{"protocol":1,"phase":"activate","generation":1,"generation":2}`} {
		if _, err := parseFrame([]byte(raw)); err == nil {
			t.Fatal(raw)
		}
	}
	t.Run("07_27_blocked_control_writer", func(t *testing.T) {
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		defer r.Close()
		defer w.Close()
		done := make(chan error, 1)
		if err := w.SetWriteDeadline(time.Now().Add(50 * time.Millisecond)); err != nil {
			t.Fatal(err)
		}
		n, fillErr := w.Write(make([]byte, 1<<20))
		if n == 0 || !errors.Is(fillErr, os.ErrDeadlineExceeded) {
			t.Fatalf("prefill n=%d error=%v", n, fillErr)
		}
		command, _ := frameBytes(frame{Protocol: 1, Phase: "activate", Generation: 1})
		go func() { done <- writeFrame(w, command, 50*time.Millisecond) }()
		select {
		case err := <-done:
			if !errors.Is(err, os.ErrDeadlineExceeded) {
				t.Fatal(err)
			}
		case <-time.After(time.Second):
			t.Fatal("writer stuck")
		}
	})
	t.Run("partial_header_survives_deadline", func(t *testing.T) {
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		defer r.Close()
		defer w.Close()
		bytes, _ := frameBytes(frame{Protocol: 1, Phase: "activate", Generation: 7})
		done := make(chan struct{})
		go func() {
			defer close(done)
			_, _ = w.Write(bytes[:2])
			time.Sleep(40 * time.Millisecond)
			_, _ = w.Write(bytes[2:])
		}()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		got, err := readFrame(ctx, r, 10*time.Millisecond)
		<-done
		if err != nil || got.Generation != 7 {
			t.Fatal(got, err)
		}
	})
}
func TestProductionTimingControl(t *testing.T) {
	if productionTiming != (timing{startup: 20 * time.Second, cleanup: 10 * time.Second, grace: 5 * time.Second, pipe: time.Second}) {
		t.Fatal(productionTiming)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	start := time.Now()
	_, g, err := launchFixture(t, ctx, "healthy", productionTiming)
	if err != nil {
		t.Fatal(err)
	}
	_ = g.close(ctx)
	elapsed := time.Since(start)
	if elapsed < 5*time.Second || elapsed >= 30*time.Second {
		t.Fatal(elapsed)
	}
	assertReaped(t, g)
}

// This holder is a second direct child of the harness, never a descendant of
// the managed fixture. Its control deadline guarantees an independent exit.
func TestManagedLogHolder(t *testing.T) {
	if len(os.Args) == 0 || os.Args[len(os.Args)-1] != "--managed-log-holder" {
		return
	}
	log := os.NewFile(3, "harness retained log writer")
	defer log.Close()
	control, err := inheritedPipe(4)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	_ = control.SetReadDeadline(time.Now().Add(3 * time.Second))
	var one [1]byte
	_, err = control.Read(one[:])
	_ = control.Close()
	if !errors.Is(err, io.EOF) {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
func TestHeldLogWriterIndependentReap(t *testing.T) {
	a := fixtureAdmission(t, "exit-after")
	args := childArguments(a, 1)
	listener, err := net.ListenTCP("tcp4", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	listenerFile, err := listener.File()
	if err != nil {
		t.Fatal(err)
	}
	defer listenerFile.Close()
	pipe := func() (*os.File, *os.File) {
		r, w, e := os.Pipe()
		if e != nil {
			t.Fatal(e)
		}
		t.Cleanup(func() { _ = r.Close(); _ = w.Close() })
		return r, w
	}
	cr, cw := pipe()
	rr, rw := pipe()
	outR, outW := pipe()
	errR, errW := pipe()
	holderR, holderW := pipe()
	holder := exec.Command(a.executable, "-test.run=^TestManagedLogHolder$", "--", "--managed-log-holder")
	holder.Env = []string{}
	holder.Dir = a.cwd
	holder.ExtraFiles = []*os.File{outW, holderR}
	if err = holder.Start(); err != nil {
		t.Fatal(err)
	}
	holderWaited := false
	defer func() {
		_ = holderW.Close()
		if !holderWaited {
			if err := holder.Wait(); err != nil {
				t.Error(err)
			}
		}
	}()
	cmd := exec.Command(a.executable, args...)
	cmd.Env = []string{}
	cmd.Dir = a.cwd
	cmd.Stdout = outW
	cmd.Stderr = errW
	cmd.ExtraFiles = []*os.File{listenerFile, cr, rw}
	if err = cmd.Start(); err != nil {
		t.Fatal(err)
	}
	for _, f := range []*os.File{listenerFile, cr, rw, outW, errW, holderR} {
		if err = f.Close(); err != nil {
			t.Fatal(err)
		}
	}
	g := &generation{cmd: cmd, process: recordOwner(cmdOwner{cmd}), sequence: 1, time: fixtureTiming, stop: make(chan struct{}), done: make(chan struct{}), waitDone: make(chan struct{}), pipes: processPipes{listener: listener, control: cw, report: rr, out: outR, stderr: errR}}
	g.expected = frame{Protocol: 1, Generation: 1, Profile: a.profile, Config: a.config, Registration: a.registration, Socket: listener.Addr().String()}
	ready := make(chan error, 1)
	go g.run(context.Background(), ready)
	t.Logf("case24 managed pid=%d separately owned holder pid=%d", cmd.Process.Pid, holder.Process.Pid)
	select {
	case err = <-ready:
		if err != nil && !errors.Is(err, io.EOF) {
			t.Error(err)
		}
	case <-time.After(3 * time.Second):
		t.Error("startup stalled")
	}
	assertReaped(t, g)
	g.stdout.mu.Lock()
	complete := g.stdout.complete
	g.stdout.mu.Unlock()
	if complete {
		t.Error("held log writer incorrectly reached EOF")
	}
	if holder.ProcessState != nil {
		t.Error("holder unexpectedly waited")
	}
	_ = holderW.Close()
	err = holder.Wait()
	holderWaited = true
	if err != nil {
		t.Fatal(err)
	}
}
func TestCleanupIncompleteRetainsOwner(t *testing.T) {
	a := fixtureAdmission(t, "exit-error")
	args := childArguments(a, 1)
	cmd := exec.Command(a.executable, args...)
	cmd.Dir = a.cwd
	cmd.Env = []string{}
	readyR, readyW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer readyR.Close()
	defer readyW.Close()
	cmd.ExtraFiles = []*os.File{readyW}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	if err = readyW.Close(); err != nil {
		t.Fatal(err)
	}
	t.Logf("case23 owned child pid=%d", cmd.Process.Pid)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_ = ctx
	g := &generation{cmd: cmd, process: recordOwner(cmdOwner{cmd}), time: timing{cleanup: 5 * time.Second, grace: 2 * time.Second}, waitDone: make(chan struct{}), cancel: cancel}
	var keep []*os.File
	pipe := func() *os.File {
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		keep = append(keep, r, w)
		return r
	}
	defer func() {
		for _, f := range keep {
			_ = f.Close()
		}
	}()
	g.pipes.control = pipe()
	g.pipes.report = pipe()
	g.pipes.out = pipe()
	g.pipes.stderr = pipe()
	for i := range g.workers {
		g.workers[i] = make(chan struct{})
	}
	if err = readyR.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
		t.Fatal(err)
	}
	var ready [1]byte
	_, readyErr := io.ReadFull(readyR, ready[:])
	if readyErr != nil || ready[0] != 1 {
		t.Error("fixture never reached explicit exit path", readyErr)
	}
	// This is a controlled worker-join failure, not a claim to induce an
	// uninterruptible kernel Wait. The real child is still waited exactly once.
	err = g.cleanup(errProtocol)
	if !errors.Is(err, errCleanup) || !errors.Is(err, errProtocol) {
		t.Fatal(err)
	}
	var exitError *exec.ExitError
	if !errors.As(err, &exitError) {
		t.Error("completed Wait result dropped by worker join timeout", err)
	}

	select {
	case <-g.waitDone:
	case <-time.After(time.Second):
		t.Fatal("real child not reaped")
	}
	c := &controller{current: g}
	if next, err := c.start(context.Background(), a); err == nil || next != nil {
		t.Fatal("unresolved owner restarted")
	}
	for _, done := range g.workers {
		close(done)
	}
}
func TestParentControlLoss(t *testing.T) {
	_, g, err := launchFixture(t, context.Background(), "healthy", fixtureTiming)
	if err != nil {
		t.Fatal(err)
	}
	if err = g.pipes.control.Close(); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(time.Second)
	for g.valid() && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if g.valid() {
		t.Error("case30 control loss did not invalidate")
	}
	assertReaped(t, g)
}
func TestInheritedDescriptorKind(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "ordinary-file")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	fd, err := syscall.Dup(int(f.Fd()))
	if err != nil {
		t.Fatal(err)
	}
	if pipe, err := inheritedPipe(uintptr(fd)); err == nil {
		_ = pipe.Close()
		t.Fatal("case28 regular file admitted as inherited pipe")
	}
}

func TestAcknowledgmentCannotBypassBlockedActivation(t *testing.T) {
	for _, mode := range []string{"premature-ack", "premature-eof", "premature-extra"} {
		t.Run(mode, func(t *testing.T) { testBlockedActivation(t, mode) })
	}
}
func testBlockedActivation(t *testing.T, mode string) {
	t.Helper()
	a := fixtureAdmission(t, mode)
	cmd, pipes, err := spawn(a, 1)
	if err != nil {
		t.Fatal(err)
	}
	// The fixture writes both reports without reading control. Fill that exact
	// parent writer before starting its lifecycle to make activation block.
	if err = pipes.control.SetWriteDeadline(time.Now().Add(50 * time.Millisecond)); err != nil {
		t.Fatal(err)
	}
	n, err := pipes.control.Write(make([]byte, 1<<20))
	if n == 0 || !errors.Is(err, os.ErrDeadlineExceeded) {
		t.Fatal(n, err)
	}
	tm := fixtureTiming
	// Keep the write blocked long enough for the explicit report loss to arrive.
	// All variants still use the fixed 3s startup/5s cleanup fixture envelope.
	if mode != "premature-ack" {
		tm.pipe = time.Second
	}
	g := &generation{cmd: cmd, process: recordOwner(cmdOwner{cmd}), pipes: pipes, sequence: 1, time: tm, stop: make(chan struct{}), done: make(chan struct{}), waitDone: make(chan struct{})}
	g.expected = frame{Protocol: 1, Generation: 1, Profile: a.profile, Config: a.config, Registration: a.registration, Socket: pipes.listener.Addr().String()}
	ready := make(chan error, 1)
	go g.run(context.Background(), ready)
	t.Logf("case7 owned child pid=%d with measured control backpressure", cmd.Process.Pid)
	select {
	case err = <-ready:
		if err == nil {
			t.Error("ack published before activation write completed")
		}
		if mode == "premature-eof" && !errors.Is(err, io.EOF) {
			t.Error("lost EOF cause", err)
		}
		if mode == "premature-extra" && !errors.Is(err, errProtocol) {
			t.Error("lost extra-byte cause", err)
		}
	case <-time.After(3 * time.Second):
		t.Error("start did not terminate")
	}
	assertReaped(t, g)
	if g.valid() {
		t.Fatal("blocked activation exposed serving")
	}
}

type trustInfo struct {
	os.FileInfo
	mode os.FileMode
	uid  uint32
}

func (i trustInfo) Mode() os.FileMode { return i.mode }
func (i trustInfo) IsDir() bool       { return i.mode.IsDir() }
func (i trustInfo) Sys() any          { return &syscall.Stat_t{Uid: i.uid} }
func TestRoleAwarePathTrust(t *testing.T) {
	info, err := os.Lstat(testRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	uid := uint32(os.Geteuid())
	for _, tt := range []struct {
		name                        string
		uid                         uint32
		mode                        os.FileMode
		ancestor, artifact, mutable bool
	}{
		{"private", uid, os.ModeDir | 0700, true, true, true},
		{"system", 0, os.ModeDir | 0755, true, true, uid == 0},
		{"root-sticky", 0, os.ModeDir | os.ModeSticky | 0777, true, false, false},
		{"owner-sticky", uid, os.ModeDir | os.ModeSticky | 0777, true, false, false},
		{"root-writable", 0, os.ModeDir | 0777, false, false, false},
		{"owner-writable", uid, os.ModeDir | 0777, false, false, false},
		{"foreign", ^uint32(0), os.ModeDir | 0700, false, false, false},
		{"foreign-sticky", ^uint32(0), os.ModeDir | os.ModeSticky | 0777, false, false, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			f := trustInfo{FileInfo: info, uid: tt.uid, mode: tt.mode}
			if trustedAncestor(f) != tt.ancestor || protectedArtifact(f) != tt.artifact || trustedMutableDirectory(f) != tt.mutable {
				t.Fatal("role trust mismatch")
			}
		})
	}
	// A real system artifact is read and hashed only; it is never executed.
	path, err := filepath.EvalSymlinks("/usr/bin/true")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = hashFile(context.Background(), time.Now().Add(time.Second), path, true, &budget{}); err != nil {
		t.Fatal("host prerequisite: /usr/bin/true must resolve to a protected artifact under trusted parents", err)
	}
	root := testRoot(t)
	writable := filepath.Join(root, "sticky")
	if err = os.Mkdir(writable, 0700); err != nil {
		t.Fatal(err)
	}
	if err = os.Chmod(writable, os.ModeSticky|0777); err != nil {
		t.Fatal(err)
	}
	leaf := filepath.Join(writable, "private")
	if err = os.Mkdir(leaf, 0700); err != nil {
		t.Fatal(err)
	}
	if _, err = decodeCanonical(context.Background(), time.Now().Add(time.Second), encoded(leaf), &budget{}); err != nil {
		t.Fatal("sticky ancestor refused", err)
	}
	if _, err = trustedDirectory(writable); !errors.Is(err, errInput) {
		t.Fatal("sticky mutable leaf admitted", err)
	}
	if err = os.Chmod(writable, 0777); err != nil {
		t.Fatal(err)
	}
	if _, err = decodeCanonical(context.Background(), time.Now().Add(time.Second), encoded(leaf), &budget{}); !errors.Is(err, errInput) {
		t.Fatal("writable nonsticky ancestor admitted", err)
	}
}

func recordsCommand(recorder *recordedProcess, cmd *exec.Cmd) bool {
	if recorder == nil || cmd == nil {
		return false
	}
	owner, ok := recorder.inner.(cmdOwner)
	return ok && owner.cmd == cmd
}

// These are unstarted command values: no extra child or signal is needed to
// discriminate a recorder that observes a different or nondelegating owner.
func TestRecorderIdentifiesExactCommand(t *testing.T) {
	cmd, other := &exec.Cmd{}, &exec.Cmd{}
	for _, tt := range []struct {
		name  string
		inner processOwner
		want  bool
	}{
		{"same", cmdOwner{cmd}, true},
		{"different", cmdOwner{other}, false},
		{"nil-command", cmdOwner{}, false},
		{"nondelegating", &recordedProcess{}, false},
		{"nil-owner", nil, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if recordsCommand(&recordedProcess{inner: tt.inner}, cmd) != tt.want {
				t.Fatal("recorder identity mismatch")
			}
		})
	}
	if recordsCommand(nil, cmd) || recordsCommand(&recordedProcess{inner: cmdOwner{cmd}}, nil) {
		t.Fatal("nil recorder or expected command accepted")
	}
}
