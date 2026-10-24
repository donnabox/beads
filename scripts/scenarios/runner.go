package main

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
)

// outcome is what happened to one scenario.
type outcome string

const (
	outcomePass     outcome = "pass"
	outcomeFail     outcome = "fail"     // an expectation did not hold
	outcomeXfail    outcome = "xfail"    // declared xfail, failed as declared
	outcomeXpass    outcome = "xpass"    // declared xfail, but it passed
	outcomeSkip     outcome = "skip"     // declared gap, not run
	outcomeDiverged outcome = "diverged" // the two workspaces disagree after normalization
)

// failing reports the outcomes that make the run exit 1.
func (o outcome) failing() bool {
	return o == outcomeFail || o == outcomeXpass || o == outcomeDiverged
}

// scenarioResult is everything the summary and the receipts need about one scenario.
type scenarioResult struct {
	Scenario *Scenario
	Outcome  outcome
	Detail   string // why it failed, diverged or unexpectedly passed
	Steps    int    // steps executed in workspace A
	A, B     []*stepResult
	Equal    bool
	Duration time.Duration
	Captures map[string]string // stable captures of workspace A
}

// stepFailure is a scenario failing, as opposed to the driver failing.
type stepFailure struct {
	Step string
	Err  error
}

func (f *stepFailure) String() string {
	if f.Step == "" {
		return f.Err.Error()
	}
	return fmt.Sprintf("step %q: %v", f.Step, f.Err)
}

// wsRun is one workspace's execution of a scenario, before normalization.
type wsRun struct {
	raw      []*stepResult
	captured map[string]string // every capture, for ${var} and assertions
	stable   map[string]string // only the captures that are not random per workspace
	volatile [][]capturedValue // per step, the token and id captures for the normalizer
}

// capture records what a step captured. A capture that cannot be read is a
// scenario failure: later steps depend on it.
func (r *wsRun) capture(st *Step, res *stepResult) ([]capturedValue, error) {
	names := make([]string, 0, len(st.Capture))
	for name := range st.Capture {
		names = append(names, name)
	}
	sort.Strings(names)
	var volatile []capturedValue
	for _, name := range names {
		c := st.Capture[name]
		var value string
		switch {
		case c.jp != nil:
			v, err := pathScalar(res.Stdout, c.jp)
			if err != nil {
				return volatile, fmt.Errorf("capture %s: %w", name, err)
			}
			value = v
		case c.Stdout:
			value = string(res.Stdout)
		case c.StderrCode:
			value = stderrCode(res.Stderr)
		}
		r.captured[name] = value
		switch kind := captureKind(c.Kind); kind {
		case kindToken, kindID:
			volatile = append(volatile, capturedValue{Kind: kind, Value: value})
		default:
			r.stable[name] = value
		}
	}
	return volatile, nil
}

// runSteps executes the scenario's steps in w. With evaluate set it checks each
// step's expectations against the raw result and stops at the first failure;
// without it the steps only run, up to limit (limit < 0 means all of them).
// The first return is what ran; the second a scenario failure; the third a
// driver failure.
func (d *driver) runSteps(ctx context.Context, sc *Scenario, w *workspace, inputs map[string][]byte, limit int, evaluate bool) (*wsRun, *stepFailure, error) {
	run := &wsRun{captured: map[string]string{}, stable: map[string]string{}}
	env := &evalEnv{Steps: map[string]*stepResult{}, Captures: run.captured, Inputs: inputs}
	steps := sc.Steps
	if limit >= 0 && limit < len(steps) {
		steps = steps[:limit]
	}
	for i := range steps {
		st := &steps[i]
		argv := make([]string, 0, len(st.Argv)+2)
		for _, arg := range st.Argv {
			v, err := expandVars(arg, run.captured)
			if err != nil {
				return run, &stepFailure{Step: st.Name, Err: err}, nil
			}
			argv = append(argv, v)
		}
		if st.Actor != "" {
			argv = append(argv, "--actor", st.Actor)
		}
		var stdin []byte
		if st.Stdin != nil {
			b, err := st.Stdin.bytes()
			if err != nil {
				return run, nil, harnessf("scenario %s step %q: %v", sc.ID, st.Name, err)
			}
			stdin = b
		}
		res, err := w.exec(ctx, d.bd, argv, stdin)
		if err != nil {
			return run, nil, err
		}
		res.Name = st.Name
		run.raw = append(run.raw, res)
		env.Steps[st.Name] = res
		volatile, err := run.capture(st, res)
		run.volatile = append(run.volatile, volatile)
		if err != nil {
			return run, &stepFailure{Step: st.Name, Err: err}, nil
		}
		if evaluate {
			for _, a := range st.Expect {
				if err := a.eval(res, env); err != nil {
					return run, &stepFailure{Step: st.Name, Err: err}, nil
				}
			}
		}
	}
	return run, nil, nil
}

// driver holds what every scenario run shares.
type driver struct {
	bd     string
	keep   bool
	stderr io.Writer
}

