package resourceget

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

const definitionScope = "https://example.test/work/"

func TestDefinitionOriginIsImmutableAndOutsideResourceJSON(t *testing.T) {
	doc, err := NewDefinitionDocument(definitionScope, DefinitionBead, definitionBytes(t, definitionFixture(t)))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := doc.Origin(); ok {
		t.Fatal("ordinary carrier invented intrinsic provenance")
	}
	const uri = "https://github.com/gastownhall/bdp/types/type-system/type"
	annotated, err := doc.WithOrigin(uri, "type-2")
	if err != nil {
		t.Fatal(err)
	}
	origin, ok := annotated.Origin()
	if !ok || origin.URI != uri || origin.Revision != "type-2" {
		t.Fatalf("lost exact provenance: %+v", origin)
	}
	origin.URI = "https://foreign.example/type"
	if original, _ := annotated.Origin(); original.URI != uri {
		t.Fatal("returned provenance mutated carrier")
	}
	if _, ok := doc.Origin(); ok {
		t.Fatal("annotation mutated original document")
	}
	before, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	after, err := json.Marshal(annotated)
	if err != nil || !bytes.Equal(before, after) || !bytes.Equal(doc.NativeJSON(), annotated.NativeJSON()) {
		t.Fatalf("display provenance changed Resource JSON: %v", err)
	}
	if _, err := annotated.WithOrigin(uri, "type-2"); err != nil {
		t.Fatal(err)
	}
	if _, err := annotated.WithOrigin(uri, "type-1"); err == nil {
		t.Fatal("replaced checked origin with another edition")
	}
	for _, input := range []DefinitionOrigin{{URI: "type", Revision: "type-2"}, {URI: uri}, {URI: uri, Revision: string([]byte{0xff})}} {
		if _, err := doc.WithOrigin(input.URI, input.Revision); err == nil {
			t.Fatalf("accepted incomplete origin: %+v", input)
		}
	}
	if _, err := (DefinitionDocument{}).WithOrigin(uri, "type-2"); err == nil {
		t.Fatal("annotated uninitialized document")
	}
}

func definitionFixture(t *testing.T) map[string]any {
	t.Helper()
	return map[string]any{
		"id":         definitionScope + "Task",
		"type":       map[string]any{"uri": "https://types.example/BeadType", "revision": "global-metatype-7"},
		"revision":   "local-task-3",
		"properties": map[string]any{"title": "Task", "schema": map[string]any{"$id": "https://types.example/task-schema", "type": "object"}},
		"metadata":   map[string]any{},
		"ownedLinks": map[string]any{
			"https://types.example/Refines": []any{map[string]any{
				"id": definitionScope + "Task-refines-Issue", "revision": "local-refines-2",
				"type":       map[string]any{"uri": "https://types.example/Refines", "revision": "refines-definition-4"},
				"source":     map[string]any{"uri": definitionScope + "Task", "revision": "earlier-owner-1"},
				"target":     map[string]any{"uri": "https://types.example/Issue", "revision": "issue-definition-5"},
				"properties": map[string]any{}, "metadata": map[string]any{},
			}},
		},
	}
}

