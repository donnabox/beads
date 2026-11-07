//go:build cgo

package bdhttp_test

import (
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/httpapi"
	"github.com/steveyegge/beads/internal/ui"
)

// TestE2E_ConnectGapsAreClosed runs, through the real bd on both protocols, the
// commands of group D of the native fixes. Each answered at the base with exit 0
// and a gap: a marker that never rendered, a graph of only the root, blank blockers,
// shallow rows, a "not set" for a key that exists. A yes applies to both protocols,
// so v0 stops matching the base for these on purpose.
func TestE2E_ConnectGapsAreClosed(t *testing.T) {
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
			dialed := func(t *testing.T, callable, v0 string) {
				t.Helper()
				for _, request := range counter.snapshot() {
					if strings.Contains(request, "/bdp/call/"+callable) || strings.Contains(request, v0) {
						return
					}
				}
				t.Errorf("no request reached %s (%s): %v", callable, v0, counter.snapshot())
			}
			_ = dialed

			t.Run("D1 list text renders the gate marker", func(t *testing.T) {
				subject := f.create("gated subject")
				f.mustRun(nil, "gate", "create", "--type=timer", "--timeout=1h", "--blocks", subject, "--json")
				if out := f.mustRun(nil, "list").stdout; !strings.Contains(out, ui.StatusIconGated+" "+subject) {
					t.Errorf("bd list does not mark %s gated: %s", subject, out)
				}
			})

			t.Run("D2 ready text shows the parent epic", func(t *testing.T) {
				epic := strings.TrimSpace(f.mustRun(nil, "create", "annotated epic", "-t", "epic", "--silent").stdout)
				f.mustRun(nil, "create", "annotated child", "--parent", epic, "--silent")
				if text := f.mustRun(nil, "ready").stdout; !strings.Contains(text, "annotated epic") {
					t.Errorf("bd ready lost the parent epic: %s", text)
				}
			})

			t.Run("D5 info carries the issue count", func(t *testing.T) {
				out := f.mustRun(nil, "info", "--json").stdout
				if !strings.Contains(out, `"issue_count"`) {
					t.Errorf("bd info --json has no issue_count: %s", out)
				}
			})

			t.Run("D3 graph reports dependents", func(t *testing.T) {
				parent := f.create("graph root")
				child := f.create("graph dependent")
				f.mustRun(nil, "dep", "add", child, parent)
				out := f.mustRun(nil, "graph", parent, "--json").stdout
				if !strings.Contains(out, child) {
					t.Errorf("bd graph %s does not name its dependent %s: %s", parent, child, out)
				}
			})
			//ROWS
		})
	}
}
