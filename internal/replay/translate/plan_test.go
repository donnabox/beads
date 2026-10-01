package translate

import (
	"context"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/replay/doltcli"
	"github.com/steveyegge/beads/internal/replay/replaytest"
)

// These tests cover the commit-step plan: Plan reads one step (the net diff from
// one commit to another), classifies every table it changed with the table
// policy, and returns every issue's bd actions in the order they must run.
//
//   - Real-store tests, one per behaviour the ruling names (TestB3*): a
//     label-only step is a recorded gap, an unclassified table is a recorded gap
//     that fails the drift test, a delete cascades its edges, creates precede the
//     edges that name them, a title that looks like a flag replays, an issue the
//     translator cannot express runs nothing while the others still run, and the
//     translator never looks bd up on PATH. Each builds its oracle with bd and
//     replays the plan into a second bd project, as the harness does.
//   - Assembler tests (TestAssemble*): the plan is built by a function of the rows
//     the step changed, so ordering, atomicity and the edge rules are checked on
//     hand-built rows with no tool installed.

// ---- helpers --------------------------------------------------------------

// createIssue creates an issue in the bd project at dir and returns its id.
// extra holds any further flags, such as an explicit --id or --deps.
func createIssue(t *testing.T, dir, title string, extra ...string) string {
	t.Helper()
	args := append([]string{"create", title, "--type", "task", "--json"}, extra...)
	return jsonID(t, runBd(t, dir, args...))
}

// firstCommit returns the oldest commit of the dolt database at dir.
func firstCommit(t *testing.T, dir string) string {
	t.Helper()
	rows := parseCSV(t, runDolt(t, dir, "sql", "-q",
		"SELECT commit_hash FROM dolt_log ORDER BY date ASC LIMIT 1", "-r", "csv"))
	if len(rows) < 2 {
		t.Fatalf("dolt_log has no commits in %s", dir)
	}
	return rows[1][0]
}

func mustPlan(t *testing.T, dataDir, from, to string) *StepPlan {
	t.Helper()
	p, err := Plan(context.Background(), dataDir, from, to)
	if err != nil {
		t.Fatalf("Plan(%s, %s): %v", from, to, err)
	}
	return p
}

// applyPlan runs every action of p, in order, through the bd binary built from
// this tree.
func applyPlan(t *testing.T, workDir string, p *StepPlan) {
	t.Helper()
	applyActions(t, workDir, p.Actions)
}

// seedWork brings the work project to the oracle's current state by replaying
// the oracle's whole history as one step.
func seedWork(t *testing.T, fx *mutationFixture) {
	t.Helper()
	p := mustPlan(t, fx.source, firstCommit(t, fx.source), headCommit(t, fx.source))
	if len(p.Untranslatable) != 0 {
		t.Fatalf("seeding the work project: untranslatable issues %v", p.Untranslatable)
	}
	applyPlan(t, fx.workDir, p)
}

func kindsOf(actions []Action) []ActionKind {
	kinds := make([]ActionKind, 0, len(actions))
	for _, a := range actions {
		kinds = append(kinds, a.Kind)
	}
	return kinds
}

func argvLines(actions []Action) []string {
	lines := make([]string, 0, len(actions))
	for _, a := range actions {
		lines = append(lines, strings.Join(a.Argv, " "))
	}
	return lines
}

// commitTableChange creates a table in a fresh dolt database in one commit and
// fills it in the next, and returns the database and the commit before and
// after the insert: a step whose only change is rows in that table.
func commitTableChange(t *testing.T, table string) (dir, from, to string) {
	t.Helper()
	requireDolt(t)
	dir = replaytest.NewDoltDB(t, "oracle")
	runDolt(t, dir, "sql", "-q", "CREATE TABLE "+table+" (k INT PRIMARY KEY, v TEXT)")
	runDolt(t, dir, "add", "-A")
	runDolt(t, dir, "commit", "-m", "create "+table)
	from = headCommit(t, dir)
	runDolt(t, dir, "sql", "-q", "INSERT INTO "+table+" VALUES (1, 'x')")
	runDolt(t, dir, "add", "-A")
	runDolt(t, dir, "commit", "-m", "fill "+table)
	return dir, from, headCommit(t, dir)
}

// ---- the table policy -----------------------------------------------------

