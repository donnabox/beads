package graphstore

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/storage/issueops"
	"github.com/steveyegge/beads/internal/types"
	publicops "github.com/steveyegge/beads/issueops"
)

// IssueChildCreateRequest binds an ordinary child create to one checked graph
// parent. The native create writer allocates the dotted ID and parent-child
// Dependency; the graph writer publishes their Issue and Link identities in
// the same transaction. Optional local Issue dependencies use that transaction
// too; an explicit child ID is still not admitted.
type IssueChildCreateRequest struct {
	ParentPath             string
	ExpectedParentRevision string
	InheritLabels          bool
	Create                 publicops.CreateRequest
	Dependencies           []IssueCreateDependency
}

type IssueChildCreateResult struct {
	Child IssueRecord `json:"child"`
	Link  LinkRecord  `json:"link"`
}

func (s *Store) CreateChildIssue(ctx context.Context, request IssueChildCreateRequest) (IssueChildCreateResult, error) {
	if err := validatePath(request.ParentPath); err != nil {
		return IssueChildCreateResult{}, err
	}
	if request.ExpectedParentRevision == "" {
		return IssueChildCreateResult{}, fmt.Errorf("%w: child create requires the parent revision used for its read", storage.ErrValidation)
	}
	create := issueops.CloneCreateRequest(request.Create)
	if create.Issue == nil || create.Issue.ID != "" {
		return IssueChildCreateResult{}, fmt.Errorf("%w: child ID must be allocated by the native Issue writer", storage.ErrValidation)
	}
	if err := validateIssueCreate(create); err != nil {
		return IssueChildCreateResult{}, err
	}
	due, err := normalizeIssueDue(create.Issue.DueAt)
	if err != nil {
		return IssueChildCreateResult{}, err
	}
	create.Issue.DueAt = due
	deferUntil, err := normalizeIssueDefer(create.Issue.DeferUntil)
	if err != nil {
		return IssueChildCreateResult{}, err
	}
	create.Issue.DeferUntil = deferUntil
	var result IssueChildCreateResult
	err = s.withTx(ctx, true, func(tx *sql.Tx) error {
		if err := checkBinding(ctx, tx, s.options); err != nil {
			return err
		}
		if err := s.requireRelationshipGenerationInTx(ctx, tx); err != nil {
			return err
		}
		issueType, err := installedIssueTypeURLInTx(ctx, tx, s.ScopeURL())
		if err != nil {
			return err
		}
		linkType, ok := nativeDependencyTypeURLForIssueType(s.ScopeURL(), issueType, types.DepParentChild)
		if !ok {
			return fmt.Errorf("%w: installed Issue Type has no parent-child Link", ErrCapabilityUnavailable)
		}
		if err := s.selectedLinkTypeInTx(ctx, tx, linkType); err != nil {
			return err
		}
		parent, err := s.requireIssueInTx(ctx, tx, request.ParentPath)
		if err != nil {
			return err
		}
		if parent.Type != issueType {
			return fmt.Errorf("%w: parent is not an Issue of the installed relationship Type", storage.ErrValidation)
		}
		if parent.Revision != request.ExpectedParentRevision {
			return fmt.Errorf("%w: parent Issue revision changed (current %s)", ErrConflict, parent.Revision)
		}
		if issueops.ResolveInfraTypesInTx(ctx, tx)[string(create.Issue.IssueType.Normalize())] {
			return fmt.Errorf("%w: infrastructure Issue types are not admitted by the graph preview", storage.ErrValidation)
		}
		// The selected graph path can differ from the native backing ID. The
		// native writer must receive that backing ID so its counter and edge
		// refer to the same parent the graph read just checked.
		resolved, native, err := s.prepareIssueCreateDependenciesInTx(ctx, tx, issueType, create, request.Dependencies)
		if err != nil {
			return err
		}
		if len(native.Dependencies)+1 > PreviewOwnedLinkLimit {
			return fmt.Errorf("%w: preview source-owned Link limit %d exceeded", storage.ErrValidation, PreviewOwnedLinkLimit)
		}
		for _, dependency := range resolved {
			if !dependency.reverse && dependency.targetID == parent.Properties.ID {
				return fmt.Errorf("%w: parent is already linked by the native child create", storage.ErrValidation)
			}
		}
		native.ParentID = parent.Properties.ID
		native.InheritLabelsFromParent = request.InheritLabels
		if err := issueops.ValidatePublicCreateRequest(native); err != nil {
			return err
		}
		if err := s.touchCoordination(ctx, tx); err != nil {
			return err
		}
		unscope := issueops.ScopeVersionedHistoryTransaction(tx, true)
		defer unscope()
		created, _, err := issueops.ExecuteCreate(ctx, tx, native)
		if err != nil {
			return err
		}
		if !sameIssueCreateFields(create.Issue, created.Issue, create.Actor) || issueops.IsWisp(created.Issue) ||
			!strings.HasPrefix(created.Issue.ID, parent.Properties.ID+".") {
			return fmt.Errorf("%w: native child Issue cannot be represented exactly", storage.ErrValidation)
		}
		if err := s.afterStage("child-issue"); err != nil {
			return err
		}
		childPath, err := generatedIssuePath(ctx, tx, created.Issue.ID)
		if err != nil {
			return err
		}
		childRevision, err := freshToken()
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO graph_preview_catalog
			(path,resource_kind,type_url,revision,allocation_state,backing,backing_key)
			VALUES (?,'bead',?,?,'live','issue',?)`, childPath, issueType, childRevision, created.Issue.ID); err != nil {
			return err
		}
		if err := s.afterStage("child-catalog"); err != nil {
			return err
		}
		var dependencyID, dependencyType string
		if err := tx.QueryRowContext(ctx, `SELECT id,type FROM dependencies
			WHERE issue_id=? AND depends_on_issue_id=?`, created.Issue.ID, parent.Properties.ID).
			Scan(&dependencyID, &dependencyType); err != nil {
			return fmt.Errorf("%w: native child has no parent Dependency: %v", ErrInvalidStore, err)
		}
		if dependencyType != string(types.DepParentChild) {
			return fmt.Errorf("%w: native child has a non-parent Dependency", ErrInvalidStore)
		}
		linkToken, err := freshToken()
		if err != nil {
			return err
		}
		linkPath := "links/" + linkToken
		var allocated int
		allocationErr := tx.QueryRowContext(ctx, `SELECT 1 FROM graph_preview_catalog WHERE path=?`, linkPath).Scan(&allocated)
		if allocationErr == nil {
			return ErrAlreadyExists
		}
		if !errors.Is(allocationErr, sql.ErrNoRows) {
			return allocationErr
		}
		linkRevision, err := freshToken()
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO graph_preview_catalog
			(path,resource_kind,type_url,revision,allocation_state,backing,backing_key)
			VALUES (?,'link',?,?,'live','dependency',?)`, linkPath, linkType, linkRevision, dependencyID); err != nil {
			return err
		}
		if err := s.afterStage("child-link-catalog"); err != nil {
			return err
		}
		link, err := s.currentLinkInTx(ctx, tx, linkPath)
		if err != nil {
			return err
		}
		if link.Source != graph.CanonicalURL(s.ScopeURL(), childPath) || link.Target != parent.ID || link.Type != linkType {
			return fmt.Errorf("%w: child Link endpoints or Type differ from native Dependency", ErrInvalidStore)
		}
		linkBytes, err := canonicalJSON(link)
		if err != nil {
			return err
		}
		if err := insertPreviewVersionInTx(ctx, tx, linkPath, linkRevision, linkBytes, create.Actor); err != nil {
			return err
		}
		if err := s.afterStage("child-link-retained"); err != nil {
			return err
		}
		if len(resolved) != 0 {
			if err := s.mapIssueCreateDependenciesInTx(ctx, tx, childPath, created.Issue.ID, create.Actor, resolved, false); err != nil {
				return fmt.Errorf("map child Issue dependencies: %w", err)
			}
		}
		// The child has no prior graph mapping whose context could be preserved.
		if err := s.recordIssueMappingContextInTx(ctx, tx, childPath, created.Issue.ID, issueRevisionContext{Inception: "create"}); err != nil {
			return err
		}
		if err := s.afterStage("child-retained"); err != nil {
			return err
		}
		if create.Issue.Notes != "" {
			if err := checkCurrentReadBytes(ctx, tx); err != nil {
				return err
			}
		}
		child, err := s.showIssueInTx(ctx, tx, childPath)
		if err != nil {
			return err
		}
		foundParent := false
		for _, owned := range child.Owned {
			foundParent = foundParent || bytes.Equal(owned, linkBytes)
		}
		if len(child.Owned) != len(native.Dependencies)+1 || !foundParent {
			return fmt.Errorf("%w: child has no exact owned parent Link", ErrInvalidStore)
		}
		result = IssueChildCreateResult{Child: child, Link: link}
		return nil
	})
	if err != nil {
		return IssueChildCreateResult{}, err
	}
	return result, nil
}
