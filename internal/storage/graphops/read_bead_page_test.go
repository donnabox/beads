package graphops

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"math"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	graph "github.com/steveyegge/beads/graphops"
)

func beadPageRow(path, typ string) resourceRow {
	r := validRow()
	r.path = path
	r.typeURL = typ
	return r
}
func beadPageRows(rows ...resourceRow) *sqlmock.Rows {
	out := sqlmock.NewRows(resourceColumns)
	for _, r := range rows {
		out.AddRow(resourceValues(r)...)
	}
	return out
}
func beadPageDescriptor(t *testing.T, typ string, owns ...graph.OwnedLinkDecl) graph.TypeDescriptor {
	t.Helper()
	d, err := graph.NewTypeDescriptor(graph.TypeDescriptorSpec{ID: typ, Name: "Page", Describes: graph.KindBead, OwnsOutgoing: owns})
	if err != nil {
		t.Fatal(err)
	}
	return d
}
func beadPageDecl(t *testing.T, typ string, max int) graph.OwnedLinkDecl {
	t.Helper()
	var d graph.OwnedLinkDecl
	var err error
	if typ == "*" {
		d, err = graph.NewWildcardOwnedLinkDecl(max)
	} else {
		d, err = graph.NewOwnedLinkDecl(typ, "", max)
	}
	if err != nil {
		t.Fatal(err)
	}
	return d
}
func expectBeadPageRows(m sqlmock.Sqlmock, typ, after string, limit, valueBytes int, rows *sqlmock.Rows) {
	query := "SELECT " + beadColumns + " FROM graph_beads WHERE path > ?"
	args := []driver.Value{valueBytes, after}
	if typ != "" {
		query += " AND type_url = ?"
		args = append(args, typ)
	}
	m.ExpectQuery(regexp.QuoteMeta(query + " ORDER BY path LIMIT ?")).WithArgs(append(args, limit+1)...).WillReturnRows(rows).RowsWillBeClosed()
}
func beadPageDescriptorRows(ds ...graph.TypeDescriptor) *sqlmock.Rows {
	rows := sqlmock.NewRows([]string{"url", "descriptor", "length", "fingerprint"})
	for _, d := range ds {
		rows.AddRow(d.ID(), d.CanonicalJSON(), len(d.CanonicalJSON()), d.Fingerprint())
	}
	return rows
}
func expectBeadPageDescriptors(m sqlmock.Sqlmock, ids []string, valueBytes int, rows *sqlmock.Rows) {
	args := []driver.Value{valueBytes}
	for _, id := range ids {
		args = append(args, id)
	}
	query := "SELECT url, CASE WHEN LENGTH(descriptor) < 0 OR LENGTH(descriptor) > ? THEN NULL ELSE descriptor END, LENGTH(descriptor), fingerprint FROM graph_type_descriptors WHERE url IN (" + strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",") + ") LIMIT ?"
	m.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(append(args, len(ids)+1)...).WillReturnRows(rows).RowsWillBeClosed()
}
func expectBeadPageOwned(m sqlmock.Sqlmock, clauses string, args []driver.Value, cap, valueBytes int, rows *sqlmock.Rows) {
	query := "SELECT " + linkColumns + " FROM graph_links WHERE source_kind = 'in' AND (" + clauses + ") ORDER BY source_path, type_url, path LIMIT ?"
	params := append([]driver.Value{valueBytes}, args...)
	params = append(params, cap+1)
	m.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(params...).WillReturnRows(rows).RowsWillBeClosed()
}
func requireZeroBeadPage(t *testing.T, page beadRowsPage) {
	t.Helper()
	if len(page.items) != 0 || page.lastPath != "" || page.hasMore {
		t.Fatalf("partial page escaped: %+v", page)
	}
}

