package translate

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/steveyegge/beads/internal/replay/doltcli"
	"github.com/steveyegge/beads/internal/storage/issueops"
)

// StepPlan is everything the translator derives from one commit step, the net
// change from one commit to another: the bd actions that replay it, in the order
// they must run, and what the step changed that it cannot replay. Nothing in it
// has run yet.
type StepPlan struct {
	From, To string
	// Actions replay the step, for every issue, in execution order.
	Actions []Action
	// Untranslatable lists the issues the step changed in a way the translator
	// cannot express. Nothing for such an issue is in Actions. Sorted by issue id.
	Untranslatable []*Untranslatable
	// Gaps lists the changed tables the translator does not replay: the
	// unsupported ones, and any table the policy does not name. Sorted by table.
	Gaps []CoverageGap
	// Derived names the changed tables bd rewrites on its own: counted, not
	// replayed. Sorted.
	Derived []string
}

// ActionsFor returns the actions that belong to issue, in execution order. An
// edge belongs to its issue_id end.
func (p *StepPlan) ActionsFor(issue string) []Action {
	var out []Action
	for _, a := range p.Actions {
		if a.Issue == issue {
			out = append(out, a)
		}
	}
	return out
}

// Untranslatable says why one issue's changes in a step cannot be replayed. A
// plan is atomic per issue: if any part of an issue's step cannot be expressed,
// none of its actions run, and the others in the step are unaffected.
type Untranslatable struct {
	Issue string
	// Columns are the issues-table columns that changed and have no bd form,
	// sorted. Empty when the cause is not a column.
	Columns []string
	Reasons []string
}

func (u *Untranslatable) Error() string {
	return fmt.Sprintf("%s is untranslatable: %s", u.Issue, strings.Join(u.Reasons, "; "))
}

// CoverageGap records that a step changed a table the translator does not
// replay, so a replay that compared only what it replayed would pass over it.
type CoverageGap struct {
	Table string
	// Unknown is set when the table is not in the table policy at all.
	Unknown bool
	Reason  string
}

// diffRow is one row of a dolt_commit_diff_* table read as text: columns named
// from_* and to_*, plus diff_type. A NULL and an empty string both read as "".
type diffRow map[string]string

// tableChange is one table the step changed, as dolt_diff_summary reports it.
type tableChange struct {
	Name     string
	DiffType string
}

// stepInput is what Plan reads from the oracle for one step; assemble turns it
// into a plan and does no I/O.
type stepInput struct {
	From, To string
	Tables   []tableChange
	// Issues and Deps are the row diffs of the replayed tables, empty when the
	// step did not change them.
	Issues []diffRow
	Deps   []diffRow
	// Present holds the edge targets, among presenceCandidates(Issues, Deps), that
	// exist in the oracle before the step: the work project holds them too.
	Present map[string]bool
	// Unsafe holds the candidates idsPresent refused to look up. Present says
	// nothing about them either way, so the issues that point at them are withheld.
	Unsafe map[string]bool
}

// Plan reads the commit step from `from` (exclusive) to `to` (inclusive) in the
// dolt data directory dataDir, classifies every table it changed with the table
// policy, and returns one plan for the whole step. The step is net: a range of
// several commits is planned as the single change from its first state to its
// last.
func Plan(ctx context.Context, dataDir, from, to string) (*StepPlan, error) {
	for _, ref := range []string{from, to} {
		if err := issueops.ValidateRef(ref); err != nil {
			return nil, fmt.Errorf("plan: %w", err)
		}
	}
	tables, err := changedTables(ctx, dataDir, from, to)
	if err != nil {
		return nil, err
	}
	in := stepInput{From: from, To: to, Tables: tables}
	issuesAdded, issuesChanged, depsChanged := false, false, false
	for _, tc := range tables {
		switch tc.Name {
		case "issues":
			issuesChanged = true
			issuesAdded = tc.DiffType == "added"
		case "dependencies":
			depsChanged = true
		}
	}
	// Issues first, then dependencies, whatever order dolt listed the tables in.
	if issuesChanged {
		if in.Issues, err = readIssueDiffs(ctx, dataDir, from, to); err != nil {
			return nil, err
		}
	}
	if depsChanged {
		if in.Deps, err = readDependencyDiffs(ctx, dataDir, from, to); err != nil {
			return nil, err
		}
	}
	// The issues table did not exist before a step that adds it, so nothing is
	// present and there is nothing to ask.
	if candidates := presenceCandidates(in.Issues, in.Deps); len(candidates) > 0 && !issuesAdded {
		var refused []string
		if in.Present, refused, err = idsPresent(ctx, dataDir, from, candidates); err != nil {
			return nil, err
		}
		in.Unsafe = make(map[string]bool, len(refused))
		for _, id := range refused {
			in.Unsafe[id] = true
		}
	}
	return assemble(in), nil
}

