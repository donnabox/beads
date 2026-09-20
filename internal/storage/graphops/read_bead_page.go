package graphops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"

	graph "github.com/steveyegge/beads/graphops"
)

// Decoded private operands/results; neither window nor lastPath is a minted
// Cursor. The caller owns one transaction and separately establishes the
// selected-state installation/physical-live-set prerequisite.
type beadRowsPage struct {
	items    []graph.BeadRecord
	lastPath string
	hasMore  bool
}

type beadPageBudget struct {
	*readBudget
	rowsLeft, groupsLeft int
}

func beadPageBudgetFor(ctx context.Context, scope, typeURL string, window pageWindow, limits readLimits) (*beadPageBudget, error) {
	if err := observationContext(ctx); err != nil {
		return nil, fmt.Errorf("Bead page context: %w", err)
	}
	b, err := budgetFor(scope, limits)
	if err != nil {
		return nil, err
	}
	if window.limit < 1 || window.limit >= limits.rows {
		return nil, fmt.Errorf("%w: Bead page must leave room for lookahead", graph.ErrValidation)
	}
	if window.afterPath != "" {
		if err := graph.ValidateBeadPath(window.afterPath); err != nil {
			return nil, err
		}
	}
	if typeURL != "" {
		if err := graph.ValidateTypeURL(typeURL); err != nil {
			return nil, err
		}
	}
	if len(window.afterPath) > limits.valueBytes || len(typeURL) > limits.valueBytes {
		return nil, errBudget
	}
	return &beadPageBudget{readBudget: b, rowsLeft: limits.rows, groupsLeft: limits.rows}, nil
}

func beadPageQuery(typeURL string, window pageWindow, valueBytes int) (string, []any) {
	query := "SELECT " + beadColumns + " FROM graph_beads WHERE path > ?"
	args := []any{valueBytes, window.afterPath}
	if typeURL != "" {
		query += " AND type_url = ?"
		args = append(args, typeURL)
	}
	return query + " ORDER BY path LIMIT ?", append(args, window.limit+1)
}

func pageBeadsInTx(ctx context.Context, tx queryer, typeURL string, window pageWindow, b *beadPageBudget) ([]graph.Bead, error) {
	query, args := beadPageQuery(typeURL, window, b.limits.valueBytes)
	previous := window.afterPath
	items, err := readRows(ctx, tx, query, args, window.limit+1, func(rows *sql.Rows) (graph.Bead, error) {
		var row resourceRow
		if err := scanResource(rows, &row); err != nil {
			return graph.Bead{}, err
		}
		if err := chargeResource(b.readBudget, row); err != nil {
			return graph.Bead{}, err
		}
		bead, err := decodeBead(row)
		if err != nil {
			return graph.Bead{}, err
		}
		if (typeURL != "" && bead.TypeURL() != typeURL) || graph.CompareCodeUnits(bead.Path(), previous) <= 0 {
			return graph.Bead{}, corrupt(errors.New("Bead page selection or order mismatch"))
		}
		previous = bead.Path()
		return bead, nil
	})
	if err != nil {
		return nil, pageRowsError(err)
	}
	b.rowsLeft -= len(items)
	return items, nil
}

func beadDescriptorBatchQuery(ids []string, valueBytes int) (string, []any) {
	args := []any{valueBytes}
	for _, id := range ids {
		args = append(args, id)
	}
	query := "SELECT url, CASE WHEN LENGTH(descriptor) < 0 OR LENGTH(descriptor) > ? THEN NULL ELSE descriptor END, LENGTH(descriptor), fingerprint FROM graph_type_descriptors WHERE url IN (" + strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",") + ") LIMIT ?"
	return query, append(args, len(ids)+1)
}

