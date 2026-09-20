package graphops

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	graph "github.com/steveyegge/beads/graphops"
)

var resourceColumns = []string{"path", "type_url", "revision", "attribution_principal", "attribution_status", "properties", "properties_length"}
var allLinkColumns = append(append([]string{}, resourceColumns...), "source_kind", "source_path", "source_url", "source_pin", "target_kind", "target_path", "target_url", "target_pin")

func asDriver(v sql.NullString) driver.Value {
	if v.Valid {
		return v.String
	}
	return nil
}
func blobLength(raw []byte) driver.Value {
	if raw == nil {
		return nil
	}
	return int64(len(raw))
}
func resourceValues(r resourceRow) []driver.Value {
	return []driver.Value{r.path, r.typeURL, r.revision, asDriver(r.principal), asDriver(r.attribution), r.properties, blobLength(r.properties)}
}
func linkValues(r linkRow) []driver.Value {
	return append(resourceValues(r.resourceRow), r.source.kind, asDriver(r.source.path), asDriver(r.source.url), asDriver(r.source.pin), r.target.kind, asDriver(r.target.path), asDriver(r.target.url), asDriver(r.target.pin))
}
func mockTx(t *testing.T) (*sql.Tx, sqlmock.Sqlmock) {
	t.Helper()
	db, m, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	m.ExpectBegin()
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		m.ExpectRollback()
		_ = tx.Rollback()
		if err := m.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
		_ = db.Close()
	})
	return tx, m
}
func expectBead(m sqlmock.Sqlmock, r resourceRow) {
	m.ExpectQuery(regexp.QuoteMeta("SELECT "+beadColumns+" FROM graph_beads WHERE path = ? LIMIT 2")).WithArgs(fixtureLimits.valueBytes, r.path).WillReturnRows(sqlmock.NewRows(resourceColumns).AddRow(resourceValues(r)...)).RowsWillBeClosed()
}
func expectDescriptor(m sqlmock.Sqlmock, d graph.TypeDescriptor) {
	m.ExpectQuery(regexp.QuoteMeta("SELECT url, CASE WHEN LENGTH(descriptor) < 0 OR LENGTH(descriptor) > ? THEN NULL ELSE descriptor END, LENGTH(descriptor), fingerprint FROM graph_type_descriptors WHERE url = ? LIMIT 2")).WithArgs(fixtureLimits.valueBytes, d.ID()).WillReturnRows(sqlmock.NewRows([]string{"url", "descriptor", "descriptor_length", "fingerprint"}).AddRow(d.ID(), d.CanonicalJSON(), blobLength(d.CanonicalJSON()), d.Fingerprint())).RowsWillBeClosed()
}
func expectLinks(m sqlmock.Sqlmock, where string, args []driver.Value, rows *sqlmock.Rows, capOverride ...int) {
	params := append([]driver.Value{fixtureLimits.valueBytes}, args...)
	limit := fixtureLimits.rows
	if len(capOverride) > 0 {
		limit = capOverride[0]
	}
	params = append(params, limit+1)
	m.ExpectQuery(regexp.QuoteMeta("SELECT " + linkColumns + " FROM graph_links WHERE " + where + " ORDER BY path LIMIT ?")).WithArgs(params...).WillReturnRows(rows).RowsWillBeClosed()
}