// B3.PolicyDrift: every table a freshly initialised store versions is
// classified, so a migration that adds a table fails here until someone decides
// what replay does with it. The set is what dolt_diff names over the store's
// history, which holds exactly the versioned tables: the local-only ones are
// dolt-ignored and never appear.
func TestB3PolicyDrift(t *testing.T) {
	requireBd(t)
	requireDolt(t)
	dir := initBdProject(t, "fresh")
	_, rows, err := doltcli.Query(context.Background(), dataDir(t, dir),
		"SELECT DISTINCT table_name FROM dolt_diff ORDER BY table_name")
	if err != nil {
		t.Fatalf("listing the tables a fresh store versions: %v", err)
	}
	if len(rows) == 0 {
		t.Fatal("a freshly initialised store versions no tables: the guard would pass over nothing")
	}
	seen := map[string]bool{}
	for _, r := range rows {
		name := r[0].Text
		seen[name] = true
		if _, ok := PolicyFor(name); !ok {
			t.Errorf("table %q is versioned by a fresh store and has no entry in the table policy: classify it as replayed, derived or unsupported, with a reason", name)
		}
	}
	for _, want := range []string{"issues", "dependencies"} {
		if !seen[want] {
			t.Errorf("a fresh store does not version %q, so this guard is not reading what it should", want)
		}
	}
}

// The policy as ruled: issues and dependencies are replayed, the tables bd
// rewrites on its own are derived, and everything else the migrations create is
// recorded as unsupported. Each entry says why.
func TestB3PolicyEntries(t *testing.T) {
	want := map[string]TableClass{
		"issues": TableReplayed, "dependencies": TableReplayed,

		"events": TableDerived, "child_counters": TableDerived, "issue_counter": TableDerived,
		// Tables the real histories carry that the ruling did not list.
		"schema_migrations": TableDerived, "dolt_ignore": TableDerived, "dolt_nonlocal_tables": TableDerived,
		"dolt_schemas": TableDerived, "dirty_issues": TableDerived, "export_hashes": TableDerived,

		"labels": TableUnsupported, "comments": TableUnsupported, "config": TableUnsupported,
		"metadata": TableUnsupported, "custom_statuses": TableUnsupported, "custom_types": TableUnsupported,
		"federation_peers": TableUnsupported, "interactions": TableUnsupported, "routes": TableUnsupported,
		"repo_mtimes": TableUnsupported, "compaction_snapshots": TableUnsupported, "issue_snapshots": TableUnsupported,
		"provenance_events": TableUnsupported, "store_epoch": TableUnsupported,
		"epoch_minted_addresses": TableUnsupported, "issue_versions": TableUnsupported,
	}
	names := PolicyTables()
	if len(names) != len(want) {
		t.Errorf("the policy classifies %d tables, want %d: %v", len(names), len(want), names)
	}
	for i := 1; i < len(names); i++ {
		if names[i-1] >= names[i] {
			t.Errorf("PolicyTables is not sorted and unique at %q, %q", names[i-1], names[i])
		}
	}
	for table, class := range want {
		got, ok := PolicyFor(table)
		if !ok {
			t.Errorf("table %q is not in the policy", table)
			continue
		}
		if got.Class != class {
			t.Errorf("table %q is %s, want %s", table, got.Class, class)
		}
		if len(strings.TrimSpace(got.Reason)) < 12 {
			t.Errorf("table %q has no real reason: %q", table, got.Reason)
		}
	}
	if _, ok := PolicyFor("zz_not_a_table"); ok {
		t.Error("PolicyFor classified a table nobody named")
	}
	for _, c := range []TableClass{TableReplayed, TableDerived, TableUnsupported} {
		if c.String() == "" || strings.Contains(c.String(), "TableClass(") {
			t.Errorf("class %d has no name", int(c))
		}
	}
}

// B3.UnknownTable: a step that changes a table the policy does not name is a
// recorded gap flagged unknown, never a silent no-op. The drift test is what
// fails the build.
func TestB3UnknownTable(t *testing.T) {
	dir, from, to := commitTableChange(t, "zz_unclassified")
	p := mustPlan(t, dir, from, to)
	if len(p.Actions) != 0 || len(p.Untranslatable) != 0 || len(p.Derived) != 0 {
		t.Errorf("a step over an unclassified table replays nothing and withholds nothing, got actions=%v untranslatable=%v derived=%v", p.Actions, p.Untranslatable, p.Derived)
	}
	if len(p.Gaps) != 1 || p.Gaps[0].Table != "zz_unclassified" || !p.Gaps[0].Unknown || p.Gaps[0].Reason == "" {
		t.Fatalf("gaps = %+v, want one unknown gap for zz_unclassified with a reason", p.Gaps)
	}
}

