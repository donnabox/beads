package graphops

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	graph "github.com/steveyegge/beads/graphops"
)

// Real embedded rows preserve native integer kinds through Scan(*any). Bypass
// sqlmock's default converter, which would widen int8 and reject high-bit uint64.
type preconditionPassthrough struct{}

func (preconditionPassthrough) ConvertValue(v any) (driver.Value, error) { return v, nil }
func nativeObservationMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, m, err := sqlmock.New(sqlmock.ValueConverterOption(preconditionPassthrough{}))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		m.ExpectClose()
		if err := db.Close(); err != nil {
			t.Error(err)
		}
		if err := m.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})
	return db, m
}

func TestPreconditionNativeIntegerRows(t *testing.T) {
	for _, scalar := range []any{int(1), int8(1), int16(1), int32(1), int64(1), uint(1), uint8(1), uint16(1), uint32(1), uint64(1)} {
		for _, f := range observationFixtures() {
			if f.name == "ledger" {
				continue
			}
			t.Run(fmt.Sprintf("%s/%T", f.name, scalar), func(t *testing.T) {
				db, m := nativeObservationMock(t)
				values := append([]driver.Value(nil), f.values...)
				values[0] = scalar
				if f.name == "lease" {
					values[1] = scalar
				}
				expectObservation(m, f).WillReturnRows(m.NewRows(f.columns).AddRow(values...)).RowsWillBeClosed()
				got, err := f.observe(preconditionContext(t), db)
				if err != nil || reflect.DeepEqual(got, f.zero) {
					t.Fatalf("native %T: %+v %v", scalar, got, err)
				}
			})
		}
	}
	for _, seq := range []uint64{1, 1 << 63, graph.MaxLedgerSeq} {
		t.Run(strconv.FormatUint(seq, 10), func(t *testing.T) {
			db, m := nativeObservationMock(t)
			hash := strings.Repeat("a", 64)
			m.ExpectQuery(regexp.QuoteMeta(ledgerObservationQuery)).WithArgs(strconv.FormatUint(seq, 10)).WillReturnRows(m.NewRows(ledgerObservationColumns).AddRow(seq, hash, seq, hash, strings.Repeat("v", 32))).RowsWillBeClosed()
			got, err := observeLedgerInTx(preconditionContext(t), db, &seq)
			if err != nil || got.recorded.seq != seq || got.tip.seq != seq || got.recorded.hash != hash {
				t.Fatalf("native uint64: %+v %v", got, err)
			}
		})
	}
	for _, bad := range []any{float32(1), float64(1), true, int8(-1), int32(-1), uint8(0), time.Now()} {
		t.Run(fmt.Sprintf("reject/%T", bad), func(t *testing.T) {
			db, m := nativeObservationMock(t)
			values := scopeObservationValues()
			values[0] = bad
			m.ExpectQuery(regexp.QuoteMeta(scopeObservationQuery)).WillReturnRows(m.NewRows(scopeObservationColumns).AddRow(values...)).RowsWillBeClosed()
			got, err := observeScopeInTx(preconditionContext(t), db)
			if !errors.Is(err, errCorrupt) || got != (scopeObservation{}) {
				t.Fatalf("accepted %T: %+v %v", bad, got, err)
			}
		})
	}
}

func TestScopeMalformedRowFields(t *testing.T) {
	for _, tc := range []struct {
		name  string
		slot  int
		value driver.Value
	}{
		{"noncanonical URL", 1, "https://example.com/scope"},
		{"upper authority", 2, strings.Repeat("A", 32)},
		{"nonhex authority", 2, strings.Repeat("g", 32)},
		{"zero epoch", 3, "0"},
		{"epoch overflow", 3, "18446744073709551616"},
		{"invalid minted date", 4, "2026-02-30 12:34:56.000000"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tx, m := mockTx(t)
			values := scopeObservationValues()
			values[tc.slot] = tc.value
			m.ExpectQuery(regexp.QuoteMeta(scopeObservationQuery)).WillReturnRows(sqlmock.NewRows(scopeObservationColumns).AddRow(values...)).RowsWillBeClosed()
			got, err := observeScopeInTx(preconditionContext(t), tx)
			if !errors.Is(err, errCorrupt) || errors.Is(err, graph.ErrValidation) || got != (scopeObservation{}) {
				t.Fatalf("malformed row: %+v %v", got, err)
			}
		})
	}
}

