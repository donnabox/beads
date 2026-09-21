package graphops

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage/authority"
)

func incidentAllocationColumns() []string {
	return append([]string{"requested", "allocation_path", "kind", "state", "bead_present", "bead_path", "candidate", "present"}, allLinkColumns...)
}
func incidentLiveValues(path string, candidate, present bool, row linkRow) []driver.Value {
	return append([]driver.Value{path, path, "bead", "live", true, path, candidate, present}, linkValues(row)...)
}
func incidentAbsenceValues(state string) []driver.Value {
	v := append([]driver.Value{"beads/plan", nil, nil, nil, false, "", false, false}, linkValues(linkRow{})...)
	if state != "" {
		v[1], v[2], v[3] = "beads/plan", "bead", state
	}
	return v
}
func TestIncidentAllocationStates(t *testing.T) {
	for _, direction := range []graph.Direction{graph.DirectionIn, graph.DirectionOut, graph.DirectionBoth} {
		for _, state := range []string{"", "live", "reserved", "pruned", "erased", "unknown"} {
			for _, present := range []bool{false, true} {
				t.Run(fmt.Sprintf("%d/%s/%t", direction, state, present), func(t *testing.T) {
					tx, m := mockTx(t)
					v := incidentAbsenceValues(state)
					if present {
						v[4], v[5] = true, "beads/plan"
					}
					expectIncidentPage(m, direction, "", 1, incidentPageRows().AddRow(v...))
					q := &preconditionQueryHook{tx: tx}
					page, err := readIncidentPageInTx(preconditionContext(t), q, fixtureScope, "beads/plan", direction, pageWindow{limit: 1}, fixtureLimits)
					requireEmptyPage(t, page)
					if q.calls != 1 {
						t.Fatal(q.calls)
					}
					if state == "live" && present {
						if err != nil {
							t.Fatal(err)
						}
						return
					}
					var absence *exactAbsence
					if !present && (state == "" || state == "reserved" || state == "pruned" || state == "erased") {
						if !errors.As(err, &absence) || absence.state != state || errors.Is(err, errAbsent) || errors.Is(err, errCorrupt) || errors.Is(err, graph.ErrNotFound) {
							t.Fatalf("state=%q err=%v", state, err)
						}
						if strings.Contains(err.Error(), "beads/") || (state != "" && strings.Contains(err.Error(), state)) {
							t.Fatal("disclosing absence")
						}
					} else if !errors.Is(err, errCorrupt) || errors.As(err, &absence) {
						t.Fatal(err)
					}
				})
			}
		}
	}
}
func TestIncidentAllocationMalformedProjection(t *testing.T) {
	for _, fault := range []string{"echo", "allocation-path", "kind", "partial-kind", "partial-state", "marker-null", "candidate-null", "hydrated-null", "bead-path", "candidate", "absent-hydrated", "payload", "duplicate", "changed", "empty"} {
		t.Run(fault, func(t *testing.T) {
			tx, m := mockTx(t)
			v := incidentAbsenceValues("reserved")
			rows := incidentPageRows()
			switch fault {
			case "echo":
				v[0] = "beads/other"
			case "allocation-path":
				v[1] = "beads/other"
			case "kind":
				v[2] = "link"
			case "partial-kind":
				v[2] = nil
			case "partial-state":
				v[3] = nil
			case "marker-null":
				v[4] = nil
			case "candidate-null":
				v[6] = nil
			case "hydrated-null":
				v[7] = nil
			case "absent-hydrated":
				v[6], v[7] = true, true
				copy(v[8:], linkValues(pageRow("links/a")))
			case "bead-path":
				v[5] = "beads/plan"
			case "candidate":
				v[6] = true
			case "payload":
				v[8] = "links/junk"
			}
			if fault != "empty" {
				rows.AddRow(v...)
			}
			if fault == "duplicate" {
				rows.AddRow(v...)
			}
			if fault == "changed" {
				rows.AddRow(incidentAbsenceValues("pruned")...)
			}
			expectIncidentPage(m, graph.DirectionBoth, "", 1, rows)
			page, err := readIncidentPageInTx(preconditionContext(t), tx, fixtureScope, "beads/plan", graph.DirectionBoth, pageWindow{limit: 1}, fixtureLimits)
			requireEmptyPage(t, page)
			var absence *exactAbsence
			if err == nil || errors.As(err, &absence) || (!strings.HasSuffix(fault, "-null") && !errors.Is(err, errCorrupt)) {
				t.Fatal(err)
			}
		})
	}
}
func TestIncidentAllocationCloseOverridesAbsence(t *testing.T) {
	for _, state := range []string{"", "reserved", "pruned", "erased"} {
		for _, cancelOnClose := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/%t", state, cancelOnClose), func(t *testing.T) {
				ctx, cancel := context.WithCancel(preconditionContext(t))
				defer cancel()
				boom := errors.New("incident close fault")
				calls := 0
				cancelHook := func() {}
				if cancelOnClose {
					cancelHook = cancel
				}
				db := sql.OpenDB(completionCloseDriver{columns: incidentAllocationColumns(), values: incidentAbsenceValues(state), cancel: cancelHook, closeError: boom, calls: &calls})
				defer db.Close()
				page, err := readIncidentPageInTx(ctx, independentRowsContext{db}, fixtureScope, "beads/plan", graph.DirectionBoth, pageWindow{limit: 1}, fixtureLimits)
				requireEmptyPage(t, page)
				var absence *exactAbsence
				if !errors.Is(err, boom) || errors.Is(err, context.Canceled) != cancelOnClose || errors.As(err, &absence) || calls != 1 {
					t.Fatalf("calls=%d err=%v", calls, err)
				}
			})
		}
	}
}
func TestIncidentAllocationOperandBounds(t *testing.T) {
	for _, operand := range []string{"path", "after", "deadline"} {
		t.Run(operand, func(t *testing.T) {
			tx, _ := mockTx(t)
			ctx := preconditionContext(t)
			path := "beads/plan"
			window := pageWindow{limit: 1}
			limits := fixtureLimits
			limits.valueBytes = 16
			want := errBudget
			switch operand {
			case "path":
				path = "beads/" + strings.Repeat("a", 17)
			case "after":
				window.afterPath = "links/" + strings.Repeat("a", 17)
			case "deadline":
				ctx = context.Background()
				want = errObservationDeadline
			}
			page, err := readIncidentPageInTx(ctx, tx, fixtureScope, path, graph.DirectionBoth, window, limits)
			requireEmptyPage(t, page)
			if !errors.Is(err, want) {
				t.Fatal(err)
			}
		})
	}
}
func TestIncidentAllocationRepeatedMetadataBudget(t *testing.T) {
	for _, short := range []bool{false, true} {
		t.Run(fmt.Sprint(short), func(t *testing.T) {
			tx, m := mockTx(t)
			limits := fixtureLimits
			limits.valueBytes = 128
			first := incidentLiveValues("beads/plan", true, true, pageRow("links/a"))
			second := incidentLiveValues("beads/plan", true, true, pageRow("links/b"))
			total := 0
			for _, v := range append(append([]driver.Value{}, first...), second...) {
				switch x := v.(type) {
				case string:
					total += len(x)
				case []byte:
					total += len(x)
				}
			}
			limits.bytes = total
			if short {
				limits.bytes--
			}
			rows := incidentPageRows().AddRow(first...).AddRow(second...)
			query, args := incidentPageQuery("beads/plan", graph.DirectionBoth, pageWindow{limit: 1}, limits.valueBytes)
			values := make([]driver.Value, len(args))
			for i, v := range args {
				values[i] = v
			}
			m.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(values...).WillReturnRows(rows).RowsWillBeClosed()
			page, err := readIncidentPageInTx(preconditionContext(t), tx, fixtureScope, "beads/plan", graph.DirectionBoth, pageWindow{limit: 1}, limits)
			if short {
				requireEmptyPage(t, page)
				if !errors.Is(err, errBudget) {
					t.Fatal(err)
				}
			} else if err != nil || len(page.items) != 1 || !page.hasMore {
				t.Fatalf("page=%+v err=%v", page, err)
			}
		})
	}
}
func TestIncidentAllocationObservationStatementBudget(t *testing.T) {
	for _, state := range []string{"", "reserved", "pruned", "erased", "live"} {
		t.Run(state, func(t *testing.T) {
			tx, m := mockTx(t)
			for _, f := range observationFixtures() {
				if f.name == "lease" {
					found := false
					for i, column := range f.columns {
						if column == "expires" {
							f.values[i] = "2026-09-20 12:35:56.123456"
							found = true
						}
					}
					if !found {
						t.Fatal("missing lease expires projection")
					}
				}
				expectObservation(m, f).WillReturnRows(sqlmock.NewRows(f.columns).AddRow(f.values...)).RowsWillBeClosed()
			}
			m.ExpectQuery(regexp.QuoteMeta(stateObservationQuery)).WillReturnRows(sqlmock.NewRows(stateColumns).AddRow(stateValues()...)).RowsWillBeClosed()
			q := &preconditionQueryHook{tx: tx}
			ctx := preconditionContext(t)
			n := uint64(3)
			facts, err := observePreconditionsInTx(ctx, q, &n)
			if err != nil {
				t.Fatal(err)
			}
			// Synthetic value-only witness for composition/counting, never a
			// provider grant. Expected values come from the independent SQL rows.
			version, err := composeStateVersion(stateLabels())
			if err != nil {
				t.Fatal(err)
			}
			witness, err := authority.NewWitness(authority.WitnessFields{
				ScopeURL: "https://example.com/scope/", AuthorityID: strings.Repeat("a", 32), Epoch: ^uint64(0),
				LedgerSeq: 3, LedgerHash: strings.Repeat("a", 64), StateVersion: version,
				StateCommit: strings.Repeat("v", 32), GrantedAt: "2026-09-20T12:34:56.123456Z",
			})
			if err != nil {
				t.Fatal(err)
			}
			key := strings.Repeat("b", 64)
			expected := expectedReadFacts{witness: &witness, installationKey: key, witnessInstallationKey: key, database: "graph_fixture", branch: "main"}
			if err := checkReadFacts(ctx, expected, facts); err != nil {
				t.Fatal(err)
			}
			if q.calls != 4 {
				t.Fatal(q.calls)
			}
			v := incidentAbsenceValues(state)
			if state == "live" {
				v = incidentLiveValues("beads/plan", true, true, pageRow("links/a"))
			}
			expectIncidentPage(m, graph.DirectionBoth, "", 1, incidentPageRows().AddRow(v...))
			page, err := readIncidentPageInTx(ctx, q, facts.scope.url, "beads/plan", graph.DirectionBoth, pageWindow{limit: 1}, fixtureLimits)
			if q.calls != 5 {
				t.Fatal(q.calls)
			}
			if state == "live" {
				if err != nil || len(page.items) != 1 {
					t.Fatal(err)
				}
			} else {
				var absence *exactAbsence
				if !errors.As(err, &absence) || absence.state != state {
					t.Fatal(err)
				}
				requireEmptyPage(t, page)
			}
		})
	}
}

