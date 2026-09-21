package graphops

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/dolthub/vitess/go/vt/sqlparser"
	graph "github.com/steveyegge/beads/graphops"
)

func pageRow(path string) linkRow { r := validLinkRow(); r.path = path; return r }
func pageMockRows(paths ...string) *sqlmock.Rows {
	rows := sqlmock.NewRows(allLinkColumns)
	for _, path := range paths {
		rows.AddRow(linkValues(pageRow(path))...)
	}
	return rows
}
func expectPage(m sqlmock.Sqlmock, after string, limit int, rows *sqlmock.Rows) {
	m.ExpectQuery(regexp.QuoteMeta("SELECT "+linkColumns+" FROM graph_links WHERE path > ? ORDER BY path LIMIT ?")).WithArgs(fixtureLimits.valueBytes, after, limit+1).WillReturnRows(rows).RowsWillBeClosed()
}
func requireEmptyPage(t *testing.T, page linkRowsPage) {
	t.Helper()
	if len(page.items) != 0 || page.hasMore || page.lastPath != "" {
		t.Fatalf("partial page escaped: %+v", page)
	}
}
func TestLinkPageBoundaries(t *testing.T) {
	for _, count := range []int{0, 1, 2, 3} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			tx, m := mockTx(t)
			rows := pageMockRows()
			for i := 0; i < count; i++ {
				rows.AddRow(linkValues(pageRow(fmt.Sprintf("links/%d", i)))...)
			}
			expectPage(m, "", 2, rows)
			page, err := readLinkPageInTx(t.Context(), tx, fixtureScope, linkPageSelection{}, pageWindow{limit: 2}, fixtureLimits)
			if err != nil {
				t.Fatal(err)
			}
			if len(page.items) != min(count, 2) || page.hasMore != (count > 2) {
				t.Fatalf("page=%+v", page)
			}
			last := ""
			if count > 0 {
				last = fmt.Sprintf("links/%d", min(count, 2)-1)
			}
			if page.lastPath != last {
				t.Fatalf("last=%q want=%q", page.lastPath, last)
			}
		})
	}
}
func TestLinkPageContinuations(t *testing.T) {
	tx, m := mockTx(t)
	expectPage(m, "", 2, pageMockRows("links/A", "links/B", "links/caf%C3%A9"))
	expectPage(m, "links/B", 2, pageMockRows("links/caf%C3%A9", "links/nested/x"))
	expectPage(m, "links/nested/x", 2, pageMockRows())
	var got []string
	for _, after := range []string{"", "links/B", "links/nested/x"} {
		page, err := readLinkPageInTx(t.Context(), tx, fixtureScope, linkPageSelection{}, pageWindow{afterPath: after, limit: 2}, fixtureLimits)
		if err != nil {
			t.Fatal(err)
		}
		for _, link := range page.items {
			got = append(got, link.Path())
		}
		if page.hasMore != (after == "") {
			t.Fatalf("wrong terminality: %+v", page)
		}
	}
	if strings.Join(got, ",") != "links/A,links/B,links/caf%C3%A9,links/nested/x" {
		t.Fatal(got)
	}
}
func TestLinkPageInvalidOperandsIssueNoQuery(t *testing.T) {
	local, err := graph.NewInScopeRef("beads/plan", "")
	if err != nil {
		t.Fatal(err)
	}
	reclassified, err := graph.ParseRef("https://elsewhere.example/", fixtureScope+"beads/plan", "")
	if err != nil {
		t.Fatal(err)
	}
	zero := graph.Ref{}
	cases := []struct {
		name, scope string
		selection   linkPageSelection
		window      pageWindow
		limits      readLimits
	}{
		{"zero limit", fixtureScope, linkPageSelection{}, pageWindow{}, fixtureLimits},
		{"negative limit", fixtureScope, linkPageSelection{}, pageWindow{limit: -1}, fixtureLimits},
		{"no lookahead reserve", fixtureScope, linkPageSelection{}, pageWindow{limit: fixtureLimits.rows}, fixtureLimits},
		{"rows one", fixtureScope, linkPageSelection{}, pageWindow{limit: 1}, readLimits{rows: 1, bytes: 8192, valueBytes: 8192}},
		{"wrong path", fixtureScope, linkPageSelection{}, pageWindow{afterPath: "beads/plan", limit: 1}, fixtureLimits},
		{"bad scope", "http://bad/local-test/", linkPageSelection{}, pageWindow{limit: 1}, fixtureLimits},
		{"zero endpoint", fixtureScope, linkPageSelection{source: &zero}, pageWindow{limit: 1}, fixtureLimits},
		{"reclassified external", fixtureScope, linkPageSelection{source: &reclassified, target: &local}, pageWindow{limit: 1}, fixtureLimits},
	}
	for _, typ := range []string{"https://GRAPH.example/types/x", "https://graph.example:443/types/x", relationType + " ", relationType + "#x", "*"} {
		cases = append(cases, struct {
			name, scope string
			selection   linkPageSelection
			window      pageWindow
			limits      readLimits
		}{typ, fixtureScope, linkPageSelection{typeURL: typ}, pageWindow{limit: 1}, fixtureLimits})
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tx, _ := mockTx(t)
			page, err := readLinkPageInTx(t.Context(), tx, tc.scope, tc.selection, tc.window, tc.limits)
			if !errors.Is(err, graph.ErrValidation) {
				t.Fatalf("error=%v", err)
			}
			requireEmptyPage(t, page)
		})
	}
}
func TestLinkPageStructuralPredicates(t *testing.T) {
	// Every nonempty combination; identity ignores independently different pins.
	source, err := graph.NewInScopeRef("beads/plan", "requested pin")
	if err != nil {
		t.Fatal(err)
	}
	target, err := graph.ParseRef(fixtureScope, "urn:external:target", "requested target pin")
	if err != nil {
		t.Fatal(err)
	}
	for mask := 0; mask < 8; mask++ {
		t.Run(fmt.Sprint(mask), func(t *testing.T) {
			selection := linkPageSelection{}
			where := "path > ?"
			args := []driver.Value{fixtureLimits.valueBytes, "links/0"}
			if mask&1 != 0 {
				selection.typeURL = relationType
				where += " AND type_url = ?"
				args = append(args, relationType)
			}
			if mask&2 != 0 {
				selection.source = &source
				where += " AND source_kind = 'in' AND source_path = ?"
				args = append(args, "beads/plan")
			}
			if mask&4 != 0 {
				selection.target = &target
				where += " AND target_kind = 'ext' AND target_url = ?"
				args = append(args, "urn:external:target")
			}
			args = append(args, 2)
			r := pageRow("links/a")
			r.source.pin = present("stored pin")
			r.target = endpointRow{kind: "ext", url: present("urn:external:target"), pin: present("stored target pin")}
			tx, m := mockTx(t)
			m.ExpectQuery(regexp.QuoteMeta("SELECT " + linkColumns + " FROM graph_links WHERE " + where + " ORDER BY path LIMIT ?")).WithArgs(args...).WillReturnRows(sqlmock.NewRows(allLinkColumns).AddRow(linkValues(r)...)).RowsWillBeClosed()
			page, err := readLinkPageInTx(t.Context(), tx, fixtureScope, selection, pageWindow{afterPath: "links/0", limit: 1}, fixtureLimits)
			if err != nil || len(page.items) != 1 || page.items[0].Source().Pin() != "stored pin" || page.items[0].Target().Pin() != "stored target pin" {
				t.Fatalf("page=%+v error=%v", page, err)
			}
		})
	}
	// Symmetric external-source/local-target selector is independently pinned.
	ext, err := graph.ParseRef(fixtureScope, "urn:external:source", "")
	if err != nil {
		t.Fatal(err)
	}
	local, err := graph.NewInScopeRef("beads/decision", "")
	if err != nil {
		t.Fatal(err)
	}
	tx, m := mockTx(t)
	r := pageRow("links/x")
	r.source = endpointRow{kind: "ext", url: present("urn:external:source")}
	m.ExpectQuery(regexp.QuoteMeta("SELECT "+linkColumns+" FROM graph_links WHERE path > ? AND source_kind = 'ext' AND source_url = ? AND target_kind = 'in' AND target_path = ? ORDER BY path LIMIT ?")).WithArgs(fixtureLimits.valueBytes, "", "urn:external:source", "beads/decision", 2).WillReturnRows(sqlmock.NewRows(allLinkColumns).AddRow(linkValues(r)...)).RowsWillBeClosed()
	if _, err := readLinkPageInTx(t.Context(), tx, fixtureScope, linkPageSelection{source: &ext, target: &local}, pageWindow{limit: 1}, fixtureLimits); err != nil {
		t.Fatal(err)
	}
}
func TestLinkPageRefusesFaultyRowsAndLookahead(t *testing.T) {
	cases := []struct {
		name  string
		rows  func() *sqlmock.Rows
		after string
		want  error
	}{
		{"duplicate", func() *sqlmock.Rows { return pageMockRows("links/a", "links/a") }, "", errCorrupt},
		{"reverse", func() *sqlmock.Rows { return pageMockRows("links/b", "links/a") }, "", errCorrupt},
		{"equal boundary", func() *sqlmock.Rows { return pageMockRows("links/b") }, "links/b", errCorrupt},
		{"below boundary", func() *sqlmock.Rows { return pageMockRows("links/a") }, "links/b", errCorrupt},
		{"past SQL cap", func() *sqlmock.Rows { return pageMockRows("links/a", "links/b", "links/c") }, "", errCorrupt},
		{"malformed lookahead", func() *sqlmock.Rows {
			r := pageRow("links/b")
			r.properties = []byte("{")
			return pageMockRows("links/a").AddRow(linkValues(r)...)
		}, "", errCorrupt},
		{"oversized lookahead", func() *sqlmock.Rows {
			r := pageRow("links/b")
			r.properties = []byte(strings.Repeat("x", fixtureLimits.valueBytes+1))
			return pageMockRows("links/a").AddRow(linkValues(r)...)
		}, "", errBudget},
		{"negative length", func() *sqlmock.Rows {
			values := linkValues(pageRow("links/b"))
			values[5] = nil
			values[6] = int64(-1)
			return pageMockRows("links/a").AddRow(values...)
		}, "", errBudget},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tx, m := mockTx(t)
			expectPage(m, tc.after, 1, tc.rows())
			page, err := readLinkPageInTx(t.Context(), tx, fixtureScope, linkPageSelection{}, pageWindow{afterPath: tc.after, limit: 1}, fixtureLimits)
			if !errors.Is(err, tc.want) || errors.Is(err, graph.ErrValidation) || errors.Is(err, graph.ErrNotFound) || (tc.want == errCorrupt && errors.Is(err, errBudget)) {
				t.Fatalf("error=%v", err)
			}
			requireEmptyPage(t, page)
		})
	}
}
func TestLinkPageRejectsMismatchedSelection(t *testing.T) {
	ref, err := graph.NewInScopeRef("beads/other", "")
	if err != nil {
		t.Fatal(err)
	}
	for _, selection := range []linkPageSelection{{typeURL: relationType + "?v=A"}, {source: &ref}, {target: &ref}} {
		t.Run(fmt.Sprintf("%+v", selection), func(t *testing.T) {
			tx, m := mockTx(t)
			m.ExpectQuery("SELECT .* FROM graph_links WHERE").WillReturnRows(pageMockRows("links/a")).RowsWillBeClosed()
			page, err := readLinkPageInTx(t.Context(), tx, fixtureScope, selection, pageWindow{limit: 1}, fixtureLimits)
			if !errors.Is(err, errCorrupt) {
				t.Fatal(err)
			}
			requireEmptyPage(t, page)
		})
	}
}
func TestLinkPagePropagatesQueryRowScanAndCloseErrors(t *testing.T) {
	boom := errors.New("page IO")
	for _, kind := range []string{"query", "row", "scan", "close"} {
		t.Run(kind, func(t *testing.T) {
			tx, m := mockTx(t)
			q := m.ExpectQuery("SELECT .* FROM graph_links WHERE")
			switch kind {
			case "query":
				q.WillReturnError(boom)
			case "row":
				q.WillReturnRows(pageMockRows("links/a").RowError(0, boom)).RowsWillBeClosed()
			case "scan":
				q.WillReturnRows(sqlmock.NewRows([]string{"bad"}).AddRow("bad")).RowsWillBeClosed()
			case "close":
				q.WillReturnRows(pageMockRows().CloseError(boom)).RowsWillBeClosed()
			}
			page, err := readLinkPageInTx(t.Context(), tx, fixtureScope, linkPageSelection{}, pageWindow{limit: 1}, fixtureLimits)
			if err == nil || (kind != "scan" && !errors.Is(err, boom)) {
				t.Fatalf("error=%v", err)
			}
			requireEmptyPage(t, page)
		})
	}
}

