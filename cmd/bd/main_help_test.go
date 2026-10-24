package main

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestSubcommandHelpOmitsInheritedFlagsWithoutChangingFlagSet(t *testing.T) {
	root := &cobra.Command{Use: "bd"}
	root.PersistentFlags().String("database", "", "Choose the database")
	child := &cobra.Command{Use: "show", Run: func(*cobra.Command, []string) {}}
	child.Flags().Bool("long", false, "Show details")
	root.AddCommand(child)

	usage := compactSubcommandUsage(child, child.UsageString())
	if strings.Contains(usage, "Global Flags:") || strings.Contains(usage, "--database") || !strings.Contains(usage, "--long") || !strings.Contains(usage, "bd --help") {
		t.Fatalf("subcommand help did not retain only local flags and the root pointer:\n%s", usage)
	}
	if child.InheritedFlags().Lookup("database") == nil || !strings.Contains(root.UsageString(), "--database") {
		t.Fatal("help formatting changed accepted or root flags")
	}
}
