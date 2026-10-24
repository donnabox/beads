package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

var graphGenericFlags = []string{"view", "direction", "depth", "max-nodes", "max-links", "link-type", "exact-link-type"}

func init() {
	graphCmd.Flags().String("view", "", "Deprecated: generic is the ordinary graph view in link mode")
	graphCmd.Flags().String("direction", "both", "Generic Link direction: in, out or both (graph_mode link workspaces only)")
	graphCmd.Flags().Int("depth", 1000, "Link hop depth (0..1000); frontier reports depth truncation (link mode only)")
	graphCmd.Flags().StringArray("link-type", nil, "Select a Link Type including refinements (repeatable OR union; default Dependency; link mode only)")
	graphCmd.Flags().StringArray("exact-link-type", nil, "Select exactly a Link Type, without refinements (repeatable OR union; link mode only)")
	graphCmd.Flags().Int("max-nodes", 1000, "Generic emitted Bead bound (1..1000); exhaustion refuses (graph_mode link workspaces only)")
	graphCmd.Flags().Int("max-links", 1000, "Generic emitted Link bound (1..1000); exhaustion refuses (graph_mode link workspaces only)")
}

func graphGenericFlagsChanged(cmd *cobra.Command) bool {
	for _, name := range graphGenericFlags {
		if cmd.Flags().Changed(name) {
			return true
		}
	}
	return false
}

func graphPreviewGenericInput(cmd *cobra.Command, args []string, scope string) (graphGenericInput, error) {
	zero := graphGenericInput{}
	if err := graphPreviewFlags(cmd, append(append([]string{}, graphGenericFlags...), "all", "compact", "box", "dot", "html", "open", "max-rows")...); err != nil {
		return zero, err
	}
	open, _ := cmd.Flags().GetBool("open")
	if open {
		return zero, graphFailure("capability_unavailable", "--open is not implemented for the generic graph view; omit it (dependency mode retains its existing open view)", 5)
	}
	view, _ := cmd.Flags().GetString("view")
	if cmd.Flags().Changed("view") && view != "generic" {
		return zero, graphFailure("invalid_selector", "--view only accepts generic in link mode", 2)
	}
	all, _ := cmd.Flags().GetBool("all")
	if (all && len(args) != 0) || (!all && len(args) != 1) {
		return zero, graphFailure("invalid_selector", "graph requires one local Bead ID or --all without an ID", 2)
	}
	if all && (cmd.Flags().Changed("direction") || cmd.Flags().Changed("depth")) {
		return zero, graphFailure("invalid_selector", "--all selects the full matching population; omit --direction and --depth", 2)
	}
	in := graphGenericInput{All: all}
	if !all {
		path, err := graphPreviewResourcePath(scope, args[0])
		if err != nil {
			return zero, graphFailure("invalid_selector", err.Error(), 2)
		}
		if err := graph.ValidateBeadPath(path); err != nil {
			return zero, graphFailure("invalid_selector", err.Error(), 2)
		}
		in.Root = graph.CanonicalURL(scope, path)
	}
	selectors, _ := cmd.Flags().GetStringArray("link-type")
	exact, _ := cmd.Flags().GetStringArray("exact-link-type")
	for _, selection := range []struct {
		flag   string
		values []string
	}{{"link-type", selectors}, {"exact-link-type", exact}} {
		if cmd.Flags().Changed(selection.flag) && len(selection.values) == 0 {
			return zero, graphFailure("invalid_selector", "--"+selection.flag+" requires a nonempty Type selector", 2)
		}
	}
	in.ImplicitDependency = !cmd.Flags().Changed("link-type") && !cmd.Flags().Changed("exact-link-type")
	if in.ImplicitDependency {
		selectors = []string{"dependency"}
	}
	resolve := func(selector, flag string) (string, error) {
		if selector == "" {
			return "", graphFailure("invalid_selector", "--"+flag+" requires a nonempty Type selector", 2)
		}
		if !strings.Contains(selector, "/") {
			for _, name := range graphstore.FreshTypeNames() {
				if strings.EqualFold(selector, name) {
					selector = name
					break
				}
			}
		}
		url, err := graphPreviewTypeURL(scope, selector)
		if err != nil {
			return "", graphFailure("invalid_selector", err.Error(), 2)
		}
		return url, nil
	}
	for _, selection := range []struct {
		flag        string
		values      []string
		destination *[]string
	}{{"link-type", selectors, &in.LinkTypes}, {"exact-link-type", exact, &in.ExactLinkTypes}} {
		seen := map[string]bool{}
		for _, selector := range selection.values {
			if selection.flag == "link-type" && selector == "any" {
				in.AllLinkTypes = true
				continue
			}
			url, err := resolve(selector, selection.flag)
			if err != nil {
				return zero, err
			}
			if !seen[url] {
				*selection.destination = append(*selection.destination, url)
				seen[url] = true
			}
		}
		sort.Strings(*selection.destination)
	}
	in.Direction, _ = cmd.Flags().GetString("direction")
	in.Depth, _ = cmd.Flags().GetInt("depth")
	in.MaxNodes, _ = cmd.Flags().GetInt("max-nodes")
	in.MaxLinks, _ = cmd.Flags().GetInt("max-links")
	if err := validateGraphGenericBounds(in); err != nil {
		return zero, graphFailure("invalid_selector", err.Error(), 2)
	}
	if err := graphPreviewRejectDependencyRowCap(cmd, "current graph acquisition"); err != nil {
		return zero, err
	}
	return in, nil
}

