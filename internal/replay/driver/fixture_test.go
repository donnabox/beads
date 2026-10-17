package driver

import (
	"os"
	"testing"

	"github.com/steveyegge/beads/internal/replay/replaytest"
)

func TestMain(m *testing.M) {
	code := replaytest.Main(m)
	removeSharedFixtures()
	os.Exit(code)
}

// ---- fixture helpers ------------------------------------------------------
//
// Thin wrappers over replaytest, so every test reads as it always did while bd
// comes from a build of this tree and dolt fixtures carry their own identity.

func requireBd(t *testing.T) {
	t.Helper()
	replaytest.Require(t, replaytest.NeedBd)
}

func requireDolt(t *testing.T) {
	t.Helper()
	replaytest.Require(t, replaytest.NeedDolt)
}

// runBd runs bd, built from this tree, in dir with the sanitized environment.
func runBd(t *testing.T, dir string, args ...string) string {
	t.Helper()
	return replaytest.RunBd(t, replaytest.BdBin(t), dir, args...)
}

// runBdBin runs a caller-chosen bd binary. Needed wherever a fixture must stay
// on the exact same schema and behavior as another binary under test in the
// same test (the e2e test's oracle side uses this to match the integration
// build): an older bd would silently lack columns the current tree already has.
func runBdBin(t *testing.T, bin, dir string, args ...string) string {
	t.Helper()
	return replaytest.RunBd(t, bin, dir, args...)
}

func runDolt(t *testing.T, dir string, args ...string) string {
	t.Helper()
	requireDolt(t)
	return replaytest.RunDolt(t, dir, args...)
}

func initBdProject(t *testing.T, name string) string {
	t.Helper()
	return initBdProjectWith(t, name, replaytest.BdBin(t))
}

func initBdProjectWith(t *testing.T, name, bdBin string) string {
	t.Helper()
	replaytest.Isolate(t)
	return replaytest.InitBdProject(t, name, bdBin)
}

// dataDir returns the embedded Dolt data directory bd init created under dir.
func dataDir(t *testing.T, dir string) string {
	t.Helper()
	return replaytest.DataDir(t, dir)
}

func headCommit(t *testing.T, dir string) string {
	t.Helper()
	return replaytest.HeadCommit(t, dir)
}

func jsonID(t *testing.T, out string) string {
	t.Helper()
	return replaytest.JSONID(t, out)
}

func parseCSV(t *testing.T, out string) [][]string {
	t.Helper()
	return replaytest.ParseCSV(t, out)
}
