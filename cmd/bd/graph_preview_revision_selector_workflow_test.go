//go:build cgo

package main

import (
	"strings"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// These process-level assertions keep the staged CLI honest: parsing a pin
// cannot turn a floating Link into a purported pinned Link while the writer's
// stored Reference and Type revisions are still missing.
func TestGraphPreviewLinkEndpointRevisionAdmission(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			const scope = "https://example.invalid/reference-cli/"
			work, home := graphListWorkspace(t, bd, engine, scope)
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			refuse := func(code string, args ...string) {
				t.Helper()
				graphPatchProcess(t, bd, work, home, nil, code, 90*time.Second, append(args, "--json")...)
			}
			call("create", "Source", "--id", "beads/source")
			call("create", "Target", "--id", "beads/target")
			sourceBefore := call("show", "beads/source")
			linksBefore := call("link", "list", "--source", "beads/source")
			for _, tc := range []struct {
				name, source, target string
				flags                []string
			}{
				{"source exact suffix", "source@observed", "target", nil},
				{"source current suffix", "source@current", "target", nil},
				{"target exact suffix", "source", "target@observed", nil},
				{"target current suffix", "source", "target@current", nil},
				{"source exact flag", "source", "target", []string{"--source-revision", "observed"}},
				{"target current flag", "source", "target", []string{"--target-revision", "current"}},
			} {
				t.Run(tc.name, func(t *testing.T) {
					args := []string{"link", "create", tc.source, tc.target, "--type", "types/preview-related"}
					args = append(args, tc.flags...)
					refuse("capability_unavailable", args...)
					if got := call("show", "beads/source"); got != sourceBefore {
						t.Fatal("refused Link creation changed the owning source")
					}
					if got := call("link", "list", "--source", "beads/source"); got != linksBefore {
						t.Fatal("refused Link creation inserted an endpoint or Link")
					}
				})
			}
			for _, args := range [][]string{
				{"link", "create", "source@unpinned", "target", "--type", "types/preview-related", "--source-revision", "unpinned"},
				{"link", "create", "source", "target@observed", "--type", "types/preview-related", "--target-revision", "observed"},
			} {
				refuse("invalid_selector", args...)
			}
			for _, args := range [][]string{
				{"link", "create", "--source", "source", "--target", "target", "--type", "types/preview-related", "--source-revision", "current"},
				{"link", "source", "[https://external.example/item]@r7", "--type", "types/preview-related"},
				{"link", "source@current", "target", "--type", "types/preview-related"},
				{"link", "source", "target@observed", "--type", "types/preview-related"},
				{"link", "source", "target", "--type", "types/preview-related", "--source-revision", "current"},
				{"link", "source", "target", "--type", "blocks", "--target-revision", "observed"},
			} {
				refuse("capability_unavailable", args...)
				if got := call("show", "beads/source"); got != sourceBefore {
					t.Fatal("refused shorthand Link creation changed the source")
				}
				if got := call("link", "list", "--source", "beads/source"); got != linksBefore {
					t.Fatal("refused shorthand Link creation changed Links")
				}
			}
			refuse("invalid_selector", "link", "source", "https://external.example/item", "--type", "types/preview-related")
			for _, tc := range []struct {
				name, source, target, id string
				flags                    []string
			}{
				{"bare", "source", "target", "links/bare", nil},
				{"explicit suffix floating", "source@unpinned", "target@unpinned", "links/suffix-floating", nil},
				{"explicit flag floating", "source", "target", "links/flag-floating", []string{"--source-revision", "unpinned", "--target-revision", "unpinned"}},
			} {
				t.Run(tc.name, func(t *testing.T) {
					args := []string{"link", "create", tc.source, tc.target, "--type", "types/preview-related", "--id", tc.id}
					args = append(args, tc.flags...)
					created := graphMixedResult[graphstore.LinkMutationResult](t, call(args...))
					if created.Link.ID != scope+tc.id || created.Link.Source != scope+"beads/source" || created.Link.Target != scope+"beads/target" {
						t.Fatalf("floating Link was not stored as the same endpoints: %+v", created.Link)
					}
				})
			}
			shorthand := graphMixedResult[graphstore.LinkMutationResult](t, call("link", "source@unpinned", "target", "--type", "types/preview-related", "--id", "links/shorthand-floating", "--target-revision", "unpinned"))
			if shorthand.Link.ID != scope+"links/shorthand-floating" || shorthand.Link.Source != scope+"beads/source" || shorthand.Link.Target != scope+"beads/target" {
				t.Fatalf("shorthand floating Link changed endpoint identity: %+v", shorthand.Link)
			}
			call("link", "source@unpinned", "target", "--type", "blocks", "--target-revision", "unpinned")
			bare := call("link", "show", "links/bare")
			refuse("invalid_selector", "link", "update", "links/bare@stale", "--properties", `{"note":"wrong target"}`)
			refuse("invalid_selector", "link", "delete", "links/bare@stale", "--force")
			if got := call("link", "show", "links/bare"); got != bare {
				t.Fatal("historical write-target syntax changed the bare Link")
			}
		})
	}
}

