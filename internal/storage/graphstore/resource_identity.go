package graphstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	graph "github.com/steveyegge/beads/graphops"
)

// AllocatedResourceIdentity is internal allocation data, not a protocol record.
// PrivatePath is a private storage locator; no caller may emit it as the public ID.
// Kind comes from the catalog, never from PublicID's path segments. State may
// be deleted: the Scope reserves this identity for the allocation's lifetime.
type AllocatedResourceIdentity struct {
	PublicID    string             `json:"-"`
	PrivatePath string             `json:"-"`
	Kind        graph.ResourceKind `json:"-"`
	State       string             `json:"-"`
}

// ResolveAllocatedResourceIdentity looks up an exact literal public ID across
// Resource kinds, including deleted allocations. It does not accept aliases,
// infer kind from a prefix, or change current/retained record projection.
func (s *Store) ResolveAllocatedResourceIdentity(ctx context.Context, id string) (AllocatedResourceIdentity, error) {
	if err := validatePublicID(id); err != nil {
		return AllocatedResourceIdentity{}, err
	}
	return s.readAllocatedResourceIdentity(ctx, func(tx *sql.Tx) (AllocatedResourceIdentity, error) {
		return resolveAllocatedResourceIdentityInTx(ctx, tx, id)
	})
}

// AllocatedResourceIdentityForPath acquires public allocation identity for a
// private locator. Full-record and endpoint projection can share this fact
// without reconstructing public identity from the private path's spelling.
func (s *Store) AllocatedResourceIdentityForPath(ctx context.Context, path string) (AllocatedResourceIdentity, error) {
	if validatePath(path) != nil && validateLinkPath(path) != nil {
		return AllocatedResourceIdentity{}, fmt.Errorf("%w: invalid private allocation locator", graph.ErrValidation)
	}
	return s.readAllocatedResourceIdentity(ctx, func(tx *sql.Tx) (AllocatedResourceIdentity, error) {
		return allocatedResourceIdentityForPathInTx(ctx, tx, path)
	})
}

func (s *Store) readAllocatedResourceIdentity(ctx context.Context, lookup func(*sql.Tx) (AllocatedResourceIdentity, error)) (AllocatedResourceIdentity, error) {
	var identity AllocatedResourceIdentity
	err := s.withTx(ctx, false, func(tx *sql.Tx) error {
		if err := checkBinding(ctx, tx, s.options); err != nil {
			return err
		}
		if err := checkCurrentReadBytes(ctx, tx); err != nil {
			return err
		}
		var err error
		identity, err = lookup(tx)
		return err
	})
	if err != nil {
		return AllocatedResourceIdentity{}, err
	}
	return identity, nil
}

// ResolveCanonicalPublicResourceIdentity resolves a safe canonical public URI through
// Bead/Link allocations only. It is not a complete public route: registered Type
// and navigation collision checks belong to the subsequent common switch. This
// preparatory API changes neither the rooted protocol readers nor snapshots.
func (s *Store) ResolveCanonicalPublicResourceIdentity(ctx context.Context, uri string) (AllocatedResourceIdentity, error) {
	id, err := graph.ScopeLocalPublicResourceID(s.ScopeURL(), uri)
	if err != nil {
		return AllocatedResourceIdentity{}, err
	}
	return s.ResolveAllocatedResourceIdentity(ctx, id)
}

const allocatedIdentityColumns = `path,resource_kind,allocation_state,backing,
 CASE WHEN backing='issue' THEN backing_key ELSE COALESCE(public_id,path) END`

// These transaction helpers require the caller's usual binding and acquisition
// checks. They do not allocate, rewrite identities, or choose URI encodings.
func resolveAllocatedResourceIdentityInTx(ctx context.Context, tx *sql.Tx, id string) (AllocatedResourceIdentity, error) {
	rows, err := tx.QueryContext(ctx, `SELECT `+allocatedIdentityColumns+` FROM graph_preview_catalog
 WHERE (backing='issue' AND backing_key=?) OR
 (backing IN ('generic','dependency','informational') AND COALESCE(public_id,path)=?) LIMIT 2`, id, id)
	if err != nil {
		return AllocatedResourceIdentity{}, err
	}
	identity, err := readAllocatedIdentityRows(rows)
	if err == nil && identity.PublicID != id {
		return AllocatedResourceIdentity{}, fmt.Errorf("%w: public identity lookup was not byte-exact", ErrInvalidStore)
	}
	return identity, err
}

func allocatedResourceIdentityForPathInTx(ctx context.Context, tx *sql.Tx, path string) (AllocatedResourceIdentity, error) {
	rows, err := tx.QueryContext(ctx, `SELECT `+allocatedIdentityColumns+` FROM graph_preview_catalog WHERE path=? LIMIT 2`, path)
	if err != nil {
		return AllocatedResourceIdentity{}, err
	}
	identity, err := readAllocatedIdentityRows(rows)
	if err == nil && identity.PrivatePath != path {
		return AllocatedResourceIdentity{}, fmt.Errorf("%w: private allocation lookup was not byte-exact", ErrInvalidStore)
	}
	return identity, err
}

func readAllocatedIdentityRows(rows *sql.Rows) (AllocatedResourceIdentity, error) {
	var result AllocatedResourceIdentity
	for rows.Next() {
		if result.PrivatePath != "" {
			_ = rows.Close()
			return AllocatedResourceIdentity{}, fmt.Errorf("%w: identity lookup has multiple allocations", ErrInvalidStore)
		}
		var backing string
		var publicID sql.NullString
		if err := rows.Scan(&result.PrivatePath, &result.Kind, &result.State, &backing, &publicID); err != nil {
			_ = rows.Close()
			return AllocatedResourceIdentity{}, err
		}
		expectedKind := graph.KindBead
		switch backing {
		case "issue", "generic":
		case "dependency", "informational":
			expectedKind = graph.KindLink
		default:
			_ = rows.Close()
			return AllocatedResourceIdentity{}, fmt.Errorf("%w: unsupported identity backing", ErrInvalidStore)
		}
		invalidLocator := expectedKind == graph.KindBead && validatePath(result.PrivatePath) != nil || expectedKind == graph.KindLink && validateLinkPath(result.PrivatePath) != nil
		if invalidLocator || result.Kind != expectedKind || (result.State != "live" && result.State != "deleted") || !publicID.Valid || validatePublicID(publicID.String) != nil {
			_ = rows.Close()
			return AllocatedResourceIdentity{}, fmt.Errorf("%w: allocation identity/kind/state disagrees with its backing", ErrInvalidStore)
		}
		result.PublicID = publicID.String
	}
	if err := errors.Join(rows.Err(), rows.Close()); err != nil {
		return AllocatedResourceIdentity{}, err
	}
	if result.PrivatePath == "" {
		return AllocatedResourceIdentity{}, ErrNotFound
	}
	return result, nil
}
