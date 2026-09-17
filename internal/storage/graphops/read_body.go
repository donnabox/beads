package graphops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"

	graph "github.com/steveyegge/beads/graphops"
)

// The caller owns one transaction. No Exec/commit/authority capability is exposed.
type queryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

var errRowOverflow = errors.New("graph row count exceeded")

type readLimits struct{ rows, bytes, valueBytes int }
type readBudget struct {
	remaining int
	limits    readLimits
}

func budgetFor(scope string, limits readLimits) (*readBudget, error) {
	if err := graph.ValidatePersistedScopeURL(scope); err != nil {
		return nil, err
	}
	// Candidate implementation caps, not advertised protocol or schema limits.
	if limits.rows < 1 || limits.rows > 1024 || limits.valueBytes < 2 || limits.valueBytes > 1<<20 || limits.bytes < limits.valueBytes || limits.bytes > 8<<20 {
		return nil, fmt.Errorf("%w: invalid private read limits", graph.ErrValidation)
	}
	return &readBudget{remaining: limits.bytes, limits: limits}, nil
}

func (b *readBudget) charge(values ...[]byte) error {
	for _, value := range values {
		if len(value) > b.limits.valueBytes || len(value) > b.remaining {
			return errBudget
		}
		b.remaining -= len(value)
	}
	return nil
}

// CAST preserves NULL while avoiding the embedded driver nullable-ENUM scan panic.
// BLOB prefixes cap materialization before JSON parsing. LIMIT + 1 detects
// incomplete expansions. Scalar columns have the finite widths in spec B4.
const beadColumns = "path, type_url, revision, attribution_principal, CAST(attribution_status AS CHAR), SUBSTRING(properties, 1, ?)"
const linkColumns = beadColumns + ", source_kind, source_path, source_url, source_pin, target_kind, target_path, target_url, target_pin"

func scanResource(rows *sql.Rows, row *resourceRow, tail ...any) error {
	args := []any{&row.path, &row.typeURL, &row.revision, &row.principal, &row.attribution, &row.properties}
	return rows.Scan(append(args, tail...)...)
}
func chargeResource(b *readBudget, row resourceRow) error {
	return b.charge([]byte(row.path), []byte(row.typeURL), []byte(row.revision), []byte(row.principal.String), []byte(row.attribution.String), row.properties)
}

