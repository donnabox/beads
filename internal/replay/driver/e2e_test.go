package driver

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/replay/doltcli"
	"github.com/steveyegge/beads/internal/replay/replaytest"
)

// mergeSide merges the side project's history into the oracle as a side branch:
// the side project, a copy of the oracle taken at its base and advanced on its
// own, is pushed to a file remote and fetched back, and its tip is merged into
// the oracle's mainline with a merge commit. Everything in it was written by bd.
func mergeSide(t *testing.T, o, side *oracleHistory) {
	t.Helper()
	remote := filepath.Join(t.TempDir(), "side-remote")
	if err := os.MkdirAll(remote, 0o755); err != nil {
		t.Fatalf("creating %s: %v", remote, err)
	}
	url := "file://" + remote
	replaytest.RunDolt(t, side.data, "remote", "add", "up", url)
	replaytest.RunDolt(t, side.data, "push", "up", "main")
	replaytest.RunDolt(t, o.data, "remote", "add", "up", url)
	replaytest.RunDolt(t, o.data, "fetch", "up")
	replaytest.RunDolt(t, o.data, "merge", "up/main", "--no-ff", "-m", "merge")
	o.step()
}

// The exit fixture: one corpus with a merge, a commit that changes several
// issues, creates that carry edges, a delete of an issue that has an edge, and
// one step the translator cannot express, replayed through the real loop with an
// Observer attached.
func TestB6ExitFixture(t *testing.T) {
	requireBd(t)
	requireDolt(t)

	o := newOracle(t, t.TempDir(), "oracle")
	base := o.head()
	// The side project is a copy of the oracle at its base, advanced on its own.
	sideDir := filepath.Join(t.TempDir(), "side")
	copyProject(t, o.dir, sideDir)
	side := &oracleHistory{t: t, bin: o.bin, dir: sideDir, data: replaytest.DataDir(t, sideDir)}

	a := o.create("Alpha")                                                                                                     // 0: a create
	b := o.create("Beta", "--deps", a)                                                                                         // 1: a create with an edge
	o.sqlStep("test: priorities of two issues in one commit", "UPDATE issues SET priority = 1 WHERE id IN ('"+a+"', '"+b+"')") // 2: one commit, two issues
	o.sqlStep("test: direct owner mutation", "UPDATE issues SET owner = 'someone-else' WHERE id = '"+a+"'")                    // 3: no bd form
	o.run("update", a, "--description", "after the step bd cannot express")                                                    // 4: touches the quarantined issue
	o.run("update", b, "--title", "Beta two")                                                                                  // 5
	c := o.create("Gamma", "--deps", b)                                                                                        // 6: a create with an edge
	o.run("delete", c, "--force")                                                                                              // 7: a delete of an issue with an edge
	m := o.create("Mainline")                                                                                                  // 8: the mainline side of the merge
	o.run("close", m, "--reason", "done")                                                                                      // 9: a close
	o.run("dep", "remove", b, a)                                                                                               // 10: an edge removed
	f1 := side.create("Side one")
	f2 := side.create("Side two")
	mergeSide(t, o, side) // 11: the merge step, first parent to the merge
	if len(o.heads) != 12 {
		t.Fatalf("the fixture made %d steps, want 12", len(o.heads))
	}
	h := o.heads

	bin, _ := bdStandIn(t, o.bin, "")
	f := newFixtureRun(t, o, bin)
	rec := &recorder{}
	f.cfg.Observer = rec
	run := mustComplete(t, f)

	want := []struct {
		step    int
		issue   string
		verdict Verdict
		kind    string
	}{
		{0, a, VerdictMatched, "create"},
		{1, b, VerdictMatched, "create+dep_add"},
		{2, a, VerdictMatched, "update"},
		{2, b, VerdictMatched, "update"},
		{3, a, VerdictUntranslatable, ""},
		{4, a, VerdictSkippedQuarantined, "update"},
		{5, b, VerdictMatched, "update"},
		{6, c, VerdictMatched, "create+dep_add"},
		{7, c, VerdictMatched, "delete"},
		{8, m, VerdictMatched, "create"},
		{9, m, VerdictMatched, "close"},
		{10, b, VerdictMatched, "dep_remove"},
		{11, f1, VerdictMatched, "merge:create"},
		{11, f2, VerdictMatched, "merge:create"},
	}
	sort.SliceStable(want, func(i, j int) bool {
		if want[i].step != want[j].step {
			return want[i].step < want[j].step
		}
		return want[i].issue < want[j].issue
	})
	rows := f.results()
	if len(rows) != len(want) {
		t.Fatalf("%d rows, want %d: %+v", len(rows), len(want), rows)
	}
	for i, w := range want {
		got := rows[i]
		if got.SourceCommit != h[w.step] || got.IssueID != w.issue {
			t.Errorf("row %d is %s at %s, want %s at step %d (%s)", i, got.IssueID, got.SourceCommit, w.issue, w.step, h[w.step])
			continue
		}
		if got.Verdict != w.verdict || got.MutationKind != w.kind || got.RunID != run.ID {
			t.Errorf("row %d (%s at step %d) = verdict %q kind %q run %q, want %q %q %q",
				i, w.issue, w.step, got.Verdict, got.MutationKind, got.RunID, w.verdict, w.kind, run.ID)
		}
		if w.step > 0 && got.FromCommit != h[w.step-1] {
			t.Errorf("row %d (%s at step %d) starts at %s, want the previous step's commit %s", i, w.issue, w.step, got.FromCommit, h[w.step-1])
		}
	}
	if got := f.resultFor(h[3], a).Detail; got == nil || !reflect.DeepEqual(got.Columns, []string{"owner"}) {
		t.Errorf("the untranslatable row names %+v, want the owner column", got)
	}
	if got := f.resultFor(h[4], a).Detail; got == nil || got.QuarantinedAt != h[3] {
		t.Errorf("the skipped row names %+v, want the step that quarantined the issue (%s)", got, h[3])
	}

	// The Observer saw exactly the rows that were written, in order, and the merge
	// step is flagged; the loop emitted no seed event.
	if len(rec.events) != len(rows) {
		t.Fatalf("%d events for %d rows", len(rec.events), len(rows))
	}
	for i, ev := range rec.events {
		got, ok := ev.(StepIssueResult)
		if !ok {
			t.Fatalf("event %d is %T, want StepIssueResult", i, ev)
		}
		if !reflect.DeepEqual(got.Result, rows[i]) {
			t.Errorf("event %d carries %+v, want the row written: %+v", i, got.Result, rows[i])
		}
		if got.Step != want[i].step || got.Merge != (want[i].step == 11) {
			t.Errorf("event %d is step %d merge=%v, want step %d merge=%v", i, got.Step, got.Merge, want[i].step, want[i].step == 11)
		}
	}

	// The summary states what the run covered.
	s := f.summary()
	wantVerdicts := map[string]int{"matched": 12, "mismatch": 0, "uncomparable": 0, "untranslatable": 1, "rejected": 0, "skipped-quarantined": 1}
	if !reflect.DeepEqual(s.Verdicts, wantVerdicts) {
		t.Errorf("summary verdicts = %v, want %v", s.Verdicts, wantVerdicts)
	}
	if s.Gaps.Columns["issues.owner"] != 1 {
		t.Errorf("summary gaps by column = %v, want issues.owner once", s.Gaps.Columns)
	}
	if s.Status != "completed" || s.Mode != "exhaustive" || s.RunID != run.ID {
		t.Errorf("summary header = %q %q %q", s.Status, s.Mode, s.RunID)
	}
	_, logRows, err := doltcli.Query(context.Background(), o.data, "SELECT COUNT(*) FROM dolt_log")
	if err != nil {
		t.Fatalf("counting the oracle's commits: %v", err)
	}
	cr := s.CoveredRange
	if cr.BaseCommit != base || cr.HeadCommit != h[11] || cr.CommitsReplayed != 12 || cr.CommitsTotal <= 12 || strconv.Itoa(cr.CommitsTotal) != logRows[0][0].Text {
		t.Errorf("covered_range = %+v, want base %s, head %s, 12 replayed and every commit of the history (%s) counted", cr, base, h[11], logRows[0][0].Text)
	}
	if cr.BaseDate == "" || cr.HeadDate == "" {
		t.Errorf("covered_range dates are empty: %+v", cr)
	}
	for _, g := range readJSONLTyped[CoverageGapRow](t, filepath.Join(f.outDir, "coverage_gaps.jsonl")) {
		if g.Unknown {
			t.Errorf("a table outside the policy changed: %+v", g)
		}
	}

	// The work clone holds what the replay built, with history on.
	if got := f.workCount("SELECT COUNT(*) FROM issues WHERE id IN ('" + m + "', '" + f1 + "', '" + f2 + "')"); got != 3 {
		t.Errorf("the work clone holds %d of the mainline and side-branch issues, want 3", got)
	}
	if got := f.workCount("SELECT COUNT(*) FROM issues WHERE id = '" + c + "'"); got != 0 {
		t.Errorf("the deleted issue %s is still in the work clone", c)
	}
	if got := f.workCount("SELECT COUNT(*) FROM issue_versions"); got == 0 {
		t.Error("issue_versions is empty after the run")
	}
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		t.Fatalf("git rev-parse --show-toplevel: %v", err)
	}
	return strings.TrimSpace(string(out))
}

func readCommitReplayResults(t *testing.T, outDir string) []CommitReplayResult {
	t.Helper()
	return readJSONLTyped[CommitReplayResult](t, filepath.Join(outDir, "commit_replay_results.jsonl"))
}

func readMismatches(t *testing.T, outDir string) []Mismatch {
	t.Helper()
	return readJSONLTyped[Mismatch](t, filepath.Join(outDir, "mismatches.jsonl"))
}

func readJSONLTyped[T any](t *testing.T, path string) []T {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("reading %s: %v", path, err)
	}
	var out []T
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line == "" {
			continue
		}
		var v T
		if err := json.Unmarshal([]byte(line), &v); err != nil {
			t.Fatalf("unmarshaling line from %s: %v (line: %s)", path, err, line)
		}
		out = append(out, v)
	}
	return out
}
