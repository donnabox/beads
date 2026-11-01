//go:build !windows

package scripts_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// This runs the real workflow shell bodies and its actual scheduling clauses.
// Commands are temporary stand-ins: the test proves ordering/status behavior,
// not Bazel internals, XML correctness, GitHub cancellation or cache health.
func TestBazelIntegrationResultFinalization(t *testing.T) {
	job := readCIWorkflow(t, bazelWorkflowName).job(t, bazelIntegJobName)
	t.Run("delayed-prior-writer", func(t *testing.T) {
		fx := newIntegrationFinalizationFixture(t, job, "delayed", "success", true)
		run := fx.start(t)
		run.await(t, "nogo-waiting")
		fx.assertAbsent(t, "checker", "analytics")
		fx.completePrior(t)
		run.finish(t, false)
		fx.assertOrder(t, "nogo-waiting", "prior-complete", "nogo-complete", "checker", "analytics", "result:failure")
	})
	t.Run("no-settlement", func(t *testing.T) {
		for _, outcome := range []string{"failure", "timeout"} {
			t.Run(outcome, func(t *testing.T) {
				// A timeout stand-in returns124 without a finalization marker;
				// this does not pretend to simulate the Actions90-minute clock.
				fx := newIntegrationFinalizationFixture(t, job, "delayed", outcome, true)
				run := fx.start(t)
				run.finish(t, false)
				fx.assertAbsent(t, "checker", "analytics", "prior-complete", "nogo-complete")
				fx.assertOrder(t, "nogo-enter", "unavailable", "result:failure")
			})
		}
	})
	t.Run("successful-test-failed-nogo", func(t *testing.T) {
		fx := newIntegrationFinalizationFixture(t, job, "success", "failure", false)
		run := fx.start(t)
		run.finish(t, false)
		fx.assertOrder(t, "test-complete", "nogo-enter", "checker", "analytics", "result:failure")
		fx.assertAbsent(t, "unavailable")
	})
	t.Run("natural-test-failure-successful-nogo", func(t *testing.T) {
		fx := newIntegrationFinalizationFixture(t, job, "failure", "success", false)
		run := fx.start(t)
		run.finish(t, false)
		fx.assertOrder(t, "test-complete", "nogo-complete", "checker", "analytics", "result:failure")
		fx.assertAbsent(t, "unavailable")
	})
	t.Run("cancel-during-serialization", func(t *testing.T) {
		fx := newIntegrationFinalizationFixture(t, job, "delayed", "success", true)
		run := fx.start(t)
		run.await(t, "nogo-waiting")
		run.cancel(t)
		run.finishCancelled(t)
		fx.assertAbsent(t, "nogo-complete", "checker", "analytics", "result:success", "result:failure")
		// No unlock/retry/successor is launched after cancellation. No inner
		// exit is invented when cancellation prevents its completion marker.
	})
}

type integrationFinalizationFixture struct {
	root, state, bash, script string
	env                       []string
}

func integrationFinalizationQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func integrationFinalizationCondition(t *testing.T, expression string) string {
	t.Helper()
	expression = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(expression, "${{"), "}}"))
	// Deliberately only the actual reviewed workflow clauses. A new clause
	// requires receiving, not an accidental expansion into an Actions parser.
	allowed := map[string]bool{
		"": true, "always()": true,
		"always() && steps.test.outcome != 'skipped'": true,
		"always() && steps.test.outcome != 'skipped' && steps.test.outcome != 'success' && steps.nogo-integration.outcome != 'success'":       true,
		"always() && steps.test.outcome != 'skipped' && (steps.test.outcome == 'success' || steps.nogo-integration.outcome == 'success')":     true,
		"!cancelled() && steps.test.outcome != 'skipped' && (steps.test.outcome == 'success' || steps.nogo-integration.outcome == 'success')": true,
	}
	if !allowed[expression] {
		t.Fatalf("unreceived integration clause: %q", expression)
	}
	if expression == "" {
		return "true"
	}
	expression = strings.NewReplacer(
		"!cancelled()", `"$cancelled" != '1'`,
		"always()", "'1' == '1'",
		"steps.test.outcome", `"$test_outcome"`,
		"steps.nogo-integration.outcome", `"$nogo_outcome"`,
	).Replace(expression)
	return "[[ " + expression + " ]]"
}

