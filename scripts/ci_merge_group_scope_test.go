package scripts_test

import (
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// The merge-group fast set (engdocs/CI_REQUIRED_CHECK_TOPOLOGY.md, "Merge
// Queue"). A merge group runs only the fast build and unit set: every
// Windows, macOS and Nix job and bazel.yml's Dolt-tier lanes are skipped
// there by design. The pull request's own run before it runs everything,
// and main's push runs run the complement after it lands, so every job the
// required gates read runs on every push head on main: the merge queue lands
// a batch of entries as one push of its last entry, whose merge group ran
// the fast set on the same tree (main fast-forwards to it), and the push
// runs (pr.yml's and pr-risk.yml's, and bazel.yml's own, which runs every
// Bazel lane) run the rest on it. That tree contains every entry of the
// batch; an earlier entry gets the complement only as part of it.
//
// One script decides, .github/scripts/ci-scope.sh, run by pr.yml's and
// pr-risk.yml's scope jobs; it alone reads the event. Every other job reads
// the scope's outputs, and the two required gates excuse a skip only where
// the scope says so: ci-scope.sh's skips for the workflows' own job sets,
// bazel-gate.sh's for the Dolt-tier lanes, which also needs the Bazel call
// to report that it ran without them. These tests run the real script, every
// job's real if: through evalGHExpr, and both gates' real steps.

const (
	ciScopeScript = ".github/scripts/ci-scope.sh"
	ciScopeJob    = "scope"
	// bazel.yml's ordinary Dolt-tier switch: each lane retains this rule,
	// with exact-fork full-local coverage overriding it to on. The call reports what it did
	// as its dolt-tiers output.
	bazelDoltTiersCond   = "inputs.dolt-tiers != 'off'"
	bazelDoltTiersOutput = "${{ jobs.rbe.outputs.local_full == 'true' && 'on' || (inputs.dolt-tiers == 'off' && 'off' || 'on') }}"
)

// ciScopeRows: ci-scope.sh's decision by event (the two workflows'
// triggers). Any other event has no row and fails the scope job.
var ciScopeRows = map[string]map[string]string{
	// Everything, as before the fast set.
	"pull_request": {"platforms": "true", "dolt-tiers": "on", "linux": "true", "advisory": "true"},
	// The requested fast set; Q still executes complete dedicated Dolt coverage.
	"merge_group": {"platforms": "false", "dolt-tiers": "off", "linux": "true", "advisory": "false"},
	// The complement, after the merge: the Windows, macOS and Nix jobs.
	// bazel.yml's own push run is every Bazel lane, so pr.yml calls none.
	"push": {"platforms": "true", "dolt-tiers": "on", "linux": "false", "advisory": "false"},
}

// ciScopePlatformsIf: the if of a job on the platforms switch (the Windows,
// macOS and Nix jobs): it runs on every pull request and every push to main,
// and only a merge group skips it, which the gates excuse there alone
// (TestCIScopePRGate, TestCIScopeRiskGate).
const ciScopePlatformsIf = "${{ needs." + ciScopeJob + ".outputs.platforms == 'true' }}"

// ciScopeOutputs: the scope job's outputs, in the gate env's order.
var ciScopeOutputs = []string{"platforms", "dolt-tiers", "linux", "advisory"}

// ciScopeJobSwitch: the scope output each job of the two workflows runs on
// (if: needs.scope.outputs.<switch> == 'true'). The scope job decides and the
// gates always run; every other job is listed, so a new job cannot escape the
// decision (TestCIScopeJobsReadTheScope).
type scopeJobContract struct {
	Needs []string
	If    string
}

var scopeJobContracts = map[string]map[string]scopeJobContract{
	"pr.yml": {
		"scope":                                  {Needs: []string{}, If: ""},
		"build-artifacts":                        {Needs: []string{"scope", "bazel-coverage"}, If: "${{ needs.scope.outputs.linux == 'true' && (needs.bazel-coverage.outputs.pr_lanes != 'true') }}"},
		"fast-checks":                            {Needs: []string{"scope"}, If: "${{ needs.scope.outputs.linux == 'true' }}"},
		"check-cmd-bd-puregeo-tests":             {Needs: []string{"scope", "bazel-coverage"}, If: "${{ needs.scope.outputs.linux == 'true' && (needs.bazel-coverage.outputs.pr_lanes != 'true') }}"},
		"check-release-target-cross-compilation": {Needs: []string{"scope"}, If: "${{ needs.scope.outputs.linux == 'true' }}"},
		"test-windows-liveness":                  {Needs: []string{"scope"}, If: "${{ needs.scope.outputs.platforms == 'true' }}"},
		"test-windows-small":                     {Needs: []string{"scope"}, If: "${{ needs.scope.outputs.platforms == 'true' }}"},
		"worktree-remove-windows":                {Needs: []string{"scope"}, If: "${{ needs.scope.outputs.platforms == 'true' }}"},
		"windows-test-binaries":                  {Needs: []string{"scope"}, If: "${{ needs.scope.outputs.platforms == 'true' }}"},
		"test-windows-liveness-prebuilt":         {Needs: []string{"scope", "windows-test-binaries"}, If: "${{ needs.scope.outputs.platforms == 'true' }}"},
		"worktree-remove-windows-prebuilt":       {Needs: []string{"scope", "windows-test-binaries"}, If: "${{ needs.scope.outputs.platforms == 'true' }}"},
		"check-doc-flags":                        {Needs: []string{"scope"}, If: "${{ needs.scope.outputs.linux == 'true' }}"},
		"check-doc-freshness-platforms":          {Needs: []string{"scope"}, If: "${{ needs.scope.outputs.platforms == 'true' }}"},
		"pr-preflight-platforms":                 {Needs: []string{"scope"}, If: "${{ needs.scope.outputs.platforms == 'true' }}"},
		"pr-policy-wrapper":                      {Needs: []string{"scope"}, If: "${{ needs.scope.outputs.linux == 'true' }}"},
		"pr-core-wrapper":                        {Needs: []string{"scope", "build-artifacts", "bazel-coverage"}, If: "${{ needs.scope.outputs.linux == 'true' && (needs.bazel-coverage.outputs.pr_lanes != 'true') }}"},
		"scripts-go-checks":                      {Needs: []string{"scope"}, If: "${{ needs.scope.outputs.linux == 'true' }}"},
		"pr-lint-wrapper":                        {Needs: []string{"scope"}, If: "${{ needs.scope.outputs.linux == 'true' }}"},
		"advisory-reports":                       {Needs: []string{"scope"}, If: "${{ needs.scope.outputs.advisory == 'true' }}"},
		"test-domain-uow":                        {Needs: []string{"scope", "build-artifacts", "bazel-coverage"}, If: "${{ needs.scope.outputs.linux == 'true' && (needs.bazel-coverage.outputs.pr_lanes != 'true') }}"},
		"test-dolt-server-fingerprint":           {Needs: []string{"scope"}, If: "${{ needs.scope.outputs.linux == 'true' }}"},
		"contract-corpus":                        {Needs: []string{"scope", "bazel-coverage"}, If: "${{ needs.scope.outputs.linux == 'true' && (needs.bazel-coverage.outputs.pr_lanes != 'true') }}"},
		"windows-make-shell":                     {Needs: []string{"scope"}, If: "${{ needs.scope.outputs.platforms == 'true' }}"},
		"graph-c0-cohort":                        {Needs: []string{"scope", "build-artifacts"}, If: "${{ needs.scope.outputs.linux == 'true' }}"},
		"graph-c0":                               {Needs: []string{"scope", "build-artifacts", "graph-c0-cohort"}, If: "${{ needs.scope.outputs.linux == 'true' && (always()) }}"},
		"bazel":                                  {Needs: []string{"scope"}, If: "${{ needs.scope.outputs.linux == 'true' }}"},
		"bazel-coverage":                         {Needs: []string{"scope"}, If: "${{ needs.scope.outputs.linux == 'true' }}"},
		"ci-gate":                                {Needs: []string{"scope", "build-artifacts", "fast-checks", "graph-c0", "check-cmd-bd-puregeo-tests", "check-release-target-cross-compilation", "test-windows-liveness", "worktree-remove-windows", "test-windows-liveness-prebuilt", "worktree-remove-windows-prebuilt", "check-doc-flags", "check-doc-freshness-platforms", "pr-preflight-platforms", "pr-policy-wrapper", "pr-core-wrapper", "scripts-go-checks", "pr-lint-wrapper", "test-domain-uow", "test-dolt-server-fingerprint", "contract-corpus", "windows-make-shell", "bazel", "bazel-coverage"}, If: "${{ always() }}"},
	},
	"pr-risk.yml": {
		"scope":                     {Needs: []string{}, If: ""},
		"detect-ci-tier":            {Needs: []string{"scope"}, If: "${{ needs.scope.outputs.linux == 'true' }}"},
		"bazel-coverage":            {Needs: []string{"scope"}, If: "${{ needs.scope.outputs.linux == 'true' }}"},
		"test-proxied-cmd":          {Needs: []string{"scope", "detect-ci-tier", "bazel-coverage", "build-embedded"}, If: "${{ needs.scope.outputs.linux == 'true' && needs.scope.outputs.dolt-tiers == 'on' && (needs.detect-ci-tier.outputs.full_embedded == 'true' && needs.bazel-coverage.outputs.dolt_server != 'true') }}"},
		"build-embedded":            {Needs: []string{"scope", "detect-ci-tier", "bazel-coverage"}, If: "${{ needs.scope.outputs.linux == 'true' && needs.scope.outputs.dolt-tiers == 'on' && (needs.detect-ci-tier.outputs.full_embedded == 'true' && (needs.bazel-coverage.outputs.embedded != 'true' || needs.bazel-coverage.outputs.dolt_server != 'true')) }}"},
		"test-embedded-storage":     {Needs: []string{"scope", "detect-ci-tier", "bazel-coverage", "build-embedded"}, If: "${{ needs.scope.outputs.linux == 'true' && needs.scope.outputs.dolt-tiers == 'on' && (needs.detect-ci-tier.outputs.full_embedded == 'true' && needs.bazel-coverage.outputs.embedded != 'true') }}"},
		"test-embedded-conformance": {Needs: []string{"scope", "detect-ci-tier", "bazel-coverage", "build-embedded"}, If: "${{ needs.scope.outputs.linux == 'true' && needs.scope.outputs.dolt-tiers == 'on' && (needs.detect-ci-tier.outputs.full_embedded == 'true' && needs.bazel-coverage.outputs.embedded != 'true') }}"},
		"test-server-storage":       {Needs: []string{"scope", "detect-ci-tier", "bazel-coverage", "build-embedded"}, If: "${{ needs.scope.outputs.linux == 'true' && needs.scope.outputs.dolt-tiers == 'on' && (needs.detect-ci-tier.outputs.full_embedded == 'true' && needs.bazel-coverage.outputs.dolt_server != 'true') }}"},
		"test-server-storage-full":  {Needs: []string{"scope", "detect-ci-tier", "bazel-coverage", "build-embedded"}, If: "${{ needs.scope.outputs.linux == 'true' && needs.scope.outputs.dolt-tiers == 'on' && (needs.detect-ci-tier.outputs.full_embedded == 'true' && needs.bazel-coverage.outputs.dolt_server != 'true') }}"},
		"test-embedded-cmd":         {Needs: []string{"scope", "detect-ci-tier", "bazel-coverage", "build-embedded"}, If: "${{ needs.scope.outputs.linux == 'true' && needs.scope.outputs.dolt-tiers == 'on' && (needs.detect-ci-tier.outputs.full_embedded == 'true' && needs.bazel-coverage.outputs.embedded != 'true') }}"},
		"test-nix":                  {Needs: []string{"scope"}, If: "${{ needs.scope.outputs.platforms == 'true' }}"},
		"ci-gate":                   {Needs: []string{"scope", "detect-ci-tier", "bazel-coverage", "build-embedded", "test-embedded-storage", "test-embedded-conformance", "test-server-storage", "test-server-storage-full", "test-embedded-cmd", "test-proxied-cmd", "test-nix"}, If: "${{ always() }}"},
	},
}

// Categories support event scenarios. Exact dependencies and predicates are
// pinned independently in scopeJobContracts above, including every Graph and
// legacy coverage dependency; these are not inferred from production YAML.
var ciScopeJobSwitch = func() map[string]map[string]string {
	out := map[string]map[string]string{}
	for workflow, contracts := range scopeJobContracts {
		out[workflow] = map[string]string{}
		for name, contract := range contracts {
			if name == ciScopeJob || name == "ci-gate" {
				continue
			}
			for _, category := range []string{"linux", "platforms", "advisory"} {
				if strings.Contains(contract.If, "needs.scope.outputs."+category+" == 'true'") {
					out[workflow][name] = category
				}
			}
		}
	}
	return out
}()

// ciScopeGateIDs: per workflow, the gate ids each scope switch turns off
// when it is false (the gate step's CI_SCOPE_<SWITCH>_IDS, which
// ci-scope.sh's skips prints): pr.yml's Windows and macOS jobs on a merge
// group, its Linux jobs and its whole Bazel call on a push to main (bazel.yml
// runs on that push by itself), pr-risk.yml's Nix job on a merge group. The
// Dolt-tier lanes' ids are bazel-gate.sh's (bazelDoltTierGateIDs).
var ciScopeGateIDs = map[string]map[string][]string{
	"pr.yml": {
		"platforms": strings.Fields("CHECK_DOC_FRESHNESS_PLATFORMS PR_PREFLIGHT_PLATFORMS TEST_WINDOWS_LIVENESS TEST_WINDOWS_LIVENESS_PREBUILT WINDOWS_MAKE_SHELL WORKTREE_REMOVE_WINDOWS WORKTREE_REMOVE_WINDOWS_PREBUILT"),
		"linux":     strings.Fields("BAZEL BAZEL_CMD_DOLT BAZEL_COVERAGE BAZEL_DOLTSERVER BAZEL_DOLT_SERVER_RETIRED BAZEL_EMBEDDED BAZEL_EMBEDDED_RETIRED BAZEL_INTEGRATION BAZEL_LOCAL_FULL BAZEL_PROXIED BAZEL_PR_LANES_RETIRED BAZEL_PURE BAZEL_RELEASE_CROSS BAZEL_REMOTE BAZEL_SERVER_STORAGE BAZEL_TEST BUILD_ARTIFACTS CHECK_BUILD_TAGS CHECK_CMD_BD_PUREGEO_TESTS CHECK_DOC_FLAGS CHECK_MIGRATION_HYGIENE CHECK_NO_BEADS_CHANGES CHECK_RELEASE_TARGET_CROSS_COMPILATION CHECK_VERSION_CONSISTENCY CONTRACT_CORPUS FAST_CHECKS FMT_CHECK GRAPH_C0 PACKAGE_MCP PACKAGE_NPM PR_CORE_WRAPPER PR_LINT_WRAPPER PR_POLICY_WRAPPER SCRIPTS_GO_CHECKS TEST_DOLT_SERVER_FINGERPRINT TEST_DOMAIN_UOW"),
	},
	prRiskWorkflowName: {
		"platforms": {"TEST_NIX"},
		"linux":     strings.Fields("DETECT_CI_TIER BAZEL_COVERAGE BUILD_EMBEDDED TEST_EMBEDDED_STORAGE TEST_EMBEDDED_CONFORMANCE TEST_SERVER_STORAGE TEST_SERVER_STORAGE_FULL TEST_EMBEDDED_CMD TEST_PROXIED_CMD"),
	},
}

// bazelDoltTierGateIDs: the gate ids of bazel.yml's Dolt-tier lanes, which a
// merge group skips. bazel-gate.sh excuses them only when the scope turned
// the tiers off and the call reports it ran without them.
var bazelDoltTierGateIDs = []string{"BAZEL_EMBEDDED", "BAZEL_DOLTSERVER", "BAZEL_PROXIED", "BAZEL_SERVER_STORAGE", "BAZEL_CMD_DOLT"}

// bazelDoltTierLanes: the bazel.yml jobs a dolt-tiers "off" call skips in
// every mode: the Dolt tiers' unions (mode remote) and their members (every
// other mode), and the embedded tier.
var bazelDoltTierLanes = []string{bazelEmbedJobName, bazelDoltJobName, bazelProxiedJobName, bazelDoltRaceJobName, bazelServerJobName, bazelCmdDoltJobName, bazelDoltIntJobName}

// withDoltTiers: a Dolt-tier lane's if, its own rule plus the caller's
// dolt-tiers switch.
func withDoltTiers(ifExpr string) string {
	return strings.TrimSuffix(ifExpr, " }}") + " && " + bazelDoltTiersCond + " }}"
}

// ciScopeEnvName: the gate env suffix of a scope output (dolt-tiers ->
// DOLT_TIERS).
func ciScopeEnvName(output string) string {
	return strings.ToUpper(strings.ReplaceAll(output, "-", "_"))
}

// ciScopeCtx: an evalGHExpr context holding the scope job's outputs.
func ciScopeCtx(scope map[string]string) map[string]string {
	ctx := map[string]string{}
	for k, v := range scope {
		ctx["needs."+ciScopeJob+".outputs."+k] = v
	}
	return ctx
}

// runCIScope runs ci-scope.sh with args and env, returning its stdout, its
// $GITHUB_OUTPUT and its error (with the output folded in).
func runCIScope(t *testing.T, env map[string]string, args ...string) (string, map[string]string, error) {
	t.Helper()
	requireHostTool(t, "bash")
	dir := t.TempDir()
	output := filepath.Join(dir, "output")
	cmd := exec.Command("bash", append([]string{filepath.Join(sourceRepoRoot(t), ciScopeScript)}, args...)...)
	cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "GITHUB_OUTPUT=" + output, "GITHUB_STEP_SUMMARY=" + filepath.Join(dir, "summary")}
	for k, v := range env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	var stdout, stderr strings.Builder
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	if err != nil {
		err = fmt.Errorf("%v: %s%s", err, stdout.String(), stderr.String())
	}
	return strings.TrimSpace(stdout.String()), rrcOutputs(t, output), err
}

