package graphstore

import (
	"context"
	"database/sql"
	"fmt"
	"unicode/utf8"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/storage/issueops"
	"github.com/steveyegge/beads/internal/types"
	publicops "github.com/steveyegge/beads/issueops"
)

// IssueCreateDependency selects an existing local Issue. Reverse corresponds
// to shipping create --deps blocks:TARGET: TARGET depends on the new Issue.
// External endpoints are deliberately outside this graph Link contract.
type IssueCreateDependency struct {
	Target  IssueSelector
	Type    types.DependencyType
	Reverse bool
}

type resolvedIssueCreateDependency struct {
	targetPath string
	targetID   string
	typeURL    string
	nativeType types.DependencyType
	reverse    bool
}

func (s *Store) prepareIssueCreateDependenciesInTx(ctx context.Context, tx *sql.Tx, issueType string, request publicops.CreateRequest, dependencies []IssueCreateDependency) ([]resolvedIssueCreateDependency, publicops.CreateRequest, error) {
	native := issueops.CloneCreateRequest(request)
	if len(dependencies) == 0 {
		return nil, native, nil
	}
	if err := s.requireRelationshipGenerationInTx(ctx, tx); err != nil {
		return nil, publicops.CreateRequest{}, err
	}
	if request.Actor == "" || !utf8.ValidString(request.Actor) {
		return nil, publicops.CreateRequest{}, fmt.Errorf("%w: dependency create requires an actor", storage.ErrValidation)
	}
	if len(dependencies) > PreviewOwnedLinkLimit {
		return nil, publicops.CreateRequest{}, fmt.Errorf("%w: preview source-owned Link limit %d exceeded", storage.ErrValidation, PreviewOwnedLinkLimit)
	}
	resolved := make([]resolvedIssueCreateDependency, 0, len(dependencies))
	seen := make(map[string]bool, len(dependencies))
	forward := 0
	for _, dependency := range dependencies {
		if err := validateIssueSelector(dependency.Target); err != nil {
			return nil, publicops.CreateRequest{}, err
		}
		nativeType, err := previewDependencyType(dependency.Type)
		if err != nil {
			return nil, publicops.CreateRequest{}, err
		}
		typeURL, ok := nativeDependencyTypeURLForIssueType(s.ScopeURL(), issueType, nativeType)
		if !ok {
			return nil, publicops.CreateRequest{}, fmt.Errorf("%w: native Dependency Link Type is unavailable", ErrCapabilityUnavailable)
		}
		if err := s.selectedLinkTypeInTx(ctx, tx, typeURL); err != nil {
			return nil, publicops.CreateRequest{}, err
		}
		targetPath, target, err := s.resolveLiveIssueSelectorInTx(ctx, tx, dependency.Target)
		if err != nil {
			return nil, publicops.CreateRequest{}, err
		}
		if target.Type != issueType {
			return nil, publicops.CreateRequest{}, fmt.Errorf("%w: dependency target is not an Issue of the installed Type", storage.ErrValidation)
		}
		key := fmt.Sprintf("%t|%s", dependency.Reverse, target.Properties.ID)
		if seen[key] {
			return nil, publicops.CreateRequest{}, fmt.Errorf("%w: duplicate or conflicting create Dependency pair", storage.ErrValidation)
		}
		seen[key] = true
		if dependency.Reverse {
			if len(target.Owned) >= PreviewOwnedLinkLimit {
				return nil, publicops.CreateRequest{}, fmt.Errorf("%w: preview source-owned Link limit %d exceeded", storage.ErrValidation, PreviewOwnedLinkLimit)
			}
		} else {
			forward++
		}
		resolved = append(resolved, resolvedIssueCreateDependency{targetPath: targetPath, targetID: target.Properties.ID, typeURL: typeURL, nativeType: nativeType, reverse: dependency.Reverse})
		if !dependency.Reverse {
			native.Dependencies = append(native.Dependencies, publicops.CreateDependency{TargetID: target.Properties.ID, Type: nativeType})
		}
	}
	if forward > PreviewOwnedLinkLimit {
		return nil, publicops.CreateRequest{}, fmt.Errorf("%w: preview source-owned Link limit %d exceeded", storage.ErrValidation, PreviewOwnedLinkLimit)
	}
	return resolved, native, nil
}

