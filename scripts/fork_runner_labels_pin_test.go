package scripts_test

import (
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// This file is the fork's own (upstream has no copy, so syncs never conflict
// on it). It pins rule R6: no workflow job may ask for a Blacksmith runner
// outside upstream's organization, because this repository has no pool.
//
// Upstream F3 moved the rbe job, PR Risk's detect-ci-tier and bazel-coverage
// jobs, and pr.yml's bazel-coverage and ci-gate onto blacksmith-2vcpu-
// ubuntu-2404 for every same-repository pull request, merge_group, push,
// dispatch and schedule run. Every pull request inside this repository is
// same-repository, so upstream's predicate is true for all of them and the job
// queues for a runner that does not exist. Every upstream test still passes:
// they pin the expression's text, not where the job can run.
//
// The fork's edit is one conjunct in front of the predicate,
// github.repository_owner == 'gastownhall', so outside that organization the
// expression always resolves to ubuntu-latest, and a missing or renamed
// context field fails towards the hosted runner. Both constants upstream
// compares the workflows to (sameRepoBlacksmith2vcpu, wantRBERunsOn) carry the
// same conjunct. This file proves it still covers the whole condition and that
// no other form of Blacksmith label is in any workflow.
//
// The bazel.yml lanes that pick Blacksmith with needs.rbe.outputs.enabled ==
// 'true' need no guard: the rbe job decides remote only when the repository
// variable RBE_WEST_WORKERS and the secret RBE_WEST_EXECUTOR both exist and the
// run is not a pull request inside a GitHub fork. This repository sets neither
// (verify with the repository's Actions variables and secrets), and every pull
// request in it has head.repo.fork true.

const forkBlacksmithGuard = "github.repository_owner == 'gastownhall' && "

var (
	forkBlacksmithName = regexp.MustCompile(`(?i)\bblacksmith-[0-9]+vcpu`)
	forkRBEKeyedRunsOn = regexp.MustCompile(`^\$\{\{ needs\.rbe\.outputs\.enabled == 'true' && 'blacksmith-[24]vcpu-ubuntu-2404' \|\| 'ubuntu-latest' \}\}$`)
	forkBlacksmithTail = regexp.MustCompile(`^ && 'blacksmith-[24]vcpu-ubuntu-2404' \|\| 'ubuntu-latest' \}\}$`)
)

// forkGuardDominates: expr is "${{ <guard>(<condition>) && 'blacksmith-...' ||
// 'ubuntu-latest' }}" and the parenthesis opened right after the guard closes
// just before the Blacksmith branch, so the guard applies to the whole
// condition rather than to its first operand.
func forkGuardDominates(expr string) bool {
	rest, ok := strings.CutPrefix(expr, "${{ "+forkBlacksmithGuard+"(")
	if !ok {
		return false
	}
	depth, quoted := 1, false
	for i := 0; i < len(rest); i++ {
		switch c := rest[i]; {
		case c == '\'':
			quoted = !quoted
		case quoted:
		case c == '(':
			depth++
		case c == ')':
			depth--
			if depth == 0 {
				return forkBlacksmithTail.MatchString(rest[i+1:])
			}
		}
	}
	return false
}

// forkRunnerLabelViolations lists every scalar that spells a Blacksmith runner
// label anywhere but a runs-on value that is rbe-keyed or guarded. A matrix
// entry, an input default, a runs-on list or mapping and a new expression shape
// all land here; prose that only says "Blacksmith" (a comment in a run script)
// does not.
func forkRunnerLabelViolations(doc *yaml.Node) []string {
	var out []string
	walkYAML(doc, "", func(path string, key bool, value string) {
		if key || !forkBlacksmithName.MatchString(value) {
			return
		}
		if strings.HasSuffix(path, ".runs-on") && (forkRBEKeyedRunsOn.MatchString(value) || forkGuardDominates(value)) {
			return
		}
		out = append(out, path+" = "+value)
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
			t.Errorf("%s: %s: a job a pull request inside this fork can schedule would queue for a Blacksmith runner the fork does not have; apply rule R6 (be-8ofr93)", rel, v)
		}
	}
}

func forkWorkflowDoc(t *testing.T, body string) *yaml.Node {
	t.Helper()
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte("jobs:\n  a:\n"+body), &doc); err != nil {
		t.Fatalf("parse %q: %v", body, err)
	}
	return doc.Content[0]
}

// Known-positive controls, independent of what the constants hold: the scan
// reports upstream's F3 literal as shipped and every other shape it must not
// accept, and accepts the shapes it must.
func TestForkRunnerLabelScanControls(t *testing.T) {
	unguarded := strings.Replace(sameRepoBlacksmith2vcpu, forkBlacksmithGuard, "", 1)
	guarded := "${{ " + forkBlacksmithGuard + strings.TrimPrefix(unguarded, "${{ ")
	partial := "${{ " + forkBlacksmithGuard + "(github.event_name == 'merge_group') || (github.event_name == 'pull_request') && 'blacksmith-2vcpu-ubuntu-2404' || 'ubuntu-latest' }}"
	rbeKeyed := "${{ needs.rbe.outputs.enabled == 'true' && 'blacksmith-4vcpu-ubuntu-2404' || 'ubuntu-latest' }}"
	cases := []struct {
		name string
		body string
		want int
	}{
		{"upstream's F3 literal as shipped", "    runs-on: " + unguarded + "\n", 1},
		{"guard that covers only the first operand", "    runs-on: " + partial + "\n", 1},
		{"matrix entry", "    strategy:\n      matrix:\n        runner: [blacksmith-2vcpu-ubuntu-2404]\n    runs-on: ${{ matrix.runner }}\n", 1},
		{"runs-on list", "    runs-on: [self-hosted, blacksmith-2vcpu-ubuntu-2404]\n", 1},
		{"guarded literal", "    runs-on: " + guarded + "\n", 0},
		{"rbe-keyed literal", "    runs-on: " + rbeKeyed + "\n", 0},
		{"hosted runner", "    runs-on: ubuntu-latest\n", 0},
		{"prose in a run script", "    steps:\n      - run: |\n          # the Blacksmith runner's environment caps argv\n          echo hi\n", 0},
	}
	for _, c := range cases {
		if got := forkRunnerLabelViolations(forkWorkflowDoc(t, c.body)); len(got) != c.want {
			t.Errorf("%s: %d violations %q, want %d", c.name, len(got), got, c.want)
		}
	}
	if !forkGuardDominates(guarded) || forkGuardDominates(unguarded) || forkGuardDominates(partial) {
		t.Errorf("forkGuardDominates: guarded %v, unguarded %v, partial %v; want true, false, false",
			forkGuardDominates(guarded), forkGuardDominates(unguarded), forkGuardDominates(partial))
	}
}
