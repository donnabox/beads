package graphops

import (
	"context"
	"database/sql/driver"
	"errors"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	graph "github.com/steveyegge/beads/graphops"
)

func exactTestRow(t *testing.T, kind graph.ResourceKind, state string, allocated, present bool) ([]string, []driver.Value) {
	t.Helper()
	path := "beads/plan"
	columns := exactBeadColumns
	if kind == graph.KindLink {
		path = "links/plan-decision"
		columns = exactLinkColumns
	}
	values := exactMissingValues(path, kind, "")
	if allocated {
		values[1], values[2], values[3] = path, string(kind), state
	}
	values[4] = present
	if present {
		body := resourceValues(validRow())
		if kind == graph.KindLink {
			r := validLinkRow()
			r.path = path
			d := relationDescriptor(t)
			body = append(linkValues(r), d.ID(), d.CanonicalJSON(), blobLength(d.CanonicalJSON()), d.Fingerprint())
		}
		values = append(values[:5], body...)
	}
	return columns, values
}
func exactTestRead(t *testing.T, ctx context.Context, tx queryer, kind graph.ResourceKind) (bool, error) {
	t.Helper()
	if kind == graph.KindBead {
		r, err := readBeadInTx(ctx, tx, fixtureScope, "beads/plan", fixtureLimits)
		return reflect.DeepEqual(r, graph.BeadRecord{}), err
	}
	r, err := readLinkInTx(ctx, tx, fixtureScope, "links/plan-decision", fixtureLimits)
	return r.IsZero(), err
}
func TestExactAllocationClassification(t *testing.T) {
	for _, kind := range []graph.ResourceKind{graph.KindBead, graph.KindLink} {
		for _, state := range []string{"", graph.AllocationLive, graph.AllocationReserved, graph.AllocationPruned, graph.AllocationErased, "invalid"} {
			for _, present := range []bool{false, true} {
				name := string(kind) + "/" + state + "/absent"
				if present {
					name = string(kind) + "/" + state + "/present"
				}
				t.Run(name, func(t *testing.T) {
					tx, m := mockTx(t)
					columns, values := exactTestRow(t, kind, state, state != "", present)
					query := exactBeadQuery
					args := []driver.Value{fixtureLimits.valueBytes, "beads/plan", "beads/plan", "beads/plan"}
					if kind == graph.KindLink {
						query = exactLinkQuery
						args = []driver.Value{fixtureLimits.valueBytes, fixtureLimits.valueBytes, "links/plan-decision", "links/plan-decision", "links/plan-decision"}
					}
					m.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(args...).WillReturnRows(sqlmock.NewRows(columns).AddRow(values...)).RowsWillBeClosed()
					live := state == graph.AllocationLive && present
					if live && kind == graph.KindBead {
						expectDescriptor(m, memoryDescriptor(t))
					}
					zero, err := exactTestRead(t, t.Context(), tx, kind)
					if live {
						if err != nil || zero {
							t.Fatalf("live: zero=%v err=%v", zero, err)
						}
						return
					}
					if !zero {
						t.Fatal("error retained resource")
					}
					gone := !present && (state == "" || state == graph.AllocationReserved || state == graph.AllocationPruned || state == graph.AllocationErased)
					var absence *exactAbsence
					if gone {
						if !errors.As(err, &absence) || absence.state != state {
							t.Fatalf("absence state=%v err=%v", absence, err)
						}
						if errors.Is(err, errAbsent) || errors.Is(err, errCorrupt) || errors.Is(err, graph.ErrNotFound) {
							t.Fatal("absence collapsed into sentinel")
						}
						if strings.Contains(err.Error(), string(kind)+"s/") || (state != "" && strings.Contains(err.Error(), state)) {
							t.Fatal("absence text leaks facts")
						}
					} else if !errors.Is(err, errCorrupt) || errors.As(err, &absence) {
						t.Fatalf("corruption=%v", err)
					} else {
						for _, sensitive := range []string{"beads/", "links/", "live", "reserved", "pruned", "erased"} {
							if strings.Contains(err.Error(), sensitive) {
								t.Fatalf("corrupt error text discloses physical facts: %v", err)
							}
						}
					}
				})
			}
		}
	}
}

