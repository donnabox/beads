package graphops

import "testing"

func TestPublicResourceMountPreservesLiteralSegments(t *testing.T) {
	const scope = "https://example.invalid/work/"
	for _, root := range []string{"beads", "links", "types", "foo/types/bar"} {
		for _, id := range []string{"x", "beads/x", "links/x", "types/x", "a%20b", "a%252Fb"} {
			mounted, err := MountCanonicalResourceID(root, id)
			if err != nil || mounted != root+"/"+id {
				t.Fatalf("mount changed literal segments root=%q id=%q: %q %v", root, id, mounted, err)
			}
			uri, err := CanonicalPublicResourceURL(scope, mounted)
			if err != nil || uri != scope+root+"/"+id {
				t.Fatalf("URL changed mounted spelling: %q %v", uri, err)
			}
			got, err := ScopeLocalPublicResourceID(scope, uri)
			if err != nil || got != mounted {
				t.Fatalf("inverse changed mount spelling: %q %v", got, err)
			}
		}
	}
}

func TestPublicResourceURLDoesNotEncodeLiteralKeyData(t *testing.T) {
	const scope = "https://example.invalid/work/"
	if _, err := CanonicalPublicResourceURL(scope, "a b"); err == nil {
		t.Fatal("helper silently selected an encoding for raw legacy data")
	}
	uri, err := CanonicalPublicResourceURL(scope, "a%20b")
	if err != nil || uri != scope+"a%20b" {
		t.Fatalf("canonical spelling changed: %q %v", uri, err)
	}
	got, err := ScopeLocalPublicResourceID(scope, uri)
	if err != nil || got != "a%20b" {
		t.Fatalf("inverse decoded canonical spelling into a different literal ID: %q %v", got, err)
	}
}

func TestPublicResourceURLValidatesCanonicalScope(t *testing.T) {
	for _, scope := range []string{"https://example.invalid/work", "https://example.invalid/work/../", "https://example.invalid/work/?q=1", "https://user@example.invalid/", "https://example.invalid/%77ork/", "https://example.invalid/work//"} {
		if _, err := CanonicalPublicResourceURL(scope, "x"); err == nil {
			t.Errorf("accepted noncanonical Scope %q", scope)
		}
		if _, err := ScopeLocalPublicResourceID(scope, scope+"x"); err == nil {
			t.Errorf("inverse accepted noncanonical Scope %q", scope)
		}
	}
}
