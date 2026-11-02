//go:build cgo

package graphstore

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	mysql "github.com/go-sql-driver/mysql"
	"github.com/steveyegge/beads/internal/intrinsicmaterialization"
	"github.com/steveyegge/beads/internal/intrinsictypecatalog"
	"github.com/steveyegge/beads/resourceget"
)

func intrinsicHardeningStore(t *testing.T) (context.Context, Options, *Store) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	t.Cleanup(cancel)
	o := testOptions(t)
	o.Binding.SchemaVersion = TypesFreshSchemaVersion
	if err := Init(ctx, o); err != nil {
		t.Fatal(err)
	}
	s, err := OpenExisting(ctx, o)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := s.Close(); err != nil {
			t.Error(err)
		}
	})
	return ctx, o, s
}

// The shipping test observer reads one complete current context in one native
// transaction. The separate oracle below retains six independent public Gets.
func intrinsicHardeningState(t *testing.T, ctx context.Context, s *Store) []string {
	t.Helper()
	state, err := readIntrinsicHardeningState(ctx, s)
	if err != nil {
		t.Fatal(err)
	}
	return state
}

func readIntrinsicHardeningState(ctx context.Context, s *Store) ([]string, error) {
	var state []string
	err := s.withDefinitionReadTx(ctx, txDefinitionRead, func(tx *sql.Tx, budget *intrinsicMetadataBudget) error {
		var err error
		state, err = captureIntrinsicHardeningStateInTx(ctx, tx, s, budget)
		return err
	})
	if err != nil {
		return nil, err
	}
	return state, nil
}

func captureIntrinsicHardeningStateInTx(ctx context.Context, tx *sql.Tx, s *Store, budget *intrinsicMetadataBudget) ([]string, error) {
	// Token, fixed headers/order map and final count strings are bounded here.
	if err := budget.chargeBody(6 * 4096); err != nil {
		return nil, err
	}
	var token string
	if err := tx.QueryRowContext(ctx, `SELECT writer_token FROM graph_preview_scope WHERE singleton=1`).Scan(&token); err != nil {
		return nil, err
	}
	contextID, err := currentIntrinsicContextInTx(ctx, tx, budget)
	if err != nil {
		return nil, err
	}
	plan, err := loadIntrinsicContextInTx(ctx, tx, s.options, contextID, budget)
	if err != nil {
		return nil, err
	}
	if err = budget.chargeBody(2 * intrinsicmaterialization.MaxContextBytes); err != nil {
		return nil, err
	}
	bindings := plan.Bindings()
	if len(bindings) != 6 {
		return nil, ErrInvalidStore
	}
	byID := make(map[string]intrinsicmaterialization.Binding, 6)
	for _, binding := range bindings {
		id := binding.Allocation.PublicID
		if _, exists := byID[id]; exists {
			return nil, ErrInvalidStore
		}
		byID[id] = binding
	}
	state := []string{token}
	for index, id := range []string{"type", "bead-type", "link-type", "refines", "bead-type-refines-type", "link-type-refines-type"} {
		binding, ok := byID[id]
		if !ok || validatePublicID(id) != nil {
			return nil, ErrInvalidStore
		}
		allocation := binding.Allocation
		wantKind := "bead"
		if index >= 4 {
			wantKind = "link"
		}
		if string(allocation.Kind) != wantKind {
			return nil, ErrInvalidStore
		}
		// The full context checks private-path correspondence. Retain the public
		// resolver too, so another backing with this ID cannot become invisible.
		identity, err := resolveAllocatedResourceIdentityInTx(ctx, tx, id, budget)
		if err != nil {
			return nil, err
		}
		if identity.PublicID != id || identity.PrivatePath != allocation.PrivatePath || identity.Kind != allocation.Kind || identity.State != "live" {
			return nil, ErrInvalidStore
		}
		// Reserve before constructor canonicalization/clones and output strings.
		if err = budget.chargeBody(8 * len(binding.Record)); err != nil {
			return nil, err
		}
		doc, err := definitionDocumentForNative(s.ScopeURL(), IntrinsicDocument{
			PublicID: id, PrivatePath: allocation.PrivatePath, Kind: allocation.Kind,
			Revision: allocation.Local.Revision, ContextID: contextID,
			Global: binding.Global, Body: binding.Record,
		})
		if err != nil {
			return nil, err
		}
		if doc.ID() != allocation.Local.URI || doc.Revision() != allocation.Local.Revision || string(doc.Kind()) != wantKind {
			return nil, ErrInvalidStore
		}
		state = append(state, string(doc.NativeJSON()))
	}
	for _, query := range []string{`SELECT COUNT(*) FROM graph_preview_intrinsic_contexts`, `SELECT COUNT(*) FROM graph_preview_intrinsic_members`} {
		var n int
		if err := tx.QueryRowContext(ctx, query).Scan(&n); err != nil {
			return nil, err
		}
		state = append(state, fmt.Sprint(n))
	}
	return state, nil
}

