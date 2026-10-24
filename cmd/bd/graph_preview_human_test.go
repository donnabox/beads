package main

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
)

func TestGraphHumanIdentityRowsAndDocuments(t *testing.T) {
	const scope = "https://example.invalid/scope/"
	task, _ := graphstore.NativeIssueTypeURL(scope, "task")
	issue := graphstore.IssueRecord{ID: scope + "beads/work", Type: task, Revision: "full-token", Properties: &types.Issue{Title: "Work", Description: "Selected body", Status: types.StatusOpen, Priority: 2}}
	if got := graphHumanRow(scope, issue); got != `beads/work [task] open P2 "Work"` {
		t.Fatal(got)
	}
	link := graphstore.LinkRecord{ID: scope + "links/L", Type: scope + "types/references", Source: issue.ID, Target: "https://foreign.invalid/scope/beads/m", TargetRevision: "old-token", Revision: "link-token"}
	want := "links/L [types/references] beads/work → https://foreign.invalid/scope/beads/m@=old-token"
	if got := graphHumanRow(scope, link); got != want {
		t.Fatal(got)
	}
	for _, retained := range []bool{false, true} {
		got := graphHumanDocument(scope, link, "@4", retained, false)
		if !strings.HasPrefix(got, want+"\nRevision: @4") || !strings.Contains(got, "To: https://foreign.invalid/scope/beads/m@=old-token") || strings.Contains(got, "link-token") || strings.Contains(got, "Attribution") {
			t.Fatal(got)
		}
	}
	got := graphHumanDocument(scope, issue, "@1", true, true)
	if !strings.Contains(got, "retained snapshot") || !strings.Contains(got, "Selected body") || !strings.Contains(got, "Exact revision token\nfull-token") {
		t.Fatal(got)
	}
	// URI prefix coincidence is not Scope membership.
	if got := graphHumanID(scope, "https://example.invalid/scope-other/x"); got != "https://example.invalid/scope-other/x" {
		t.Fatal(got)
	}
}

func TestGraphHumanTimelineKeepsDeletionSequence(t *testing.T) {
	rows := []graphstore.VersionRow{{Ordinal: 3, Version: "restored", ChangeAt: time.Unix(3, 0)}, {Ordinal: 2, Version: "deletion-token-long", Removed: true, ChangeAt: time.Unix(2, 0)}, {Ordinal: 1, Version: "initial", ChangeAt: time.Unix(1, 0)}}
	got := renderGraphVersions("beads/a", graphstore.KindMemory, rows, true, false, false)
	for _, text := range []string{"SEQ", "(deleted; not a live snapshot)", "  3", "  2", "  1"} {
		if !strings.Contains(got, text) {
			t.Fatalf("missing %q: %s", text, got)
		}
	}
}

func TestGraphHumanRevisionLookupIsOptional(t *testing.T) {
	// A canceled display lookup must fall back without trying to dereference a
	// provider. Successful command content does not depend on acquiring History.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if got := graphHumanRevision(ctx, nil, "https://example.invalid/a", "full-token"); got != "@=full-token" {
		t.Fatal(got)
	}
}

func TestGraphHumanPreservesDomainProperties(t *testing.T) {
	const scope = "https://example.invalid/scope/"
	r := graphstore.LinkRecord{ID: scope + "L", Type: scope + "types/references", Properties: map[string]any{"title": 17, "body": map[string]any{"nested": true}, "description": "domain description", "metadata": map[string]any{"nested": "content"}, "revision": "domain revision", "id": "domain id", "created_at": "domain date", "empty": ""}}
	for _, long := range []bool{false, true} {
		got := graphHumanDocument(scope, r, "@1", false, long)
		for _, text := range []string{`"title": 17`, `"body": {`, `"description": "domain description"`, `"metadata": {`, `"revision": "domain revision"`, `"id": "domain id"`, `"created_at": "domain date"`, `"empty": ""`} {
			if !strings.Contains(got, text) {
				t.Fatalf("long=%t missing %s: %s", long, text, got)
			}
		}
	}
	var props graphstore.Properties
	if err := json.Unmarshal([]byte(`{"title":"M","body":"B","description":"extra","revision":"user rev","metadata":{"content":true}}`), &props); err != nil {
		t.Fatal(err)
	}
	got := graphHumanDocument(scope, graphstore.Record{ID: scope + "M", Type: scope + "types/memory", Properties: props}, "@1", false, false)
	for _, text := range []string{`"description": "extra"`, `"revision": "user rev"`, `"metadata": {`} {
		if !strings.Contains(got, text) {
			t.Fatal(got)
		}
	}
}

func TestGraphHumanMetadataPrecisionAndLongTimeline(t *testing.T) {
	raw := json.RawMessage(`{"integer":9007199254740993,"decimal":0.12345678901234567890123456789}`)
	got := graphHumanDocument("https://example.invalid/scope/", graphstore.Record{ID: "https://example.invalid/scope/M", Metadata: raw}, "@1", false, false)
	for _, text := range []string{"9007199254740993", "0.12345678901234567890123456789"} {
		if !strings.Contains(got, text) {
			t.Fatal(got)
		}
	}
	rows := []graphstore.VersionRow{{Ordinal: 2, Version: "exact-deletion-token", Removed: true}, {Ordinal: 1, Version: "exact-live-token"}}
	short := renderGraphVersions("M", graphstore.KindMemory, rows, true, true, false)
	detail := renderGraphVersions("M", graphstore.KindMemory, rows, true, true, true)
	if strings.Contains(short, "exact-deletion-token") || strings.Contains(short, "exact-live-token") || !strings.Contains(detail, "exact-deletion-token") || !strings.Contains(detail, "exact-live-token") {
		t.Fatalf("short=%s detail=%s", short, detail)
	}
}

func TestGraphHumanCurrentAdjunctsDiscloseGatesAndCommentBounds(t *testing.T) {
	at := time.Date(2026, 10, 23, 12, 30, 0, 0, time.UTC)
	details := &types.IssueDetails{GatedBy: []types.GateRef{{ID: "https://example.invalid/scope/g", Type: "human", Reason: "approval required"}}, Comments: []*types.Comment{{Author: "A", Text: "first", CreatedAt: at}, {Author: "B", Text: "second", CreatedAt: at}, {Author: "C", Text: "third", CreatedAt: at}}}
	got := graphHumanCurrentIssueDetails("https://example.invalid/scope/", details, graphIssueDetailOptions{CommentsTail: 1, Long: true, LocalTime: true})
	for _, text := range []string{"Current gates", "g [human]: approval required", "2 older comments hidden", at.Local().Format("2006-01-02 15:04:05 -07:00"), "C: third"} {
		if !strings.Contains(got, text) {
			t.Fatal(got)
		}
	}
	if strings.Contains(got, "A: first") || strings.Contains(got, "B: second") {
		t.Fatal(got)
	}
}
