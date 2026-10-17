package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/internal/ui"
	"github.com/steveyegge/beads/internal/validation"
	publicops "github.com/steveyegge/beads/issueops"
)

// TODOs remain ordinary task Issues. The graph adapter selects the same native
// writer used by create/close, so each changed Issue has one retained version.
func runGraphPreviewTodoAdd(cmd *cobra.Command, args []string) error {
	if err := graphPreviewWritePolicy(); err != nil {
		return err
	}
	if err := graphPreviewFlags(cmd, "priority", "description"); err != nil {
		return err
	}
	title := strings.Join(args, " ")
	if strings.TrimSpace(title) == "" {
		return graphFailure("invalid_properties", "TODO title must not be blank", 2)
	}
	priority, _ := cmd.Flags().GetInt("priority")
	if _, err := validation.ValidatePriority(fmt.Sprint(priority)); err != nil {
		return graphFailure("invalid_properties", err.Error(), 2)
	}
	description, _ := cmd.Flags().GetString("description")
	actor := getActorWithGit()
	var path string
	err := withGraphStoreOutput(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		record, err := store.CreateIssue(ctx, "", publicops.CreateRequest{Actor: actor, Issue: &types.Issue{
			Title: title, Description: description, Priority: priority,
			IssueType: types.TypeTask, Status: types.StatusOpen,
			Assignee: actor, Owner: getOwner(), CreatedBy: actor,
		}})
		if err != nil {
			return nil, "", err
		}
		path = graphPreviewDisplayLocalURL(record.ID)
		issue := *record.Properties
		issue.ID = path
		issue.Metadata = record.Metadata
		return &issue, fmt.Sprintf("Created %s: %s\n", ui.RenderID(path), issue.Title), nil
	}, func(result any, human string) error {
		if jsonOutput {
			return graphTodoJSON(cmd, result)
		}
		_, err := fmt.Fprint(cmd.OutOrStdout(), human)
		return err
	})
	if err == nil {
		SetLastTouchedID(path)
	}
	return err
}

func runGraphPreviewTodoList(cmd *cobra.Command) error {
	if err := graphPreviewFlags(cmd, "all"); err != nil {
		return err
	}
	showAll, _ := cmd.Flags().GetBool("all")
	status := string(types.StatusOpen)
	if showAll {
		status = ""
	}
	// Native todo has no pagination. Zero asks the shared Issue query for the
	// complete bounded graph inventory, and HasMore is refused defensively.
	limit := 0
	return withGraphStoreOutput(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		page, err := store.ListIssues(ctx, publicops.ListRequest{
			IssueType: string(types.TypeTask), Status: status, AllFlag: true,
			Limit: &limit, MaxRows: graphstore.PreviewSnapshotLimit,
		})
		if err != nil {
			return nil, "", err
		}
		if page.HasMore {
			return nil, "", fmt.Errorf("%w: TODO list would be incomplete", graphstore.ErrLimitExceeded)
		}
		issues := make([]*types.Issue, 0, len(page.Items))
		for _, record := range page.Items {
			if record.Properties == nil || record.Properties.IssueType.Normalize() != types.TypeTask {
				return nil, "", fmt.Errorf("%w: TODO query returned a non-task Issue", graphstore.ErrInvalidStore)
			}
			issue := *record.Properties
			issue.ID = graphPreviewDisplayLocalURL(record.ID)
			issue.Metadata = record.Metadata
			issues = append(issues, &issue)
		}
		if jsonOutput {
			data, err := json.MarshalIndent(issues, "", "  ")
			if err != nil {
				return nil, "", err
			}
			if len(data) > graphIssueListOutputLimit {
				return nil, "", fmt.Errorf("%w: TODO JSON exceeds the graph list output bound", graphstore.ErrLimitExceeded)
			}
			return issues, "", nil
		}
		if len(issues) == 0 {
			return issues, "No TODOs found\n", nil
		}
		todoSortIssues(issues)
		var human strings.Builder
		for _, issue := range issues {
			fmt.Fprintf(&human, "  %s %s  %-40s  %s  %s\n",
				ui.RenderStatusIcon(string(issue.Status)), ui.RenderID(issue.ID),
				todoTruncate(issue.Title, 40), ui.RenderPriority(issue.Priority), issue.Status)
		}
		fmt.Fprintf(&human, "\nTotal: %d TODOs\n", len(issues))
		return issues, human.String(), nil
	}, func(result any, human string) error {
		if jsonOutput {
			return graphTodoJSON(cmd, result)
		}
		_, err := fmt.Fprint(cmd.OutOrStdout(), human)
		return err
	})
}