// This fake driver cancels the observer's context specifically at Rows.Close.
// The SQL rows use an independent context, so database/sql cannot race to supply
// the cancellation error: observeRow's post-read context guard must refuse.
type closeCancelDriver struct{ cancel context.CancelFunc }

func (d closeCancelDriver) Open(string) (driver.Conn, error)             { return closeCancelConn{d.cancel}, nil }
func (d closeCancelDriver) Connect(context.Context) (driver.Conn, error) { return d.Open("") }
func (d closeCancelDriver) Driver() driver.Driver                        { return d }

type closeCancelConn struct{ cancel context.CancelFunc }

func (closeCancelConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("unexpected prepare")
}
func (closeCancelConn) Close() error              { return nil }
func (closeCancelConn) Begin() (driver.Tx, error) { return nil, errors.New("unexpected transaction") }
func (c closeCancelConn) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	return &closeCancelRows{cancel: c.cancel}, nil
}

type closeCancelRows struct {
	cancel    context.CancelFunc
	delivered bool
}

func (*closeCancelRows) Columns() []string { return scopeObservationColumns }
func (r *closeCancelRows) Close() error    { r.cancel(); return nil }
func (r *closeCancelRows) Next(dest []driver.Value) error {
	if r.delivered {
		return io.EOF
	}
	r.delivered = true
	copy(dest, scopeObservationValues())
	return nil
}

type independentRowsContext struct{ db *sql.DB }

func (q independentRowsContext) QueryContext(_ context.Context, query string, args ...any) (*sql.Rows, error) {
	return q.db.QueryContext(context.Background(), query, args...)
}
func TestPreconditionCancellationOnRowsClose(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	db := sql.OpenDB(closeCancelDriver{cancel})
	defer func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	}()
	got, err := observeScopeInTx(ctx, independentRowsContext{db})
	if !errors.Is(err, context.Canceled) || got != (scopeObservation{}) {
		t.Fatalf("close cancellation: %+v %v", got, err)
	}
}

// Independent ordered projection expectations cover the fields consumed by
// Scan. Clause checks separately retain bounds and absent-row anchors; actual
// embedded controls also exercise non-null fields and singleton absence.
func TestPreconditionProjectionAndAnchors(t *testing.T) {
	normalize := func(s string) string { return strings.Join(strings.Fields(s), " ") }
	cases := []struct {
		query, projection string
		clauses           []string
	}{
		{scopeObservationQuery, "id, scope_url, authority_id, CAST(authority_epoch AS CHAR), DATE_FORMAT(minted_at, '%Y-%m-%d %H:%i:%s.%f')", []string{"FROM graph_scope LIMIT 2"}},
		{ledgerObservationQuery, "r.seq, r.hash, tip.seq, tip.hash, DOLT_HASHOF('HEAD')", []string{"FROM (SELECT 1 AS anchor) AS a", "WHERE seq = ? LIMIT 2", "ORDER BY seq DESC LIMIT 1", ") AS tip ON TRUE LIMIT 2"}},
		{leaseObservationQuery, "l.row_present, l.id, l.scope_url, l.authority_id, l.holder_installation_key, l.renewer, CAST(l.authority_epoch AS CHAR), DATE_FORMAT(l.granted_at, '%Y-%m-%d %H:%i:%s.%f'), DATE_FORMAT(l.expires_at, '%Y-%m-%d %H:%i:%s.%f'), DATE_FORMAT(l.heartbeat_at, '%Y-%m-%d %H:%i:%s.%f'), l.fence, DATE_FORMAT(NOW(6), '%Y-%m-%d %H:%i:%s.%f'), @@session.time_zone", []string{"FROM (SELECT 1 AS anchor) AS a", "SELECT 1 AS row_present, id, scope_url, authority_id, holder_installation_key, renewer, authority_epoch, granted_at, expires_at, heartbeat_at, fence FROM graph_authority_lease LIMIT 2", ") AS l ON TRUE LIMIT 2"}},
	}
	for i, tc := range cases {
		query := normalize(tc.query)
		before, _, ok := strings.Cut(query, " FROM ")
		if !ok || before != "SELECT "+tc.projection {
			t.Fatalf("query%d projection: %s", i, before)
		}
		for _, clause := range tc.clauses {
			if !strings.Contains(query, clause) {
				t.Fatalf("query%d lost clause %q", i, clause)
			}
		}
	}
}
