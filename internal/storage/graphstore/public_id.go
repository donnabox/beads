package graphstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	graph "github.com/steveyegge/beads/graphops"
	"strings"
	"unicode/utf8"

	"github.com/steveyegge/beads/internal/storage"
)

// A public ID is reserved across Issues, Memories, and Links for the Scope's
// lifetime, including deleted allocations. Native Issue backing keys and explicit
// Memory public_id values are identities; legacy NULL Memory/Link allocations
// continue to publish their original path. The writer_token coordination fence
// serializes check-and-insert reservations; no suffix alias is reserved.
func publicResourceIDExistsInTx(ctx context.Context, tx *sql.Tx, id string) (bool, error) {
	var exists int
	err := tx.QueryRowContext(ctx, `SELECT 1 FROM graph_preview_catalog
		WHERE (
			(backing='issue' AND backing_key=?) OR
			(backing IN ('generic','dependency','informational') AND COALESCE(public_id,path)=?)
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
	if err := validatePublicID(id); err != nil {
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
			(backing IN ('generic','dependency','informational') AND COALESCE(public_id,path)=?) LIMIT 2`, id, id)
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
			err := tx.QueryRowContext(ctx, `SELECT backing FROM graph_preview_catalog WHERE path=? AND (backing<>'generic' OR public_id IS NULL)`, candidate.path).Scan(&backing)
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

// validatePublicID deliberately imposes no path syntax, case folding or trim.
// Existing acquisition budgets bound bytes before materializing stored IDs.
func validatePublicID(id string) error {
	if !utf8.ValidString(id) || strings.TrimSpace(id) == "" {
		return fmt.Errorf("%w: public ID must be nonblank UTF-8", storage.ErrValidation)
	}
	return nil
}

// PrepareMemoryIdentity prepares a private locator for an explicit public ID.
// It does not reserve the identity: Create/createInTx reserves atomically with
// payload and history. Importers can prepare requests before their batch tx.
func PrepareMemoryIdentity(req CreateRequest) (CreateRequest, error) {
	if req.PublicID == "" {
		return req, nil
	}
	if err := validatePublicID(req.PublicID); err != nil {
		return CreateRequest{}, err
	}
	if req.Path == "" {
		// Preserve the established canonical layout when it is a valid private
		// address. The public bytes are independent and never normalized.
		req.Path = req.PublicID
		if validatePath(req.Path) != nil {
			req.Path = "beads/" + req.PublicID
		}
		if validatePath(req.Path) != nil {
			token, err := freshToken()
			if err != nil {
				return CreateRequest{}, err
			}
			req.Path = "beads/" + token
		}
		req.privatePathGenerated = true
	}

	return req, nil
}

func memoryPublicIDInTx(ctx context.Context, tx *sql.Tx, path string) (string, error) {
	var size uint64
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(OCTET_LENGTH(public_id),0) FROM graph_preview_catalog WHERE path=?`, path).Scan(&size); err != nil {
		return "", fmt.Errorf("%w: public ID allocation size: %v", ErrInvalidStore, err)
	}
	if size > PreviewCurrentReadByteLimit {
		return "", ErrLimitExceeded
	}
	var id sql.NullString
	if err := tx.QueryRowContext(ctx, `SELECT public_id FROM graph_preview_catalog WHERE path=?`, path).Scan(&id); err != nil {
		return "", fmt.Errorf("%w: public ID allocation is missing: %v", ErrInvalidStore, err)
	}
	if !id.Valid {
		return "", nil
	}
	if err := validatePublicID(id.String); err != nil {
		return "", fmt.Errorf("%w: invalid allocation public ID", ErrInvalidStore)
	}
	return id.String, nil
}

// ResolveMemorySelector separates literal public identity from canonical locator
// syntax. If both interpretations name different allocations, it refuses rather
// than guessing. Returned allocation paths never change or become reusable.
func (s *Store) ResolveMemorySelector(ctx context.Context, selector string) (string, error) {
	var selected string
	err := s.withTx(ctx, false, func(tx *sql.Tx) error {
		if err := checkBinding(ctx, tx, s.options); err != nil {
			return err
		}
		if err := checkCurrentReadBytes(ctx, tx); err != nil {
			return err
		}
		if err := validatePublicID(selector); err != nil {
			return err
		}
		var backing string
		rows, err := tx.QueryContext(ctx, `SELECT path,backing FROM graph_preview_catalog WHERE
 (backing='issue' AND backing_key=?) OR
 (backing IN ('generic','dependency','informational') AND COALESCE(public_id,path)=?) LIMIT 2`, selector, selector)
		if err != nil {
			return err
		}
		for rows.Next() {
			if selected != "" {
				_ = rows.Close()
				return fmt.Errorf("%w: public ID has multiple allocations", ErrInvalidStore)
			}
			if err := rows.Scan(&selected, &backing); err != nil {
				_ = rows.Close()
				return err
			}
		}
		if err := errors.Join(rows.Err(), rows.Close()); err != nil {
			return err
		}

		locator, kind, canonical := graph.SplitCanonicalURL(s.ScopeURL(), selector)
		if canonical && kind == graph.KindBead {
			var locatorBacking string
			e := tx.QueryRowContext(ctx, `SELECT backing FROM graph_preview_catalog WHERE path=?`, locator).Scan(&locatorBacking)
			if e != nil && !errors.Is(e, sql.ErrNoRows) {
				return e
			}
			if e == nil {
				if selected != "" && selected != locator {
					return fmt.Errorf("%w: selector names both a public ID and a different canonical locator", storage.ErrValidation)
				}
				selected, backing = locator, locatorBacking
			}
		}
		if selected == "" {
			return ErrNotFound
		}
		if backing != "generic" {
			return fmt.Errorf("%w: selected public ID is not a Memory", ErrCapabilityUnavailable)
		}
		return nil
	})
	return selected, err
}
