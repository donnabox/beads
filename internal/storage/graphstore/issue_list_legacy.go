package graphstore

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/internal/workapi"
	publicops "github.com/steveyegge/beads/issueops"
)

// LegacyIssueListPage preserves ordinary bd list --json's Issue row shape.
// It is separate from the complete graph-record projection used by
// --format records-json.
type LegacyIssueListPage struct {
	Items   []*types.IssueWithCounts
	HasMore bool
}

// ListIssuesLegacy reads counted native Issue rows and projects their IDs in
// one checked graph snapshot. The existing graph Issue list first validates
// every selected record, including the over-fetched pagination probe.
func (s *Store) ListIssuesLegacy(ctx context.Context, request publicops.ListRequest) (LegacyIssueListPage, error) {
	return s.listIssuesLegacy(ctx, request, nil)
}

// ListIssuesLegacySelected keeps the parent resolution and counted legacy
// projection in the same checked read transaction.
func (s *Store) ListIssuesLegacySelected(ctx context.Context, request publicops.ListRequest, parent IssueSelector) (LegacyIssueListPage, error) {
	if err := validateIssueSelector(parent); err != nil {
		return LegacyIssueListPage{}, err
	}
	return s.listIssuesLegacy(ctx, request, &parent)
}

func (s *Store) listIssuesLegacy(ctx context.Context, request publicops.ListRequest, parent *IssueSelector) (LegacyIssueListPage, error) {
	includeComments := request.IncludeComments
	request.IncludeComments = false
	request, err := prepareIssueListRequest(request)
	if err != nil {
		return LegacyIssueListPage{}, err
	}
	request.IncludeComments = includeComments
	if request.ReadyFlag && parent != nil {
		return LegacyIssueListPage{}, fmt.Errorf("%w: graph Issue list cannot combine --ready with --parent", ErrCapabilityUnavailable)
	}
	if request.ReadyFlag {
		s.wakeExpiredDefersAdvisory(ctx)
	}
	var result LegacyIssueListPage
	err = s.withTx(ctx, false, func(tx *sql.Tx) error {
		page, err := s.listIssuesWithParentInTx(ctx, tx, &request, parent)
		if err != nil {
			return err
		}
		if request.IDFilter != "" && len(page.Items) == 0 {
			// The checked query already validated the entire selected set.
			// In particular, an unmatched --id set has no native keys for the
			// counted query to project.
			result = LegacyIssueListPage{Items: []*types.IssueWithCounts{}, HasMore: page.HasMore}
			return nil
		}
		cfg, err := issueListConfigInTx(ctx, tx)
		if err != nil {
			return err
		}
		filter, err := workapi.BuildListFilter(request, cfg)
		if err != nil {
			return err
		}
		filter.SkipCounts = false
		if parent != nil || filter.Limit > PreviewSnapshotLimit {
			filter.Limit = PreviewSnapshotLimit
		}
		if parent == nil {
			filter = workapi.WithFetchOneExtra(workapi.WithRowsBeforeThePage(filter, request.Offset))
		} else {
			filter.MaxRows = 0
		}
		rows, err := queryIssueListRowsInTx(ctx, tx, filter, request.ReadyFlag)
		if err != nil {
			return err
		}
		if parent != nil {
			parentNativeID, parentPath, err := s.resolveIssueListParentInTx(ctx, tx, *parent)
			if err != nil {
				return err
			}
			selected := make([]*types.IssueWithCounts, 0, len(rows))
			for _, row := range rows {
				if row == nil || row.Issue == nil {
					return fmt.Errorf("%w: nil counted Issue list row", ErrInvalidStore)
				}
				var path string
				if err := tx.QueryRowContext(ctx, "SELECT path FROM graph_preview_catalog WHERE backing='issue' AND backing_key=? AND allocation_state='live'", row.ID).Scan(&path); err != nil {
					return fmt.Errorf("%w: unmapped counted Issue list row: %v", ErrInvalidStore, err)
				}
				if err := validatePath(path); err != nil {
					return fmt.Errorf("%w: malformed counted Issue list path: %v", ErrInvalidStore, err)
				}
				if issueListParentMatches(row, path, parentNativeID, parentPath) {
					selected = append(selected, row)
				}
			}
			rows = selected
		}
		var more bool
		if request.SortBy == "id" {
			// The checked graph page is ordered by visible Bead IDs. The native
			// counted query still names backing IDs, so select its rows in that
			// checked order instead of sorting those opaque keys a second time.
			more = len(rows) > len(page.Items)
			byID := make(map[string]*types.IssueWithCounts, len(rows))
			for _, row := range rows {
				if row == nil || row.Issue == nil || byID[row.ID] != nil {
					return fmt.Errorf("%w: invalid counted Issue list row", ErrInvalidStore)
				}
				byID[row.ID] = row
			}
			ordered := make([]*types.IssueWithCounts, 0, len(page.Items))
			for _, item := range page.Items {
				if item.Properties == nil || byID[item.Properties.ID] == nil {
					return fmt.Errorf("%w: counted Issue list disagrees with checked graph page", ErrInvalidStore)
				}
				ordered = append(ordered, byID[item.Properties.ID])
			}
			rows = ordered
		} else {
			rows, more = workapi.FinishPageAt(rows, request.SortBy, request.Reverse, request.Offset, workapi.PageLimit(request), false)
		}
		if rows == nil {
			rows = []*types.IssueWithCounts{}
		}
		if more != page.HasMore || len(rows) != len(page.Items) {
			return fmt.Errorf("%w: counted Issue list disagrees with checked graph page", ErrInvalidStore)
		}
		for i, row := range rows {
			if row == nil || row.Issue == nil || page.Items[i].Properties == nil || row.ID != page.Items[i].Properties.ID {
				return fmt.Errorf("%w: counted Issue list disagrees with checked graph row", ErrInvalidStore)
			}
		}
		if err := workapi.HydrateListComments(ctx, func() workapi.CommentStreamer { return graphIssueDetailSource{tx: tx} }, workapi.SourceRoutesCommentPlanes, rows, request.IncludeComments, true); err != nil {
			return err
		}
		for i, row := range rows {
			canonical := page.Items[i].ID
			if !strings.HasPrefix(canonical, s.ScopeURL()) {
				return fmt.Errorf("%w: listed Issue escaped graph scope", ErrInvalidStore)
			}
			path := strings.TrimPrefix(canonical, s.ScopeURL())
			if err := validatePath(path); err != nil {
				return fmt.Errorf("%w: invalid listed Issue path: %v", ErrInvalidStore, err)
			}
			row.ID = path
			if row.Parent != nil {
				parent, err := s.canonicalIssuePathInTx(ctx, tx, *row.Parent)
				if err != nil {
					return err
				}
				row.Parent = &parent
			}
			for _, comment := range row.Comments {
				if comment != nil {
					comment.IssueID = path
				}
			}
		}
		result = LegacyIssueListPage{Items: rows, HasMore: more}
		return nil
	})
	return result, err
}

func (s *Store) canonicalIssuePathInTx(ctx context.Context, tx *sql.Tx, nativeID string) (string, error) {
	issueType, err := installedIssueTypeURLInTx(ctx, tx, s.ScopeURL())
	if err != nil {
		return "", err
	}
	var path, kind, typ, state string
	if err := tx.QueryRowContext(ctx, `SELECT path,resource_kind,type_url,allocation_state FROM graph_preview_catalog WHERE backing='issue' AND backing_key=?`, nativeID).Scan(&path, &kind, &typ, &state); err != nil {
		return "", fmt.Errorf("%w: unmapped Issue list parent: %v", ErrInvalidStore, err)
	}
	if kind != "bead" || typ != issueType || state != "live" || validatePath(path) != nil {
		return "", fmt.Errorf("%w: invalid Issue list parent allocation", ErrInvalidStore)
	}
	return path, nil
}
