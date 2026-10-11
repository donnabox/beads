package main

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func runGraphPreviewState(cmd *cobra.Command, selector, dimension string, list bool) error {
	if err := graphPreviewFlags(cmd); err != nil {
		return err
	}
	path, err := graphPreviewResourcePath(graphPreviewConfig.GraphScopeURL, selector)
	if err != nil || graph.ValidateBeadPath(path) != nil {
		return graphFailure("invalid_selector", "state requires a local Issue ID or beads/PATH", 2)
	}
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		resource, err := store.Read(ctx, path)
		if err != nil {
			return nil, "", err
		}
		issue, ok := resource.(graphstore.IssueRecord)
		if !ok {
			return nil, "", graphstore.ErrCapabilityUnavailable
		}
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
			human := fmt.Sprintf("%s has no state labels", issue.ID)
			if len(lines) > 0 {
				human = strings.Join(lines, "\n")
			}
			return map[string]any{"issue_id": issue.ID, "states": states}, human, nil
		}
		value, _ := findStateLabel(issue.Properties.Labels, dimension)
		var outputValue any
		human := fmt.Sprintf("(no %s state set)", dimension)
		if value != "" {
			outputValue = value
			human = value
		}
		return map[string]any{"issue_id": issue.ID, "dimension": dimension, "value": outputValue}, human, nil
	})
}