// ciScopeFor: the scope job's outputs for an event, from the real script.
func ciScopeFor(t *testing.T, event string) map[string]string {
	t.Helper()
	_, out, err := runCIScope(t, map[string]string{"GITHUB_EVENT_NAME": event}, "decide")
	if err != nil {
		t.Fatalf("%s decide on %s: %v", ciScopeScript, event, err)
	}
	return out
}

// prCallWith: pr.yml's bazel call inputs with the scope's outputs resolved.
// Only `${{ needs.scope.outputs.<x> }}` may be an expression there.
func prCallWith(t *testing.T, with, scope map[string]string) map[string]string {
	t.Helper()
	ref := regexp.MustCompile(`^\$\{\{ needs\.` + ciScopeJob + `\.outputs\.([a-z-]+) \}\}$`)
	out := map[string]string{}
	for k, v := range with {
		if strings.Contains(v, "${{") {
			m := ref.FindStringSubmatch(v)
			if m == nil {
				t.Fatalf("pr.yml bazel call with %s = %q: only the scope's outputs may be expressions there", k, v)
			}
			v = scope[m[1]]
		}
		out[k] = v
	}
	return out
}

// prCallWithFor: pr.yml's bazel call inputs on an event (ciScopeRows).
func prCallWithFor(t *testing.T, event string) map[string]string {
	t.Helper()
	return prCallWith(t, readCIWorkflow(t, "pr.yml").job(t, "bazel").With, ciScopeRows[event])
}

// bazelPRCallLanesFor: the modes each bazel.yml lane runs in under pr.yml's
// call on an event.
func bazelPRCallLanesFor(t *testing.T, event string) map[string]map[string]bool {
	t.Helper()
	return bazelPRCallLanes(t, prCallWithFor(t, event))
}

// TestCIScopeDecide: the decision table, from the real script. Any event the
// workflows do not trigger on fails the scope job, and so both gates. The
// script reads the event and nothing else.
func TestCIScopeDecide(t *testing.T) {
	for event, want := range ciScopeRows {
		if got := ciScopeFor(t, event); !reflect.DeepEqual(got, want) {
			t.Errorf("%s decide on %s = %v, want %v", ciScopeScript, event, got, want)
		}
	}
	for _, event := range []string{"", "workflow_dispatch", "schedule", "pull_request_target", "workflow_call", "workflow_run", "PUSH", "merge_group "} {
		if _, out, err := runCIScope(t, map[string]string{"GITHUB_EVENT_NAME": event}, "decide"); err == nil || len(out) != 0 {
			t.Errorf("%s decide on %q: outputs %v, err %v; want a failure and no outputs", ciScopeScript, event, out, err)
		}
	}
	if _, _, err := runCIScope(t, map[string]string{"GITHUB_EVENT_NAME": "pull_request"}, "bogus"); err == nil {
		t.Errorf("%s bogus succeeded; want a usage error", ciScopeScript)
	}
	vars := regexp.MustCompile(`\$\{?([A-Z][A-Z0-9_]*)`)
	for _, line := range strings.Split(readPolicyFile(t, sourceRepoRoot(t), ciScopeScript), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		for _, m := range vars.FindAllStringSubmatch(line, -1) {
			if strings.HasPrefix(m[1], "GITHUB_") && !slices.Contains([]string{"GITHUB_EVENT_NAME", "GITHUB_OUTPUT", "GITHUB_STEP_SUMMARY"}, m[1]) {
				t.Errorf("%s reads %s; the scope is a function of the event alone: %q", ciScopeScript, m[1], line)
			}
		}
		if strings.Contains(line, "${{") {
			t.Errorf("%s holds an expression: %q", ciScopeScript, line)
		}
	}
}

