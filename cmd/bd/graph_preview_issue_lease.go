package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
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
	leaseActor := graphPreviewCoordinationActor()
	selected, err := graphPreviewIssueSelector(selector)
	if err != nil {
		return graphFailure("invalid_selector", "heartbeat requires a local Issue ID or beads/PATH", 2)
	}
	var touched string
	err = withGraphFamiliarStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		issue, err := store.HeartbeatIssueSelected(ctx, selected, leaseActor)
		if err != nil {
			return nil, "", err
		}
		touched = issue.ID
		return graphHeartbeatReport(issue, leaseActor)
	})
	if err == nil {
		SetLastTouchedID(touched)
	}
	return err
}

func graphHeartbeatReport(issue graphstore.IssueRecord, actor string) (any, string, error) {
	return map[string]string{"id": issue.Properties.ID, "status": "heartbeat", "owner": actor},
		fmt.Sprintf("Heartbeat %s (lease refreshed)", issue.Properties.ID), nil
}

func runGraphPreviewReclaim(cmd *cobra.Command) error {
	if err := graphPreviewWritePolicy(); err != nil {
		return err
	}
	if err := graphPreviewFlags(cmd, "older-than", "id", "assignee", "label", "label-any", "exclude-label", "any-replica"); err != nil {
		return err
	}
	leaseActor := graphPreviewCoordinationActor()
	olderThan, _ := cmd.Flags().GetDuration("older-than")
	if olderThan < 0 {
		return graphFailure("invalid_properties", "--older-than must not be negative", 2)
	}
	filter, err := reclaimFilterFromFlags(cmd)
	if err != nil {
		return graphFailure("invalid_properties", err.Error(), 2)
	}
	// The native report describes the operator's scope, not the mapped Issue
	// IDs supplied internally to fence this preview to graph-owned rows.
	scoped := !filter.IsEmpty()
	// Parse every selector before storage opens. Mapping to native IDs happens
	// below through the graph catalog, never by assuming path == Issue ID.
	selectors := make([]graphstore.IssueSelector, len(filter.IDs))
	for i, selector := range filter.IDs {
		selected, err := graphPreviewIssueSelector(selector)
		if err != nil {
			return graphFailure("invalid_selector", "reclaim --id requires local Issue IDs or beads/PATH", 2)
		}
		selectors[i] = selected
	}
	return withGraphFamiliarStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		reclaimed, err := store.ReclaimExpiredIssueLeasesSelected(ctx, olderThan, filter, selectors, leaseActor)
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
		return graphReclaimReport(reclaimed, scoped, human)
	})
}

func graphReclaimReport(reclaimed []types.ReclaimedLease, scoped bool, human string) (any, string, error) {
	return map[string]any{"reclaimed": reclaimed, "count": len(reclaimed), "scoped": scoped}, human, nil
}
