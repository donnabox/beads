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
			t.Run("B3 note", func(t *testing.T) {
				id := f.create("note me")
				f.mustRun(nil, "note", id, "a note")
				if row := c.show(t, id); !strings.Contains(row["notes"].(string), "a note") {
					t.Fatalf("after bd note: %v", row)
				}
			})
			t.Run("B4 q and q --parent", func(t *testing.T) {
				id := strings.TrimSpace(f.mustRun(nil, "q", "quick capture", "-p", "1", "-l", "x").stdout)
				if row := c.show(t, id); row["title"] != "quick capture" || row["priority"] != float64(1) {
					t.Fatalf("after bd q: %v", row)
				}
				parent := strings.TrimSpace(f.mustRun(nil, "create", "q parent", "-t", "epic", "-l", "inherited", "--silent").stdout)
				child := strings.TrimSpace(f.mustRun(nil, "q", "quick child", "--parent", parent).stdout)
				if !strings.HasPrefix(child, parent+".") {
					t.Errorf("bd q --parent minted %s under %s", child, parent)
				}
				if labels, _ := c.show(t, child)["labels"].([]any); !slices.Contains(labels, any("inherited")) {
					t.Errorf("bd q --parent did not inherit labels: %v", labels)
				}
			})
			t.Run("B5 todo add", func(t *testing.T) {
				f.mustRun(nil, "todo", "add", "a todo", "-d", "todo body")
				if out := f.mustRun(nil, "todo", "--json").stdout; !strings.Contains(out, "a todo") {
					t.Errorf("bd todo: %s", out)
				}
			})
			t.Run("B6 duplicate and supersede", func(t *testing.T) {
				a, b := f.create("duplicate of"), f.create("canonical")
				f.mustRun(nil, "duplicate", a, "--of", b)
				if c.show(t, a)["status"] != "closed" {
					t.Errorf("bd duplicate did not close %s", a)
				}
				old, replacement := f.create("superseded"), f.create("replacement")
				f.mustRun(nil, "supersede", old, "--with", replacement)
				if c.show(t, old)["status"] != "closed" {
					t.Errorf("bd supersede did not close %s", old)
				}
			})
			t.Run("B7 dep relate and unrelate", func(t *testing.T) {
				e, g := f.create("relate one"), f.create("relate two")
				f.mustRun(nil, "dep", "relate", e, g)
				if out := f.mustRun(nil, "dep", "list", e, "--json").stdout; !depListHas(t, out, g, "relates-to") {
					t.Errorf("bd dep relate left no edge: %s", out)
				}
				if out := f.mustRun(nil, "dep", "list", g, "--json").stdout; !depListHas(t, out, e, "relates-to") {
					t.Errorf("bd dep relate left no edge back: %s", out)
				}
				f.mustRun(nil, "dep", "unrelate", e, g)
				if out := f.mustRun(nil, "dep", "list", e, "--json").stdout; strings.Contains(out, g) {
					t.Errorf("bd dep unrelate left the edge: %s", out)
				}
			})
			t.Run("B8 defer and undefer write", func(t *testing.T) {
				id := f.create("defer me")
				c.counter.reset()
				f.mustRun(nil, "defer", id, "--until", "2099-01-01T00:00:00Z", "--reason", "later")
				c.dialed(t, "issue/update", "PATCH /v0/beads/issues/"+id)
				row := c.show(t, id)
				if row["status"] != "deferred" || !strings.Contains(row["notes"].(string), "later") || row["defer_until"] == nil {
					t.Fatalf("after bd defer: %v", row)
				}
				f.mustRun(nil, "undefer", id)
				if row := c.show(t, id); row["status"] != "open" || row["defer_until"] != nil {
					t.Fatalf("after bd undefer: %v", row)
				}
			})
			t.Run("B9 unclaim dials the release", func(t *testing.T) {
				id := f.create("unclaim me")
				f.mustRun(nil, "update", id, "--claim")
				c.counter.reset()
				f.mustRun(nil, "unclaim", id, "--reason", "changed my mind")
				c.dialed(t, "issue/unclaim", ":release")
				row := c.show(t, id)
				if row["status"] != "open" || row["assignee"] != nil {
					t.Fatalf("after bd unclaim: %v", row)
				}
				if comments := f.mustRun(nil, "comments", id).stdout; !strings.Contains(comments, "changed my mind") {
					t.Errorf("the reason did not land as a comment: %s", comments)
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
