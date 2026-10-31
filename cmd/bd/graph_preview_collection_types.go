package main

import (
	"context"
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
	return graphPreviewCollectionTypesWithDefault(cmd, scope, links, true)
}

// issueDefault is a caller-selected population policy, independent of Resource
// kind and explicit selectors. Canonical List selects all ordinary Beads.
func graphPreviewCollectionTypesWithDefault(cmd *cobra.Command, scope string, links, issueDefault bool) (exact, family string, err error) {
	layout := graphstore.SchemaVersion
	if graphPreviewConfig != nil {
		layout = graphPreviewConfig.GraphSchemaVersion
	}
	return graphPreviewCollectionTypesForLayoutWithDefault(cmd, scope, links, layout, issueDefault)
}

func graphPreviewCollectionTypesForLayout(cmd *cobra.Command, scope string, links bool, layout int) (exact, family string, err error) {
	return graphPreviewCollectionTypesForLayoutWithDefault(cmd, scope, links, layout, true)
}

func graphPreviewCollectionTypesForLayoutWithDefault(cmd *cobra.Command, scope string, links bool, layout int, issueDefault bool) (exact, family string, err error) {
	changed := func(name string) bool {
		return cmd.Flags().Lookup(name) != nil && cmd.Flags().Changed(name)
	}
	if changed("type") && changed("conforms-to") {
		return "", "", fmt.Errorf("--type and --conforms-to both select the Type family; supply only one")
	}
	if !changed("type") && !changed("exact-type") && !changed("conforms-to") {
		if !links && issueDefault && layout != graphstore.TypesFreshSchemaVersion {
			return "", graphstore.IssueTypeURL(scope), nil
		}
		return "", "", nil
	}
	if changed("exact-type") {
		raw, _ := cmd.Flags().GetString("exact-type")
		if raw == "" {
			return "", "", fmt.Errorf("--exact-type requires a Type")
		}
		if raw == "any" {
			return "", "", fmt.Errorf("any is the --type wildcard, not a Type name")
		}
		exact, err = graphPreviewCollectionTypeURL(scope, raw, links)
		if err != nil {
			return "", "", err
		}
	}
	if changed("conforms-to") {
		raw, _ := cmd.Flags().GetString("conforms-to")
		// Preserve the existing canonical-URI alias; product shorthand belongs
		// to --type and does not change raw BDP relative-operand semantics.
		if !strings.Contains(raw, "://") {
			return "", "", fmt.Errorf("--conforms-to requires a canonical absolute HTTP(S) Type URI")
		}
		if err := graph.ValidateTypeURL(raw); err != nil {
			return "", "", err
		}
		return exact, raw, nil
	}
	if !changed("type") {
		return exact, "", nil
	}
	raw, _ := cmd.Flags().GetString("type")
	if raw == "" {
		return "", "", fmt.Errorf("--type requires a Type")
	}
	if raw == "any" {
		return exact, "", nil
	}
	family, err = graphPreviewCollectionTypeURL(scope, raw, links)
	if err != nil {
		return "", "", err
	}
	return exact, family, nil
}

func graphPreviewCollectionTypeURL(scope, raw string, links bool) (string, error) {
	if strings.Contains(raw, "://") {
		if err := graph.ValidateTypeURL(raw); err != nil {
			return "", err
		}
		return raw, nil
	}
	// Bare native Issue shorthand retains the shipping alias policy shared by
	// create/update and collection/count. Full URLs and literal segmented IDs
	// are handled as complete identities, never normalized.
	if !links && !graphPreviewNominalSelector(raw) {
		classification := types.IssueType(raw).Normalize()
		if classification.IsBuiltIn() {
			return graphstore.NativeIssueTypeURL(scope, string(classification))
		}
	}
	// A canonical Scope-local ID is already a complete value, including its
	// percent escapes and literal path segments. Only a raw noncanonical native
	// label reaches encoding at this role-specific input boundary.
	if id, err := graph.CanonicalPublicResourceURL(scope, raw); err == nil {
		return id, nil
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
		// Collection reads select the exact stored native kind. The
		// depends-on/blocked-by aliases belong only to dep add writes.
		return graphstore.NativeDependencyTypeURL(scope, raw)
	}
}

// graphPreviewReadBeadDocuments resolves product selectors and pages one checked
// capture. A local definition identity is not an alias in the raw BDP predicate.
func graphPreviewReadBeadDocuments(ctx context.Context, store *graphstore.Store, request graphstore.BeadListRequest) (graphstore.BeadDocumentPage, error) {
	snapshot, err := store.CurrentResourceDocuments(ctx)
	if err != nil {
		return graphstore.BeadDocumentPage{}, err
	}
	return graphPreviewSelectCapturedBeads(snapshot, request)
}

func graphPreviewSelectCapturedBeads(snapshot graphstore.ResourceSnapshot, request graphstore.BeadListRequest) (graphstore.BeadDocumentPage, error) {
	if request.TypeURL != "" {
		selected, err := snapshot.ResolveType(request.TypeURL)
		if err != nil {
			return graphstore.BeadDocumentPage{}, err
		}
		request.TypeURL = selected.URI
	}
	if request.ConformsTo != "" {
		selected, err := snapshot.ResolveType(request.ConformsTo)
		if err != nil {
			return graphstore.BeadDocumentPage{}, err
		}
		request.ConformsTo = selected.URI
	}
	return snapshot.SelectBeadDocuments(request)
}
