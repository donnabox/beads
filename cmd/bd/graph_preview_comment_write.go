package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/ui"
)

// Both familiar comment entry points use the native comment writer. A comment
// is a feed entry, not a new Issue Resource version or History record.
func runGraphPreviewAddComment(cmd *cobra.Command, args []string, plural bool) error {
	if err := graphPreviewWritePolicy(); err != nil {
		return err
	}
	if plural {
		if err := graphPreviewFlags(cmd, "file", "author"); err != nil {
			return err
		}
	} else if err := graphPreviewFlags(cmd, "stdin", "file"); err != nil {
		return err
	}
	path, err := graphPreviewResourcePath(graphPreviewConfig.GraphScopeURL, args[0])
	if err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	hint := "use positional args, --stdin, or --file"
	if plural {
		hint = "use positional args or -f to read from file"
	}
	commentText, err := requireTextFromSources("comment text", hint, cmdTextSources(cmd, args[1:]))
	if err != nil {
		return graphFailure("invalid_properties", err.Error(), 2)
	}
	author := getActorWithGit()
	if plural {
		if override, _ := cmd.Flags().GetString("author"); override != "" {
			author = override
		}
	}
	err = withGraphStoreOutput(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		comment, issue, err := store.AddIssueCommentWithIssue(ctx, path, author, commentText)
		if err != nil {
			return nil, "", err
		}
		if plural {
			return comment, "Comment added to " + path + "\n", nil
		}
		return comment, fmt.Sprintf("%s Comment added to %s\n", ui.RenderPass("✓"), formatFeedbackID(path, issue.Properties.Title)), nil
	}, func(result any, human string) error {
		if jsonOutput {
			return outputJSON(result)
		}
		if quietFlag {
			return nil
		}
		_, err := fmt.Fprint(cmd.OutOrStdout(), human)
		return err
	})
	if err == nil && !plural {
		SetLastTouchedID(path)
	}
	return err
}