// TestCIScopeJobs: pr.yml's and pr-risk.yml's scope jobs are one job: no needs
// and no if (it always runs), the gates' Blacksmith runner, read-only, only
// the script checked out (sparse, no persisted credentials), the script's
// decide as its one step, and its four outputs.
func TestCIScopeJobs(t *testing.T) {
	wantOutputs := map[string]string{}
	for _, o := range ciScopeOutputs {
		wantOutputs[o] = "${{ steps.decide.outputs." + o + " }}"
	}
	wantCheckout := map[string]string{"persist-credentials": "false", "sparse-checkout": ciScopeScript, "sparse-checkout-cone-mode": "false"}
	var jobs []ciWorkflowJob
	for _, name := range []string{"pr.yml", prRiskWorkflowName} {
		job := readCIWorkflow(t, name).job(t, ciScopeJob)
		jobs = append(jobs, job)
		if job.If != "" || len(job.Needs) != 0 || job.RunsOn != sameRepoBlacksmith2vcpu || job.TimeoutMinutes != 5 || len(job.Env) != 0 || job.ContinueOnError || job.Uses != "" {
			t.Errorf("%s scope job: if %q, needs %v, runs-on %q, timeout %d, env %v, continue-on-error %v, uses %q; want an unconditional 5-minute job on %q",
				name, job.If, job.Needs, job.RunsOn, job.TimeoutMinutes, job.Env, job.ContinueOnError, job.Uses, sameRepoBlacksmith2vcpu)
		}
		if !reflect.DeepEqual(job.Permissions, map[string]any{"contents": "read"}) {
			t.Errorf("%s scope job permissions = %v, want contents: read", name, job.Permissions)
		}
		if !reflect.DeepEqual(job.Outputs, wantOutputs) {
			t.Errorf("%s scope job outputs = %v, want %v", name, job.Outputs, wantOutputs)
		}
		if len(job.Steps) != 2 {
			t.Fatalf("%s scope job has %d steps, want the script's checkout and decide", name, len(job.Steps))
		}
		checkout, decide := job.Steps[0], job.Steps[1]
		if checkout.Uses != "actions/checkout@"+checkoutSHA || !reflect.DeepEqual(checkout.With, wantCheckout) || checkout.If != "" || len(checkout.Env) != 0 {
			t.Errorf("%s scope checkout = %+v; want %s with %v", name, checkout, checkoutSHA, wantCheckout)
		}
		if decide.ID != "decide" || decide.Run != "bash "+ciScopeScript+" decide" || len(decide.Env) != 0 || decide.If != "" || decide.Uses != "" || decide.ContinueOnError != nil {
			t.Errorf("%s scope decide step = %+v; want id decide running %q with no env or if (the script reads GITHUB_EVENT_NAME alone)", name, decide, "bash "+ciScopeScript+" decide")
		}
	}
	if len(jobs) == 2 && !reflect.DeepEqual(jobs[0], jobs[1]) {
		t.Errorf("pr.yml and %s scope jobs differ; both gates must see the same decision:\n%+v\n%+v", prRiskWorkflowName, jobs[0], jobs[1])
	}
}

// Every physical job, including result gates, retains its complete ordered
// dependencies and composed predicate. Scope alone does not describe a job
// whose execution also depends on build artifacts, cohorts or retirement.
func TestCIScopeJobsReadTheScope(t *testing.T) {
	for name, contracts := range scopeJobContracts {
		w := readCIWorkflow(t, name)
		if len(w.Jobs) != len(contracts) {
			t.Errorf("%s job census = %d, want %d", name, len(w.Jobs), len(contracts))
		}
		for jobName, job := range w.Jobs {
			want, ok := contracts[jobName]
			if !ok {
				t.Errorf("%s has unclassified job %s", name, jobName)
				continue
			}
			if !slices.Equal(job.Needs, want.Needs) {
				t.Errorf("%s %s needs = %v, want %v", name, jobName, job.Needs, want.Needs)
			}
			if job.If != want.If {
				t.Errorf("%s %s if = %q, want %q", name, jobName, job.If, want.If)
			}
		}
		for jobName := range contracts {
			if _, ok := w.Jobs[jobName]; !ok {
				t.Errorf("%s missing required job %s", name, jobName)
			}
		}
	}
	if got := readCIWorkflow(t, "pr.yml").job(t, "bazel").With["dolt-tiers"]; got != "${{ needs."+ciScopeJob+".outputs.dolt-tiers }}" {
		t.Errorf("pr.yml bazel call dolt-tiers = %q, want scope output", got)
	}
}

// TestCIScopeGateEnv: both gates require the scope (SCOPE, which their run
// script sets from ci-scope.sh's result), read the scope job's result and
// outputs, and declare per switch exactly the ids it turns off, each one an
// id the gate requires.
func TestCIScopeGateEnv(t *testing.T) {
	for name, idsBySwitch := range ciScopeGateIDs {
		step := readCIWorkflow(t, name).job(t, "ci-gate").step(t, "Evaluate CI gate")
		required := strings.Fields(step.Env["CI_GATE_REQUIRED"])
		if !contains(required, "SCOPE") {
			t.Errorf("%s gate does not require SCOPE", name)
		}
		if _, ok := step.Env["SCOPE"]; ok {
			t.Errorf("%s gate env sets SCOPE; only its run script may (from %s result)", name, ciScopeScript)
		}
		want := map[string]string{"SCOPE_RESULT": "${{ needs." + ciScopeJob + ".result }}"}
		for _, o := range ciScopeOutputs {
			want["SCOPE_"+ciScopeEnvName(o)] = "${{ needs." + ciScopeJob + ".outputs." + o + " }}"
		}
		for k, v := range want {
			if step.Env[k] != v {
				t.Errorf("%s gate env %s = %q, want %q", name, k, step.Env[k], v)
			}
		}
		flagGated := map[string]bool{}
		for _, g := range bazelFlagGatedLanes {
			flagGated[g.id] = true
		}
		// The run script selects exactly one Windows pair and one proof
		// family. Scope skips may list their inactive alternatives too.
		for _, id := range []string{"BAZEL_REMOTE", "BAZEL_LOCAL_FULL", "TEST_WINDOWS_LIVENESS", "WORKTREE_REMOVE_WINDOWS", "TEST_WINDOWS_LIVENESS_PREBUILT", "WORKTREE_REMOVE_WINDOWS_PREBUILT"} {
			flagGated[id] = true
		}
		for _, o := range ciScopeOutputs {
			key := "CI_SCOPE_" + ciScopeEnvName(o) + "_IDS"
			got := strings.Fields(step.Env[key])
			if !reflect.DeepEqual(slices.Sorted(slices.Values(got)), slices.Sorted(slices.Values(idsBySwitch[o]))) {
				t.Errorf("%s gate env %s = %v, want exactly %v", name, key, got, idsBySwitch[o])
			}
			for _, id := range got {
				if !contains(required, id) && !(name == "pr.yml" && flagGated[id]) {
					t.Errorf("%s gate env %s lists %s, which the gate does not require", name, key, id)
				}
			}
		}
		for _, frag := range []string{`SCOPE="$(bash ` + ciScopeScript + ` result)"`, "bash " + ciScopeScript + " skips"} {
			if strings.Count(step.Run, frag) != 1 {
				t.Errorf("%s gate run lacks exactly one %q:\n%s", name, frag, step.Run)
			}
		}
	}
}

// ciGateAllowedSkips: the ids a gate run accepted as skipped (ci-gate.sh's
// "ok: X=skipped (allowed)" lines).
func ciGateAllowedSkips(log string) []string {
	var out []string
	for _, m := range regexp.MustCompile(`(?m)^ok: ([A-Z0-9_]+)=skipped \(allowed\)$`).FindAllStringSubmatch(log, -1) {
		out = append(out, m[1])
	}
	slices.Sort(out)
	return out
}

// prScopeScenario: pr.yml's gate scenario for one run on event with this
// scope: every job the scope runs succeeded and every job it turned off was
// skipped; the Bazel call, when the scope runs it, took mode remote with
// every lane that runs there passing, and otherwise never started.
func prScopeScenario(t *testing.T, event string, scope map[string]string) (bazelGateScenario, map[string]map[string]bool) {
	t.Helper()
	pr := readCIWorkflow(t, "pr.yml")
	ctx := ciScopeCtx(scope)
	sc := bazelGateScenario{name: event, event: event, repository: "gastownhall/beads", scope: scope, results: map[string]string{ciScopeJob: "success"}, flags: map[string]string{}, coverage: "success", coverageSet: true, covered: map[string]string{}}
	// Explicit normal-fork decision facts, separately from job outcomes.
	for _, tier := range retiredTiers {
		sc.covered[tier.output] = "false"
	}
	if event == "merge_group" {
		sc.covered["embedded"] = "true"
	}
	if scope["linux"] != "true" {
		sc.coverage = "skipped"
		sc.covered = map[string]string{}
	}
	for key, value := range sc.covered {
		ctx["needs.bazel-coverage.outputs."+key] = value
	}
	for _, g := range bazelFlagGatedLanes {
		sc.flags[g.flag] = pr.Env[g.flag]
	}
	sc.flags["WINDOWS_PREBUILT_REQUIRED"] = pr.Env["WINDOWS_PREBUILT_REQUIRED"]
	for jobName := range ciScopeJobSwitch["pr.yml"] {
		if jobName == "bazel" {
			continue
		}
		sc.results[jobName] = "skipped"
		if evalRRCIf(t, pr.job(t, jobName).If, ctx) {
			sc.results[jobName] = "success"
		}
	}
	lanes := map[string]map[string]bool{}
	if evalRRCIf(t, pr.job(t, "bazel").If, ctx) {
		sc.callWith = prCallWith(t, pr.job(t, "bazel").With, scope)
		lanes = bazelPRCallLanes(t, sc.callWith)
		sc.mode, sc.enabled, sc.call, sc.outputs = "remote", "true", "success", map[string]string{}
		for lane, modes := range lanes {
			if modes["remote"] {
				sc.outputs[lane] = "success"
			}
		}
	} else {
		sc.call, sc.callSkipped = "skipped", true
	}
	return sc, lanes
}

// prGateSetter: how a scenario sets the result one gate id reads. bazel:
// the id reads a Bazel call output (a lane's, or a package gate's).
type prGateSetter struct {
	bazel bool
	set   func(sc *bazelGateScenario, result string)
}

// prGateSetters: for every id pr.yml's gate reads from a job (not SCOPE,
// BAZEL or BAZEL_REMOTE, which its run script derives), how a scenario sets
// that job's result. A skipped Bazel lane or fast-checks step reports no
// output. A lane that runs in neither arm of its union in this mode "runs
// anyway" as the arm the mode would pick (the union in mode remote).
func prGateSetters(t *testing.T, step ciWorkflowStep, lanes map[string]map[string]bool) map[string]prGateSetter {
	t.Helper()
	re := regexp.MustCompile(`^\$\{\{ needs\.([A-Za-z0-9_-]+)\.(result|outputs\.([A-Za-z0-9_-]+))( \|\| 'skipped')? \}\}$`)
	setters := map[string]prGateSetter{}
	for key, value := range step.Env {
		m := re.FindStringSubmatch(value)
		if m == nil {
			continue
		}
		job, output := m[1], m[3]
		switch {
		case job == ciScopeJob, job == "bazel" && output == "", job == "bazel" && slices.Contains([]string{"rbe-mode", "rbe-enabled", "dolt-tiers", "local-full-intent"}, output):
			continue
		case job == "bazel-coverage" && output != "":
			continue // retirement decision metadata, never a job result
		case job == "bazel-coverage":
			setters[key] = prGateSetter{set: func(sc *bazelGateScenario, r string) { sc.coverage, sc.coverageSet = r, true }}
		case job == "bazel" && output == "local-full-result":
			setters[key] = prGateSetter{bazel: true, set: func(sc *bazelGateScenario, r string) { sc.localResult = r }}
		case job == "bazel":
			setters[key] = prGateSetter{bazel: true, set: func(sc *bazelGateScenario, r string) {
				reporter := bazelLaneReporter(lanes, output, sc.mode)
				if reporter == "" {
					reporter = output
					if union := bazelUnionOf(output); union != "" && sc.mode == "remote" {
						reporter = union
					}
				}
				if r == "skipped" {
					r = ""
				}
				sc.outputs = maps.Clone(sc.outputs)
				if sc.outputs == nil {
					sc.outputs = map[string]string{}
				}
				sc.outputs[reporter] = r
			}}
		case job == "fast-checks" && output != "":
			legacy := fastChecksLegacyJobNames[output]
			setters[key] = prGateSetter{set: func(sc *bazelGateScenario, r string) {
				sc.results = maps.Clone(sc.results)
				if r == "skipped" {
					r = ""
				} else if sc.results["fast-checks"] == "skipped" {
					sc.results["fast-checks"] = r // a step outcome needs a job that ran
				}
				sc.results[legacy] = r
			}}
		default:
			setters[key] = prGateSetter{set: func(sc *bazelGateScenario, r string) {
				sc.results = maps.Clone(sc.results)
				sc.results[job] = r
			}}
		}
	}
	return setters
}

