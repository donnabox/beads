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
	Short: "Remove a Link in an experimental graph workspace",
	Long: `Remove one informational Link or blocking Dependency by canonical Link ID.
Informational Links also accept an unambiguous source/target pair selected with
--link-type. Requires a Link revision guard. Source revision protection is
optional for Memory-owned Links and required for blocking Dependencies. The ID
stays reserved and prior snapshots remain retained. Blocking pair selection is
not supported.`,
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
	registerGraphLinkTypeFlag(graphUnlinkCmd)
	graphUnlinkCmd.Flags().String("if-revision", "", "Require this observed Link revision")
	graphUnlinkCmd.Flags().Bool("unconditional", false, "Explicitly accept the current Link revision")
	graphUnlinkCmd.Flags().String("if-source-revision", "", "Require this observed owning-source revision")
	graphUnlinkCmd.Flags().Bool("unconditional-source", false, "Accept the current owning-source revision (default without --if-source-revision)")
	graphLinksCmd.Flags().String("direction", "both", "Incident direction: in, out, or both")
	registerGraphLinkTypeFlag(graphLinksCmd)
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
	if err := graphPreviewFlags(cmd, "direction", "link-type", "resource-type"); err != nil {
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
	typ, err := graphPreviewLinkType(cmd)
	if err != nil {
		return err
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
	if err := graphPreviewFlags(cmd, "link-type", "resource-type", "if-revision", "unconditional", "if-source-revision", "unconditional-source"); err != nil {
		return err
	}
	if len(args) != 1 && len(args) != 2 {
		return graphFailure("invalid_selector", "unlink requires one canonical Link or two canonical Bead selectors", 2)
	}
	request := graphstore.LinkDeleteRequest{Actor: getActorWithGit()}
	if len(args) == 1 {
		if graphPreviewLinkTypeChanged(cmd) {
			return graphFailure("invalid_selector", "--link-type is only used for pair selection", 2)
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
		request.TypeURL, err = graphPreviewLinkType(cmd)
		if err != nil {
			return err
		}
		if request.TypeURL == "" {
			return graphFailure("invalid_selector", "pair unlink requires --link-type", 2)
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
	if !cmd.Flags().Changed("if-source-revision") && !cmd.Flags().Changed("unconditional-source") {
		request.UnconditionalSource = false
		request.DefaultInformationalSource = true
	}
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		result, err := store.Unlink(ctx, request)
		return result, graphPreviewReplacementSummary(fmt.Sprintf("Unlinked %s; identity remains reserved", result.Link.ID), result.ReplacedSource), err
	})
}