// Capture exactly the protected intrinsic records and fence, not an entire
// database dump. Readbacks use production readers, including retained captures.
func intrinsicHardeningStateOracle(t *testing.T, ctx context.Context, s *Store) []string {
	t.Helper()
	var token string
	if err := s.db.QueryRowContext(ctx, `SELECT writer_token FROM graph_preview_scope WHERE singleton=1`).Scan(&token); err != nil {
		t.Fatal(err)
	}
	state := []string{token}
	for _, id := range []string{"type", "bead-type", "link-type", "refines", "bead-type-refines-type", "link-type-refines-type"} {
		doc, err := s.GetDefinition(ctx, id, "")
		if err != nil {
			t.Fatal(err)
		}
		state = append(state, string(doc.NativeJSON()))
	}
	for _, query := range []string{`SELECT COUNT(*) FROM graph_preview_intrinsic_contexts`, `SELECT COUNT(*) FROM graph_preview_intrinsic_members`} {
		var n int
		if err := s.db.QueryRowContext(ctx, query).Scan(&n); err != nil {
			t.Fatal(err)
		}
		state = append(state, fmt.Sprint(n))
	}
	return state
}

func assertCustomDefinitionRows(t *testing.T, ctx context.Context, s *Store, id string, want int) {
	t.Helper()
	for _, query := range []string{
		`SELECT COUNT(*) FROM graph_preview_catalog WHERE public_id=?`,
		`SELECT COUNT(*) FROM graph_preview_payloads p JOIN graph_preview_catalog c ON c.path=p.path WHERE c.public_id=?`,
		`SELECT COUNT(*) FROM graph_preview_versions v JOIN graph_preview_catalog c ON c.path=v.path WHERE c.public_id=?`,
	} {
		var n int
		if err := s.db.QueryRowContext(ctx, query, id).Scan(&n); err != nil || n != want {
			t.Fatalf("partial definition: count=%d want=%d err=%v", n, want, err)
		}
	}
}

func TestCustomDefinitionAtomicRollbackAndReservations(t *testing.T) {
	ctx, o, s := intrinsicHardeningStore(t)
	fault := errors.New("injected definition stage failure")
	for _, stage := range []string{"definition-validated", "coordination", "definition-catalog", "definition-payload", "definition-version", "definition-readback"} {
		t.Run(stage, func(t *testing.T) {
			before := intrinsicHardeningState(t, ctx, s)
			var catalogs, payloads, versions int
			if err := s.db.QueryRowContext(ctx, `SELECT (SELECT COUNT(*) FROM graph_preview_catalog),(SELECT COUNT(*) FROM graph_preview_payloads),(SELECT COUNT(*) FROM graph_preview_versions)`).Scan(&catalogs, &payloads, &versions); err != nil {
				t.Fatal(err)
			}
			calls := 0
			s.afterWrite = func(actual string) error {
				if actual == stage {
					calls++
					return fault
				}
				return nil
			}
			id := "links/" + stage
			doc, err := s.CreateBeadType(ctx, id, "hardening")
			s.afterWrite = nil
			if !errors.Is(err, fault) || errors.Is(err, ErrOutcomeUnknown) || doc.ID() != "" || calls != 1 {
				t.Fatalf("failure leaked result/replayed: %v %s calls%d", err, doc.ID(), calls)
			}
			assertCustomDefinitionRows(t, ctx, s, id, 0)
			var c, p, v int
			if err := s.db.QueryRowContext(ctx, `SELECT (SELECT COUNT(*) FROM graph_preview_catalog),(SELECT COUNT(*) FROM graph_preview_payloads),(SELECT COUNT(*) FROM graph_preview_versions)`).Scan(&c, &p, &v); err != nil || c != catalogs || p != payloads || v != versions {
				t.Fatalf("rollback left orphan rows: %d/%d/%d: %v", c, p, v, err)
			}
			if after := intrinsicHardeningState(t, ctx, s); !reflect.DeepEqual(before, after) {
				t.Fatal("rollback changed intrinsic records/context/fence")
			}
			doc, err = s.CreateBeadType(ctx, id, "hardening")
			if err != nil || doc.ID() != s.ScopeURL()+id {
				t.Fatalf("same ID not reusable after confirmed rollback: %s %v", doc.ID(), err)
			}
			assertCustomDefinitionRows(t, ctx, s, id, 1)
		})
	}
	before := intrinsicHardeningState(t, ctx, s)
	if doc, err := s.CreateBeadType(ctx, "bead-type-refines-type", "hardening"); (!errors.Is(err, ErrAlreadyExists) && !errors.Is(err, ErrConflict)) || doc.ID() != "" {
		t.Fatalf("cross-kind lifetime reservation escaped: %v", err)
	}
	if after := intrinsicHardeningState(t, ctx, s); !reflect.DeepEqual(before, after) {
		t.Fatal("cross-kind refusal changed owned Link/fence/context")
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	if err := Init(ctx, o); err == nil {
		t.Fatal("fresh initializer adopted an existing database")
	}
	reopened, err := OpenExisting(ctx, o)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := reopened.Close(); err != nil {
			t.Error(err)
		}
	}()
	if after := intrinsicHardeningState(t, ctx, reopened); !reflect.DeepEqual(before, after) {
		t.Fatal("failed repeated Init changed existing records")
	}
}

