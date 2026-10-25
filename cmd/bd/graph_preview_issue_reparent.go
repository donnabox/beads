package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// Parent replacement and admitted scalar edits share one checked native Issue
// writer and canonical Link transaction.
func runGraphPreviewReparentIssue(cmd *cobra.Command, sourceArg string) error {
	parentArg, _ := cmd.Flags().GetString("parent")
	if parentArg == "" {
		return graphFailure("capability_unavailable", "graph update --parent cannot detach an Issue while dotted IDs imply a parent; select a nonempty local Issue", 5)
	}
	if _, err := graphPreviewWritableIdentity(sourceArg); err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	if _, err := graphPreviewWritableIdentity(parentArg); err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	source, err := graphPreviewIssueSelector(sourceArg)
	if err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	parent, err := graphPreviewIssueSelector(parentArg)
	if err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	// Refuse malformed operands before edit flags can consume body stdin.
	edit, err := graphPreviewIssueEditRequestWithParent(cmd, "", true, sourceArg)
	if err != nil {
		return err
	}
	if err := graphPreviewIssuePropertyMerge(cmd, &edit); err != nil {
		return err
	}
	revision, unconditional := edit.ExpectedRevision, edit.Unconditional
	edit.ExpectedRevision, edit.Unconditional = "", false
	var touchedID string
	err = withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		result, err := store.ReparentIssue(ctx, graphstore.IssueReparentRequest{
			Path: source.Path, NativeID: source.NativeID, ParentPath: parent.Path, ParentNativeID: parent.NativeID,
			Exact: source.Exact, ParentExact: parent.Exact,
			Actor: getActorWithGit(), ExpectedRevision: revision, Unconditional: unconditional, Edit: edit,
		})
		if err != nil {
			return nil, "", err
		}
		verb := "Unchanged"
		if result.Changed {
			verb = "Updated"
		}
		touchedID = result.Issue.ID
		// Keep the ordinary Issue-update feedback shape; the Link identities
		// remain available through bd link rather than changing update JSON.
		return graphstore.IssueMutationResult{Issue: result.Issue, Changed: result.Changed},
			fmt.Sprintf("%s %s", verb, result.Issue.ID), nil
	})
	if err == nil {
		SetLastTouchedID(touchedID)
	}
	return err
}
