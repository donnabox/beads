package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// Match the existing single-Issue graph close route. Wider routing, batching and
// the future claim-fence contract remain outside this reversible preview.
func runGraphPreviewReopen(cmd *cobra.Command, args []string) error {
	if err := graphPreviewWritePolicy(); err != nil {
		return err
	}
	if err := graphPreviewFlags(cmd, "reason"); err != nil {
		return err
	}
	if len(args) != 1 {
		return graphFailure("invalid_selector", "graph reopen requires one Bead ID or beads/PATH", 2)
	}
	path, err := graphPreviewResourcePath(graphPreviewConfig.GraphScopeURL, args[0])
	if err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	if err := graph.ValidateBeadPath(path); err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	reason, _ := cmd.Flags().GetString("reason")
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		result, err := store.ReopenIssue(ctx, path, reason, getActorWithGit())
		if err != nil {
			return nil, "", err
		}
		if !result.Changed {
			return result, fmt.Sprintf("Unchanged %s (status %q)\n", result.Issue.ID, result.Issue.Properties.Status), nil
		}
		return result, fmt.Sprintf("Reopened %s\n", result.Issue.ID), nil
	})
}