// A derived table is counted and not replayed; an unsupported one is a known gap
// with the policy's reason.
func TestB3DerivedAndUnsupportedTables(t *testing.T) {
	t.Run("derived", func(t *testing.T) {
		dir, from, to := commitTableChange(t, "child_counters")
		p := mustPlan(t, dir, from, to)
		if !reflect.DeepEqual(p.Derived, []string{"child_counters"}) {
			t.Errorf("derived = %v, want [child_counters]", p.Derived)
		}
		if len(p.Gaps) != 0 || len(p.Actions) != 0 {
			t.Errorf("a derived table is neither a gap nor an action, got gaps=%v actions=%v", p.Gaps, p.Actions)
		}
	})
	t.Run("unsupported", func(t *testing.T) {
		dir, from, to := commitTableChange(t, "config")
		p := mustPlan(t, dir, from, to)
		policy, _ := PolicyFor("config")
		if len(p.Gaps) != 1 || p.Gaps[0].Table != "config" || p.Gaps[0].Unknown || p.Gaps[0].Reason != policy.Reason {
			t.Errorf("gaps = %+v, want one known gap for config carrying the policy's reason", p.Gaps)
		}
		if len(p.Derived) != 0 || len(p.Actions) != 0 {
			t.Errorf("an unsupported table is neither derived nor an action, got derived=%v actions=%v", p.Derived, p.Actions)
		}
	})
}

// B3.LabelOnlyIsGap: a step whose only change is a label added through bd used to
// classify as a no-op and "match" a target that never received the label. It is
// now a recorded gap on the labels table.
func TestB3LabelOnlyIsGap(t *testing.T) {
	fx := newMutationFixture(t)
	id := createIssue(t, fx.sourceDir, "Labelled")
	from := headCommit(t, fx.source)
	runBd(t, fx.sourceDir, "label", "add", id, "triage")
	to := headCommit(t, fx.source)

	p := mustPlan(t, fx.source, from, to)
	if len(p.Actions) != 0 {
		t.Errorf("a label-only step replays nothing, got %v", argvLines(p.Actions))
	}
	if len(p.Gaps) != 1 || p.Gaps[0].Table != "labels" || p.Gaps[0].Unknown || p.Gaps[0].Reason == "" {
		t.Fatalf("gaps = %+v, want one known gap for labels with a reason", p.Gaps)
	}
	if len(p.Untranslatable) != 0 || len(p.Derived) != 0 {
		t.Errorf("a label-only step withholds no issue and counts no derived table, got untranslatable=%v derived=%v", p.Untranslatable, p.Derived)
	}
}

// ---- ordering and atomicity on real stores --------------------------------

// B3.DeleteWithDep: deleting an issue cascades its edges in the source, so the
// step carries a removed issue and a removed dependency. The plan is the delete
// alone: replaying a dependency remove after (or before) the cascade is what made
// this step fail half-applied.
func TestB3DeleteWithDep(t *testing.T) {
	fx := newMutationFixture(t)
	a := createIssue(t, fx.sourceDir, "Depends")
	b := createIssue(t, fx.sourceDir, "Target")
	runBd(t, fx.sourceDir, "dep", "add", a, b)
	seedWork(t, fx)
	if !depEdgeExists(t, fx.workDir, a, b) {
		t.Fatalf("seeding did not carry the edge %s -> %s", a, b)
	}

	from := headCommit(t, fx.source)
	runBd(t, fx.sourceDir, "delete", b, "--force")
	to := headCommit(t, fx.source)

	p := mustPlan(t, fx.source, from, to)
	if got, want := kindsOf(p.Actions), []ActionKind{KindDelete}; !reflect.DeepEqual(got, want) {
		t.Fatalf("plan = %v, want exactly one delete: the delete cascades the edge, so there is no dependency remove", argvLines(p.Actions))
	}
	applyPlan(t, fx.workDir, p)
	if issueExists(t, fx.workDir, b) {
		t.Errorf("%s still exists in the work project after the plan", b)
	}
	if !issueExists(t, fx.workDir, a) {
		t.Errorf("%s is gone from the work project: the plan deleted more than the step did", a)
	}
}

// B3.CreateWithDeps: one bd create that names a dependency is one commit with an
// issue row and an edge row. The plan creates the issue once, then adds the edge.
func TestB3CreateWithDeps(t *testing.T) {
	fx := newMutationFixture(t)
	parent := createIssue(t, fx.sourceDir, "Parent")
	seedWork(t, fx)

	from := headCommit(t, fx.source)
	child := createIssue(t, fx.sourceDir, "Child", "--deps", parent)
	to := headCommit(t, fx.source)

	p := mustPlan(t, fx.source, from, to)
	if got, want := kindsOf(p.Actions), []ActionKind{KindCreate, KindDepAdd}; !reflect.DeepEqual(got, want) {
		t.Fatalf("plan = %v, want one create then one dependency add", argvLines(p.Actions))
	}
	for _, a := range p.Actions {
		if a.Issue != child {
			t.Errorf("action %v belongs to %q, want %q: the child owns both its create and its edge", a.Argv, a.Issue, child)
		}
	}
	applyPlan(t, fx.workDir, p)
	if !issueExists(t, fx.workDir, child) || !depEdgeExists(t, fx.workDir, child, parent) {
		t.Fatalf("the work project lacks %s or its edge to %s after the plan", child, parent)
	}
}

