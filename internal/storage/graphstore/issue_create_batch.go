package graphstore

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/storage/issueops"
	publicops "github.com/steveyegge/beads/issueops"
)

// CreateIssueBatch uses the native batch writer and publishes every graph
// identity and local dependency Link in that same transaction. A refused item
// cannot leave earlier headings or edges behind. Dependency targets must
// already exist; markdown headings have no caller-assigned IDs to refer to one
// another within the batch.
func (s *Store) CreateIssueBatch(ctx context.Context, request publicops.CreateBatchRequest) ([]IssueRecord, error) {
	request = issueops.CloneCreateBatchRequest(request)
	if err := issueops.ValidateCreateBatchRequest(request); err != nil {
		return nil, err
	}
	for index, item := range request.Items {
		// The graph adapter validates and resolves the relationships in its
		// transaction below. validateIssueCreate only checks Issue fields.
		fieldsOnly := item
		fieldsOnly.Dependencies = nil
		if err := validateIssueCreate(issueops.CreateBatchItemRequest(request, fieldsOnly)); err != nil {
			return nil, issueops.CreateBatchItemError(index, err)
		}
		due, err := normalizeIssueDue(item.Issue.DueAt)
		if err != nil {
			return nil, issueops.CreateBatchItemError(index, err)
		}
		item.Issue.DueAt = due
		request.Items[index] = item
	}
	var records []IssueRecord
	err := s.withTx(ctx, true, func(tx *sql.Tx) error {
		if err := checkBinding(ctx, tx, s.options); err != nil {
			return err
		}
		issueType, err := installedIssueTypeURLInTx(ctx, tx, s.ScopeURL())
		if err != nil {
			return err
		}
		resolvedByItem := make([][]resolvedIssueCreateDependency, len(request.Items))
		reverseBySource := make(map[string]int)
		for index, item := range request.Items {
			if issueops.ResolveInfraTypesInTx(ctx, tx)[string(item.Issue.IssueType.Normalize())] {
				return issueops.CreateBatchItemError(index, fmt.Errorf("%w: infrastructure Issue types are not admitted by the graph preview", storage.ErrValidation))
			}
			// Explicit IDs may remain reserved by another Resource or a deleted
			// allocation after the native Issue row is gone. Check before the
			// native batch writer can reuse retained History for that spelling.
			if item.Issue.ID != "" {
				if exists, err := publicResourceIDExistsInTx(ctx, tx, item.Issue.ID); err != nil {
					return issueops.CreateBatchItemError(index, err)
				} else if exists {
					return issueops.CreateBatchItemError(index, ErrAlreadyExists)
				}
			}
			dependencies := make([]IssueCreateDependency, 0, len(item.Dependencies))
			for _, dep := range item.Dependencies {
				if dep.Metadata != "" || dep.ThreadID != "" {
					return issueops.CreateBatchItemError(index, fmt.Errorf("%w: markdown Dependency metadata and threads are unavailable in graph mode", ErrCapabilityUnavailable))
				}
				targetID := dep.TargetID
				if strings.HasPrefix(targetID, s.ScopeURL()) {
					var err error
					targetID, err = graph.ScopeLocalPublicResourceID(s.ScopeURL(), targetID)
					if err != nil {
						return issueops.CreateBatchItemError(index, err)
					}
				}
				if err := graph.ValidateCanonicalResourceID(targetID); err != nil {
					return issueops.CreateBatchItemError(index, err)
				}
				// Dependency targets are opaque public IDs, even when the
				// spelling happens to begin with beads/.
				selector := IssueSelector{Path: "beads/" + targetID, NativeID: targetID, Exact: true}
				dependencies = append(dependencies, IssueCreateDependency{Target: selector, Type: dep.Type, Reverse: dep.Reverse})
			}
			create := issueops.CreateBatchItemRequest(request, item)
			create.Dependencies = nil
			resolved, _, err := s.prepareIssueCreateDependenciesInTx(ctx, tx, issueType, create, dependencies)
			if err != nil {
				return issueops.CreateBatchItemError(index, err)
			}
			resolvedByItem[index] = resolved
			for _, dep := range resolved {
				if !dep.reverse {
					continue
				}
				source, err := s.requireIssueInTx(ctx, tx, dep.targetPath)
				if err != nil {
					return issueops.CreateBatchItemError(index, err)
				}
				reverseBySource[dep.targetPath]++
				if len(source.Owned)+reverseBySource[dep.targetPath] > PreviewOwnedLinkLimit {
					return issueops.CreateBatchItemError(index, fmt.Errorf("%w: preview source-owned Link limit %d exceeded", storage.ErrValidation, PreviewOwnedLinkLimit))
				}
			}
			request.Items[index].Dependencies = nil
		}
		if err := s.touchCoordination(ctx, tx); err != nil {
			return err
		}
		unscope := issueops.ScopeVersionedHistoryTransaction(tx, true)
		defer unscope()
		created, _, err := issueops.ExecuteCreateBatch(ctx, tx, request)
		if err != nil {
			return err
		}
		if err := s.afterStage("issue"); err != nil {
			return err
		}
		// Reserve every prospective Type before checking any new Resource ID.
		// Item order cannot let an earlier Resource claim a later item's Type.
		declaredTypes := make([]string, len(created.Issues))
		records = make([]IssueRecord, 0, len(created.Issues))
		paths := make([]string, len(created.Issues))
		hasInitialNotes := false
		for index, issue := range created.Issues {
			if !sameIssueCreateFields(request.Items[index].Issue, issue, request.Actor) || issueops.IsWisp(issue) {
				return fmt.Errorf("%w: markdown item %d cannot be represented exactly as a durable graph Issue", storage.ErrValidation, index)
			}
			declaredTypes[index], err = NativeIssueTypeURL(s.ScopeURL(), string(issue.IssueType))
			if err != nil {
				return err
			}
			if err := s.ensureNativeTypeInTx(ctx, tx, declaredTypes[index]); err != nil {
				return err
			}
		}
		// Types have all been declared before their corresponding public IDs
		// are tested against the lifetime allocation set.
		for index, issue := range created.Issues {
			// Generated IDs are known only after the native batch writer runs.
			// Roll every item back if any public ID is already allocated.
			if exists, err := publicResourceIDExistsInTx(ctx, tx, issue.ID); err != nil {
				return issueops.CreateBatchItemError(index, err)
			} else if exists {
				return issueops.CreateBatchItemError(index, ErrAlreadyExists)
			}
			path, err := generatedIssuePath(ctx, tx, issue.ID)
			if err != nil {
				return err
			}
			revision, err := freshToken()
			if err != nil {
				return err
			}
			var ordinal int64
			if err := tx.QueryRowContext(ctx, `SELECT current_revision FROM issues WHERE id=?`, issue.ID).Scan(&ordinal); err != nil {
				return err
			}
			issueType := declaredTypes[index]
			if _, err := tx.ExecContext(ctx, `INSERT INTO graph_preview_catalog
				(path,resource_kind,type_url,revision,allocation_state,backing,backing_key)
				VALUES (?,'bead',?,?,'live','issue',?)`, path, issueType, revision, issue.ID); err != nil {
				return err
			}
			if err := s.afterStage("issue-catalog"); err != nil {
				return err
			}
			paths[index] = path
			if len(resolvedByItem[index]) == 0 {
				if _, err := tx.ExecContext(ctx, `INSERT INTO graph_preview_issue_versions
					(path,version,issue_id,issue_revision,owned,revision_context,history_ordinal) VALUES (?,?,?,?,'[]','{"inception":"create"}',?)`, path, revision, issue.ID, ordinal, ordinal); err != nil {
					return err
				}
			}
			if err := s.afterStage("retained"); err != nil {
				return err
			}
		}
		for index, issue := range created.Issues {
			path := paths[index]
			if len(resolvedByItem[index]) != 0 {
				if err := s.mapIssueCreateDependenciesInTx(ctx, tx, path, issue.ID, request.Actor, resolvedByItem[index], true); err != nil {
					return issueops.CreateBatchItemError(index, err)
				}
			}
			record, err := s.showIssueInTx(ctx, tx, path)
			if err != nil {
				return err
			}
			records = append(records, record)
			hasInitialNotes = hasInitialNotes || issue.Notes != ""
		}
		if hasInitialNotes {
			if err := checkCurrentReadBytes(ctx, tx); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return records, nil
}
