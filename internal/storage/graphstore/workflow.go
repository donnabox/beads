package graphstore

import (
	"context"
	"database/sql"
	"fmt"
	"unicode/utf8"

	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/storage/issueops"
	"github.com/steveyegge/beads/internal/types"
	publicops "github.com/steveyegge/beads/issueops"
)

// CloseIssue delegates to the checked Issue-domain close, in the same fenced
// transaction as retained state and graph identity. It offers no force bypass.
func (s *Store) CloseIssue(ctx context.Context, path, reason, actor string) (IssueMutationResult, error) {
	if err := validatePath(path); err != nil {
		return IssueMutationResult{}, err
	}
	if actor == "" || !utf8.ValidString(actor) || !utf8.ValidString(reason) {
		return IssueMutationResult{}, fmt.Errorf("%w: close requires UTF-8 actor and reason", storage.ErrValidation)
	}
	var result IssueMutationResult
	err := s.withTx(ctx, true, func(tx *sql.Tx) error {
		if err := checkBinding(ctx, tx, s.options); err != nil {
			return err
		}
		before, err := s.requireIssueInTx(ctx, tx, path)
		if err != nil {
			return err
		}
		if before.Properties.Status == types.StatusClosed {
			result = IssueMutationResult{Issue: before}
			return nil
		}
		if err := s.touchCoordination(ctx, tx); err != nil {
			return err
		}
		unscope := issueops.ScopeVersionedHistoryTransaction(tx, true)
		defer unscope()
		closed, _, err := issueops.ExecuteClose(ctx, tx, publicops.CloseRequest{IssueID: before.Properties.ID, Reason: reason, Actor: actor})
		if err != nil {
			return err
		}
		if !closed.Changed {
			return fmt.Errorf("%w: close unexpectedly became a no-op", ErrInvalidStore)
		}
		if err := s.afterStage("close"); err != nil {
			return err
		}
		if err := s.recordIssueMappingInTx(ctx, tx, path, before.Properties.ID); err != nil {
			return err
		}
		after, err := s.showIssueInTx(ctx, tx, path)
		if err != nil {
			return err
		}
		result = IssueMutationResult{Issue: after, Changed: true}
		return nil
	})
	if err != nil {
		return IssueMutationResult{}, err
	}
	return result, nil
}

// ReadyIssues uses the existing scheduling query, then resolves every result
// into the same canonical graph. Readiness is derived, not retained Bead state.
func (s *Store) ReadyIssues(ctx context.Context) ([]IssueRecord, error) {
	result := []IssueRecord{}
	err := s.withTx(ctx, false, func(tx *sql.Tx) error {
		if err := checkBinding(ctx, tx, s.options); err != nil {
			return err
		}
		ready, err := issueops.GetReadyWorkInTx(ctx, tx, types.WorkFilter{})
		if err != nil {
			return err
		}
		for _, issue := range ready {
			var path string
			if err := tx.QueryRowContext(ctx, `SELECT path FROM graph_preview_catalog WHERE backing='issue' AND backing_key=? AND allocation_state='live'`, issue.ID).Scan(&path); err != nil {
				return fmt.Errorf("%w: unmapped ready Issue: %v", ErrInvalidStore, err)
			}
			record, err := s.showIssueInTx(ctx, tx, path)
			if err != nil {
				return err
			}
			result = append(result, record)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
