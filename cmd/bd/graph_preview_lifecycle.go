package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cobra"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
)

var graphUnlinkCmd = &cobra.Command{
	Use: "unlink LINK | SOURCE TARGET", GroupID: "issues",
	Short: "Remove a Link in an experimental graph workspace",
	Long: `Remove one informational Link or blocking Dependency by canonical Link ID.
Informational Links also accept an unambiguous source/target pair selected with
--type; local endpoint spellings select public Bead IDs, while canonical Scope
URLs select exact graph paths. Requires a Link revision guard. Source revision protection is
optional for Memory-owned Links and required for blocking Dependencies. The ID
stays reserved and prior snapshots remain retained. Blocking pair selection is
not supported by unlink; use bd dep remove ISSUE BLOCKER for the native pair
verb and its absent-pair no-op.`,
	SilenceUsage: true, SilenceErrors: true,
	RunE: runGraphPreviewUnlink,
}

var graphLinksCmd = &cobra.Command{
	Use: "links BEAD", GroupID: "issues",
	Short: "Inspect incident Links in an experimental graph workspace",
	Long: `List current informational Links and blocking Dependencies incident to a
public Bead ID (including a literal beads/PATH spelling) or canonical Scope URL.
Direction is relative to that Bead; default is both. This bounded
preview returns a complete result or refuses above its advertised limit. It does
not paginate, read historical state, or resolve remote targets.`,
	SilenceUsage: true, SilenceErrors: true,
	RunE: runGraphPreviewLinks,
}

var linkCreateCmd = &cobra.Command{
	Use: "create [SOURCE TARGET]", Short: "Create a Link in a graph workspace",
	Args: cobra.MaximumNArgs(2), SilenceUsage: true, SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !cmd.Flags().Changed("type") {
			return graphFailure("invalid_selector", "link create requires --type", 2)
		}
		if err := canonicalExactOccurrences(cmd, os.Args[1:], "from", "to"); err != nil {
			return graphFailure("invalid_selector", "link create "+err.Error(), 2)
		}
		source, target, err := graphPreviewAuthoringEndpoints(cmd, args)
		if err != nil {
			return graphFailure("invalid_selector", "link create "+err.Error(), 2)
		}
		return runGraphPreviewLink(cmd, []string{source, target})
	},
}

var linkListCmd = &cobra.Command{
	Use: "list", Short: "List current Links in a graph workspace",
	Args: cobra.NoArgs, SilenceUsage: true, SilenceErrors: true, RunE: runGraphPreviewLinkCollection,
}

var linkCountCmd = &cobra.Command{
	Use: "count", Short: "Count current Links matching the collection filters",
	Args: cobra.NoArgs, SilenceUsage: true, SilenceErrors: true, RunE: runGraphPreviewLinkCollection,
}

var linkShowCmd = &cobra.Command{
	Use: "show LINK", Short: "Show a Link by ID",
	Args: func(cmd *cobra.Command, args []string) error {
		if canonicalOutputSelected(cmd) {
			return nil
		}
		return cobra.ExactArgs(1)(cmd, args)
	}, SilenceUsage: true, SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGraphPreviewShowKind(cmd, args, true)
	},
}

var linkUpdateCmd = &cobra.Command{
	Use: "update LINK", Short: "Update a Link or recover it from a retained revision",
	Args: cobra.ExactArgs(1), SilenceUsage: true, SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		selector, err := graphPreviewWritableResourceArgument(args[0])
		if err != nil {
			return err
		}
		path, err := graphPreviewResolvedHistoricalPath(selector)
		if err != nil {
			return err
		}
		if err := graph.ValidateLinkPath(path); err != nil {
			return graphFailure("invalid_selector", err.Error(), 2)
		}
		return runGraphPreviewUpdate(cmd, []string{selector})
	},
}

var linkDeleteCmd = &cobra.Command{
	Use: "delete LINK", Short: "Preview or delete a Link by ID",
	Args: cobra.ExactArgs(1), SilenceUsage: true, SilenceErrors: true, RunE: runGraphPreviewLinkDelete,
}

var linkVersionsCmd = &cobra.Command{
	Use: "versions LINK", Short: "List a Link's retained versions",
	Args: cobra.ExactArgs(1), SilenceUsage: true, SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if canonicalOutputSelected(cmd) {
			return runGraphPreviewVersions(cmd, args)
		}
		path, err := graphPreviewResolvedHistoricalPath(args[0])
		if err != nil {
			return err
		}
		if err := graph.ValidateLinkPath(path); err != nil {
			return graphFailure("invalid_selector", err.Error(), 2)
		}
		return runGraphPreviewVersions(cmd, args)
	},
}

