package graphstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	graph "github.com/steveyegge/beads/graphops"
)

// AllocatedResourceIdentity is internal allocation data, not a protocol record.
// Path is a private storage locator; no caller may emit it as the public ID.
// Kind comes from the catalog, never from PublicID's path segments. State may
// be deleted: the Scope reserves this identity for the allocation's lifetime.
type AllocatedResourceIdentity struct {
	PublicID string             `json:"-"`
	Path     string             `json:"-"`
	Kind     graph.ResourceKind `json:"-"`
	State    string             `json:"-"`
}

// ResolveAllocatedResourceIdentity looks up an exact literal public ID across
// Resource kinds, including deleted allocations. It does not accept aliases,
// infer kind from a prefix, or change current/retained record projection.
func (s *Store) ResolveAllocatedResourceIdentity(ctx context.Context, id string) (AllocatedResourceIdentity, error) {
	if err := validatePublicID(id); err != nil {
		return AllocatedResourceIdentity{}, err
	}
	var identity AllocatedResourceIdentity
	err := s.withTx(ctx, false, func(tx *sql.Tx) error {
		if err := checkBinding(ctx, tx, s.options); err != nil {
			return err
		}
		if err := checkCurrentReadBytes(ctx, tx); err != nil {
			return err
		}
		var err error
		identity, err = resolveAllocatedResourceIdentityInTx(ctx, tx, id)
		return err
	})
	if err != nil {
		return AllocatedResourceIdentity{}, err
	}
	return identity, nil
}

// ResolveCanonicalResourceIdentity resolves a safe canonical public URI through
// the allocation catalog. This preparatory API is intentionally separate from
// the existing rooted protocol readers; adopting it does not rewrite snapshots.
func (s *Store) ResolveCanonicalResourceIdentity(ctx context.Context, uri string) (AllocatedResourceIdentity, error) {
	id, err := graph.ScopeLocalResourceID(s.ScopeURL(), uri)
	if err != nil {
		return AllocatedResourceIdentity{}, err
	}
	return s.ResolveAllocatedResourceIdentity(ctx, id)
}

// Callers inside checked writer/import transactions may use this exact lookup
// after their usual binding and byte-budget checks. It performs no allocation
// or canonical URI conversion and leaves unsupported legacy keys unchanged.
func resolveAllocatedResourceIdentityInTx(ctx context.Context, tx *sql.Tx, id string) (AllocatedResourceIdentity, error) {
	rows, err := tx.QueryContext(ctx, `SELECT path,resource_kind,allocation_state,backing
 FROM graph_preview_catalog WHERE (backing='issue' AND backing_key=?) OR
 (backing IN ('generic','dependency','informational') AND COALESCE(public_id,path)=?) LIMIT 2`, id, id)
	if err != nil {
		return AllocatedResourceIdentity{}, err
	}
	var result AllocatedResourceIdentity
	for rows.Next() {
		if result.Path != "" {
			_ = rows.Close()
			return AllocatedResourceIdentity{}, fmt.Errorf("%w: public ID has multiple allocations", ErrInvalidStore)
		}
		var backing string
		if err := rows.Scan(&result.Path, &result.Kind, &result.State, &backing); err != nil {
			_ = rows.Close()
			return AllocatedResourceIdentity{}, err
		}
		if (result.State != "live" && result.State != "deleted") ||
			((backing == "issue" || backing == "generic") && result.Kind != graph.KindBead) ||
			((backing == "dependency" || backing == "informational") && result.Kind != graph.KindLink) {
			_ = rows.Close()
			return AllocatedResourceIdentity{}, fmt.Errorf("%w: allocation kind/state disagrees with its backing", ErrInvalidStore)
		}
	}
	if err := errors.Join(rows.Err(), rows.Close()); err != nil {
		return AllocatedResourceIdentity{}, err
	}
	if result.Path == "" {
		return AllocatedResourceIdentity{}, ErrNotFound
	}
	result.PublicID = id
	return result, nil
}
