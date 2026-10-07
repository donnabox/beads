package scripts_test

import (
	"fmt"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// This file is the fork's own (upstream has no copy, so syncs never conflict
// on it). It pins rule R6: no workflow job a pull request in this repository
// can schedule may ask for a Blacksmith runner, because this repository has no
// pool and such a job queues forever.
//
// Upstream puts blacksmith-<N>vcpu-ubuntu-2404 labels on `runs-on` in two
// ways. (1) Behind needs.rbe.outputs.mode == 'remote' (bazel.yml's lanes):
// that needs the repository variable RBE_WEST_WORKERS and the secret
// RBE_WEST_EXECUTOR, which this repository does not have, and rbe's fork modes
// (fork-ro, fork-rw) stay GitHub-hosted. (2) Behind a same-repository
// pull_request / merge_group predicate (the sameRepoBlacksmith* constants):
// every pull request inside this repository is same-repository, so it is true
// for all of them. The fork's edit is one conjunct in front of (2),
// github.repository_owner == 'gastownhall', so anywhere else the expression is
// ubuntu-latest, and a missing context field fails towards the hosted runner.
//
// This file does not match the expressions' text. It evaluates each one with
// upstream's own evaluator (ci_blacksmith_runner_test.go) over every context
// this repository can produce and fails if any of them reaches a Blacksmith
// label, so a guard on only the first operand, a lane re-keyed on
// needs.rbe.outputs.enabled (true in the fork modes), or a new expression
// shape all fail here, not on the first pull request.
//
// The evaluator reads an identifier the context does not mention as null (the
// real evaluator does), which can only make a comparison false. So a scan that
// forgets to supply something an expression reads can only err towards "never
// reaches Blacksmith": it passes, silently. It did, on F7b's include-matrix
// marker (strategy.matrix.include[*].runner), whose key an earlier version of
// this scan never read. The scan therefore reads include entries, and fails
// closed on any identifier it does not supply (forkRunsOnReachesBlacksmith).
//
// A workflow whose only trigger is a push to main never runs from this tree:
// main is a byte-identical upstream mirror and runs its own copy, so such a
// file is not scanned (its runs on the mirror gate nothing).

const forkBlacksmithGuard = "github.repository_owner == 'gastownhall' && "

var forkBlacksmithName = regexp.MustCompile(`(?i)\bblacksmith-[0-9]+vcpu`)

// forkModelledIdents: the identifiers forkRunsOnContexts supplies. A runs-on
// that reads any other identifier (and is not a matrix key of its own job) is
// a violation: the evaluator would read it as null and the scan would pass
// without having looked.
var forkModelledIdents = map[string]bool{
	"github.event_name":                             true,
	"github.event.pull_request.head.repo.full_name": true,
	"github.repository":                             true,
	"github.repository_owner":                       true,
	"github.actor":                                  true,
	"needs.rbe.outputs.mode":                        true,
	"needs.rbe.outputs.enabled":                     true,
}

// forkExprIdents: the identifiers expr reads, sorted and unique.
func forkExprIdents(expr string) ([]string, error) {
	body := strings.TrimSpace(expr)
	body = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(body, "${{"), "}}"))
	toks, err := ghTokenize(body)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var out []string
	for _, tok := range toks {
		if tok.kind == "ident" && !seen[tok.val] {
			seen[tok.val] = true
			out = append(out, tok.val)
		}
	}
	sort.Strings(out)
	return out, nil
}

// forkMatrixKey: the matrix key a scalar under .strategy.matrix. feeds, from
// its path after that marker. "os[1]" feeds os; "include[2].runner" feeds
// runner (an include entry's own value for it, and viaInclude says so: another
// combination may not get the key at all); an exclude entry feeds nothing,
// since leaving it out only widens the combinations tried.
func forkMatrixKey(rest string) (name string, ok, viaInclude bool) {
	head, tail, indexed := strings.Cut(rest, "[")
	switch {
	case head == "exclude":
		return "", false, false
	case head == "include" && indexed:
		_, entry, found := strings.Cut(tail, "].")
		if !found {
			return "", false, false
		}
		name, _, _ = strings.Cut(entry, "[")
		return name, name != "", true
	default:
		return head, true, false
	}
}

