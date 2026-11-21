//go:build cgo

package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func TestGraphPreviewEditWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			initArgs := []string{"init", "--graph-mode", "link", "--scope-url", "https://example.invalid/edit/", "--skip-hooks", "--skip-agents", "--non-interactive"}
			if engine == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for ordinary shared-server edit qualification")
				}
				initArgs = append(initArgs, "--server", "--external", "--server-host", "127.0.0.1", "--server-port", port, "--server-user", "root")
			}
			call := func(code string, env []string, args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, env, code, append(args, "--json")...)
			}
			editor := func(body string) []string {
				t.Helper()
				path := filepath.Join(t.TempDir(), "editor")
				if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body+"\n"), 0700); err != nil {
					t.Fatal(err)
				}
				return []string{"EDITOR=" + path}
			}
			call("", nil, initArgs...)
			call("", nil, "remember", "Memory body", "--id", "beads/memory", "--title", "Memory")
			original := graphMixedResult[graphstore.IssueRecord](t, call("", nil, "create", "Edit me", "--id", "beads/work", "--description", "Original description", "--notes", "Original notes"))
			versions := call("", nil, "versions", "work")
			unchanged := graphMixedResult[graphstore.IssueMutationResult](t, call("", editor("exit 0"), "edit", "work"))
			if unchanged.Changed || !reflect.DeepEqual(unchanged.Issue, original) || call("", nil, "versions", "work") != versions {
				t.Fatal("unedited tempfile changed Issue or History")
			}
			changed := graphMixedResult[graphstore.IssueMutationResult](t, call("", editor(`printf 'Edited description\n' > "$1"`), "edit", "work"))
			if !changed.Changed || changed.Issue.Properties.Description != "Edited description" || changed.Issue.Revision == original.Revision {
				t.Fatalf("description edit: %+v", changed)
			}
			_, _, rows, _ := graphVersionsListed(t, call("", nil, "versions", "work"))
			if len(rows) != 2 {
				t.Fatalf("description edit should record one native version, got %d", len(rows))
			}
			notes := graphMixedResult[graphstore.IssueMutationResult](t, call("", editor(`printf 'Replacement notes\n' > "$1"`), "edit", "work", "--notes"))
			if !notes.Changed || notes.Issue.Properties.Notes != "Replacement notes" || notes.Issue.Properties.Description != "Edited description" {
				t.Fatalf("notes edit changed wrong field: %+v", notes)
			}
			beforeRefusals := call("", nil, "versions", "work")
			call("invalid_properties", nil, "edit", "work", "--title", "--notes")
			call("invalid_properties", editor(`printf '   ' > "$1"`), "edit", "work", "--title")
			call("permission_denied", nil, "edit", "work", "--readonly")
			call("capability_unavailable", nil, "edit", "memory")
			if got := call("", nil, "versions", "work"); got != beforeRefusals {
				t.Fatal("refused edits changed History")
			}
			if shown := graphMixedResult[graphstore.IssueRecord](t, call("", nil, "show", "work", "--format", "graph-json")); !reflect.DeepEqual(shown, notes.Issue) {
				t.Fatal("refused edits changed Issue")
			}
			if !strings.Contains(call("", editor(`printf 'Renamed\n' > "$1"`), "edit", "work", "--title"), "Renamed") {
				t.Fatal("title edit did not return the updated Issue")
			}
			beforeRace := call("", nil, "versions", "work")
			raceEditor := editor(`printf 'Lost description\n' > "$1"
"$BD_BIN" update work --title Concurrent --json >/dev/null`)
			call("revision_conflict", append(raceEditor, "BD_BIN="+bd), "edit", "work")
			afterRace := graphMixedResult[graphstore.IssueRecord](t, call("", nil, "show", "work", "--format", "graph-json"))
			if afterRace.Properties.Title != "Concurrent" || afterRace.Properties.Description != "Edited description" || call("", nil, "versions", "work") == beforeRace {
				t.Fatal("stale editor overwrote a concurrent Issue update")
			}
			_, _, rowsBeforeUntouchedRace, _ := graphVersionsListed(t, call("", nil, "versions", "work"))
			untouchedRaceEditor := editor(`"$BD_BIN" update work --title UntouchedRace --json >/dev/null`)
			call("revision_conflict", append(untouchedRaceEditor, "BD_BIN="+bd), "edit", "work")
			current := graphMixedResult[graphstore.IssueRecord](t, call("", nil, "show", "work", "--format", "graph-json"))
			_, _, rowsAfterUntouchedRace, _ := graphVersionsListed(t, call("", nil, "versions", "work"))
			if current.Properties.Title != "UntouchedRace" || current.Properties.Description != "Edited description" || len(rowsAfterUntouchedRace) != len(rowsBeforeUntouchedRace)+1 {
				t.Fatal("unchanged editor content returned stale success or added History")
			}
		})
	}
}
