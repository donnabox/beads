package main

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// Historical acquisition/composition is a single storage Update transaction;
// this adapter never simulates it with a historical read and replacement.
func runGraphPreviewHistoricalUpdate(cmd *cobra.Command, path string) error {
	if err := graphPreviewFlags(cmd, "from-revision", "properties", "patch", "replace-properties", "metadata", "replace-metadata", "set-metadata", "unset-metadata", "if-revision"); err != nil {
		return err
	}
	if !strings.HasPrefix(path, "beads/") {
		return graphFailure("capability_unavailable", "historical Link Update/recovery is not supported", 5)
	}
	source, _ := cmd.Flags().GetString("from-revision")
	if source == "" || !utf8.ValidString(source) || len(source) > graphstore.PreviewVersionTokenLimit {
		return graphFailure("invalid_selector", fmt.Sprintf("--from-revision requires a nonempty UTF-8 token of at most %d bytes", graphstore.PreviewVersionTokenLimit), 2)
	}
	revision, unconditional, err := graphPreviewEditRevisionGuard(cmd)
	if err != nil {
		return err
	}
	metadata, err := graphPreviewMetadataPatch(cmd)
	if err != nil {
		return err
	}
	request := graphstore.MemoryPatchRequest{Path: path, FromRevision: source, Actor: getActorWithGit(), ExpectedRevision: revision, Unconditional: unconditional, Metadata: metadata}
	if cmd.Flags().Changed("properties") {
		if cmd.Flags().Changed("patch") || cmd.Flags().Changed("replace-properties") {
			return graphFailure("invalid_properties", "--properties cannot combine with --patch or --replace-properties", 2)
		}
		input, _ := cmd.Flags().GetString("properties")
		values, err := graphPreviewPropertiesJSON(input, cmd.InOrStdin())
		if err != nil {
			return graphFailure("invalid_properties", err.Error(), 2)
		}
		request.PropertiesMerge = values
		request.PropertiesProvided = true
	}
	var patch []byte
	if cmd.Flags().Changed("patch") || cmd.Flags().Changed("replace-properties") {
		patch, err = graphPreviewPropertyOperations(cmd)
		if err != nil {
			return graphFailure("invalid_properties", err.Error(), 2)
		}
	}
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		var result graphstore.MemoryMutationResult
		var err error
		if patch != nil {
			result, err = store.PatchMemoryProperties(ctx, graphstore.MemoryPropertiesPatchRequest{Path: path, FromRevision: source, Patch: patch, Actor: request.Actor, ExpectedRevision: revision, Unconditional: unconditional, Metadata: metadata})
		} else {
			result, err = store.PatchMemory(ctx, request)
		}
		verb := "Updated"
		if !result.Changed {
			verb = "Unchanged"
		}
		return graphPreviewReplacementResult(result, fmt.Sprintf("%s %s", verb, result.Memory.ID), result.Replaced, err)
	})
}