func incidentPageRows() *sqlmock.Rows {
	return sqlmock.NewRows(incidentAllocationColumns())
}
func addIncidentPageRow(rows *sqlmock.Rows, anchor string, candidate, present bool, r linkRow) *sqlmock.Rows {
	return rows.AddRow(incidentLiveValues(anchor, candidate, present, r)...)
}
func expectIncidentPage(m sqlmock.Sqlmock, direction graph.Direction, after string, limit int, rows *sqlmock.Rows) {
	branch := "SELECT path FROM graph_links WHERE source_kind = 'in' AND source_path = ? AND path > ?"
	args := []driver.Value{fixtureLimits.valueBytes, "beads/plan", "beads/plan", "beads/plan", "beads/plan", after}
	if direction == graph.DirectionIn {
		branch = "SELECT path FROM graph_links WHERE target_kind = 'in' AND target_path = ? AND path > ?"
	}
	if direction == graph.DirectionBoth {
		branch += " UNION SELECT path FROM graph_links WHERE target_kind = 'in' AND target_path = ? AND path > ?"
		branch = "SELECT path FROM (" + branch + ") candidates"
		args = append(args, "beads/plan", after)
	}
	query := "SELECT req.requested_path, a.path, CAST(a.resource_kind AS CHAR), CAST(a.state AS CHAR), b.path IS NOT NULL, COALESCE(b.path, ''), l.candidate_path IS NOT NULL, l.path IS NOT NULL, " + joinedLinkColumns + " FROM (SELECT ? AS requested_path) req LEFT JOIN (SELECT path, resource_kind, state FROM graph_allocations WHERE path = ? LIMIT 2) a ON TRUE LEFT JOIN (SELECT path FROM graph_beads WHERE path = ? LIMIT 2) b ON TRUE LEFT JOIN (SELECT /*+ LEFT_OUTER_LOOKUP_JOIN(i,src) */ i.path AS candidate_path, src.path, src.type_url, src.revision, src.attribution_principal, src.attribution_status, src.properties, src.source_kind, src.source_path, src.source_url, src.source_pin, src.target_kind, src.target_path, src.target_url, src.target_pin FROM (" + branch + " ORDER BY path LIMIT ?) i LEFT JOIN graph_links src ON src.path = i.path) l ON b.path IS NOT NULL ORDER BY l.candidate_path"
	args = append(args, limit+1)
	m.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(args...).WillReturnRows(rows).RowsWillBeClosed()
}
func TestIncidentPageDirectionsAndSelfLoop(t *testing.T) {
	for _, direction := range []graph.Direction{graph.DirectionBoth, graph.DirectionIn, graph.DirectionOut} {
		t.Run(fmt.Sprint(direction), func(t *testing.T) {
			r := pageRow("links/a")
			if direction == graph.DirectionIn {
				r.source.path = present("beads/decision")
				r.target.path = present("beads/plan")
			}
			self := pageRow("links/b")
			self.target.path = present("beads/plan")
			rows := addIncidentPageRow(incidentPageRows(), "beads/plan", true, true, r)
			addIncidentPageRow(rows, "beads/plan", true, true, self)
			tx, m := mockTx(t)
			expectIncidentPage(m, direction, "links/0", 1, rows)
			page, err := readIncidentPageInTx(preconditionContext(t), tx, fixtureScope, "beads/plan", direction, pageWindow{afterPath: "links/0", limit: 1}, fixtureLimits)
			if err != nil || len(page.items) != 1 || !page.hasMore || page.lastPath != "links/a" {
				t.Fatalf("page=%+v error=%v", page, err)
			}
		})
	}
}
func TestIncidentPageAbsentEmptyAndCorruption(t *testing.T) {
	for _, kind := range []string{"absent", "empty", "empty with payload", "missing link", "missing candidate", "wrong anchor", "mixed", "unrelated", "equal boundary", "below boundary", "overflow"} {
		t.Run(kind, func(t *testing.T) {
			rows := incidentPageRows()
			r := pageRow("links/b")
			want := errCorrupt
			switch kind {
			case "absent":
				want = errCorrupt // no request singleton is an impossible projection
			case "empty":
				addIncidentPageRow(rows, "beads/plan", false, false, linkRow{})
				want = nil
			case "empty with payload":
				addIncidentPageRow(rows, "beads/plan", false, false, r)
			case "missing link":
				addIncidentPageRow(rows, "beads/plan", true, false, linkRow{})
			case "missing candidate":
				addIncidentPageRow(rows, "beads/plan", false, true, r)
			case "wrong anchor":
				addIncidentPageRow(rows, "beads/other", true, true, r)
			case "mixed":
				addIncidentPageRow(rows, "beads/plan", false, false, linkRow{})
				addIncidentPageRow(rows, "beads/plan", true, true, r)
			case "unrelated":
				r.source.path = present("beads/other")
				addIncidentPageRow(rows, "beads/plan", true, true, r)
			case "equal boundary":
				r.path = "links/a"
				addIncidentPageRow(rows, "beads/plan", true, true, r)
			case "below boundary":
				r.path = "links/0"
				addIncidentPageRow(rows, "beads/plan", true, true, r)
			case "overflow":
				for _, path := range []string{"links/b", "links/c", "links/d"} {
					addIncidentPageRow(rows, "beads/plan", true, true, pageRow(path))
				}
			}
			tx, m := mockTx(t)
			expectIncidentPage(m, graph.DirectionOut, "links/a", 1, rows)
			page, err := readIncidentPageInTx(preconditionContext(t), tx, fixtureScope, "beads/plan", graph.DirectionOut, pageWindow{afterPath: "links/a", limit: 1}, fixtureLimits)
			if (want == nil && err != nil) || (want != nil && !errors.Is(err, want)) {
				t.Fatalf("error=%v want=%v", err, want)
			}
			requireEmptyPage(t, page)
		})
	}
}
func TestPagePrivateCapAndAggregateLookahead(t *testing.T) {
	r := pageRow("links/a")
	r.propertiesLength = sql.NullInt64{Int64: int64(len(r.properties)), Valid: true}
	b, _ := budgetFor(fixtureScope, fixtureLimits)
	if _, err := chargedLink(fixtureScope, r, b); err != nil {
		t.Fatal(err)
	}
	one := fixtureLimits.bytes - b.remaining
	limits := readLimits{rows: 2, bytes: one*2 - 1, valueBytes: one}
	tx, m := mockTx(t)
	m.ExpectQuery("SELECT .* FROM graph_links WHERE").WithArgs(one, "", 2).WillReturnRows(pageMockRows("links/a", "links/b")).RowsWillBeClosed()
	page, err := readLinkPageInTx(t.Context(), tx, fixtureScope, linkPageSelection{}, pageWindow{limit: 1}, limits)
	if !errors.Is(err, errBudget) {
		t.Fatal(err)
	}
	requireEmptyPage(t, page)
	// Maximum legal private page retains lookahead room; no public limit implied.
	tx, m = mockTx(t)
	m.ExpectQuery("SELECT .* FROM graph_links WHERE").WithArgs(fixtureLimits.valueBytes, "", fixtureLimits.rows).WillReturnRows(pageMockRows()).RowsWillBeClosed()
	if _, err := readLinkPageInTx(t.Context(), tx, fixtureScope, linkPageSelection{}, pageWindow{limit: fixtureLimits.rows - 1}, fixtureLimits); err != nil {
		t.Fatal(err)
	}
}

