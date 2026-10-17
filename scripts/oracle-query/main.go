// Command oracle-query prints, as JSON, the row an issue had at a given commit
// of a corpus clone: the read side of the replay harness. It is a thin wrapper
// over internal/replay/oracle, which holds all of the logic.
//
// Usage:
//
//	go run ./scripts/oracle-query -dir=<clone-dir> -ref=<commit-hash> -issue=<issue-id>
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/steveyegge/beads/internal/replay/oracle"
)

func main() {
	dir := flag.String("dir", "", "corpus clone directory produced by scripts/corpus-acquire (required)")
	ref := flag.String("ref", "", "commit hash to query the issue's row AS OF (required)")
	issue := flag.String("issue", "", "issue id to query (required)")
	flag.Parse()

	if *dir == "" || *ref == "" || *issue == "" {
		fmt.Fprintln(os.Stderr, "usage: oracle-query -dir=<clone-dir> -ref=<commit-hash> -issue=<issue-id>")
		os.Exit(2)
	}

	row, err := oracle.QueryAsOf(context.Background(), *dir, *ref, *issue)
	if err != nil {
		fmt.Fprintln(os.Stderr, "oracle-query:", err)
		os.Exit(1)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(row); err != nil {
		fmt.Fprintln(os.Stderr, "oracle-query: encoding result:", err)
		os.Exit(1)
	}
}
