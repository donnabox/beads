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
			refuse("invalid_selector", "remember", "No", "--update", memory.ID, "--if-revision", edited.Memory.Revision, "--unconditional")
			refuse("invalid_selector", "remember", "No", "--update", memory.ID, "--unconditional=false")
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
			refuse("invalid_selector", "update", link.Link.ID, "--properties", `{}`)
			changed := graphMixedResult[graphstore.LinkMutationResult](t, call("update", link.Link.ID, "--properties", `{"note":"current source"}`, "--if-revision", link.Link.Revision))
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
			call("unlink", unowned.Link.ID, "--if-revision", unowned.Link.Revision)
			if call("show", issue.ID) != issueBefore {
				t.Fatal("default informational unlink changed its Issue source")
			}
			// Resource/delete/Issue guards remain required explicit choices.
			refuse("invalid_selector", "delete", memory.ID, "--force")
			refuse("invalid_selector", "update", issue.ID, "--title", "No implicit Issue write")
			gate := graphMixedResult[graphstore.IssueRecord](t, call("create", "Prerequisite"))
			dependency := graphMixedResult[graphstore.DependencyResult](t, call("dep", "add", issue.ID, gate.ID))
			refuse("capability_unavailable", "unlink", issue.ID, gate.ID, "--resource-type", graphstore.DependencyTypeURL(scope), "--if-revision", dependency.Link.Revision)
			refuse("invalid_properties", "unlink", dependency.Link.ID, "--if-revision", dependency.Link.Revision)
			call("unlink", dependency.Link.ID, "--if-revision", dependency.Link.Revision, "--unconditional-source")
			propertyMemory := graphMixedResult[graphstore.Record](t, call("remember", "--id", "beads/property-memory", "--properties", `{"title":"Property title","body":"First body"}`))
			merged := graphMixedResult[graphstore.MemoryMutationResult](t, call("update", propertyMemory.ID, "--properties", `{"body":"Second body"}`, "--if-revision", propertyMemory.Revision))
			if merged.Memory.Properties.Title != "Property title" || merged.Memory.Properties.Body != "Second body" {
				t.Fatal("Memory property merge lost an omitted title")
			}
			emptyMerge := graphMixedResult[graphstore.MemoryMutationResult](t, call("update", propertyMemory.ID, "--properties", `{}`, "--if-revision", merged.Memory.Revision))
			if emptyMerge.Changed || emptyMerge.Memory.Revision != merged.Memory.Revision {
				t.Fatal("empty Memory property merge created a version")
			}
			rememberMerge := graphMixedResult[graphstore.MemoryMutationResult](t, call("remember", "--update", propertyMemory.ID, "--properties", `{"title":"Retitled"}`))
			if rememberMerge.Memory.Properties.Body != "Second body" || rememberMerge.Memory.Properties.Title != "Retitled" {
				t.Fatal("remember property merge lost the omitted body")
			}
			propertyLink := graphMixedResult[graphstore.LinkMutationResult](t, call("link", propertyMemory.ID, issue.ID, "--link-type", related, "--properties", `{"note":"keep"}`))
			linkEmptyMerge := graphMixedResult[graphstore.LinkMutationResult](t, call("update", propertyLink.Link.ID, "--properties", `{}`, "--if-revision", propertyLink.Link.Revision))
			if linkEmptyMerge.Changed || linkEmptyMerge.Link.Revision != propertyLink.Link.Revision || linkEmptyMerge.Link.Properties["note"] != "keep" {
				t.Fatal("empty Link property merge cleared a field or created a version")
			}
			linkCleared := graphMixedResult[graphstore.LinkMutationResult](t, call("update", propertyLink.Link.ID, "--patch", `[{"op":"replace","path":"","value":{}}]`, "--if-revision", propertyLink.Link.Revision))
			if !linkCleared.Changed || len(linkCleared.Link.Properties) != 0 {
				t.Fatal("explicit Link root patch did not clear properties")
			}
			propertyIssue := graphMixedResult[graphstore.IssueRecord](t, call("create", "--id", "beads/property-issue", "--properties", `{"title":"Property Issue","priority":1,"description":"Initial","estimated_minutes":7}`))
			if propertyIssue.Properties.Title != "Property Issue" || propertyIssue.Properties.Priority != 1 || propertyIssue.Properties.Description != "Initial" || propertyIssue.Properties.EstimatedMinutes == nil || *propertyIssue.Properties.EstimatedMinutes != 7 {
				t.Fatal("Issue property initialization lost typed values")
			}
			issueMerge := graphMixedResult[graphstore.IssueMutationResult](t, call("update", propertyIssue.ID, "--properties", `{"description":"Revised"}`, "--if-revision", propertyIssue.Revision))
			if !issueMerge.Changed || issueMerge.Issue.Properties.Title != "Property Issue" || issueMerge.Issue.Properties.Priority != 1 || issueMerge.Issue.Properties.Description != "Revised" {
				t.Fatal("Issue property merge lost an omitted field")
			}
			issueNoop := graphMixedResult[graphstore.IssueMutationResult](t, call("update", propertyIssue.ID, "--properties", `{}`, "--if-revision", issueMerge.Issue.Revision))
			if issueNoop.Changed || issueNoop.Issue.Revision != issueMerge.Issue.Revision {
				t.Fatal("empty Issue property merge created a version")
			}
			issuePatched := graphMixedResult[graphstore.IssueMutationResult](t, call("update", propertyIssue.ID, "--patch", `[{"op":"replace","path":"/description","value":"Patched"},{"op":"remove","path":"/estimated_minutes"}]`, "--if-revision", issueMerge.Issue.Revision))
			if !issuePatched.Changed || issuePatched.Issue.Properties.Description != "Patched" || issuePatched.Issue.Properties.EstimatedMinutes != nil {
				t.Fatal("Issue property patch bypassed native field semantics")
			}
			refuse("revision_conflict", "update", propertyIssue.ID, "--properties", `{"title":"stale"}`, "--if-revision", propertyIssue.Revision)
			refuse("invalid_properties", "update", propertyIssue.ID, "--properties", `{"status":"closed"}`, "--if-revision", issuePatched.Issue.Revision)
		})
	}
}
