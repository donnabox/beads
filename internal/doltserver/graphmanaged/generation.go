package graphmanaged

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

const maxFrame = 65536

var productionTiming = timing{startup: 20 * time.Second, cleanup: 10 * time.Second, grace: 5 * time.Second, pipe: time.Second}

type timing struct{ startup, cleanup, grace, pipe time.Duration }
type frame struct {
	Protocol     uint64 `json:"protocol"`
	Phase        string `json:"phase"`
	Generation   uint64 `json:"generation"`
	Profile      string `json:"profile,omitempty"`
	Config       string `json:"config,omitempty"`
	Registration string `json:"registration,omitempty"`
	Socket       string `json:"socket,omitempty"`
}

func frameBytes(f frame) ([]byte, error) {
	payload, err := json.Marshal(f)
	if err != nil {
		return nil, err
	}
	if len(payload) == 0 || len(payload) > maxFrame {
		return nil, errProtocol
	}
	out := make([]byte, 4+len(payload))
	binary.BigEndian.PutUint32(out, uint32(len(payload)))
	copy(out[4:], payload)
	return out, nil
}
func parseFrame(payload []byte) (frame, error) {
	var f frame
	if len(payload) == 0 || len(payload) > maxFrame {
		return f, errProtocol
	}
	b := budget{}
	if err := validateJSON(payload, &b); err != nil {
		return f, errors.Join(errProtocol, err)
	}
	// Unmarshal through a closed map additionally rejects null rather than treating
	// it as a zero/default value. Activation has exactly three fields.
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(payload, &fields); err != nil || fields == nil {
		return f, errProtocol
	}
	for key, value := range fields {
		switch key {
		case "protocol", "phase", "generation", "profile", "config", "registration", "socket":
		default:
			return f, errProtocol
		}
		if string(value) == "null" {
			return f, errProtocol
		}
	}
	if err := json.Unmarshal(payload, &f); err != nil {
		return f, errProtocol
	}
	if f.Protocol != 1 || f.Generation == 0 {
		return f, errProtocol
	}
	switch f.Phase {
	case "activate":
		if len(fields) != 3 {
			return f, errProtocol
		}
	case "prepared", "activated":
		if len(fields) != 7 || !digestValid(f.Profile) || !digestValid(f.Config) || !digestValid(f.Registration) || f.Socket == "" {
			return f, errProtocol
		}
	default:
		return f, errProtocol
	}
	return f, nil
}

func readExact(ctx context.Context, f *os.File, buf []byte, interval time.Duration) error {
	offset := 0
	for offset < len(buf) {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := f.SetReadDeadline(time.Now().Add(interval)); err != nil {
			return err
		}
		n, err := f.Read(buf[offset:])
		offset += n
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				continue
			}
			if offset == len(buf) {
				return nil
			}
			return err
		}
		if n == 0 {
			return io.ErrNoProgress
		}
	}
	return nil
}
func readFrame(ctx context.Context, f *os.File, interval time.Duration) (frame, error) {
	var header [4]byte
	if err := readExact(ctx, f, header[:], interval); err != nil {
		return frame{}, err
	}
	size := binary.BigEndian.Uint32(header[:])
	if size == 0 || size > maxFrame {
		return frame{}, errProtocol
	}
	payload := make([]byte, size)
	if err := readExact(ctx, f, payload, interval); err != nil {
		return frame{}, err
	}
	return parseFrame(payload)
}
func writeFrame(f *os.File, bytes []byte, interval time.Duration) error {
	if err := f.SetWriteDeadline(time.Now().Add(interval)); err != nil {
		return err
	}
	for len(bytes) > 0 {
		n, err := f.Write(bytes)
		bytes = bytes[n:]
		if err != nil {
			return err
		}
		if n == 0 {
			return io.ErrNoProgress
		}
	}
	return nil
}

type logTail struct {
	mu        sync.Mutex
	bytes     []byte
	discarded uint64
	err       error
	complete  bool
}

func (l *logTail) append(p []byte) {
	l.mu.Lock()
	defer l.mu.Unlock()
	overflow := max(0, len(l.bytes)+len(p)-(64<<10))
	if uint64(overflow) > ^uint64(0)-l.discarded {
		l.discarded = ^uint64(0)
	} else {
		l.discarded += uint64(overflow)
	}
	if len(p) >= 64<<10 {
		l.bytes = append(l.bytes[:0], p[len(p)-(64<<10):]...)
		return
	}
	if overflow > 0 {
		copy(l.bytes, l.bytes[overflow:])
		l.bytes = l.bytes[:len(l.bytes)-overflow]
	}
	l.bytes = append(l.bytes, p...)
}
func (l *logTail) drain(f *os.File, done chan<- struct{}) {
	defer close(done)
	buf := make([]byte, 32<<10)
	for {
		n, err := f.Read(buf)
		if n > 0 {
			l.append(buf[:n])
		}
		if err != nil {
			l.mu.Lock()
			l.complete = errors.Is(err, io.EOF)
			if !l.complete && !errors.Is(err, os.ErrClosed) {
				l.err = err
			}
			l.mu.Unlock()
			return
		}
	}
}

