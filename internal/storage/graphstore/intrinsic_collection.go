package graphstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/intrinsicmaterialization"
	"github.com/steveyegge/beads/internal/intrinsictypecatalog"
	"github.com/steveyegge/beads/resourceget"
)

// ListDefinitions returns complete current local definition Beads. Type-family
// and exact-affiliation predicates intersect. Selectors are the three explicit
// intrinsic category labels or their checked global URIs, never basename,
// Scope-local identity aliases, installed/default claims or string-Type DTOs.
// The checked context supplies four intrinsic records; admitted local definitions
// are acquired under the same operation budget. CLI paging is separate.
func (s *Store) ListDefinitions(ctx context.Context, family, exact string) ([]resourceget.DefinitionDocument, error) {
	if s.options.Binding.SchemaVersion != TypesFreshSchemaVersion {
		return nil, fmt.Errorf("%w: definition collection requires Types-only layout", ErrCapabilityUnavailable)
	}
	var result []resourceget.DefinitionDocument
	err := s.withDefinitionReadTx(ctx, txDefinitionList, func(tx *sql.Tx, budget *intrinsicMetadataBudget) error {
		id, err := currentIntrinsicContextInTx(ctx, tx, budget)
		if err != nil {
			return err
		}
		plan, err := loadIntrinsicContextInTx(ctx, tx, s.options, id, budget)
		if err != nil {
			return err
		}
		// This second catalog value supplies only the fixed source's exact Refines
		// ancestry; it is not another installed inventory or a latest lookup.
		if err = budget.chargeBody(4 * plan.Source().Bytes); err != nil {
			return err
		}
		catalog, err := intrinsictypecatalog.LoadSource(plan.Source())
		if err != nil {
			return err
		}
		result, err = selectIntrinsicDefinitionBeads(s.ScopeURL(), catalog, plan, family, exact, budget)
		if err != nil {
			return err
		}
		custom, err := s.listCustomDefinitionsInTx(ctx, tx, plan, budget)
		if err != nil {
			return err
		}
		required, err := resolveIntrinsicDefinitionCategory(catalog, plan, family)
		if err != nil {
			return err
		}
		nominal, err := resolveIntrinsicDefinitionCategory(catalog, plan, exact)
		if err != nil {
			return err
		}
		ownType, err := customDefinitionAffiliation(plan)
		if err != nil {
			return err
		}
		ancestors, known := catalog.Ancestors(ownType)
		if !known {
			return ErrInvalidStore
		}
		for _, doc := range custom {
			if nominal.URI != "" && ownType != nominal {
				continue
			}
			if required.URI != "" && ownType != required && !slices.Contains(ancestors, required) {
				continue
			}
			result = append(result, doc)
		}
		slices.SortFunc(result, func(a, b resourceget.DefinitionDocument) int { return graph.CompareCodeUnits(a.ID(), b.ID()) })
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Locate the current context from actual current allocation/member pairs,
// rather than choosing an arbitrary retained context or reconstructing a head.
// Old contexts may remain present. Two live Type-role contexts are corruption.
func currentIntrinsicContextInTx(ctx context.Context, tx *sql.Tx, budget *intrinsicMetadataBudget) (string, error) {
	if err := budget.chargeBody(4096); err != nil {
		return "", err
	}
	rows, err := tx.QueryContext(ctx, `SELECT m.context_id FROM graph_preview_intrinsic_members m JOIN graph_preview_catalog c ON c.path=m.path AND c.revision=m.revision WHERE m.role=? AND c.backing=? AND c.allocation_state=? LIMIT 2`, string(intrinsicmaterialization.Type), "intrinsic", "live")
	if err != nil {
		return "", err
	}
	var ids []string
	for rows.Next() {
		if len(ids) == 1 {
			err = fmt.Errorf("%w: multiple current intrinsic contexts", ErrInvalidStore)
			break
		}
		var id string
		if err = rows.Scan(&id); err != nil {
			break
		}
		if !authorityID.MatchString(id) {
			err = fmt.Errorf("%w: current intrinsic context identity", ErrInvalidStore)
			break
		}
		ids = append(ids, id)
	}
	if err == nil && rows.NextResultSet() {
		err = fmt.Errorf("%w: extra current-context result set", ErrInvalidStore)
	}
	err = errors.Join(err, rows.Err(), rows.Close())
	if err != nil {
		return "", err
	}
	if len(ids) != 1 {
		return "", fmt.Errorf("%w: expected one current intrinsic context", ErrInvalidStore)
	}
	return ids[0], nil
}

func selectIntrinsicDefinitionBeads(scope string, catalog *intrinsictypecatalog.Catalog, plan *intrinsicmaterialization.Plan, family, exact string, budget *intrinsicMetadataBudget) ([]resourceget.DefinitionDocument, error) {
	if catalog == nil || plan == nil || catalog.Source() != plan.Source() {
		return nil, ErrInvalidStore
	}
	// Bindings copies original and local records; reserve before requesting
	// those copies, independently of the validated context's structural cap.
	if err := budget.chargeBody(2 * intrinsicmaterialization.MaxContextBytes); err != nil {
		return nil, err
	}
	bindings := plan.Bindings()
	if len(bindings) != 6 {
		return nil, ErrInvalidStore
	}

	required, err := resolveIntrinsicDefinitionCategory(catalog, plan, family)
	if err != nil {
		return nil, err
	}
	nominal, err := resolveIntrinsicDefinitionCategory(catalog, plan, exact)
	if err != nil {
		return nil, err
	}
	result := make([]resourceget.DefinitionDocument, 0, 4)
	for _, binding := range bindings {
		if binding.Allocation.Kind != graph.KindBead {
			continue
		}
		var header struct {
			Type intrinsictypecatalog.Reference `json:"type"`
		}
		if err = json.Unmarshal(binding.Record, &header); err != nil {
			return nil, fmt.Errorf("%w: definition affiliation: %v", ErrInvalidStore, err)
		}
		ancestors, known := catalog.Ancestors(header.Type)
		if !known {
			return nil, fmt.Errorf("%w: unknown definition affiliation", ErrInvalidStore)
		}
		if nominal.URI != "" && header.Type != nominal {
			continue
		}
		if required.URI != "" && header.Type != required && !slices.Contains(ancestors, required) {
			continue
		}
		if err = budget.chargeBody(4 * len(binding.Record)); err != nil {
			return nil, err
		}
		doc, e := resourceget.NewDefinitionDocument(scope, resourceget.DefinitionBead, binding.Record)
		if e != nil {
			return nil, fmt.Errorf("%w: definition carrier: %v", ErrInvalidStore, e)
		}
		doc, e = doc.WithOrigin(binding.Global.URI, binding.Global.Revision)
		if e != nil {
			return nil, fmt.Errorf("%w: checked definition origin: %v", ErrInvalidStore, e)
		}
		result = append(result, doc)
	}
	slices.SortFunc(result, func(a, b resourceget.DefinitionDocument) int { return graph.CompareCodeUnits(a.ID(), b.ID()) })
	return result, nil
}

// These are exact selected protocol spellings, not basename/URL rewriting.
func resolveIntrinsicDefinitionCategory(catalog *intrinsictypecatalog.Catalog, plan *intrinsicmaterialization.Plan, raw string) (intrinsictypecatalog.Reference, error) {
	if raw == "" {
		return intrinsictypecatalog.Reference{}, nil
	}
	var name string
	switch raw {
	case "type", "https://github.com/gastownhall/bdp/types/type-system/type":
		name = "type"
	case "bead-type", "https://github.com/gastownhall/bdp/types/type-system/bead-type":
		name = "bead-type"
	case "link-type", "https://github.com/gastownhall/bdp/types/type-system/link-type":
		name = "link-type"
	default:
		return intrinsictypecatalog.Reference{}, fmt.Errorf("%w: unsupported intrinsic Bead category selector", ErrCapabilityUnavailable)
	}
	if catalog == nil || plan == nil || catalog.Source() != plan.Source() {
		return intrinsictypecatalog.Reference{}, ErrInvalidStore
	}
	ref, known := catalog.Pin(name)
	if !known {
		return intrinsictypecatalog.Reference{}, ErrInvalidStore
	}
	if _, known := catalog.Ancestors(ref); !known {
		return intrinsictypecatalog.Reference{}, ErrInvalidStore
	}
	if _, known := plan.LocalFor(ref); !known {
		return intrinsictypecatalog.Reference{}, ErrInvalidStore
	}
	return ref, nil
}
