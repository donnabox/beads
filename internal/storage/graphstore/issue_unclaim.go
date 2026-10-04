package graphstore

import (
	"context"
	"database/sql"
	"fmt"
	"unicode/utf8"

	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/storage/issueops"
	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/internal/workapi"
	publicops "github.com/steveyegge/beads/issueops"
)

// UnclaimIssue releases only the current holder's in-progress graph Issue.
// Force, expected-holder supervision and lease-expiry recovery are separate
// policy decisions. The native release owns its lease, event and History stamp;
// the graph layer records only the complete retained projection in that tx.
func (s *Store) UnclaimIssue(ctx context.Context, path, actor string) (IssueMutationResult, error) {
	if err := validatePath(path); err != nil {
		return IssueMutationResult{}, fmt.Errorf("%w: %v", storage.ErrValidation, err)
	}
	if actor == "" || !utf8.ValidString(actor) {
		return IssueMutationResult{}, fmt.Errorf("%w: Issue unclaim requires a nonempty UTF-8 actor", storage.ErrValidation)
	}
	if err := types.CheckFieldLen("actor", actor); err != nil {
		return IssueMutationResult{}, fmt.Errorf("%w: %w", storage.ErrValidation, err)
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
		if before.Properties.Status != types.StatusInProgress {
			return fmt.Errorf("%w: graph unclaim requires an in-progress Issue", publicops.ErrNotReleasable)
		}
		request := publicops.ReleaseRequest{IssueID: before.Properties.ID, Actor: actor}
		if err := workapi.ValidateReleaseRequest(request); err != nil {
			return err
		}
		if err := s.touchCoordination(ctx, tx); err != nil {
			return err
		}
		unscope := issueops.ScopeVersionedHistoryTransaction(tx, true)
		defer unscope()
		released, _, err := issueops.ReleaseIssueInTx(ctx, tx, request)
		if err != nil {
			return err
		}
		if !released.Changed || released.Issue == nil || released.Issue.RowVersion == before.Properties.RowVersion {
			return fmt.Errorf("%w: Issue unclaim did not mint one native version", ErrInvalidStore)
		}
		if err := s.afterStage("issue-unclaim"); err != nil {
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
