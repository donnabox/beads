package main

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/ui"
	"github.com/steveyegge/beads/internal/workapi"
)

// Label listing is an Issue detail read. Resolve the native selector and
// project its labels in the same checked graph snapshot as other Issue views.
func runGraphPreviewLabelList(cmd *cobra.Command, raw string) error {
	if err := graphPreviewFlags(cmd); err != nil {
		return err
	}
	selector, err := graphPreviewIssueSelector(raw)
	if err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	return withGraphStoreOutput(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		details, err := store.ShowIssueDetailsSelected(ctx, selector, workapi.DetailOptions{})
		if err != nil {
			return nil, "", err
		}
		labels := append([]string{}, details.Labels...)
		sort.Strings(labels)
		id := graphPreviewDisplayLocalURL(details.ID)
		if len(labels) == 0 {
			return labels, fmt.Sprintf("\n%s has no labels\n", id), nil
		}
		var human strings.Builder
		fmt.Fprintf(&human, "\n%s Labels for %s:\n", ui.RenderAccent("🏷"), id)
		for _, label := range labels {
			fmt.Fprintf(&human, "  - %s\n", label)
		}
		human.WriteByte('\n')
		return labels, human.String(), nil
	}, func(result any, human string) error {
		if jsonOutput {
			return outputJSON(result)
		}
		if quietFlag {
			return nil
		}
		_, err := io.WriteString(cmd.OutOrStdout(), human)
		return err
	})
}
