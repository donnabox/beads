package scripts_test

import (
	"sort"
	"strings"
	"testing"
)

// This file is the fork's own (upstream has no copy, so syncs never conflict
// on it). It pins the one upstream decision the fork relies on and upstream
// does not test for it.
//
// Upstream retires its legacy PR jobs (PR Core, build-artifacts, the pure-Go
// check, domain+uow, contract corpus, and PR Risk's embedded and server-Dolt
// tiers) in favour of remote Bazel lanes, but only on same-repository,
// non-Dependabot pull requests whose head repository is not a GitHub fork
// (bazel-coverage's decision step). The fork has no Bazel farm, so it keeps
// those jobs by taking upstream's workflows unchanged: its repository is
// itself a GitHub fork, so every pull request inside it, same-repository
// branches included, has head.repo.fork true and the decision retires
// nothing. Graph core (graph-c0) relies on the same decision: it installs bd
// from build-artifacts' ci-build-artifacts, and upstream stands build-artifacts
// down wherever pr_lanes is true. If upstream changes the decision, these
// tests fail in the sync pull request instead of the fork losing its lanes.

// A pull request inside the fork with every retirement flag on: the decision
// step, run as GitHub runs it, retires no tier in either workflow.
func TestForkPullRequestsKeepLegacyLanes(t *testing.T) {
	requireHostTool(t, "bash")
	facts := rbeFacts{event: "pull_request", fork: true, dependabot: false}
	for i := range facts.retired {
		facts.retired[i] = "true"
	}
	wantEnv := map[string]string{"PULL_REQUEST": "true", "FORK": "true", "DEPENDABOT": "false"}
	for _, r := range retiredTiers {
		wantEnv[r.envKey] = "true"
	}
	for _, workflow := range []string{"pr.yml", prRiskWorkflowName} {
		step := coverageStep(t, workflow)
		env := map[string]string{}
		for k, v := range step.Env {
			env[k] = evalRBEExpr(t, v, facts, nil)
		}
		for _, k := range sortedKeys(wantEnv) {
			if env[k] != wantEnv[k] {
				t.Errorf("%s %s: %s = %q for a pull request inside a GitHub fork, want %q", workflow, prRiskCoverageJobName, k, env[k], wantEnv[k])
			}
		}
		out, err := runBazelRBEDecision(t, step.Run, env)
		if err != nil {
			t.Fatalf("%s %s decision step: %v", workflow, prRiskCoverageJobName, err)
		}
		for _, output := range []string{"embedded", "dolt_server", "pr_lanes"} {
			if _, ok := out[output]; !ok {
				t.Errorf("%s %s reported no %s output: %v", workflow, prRiskCoverageJobName, output, out)
			}
		}
		var retired []string
		for output, v := range out {
			if v != "false" {
				retired = append(retired, output+"="+v)
			}
		}
		sort.Strings(retired)
		if len(retired) > 0 {
			t.Errorf("%s %s retires legacy tiers on a pull request inside a GitHub fork (%s); the fork has no Bazel farm to replace them", workflow, prRiskCoverageJobName, strings.Join(retired, ", "))
		}
	}
}

// The jobs that rely on that decision are still defined, Graph core still
// takes its bd from build-artifacts, and ci-gate still requires Graph core.
func TestForkKeepsLegacyJobsAndGraphCore(t *testing.T) {
	pr := readCIWorkflow(t, "pr.yml")
	for _, name := range []string{"pr-core-wrapper", "build-artifacts", "graph-c0"} {
		if _, ok := pr.Jobs[name]; !ok {
			t.Errorf("pr.yml has no %s job", name)
		}
	}
	if graph, ok := pr.Jobs["graph-c0"]; ok && !contains(graph.Needs, "build-artifacts") {
		t.Errorf("graph-c0 needs %v, want build-artifacts (it installs bd from ci-build-artifacts)", graph.Needs)
	}
	gate := pr.job(t, "ci-gate")
	step := gate.step(t, "Evaluate CI gate")
	if !contains(gate.Needs, "graph-c0") || !contains(strings.Fields(step.Env["CI_GATE_REQUIRED"]), "GRAPH_C0") ||
		step.Env["GRAPH_C0"] != "${{ needs.graph-c0.result }}" {
		t.Errorf("ci-gate does not require graph-c0's result as GRAPH_C0 (needs %v, GRAPH_C0 = %q)", gate.Needs, step.Env["GRAPH_C0"])
	}
}
