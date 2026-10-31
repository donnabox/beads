package main

import (
	"encoding/json"
	"net/url"
	"reflect"
	"testing"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/clioutput"
	"github.com/steveyegge/beads/internal/httpapi/graphread"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/resourceget"
)

// This captured-input corpus tests the CLI's common selectors and renderers.
// It does not replace native installation, command dispatch or reopen tests.
func mixedCanonicalResourceCapture(t *testing.T) graphstore.ResourceSnapshot {
	t.Helper()
	snapshot := commonGraphCapture(t)
	memory := resourceget.TypeReference{URI: genericTestScope + "memory", Revision: "memory-1"}
	refines := resourceget.TypeReference{URI: genericTestScope + "refines", Revision: "refines-1"}
	for _, value := range []struct {
		reference resourceget.TypeReference
		kind      graph.ResourceKind
	}{{memory, graph.KindBead}, {refines, graph.KindLink}} {
		snapshot.Types = append(snapshot.Types, graphstore.ResourceTypeFact{
			Reference: value.reference, Definition: value.reference, Describes: value.kind,
			Current: true, Ancestors: []resourceget.TypeReference{value.reference},
		})
	}
	add := func(kind graph.ResourceKind, body map[string]any) resourceget.ResourceDocument {
		t.Helper()
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		doc, err := resourceget.NewResourceDocument(genericTestScope, kind, raw)
		if err != nil {
			t.Fatal(err)
		}
		return doc
	}
	for _, value := range []struct {
		id         string
		properties map[string]any
	}{
		{"memory-instance", map[string]any{"title": "Retained memory", "body": "ordinary application content"}},
		{"descriptor-shaped-data", map[string]any{"name": "bead-type", "propertiesSchema": map[string]any{"type": "object"}, "linkConstraints": map[string]any{}, "title": "Still a Memory"}},
	} {
		snapshot.Documents = append(snapshot.Documents, add(graph.KindBead, map[string]any{
			"id": genericTestScope + value.id, "revision": value.id + "-1", "type": memory,
			"properties": value.properties, "metadata": map[string]any{"corpus": "mixed", "nested": map[string]any{"keep": true}}, "ownedLinks": map[string]any{},
		}))
	}
	link := map[string]any{
		"id": genericTestScope + "refinement", "revision": "refinement-1", "type": refines,
		"source":     resourceget.TypeReference{URI: genericTestScope + "type-child", Revision: "child-1"},
		"target":     resourceget.TypeReference{URI: genericTestScope + "type-base", Revision: "base-1"},
		"properties": map[string]any{"note": "derived to exact base"},
	}
	for i, doc := range snapshot.Documents {
		if doc.ID() != genericTestScope+"type-child" {
			continue
		}
		var body map[string]any
		if err := json.Unmarshal(doc.NativeJSON(), &body); err != nil {
			t.Fatal(err)
		}
		body["ownedLinks"].(map[string]any)[refines.URI] = []any{link}
		snapshot.Documents[i] = add(graph.KindBead, body)
	}
	snapshot.Documents = append(snapshot.Documents, add(graph.KindLink, link))
	return snapshot
}

func assertMixedResourceJSON(t *testing.T, got, want []byte) {
	t.Helper()
	var actual, expected any
	if err := json.Unmarshal(got, &actual); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(want, &expected); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("complete Resource changed:\ngot %s\nwant %s", got, want)
	}
}

