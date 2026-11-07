package createbatchequiv

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/testutil/bazeltest"
)

// The tests here pin fork_golden_deltas.go, the fork-owned and exact override
// of the checked-in (upstream) golden digests. None of them needs a database:
// the drivers in the embeddeddolt and dolt packages are what hold the real
// code to the pinned digests. These hold the override itself to its contract.

// forkGoldenDeltaWorkOrder is what architect Ruling B of be-0a9kgp says this
// fork stores differently from upstream's pre-change golden, by design and
// nowhere else. Per (scenario, table): the mechanism, the row counts on both
// sides, and the issues whose rows differ (nil: every upstream row differs, so
// the fork stores none of them).
var forkGoldenDeltaWorkOrder = []struct {
	scenario, table        string
	mechanism              forkMechanism
	forkRows, upstreamRows int
	issueIDs               []string
}{
	{"small", "issue_versions", mechanismFenceLegacyUnversioned, 37, 40, []string{"eq-s1", "eq-s2", "eq-s4"}},
	{"depadd", "issue_versions", mechanismFenceLegacyUnversioned, 0, 11, nil},
	{"waitsfor", "issue_versions", mechanismFenceLegacyUnversioned, 0, 15, nil},
	{"import458", "issue_versions", mechanismFenceLegacyUnversioned, 394, 398, []string{"eq-ep", "eq-s1", "eq-s4", "eq-s5"}},
	{"import458-reject-stale", "issue_versions", mechanismFenceLegacyUnversioned, 394, 396, []string{"eq-ep", "eq-s1"}},
	{"small", "issues_updated_at", mechanismRecorderKeepsUpdatedAt, 42, 41, []string{"eq-n2"}},
	{"depadd", "issues_updated_at", mechanismRecorderKeepsUpdatedAt, 15, 13, []string{"eq-db1", "eq-dl"}},
	{"import458", "issues_updated_at", mechanismRecorderKeepsUpdatedAt, 403, 401, []string{"eq-b250", "eq-b3"}},
	{"import458-reject-stale", "issues_updated_at", mechanismRecorderKeepsUpdatedAt, 402, 400, []string{"eq-b250", "eq-b3"}},
}

// forkGoldenDeltaDigest is the digest the harness would record for a table
// holding rows.
func forkGoldenDeltaDigest(rows ...string) goldenTable {
	return digestOf(Outcome{Tables: map[string][]string{"t": rows}}).Tables["t"]
}

// forkGoldenDeltaText renders a digest the way checkGolden's own failure does.
func forkGoldenDeltaText(g goldenTable) string {
	return fmt.Sprintf("%d rows %s", g.Rows, g.SHA256)
}

// forkGoldenDeltaCheckedIn reads one scenario's checked-in golden.
func forkGoldenDeltaCheckedIn(t *testing.T, name string) golden {
	t.Helper()
	b, err := goldenFS.ReadFile("golden/" + name + ".json")
	if err != nil {
		t.Fatalf("golden/%s.json: %v", name, err)
	}
	var g golden
	if err := json.Unmarshal(b, &g); err != nil {
		t.Fatalf("golden/%s.json: %v", name, err)
	}
	return g
}

// forkGoldenDeltaIssueIDs reduces pinned row ids to the sorted, distinct ids
// of the issues they belong to: an issue_versions row id is
// "<issue id>@<revision>", an issues_updated_at row id is the issue id.
func forkGoldenDeltaIssueIDs(rowIDs []string) []string {
	seen := map[string]bool{}
	for _, id := range rowIDs {
		issue, _, _ := strings.Cut(id, "@")
		seen[issue] = true
	}
	return sortedKeys(seen)
}

// forkGoldenDeltaProbeEnv switches a test into its child half. checkGolden
// reports through *testing.T, so a test that wants to watch it fail re-runs
// itself in a child process with the probe mode set and reads what the child
// printed (the repo's idiom for asserting a failure; see
// backend/conformance/role_bundle_test.go). Only the child ever replaces
// forkGoldenDeltas.
const forkGoldenDeltaProbeEnv = "CREATEBATCHEQUIV_FORKGOLDENDELTA_PROBE"

// runForkGoldenDeltaProbe re-runs the calling test in a child process with the
// probe mode set, and returns what the child printed and whether it failed.
func runForkGoldenDeltaProbe(t *testing.T, mode string) (out string, failed bool) {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=^"+t.Name()+"$", "-test.v")
	cmd.Env = append(bazeltest.ShardFreeEnv(os.Environ()), forkGoldenDeltaProbeEnv+"="+mode)
	b, err := cmd.CombinedOutput()
	var exit *exec.ExitError
	if err != nil && !errors.As(err, &exit) {
		t.Fatalf("running the %s probe: %v\n%s", mode, err, b)
	}
	// A child that selected no test exits 0 and prints nothing about the
	// scenario; reading that as "checkGolden did not fail" would pass vacuously.
	if !regexp.MustCompile(`(?m)^=== RUN\s+` + regexp.QuoteMeta(t.Name()) + `$`).Match(b) {
		t.Fatalf("the %s probe did not run %s:\n%s", mode, t.Name(), b)
	}
	return string(b), err != nil
}

