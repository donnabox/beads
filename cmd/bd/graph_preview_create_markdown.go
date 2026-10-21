package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/config"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/internal/validation"
)

// Markdown headings are an ordinary Issue batch, not independent graph
// creates. In particular, any refusal must roll back the entire file.
func runGraphPreviewCreateMarkdown(cmd *cobra.Command, args []string) error {
	if err := graphPreviewFlags(cmd, "file", "validate", "silent"); err != nil {
		return err
	}
	if len(args) != 0 {
		return graphFailure("invalid_selector", "cannot specify both title and --file flag; use --body-file for one Issue's description", 2)
	}
	file, _ := cmd.Flags().GetString("file")
	if file == "" {
		return graphFailure("invalid_selector", "--file requires a markdown path", 2)
	}
	in := createInput{markdownFile: file, createdBy: getActorWithGit(), owner: getOwner(), validationMode: config.GetString("validation.on-create"), jsonOutput: jsonOutput}
	validate, _ := cmd.Flags().GetBool("validate")
	if validate {
		in.validationMode = "error"
	}
	templates, err := parseMarkdownFile(file)
	if err != nil {
		return graphFailure("invalid_properties", fmt.Sprintf("parsing markdown file: %v", err), 2)
	}
	// The ordinary gather/build helpers report their own errors in plain text.
	// Validate this narrower graph route first so --json failures keep the
	// graph diagnostic contract and the shared builder cannot partially print.
	for _, template := range templates {
		for _, raw := range template.Dependencies {
			dependency := strings.TrimSpace(raw)
			if strings.HasPrefix(dependency, "external:") || strings.Contains(dependency, ":external:") {
				return graphFailure("capability_unavailable", "markdown Dependencies on external targets are unavailable in graph mode", 5)
			}
		}
		if in.validationMode == "error" {
			lintIssue := &types.Issue{IssueType: template.IssueType, Description: template.Description, AcceptanceCriteria: template.AcceptanceCriteria}
			if err := validation.LintIssue(lintIssue); err != nil {
				return graphFailure("invalid_properties", fmt.Sprintf("template %q: %v", template.Title, err), 2)
			}
		}
		class, ephemeral, err := resolveCreateStorageClass("", template.IssueType, false, false)
		if err != nil {
			return graphFailure("invalid_properties", fmt.Sprintf("template %q: %v", template.Title, err), 2)
		}
		if class.Normalize() != "" || ephemeral {
			return graphFailure("capability_unavailable", fmt.Sprintf("template %q requires versioned Issue storage in graph mode", template.Title), 5)
		}
	}
	request, err := buildMarkdownBatchRequest(templates, in)
	if err != nil {
		return err
	}
	var last string
	err = withGraphFamiliarStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		records, err := store.CreateIssueBatch(ctx, request)
		if err != nil {
			return nil, "", err
		}
		var summary strings.Builder
		// Ordinary --file --json returns a bare Issue array. Keep that shape
		// and the familiar field names while replacing only the IDs with
		// locally resolvable graph paths.
		issues := make([]*types.Issue, 0, len(records))
		fmt.Fprintf(&summary, "Created %d issues from %s:\n", len(records), in.markdownFile)
		for _, record := range records {
			id := graphPreviewDisplayLocalURL(record.ID)
			fmt.Fprintf(&summary, "  %s: %s [P%d, %s]\n", id, record.Properties.Title, record.Properties.Priority, record.Properties.IssueType)
			issue := *record.Properties
			issue.ID = id
			issues = append(issues, &issue)
			last = id
		}
		return issues, strings.TrimSuffix(summary.String(), "\n"), nil
	})
	if err == nil && last != "" {
		SetLastTouchedID(last)
	}
	return err
}