func TestReadBeadCompleteOwnedGroups(t *testing.T) {
	tx, m := mockTx(t)
	r := validRow()
	expectBead(m, r)
	decl, err := graph.NewOwnedLinkDecl(relationType, "Explains", 2)
	if err != nil {
		t.Fatal(err)
	}
	expectDescriptor(m, memoryDescriptor(t, decl))
	expectLinks(m, "source_kind = 'in' AND source_path = ? AND type_url IN (?)", []driver.Value{r.path, relationType}, sqlmock.NewRows(allLinkColumns), 2)
	record, err := readBeadInTx(t.Context(), tx, fixtureScope, r.path, fixtureLimits)
	if err != nil {
		t.Fatal(err)
	}
	if len(record.OwnedLinks) != 1 || len(record.OwnedLinks[0].Links) != 0 || record.OwnedLinks[0].TypeURL != relationType {
		t.Fatal("missing explicit empty owned group")
	}
}
func TestReadOwnedWildcardAndDescriptorLimit(t *testing.T) {
	for _, over := range []bool{false, true} {
		t.Run(map[bool]string{false: "present group", true: "declared limit refusal"}[over], func(t *testing.T) {
			tx, m := mockTx(t)
			expectBead(m, validRow())
			wild, err := graph.NewWildcardOwnedLinkDecl(1)
			if err != nil {
				t.Fatal(err)
			}
			expectDescriptor(m, memoryDescriptor(t, wild))
			rows := sqlmock.NewRows(allLinkColumns).AddRow(linkValues(validLinkRow())...)
			if over {
				r := validLinkRow()
				r.path = "links/second"
				rows.AddRow(linkValues(r)...)
			}
			expectLinks(m, "source_kind = 'in' AND source_path = ?", []driver.Value{"beads/plan"}, rows, 1)
			record, err := readBeadInTx(t.Context(), tx, fixtureScope, "beads/plan", fixtureLimits)
			if over {
				if !errors.Is(err, errCorrupt) {
					t.Fatalf("error=%v", err)
				}
			} else if err != nil || len(record.OwnedLinks) != 1 || len(record.OwnedLinks[0].Links) != 1 {
				t.Fatalf("record=%v error=%v", record, err)
			}
		})
	}
}
func TestReadLinkUsesPersistedDescriptorAndExactEndpoint(t *testing.T) {
	tx, m := mockTx(t)
	r := validLinkRow()
	d := relationDescriptor(t)
	columns := append(append([]string{}, allLinkColumns...), "url", "descriptor", "descriptor_length", "fingerprint")
	values := append(linkValues(r), d.ID(), d.CanonicalJSON(), blobLength(d.CanonicalJSON()), d.Fingerprint())
	m.ExpectQuery(regexp.QuoteMeta(exactLinkQuery)).WithArgs(fixtureLimits.valueBytes, fixtureLimits.valueBytes, r.path).WillReturnRows(sqlmock.NewRows(columns).AddRow(values...)).RowsWillBeClosed()
	link, err := readLinkInTx(t.Context(), tx, fixtureScope, r.path, fixtureLimits)
	if err != nil {
		t.Fatal(err)
	}
	if link.Target().Path() != "beads/decision" || link.Revision().String() != r.revision {
		t.Fatal("link changed")
	}
}
func incidentRows(rows ...linkRow) *sqlmock.Rows {
	columns := append([]string{"bead_path", "link_present"}, allLinkColumns...)
	result := sqlmock.NewRows(columns)
	for _, r := range rows {
		result.AddRow(append([]driver.Value{"beads/plan", true}, linkValues(r)...)...)
	}
	return result
}
func expectIncident(m sqlmock.Sqlmock, direction graph.Direction, limits readLimits, rows *sqlmock.Rows) {
	query, count := incidentQuery(direction)
	args := []driver.Value{limits.valueBytes}
	for i := 0; i < count; i++ {
		args = append(args, "beads/plan")
	}
	args = append(args, "beads/plan", limits.rows+1)
	m.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(args...).WillReturnRows(rows).RowsWillBeClosed()
}
func TestReadIncidentUnionHasOneSelfLoopAndDomainOrdering(t *testing.T) {
	tx, m := mockTx(t)
	first := validLinkRow()
	first.path = "links/b"
	first.target = first.source
	second := validLinkRow()
	second.path = "links/a"
	query, _ := incidentQuery(graph.DirectionBoth)
	if !strings.Contains(query, " UNION SELECT path FROM graph_links WHERE target_kind") || strings.Contains(query, "source_path = ? OR") {
		t.Fatal("incident query lost indexed union shape")
	}
	expectIncident(m, graph.DirectionBoth, fixtureLimits, incidentRows(first, second))
	links, err := readIncidentLinksInTx(t.Context(), tx, fixtureScope, "beads/plan", graph.DirectionBoth, fixtureLimits)
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 2 || links[0].Path() != "links/a" || links[1].Path() != "links/b" {
		t.Fatalf("links=%v", links)
	}
}
func TestIncidentDirectionsEmptyAndPhysicalAbsence(t *testing.T) {
	for _, direction := range []graph.Direction{graph.DirectionIn, graph.DirectionOut} {
		t.Run(direction.String(), func(t *testing.T) {
			tx, m := mockTx(t)
			row := validLinkRow()
			if direction == graph.DirectionIn {
				row.target, row.source = row.source, row.target
			}
			expectIncident(m, direction, fixtureLimits, incidentRows(row))
			links, err := readIncidentLinksInTx(t.Context(), tx, fixtureScope, "beads/plan", direction, fixtureLimits)
			if err != nil || len(links) != 1 {
				t.Fatalf("links=%v err=%v", links, err)
			}
		})
	}
	t.Run("empty anchor", func(t *testing.T) {
		tx, m := mockTx(t)
		rows := incidentRows()
		empty := linkValues(linkRow{})
		empty[5] = []byte("{}")
		rows.AddRow(append([]driver.Value{"beads/plan", false}, empty...)...)
		expectIncident(m, graph.DirectionBoth, fixtureLimits, rows)
		links, err := readIncidentLinksInTx(t.Context(), tx, fixtureScope, "beads/plan", graph.DirectionBoth, fixtureLimits)
		if err != nil || len(links) != 0 {
			t.Fatalf("links=%v error=%v", links, err)
		}
	})
	t.Run("absent anchor", func(t *testing.T) {
		tx, m := mockTx(t)
		expectIncident(m, graph.DirectionBoth, fixtureLimits, incidentRows())
		_, err := readIncidentLinksInTx(t.Context(), tx, fixtureScope, "beads/plan", graph.DirectionBoth, fixtureLimits)
		if !errors.Is(err, errAbsent) {
			t.Fatal(err)
		}
	})
}
func TestIncidentRowAndAggregateByteBoundaries(t *testing.T) {
	for _, kind := range []string{"rows", "aggregate bytes"} {
		t.Run(kind, func(t *testing.T) {
			tx, m := mockTx(t)
			limits := fixtureLimits
			row := validLinkRow()
			other := row
			other.path = "links/second"
			if kind == "rows" {
				limits.rows = 1
			} else {
				limits.valueBytes = 40
				limits.bytes = 40
			}
			expectIncident(m, graph.DirectionOut, limits, incidentRows(row, other))
			links, err := readIncidentLinksInTx(t.Context(), tx, fixtureScope, "beads/plan", graph.DirectionOut, limits)
			if !errors.Is(err, errBudget) || links != nil {
				t.Fatalf("links=%v error=%v", links, err)
			}
		})
	}
}
func TestExplicitOwnedLinksCountTowardWildcardWholeSet(t *testing.T) {
	tx, m := mockTx(t)
	expectBead(m, validRow())
	decl, err := graph.NewOwnedLinkDecl(relationType, "Explicit", 1)
	if err != nil {
		t.Fatal(err)
	}
	wild, err := graph.NewWildcardOwnedLinkDecl(1)
	if err != nil {
		t.Fatal(err)
	}
	expectDescriptor(m, memoryDescriptor(t, decl, wild))
	row := validLinkRow()
	other := row
	other.path = "links/other"
	other.typeURL = "https://graph.example/types/other"
	expectLinks(m, "source_kind = 'in' AND source_path = ?", []driver.Value{"beads/plan"}, sqlmock.NewRows(allLinkColumns).AddRow(linkValues(row)...).AddRow(linkValues(other)...), 1)
	_, err = readBeadInTx(t.Context(), tx, fixtureScope, "beads/plan", fixtureLimits)
	if !errors.Is(err, errCorrupt) {
		t.Fatal(err)
	}
}

