package graphsession

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
)

type infileObservation struct {
	requested bool
	bytes     []byte
	cause     error
}

type infileProbe struct {
	name    string
	started atomic.Bool
	done    chan infileObservation
}

func newInfileProbe(name string) *infileProbe {
	return &infileProbe{name: name, done: make(chan infileObservation, 1)}
}

// observe owns the post-request stream. The final cause is evidence: zero bytes
// plus a read timeout or a local cleanup close is not proof of remote disposal.
// A broken policy can upload only the small fixture payload; acknowledge its
// terminator once so the negative control cannot hang in the driver handler.
func (p *infileProbe) observe(c net.Conn, seq byte) (err error) {
	if !p.started.CompareAndSwap(false, true) {
		return errors.New("infile probe reused")
	}
	var observed infileObservation
	defer func() { observed.cause = err; p.done <- observed; close(p.done) }()
	if err = c.SetDeadline(time.Now().Add(2 * time.Second)); err != nil {
		return err
	}
	if err = packet(c, seq, append([]byte{0xfb}, p.name...)); err != nil {
		return err
	}
	observed.requested = true
	acknowledged := false
	for range 8 {
		var header [4]byte
		n, e := io.ReadFull(c, header[:])
		observed.bytes = append(observed.bytes, header[:n]...)
		if e != nil {
			return e
		}
		size := int(header[0]) | int(header[1])<<8 | int(header[2])<<16
		if size > 4096 {
			return fmt.Errorf("infile fixture packet exceeds bound: %d", size)
		}
		payload := make([]byte, size)
		n, e = io.ReadFull(c, payload)
		observed.bytes = append(observed.bytes, payload[:n]...)
		if e != nil {
			return e
		}
		if size == 0 && !acknowledged {
			acknowledged = true
			if e = packet(c, header[3]+1, []byte{0, 0, 0, 2, 0, 0, 0}); e != nil {
				return e
			}
		}
	}
	return errors.New("infile fixture packet count exceeded")
}

func assertInfileDisposal(t *testing.T, p *peer, probe *infileProbe) {
	t.Helper()
	var observation infileObservation
	select {
	case observation = <-probe.done:
	case <-time.After(3 * time.Second):
		t.Fatal("infile probe did not complete")
	}
	if !observation.requested {
		t.Errorf("LOCAL INFILE request was not sent: %v", observation.cause)
	}
	// TCP peers may report reset rather than EOF. A timeout, net.ErrClosed or
	// partial header is never accepted as the independent remote-close oracle.
	if !errors.Is(observation.cause, io.EOF) && !errors.Is(observation.cause, syscall.ECONNRESET) {
		t.Errorf("infile termination: %v; want remote EOF/reset", observation.cause)
	}
	if len(observation.bytes) != 0 {
		t.Errorf("bytes sent after LOCAL INFILE: %x", observation.bytes)
	}
	p.mu.Lock()
	connections := p.connections
	capabilities := append([]uint32(nil), p.capabilities...)
	p.mu.Unlock()
	if connections != 1 || len(capabilities) != 1 {
		t.Fatalf("connections=%d handshake records=%v", connections, capabilities)
	}
	const expected = uint32(512 | 8) // PROTOCOL_41 and CONNECT_WITH_DB, both offered.
	if capabilities[0]&expected != expected || capabilities[0]&128 != 0 {
		t.Errorf("parsed capabilities=%#x: require %#x and prohibit LOCAL FILES", capabilities[0], expected)
	}
}

func infileSource(t *testing.T, source string) (string, *atomic.Int32) {
	t.Helper()
	calls := new(atomic.Int32)
	const payload = "owned graphsession fixture payload"
	if source == "reader" {
		name := t.Name()
		mysql.RegisterReaderHandler(name, func() io.Reader { calls.Add(1); return strings.NewReader(payload) })
		t.Cleanup(func() { mysql.DeregisterReaderHandler(name) })
		return "Reader::" + name, calls
	}
	path := filepath.Join(t.TempDir(), "input.txt")
	if err := os.WriteFile(path, []byte(payload), 0600); err != nil {
		t.Fatal(err)
	}
	mysql.RegisterLocalFile(path)
	t.Cleanup(func() { mysql.DeregisterLocalFile(path) })
	return path, calls
}

