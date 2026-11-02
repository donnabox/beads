package graphstore

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/steveyegge/beads/internal/typecatalog"
	"github.com/steveyegge/beads/internal/typeconstraints"
	"github.com/steveyegge/beads/resourceget"
)

type resourceReadContextKey struct{}
type resourceReadContext struct {
	store       *Store
	tx          *sql.Tx
	budget      *intrinsicMetadataBudget
	installed   checkedInstalledTypes
	definitions []resourceget.ResourceDocument
}

// selectedTypesInTx consumes executable contracts and domain obligations from
// the SAME checked stored Type-Bead editions as this operation. The pure
// projection supplies only the immutable selected application's binding map;
// every declaration in definitions below is the checked stored copy.
func (s *Store) selectedTypesInTx(ctx context.Context, tx *sql.Tx) (*selectedTypeCatalog, error) {
	if s.LayoutVersion() != TypesFreshSchemaVersion {
		return s.selectedTypes()
	}
	var installed checkedInstalledTypes
	if read, ok := ctx.Value(resourceReadContextKey{}).(*resourceReadContext); ok && read != nil && read.store == s && read.tx == tx && read.budget != nil {
		installed = read.installed
	} else {
		state, err := s.nativeOperationInContext(ctx, tx)
		if err != nil {
			return nil, err
		}
		installed = state.capture.installed
	}
	if installed.catalog == nil || installed.memories == nil {
		return nil, ErrInvalidStore
	}
	projection, err := newSelectedMemoryProjection(s.ScopeURL())
	if err != nil {
		return nil, err
	}
	selected := &selectedTypeCatalog{projection: projection, types: installed.catalog, definitions: installed.memories}
	rules := []typeconstraints.Rule{}
	for id, declaration := range installed.memories {
		definition, ok := installed.catalog.Definition(id)
		if !ok || definition.Identity.URI != id || definition.Kind != declaration.Kind {
			return nil, fmt.Errorf("%w: stored Memory Type/domain obligation differs", ErrInvalidStore)
		}
		if declaration.Kind != "link" {
			continue
		}
		rule := declaration.LocalGraphRule
		if rule.Family.URI != id {
			return nil, ErrInvalidStore
		}
		rules = append(rules, typeconstraints.Rule{Family: typecatalog.Reference{URI: id, Revision: definition.Identity.Revision}, SourceMax: typeconstraints.Limit{Set: rule.SourceMax.Present, Value: rule.SourceMax.Value}, TargetMax: typeconstraints.Limit{Set: rule.TargetMax.Present, Value: rule.TargetMax.Value}, Acyclic: rule.Acyclic, Unique: rule.Unique, ForbidIdentitySelf: rule.ForbidIdentitySelf, SameScope: rule.SameScope, FloatingSource: rule.FloatingSource, FloatingTarget: rule.FloatingTarget})
	}
	selected.graph, err = typeconstraints.New(installed.catalog, rules)
	if err != nil {
		return nil, err
	}
	return selected, nil
}

func (s *Store) contextForResourceCapture(ctx context.Context, tx *sql.Tx, capture checkedResourceCapture, budget *intrinsicMetadataBudget) context.Context {
	return context.WithValue(ctx, resourceReadContextKey{}, &resourceReadContext{store: s, tx: tx, budget: budget, installed: capture.installed, definitions: capture.snapshot.Documents})
}
