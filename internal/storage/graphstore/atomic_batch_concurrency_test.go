//go:build cgo

package graphstore

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/canonicalbatch"
	"github.com/steveyegge/beads/internal/canonicalbatchstate"
)

// The embedded production graphstore pool has one session. A goroutine race on
// that connection would only prove serialized calls; reopening its directory
// would test an engine file lock. These tests instead require an explicitly
// supplied released server and two independent production one-session pools.
// They own a unique disposable layout-15 database and do not provision a server.
func atomicBatchRaceServerFixture(t *testing.T) (context.Context, Options) {
	t.Helper()
	portText := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
	if portText == "" {
		t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for actual overlapping canonical-batch transactions; serialized embedded calls are not this proof")
	}
	port, err := strconv.Atoi(portText)
	if err != nil || port < 1 || port > 65535 {
		t.Fatal("invalid BEADS_GRAPH_TEST_SERVER_PORT")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	t.Cleanup(cancel)
	o := testOptions(t)
	token, err := freshToken()
	if err != nil {
		t.Fatal(err)
	}
	o.Backend, o.Database = "server", "graph_atomic_race_"+token
	o.Binding.SchemaVersion = AtomicBatchSchemaVersion
	o.ServerHost, o.ServerPort, o.ServerUser = "127.0.0.1", port, "root"
	if err := InitAtomicBatch(ctx, o); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanupCtx, stop := context.WithTimeout(context.Background(), 10*time.Second)
		defer stop()
		admin, err := openBackend(cleanupCtx, o, "")
		if err != nil {
			t.Error(err)
			return
		}
		if _, err := admin.db.ExecContext(cleanupCtx, "DROP DATABASE `"+o.Database+"`"); err != nil {
			t.Error(err)
		}
		if err := admin.Close(); err != nil {
			t.Error(err)
		}
	})
	return ctx, o
}

func openAtomicBatchRaceStore(t *testing.T, ctx context.Context, o Options) *Store {
	t.Helper()
	s, err := OpenAtomicBatch(ctx, o)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := s.Close(); err != nil {
			t.Error(err)
		}
	})
	return s
}

type atomicBatchRaceOutcome struct {
	index  int
	result BatchResult
	err    error
}