// No engine or authority is involved in this seam.
var _ queryer = (*sql.Tx)(nil)

func TestIncidentPageInvalidOperandsIssueNoQuery(t *testing.T) {
	for _, tc := range []struct {
		name, path string
		direction  graph.Direction
		window     pageWindow
	}{
		{"invalid path", "links/x", graph.DirectionBoth, pageWindow{limit: 1}},
		{"invalid direction", "beads/plan", graph.Direction(255), pageWindow{limit: 1}},
		{"invalid window", "beads/plan", graph.DirectionBoth, pageWindow{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tx, _ := mockTx(t)
			page, err := readIncidentPageInTx(preconditionContext(t), tx, fixtureScope, tc.path, tc.direction, tc.window, fixtureLimits)
			if !errors.Is(err, graph.ErrValidation) {
				t.Fatalf("error=%v", err)
			}
			requireEmptyPage(t, page)
		})
	}
}
func TestIncidentPageLookaheadAndDuplicateFailures(t *testing.T) {
	for _, kind := range []string{"malformed", "oversized", "negative length", "duplicate self-loop"} {
		t.Run(kind, func(t *testing.T) {
			first := pageRow("links/a")
			first.target.path = present("beads/plan")
			second := pageRow("links/b")
			want := errCorrupt
			switch kind {
			case "malformed":
				second.properties = []byte("{")
			case "oversized":
				second.properties = []byte(strings.Repeat("x", fixtureLimits.valueBytes+1))
				want = errBudget
			case "negative length":
				want = errBudget
			case "duplicate self-loop":
				second = first
			}
			rows := addIncidentPageRow(incidentPageRows(), "beads/plan", true, true, first)
			if kind == "negative length" {
				values := linkValues(second)
				values[5] = nil
				values[6] = int64(-1)
				rows.AddRow(append([]driver.Value{"beads/plan", "beads/plan", "bead", "live", true, "beads/plan", true, true}, values...)...)
			} else {
				addIncidentPageRow(rows, "beads/plan", true, true, second)
			}
			tx, m := mockTx(t)
			expectIncidentPage(m, graph.DirectionBoth, "", 1, rows)
			page, err := readIncidentPageInTx(preconditionContext(t), tx, fixtureScope, "beads/plan", graph.DirectionBoth, pageWindow{limit: 1}, fixtureLimits)
			if !errors.Is(err, want) {
				t.Fatalf("error=%v", err)
			}
			requireEmptyPage(t, page)
		})
	}
}
func TestIncidentPageEmptyResultCloseFailurePropagates(t *testing.T) {
	boom := errors.New("incident close")
	tx, m := mockTx(t)
	expectIncidentPage(m, graph.DirectionIn, "", 1, incidentPageRows().CloseError(boom))
	page, err := readIncidentPageInTx(preconditionContext(t), tx, fixtureScope, "beads/plan", graph.DirectionIn, pageWindow{limit: 1}, fixtureLimits)
	if !errors.Is(err, boom) || errors.Is(err, errAbsent) {
		t.Fatalf("error=%v", err)
	}
	requireEmptyPage(t, page)
}