func TestCustomDefinitionConcurrentSameID(t *testing.T) {
	ctx, o, left := intrinsicHardeningStore(t)
	// Distinct database/sql connections can race against the same embedded
	// engine. Both callbacks rendezvous after their complete checked read
	// snapshots, before either attempts the shared writer-fence mutation.
	left.db.SetMaxOpenConns(2)
	right := &Store{db: left.db, options: o}
	before := intrinsicHardeningState(t, ctx, left)
	barrierCtx, stopBarrier := context.WithTimeout(ctx, 10*time.Second)
	defer stopBarrier()
	start := make(chan struct{})
	reached := make(chan struct{}, 2)
	release := make(chan struct{})
	var releaseOnce sync.Once
	releaseWriters := func() { releaseOnce.Do(func() { close(release) }) }
	pause := func(stage string) error {
		if stage != "definition-validated" {
			return nil
		}
		select {
		case reached <- struct{}{}:
		case <-barrierCtx.Done():
			return barrierCtx.Err()
		}
		select {
		case <-release:
			return nil
		case <-barrierCtx.Done():
			return barrierCtx.Err()
		}
	}
	var traceMu sync.Mutex
	traces := make(map[string][]string)
	tracePause := func(writer string) func(string) error {
		return func(stage string) error {
			traceMu.Lock()
			traces[writer] = append(traces[writer], stage)
			traceMu.Unlock()
			return pause(stage)
		}
	}
	left.afterWrite, right.afterWrite = tracePause("left"), tracePause("right")
	type result struct {
		writer string
		doc    resourceget.DefinitionDocument
		err    error
	}
	results := make(chan result, 2)
	var workers sync.WaitGroup
	t.Cleanup(func() {
		releaseWriters()
		workers.Wait()
		left.afterWrite, right.afterWrite = nil, nil
	})
	for i, s := range []*Store{left, right} {
		writer := []string{"left", "right"}[i]
		workers.Add(1)
		go func(s *Store, writer string) {
			defer workers.Done()
			<-start
			d, e := s.CreateBeadType(ctx, "Race/Foo", "hardening")
			results <- result{writer, d, e}
		}(s, writer)
	}
	close(start)
	arrivals := 0
	for arrivals < 2 {
		select {
		case <-reached:
			arrivals++
		case <-barrierCtx.Done():
			releaseWriters()
			workers.Wait()
			t.Fatalf("only %d validated snapshots reached rendezvous: %v", arrivals, barrierCtx.Err())
		}
	}
	releaseWriters()
	workers.Wait()
	left.afterWrite, right.afterWrite = nil, nil
	close(results)
	var winner resourceget.DefinitionDocument
	success, refused := 0, 0
	for r := range results {
		stages := traces[r.writer] // all writers joined before these reads
		var typed *mysql.MySQLError
		hasTyped := errors.As(r.err, &typed)
		var number uint16
		var state [5]byte
		if hasTyped {
			number, state = typed.Number, typed.SQLState
		}
		t.Logf("ACTUAL_WRITER writer=%s stages=%q errorType=%T mysql=%t number=%d sqlState=%v err=%v", r.writer, stages, r.err, hasTyped, number, state, r.err)
		// This is the callback's final fallible step; the hook returns nil here.
		// With validateTypes=false, only the real Commit remains afterward.
		if !slices.Contains(stages, "definition-readback") {
			t.Fatalf("writer %s did not establish commit-phase witness: %v", r.writer, r.err)
		}
		if r.err == nil {
			success++
			winner = r.doc
		} else if (errors.Is(r.err, ErrConflict) || errors.Is(r.err, ErrAlreadyExists)) && !errors.Is(r.err, ErrOutcomeUnknown) && r.doc.ID() == "" {
			refused++
		} else {
			t.Fatalf("unclassified racing writer: %v", r.err)
		}
	}
	if success != 1 || refused != 1 {
		t.Fatalf("writers success%d refused%d", success, refused)
	}
	assertCustomDefinitionRows(t, ctx, left, "Race/Foo", 1)
	if after := intrinsicHardeningState(t, ctx, left); !reflect.DeepEqual(before[1:], after[1:]) {
		t.Fatal("racing writers changed protected intrinsic records/context")
	}
	actual, err := left.GetDefinition(ctx, "Race/Foo", winner.Revision())
	if err != nil || !bytes.Equal(actual.NativeJSON(), winner.NativeJSON()) {
		t.Fatalf("winner differs from retained commit: %v", err)
	}
	lower, err := left.CreateBeadType(ctx, "race/Foo", "hardening")
	if err != nil || lower.ID() == winner.ID() {
		t.Fatalf("case-folded literal IDs: %v", err)
	}
}