// forkRunsOnContexts: every context a runs-on expression can see in this
// repository. The owner is never gastownhall (or is missing); rbe's mode is
// every value its decision step can reach here, not "remote", with enabled
// derived from it the way that step derives it; every event shape, head
// repository and actor is tried, realistic or not, because one that reaches a
// Blacksmith label is a violation whatever the combination.
func forkRunsOnContexts(matrix map[string][]string) []map[string]string {
	ctxs := []map[string]string{}
	for _, owner := range []string{"versioned-beads", "someone-else", ""} {
		for _, event := range []string{"pull_request", "pull_request_target", "merge_group", "push", "workflow_dispatch", "schedule", "workflow_call"} {
			for _, head := range []string{"versioned-beads/beads", "other/beads", ""} {
				for _, actor := range []string{"alice", "dependabot[bot]"} {
					for _, mode := range []string{"skip", "cache", "local", "fork-ro", "fork-rw"} {
						ctx := map[string]string{
							"github.event_name":                             event,
							"github.event.pull_request.head.repo.full_name": head,
							"github.repository":                             "versioned-beads/beads",
							"github.actor":                                  actor,
							"needs.rbe.outputs.mode":                        mode,
							"needs.rbe.outputs.enabled":                     strconv.FormatBool(strings.HasPrefix(mode, "fork-")),
						}
						if owner != "" {
							ctx["github.repository_owner"] = owner
						}
						ctxs = append(ctxs, ctx)
					}
				}
			}
		}
	}
	names := make([]string, 0, len(matrix))
	for name := range matrix {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		var next []map[string]string
		for _, ctx := range ctxs {
			for _, v := range matrix[name] {
				c := map[string]string{"matrix." + name: v}
				for k, val := range ctx {
					c[k] = val
				}
				next = append(next, c)
			}
		}
		ctxs = next
	}
	return ctxs
}

// forkRunsOnReachesBlacksmith: "" when expr never resolves to a Blacksmith
// label in this repository, else why. An expression the evaluator cannot read,
// or that reads an identifier the scan does not supply, fails closed.
func forkRunsOnReachesBlacksmith(expr string, matrix map[string][]string) string {
	if !strings.HasPrefix(strings.TrimSpace(expr), "${{") {
		return "a literal Blacksmith label"
	}
	idents, err := forkExprIdents(expr)
	if err != nil {
		return "cannot be evaluated (" + err.Error() + "); extend this test or ask the architect"
	}
	read := map[string][]string{}
	for _, id := range idents {
		if key, ok := strings.CutPrefix(id, "matrix."); ok {
			if len(matrix[key]) == 0 {
				return "reads " + id + ", which this job's strategy.matrix never gives a value (a dynamic or unreadable matrix); extend this test or ask the architect"
			}
			read[key] = matrix[key]
		} else if !forkModelledIdents[id] {
			return "reads " + id + ", which this scan does not supply (the evaluator reads it as null, so a scan that leaves it out can only pass); extend this test or ask the architect"
		}
	}
	for _, ctx := range forkRunsOnContexts(read) {
		v, err := evalGHExpr(expr, ctx)
		if err != nil {
			return "cannot be evaluated (" + err.Error() + "); extend this test or ask the architect"
		}
		s, ok := v.(string)
		if !ok {
			return fmt.Sprintf("evaluates to %#v, not a runner label", v)
		}
		if forkBlacksmithName.MatchString(s) {
			return fmt.Sprintf("resolves to %s with owner %q, event %s, mode %s", s, ctx["github.repository_owner"], ctx["github.event_name"], ctx["needs.rbe.outputs.mode"])
		}
	}
	return ""
}

// forkMirrorOnly: the workflow's only trigger is a push to main.
func forkMirrorOnly(root *yaml.Node) bool {
	for i := 0; i+1 < len(root.Content); i += 2 {
		if root.Content[i].Value != "on" {
			continue
		}
		var on map[string]any
		if root.Content[i+1].Decode(&on) != nil {
			return false
		}
		return reflect.DeepEqual(on, map[string]any{"push": map[string]any{"branches": []any{"main"}}})
	}
	return false
}

