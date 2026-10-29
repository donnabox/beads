package graphread

import (
	"context"
	"errors"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/resourcequery"
)

// QueryResourceByteLimit bounds the mapped evaluator input before validation
// and evaluation retain their own copies, independently of persisted byte limits.
const QueryResourceByteLimit = 16 * 1024 * 1024

func queryResourceWithinBudget(kind graph.ResourceKind, value any, remaining *int) (resourcequery.Resource, error) {
	raw, err := queryResourceJSON(value)
	if err != nil {
		return resourcequery.Resource{}, err
	}
	if len(raw) > *remaining {
		return resourcequery.Resource{}, resourcequery.ErrLimit
	}
	*remaining -= len(raw)
	return resourcequery.NewResource(kind, raw)
}

// ExecuteResourceQuery is the shared local service entry point for CLI and BDP
// Query. Admission precedes acquisition. The store captures one checked read
// context; projection and evaluation use only captured values afterwards.
// Neither adapter opens another read to fill a missing field or endpoint.
func (r *Reader) ExecuteResourceQuery(ctx context.Context, raw []byte) (resourcequery.Response, error) {
	request, err := resourcequery.Decode(raw)
	if err != nil {
		return resourcequery.Response{}, err
	}
	limits := resourcequery.DefaultLimits()
	if err := resourcequery.Admit(request, limits); err != nil {
		return resourcequery.Response{}, err
	}
	_, data, err := r.queryInventory(ctx)
	if err != nil {
		return resourcequery.Response{}, err
	}
	return resourcequery.Execute(ctx, raw, data, limits)
}

// queryInventory captures public Resources and their Type authority together.
// Query and candidate-local List evaluate only this detached context.
func (r *Reader) queryInventory(ctx context.Context) (Inventory, resourcequery.Dataset, error) {
	if ctx == nil || r == nil || r.store == nil {
		return Inventory{}, resourcequery.Dataset{}, errors.New("Query requires a reader and context")
	}
	if err := ctx.Err(); err != nil {
		return Inventory{}, resourcequery.Dataset{}, err
	}
	snapshot, err := r.store.ReadQuerySnapshot(ctx)
	if err != nil {
		return Inventory{}, resourcequery.Dataset{}, err
	}
	inventory, err := projectInventory(snapshot.Snapshot)
	if err != nil {
		return Inventory{}, resourcequery.Dataset{}, err
	}
	data := resourcequery.Dataset{TypeFamilies: snapshot.TypeFamilies, TypeKinds: snapshot.TypeKinds,
		Resources: make([]resourcequery.Resource, 0, len(inventory.Beads)+len(inventory.Links))}
	remaining := QueryResourceByteLimit
	appendResource := func(kind graph.ResourceKind, value any) error {
		record, err := queryResourceWithinBudget(kind, value, &remaining)
		if err != nil {
			return err
		}
		data.Resources = append(data.Resources, record)
		return nil
	}
	for _, bead := range inventory.Beads {
		if err := appendResource(graph.KindBead, bead); err != nil {
			return Inventory{}, resourcequery.Dataset{}, err
		}
	}
	for _, link := range inventory.Links {
		if err := appendResource(graph.KindLink, link); err != nil {
			return Inventory{}, resourcequery.Dataset{}, err
		}
	}
	return inventory, data, nil
}