func TestExactAllocationMalformedAndIncompleteResults(t *testing.T) {
	for _, kind := range []graph.ResourceKind{graph.KindBead, graph.KindLink} {
		for _, fault := range []string{"echo", "path", "kind", "partial-kind", "partial-state", "marker", "resource-path", "missing-anchor", "duplicate-absence", "iteration", "close"} {
			t.Run(string(kind)+"/"+fault, func(t *testing.T) {
				tx, m := mockTx(t)
				columns, values := exactTestRow(t, kind, graph.AllocationReserved, true, false)
				boom := errors.New("driver fault")
				switch fault {
				case "echo":
					values[0] = "beads/elsewhere"
				case "path":
					values[1] = "beads/elsewhere"
				case "kind":
					values[2] = "unknown"
				case "partial-kind":
					values[2] = nil
				case "partial-state":
					values[3] = nil
				case "marker":
					values[4] = true
				case "resource-path":
					values[5] = "beads/elsewhere"
				}
				rows := sqlmock.NewRows(columns)
				if fault != "missing-anchor" && fault != "close" {
					rows.AddRow(values...)
				}
				if fault == "duplicate-absence" {
					rows.AddRow(values...)
				}
				if fault == "iteration" {
					rows.RowError(0, boom)
				}
				if fault == "close" {
					rows.CloseError(boom)
				}
				query := exactBeadQuery
				if kind == graph.KindLink {
					query = exactLinkQuery
				}
				m.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(rows).RowsWillBeClosed()
				zero, err := exactTestRead(t, t.Context(), tx, kind)
				var absence *exactAbsence
				if !zero || errors.As(err, &absence) {
					t.Fatalf("partial result: zero=%v err=%v", zero, err)
				}
				want := errCorrupt
				if fault == "iteration" || fault == "close" {
					want = boom
				}
				if !errors.Is(err, want) {
					t.Fatalf("error=%v want=%v", err, want)
				}
			})
		}
	}
}
func TestExactPathBudgetRefusesBeforeSQL(t *testing.T) {
	tx, _ := mockTx(t)
	suffix := strings.Repeat("x", fixtureLimits.valueBytes)
	if _, err := readBeadInTx(t.Context(), tx, fixtureScope, "beads/"+suffix, fixtureLimits); !errors.Is(err, errBudget) {
		t.Fatal(err)
	}
	if _, err := readLinkInTx(t.Context(), tx, fixtureScope, "links/"+suffix, fixtureLimits); !errors.Is(err, errBudget) {
		t.Fatal(err)
	}
}

func TestExactNativePresenceMarker(t *testing.T) {
	for _, marker := range []driver.Value{false, int8(0), int64(0), []byte("0")} {
		db, m := nativeObservationMock(t)
		columns, values := exactTestRow(t, graph.KindBead, graph.AllocationPruned, true, false)
		values[4] = marker
		m.ExpectQuery(regexp.QuoteMeta(exactBeadQuery)).WillReturnRows(m.NewRows(columns).AddRow(values...)).RowsWillBeClosed()
		_, err := readBeadInTx(t.Context(), db, fixtureScope, "beads/plan", fixtureLimits)
		var absence *exactAbsence
		if !errors.As(err, &absence) || absence.state != graph.AllocationPruned {
			t.Fatalf("native marker %T: %v", marker, err)
		}
	}
}

func TestAbsentExactReadsUseFiveStatementsWithPreconditions(t *testing.T) {
	for _, kind := range []graph.ResourceKind{graph.KindBead, graph.KindLink} {
		for _, state := range []string{"", graph.AllocationReserved, graph.AllocationPruned, graph.AllocationErased} {
			t.Run(string(kind)+"/"+state, func(t *testing.T) {
				tx, m := mockTx(t)
				for _, f := range observationFixtures() {
					expectObservation(m, f).WillReturnRows(sqlmock.NewRows(f.columns).AddRow(f.values...)).RowsWillBeClosed()
				}
				m.ExpectQuery(regexp.QuoteMeta(stateObservationQuery)).WillReturnRows(sqlmock.NewRows(stateColumns).AddRow(stateValues()...)).RowsWillBeClosed()
				q := &preconditionQueryHook{tx: tx}
				n := uint64(3)
				ctx := preconditionContext(t)
				if _, err := observePreconditionsInTx(ctx, q, &n); err != nil {
					t.Fatal(err)
				}
				columns, values := exactTestRow(t, kind, state, state != "", false)
				query := exactBeadQuery
				if kind == graph.KindLink {
					query = exactLinkQuery
				}
				m.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(sqlmock.NewRows(columns).AddRow(values...)).RowsWillBeClosed()
				zero, err := exactTestRead(t, ctx, q, kind)
				var absence *exactAbsence
				if !zero || !errors.As(err, &absence) || absence.state != state || q.calls != 5 {
					t.Fatalf("zero=%v calls=%d err=%v", zero, q.calls, err)
				}
			})
		}
	}
}

