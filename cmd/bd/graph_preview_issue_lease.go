package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
)

func runGraphPreviewHeartbeat(cmd *cobra.Command, selector string) error {
	if err := graphPreviewWritePolicy(); err != nil {
		return err
	}
	if err := graphPreviewFlags(cmd); err != nil {
		return err
	}
	path, err := graphPreviewResourcePath(graphPreviewConfig.GraphScopeURL, selector)
	if err != nil || graph.ValidateBeadPath(path) != nil {
		return graphFailure("invalid_selector", "heartbeat requires a local Issue ID or beads/PATH", 2)
	}
	err = withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		issue, err := store.HeartbeatIssue(ctx, path, actor)
		if err != nil {
			return nil, "", err
		}
		return map[string]string{"id": issue.ID, "status": "heartbeat", "owner": actor},
			fmt.Sprintf("Heartbeat %s (lease refreshed)", issue.ID), nil
	})
	if err == nil {
		SetLastTouchedID(path)
	}
	return err
}

func runGraphPreviewReclaim(cmd *cobra.Command) error {
	if err := graphPreviewWritePolicy(); err != nil {
		return err
	}
	if err := graphPreviewFlags(cmd, "older-than", "id", "assignee", "label", "label-any", "exclude-label", "any-replica"); err != nil {
		return err
	}
	olderThan, _ := cmd.Flags().GetDuration("older-than")
	if olderThan < 0 {
		return graphFailure("invalid_properties", "--older-than must not be negative", 2)
	}
	filter, err := reclaimFilterFromFlags(cmd)
	if err != nil {
		return graphFailure("invalid_properties", err.Error(), 2)
	}
	// Parse every selector before storage opens. Mapping to native IDs happens
	// below through the graph catalog, never by assuming path == Issue ID.
	paths := make([]string, len(filter.IDs))
	for i, selector := range filter.IDs {
		path, err := graphPreviewResourcePath(graphPreviewConfig.GraphScopeURL, selector)
		if err != nil || graph.ValidateBeadPath(path) != nil {
			return graphFailure("invalid_selector", "reclaim --id requires local Issue IDs or beads/PATH", 2)
		}
		paths[i] = path
	}
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		if len(paths) > 0 {
			filter.IDs = make([]string, 0, len(paths))
			for _, path := range paths {
				issue, err := store.ShowIssue(ctx, path)
				if err != nil {
					return nil, "", err
				}
				filter.IDs = append(filter.IDs, issue.Properties.ID)
			}
		}
		reclaimed, err := store.ReclaimExpiredIssueLeases(ctx, olderThan, filter, actor)
		if err != nil {
			return nil, "", err
		}
		if reclaimed == nil {
			reclaimed = []types.ReclaimedLease{}
		}
		human := "No stale leases to reclaim"
		if len(reclaimed) > 0 {
			ids := make([]string, 0, len(reclaimed))
			for _, item := range reclaimed {
				ids = append(ids, item.ID)
			}
			human = fmt.Sprintf("Reclaimed %d stale-lease Issue(s): %s", len(reclaimed), strings.Join(ids, ", "))
		}
		return map[string]any{"reclaimed": reclaimed, "count": len(reclaimed), "scoped": !filter.IsEmpty()}, human, nil
	})
}
