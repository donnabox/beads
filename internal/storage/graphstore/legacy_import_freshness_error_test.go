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

	"github.com/steveyegge/beads/internal/storage"
)

// This tests error wrapping at the SQL boundary without opening an engine.
// The double only supplies COUNT results or the selected query failure;
// it does not model tables, mutations or transaction behavior.
func TestLegacyImportFreshnessQueryErrors(t *testing.T) {
	for _, test := range []struct {
		name, query, plane string
	}{
		{"table", "SELECT COUNT(*) FROM graph_preview_catalog", "graph_preview_catalog"},
		{"journal-sequence", "SELECT COUNT(*) FROM bd_events_seq WHERE id<>0 OR next_seq<>0", "bd_events_seq"},
		{"epoch-address", "SELECT COUNT(*) FROM epoch_minted_addresses", "epoch_minted_addresses"},
		{"interaction", "SELECT COUNT(*) FROM interactions", "interactions"},
		{"epoch-state", legacyFreshnessChangedEpochQuery, "store_epoch"},
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

const legacyFreshnessChangedEpochQuery = "SELECT COUNT(*) FROM store_epoch WHERE id<>1 OR epoch<>1 OR bumped_at IS NOT NULL OR bumped_reason IS NOT NULL"

// These are prefiltered SQL COUNT results. The engine matrix separately tests
// the actual absent/pristine singleton and each bumped-field predicate.
func TestLegacyImportFreshnessPlaneAdmission(t *testing.T) {
	for _, test := range []struct {
		name, occupiedQuery, plane string
	}{
		{"absent-or-pristine", "", ""},
		{"minted-address", "SELECT COUNT(*) FROM epoch_minted_addresses", "epoch_minted_addresses"},
		{"interaction", "SELECT COUNT(*) FROM interactions", "interactions"},
		{"changed-epoch", legacyFreshnessChangedEpochQuery, "store_epoch"},
	} {
		t.Run(test.name, func(t *testing.T) {
			connector := &legacyFreshnessCountDriver{counts: map[string]int64{}, seen: map[string]bool{}}
			if test.occupiedQuery != "" {
				connector.counts[test.occupiedQuery] = 1
			}
			db := sql.OpenDB(connector)
			defer func() {
				if err := db.Close(); err != nil {
					t.Error(err)
				}
			}()
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
			if test.occupiedQuery == "" {
				if err != nil || !connector.seen["SELECT COUNT(*) FROM epoch_minted_addresses"] || !connector.seen[legacyFreshnessChangedEpochQuery] {
					t.Fatalf("fresh epoch baseline refused or unchecked: %v", err)
				}
			} else if !connector.seen[test.occupiedQuery] || !errors.Is(err, storage.ErrValidation) || !strings.Contains(err.Error(), "; "+test.plane+" is occupied") {
				t.Fatalf("occupied epoch plane admitted or misclassified: %v", err)
			}
		})
	}
}

type legacyFreshnessCountDriver struct {
	query  string
	cause  error
	failed bool
	counts map[string]int64
	seen   map[string]bool
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
	if c.source.seen != nil {
		c.source.seen[query] = true
	}
	if query == c.source.query {
		c.source.failed = true
		return nil, c.source.cause
	}
	if !strings.HasPrefix(query, "SELECT COUNT(*) FROM ") {
		return nil, fmt.Errorf("unexpected freshness query %q", query)
	}
	return &legacyFreshnessCountRows{count: c.source.counts[query]}, nil
}

type legacyFreshnessCountTx struct{}

func (legacyFreshnessCountTx) Commit() error   { return nil }
func (legacyFreshnessCountTx) Rollback() error { return nil }

type legacyFreshnessCountRows struct {
	read  bool
	count int64
}

func (*legacyFreshnessCountRows) Columns() []string { return []string{"count"} }
func (*legacyFreshnessCountRows) Close() error      { return nil }
func (r *legacyFreshnessCountRows) Next(values []driver.Value) error {
	if r.read {
		return io.EOF
	}
	r.read = true
	values[0] = r.count
	return nil
}
