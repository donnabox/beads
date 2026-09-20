package graphops

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	graph "github.com/steveyegge/beads/graphops"
)

const observationTime = "2026-09-20 12:34:56.123456"

var scopeObservationColumns = []string{"id", "url", "authority", "epoch", "minted"}
var ledgerObservationColumns = []string{"recorded_seq", "recorded_hash", "tip_seq", "tip_hash", "head"}
var leaseObservationColumns = []string{"present", "id", "url", "authority", "holder", "renewer", "epoch", "granted", "expires", "heartbeat", "fence", "clock", "zone"}

func scopeObservationValues() []driver.Value {
	return []driver.Value{int64(1), "https://example.com/scope/", strings.Repeat("a", 32), "18446744073709551615", observationTime}
}
func ledgerObservationValues() []driver.Value {
	return []driver.Value{int64(3), strings.Repeat("a", 64), int64(8), strings.Repeat("b", 64), strings.Repeat("v", 32)}
}
func leaseObservationValues() []driver.Value {
	return []driver.Value{int64(1), int64(1), "https://example.com/scope/", strings.Repeat("a", 32), strings.Repeat("b", 64), strings.Repeat("c", 32), "18446744073709551615", observationTime, "2026-09-20 12:34:50.000000", observationTime, strings.Repeat("d", 32), observationTime, "+00:00"}
}
func preconditionContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	t.Cleanup(cancel)
	return ctx
}

type observationFixture struct {
	name, query string
	columns     []string
	values      []driver.Value
	args        []driver.Value
	observe     func(context.Context, queryer) (any, error)
	zero        any
}

func observationFixtures() []observationFixture {
	n := uint64(3)
	return []observationFixture{
		{"scope", scopeObservationQuery, scopeObservationColumns, scopeObservationValues(), nil, func(c context.Context, q queryer) (any, error) { return observeScopeInTx(c, q) }, scopeObservation{}},
		{"ledger", ledgerObservationQuery, ledgerObservationColumns, ledgerObservationValues(), []driver.Value{"3"}, func(c context.Context, q queryer) (any, error) { return observeLedgerInTx(c, q, &n) }, ledgerObservation{}},
		{"lease", leaseObservationQuery, leaseObservationColumns, leaseObservationValues(), nil, func(c context.Context, q queryer) (any, error) { return observeLeaseInTx(c, q) }, leaseObservation{}},
	}
}
func expectObservation(m sqlmock.Sqlmock, f observationFixture) *sqlmock.ExpectedQuery {
	return m.ExpectQuery(regexp.QuoteMeta(f.query)).WithArgs(f.args...)
}

func TestPreconditionFactsPreserveDriverTextAndUnsignedRange(t *testing.T) {
	for _, f := range observationFixtures() {
		for _, asBytes := range []bool{false, true} {
			t.Run(f.name+strconv.FormatBool(asBytes), func(t *testing.T) {
				tx, m := mockTx(t)
				values := append([]driver.Value(nil), f.values...)
				if asBytes {
					for i, value := range values {
						if s, ok := value.(string); ok {
							values[i] = []byte(s)
						}
					}
				}
				expectObservation(m, f).WillReturnRows(sqlmock.NewRows(f.columns).AddRow(values...)).RowsWillBeClosed()
				got, err := f.observe(preconditionContext(t), tx)
				if err != nil {
					t.Fatal(err)
				}
				switch v := got.(type) {
				case scopeObservation:
					if v.presence != observationPresent || v.epoch != ^uint64(0) || v.url != "https://example.com/scope/" || v.mintedAt != observationTime {
						t.Fatalf("scope=%+v", v)
					}
				case ledgerObservation:
					if v.requested == nil || *v.requested != 3 || v.recorded.seq != 3 || v.tip.seq != 8 || v.head != strings.Repeat("v", 32) {
						t.Fatalf("ledger=%+v", v)
					}
				case leaseObservation:
					if v.presence != observationPresent || v.epoch != ^uint64(0) || v.renewer != strings.Repeat("c", 32) || v.grantedAt != observationTime || v.heartbeatAt != observationTime || v.expiresAt != "2026-09-20 12:34:50.000000" || v.zone != "+00:00" || v.queryStarted.IsZero() || v.queryFinished.Before(v.queryStarted) {
						t.Fatalf("lease=%+v", v)
					}
				default:
					t.Fatalf("unexpected type %T", got)
				}
			})
		}
	}
}

