//go:build cgo

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/memoryapi"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func TestGraphPreviewMemoryPublicIdentityWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			const scope = "https://example.invalid/memory-public-identity/"
			work, home := graphListWorkspace(t, bd, engine, scope)
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			for _, key := range []string{"TWO", "hello", "%E9%9B%AA/@literal", "beads/literal", "links/memory"} {
				body := "  Body  summary\nExact second line\n"
				created := graphMixedResult[graphstore.Record](t, call("remember", body, "--key", key))
				if created.ID != scope+key || created.Properties.Title != memoryapi.TitleFromBody(body) || created.Properties.Body != body {
					t.Fatal("creation changed title/body")
				}
				out, stderr, exit := graphGenerationRun(t, bd, work, home, "recall", key)
				if exit != 0 || out != body+"\n" {
					t.Fatalf("recall %q: %q %s exit%d", key, out, stderr, exit)
				}
				edited := graphMixedResult[graphstore.MemoryMutationResult](t, call("remember", "Different body", "--id", key))
				if edited.Memory.ID != created.ID || edited.Memory.Properties.Title != created.Properties.Title {
					t.Fatal("overwrite changed identity/title")
				}
				shown := graphMixedResult[graphstore.Record](t, call("show", key, "--format", "graph-json"))
				if shown.ID != created.ID || shown.Properties.Body != "Different body" {
					t.Fatal("returned key did not select same Memory")
				}
				summary := graphMixedResult[graphMemoryDiscoveryResult](t, graphPolicyCLI(t, bd, work, home, nil, "", "memories", key, "--format", "records-json"))
				found := false
				for _, item := range summary.Items {
					if item.ID == key {
						found = true
						if item.Title != created.Properties.Title || item.Excerpt == nil {
							t.Fatal("summary lost title/body")
						}
					}
				}
				if !found {
					t.Fatalf("summary did not return exact ID %q: %+v", key, summary)
				}
				graphPolicyCLI(t, bd, work, home, nil, "invalid_selector", "remember", "Refused", "--id", key, "--key", key, "--json")
				after := graphMixedResult[graphstore.Record](t, call("show", key, "--format", "graph-json"))
				if after.Revision != shown.Revision {
					t.Fatal("conflicting flags changed Memory")
				}
				removed := graphMixedResult[graphstore.MemoryDeleteResult](t, call("forget", key, "--if-revision", shown.Revision))
				if removed.Terminal == nil {
					t.Fatal("forget omitted terminal revision")
				}
				graphPolicyCLI(t, bd, work, home, nil, "identity_reserved", "remember", "No auto recovery", "--key", key, "--json")
				restored := graphMixedResult[graphstore.MemoryMutationResult](t, call("update", key, "--from-revision", created.Revision, "--if-revision", removed.Terminal.Revision))
				if restored.Memory.ID != created.ID || restored.Memory.Properties.Body != body {
					t.Fatal("restore changed identity or source content")
				}
			}
			const sourceKey, targetKey = "source/%E6%BA%90/@literal", "target-literal"
			source := graphMixedResult[graphstore.Record](t, call("remember", "Source body", "--id", sourceKey))
			target := graphMixedResult[graphstore.Record](t, call("remember", "Target body", "--id", targetKey))
			linked := graphMixedResult[graphstore.LinkMutationResult](t, call("link", "create", "--from", sourceKey, "--to", targetKey, "--type", "types/references", "--id", "literal-endpoints"))
			if linked.Link.Source != source.ID || linked.Link.Target != target.ID {
				t.Fatal("literal public endpoints selected different allocations")
			}
			const content = "The exact legacy derived identity must overwrite"
			first := graphMixedResult[graphstore.Record](t, call("remember", content))
			second := graphMixedResult[graphstore.MemoryMutationResult](t, call("remember", content))
			if second.Memory.ID != first.ID || second.Changed {
				t.Fatal("derived remember allocated a second Memory")
			}
			summary := graphMixedResult[graphMemoryDiscoveryResult](t, graphPolicyCLI(t, bd, work, home, nil, "", "memories", memoryapi.DeriveKey(content), "--format", "records-json"))
			if len(summary.Items) != 1 || summary.Items[0].ID != memoryapi.DeriveKey(content) {
				t.Fatal("legacy derivation changed")
			}
			one := graphMixedResult[graphstore.Record](t, call("remember", "One body", "--id", "one"))
			before := graphMixedResult[graphMemoryDiscoveryResult](t, graphPolicyCLI(t, bd, work, home, nil, "", "memories", "--all", "--format", "records-json"))
			for _, invalid := range []string{one.ID, "..", "a//b", "a?b", "雪"} {
				graphPolicyCLI(t, bd, work, home, nil, "invalid_properties", "remember", "Rejected compatibility key", "--id", invalid, "--json")
			}
			selected := graphMixedResult[graphstore.Record](t, call("show", one.ID, "--format", "graph-json"))
			if selected.ID != one.ID {
				t.Fatal("returned canonical URL selected different allocation")
			}
			after := graphMixedResult[graphMemoryDiscoveryResult](t, graphPolicyCLI(t, bd, work, home, nil, "", "memories", "--all", "--format", "records-json"))
			if len(before.Items) != len(after.Items) {
				t.Fatal("rejected identities changed workspace")
			}
			if strings.TrimSpace(first.Properties.Body) != content {
				t.Fatal("body changed")
			}
		})
	}
}