// forkRunnerLabelViolations lists every scalar that spells a Blacksmith runner
// label anywhere but a runs-on value that cannot resolve to it here. A matrix
// entry, an input default, a runs-on list or mapping and a literal label all
// land here; prose that only says "Blacksmith" (a comment in a run script)
// does not.
func forkRunnerLabelViolations(root *yaml.Node) []string {
	if forkMirrorOnly(root) {
		return nil
	}
	const marker = ".strategy.matrix."
	matrix := map[string]map[string][]string{} // job path -> matrix key -> values
	viaInclude := map[string]map[string]bool{}
	walkYAML(root, "", func(path string, key bool, value string) {
		i := strings.Index(path, marker)
		if key || i < 0 {
			return
		}
		name, ok, include := forkMatrixKey(path[i+len(marker):])
		if !ok {
			return
		}
		job := path[:i]
		if matrix[job] == nil {
			matrix[job], viaInclude[job] = map[string][]string{}, map[string]bool{}
		}
		matrix[job][name] = append(matrix[job][name], value)
		viaInclude[job][name] = viaInclude[job][name] || include
	})
	for job, keys := range matrix {
		for name, values := range keys {
			if viaInclude[job][name] {
				values = append(values, "") // a combination no include entry reaches leaves the key unset
			}
			sort.Strings(values)
			keys[name] = slices.Compact(values)
		}
	}
	var out []string
	walkYAML(root, "", func(path string, key bool, value string) {
		if key || !forkBlacksmithName.MatchString(value) {
			return
		}
		job, ok := strings.CutSuffix(path, ".runs-on")
		if !ok {
			out = append(out, path+" = "+value+": a Blacksmith label outside a runs-on value")
			return
		}
		if why := forkRunsOnReachesBlacksmith(value, matrix[job]); why != "" {
			out = append(out, path+" = "+value+": "+why)
		}
	})
	return out
}

// Every workflow in the repository, not a fixed list, so a Blacksmith label in
// a file upstream adds later is caught too.
func TestForkWorkflowsNeverAskForBlacksmithOutsideUpstream(t *testing.T) {
	root := sourceRepoRoot(t)
	files, err := filepath.Glob(filepath.Join(root, ".github", "workflows", "*.y*ml"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no workflows under %s: %v", root, err)
	}
	sort.Strings(files)
	for _, file := range files {
		rel, err := filepath.Rel(root, file)
		if err != nil {
			t.Fatal(err)
		}
		for _, v := range forkRunnerLabelViolations(readYAMLNode(t, rel)) {
			t.Errorf("%s: %s; a job a pull request inside this fork can schedule would queue for a Blacksmith runner the fork does not have: apply rule R6 (be-8ofr93, be-6i64d8)", rel, v)
		}
	}
}

func forkWorkflowDoc(t *testing.T, trigger, body string) *yaml.Node {
	t.Helper()
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(trigger+"jobs:\n  a:\n"+body), &doc); err != nil {
		t.Fatalf("parse %q: %v", body, err)
	}
	return doc.Content[0]
}

