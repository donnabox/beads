package graphops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	graph "github.com/steveyegge/beads/graphops"
)

// These values are decoded private operands, never public cursors or authority.
// A protected caller must separately validate selected-state physical rows as
// exactly the live set. In particular, an afterPath is not a minted Cursor.
type pageWindow struct {
	afterPath string
	limit     int
}
type linkPageSelection struct {
	typeURL        string
	source, target *graph.Ref
}
type linkRowsPage struct {
	items    []graph.Link
	lastPath string
	hasMore  bool
}

func pageBudget(scope string, window pageWindow, limits readLimits) (*readBudget, error) {
	b, err := budgetFor(scope, limits)
	if err != nil {
		return nil, err
	}
	// Reserve one row for charged, validated lookahead. This is a private scan
	// bound, not an advertised page maximum or default. Compare before adding.
	if window.limit < 1 || window.limit >= limits.rows {
		return nil, fmt.Errorf("%w: page must leave room for lookahead", graph.ErrValidation)
	}
	if window.afterPath != "" {
		if err := graph.ValidateLinkPath(window.afterPath); err != nil {
			return nil, err
		}
	}
	return b, nil
}

func validatePageSelection(scope string, selection linkPageSelection) error {
	if selection.typeURL != "" {
		if err := graph.ValidateTypeURL(selection.typeURL); err != nil {
			return err
		}
	}
	for _, ref := range []*graph.Ref{selection.source, selection.target} {
		if ref == nil {
			continue
		}
		if ref.IsZero() {
			return fmt.Errorf("%w: zero endpoint filter", graph.ErrValidation)
		}
		parsed, err := graph.ParseRef(scope, ref.URL(scope), ref.Pin())
		if err != nil {
			return err
		}
		if !parsed.Equal(*ref) {
			return fmt.Errorf("%w: endpoint filter changes Scope classification", graph.ErrValidation)
		}
	}
	return nil
}

// Only these fixed internal column names can enter the SQL. Request values are
// parameters; pins never participate in endpoint identity.
func linkPageQuery(selection linkPageSelection, window pageWindow, valueBytes int) (string, []any) {
	clauses := []string{"path > ?"}
	args := []any{valueBytes, window.afterPath}
	if selection.typeURL != "" {
		clauses = append(clauses, "type_url = ?")
		args = append(args, selection.typeURL)
	}
	for _, endpoint := range []struct {
		name string
		ref  *graph.Ref
	}{{"source", selection.source}, {"target", selection.target}} {
		if endpoint.ref == nil {
			continue
		}
		if endpoint.ref.InScope() {
			clauses = append(clauses, endpoint.name+"_kind = 'in' AND "+endpoint.name+"_path = ?")
			args = append(args, endpoint.ref.Path())
		} else {
			clauses = append(clauses, endpoint.name+"_kind = 'ext' AND "+endpoint.name+"_url = ?")
			args = append(args, endpoint.ref.URI())
		}
	}
	args = append(args, window.limit+1)
	return "SELECT " + linkColumns + " FROM graph_links WHERE " + strings.Join(clauses, " AND ") + " ORDER BY path LIMIT ?", args
}

func checkPageLink(link graph.Link, selection linkPageSelection, window pageWindow, previous string) error {
	if (selection.typeURL != "" && link.TypeURL() != selection.typeURL) ||
		(selection.source != nil && !link.Source().SameURI(*selection.source)) ||
		(selection.target != nil && !link.Target().SameURI(*selection.target)) ||
		graph.CompareCodeUnits(link.Path(), window.afterPath) <= 0 ||
		(previous != "" && graph.CompareCodeUnits(link.Path(), previous) <= 0) {
		return corrupt(errors.New("Link page row violates selection or keyset order"))
	}
	return nil
}

func finishLinkPage(links []graph.Link, window pageWindow) linkRowsPage {
	result := linkRowsPage{items: links, hasMore: len(links) > window.limit}
	if result.hasMore {
		result.items = result.items[:window.limit]
	}
	if len(result.items) > 0 {
		result.lastPath = result.items[len(result.items)-1].Path()
	}
	return result
}

func pageRowsError(err error) error {
	if errors.Is(err, errRowOverflow) {
		return errors.Join(errCorrupt, err)
	}
	return err
}

func readLinkPageInTx(ctx context.Context, tx queryer, scope string, selection linkPageSelection, window pageWindow, limits readLimits) (linkRowsPage, error) {
	b, err := pageBudget(scope, window, limits)
	if err != nil {
		return linkRowsPage{}, err
	}
	if err := validatePageSelection(scope, selection); err != nil {
		return linkRowsPage{}, err
	}
	query, args := linkPageQuery(selection, window, limits.valueBytes)
	previous := ""
	links, err := readRows(ctx, tx, query, args, window.limit+1, func(rows *sql.Rows) (graph.Link, error) {
		var row linkRow
		if err := rows.Scan(linkScanTargets(&row)...); err != nil {
			return graph.Link{}, err
		}
		link, err := chargedLink(scope, row, b)
		if err != nil {
			return graph.Link{}, err
		}
		if err := checkPageLink(link, selection, window, previous); err != nil {
			return graph.Link{}, err
		}
		previous = link.Path()
		return link, nil
	})
	if err != nil {
		return linkRowsPage{}, pageRowsError(err)
	}
	return finishLinkPage(links, window), nil
}

