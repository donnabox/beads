package scripts_test

import (
	"fmt"
	"path/filepath"
	"reflect"
	"regexp"
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
// A workflow whose only trigger is a push to main never runs from this tree:
// main is a byte-identical upstream mirror and runs its own copy, so such a
// file is not scanned (its runs on the mirror gate nothing).

const forkBlacksmithGuard = "github.repository_owner == 'gastownhall' && "

var forkBlacksmithName = regexp.MustCompile(`(?i)\bblacksmith-[0-9]+vcpu`)

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
// label in this repository, else why. An expression the evaluator cannot read
// fails closed.
func forkRunsOnReachesBlacksmith(expr string, matrix map[string][]string) string {
	if !strings.HasPrefix(strings.TrimSpace(expr), "${{") {
		return "a literal Blacksmith label"
	}
	for _, ctx := range forkRunsOnContexts(matrix) {
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
	walkYAML(root, "", func(path string, key bool, value string) {
		i := strings.Index(path, marker)
		if key || i < 0 {
			return
		}
		name, _, _ := strings.Cut(path[i+len(marker):], "[")
		if matrix[path[:i]] == nil {
			matrix[path[:i]] = map[string][]string{}
		}
		matrix[path[:i]][name] = append(matrix[path[:i]][name], value)
	})
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
	const pr = "on:\n  pull_request:\n"
	cases := []struct {
		name    string
		trigger string
		body    string
		want    int
	}{
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
		{"guarded literal", pr, "    runs-on: " + guarded + "\n", 0},
		{"mode-keyed literal", pr, "    runs-on: " + modeKeyed + "\n", 0},
		{"hosted runner", pr, "    runs-on: ubuntu-latest\n", 0},
		{"prose in a run script", pr, "    steps:\n      - run: |\n          # the Blacksmith runner's environment caps argv\n          echo hi\n", 0},
		{"matrix ternary on a workflow that only runs on a push to main", "on:\n  push:\n    branches: [main]\n", matrixTernary, 0},
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
