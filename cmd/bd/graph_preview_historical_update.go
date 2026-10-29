package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// Historical acquisition/composition is a single storage Update transaction;
// this adapter never simulates it with a historical read and replacement.
func runGraphPreviewHistoricalUpdate(cmd *cobra.Command, path string) error {
	if strings.HasPrefix(path, "links/") {
		return runGraphPreviewHistoricalLinkUpdate(cmd, path)
	}
	if err := graphPreviewFlags(cmd, "type", "from-revision", "properties", "patch", "replace-properties", "metadata", "replace-metadata", "patch-metadata", "set-metadata", "unset-metadata", "if-revision", "force"); err != nil {
		return err
	}
	if !strings.HasPrefix(path, "beads/") {
		return graphFailure("capability_unavailable", "historical Update requires a Bead or Link identity", 5)
	}
	source, err := graphPreviewRevisionFlag(cmd, "from-revision", path, true)
	if err != nil {
		return err
	}
	revision, unconditional, err := graphPreviewEditRevisionGuard(cmd, path, true)
	if err != nil {
		return err
	}
	metadata, err := graphPreviewMetadataPatch(cmd)
	if err != nil {
		return err
	}
	typeURL, err := graphPreviewMemoryTypeFlag(cmd)
	if err != nil {
		return err
	}
	request := graphstore.MemoryPatchRequest{TypeURL: typeURL, Path: path, FromRevision: source, Actor: getActorWithGit(), ExpectedRevision: revision, Unconditional: unconditional, Metadata: metadata}
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
		// This dispatch reads only the current backing. Historical acquisition and
		// composition remain inside UpdateIssue's single storage transaction.
		native, err := store.NativeIssueAllocation(ctx, path)
		if err != nil {
			return nil, "", err
		}
		if native {
			nativeType, err := graphPreviewNativeIssueTypeFlag(cmd, store.ScopeURL())
			if err != nil {
				return nil, "", err
			}
			nativeRequest := graphstore.UpdateIssueRequest{TypeURL: nativeType, Path: path, FromRevision: source, Actor: request.Actor, ExpectedRevision: revision, Unconditional: unconditional, Metadata: metadata, PropertiesPatch: patch}
			nativeRequest.ForceNotesOverwrite, _ = cmd.Flags().GetBool("force")
			if request.PropertiesProvided {
				var values map[string]any
				if err := json.Unmarshal(request.PropertiesMerge, &values); err != nil {
					return nil, "", err
				}
				if err := graphPreviewApplyIssueProperties(cmd, &nativeRequest, values); err != nil {
					return nil, "", err
				}
			}
			result, err := store.UpdateIssue(ctx, nativeRequest)
			if err != nil {
				return nil, "", err
			}
			verb := "Updated"
			if !result.Changed {
				verb = "Unchanged"
			}
			return result, fmt.Sprintf("%s %s", verb, result.Issue.ID), nil
		}
		if cmd.Flags().Changed("force") {
			return nil, "", graphFailure("invalid_properties", "--force applies only to a native Issue notes edit", 2)
		}
		var result graphstore.MemoryMutationResult
		if patch != nil {
			result, err = store.PatchMemoryProperties(ctx, graphstore.MemoryPropertiesPatchRequest{TypeURL: typeURL, Path: path, FromRevision: source, Patch: patch, Actor: request.Actor, ExpectedRevision: revision, Unconditional: unconditional, Metadata: metadata})
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
