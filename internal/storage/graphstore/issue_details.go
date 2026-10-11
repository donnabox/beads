package graphstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/storage/domain"
	"github.com/steveyegge/beads/internal/storage/issueops"
	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/internal/workapi"
)

// ShowIssueDetails projects a current graph Issue through the same detail
// assembler as ordinary bd show. The graph revision is the equality token
// accepted by graph --if-revision; native row_lock is never exposed here.
func (s *Store) ShowIssueDetails(ctx context.Context, path string, opts workapi.DetailOptions) (*types.IssueDetails, error) {
	if err := validatePath(path); err != nil {
		return nil, err
	}
	var details *types.IssueDetails
	err := s.withTx(ctx, false, func(tx *sql.Tx) error {
		if err := checkCurrentReadBytes(ctx, tx); err != nil {
			return err
		}
		if err := checkBinding(ctx, tx, s.options); err != nil {
			return err
		}
		record, err := s.showIssueInTx(ctx, tx, path)
		if err != nil {
			return err
		}
		if record.Properties == nil {
			return fmt.Errorf("%w: Issue has no properties", ErrInvalidStore)
		}
		// Read the native Issue in this same verified snapshot. Its metadata
		// and labels keep their ordinary JSON representation; the graph
		// record's common metadata and separated owned Links are different
		// projections of the same state.
		issue, err := issueops.GetIssueInTx(ctx, tx, record.Properties.ID)
		if err != nil {
			return err
		}
		details, err = workapi.BuildIssueDetails(ctx, graphIssueDetailSource{tx: tx}, issue, false, opts)
		if err != nil {
			return err
		}
		details.Revision = record.Revision
		// Native Issue IDs are backing keys, not graph selectors. Project every
		// Issue reference through this same transaction's catalog before the
		// detail array escapes; a bare backing ID can otherwise resolve to an
		// unrelated Memory with the same Bead path.
		if err := s.projectIssueDetailIDsInTx(ctx, tx, details); err != nil {
			return err
		}
		if details.ID != path {
			return fmt.Errorf("%w: Issue allocation changed during detail projection", ErrInvalidStore)
		}
		return nil
	})
	return details, err
}

func (s *Store) projectIssueDetailIDsInTx(ctx context.Context, tx *sql.Tx, details *types.IssueDetails) error {
	paths := map[string]string{}
	canonicalPath := func(nativeID string) (string, error) {
		if path, ok := paths[nativeID]; ok {
			return path, nil
		}
		var path, kind, typ, state string
		err := tx.QueryRowContext(ctx, `SELECT path,resource_kind,type_url,allocation_state FROM graph_preview_catalog WHERE backing='issue' AND backing_key=?`, nativeID).Scan(&path, &kind, &typ, &state)
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("%w: unmapped Issue detail ID", ErrInvalidStore)
		}
		if err != nil {
			return "", err
		}
		if kind != "bead" || typ != IssueTypeURL(s.ScopeURL()) || state != "live" || validatePath(path) != nil {
			return "", fmt.Errorf("%w: invalid Issue detail allocation", ErrInvalidStore)
		}
		paths[nativeID] = path
		return path, nil
	}
	projectIssue := func(issue *types.Issue) error {
		path, err := canonicalPath(issue.ID)
		if err != nil {
			return err
		}
		issue.ID = path
		return nil
	}
	if err := projectIssue(&details.Issue); err != nil {
		return err
	}
	for _, relation := range [][]*types.IssueWithDependencyMetadata{details.Dependencies, details.Dependents} {
		for _, related := range relation {
			if related != nil {
				if err := projectIssue(&related.Issue); err != nil {
					return err
				}
			}
		}
	}
	if details.Parent != nil {
		path, err := canonicalPath(*details.Parent)
		if err != nil {
			return err
		}
		details.Parent = &path
	}
	for _, comment := range details.Comments {
		if comment != nil {
			path, err := canonicalPath(comment.IssueID)
			if err != nil {
				return err
			}
			comment.IssueID = path
		}
	}
	return nil
}

// graphIssueDetailSource uses the native Issue read primitives in the same
// snapshot that verified the graph catalog and retained Issue head.
type graphIssueDetailSource struct{ tx *sql.Tx }

func (s graphIssueDetailSource) GetIssue(ctx context.Context, id string) (*types.Issue, error) {
	return issueops.GetIssueInTx(ctx, s.tx, id)
}

func (graphIssueDetailSource) GetWisp(_ context.Context, _ string) (*types.Issue, error) {
	return nil, storage.ErrNotFound
}

func (s graphIssueDetailSource) Labels(ctx context.Context, id string, _ bool) ([]string, error) {
	return issueops.GetLabelsInTx(ctx, s.tx, "labels", id)
}

func (s graphIssueDetailSource) Dependencies(ctx context.Context, id string, _ bool) ([]*types.IssueWithDependencyMetadata, error) {
	return issueops.GetDependenciesWithMetadataInTx(ctx, s.tx, id)
}

func (s graphIssueDetailSource) CountDependencies(ctx context.Context, id string, _ bool) (int64, error) {
	return issueops.CountDependencyEdgesInTx(ctx, s.tx, id, domain.DepDirectionOut, nil)
}

func (s graphIssueDetailSource) CountDependents(ctx context.Context, id string, _ bool) (int64, error) {
	return issueops.CountDependencyEdgesInTx(ctx, s.tx, id, domain.DepDirectionIn, nil)
}

func (s graphIssueDetailSource) CountComments(ctx context.Context, id string, _ bool) (int64, error) {
	counts, err := issueops.GetCommentCountsInTx(ctx, s.tx, []string{id})
	return int64(counts[id]), err
}

func (s graphIssueDetailSource) IterDependents(ctx context.Context, id string, _ bool) (storage.Iter[types.IssueWithDependencyMetadata], error) {
	items, err := issueops.GetDependentsWithMetadataInTx(ctx, s.tx, id)
	if err != nil {
		return nil, err
	}
	return storage.NewSliceIter(items), nil
}

func (s graphIssueDetailSource) IterComments(ctx context.Context, id string, _ bool) (storage.Iter[types.Comment], error) {
	items, err := issueops.GetIssueCommentsInTx(ctx, s.tx, id)
	if err != nil {
		return nil, err
	}
	return storage.NewSliceIter(items), nil
}