func TestScopeAbsenceAndAnchoredAbsence(t *testing.T) {
	t.Run("scope", func(t *testing.T) {
		tx, m := mockTx(t)
		m.ExpectQuery(regexp.QuoteMeta(scopeObservationQuery)).WillReturnRows(sqlmock.NewRows(scopeObservationColumns)).RowsWillBeClosed()
		got, err := observeScopeInTx(preconditionContext(t), tx)
		if err != nil || got != (scopeObservation{presence: observationAbsent}) {
			t.Fatalf("got=%+v err=%v", got, err)
		}
	})
	for _, requested := range []*uint64{nil, new(uint64(12))} {
		t.Run("ledger requested="+strconv.FormatBool(requested != nil), func(t *testing.T) {
			tx, m := mockTx(t)
			var operand driver.Value
			if requested != nil {
				operand = "12"
			}
			m.ExpectQuery(regexp.QuoteMeta(ledgerObservationQuery)).WithArgs(operand).WillReturnRows(sqlmock.NewRows(ledgerObservationColumns).AddRow(nil, nil, nil, nil, strings.Repeat("0", 32))).RowsWillBeClosed()
			got, err := observeLedgerInTx(preconditionContext(t), tx, requested)
			if err != nil || got.recorded.presence != observationAbsent || got.tip.presence != observationAbsent || got.head != strings.Repeat("0", 32) || (got.requested == nil) != (requested == nil) {
				t.Fatalf("got=%+v err=%v", got, err)
			}
			if requested != nil {
				*requested = 99
				if *got.requested != 12 {
					t.Fatal("operand was not copied")
				}
			}
		})
	}
	for _, zone := range []string{"+00:00", "SYSTEM", "+05:00", "-07:00"} {
		t.Run("absent lease "+zone, func(t *testing.T) {
			tx, m := mockTx(t)
			values := make([]driver.Value, 13)
			values[11], values[12] = observationTime, zone
			m.ExpectQuery(regexp.QuoteMeta(leaseObservationQuery)).WillReturnRows(sqlmock.NewRows(leaseObservationColumns).AddRow(values...)).RowsWillBeClosed()
			got, err := observeLeaseInTx(preconditionContext(t), tx)
			if err != nil || got.presence != observationAbsent || got.clock != observationTime || got.zone != zone || got.holder != "" {
				t.Fatalf("got=%+v err=%v", got, err)
			}
		})
	}
}

func TestPreconditionFailuresDiscardEveryFact(t *testing.T) {
	boom := errors.New("observed engine error")
	for _, f := range observationFixtures() {
		for _, mode := range []string{"query", "scan", "iteration", "close", "duplicate", "empty"} {
			if f.name == "scope" && mode == "empty" {
				continue
			}
			t.Run(f.name+"/"+mode, func(t *testing.T) {
				tx, m := mockTx(t)
				q := expectObservation(m, f)
				rows := sqlmock.NewRows(f.columns)
				switch mode {
				case "query":
					q.WillReturnError(boom)
				case "scan":
					rows = sqlmock.NewRows([]string{"wrong"}).AddRow("bad")
				case "iteration":
					rows.AddRow(f.values...).AddRow(f.values...).RowError(1, boom)
				case "close":
					rows.AddRow(f.values...).CloseError(boom)
				case "duplicate":
					rows.AddRow(f.values...).AddRow(f.values...)
				}
				if mode != "query" {
					q.WillReturnRows(rows).RowsWillBeClosed()
				}
				got, err := f.observe(preconditionContext(t), tx)
				if err == nil || !reflect.DeepEqual(got, f.zero) {
					t.Fatalf("partial=%+v err=%v", got, err)
				}
				if (mode == "query" || mode == "iteration" || mode == "close") && !errors.Is(err, boom) {
					t.Fatalf("lost error %v", err)
				}
				if (mode == "duplicate" || mode == "empty") && (!errors.Is(err, errCorrupt) || errors.Is(err, graph.ErrValidation)) {
					t.Fatalf("wrong classification %v", err)
				}
			})
		}
	}
}