func TestMixedCanonicalResourceShowAndListPreserveEveryCompleteBody(t *testing.T) {
	capture := mixedCanonicalResourceCapture(t)
	for _, doc := range capture.Documents {
		t.Run(doc.ID()[len(genericTestScope):], func(t *testing.T) {
			renderer, err := clioutput.New(canonicalShowDescriptor(doc.ID(), doc), clioutput.Options{Formats: []string{"json"}})
			if err != nil {
				t.Fatal(err)
			}
			out, err := renderer.Render(doc.NativeJSON())
			if err != nil {
				t.Fatal(err)
			}
			var envelope struct {
				Data json.RawMessage `json:"data"`
			}
			if err := json.Unmarshal(out, &envelope); err != nil {
				t.Fatal(err)
			}
			assertMixedResourceJSON(t, envelope.Data, doc.NativeJSON())
		})
	}
	for _, collection := range []string{"beads", "links"} {
		selection, err := graphread.CompileCollection(genericTestScope, collection, url.Values{}, graphread.SelectorLimits{})
		if err != nil {
			t.Fatal(err)
		}
		rows, err := selection.SelectResources(capture)
		if err != nil {
			t.Fatal(err)
		}
		want := map[string][]byte{}
		for _, doc := range capture.Documents {
			if (collection == "links") == (doc.Kind() == graph.KindLink) {
				want[doc.ID()] = doc.NativeJSON()
			}
		}
		renderer, err := clioutput.NewCollection(canonicalResourceCollectionDescriptor(collection == "links", rows), clioutput.Options{Formats: []string{"json"}})
		if err != nil {
			t.Fatal(err)
		}
		payload, err := json.Marshal(graphread.Page{Items: rows})
		if err != nil {
			t.Fatal(err)
		}
		// SelectResources enumerates the entire captured collection.
		exhausted := true
		out, err := renderer.Render(payload, clioutput.Completion{Returned: len(rows), Exhausted: &exhausted, StopReason: "exhausted"})
		if err != nil {
			t.Fatal(err)
		}
		var envelope struct {
			Data struct {
				Items []json.RawMessage `json:"items"`
			} `json:"data"`
		}
		if err := json.Unmarshal(out, &envelope); err != nil {
			t.Fatal(err)
		}
		for _, raw := range envelope.Data.Items {
			var identity struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal(raw, &identity); err != nil {
				t.Fatal(err)
			}
			body, exists := want[identity.ID]
			if !exists {
				t.Fatalf("unexpected or duplicate Resource %s", identity.ID)
			}
			assertMixedResourceJSON(t, raw, body)
			delete(want, identity.ID)
		}
		if len(want) != 0 {
			t.Fatalf("%s omitted ordinary Resources: %v", collection, want)
		}
	}
}

func TestMixedCanonicalResourceGraphAndTypeSelectionIgnoreDescriptorShape(t *testing.T) {
	capture := mixedCanonicalResourceCapture(t)
	result, err := projectResourceGraphGeneric(capture, genericTestScope, graphGenericInput{All: true, AllLinkTypes: true, Direction: "both", Depth: 10, MaxNodes: 10, MaxLinks: 10})
	if err != nil || !result.Complete || len(result.Nodes) != 6 || len(result.Links) != 2 {
		t.Fatalf("mixed ordinary graph population: %+v %v", result, err)
	}
	selected, err := graphSelectedResourceRecords(result, capture)
	if err != nil || len(selected.Beads) != 6 || len(selected.Links) != 2 {
		t.Fatalf("mixed graph payload: %+v %v", selected, err)
	}
	for _, value := range []struct {
		predicate, typ string
		ids            []string
	}{
		{"type", "bead-type", []string{"isolated-type", "type-base", "type-child"}},
		{"type", "task", []string{"333"}},
		{"type", "memory", []string{"descriptor-shaped-data", "memory-instance"}},
		{"type", "unknown", []string{}},
		{"conformsTo", "issue", []string{"333"}},
		{"conformsTo", "memory", []string{"descriptor-shaped-data", "memory-instance"}},
	} {
		selection, err := graphread.CompileCollection(genericTestScope, "beads", url.Values{value.predicate: {genericTestScope + value.typ}}, graphread.SelectorLimits{})
		if err != nil {
			t.Fatal(err)
		}
		rows, err := selection.SelectResources(capture)
		if err != nil {
			t.Fatal(err)
		}
		ids := []string{}
		for _, raw := range rows {
			var identity struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal(raw, &identity); err != nil {
				t.Fatal(err)
			}
			ids = append(ids, identity.ID[len(genericTestScope):])
		}
		if !reflect.DeepEqual(ids, value.ids) {
			t.Fatalf("literal %s Type selection inferred descriptor role: got %v want %v", value.typ, ids, value.ids)
		}
	}
}
