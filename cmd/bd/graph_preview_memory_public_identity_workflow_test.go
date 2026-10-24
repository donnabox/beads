//go:build cgo

package main

import (
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
			linked := graphMixedResult[graphstore.LinkMutationResult](t, call("link", "create", "--source", sourceKey, "--target", targetKey, "--type", "types/references", "--id", "literal-endpoints"))
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
			link := graphMixedResult[graphstore.LinkMutationResult](t, call("link", "create", "--source", source.ID, "--target", target.ID, "--type", "types/references", "--id", "beads/edge"))
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
			dependency := graphMixedResult[graphstore.DependencyResult](t, call("link", "create", "--source", one.ID, "--target", two.ID, "--type", "types/dependency", "--id", "dependency-rootless"))
			if dependency.Link.ID != scope+"dependency-rootless" || dependency.Link.Source != one.ID || dependency.Link.Target != two.ID {
				t.Fatal("native Link identity changed", dependency)
			}
			nativeGraph := call("graph", one.ID)
			for _, id := range []string{one.ID, two.ID, dependency.Link.ID} {
				if !strings.Contains(nativeGraph, id) {
					t.Fatalf("native graph omitted %q: %s", id, nativeGraph)
				}
			}
			call("dep", "remove", one.ID, two.ID)
		})
	}
}