var linkCompareCmd = &cobra.Command{
	Use: "compare LINK --from TOKEN --to TOKEN", Short: "Compare two retained Link versions",
	Args: cobra.ExactArgs(1), SilenceUsage: true, SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if canonicalOutputSelected(cmd) {
			return runGraphPreviewCompare(cmd, args)
		}
		path, err := graphPreviewResolvedHistoricalPath(args[0])
		if err != nil {
			return err
		}
		if err := graph.ValidateLinkPath(path); err != nil {
			return graphFailure("invalid_selector", err.Error(), 2)
		}
		return runGraphPreviewCompare(cmd, args)
	},
}

func init() {
	linkCmd.AddCommand(linkCreateCmd, linkListCmd, linkShowCmd, linkUpdateCmd, linkDeleteCmd, linkVersionsCmd, linkCompareCmd, linkCountCmd)
	linkCountCmd.Flags().String("exact-type", "", "Select only the exact declared Link Type")
	graphLinksCmd.Flags().String("exact-type", "", "Select only the exact declared Link Type")
	linkCreateCmd.Flags().String("type", "", "Installed Link Type or built-in Dependency type")
	for _, flag := range []string{"from", "to"} {
		linkCreateCmd.Flags().String(flag, "", "Local public Bead ID or canonical Scope URL")
	}
	linkCreateCmd.Flags().String("source-revision", "", "Source endpoint revision: exact token, current or unpinned; owned source pins refuse (distinct from --if-source-revision)")
	linkCreateCmd.Flags().String("target-revision", "", "Target endpoint revision: exact token, current or unpinned (informational Links only)")
	linkCreateCmd.Flags().String("id-root", "", "Prefix for an automatically generated Link ID")
	linkCreateCmd.Flags().String("id", "", "Optional local Link ID or links/PATH")
	linkCreateCmd.Flags().String("properties", "", "Initial informational Link properties as JSON, @file, or @-")
	linkCreateCmd.Flags().String("metadata", "", "Initial informational Link metadata as JSON or @file")
	linkCreateCmd.Flags().String("if-source-revision", "", "Require the owning Bead revision")
	for _, command := range []*cobra.Command{linkListCmd, linkCountCmd} {
		registerGraphLinkTypeFlag(command)
		if command == linkListCmd {
			command.Flags().Lookup("type").Usage += "; includes refinements (use --exact-type for nominal selection)"
		}
		command.Flags().String("from", "", "Select Links with this local source Bead")
		command.Flags().String("to", "", "Select Links with this local target Bead")
		command.Flags().String("incident-to", "", "Select Links incident to this local Bead")
		command.Flags().String("direction", "both", "With --incident-to: in, out, or both")
	}
	linkShowCmd.Flags().Bool("long", false, "Include exact revision token and administrative details")
	linkShowCmd.Flags().String("revision", "", "Read an exact retained Link version")
	linkShowCmd.Flags().String("format", "", "Use graph-json for the complete graph record")
	for _, command := range []*cobra.Command{linkUpdateCmd} {
		command.Flags().SetNormalizeFunc(normalizeHistoricalRevisionFlag)
		command.Flags().String("from-revision", "", "Compose editable Link content from this exact retained live revision; recover a deleted Link with the same ID")
		command.Flags().String("source-revision", "", "Re-pin the source to an exact token or current, or unpin with unpinned; owned source pins refuse")
		command.Flags().String("target-revision", "", "Re-pin the target to an exact token or current, or unpin with unpinned")
		command.Flags().String("properties", "", "Merge named Link properties from JSON, @file, or @-")
		command.Flags().String("replace-properties", "", "Replace writable Link properties from JSON, @file, or @-")
		command.Flags().String("patch", "", "Apply ordered Link property operations from JSON, @file, or @-")
		command.Flags().String("metadata", "", "Merge Link metadata from JSON or @file")
		command.Flags().String("replace-metadata", "", "Replace Link metadata from JSON or @file")
		command.Flags().StringArray("set-metadata", nil, "Set a Link metadata key=value")
		command.Flags().StringArray("unset-metadata", nil, "Unset a Link metadata key")
		command.Flags().String("if-revision", "", "Require the Link revision")
		command.Flags().String("if-source-revision", "", "Require the owning Bead revision")
	}
	linkDeleteCmd.Flags().Bool("force", false, "Apply deletion after preview")
	linkDeleteCmd.Flags().String("if-revision", "", "Require the Link revision")
	linkDeleteCmd.Flags().String("if-source-revision", "", "Require the owning Bead revision")
	linkCompareCmd.Flags().String("from", "", "Exact retained source version token")
	linkCompareCmd.Flags().String("to", "", "Exact retained destination version token")
	registerGraphLinkTypeFlag(graphUnlinkCmd)
	graphUnlinkCmd.Flags().String("if-revision", "", "Require this observed Link revision")
	graphUnlinkCmd.Flags().String("if-source-revision", "", "Require this observed owning-source revision")
	graphLinksCmd.Flags().String("direction", "both", "Incident direction: in, out, or both")
	registerGraphLinkTypeFlag(graphLinksCmd)
}

