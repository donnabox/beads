package issueops

import (
	"context"
	"database/sql"
	"fmt"
	"slices"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/types"
)

// CanonicalDependencyRows is an opaque transaction-bound intermediate token.
// Only SettleCanonicalDependencyReadinessInTx may consume it. Any error requires
// rollback of the enclosing transaction; failed phases cannot be retried.
type CanonicalDependencyRows struct {
	tx       *sql.Tx
	plan     CanonicalDependencyIntentPlan
	baseline canonicalDependencyInventory
	blocked  blockedJournalSnapshot
	consumed bool
}

// CanonicalDependencyRowChange contains detached actual row pre/postimages.
// Reassert remains a changed assertion even if its field values happen to match.
type CanonicalDependencyRowChange struct {
	Transition    CanonicalDependencyTransition
	Pair          CanonicalDependencyPair
	Before, After *types.Dependency
}

// CanonicalDependencyFinalization is ready for the graph publisher to retain
// final Links and owners. Emission follows that publication in the same tx.
type CanonicalDependencyFinalization struct {
	tx         *sql.Tx
	actor      string
	population []string
	blocked    blockedJournalSnapshot
	changes    []CanonicalDependencyRowChange
	emitted    bool
}

// RowChanges returns copies; callers cannot alter the deferred facts.
func (f *CanonicalDependencyFinalization) RowChanges() []CanonicalDependencyRowChange {
	if f == nil {
		return nil
	}
	result := make([]CanonicalDependencyRowChange, len(f.changes))
	for i, c := range f.changes {
		result[i] = c
		result[i].Before = cloneCanonicalDependency(c.Before)
		result[i].After = cloneCanonicalDependency(c.After)
	}
	return result
}

// OwnerIDs excludes neighbors whose only change is derived readiness.
func (f *CanonicalDependencyFinalization) OwnerIDs() []string {
	if f == nil {
		return nil
	}
	owners := make(map[string]bool)
	for _, c := range f.changes {
		owners[c.Pair.SourceID] = true
	}
	result := make([]string, 0, len(owners))
	for id := range owners {
		result = append(result, id)
	}
	slices.SortFunc(result, graph.CompareCodeUnits)
	return result
}

// ApplyCanonicalDependencyFinalRowsInTx validates a complete local baseline,
// then applies final native pair transitions through private native writers.
// Intermediate graph checks, readiness, facts and versions are deferred as one
// fixed policy. The caller must hold its controlled-writer fence and separately
// prove complete canonical mappings/history. It must roll back on any error.
func ApplyCanonicalDependencyFinalRowsInTx(ctx context.Context, tx *sql.Tx, request CanonicalDependencyFinalRequest) (*CanonicalDependencyRows, error) {
	if tx == nil {
		return nil, fmt.Errorf("%w: canonical Dependency edit requires a transaction", storage.ErrValidation)
	}
	plan, err := NormalizeCanonicalDependencyFinalRequest(request)
	if err != nil {
		return nil, err
	}
	inventory, err := readCanonicalDependencyInventory(ctx, tx, plan.Population)
	if err != nil {
		return nil, err
	}
	if err := validateCanonicalDependencyBaseline(ctx, tx, plan, inventory); err != nil {
		return nil, err
	}
	before, err := captureBlockedJournalSnapshot(ctx, tx, plan.Population, nil)
	if err != nil {
		return nil, err
	}
	token := &CanonicalDependencyRows{tx: tx, plan: plan, baseline: inventory, blocked: before}
	// Remove all departing assertions before any new assertion, so no transient
	// pair/hierarchy conflict is mistaken for the validity of the final graph.
	for _, intent := range plan.Changes {
		if intent.Transition == CanonicalDependencyDeleteTransition || intent.Transition == CanonicalDependencyReassertTransition {
			if _, err := removeDependencyWithPolicyInTx(ctx, tx, intent.Pair.SourceID, intent.Pair.TargetID, plan.Actor, false, nil, false, dependencyCanonicalFinalRows); err != nil {
				return nil, err
			}
		}
	}
	kind := DepTargetIssue
	for _, intent := range plan.Changes {
		if intent.Transition == CanonicalDependencyDeleteTransition {
			continue
		}
		dep := &types.Dependency{IssueID: intent.Pair.SourceID, DependsOnID: intent.Pair.TargetID, Type: intent.Type, Metadata: string(intent.Metadata)}
		if _, err := addDependencyWithPolicyInTx(ctx, tx, dep, plan.Actor, AddDependencyOpts{SourceTable: "issues", TargetTable: "issues", WriteTable: "dependencies", TargetKind: &kind}, nil, false, dependencyCanonicalFinalRows); err != nil {
			return nil, err
		}
	}
	return token, nil
}