// Known-positive and known-negative controls, independent of what the
// workflows hold: the scan reports upstream's F3 literal as shipped and every
// other shape it must not accept, and accepts the shapes it must. The first
// case proves the contexts can reach a Blacksmith label at all, so a scan that
// passes cannot be passing vacuously.
func TestForkRunnerLabelScanControls(t *testing.T) {
	unguarded := strings.Replace(sameRepoBlacksmith2vcpu, forkBlacksmithGuard, "", 1)
	guarded := "${{ " + forkBlacksmithGuard + strings.TrimPrefix(unguarded, "${{ ")
	firstOperandOnly := "${{ " + forkBlacksmithGuard + "(github.event_name == 'merge_group') || (github.event_name == 'pull_request') && 'blacksmith-2vcpu-ubuntu-2404' || 'ubuntu-latest' }}"
	modeKeyed := "${{ needs.rbe.outputs.mode == 'remote' && 'blacksmith-4vcpu-ubuntu-2404' || 'ubuntu-latest' }}"
	enabledKeyed := "${{ needs.rbe.outputs.enabled == 'true' && 'blacksmith-4vcpu-ubuntu-2404' || 'ubuntu-latest' }}"
	matrixTernary := "    strategy:\n      matrix:\n        runner: [blacksmith, github]\n    runs-on: ${{ matrix.runner == 'blacksmith' && 'blacksmith-8vcpu-ubuntu-2404' || 'ubuntu-latest' }}\n"
	// The marker's legs come from upstream's pinned constant, so a leg upstream
	// adds (#7231 added macOS, a third) is exercised without an edit here, and
	// every leg gets an include entry: a leg this control cannot reach would pass
	// unexamined (rule R6, be-5g3cmt). os is only the marker's last fallback.
	legKeys := forkMarkerLegKeys(t)
	everyLeg := 1<<len(legKeys) - 1
	includeMatrix := func(runsOn string) string {
		var b strings.Builder
		b.WriteString("    strategy:\n      matrix:\n        include:\n")
		for _, key := range legKeys {
			b.WriteString("          - os: hosted\n            runner: " + key + "\n")
		}
		return b.String() + "    runs-on: " + runsOn + "\n"
	}
	const pr = "on:\n  pull_request:\n"
	type scanCase struct {
		name    string
		trigger string
		body    string
		want    int
	}
	cases := []scanCase{
		{"upstream's F3 literal as shipped", pr, "    runs-on: " + unguarded + "\n", 1},
		{"upstream's 8 vCPU literal as shipped", pr, "    runs-on: " + strings.Replace(unguarded, "2vcpu", "8vcpu", 1) + "\n", 1},
		{"guard that covers only the first operand", pr, "    runs-on: " + firstOperandOnly + "\n", 1},
		{"lane keyed on rbe's enabled output", pr, "    runs-on: " + enabledKeyed + "\n", 1},
		{"matrix entry", pr, "    strategy:\n      matrix:\n        runner: [blacksmith-2vcpu-ubuntu-2404]\n    runs-on: ${{ matrix.runner }}\n", 1},
		{"runs-on list", pr, "    runs-on: [self-hosted, blacksmith-2vcpu-ubuntu-2404]\n", 1},
		{"literal label", pr, "    runs-on: blacksmith-2vcpu-ubuntu-2404\n", 1},
		{"label in an env value", pr, "    runs-on: ubuntu-latest\n    env:\n      RUNNER: blacksmith-2vcpu-ubuntu-2404\n", 1},
		{"expression the evaluator cannot read", pr, "    runs-on: ${{ contains(github.ref, 'x') && 'blacksmith-2vcpu-ubuntu-2404' || 'ubuntu-latest' }}\n", 1},
		{"matrix ternary on a pull_request workflow", pr, matrixTernary, 1},
		{"matrix ternary on a workflow that also has a dispatch trigger", "on:\n  push:\n    branches: [main]\n  workflow_dispatch:\n", matrixTernary, 1},
		{"include key unset in a combination no entry reaches", pr, "    strategy:\n      matrix:\n        os: [a, b]\n        include:\n          - os: a\n            runner: x\n    runs-on: ${{ matrix.runner != 'x' && 'blacksmith-2vcpu-ubuntu-2404' || 'ubuntu-latest' }}\n", 1},
		{"marker key no include entry or list gives", pr, "    strategy:\n      matrix:\n        os: [ubuntu-latest]\n    runs-on: " + forkMarkerRunsOn(t, everyLeg) + "\n", 1},
		{"runs-on reads a context the scan does not supply", pr, "    runs-on: ${{ vars.USE_BLACKSMITH == 'true' && 'blacksmith-2vcpu-ubuntu-2404' || 'ubuntu-latest' }}\n", 1},
		{"guarded literal", pr, "    runs-on: " + guarded + "\n", 0},
		{"mode-keyed literal", pr, "    runs-on: " + modeKeyed + "\n", 0},
		{"hosted runner", pr, "    runs-on: ubuntu-latest\n", 0},
		{"prose in a run script", pr, "    steps:\n      - run: |\n          # the Blacksmith runner's environment caps argv\n          echo hi\n", 0},
		{"matrix ternary on a workflow that only runs on a push to main", "on:\n  push:\n    branches: [main]\n", matrixTernary, 0},
	}
	// The marker with every subset of its legs guarded (leg i is bit i, in the
	// constant's order): the scan accepts it only when every leg is. A fixed list
	// of rows breaks when upstream adds a leg, and a row that is missing is a leg
	// that passes unexamined.
	for guarded := 0; guarded <= everyLeg; guarded++ {
		var on, off []string
		for i, key := range legKeys {
			if guarded>>i&1 == 1 {
				on = append(on, key)
			} else {
				off = append(off, key)
			}
		}
		c := scanCase{"include-matrix marker with " + strings.Join(on, ", ") + " guarded and " + strings.Join(off, ", ") + " open", pr, includeMatrix(forkMarkerRunsOn(t, guarded)), 1}
		switch guarded {
		case 0:
			c.name = "include-matrix marker as F7b shipped it (no leg guarded)"
		case everyLeg:
			c.name, c.want = "include-matrix marker with every leg guarded", 0
		}
		cases = append(cases, c)
	}
	for _, c := range cases {
		if got := forkRunnerLabelViolations(forkWorkflowDoc(t, c.trigger, c.body)); len(got) != c.want {
			t.Errorf("%s: %d violations %q, want %d", c.name, len(got), got, c.want)
		}
	}
	// Not vacuous: outside this fork the unguarded literal does reach Blacksmith.
	upstream := map[string]string{
		"github.repository_owner": "gastownhall", "github.event_name": "pull_request", "github.repository": "gastownhall/beads",
		"github.event.pull_request.head.repo.full_name": "gastownhall/beads", "github.actor": "alice",
	}
	if got := mustEvalGHRunsOn(t, unguarded, upstream); !strings.HasPrefix(got, "blacksmith-") {
		t.Errorf("upstream's literal on a same-repository pull request in gastownhall = %q, want a Blacksmith label", got)
	}
	if got := mustEvalGHRunsOn(t, guarded, upstream); !strings.HasPrefix(got, "blacksmith-") {
		t.Errorf("the guarded literal on a same-repository pull request in gastownhall = %q, want it unchanged from upstream", got)
	}
	upstream["github.repository_owner"] = "versioned-beads"
	if got := mustEvalGHRunsOn(t, guarded, upstream); got != "ubuntu-latest" {
		t.Errorf("the guarded literal in versioned-beads = %q, want ubuntu-latest", got)
	}
}

