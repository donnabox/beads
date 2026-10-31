package graphread

import (
	"context"
	"encoding/json"
	"fmt"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/resourcequery"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/resourceget"
)

// ResourceInventoryMatchingList evaluates the ordinary List predicate over one
// complete native acquisition. The caller keeps normal collection selection,
// endpoint rules, output and counting. No definition is enrolled separately.
func (r *Reader) ResourceInventoryMatchingList(ctx context.Context, raw []byte, kind graph.ResourceKind) (graphstore.ResourceSnapshot, error) {
	if err := AdmitResourceList(raw, kind); err != nil {
		return graphstore.ResourceSnapshot{}, err
	}
	if ctx == nil || r == nil || r.store == nil {
		return graphstore.ResourceSnapshot{}, fmt.Errorf("%w: List reader and context required", resourcequery.ErrDataset)
	}
	snapshot, data, err := r.resourceQueryCapture(ctx)
	if err != nil {
		return graphstore.ResourceSnapshot{}, err
	}
	return matchResourceListSnapshot(ctx, raw, kind, snapshot, data)
}

func matchResourceListSnapshot(ctx context.Context, raw []byte, kind graph.ResourceKind, snapshot graphstore.ResourceSnapshot, data resourcequery.Dataset) (graphstore.ResourceSnapshot, error) {
	response, err := resourcequery.Execute(ctx, raw, data, resourcequery.DefaultLimits())
	if err != nil {
		return graphstore.ResourceSnapshot{}, err
	}
	result, ok := response.Result.(resourcequery.ResourcesResult)
	if !ok || !response.Evaluation.Complete || result.ResourceKind != kind || result.Next != nil {
		return graphstore.ResourceSnapshot{}, fmt.Errorf("%w: List requires a complete current Resource selection", resourcequery.ErrDataset)
	}
	selected := make(map[resourcequery.Edition]bool, len(result.Items))
	for _, raw := range result.Items {
		if err := ctx.Err(); err != nil {
			return graphstore.ResourceSnapshot{}, err
		}
		var edition resourcequery.Edition
		if err := json.Unmarshal(raw, &edition); err != nil || edition.ID == "" || edition.Revision == "" || selected[edition] {
			return graphstore.ResourceSnapshot{}, fmt.Errorf("%w: invalid or duplicate List result edition", resourcequery.ErrDataset)
		}
		selected[edition] = true
	}
	// Retain the other Resource kind and every exact Type fact from this same
	// acquisition. In particular, Link endpoint selection must see all Beads.
	documents := make([]resourceget.ResourceDocument, 0, len(snapshot.Documents))
	matched := make(map[resourcequery.Edition]bool, len(selected))
	for _, document := range snapshot.Documents {
		if err := ctx.Err(); err != nil {
			return graphstore.ResourceSnapshot{}, err
		}
		if document.Kind() != kind {
			documents = append(documents, document)
			continue
		}
		edition := resourcequery.Edition{ID: document.ID(), Revision: document.Revision()}
		if selected[edition] {
			if matched[edition] {
				return graphstore.ResourceSnapshot{}, fmt.Errorf("%w: duplicate captured List edition", resourcequery.ErrDataset)
			}
			documents = append(documents, document)
			matched[edition] = true
		}
	}
	if len(matched) != len(selected) {
		return graphstore.ResourceSnapshot{}, fmt.Errorf("%w: List selection differs from captured inventory", resourcequery.ErrDataset)
	}
	if err := ctx.Err(); err != nil {
		return graphstore.ResourceSnapshot{}, err
	}
	snapshot.Documents = documents
	return snapshot, nil
}
