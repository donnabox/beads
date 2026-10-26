package main

import (
	"context"
	"errors"
	"fmt"
	graph "github.com/steveyegge/beads/graphops"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/internal/utils"
	"github.com/steveyegge/beads/internal/validation"
	publicops "github.com/steveyegge/beads/issueops"
)

var graphPreviewIssueEditFlags = []string{"title", "description", "body", "message", "body-file", "description-file", "stdin", "design", "acceptance", "priority", "assignee", "notes", "clear-notes", "append-notes", "estimate", "external-ref", "spec-id", "due", "status", "add-label", "remove-label", "set-labels"}

func graphPreviewDescriptionFlagsChanged(cmd *cobra.Command) bool {
	for _, name := range []string{"description", "body", "message", "body-file", "description-file"} {
		if cmd.Flags().Changed(name) {
			return true
		}
	}
	stdin, _ := cmd.Flags().GetBool("stdin")
	return stdin
}

func graphPreviewIssueEditFlagsChanged(cmd *cobra.Command) bool {
	for _, name := range graphPreviewIssueEditFlags {
		if cmd.Flags().Changed(name) {
			return true
		}
	}
	return false
}

// This narrow preview keeps familiar Issue field meanings while requiring the
// existing opaque graph guard for a canonical Resource selector. Broader
// workflow flags remain explicit refusals; legacy routing is unchanged.
func graphPreviewIssueEditRequest(cmd *cobra.Command, path string, identity ...string) (graphstore.UpdateIssueRequest, error) {
	return graphPreviewIssueEditRequestWithParent(cmd, path, false, identity...)
}

