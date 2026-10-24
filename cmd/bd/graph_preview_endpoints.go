package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// graphPreviewAuthoringEndpoints accepts the same three endpoint forms for
// Link creation and the Memory authoring facades. A repeated endpoint is a
// refusal even when both spellings carry the same value.
func graphPreviewAuthoringEndpoints(cmd *cobra.Command, args []string) (string, string, error) {
	if err := canonicalExactOccurrences(cmd, os.Args[1:], "from", "to"); err != nil {
		return "", "", err
	}
	fromFlag := cmd.Flags().Changed("from")
	toFlag := cmd.Flags().Changed("to")
	from, _ := cmd.Flags().GetString("from")
	to, _ := cmd.Flags().GetString("to")
	switch len(args) {
	case 0:
		if !fromFlag || !toFlag {
			return "", "", fmt.Errorf("requires SOURCE TARGET or --from SOURCE --to TARGET")
		}
	case 1:
		if fromFlag || !toFlag {
			return "", "", fmt.Errorf("requires SOURCE TARGET, SOURCE --to TARGET, or --from SOURCE --to TARGET")
		}
		from = args[0]
	case 2:
		if fromFlag || toFlag {
			return "", "", fmt.Errorf("endpoint supplied both positionally and by flag")
		}
		from, to = args[0], args[1]
	default:
		return "", "", fmt.Errorf("requires at most SOURCE TARGET")
	}
	if from == "" || to == "" {
		return "", "", fmt.Errorf("source and target require nonempty Bead identities")
	}
	return from, to, nil
}
