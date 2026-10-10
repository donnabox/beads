//go:build unix

package graphsession

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
)

// These are actual driver/loopback protocol controls, not an engine. The
// process boundary confines eligible LOCAL INFILE registrations to test children.
const compositionPhase = 5 * time.Second

type compositionReceipt struct {
	Capabilities     uint32   `json:"client_capabilities"`
	Commands         []string `json:"commands"`
	Request          string   `json:"local_infile_request,omitempty"`
	PostRequestBytes int64    `json:"post_request_bytes"`
	Terminal         string   `json:"terminal"`
	Quit             bool     `json:"quit"`
	Error            string   `json:"error,omitempty"`
}
type compositionPeer struct {
	listener   net.Listener
	mu         sync.Mutex
	conns      []net.Conn
	receipts   []compositionReceipt
	failures   error
	stopping   bool
	acceptDone chan struct{}
	workers    sync.WaitGroup
	joined     bool
	once       sync.Once
	done       chan struct{}
	offered    bool
	inject     string
	name       string
	rowValue   bool
}

func newCompositionPeer(t *testing.T, offered bool, inject, name string, rowValue bool) (*compositionPeer, endpoint) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	p := &compositionPeer{listener: listener, offered: offered, inject: inject, name: name, rowValue: rowValue, acceptDone: make(chan struct{}), done: make(chan struct{})}
	// Registered before workers: assertion failures still close and join everything.
	t.Cleanup(func() { p.stop(t) })
	go p.accept()
	return p, endpoint{address: listener.Addr().String(), base: "sales", branch: "main", user: "fixture"}
}
func (p *compositionPeer) record(err error) {
	if err != nil {
		p.mu.Lock()
		p.failures = errors.Join(p.failures, err)
		p.mu.Unlock()
	}
}
func (p *compositionPeer) accept() {
	defer close(p.acceptDone)
	for {
		c, err := p.listener.Accept()
		if err != nil {
			p.mu.Lock()
			expected := p.stopping && errors.Is(err, net.ErrClosed)
			p.mu.Unlock()
			if !expected {
				p.record(err)
			}
			return
		}
		p.mu.Lock()
		p.conns = append(p.conns, c)
		count := len(p.conns)
		p.mu.Unlock()
		if count != 1 {
			p.record(errors.New("unexpected second connection"))
			p.record(c.Close())
			continue
		}
		p.workers.Add(1)
		go func() {
			defer p.workers.Done()
			receipt, err := p.serve(c)
			if err != nil {
				receipt.Error = err.Error()
				p.record(err)
			}
			p.record(c.Close())
			p.mu.Lock()
			p.receipts = append(p.receipts, receipt)
			p.mu.Unlock()
			close(p.done)
		}()
	}
}
func (p *compositionPeer) serve(c net.Conn) (r compositionReceipt, err error) {
	if err = c.SetDeadline(time.Now().Add(compositionPhase)); err != nil {
		return
	}
	r.Capabilities, err = handshakeCapabilities(c, p.offered)
	if err != nil {
		return
	}
	for {
		if err = c.SetDeadline(time.Now().Add(compositionPhase)); err != nil {
			return
		}
		var b []byte
		b, err = readPacket(c)
		if err != nil {
			return
		}
		if len(b) == 1 && b[0] == 1 {
			r.Quit = true
			r.Terminal = "COM_QUIT"
			return r, nil
		}
		if len(b) < 2 || b[0] != 3 {
			return r, errors.New("unexpected command packet")
		}
		q := string(b[1:])
		r.Commands = append(r.Commands, q)
		if (p.inject == "initialize" && strings.HasPrefix(q, "SELECT @@autocommit")) || (p.inject == "commit" && q == "COMMIT") {
			r.Request = p.name
			if err = packet(c, 1, append([]byte{0xfb}, []byte(p.name)...)); err != nil {
				return
			}
			// EOF/reset must actually follow the sent request. Timeouts never pass.
			r.PostRequestBytes, err = io.CopyN(io.Discard, c, (1<<20)+1)
			switch {
			case errors.Is(err, io.EOF):
				r.Terminal = "EOF"
				err = nil
			case errors.Is(err, syscall.ECONNRESET):
				r.Terminal = "ECONNRESET"
				err = nil
			case err == nil:
				err = errors.New("post-request upload exceeded fixture bound")
			}
			return
		}
		response := defaultReply(q)
		if p.rowValue && strings.HasPrefix(q, "CALL DOLT_MERGE") {
			response = reply{rows: [][]string{{"\xfbordinary"}}}
		}
		if err = sendReply(c, response); err != nil {
			return
		}
	}
}
func (p *compositionPeer) stop(t *testing.T) {
	t.Helper()
	p.once.Do(func() {
		p.mu.Lock()
		p.stopping = true
		p.mu.Unlock()
		p.record(p.listener.Close())
		select {
		case <-p.acceptDone:
		case <-time.After(time.Second):
			t.Error("accept worker not joined")
		}
		p.mu.Lock()
		conns := append([]net.Conn(nil), p.conns...)
		p.mu.Unlock()
		for _, c := range conns {
			if err := c.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
				p.record(err)
			}
		}
		joined := make(chan struct{})
		go func() { p.workers.Wait(); close(joined) }()
		select {
		case <-joined:
			p.joined = true
		case <-time.After(time.Second):
			t.Error("protocol worker not joined")
		}
	})
}
func (p *compositionPeer) finish(t *testing.T) compositionReceipt {
	t.Helper()
	select {
	case <-p.done:
	case <-time.After(compositionPhase):
		t.Error("protocol terminality missing; forced cleanup fails")
	}
	p.stop(t)
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.failures != nil {
		t.Errorf("peer failures: %v", p.failures)
	}
	if len(p.conns) != 1 || len(p.receipts) != 1 {
		t.Fatalf("connections=%d receipts=%d", len(p.conns), len(p.receipts))
	}
	r := p.receipts[0]
	if r.Capabilities&0x80 != 0 {
		t.Error("actual client handshake advertised LOCAL INFILE")
	}
	b, err := json.Marshal(struct {
		Accepted int  `json:"accepted"`
		Joined   bool `json:"workers_joined"`
		compositionReceipt
	}{len(p.conns), p.joined, r})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("composition_protocol=%s", b)
	return r
}
func compositionContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), compositionPhase)
	t.Cleanup(cancel)
	return ctx
}
func assertCompositionFailure(t *testing.T, s *session, err error, cleanupCause error) {
	t.Helper()
	if !errors.Is(err, mysql.ErrLocalInfileDisabled) || !errors.Is(err, errUncertain) || errors.Is(err, driver.ErrBadConn) {
		t.Fatalf("lost/non-replayable primary: %v", err)
	}
	if cleanupCause != nil && !errors.Is(err, cleanupCause) {
		t.Fatalf("lost observed close error: %v", err)
	}
	if s.state != discarded || s.db.Stats().OpenConnections != 0 {
		t.Fatal("failed owner/pool not discarded")
	}
	ctx := compositionContext(t)
	_, again := s.execute(ctx, commitOperation, "")
	for _, e := range []error{again, s.acquire(ctx, ioBound), s.refresh(ctx), s.abort(ctx), s.finish(ctx)} {
		if !errors.Is(e, errState) {
			t.Fatalf("discarded operation accepted: %v", e)
		}
	}
	if e := s.close(); !errors.Is(e, mysql.ErrLocalInfileDisabled) || !errors.Is(e, errUncertain) || (cleanupCause != nil && !errors.Is(e, cleanupCause)) {
		t.Fatalf("repeat close lost failure: %v", e)
	}
}
func compositionAdversary(t *testing.T, stage, kind string, offered bool) {
	t.Helper()
	var calls atomic.Int32
	var name string
	if kind == "reader" {
		key := fmt.Sprintf("composition-%d", os.Getpid())
		name = "Reader::" + key
		mysql.RegisterReaderHandler(key, func() io.Reader { calls.Add(1); return strings.NewReader("registered-reader-marker") })
		t.Cleanup(func() { mysql.DeregisterReaderHandler(key) })
	} else {
		name = filepath.Join(t.TempDir(), "registered-file")
		if err := os.WriteFile(name, []byte("registered-file-marker"), 0600); err != nil {
			t.Fatal(err)
		}
		mysql.RegisterLocalFile(name)
		t.Cleanup(func() { mysql.DeregisterLocalFile(name) })
	}
	p, e := newCompositionPeer(t, offered, stage, name, false)
	ctx := compositionContext(t)
	s, err := open(ctx, e)
	if stage == "initialize" {
		if s != nil || !errors.Is(err, mysql.ErrLocalInfileDisabled) || errors.Is(err, driver.ErrBadConn) {
			t.Fatalf("initialization admitted/lost primary: %v", err)
		}
	} else {
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if e := s.close(); e != nil && !errors.Is(e, mysql.ErrLocalInfileDisabled) {
				t.Errorf("cleanup: %v", e)
			}
		})
		if err = s.acquire(ctx, ioBound); err != nil {
			t.Fatal(err)
		}
		_, err = s.execute(ctx, commitOperation, "")
		assertCompositionFailure(t, s, err, nil)
	}
	r := p.finish(t)
	if r.Request != name || r.PostRequestBytes != 0 || r.Quit || r.Terminal == "" || calls.Load() != 0 {
		t.Fatalf("infile leaked/admitted: %+v calls=%d", r, calls.Load())
	}
	expected := 1
	if stage == "commit" {
		expected = 5
	}
	if len(r.Commands) != expected {
		t.Fatalf("unexpected replay/release/command sequence: %q", r.Commands)
	}
	if stage == "commit" && r.Commands[4] != "COMMIT" {
		t.Fatalf("mutation command lost: %q", r.Commands)
	}
	t.Logf("reader_callbacks=%d primary=%v", calls.Load(), err)
}
func compositionNormal(t *testing.T, offered, rowValue bool) {
	t.Helper()
	p, e := newCompositionPeer(t, offered, "", "", rowValue)
	ctx := compositionContext(t)
	s, err := open(ctx, e)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := s.close(); err != nil {
			t.Error(err)
		}
	})
	if err = s.acquire(ctx, ioBound); err != nil {
		t.Fatal(err)
	}
	result, err := s.execute(ctx, mergeOperation, testCommit)
	if err != nil {
		t.Fatal(err)
	}
	if rowValue && (len(result) != 1 || len(result[0].rows) != 1 || !result[0].rows[0][0].Valid || result[0].rows[0][0].String != "\xfbordinary") {
		t.Fatalf("ordinary row misread: %#v", result)
	}
	if err = s.refresh(ctx); err != nil {
		t.Fatal(err)
	}
	if err = s.finish(ctx); err != nil {
		t.Fatal(err)
	}
	r := p.finish(t)
	want := []string{"SELECT @@autocommit", "SELECT GET_LOCK", "ROLLBACK", "SELECT DATABASE", "CALL DOLT_MERGE('" + testCommit + "')", "ROLLBACK", "SELECT DATABASE", "SELECT RELEASE_LOCK", "SELECT IS_USED_LOCK"}
	if len(r.Commands) != len(want) || !r.Quit || s.db.Stats().OpenConnections != 0 {
		t.Fatalf("normal lifecycle changed: %+v", r)
	}
	for i, prefix := range want {
		if !strings.HasPrefix(r.Commands[i], prefix) {
			t.Fatalf("command %d: %q", i, r.Commands[i])
		}
	}
}