// Candidate ordering and global lookahead LIMIT happen BEFORE the sentinel
// join. Branches are deliberately not separately limited: such limits require
// their own path ordering and at least limit+1 candidates. UNION deduplicates a
// self-loop before the global limit. Hydration has its own two-input lookup
// hint; the pinned plan oracle must prove it is honored, including the global
// candidate cap before hydration. Raw src values reach the existing guarded
// outer projection unchanged. Candidate work itself is not cost-bounded.
// This SQL is provisional: a future protected incident call must compose the
// allocation-anchor diagnosis into this same statement, never a sixth query.
func incidentPageQuery(path string, direction graph.Direction, window pageWindow, valueBytes int) (string, []any) {
	branch := func(endpoint string) string {
		return "SELECT path FROM graph_links WHERE " + endpoint + "_kind = 'in' AND " + endpoint + "_path = ? AND path > ?"
	}
	selection := branch("source")
	if direction == graph.DirectionIn {
		selection = branch("target")
	}
	args := []any{valueBytes, path, window.afterPath}
	if direction == graph.DirectionBoth {
		selection += " UNION " + branch("target")
		// The pinned server's SetOp walker omits UNION's own LIMIT from
		// prepared parameter counting. Keep the bound global ORDER/LIMIT
		// on an ordinary SELECT after deduplication and before the sentinel.
		selection = "SELECT path FROM (" + selection + ") candidates"
		args = append(args, path, window.afterPath)
	}
	args = append(args, window.limit+1, path)
	query := "SELECT b.path, l.candidate_path IS NOT NULL, l.path IS NOT NULL, " + joinedLinkColumns + " FROM graph_beads b LEFT JOIN (SELECT /*+ LEFT_OUTER_LOOKUP_JOIN(i,src) */ i.path AS candidate_path, src.path, src.type_url, src.revision, src.attribution_principal, src.attribution_status, src.properties, src.source_kind, src.source_path, src.source_url, src.source_pin, src.target_kind, src.target_path, src.target_url, src.target_pin FROM (" + selection + " ORDER BY path LIMIT ?) i LEFT JOIN graph_links src ON src.path = i.path) l ON TRUE WHERE b.path = ? ORDER BY l.candidate_path"
	return query, args
}

// With no joined Link, joinedLinkColumns yields only its explicit empty-string
// fallbacks and SQL NULLs. A contradictory payload is an engine/query-shape
// failure, not a new allocation or endpoint-liveness rule.
func emptyPageLinkProjection(row linkRow) bool {
	return row.path == "" && row.typeURL == "" && row.revision == "" &&
		row.principal == (sql.NullString{}) && row.attribution == (sql.NullString{}) &&
		row.properties == nil && row.propertiesLength == (sql.NullInt64{}) &&
		row.source == (endpointRow{}) && row.target == (endpointRow{})
}

func readIncidentPageInTx(ctx context.Context, tx queryer, scope, path string, direction graph.Direction, window pageWindow, limits readLimits) (linkRowsPage, error) {
	b, err := pageBudget(scope, window, limits)
	if err != nil {
		return linkRowsPage{}, err
	}
	if err := graph.ValidateBeadPath(path); err != nil {
		return linkRowsPage{}, err
	}
	if !direction.Valid() {
		return linkRowsPage{}, fmt.Errorf("%w: invalid direction", graph.ErrValidation)
	}
	query, args := incidentPageQuery(path, direction, window, limits.valueBytes)
	previous := ""
	type incidentRow struct {
		link    graph.Link
		present bool
	}
	rows, err := readRows(ctx, tx, query, args, window.limit+1, func(rows *sql.Rows) (incidentRow, error) {
		var anchor string
		var candidate, present bool
		var row linkRow
		targets := append([]any{&anchor, &candidate, &present}, linkScanTargets(&row)...)
		if err := rows.Scan(targets...); err != nil {
			return incidentRow{}, err
		}
		if anchor != path || candidate != present {
			return incidentRow{}, corrupt(errors.New("incident page anchor or candidate mismatch"))
		}
		// Charge each projected copy: this budget bounds transferred/decoded
		// bytes, including the repeated anchor and the lookahead row.
		if err := b.charge([]byte(anchor)); err != nil {
			return incidentRow{}, err
		}
		if !present {
			if !emptyPageLinkProjection(row) {
				return incidentRow{}, corrupt(errors.New("payload on empty incident page sentinel"))
			}
			return incidentRow{}, nil
		}
		link, err := chargedLink(scope, row, b)
		if err != nil {
			return incidentRow{}, err
		}
		outbound := link.Source().InScope() && link.Source().Path() == path
		inbound := link.Target().InScope() && link.Target().Path() == path
		if (direction == graph.DirectionIn && !inbound) || (direction == graph.DirectionOut && !outbound) || (direction == graph.DirectionBoth && !inbound && !outbound) {
			return incidentRow{}, corrupt(errors.New("unrelated incident page Link"))
		}
		if err := checkPageLink(link, linkPageSelection{}, window, previous); err != nil {
			return incidentRow{}, err
		}
		previous = link.Path()
		return incidentRow{link: link, present: true}, nil
	})
	if err != nil {
		return linkRowsPage{}, pageRowsError(err)
	}
	if len(rows) == 0 {
		return linkRowsPage{}, errAbsent
	}
	var links []graph.Link
	for _, row := range rows {
		if !row.present {
			if len(rows) != 1 {
				return linkRowsPage{}, corrupt(errors.New("mixed empty incident page"))
			}
		} else {
			links = append(links, row.link)
		}
	}
	return finishLinkPage(links, window), nil
}
