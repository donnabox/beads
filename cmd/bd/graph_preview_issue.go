package main

import (
	"context"
	"fmt"
	"unicode/utf8"

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
	if err := graphPreviewFlags(cmd, "id", "title", "description", "body", "message", "type", "priority", "labels", "label", "design", "acceptance", "assignee", "estimate", "external-ref", "spec-id", "notes", "due"); err != nil {
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
	creator := getActorWithGit()
	request := publicops.CreateRequest{Actor: creator, Issue: &types.Issue{
		CreatedBy: creator, Owner: getOwner(),
		Title: title, Description: description, Status: types.StatusOpen,
		Priority: priority, IssueType: issueType, Labels: utils.NormalizeLabels(append(labels, aliasLabels...)),
	}}
	if err := graphPreviewIssueCreateFields(cmd, request.Issue); err != nil {
		return err
	}
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		record, err := store.CreateIssue(ctx, path, request)
		if err != nil {
			return nil, "", err
		}
		return record, fmt.Sprintf("Created %s\n", path), nil
	})
}

// Initial assignment is an open Issue property, not a claim. Keep optional
// estimate presence and the ordinary CLI's empty-external-reference convention.
func graphPreviewIssueCreateFields(cmd *cobra.Command, issue *types.Issue) error {
	for _, field := range []struct {
		name string
		dest *string
	}{
		{"design", &issue.Design}, {"acceptance", &issue.AcceptanceCriteria}, {"notes", &issue.Notes},
		{"assignee", &issue.Assignee}, {"spec-id", &issue.SpecID},
	} {
		value, _ := cmd.Flags().GetString(field.name)
		if !utf8.ValidString(value) {
			return graphFailure("invalid_properties", "Issue "+field.name+" must be valid UTF-8", 2)
		}
		*field.dest = value
	}
	if err := types.CheckFieldLen("assignee", issue.Assignee); err != nil {
		return graphFailure("invalid_properties", err.Error(), 2)
	}
	external, _ := cmd.Flags().GetString("external-ref")
	if !utf8.ValidString(external) {
		return graphFailure("invalid_properties", "Issue external-ref must be valid UTF-8", 2)
	}
	if external != "" {
		issue.ExternalRef = &external
	}
	if cmd.Flags().Changed("estimate") {
		value, err := cmd.Flags().GetInt("estimate")
		if err != nil {
			return graphFailure("invalid_properties", err.Error(), 2)
		}
		if err := types.ValidateIssueEstimatedMinutes(&value); err != nil {
			return graphFailure("invalid_properties", err.Error(), 2)
		}
		issue.EstimatedMinutes = &value
	}
	if cmd.Flags().Changed("due") {
		due, err := graphPreviewIssueDueInput(cmd)
		if err != nil {
			return err
		}
		issue.DueAt = due
	}
	return nil
}
