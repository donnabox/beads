// Package compatwindow is a scenario test for the window in which an old bd
// binary and the current one share one store.
//
// The current binary has versioned history switched on in the store, so its
// writes record versions and stamp each record it creates as taking part in
// history (participation_generation holds a value). An old binary predates that
// column. It is built here, hermetically, from a named upstream commit, never
// taken from PATH, and runs against a store the current binary has already
// stamped. Both are real processes sharing one embedded Dolt store, and the
// scenario reads the store's version rows, current_revision and
// participation_generation (NULL meaning legacy) before and after each step.
//
// What the scenario pins:
//
//  1. The old binary is refused by default. Opening a store whose schema is
//     ahead of it, it exits non-zero, names both schema versions, carries a
//     structured schema_skew block under --json, and writes nothing.
//  2. The schema-skew hatch downgrades that refusal to a warning and the old
//     binary then writes. What its writes do to the versioning columns is
//     recorded, not presumed: see TestHatchAdmitsLegacyWriter.
//  3. The write fence is a skip performed by the current binary, not a refusal,
//     and it does not run inside the old one. When the current binary updates a
//     record the legacy writer last wrote, which still carries no participation
//     generation, it mints no version row and moves no revision, with the hatch
//     set in its own environment or not.
//  4. Fixture only. Every bd runs in a project created under the test's own
//     temporary directory, and the scenario stops before any write unless bd
//     resolves that project's workspace and nothing above it holds another.
//
// Versioned history is read back and checked live before anything else is
// asserted, because with history off the fence never runs and every later
// assertion would pass without testing anything.
//
// The in-process cases for the fence, including the hatch's lack of any bearing
// on it, live with the storage conformance contracts; this scenario adds only
// the dimension that needs two binaries.
package compatwindow

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/replay/replaytest"
)

func TestMain(m *testing.M) {
	code := replaytest.Main(m)
	if oldBuild.dir != "" {
		_ = os.RemoveAll(oldBuild.dir)
	}
	os.Exit(code)
}

// The old binary opens a store that is ahead of it only when told to. Without
// the hatch every command is refused, a refusal is a statement and not a
// partial write, and the structured form carries the two versions.
func TestOldWriterIsRefusedByDefault(t *testing.T) {
	c := newCompat(t)
	before := c.record(c.participating)

	out, code := c.run(c.oldBin, nil, "list")
	if code == 0 {
		t.Fatalf("the legacy binary opened a store ahead of it and exited 0:\n%s", out)
	}
	for _, want := range []string{"schema version mismatch", fmt.Sprintf("v%d", c.dbSchema), fmt.Sprintf("v%d", c.oldSchema)} {
		if !strings.Contains(out, want) {
			t.Errorf("the refusal does not contain %q:\n%s", want, out)
		}
	}

	jsonOut, code := c.run(c.oldBin, nil, "list", "--json")
	if code == 0 {
		t.Fatalf("the legacy binary opened a store ahead of it under --json and exited 0:\n%s", jsonOut)
	}
	skew, _ := jsonObject(jsonOut, "schema_skew")["schema_skew"].(map[string]any)
	if skew == nil {
		t.Fatalf("the --json refusal carries no schema_skew block:\n%s", jsonOut)
	}
	for key, want := range map[string]int{
		"current_version":  c.dbSchema,
		"required_version": c.oldSchema,
		"delta":            c.dbSchema - c.oldSchema,
	} {
		if got, _ := skew[key].(float64); got != float64(want) {
			t.Errorf("schema_skew.%s = %v, want %d:\n%s", key, skew[key], want, jsonOut)
		}
	}

	out, code = c.update(c.oldBin, nil, c.participating, "a title that must not land")
	if code == 0 {
		t.Fatalf("the legacy binary wrote to a store ahead of it without the hatch and exited 0:\n%s", out)
	}
	if after := c.record(c.participating); after != before {
		t.Fatalf("a refused write changed the record:\n before: %v\n  after: %v", before, after)
	}
}

