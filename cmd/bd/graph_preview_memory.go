package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func runGraphPreviewUpdate(cmd *cobra.Command, args []string) error {
	if err := graphPreviewWritePolicy(); err != nil {
		return err
	}
	if len(args) == 0 {
		if cmd.Flags().Changed("properties") || cmd.Flags().Changed("patch") {
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
	path, err := graphPreviewResourcePath(graphPreviewConfig.GraphScopeURL, args[0])
	if err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	if cmd.Flags().Changed("patch") {
		if strings.HasPrefix(path, "links/") {
			return runGraphPreviewLinkPropertiesPatch(cmd, path)
		}
		return runGraphPreviewMemoryPropertiesPatch(cmd, path)
	}
	if cmd.Flags().Changed("claim") {
		return runGraphPreviewClaimIssue(cmd, path)
	}
	if graphPreviewMetadataFlagsChanged(cmd) && !cmd.Flags().Changed("properties") && !graphPreviewIssueEditFlagsChanged(cmd) {
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

// Replacement is explicit: a missing string is not an instruction to clear it.
func graphPreviewMemoryProperties(properties map[string]any) (graphstore.Properties, error) {
	title, titleOK := properties["title"].(string)
	body, bodyOK := properties["body"].(string)
	if len(properties) != 2 || !titleOK || !bodyOK {
		return graphstore.Properties{}, fmt.Errorf("Memory replacement requires exactly the title and body string properties")
	}
	return graphstore.Properties{Title: title, Body: body}, nil
}

func runGraphPreviewUpdateMemory(cmd *cobra.Command, path string) error {
	if err := graphPreviewFlags(cmd, "properties", "metadata", "set-metadata", "unset-metadata", "if-revision", "unconditional"); err != nil {
		return err
	}
	if !cmd.Flags().Changed("properties") {
		return graphFailure("invalid_properties", "Memory update requires --properties to explicitly replace title and body", 2)
	}
	revision, unconditional, err := graphPreviewRevisionGuard(cmd, false, true)
	if err != nil {
		return err
	}
	input, _ := cmd.Flags().GetString("properties")
	values, err := graphPreviewProperties(input, cmd.InOrStdin())
	if err != nil {
		return graphFailure("invalid_properties", err.Error(), 2)
	}
	properties, err := graphPreviewMemoryProperties(values)
	if err != nil {
		return graphFailure("invalid_properties", err.Error(), 2)
	}
	metadata, _, err := graphPreviewMetadataPatch(cmd)
	if err != nil {
		return err
	}
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		result, err := store.UpdateMemory(ctx, graphstore.MemoryUpdateRequest{Path: path, Properties: properties, Metadata: metadata, Actor: getActorWithGit(), ExpectedRevision: revision, Unconditional: unconditional})
		verb := "Updated"
		if !result.Changed {
			verb = "Unchanged"
		}
		return result, graphPreviewReplacementSummary(fmt.Sprintf("%s %s", verb, result.Memory.ID), result.Replaced), err
	})
}