func graphPreviewIssueEditRequestWithParent(cmd *cobra.Command, path string, allowParent bool, identity ...string) (graphstore.UpdateIssueRequest, error) {
	request := graphstore.UpdateIssueRequest{Path: path}
	allowed := append([]string{"type", "if-revision", "if-status", "if-assignee", "force", "session", "metadata", "replace-metadata", "set-metadata", "unset-metadata", "properties", "allow-empty-description"}, graphPreviewIssueEditFlags...)
	if allowParent {
		allowed = append(allowed, "parent")
	}
	if err := graphPreviewFlags(cmd, allowed...); err != nil {
		return request, err
	}
	if !graphPreviewIssueEditFlagsChanged(cmd) && !graphPreviewMetadataFlagsChanged(cmd) && !cmd.Flags().Changed("properties") && !cmd.Flags().Changed("type") && !allowParent {
		return request, graphFailure("invalid_properties", "Issue update requires at least one supported field", 2)
	}
	var err error
	request.TypeURL, err = graphPreviewNativeIssueTypeFlag(cmd, graphPreviewConfig.GraphScopeURL)
	if err != nil {
		return request, err
	}
	metadata, err := graphPreviewMetadataPatch(cmd)
	if err != nil {
		return request, err
	}
	request.Metadata = metadata
	if cmd.Flags().Changed("set-labels") {
		values, _ := cmd.Flags().GetStringSlice("set-labels")
		request.Labels.Replace = publicops.Field[[]string]{Set: true, Value: utils.NormalizeLabels(values)}
		warnLabelsContainingWhitespace(request.Labels.Replace.Value)
	}
	if cmd.Flags().Changed("add-label") {
		values, _ := cmd.Flags().GetStringSlice("add-label")
		request.Labels.Add = utils.NormalizeLabels(values)
		warnLabelsContainingWhitespace(request.Labels.Add)
	}
	if cmd.Flags().Changed("remove-label") {
		values, _ := cmd.Flags().GetStringSlice("remove-label")
		request.Labels.Remove = utils.NormalizeLabels(values)
	}
	revisionIdentity, privatePath := path, true
	if len(identity) != 0 {
		revisionIdentity, privatePath = identity[0], false
	}
	revision, unconditional, err := graphPreviewEditRevisionGuard(cmd, revisionIdentity, privatePath)
	if err != nil {
		return request, err
	}
	request.ExpectedRevision, request.Unconditional = revision, unconditional
	if cmd.Flags().Changed("if-assignee") {
		value, _ := cmd.Flags().GetString("if-assignee")
		if !utf8.ValidString(value) {
			return request, graphFailure("invalid_properties", "expected Issue assignee must be valid UTF-8", 2)
		}
		request.ExpectedAssignee = &value
	}
	if cmd.Flags().Changed("if-status") {
		status, _ := cmd.Flags().GetString("if-status")
		if status == "" || !utf8.ValidString(status) {
			return request, graphFailure("invalid_properties", "expected Issue status must be a nonempty UTF-8 name", 2)
		}
		value := types.Status(status)
		request.ExpectedStatus = &value
	}
	if cmd.Flags().Changed("notes") && (cmd.Flags().Changed("clear-notes") || cmd.Flags().Changed("append-notes")) ||
		cmd.Flags().Changed("clear-notes") && cmd.Flags().Changed("append-notes") {
		return request, graphFailure("invalid_properties", "--notes, --clear-notes and --append-notes are mutually exclusive", 2)
	}
	if cmd.Flags().Changed("clear-notes") && !clearNotesRequested(cmd) {
		return request, graphFailure("invalid_properties", "--clear-notes=false does not request an edit", 2)
	}
	force, _ := cmd.Flags().GetBool("force")
	if force && !cmd.Flags().Changed("notes") && !cmd.Flags().Changed("properties") && !cmd.Flags().Changed("assignee") && !cmd.Flags().Changed("status") {
		return request, graphFailure("invalid_properties", "graph Issue update accepts --force only with notes replacement, assignee, or status", 2)
	}
	request.ForceNotesOverwrite = force && cmd.Flags().Changed("notes")
	if cmd.Flags().Changed("status") {
		status, _ := cmd.Flags().GetString("status")
		if status == "" || !utf8.ValidString(status) {
			return request, graphFailure("invalid_properties", "Issue status must be a nonempty UTF-8 name", 2)
		}
		value := types.Status(status)
		request.Status = &value
		request.ForceClosePolicy = force
		if status == string(types.StatusClosed) {
			session, _ := cmd.Flags().GetString("session")
			if session == "" {
				session = os.Getenv("CLAUDE_SESSION_ID")
			}
			if session != "" {
				request.ClosedBySession = &session
			}
		} else if cmd.Flags().Changed("session") {
			return request, graphFailure("invalid_properties", "--session requires --status=closed", 2)
		}
	} else if cmd.Flags().Changed("session") {
		return request, graphFailure("invalid_properties", "--session requires --status=closed", 2)
	}
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
		// A matching assignee guard authorizes a transfer in the native writer.
		// It is mutually exclusive with the unconditional force-transfer bit.
		request.ForceAssigneeTransfer = force && request.ExpectedAssignee == nil
	}
	if graphPreviewDescriptionFlagsChanged(cmd) {
		// Keep graph's structured conflict error for incompatible inline aliases;
		// the shipping parser otherwise writes directly to stderr.
		var prior *string
		for _, name := range []string{"description", "body", "message"} {
			if !cmd.Flags().Changed(name) {
				continue
			}
			value, _ := cmd.Flags().GetString(name)
			if prior != nil && *prior != value {
				return request, graphFailure("invalid_properties", "description aliases must have identical values", 2)
			}
			prior = &value
		}
		value, changed, err := getDescriptionFlag(cmd)
		if err != nil {
			return request, err
		}
		if err := validateDescriptionUpdate(cmd, value, changed); err != nil {
			return request, graphFailure("invalid_properties", err.Error(), 2)
		}
		if changed {
			if !utf8.ValidString(value) {
				return request, graphFailure("invalid_properties", "Issue description must be valid UTF-8", 2)
			}
			request.Description = &value
		}
	}

	fields := []struct {
		name string
		out  **string
	}{
		{"title", &request.Title},
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

// The JSON object names native Issue properties, while the existing Issue
// writer keeps ownership of validation, no-ops, History and assignment fences.
func graphPreviewIssuePropertyMerge(cmd *cobra.Command, request *graphstore.UpdateIssueRequest) error {
	values, err := graphPreviewIssuePropertyValues(cmd)
	if err != nil || values == nil {
		return err
	}
	return graphPreviewApplyIssueProperties(cmd, request, values)
}

func graphPreviewIssuePropertyValues(cmd *cobra.Command) (map[string]any, error) {
	if !cmd.Flags().Changed("properties") {
		return nil, nil
	}
	input, _ := cmd.Flags().GetString("properties")
	if input == "-" && graphPreviewDescriptionUsesStdin(cmd) {
		return nil, graphFailure("invalid_properties", "--properties - and a description source cannot both read stdin", 2)
	}
	values, err := graphPreviewProperties(input, cmd.InOrStdin())
	if err != nil {
		return nil, graphFailure("invalid_properties", err.Error(), 2)
	}
	if _, ok := values["description"]; ok && graphPreviewDescriptionFlagsChanged(cmd) {
		return nil, graphFailure("invalid_properties", "Issue description was supplied by both --properties and a shorthand flag", 2)
	}
	return values, nil
}

func graphPreviewDescriptionUsesStdin(cmd *cobra.Command) bool {
	if stdin, _ := cmd.Flags().GetBool("stdin"); stdin {
		return true
	}
	for _, name := range []string{"description", "body", "message", "body-file", "description-file"} {
		if cmd.Flags().Changed(name) {
			value, _ := cmd.Flags().GetString(name)
			if value == "-" {
				return true
			}
		}
	}
	return false
}

func graphPreviewApplyIssueProperties(cmd *cobra.Command, request *graphstore.UpdateIssueRequest, values map[string]any) error {
	request.PropertiesProvided = true
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		value := values[name]
		if dest := map[string]**string{
			"title": &request.Title, "description": &request.Description,
			"design": &request.Design, "acceptance_criteria": &request.AcceptanceCriteria,
			"notes": &request.Notes, "assignee": &request.Assignee,
			"spec_id": &request.SpecID, "external_ref": &request.ExternalRef,
		}[name]; dest != nil {
			if *dest != nil || name == "notes" && request.AppendNotes != nil {
				return graphFailure("invalid_properties", "Issue "+name+" was supplied by both --properties and a shorthand flag", 2)
			}
			if name == "description" && (cmd.Flags().Changed("body") || cmd.Flags().Changed("message")) {
				return graphFailure("invalid_properties", "Issue description was supplied by both --properties and a shorthand flag", 2)
			}
			if value == nil && name == "external_ref" {
				cleared := ""
				*dest = &cleared
				continue
			}
			text, ok := value.(string)
			if !ok {
				return graphFailure("invalid_properties", "Issue "+name+" must be a JSON string", 2)
			}
			if name == "title" {
				text = strings.TrimSpace(text)
				if err := types.ValidateIssueTitle(text); err != nil {
					return graphFailure("invalid_properties", err.Error(), 2)
				}
			}
			if name == "assignee" {
				if err := types.CheckFieldLen("assignee", text); err != nil {
					return graphFailure("invalid_properties", err.Error(), 2)
				}
			}
			if name == "notes" {
				if err := validateNotesUpdate(text); err != nil {
					return graphFailure("invalid_properties", err.Error(), 2)
				}
			}
			*dest = &text
			continue
		}
		switch name {
		case "issue_type":
			text, ok := value.(string)
			if !ok {
				return graphFailure("invalid_properties", "Issue issue_type must be a JSON string", 2)
			}
			classification := types.IssueType(text).Normalize()
			request.IssueType = &classification
		case "priority":
			if request.Priority != nil {
				return graphFailure("invalid_properties", "Issue priority was supplied by both --properties and --priority", 2)
			}
			integer, err := graphPreviewIssuePropertyInt(name, value)
			if err != nil {
				return err
			}
			priority, err := validation.ValidatePriority(strconv.Itoa(integer))
			if err != nil {
				return graphFailure("invalid_properties", err.Error(), 2)
			}
			request.Priority = &priority
		case "estimated_minutes":
			if request.EstimatedMinutes != nil {
				return graphFailure("invalid_properties", "Issue estimated_minutes was supplied by both --properties and --estimate", 2)
			}
			if value == nil {
				request.ClearEstimatedMinutes = true
				continue
			}
			integer, err := graphPreviewIssuePropertyInt(name, value)
			if err != nil {
				return err
			}
			if err := types.ValidateIssueEstimatedMinutes(&integer); err != nil {
				return graphFailure("invalid_properties", err.Error(), 2)
			}
			request.EstimatedMinutes = &integer
		case "due_at":
			if request.DueAt.Set {
				return graphFailure("invalid_properties", "Issue due_at was supplied by both --properties and --due", 2)
			}
			if value == nil {
				request.DueAt = publicops.Field[*time.Time]{Set: true}
				continue
			}
			text, ok := value.(string)
			if !ok {
				return graphFailure("invalid_properties", "Issue due_at must be an RFC 3339 JSON string or null", 2)
			}
			due, err := time.Parse(time.RFC3339Nano, text)
			if err != nil {
				return graphFailure("invalid_properties", "Issue due_at must be an RFC 3339 timestamp", 2)
			}
			request.DueAt = publicops.Field[*time.Time]{Set: true, Value: &due}
		default:
			return graphFailure("invalid_properties", "Issue property "+name+" is not writable through update", 2)
		}
	}
	force, _ := cmd.Flags().GetBool("force")
	if request.FromRevision == "" {
		request.ForceNotesOverwrite = force && request.Notes != nil
		request.ForceAssigneeTransfer = force && request.Assignee != nil && request.ExpectedAssignee == nil
		if force && !request.ForceNotesOverwrite && request.Assignee == nil && !request.ForceClosePolicy {
			return graphFailure("invalid_properties", "graph Issue update accepts --force only with notes replacement, assignee, or status", 2)
		}
	}
	return nil
}

