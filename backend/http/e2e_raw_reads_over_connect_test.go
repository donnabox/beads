//go:build cgo

package bdhttp_test

import (
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/httpapi"
)

// TestE2E_RawReadsServeCommandsOverConnect runs, through the real bd on both
// protocols, the commands rows B15 and B16 stop refusing: each used to answer
// "not supported" or "not expressible", and each now answers from v0 operations.
func TestE2E_RawReadsServeCommandsOverConnect(t *testing.T) {
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
			addr := startProtocolServer(t, tc.server)
			f := &remoteRolesFixture{t: t, bin: bin, dir: t.TempDir(), ids: map[string]string{}}
			f.mustRun(nil, append([]string{"connect", "http://" + addr}, tc.connect...)...)

			target := f.create("listing target")
			source := f.create("listing source")
			f.mustRun(nil, "dep", "add", source, target)

			t.Run("B16 listings through a raw search", func(t *testing.T) {
				for _, args := range [][]string{
					{"lint", "--json"},
					{"orphans", "--json"},
					{"duplicates", "--json"},
					{"find-duplicates", "--json"},
					{"human", "list", "--json"},
					{"label", "list-all", "--json"},
					{"swarm", "list", "--json"},
					{"graph", "--all", "--json"},
					{"todo", "--json"},
				} {
					r := f.run(nil, args...)
					said := r.stdout + r.stderr
					if strings.Contains(said, "not supported") || strings.Contains(said, "not expressible") || strings.Contains(said, "matching neither") {
						t.Errorf("bd %v still refuses: %s", args, said)
					}
				}
			})

			t.Run("B15 B16 ready --gated and mol ready", func(t *testing.T) {
				for _, args := range [][]string{{"ready", "--gated", "--json"}, {"mol", "ready", "--json"}} {
					if r := f.run(nil, args...); r.code != 0 {
						t.Errorf("bd %v exit %d: %s %s", args, r.code, r.stdout, r.stderr)
					}
				}
			})

			t.Run("B15 stale, epic status and mol stale", func(t *testing.T) {
				if out := f.mustRun(nil, "stale", "--json").stdout; strings.Contains(out, target) {
					t.Errorf("bd stale calls %s stale the moment it was made: %s", target, out)
				}
				epic := strings.TrimSpace(f.mustRun(nil, "create", "status epic", "-t", "epic", "--silent").stdout)
				f.mustRun(nil, "create", "status child", "--parent", epic, "--silent")
				if out := f.mustRun(nil, "epic", "status", "--json").stdout; !strings.Contains(out, "status epic") {
					t.Errorf("bd epic status does not name the epic: %s", out)
				}
				f.mustRun(nil, "mol", "stale", "--json")
			})

			t.Run("B15 delete previews", func(t *testing.T) {
				leaf := f.create("delete preview leaf")
				if out := f.mustRun(nil, "delete", leaf).stdout; !strings.Contains(out, "delete preview leaf") {
					t.Errorf("bd delete's preview does not name the issue: %s", out)
				}
				f.mustRun(nil, "show", leaf, "--json")
				if out := f.mustRun(nil, "delete", source, "--dry-run").stdout; !strings.Contains(out, target) {
					t.Errorf("bd delete --dry-run does not name the edge's target: %s", out)
				}
				f.mustRun(nil, "show", source, "--json")
			})
		})
	}
}
