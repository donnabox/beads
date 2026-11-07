package main

import (
	"context"
	"fmt"
	"os"

	"github.com/steveyegge/beads/internal/remedy"
	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/internal/ui"
)

// showIssueRefs displays issues that reference the given issue(s), grouped by relationship type
func showIssueRefs(ctx context.Context, args []string, jsonOut bool) error {
	// Collect all refs for all issues
	allRefs := make(map[string][]*types.IssueWithDependencyMetadata)

	// Process each issue
	processIssue := func(issueID string, issueStore storage.DoltStorage) error {
		refs, err := issueStore.GetDependentsWithMetadata(ctx, issueID)
		if err != nil {
			return err
		}
		refs, err = hydrateRemoteDependents(ctx, issueStore, refs)
		if err != nil {
			return err
		}
		allRefs[issueID] = refs
		return nil
	}

	// Process each arg via routing-aware resolution
	for _, id := range args {
		result, err := resolveAndGetIssueWithRouting(ctx, store, id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving %s: %s\n", id, remedy.PrintedMessage(err))
			continue
		}
		if result == nil || result.Issue == nil {
			if result != nil {
				result.Close()
			}
			fmt.Fprintf(os.Stderr, "Issue %s not found\n", id)
			continue
		}
		if err := processIssue(result.ResolvedID, result.Store); err != nil {
			fmt.Fprintf(os.Stderr, "Error getting refs for %s: %v\n", id, err)
		}
		result.Close()
	}

	// Output results
	if jsonOut {
		return outputJSON(allRefs)
	}

	// Display refs grouped by issue and relationship type
	for issueID, refs := range allRefs {
		if len(refs) == 0 {
			fmt.Printf("\n%s: No references found\n", ui.RenderAccent(issueID))
			continue
		}

		fmt.Printf("\n%s References to %s:\n", ui.RenderAccent("📎"), issueID)

		// Every ref is an edge pointing AT this issue, so each group is named
		// from this issue's end. The bare type name would read from the other
		// end for the types whose name runs source-first: a (dup, canonical)
		// edge under a "duplicates" heading says the canonical is the copy.
		for _, sec := range groupDepSections(refs, false, nil) {
			displayRefGroup(sec)
		}
		fmt.Println()
	}
	return nil
}

// displayRefGroup displays one group of references under its relationship name
// Closed items get entire row muted - the work is done, no need for attention
func displayRefGroup(sec depSection) {
	emoji := getRefTypeEmoji(sec.Type)
	fmt.Printf("\n  %s %s (%d):\n", emoji, sec.Heading, len(sec.Deps))

	for _, ref := range sec.Deps {
		// Closed items: mute entire row since the work is complete
		if ref.Status == types.StatusClosed {
			fmt.Printf("    %s: %s %s\n",
				ui.RenderMuted(ref.ID),
				ui.RenderMuted(ref.Title),
				ui.RenderMuted(fmt.Sprintf("[P%d - %s]", ref.Priority, ref.Status)))
			continue
		}

		// Active items: color ID based on status
		var idStr string
		switch ref.Status {
		case types.StatusOpen:
			idStr = ui.StatusOpenStyle.Render(ref.ID)
		case types.StatusInProgress:
			idStr = ui.StatusInProgressStyle.Render(ref.ID)
		case types.StatusBlocked:
			idStr = ui.StatusBlockedStyle.Render(ref.ID)
		default:
			idStr = ref.ID
		}
		fmt.Printf("    %s: %s [P%d - %s]\n", idStr, ref.Title, ref.Priority, ref.Status)
	}
}

// getRefTypeEmoji returns an emoji for a dependency/reference type
func getRefTypeEmoji(depType types.DependencyType) string {
	switch depType {
	case types.DepUntil:
		return "⏳" // Hourglass - waiting until
	case types.DepCausedBy:
		return "⚡" // Lightning - triggered by
	case types.DepValidates:
		return "✅" // Checkmark - validates
	case types.DepBlocks:
		return "🚫" // Blocked
	case types.DepParentChild:
		return "↳" // Child arrow
	case types.DepRelatesTo, types.DepRelated:
		return "↔" // Bidirectional
	case types.DepTracks:
		return "👁" // Watching
	case types.DepDiscoveredFrom:
		return "◊" // Diamond - discovered
	case types.DepSupersedes:
		return "⬆" // Upgrade
	case types.DepDuplicates:
		return "🔄" // Duplicate
	case types.DepRepliesTo:
		return "💬" // Chat
	case types.DepApprovedBy:
		return "👍" // Approved
	case types.DepAuthoredBy:
		return "✏" // Authored
	case types.DepAssignedTo:
		return "👤" // Assigned
	default:
		return "→" // Default arrow
	}
}

// hydrateRemoteDependents replaces the shallow rows a remote store's dependents
// read answers with whole ones. bd serve projects that read onto id, status,
// type, priority, title and edge type, which is enough for `bd show`'s
// dependents section and not for a caller rendering the row whole: every other
// field would print blank and every timestamp as 0001-01-01. A row the follow-up
// read cannot supply refuses rather than printing that.
func hydrateRemoteDependents(ctx context.Context, s storage.DoltStorage, refs []*types.IssueWithDependencyMetadata) ([]*types.IssueWithDependencyMetadata, error) {
	if !storeIsRemoteBackend(s) || len(refs) == 0 {
		return refs, nil
	}
	ids := make([]string, 0, len(refs))
	for _, ref := range refs {
		ids = append(ids, ref.ID)
	}
	rows, err := s.GetIssuesByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	byID := make(map[string]*types.Issue, len(rows))
	for _, row := range rows {
		byID[row.ID] = row
	}
	hydrated := make([]*types.IssueWithDependencyMetadata, 0, len(refs))
	for _, ref := range refs {
		row, ok := byID[ref.ID]
		if !ok {
			return nil, fmt.Errorf("the server listed %q as a dependent but did not return its row", ref.ID)
		}
		hydrated = append(hydrated, &types.IssueWithDependencyMetadata{Issue: *row, DependencyType: ref.DependencyType})
	}
	return hydrated, nil
}