func TestCustomDefinitionPersistedCorruption(t *testing.T) {
	ctx, _, s := intrinsicHardeningStore(t)
	healthy, err := s.CreateBeadType(ctx, "foo", "hardening")
	if err != nil {
		t.Fatal(err)
	}
	var path string
	if err := s.db.QueryRowContext(ctx, `SELECT path FROM graph_preview_catalog WHERE public_id=?`, "foo").Scan(&path); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name, query string
		args        []any
	}{
		{"catalog-type", `UPDATE graph_preview_catalog SET type_url=? WHERE path=?`, []any{"https://wrong.invalid/type", path}},
		{"allocation-key", `UPDATE graph_preview_catalog SET backing_key=? WHERE path=?`, []any{"wrong", path}},
		{"missing-payload", `DELETE FROM graph_preview_payloads WHERE path=?`, []any{path}},
		{"private-context", `UPDATE graph_preview_payloads SET metadata=? WHERE path=?`, []any{[]byte(`{"encoding":"unsupported"}`), path}},
		{"current-body", `UPDATE graph_preview_payloads SET properties=? WHERE path=?`, []any{[]byte(`{}`), path}},
		{"missing-snapshot", `DELETE FROM graph_preview_versions WHERE path=?`, []any{path}},
		{"snapshot-bytes", `UPDATE graph_preview_versions SET snapshot=? WHERE path=?`, []any{[]byte(`{}`), path}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := s.withDefinitionReadTx(ctx, txDefinitionRead, func(tx *sql.Tx, budget *intrinsicMetadataBudget) error {
				if _, err := tx.ExecContext(ctx, tc.query, tc.args...); err != nil {
					return err
				}
				id, err := currentIntrinsicContextInTx(ctx, tx, budget)
				if err != nil {
					return err
				}
				plan, err := loadIntrinsicContextInTx(ctx, tx, s.options, id, budget)
				if err != nil {
					return err
				}
				doc, err := readCustomDefinitionInTx(ctx, tx, s.options, path, healthy.Revision(), plan, budget)
				if !errors.Is(err, ErrInvalidStore) || doc.ID() != "" {
					return fmt.Errorf("corrupt read escaped: %s %v", doc.ID(), err)
				}
				docs, err := s.listCustomDefinitionsInTx(ctx, tx, plan, budget)
				if !errors.Is(err, ErrInvalidStore) || docs != nil {
					return fmt.Errorf("partial corrupt collection: %d %v", len(docs), err)
				}
				return nil // wrapper rolls back: no committed fixture corruption
			})
			if err != nil {
				t.Fatal(err)
			}
			doc, err := s.GetDefinition(ctx, "foo", healthy.Revision())
			if err != nil || !bytes.Equal(doc.NativeJSON(), healthy.NativeJSON()) {
				t.Fatalf("corruption rollback changed record: %v", err)
			}
		})
	}
	t.Run("public-list-discards-partial-intrinsics", func(t *testing.T) {
		// Commit corruption only inside this exclusively owned disposable fixture,
		// so the public operation's own transaction observes it. This proves the
		// outward List result is nil even after its intrinsic records were selected.
		bad, err := s.CreateBeadType(ctx, "list-corrupt", "hardening")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := s.db.ExecContext(ctx, `UPDATE graph_preview_versions SET snapshot=? WHERE version=?`, []byte(`{}`), bad.Revision()); err != nil {
			t.Fatal(err)
		}
		docs, err := s.ListDefinitions(ctx, "type", "")
		if !errors.Is(err, ErrInvalidStore) || docs != nil {
			t.Fatalf("public partial collection: %d %v", len(docs), err)
		}
	})

}

