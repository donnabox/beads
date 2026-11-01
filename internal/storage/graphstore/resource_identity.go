package graphstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

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
	return s.readAllocatedResourceIdentity(ctx, func(tx *sql.Tx, budget *intrinsicMetadataBudget) (AllocatedResourceIdentity, error) {
		return resolveAllocatedResourceIdentityInTx(ctx, tx, id, budget)
	})
}

// AllocatedResourceIdentityForPath acquires public allocation identity for a
// private locator. Full-record and endpoint projection can share this fact
// without reconstructing public identity from the private path's spelling.
func (s *Store) AllocatedResourceIdentityForPath(ctx context.Context, path string) (AllocatedResourceIdentity, error) {
	if validatePath(path) != nil && validateLinkPath(path) != nil {
		return AllocatedResourceIdentity{}, fmt.Errorf("%w: invalid private allocation locator", graph.ErrValidation)
	}
	return s.readAllocatedResourceIdentity(ctx, func(tx *sql.Tx, budget *intrinsicMetadataBudget) (AllocatedResourceIdentity, error) {
		return allocatedResourceIdentityForPathInTx(ctx, tx, path, budget)
	})
}

func (s *Store) readAllocatedResourceIdentity(ctx context.Context, lookup func(*sql.Tx, *intrinsicMetadataBudget) (AllocatedResourceIdentity, error)) (AllocatedResourceIdentity, error) {
	var identity AllocatedResourceIdentity
	var err error
	if s.options.Binding.SchemaVersion == TypesFreshSchemaVersion {
		err = s.withDefinitionReadTx(ctx, txDefinitionRead, func(tx *sql.Tx, budget *intrinsicMetadataBudget) error {
			var e error
			identity, e = lookup(tx, budget)
			return e
		})
	} else {
		err = s.withTx(ctx, false, func(tx *sql.Tx) error {
			if e := checkBinding(ctx, tx, s.options); e != nil {
				return e
			}
			var e error
			identity, e = lookup(tx, nil)
			return e
		})
	}
	if err != nil {
		return AllocatedResourceIdentity{}, err
	}
	return identity, nil
}

// ResolveCanonicalPublicResourceIdentity resolves an exact same-Scope public URI
// through the lifetime Bead/Link catalog. Registered Type and navigation URLs
// are separate authority facts, not inferred Resource-kind prefix families.
func (s *Store) ResolveCanonicalPublicResourceIdentity(ctx context.Context, uri string) (AllocatedResourceIdentity, error) {
	id, err := graph.ScopeLocalPublicResourceID(s.ScopeURL(), uri)
	if err != nil {
		return AllocatedResourceIdentity{}, err
	}
	return s.ResolveAllocatedResourceIdentity(ctx, id)
}

const allocatedIdentityColumns = `path,resource_kind,allocation_state,backing,
 CASE WHEN backing='issue' THEN backing_key ELSE COALESCE(public_id,path) END`

// publicResourceURLInTx is the single record/reference identity projection.
// Callers retain private locators for SQL only and acquire allocation bytes
// under their checked current/retained read or mutation budget.
func (s *Store) publicResourceURLInTx(ctx context.Context, tx *sql.Tx, path string) (string, error) {
	identity, err := allocatedResourceIdentityForPathInTx(ctx, tx, path)
	if err != nil {
		return "", err
	}
	return graph.CanonicalPublicResourceURL(s.ScopeURL(), identity.PublicID)
}

func (s *Store) publicResourcePathInTx(ctx context.Context, tx *sql.Tx, uri string, kind graph.ResourceKind) (string, error) {
	id, err := graph.ScopeLocalPublicResourceID(s.ScopeURL(), uri)
	if err != nil {
		return "", err
	}
	identity, err := resolveAllocatedResourceIdentityInTx(ctx, tx, id)
	if err != nil {
		return "", err
	}
	if identity.Kind != kind {
		return "", fmt.Errorf("%w: public identity is not a %s", graph.ErrValidation, kind)
	}
	return identity.PrivatePath, nil
}

// Charge only the selected allocation facts before acquiring their bytes.
// In particular, a scoped Memory question must not scan unrelated lifetime
// identities or inherit a whole-Scope retained-history acquisition gate.
func checkAllocatedIdentityReadBytes(ctx context.Context, tx *sql.Tx, predicate string, args ...any) error {
	return checkAllocatedIdentityBytesInBudget(ctx, tx, nil, predicate, args...)
}

