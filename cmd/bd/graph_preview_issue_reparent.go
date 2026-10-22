package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// A parent replacement and admitted scalar edits share one checked native
// Issue writer and canonical Link transaction.
func runGraphPreviewReparentIssue(cmd *cobra.Command, sourceArg string) error {
	edit, err := graphPreviewIssueEditRequestWithParent(cmd, "", true)
	if err != nil {
		return err
	}
	if err := graphPreviewIssuePropertyMerge(cmd, &edit); err != nil {
		return err
	}
	parentArg, _ := cmd.Flags().GetString("parent")
	if parentArg == "" {
		return graphFailure("capability_unavailable", "graph update --parent cannot detach an Issue while dotted IDs imply a parent; select a nonempty local Issue", 5)
	}
	source, err := graphPreviewIssueSelector(sourceArg)
	if err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	parent, err := graphPreviewIssueSelector(parentArg)
	if err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	revision, unconditional := edit.ExpectedRevision, edit.Unconditional
	edit.ExpectedRevision, edit.Unconditional = "", false
	var touchedPath string
	err = withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		result, err := store.ReparentIssue(ctx, graphstore.IssueReparentRequest{
			Path: source.Path, NativeID: source.NativeID, ParentPath: parent.Path, ParentNativeID: parent.NativeID,
			Actor: getActorWithGit(), ExpectedRevision: revision, Unconditional: unconditional, Edit: edit,
		})
		if err != nil {
			return nil, "", err
		}
		verb := "Unchanged"
		if result.Changed {
			verb = "Updated"
		}
		touchedPath = strings.TrimPrefix(result.Issue.ID, graphPreviewConfig.GraphScopeURL)
		// Keep the ordinary Issue-update feedback shape; the Link identities
		// remain available through bd link rather than changing update JSON.
		return graphstore.IssueMutationResult{Issue: result.Issue, Changed: result.Changed},
			fmt.Sprintf("%s %s", verb, result.Issue.ID), nil
	})
	if err == nil {
		SetLastTouchedID(touchedPath)
	}
	return err
}
