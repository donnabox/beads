package graphread

import "github.com/steveyegge/beads/internal/storage/graphstore"

// ProjectCurrentRecords projects a captured, checked current selection using the
// same Resource/ownership conversion as Inventory. The caller supplies all Type
// and ownership context from that read view. It performs no storage reads and
// does not assert that the selection is the whole workspace or an export.
func ProjectCurrentRecords(selection graphstore.Snapshot) (Inventory, error) {
	return projectInventory(selection)
}