func TestExactQueriesKeepIndependentAllocationAndResourceAnchors(t *testing.T) {
	for _, tc := range []struct {
		query, table, alias string
		parameters          int
	}{{exactBeadQuery, "graph_beads", "b", 4}, {exactLinkQuery, "graph_links", "l", 5}} {
		for _, part := range []string{"req.requested_path, a.path, CAST(a.resource_kind AS CHAR), CAST(a.state AS CHAR)", tc.alias + ".path IS NOT NULL", "FROM (SELECT ? AS requested_path) req", "LEFT JOIN (SELECT path, resource_kind, state FROM graph_allocations WHERE path = ? LIMIT 2) a ON TRUE"} {
			if !strings.Contains(tc.query, part) {
				t.Fatalf("missing required query projection/join %q", part)
			}
		}
		suffix := "FROM graph_beads WHERE path = ? LIMIT 2) b ON TRUE LIMIT 2"
		if tc.alias == "l" {
			for _, part := range []string{"LEFT JOIN (SELECT /*+ LEFT_OUTER_LOOKUP_JOIN(src,d) */ src.path", "d.url AS descriptor_url, d.descriptor AS descriptor_payload, d.fingerprint AS descriptor_fingerprint", ", l.descriptor_url, CASE WHEN LENGTH(l.descriptor_payload) < 0 OR LENGTH(l.descriptor_payload) > ? THEN NULL ELSE l.descriptor_payload END, LENGTH(l.descriptor_payload), l.descriptor_fingerprint FROM "} {
				if !strings.Contains(tc.query, part) {
					t.Fatalf("missing bounded descriptor projection/join %q", part)
				}
			}
			if !strings.HasPrefix(tc.query, "SELECT req.requested_path") || strings.Count(tc.query, "LEFT_OUTER_LOOKUP_JOIN") != 1 {
				t.Fatal("descriptor lookup hint not confined to inner Link query")
			}
			suffix = "FROM graph_links src LEFT JOIN graph_type_descriptors d ON d.url = src.type_url WHERE src.path = ? LIMIT 2) l ON TRUE LIMIT 2"
		} else if strings.Contains(tc.query, "/*+") {
			t.Fatal("unexpected optimizer hint in Bead query")
		}
		if !strings.HasSuffix(tc.query, suffix) {
			t.Fatal("exact lookup anchor or descriptor join changed")
		}
		if strings.Count(tc.query, "?") != tc.parameters || strings.Count(tc.query, " WHERE ") != 2 || strings.Count(tc.query, "LIMIT 2") != 3 {
			t.Fatal("query cardinality or argument shape changed")
		}
	}
}

func TestExactAllocationMetadataBudget(t *testing.T) {
	for _, kind := range []graph.ResourceKind{graph.KindBead, graph.KindLink} {
		t.Run(string(kind), func(t *testing.T) {
			tx, m := mockTx(t)
			columns, values := exactTestRow(t, kind, graph.AllocationPruned, true, false)
			path := values[0].(string)
			limits := fixtureLimits
			limits.valueBytes, limits.bytes = len(path)+1, len(path)+1
			query := exactBeadQuery
			args := []driver.Value{limits.valueBytes, path, path, path}
			if kind == graph.KindLink {
				query = exactLinkQuery
				args = []driver.Value{limits.valueBytes, limits.valueBytes, path, path, path}
			}
			m.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(args...).WillReturnRows(sqlmock.NewRows(columns).AddRow(values...)).RowsWillBeClosed()
			q := &preconditionQueryHook{tx: tx}
			var err error
			var zero bool
			if kind == graph.KindBead {
				var r graph.BeadRecord
				r, err = readBeadInTx(t.Context(), q, fixtureScope, path, limits)
				zero = reflect.DeepEqual(r, graph.BeadRecord{})
			} else {
				var r graph.Link
				r, err = readLinkInTx(t.Context(), q, fixtureScope, path, limits)
				zero = r.IsZero()
			}
			var absence *exactAbsence
			if !zero || !errors.Is(err, errBudget) || errors.As(err, &absence) || q.calls != 1 {
				t.Fatalf("zero=%v calls=%d err=%v", zero, q.calls, err)
			}
		})
	}
}