func TestPhysicalAbsenceDoesNotInventPublicGoneMeaning(t *testing.T) {
	tx, m := mockTx(t)
	m.ExpectQuery("FROM graph_beads WHERE").WillReturnRows(sqlmock.NewRows(resourceColumns)).RowsWillBeClosed()
	_, err := readBeadInTx(t.Context(), tx, fixtureScope, "beads/plan", fixtureLimits)
	if !errors.Is(err, errAbsent) || errors.Is(err, graph.ErrNotFound) {
		t.Fatalf("private absence=%v", err)
	}
}
func TestReadFailuresRetainNoPartialResult(t *testing.T) {
	boom := errors.New("engine failure")
	for _, kind := range []string{"query", "scan", "iteration", "close", "bytes", "duplicate"} {
		t.Run(kind, func(t *testing.T) {
			tx, m := mockTx(t)
			q := m.ExpectQuery("FROM graph_beads WHERE")
			switch kind {
			case "query":
				q.WillReturnError(boom)
			case "scan":
				q.WillReturnRows(sqlmock.NewRows([]string{"one"}).AddRow("bad")).RowsWillBeClosed()
			case "iteration":
				q.WillReturnRows(sqlmock.NewRows(resourceColumns).AddRow(resourceValues(validRow())...).RowError(0, boom)).RowsWillBeClosed()
			case "close":
				q.WillReturnRows(sqlmock.NewRows(resourceColumns).CloseError(boom)).RowsWillBeClosed()
			case "bytes":
				r := validRow()
				r.properties = []byte(strings.Repeat(" ", fixtureLimits.valueBytes+1))
				q.WillReturnRows(sqlmock.NewRows(resourceColumns).AddRow(resourceValues(r)...)).RowsWillBeClosed()
			case "duplicate":
				q.WillReturnRows(sqlmock.NewRows(resourceColumns).AddRow(resourceValues(validRow())...).AddRow(resourceValues(validRow())...)).RowsWillBeClosed()
			}
			result, err := readBeadInTx(t.Context(), tx, fixtureScope, "beads/plan", fixtureLimits)
			if err == nil || !result.Bead.IsZero() {
				t.Fatalf("result=%v error=%v", result, err)
			}
			if kind == "duplicate" && (!errors.Is(err, errCorrupt) || errors.Is(err, errBudget)) {
				t.Fatalf("duplicate classification: %v", err)
			}
			if kind == "bytes" && (!errors.Is(err, errBudget) || errors.Is(err, errCorrupt)) {
				t.Fatalf("byte classification: %v", err)
			}
			if (kind == "query" || kind == "iteration" || kind == "close") && !errors.Is(err, boom) {
				t.Fatalf("engine error lost: %v", err)
			}
		})
	}
}
func TestInvalidRequestsIssueNoSQL(t *testing.T) {
	tx, _ := mockTx(t)
	if _, err := readBeadInTx(t.Context(), tx, fixtureScope, "beads/../x", fixtureLimits); !errors.Is(err, graph.ErrValidation) {
		t.Fatal(err)
	}
	if _, err := readLinkInTx(t.Context(), tx, fixtureScope, "links/x", readLimits{}); !errors.Is(err, graph.ErrValidation) {
		t.Fatal(err)
	}
	if _, err := readIncidentLinksInTx(t.Context(), tx, fixtureScope, "beads/plan", graph.Direction(99), fixtureLimits); !errors.Is(err, graph.ErrValidation) {
		t.Fatal(err)
	}
}

