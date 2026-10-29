package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// An explicit fresh profile selects one supported layout. It never changes an
// existing workspace's format or the default mixed graph initializer.
func graphPreviewInitSchemaVersion(cmd *cobra.Command, mode string) (int, error) {
	if cmd.Flags().Lookup("graph-profile") == nil || !cmd.Flags().Changed("graph-profile") {
		return graphstore.SchemaVersion, nil
	}
	profile, err := cmd.Flags().GetString("graph-profile")
	if err != nil {
		return 0, err
	}
	if mode != "link" {
		return 0, fmt.Errorf("--graph-profile requires --graph-mode link")
	}
	if profile != "types" {
		return 0, fmt.Errorf("unsupported graph profile %q; expected types", profile)
	}
	return graphstore.TypesFreshSchemaVersion, nil
}