func TestGraphPreviewExactReadIDRevisionRoutesToRetainedState(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := graphListWorkspace(t, bd, engine, "https://example.invalid/reference-read/")
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			created := graphMixedResult[graphstore.Record](t, call("remember", "First body", "--id", "beads/plan"))
			call("update", "beads/plan", "--properties", `{"body":"Second body"}`)
			oldByFlag := call("show", "beads/plan", "--version", created.Revision)
			oldBySuffix := call("show", "beads/plan@"+created.Revision)
			if got := call("show", "[https://example.invalid/reference-read/beads/plan]@"+created.Revision); got != oldByFlag {
				t.Fatal("URI@revision did not select exact retained state")
			}
			if oldBySuffix != oldByFlag {
				t.Fatal("ID@REV did not select the same retained Bead revision as --version")
			}
			current := call("show", "beads/plan")
			if oldBySuffix == current || call("show", "beads/plan@current") != current || call("show", "beads/plan@unpinned") != current {
				t.Fatal("exact and floating/current Bead selectors were conflated")
			}
			graphPatchProcess(t, bd, work, home, nil, "invalid_selector", 90*time.Second, "show", "beads/plan@"+created.Revision, "--version", created.Revision, "--json")
			// Unknown suffixes must not fall back to the current record. The
			// existing raw-read helper verifies numeric exit3 and no stdout.
			for _, selector := range []string{"plan@no-such-token", "[https://example.invalid/reference-read/beads/plan]@no-such-token"} {
				graphMemoryReadRaw(t, bd, work, home, "", "revision_unknown", "show", selector)
			}
			graphPatchProcess(t, bd, work, home, nil, "capability_unavailable", 90*time.Second, "show", "plan@"+created.Revision, "--include-dependents", "--json")
			for _, word := range []string{"current", "unpinned"} {
				// The flag is an exact-token escape, even for reserved suffix
				// words. These words aren't retained tokens in this fixture,
				// so exact lookup must refuse instead of reading current.
				graphPolicyCLI(t, bd, work, home, nil, "revision_unknown", "show", "plan", "--version", word, "--json")
			}
			for _, suffix := range []string{"@" + created.Revision, "@current", "@unpinned"} {
				for _, command := range []string{"update", "delete", "forget"} {
					args := []string{command, "plan" + suffix}
					if command == "update" {
						args = append(args, "--properties", `{"body":"wrong"}`)
					}
					graphPatchProcess(t, bd, work, home, nil, "invalid_selector", 90*time.Second, append(args, "--json")...)
				}
			}
			if got := call("show", "plan"); got != current {
				t.Fatal("refused historical write changed current Memory")
			}
			call("create", "Target", "--id", "beads/target")
			linked := graphMixedResult[graphstore.LinkMutationResult](t, call("link", "create", "plan", "target", "--type", "types/preview-related", "--id", "links/context"))
			call("link", "update", "links/context", "--properties", `{"note":"changed"}`)
			graphMemoryReadRaw(t, bd, work, home, "", "revision_unknown", "link", "show", "links/context@no-such-token")
			graphMemoryReadRaw(t, bd, work, home, "", "revision_unknown", "link", "show", "[https://example.invalid/reference-read/links/context]@no-such-token")
			graphPatchProcess(t, bd, work, home, nil, "invalid_selector", 90*time.Second, "link", "show", "links/context@"+linked.Link.Revision, "--version", linked.Link.Revision, "--json")
			linkByFlag := call("link", "show", "links/context", "--version", linked.Link.Revision)
			linkBySuffix := call("link", "show", "links/context@"+linked.Link.Revision)
			if linkBySuffix != linkByFlag || linkBySuffix == call("link", "show", "links/context") {
				t.Fatal("Link ID@REV did not route to exact retained Link state")
			}
		})
	}
}