func runGraphPreviewUpdateIssue(cmd *cobra.Command, path, identity string, revisionIdentity ...string) error {
	// Resolve the structured input and its source conflicts before a description
	// flag can consume stdin or open a file.
	values, err := graphPreviewIssuePropertyValues(cmd)
	if err != nil {
		return err
	}
	guardIdentity := identity
	if len(revisionIdentity) != 0 {
		guardIdentity = revisionIdentity[0]
	}
	request, err := graphPreviewIssueEditRequest(cmd, path, guardIdentity)
	if err != nil {
		return err
	}
	selected, err := graphPreviewIssueSelector(identity)
	if err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	if path == selected.Path {
		request.NativeID = selected.NativeID
	}
	if values != nil {
		if err := graphPreviewApplyIssueProperties(cmd, &request, values); err != nil {
			return err
		}
	}
	request.Actor = getActorWithGit()
	var touchedID string
	err = withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		result, err := store.UpdateIssue(ctx, request)
		if err != nil {
			return nil, "", err
		}
		touchedID = result.Issue.ID
		verb := "Unchanged"
		if result.Changed {
			verb = "Updated"
		}
		return result, fmt.Sprintf("%s %s", verb, result.Issue.ID), nil
	})
	if err == nil {
		SetLastTouchedID(touchedID)
	}
	return err
}

