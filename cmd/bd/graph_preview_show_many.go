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
	if format != "" || cmd.Flags().Changed("revision") {
		return graphFailure("capability_unavailable", "multi-ID show supports current Issue details only", 5)
	}
	selectors := make([]struct{ path, localID string }, 0, len(args))
	literalIDs, _ := cmd.Flags().GetStringArray("id")
	for index, arg := range args {
		resolve := graphPreviewResolveMemoryArgument
		if index >= len(args)-len(literalIDs) {
			resolve = graphPreviewResolveLiteralArgument
		}
		argument, err := resolve(arg)
		if err != nil {
			return err
		}
		ref, err := graphPreviewParseReference(argument)
		if err != nil {
			return graphFailure("invalid_selector", err.Error(), 2)
		}
		if ref.Explicit && !ref.currentSelection() {
			return graphFailure("capability_unavailable", "multi-ID show supports current Issue details only", 5)
		}
		path, localID, err := graphPreviewResourceReadSelector(ref.Identity)
		if err != nil {
			return graphFailure("invalid_selector", err.Error(), 2)
		}
		selectors = append(selectors, struct{ path, localID string }{path, localID})
	}
	includeDependents, _ := cmd.Flags().GetBool("include-dependents")
	includeComments, _ := cmd.Flags().GetBool("include-comments")
	briefDeps, _ := cmd.Flags().GetBool("brief-deps")
	var missing []string
	var firstSelected string
	err := withGraphStoreOutput(func(ctx context.Context, s *graphstore.Store) (any, string, error) {
		details := make([]*types.IssueDetails, 0, len(selectors))
		lines := make([]string, 0, len(selectors))
		for index, selector := range selectors {
			opts := workapi.DetailOptions{IncludeDependents: includeDependents, IncludeComments: includeComments, BriefDeps: briefDeps}
			if !jsonOutput {
				opts = workapi.DetailOptions{IncludeDependents: true, IncludeComments: true}
			}
			_, record, detail, err := s.ReadResourceWithIssueDetailsSelected(ctx, selector.path, selector.localID, opts)
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
			if firstSelected == "" {
				firstSelected = issue.ID
			}
			if shortMode {
				shorthand := *issue.Properties
				lines = append(lines, formatShortIssue(&shorthand))
				continue
			}
			if jsonOutput {
				details = append(details, detail)
			} else {
				lines = append(lines, graphHumanDocument(s.ScopeURL(), issue, "", false, longMode)+graphHumanCurrentIssueDetails(s.ScopeURL(), detail, graphIssueDetailOptions{Long: longMode, LocalTime: localTime, CommentsTail: commentsTail}))
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
	if err == nil && firstSelected != "" {
		SetLastTouchedID(firstSelected)
	}
	return err
}