// Raw @ remains valid in stored IDs. Full canonical URLs escape the new local
// shorthand through every affected read/write adapter, including Link aliases.
func TestGraphPreviewCanonicalURIAtIdentityWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			const scope = "https://example.invalid/raw-at/"
			work, home := graphListWorkspace(t, bd, engine, scope)
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			const memoryURI = scope + "beads/plan@literal"
			const linkURI = scope + "links/context@literal"
			created := graphMixedResult[graphstore.Record](t, call("remember", "Raw at", "--id", "beads/plan@literal"))
			if got := graphMixedResult[graphstore.Record](t, call("show", memoryURI)); got.ID != memoryURI {
				t.Fatalf("wrong identity: %+v", got)
			}
			call("update", memoryURI, "--properties", `{ "body":"Updated raw at" }`)
			if got := call("show", "["+memoryURI+"]@"+created.Revision); got != call("show", memoryURI, "--version", created.Revision) {
				t.Fatal("canonical raw@ exact read mismatch")
			}
			call("create", "Target", "--id", "beads/target")
			call("link", "create", memoryURI, "target", "--type", "types/preview-related", "--id", "links/context@literal")
			shown := graphMixedResult[graphstore.LinkRecord](t, call("link", "show", linkURI))
			if shown.ID != linkURI || shown.Source != memoryURI {
				t.Fatalf("canonical raw@ Link identity lost: %+v", shown)
			}
			call("link", "update", linkURI, "--properties", `{ "note":"Changed" }`)
			if got := call("link", "show", "["+linkURI+"]@"+shown.Revision); got != call("link", "show", linkURI, "--version", shown.Revision) {
				t.Fatal("canonical raw@ Link exact read mismatch")
			}
			call("link", "delete", linkURI, "--force")
			call("link", memoryURI, "target", "--type", "types/preview-related", "--id", "links/shorthand@literal")
			call("link", "delete", scope+"links/shorthand@literal", "--force")
			call("delete", memoryURI, "--force")
			call("remember", "Forget raw at", "--id", "beads/forget@literal")
			call("forget", scope+"beads/forget@literal")
			// Plain mutation output discloses the writer's actual current
			// revision, not the replaced predecessor or input guard.
			call("remember", "Before disclosure", "--id", "beads/disclosure")
			before := graphMixedResult[graphstore.Record](t, call("show", "disclosure"))
			human := graphPolicyCLI(t, bd, work, home, nil, "", "update", "disclosure", "--properties", `{ "body":"After disclosure" }`)
			after := graphMixedResult[graphstore.Record](t, call("show", "disclosure"))
			if !strings.Contains(human, "Current Memory revision: ["+after.ID+"]@"+after.Revision) || !strings.Contains(human, "Replaced Memory "+before.ID+" version "+before.Revision) {
				t.Fatalf("ambiguous Memory update output: %q", human)
			}
			human = graphPolicyCLI(t, bd, work, home, nil, "", "link", "disclosure", "target", "--type", "types/preview-related", "--id", "links/disclosed")
			owner := graphMixedResult[graphstore.Record](t, call("show", "disclosure"))
			linked := graphMixedResult[graphstore.LinkRecord](t, call("show", "links/disclosed"))
			if !strings.Contains(human, "Current Link revision: ["+linked.ID+"]@"+linked.Revision) || !strings.Contains(human, "Current source revision: ["+owner.ID+"]@"+owner.Revision) {
				t.Fatalf("Link write omitted current source revision: %q", human)
			}
			call("link", "create", "disclosure", "target", "--type", "types/preview-related", "--id", "links/guarded", "--if-source-revision", owner.Revision)
			graphPatchProcess(t, bd, work, home, nil, "revision_conflict", 90*time.Second, "link", "create", "disclosure", "target", "--type", "types/preview-related", "--id", "links/stale", "--if-source-revision", owner.Revision, "--json")
		})
	}
}

