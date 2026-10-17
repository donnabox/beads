// Command scenarios is a black-box scenario driver for Memory Beads on the
// fork's graph mode. It runs JSON scenario files against a real bd binary in
// two fresh hermetic workspaces, checks each scenario's expectations (declared
// from spec text, never recorded from observed behavior) on the first, and
// requires the normalized transcripts of both to be byte-identical.
//
// See README.md for the scenario format, the assertion vocabulary, the
// normalization rules, the exit-code contract and the guards.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Exit codes. The CI lane builds its matrix from --list and reads these.
const (
	exitOK       = 0 // every selected scenario is pass, xfail or a declared skip; discovered == executed; determinism verified
	exitFail     = 1 // a scenario failed, an xfail passed (xpass), or the two workspaces diverged after normalization
	exitHarness  = 2 // harness error or a guard tripped: bd missing, graph mode not active, unsafe workspace, timeout, bad usage
	exitManifest = 3 // a scenario file or the scenarios directory is invalid
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

type stringList []string

func (l *stringList) String() string     { return strings.Join(*l, ",") }
func (l *stringList) Set(v string) error { *l = append(*l, v); return nil }

// exitFor maps an error to the exit code of its class.
func exitFor(err error) int {
	var me *manifestError
	if errors.As(err, &me) {
		return exitManifest
	}
	return exitHarness
}

func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("scenarios", flag.ContinueOnError)
	fs.SetOutput(stderr)
	bd := fs.String("bd", "", "path of the bd binary under test (required unless --list); never looked up on PATH")
	out := fs.String("out", "", "directory for transcripts and receipts.json (required unless --list); must not exist or be empty")
	scenariosDir := fs.String("scenarios", "", "directory of scenario *.json files (default scripts/scenarios/scenarios under the repo root)")
	engine := fs.String("engine", "embedded", "storage engine; only embedded is supported")
	list := fs.Bool("list", false, "print one scenario id per line and exit")
	keep := fs.Bool("keep-workspaces", false, "keep the temporary workspaces for inspection")
	var only stringList
	fs.Var(&only, "scenario", "run only this scenario id (repeatable)")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return exitOK
		}
		return exitHarness
	}
	if fs.NArg() > 0 {
		fmt.Fprintf(stderr, "scenarios: unexpected argument %q\n", fs.Arg(0))
		return exitHarness
	}

	dir := *scenariosDir
	if dir == "" {
		dir = defaultScenariosDir()
	}
	if *list {
		scs, err := loadScenarios(dir)
		if err != nil {
			fmt.Fprintf(stderr, "scenarios: %v\n", err)
			return exitFor(err)
		}
		for _, sc := range scs {
			fmt.Fprintln(stdout, sc.ID)
		}
		return exitOK
	}
	if *bd == "" || *out == "" {
		fmt.Fprintln(stderr, "scenarios: --bd and --out are required (or --list)")
		return exitHarness
	}
	if *engine != "embedded" {
		fmt.Fprintf(stderr, "scenarios: engine %q is not supported: only embedded (the server engine is a follow-up)\n", *engine)
		return exitHarness
	}

	// The manifest is checked before bd is touched: a bad scenario must never
	// cost a workspace, and must not depend on a working bd to be reported.
	all, err := loadScenarios(dir)
	if err != nil {
		fmt.Fprintf(stderr, "scenarios: %v\n", err)
		return exitFor(err)
	}
	selected, err := selectScenarios(all, only)
	if err != nil {
		fmt.Fprintf(stderr, "scenarios: %v\n", err)
		return exitFor(err)
	}
	bdPath, err := resolveBD(*bd)
	if err == nil {
		err = prepareOut(*out)
	}
	if err != nil {
		fmt.Fprintf(stderr, "scenarios: %v\n", err)
		return exitFor(err)
	}

	code, err := execute(context.Background(), &driver{bd: bdPath, keep: *keep, stderr: stderr}, dir, *out, *engine, all, selected, len(only) == 0, stdout, stderr)
	if err != nil {
		fmt.Fprintf(stderr, "scenarios: %v\n", err)
		return exitFor(err)
	}
	return code
}

