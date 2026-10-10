package main

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/spf13/cobra"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/graphpatch"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
)

// Issue patch uses the same Property Change evaluator as Memory and Link, but
// projects only fields owned by the native Issue update writer. Its final write
// still passes through that writer, retaining assignment fences and History.
func graphPreviewPatchIssueProperties(ctx context.Context, store *graphstore.Store, cmd *cobra.Command, input graphstore.MemoryPropertiesPatchRequest, issue graphstore.IssueRecord) (any, string, error) {
	if issue.Properties == nil {
		return nil, "", graphFailure("invalid_properties", "Issue properties are unavailable", 2)
	}
	before := graphPreviewIssuePatchProjection(issue.Properties)
	encoded, err := json.Marshal(before)
	if err != nil {
		return nil, "", err
	}
	properties, err := graph.NewProperties(encoded)
	if err != nil {
		return nil, "", graphFailure("invalid_properties", err.Error(), 2)
	}
	patch, err := graphpatch.Parse(input.Patch)
	if err != nil {
		return nil, "", graphFailure("invalid_properties", err.Error(), 2)
	}
	afterProperties, err := patch.Apply(properties)
	if err != nil {
		return nil, "", graphFailure("invalid_properties", err.Error(), 2)
	}
	var after map[string]any
	if err := json.Unmarshal(afterProperties.Bytes(), &after); err != nil {
		return nil, "", graphFailure("invalid_properties", err.Error(), 2)
	}
	changes := make(map[string]any)
	for name, value := range after {
		old, admitted := before[name]
		if !admitted {
			return nil, "", graphFailure("invalid_properties", "Issue property "+name+" is not writable through patch", 2)
		}
		if !reflect.DeepEqual(old, value) {
			changes[name] = value
		}
	}
	for name := range before {
		if _, present := after[name]; present {
			continue
		}
		switch name {
		case "title", "priority":
			return nil, "", graphFailure("invalid_properties", "Issue "+name+" cannot be removed", 2)
		case "estimated_minutes", "external_ref", "due_at":
			changes[name] = nil
		default:
			changes[name] = ""
		}
	}
	request := graphstore.UpdateIssueRequest{Path: input.Path, Actor: input.Actor, ExpectedRevision: input.ExpectedRevision, Unconditional: input.Unconditional, Metadata: input.Metadata}
	if request.ExpectedRevision == "" {
		// Omission means edit the current resource. Pin the pre-read snapshot so a
		// concurrent writer cannot turn the computed patch into a lost update.
		request.ExpectedRevision = issue.Revision
		request.Unconditional = false
	}
	if err := graphPreviewApplyIssueProperties(cmd, &request, changes); err != nil {
		return nil, "", err
	}
	result, err := store.UpdateIssue(ctx, request)
	if err != nil {
		return nil, "", err
	}
	verb := "Updated"
	if !result.Changed {
		verb = "Unchanged"
	}
	return result, fmt.Sprintf("%s %s", verb, result.Issue.ID), nil
}

func graphPreviewIssuePatchProjection(issue *types.Issue) map[string]any {
	projection := map[string]any{
		"title": issue.Title, "description": issue.Description, "design": issue.Design,
		"acceptance_criteria": issue.AcceptanceCriteria, "notes": issue.Notes,
		"spec_id": issue.SpecID, "assignee": issue.Assignee,
		"priority": issue.Priority, "estimated_minutes": issue.EstimatedMinutes,
		"external_ref": issue.ExternalRef,
	}
	if issue.DueAt == nil {
		projection["due_at"] = nil
	} else {
		projection["due_at"] = issue.DueAt.Format(time.RFC3339Nano)
	}
	return projection
}