func runGraphPreviewGeneric(cmd *cobra.Command, args []string) error {
	in, err := graphPreviewGenericInput(cmd, args, graphPreviewConfig.GraphScopeURL)
	if err != nil {
		return err
	}
	return withGraphStoreOutput(func(ctx context.Context, s *graphstore.Store) (any, string, error) {
		snapshot, err := s.CurrentSnapshot(ctx)
		if err != nil {
			return nil, "", err
		}
		// Dependency --all follows the familiar open-Issue population. Other
		// selectors include only matching components, never unrelated isolates.
		in.OpenIssuesOnly = in.All && !in.AllLinkTypes && len(in.ExactLinkTypes) == 0 && len(in.LinkTypes) == 1 && in.LinkTypes[0] == graphstore.DependencyTypeURL(s.ScopeURL())
		result, err := projectGraphGeneric(snapshot, s.ScopeURL(), in)
		if err != nil {
			return nil, "", err
		}

		output, err := renderGraphSelected(cmd, result, snapshot)
		return result.OmittedLinkTypes, output, err
	}, func(value any, output string) error {
		omitted, _ := value.([]string)
		return writeGraphGenericSelection(cmd.OutOrStdout(), cmd.ErrOrStderr(), output, omitted)
	})
}

func writeGraphGenericSelection(out, diagnostics io.Writer, output string, omitted []string) error {
	if err := writeGraphGeneric(out, output); err != nil {
		return err
	}
	return writeGraphGenericOmittedNotice(diagnostics, omitted)
}

func writeGraphGenericOmittedNotice(out io.Writer, omitted []string) error {
	if len(omitted) == 0 {
		return nil
	}
	return writeGraphGeneric(out, fmt.Sprintf("Graph uses the default Dependency family; reached Beads also have omitted Link Types: %s. Use --link-type any to select all Link Types; direction and depth still apply.\n", strings.Join(omitted, ", ")))
}

func writeGraphGeneric(out io.Writer, output string) error {
	n, err := io.WriteString(out, output)
	if err == nil && n != len(output) {
		return io.ErrShortWrite
	}
	return err
}

func renderGraphGeneric(result graphGenericResult, structured, quiet bool) (string, error) {
	var human strings.Builder
	if !structured && !quiet {
		fmt.Fprintf(&human, "Generic graph (summary; graph preview)\nRoot %q direction %s depth %d\n", result.Root, result.Direction, result.Depth)
		for _, node := range result.Nodes {
			fmt.Fprintf(&human, "Bead %q %q\n", node.ID, node.Title)
		}
		for _, link := range result.Links {
			fmt.Fprintf(&human, "Link %q %q -> %q\n", link.ID, link.Source, link.Target)
		}
		fmt.Fprintf(&human, "Complete: %t; frontier: %d\n", result.Complete, len(result.Frontier))
		for _, id := range result.Frontier {
			fmt.Fprintf(&human, "Frontier %q\n", id)
		}
	}
	var output bytes.Buffer
	if err := graphPrintTo(&output, result, strings.TrimSuffix(human.String(), "\n"), quiet, structured); err != nil {
		return "", err
	}
	if output.Len() > graphGenericOutputLimit {
		return "", fmt.Errorf("%w: generic summary output exceeds %d bytes", graphstore.ErrLimitExceeded, graphGenericOutputLimit)
	}
	return output.String(), nil
}
