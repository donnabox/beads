package main

import (
	"context"
	"fmt"
	"unicode/utf8"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// Delete dispatches by immutable Bead backing; forget remains Memory-only.
// The default preview is a read; neither --force nor forget implies a cascade
// or a missing guard.
func runGraphPreviewDeleteMemory(cmd *cobra.Command, args []string, forget bool) error {
	request, err := graphPreviewMemoryDeleteInput(cmd, args, forget)
	if err != nil {
		return err
	}
	if !request.Preview {
		if err := graphPreviewWritePolicy(); err != nil {
			return err
		}
	}
	request.Actor = getActorWithGit()
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		if !forget {
			current, err := store.Read(ctx, request.Path)
			if err != nil {
				return nil, "", err
			}
			if _, issue := current.(graphstore.IssueRecord); issue {
				result, err := store.DeleteIssue(ctx, graphstore.IssueDeleteRequest(request))
				if err != nil {
					return nil, "", err
				}
				verb := "Deleted"
				if result.Preview {
					verb = "Would delete"
				}
				summary := fmt.Sprintf("%s %s at final live revision %s; identity and prior snapshots remain retained", verb, result.Issue.ID, result.Issue.Revision)
				if result.Terminal != nil {
					summary = fmt.Sprintf("%s %s at deletion revision %s; final live revision %s remains retained", verb, result.Issue.ID, result.Terminal.Revision, result.Issue.Revision)
				}
				return result, summary, nil
			}
		}
		result, err := store.DeleteMemory(ctx, request)
		if err != nil {
			return nil, "", err
		}
		verb := "Deleted"
		if result.Preview {
			verb = "Would delete"
		}
		summary := fmt.Sprintf("%s %s at final live revision %s; identity and prior snapshots remain retained", verb, result.Memory.ID, result.Memory.Revision)
		if result.Terminal != nil {
			summary += fmt.Sprintf("; deletion revision %s", result.Terminal.Revision)
		}
		return result, summary, nil
	})
}

func graphPreviewMemoryDeleteInput(cmd *cobra.Command, args []string, forget bool) (graphstore.MemoryDeleteRequest, error) {
	allowed := []string{"if-revision"}
	if !forget {
		allowed = append(allowed, "force")
	}
	if err := graphPreviewFlags(cmd, allowed...); err != nil {
		return graphstore.MemoryDeleteRequest{}, err
	}
	if len(args) != 1 {
		return graphstore.MemoryDeleteRequest{}, graphFailure("invalid_selector", "deletion requires exactly one Bead ID or beads/PATH", 2)
	}
	selector, err := graphPreviewWritableIdentity(args[0])
	if err != nil {
		return graphstore.MemoryDeleteRequest{}, graphFailure("invalid_selector", err.Error(), 2)
	}
	path, err := graphPreviewBeadSelector(selector)
	if err != nil {
		return graphstore.MemoryDeleteRequest{}, err
	}
	force, _ := cmd.Flags().GetBool("force")
	preview := !forget && !force
	revision, unconditional, err := graphPreviewRevisionGuard(cmd, false)
	if err != nil {
		return graphstore.MemoryDeleteRequest{}, err
	}
	if !utf8.ValidString(revision) || len(revision) > graphstore.PreviewVersionTokenLimit {
		return graphstore.MemoryDeleteRequest{}, graphFailure("invalid_selector", fmt.Sprintf("--if-revision requires a UTF-8 token of at most %d bytes", graphstore.PreviewVersionTokenLimit), 2)
	}
	return graphstore.MemoryDeleteRequest{Path: path, ExpectedRevision: revision, Unconditional: unconditional, Preview: preview}, nil
}
