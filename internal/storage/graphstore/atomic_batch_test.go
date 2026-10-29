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
	"strings"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/canonicalbatch"
	"github.com/steveyegge/beads/internal/canonicalbatchstate"
)

// These tests own a fresh embedded engine/database each. They do not borrow an
// unauthenticated external server or touch an installed workspace. The root
// qualification lane controls execution; source authoring is not an engine PASS.
func atomicBatchFixture(t *testing.T) (context.Context, Options, *Store) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	t.Cleanup(cancel)
	o := testOptions(t)
	o.Binding.SchemaVersion = AtomicBatchSchemaVersion
	if err := InitAtomicBatch(ctx, o); err != nil {
		t.Fatal(err)
	}
	s, err := OpenAtomicBatch(ctx, o)
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

func atomicBatchCompile(t *testing.T, s *Store, input string) *canonicalbatch.Batch {
	t.Helper()
	batch, err := canonicalbatch.Compile([]byte(input), canonicalbatch.Options{Scope: s.ScopeURL()})
	if err != nil {
		t.Fatal(err)
	}
	return batch
}

func atomicBatchSeed(t *testing.T, ctx context.Context, s *Store) BatchResult {
	t.Helper()
	batch := atomicBatchCompile(t, s, `{"operation":"createBead","name":"owner","id":"owner","type":"memory","properties":{"title":"Plan","body":"original"}}
{"operation":"createBead","name":"rule","id":"rule","type":"constraint","properties":{"title":"Rule","body":"constraint"}}
{"operation":"createLink","id":"binding","type":"subject-to","source":{"binding":"owner"},"target":{"binding":"rule"},"properties":{}}`)
	result, err := s.ApplyCanonicalBatch(ctx, batch, "seed", "test-author")
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func atomicBatchOutcome(t *testing.T, result BatchResult, id string) BatchResourceOutcome {
	t.Helper()
	for _, row := range result.Resources {
		if row.ID == id {
			return row
		}
	}
	t.Fatalf("missing outcome %s: %+v", id, result)
	return BatchResourceOutcome{}
}

func atomicBatchMemory(t *testing.T, s *Store, result BatchResult, id string) Record {
	t.Helper()
	row := atomicBatchOutcome(t, result, s.ScopeURL()+id)
	var memory Record
	if err := json.Unmarshal(row.Resource, &memory); err != nil {
		t.Fatal(err)
	}
	return memory
}

func TestCanonicalBatchNativeCreateGraphAndRecovery(t *testing.T) {
	ctx, o, s := atomicBatchFixture(t)
	seed := atomicBatchSeed(t, ctx, s)
	if !seed.Changed || seed.Transaction == "" || len(seed.Results) != 3 || len(seed.Resources) != 3 {
		t.Fatalf("incomplete result: %+v", seed)
	}
	owner := atomicBatchMemory(t, s, seed, "owner")
	if len(owner.Owned) != 1 {
		t.Fatalf("creation omitted final owned Link: %+v", owner)
	}
	linkRow := atomicBatchOutcome(t, seed, s.ScopeURL()+"binding")
	var link LinkRecord
	if err := json.Unmarshal(linkRow.Resource, &link); err != nil {
		t.Fatal(err)
	}
	var captured LinkRecord
	if err := json.Unmarshal(owner.Owned[0], &captured); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(link, captured) || len(linkRow.OwnerRevisions) != 1 || linkRow.OwnerRevisions[0].Revision != owner.Revision {
		t.Fatalf("owner captured an intermediate Link: %+v %+v", linkRow, owner)
	}
	for _, id := range []string{"owner", "rule", "binding"} {
		path, err := s.ResolveExactAllocatedResourcePath(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
		var count int
		if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM graph_preview_versions WHERE path=?`, path).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Fatalf("%s published %d versions", id, count)
		}
	}
	// Recovery survives a real close/reopen and returns the original complete
	// outcome, even after a later successful decision changes today's state.
	change := atomicBatchCompile(t, s, `{"operation":"updateBead","bead":"owner","propertiesChange":[{"op":"replace","path":"/title","value":"later"}]}`)
	if _, err := s.ApplyCanonicalBatch(ctx, change, "later", "test-author"); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err := OpenAtomicBatch(ctx, o)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := s.Close(); err != nil {
			t.Error(err)
		}
	})
	seedBatch := atomicBatchCompile(t, s, `{"operation":"createBead","name":"owner","id":"owner","type":"memory","properties":{"title":"Plan","body":"original"}}
{"operation":"createBead","name":"rule","id":"rule","type":"constraint","properties":{"title":"Rule","body":"constraint"}}
{"operation":"createLink","id":"binding","type":"subject-to","source":{"binding":"owner"},"target":{"binding":"rule"},"properties":{}}`)
	replayed, err := s.ApplyCanonicalBatch(ctx, seedBatch, "seed", "test-author")
	if err != nil || !reflect.DeepEqual(seed, replayed) {
		t.Fatalf("recovery changed disposition: %+v %v", replayed, err)
	}
	if _, err := s.ApplyCanonicalBatch(ctx, seedBatch, "seed", "different-actor"); !errors.Is(err, ErrBatchKeyConflict) {
		t.Fatalf("actor created a separate key namespace: %v", err)
	}
	renamed := strings.NewReplacer(`"name":"owner"`, `"name":"renamed-owner"`, `"binding":"owner"`, `"binding":"renamed-owner"`, `"name":"rule"`, `"name":"renamed-rule"`, `"binding":"rule"`, `"binding":"renamed-rule"`).Replace(string(seedBatch.NativeJSON()))
	renamedBatch, err := canonicalbatch.Compile([]byte(renamed), canonicalbatch.Options{Format: "json", Scope: s.ScopeURL()})
	if err != nil {
		t.Fatal(err)
	}
	renamedResult, err := s.ApplyCanonicalBatch(ctx, renamedBatch, "seed", "test-author")
	if err != nil {
		t.Fatal(err)
	}
	want := seed
	want.Results = append([]BatchOperationResult(nil), seed.Results...)
	want.Results[0].OperationName, want.Results[1].OperationName = "renamed-owner", "renamed-rule"
	if !reflect.DeepEqual(want, renamedResult) {
		t.Fatalf("renamed replay changed committed outcome: %+v %+v", want, renamedResult)
	}
}

func TestCanonicalBatchNativeFinalNoop(t *testing.T) {
	ctx, _, s := atomicBatchFixture(t)
	seed := atomicBatchSeed(t, ctx, s)
	owner := atomicBatchMemory(t, s, seed, "owner")
	batch := atomicBatchCompile(t, s, fmt.Sprintf(`{"operation":"updateBead","bead":"owner","expectedRevision":%q,"propertiesChange":[{"op":"replace","path":"/title","value":"temporary"}]}
{"operation":"updateBead","bead":"owner","expectedRevision":%q,"propertiesChange":[{"op":"replace","path":"/title","value":"Plan"}]}`, owner.Revision, owner.Revision))
	result, err := s.ApplyCanonicalBatch(ctx, batch, "no-op", "test-author")
	if err != nil {
		t.Fatal(err)
	}
	if result.Changed || result.Transaction != "" {
		t.Fatalf("staged changes escaped no-op: %+v", result)
	}
	after := atomicBatchMemory(t, s, result, "owner")
	beforeJSON, _ := canonicalJSON(owner)
	afterJSON, _ := canonicalJSON(after)
	if !bytes.Equal(beforeJSON, afterJSON) {
		t.Fatalf("net no-op changed envelope: %s -> %s", beforeJSON, afterJSON)
	}
	var decisions int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM graph_preview_batch_decisions WHERE decision_key='no-op'`).Scan(&decisions); err != nil {
		t.Fatal(err)
	}
	if decisions != 1 {
		t.Fatal("successful no-op lost anti-reexecution disposition")
	}
}

func TestCanonicalBatchNativeStaleGuardRollback(t *testing.T) {
	ctx, _, s := atomicBatchFixture(t)
	atomicBatchSeed(t, ctx, s)
	batch := atomicBatchCompile(t, s, `{"operation":"createBead","id":"rolled-back","type":"memory","properties":{"title":"Never committed","body":""}}
{"operation":"updateBead","bead":"owner","expectedRevision":"stale","propertiesChange":[{"op":"replace","path":"/title","value":"bad"}]}`)
	if _, err := s.ApplyCanonicalBatch(ctx, batch, "stale", "test-author"); !errors.Is(err, canonicalbatchstate.ErrRevision) {
		t.Fatalf("guard did not refuse: %v", err)
	}
	var allocated, decisions int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM graph_preview_catalog WHERE public_id='rolled-back'`).Scan(&allocated); err != nil {
		t.Fatal(err)
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM graph_preview_batch_decisions WHERE decision_key='stale'`).Scan(&decisions); err != nil {
		t.Fatal(err)
	}
	if allocated != 0 || decisions != 0 {
		t.Fatalf("failed request left effects: allocated=%d decisions=%d", allocated, decisions)
	}
}

func TestCanonicalBatchNativeAllocationOnly(t *testing.T) {
	ctx, _, s := atomicBatchFixture(t)
	batch := atomicBatchCompile(t, s, `{"operation":"createBead","name":"temporary","id":"team:consumed","type":"memory","properties":{"title":"Never published","body":""}}
{"operation":"deleteBead","bead":{"binding":"temporary"}}`)
	result, err := s.ApplyCanonicalBatch(ctx, batch, "consume", "test-author")
	if err != nil {
		t.Fatal(err)
	}
	if !result.Changed || result.Transaction == "" || len(result.Resources) != 1 || result.Resources[0].Outcome != "allocated" || len(result.Resources[0].Resource) != 0 || result.Resources[0].Deleted != nil {
		t.Fatalf("allocated identity became a Resource: %+v", result)
	}
	var catalogs, versions, allocations int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM graph_preview_catalog`).Scan(&catalogs); err != nil {
		t.Fatal(err)
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM graph_preview_versions`).Scan(&versions); err != nil {
		t.Fatal(err)
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM graph_preview_batch_allocations WHERE public_id='team:consumed'`).Scan(&allocations); err != nil {
		t.Fatal(err)
	}
	if catalogs != 0 || versions != 0 || allocations != 1 {
		t.Fatalf("fictional state: catalog=%d versions=%d allocations=%d", catalogs, versions, allocations)
	}
	if _, err := s.Create(ctx, CreateRequest{PublicID: "team:consumed", Title: "Reused"}); !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("ordinary writer reused batch allocation: %v", err)
	}
	replay, err := s.ApplyCanonicalBatch(ctx, batch, "consume", "test-author")
	if err != nil || !reflect.DeepEqual(result, replay) {
		t.Fatalf("allocation recovery changed: %+v %v", replay, err)
	}
}

func TestCanonicalBatchNativeDeleteOwnedLinkAndOwner(t *testing.T) {
	ctx, _, s := atomicBatchFixture(t)
	seed := atomicBatchSeed(t, ctx, s)
	owner := atomicBatchMemory(t, s, seed, "owner")
	ownerPath, err := s.ResolveExactAllocatedResourcePath(ctx, "owner")
	if err != nil {
		t.Fatal(err)
	}
	batch := atomicBatchCompile(t, s, `{"operation":"deleteLink","link":"binding"}
{"operation":"deleteBead","bead":"owner"}`)
	result, err := s.ApplyCanonicalBatch(ctx, batch, "remove-owner", "test-author")
	if err != nil {
		t.Fatal(err)
	}
	row := atomicBatchOutcome(t, result, owner.ID)
	if row.Outcome != "deleted" || row.Deleted == nil || row.Deleted.Revision != owner.Revision {
		t.Fatalf("deletion invented last-live revision: %+v", row)
	}
	linkRow := atomicBatchOutcome(t, result, s.ScopeURL()+"binding")
	if len(linkRow.OwnerRevisions) != 1 || linkRow.OwnerRevisions[0].ID != owner.ID || linkRow.OwnerRevisions[0].Revision != owner.Revision {
		t.Fatalf("deleted owner's final published live revision was lost: %+v", linkRow.OwnerRevisions)
	}
	var baseline Record
	err = s.withTx(ctx, false, func(tx *sql.Tx) error {
		var err error
		baseline, err = s.deletedMemoryInTx(ctx, tx, ownerPath)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if !equalBatchOwned(owner.Owned, baseline.Owned) || baseline.Revision != owner.Revision {
		t.Fatalf("actual last-live capture was lost: %+v", baseline)
	}
	var versions int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM graph_preview_versions WHERE path=?`, ownerPath).Scan(&versions); err != nil {
		t.Fatal(err)
	}
	if versions != 2 {
		t.Fatalf("intermediate owner version published: %d", versions)
	}
}
