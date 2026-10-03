package driver

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/steveyegge/beads/internal/replay/doltcli"
)

// Options are the settings a driver-core command line carries.
type Options struct {
	IntegrationRef  string // git ref of the integration build under test
	IntegrationRepo string // checkout to build the integration binary from
	OracleDataDir   string // dolt data directory holding the historical corpus
	WorkDir         string // bd project directory to replay into
	OutDir          string // directory for the JSONL result files
	SampleSize      int    // evenly-spaced commits to sample; 0 replays everything
}

// Run builds the integration binary under test, prepares the work project, and
// replays the oracle's history into it. The oracle read and the translator are
// libraries called in process, so the integration binary is the only thing
// built.
func (o Options) Run(ctx context.Context) (ReplayRun, error) {
	toolsDir, err := os.MkdirTemp("", "driver-core-tools-*")
	if err != nil {
		return ReplayRun{}, fmt.Errorf("creating tools dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(toolsDir) }()

	integrationBin := filepath.Join(toolsDir, "integration-bd")
	if _, err := BuildIntegration(ctx, o.IntegrationRepo, o.IntegrationRef, "./cmd/bd", integrationBin); err != nil {
		return ReplayRun{}, fmt.Errorf("building integration binary: %w", err)
	}

	if err := ensureWorkProject(ctx, o.WorkDir, integrationBin); err != nil {
		return ReplayRun{}, fmt.Errorf("preparing work project: %w", err)
	}
	workDataDir, err := findEmbeddedDoltDir(o.WorkDir)
	if err != nil {
		return ReplayRun{}, fmt.Errorf("locating work project's data dir: %w", err)
	}

	return Run(ctx, RunConfig{
		IntegrationRef:  o.IntegrationRef,
		IntegrationRepo: o.IntegrationRepo,
		OracleDataDir:   o.OracleDataDir,
		WorkDir:         o.WorkDir,
		WorkDataDir:     workDataDir,
		OutDir:          o.OutDir,
		SampleSize:      o.SampleSize,
		Tools:           Tools{IntegrationBin: integrationBin},
	})
}

// ensureWorkProject initializes workDir as a bd project via integrationBin if
// it isn't one already, so a caller can point driver-core at a fresh empty
// directory without a separate manual bootstrap step.
func ensureWorkProject(ctx context.Context, workDir, integrationBin string) error {
	if _, err := os.Stat(filepath.Join(workDir, ".beads")); err == nil {
		return nil
	}
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return fmt.Errorf("ensure work project: %w", err)
	}
	cmd := exec.CommandContext(ctx, integrationBin, "init", "--non-interactive", "--role=maintainer")
	cmd.Dir = workDir
	cmd.Env = doltcli.SanitizedEnv(os.Environ())
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ensure work project: init: %w\n%s", err, out)
	}
	return nil
}

// findEmbeddedDoltDir returns the embedded Dolt data directory bd init created
// under dir. Glob's "*" also matches a sibling ".lock" file, so matches must be
// filtered to directories.
func findEmbeddedDoltDir(dir string) (string, error) {
	matches, err := filepath.Glob(filepath.Join(dir, ".beads", "embeddeddolt", "*"))
	if err != nil {
		return "", fmt.Errorf("find embedded dolt dir under %s: %w", dir, err)
	}
	for _, m := range matches {
		if info, statErr := os.Stat(m); statErr == nil && info.IsDir() {
			return m, nil
		}
	}
	return "", fmt.Errorf("find embedded dolt dir under %s: no directory among matches=%v", dir, matches)
}