// Effective required IDs include the mutually exclusive dynamic pairs.
// Scope may excuse inactive alternatives, but that does not require them.
func prEffectiveRequired(step ciWorkflowStep, sc bazelGateScenario) []string {
	required := strings.Fields(step.Env["CI_GATE_REQUIRED"])
	if sc.flags["WINDOWS_PREBUILT_REQUIRED"] == "true" {
		required = append(required, "TEST_WINDOWS_LIVENESS_PREBUILT", "WORKTREE_REMOVE_WINDOWS_PREBUILT")
	} else {
		required = append(required, "TEST_WINDOWS_LIVENESS", "WORKTREE_REMOVE_WINDOWS")
	}
	if sc.flags["BAZEL_CMD_DOLT_REQUIRED"] == "true" {
		required = append(required, "BAZEL_CMD_DOLT")
	}
	if sc.localIntent == "true" {
		required = append(required, "BAZEL_LOCAL_FULL")
	} else {
		required = append(required, "BAZEL_REMOTE")
	}
	return required
}

// TestCIScopePRGate runs pr.yml's real gate step over each event with the
// real scope: green with exactly the skips the scope (and, for the Dolt
// tiers, the call's report) excuses: none on a pull request, the Windows and
// macOS jobs and the Dolt-tier lanes on a merge group, the Linux jobs and the
// whole Bazel call on a push to main. Every job that runs and fails, is
// cancelled or goes missing reds it, and so does a job the scope turned off
// that ran anyway and failed (a Dolt-tier lane in a merge group).
func TestCIScopePRGate(t *testing.T) {
	requireHostTool(t, "bash")
	step := readCIWorkflow(t, "pr.yml").job(t, "ci-gate").step(t, "Evaluate CI gate")
	wantSkips := map[string][]string{
		"pull_request": nil,
		"merge_group":  slices.Concat(ciScopeGateIDs["pr.yml"]["platforms"], bazelDoltTierGateIDs),
		"push":         ciScopeGateIDs["pr.yml"]["linux"],
	}
	for _, event := range ciScopeEvents() {
		scope := ciScopeFor(t, event)
		base, lanes := prScopeScenario(t, event, scope)
		required := prEffectiveRequired(step, base)
		pass, log := runPRGateStep(t, step, base)
		if !pass {
			t.Fatalf("%s, every job the scope runs passed: gate red\n%s", event, log)
		}
		var want []string
		for _, id := range wantSkips[event] {
			// With coverage false, retirement derivatives succeed; only
			// the absent raw execution facts consume a scope skip.
			if event == "push" && slices.Contains([]string{"BAZEL_EMBEDDED_RETIRED", "BAZEL_DOLT_SERVER_RETIRED", "BAZEL_PR_LANES_RETIRED"}, id) {
				continue
			}
			if contains(required, id) {
				want = append(want, id)
			}
		}
		slices.Sort(want)
		if got := ciGateAllowedSkips(log); !slices.Equal(got, want) {
			t.Errorf("%s: gate accepted skips %v, want exactly %v\n%s", event, got, want, log)
		}
		skipped := map[string]bool{}
		for _, id := range ciGateAllowedSkips(log) {
			skipped[id] = true
		}
		ids := 0
		setters := prGateSetters(t, step, lanes)
		for _, id := range required {
			if _, ok := setters[id]; !ok && !slices.Contains([]string{"SCOPE", "BAZEL", "BAZEL_REMOTE", "BAZEL_EMBEDDED_RETIRED", "BAZEL_DOLT_SERVER_RETIRED", "BAZEL_PR_LANES_RETIRED"}, id) {
				t.Errorf("%s: required mutable result %s has no setter", event, id)
			}
		}
		for id, setter := range setters {
			if !contains(required, id) {
				continue // advisory
			}
			if setter.bazel && base.callSkipped {
				continue // no lane can run in a call that never started
			}
			ids++
			results := []string{"failure", "cancelled"}
			if !skipped[id] {
				results = append(results, "skipped") // ran, and must not go missing
			}
			for _, r := range results {
				sc := base
				setter.set(&sc, r)
				if pass, log := runPRGateStep(t, step, sc); pass || !regexp.MustCompile(`::error::`+id+`\b`).MatchString(log) {
					t.Errorf("%s, %s %s (scope skipped it: %v): gate pass = %v, want red naming %s\n%s", event, id, r, skipped[id], pass, id, log)
				}
			}
		}
		// pull_request and merge_group: every job and lane id; push: the
		// Windows and macOS jobs and fast-checks' and check-doc-flags'.
		if min := map[string]int{"pull_request": 18, "merge_group": 18, "push": 9}[event]; ids < min {
			t.Errorf("%s: mutated %d gate ids, want at least %d; the setter scan is broken", event, ids, min)
		}
		// The scope job's own failure, or a broken output, reds the gate
		// however the rest went.
		for _, c := range []struct {
			name   string
			result string
			scope  map[string]string
		}{
			{"scope failed", "failure", nil},
			{"scope cancelled", "cancelled", nil},
			{"scope skipped", "skipped", nil},
			{"scope succeeded without outputs", "success", map[string]string{}},
			{"scope output upper-cased", "success", withScope(scope, "linux", strings.ToUpper(scope["linux"]))},
			{"scope output not a row", "success", withScope(scope, "platforms", map[string]string{"true": "false", "false": "true"}[scope["platforms"]])},
			{"scope output missing a switch", "success", withScope(scope, "advisory", "")},
		} {
			sc := base
			sc.results = maps.Clone(base.results)
			sc.results[ciScopeJob] = c.result
			sc.scope = c.scope
			if c.scope == nil {
				sc.scope = map[string]string{}
			}
			if pass, log := runPRGateStep(t, step, sc); pass || !regexp.MustCompile(`::error::SCOPE\b`).MatchString(log) {
				t.Errorf("%s, %s: gate pass = %v, want red naming SCOPE\n%s", event, c.name, pass, log)
			}
		}
	}

	// The Dolt-tier skip needs both halves: the scope turned the tiers off
	// and the call reports it ran without them.
	mg, _ := prScopeScenario(t, "merge_group", ciScopeFor(t, "merge_group"))
	mg.callWith = maps.Clone(mg.callWith)
	mg.callWith["dolt-tiers"] = "on" // a call that ran the Dolt tiers, yet they reported nothing
	if pass, log := runPRGateStep(t, step, mg); pass || !regexp.MustCompile(`::error::BAZEL_EMBEDDED\b`).MatchString(log) {
		t.Errorf("merge group, the call reports dolt-tiers on but the Dolt-tier lanes reported nothing: gate pass = %v, want red\n%s", pass, log)
	}
	prRun, _ := prScopeScenario(t, "pull_request", ciScopeFor(t, "pull_request"))
	prRun.callWith = maps.Clone(prRun.callWith)
	prRun.callWith["dolt-tiers"] = "off"
	for lane := range prRun.outputs {
		if slices.Contains(bazelDoltTierLanes, lane) {
			delete(prRun.outputs, lane)
		}
	}
	if pass, log := runPRGateStep(t, step, prRun); pass || !regexp.MustCompile(`::error::BAZEL_EMBEDDED\b`).MatchString(log) {
		t.Errorf("pull request, the call ran without the Dolt tiers: gate pass = %v, want red (the scope runs them on a PR)\n%s", pass, log)
	}
}

// withScope: scope with one output replaced ("" deletes it).
func withScope(scope map[string]string, key, value string) map[string]string {
	out := maps.Clone(scope)
	if value == "" {
		delete(out, key)
	} else {
		out[key] = value
	}
	return out
}

// ciScopeEvents: the events the two workflows run on, in a fixed order.
func ciScopeEvents() []string { return []string{"pull_request", "merge_group", "push"} }

// runRiskGateStep runs pr-risk.yml's real gate step with the jobs' results
// and explicit output facts. Missing results stay missing and fail closed.
func runRiskGateStep(t *testing.T, step ciWorkflowStep, event string, results, scope map[string]string, outputs map[string]map[string]string) (bool, string) {
	t.Helper()
	re := regexp.MustCompile(`^\$\{\{ needs\.([A-Za-z0-9_-]+)\.(result|outputs\.([A-Za-z0-9_-]+)) \}\}$`)
	env := []string{"PATH=" + os.Getenv("PATH"), "GITHUB_EVENT_NAME=" + event}
	for key, value := range step.Env {
		if !strings.Contains(value, "${{") {
			env = append(env, key+"="+value)
			continue
		}
		m := re.FindStringSubmatch(value)
		if m == nil {
			t.Fatalf("pr-risk.yml gate env %s = %q: the simulation cannot evaluate it", key, value)
		}
		got := ""
		switch {
		case m[2] == "result":
			got = results[m[1]]
		case m[1] == ciScopeJob:
			switch m[3] {
			case "platforms", "dolt-tiers", "linux", "advisory":
				got = scope[m[3]]
			default:
				t.Fatalf("unknown Risk scope output %q", m[3])
			}
		case m[1] == "detect-ci-tier" && m[3] == "full_embedded":
			got = outputs[m[1]][m[3]]
		case m[1] == "bazel-coverage" && (m[3] == "embedded" || m[3] == "dolt_server"):
			got = outputs[m[1]][m[3]]
		default:
			t.Fatalf("pr-risk.yml gate env %s = %q: the simulation cannot evaluate it", key, value)
		}
		env = append(env, key+"="+got)
	}
	cmd := exec.Command("bash", "--noprofile", "--norc", "-eo", "pipefail", "-c", step.Run)
	cmd.Dir = sourceRepoRoot(t)
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	return err == nil, string(out)
}

var riskEngineJobIDs = map[string]string{
	"build-embedded":            "BUILD_EMBEDDED",
	"test-embedded-storage":     "TEST_EMBEDDED_STORAGE",
	"test-embedded-conformance": "TEST_EMBEDDED_CONFORMANCE",
	"test-server-storage":       "TEST_SERVER_STORAGE",
	"test-server-storage-full":  "TEST_SERVER_STORAGE_FULL",
	"test-embedded-cmd":         "TEST_EMBEDDED_CMD",
	"test-proxied-cmd":          "TEST_PROXIED_CMD",
}

