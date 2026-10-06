package driver

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/steveyegge/beads/internal/replay/doltcli"
	"github.com/steveyegge/beads/internal/replay/oracle"
	"github.com/steveyegge/beads/internal/replay/translate"
)

// Tools pins the binary a Run invocation replays through: the integration
// build under test (NFR4). The read side and the translator are libraries, so
// there is nothing else to pin.
type Tools struct {
	IntegrationBin string
}

// RunConfig is everything one Run invocation needs: which integration build to
// test (IntegrationRef/IntegrationRepo), which historical corpus is the oracle
// (OracleDataDir), which fresh project to replay mutations into
// (WorkDir/WorkDataDir), where to persist results (OutDir), and the binary to
// replay through. SampleSize <= 0 means UC2's exhaustive loop; SampleSize > 0
// selects that many evenly-spaced commits (AF1 sampled mode).
type RunConfig struct {
	IntegrationRef  string
	IntegrationRepo string
	OracleDataDir   string
	WorkDir         string
	WorkDataDir     string
	OutDir          string
	SampleSize      int
	Tools           Tools
	// Observer, when set, is told about every (step, issue) result as it is
	// written. It may be nil.
	Observer Observer
	// Seed is the record of the seeding that gave the work clone the oracle's base
	// state, nil when the work clone was not seeded. A base that holds issues is
	// refused without one.
	Seed *SeedRecord
}

// ErrLegacyRowsPresent reports a work clone that holds issues created while
// versioned history was off. Such a row never records versions, so a run that
// finds one has not measured what it set out to. It is a fault of the harness
// and is never a finding about the build under test.
var ErrLegacyRowsPresent = errors.New("the work clone holds issues created while versioned history was off")

// Run replays the oracle's history into the work project, step by step along the
// first parent of each commit, and compares the work project's view of every
// touched issue with the oracle's, persisting a CommitReplayResult for every
// (step, issue) pair (and a Mismatch, a CoverageGapRow or a rejection where there
// is one) and a summary of what the run covered. A mismatch does not abort the run
// -- AC2 requires a result row for every pair the run visits, matched or not;
// only an infrastructure failure (a broken ref, a bd that cannot run or is killed,
// a write error) does.
//
// The work project has versioned history switched on before the first step, so
// everything the replay creates records versions. A base that already holds
// issues is refused, before anything is written, unless the work clone was
// seeded from it.
func Run(ctx context.Context, cfg RunConfig) (ReplayRun, error) {
	walk, err := ReadWalk(ctx, cfg.OracleDataDir)
	if err != nil {
		return ReplayRun{}, fmt.Errorf("run: %w", err)
	}
	if err := checkSeedGuard(ctx, cfg.OracleDataDir, walk, cfg.Seed); err != nil {
		return ReplayRun{}, fmt.Errorf("run: %w", err)
	}

	sha, err := runGit(ctx, cfg.IntegrationRepo, "rev-parse", cfg.IntegrationRef)
	if err != nil {
		return ReplayRun{}, fmt.Errorf("run: resolve integration ref %q: %w", cfg.IntegrationRef, err)
	}

	mode := "exhaustive"
	if cfg.SampleSize > 0 {
		mode = "sampled"
	}
	run := ReplayRun{
		ID:             GenerateRunID(),
		IntegrationRef: cfg.IntegrationRef,
		IntegrationSHA: sha,
		Mode:           mode,
		SampleSize:     cfg.SampleSize,
		StartedAt:      time.Now().UTC(),
		Status:         "running",
	}

	store, err := NewStore(cfg.OutDir)
	if err != nil {
		return run, fmt.Errorf("run: %w", err)
	}

	steps := walk.Steps(cfg.SampleSize)
	r := newRunner(run.ID, store, cfg.Observer)
	loopErr := replayHistory(ctx, cfg, r, steps)

	run.Status = "completed"
	if loopErr != nil {
		run.Status = "failed"
	}
	run.FinishedAt = time.Now().UTC()
	if err := store.WriteSummary(r.summarize(run, walk, steps, cfg.Seed)); err != nil {
		if loopErr == nil {
			loopErr = fmt.Errorf("writing the summary: %w", err)
			run.Status = "failed"
		} else {
			loopErr = fmt.Errorf("%w (and writing the summary: %v)", loopErr, err)
		}
	}
	if err := store.WriteReplayRun(run); err != nil {
		if loopErr == nil {
			return run, fmt.Errorf("run: writing completed replay run: %w", err)
		}
		return run, fmt.Errorf("run: %w (and failed to record failed status: %v)", loopErr, err)
	}
	if loopErr != nil {
		return run, fmt.Errorf("run: %w", loopErr)
	}
	return run, nil
}