func TestGraphPreviewCommonIdentityGraphAndNativeLinkWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			const scope = "https://example.invalid/common-cli/"
			work, home := graphListWorkspace(t, bd, engine, scope)
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			source := graphMixedResult[graphstore.Record](t, call("remember", "Source", "--id", "source"))
			target := graphMixedResult[graphstore.Record](t, call("remember", "Target", "--id", "links/target"))
			link := graphMixedResult[graphstore.LinkMutationResult](t, call("link", "create", "--from", source.ID, "--to", target.ID, "--type", "types/references", "--id", "beads/edge"))
			if link.Link.ID != scope+"beads/edge" || link.Link.Source != source.ID || link.Link.Target != target.ID {
				t.Fatal("informational identity changed", link)
			}
			selected := call("graph", "source", "--link-type", "types/references", "--direction", "out")
			for _, id := range []string{source.ID, target.ID, link.Link.ID} {
				if !strings.Contains(selected, id) {
					t.Fatalf("graph omitted exact identity %q: %s", id, selected)
				}
			}
			call("link", "versions", link.Link.ID)
			call("link", "update", "beads/edge", "--properties", `{"note":"Updated"}`)
			call("link", "delete", "beads/edge", "--force")
			call("link", "versions", "beads/edge")
			one := graphMixedResult[graphstore.IssueRecord](t, call("create", "One"))
			two := graphMixedResult[graphstore.IssueRecord](t, call("create", "Two"))
			if one.ID != scope+one.Properties.ID || two.ID != scope+two.Properties.ID {
				t.Fatal("Issue ID acquired a kind prefix")
			}
			dependency := graphMixedResult[graphstore.DependencyResult](t, call("link", "create", "--from", one.ID, "--to", two.ID, "--type", "types/dependency", "--id", "dependency-rootless"))
			if dependency.Link.ID != scope+"dependency-rootless" || dependency.Link.Source != one.ID || dependency.Link.Target != two.ID {
				t.Fatal("native Link identity changed", dependency)
			}
			nativeGraph := call("graph", one.ID)
			for _, id := range []string{one.ID, two.ID, dependency.Link.ID} {
				if !strings.Contains(nativeGraph, id) {
					t.Fatalf("native graph omitted %q: %s", id, nativeGraph)
				}
			}
			tree := graphPolicyCLI(t, bd, work, home, nil, "", "dep", "tree", one.ID)
			if !strings.HasPrefix(tree, "Blocking dependency tree for "+one.Properties.ID+" ") || !strings.Contains(tree, two.Properties.ID) || strings.Contains(tree, "beads/"+one.Properties.ID) {
				t.Fatalf("dependency tree lost public native IDs: %s", tree)
			}
			call("dep", "remove", one.ID, two.ID)
			isolated := graphPolicyCLI(t, bd, work, home, nil, "", "dep", "tree", one.ID)
			if strings.TrimSpace(isolated) != one.Properties.ID+" has no dependencies" {
				t.Fatalf("isolated tree leaked private identity: %s", isolated)
			}
		})
	}
}

