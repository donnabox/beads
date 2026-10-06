package main

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cobra"
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
	beadType, err := graphPreviewCreateBeadType(cmd, graphPreviewConfig.GraphScopeURL)
	if err != nil {
		return err
	}
	if beadType == graphstore.MemoryTypeURL(graphPreviewConfig.GraphScopeURL) {
		return runGraphPreviewCreateMemory(cmd, args)
	}
	if err := graphPreviewFlags(cmd, "bead-type", "id", "title", "description", "body", "message", "type", "priority", "labels", "label", "design", "acceptance", "assignee", "estimate", "external-ref", "spec-id", "notes", "due"); err != nil {
		return err
	}
	path, err := graphPreviewCreateBeadPath(cmd)
	if err != nil {
		return err
	}
	titleFlag, _ := cmd.Flags().GetString("title")
	title, err := resolveTitle(args, titleFlag, "", "")
	if err != nil {
		return graphFailure("invalid_properties", err.Error(), 2)
	}
	// Ordinary create treats "-" as stdin; this bounded graph route accepts
	// inline Issue text only and refuses before acquiring input or opening storage.
	for _, name := range []string{"description", "body", "message"} {
		value, _ := cmd.Flags().GetString(name)
		if cmd.Flags().Changed(name) && value == "-" {
			return graphFailure("capability_unavailable", "graph Issue create currently accepts inline description text only; stdin and file sources are unavailable", 5)
		}
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

// Nominal bead type is separate from the Issue classification selected by -t.
// Only the two installed graph descriptors have a create writer in this slice.
func graphPreviewCreateBeadType(cmd *cobra.Command, scope string) (string, error) {
	if !cmd.Flags().Changed("bead-type") {
		return graphstore.IssueTypeURL(scope), nil
	}
	selector, _ := cmd.Flags().GetString("bead-type")
	typeURL, err := graphPreviewTypeURL(scope, selector)
	if err != nil {
		return "", graphFailure("invalid_selector", err.Error(), 2)
	}
	if typeURL != graphstore.IssueTypeURL(scope) && typeURL != graphstore.MemoryTypeURL(scope) {
		return "", graphFailure("capability_unavailable", "graph create supports only the installed Issue and Memory bead types", 5)
	}
	return typeURL, nil
}

func runGraphPreviewCreateMemory(cmd *cobra.Command, args []string) error {
	title, body, err := graphPreviewCreateMemoryInput(cmd, args)
	if err != nil {
		return err
	}
	path, err := graphPreviewCreateBeadPath(cmd)
	if err != nil {
		return err
	}
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		record, err := store.Create(ctx, graphstore.CreateRequest{Path: path, Title: title, Body: body, Actor: getActorWithGit()})
		if err != nil {
			return nil, "", err
		}
		return record, fmt.Sprintf("Created %s\n", path), nil
	})
}

// Inline text only: validate aliases and explicit presence before the ordinary
// description helper can read stdin. Issue-only flags cannot be silently lost.
func graphPreviewCreateMemoryInput(cmd *cobra.Command, args []string) (string, string, error) {
	if err := graphPreviewFlags(cmd, "bead-type", "id", "title", "description", "body", "message"); err != nil {
		return "", "", err
	}
	if len(args) > 1 {
		return "", "", graphFailure("invalid_selector", "Memory create accepts at most one title", 2)
	}
	bodyPresent, bodyValue := false, ""
	for _, name := range []string{"description", "body", "message"} {
		if !cmd.Flags().Changed(name) {
			continue
		}
		value, _ := cmd.Flags().GetString(name)
		if value == "-" {
			return "", "", graphFailure("capability_unavailable", "graph Memory create accepts inline body text only; stdin and file sources are unavailable", 5)
		}
		if bodyPresent && value != bodyValue {
			return "", "", graphFailure("invalid_properties", "Memory create description/body/message aliases must have the same value", 2)
		}
		bodyPresent, bodyValue = true, value
	}
	body, _, err := getDescriptionFlag(cmd)
	if err != nil {
		return "", "", err
	}
	title, _ := cmd.Flags().GetString("title")
	if cmd.Flags().Changed("title") && strings.TrimSpace(title) == "" {
		return "", "", graphFailure("invalid_properties", "an explicit --title must be nonempty", 2)
	}
	if len(args) > 0 || cmd.Flags().Changed("title") {
		title, err = resolveTitle(args, title, "", "")
		if err != nil {
			return "", "", graphFailure("invalid_properties", err.Error(), 2)
		}
	} else if bodyPresent {
		title = graphPreviewMemoryTitle(body)
	} else {
		return "", "", graphFailure("invalid_properties", "Memory create requires a title or inline body", 2)
	}
	if !utf8.ValidString(title) || !utf8.ValidString(body) {
		return "", "", graphFailure("invalid_properties", "Memory title and body must be valid UTF-8", 2)
	}
	return title, body, nil
}
