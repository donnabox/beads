package graphstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/steveyegge/beads/internal/storage"
)

// IssueSelector keeps the explicit graph path separate from a bare native
// Issue ID. NativeID is set only for bare CLI spelling, never for beads/PATH
// or a canonical URL.
type IssueSelector struct {
	Path     string
	NativeID string
}

func validateIssueSelector(selector IssueSelector) error {
	if err := validatePath(selector.Path); err != nil {
		return err
	}
	if selector.NativeID != "" && selector.Path != "beads/"+selector.NativeID {
		return fmt.Errorf("%w: native Issue fallback requires its matching bare selector", storage.ErrValidation)
	}
	return nil
}

// Resolve against the checked live catalog in the caller's transaction. A
// bare spelling that also names a different graph path is ambiguous, even if
// that path is a Memory or a retained deletion. Explicit paths remain exact.
func (s *Store) resolveLiveIssueSelectorInTx(ctx context.Context, tx *sql.Tx, selector IssueSelector) (string, IssueRecord, error) {
	if selector.NativeID == "" {
		issue, err := s.requireIssueInTx(ctx, tx, selector.Path)
		return selector.Path, issue, err
	}
	var exists int
	err := tx.QueryRowContext(ctx, `SELECT 1 FROM graph_preview_catalog WHERE path=?`, selector.Path).Scan(&exists)
	pathExists := err == nil
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", IssueRecord{}, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT path FROM graph_preview_catalog WHERE backing='issue' AND backing_key=? AND allocation_state='live' LIMIT 2`, selector.NativeID)
	if err != nil {
		return "", IssueRecord{}, err
	}
	var paths []string
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			_ = rows.Close()
			return "", IssueRecord{}, err
		}
		paths = append(paths, path)
	}
	if err := errors.Join(rows.Err(), rows.Close()); err != nil {
		return "", IssueRecord{}, err
	}
	if len(paths) > 1 {
		return "", IssueRecord{}, fmt.Errorf("%w: native Issue ID has multiple live graph allocations", ErrInvalidStore)
	}
	if len(paths) == 1 && pathExists && paths[0] != selector.Path {
		return "", IssueRecord{}, fmt.Errorf("%w: bare Issue ID %q matches both a graph path and a different native Issue; use an explicit beads/PATH", storage.ErrValidation, selector.NativeID)
	}
	path := selector.Path
	if !pathExists {
		if len(paths) == 0 {
			return "", IssueRecord{}, ErrNotFound
		}
		path = paths[0]
		if err := validatePath(path); err != nil {
			return "", IssueRecord{}, fmt.Errorf("%w: invalid native Issue graph allocation: %v", ErrInvalidStore, err)
		}
	}
	issue, err := s.requireIssueInTx(ctx, tx, path)
	return path, issue, err
}

func (s *Store) ShowIssueSelected(ctx context.Context, selector IssueSelector) (IssueRecord, error) {
	if err := validateIssueSelector(selector); err != nil {
		return IssueRecord{}, err
	}
	var result IssueRecord
	err := s.withTx(ctx, false, func(tx *sql.Tx) error {
		if err := checkBinding(ctx, tx, s.options); err != nil {
			return err
		}
		_, issue, err := s.resolveLiveIssueSelectorInTx(ctx, tx, selector)
		result = issue
		return err
	})
	if err != nil {
		return IssueRecord{}, err
	}
	return result, nil
}
