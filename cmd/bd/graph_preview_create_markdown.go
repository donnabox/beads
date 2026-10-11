package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// Markdown headings are an ordinary Issue batch, not independent graph
// creates. In particular, any refusal must roll back the entire file.
func runGraphPreviewCreateMarkdown(cmd *cobra.Command, args []string) error {
	if err := graphPreviewFlags(cmd, "file", "validate", "silent"); err != nil {
		return err
	}
	in, err := gatherCreateInput(cmd, args)
	if err != nil {
		return err
	}
	templates, err := parseMarkdownFile(in.markdownFile)
	if err != nil {
		return graphFailure("invalid_properties", fmt.Sprintf("parsing markdown file: %v", err), 2)
	}
	request, err := buildMarkdownBatchRequest(templates, in)
	if err != nil {
		return err
	}
	var last string
	err = withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		records, err := store.CreateIssueBatch(ctx, request)
		if err != nil {
			return nil, "", err
		}
		var summary strings.Builder
		fmt.Fprintf(&summary, "Created %d issues from %s:\n", len(records), in.markdownFile)
		for _, record := range records {
			id := graphPreviewDisplayLocalURL(record.ID)
			fmt.Fprintf(&summary, "  %s: %s [P%d, %s]\n", id, record.Properties.Title, record.Properties.Priority, record.Properties.IssueType)
			last = id
		}
		return records, summary.String(), nil
	})
	if err == nil && last != "" {
		SetLastTouchedID(last)
	}
	return err
}
