package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/ui"
)

func runGraphPreviewMergeSlot(cmd *cobra.Command) error {
	op := cmd.Name()
	write := op != "check"
	if write {
		if err := graphPreviewWritePolicy(); err != nil {
			return err
		}
	}
	if op == "acquire" || op == "release" {
		if err := graphPreviewFlags(cmd, "holder", "wait"); err != nil {
			return err
		}
	} else if err := graphPreviewFlags(cmd); err != nil {
		return err
	}
	leaseActor := graphPreviewCoordinationActor()

	var idleExit bool
	err := withGraphFamiliarStore(func(ctx context.Context, s *graphstore.Store) (any, string, error) {
		switch op {
		case "create":
			issue, err := s.MergeSlotCreate(ctx, leaseActor)
			if err != nil {
				return nil, "", err
			}
			commandDidWrite.Store(true)
			return map[string]any{"id": issue.Properties.ID, "status": string(issue.Properties.Status)},
				fmt.Sprintf("%s Created merge slot: %s", ui.RenderPass("✓"), issue.Properties.ID), nil
		case "check":
			status, err := s.MergeSlotCheck(ctx)
			if errors.Is(err, graphstore.ErrNotFound) {
				id, idErr := s.MergeSlotID(ctx)
				if idErr != nil {
					return nil, "", idErr
				}
				return map[string]any{"id": id, "available": false, "error": "not found"},
					fmt.Sprintf("Merge slot not found: %s\nRun 'bd merge-slot create' to create one.", id), nil
			}
			if err != nil {
				return nil, "", err
			}
			human := fmt.Sprintf("%s Merge slot available: %s", ui.RenderPass("✓"), status.SlotID)
			if !status.Available {
				human = fmt.Sprintf("%s Merge slot held: %s\n  Holder: %s", ui.RenderAccent("○"), status.SlotID, status.Holder)
				if len(status.Waiters) > 0 {
					human += fmt.Sprintf("\n  Waiters: %d", len(status.Waiters))
					for i, waiter := range status.Waiters {
						human += fmt.Sprintf("\n    %d. %s", i+1, waiter)
					}
				}
			}
			return map[string]any{"id": status.SlotID, "available": status.Available, "holder": nilIfEmpty(status.Holder), "waiters": status.Waiters}, human, nil
		case "acquire":
			holder := mergeSlotHolder
			if holder == "" {
				holder = leaseActor
			}
			if holder == "" {
				return nil, "", fmt.Errorf("no holder specified; use --holder or set BEADS_ACTOR env var")
			}
			result, changed, err := s.MergeSlotAcquire(ctx, holder, leaseActor, mergeSlotAddWaiter)
			if err != nil {
				return nil, "", err
			}
			if changed {
				commandDidWrite.Store(true)
			}
			if result.Acquired {
				return map[string]any{"id": result.SlotID, "acquired": true, "holder": holder},
					fmt.Sprintf("%s Acquired merge slot: %s\n  Holder: %s", ui.RenderPass("✓"), result.SlotID, holder), nil
			}
			idleExit = true
			if result.Waiting {
				return map[string]any{"id": result.SlotID, "acquired": false, "waiting": true, "holder": result.Holder, "position": result.Position},
					fmt.Sprintf("%s Slot held by %s, added to waiters queue (position %d)", ui.RenderAccent("○"), result.Holder, result.Position), nil
			}
			return map[string]any{"id": result.SlotID, "acquired": false, "holder": result.Holder},
				fmt.Sprintf("slot held by: %s\nUse --wait to add yourself to the waiters queue.", result.Holder), nil
		case "release":
			id, changed, err := s.MergeSlotRelease(ctx, mergeSlotHolder, leaseActor)
			if err != nil {
				return nil, "", err
			}
			if changed {
				commandDidWrite.Store(true)
			}
			return map[string]any{"id": id, "released": true}, fmt.Sprintf("%s Released merge slot: %s", ui.RenderPass("✓"), id), nil
		}
		return nil, "", fmt.Errorf("unsupported merge-slot command")
	})
	if err != nil {
		return err
	}
	if idleExit {
		return SilentExit()
	}
	return nil
}
