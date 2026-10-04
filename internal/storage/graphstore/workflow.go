package graphstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
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

// ErrIssueDeferralConstraint refuses an Issue whose assignment or lifecycle
// state has no settled dateless deferral policy in the graph preview.
var ErrIssueDeferralConstraint = errors.New("Issue state cannot be deferred or undeferred")

// IssueDeferralRequest changes one durable Issue's dateless icebox state. The
// graph revision guard is checked against the complete owned graph before the
// native Issue writer records the sole retained successor.
type IssueDeferralRequest struct {
	Path, Actor, ExpectedRevision string
	Unconditional, Deferred       bool
}

// SetIssueDeferred uses the existing native Issue update funnel. In particular,
// the native writer owns its History stamp; graphstore only records the resulting
// projection in the same transaction. No clock or automatic wake-up is added.
func (s *Store) SetIssueDeferred(ctx context.Context, request IssueDeferralRequest) (IssueMutationResult, error) {
	if err := validatePath(request.Path); err != nil {
		return IssueMutationResult{}, fmt.Errorf("%w: %v", storage.ErrValidation, err)
	}
	if request.Actor == "" || !utf8.ValidString(request.Actor) {
		return IssueMutationResult{}, fmt.Errorf("%w: deferral requires a nonempty UTF-8 actor", storage.ErrValidation)
	}
	var result IssueMutationResult
	err := s.withTx(ctx, true, func(tx *sql.Tx) error {
		if err := checkBinding(ctx, tx, s.options); err != nil {
			return err
		}
		before, err := s.requireIssueInTx(ctx, tx, request.Path)
		if err != nil {
			return err
		}
		if err := checkRevisionGuard(request.ExpectedRevision, request.Unconditional, before.Revision, true, "Issue"); err != nil {
			return err
		}
		if (before.Properties.Status != types.StatusOpen && before.Properties.Status != types.StatusDeferred) || before.Properties.Assignee != "" {
			return fmt.Errorf("%w: dateless deferral requires an unassigned open or deferred Issue; claimed, in-progress, closed and pinned Issues need an explicit release or lifecycle decision", ErrIssueDeferralConstraint)
		}
		if request.Deferred == (before.Properties.Status == types.StatusDeferred) {
			result = IssueMutationResult{Issue: before}
			return nil
		}
		status := types.StatusOpen
		if request.Deferred {
			status = types.StatusDeferred
		}
		patch := publicops.IssuePatch{
			Status:     publicops.Field[types.Status]{Set: true, Value: status},
			DeferUntil: publicops.Field[*time.Time]{Set: true},
		}
		if err := s.touchCoordination(ctx, tx); err != nil {
			return err
		}
		unscope := issueops.ScopeVersionedHistoryTransaction(tx, true)
		defer unscope()
		updated, _, err := issueops.ExecuteUpdate(ctx, tx, publicops.UpdateRequest{
			IssueID: before.Properties.ID, Actor: request.Actor, Patch: patch, IssuePlaneOnly: true,
		})
		if err != nil {
			return err
		}
		if !updated.Changed {
			return fmt.Errorf("%w: deferral unexpectedly became a no-op", ErrInvalidStore)
		}
		if err := s.afterStage("issue-deferral"); err != nil {
			return err
		}
		if err := s.recordIssueMappingInTx(ctx, tx, request.Path, before.Properties.ID); err != nil {
			return err
		}
		after, err := s.showIssueInTx(ctx, tx, request.Path)
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