// SettleCanonicalDependencyReadinessInTx verifies exact final row intent,
// validates the whole final scheduling/hierarchy graph, then recomputes native
// readiness once over the complete population. It mints no history. A failed or
// successful attempt consumes rows; failure requires enclosing rollback.
func SettleCanonicalDependencyReadinessInTx(ctx context.Context, tx *sql.Tx, rows *CanonicalDependencyRows) (*CanonicalDependencyFinalization, error) {
	if tx == nil || rows == nil || rows.tx != tx || rows.consumed {
		return nil, fmt.Errorf("%w: invalid or consumed canonical Dependency rows", storage.ErrValidation)
	}
	rows.consumed = true
	final, err := readCanonicalDependencyInventory(ctx, tx, rows.plan.Population)
	if err != nil {
		return nil, err
	}
	changes, err := checkCanonicalDependencyFinalRows(rows.plan, rows.baseline.rows, final.rows)
	if err != nil {
		return nil, err
	}
	if err := validateCanonicalDependencyGraph(ctx, tx, final.rows); err != nil {
		return nil, err
	}
	deferred := context.WithValue(ctx, journalContextKey{}, false)
	if _, err := RecomputeIsBlockedInTxWithResult(deferred, tx, rows.plan.Population, nil); err != nil {
		return nil, err
	}
	return &CanonicalDependencyFinalization{tx: tx, actor: rows.plan.Actor, population: slices.Clone(rows.plan.Population), blocked: rows.blocked, changes: changes}, nil
}

func checkCanonicalDependencyFinalRows(plan CanonicalDependencyIntentPlan, before, after map[CanonicalDependencyPair]*types.Dependency) ([]CanonicalDependencyRowChange, error) {
	expected := make(map[CanonicalDependencyPair]bool, len(before)+len(plan.Changes))
	for pair := range before {
		expected[pair] = true
	}
	changed := make(map[CanonicalDependencyPair]bool)
	var result []CanonicalDependencyRowChange
	for _, intent := range plan.Changes {
		pair := intent.Pair
		a, b := before[pair], after[pair]
		changed[pair] = true
		if intent.Transition == CanonicalDependencyDeleteTransition {
			delete(expected, pair)
			if b != nil {
				return nil, fmt.Errorf("%w: canonical Dependency deletion not retained", storage.ErrValidation)
			}
		} else {
			expected[pair] = true
			if b == nil || b.Type != intent.Type || !DependencyMetadataEqual(b.Metadata, string(intent.Metadata)) {
				return nil, fmt.Errorf("%w: canonical Dependency final intent not retained", storage.ErrValidation)
			}
			if intent.Transition == CanonicalDependencyMetadataTransition {
				preserved := cloneCanonicalDependency(a)
				preserved.Metadata = b.Metadata
				if !equalCanonicalDependency(preserved, b) {
					return nil, fmt.Errorf("%w: canonical Dependency metadata edit changed assertion facts", storage.ErrValidation)
				}
				if equalCanonicalDependency(a, b) {
					continue
				}
			} else if b.CreatedBy != plan.Actor || b.ThreadID != "" {
				return nil, fmt.Errorf("%w: canonical Dependency creation facts not retained", storage.ErrValidation)
			}
		}
		result = append(result, CanonicalDependencyRowChange{Transition: intent.Transition, Pair: pair, Before: cloneCanonicalDependency(a), After: cloneCanonicalDependency(b)})
	}
	if len(expected) != len(after) {
		return nil, fmt.Errorf("%w: canonical Dependency final population changed", storage.ErrValidation)
	}
	for pair := range expected {
		if after[pair] == nil || (!changed[pair] && !equalCanonicalDependency(before[pair], after[pair])) {
			return nil, fmt.Errorf("%w: unplanned canonical Dependency row change", storage.ErrValidation)
		}
	}
	return result, nil
}

// EmitCanonicalDependencyFinalFactsInTx emits actual final native facts after
// all final native/Resource versions and owner snapshots exist in this tx.
// It never mints a version or allocates a public identity. On error, roll back;
// the token is consumed before emission so a partial emission is not retryable.
func EmitCanonicalDependencyFinalFactsInTx(ctx context.Context, tx *sql.Tx, final *CanonicalDependencyFinalization) error {
	if tx == nil || final == nil || final.tx != tx || final.emitted {
		return fmt.Errorf("%w: invalid or emitted canonical Dependency finalization", storage.ErrValidation)
	}
	final.emitted = true
	for _, c := range final.changes {
		if c.Transition == CanonicalDependencyDeleteTransition || c.Transition == CanonicalDependencyReassertTransition {
			if err := RecordEventInTable(ctx, tx, "events", c.Pair.SourceID, types.EventDependencyRemoved, final.actor, fmt.Sprintf("Removed dependency on %s", c.Pair.TargetID)); err != nil {
				return err
			}
			if err := RecordDepEventInTx(ctx, tx, EventDepRemove, c.Pair.SourceID, string(c.Before.Type), c.Pair.TargetID, c.Before.Metadata, final.actor); err != nil {
				return err
			}
		}
		if c.After != nil {
			if c.Transition != CanonicalDependencyMetadataTransition {
				if err := RecordEventInTable(ctx, tx, "events", c.Pair.SourceID, types.EventDependencyAdded, final.actor, fmt.Sprintf("Added dependency: %s %s %s", c.Pair.SourceID, c.After.Type, c.Pair.TargetID)); err != nil {
					return err
				}
			}
			if err := RecordDepEventInTx(ctx, tx, EventDepAdd, c.Pair.SourceID, string(c.After.Type), c.Pair.TargetID, c.After.Metadata, final.actor); err != nil {
				return err
			}
		}
	}
	return recordBlockedJournalChanges(ctx, tx, final.blocked, final.population, nil)
}
