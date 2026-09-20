package graphops

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"testing"
)

// The request context is independent from database/sql's row context, avoiding
// a race with its own cancellation watcher. These controls test the observed
// Rows.Close boundary, not an impossible guarantee against cancellation after
// the helper's final check.
type completionCloseDriver struct {
	columns    []string
	values     []driver.Value
	cancel     context.CancelFunc
	closeError error
	calls      *int
}

func (d completionCloseDriver) Open(string) (driver.Conn, error)             { return completionCloseConn{d}, nil }
func (d completionCloseDriver) Connect(context.Context) (driver.Conn, error) { return d.Open("") }
func (d completionCloseDriver) Driver() driver.Driver                        { return d }

type completionCloseConn struct{ d completionCloseDriver }

func (completionCloseConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("unexpected prepare")
}
func (completionCloseConn) Close() error { return nil }
func (completionCloseConn) Begin() (driver.Tx, error) {
	return nil, errors.New("unexpected transaction")
}
func (c completionCloseConn) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	*c.d.calls++
	return &completionCloseRows{d: c.d}, nil
}

type completionCloseRows struct {
	d         completionCloseDriver
	delivered bool
}

func (r *completionCloseRows) Columns() []string { return r.d.columns }
func (r *completionCloseRows) Close() error      { r.d.cancel(); return r.d.closeError }
func (r *completionCloseRows) Next(dest []driver.Value) error {
	if r.delivered || r.d.values == nil {
		return io.EOF
	}
	r.delivered = true
	copy(dest, r.d.values)
	return nil
}

func TestReadRowsCancellationAtCompletion(t *testing.T) {
	boom := errors.New("driver close failed")
	for _, present := range []bool{false, true} {
		for _, failClose := range []bool{false, true} {
			t.Run(map[bool]string{false: "empty", true: "row"}[present]+"/"+map[bool]string{false: "clean-close", true: "close-error"}[failClose], func(t *testing.T) {
				ctx, cancel := context.WithCancel(preconditionContext(t))
				defer cancel()
				var values []driver.Value
				if present {
					values = []driver.Value{int64(7)}
				}
				var closeErr error
				if failClose {
					closeErr = boom
				}
				calls := 0
				db := sql.OpenDB(completionCloseDriver{columns: []string{"value"}, values: values, cancel: cancel, closeError: closeErr, calls: &calls})
				defer func() {
					if err := db.Close(); err != nil {
						t.Error(err)
					}
				}()
				got, err := readRows(ctx, independentRowsContext{db}, "SELECT value", nil, 1, func(rows *sql.Rows) (int, error) { var v int; err := rows.Scan(&v); return v, err })
				if got != nil || !errors.Is(err, context.Canceled) || calls != 1 {
					t.Fatalf("result=%v calls=%d err=%v", got, calls, err)
				}
				if closeErr != nil && !errors.Is(err, closeErr) {
					t.Fatalf("lost close error: %v", err)
				}
			})
		}
	}
}

func TestReadRowsCanceledBeforeDispatch(t *testing.T) {
	ctx, cancel := context.WithCancel(preconditionContext(t))
	cancel()
	calls := 0
	db := sql.OpenDB(completionCloseDriver{columns: []string{"value"}, values: []driver.Value{int64(7)}, cancel: cancel, calls: &calls})
	defer func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	}()
	got, err := readRows(ctx, independentRowsContext{db}, "SELECT value", nil, 1, func(rows *sql.Rows) (int, error) { var v int; err := rows.Scan(&v); return v, err })
	if got != nil || !errors.Is(err, context.Canceled) || calls != 0 {
		t.Fatalf("result=%v calls=%d err=%v", got, calls, err)
	}
}
