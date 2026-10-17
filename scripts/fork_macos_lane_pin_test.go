package scripts_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// This file is the fork's own (upstream has no copy, so syncs never conflict
// on it). It pins .github/workflows/fork-macos.yml, the one macOS lane in this
// repository (rule M1, be-6i64d8).
//
// Upstream's F4 deleted pr.yml's PR macOS job and left main.yml's push-to-main
// leg as its only macOS lane. This repository's main is a byte-identical
// upstream mirror, so that leg never sees a fork change; the job upstream
// deleted lives on here, in a workflow of its own, so pr.yml stays upstream's
// text. Its assertions were pr.yml's TestMacOSTestJobsReuseWorkspaceBDBinary
// hunks before the move.

const forkMacOSWorkflow = "fork-macos.yml"

func TestForkMacOSLane(t *testing.T) {
	const (
		workspaceBDBinary = "${{ github.workspace }}/bd"
		buildCommand      = "go build -v -tags gms_pure_go ./cmd/bd"
		// Graphstore is exhaustively grouped, preserving each process's 30m
		// alarm.
		testCommand = "python3 scripts/ci/macos-go-test.py"
	)
	job := readCIWorkflow(t, forkMacOSWorkflow).job(t, "test-macos")
	if job.RunsOn != macOSRunner {
		t.Errorf("macOS test runner = %q, want %q", job.RunsOn, macOSRunner)
	}
	assertStepRunsExactly(t, job, "Build", buildCommand)
	assertStepRunsExactly(t, job, "Test", testCommand)
	assertStepsBefore(t, job, []string{"Build"}, []string{"Test"})
	assertStepEnvValue(t, job, "Test", "BEADS_TEST_BD_BINARY", workspaceBDBinary)
	if step := job.step(t, "Test"); step.TimeoutMinutes != 120 || step.ContinueOnError != nil || step.If != "" {
		t.Error("macOS dispatch must remain bounded, unconditional and failure-propagating")
	}
	upload := job.step(t, "Upload macOS test evidence")
	if upload.If != "always()" || upload.TimeoutMinutes != 5 || upload.With["name"] != "macos-go-test-${{ github.sha }}" || upload.With["path"] != "artifacts/macos-go-test" || upload.With["retention-days"] != "7" || upload.With["if-no-files-found"] != "warn" {
		t.Error("macOS evidence must survive failure with a bounded commit-specific upload")
	}
}

// The lane runs pull-request code, so it may only run unprivileged: triggers
// that carry no secrets, a read-only token, no secret reference anywhere, and
// no job that widens its own permissions. It is also not part of pr.yml's
// CI Gate: a macOS run takes about 90 minutes.
func TestForkMacOSLaneIsUnprivilegedAndOutsideTheGate(t *testing.T) {
	rel := filepath.Join(".github", "workflows", forkMacOSWorkflow)
	var wf struct {
		On          map[string]any `yaml:"on"`
		Permissions map[string]any `yaml:"permissions"`
	}
	data, err := os.ReadFile(filepath.Join(sourceRepoRoot(t), rel))
	if err != nil {
		t.Fatal(err)
	}
	if err := yaml.Unmarshal(data, &wf); err != nil {
		t.Fatalf("parse %s: %v", rel, err)
	}
	wantOn := map[string]any{"pull_request": map[string]any{"branches": []any{"integration"}}, "workflow_dispatch": nil}
	if !reflect.DeepEqual(wf.On, wantOn) {
		t.Errorf("%s triggers = %v, want %v (never pull_request_target or workflow_run: it runs pull-request code)", rel, wf.On, wantOn)
	}
	if !reflect.DeepEqual(wf.Permissions, map[string]any{"contents": "read"}) {
		t.Errorf("%s permissions = %v, want exactly contents: read", rel, wf.Permissions)
	}
	if strings.Contains(string(data), "secrets.") {
		t.Errorf("%s references a secret; the lane runs pull-request code", rel)
	}
	for name, job := range readCIWorkflow(t, forkMacOSWorkflow).Jobs {
		if job.Permissions != nil {
			t.Errorf("%s job %s sets its own permissions %v", rel, name, job.Permissions)
		}
	}
	if contains(readCIWorkflow(t, "pr.yml").job(t, "ci-gate").Needs, "test-macos") {
		t.Error("pr.yml's ci-gate needs test-macos; the fork's macOS lane is a separate check")
	}
}
