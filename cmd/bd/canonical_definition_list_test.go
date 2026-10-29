package main

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/steveyegge/beads/internal/clioutput"
)

func TestIntrinsicDefinitionCollectionPreservesExactRecords(t *testing.T) {
	const scope = "https://example.invalid/definitions/"
	raw := json.RawMessage(`{"id":"https://example.invalid/definitions/local-type","revision":"bead-r1","type":{"uri":"https://github.com/gastownhall/bdp/types/type-system/bead-type","revision":"category-r1"},"properties":{"propertiesSchema":{"type":"object"}},"ownedLinks":{"https://github.com/gastownhall/bdp/types/type-system/refines":[{"id":"https://example.invalid/definitions/edge","revision":"link-r1","type":{"uri":"https://github.com/gastownhall/bdp/types/type-system/refines","revision":"refines-r1"},"source":"https://example.invalid/definitions/local-type","target":{"uri":"https://example.invalid/definitions/parent","revision":"parent-r1"},"properties":{}}]}}`)
	descriptor, err := intrinsicDefinitionCollectionDescriptor(scope, []json.RawMessage{raw})
	if err != nil {
		t.Fatal(err)
	}
	renderer, err := clioutput.NewCollection(descriptor, clioutput.Options{Formats: []string{"json"}})
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(map[string]any{"items": []json.RawMessage{raw}, "next": nil})
	if err != nil {
		t.Fatal(err)
	}
	exhausted := true
	output, err := renderer.Render(payload, clioutput.Completion{Returned: 1, Exhausted: &exhausted, StopReason: "exhausted"})
	if err != nil {
		t.Fatal(err)
	}
	var envelope struct {
		Spec string          `json:"bdp_spec"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(output, &envelope); err != nil || envelope.Spec != intrinsicDefinitionContract {
		t.Fatalf("wrong definition collection contract: %v %s", err, output)
	}
	var before, after any
	if json.Unmarshal(payload, &before) != nil || json.Unmarshal(envelope.Data, &after) != nil || !reflect.DeepEqual(before, after) {
		t.Fatalf("collection changed full records or exact affiliations: %s", output)
	}
	for _, omit := range []string{"/type/revision", "/ownedLinks/https:~1~1github.com~1gastownhall~1bdp~1types~1type-system~1refines/0/target/revision"} {
		r, err := clioutput.NewCollection(descriptor, clioutput.Options{Formats: []string{"json"}, Omit: []string{omit}})
		if errors.Is(err, clioutput.ErrInvalidProjection) {
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		if _, err := r.Render(payload, clioutput.Completion{Returned: 1, Exhausted: &exhausted, StopReason: "exhausted"}); !errors.Is(err, clioutput.ErrInvalidProjection) {
			t.Fatalf("mask removed exact context %s: %v", omit, err)
		}
	}
	legacy := json.RawMessage(`{"id":"https://example.invalid/definitions/local-type","revision":"r1","type":"https://example.invalid/definitions/type","properties":{},"ownedLinks":{}}`)
	if _, err := intrinsicDefinitionCollectionDescriptor(scope, []json.RawMessage{legacy}); err == nil {
		t.Fatal("definition collection admitted legacy string-Type projection")
	}
}
