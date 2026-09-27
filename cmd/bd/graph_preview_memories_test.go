package main

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func TestGraphMemoryDiscoveryRendering(t *testing.T) {
	result := graphMemoryDiscoveryResult{Projection: "summary", Scope: "https://example.invalid/", Complete: true,
		Items: []graphMemoryDiscoveryItem{{ID: "https://example.invalid/beads/plan", Title: "Plan\n\x1b[31m", Version: "saved'quoted",
			Attribution:   graphstore.Attribution{Actor: "actor\nname", Status: "claimed", RecordedAt: "observed"},
			MatchedFields: []string{"body"}, Excerpt: &graphMemoryDiscoveryExcerpt{Field: "body", Text: "line\nbody", Truncated: true},
			Details: &graphMemoryDiscoveryDetails{OwnedLinkCount: 2}}}}
	human, err := renderGraphMemoryDiscovery(result, false, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, unsafe := range []string{"Plan\n", "\x1b", "actor\nname", "line\nbody"} {
		if strings.Contains(human, unsafe) {
			t.Fatalf("unescaped data in human output: %q", human)
		}
	}
	if !strings.Contains(human, "--version 'saved'\"'\"'quoted'") || !strings.Contains(human, "Owned Links: 2") {
		t.Fatalf("missing exact recall or details: %s", human)
	}
	structured, err := renderGraphMemoryDiscovery(result, true, true)
	if err != nil {
		t.Fatal(err)
	}
	var envelope struct {
		SchemaVersion int                        `json:"schemaVersion"`
		Preview       bool                       `json:"preview"`
		Result        graphMemoryDiscoveryResult `json:"result"`
	}
	if err := json.Unmarshal([]byte(structured), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.SchemaVersion != 1 || !envelope.Preview || !envelope.Result.Complete || envelope.Result.Next != nil || envelope.Result.Items[0].Title != result.Items[0].Title {
		t.Fatalf("structured summary changed or disappeared under quiet: %+v", envelope)
	}
	quiet, err := renderGraphMemoryDiscovery(result, false, true)
	if err != nil || quiet != "" {
		t.Fatalf("quiet human output: %q, %v", quiet, err)
	}
}

func TestGraphMemoryDiscoveryRenderedByteLimit(t *testing.T) {
	result := graphMemoryDiscoveryResult{Items: []graphMemoryDiscoveryItem{{Title: strings.Repeat("\x00", graphMemoryDiscoveryOutputLimit/4)}}}
	for _, structured := range []bool{false, true} {
		output, err := renderGraphMemoryDiscovery(result, structured, false)
		if !errors.Is(err, graphstore.ErrLimitExceeded) || output != "" {
			t.Fatalf("oversized escaped output must refuse atomically: structured=%v bytes=%d err=%v", structured, len(output), err)
		}
	}
}
