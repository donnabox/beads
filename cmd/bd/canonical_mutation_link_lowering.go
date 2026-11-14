package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/canonicalbatch"
)

// Lower only explicitly supplied complete endpoint References. Omission keeps
// the old Reference; a bare replacement deliberately has no revision pin. The
// existing member compiler and publisher retain all kind/Scope/pin/Type checks.
// This function performs no input, definition or Resource acquisition.
func canonicalMutationLinkEndpointEdits(cmd *cobra.Command, scope string) (map[string]canonicalbatch.Operand, error) {
	result := map[string]canonicalbatch.Operand{}
	for _, endpoint := range []struct { flag, member string }{{"from", "source"}, {"to", "target"}} {
		if !cmd.Flags().Changed(endpoint.flag) {
			continue
		}
		value, err := cmd.Flags().GetString(endpoint.flag)
		if err != nil {
			return nil, err
		}
		ref, err := canonicalbatch.ParseOperand(value, scope)
		if err != nil {
			return nil, fmt.Errorf("--%s: %w", endpoint.flag, err)
		}
		result[endpoint.member] = ref
	}
	return result, nil
}
