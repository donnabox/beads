//go:build cgo

package main

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/httpapi/bdpwire"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// Process-level admission checks distinguish exact stored endpoint References
// from current write guards, unsupported native pins and malformed selectors.
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
			call("create", "Source", "--id", "source", "--force")
			call("remember", "Target", "--id", "target")
			call("create", "Native target", "--id", "native-target", "--force")
			sourceBefore := call("show", scope+"source")
			linksBefore := call("link", "list", "--source", scope+"source")
			for _, tc := range []struct {
				name, source, target string
				flags                []string
			}{
				{"source exact suffix", "source@observed", "target", nil},
				{"target exact suffix", "source", "target@observed", nil},
				{"source exact flag", "source", "target", []string{"--source-revision", "observed"}},
			} {
				t.Run(tc.name, func(t *testing.T) {
					args := []string{"link", "create", tc.source, tc.target, "--type", "types/references"}
					args = append(args, tc.flags...)
					refuse("revision_unknown", args...)
					if got := call("show", scope+"source"); got != sourceBefore {
						t.Fatal("refused Link creation changed the owning source")
					}
					if got := call("link", "list", "--source", scope+"source"); got != linksBefore {
						t.Fatal("refused Link creation inserted an endpoint or Link")
					}
				})
			}
			for _, args := range [][]string{
				{"link", "create", "source@unpinned", "target", "--type", "types/references", "--source-revision", "unpinned"},
				{"link", "create", "source", "target@observed", "--type", "types/references", "--target-revision", "observed"},
			} {
				refuse("invalid_selector", args...)
			}
			for _, tc := range []struct {
				code string
				args []string
			}{
				{"invalid_selector", []string{"link", "source", "[https://external.example/item]@r7", "--type", "types/references"}},
				{"revision_unknown", []string{"link", "source", "target@observed", "--type", "types/references"}},
				{"capability_unavailable", []string{"link", "source", "native-target@current", "--type", "blocks"}},
				{"capability_unavailable", []string{"link", "source", "native-target", "--type", "blocks", "--target-revision", "observed"}},
			} {
				refuse(tc.code, tc.args...)
				if got := call("show", scope+"source"); got != sourceBefore {
					t.Fatal("refused shorthand Link creation changed the source")
				}
				if got := call("link", "list", "--source", scope+"source"); got != linksBefore {
					t.Fatal("refused shorthand Link creation changed Links")
				}
			}
			refuse("invalid_selector", "link", "source", "https://external.example/item", "--type", "types/references")
			for _, tc := range []struct {
				name, source, target, id string
				flags                    []string
			}{
				{"bare", "source", "target", "links/bare", nil},
				{"explicit suffix floating", "source@unpinned", "target@unpinned", "links/suffix-floating", nil},
				{"explicit flag floating", "source", "target", "links/flag-floating", []string{"--source-revision", "unpinned", "--target-revision", "unpinned"}},
			} {
				t.Run(tc.name, func(t *testing.T) {
					args := []string{"link", "create", tc.source, tc.target, "--type", "types/references", "--id", tc.id}
					args = append(args, tc.flags...)
					created := graphMixedResult[graphstore.LinkMutationResult](t, call(args...))
					if created.Link.ID != scope+tc.id || created.Link.Source != scope+"source" || created.Link.Target != scope+"target" {
						t.Fatalf("floating Link was not stored as the same endpoints: %+v", created.Link)
					}
				})
			}
			for i, args := range [][]string{
				{"link", "create", "source@current", "target@current", "--type", "types/references", "--id", "links/pinned-current"},
				{"link", "create", "--source", "source", "--target", "target", "--source-revision", "current", "--target-revision", "current", "--type", "types/references", "--id", "links/pinned-flags"},
				{"link", "source", "target", "--source-revision", "current", "--type", "types/references", "--id", "links/source-pin"},
			} {
				created := graphMixedResult[graphstore.LinkMutationResult](t, call(args...))
				if created.Link.SourceRevision == "" || (i < 2 && created.Link.TargetRevision == "") || created.Link.Source != scope+"source" || created.Link.Target != scope+"target" {
					t.Fatalf("lost current endpoint pins: %+v", created)
				}
			}
			shorthand := graphMixedResult[graphstore.LinkMutationResult](t, call("link", "source@unpinned", "target", "--type", "types/references", "--id", "links/shorthand-floating", "--target-revision", "unpinned"))
			if shorthand.Link.ID != scope+"links/shorthand-floating" || shorthand.Link.Source != scope+"source" || shorthand.Link.Target != scope+"target" {
				t.Fatalf("shorthand floating Link changed endpoint identity: %+v", shorthand.Link)
			}
			call("link", "["+scope+"source]@unpinned", scope+"native-target", "--type", "blocks", "--target-revision", "unpinned")
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
			const planURL = "https://example.invalid/reference-read/beads/plan"
			oldByFlag := call("show", planURL, "--version", created.Revision)
			oldBySuffix := call("show", "["+planURL+"]@"+created.Revision)
			if got := call("show", "["+planURL+"]@"+created.Revision); got != oldByFlag {
				t.Fatal("URI@revision did not select exact retained state")
			}
			if oldBySuffix != oldByFlag {
				t.Fatal("canonical URL@REV did not select the same retained Bead revision as --version")
			}
			current := call("show", planURL)
			if oldBySuffix == current || call("show", "["+planURL+"]@current") != current || call("show", "["+planURL+"]@unpinned") != current {
				t.Fatal("exact and floating/current Bead selectors were conflated")
			}
			graphPatchProcess(t, bd, work, home, nil, "invalid_selector", 90*time.Second, "show", "["+planURL+"]@"+created.Revision, "--version", created.Revision, "--json")
			// Unknown suffixes must not fall back to the current record. The
			// existing raw-read helper verifies numeric exit3 and no stdout.
			for _, selector := range []string{"[" + planURL + "]@no-such-token"} {
				graphMemoryReadRaw(t, bd, work, home, "", "revision_unknown", "show", selector)
			}
			graphMemoryReadRaw(t, bd, work, home, "", "capability_unavailable", "show", "beads/plan@no-such-token")
			graphPatchProcess(t, bd, work, home, nil, "capability_unavailable", 90*time.Second, "show", "beads/plan@"+created.Revision, "--include-dependents", "--json")
			for _, word := range []string{"current", "unpinned"} {
				// The flag is an exact-token escape, even for reserved suffix
				// words. These words aren't retained tokens in this fixture,
				// so exact lookup must refuse instead of reading current.
				graphPolicyCLI(t, bd, work, home, nil, "revision_unknown", "show", planURL, "--version", word, "--json")
			}
			for _, suffix := range []string{"@" + created.Revision, "@current", "@unpinned"} {
				for _, command := range []string{"update", "delete", "forget"} {
					args := []string{command, "beads/plan" + suffix}
					if command == "update" {
						args = append(args, "--properties", `{"body":"wrong"}`)
					}
					code := "invalid_selector"
					if command == "forget" {
						code = "not_found" // forget selects literal key data, never a revision.
					}
					graphPatchProcess(t, bd, work, home, nil, code, 90*time.Second, append(args, "--json")...)
				}
			}
			if got := call("show", "beads/plan"); got != current {
				t.Fatal("refused historical write changed current Memory")
			}
			call("create", "Target", "--type", "types/constraint", "--id", "beads/target")
			linked := graphMixedResult[graphstore.LinkMutationResult](t, call("link", "create", "beads/plan", "beads/target", "--type", "types/subject-to", "--id", "links/context"))
			call("link", "update", "links/context", "--properties", `{"note":"changed"}`)
			graphMemoryReadRaw(t, bd, work, home, "", "capability_unavailable", "link", "show", "links/context@no-such-token")
			graphMemoryReadRaw(t, bd, work, home, "", "revision_unknown", "link", "show", "[https://example.invalid/reference-read/links/context]@no-such-token")
			graphPatchProcess(t, bd, work, home, nil, "invalid_selector", 90*time.Second, "link", "show", "links/context@"+linked.Link.Revision, "--version", linked.Link.Revision, "--json")
			const linkURL = "https://example.invalid/reference-read/links/context"
			linkByFlag := call("link", "show", linkURL, "--version", linked.Link.Revision)
			linkBySuffix := call("link", "show", "["+linkURL+"]@"+linked.Link.Revision)
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
			call("create", "Target", "--type", "types/constraint", "--id", "beads/target")
			call("link", "create", memoryURI, "beads/target", "--type", "types/subject-to", "--id", "links/context@literal")
			shown := graphMixedResult[graphstore.LinkRecord](t, call("link", "show", linkURI))
			if shown.ID != linkURI || shown.Source != memoryURI {
				t.Fatalf("canonical raw@ Link identity lost: %+v", shown)
			}
			call("link", "update", linkURI, "--properties", `{ "note":"Changed" }`)
			if got := call("link", "show", "["+linkURI+"]@"+shown.Revision); got != call("link", "show", linkURI, "--version", shown.Revision) {
				t.Fatal("canonical raw@ Link exact read mismatch")
			}
			call("link", "delete", linkURI, "--force")
			call("link", memoryURI, "beads/target", "--type", "types/subject-to", "--id", "links/shorthand@literal")
			call("link", "delete", scope+"links/shorthand@literal", "--force")
			call("delete", memoryURI, "--force")
			call("remember", "Forget raw at", "--id", "beads/forget@literal")
			call("forget", scope+"beads/forget@literal")
			// Plain mutation output discloses the writer's actual current
			// revision, not the replaced predecessor or input guard.
			call("remember", "Before disclosure", "--id", "beads/disclosure")
			before := graphMixedResult[graphstore.Record](t, call("show", "beads/disclosure"))
			human := graphPolicyCLI(t, bd, work, home, nil, "", "update", "beads/disclosure", "--properties", `{ "body":"After disclosure" }`)
			after := graphMixedResult[graphstore.Record](t, call("show", "beads/disclosure"))
			if !strings.Contains(human, "Current Memory revision: ["+after.ID+"]@"+after.Revision) || !strings.Contains(human, "Replaced Memory "+before.ID+" version "+before.Revision) {
				t.Fatalf("ambiguous Memory update output: %q", human)
			}
			human = graphPolicyCLI(t, bd, work, home, nil, "", "link", "beads/disclosure", "beads/target", "--type", "types/subject-to", "--id", "links/disclosed")
			owner := graphMixedResult[graphstore.Record](t, call("show", "beads/disclosure"))
			linked := graphMixedResult[graphstore.LinkRecord](t, call("show", scope+"links/disclosed"))
			if !strings.Contains(human, "Current Link revision: ["+linked.ID+"]@"+linked.Revision) || !strings.Contains(human, "Current source revision: ["+owner.ID+"]@"+owner.Revision) {
				t.Fatalf("Link write omitted current source revision: %q", human)
			}
			call("link", "create", "beads/disclosure", "beads/target", "--type", "types/subject-to", "--id", "links/guarded", "--if-source-revision", owner.Revision)
			graphPatchProcess(t, bd, work, home, nil, "revision_conflict", 90*time.Second, "link", "create", "beads/disclosure", "beads/target", "--type", "types/subject-to", "--id", "links/stale", "--if-source-revision", owner.Revision, "--json")
		})
	}
}

