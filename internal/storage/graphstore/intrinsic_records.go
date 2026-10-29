package graphstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	graph "github.com/steveyegge/beads/graphops"

	"github.com/steveyegge/beads/resourceget"
)

// IntrinsicDocument is an internal full exact-affiliation record carrier. It is
// deliberately distinct from the legacy Memory Record.Type string projection.
// Body is freshly copied for each result and owns its nested/captured records.
type IntrinsicDocument struct {
	PublicID    string
	PrivatePath string
	Kind        graph.ResourceKind
	Revision    string
	ContextID   string
	Body        json.RawMessage
}

// GetIntrinsicDefinition reads an actual local allocation, never the publication
// as a substitute for ownership. Empty revision selects its actual current row.
func (s *Store) GetIntrinsicDefinition(ctx context.Context, address, revision string) (IntrinsicDocument, error) {
	if s.options.Binding.SchemaVersion != TypesFreshSchemaVersion {
		return IntrinsicDocument{}, ErrNotFound
	}
	var result IntrinsicDocument
	err := s.withDefinitionReadTx(ctx, txDefinitionRead, func(tx *sql.Tx, budget *intrinsicMetadataBudget) error {
		var identity AllocatedResourceIdentity
		var err error
		if strings.HasPrefix(address, s.ScopeURL()) {
			id, parseErr := graph.ScopeLocalPublicResourceID(s.ScopeURL(), address)
			if parseErr != nil {
				return parseErr
			}
			identity, err = resolveAllocatedResourceIdentityInTx(ctx, tx, id, budget)
		} else {
			if err = validatePublicID(address); err != nil {
				return err
			}
			identity, err = resolveAllocatedResourceIdentityInTx(ctx, tx, address, budget)
		}
		if err != nil {
			return err
		}
		return s.readSelectedDefinitionInTx(ctx, tx, identity.PrivatePath, revision, &result, budget)
	})
	if err != nil {
		return IntrinsicDocument{}, err
	}
	return result, nil
}

func (s *Store) readIntrinsicDocumentInTx(ctx context.Context, tx *sql.Tx, path, revision string, out *IntrinsicDocument, budget *intrinsicMetadataBudget) error {
	if s.options.Binding.SchemaVersion != TypesFreshSchemaVersion {
		return ErrNotFound
	}
	if err := budget.chargeBody(4096 + 2*len(path) + 2*len(revision)); err != nil {
		return err
	}
	var current, backing, state string
	if err := tx.QueryRowContext(ctx, `SELECT revision,backing,allocation_state FROM graph_preview_catalog WHERE path=?`, path).Scan(&current, &backing, &state); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if backing != "intrinsic" && backing != "intrinsic-link" {
		return ErrNotFound
	}
	if state != "live" {
		return ErrGone
	}
	if revision == "" {
		revision = current
	}
	var contextID string
	if err := tx.QueryRowContext(ctx, `SELECT context_id FROM graph_preview_intrinsic_members WHERE path=? AND revision=?`, path, revision).Scan(&contextID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if revision == current {
				return fmt.Errorf("%w: live intrinsic current member missing", ErrInvalidStore)
			}
			var evidence int
			if e := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM graph_preview_versions WHERE path=? AND version=?`, path, revision).Scan(&evidence); e != nil {
				return e
			}
			if evidence != 0 {
				return fmt.Errorf("%w: intrinsic retained member missing", ErrInvalidStore)
			}
			return ErrVersionUnknown
		}
		return err
	}
	plan, err := loadIntrinsicContextInTx(ctx, tx, s.options, contextID, budget)
	if err != nil {
		return err
	}
	for _, member := range plan.Bindings() {
		a := member.Allocation
		if a.PrivatePath == path && a.Local.Revision == revision {
			if err = budget.chargeBody(4 * len(member.Record)); err != nil {
				return err
			}
			*out = IntrinsicDocument{PublicID: a.PublicID, PrivatePath: path, Kind: a.Kind, Revision: revision, ContextID: contextID, Body: append(json.RawMessage(nil), member.Record...)}
			return nil
		}
	}
	return fmt.Errorf("%w: intrinsic mapping missing selected edition", ErrInvalidStore)
}

// GetDefinition is the dedicated exact-affiliation native receiving seam. Only
// a checked persisted local allocation/context can construct its document.
func (s *Store) GetDefinition(ctx context.Context, address, revision string) (resourceget.DefinitionDocument, error) {
	if s.options.Binding.SchemaVersion != TypesFreshSchemaVersion {
		return resourceget.DefinitionDocument{}, ErrNotFound
	}
	native, err := s.GetIntrinsicDefinition(ctx, address, revision)
	if err != nil {
		return resourceget.DefinitionDocument{}, err
	}
	kind := resourceget.DefinitionBead
	if native.Kind == graph.KindLink {
		kind = resourceget.DefinitionLink
	}
	return resourceget.NewDefinitionDocument(s.ScopeURL(), kind, native.Body)
}

// GetDefinitionForProjection is the internal reader adapter: path is explicitly
// a private locator (or canonical Scope URL), localID is a separate whole public
// ID. It never guesses private kind/path from an opaque caller ID.
func (s *Store) GetDefinitionForProjection(ctx context.Context, path, localID, revision string) (resourceget.DefinitionDocument, error) {
	if s.options.Binding.SchemaVersion != TypesFreshSchemaVersion {
		return resourceget.DefinitionDocument{}, ErrNotFound
	}
	if localID != "" {
		if path != intrinsicBeadBackingPath(localID) {
			return resourceget.DefinitionDocument{}, fmt.Errorf("%w: inconsistent local selector", graph.ErrValidation)
		}
		return s.GetDefinition(ctx, localID, revision)
	}
	if strings.HasPrefix(path, s.ScopeURL()) {
		return s.GetDefinition(ctx, path, revision)
	}
	if err := validateResourcePath(path); err != nil {
		return resourceget.DefinitionDocument{}, err
	}
	var native IntrinsicDocument
	err := s.withDefinitionReadTx(ctx, txDefinitionRead, func(tx *sql.Tx, budget *intrinsicMetadataBudget) error {
		return s.readSelectedDefinitionInTx(ctx, tx, path, revision, &native, budget)
	})
	if err != nil {
		return resourceget.DefinitionDocument{}, err
	}
	kind := resourceget.DefinitionBead
	if native.Kind == graph.KindLink {
		kind = resourceget.DefinitionLink
	}
	return resourceget.NewDefinitionDocument(s.ScopeURL(), kind, native.Body)
}

// LayoutVersion is an immutable request-metadata routing hint, like ScopeURL.
// It grants no support/ownership: every definition read checks the actual
// same-transaction persisted binding/schema/policy before acquiring bodies.
func (s *Store) LayoutVersion() int { return s.options.Binding.SchemaVersion }

// intrinsicBeadBackingPath constructs only a private Bead backing locator.
// Callers validate its key or independently match the whole public selector.
func intrinsicBeadBackingPath(key string) string { return "beads/" + key }
