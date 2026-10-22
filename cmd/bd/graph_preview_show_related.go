package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/internal/ui"
	"github.com/steveyegge/beads/internal/workapi"
)

// The familiar reverse-reference views use the graph detail reader's checked
// native dependency projection. Their JSON shape remains the shipping map of
// Issue IDs to related Issue rows, including an empty row set for an Issue
// without references.
func runGraphPreviewShowRelated(cmd *cobra.Command, args []string, refs, short bool, format string) error {
	if len(args) == 0 {
		return graphFailure("invalid_selector", "at least one Issue ID is required", 2)
	}
	if format != "" || cmd.Flags().Changed("version") {
		return graphFailure("capability_unavailable", "--refs and --children require current Issues in the familiar output format", 5)
	}
	paths := make([]string, 0, len(args))
	for _, arg := range args {
		ref, err := graphPreviewParseReference(arg)
		if err != nil {
			return graphFailure("invalid_selector", err.Error(), 2)
		}
		if ref.Explicit && ref.Revision != "current" && ref.Revision != "unpinned" {
			return graphFailure("capability_unavailable", "--refs and --children require current Issues", 5)
		}
		path, err := graphPreviewResourcePath(graphPreviewConfig.GraphScopeURL, ref.Identity)
		if err != nil {
			return graphFailure("invalid_selector", err.Error(), 2)
		}
		if !strings.HasPrefix(path, "beads/") {
			return graphFailure("capability_unavailable", "--refs and --children require Issues", 5)
		}
		paths = append(paths, path)
	}
	var missing []string
	err := withGraphStoreOutput(func(ctx context.Context, s *graphstore.Store) (any, string, error) {
		rows := make(map[string][]*types.IssueWithDependencyMetadata, len(paths))
		var human strings.Builder
		for index, path := range paths {
			record, err := s.Read(ctx, path)
			if errors.Is(err, graphstore.ErrNotFound) || errors.Is(err, graphstore.ErrGone) {
				missing = append(missing, args[index])
				continue
			}
			if err != nil {
				return nil, "", err
			}
			if _, ok := record.(graphstore.IssueRecord); !ok {
				return nil, "", graphFailure("capability_unavailable", "--refs and --children require Issues", 5)
			}
			detail, err := s.ShowIssueDetails(ctx, path, workapi.DetailOptions{IncludeDependents: true})
			if err != nil {
				return nil, "", err
			}
			related := make([]*types.IssueWithDependencyMetadata, 0, len(detail.Dependents))
			for _, row := range detail.Dependents {
				if !refs && row.DependencyType != types.DepParentChild {
					continue
				}
				related = append(related, row)
			}
			rows[path] = related
			if refs {
				renderGraphShowRefs(&human, path, related)
			} else {
				renderGraphShowChildren(&human, path, related, short)
			}
		}
		return rows, human.String(), nil
	}, func(result any, human string) error {
		for _, arg := range missing {
			fmt.Fprintf(cmd.ErrOrStderr(), "Issue %s not found\n", arg)
		}
		if jsonOutput {
			return outputJSON(result)
		}
		if !quietFlag {
			_, err := fmt.Fprint(cmd.OutOrStdout(), human)
			return err
		}
		return nil
	})
	return err
}

func renderGraphShowChildren(out *strings.Builder, path string, children []*types.IssueWithDependencyMetadata, short bool) {
	if len(children) == 0 {
		fmt.Fprintf(out, "%s: No children found\n", ui.RenderAccent(path))
		return
	}
	fmt.Fprintf(out, "%s Children of %s (%d):\n", ui.RenderAccent("↳"), path, len(children))
	for _, child := range children {
		if short {
			fmt.Fprintf(out, "  %s\n", formatShortIssue(&child.Issue))
		} else {
			fmt.Fprintln(out, formatDependencyLine("↳", child))
		}
	}
	fmt.Fprintln(out)
}

func renderGraphShowRefs(out *strings.Builder, path string, refs []*types.IssueWithDependencyMetadata) {
	if len(refs) == 0 {
		fmt.Fprintf(out, "\n%s: No references found\n", ui.RenderAccent(path))
		return
	}
	fmt.Fprintf(out, "\n%s References to %s:\n", ui.RenderAccent("📎"), path)
	for _, section := range groupDepSections(refs, false, nil) {
		writeRefGroup(out, section)
	}
	fmt.Fprintln(out)
}
