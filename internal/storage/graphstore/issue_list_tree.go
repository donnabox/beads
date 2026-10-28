package graphstore

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
	"strings"

	"github.com/steveyegge/beads/internal/storage/issueops"
	"github.com/steveyegge/beads/internal/types"
	publicops "github.com/steveyegge/beads/issueops"
)

// IssueTreeSelected is a checked, unpaged hierarchy for the ordinary human
// `bd list --parent` view. IDs in the result are native public Issue IDs. The root
// is separate because the row cap and list filters apply to descendants only.
type IssueTreeSelected struct {
	Root     *types.Issue
	Items    []*types.Issue
	Children map[string][]string
}

// ListIssueTreeSelected gathers filtered descendants and their hierarchy in
// one graph read transaction. The graph preview admits at most 1,000 live
// Resources, so a complete candidate scan is bounded without page cursors.
// Filtered-out intermediary Issues do not lead to edge descendants; dotted
// graph IDs retain the shipping direct-prefix fallback.
func (s *Store) ListIssueTreeSelected(ctx context.Context, request publicops.ListRequest, parent IssueSelector) (IssueTreeSelected, error) {
	if err := validateIssueSelector(parent); err != nil {
		return IssueTreeSelected{}, err
	}
	prepared, err := prepareIssueListRequest(request)
	if err != nil {
		return IssueTreeSelected{}, err
	}
	var result IssueTreeSelected
	err = s.withTx(ctx, false, func(tx *sql.Tx) error {
		// Validate the complete catalog and bound the candidate set before
		// resolving the root against it. Both reads use this transaction.
		all := prepared
		all.Limit = treeIssueListInt(PreviewSnapshotLimit)
		all.MaxRows = 0
		page, err := s.listIssuesWithParentInTx(ctx, tx, &all, nil)
		if err != nil {
			return err
		}
		if page.HasMore {
			return fmt.Errorf("%w: Issue tree candidate scan was truncated", ErrInvalidStore)
		}
		_, rootRecord, err := s.resolveLiveIssueSelectorInTx(ctx, tx, parent)
		if err != nil {
			return err
		}
		if rootRecord.Properties == nil || rootRecord.Properties.ID == "" {
			return fmt.Errorf("%w: incomplete parent Issue mapping", ErrInvalidStore)
		}
		rootNative := rootRecord.Properties.ID
		root, err := issueops.GetIssueInTx(ctx, tx, rootNative)
		if err != nil {
			return err
		}
		root.ID = rootNative

		// The ordinary parent tree is not page-truncated. The original caller
		// cap is enforced on deduplicated descendants below.
		byNative := make(map[string]*types.Issue, len(page.Items))
		for _, record := range page.Items {
			if record.Properties == nil || record.Properties.ID == "" || !strings.HasPrefix(record.ID, s.ScopeURL()) {
				return fmt.Errorf("%w: incomplete Issue tree candidate", ErrInvalidStore)
			}
			native := record.Properties.ID
			if native == rootNative {
				continue
			}
			if byNative[native] != nil {
				return fmt.Errorf("%w: duplicate Issue tree candidate", ErrInvalidStore)
			}
			candidate, err := issueops.GetIssueInTx(ctx, tx, native)
			if err != nil {
				return err
			}
			candidate.ID = native
			byNative[native] = candidate
		}

		edges, err := tx.QueryContext(ctx, fmt.Sprintf("SELECT issue_id, %s FROM dependencies WHERE type='parent-child'", issueops.DepTargetExpr))
		if err != nil {
			return err
		}
		childrenByNative := make(map[string][]string)
		parentsByNative := make(map[string][]string)
		hasParentEdge := make(map[string]bool)
		for edges.Next() {
			var child, target string
			if err := edges.Scan(&child, &target); err != nil {
				_ = edges.Close()
				return err
			}
			if byNative[child] != nil {
				hasParentEdge[child] = true
				if target == rootNative || byNative[target] != nil {
					childrenByNative[target] = append(childrenByNative[target], child)
					parentsByNative[child] = append(parentsByNative[child], target)
				}
			}
		}
		if err := edges.Err(); err != nil {
			_ = edges.Close()
			return err
		}
		if err := edges.Close(); err != nil {
			return err
		}

		selected := make(map[string]bool)
		queue := []string{rootNative}
		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]
			currentPath := rootNative
			if current != rootNative {
				currentPath = byNative[current].ID
			}
			for _, child := range childrenByNative[current] {
				if !selected[child] {
					selected[child] = true
					queue = append(queue, child)
				}
			}
			for native, candidate := range byNative {
				if !selected[native] && !hasParentEdge[native] && strings.HasPrefix(candidate.ID, currentPath+".") {
					selected[native] = true
					queue = append(queue, native)
				}
			}
		}
		if err := issueops.EnforceMaxRowsCap(len(selected), prepared.MaxRows, prepared.MaxRowsSource); err != nil {
			return fmt.Errorf("%w: %w", ErrLimitExceeded, err)
		}
		result = IssueTreeSelected{Root: root, Items: make([]*types.Issue, 0, len(selected)), Children: make(map[string][]string)}
		for native := range selected {
			result.Items = append(result.Items, byNative[native])
		}
		for native := range selected {
			childPath := byNative[native].ID
			linked := false
			for _, target := range parentsByNative[native] {
				if target == rootNative || selected[target] {
					parentPath := rootNative
					if target != rootNative {
						parentPath = byNative[target].ID
					}
					result.Children[parentPath] = append(result.Children[parentPath], childPath)
					linked = true
				}
			}
			if linked {
				continue
			}
			// The dotted fallback nests under the nearest visible ancestor.
			// A missing or filtered intermediate never hides a descendant.
			parentPath := rootNative
			for _, candidate := range result.Items {
				if candidate.ID != childPath && strings.HasPrefix(childPath, candidate.ID+".") && len(candidate.ID) > len(parentPath) {
					parentPath = candidate.ID
				}
			}
			result.Children[parentPath] = append(result.Children[parentPath], childPath)
		}
		for parentPath, children := range result.Children {
			slices.Sort(children)
			result.Children[parentPath] = slices.Compact(children)
		}
		return nil
	})
	if err != nil {
		return IssueTreeSelected{}, err
	}
	return result, nil
}

func treeIssueListInt(value int) *int { return &value }
