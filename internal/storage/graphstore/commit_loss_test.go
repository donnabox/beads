//go:build cgo

package graphstore

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestServerLostCommitResponse(t *testing.T) {
	portText := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
	if portText == "" {
		t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for ordinary released-server qualification")
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	o := testOptions(t)
	token, err := freshToken()
	if err != nil {
		t.Fatal(err)
	}
	o.Backend, o.Database = "server", "graph_loss_"+token
	o.ServerHost, o.ServerPort, o.ServerUser = "127.0.0.1", port, "root"
	if err := Init(ctx, o); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanupCtx, stop := context.WithTimeout(context.Background(), 10*time.Second)
		defer stop()
		admin, err := openBackend(cleanupCtx, o, "")
		if err != nil {
			t.Error(err)
			return
		}
		if _, err = admin.db.ExecContext(cleanupCtx, "DROP DATABASE `"+o.Database+"`"); err != nil {
			t.Error(err)
		}
		if err = admin.Close(); err != nil {
			t.Error(err)
		}
	})
	proxyPort, observed := startCommitLossProxy(t, net.JoinHostPort(o.ServerHost, portText))
	proxied := o
	proxied.ServerPort = proxyPort
	s, err := OpenExisting(ctx, proxied)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := s.Close(); err != nil {
			t.Error(err)
		}
	})
	_, err = s.Create(ctx, CreateRequest{Path: "beads/uncertain", Title: "Accepted, response lost", Body: "Still complete", Actor: "test-author"})
	if !errors.Is(err, ErrOutcomeUnknown) || errors.Is(err, ErrConflict) {
		t.Fatalf("lost COMMIT response must be unknown, got %v", err)
	}
	select {
	case packet := <-observed:
		// This witness belongs to the fault harness, not the client. An OK
		// response proves this particular unknown outcome actually committed.
		if len(packet) == 0 || packet[0] != 0x00 {
			t.Fatalf("server did not acknowledge COMMIT with OK (payload length %d)", len(packet))
		}
	case <-ctx.Done():
		t.Fatal("proxy did not observe COMMIT response", ctx.Err())
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	direct, err := OpenExisting(ctx, o)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := direct.Close(); err != nil {
			t.Error(err)
		}
	})
	r, err := direct.Show(ctx, "beads/uncertain")
	if err != nil {
		t.Fatal(err)
	}
	if r.Properties.Title != "Accepted, response lost" || r.Properties.Body != "Still complete" || r.Attribution.Actor != "test-author" {
		t.Fatalf("wrong recovered record: %+v", r)
	}
	for _, table := range []string{"graph_preview_catalog", "graph_preview_payloads", "graph_preview_versions"} {
		var count int
		if err := direct.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table+" WHERE path='beads/uncertain'").Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Fatalf("unknown outcome has %d rows in %s", count, table)
		}
	}
	if _, err := direct.Create(ctx, CreateRequest{Path: "beads/successor"}); err != nil {
		t.Fatal(err)
	}
	if _, err := direct.Show(ctx, "beads/uncertain"); err != nil {
		t.Fatal(err)
	}
}

// startCommitLossProxy is a single-client, plaintext, loopback fault control.
// Authentication and SQL packets are forwarded without inspection/logging,
// except exact COM_QUERY COMMIT. That command reaches the real server, whose
// response is consumed and withheld from the client. No mock engine is used.
func startCommitLossProxy(t *testing.T, target string) (int, <-chan []byte) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	var mu sync.Mutex
	var front, back net.Conn
	done := make(chan struct{})
	observed := make(chan []byte, 1)
	closeConnections := func() {
		mu.Lock()
		defer mu.Unlock()
		if front != nil {
			_ = front.Close()
		}
		if back != nil {
			_ = back.Close()
		}
	}
	t.Cleanup(func() {
		_ = listener.Close()
		closeConnections()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Error("test COMMIT proxy did not stop")
		}
	})
	go func() {
		defer close(done)
		defer closeConnections()
		client, err := listener.Accept()
		if err != nil {
			return
		}
		mu.Lock()
		front = client
		mu.Unlock()
		server, err := net.DialTimeout("tcp", target, 5*time.Second)
		if err != nil {
			return
		}
		mu.Lock()
		back = server
		mu.Unlock()
		_ = client.SetDeadline(time.Now().Add(45 * time.Second))
		_ = server.SetDeadline(time.Now().Add(45 * time.Second))
		var commitSent atomic.Bool
		uploadDone := make(chan struct{})
		go func() {
			defer close(uploadDone)
			defer closeConnections()
			for {
				packet, err := readTestMySQLPacket(client)
				if err != nil {
					return
				}
				payload := packet[4:]
				if len(payload) > 1 && payload[0] == 0x03 && strings.EqualFold(strings.TrimSpace(string(payload[1:])), "COMMIT") {
					commitSent.Store(true)
				}
				if _, err := server.Write(packet); err != nil {
					return
				}
			}
		}()
		for {
			packet, err := readTestMySQLPacket(server)
			if err != nil {
				break
			}
			if commitSent.Load() {
				observed <- append([]byte(nil), packet[4:]...)
				break
			}
			if _, err := client.Write(packet); err != nil {
				break
			}
		}
		closeConnections()
		<-uploadDone
	}()
	return listener.Addr().(*net.TCPAddr).Port, observed
}

func readTestMySQLPacket(r io.Reader) ([]byte, error) {
	var header [4]byte
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return nil, err
	}
	n := int(binary.LittleEndian.Uint32(header[:]) & 0x00ffffff)
	if n > 1<<20 {
		return nil, fmt.Errorf("test proxy packet exceeds 1 MiB")
	}
	packet := make([]byte, 4+n)
	copy(packet, header[:])
	_, err := io.ReadFull(r, packet[4:])
	return packet, err
}