func TestPreconditionMalformedColumnsNeverBecomeAbsence(t *testing.T) {
	for _, f := range observationFixtures() {
		for column := range f.values {
			t.Run(f.name+"/NULL/"+strconv.Itoa(column), func(t *testing.T) {
				tx, m := mockTx(t)
				values := append([]driver.Value(nil), f.values...)
				values[column] = nil
				expectObservation(m, f).WillReturnRows(sqlmock.NewRows(f.columns).AddRow(values...)).RowsWillBeClosed()
				got, err := f.observe(preconditionContext(t), tx)
				if !errors.Is(err, errCorrupt) || errors.Is(err, graph.ErrValidation) || !reflect.DeepEqual(got, f.zero) {
					t.Fatalf("partial=%+v err=%v", got, err)
				}
			})
		}
	}
	for _, f := range observationFixtures() {
		for badIndex, bad := range []driver.Value{int64(0), int64(-1), int64(2), "01", "+1", true, float64(1)} {
			if f.name == "ledger" {
				continue
			}
			t.Run(f.name+"/bad-id/"+strconv.Itoa(badIndex), func(t *testing.T) {
				tx, m := mockTx(t)
				values := append([]driver.Value(nil), f.values...)
				slot := 0
				if f.name == "lease" {
					slot = 1
				}
				values[slot] = bad
				expectObservation(m, f).WillReturnRows(sqlmock.NewRows(f.columns).AddRow(values...)).RowsWillBeClosed()
				got, err := f.observe(preconditionContext(t), tx)
				if !errors.Is(err, errCorrupt) || !reflect.DeepEqual(got, f.zero) {
					t.Fatalf("partial=%+v err=%v", got, err)
				}
			})
		}
	}
}

func TestObservationScalarParsing(t *testing.T) {
	for _, good := range []any{int64(1), uint64(1), "1", []byte("1"), "18446744073709551615", uint64(1) << 63} {
		if _, err := observationUint(good, ^uint64(0)); err != nil {
			t.Fatalf("%v: %v", good, err)
		}
	}
	for _, bad := range []any{nil, int64(-1), uint64(0), "0", "01", "+1", " 1", "1 ", "1.0", "1e1", "18446744073709551616", true, float64(1), time.Now()} {
		if _, err := observationUint(bad, ^uint64(0)); !errors.Is(err, errCorrupt) {
			t.Fatalf("accepted %v: %v", bad, err)
		}
	}
	for _, bad := range []string{"2026-09-20 12:34:56", "2026-09-20 12:34:56.12345", "2026-09-20 12:34:56.1234567", "2026-02-30 12:34:56.000000", "0000-00-00 00:00:00.000000", "0000-01-01 00:00:00.000000", "2026-09-20T12:34:56.123456Z"} {
		if _, err := observationCivil(bad); !errors.Is(err, errCorrupt) {
			t.Fatalf("accepted timestamp %q: %v", bad, err)
		}
	}
	for _, good := range []string{observationTime, "1960-01-01 00:00:00.000000"} {
		if got, err := observationCivil(good); err != nil || string(got) != good {
			t.Fatalf("civil rewrite: %q %v", got, err)
		}
	}
	for _, bad := range []string{"HTTP://EXAMPLE.COM/scope/", "https://example.com/scope", string([]byte{255}), strings.Repeat("u", 8193)} {
		if _, err := observationScopeURL(bad); err == nil {
			t.Fatalf("accepted URL %q", bad)
		}
	}
	for _, bad := range []string{strings.Repeat("A", 32), strings.Repeat("g", 32), strings.Repeat("a", 31), strings.Repeat("a", 33)} {
		if _, err := observationLabel(bad, 32, 'f'); !errors.Is(err, errCorrupt) {
			t.Fatalf("accepted hex %q", bad)
		}
	}
}

func TestLedgerOperandsAndDivergentFacts(t *testing.T) {
	for _, bad := range []uint64{0, ^uint64(0)} {
		tx, _ := mockTx(t)
		got, err := observeLedgerInTx(preconditionContext(t), tx, &bad)
		if !errors.Is(err, errObservationOperand) || got.requested != nil {
			t.Fatalf("bad operand got=%+v err=%v", got, err)
		}
	}
	for _, seq := range []uint64{3, 1 << 63, graph.MaxLedgerSeq} {
		tx, m := mockTx(t)
		text := strconv.FormatUint(seq, 10)
		m.ExpectQuery(regexp.QuoteMeta(ledgerObservationQuery)).WithArgs(text).WillReturnRows(sqlmock.NewRows(ledgerObservationColumns).AddRow(text, strings.Repeat("c", 64), text, strings.Repeat("c", 64), strings.Repeat("a", 32))).RowsWillBeClosed()
		got, err := observeLedgerInTx(preconditionContext(t), tx, &seq)
		if err != nil || got.recorded.seq != seq || got.recorded.hash != strings.Repeat("c", 64) {
			t.Fatalf("got=%+v err=%v", got, err)
		}
	}
	tx, m := mockTx(t)
	seq := uint64(20)
	m.ExpectQuery(regexp.QuoteMeta(ledgerObservationQuery)).WithArgs("20").WillReturnRows(sqlmock.NewRows(ledgerObservationColumns).AddRow(nil, nil, int64(8), strings.Repeat("b", 64), strings.Repeat("a", 32))).RowsWillBeClosed()
	got, err := observeLedgerInTx(preconditionContext(t), tx, &seq)
	if err != nil || got.recorded.presence != observationAbsent || got.tip.seq != 8 {
		t.Fatalf("missing recorded row lost tip: %+v %v", got, err)
	}
}