func TestLocalInfileDispatchRefusal(t *testing.T) {
	for _, position := range []string{"first", "later"} {
		for _, source := range []string{"reader", "registered-file"} {
			t.Run(position+"/"+source, func(t *testing.T) {
				name, calls := infileSource(t, source)
				probe := newInfileProbe(name)
				s, p := openPeer(t, func(q string) reply {
					if q != "COMMIT" {
						return defaultReply(q)
					}
					request := reply{infile: probe}
					if position == "later" {
						return reply{rows: [][]string{{"first result"}}, nextSets: []reply{request}}
					}
					return request
				})
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				if err := s.acquire(ctx, ioBound); err != nil {
					t.Fatal(err)
				}
				before := p.commands()
				sets, err := s.execute(ctx, commitOperation, "")
				if !errors.Is(err, mysql.ErrLocalInfileDisabled) || !errors.Is(err, errUncertain) {
					t.Errorf("dispatch refusal=%v, sets=%v", err, sets)
				}
				if errors.Is(err, driver.ErrBadConn) || errors.Is(err, errReconnect) {
					t.Errorf("retryable/reconnected refusal: %v", err)
				}
				if s.state != discarded || s.db.Stats().OpenConnections != 0 {
					t.Errorf("state=%v pool=%+v; want discarded/empty", s.state, s.db.Stats())
				}
				if source == "reader" && calls.Load() != 0 {
					t.Errorf("registered reader invoked %d times", calls.Load())
				}
				if position == "first" && len(sets) != 0 {
					t.Errorf("evidence fabricated on first-header refusal: %+v", sets)
				}
				// Later partial evidence may accompany refusal; it must not be success.
				if position == "later" && (len(sets) != 1 || len(sets[0].rows) != 1 || len(sets[0].rows[0]) != 1 || sets[0].rows[0][0].String != "first result") {
					t.Errorf("first result not consumed before later refusal: %+v", sets)
				}
				assertInfileDisposal(t, p, probe)
				if _, e := s.execute(ctx, commitOperation, ""); !errors.Is(e, errState) {
					t.Errorf("reexecute=%v", e)
				}
				if e := s.refresh(ctx); !errors.Is(e, errState) {
					t.Errorf("refresh=%v", e)
				}
				if e := s.finish(ctx); !errors.Is(e, errState) {
					t.Errorf("finish=%v", e)
				}
				// The probe consumes all later bytes rather than adding them to commands.
				// This suffix oracle supplements, and cannot replace, zero-byte disposal.
				commands := p.commands()
				if len(commands) != len(before)+1 || commands[len(before)] != "COMMIT" {
					t.Errorf("before=%v after=%v; want one COMMIT and no recovery", before, commands)
				}
			})
		}
	}
}

func TestLocalInfileInitializationRefusal(t *testing.T) {
	name, calls := infileSource(t, "reader")
	probe := newInfileProbe(name)
	e, p := startPeer(t, func(q string) reply {
		if strings.HasPrefix(q, "SELECT @@autocommit") {
			return reply{infile: probe}
		}
		return defaultReply(q)
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	s, err := open(ctx, e)
	if s != nil {
		defer s.close()
		t.Error("initialization returned a session")
	}
	if !errors.Is(err, mysql.ErrLocalInfileDisabled) {
		t.Errorf("initialization refusal=%v", err)
	}
	if errors.Is(err, driver.ErrBadConn) || errors.Is(err, errReconnect) {
		t.Errorf("initialization replay: %v", err)
	}
	assertInfileDisposal(t, p, probe)
	if calls.Load() != 0 {
		t.Errorf("registered reader invoked %d times", calls.Load())
	}
	commands := p.commands()
	if len(commands) != 1 || !strings.HasPrefix(commands[0], "SELECT @@autocommit") {
		t.Errorf("initialization commands=%v; want only initial SELECT", commands)
	}
}
