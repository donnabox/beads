package resourceget

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/steveyegge/beads/graphops"
)

func TestResourceDocumentPreservesNativeIdentityAndExactType(t *testing.T) {
	fixture := definitionFixture(t)
	fixture["id"] = definitionScope + "333"
	fixture["properties"] = map[string]any{"title": "Hello world", "issue_type": "task", "priority": 2}
	// The fixture's whole exact Type URI is opaque; it declares no native alias.
	fixture["type"] = map[string]any{"uri": "https://types.example/native-task", "revision": "actual-type-edition"}
	fixture["ownedLinks"] = map[string]any{}
	raw := definitionBytes(t, fixture)
	doc, err := NewResourceDocument(definitionScope, graphops.KindBead, raw)
	if err != nil {
		t.Fatal(err)
	}
	if doc.ID() != definitionScope+"333" || doc.Kind() != graphops.KindBead || doc.Type() != (TypeReference{URI: "https://types.example/native-task", Revision: "actual-type-edition"}) {
		t.Fatal("lost native identity or exact affiliation")
	}
	if _, ok := doc.Definition(); ok {
		t.Fatal("inferred definition authority from a Resource shape")
	}
	before := doc.NativeJSON()
	raw[0] = 'x'
	copy := doc.NativeJSON()
	copy[0] = 'x'
	if !bytes.Equal(before, doc.NativeJSON()) {
		t.Fatal("mutable Resource bytes")
	}
	encoded, err := json.Marshal(doc)
	if err != nil || !bytes.Equal(encoded, before) {
		t.Fatal("changed complete Resource representation", err)
	}
	for _, value := range []any{"https://types.example/native-task", map[string]any{"uri": "https://types.example/native-task"}, nil} {
		fixture["type"] = value
		if _, err := NewResourceDocument(definitionScope, graphops.KindBead, definitionBytes(t, fixture)); err == nil {
			t.Fatal("admitted missing exact affiliation", value)
		}
	}
}

func TestDefinitionResourceConversionPreservesCheckedDisplayOrigin(t *testing.T) {
	doc, err := NewDefinitionDocument(definitionScope, DefinitionBead, definitionBytes(t, definitionFixture(t)))
	if err != nil {
		t.Fatal(err)
	}
	doc, err = doc.WithOrigin("https://types.example/Type", "type-2")
	if err != nil {
		t.Fatal(err)
	}
	r, err := doc.Resource()
	if err != nil {
		t.Fatal(err)
	}
	view, ok := r.Definition()
	if !ok {
		t.Fatal("lost explicit definition conversion")
	}
	origin, ok := view.Origin()
	if !ok || origin != (DefinitionOrigin{URI: "https://types.example/Type", Revision: "type-2"}) {
		t.Fatal("lost exact carried correspondence")
	}
	if !bytes.Equal(doc.NativeJSON(), r.NativeJSON()) || !bytes.Equal(view.NativeJSON(), r.NativeJSON()) {
		t.Fatal("display metadata changed Resource bytes")
	}
	if _, err := (DefinitionDocument{}).Resource(); err == nil {
		t.Fatal("converted zero definition")
	}
	if _, err := json.Marshal(ResourceDocument{}); err == nil {
		t.Fatal("marshaled zero Resource")
	}
}

func TestResourceOwnedAbsencePreservesDefinitionRequirements(t *testing.T) {
	fixture := definitionFixture(t)
	delete(fixture, "ownedLinks")
	raw := definitionBytes(t, fixture)
	document, err := NewResourceDocument(definitionScope, graphops.KindBead, raw)
	if err != nil || !bytes.Equal(document.NativeJSON(), raw) {
		t.Fatalf("ordinary absence was changed: %s, %v", document.NativeJSON(), err)
	}
	if _, err := NewDefinitionDocument(definitionScope, DefinitionBead, raw); err == nil {
		t.Fatal("DefinitionBead lost mandatory ownedLinks")
	}
	for name, value := range map[string]any{"null": nil, "array": []any{}, "scalar": true} {
		t.Run(name, func(t *testing.T) {
			fixture["ownedLinks"] = value
			if _, err := NewResourceDocument(definitionScope, graphops.KindBead, definitionBytes(t, fixture)); err == nil {
				t.Fatal("malformed present ownedLinks accepted")
			}
		})
	}
	fixture["ownedLinks"] = map[string]any{"https://types.example/historical-selector": []any{}}
	historical := definitionBytes(t, fixture)
	ordinary, err := NewResourceDocument(definitionScope, graphops.KindBead, historical)
	if err != nil || !bytes.Equal(ordinary.NativeJSON(), historical) {
		t.Fatalf("historical empty bucket was rewritten: %s, %v", ordinary.NativeJSON(), err)
	}
	definition, err := NewDefinitionDocument(definitionScope, DefinitionBead, historical)
	if err != nil || !bytes.Equal(definition.NativeJSON(), historical) {
		t.Fatalf("historical Definition shape was rewritten: %s, %v", definition.NativeJSON(), err)
	}
}
