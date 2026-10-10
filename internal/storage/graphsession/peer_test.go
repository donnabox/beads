package graphsession

import (
	"context"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

// A protocol fixture, not a database: predefined replies test ownership of real
// mysql driver I/O. It implements no locks, transactions or engine semantics.
type reply struct {
	tailError bool
	multi     bool
	rows      [][]string
	err       bool
	drop      bool
	null      bool
}
type peer struct {
	mu      sync.Mutex
	queries []string
	fn      func(string) reply
}

func defaultReply(q string) reply {
	switch {
	case strings.HasPrefix(q, "SELECT @@autocommit"):
		return reply{rows: [][]string{{"1", "0", "0", "41"}}}
	case strings.HasPrefix(q, "SELECT GET_LOCK"):
		return reply{rows: [][]string{{"1", "41"}}}
	case strings.HasPrefix(q, "SELECT DATABASE"):
		return reply{rows: [][]string{{"Sales/main", "main", "41", "41"}}}
	case strings.HasPrefix(q, "SELECT RELEASE_LOCK"):
		return reply{rows: [][]string{{"1"}}}
	case strings.HasPrefix(q, "SELECT IS_USED_LOCK"):
		return reply{rows: [][]string{{"42"}}}
	case strings.HasPrefix(q, "CALL DOLT_MERGE"):
		return reply{rows: [][]string{{"commit", "0", "0", "completed"}}}
	default:
		return reply{}
	}
}
func (p *peer) commands() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]string(nil), p.queries...)
}
func packet(c net.Conn, seq byte, b []byte) error {
	h := []byte{byte(len(b)), byte(len(b) >> 8), byte(len(b) >> 16), seq}
	if _, err := c.Write(h); err != nil {
		return err
	}
	_, err := c.Write(b)
	return err
}
func readPacket(c net.Conn) ([]byte, error) {
	var h [4]byte
	if _, err := io.ReadFull(c, h[:]); err != nil {
		return nil, err
	}
	n := int(h[0]) | int(h[1])<<8 | int(h[2])<<16
	if n > 1<<20 {
		return nil, errors.New("fixture packet too large")
	}
	b := make([]byte, n)
	_, err := io.ReadFull(c, b)
	return b, err
}
func handshake(c net.Conn) error {
	caps := uint32(1 | 4 | 8 | 512 | 8192 | 32768 | 131072 | 524288)
	b := append([]byte{10}, []byte("8.0.0-fixture\x00")...)
	b = binary.LittleEndian.AppendUint32(b, 41)
	b = append(b, []byte("12345678\x00")...)
	b = binary.LittleEndian.AppendUint16(b, uint16(caps))
	b = append(b, 33)
	b = binary.LittleEndian.AppendUint16(b, 2)
	b = binary.LittleEndian.AppendUint16(b, uint16(caps>>16))
	b = append(b, 21)
	b = append(b, make([]byte, 10)...)
	b = append(b, []byte("abcdefghijkl\x00mysql_native_password\x00")...)
	if err := packet(c, 0, b); err != nil {
		return err
	}
	if _, err := readPacket(c); err != nil {
		return err
	}
	return packet(c, 2, []byte{0, 0, 0, 2, 0, 0, 0})
}
func lenString(b []byte, s string) []byte {
	if len(s) < 251 {
		b = append(b, byte(len(s)))
	} else {
		b = append(b, 0xfc, byte(len(s)), byte(len(s)>>8))
	}
	return append(b, []byte(s)...)
}
func sendReply(c net.Conn, r reply) error {
	if r.drop {
		return c.Close()
	}
	if r.err {
		return packet(c, 1, []byte{0xff, 0x15, 0x04, '#', 'H', 'Y', '0', '0', '0', 'f', 'i', 'x', 't', 'u', 'r', 'e'})
	}
	if len(r.rows) == 0 {
		return packet(c, 1, []byte{0, 0, 0, 2, 0, 0, 0})
	}
	if err := packet(c, 1, []byte{byte(len(r.rows[0]))}); err != nil {
		return err
	}
	seq := byte(2)
	for range r.rows[0] {
		var b []byte
		for _, s := range []string{"def", "", "", "", "value", ""} {
			b = lenString(b, s)
		}
		b = append(b, 12, 33, 0, 255, 0, 0, 0, 0xfd, 0, 0, 0, 0, 0)
		if err := packet(c, seq, b); err != nil {
			return err
		}
		seq++
	}
	eof := []byte{0xfe, 0, 0, 2, 0}
	if err := packet(c, seq, eof); err != nil {
		return err
	}
	seq++
	for _, row := range r.rows {
		var b []byte
		for _, v := range row {
			if r.null {
				b = append(b, 0xfb)
			} else {
				b = lenString(b, v)
			}
		}
		if err := packet(c, seq, b); err != nil {
			return err
		}
		seq++
	}
	if r.tailError {
		return packet(c, seq, []byte{0xff, 0x15, 0x04, '#', 'H', 'Y', '0', '0', '0', 'd', 'r', 'a', 'i', 'n'})
	}
	if r.multi {
		if err := packet(c, seq, []byte{0xfe, 0, 0, 10, 0}); err != nil {
			return err
		}
		seq++
		if err := packet(c, seq, []byte{1}); err != nil {
			return err
		}
		seq++
		var column []byte
		for _, name := range []string{"def", "", "", "", "second", ""} {
			column = lenString(column, name)
		}
		column = append(column, 12, 33, 0, 255, 0, 0, 0, 0xfd, 0, 0, 0, 0, 0)
		if err := packet(c, seq, column); err != nil {
			return err
		}
		seq++
		if err := packet(c, seq, eof); err != nil {
			return err
		}
		seq++
		if err := packet(c, seq, lenString(nil, "second")); err != nil {
			return err
		}
		seq++
		return packet(c, seq, eof)
	}
	return packet(c, seq, eof)
}
func (p *peer) serve(c net.Conn, stopAfterCommit <-chan struct{}) error {
	if err := handshake(c); err != nil {
		return err
	}
	for {
		b, err := readPacket(c)
		if err != nil {
			return err
		}
		if len(b) == 0 {
			return io.ErrUnexpectedEOF
		}
		if b[0] == 1 {
			return nil
		}
		if b[0] != 3 {
			return errors.New("fixture unexpected command")
		}
		q := string(b[1:])
		p.mu.Lock()
		p.queries = append(p.queries, q)
		p.mu.Unlock()
		r := defaultReply(q)
		if p.fn != nil {
			r = p.fn(q)
		}
		if err := sendReply(c, r); err != nil {
			return err
		}
		if q == "COMMIT" && stopAfterCommit != nil {
			<-stopAfterCommit
			return nil
		}
	}
}
func startPeer(t *testing.T, fn func(string) reply) (endpoint, *peer) {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	p := &peer{fn: fn}
	var wg sync.WaitGroup
	var mu sync.Mutex
	var conns []net.Conn
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			c, e := l.Accept()
			if e != nil {
				return
			}
			mu.Lock()
			conns = append(conns, c)
			mu.Unlock()
			wg.Add(1)
			go func() { defer wg.Done(); defer c.Close(); _ = p.serve(c, nil) }()
		}
	}()
	t.Cleanup(func() {
		_ = l.Close()
		<-done
		mu.Lock()
		for _, c := range conns {
			_ = c.Close()
		}
		mu.Unlock()
		joined := make(chan struct{})
		go func() { wg.Wait(); close(joined) }()
		select {
		case <-joined:
		case <-time.After(3 * time.Second):
			t.Error("fixture workers not joined")
		}
	})
	return endpoint{address: l.Addr().String(), base: "sales", branch: "main", user: "fixture"}, p
}
func openPeer(t *testing.T, fn func(string) reply) (*session, *peer) {
	t.Helper()
	e, p := startPeer(t, fn)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	s, err := open(ctx, e)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.close() })
	return s, p
}
