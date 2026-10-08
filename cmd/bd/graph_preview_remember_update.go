package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cobra"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/storage/issueops"
)

// An explicit ID follows ordinary remember's create-or-update convenience.
// Create reserves the ID atomically. Only an already allocated ID falls through
// to the existing Memory writer, which resolves omitted fields and the accepted
// predecessor in its own transaction. A deleted ID cannot be reused, and an
// Issue at that ID cannot be patched as a Memory.
func runGraphPreviewRememberUpsert(cmd *cobra.Command, args []string) error {
	path, err := graphPreviewCreateBeadPath(cmd)
	if err != nil {
		return err
	}
	revision, unconditional, err := graphPreviewRevisionGuard(cmd, false, false)
	if err != nil {
		return err
	}
	if !cmd.Flags().Changed("if-revision") && !cmd.Flags().Changed("unconditional") {
		unconditional = true
	}
	if !utf8.ValidString(revision) || len(revision) > graphstore.PreviewVersionTokenLimit {
		return graphFailure("invalid_selector", fmt.Sprintf("--if-revision requires a UTF-8 token of at most %d bytes", graphstore.PreviewVersionTokenLimit), 2)
	}
	title, body, err := graphPreviewRememberPatchInput(cmd, args)
	if err != nil {
		return err
	}
	metadata, _, err := graphPreviewMetadataPatch(cmd)
	if err != nil {
		return err
	}
	createMetadata, _, err := issueops.ApplyMetadataPatch(nil, metadata)
	if err != nil {
		return graphFailure("invalid_properties", err.Error(), 2)
	}
	actor := getActorWithGit()
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		if body != nil && !cmd.Flags().Changed("if-revision") && !(title != nil && strings.TrimSpace(*title) == "") {
			createTitle := graphPreviewMemoryTitle(*body)
			if title != nil {
				createTitle = *title
			}
			created, createErr := store.Create(ctx, graphstore.CreateRequest{Path: path, Title: createTitle, Body: *body, Actor: actor, Metadata: createMetadata})
			if createErr == nil {
				return created, fmt.Sprintf("Created %s\n", path), nil
			}
			if !errors.Is(createErr, graphstore.ErrAlreadyExists) {
				return nil, "", createErr
			}
			result, patchErr := store.PatchMemory(ctx, graphstore.MemoryPatchRequest{
				Path: path, Title: title, Body: body, Actor: actor,
				ExpectedRevision: revision, Unconditional: unconditional, Metadata: metadata,
			})
			if errors.Is(patchErr, graphstore.ErrNotFound) || errors.Is(patchErr, graphstore.ErrGone) || errors.Is(patchErr, graphstore.ErrCapabilityUnavailable) {
				return nil, "", createErr // Deleted or non-Memory identity remains reserved.
			}
			return graphPreviewRememberPatchResult(result, patchErr)
		}
		result, err := store.PatchMemory(ctx, graphstore.MemoryPatchRequest{
			Path: path, Title: title, Body: body, Actor: actor,
			ExpectedRevision: revision, Unconditional: unconditional, Metadata: metadata,
		})
		if errors.Is(err, graphstore.ErrNotFound) && title != nil && strings.TrimSpace(*title) == "" {
			return nil, "", graphFailure("invalid_properties", "an explicit creation --title must be nonempty", 2)
		}
		return graphPreviewRememberPatchResult(result, err)
	})
}

func graphPreviewRememberPatchResult(result graphstore.MemoryMutationResult, err error) (any, string, error) {
	if err != nil {
		return nil, "", err
	}
	verb := "Updated"
	if !result.Changed {
		verb = "Unchanged"
	}
	return result, graphPreviewReplacementSummary(fmt.Sprintf("%s %s: %q", verb, result.Memory.ID, result.Memory.Properties.Title), result.Replaced), nil
}

// Omitted fields are resolved by the existing writer inside its transaction,
// from the actual accepted predecessor. This command never reads and refreshes
// a guard, manufactures an unconditional replacement from stale content, or retries.
func runGraphPreviewRememberUpdate(cmd *cobra.Command, args []string) error {
	selector, _ := cmd.Flags().GetString("update")
	path, err := graphPreviewResourcePath(graphPreviewConfig.GraphScopeURL, selector)
	if err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	if err := graph.ValidateBeadPath(path); err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	revision, unconditional, err := graphPreviewRevisionGuard(cmd, false, false)
	if err != nil {
		return err
	}
	if !cmd.Flags().Changed("if-revision") && !cmd.Flags().Changed("unconditional") {
		unconditional = true
	}
	if !utf8.ValidString(revision) || len(revision) > graphstore.PreviewVersionTokenLimit {
		return graphFailure("invalid_selector", fmt.Sprintf("--if-revision requires a UTF-8 token of at most %d bytes", graphstore.PreviewVersionTokenLimit), 2)
	}
	title, body, err := graphPreviewRememberPatchInput(cmd, args)
	if err != nil {
		return err
	}
	metadata, _, err := graphPreviewMetadataPatch(cmd)
	if err != nil {
		return err
	}
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		result, err := store.PatchMemory(ctx, graphstore.MemoryPatchRequest{
			Path: path, Title: title, Body: body, Actor: getActorWithGit(),
			ExpectedRevision: revision, Unconditional: unconditional, Metadata: metadata,
		})
		if err != nil {
			return nil, "", err
		}
		verb := "Updated"
		if !result.Changed {
			verb = "Unchanged"
		}
		return result, graphPreviewReplacementSummary(fmt.Sprintf("%s %s: %q", verb, result.Memory.ID, result.Memory.Properties.Title), result.Replaced), nil
	})
}

// A selected edit must supply at least one field. Nil means preserve, while an
// explicitly empty string means clear. Without a body source, title-only edits
// do not inspect stdin, regardless of whether a pipe is attached.
func graphPreviewRememberPatchInput(cmd *cobra.Command, args []string) (*string, *string, error) {
	var title *string
	if cmd.Flags().Changed("title") {
		value, _ := cmd.Flags().GetString("title")
		if !utf8.ValidString(value) {
			return nil, nil, graphFailure("invalid_properties", "Memory title must be valid UTF-8", 2)
		}
		title = &value
	}
	if len(args) == 0 && !cmd.Flags().Changed("body-file") && !cmd.Flags().Changed("stdin") {
		if title == nil && !graphPreviewMetadataFlagsChanged(cmd) {
			return nil, nil, graphFailure("invalid_properties", "an existing-ID remember write requires --title, a metadata edit or one explicit body source", 2)
		}
		return title, nil, nil
	}
	body, err := graphPreviewRememberBody(cmd, args)
	if err != nil {
		return nil, nil, err
	}
	return title, &body, nil
}