func TestBeadPageBoundaries(t *testing.T) {
	for _, count := range []int{0, 1, 2, 3, 4} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			tx, m := mockTx(t)
			rows := beadPageRows()
			for i := 0; i < count; i++ {
				rows.AddRow(resourceValues(beadPageRow(fmt.Sprintf("beads/%d", i), memoryType))...)
			}
			expectBeadPageRows(m, "", "", 2, fixtureLimits.valueBytes, rows)
			if count > 0 && count < 4 {
				expectBeadPageDescriptors(m, []string{memoryType}, fixtureLimits.valueBytes, beadPageDescriptorRows(beadPageDescriptor(t, memoryType)))
			}
			q := &preconditionQueryHook{tx: tx}
			page, err := readBeadPageInTx(preconditionContext(t), q, fixtureScope, "", pageWindow{limit: 2}, fixtureLimits)
			if count == 4 {
				requireZeroBeadPage(t, page)
				if !errors.Is(err, errCorrupt) {
					t.Fatal(err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(page.items) != min(count, 2) || page.hasMore != (count > 2) {
				t.Fatalf("%+v", page)
			}
			if count == 0 {
				requireZeroBeadPage(t, page)
				if q.calls != 1 {
					t.Fatal(q.calls)
				}
			} else if page.lastPath != fmt.Sprintf("beads/%d", min(count, 2)-1) || q.calls != 2 {
				t.Fatalf("page=%+v calls=%d", page, q.calls)
			}
		})
	}
}
func TestBeadPageExactTypeAndGap(t *testing.T) {
	tx, m := mockTx(t)
	typ := memoryType + "/A"
	expectBeadPageRows(m, typ, "beads/absent", 2, fixtureLimits.valueBytes, beadPageRows(beadPageRow("beads/caf%C3%A9", typ), beadPageRow("beads/nested/x", typ)))
	expectBeadPageDescriptors(m, []string{typ}, fixtureLimits.valueBytes, beadPageDescriptorRows(beadPageDescriptor(t, typ)))
	page, err := readBeadPageInTx(preconditionContext(t), tx, fixtureScope, typ, pageWindow{afterPath: "beads/absent", limit: 2}, fixtureLimits)
	if err != nil || len(page.items) != 2 || page.hasMore || page.lastPath != "beads/nested/x" {
		t.Fatalf("%+v %v", page, err)
	}
}
func TestBeadPageInvalidOperandsNoSQL(t *testing.T) {
	for _, tc := range []struct {
		name, typ, after string
		limit            int
		ctx              context.Context
		want             error
	}{
		{"zero", "", "", 0, nil, graph.ErrValidation}, {"no-lookahead", "", "", fixtureLimits.rows, nil, graph.ErrValidation},
		{"overflow", "", "", math.MaxInt, nil, graph.ErrValidation}, {"link-path", "", "links/x", 1, nil, graph.ErrValidation},
		{"rawBMP", "", "beads/\uffff", 1, nil, graph.ErrValidation}, {"rawSupplementary", "", "beads/\U00010000", 1, nil, graph.ErrValidation},
		{"type-case", "https://EXAMPLE.test/type", "", 1, nil, graph.ErrValidation}, {"fragment", memoryType + "#x", "", 1, nil, graph.ErrValidation},
		{"long-path", "", "beads/" + strings.Repeat("x", fixtureLimits.valueBytes), 1, nil, errBudget},
		{"undeadlined", "", "", 1, context.Background(), errObservationDeadline},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tx, _ := mockTx(t)
			ctx := tc.ctx
			if ctx == nil {
				ctx = preconditionContext(t)
			}
			q := &preconditionQueryHook{tx: tx}
			page, err := readBeadPageInTx(ctx, q, fixtureScope, tc.typ, pageWindow{afterPath: tc.after, limit: tc.limit}, fixtureLimits)
			requireZeroBeadPage(t, page)
			if !errors.Is(err, tc.want) || q.calls != 0 {
				t.Fatalf("calls=%d err=%v", q.calls, err)
			}
		})
	}
}
func TestBeadPageLookaheadRefusals(t *testing.T) {
	for _, kind := range []string{"properties", "path", "supplementary", "type", "duplicate", "backwards", "length", "bytes"} {
		t.Run(kind, func(t *testing.T) {
			tx, m := mockTx(t)
			first := beadPageRow("beads/A", memoryType)
			next := beadPageRow("beads/a", memoryType)
			want := errCorrupt
			switch kind {
			case "properties":
				next.properties = []byte("null")
			case "path":
				next.path = "beads/\uffff"
			case "supplementary":
				next.path = "beads/\U00010000"
			case "type":
				next.typeURL = memoryType + "/other"
			case "duplicate":
				next.path = first.path
			case "backwards":
				next.path = "beads/0"
			}
			rows := beadPageRows(first)
			values := resourceValues(next)
			if kind == "length" {
				values[5] = nil
				values[6] = int64(-1)
				want = errBudget
			}
			if kind == "bytes" {
				values[5] = nil
				values[6] = int64(fixtureLimits.valueBytes + 1)
				want = errBudget
			}
			rows.AddRow(values...)
			expectBeadPageRows(m, memoryType, "", 1, fixtureLimits.valueBytes, rows)
			page, err := readBeadPageInTx(preconditionContext(t), tx, fixtureScope, memoryType, pageWindow{limit: 1}, fixtureLimits)
			requireZeroBeadPage(t, page)
			if !errors.Is(err, want) {
				t.Fatal(err)
			}
		})
	}
}
func TestBeadPageDescriptorBatchIncludesLookahead(t *testing.T) {
	a, z := memoryType+"/A", memoryType+"/z"
	for _, kind := range []string{"valid", "missing-lookahead", "duplicate", "unexpected", "wrong-kind", "noncanonical", "fingerprint", "null", "oversized"} {
		t.Run(kind, func(t *testing.T) {
			tx, m := mockTx(t)
			expectBeadPageRows(m, "", "", 1, fixtureLimits.valueBytes, beadPageRows(beadPageRow("beads/A", a), beadPageRow("beads/Z", z)))
			da, dz := beadPageDescriptor(t, a), beadPageDescriptor(t, z)
			rows := beadPageDescriptorRows(da)
			want := errCorrupt
			switch kind {
			case "valid":
				rows = beadPageDescriptorRows(da, dz)
			case "duplicate":
				rows = beadPageDescriptorRows(da, da)
			case "unexpected":
				rows = beadPageDescriptorRows(da, beadPageDescriptor(t, memoryType))
			case "wrong-kind":
				endpoint, err := graph.NewEndpointConstraint(nil, graph.ExternalOpaque)
				if err != nil {
					t.Fatal(err)
				}
				d, err := graph.NewTypeDescriptor(graph.TypeDescriptorSpec{ID: z, Name: "Wrong kind", Describes: graph.KindLink, Source: &endpoint, Target: &endpoint})
				if err != nil {
					t.Fatal(err)
				}
				rows.AddRow(z, d.CanonicalJSON(), len(d.CanonicalJSON()), d.Fingerprint())
			case "noncanonical":
				raw := append([]byte(" "), dz.CanonicalJSON()...)
				rows.AddRow(z, raw, len(raw), dz.Fingerprint())
			case "fingerprint":
				rows.AddRow(z, dz.CanonicalJSON(), len(dz.CanonicalJSON()), strings.Repeat("0", 64))
			case "null":
				rows.AddRow(z, nil, nil, dz.Fingerprint())
			case "oversized":
				rows.AddRow(z, nil, fixtureLimits.valueBytes+1, dz.Fingerprint())
				want = errBudget
			}
			expectBeadPageDescriptors(m, []string{a, z}, fixtureLimits.valueBytes, rows)
			page, err := readBeadPageInTx(preconditionContext(t), tx, fixtureScope, "", pageWindow{limit: 1}, fixtureLimits)
			if kind == "valid" {
				if err != nil || len(page.items) != 1 || !page.hasMore || page.lastPath != "beads/A" {
					t.Fatalf("%+v %v", page, err)
				}
			} else {
				requireZeroBeadPage(t, page)
				if !errors.Is(err, want) {
					t.Fatal(err)
				}
			}
		})
	}
}
func pageOwnedLink(source, typ, path string) linkRow {
	r := validLinkRow()
	r.source.path.String = source
	r.typeURL = typ
	r.path = path
	return r
}
func TestBeadPageBatchesOwnersAndPreservesGroups(t *testing.T) {
	tx, m := mockTx(t)
	a, z := memoryType+"/A", memoryType+"/z"
	other := relationType + "/other"
	explicit := beadPageDecl(t, relationType, 2)
	wild := beadPageDecl(t, "*", 3)
	expectBeadPageRows(m, "", "", 2, fixtureLimits.valueBytes, beadPageRows(beadPageRow("beads/A", a), beadPageRow("beads/B", z), beadPageRow("beads/Z", z)))
	expectBeadPageDescriptors(m, []string{a, z}, fixtureLimits.valueBytes, beadPageDescriptorRows(beadPageDescriptor(t, z, explicit, wild), beadPageDescriptor(t, a, explicit)))
	rows := sqlmock.NewRows(allLinkColumns).AddRow(linkValues(pageOwnedLink("beads/B", other, "links/B"))...)
	expectBeadPageOwned(m, "(source_path = ? AND type_url IN (?)) OR (source_path = ?)", []driver.Value{"beads/A", relationType, "beads/B"}, 5, fixtureLimits.valueBytes, rows)
	q := &preconditionQueryHook{tx: tx}
	page, err := readBeadPageInTx(preconditionContext(t), q, fixtureScope, "", pageWindow{limit: 2}, fixtureLimits)
	if err != nil {
		t.Fatal(err)
	}
	if q.calls != 3 || len(page.items) != 2 || !page.hasMore || page.lastPath != "beads/B" {
		t.Fatalf("%+v calls=%d", page, q.calls)
	}
	if len(page.items[0].OwnedLinks) != 1 || len(page.items[0].OwnedLinks[0].Links) != 0 || len(page.items[1].OwnedLinks) != 2 || len(page.items[1].OwnedLinks[1].Links) != 1 {
		t.Fatal(page)
	}
}
func TestBeadPageOwnedOverflowWitness(t *testing.T) {
	for _, corruptWitness := range []bool{false, true} {
		t.Run(fmt.Sprint(corruptWitness), func(t *testing.T) {
			tx, m := mockTx(t)
			limits := fixtureLimits
			limits.rows = 16 // 3 Beads + 3 descriptors leaves R=10.
			types := []string{memoryType + "/A", memoryType + "/B", memoryType + "/C"}
			rs := beadPageRows()
			ds := beadPageDescriptorRows()
			for i, max := range []int{4, 6, 100} {
				rs.AddRow(resourceValues(beadPageRow(fmt.Sprintf("beads/%c", 'A'+i), types[i]))...)
				d := beadPageDescriptor(t, types[i], beadPageDecl(t, relationType, max))
				ds.AddRow(d.ID(), d.CanonicalJSON(), len(d.CanonicalJSON()), d.Fingerprint())
			}
			expectBeadPageRows(m, "", "", 3, limits.valueBytes, rs)
			expectBeadPageDescriptors(m, types, limits.valueBytes, ds)
			links := sqlmock.NewRows(allLinkColumns)
			for i := 0; i < 11; i++ {
				source := "beads/B"
				if i < 4 {
					source = "beads/A"
				}
				if i == 10 && !corruptWitness {
					source = "beads/C"
				}
				links.AddRow(linkValues(pageOwnedLink(source, relationType, fmt.Sprintf("links/%02d", i)))...)
			}
			expectBeadPageOwned(m, "(source_path = ? AND type_url IN (?)) OR (source_path = ? AND type_url IN (?)) OR (source_path = ? AND type_url IN (?))", []driver.Value{"beads/A", relationType, "beads/B", relationType, "beads/C", relationType}, 10, limits.valueBytes, links)
			page, err := readBeadPageInTx(preconditionContext(t), tx, fixtureScope, "", pageWindow{limit: 3}, limits)
			requireZeroBeadPage(t, page)
			want := errBudget
			if corruptWitness {
				want = errCorrupt
			}
			if !errors.Is(err, want) || errors.Is(err, graph.ErrValidation) {
				t.Fatalf("want=%v err=%v", want, err)
			}
		})
	}
}
func TestBeadPageOwnedMalformedRows(t *testing.T) {
	for _, kind := range []string{"foreign", "unowned", "duplicate", "unordered", "wild-total", "explicit-max", "null"} {
		t.Run(kind, func(t *testing.T) {
			tx, m := mockTx(t)
			expectBeadPageRows(m, "", "", 1, fixtureLimits.valueBytes, beadPageRows(beadPageRow("beads/A", memoryType)))
			owns := []graph.OwnedLinkDecl{beadPageDecl(t, relationType, 1)}
			cap := 1
			clause := "(source_path = ? AND type_url IN (?))"
			args := []driver.Value{"beads/A", relationType}
			if kind == "wild-total" || kind == "explicit-max" || kind == "unordered" || kind == "duplicate" {
				owns = append(owns, beadPageDecl(t, "*", 3))
				cap = 3
				clause = "(source_path = ?)"
				args = args[:1]
			}
			expectBeadPageDescriptors(m, []string{memoryType}, fixtureLimits.valueBytes, beadPageDescriptorRows(beadPageDescriptor(t, memoryType, owns...)))
			r := pageOwnedLink("beads/A", relationType, "links/A")
			rows := sqlmock.NewRows(allLinkColumns)
			switch kind {
			case "foreign":
				r.source.path.String = "beads/elsewhere"
			case "unowned":
				r.typeURL = relationType + "/other"
			case "null":
				r.properties = nil
			}
			if kind == "wild-total" {
				for i := 0; i < 4; i++ {
					r.typeURL = fmt.Sprintf("%s/%d", relationType, i)
					r.path = fmt.Sprintf("links/%d", i)
					rows.AddRow(linkValues(r)...)
				}
			} else {
				rows.AddRow(linkValues(r)...)
			}
			if kind == "duplicate" {
				rows = sqlmock.NewRows(allLinkColumns).AddRow(linkValues(pageOwnedLink("beads/A", relationType+"/a", "links/0"))...).AddRow(linkValues(pageOwnedLink("beads/A", relationType+"/b", "links/0"))...)
			}
			if kind == "unordered" {
				r.typeURL = relationType + "/other"
				r.path = "links/0"
				rows.AddRow(linkValues(r)...)
				rows.AddRow(linkValues(pageOwnedLink("beads/A", relationType+"/aaa", "links/1"))...)
			}
			if kind == "explicit-max" {
				r.path = "links/B"
				rows.AddRow(linkValues(r)...)
			}
			expectBeadPageOwned(m, clause, args, cap, fixtureLimits.valueBytes, rows)
			page, err := readBeadPageInTx(preconditionContext(t), tx, fixtureScope, "", pageWindow{limit: 1}, fixtureLimits)
			requireZeroBeadPage(t, page)
			if !errors.Is(err, errCorrupt) {
				t.Fatal(err)
			}
		})
	}
}
func TestBeadPageAggregateRowAndMetadataBudgets(t *testing.T) {
	for _, kind := range []string{"descriptor-rows", "declarations", "aggregate-bytes", "small-terminal"} {
		t.Run(kind, func(t *testing.T) {
			tx, m := mockTx(t)
			limits := fixtureLimits
			limits.rows = 4
			window := pageWindow{limit: 3}
			rows := beadPageRows(beadPageRow("beads/A", memoryType))
			if kind == "descriptor-rows" {
				for _, path := range []string{"beads/B", "beads/C", "beads/D"} {
					rows.AddRow(resourceValues(beadPageRow(path, memoryType))...)
				}
			}
			if kind == "aggregate-bytes" {
				limits.valueBytes = 512
				limits.bytes = 512
			}
			expectBeadPageRows(m, "", "", 3, limits.valueBytes, rows)
			if kind != "descriptor-rows" {
				var owns []graph.OwnedLinkDecl
				if kind == "declarations" {
					for i := 0; i < 5; i++ {
						owns = append(owns, beadPageDecl(t, fmt.Sprintf("%s/%d", relationType, i), 1))
					}
				}
				d := beadPageDescriptor(t, memoryType, owns...)
				if kind == "aggregate-bytes" {
					var err error
					d, err = graph.NewTypeDescriptor(graph.TypeDescriptorSpec{ID: memoryType, Name: strings.Repeat("n", 350), Describes: graph.KindBead})
					if err != nil {
						t.Fatal(err)
					}
					if len(d.CanonicalJSON()) > limits.valueBytes || len(d.CanonicalJSON())+len(d.ID())+len(d.Fingerprint()) <= limits.bytes {
						t.Fatal("aggregate-byte control does not cross the intended boundary")
					}
				}
				expectBeadPageDescriptors(m, []string{memoryType}, limits.valueBytes, beadPageDescriptorRows(d))
			}
			page, err := readBeadPageInTx(preconditionContext(t), tx, fixtureScope, "", window, limits)
			if kind == "small-terminal" {
				if err != nil || len(page.items) != 1 {
					t.Fatalf("%+v %v", page, err)
				}
			} else {
				requireZeroBeadPage(t, page)
				if !errors.Is(err, errBudget) {
					t.Fatal(err)
				}
			}
		})
	}
}
func TestBeadPageSaturatingCapAndSQLShape(t *testing.T) {
	big := int(9007199254740994)
	owner := &beadPageOwner{owns: []graph.OwnedLinkDecl{beadPageDecl(t, relationType, big), beadPageDecl(t, relationType+"/b", big)}}
	if beadOwnedCap([]*beadPageOwner{owner, owner}, 1025) != 1025 {
		t.Fatal("cap did not saturate")
	}
	q, args := beadPageQuery(memoryType, pageWindow{afterPath: "beads/A", limit: 3}, 100)
	if !strings.Contains(q, "WHERE path > ? AND type_url = ? ORDER BY path LIMIT ?") || !slices.Equal(args, []any{100, "beads/A", memoryType, 4}) {
		t.Fatalf("%s %v", q, args)
	}
	var err error
	owner.record.Bead, err = decodeBead(beadPageRow("beads/A", memoryType))
	if err != nil {
		t.Fatal(err)
	}
	q, args = beadOwnedBatchQuery([]*beadPageOwner{owner}, 100, 10)
	if !strings.Contains(q, "source_kind = 'in' AND ((source_path = ? AND type_url IN (?,?))) ORDER BY source_path, type_url, path LIMIT ?") || !slices.Equal(args, []any{100, "beads/A", relationType, relationType + "/b", 11}) {
		t.Fatalf("%s %v", q, args)
	}
}

