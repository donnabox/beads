package main

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// Help is rendered before graph workspace admission, so it must describe both
// ordinary behavior and the graph preview without opening storage.
func TestGraphPreviewHelpDocumentsPlaytestCommands(t *testing.T) {
	for _, cmd := range []*cobra.Command{
		initCmd, rememberCmd, memoriesCmd, recallCmd, createCmd, showCmd,
		updateCmd, deleteCmd, forgetCmd, depCmd, depAddCmd, depRemoveCmd, linkCmd, quickCmd, closeCmd,
		reopenCmd, unclaimCmd, commentsCmd, deferCmd, undeferCmd, readyCmd, listCmd, blockedCmd, graphCmd, statusCmd,
		typesCmd, versionsCmd, historyCmd,
	} {
		t.Run(cmd.Name()+"-scope", func(t *testing.T) {
			if out := captureStdout(t, cmd.Help); !strings.Contains(out, "Graph preview workspaces:") {
				t.Fatalf("%s help omits graph preview guidance", cmd.Name())
			}
		})
	}

	for _, tc := range []struct {
		name string
		cmd  *cobra.Command
		want []string
	}{
		{"remember", rememberCmd, []string{"Graph preview workspaces:", "bd remember 'Revised policy' --id policy", "bd remember 'Revised policy' --update policy", "--create-only", "--if-revision TOKEN", "--replace-properties", "--replace-metadata", "Omitted fields remain unchanged"}},
		{"create", createCmd, []string{"Graph preview workspaces:", "--type types/memory", "--id policy", "bare --id ID keeps the shipping native Issue ID", "beads/PATH selects a graph path", "--properties JSON", "types/NAME", "--parent ID creates a dotted child ID"}},
		{"q", quickCmd, []string{"Graph preview workspaces:", "bd q -- -follow-up", "--parent is unavailable here", "Ordinary Issue workspaces:", "bd q \"Subtask\" --parent"}},
		{"dep", depCmd, []string{"Graph preview workspaces:", "bd dep BLOCKER --blocks BLOCKED", "--no-cycle-check"}},
		{"dep-remove", depRemoveCmd, []string{"Graph preview workspaces:", "bd dep remove issue blocker", "status not_found", "dependency_removed", "bd link delete ID --force"}},
		{"link", linkCmd, []string{
			"Graph preview workspaces:", "bd link", "create/list/show/update/delete/versions/compare/count", "--type types/references", "Memory or Issue",
			"accepts a built-in Dependency kind, types/NAME or a full local Type URL",
			"The single Type types/dependency", "--if-source-revision TOKEN to reject a stale source.", "Owned-source pins remain unavailable.", "current resolves",
		}},
		{"show", showCmd, []string{"Graph preview workspaces:", "ordinary detail array", "--format graph-json"}},
		{"update", updateCmd, []string{"Graph preview workspaces:", "--properties", "--replace-properties", "--replace-metadata", "--if-revision TOKEN", "omission", "preserves the existing pin"}},
		{"unclaim", unclaimCmd, []string{"Graph preview workspaces:", "bd unclaim ID...", "--force", "--if-assignee HOLDER", "--reason TEXT", "bd comments ID"}},
		{"comments", commentsCmd, []string{"Graph preview workspaces:", "bd comments ID", "outside the", "Issue version"}},
		{"list", listCmd, []string{"Graph preview workspaces:", "all installed Bead Types", "--type types/NAME", "Issue-specific filters", "newest recorded change first", "hidden unless --all is given", "--all also removes the row limit", "BEADS_MAX_ROWS refuses a page of more Beads", "--sort, --reverse", "Issues only", "a line under the header saying Memories are", "it is simply not"}},
		{"types", typesCmd, []string{"Graph preview workspaces:", "types/NAME", "--details", "--type"}},
		{"versions", versionsCmd, []string{"Graph preview workspaces:", "Use bd versions ID to list one Memory, Issue or Link's retained versions newest", "Bare ID means beads/ID; use links/PATH for a Link.", "--version TOKEN or bd compare ID --from TOKEN --to TOKEN", "BDP HTTP History", "Ordinary Issue workspaces:", "List the versions recorded for a bead by versioned history."}},
		{"history", historyCmd, []string{"Graph preview workspaces:", "bd history ID is an alias for bd versions ID.", "--limit and --events are not supported by the graph alias.", "BDP HTTP History", "Ordinary Issue workspaces:", "Show the complete version history of an issue"}},
		{"defer", deferCmd, []string{"Graph preview workspaces:", "bd defer ID...", "current revision is accepted", "--until", "bd ready wakes", "Ordinary Issue workspaces:"}},
		{"undefer", undeferCmd, []string{"Graph preview workspaces:", "bd undefer ID...", "current revision is accepted", "clears a stale defer date", "Ordinary Issue workspaces:"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := captureStdout(t, tc.cmd.Help)
			for _, want := range tc.want {
				if !strings.Contains(out, want) {
					t.Fatalf("help missing %q: %s", want, out)
				}
			}
		})
	}

	// Versions, history and dateless deferral lead with the graph section, ahead of their long
	// ordinary descriptions (versioned-history recording, Dolt commits).
	for _, cmd := range []*cobra.Command{versionsCmd, historyCmd, deferCmd, undeferCmd} {
		t.Run(cmd.Name()+"-graph-first", func(t *testing.T) {
			out := captureStdout(t, cmd.Help)
			graphAt, ordinaryAt := strings.Index(out, "Graph preview workspaces:"), strings.Index(out, "Ordinary Issue workspaces:")
			if graphAt < 0 || ordinaryAt < 0 || graphAt > ordinaryAt {
				t.Fatalf("%s help must put graph guidance before the ordinary description: %s", cmd.Name(), out)
			}
		})
	}
}
