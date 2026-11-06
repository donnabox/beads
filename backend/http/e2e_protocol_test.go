//go:build cgo

package bdhttp_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/httpapi"
	"github.com/steveyegge/beads/internal/storage/embeddeddolt"
)

// startProtocolServer is startE2EServer on one protocol.
func startProtocolServer(t *testing.T, protocol httpapi.Protocol) string {
	t.Helper()
	ctx := context.Background()
	reference, err := embeddeddolt.Open(ctx, t.TempDir(), e2eDatabase, "main")
	if err != nil {
		t.Fatalf("open the reference store: %v", err)
	}
	t.Cleanup(func() { _ = reference.Close() })
	if err := reference.SetConfig(ctx, "issue_prefix", "e2e"); err != nil {
		t.Fatalf("set the issue prefix: %v", err)
	}
	cfg := e2eServeConfig(t, reference, e2eProjectID, e2eDatabase)
	cfg.Protocol = protocol
	srv, err := httpapi.Listen(cfg)
	if err != nil {
		t.Fatalf("bind the in-process server: %v", err)
	}
	serveCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() { defer close(done); _ = srv.Serve(serveCtx) }()
	t.Cleanup(func() {
		cancel()
		select {
		case <-done:
		case <-time.After(30 * time.Second):
			t.Error("the in-process server did not shut down")
		}
	})
	return srv.Addr()
}

// Every pairing of a server's protocol and a connection's, through the real bd: the
// pairs that share one work end to end and on the wire they asked for, the handshake
// included, and the two that do not refuse at connect, nonzero, in words that are true
// of them, and leave nothing behind.
func TestE2E_ProtocolPairings(t *testing.T) {
	skipUnlessEmbeddedDolt(t)
	bin := buildBD(t)

	onV0 := func(request string) bool {
		return strings.Contains(request, " /v0/") || strings.Contains(request, " /healthz")
	}
	onBDP := func(request string) bool { return strings.Contains(request, " /bdp/call/") }

	for _, tc := range []struct {
		name    string
		server  httpapi.Protocol
		connect []string
		sidecar string
		wire    func(string) bool
		refuses []string
	}{
		{name: "default to default", connect: nil, wire: onV0},
		{name: "v0 to v0", server: httpapi.ProtocolV0, connect: []string{"--protocol", "v0"}, wire: onV0},
		{name: "v0 to both", server: httpapi.ProtocolBoth, connect: nil, wire: onV0},
		{name: "bdp to both", server: httpapi.ProtocolBoth, connect: []string{"--protocol", "bdp"}, sidecar: "bdp", wire: onBDP},
		{name: "bdp to bdp", server: httpapi.ProtocolBDP, connect: []string{"--protocol", "bdp"}, sidecar: "bdp", wire: onBDP},
		{
			name: "bdp to a v0-only server", server: httpapi.ProtocolV0, connect: []string{"--protocol", "bdp"},
			refuses: []string{"has no /bdp/call", "serving v0 only", "from a release without BDP", "bd serve --protocol both", "bd connect --protocol v0"},
		},
		{
			name: "v0 to a bdp-only server", server: httpapi.ProtocolBDP, connect: nil,
			refuses: []string{"protocol_not_served", "runs with --protocol bdp", "bd connect --protocol bdp", "bd serve --protocol both"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			proxy, counter := startCountingProxy(t, startProtocolServer(t, tc.server))
			f := &remoteRolesFixture{t: t, bin: bin, dir: t.TempDir(), ids: map[string]string{}}
			connected := f.run(nil, append([]string{"connect", "http://" + proxy}, tc.connect...)...)

			sidecar := filepath.Join(f.dir, ".beads", "http_target.json")
			if tc.refuses != nil {
				if connected.code == 0 {
					t.Fatalf("bd connect exited 0: %s", connected.stdout)
				}
				for _, want := range tc.refuses {
					if !strings.Contains(connected.stderr, want) {
						t.Errorf("missing %q in: %s", want, connected.stderr)
					}
				}
				if strings.Contains(connected.stderr, "predates") {
					t.Errorf("says the server predates this: %s", connected.stderr)
				}
				if _, err := os.Stat(sidecar); !os.IsNotExist(err) {
					t.Errorf("a refused connect left a sidecar: %v", err)
				}
				return
			}
			if connected.code != 0 {
				t.Fatalf("bd connect (exit %d): %s", connected.code, connected.stderr)
			}
			raw, err := os.ReadFile(sidecar)
			if err != nil {
				t.Fatal(err)
			}
			if tc.sidecar == "" && strings.Contains(string(raw), `"protocol"`) {
				t.Errorf("a v0 connection's sidecar names a protocol: %s", raw)
			}
			if tc.sidecar != "" && !strings.Contains(string(raw), `"protocol": "`+tc.sidecar+`"`) {
				t.Errorf("the sidecar does not record %s: %s", tc.sidecar, raw)
			}

			id := f.create("pairing")
			f.mustRun(nil, "update", id, "--priority", "1")
			f.mustRun(nil, "show", id, "--json")
			f.mustRun(nil, "list", "--json")
			f.mustRun(nil, "close", id, "--reason", "done")
			for _, request := range counter.snapshot() {
				if !tc.wire(request) {
					t.Errorf("sent %s", request)
				}
			}
		})
	}

	// A saved connection outliving its server's protocol. A read of an issue that
	// exists must not come back as "no such issue".
	t.Run("a v0 connection whose server moved to bdp", func(t *testing.T) {
		v0Addr := startProtocolServer(t, httpapi.ProtocolV0)
		f := &remoteRolesFixture{t: t, bin: bin, dir: t.TempDir(), ids: map[string]string{}}
		f.mustRun(nil, "connect", "http://"+v0Addr)
		sidecar := filepath.Join(f.dir, ".beads", "http_target.json")
		raw, err := os.ReadFile(sidecar)
		if err != nil {
			t.Fatal(err)
		}
		bdpAddr := startProtocolServer(t, httpapi.ProtocolBDP)
		moved := strings.ReplaceAll(string(raw), v0Addr, bdpAddr)
		if err := os.WriteFile(sidecar, []byte(moved), 0o600); err != nil {
			t.Fatal(err)
		}
		r := f.run(nil, "show", "e2e-anything")
		if r.code == 0 {
			t.Fatalf("bd show exited 0: %s", r.stdout)
		}
		if !strings.Contains(r.stdout+r.stderr, "protocol_not_served") {
			t.Errorf("bd show did not say why: stdout=%s stderr=%s", r.stdout, r.stderr)
		}
	})
}
