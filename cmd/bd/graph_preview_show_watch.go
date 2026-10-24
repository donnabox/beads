package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/internal/workapi"
)

// Keep the shipping watch loop's redraw and signal behavior. Each graph read
// opens a short-lived store so a long watch never holds one transaction and a
// separate CLI process can publish the next revision.
func runGraphPreviewShowWatch(cmd *cobra.Command, args []string, format string, commentsTail int) error {
	if len(args) != 1 {
		return graphFailure("invalid_selector", "watch mode requires exactly one Issue ID", 2)
	}
	if format != "" || cmd.Flags().Changed("revision") {
		return graphFailure("capability_unavailable", "--watch requires a current Issue in the familiar output format", 5)
	}
	resolve := graphPreviewResolveMemoryArgument
	if cmd.Flags().Changed("id") {
		resolve = graphPreviewResolveLiteralArgument
	}
	argument, err := resolve(args[0])
	if err != nil {
		return err
	}
	ref, err := graphPreviewParseReference(argument)
	if err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	if ref.Explicit && !ref.currentSelection() {
		return graphFailure("capability_unavailable", "--watch requires a current Issue", 5)
	}
	selectorPath, localID, err := graphPreviewResourceReadSelector(ref.Identity)
	if err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	// Refuse a Memory or missing Issue before starting a watch. The shared
	// loop's initial-render failure is otherwise silent by design.
	if err := withGraphStoreOutput(func(ctx context.Context, s *graphstore.Store) (any, string, error) {
		_, record, _, err := s.ReadResourceWithIssueDetailsSelected(ctx, selectorPath, localID, workapi.DetailOptions{})
		if err != nil {
			return nil, "", err
		}
		if _, ok := record.(graphstore.IssueRecord); !ok {
			return nil, "", graphFailure("capability_unavailable", "--watch requires an Issue", 5)
		}
		return nil, "", nil
	}, func(any, string) error { return nil }); err != nil {
		return err
	}
	var revision string
	return runIssueWatch(getRootContext(), issueWatchSource{
		// A native timestamp can be rounded by the server backend. The graph
		// token changes for every authored Issue or owned-Link revision.
		snapshot: func(*types.Issue) string { return revision },
		render: func(ctx context.Context) *types.Issue {
			var issue *types.Issue
			err := withGraphStoreOutput(func(ctx context.Context, s *graphstore.Store) (any, string, error) {
				_, record, detail, err := s.ReadResourceWithIssueDetailsSelected(ctx, selectorPath, localID, workapi.DetailOptions{IncludeDependents: true, IncludeComments: true})
				if err != nil {
					return nil, "", err
				}
				if _, ok := record.(graphstore.IssueRecord); !ok {
					return nil, "", graphFailure("capability_unavailable", "--watch requires an Issue", 5)
				}
				return detail, graphHumanDocument(s.ScopeURL(), record, "", false, false) + graphHumanCurrentIssueDetails(s.ScopeURL(), detail, graphIssueDetailOptions{CommentsTail: commentsTail}), nil
			}, func(result any, human string) error {
				_, err := fmt.Fprintln(cmd.OutOrStdout(), human)
				if err == nil {
					detail := result.(*types.IssueDetails)
					issue = &detail.Issue
					revision = detail.Revision
				}
				return err
			})
			if err != nil {
				return nil
			}
			return issue
		},
		fetch: func(ctx context.Context) (*types.Issue, error) {
			readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()
			s, err := graphstore.OpenExisting(readCtx, graphOptions(graphPreviewConfig))
			if err != nil {
				return nil, err
			}
			_, record, _, readErr := s.ReadResourceWithIssueDetailsSelected(readCtx, selectorPath, localID, workapi.DetailOptions{})
			closeErr := s.Close()
			if err := errors.Join(readErr, closeErr); err != nil {
				return nil, err
			}
			issue, ok := record.(graphstore.IssueRecord)
			if !ok {
				return nil, fmt.Errorf("%w: watched Bead is no longer an Issue", graphstore.ErrCapabilityUnavailable)
			}
			selected := *issue.Properties
			revision = issue.Revision
			return &selected, nil
		},
	})
}