func readRows[T any](ctx context.Context, tx queryer, query string, args []any, max int, scan func(*sql.Rows) (T, error)) (result []T, err error) {
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = errors.Join(err, rows.Close())
		if err != nil {
			result = nil
		}
	}()
	for rows.Next() {
		if len(result) == max {
			return nil, errRowOverflow
		}
		value, err := scan(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	return result, rows.Err()
}

func beadRowInTx(ctx context.Context, tx queryer, path string, b *readBudget) (graph.Bead, error) {
	items, err := readRows(ctx, tx, "SELECT "+beadColumns+" FROM graph_beads WHERE path = ? LIMIT 2", []any{b.limits.valueBytes + 1, path}, 1, func(rows *sql.Rows) (graph.Bead, error) {
		var row resourceRow
		if err := scanResource(rows, &row); err != nil {
			return graph.Bead{}, err
		}
		if err := chargeResource(b, row); err != nil {
			return graph.Bead{}, err
		}
		if row.path != path {
			return graph.Bead{}, corrupt(errors.New("bead lookup returned another path"))
		}
		return decodeBead(row)
	})
	if errors.Is(err, errRowOverflow) {
		err = errors.Join(err, corrupt(errors.New("duplicate singleton row")))
	}
	if err != nil {
		return graph.Bead{}, err
	}
	if len(items) == 0 {
		return graph.Bead{}, errAbsent
	}
	return items[0], nil
}

func descriptorInTx(ctx context.Context, tx queryer, id string, kind graph.ResourceKind, b *readBudget) (graph.TypeDescriptor, error) {
	items, err := readRows(ctx, tx, "SELECT url, SUBSTRING(descriptor, 1, ?), fingerprint FROM graph_type_descriptors WHERE url = ? LIMIT 2", []any{b.limits.valueBytes + 1, id}, 1, func(rows *sql.Rows) (graph.TypeDescriptor, error) {
		var found, fingerprint string
		var raw []byte
		if err := rows.Scan(&found, &raw, &fingerprint); err != nil {
			return graph.TypeDescriptor{}, err
		}
		if err := b.charge([]byte(found), raw, []byte(fingerprint)); err != nil {
			return graph.TypeDescriptor{}, err
		}
		if found != id {
			return graph.TypeDescriptor{}, corrupt(errors.New("descriptor lookup returned another id"))
		}
		return decodeDescriptor(id, raw, fingerprint, kind)
	})
	if errors.Is(err, errRowOverflow) {
		err = errors.Join(err, corrupt(errors.New("duplicate singleton row")))
	}
	if err != nil {
		return graph.TypeDescriptor{}, err
	}
	if len(items) == 0 {
		return graph.TypeDescriptor{}, corrupt(errors.New("missing declared descriptor"))
	}
	return items[0], nil
}

func linksInTx(ctx context.Context, tx queryer, scope, where string, args []any, b *readBudget) ([]graph.Link, error) {
	query := "SELECT " + linkColumns + " FROM graph_links WHERE " + where + " ORDER BY path LIMIT ?"
	parameters := append([]any{b.limits.valueBytes + 1}, args...)
	parameters = append(parameters, b.limits.rows+1)
	links, err := readRows(ctx, tx, query, parameters, b.limits.rows, func(rows *sql.Rows) (graph.Link, error) {
		var row linkRow
		if err := scanResource(rows, &row.resourceRow, &row.source.kind, &row.source.path, &row.source.url, &row.source.pin, &row.target.kind, &row.target.path, &row.target.url, &row.target.pin); err != nil {
			return graph.Link{}, err
		}
		return chargedLink(scope, row, b)
	})
	if err != nil {
		return nil, err
	}
	// SQL binary collation is byte ordering; the domain law is UTF-16 code units.
	return orderedLinks(links)
}

// Nullable outer-join rows have a separate presence flag; coalescing only makes
// them scannable, never converts an absent Link into a domain value. CAST before
// COALESCE retains ENUM labels: this engine otherwise returns their numeric index.
const joinedLinkColumns = "COALESCE(l.path, ''), COALESCE(l.type_url, ''), COALESCE(l.revision, ''), l.attribution_principal, CAST(l.attribution_status AS CHAR), SUBSTRING(l.properties, 1, ?), COALESCE(CAST(l.source_kind AS CHAR), ''), l.source_path, l.source_url, l.source_pin, COALESCE(CAST(l.target_kind AS CHAR), ''), l.target_path, l.target_url, l.target_pin"
const exactLinkQuery = "SELECT " + joinedLinkColumns + ", d.url, SUBSTRING(d.descriptor, 1, ?), d.fingerprint FROM graph_links l LEFT JOIN graph_type_descriptors d ON d.url = l.type_url WHERE l.path = ? LIMIT 2"

func linkScanTargets(row *linkRow) []any {
	return []any{&row.path, &row.typeURL, &row.revision, &row.principal, &row.attribution, &row.properties, &row.source.kind, &row.source.path, &row.source.url, &row.source.pin, &row.target.kind, &row.target.path, &row.target.url, &row.target.pin}
}
func chargedLink(scope string, row linkRow, b *readBudget) (graph.Link, error) {
	if err := chargeResource(b, row.resourceRow); err != nil {
		return graph.Link{}, err
	}
	if err := b.charge([]byte(row.source.kind), []byte(row.source.path.String), []byte(row.source.url.String), []byte(row.source.pin.String), []byte(row.target.kind), []byte(row.target.path.String), []byte(row.target.url.String), []byte(row.target.pin.String)); err != nil {
		return graph.Link{}, err
	}
	return decodeLink(scope, row)
}
func orderedLinks(links []graph.Link) ([]graph.Link, error) {
	sort.Slice(links, func(i, j int) bool { return graph.CompareCodeUnits(links[i].Path(), links[j].Path()) < 0 })
	for i := 1; i < len(links); i++ {
		if links[i-1].Path() == links[i].Path() {
			return nil, corrupt(errors.New("duplicate Link path"))
		}
	}
	return links, nil
}

func readLinkInTx(ctx context.Context, tx queryer, scope, path string, limits readLimits) (graph.Link, error) {
	b, err := budgetFor(scope, limits)
	if err != nil {
		return graph.Link{}, err
	}
	if err := graph.ValidateLinkPath(path); err != nil {
		return graph.Link{}, err
	}
	links, err := readRows(ctx, tx, exactLinkQuery, []any{limits.valueBytes + 1, limits.valueBytes + 1, path}, 1, func(rows *sql.Rows) (graph.Link, error) {
		var row linkRow
		var id, fingerprint sql.NullString
		var raw []byte
		targets := append(linkScanTargets(&row), &id, &raw, &fingerprint)
		if err := rows.Scan(targets...); err != nil {
			return graph.Link{}, err
		}
		if row.path != path || !id.Valid || !fingerprint.Valid || id.String != row.typeURL {
			return graph.Link{}, corrupt(errors.New("exact Link or declared descriptor mismatch"))
		}
		link, err := chargedLink(scope, row, b)
		if err != nil {
			return graph.Link{}, err
		}
		if err := b.charge([]byte(id.String), raw, []byte(fingerprint.String)); err != nil {
			return graph.Link{}, err
		}
		if _, err := decodeDescriptor(id.String, raw, fingerprint.String, graph.KindLink); err != nil {
			return graph.Link{}, err
		}
		return link, nil
	})
	if errors.Is(err, errRowOverflow) {
		err = errors.Join(err, corrupt(errors.New("duplicate singleton row")))
	}
	if err != nil {
		return graph.Link{}, err
	}
	if len(links) == 0 {
		return graph.Link{}, errAbsent
	}
	return links[0], nil
}

func incidentQuery(direction graph.Direction) (string, int) {
	selection := "SELECT path FROM graph_links WHERE source_kind = 'in' AND source_path = ?"
	count := 1
	if direction == graph.DirectionIn {
		selection = "SELECT path FROM graph_links WHERE target_kind = 'in' AND target_path = ?"
	}
	if direction == graph.DirectionBoth {
		selection += " UNION SELECT path FROM graph_links WHERE target_kind = 'in' AND target_path = ?"
		count = 2
	}
	return "SELECT b.path, l.path IS NOT NULL, " + joinedLinkColumns + " FROM graph_beads b LEFT JOIN (" + selection + ") i ON TRUE LEFT JOIN graph_links l ON l.path = i.path WHERE b.path = ? ORDER BY l.path LIMIT ?", count
}

func readIncidentLinksInTx(ctx context.Context, tx queryer, scope, path string, direction graph.Direction, limits readLimits) ([]graph.Link, error) {
	b, err := budgetFor(scope, limits)
	if err != nil {
		return nil, err
	}
	if err := graph.ValidateBeadPath(path); err != nil {
		return nil, err
	}
	if !direction.Valid() {
		return nil, fmt.Errorf("%w: invalid direction", graph.ErrValidation)
	}
	query, count := incidentQuery(direction)
	args := []any{limits.valueBytes + 1}
	for i := 0; i < count; i++ {
		args = append(args, path)
	}
	args = append(args, path, limits.rows+1)
	type incident struct {
		link    graph.Link
		present bool
	}
	result, err := readRows(ctx, tx, query, args, limits.rows, func(rows *sql.Rows) (incident, error) {
		var row linkRow
		var anchor string
		var present bool
		targets := append([]any{&anchor, &present}, linkScanTargets(&row)...)
		if err := rows.Scan(targets...); err != nil {
			return incident{}, err
		}
		if anchor != path {
			return incident{}, corrupt(errors.New("incident anchor mismatch"))
		}
		if err := b.charge([]byte(anchor)); err != nil {
			return incident{}, err
		}
		if !present {
			return incident{}, nil
		}
		link, err := chargedLink(scope, row, b)
		if err != nil {
			return incident{}, err
		}
		outbound := link.Source().InScope() && link.Source().Path() == path
		inbound := link.Target().InScope() && link.Target().Path() == path
		if (direction == graph.DirectionIn && !inbound) || (direction == graph.DirectionOut && !outbound) || (direction == graph.DirectionBoth && !inbound && !outbound) {
			return incident{}, corrupt(errors.New("unrelated incident Link"))
		}
		return incident{link: link, present: true}, nil
	})
	if errors.Is(err, errRowOverflow) {
		err = errors.Join(err, errBudget)
	}
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, errAbsent
	}
	var links []graph.Link
	for _, row := range result {
		if row.present {
			links = append(links, row.link)
		} else if len(result) != 1 {
			return nil, corrupt(errors.New("mixed empty incident result"))
		}
	}
	return orderedLinks(links)
}