func newIntegrationFinalizationFixture(t *testing.T, job ciWorkflowJob, testMode, nogoMode string, locked bool) *integrationFinalizationFixture {
	t.Helper()
	fx := &integrationFinalizationFixture{root: t.TempDir(), bash: requireHostTool(t, "bash")}
	fx.state = filepath.Join(fx.root, "state")
	bin, tmp := filepath.Join(fx.root, "bin"), filepath.Join(fx.root, "tmp")
	for _, dir := range []string{fx.state, bin, filepath.Join(tmp, "ci-analytics")} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	if locked {
		if err := os.Mkdir(filepath.Join(fx.state, "lock"), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	// Actual shell utilities stay explicit host dependencies, matching the
	// existing scripts tests. No real bazel/python/zstd can escape this PATH.
	for _, name := range []string{"date", "tee", "grep", "tail", "sleep"} {
		if err := os.Symlink(requireHostTool(t, name), filepath.Join(bin, name)); err != nil {
			t.Fatal(err)
		}
	}
	write := func(path, body string) {
		t.Helper()
		if err := os.WriteFile(path, []byte("#!"+fx.bash+"\nset -eu\n"+body), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join(bin, "bazel"), `
event() { printf '%s\n' "$1" >> "$BD_STANDIN_STATE/events"; }
case "$1" in
  test)
    printf '{}\n' > "$RUNNER_TEMP/bazel-bep.json"
    event test-enter
    if [[ "$BD_STANDIN_TEST" == delayed ]]; then exit 124; fi
    printf 'complete\n' > "$BD_STANDIN_STATE/xml-ready"
    event test-complete
    if [[ "$BD_STANDIN_TEST" == failure ]]; then exit 3; fi
    exit 0 ;;
  build)
    event nogo-enter
    if [[ "$BD_STANDIN_NOGO" == failure ]]; then exit 9; fi
    if [[ "$BD_STANDIN_NOGO" == timeout ]]; then exit 124; fi
    if [[ -d "$BD_STANDIN_STATE/lock" ]]; then
      event nogo-waiting
      while [[ -d "$BD_STANDIN_STATE/lock" ]]; do sleep 0.01; done
    fi
    [[ -f "$BD_STANDIN_STATE/xml-ready" ]] || exit 72
    event nogo-complete
    exit 0 ;;
  *) exit 73 ;;
esac
`)
	write(filepath.Join(bin, "python3"), `
case "$1" in
  tools/bazel/check_testcases.py)
    [[ -f "$BD_STANDIN_STATE/xml-ready" ]] || exit 74
    printf 'checker\n' >> "$BD_STANDIN_STATE/events" ;;
  tools/bazel/ci_analytics_extract.py)
    [[ -f "$BD_STANDIN_STATE/xml-ready" ]] || exit 75
    printf 'analytics\n' >> "$BD_STANDIN_STATE/events"
    printf '{}\n' > "$RUNNER_TEMP/ci-analytics/ci-analytics.json" ;;
  *) exit 76 ;;
esac
`)
	write(filepath.Join(bin, "zstd"), "exit 0\n")
	var driver strings.Builder
	driver.WriteString("set -eu\njob_failed=0\ncancelled=0\ntest_outcome=skipped\nnogo_outcome=skipped\n")
	selected := map[string]string{
		"bazel test //... --config=integration": "test",
		"Nogo: integration library files":       "nogo",
		"Final test results unavailable":        "unavailable",
		"Every target and shard ran tests":      "checker",
		"CI analytics summary":                  "analytics",
		"Record job result":                     "result",
	}
	seen := map[string]bool{}
	for i, step := range job.Steps {
		role, wanted := selected[step.Name]
		if !wanted {
			continue
		}
		if seen[role] {
			t.Fatalf("duplicate selected workflow step: %s", role)
		}
		seen[role] = true
		path := filepath.Join(fx.root, fmt.Sprintf("step-%d.sh", i))
		if err := os.WriteFile(path, []byte(step.Run), 0o600); err != nil {
			t.Fatal(err)
		}
		// Execute the literal workflow Run body, in literal workflow order.
		fmt.Fprintf(&driver, "if %s; then\n", integrationFinalizationCondition(t, step.If))
		if role == "result" {
			driver.WriteString("if [[ $job_failed == 1 ]]; then export JOB_STATUS=failure; else export JOB_STATUS=success; fi\n")
		}
		if role == "unavailable" {
			driver.WriteString("printf 'unavailable\\n' >> \"$BD_STANDIN_STATE/events\"\n")
		}
		fmt.Fprintf(&driver, "if \"$BD_STANDIN_BASH\" -e -- %s; then step_outcome=success; else step_outcome=failure; job_failed=1; fi\n", integrationFinalizationQuote(path))
		if role == "test" {
			driver.WriteString("test_outcome=$step_outcome\n")
		}
		if role == "nogo" {
			driver.WriteString("nogo_outcome=$step_outcome\n")
		}
		if role == "result" {
			driver.WriteString("printf 'result:%s\\n' \"$JOB_STATUS\" >> \"$BD_STANDIN_STATE/events\"\n")
		}
		driver.WriteString("fi\n")
	}
	for _, role := range selected {
		if !seen[role] {
			t.Fatalf("missing selected workflow step: %s", role)
		}
	}
	driver.WriteString("if [[ $job_failed == 1 ]]; then exit 1; fi\n")
	fx.script = filepath.Join(fx.root, "drive.sh")
	if err := os.WriteFile(fx.script, []byte(driver.String()), 0o600); err != nil {
		t.Fatal(err)
	}
	fx.env = []string{"PATH=" + bin, "HOME=" + fx.root, "RUNNER_TEMP=" + tmp,
		"GITHUB_STEP_SUMMARY=" + filepath.Join(fx.root, "summary"), "GITHUB_OUTPUT=" + filepath.Join(fx.root, "output"),
		"BD_STANDIN_STATE=" + fx.state, "BD_STANDIN_TEST=" + testMode, "BD_STANDIN_NOGO=" + nogoMode,
		"BD_STANDIN_BASH=" + fx.bash, "LANE=integration", "MODE=cache", "CHECK_RUN_ID=1", "PR_HINT=350",
		"BAZEL_FRESH=", "EXCLUDE_TARGETS=", "BASH_ENV=", "ENV="}
	return fx
}

type integrationFinalizationRun struct {
	fx                                            *integrationFinalizationFixture
	cmd                                           *exec.Cmd
	stdout, stderr                                *os.File
	done                                          chan error
	ctx                                           context.Context
	cancelled, signalsClosed, waitStarted, reaped bool
	waitErr, stopErr                              error
}

func (fx *integrationFinalizationFixture) start(t *testing.T) *integrationFinalizationRun {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	r := &integrationFinalizationRun{fx: fx, ctx: ctx, done: make(chan error, 1)}
	var err error
	r.stdout, err = os.OpenFile(filepath.Join(fx.root, "supervisor.stdout"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	r.stderr, err = os.OpenFile(filepath.Join(fx.root, "supervisor.stderr"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		_ = r.stdout.Close()
		cancel()
		t.Fatal(err)
	}
	// The outer leader waits for the literal inner workflow, records its REAL
	// exit code, then holds. Even an exited leader is not reaped by Go until
	// all group signaling is permanently closed. Private regular output files
	// avoid pipe-copy waiters. This is a fixture, not a production supervisor.
	supervisor := `set +e
"$BD_STANDIN_BASH" --noprofile --norc -e -- "$1"
workflow_status=$?
printf '%d\n' "$workflow_status" > "$BD_STANDIN_STATE/workflow-exit" || exit 79
while :; do sleep 1; done
`
	r.cmd = exec.Command(fx.bash, "--noprofile", "--norc", "-c", supervisor, "standin-supervisor", fx.script)
	r.cmd.Env, r.cmd.Dir = fx.env, fx.root
	r.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	r.cmd.Stdout, r.cmd.Stderr = r.stdout, r.stderr
	if err := r.cmd.Start(); err != nil {
		_ = r.stdout.Close()
		_ = r.stderr.Close()
		cancel()
		t.Fatal(err)
	}
	// No Wait goroutine exists during any allowed negative-PGID signal.
	// This cleanup is registered after TempDir: it runs before file removal.
	t.Cleanup(func() {
		if err := r.stop(); err != nil {
			t.Errorf("stand-in cleanup unsettled/failed: %v", err)
		}
		if err := errors.Join(r.stdout.Close(), r.stderr.Close()); err != nil {
			t.Errorf("close stand-in output: %v", err)
		}
		cancel()
	})
	return r
}

// stop sends the FINAL group signal while the leader is OS-unreaped, closes
// signaling permanently, then starts the sole Wait. A timed-out Wait remains
// unsettled: cleanup cannot signal, probe or start a second Wait afterward.
func (r *integrationFinalizationRun) stop() error {
	if r.waitStarted {
		return r.stopErr
	}
	killErr := syscall.Kill(-r.cmd.Process.Pid, syscall.SIGKILL)
	if errors.Is(killErr, syscall.ESRCH) {
		killErr = nil
	}
	r.signalsClosed = true
	r.waitStarted = true
	go func() { r.done <- r.cmd.Wait() }()
	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()
	select {
	case r.waitErr = <-r.done:
		r.reaped = true
		var exited *exec.ExitError
		if !errors.As(r.waitErr, &exited) {
			r.stopErr = errors.Join(killErr, fmt.Errorf("holding supervisor ended unexpectedly: %v", r.waitErr))
			return r.stopErr
		}
		status, ok := exited.Sys().(syscall.WaitStatus)
		expected := ok && status.Signaled() && (status.Signal() == syscall.SIGKILL || (r.cancelled && status.Signal() == syscall.SIGTERM))
		if !expected {
			r.stopErr = errors.Join(killErr, fmt.Errorf("holding supervisor exit is not the final owned signal: %v", r.waitErr))
			return r.stopErr
		}
		r.stopErr = killErr
	case <-timer.C:
		r.stopErr = errors.Join(killErr, fmt.Errorf("holding supervisor Wait unsettled after2s; no further group signal/probe/Wait is permitted"))
	}
	return r.stopErr
}

func (r *integrationFinalizationRun) output() string {
	read := func(name string) string {
		raw, err := os.ReadFile(filepath.Join(r.fx.root, name))
		if err != nil {
			return fmt.Sprintf("[%s unavailable: %v]", name, err)
		}
		return string(raw)
	}
	return read("supervisor.stdout") + read("supervisor.stderr")
}

// workflowExit observes only the private actual-inner-exit marker. A partial
// write is not completion; a complete malformed marker fails. This value is
// never derived from the forced supervisor SIGKILL/TERM Wait status.
func (r *integrationFinalizationRun) workflowExit() (int, bool, error) {
	raw, err := os.ReadFile(filepath.Join(r.fx.state, "workflow-exit"))
	if os.IsNotExist(err) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	if len(raw) > 4 {
		return 0, false, fmt.Errorf("malformed workflow exit marker %q", raw)
	}
	if len(raw) == 0 || raw[len(raw)-1] != '\n' {
		return 0, false, nil
	}
	value, err := strconv.Atoi(string(raw[:len(raw)-1]))
	if err != nil || value < 0 || value > 255 || string(raw) != strconv.Itoa(value)+"\n" {
		return 0, false, fmt.Errorf("malformed workflow exit marker %q", raw)
	}
	return value, true, nil
}

func (r *integrationFinalizationRun) await(t *testing.T, event string) {
	t.Helper()
	for {
		if err := r.ctx.Err(); err != nil {
			t.Fatalf("workflow did not reach %s within20s: %v", event, err)
		}
		if strings.Contains(r.fx.events(t), event+"\n") {
			if err := r.ctx.Err(); err != nil {
				t.Fatalf("event %s observed after bound: %v", event, err)
			}
			return
		}
		if status, complete, err := r.workflowExit(); err != nil {
			t.Fatal(err)
		} else if complete {
			t.Fatalf("inner workflow exited%d before %s\n%s", status, event, r.output())
		}
		select {
		case <-r.ctx.Done():
			t.Fatalf("workflow did not reach %s within20s", event)
		case <-time.After(5 * time.Millisecond):
		}
	}
}

func (r *integrationFinalizationRun) finish(t *testing.T, success bool) {
	t.Helper()
	for {
		if err := r.ctx.Err(); err != nil {
			t.Fatalf("inner workflow exceeded20s: %v", err)
		}
		status, complete, err := r.workflowExit()
		if err != nil {
			t.Fatal(err)
		}
		if complete {
			if err := r.ctx.Err(); err != nil {
				t.Fatalf("inner completion observed after bound: %v", err)
			}
			if err := r.stop(); err != nil {
				t.Fatal(err)
			}
			if (success && status != 0) || (!success && status == 0) {
				t.Fatalf("actual inner workflow exit%d, expected success=%v\n%s", status, success, r.output())
			}
			return
		}
		select {
		case <-r.ctx.Done():
			t.Fatal("inner workflow did not complete within20s")
		case <-time.After(5 * time.Millisecond):
		}
	}
}

func (r *integrationFinalizationRun) cancel(t *testing.T) {
	t.Helper()
	if r.signalsClosed || r.waitStarted {
		t.Fatal("cancellation after signaling closed")
	}
	if err := r.ctx.Err(); err != nil {
		t.Fatalf("cancellation after20s bound: %v", err)
	}
	if _, complete, err := r.workflowExit(); err != nil {
		t.Fatal(err)
	} else if complete {
		t.Fatal("inner workflow completed before requested cancellation")
	}
	// No Wait exists yet: the outer leader anchors this numeric group even
	// if it exits immediately after TERM. Only stop may begin the final reap.
	if err := syscall.Kill(-r.cmd.Process.Pid, syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	r.cancelled = true
}

func (r *integrationFinalizationRun) finishCancelled(t *testing.T) {
	t.Helper()
	if !r.cancelled {
		t.Fatal("cancellation completion without owned TERM")
	}
	if err := r.stop(); err != nil {
		t.Fatal(err)
	}
	status, complete, err := r.workflowExit()
	if err != nil {
		t.Fatal(err)
	}
	if complete && status == 0 {
		t.Fatal("canceled inner workflow recorded success")
	}
	// Absence is expected if TERM prevented the actual inner-exit marker.
	// No fake failure/completion marker is written by Go or cleanup.
}

func (fx *integrationFinalizationFixture) events(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(fx.state, "events"))
	if os.IsNotExist(err) {
		return ""
	}
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func (fx *integrationFinalizationFixture) completePrior(t *testing.T) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(fx.state, "xml-ready"), []byte("complete\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile(filepath.Join(fx.state, "events"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	_, writeErr := f.WriteString("prior-complete\n")
	closeErr := f.Close()
	if err := errors.Join(writeErr, closeErr); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(fx.state, "lock")); err != nil {
		t.Fatal(err)
	}
}

func (fx *integrationFinalizationFixture) assertAbsent(t *testing.T, events ...string) {
	t.Helper()
	got := fx.events(t)
	for _, event := range events {
		if strings.Contains(got, event+"\n") {
			t.Fatalf("unexpected successor %s:\n%s", event, got)
		}
	}
}

func (fx *integrationFinalizationFixture) assertOrder(t *testing.T, events ...string) {
	t.Helper()
	got, at := fx.events(t), -1
	for _, event := range events {
		next := strings.Index(got, event+"\n")
		if next <= at {
			t.Fatalf("missing/reordered event %s:\n%s", event, got)
		}
		at = next
	}
}

// This executes literal workflow bodies and clauses with private stand-ins.
// It proves scheduling/status only, not Actions signaling or Bazel settlement.
func TestBazelUnitResultFinalization(t *testing.T) {
	job := readCIWorkflow(t, bazelWorkflowName).job(t, bazelJobName)
	for _, mode := range []string{"success", "failure", "timeout", "skipped"} {
		t.Run(mode, func(t *testing.T) {
			fx := newUnitFinalizationFixture(t, job, mode)
			run := fx.start(t)
			run.finish(t, mode == "success" || mode == "skipped")
			switch mode {
			case "success":
				fx.assertOrder(t, "test-complete", "package", "profile", "equivalence", "sync-check", "sync", "raw-upload", "analytics", "result:success")
				fx.assertAbsent(t, "unavailable")
			case "failure", "timeout":
				fx.assertOrder(t, "test-enter", "unavailable", "unavailable:failure", "raw-upload", "result:failure")
				fx.assertAbsent(t, "package", "profile", "equivalence", "sync-check", "sync", "analytics", "result:success")
				if mode == "failure" {
					// Even a settled natural failure is conservatively unavailable:
					// step outcome alone does not distinguish it from a timeout.
					fx.assertOrder(t, "test-complete", "unavailable")
				} else {
					fx.assertAbsent(t, "test-complete")
				}
				if !strings.Contains(run.output(), "Bazel test completion settlement is not established") {
					t.Fatalf("missing explicit unavailable diagnostic:\n%s", run.output())
				}
			case "skipped":
				fx.assertAbsent(t, "test-enter", "unavailable", "package", "profile", "equivalence", "sync-check", "sync", "analytics", "raw-upload")
				fx.assertOrder(t, "result:success")
			}
		})
	}
	t.Run("cancel", func(t *testing.T) {
		fx := newUnitFinalizationFixture(t, job, "cancel")
		run := fx.start(t)
		run.await(t, "test-enter")
		run.cancel(t)
		run.finishCancelled(t)
		fx.assertAbsent(t, "test-complete", "package", "profile", "equivalence", "sync-check", "sync", "analytics", "raw-upload", "result:success", "result:failure")
	})
}

func unitFinalizationCondition(t *testing.T, expression string) string {
	t.Helper()
	expression = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(expression, "${{"), "}}"))
	switch expression {
	case "always() && steps.test.outcome == 'success'":
		return "[[ $test_outcome == 'success' ]]"
	case "always() && steps.test.outcome != 'skipped' && steps.test.outcome != 'success'":
		return "[[ $test_outcome != 'skipped' && $test_outcome != 'success' ]]"
	case "always() && steps.test.outcome != '' && steps.test.outcome != 'skipped'":
		return "[[ -n $test_outcome && $test_outcome != 'skipped' ]]"
	case "always() && steps.bazel.outcome == 'success' && steps.test.outcome == 'success'":
		return "[[ $bazel_outcome == 'success' && $test_outcome == 'success' ]]"
	case "!cancelled() && steps.test.outcome == 'success'":
		return "[[ $cancelled != 1 && $test_outcome == 'success' ]]"
	default:
		return integrationFinalizationCondition(t, expression)
	}
}

func newUnitFinalizationFixture(t *testing.T, job ciWorkflowJob, mode string) *integrationFinalizationFixture {
	t.Helper()
	// Reuse the existing process-group owner/deadline/private outputs and
	// single-final-signal/Wait implementation. No new supervisor is introduced.
	fx := &integrationFinalizationFixture{root: t.TempDir(), bash: requireHostTool(t, "bash")}
	fx.state = filepath.Join(fx.root, "state")
	bin, tmp := filepath.Join(fx.root, "bin"), filepath.Join(fx.root, "tmp")
	for _, dir := range []string{fx.state, bin, tmp, filepath.Join(tmp, "ci-analytics"), filepath.Join(fx.root, "scripts", "ci")} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"date", "tee", "sleep"} {
		if err := os.Symlink(requireHostTool(t, name), filepath.Join(bin, name)); err != nil {
			t.Fatal(err)
		}
	}
	write := func(path, body string) {
		t.Helper()
		if err := os.WriteFile(path, []byte("#!"+fx.bash+"\nset -eu\n"+body), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join(bin, "bazel"), `
[[ "$1" == test ]] || exit 70
printf '{}\n' > "$RUNNER_TEMP/bazel-bep.json"
printf 'test-enter\n' >> "$BD_STANDIN_STATE/events"
case "$BD_STANDIN_TEST" in
  cancel) while :; do sleep 1; done ;;
  timeout) exit 124 ;;
  success|failure)
    printf 'complete\n' > "$BD_STANDIN_STATE/xml-ready"
    printf 'test-complete\n' >> "$BD_STANDIN_STATE/events"
    [[ "$BD_STANDIN_TEST" == success ]] || exit 3 ;;
  *) exit 71 ;;
esac
`)
	write(filepath.Join(fx.root, "scripts", "ci", "package-bazel-bd.sh"), `
printf 'package\n' >> "$BD_STANDIN_STATE/events"
[[ -f "$BD_STANDIN_STATE/xml-ready" ]] || exit 72
`)
	write(filepath.Join(bin, "python3"), `
case "$1" in
  tools/bazel/critpath.py) printf 'profile\n' >> "$BD_STANDIN_STATE/events" ;;
  tools/bazel/equivalence.py) printf 'equivalence\n' >> "$BD_STANDIN_STATE/events" ;;
  tools/bazel/ci_analytics_extract.py) printf 'analytics\n' >> "$BD_STANDIN_STATE/events" ;;
  *) exit 74 ;;
esac
[[ -f "$BD_STANDIN_STATE/xml-ready" ]] || exit 73
`)
	write(filepath.Join(bin, "zstd"), "exit 0\n")
	write(filepath.Join(bin, "make"), `
case "$1" in
  bazel-sync-check) printf 'sync-check\n' >> "$BD_STANDIN_STATE/events" ;;
  bazel-sync) printf 'sync\n' >> "$BD_STANDIN_STATE/events" ;;
  *) exit 76 ;;
esac
[[ -f "$BD_STANDIN_STATE/xml-ready" ]] || exit 75
`)
	write(filepath.Join(bin, "git"), `
[[ "$*" == 'status --porcelain --untracked-files=all' ]] || exit 77
`)
	selected := map[string]string{
		"bazel test //... --config=ci":                         "test",
		"Final test diagnostics unavailable":                   "unavailable",
		"Package bd (ci-build-artifacts layout)":               "package",
		"Critical-path report":                                 "profile",
		"Go test equivalence":                                  "equivalence",
		"BUILD files in sync (gazelle, go_srcs, MODULE.bazel)": "sync",
		"Upload test logs":                                     "raw-upload",
		"CI analytics summary":                                 "analytics",
		"Record job result":                                    "result",
	}
	seen := map[string]bool{}
	var driver strings.Builder
	driver.WriteString("set -eu\njob_failed=0\ncancelled=0\nbazel_outcome=success\ntest_outcome=skipped\nnogo_outcome=skipped\n")
	for i, step := range job.Steps {
		role, wanted := selected[step.Name]
		if !wanted {
			continue
		}
		if seen[role] {
			t.Fatalf("duplicate selected unit workflow step: %s", role)
		}
		seen[role] = true
		if role == "test" && mode == "skipped" {
			continue // The scheduler skipped this step: execute no shell body.
		}
		fmt.Fprintf(&driver, "if %s; then\n", unitFinalizationCondition(t, step.If))
		if role == "raw-upload" {
			// An action has no Run body. Test its actual clause/order only;
			// record admission rather than pretending to upload anything.
			driver.WriteString("printf 'raw-upload\\n' >> \"$BD_STANDIN_STATE/events\"\nfi\n")
			continue
		}
		if step.Run == "" {
			t.Fatalf("selected unit workflow shell body is empty: %s", role)
		}
		path := filepath.Join(fx.root, fmt.Sprintf("unit-step-%d.sh", i))
		if err := os.WriteFile(path, []byte(step.Run), 0o600); err != nil {
			t.Fatal(err)
		}
		if role == "result" {
			driver.WriteString("if [[ $job_failed == 1 ]]; then export JOB_STATUS=failure; else export JOB_STATUS=success; fi\n")
		}
		if role == "unavailable" {
			driver.WriteString("printf 'unavailable\\n' >> \"$BD_STANDIN_STATE/events\"\n")
		}
		fmt.Fprintf(&driver, "if \"$BD_STANDIN_BASH\" -e -- %s; then step_outcome=success; else step_outcome=failure; job_failed=1; fi\n", integrationFinalizationQuote(path))
		if role == "test" {
			driver.WriteString("test_outcome=$step_outcome\n")
		}
		if role == "unavailable" {
			driver.WriteString("printf 'unavailable:%s\\n' \"$step_outcome\" >> \"$BD_STANDIN_STATE/events\"\n")
		}
		if role == "result" {
			driver.WriteString("printf 'result:%s\\n' \"$JOB_STATUS\" >> \"$BD_STANDIN_STATE/events\"\n")
		}
		driver.WriteString("fi\n")
	}
	for _, role := range selected {
		if !seen[role] {
			t.Fatalf("missing selected unit workflow step: %s", role)
		}
	}
	driver.WriteString("if [[ $job_failed == 1 ]]; then exit 1; fi\n")
	fx.script = filepath.Join(fx.root, "drive.sh")
	if err := os.WriteFile(fx.script, []byte(driver.String()), 0o600); err != nil {
		t.Fatal(err)
	}
	fx.env = []string{"PATH=" + bin, "HOME=" + fx.root, "RUNNER_TEMP=" + tmp,
		"GITHUB_STEP_SUMMARY=" + filepath.Join(fx.root, "summary"), "GITHUB_OUTPUT=" + filepath.Join(fx.root, "output"),
		"BD_STANDIN_STATE=" + fx.state, "BD_STANDIN_TEST=" + mode, "BD_STANDIN_BASH=" + fx.bash,
		"BAZEL_SOLE_RUN=", "BAZEL_FRESH=", "TIER=local", "DETAILS=", "LANE=test", "MODE=local", "CHECK_RUN_ID=1", "PR_HINT=363", "BASH_ENV=", "ENV="}
	return fx
}

// Stand-ins exercise the complete selected workflow shell body. This does not
// create host namespaces or establish the cause of a hosted policy refusal.
func TestHostedNamespacePreflightDiagnostics(t *testing.T) {
	job := readCIWorkflow(t, bazelWorkflowName).job(t, "bazel-proxied")
	var body string
	for _, step := range job.Steps {
		if step.Name == "Hosted offline namespace preflight" {
			if body != "" {
				t.Fatal("duplicate hosted namespace preflight")
			}
			body = step.Run
			if step.If != "${{ needs.rbe.outputs.local_full == 'true' }}" {
				t.Fatalf("changed admission clause: %q", step.If)
			}
		}
	}
	if body == "" {
		t.Fatal("missing hosted namespace preflight")
	}
	const isolation = `unshare --user --map-root-user --net -- true
unshare --user --map-root-user --net -- bash -euc '
  ip link set lo up
  if timeout 3 bash -c "exec 3<>/dev/tcp/1.1.1.1/443" 2>/dev/null; then
    echo "::error::outbound network is reachable inside the test namespace"
    exit 1
  fi
'
`
	if !strings.HasSuffix(body, isolation) {
		t.Fatal("original namespace/loopback/probe body changed")
	}
	for _, mode := range []string{"first-refusal", "positive", "outbound-refusal", "unavailable", "read-failure"} {
		t.Run(mode, func(t *testing.T) {
			fx := newNamespacePreflightFixture(t, body, mode)
			run := fx.start(t)
			run.finish(t, mode != "first-refusal" && mode != "outbound-refusal")
			status, complete, err := run.workflowExit()
			if err != nil || !complete {
				t.Fatalf("missing actual exit: %d %v %v", status, complete, err)
			}
			wantStatus := 0
			wantEvents := "first\nsecond\nloopback\nprobe\n"
			if mode == "first-refusal" {
				wantStatus, wantEvents = 1, "first\n"
			}
			if mode == "outbound-refusal" {
				wantStatus = 1
			}
			if status != wantStatus || fx.events(t) != wantEvents {
				t.Fatalf("exit/events: %d %q; want %d %q", status, fx.events(t), wantStatus, wantEvents)
			}
			out := run.output()
			if mode == "first-refusal" && !strings.Contains(out, "unshare: write failed /proc/self/uid_map: Operation not permitted") {
				t.Fatalf("lost original refusal: %s", out)
			}
			if mode == "outbound-refusal" && !strings.Contains(out, "outbound network is reachable inside the test namespace") {
				t.Fatalf("lost original outbound refusal: %s", out)
			}
			if strings.Contains(out, "SECRET_STATUS") || strings.Contains(out, "SECRET_ENV") || strings.Contains(out, strings.Repeat("V", 513)) {
				t.Fatalf("unbounded/unselected diagnostic output: %s", out)
			}
			switch mode {
			case "unavailable":
				if !strings.Contains(out, "unavailable") || !strings.Contains(out, "uid unavailable") {
					t.Fatalf("missing unavailable diagnostics: %s", out)
				}
			case "read-failure":
				if !strings.Contains(out, "namespace-policy "+filepath.Join(fx.root, "read-error")+"=\n") || strings.Contains(out, strings.Repeat("V", 512)) {
					t.Fatalf("diagnostic read failure not exercised: %s", out)
				}
			default:
				if !strings.Contains(out, "CapEff:") || !strings.Contains(out, strings.Repeat("V", 512)) {
					t.Fatalf("selected bounded diagnostics not exercised: %s", out)
				}
			}
		})
	}
}

func newNamespacePreflightFixture(t *testing.T, body, mode string) *integrationFinalizationFixture {
	t.Helper()
	fx := &integrationFinalizationFixture{root: t.TempDir(), bash: requireHostTool(t, "bash")}
	fx.state = filepath.Join(fx.root, "state")
	bin := filepath.Join(fx.root, "bin")
	for _, dir := range []string{fx.state, bin} {
		if err := os.Mkdir(dir, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(requireHostTool(t, "sleep"), filepath.Join(bin, "sleep")); err != nil {
		t.Fatal(err)
	}
	write := func(path, text string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(text), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	standin := func(name, text string) { write(filepath.Join(bin, name), "#!"+fx.bash+"\nset -eu\n"+text) }
	standin("id", `if [[ "$BD_STANDIN_TEST" == unavailable ]]; then exit 42; fi
case "$1" in -u|-g) printf '1001\n' ;; *) exit 70 ;; esac
`)
	standin("unshare", `[[ "$1 $2 $3 $4" == '--user --map-root-user --net --' ]] || exit 71
shift 4
if [[ "$1" == true && $# == 1 ]]; then
  printf 'first\n' >> "$BD_STANDIN_STATE/events"
  if [[ "$BD_STANDIN_TEST" == first-refusal ]]; then
    printf 'unshare: write failed /proc/self/uid_map: Operation not permitted\n' >&2
    exit 1
  fi
  exit 0
fi
[[ "$1" == bash && "$2" == -euc && $# == 3 ]] || exit 72
printf 'second\n' >> "$BD_STANDIN_STATE/events"
shift
exec "$BD_STANDIN_BASH" "$@"
`)
	standin("ip", `[[ "$*" == 'link set lo up' ]] || exit 73
printf 'loopback\n' >> "$BD_STANDIN_STATE/events"
`)
	standin("timeout", `[[ "$*" == '3 bash -c exec 3<>/dev/tcp/1.1.1.1/443' ]] || exit 74
printf 'probe\n' >> "$BD_STANDIN_STATE/events"
[[ "$BD_STANDIN_TEST" == outbound-refusal ]]
`)
	// Only copied-body proc paths are substituted; production has no test
	// override. No host proc/security state or real unshare/ip/network is used.
	status := filepath.Join(fx.root, "status")
	scalar := filepath.Join(fx.root, "scalar")
	write(status, "Name:\tSECRET_STATUS\nUid:\t1001 1001 1001 1001\nCapEff:\t0000000000000000\nSeccomp:\t2\n")
	write(scalar, strings.Repeat("V", 700)+"\n")
	if mode == "unavailable" {
		status, scalar = filepath.Join(fx.root, "missing-status"), filepath.Join(fx.root, "missing-scalar")
	}
	if mode == "read-failure" {
		// A readable directory deterministically fails bash read, including
		// for elevated test users; chmod000 would not prove that branch.
		scalar = filepath.Join(fx.root, "read-error")
		if err := os.Mkdir(scalar, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	body = strings.ReplaceAll(body, "/proc/$$/status", integrationFinalizationQuote(status))
	for _, path := range []string{"/proc/sys/user/max_user_namespaces", "/proc/sys/user/max_net_namespaces", "/proc/sys/kernel/unprivileged_userns_clone", "/proc/sys/kernel/apparmor_restrict_unprivileged_userns", "/proc/sys/kernel/apparmor_restrict_unprivileged_unconfined", "/proc/$$/attr/current"} {
		if !strings.Contains(body, path) {
			t.Fatalf("missing selected diagnostic path %s", path)
		}
		body = strings.ReplaceAll(body, path, integrationFinalizationQuote(scalar))
	}
	if strings.Contains(body, "/proc/") {
		t.Fatal("unreceived host proc read in copied preflight")
	}
	fx.script = filepath.Join(fx.root, "preflight.sh")
	write(fx.script, body)
	fx.env = []string{"PATH=" + bin, "HOME=" + fx.root, "BD_STANDIN_BASH=" + fx.bash, "BD_STANDIN_STATE=" + fx.state, "BD_STANDIN_TEST=" + mode, "UNRELATED_SECRET=SECRET_ENV", "BASH_ENV=", "ENV="}
	return fx
}