// forkMarkerLeg heads the same-repository predicate each leg of upstream's
// platforms marker (the chained ternary on matrix.runner, F7b) carries once the
// guard is taken out; forkMarkerLegKey finds the matrix.runner value that leg
// tests for.
const forkMarkerLeg = "((github.event_name == 'merge_group'"

var forkMarkerLegKey = regexp.MustCompile(`matrix\.runner == '([^']+)' && `)

// forkMarkerLegKeys: the matrix.runner value of each leg of the marker, in the
// constant's order, read from the pinned constant so the controls follow
// upstream's text. The legs are counted twice, by key and by predicate. A marker
// whose counts differ, with fewer than two legs ("some legs guarded" means
// nothing) or with more than six (the controls try every subset: 2^legs rows)
// is a shape this control was not written for: ask the architect (rule R6,
// be-5g3cmt; #7231 took the marker from two legs to three).
func forkMarkerLegKeys(t *testing.T) []string {
	t.Helper()
	var keys []string
	for _, m := range forkMarkerLegKey.FindAllStringSubmatch(sameRepoPlatformsMatrixMarkerRunsOn, -1) {
		keys = append(keys, m[1])
	}
	legs := strings.Count(strings.ReplaceAll(sameRepoPlatformsMatrixMarkerRunsOn, forkBlacksmithGuard, ""), forkMarkerLeg)
	if len(keys) != legs || legs < 2 || legs > 6 {
		t.Fatalf("sameRepoPlatformsMatrixMarkerRunsOn has %d matrix.runner keys %q and %d same-repository predicates, want the same number, from 2 to 6: update this control (rule R6, be-5g3cmt)", len(keys), keys, legs)
	}
	return keys
}

