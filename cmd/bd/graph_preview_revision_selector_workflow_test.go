//go:build cgo

package main

import (
	"testing"

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
				graphPolicyCLI(t, bd, work, home, nil, code, append(args, "--json")...)
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
			if oldBySuffix != oldByFlag {
				t.Fatal("ID@REV did not select the same retained Bead revision as --version")
			}
			current := call("show", "beads/plan")
			if oldBySuffix == current || call("show", "beads/plan@current") != current || call("show", "beads/plan@unpinned") != current {
				t.Fatal("exact and floating/current Bead selectors were conflated")
			}
			graphPolicyCLI(t, bd, work, home, nil, "invalid_selector", "show", "beads/plan@"+created.Revision, "--version", created.Revision, "--json")
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
