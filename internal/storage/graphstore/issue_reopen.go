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

// ReopenIssue uses the existing Issue-domain done-category transition and keeps
// its retained state and complete owned graph mapping in the same transaction.
// ReopenIssue keeps the unguarded form for existing callers.
func (s *Store) ReopenIssue(ctx context.Context, path, reason, actor string) (IssueMutationResult, error) {
	return s.ReopenIssueWithOptions(ctx, IssueReopenRequest{Path: path, Reason: reason, Actor: actor})
}

// IssueReopenRequest optionally guards the complete Issue and owned-Link
// projection, including an already-open no-op.
type IssueReopenRequest struct {
	Path, Reason, Actor string
	ExpectedRevision    string
}

func (s *Store) ReopenIssueWithOptions(ctx context.Context, request IssueReopenRequest) (IssueMutationResult, error) {
	if err := validatePath(request.Path); err != nil {
		return IssueMutationResult{}, fmt.Errorf("%w: %v", storage.ErrValidation, err)
	}
	if request.Actor == "" || !utf8.ValidString(request.Actor) || !utf8.ValidString(request.Reason) {
		return IssueMutationResult{}, fmt.Errorf("%w: reopen requires a nonempty UTF-8 actor and UTF-8 reason", storage.ErrValidation)
	}
	var result IssueMutationResult
	err := s.withNativeDomainTx(ctx, txNativeIssueMutation, false, func(ctx context.Context, tx *sql.Tx) error {
		if err := checkBinding(ctx, tx, s.options); err != nil {
			return err
		}
		before, err := s.requireIssueInTx(ctx, tx, request.Path)
		if err != nil {
			return err
		}
		if err := checkRevisionGuard(request.ExpectedRevision, false, before.Revision, false, "Issue"); err != nil {
			return err
		}
		category, err := issueops.ReopenCategoryInTx(ctx, tx, before.Properties.Status)
		if err != nil {
			return err
		}
		if category != types.CategoryDone {
			result = IssueMutationResult{Issue: before}
			return nil
		}
		if err := s.touchCoordination(ctx, tx); err != nil {
			return err
		}
		unscope := issueops.ScopeVersionedHistoryTransaction(tx, true)
		defer unscope()
		reopenRequest := publicops.ReopenRequest{IssueID: before.Properties.ID, Reason: request.Reason, Actor: request.Actor}
		if request.ExpectedRevision != "" {
			reopenRequest.ExpectedVersion = &before.Properties.RowVersion
		}
		reopened, _, err := issueops.ExecuteReopen(ctx, tx, reopenRequest)
		if err != nil {
			return err
		}
		if !reopened.Changed {
			return fmt.Errorf("%w: reopen unexpectedly became a no-op", ErrInvalidStore)
		}
		if err := s.afterStage("issue-reopen"); err != nil {
			return err
		}
		// The target shared writer owns the one native retained version.
		if err := s.afterStage("issue-retained"); err != nil {
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
