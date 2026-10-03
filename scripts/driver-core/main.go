// Command driver-core replays a historical corpus's mutations, commit by
// commit, through an integration build of bd and compares each result against
// the corpus's own row: the driver of the replay harness. It is a thin wrapper
// over internal/replay/driver, which holds all of the logic.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/steveyegge/beads/internal/replay/driver"
)

func main() {
	integrationRef := flag.String("integration-ref", "HEAD", "git ref of the integration build under test")
	integrationRepo := flag.String("integration-repo", "", "path to the beads repo checkout to build the integration binary from (required)")
	oracleDataDir := flag.String("oracle-data-dir", "", "dolt data directory containing the historical corpus to replay (required)")
	workDir := flag.String("work-dir", "", "bd project directory to replay mutations into; created and initialized if it doesn't already exist (required)")
	outDir := flag.String("out-dir", "", "directory to write replay_runs/commit_replay_results/mismatches/metric_samples JSONL files to (required)")
	sampleSize := flag.Int("sample-size", 0, "number of evenly-spaced commits to sample; 0 replays the full history exhaustively")
	flag.Parse()

	if *integrationRepo == "" || *oracleDataDir == "" || *workDir == "" || *outDir == "" {
		fmt.Fprintln(os.Stderr, "usage: driver-core --integration-repo DIR --oracle-data-dir DIR --work-dir DIR --out-dir DIR [--integration-ref REF] [--sample-size N]")
		os.Exit(2)
	}

	replayRun, err := driver.Options{
		IntegrationRef:  *integrationRef,
		IntegrationRepo: *integrationRepo,
		OracleDataDir:   *oracleDataDir,
		WorkDir:         *workDir,
		OutDir:          *outDir,
		SampleSize:      *sampleSize,
	}.Run(context.Background())
	if err != nil {
		log.Fatalf("driver-core: %v", err)
	}

	fmt.Printf("replay run %s: status=%s mode=%s integration_sha=%s\n", replayRun.ID, replayRun.Status, replayRun.Mode, replayRun.IntegrationSHA)
}