func graphPreviewBeadSelector(selector string) (string, error) {
	path, err := graphPreviewResolvedHistoricalPath(selector)
	if err != nil {
		return "", err
	}
	if err := graph.ValidateBeadPath(path); err != nil {
		return "", graphFailure("invalid_selector", err.Error(), 2)
	}
	return path, nil
}

func runGraphPreviewLinks(cmd *cobra.Command, args []string) error {
	if err := graphPreviewFlags(cmd, "direction", "type", "exact-type"); err != nil {
		return err
	}
	if len(args) != 1 {
		return graphFailure("invalid_selector", "links requires exactly one Bead ID or beads/PATH", 2)
	}
	path, err := graphPreviewBeadSelector(args[0])
	if err != nil {
		return err
	}
	direction, _ := cmd.Flags().GetString("direction")
	if direction != "in" && direction != "out" && direction != "both" {
		return graphFailure("invalid_selector", "direction must be in, out, or both", 2)
	}
	typ, family, err := graphPreviewCollectionTypes(cmd, graphPreviewConfig.GraphScopeURL, true)
	if err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		if typ != "" {
			if err := graphPreviewRequireInstalledType(ctx, store, typ, graph.KindLink, false); err != nil {
				return nil, "", err
			}
		}
		result, err := store.ListLinks(ctx, graphstore.LinksRequest{BeadPath: path, Direction: direction, TypeURL: typ, ConformsTo: family})
		if err != nil {
			return nil, "", err
		}
		var human strings.Builder
		for _, link := range result {
			fmt.Fprintln(&human, graphHumanRow(store.ScopeURL(), link))
		}
		if len(result) == 0 {
			human.WriteString("No incident Links.\n")
		}
		return result, strings.TrimSuffix(human.String(), "\n"), nil
	})
}

func graphPreviewFilterLinkKind(cmd *cobra.Command, links []graphstore.LinkRecord) []graphstore.LinkRecord {
	if !cmd.Flags().Changed("type") {
		return links
	}
	selector, _ := cmd.Flags().GetString("type")
	if graphPreviewNominalSelector(selector) {
		return links
	}
	kind := string(canonicalDependencyType(types.DependencyType(selector)))
	matching := links[:0]
	for _, link := range links {
		if link.Properties["dep_type"] == kind {
			matching = append(matching, link)
		}
	}
	return matching
}

// Only the human view shortens URLs. JSON retains canonical IDs and foreign
// endpoints are never made to look local.
func graphPreviewDisplayLocalURL(url string) string {
	if id, ok := graphPreviewPublicResourceID(graphPreviewConfig.GraphScopeURL, url); ok {
		return id
	}
	return url
}

