package graphops

import "testing"

func TestCanonicalResourceIdentityWithoutKindPrefixes(t *testing.T) {
	const scope = "https://example.invalid/work/"
	for _, id := range []string{"x", "beads/x", "links/x", "types/x", "foo/bar", "snow/%E9%9B%AA", "a%252Fb", "a@b", "Case"} {
		uri, err := CanonicalResourceURL(scope, id)
		if err != nil || uri != scope+id {
			t.Fatalf("canonical composition changed %q: %q %v", id, uri, err)
		}
		got, err := ScopeLocalResourceID(scope, uri)
		if err != nil || got != id {
			t.Fatalf("inverse changed %q: %q %v", id, got, err)
		}
		mounted, err := MountCanonicalResourceID("foo/bar", id)
		if err != nil || mounted != "foo/bar/"+id {
			t.Fatalf("explicit root changed source spelling %q: %q %v", id, mounted, err)
		}
		unmounted, err := MountCanonicalResourceID("", id)
		if err != nil || unmounted != id {
			t.Fatalf("omitted root changed %q: %q %v", id, unmounted, err)
		}
	}
	if _, err := ScopeLocalResourceID(scope, "https://example.invalid/work-other/x"); err == nil {
		t.Fatal("accepted a different Scope")
	}
}

func TestCanonicalResourceIdentityRetainsSegmentGrammar(t *testing.T) {
	for _, id := range []string{"", ".", "..", "a//b", "\x00", "%00", "a%2Fb", "a%5Cb", "%2E", "%2e", "a?b", "a#b", "/x", "//host/x", "x/", "a\\b", "雪", "%61", "a%", "%FF"} {
		if err := ValidateCanonicalResourceID(id); err == nil {
			t.Errorf("accepted noncanonical ID %q", id)
		}
		if _, err := CanonicalResourceURL("https://example.invalid/", id); err == nil {
			t.Errorf("composed noncanonical ID %q", id)
		}
		if _, err := MountCanonicalResourceID(id, "valid"); id != "" && err == nil {
			t.Errorf("accepted invalid explicit root %q", id)
		}
	}
}