// changedTables lists the tables dolt says differ between from and to. For a
// table that exists on only one side, dolt reports the missing side as an empty
// string, so each name is taken from whichever side has one.
func changedTables(ctx context.Context, dataDir, from, to string) ([]tableChange, error) {
	query := fmt.Sprintf("SELECT from_table_name, to_table_name, diff_type FROM dolt_diff_summary(%s, %s)",
		doltcli.SQLQuote(from), doltcli.SQLQuote(to))
	_, rows, err := doltcli.Query(ctx, dataDir, query)
	if err != nil {
		return nil, fmt.Errorf("plan: listing the tables changed between %s and %s: %w", from, to, err)
	}
	var out []tableChange
	for _, r := range rows {
		fromName, toName, diffType := r[0].Text, r[1].Text, r[2].Text
		if toName != "" {
			out = append(out, tableChange{Name: toName, DiffType: diffType})
		}
		if fromName != "" && fromName != toName {
			out = append(out, tableChange{Name: fromName, DiffType: diffType})
		}
	}
	return out, nil
}

func readIssueDiffs(ctx context.Context, dataDir, from, to string) ([]diffRow, error) {
	query := fmt.Sprintf("SELECT * FROM dolt_commit_diff_issues WHERE to_commit=%s AND from_commit=%s",
		doltcli.SQLQuote(to), doltcli.SQLQuote(from))
	rows, err := readDiffRows(ctx, dataDir, query)
	if err != nil {
		return nil, fmt.Errorf("plan: reading the issue rows changed between %s and %s: %w", from, to, err)
	}
	return rows, nil
}

func readDependencyDiffs(ctx context.Context, dataDir, from, to string) ([]diffRow, error) {
	query := fmt.Sprintf("SELECT * FROM dolt_commit_diff_dependencies WHERE to_commit=%s AND from_commit=%s",
		doltcli.SQLQuote(to), doltcli.SQLQuote(from))
	rows, err := readDiffRows(ctx, dataDir, query)
	if err != nil {
		return nil, fmt.Errorf("plan: reading the dependency rows changed between %s and %s: %w", from, to, err)
	}
	return rows, nil
}

func readDiffRows(ctx context.Context, dataDir, query string) ([]diffRow, error) {
	header, rows, err := doltcli.Query(ctx, dataDir, query)
	if err != nil {
		return nil, err
	}
	out := make([]diffRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, diffRow(doltcli.RowMap(header, r)))
	}
	return out, nil
}

// idsPresentBatch bounds the length of one presence query.
const idsPresentBatch = 400

