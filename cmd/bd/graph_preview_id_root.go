package main

import (
	"fmt"

	"github.com/spf13/cobra"
	graph "github.com/steveyegge/beads/graphops"
)

func graphPreviewGeneratedIDRoot(cmd *cobra.Command) (string, bool, error) {
	if cmd.Flags().Lookup("id-root") == nil || !cmd.Flags().Changed("id-root") {
		return "", false, nil
	}
	if cmd.Flags().Changed("id") {
		return "", false, fmt.Errorf("choose --id or --id-root, not both")
	}
	root, _ := cmd.Flags().GetString("id-root")
	if err := graph.ValidateCanonicalResourceID(root); err != nil {
		return "", false, err
	}
	return root, true, nil
}