// Reuse the independently qualified driver seam: cancellation occurs inside
// actual database/sql Rows.Close, with no race against its own context watcher.
func TestPageCancellationAtRowsCompletion(t *testing.T) {
	boom := errors.New("page driver close failed")
	for _, kind := range []string{"links empty", "links row", "incident missing", "incident empty", "incident row"} {
		for _, failClose := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/close-error=%t", kind, failClose), func(t *testing.T) {
				ctx, cancel := context.WithCancel(preconditionContext(t))
				defer cancel()
				columns := allLinkColumns
				var values []driver.Value
				if strings.HasPrefix(kind, "incident") {
					columns = incidentAllocationColumns()
				}
				switch kind {
				case "links row":
					values = linkValues(pageRow("links/a"))
				case "incident empty":
					values = incidentLiveValues("beads/plan", false, false, linkRow{})
				case "incident row":
					values = incidentLiveValues("beads/plan", true, true, pageRow("links/a"))
				}
				var closeErr error
				if failClose {
					closeErr = boom
				}
				calls := 0
				db := sql.OpenDB(completionCloseDriver{columns: columns, values: values, cancel: cancel, closeError: closeErr, calls: &calls})
				defer func() {
					if err := db.Close(); err != nil {
						t.Error(err)
					}
				}()
				var page linkRowsPage
				var err error
				if strings.HasPrefix(kind, "incident") {
					page, err = readIncidentPageInTx(ctx, independentRowsContext{db}, fixtureScope, "beads/plan", graph.DirectionBoth, pageWindow{limit: 1}, fixtureLimits)
				} else {
					page, err = readLinkPageInTx(ctx, independentRowsContext{db}, fixtureScope, linkPageSelection{}, pageWindow{limit: 1}, fixtureLimits)
				}
				if !errors.Is(err, context.Canceled) || errors.Is(err, errAbsent) || calls != 1 || (failClose && !errors.Is(err, boom)) {
					t.Fatalf("calls=%d error=%v", calls, err)
				}
				requireEmptyPage(t, page)
			})
		}
	}
}
func TestPageCancellationBeforeDispatch(t *testing.T) {
	for _, incident := range []bool{false, true} {
		t.Run(fmt.Sprint(incident), func(t *testing.T) {
			ctx, cancel := context.WithCancel(preconditionContext(t))
			cancel()
			calls := 0
			db := sql.OpenDB(completionCloseDriver{columns: allLinkColumns, values: linkValues(pageRow("links/a")), cancel: cancel, calls: &calls})
			defer func() {
				if err := db.Close(); err != nil {
					t.Error(err)
				}
			}()
			var page linkRowsPage
			var err error
			if incident {
				page, err = readIncidentPageInTx(ctx, independentRowsContext{db}, fixtureScope, "beads/plan", graph.DirectionBoth, pageWindow{limit: 1}, fixtureLimits)
			} else {
				page, err = readLinkPageInTx(ctx, independentRowsContext{db}, fixtureScope, linkPageSelection{}, pageWindow{limit: 1}, fixtureLimits)
			}
			if !errors.Is(err, context.Canceled) || calls != 0 {
				t.Fatalf("calls=%d error=%v", calls, err)
			}
			requireEmptyPage(t, page)
		})
	}
}

