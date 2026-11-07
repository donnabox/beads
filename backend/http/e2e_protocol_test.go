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
	// As bd serve's own does, so the client's external-dependency decorator stands
	// down and what goes out is the command's own reads.
	cfg.ExternalDependencyPolicy = true
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

	// A connect states the protocol afresh, so one with no flag is v0 again.
	t.Run("a reconnect with no flag goes back to v0", func(t *testing.T) {
		addr := startProtocolServer(t, httpapi.ProtocolBoth)
		f := &remoteRolesFixture{t: t, bin: bin, dir: t.TempDir(), ids: map[string]string{}}
		f.mustRun(nil, "connect", "http://"+addr, "--protocol", "bdp")
		sidecar := filepath.Join(f.dir, ".beads", "http_target.json")
		raw, err := os.ReadFile(sidecar)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(raw), `"protocol": "bdp"`) {
			t.Fatalf("not recorded: %s", raw)
		}
		f.mustRun(nil, "connect", "http://"+addr)
		if raw, err = os.ReadFile(sidecar); err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(raw), `"protocol"`) {
			t.Errorf("still names a protocol: %s", raw)
		}
	})

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

// A command answers the same on either protocol, refusals included: BDP changes what
// goes on the wire and nothing a command does. Each pair runs through the real bd,
// one server of each kind, on the same seed, and the request count is what tells a
// command that reads more on one protocol.
func TestE2E_ACommandAnswersTheSameOnBothProtocols(t *testing.T) {
	skipUnlessEmbeddedDolt(t)
	bin := buildBD(t)

	type side struct {
		f       *remoteRolesFixture
		counter *requestCounter
		blocker string
		blocked string
	}
	open := func(t *testing.T, server httpapi.Protocol, connect ...string) *side {
		t.Helper()
		proxy, counter := startCountingProxy(t, startProtocolServer(t, server))
		f := &remoteRolesFixture{t: t, bin: bin, dir: t.TempDir(), ids: map[string]string{}}
		f.mustRun(nil, append([]string{"connect", "http://" + proxy}, connect...)...)
		s := &side{f: f, counter: counter, blocker: f.create("blocker"), blocked: f.create("blocked")}
		f.mustRun(nil, "dep", "add", s.blocked, s.blocker)
		return s
	}

	for _, tc := range []struct {
		name string
		args []string
		// refuses, when set, is what both protocols' refusal has to name.
		refuses string
	}{
		{name: "list", args: []string{"list"}},
		{name: "ready", args: []string{"ready"}},
		{name: "graph", args: []string{"graph", "$blocker", "--json"}},
		{name: "info", args: []string{"info", "--json"}},
		{name: "update --claim", args: []string{"update", "$blocked", "--claim", "--json"}},
		{name: "ready --explain", args: []string{"ready", "--explain", "--json"}},
		{name: "show --refs", args: []string{"show", "$blocker", "--refs", "--json"}},
		{name: "stale refuses", args: []string{"stale"}, refuses: "GetStaleIssues"},
		{name: "kv get", args: []string{"kv", "get", "nothing"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v0 := open(t, httpapi.ProtocolV0)
			bdp := open(t, httpapi.ProtocolBoth, "--protocol", "bdp")
			run := func(s *side) (bdResult, int) {
				args := make([]string, len(tc.args))
				for i, arg := range tc.args {
					args[i] = strings.NewReplacer("$blocker", s.blocker, "$blocked", s.blocked).Replace(arg)
				}
				s.counter.reset()
				r := s.f.run(nil, args...)
				return r, len(s.counter.uriSnapshot())
			}
			onV0, sentV0 := run(v0)
			onBDP, sentBDP := run(bdp)
			if onV0.code != onBDP.code {
				t.Errorf("exit %d on v0, %d on BDP:\nv0: %s\nbdp: %s", onV0.code, onBDP.code, onV0.stderr, onBDP.stderr)
			}
			if sentV0 != sentBDP {
				t.Errorf("%d requests on v0, %d on BDP", sentV0, sentBDP)
			}
			if tc.refuses != "" {
				for name, r := range map[string]bdResult{"v0": onV0, "bdp": onBDP} {
					if r.code == 0 || !strings.Contains(r.stdout+r.stderr, tc.refuses) {
						t.Errorf("on %s, exit %d, want a refusal naming %q: %s %s", name, r.code, tc.refuses, r.stdout, r.stderr)
					}
				}
			}
		})
	}

	// `bd update --claim` is an update on both protocols, not a claimIssue.
	t.Run("update --claim is issue/update on BDP", func(t *testing.T) {
		s := open(t, httpapi.ProtocolBoth, "--protocol", "bdp")
		s.counter.reset()
		s.f.mustRun(nil, "update", s.blocked, "--claim", "--json")
		sent := strings.Join(s.counter.uriSnapshot(), "\n")
		if !strings.Contains(sent, "issue/update") || strings.Contains(sent, "issue/claim") {
			t.Errorf("sent:\n%s", sent)
		}
	})

	// `bd comments add --author X` attributes the comment to X on both: on BDP the
	// wire layer moves the author to revisionProvenance.
	t.Run("--author names the comment's author on both", func(t *testing.T) {
		for _, s := range []*side{open(t, httpapi.ProtocolV0), open(t, httpapi.ProtocolBoth, "--protocol", "bdp")} {
			s.f.mustRun(nil, "comments", "add", s.blocked, "hello", "--author", "kim")
			listed := s.f.mustRun(nil, "comments", s.blocked, "--json")
			if !strings.Contains(listed.stdout, `"author": "kim"`) {
				t.Errorf("the comment is not kim's: %s", listed.stdout)
			}
		}
	})

	// C1: bd undefer's per-id failure is the command's failure, on both.
	t.Run("undefer of an issue that is not deferred exits nonzero on both", func(t *testing.T) {
		for _, s := range []*side{open(t, httpapi.ProtocolV0), open(t, httpapi.ProtocolBoth, "--protocol", "bdp")} {
			r := s.f.run(nil, "undefer", s.blocked)
			if r.code == 0 {
				t.Errorf("exit 0: %s %s", r.stdout, r.stderr)
			}
			if !strings.Contains(r.stdout+r.stderr, "is not deferred") {
				t.Errorf("did not say why: %s %s", r.stdout, r.stderr)
			}
		}
	})

	// C2: --author is an alias of --actor, so two different values refuse and write
	// nothing, on both.
	t.Run("--actor and --author with two names refuse on both", func(t *testing.T) {
		for _, s := range []*side{open(t, httpapi.ProtocolV0), open(t, httpapi.ProtocolBoth, "--protocol", "bdp")} {
			r := s.f.run(nil, "--actor", "sam", "comments", "add", s.blocked, "must not land", "--author", "lee")
			if r.code == 0 || !strings.Contains(r.stdout+r.stderr, "alias of --actor") {
				t.Errorf("exit %d: %s %s", r.code, r.stdout, r.stderr)
			}
			if listed := s.f.mustRun(nil, "comments", s.blocked, "--json"); strings.Contains(listed.stdout, "must not land") {
				t.Errorf("the refused comment was written: %s", listed.stdout)
			}
		}
	})
}
