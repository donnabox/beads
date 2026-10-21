package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/internal/utils"
)

func runGraphPreviewUpdate(cmd *cobra.Command, args []string) error {
	if err := graphPreviewWritePolicy(); err != nil {
		return err
	}
	if cmd.Flags().Changed("replace-properties") && (cmd.Flags().Changed("properties") || cmd.Flags().Changed("patch") || graphPreviewIssueEditFlagsChanged(cmd)) {
		return graphFailure("invalid_properties", "--replace-properties cannot combine with --properties, --patch or Issue property flags", 2)
	}
	if len(args) == 0 && cmd.Flags().Changed("from-revision") {
		return graphFailure("invalid_selector", "historical Update requires an explicit Bead ID", 2)
	}
	if len(args) == 0 {
		if graphPreviewLinkPinFlagsChanged(cmd) || cmd.Flags().Changed("properties") || cmd.Flags().Changed("replace-properties") || cmd.Flags().Changed("patch") || cmd.Flags().Changed("type") {
			return graphFailure("invalid_selector", "Memory and Link property updates require an explicit Resource ID", 2)
		}
		lastTouched := GetLastTouchedID()
		if lastTouched == "" {
			return graphFailure("invalid_selector", "no Issue ID provided and no last touched Issue", 2)
		}
		args = []string{lastTouched}
	}
	if len(args) != 1 {
		return graphFailure("invalid_selector", "graph update requires one Bead ID (or beads/PATH) or explicit links/PATH", 2)
	}
	selector, err := graphPreviewWritableIdentity(args[0])
	if err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	path, err := graphPreviewResourcePath(graphPreviewConfig.GraphScopeURL, selector)
	if err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	if graphPreviewLinkPinFlagsChanged(cmd) && !strings.HasPrefix(path, "links/") {
		return graphFailure("invalid_selector", "endpoint revision flags require an informational Link", 2)
	}
	if cmd.Flags().Changed("from-revision") {
		return runGraphPreviewHistoricalUpdate(cmd, path)
	}
	if cmd.Flags().Changed("patch") || cmd.Flags().Changed("replace-properties") {
		if strings.HasPrefix(path, "links/") {
			return runGraphPreviewLinkPropertiesPatch(cmd, path)
		}
		return runGraphPreviewMemoryPropertiesPatch(cmd, path)
	}
	if cmd.Flags().Changed("type") {
		if strings.HasPrefix(path, "links/") {
			return graphFailure("capability_unavailable", "Link retyping is not supported", 5)
		}
		if graphPreviewIssueEditFlagsChanged(cmd) {
			return runGraphPreviewUpdateIssue(cmd, path)
		}
		return runGraphPreviewUpdateMemory(cmd, path)
	}
	if cmd.Flags().Changed("claim") {
		return runGraphPreviewClaimIssue(cmd, path)
	}
	if graphPreviewMetadataFlagsChanged(cmd) && !graphPreviewLinkPinFlagsChanged(cmd) && !cmd.Flags().Changed("properties") && !graphPreviewIssueEditFlagsChanged(cmd) {
		return runGraphPreviewUpdateMetadataOnly(cmd, path)
	}
	if strings.HasPrefix(path, "links/") {
		return runGraphPreviewUpdateLink(cmd, args)
	}
	if graphPreviewIssueEditFlagsChanged(cmd) {
		return runGraphPreviewUpdateIssue(cmd, path)
	}
	return runGraphPreviewUpdateMemory(cmd, path)
}

// A top-level properties object edits only named Memory members. Missing
// members preserve their values in the checked storage transaction.
func graphPreviewMemoryProperties(properties map[string]any) (*string, *string, error) {
	var title, body *string
	for name, value := range properties {
		text, ok := value.(string)
		if !ok || (name != "title" && name != "body") {
			return nil, nil, fmt.Errorf("Memory properties admit only title and body JSON strings")
		}
		if name == "title" {
			title = &text
		} else {
			body = &text
		}
	}
	return title, body, nil
}

func runGraphPreviewUpdateMemory(cmd *cobra.Command, path string) error {
	if err := graphPreviewFlags(cmd, "type", "properties", "metadata", "replace-metadata", "set-metadata", "unset-metadata", "if-revision", "force"); err != nil {
		return err
	}
	if !cmd.Flags().Changed("properties") && !cmd.Flags().Changed("type") {
		return graphFailure("invalid_properties", "Memory update requires --properties or --type", 2)
	}
	typeURL, err := graphPreviewMemoryTypeFlag(cmd)
	if err != nil {
		return err
	}
	revision, unconditional, err := graphPreviewEditRevisionGuard(cmd)
	if err != nil {
		return err
	}
	var rawValues json.RawMessage
	if cmd.Flags().Changed("properties") {
		input, _ := cmd.Flags().GetString("properties")
		rawValues, err = graphPreviewPropertiesJSON(input, cmd.InOrStdin())
		if err != nil {
			return graphFailure("invalid_properties", err.Error(), 2)
		}
	}
	metadata, err := graphPreviewMetadataPatch(cmd)
	if err != nil {
		return err
	}
	issueUpdated := false
	err = withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		current, err := store.Read(ctx, path)
		if err != nil {
			return nil, "", err
		}
		if _, ok := current.(graphstore.IssueRecord); ok {
			values := make(map[string]any)
			if len(rawValues) != 0 {
				if err := json.Unmarshal(rawValues, &values); err != nil {
					return nil, "", err
				}
			}
			request := graphstore.UpdateIssueRequest{Path: path, Actor: getActorWithGit(), ExpectedRevision: revision, Unconditional: unconditional, Metadata: metadata}
			if typeURL != "" {
				selector, _ := cmd.Flags().GetString("type")
				if graphPreviewNominalSelector(selector) {
					return nil, "", graphFailure("capability_unavailable", "native Issue Type changes require their domain writer", 5)
				}
				classification := types.IssueType(utils.NormalizeIssueType(selector))
				request.IssueType = &classification
			}
			request.ForceNotesOverwrite, _ = cmd.Flags().GetBool("force")
			if err := graphPreviewApplyIssueProperties(cmd, &request, values); err != nil {
				return nil, "", err
			}
			result, err := store.UpdateIssue(ctx, request)
			if err != nil {
				return nil, "", err
			}
			issueUpdated = true
			verb := "Updated"
			if !result.Changed {
				verb = "Unchanged"
			}
			return result, graphPreviewMutationSummary(fmt.Sprintf("%s %s", verb, result.Issue.ID), result), nil
		}
		if cmd.Flags().Changed("force") {
			return nil, "", graphFailure("invalid_properties", "--force applies only to an Issue notes edit", 2)
		}
		if typeURL != "" {
			if err := graphPreviewRequireInstalledType(ctx, store, typeURL, graph.KindBead, true); err != nil {
				return nil, "", err
			}
		}
		result, err := store.PatchMemory(ctx, graphstore.MemoryPatchRequest{Path: path, TypeURL: typeURL, PropertiesMerge: rawValues, PropertiesProvided: cmd.Flags().Changed("properties"), Metadata: metadata, Actor: getActorWithGit(), ExpectedRevision: revision, Unconditional: unconditional})
		verb := "Updated"
		if !result.Changed {
			verb = "Unchanged"
		}
		return graphPreviewReplacementResult(result, fmt.Sprintf("%s %s", verb, result.Memory.ID), result.Replaced, err)
	})
	if err == nil && issueUpdated {
		SetLastTouchedID(path)
	}
	return err
}
