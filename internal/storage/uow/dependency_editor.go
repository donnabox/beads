package uow

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/steveyegge/beads/internal/storage/domain"
	storageissueops "github.com/steveyegge/beads/internal/storage/issueops"
	"github.com/steveyegge/beads/internal/types"
	publicops "github.com/steveyegge/beads/issueops"
)

// DependencyEditorSource is the capability accessor a unit-of-work provider
// offers for the dependency-edge role, the sibling of IssueLifecycleSource and
// BatchCloserSource.
type DependencyEditorSource interface {
	DependencyEditor() (publicops.DependencyEditor, error)
}

// dependencyEditor edits the dependency graph through a unit of work.
type dependencyEditor struct {
	provider UnitOfWorkProvider
}

// DependencyEditor returns the guarded dependency-edge surface for this provider.
func (p *doltSQLProvider) DependencyEditor() (publicops.DependencyEditor, error) {
	return NewDependencyEditor(p)
}

// NewDependencyEditor constructs a public dependency editor backed by provider.
func NewDependencyEditor(provider UnitOfWorkProvider) (publicops.DependencyEditor, error) {
	if isNilUnitOfWorkProvider(provider) {
		return nil, fmt.Errorf("new dependency editor: unit-of-work provider must not be nil")
	}
	return &dependencyEditor{provider: provider}, nil
}

var _ publicops.DependencyEditor = (*dependencyEditor)(nil)

// AddDependencies asserts every edge in ONE unit of work. The request is the
// transaction: N edges are one commit, and the first refusal returns an error
// that rolls every earlier edge back with it, which is what makes the request
// all-or-nothing without any bookkeeping of its own.
//
// The use case applies the parent-child-first ordering and the whole-graph
// final gate the store-backed sibling applies, through the same rules, so the
// two answer the same refusals to the same requests. It is the SOURCE-ROUTED
// bulk verb for the same reason: each edge follows the plane its source lives
// in, and a request may mix them, so the two implementations also land the
// same edges in the same tables.
func (e *dependencyEditor) AddDependencies(ctx context.Context, request publicops.AddDependenciesRequest) (publicops.AddDependenciesResult, error) {
	if err := storageissueops.ValidateAddDependenciesRequest(request); err != nil {
		return publicops.AddDependenciesResult{}, err
	}
	deps := make([]*types.Dependency, 0, len(request.Edges))
	for _, edge := range request.Edges {
		deps = append(deps, &types.Dependency{
			IssueID:     edge.IssueID,
			DependsOnID: edge.DependsOnID,
			Type:        edge.Type,
			Metadata:    edge.Metadata,
			ThreadID:    edge.ThreadID,
		})
	}
	added := make([]publicops.DependencyEdge, len(request.Edges))
	copy(added, request.Edges)

	write := func() (publicops.AddDependenciesResult, error) {
		return RunTxResult(ctx, e.provider, func(ctx context.Context, uw UnitOfWork) (publicops.AddDependenciesResult, string, error) {
			if request.Exact {
				if err := checkExactEdgesInUOW(ctx, uw, request.Edges); err != nil {
					return publicops.AddDependenciesResult{}, "", err
				}
			}
			if _, err := uw.DependencyUseCase().AddDependencies(ctx, deps, request.Actor, domain.BulkAddDepsOpts{
				SkipPerEdgeCycleCheck: request.SkipPerEdgeCycleCheck,
			}); err != nil {
				return publicops.AddDependenciesResult{}, "", err
			}
			result := publicops.AddDependenciesResult{Added: added}
			if request.Exact {
				stored, err := storedEdgesInUOW(ctx, uw, request.Edges)
				if err != nil {
					return publicops.AddDependenciesResult{}, "", err
				}
				result.StoredEdges = stored
			}
			return result, storageissueops.AddDependenciesCommitMessage(request), nil
		})
	}
	if !request.Exact {
		return write()
	}
	// An exact edge runs under the delete fence on each of its row endpoints. Two
	// creates of one pair in one second are identical diffs that Dolt lands with no
	// conflict, so both would report having made it, and a create racing a guarded
	// delete of its endpoint could leave an edge on a row that is gone. See
	// withDeleteFence, and Exact's own request rules, for what the queue buys.
	var result publicops.AddDependenciesResult
	err := withEdgeFences(ctx, e.provider, exactEdgeFenceKeys(request.Edges), func() error {
		var err error
		result, err = write()
		return err
	})
	if err != nil {
		return publicops.AddDependenciesResult{}, err
	}
	return result, nil
}

// exactEdgeFenceKeys is every row endpoint named, in one order, so that two requests
// over the same rows cannot wait on each other's locks.
func exactEdgeFenceKeys(edges []publicops.DependencyEdge) []string {
	keys := make([]string, 0, 2*len(edges))
	for _, edge := range edges {
		keys = append(keys, edge.IssueID)
		if !strings.HasPrefix(edge.DependsOnID, "external:") {
			keys = append(keys, edge.DependsOnID)
		}
	}
	slices.Sort(keys)
	return slices.Compact(keys)
}