func TestIncidentPageExactLimitTerminal(t *testing.T) {
	for _, direction := range []graph.Direction{graph.DirectionIn, graph.DirectionOut, graph.DirectionBoth} {
		t.Run(fmt.Sprint(direction), func(t *testing.T) {
			r := pageRow("links/a")
			r.target.path = present("beads/plan")
			tx, m := mockTx(t)
			expectIncidentPage(m, direction, "", 1, addIncidentPageRow(incidentPageRows(), "beads/plan", true, true, r))
			page, err := readIncidentPageInTx(preconditionContext(t), tx, fixtureScope, "beads/plan", direction, pageWindow{limit: 1}, fixtureLimits)
			if err != nil || len(page.items) != 1 || page.hasMore || page.lastPath != "links/a" {
				t.Fatalf("page=%+v error=%v", page, err)
			}
		})
	}
}

// A page-only multirow extension preserves the existing deterministic Close
// seam. The third row exceeds limit+lookahead before it can be decoded.
type pageOverflowDriver struct {
	completionCloseDriver
	rows [][]driver.Value
}

func (d pageOverflowDriver) Open(string) (driver.Conn, error) {
	return pageOverflowConn{completionCloseConn{d.completionCloseDriver}, d.rows}, nil
}
func (d pageOverflowDriver) Connect(context.Context) (driver.Conn, error) { return d.Open("") }
func (d pageOverflowDriver) Driver() driver.Driver                        { return d }