// forkMarkerRunsOn builds upstream's platforms marker from its pinned constant,
// so it follows upstream's text, with the guard in front of the same-repository
// predicate of leg i when bit i of guarded is set.
func forkMarkerRunsOn(t *testing.T, guarded int) string {
	t.Helper()
	parts := strings.Split(strings.ReplaceAll(sameRepoPlatformsMatrixMarkerRunsOn, forkBlacksmithGuard, ""), forkMarkerLeg)
	if len(parts)-1 != len(forkMarkerLegKeys(t)) {
		t.Fatalf("sameRepoPlatformsMatrixMarkerRunsOn has %d same-repository legs: update this control (rule R6, be-5g3cmt)", len(parts)-1)
	}
	var b strings.Builder
	b.WriteString(parts[0])
	for i, rest := range parts[1:] {
		if guarded>>i&1 == 1 {
			b.WriteString("(" + forkBlacksmithGuard + "(github.event_name == 'merge_group'")
		} else {
			b.WriteString(forkMarkerLeg)
		}
		b.WriteString(rest)
	}
	return b.String()
}

// forkUpstreamWorld: the context an upstream evaluator test means. Upstream's
// tests model a pull request in gastownhall/beads and know nothing of an owner,
// so a context that does not name one gets upstream's. Only the two helpers that
// evaluate a runs-on for upstream's tests call this (mustEvalGHRunsOn,
// resolveRunsOnLabel; rule R6's E1 puts the call in); this file's scan calls
// evalGHExpr itself and always names the owner, because for the scan a missing
// owner is a case to try, not a default.
func forkUpstreamWorld(ctx map[string]string) map[string]string {
	if _, ok := ctx["github.repository_owner"]; ok {
		return ctx
	}
	world := make(map[string]string, len(ctx)+1)
	for k, v := range ctx {
		world[k] = v
	}
	world["github.repository_owner"] = "gastownhall"
	return world
}

// Upstream's cache sweep (TestBlacksmithReachableAdvisoryJobsNeverSaveACache)
// skips every job whose runs-on does not resolve to a Blacksmith label under
// its two trust contexts. Those contexts name no owner, so once the guard is in
// every job resolves to the hosted label, the sweep examines nothing, and it
// still passes: 16 of its 76 (job, context) pairs before rule R6, 0 after (the
// first R6 v2 sync, vb #101). This pins that the sweep still reaches every job
// whose runs-on spells the same-repository predicate in front of a Blacksmith
// label. If F7c's sweep is renamed or removed, delete this with it.
func TestForkUpstreamSweepStillReachesBlacksmithJobs(t *testing.T) {
	ctxNames := make([]string, 0, len(blacksmithTrustContexts))
	for ctxName := range blacksmithTrustContexts {
		ctxNames = append(ctxNames, ctxName)
	}
	sort.Strings(ctxNames)
	spelled := 0
	for _, file := range generalCacheSweepWorkflows {
		jobs := readCIWorkflow(t, file).Jobs
		names := make([]string, 0, len(jobs))
		for name := range jobs {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			runsOn := jobs[name].RunsOn
			if !forkBlacksmithName.MatchString(runsOn) || !strings.Contains(runsOn, "head.repo.full_name == github.repository") {
				continue
			}
			for _, ctxName := range ctxNames {
				spelled++
				if got := resolveRunsOnLabel(runsOn, blacksmithTrustContexts[ctxName]); !strings.HasPrefix(got, "blacksmith-") {
					t.Errorf("%s job %s: runs-on resolves to %q under upstream's %s context, so upstream's cache sweep skips it; put the owner in upstream's world (rule R6, E1, be-ju6she)", file, name, got, ctxName)
				}
			}
		}
	}
	if spelled == 0 {
		t.Errorf("no job in the cache sweep's workflows spells the same-repository Blacksmith predicate: the sweep examines nothing, or upstream changed the shape; ask the architect")
	}
}
