package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func TestGraphPreviewAttributionBasisPreservesStoredStatusAndUserProperties(t *testing.T) {
	record := graphstore.Record{
		ID:       "https://example.test/beads/one",
		Type:     "https://example.test/types/memory",
		Version:  "v1",
		Revision: "v1",
		Properties: graphstore.Properties{
			Title: "One",
			Body:  `{"attribution":{"status":"claimed"}}`,
		},
		Attribution: graphstore.Attribution{Actor: "agent:writer", Status: "claimed", RecordedAt: "2026-10-09T00:00:00Z"},
	}
	stored, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(stored), `"status":"claimed"`) {
		t.Fatalf("stored snapshot unexpectedly changed: %s", stored)
	}
	projected, err := graphProjectCompleteRecords(record)
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]any
	if err := json.Unmarshal(projected, &result); err != nil {
		t.Fatal(err)
	}
	attribution := result["attribution"].(map[string]any)
	if attribution["actor"] != "agent:writer" || attribution["basis"] != "writer-supplied" {
		t.Fatalf("wrong public attribution: %v", attribution)
	}
	if _, old := attribution["status"]; old {
		t.Fatalf("old status leaked into public output: %v", attribution)
	}
	if record.Attribution.Status != "claimed" || result["properties"].(map[string]any)["body"] != record.Properties.Body {
		t.Fatal("projection changed stored state or user content")
	}
}

func TestGraphPreviewAbsentAttributionHasNoPrincipal(t *testing.T) {
	for _, attribution := range []any{graphstore.Attribution{Status: "unknown"}, graphstore.Attribution{}, nil} {
		projected, err := graphProjectCompleteRecords(map[string]any{
			"attribution": attribution,
			"properties":  map[string]any{"attribution": map[string]any{"status": "claimed"}},
		})
		if err != nil {
			t.Fatal(err)
		}
		var result map[string]any
		if err := json.Unmarshal(projected, &result); err != nil {
			t.Fatal(err)
		}
		if _, present := result["attribution"]; present {
			t.Fatalf("principal-free attribution should be absent: %s", projected)
		}
		if result["properties"].(map[string]any)["attribution"].(map[string]any)["status"] != "claimed" {
			t.Fatalf("user property was transformed: %s", projected)
		}
	}
}

func TestGraphPreviewNativeIssueVersionAttributionLabelUnchanged(t *testing.T) {
	projected, err := graphProjectCompleteRecords(map[string]any{"attribution": "native-issue-writer"})
	if err != nil {
		t.Fatal(err)
	}
	if string(projected) != `{"attribution":"native-issue-writer"}` {
		t.Fatalf("native Issue version attribution changed: %s", projected)
	}
}
