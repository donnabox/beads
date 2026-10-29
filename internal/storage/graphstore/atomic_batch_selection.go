package graphstore

import (
	"context"
	"fmt"
	"sort"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/canonicalbatch"
	"github.com/steveyegge/beads/internal/canonicalbatchstate"
)

// Authority comes from the complete checked descriptor inventory acquired by
// the batch transaction. A populated Resource set cannot establish ancestry or
// prove the absence of an abstract Type. No local-name inference is involved.
func atomicBatchTypeAuthority(ctx context.Context, descriptors []graph.TypeDescriptor) (map[string]graph.ResourceKind, map[string][]string, error) {
	kinds := map[string]graph.ResourceKind{}
	parents := map[string][]string{}
	for _, descriptor := range descriptors {
		if err := ctx.Err(); err != nil {
			return nil, nil, err
		}
		id := descriptor.ID()
		if id == "" || !descriptor.Describes().Valid() || kinds[id] != "" {
			return nil, nil, fmt.Errorf("%w: invalid checked Type inventory", ErrInvalidStore)
		}
		kinds[id], parents[id] = descriptor.Describes(), descriptor.ConformsTo()
	}
	families := map[string][]string{}
	work := 0
	for id, kind := range kinds {
		seen := map[string]bool{}
		pending := []string{id}
		for len(pending) != 0 {
			if err := ctx.Err(); err != nil {
				return nil, nil, err
			}
			work++
			if work > canonicalbatchstate.MaxSelectionWork {
				return nil, nil, fmt.Errorf("%w: checked Type ancestry work", ErrLimitExceeded)
			}
			ancestor := pending[len(pending)-1]
			pending = pending[:len(pending)-1]
			if kinds[ancestor] == "" || kinds[ancestor] != kind {
				return nil, nil, fmt.Errorf("%w: incomplete checked Type ancestry", ErrInvalidStore)
			}
			if seen[ancestor] {
				continue
			}
			seen[ancestor] = true
			pending = append(pending, parents[ancestor]...)
		}
		ancestors := make([]string, 0, len(seen))
		for ancestor := range seen {
			ancestors = append(ancestors, ancestor)
		}
		sort.Slice(ancestors, func(i, j int) bool { return graph.CompareCodeUnits(ancestors[i], ancestors[j]) < 0 })
		families[id] = ancestors
	}
	return kinds, families, nil
}

// Every member has one identity array, including empty selections and empty set
// writes. These are operation mappings, not intermediate public postimages.
func atomicBatchOperationResults(members []canonicalbatch.Member, plan canonicalbatchstate.Plan) ([]BatchOperationResult, error) {
	setMembers := atomicBatchSetMembers(members)
	result := make([]BatchOperationResult, len(members))
	for i, member := range members {
		result[i] = BatchOperationResult{OperationIndex: i, Operation: member.Operation, OperationName: member.Name, IDs: []string{}}
	}
	for _, mapping := range plan.Mappings {
		if mapping.OperationIndex < 0 || mapping.OperationIndex >= len(result) || result[mapping.OperationIndex].Operation != mapping.Operation || mapping.ID == "" {
			return nil, fmt.Errorf("%w: invalid staged operation mapping", ErrInvalidStore)
		}
		row := &result[mapping.OperationIndex]
		row.IDs = append(row.IDs, mapping.ID)
	}
	for _, capture := range plan.Selections {
		if capture.OperationIndex < 0 || capture.OperationIndex >= len(result) || result[capture.OperationIndex].Operation != "select" {
			return nil, fmt.Errorf("%w: invalid staged selection mapping", ErrInvalidStore)
		}
		result[capture.OperationIndex].IDs = append([]string{}, capture.IDs...)
	}
	for i := range result {
		ids := result[i].IDs
		sort.Slice(ids, func(i, j int) bool { return graph.CompareCodeUnits(ids[i], ids[j]) < 0 })
		for j := 1; j < len(ids); j++ {
			if ids[j] == ids[j-1] {
				return nil, fmt.Errorf("%w: duplicate staged identity mapping", ErrInvalidStore)
			}
		}
		if members[i].Operation != "select" && !setMembers[i] && len(ids) != 1 {
			return nil, fmt.Errorf("%w: missing staged singleton mapping", ErrInvalidStore)
		}
	}
	return result, nil
}

// Set cardinality belongs to the defining select, not to a separate write flag.
// A creation binding remains scalar, even if its target syntax is identical.
func atomicBatchSetMembers(members []canonicalbatch.Member) []bool {
	sets := make([]bool, len(members))
	definitions := map[string]bool{}
	for i, member := range members {
		binding := member.Bead.Binding
		if member.Link.Binding != "" {
			binding = member.Link.Binding
		}
		sets[i] = definitions[binding]
		if member.Name != "" {
			definitions[member.Name] = member.Operation == "select"
		}
	}
	return sets
}
