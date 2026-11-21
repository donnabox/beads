// Command mutation-translator classifies, for one issue, the bd CLI
// invocation(s) that replay a historical dolt_log commit step's row-diff, and
// either prints them (--dry-run) or executes them against a working clone
// through the bd binary named by --bd. It never looks bd up on PATH. It is a thin
// wrapper over internal/replay/translate, which holds all of the logic.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/steveyegge/beads/internal/replay/translate"
)

func main() {
	dataDir := flag.String("data-dir", "", "dolt data directory to read history from (required)")
	workDir := flag.String("work-dir", "", "bd project directory to replay into (required)")
	from := flag.String("from", "", "source commit hash, exclusive (required)")
	to := flag.String("to", "", "target commit hash, inclusive (required)")
	issueID := flag.String("issue", "", "issue ID to classify and replay (required)")
	dryRun := flag.Bool("dry-run", false, "classify and print actions without executing them")
	bdPath := flag.String("bd", "", "bd binary that executes the actions (required unless --dry-run; never looked up on PATH)")
	flag.Parse()

	if *dataDir == "" || *workDir == "" || *from == "" || *to == "" || *issueID == "" || (!*dryRun && *bdPath == "") {
		fmt.Fprintln(os.Stderr, "usage: mutation-translator --data-dir DIR --work-dir DIR --from SHA --to SHA --issue ID (--dry-run | --bd PATH)")
		os.Exit(2)
	}

	ctx := context.Background()
	actions, err := translate.Classify(ctx, *dataDir, *from, *to, *issueID)
	if err != nil {
		log.Fatalf("classify: %v", err)
	}

	for _, a := range actions {
		if *dryRun {
			fmt.Println(strings.Join(a.Argv, " "))
			continue
		}
		if err := translate.ExecuteWith(ctx, *bdPath, *workDir, a); err != nil {
			log.Fatalf("execute %v: %v", a.Argv, err)
		}
	}
}