// replayHistory switches versioned history on in the work clone, replays the
// steps, and checks that every issue it made records versions.
func replayHistory(ctx context.Context, cfg RunConfig, r *runner, steps []Step) error {
	if err := enableVersionedHistory(ctx, cfg); err != nil {
		return err
	}
	if err := r.runSteps(ctx, realEnv{cfg: cfg}, steps); err != nil {
		return err
	}
	return verifyNoLegacyRows(ctx, cfg)
}

// enableVersionedHistory switches versioned history on in the work clone through
// the bd under test. It must happen before the first issue is created: a row made
// while history is off is legacy for its whole life and never records.
func enableVersionedHistory(ctx context.Context, cfg RunConfig) error {
	if out, err := execBd(ctx, cfg.Tools.IntegrationBin, cfg.WorkDir, "config", "set", "versioned-history.enabled", "true"); err != nil {
		return fmt.Errorf("switching versioned history on in the work clone: %w\n%s", err, out)
	}
	return nil
}

// verifyNoLegacyRows is the check that the switch took. History that silently
// stayed off would leave every created row legacy and a run that measured nothing
// it claims to, so a legacy row is a harness error.
func verifyNoLegacyRows(ctx context.Context, cfg RunConfig) error {
	const query = "SELECT COUNT(*) FROM issues WHERE participation_generation IS NULL"
	_, rows, err := doltcli.Query(ctx, cfg.WorkDataDir, query)
	if err != nil {
		return fmt.Errorf("checking the work clone for legacy rows: %w", err)
	}
	if len(rows) != 1 || len(rows[0]) != 1 {
		return fmt.Errorf("checking the work clone for legacy rows: unexpected result %v", rows)
	}
	n, err := strconv.Atoi(rows[0][0].Text)
	if err != nil {
		return fmt.Errorf("checking the work clone for legacy rows: %q is not a count", rows[0][0].Text)
	}
	if n > 0 {
		return fmt.Errorf("%w: %d issues", ErrLegacyRowsPresent, n)
	}
	return nil
}

// realEnv is the step environment of a real run: the oracle's dolt database, the
// work clone's, and the bd under test.
type realEnv struct {
	cfg RunConfig
}

func (e realEnv) Plan(ctx context.Context, from, to string) (*translate.StepPlan, error) {
	return translate.Plan(ctx, e.cfg.OracleDataDir, from, to)
}

func (e realEnv) OracleView(ctx context.Context, ref, issue string) (*oracle.View, error) {
	return oracle.ReadView(ctx, e.cfg.OracleDataDir, ref, issue)
}

func (e realEnv) Execute(ctx context.Context, action translate.Action) error {
	return translate.ExecuteWith(ctx, e.cfg.Tools.IntegrationBin, e.cfg.WorkDir, action)
}

func (e realEnv) WorkHead(ctx context.Context) (string, error) {
	return headCommitOf(ctx, e.cfg.WorkDataDir)
}

func (e realEnv) CandidateView(ctx context.Context, ref, issue string) (*oracle.View, error) {
	return oracle.ReadView(ctx, e.cfg.WorkDataDir, ref, issue)
}

// selectSample returns all of steps, copied, when n<=0 or n>=len(steps)
// (UC2's exhaustive loop); otherwise it returns exactly n elements as a
// deterministic, strictly-ascending-index subsequence of steps (AF1's
// evenly-spaced sampled mode) -- never reordered, never randomized.
func selectSample[T any](steps []T, n int) []T {
	if n <= 0 || n >= len(steps) {
		out := make([]T, len(steps))
		copy(out, steps)
		return out
	}
	if n == 1 {
		return []T{steps[0]}
	}
	last := len(steps) - 1
	out := make([]T, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, steps[i*last/(n-1)])
	}
	return out
}

// headCommitOf returns dataDir's current head commit hash, as dolt itself names
// it, rather than as the newest row of dolt_log, which promises no order.
func headCommitOf(ctx context.Context, dataDir string) (string, error) {
	_, rows, err := doltcli.Query(ctx, dataDir, "SELECT hashof('HEAD')")
	if err != nil {
		return "", fmt.Errorf("head commit of %s: %w", dataDir, err)
	}
	if len(rows) != 1 || len(rows[0]) != 1 || rows[0][0].Text == "" {
		return "", fmt.Errorf("head commit of %s: dolt returned no head", dataDir)
	}
	return rows[0][0].Text, nil
}