type beadPageStageQuery struct {
	dbs   []*sql.DB
	calls int
}

func (q *beadPageStageQuery) QueryContext(_ context.Context, query string, args ...any) (*sql.Rows, error) {
	if q.calls >= len(q.dbs) {
		return nil, errors.New("unexpected query")
	}
	db := q.dbs[q.calls]
	q.calls++
	return db.QueryContext(context.Background(), query, args...)
}
func TestBeadPageStageCloseCancellationDiscardsAll(t *testing.T) {
	for stage := 0; stage < 3; stage++ {
		t.Run(fmt.Sprint(stage), func(t *testing.T) {
			ctx, cancel := context.WithCancel(preconditionContext(t))
			defer cancel()
			boom := errors.New("close failed")
			d := beadPageDescriptor(t, memoryType, beadPageDecl(t, relationType, 2))
			calls := 0
			fixtures := []bodyCloseDriver{{columns: resourceColumns, values: resourceValues(beadPageRow("beads/A", memoryType))}, {columns: []string{"url", "descriptor", "length", "fingerprint"}, values: []driver.Value{d.ID(), d.CanonicalJSON(), int64(len(d.CanonicalJSON())), d.Fingerprint()}}, {columns: allLinkColumns, values: linkValues(pageOwnedLink("beads/A", relationType, "links/A"))}}
			q := &beadPageStageQuery{}
			for i, f := range fixtures {
				f.calls = &calls
				f.cancel = func() {}
				if i == stage {
					f.cancel = cancel
					f.closeError = boom
				}
				db := sql.OpenDB(f)
				q.dbs = append(q.dbs, db)
				t.Cleanup(func() {
					if err := db.Close(); err != nil {
						t.Error(err)
					}
				})
			}
			page, err := readBeadPageInTx(ctx, q, fixtureScope, "", pageWindow{limit: 1}, fixtureLimits)
			requireZeroBeadPage(t, page)
			if !errors.Is(err, context.Canceled) || !errors.Is(err, boom) || q.calls != stage+1 {
				t.Fatalf("calls=%d err=%v", q.calls, err)
			}
		})
	}
}

