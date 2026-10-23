package issueops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/types"
	publicops "github.com/steveyegge/beads/issueops"
)

// WriteGraphDependencyInTx composes a checked graph Link edit through the native
// Dependency writer. The caller owns the transaction and graph identity guards.
// Native validation, journals and readiness run before one final source version.
// Recovery requires an absent pair; it cannot adopt a different live assertion.
func WriteGraphDependencyInTx(ctx context.Context, tx *sql.Tx, dep types.Dependency, actor string, recoverDeleted bool) error {
	if err := ValidateAddDependenciesRequest(publicops.AddDependenciesRequest{Actor: actor, Edges: []publicops.DependencyEdge{{IssueID: dep.IssueID, DependsOnID: dep.DependsOnID, Type: dep.Type}}}); err != nil {
		return err
	}
	var previousType, creator string
	var createdAt time.Time
	err := tx.QueryRowContext(ctx, `SELECT type,created_by,created_at FROM dependencies WHERE issue_id=? AND depends_on_issue_id=?`, dep.IssueID, dep.DependsOnID).Scan(&previousType, &creator, &createdAt)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if recoverDeleted && err == nil {
		return fmt.Errorf("%w: Dependency pair already has a live assertion", storage.ErrValidation)
	}
	if !recoverDeleted && errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w: checked Dependency disappeared", storage.ErrValidation)
	}
	if !recoverDeleted && previousType != string(dep.Type) {
		if _, err := removeDependencyInTx(ctx, tx, dep.IssueID, dep.DependsOnID, actor, true, nil, false); err != nil {
			return err
		}
	}
	if _, err := addDependencyInTx(ctx, tx, &dep, actor, AddDependencyOpts{SourceTable: "issues", TargetTable: "issues", WriteTable: "dependencies", EmitEvent: recoverDeleted || previousType != string(dep.Type)}, nil, false); err != nil {
		return err
	}
	if !recoverDeleted {
		// A live edit changes this assertion; native creation facts stay creation
		// facts even when changing kind required a native remove/add operation.
		if _, err := tx.ExecContext(ctx, `UPDATE dependencies SET created_by=?,created_at=? WHERE issue_id=? AND depends_on_issue_id=?`, creator, createdAt, dep.IssueID, dep.DependsOnID); err != nil {
			return err
		}
	}
	// The native same-type metadata re-add returns before readiness maintenance.
	// Metadata carries waits-for/conditional-blocks semantics, so recompute the
	// affected set after the final metadata, before capturing the source version.
	issues, wisps, err := AffectedByDepChangeInTx(ctx, tx, dep.IssueID, dep.DependsOnID, dep.Type)
	if err != nil {
		return err
	}
	if err := RecomputeIsBlockedInTx(ctx, tx, issues, wisps); err != nil {
		return err
	}
	return RecordVersionInTx(ctx, tx, dep.IssueID, actor)
}