// withEdgeFences runs fn holding the delete fence on each key, on a provider that has
// one to take.
func withEdgeFences(ctx context.Context, provider UnitOfWorkProvider, keys []string, fn func() error) error {
	sp, ok := fenceTarget(provider)
	if !ok || len(keys) == 0 {
		return fn()
	}
	return withDeleteFence(ctx, sp, keys[0], func() error {
		return withEdgeFences(ctx, provider, keys[1:], fn)
	})
}

// checkExactEdgesInUOW is AddDependenciesRequest.Exact's refusals.
func checkExactEdgesInUOW(ctx context.Context, uw UnitOfWork, edges []publicops.DependencyEdge) error {
	for _, edge := range edges {
		existing, err := uw.DependencyUseCase().GetForIssueIDs(ctx, []string{edge.IssueID})
		if err != nil {
			return err
		}
		for _, dep := range existing[edge.IssueID] {
			if dep != nil && dep.DependsOnID == edge.DependsOnID {
				return &publicops.DependencyOccupiedError{Occupant: dep}
			}
		}
		if strings.HasPrefix(edge.DependsOnID, "external:") {
			continue
		}
		if _, _, err := operationIssue(ctx, uw, edge.DependsOnID, false); err != nil {
			if errors.Is(err, publicops.ErrNotFound) {
				return &publicops.DependencyEndpointNotFoundError{
					IssueID: edge.IssueID, DependsOnID: edge.DependsOnID, MissingID: edge.DependsOnID,
					Err: publicops.ErrDependencyTargetNotFound,
				}
			}
			return err
		}
	}
	return nil
}

// storedEdgesInUOW reads back the rows this unit of work wrote.
func storedEdgesInUOW(ctx context.Context, uw UnitOfWork, edges []publicops.DependencyEdge) ([]*types.Dependency, error) {
	stored := make([]*types.Dependency, 0, len(edges))
	for _, edge := range edges {
		rows, err := uw.DependencyUseCase().GetForIssueIDs(ctx, []string{edge.IssueID})
		if err != nil {
			return nil, err
		}
		for _, dep := range rows[edge.IssueID] {
			if dep != nil && dep.DependsOnID == edge.DependsOnID {
				stored = append(stored, dep)
				break
			}
		}
	}
	return stored, nil
}

// RemoveDependency removes one edge in one unit of work.
//
// It is SOURCE-ROUTED, like the add: an edge lives in the plane its source
// lives in, so the removal reads that plane rather than pinning the durable
// one — a pinned removal cannot see a wisp-sourced edge and would report an
// edge that is plainly there as missing (bd-yby99.17).
//
// The delete inside the transaction is the found/not-found verdict. Taking it
// from the write itself rather than from a lookup before it is what keeps the
// answer true: a separate check is a read-then-write window in which a racing
// writer could add or drop the very edge being reported on.
func (e *dependencyEditor) RemoveDependency(ctx context.Context, request publicops.RemoveDependencyRequest) (publicops.RemoveDependencyResult, error) {
	if err := storageissueops.ValidateRemoveDependencyRequest(request); err != nil {
		return publicops.RemoveDependencyResult{}, err
	}
	write := func() (publicops.RemoveDependencyResult, error) {
		return RunTxResult(ctx, e.provider, func(ctx context.Context, uw UnitOfWork) (publicops.RemoveDependencyResult, string, error) {
			if request.InTx != nil {
				existing, err := uw.DependencyUseCase().GetForIssueIDs(ctx, []string{request.IssueID})
				if err != nil {
					return publicops.RemoveDependencyResult{}, "", err
				}
				found := false
				for _, dep := range existing[request.IssueID] {
					if dep != nil && dep.DependsOnID == request.DependsOnID {
						found = true
						if err := request.InTx.Run(dep); err != nil {
							return publicops.RemoveDependencyResult{}, "", err
						}
						break
					}
				}
				if !found {
					return publicops.RemoveDependencyResult{Removed: false}, "", nil
				}
			}
			removed, err := uw.DependencyUseCase().RemoveDependencyBySource(ctx, request.IssueID, request.DependsOnID, request.Actor)
			if err != nil {
				return publicops.RemoveDependencyResult{}, "", err
			}
			if !removed {
				// Nothing to write, so nothing to version: an empty commit message
				// is how RunTxResult is told to commit nothing.
				return publicops.RemoveDependencyResult{Removed: false}, "", nil
			}
			// A wisp-plane removal touched no versioned table, so the commit this
			// asks for finds nothing to commit and RunTxResult absorbs that — the
			// same way the all-ephemeral add records no history entry.
			return publicops.RemoveDependencyResult{Removed: true},
				storageissueops.RemoveDependencyCommitMessage(request.IssueID, request.DependsOnID), nil
		})
	}
	if request.InTx == nil {
		return write()
	}
	// Two removals of one edge are identical diffs that Dolt lands with no conflict,
	// so a guarded removal queues, as a guarded delete does: the loser finds it gone.
	var result publicops.RemoveDependencyResult
	err := withEdgeFences(ctx, e.provider, []string{"edge:" + request.IssueID + ":" + request.DependsOnID}, func() error {
		var err error
		result, err = write()
		return err
	})
	if err != nil {
		return publicops.RemoveDependencyResult{}, err
	}
	return result, nil
}
