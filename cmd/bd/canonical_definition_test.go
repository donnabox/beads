package main

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/clioutput"
	"github.com/steveyegge/beads/resourceget"
)

func TestDefinitionShowPreservesExactAffiliations(t *testing.T) {
	const scope = "https://example.invalid/definitions/"
	const typeURI = "https://github.com/gastownhall/bdp/types/type-system/refines"
	raw := []byte(`{"id":"https://example.invalid/definitions/local-type","revision":"bead-r1","type":{"uri":"https://github.com/gastownhall/bdp/types/type-system/bead-type","revision":"category-r1"},"properties":{"propertiesSchema":{"type":"object"}},"metadata":{},"ownedLinks":{"https://github.com/gastownhall/bdp/types/type-system/refines":[{"id":"https://example.invalid/definitions/edge","revision":"link-r1","type":{"uri":"https://github.com/gastownhall/bdp/types/type-system/refines","revision":"refines-r1"},"source":{"uri":"https://example.invalid/definitions/local-type","revision":"bead-r1"},"target":{"uri":"https://example.invalid/definitions/parent","revision":"parent-r1"},"properties":{},"metadata":{}}]}}`)
	document, err := resourceget.NewDefinitionDocument(scope, resourceget.DefinitionBead, raw)
	if err != nil {
		t.Fatal(err)
	}
	descriptor := canonicalShowDescriptor(document.ID(), document)
	renderer, err := clioutput.New(descriptor, clioutput.Options{Formats: []string{"json"}})
	if err != nil {
		t.Fatal(err)
	}
	emitted, err := renderer.Render(document.NativeJSON())
	if err != nil {
		t.Fatal(err)
	}
	var envelope struct {
		Data json.RawMessage `json:"data"`
		Spec string          `json:"bdp_spec"`
	}
	if err := json.Unmarshal(emitted, &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Spec != intrinsicDefinitionContract {
		t.Fatalf("definition mislabeled with legacy wire contract: %s", envelope.Spec)
	}
	var expected, actual any
	if json.Unmarshal(raw, &expected) != nil || json.Unmarshal(envelope.Data, &actual) != nil {
		t.Fatal("definition output is not a complete Resource")
	}
	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("exact affiliations or owned capture lost: %s", envelope.Data)
	}
	// A mask cannot remove an exact affiliation or an owned Link endpoint pin.
	for _, pointer := range []string{"/type/revision", "/ownedLinks/" + strings.ReplaceAll(typeURI, "/", "~1") + "/0/target/revision"} {
		r, err := clioutput.New(descriptor, clioutput.Options{Formats: []string{"json"}, Omit: []string{pointer}})
		if errors.Is(err, clioutput.ErrInvalidProjection) {
			continue // Invalid masks may be rejected before any record is acquired.
		}
		if err != nil {
			t.Fatal(err)
		}
		if _, err := r.Render(document.NativeJSON()); !errors.Is(err, clioutput.ErrInvalidProjection) {
			t.Fatalf("omitting exact context %s: %v", pointer, err)
		}
	}
}

func TestCommonResourceShowRetainsInstanceAndLinkPins(t *testing.T) {
	const scope = "https://example.invalid/resources/"
	for _, tc := range []struct {
		name      string
		kind      graph.ResourceKind
		raw       string
		protected []string
	}{
		{"instance", graph.KindBead,
			`{"id":"https://example.invalid/resources/333","revision":"bead-r1","type":{"uri":"https://example.invalid/resources/task","revision":"task-r1"},"properties":{"title":"Hello world","propertiesSchema":{"type":"object"}},"ownedLinks":{"https://example.invalid/resources/blocks":[{"id":"https://example.invalid/resources/edge","revision":"link-r1","type":{"uri":"https://example.invalid/resources/blocks","revision":"blocks-r1"},"source":{"uri":"https://example.invalid/resources/333","revision":"bead-r1"},"target":{"uri":"https://example.invalid/resources/other","revision":"other-r1"},"properties":{}}]}}`,
			[]string{"/type/revision", "/ownedLinks/https:~1~1example.invalid~1resources~1blocks/0/target/revision"}},
		{"link", graph.KindLink,
			`{"id":"https://example.invalid/resources/edge","revision":"link-r1","type":{"uri":"https://example.invalid/resources/blocks","revision":"blocks-r1"},"source":{"uri":"https://example.invalid/resources/333","revision":"bead-r1"},"target":{"uri":"https://example.invalid/resources/other","revision":"other-r1"},"properties":{}}`,
			[]string{"/type/revision", "/source/revision", "/target/revision"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			doc, err := resourceget.NewResourceDocument(scope, tc.kind, []byte(tc.raw))
			if err != nil {
				t.Fatal(err)
			}
			// A property called propertiesSchema does not make an instance a
			// Type definition. Only the checked reader's conversion carries it.
			if _, definition := canonicalDefinitionForDisplay(doc); definition {
				t.Fatal("ordinary Resource inferred to be a Type definition")
			}
			descriptor := canonicalShowDescriptor(doc.ID(), doc)
			renderer, err := clioutput.New(descriptor, clioutput.Options{Formats: []string{"json"}})
			if err != nil {
				t.Fatal(err)
			}
			emitted, err := renderer.Render(doc.NativeJSON())
			if err != nil {
				t.Fatal(err)
			}
			var envelope struct {
				Data json.RawMessage `json:"data"`
				Spec string          `json:"bdp_spec"`
			}
			if err := json.Unmarshal(emitted, &envelope); err != nil {
				t.Fatal(err)
			}
			var want, got any
			if err := json.Unmarshal([]byte(tc.raw), &want); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(envelope.Data, &got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(want, got) || envelope.Spec != intrinsicDefinitionContract {
				t.Fatalf("ordinary Resource context lost or mislabeled: %s", emitted)
			}
			for _, pointer := range tc.protected {
				r, err := clioutput.New(descriptor, clioutput.Options{Formats: []string{"json"}, Omit: []string{pointer}})
				if errors.Is(err, clioutput.ErrInvalidProjection) {
					continue
				}
				if err != nil {
					t.Fatal(err)
				}
				if _, err := r.Render(doc.NativeJSON()); !errors.Is(err, clioutput.ErrInvalidProjection) {
					t.Fatalf("ordinary context %s was removed: %v", pointer, err)
				}
			}
		})
	}
}

func TestCommonResourceDefinitionDisplayKeepsCheckedOrigin(t *testing.T) {
	definition := definitionHumanFixture(t, "bead-type", map[string]any{})
	definition, err := definition.WithOrigin("https://github.com/gastownhall/bdp/types/type-system/bead-type", "bead-type-2")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := definition.Resource()
	if err != nil {
		t.Fatal(err)
	}
	display, ok := canonicalDefinitionForDisplay(doc)
	if !ok || definitionHumanConcept(display) != "bead-type" || !reflect.DeepEqual(doc.NativeJSON(), definition.NativeJSON()) {
		t.Fatal("common read lost checked provenance or changed its Resource")
	}
}
