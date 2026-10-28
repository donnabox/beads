//go:build cgo

package main

import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// Author all state using installed commands. Every call is a separate process;
// inspection is read-only and never supplies fixture rows or bootstrap schema.
func TestGraphPreviewIssueAssignmentWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			const scope = "https://example.invalid/assignment/"
			args := []string{"init", "--graph-mode", "link", "--scope-url", scope, "--skip-hooks", "--skip-agents", "--non-interactive"}
			if engine == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for ordinary shared-server assignment qualification")
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
			list := func(want []string, extra ...string) {
				t.Helper()
				args := append([]string{"list", "--format", "records-json", "--type", "issue", "--all"}, extra...)
				page := graphMixedResult[graphstore.IssueListPage](t, graphPolicyCLI(t, bd, work, home, nil, "", args...))
				got := make([]string, 0, len(page.Items))
				for _, item := range page.Items {
					got = append(got, item.ID)
				}
				slices.Sort(got)
				if page.Items == nil || page.HasMore || !slices.Equal(got, want) {
					t.Fatalf("list %v: ids=%v more=%v want=%v", extra, got, page.HasMore, want)
				}
			}
			exact := func(record graphstore.IssueRecord) {
				t.Helper()
				got := graphMixedResult[graphstore.IssueRecord](t, call("show", record.ID, "--revision", "="+record.Revision))
				properties := *record.Properties
				properties.ContentHash, properties.RowVersion = "", 0
				record.Properties = &properties
				if !reflect.DeepEqual(got, record) {
					t.Fatalf("exact retained Issue differs: got=%+v want=%+v", got, record)
				}
			}
			call(args...)
			memory := call("remember", "Assignment context", "--id", "beads/context", "--title", "Context")
			call("create", "Prioritize work", "--id", "beads/work", "--force", "--description", "Untouched description")
			gate := call("create", "Prerequisite", "--id", "beads/gate", "--force")
			dependency := graphMixedResult[graphstore.DependencyResult](t, call("dep", "add", scope+"beads/work", scope+"beads/gate"))
			informational := call("link", scope+"beads/work", "beads/context", "--id", "links/context", "--type", scope+"references")
			original := graphMixedResult[graphstore.IssueRecord](t, call("show", "--format", "graph-json", scope+"beads/work"))
			edited := graphMixedResult[graphstore.IssueMutationResult](t, call("update", scope+"beads/work", "--priority", "P0", "--assignee", "alice", "--title", "Assigned work", "--if-revision", "="+original.Revision))
			current := edited.Issue
			if !edited.Changed || current.Properties.Priority != 0 || current.Properties.Assignee != "alice" || current.Properties.Title != "Assigned work" || current.Properties.Description != original.Properties.Description || current.Properties.Status != original.Properties.Status || current.Revision == original.Revision || !reflect.DeepEqual(current.Owned, original.Owned) {
				t.Fatalf("combined edit: %+v", edited)
			}
			if got := graphMixedResult[graphstore.IssueRecord](t, call("show", "--format", "graph-json", current.ID)); !reflect.DeepEqual(got, current) {
				t.Fatal("fresh-process current read differs")
			}
			exact(original)
			exact(current)
			list([]string{current.ID}, "--assignee=alice")
			list([]string{scope + "beads/gate"}, "--no-assignee")
			list([]string{}, "--assignee=alice", "--no-assignee")
			before := graphMemoryReadSnapshot(t, work)
			noop := graphMixedResult[graphstore.IssueMutationResult](t, call("update", scope+"beads/work", "--priority", "0", "--assignee", "alice", "--if-revision", "="+current.Revision))
			if noop.Changed || !reflect.DeepEqual(noop.Issue, current) {
				t.Fatal("same-value edit was not a semantic no-op")
			}
			refuse("revision_conflict", "update", scope+"beads/work", "--priority", "0", "--assignee", "alice", "--if-revision", "="+original.Revision)
			refuse("permission_denied", "update", scope+"beads/work", "--priority", "1", "--readonly")
			refuse("capability_unavailable", "update", scope+"beads/work", "--assignee", "bob", "--claim")
			refuse("capability_unavailable", "update", "beads/context", "--priority", "1")
			refuse("capability_unavailable", "update", "links/context", "--assignee", "bob")
			// A freeze must block this newly admitted edit before its writer opens.
			writeFile(t, filepath.Join(work, "mayor", "town.json"), []byte("{}\n"))
			freeze := filepath.Join(work, "MIGRATION-FREEZE")
			writeFile(t, freeze, []byte("assignment-test\t2026-09-28T00:00:00Z\tfreeze\n"))
			refuse("permission_denied", "update", scope+"beads/work", "--assignee", "bob")
			if err := os.Remove(freeze); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(before, graphMemoryReadSnapshot(t, work)) {
				t.Fatal("no-op or refusal changed current graph state")
			}
			exact(original)
			exact(current)
			omitted := graphMixedResult[graphstore.IssueMutationResult](t, call("update", scope+"beads/work", "--description", "Revised description", "--if-revision", "="+current.Revision))
			if !omitted.Changed || omitted.Issue.Properties.Priority != 0 || omitted.Issue.Properties.Assignee != "alice" {
				t.Fatal("omitted fields were changed")
			}
			current = omitted.Issue
			cleared := graphMixedResult[graphstore.IssueMutationResult](t, call("update", scope+"beads/work", "--assignee=", "--if-revision", "="+current.Revision))
			if !cleared.Changed || cleared.Issue.Properties.Assignee != "" || cleared.Issue.Properties.Status != current.Properties.Status || cleared.Issue.Properties.Priority != 0 {
				t.Fatal("assignee clear changed unrelated state")
			}
			current = cleared.Issue
			list([]string{}, "--assignee=alice")
			list([]string{scope + "beads/gate", scope + "beads/work"}, "--no-assignee")
			if call("show", scope+"beads/context") != memory || call("show", scope+"beads/gate", "--format", "graph-json") != gate {
				t.Fatal("assignment changed unrelated Beads")
			}
			if got := graphMixedResult[graphstore.LinkRecord](t, call("show", dependency.Link.ID)); !reflect.DeepEqual(got, dependency.Link) {
				t.Fatal("assignment changed blocking Dependency")
			}
			if got := graphMixedResult[graphstore.LinkRecord](t, call("show", scope+"links/context")); !reflect.DeepEqual(got, graphMixedResult[graphstore.LinkMutationResult](t, informational).Link) {
				t.Fatal("assignment changed informational Link")
			}
			// Independent server processes compete against one observed graph revision.
			// The storage tests force transaction overlap; this boundary proves actual
			// installed callers and refuses generic/infrastructure errors as losers.
			if engine == "server" {
				results := graphDeleteRacePair(t, bd, work, home, [2][]string{
					{"update", scope + "beads/work", "--priority=1", "--assignee=first", "--if-revision", "=" + current.Revision, "--json"},
					{"update", scope + "beads/work", "--priority=4", "--assignee=second", "--if-revision", "=" + current.Revision, "--json"},
				})
				winner := -1
				for i, result := range results {
					if result.code == "" {
						if winner != -1 {
							t.Fatal("both guarded callers committed")
						}
						winner = i
					} else if result.code != "revision_conflict" {
						t.Fatalf("unexpected losing outcome: %s", result.code)
					}
				}
				if winner < 0 {
					t.Fatal("neither guarded caller committed")
				}
				won := graphMixedResult[graphstore.IssueMutationResult](t, results[winner].stdout)
				wantPriority, wantAssignee := 1, "first"
				if winner == 1 {
					wantPriority, wantAssignee = 4, "second"
				}
				if !won.Changed || won.Issue.Properties.Priority != wantPriority || won.Issue.Properties.Assignee != wantAssignee || won.Issue.Revision == current.Revision || !reflect.DeepEqual(won.Issue.Owned, current.Owned) {
					t.Fatalf("torn/wrong winner: %+v", won)
				}
				if got := graphMixedResult[graphstore.IssueRecord](t, call("show", "--format", "graph-json", scope+"beads/work")); !reflect.DeepEqual(got, won.Issue) {
					t.Fatal("fresh process did not read exact winner")
				}
				exact(current)
				current = won.Issue
				exact(current)
			}
			// Closed scalar edits must preserve the native closure, not reopen work.
			call("close", scope+"beads/gate", "--reason", "Done")
			closed := graphMixedResult[graphstore.IssueRecord](t, call("show", "--format", "graph-json", scope+"beads/gate"))
			afterClose := graphMixedResult[graphstore.IssueMutationResult](t, call("update", scope+"beads/gate", "--priority=4", "--assignee=reviewer", "--if-revision", "="+closed.Revision))
			if !afterClose.Changed || afterClose.Issue.Properties.Status != closed.Properties.Status || !reflect.DeepEqual(afterClose.Issue.Properties.ClosedAt, closed.Properties.ClosedAt) || afterClose.Issue.Properties.CloseReason != closed.Properties.CloseReason {
				t.Fatal("closed Issue edit changed closure")
			}
			exact(closed)
			exact(afterClose.Issue)
		})
	}
}

