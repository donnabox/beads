//go:build cgo

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
)

func TestGraphPreviewIssueCommentWriteWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			var workURL string
			args := []string{"init", "--graph-mode", "link", "--scope-url", "https://example.invalid/comments/", "--skip-hooks", "--skip-agents", "--non-interactive"}
			if engine == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for shared-server qualification")
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
			read := func() graphstore.IssueRecord {
				t.Helper()
				return graphMixedResult[graphstore.IssueRecord](t, call("show", workURL, "--format", "graph-json"))
			}
			decode := func(output string) types.Comment {
				t.Helper()
				var comment types.Comment
				if err := json.Unmarshal([]byte(output), &comment); err != nil {
					t.Fatalf("decode native comment: %v\n%s", err, output)
				}
				return comment
			}
			call(args...)
			created := graphMixedResult[graphstore.IssueRecord](t, call("create", "Comment target", "--graph-path", "beads/work"))
			workURL = created.ID
			if workURL != "https://example.invalid/comments/"+created.Properties.ID || workURL == "https://example.invalid/comments/beads/work" {
				t.Fatal("fixture lost public/private identity separation")
			}
			call("remember", "Context", "--id", "beads/memory")
			before := read()
			call("remember", "Collision", "--id", "beads/"+before.Properties.ID)
			versions := call("versions", workURL)
			first := decode(call("comment", workURL, "Starting", "review", "--actor", "alice"))
			if first.ID == "" || first.IssueID != before.Properties.ID || first.Author != "alice" || first.Text != "Starting review" {
				t.Fatalf("singular comment did not return stored row: %+v", first)
			}
			file := filepath.Join(work, "followup.txt")
			writeFile(t, file, []byte("Follow-up from file\n"))
			second := decode(call("comments", "add", workURL, "--file", file, "--author", "bob"))
			if second.ID == first.ID || second.IssueID != before.Properties.ID || second.Author != "bob" || second.Text != "Follow-up from file\n" {
				t.Fatalf("plural comment did not return stored row: %+v", second)
			}
			var listed []*types.Comment
			listedJSON := call("comments", workURL)
			if err := json.Unmarshal([]byte(listedJSON), &listed); err != nil {
				t.Fatalf("comments --json must retain ordinary array shape: %v\n%s", err, listedJSON)
			}
			if len(listed) != 2 || listed[0].ID != first.ID || listed[1].ID != second.ID || listed[0].IssueID != before.Properties.ID || listed[1].IssueID != before.Properties.ID {
				t.Fatalf("comment feed lost or duplicated a row: %+v", listed)
			}
			after := read()
			if after.Revision != before.Revision || call("versions", workURL) != versions {
				t.Fatal("comments minted an Issue Resource version")
			}
			for _, command := range [][]string{
				{"comment", workURL, "Quiet singular", "--quiet"},
				{"comments", "add", workURL, "Quiet plural", "--quiet"},
			} {
				if got := graphPolicyCLI(t, bd, work, home, nil, "", command...); got != "" {
					t.Fatalf("quiet comment %v printed %q", command, got)
				}
			}
			var afterQuiet []*types.Comment
			quietFeed := call("comments", workURL)
			if err := json.Unmarshal([]byte(quietFeed), &afterQuiet); err != nil || len(afterQuiet) != 4 {
				t.Fatalf("quiet comment writes did not both persist: %v\n%s", err, quietFeed)
			}
			refuse("invalid_properties", "comment", workURL, " ")
			refuse("permission_denied", "comment", workURL, "No", "--readonly")
			refuse("permission_denied", "comments", "add", workURL, "No", "--readonly")
			refuse("invalid_properties", "comment", "beads/memory", "No")
			refuse("invalid_properties", "comments", "add", "beads/memory", "No")
		})
	}
}
