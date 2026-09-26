package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

var graphUnlinkCmd = &cobra.Command{
	Use: "unlink LINK | SOURCE TARGET", GroupID: "issues",
	Short: "Remove an informational Link in an experimental graph workspace",
	Long: `Remove one informational Link by its canonical identity, or select one
unambiguous source/target pair using --resource-type. Requires a Link revision
guard and, when Memory owns it, a source guard. The ID stays reserved and prior
snapshots remain retained. Blocking Dependency removal is not yet supported.`,
	SilenceUsage: true, SilenceErrors: true,
	RunE: runGraphPreviewUnlink,
}

var graphLinksCmd = &cobra.Command{
	Use: "links BEAD", GroupID: "issues",
	Short: "Inspect incident Links in an experimental graph workspace",
	Long: `List current informational Links and blocking Dependencies incident to a
canonical Bead. Direction is relative to that Bead; default is both. This bounded
preview returns a complete result or refuses above its advertised limit. It does
not paginate, read historical state, or resolve remote targets.`,
	SilenceUsage: true, SilenceErrors: true,
	RunE: runGraphPreviewLinks,
}

func init() {
	rootCmd.AddCommand(graphUnlinkCmd, graphLinksCmd)
	graphUnlinkCmd.Flags().String("resource-type", "", "Installed Link Type URL for pair selection")
	graphUnlinkCmd.Flags().String("if-revision", "", "Require this observed Link revision")
	graphUnlinkCmd.Flags().Bool("unconditional", false, "Explicitly accept the current Link revision")
	graphUnlinkCmd.Flags().String("if-source-revision", "", "Require this observed owning-source revision")
	graphUnlinkCmd.Flags().Bool("unconditional-source", false, "Explicitly accept the current owning-source revision")
	graphLinksCmd.Flags().String("direction", "both", "Incident direction: in, out, or both")
	graphLinksCmd.Flags().String("resource-type", "", "Filter by exact installed Link Type URL")
}

func graphPreviewBeadSelector(selector string) (string, error) {
	path, err := graphPreviewResourcePath(graphPreviewConfig.GraphScopeURL, selector)
	if err != nil {
		return "", graphFailure("invalid_selector", err.Error(), 2)
	}
	if err := graph.ValidateBeadPath(path); err != nil {
		return "", graphFailure("invalid_selector", err.Error(), 2)
	}
	return path, nil
}

func runGraphPreviewLinks(cmd *cobra.Command, args []string) error {
	if err := graphPreviewFlags(cmd, "direction", "resource-type"); err != nil {
		return err
	}
	if len(args) != 1 {
		return graphFailure("invalid_selector", "links requires exactly one canonical Bead selector", 2)
	}
	path, err := graphPreviewBeadSelector(args[0])
	if err != nil {
		return err
	}
	direction, _ := cmd.Flags().GetString("direction")
	if direction != "in" && direction != "out" && direction != "both" {
		return graphFailure("invalid_selector", "direction must be in, out, or both", 2)
	}
	typ, _ := cmd.Flags().GetString("resource-type")
	if cmd.Flags().Changed("resource-type") && typ == "" {
		return graphFailure("invalid_selector", "resource-type must be an installed Link Type URL", 2)
	}
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		result, err := store.ListLinks(ctx, graphstore.LinksRequest{BeadPath: path, Direction: direction, TypeURL: typ})
		if err != nil {
			return nil, "", err
		}
		var human strings.Builder
		for _, link := range result {
			fmt.Fprintf(&human, "%s  %s → %s\n", link.ID, link.Source, link.Target)
		}
		if len(result) == 0 {
			human.WriteString("No incident Links.\n")
		}
		return result, strings.TrimSuffix(human.String(), "\n"), nil
	})
}

func runGraphPreviewUnlink(cmd *cobra.Command, args []string) error {
	if err := graphPreviewWritePolicy(); err != nil {
		return err
	}
	if err := graphPreviewFlags(cmd, "resource-type", "if-revision", "unconditional", "if-source-revision", "unconditional-source"); err != nil {
		return err
	}
	if len(args) != 1 && len(args) != 2 {
		return graphFailure("invalid_selector", "unlink requires one canonical Link or two canonical Bead selectors", 2)
	}
	request := graphstore.LinkDeleteRequest{Actor: getActorWithGit()}
	if len(args) == 1 {
		if cmd.Flags().Changed("resource-type") {
			return graphFailure("invalid_selector", "resource-type is only used for pair selection", 2)
		}
		path, err := graphPreviewResourcePath(graphPreviewConfig.GraphScopeURL, args[0])
		if err != nil {
			return graphFailure("invalid_selector", err.Error(), 2)
		}
		if err := graph.ValidateLinkPath(path); err != nil {
			return graphFailure("invalid_selector", err.Error(), 2)
		}
		request.Path = path
	} else {
		var err error
		request.SourcePath, err = graphPreviewBeadSelector(args[0])
		if err != nil {
			return err
		}
		request.TargetPath, err = graphPreviewBeadSelector(args[1])
		if err != nil {
			return err
		}
		request.TypeURL, _ = cmd.Flags().GetString("resource-type")
		if request.TypeURL == "" {
			return graphFailure("invalid_selector", "pair unlink requires --resource-type", 2)
		}
	}
	var err error
	request.ExpectedRevision, request.Unconditional, err = graphPreviewRevisionGuard(cmd, false, true)
	if err != nil {
		return err
	}
	request.ExpectedSourceRevision, request.UnconditionalSource, err = graphPreviewRevisionGuard(cmd, true, false)
	if err != nil {
		return err
	}
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		result, err := store.Unlink(ctx, request)
		return result, fmt.Sprintf("Unlinked %s; identity remains reserved", result.Link.ID), err
	})
}
