package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/ui"
	"github.com/steveyegge/beads/internal/validation"
)

// These familiar Issue entry points use the same graph Issue writer as update.
// The native writer remains responsible for no-ops and exactly-once History.
func runGraphPreviewAssign(cmd *cobra.Command, args []string) error {
	if err := graphPreviewWritePolicy(); err != nil {
		return err
	}
	if err := graphPreviewFlags(cmd, "force", "if-revision"); err != nil {
		return err
	}
	path, err := graphPreviewResourcePath(graphPreviewConfig.GraphScopeURL, args[0])
	if err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	revision, unconditional, err := graphPreviewEditRevisionGuard(cmd)
	if err != nil {
		return err
	}
	force, _ := cmd.Flags().GetBool("force")
	assignee := args[1]
	err = withGraphStoreOutput(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		result, err := store.UpdateIssue(ctx, graphstore.UpdateIssueRequest{
			Path: path, Actor: getActorWithGit(), ExpectedRevision: revision,
			Unconditional: unconditional, Assignee: &assignee, ForceAssigneeTransfer: force,
		})
		if err != nil {
			return nil, "", err
		}
		issue := *result.Issue.Properties
		issue.ID = path
		if assignee == "" {
			return &issue, fmt.Sprintf("%s Unassigned %s\n", ui.RenderPass("✓"), formatFeedbackID(issue.ID, issue.Title)), nil
		}
		return &issue, fmt.Sprintf("%s Assigned %s to %s\n", ui.RenderPass("✓"), formatFeedbackID(issue.ID, issue.Title), assignee), nil
	}, func(result any, human string) error {
		if jsonOutput {
			return outputJSON(result)
		}
		_, err := fmt.Fprint(cmd.OutOrStdout(), human)
		return err
	})
	if err == nil {
		SetLastTouchedID(path)
	}
	return err
}

func runGraphPreviewPriority(cmd *cobra.Command, args []string) error {
	if err := graphPreviewWritePolicy(); err != nil {
		return err
	}
	if err := graphPreviewFlags(cmd); err != nil {
		return err
	}
	path, err := graphPreviewResourcePath(graphPreviewConfig.GraphScopeURL, args[0])
	if err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	priority, err := validation.ValidatePriority(args[1])
	if err != nil {
		return graphFailure("invalid_properties", err.Error(), 2)
	}
	err = withGraphStoreOutput(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		result, err := store.UpdateIssue(ctx, graphstore.UpdateIssueRequest{
			Path: path, Actor: getActorWithGit(), Unconditional: true, Priority: &priority,
		})
		if err != nil {
			return nil, "", err
		}
		issue := *result.Issue.Properties
		issue.ID = path // returned IDs must resolve to this graph Resource
		return &issue, fmt.Sprintf("%s Set priority of %s to P%d\n", ui.RenderPass("✓"), formatFeedbackID(issue.ID, issue.Title), priority), nil
	}, func(result any, human string) error {
		if jsonOutput {
			return outputJSON(result)
		}
		_, err := fmt.Fprint(cmd.OutOrStdout(), human)
		return err
	})
	if err == nil {
		SetLastTouchedID(path)
	}
	return err
}

func runGraphPreviewNote(cmd *cobra.Command, args []string) error {
	if err := graphPreviewWritePolicy(); err != nil {
		return err
	}
	if err := graphPreviewFlags(cmd, "stdin", "file"); err != nil {
		return err
	}
	path, err := graphPreviewResourcePath(graphPreviewConfig.GraphScopeURL, args[0])
	if err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	note, err := requireTextFromSources("note text", "use positional args, --stdin, or --file", cmdTextSources(cmd, args[1:]))
	if err != nil {
		return graphFailure("invalid_properties", err.Error(), 2)
	}
	err = withGraphStoreOutput(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		result, err := store.UpdateIssue(ctx, graphstore.UpdateIssueRequest{
			Path: path, Actor: getActorWithGit(), Unconditional: true, AppendNotes: &note,
		})
		if err != nil {
			return nil, "", err
		}
		issue := *result.Issue.Properties
		issue.ID = path // backing IDs can name a different canonical Bead
		return &issue, fmt.Sprintf("%s Note added to %s\n", ui.RenderPass("✓"), formatFeedbackID(issue.ID, issue.Title)), nil
	}, func(result any, human string) error {
		if jsonOutput {
			return outputJSON(result)
		}
		_, err := fmt.Fprint(cmd.OutOrStdout(), human)
		return err
	})
	if err == nil {
		SetLastTouchedID(path)
	}
	return err
}