// Raw observations plus body SQL, not a claimed authority grant. Varying page
// size/types must not turn either descriptor or owned expansion into N+1.
func TestBeadPageSevenStatementComposition(t *testing.T) {
	for _, count := range []int{1, 3} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			tx, m := mockTx(t)
			for _, f := range observationFixtures() {
				expectObservation(m, f).WillReturnRows(sqlmock.NewRows(f.columns).AddRow(f.values...)).RowsWillBeClosed()
			}
			m.ExpectQuery(regexp.QuoteMeta(stateObservationQuery)).WillReturnRows(sqlmock.NewRows(stateColumns).AddRow(stateValues()...)).RowsWillBeClosed()
			q := &preconditionQueryHook{tx: tx}
			ctx := preconditionContext(t)
			seq := uint64(3)
			if _, err := observePreconditionsInTx(ctx, q, &seq); err != nil {
				t.Fatal(err)
			}
			rows := beadPageRows()
			ds := beadPageDescriptorRows()
			var ids, clauses []string
			var args []driver.Value
			for i := 0; i < count; i++ {
				path := fmt.Sprintf("beads/%d", i)
				typ := fmt.Sprintf("%s/%d", memoryType, i)
				ids = append(ids, typ)
				rows.AddRow(resourceValues(beadPageRow(path, typ))...)
				d := beadPageDescriptor(t, typ, beadPageDecl(t, relationType, 1))
				ds.AddRow(d.ID(), d.CanonicalJSON(), len(d.CanonicalJSON()), d.Fingerprint())
				clauses = append(clauses, "(source_path = ? AND type_url IN (?))")
				args = append(args, path, relationType)
			}
			expectBeadPageRows(m, "", "", count, fixtureLimits.valueBytes, rows)
			expectBeadPageDescriptors(m, ids, fixtureLimits.valueBytes, ds)
			expectBeadPageOwned(m, strings.Join(clauses, " OR "), args, count, fixtureLimits.valueBytes, sqlmock.NewRows(allLinkColumns))
			page, err := readBeadPageInTx(ctx, q, fixtureScope, "", pageWindow{limit: count}, fixtureLimits)
			if err != nil || q.calls != 7 || len(page.items) != count {
				t.Fatalf("calls=%d page=%+v err=%v", q.calls, page, err)
			}
		})
	}
}