type preconditionQueryHook struct {
	tx         *sql.Tx
	calls      int
	afterQuery func()
}

func (q *preconditionQueryHook) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	q.calls++
	rows, err := q.tx.QueryContext(ctx, query, args...)
	if q.afterQuery != nil {
		q.afterQuery()
	}
	return rows, err
}
func TestPreconditionDeadlineBeforeAndAfterQuery(t *testing.T) {
	for _, f := range observationFixtures() {
		t.Run(f.name, func(t *testing.T) {
			tx, _ := mockTx(t)
			q := &preconditionQueryHook{tx: tx}
			for _, ctx := range []context.Context{nil, context.Background()} {
				got, err := f.observe(ctx, q)
				if !errors.Is(err, errObservationDeadline) || !reflect.DeepEqual(got, f.zero) {
					t.Fatalf("deadline: %+v %v", got, err)
				}
			}
			ctx, cancel := context.WithTimeout(t.Context(), time.Second)
			cancel()
			got, err := f.observe(ctx, q)
			if !errors.Is(err, context.Canceled) || !reflect.DeepEqual(got, f.zero) || q.calls != 0 {
				t.Fatalf("cancel: %+v %v calls%d", got, err, q.calls)
			}
		})
	}
	for _, f := range observationFixtures() {
		t.Run(f.name+"/after-query", func(t *testing.T) {
			tx, m := mockTx(t)
			expectObservation(m, f).WillReturnRows(sqlmock.NewRows(f.columns).AddRow(f.values...)).RowsWillBeClosed()
			ctx, cancel := context.WithTimeout(t.Context(), time.Second)
			defer cancel()
			q := &preconditionQueryHook{tx: tx, afterQuery: cancel}
			got, err := f.observe(ctx, q)
			if !errors.Is(err, context.Canceled) || !reflect.DeepEqual(got, f.zero) || q.calls != 1 {
				t.Fatalf("late cancel: %+v %v", got, err)
			}
		})
	}
}

func TestPreconditionQueryStructure(t *testing.T) {
	for _, query := range []string{scopeObservationQuery, ledgerObservationQuery, leaseObservationQuery} {
		if strings.Contains(query, ";") {
			t.Fatal("multi-statement query")
		}
	}
	if strings.Contains(scopeObservationQuery, "WHERE") || strings.Contains(leaseObservationQuery, "WHERE") {
		t.Fatal("singleton predicate hides corrupt extra rows")
	}
	if !strings.Contains(ledgerObservationQuery, "ORDER BY seq DESC LIMIT 1") || !strings.Contains(ledgerObservationQuery, "DOLT_HASHOF('HEAD')") || strings.Count(ledgerObservationQuery, "?") != 1 {
		t.Fatal("lost tip/HEAD or bound operand")
	}
	if !strings.Contains(leaseObservationQuery, "LEFT JOIN") || !strings.Contains(leaseObservationQuery, "NOW(6)") || !strings.Contains(leaseObservationQuery, "@@session.time_zone") {
		t.Fatal("absent lease lost clock/zone")
	}
}