func TestPresentLinkNullPropertiesNeverBecomesEmptyObject(t *testing.T) {
	tx, m := mockTx(t)
	r := validLinkRow()
	r.properties = nil
	d := relationDescriptor(t)
	columns := append(append([]string{}, allLinkColumns...), "url", "descriptor", "descriptor_length", "fingerprint")
	values := append(linkValues(r), d.ID(), d.CanonicalJSON(), blobLength(d.CanonicalJSON()), d.Fingerprint())
	m.ExpectQuery(regexp.QuoteMeta(exactLinkQuery)).WillReturnRows(sqlmock.NewRows(columns).AddRow(values...)).RowsWillBeClosed()
	if _, err := readLinkInTx(t.Context(), tx, fixtureScope, r.path, fixtureLimits); !errors.Is(err, errCorrupt) {
		t.Fatalf("NULL properties repaired: %v", err)
	}
}

func TestPersistedScopePreconditionRefusesDevelopmentIdentity(t *testing.T) {
	if _, err := budgetFor("http://127.0.0.1:8080/demo/", fixtureLimits); err != nil {
		t.Fatalf("ordinary loopback scope: %v", err)
	}
	tx, _ := mockTx(t)
	if _, err := readBeadInTx(t.Context(), tx, "http://127.0.0.1:8080/local-test/", "beads/plan", fixtureLimits); !errors.Is(err, graph.ErrValidation) {
		t.Fatal(err)
	}
}

