//go:build cgo

package main

import (
	"encoding/json"
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
			args := []string{"init", "--graph-mode", "link", "--prefix", "compat", "--scope-url", scope, "--skip-hooks", "--skip-agents", "--non-interactive"}
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
			other := graphMixedResult[graphstore.MemoryMutationResult](t, call("remember", body)).Memory
			issue := graphMixedResult[graphstore.IssueRecord](t, call("create", "Update integration target"))
			if memory.ID != other.ID || memory.ID == issue.ID {
				t.Fatalf("derived Memory identity did not overwrite while Issue allocation stayed separate: Memory=%q repeated=%q Issue=%q", memory.ID, other.ID, issue.ID)
			}
			if want := scope + "code-flow-policy-target-integration-retain-the-previous"; memory.ID != want {
				t.Fatalf("legacy content derivation changed: got %q want %q", memory.ID, want)
			}
			if !strings.HasPrefix(issue.Properties.ID, "compat-") || issue.ID != scope+issue.Properties.ID {
				t.Fatalf("generated Issue path did not use the configured native prefix: %+v", issue)
			}
			parent := graphMixedResult[graphstore.IssueRecord](t, call("create", "Explicit parent", "--id", "compat-parent", "--type", "epic"))
			if parent.ID != scope+"compat-parent" || parent.Properties.ID != "compat-parent" {
				t.Fatalf("bare explicit Issue ID diverged from its native identity: %+v", parent)
			}
			child := graphMixedResult[graphstore.IssueRecord](t, call("create", "Child", "--parent", "compat-parent"))
			if child.ID != scope+"compat-parent.1" || child.Properties.ID != "compat-parent.1" {
				t.Fatalf("child did not retain the shipping parent ID: %+v", child)
			}
			_, _, exit := graphGenerationRun(t, bd, work, home, "create", "Wrong prefix", "--id", "other-parent", "--json")
			if exit == 0 {
				t.Fatal("bare explicit Issue ID bypassed the native prefix rule")
			}
			refuse("not_found", "show", "other-parent")
			forced := graphMixedResult[graphstore.IssueRecord](t, call("create", "Forced prefix", "--id", "other-parent", "--force"))
			if forced.ID != scope+"other-parent" || forced.Properties.ID != "other-parent" {
				t.Fatalf("forced bare Issue ID diverged from shipping identity: %+v", forced)
			}
			if memory.Properties.Title != "Code flow policy" || memory.Properties.Body != body {
				t.Fatal("default title changed body or failed to summarize")
			}
			fixed := graphMixedResult[graphstore.Record](t, call("remember", "Body", "--id", "beads/fixed", "--title", "  Explicit title  "))
			if fixed.ID != scope+"beads/fixed" || fixed.Properties.Title != "  Explicit title  " {
				t.Fatal("explicit identity/title not preserved")
			}
			explicitIssue := graphMixedResult[graphstore.IssueRecord](t, call("create", "Explicit Issue", "--graph-path", "beads/fixed-issue"))
			if explicitIssue.ID != scope+explicitIssue.Properties.ID || explicitIssue.ID == scope+"beads/fixed-issue" {
				t.Fatalf("path-created Issue did not expose its distinct public identity: %+v", explicitIssue)
			}
			literalIssue := graphMixedResult[graphstore.IssueRecord](t, call("create", "Literal Issue ID", "--id", "beads/literal-issue", "--type", "epic", "--force"))
			if literalIssue.Properties.ID != "beads/literal-issue" || literalIssue.ID != scope+"beads/literal-issue" {
				t.Fatalf("slash-bearing --id was not kept literal: %+v", literalIssue)
			}
			if got := graphMixedResult[graphstore.IssueRecord](t, call("show", "beads/literal-issue", "--format", "graph-json")); got.ID != literalIssue.ID || got.Properties.ID != literalIssue.Properties.ID {
				t.Fatalf("local slash-bearing ID selected a graph path instead of the public Issue: %+v", got)
			}
			if got := graphMixedResult[graphstore.IssueRecord](t, call("show", literalIssue.ID, "--format", "graph-json")); got.ID != literalIssue.ID {
				t.Fatalf("canonical Scope URL did not select its exact public identity: %+v", got)
			}
			if out := graphPolicyCLI(t, bd, work, home, nil, "", "show", "beads/literal-issue", "--output-format=json"); !strings.Contains(out, literalIssue.ID) {
				t.Fatalf("canonical output did not resolve the literal public ID: %s", out)
			}
			literalChild := graphMixedResult[graphstore.IssueRecord](t, call("create", "Literal child", "--parent", "beads/literal-issue", "--force"))
			if literalChild.Properties.ID != "beads/literal-issue.1" || literalChild.ID != scope+"beads/literal-issue.1" {
				t.Fatalf("slash-bearing parent ID did not retain dotted child identity: %+v", literalChild)
			}
			if many := call("show", "beads/literal-issue", "--id=beads/literal-issue.1"); !strings.Contains(many, `"id": "beads/literal-issue"`) || !strings.Contains(many, `"id": "beads/literal-issue.1"`) {
				t.Fatalf("multi-ID show did not resolve literal parent and child IDs: %s", many)
			}
			if children := call("show", "beads/literal-issue", "--children"); !strings.Contains(children, literalChild.Properties.ID) {
				t.Fatalf("--children did not resolve the literal parent ID: %s", children)
			}
			graphMixedResult[graphstore.IssueRecord](t, call("create", "Literal ID distinct from path", "--id", "beads/fixed-issue", "--force"))
			graphMixedResult[graphstore.IssueRecord](t, call("create", "Graph path distinct from literal ID", "--graph-path", "beads/literal-issue"))
			collision := graphMixedResult[graphstore.Record](t, call("remember", "Existing Memory", "--id", "beads/compat-collision"))
			coexistingIssue := graphMixedResult[graphstore.IssueRecord](t, call("create", "Distinct opaque ID", "--id", "compat-collision"))
			if coexistingIssue.Properties.ID != "compat-collision" || coexistingIssue.ID != scope+"compat-collision" {
				t.Fatalf("different opaque IDs did not retain their own identity: %+v", coexistingIssue)
			}
			if got := graphMixedResult[graphstore.Record](t, call("show", collision.ID)); !reflect.DeepEqual(got, collision) {
				t.Fatal("coexisting Issue changed the existing Memory")
			}
			if got := graphMixedResult[graphstore.IssueRecord](t, call("show", coexistingIssue.ID, "--format", "graph-json")); got.Properties.ID != coexistingIssue.Properties.ID {
				t.Fatal("coexisting Issue cannot be selected by its canonical URL")
			}
			refuse("identity_reserved", "remember", "Overwrite", "--id", "beads/fixed", "--create-only")
			refuse("identity_reserved", "create", "Overwrite", "--graph-path", "beads/fixed-issue")
			refuse("identity_reserved", "create", "Duplicate literal", "--id", "beads/literal-issue", "--force")
			refuse("identity_reserved", "create", "Wrong kind", "--id", "beads/fixed", "--force")
			refuse("identity_reserved", "remember", "Wrong kind", "--id", "beads/fixed-issue")
			refuse("invalid_selector", "remember", "Body", "--id=")
			refuse("invalid_selector", "create", "Title", "--id=")
			refuse("invalid_selector", "create", "Both", "--id", "compat-both", "--graph-path", "beads/both")
			refuse("invalid_selector", "create", "Bad path", "--graph-path", "links/not-a-bead")
			refuse("invalid_properties", "remember", "Body", "--title=")
			if got := graphMixedResult[graphstore.Record](t, call("show", fixed.ID)); !reflect.DeepEqual(got, fixed) {
				t.Fatal("refused creation changed existing Memory")
			}
			upserted := graphMixedResult[graphstore.MemoryMutationResult](t, call("remember", "Replacement body", "--id", "beads/fixed"))
			if !upserted.Changed || upserted.Replaced == nil || upserted.Memory.ID != fixed.ID || upserted.Memory.Properties.Title != fixed.Properties.Title || upserted.Memory.Properties.Body != "Replacement body" {
				t.Fatal("existing-ID remember failed to preserve title and disclose predecessor")
			}
			guardedUpsert := graphMixedResult[graphstore.MemoryMutationResult](t, call("remember", "--id", "beads/fixed", "--title", "Current policy", "--if-revision", "="+upserted.Memory.Revision))
			if !guardedUpsert.Changed || guardedUpsert.Replaced != nil || guardedUpsert.Memory.Properties.Body != "Replacement body" {
				t.Fatal("guarded existing-ID title edit lost the omitted body or claimed an unconditional predecessor")
			}
			refuse("identity_reserved", "remember", "Another body", "--id", "beads/fixed", "--create-only")
			refuse("revision_conflict", "remember", "Stale body", "--id", "beads/fixed", "--if-revision", "="+fixed.Revision)
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
			refuse("revision_conflict", "remember", "New body", "--update", memory.ID, "--if-revision", "="+memory.Revision)
			refuse("invalid_selector", "remember", "No", "--update", memory.ID, "--if-revision=")
			titled := graphMixedResult[graphstore.MemoryMutationResult](t, call("remember", "--update", memory.ID, "--title", "New title", "--if-revision", "="+edited.Memory.Revision))
			if titled.Memory.Properties.Body != "New body" || titled.Replaced != nil {
				t.Fatal("guarded title edit lost omitted body")
			}

			// Subject To exercises owned-source guards; References below exercises unowned Links.
			constraint := graphMixedResult[graphstore.Record](t, call("create", "First constraint", "--type", "constraint", "--id", "constraint-one"))
			secondConstraint := graphMixedResult[graphstore.Record](t, call("create", "Second constraint", "--type", "constraint", "--id", "constraint-two"))
			related := scope + "subject-to"
			link := graphMixedResult[graphstore.LinkMutationResult](t, call("link", memory.ID, constraint.ID, "--type", related))
			if link.ReplacedSource == nil {
				t.Fatal("default Link creation lost source predecessor")
			}
			before := call("show", memory.ID)
			refuse("revision_conflict", "link", memory.ID, secondConstraint.ID, "--type", related, "--if-source-revision", titled.Memory.Revision)
			refuse("invalid_selector", "link", memory.ID, secondConstraint.ID, "--type", related, "--if-source-revision=")
			if call("show", memory.ID) != before {
				t.Fatal("stale/conflicting source guards changed Memory")
			}
			mixed := graphMixedResult[graphstore.LinkMutationResult](t, call("link", memory.ID, secondConstraint.ID, "--type", related))
			changed := graphMixedResult[graphstore.LinkMutationResult](t, call("update", link.Link.ID, "--properties", `{"note":"current source"}`))
			if !changed.Changed || changed.ReplacedSource == nil {
				t.Fatal("default source guard did not support Link replacement")
			}
			linkNoop := graphMixedResult[graphstore.LinkMutationResult](t, call("update", link.Link.ID, "--properties", `{"note":"current source"}`, "--if-revision", "="+changed.Link.Revision))
			if linkNoop.Changed || linkNoop.ReplacedSource != nil || !reflect.DeepEqual(linkNoop.Link, changed.Link) {
				t.Fatal("default source acceptance disclosed or versioned a no-op")
			}
			patched := graphMixedResult[graphstore.LinkMutationResult](t, call("update", link.Link.ID, "--patch", `[{"op":"replace","path":"/note","value":"patched"}]`, "--if-revision", "="+changed.Link.Revision))
			if !patched.Changed || patched.ReplacedSource == nil {
				t.Fatal("default source guard did not support ordered Link patch")
			}
			refuse("revision_conflict", "update", link.Link.ID, "--properties", `{}`, "--if-revision", "="+link.Link.Revision)
			preview := graphMixedResult[map[string]any](t, call("link", "delete", mixed.Link.ID))
			if preview["preview"] != true {
				t.Fatal("Link deletion did not preview by default")
			}
			call("link", "delete", "--force", mixed.Link.ID, "--if-revision", "="+mixed.Link.Revision)
			call("link", "delete", "--force", link.Link.ID, "--if-revision", "="+patched.Link.Revision)
			issueBefore := call("show", issue.ID)
			unowned := graphMixedResult[graphstore.LinkMutationResult](t, call("link", issue.ID, memory.ID, "--type", "references"))
			if unowned.ReplacedSource != nil || call("show", issue.ID) != issueBefore {
				t.Fatal("default informational Link changed or disclosed its Issue source")
			}
			human := graphPolicyCLI(t, bd, work, home, nil, "", "link", "list", "--incident-to", issue.ID)
			if !strings.Contains(human, strings.TrimPrefix(unowned.Link.ID, scope)+" [references] "+strings.TrimPrefix(issue.ID, scope)+" → "+strings.TrimPrefix(memory.ID, scope)) {
				t.Fatalf("links omitted identity-first local display: %q", human)
			}
			call("link", "delete", "--force", unowned.Link.ID, "--if-revision", "="+unowned.Link.Revision)
			localLinkID := "links/canonical"
			canonicalID := scope + localLinkID
			refuse("invalid_selector", "link", issue.ID, other.ID, "--type", "references", "--id", canonicalID)
			canonical := graphMixedResult[graphstore.LinkMutationResult](t, call("link", issue.ID, other.ID, "--type", "references", "--id", localLinkID))
			if canonical.Link.ID != canonicalID {
				t.Fatal("informational Link did not preserve the literal local ID in its canonical URL")
			}
			call("link", "delete", "--force", canonicalID, "--if-revision", "="+canonical.Link.Revision)
			if call("show", issue.ID) != issueBefore {
				t.Fatal("default informational unlink changed its Issue source")
			}
			// Destructive operations accept the current revision when no guard is supplied.
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
			sameExplicit := graphMixedResult[graphstore.DependencyResult](t, call("dep", "add", issue.ID, otherBlocker.ID, "--id", "links/chosen-block"))
			if sameExplicit.Changed || sameExplicit.Link.ID != explicit.Link.ID {
				t.Fatal("dep add --id did not preserve the matching pair as a no-op")
			}
			refuse("invalid_properties", "dep", "add", issue.ID, otherBlocker.ID, "--id", "links/different-block")
			thirdBlocker := graphMixedResult[graphstore.IssueRecord](t, call("create", "Third prerequisite"))
			refuse("identity_reserved", "dep", "add", issue.ID, thirdBlocker.ID, "--id", "links/chosen-block")
			// The shipping positional form remains a Link creation route. Its
			// first Issue is blocked by the second, and repeating the same pair
			// must not allocate another Link or Issue version.
			decodeShorthand := func(raw string) map[string]any {
				t.Helper()
				var output map[string]any
				if err := json.Unmarshal([]byte(raw), &output); err != nil {
					t.Fatalf("decode familiar positional link: %v\n%s", err, raw)
				}
				return output
			}
			shorthandBlock := decodeShorthand(call("link", issue.ID, thirdBlocker.ID))
			if shorthandBlock["status"] != "added" || shorthandBlock["issue_id"] != issue.Properties.ID || shorthandBlock["depends_on_id"] != thirdBlocker.Properties.ID || shorthandBlock["type"] != "blocks" || shorthandBlock["schema_version"] != float64(1) {
				t.Fatalf("positional link lost shipping JSON fields or pair direction: %+v", shorthandBlock)
			}
			beforeShorthandNoop := call("versions", issue.ID)
			beforeShorthandLinks := call("link", "list", "--from", issue.ID, "--type", "blocks")
			if shorthandNoop := decodeShorthand(call("link", issue.ID, thirdBlocker.ID)); !reflect.DeepEqual(shorthandNoop, shorthandBlock) {
				t.Fatalf("repeated positional link changed its familiar result: %+v", shorthandNoop)
			}
			humanShorthand := graphPolicyCLI(t, bd, work, home, nil, "", "link", issue.ID, thirdBlocker.ID)
			if !strings.Contains(humanShorthand, "Linked:") || !strings.Contains(humanShorthand, issue.Properties.ID) || !strings.Contains(humanShorthand, thirdBlocker.Properties.ID) || !strings.Contains(humanShorthand, "(blocks)") {
				t.Fatalf("positional link lost familiar human feedback: %s", humanShorthand)
			}
			if afterShorthandNoop := call("versions", issue.ID); afterShorthandNoop != beforeShorthandNoop {
				t.Fatal("repeated positional link allocated another Issue version")
			}
			if afterShorthandLinks := call("link", "list", "--from", issue.ID, "--type", "blocks"); afterShorthandLinks != beforeShorthandLinks {
				t.Fatal("repeated positional link allocated another Link")
			}
			refuse("invalid_properties", "link", issue.ID, thirdBlocker.ID, "--type", "tracks")
			shorthandTracks := decodeShorthand(call("link", otherBlocker.ID, thirdBlocker.ID, "--type", "tracks"))
			if shorthandTracks["issue_id"] != otherBlocker.Properties.ID || shorthandTracks["depends_on_id"] != thirdBlocker.Properties.ID || shorthandTracks["type"] != "tracks" {
				t.Fatalf("positional link lost its explicit Dependency kind: %+v", shorthandTracks)
			}
			call("link", "delete", "--force", dependency.Link.ID, "--if-revision", "="+dependency.Link.Revision)
			propertyMemory := graphMixedResult[graphstore.Record](t, call("remember", "--id", "beads/property-memory", "--properties", `{"title":"Property title","body":"First body"}`))
			merged := graphMixedResult[graphstore.MemoryMutationResult](t, call("update", propertyMemory.ID, "--properties", `{"body":"Second body"}`, "--if-revision", "="+propertyMemory.Revision))
			if merged.Memory.Properties.Title != "Property title" || merged.Memory.Properties.Body != "Second body" {
				t.Fatal("Memory property merge lost an omitted title")
			}
			emptyMerge := graphMixedResult[graphstore.MemoryMutationResult](t, call("update", propertyMemory.ID, "--properties", `{}`, "--if-revision", "="+merged.Memory.Revision))
			if emptyMerge.Changed || emptyMerge.Memory.Revision != merged.Memory.Revision {
				t.Fatal("empty Memory property merge created a version")
			}
			rememberMerge := graphMixedResult[graphstore.MemoryMutationResult](t, call("remember", "--update", propertyMemory.ID, "--properties", `{"title":"Retitled"}`))
			if rememberMerge.Memory.Properties.Body != "Second body" || rememberMerge.Memory.Properties.Title != "Retitled" {
				t.Fatal("remember property merge lost the omitted body")
			}
			propertyLink := graphMixedResult[graphstore.LinkMutationResult](t, call("link", propertyMemory.ID, constraint.ID, "--type", related, "--properties", `{"note":"keep"}`))
			linkEmptyMerge := graphMixedResult[graphstore.LinkMutationResult](t, call("update", propertyLink.Link.ID, "--properties", `{}`, "--if-revision", "="+propertyLink.Link.Revision))
			if linkEmptyMerge.Changed || linkEmptyMerge.Link.Revision != propertyLink.Link.Revision || linkEmptyMerge.Link.Properties["note"] != "keep" {
				t.Fatal("empty Link property merge cleared a field or created a version")
			}
			linkCleared := graphMixedResult[graphstore.LinkMutationResult](t, call("update", propertyLink.Link.ID, "--patch", `[{"op":"replace","path":"","value":{}}]`, "--if-revision", "="+propertyLink.Link.Revision))
			if !linkCleared.Changed || len(linkCleared.Link.Properties) != 0 {
				t.Fatal("explicit Link root patch did not clear properties")
			}
			propertyIssue := graphMixedResult[graphstore.IssueRecord](t, call("create", "--graph-path", "beads/property-issue", "--properties", `{"title":"Property Issue","priority":1,"description":"Initial","estimated_minutes":7}`))
			if propertyIssue.Properties.Title != "Property Issue" || propertyIssue.Properties.Priority != 1 || propertyIssue.Properties.Description != "Initial" || propertyIssue.Properties.EstimatedMinutes == nil || *propertyIssue.Properties.EstimatedMinutes != 7 {
				t.Fatal("Issue property initialization lost typed values")
			}
			issueMerge := graphMixedResult[graphstore.IssueMutationResult](t, call("update", propertyIssue.ID, "--properties", `{"description":"Revised"}`, "--if-revision", "="+propertyIssue.Revision))
			if !issueMerge.Changed || issueMerge.Issue.Properties.Title != "Property Issue" || issueMerge.Issue.Properties.Priority != 1 || issueMerge.Issue.Properties.Description != "Revised" {
				t.Fatal("Issue property merge lost an omitted field")
			}
			issueNoop := graphMixedResult[graphstore.IssueMutationResult](t, call("update", propertyIssue.ID, "--properties", `{}`, "--if-revision", "="+issueMerge.Issue.Revision))
			if issueNoop.Changed || issueNoop.Issue.Revision != issueMerge.Issue.Revision {
				t.Fatal("empty Issue property merge created a version")
			}
			issuePatched := graphMixedResult[graphstore.IssueMutationResult](t, call("update", propertyIssue.ID, "--patch", `[{"op":"replace","path":"/description","value":"Patched"},{"op":"remove","path":"/estimated_minutes"}]`, "--if-revision", "="+issueMerge.Issue.Revision))
			if !issuePatched.Changed || issuePatched.Issue.Properties.Description != "Patched" || issuePatched.Issue.Properties.EstimatedMinutes != nil {
				t.Fatal("Issue property patch bypassed native field semantics")
			}
			withNotes := graphMixedResult[graphstore.IssueMutationResult](t, call("update", propertyIssue.ID, "--properties", `{"notes":"first"}`))
			refuse("notes_overwrite_refused", "update", propertyIssue.ID, "--properties", `{"notes":"second"}`)
			forcedNotes := graphMixedResult[graphstore.IssueMutationResult](t, call("update", propertyIssue.ID, "--properties", `{"notes":"second"}`, "--force"))
			if !withNotes.Changed || !forcedNotes.Changed || forcedNotes.Issue.Properties.Notes != "second" {
				t.Fatal("Issue notes property merge did not preserve the native overwrite fence")
			}
			forcedPatch := graphMixedResult[graphstore.IssueMutationResult](t, call("update", propertyIssue.ID, "--patch", `[{"op":"replace","path":"/notes","value":"third"}]`, "--force"))
			if !forcedPatch.Changed || forcedPatch.Issue.Properties.Notes != "third" {
				t.Fatal("Issue notes property patch did not honor deliberate force")
			}
			refuse("revision_conflict", "update", propertyIssue.ID, "--properties", `{"title":"stale"}`, "--if-revision", "="+propertyIssue.Revision)
			refuse("invalid_properties", "update", propertyIssue.ID, "--properties", `{"status":"closed"}`, "--if-revision", "="+issuePatched.Issue.Revision)
		})
	}
}

