//go:build cgo

package graphstore

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

// This boundary is persistence and production initialization, not a parser or
// mocked-engine proof. No preinserted record or publication lookup can pass it.
func TestIntrinsicBeadTypeProductionInitReopen(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	o := testOptions(t)
	o.Binding.SchemaVersion = TypesFreshSchemaVersion
	if err := Init(ctx, o); err != nil {
		t.Fatal(err)
	}
	s, err := OpenExisting(ctx, o)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if e := s.Close(); e != nil {
			t.Error(e)
		}
	}()
	first, err := s.GetIntrinsicDefinition(ctx, "type-for-bead-type", "")
	if err != nil {
		t.Fatal(err)
	}
	var body struct {
		ID       string                         `json:"id"`
		Revision string                         `json:"revision"`
		Type     struct{ URI, Revision string } `json:"type"`
		Owned    []json.RawMessage              `json:"owned"`
	}
	if err = json.Unmarshal(first.Body, &body); err != nil {
		t.Fatal(err)
	}
	if first.PublicID != "type-for-bead-type" || body.ID != o.Binding.ScopeURL+"type-for-bead-type" || body.Revision != first.Revision || body.Type.URI != "https://github.com/gastownhall/bdp/types/type-system/bead-type" || body.Type.Revision != "bead-type-2" {
		t.Fatalf("wrong persisted definition: %s", first.Body)
	}
	if first.Global.URI != "https://github.com/gastownhall/bdp/types/type-system/bead-type" || first.Global.Revision != "bead-type-2" {
		t.Fatal("native origin differs from checked shared definition")
	}
	first.Body[0] = '!'
	again, err := s.GetIntrinsicDefinition(ctx, "type-for-bead-type", first.Revision)
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(again.Body) {
		t.Fatal("caller mutation changed retained body")
	}
	if err = s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err = OpenExisting(ctx, o)
	if err != nil {
		t.Fatal(err)
	}
	reopened, err := s.GetIntrinsicDefinition(ctx, "type-for-bead-type", "")
	if err != nil {
		t.Fatal(err)
	}
	if reopened.Revision != again.Revision || !bytes.Equal(reopened.Body, again.Body) {
		t.Fatal("reopen changed actual local record")
	}
	for _, localID := range []string{"type-for-bead-type", ""} {
		doc, err := s.GetDefinitionForProjection(ctx, "beads/type-for-bead-type", localID, again.Revision)
		if err != nil {
			t.Fatal(err)
		}
		origin, ok := doc.Origin()
		if !ok || origin.URI != first.Global.URI || origin.Revision != first.Global.Revision {
			t.Fatal("projection origin did not follow checked retained binding")
		}
	}
	foo, err := s.CreateBeadType(ctx, "omitted-schema-Foo", "omission-control")
	if err != nil {
		t.Fatal(err)
	}
	if _, present := foo.Origin(); present {
		t.Fatal("custom definition has intrinsic display origin")
	}
	var definition struct {
		Properties map[string]json.RawMessage     `json:"properties"`
		Type       struct{ URI, Revision string } `json:"type"`
	}
	if err := json.Unmarshal(foo.NativeJSON(), &definition); err != nil {
		t.Fatal(err)
	}
	if _, present := definition.Properties["propertiesSchema"]; present || definition.Type.Revision != "bead-type-2" {
		t.Fatal("new definition manufactured a local schema or lost exact pin")
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err = OpenExisting(ctx, o)
	if err != nil {
		t.Fatal(err)
	}
	retained, err := s.GetDefinition(ctx, "omitted-schema-Foo", foo.Revision())
	if _, present := retained.Origin(); present {
		t.Fatal("custom exact read acquired intrinsic display origin")
	}
	if err != nil || !bytes.Equal(retained.NativeJSON(), foo.NativeJSON()) {
		t.Fatalf("omission changed on reopen: %v", err)
	}
	t.Run("unknown-exact-edition", func(t *testing.T) {
		if _, e := s.GetDefinition(ctx, "type-for-bead-type", "unavailable-exact-edition"); !errors.Is(e, ErrVersionUnknown) {
			t.Fatalf("unknown exact: %v", e)
		}
	})
	t.Run("catalog-affiliation-only", func(t *testing.T) {
		wrong := body.Type.URI[:len(body.Type.URI)-1] + "X"
		if _, e := s.db.ExecContext(ctx, `UPDATE graph_preview_catalog SET type_url=? WHERE path=?`, wrong, "beads/type-for-bead-type"); e != nil {
			t.Fatal(e)
		}
		defer func() {
			if _, e := s.db.ExecContext(ctx, `UPDATE graph_preview_catalog SET type_url=? WHERE path=?`, body.Type.URI, "beads/type-for-bead-type"); e != nil {
				t.Error(e)
			}
		}()
		if _, e := s.GetDefinition(ctx, "type-for-bead-type", ""); !errors.Is(e, ErrInvalidStore) {
			t.Fatalf("catalog-only corruption escaped: %v", e)
		}
	})
	t.Run("missing-current-member", func(t *testing.T) {
		// Stage the corruption in an ordinary SQL Tx and roll back; production
		// reader executes in that SAME Tx so zero committed test damage remains.
		tx, e := s.db.BeginTx(ctx, nil)
		if e != nil {
			t.Fatal(e)
		}
		defer func() {
			if e := tx.Rollback(); e != nil {
				t.Error(e)
			}
		}()
		if _, e = tx.ExecContext(ctx, `DELETE FROM graph_preview_intrinsic_members WHERE path=? AND revision=?`, "beads/type-for-bead-type", again.Revision); e != nil {
			t.Fatal(e)
		}
		budget := newIntrinsicMetadataBudget()
		var doc IntrinsicDocument
		if e = s.readIntrinsicDocumentInTx(ctx, tx, "beads/type-for-bead-type", "", &doc, budget); !errors.Is(e, ErrInvalidStore) {
			t.Fatalf("missing current: %v", e)
		}
		if e = s.readIntrinsicDocumentInTx(ctx, tx, "beads/type-for-bead-type", again.Revision, &doc, budget); !errors.Is(e, ErrInvalidStore) {
			t.Fatalf("missing retained mapping: %v", e)
		}
	})

}