// Normal-fork fixtures pin explicit committed retirement semantics. These
// rows are complete; a mutation deleting a result cannot inherit success.
func riskScopeScenario(event string) (map[string]string, map[string]map[string]string) {
	results := map[string]string{ciScopeJob: "success", "detect-ci-tier": "success", "bazel-coverage": "success", "test-nix": "success"}
	outputs := map[string]map[string]string{
		"detect-ci-tier": {"full_embedded": "true"},
		"bazel-coverage": {"embedded": "false", "dolt_server": "false"},
	}
	for job := range riskEngineJobIDs {
		results[job] = "success"
	}
	switch event {
	case "pull_request":
	case "merge_group":
		for job := range riskEngineJobIDs {
			results[job] = "skipped"
		}
		results["test-nix"] = "skipped"
		outputs["bazel-coverage"]["embedded"] = "true"
	case "push":
		for job := range riskEngineJobIDs {
			results[job] = "skipped"
		}
		results["detect-ci-tier"], results["bazel-coverage"] = "skipped", "skipped"
		outputs = map[string]map[string]string{}
	default:
		panic("unknown Risk scope event: " + event)
	}
	return results, outputs
}

func TestCIScopeRiskGate(t *testing.T) {
	requireHostTool(t, "bash")
	w := readCIWorkflow(t, prRiskWorkflowName)
	step := w.job(t, "ci-gate").step(t, "Evaluate CI gate")
	required := []string{"SCOPE", "DETECT_CI_TIER", "BAZEL_COVERAGE", "BUILD_EMBEDDED", "TEST_EMBEDDED_STORAGE", "TEST_EMBEDDED_CONFORMANCE", "TEST_SERVER_STORAGE", "TEST_SERVER_STORAGE_FULL", "TEST_EMBEDDED_CMD", "TEST_PROXIED_CMD", "TEST_NIX"}
	if got := strings.Fields(step.Env["CI_GATE_REQUIRED"]); !slices.Equal(got, required) {
		t.Errorf("Risk required = %v, want %v", got, required)
	}
	for _, event := range ciScopeEvents() {
		scope := ciScopeFor(t, event)
		results, outputs := riskScopeScenario(event)
		pass, log := runRiskGateStep(t, step, event, results, scope, outputs)
		if !pass {
			t.Errorf("%s: normal Risk fixture red\n%s", event, log)
		}
		var wantSkips []string
		if event != "pull_request" {
			for _, id := range riskEngineJobIDs {
				wantSkips = append(wantSkips, id)
			}
		}
		switch event {
		case "merge_group":
			wantSkips = append(wantSkips, "TEST_NIX")
		case "push":
			wantSkips = append(wantSkips, "DETECT_CI_TIER", "BAZEL_COVERAGE")
		}
		slices.Sort(wantSkips)
		if got := ciGateAllowedSkips(log); !slices.Equal(got, wantSkips) {
			t.Errorf("%s skips = %v, want %v", event, got, wantSkips)
		}
		// Every actual job has an explicit independently failing result, even
		// when its skip is permitted by a validated scope/retirement decision.
		for job := range results {
			id := "SCOPE"
			if job != ciScopeJob {
				for _, candidate := range required {
					if step.Env[candidate] == "${{ needs."+job+".result }}" {
						id = candidate
						break
					}
				}
				if id == "SCOPE" {
					t.Fatalf("unmapped required Risk result %s", job)
				}
			}
			for _, result := range []string{"failure", "cancelled", "", "skipped"} {
				if result == "skipped" && results[job] == "skipped" {
					continue
				}
				broken := maps.Clone(results)
				broken[job] = result
				if pass, log := runRiskGateStep(t, step, event, broken, scope, outputs); pass || !strings.Contains(log, "::error::"+id) {
					t.Errorf("%s %s=%q: pass%v, want red naming%s\n%s", event, job, result, pass, id, log)
				}
			}
		}
		for name, c := range map[string]struct {
			result string
			scope  map[string]string
		}{
			"scope failed":            {"failure", map[string]string{}},
			"scope without outputs":   {"success", map[string]string{}},
			"scope output not a row":  {"success", withScope(scope, "dolt-tiers", map[string]string{"on": "off", "off": "on"}[scope["dolt-tiers"]])},
			"scope output upper-case": {"success", withScope(scope, "platforms", strings.ToUpper(scope["platforms"]))},
		} {
			broken := maps.Clone(results)
			broken[ciScopeJob] = c.result
			if pass, log := runRiskGateStep(t, step, event, broken, c.scope, outputs); pass || !strings.Contains(log, "::error::SCOPE") {
				t.Errorf("%s %s pass%v, want SCOPE red\n%s", event, name, pass, log)
			}
		}
	}
	// Docs-only PRs skip engines for an explicit successful detector result;
	// Nix, scope and the coverage decision still need actual success.
	results, outputs := riskScopeScenario("pull_request")
	outputs["detect-ci-tier"]["full_embedded"] = "false"
	for job := range riskEngineJobIDs {
		results[job] = "skipped"
	}
	if pass, log := runRiskGateStep(t, step, "pull_request", results, ciScopeFor(t, "pull_request"), outputs); !pass {
		t.Errorf("docs-only Risk fixture red\n%s", log)
	}
}

// TestCIScopeSkipsNeedARow: ci-scope.sh's result and skips accept only an
// exact row of the decision table: anything else is an invalid scope and
// excuses nothing.
func TestCIScopeSkipsNeedARow(t *testing.T) {
	ids := map[string]string{"CI_SCOPE_PLATFORMS_IDS": "P1 P2", "CI_SCOPE_LINUX_IDS": "L1"}
	env := func(result string, scope map[string]string) map[string]string {
		e := maps.Clone(ids)
		e["SCOPE_RESULT"] = result
		for _, o := range ciScopeOutputs {
			e["SCOPE_"+ciScopeEnvName(o)] = scope[o]
		}
		return e
	}
	wantSkips := map[string]string{"pull_request": "", "merge_group": "P1 P2", "push": "L1"}
	for _, event := range ciScopeEvents() {
		scope := ciScopeRows[event]
		if out, _, err := runCIScope(t, env("success", scope), "result"); err != nil || out != "success" {
			t.Errorf("%s row: result %q (%v), want success", event, out, err)
		}
		if out, _, err := runCIScope(t, env("success", scope), "skips"); err != nil || out != wantSkips[event] {
			t.Errorf("%s row: skips %q (%v), want %q", event, out, err, wantSkips[event])
		}
		for _, r := range []string{"failure", "cancelled", "skipped", ""} {
			if out, _, err := runCIScope(t, env(r, scope), "result"); err != nil || out != r {
				t.Errorf("%s row, scope job %q: result %q (%v), want the job's result", event, r, out, err)
			}
		}
	}
	for name, scope := range map[string]map[string]string{
		"empty":                        {},
		"platforms and linux both off": {"platforms": "false", "dolt-tiers": "off", "linux": "false", "advisory": "false"},
		"upper case":                   {"platforms": "FALSE", "dolt-tiers": "off", "linux": "true", "advisory": "false"},
		"padded":                       {"platforms": "false ", "dolt-tiers": "off", "linux": "true", "advisory": "false"},
		"injected":                     {"platforms": "false\n::error::x", "dolt-tiers": "off", "linux": "true", "advisory": "false"},
		"missing advisory":             {"platforms": "false", "dolt-tiers": "off", "linux": "true"},
	} {
		out, _, err := runCIScope(t, env("success", scope), "result")
		if err != nil || !regexp.MustCompile(`^invalid-scope:[A-Za-z0-9=,-]*$`).MatchString(out) {
			t.Errorf("%s: result %q (%v), want one invalid-scope:<sanitized> line", name, out, err)
		}
		if out, _, err := runCIScope(t, env("success", scope), "skips"); err != nil || out != "" {
			t.Errorf("%s: skips %q (%v), want none", name, out, err)
		}
	}
}

// TestBazelGateDoltTiersCrossCheck: bazel-gate.sh excuses the Dolt-tier
// lanes only when the scope turned the tiers off and the call reports it ran
// without them, in a valid mode; a call that never started (pr.yml's push run)
// reads skipped for its aggregate and its remote check, which only the
// scope's own skips may excuse.
func TestBazelGateDoltTiersCrossCheck(t *testing.T) {
	requireHostTool(t, "bash")
	run := func(arg string, env map[string]string) string {
		t.Helper()
		cmd := exec.Command("bash", bazelGateScript, arg)
		cmd.Dir = sourceRepoRoot(t)
		cmd.Env = []string{"PATH=" + os.Getenv("PATH")}
		for k, v := range env {
			cmd.Env = append(cmd.Env, k+"="+v)
		}
		out, err := cmd.Output()
		if err != nil {
			t.Fatalf("%s %s %v: %v", bazelGateScript, arg, env, err)
		}
		return strings.TrimSpace(string(out))
	}
	for _, mode := range bazelRBEModes {
		base := run("skips", map[string]string{"BAZEL_RBE_MODE": mode, "BAZEL_RBE_ENABLED": bazelModeEnabled(mode)})
		for _, scopeDolt := range []string{"on", "off", ""} {
			for _, callDolt := range []string{"on", "off", ""} {
				got := strings.Fields(run("skips", map[string]string{"BAZEL_RBE_MODE": mode, "BAZEL_RBE_ENABLED": bazelModeEnabled(mode), "SCOPE_DOLT_TIERS": scopeDolt, "BAZEL_DOLT_TIERS": callDolt}))
				want := strings.Fields(base)
				if scopeDolt == "off" && callDolt == "off" {
					for _, id := range bazelDoltTierGateIDs {
						if !slices.Contains(want, id) {
							want = append(want, id)
						}
					}
				}
				slices.Sort(got)
				slices.Sort(want)
				if !slices.Equal(slices.Compact(got), want) {
					t.Errorf("mode %s, scope dolt-tiers %q, call dolt-tiers %q: skips %v, want %v", mode, scopeDolt, callDolt, got, want)
				}
			}
		}
	}
	if got := run("skips", map[string]string{"BAZEL_RBE_MODE": "bogus", "BAZEL_RBE_ENABLED": "true", "SCOPE_DOLT_TIERS": "off", "BAZEL_DOLT_TIERS": "off"}); got != "" {
		t.Errorf("invalid mode with the Dolt tiers off: skips %q, want none", got)
	}
	for _, c := range []struct {
		name, call, mode, enabled string
		aggregate, remote         string
	}{
		{"call never started", "skipped", "", "", "skipped", "skipped"},
		{"rbe job failed", "failure", "", "", "invalid-rbe-mode:/", "not-run-remotely"},
		{"call cancelled before the decision", "cancelled", "", "", "invalid-rbe-mode:/", "not-run-remotely"},
		{"remote", "success", "remote", "true", "success", "success"},
		{"fork-ro", "success", "fork-ro", "true", "success", "success"},
		{"cache", "success", "cache", "false", "success", "not-run-remotely"},
		{"skip", "success", "skip", "false", "success", "not-run-remotely"},
		{"inconsistent", "success", "remote", "false", "invalid-rbe-mode:remote/false", "not-run-remotely"},
		{"skipped with a mode", "skipped", "remote", "true", "skipped", "success"},
	} {
		env := map[string]string{"BAZEL_CALL": c.call, "BAZEL_RBE_MODE": c.mode, "BAZEL_RBE_ENABLED": c.enabled}
		if got := run("aggregate", env); got != c.aggregate {
			t.Errorf("%s: aggregate %q, want %q", c.name, got, c.aggregate)
		}
		if got := run("remote", env); got != c.remote {
			t.Errorf("%s: remote %q, want %q", c.name, got, c.remote)
		}
		if c.name == "call never started" {
			if got := run("skips", env); got != "" {
				t.Errorf("%s: skips %q, want none (the scope excuses what it turned off)", c.name, got)
			}
		}
	}
}

