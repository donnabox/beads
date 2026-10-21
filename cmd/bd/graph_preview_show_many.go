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

// The shipping multi-ID form is an Issue detail command. Keep the graph
// Resource and exact-version presentations on the single-selector route.
func runGraphPreviewShowMany(cmd *cobra.Command, args []string, shortMode, longMode, localTime bool, commentsTail int, format string) error {
	if format != "" || cmd.Flags().Changed("version") {
		return graphFailure("capability_unavailable", "multi-ID show supports current Issue details only", 5)
	}
	paths := make([]string, 0, len(args))
	for _, arg := range args {
		ref, err := graphPreviewParseReference(arg)
		if err != nil {
			return graphFailure("invalid_selector", err.Error(), 2)
		}
		if ref.Explicit && ref.Revision != "current" && ref.Revision != "unpinned" {
			return graphFailure("capability_unavailable", "multi-ID show supports current Issue details only", 5)
		}
		path, err := graphPreviewResourcePath(graphPreviewConfig.GraphScopeURL, ref.Identity)
		if err != nil {
			return graphFailure("invalid_selector", err.Error(), 2)
		}
		if !strings.HasPrefix(path, "beads/") {
			return graphFailure("capability_unavailable", "multi-ID show supports current Issues only", 5)
		}
		paths = append(paths, path)
	}
	includeDependents, _ := cmd.Flags().GetBool("include-dependents")
	includeComments, _ := cmd.Flags().GetBool("include-comments")
	briefDeps, _ := cmd.Flags().GetBool("brief-deps")
	var missing []string
	err := withGraphStoreOutput(func(ctx context.Context, s *graphstore.Store) (any, string, error) {
		details := make([]*types.IssueDetails, 0, len(paths))
		lines := make([]string, 0, len(paths))
		for index, path := range paths {
			record, err := s.Read(ctx, path)
			if errors.Is(err, graphstore.ErrNotFound) || errors.Is(err, graphstore.ErrGone) {
				missing = append(missing, args[index])
				continue
			}
			if err != nil {
				return nil, "", err
			}
			issue, ok := record.(graphstore.IssueRecord)
			if !ok {
				return nil, "", graphFailure("capability_unavailable", "multi-ID show supports current Issues only", 5)
			}
			if shortMode {
				shorthand := *issue.Properties
				shorthand.ID = path
				lines = append(lines, formatShortIssue(&shorthand))
				continue
			}
			opts := workapi.DetailOptions{IncludeDependents: includeDependents, IncludeComments: includeComments, BriefDeps: briefDeps}
			if !jsonOutput {
				opts = workapi.DetailOptions{IncludeDependents: true, IncludeComments: true}
			}
			detail, err := s.ShowIssueDetails(ctx, path, opts)
			if errors.Is(err, graphstore.ErrNotFound) || errors.Is(err, graphstore.ErrGone) {
				missing = append(missing, args[index])
				continue
			}
			if err != nil {
				return nil, "", err
			}
			if jsonOutput {
				details = append(details, detail)
			} else {
				lines = append(lines, renderGraphIssueDetails(detail, graphIssueDetailOptions{Long: longMode, LocalTime: localTime, CommentsTail: commentsTail}))
			}
		}
		if len(details)+len(lines) == 0 {
			return nil, "", fmt.Errorf("%w: no issues found matching the provided IDs", graphstore.ErrNotFound)
		}
		if shortMode {
			return nil, strings.Join(lines, "\n"), nil
		}
		return details, strings.Join(lines, "\n\n"+ui.RenderMuted(strings.Repeat("─", 60))+"\n\n"), nil
	}, func(result any, human string) error {
		for _, path := range missing {
			fmt.Fprintf(cmd.ErrOrStderr(), "Issue %s not found\nHint: %s\n", path, showNotFoundHint(path))
		}
		if shortMode {
			if !quietFlag {
				_, err := fmt.Fprintln(cmd.OutOrStdout(), human)
				return err
			}
			return nil
		}
		if jsonOutput {
			return outputJSON(result)
		}
		if !quietFlag {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), human)
			return err
		}
		return nil
	})
	if err == nil {
		SetLastTouchedID(paths[0])
	}
	return err
}
