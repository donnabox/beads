package graphstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/steveyegge/beads/internal/storage/issueops"
	"github.com/steveyegge/beads/internal/types"
)

// SearchIssueRows retains the native filter/count projection, delegating
// without selectors. Selected14 and legacy11 preserve their own admission.
func (s *Store) SearchIssueRows(ctx context.Context, query string, filter types.IssueFilter) ([]*types.IssueWithCounts, error) {
	rows, _, err := s.SearchIssueRowsWithSelectors(ctx, query, filter, nil)
	return rows, err
}

// SearchIssueRowsWithSelectors resolves every selector and hydrates native
// IssueWithCounts rows in ONE checked transaction. Paths retain input order and
// duplicates. No Resource-document-to-Issue reconstruction or second lookup is
// performed by the caller. Every14 failure returns nil rows and nil paths.
func (s *Store) SearchIssueRowsWithSelectors(ctx context.Context, query string, filter types.IssueFilter, selectors []IssueSelector) ([]*types.IssueWithCounts, []string, error) {
	if !filter.SkipWisps || (filter.Ephemeral != nil && *filter.Ephemeral) {
		return nil, nil, fmt.Errorf("%w: graph Issue discovery requires durable scope", ErrCapabilityUnavailable)
	}
	selected14 := s.LayoutVersion() == TypesFreshSchemaVersion
	if selected14 && (len(selectors) > PreviewSnapshotLimit || len(query) > PreviewIssueListRequestByteLimit) {
		return nil, nil, ErrLimitExceeded
	}
	for _, selector := range selectors {
		if err := validateIssueSelector(selector); err != nil {
			return nil, nil, err
		}
	}
	filter.Limit = PreviewSnapshotLimit
	var rows []*types.IssueWithCounts
	var paths []string
	var err error
	if selected14 {
		err = s.withDefinitionReadTx(ctx, txIssueDiscovery, func(tx *sql.Tx, budget *intrinsicMetadataBudget) error {
			if e := checkResourceQueryReadBytesInTx(ctx, tx); e != nil {
				return e
			}
			if e := chargeIssueDiscoveryPolicyInTx(ctx, tx, budget); e != nil {
				return e
			}
			if e := budget.chargeBody(128 + 4*len(query) + 32*len(selectors)); e != nil {
				return e
			}
			capture, e := s.captureTypeRecordsInTx(ctx, tx, budget)
			if e != nil {
				return e
			}
			readctx := s.contextForResourceCapture(ctx, tx, capture, budget)
			readctx = context.WithValue(readctx, issueDiscoveryReadKey{}, &issueDiscoveryRead{store: s, tx: tx, budget: budget, capture: capture})
			rows, paths, e = s.searchIssueRowsWithSelectorsInTx(readctx, tx, query, filter, selectors, budget)
			return e
		})
	} else {
		err = s.withTx(ctx, false, func(tx *sql.Tx) error {
			var e error
			rows, paths, e = s.searchIssueRowsWithSelectorsInTx(ctx, tx, query, filter, selectors, nil)
			return e
		})
	}
	if selected14 && ctx != nil {
		err = errors.Join(err, ctx.Err())
	}
	if err != nil {
		if selected14 || len(selectors) > 0 {
			return nil, nil, err
		}
		// Preserve the historical11 SearchIssueRows finalization return shape.
		return rows, nil, err
	}
	return rows, paths, nil
}

func (s *Store) searchIssueRowsWithSelectorsInTx(ctx context.Context, tx *sql.Tx, query string, filter types.IssueFilter, selectors []IssueSelector, budget *intrinsicMetadataBudget) ([]*types.IssueWithCounts, []string, error) {
	if budget == nil {
		if err := checkCurrentReadBytes(ctx, tx); err != nil {
			return nil, nil, err
		}
		if err := checkBinding(ctx, tx, s.options); err != nil {
			return nil, nil, err
		}
	}
	if err := s.checkCollectionMappingsInTx(ctx, tx); err != nil {
		return nil, nil, err
	}
	var count int
	if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM graph_preview_catalog WHERE allocation_state='live'").Scan(&count); err != nil {
		return nil, nil, err
	}
	if count > PreviewSnapshotLimit {
		return nil, nil, fmt.Errorf("%w: graph Issue discovery exceeds %d live Resources", ErrLimitExceeded, PreviewSnapshotLimit)
	}
	if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM graph_preview_catalog WHERE allocation_state NOT IN ('live','deleted')").Scan(&count); err != nil {
		return nil, nil, err
	}
	if count != 0 {
		return nil, nil, fmt.Errorf("%w: unknown allocation state", ErrInvalidStore)
	}
	if err := s.checkIssueListCatalogInTx(ctx, tx); err != nil {
		return nil, nil, err
	}
	paths, err := s.resolveIssuePathsInTx(ctx, tx, selectors)
	if err != nil {
		return nil, nil, err
	}
	if budget != nil {
		if err := chargeCommonNativeRowsInTx(ctx, tx, budget); err != nil {
			return nil, nil, err
		}
	}
	rows, err := issueops.SearchIssuesWithCountsInTx(ctx, tx, query, filter)
	if err != nil {
		var capError *issueops.ErrTooManyRows
		if errors.As(err, &capError) {
			return nil, nil, fmt.Errorf("%w: %w", ErrLimitExceeded, err)
		}
		return nil, nil, err
	}
	for _, row := range rows {
		if budget != nil {
			if err := ctx.Err(); err != nil {
				return nil, nil, err
			}
		}
		if row == nil || row.Issue == nil {
			return nil, nil, fmt.Errorf("%w: nil Issue discovery row", ErrInvalidStore)
		}
		var path string
		if err := tx.QueryRowContext(ctx, "SELECT path FROM graph_preview_catalog WHERE backing='issue' AND backing_key=? AND allocation_state='live'", row.ID).Scan(&path); err != nil {
			return nil, nil, fmt.Errorf("%w: unmapped discovered Issue: %v", ErrInvalidStore, err)
		}
		record, err := s.showIssueInTx(ctx, tx, path)
		if err != nil {
			return nil, nil, err
		}
		row.ID = strings.TrimPrefix(record.ID, s.ScopeURL())
		if row.Parent != nil {
			var parentPath string
			if err := tx.QueryRowContext(ctx, "SELECT path FROM graph_preview_catalog WHERE backing='issue' AND backing_key=? AND allocation_state='live'", *row.Parent).Scan(&parentPath); err != nil {
				return nil, nil, fmt.Errorf("%w: unmapped discovered parent Issue: %v", ErrInvalidStore, err)
			}
			if err := validatePath(parentPath); err != nil {
				return nil, nil, fmt.Errorf("%w: malformed discovered parent Issue", ErrInvalidStore)
			}
			parent, err := s.showIssueInTx(ctx, tx, parentPath)
			if err != nil {
				return nil, nil, err
			}
			*row.Parent = strings.TrimPrefix(parent.ID, s.ScopeURL())
		}
	}
	if budget != nil {
		if err := ctx.Err(); err != nil {
			return nil, nil, err
		}
	}
	if rows == nil {
		rows = []*types.IssueWithCounts{}
	}
	return rows, paths, nil
}