func graphPreviewLinkCollectionRequest(cmd *cobra.Command) (graphstore.LinkCollectionRequest, error) {
	var request graphstore.LinkCollectionRequest
	if err := graphPreviewFlags(cmd, "type", "exact-type", "conforms-to", "from", "to", "incident-to", "direction"); err != nil {
		return request, err
	}
	var err error
	request.TypeURL, request.ConformsTo, err = graphPreviewCollectionTypes(cmd, graphPreviewConfig.GraphScopeURL, true)
	if err != nil {
		return request, graphFailure("invalid_selector", err.Error(), 2)
	}
	for _, field := range []struct {
		name string
		dest *string
	}{
		{"from", &request.SourcePath}, {"to", &request.TargetPath}, {"incident-to", &request.IncidentPath},
	} {
		if !cmd.Flags().Changed(field.name) {
			continue
		}
		selector, _ := cmd.Flags().GetString(field.name)
		if selector == "" {
			return request, graphFailure("invalid_selector", "--"+field.name+" requires a Bead", 2)
		}
		*field.dest, err = graphPreviewBeadSelector(selector)
		if err != nil {
			return request, err
		}
	}
	direction, _ := cmd.Flags().GetString("direction")
	if direction != "both" && direction != "in" && direction != "out" {
		return request, graphFailure("invalid_selector", "direction must be in, out, or both", 2)
	}
	if cmd.Flags().Changed("direction") && request.IncidentPath == "" {
		return request, graphFailure("invalid_selector", "--direction requires --incident-to", 2)
	}
	if direction == "in" && request.TargetPath != "" || direction == "out" && request.SourcePath != "" {
		return request, graphFailure("invalid_selector", "--direction conflicts with an explicit endpoint filter on the same side", 2)
	}
	if direction == "in" {
		request.TargetPath, request.IncidentPath = request.IncidentPath, ""
	}
	if direction == "out" {
		request.SourcePath, request.IncidentPath = request.IncidentPath, ""
	}
	return request, nil
}

func runGraphPreviewLinkCollection(cmd *cobra.Command, _ []string) error {
	if cmd.Name() == "list" && canonicalOutputSelected(cmd) {
		return runCanonicalCollection(cmd, true)
	}
	request, err := graphPreviewLinkCollectionRequest(cmd)
	if err != nil {
		return err
	}
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		if request.TypeURL != "" {
			if err := graphPreviewRequireInstalledType(ctx, store, request.TypeURL, graph.KindLink, false); err != nil {
				return nil, "", err
			}
		}
		links, err := store.CollectLinks(ctx, request)
		if err != nil {
			return nil, "", err
		}
		if cmd.Name() == "count" {
			result := map[string]any{"count": len(links)}
			return result, fmt.Sprintf("%d", len(links)), nil
		}
		var human strings.Builder
		for _, link := range links {
			fmt.Fprintln(&human, graphHumanRow(store.ScopeURL(), link))
		}
		if len(links) == 0 {
			human.WriteString("No Links.\n")
		}
		return links, strings.TrimSuffix(human.String(), "\n"), nil
	})
}

// The default is an informational preview. --force enters the existing
// transactional Unlink writer, which checks both Link and source revisions.
func runGraphPreviewLinkDelete(cmd *cobra.Command, args []string) error {
	if err := graphPreviewFlags(cmd, "force", "if-revision", "if-source-revision"); err != nil {
		return err
	}
	selector, err := graphPreviewWritableResourceArgument(args[0])
	if err != nil {
		return err
	}
	path, err := graphPreviewResolvedHistoricalPath(selector)
	if err != nil {
		return err
	}
	revision, _ := cmd.Flags().GetString("if-revision")
	sourceRevision, _ := cmd.Flags().GetString("if-source-revision")
	if cmd.Flags().Changed("if-source-revision") && sourceRevision == "" {
		return graphFailure("invalid_selector", "--if-source-revision requires a token", 2)
	}
	if cmd.Flags().Changed("if-revision") && revision == "" {
		return graphFailure("invalid_selector", "--if-revision requires a token", 2)
	}
	for _, token := range []struct{ name, value string }{{"if-revision", revision}, {"if-source-revision", sourceRevision}} {
		if !utf8.ValidString(token.value) || len(token.value) > graphstore.PreviewVersionTokenLimit {
			return graphFailure("invalid_selector", fmt.Sprintf("--%s requires a UTF-8 token of at most %d bytes", token.name, graphstore.PreviewVersionTokenLimit), 2)
		}
	}
	force, _ := cmd.Flags().GetBool("force")
	if force {
		if err := graphPreviewWritePolicy(); err != nil {
			return err
		}
		request := graphstore.LinkDeleteRequest{Path: path, Actor: getActorWithGit(),
			ExpectedRevision: revision, Unconditional: revision == "", ExpectedSourceRevision: sourceRevision,
			UnconditionalSource: sourceRevision == ""}
		return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
			result, err := store.Unlink(ctx, request)
			return graphPreviewReplacementResult(result, fmt.Sprintf("Unlinked %s; identity remains reserved", result.Link.ID), result.ReplacedSource, err)
		})
	}
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		record, err := store.Read(ctx, path)
		if err != nil {
			return nil, "", err
		}
		link, ok := record.(graphstore.LinkRecord)
		if !ok {
			return nil, "", graphFailure("invalid_selector", "expected a Link", 2)
		}
		if revision != "" && revision != link.Revision {
			return nil, "", graphFailure("revision_conflict", "Link revision differs from --if-revision", 4)
		}
		if sourceRevision != "" {
			_, local := graphPreviewPublicResourceID(graphPreviewConfig.GraphScopeURL, link.Source)
			if !local {
				return nil, "", graphFailure("not_authority", "Link source is not a local Bead", 5)
			}
			source, err := store.Read(ctx, link.Source)
			if err != nil {
				return nil, "", err
			}
			var current string
			switch value := source.(type) {
			case graphstore.Record:
				current = value.Revision
			case graphstore.IssueRecord:
				current = value.Revision
			}
			if current != sourceRevision {
				return nil, "", graphFailure("revision_conflict", "source revision differs from --if-source-revision", 4)
			}
		}
		return map[string]any{"preview": true, "link": link}, fmt.Sprintf("Would delete %s; use --force to apply", graphPreviewDisplayLocalURL(link.ID)), nil
	})
}

