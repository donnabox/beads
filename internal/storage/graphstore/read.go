package graphstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// Read preserves whole-workspace acquisition and binding checks, but exposes
// only the Memory backing in this C0 landing.
func (s *Store) Read(ctx context.Context, path string) (any, error) {
	if err := validatePath(path); err != nil {
		return nil, err
	}
	var result any
	err := s.withTx(ctx, false, func(tx *sql.Tx) error {
		if err := checkCurrentReadBytes(ctx, tx); err != nil {
			return err
		}
		if err := checkBinding(ctx, tx, s.options); err != nil {
			return err
		}
		var backing string
		if err := tx.QueryRowContext(ctx, `SELECT backing FROM graph_preview_catalog WHERE path=?`, path).Scan(&backing); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrNotFound
			}
			return err
		}
		var err error
		switch backing {
		case "generic":
			result, err = s.showMemoryInTx(ctx, tx, path)
		default:
			err = fmt.Errorf("%w: unsupported backing", ErrCapabilityUnavailable)
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
