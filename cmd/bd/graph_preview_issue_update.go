package main

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/internal/validation"
	publicops "github.com/steveyegge/beads/issueops"
)

var graphPreviewIssueEditFlags = []string{"title", "description", "body", "message", "design", "acceptance", "priority", "assignee", "notes", "clear-notes", "append-notes", "estimate", "external-ref", "spec-id", "due"}

func graphPreviewIssueEditFlagsChanged(cmd *cobra.Command) bool {
	for _, name := range graphPreviewIssueEditFlags {
		if cmd.Flags().Changed(name) {
			return true
		}
	}
	return false
}

// This narrow preview keeps familiar Issue field meanings while requiring the
// existing opaque graph guard for a canonical Resource selector. File/stdin and
// broader workflow flags remain explicit refusals; legacy routing is unchanged.
func graphPreviewIssueEditRequest(cmd *cobra.Command, path string) (graphstore.UpdateIssueRequest, error) {
	request := graphstore.UpdateIssueRequest{Path: path}
	allowed := append([]string{"if-revision", "unconditional", "force"}, graphPreviewIssueEditFlags...)
	if err := graphPreviewFlags(cmd, allowed...); err != nil {
		return request, err
	}
	if !graphPreviewIssueEditFlagsChanged(cmd) {
		return request, graphFailure("invalid_properties", "Issue update requires at least one supported field", 2)
	}
	revision, unconditional, err := graphPreviewRevisionGuard(cmd, false, true)
	if err != nil {
		return request, err
	}
	request.ExpectedRevision, request.Unconditional = revision, unconditional
	if cmd.Flags().Changed("notes") && (cmd.Flags().Changed("clear-notes") || cmd.Flags().Changed("append-notes")) ||
		cmd.Flags().Changed("clear-notes") && cmd.Flags().Changed("append-notes") {
		return request, graphFailure("invalid_properties", "--notes, --clear-notes and --append-notes are mutually exclusive", 2)
	}
	if cmd.Flags().Changed("clear-notes") && !clearNotesRequested(cmd) {
		return request, graphFailure("invalid_properties", "--clear-notes=false does not request an edit", 2)
	}
	if cmd.Flags().Changed("force") && !cmd.Flags().Changed("notes") {
		return request, graphFailure("invalid_properties", "graph Issue update accepts --force only with --notes", 2)
	}
	request.ForceNotesOverwrite, _ = cmd.Flags().GetBool("force")
	if clearNotesRequested(cmd) {
		cleared := ""
		request.Notes = &cleared
	}
	if cmd.Flags().Changed("due") {
		value, err := graphPreviewIssueDueInput(cmd)
		if err != nil {
			return request, err
		}
		request.DueAt = publicops.Field[*time.Time]{Set: true, Value: value}
	}

	if cmd.Flags().Changed("priority") {
		value, _ := cmd.Flags().GetString("priority")
		priority, err := validation.ValidatePriority(value)
		if err != nil {
			return request, graphFailure("invalid_properties", err.Error(), 2)
		}
		request.Priority = &priority
	}
	if cmd.Flags().Changed("estimate") {
		value, err := cmd.Flags().GetInt("estimate")
		if err != nil {
			return request, graphFailure("invalid_properties", err.Error(), 2)
		}
		if err := types.ValidateIssueEstimatedMinutes(&value); err != nil {
			return request, graphFailure("invalid_properties", err.Error(), 2)
		}
		request.EstimatedMinutes = &value
	}
	if cmd.Flags().Changed("assignee") {
		value, _ := cmd.Flags().GetString("assignee")
		if !utf8.ValidString(value) {
			return request, graphFailure("invalid_properties", "Issue assignee must be valid UTF-8", 2)
		}
		if err := types.CheckFieldLen("assignee", value); err != nil {
			return request, graphFailure("invalid_properties", err.Error(), 2)
		}
		// Preserve the existing update's literal value. A graph revision guard
		// does not grant permission to take another actor's active assignment.
		request.Assignee = &value
	}

	fields := []struct {
		name string
		out  **string
	}{
		{"title", &request.Title}, {"description", &request.Description},
		{"body", &request.Description}, {"message", &request.Description},
		{"design", &request.Design},
		{"acceptance", &request.AcceptanceCriteria},
		{"notes", &request.Notes},
		{"append-notes", &request.AppendNotes},
		{"external-ref", &request.ExternalRef}, {"spec-id", &request.SpecID},
	}
	for _, field := range fields {
		if !cmd.Flags().Changed(field.name) {
			continue
		}
		value, _ := cmd.Flags().GetString(field.name)
		if !utf8.ValidString(value) {
			return request, graphFailure("invalid_properties", "Issue "+field.name+" must be valid UTF-8", 2)
		}
		if field.name == "notes" {
			if err := validateNotesUpdate(value); err != nil {
				return request, graphFailure("invalid_properties", err.Error(), 2)
			}
		}
		if value == "-" && field.out == &request.Description {
			return request, graphFailure("capability_unavailable", "graph Issue update currently accepts inline text only; stdin and file sources are unavailable", 5)
		}
		if field.name == "title" {
			value = strings.TrimSpace(value)
			if err := types.ValidateIssueTitle(value); err != nil {
				return request, graphFailure("invalid_properties", err.Error(), 2)
			}
		}
		if *field.out != nil && **field.out != value {
			return request, graphFailure("invalid_properties", "description aliases must have identical values", 2)
		}
		*field.out = &value
	}
	return request, nil
}

func runGraphPreviewUpdateIssue(cmd *cobra.Command, path string) error {
	request, err := graphPreviewIssueEditRequest(cmd, path)
	if err != nil {
		return err
	}
	request.Actor = getActorWithGit()
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		result, err := store.UpdateIssue(ctx, request)
		if err != nil {
			return nil, "", err
		}
		verb := "Unchanged"
		if result.Changed {
			verb = "Updated"
		}
		return result, fmt.Sprintf("%s %s", verb, result.Issue.ID), nil
	})
}
