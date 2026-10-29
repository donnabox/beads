package main

import (
	"context"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func init() {
	initCmd.Flags().Bool("atomic-batch", false, "Fresh graph workspace with the experimental canonical atomic batch profile; never upgrades an existing store")
}

func canonicalGraphSchemaSupported(version int) bool {
	return version == graphstore.SchemaVersion || version == graphstore.AtomicBatchSchemaVersion
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
