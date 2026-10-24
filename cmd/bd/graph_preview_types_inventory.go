package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// graphPreviewTypesInventory renders the provider's selected definitions. It
// neither installs definitions nor assumes a particular Memory catalog or parent
// graph. Text, JSON and details describe the actual declared Types; native
// classification shorthands do not form a second inventory authority.
func graphPreviewTypesInventory(scope string, selected []graph.TypeDescriptor, details, assessmentsEnabled bool) (graphTypeInventory, string, error) {
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
		if (!graphstore.IsNativeConcreteTypeDescriptor(scope, descriptor) && graphPreviewIntrinsicType(descriptor)) || !graphstore.AssessmentTypeAvailable(scope, descriptor.ID(), assessmentsEnabled) {
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
		id := strings.TrimPrefix(descriptor.ID(), scope)
		label := descriptor.Name()
		if descriptor.ID() == graphstore.IssueTypeURL(scope) || descriptor.ID() == graphstore.DependencyTypeURL(scope) {
			label += " (abstract)"
		}
		add(out, count, id, label)
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
	var text strings.Builder
	fmt.Fprintf(&text, "%sBead Types (%d):\n%s", prefix, beadCount, beads.String())
	fmt.Fprintf(&text, "\n%sLink Types (%d):\n%s", prefix, linkCount, links.String())
	return inventory, text.String(), nil
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
