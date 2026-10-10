//go:build cgo

package main

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// Exercise defaults through installed processes and both real engines. Explicit
// optimistic guards, identity reservation, and omitted-field preservation stay
// observable alongside the compatibility path.
func TestGraphPreviewCompatibilityDefaultsWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			const scope = "https://example.invalid/compatibility/"
			args := []string{"init", "--graph-mode", "link", "--scope-url", scope, "--skip-hooks", "--skip-agents", "--non-interactive"}
			if engine == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for ordinary shared-server CLI qualification")
				}
				args = append(args, "--server", "--external", "--server-host", "127.0.0.1", "--server-port", port, "--server-user", "root")
			}
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			refuse := func(code string, args ...string) {
				t.Helper()
				graphPolicyCLI(t, bd, work, home, nil, code, append(args, "--json")...)
			}
			call(args...)
			const body = "  Code  flow policy\nTarget integration; retain the previous policy."
			memory := graphMixedResult[graphstore.Record](t, call("remember", body))
			other := graphMixedResult[graphstore.Record](t, call("remember", body))
			issue := graphMixedResult[graphstore.IssueRecord](t, call("create", "Update integration target"))
			if memory.ID == other.ID || memory.ID == issue.ID || other.ID == issue.ID || !strings.HasPrefix(memory.ID, scope+"beads/") {
				t.Fatal("omitted IDs did not allocate distinct canonical identities")
			}
			if memory.Properties.Title != "Code flow policy" || memory.Properties.Body != body {
				t.Fatal("default title changed body or failed to summarize")
			}
			fixed := graphMixedResult[graphstore.Record](t, call("remember", "Body", "--id", "beads/fixed", "--title", "  Explicit title  "))
			if fixed.ID != scope+"beads/fixed" || fixed.Properties.Title != "  Explicit title  " {
				t.Fatal("explicit identity/title not preserved")
			}
			call("create", "Explicit Issue", "--id", "beads/fixed-issue")
			refuse("identity_reserved", "remember", "Overwrite", "--id", "beads/fixed", "--create-only")
			refuse("identity_reserved", "create", "Overwrite", "--id", "beads/fixed-issue")
			refuse("identity_reserved", "create", "Wrong kind", "--id", "beads/fixed")
			refuse("identity_reserved", "remember", "Wrong kind", "--id", "beads/fixed-issue")
			refuse("invalid_selector", "remember", "Body", "--id=")
			refuse("invalid_selector", "create", "Title", "--id=")
			refuse("invalid_properties", "remember", "Body", "--title=")
			if got := graphMixedResult[graphstore.Record](t, call("show", fixed.ID)); !reflect.DeepEqual(got, fixed) {
				t.Fatal("refused creation changed existing Memory")
			}
			upserted := graphMixedResult[graphstore.MemoryMutationResult](t, call("remember", "Replacement body", "--id", "beads/fixed"))
			if !upserted.Changed || upserted.Replaced == nil || upserted.Memory.ID != fixed.ID || upserted.Memory.Properties.Title != fixed.Properties.Title || upserted.Memory.Properties.Body != "Replacement body" {
				t.Fatal("existing-ID remember failed to preserve title and disclose predecessor")
			}
			guardedUpsert := graphMixedResult[graphstore.MemoryMutationResult](t, call("remember", "--id", "beads/fixed", "--title", "Current policy", "--if-revision", upserted.Memory.Revision))
			if !guardedUpsert.Changed || guardedUpsert.Replaced != nil || guardedUpsert.Memory.Properties.Body != "Replacement body" {
				t.Fatal("guarded existing-ID title edit lost the omitted body or claimed an unconditional predecessor")
			}
			refuse("identity_reserved", "remember", "Another body", "--id", "beads/fixed", "--create-only")
			refuse("revision_conflict", "remember", "Stale body", "--id", "beads/fixed", "--if-revision", fixed.Revision)
			if current := graphMixedResult[graphstore.Record](t, call("show", fixed.ID)); !reflect.DeepEqual(current, guardedUpsert.Memory) {
				t.Fatal("create-only or stale guard changed existing Memory")
			}

			edited := graphMixedResult[graphstore.MemoryMutationResult](t, call("remember", "New body", "--update", memory.ID))
			if !edited.Changed || edited.Replaced == nil || edited.Memory.Properties.Title != memory.Properties.Title || edited.Memory.Properties.Body != "New body" {
				t.Fatal("default edit lost omitted title or predecessor disclosure")
			}
			noop := graphMixedResult[graphstore.MemoryMutationResult](t, call("remember", "New body", "--update", memory.ID))
			if noop.Changed || noop.Replaced != nil || !reflect.DeepEqual(noop.Memory, edited.Memory) {
				t.Fatal("default edit created a no-op version")
			}
			refuse("revision_conflict", "remember", "New body", "--update", memory.ID, "--if-revision", memory.Revision)
			refuse("invalid_selector", "remember", "No", "--update", memory.ID, "--if-revision=")
			titled := graphMixedResult[graphstore.MemoryMutationResult](t, call("remember", "--update", memory.ID, "--title", "New title", "--if-revision", edited.Memory.Revision))
			if titled.Memory.Properties.Body != "New body" || titled.Replaced != nil {
				t.Fatal("guarded title edit lost omitted body")
			}

			related := scope + "types/preview-related-v2"
			link := graphMixedResult[graphstore.LinkMutationResult](t, call("link", memory.ID, other.ID, "--resource-type", related))
			if link.ReplacedSource == nil {
				t.Fatal("default Link creation lost source predecessor")
			}
			before := call("show", memory.ID)
			refuse("revision_conflict", "link", memory.ID, issue.ID, "--resource-type", related, "--if-source-revision", titled.Memory.Revision)
			refuse("invalid_selector", "link", memory.ID, issue.ID, "--resource-type", related, "--if-source-revision", titled.Memory.Revision, "--unconditional-source")
			refuse("invalid_selector", "link", memory.ID, issue.ID, "--resource-type", related, "--unconditional-source=false")
			if call("show", memory.ID) != before {
				t.Fatal("stale/conflicting source guards changed Memory")
			}
			mixed := graphMixedResult[graphstore.LinkMutationResult](t, call("link", memory.ID, issue.ID, "--resource-type", related))
			changed := graphMixedResult[graphstore.LinkMutationResult](t, call("update", link.Link.ID, "--properties", `{"note":"current source"}`))
			if !changed.Changed || changed.ReplacedSource == nil {
				t.Fatal("default source guard did not support Link replacement")
			}
			linkNoop := graphMixedResult[graphstore.LinkMutationResult](t, call("update", link.Link.ID, "--properties", `{"note":"current source"}`, "--if-revision", changed.Link.Revision))
			if linkNoop.Changed || linkNoop.ReplacedSource != nil || !reflect.DeepEqual(linkNoop.Link, changed.Link) {
				t.Fatal("default source acceptance disclosed or versioned a no-op")
			}
			patched := graphMixedResult[graphstore.LinkMutationResult](t, call("update", link.Link.ID, "--patch", `[{"op":"replace","path":"/note","value":"patched"}]`, "--if-revision", changed.Link.Revision))
			if !patched.Changed || patched.ReplacedSource == nil {
				t.Fatal("default source guard did not support ordered Link patch")
			}
			refuse("revision_conflict", "update", link.Link.ID, "--properties", `{}`, "--if-revision", link.Link.Revision)
			refuse("invalid_selector", "unlink", mixed.Link.ID)
			call("unlink", memory.ID, issue.ID, "--resource-type", related, "--if-revision", mixed.Link.Revision)
			call("unlink", link.Link.ID, "--if-revision", patched.Link.Revision)
			issueBefore := call("show", issue.ID)
			unowned := graphMixedResult[graphstore.LinkMutationResult](t, call("link", issue.ID, memory.ID, "--resource-type", related))
			if unowned.ReplacedSource != nil || call("show", issue.ID) != issueBefore {
				t.Fatal("default informational Link changed or disclosed its Issue source")
			}
			human := graphPolicyCLI(t, bd, work, home, nil, "", "links", issue.ID)
			if !strings.Contains(human, "types/preview-related-v2  "+strings.TrimPrefix(unowned.Link.ID, scope)+"  "+strings.TrimPrefix(issue.ID, scope)+" → "+strings.TrimPrefix(memory.ID, scope)) {
				t.Fatalf("links omitted type-first local display: %q", human)
			}
			call("unlink", unowned.Link.ID, "--if-revision", unowned.Link.Revision)
			if call("show", issue.ID) != issueBefore {
				t.Fatal("default informational unlink changed its Issue source")
			}
			// Destructive operations retain their explicit revision choice.
			refuse("invalid_selector", "delete", memory.ID, "--force")
			updatedIssue := graphMixedResult[graphstore.IssueMutationResult](t, call("update", issue.ID, "--title", "Issue edits accept current"))
			if updatedIssue.Issue.Properties.Title != "Issue edits accept current" {
				t.Fatal("ordinary Issue update did not accept the current predecessor")
			}
			gate := graphMixedResult[graphstore.IssueRecord](t, call("create", "Prerequisite"))
			dependency := graphMixedResult[graphstore.DependencyResult](t, call("dep", "add", issue.ID, gate.ID))
			shortcut := graphMixedResult[graphstore.DependencyResult](t, call("dep", gate.ID, "--blocks", issue.ID))
			if shortcut.Changed || shortcut.Link.ID != dependency.Link.ID {
				t.Fatal("dep --blocks did not preserve the existing blocking Link as a no-op")
			}
			otherBlocker := graphMixedResult[graphstore.IssueRecord](t, call("create", "Another prerequisite"))
			explicit := graphMixedResult[graphstore.DependencyResult](t, call("dep", "add", issue.ID, otherBlocker.ID, "--id", "links/chosen-block"))
			if explicit.Link.ID != scope+"links/chosen-block" {
				t.Fatal("dep add --id did not allocate the selected Link")
			}
			refuse("capability_unavailable", "unlink", issue.ID, gate.ID, "--resource-type", graphstore.DependencyTypeURL(scope), "--if-revision", dependency.Link.Revision)
			refuse("invalid_properties", "unlink", dependency.Link.ID, "--if-revision", dependency.Link.Revision)
			call("unlink", dependency.Link.ID, "--if-revision", dependency.Link.Revision, "--unconditional-source")
		})
	}
}