// The coordination hook occurs only after each fresh decision acquired its
// complete baseline and checked its guard. Requiring BOTH hooks before release
// prevents a serialized successful replay or stale-guard refusal from counting
// as overlapping transactions. Context cancellation releases every goroutine.
func raceAtomicBatchServerDecisions(t *testing.T, parent context.Context, stores [2]*Store, batches [2]*canonicalbatch.Batch, keys [2]string) [2]atomicBatchRaceOutcome {
	t.Helper()
	if stores[0].db == stores[1].db || stores[0].db.Stats().MaxOpenConnections != 1 || stores[1].db.Stats().MaxOpenConnections != 1 {
		t.Fatal("race requires two independent ordinary one-session server pools")
	}
	ctx, cancel := context.WithTimeout(parent, 60*time.Second)
	reached, release := make(chan int, 2), make(chan struct{})
	results := make(chan atomicBatchRaceOutcome, 2)
	var writers sync.WaitGroup
	defer func() {
		cancel()
		writers.Wait()
		for _, store := range stores {
			store.afterWrite = nil
		}
	}()
	for index, store := range stores {
		store.afterWrite = func(stage string) error {
			if stage != "coordination" {
				return nil
			}
			select {
			case reached <- index:
			case <-ctx.Done():
				return ctx.Err()
			}
			select {
			case <-release:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		writers.Add(1)
		go func() {
			defer writers.Done()
			result, err := store.ApplyCanonicalBatch(ctx, batches[index], keys[index], "race-author")
			results <- atomicBatchRaceOutcome{index: index, result: result, err: err}
		}()
	}
	seen := map[int]bool{}
	for range 2 {
		select {
		case index := <-reached:
			if seen[index] {
				t.Fatal("one writer reached the overlap barrier twice")
			}
			seen[index] = true
		case early := <-results:
			t.Fatalf("writer %d completed before both fresh transactions overlapped: %v", early.index, early.err)
		case <-ctx.Done():
			t.Fatal("two independent transactions did not reach overlap barrier: ", ctx.Err())
		}
	}
	close(release)
	var outcomes [2]atomicBatchRaceOutcome
	for range 2 {
		select {
		case outcome := <-results:
			outcomes[outcome.index] = outcome
		case <-ctx.Done():
			t.Fatal("overlapping decisions did not terminate: ", ctx.Err())
		}
	}
	writers.Wait()
	return outcomes
}

func atomicBatchRaceCount(t *testing.T, ctx context.Context, s *Store, query string, args ...any) int {
	t.Helper()
	var count int
	if err := s.db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}

func TestCanonicalBatchServerConcurrentSameKey(t *testing.T) {
	ctx, options := atomicBatchRaceServerFixture(t)
	first, second := openAtomicBatchRaceStore(t, ctx, options), openAtomicBatchRaceStore(t, ctx, options)
	batch := atomicBatchCompile(t, first, `{"operation":"createBead","id":"winner","type":"memory","properties":{"title":"Committed once","body":"complete"}}
{"operation":"createBead","id":"consumed-once","type":"memory","properties":{"title":"Temporary","body":""}}
{"operation":"deleteBead","bead":"consumed-once"}`)
	outcomes := raceAtomicBatchServerDecisions(t, ctx, [2]*Store{first, second}, [2]*canonicalbatch.Batch{batch, batch}, [2]string{"shared-key", "shared-key"})
	winner, conflicts := -1, 0
	for i, outcome := range outcomes {
		if outcome.err == nil {
			if winner != -1 {
				t.Fatal("two fresh transactions reported success for one key")
			}
			winner = i
		} else if errors.Is(outcome.err, ErrConflict) && !errors.Is(outcome.err, ErrOutcomeUnknown) {
			conflicts++
		} else {
			t.Fatalf("writer %d was not a success or known conflict: %v", i, outcome.err)
		}
	}
	if winner == -1 || conflicts != 1 {
		t.Fatalf("winner=%d known-conflicts=%d", winner, conflicts)
	}
	before := workflowState(t, ctx, first)
	// This single explicit caller retry follows a KNOWN conflict. It must
	// recover the winner's complete immutable result, never execute the batch.
	replayed, err := second.ApplyCanonicalBatch(ctx, batch, "shared-key", "race-author")
	if err != nil || !reflect.DeepEqual(outcomes[winner].result, replayed) {
		t.Fatalf("same-key conflict retry did not recover original outcome: %+v %v", replayed, err)
	}
	if !reflect.DeepEqual(before, workflowState(t, ctx, first)) {
		t.Fatal("retained replay changed committed graph or writer token")
	}
	for _, query := range []string{
		`SELECT COUNT(*) FROM graph_preview_batch_decisions WHERE decision_key='shared-key'`,
		`SELECT COUNT(*) FROM graph_preview_catalog WHERE public_id='winner' AND allocation_state='live'`,
		`SELECT COUNT(*) FROM graph_preview_payloads`,
		`SELECT COUNT(*) FROM graph_preview_versions`,
		`SELECT COUNT(*) FROM graph_preview_batch_allocations WHERE public_id='consumed-once'`,
	} {
		if count := atomicBatchRaceCount(t, ctx, first, query); count != 1 {
			t.Fatalf("duplicate or partial publication (%d): %s", count, query)
		}
	}
	if count := atomicBatchRaceCount(t, ctx, first, `SELECT COUNT(*) FROM graph_preview_batch_allocations`); count != 2 {
		t.Fatalf("committed identities were not consumed exactly once: %d", count)
	}
	if _, err := first.ResolveExactAllocatedResourcePath(ctx, "consumed-once"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("allocation-only identity became a Resource: %v", err)
	}
}

func TestCanonicalBatchServerConcurrentRevisionGuard(t *testing.T) {
	ctx, options := atomicBatchRaceServerFixture(t)
	first, second := openAtomicBatchRaceStore(t, ctx, options), openAtomicBatchRaceStore(t, ctx, options)
	seed, err := first.ApplyCanonicalBatch(ctx, atomicBatchCompile(t, first, `{"operation":"createBead","id":"guarded","type":"memory","properties":{"title":"Original","body":""}}`), "seed", "race-author")
	if err != nil {
		t.Fatal(err)
	}
	original := atomicBatchMemory(t, first, seed, "guarded")
	var batches [2]*canonicalbatch.Batch
	keys := [2]string{"left-key", "right-key"}
	for i := range batches {
		batches[i] = atomicBatchCompile(t, first, fmt.Sprintf(`{"operation":"createBead","id":"consumed-%d","type":"memory","properties":{"title":"Transient","body":""}}
{"operation":"deleteBead","bead":"consumed-%d"}
{"operation":"updateBead","bead":"guarded","expectedRevision":%q,"propertiesChange":[{"op":"replace","path":"/title","value":"candidate-%d"}]}`, i, i, original.Revision, i))
	}
	outcomes := raceAtomicBatchServerDecisions(t, ctx, [2]*Store{first, second}, batches, keys)
	winner, loser := -1, -1
	for i, outcome := range outcomes {
		if outcome.err == nil {
			if winner != -1 {
				t.Fatal("two decisions committed against one original revision")
			}
			winner = i
		} else if errors.Is(outcome.err, ErrConflict) && !errors.Is(outcome.err, ErrOutcomeUnknown) {
			loser = i
		} else {
			t.Fatalf("guard competitor %d was not a known conflict: %v", i, outcome.err)
		}
	}
	if winner == -1 || loser == -1 {
		t.Fatalf("winner=%d known-conflict loser=%d", winner, loser)
	}
	path, err := first.ResolveExactAllocatedResourcePath(ctx, "guarded")
	if err != nil {
		t.Fatal(err)
	}
	current, err := first.Show(ctx, path)
	if err != nil || current.Properties.Title != fmt.Sprintf("candidate-%d", winner) || current.Revision == original.Revision {
		t.Fatalf("winner/loser content or revision mismatch: %+v %v", current, err)
	}
	if count := atomicBatchRaceCount(t, ctx, first, `SELECT COUNT(*) FROM graph_preview_versions WHERE path=?`, path); count != 2 {
		t.Fatalf("guard race minted %d versions instead of seed plus one", count)
	}
	for i := range outcomes {
		want := 0
		if i == winner {
			want = 1
		}
		if count := atomicBatchRaceCount(t, ctx, first, `SELECT COUNT(*) FROM graph_preview_batch_decisions WHERE decision_key=?`, keys[i]); count != want {
			t.Fatalf("competitor %d retained %d decisions, want %d", i, count, want)
		}
		if count := atomicBatchRaceCount(t, ctx, first, `SELECT COUNT(*) FROM graph_preview_batch_allocations WHERE public_id=?`, fmt.Sprintf("consumed-%d", i)); count != want {
			t.Fatalf("competitor %d retained %d allocations, want %d", i, count, want)
		}
	}
	before := workflowState(t, ctx, first)
	// One explicit retry is now evaluated against the successor, so the same
	// frozen original guard must refuse without a new decision or allocation.
	if _, err := second.ApplyCanonicalBatch(ctx, batches[loser], keys[loser], "race-author"); !errors.Is(err, canonicalbatchstate.ErrRevision) || errors.Is(err, ErrOutcomeUnknown) {
		t.Fatalf("known-conflict retry did not refuse stale original revision: %v", err)
	}
	if !reflect.DeepEqual(before, workflowState(t, ctx, first)) || atomicBatchRaceCount(t, ctx, first, `SELECT COUNT(*) FROM graph_preview_batch_decisions WHERE decision_key=?`, keys[loser]) != 0 || atomicBatchRaceCount(t, ctx, first, `SELECT COUNT(*) FROM graph_preview_batch_allocations WHERE public_id=?`, fmt.Sprintf("consumed-%d", loser)) != 0 {
		t.Fatal("failed guard retry retained graph, allocation or decision effects")
	}
}

func TestCanonicalBatchServerLostCommitResponseRecovery(t *testing.T) {
	ctx, options := atomicBatchRaceServerFixture(t)
	proxyPort, observed := startCommitLossProxy(t, net.JoinHostPort(options.ServerHost, strconv.Itoa(options.ServerPort)))
	proxied := options
	proxied.ServerPort = proxyPort
	uncertain := openAtomicBatchRaceStore(t, ctx, proxied)
	batch := atomicBatchCompile(t, uncertain, `{"operation":"createBead","id":"uncertain","type":"memory","properties":{"title":"Accepted response lost","body":"Still complete"}}`)
	result, err := uncertain.ApplyCanonicalBatch(ctx, batch, "lost-response", "race-author")
	if !errors.Is(err, ErrOutcomeUnknown) || errors.Is(err, ErrConflict) || !reflect.ValueOf(result).IsZero() {
		t.Fatalf("lost COMMIT response must return unknown with zero result: %+v %v", result, err)
	}
	select {
	case packet := <-observed:
		// This OK witness belongs to the harness; the interrupted client was
		// correctly uncertain. It must not infer rollback and replay blindly.
		if len(packet) == 0 || packet[0] != 0x00 {
			t.Fatalf("server did not acknowledge the withheld COMMIT with OK: %d bytes", len(packet))
		}
	case <-ctx.Done():
		t.Fatal("proxy did not witness COMMIT response: ", ctx.Err())
	}
	if err := uncertain.Close(); err != nil {
		t.Fatal(err)
	}
	direct := openAtomicBatchRaceStore(t, ctx, options)
	// This is an explicit recovery of the frozen key, not automatic replay of
	// a newly generated request. Its created result must already be retained.
	recovered, err := direct.ApplyCanonicalBatch(ctx, batch, "lost-response", "race-author")
	if err != nil || !recovered.Changed || recovered.Transaction == "" || len(recovered.Resources) != 1 || recovered.Resources[0].Outcome != "created" {
		t.Fatalf("committed frozen outcome was not recovered: %+v %v", recovered, err)
	}
	memory := atomicBatchMemory(t, direct, recovered, "uncertain")
	if memory.Properties.Title != "Accepted response lost" || memory.Properties.Body != "Still complete" || memory.Attribution.Actor != "race-author" {
		t.Fatalf("incomplete recovered postimage: %+v", memory)
	}
	before := workflowState(t, ctx, direct)
	replayed, err := direct.ApplyCanonicalBatch(ctx, batch, "lost-response", "race-author")
	if err != nil || !reflect.DeepEqual(recovered, replayed) || !reflect.DeepEqual(before, workflowState(t, ctx, direct)) {
		t.Fatalf("subsequent replay changed original outcome or graph: %+v %v", replayed, err)
	}
	for _, query := range []string{
		`SELECT COUNT(*) FROM graph_preview_batch_decisions WHERE decision_key='lost-response'`,
		`SELECT COUNT(*) FROM graph_preview_catalog WHERE public_id='uncertain'`,
		`SELECT COUNT(*) FROM graph_preview_payloads`,
		`SELECT COUNT(*) FROM graph_preview_versions`,
		`SELECT COUNT(*) FROM graph_preview_batch_allocations WHERE public_id='uncertain'`,
	} {
		if count := atomicBatchRaceCount(t, ctx, direct, query); count != 1 {
			t.Fatalf("lost-response recovery duplicated or omitted state (%d): %s", count, query)
		}
	}
}