type pageOverflowConn struct {
	completionCloseConn
	rows [][]driver.Value
}

func (c pageOverflowConn) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	*c.d.calls++
	return &pageOverflowRows{completionCloseRows: completionCloseRows{d: c.d}, rows: c.rows}, nil
}

type pageOverflowRows struct {
	completionCloseRows
	rows [][]driver.Value
	next int
}

func (r *pageOverflowRows) Next(dest []driver.Value) error {
	if r.next == len(r.rows) {
		return io.EOF
	}
	copy(dest, r.rows[r.next])
	r.next++
	return nil
}
func TestPageOverflowPreservesCloseAndCancellation(t *testing.T) {
	for _, incident := range []bool{false, true} {
		t.Run(fmt.Sprint(incident), func(t *testing.T) {
			ctx, cancel := context.WithCancel(preconditionContext(t))
			defer cancel()
			boom := errors.New("overflow driver close failed")
			columns := allLinkColumns
			var values [][]driver.Value
			for _, path := range []string{"links/a", "links/b", "links/c"} {
				row := linkValues(pageRow(path))
				if incident {
					row = append([]driver.Value{"beads/plan", "beads/plan", "bead", "live", true, "beads/plan", true, true}, row...)
				}
				values = append(values, row)
			}
			if incident {
				columns = incidentAllocationColumns()
			}
			calls := 0
			db := sql.OpenDB(pageOverflowDriver{completionCloseDriver{columns: columns, cancel: cancel, closeError: boom, calls: &calls}, values})
			defer func() {
				if err := db.Close(); err != nil {
					t.Error(err)
				}
			}()
			var page linkRowsPage
			var err error
			if incident {
				page, err = readIncidentPageInTx(ctx, independentRowsContext{db}, fixtureScope, "beads/plan", graph.DirectionBoth, pageWindow{limit: 1}, fixtureLimits)
			} else {
				page, err = readLinkPageInTx(ctx, independentRowsContext{db}, fixtureScope, linkPageSelection{}, pageWindow{limit: 1}, fixtureLimits)
			}
			if !errors.Is(err, errCorrupt) || !errors.Is(err, errRowOverflow) || !errors.Is(err, context.Canceled) || !errors.Is(err, boom) || errors.Is(err, errBudget) || calls != 1 {
				t.Fatalf("calls=%d error=%v", calls, err)
			}
			requireEmptyPage(t, page)
		})
	}
}

