//go:build cgo

package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// Every operation is a fresh installed CLI process after normal initialization.
// The required graph lane must supply the ordinary server and reject its skip.
func TestGraphPreviewMemoryDeleteWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			const scope = "https://example.invalid/delete/"
			initArgs := []string{"init", "--graph-mode", "link", "--scope-url", scope, "--skip-hooks", "--skip-agents", "--non-interactive"}
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
			refuse := func(code string, args ...string) {
				t.Helper()
				graphPolicyCLI(t, bd, work, home, nil, code, append(args, "--json")...)
			}
			call(initArgs...)
			// Empty content is present, a lesson from sjarmak's gastownhall/beads#5964
			// ordinary KV-memory fix. This separate graph path uses normal CLI creation.
			original := graphMixedResult[graphstore.Record](t, call("remember", "", "--id", "beads/empty", "--title", "Empty body"))
			if original.Properties.Body != "" || original.Revision == "" {
				t.Fatal("empty-body Memory was not present")
			}
			preview := graphMixedResult[graphstore.MemoryDeleteResult](t, call("delete", original.ID, "--readonly"))
			if !preview.Preview || preview.Deleted || !reflect.DeepEqual(preview.Memory, original) {
				t.Fatal("default read-only preview differs from live Memory")
			}
			if got := graphMixedResult[graphstore.Record](t, call("show", "beads/empty")); !reflect.DeepEqual(got, original) {
				t.Fatal("preview changed current state")
			}
			updated := graphMixedResult[graphstore.MemoryMutationResult](t, call("remember", "--update", "beads/empty", "--title", "Final empty body", "--if-revision", original.Revision)).Memory
			for _, args := range [][]string{
				{"delete", "beads/empty", "--if-revision", original.Revision},
				{"delete", "beads/empty", "--force", "--if-revision", original.Revision},
				{"forget", "beads/empty", "--if-revision", original.Revision},
			} {
				refuse("revision_conflict", args...)
			}
			for _, args := range [][]string{
				{"delete", "beads/empty", "--if-revision="},
				{"delete", "beads/empty", "beads/other"}, {"delete", "links/context"},
				{"delete", "https://foreign.invalid/beads/empty"},
			} {
				refuse("invalid_selector", args...)
			}
			for _, args := range [][]string{
				{"delete", "beads/empty", "--force", "--readonly"},
				{"forget", "beads/empty", "--readonly"},
			} {
				refuse("permission_denied", args...)
			}
			for _, flag := range []string{"--cascade", "--cascade=false", "--from-file=/unread/deletions", "--dry-run"} {
				refuse("capability_unavailable", "delete", "beads/empty", flag)
			}
			refuse("not_found", "delete", "beads/missing")
			// The graph adapter must honor bound read-only policy and migration
			// freezes while leaving its default preview available as a read.
			graphPolicyCLI(t, bd, work, home, []string{"BD_READONLY=true"}, "permission_denied", "delete", "beads/empty", "--force", "--json")
			writeFile(t, filepath.Join(work, "mayor", "town.json"), []byte("{}\n"))
			freezePath := filepath.Join(work, "MIGRATION-FREEZE")
			writeFile(t, freezePath, []byte("memory-delete-test\t2026-09-28T00:00:00Z\tdeletion policy control\n"))
			call("delete", "beads/empty")
			refuse("permission_denied", "delete", "beads/empty", "--force")
			refuse("permission_denied", "forget", "beads/empty")
			if err := os.Remove(freezePath); err != nil {
				t.Fatal(err)
			}
			if got := graphMixedResult[graphstore.Record](t, call("show", updated.ID)); !reflect.DeepEqual(got, updated) {
				t.Fatal("refused deletion changed final live state")
			}
			deleted := graphMixedResult[graphstore.MemoryDeleteResult](t, call("delete", "beads/empty", "--force", "--if-revision", updated.Revision))
			if deleted.Preview || !deleted.Deleted || !reflect.DeepEqual(deleted.Memory, updated) {
				t.Fatal("delete did not disclose exact final live state")
			}
			refuse("gone", "show", updated.ID)
			refuse("gone", "delete", "beads/empty")
			refuse("gone", "forget", "beads/empty")
			refuse("identity_reserved", "remember", "replacement", "--id", "beads/empty", "--title", "Cannot reuse")
			unguarded := graphMixedResult[graphstore.Record](t, call("remember", "Disposable", "--id", "beads/unguarded", "--title", "Unguarded"))
			applied := graphMixedResult[graphstore.MemoryDeleteResult](t, call("delete", "beads/unguarded", "--force"))
			if !applied.Deleted || applied.Preview || !reflect.DeepEqual(applied.Memory, unguarded) {
				t.Fatal("unguarded --force did not delete the current Memory")
			}
			refuse("gone", "show", "beads/unguarded")

			plan := graphMixedResult[graphstore.Record](t, call("remember", "Plan", "--id", "beads/plan", "--title", "Plan"))
			other := call("remember", "Context", "--id", "beads/other", "--title", "Context")
			issue := call("create", "Keep Issue", "--id", "beads/work")
			issuePreview := graphMixedResult[graphstore.IssueDeleteResult](t, call("delete", "beads/work"))
			if !issuePreview.Preview || issuePreview.Deleted || issuePreview.Issue.ID != scope+"beads/work" {
				t.Fatal("Issue deletion preview did not preserve live Issue")
			}
			refuse("capability_unavailable", "forget", "beads/work")
			related := scope + "types/superseded-by"
			call("link", "beads/plan", "beads/other", "--id", "links/outgoing", "--type", related, "--if-source-revision", plan.Revision)
			call("link", "beads/work", "beads/plan", "--id", "links/incoming", "--type", "types/references")
			for _, path := range []string{"links/outgoing", "links/incoming"} {
				before := call("show", "beads/plan")
				linkBefore := call("show", path)
				current := graphMixedResult[graphstore.Record](t, before)
				refuse("constraint_violation", "delete", "beads/plan")
				refuse("constraint_violation", "delete", "beads/plan", "--force", "--if-revision", current.Revision)
				refuse("constraint_violation", "forget", "beads/plan")
				if call("show", "beads/plan") != before || call("show", path) != linkBefore || call("show", "beads/other") != other || call("show", "beads/work", "--format", "graph-json") != issue {
					t.Fatal("incident-Link refusal changed Memory, Link or another endpoint")
				}
				link := graphMixedResult[graphstore.LinkRecord](t, linkBefore)
				args := []string{"link", "delete", "--force", path, "--if-revision", link.Revision}
				if path == "links/outgoing" {
					args = append(args, "--if-source-revision", current.Revision)
				}
				call(args...)
			}
			final := graphMixedResult[graphstore.Record](t, call("show", "beads/plan"))
			forgotten := graphMixedResult[graphstore.MemoryDeleteResult](t, call("forget", final.ID))
			if !forgotten.Deleted || forgotten.Preview || !reflect.DeepEqual(forgotten.Memory, final) || len(final.Owned) != 0 {
				t.Fatal("forget after explicit unlink lost final predecessor or invented a live version")
			}
			refuse("gone", "show", "beads/plan")
			refuse("identity_reserved", "remember", "replacement", "--id", "beads/plan", "--title", "Cannot reuse")
			if call("show", "beads/other") != other || call("show", "beads/work", "--format", "graph-json") != issue {
				t.Fatal("Memory deletion changed unrelated surviving records")
			}
		})
	}
}