func TestGraphPreviewCommonIdentityCurrentAndRelated(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			const scope = "https://example.invalid/current-public/"
			work, home := graphListWorkspace(t, bd, engine, scope)
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			one := graphMixedResult[graphstore.IssueRecord](t, call("create", "Actual Issue"))
			two := graphMixedResult[graphstore.IssueRecord](t, call("create", "Other Issue"))
			call("remember", "Different allocation", "--id", "beads/"+one.Properties.ID)
			call("show", one.ID, "--format", "graph-json")
			current := graphMixedResult[graphstore.IssueRecord](t, call("show", "--current", "--actor", "none", "--format", "graph-json"))
			if current.ID != one.ID {
				t.Fatalf("last-touched selected private spelling: %+v", current)
			}
			var many []map[string]any
			if err := json.Unmarshal([]byte(call("show", one.ID, two.ID)), &many); err != nil {
				t.Fatal(err)
			}
			if len(many) != 2 || many[0]["id"] != one.Properties.ID || many[1]["id"] != two.Properties.ID {
				t.Fatalf("multi show: %+v", many)
			}
			for _, flag := range []string{"--refs", "--children"} {
				var related map[string]json.RawMessage
				if err := json.Unmarshal([]byte(call("show", one.ID, flag)), &related); err != nil {
					t.Fatal(err)
				}
				if _, ok := related[one.Properties.ID]; !ok {
					t.Fatalf("related map used private path: %v", related)
				}
			}
			versions := graphMixedResult[map[string]any](t, call("versions", one.ID))
			if versions["resource"] != one.Properties.ID {
				t.Fatalf("versions leaked locator: %+v", versions)
			}
			raw := graphPolicyCLI(t, bd, work, home, nil, "", "versions", one.ID, "--output-format", "jsonl")
			if !strings.Contains(raw, `"resource":"`+one.Properties.ID+`"`) || strings.Contains(raw, `"resource":"beads/`+one.Properties.ID+`"`) {
				t.Fatalf("JSONL versions identity: %s", raw)
			}
		})
	}
}

func TestGraphPreviewCommonIdentityIssueDelete(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			const scope = "https://example.invalid/delete-public/"
			work, home := graphListWorkspace(t, bd, engine, scope)
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			issue := graphMixedResult[graphstore.IssueRecord](t, call("create", "Native deletion subject"))
			memory := graphMixedResult[graphstore.Record](t, call("remember", "Different allocation", "--id", "beads/"+issue.Properties.ID))
			issueBefore := call("show", issue.ID, "--format", "graph-json")
			memoryBefore := call("show", memory.ID)
			for _, selector := range []string{issue.Properties.ID, issue.ID} {
				preview := graphMixedResult[graphstore.IssueDeleteResult](t, call("delete", selector))
				if !preview.Preview || preview.Deleted || preview.Issue.ID != issue.ID {
					t.Fatalf("preview selected wrong allocation: %+v", preview)
				}
			}
			graphPolicyCLI(t, bd, work, home, nil, "revision_conflict", "delete", issue.ID, "--force", "--if-revision", "stale", "--json")
			if call("show", issue.ID, "--format", "graph-json") != issueBefore || call("show", memory.ID) != memoryBefore {
				t.Fatal("preview or stale delete changed Issue or competing Memory")
			}
			deleted := graphMixedResult[graphstore.IssueDeleteResult](t, call("delete", issue.Properties.ID, "--force", "--if-revision", issue.Revision))
			if !deleted.Deleted || deleted.Issue.ID != issue.ID || deleted.Terminal == nil || deleted.Terminal.ID != issue.ID {
				t.Fatalf("delete lost public identity: %+v", deleted)
			}
			graphPolicyCLI(t, bd, work, home, nil, "gone", "show", issue.ID, "--json")
			if call("show", memory.ID) != memoryBefore {
				t.Fatal("Issue deletion selected the Memory whose public ID matches its private locator")
			}
		})
	}
}

