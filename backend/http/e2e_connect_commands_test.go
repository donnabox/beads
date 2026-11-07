//go:build cgo

package bdhttp_test

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/httpapi"
)

// TestE2E_ConnectCommandParity runs, through the real bd binary against an
// in-process bd serve, the commands of group B of the native fixes: each used to
// reach the refusing shell, and some printed the refusal, wrote nothing and exited
// 0. A yes applies to both protocols, so it runs on both.
func TestE2E_ConnectCommandParity(t *testing.T) {
	skipUnlessEmbeddedDolt(t)
	bin := buildBD(t)

	for _, tc := range []struct {
		protocol string
		server   httpapi.Protocol
		connect  []string
	}{
		{protocol: "v0", server: httpapi.ProtocolV0},
		{protocol: "bdp", server: httpapi.ProtocolBoth, connect: []string{"--protocol", "bdp"}},
	} {
		t.Run(tc.protocol, func(t *testing.T) {
			proxyAddr, counter := startCountingProxy(t, startProtocolServer(t, tc.server))
			f := &remoteRolesFixture{t: t, bin: bin, dir: t.TempDir(), ids: map[string]string{}}
			f.mustRun(nil, append([]string{"connect", "http://" + proxyAddr, "--expect-project-id", e2eProjectID, "--json"}, tc.connect...)...)
			c := &connectCommands{f: f, counter: counter}
			t.Run("B1 priority", func(t *testing.T) {
				id := f.create("priority me")
				f.mustRun(nil, "priority", id, "1")
				if row := c.show(t, id); row["priority"] != float64(1) {
					t.Fatalf("after bd priority: %v", row)
				}
			})
			t.Run("B2 tag", func(t *testing.T) {
				id := f.create("tag me")
				f.mustRun(nil, "tag", id, "t1")
				if labels, _ := c.show(t, id)["labels"].([]any); !slices.Contains(labels, any("t1")) {
					t.Fatalf("after bd tag: %v", labels)
				}
			})
			//ROWS
		})
	}
}

// connectCommands is what each row's subtest asks of the connection.
type connectCommands struct {
	f       *remoteRolesFixture
	counter *requestCounter
}

// dialed names a request both ways, by its v0 path and by its Callable, because
// which of the two the client speaks is the wire's business.
func (c *connectCommands) dialed(t *testing.T, callable, v0 string) {
	t.Helper()
	for _, request := range c.counter.snapshot() {
		if strings.Contains(request, "/bdp/call/"+callable) || strings.Contains(request, v0) {
			return
		}
	}
	t.Errorf("no request reached %s (%s): %v", callable, v0, c.counter.snapshot())
}

func (c *connectCommands) show(t *testing.T, id string) map[string]any {
	t.Helper()
	out := c.f.mustRun(nil, "show", id, "--json").stdout
	var rows []map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &rows); err != nil {
		t.Fatalf("parse bd show: %v\n%s", err, out)
	}
	if len(rows) != 1 {
		t.Fatalf("bd show %s answered %d rows: %s", id, len(rows), out)
	}
	return rows[0]
}

func (c *connectCommands) refused(t *testing.T, want string, args ...string) {
	t.Helper()
	r := c.f.run(nil, args...)
	if r.code == 0 {
		t.Fatalf("bd %v exited 0: stdout=%s stderr=%s", args, r.stdout, r.stderr)
	}
	if !strings.Contains(r.stdout+r.stderr, want) {
		t.Errorf("bd %v: want %q in stdout=%s stderr=%s", args, want, r.stdout, r.stderr)
	}
}
