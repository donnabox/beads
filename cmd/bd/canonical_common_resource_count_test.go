package main

import (
	"encoding/json"
	"testing"

	graph "github.com/steveyegge/beads/graphops"
)

func TestCanonicalCommonResourceCountPrecedesDeliveryLimit(t *testing.T) {
	capture := commonGraphCapture(t)
	records := []json.RawMessage{}
	for _, document := range capture.Documents {
		if document.Kind() == graph.KindBead {
			records = append(records, document.NativeJSON())
		}
	}
	if len(records) < 2 {
		t.Fatal("mixed capture must exceed the requested delivery limit")
	}
	cmd, out, _ := canonicalCollectionTestCommand(t, false, "--output-format=json", "--count-only", "--limit=1")
	plan, failure := canonicalCollectionTestPlan(t, cmd, false)
	plan.resourceScope = "https://example.test/graph/"
	if err := renderCanonicalCollection(cmd, plan, records, failure); err != nil {
		t.Fatal(err)
	}
	var envelope struct {
		Data struct {
			Count int `json:"count"`
		} `json:"data"`
	}
	if err := json.Unmarshal(out.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Data.Count != len(records) {
		t.Fatalf("common-resource count truncated or omitted records: %s", out)
	}
}
