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
// descendants, and accepts only its private test switches after --.
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
	if mode == "ignore-term" {
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
	if mode == "premature-ack" {
		ack := prepared
		ack.Phase = "activated"
		bytes, err := frameBytes(ack)
		if err != nil {
			return err
		}
		if err = writeFrame(report, bytes, time.Second); err != nil {
			return err
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
	var extra [1]byte
	err = readExact(context.Background(), control, extra[:], time.Second)
	if errors.Is(err, io.EOF) {
		return nil
	}
	return err
}

var fixtureTiming = timing{startup: 3 * time.Second, cleanup: 5 * time.Second, grace: 50 * time.Millisecond, pipe: 100 * time.Millisecond}

func fixtureAdmission(t *testing.T) (admitted, []string) {
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
	return a, []string{"-test.run=^TestManagedFixture$", "--", "--profile=" + a.profile, "--config=" + a.config, "--registration=" + a.registration}
}
func filepathCanonical(p string) (string, error) { return filepath.EvalSymlinks(p) }
func launchFixture(t *testing.T, ctx context.Context, mode string, tm timing) (*controller, *generation, error) {
	t.Helper()
	a, args := fixtureAdmission(t)
	args = append(args, "--managed-fixture="+mode)
	c := &controller{}
	g, err := c.startWithTiming(ctx, a, args, tm)
	if g != nil {
		t.Logf("owned child pid=%d mode=%s", g.cmd.Process.Pid, mode)
		t.Cleanup(func() {
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
			defer cancel()
			_ = g.close(cleanupCtx)
			select {
			case <-g.waitDone:
			default:
				t.Errorf("child %d not reaped", g.cmd.Process.Pid)
			}
		})
	}
	return c, g, err
}
func assertReaped(t *testing.T, g *generation) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = g.close(ctx)
	select {
	case <-g.waitDone:
	default:
		t.Fatal("not reaped")
	}
	if !slices.Equal(g.events, []string{"term", "kill", "signals-ended", "wait"}) {
		t.Fatal(g.events)
	}
	if g.valid() {
		t.Fatal("still valid")
	}
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
		{"18_control_eof", "healthy", true},
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
			_, g, err := launchFixture(t, ctx, tt.mode, fixtureTiming)
			if (err == nil) != tt.healthy {
				t.Fatalf("healthy=%v err=%v", tt.healthy, err)
			}
			if g == nil {
				t.Fatal("no child")
			}
			if tt.mode == "healthy" {
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
			assertReaped(t, g)
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
		g, err := c.start(ctx, admitted{}, nil)
		if g != nil || !errors.Is(err, context.Canceled) {
			t.Fatal(g, err)
		}
	})
	t.Run("02_spawn_failure", func(t *testing.T) {
		a, _ := fixtureAdmission(t)
		a.executable = "/not/a/managed/executable"
		a.pins = nil
		c := &controller{}
		g, err := c.startWithTiming(context.Background(), a, nil, fixtureTiming)
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
		if next, err := c.start(context.Background(), admitted{}, nil); err == nil || next != nil {
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
	a, args := fixtureAdmission(t)
	args = append(args, "--managed-fixture=exit-after", "--managed-protocol=1", "--generation=1")
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
	g := &generation{cmd: cmd, sequence: 1, time: fixtureTiming, stop: make(chan struct{}), done: make(chan struct{}), waitDone: make(chan struct{}), pipes: processPipes{listener: listener, control: cw, report: rr, out: outR, stderr: errR}}
	g.expected = frame{Protocol: 1, Generation: 1, Profile: a.profile, Config: a.config, Registration: a.registration, Socket: listener.Addr().String()}
	ready := make(chan error, 1)
	go g.run(context.Background(), ready)
	t.Logf("case24 managed pid=%d separately owned holder pid=%d", cmd.Process.Pid, holder.Process.Pid)
	select {
	case err = <-ready:
		if err != nil {
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
	a, args := fixtureAdmission(t)
	args = append(args, "--managed-fixture=exit-error")
	cmd := exec.Command(a.executable, args...)
	cmd.Dir = a.cwd
	cmd.Env = []string{}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Logf("case23 owned child pid=%d", cmd.Process.Pid)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_ = ctx
	g := &generation{cmd: cmd, time: timing{cleanup: 100 * time.Millisecond, grace: 10 * time.Millisecond}, waitDone: make(chan struct{}), cancel: cancel}
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
	// This is a controlled worker-join failure, not a claim to induce an
	// uninterruptible kernel Wait. The real child is still waited exactly once.
	err := g.cleanup(errProtocol)
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
	if next, err := c.start(context.Background(), a, args); err == nil || next != nil {
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
	a, args := fixtureAdmission(t)
	args = append(args, "--managed-fixture=premature-ack", "--managed-protocol=1", "--generation=1")
	cmd, pipes, err := spawn(a, args)
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
	g := &generation{cmd: cmd, pipes: pipes, sequence: 1, time: fixtureTiming, stop: make(chan struct{}), done: make(chan struct{}), waitDone: make(chan struct{})}
	g.expected = frame{Protocol: 1, Generation: 1, Profile: a.profile, Config: a.config, Registration: a.registration, Socket: pipes.listener.Addr().String()}
	ready := make(chan error, 1)
	go g.run(context.Background(), ready)
	t.Logf("case7 owned child pid=%d with measured control backpressure", cmd.Process.Pid)
	select {
	case err = <-ready:
		if err == nil {
			t.Error("ack published before activation write completed")
		}
	case <-time.After(3 * time.Second):
		t.Error("start did not terminate")
	}
	assertReaped(t, g)
}