func TestGraphPreviewUpdateStatusTwoEngines(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			args := []string{"init", "--graph-mode", "link", "--prefix", "status", "--scope-url", "https://example.invalid/update-status/", "--skip-hooks", "--skip-agents", "--non-interactive"}
			if engine == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for shared-server status qualification")
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
			call("create", "Work", "--id", "beads/work", "--force")
			call("create", "Prerequisite", "--id", "beads/gate", "--force")
			call("dep", "add", "beads/work", "beads/gate")
			before := graphMixedResult[graphstore.IssueRecord](t, call("show", "beads/work", "--format", "graph-json"))
			refuse("constraint_violation", "update", "beads/work", "--status=closed", "--if-revision", "="+before.Revision)
			refuse("invalid_properties", "update", "beads/work", "--status=bogus")
			refuse("invalid_properties", "update", "beads/work", "--status=open", "--session=review")
			if got := graphMixedResult[graphstore.IssueRecord](t, call("show", "beads/work", "--format", "graph-json")); !reflect.DeepEqual(got, before) {
				t.Fatal("refused status edits changed the Issue")
			}
			closed := graphMixedResult[graphstore.IssueMutationResult](t, call("update", "beads/work", "--status=closed", "--session=review", "--force", "--if-revision", "="+before.Revision))
			if !closed.Changed || closed.Issue.Properties.Status != "closed" || closed.Issue.Properties.ClosedAt == nil || closed.Issue.Properties.ClosedBySession != "review" {
				t.Fatalf("forced status did not close with attribution: %+v", closed)
			}
			refuse("revision_conflict", "update", "beads/work", "--status=open", "--if-revision", "="+before.Revision)
			reopened := graphMixedResult[graphstore.IssueMutationResult](t, call("update", "beads/work", "--status=open", "--if-revision", "="+closed.Issue.Revision))
			if !reopened.Changed || reopened.Issue.Properties.Status != "open" || reopened.Issue.Properties.ClosedAt != nil {
				t.Fatalf("status update did not reopen: %+v", reopened)
			}
			noop := graphMixedResult[graphstore.IssueMutationResult](t, call("update", "beads/work", "--status=open", "--if-revision", "="+reopened.Issue.Revision))
			if noop.Changed || noop.Issue.Revision != reopened.Issue.Revision {
				t.Fatalf("same-status update minted a revision: %+v", noop)
			}
			refuse("constraint_violation", "update", "beads/work", "--status=open", "--if-status=closed", "--if-revision", "="+reopened.Issue.Revision)
			refuse("invalid_properties", "update", "beads/work", "--status=open", "--if-status=unconfigured")
			if current := graphMixedResult[graphstore.IssueRecord](t, call("show", "beads/work", "--format", "graph-json")); !reflect.DeepEqual(current, reopened.Issue) {
				t.Fatal("refused status guards changed the Issue")
			}
			guarded := graphMixedResult[graphstore.IssueMutationResult](t, call("update", "beads/work", "--status=in_progress", "--if-status=open", "--if-revision", "="+reopened.Issue.Revision))
			if !guarded.Changed || guarded.Issue.Properties.Status != "in_progress" {
				t.Fatalf("matching status guard did not admit edit: %+v", guarded)
			}
		})
	}
}

