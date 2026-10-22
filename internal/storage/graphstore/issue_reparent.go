package graphstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"unicode/utf8"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/storage/issueops"
	"github.com/steveyegge/beads/internal/types"
	publicops "github.com/steveyegge/beads/issueops"
)

// IssueReparentRequest replaces a durable Issue's parent-child edge with one
// local Issue. Empty parent (detach) is deliberately outside this operation:
// dotted native IDs still imply their original parent in legacy selectors.
type IssueReparentRequest struct {
	Path, NativeID, ParentPath, ParentNativeID string
	Actor, ExpectedRevision                    string
	Unconditional                              bool
	// Edit is applied by the same native writer that replaces ParentID. Its
	// selector, guard and actor come from the enclosing reparent request.
	Edit UpdateIssueRequest
}

type IssueReparentResult struct {
	Issue   IssueRecord     `json:"issue"`
	Removed []LinkTombstone `json:"removed,omitempty"`
	Added   *LinkRecord     `json:"added,omitempty"`
	Changed bool            `json:"changed"`
}

// ReparentIssue composes the native parent replacement, old canonical Link
// tombstones, new canonical Link and child-owned History in one transaction.
// The native Issue writer remains the only writer of dependency edges.
func (s *Store) ReparentIssue(ctx context.Context, request IssueReparentRequest) (IssueReparentResult, error) {
	for _, path := range []string{request.Path, request.ParentPath} {
		if err := validatePath(path); err != nil {
			return IssueReparentResult{}, err
		}
	}
	for _, selector := range []struct{ path, nativeID string }{
		{request.Path, request.NativeID}, {request.ParentPath, request.ParentNativeID},
	} {
		if selector.nativeID != "" && selector.path != "beads/"+selector.nativeID {
			return IssueReparentResult{}, fmt.Errorf("%w: native Issue ID requires its matching bare selector", storage.ErrValidation)
		}
	}
	if request.Path == request.ParentPath || request.Actor == "" || !utf8.ValidString(request.Actor) {
		return IssueReparentResult{}, fmt.Errorf("%w: reparent requires distinct local Issues and a UTF-8 actor", storage.ErrValidation)
	}
	if request.Edit.Path != "" || request.Edit.NativeID != "" || request.Edit.Actor != "" ||
		request.Edit.ExpectedRevision != "" || request.Edit.Unconditional || request.Edit.FromRevision != "" ||
		len(request.Edit.PropertiesPatch) != 0 {
		return IssueReparentResult{}, fmt.Errorf("%w: reparent edit cannot supply a second selector, guard or historical source", storage.ErrValidation)
	}
	patch, _, err := prepareIssueUpdatePatch(request.Edit)
	if err != nil {
		return IssueReparentResult{}, err
	}
	patch.ParentID = publicops.Field[string]{Set: true}
	var result IssueReparentResult
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
		childPath, child, err := s.resolveLiveIssueSelectorInTx(ctx, tx, IssueSelector{Path: request.Path, NativeID: request.NativeID})
		if err != nil {
			return err
		}
		parentPath, parent, err := s.resolveLiveIssueSelectorInTx(ctx, tx, IssueSelector{Path: request.ParentPath, NativeID: request.ParentNativeID})
		if err != nil {
			return err
		}
		if childPath == parentPath || child.Type != issueType || parent.Type != issueType || child.Properties.ID == parent.Properties.ID {
			return fmt.Errorf("%w: reparent requires two distinct Issues of the installed relationship Type", storage.ErrValidation)
		}
		if err := checkRevisionGuard(request.ExpectedRevision, request.Unconditional, child.Revision, true, "Issue"); err != nil {
			return err
		}
		// Capture canonical Links before the native writer removes their backing
		// rows. Every old parent edge must have exactly one live graph identity.
		rows, err := tx.QueryContext(ctx, `SELECT id,depends_on_issue_id FROM dependencies WHERE issue_id=? AND type=? ORDER BY id`, child.Properties.ID, string(types.DepParentChild))
		if err != nil {
			return err
		}
		type parentEdge struct{ id, target string }
		var edges []parentEdge
		for rows.Next() {
			var edge parentEdge
			if err := rows.Scan(&edge.id, &edge.target); err != nil {
				_ = rows.Close()
				return err
			}
			edges = append(edges, edge)
		}
		if err := rows.Err(); err != nil {
			_ = rows.Close()
			return err
		}
		if err := rows.Close(); err != nil {
			return err
		}
		if len(edges) == 0 && len(child.Owned) >= PreviewOwnedLinkLimit {
			return fmt.Errorf("%w: preview source-owned Link limit %d exceeded", storage.ErrValidation, PreviewOwnedLinkLimit)
		}
		type oldLink struct {
			path   string
			record LinkRecord
		}
		old := make([]oldLink, 0, len(edges))
		for _, edge := range edges {
			var path string
			if err := tx.QueryRowContext(ctx, `SELECT path FROM graph_preview_catalog WHERE resource_kind='link' AND allocation_state='live' AND backing='dependency' AND backing_key=?`, edge.id).Scan(&path); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return fmt.Errorf("%w: parent edge has no canonical Link", ErrInvalidStore)
				}
				return err
			}
			link, err := s.currentLinkInTx(ctx, tx, path)
			if err != nil {
				return err
			}
			if link.Source != child.ID || link.Type != linkType || link.Properties["dep_type"] != string(types.DepParentChild) {
				return fmt.Errorf("%w: parent Link differs from native edge", ErrInvalidStore)
			}
			old = append(old, oldLink{path, link})
		}
		sameParent := len(edges) == 1 && edges[0].target == parent.Properties.ID
		patch.ParentID.Value = parent.Properties.ID
		attempt := publicops.UpdateRequest{Actor: request.Actor, IssueID: child.Properties.ID, IssuePlaneOnly: true,
			Patch: patch, ForceNotesOverwrite: request.Edit.ForceNotesOverwrite,
			ForceAssigneeTransfer: request.Edit.ForceAssigneeTransfer}
		if err := issueops.ValidateUpdateRequest(attempt); err != nil {
			return err
		}
		if sameParent {
			beforeNative := *child.Properties
			beforeNative.Metadata = child.Metadata
			updates, err := issueops.ResolveMergeOps(&beforeNative, issueops.UpdateFields(patch))
			if err != nil {
				return err
			}
			updates, err = issueops.DiscardNoopIssueUpdates(&beforeNative, updates)
			if err != nil {
				return err
			}
			resolvedMetadata, metadataChanged, err := issueops.ApplyMetadataPatch(child.Metadata, patch.Metadata)
			if err != nil {
				return err
			}
			if _, err := commonMetadata(resolvedMetadata); err != nil {
				return err
			}
			if len(updates) == 0 && !metadataChanged {
				result = IssueReparentResult{Issue: child}
				return nil
			}
		}
		if err := s.touchCoordination(ctx, tx); err != nil {
			return err
		}
		unscope := issueops.ScopeVersionedHistoryTransaction(tx, true)
		defer unscope()
		updated, _, err := issueops.ExecuteUpdate(ctx, tx, attempt)
		if err != nil {
			return err
		}
		if !updated.Changed {
			if sameParent {
				result = IssueReparentResult{Issue: child}
				return nil
			}
			return fmt.Errorf("%w: parent replacement unexpectedly became a no-op", ErrInvalidStore)
		}
		if err := checkIssueScalarStoredIntent(patch, updated.Issue); err != nil {
			return err
		}
		if err := s.afterStage("issue-reparent"); err != nil {
			return err
		}
		if sameParent {
			if err := s.recordIssueMappingInTx(ctx, tx, childPath, child.Properties.ID); err != nil {
				return err
			}
			if patch.Notes.Set || patch.AppendNotes.Set || patch.Metadata.Replace.Set || patch.Metadata.Merge.Set ||
				len(patch.Metadata.Set) > 0 || len(patch.Metadata.Unset) > 0 {
				if err := checkCurrentReadBytes(ctx, tx); err != nil {
					return err
				}
			}
			current, err := s.showIssueInTx(ctx, tx, childPath)
			if err != nil {
				return err
			}
			result = IssueReparentResult{Issue: current, Changed: true}
			return nil
		}
		result.Removed = make([]LinkTombstone, 0, len(old))
		for _, link := range old {
			revision, err := freshToken()
			if err != nil {
				return err
			}
			tombstone := LinkTombstone{ID: link.record.ID, Type: link.record.Type, Revision: revision, Version: revision,
				State: "deleted", PreviousVersion: link.record.Version, Attribution: writeAttribution(request.Actor)}
			if _, err := tx.ExecContext(ctx, `UPDATE graph_preview_catalog SET allocation_state='deleted',revision=?,backing_key=NULL WHERE path=?`, revision, link.path); err != nil {
				return err
			}
			snapshot, err := canonicalJSON(tombstone)
			if err != nil {
				return err
			}
			if err := insertPreviewVersionInTx(ctx, tx, link.path, revision, snapshot, request.Actor); err != nil {
				return err
			}
			result.Removed = append(result.Removed, tombstone)
		}
		if err := s.afterStage("reparent-old-links"); err != nil {
			return err
		}
		var dependencyID, dependencyType string
		if err := tx.QueryRowContext(ctx, `SELECT id,type FROM dependencies WHERE issue_id=? AND depends_on_issue_id=?`, child.Properties.ID, parent.Properties.ID).
			Scan(&dependencyID, &dependencyType); err != nil {
			return err
		}
		if dependencyType != string(types.DepParentChild) {
			return fmt.Errorf("%w: replacement edge has wrong native kind", ErrInvalidStore)
		}
		linkToken, err := freshToken()
		if err != nil {
			return err
		}
		linkPath := "links/" + linkToken
		linkRevision, err := freshToken()
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO graph_preview_catalog
			(path,resource_kind,type_url,revision,allocation_state,backing,backing_key)
			VALUES (?,'link',?,?,'live','dependency',?)`, linkPath, linkType, linkRevision, dependencyID); err != nil {
			return err
		}
		if err := s.afterStage("reparent-new-link"); err != nil {
			return err
		}
		added, err := s.currentLinkInTx(ctx, tx, linkPath)
		if err != nil {
			return err
		}
		if added.Source != child.ID || added.Target != graph.CanonicalURL(s.ScopeURL(), parentPath) || added.Type != linkType {
			return fmt.Errorf("%w: replacement Link differs from native edge", ErrInvalidStore)
		}
		snapshot, err := canonicalJSON(added)
		if err != nil {
			return err
		}
		if err := insertPreviewVersionInTx(ctx, tx, linkPath, linkRevision, snapshot, request.Actor); err != nil {
			return err
		}
		if err := s.afterStage("reparent-link-retained"); err != nil {
			return err
		}
		if err := s.recordIssueMappingInTx(ctx, tx, childPath, child.Properties.ID); err != nil {
			return err
		}
		if patch.Notes.Set || patch.AppendNotes.Set || patch.Metadata.Replace.Set || patch.Metadata.Merge.Set ||
			len(patch.Metadata.Set) > 0 || len(patch.Metadata.Unset) > 0 {
			if err := checkCurrentReadBytes(ctx, tx); err != nil {
				return err
			}
		}
		if err := s.afterStage("reparent-issue-retained"); err != nil {
			return err
		}
		current, err := s.showIssueInTx(ctx, tx, childPath)
		if err != nil {
			return err
		}
		result.Issue, result.Added, result.Changed = current, &added, true
		return nil
	})
	if err != nil {
		return IssueReparentResult{}, err
	}
	return result, nil
}