// TestBazelDoltTiersInput: bazel.yml's dolt-tiers input (a string, default
// "on": push, dispatch, nightly.yml and bazel-farm.yml keep every lane) turns
// off exactly the ordinary Dolt-tier lanes when it is "off"; Q requires the
// dedicated replacement regardless. No test step may filter tiers; nothing
// else reads it; the call reports what it did as its dolt-tiers output, which
// agrees with the lanes for every value; and only pr.yml passes it, from the
// scope.
func TestBazelDoltTiersInput(t *testing.T) {
	call := readBazelWorkflowCall(t)
	if in, ok := call.Inputs["dolt-tiers"]; !ok || in.Type != "string" || in.Default != "on" {
		t.Errorf("%s workflow_call input dolt-tiers = %+v (present %v), want type string, default on", bazelWorkflowName, in, ok)
	}
	if got := call.Outputs["dolt-tiers"].Value; got != bazelDoltTiersOutput {
		t.Errorf("%s workflow_call output dolt-tiers = %q, want %q", bazelWorkflowName, got, bazelDoltTiersOutput)
	}
	workflow := readCIWorkflow(t, bazelWorkflowName)
	suffix := " && " + bazelDoltTiersCond + " }}"
	for name, job := range workflow.Jobs {
		dolt := slices.Contains(bazelDoltTierLanes, name)
		ordinaryIf := job.If
		if body, ok := strings.CutPrefix(ordinaryIf, "${{ needs.rbe.outputs.local_full == 'true' || ("); ok {
			if !slices.Contains([]string{bazelEmbedJobName, bazelIntegJobName, bazelDoltJobName, bazelProxiedJobName, bazelServerJobName, bazelCmdDoltJobName}, name) {
				t.Fatalf("unexpected Q alternate: %s", name)
			}
			body, ok = strings.CutSuffix(body, ") }}")
			if !ok {
				t.Fatalf("malformed Q alternate: %s", ordinaryIf)
			}
			ordinaryIf = "${{ " + body + " }}"
		}
		if got := strings.HasSuffix(ordinaryIf, suffix); got != dolt {
			t.Errorf("%s if %q ends with the Dolt-tier switch: %v, want %v", name, job.If, got, dolt)
		}
		if strings.Count(job.If, "inputs.dolt-tiers") > btoi(dolt) {
			t.Errorf("%s if %q reads inputs.dolt-tiers outside the switch", name, job.If)
		}
		for _, s := range job.Steps {
			if strings.Contains(fmt.Sprint(s.If, s.Run, s.Env, s.With), "dolt-tiers") {
				t.Errorf("%s step %q reads dolt-tiers; only the lanes' if may", name, s.Name)
			}
		}
		if strings.Contains(fmt.Sprint(job.Env, job.RunsOn, job.Outputs), "dolt-tiers") {
			t.Errorf("%s reads dolt-tiers outside its if", name)
		}
	}
	for _, name := range bazelDoltTierLanes {
		if _, ok := workflow.Jobs[name]; !ok {
			t.Errorf("bazelDoltTierLanes lists %s, which %s does not have", name, bazelWorkflowName)
		}
	}
	// The lanes' switch and the reported output agree on every value a
	// caller could pass (GitHub's == is case-insensitive).
	for _, v := range []any{nil, "", "on", "off", "OFF", "Off", "no", "false", false, true} {
		ctx := map[string]any{"inputs.dolt-tiers": v, "jobs.rbe.outputs.local_full": "false"}
		runs, err := evalGHExprTyped("${{ "+bazelDoltTiersCond+" }}", ctx)
		if err != nil {
			t.Fatal(err)
		}
		out, err := evalGHExprTyped(bazelDoltTiersOutput, ctx)
		if err != nil {
			t.Fatal(err)
		}
		if (out == "on") != ghTruthy(runs) || (out != "on" && out != "off") {
			t.Errorf("dolt-tiers %#v: lanes run %v, the call reports %q", v, runs, out)
		}
		ctx["jobs.rbe.outputs.local_full"] = "true"
		if qout, err := evalGHExprTyped(bazelDoltTiersOutput, ctx); err != nil || qout != "on" {
			t.Errorf("Q dolt-tiers %#v reports%v/%v, want on", v, qout, err)
		}
	}
	// In every mode "off" stops exactly the Dolt-tier lanes that would run.
	for _, mode := range bazelRBEModes {
		for name, job := range workflow.Jobs {
			if name == bazelRBEJobName || isBazelRRCJob(name) || name == bazelDispatchCoverageJobName || name == "bazel-local-coverage" {
				continue
			}
			for _, with := range []map[string]string{{}, {"integration": "on", "package-gates": "on"}} {
				on := bazelLaneRunModes(t, name, job.If, with)[mode]
				offWith := maps.Clone(with)
				offWith["dolt-tiers"] = "off"
				off := bazelLaneRunModes(t, name, job.If, offWith)[mode]
				if want := on && !slices.Contains(bazelDoltTierLanes, name); off != want {
					t.Errorf("mode %s, %s (with %v): runs with dolt-tiers off = %v, want %v", mode, name, with, off, want)
				}
			}
		}
	}
	// Only pr.yml passes it, and only the scope's switch.
	entries, err := os.ReadDir(filepath.Join(sourceRepoRoot(t), ".github", "workflows"))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".yml") || e.Name() == bazelWorkflowName {
			continue
		}
		for jobName, job := range readCIWorkflow(t, e.Name()).Jobs {
			v, ok := job.With["dolt-tiers"]
			if !ok || !strings.HasSuffix(job.Uses, "/"+bazelWorkflowName) {
				continue
			}
			if e.Name() != "pr.yml" || jobName != "bazel" || v != "${{ needs."+ciScopeJob+".outputs.dolt-tiers }}" {
				t.Errorf("%s job %s passes dolt-tiers %q to %s; only pr.yml's call may, from its scope", e.Name(), jobName, v, bazelWorkflowName)
			}
		}
	}
}

// Legacy Risk engine definitions defer on merge groups and push. Their
// surface is carried by these actual dedicated replacement results; a build
// artifact is useful only when every consumer tier is actually covered.
var riskReplacementIDs = map[string][]string{
	"BUILD_EMBEDDED":            {"BAZEL_EMBEDDED", "BAZEL_PROXIED", "BAZEL_SERVER_STORAGE", "BAZEL_CMD_DOLT_DIRECT"},
	"TEST_EMBEDDED_STORAGE":     {"BAZEL_EMBEDDED"},
	"TEST_EMBEDDED_CONFORMANCE": {"BAZEL_EMBEDDED"},
	"TEST_EMBEDDED_CMD":         {"BAZEL_EMBEDDED"},
	"TEST_SERVER_STORAGE":       {"BAZEL_SERVER_STORAGE"},
	"TEST_SERVER_STORAGE_FULL":  {"BAZEL_SERVER_STORAGE"},
	"TEST_PROXIED_CMD":          {"BAZEL_PROXIED"},
}

func TestForkRiskReplacementRequiresActualCompleteLocalResults(t *testing.T) {
	requireHostTool(t, "bash")
	for legacy, replacements := range riskReplacementIDs {
		for _, id := range replacements {
			if !slices.Contains(fullLocalRequiredResults, id) {
				t.Fatalf("%s replacement %s is not a full-local required result", legacy, id)
			}
			for _, result := range []string{"failure", "cancelled", "skipped", ""} {
				facts := fullLocalGateFacts()
				facts["SCOPE_DOLT_TIERS"] = "off" // deferral cannot excuse replacement
				facts[id] = result
				if out, err := runFullLocalGate(t, "local-full", facts); err == nil || !strings.Contains(out, id+"=") {
					t.Errorf("%s replacement%s=%q did not fail complete local gate: %v\n%s", legacy, id, result, err, out)
				}
			}
		}
	}
	for _, id := range riskEngineJobIDs {
		if len(riskReplacementIDs[id]) == 0 {
			t.Errorf("Risk engine%s has no explicit replacement", id)
		}
	}
}

// Q is a workflow-owned selection fact: caller fast-set switches cannot
// suppress any dedicated tier, and no remote union can stand in for one.
func TestForkFullLocalSchedulesAllDedicatedTiersDespiteFastSetInputs(t *testing.T) {
	workflow := readCIWorkflow(t, bazelWorkflowName)
	want := []string{bazelJobName, bazelPureJobName, bazelReleaseCrossJobName, bazelEmbedJobName, bazelIntegJobName, bazelDoltJobName, bazelProxiedJobName, bazelServerJobName, bazelCmdDoltJobName, "bazel-server-storage-hosted"}
	slices.Sort(want)
	var got []string
	with := map[string]string{"dolt-tiers": "off", "integration": "off", "package-gates": "off"}
	for name, job := range workflow.Jobs {
		if name == bazelRBEJobName || isBazelRRCJob(name) || name == bazelDispatchCoverageJobName || name == "bazel-local-coverage" {
			continue
		}
		if bazelLaneRunModesWithLocalProof(t, name, job.If, with, true)["local"] {
			got = append(got, name)
		}
		if slices.Contains(bazelDoltTierLanes, name) && bazelLaneRunModesWithLocalProof(t, name, job.If, with, false)["local"] {
			t.Errorf("non-Q fast-set off still runs%s", name)
		}
	}
	slices.Sort(got)
	if !slices.Equal(got, want) {
		t.Errorf("Q dedicated schedule%v, want exact%v", got, want)
	}
}

