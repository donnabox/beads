package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func TestGraphPreviewCurrentRevisionDisclosure(t *testing.T) {
	const id = "https://example.invalid/scope/beads/a@b"
	const link = "https://example.invalid/scope/links/context"
	for _, tc := range []struct {
		name   string
		result any
		want   []string
		absent string
	}{
		{"Memory update", graphstore.MemoryMutationResult{Memory: graphstore.Record{ID: id, Revision: "new-memory"}}, []string{"Current Memory revision: [" + id + "]@=new-memory"}, ""},
		{"Issue update", graphstore.IssueMutationResult{Issue: graphstore.IssueRecord{ID: id, Revision: "new-issue"}}, []string{"Current Issue revision: [" + id + "]@=new-issue"}, ""},
		{"owned informational Link", graphstore.LinkMutationResult{Link: graphstore.LinkRecord{ID: link, Revision: "new-link"}, Source: graphstore.Record{ID: id, Revision: "new-owner"}}, []string{"Current Link revision: [" + link + "]@=new-link", "Current source revision: [" + id + "]@=new-owner"}, ""},
		{"native Dependency", graphstore.DependencyResult{Link: graphstore.LinkRecord{ID: link, Revision: "new-dependency"}, Source: graphstore.IssueRecord{ID: id, Revision: "new-owner"}}, []string{"Current Link revision: [" + link + "]@=new-dependency", "Current source revision: [" + id + "]@=new-owner"}, ""},
		{"removed Link", graphstore.LinkDeleteResult{Link: graphstore.LinkTombstone{ID: link, Revision: "old-link"}, Source: graphstore.Record{ID: id, Revision: "new-owner"}}, []string{"Current source revision: [" + id + "]@=new-owner"}, "Current Link revision"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, summary, err := graphPreviewReplacementResult(tc.result, "Successful write", &graphstore.ReplacedMemory{ID: id, Version: "old-memory"}, nil)
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range append(tc.want, "Replaced Memory "+id+" version old-memory") {
				if !strings.Contains(summary, want) {
					t.Fatalf("summary %q lacks %q", summary, want)
				}
			}
			if tc.absent != "" && strings.Contains(summary, tc.absent) {
				t.Fatalf("misleading current state: %q", summary)
			}
			if strings.Contains(summary, "]@old-memory") || strings.Contains(summary, "]@=old-memory") {
				t.Fatalf("predecessor advertised as current: %q", summary)
			}
		})
	}
	failure := errors.New("refused")
	if _, summary, err := graphPreviewReplacementResult(nil, "not committed", nil, failure); !errors.Is(err, failure) || summary != "" {
		t.Fatalf("failed write produced a success summary: %q, %v", summary, err)
	}
}
