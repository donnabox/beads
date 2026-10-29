package main

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

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
