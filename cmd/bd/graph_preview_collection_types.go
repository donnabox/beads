package main

import (
	"fmt"
	"strings"

	graph "github.com/steveyegge/beads/graphops"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
)

func init() {
	countCmd.Flags().String("exact-type", "", "Select only the exact declared Bead Type (graph preview only)")
}

// Product collection selectors lower into BDP's unchanged nominal type and
// conformsTo predicates. Native bare labels are Beads adapter conveniences;
// generic BDP relative operands do not use this resolver.
func graphPreviewCollectionTypes(cmd *cobra.Command, scope string, links bool) (exact, family string, err error) {
	selected := ""
	for _, name := range []string{"type", "exact-type", "conforms-to"} {
		if cmd.Flags().Lookup(name) != nil && cmd.Flags().Changed(name) {
			if selected != "" {
				return "", "", fmt.Errorf("--type, --exact-type and --conforms-to are mutually exclusive")
			}
			selected = name
		}
	}
	if selected == "" {
		if !links {
			return "", graphstore.IssueTypeURL(scope), nil
		}
		return "", "", nil
	}
	raw, _ := cmd.Flags().GetString(selected)
	if raw == "" {
		return "", "", fmt.Errorf("--%s requires a Type", selected)
	}
	if selected == "conforms-to" {
		// Preserve the existing canonical-URI alias; product shorthand belongs
		// to --type and does not change raw BDP relative-operand semantics.
		if !strings.Contains(raw, "://") {
			return "", "", fmt.Errorf("--conforms-to requires a canonical absolute HTTP(S) Type URI")
		}
		if err := graph.ValidateTypeURL(raw); err != nil {
			return "", "", err
		}
		return "", raw, nil
	}
	if selected == "type" && raw == "any" {
		return "", "", nil
	}
	if raw == "any" {
		return "", "", fmt.Errorf("any is the --type wildcard, not a Type name")
	}
	id, err := graphPreviewCollectionTypeURL(scope, raw, links)
	if err != nil {
		return "", "", err
	}
	if selected == "exact-type" {
		return id, "", nil
	}
	return "", id, nil
}

func graphPreviewCollectionTypeURL(scope, raw string, links bool) (string, error) {
	if strings.Contains(raw, "://") {
		if err := graph.ValidateTypeURL(raw); err != nil {
			return "", err
		}
		return raw, nil
	}
	if graphPreviewNominalSelector(raw) {
		return graphPreviewTypeURL(scope, raw)
	}
	if !links {
		if raw == "issue" {
			return graphstore.IssueTypeURL(scope), nil
		}
		classification := types.IssueType(raw).Normalize()
		if classification.IsBuiltIn() {
			return graphstore.NativeIssueTypeURL(scope, string(classification))
		}
		for _, name := range graphstore.FreshTypeNames() {
			if raw == name {
				return graphPreviewTypeURL(scope, raw)
			}
		}
		// Configured-unused and formerly-used custom declarations are resolved
		// by the checked Type inventory in the actual read transaction.
		return graphstore.NativeIssueTypeURL(scope, string(classification))
	} else {
		if raw == "dependency" {
			return graphstore.DependencyTypeURL(scope), nil
		}
		// Known Memory Link names keep their own namespace. Other native-valid
		// labels may name arbitrary native Dependency kinds without global IDs.
		for _, name := range graphstore.FreshTypeNames() {
			if raw == name {
				return graphPreviewTypeURL(scope, raw)
			}
		}
		return graphstore.NativeDependencyTypeURL(scope, string(canonicalDependencyType(types.DependencyType(raw))))
	}
}
