package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// The ordinary single-Bead route remains the default. These forms select
// native Issue set semantics, including the bounded cascade closure.
func graphPreviewDeleteSetRequested(cmd *cobra.Command, args []string) bool {
	return len(args) > 1 || cmd.Flags().Changed("cascade") || cmd.Flags().Changed("from-file")
}

func runGraphPreviewDeleteIssueSet(cmd *cobra.Command, args []string) error {
	if err := graphPreviewFlags(cmd, "force", "if-revision", "cascade", "from-file", "dry-run"); err != nil {
		return err
	}
	fromFile, _ := cmd.Flags().GetString("from-file")
	if cmd.Flags().Changed("from-file") {
		if fromFile == "" {
			return graphFailure("invalid_selector", "--from-file requires a path", 2)
		}
		fileIDs, err := readIssueIDsFromFile(fromFile)
		if err != nil {
			return graphFailure("invalid_selector", "reading --from-file: "+err.Error(), 2)
		}
		args = append(args, fileIDs...)
	}
	if len(args) == 0 {
		return graphFailure("invalid_selector", "delete requires at least one Issue ID or beads/PATH", 2)
	}
	selectors := make([]graphstore.IssueSelector, 0, len(args))
	seen := make(map[graphstore.IssueSelector]bool, len(args))
	for _, input := range args {
		selector, err := graphPreviewWritableIdentity(input)
		if err != nil {
			return graphFailure("invalid_selector", err.Error(), 2)
		}
		selected, err := graphPreviewIssueSelector(selector)
		if err != nil {
			return graphFailure("invalid_selector", err.Error(), 2)
		}
		if !seen[selected] {
			seen[selected] = true
			selectors = append(selectors, selected)
		}
	}
	if len(selectors) > graphstore.PreviewSnapshotLimit {
		return graphFailure("limit_exceeded", fmt.Sprintf("delete exceeds %d Issue selectors", graphstore.PreviewSnapshotLimit), 2)
	}
	revision, unconditional, err := graphPreviewRevisionGuard(cmd, false)
	if err != nil {
		return err
	}
	if revision != "" && len(selectors) != 1 {
		return graphFailure("invalid_selector", "--if-revision requires exactly one Issue", 2)
	}
	force, _ := cmd.Flags().GetBool("force")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	cascade, _ := cmd.Flags().GetBool("cascade")
	preview := dryRun || !force
	if !preview {
		if err := graphPreviewWritePolicy(); err != nil {
			return err
		}
	}
	expected := make(map[string]string)
	if revision != "" {
		expected[selectors[0].Path] = revision
	}
	request := graphstore.IssueDeleteSetRequest{
		Selectors: selectors, ExpectedRevisions: expected, Unconditional: unconditional,
		Force: force, Cascade: cascade, Preview: preview, Actor: getActorWithGit(),
	}
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		result, err := store.DeleteIssueSetWithIncidentLinks(ctx, request)
		if err != nil {
			return nil, "", err
		}
		verb := "Deleted"
		if result.Preview {
			verb = "Would delete"
		}
		names := make([]string, len(result.Issues))
		for i, issue := range result.Issues {
			names[i] = issue.Issue.ID
		}
		summary := fmt.Sprintf("%s %d Issue(s): %s; %d incident Link(s); identity and prior snapshots remain retained",
			verb, len(names), strings.Join(names, ", "), len(result.IncidentLinkIDs))
		return result, summary, nil
	})
}