func TestIncidentAllocationSentinelBudgetPrecedesAbsence(t *testing.T) {
	for _, bytes := range []int{32, 31} {
		t.Run(fmt.Sprint(bytes), func(t *testing.T) {
			tx, m := mockTx(t)
			limits := fixtureLimits
			limits.valueBytes, limits.bytes = 10, bytes
			query, args := incidentPageQuery("beads/plan", graph.DirectionBoth, pageWindow{limit: 1}, limits.valueBytes)
			values := make([]driver.Value, len(args))
			for i, v := range args {
				values[i] = v
			}
			m.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(values...).WillReturnRows(incidentPageRows().AddRow(incidentAbsenceValues("reserved")...)).RowsWillBeClosed()
			page, err := readIncidentPageInTx(preconditionContext(t), tx, fixtureScope, "beads/plan", graph.DirectionBoth, pageWindow{limit: 1}, limits)
			requireEmptyPage(t, page)
			var absence *exactAbsence
			// Independently specified request(10)+allocation path(10)+kind(4)+state(8).
			if bytes == 32 {
				if !errors.As(err, &absence) || absence.state != "reserved" || errors.Is(err, errBudget) {
					t.Fatal(err)
				}
			} else if !errors.Is(err, errBudget) || errors.As(err, &absence) {
				t.Fatal(err)
			}
		})
	}
}
