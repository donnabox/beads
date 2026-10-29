package graphread

import (
	"context"
	"encoding/json"
	"fmt"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/resourcequery"
)

// AdmitResourceList validates the complete public List AST before acquisition.
// Public List retains real editions and its own projection/delivery controls;
// it is distinct from staged Batch's deliberately revisionless List-ID view.
func AdmitResourceList(raw []byte, kind graph.ResourceKind) error {
	request, err := resourcequery.Decode(raw)
	if err != nil {
		return err
	}
	if err := resourcequery.Admit(request, resourcequery.DefaultLimits()); err != nil {
		return err
	}
	if !kind.Valid() || request.Profile() != "list" || request.ResourceKind() != kind {
		return fmt.Errorf("%w: use a List AST scanning this command's Resource kind", resourcequery.ErrInvalid)
	}
	if request.ResultForm() != "resources" {
		return fmt.Errorf("%w: List AST returns resources; use query --ast for an aggregate", resourcequery.ErrUnsupported)
	}
	// The retained List codec still uses attribution/changeContext, while Query
	// maps those fields (including nested owned Links) to revisionProvenance.
	// Refuse only these known mismatches instead of silently testing one shape
	// and returning another, changing a pin, or redefining missing properties.
	for _, root := range request.FieldRoots() {
		switch root {
		case "revisionProvenance", "attribution", "changeContext", "ownedLinks":
			return fmt.Errorf("%w: List AST field %q requires a shared predicate/result representation; use query --ast for the Query representation", resourcequery.ErrUnsupported, root)
		}
	}
	return nil
}

// InventoryMatchingList returns the same checked inventory with only its
// candidate side filtered. It retains endpoint Beads when listing Links, so the
// caller applies Type and endpoint selectors without acquiring another current
// inventory. It opens no provider or remote Type authority during evaluation.
func (r *Reader) InventoryMatchingList(ctx context.Context, raw []byte, kind graph.ResourceKind) (Inventory, error) {
	if err := AdmitResourceList(raw, kind); err != nil {
		return Inventory{}, err
	}
	inventory, data, err := r.queryInventory(ctx)
	if err != nil {
		return Inventory{}, err
	}
	return matchListInventory(ctx, raw, kind, inventory, data)
}

func matchListInventory(ctx context.Context, raw []byte, kind graph.ResourceKind, inventory Inventory, data resourcequery.Dataset) (Inventory, error) {
	response, err := resourcequery.Execute(ctx, raw, data, resourcequery.DefaultLimits())
	if err != nil {
		return Inventory{}, err
	}
	result, ok := response.Result.(resourcequery.ResourcesResult)
	if !ok || !response.Evaluation.Complete || result.ResourceKind != kind || result.Next != nil {
		return Inventory{}, fmt.Errorf("%w: List requires a complete current Resource selection", resourcequery.ErrDataset)
	}
	selected := make(map[resourcequery.Edition]bool, len(result.Items))
	for _, raw := range result.Items {
		var edition resourcequery.Edition
		if err := json.Unmarshal(raw, &edition); err != nil {
			return Inventory{}, fmt.Errorf("%w: List result edition", resourcequery.ErrDataset)
		}
		selected[edition] = true
	}
	matched := 0
	if kind == graph.KindBead {
		beads := inventory.Beads[:0:0]
		for _, bead := range inventory.Beads {
			if selected[resourcequery.Edition{ID: bead.ID, Revision: bead.Revision}] {
				beads = append(beads, bead)
				matched++
			}
		}
		inventory.Beads = beads
	} else {
		links := inventory.Links[:0:0]
		for _, link := range inventory.Links {
			if selected[resourcequery.Edition{ID: link.ID, Revision: link.Revision}] {
				links = append(links, link)
				matched++
			}
		}
		inventory.Links = links
	}
	if matched != len(selected) {
		return Inventory{}, fmt.Errorf("%w: List selection differs from captured inventory", resourcequery.ErrDataset)
	}
	if err := ctx.Err(); err != nil {
		return Inventory{}, err
	}
	return inventory, nil
}
