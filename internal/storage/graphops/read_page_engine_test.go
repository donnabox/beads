//go:build cgo

package graphops

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	graph "github.com/steveyegge/beads/graphops"
)

type pageFixtureLink struct{ path, typeURL, source, target, revision string }

// Shared by the existing parent-owned embedded and managed fixture lifecycles.
// All added rows live in this one rollback-only transaction, after the original
// positive read transaction has closed. This is not a schema/authority install.
func verifyEnginePageControls(t *testing.T, ctx context.Context, db fixtureDB) {
	t.Helper()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()
	// As in the allocation fixture, bound EXPLAIN is an embedded diagnostic.
	// Managed MySQL reports zero prepared parameters for EXPLAIN; all actual
	// page reads below still run with their unchanged bound operands.
	embedded := false
	switch db.(type) {
	case *sql.DB:
		embedded = true
	case *sql.Conn:
		fmt.Println("GRAPH_PAGE_PLAN_DIAGNOSTIC_OMITTED managed prepared EXPLAIN unsupported; page semantic controls remain mandatory")
	default:
		t.Fatal("unknown page fixture database leg")
	}
	capturePlan := func(name, after, query string, args []any) {
		if embedded {
			capturePageExplain(t, ctx, tx, name, after, query, args)
		}
	}
	scope := scopeFromFixtureTx(t, ctx, tx)
	typeA, typeLower := relationType+"/A", relationType+"/a"
	endpoint, err := graph.NewEndpointConstraint(nil, graph.ExternalOpaque)
	if err != nil {
		t.Fatal(err)
	}
	for _, typ := range []string{typeA, typeLower} {
		descriptor, err := graph.NewTypeDescriptor(graph.TypeDescriptorSpec{ID: typ, Name: "Page relation", Describes: graph.KindLink, Source: &endpoint, Target: &endpoint})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := tx.ExecContext(ctx, "INSERT INTO graph_type_descriptors (url, descriptor, fingerprint, installed_seq, installed_at, last_authority_id, last_epoch) VALUES (?, ?, ?, 1, '2026-09-17 00:00:00', ?, 1)", typ, descriptor.CanonicalJSON(), descriptor.Fingerprint(), strings.Repeat("a", 32)); err != nil {
			t.Fatal(err)
		}
	}
	anchor, other, empty := "beads/page-anchor", "beads/page-other", "beads/page-empty"
	for _, path := range []string{anchor, other, empty} {
		if _, err := tx.ExecContext(ctx, "INSERT INTO graph_beads (path,type_url,revision,properties,last_authority_id,last_epoch,created_at,updated_at) VALUES (?,?,?,?,?,1,'2026-09-17 00:00:00','2026-09-17 00:00:00')", path, memoryType, graph.MintRevision().String(), []byte("{}"), strings.Repeat("a", 32)); err != nil {
			t.Fatal(err)
		}
	}
	// More than pageLimit+1 outgoing paths precede the first incoming path;
	// subsequent directions interleave. The self-loop must appear only once.
	fixtures := []pageFixtureLink{
		{"links/page/00", typeA, anchor, other, ""},
		{"links/page/01", typeA, anchor, other, ""},
		{"links/page/02", typeA, anchor, other, ""},
		{"links/page/03", typeLower, anchor, other, ""},
		{"links/page/04", typeA, other, anchor, ""},
		{"links/page/05", typeA, anchor, anchor, ""},
		{"links/page/06", typeLower, anchor, other, ""},
		{"links/page/07", typeA, other, anchor, ""},
		{"links/page/08", typeA, "urn:external:source", anchor, ""},
		{"links/page/09", typeLower, anchor, "urn:external:target", ""},
		{"links/page/A", typeA, anchor, other, ""},
		{"links/page/a", typeA, other, anchor, ""},
		{"links/page/caf%C3%A9", typeA, anchor, other, ""},
		{"links/page/nested/x", typeLower, other, anchor, ""},
	}
	expected := map[string]graph.Link{}
	for i := range fixtures {
		f := &fixtures[i]
		f.revision = graph.MintRevision().String()
		source, err := graph.ParseRef(scope, f.source, "source stored pin")
		if err != nil {
			t.Fatal(err)
		}
		target, err := graph.ParseRef(scope, f.target, "target stored pin")
		if err != nil {
			t.Fatal(err)
		}
		sk, sp, su := pageFixtureEndpoint(source)
		tk, tp, tu := pageFixtureEndpoint(target)
		if _, err := tx.ExecContext(ctx, "INSERT INTO graph_links (path,type_url,revision,properties,source_kind,source_path,source_url,source_pin,target_kind,target_path,target_url,target_pin,last_authority_id,last_epoch,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,1,'2026-09-17 00:00:00','2026-09-17 00:00:00')", f.path, f.typeURL, f.revision, []byte("{}"), sk, sp, su, source.Pin(), tk, tp, tu, target.Pin(), strings.Repeat("a", 32)); err != nil {
			t.Fatalf("seed page Link %s: %v", f.path, err)
		}
		revision, err := graph.NewRevision(f.revision)
		if err != nil {
			t.Fatal(err)
		}
		properties, err := graph.NewProperties([]byte("{}"))
		if err != nil {
			t.Fatal(err)
		}
		link, err := graph.NewLink(graph.LinkSpec{Path: f.path, TypeURL: f.typeURL, Revision: revision, Properties: properties, Source: source, Target: target})
		if err != nil {
			t.Fatal(err)
		}
		expected[f.path] = link
	}
	ref := func(raw string) *graph.Ref {
		v, err := graph.ParseRef(scope, raw, "filter pin differs")
		if err != nil {
			t.Fatal(err)
		}
		return &v
	}
	paths := func(indices ...int) []string {
		var out []string
		for _, i := range indices {
			out = append(out, fixtures[i].path)
		}
		return out
	}
	cases := []struct {
		name      string
		selection linkPageSelection
		want      []string
	}{
		{"typed", linkPageSelection{typeURL: typeA}, paths(0, 1, 2, 4, 5, 7, 8, 10, 11, 12)},
		{"case distinct type", linkPageSelection{typeURL: typeLower}, paths(3, 6, 9, 13)},
		{"typed local source", linkPageSelection{typeURL: typeA, source: ref(anchor)}, paths(0, 1, 2, 5, 10, 12)},
		{"untyped local source", linkPageSelection{source: ref(anchor)}, paths(0, 1, 2, 3, 5, 6, 9, 10, 12)},
		{"typed local target", linkPageSelection{typeURL: typeA, target: ref(anchor)}, paths(4, 5, 7, 8, 11)},
		{"untyped local target", linkPageSelection{target: ref(anchor)}, paths(4, 5, 7, 8, 11, 13)},
		{"external source", linkPageSelection{source: ref("urn:external:source")}, paths(8)},
		{"external target", linkPageSelection{target: ref("urn:external:target")}, paths(9)},
		{"combined local", linkPageSelection{source: ref(anchor), target: ref(other)}, paths(0, 1, 2, 3, 6, 10, 12)},
		{"type plus both endpoints", linkPageSelection{typeURL: typeA, source: ref(anchor), target: ref(other)}, paths(0, 1, 2, 10, 12)},
	}
	for _, tc := range cases {
		verifyFixturePageTraversal(t, tc.name, tc.want, expected, func(after string) (linkRowsPage, error) {
			window := pageWindow{afterPath: after, limit: 2}
			query, args := linkPageQuery(tc.selection, window, fixtureLimits.valueBytes)
			capturePlan(tc.name, after, query, args)
			return readLinkPageInTx(ctx, tx, scope, tc.selection, window, fixtureLimits)
		})
	}
	// Unfiltered includes the original persisted receipt's three Links. Read only
	// those exact known paths for value expectations; membership remains explicit.
	all := paths(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13)
	for _, path := range []string{"links/plan-decision", "links/decision-finding", "links/external-plan"} {
		link, err := readLinkInTx(ctx, tx, scope, path, fixtureLimits)
		if err != nil {
			t.Fatal(err)
		}
		expected[path] = link
		all = append(all, path)
	}
	slices.SortFunc(all, graph.CompareCodeUnits)
	verifyFixturePageTraversal(t, "unfiltered", all, expected, func(after string) (linkRowsPage, error) {
		window := pageWindow{afterPath: after, limit: 2}
		query, args := linkPageQuery(linkPageSelection{}, window, fixtureLimits.valueBytes)
		capturePlan("unfiltered", after, query, args)
		return readLinkPageInTx(ctx, tx, scope, linkPageSelection{}, window, fixtureLimits)
	})
	for _, tc := range []struct {
		name      string
		direction graph.Direction
		want      []string
	}{
		{"incident out", graph.DirectionOut, paths(0, 1, 2, 3, 5, 6, 9, 10, 12)},
		{"incident in", graph.DirectionIn, paths(4, 5, 7, 8, 11, 13)},
		{"incident both", graph.DirectionBoth, paths(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13)},
	} {
		verifyFixturePageTraversal(t, tc.name, tc.want, expected, func(after string) (linkRowsPage, error) {
			window := pageWindow{afterPath: after, limit: 2}
			query, args := incidentPageQuery(anchor, tc.direction, window, fixtureLimits.valueBytes)
			capturePlan(tc.name, after, query, args)
			return readIncidentPageInTx(ctx, tx, scope, anchor, tc.direction, window, fixtureLimits)
		})
	}
	// A continuation is a keyset boundary, not a requirement that the path exists.
	gap := pageWindow{afterPath: "links/page/035", limit: 2}
	query, args := linkPageQuery(linkPageSelection{}, gap, fixtureLimits.valueBytes)
	capturePlan("unfiltered missing boundary", gap.afterPath, query, args)
	page, err := readLinkPageInTx(ctx, tx, scope, linkPageSelection{}, gap, fixtureLimits)
	if err != nil || len(page.items) != 2 || page.items[0].Path() != "links/page/04" || page.items[1].Path() != "links/page/05" || page.lastPath != "links/page/05" || !page.hasMore {
		t.Fatalf("missing-boundary page=%+v error=%v", page, err)
	}
	for _, direction := range []graph.Direction{graph.DirectionIn, graph.DirectionOut, graph.DirectionBoth} {
		for _, anchorPath := range []string{empty, "beads/page/missing"} {
			for _, after := range []string{"", "links/zzzz"} {
				window := pageWindow{afterPath: after, limit: 2}
				query, args := incidentPageQuery(anchorPath, direction, window, fixtureLimits.valueBytes)
				capturePlan(fmt.Sprintf("incident anchor=%s direction=%v", anchorPath, direction), after, query, args)
				page, err := readIncidentPageInTx(ctx, tx, scope, anchorPath, direction, window, fixtureLimits)
				if (anchorPath == empty && err != nil) || (anchorPath != empty && !errors.Is(err, errAbsent)) {
					t.Fatalf("anchor=%s direction=%v error=%v", anchorPath, direction, err)
				}
				requireEmptyPage(t, page)
			}
		}
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	fmt.Println("GRAPH_PAGE_CONTROLS_RECORDED rollback-only single-transaction traversal; plans require independent index qualification")
}

func pageFixtureEndpoint(ref graph.Ref) (string, any, any) {
	if ref.InScope() {
		return "in", ref.Path(), nil
	}
	return "ext", nil, ref.URI()
}

func verifyFixturePageTraversal(t *testing.T, name string, want []string, expected map[string]graph.Link, read func(string) (linkRowsPage, error)) {
	t.Helper()
	after := ""
	offset := 0
	for attempt := 0; attempt <= len(want)+1; attempt++ {
		page, err := read(after)
		if err != nil {
			t.Fatalf("page %s after=%q: %v", name, after, err)
		}
		count := min(2, len(want)-offset)
		if len(page.items) != count {
			t.Fatalf("page %s count=%d want=%d", name, len(page.items), count)
		}
		for i, link := range page.items {
			path := want[offset+i]
			w := expected[path]
			if link.Path() != path || link.TypeURL() != w.TypeURL() || link.Revision() != w.Revision() || link.Properties().String() != w.Properties().String() || !link.Source().Equal(w.Source()) || !link.Target().Equal(w.Target()) {
				t.Fatalf("page %s returned altered/wrong Link %s wanted %s", name, link.Path(), path)
			}
		}
		offset += count
		more := offset < len(want)
		last := ""
		if count > 0 {
			last = want[offset-1]
		}
		if page.hasMore != more || page.lastPath != last {
			t.Fatalf("page %s continuation=%+v expected last=%q more=%t", name, page, last, more)
		}
		if !more {
			// Exercise a continuation past the final row even when the public-facing
			// result would have been terminal. No existence lookup of afterPath.
			if last != "" {
				terminal, err := read(last)
				if err != nil {
					t.Fatal(err)
				}
				requireEmptyPage(t, terminal)
			}
			return
		}
		after = last
	}
	t.Fatalf("page %s failed finite traversal", name)
}

func capturePageExplain(t *testing.T, ctx context.Context, tx *sql.Tx, name, after, query string, args []any) {
	t.Helper()
	rows, err := tx.QueryContext(ctx, "EXPLAIN FORMAT=tree "+query, args...)
	if err != nil {
		t.Fatalf("EXPLAIN %s: %v", name, err)
	}
	columns, err := rows.Columns()
	if err != nil || len(columns) != 1 || columns[0] != "plan" {
		_ = rows.Close()
		t.Fatalf("unexpected EXPLAIN tree columns %v: %v", columns, err)
	}
	var plan [][]string
	for rows.Next() {
		values := make([]any, len(columns))
		targets := make([]any, len(columns))
		for i := range values {
			targets[i] = &values[i]
		}
		if err := rows.Scan(targets...); err != nil {
			_ = rows.Close()
			t.Fatal(err)
		}
		var line []string
		for _, v := range values {
			if raw, ok := v.([]byte); ok {
				line = append(line, string(raw))
			} else {
				line = append(line, fmt.Sprint(v))
			}
		}
		if len(line[0]) > 4096 {
			_ = rows.Close()
			t.Fatal("EXPLAIN line exceeds fixture bound")
		}
		plan = append(plan, line)
		if len(plan) > 256 {
			_ = rows.Close()
			t.Fatal("EXPLAIN exceeds fixture bound")
		}
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		t.Fatal(err)
	}
	if err := rows.Close(); err != nil {
		t.Fatal(err)
	}
	if len(plan) == 0 {
		t.Fatal("empty EXPLAIN evidence")
	}
	raw, err := json.Marshal(struct {
		Name, After, Query string
		Arguments          []any
		Columns            []string
		Plan               [][]string
	}{name, after, query, args, columns, plan})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("GRAPH_PAGE_EXPLAIN %s\n", raw)
}