// Exercise the production materializer in a real caller-owned transaction.
// DDL is deliberately outside this claim: a failed fresh Init may stay pending.
func TestIntrinsicMaterializationFinalCensusRollback(t *testing.T) {
	ctx, o, s := intrinsicHardeningStore(t)
	before := intrinsicHardeningState(t, ctx, s)
	// The complete initialized population now includes ordinary Type Beads and
	// their Refines Links; only the immutable intrinsic-member context stays six.
	// Retain every column/cell/row in each table that bootstrap can mutate,
	// independently of the protected-intrinsic observer above.
	tables := []struct{ name, order string }{
		{"graph_preview_scope", "singleton"},
		{"graph_preview_catalog", "path"},
		{"graph_preview_payloads", "path"},
		{"graph_preview_versions", "path,version"},
		{"graph_preview_intrinsic_contexts", "context_id"},
		{"graph_preview_intrinsic_members", "path,revision"},
	}
	committedRows := func() map[string][]string {
		t.Helper()
		state := make(map[string][]string, len(tables))
		retained := 0
		for _, table := range tables {
			rows, err := s.db.QueryContext(ctx, "SELECT * FROM "+table.name+" ORDER BY "+table.order)
			if err != nil {
				t.Fatal(err)
			}
			columns, err := rows.Columns()
			if err != nil {
				_ = rows.Close()
				t.Fatal(err)
			}
			appendRow := func(value any) {
				t.Helper()
				encoded, err := json.Marshal(value)
				if err != nil || len(encoded) > PreviewCurrentReadByteLimit-retained {
					_ = rows.Close()
					t.Fatalf("bounded complete bootstrap row oracle: %v", err)
				}
				retained += len(encoded)
				state[table.name] = append(state[table.name], string(encoded))
			}
			appendRow(columns)
			for rows.Next() {
				values := make([]any, len(columns))
				pointers := make([]any, len(columns))
				for i := range values {
					pointers[i] = &values[i]
				}
				if err = rows.Scan(pointers...); err != nil {
					break
				}
				appendRow(values)
			}
			if err = errors.Join(err, rows.Err(), rows.Close()); err != nil {
				t.Fatal(err)
			}
		}
		return state
	}
	baseline := committedRows()
	if len(baseline["graph_preview_intrinsic_members"])-1 != 6 || len(baseline["graph_preview_intrinsic_contexts"])-1 != 1 || len(baseline["graph_preview_catalog"])-1 <= 6 || len(baseline["graph_preview_payloads"])-1 == 0 {
		t.Fatal("expected six intrinsic members plus complete ordinary Type bootstrap population")
	}
	clearStaged := func(tx *sql.Tx) {
		t.Helper()
		// Payloads are ordinary-Type backing state too. These deletes are staged
		// only in the caller-owned transaction, never reset/commit the fixture.
		for _, q := range []string{`DELETE FROM graph_preview_payloads`, `DELETE FROM graph_preview_versions`, `DELETE FROM graph_preview_intrinsic_members`, `DELETE FROM graph_preview_intrinsic_contexts`, `DELETE FROM graph_preview_catalog`} {
			if _, err := tx.ExecContext(ctx, q); err != nil {
				_ = tx.Rollback()
				t.Fatal(err)
			}
		}
	}
	assertStagedPopulation := func(tx *sql.Tx) {
		t.Helper()
		for _, table := range tables[1:] {
			var n int
			want := len(baseline[table.name]) - 1 // first entry is the full column list
			if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table.name).Scan(&n); err != nil || n != want {
				_ = tx.Rollback()
				t.Fatalf("incomplete staged bootstrap %s: %d want %d: %v", table.name, n, want, err)
			}
		}
	}
	assertRestored := func(label string) {
		t.Helper()
		if after := committedRows(); !reflect.DeepEqual(baseline, after) {
			t.Fatalf("%s changed committed bootstrap table cells/rows/fence", label)
		}
		if after := intrinsicHardeningState(t, ctx, s); !reflect.DeepEqual(before, after) {
			t.Fatal("failed materialization changed committed intrinsic state")
		}
	}
	// Calibrate using the actual successful production materializer and an
	// independent exact final-context acquisition. Both use the normal cap,
	// inside this SAME original90 fixture; the caller rolls calibration back.
	// Fixed-length native tokens keep equivalent fresh bootstrap debit sizes.
	preCensus := func() int {
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := tx.Rollback(); err != nil {
				t.Fatal(err)
			}
		}()
		clearStaged(tx)
		full := newIntrinsicMetadataBudget()
		if err := installIntrinsicMembers(ctx, tx, o, full); err != nil {
			t.Fatalf("normal-cap materialization calibration: %v", err)
		}
		assertStagedPopulation(tx)
		var contextID string
		if err := tx.QueryRowContext(ctx, `SELECT context_id FROM graph_preview_intrinsic_contexts`).Scan(&contextID); err != nil {
			t.Fatal(err)
		}
		census := newIntrinsicMetadataBudget()
		if _, err := loadIntrinsicContextInTx(ctx, tx, o, contextID, census); err != nil {
			t.Fatalf("independent final-context calibration: %v", err)
		}
		spent := PreviewCurrentReadByteLimit - full.remaining
		finalDebit := PreviewCurrentReadByteLimit - census.remaining
		beforeFinal := spent - finalDebit
		initial := 4*intrinsictypecatalog.CurrentSource().Bytes + 2*intrinsicmaterialization.MaxContextBytes
		if finalDebit <= 0 || beforeFinal <= initial+4*intrinsicmaterialization.MaxContextBytes || beforeFinal >= PreviewCurrentReadByteLimit {
			t.Fatalf("invalid bounded final-census calibration: total=%d final=%d before=%d", spent, finalDebit, beforeFinal)
		}
		return beforeFinal
	}()
	assertRestored("successful calibration rollback")
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	clearStaged(tx)
	// Exactly the measured source/planning/write/readback debit, including ALL
	// installed Types. No allowance remains for the real final integrity census.
	budget := &intrinsicMetadataBudget{remaining: preCensus}
	err = installIntrinsicMembers(ctx, tx, o, budget)
	if !errors.Is(err, ErrInvalidStore) || !strings.Contains(err.Error(), ErrLimitExceeded.Error()) || !strings.Contains(err.Error(), "final intrinsic census:") {
		_ = tx.Rollback()
		t.Fatalf("final census did not refuse: %v", err)
	}
	assertStagedPopulation(tx)
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	assertRestored("final-census refusal rollback")
}

