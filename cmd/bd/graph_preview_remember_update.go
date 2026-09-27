package main

import (
	"context"
	"fmt"
	"unicode/utf8"

	"github.com/spf13/cobra"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

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
	revision, unconditional, err := graphPreviewRevisionGuard(cmd, false, true)
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
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		result, err := store.PatchMemory(ctx, graphstore.MemoryPatchRequest{
			Path: path, Title: title, Body: body, Actor: getActorWithGit(),
			ExpectedRevision: revision, Unconditional: unconditional,
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
		if title == nil {
			return nil, nil, graphFailure("invalid_properties", "remember --update requires --title or one explicit body source", 2)
		}
		return title, nil, nil
	}
	body, err := graphPreviewRememberBody(cmd, args)
	if err != nil {
		return nil, nil, err
	}
	return title, &body, nil
}
