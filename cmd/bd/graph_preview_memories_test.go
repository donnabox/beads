package main

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func TestGraphPreviewMemoryDiscoveryRendering(t *testing.T) {
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
	if envelope.SchemaVersion != 1 || !envelope.Preview || !envelope.Result.Complete || envelope.Result.Next != nil || !reflect.DeepEqual(envelope.Result, result) {
		t.Fatalf("structured summary changed or disappeared under quiet: %+v", envelope)
	}
	quiet, err := renderGraphMemoryDiscovery(result, false, true)
	if err != nil || quiet != "" {
		t.Fatalf("quiet human output: %q, %v", quiet, err)
	}
}

func TestGraphPreviewMemoryDiscoveryRenderedByteLimit(t *testing.T) {
	result := graphMemoryDiscoveryResult{Items: []graphMemoryDiscoveryItem{{Title: strings.Repeat("\x00", graphMemoryDiscoveryOutputLimit/4)}}}
	for _, structured := range []bool{false, true} {
		output, err := renderGraphMemoryDiscovery(result, structured, false)
		if !errors.Is(err, graphstore.ErrLimitExceeded) || output != "" {
			t.Fatalf("oversized escaped output must refuse atomically: structured=%v bytes=%d err=%v", structured, len(output), err)
		}
	}
}

func TestGraphPreviewMemoryDiscoveryCompactRendering(t *testing.T) {
	result := graphMemoryDiscoveryResult{Scope: "https://example.invalid/project/", Items: []graphMemoryDiscoveryItem{
		{ID: "https://example.invalid/project/beads/policy", Title: `Code "flow" policy — 雪`, Version: "saved-version", Attribution: graphstore.Attribution{Actor: "actor-secret", Status: "claimed", RecordedAt: "timestamp-secret"}},
		{ID: "https://example.invalid/project/beads/tests", Title: "Testing policy"},
	}}
	output, err := renderGraphMemoryDiscovery(result, false, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Memories (2):\n\n", "  beads/policy  Code \"flow\" policy — 雪\n", "  beads/tests  Testing policy\n", "bd recall <id>", "--details"} {
		if !strings.Contains(output, want) {
			t.Fatalf("missing compact content %q in %q", want, output)
		}
	}
	for _, unwanted := range []string{result.Scope, "saved-version", "actor-secret", "timestamp-secret", "Attribution:", "Owned Links:", "--version", "Matched:"} {
		if strings.Contains(output, unwanted) {
			t.Fatalf("default output exposed details %q: %q", unwanted, output)
		}
	}
}

func TestGraphPreviewMemoryDiscoverySearchRendering(t *testing.T) {
	body := strings.Repeat("leading text ", 20) + "MATCH\n\x1b[31m" + strings.Repeat(" trailing text", 20)
	excerpt := graphMemoryBodyExcerpt(body, strings.Index(body, "MATCH"))
	result := graphMemoryDiscoveryResult{Scope: "https://example.invalid/", Items: []graphMemoryDiscoveryItem{{ID: "https://example.invalid/beads/search", Title: "Search result", MatchedFields: []string{"body"}, Excerpt: excerpt}}}
	output, err := renderGraphMemoryDiscovery(result, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, "MATCH\\n\\x1b[31m") || !strings.Contains(output, "(excerpt, shortened)") || strings.Contains(output, body) || strings.Contains(output, "\x1b") || len(output) > 512 {
		t.Fatalf("search excerpt must be bounded, labeled and escaped: %q", output)
	}
	// An unshortened matching body remains explicitly identified as an excerpt.
	result.Items[0].Excerpt = &graphMemoryDiscoveryExcerpt{Field: "body", Text: "Short match"}
	output, err = renderGraphMemoryDiscovery(result, false, false)
	if err != nil || !strings.Contains(output, "    Short match (excerpt)\n") || strings.Contains(output, "shortened") {
		t.Fatalf("short excerpt: %q (%v)", output, err)
	}
}

func TestGraphPreviewMemoryDiscoveryTerminalEscaping(t *testing.T) {
	result := graphMemoryDiscoveryResult{Scope: "https://example.invalid/", Items: []graphMemoryDiscoveryItem{{
		ID: "https://example.invalid/beads/id\r\t\x1b[2J", Title: "Policy\n\u202eend", Version: "version\x1b[31m\n",
		Attribution: graphstore.Attribution{Actor: "actor\x00", RecordedAt: "time\r"}, Details: &graphMemoryDiscoveryDetails{},
	}}}
	output, err := renderGraphMemoryDiscovery(result, false, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, unsafe := range []string{"\x1b", "\r", "\t", "\x00", "\u202e", "Policy\n", "version\x1b"} {
		if strings.Contains(output, unsafe) {
			t.Fatalf("unsafe terminal character %q in %q", unsafe, output)
		}
	}
	for _, escaped := range []string{`beads/id\r\t\x1b[2J`, `Policy\n\u202eend`, `version\x1b[31m\n`} {
		if !strings.Contains(output, escaped) {
			t.Fatalf("data vanished instead of being escaped: %q in %q", escaped, output)
		}
	}
}
