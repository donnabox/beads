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
	if len(args) != 1 {
		return graphFailure("invalid_selector", "graph update requires one canonical beads/PATH or links/PATH", 2)
	}
	path, err := graphPreviewResourcePath(graphPreviewConfig.GraphScopeURL, args[0])
	if err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	if strings.HasPrefix(path, "links/") {
		return runGraphPreviewUpdateLink(cmd, args)
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
	if err := graphPreviewFlags(cmd, "properties", "if-revision", "unconditional"); err != nil {
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
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		result, err := store.UpdateMemory(ctx, graphstore.MemoryUpdateRequest{Path: path, Properties: properties, Actor: getActorWithGit(), ExpectedRevision: revision, Unconditional: unconditional})
		verb := "Updated"
		if !result.Changed {
			verb = "Unchanged"
		}
		return result, fmt.Sprintf("%s %s", verb, result.Memory.ID), err
	})
}