// With the hatch the old binary is admitted, and still warns. The assertions on
// what its writes do to the versioning columns record the behavior of the pinned
// legacy binary as observed. They document the compatibility window as it is and
// do not say that behavior is desirable: the binary carries the function that
// records versions but nothing on its command line turns it on, so it never
// records one, and a record it changes keeps its old history.
func TestHatchAdmitsLegacyWriter(t *testing.T) {
	c := newCompat(t)

	// A record that takes part in history.
	before := c.record(c.participating)
	retitled := "participating record, retitled by the legacy writer"
	out, code := c.update(c.oldBin, hatch, c.participating, retitled)
	if code != 0 {
		t.Fatalf("the hatch did not admit the legacy writer (exit %d):\n%s", code, out)
	}
	if !strings.Contains(strings.ToLower(out), "schema skew ignored") {
		t.Errorf("the downgraded refusal did not warn:\n%s", out)
	}
	after := c.record(c.participating)
	t.Logf("legacy writer under the hatch, participating record:\n before: %v\n  after: %v", before, after)
	if after.Title != retitled {
		t.Fatalf("the legacy writer was admitted but its write did not land: %v", after)
	}
	if after.Versions != before.Versions {
		t.Errorf("the legacy writer minted a version row: %d -> %d", before.Versions, after.Versions)
	}
	if after.Revision != before.Revision {
		t.Errorf("the legacy writer moved current_revision: %s -> %s", before.Revision, after.Revision)
	}
	if after.Generation != before.Generation {
		t.Errorf("the legacy writer changed participation_generation: %v -> %v", before.Generation, after.Generation)
	}

	// A record the legacy writer creates takes no part in history, and writing
	// to it again changes nothing about that.
	legacy := c.legacyRecord()
	created := c.record(legacy)
	retitled = "legacy record, retitled by the legacy writer"
	out, code = c.update(c.oldBin, hatch, legacy, retitled)
	if code != 0 {
		t.Fatalf("the legacy writer could not update its own record (exit %d):\n%s", code, out)
	}
	after = c.record(legacy)
	t.Logf("legacy writer under the hatch, legacy record:\n before: %v\n  after: %v", created, after)
	if after.Title != retitled {
		t.Fatalf("the legacy writer's update of its own record did not land: %v", after)
	}
	if !after.Generation.Null || after.Versions != 0 || after.Revision != created.Revision {
		t.Errorf("the legacy writer promoted or versioned a legacy record:\n before: %v\n  after: %v", created, after)
	}
}

// After the legacy writer has been through the store, the current writer still
// skips a record the legacy writer last wrote and that carries no participation
// generation: no version row, no revision bump, no stamp. It does so whether or
// not the hatch is set in its own environment, because the fence reads the row,
// not the schema cursor the hatch downgrades. The first step is the control:
// the same writer does record a version for a record that takes part, so the
// skip below is a decision and not history being off.
func TestCurrentWriterSkipsLegacyRecord(t *testing.T) {
	c := newCompat(t)
	legacy := c.legacyRecord()

	before := c.record(c.participating)
	out, code := c.update(c.newBin, nil, c.participating, "participating record, retitled by the current writer")
	if code != 0 {
		t.Fatalf("the current writer could not update a participating record (exit %d):\n%s", code, out)
	}
	after := c.record(c.participating)
	if after.Versions != before.Versions+1 || after.Revision == before.Revision {
		t.Fatalf("control failed: the current writer recorded no version for a participating record, so a skip below would prove nothing:\n before: %v\n  after: %v", before, after)
	}
	if after.Generation != before.Generation {
		t.Errorf("the current writer changed a participation generation it had no reason to touch: %v -> %v", before.Generation, after.Generation)
	}

	for _, tc := range []struct {
		name string
		env  []string
	}{
		{"hatch unset", nil},
		{"hatch set in the current writer's environment", hatch},
	} {
		title := "legacy record, retitled by the current writer (" + tc.name + ")"
		was := c.record(legacy)
		out, code := c.update(c.newBin, tc.env, legacy, title)
		if code != 0 {
			t.Fatalf("%s: the current writer could not update a legacy record (exit %d):\n%s", tc.name, code, out)
		}
		now := c.record(legacy)
		if now.Title != title {
			t.Fatalf("%s: the update did not land: %v", tc.name, now)
		}
		if now.Versions != was.Versions || now.Revision != was.Revision || now.Generation != was.Generation {
			t.Errorf("%s: the current writer did not skip a legacy record:\n before: %v\n  after: %v", tc.name, was, now)
		}
	}
}

// The scenario's isolation rests on seeing a workspace that sits above its
// project, so that check is pinned directly.
func TestWorkspaceAboveFindsNearestAncestor(t *testing.T) {
	root := t.TempDir()
	mkdir := func(rel string) string {
		t.Helper()
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatal(err)
		}
		return p
	}
	mkdir(".beads")
	mkdir("a/.beads")
	leaf := mkdir("a/b/c")

	if got, want := workspaceAbove(leaf), filepath.Join(root, "a"); got != want {
		t.Errorf("workspaceAbove(%s) = %q, want the nearest ancestor %q", leaf, got, want)
	}
	if got, want := workspaceAbove(filepath.Join(root, "a")), root; got != want {
		t.Errorf("a directory's own workspace is not above it: workspaceAbove = %q, want %q", got, want)
	}
}
