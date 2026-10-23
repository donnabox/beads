package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/storage/issueops"
)

// An explicit ID follows ordinary remember's create-or-update convenience.
// Create reserves the ID atomically. Only an already allocated ID falls through
// to the existing Memory writer, which resolves omitted fields and the accepted
// predecessor in its own transaction. A deleted ID cannot be reused, and an
// Issue at that ID cannot be patched as a Memory.
func runGraphPreviewRememberUpsert(cmd *cobra.Command, publicID string, title, body *string) error {
	path := ""
	var err error
	revision, unconditional, err := graphPreviewEditRevisionGuard(cmd)
	if err != nil {
		return err
	}
	if !utf8.ValidString(revision) || len(revision) > graphstore.PreviewVersionTokenLimit {
		return graphFailure("invalid_selector", fmt.Sprintf("--if-revision requires a UTF-8 token of at most %d bytes", graphstore.PreviewVersionTokenLimit), 2)
	}
	metadata, err := graphPreviewMetadataPatch(cmd)
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
			created, createErr := store.Create(ctx, graphstore.CreateRequest{PublicID: publicID, Title: createTitle, Body: *body, Actor: actor, Metadata: createMetadata})
			if createErr == nil {
				return created, fmt.Sprintf("Created %s\n", publicID), nil
			}
			if !graphPreviewRememberDefiniteError(createErr, graphstore.ErrAlreadyExists) {
				return nil, "", createErr
			}
			var resolveErr error
			path, resolveErr = store.ResolveExactAllocatedResourcePath(ctx, publicID)
			if resolveErr != nil {
				return nil, "", resolveErr
			}
			result, patchErr := store.PatchMemory(ctx, graphstore.MemoryPatchRequest{
				Path: path, Title: title, Body: body, PropertiesProvided: cmd.Flags().Changed("properties"), Actor: actor,
				ExpectedRevision: revision, Unconditional: unconditional, Metadata: metadata,
			})
			if graphPreviewRememberDefiniteError(patchErr, graphstore.ErrNotFound) || graphPreviewRememberDefiniteError(patchErr, graphstore.ErrGone) || graphPreviewRememberDefiniteError(patchErr, graphstore.ErrCapabilityUnavailable) {
				return nil, "", createErr // Deleted or non-Memory identity remains reserved.
			}
			return graphPreviewRememberPatchResult(result, patchErr)
		}
		path, err = store.ResolveExactAllocatedResourcePath(ctx, publicID)
		if err != nil {
			return nil, "", err
		}
		result, err := store.PatchMemory(ctx, graphstore.MemoryPatchRequest{
			Path: path, Title: title, Body: body, PropertiesProvided: cmd.Flags().Changed("properties"), Actor: actor,
			ExpectedRevision: revision, Unconditional: unconditional, Metadata: metadata,
		})
		if graphPreviewRememberDefiniteError(err, graphstore.ErrNotFound) && title != nil && strings.TrimSpace(*title) == "" {
			return nil, "", graphFailure("invalid_properties", "an explicit creation --title must be nonempty", 2)
		}
		return graphPreviewRememberPatchResult(result, err)
	})
}

// A failed rollback or lost commit acknowledgment can wrap a more specific
// storage error. Do not turn that ambiguous write into a definite fallback or
// validation error: the caller must inspect state before deciding to retry.
func graphPreviewRememberDefiniteError(err, kind error) bool {
	return errors.Is(err, kind) && !errors.Is(err, graphstore.ErrOutcomeUnknown)
}

func graphPreviewRememberPatchResult(result graphstore.MemoryMutationResult, err error) (any, string, error) {
	if err != nil {
		return nil, "", err
	}
	verb := "Updated"
	if !result.Changed {
		verb = "Unchanged"
	}
	return graphPreviewReplacementResult(result, fmt.Sprintf("%s %s: %q", verb, graphPreviewMemoryDisplayID(result.Memory), result.Memory.Properties.Title), result.Replaced, nil)
}

