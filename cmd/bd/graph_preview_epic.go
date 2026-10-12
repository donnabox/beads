package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
)

func runGraphPreviewEpicStatus(cmd *cobra.Command) error {
	if err := graphPreviewFlags(cmd, "eligible-only"); err != nil {
		return err
	}
	eligibleOnly, _ := cmd.Flags().GetBool("eligible-only")
	return withGraphStoreOutput(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		statuses, err := store.EpicStatuses(ctx)
		return statuses, "", err
	}, func(result any, _ string) error {
		return renderEpicStatus(result.([]*types.EpicStatus), eligibleOnly)
	})
}

type graphPreviewEpicCloseResult struct {
	eligible []*types.EpicStatus
	closed   []string
	errors   []error
	dryRun   bool
	reason   string
}

func runGraphPreviewEpicCloseEligible(cmd *cobra.Command) error {
	if err := graphPreviewFlags(cmd, "dry-run", "reason"); err != nil {
		return err
	}
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	if !dryRun {
		if err := graphPreviewWritePolicy(); err != nil {
			return err
		}
	}
	reason := resolveCloseEligibleReason(cmd)
	if err := validateCloseReasons([]string{reason}); err != nil {
		return err
	}
	return withGraphStoreOutput(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		statuses, err := store.EpicStatuses(ctx)
		if err != nil {
			return nil, "", err
		}
		result := graphPreviewEpicCloseResult{eligible: filterEligibleEpics(statuses), dryRun: dryRun, reason: reason}
		if dryRun {
			return result, "", nil
		}
		for _, status := range result.eligible {
			path := status.Epic.ID
			current, err := store.ShowIssue(ctx, path)
			if err == nil {
				_, err = store.CloseEligibleEpic(ctx, graphstore.EpicCloseRequest{
					Path: path, ExpectedRevision: current.Revision, Reason: reason, Actor: "system",
				})
			}
			if err != nil {
				result.errors = append(result.errors, fmt.Errorf("%s: %w", path, err))
				continue
			}
			result.closed = append(result.closed, path)
		}
		return result, "", nil
	}, func(raw any, _ string) error {
		result := raw.(graphPreviewEpicCloseResult)
		if len(result.eligible) == 0 {
			return outputNoEligibleEpics(result.reason)
		}
		if result.dryRun {
			return outputCloseEligibleDryRun(result.eligible, result.reason)
		}
		for _, err := range result.errors {
			fmt.Fprintf(os.Stderr, "Error closing %v\n", err)
		}
		if len(result.closed) > 0 {
			commandDidWrite.Store(true)
		}
		return outputCloseEligibleResult(result.closed, result.reason)
	})
}
