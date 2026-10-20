package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// Historical acquisition/composition is a single storage Update transaction;
// this adapter never simulates it with a historical read and replacement.
func runGraphPreviewHistoricalUpdate(cmd *cobra.Command, path string) error {
	if err := graphPreviewFlags(cmd, "from-revision", "properties", "patch", "replace-properties", "metadata", "replace-metadata", "set-metadata", "unset-metadata", "if-revision", "force"); err != nil {
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
		// This dispatch reads only the current backing. Historical acquisition and
		// composition remain inside UpdateIssue's single storage transaction.
		current, readErr := store.Read(ctx, path)
		if readErr != nil && !errors.Is(readErr, graphstore.ErrGone) {
			return nil, "", readErr
		}
		if _, native := current.(graphstore.IssueRecord); native {
			if patch != nil {
				return nil, "", graphFailure("capability_unavailable", "historical native Issue ordered property changes are not yet supported", 5)
			}
			nativeRequest := graphstore.UpdateIssueRequest{Path: path, FromRevision: source, Actor: request.Actor, ExpectedRevision: revision, Unconditional: unconditional, Metadata: metadata}
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