// forkGoldenDeltaProbeTarget is what the probes run checkGolden against: the
// "apply" scenario, which carries no real pin, one table of its checked-in
// golden (the first by name, so the pick is stable) and that table's
// checked-in digest. Each probe installs its own pin over it.
func forkGoldenDeltaProbeTarget(t *testing.T) (name, table string, checkedIn goldenTable) {
	t.Helper()
	name = "apply"
	g := forkGoldenDeltaCheckedIn(t, name)
	table = sortedKeys(g.Tables)[0]
	return name, table, g.Tables[table]
}

// A pin derived against a golden that has since changed (an upstream sync
// re-recorded it) must stop the check at once, naming the entry, rather than
// compare against a digest nobody has vouched for.
func TestForkGoldenDeltaStalePinFailsLoudly(t *testing.T) {
	name, table, checkedIn := forkGoldenDeltaProbeTarget(t)
	derivedAgainst := forkGoldenDeltaDigest("the golden this pin was derived against")
	if os.Getenv(forkGoldenDeltaProbeEnv) == "stale" {
		forkGoldenDeltas = map[forkGoldenKey]forkGoldenDelta{
			{name, table}: {
				Mechanism: mechanismFenceLegacyUnversioned,
				Upstream:  derivedAgainst,
				Fork:      forkGoldenDeltaDigest("what the fork stores"),
				RowIDs:    []string{"eq-x@1"},
			},
		}
		checkGolden(t, name, Outcome{})
		return
	}

	out, failed := runForkGoldenDeltaProbe(t, "stale")
	if !failed {
		t.Fatalf("a pin derived against a golden that no longer states that digest was accepted:\n%s", out)
	}
	for _, want := range []string{
		"stale pin",
		name,
		table,
		string(mechanismFenceLegacyUnversioned),
		forkGoldenDeltaText(checkedIn),
		forkGoldenDeltaText(derivedAgainst),
	} {
		if !strings.Contains(out, want) {
			t.Errorf("the failure does not mention %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "differs from the pre-change golden") {
		t.Errorf("the check went on to compare digests against a stale pin instead of stopping on it:\n%s", out)
	}
}

// An override whose fork digest equals the upstream digest overrides nothing:
// the fork stores what upstream stores, and the entry is dead weight that
// would hide a real divergence the day one appears. It must fail, not pass.
func TestForkGoldenDeltaUnneededOverrideFails(t *testing.T) {
	name, table, checkedIn := forkGoldenDeltaProbeTarget(t)
	if os.Getenv(forkGoldenDeltaProbeEnv) == "unneeded" {
		forkGoldenDeltas = map[forkGoldenKey]forkGoldenDelta{
			{name, table}: {
				Mechanism: mechanismRecorderKeepsUpdatedAt,
				Upstream:  checkedIn,
				Fork:      checkedIn,
				RowIDs:    []string{"eq-x"},
			},
		}
		checkGolden(t, name, Outcome{})
		return
	}

	out, failed := runForkGoldenDeltaProbe(t, "unneeded")
	if !failed {
		t.Fatalf("an override whose fork digest equals the upstream digest was accepted:\n%s", out)
	}
	for _, want := range []string{"no longer needed", name, table, string(mechanismRecorderKeepsUpdatedAt)} {
		if !strings.Contains(out, want) {
			t.Errorf("the failure does not mention %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "differs from the pre-change golden") {
		t.Errorf("the check went on to compare digests against an unneeded override instead of stopping on it:\n%s", out)
	}
}

// With a pin in place the check is exact: the actual digest must equal the
// fork digest. A table that still matches the pin passes, and one that does
// not (including one that fell back to what upstream stores) fails against
// the fork digest, not the upstream one.
func TestForkGoldenDeltaActualDigestMustEqualTheForkDigest(t *testing.T) {
	name, table, checkedIn := forkGoldenDeltaProbeTarget(t)
	forkRows := []string{`"fork-row-1"`, `"fork-row-2"`}
	fork := forkGoldenDeltaDigest(forkRows...)
	pin := func() {
		forkGoldenDeltas = map[forkGoldenKey]forkGoldenDelta{
			{name, table}: {
				Mechanism: mechanismFenceLegacyUnversioned,
				Upstream:  checkedIn,
				Fork:      fork,
				RowIDs:    []string{"eq-x@1", "eq-y@1"},
			},
		}
	}
	// Every other table of the golden goes unmatched in these probes (they
	// supply none), so the child fails either way; only this table's line says
	// whether the override was applied.
	switch os.Getenv(forkGoldenDeltaProbeEnv) {
	case "exact":
		pin()
		checkGolden(t, name, Outcome{Tables: map[string][]string{table: forkRows}})
		t.Log("probe finished: exact")
		return
	case "drift":
		pin()
		checkGolden(t, name, Outcome{Tables: map[string][]string{table: forkRows[:1]}})
		t.Log("probe finished: drift")
		return
	}

	differs := fmt.Sprintf("%s: %s differs from the pre-change golden", name, table)
	line := func(out string) string {
		for _, l := range strings.Split(out, "\n") {
			if strings.Contains(l, differs) {
				return l
			}
		}
		return ""
	}

	out, _ := runForkGoldenDeltaProbe(t, "exact")
	if !strings.Contains(out, "probe finished: exact") {
		t.Fatalf("the exact probe did not get through checkGolden:\n%s", out)
	}
	if l := line(out); l != "" {
		t.Errorf("a table whose digest equals the fork digest was reported as differing:\n%s", l)
	}

	out, _ = runForkGoldenDeltaProbe(t, "drift")
	if !strings.Contains(out, "probe finished: drift") {
		t.Fatalf("the drift probe did not get through checkGolden:\n%s", out)
	}
	l := line(out)
	if l == "" {
		t.Fatalf("a table whose digest is not the fork digest was accepted:\n%s", out)
	}
	if want := "want " + forkGoldenDeltaText(fork); !strings.Contains(l, want) {
		t.Errorf("the table was not held to the fork digest (%q):\n%s", want, l)
	}
	if strings.Contains(l, "want "+forkGoldenDeltaText(checkedIn)) {
		t.Errorf("the table was still held to the upstream digest:\n%s", l)
	}
}

// The swap touches exactly the pinned (scenario, table): the scenario's other
// tables and its verbatim lists stay as the checked-in golden states them, and
// another scenario's pin is neither applied nor judged against this golden.
func TestForkGoldenDeltaSwapsOnlyThePinnedTable(t *testing.T) {
	upstream, fork, other := forkGoldenDeltaDigest("up"), forkGoldenDeltaDigest("fork"), forkGoldenDeltaDigest("other")
	deltas := map[forkGoldenKey]forkGoldenDelta{
		{"s1", "pinned"}: {
			Mechanism: mechanismFenceLegacyUnversioned,
			Upstream:  upstream,
			Fork:      fork,
			RowIDs:    []string{"eq-x@1"},
		},
		{"s2", "elsewhere"}: {
			Mechanism: mechanismRecorderKeepsUpdatedAt,
			Upstream:  other,
			Fork:      fork,
			RowIDs:    []string{"eq-y"},
		},
	}
	newWant := func() golden {
		return golden{
			Tables:  map[string]goldenTable{"pinned": upstream, "untouched": other},
			Skipped: []string{"eq-a -> eq-b: reason"},
		}
	}

	want := newWant()
	if err := swapForkGoldenDeltas("s1", &want, deltas); err != nil {
		t.Fatalf("a current pin was refused: %v", err)
	}
	if want.Tables["pinned"] != fork {
		t.Errorf("the pinned table holds %v, want the fork digest %v", want.Tables["pinned"], fork)
	}
	if want.Tables["untouched"] != other || len(want.Tables) != 2 {
		t.Errorf("the swap changed tables it does not pin: %v", want.Tables)
	}
	if !slices.Equal(want.Skipped, newWant().Skipped) {
		t.Errorf("the swap changed a verbatim list: %v", want.Skipped)
	}

	unpinned := newWant()
	if err := swapForkGoldenDeltas("s3", &unpinned, deltas); err != nil {
		t.Fatalf("a scenario with no pin was refused: %v", err)
	}
	if unpinned.Tables["pinned"] != upstream || unpinned.Tables["untouched"] != other {
		t.Errorf("a scenario with no pin was changed: %v", unpinned.Tables)
	}
}

// Ruling B names nine entries and no others: a tenth would be a divergence
// nobody has approved, and one fewer would leave a driver failing.
func TestForkGoldenDeltaPinsTheNineWorkOrderEntries(t *testing.T) {
	if len(forkGoldenDeltas) != len(forkGoldenDeltaWorkOrder) {
		t.Errorf("forkGoldenDeltas pins %d entries, the work order names %d", len(forkGoldenDeltas), len(forkGoldenDeltaWorkOrder))
	}
	for _, w := range forkGoldenDeltaWorkOrder {
		key := forkGoldenKey{w.scenario, w.table}
		d, ok := forkGoldenDeltas[key]
		if !ok {
			t.Errorf("%s is not pinned", key)
			continue
		}
		if d.Mechanism != w.mechanism {
			t.Errorf("%s: mechanism %q, want %q", key, d.Mechanism, w.mechanism)
		}
		if d.Fork.Rows != w.forkRows || d.Upstream.Rows != w.upstreamRows {
			t.Errorf("%s: %d rows (fork) vs %d (upstream), want %d vs %d", key, d.Fork.Rows, d.Upstream.Rows, w.forkRows, w.upstreamRows)
		}
		if w.issueIDs == nil {
			if d.Fork.Rows != 0 || len(d.RowIDs) != d.Upstream.Rows {
				t.Errorf("%s: want every one of upstream's %d rows pinned as missing, got %d row ids and %d fork rows",
					key, d.Upstream.Rows, len(d.RowIDs), d.Fork.Rows)
			}
			continue
		}
		if got := forkGoldenDeltaIssueIDs(d.RowIDs); !slices.Equal(got, w.issueIDs) {
			t.Errorf("%s: differing issues %v, want %v", key, got, w.issueIDs)
		}
	}
}

// Whatever the pins say, they must be internally consistent and still state
// the checked-in golden: the same stale-pin check the drivers run, here
// against every golden without a database, so a re-recorded golden fails
// this package's own tests and not only the two backends'.
func TestForkGoldenDeltaPinsStateTheCheckedInGolden(t *testing.T) {
	known := map[string]bool{}
	for _, sc := range append(scenarios(), largeScenarios()...) {
		known[sc.name] = true
	}
	// Which table each mechanism accounts for, how the fork's row count moves
	// against upstream's per differing row, and what a row id looks like.
	shape := map[forkMechanism]struct {
		table string
		sign  int
		idRe  *regexp.Regexp
	}{
		mechanismFenceLegacyUnversioned: {"issue_versions", -1, regexp.MustCompile(`^eq-[^@\s]+@[1-9][0-9]*$`)},
		mechanismRecorderKeepsUpdatedAt: {"issues_updated_at", +1, regexp.MustCompile(`^eq-[^@\s]+$`)},
	}
	sha256Hex := regexp.MustCompile(`^[0-9a-f]{64}$`)

	pinned := map[string]bool{}
	for _, key := range sortedForkGoldenKeys(forkGoldenDeltas) {
		d := forkGoldenDeltas[key]
		pinned[key.Scenario] = true
		if !known[key.Scenario] {
			t.Errorf("%s names a scenario no driver runs", key)
		}
		s, ok := shape[d.Mechanism]
		if !ok {
			t.Errorf("%s: unknown mechanism %q", key, d.Mechanism)
			continue
		}
		if key.Table != s.table {
			t.Errorf("%s: mechanism %s accounts for table %s only", key, d.Mechanism, s.table)
		}
		if d.Fork == d.Upstream {
			t.Errorf("%s: the fork digest equals the upstream digest, so there is nothing to override", key)
		}
		if !sha256Hex.MatchString(d.Upstream.SHA256) || !sha256Hex.MatchString(d.Fork.SHA256) {
			t.Errorf("%s: a digest is not a lowercase hex sha256: upstream %q, fork %q", key, d.Upstream.SHA256, d.Fork.SHA256)
		}
		if len(d.RowIDs) == 0 || !sort.StringsAreSorted(d.RowIDs) || len(slices.Compact(slices.Clone(d.RowIDs))) != len(d.RowIDs) {
			t.Errorf("%s: row ids must be non-empty, sorted and distinct: %v", key, d.RowIDs)
		}
		for _, id := range d.RowIDs {
			if !s.idRe.MatchString(id) {
				t.Errorf("%s: row id %q is not shaped like a %s row id", key, id, key.Table)
			}
		}
		if want := d.Upstream.Rows + s.sign*len(d.RowIDs); d.Fork.Rows != want {
			t.Errorf("%s: the fork has %d rows against upstream's %d with %d differing rows; %s implies %d",
				key, d.Fork.Rows, d.Upstream.Rows, len(d.RowIDs), d.Mechanism, want)
		}
	}

	for _, name := range sortedKeys(pinned) {
		if !known[name] {
			continue
		}
		want := forkGoldenDeltaCheckedIn(t, name)
		if err := swapForkGoldenDeltas(name, &want, forkGoldenDeltas); err != nil {
			t.Errorf("the pins for %s no longer match golden/%s.json:\n%v", name, name, err)
			continue
		}
		for _, key := range sortedForkGoldenKeys(forkGoldenDeltas) {
			if key.Scenario == name && want.Tables[key.Table] != forkGoldenDeltas[key].Fork {
				t.Errorf("%s: the swap left %v in place of the fork digest", key, want.Tables[key.Table])
			}
		}
	}
}