type processPipes struct {
	startupError                 error
	listener                     net.Listener
	control, report, out, stderr *os.File
}
type generation struct {
	mu                 sync.Mutex
	cmd                *exec.Cmd
	process            processOwner
	pipes              processPipes
	sequence           uint64
	expected           frame
	time               timing
	serving            bool
	cause, errorResult error
	stop               chan struct{}
	stopOnce           sync.Once
	done, waitDone     chan struct{}
	waitResult         error
	stdout, stderr     logTail
	workers            [4]chan struct{}
	cancel             context.CancelFunc
}

// controller retains unresolved child ownership. It deliberately has no reset,
// replacement or production constructor; another start on this owner refuses.
// Host exclusive-wait and adapter non-forking source gates remain external work.
type controller struct {
	mu       sync.Mutex
	current  *generation
	sequence uint64
	// A scoped test decorator observes the real owned process boundaries.
	// Production leaves it nil. The test harness verifies that the decorator
	// delegates to this generation's exact cmdOwner; this type alone does not.
	observe func(processOwner) processOwner
}

type processOwner interface {
	Signal(os.Signal) error
	Kill() error
	Wait() error
}
type cmdOwner struct{ cmd *exec.Cmd }

func (p cmdOwner) Signal(s os.Signal) error { return p.cmd.Process.Signal(s) }
func (p cmdOwner) Kill() error              { return p.cmd.Process.Kill() }
func (p cmdOwner) Wait() error              { return p.cmd.Wait() }

func childArguments(a admitted, sequence uint64) []string {
	args := append([]string(nil), a.argv...)
	return append(args, "--managed-protocol=1", "--generation="+strconv.FormatUint(sequence, 10))
}

