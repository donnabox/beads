package issueops

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
	"unicode/utf8"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/storage/depid"
	"github.com/steveyegge/beads/internal/storage/domain"
	"github.com/steveyegge/beads/internal/types"
	publicops "github.com/steveyegge/beads/issueops"
)

// canonicalDependencyInventory is a complete, bounded, durable local population.
// Catalog bijections and participation in retained history are a separate proof
// the graph publisher must supply. This module never infers them from ID spelling.
type canonicalDependencyInventory struct {
	templates map[string]bool
	rows      map[CanonicalDependencyPair]*types.Dependency
}

func readCanonicalDependencyInventory(ctx context.Context, tx *sql.Tx, population []string) (canonicalDependencyInventory, error) {
	inventory := canonicalDependencyInventory{templates: make(map[string]bool), rows: make(map[CanonicalDependencyPair]*types.Dependency)}
	rows, err := tx.QueryContext(ctx, `SELECT id, ephemeral, no_history, wisp_type, is_template FROM issues LIMIT 1001`)
	if err != nil {
		return inventory, err
	}
	for rows.Next() {
		var id string
		var ephemeral, noHistory, template sql.NullBool
		var wispType sql.NullString
		if err := rows.Scan(&id, &ephemeral, &noHistory, &wispType, &template); err != nil {
			_ = rows.Close()
			return inventory, err
		}
		if len(inventory.templates) >= CanonicalDependencyPlanLimit || id == "" || !utf8.ValidString(id) || ephemeral.Bool || noHistory.Bool || wispType.String != "" {
			_ = rows.Close()
			return inventory, fmt.Errorf("%w: canonical Dependency requires a bounded durable Issue population", storage.ErrValidation)
		}
		if _, duplicate := inventory.templates[id]; duplicate {
			_ = rows.Close()
			return inventory, fmt.Errorf("%w: duplicate native Issue identity", storage.ErrValidation)
		}
		inventory.templates[id] = template.Bool
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return inventory, err
	}
	if err := rows.Close(); err != nil {
		return inventory, err
	}
	if len(inventory.templates) != len(population) {
		return inventory, fmt.Errorf("%w: incomplete canonical native Issue population", storage.ErrValidation)
	}
	for _, id := range population {
		if _, ok := inventory.templates[id]; !ok {
			return inventory, fmt.Errorf("%w: canonical native Issue population changed", storage.ErrValidation)
		}
	}
	// Neither plane can hide authority from final graph validation/readiness.
	for _, query := range []string{`SELECT COUNT(*) FROM wisps`, `SELECT COUNT(*) FROM wisp_dependencies`} {
		var count int64
		if err := tx.QueryRowContext(ctx, query).Scan(&count); err != nil {
			return inventory, err
		}
		if count != 0 {
			return inventory, fmt.Errorf("%w: canonical Dependency candidate requires empty wisp authority", storage.ErrValidation)
		}
	}
	rows, err = tx.QueryContext(ctx, `SELECT id, issue_id, depends_on_issue_id, depends_on_wisp_id, depends_on_external, type, created_at, created_by, metadata, thread_id FROM dependencies LIMIT 1001`)
	if err != nil {
		return inventory, err
	}
	for rows.Next() {
		dep := new(types.Dependency)
		var target, wisp, external, actor, metadata, thread sql.NullString
		if err := rows.Scan(&dep.ID, &dep.IssueID, &target, &wisp, &external, &dep.Type, &dep.CreatedAt, &actor, &metadata, &thread); err != nil {
			_ = rows.Close()
			return inventory, err
		}
		dep.DependsOnID, dep.CreatedBy, dep.Metadata, dep.ThreadID = target.String, actor.String, metadata.String, thread.String
		if !metadata.Valid {
			dep.Metadata = "{}"
		}
		pair := CanonicalDependencyPair{SourceID: dep.IssueID, TargetID: dep.DependsOnID}
		_, sourceOK := inventory.templates[pair.SourceID]
		_, targetOK := inventory.templates[pair.TargetID]
		if len(inventory.rows) >= CanonicalDependencyPlanLimit || !target.Valid || wisp.Valid || external.Valid || !sourceOK || !targetOK || pair.SourceID == pair.TargetID || !dep.Type.IsWellKnown() || dep.ID != depid.New(pair.SourceID, pair.TargetID) || dep.ThreadID != "" || !utf8.ValidString(dep.CreatedBy) || dep.CreatedAt.IsZero() {
			_ = rows.Close()
			return inventory, fmt.Errorf("%w: unsupported or malformed canonical native Dependency baseline", storage.ErrValidation)
		}
		if err := validateCanonicalMetadata([]byte(dep.Metadata)); err != nil {
			_ = rows.Close()
			return inventory, err
		}
		if _, duplicate := inventory.rows[pair]; duplicate {
			_ = rows.Close()
			return inventory, fmt.Errorf("%w: ambiguous native Dependency pair", storage.ErrValidation)
		}
		inventory.rows[pair] = dep
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return inventory, err
	}
	return inventory, rows.Close()
}

