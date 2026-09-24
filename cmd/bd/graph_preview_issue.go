package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/config"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/internal/utils"
	"github.com/steveyegge/beads/internal/validation"
	publicops "github.com/steveyegge/beads/issueops"
)

// This disposable adapter preserves the Issue writer and classification rules.
// Its canonical path is distinct from the internal backing ID; no alias or
// production nominal-Type contract is established by this preview.
func runGraphPreviewCreateIssue(cmd *cobra.Command, args []string) error {
	if err := graphPreviewWritePolicy(); err != nil {
		return err
	}
	if err := graphPreviewFlags(cmd, "id", "title", "description", "body", "message", "type", "priority", "labels", "label"); err != nil {
		return err
	}
	path, _ := cmd.Flags().GetString("id")
	if err := graph.ValidateBeadPath(path); err != nil {
		return graphFailure("invalid_selector", "graph create requires --id beads/PATH: "+err.Error(), 2)
	}
	titleFlag, _ := cmd.Flags().GetString("title")
	title, err := resolveTitle(args, titleFlag, "", "")
	if err != nil {
		return graphFailure("invalid_properties", err.Error(), 2)
	}
	description, _, err := getDescriptionFlag(cmd)
	if err != nil {
		return err
	}
	if description == "" && config.GetBool("create.require-description") {
		return graphFailure("invalid_properties", "description is required by create.require-description", 2)
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
	aliasLabels, _ := cmd.Flags().GetStringSlice("label")
	request := publicops.CreateRequest{Actor: getActorWithGit(), Issue: &types.Issue{
		Title: title, Description: description, Status: types.StatusOpen,
		Priority: priority, IssueType: issueType, Labels: utils.NormalizeLabels(append(labels, aliasLabels...)),
	}}
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		record, err := store.CreateIssue(ctx, path, request)
		if err != nil {
			return nil, "", err
		}
		return record, fmt.Sprintf("Created %s\n", path), nil
	})
}
