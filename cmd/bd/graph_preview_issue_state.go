package main

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func runGraphPreviewState(cmd *cobra.Command, selector, dimension string, list bool) error {
	if err := graphPreviewFlags(cmd); err != nil {
		return err
	}
	selected, err := graphPreviewIssueSelector(selector)
	if err != nil {
		return graphFailure("invalid_selector", "state requires a local Issue ID or beads/PATH", 2)
	}
	return withGraphFamiliarStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		issue, err := store.ShowIssueSelected(ctx, selected)
		if err != nil {
			return nil, "", err
		}
		return graphStateReport(issue, dimension, list)
	})
}

func graphStateReport(issue graphstore.IssueRecord, dimension string, list bool) (any, string, error) {
	if list {
		states := collectStateLabels(issue.Properties.Labels)
		names := make([]string, 0, len(states))
		for name := range states {
			names = append(names, name)
		}
		sort.Strings(names)
		lines := make([]string, 0, len(names))
		for _, name := range names {
			lines = append(lines, fmt.Sprintf("%s: %s", name, states[name]))
		}
		human := fmt.Sprintf("%s has no state labels", issue.Properties.ID)
		if len(lines) > 0 {
			human = strings.Join(lines, "\n")
		}
		return map[string]any{"issue_id": issue.Properties.ID, "states": states}, human, nil
	}
	value, _ := findStateLabel(issue.Properties.Labels, dimension)
	var outputValue any
	human := fmt.Sprintf("(no %s state set)", dimension)
	if value != "" {
		outputValue = value
		human = value
	}
	return map[string]any{"issue_id": issue.Properties.ID, "dimension": dimension, "value": outputValue}, human, nil
}
