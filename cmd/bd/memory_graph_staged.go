package main

import (
	"fmt"
	"slices"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// Phase membership is query context, not Resource properties or a new BDP IR.
// Keep it inside each protected root view; union membership cannot reconstruct it.
type memoryGraphStages struct {
	SubjectTo []memoryGraphSeed        `json:"subjectTo"`
	Requires  memoryGraphRequiresStage `json:"requires"`
}
type memoryGraphSeed struct {
	Link   string `json:"link"`
	Target string `json:"target"`
}
type memoryGraphRequiresStage struct {
	Depth    int      `json:"depth"`
	Seeds    []string `json:"seeds"`
	Beads    []string `json:"beads"`
	Links    []string `json:"links"`
	Frontier []string `json:"frontier"`
}

func memoryConstraintStages(result *memoryGraphResult, read graphstore.CurrentConstraintsRead) error {
	invalid := func() error {
		return fmt.Errorf("%w: inconsistent Constraint stage context", graphstore.ErrInvalidStore)
	}
	if len(read.Stages) != len(result.Traversal.Roots) {
		return invalid()
	}
	// Charge the already bounded ordinary graph before adding duplicated stage
	// references; a stage refusal never emits part of a successful document.
	raw, err := encodeMemoryGraphJSON(result)
	if err != nil {
		return err
	}
	remaining := graphstore.PreviewCurrentReadByteLimit - len(raw)
	canonical := func(paths []string) []string {
		ids := make([]string, 0, len(paths))
		for _, path := range paths {
			ids = append(ids, graph.CanonicalURL(result.Scope, path))
		}
		return ids
	}
	projectedStages := make([]*memoryGraphStages, 0, len(read.Stages))
	linkEndpoints := map[string][2]string{}
	for _, l := range result.Links {
		linkEndpoints[l.ID] = [2]string{l.Source.URI, l.Target.URI}
	}
	sameSet := func(a, b []string) bool {
		x, y := slices.Clone(a), slices.Clone(b)
		slices.Sort(x)
		slices.Sort(y)
		return slices.Equal(x, y)
	}
	for i, stage := range read.Stages {
		root := &result.Traversal.Roots[i]
		if root.ID != graph.CanonicalURL(result.Scope, stage.Root.Path) || (root.Depth != nil && *root.Depth != stage.Root.RequiresDepth) {
			return invalid()
		}
		nodes, links := map[string]bool{}, map[string]bool{}
		for _, id := range root.Beads {
			nodes[id] = true
		}
		for _, id := range root.Links {
			links[id] = true
		}
		projected := &memoryGraphStages{SubjectTo: []memoryGraphSeed{}}
		phaseNodes := map[string]bool{}
		for _, path := range stage.Requires.Nodes {
			id := graph.CanonicalURL(result.Scope, path)
			if !nodes[id] || phaseNodes[id] {
				return invalid()
			}
			phaseNodes[id] = true
		}
		targets := []string{}
		edgeIDs := []string{}
		declarations, requirements := map[string]bool{}, map[string]bool{}
		for _, seed := range stage.SubjectTo {
			id, target := graph.CanonicalURL(result.Scope, seed.Link), graph.CanonicalURL(result.Scope, seed.Target)
			if !links[id] || declarations[id] || !phaseNodes[target] || linkEndpoints[id] != [2]string{root.ID, target} {
				return invalid()
			}
			declarations[id] = true
			projected.SubjectTo = append(projected.SubjectTo, memoryGraphSeed{Link: id, Target: target})
			targets = append(targets, seed.Target)
			edgeIDs = append(edgeIDs, id)
		}
		slices.Sort(targets)
		targets = slices.Compact(targets)
		if !sameSet(targets, stage.Requires.Seeds) {
			return invalid()
		}
		for _, path := range stage.Requires.Links {
			id := graph.CanonicalURL(result.Scope, path)
			ends, ok := linkEndpoints[id]
			if !ok || !links[id] || declarations[id] || requirements[id] || !phaseNodes[ends[0]] || !phaseNodes[ends[1]] {
				return invalid()
			}
			requirements[id] = true
			edgeIDs = append(edgeIDs, id)
		}
		for _, path := range stage.Requires.Frontier {
			if !phaseNodes[graph.CanonicalURL(result.Scope, path)] {
				return invalid()
			}
		}
		slices.Sort(edgeIDs)
		edgeIDs = slices.Compact(edgeIDs)
		if !sameSet(edgeIDs, root.Links) {
			return invalid()
		}
		combinedNodes := make([]string, 0, len(phaseNodes)+1)
		for id := range phaseNodes {
			combinedNodes = append(combinedNodes, id)
		}
		combinedNodes = append(combinedNodes, root.ID)
		slices.Sort(combinedNodes)
		combinedNodes = slices.Compact(combinedNodes)
		if !sameSet(combinedNodes, root.Beads) {
			return invalid()
		}
		// Every phase member must be reached forward from an explicit seed.
		reached, queue := map[string]bool{}, []string{}
		for _, seed := range stage.Requires.Seeds {
			id := graph.CanonicalURL(result.Scope, seed)
			if !phaseNodes[id] {
				return invalid()
			}
			reached[id] = true
			queue = append(queue, id)
		}
		outgoing := map[string][]string{}
		for id := range requirements {
			ends := linkEndpoints[id]
			outgoing[ends[0]] = append(outgoing[ends[0]], ends[1])
		}
		for i := 0; i < len(queue); i++ {
			for _, target := range outgoing[queue[i]] {
				if !reached[target] {
					reached[target] = true
					queue = append(queue, target)
				}
			}
		}
		if len(reached) != len(phaseNodes) {
			return invalid()
		}
		projected.Requires.Depth = stage.Root.RequiresDepth
		projected.Requires.Seeds = canonical(stage.Requires.Seeds)
		projected.Requires.Beads = canonical(stage.Requires.Nodes)
		projected.Requires.Links = canonical(stage.Requires.Links)
		projected.Requires.Frontier = canonical(stage.Requires.Frontier)
		if !sameSet(projected.Requires.Frontier, root.Frontier) {
			return invalid()
		}
		encoded, err := encodeMemoryGraphJSON(projected)
		if err != nil {
			return err
		}
		if len(encoded)+16 > remaining {
			return fmt.Errorf("%w: Constraint stage output budget", graphstore.ErrLimitExceeded)
		}
		remaining -= len(encoded) + 16
		projectedStages = append(projectedStages, projected)
	}
	for i, stage := range projectedStages {
		result.Traversal.Roots[i].Stages = stage
	}
	return nil
}

// Disclosure qualifies the operation's discovery scope, once per human document.
// It is outside the per-root forest and is never emitted as a JSON data row.
func memoryDiscoveryLine(operation string) string {
	switch operation {
	case "constraints":
		return "Discovery: explicit Subject To declarations and their Requires dependencies\n"
	case "subjects":
		return "Discovery: direct incoming Subject To declarations only\n"
	default:
		return ""
	}
}