func runGraphPreviewUnlink(cmd *cobra.Command, args []string) error {
	if err := graphPreviewWritePolicy(); err != nil {
		return err
	}
	if err := graphPreviewFlags(cmd, "type", "if-revision", "if-source-revision"); err != nil {
		return err
	}
	if len(args) != 1 && len(args) != 2 {
		return graphFailure("invalid_selector", "unlink requires one Link ID (or links/PATH) or two Bead IDs", 2)
	}
	request := graphstore.LinkDeleteRequest{Actor: getActorWithGit()}
	if len(args) == 1 {
		if graphPreviewLinkTypeChanged(cmd) {
			return graphFailure("invalid_selector", "--type is only used for pair selection", 2)
		}
		selector, err := graphPreviewWritableResourceArgument(args[0])
		if err != nil {
			return err
		}
		path, err := graphPreviewResolvedHistoricalPath(selector)
		if err != nil {
			return err
		}
		if err := graph.ValidateLinkPath(path); err != nil {
			return graphFailure("invalid_selector", err.Error(), 2)
		}
		request.Path = path
	} else {
		var err error
		source, err := graphPreviewWritableResourceArgument(args[0])
		if err != nil {
			return err
		}
		target, err := graphPreviewWritableResourceArgument(args[1])
		if err != nil {
			return err
		}
		sourceSelector, err := graphPreviewIssueSelector(source)
		if err != nil {
			return graphFailure("invalid_selector", err.Error(), 2)
		}
		targetSelector, err := graphPreviewIssueSelector(target)
		if err != nil {
			return graphFailure("invalid_selector", err.Error(), 2)
		}
		request.SourcePath, request.SourceLocalID = sourceSelector.Path, sourceSelector.NativeID
		request.TargetPath, request.TargetLocalID = targetSelector.Path, targetSelector.NativeID
		request.TypeURL, err = graphPreviewLinkType(cmd)
		if err != nil {
			return err
		}
		if request.TypeURL == "" {
			return graphFailure("invalid_selector", "pair unlink requires --type", 2)
		}
		if _, native := graphPreviewNativeDependencyType(graphPreviewConfig.GraphScopeURL, request.TypeURL); native {
			return graphFailure("capability_unavailable", "remove a Dependency by Link ID or use bd dep remove for a blocking pair", 5)
		}
	}
	var err error
	request.ExpectedRevision, request.Unconditional, err = graphPreviewRevisionGuard(cmd, false)
	if err != nil {
		return err
	}
	request.ExpectedSourceRevision, request.UnconditionalSource, err = graphPreviewRevisionGuard(cmd, true)
	if err != nil {
		return err
	}
	if !cmd.Flags().Changed("if-source-revision") {
		request.UnconditionalSource = false
		request.DefaultInformationalSource = true
	}
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		result, err := store.Unlink(ctx, request)
		return graphPreviewReplacementResult(result, fmt.Sprintf("Unlinked %s; identity remains reserved", result.Link.ID), result.ReplacedSource, err)
	})
}
