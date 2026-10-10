package main

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/legacyimport"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func runGraphPreviewImport(cmd *cobra.Command, args []string) error {
	if err := graphPreviewWritePolicy(); err != nil {
		return err
	}
	if err := graphPreviewFlags(cmd, "input", "dry-run"); err != nil {
		return err
	}
	if len(args) > 1 {
		return graphFailure("invalid_selector", "import accepts one file or '-'", 2)
	}
	if err := runImportInner(args); err != nil {
		if _, ok := err.(*exitError); ok {
			return err
		}
		return graphFailure("invalid_properties", err.Error(), 2)
	}
	return nil
}

func runGraphPreviewImportReader(reader io.Reader, source string) error {
	batch, err := legacyimport.Parse(reader)
	if err != nil {
		return graphFailure("invalid_properties", err.Error(), 2)
	}
	return withGraphStore(func(ctx context.Context, s *graphstore.Store) (any, string, error) {
		result, err := s.ImportLegacy(ctx, batch, getActorWithGit(), importDryRun)
		if err != nil {
			return nil, "", err
		}
		action := "Imported"
		if importDryRun {
			action = "Would import"
		}
		return result, fmt.Sprintf("%s %d Issues, %d Memories, %d Dependencies and %d comments from %s. Source history is not present in legacy exports.", action, result.Issues, result.Memories, result.Dependencies, result.Comments, source), nil
	})
}