// Colon-only spellings are local selectors even when net/url would call them
// opaque URIs. Literal-@ siblings are accessible only by canonical Scope URL.
func TestGraphPreviewColonReferencePrecedenceWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			const scope = "https://example.invalid/colon-reference/"
			work, home := graphListWorkspace(t, bd, engine, scope)
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			refuse := func(code string, args ...string) {
				t.Helper()
				graphPatchProcess(t, bd, work, home, nil, code, 90*time.Second, append(args, "--json")...)
			}
			call("create", "Target", "--id", "beads/target")
			for _, tc := range []struct{ name, id, unknown string }{
				{"colon local", "plan:v1", "unknown-token"},
				{"opaque looking local", "urn:x", "y"},
			} {
				t.Run(tc.name, func(t *testing.T) {
					created := graphMixedResult[graphstore.Record](t, call("remember", "Subject", "--id", "beads/"+tc.id))
					subject := scope + "beads/" + tc.id
					sibling := subject + "@" + tc.unknown
					literal := graphMixedResult[graphstore.Record](t, call("remember", "Literal sibling", "--id", "beads/"+tc.id+"@"+tc.unknown))
					call("update", subject, "--properties", `{"body":"Later subject"}`)
					beforeSubject, beforeSibling := call("show", subject), call("show", sibling)
					beforeLinks := call("link", "list", "--source", subject)
					for _, prefix := range []string{"", "beads/"} {
						if got := call("show", prefix+tc.id+"@"+created.Revision); got != call("show", subject, "--version", created.Revision) {
							t.Fatal("colon selector did not select the exact retained subject")
						}
						selector := prefix + tc.id + "@" + tc.unknown
						graphMemoryReadRaw(t, bd, work, home, "", "revision_unknown", "show", selector)
						for _, command := range []string{"update", "delete", "forget"} {
							args := []string{command, selector}
							if command == "update" {
								args = append(args, "--properties", `{"body":"Must not land"}`)
							}
							if command == "delete" {
								args = append(args, "--force")
							}
							refuse("invalid_selector", args...)
						}
						refuse("capability_unavailable", "link", "create", selector, "target", "--type", "types/preview-related")
						refuse("capability_unavailable", "link", "create", "target", selector, "--type", "types/preview-related")
					}
					if call("show", subject) != beforeSubject || call("show", sibling) != beforeSibling || call("link", "list", "--source", subject) != beforeLinks {
						t.Fatal("a refused colon selector changed subject, literal sibling or Links")
					}
					// Canonical URL fallback addresses the literal colon+@ identity
					// for both writes and exact reads, without touching the subject.
					call("update", sibling, "--properties", `{"body":"Literal sibling updated"}`)
					if call("show", sibling) == beforeSibling || call("show", subject) != beforeSubject {
						t.Fatal("canonical colon+@ mutation selected the wrong identity")
					}
					if got := call("show", "["+sibling+"]@"+literal.Revision); got != call("show", sibling, "--version", literal.Revision) || got == call("show", sibling) {
						t.Fatal("canonical colon+@ exact read did not retain the literal sibling")
					}
				})
			}
			t.Run("colon floating endpoint", func(t *testing.T) {
				const sibling = scope + "beads/plan:v1@unpinned"
				call("remember", "Literal unpinned sibling", "--id", "beads/plan:v1@unpinned")
				beforeSibling := call("show", sibling)
				beforeUnknownSibling := call("show", scope+"beads/plan:v1@unknown-token")
				created := graphMixedResult[graphstore.LinkMutationResult](t, call("link", "create", "plan:v1@unpinned", "target", "--type", "types/preview-related", "--id", "links/colon-floating"))
				if created.Link.Source != scope+"beads/plan:v1" || created.Link.Target != scope+"beads/target" || created.Link.ID != scope+"links/colon-floating" {
					t.Fatalf("floating colon endpoint selected the wrong identity: %+v", created.Link)
				}
				if call("show", sibling) != beforeSibling || call("show", scope+"beads/plan:v1@unknown-token") != beforeUnknownSibling {
					t.Fatal("floating colon endpoint changed a literal-@ sibling")
				}
			})
			linked := graphMixedResult[graphstore.LinkMutationResult](t, call("link", "create", "plan:v1", "target", "--type", "types/preview-related", "--id", "links/context:v1"))
			call("link", "create", "plan:v1", "target", "--type", "types/preview-related", "--id", "links/context:v1@unknown-token")
			const link = scope + "links/context:v1"
			const literalLink = link + "@unknown-token"
			beforeLink, beforeLiteralLink := call("link", "show", link), call("link", "show", literalLink)
			for _, prefix := range []string{"", "links/"} {
				if got := call("link", "show", prefix+"context:v1@"+linked.Link.Revision); got != call("link", "show", link, "--version", linked.Link.Revision) {
					t.Fatal("colon Link selector did not select exact retained Link")
				}
				selector := prefix + "context:v1@unknown-token"
				graphMemoryReadRaw(t, bd, work, home, "", "revision_unknown", "link", "show", selector)
				refuse("invalid_selector", "link", "update", selector, "--properties", `{"note":"Must not land"}`)
				refuse("invalid_selector", "link", "delete", selector, "--force")
			}
			if call("link", "show", link) != beforeLink || call("link", "show", literalLink) != beforeLiteralLink {
				t.Fatal("refused colon Link mutation selected a literal-@ sibling")
			}
			call("link", "update", literalLink, "--properties", `{"note":"Literal Link updated"}`)
			if call("link", "show", literalLink) == beforeLiteralLink || call("link", "show", link) != beforeLink {
				t.Fatal("canonical colon+@ Link mutation selected the wrong identity")
			}
			beforeSubject := call("show", "plan:v1")
			beforeLinks := call("link", "list", "--source", "plan:v1")
			for _, tc := range []struct{ name, selector string }{
				{"opaque bracket", "[urn:x@y]@unpinned"},
				{"no authority bracket", "[https:///beads/plan:v1]@unpinned"},
				{"malformed bracket", "[https://[broken/beads/plan:v1]@unpinned"},
				{"foreign bracket", "[https://foreign.example/beads/plan:v1]@unpinned"},
			} {
				t.Run(tc.name, func(t *testing.T) {
					refuse("invalid_selector", "show", tc.selector)
					refuse("invalid_selector", "link", "show", tc.selector)
					refuse("invalid_selector", "update", tc.selector, "--properties", `{"body":"Must not land"}`)
					refuse("invalid_selector", "link", "create", "plan:v1", tc.selector, "--type", "types/preview-related")
				})
			}
			if call("show", "plan:v1") != beforeSubject || call("link", "list", "--source", "plan:v1") != beforeLinks {
				t.Fatal("refused bracketed locator changed current state")
			}
		})
	}
}
