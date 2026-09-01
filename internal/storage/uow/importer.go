package uow

import (
	"context"
	"fmt"

	"github.com/steveyegge/beads/internal/storage"
	storageissueops "github.com/steveyegge/beads/internal/storage/issueops"
	"github.com/steveyegge/beads/internal/types"
	publicops "github.com/steveyegge/beads/issueops"
)

// ImporterSource is the capability accessor a unit-of-work provider offers
// for the import role, the sibling of BatchCloserSource and IssueReaderSource.
type ImporterSource interface {
	Importer() (publicops.Importer, error)
}

// Importer returns the guarded bulk-import surface for this provider.
func (p *doltSQLProvider) Importer() (publicops.Importer, error) {
	return NewImporter(p)
}

// NewImporter constructs a public importer backed by provider.
func NewImporter(provider UnitOfWorkProvider) (publicops.Importer, error) {
	if isNilUnitOfWorkProvider(provider) {
		return nil, fmt.Errorf("new importer: unit-of-work provider must not be nil")
	}
	return &importer{provider: provider}, nil
}

// importer lands one whole import batch through one unit of work.
type importer struct {
	provider UnitOfWorkProvider
}

var _ publicops.Importer = (*importer)(nil)

// ImportBatch writes the whole batch in ONE unit of work and commits it as
// ONE history entry: the issue rows through the SAME batch-upsert engine the
// classic stores run (internal/storage/issueops.CreateIssuesInTxWithResult —
// conditional row upsert, idempotent label/comment/dependency merge, child
// counters, blocked recompute), the memory records, and the optional
// issue_prefix reconciliation. A request-level failure rolls all of it back.
//
// The engine's callbacks land in the result instead of being exposed on the
// request: RunTxResult retries the whole attempt on a serialization failure,
// and result state declared inside the attempt cannot leak between retries
// the way a caller's callback accumulator would.
func (o *importer) ImportBatch(ctx context.Context, request publicops.ImportBatchRequest) (publicops.ImportBatchResult, error) {
	if request.Actor == "" {
		return publicops.ImportBatchResult{}, fmt.Errorf("import batch: actor must not be empty")
	}
	return RunTxResult(ctx, o.provider, func(ctx context.Context, uw UnitOfWork) (publicops.ImportBatchResult, string, error) {
		var result publicops.ImportBatchResult

		if len(request.Issues) > 0 {
			runner, err := importStatementRunner(uw)
			if err != nil {
				return publicops.ImportBatchResult{}, "", err
			}
			staleRejected := make(map[string]struct{})
			skippedSeen := make(map[string]struct{})
			opts := storage.BatchCreateOptions{
				SkipPrefixValidation:           request.SkipPrefixValidation,
				RejectStaleUpserts:             !request.AllowStale,
				SkipDependencyValidationErrors: true,
				OnSkippedDependency: func(issueID, dependsOnID, reason string) {
					key := issueID + "\x00" + dependsOnID + "\x00" + reason
					if _, ok := skippedSeen[key]; ok {
						return
					}
					skippedSeen[key] = struct{}{}
					result.SkippedDependencies = append(result.SkippedDependencies, publicops.SkippedDependency{
						IssueID:     issueID,
						DependsOnID: dependsOnID,
						Reason:      reason,
					})
				},
				OnStaleRejected: func(issueID string) {
					if _, ok := staleRejected[issueID]; ok {
						return
					}
					staleRejected[issueID] = struct{}{}
					result.StaleRejectedIDs = append(result.StaleRejectedIDs, issueID)
				},
			}
			// TWO PASSES OVER ONE TRANSACTION, and only when the batch
			// actually carries a regular<->wisp edge onto another of its own
			// rows: the engine skip-reports such an edge while both endpoints
			// are rows of one batch, so writing the rows first and those edges
			// after is what wires them (wy-zdfs6r; the classic chunked import
			// does the same across its chunks, wy-4276q8). It is still ONE
			// unit of work and ONE history entry — the passes share this
			// transaction, so the second one's targets are rows this same
			// transaction already wrote, and a failure in either rolls the
			// whole import back.
			rows, deferredPlanes := storageissueops.SplitCrossPlaneBatchEdges(request.Issues)
			if _, err := storageissueops.CreateIssuesInTxWithResult(ctx, runner, rows, request.Actor, opts); err != nil {
				return publicops.ImportBatchResult{}, "", err
			}
			// Created counts the ROWS the batch wrote, so it is computed from
			// the request before the edge pass, which writes none.
			result.Created = len(request.Issues) - len(staleRejected)
			if err := wireDeferredCrossPlaneEdges(ctx, runner, deferredPlanes, staleRejected, request.Actor, opts); err != nil {
				return publicops.ImportBatchResult{}, "", err
			}
		}

		for _, memory := range request.Memories {
			if err := uw.ConfigUseCase().SetConfig(ctx, memory.Key, memory.Value); err != nil {
				return publicops.ImportBatchResult{}, "", fmt.Errorf("import memory %q: %w", memory.Key, err)
			}
			result.MemoriesImported++
		}

		// config.yaml is authoritative for issue_prefix on the import flow
		// (be-llaf); a read or write failure here degrades to "not synced"
		// rather than failing the batch, matching the classic path.
		if request.SyncIssuePrefix != "" {
			stored, _ := uw.ConfigUseCase().GetConfig(ctx, "issue_prefix")
			if stored != request.SyncIssuePrefix {
				if err := uw.ConfigUseCase().SetConfig(ctx, "issue_prefix", request.SyncIssuePrefix); err == nil {
					result.PrefixSynced = true
				}
			}
		}

		return result, importBatchCommitMessage(request, result), nil
	})
}