// Independent public-API oracle comparisons, not the observer compared to itself.
func TestIntrinsicHardeningStateObserverEquivalence(t *testing.T) {
	ctx, _, s := intrinsicHardeningStore(t)
	assertEquivalent := func(label string) {
		t.Helper()
		want := intrinsicHardeningStateOracle(t, ctx, s)
		got, err := readIntrinsicHardeningState(ctx, s)
		if err != nil || !reflect.DeepEqual(got, want) || len(got) != 9 {
			t.Fatalf("%s complete protected state differs: %v", label, err)
		}
		for i, raw := range got[1:7] {
			kind := resourceget.DefinitionBead
			if i >= 4 {
				kind = resourceget.DefinitionLink
			}
			if _, err := resourceget.NewDefinitionDocument(s.ScopeURL(), kind, []byte(raw)); err != nil {
				t.Fatalf("%s full record %d: %v", label, i, err)
			}
		}
	}
	assertEquivalent("before-custom-write")
	if _, err := s.CreateBeadType(ctx, "observer-foo", "hardening"); err != nil {
		t.Fatal(err)
	}
	assertEquivalent("after-custom-write")
}

func TestIntrinsicHardeningStateObserverRefusals(t *testing.T) {
	for _, mode := range []string{"retained-member", "retained-snapshot", "public-id-ambiguity", "owned-capture"} {
		t.Run(mode, func(t *testing.T) {
			// Independent finite fresh fixtures, not a new cumulative fault-root
			// sequence or a reset of the original shared90 context.
			ctx, _, s := intrinsicHardeningStore(t)
			before := intrinsicHardeningStateOracle(t, ctx, s)
			var path string
			if err := s.db.QueryRowContext(ctx, `SELECT path FROM graph_preview_catalog WHERE public_id=?`, "bead-type").Scan(&path); err != nil {
				t.Fatal(err)
			}
			err := s.withDefinitionReadTx(ctx, txDefinitionRead, func(tx *sql.Tx, budget *intrinsicMetadataBudget) error {
				switch mode {
				case "retained-member", "owned-capture":
					raw := []byte(`{}`)
					if mode == "owned-capture" {
						var current []byte
						if err := tx.QueryRowContext(ctx, `SELECT record FROM graph_preview_intrinsic_members WHERE path=?`, path).Scan(&current); err != nil {
							return err
						}
						var fields map[string]json.RawMessage
						if err := json.Unmarshal(current, &fields); err != nil {
							return err
						}
						fields["ownedLinks"] = json.RawMessage(`{}`)
						var err error
						raw, err = json.Marshal(fields)
						if err != nil {
							return err
						}
					}
					// Keep digest coherent: the source/body correspondence must refuse.
					if _, err := tx.ExecContext(ctx, `UPDATE graph_preview_intrinsic_members SET record=?,record_sha256=? WHERE path=?`, raw, intrinsicDigest(raw), path); err != nil {
						return err
					}
				case "retained-snapshot":
					if _, err := tx.ExecContext(ctx, `UPDATE graph_preview_versions SET snapshot=? WHERE path=?`, []byte(`{}`), path); err != nil {
						return err
					}
				case "public-id-ambiguity":
					key := strings.Repeat("f", 32)
					if _, err := tx.ExecContext(ctx, `INSERT INTO graph_preview_catalog (path,resource_kind,type_url,revision,allocation_state,backing,backing_key,public_id) SELECT ?,resource_kind,type_url,revision,allocation_state,'generic',?,public_id FROM graph_preview_catalog WHERE path=?`, "beads/"+key, key, path); err != nil {
						return err
					}
				}
				got, err := captureIntrinsicHardeningStateInTx(ctx, tx, s, budget)
				if !errors.Is(err, ErrInvalidStore) || got != nil {
					return fmt.Errorf("%s corrupt observer escaped: %d records, %w", mode, len(got), err)
				}
				return nil // Actual production wrapper rolls back the owned mutation.
			})
			if err != nil {
				t.Fatal(err)
			}
			got, err := readIntrinsicHardeningState(ctx, s)
			if err != nil || !reflect.DeepEqual(got, before) {
				t.Fatalf("%s callback rollback/cleanup changed state: %v", mode, err)
			}
		})
	}
}