// Creates come before every edge and are ordered by id, whatever order the step
// made them in: here the edge's owner sorts before its target, so replaying each
// issue's actions in turn would add the edge before the target exists.
func TestB3CreatesPrecedeEdgesAcrossIssues(t *testing.T) {
	fx := newMutationFixture(t)
	from := headCommit(t, fx.source)
	runBd(t, fx.sourceDir, "create", "Zed", "--id", "source-zzz", "--force", "--type", "task")
	runBd(t, fx.sourceDir, "create", "Aye", "--id", "source-aaa", "--force", "--type", "task", "--deps", "source-zzz")
	to := headCommit(t, fx.source)

	p := mustPlan(t, fx.source, from, to)
	if got, want := kindsOf(p.Actions), []ActionKind{KindCreate, KindCreate, KindDepAdd}; !reflect.DeepEqual(got, want) {
		t.Fatalf("plan = %v, want both creates, then the edge", argvLines(p.Actions))
	}
	if p.Actions[0].Issue != "source-aaa" || p.Actions[1].Issue != "source-zzz" {
		t.Errorf("creates are %q then %q, want them ordered by id", p.Actions[0].Issue, p.Actions[1].Issue)
	}
	applyPlan(t, fx.workDir, p)
	if !depEdgeExists(t, fx.workDir, "source-aaa", "source-zzz") {
		t.Error("the edge is missing from the work project after the plan")
	}
}

// B3.DashTitle: a title that begins with a dash is not a flag. bd refuses such a
// title as a positional even after a "--", so create gets it as the --title flag;
// the positionals bd accepts after "--" (update, close, delete, dependency
// commands) go there.
func TestB3DashTitle(t *testing.T) {
	fx := newMutationFixture(t)
	from := headCommit(t, fx.source)
	const created = "-leading dash title"
	id := jsonID(t, runBd(t, fx.sourceDir, "create", "--title", created, "--type", "task", "--json"))
	to := headCommit(t, fx.source)

	p := mustPlan(t, fx.source, from, to)
	if len(p.Actions) != 1 || p.Actions[0].Kind != KindCreate {
		t.Fatalf("plan = %v, want one create", argvLines(p.Actions))
	}
	argv := p.Actions[0].Argv
	if i := indexOf(argv, "--title"); i < 0 || i+1 >= len(argv) || argv[i+1] != created {
		t.Errorf("create argv = %q, want --title followed by the title as its own element", argv)
	}
	applyPlan(t, fx.workDir, p)
	if got := bdShowField(t, fx.workDir, id, "title"); got != created {
		t.Errorf("work title = %q, want %q", got, created)
	}

	from = headCommit(t, fx.source)
	const renamed = "--renamed like a flag"
	runBd(t, fx.sourceDir, "update", id, "--title", renamed)
	to = headCommit(t, fx.source)
	p = mustPlan(t, fx.source, from, to)
	if len(p.Actions) != 1 || p.Actions[0].Kind != KindUpdate {
		t.Fatalf("plan = %v, want one update", argvLines(p.Actions))
	}
	argv = p.Actions[0].Argv
	if n := len(argv); n < 2 || argv[n-2] != "--" || argv[n-1] != id {
		t.Errorf("update argv = %q, want the id after a \"--\" as the last two elements", argv)
	}
	applyPlan(t, fx.workDir, p)
	if got := bdShowField(t, fx.workDir, id, "title"); got != renamed {
		t.Errorf("work title after the update = %q, want %q", got, renamed)
	}
}

func indexOf(s []string, v string) int {
	for i, e := range s {
		if e == v {
			return i
		}
	}
	return -1
}