func (c *controller) start(ctx context.Context, a admitted) (*generation, error) {
	return c.startWithTiming(ctx, a, productionTiming)
}
func (c *controller) startWithTiming(ctx context.Context, a admitted, t timing) (*generation, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.current != nil || c.sequence == ^uint64(0) {
		return nil, errInput
	}
	if !platformSupported() {
		return nil, errUnsupported
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := a.recheck(ctx); err != nil {
		return nil, err
	}
	c.sequence++
	g := &generation{sequence: c.sequence, time: t, stop: make(chan struct{}), done: make(chan struct{}), waitDone: make(chan struct{})}
	g.expected = frame{Protocol: 1, Generation: g.sequence, Profile: a.profile, Config: a.config, Registration: a.registration}
	cmd, pipes, err := spawn(a, c.sequence)
	if err != nil {
		return nil, err
	}
	g.cmd = cmd
	g.process = cmdOwner{cmd}
	if c.observe != nil {
		g.process = c.observe(g.process)
	}
	g.pipes = pipes
	g.expected.Socket = pipes.listener.Addr().String()
	c.current = g
	ready := make(chan error, 1)
	go g.run(ctx, ready)
	err = <-ready
	return g, err
}
func (g *generation) valid() bool { g.mu.Lock(); defer g.mu.Unlock(); return g.serving }
func (g *generation) close(ctx context.Context) error {
	g.stopOnce.Do(func() { close(g.stop) })
	select {
	case <-g.done:
		g.mu.Lock()
		defer g.mu.Unlock()
		return g.errorResult
	case <-ctx.Done():
		return errors.Join(errCleanup, ctx.Err())
	}
}
func (g *generation) matches(f frame, phase string) bool {
	expected := g.expected
	expected.Phase = phase
	return f == expected
}

type reportEvent struct {
	f   frame
	err error
}

func (g *generation) run(ctx context.Context, ready chan<- error) {
	workCtx, cancel := context.WithCancel(context.Background())
	g.cancel = cancel
	for i := range g.workers {
		g.workers[i] = make(chan struct{})
	}
	go g.stdout.drain(g.pipes.out, g.workers[0])
	go g.stderr.drain(g.pipes.stderr, g.workers[1])
	reports := make(chan reportEvent, 1)
	go func() {
		defer close(g.workers[2])
		for i := 0; i < 2; i++ {
			f, err := readFrame(workCtx, g.pipes.report, g.time.pipe)
			select {
			case reports <- reportEvent{f, err}:
			case <-workCtx.Done():
				return
			}
			if err != nil {
				return
			}
		}
		var extra [1]byte
		err := readExact(workCtx, g.pipes.report, extra[:], g.time.pipe)
		if err == nil {
			err = errProtocol
		}
		select {
		case reports <- reportEvent{err: err}:
		case <-workCtx.Done():
		}
	}()
	commands := make(chan []byte, 1)
	writes := make(chan error, 1)
	go func() {
		defer close(g.workers[3])
		select {
		case bytes := <-commands:
			err := writeFrame(g.pipes.control, bytes, g.time.pipe)
			select {
			case writes <- err:
			case <-workCtx.Done():
			}
		case <-workCtx.Done():
		}
	}()
	startupDeadline := time.Now().Add(g.time.startup)
	timer := time.NewTimer(g.time.startup)
	defer timer.Stop()
	phase := "prepared"
	published := false
	activationWritten, acknowledged := false, false
	cause := g.pipes.startupError
lifecycle:
	for cause == nil {
		select {
		case <-ctx.Done():
			cause = ctx.Err()
			break lifecycle
		case <-g.stop:
			if !published && cause == nil {
				cause = errInput
			}
			break lifecycle
		case <-timer.C:
			cause = context.DeadlineExceeded
			break lifecycle
		case err := <-writes:
			if err != nil {
				cause = err
				break lifecycle
			}
			activationWritten = true
		case event := <-reports:
			if event.err != nil {
				cause = event.err
				break lifecycle
			}
			if !g.matches(event.f, phase) {
				cause = errProtocol
				break lifecycle
			}
			if phase == "prepared" {
				phase = "activated"
				bytes, err := frameBytes(frame{Protocol: 1, Phase: "activate", Generation: g.sequence})
				if err != nil {
					cause = err
					break lifecycle
				}
				commands <- bytes

			} else {
				acknowledged = true
			}
		}
		// Neither a successful write nor a matching acknowledgment alone publishes
		// startup. In particular, an early acknowledgment cannot outrun a blocked
		// or partially failed activation write.
		if !published && activationWritten && acknowledged {
			if ctx.Err() != nil {
				cause = ctx.Err()
				break lifecycle
			}
			if !time.Now().Before(startupDeadline) || !timer.Stop() {
				cause = context.DeadlineExceeded
				break lifecycle
			}
			if err := g.pipes.listener.Close(); err != nil {
				cause = err
				break lifecycle
			}
			g.pipes.listener = nil
			g.mu.Lock()
			g.serving = true
			g.mu.Unlock()
			published = true
			ready <- nil
		}
	}

	g.mu.Lock()
	g.serving = false
	g.cause = cause
	g.mu.Unlock()
	result := g.cleanup(cause)
	g.mu.Lock()
	g.errorResult = result
	g.mu.Unlock()
	close(g.done)
	if !published {
		ready <- result
	}
}
func (g *generation) cleanup(cause error) error {
	deadline := time.Now().Add(g.time.cleanup)
	result := errors.Join(cause, g.pipes.control.Close())
	if g.pipes.listener != nil {
		result = errors.Join(result, g.pipes.listener.Close())
	}
	// This coordinator is the ONLY signaller. Wait has not started, including
	// when the report pipe already reached EOF. No goroutine uses CommandContext.
	result = errors.Join(result, signalTerm(g.process))
	timer := time.NewTimer(g.time.grace)
	<-timer.C
	result = errors.Join(result, g.process.Kill())
	go func() { g.waitResult = g.process.Wait(); close(g.waitDone) }()
	// Nothing after this line may signal, even if waiting exceeds our envelope.
	g.cancel()
	for _, f := range []*os.File{g.pipes.report, g.pipes.out, g.pipes.stderr} {
		result = errors.Join(result, f.Close())
	}
	remaining := time.Until(deadline)
	if remaining <= 0 {
		return g.observedCleanup(errors.Join(result, errCleanup))
	}
	timer = time.NewTimer(remaining)
	defer timer.Stop()
	for _, done := range append([]chan struct{}{g.waitDone}, g.workers[:]...) {
		select {
		case <-done:
		case <-timer.C:
			return g.observedCleanup(errors.Join(result, errCleanup))
		}
	}

	return g.observedCleanup(result)
}

// A join timeout must not erase outcomes that are already available. The
// closed waitDone channel synchronizes the sole waiter's result; an unfinished
// Wait remains owned without reading its result concurrently.
func (g *generation) observedCleanup(result error) error {
	select {
	case <-g.waitDone:
		result = errors.Join(result, g.waitResult)
	default:
	}
	for _, tail := range []*logTail{&g.stdout, &g.stderr} {
		tail.mu.Lock()
		result = errors.Join(result, tail.err)
		tail.mu.Unlock()
	}
	return result
}

func (g *generation) String() string { return fmt.Sprintf("managed generation %d", g.sequence) }
