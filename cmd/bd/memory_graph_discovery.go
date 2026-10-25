package main

import (
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// Discovery scope is validated before choosing any output mode, including quiet.
// This consumes only the captured incidence and phase witnesses, never Owned Links.
func validateMemoryDiscoveryCapture(read graphstore.CurrentFamilyRead, staged *graphstore.CurrentConstraintsRead, scope, operation string) error {
	if operation != "subjects" && operation != "constraints" {
		return nil
	}
	c, err := bindMemoryConstraintCapture(read, scope)
	if err != nil {
		return err
	}
	if operation == "constraints" && (staged == nil || len(staged.Stages) != len(read.Graph.Views)) {
		return invalidMemoryConstraintCapture()
	}
	for slot, view := range read.Graph.Views {
		selected, err := c.view(view)
		if err != nil {
			return err
		}
		if operation == "subjects" {
			if view.Root.Depth != memoryDirectSubjectsDepth || len(view.Frontier) != 0 {
				return invalidMemoryConstraintCapture()
			}
			members := map[string]bool{view.Root.Path: true}
			for _, path := range view.Links {
				link := c.links[path]
				if link.Target != selected.Root {
					return invalidMemoryConstraintCapture()
				}
				source, ok := graphPreviewPublicResourceID(scope, link.Source)
				if !ok {
					return invalidMemoryConstraintCapture()
				}
				members[source] = true
			}
			if !sameMemoryConstraintSet(view.Nodes, members) {
				return invalidMemoryConstraintCapture()
			}
			continue
		}
		stage := staged.Stages[slot]
		if stage.Root.Path != view.Root.Path || stage.Root.RequiresDepth != view.Root.Depth {
			return invalidMemoryConstraintCapture()
		}
		phaseNodes, err := memoryConstraintSet(stage.Requires.Nodes)
		if err != nil {
			return err
		}
		phaseLinks, err := memoryConstraintSet(stage.Requires.Links)
		if err != nil {
			return err
		}
		seeds, err := memoryConstraintSet(stage.Requires.Seeds)
		if err != nil {
			return err
		}
		unionNodes := map[string]bool{view.Root.Path: true}
		unionLinks, targets := map[string]bool{}, map[string]bool{}
		outgoing := map[string][]string{}
		for node := range phaseNodes {
			unionNodes[node] = true
		}
		for path := range phaseLinks {
			link, ok := c.links[path]
			source, src := graphPreviewPublicResourceID(scope, link.Source)
			target, dst := graphPreviewPublicResourceID(scope, link.Target)
			if !ok || !src || !dst || !phaseNodes[source] || !phaseNodes[target] {
				return invalidMemoryConstraintCapture()
			}
			unionLinks[path] = true
			outgoing[source] = append(outgoing[source], target)
		}
		for _, seed := range stage.SubjectTo {
			link, ok := c.links[seed.Link]
			if !ok || unionLinks[seed.Link] || link.Source != selected.Root || link.Target != scope+seed.Target || !seeds[seed.Target] {
				return invalidMemoryConstraintCapture()
			}
			unionLinks[seed.Link] = true
			targets[seed.Target] = true
		}
		frontier, err := memoryConstraintSet(stage.Requires.Frontier)
		if err != nil {
			return err
		}
		if !sameMemoryConstraintSet(stage.Requires.Seeds, targets) || !sameMemoryConstraintSet(view.Nodes, unionNodes) || !sameMemoryConstraintSet(view.Links, unionLinks) || !sameMemoryConstraintSet(view.Frontier, frontier) {
			return invalidMemoryConstraintCapture()
		}
		reached, queue := map[string]bool{}, []string{}
		for seed := range seeds {
			if !phaseNodes[seed] {
				return invalidMemoryConstraintCapture()
			}
			reached[seed] = true
			queue = append(queue, seed)
		}
		for i := 0; i < len(queue); i++ {
			for _, target := range outgoing[queue[i]] {
				if !reached[target] {
					reached[target] = true
					queue = append(queue, target)
				}
			}
		}
		if !sameMemoryConstraintSet(stage.Requires.Nodes, reached) {
			return invalidMemoryConstraintCapture()
		}
		for node := range frontier {
			if !phaseNodes[node] {
				return invalidMemoryConstraintCapture()
			}
		}
	}
	return nil
}
