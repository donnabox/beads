package graphstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/steveyegge/beads/internal/storage"
)

// IssueSelector keeps an exact canonical graph path separate from an opaque
// local native Issue ID. NativeID is set for any local spelling, including
// strings that happen to contain beads/; canonical URLs select paths only.
type IssueSelector struct {
	Path     string
	NativeID string
}

// A non-Issue selection has a different user-visible state-read refusal from
// a bare name that ambiguously identifies an Issue and another graph Bead.
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
// local spelling may name a native Issue ID or a graph path retained by the
// current-v0 preview. If those identify different records, refuse rather than
// selecting one by its spelling. Canonical URLs remain exact paths.
func (s *Store) resolveLiveIssueSelectorInTx(ctx context.Context, tx *sql.Tx, selector IssueSelector) (string, IssueRecord, error) {
	if selector.NativeID == "" {
		var backing string
		err := tx.QueryRowContext(ctx, `SELECT backing FROM graph_preview_catalog WHERE path=?`, selector.Path).Scan(&backing)
		if errors.Is(err, sql.ErrNoRows) {
			return "", IssueRecord{}, ErrNotFound
		}
		if err != nil {
			return "", IssueRecord{}, err
		}
		if backing != "issue" {
			return "", IssueRecord{}, nonIssueSelectorError(backing)
		}
		issue, err := s.requireIssueInTx(ctx, tx, selector.Path)
		return selector.Path, issue, err
	}
	backingAt := func(path string) (string, bool, error) {
		var backing string
		err := tx.QueryRowContext(ctx, `SELECT backing FROM graph_preview_catalog WHERE path=?`, path).Scan(&backing)
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}
		return backing, err == nil, err
	}
	backing, pathExists, err := backingAt(selector.Path)
	if err != nil {
		return "", IssueRecord{}, err
	}
	legacyPath := ""
	legacyBacking := ""
	legacyExists := false
	if selector.NativeID != selector.Path && validatePath(selector.NativeID) == nil {
		legacyPath = selector.NativeID
		legacyBacking, legacyExists, err = backingAt(legacyPath)
		if err != nil {
			return "", IssueRecord{}, err
		}
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
	if pathExists && legacyExists {
		return "", IssueRecord{}, fmt.Errorf("%w: Issue ID %q matches two graph paths; use a canonical Scope URL", storage.ErrValidation, selector.NativeID)
	}
	if len(paths) == 1 && ((pathExists && paths[0] != selector.Path) || (legacyExists && paths[0] != legacyPath)) {
		return "", IssueRecord{}, fmt.Errorf("%w: Issue ID %q matches a graph path and a different native Issue; use a canonical Scope URL", storage.ErrValidation, selector.NativeID)
	}
	path := selector.Path
	if !pathExists {
		if len(paths) == 0 && !legacyExists {
			return "", IssueRecord{}, ErrNotFound
		}
		if len(paths) == 1 {
			path = paths[0]
		} else {
			path = legacyPath
			backing = legacyBacking
		}
		if err := validatePath(path); err != nil {
			return "", IssueRecord{}, fmt.Errorf("%w: invalid native Issue graph allocation: %v", ErrInvalidStore, err)
		}
	}
	if backing != "issue" && (pathExists || (legacyExists && len(paths) == 0)) {
		return "", IssueRecord{}, nonIssueSelectorError(backing)
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
		if errors.Is(err, errNonIssueSelector) {
			return IssueRecord{}, ErrCapabilityUnavailable
		}
		return IssueRecord{}, err
	}
	return result, nil
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
			path, _, err := s.resolveLiveIssueSelectorInTx(ctx, tx, selector)
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
