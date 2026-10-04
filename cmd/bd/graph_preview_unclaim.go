package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func runGraphPreviewUnclaim(cmd *cobra.Command, args []string) error {
	if err := graphPreviewWritePolicy(); err != nil {
		return err
	}
	if err := graphPreviewFlags(cmd); err != nil {
		return err
	}
	if len(args) != 1 {
		return graphFailure("invalid_selector", "graph unclaim requires one Bead ID or beads/PATH", 2)
	}
	path, err := graphPreviewResourcePath(graphPreviewConfig.GraphScopeURL, args[0])
	if err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	if err := graph.ValidateBeadPath(path); err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	actor := getActorWithGit()
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		result, err := store.UnclaimIssue(ctx, path, actor)
		if err != nil {
			return nil, "", err
		}
		return result, fmt.Sprintf("Unclaimed %s", result.Issue.ID), nil
	})
}