// B3.Atomic: an issue with a change the translator cannot express runs nothing
// at all, not even the parts it could express, and the other issues in the step
// still run. The withheld issue is typed, with the column named, and a single
// issue's view of the plan reports it as an error.
func TestB3Atomic(t *testing.T) {
	fx := newMutationFixture(t)
	x := createIssue(t, fx.sourceDir, "Unexpressible")
	y := createIssue(t, fx.sourceDir, "Fine")
	seedWork(t, fx)

	from := headCommit(t, fx.source)
	runBd(t, fx.sourceDir, "update", y, "--description", "y after")
	runBd(t, fx.sourceDir, "update", x, "--description", "x after")
	// owner has no bd update flag, so this column change has no replay form.
	runDolt(t, fx.source, "sql", "-q", "UPDATE issues SET owner = 'someone-else' WHERE id = '"+x+"'")
	runDolt(t, fx.source, "add", "-A")
	runDolt(t, fx.source, "commit", "-m", "test: direct owner mutation")
	to := headCommit(t, fx.source)

	p := mustPlan(t, fx.source, from, to)
	if len(p.Actions) != 1 || p.Actions[0].Kind != KindUpdate || p.Actions[0].Issue != y {
		t.Fatalf("plan = %+v, want exactly the update for %s and nothing for %s", p.Actions, y, x)
	}
	if len(p.Untranslatable) != 1 || p.Untranslatable[0].Issue != x {
		t.Fatalf("untranslatable = %v, want only %s", p.Untranslatable, x)
	}
	if got := p.Untranslatable[0].Columns; !reflect.DeepEqual(got, []string{"owner"}) {
		t.Errorf("untranslatable columns = %v, want [owner]", got)
	}
	if len(p.Untranslatable[0].Reasons) == 0 || !strings.Contains(p.Untranslatable[0].Error(), x) {
		t.Errorf("the withheld issue says nothing about why: %q", p.Untranslatable[0].Error())
	}

	applyPlan(t, fx.workDir, p)
	if got := bdShowField(t, fx.workDir, y, "description"); got != "y after" {
		t.Errorf("%s description = %q, want %q: the other issue must still run", y, got, "y after")
	}
	if got := bdShowField(t, fx.workDir, x, "description"); got == "x after" {
		t.Errorf("%s description changed to %q: an untranslatable issue must run nothing", x, got)
	}

	var u *Untranslatable
	if _, err := Classify(context.Background(), fx.source, from, to, x); !errors.As(err, &u) || u.Issue != x {
		t.Errorf("Classify(%s) = %v, want an *Untranslatable for it", x, err)
	}
	if actions, err := Classify(context.Background(), fx.source, from, to, y); err != nil || len(actions) != 1 || actions[0].Kind != KindUpdate {
		t.Errorf("Classify(%s) = %v, %v, want its one update", y, actions, err)
	}
}

// An edge to a reference outside the store replays as bd writes it: the target
// lives in depends_on_external, not depends_on_issue_id.
func TestB3ExternalDependency(t *testing.T) {
	fx := newMutationFixture(t)
	a := createIssue(t, fx.sourceDir, "Has an external dependency")
	seedWork(t, fx)

	from := headCommit(t, fx.source)
	runBd(t, fx.sourceDir, "dep", "add", a, "external:proj:cap")
	to := headCommit(t, fx.source)

	p := mustPlan(t, fx.source, from, to)
	if len(p.Actions) != 1 || p.Actions[0].Kind != KindDepAdd {
		t.Fatalf("plan = %v, want one dependency add", argvLines(p.Actions))
	}
	if got := p.Actions[0].Argv; got[len(got)-1] != "external:proj:cap" || got[len(got)-2] != a {
		t.Errorf("dependency argv = %q, want the owner and the external reference last", got)
	}
	applyPlan(t, fx.workDir, p)
	_, rows, err := doltcli.Query(context.Background(), dataDir(t, fx.workDir),
		"SELECT depends_on_external FROM dependencies WHERE issue_id = "+doltcli.SQLQuote(a))
	if err != nil {
		t.Fatalf("reading the work project's dependencies: %v", err)
	}
	if len(rows) != 1 || rows[0][0].Text != "external:proj:cap" {
		t.Errorf("work dependencies = %v, want one row for external:proj:cap", rows)
	}
}

