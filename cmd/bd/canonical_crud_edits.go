package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/steveyegge/beads/internal/canonicaledit"
)

func canonicalCRUDConsumedPropertyAlias(cmd *cobra.Command, flag *pflag.Flag) bool {
	if flag.Changed || !cmd.Flags().Changed("patch") {
		return false
	}
	switch flag.Name {
	case "patch-properties", "set-property", "unset-property":
		return true
	default:
		return false
	}
}

// Register only on graph mutation commands. Legacy workspace dispatch retains
// its existing contract and refuses these graph-only options.
func canonicalCRUDRegisterFlags(cmd *cobra.Command) {
	cmd.Flags().String("patch-properties", "", "Apply ordered strict JSON property operations from JSON, @file, or @-")
	cmd.Flags().StringArray("set-property", nil, "Set a literal top-level property KEY=VALUE (repeatable)")
	cmd.Flags().StringArray("unset-property", nil, "Remove a literal top-level property, including an absent key (repeatable)")
	cmd.Flags().String("patch-metadata", "", "Apply ordered strict JSON metadata operations from JSON, @file, or @-")
}

func canonicalCRUDFlagsChanged(cmd *cobra.Command) bool {
	for _, name := range []string{"patch-properties", "set-property", "unset-property", "patch-metadata"} {
		if cmd.Flags().Changed(name) {
			return true
		}
	}
	return false
}

// Normalize only after graph admission, before storage is opened. The legacy
// --patch spelling remains compatible, but cannot select a second edit mode.
func canonicalCRUDNormalizeUpdate(cmd *cobra.Command) error {
	stdin := 0
	if graphPreviewDescriptionUsesStdin(cmd) {
		stdin++
	}
	for _, name := range []string{"properties", "replace-properties", "patch-properties", "patch", "metadata", "replace-metadata", "patch-metadata"} {
		if cmd.Flags().Changed(name) {
			value, _ := cmd.Flags().GetString(name)
			if value == "@-" {
				stdin++
			}
		}
	}
	if stdin > 1 {
		return graphFailure("invalid_properties", "only one object input may consume stdin", 2)
	}
	modes := 0
	for _, name := range []string{"properties", "replace-properties", "patch-properties", "patch"} {
		if cmd.Flags().Changed(name) {
			modes++
		}
	}
	keyed := cmd.Flags().Changed("set-property") || cmd.Flags().Changed("unset-property")
	if keyed {
		modes++
	}
	if modes > 1 {
		return graphFailure("invalid_properties", "choose one properties edit mode: merge, replace, patch, or set/unset", 2)
	}
	if (keyed || cmd.Flags().Changed("patch-properties")) && graphPreviewIssueEditFlagsChanged(cmd) {
		return graphFailure("invalid_properties", "property operations cannot combine with Issue property flags", 2)
	}
	if !keyed && !cmd.Flags().Changed("patch-properties") {
		return nil
	}
	var raw []byte
	var err error
	if keyed {
		sets, _ := cmd.Flags().GetStringArray("set-property")
		unsets, _ := cmd.Flags().GetStringArray("unset-property")
		raw, err = canonicaledit.KeyedOperations(sets, unsets)
	} else {
		input, _ := cmd.Flags().GetString("patch-properties")
		raw, err = graphPreviewMemoryPropertiesPatchInput(input, cmd.InOrStdin())
	}
	if err != nil {
		return graphFailure("invalid_properties", err.Error(), 2)
	}
	if err := cmd.Flags().Set("patch", string(raw)); err != nil {
		return err
	}
	for _, name := range []string{"patch-properties", "set-property", "unset-property"} {
		if flag := cmd.Flags().Lookup(name); flag != nil {
			flag.Changed = false
		}
	}
	return nil
}