// Replacement needs an existing Memory. Creation initializes the complete
// document with --properties; a missing explicit ID is never silently created
// by this edit route. The checked Memory writer applies properties and metadata
// together, retaining its revision/no-op rules.
func runGraphPreviewRememberPropertiesReplacement(cmd *cobra.Command, args []string) error {
	if len(args) != 0 || cmd.Flags().Changed("title") || cmd.Flags().Changed("body-file") || cmd.Flags().Changed("stdin") || cmd.Flags().Changed("properties") || cmd.Flags().Changed("create-only") {
		return graphFailure("invalid_properties", "--replace-properties requires an existing --id or --update and cannot combine with body, title or --properties", 2)
	}
	literal := cmd.Flags().Changed("id") || cmd.Flags().Changed("key")
	if literal == cmd.Flags().Changed("update") {
		return graphFailure("invalid_selector", "replacement requires exactly one identity selector", 2)
	}
	selector := ""
	if cmd.Flags().Changed("id") {
		selector, _ = cmd.Flags().GetString("id")
	} else if cmd.Flags().Changed("key") {
		selector, _ = cmd.Flags().GetString("key")
	} else {
		selector, _ = cmd.Flags().GetString("update")
	}
	var path string
	var err error
	revision, unconditional, err := graphPreviewEditRevisionGuard(cmd)
	if err != nil {
		return err
	}
	if !utf8.ValidString(revision) || len(revision) > graphstore.PreviewVersionTokenLimit {
		return graphFailure("invalid_selector", fmt.Sprintf("--if-revision requires a UTF-8 token of at most %d bytes", graphstore.PreviewVersionTokenLimit), 2)
	}
	raw, err := graphPreviewPropertyOperations(cmd)
	if err != nil {
		return graphFailure("invalid_properties", err.Error(), 2)
	}
	metadata, err := graphPreviewMetadataPatch(cmd)
	if err != nil {
		return err
	}
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		if literal {
			path, err = store.ResolveExactAllocatedResourcePath(ctx, selector)
		} else {
			path, err = store.ResolveMemorySelector(ctx, selector)
		}
		if err != nil {
			return nil, "", err
		}

		result, err := store.PatchMemoryProperties(ctx, graphstore.MemoryPropertiesPatchRequest{
			Path: path, Patch: raw, Actor: getActorWithGit(), ExpectedRevision: revision,
			Unconditional: unconditional, Metadata: metadata,
		})
		return graphPreviewRememberPatchResult(result, err)
	})
}

// Omitted fields are resolved by the existing writer inside its transaction,
// from the actual accepted predecessor. This command never reads and refreshes
// a guard, manufactures an unconditional replacement from stale content, or retries.
func runGraphPreviewRememberUpdate(cmd *cobra.Command, args []string) error {
	selector, _ := cmd.Flags().GetString("update")
	var path string
	var err error
	revision, unconditional, err := graphPreviewEditRevisionGuard(cmd)
	if err != nil {
		return err
	}
	if !utf8.ValidString(revision) || len(revision) > graphstore.PreviewVersionTokenLimit {
		return graphFailure("invalid_selector", fmt.Sprintf("--if-revision requires a UTF-8 token of at most %d bytes", graphstore.PreviewVersionTokenLimit), 2)
	}
	title, body, err := graphPreviewRememberPatchInput(cmd, args)
	if err != nil {
		return err
	}
	metadata, err := graphPreviewMetadataPatch(cmd)
	if err != nil {
		return err
	}
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		path, err = store.ResolveMemorySelector(ctx, selector)
		if err != nil {
			return nil, "", err
		}

		result, err := store.PatchMemory(ctx, graphstore.MemoryPatchRequest{
			Path: path, Title: title, Body: body, PropertiesProvided: cmd.Flags().Changed("properties"), Actor: getActorWithGit(),
			ExpectedRevision: revision, Unconditional: unconditional, Metadata: metadata,
		})
		if err != nil {
			return nil, "", err
		}
		verb := "Updated"
		if !result.Changed {
			verb = "Unchanged"
		}
		return graphPreviewReplacementResult(result, fmt.Sprintf("%s %s: %q", verb, graphPreviewMemoryDisplayID(result.Memory), result.Memory.Properties.Title), result.Replaced, nil)
	})
}

// A selected edit must supply at least one field. Nil means preserve, while an
// explicitly empty string means clear. Without a body source, title-only edits
// do not inspect stdin, regardless of whether a pipe is attached.
func graphPreviewRememberPatchInput(cmd *cobra.Command, args []string) (*string, *string, error) {
	var propertyTitle, propertyBody *string
	if cmd.Flags().Changed("properties") {
		input, _ := cmd.Flags().GetString("properties")
		values, err := graphPreviewProperties(input, cmd.InOrStdin())
		if err != nil {
			return nil, nil, graphFailure("invalid_properties", err.Error(), 2)
		}
		propertyTitle, propertyBody, err = graphPreviewMemoryProperties(values)
		if err != nil {
			return nil, nil, graphFailure("invalid_properties", err.Error(), 2)
		}
	}
	var title *string
	if cmd.Flags().Changed("title") {
		if propertyTitle != nil {
			return nil, nil, graphFailure("invalid_properties", "Memory title was supplied by both --properties and --title", 2)
		}
		value, _ := cmd.Flags().GetString("title")
		if !utf8.ValidString(value) {
			return nil, nil, graphFailure("invalid_properties", "Memory title must be valid UTF-8", 2)
		}
		title = &value
	} else {
		title = propertyTitle
	}
	if len(args) == 0 && !cmd.Flags().Changed("body-file") && !cmd.Flags().Changed("stdin") {
		if title == nil && propertyBody == nil && !cmd.Flags().Changed("properties") && !graphPreviewMetadataFlagsChanged(cmd) {
			return nil, nil, graphFailure("invalid_properties", "an existing-ID remember write requires --properties, --title, a metadata edit or one explicit body source", 2)
		}
		return title, propertyBody, nil
	}
	if propertyBody != nil {
		return nil, nil, graphFailure("invalid_properties", "Memory body was supplied by both --properties and a shorthand body source", 2)
	}
	body, err := graphPreviewRememberBody(cmd, args)
	if err != nil {
		return nil, nil, err
	}
	return title, &body, nil
}