func beadDescriptorsInTx(ctx context.Context, tx queryer, beads []graph.Bead, b *beadPageBudget) (map[string]graph.TypeDescriptor, error) {
	wanted := map[string]bool{}
	for _, bead := range beads {
		wanted[bead.TypeURL()] = true
	}
	if len(wanted) > b.rowsLeft {
		return nil, errBudget
	}
	ids := make([]string, 0, len(wanted))
	for id := range wanted {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	query, args := beadDescriptorBatchQuery(ids, b.limits.valueBytes)
	found := make(map[string]graph.TypeDescriptor, len(ids))
	_, err := readRows(ctx, tx, query, args, len(ids)+1, func(rows *sql.Rows) (struct{}, error) {
		var id, fingerprint string
		var raw []byte
		var length sql.NullInt64
		if err := rows.Scan(&id, &raw, &length, &fingerprint); err != nil {
			return struct{}{}, err
		}
		if err := b.charge([]byte(id), []byte(fingerprint)); err != nil {
			return struct{}{}, err
		}
		if err := chargeBlob(b.readBudget, raw, length); err != nil {
			return struct{}{}, err
		}
		if !wanted[id] {
			return struct{}{}, corrupt(errors.New("unexpected or duplicate Bead descriptor"))
		}
		d, err := decodeDescriptor(id, raw, fingerprint, graph.KindBead)
		if err != nil {
			return struct{}{}, err
		}
		found[id] = d
		delete(wanted, id)
		return struct{}{}, nil
	})
	if err != nil {
		return nil, pageRowsError(err)
	}
	if len(wanted) != 0 {
		return nil, corrupt(errors.New("missing Bead page descriptor"))
	}
	b.rowsLeft -= len(found)
	return found, nil
}

type beadPageOwner struct {
	record graph.BeadRecord
	owns   []graph.OwnedLinkDecl
	groups map[string][]graph.Link
	total  int
}

func (b *beadPageBudget) group(key string) error {
	if b.groupsLeft == 0 {
		return errBudget
	}
	if err := b.charge([]byte(key)); err != nil {
		return err
	}
	b.groupsLeft--
	return nil
}

func makeBeadPageOwners(beads []graph.Bead, descriptors map[string]graph.TypeDescriptor, b *beadPageBudget) ([]*beadPageOwner, error) {
	owners := make([]*beadPageOwner, 0, len(beads))
	for _, bead := range beads {
		// OwnsOutgoing copies its slice before retained metadata accounting. The
		// already charged descriptor byte limit bounds that transient parse/copy;
		// this is not a driver/engine allocation guarantee.
		descriptor, ok := descriptors[bead.TypeURL()]
		if !ok {
			return nil, corrupt(errors.New("owner without decoded descriptor"))
		}
		owns := descriptor.OwnsOutgoing()
		if len(owns) > b.groupsLeft {
			return nil, errBudget
		}
		owner := &beadPageOwner{record: graph.BeadRecord{Bead: bead}, owns: owns, groups: map[string][]graph.Link{}}
		for _, decl := range owns {
			if err := b.group(decl.TypeURL()); err != nil {
				return nil, err
			}
			if !decl.Wildcard() {
				owner.groups[decl.TypeURL()] = nil
			}
		}
		owners = append(owners, owner)
	}
	return owners, nil
}

// Each owner's cap is summed with saturation at stop, avoiding overflow even
// when a descriptor contains a large admissible Max. A wildcard bounds its
// whole set, including explicit Types.
func beadOwnedCap(owners []*beadPageOwner, stop int) int {
	total := 0
	for _, owner := range owners {
		cap := 0
		for _, decl := range owner.owns {
			if decl.Wildcard() {
				cap = min(stop, decl.Max())
				break
			}
			cap += min(stop-cap, decl.Max())
		}
		total += min(stop-total, cap)
	}
	return total
}

func beadOwnedBatchQuery(owners []*beadPageOwner, valueBytes, cap int) (string, []any) {
	var clauses []string
	args := []any{valueBytes}
	for _, owner := range owners {
		if len(owner.owns) == 0 {
			continue
		}
		var types []string
		wildcard := false
		for _, decl := range owner.owns {
			if decl.Wildcard() {
				wildcard = true
				break
			}
			types = append(types, decl.TypeURL())
		}
		clause := "(source_path = ?"
		args = append(args, owner.record.Bead.Path())
		if !wildcard {
			clause += " AND type_url IN (" + strings.TrimSuffix(strings.Repeat("?,", len(types)), ",") + ")"
			for _, typ := range types {
				args = append(args, typ)
			}
		}
		clauses = append(clauses, clause+")")
	}
	return "SELECT " + linkColumns + " FROM graph_links WHERE source_kind = 'in' AND (" + strings.Join(clauses, " OR ") + ") ORDER BY source_path, type_url, path LIMIT ?", append(args, cap+1)
}

func (owner *beadPageOwner) add(link graph.Link, b *beadPageBudget, retain bool) error {
	var explicit, wildcard graph.OwnedLinkDecl
	for _, decl := range owner.owns {
		if decl.Wildcard() {
			wildcard = decl
		} else if decl.TypeURL() == link.TypeURL() {
			explicit = decl
		}
	}
	if explicit.TypeURL() == "" && !wildcard.Wildcard() {
		return corrupt(errors.New("unowned Link in Bead page"))
	}
	if (explicit.TypeURL() != "" && len(owner.groups[link.TypeURL()]) >= explicit.Max()) ||
		(wildcard.Wildcard() && owner.total >= wildcard.Max()) {
		return corrupt(errors.New("Bead page owned set exceeds descriptor maximum"))
	}
	// The cap+1 witness is validated but never admitted to a retained group.
	if !retain {
		return nil
	}
	if _, ok := owner.groups[link.TypeURL()]; !ok {
		if err := b.group(link.TypeURL()); err != nil {
			return err
		}
	}
	owner.total++
	owner.groups[link.TypeURL()] = append(owner.groups[link.TypeURL()], link)
	return nil
}

func beadOwnedInTx(ctx context.Context, tx queryer, scope string, owners []*beadPageOwner, b *beadPageBudget) error {
	byPath := map[string]*beadPageOwner{}
	for _, owner := range owners {
		if len(owner.owns) > 0 {
			byPath[owner.record.Bead.Path()] = owner
		}
	}
	if len(byPath) == 0 {
		return nil
	}
	descriptorCap := beadOwnedCap(owners, b.rowsLeft+1)
	cap := min(b.rowsLeft, descriptorCap)
	query, args := beadOwnedBatchQuery(owners, b.limits.valueBytes, cap)
	previous := [3]string{}
	seen := map[string]bool{}
	items, err := readRows(ctx, tx, query, args, cap+1, func(rows *sql.Rows) (struct{}, error) {
		var row linkRow
		if err := rows.Scan(linkScanTargets(&row)...); err != nil {
			return struct{}{}, err
		}
		link, err := chargedLink(scope, row, b.readBudget)
		if err != nil {
			return struct{}{}, err
		}
		owner := byPath[link.Source().Path()]
		if !link.Source().InScope() || owner == nil {
			return struct{}{}, corrupt(errors.New("foreign source in Bead page owned batch"))
		}
		current := [3]string{link.Source().Path(), link.TypeURL(), link.Path()}
		if seen[link.Path()] || !beadOwnedTupleAfter(current, previous) {
			return struct{}{}, corrupt(errors.New("duplicate or unordered Bead page owned Link"))
		}
		seen[link.Path()] = true
		previous = current
		return struct{}{}, owner.add(link, b, len(seen) <= cap)
	})
	if err != nil {
		return pageRowsError(err)
	}
	if len(items) > cap {
		if descriptorCap <= b.rowsLeft {
			return corrupt(errors.New("Bead page aggregate descriptor maximum exceeded"))
		}
		return errBudget
	}
	b.rowsLeft -= len(items)
	return nil
}

func beadOwnedTupleAfter(current, previous [3]string) bool {
	return slices.Compare(current[:], previous[:]) > 0
}

func finishBeadOwners(owners []*beadPageOwner) ([]graph.BeadRecord, error) {
	items := make([]graph.BeadRecord, 0, len(owners))
	for _, owner := range owners {
		for typ, links := range owner.groups {
			owner.record.OwnedLinks = append(owner.record.OwnedLinks, graph.OwnedLinkGroup{TypeURL: typ, Links: links})
		}
		sort.Slice(owner.record.OwnedLinks, func(i, j int) bool {
			return graph.CompareCodeUnits(owner.record.OwnedLinks[i].TypeURL, owner.record.OwnedLinks[j].TypeURL) < 0
		})
		if err := graph.CheckBeadRecord(owner.record, owner.owns); err != nil {
			return nil, corrupt(err)
		}
		items = append(items, owner.record)
	}
	return items, nil
}

// Three body statements at most, shared row/byte/group limits. A legal record
// can exceed these private limits even when singleton reads admit it. Never use
// this as a public traversal guarantee without installer/capacity reconciliation.
func readBeadPageInTx(ctx context.Context, tx queryer, scope, typeURL string, window pageWindow, limits readLimits) (beadRowsPage, error) {
	b, err := beadPageBudgetFor(ctx, scope, typeURL, window, limits)
	if err != nil {
		return beadRowsPage{}, err
	}
	beads, err := pageBeadsInTx(ctx, tx, typeURL, window, b)
	if err != nil || len(beads) == 0 {
		return beadRowsPage{}, err
	}
	descriptors, err := beadDescriptorsInTx(ctx, tx, beads, b)
	if err != nil {
		return beadRowsPage{}, err
	}
	more := len(beads) > window.limit
	if more {
		beads = beads[:window.limit]
	}
	owners, err := makeBeadPageOwners(beads, descriptors, b)
	if err != nil {
		return beadRowsPage{}, err
	}
	if err := beadOwnedInTx(ctx, tx, scope, owners, b); err != nil {
		return beadRowsPage{}, err
	}
	items, err := finishBeadOwners(owners)
	if err != nil {
		return beadRowsPage{}, err
	}
	if err := ctx.Err(); err != nil {
		return beadRowsPage{}, err
	}
	return beadRowsPage{items: items, lastPath: beads[len(beads)-1].Path(), hasMore: more}, nil
}