// idsPresent returns which of ids exist in the oracle's issues table as of ref, and
// the ids it refused to look up. The ids come out of the oracle's own rows, so each
// is data, not a value the harness chose. Each is checked here, where it is put
// into a query, with issueops.ValidateRef, the same check every other id and ref
// the harness queries by passes, and one that fails is never sent to dolt. A refused
// id is not present; the caller says why.
func idsPresent(ctx context.Context, dataDir, ref string, ids []string) (map[string]bool, []string, error) {
	present := make(map[string]bool, len(ids))
	var refused []string
	lookable := make([]string, 0, len(ids))
	for _, id := range ids {
		if issueops.ValidateRef(id) != nil {
			refused = append(refused, id)
			continue
		}
		lookable = append(lookable, id)
	}
	for start := 0; start < len(lookable); start += idsPresentBatch {
		end := min(start+idsPresentBatch, len(lookable))
		quoted := make([]string, 0, end-start)
		for _, id := range lookable[start:end] {
			quoted = append(quoted, doltcli.SQLQuote(id))
		}
		query := fmt.Sprintf("SELECT id FROM issues AS OF %s WHERE id IN (%s)", doltcli.SQLQuote(ref), strings.Join(quoted, ", "))
		_, rows, err := doltcli.Query(ctx, dataDir, query)
		if err != nil {
			return nil, nil, fmt.Errorf("plan: checking which edge targets exist as of %s: %w", ref, err)
		}
		for _, r := range rows {
			present[r[0].Text] = true
		}
	}
	return present, refused, nil
}

// edgeTarget names the target of a dependency row, reading the columns under
// prefix ("to_" or "from_"). It reports external when the target is a reference
// outside the store, which bd keeps in its own column, and a reason when the row
// names nothing a bd command can address.
func edgeTarget(r diffRow, prefix string) (target string, external bool, reason string) {
	if v := r[prefix+"depends_on_issue_id"]; v != "" {
		return v, false, ""
	}
	if v := r[prefix+"depends_on_external"]; v != "" {
		return v, true, ""
	}
	if v := r[prefix+"depends_on_wisp_id"]; v != "" {
		return "", false, fmt.Sprintf("dependency target %s is a wisp, which the work project does not hold", v)
	}
	return "", false, "dependency row names no target"
}

// presenceCandidates returns the sorted ids of the issues that edges added by the
// step point at and that the step itself does not create: the only ones whose
// existence before the step the plan cannot know from the step alone.
func presenceCandidates(issues, deps []diffRow) []string {
	created := map[string]bool{}
	for _, r := range issues {
		if r["diff_type"] == "added" {
			created[r["to_id"]] = true
		}
	}
	set := map[string]bool{}
	for _, r := range deps {
		if r["diff_type"] != "added" {
			continue
		}
		target, external, reason := edgeTarget(r, "to_")
		if reason != "" || external || created[target] {
			continue
		}
		set[target] = true
	}
	return sortedKeys(set)
}

