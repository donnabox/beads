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

// A non-Issue selection has a different refusal from a bare name that
// ambiguously identifies an Issue and another graph Bead.
var errNonIssueSelector = errors.New("selector is not an Issue")

func nonIssueSelectorError(backing string) error {
	return fmt.Errorf("%w: selector names %s, not an Issue: %w", storage.ErrValidation, backing, errNonIssueSelector)
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
func (s *Store) resolveLiveIssueSelectorInTx(ctx context.Context, tx *sql.Tx, selector IssueSelector) (string, error) {
	if selector.NativeID == "" {
		var backing string
		err := tx.QueryRowContext(ctx, `SELECT backing FROM graph_preview_catalog WHERE path=?`, selector.Path).Scan(&backing)
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotFound
		}
		if err != nil {
			return "", err
		}
		if backing != "issue" {
			return "", nonIssueSelectorError(backing)
		}
		_, err = s.requireIssueInTx(ctx, tx, selector.Path)
		return selector.Path, err
	}
	var backing string
	err := tx.QueryRowContext(ctx, `SELECT backing FROM graph_preview_catalog WHERE path=?`, selector.Path).Scan(&backing)
	pathExists := err == nil
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	rows, err := tx.QueryContext(ctx, `SELECT path FROM graph_preview_catalog WHERE backing='issue' AND backing_key=? AND allocation_state='live' LIMIT 2`, selector.NativeID)
	if err != nil {
		return "", err
	}
	var paths []string
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			_ = rows.Close()
			return "", err
		}
		paths = append(paths, path)
	}
	if err := errors.Join(rows.Err(), rows.Close()); err != nil {
		return "", err
	}
	if len(paths) > 1 {
		return "", fmt.Errorf("%w: native Issue ID has multiple live graph allocations", ErrInvalidStore)
	}
	if len(paths) == 1 && pathExists && paths[0] != selector.Path {
		return "", fmt.Errorf("%w: bare Issue ID %q matches both a graph path and a different native Issue; use an explicit beads/PATH", storage.ErrValidation, selector.NativeID)
	}
	path := selector.Path
	if !pathExists {
		if len(paths) == 0 {
			return "", ErrNotFound
		}
		path = paths[0]
		if err := validatePath(path); err != nil {
			return "", fmt.Errorf("%w: invalid native Issue graph allocation: %v", ErrInvalidStore, err)
		}
	}
	if pathExists && backing != "issue" {
		return "", nonIssueSelectorError(backing)
	}
	_, err = s.requireIssueInTx(ctx, tx, path)
	return path, err
}

// ResolveIssuePaths resolves count filters against the checked Issue catalog.
// Missing and non-Issue IDs retain their literal graph path so a count can
// return zero, as the ordinary command does. Ambiguous bare IDs fail.
func (s *Store) ResolveIssuePaths(ctx context.Context, selectors []IssueSelector) ([]string, error) {
	if len(selectors) == 0 {
		return []string{}, nil
	}
	for _, selector := range selectors {
		if err := validateIssueSelector(selector); err != nil {
			return nil, err
		}
	}
	paths := make([]string, len(selectors))
	err := s.withTx(ctx, false, func(tx *sql.Tx) error {
		if err := checkBinding(ctx, tx, s.options); err != nil {
			return err
		}
		for i, selector := range selectors {
			path, err := s.resolveLiveIssueSelectorInTx(ctx, tx, selector)
			switch {
			case errors.Is(err, ErrNotFound):
				paths[i] = selector.Path
			case errors.Is(err, errNonIssueSelector):
				paths[i] = selector.Path
			case err != nil:
				return err
			default:
				paths[i] = path
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return paths, nil
}
