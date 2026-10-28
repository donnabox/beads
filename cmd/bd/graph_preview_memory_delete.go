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
	if !forget && graphPreviewDeleteSetRequested(cmd, args) {
		return runGraphPreviewDeleteIssueSet(cmd, args)
	}
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
	selectedArgument := args[0]
	if !forget {
		selectedArgument, err = graphPreviewWritableResourceArgument(selectedArgument)
		if err != nil {
			return err
		}
	}
	readPath, localID, err := graphPreviewResourceReadSelector(selectedArgument)
	if err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		var selectedPath string
		var current any
		var err error
		if forget {
			selectedPath, err = store.ResolveMemorySelector(ctx, args[0])
			if err == nil {
				current, err = store.Read(ctx, selectedPath)
			}
		} else {
			selectedPath, current, err = store.ReadResourceSelected(ctx, readPath, localID)
		}
		if err != nil {
			return nil, "", err
		}
		request.Path = selectedPath
		if !forget {
			if _, issue := current.(graphstore.IssueRecord); issue {
				issueRequest := graphstore.IssueDeleteRequest{Path: selectedPath,
					Actor: request.Actor, ExpectedRevision: request.ExpectedRevision,
					Unconditional: request.Unconditional, Preview: request.Preview}
				var result graphstore.IssueDeleteResult
				if issueRequest.Preview {
					result, err = store.PreviewDeleteIssueWithIncidentLinks(ctx, issueRequest)
				} else {
					result, err = store.DeleteIssueWithIncidentLinks(ctx, issueRequest)
				}
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
				if count := len(result.IncidentLinkIDs); count != 0 {
					if result.Preview {
						summary += fmt.Sprintf("; %d incident Link(s) would be removed", count)
					} else {
						summary += fmt.Sprintf("; %d incident Link(s) removed", count)
					}
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
		summary := fmt.Sprintf("%s %s at final live revision %s; identity and prior snapshots remain retained", verb, graphPreviewMemoryDisplayID(result.Memory), result.Memory.Revision)
		if result.Terminal != nil {
			summary += fmt.Sprintf("; deletion revision %s", result.Terminal.Revision)
		}
		return result, summary, nil
	})
}

func graphPreviewMemoryDeleteInput(cmd *cobra.Command, args []string, forget bool) (graphstore.MemoryDeleteRequest, error) {
	allowed := []string{"if-revision"}
	if !forget {
		allowed = append(allowed, "force", "dry-run")
	}
	if err := graphPreviewFlags(cmd, allowed...); err != nil {
		return graphstore.MemoryDeleteRequest{}, err
	}
	if len(args) != 1 {
		return graphstore.MemoryDeleteRequest{}, graphFailure("invalid_selector", "deletion requires exactly one Bead ID or beads/PATH", 2)
	}
	path, _, err := graphPreviewResourceReadSelector(args[0])
	if err != nil {
		return graphstore.MemoryDeleteRequest{}, graphFailure("invalid_selector", err.Error(), 2)
	}
	force, _ := cmd.Flags().GetBool("force")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	preview := !forget && (dryRun || !force)
	revision, unconditional, err := graphPreviewRevisionGuard(cmd, false, path, true)
	if err != nil {
		return graphstore.MemoryDeleteRequest{}, err
	}
	if !utf8.ValidString(revision) || len(revision) > graphstore.PreviewVersionTokenLimit {
		return graphstore.MemoryDeleteRequest{}, graphFailure("invalid_selector", fmt.Sprintf("--if-revision requires a UTF-8 token of at most %d bytes", graphstore.PreviewVersionTokenLimit), 2)
	}
	return graphstore.MemoryDeleteRequest{Path: path, ExpectedRevision: revision, Unconditional: unconditional, Preview: preview}, nil
}
