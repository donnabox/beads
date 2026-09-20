//go:build cgo

package graphops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	graph "github.com/steveyegge/beads/graphops"
)

// Shared embedded/managed controls qualify projection reads only. Their rows
// intentionally do not claim lawful ledger provenance or public authority.
func verifyEngineAllocationControls(t *testing.T, ctx context.Context, db fixtureDB) {
	t.Helper()
	before := fixtureTablesDigest(t, ctx, db)
	for _, kind := range []graph.ResourceKind{graph.KindBead, graph.KindLink} {
		for _, state := range []string{"", graph.AllocationLive, graph.AllocationReserved, graph.AllocationPruned, graph.AllocationErased} {
			for _, present := range []bool{false, true} {
				tx, err := db.BeginTx(ctx, nil)
				if err != nil {
					t.Fatal(err)
				}
				func() {
					defer func() { _ = tx.Rollback() }()
					path := string(kind) + "s/allocation-control"
					if present {
						allocationResource(t, ctx, tx, path, kind)
					}
					if state != "" {
						seedFixtureAllocation(t, ctx, tx, path, kind)
						if _, err := tx.ExecContext(ctx, "UPDATE graph_allocations SET state = ? WHERE path = ?", state, path); err != nil {
							t.Fatal(err)
						}
					}
					err = allocationRead(t, ctx, tx, path, kind)
					live := state == graph.AllocationLive && present
					gone := !present && state != graph.AllocationLive
					var absence *exactAbsence
					switch {
					case live:
						if err != nil {
							t.Fatalf("%s live: %v", kind, err)
						}
					case gone:
						if !errors.As(err, &absence) || absence.state != state {
							t.Fatalf("%s state=%s absence: %v", kind, state, err)
						}
					default:
						if !errors.Is(err, errCorrupt) || errors.As(err, &absence) {
							t.Fatalf("%s state=%s present=%v: %v", kind, state, present, err)
						}
					}
					if err := tx.Rollback(); err != nil {
						t.Fatal(err)
					}
				}()
				if got := fixtureTablesDigest(t, ctx, db); got != before {
					t.Fatal("allocation control retained mutation")
				}
			}
		}
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()
	// Distinct canonical paths must stay distinct under the actual join collation.
	for _, entry := range []struct {
		path string
		kind graph.ResourceKind
	}{{"beads/Plan", graph.KindBead}, {"links/Plan-decision", graph.KindLink}} {
		var absence *exactAbsence
		if err := allocationRead(t, ctx, tx, entry.path, entry.kind); !errors.As(err, &absence) || absence.state != "" {
			t.Fatalf("case-differing lookup: %v", err)
		}
	}

	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	verifyAllocationMissingDescriptor(t, ctx, db)
	if got := fixtureTablesDigest(t, ctx, db); got != before {
		t.Fatal("allocation read controls changed tables")
	}
	fmt.Println("GRAPH_ALLOCATION_PROJECTION_CONTROLS_OK")
}

// Plan diagnostics use the already-qualified embedded EXPLAIN leg. Managed
// classification stays independent of SQL-server prepared-EXPLAIN support.
func logAllocationPlans(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()
	for _, entry := range []struct{ query, path string }{{exactBeadQuery, "beads/plan"}, {exactBeadQuery, "beads/absent"}, {exactBeadQuery, "beads/Plan"}, {exactLinkQuery, "links/plan-decision"}, {exactLinkQuery, "links/absent"}, {exactLinkQuery, "links/Plan-decision"}} {
		args := []any{fixtureLimits.valueBytes, entry.path, entry.path, entry.path}
		if entry.query == exactLinkQuery {
			args = []any{fixtureLimits.valueBytes, fixtureLimits.valueBytes, entry.path, entry.path, entry.path}
		}
		rows, err := tx.QueryContext(ctx, "EXPLAIN FORMAT=tree "+entry.query, args...)
		if err != nil {
			t.Fatal(err)
		}
		columns, err := rows.Columns()
		if err != nil || len(columns) != 1 || columns[0] != "plan" {
			_ = rows.Close()
			t.Fatalf("allocation plan columns %v: %v", columns, err)
		}
		var lines []string
		for rows.Next() {
			var line string
			if err := rows.Scan(&line); err != nil {
				_ = rows.Close()
				t.Fatal(err)
			}
			if len(line) > 4096 || len(lines) >= 64 {
				_ = rows.Close()
				t.Fatal("allocation plan diagnostic bound")
			}
			lines = append(lines, line)
		}
		if err := errors.Join(rows.Err(), rows.Close()); err != nil {
			t.Fatal(err)
		}
		if len(lines) == 0 {
			t.Fatal("empty allocation query plan")
		}
		// Print the complete bounded plan before applying the qualification oracle.
		t.Logf("allocation exact plan %s:\n%s", entry.path, strings.Join(lines, "\n"))
		tables := []string{"graph_allocations", "graph_beads"}
		if entry.query == exactLinkQuery {
			tables = []string{"graph_allocations", "graph_links", "graph_type_descriptors"}
		}
		if err := checkAllocationLookupPlan(lines, tables, entry.path); err != nil {
			t.Error(err)
		}
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if !t.Failed() {
		fmt.Println("GRAPH_ALLOCATION_PLANS_RECORDED")
	}
}

func allocationResource(t *testing.T, ctx context.Context, tx *sql.Tx, path string, kind graph.ResourceKind) {
	t.Helper()
	revision := graph.MintRevision().String()
	var err error
	if kind == graph.KindBead {
		_, err = tx.ExecContext(ctx, "INSERT INTO graph_beads (path,type_url,revision,properties,last_authority_id,last_epoch,created_at,updated_at) VALUES (?,?,?, ?,REPEAT('a',32),1,'2026-09-17 00:00:00','2026-09-17 00:00:00')", path, memoryType, revision, []byte(`{}`))
	} else {
		_, err = tx.ExecContext(ctx, "INSERT INTO graph_links (path,type_url,revision,properties,source_kind,source_path,target_kind,target_path,last_authority_id,last_epoch,created_at,updated_at) VALUES (?,?,?,?,'in','beads/plan','in','beads/finding',REPEAT('a',32),1,'2026-09-17 00:00:00','2026-09-17 00:00:00')", path, relationType, revision, []byte(`{}`))
	}
	if err != nil {
		t.Fatal(err)
	}
}
func allocationRead(t *testing.T, ctx context.Context, tx *sql.Tx, path string, kind graph.ResourceKind) error {
	t.Helper()
	if kind == graph.KindBead {
		record, err := readBeadInTx(ctx, tx, fixtureScope, path, fixtureLimits)
		if err != nil && (!record.Bead.IsZero() || len(record.OwnedLinks) != 0) {
			t.Fatal("failed exact Bead retained fields")
		}
		if err == nil && record.Bead.Path() != path {
			t.Fatal("exact Bead identity mismatch")
		}
		return err
	}
	record, err := readLinkInTx(ctx, tx, fixtureScope, path, fixtureLimits)
	if err != nil && !record.IsZero() {
		t.Fatal("failed exact Link retained fields")
	}
	if err == nil && record.Path() != path {
		t.Fatal("exact Link identity mismatch")
	}
	return err
}

// This is the embedded two-session leg. The managed worker deliberately owns
// one connection; shared classification controls above also run over MySQL.
const allocationSnapshotIncidentPath = "links/incident-allocation-snapshot"

func verifyAllocationSnapshot(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	before := fixtureTablesDigest(t, ctx, db)
	const path = "beads/allocation-snapshot"
	setup, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = setup.Rollback() }()
	allocationResource(t, ctx, setup, path, graph.KindBead)
	seedFixtureAllocation(t, ctx, setup, path, graph.KindBead)
	expectedIncident := seedIncidentTargetLink(t, ctx, setup, allocationSnapshotIncidentPath, path)
	if err := setup.Commit(); err != nil {
		t.Fatal(err)
	}
	// Register before reserved sessions/transactions, whose defers run first.
	// This function-level defer also runs before the caller closes the engine.
	cleaned := false
	defer func() {
		if cleaned {
			return
		}
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := cleanupAllocationSnapshot(cleanupCtx, db, path, false); err != nil {
			t.Errorf("allocation snapshot failure-path cleanup: %v", err)
		} else {
			t.Log("allocation snapshot failure-path projection cleanup completed")
		}
	}()
	reader, err := db.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	writer, err := db.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer writer.Close()
	held, err := reader.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = held.Rollback() }()
	original, err := readBeadInTx(ctx, held, fixtureScope, path, fixtureLimits)
	if err != nil {
		t.Fatal(err)
	}
	checkHeldIncident := func() {
		for _, direction := range []graph.Direction{graph.DirectionIn, graph.DirectionBoth} {
			page, err := readIncidentPageInTx(ctx, held, fixtureScope, path, direction, pageWindow{limit: 1}, fixtureLimits)
			if err != nil {
				t.Fatal(err)
			}
			requireIncidentTargetPage(t, page, expectedIncident)
		}
	}
	checkHeldIncident()
	write, err := writer.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = write.Rollback() }()
	for _, query := range []string{"DELETE FROM graph_beads WHERE path = ?", "UPDATE graph_allocations SET state = 'pruned', tombstone_seq = 2 WHERE path = ?"} {
		result, err := write.ExecContext(ctx, query, path)
		if err != nil {
			t.Fatal(err)
		}
		if n, err := result.RowsAffected(); err != nil || n != 1 {
			t.Fatalf("allocation transition affected%d: %v", n, err)
		}
	}
	if err := write.Commit(); err != nil {
		t.Fatal(err)
	}
	still, err := readBeadInTx(ctx, held, fixtureScope, path, fixtureLimits)
	if err != nil || still.Bead.Revision() != original.Bead.Revision() || still.Bead.Properties().String() != original.Bead.Properties().String() {
		t.Fatalf("held allocation/resource pair changed: %v", err)
	}
	checkHeldIncident()
	if err := held.Rollback(); err != nil {
		t.Fatal(err)
	}
	fresh, err := reader.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = fresh.Rollback() }()
	var absence *exactAbsence
	if err := allocationRead(t, ctx, fresh, path, graph.KindBead); !errors.As(err, &absence) || absence.state != graph.AllocationPruned {
		t.Fatalf("fresh allocation/resource pair: %v", err)
	}
	for _, direction := range []graph.Direction{graph.DirectionIn, graph.DirectionOut, graph.DirectionBoth} {
		page, err := readIncidentPageInTx(ctx, fresh, fixtureScope, path, direction, pageWindow{limit: 1}, fixtureLimits)
		requireEmptyPage(t, page)
		if !errors.As(err, &absence) || absence.state != graph.AllocationPruned {
			t.Fatalf("fresh incident snapshot: %v", err)
		}
	}
	if err := fresh.Rollback(); err != nil {
		t.Fatal(err)
	}
	// Restore only this test's newly created projection, not a production path.
	// This is explicit cleanup after committed controls, not claimed rollback.
	if err := cleanupAllocationSnapshot(ctx, writer, path, true); err != nil {
		t.Fatal(err)
	}
	cleaned = true
	if err := errors.Join(reader.Close(), writer.Close()); err != nil {
		t.Fatal(err)
	}
	if got := fixtureTablesDigest(t, ctx, db); got != before {
		t.Fatal("allocation snapshot controls retained changes")
	}
	fmt.Println("GRAPH_ALLOCATION_SNAPSHOT_OK")
}

