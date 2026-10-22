package graphstore

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/storage/issueops"
	publicops "github.com/steveyegge/beads/issueops"
)

// CreateIssueBatch uses the native batch writer and publishes every graph
// identity and local dependency Link in that same transaction. A refused item
// cannot leave earlier headings or edges behind. A typed batch request may
// name an earlier item's explicit native ID; generated IDs cannot be named
// before creation. The current Markdown parser does not expose explicit IDs.
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
		batchTargetByItem := make([]map[int]int, len(request.Items))
		earlierIDs := make(map[string]int, len(request.Items))
		reverseBySource := make(map[string]int)
		for index, item := range request.Items {
			if issueops.ResolveInfraTypesInTx(ctx, tx)[string(item.Issue.IssueType.Normalize())] {
				return issueops.CreateBatchItemError(index, fmt.Errorf("%w: infrastructure Issue types are not admitted by the graph preview", storage.ErrValidation))
			}
			dependencies := make([]IssueCreateDependency, 0, len(item.Dependencies))
			local := make([]resolvedIssueCreateDependency, 0)
			localTarget := make(map[int]int)
			for _, dep := range item.Dependencies {
				if dep.Metadata != "" || dep.ThreadID != "" {
					return issueops.CreateBatchItemError(index, fmt.Errorf("%w: markdown Dependency metadata and threads are unavailable in graph mode", ErrCapabilityUnavailable))
				}
				if targetIndex, ok := earlierIDs[dep.TargetID]; ok {
					if dep.Reverse {
						return issueops.CreateBatchItemError(index, fmt.Errorf("%w: reverse Dependency to an earlier batch Issue is unavailable", ErrCapabilityUnavailable))
					}
					if err := s.requireRelationshipGenerationInTx(ctx, tx); err != nil {
						return issueops.CreateBatchItemError(index, err)
					}
					nativeType, err := previewDependencyType(dep.Type)
					if err != nil {
						return issueops.CreateBatchItemError(index, err)
					}
					typeURL, ok := nativeDependencyTypeURLForIssueType(s.ScopeURL(), issueType, nativeType)
					if !ok {
						return issueops.CreateBatchItemError(index, fmt.Errorf("%w: native Dependency Link Type is unavailable", ErrCapabilityUnavailable))
					}
					if err := s.selectedLinkTypeInTx(ctx, tx, typeURL); err != nil {
						return issueops.CreateBatchItemError(index, err)
					}
					localTarget[len(local)] = targetIndex
					local = append(local, resolvedIssueCreateDependency{targetID: dep.TargetID, typeURL: typeURL, nativeType: nativeType})
					continue
				}
				selector := IssueSelector{Path: "beads/" + dep.TargetID, NativeID: dep.TargetID}
				if strings.HasPrefix(dep.TargetID, "beads/") {
					selector = IssueSelector{Path: dep.TargetID}
				}
				dependencies = append(dependencies, IssueCreateDependency{Target: selector, Type: dep.Type, Reverse: dep.Reverse})
			}
			create := issueops.CreateBatchItemRequest(request, item)
			create.Dependencies = nil
			resolved, _, err := s.prepareIssueCreateDependenciesInTx(ctx, tx, issueType, create, dependencies)
			if err != nil {
				return issueops.CreateBatchItemError(index, err)
			}
			batchTargets := make(map[int]int, len(localTarget))
			for localIndex, targetIndex := range localTarget {
				batchTargets[len(resolved)+localIndex] = targetIndex
			}
			resolved = append(resolved, local...)
			if len(resolved) > PreviewOwnedLinkLimit {
				return issueops.CreateBatchItemError(index, fmt.Errorf("%w: preview source-owned Link limit %d exceeded", storage.ErrValidation, PreviewOwnedLinkLimit))
			}
			seen := make(map[string]bool, len(resolved))
			for _, dependency := range resolved {
				key := fmt.Sprintf("%t|%s", dependency.reverse, dependency.targetID)
				if seen[key] {
					return issueops.CreateBatchItemError(index, fmt.Errorf("%w: duplicate or conflicting create Dependency pair", storage.ErrValidation))
				}
				seen[key] = true
			}
			resolvedByItem[index] = resolved
			batchTargetByItem[index] = batchTargets
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
			for _, dep := range resolved {
				if !dep.reverse {
					request.Items[index].Dependencies = append(request.Items[index].Dependencies, publicops.CreateDependency{TargetID: dep.targetID, Type: dep.nativeType})
				}
			}
			if item.Issue.ID != "" {
				if _, exists := earlierIDs[item.Issue.ID]; exists {
					return issueops.CreateBatchItemError(index, fmt.Errorf("%w: duplicate explicit Issue ID in batch", storage.ErrValidation))
				}
				earlierIDs[item.Issue.ID] = index
			}
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
		records = make([]IssueRecord, 0, len(created.Issues))
		paths := make([]string, len(created.Issues))
		hasInitialNotes := false
		for index, issue := range created.Issues {
			if !sameIssueCreateFields(request.Items[index].Issue, issue, request.Actor) || issueops.IsWisp(issue) {
				return fmt.Errorf("%w: markdown item %d cannot be represented exactly as a durable graph Issue", storage.ErrValidation, index)
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
			for dependencyIndex, targetIndex := range batchTargetByItem[index] {
				resolvedByItem[index][dependencyIndex].targetPath = paths[targetIndex]
			}
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
