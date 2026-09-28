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

// UpdateIssueRequest admits assignee, priority, title, description, design and acceptance criteria.
// AppendNotes reuses the native transactional append operation. Notes replacement
// and clearing remain reserved for contributor safeguard reconciliation. A nil field leaves that property unchanged; an explicit empty
// string clears it where the Issue domain permits. An explicit priority of zero
// sets P0; a nil priority leaves it unchanged. The guard addresses the complete graph revision, including
// the Issue's owned blocking Dependencies, not its private storage ordinal.
// Assignee nil preserves the current value; empty clears it. Neither the graph
// guard nor Unconditional bypasses the ordinary active-assignment transfer fence.
type UpdateIssueRequest struct {
	Path, Actor, ExpectedRevision                  string
	Unconditional                                  bool
	Title, Description, Design, AcceptanceCriteria *string
	Priority                                       *int
	Assignee                                       *string
	AppendNotes                                    *string
}

// UpdateIssue delegates admitted scalar edits to the existing Issue domain writer and
// retained-version recorder. Payload, opaque graph revision and complete owned
// state commit together; this adds no public History ordering or change context.
func (s *Store) UpdateIssue(ctx context.Context, request UpdateIssueRequest) (IssueMutationResult, error) {
	if err := validatePath(request.Path); err != nil {
		return IssueMutationResult{}, fmt.Errorf("%w: %v", storage.ErrValidation, err)
	}
	if request.Actor == "" || !utf8.ValidString(request.Actor) {
		return IssueMutationResult{}, fmt.Errorf("%w: Issue update requires a nonempty UTF-8 actor", storage.ErrValidation)
	}
	patch := publicops.IssuePatch{}
	count := 0
	for _, field := range []struct {
		name  string
		value *string
		dest  *publicops.Field[string]
	}{
		{"title", request.Title, &patch.Title},
		{"assignee", request.Assignee, &patch.Assignee},
		{"description", request.Description, &patch.Description},
		{"design", request.Design, &patch.Design},
		{"acceptance_criteria", request.AcceptanceCriteria, &patch.AcceptanceCriteria},
		{"append_notes", request.AppendNotes, &patch.AppendNotes},
	} {
		if field.value == nil {
			continue
		}
		if !utf8.ValidString(*field.value) {
			return IssueMutationResult{}, fmt.Errorf("%w: Issue %s must be UTF-8", storage.ErrValidation, field.name)
		}
		*field.dest = publicops.Field[string]{Set: true, Value: *field.value}
		count++
	}
	if patch.Assignee.Set {
		if err := types.CheckFieldLen("assignee", patch.Assignee.Value); err != nil {
			return IssueMutationResult{}, fmt.Errorf("%w: Issue assignee: %w", storage.ErrValidation, err)
		}
	}
	if request.Priority != nil {
		patch.Priority = publicops.Field[int]{Set: true, Value: *request.Priority}
		count++
	}
	if count == 0 {
		return IssueMutationResult{}, fmt.Errorf("%w: Issue update requires an admitted field", storage.ErrValidation)
	}
	attempt := publicops.UpdateRequest{Actor: request.Actor, Patch: patch, IssuePlaneOnly: true}
	if err := issueops.ValidateUpdateRequest(attempt); err != nil {
		return IssueMutationResult{}, err
	}
	var result IssueMutationResult
	err := s.withTx(ctx, true, func(tx *sql.Tx) error {
		if err := checkBinding(ctx, tx, s.options); err != nil {
			return err
		}
		current, revision, _, err := s.beadEndpointInTx(ctx, tx, request.Path)
		if err != nil {
			return err
		}
		before, ok := current.(IssueRecord)
		if !ok {
			return fmt.Errorf("%w: Issue update requires the experimental Issue Type", ErrCapabilityUnavailable)
		}
		if err := checkRevisionGuard(request.ExpectedRevision, request.Unconditional, revision, true, "Issue"); err != nil {
			return err
		}
		// Resolve append intent only to plan the no-op against this checked snapshot.
		// ExecuteUpdate must still receive the original typed intent so its shared
		// writer owns atomic read/append semantics and future overwrite safeguards.
		updates, err := issueops.ResolveMergeOps(before.Properties, issueops.UpdateFields(patch))
		if err != nil {
			return err
		}
		updates, err = issueops.DiscardNoopIssueUpdates(before.Properties, updates)
		if err != nil {
			return err
		}
		if len(updates) == 0 {
			result = IssueMutationResult{Issue: before}
			return nil
		}
		if err := s.touchCoordination(ctx, tx); err != nil {
			return err
		}
		unscope := issueops.ScopeVersionedHistoryTransaction(tx, true)
		defer unscope()
		attempt.IssueID = before.Properties.ID
		updated, _, err := issueops.ExecuteUpdate(ctx, tx, attempt)
		if err != nil {
			return err
		}
		if !updated.Changed {
			return fmt.Errorf("%w: Issue update unexpectedly became a no-op", ErrInvalidStore)
		}
		if err := s.afterStage("issue-update"); err != nil {
			return err
		}
		// ExecuteUpdate records its ordinary domain event, but does not invoke the
		// retained Issue recorder. Use Jim's existing recorder once for this edit.
		if err := issueops.RecordVersionInTx(ctx, tx, before.Properties.ID, request.Actor); err != nil {
			return err
		}
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
