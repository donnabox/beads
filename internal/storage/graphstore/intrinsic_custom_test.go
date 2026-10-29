//go:build cgo

package graphstore

import (
	"bytes"
	"context"
	"database/sql"
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

// Capture exactly the protected intrinsic records and fence, not an entire
// database dump. Readbacks use production readers, including retained captures.
func intrinsicHardeningState(t *testing.T, ctx context.Context, s *Store) []string {
	t.Helper()
	var token string
	if err := s.db.QueryRowContext(ctx, `SELECT writer_token FROM graph_preview_scope WHERE singleton=1`).Scan(&token); err != nil {
		t.Fatal(err)
	}
	state := []string{token}
	for _, id := range []string{"type-for-type", "type-for-bead-type", "type-for-link-type", "type-for-refines", "bead-type-refines-type", "link-type-refines-type"} {
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
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	// These staged deletes make the already-correct physical schema empty for
	// the production bootstrap function; no deleted bytes are ever committed.
	for _, q := range []string{`DELETE FROM graph_preview_versions`, `DELETE FROM graph_preview_intrinsic_members`, `DELETE FROM graph_preview_intrinsic_contexts`, `DELETE FROM graph_preview_catalog`} {
		if _, err := tx.ExecContext(ctx, q); err != nil {
			_ = tx.Rollback()
			t.Fatal(err)
		}
	}
	// Exactly its initial source/copy reservation: all six INSERTs are staged,
	// then the independent final integrity-census acquisition must refuse.
	budget := &intrinsicMetadataBudget{remaining: 4*intrinsictypecatalog.CurrentSource().Bytes + 2*intrinsicmaterialization.MaxContextBytes}
	err = installIntrinsicMembers(ctx, tx, o, budget)
	if !errors.Is(err, ErrInvalidStore) || !strings.Contains(err.Error(), ErrLimitExceeded.Error()) {
		_ = tx.Rollback()
		t.Fatalf("final census did not refuse: %v", err)
	}
	for _, q := range []string{`SELECT COUNT(*) FROM graph_preview_catalog`, `SELECT COUNT(*) FROM graph_preview_intrinsic_members`, `SELECT COUNT(*) FROM graph_preview_versions`} {
		var n int
		if err := tx.QueryRowContext(ctx, q).Scan(&n); err != nil || n != 6 {
			_ = tx.Rollback()
			t.Fatalf("expected six staged members: %d %v", n, err)
		}
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if after := intrinsicHardeningState(t, ctx, s); !reflect.DeepEqual(before, after) {
		t.Fatal("failed materialization changed committed intrinsic state")
	}
}
