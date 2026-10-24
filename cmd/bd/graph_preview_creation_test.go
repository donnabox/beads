package main

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestGraphPreviewCreationDefaults(t *testing.T) {
	for _, tc := range []struct{ body, title string }{
		{"  Code  flow\tpolicy\r\nKeep the whole body.", "Code flow policy"},
		{"\n\t", "Untitled memory"},
		{strings.Repeat("雪", 80), strings.Repeat("雪", 80)},
		{strings.Repeat("雪", 81), strings.Repeat("雪", 79) + "…"},
	} {
		if got := graphPreviewMemoryTitle(tc.body); got != tc.title || !utf8.ValidString(got) {
			t.Fatalf("title=%q want=%q", got, tc.title)
		}
	}
}

// A URL is never a bare ID. The refusal names the forms that are accepted, and
// says --id takes no URL at all, instead of implying that a Scope URL works
// wherever an ID does.
func TestGraphPreviewURLSelectorRefusalNamesAcceptedForms(t *testing.T) {
	const url = "https://example.invalid/other/beads/policy"
	want := []string{
		`invalid local selector "` + url + `"`,
		"this workspace's Scope URL followed by beads/PATH or links/PATH",
		"--id accepts only a bare ID or beads/PATH",
	}
	if _, err := graphPreviewBareBeadPath(url); err == nil {
		t.Fatal("URL accepted as a bare Bead ID")
	} else {
		for _, fragment := range want {
			if !strings.Contains(err.Error(), fragment) {
				t.Errorf("selector refusal omits %q: %v", fragment, err)
			}
		}
	}

}
