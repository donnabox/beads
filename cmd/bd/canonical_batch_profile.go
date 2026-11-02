package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func init() {
	initCmd.Flags().Bool("atomic-batch", false, "Fresh graph workspace with the experimental canonical atomic batch profile; never upgrades an existing store")
}

func canonicalGraphSchemaSupported(version int) bool {
	return version == graphstore.SchemaVersion || version == graphstore.TypesFreshSchemaVersion || version == graphstore.AtomicBatchSchemaVersion
}

// Select one explicit fresh profile without combining or upgrading layouts.
func selectedGraphPreviewInitSchemaVersion(cmd *cobra.Command, mode string) (int, error) {
	version, err := graphPreviewInitSchemaVersion(cmd, mode)
	if err != nil {
		return 0, err
	}
	if cmd.Flags().Lookup("atomic-batch") == nil {
		return version, nil
	}
	enabled, err := cmd.Flags().GetBool("atomic-batch")
	if err != nil {
		return 0, err
	}
	if !enabled {
		return version, nil
	}
	if mode != "link" {
		return 0, fmt.Errorf("--atomic-batch requires --graph-mode link")
	}
	if version != graphstore.SchemaVersion {
		return 0, fmt.Errorf("--atomic-batch and --graph-profile select separate fresh layouts")
	}
	return graphstore.AtomicBatchSchemaVersion, nil
}

// Select only the explicitly persisted local profile. This is not a fallback
// open, migration, or an implication that another optional profile is included.
func openSelectedGraphStore(ctx context.Context, options graphstore.Options) (*graphstore.Store, error) {
	if options.Binding.SchemaVersion == graphstore.AtomicBatchSchemaVersion {
		return graphstore.OpenAtomicBatch(ctx, options)
	}
	return graphstore.OpenExisting(ctx, options)
}

func initSelectedGraphStore(ctx context.Context, options graphstore.Options) error {
	if options.Binding.SchemaVersion == graphstore.AtomicBatchSchemaVersion {
		return graphstore.InitAtomicBatch(ctx, options)
	}
	return graphstore.Init(ctx, options)
}