func definitionBytes(t *testing.T, value any) []byte {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestDefinitionDocumentPreservesExactAffiliationsAndEndpoints(t *testing.T) {
	native := definitionBytes(t, definitionFixture(t))
	doc, err := NewDefinitionDocument(definitionScope, DefinitionBead, native)
	if err != nil {
		t.Fatal(err)
	}
	if doc.ID() != definitionScope+"Task" || doc.Revision() != "local-task-3" || doc.ScopeURL() != definitionScope || doc.Kind() != DefinitionBead {
		t.Fatalf("incorrect identity accessors: %+v", doc)
	}
	for _, literal := range []string{"global-metatype-7", "local-refines-2", "refines-definition-4", "earlier-owner-1", "issue-definition-5", "https://types.example/task-schema"} {
		if !bytes.Contains(doc.NativeJSON(), []byte(literal)) {
			t.Fatalf("lost recorded %q", literal)
		}
	}
	encoded, err := json.Marshal(doc)
	if err != nil || !bytes.Equal(encoded, doc.NativeJSON()) {
		t.Fatalf("marshal changed carrier: %s %v", encoded, err)
	}
	for i := range native {
		native[i] = 'x'
	}
	exposed := doc.NativeJSON()
	exposed[0] = 'x'
	if doc.NativeJSON()[0] != '{' {
		t.Fatal("document retained mutable input/output bytes")
	}
}

func TestDefinitionDocumentStandaloneLinkAndRecordedEmptyBucket(t *testing.T) {
	fixture := definitionFixture(t)
	link := fixture["ownedLinks"].(map[string]any)["https://types.example/Refines"].([]any)[0]
	doc, err := NewDefinitionDocument(definitionScope, DefinitionLink, definitionBytes(t, link))
	if err != nil || doc.Kind() != DefinitionLink {
		t.Fatalf("standalone Link: %+v %v", doc, err)
	}
	fixture["ownedLinks"] = map[string]any{"https://types.example/Refines": []any{}}
	doc, err = NewDefinitionDocument(definitionScope, DefinitionBead, definitionBytes(t, fixture))
	if err != nil || !bytes.Contains(doc.NativeJSON(), []byte(`"https://types.example/Refines":[]`)) {
		t.Fatalf("rewrote retained empty bucket: %s %v", doc.NativeJSON(), err)
	}
}

func TestDefinitionDocumentPreservesOptionalMetadataAbsence(t *testing.T) {
	fixture := definitionFixture(t)
	delete(fixture, "metadata")
	link := fixture["ownedLinks"].(map[string]any)["https://types.example/Refines"].([]any)[0].(map[string]any)
	delete(link, "metadata")
	for _, tc := range []struct {
		kind DefinitionKind
		body map[string]any
	}{{DefinitionBead, fixture}, {DefinitionLink, link}} {
		t.Run(string(tc.kind), func(t *testing.T) {
			native := definitionBytes(t, tc.body)
			doc, err := NewDefinitionDocument(definitionScope, tc.kind, native)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(native, doc.NativeJSON()) || bytes.Contains(doc.NativeJSON(), []byte(`"metadata"`)) {
				t.Fatalf("changed absent metadata: %s", doc.NativeJSON())
			}
		})
	}
}

func TestDefinitionDocumentRejectsWrongOrIncompleteCarrier(t *testing.T) {
	cases := []struct {
		name string
		edit func(map[string]any)
	}{
		{"unpinned-own-Type", func(v map[string]any) { v["type"] = "https://types.example/BeadType" }},
		{"missing-own-Type-pin", func(v map[string]any) { v["type"] = map[string]any{"uri": "https://types.example/BeadType"} }},
		{"foreign-ID", func(v map[string]any) { v["id"] = "https://other.example/work/Task" }},
		{"different-Scope-prefix", func(v map[string]any) { v["id"] = "https://example.test/work-extra/Task" }},
		{"missing-revision", func(v map[string]any) { delete(v, "revision") }},
		{"null-properties", func(v map[string]any) { v["properties"] = nil }},
		{"null-metadata", func(v map[string]any) { v["metadata"] = nil }},
		{"array-metadata", func(v map[string]any) { v["metadata"] = []any{} }},
		{"missing-owned-plane", func(v map[string]any) { delete(v, "ownedLinks") }},
		{"null-owned-plane", func(v map[string]any) { v["ownedLinks"] = nil }},
		{"incident-aggregate", func(v map[string]any) { v["links"] = map[string]any{"items": []any{}, "next": nil} }},
		{"legacy-attribution", func(v map[string]any) { v["attribution"] = map[string]any{"principal": "agent:a", "basis": "unknown"} }},
		{"mapping-extra", func(v map[string]any) { v["definitionContext"] = "private-context" }},
		{"invalid-provenance", func(v map[string]any) { v["revisionProvenance"] = map[string]any{} }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := definitionFixture(t)
			tc.edit(fixture)
			if _, err := NewDefinitionDocument(definitionScope, DefinitionBead, definitionBytes(t, fixture)); err == nil {
				t.Fatal("accepted invalid carrier")
			}
		})
	}
	if _, err := NewDefinitionDocument(definitionScope, "unknown", definitionBytes(t, definitionFixture(t))); err == nil {
		t.Fatal("accepted unknown Resource kind")
	}
	if _, err := (DefinitionDocument{}).MarshalJSON(); err == nil {
		t.Fatal("serialized uninitialized document")
	}
}