func TestSingletonLinkAndDescriptorOverflowClassification(t *testing.T) {
	for _, target := range []string{"link", "descriptor"} {
		for _, mode := range []string{"duplicate", "bytes"} {
			t.Run(target+"/"+mode, func(t *testing.T) {
				tx, m := mockTx(t)
				d := relationDescriptor(t)
				raw := d.CanonicalJSON()
				if mode == "bytes" {
					raw = []byte(strings.Repeat(" ", fixtureLimits.valueBytes+1))
				}
				var err error
				if target == "link" {
					row := validLinkRow()
					columns := append(append([]string{}, allLinkColumns...), "url", "descriptor", "descriptor_length", "fingerprint")
					values := append(linkValues(row), d.ID(), raw, blobLength(raw), d.Fingerprint())
					rows := sqlmock.NewRows(columns).AddRow(values...)
					if mode == "duplicate" {
						rows.AddRow(values...)
					}
					m.ExpectQuery(regexp.QuoteMeta(exactLinkQuery)).WillReturnRows(rows).RowsWillBeClosed()
					_, err = readLinkInTx(t.Context(), tx, fixtureScope, row.path, fixtureLimits)
				} else {
					rows := sqlmock.NewRows([]string{"url", "descriptor", "descriptor_length", "fingerprint"}).AddRow(d.ID(), raw, blobLength(raw), d.Fingerprint())
					if mode == "duplicate" {
						rows.AddRow(d.ID(), raw, blobLength(raw), d.Fingerprint())
					}
					m.ExpectQuery("FROM graph_type_descriptors WHERE").WillReturnRows(rows).RowsWillBeClosed()
					budget, e := budgetFor(fixtureScope, fixtureLimits)
					if e != nil {
						t.Fatal(e)
					}
					_, err = descriptorInTx(t.Context(), tx, d.ID(), graph.KindLink, budget)
				}
				want, other := errCorrupt, errBudget
				if mode == "bytes" {
					want, other = errBudget, errCorrupt
				}
				if !errors.Is(err, want) || errors.Is(err, other) {
					t.Fatalf("classification: %v", err)
				}
			})
		}
	}
}
func TestOwnedBudgetExhaustionIsNotPersistedCorruption(t *testing.T) {
	for _, mode := range []string{"private rows", "bytes under descriptor cap"} {
		t.Run(mode, func(t *testing.T) {
			tx, m := mockTx(t)
			expectBead(m, validRow())
			decl, err := graph.NewOwnedLinkDecl(relationType, "Explains", 2)
			if err != nil {
				t.Fatal(err)
			}
			expectDescriptor(m, memoryDescriptor(t, decl))
			limits := fixtureLimits
			row := validLinkRow()
			rows := sqlmock.NewRows(allLinkColumns)
			cap := 2
			if mode == "private rows" {
				limits.rows = 1
				cap = 1
				rows.AddRow(linkValues(row)...)
				row.path = "links/second"
			} else {
				row.properties = []byte(strings.Repeat(" ", fixtureLimits.valueBytes+1))
			}
			rows.AddRow(linkValues(row)...)
			expectLinks(m, "source_kind = 'in' AND source_path = ? AND type_url IN (?)", []driver.Value{"beads/plan", relationType}, rows, cap)
			_, err = readBeadInTx(t.Context(), tx, fixtureScope, "beads/plan", limits)
			if !errors.Is(err, errBudget) || errors.Is(err, errCorrupt) {
				t.Fatalf("budget classification: %v", err)
			}
		})
	}
}

func TestBlobProjectionLengthGuards(t *testing.T) {
	for _, test := range []struct {
		name   string
		raw    []byte
		length sql.NullInt64
		want   error
	}{
		{"negative engine length", nil, sql.NullInt64{Int64: -1, Valid: true}, errBudget},
		{"oversized suppressed payload", nil, sql.NullInt64{Int64: int64(fixtureLimits.valueBytes) + 1, Valid: true}, errBudget},
		{"mismatched payload", []byte("{}"), sql.NullInt64{Int64: 3, Valid: true}, errCorrupt},
		{"missing length", []byte("{}"), sql.NullInt64{}, errCorrupt},
		{"exact raw bytes", []byte("{}"), sql.NullInt64{Int64: 2, Valid: true}, nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			b, err := budgetFor(fixtureScope, fixtureLimits)
			if err != nil {
				t.Fatal(err)
			}
			err = chargeBlob(b, test.raw, test.length)
			if !errors.Is(err, test.want) {
				t.Fatalf("classification: %v", err)
			}
		})
	}
}
func TestOwnedDescriptorMaximumEqualsPrivateCap(t *testing.T) {
	tx, m := mockTx(t)
	expectBead(m, validRow())
	decl, err := graph.NewOwnedLinkDecl(relationType, "Explains", 1)
	if err != nil {
		t.Fatal(err)
	}
	expectDescriptor(m, memoryDescriptor(t, decl))
	row := validLinkRow()
	other := row
	other.path = "links/extra"
	expectLinks(m, "source_kind = 'in' AND source_path = ? AND type_url IN (?)", []driver.Value{"beads/plan", relationType}, sqlmock.NewRows(allLinkColumns).AddRow(linkValues(row)...).AddRow(linkValues(other)...), 1)
	limits := fixtureLimits
	limits.rows = 1
	_, err = readBeadInTx(t.Context(), tx, fixtureScope, "beads/plan", limits)
	if !errors.Is(err, errCorrupt) || errors.Is(err, errBudget) {
		t.Fatalf("equal-bound classification: %v", err)
	}
}
