//go:build cgo

package graphstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

// This test alters only an owned disposable store. A retained decision must be
// checked as historical data, never normalized into a success or re-executed.
func TestCanonicalBatchNativeCorruptRecovery(t *testing.T) {
	ctx, _, s := atomicBatchFixture(t)
	batch := atomicBatchCompile(t, s, `{"operation":"createBead","id":"a","type":"memory","properties":{"title":"Original","body":""}}`)
	original, err := s.ApplyCanonicalBatch(ctx, batch, "recover", "test-author")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := canonicalJSON(original)
	if err != nil {
		t.Fatal(err)
	}
	for _, scenario := range []string{"empty-resource", "unknown-field", "false-change"} {
		var fields map[string]any
		if err := json.Unmarshal(raw, &fields); err != nil {
			t.Fatal(err)
		}
		switch scenario {
		case "empty-resource":
			fields["resources"] = []any{map[string]any{}}
		case "unknown-field":
			fields["future"] = true
		case "false-change":
			fields["changed"] = false
		}
		corrupt, err := canonicalJSON(fields)
		if err != nil {
			t.Fatal(err)
		}
		if err := s.withTx(ctx, true, func(tx *sql.Tx) error {
			_, err := tx.ExecContext(ctx, `UPDATE graph_preview_batch_decisions SET result=? WHERE decision_key=?`, corrupt, "recover")
			return err
		}); err != nil {
			t.Fatal(err)
		}
		got, err := s.ApplyCanonicalBatch(ctx, batch, "recover", "test-author")
		if !errors.Is(err, ErrInvalidStore) || !reflect.DeepEqual(got, BatchResult{}) {
			t.Fatalf("corrupt recovery invented a result: %+v %v", got, err)
		}
	}
	if err := s.withTx(ctx, true, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `UPDATE graph_preview_batch_decisions SET result=? WHERE decision_key=?`, raw, "recover")
		return err
	}); err != nil {
		t.Fatal(err)
	}
	got, err := s.ApplyCanonicalBatch(ctx, batch, "recover", "test-author")
	if err != nil || !reflect.DeepEqual(got, original) {
		t.Fatalf("valid recovery changed after rejecting corruption: %+v %v", got, err)
	}
	var count int
	path, err := s.ResolveExactAllocatedResourcePath(ctx, "a")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM graph_preview_versions WHERE path=?`, path).Scan(&count); err != nil || count != 1 {
		t.Fatalf("replay published another version: count=%d err=%v", count, err)
	}
}

func TestCanonicalBatchNativeProfileIsolation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	o := testOptions(t)
	if err := Init(ctx, o); err != nil {
		t.Fatal(err)
	}
	atomic := o
	atomic.Binding.SchemaVersion = AtomicBatchSchemaVersion
	if wrong, err := OpenAtomicBatch(ctx, atomic); err == nil {
		_ = wrong.Close()
		t.Fatal("ordinary layout opened as atomic batch")
	}
	if err := InitAtomicBatch(ctx, atomic); err == nil {
		t.Fatal("atomic initialization upgraded an ordinary workspace")
	}
	ordinary, err := OpenExisting(ctx, o)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = ordinary.Close() }()
	batch := atomicBatchCompile(t, ordinary, `{"operation":"createBead","id":"a","type":"memory","properties":{"title":"Must refuse","body":""}}`)
	if _, err := ordinary.ApplyCanonicalBatch(ctx, batch, "not-admitted", "test-author"); !errors.Is(err, ErrCapabilityUnavailable) {
		t.Fatalf("ordinary store executed a batch: %v", err)
	}
	_, ao, candidate := atomicBatchFixture(t)
	if err := candidate.Close(); err != nil {
		t.Fatal(err)
	}
	wrong := ao
	wrong.Binding.SchemaVersion = SchemaVersion
	if reopened, err := OpenExisting(ctx, wrong); err == nil {
		_ = reopened.Close()
		t.Fatal("atomic workspace silently fell back to ordinary layout")
	}
	reopened, err := OpenAtomicBatch(ctx, ao)
	if err != nil {
		t.Fatalf("wrong-profile open altered candidate: %v", err)
	}
	_ = reopened.Close()
}

func TestCanonicalBatchNativeOwnedCaptureRecoveryOrder(t *testing.T) {
	ctx, _, s := atomicBatchFixture(t)
	_, err := s.ApplyCanonicalBatch(ctx, atomicBatchCompile(t, s, `{"operation":"createBead","id":"owner","type":"memory","properties":{"title":"Owner","body":""}}
{"operation":"createBead","id":"rule","type":"constraint","properties":{"title":"Rule","body":""}}`), "seed", "test-author")
	if err != nil {
		t.Fatal(err)
	}
	// Ordinary checked writers can assign public IDs independently of private
	// paths. These two valid captures deliberately reverse public-ID ordering.
	for _, identity := range [][2]string{{"links/a", "z-owned"}, {"links/z", "a-owned"}} {
		if _, err := s.AddInformationalLink(ctx, LinkCreateRequest{
			Path: identity[0], PublicID: identity[1], TypeURL: s.ScopeURL() + "subject-to",
			SourcePath: s.ScopeURL() + "owner", TargetPath: s.ScopeURL() + "rule", UnconditionalSource: true, Actor: "test-author",
		}); err != nil {
			t.Fatal(err)
		}
	}
	batch := atomicBatchCompile(t, s, `{"operation":"updateBead","bead":"owner","propertiesChange":[{"op":"replace","path":"/title","value":"Final"}]}`)
	result, err := s.ApplyCanonicalBatch(ctx, batch, "ordered-capture", "test-author")
	if err != nil {
		t.Fatal(err)
	}
	owner := atomicBatchMemory(t, s, result, "owner")
	if len(owner.Owned) != 2 {
		t.Fatalf("expected two owned captures: %+v", owner)
	}
	for i, want := range []string{"z-owned", "a-owned"} {
		var link LinkRecord
		if err := json.Unmarshal(owner.Owned[i], &link); err != nil || link.ID != s.ScopeURL()+want {
			t.Fatalf("fixture did not reverse private/public order: %+v %v", link, err)
		}
	}
	replayed, err := s.ApplyCanonicalBatch(ctx, batch, "ordered-capture", "test-author")
	if err != nil || !reflect.DeepEqual(result, replayed) {
		t.Fatalf("valid owned carrier order was refused or rewritten: %+v %v", replayed, err)
	}
}
