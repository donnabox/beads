package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// Historical content acquisition and all guards remain inside the Link writer.
func runGraphPreviewHistoricalLinkUpdate(cmd *cobra.Command, path string) error {
	if err := graphPreviewFlags(cmd, "from-revision", "properties", "patch", "replace-properties", "metadata", "replace-metadata", "set-metadata", "unset-metadata", "if-revision", "if-source-revision", "source-revision", "target-revision"); err != nil {
		return err
	}
	selected, err := graphPreviewRevisionFlag(cmd, "from-revision", path, true)
	if err != nil {
		return err
	}
	revision, unconditional, err := graphPreviewEditRevisionGuard(cmd, path, true)
	if err != nil {
		return err
	}
	sourceRevision, unconditionalSource, err := graphPreviewRevisionGuard(cmd, true, path, true)
	if err != nil {
		return err
	}
	metadata, err := graphPreviewMetadataPatch(cmd)
	if err != nil {
		return err
	}
	sourcePin, targetPin, err := graphPreviewLinkPinIntents(cmd)
	if err != nil {
		return err
	}
	request := graphstore.LinkUpdateRequest{Path: path, FromRevision: selected, Actor: getActorWithGit(), ExpectedRevision: revision, Unconditional: unconditional, ExpectedSourceRevision: sourceRevision, UnconditionalSource: unconditionalSource, Metadata: metadata, SourcePin: sourcePin, TargetPin: targetPin}
	if cmd.Flags().Changed("properties") {
		if cmd.Flags().Changed("patch") || cmd.Flags().Changed("replace-properties") {
			return graphFailure("invalid_properties", "--properties cannot combine with --patch or --replace-properties", 2)
		}
		input, _ := cmd.Flags().GetString("properties")
		request.Properties, err = graphPreviewProperties(input, cmd.InOrStdin())
		if err != nil {
			return graphFailure("invalid_properties", err.Error(), 2)
		}
		request.MergeProperties = true
	}
	var patch []byte
	if cmd.Flags().Changed("patch") || cmd.Flags().Changed("replace-properties") {
		patch, err = graphPreviewPropertyOperations(cmd)
		if err != nil {
			return graphFailure("invalid_properties", err.Error(), 2)
		}
	}
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		var result graphstore.LinkMutationResult
		var err error
		if patch != nil {
			result, err = store.PatchLinkProperties(ctx, graphstore.LinkPropertiesPatchRequest{Path: path, FromRevision: selected, Patch: patch, Actor: request.Actor, ExpectedRevision: revision, Unconditional: unconditional, ExpectedSourceRevision: sourceRevision, UnconditionalSource: unconditionalSource, Metadata: metadata, SourcePin: sourcePin, TargetPin: targetPin})
		} else {
			result, err = store.UpdateLink(ctx, request)
		}
		if err != nil {
			return nil, "", err
		}
		verb := "Updated"
		if !result.Changed {
			verb = "Unchanged"
		}
		return graphPreviewReplacementResult(result, fmt.Sprintf("%s %s", verb, result.Link.ID), result.ReplacedSource, nil)
	})
}