func readBeadInTx(ctx context.Context, tx queryer, scope, path string, limits readLimits) (graph.BeadRecord, error) {
	b, err := budgetFor(scope, limits)
	if err != nil {
		return graph.BeadRecord{}, err
	}
	if err := graph.ValidateBeadPath(path); err != nil {
		return graph.BeadRecord{}, err
	}
	bead, err := beadRowInTx(ctx, tx, path, b)
	if err != nil {
		return graph.BeadRecord{}, err
	}
	descriptor, err := descriptorInTx(ctx, tx, bead.TypeURL(), graph.KindBead, b)
	if err != nil {
		return graph.BeadRecord{}, err
	}
	owns := descriptor.OwnsOutgoing()
	record := graph.BeadRecord{Bead: bead}
	if len(owns) == 0 {
		return record, nil
	}
	if len(owns) > limits.rows {
		return graph.BeadRecord{}, errBudget
	}
	groups := map[string][]graph.Link{}
	wildcard := false
	args := []any{path}
	var marks []string
	for _, decl := range owns {
		if decl.Wildcard() {
			wildcard = true
			continue
		}
		groups[decl.TypeURL()] = nil
		args = append(args, decl.TypeURL())
		marks = append(marks, "?")
	}
	where := "source_kind = 'in' AND source_path = ?"
	if !wildcard {
		where += " AND type_url IN (" + strings.Join(marks, ",") + ")"
	} else {
		args = args[:1]
	}
	// Bound the materialized owned set by both descriptor declarations and the
	// private expansion budget. Saturating addition avoids overflow from large Max.
	ownedCap := 0
	for _, decl := range owns {
		if decl.Wildcard() {
			ownedCap = min(limits.rows, decl.Max())
			break
		}
		ownedCap += min(limits.rows-ownedCap, decl.Max())
	}
	b.limits.rows = ownedCap
	links, err := linksInTx(ctx, tx, scope, where, args, b)
	if errors.Is(err, errRowOverflow) {
		if ownedCap < limits.rows {
			err = errors.Join(err, corrupt(errors.New("owned set exceeds descriptor maximum")))
		} else {
			err = errors.Join(err, errBudget)
		}
	}
	if err != nil {
		return graph.BeadRecord{}, err
	}
	for _, link := range links {
		groups[link.TypeURL()] = append(groups[link.TypeURL()], link)
	}
	for id, links := range groups {
		record.OwnedLinks = append(record.OwnedLinks, graph.OwnedLinkGroup{TypeURL: id, Links: links})
	}
	sort.Slice(record.OwnedLinks, func(i, j int) bool {
		return graph.CompareCodeUnits(record.OwnedLinks[i].TypeURL, record.OwnedLinks[j].TypeURL) < 0
	})
	if err := graph.CheckBeadRecord(record, owns); err != nil {
		return graph.BeadRecord{}, corrupt(err)
	}
	return record, nil
}