func sortedKeys(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// edge is one dependency the step adds or removes, owned by its issue_id end.
type edge struct {
	owner, target, typ string
	external           bool
}

// issueWork is everything the step does to one issue before atomicity is
// applied.
type issueWork struct {
	create  *Action
	mods    []Action
	del     *Action
	reasons []string
	columns map[string]bool
}

func (w *issueWork) block(reason string, columns ...string) {
	w.reasons = append(w.reasons, reason)
	for _, c := range columns {
		if w.columns == nil {
			w.columns = map[string]bool{}
		}
		w.columns[c] = true
	}
}

func (w *issueWork) blocked() bool { return len(w.reasons) > 0 }

// assemble builds the plan for one step from the rows it changed.
//
// Atomic per issue: the whole plan for an issue is built before anything runs,
// and an issue with any part the translator cannot express is withheld
// entirely, including the edges it owns. Withholding a created issue makes it a
// missing target for the edges that name it, so the loop below runs until no
// further issue is withheld.
//
// Order is fixed so a replay is deterministic and nothing is asked of bd before
// it can be done: creates, then each issue's modifications, then dependency
// removes, then dependency adds, then deletes last, each phase by issue id. A
// dependency remove is dropped when either end is deleted in the step, because
// the delete cascades the edge.
func assemble(in stepInput) *StepPlan {
	p := &StepPlan{From: in.From, To: in.To}
	classifyTables(p, in.Tables)

	issues := map[string]*issueWork{}
	work := func(id string) *issueWork {
		w := issues[id]
		if w == nil {
			w = &issueWork{}
			issues[id] = w
		}
		return w
	}
	created := map[string]bool{}
	deleted := map[string]bool{}

	for _, r := range in.Issues {
		switch r["diff_type"] {
		case "added":
			id := r["to_id"]
			work(id).create = &Action{Kind: KindCreate, Issue: id, Argv: createArgv(r, id)}
			created[id] = true
		case "removed":
			id := r["from_id"]
			work(id).del = &Action{Kind: KindDelete, Issue: id, Argv: []string{"delete", "--force", "--", id}}
			deleted[id] = true
		case "modified":
			id := r["to_id"]
			actions, unmapped := modifiedActions(r, id)
			if len(unmapped) > 0 {
				work(id).block(fmt.Sprintf("issues column(s) changed that have no bd form: %s", strings.Join(unmapped, ", ")), unmapped...)
				continue
			}
			work(id).mods = actions
		default:
			id := firstNonEmpty(r["to_id"], r["from_id"])
			work(id).block(fmt.Sprintf("issues row has an unrecognized diff_type %q", r["diff_type"]))
		}
	}

	var removes, adds []edge
	for _, r := range in.Deps {
		switch r["diff_type"] {
		case "added":
			owner := r["to_issue_id"]
			target, external, reason := edgeTarget(r, "to_")
			if reason != "" {
				work(owner).block(reason)
				continue
			}
			adds = append(adds, edge{owner: owner, target: target, typ: r["to_type"], external: external})
		case "removed":
			owner := r["from_issue_id"]
			if deleted[owner] {
				continue
			}
			target, external, reason := edgeTarget(r, "from_")
			if reason != "" {
				work(owner).block(reason)
				continue
			}
			if !external && deleted[target] {
				continue
			}
			removes = append(removes, edge{owner: owner, target: target, external: external})
		case "modified":
			owner := firstNonEmpty(r["to_issue_id"], r["from_issue_id"])
			work(owner).block(fmt.Sprintf("dependency %s -> %s was changed in place, which has no single bd form", owner, firstNonEmpty(r["to_depends_on_issue_id"], r["from_depends_on_issue_id"])))
		default:
			owner := firstNonEmpty(r["to_issue_id"], r["from_issue_id"])
			work(owner).block(fmt.Sprintf("dependency row has an unrecognized diff_type %q", r["diff_type"]))
		}
	}

	// An edge needs its target to exist when it runs: in the work project, or
	// created earlier in this plan by an issue that is not itself withheld.
	exists := func(id string) bool {
		if w := issues[id]; created[id] && w != nil && !w.blocked() {
			return true
		}
		return in.Present[id]
	}
	for changed := true; changed; {
		changed = false
		for _, e := range adds {
			if w := issues[e.owner]; w != nil && w.blocked() {
				continue
			}
			if e.external || exists(e.target) {
				continue
			}
			reason := fmt.Sprintf("dependency %s -> %s: the target is neither in the work project nor created earlier in this plan", e.owner, e.target)
			if in.Unsafe[e.target] {
				reason = fmt.Sprintf("dependency %s -> %q: the target id cannot be looked up (an id holds only letters, digits and _ . / -, up to 128 of them), so the plan cannot tell whether the work project has it", e.owner, e.target)
			}
			work(e.owner).block(reason)
			changed = true
		}
	}

	live := func(id string) bool { w := issues[id]; return w == nil || !w.blocked() }
	ids := make([]string, 0, len(issues))
	for id := range issues {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		if w := issues[id]; w.blocked() {
			p.Untranslatable = append(p.Untranslatable, &Untranslatable{Issue: id, Columns: sortedKeys(w.columns), Reasons: w.reasons})
		}
	}
	for _, id := range ids {
		if w := issues[id]; live(id) && w.create != nil {
			p.Actions = append(p.Actions, *w.create)
		}
	}
	for _, id := range ids {
		if live(id) {
			p.Actions = append(p.Actions, issues[id].mods...)
		}
	}
	sortEdges(removes)
	for _, e := range removes {
		if live(e.owner) {
			p.Actions = append(p.Actions, Action{Kind: KindDepRemove, Issue: e.owner, Argv: []string{"dep", "remove", "--", e.owner, e.target}})
		}
	}
	sortEdges(adds)
	for _, e := range adds {
		if live(e.owner) {
			argv := []string{"dep", "add"}
			if e.typ != "" {
				argv = append(argv, "--type", e.typ)
			}
			p.Actions = append(p.Actions, Action{Kind: KindDepAdd, Issue: e.owner, Argv: append(argv, "--", e.owner, e.target)})
		}
	}
	for _, id := range ids {
		if w := issues[id]; live(id) && w.del != nil {
			p.Actions = append(p.Actions, *w.del)
		}
	}
	return p
}

func sortEdges(edges []edge) {
	sort.Slice(edges, func(i, j int) bool {
		a, b := edges[i], edges[j]
		if a.owner != b.owner {
			return a.owner < b.owner
		}
		if a.target != b.target {
			return a.target < b.target
		}
		return a.typ < b.typ
	})
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// classifyTables sorts the tables a step changed by the policy: unsupported and
// unnamed ones become gaps, derived ones are listed, replayed ones are read as
// row diffs by Plan.
func classifyTables(p *StepPlan, tables []tableChange) {
	seen := map[string]bool{}
	for _, tc := range tables {
		if seen[tc.Name] {
			continue
		}
		seen[tc.Name] = true
		policy, known := PolicyFor(tc.Name)
		switch {
		case !known:
			p.Gaps = append(p.Gaps, CoverageGap{Table: tc.Name, Unknown: true, Reason: "the table is not in the table policy"})
		case policy.Class == TableUnsupported:
			p.Gaps = append(p.Gaps, CoverageGap{Table: tc.Name, Reason: policy.Reason})
		case policy.Class == TableDerived:
			p.Derived = append(p.Derived, tc.Name)
		}
	}
	sort.Slice(p.Gaps, func(i, j int) bool { return p.Gaps[i].Table < p.Gaps[j].Table })
	sort.Strings(p.Derived)
}

// createArgv builds the bd create invocation for an added issue row: every
// settable column the row holds goes in as a flag, the title too. bd refuses a
// title that begins with a dash when it is a positional, even after a "--", so
// nothing about create is positional.
func createArgv(row diffRow, id string) []string {
	// --force: a work project's bd usually has a different ID prefix than the
	// source it replays, which create otherwise refuses as a prefix mismatch.
	argv := []string{"create", "--id", id, "--force"}
	for _, ff := range settableFields {
		if v := row["to_"+ff.column]; v != "" {
			argv = append(argv, ff.flag, v)
		}
	}
	return argv
}

// modifiedActions turns a modified issue row into its update and close actions,
// or returns the changed columns that have no bd form. The positional id goes
// after a "--" so an id that looks like a flag is still an id.
func modifiedActions(row diffRow, id string) (actions []Action, unmapped []string) {
	changed := make(map[string]bool)
	for k, v := range row {
		if !strings.HasPrefix(k, "to_") {
			continue
		}
		base := k[len("to_"):]
		fromVal, ok := row["from_"+base]
		if !ok || ignoredColumns[base] {
			continue
		}
		if v != fromVal {
			changed[base] = true
		}
	}

	closing := row["from_status"] != "closed" && row["to_status"] == "closed"
	if closing {
		for c := range closingColumns {
			delete(changed, c)
		}
	}

	var flags []string
	for _, ff := range settableFields {
		if !changed[ff.column] {
			continue
		}
		flags = append(flags, ff.flag, row["to_"+ff.column])
		delete(changed, ff.column)
	}
	if len(changed) > 0 {
		return nil, sortedKeys(changed)
	}
	if len(flags) > 0 {
		argv := append(append([]string{"update"}, flags...), "--", id)
		actions = append(actions, Action{Kind: KindUpdate, Issue: id, Argv: argv})
	}
	if closing {
		argv := []string{"close"}
		if reason := row["to_close_reason"]; reason != "" {
			argv = append(argv, "--reason", reason)
		}
		actions = append(actions, Action{Kind: KindClose, Issue: id, Argv: append(argv, "--", id)})
	}
	return actions, nil
}
