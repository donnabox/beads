package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// These are alternate paths to the existing commands, not new Memory routes.
// Keep distinct Cobra flags so each path records its own Changed set while
// sharing the original flag values and handlers.
var (
	memoryRememberAliasCmd *cobra.Command
	memoryRecallAliasCmd   *cobra.Command
	memoryMemoriesAliasCmd *cobra.Command
	memoryForgetAliasCmd   *cobra.Command
)

func init() {
	memoryRememberAliasCmd = newMemoryAliasCommand(rememberCmd)
	memoryRecallAliasCmd = newMemoryAliasCommand(recallCmd)
	memoryMemoriesAliasCmd = newMemoryAliasCommand(memoriesCmd)
	memoryForgetAliasCmd = newMemoryAliasCommand(forgetCmd)
	memoryGraphCmd.AddCommand(memoryRememberAliasCmd, memoryRecallAliasCmd, memoryMemoriesAliasCmd, memoryForgetAliasCmd)
}

func newMemoryAliasCommand(original *cobra.Command) *cobra.Command {
	alias := &cobra.Command{
		Use: original.Use, Short: original.Short, Long: original.Long,
		Args: original.Args, RunE: original.RunE,
		SilenceUsage: original.SilenceUsage, SilenceErrors: original.SilenceErrors,
	}
	original.Flags().VisitAll(func(flag *pflag.Flag) {
		copy := *flag
		copy.Changed = false
		alias.Flags().AddFlag(&copy)
	})
	return alias
}

// Admission treats a nested convenience spelling exactly like its top-level
// command. Relationship reads under `bd memory` retain their separate route.
func graphPreviewMemoryTopCommand(cmd *cobra.Command) *cobra.Command {
	switch cmd {
	case memoryRememberAliasCmd:
		return rememberCmd
	case memoryRecallAliasCmd:
		return recallCmd
	case memoryMemoriesAliasCmd:
		return memoriesCmd
	case memoryForgetAliasCmd:
		return forgetCmd
	default:
		return nil
	}
}