// TestCIScopeCoversEveryGatedJobOnMain is the invariant the fast set rests
// on: every job pr.yml's and pr-risk.yml's gates read on a pull request runs
// on every push head on main, in the merge group of the batch's last entry
// (main fast-forwards to that tested merge-group commit) or in main's push
// runs: pr.yml's and pr-risk.yml's (the Windows, macOS and Nix jobs) and
// bazel.yml's own push run (every Bazel lane: it has no inputs, so neither
// the Dolt tiers nor the integration build's //cmd/bd:bd_test are turned
// off). The merge group runs the fast set, the push runs the rest (and
// TestCIScopeExcusedIDsRedTheirPushRun: turn red when it fails). Jobs the
// gates do not read (advisory-reports) are not covered: pull requests only.
func TestCIScopeCoversEveryGatedJobOnMain(t *testing.T) {
	type where struct{ mergeGroup, push bool }
	runs := map[string]where{} // gate id -> where it runs after the PR run
	for name := range ciScopeJobSwitch {
		w := readCIWorkflow(t, name)
		step := w.job(t, "ci-gate").step(t, "Evaluate CI gate")
		knownIDs := strings.Fields(step.Env["CI_GATE_REQUIRED"])
		for _, id := range []string{"TEST_WINDOWS_LIVENESS", "WORKTREE_REMOVE_WINDOWS", "TEST_WINDOWS_LIVENESS_PREBUILT", "WORKTREE_REMOVE_WINDOWS_PREBUILT"} {
			knownIDs = append(knownIDs, id)
		}
		var mgResults, pushResults map[string]string
		if name == "pr.yml" {
			mg, _ := prScopeScenario(t, "merge_group", ciScopeRows["merge_group"])
			push, _ := prScopeScenario(t, "push", ciScopeRows["push"])
			mgResults, pushResults = mg.results, push.results
		} else {
			mgResults, _ = riskScopeScenario("merge_group")
			pushResults, _ = riskScopeScenario("push")
		}
		for key, value := range step.Env {
			if !contains(knownIDs, key) {
				continue
			} // exclude decision metadata
			m := regexp.MustCompile(`^\$\{\{ needs\.([a-z0-9-]+)\.(result|outputs\.[a-z0-9-]+)`).FindStringSubmatch(value)
			if m == nil || m[1] == ciScopeJob || m[1] == "bazel" {
				continue
			}
			runs[key] = where{mgResults[m[1]] == "success", pushResults[m[1]] == "success"}
		}
	}
	// The Bazel lanes, mode remote: pr.yml's merge-group call, and
	// bazel.yml's own push run (no inputs at all: every input reads null).
	mgLanes := bazelPRCallLanesFor(t, "merge_group")
	pushLanes := map[string]map[string]bool{}
	for name, job := range readCIWorkflow(t, bazelWorkflowName).Jobs {
		if name != bazelRBEJobName && !isBazelRRCJob(name) && name != bazelDispatchCoverageJobName && name != "bazel-local-coverage" {
			pushLanes[name] = bazelLaneRunModes(t, name, job.If, map[string]string{})
		}
	}
	bazelIDs := maps.Clone(bazelLaneGateIDs)
	for lane, g := range bazelFlagGatedLanes {
		bazelIDs[lane] = g.id
	}
	for lane, id := range bazelIDs {
		runs[id] = where{bazelLaneReporter(mgLanes, lane, "remote") != "", bazelLaneReporter(pushLanes, lane, "remote") != ""}
	}
	// Risk's raw legacy jobs intentionally do not run in either complement.
	// Verify their dedicated replacement surfaces rather than invent a run.
	for legacy, replacements := range riskReplacementIDs {
		covered := true
		for _, replacement := range replacements {
			id := replacement
			if id == "BAZEL_CMD_DOLT_DIRECT" {
				id = "BAZEL_CMD_DOLT"
			}
			covered = covered && runs[id].push
		}
		runs[legacy] = where{false, covered}
	}
	for id, w := range runs {
		if !w.mergeGroup && !w.push {
			t.Errorf("%s runs in neither the merge group nor main's push runs: a push head on main would never run it", id)
		}
	}
	// Where each set runs, spelled out.
	for _, id := range slices.Concat(ciScopeGateIDs["pr.yml"]["platforms"], ciScopeGateIDs[prRiskWorkflowName]["platforms"], bazelDoltTierGateIDs) {
		if w := runs[id]; w.mergeGroup || !w.push {
			t.Errorf("%s: merge group %v, push %v; want the push run only (the complement)", id, w.mergeGroup, w.push)
		}
	}
	for _, id := range []string{"FAST_CHECKS", "CHECK_MIGRATION_HYGIENE", "CHECK_NO_BEADS_CHANGES", "CHECK_DOC_FLAGS", "PACKAGE_MCP", "PACKAGE_NPM", "BAZEL_TEST", "BAZEL_PURE", "BAZEL_RELEASE_CROSS", "BAZEL_INTEGRATION"} {
		if w := runs[id]; !w.mergeGroup {
			t.Errorf("%s does not run in the merge group: it is the fast set", id)
		}
	}
	// The integration build's //cmd/bd:bd_test: pr.yml's call leaves it to
	// the cmd/bd Dolt-server tier (cmd-dolt-required), which a merge group
	// skips, so the merge group runs it nowhere; the push run runs it in both
	// (no cmd-dolt-required: bazel-integration keeps it, and the tier runs).
	exclude := readCIWorkflow(t, bazelWorkflowName).job(t, bazelIntegJobName).step(t, "bazel test //... --config=integration").Env["EXCLUDE_TARGETS"]
	for _, c := range []struct {
		name string
		ctx  map[string]any
		want string
	}{
		{"merge group call", map[string]any{"inputs.cmd-dolt-required": true}, "-//cmd/bd:bd_test"},
		{"push run", map[string]any{}, ""},
	} {
		if got, err := evalGHExprTyped(exclude, c.ctx); err != nil || got != c.want {
			t.Errorf("%s: bazel-integration EXCLUDE_TARGETS = %#v (%v), want %q", c.name, got, err, c.want)
		}
	}
	if with := prCallWithFor(t, "merge_group"); with["cmd-dolt-required"] != "true" || bazelLaneReporter(mgLanes, bazelCmdDoltJobName, "remote") != "" {
		t.Errorf("merge group call: cmd-dolt-required %q, cmd/bd tier reporter %q; want bd_test left to the push run", with["cmd-dolt-required"], bazelLaneReporter(mgLanes, bazelCmdDoltJobName, "remote"))
	}
	if bazelLaneReporter(pushLanes, bazelCmdDoltJobName, "remote") == "" {
		t.Error("bazel.yml's push run does not run the cmd/bd Dolt-server tier")
	}
	// pr.yml never calls bazel.yml on a push: bazel.yml's own push run is
	// every lane, and its rrc-seed mints only for a main-push bazel.yml token.
	if evalRRCIf(t, readCIWorkflow(t, "pr.yml").job(t, "bazel").If, ciScopeCtx(ciScopeRows["push"])) {
		t.Error("pr.yml calls bazel.yml on a push to main; bazel.yml's own push run covers the lanes")
	}
	var on struct {
		On map[string]yaml.Node `yaml:"on"`
	}
	if err := yaml.Unmarshal([]byte(readPolicyFile(t, sourceRepoRoot(t), ".github/workflows/"+bazelWorkflowName)), &on); err != nil {
		t.Fatal(err)
	}
	var push struct {
		Branches []string `yaml:"branches"`
	}
	if node, ok := on.On["push"]; !ok || node.Decode(&push) != nil || !slices.Equal(push.Branches, []string{"main"}) {
		t.Errorf("%s push trigger = %v; want push to main (the Bazel half of the complement)", bazelWorkflowName, push.Branches)
	}
}

