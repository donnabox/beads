package main

import "testing"

func TestGraphPreviewParseReference(t *testing.T) {
	for _, tc := range []struct {
		name, input, identity, revision string
		explicit, wrapped               bool
	}{
		{"bare local", "beads/plan", "beads/plan", "", false, false},
		{"local pin", "plan@r7", "plan", "r7", true, false},
		{"local current", "plan@current", "plan", "current", true, false},
		{"local floating", "plan@unpinned", "plan", "unpinned", true, false},
		{"raw uri userinfo", "https://user@example.test/beads/x", "https://user@example.test/beads/x", "", false, false},
		{"raw uri path", "https://example.test/a@b", "https://example.test/a@b", "", false, false},
		{"wrapped uri", "[https://example.test/beads/x]@r7", "https://example.test/beads/x", "r7", true, true},
		{"wrapped ipv6", "[https://[2001:db8::1]/beads/x]@r7", "https://[2001:db8::1]/beads/x", "r7", true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := graphPreviewParseReference(tc.input)
			if err != nil {
				t.Fatal(err)
			}
			if got.Identity != tc.identity || got.Revision != tc.revision || got.Explicit != tc.explicit || got.Wrapped != tc.wrapped {
				t.Fatalf("parsed %+v", got)
			}
		})
	}
}

func TestGraphPreviewReferenceRefusals(t *testing.T) {
	for _, input := range []string{
		"", "@r7", "plan@", "[https://example.test/x]", "[https://[2001:db8::1]/x@r7",
		"[https://example.test/x]oops@r7", "[not-a-uri]@r7", "https://[broken/x@r7",
	} {
		if _, err := graphPreviewParseReference(input); err == nil {
			t.Errorf("accepted %q", input)
		}
	}
	if _, err := graphPreviewReferenceWithFlag("plan@r7", "r8", true); err == nil {
		t.Fatal("suffix and flag both accepted")
	}
	if _, err := graphPreviewReferenceWithFlag("plan", "", true); err == nil {
		t.Fatal("empty explicit flag accepted")
	}
	if _, err := graphPreviewWritableIdentity("links/context@r7"); err == nil {
		t.Fatal("historical write target accepted")
	}
	if got, err := graphPreviewWritableIdentity("links/context"); err != nil || got != "links/context" {
		t.Fatalf("bare write target = %q, %v", got, err)
	}
}

func TestGraphPreviewLinkReferenceFlagsAreSeparateFromWriteGuard(t *testing.T) {
	for _, name := range []string{"source-revision", "target-revision", "if-source-revision"} {
		if linkCreateCmd.Flags().Lookup(name) == nil {
			t.Errorf("link create lacks --%s", name)
		}
	}
}
