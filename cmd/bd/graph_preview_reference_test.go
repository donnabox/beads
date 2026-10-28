package main

import (
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"

	"github.com/spf13/cobra"
	graph "github.com/steveyegge/beads/graphops"
)

func TestGraphPreviewParseReference(t *testing.T) {
	for _, tc := range []struct {
		name, input, identity, revision string
		explicit, wrapped               bool
	}{
		{"bare local", "beads/plan", "beads/plan", "", false, false},
		{"rooted colon local pin", "beads/plan:v1@=r7", "beads/plan:v1", "r7", true, false},
		{"canonical colon literal at", "https://example.test/beads/plan:v1@=r7", "https://example.test/beads/plan:v1@=r7", "", false, false},
		{"local pin", "plan@=r7", "plan", "r7", true, false},
		{"local current", "plan@current", "plan", "current", true, false},
		{"raw uri userinfo", "https://user@example.test/beads/x", "https://user@example.test/beads/x", "", false, false},
		{"raw uri path", "https://example.test/a@b", "https://example.test/a@b", "", false, false},
		{"wrapped uri", "[https://example.test/beads/x]@=r7", "https://example.test/beads/x", "r7", true, true},
		{"literal local canonical URI", "https://example.test/beads/a@b", "https://example.test/beads/a@b", "", false, false},
		{"opaque suffix containing at", "plan@=r@7", "plan", "r@7", true, false},
		{"wrapped uppercase scheme ipv6", "[HTTPS://[2001:db8::1]/beads/x]@=r7", "HTTPS://[2001:db8::1]/beads/x", "r7", true, true},
		{"wrapped mixed scheme ipv6", "[HtTpS://[2001:db8::1]/beads/x]@=r7", "HtTpS://[2001:db8::1]/beads/x", "r7", true, true},
		{"wrapped ipv6", "[https://[2001:db8::1]/beads/x]@=r7", "https://[2001:db8::1]/beads/x", "r7", true, true},
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
		"", "plan@unpinned", "plan@r7", "@=r7", "plan@", "[https://example.test/x]", "[https://[2001:db8::1]/x@=r7",
		"[HTTPS://[2001:db8::1]/x[y]]@=r7", "[HtTpS://[2001:db8::1]/x?y=[z]]@=r7",
		"[https://example.test/x]oops@=r7", "[https://example.test/x[y]]@=r7", "[https://example.test/x?y=[z]]@=r7", "[not-a-uri]@=r7", "https://[broken/x@=r7", "beads/a://b",
	} {
		if _, err := graphPreviewParseReference(input); err == nil {
			t.Errorf("accepted %q", input)
		}
	}
	for _, input := range []string{"plan@current", "plan@4", "plan@=r7"} {
		if _, err := graphPreviewWritableIdentity(input); err == nil {
			t.Errorf("accepted suffixed mutation target %q", input)
		}
	}
	for _, input := range []string{"@donna", "draft@"} {
		if _, err := graphPreviewParseReference(input); err == nil || !strings.Contains(err.Error(), "full canonical Scope URL") {
			t.Errorf("missing literal-@ escape guidance for %q: %v", input, err)
		}
	}
	if _, err := graphPreviewWritableIdentity("beads/@donna"); err == nil || !strings.Contains(err.Error(), "full canonical Scope URL") {
		t.Fatalf("missing writable literal-@ escape guidance: %v", err)
	}
	for _, revision := range []string{string([]byte{255}), strings.Repeat("x", graphstore.PreviewVersionTokenLimit+1)} {
		if _, err := graphPreviewReferenceWithFlag("plan", revision, true); err == nil {
			t.Error("accepted invalid revision flag")
		}
	}
	if got, err := graphPreviewWritableIdentity("https://example.test/beads/a@b"); err != nil || got != "https://example.test/beads/a@b" {
		t.Fatalf("canonical URL escape = %q, %v", got, err)
	}
	if _, err := graphPreviewReferenceWithFlag("plan@=r7", "r8", true); err == nil {
		t.Fatal("suffix and flag both accepted")
	}
	if _, err := graphPreviewReferenceWithFlag("plan", "", true); err == nil {
		t.Fatal("empty explicit flag accepted")
	}
	if _, err := graphPreviewWritableIdentity("links/context@=r7"); err == nil {
		t.Fatal("historical write target accepted")
	}
	if got, err := graphPreviewWritableIdentity("links/context"); err != nil || got != "links/context" {
		t.Fatalf("bare write target = %q, %v", got, err)
	}
}

func TestGraphPreviewLinkReferenceFlagsAreSeparateFromWriteGuard(t *testing.T) {
	for _, command := range []struct {
		name string
		cmd  *cobra.Command
	}{{"link", linkCmd}, {"link create", linkCreateCmd}, {"update", updateCmd}, {"link update", linkUpdateCmd}} {
		for _, name := range []string{"source-revision", "target-revision", "if-source-revision"} {
			if command.cmd.Flags().Lookup(name) == nil {
				t.Errorf("%s lacks --%s", command.name, name)
			}
		}
	}
}

