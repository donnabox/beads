package main

import (
	"context"
	"fmt"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func runGraphPreviewIssueStatus() error {
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		stats, err := store.IssueStatistics(ctx)
		if err != nil {
			return nil, "", err
		}
		return map[string]any{"summary": stats, "blocked_count_skipped": false},
			fmt.Sprintf("Issue status: %d total, %d open, %d in progress, %d closed, %d blocked, %d ready",
				stats.TotalIssues, stats.OpenIssues, stats.InProgressIssues, stats.ClosedIssues, *stats.BlockedIssues, *stats.ReadyIssues), nil
	})
}
