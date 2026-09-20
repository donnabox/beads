package graphops

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"testing"

	graph "github.com/steveyegge/beads/graphops"
)

// The request context is independent from database/sql's row context, avoiding
// a race with its own cancellation watcher. These controls test the observed
// Rows.Close boundary, not an impossible guarantee against cancellation after
// the helper's final check.
type bodyCloseDriver struct {
	columns    []string
	values     []driver.Value
	cancel     context.CancelFunc
	closeError error
	calls      *int
}

func (d bodyCloseDriver) Open(string) (driver.Conn, error)             { return bodyCloseConn{d}, nil }
func (d bodyCloseDriver) Connect(context.Context) (driver.Conn, error) { return d.Open("") }
func (d bodyCloseDriver) Driver() driver.Driver                        { return d }

type bodyCloseConn struct{ d bodyCloseDriver }

func (bodyCloseConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("unexpected prepare")
}
func (bodyCloseConn) Close() error              { return nil }
func (bodyCloseConn) Begin() (driver.Tx, error) { return nil, errors.New("unexpected transaction") }
func (c bodyCloseConn) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	*c.d.calls++
	return &bodyCloseRows{d: c.d}, nil
}

type bodyCloseRows struct {
	d         bodyCloseDriver
	delivered bool
}

func (r *bodyCloseRows) Columns() []string { return r.d.columns }
func (r *bodyCloseRows) Close() error      { r.d.cancel(); return r.d.closeError }
func (r *bodyCloseRows) Next(dest []driver.Value) error {
	if r.delivered {
		return io.EOF
	}
	r.delivered = true
	copy(dest, r.d.values)
	return nil
}

func TestBodyCancellationOnRowsClose(t *testing.T) {
	boom := errors.New("driver close failed")
	for _, kind := range []graph.ResourceKind{graph.KindBead, graph.KindLink} {
		for _, present := range []bool{false, true} {
			for _, closeErr := range []error{nil, boom} {
				state := graph.AllocationPruned
				if present {
					state = graph.AllocationLive
				}
				t.Run(string(kind)+"/"+state+"/"+map[bool]string{false: "clean-close", true: "close-error"}[closeErr != nil], func(t *testing.T) {
					ctx, cancel := context.WithCancel(preconditionContext(t))
					defer cancel()
					columns, values := exactTestRow(t, kind, state, true, present)
					calls := 0
					db := sql.OpenDB(bodyCloseDriver{columns: columns, values: values, cancel: cancel, closeError: closeErr, calls: &calls})
					defer func() {
						if err := db.Close(); err != nil {
							t.Error(err)
						}
					}()
					zero, err := exactTestRead(t, ctx, independentRowsContext{db}, kind)
					var absence *exactAbsence
					if !zero || !errors.Is(err, context.Canceled) || errors.As(err, &absence) || calls != 1 {
						t.Fatalf("zero=%v calls=%d error=%v", zero, calls, err)
					}
					if closeErr != nil && !errors.Is(err, closeErr) {
						t.Fatalf("lost driver close error: %v", err)
					}
				})
			}
		}
	}
}
func TestBodyCanceledBeforeQuery(t *testing.T) {
	ctx, cancel := context.WithCancel(preconditionContext(t))
	cancel()
	tx, _ := mockTx(t)
	q := &preconditionQueryHook{tx: tx}
	for _, kind := range []graph.ResourceKind{graph.KindBead, graph.KindLink} {
		zero, err := exactTestRead(t, ctx, q, kind)
		if !zero || !errors.Is(err, context.Canceled) || q.calls != 0 {
			t.Fatalf("zero=%v calls=%d err=%v", zero, q.calls, err)
		}
	}
}
