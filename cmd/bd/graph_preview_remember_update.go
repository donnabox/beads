package main

import (
	"context"
	"fmt"
	"unicode/utf8"

	"github.com/spf13/cobra"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// This guarded composition is a disposable authoring convenience, not upsert.
// The read supplies omitted title bytes. UpdateMemory checks the caller's
// unchanged revision again in its own transaction; never refresh it or retry.
func runGraphPreviewRememberUpdate(cmd *cobra.Command, args []string) error {
	selector, _ := cmd.Flags().GetString("update")
	path, err := graphPreviewResourcePath(graphPreviewConfig.GraphScopeURL, selector)
	if err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	if err := graph.ValidateBeadPath(path); err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	revision, _ := cmd.Flags().GetString("if-revision")
	if !cmd.Flags().Changed("if-revision") || revision == "" || !utf8.ValidString(revision) || len(revision) > graphstore.PreviewVersionTokenLimit {
		return graphFailure("invalid_selector", fmt.Sprintf("remember --update requires --if-revision with a nonempty UTF-8 token of at most %d bytes", graphstore.PreviewVersionTokenLimit), 2)
	}
	var title *string
	if cmd.Flags().Changed("title") {
		value, _ := cmd.Flags().GetString("title")
		if !utf8.ValidString(value) {
			return graphFailure("invalid_properties", "Memory title must be valid UTF-8", 2)
		}
		title = &value
	}
	body, err := graphPreviewRememberBody(cmd, args)
	if err != nil {
		return err
	}
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		current, err := store.Read(ctx, path)
		if err != nil {
			return nil, "", err
		}
		properties, err := graphPreviewRememberUpdateProperties(current, body, title, revision)
		if err != nil {
			return nil, "", err
		}
		result, err := store.UpdateMemory(ctx, graphstore.MemoryUpdateRequest{
			Path: path, Properties: properties, Actor: getActorWithGit(), ExpectedRevision: revision,
		})
		if err != nil {
			return nil, "", err
		}
		verb := "Updated"
		if !result.Changed {
			verb = "Unchanged"
		}
		return result, fmt.Sprintf("%s %s: %q", verb, result.Memory.ID, result.Memory.Properties.Title), nil
	})
}

// Build the complete replacement only from an already checked Memory read.
// A nil title means preserve; an explicit empty string is a value. The guard
// below detects a stale composition read, but is not the authority's write CAS.
func graphPreviewRememberUpdateProperties(current any, body string, title *string, revision string) (graphstore.Properties, error) {
	memory, ok := current.(graphstore.Record)
	if !ok {
		return graphstore.Properties{}, fmt.Errorf("%w: remember --update requires the experimental Memory Type", graphstore.ErrCapabilityUnavailable)
	}
	if revision == "" {
		return graphstore.Properties{}, fmt.Errorf("%w: remember --update requires an observed revision", storage.ErrValidation)
	}
	if memory.Revision != revision {
		return graphstore.Properties{}, fmt.Errorf("%w: Memory revision changed (current %s)", graphstore.ErrConflict, memory.Revision)
	}
	properties := memory.Properties
	properties.Body = body
	if title != nil {
		properties.Title = *title
	}
	return properties, nil
}
