package main

import (
	"github.com/spf13/cobra"
	graph "github.com/steveyegge/beads/graphops"
)

// The noun/verb surface uses the same graph operations as the established
// spellings. The bare `bd link A B` path keeps its ordinary Issue semantics.
var (
	graphLinkAddCmd = &cobra.Command{
		Use: "add SOURCE TARGET", Short: "Create a typed Link between two Beads",
		Args: cobra.ExactArgs(2), SilenceUsage: true, SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !graphPreviewActive {
				return graphFailure("capability_unavailable", "link add requires a graph preview workspace", 5)
			}
			if !graphPreviewLinkTypeChanged(cmd) {
				return graphFailure("invalid_selector", "link add requires --link-type types/NAME", 2)
			}
			return runGraphPreviewLink(cmd, args)
		},
	}
	graphLinkListCmd = &cobra.Command{
		Use: "list BEAD", Short: "List Links incident to a Bead",
		Args: cobra.ExactArgs(1), SilenceUsage: true, SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !graphPreviewActive {
				return graphFailure("capability_unavailable", "link list requires a graph preview workspace", 5)
			}
			return runGraphPreviewLinks(cmd, args)
		},
	}
	graphLinkShowCmd = &cobra.Command{
		Use: "show LINK", Short: "Show a Link by ID",
		Args: cobra.ExactArgs(1), SilenceUsage: true, SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !graphPreviewActive {
				return graphFailure("capability_unavailable", "link show requires a graph preview workspace", 5)
			}
			path, err := graphPreviewLinkCommandPath(args[0])
			if err != nil {
				return err
			}
			return runGraphPreviewShow(cmd, []string{path})
		},
	}
	graphLinkUpdateCmd = &cobra.Command{
		Use: "update LINK", Short: "Update a Link's properties or metadata",
		Args: cobra.ExactArgs(1), SilenceUsage: true, SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !graphPreviewActive {
				return graphFailure("capability_unavailable", "link update requires a graph preview workspace", 5)
			}
			path, err := graphPreviewLinkCommandPath(args[0])
			if err != nil {
				return err
			}
			return runGraphPreviewUpdate(cmd, []string{path})
		},
	}
	graphLinkRemoveCmd = &cobra.Command{
		Use: "remove LINK | SOURCE TARGET", Short: "Remove a Link by ID or unambiguous endpoint pair",
		Args: cobra.RangeArgs(1, 2), SilenceUsage: true, SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !graphPreviewActive {
				return graphFailure("capability_unavailable", "link remove requires a graph preview workspace", 5)
			}
			return runGraphPreviewUnlink(cmd, args)
		},
	}
)

func graphPreviewLinkCommandPath(selector string) (string, error) {
	path, err := graphPreviewLinkPath(graphPreviewConfig.GraphScopeURL, selector)
	if err != nil {
		return "", graphFailure("invalid_selector", err.Error(), 2)
	}
	if err := graph.ValidateLinkPath(path); err != nil {
		return "", graphFailure("invalid_selector", err.Error(), 2)
	}
	return path, nil
}

func init() {
	linkCmd.AddCommand(graphLinkAddCmd, graphLinkListCmd, graphLinkShowCmd, graphLinkUpdateCmd, graphLinkRemoveCmd)
	registerGraphLinkTypeFlag(graphLinkAddCmd)
	graphLinkAddCmd.Flags().String("id", "", "New Link ID or links/PATH")
	graphLinkAddCmd.Flags().String("properties", "", "Initial Link properties as JSON, @file, or @-")
	graphLinkAddCmd.Flags().String("metadata", "", "Initial Link metadata as a JSON object or @file")
	graphLinkAddCmd.Flags().String("if-source-revision", "", "Require this observed owning-source revision")
	graphLinkAddCmd.Flags().Bool("unconditional-source", false, "Accept the current owning-source revision")

	registerGraphLinkTypeFlag(graphLinkListCmd)
	graphLinkListCmd.Flags().String("direction", "both", "Incident direction: in, out, or both")
	graphLinkShowCmd.Flags().String("version", "", "Read an exact retained Link version")

	graphLinkUpdateCmd.Flags().String("properties", "", "Replace Link properties from JSON, @file, or @-")
	graphLinkUpdateCmd.Flags().String("patch", "", "Apply ordered Link property operations from JSON, @file, or @-")
	graphLinkUpdateCmd.Flags().String("metadata", "", "Merge a JSON object into Link metadata")
	graphLinkUpdateCmd.Flags().StringArray("set-metadata", nil, "Set Link metadata key=value (repeatable)")
	graphLinkUpdateCmd.Flags().StringArray("unset-metadata", nil, "Remove Link metadata key (repeatable)")
	graphLinkUpdateCmd.Flags().String("if-revision", "", "Require this observed Link revision")
	graphLinkUpdateCmd.Flags().String("if-source-revision", "", "Require this observed owning-source revision")
	graphLinkUpdateCmd.Flags().Bool("unconditional-source", false, "Accept the current owning-source revision")

	registerGraphLinkTypeFlag(graphLinkRemoveCmd)
	graphLinkRemoveCmd.Flags().String("if-revision", "", "Require this observed Link revision")
	graphLinkRemoveCmd.Flags().Bool("unconditional", false, "Explicitly accept the current Link revision")
	graphLinkRemoveCmd.Flags().String("if-source-revision", "", "Require this observed owning-source revision")
	graphLinkRemoveCmd.Flags().Bool("unconditional-source", false, "Accept the current owning-source revision")
}
