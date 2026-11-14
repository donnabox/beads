package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/internal/utils"
	"github.com/steveyegge/beads/internal/validation"
	publicops "github.com/steveyegge/beads/issueops"
)

// Quick capture keeps the ordinary one-line ID output while the native Issue
// writer owns creation, its initial History entry, and the graph mapping.
func runGraphPreviewQuick(cmd *cobra.Command, args []string) error {
	if err := graphPreviewWritePolicy(); err != nil {
		return err
	}
	if err := graphPreviewFlags(cmd, "priority", "type", "labels", "parent"); err != nil {
		return err
	}
	parent, _ := cmd.Flags().GetString("parent")
	if parent != "" {
		return graphFailure("capability_unavailable", "graph quick capture does not yet support --parent; use bd create --parent PARENT in a fresh graph workspace", 5)
	}
	// Ordinary q joins its positional arguments without create's flag-looking
	// title check. Cobra's -- terminator permits titles beginning with '-'.
	title := strings.Join(args, " ")
	if strings.TrimSpace(title) == "" {
		return graphFailure("invalid_properties", "title cannot be empty or whitespace-only", 2)
	}
	priorityText, _ := cmd.Flags().GetString("priority")
	priority, err := validation.ValidatePriority(priorityText)
	if err != nil {
		return graphFailure("invalid_properties", err.Error(), 2)
	}
	classification, _ := cmd.Flags().GetString("type")
	issueType := types.IssueType(classification).Normalize()
	storageClass, err := resolveStorageClass("", issueType)
	if err != nil {
		return graphFailure("invalid_properties", err.Error(), 2)
	}
	if storageClass.Normalize() != "" {
		return graphFailure("capability_unavailable", "graph Issue preview requires versioned storage; configured ephemeral or unversioned storage is unsupported", 5)
	}
	labels, _ := cmd.Flags().GetStringSlice("labels")
	creator := getActorWithGit()
	request := publicops.CreateRequest{Actor: creator, Issue: &types.Issue{
		CreatedBy: creator, Owner: getOwner(), Title: title,
		Status: types.StatusOpen, Priority: priority, IssueType: issueType,
		Labels: utils.NormalizeLabels(labels),
	}}
	path := ""
	err = withGraphStoreOutput(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		record, err := store.CreateIssue(ctx, "", request)
		if err != nil {
			return nil, "", err
		}
		path = graphPreviewDisplayLocalURL(record.ID)
		return record, path, nil
	}, func(_ any, id string) error {
		_, err := fmt.Fprintln(cmd.OutOrStdout(), id)
		return err
	})
	if err == nil {
		SetLastTouchedID(path)
	}
	return err
}
