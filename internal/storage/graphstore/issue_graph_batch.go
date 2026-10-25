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

// ApplyNewIssueGraph is the bounded storage proof for a graph JSON plan whose
// nodes are all new durable Issues. An edge may start at a new key and end at
// a new key or exact existing local Issue, or start at an exact existing local
// Issue and end at a new key. Several edges may share an existing source; each
// native revision receives its matching retained canonical owned-Link set.
// The native batch body and every canonical projection run in one graph write
// transaction. Other ApplyBatch forms refuse before any write; this method is
// not the CLI's complete graph-plan contract.
func (s *Store) ApplyNewIssueGraph(ctx context.Context, request publicops.ApplyBatchRequest) (map[string]IssueRecord, error) {
	plan, err := storage.PlanApplyBatch(request)
	if err != nil {
		return nil, err
	}
	type edge struct {
		sourceKey, sourceID string
		sourcePath          string
		targetKey, targetID string
		dependency          resolvedIssueCreateDependency
	}
	var edges []edge
	creates := make(map[string]*publicops.CreateItem)
	seen := make(map[string]bool)
	for index, item := range plan.Items {
		switch item.Kind {
		case publicops.ItemCreate:
			if item.Create.Key == "" || len(item.Create.MetadataRefs) != 0 {
				return nil, fmt.Errorf("%w: graph batch item %d requires a key and cannot carry metadata refs", ErrCapabilityUnavailable, index)
			}
			if err := validateIssueCreate(publicops.CreateRequest{Actor: plan.Actor, Issue: item.Create.Issue, ForceIDPrefix: plan.ForceIDPrefix, DefaultPriority: item.Create.DefaultPriority}); err != nil {
				return nil, fmt.Errorf("graph batch item %d: %w", index, err)
			}
			creates[item.Create.Key] = item.Create
		case publicops.ItemDepAdd:
			dep := item.DepAdd
			if (dep.Source.Key == "") == (dep.Source.ID == "") || (dep.Target.Key == "") == (dep.Target.ID == "") ||
				(dep.Source.ID != "" && dep.Target.Key == "") || dep.Metadata != "" || dep.ThreadID != "" || dep.HasSpawner {
				return nil, fmt.Errorf("%w: graph batch item %d requires a new source or an exact existing source with a new keyed target, and no edge metadata", ErrCapabilityUnavailable, index)
			}
			nativeType, err := previewDependencyType(dep.Type)
			if err != nil {
				return nil, err
			}
			pair := dep.Source.Key + "\x00" + dep.Source.ID + "\x00" + dep.Target.Key + "\x00" + dep.Target.ID
			if seen[pair] {
				return nil, fmt.Errorf("%w: duplicate graph batch Dependency pair", storage.ErrValidation)
			}
			seen[pair] = true
			edges = append(edges, edge{sourceKey: dep.Source.Key, sourceID: dep.Source.ID, targetKey: dep.Target.Key, targetID: dep.Target.ID, dependency: resolvedIssueCreateDependency{nativeType: nativeType}})
		default:
			return nil, fmt.Errorf("%w: graph batch item %d is not a new Issue or local Dependency", ErrCapabilityUnavailable, index)
		}
	}
	if len(creates) == 0 {
		return nil, fmt.Errorf("%w: graph batch requires a new Issue", storage.ErrValidation)
	}
	for _, item := range edges {
		if (item.sourceKey != "" && creates[item.sourceKey] == nil) || (item.targetKey != "" && creates[item.targetKey] == nil) {
			return nil, fmt.Errorf("%w: graph batch keyed endpoints must name new Issues", ErrCapabilityUnavailable)
		}
	}
	ownedCount := make(map[string]int)
	for _, item := range edges {
		if item.sourceKey == "" {
			continue
		}
		ownedCount[item.sourceKey]++
		if ownedCount[item.sourceKey] > PreviewOwnedLinkLimit {
			return nil, fmt.Errorf("%w: preview source-owned Link limit %d exceeded", storage.ErrValidation, PreviewOwnedLinkLimit)
		}
	}
	// The ordinary apply role classifies dependencies from textual ID prefixes.
	// These endpoints were resolved as local Issues by the graph plan, so let
	// its create role allocate the Issues and add each edge through the checked
	// local writer in the same transaction below.
	nativeItems := make([]publicops.ApplyItem, 0, len(creates))
	for _, item := range plan.Items {
		if item.Kind == publicops.ItemCreate {
			nativeItems = append(nativeItems, item)
		}
	}
	nativePlan, err := storage.PlanApplyBatch(publicops.ApplyBatchRequest{
		Actor: plan.Actor, Provenance: plan.Provenance, ForceIDPrefix: plan.ForceIDPrefix,
		SkipPerEdgeCycleCheck: plan.SkipPerEdgeCycleCheck, Items: nativeItems,
	})
	if err != nil {
		return nil, err
	}

	var records map[string]IssueRecord
	err = s.withTx(ctx, true, func(tx *sql.Tx) error {
		if err := checkBinding(ctx, tx, s.options); err != nil {
			return err
		}
		issueType, err := installedIssueTypeURLInTx(ctx, tx, s.ScopeURL())
		if err != nil {
			return err
		}
		if len(edges) != 0 {
			if err := s.requireRelationshipGenerationInTx(ctx, tx); err != nil {
				return err
			}
		}
		resolvedPairs := make(map[string]bool, len(edges))
		existingSources := make(map[string]existingBatchSource)
		for index := range edges {
			if edges[index].sourceID != "" {
				sourceID := edges[index].sourceID
				if strings.HasPrefix(sourceID, s.ScopeURL()) {
					var err error
					sourceID, err = graph.ScopeLocalPublicResourceID(s.ScopeURL(), sourceID)
					if err != nil {
						return err
					}
				}
				if err := graph.ValidateCanonicalResourceID(sourceID); err != nil {
					return err
				}
				selector := IssueSelector{Path: "beads/" + sourceID, NativeID: sourceID, Exact: true}
				path, source, err := s.resolveLiveIssueSelectorInTx(ctx, tx, selector)
				if err != nil {
					return err
				}
				if !nativeIssueTypeMatches(s.ScopeURL(), source.Type, issueType) || len(source.Owned) >= PreviewOwnedLinkLimit {
					return fmt.Errorf("%w: graph batch existing source is not a local Issue with Link capacity", storage.ErrValidation)
				}
				if _, known := existingSources[path]; !known {
					var nativeRevision int64
					if err := tx.QueryRowContext(ctx, `SELECT current_revision FROM issues WHERE id=?`, source.Properties.ID).Scan(&nativeRevision); err != nil {
						return err
					}
					existingSources[path] = existingBatchSource{record: source, nativeRevision: nativeRevision}
				}
				edges[index].sourcePath = path
				edges[index].sourceID = source.Properties.ID
			}
			if edges[index].targetID != "" {
				targetID := edges[index].targetID
				if strings.HasPrefix(targetID, s.ScopeURL()) {
					var err error
					targetID, err = graph.ScopeLocalPublicResourceID(s.ScopeURL(), targetID)
					if err != nil {
						return err
					}
				}
				if err := graph.ValidateCanonicalResourceID(targetID); err != nil {
					return err
				}
				selector := IssueSelector{Path: "beads/" + targetID, NativeID: targetID, Exact: true}
				path, target, err := s.resolveLiveIssueSelectorInTx(ctx, tx, selector)
				if err != nil {
					return err
				}
				if !nativeIssueTypeMatches(s.ScopeURL(), target.Type, issueType) {
					return fmt.Errorf("%w: graph batch target is not a local Issue", storage.ErrValidation)
				}
				edges[index].dependency.targetID = target.Properties.ID
				edges[index].dependency.targetPath = path
			}
			target := "key:" + edges[index].targetKey
			if edges[index].targetID != "" {
				target = "id:" + edges[index].dependency.targetID
			}
			source := "key:" + edges[index].sourceKey
			if edges[index].sourceID != "" {
				source = "id:" + edges[index].sourceID
			}
			pair := source + "\x00" + target
			if resolvedPairs[pair] {
				return fmt.Errorf("%w: duplicate graph batch Dependency pair", storage.ErrValidation)
			}
			resolvedPairs[pair] = true
			typeURL, ok := nativeDependencyTypeURLForIssueType(s.ScopeURL(), issueType, edges[index].dependency.nativeType)
			if !ok {
				return fmt.Errorf("%w: native Dependency Link Type is unavailable", ErrCapabilityUnavailable)
			}
			if err := s.selectedLinkTypeInTx(ctx, tx, typeURL); err != nil {
				return err
			}
			edges[index].dependency.typeURL = typeURL
		}
		for path, source := range existingSources {
			count := 0
			for _, item := range edges {
				if item.sourcePath == path {
					count++
				}
			}
			if len(source.record.Owned)+count > PreviewOwnedLinkLimit {
				return fmt.Errorf("%w: preview source-owned Link limit %d exceeded", storage.ErrValidation, PreviewOwnedLinkLimit)
			}
		}
		if err := s.touchCoordination(ctx, tx); err != nil {
			return err
		}
		unscope := issueops.ScopeVersionedHistoryTransaction(tx, true)
		defer unscope()
		applied, _, err := issueops.ApplyBatchInTx(ctx, tx, nativePlan)
		if err != nil {
			return err
		}
		if err := s.afterStage("issue"); err != nil {
			return err
		}
		// Declare every concrete native Issue Type before testing any public
		// ID. A first-use Type in a later item must reserve its spelling for
		// the entire batch, independent of map iteration order.
		declaredTypes := make(map[string]string, len(creates))
		for key, create := range creates {
			id := applied.Keys[key]
			if id == "" {
				return fmt.Errorf("%w: graph batch did not bind key %q", ErrInvalidStore, key)
			}
			var issue *publicops.Issue
			for _, result := range applied.Items {
				if result.Kind == publicops.ItemCreate && result.IssueID == id {
					issue = result.Issue
					break
				}
			}
			if !sameIssueCreateFields(create.Issue, issue, plan.Actor) || issueops.IsWisp(issue) {
				return fmt.Errorf("%w: graph batch Issue %q cannot be represented exactly", storage.ErrValidation, key)
			}
			declared, err := NativeIssueTypeURL(s.ScopeURL(), string(issue.IssueType))
			if err != nil {
				return err
			}
			if err := s.ensureNativeTypeInTx(ctx, tx, declared); err != nil {
				return err
			}
			declaredTypes[key] = declared
		}
		paths := make(map[string]string, len(creates))
		for key := range creates {
			id := applied.Keys[key]
			if exists, err := publicResourceIDExistsInTx(ctx, tx, id); err != nil {
				return err
			} else if exists {
				return ErrAlreadyExists
			}
			path, err := generatedIssuePath(ctx, tx, id)
			if err != nil {
				return err
			}
			revision, err := freshToken()
			if err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, `INSERT INTO graph_preview_catalog
				(path,resource_kind,type_url,revision,allocation_state,backing,backing_key)
				VALUES (?,'bead',?,?,'live','issue',?)`, path, declaredTypes[key], revision, id); err != nil {
				return err
			}
			if err := s.afterStage("issue-catalog"); err != nil {
				return err
			}
			paths[key] = path
		}
		bySource := make(map[string][]resolvedIssueCreateDependency)
		byExistingSource := make(map[string][]resolvedIssueCreateDependency)
		for _, item := range edges {
			dependency := item.dependency
			if item.targetKey != "" {
				dependency.targetID = applied.Keys[item.targetKey]
				dependency.targetPath = paths[item.targetKey]
			}
			if item.sourcePath != "" {
				byExistingSource[item.sourcePath] = append(byExistingSource[item.sourcePath], dependency)
			} else {
				bySource[item.sourceKey] = append(bySource[item.sourceKey], dependency)
			}
		}
		records = make(map[string]IssueRecord, len(creates))
		for key := range creates {
			id, path := applied.Keys[key], paths[key]
			if dependencies := bySource[key]; len(dependencies) != 0 {
				if err := s.mapIssueCreateDependenciesInTx(ctx, tx, path, id, plan.Actor, dependencies, true); err != nil {
					return err
				}
			} else if err := s.recordIssueMappingContextInTx(ctx, tx, path, id, issueRevisionContext{Inception: "create"}); err != nil {
				return err
			}
			record, err := s.showIssueInTx(ctx, tx, path)
			if err != nil {
				return err
			}
			records[key] = record
		}
		for path, dependencies := range byExistingSource {
			source := existingSources[path]
			if err := s.mapExistingBatchSourceInTx(ctx, tx, path, source, plan.Actor, dependencies); err != nil {
				return err
			}
		}
		return checkCurrentReadBytes(ctx, tx)
	})
	if err != nil {
		return nil, err
	}
	return records, nil
}

