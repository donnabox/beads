package graphstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/steveyegge/beads/internal/storage"
)

// A public ID is reserved across Issues, Memories, and Links for the Scope's
// lifetime, including deleted allocations. Issue
// backing keys are public IDs; graph paths are private addresses. The current
// Memory and Link allocations publish their full allocated path as public ID.
func publicResourceIDExistsInTx(ctx context.Context, tx *sql.Tx, id string) (bool, error) {
	var exists int
	err := tx.QueryRowContext(ctx, `SELECT 1 FROM graph_preview_catalog
		WHERE (
			(backing='issue' AND backing_key=?) OR
			(backing IN ('generic','dependency','informational') AND path=?)
		) LIMIT 1`, id, id).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

// ResolveAllocatedResourcePath finds a public ID even after its Resource has
// been deleted. Historical update needs the durable allocation path before
// its checked writer can compare the terminal revision and restore content.
// An exact public ID wins over old Bead/Link suffix shorthand.
func (s *Store) ResolveAllocatedResourcePath(ctx context.Context, id string) (string, error) {
	return s.resolveAllocatedResourcePath(ctx, id, true)
}

// ResolveExactAllocatedResourcePath excludes compatibility suffixes for a
// write that will commit in a later transaction. The returned allocation is
// stable for this Scope's lifetime, including across deletion and recovery.
func (s *Store) ResolveExactAllocatedResourcePath(ctx context.Context, id string) (string, error) {
	return s.resolveAllocatedResourcePath(ctx, id, false)
}

func (s *Store) resolveAllocatedResourcePath(ctx context.Context, id string, allowShorthand bool) (string, error) {
	if id == "" || validatePath("beads/"+id) != nil || validateLinkPath("links/"+id) != nil {
		return "", fmt.Errorf("%w: invalid local Resource ID", storage.ErrValidation)
	}
	var selected string
	err := s.withTx(ctx, false, func(tx *sql.Tx) error {
		if err := checkCurrentReadBytes(ctx, tx); err != nil {
			return err
		}
		if err := checkBinding(ctx, tx, s.options); err != nil {
			return err
		}
		rows, err := tx.QueryContext(ctx, `SELECT path FROM graph_preview_catalog WHERE
			(backing='issue' AND backing_key=?) OR
			(backing IN ('generic','dependency','informational') AND path=?) LIMIT 2`, id, id)
		if err != nil {
			return err
		}
		for rows.Next() {
			var path string
			if err := rows.Scan(&path); err != nil {
				_ = rows.Close()
				return err
			}
			if selected != "" {
				_ = rows.Close()
				return fmt.Errorf("%w: public ID has multiple graph allocations", ErrInvalidStore)
			}
			selected = path
		}
		if err := errors.Join(rows.Err(), rows.Close()); err != nil {
			return err
		}
		if selected != "" {
			return nil
		}
		if !allowShorthand {
			return ErrNotFound
		}
		// Suffixes are compatibility syntax, never public identity. A native
		// Issue's private path cannot answer an unqualified local selector.
		for _, candidate := range []struct{ path, backing string }{{"beads/" + id, "generic"}, {"links/" + id, "link"}} {
			var backing string
			err := tx.QueryRowContext(ctx, `SELECT backing FROM graph_preview_catalog WHERE path=?`, candidate.path).Scan(&backing)
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			if err != nil {
				return err
			}
			if candidate.backing == "generic" && backing != "generic" ||
				candidate.backing == "link" && backing != "dependency" && backing != "informational" {
				continue
			}
			if selected != "" {
				return fmt.Errorf("%w: local shorthand identifies multiple Resources; use a public ID or canonical Scope URL", storage.ErrValidation)
			}
			selected = candidate.path
		}
		if selected == "" {
			return ErrNotFound
		}
		return nil
	})
	return selected, err
}