// Match the pinned MySQL server's COM_STMT_PREPARE parameter enumeration.
// Parsing succeeds for both shapes; SetOp traversal loses its own LIMIT operand.
func TestIncidentPagePreparedParameterEnumeration(t *testing.T) {
	enumerate := func(t *testing.T, query string) []string {
		t.Helper()
		statement, err := sqlparser.Parse(query)
		if err != nil {
			t.Fatal(err)
		}
		var names []string
		err = sqlparser.Walk(func(node sqlparser.SQLNode) (bool, error) {
			if value, ok := node.(*sqlparser.SQLVal); ok && strings.HasPrefix(string(value.Val), ":v") {
				names = append(names, string(value.Val))
			}
			return true, nil
		}, statement)
		if err != nil {
			t.Fatal(err)
		}
		slices.Sort(names)
		return names
	}
	for _, direction := range []graph.Direction{graph.DirectionIn, graph.DirectionOut, graph.DirectionBoth} {
		t.Run(fmt.Sprint(direction), func(t *testing.T) {
			query, args := incidentPageQuery("beads/plan", direction, pageWindow{afterPath: "links/a", limit: 2}, fixtureLimits.valueBytes)
			want := []string{":v1", ":v2", ":v3", ":v4", ":v5", ":v6", ":v7"}
			if direction == graph.DirectionBoth {
				want = append(want, ":v8", ":v9")
			}
			if got := enumerate(t, query); !slices.Equal(got, want) || len(args) != len(want) || args[len(args)-1] != 3 {
				t.Fatalf("parameters=%v want=%v args=%v", got, want, args)
			}
			missing := strings.Replace(query, "src.path = i.path", "src.path = i.path AND FALSE", 1)
			if missing == query || !slices.Equal(enumerate(t, missing), want) {
				t.Fatal("diagnostic must preserve independently counted parameters")
			}
			if strings.Count(query, " LIMIT ?") != 1 || strings.Count(query, "LEFT_OUTER_LOOKUP_JOIN(i,src)") != 1 || !strings.Contains(query, "i.path AS candidate_path, src.path") {
				t.Fatal("candidate cap or independent projection shape changed")
			}
			if direction == graph.DirectionBoth {
				original := strings.Replace(query, "SELECT path FROM (SELECT path FROM graph_links", "SELECT path FROM graph_links", 1)
				if original == query {
					t.Fatal("predecessor SELECT reconstruction did not apply")
				}
				beforeLimit := original
				original = strings.Replace(original, ") candidates ORDER BY path LIMIT ?", " ORDER BY path LIMIT ?", 1)
				if original == beforeLimit {
					t.Fatal("predecessor LIMIT reconstruction did not apply")
				}
				// Keep the actual failing predecessor visible: only v9, the global
				// UNION LIMIT, disappears. This does not endorse that query for dispatch.
				// A future parser pin may repair its walk: update this pin tripwire,
				// not the still-valid production wrapper.
				if got := enumerate(t, original); !slices.Equal(got, []string{":v1", ":v2", ":v3", ":v4", ":v5", ":v6", ":v7", ":v8"}) {
					t.Fatalf("pinned predecessor behavior changed: %v", got)
				}
			} else if strings.Contains(query, " candidates") || strings.Contains(query, " UNION ") {
				t.Fatalf("single direction unnecessarily changed: %s", query)
			}
		})
	}
}