func runGraphPreviewTodoDone(cmd *cobra.Command, args []string) error {
	if err := graphPreviewWritePolicy(); err != nil {
		return err
	}
	if err := graphPreviewFlags(cmd, "reason", "if-revision"); err != nil {
		return err
	}
	revision, _, err := graphPreviewEditRevisionGuard(cmd)
	if err != nil {
		return err
	}
	if revision != "" && len(args) != 1 {
		return graphFailure("invalid_selector", "todo done --if-revision requires exactly one Issue", 2)
	}
	reason, _ := cmd.Flags().GetString("reason")
	if reason == "" {
		reason = "Completed"
	}
	// A guarded write is one atomic decision and keeps typed refusals. Native
	// unguarded batches instead try every selector independently, even after a
	// malformed or missing ID, and return success with the IDs they closed.
	if revision != "" {
		if _, err := graphTodoIssuePath(args[0]); err != nil {
			return graphFailure("invalid_selector", err.Error(), 2)
		}
	}
	var firstSettled string
	err = withGraphStoreOutput(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		closed := make([]string, 0, len(args))
		var human strings.Builder
		for _, selector := range args {
			path, err := graphTodoIssuePath(selector)
			if err != nil {
				graphTodoTargetError(selector, err, true)
				continue
			}
			result, err := store.CloseIssueWithOptions(ctx, graphstore.IssueCloseRequest{
				Path: path, Reason: reason, Actor: getActorWithGit(), ExpectedRevision: revision,
			})
			if err != nil {
				if revision != "" {
					return nil, "", err
				}
				graphTodoTargetError(selector, err, errors.Is(err, graphstore.ErrNotFound))
				continue
			}
			if firstSettled == "" {
				firstSettled = path
			}
			closed = append(closed, path)
			fmt.Fprintf(&human, "Closed %s\n", ui.RenderID(path))
			if result.OpenChildren > 0 {
				fmt.Fprintf(os.Stderr, "warning: closing %s with %d open child issue(s) still active\n", path, result.OpenChildren)
			}
		}
		return map[string]any{"closed": closed, "reason": reason}, human.String(), nil
	}, func(result any, human string) error {
		if jsonOutput {
			return graphTodoJSON(cmd, result)
		}
		_, err := fmt.Fprint(cmd.OutOrStdout(), human)
		return err
	})
	if err != nil {
		return err
	}
	if firstSettled != "" {
		SetLastTouchedID(firstSettled)
	}
	return nil
}

func graphTodoIssuePath(selector string) (string, error) {
	path, err := graphPreviewResourcePath(graphPreviewConfig.GraphScopeURL, selector)
	if err != nil {
		return "", err
	}
	if err := graph.ValidateBeadPath(path); err != nil {
		return "", err
	}
	return path, nil
}

func graphTodoTargetError(selector string, err error, lookup bool) {
	verb := "close"
	if lookup {
		verb = "get issue"
	}
	fmt.Fprintf(os.Stderr, "Error: failed to %s %s: %v\n", verb, selector, err) //nolint:gosec // Stderr diagnostics, not a browser context.
}

// The ordinary todo command emits raw JSON for add, list and done. Preserve
// that shape instead of adding the graph envelope or a schema_version field.
func graphTodoJSON(cmd *cobra.Command, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(cmd.OutOrStdout(), string(data))
	return err
}
