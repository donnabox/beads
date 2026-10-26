package main

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

type graphTypeInventory struct {
	BeadTypes []json.RawMessage `json:"beadTypes"`
	LinkTypes []json.RawMessage `json:"linkTypes"`
}

// Inventory comes from checked workspace declarations, including immutable
// used native bindings and configured unused native Issue definitions.
func runGraphPreviewTypes(cmd *cobra.Command, args []string) error {
	if err := graphPreviewFlags(cmd, "details"); err != nil {
		return err
	}
	if len(args) != 0 {
		return graphFailure("invalid_selector", "bd types takes no arguments", 2)
	}
	details, _ := cmd.Flags().GetBool("details")
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		descriptors, err := store.ListInstalledTypes(ctx)
		if err != nil {
			return nil, "", err
		}
		return graphPreviewTypesInventory(store.ScopeURL(), descriptors, details, store.AssessmentsEnabled())
	})
}

func graphPreviewTypeTitle(name string) string {
	parts := strings.FieldsFunc(name, func(r rune) bool { return r == '-' || r == '_' })
	for i, part := range parts {
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}