func TestPreconditionAggregateOrderAndFailureAtomicity(t *testing.T) {
	boom := errors.New("precondition query refused")
	for stop := 0; stop <= 4; stop++ {
		t.Run(strconv.Itoa(stop), func(t *testing.T) {
			tx, m := mockTx(t)
			fixtures := observationFixtures()
			for i, f := range fixtures {
				q := expectObservation(m, f)
				if stop == i+1 {
					q.WillReturnError(boom)
					break
				}
				q.WillReturnRows(sqlmock.NewRows(f.columns).AddRow(f.values...)).RowsWillBeClosed()
			}
			if stop == 0 || stop == 4 {
				q := m.ExpectQuery(regexp.QuoteMeta(stateObservationQuery)).WithoutArgs()
				if stop == 4 {
					q.WillReturnError(boom)
				} else {
					q.WillReturnRows(sqlmock.NewRows(stateColumns).AddRow(stateValues()...)).RowsWillBeClosed()
				}
			}
			n := uint64(3)
			counter := &preconditionQueryHook{tx: tx}
			got, err := observePreconditionsInTx(preconditionContext(t), counter, &n)
			if stop != 0 {
				if !errors.Is(err, boom) || !reflect.DeepEqual(got, preconditionObservations{}) || counter.calls != stop {
					t.Fatalf("stop=%d partial=%+v calls=%d err=%v", stop, got, counter.calls, err)
				}
			} else if err != nil || counter.calls != 4 || got.scope.presence != observationPresent || got.ledger.tip.seq != 8 || got.lease.zone != "+00:00" || got.state.descriptorHash() != strings.Repeat("2", 32) {
				t.Fatalf("facts=%+v calls=%d err=%v", got, counter.calls, err)
			}
		})
	}
	tx, _ := mockTx(t)
	n := uint64(0)
	got, err := observePreconditionsInTx(preconditionContext(t), tx, &n)
	if !errors.Is(err, errObservationOperand) || !reflect.DeepEqual(got, preconditionObservations{}) {
		t.Fatalf("bad aggregate operand issued query: %+v %v", got, err)
	}
}

func TestPreconditionBoundsAndLeaseAbsenceRefusals(t *testing.T) {
	for _, tc := range []struct {
		name   string
		slot   int
		value  driver.Value
		budget bool
	}{
		{"zone bound", 12, strings.Repeat("z", 257), true},
		{"scope byte bound", 2, strings.Repeat("u", 8193), true},
		{"row bound", 2, strings.Repeat("u", observationRowBytes+1), true},
		{"wrong marker", 0, int64(2), false},
		{"invalid scope", 2, "https://example.com/scope", false},
		{"wrong holder width", 4, strings.Repeat("b", 32), false},
		{"upper authority", 3, strings.Repeat("A", 32), false},
		{"nonhex authority", 3, strings.Repeat("g", 32), false},
		{"upper holder", 4, strings.Repeat("B", 64), false},
		{"nonhex holder", 4, strings.Repeat("g", 64), false},
		{"nonhex renewer", 5, strings.Repeat("g", 32), false},
		{"zero epoch", 6, "0", false},
		{"invalid grant", 7, "0000-00-00 00:00:00.000000", false},
		{"invalid expiry", 8, "2026-02-30 00:00:00.000000", false},
		{"invalid heartbeat", 9, "2026-09-20T12:34:56Z", false},
		{"upper fence", 10, strings.Repeat("D", 32), false},
		{"invalid clock", 11, "2026-09-20 12:34:56.1234567", false},
		{"invalid zone UTF8", 12, []byte{255}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tx, m := mockTx(t)
			values := leaseObservationValues()
			values[tc.slot] = tc.value
			m.ExpectQuery(regexp.QuoteMeta(leaseObservationQuery)).WillReturnRows(sqlmock.NewRows(leaseObservationColumns).AddRow(values...)).RowsWillBeClosed()
			got, err := observeLeaseInTx(preconditionContext(t), tx)
			want := errCorrupt
			if tc.budget {
				want = errBudget
			}
			if !errors.Is(err, want) || got != (leaseObservation{}) {
				t.Fatalf("partial=%+v err=%v", got, err)
			}
		})
	}
	for _, slot := range []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12} {
		t.Run("absent lease column "+strconv.Itoa(slot), func(t *testing.T) {
			tx, m := mockTx(t)
			values := make([]driver.Value, 13)
			values[11], values[12] = observationTime, "+00:00"
			if slot < 11 {
				values[slot] = leaseObservationValues()[slot]
			} else {
				values[slot] = nil
			}
			m.ExpectQuery(regexp.QuoteMeta(leaseObservationQuery)).WillReturnRows(sqlmock.NewRows(leaseObservationColumns).AddRow(values...)).RowsWillBeClosed()
			got, err := observeLeaseInTx(preconditionContext(t), tx)
			if !errors.Is(err, errCorrupt) || got != (leaseObservation{}) {
				t.Fatalf("partial=%+v err=%v", got, err)
			}
		})
	}
}

