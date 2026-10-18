//go:build cgo

package graphstore

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
)

// This tests error wrapping at the SQL boundary without opening an engine.
// The double only returns an empty COUNT result or the selected query failure;
// it does not model tables, mutations or transaction behavior.
func TestLegacyImportFreshnessQueryErrors(t *testing.T) {
	for _, test := range []struct {
		name, query, plane string
	}{
		{"table", "SELECT COUNT(*) FROM graph_preview_catalog", "graph_preview_catalog"},
		{"journal-sequence", "SELECT COUNT(*) FROM bd_events_seq WHERE id<>0 OR next_seq<>0", "bd_events_seq"},
		{"config", "SELECT COUNT(*) FROM config WHERE LEFT(`key`,?)=?", "config"},
	} {
		t.Run(test.name, func(t *testing.T) {
			cause := errors.New("freshness query unavailable")
			connector := &legacyFreshnessCountDriver{query: test.query, cause: cause}
			db := sql.OpenDB(connector)
			t.Cleanup(func() {
				if err := db.Close(); err != nil {
					t.Error(err)
				}
			})
			tx, err := db.BeginTx(t.Context(), nil)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := tx.Rollback(); err != nil {
					t.Error(err)
				}
			}()
			err = requireEmptyLegacyTarget(t.Context(), tx)
			if !connector.failed || !errors.Is(err, ErrInvalidStore) || !errors.Is(err, cause) || !strings.Contains(err.Error(), "from "+test.plane+";") {
				t.Fatalf("freshness error lost plane or cause: reached=%t err=%v", connector.failed, err)
			}
		})
	}
}

type legacyFreshnessCountDriver struct {
	query  string
	cause  error
	failed bool
}

func (d *legacyFreshnessCountDriver) Connect(context.Context) (driver.Conn, error) {
	return &legacyFreshnessCountConn{source: d}, nil
}
func (d *legacyFreshnessCountDriver) Driver() driver.Driver { return d }
func (d *legacyFreshnessCountDriver) Open(string) (driver.Conn, error) {
	return d.Connect(context.Background())
}

type legacyFreshnessCountConn struct{ source *legacyFreshnessCountDriver }

func (*legacyFreshnessCountConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("freshness count double does not prepare statements")
}
func (*legacyFreshnessCountConn) Close() error              { return nil }
func (*legacyFreshnessCountConn) Begin() (driver.Tx, error) { return legacyFreshnessCountTx{}, nil }
func (c *legacyFreshnessCountConn) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	if query == c.source.query {
		c.source.failed = true
		return nil, c.source.cause
	}
	if !strings.HasPrefix(query, "SELECT COUNT(*) FROM ") {
		return nil, fmt.Errorf("unexpected freshness query %q", query)
	}
	return &legacyFreshnessCountRows{}, nil
}

type legacyFreshnessCountTx struct{}

func (legacyFreshnessCountTx) Commit() error   { return nil }
func (legacyFreshnessCountTx) Rollback() error { return nil }

type legacyFreshnessCountRows struct{ read bool }

func (*legacyFreshnessCountRows) Columns() []string { return []string{"count"} }
func (*legacyFreshnessCountRows) Close() error      { return nil }
func (r *legacyFreshnessCountRows) Next(values []driver.Value) error {
	if r.read {
		return io.EOF
	}
	r.read = true
	values[0] = int64(0)
	return nil
}
