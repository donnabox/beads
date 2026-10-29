package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/clioutput"
	"github.com/steveyegge/beads/resourceget"
)

func definitionHumanFixture(t *testing.T, id string, properties map[string]any) resourceget.DefinitionDocument {
	t.Helper()
	const scope = "https://example.invalid/types/"
	raw, err := json.Marshal(map[string]any{
		"id": scope + id, "revision": "local-2",
		"type":       map[string]any{"uri": "https://github.com/gastownhall/bdp/types/type-system/bead-type", "revision": "bead-type-2"},
		"properties": properties, "ownedLinks": map[string]any{},
	})
	if err != nil {
		t.Fatal(err)
	}
	doc, err := resourceget.NewDefinitionDocument(scope, resourceget.DefinitionBead, raw)
	if err != nil {
		t.Fatal(err)
	}
	return doc
}

func TestDefinitionHumanConceptRequiresCheckedExactOrigin(t *testing.T) {
	const prefix = "https://github.com/gastownhall/bdp/types/type-system/"
	for _, concept := range []string{"type", "bead-type", "link-type", "refines"} {
		t.Run(concept, func(t *testing.T) {
			doc := definitionHumanFixture(t, "opaque/whole-ID", map[string]any{})
			for _, revision := range []string{concept + "-1", concept + "-2"} {
				annotated, err := doc.WithOrigin(prefix+concept, revision)
				if err != nil {
					t.Fatal(err)
				}
				if got := definitionHumanConcept(annotated); got != concept {
					t.Fatalf("exact origin lost: %s", got)
				}
				view, err := renderDefinitionHuman(annotated)
				if err != nil || !strings.Contains(string(view), "Local ID: opaque/whole-ID\n") || !strings.Contains(string(view), "Metatype: "+prefix+"bead-type @ bead-type-2") {
					t.Fatalf("identity/metatype confused: %s %v", view, err)
				}
			}
		})
	}
	for _, origin := range []resourceget.DefinitionOrigin{
		{URI: "https://foreign.example/bead-type", Revision: "bead-type-2"},
		{URI: prefix + "bead-type", Revision: "bead-type-future"},
	} {
		doc := definitionHumanFixture(t, "type-for-bead-type", map[string]any{})
		annotated, err := doc.WithOrigin(origin.URI, origin.Revision)
		if err != nil {
			t.Fatal(err)
		}
		if definitionHumanConcept(annotated) != "" || definitionHumanConcept(doc) != "" {
			t.Fatal("inferred concept from spelling or an unrecognized edition")
		}
	}
}

func TestDefinitionHumanDescriptorShowsOmissionAndExplicitEnvironmentSummary(t *testing.T) {
	doc := definitionHumanFixture(t, "Foo/Child", map[string]any{})
	view, err := renderDefinitionHuman(doc)
	if err != nil || !strings.Contains(string(view), "propertiesSchema: absent") || !strings.Contains(string(view), "inherited constraints still apply") {
		t.Fatalf("omission meaning missing: %s %v", view, err)
	}
	doc = definitionHumanFixture(t, "Foo/Child", map[string]any{
		"propertiesSchema": map[string]any{"type": "object", "description": "literal\u202evalue\u007f\u0085"},
		"schemaResources":  map[string]any{"https://example.invalid/schema": map[string]any{"type": "object", "description": "retained-environment-body"}},
	})
	view, err = renderDefinitionHuman(doc)
	if err != nil {
		t.Fatal(err)
	}
	text := string(view)
	for _, wanted := range []string{"propertiesSchema:\n", "\n      \"type\": \"object\"", "schemaResources: 1 retained resources", "--format detail or --json", "\\u202E", "\\u007F", "\\u0085"} {
		if !strings.Contains(text, wanted) {
			t.Fatalf("missing %q in %s", wanted, text)
		}
	}
	if strings.Contains(text, "retained-environment-body") || strings.ContainsAny(text, "\u202e\u007f\u0085") || strings.Contains(text, "\\u000A") {
		t.Fatalf("unsafe or unintelligible summary: %s", text)
	}
}

func TestDefinitionHumanCollectionAndCanonicalModesStaySeparate(t *testing.T) {
	const scope = "https://example.invalid/types/"
	doc := definitionHumanFixture(t, "Foo/Whole", map[string]any{})
	view, err := renderDefinitionHumanCollection(scope, []resourceget.DefinitionDocument{doc})
	if err != nil || !strings.Contains(string(view), "Local ID: Foo/Whole") || !strings.Contains(string(view), "1 definitions (complete selection)") || strings.Contains(string(view), "Descriptor:") {
		t.Fatalf("collection: %s %v", view, err)
	}
	if _, err := renderDefinitionHumanCollection("https://foreign.example/", []resourceget.DefinitionDocument{doc}); err == nil {
		t.Fatal("displayed a document outside the selected Scope")
	}
	for _, options := range []clioutput.Options{{JSON: true}, {Formats: []string{"json"}}, {Formats: []string{"jsonl"}}, {Formats: []string{"detail"}}, {Include: []string{"/properties"}}, {Omit: []string{"/metadata"}}} {
		if definitionHumanView(options) {
			t.Fatalf("changed explicit canonical mode: %+v", options)
		}
	}
	if !definitionHumanView(clioutput.Options{}) || !definitionHumanView(clioutput.Options{Formats: []string{"default"}}) {
		t.Fatal("ordinary human output did not select readable view")
	}
}

func TestDefinitionHumanRefinesRetainsExactCapturedTarget(t *testing.T) {
	doc := definitionHumanFixture(t, "Child/Whole", map[string]any{})
	var record map[string]any
	if err := json.Unmarshal(doc.NativeJSON(), &record); err != nil {
		t.Fatal(err)
	}
	const refines = "https://github.com/gastownhall/bdp/types/type-system/refines"
	const target = "https://foreign.example/types/Parent/Whole"
	record["ownedLinks"] = map[string]any{refines: []any{map[string]any{
		"id": doc.ScopeURL() + "Owned/Declaration", "revision": "owned-3",
		"type":       map[string]any{"uri": refines, "revision": "refines-2"},
		"source":     map[string]any{"uri": doc.ID(), "revision": "owner-capture-1"},
		"target":     map[string]any{"uri": target, "revision": "parent-capture-4"},
		"properties": map[string]any{},
	}}}
	raw, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	doc, err = resourceget.NewDefinitionDocument(doc.ScopeURL(), resourceget.DefinitionBead, raw)
	if err != nil {
		t.Fatal(err)
	}
	view, err := renderDefinitionHuman(doc)
	if err != nil || !strings.Contains(string(view), "Refines: "+target+" @ parent-capture-4") || !strings.Contains(string(view), "Owned/Declaration @ owned-3") {
		t.Fatalf("changed retained Refines identity: %s %v", view, err)
	}
	if got := definitionHumanText("safe\x1b[31m\u2066text"); got != "safe\\u001B[31m\\u2066text" {
		t.Fatalf("unsafe display: %q", got)
	}
}
