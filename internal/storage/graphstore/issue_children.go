package graphstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/steveyegge/beads/internal/types"
)

// ListChildren reads native immediate parent-child edges and returns complete
// mapped Issue records. Dotted IDs alone are not proof of parentage: the native
// dependency is authoritative. This read does not allocate a Link or change
// the existing preview's admitted dependency writer.
func (s *Store) ListChildren(ctx context.Context, parentPath string) ([]IssueRecord, error) {
	if err := validatePath(parentPath); err != nil {
		return nil, err
	}
	var children []IssueRecord
	err := s.withTx(ctx, false, func(tx *sql.Tx) error {
		if err := checkCurrentReadBytes(ctx, tx); err != nil {
			return err
		}
		if err := checkBinding(ctx, tx, s.options); err != nil {
			return err
		}
		parent, err := s.requireIssueInTx(ctx, tx, parentPath)
		if err != nil {
			return err
		}
		rows, err := tx.QueryContext(ctx, `SELECT c.path
			FROM dependencies d LEFT JOIN graph_preview_catalog c
			ON c.backing='issue' AND c.backing_key=d.issue_id AND c.allocation_state='live'
			WHERE d.depends_on_issue_id=? AND d.type=? ORDER BY c.path LIMIT ?`,
			parent.Properties.ID, string(types.DepParentChild), PreviewSnapshotLimit+1)
		if err != nil {
			return err
		}
		var paths []string
		for rows.Next() {
			var path sql.NullString
			if err := rows.Scan(&path); err != nil {
				_ = rows.Close()
				return err
			}
			if !path.Valid {
				_ = rows.Close()
				return fmt.Errorf("%w: unmapped child Issue", ErrInvalidStore)
			}
			paths = append(paths, path.String)
		}
		if err := errors.Join(rows.Err(), rows.Close()); err != nil {
			return err
		}
		if len(paths) > PreviewSnapshotLimit {
			return fmt.Errorf("%w: child Issue limit exceeded", ErrLimitExceeded)
		}
		children = make([]IssueRecord, 0, len(paths))
		for _, path := range paths {
			child, err := s.showIssueInTx(ctx, tx, path)
			if err != nil {
				return err
			}
			children = append(children, child)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return children, nil
}

// ListIssueDependencies reads only the installed blocking Dependency Type,
// checking the anchor as a native Issue in the same transaction as its Links.
func (s *Store) ListIssueDependencies(ctx context.Context, issuePath, direction string) ([]LinkRecord, error) {
	if err := validatePath(issuePath); err != nil {
		return nil, err
	}
	if direction != "in" && direction != "out" {
		return nil, fmt.Errorf("%w: dependency direction must be in or out", ErrCapabilityUnavailable)
	}
	var links []LinkRecord
	err := s.withTx(ctx, false, func(tx *sql.Tx) error {
		if err := checkBinding(ctx, tx, s.options); err != nil {
			return err
		}
		if _, err := s.requireIssueInTx(ctx, tx, issuePath); err != nil {
			return err
		}
		var err error
		links, err = s.incidentLinksInTx(ctx, tx, LinksRequest{
			BeadPath: issuePath, Direction: direction, TypeURL: DependencyTypeURL(s.ScopeURL()),
		}, "")
		return err
	})
	if err != nil {
		return nil, err
	}
	return links, nil
}
