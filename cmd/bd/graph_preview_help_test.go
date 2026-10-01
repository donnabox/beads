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
		updateCmd, deleteCmd, forgetCmd, depAddCmd, linkCmd, closeCmd,
		reopenCmd, readyCmd, listCmd, blockedCmd, graphCmd, statusCmd,
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
		{"remember", rememberCmd, []string{"Graph preview workspaces:", "bd remember 'Revised policy' --update beads/policy", "--if-revision TOKEN", "Omitted fields remain unchanged"}},
		{"create", createCmd, []string{"Graph preview workspaces:", "--bead-type types/preview-memory-v2", "--id beads/"}},
		{"link", linkCmd, []string{"Graph preview workspaces:", "--link-type types/example-cites", "Memory or Issue"}},
		{"update", updateCmd, []string{"Graph preview workspaces:", "--properties", "--if-revision TOKEN"}},
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
}