func TestLedgerQueryConsistencyAndCommitAlphabet(t *testing.T) {
	for _, tc := range []struct {
		name      string
		requested *uint64
		change    func([]driver.Value)
	}{
		{"no requested row", nil, func(v []driver.Value) {}},
		{"wrong recorded seq", new(uint64(4)), func(v []driver.Value) {}},
		{"recorded beyond maximum", new(uint64(3)), func(v []driver.Value) { v[2] = int64(2) }},
		{"recorded without tip", new(uint64(3)), func(v []driver.Value) { v[2], v[3] = nil, nil }},
		{"same seq different hashes", new(uint64(3)), func(v []driver.Value) { v[2] = int64(3) }},
		{"tip out of range", new(uint64(3)), func(v []driver.Value) { v[2] = "18446744073709551615" }},
		{"nonbase32 HEAD", new(uint64(3)), func(v []driver.Value) { v[4] = strings.Repeat("z", 32) }},
		{"upper HEAD", new(uint64(3)), func(v []driver.Value) { v[4] = strings.Repeat("A", 32) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tx, m := mockTx(t)
			values := ledgerObservationValues()
			tc.change(values)
			var arg driver.Value
			if tc.requested != nil {
				arg = strconv.FormatUint(*tc.requested, 10)
			}
			m.ExpectQuery(regexp.QuoteMeta(ledgerObservationQuery)).WithArgs(arg).WillReturnRows(sqlmock.NewRows(ledgerObservationColumns).AddRow(values...)).RowsWillBeClosed()
			got, err := observeLedgerInTx(preconditionContext(t), tx, tc.requested)
			if !errors.Is(err, errCorrupt) || !reflect.DeepEqual(got, ledgerObservation{}) {
				t.Fatalf("partial=%+v err=%v", got, err)
			}
		})
	}
}

// This counts raw observations plus row bodies in one transaction. It does not
// turn the fixture facts into a witness, claim or protected Reader.
func TestPreconditionAndBodyStatementBudgets(t *testing.T) {
	for _, method := range []string{"bead", "link", "incident"} {
		t.Run(method, func(t *testing.T) {
			tx, m := mockTx(t)
			for _, f := range observationFixtures() {
				expectObservation(m, f).WillReturnRows(sqlmock.NewRows(f.columns).AddRow(f.values...)).RowsWillBeClosed()
			}
			m.ExpectQuery(regexp.QuoteMeta(stateObservationQuery)).WillReturnRows(sqlmock.NewRows(stateColumns).AddRow(stateValues()...)).RowsWillBeClosed()
			q := &preconditionQueryHook{tx: tx}
			ctx := preconditionContext(t)
			n := uint64(3)
			if _, err := observePreconditionsInTx(ctx, q, &n); err != nil {
				t.Fatal(err)
			}
			var err error
			want := 5
			switch method {
			case "bead":
				r := validRow()
				expectBead(m, r)
				decl, e := graph.NewOwnedLinkDecl(relationType, "Explains", 2)
				if e != nil {
					t.Fatal(e)
				}
				expectDescriptor(m, memoryDescriptor(t, decl))
				expectLinks(m, "source_kind = 'in' AND source_path = ? AND type_url IN (?)", []driver.Value{r.path, relationType}, sqlmock.NewRows(allLinkColumns), 2)
				_, err = readBeadInTx(ctx, q, fixtureScope, r.path, fixtureLimits)
				want = 7
			case "link":
				r := validLinkRow()
				d := relationDescriptor(t)
				columns := append(append([]string{}, allLinkColumns...), "url", "descriptor", "descriptor_length", "fingerprint")
				values := append(linkValues(r), d.ID(), d.CanonicalJSON(), blobLength(d.CanonicalJSON()), d.Fingerprint())
				m.ExpectQuery(regexp.QuoteMeta(exactLinkQuery)).WithArgs(fixtureLimits.valueBytes, fixtureLimits.valueBytes, r.path).WillReturnRows(sqlmock.NewRows(columns).AddRow(values...)).RowsWillBeClosed()
				_, err = readLinkInTx(ctx, q, fixtureScope, r.path, fixtureLimits)
			case "incident":
				expectIncident(m, graph.DirectionBoth, fixtureLimits, incidentRows(validLinkRow()))
				_, err = readIncidentLinksInTx(ctx, q, fixtureScope, "beads/plan", graph.DirectionBoth, fixtureLimits)
			}
			if err != nil || q.calls != want {
				t.Fatalf("%s calls%d want%d err%v", method, q.calls, want, err)
			}
		})
	}
}