func orderedCanonicalDependencyPairs(rows map[CanonicalDependencyPair]*types.Dependency) []CanonicalDependencyPair {
	pairs := make([]CanonicalDependencyPair, 0, len(rows))
	for pair := range rows {
		pairs = append(pairs, pair)
	}
	slices.SortFunc(pairs, func(a, b CanonicalDependencyPair) int {
		if n := graph.CompareCodeUnits(a.SourceID, b.SourceID); n != 0 {
			return n
		}
		return graph.CompareCodeUnits(a.TargetID, b.TargetID)
	})
	return pairs
}

func validateCanonicalDependencyGraph(ctx context.Context, tx *sql.Tx, rows map[CanonicalDependencyPair]*types.Dependency) error {
	adjacency := make(map[string][]string)
	var scheduling [][2]string
	pairs := orderedCanonicalDependencyPairs(rows)
	for _, pair := range pairs {
		if types.IsSchedulingEdge(rows[pair].Type) {
			adjacency[pair.SourceID] = append(adjacency[pair.SourceID], pair.TargetID)
			scheduling = append(scheduling, [2]string{pair.SourceID, pair.TargetID})
		}
	}
	if cycle := CycleThroughEdgesInGraph(adjacency, scheduling); cycle != "" {
		return fmt.Errorf("%w: %w: canonical native scheduling cycle: %s", storage.ErrValidation, domain.ErrDependencyCycle, cycle)
	}
	for _, pair := range pairs {
		if err := CheckBlockingHierarchyInTx(ctx, tx, rows[pair], []string{"dependencies"}); err != nil {
			return err
		}
	}
	return nil
}

func validateCanonicalDependencyBaseline(ctx context.Context, tx *sql.Tx, plan CanonicalDependencyIntentPlan, inventory canonicalDependencyInventory) error {
	finalCount := len(inventory.rows)
	for _, intent := range plan.Changes {
		if intent.Transition == CanonicalDependencyInsertTransition {
			finalCount++
		}
		if intent.Transition == CanonicalDependencyDeleteTransition {
			finalCount--
		}
		current := inventory.rows[intent.Pair]
		if intent.Existing == nil {
			if current != nil {
				return fmt.Errorf("%w: canonical Dependency insertion pair already exists", storage.ErrValidation)
			}
		} else if current == nil || current.ID != intent.Existing.NativeID || current.Type != intent.Type {
			return fmt.Errorf("%w: canonical Dependency baseline identity or classification changed", storage.ErrValidation)
		}
		if err := AuthorizeTemplateUpdate(&types.Issue{ID: intent.Pair.SourceID, IsTemplate: inventory.templates[intent.Pair.SourceID]}, publicops.UpdateRequest{}); err != nil {
			return err
		}
	}
	if finalCount > CanonicalDependencyPlanLimit {
		return fmt.Errorf("%w: final canonical Dependency population exceeds candidate bound", storage.ErrValidation)
	}
	if err := validateCanonicalDependencyGraph(ctx, tx, inventory.rows); err != nil {
		return err
	}
	count, err := CountIsBlockedInconsistenciesInTx(ctx, tx)
	if err != nil {
		return err
	}
	if count != 0 {
		return fmt.Errorf("%w: canonical Dependency baseline readiness is stale", storage.ErrValidation)
	}
	return nil
}

func cloneCanonicalDependency(dep *types.Dependency) *types.Dependency {
	if dep == nil {
		return nil
	}
	copy := *dep
	return &copy
}

func equalCanonicalDependency(a, b *types.Dependency) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return a.ID == b.ID && a.IssueID == b.IssueID && a.DependsOnID == b.DependsOnID && a.Type == b.Type && a.CreatedAt.Equal(b.CreatedAt) && a.CreatedBy == b.CreatedBy && a.ThreadID == b.ThreadID && DependencyMetadataEqual(a.Metadata, b.Metadata)
}