func TestGraphPreviewUpdateForceAssigneeTwoEngines(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			const scope = "https://example.invalid/update-force-assignee/"
			args := []string{"init", "--graph-mode", "link", "--prefix", "transfer", "--scope-url", scope, "--skip-hooks", "--skip-agents", "--non-interactive"}
			if engine == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for shared-server update qualification")
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
			call("create", "Claimed work", "--id", "beads/work", "--notes", "Original", "--force")
			claimed := graphMixedResult[graphstore.IssueMutationResult](t, call("update", "beads/work", "--claim", "--actor", "holder"))
			refuse("constraint_violation", "update", "beads/work", "--assignee", "successor", "--actor", "successor")
			refuse("constraint_violation", "update", "beads/work", "--assignee", "successor", "--force=false", "--actor", "successor")
			unforced := graphMixedResult[graphstore.IssueMutationResult](t, call("update", "beads/work", "--title", "Still claimed", "--force=false", "--if-revision", "="+claimed.Issue.Revision, "--actor", "holder"))
			if !unforced.Changed || unforced.Issue.Properties.Title != "Still claimed" || unforced.Issue.Properties.Assignee != "holder" {
				t.Fatalf("--force=false changed update semantics: %+v", unforced)
			}
			refuse("revision_conflict", "update", "beads/work", "--assignee", "successor", "--force", "--if-revision", "="+claimed.Issue.Revision, "--actor", "successor")
			if current := graphMixedResult[graphstore.IssueRecord](t, call("show", "beads/work", "--format", "graph-json")); !reflect.DeepEqual(current, unforced.Issue) {
				t.Fatal("stale forced transfer changed the Issue")
			}
			transferred := graphMixedResult[graphstore.IssueMutationResult](t, call("update", "beads/work", "--assignee", "successor", "--notes", "Revised", "--force", "--if-revision", "="+unforced.Issue.Revision, "--actor", "successor"))
			if !transferred.Changed || transferred.Issue.Properties.Assignee != "successor" || transferred.Issue.Properties.Notes != "Revised" || transferred.Issue.Properties.Status != unforced.Issue.Properties.Status || transferred.Issue.Revision == unforced.Issue.Revision {
				t.Fatalf("forced transfer and notes were not one guarded edit: %+v", transferred)
			}
			appended := graphMixedResult[graphstore.IssueMutationResult](t, call("update", "beads/work", "--append-notes", "Progress", "--force=false", "--if-revision", "="+transferred.Issue.Revision, "--actor", "successor"))
			if !appended.Changed || appended.Issue.Properties.Assignee != "successor" || appended.Issue.Properties.Notes == "Revised" || !strings.HasSuffix(appended.Issue.Properties.Notes, "Progress") || appended.Issue.Revision == transferred.Issue.Revision {
				t.Fatalf("false force changed append semantics: %+v", appended)
			}
			refuse("invalid_properties", "update", "beads/work", "--append-notes", "Must refuse", "--force")
			if current := graphMixedResult[graphstore.IssueRecord](t, call("show", "beads/work", "--format", "graph-json")); !reflect.DeepEqual(current, appended.Issue) {
				t.Fatal("refused forced append changed the Issue")
			}
			refuse("invalid_properties", "update", "beads/work", "--title", "Unsupported force", "--force")
			properties := graphMixedResult[graphstore.IssueMutationResult](t, call("update", "beads/work", `--properties={"assignee":"third","notes":"Third"}`, "--force", "--actor", "third"))
			if !properties.Changed || properties.Issue.Properties.Assignee != "third" || properties.Issue.Properties.Notes != "Third" {
				t.Fatalf("property force intent lost: %+v", properties)
			}
			human := graphPolicyCLI(t, bd, work, home, nil, "", "update", "beads/work", "--assignee", "fourth", "--force", "--actor", "fourth")
			if !strings.Contains(human, "Updated "+scope+"beads/work") {
				t.Fatalf("human update output changed: %q", human)
			}
			current := graphMixedResult[graphstore.IssueRecord](t, call("show", "beads/work", "--format", "graph-json"))
			refuse("constraint_violation", "update", "beads/work", "--assignee", "fourth", "--if-assignee", "third", "--if-revision", "="+current.Revision)
			if after := graphMixedResult[graphstore.IssueRecord](t, call("show", "beads/work", "--format", "graph-json")); !reflect.DeepEqual(after, current) {
				t.Fatal("stale assignee guard changed the Issue")
			}
			guarded := graphMixedResult[graphstore.IssueMutationResult](t, call("update", "beads/work", "--assignee", "fifth", "--if-assignee", "fourth", "--force", "--if-revision", "="+current.Revision))
			if !guarded.Changed || guarded.Issue.Properties.Assignee != "fifth" {
				t.Fatalf("matching assignee guard did not authorize transfer: %+v", guarded)
			}
			call("create", "Unassigned", "--id", "beads/unassigned", "--force")
			unassigned := graphMixedResult[graphstore.IssueMutationResult](t, call("update", "beads/unassigned", "--assignee", "sixth", "--if-assignee="))
			if !unassigned.Changed || unassigned.Issue.Properties.Assignee != "sixth" {
				t.Fatalf("empty assignee guard did not match unassigned Issue: %+v", unassigned)
			}
		})
	}
}
