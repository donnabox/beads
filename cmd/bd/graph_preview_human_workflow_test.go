//go:build cgo

package main

import (
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func TestGraphHumanInstalledWorkflow(t *testing.T) {
	isolateBeadsDirForTest(t)
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			const scope = "https://example.invalid/human-output/"
			work, home := graphListWorkspace(t, bd, engine, scope)
			call := func(args ...string) string { t.Helper(); return graphPolicyCLI(t, bd, work, home, nil, "", args...) }
			create := call("create", "First", "--type", "types/memory", "--id", "beads/first", "--body", "Old body")
			if strings.TrimSpace(create) != `beads/first [types/memory] "First"` {
				t.Fatal(create)
			}
			one := graphMixedResult[graphstore.Record](t, call("show", "beads/first", "--json"))
			call("create", "Second", "--type", "types/memory", "--id", "second")
			shown := call("show", "beads/first")
			if !strings.Contains(shown, "Revision: @1 (current)") || !strings.Contains(shown, "Old body") || strings.Contains(shown, one.Revision) {
				t.Fatal(shown)
			}
			call("update", "beads/first", "--properties", `{"body":"New body"}`)
			past := call("show", "beads/first@1")
			if !strings.Contains(past, "Revision: @1 (retained snapshot)") || !strings.Contains(past, "Old body") || strings.Contains(past, "New body") {
				t.Fatal(past)
			}
			if got := call("show", "beads/first@1", "--long"); !strings.Contains(got, one.Revision) {
				t.Fatal(got)
			}
			createdLink := call("link", "create", "beads/first", "second@1", "--type", "types/references", "--id", "links/context")
			second := graphMixedResult[graphstore.Record](t, call("show", "second", "--json"))
			row := "links/context [types/references] beads/first → second@=" + second.Revision
			if strings.TrimSpace(createdLink) != row {
				t.Fatal(createdLink)
			}
			if got := call("link", "list"); !strings.Contains(got, row) {
				t.Fatal(got)
			}
			link := graphMixedResult[graphstore.LinkRecord](t, call("link", "show", "links/context", "--json"))
			call("link", "update", "links/context", "--properties", `{"body":"New Link body"}`)
			for _, selector := range []string{"links/context", "links/context@1"} {
				got := call("link", "show", selector)
				if !strings.HasPrefix(got, row+"\nRevision:") || !strings.Contains(got, "From: beads/first") || !strings.Contains(got, "To: second@="+second.Revision) {
					t.Fatal(got)
				}
				if selector == "links/context@1" && strings.Contains(got, "New Link body") {
					t.Fatal(got)
				}
			}
			if got := graphMixedResult[graphstore.LinkRecord](t, call("link", "show", "links/context@1", "--json")); got.Revision != link.Revision || got.TargetRevision != second.Revision {
				t.Fatal(got)
			}
			issue := call("create", "Native", "--id-root", "work")
			if !strings.Contains(issue, "[task] open P2") {
				t.Fatal(issue)
			}
			if got := call("list", "--type", "any"); !strings.Contains(got, strings.TrimSpace(create)) || !strings.Contains(got, strings.TrimSpace(issue)) {
				t.Fatal(got)
			}
			// All five authoring facades use the generic Link row. Constraints admit
			// the three constraint relationships without bypassing their real writer.
			for i, op := range memoryLinkOperations {
				src, dst := "beads/first", "second"
				if i >= 2 {
					src, dst = "c-source-"+op.name, "c-target-"+op.name
					call("create", src, "--type", "types/constraint", "--id", src)
					call("create", dst, "--type", "types/constraint", "--id", dst)
				}
				id := "edge-" + op.name
				got := call("memory", op.name, src, dst, "--id", id)
				want := id + " [types/" + op.family + "] " + src + " → " + dst
				if strings.TrimSpace(got) != want {
					t.Fatalf("%s got %s want %s", op.name, got, want)
				}
				if listed := call("link", "list"); !strings.Contains(listed, want) {
					t.Fatal(listed)
				}
			}
			recoverable := graphMixedResult[graphstore.Record](t, call("create", "Recovery", "--type", "types/memory", "--id", "recovery", "--json"))
			call("delete", "recovery", "--force")
			call("update", "recovery", "--from-revision", recoverable.Revision)
			timeline := call("versions", "recovery")
			if !strings.Contains(timeline, "  SEQ  WHEN") || !strings.Contains(timeline, "  2") || !strings.Contains(timeline, "(deleted; not a live snapshot)") || !strings.Contains(timeline, "  3") {
				t.Fatal(timeline)
			}
		})
	}
}
