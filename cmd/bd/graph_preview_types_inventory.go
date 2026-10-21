package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
)

// graphPreviewTypesInventory renders the provider's selected definitions. It
// neither installs definitions nor assumes a particular Memory catalog or parent
// graph. The default text expands the existing Issue/Dependency classifications;
// JSON and details describe the installed nominal Types themselves.
func graphPreviewTypesInventory(scope string, selected []graph.TypeDescriptor, customTypes []string, details bool) (graphTypeInventory, string, error) {
	descriptors := append([]graph.TypeDescriptor(nil), selected...)
	sort.Slice(descriptors, func(i, j int) bool {
		return graph.CompareCodeUnits(descriptors[i].ID(), descriptors[j].ID()) < 0
	})
	inventory := graphTypeInventory{BeadTypes: []json.RawMessage{}, LinkTypes: []json.RawMessage{}}
	var beads, links strings.Builder
	beadCount, linkCount := 0, 0
	add := func(out *strings.Builder, count *int, name, description string) {
		fmt.Fprintf(out, "  %-25s %s\n", graphMemoryDisplayText(name), graphMemoryDisplayText(description))
		*count = *count + 1
	}
	for _, descriptor := range descriptors {
		if graphPreviewIntrinsicType(descriptor) {
			continue
		}
		var out *strings.Builder
		var count *int
		switch descriptor.Describes() {
		case graph.KindBead:
			inventory.BeadTypes = append(inventory.BeadTypes, json.RawMessage(descriptor.CanonicalJSON()))
			out, count = &beads, &beadCount
		case graph.KindLink:
			inventory.LinkTypes = append(inventory.LinkTypes, json.RawMessage(descriptor.CanonicalJSON()))
			out, count = &links, &linkCount
		default:
			return graphTypeInventory{}, "", fmt.Errorf("%w: installed Type has unknown resource category", graphstore.ErrInvalidStore)
		}
		if !details && descriptor.Describes() == graph.KindBead && descriptor.ID() == graphstore.IssueTypeURL(scope) {
			for _, entry := range coreWorkTypes {
				name := string(entry.Type)
				add(out, count, name, graphPreviewTypeTitle(name)+" Issue")
			}
			for _, entry := range systemWorkTypes() {
				add(out, count, entry.Name, graphPreviewTypeTitle(entry.Name)+" Issue")
			}
			for _, name := range customTypes {
				add(out, count, name, graphPreviewTypeTitle(name)+" Issue")
			}
			continue
		}
		if !details && descriptor.Describes() == graph.KindLink && descriptor.ID() == graphstore.DependencyTypeURL(scope) {
			for _, kind := range types.WellKnownDependencyTypes() {
				name, title := string(kind), graphPreviewTypeTitle(string(kind))
				if kind == types.DepBlocks {
					title = "Blocking"
				}
				if kind == types.DepParentChild {
					title = "Parent/Child"
				}
				add(out, count, name, title+" Dependency")
			}
			continue
		}
		id := strings.TrimPrefix(descriptor.ID(), scope)
		add(out, count, id, descriptor.Name())
		if details {
			var pretty bytes.Buffer
			if err := json.Indent(&pretty, descriptor.CanonicalJSON(), "", "  "); err != nil {
				return graphTypeInventory{}, "", err
			}
			for _, line := range strings.Split(pretty.String(), "\n") {
				fmt.Fprintf(out, "    %s\n", line)
			}
		}
	}
	prefix := ""
	if details {
		prefix = "Installed "
	}
	return inventory, fmt.Sprintf("%sBead Types (%d):\n%s\n%sLink Types (%d):\n%s", prefix, beadCount, beads.String(), prefix, linkCount, links.String()), nil
}

// The legacy descriptor has no visibility member. Reserve both intrinsic names
// and their local identity spellings until the selected registry provides its
// explicit public projection. Refines is the replacement for ConformsTo; the
// ordinary Memory Link Type Conformance remains public.
func graphPreviewIntrinsicType(descriptor graph.TypeDescriptor) bool {
	reserved := func(value string) bool {
		value = strings.ToLower(strings.NewReplacer("-", "", "_", "", " ", "").Replace(value))
		switch value {
		case "type", "beadtype", "linktype", "conformsto", "refines":
			return true
		}
		return false
	}
	id := descriptor.ID()
	return reserved(descriptor.Name()) || reserved(id[strings.LastIndex(id, "/")+1:])
}
