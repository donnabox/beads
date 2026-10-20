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
			refuse("invalid_selector", "unlink", "links/bare@current", "--unconditional")
			refuse("invalid_selector", "unlink", "source@unpinned", "target", "--type", "types/preview-related", "--unconditional")
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
			call("link", "delete", linkURI, "--force", "--unconditional")
			call("link", memoryURI, "target", "--type", "types/preview-related", "--id", "links/unlink@literal")
			call("unlink", scope+"links/unlink@literal", "--unconditional")
			call("link", memoryURI, "target", "--type", "types/preview-related", "--id", "links/pair@literal")
			call("unlink", memoryURI, scope+"beads/target", "--type", "types/preview-related", "--unconditional")
			call("delete", memoryURI, "--force", "--unconditional")
			call("remember", "Forget raw at", "--id", "beads/forget@literal")
			call("forget", scope+"beads/forget@literal", "--unconditional")
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