// wsOutcome is one workspace's run, normalized while its root still exists (so
// both spellings of the root can be masked) and before it is removed.
type wsOutcome struct {
	failure *stepFailure
	order   error // recordedAt ordering violation
	norm    []*stepResult
	stable  map[string]string
}

func (d *driver) release(w *workspace) {
	if d.keep {
		fmt.Fprintf(d.stderr, "kept workspace %s\n", w.Root)
		return
	}
	if err := w.close(); err != nil {
		fmt.Fprintf(d.stderr, "warning: %v\n", err)
	}
}

func (d *driver) runIn(ctx context.Context, sc *Scenario, inputs map[string][]byte, limit int, evaluate bool) (*wsOutcome, error) {
	w, err := newWorkspace()
	if err != nil {
		return nil, harnessf("cannot create a workspace: %v", err)
	}
	defer d.release(w)
	if err := w.initGraph(ctx, d.bd); err != nil {
		return nil, err
	}
	run, failure, err := d.runSteps(ctx, sc, w, inputs, limit, evaluate)
	if err != nil {
		return nil, err
	}
	n := newNormalizer(w.Root)
	norm := make([]*stepResult, len(run.raw))
	for i, raw := range run.raw {
		norm[i] = n.normalize(raw, run.volatile[i])
	}
	return &wsOutcome{failure: failure, order: checkRecordedAtOrder(run.raw), norm: norm, stable: run.stable}, nil
}

// runScenario runs sc into workspace A, checking its expectations on A's raw
// values, then replays the same steps into workspace B. The two normalized
// transcripts must be byte-identical: that is the determinism check.
func (d *driver) runScenario(ctx context.Context, sc *Scenario) (*scenarioResult, error) {
	start := time.Now()
	res := &scenarioResult{Scenario: sc}
	if sc.State == "skip" {
		res.Outcome = outcomeSkip
		res.Detail = "declared gap " + sc.Gap
		return res, nil
	}
	inputs, err := sc.inputs()
	if err != nil {
		return nil, harnessf("scenario %s: %v", sc.ID, err)
	}
	a, err := d.runIn(ctx, sc, inputs, -1, true)
	if err != nil {
		return nil, err
	}
	b, err := d.runIn(ctx, sc, inputs, len(a.norm), false)
	if err != nil {
		return nil, err
	}
	res.A, res.B, res.Steps = a.norm, b.norm, len(a.norm)
	res.Captures = a.stable
	res.Duration = time.Since(start)

	failure := a.failure
	if failure == nil && a.order != nil {
		failure = &stepFailure{Err: a.order}
	}
	if failure == nil && b.order != nil {
		failure = &stepFailure{Err: fmt.Errorf("workspace B: %w", b.order)}
	}

	if detail := diverged(a, b); detail != "" {
		res.Outcome, res.Detail = outcomeDiverged, detail
		return res, nil
	}
	res.Equal = true
	switch {
	case failure != nil && sc.State == "xfail":
		res.Outcome, res.Detail = outcomeXfail, failure.String()
	case failure != nil:
		res.Outcome, res.Detail = outcomeFail, failure.String()
	case sc.State == "xfail":
		res.Outcome = outcomeXpass
		res.Detail = fmt.Sprintf("declared xfail (finding %s) but every expectation held: remove the xfail or update the finding", sc.Finding)
	default:
		res.Outcome = outcomePass
	}
	return res, nil
}

// diverged explains how two workspaces differ, or returns "" when they agree.
func diverged(a, b *wsOutcome) string {
	// B does not check expectations, so it only fails on a capture it cannot
	// read. That is consistent with A when A failed at the very same step; it is
	// a divergence when A read the value and B could not.
	if b.failure != nil && (a.failure == nil || a.failure.Step != b.failure.Step) {
		return fmt.Sprintf("workspaces diverged: workspace B failed at %v where workspace A did not", b.failure)
	}
	if len(a.norm) != len(b.norm) {
		return fmt.Sprintf("workspaces diverged: A ran %d steps, B ran %d", len(a.norm), len(b.norm))
	}
	for i := range a.norm {
		if stepDigest(a.norm[i]) == stepDigest(b.norm[i]) {
			continue
		}
		var fields []string
		if strings.Join(a.norm[i].Argv, "\x00") != strings.Join(b.norm[i].Argv, "\x00") {
			fields = append(fields, "argv")
		}
		if a.norm[i].Exit != b.norm[i].Exit {
			fields = append(fields, "exit")
		}
		if string(a.norm[i].Stdout) != string(b.norm[i].Stdout) {
			fields = append(fields, "stdout")
		}
		if string(a.norm[i].Stderr) != string(b.norm[i].Stderr) {
			fields = append(fields, "stderr")
		}
		return fmt.Sprintf("workspaces diverged after normalization at step %02d-%s: %s differs",
			i+1, a.norm[i].Name, strings.Join(fields, ", "))
	}
	return ""
}