func TestGraphPreviewCommonIdentityLiteralIssueAdapters(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			const scope = "https://example.invalid/literal-issue/"
			work, home := graphListWorkspace(t, bd, engine, scope)
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			one := graphMixedResult[graphstore.IssueRecord](t, call("create", "Literal Issue", "--id", "work@literal", "--force"))
			two := graphMixedResult[graphstore.IssueRecord](t, call("create", "Other Issue"))
			for _, args := range [][]string{
				{"show", one.Properties.ID, two.Properties.ID},
				{"show", "--id", one.Properties.ID, "--id", two.Properties.ID},
				{"show", one.Properties.ID, "--id", two.Properties.ID},
			} {
				var rows []map[string]any
				if err := json.Unmarshal([]byte(call(args...)), &rows); err != nil {
					t.Fatal(err)
				}
				if len(rows) != 2 || rows[0]["id"] != one.Properties.ID || rows[1]["id"] != two.Properties.ID {
					t.Fatalf("literal multi selector changed identity: %+v", rows)
				}
			}
			for _, flag := range []string{"--refs", "--children"} {
				for _, selector := range [][]string{{one.Properties.ID}, {"--id", one.Properties.ID}} {
					var rows map[string]json.RawMessage
					if err := json.Unmarshal([]byte(call(append([]string{"show", flag}, selector...)...)), &rows); err != nil {
						t.Fatal(err)
					}
					if _, ok := rows[one.Properties.ID]; !ok {
						t.Fatalf("literal related selector changed identity: %+v", rows)
					}
				}
			}
			graphPolicyCLI(t, bd, work, home, nil, "not_found", "show", "--refs", "--id", "work@no-such-key", "--json")
			empty := graphPolicyCLI(t, bd, work, home, nil, "", "comments", one.Properties.ID)
			if strings.TrimSpace(empty) != "No comments on "+one.Properties.ID {
				t.Fatalf("empty comments leaked private locator: %s", empty)
			}
			call("comments", "add", one.Properties.ID, "A native comment")
			comments := graphPolicyCLI(t, bd, work, home, nil, "", "comments", one.Properties.ID)
			if !strings.Contains(comments, "Comments on "+one.Properties.ID+":") || strings.Contains(comments, "beads/"+one.Properties.ID) {
				t.Fatalf("comments leaked private locator: %s", comments)
			}
			editor := filepath.Join(t.TempDir(), "editor")
			if err := os.WriteFile(editor, []byte("#!/bin/sh\nprintf 'Edited native title\\n' > \"$1\"\n"), 0700); err != nil {
				t.Fatal(err)
			}
			edited := graphPolicyCLI(t, bd, work, home, []string{"EDITOR=" + editor}, "", "edit", one.Properties.ID, "--title")
			if !strings.Contains(edited, one.Properties.ID) || strings.Contains(edited, "beads/"+one.Properties.ID) {
				t.Fatalf("editor feedback leaked private locator: %s", edited)
			}
			current := graphMixedResult[graphstore.IssueRecord](t, call("show", one.ID, "--format", "graph-json"))
			if current.ID != one.ID || current.Properties.Title != "Edited native title" {
				t.Fatalf("editor changed identity or lost title: %+v", current)
			}
		})
	}
}