// Colon-only spellings are local selectors even when net/url would call them
// opaque URIs. Exact allocated literal-@ IDs win before revision parsing;
// bracketed canonical references remain the exact retained-state escape.
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
			call("create", "Target", "--type", "types/constraint", "--id", "beads/target")
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
					if got := call("show", "["+subject+"]@"+created.Revision); got != call("show", subject, "--version", created.Revision) {
						t.Fatal("canonical colon selector did not select the exact retained subject")
					}
					for _, prefix := range []string{"beads/"} {
						selector := prefix + tc.id + "@no-such-revision"
						graphMemoryReadRaw(t, bd, work, home, "", "capability_unavailable", "show", selector)
						for _, command := range []string{"update", "delete", "forget"} {
							args := []string{command, selector}
							if command == "update" {
								args = append(args, "--properties", `{"body":"Must not land"}`)
							}
							if command == "delete" {
								args = append(args, "--force")
							}
							code := "invalid_selector"
							if command == "forget" {
								code = "not_found" // literal unallocated key, no revision parsing.
							}
							refuse(code, args...)
						}
						refuse("revision_unknown", "link", "create", selector, "beads/target", "--type", "types/subject-to")
						refuse("revision_unknown", "link", "create", "beads/target", selector, "--type", "types/subject-to")
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
				created := graphMixedResult[graphstore.LinkMutationResult](t, call("link", "create", "beads/plan:v1", "beads/target", "--source-revision", "unpinned", "--type", "types/subject-to", "--id", "links/colon-floating"))
				if created.Link.Source != scope+"beads/plan:v1" || created.Link.Target != scope+"beads/target" || created.Link.ID != scope+"links/colon-floating" {
					t.Fatalf("floating colon endpoint selected the wrong identity: %+v", created.Link)
				}
				if call("show", sibling) != beforeSibling || call("show", scope+"beads/plan:v1@unknown-token") != beforeUnknownSibling {
					t.Fatal("floating colon endpoint changed a literal-@ sibling")
				}
			})
			linked := graphMixedResult[graphstore.LinkMutationResult](t, call("link", "create", "beads/plan:v1", "beads/target", "--type", "types/subject-to", "--id", "links/context:v1"))
			call("link", "create", "beads/plan:v1", "beads/target", "--type", "types/subject-to", "--id", "links/context:v1@unknown-token")
			const link = scope + "links/context:v1"
			const literalLink = link + "@unknown-token"
			beforeLink, beforeLiteralLink := call("link", "show", link), call("link", "show", literalLink)
			if got := call("link", "show", "["+link+"]@"+linked.Link.Revision); got != call("link", "show", link, "--version", linked.Link.Revision) {
				t.Fatal("canonical colon Link selector did not select exact retained Link")
			}
			for _, prefix := range []string{"", "links/"} {
				selector := prefix + "context:v1@no-such-revision"
				graphMemoryReadRaw(t, bd, work, home, "", "capability_unavailable", "link", "show", selector)
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
			beforeSubject := call("show", "beads/plan:v1")
			beforeLinks := call("link", "list", "--source", "beads/plan:v1")
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
					refuse("invalid_selector", "link", "create", "beads/plan:v1", tc.selector, "--type", "types/subject-to")
				})
			}
			if call("show", "beads/plan:v1") != beforeSubject || call("link", "list", "--source", "beads/plan:v1") != beforeLinks {
				t.Fatal("refused bracketed locator changed current state")
			}
		})
	}
}

func TestGraphPreviewEndpointRepinWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			const scope = "https://example.invalid/endpoint-repin/"
			work, home := graphListWorkspace(t, bd, engine, scope)
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			refuse := func(code string, args ...string) {
				t.Helper()
				graphPatchProcess(t, bd, work, home, nil, code, 90*time.Second, append(args, "--json")...)
			}
			source := graphMixedResult[graphstore.Record](t, call("remember", "Source", "--id", "source"))
			target := graphMixedResult[graphstore.Record](t, call("create", "Rule one", "--id", "rule", "--type", "types/constraint"))
			created := graphMixedResult[graphstore.LinkMutationResult](t, call("link", "create", "source@"+source.Revision, "rule@"+target.Revision, "--type", "types/references", "--id", "citation", "--properties", `{"note":"keep"}`, "--metadata", `{"origin":"manual"}`))
			if created.Link.SourceRevision != source.Revision || created.Link.TargetRevision != target.Revision {
				t.Fatalf("creation lost exact pins: %+v", created)
			}
			old := call("link", "show", created.Link.ID)
			changedTarget := graphMixedResult[graphstore.MemoryMutationResult](t, call("update", target.ID, "--properties", `{"body":"Rule two"}`, "--if-revision", target.Revision)).Memory
			repinned := graphMixedResult[graphstore.LinkMutationResult](t, call("update", created.Link.ID, "--target-revision", "current", "--if-revision", created.Link.Revision))
			if !repinned.Changed || repinned.Link.ID != created.Link.ID || repinned.Link.Revision == created.Link.Revision || repinned.Link.SourceRevision != source.Revision || repinned.Link.TargetRevision != changedTarget.Revision || !reflect.DeepEqual(repinned.Link.Properties, created.Link.Properties) || string(repinned.Link.Metadata) != string(created.Link.Metadata) {
				t.Fatalf("repin lost identity/omissions: %+v", repinned)
			}
			if got := call("link", "show", created.Link.ID, "--version", created.Link.Revision); got != old {
				t.Fatal("repin changed old retained Link")
			}
			for _, expected := range []graphstore.LinkRecord{created.Link, repinned.Link} {
				raw := graphPolicyCLI(t, bd, work, home, nil, "", "link", "show", expected.ID, "--version", expected.Revision, "--output-format", "json")
				var envelope struct {
					Data bdpwire.LinkRecord `json:"data"`
				}
				if err := json.Unmarshal([]byte(raw), &envelope); err != nil || envelope.Data.ID != expected.ID || envelope.Data.Revision != expected.Revision || envelope.Data.Source.URI != expected.Source || envelope.Data.Source.Revision != expected.SourceRevision || envelope.Data.Target.URI != expected.Target || envelope.Data.Target.Revision != expected.TargetRevision {
					t.Fatalf("canonical retained Link lost exact endpoint References: decode=%v got=%s", err, raw)
				}
			}
			same := graphMixedResult[graphstore.LinkMutationResult](t, call("link", "update", created.Link.ID, "--target-revision", changedTarget.Revision, "--if-revision", repinned.Link.Revision))
			if same.Changed || !reflect.DeepEqual(same.Link, repinned.Link) {
				t.Fatal("same full Reference minted a revision")
			}
			before := call("link", "show", created.Link.ID)
			refuse("revision_conflict", "update", created.Link.ID, "--target-revision", changedTarget.Revision, "--if-revision", created.Link.Revision)
			refuse("revision_unknown", "update", created.Link.ID, "--target-revision", "missing-retained")
			refuse("invalid_selector", "update", created.Link.ID, "--target-revision=")
			if call("link", "show", created.Link.ID) != before {
				t.Fatal("refused re-pin changed current Link")
			}
			merged := graphMixedResult[graphstore.LinkMutationResult](t, call("link", "update", created.Link.ID, "--source-revision", "unpinned", "--set-metadata", "reviewed=true", "--if-revision", repinned.Link.Revision))
			if merged.Link.SourceRevision != "" || merged.Link.TargetRevision != changedTarget.Revision || !reflect.DeepEqual(merged.Link.Properties, repinned.Link.Properties) || !strings.Contains(string(merged.Link.Metadata), `"reviewed":true`) {
				t.Fatalf("pin/metadata merge lost omissions: %+v", merged)
			}
			patched := graphMixedResult[graphstore.LinkMutationResult](t, call("update", created.Link.ID, "--target-revision", "unpinned", "--patch", `[{"op":"replace","path":"/note","value":"second"}]`, "--if-revision", merged.Link.Revision))
			if patched.Link.TargetRevision != "" || patched.Link.SourceRevision != "" || patched.Link.Properties["note"] != "second" || string(patched.Link.Metadata) != string(merged.Link.Metadata) {
				t.Fatalf("combined ordered patch lost re-pin: %+v", patched)
			}
			owned := graphMixedResult[graphstore.LinkMutationResult](t, call("link", "source", "rule@"+target.Revision, "--type", "types/subject-to", "--id", "obligation", "--if-source-revision", source.Revision))
			owner := graphMixedSourceRecord(t, owned.Source)
			oldOwner := call("show", source.ID)
			ownedNext := graphMixedResult[graphstore.LinkMutationResult](t, call("link", "update", owned.Link.ID, "--target-revision", "current", "--if-revision", owned.Link.Revision, "--if-source-revision", owner.Revision))
			ownerNext := graphMixedSourceRecord(t, ownedNext.Source)
			if ownerNext.Revision == owner.Revision || ownedNext.Link.TargetRevision != changedTarget.Revision {
				t.Fatalf("owned target re-pin did not mint source: %+v", ownedNext)
			}
			if call("show", source.ID, "--version", owner.Revision) != oldOwner {
				t.Fatal("owned re-pin rewrote old source")
			}
			for _, expected := range []graphstore.Record{owner, ownerNext} {
				raw := graphPolicyCLI(t, bd, work, home, nil, "", "show", expected.ID, "--version", expected.Revision, "--output-format", "json")
				var envelope struct {
					Data bdpwire.BeadRecord `json:"data"`
				}
				if err := json.Unmarshal([]byte(raw), &envelope); err != nil {
					t.Fatal(err)
				}
				group := envelope.Data.OwnedLinks[scope+"types/subject-to"]
				var wanted graphstore.LinkRecord
				if err := json.Unmarshal(expected.Owned[0], &wanted); err != nil {
					t.Fatal(err)
				}
				if envelope.Data.Revision != expected.Revision || len(group) != 1 || group[0].ID != wanted.ID || group[0].Target.URI != wanted.Target || group[0].Target.Revision != wanted.TargetRevision {
					t.Fatalf("canonical retained selected-Type ownership lost endpoint pin: got=%s", raw)
				}
			}
			before = call("show", source.ID)
			refuse("capability_unavailable", "update", owned.Link.ID, "--source-revision", "current", "--if-revision", ownedNext.Link.Revision, "--if-source-revision", ownerNext.Revision)
			refuse("invalid_selector", "update", source.ID, "--target-revision", "current")
			if call("show", source.ID) != before {
				t.Fatal("refused pin changed owner")
			}
			literal := graphMixedResult[graphstore.Record](t, call("remember", "Literal at", "--id", "plan:v1@draft"))
			sibling := graphMixedResult[graphstore.Record](t, call("remember", "Sibling", "--id", "plan:v1"))
			literalPin := graphMixedResult[graphstore.LinkMutationResult](t, call("link", "create", literal.ID, "rule", "--source-revision", "current", "--type", "types/references", "--id", "literal-pin"))
			if literalPin.Link.Source != literal.ID || literalPin.Link.SourceRevision != literal.Revision {
				t.Fatal("canonical URL literal at resolved wrong source")
			}
			literalBare := graphMixedResult[graphstore.LinkMutationResult](t, call("link", "create", "plan:v1@draft", "rule", "--type", "types/references", "--id", "literal-bare"))
			if literalBare.Link.Source != literal.ID || literalBare.Link.SourceRevision != "" {
				t.Fatal("exact allocated local ID with @ did not remain a floating literal endpoint")
			}
			refuse("revision_unknown", "link", "create", "plan:v1@no-such-revision", "rule", "--type", "types/references")
			if got := graphMixedResult[graphstore.Record](t, call("show", sibling.ID)); !reflect.DeepEqual(got, sibling) {
				t.Fatal("literal canonical fallback changed sibling")
			}
		})
	}
}

func graphMixedSourceRecord(t *testing.T, value any) graphstore.Record {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var record graphstore.Record
	if err := json.Unmarshal(raw, &record); err != nil {
		t.Fatal(err)
	}
	if record.ID == "" || record.Revision == "" {
		t.Fatalf("missing owner: %s", raw)
	}
	return record
}
