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
			work, home := graphListWorkspace(t, bd, engine, "https://example.invalid/memory-public-identity/")
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			for _, key := range []string{"TWO", " hello ", "雪/%@?#", "beads/literal", "https://foreign.invalid/a@b"} {
				body := "  Body  summary\nExact second line\n"
				created := graphMixedResult[graphstore.Record](t, call("remember", body, "--key", key))
				if created.Properties.Title != memoryapi.TitleFromBody(body) || created.Properties.Body != body {
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
			literal := graphMixedResult[graphstore.Record](t, call("remember", "URL key body", "--id", one.ID))
			selected := graphMixedResult[graphstore.Record](t, call("show", "--id", one.ID, "--format", "graph-json"))
			if selected.ID != literal.ID {
				t.Fatal("literal --id selected locator allocation")
			}
			graphPolicyCLI(t, bd, work, home, nil, "invalid_properties", "show", one.ID, "--format", "graph-json", "--json")
			before := graphMixedResult[graphMemoryDiscoveryResult](t, graphPolicyCLI(t, bd, work, home, nil, "", "memories", "--all", "--format", "records-json"))
			graphPolicyCLI(t, bd, work, home, nil, "invalid_properties", "forget", one.ID, "--if-revision", one.Revision, "--json")
			after := graphMixedResult[graphMemoryDiscoveryResult](t, graphPolicyCLI(t, bd, work, home, nil, "", "memories", "--all", "--format", "records-json"))
			if len(before.Items) != len(after.Items) {
				t.Fatal("ambiguous deletion changed workspace")
			}
			if strings.TrimSpace(first.Properties.Body) != content {
				t.Fatal("body changed")
			}
		})
	}
}