func TestBeadPageSharedDescriptorStillChargesEachRecordsGroups(t *testing.T) {
	tx, m := mockTx(t)
	limits := fixtureLimits
	limits.rows = 4
	expectBeadPageRows(m, "", "", 2, limits.valueBytes, beadPageRows(beadPageRow("beads/A", memoryType), beadPageRow("beads/B", memoryType)))
	owns := []graph.OwnedLinkDecl{beadPageDecl(t, relationType, 1), beadPageDecl(t, relationType+"/b", 1), beadPageDecl(t, relationType+"/c", 1)}
	expectBeadPageDescriptors(m, []string{memoryType}, limits.valueBytes, beadPageDescriptorRows(beadPageDescriptor(t, memoryType, owns...)))
	q := &preconditionQueryHook{tx: tx}
	page, err := readBeadPageInTx(preconditionContext(t), q, fixtureScope, "", pageWindow{limit: 2}, limits)
	requireZeroBeadPage(t, page)
	if !errors.Is(err, errBudget) || q.calls != 2 {
		t.Fatalf("calls=%d err=%v", q.calls, err)
	}
}
func TestBeadPageAggregateBytesAcrossAllThreeStages(t *testing.T) {
	tx, m := mockTx(t)
	limits := fixtureLimits
	limits.valueBytes = 512
	bead := beadPageRow("beads/A", memoryType)
	link := pageOwnedLink("beads/A", relationType, "links/A")
	link.properties = []byte(`{"text":"` + strings.Repeat("x", 300) + `"}`)
	d := beadPageDescriptor(t, memoryType, beadPageDecl(t, relationType, 2))
	// Independent transferred scalar accounting, including one explicit group
	// key. Every individual value fits 512; the final Link crosses the shared sum.
	total := len(bead.path) + len(bead.typeURL) + len(bead.revision) + len(bead.properties) + len(d.ID()) + len(d.CanonicalJSON()) + len(d.Fingerprint()) + len(relationType)
	total += len(link.path) + len(link.typeURL) + len(link.revision) + len(link.properties) + len(link.source.kind) + len(link.source.path.String) + len(link.source.pin.String) + len(link.target.kind) + len(link.target.path.String) + len(link.target.pin.String)
	limits.bytes = total - 1
	if limits.bytes < limits.valueBytes {
		t.Fatal("test does not reach shared-byte boundary")
	}
	expectBeadPageRows(m, "", "", 1, limits.valueBytes, beadPageRows(bead))
	expectBeadPageDescriptors(m, []string{memoryType}, limits.valueBytes, beadPageDescriptorRows(d))
	expectBeadPageOwned(m, "(source_path = ? AND type_url IN (?))", []driver.Value{"beads/A", relationType}, 2, limits.valueBytes, sqlmock.NewRows(allLinkColumns).AddRow(linkValues(link)...))
	q := &preconditionQueryHook{tx: tx}
	page, err := readBeadPageInTx(preconditionContext(t), q, fixtureScope, "", pageWindow{limit: 1}, limits)
	requireZeroBeadPage(t, page)
	if !errors.Is(err, errBudget) || q.calls != 3 {
		t.Fatalf("calls=%d error=%v", q.calls, err)
	}
}