func TestGraphPreviewPositionalLinkPartialIDsWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			const scope = "https://example.invalid/link-partial/"
			initArgs := []string{"init", "--graph-mode", "link", "--prefix", "compat", "--scope-url", scope, "--skip-hooks", "--skip-agents", "--non-interactive"}
			if engine == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for ordinary shared-server CLI qualification")
				}
				initArgs = append(initArgs, "--server", "--external", "--server-host", "127.0.0.1", "--server-port", port, "--server-user", "root")
			}
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			call(initArgs...)
			source := graphMixedResult[graphstore.IssueRecord](t, call("create", "Source"))
			target := graphMixedResult[graphstore.IssueRecord](t, call("create", "Target"))
			abbreviate := func(id string) string {
				t.Helper()
				if !strings.HasPrefix(id, "compat-") || len(id) < len("compat-")+2 {
					t.Fatalf("unexpected generated native Issue ID %q", id)
				}
				return id[len("compat-") : len(id)-1]
			}
			from, to := abbreviate(source.Properties.ID), abbreviate(target.Properties.ID)
			decodeLink := func(raw string) map[string]any {
				t.Helper()
				var result map[string]any
				if err := json.Unmarshal([]byte(raw), &result); err != nil {
					t.Fatalf("decode familiar positional Link JSON: %v\n%s", err, raw)
				}
				return result
			}
			result := decodeLink(call("link", from, to))
			if result["issue_id"] != source.Properties.ID || result["depends_on_id"] != target.Properties.ID || result["type"] != "blocks" {
				t.Fatalf("positional partial Link resolved wrong Issues: %+v", result)
			}
			if repeated := decodeLink(call("link", from, to)); !reflect.DeepEqual(repeated, result) {
				t.Fatalf("repeated positional Link changed familiar result: %+v", repeated)
			}
			var shown []struct {
				Dependencies []json.RawMessage `json:"dependencies"`
			}
			if raw := call("show", source.ID); json.Unmarshal([]byte(raw), &shown) != nil || len(shown) != 1 || len(shown[0].Dependencies) != 1 {
				t.Fatalf("positional Link did not preserve one native relation: %s", raw)
			}
			other := graphMixedResult[graphstore.IssueRecord](t, call("create", "Other blocker"))
			call("dep", "add", from, "--blocked-by", abbreviate(other.Properties.ID))
			if raw := call("show", source.ID); json.Unmarshal([]byte(raw), &shown) != nil || len(shown) != 1 || len(shown[0].Dependencies) != 2 || !strings.Contains(raw, other.Properties.ID) {
				t.Fatalf("dep add partial IDs did not preserve the second native relation: %s", raw)
			}
			fourth := graphMixedResult[graphstore.IssueRecord](t, call("create", "Fourth blocker"))
			call("dep", abbreviate(fourth.Properties.ID), "--blocks", from)
			if raw := call("show", source.ID); json.Unmarshal([]byte(raw), &shown) != nil || len(shown) != 1 || len(shown[0].Dependencies) != 3 || !strings.Contains(raw, fourth.Properties.ID) {
				t.Fatalf("dep --blocks partial IDs did not preserve the third native relation: %s", raw)
			}
			plain := graphMixedResult[graphstore.IssueRecord](t, call("create", "Plain", "--id", "x", "--force"))
			slash := graphMixedResult[graphstore.IssueRecord](t, call("create", "Slash", "--id", "beads/x", "--force"))
			if plain.ID == slash.ID {
				t.Fatal("distinct public IDs shared one graph path")
			}
			if linked := decodeLink(call("link", "beads/x", target.Properties.ID)); linked["issue_id"] != "beads/x" || linked["depends_on_id"] != target.Properties.ID {
				t.Fatalf("positional Link treated slash ID as another Issue's graph path: %+v", linked)
			}
			shown = nil
			if raw := call("show", "x"); json.Unmarshal([]byte(raw), &shown) != nil || len(shown) != 1 || len(shown[0].Dependencies) != 0 {
				t.Fatalf("linking beads/x changed x: %s", raw)
			}
			shown = nil
			if raw := call("show", "beads/x"); json.Unmarshal([]byte(raw), &shown) != nil || len(shown) != 1 || len(shown[0].Dependencies) != 1 {
				t.Fatalf("positional slash-ID Link missed its source: %s", raw)
			}
		})
	}
}