func TestDefinitionOwnedLinksAreCompleteUniqueAndOwnerMatched(t *testing.T) {
	for _, test := range []string{"unpinned-Type", "wrong-source", "wrong-bucket", "duplicate-ID", "foreign-ID", "null-metadata", "partial-endpoint"} {
		t.Run(test, func(t *testing.T) {
			fixture := definitionFixture(t)
			owned := fixture["ownedLinks"].(map[string]any)
			link := owned["https://types.example/Refines"].([]any)[0].(map[string]any)
			switch test {
			case "unpinned-Type":
				link["type"] = "https://types.example/Refines"
			case "wrong-source":
				link["source"] = definitionScope + "other"
			case "wrong-bucket":
				link["type"] = map[string]any{"uri": "https://types.example/Other", "revision": "other-1"}
			case "duplicate-ID":
				owned["https://types.example/Refines"] = []any{link, link}
			case "foreign-ID":
				link["id"] = "https://other.example/link-1"
			case "null-metadata":
				link["metadata"] = nil
			case "partial-endpoint":
				link["target"] = map[string]any{"uri": "https://types.example/Issue"}
			}
			if _, err := NewDefinitionDocument(definitionScope, DefinitionBead, definitionBytes(t, fixture)); err == nil {
				t.Fatal("accepted invalid owned Link")
			}
		})
	}
}

func TestDefinitionProvenanceRetainsAbsenceAndValueStates(t *testing.T) {
	fixture := definitionFixture(t)
	fixture["revisionProvenance"] = map[string]any{
		"actor":       map[string]any{"id": "agent:trish", "basis": "unknown"},
		"committedAt": map[string]any{"state": "present", "value": "2026-10-28T20:15:00Z"},
		"agent":       map[string]any{"state": "absent"},
		"message":     map[string]any{"state": "present", "value": ""},
	}
	doc, err := NewDefinitionDocument(definitionScope, DefinitionBead, definitionBytes(t, fixture))
	if err != nil || !bytes.Contains(doc.NativeJSON(), []byte(`"message":{"state":"present","value":""}`)) {
		t.Fatalf("lost provenance states: %s %v", doc.NativeJSON(), err)
	}
	for _, raw := range []string{
		`{"actor":{"id":"agent:a","basis":"authenticated"}}`,
		`{"committedAt":{"state":"undetermined"}}`,
		`{"committedAt":{"state":"absent"},"agent":{"state":"absent"},"message":{"state":"absent"}}`,
		`{"committedAt":{"state":"undetermined"},"agent":{"state":"present","value":""},"message":{"state":"absent"}}`,
		`{"actor":{"id":"agent:a","basis":"unknown"},"extra":true}`,
	} {
		fixture["revisionProvenance"] = json.RawMessage(raw)
		if _, err := NewDefinitionDocument(definitionScope, DefinitionBead, definitionBytes(t, fixture)); err == nil {
			t.Fatalf("accepted provenance %s", raw)
		}
	}
}

func TestDefinitionInputUsesStrictCommonJSON(t *testing.T) {
	raw := string(definitionBytes(t, definitionFixture(t)))
	for _, invalid := range []string{
		strings.Replace(raw, `"local-task-3"`, `"local-task-3","revision":"other"`, 1),
		strings.Replace(raw, `"Task"`, `"\ud800"`, 1),
		strings.Replace(raw, `"Task"`, `1e999`, 1),
	} {
		if _, err := NewDefinitionDocument(definitionScope, DefinitionBead, []byte(invalid)); err == nil {
			t.Fatalf("accepted noncommon JSON %s", invalid)
		}
	}
}

func TestDefinitionDateTimeDoesNotNormalizeRecordedContext(t *testing.T) {
	for _, value := range []string{"2026-10-28T20:15:00Z", "2016-12-31T23:59:60Z", "2017-01-01T01:59:60+02:00"} {
		if !definitionDateTime(value) {
			t.Errorf("rejected date-time %s", value)
		}
	}
	for _, value := range []string{"2026-10-28t20:15:00z", "2026-02-30T20:15:00Z", "2026-10-28T24:00:00Z", "2026-10-28T20:15:00+02", "2026-10-28T20:15:00+0200", "2026-10-28T20:15:00+24:00", "2026-10-28T23:99:60+00:40", "2026-10-28T20:15:60Z"} {
		if definitionDateTime(value) {
			t.Errorf("accepted invalid native date-time %s", value)
		}
	}
}
