package main

import (
	"context"
	"fmt"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func runGraphPreviewIssueStatus() error {
	return withGraphFamiliarStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		stats, err := store.IssueStatistics(ctx)
		if err != nil {
			return nil, "", err
		}
		return &StatusOutput{Summary: stats},
			fmt.Sprintf("Issue status: %d total, %d open, %d in progress, %d closed, %d blocked, %d ready",
				stats.TotalIssues, stats.OpenIssues, stats.InProgressIssues, stats.ClosedIssues, *stats.BlockedIssues, *stats.ReadyIssues), nil
	})
}