// B3.PinnedBd: bd runs from the path the caller gives, whatever PATH holds. A
// decoy bd first on PATH is never started while the plan replays through the
// real one. (The source guard that nothing looks bd up on PATH is
// TestB3PinnedBdIsNeverLookedUp in the layout package.)
func TestB3PinnedBd(t *testing.T) {
	decoyDir, decoyLog := putBdFirst(t)
	t.Setenv("PATH", decoyDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	fx := newMutationFixture(t)
	from := headCommit(t, fx.source)
	id := createIssue(t, fx.sourceDir, "Replayed by the pinned bd")
	to := headCommit(t, fx.source)

	p := mustPlan(t, fx.source, from, to)
	applyPlan(t, fx.workDir, p)
	if !issueExists(t, fx.workDir, id) {
		t.Fatalf("%s is missing from the work project: the pinned bd did not run the plan", id)
	}
	if got := readCalls(t, decoyLog); got != "" {
		t.Errorf("the decoy bd on PATH was started: %q", got)
	}
}

// ---- the assembler on hand-built rows -------------------------------------

func row(kv ...string) diffRow {
	r := diffRow{}
	for i := 0; i+1 < len(kv); i += 2 {
		r[kv[i]] = kv[i+1]
	}
	return r
}

func issuesAndDeps() []tableChange {
	return []tableChange{{Name: "issues", DiffType: "modified"}, {Name: "dependencies", DiffType: "modified"}}
}

// A step's actions run in one fixed order, with ids sorted inside each phase:
// creates, modifications, dependency removes, dependency adds, deletes. An edge
// whose owner or target is deleted in the step is dropped, because the delete
// cascades it. An edge is owned by its issue_id end.
func TestAssembleOrdering(t *testing.T) {
	p := assemble(stepInput{
		From: "f", To: "t", Tables: issuesAndDeps(),
		Issues: []diffRow{
			row("diff_type", "removed", "from_id", "d-4"),
			row("diff_type", "added", "to_id", "b-2", "to_title", "Bee"),
			row("diff_type", "modified", "to_id", "c-3", "from_id", "c-3", "to_description", "new", "from_description", "old"),
			row("diff_type", "added", "to_id", "a-1", "to_title", "Aye"),
		},
		Deps: []diffRow{
			row("diff_type", "added", "to_issue_id", "c-3", "to_depends_on_issue_id", "a-1", "to_type", "blocks"),
			row("diff_type", "added", "to_issue_id", "a-1", "to_depends_on_issue_id", "b-2", "to_type", "related"),
			row("diff_type", "removed", "from_issue_id", "c-3", "from_depends_on_issue_id", "e-5"),
			row("diff_type", "removed", "from_issue_id", "d-4", "from_depends_on_issue_id", "e-5"),
			row("diff_type", "removed", "from_issue_id", "e-5", "from_depends_on_issue_id", "d-4"),
		},
	})
	want := []string{
		"create --id a-1 --force --title Aye",
		"create --id b-2 --force --title Bee",
		"update --description new -- c-3",
		"dep remove -- c-3 e-5",
		"dep add --type related -- a-1 b-2",
		"dep add --type blocks -- c-3 a-1",
		"delete --force -- d-4",
	}
	if got := argvLines(p.Actions); !reflect.DeepEqual(got, want) {
		t.Errorf("actions =\n  %s\nwant\n  %s", strings.Join(got, "\n  "), strings.Join(want, "\n  "))
	}
	if len(p.Untranslatable) != 0 || len(p.Gaps) != 0 {
		t.Errorf("nothing here is withheld or a gap, got untranslatable=%v gaps=%v", p.Untranslatable, p.Gaps)
	}
	for _, a := range p.Actions {
		if a.Issue == "" {
			t.Errorf("action %v names no issue", a.Argv)
		}
	}
	if got := argvLines(p.ActionsFor("a-1")); !reflect.DeepEqual(got, []string{"create --id a-1 --force --title Aye", "dep add --type related -- a-1 b-2"}) {
		t.Errorf("ActionsFor(a-1) = %v: it owns its create and the edge it declares, not the edge pointing at it", got)
	}
}

// Everything a created issue is built with goes in as a flag, in a fixed order,
// and an empty value is left out.
func TestAssembleCreateFlags(t *testing.T) {
	p := assemble(stepInput{
		From: "f", To: "t", Tables: issuesAndDeps()[:1],
		Issues: []diffRow{row("diff_type", "added", "to_id", "x-1", "to_title", "Widget, deluxe", "to_description", "Does things",
			"to_priority", "2", "to_issue_type", "task", "to_status", "open", "to_notes", "", "to_assignee", "")},
	})
	want := []string{"create --id x-1 --force --title Widget, deluxe --description Does things --priority 2 --type task --status open"}
	if got := argvLines(p.Actions); !reflect.DeepEqual(got, want) {
		t.Errorf("actions = %v, want %v", got, want)
	}
}

// A target the work project will not hold makes the issue that names it
// untranslatable at plan time: nothing for it runs, the rest is unaffected.
func TestAssembleMissingTargetIsUntranslatable(t *testing.T) {
	p := assemble(stepInput{
		From: "f", To: "t", Tables: issuesAndDeps(),
		Issues: []diffRow{
			row("diff_type", "modified", "to_id", "x-1", "from_id", "x-1", "to_description", "new", "from_description", "old"),
		},
		Deps: []diffRow{
			row("diff_type", "added", "to_issue_id", "x-1", "to_depends_on_issue_id", "y-2", "to_type", "blocks"),
			row("diff_type", "added", "to_issue_id", "z-3", "to_depends_on_issue_id", "x-1", "to_type", "blocks"),
		},
		Present: map[string]bool{"x-1": true},
	})
	if got, want := argvLines(p.Actions), []string{"dep add --type blocks -- z-3 x-1"}; !reflect.DeepEqual(got, want) {
		t.Errorf("actions = %v, want %v", got, want)
	}
	if len(p.Untranslatable) != 1 || p.Untranslatable[0].Issue != "x-1" || !strings.Contains(p.Untranslatable[0].Error(), "y-2") {
		t.Errorf("untranslatable = %v, want x-1, naming the missing target y-2", p.Untranslatable)
	}
}

// Withholding a created issue makes it a missing target for the edges that name
// it, so the issues that own those edges are withheld too, however deep the chain.
func TestAssembleWithholdingCascades(t *testing.T) {
	p := assemble(stepInput{
		From: "f", To: "t", Tables: issuesAndDeps(),
		Issues: []diffRow{
			row("diff_type", "added", "to_id", "a-1", "to_title", "Fine"),
			row("diff_type", "added", "to_id", "c-3", "to_title", "Points at nothing"),
			row("diff_type", "added", "to_id", "d-4", "to_title", "Points at c-3"),
			row("diff_type", "added", "to_id", "e-5", "to_title", "Points at d-4"),
		},
		Deps: []diffRow{
			row("diff_type", "added", "to_issue_id", "c-3", "to_depends_on_issue_id", "m-9", "to_type", "blocks"),
			row("diff_type", "added", "to_issue_id", "d-4", "to_depends_on_issue_id", "c-3", "to_type", "blocks"),
			row("diff_type", "added", "to_issue_id", "e-5", "to_depends_on_issue_id", "d-4", "to_type", "blocks"),
		},
	})
	if got, want := argvLines(p.Actions), []string{"create --id a-1 --force --title Fine"}; !reflect.DeepEqual(got, want) {
		t.Errorf("actions = %v, want %v", got, want)
	}
	var withheld []string
	for _, u := range p.Untranslatable {
		withheld = append(withheld, u.Issue)
	}
	if !reflect.DeepEqual(withheld, []string{"c-3", "d-4", "e-5"}) {
		t.Errorf("withheld = %v, want [c-3 d-4 e-5] in id order", withheld)
	}
}

// A target outside the store is written to the external column; a target that is
// a wisp, or no target at all, cannot be replayed.
func TestAssembleEdgeTargets(t *testing.T) {
	p := assemble(stepInput{
		From: "f", To: "t", Tables: issuesAndDeps()[1:],
		Deps: []diffRow{
			row("diff_type", "added", "to_issue_id", "x-1", "to_depends_on_issue_id", "", "to_depends_on_external", "external:proj:cap", "to_type", "blocks"),
			row("diff_type", "added", "to_issue_id", "y-2", "to_depends_on_issue_id", "", "to_depends_on_wisp_id", "w-7", "to_type", "blocks"),
			row("diff_type", "added", "to_issue_id", "z-3", "to_depends_on_issue_id", "", "to_depends_on_external", "", "to_type", "blocks"),
			row("diff_type", "removed", "from_issue_id", "x-1", "from_depends_on_issue_id", "", "from_depends_on_external", "external:proj:old"),
		},
	})
	want := []string{"dep remove -- x-1 external:proj:old", "dep add --type blocks -- x-1 external:proj:cap"}
	if got := argvLines(p.Actions); !reflect.DeepEqual(got, want) {
		t.Errorf("actions = %v, want %v", got, want)
	}
	if len(p.Untranslatable) != 2 || p.Untranslatable[0].Issue != "y-2" || p.Untranslatable[1].Issue != "z-3" {
		t.Fatalf("untranslatable = %v, want y-2 (wisp target) and z-3 (no target)", p.Untranslatable)
	}
	if !strings.Contains(p.Untranslatable[0].Error(), "wisp") {
		t.Errorf("the wisp edge's reason does not say wisp: %q", p.Untranslatable[0].Error())
	}
}

// An edge changed in place (its type, say) has no single bd form.
func TestAssembleModifiedEdgeIsUntranslatable(t *testing.T) {
	p := assemble(stepInput{
		From: "f", To: "t", Tables: issuesAndDeps()[1:],
		Deps: []diffRow{
			row("diff_type", "modified", "to_issue_id", "x-1", "to_depends_on_issue_id", "y-2", "to_type", "related",
				"from_issue_id", "x-1", "from_depends_on_issue_id", "y-2", "from_type", "blocks"),
		},
		Present: map[string]bool{"y-2": true},
	})
	if len(p.Actions) != 0 || len(p.Untranslatable) != 1 || p.Untranslatable[0].Issue != "x-1" {
		t.Errorf("actions=%v untranslatable=%v, want x-1 withheld and nothing run", p.Actions, p.Untranslatable)
	}
}

// What a modified row changes decides what runs: settable columns become one
// update, the bookkeeping columns bd rewrites itself change nothing, a close is
// a close, and a column with no bd form withholds the issue.
func TestAssembleModifiedRows(t *testing.T) {
	derived := []string{"to_updated_at", "2026-02-01", "from_updated_at", "2026-01-01", "to_content_hash", "b", "from_content_hash", "a",
		"to_is_blocked", "0", "from_is_blocked", "1", "to_row_lock", "2", "from_row_lock", "1"}
	p := assemble(stepInput{
		From: "f", To: "t", Tables: issuesAndDeps()[:1],
		Issues: []diffRow{
			row(append([]string{"diff_type", "modified", "to_id", "a-1", "from_id", "a-1"}, derived...)...),
			row("diff_type", "modified", "to_id", "b-2", "from_id", "b-2", "to_title", "New", "from_title", "Old", "to_priority", "1", "from_priority", "3"),
			row("diff_type", "modified", "to_id", "c-3", "from_id", "c-3", "to_status", "closed", "from_status", "open",
				"to_close_reason", "done", "from_close_reason", "", "to_closed_at", "2026-01-01", "from_closed_at", ""),
			row("diff_type", "modified", "to_id", "d-4", "from_id", "d-4", "to_description", "x", "from_description", "y", "to_owner", "someone", "from_owner", "nobody"),
		},
	})
	want := []string{"update --title New --priority 1 -- b-2", "close --reason done -- c-3"}
	if got := argvLines(p.Actions); !reflect.DeepEqual(got, want) {
		t.Errorf("actions = %v, want %v", got, want)
	}
	if len(p.Untranslatable) != 1 || p.Untranslatable[0].Issue != "d-4" || !reflect.DeepEqual(p.Untranslatable[0].Columns, []string{"owner"}) {
		t.Errorf("untranslatable = %v, want d-4 withheld for the column owner", p.Untranslatable)
	}
}

// The tables a step changed are classified once each: unsupported ones and
// unclassified ones are gaps (the unclassified flagged), derived ones are listed.
func TestAssembleTableClasses(t *testing.T) {
	p := assemble(stepInput{
		From: "f", To: "t",
		Tables: []tableChange{
			{Name: "labels", DiffType: "modified"}, {Name: "zz_unclassified", DiffType: "added"}, {Name: "events", DiffType: "modified"},
			{Name: "comments", DiffType: "modified"}, {Name: "labels", DiffType: "modified"}, {Name: "issues", DiffType: "modified"},
		},
	})
	var gaps []string
	for _, g := range p.Gaps {
		gaps = append(gaps, g.Table)
		if g.Unknown != (g.Table == "zz_unclassified") || g.Reason == "" {
			t.Errorf("gap %+v: Unknown must be set exactly for the unclassified table, and every gap needs a reason", g)
		}
	}
	if !reflect.DeepEqual(gaps, []string{"comments", "labels", "zz_unclassified"}) {
		t.Errorf("gaps = %v, want [comments labels zz_unclassified] sorted, labels once", gaps)
	}
	if !reflect.DeepEqual(p.Derived, []string{"events"}) {
		t.Errorf("derived = %v, want [events]", p.Derived)
	}
	if len(p.Actions) != 0 {
		t.Errorf("no row changed, so nothing runs; got %v", p.Actions)
	}
}

// Only the targets the plan cannot vouch for need a look at the oracle: not the
// issues the step creates, not references outside the store, not wisps.
func TestPresenceCandidates(t *testing.T) {
	issues := []diffRow{row("diff_type", "added", "to_id", "made-1")}
	deps := []diffRow{
		row("diff_type", "added", "to_issue_id", "a-1", "to_depends_on_issue_id", "made-1"),
		row("diff_type", "added", "to_issue_id", "a-1", "to_depends_on_issue_id", "old-2"),
		row("diff_type", "added", "to_issue_id", "b-2", "to_depends_on_issue_id", "old-2"),
		row("diff_type", "added", "to_issue_id", "b-2", "to_depends_on_issue_id", "", "to_depends_on_external", "external:p:c"),
		row("diff_type", "added", "to_issue_id", "b-2", "to_depends_on_issue_id", "", "to_depends_on_wisp_id", "w-1"),
		row("diff_type", "removed", "from_issue_id", "c-3", "from_depends_on_issue_id", "gone-3"),
		row("diff_type", "added", "to_issue_id", "c-3", "to_depends_on_issue_id", "another-4"),
	}
	if got, want := presenceCandidates(issues, deps), []string{"another-4", "old-2"}; !reflect.DeepEqual(got, want) {
		t.Errorf("presenceCandidates = %v, want %v", got, want)
	}
}