type compositionCloseFault struct {
	net.Conn
	calls atomic.Int32
	cause error
}

func (c *compositionCloseFault) Close() error {
	c.calls.Add(1)
	return errors.Join(c.Conn.Close(), c.cause)
}
func compositionCleanup(t *testing.T) {
	t.Helper()
	p, e := newCompositionPeer(t, true, "commit", "Reader::cleanup-unregistered", false)
	ctx := compositionContext(t)
	// Separate observation seam: a real TCP connector/parser, not actual open.
	s := &session{target: e, name: lockName(e.base)}
	cause := errors.New("observed physical close failure")
	cfg, err := s.config(e)
	if err != nil {
		t.Fatal(err)
	}
	var observed *compositionCloseFault
	cfg.DialFunc = func(ctx context.Context, network, address string) (net.Conn, error) {
		var d net.Dialer
		c, err := d.DialContext(ctx, network, address)
		if err != nil {
			return nil, err
		}
		observed = &compositionCloseFault{Conn: c, cause: cause}
		s.transport = &transport{Conn: observed}
		return s.transport, nil
	}
	c, err := mysql.NewConnector(cfg)
	if err != nil {
		t.Fatal(err)
	}
	s.db = sql.OpenDB(&connector{Connector: c})
	s.db.SetMaxOpenConns(1)
	s.db.SetMaxIdleConns(0)
	t.Cleanup(func() {
		if err := s.close(); err != nil && !errors.Is(err, cause) {
			t.Error(err)
		}
	})
	s.conn, err = s.db.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.initialize(ctx); err != nil {
		t.Fatal(err)
	}
	if err = s.acquire(ctx, ioBound); err != nil {
		t.Fatal(err)
	}
	_, err = s.execute(ctx, commitOperation, "")
	assertCompositionFailure(t, s, err, cause)
	if observed == nil || observed.calls.Load() != 1 {
		t.Fatal("physical close was not exactly once")
	}
	r := p.finish(t)
	if r.PostRequestBytes != 0 || r.Terminal == "" || r.Quit {
		t.Fatalf("cleanup sent unexpected data: %+v", r)
	}
	t.Logf("observed_close_calls=%d joined_failure=%v", observed.calls.Load(), err)
}