func (s *Store) mapIssueCreateDependenciesInTx(ctx context.Context, tx *sql.Tx, issuePath, issueID, actor string, dependencies []resolvedIssueCreateDependency) error {
	reverseSources := make(map[string]string)
	var reverseEdges []publicops.DependencyEdge
	for _, dependency := range dependencies {
		if dependency.reverse {
			reverseEdges = append(reverseEdges, publicops.DependencyEdge{IssueID: dependency.targetID, DependsOnID: issueID, Type: dependency.nativeType})
		}
	}
	if len(reverseEdges) != 0 {
		if _, _, err := issueops.ExecuteAddDependencies(ctx, tx, publicops.AddDependenciesRequest{Actor: actor, Edges: reverseEdges}); err != nil {
			return fmt.Errorf("reverse native create Dependencies: %w", err)
		}
	}
	for _, dependency := range dependencies {
		sourcePath, sourceID := issuePath, issueID
		targetPath, targetID := dependency.targetPath, dependency.targetID
		if dependency.reverse {
			sourcePath, sourceID = dependency.targetPath, dependency.targetID
			targetPath, targetID = issuePath, issueID
			reverseSources[sourcePath] = sourceID
		}
		var nativeID, nativeType string
		if err := tx.QueryRowContext(ctx, `SELECT id,type FROM dependencies WHERE issue_id=? AND depends_on_issue_id=? AND depends_on_wisp_id IS NULL AND depends_on_external IS NULL`, sourceID, targetID).Scan(&nativeID, &nativeType); err != nil {
			return fmt.Errorf("%w: native create Dependency missing: %v", ErrInvalidStore, err)
		}
		if nativeType != string(dependency.nativeType) {
			return fmt.Errorf("%w: native create Dependency kind changed", ErrInvalidStore)
		}
		pathToken, err := freshToken()
		if err != nil {
			return err
		}
		linkPath := "links/" + pathToken
		linkRevision, err := freshToken()
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO graph_preview_catalog
			(path,resource_kind,type_url,revision,allocation_state,backing,backing_key)
			VALUES (?,'link',?,?,'live','dependency',?)`, linkPath, dependency.typeURL, linkRevision, nativeID); err != nil {
			return err
		}
		if err := s.afterStage("create-dependency-catalog"); err != nil {
			return err
		}
		link, err := s.currentLinkInTx(ctx, tx, linkPath)
		if err != nil {
			return err
		}
		if link.Source != graph.CanonicalURL(s.ScopeURL(), sourcePath) || link.Target != graph.CanonicalURL(s.ScopeURL(), targetPath) || link.Type != dependency.typeURL {
			return fmt.Errorf("%w: native create Dependency cannot be represented exactly as a Link", ErrInvalidStore)
		}
		snapshot, err := canonicalJSON(link)
		if err != nil {
			return err
		}
		if err := insertPreviewVersionInTx(ctx, tx, linkPath, linkRevision, snapshot, actor); err != nil {
			return err
		}
		if err := s.afterStage("create-dependency-retained"); err != nil {
			return err
		}
	}
	if err := s.recordIssueMappingContextInTx(ctx, tx, issuePath, issueID, issueRevisionContext{Inception: "create"}); err != nil {
		return fmt.Errorf("new Issue mapping: %w", err)
	}
	for sourcePath, sourceID := range reverseSources {
		if err := s.recordIssueMappingInTx(ctx, tx, sourcePath, sourceID); err != nil {
			return fmt.Errorf("reverse source %s mapping: %w", sourcePath, err)
		}
	}
	return nil
}
