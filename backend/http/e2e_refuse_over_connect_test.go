//go:build cgo

package bdhttp_test

import (
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/httpapi"
)

// A command bd serve cannot serve refuses over bd connect, nonzero, and says why
// in words that are true of a client of a server. Each row names what the old
// message said, or what the command did instead of refusing, as a thing that must
// not come back. A yes on these rows applies to both protocols.
func TestE2E_ConnectRefusalsSayWhy(t *testing.T) {
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

			for _, row := range []struct {
				name string
				args []string
				want string
				not  []string
			}{
				{
					name: "E2 gc", args: []string{"gc", "--dry-run"},
					want: "bd gc is not available over bd connect: it maintains the server's own database",
					not:  []string{"SearchIssues", "searching closed issues"},
				},
				{
					name: "E2 gc --json", args: []string{"gc", "--dry-run", "--json"},
					want: "bd gc is not available over bd connect",
					not:  []string{"SearchIssues", "searching closed issues"},
				},
				{
					name: "E1 sql", args: []string{"sql", "select 1"},
					want: "'bd sql' is not available over bd connect: bd serve publishes no SQL surface",
					not:  []string{"embedded mode"},
				},
				{
					name: "E3 sync", args: []string{"sync"},
					want: "bd sync is not available over bd connect: bd serve is the sync boundary",
					not:  []string{"is_blocked recompute"},
				},
			} {
				t.Run(row.name, func(t *testing.T) {
					r := f.run(nil, row.args...)
					said := r.stdout + r.stderr
					if r.code == 0 {
						t.Fatalf("bd %v exited 0: %s", row.args, said)
					}
					if !strings.Contains(said, row.want) {
						t.Errorf("bd %v: want %q in: %s", row.args, row.want, said)
					}
					for _, stale := range row.not {
						if strings.Contains(said, stale) {
							t.Errorf("bd %v still says %q: %s", row.args, stale, said)
						}
					}
				})
			}
		})
	}
}