// A native bare classification is a product shorthand, not BDP relative URI
// resolution. Full nominal identities keep their exact declared meaning.
func graphPreviewNativeIssueTypeFlag(cmd *cobra.Command, scope string) (string, error) {
	if !cmd.Flags().Changed("type") {
		return "", nil
	}
	raw, _ := cmd.Flags().GetString("type")
	id, err := graphPreviewCollectionTypeURL(scope, raw, false)
	if err != nil {
		return "", graphFailure("invalid_selector", err.Error(), 2)
	}
	base, _, native := graphstore.NativeTypeClassification(scope, id, graph.KindBead)
	if !native || base != graphstore.IssueTypeURL(scope) {
		return "", graphFailure("invalid_selector", "native Issue requires a concrete native Issue Type", 2)
	}
	return id, nil
}

// Shipping bd update treats multiple IDs as independent writes. Keep each
// checked Issue edit in its own transaction so a missing sibling cannot undo
// successful edits, and do not reinterpret a single revision as a set guard.
func runGraphPreviewUpdateIssues(cmd *cobra.Command, args []string) error {
	if cmd.Flags().Changed("if-revision") {
		return graphFailure("invalid_selector", "multi-ID graph update cannot share one --if-revision; update each Issue separately", 2)
	}
	if cmd.Flags().Changed("patch") {
		return graphFailure("invalid_selector", "multi-ID graph update --patch requires one Bead ID", 2)
	}
	if cmd.Flags().Changed("parent") || cmd.Flags().Changed("type") || cmd.Flags().Changed("claim") ||
		cmd.Flags().Changed("from-revision") || cmd.Flags().Changed("replace-properties") ||
		graphPreviewLinkPinFlagsChanged(cmd) {
		return graphFailure("capability_unavailable", "multi-ID graph update admits only checked Issue field and property edits", 5)
	}
	selected := make([]graphstore.IssueSelector, len(args))
	for i, arg := range args {
		identity, err := graphPreviewWritableIdentity(arg)
		if err != nil {
			return graphFailure("invalid_selector", err.Error(), 2)
		}
		selected[i], err = graphPreviewIssueSelector(identity)
		if err != nil {
			return graphFailure("invalid_selector", err.Error(), 2)
		}
	}
	// Preflight every operand before edit flags can consume body stdin.
	request, err := graphPreviewIssueEditRequest(cmd, "")
	if err != nil {
		return err
	}
	if err := graphPreviewIssuePropertyMerge(cmd, &request); err != nil {
		return err
	}
	request.Actor = getActorWithGit()
	results := make([]graphstore.IssueMutationResult, 0, len(args))
	human := make([]string, 0, len(args))
	failed, unknown := false, false
	for i, selector := range selected {
		item := request
		item.Path, item.NativeID, item.Exact = selector.Path, selector.NativeID, selector.Exact
		err := withGraphStoreOutput(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
			result, err := store.UpdateIssue(ctx, item)
			if err != nil {
				return nil, "", fmt.Errorf("updating %s: %w", args[i], err)
			}
			verb := "Unchanged"
			if result.Changed {
				verb = "Updated"
			}
			return result, fmt.Sprintf("%s %s", verb, result.Issue.ID), nil
		}, func(value any, line string) error {
			results = append(results, value.(graphstore.IssueMutationResult))
			human = append(human, line)
			SetLastTouchedID(value.(graphstore.IssueMutationResult).Issue.ID)
			return nil
		})
		if err != nil {
			failed = true
			var reported *exitError
			if errors.As(err, &reported) && reported.Code == 6 {
				unknown = true
				break // An uncertain commit needs inspection before any further writes.
			}
		}
	}
	if len(results) > 0 {
		if err := graphPrint(results, strings.Join(human, "\n"), quietFlag); err != nil {
			return err
		}
	}
	if unknown {
		return &exitError{Code: 6}
	}
	if failed {
		return &exitError{Code: 1}
	}
	return nil
}