// execute runs the selected scenarios, writes the transcripts and receipts,
// and reports. exhaustive means no --scenario filter was given.
func execute(ctx context.Context, d *driver, dir, out, engine string, all, selected []*Scenario, exhaustive bool, stdout, stderr io.Writer) (int, error) {
	info, err := d.bdInfo(ctx)
	if err != nil {
		return 0, err
	}
	doc := &receiptsDoc{
		DriverVersion: driverVersion,
		Source:        gitProvenance(dir),
		BD:            info,
		Env:           receiptEnv(),
		ScenariosDir:  dir,
		Engine:        engine,
		Discovered:    len(all),
		Scenarios:     []scenarioReceipt{},
	}
	counts := map[outcome]int{}
	failing := false
	for _, sc := range selected {
		r, err := d.runScenario(ctx, sc)
		if err != nil {
			return 0, err
		}
		if r.Outcome != outcomeSkip {
			if err := writeTranscripts(out, sc.ID, "A", r.A); err != nil {
				return 0, harnessf("writing transcripts: %v", err)
			}
			if err := writeTranscripts(out, sc.ID, "B", r.B); err != nil {
				return 0, harnessf("writing transcripts: %v", err)
			}
		}
		doc.Executed++
		doc.Scenarios = append(doc.Scenarios, newReceipt(engine, r))
		counts[r.Outcome]++
		report(stdout, stderr, r)
		failing = failing || r.Outcome.failing()
	}
	if exhaustive && doc.Executed != doc.Discovered {
		return 0, harnessf("discovered %d scenarios but executed %d: the receipts are not exhaustive", doc.Discovered, doc.Executed)
	}
	if err := writeReceipts(out, doc); err != nil {
		return 0, harnessf("writing receipts: %v", err)
	}
	fmt.Fprintf(stdout, "scenarios: %d discovered, %d executed; pass=%d xfail=%d skip=%d fail=%d xpass=%d diverged=%d\n",
		doc.Discovered, doc.Executed, counts[outcomePass], counts[outcomeXfail], counts[outcomeSkip],
		counts[outcomeFail], counts[outcomeXpass], counts[outcomeDiverged])
	if failing {
		return exitFail, nil
	}
	return exitOK, nil
}

func report(stdout, stderr io.Writer, r *scenarioResult) {
	sc := r.Scenario
	switch r.Outcome {
	case outcomeSkip:
		fmt.Fprintf(stdout, "%s: skip (gap %s)\n", sc.ID, sc.Gap)
	case outcomeXfail:
		fmt.Fprintf(stdout, "%s: xfail (finding %s; %d steps, %.1fs)\n", sc.ID, sc.Finding, r.Steps, r.Duration.Seconds())
	default:
		fmt.Fprintf(stdout, "%s: %s (%d steps, %.1fs)\n", sc.ID, r.Outcome, r.Steps, r.Duration.Seconds())
	}
	if r.Outcome.failing() {
		fmt.Fprintf(stderr, "%s: %s: %s\n", sc.ID, r.Outcome, r.Detail)
	}
}

// selectScenarios applies --scenario. An id that is not in the directory is a
// manifest error, not a scenario that quietly runs nothing.
func selectScenarios(all []*Scenario, only []string) ([]*Scenario, error) {
	if len(only) == 0 {
		return all, nil
	}
	have := map[string]bool{}
	ids := make([]string, 0, len(all))
	for _, sc := range all {
		have[sc.ID] = true
		ids = append(ids, sc.ID)
	}
	want := map[string]bool{}
	var problems []string
	for _, id := range only {
		if !have[id] {
			problems = append(problems, fmt.Sprintf("--scenario %q is not a scenario in the directory (have: %s)", id, strings.Join(ids, ", ")))
		}
		want[id] = true
	}
	if len(problems) > 0 {
		return nil, &manifestError{problems}
	}
	var selected []*Scenario
	for _, sc := range all {
		if want[sc.ID] {
			selected = append(selected, sc)
		}
	}
	return selected, nil
}

// resolveBD makes --bd an explicit absolute path to an executable file. It is
// never resolved through PATH: a bare "bd" becomes ./bd, which must exist.
func resolveBD(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", harnessf("--bd %q: %v", path, err)
	}
	fi, err := os.Stat(abs)
	if err != nil {
		return "", harnessf("bd binary %s: %v", abs, err)
	}
	if !fi.Mode().IsRegular() || fi.Mode().Perm()&0o111 == 0 {
		return "", harnessf("bd binary %s is not an executable file", abs)
	}
	return abs, nil
}

// prepareOut creates --out, or accepts an existing empty directory. A
// non-empty one is refused so receipts from two runs cannot mix.
func prepareOut(out string) error {
	fi, err := os.Stat(out)
	switch {
	case err == nil:
		if !fi.IsDir() {
			return harnessf("--out %s exists and is not a directory", out)
		}
		entries, err := os.ReadDir(out)
		if err != nil {
			return harnessf("--out %s: %v", out, err)
		}
		if len(entries) > 0 {
			return harnessf("--out %s is not empty; use a fresh directory so receipts cannot mix runs", out)
		}
	case errors.Is(err, os.ErrNotExist):
		if err := os.MkdirAll(out, 0o750); err != nil {
			return harnessf("--out %s: %v", out, err)
		}
	default:
		return harnessf("--out %s: %v", out, err)
	}
	return nil
}

// defaultScenariosDir is scripts/scenarios/scenarios under the repo root, found
// by walking up from the working directory to the nearest go.mod. Outside a
// checkout it falls back to the relative path, which is then simply missing.
func defaultScenariosDir() string {
	rel := filepath.Join("scripts", "scenarios", "scenarios")
	dir, err := os.Getwd()
	if err != nil {
		return rel
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return filepath.Join(dir, rel)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return rel
		}
		dir = parent
	}
}