func TestIntrinsicHardeningStateObserverErrors(t *testing.T) {
	t.Run("canceled-publication-is-empty", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		got, err := readIntrinsicHardeningState(ctx, &Store{})
		if !errors.Is(err, context.Canceled) || got != nil {
			t.Fatalf("cancellation cause/result lost: %v %v", got, err)
		}
	})
	t.Run("budget-before-query", func(t *testing.T) {
		tx, _ := intrinsicProfileTransaction(t)
		got, err := captureIntrinsicHardeningStateInTx(context.Background(), tx, &Store{}, &intrinsicMetadataBudget{})
		if !errors.Is(err, ErrLimitExceeded) || got != nil {
			t.Fatalf("budget/result lost: %v %v", got, err)
		}
	})
	t.Run("token-query-cause", func(t *testing.T) {
		tx, mock := intrinsicProfileTransaction(t)
		injected := errors.New("observer token read failed")
		mock.ExpectQuery(`SELECT writer_token FROM graph_preview_scope WHERE singleton=1`).WillReturnError(injected)
		got, err := captureIntrinsicHardeningStateInTx(context.Background(), tx, &Store{}, newIntrinsicMetadataBudget())
		if !errors.Is(err, injected) || got != nil {
			t.Fatalf("query cause/partial result lost: %v %v", got, err)
		}
	})
}