func checkAllocatedIdentityBytesInBudget(ctx context.Context, tx *sql.Tx, budget *intrinsicMetadataBudget, predicate string, args ...any) error {
	var size uint64
	query := `SELECT CAST(LEAST(COALESCE(SUM(
 4*(256+OCTET_LENGTH(path)+COALESCE(OCTET_LENGTH(backing_key),0)+COALESCE(OCTET_LENGTH(public_id),0))
 ),0),?) AS UNSIGNED) FROM graph_preview_catalog WHERE ` + predicate
	parameters := append([]any{PreviewCurrentReadByteLimit + 1}, args...)
	err := tx.QueryRowContext(ctx, query, parameters...).Scan(&size)
	if err != nil {
		return err
	}
	if size > PreviewCurrentReadByteLimit {
		return fmt.Errorf("%w: allocation identities exceed acquisition budget", ErrLimitExceeded)
	}
	if budget != nil {
		return budget.chargeBody(int(size) + 256)
	}
	return nil
}

// Private locators remain explicit internal addresses. A complete canonical
// Scope URL is a disjoint public address and resolves only exact allocation
// identity; it never enables abbreviation or private suffix compatibility.
func (s *Store) validateResourceReadAddress(address string) error {
	if strings.HasPrefix(address, s.ScopeURL()) {
		_, err := graph.ScopeLocalPublicResourceID(s.ScopeURL(), address)
		return err
	}
	return validateResourcePath(address)
}

func (s *Store) resolveResourceReadAddressInTx(ctx context.Context, tx *sql.Tx, address string) (string, error) {
	if !strings.HasPrefix(address, s.ScopeURL()) {
		return address, nil
	}
	id, err := graph.ScopeLocalPublicResourceID(s.ScopeURL(), address)
	if err != nil {
		return "", err
	}
	identity, err := resolveAllocatedResourceIdentityInTx(ctx, tx, id)
	if err != nil {
		return "", err
	}
	return identity.PrivatePath, nil
}

// These transaction helpers require the caller's usual binding and acquisition
// checks. They do not allocate, rewrite identities, or choose URI encodings.
const allocatedPublicIdentityPredicate = `(backing='issue' AND backing_key=?) OR
 (backing IN ('generic','dependency','informational','intrinsic','intrinsic-link','type-definition') AND COALESCE(public_id,path)=?)`

func resolveAllocatedResourceIdentityInTx(ctx context.Context, tx *sql.Tx, id string, budgets ...*intrinsicMetadataBudget) (AllocatedResourceIdentity, error) {
	var budget *intrinsicMetadataBudget
	if len(budgets) > 1 {
		return AllocatedResourceIdentity{}, ErrInvalidStore
	}
	if len(budgets) == 1 {
		budget = budgets[0]
	}
	if err := checkAllocatedIdentityBytesInBudget(ctx, tx, budget, allocatedPublicIdentityPredicate, id, id); err != nil {
		return AllocatedResourceIdentity{}, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT `+allocatedIdentityColumns+` FROM graph_preview_catalog
 WHERE `+allocatedPublicIdentityPredicate+` LIMIT 2`, id, id)
	if err != nil {
		return AllocatedResourceIdentity{}, err
	}
	identity, err := readAllocatedIdentityRows(rows)
	if err == nil && identity.PublicID != id {
		return AllocatedResourceIdentity{}, fmt.Errorf("%w: public identity lookup was not byte-exact", ErrInvalidStore)
	}
	return identity, err
}

func allocatedResourceIdentityForPathInTx(ctx context.Context, tx *sql.Tx, path string, budgets ...*intrinsicMetadataBudget) (AllocatedResourceIdentity, error) {
	var budget *intrinsicMetadataBudget
	if len(budgets) > 1 {
		return AllocatedResourceIdentity{}, ErrInvalidStore
	}
	if len(budgets) == 1 {
		budget = budgets[0]
	}
	if err := checkAllocatedIdentityBytesInBudget(ctx, tx, budget, "path=?", path); err != nil {
		return AllocatedResourceIdentity{}, err
	}
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
		case "issue", "generic", "intrinsic", "type-definition":
		case "dependency", "informational", "intrinsic-link":
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

// Public canonical endpoints are exact catalog addresses. Compatibility
// operands keep their existing checked path/local-ID agreement rules.
func (s *Store) validateBeadEndpointOperand(path, localID string) error {
	if strings.HasPrefix(path, s.ScopeURL()) {
		if localID != "" {
			return fmt.Errorf("%w: canonical endpoint cannot also supply a local-ID selector", graph.ErrValidation)
		}
		_, err := graph.ScopeLocalPublicResourceID(s.ScopeURL(), path)
		return err
	}
	return validateBeadWriteOperand(path, localID)
}

func (s *Store) resolveBeadEndpointOperandInTx(ctx context.Context, tx *sql.Tx, path, localID string) (string, error) {
	if strings.HasPrefix(path, s.ScopeURL()) {
		if localID != "" {
			return "", fmt.Errorf("%w: canonical endpoint cannot also supply a local-ID selector", graph.ErrValidation)
		}
		var err error
		path, err = s.publicResourcePathInTx(ctx, tx, path, graph.KindBead)
		if err != nil {
			return "", err
		}
	}
	return s.resolveLiveBeadWriteSelectorInTx(ctx, tx, path, localID)
}
