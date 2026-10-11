package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/legacyimport"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func runGraphPreviewImport(cmd *cobra.Command, args []string) error {
	if err := graphPreviewWritePolicy(); err != nil {
		return err
	}
	if err := graphPreviewFlags(cmd, "input", "dry-run", "source-prefix"); err != nil {
		return err
	}
	if cmd.Flags().Changed("source-prefix") && importSourcePrefix == "" {
		return graphFailure("invalid_properties", "--source-prefix requires a nonempty prefix", 2)
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
		if errors.Is(err, legacyimport.ErrLimitExceeded) {
			return graphStorageError(fmt.Errorf("%w: %w", graphstore.ErrLimitExceeded, err))
		}
		return graphFailure("invalid_properties", err.Error(), 2)
	}
	return withGraphStoreBudget(5*time.Minute, func(ctx context.Context, s *graphstore.Store) (any, string, error) {
		result, err := s.ImportLegacyWithPrefix(ctx, batch, getActorWithGit(), importDryRun, importSourcePrefix)
		if err != nil {
			return nil, "", err
		}
		action := "Imported"
		if importDryRun {
			action = "Would import"
		}
		return result, graphPreviewLegacyImportSummary(action, source, result), nil
	}, func(result any, human string) error {
		return graphPrint(result, human, quietFlag)
	})
}

func graphPreviewLegacyImportSummary(action, source string, result graphstore.LegacyImportResult) string {
	prefix := ""
	if result.IssuePrefix != "" {
		if result.DryRun {
			prefix = fmt.Sprintf(" Would adopt Issue prefix: %s.", result.IssuePrefix)
		} else {
			prefix = fmt.Sprintf(" Workspace Issue prefix: %s.", result.IssuePrefix)
		}
	}
	return fmt.Sprintf("%s %d Issues, %d Memories, %d Dependencies and %d comments from %s.%s Source history is not present in legacy exports.", action, result.Issues, result.Memories, result.Dependencies, result.Comments, source, prefix)
}
