package scripts_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func fullLocalGateFacts() map[string]string {
	facts := map[string]string{
		"GITHUB_REPOSITORY": "donnabox/beads", "BAZEL_RBE_MODE": "local",
		"BAZEL_RBE_ENABLED": "false", "BAZEL_LOCAL_FULL_INTENT": "true",
		"BAZEL_DOLT_TIERS": "on", "BAZEL_LOCAL_FULL": "success",
	}
	for _, id := range fullLocalRequiredResults {
		facts[id] = "success"
	}
	return facts
}

var fullLocalRequiredResults = []string{
	"BAZEL_TEST", "BAZEL_PURE", "BAZEL_RELEASE_CROSS", "BAZEL_EMBEDDED",
	"BAZEL_INTEGRATION", "BAZEL_DOLTSERVER", "BAZEL_PROXIED",
	"BAZEL_SERVER_STORAGE", "BAZEL_CMD_DOLT_DIRECT",
}

func runFullLocalGate(t *testing.T, command string, facts map[string]string) (string, error) {
	t.Helper()
	cmd := exec.Command("bash", filepath.Join(bazelPolicyRoot(t), ".github/scripts/bazel-gate.sh"), command)
	// No ambient result, repository, scope, credential or routing fact may
	// make a deliberately absent fixture fact succeed.
	cmd.Env = []string{"PATH=" + os.Getenv("PATH")}
	for key, value := range facts {
		cmd.Env = append(cmd.Env, key+"="+value)
	}
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func TestForkFullLocalCoverageRequiresEveryActualTier(t *testing.T) {
	if out, err := runFullLocalGate(t, "local-full", fullLocalGateFacts()); err != nil {
		t.Fatalf("complete dedicated local proof rejected: %v: %s", err, out)
	}
	for _, id := range fullLocalRequiredResults {
		for _, result := range []string{"", "skipped", "cancelled", "failure", "unknown"} {
			t.Run(id+"/"+result, func(t *testing.T) {
				facts := fullLocalGateFacts()
				facts[id] = result
				if out, err := runFullLocalGate(t, "local-full", facts); err == nil || !strings.Contains(out, id+"=") {
					t.Fatalf("%s=%q did not fail on its own required result: %v: %s", id, result, err, out)
				}
			})
		}
	}
	for _, id := range []string{"BAZEL_DOLT_RACE", "BAZEL_DOLT_INT"} {
		facts := fullLocalGateFacts()
		facts[id] = "success"
		if out, err := runFullLocalGate(t, "local-full", facts); err == nil {
			t.Fatalf("active remote union %s accepted in local execution: %s", id, out)
		}
	}
}

func TestForkFullLocalCoverageCannotClaimRemoteOrExcuseSkips(t *testing.T) {
	facts := fullLocalGateFacts()
	out, err := runFullLocalGate(t, "remote", facts)
	if err != nil || strings.TrimSpace(out) != "not-run-remotely" {
		t.Fatalf("local execution relabeled remote: %v: %s", err, out)
	}
	facts["SCOPE_DOLT_TIERS"] = "off"
	facts["BAZEL_DOLT_TIERS"] = "off"
	if out, err := runFullLocalGate(t, "skips", facts); err != nil || strings.TrimSpace(out) != "" {
		t.Fatalf("scope excused required full-local tiers: %v: %s", err, out)
	}
	if out, err := runFullLocalGate(t, "local-full", facts); err == nil {
		t.Fatalf("effective Dolt tiers off accepted: %s", out)
	}
	for key, values := range map[string][]string{
		"GITHUB_REPOSITORY":       {"gastownhall/beads", "other/beads", ""},
		"BAZEL_RBE_MODE":          {"remote", "cache", "skip", ""},
		"BAZEL_RBE_ENABLED":       {"true", ""},
		"BAZEL_LOCAL_FULL_INTENT": {"false", "", "unknown"},
	} {
		for _, value := range values {
			candidate := fullLocalGateFacts()
			candidate[key] = value
			if out, err := runFullLocalGate(t, "local-full", candidate); err == nil {
				t.Fatalf("invalid %s=%q accepted: %s", key, value, out)
			}
		}
	}
}

func TestForkLocalDispatchRequiresActualCompleteGateAndCommand(t *testing.T) {
	if out, err := runFullLocalGate(t, "dispatch-cmd", fullLocalGateFacts()); err != nil {
		t.Fatalf("complete local dispatch rejected: %v: %s", err, out)
	}
	for _, key := range []string{"BAZEL_LOCAL_FULL", "BAZEL_INTEGRATION", "BAZEL_CMD_DOLT_DIRECT"} {
		for _, result := range []string{"", "skipped", "cancelled", "failure", "unknown"} {
			facts := fullLocalGateFacts()
			facts[key] = result
			facts["BAZEL_CMD_DOLT_COMBINED"] = "success"
			if out, err := runFullLocalGate(t, "dispatch-cmd", facts); err == nil {
				t.Fatalf("local dispatch %s=%q masked by unrelated success: %s", key, result, out)
			}
		}
	}
}

func TestForkFullLocalWorkflowUsesDedicatedRecordedResults(t *testing.T) {
	var workflow struct {
		Jobs map[string]struct {
			If              string               `yaml:"if"`
			ContinueOnError bool                 `yaml:"continue-on-error"`
			RunsOn          string               `yaml:"runs-on"`
			TimeoutMinutes  int                  `yaml:"timeout-minutes"`
			Needs           ciWorkflowStringList `yaml:"needs"`
			Outputs         map[string]string    `yaml:"outputs"`
			Steps           []struct {
				ID              string            `yaml:"id"`
				If              string            `yaml:"if"`
				ContinueOnError any               `yaml:"continue-on-error"`
				Run             string            `yaml:"run"`
				Env             map[string]string `yaml:"env"`
			} `yaml:"steps"`
		} `yaml:"jobs"`
	}
	data := readPolicyFile(t, bazelPolicyRoot(t), ".github/workflows/bazel.yml")
	if err := decodeCIWorkflowYAMLForRBE("bazel.yml", []byte(data), false, &workflow); err != nil {
		t.Fatal(err)
	}
	gate, ok := workflow.Jobs["bazel-local-coverage"]
	if !ok {
		t.Fatal("missing independent full-local coverage gate")
	}
	if gate.If != "${{ always() && needs.rbe.outputs.local_full == 'true' }}" || gate.ContinueOnError || gate.RunsOn != "ubuntu-latest" || gate.TimeoutMinutes != 5 {
		t.Fatalf("full-local gate must always run fail-closed under actual Q: %+v", gate)
	}
	wantNeeds := []string{"rbe", "bazel-test", "bazel-pure", "bazel-release-cross", "bazel-embedded", "bazel-integration", "bazel-doltserver", "bazel-proxied", "bazel-server-storage", "bazel-cmd-dolt", "bazel-dolt-race", "bazel-dolt-int"}
	if !slices.Equal([]string(gate.Needs), wantNeeds) {
		t.Errorf("full-local gate needs%v, want exact%v", gate.Needs, wantNeeds)
	}
	dedicated := map[string]string{
		"BAZEL_TEST": "bazel-test", "BAZEL_PURE": "bazel-pure", "BAZEL_RELEASE_CROSS": "bazel-release-cross",
		"BAZEL_EMBEDDED": "bazel-embedded", "BAZEL_INTEGRATION": "bazel-integration", "BAZEL_DOLTSERVER": "bazel-doltserver",
		"BAZEL_PROXIED": "bazel-proxied", "BAZEL_SERVER_STORAGE": "bazel-server-storage", "BAZEL_CMD_DOLT_DIRECT": "bazel-cmd-dolt",
	}
	var actual map[string]string
	recorded := false
	for _, step := range gate.Steps {
		if strings.TrimSpace(step.Run) == ".github/scripts/bazel-gate.sh local-full" {
			if step.If != "" || (step.ContinueOnError != nil && step.ContinueOnError != false) {
				t.Fatal("local coverage evaluation may not skip or mask failures")
			}
			actual = step.Env
		}
		if step.ID == "result" && step.If == "always()" && step.Run == `echo "result=${{ job.status }}" >> "$GITHUB_OUTPUT"` && (step.ContinueOnError == nil || step.ContinueOnError == false) {
			recorded = true
		}
	}
	if actual == nil || !recorded || gate.Outputs["result"] != "${{ steps.result.outputs.result }}" {
		t.Fatal("gate must expose its actual always-recorded job.status")
	}
	for id, job := range dedicated {
		found := false
		for _, need := range gate.Needs {
			if need == job {
				found = true
			}
		}
		if !found || actual[id] != "${{ needs."+job+".outputs.result }}" {
			t.Errorf("required dedicated result %s is missing or mapped from another/masked arm", id)
		}
	}
	for _, job := range []string{"bazel-embedded", "bazel-integration", "bazel-doltserver", "bazel-proxied", "bazel-server-storage", "bazel-cmd-dolt"} {
		if !strings.Contains(workflow.Jobs[job].If, "needs.rbe.outputs.local_full == 'true'") {
			t.Errorf("%s excludes full-local selection", job)
		}
	}
	for _, job := range []string{"bazel-dolt-race", "bazel-dolt-int", "rbe-prewarm"} {
		if strings.Contains(workflow.Jobs[job].If, "local_full") {
			t.Errorf("full-local selection enabled remote-only %s", job)
		}
	}
}

// Exercise the receiving workflow as well as its result-only helper. A valid
// full-local result does not mask an individual tier failure, and a fabricated
// selection marker is never sufficient to turn a non-Q execution green.
func TestForkFullLocalReceivingGate(t *testing.T) {
	step := readCIWorkflow(t, "pr.yml").job(t, "ci-gate").step(t, "Evaluate CI gate")
	workflow := readCIWorkflow(t, bazelWorkflowName)
	for _, event := range []string{"pull_request", "merge_group"} {
		t.Run(event, func(t *testing.T) {
			with := prCallWithFor(t, event)
			outputs := map[string]string{}
			for name, job := range workflow.Jobs {
				if name == bazelRBEJobName || name == bazelDispatchCoverageJobName || name == "bazel-local-coverage" || isBazelRRCJob(name) {
					continue
				}
				if bazelLaneRunModesWithLocalProof(t, name, job.If, with, true)["local"] {
					outputs[name] = "success"
				}
			}
			covered := coveredAll("false")
			if event == "merge_group" {
				covered["embedded"] = "true"
			}
			base := bazelGateScenario{name: event + " exact fork full-local", event: event, repository: "donnabox/beads", mode: "local", enabled: "false", call: "success", outputs: outputs, covered: covered, callWith: with, localIntent: "true", localResult: "success"}
			base.flags = map[string]string{}
			for _, g := range bazelFlagGatedLanes {
				base.flags[g.flag] = readCIWorkflow(t, "pr.yml").Env[g.flag]
			}
			if pass, log := runPRGateStep(t, step, base); !pass {
				t.Fatalf("complete local receiving proof rejected: %s", log)
			}
			for _, result := range []string{"", "skipped", "cancelled", "failure", "unknown"} {
				sc := base
				sc.localResult = result
				if pass, log := runPRGateStep(t, step, sc); pass {
					t.Errorf("full-local result %q accepted: %s", result, log)
				}
			}
			for _, lane := range []string{bazelJobName, bazelPureJobName, bazelReleaseCrossJobName, bazelEmbedJobName, bazelIntegJobName, bazelDoltJobName, bazelProxiedJobName, bazelServerJobName, bazelCmdDoltJobName} {
				for _, result := range []string{"", "skipped", "cancelled", "failure"} {
					sc := base
					sc.outputs = copyMap(base.outputs)
					sc.outputs[lane] = result
					if pass, log := runPRGateStep(t, step, sc); pass {
						t.Errorf("%s=%q masked by complete aggregate: %s", lane, result, log)
					}
				}
			}
			for _, repo := range []string{"", "gastownhall/beads", "other/beads"} {
				sc := base
				sc.repository = repo
				if pass, log := runPRGateStep(t, step, sc); pass {
					t.Errorf("foreign repository %q accepted full-local marker: %s", repo, log)
				}
			}
			for _, intent := range []string{"", "false", "unknown"} {
				sc := base
				sc.localIntent = intent
				if pass, log := runPRGateStep(t, step, sc); pass {
					t.Errorf("local intent %q accepted: %s", intent, log)
				}
			}
		})
	}
}

func TestForkFullLocalSelectionRequiresExactRepository(t *testing.T) {
	step := readCIWorkflow(t, bazelWorkflowName).job(t, bazelRBEJobName).Steps[0]
	for _, repo := range []string{"donnabox/beads", "someone/beads", "donnabox/other", ""} {
		t.Run(repo, func(t *testing.T) {
			env := map[string]string{
				"GITHUB_REPOSITORY": repo,
				"RBE_VAR_ON":        "true", "RBE_INPUT_OFF": "false", "RBE_INPUT_CACHE": "false",
				"PULL_REQUEST": "true", "FORK": "true", "FORK_REPO": "true",
				"DEPENDABOT": "false", "FORK_FARM": "false", "HAS_EXECUTOR": "false",
				"PR_NUMBER": "7123", bazelTestMintEnv: "rw",
			}
			out, log, _, err := runBazelRBEDecisionSummary(t, step.Run, env)
			if err != nil {
				t.Fatalf("decision failed: %v", err)
			}
			want := "false"
			if repo == "donnabox/beads" {
				want = "true"
			}
			if out["mode"] != "local" || out["enabled"] != "false" || out["local_full"] != want || log != "" {
				t.Errorf("repository %q: outputs %v, mint log %q; want local/false, intent %s and no mint access", repo, out, log, want)
			}
		})
	}
}