// TestCIScopeExcusedIDsRedTheirPushRun: every gate id a merge group's gates
// excuse (read from the gates' own merge-group runs) turns main's push run red
// when its job fails there, which main-ci-alert.yml reports; running on the
// push is not enough.
//   - pr.yml's and pr-risk.yml's own jobs (the Windows, macOS and Nix jobs):
//     the job runs on the push, carries no job-level continue-on-error (its
//     failure fails the run), and the push run's real gate step goes red on
//     its failure.
//   - bazel.yml's Dolt-tier lanes, in its own push run (mode remote, no
//     inputs): the job reporting the id runs and carries no job-level
//     continue-on-error, and the tier's failure fails it. A lane's job.status
//     is its result; bazel-dolt-int's per-tier results fail it through the
//     classify step's --gate-tier (server-storage) or through a later step
//     that exits non-zero on the tier's failure outside pr.yml's call, and
//     only there: in pr.yml's call (pull_request, merge_group) the gate reads
//     the per-tier output, so a PR shows the failure exactly as before.
func TestCIScopeExcusedIDsRedTheirPushRun(t *testing.T) {
	requireHostTool(t, "bash")
	pr := readCIWorkflow(t, "pr.yml")
	prStep := pr.job(t, "ci-gate").step(t, "Evaluate CI gate")
	risk := readCIWorkflow(t, prRiskWorkflowName)
	riskStep := risk.job(t, "ci-gate").step(t, "Evaluate CI gate")
	mgScope, pushScope := ciScopeFor(t, "merge_group"), ciScopeFor(t, "push")

	excused := map[string]string{} // gate id -> the workflow whose gate excuses it
	mg, _ := prScopeScenario(t, "merge_group", mgScope)
	if pass, log := runPRGateStep(t, prStep, mg); !pass {
		t.Fatalf("pr.yml's merge-group gate is red:\n%s", log)
	} else {
		for _, id := range ciGateAllowedSkips(log) {
			excused[id] = "pr.yml"
		}
	}
	mgRiskResults, mgRiskOutputs := riskScopeScenario("merge_group")
	if pass, log := runRiskGateStep(t, riskStep, "merge_group", mgRiskResults, mgScope, mgRiskOutputs); !pass {
		t.Fatalf("%s's merge-group gate is red:\n%s", prRiskWorkflowName, log)
	} else {
		for _, id := range ciGateAllowedSkips(log) {
			excused[id] = prRiskWorkflowName
		}
	}
	wantExcused := append([]string(nil), bazelDoltTierGateIDs...)
	for _, id := range ciScopeGateIDs["pr.yml"]["platforms"] {
		if contains(prEffectiveRequired(prStep, mg), id) {
			wantExcused = append(wantExcused, id)
		}
	}
	wantExcused = append(wantExcused, "TEST_NIX")
	for _, id := range riskEngineJobIDs {
		wantExcused = append(wantExcused, id)
	}
	slices.Sort(wantExcused)
	if got := slices.Sorted(maps.Keys(excused)); !slices.Equal(got, wantExcused) {
		t.Fatalf("merge-group skips%v, want exact%v", got, wantExcused)
	}

	bazel := readCIWorkflow(t, bazelWorkflowName)
	pushLanes := map[string]map[string]bool{} // bazel.yml's own push run: no inputs
	for name, job := range bazel.Jobs {
		if name != bazelRBEJobName && !isBazelRRCJob(name) && name != bazelDispatchCoverageJobName && name != "bazel-local-coverage" {
			pushLanes[name] = bazelLaneRunModes(t, name, job.If, map[string]string{})
		}
	}
	laneOf := map[string]string{} // gate id -> bazel.yml lane
	for lane, id := range bazelLaneGateIDs {
		laneOf[id] = lane
	}
	for lane, g := range bazelFlagGatedLanes {
		laneOf[g.id] = lane
	}
	needsResult := regexp.MustCompile(`^\$\{\{ needs\.([a-z0-9-]+)\.result \}\}$`)

	for id, wf := range excused {
		if replacements, ok := riskReplacementIDs[id]; ok && wf == prRiskWorkflowName {
			// Raw Risk jobs remain skipped on push too. The separate actual
			// helper negative tests require each replacement, including cmd
			// despite its historical advisory job status.
			for _, replacement := range replacements {
				if !slices.Contains(fullLocalRequiredResults, replacement) {
					t.Errorf("%s has unrequired replacement%s", id, replacement)
				}
			}
			continue
		}
		if lane, ok := laneOf[id]; ok && wf == "pr.yml" {
			reporter := bazelLaneReporter(pushLanes, lane, "remote")
			if reporter == "" {
				t.Errorf("%s: no job of bazel.yml's push run (mode remote) reports it", id)
				continue
			}
			job := bazel.job(t, reporter)
			if job.ContinueOnError {
				t.Errorf("%s: %s carries job-level continue-on-error; its failure would leave main's push run green", id, reporter)
			}
			key := "result"
			if bazelUnionOf(lane) == reporter && bazelUnionMemberOutputKey[lane] != "" {
				key = bazelUnionMemberOutputKey[lane]
			}
			if key == "result" {
				if job.Outputs["result"] != "${{ steps.result.outputs.result }}" {
					t.Errorf("%s: %s reports %v, not its job.status", id, reporter, job.Outputs)
				}
				continue // the job's own status: its failure is the run's
			}
			tier := strings.ReplaceAll(strings.TrimSuffix(key, "_result"), "_", "-")
			if strings.Contains(job.step(t, "Classify tier results").Run, "--gate-tier "+tier+" ") {
				continue // the classify step exits 1 on this tier's failure
			}
			failing := 0
			for _, step := range job.Steps {
				if !strings.Contains(step.If, "steps.classify.outputs."+key) {
					continue
				}
				// always() is true in a job that got this far.
				expr := strings.ReplaceAll(step.If, "always()", "true")
				runs := func(event, result string) bool {
					return evalRRCIf(t, expr, map[string]string{"github.event_name": event, "steps.classify.outputs." + key: result})
				}
				if !runs("push", "failure") {
					continue
				}
				failing++
				for _, event := range []string{"schedule", "workflow_dispatch", "pull_request_target"} {
					if !runs(event, "failure") {
						t.Errorf("%s: %s step %q does not fail a %s run on the %s tier's failure, where no gate reads %s", id, reporter, step.Name, event, tier, key)
					}
				}
				for _, event := range []string{"pull_request", "merge_group"} {
					if runs(event, "failure") {
						t.Errorf("%s: %s step %q fails pr.yml's %s call too, whose gate reads %s (a PR would show it twice)", id, reporter, step.Name, event, key)
					}
				}
				if runs("push", "success") || runs("push", "") {
					t.Errorf("%s: %s step %q runs without the %s tier's failure", id, reporter, step.Name, tier)
				}
				if step.ContinueOnError != nil && step.ContinueOnError != false {
					t.Errorf("%s: %s step %q carries continue-on-error", id, reporter, step.Name)
				}
				cmd := exec.Command("bash", "--noprofile", "--norc", "-eo", "pipefail", "-c", step.Run)
				cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "GITHUB_EVENT_NAME=push"}
				if out, err := cmd.CombinedOutput(); err == nil {
					t.Errorf("%s: %s step %q exits 0 on the %s tier's failure:\n%s", id, reporter, step.Name, tier, out)
				}
			}
			if failing == 0 {
				t.Errorf("%s: a %s tier failure leaves %s green in main's push run (no --gate-tier %s, no step failing on %s); main-ci-alert.yml would never see it", id, tier, reporter, tier, key)
			}
			continue
		}
		// One of the workflow's own jobs.
		w, step := pr, prStep
		if wf == prRiskWorkflowName {
			w, step = risk, riskStep
		}
		m := needsResult.FindStringSubmatch(step.Env[id])
		if m == nil {
			t.Errorf("%s gate env %s = %q: want a job's result", wf, id, step.Env[id])
			continue
		}
		job := w.job(t, m[1])
		if !evalRRCIf(t, job.If, ciScopeCtx(pushScope)) || job.ContinueOnError {
			t.Errorf("%s: %s job %s on push: runs %v, continue-on-error %v; want it to run and to fail the run when it fails", id, wf, m[1], evalRRCIf(t, job.If, ciScopeCtx(pushScope)), job.ContinueOnError)
		}
		var pass bool
		var log string
		if wf == "pr.yml" {
			sc, _ := prScopeScenario(t, "push", pushScope)
			sc.results = maps.Clone(sc.results)
			sc.results[m[1]] = "failure"
			pass, log = runPRGateStep(t, step, sc)
		} else {
			riskResults, riskOutputs := riskScopeScenario("push")
			riskResults[m[1]] = "failure"
			pass, log = runRiskGateStep(t, step, "push", riskResults, pushScope, riskOutputs)
		}
		if pass || !regexp.MustCompile(`::error::`+id+`\b`).MatchString(log) {
			t.Errorf("%s: %s failed on the push to main, %s's gate pass = %v; want red naming %s\n%s", id, m[1], wf, pass, id, log)
		}
	}
}

// TestCIScopeWorkflowsRunOnPushToMain: both workflows add exactly a push to
// main to their pull_request and merge_group triggers (the complement runs
// there), with no path filter.
func TestCIScopeWorkflowsRunOnPushToMain(t *testing.T) {
	for name := range ciScopeJobSwitch {
		var doc struct {
			On map[string]yaml.Node `yaml:"on"`
		}
		if err := yaml.Unmarshal([]byte(readPolicyFile(t, sourceRepoRoot(t), ".github/workflows/"+name)), &doc); err != nil {
			t.Fatal(err)
		}
		if got := slices.Sorted(maps.Keys(doc.On)); !slices.Equal(got, []string{"merge_group", "pull_request", "push"}) {
			t.Errorf("%s triggers = %v, want merge_group, pull_request and push", name, got)
		}
		var push map[string]any
		if node, ok := doc.On["push"]; !ok || node.Decode(&push) != nil || !reflect.DeepEqual(push, map[string]any{"branches": []any{"main"}}) {
			t.Errorf("%s push trigger = %v, want exactly branches: [main]", name, push)
		}
	}
}

// TestCIScopePushRunsHandleNullPullRequestFields: a push run has no
// pull_request and no merge_group payload. Every job that now also runs on a
// push to main reads neither, except through the Blacksmith venue ternary
// (which selects its pinned OS fallback there) and the concurrency group (one
// group per pushed commit, TestCIScopePushConcurrency).
func TestCIScopePushRunsHandleNullPullRequestFields(t *testing.T) {
	read := regexp.MustCompile(`github\.event\.(pull_request|merge_group)\.`)
	pushCtx := map[string]string{"github.event_name": "push", "github.repository": mergeQueueOwnRepo, "github.repository_owner": "gastownhall", "github.actor": "alice", "github.ref": "refs/heads/main", "github.sha": strings.Repeat("a", 40)}
	jobPath := regexp.MustCompile(`^\.jobs\.([^.]+)\.`)
	for name, switches := range ciScopeJobSwitch {
		onPush := map[string]bool{ciScopeJob: true, "ci-gate": true}
		for job, sw := range switches {
			onPush[job] = ciScopeRows["push"][sw] == "true"
		}
		seen := 0
		walkYAML(readYAMLNode(t, filepath.Join(".github", "workflows", name)), "", func(path string, key bool, value string) {
			if key || !read.MatchString(value) || path == ".concurrency.group" {
				return
			}
			m := jobPath.FindStringSubmatch(path)
			if m == nil {
				t.Errorf("%s %s reads %q outside a job", name, path, value)
				return
			}
			if !onPush[m[1]] {
				return
			}
			seen++
			if strings.HasSuffix(path, ".runs-on") {
				known := map[string]string{
					sameRepoBlacksmith2vcpu: "ubuntu-latest", sameRepoBlacksmith4vcpu: "ubuntu-latest", sameRepoBlacksmith8vcpu: "ubuntu-latest",
					sameRepoBlacksmithWindows2vcpu: "windows-latest", sameRepoBlacksmithWindows4vcpu: "windows-latest", sameRepoBlacksmithWindows8vcpu: "windows-latest", sameRepoBlacksmithWindows16vcpu: "windows-latest",
				}
				if value == sameRepoPlatformsMatrixMarkerRunsOn {
					for _, row := range []struct{ runner, os, want string }{
						{"same-repo-linux", "", "ubuntu-latest"}, {"same-repo-windows", "", "windows-latest"}, {"same-repo-macos", "", "macos-latest"},
						{"", "ubuntu-latest", "ubuntu-latest"}, {"", "windows-latest", "windows-latest"}, {"", "macos-latest", "macos-latest"},
					} {
						ctx := maps.Clone(pushCtx)
						ctx["matrix.runner"], ctx["matrix.os"] = row.runner, row.os
						if got := mustEvalGHRunsOn(t, value, ctx); got != row.want {
							t.Errorf("%s %s push matrix%+v =%q, want%q", name, path, row, got, row.want)
						}
					}
				} else if want, ok := known[value]; !ok {
					t.Errorf("%s %s reads PR fields through unreviewed runner expression%q", name, path, value)
				} else if got := mustEvalGHRunsOn(t, value, pushCtx); got != want {
					t.Errorf("%s %s on push = %q, want %q", name, path, got, want)
				}
				return
			}
			t.Errorf("%s %s reads %q, which is null on a push to main; this job runs there now", name, path, value)
		})
		if seen == 0 {
			t.Errorf("%s: found no push-run job reading a pull_request field; the walk is broken (the gates' runs-on reads one)", name)
		}
	}
}

// TestCIScopePushConcurrency: on a push to main both workflows (and
// bazel.yml, the Bazel half of the complement) key their concurrency group on
// the pushed commit, so no push run cancels another or waits behind it
// (GitHub cancels a group's pending run when a newer one queues): every
// push's complement runs to completion. Pull requests and merge groups
// keep their groups (TestMergeQueueConcurrencyNeverCancelsQueueRuns).
func TestCIScopePushConcurrency(t *testing.T) {
	for _, name := range []string{"pr.yml", prRiskWorkflowName, bazelWorkflowName} {
		var doc struct {
			Concurrency struct {
				Group            string `yaml:"group"`
				CancelInProgress any    `yaml:"cancel-in-progress"`
			} `yaml:"concurrency"`
		}
		if err := yaml.Unmarshal([]byte(readPolicyFile(t, sourceRepoRoot(t), ".github/workflows/"+name)), &doc); err != nil {
			t.Fatal(err)
		}
		push := func(sha string) map[string]string {
			return map[string]string{"github.workflow": strings.TrimSuffix(name, ".yml"), "github.event_name": "push", "github.ref": "refs/heads/main", "github.sha": sha}
		}
		a, b := strings.Repeat("a", 40), strings.Repeat("b", 40)
		ga, gb := interpolateGH(t, doc.Concurrency.Group, push(a)), interpolateGH(t, doc.Concurrency.Group, push(b))
		if ga == gb || !strings.Contains(ga, a) {
			t.Errorf("%s concurrency group on push: %q and %q; want one group per pushed commit", name, ga, gb)
		}
		for _, other := range []map[string]string{mergeGroupCtx(1, map[string]string{"github.sha": a}), pullRequestCtx(1, map[string]string{"github.sha": a})} {
			ctx := maps.Clone(other)
			ctx["github.workflow"] = strings.TrimSuffix(name, ".yml")
			if g := interpolateGH(t, doc.Concurrency.Group, ctx); g == ga || strings.Contains(g, a) {
				t.Errorf("%s concurrency group for %s = %q: keyed on the commit (or equal to the push group %q); want the PR number or the queue ref", name, other["github.event_name"], g, ga)
			}
		}
	}
}