func TestGraphPreviewCanonicalLocalIDCanContainAt(t *testing.T) {
	// Local CLI shorthand now reserves raw @ as a separator, but the stored
	// identity grammar still permits it. Full canonical URIs escape shorthand;
	// do not tighten identity validation as a side effect of CLI parsing.
	if got := graph.CanonicalSegment("a@b"); got != "a@b" {
		t.Fatalf("CanonicalSegment(a@b) = %q", got)
	}
	if err := graph.ValidateBeadPath("beads/a@b"); err != nil {
		t.Fatalf("current canonical Bead ID refuses @: %v", err)
	}
	if err := graph.ValidateLinkPath("links/a@b"); err != nil {
		t.Fatalf("current canonical Link ID refuses @: %v", err)
	}
}

// URI-looking local segments obey the same first-raw-@ rule as other IDs.
// This decision is lexical; an existing literal-@ sibling cannot select it.
func TestGraphPreviewColonReferencePrecedence(t *testing.T) {
	for _, tc := range []struct{ input, identity, revision string }{
		{"plan:v1@=r7", "plan:v1", "r7"},
		{"urn:x@=y", "urn:x", "y"},
		{"plan:v1@=r@7", "plan:v1", "r@7"},
		{"beads/plan:v1@=r7", "beads/plan:v1", "r7"},
		{"context:v1@=r7", "context:v1", "r7"},
	} {
		t.Run(tc.input, func(t *testing.T) {
			got, err := graphPreviewParseReference(tc.input)
			if err != nil || got.Identity != tc.identity || got.Revision != tc.revision || !got.Explicit || got.Wrapped {
				t.Fatalf("local precedence = %+v, %v", got, err)
			}
			if _, err := graphPreviewWritableIdentity(tc.input); err == nil {
				t.Fatal("colon-bearing revision selector accepted as mutation identity")
			}
		})
	}
	for _, identity := range []string{"plan:v1", "urn:x", "https://example.test/beads/plan:v1@=r7"} {
		if got, err := graphPreviewWritableIdentity(identity); err != nil || got != identity {
			t.Errorf("bare identity %q = %q, %v", identity, got, err)
		}
	}
}

func TestGraphPreviewBracketedLocatorAdmission(t *testing.T) {
	for _, input := range []string{
		"[urn:x@y]@=r7", "[https:plan]@=r7", "[https:/plan]@=r7",
		"[https:///beads/plan]@=r7", "[file:///beads/plan]@=r7",
		"[ftp://example.test/beads/plan]@=r7", "[//example.test/beads/plan]@=r7",
		"[https://[broken/beads/plan]@=r7",
	} {
		t.Run(input, func(t *testing.T) {
			if _, err := graphPreviewParseReference(input); err == nil {
				t.Fatal("noncanonical bracketed locator accepted")
			}
		})
	}
	if _, err := graphPreviewParseReference("[urn:x@y]@=r7"); err == nil || !strings.Contains(err.Error(), "canonical HTTP(S) Scope URL") {
		t.Fatalf("opaque URI refusal lacks canonical Scope guidance: %v", err)
	}
	// Scope membership is checked downstream; parsing must preserve a foreign
	// canonical locator verbatim rather than reinterpret it as a local ID.
	if got, err := graphPreviewParseReference("[https://foreign.example/beads/plan]@=r7"); err != nil || got.Identity != "https://foreign.example/beads/plan" || !got.Wrapped {
		t.Fatalf("foreign locator parsing = %+v, %v", got, err)
	}
}

func TestGraphPreviewEndpointPinIntents(t *testing.T) {
	for _, tc := range []struct {
		name           string
		flags          []string
		source, target *graphstore.EndpointPin
		bad            bool
	}{
		{"omitted", nil, nil, nil, false},
		{"explicit-unpin", []string{"--source-revision=unpinned"}, &graphstore.EndpointPin{}, nil, false},
		{"current-independent-exact", []string{"--source-revision=current", "--target-revision=r@7"}, &graphstore.EndpointPin{Current: true}, &graphstore.EndpointPin{Revision: "r@7"}, false},
		{"empty-refuses", []string{"--target-revision="}, nil, nil, true},
		{"oversized-refuses", []string{"--target-revision=" + strings.Repeat("x", graphstore.PreviewVersionTokenLimit+1)}, nil, nil, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().String("source-revision", "", "")
			cmd.Flags().String("target-revision", "", "")
			if err := cmd.ParseFlags(tc.flags); err != nil {
				t.Fatal(err)
			}
			source, target, err := graphPreviewLinkPinIntents(cmd)
			if (err != nil) != tc.bad {
				t.Fatalf("error: %v", err)
			}
			equal := func(a, b *graphstore.EndpointPin) bool {
				return (a == nil && b == nil) || (a != nil && b != nil && *a == *b)
			}
			if err == nil && (!equal(source, tc.source) || !equal(target, tc.target)) {
				t.Fatalf("lost omission/pin intent: %+v %+v", source, target)
			}
		})
	}
}
