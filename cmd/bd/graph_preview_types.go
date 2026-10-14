package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
)

type graphTypeInventory struct {
	BeadTypes []json.RawMessage `json:"beadTypes"`
	LinkTypes []json.RawMessage `json:"linkTypes"`
}

// The source of this inventory is the selected workspace's checked, persisted
// Type installation.
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
		inventory := graphTypeInventory{
			BeadTypes: []json.RawMessage{}, LinkTypes: []json.RawMessage{},
		}
		var beadText, linkText strings.Builder
		for _, descriptor := range descriptors {
			id := strings.TrimPrefix(descriptor.ID(), store.ScopeURL())
			var destination *strings.Builder
			switch descriptor.Describes() {
			case graph.KindBead:
				inventory.BeadTypes = append(inventory.BeadTypes, json.RawMessage(descriptor.CanonicalJSON()))
				destination = &beadText
			case graph.KindLink:
				inventory.LinkTypes = append(inventory.LinkTypes, json.RawMessage(descriptor.CanonicalJSON()))
				destination = &linkText
			default:
				return nil, "", fmt.Errorf("%w: installed Type has unknown resource category", graphstore.ErrInvalidStore)
			}
			fmt.Fprintf(destination, "  %-30s %s\n", graphMemoryDisplayText(id), graphMemoryDisplayText(descriptor.Name()))
			if details {
				var pretty bytes.Buffer
				if err := json.Indent(&pretty, descriptor.CanonicalJSON(), "    ", "  "); err != nil {
					return nil, "", err
				}
				for _, line := range strings.Split(pretty.String(), "\n") {
					fmt.Fprintf(destination, "    %s\n", line)
				}
			}
		}
		if !details {
			customTypes, err := resolveWorkspaceCustomTypes(ctx)
			if err != nil {
				return nil, "", err
			}
			return inventory, graphPreviewEffectiveTypesText(customTypes), nil
		}
		human := fmt.Sprintf("Installed Bead Types (%d):\n%s\nInstalled Link Types (%d):\n%s", len(inventory.BeadTypes), beadText.String(), len(inventory.LinkTypes), linkText.String())
		return inventory, human, nil
	})
}

// The default human view lists usable --type selectors. Issue and Dependency
// values share their respective installed graph Types; --json and --details
// continue to expose the actual installed descriptors.
func graphPreviewEffectiveTypesText(customTypes []string) string {
	var beads, links strings.Builder
	beadCount, linkCount := 0, 0
	add := func(out *strings.Builder, count *int, name, description string) {
		fmt.Fprintf(out, "  %-25s %s\n", name, description)
		*count = *count + 1
	}
	add(&beads, &beadCount, "preview-memory", "Memory Preview")
	for _, entry := range coreWorkTypes {
		name := string(entry.Type)
		add(&beads, &beadCount, name, graphPreviewTypeTitle(name)+" Issue")
	}
	for _, entry := range systemWorkTypes() {
		add(&beads, &beadCount, entry.Name, graphPreviewTypeTitle(entry.Name)+" Issue")
	}
	for _, name := range customTypes {
		add(&beads, &beadCount, name, graphPreviewTypeTitle(name)+" Issue")
	}
	add(&links, &linkCount, "preview-related-memory", "Related Memory Preview")
	add(&links, &linkCount, "example-cites", "Example Cites")
	add(&links, &linkCount, "example-follows", "Example Follows")
	for _, kind := range types.WellKnownDependencyTypes() {
		name := string(kind)
		title := graphPreviewTypeTitle(name)
		if kind == types.DepBlocks {
			title = "Blocking"
		} else if kind == types.DepParentChild {
			title = "Parent/Child"
		}
		add(&links, &linkCount, name, title+" Dependency")
	}
	return fmt.Sprintf("Bead Types (%d):\n%s\nLink Types (%d):\n%s", beadCount, beads.String(), linkCount, links.String())
}

func graphPreviewTypeTitle(name string) string {
	parts := strings.FieldsFunc(name, func(r rune) bool { return r == '-' || r == '_' })
	for i, part := range parts {
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}
