package main

import (
	"strings"
	"testing"
	"unicode/utf8"

	graph "github.com/steveyegge/beads/graphops"
)

func TestGraphPreviewCreationDefaults(t *testing.T) {
	seen := map[string]bool{}
	for range 32 {
		path, err := graphPreviewCreateBeadPath(selectedRememberCommand(t, nil))
		if err != nil || graph.ValidateBeadPath(path) != nil || seen[path] {
			t.Fatalf("invalid or reused generated path %q: %v", path, err)
		}
		seen[path] = true
	}
	for _, id := range []string{"beads/explicit", "", "links/wrong", "../wrong"} {
		path, err := graphPreviewCreateBeadPath(selectedRememberCommand(t, []string{"--id=" + id}))
		if id == "beads/explicit" {
			if err != nil || path != id {
				t.Fatalf("explicit ID changed: %q %v", path, err)
			}
		} else if err == nil {
			t.Fatalf("invalid explicit ID %q was replaced with %q", id, path)
		}
	}
	for _, tc := range []struct{ body, title string }{
		{"  Code  flow\tpolicy\r\nKeep the whole body.", "Code flow policy"},
		{"\n\t", ""},
		{strings.Repeat("雪", 80), strings.Repeat("雪", 80)},
		{strings.Repeat("雪", 81), strings.Repeat("雪", 79) + "…"},
	} {
		if got := graphPreviewMemoryTitle(tc.body); got != tc.title || !utf8.ValidString(got) {
			t.Fatalf("title=%q want=%q", got, tc.title)
		}
	}
}