// wireDeferredCrossPlaneEdges applies the regular<->wisp edges
// SplitCrossPlaneBatchEdges held back, one single-plane batch at a time, now
// that every target is a row this transaction has already written.
//
// ConflictSkip because the row write already happened in the first pass and a
// second full upsert here would rewrite it; with it the engine leaves the
// stored row untouched and still wires the batch's dependencies. The
// stale-rejection callback is cleared for the same reason no row write here
// can be rejected: a second signal could only misreport a row whose first
// pass landed. A row the stale guard DID reject keeps its deferred edges out
// too — its snapshot is the older version of the bead, and merging its aux
// data is the loss bd-578h9.8 closed.
//
// Every edge dropped for an ordinary reason (an absent target, a cycle) is
// still reported: opts keeps the caller's OnSkippedDependency, so this pass
// reports exactly what the one-pass write would have.
func wireDeferredCrossPlaneEdges(ctx context.Context, runner storageissueops.DBTX, planes [][]*types.Issue, staleRejected map[string]struct{}, actor string, opts storage.BatchCreateOptions) error {
	if len(planes) == 0 {
		return nil
	}
	depOpts := opts
	depOpts.ConflictSkip = true
	depOpts.OnStaleRejected = nil
	for _, plane := range planes {
		rows := plane
		if len(staleRejected) > 0 {
			rows = plane[:0:0]
			for _, row := range plane {
				if _, stale := staleRejected[row.ID]; stale {
					continue
				}
				rows = append(rows, row)
			}
		}
		if len(rows) == 0 {
			continue
		}
		if _, err := storageissueops.CreateIssuesInTxWithResult(ctx, runner, rows, actor, depOpts); err != nil {
			return err
		}
	}
	return nil
}

// importBatchCommitMessage names what LANDED, in the exact shape the classic
// `bd import` has always committed. A batch that landed no issue rows and no
// memories but did reconcile the prefix commits under the sync message the
// classic path used for its separate config commit; a batch that landed
// nothing at all returns "" and commits nothing.
func importBatchCommitMessage(request publicops.ImportBatchRequest, result publicops.ImportBatchResult) string {
	if result.Created > 0 || result.MemoriesImported > 0 {
		msg := fmt.Sprintf("bd import: %d issues", result.Created)
		if result.MemoriesImported > 0 {
			msg += fmt.Sprintf(", %d memories", result.MemoriesImported)
		}
		return msg + fmt.Sprintf(" from %s", request.Source)
	}
	if result.PrefixSynced {
		return "bd import: sync issue_prefix from config.yaml"
	}
	return ""
}

// importStatementRunner exposes the unit of work's statement runner to the
// batch-upsert engine. The import role is deliberately NOT a use-case loop:
// upsert semantics, aux-data merge, child counters and the blocked recompute
// already live in internal/storage/issueops behind the DBTX seam, and the
// domain/db Runner satisfies that seam, so the proxied path runs the SAME
// code the classic stores run instead of mirroring it in parallel SQL.
//
// The unit of work is UNWRAPPED first because a decorated one is still the
// caller's unit of work: the hook-notifying wrapper stands in front of the use
// cases (notifying.go), and asserting on the concrete type through it would
// fail every import in a workspace that has hooks. Reaching the runner past a
// decorator is exactly what unwrapUOW is for, and it is why an import fires no
// hooks on this plumbing — the statements never pass a use case for the
// recorder to see. Same on the DoltStorage plumbing, where import runs through
// the batch engine rather than the hook-firing store methods.
func importStatementRunner(uw UnitOfWork) (storageissueops.DBTX, error) {
	b, ok := unwrapUOW(uw).(*baseUOW)
	if !ok {
		return nil, fmt.Errorf("uow: importer: %T does not expose a statement runner", uw)
	}
	return b.tx.Runner(), nil
}
