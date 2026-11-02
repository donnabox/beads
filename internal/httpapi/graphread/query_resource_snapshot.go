package graphread

import (
	"context"
	"fmt"

	"github.com/steveyegge/beads/internal/resourcequery"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// resourceQueryDataset maps one checked native acquisition. It does not resolve
// a definition, reopen the store, substitute current ancestry or rewrite a URI.
// Definition is deliberately absent from the evaluator authority boundary.
func resourceQueryDataset(ctx context.Context, snapshot graphstore.ResourceSnapshot) (resourcequery.Dataset, error) {
	if ctx == nil {
		return resourcequery.Dataset{}, fmt.Errorf("%w: Query context absent", resourcequery.ErrDataset)
	}
	if err := ctx.Err(); err != nil {
		return resourcequery.Dataset{}, err
	}
	if snapshot.Types == nil {
		return resourcequery.Dataset{}, fmt.Errorf("%w: exact captured Type authority absent", resourcequery.ErrDataset)
	}
	data := resourcequery.Dataset{
		Resources:    make([]resourcequery.Resource, 0, len(snapshot.Documents)),
		TypeEditions: make([]resourcequery.TypeEditionFact, 0, len(snapshot.Types)),
	}
	for _, fact := range snapshot.Types {
		if err := ctx.Err(); err != nil {
			return resourcequery.Dataset{}, err
		}
		ancestors := make([]resourcequery.Edition, len(fact.Ancestors))
		for i, ancestor := range fact.Ancestors {
			if err := ctx.Err(); err != nil {
				return resourcequery.Dataset{}, err
			}
			ancestors[i] = resourcequery.Edition{ID: ancestor.URI, Revision: ancestor.Revision}
		}
		data.TypeEditions = append(data.TypeEditions, resourcequery.TypeEditionFact{
			Reference: resourcequery.Edition{ID: fact.Reference.URI, Revision: fact.Reference.Revision},
			Kind:      fact.Describes, Current: fact.Current, Ancestors: ancestors,
		})
	}
	remaining := QueryResourceByteLimit
	for _, document := range snapshot.Documents {
		if err := ctx.Err(); err != nil {
			return resourcequery.Dataset{}, err
		}
		// NativeJSON returns a detached complete record with the captured exact
		// Type, endpoints, owned Links and provenance. No legacy DTO projection.
		raw := document.NativeJSON()
		if len(raw) > remaining {
			return resourcequery.Dataset{}, resourcequery.ErrLimit
		}
		remaining -= len(raw)
		record, err := resourcequery.NewResource(document.Kind(), raw)
		if err != nil {
			return resourcequery.Dataset{}, err
		}
		data.Resources = append(data.Resources, record)
	}
	if err := ctx.Err(); err != nil {
		return resourcequery.Dataset{}, err
	}
	return data, nil
}

// resourceQueryCapture is the owning bridge from a successful native read to
// exact CURRENT Resource acquisition. A caller-supplied ResourceSnapshot or a
// syntax-valid document cannot opt in through the representation-only mapper.
// ReadResourceQuerySnapshot admits only the checked Types-aware store and
// returns zero on acquisition, validation, cancellation or transaction failure.
// Mapping, fact construction and evaluation perform no second native read.
// This authority discharges dynamic stored endpoint pins only; ordinary request
// admission still refuses static pinned refs before any capture.
func (r *Reader) resourceQueryCapture(ctx context.Context) (graphstore.ResourceSnapshot, resourcequery.Dataset, error) {
	if ctx == nil || r == nil || r.store == nil {
		return graphstore.ResourceSnapshot{}, resourcequery.Dataset{}, fmt.Errorf("%w: checked Resource reader and context required", resourcequery.ErrDataset)
	}
	if err := ctx.Err(); err != nil {
		return graphstore.ResourceSnapshot{}, resourcequery.Dataset{}, err
	}
	snapshot, err := r.store.ReadResourceQuerySnapshot(ctx)
	if err != nil {
		return graphstore.ResourceSnapshot{}, resourcequery.Dataset{}, err
	}
	data, err := resourceQueryDataset(ctx, snapshot)
	if err != nil {
		return graphstore.ResourceSnapshot{}, resourcequery.Dataset{}, err
	}
	// These top-level documents are exactly the live allocation/record editions
	// in the successful checked capture. Retained Type facts and owned Link
	// bodies inside them are not independent CURRENT selections. WriterToken
	// is private coordination data and never supplies a public revision.
	facts := make([]resourcequery.CurrentEditionFact, 0, len(snapshot.Documents))
	for _, document := range snapshot.Documents {
		if err := ctx.Err(); err != nil {
			return graphstore.ResourceSnapshot{}, resourcequery.Dataset{}, err
		}
		facts = append(facts, resourcequery.CurrentEditionFact{
			Reference: resourcequery.Edition{ID: document.ID(), Revision: document.Revision()},
			Kind:      document.Kind(),
		})
	}
	if err := ctx.Err(); err != nil {
		return graphstore.ResourceSnapshot{}, resourcequery.Dataset{}, err
	}
	data.CurrentEditions = resourcequery.NewCurrentEditionAuthorityFromTrustedCapture(facts)
	if err := ctx.Err(); err != nil {
		return graphstore.ResourceSnapshot{}, resourcequery.Dataset{}, err
	}
	return snapshot, data, nil
}
