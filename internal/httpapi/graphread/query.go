package graphread

import (
	"context"
	"errors"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/resourcequery"
)

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
	if ctx == nil || r == nil || r.store == nil {
		return resourcequery.Response{}, errors.New("Query requires a reader and context")
	}
	if err := ctx.Err(); err != nil {
		return resourcequery.Response{}, err
	}
	snapshot, err := r.store.ReadQuerySnapshot(ctx)
	if err != nil {
		return resourcequery.Response{}, err
	}
	inventory, err := projectInventory(snapshot.Snapshot)
	if err != nil {
		return resourcequery.Response{}, err
	}
	data := resourcequery.Dataset{TypeFamilies: snapshot.TypeFamilies, TypeKinds: snapshot.TypeKinds,
		Resources: make([]resourcequery.Resource, 0, len(inventory.Beads)+len(inventory.Links))}
	appendResource := func(kind graph.ResourceKind, value any) error {
		rawRecord, err := queryResourceJSON(value)
		if err != nil {
			return err
		}
		record, err := resourcequery.NewResource(kind, rawRecord)
		if err != nil {
			return err
		}
		data.Resources = append(data.Resources, record)
		return nil
	}
	for _, bead := range inventory.Beads {
		if err := appendResource(graph.KindBead, bead); err != nil {
			return resourcequery.Response{}, err
		}
	}
	for _, link := range inventory.Links {
		if err := appendResource(graph.KindLink, link); err != nil {
			return resourcequery.Response{}, err
		}
	}
	return resourcequery.Execute(ctx, raw, data, limits)
}