func TestBeadPageOwnedSQLLimitVersusBudgetWitness(t *testing.T) {
	for _, count := range []int{4, 5} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			tx, m := mockTx(t)
			limits := fixtureLimits
			limits.rows = 5
			expectBeadPageRows(m, "", "", 1, limits.valueBytes, beadPageRows(beadPageRow("beads/A", memoryType)))
			d := beadPageDescriptor(t, memoryType, beadPageDecl(t, "*", 100))
			expectBeadPageDescriptors(m, []string{memoryType}, limits.valueBytes, beadPageDescriptorRows(d))
			rows := sqlmock.NewRows(allLinkColumns)
			for i := 0; i < count; i++ {
				rows.AddRow(linkValues(pageOwnedLink("beads/A", relationType, fmt.Sprintf("links/%d", i)))...)
			}
			expectBeadPageOwned(m, "(source_path = ?)", []driver.Value{"beads/A"}, 3, limits.valueBytes, rows)
			page, err := readBeadPageInTx(preconditionContext(t), tx, fixtureScope, "", pageWindow{limit: 1}, limits)
			requireZeroBeadPage(t, page)
			want, other := errBudget, errCorrupt
			if count == 5 {
				want, other = other, want
			}
			if !errors.Is(err, want) || errors.Is(err, other) {
				t.Fatalf("count=%d error=%v", count, err)
			}
		})
	}
}
func TestBeadPageWildcardGroupsExhaustMetadataAfterQuery(t *testing.T) {
	tx, m := mockTx(t)
	limits := fixtureLimits
	limits.rows = 5
	expectBeadPageRows(m, "", "", 1, limits.valueBytes, beadPageRows(beadPageRow("beads/A", memoryType)))
	d := beadPageDescriptor(t, memoryType, beadPageDecl(t, "*", 3), beadPageDecl(t, relationType+"/x", 1), beadPageDecl(t, relationType+"/y", 1))
	expectBeadPageDescriptors(m, []string{memoryType}, limits.valueBytes, beadPageDescriptorRows(d))
	rows := sqlmock.NewRows(allLinkColumns)
	for i, suffix := range []string{"a", "b", "c"} {
		rows.AddRow(linkValues(pageOwnedLink("beads/A", relationType+"/"+suffix, fmt.Sprintf("links/%d", i)))...)
	}
	expectBeadPageOwned(m, "(source_path = ?)", []driver.Value{"beads/A"}, 3, limits.valueBytes, rows)
	q := &preconditionQueryHook{tx: tx}
	page, err := readBeadPageInTx(preconditionContext(t), q, fixtureScope, "", pageWindow{limit: 1}, limits)
	requireZeroBeadPage(t, page)
	if !errors.Is(err, errBudget) || errors.Is(err, errCorrupt) || q.calls != 3 {
		t.Fatalf("calls=%d error=%v", q.calls, err)
	}
}
func TestBeadPageLookaheadDeclarationsDoNotConsumeGroups(t *testing.T) {
	tx, m := mockTx(t)
	limits := fixtureLimits
	limits.rows = 6
	z := memoryType + "/z"
	expectBeadPageRows(m, "", "", 1, limits.valueBytes, beadPageRows(beadPageRow("beads/A", memoryType), beadPageRow("beads/Z", z)))
	var owns []graph.OwnedLinkDecl
	for i := 0; i < 7; i++ {
		owns = append(owns, beadPageDecl(t, fmt.Sprintf("%s/%d", relationType, i), 1))
	}
	expectBeadPageDescriptors(m, []string{memoryType, z}, limits.valueBytes, beadPageDescriptorRows(beadPageDescriptor(t, memoryType), beadPageDescriptor(t, z, owns...)))
	q := &preconditionQueryHook{tx: tx}
	page, err := readBeadPageInTx(preconditionContext(t), q, fixtureScope, "", pageWindow{limit: 1}, limits)
	if err != nil || len(page.items) != 1 || !page.hasMore || page.lastPath != "beads/A" || q.calls != 2 {
		t.Fatalf("page=%+v calls=%d error=%v", page, q.calls, err)
	}
}
func TestBeadPageWitnessByteRefusalPrecedesDescriptorOverflow(t *testing.T) {
	tx, m := mockTx(t)
	limits := fixtureLimits
	limits.valueBytes = 512
	bead := beadPageRow("beads/A", memoryType)
	first := pageOwnedLink("beads/A", relationType, "links/A")
	first.properties = []byte(`{"text":"` + strings.Repeat("x", 300) + `"}`)
	second := first
	second.path = "links/B"
	d := beadPageDescriptor(t, memoryType, beadPageDecl(t, relationType, 1))
	scalarBytes := func(values []driver.Value) int {
		n := 0
		for _, v := range values {
			switch v := v.(type) {
			case string:
				n += len(v)
			case []byte:
				n += len(v)
			}
		}
		return n
	}
	limits.bytes = scalarBytes(resourceValues(bead)) + len(d.ID()) + len(d.CanonicalJSON()) + len(d.Fingerprint()) + len(relationType) + scalarBytes(linkValues(first)) + 50
	if limits.bytes < limits.valueBytes || scalarBytes(linkValues(second)) <= 50 {
		t.Fatal("invalid byte witness stimulus")
	}
	expectBeadPageRows(m, "", "", 1, limits.valueBytes, beadPageRows(bead))
	expectBeadPageDescriptors(m, []string{memoryType}, limits.valueBytes, beadPageDescriptorRows(d))
	expectBeadPageOwned(m, "(source_path = ? AND type_url IN (?))", []driver.Value{"beads/A", relationType}, 1, limits.valueBytes, sqlmock.NewRows(allLinkColumns).AddRow(linkValues(first)...).AddRow(linkValues(second)...))
	page, err := readBeadPageInTx(preconditionContext(t), tx, fixtureScope, "", pageWindow{limit: 1}, limits)
	requireZeroBeadPage(t, page)
	if !errors.Is(err, errBudget) || errors.Is(err, errCorrupt) {
		t.Fatal(err)
	}
}
func TestBeadPageOwnerRequiresDecodedDescriptor(t *testing.T) {
	bead, err := decodeBead(validRow())
	if err != nil {
		t.Fatal(err)
	}
	budget, err := beadPageBudgetFor(preconditionContext(t), fixtureScope, "", pageWindow{limit: 1}, fixtureLimits)
	if err != nil {
		t.Fatal(err)
	}
	owners, err := makeBeadPageOwners([]graph.Bead{bead}, map[string]graph.TypeDescriptor{}, budget)
	if !errors.Is(err, errCorrupt) || len(owners) != 0 {
		t.Fatalf("owners=%v error=%v", owners, err)
	}
}