type existingBatchSource struct {
	record         IssueRecord
	nativeRevision int64
}

// mapExistingBatchSourceInTx adds each Link through the checked local writer
// and retains the source's canonical owned-Link set at each native revision.
// This preserves intermediate History while keeping all writes in one graph
// transaction.
func (s *Store) mapExistingBatchSourceInTx(ctx context.Context, tx *sql.Tx, path string, source existingBatchSource, actor string, edges []resolvedIssueCreateDependency) error {
	id := source.record.Properties.ID
	for i, item := range edges {
		if err := s.mapIssueCreateDependenciesInTx(ctx, tx, path, id, actor, []resolvedIssueCreateDependency{item}, false); err != nil {
			return err
		}
		var ordinal int64
		if err := tx.QueryRowContext(ctx, `SELECT current_revision FROM issues WHERE id=?`, id).Scan(&ordinal); err != nil {
			return err
		}
		if ordinal != source.nativeRevision+int64(i)+1 {
			return fmt.Errorf("%w: graph batch source Link did not mint one native version", ErrInvalidStore)
		}
		var retained int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM issue_versions WHERE issue_id=? AND revision=?`, id, ordinal).Scan(&retained); err != nil {
			return err
		}
		if retained != 1 {
			return fmt.Errorf("%w: graph batch native source revision is missing", ErrInvalidStore)
		}
		owned, err := s.ownedLinksInTx(ctx, tx, id)
		if err != nil {
			return err
		}
		previous, err := currentIssueRevisionContextInTx(ctx, tx, path)
		if err != nil {
			return err
		}
		if err := s.recordIssueMappingSnapshotInTx(ctx, tx, path, id, ordinal, owned, previous); err != nil {
			return err
		}
	}
	return nil
}