// Delete only rows at the two test-owned paths. Normal cleanup proves the exact
// post-transition population; failure cleanup tolerates partially finished setup.
func cleanupAllocationSnapshot(ctx context.Context, db fixtureDB, path string, completed bool) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	for _, deletion := range []struct {
		query, path string
		want        int64
	}{
		{"DELETE FROM graph_links WHERE path = ?", allocationSnapshotIncidentPath, 1},
		{"DELETE FROM graph_allocations WHERE path = ?", allocationSnapshotIncidentPath, 1},
		{"DELETE FROM graph_allocations WHERE path = ?", path, 1},
		{"DELETE FROM graph_beads WHERE path = ?", path, 0},
	} {
		result, err := tx.ExecContext(ctx, deletion.query, deletion.path)
		if err != nil {
			return err
		}
		count, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if count < 0 || count > 1 || (completed && count != deletion.want) {
			return fmt.Errorf("snapshot cleanup affected unexpected row count: %d", count)
		}
	}
	return tx.Commit()
}

// The raw outer join keeps the Link, while the domain reader rejects the
// incomplete descriptor. This mutation is rolled back before the digest check.
func verifyAllocationMissingDescriptor(t *testing.T, ctx context.Context, db fixtureDB) {
	t.Helper()
	before := fixtureTablesDigest(t, ctx, db)
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()
	result, err := tx.ExecContext(ctx, "DELETE FROM graph_type_descriptors WHERE url = ?", relationType)
	if err != nil {
		t.Fatal(err)
	}
	if n, err := result.RowsAffected(); err != nil || n != 1 {
		t.Fatalf("descriptor delete affected%d: %v", n, err)
	}
	const path = "links/plan-decision"
	results, err := readRows(ctx, tx, exactLinkQuery, []any{fixtureLimits.valueBytes, fixtureLimits.valueBytes, path, path, path}, 1, func(rows *sql.Rows) (bool, error) {
		var allocation exactPathRow
		var row linkRow
		var id, fingerprint sql.NullString
		var raw []byte
		var length sql.NullInt64
		targets := append(append(allocation.targets(), linkScanTargets(&row)...), &id, &raw, &length, &fingerprint)
		if err := rows.Scan(targets...); err != nil {
			return false, err
		}
		return allocation.present && row.path == path && !id.Valid && raw == nil && !length.Valid && !fingerprint.Valid, nil
	})
	if err != nil || len(results) != 1 || !results[0] {
		t.Fatalf("missing descriptor did not preserve raw Link projection: %v", err)
	}
	link, err := readLinkInTx(ctx, tx, fixtureScope, path, fixtureLimits)
	var absence *exactAbsence
	if !link.IsZero() || !errors.Is(err, errCorrupt) || errors.As(err, &absence) {
		t.Fatalf("missing descriptor domain refusal: %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if got := fixtureTablesDigest(t, ctx, db); got != before {
		t.Fatal("missing descriptor control retained changes")
	}
}

// A dedicated external-source edge makes absent-target suppression reachable
// without disabling the source FK. It is never a production seeder.
func seedIncidentTargetLink(t *testing.T, ctx context.Context, tx *sql.Tx, path, targetPath string) graph.Link {
	t.Helper()
	source, err := graph.ParseRef(fixtureScope, "urn:incident-allocation:external", "source opaque")
	if err != nil {
		t.Fatal(err)
	}
	target, err := graph.NewInScopeRef(targetPath, "target opaque")
	if err != nil {
		t.Fatal(err)
	}
	properties, err := graph.NewProperties([]byte(`{"incident":true}`))
	if err != nil {
		t.Fatal(err)
	}
	link, err := graph.NewLink(graph.LinkSpec{Path: path, TypeURL: relationType, Revision: graph.MintRevision(), Properties: properties, Source: source, Target: target})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO graph_links (path,type_url,revision,properties,source_kind,source_url,source_pin,target_kind,target_path,target_pin,last_authority_id,last_epoch,created_at,updated_at) VALUES (?,?,?,?,'ext',?,?,'in',?,?,REPEAT('a',32),1,'2026-09-20 00:00:00','2026-09-20 00:00:00')", path, relationType, link.Revision().String(), properties.Bytes(), source.URI(), source.Pin(), target.Path(), target.Pin()); err != nil {
		t.Fatal(err)
	}
	seedFixtureAllocation(t, ctx, tx, path, graph.KindLink)
	return link
}
func requireIncidentTargetPage(t *testing.T, page linkRowsPage, want graph.Link) {
	t.Helper()
	if len(page.items) != 1 || page.hasMore || page.lastPath != want.Path() {
		t.Fatalf("incident page=%+v", page)
	}
	got := page.items[0]
	attribution, present := got.Attribution()
	wantAttribution, wantPresent := want.Attribution()
	if attribution != wantAttribution || present != wantPresent {
		t.Fatal("incident snapshot attribution changed")
	}
	if got.Path() != want.Path() || got.TypeURL() != want.TypeURL() || got.Revision() != want.Revision() || got.Properties().String() != want.Properties().String() || !got.Source().Equal(want.Source()) || !got.Target().Equal(want.Target()) {
		t.Fatal("incident snapshot Link changed")
	}
}
