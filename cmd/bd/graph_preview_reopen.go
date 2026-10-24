package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func runGraphPreviewReopen(cmd *cobra.Command, args []string) error {
	if err := graphPreviewWritePolicy(); err != nil {
		return err
	}
	if err := graphPreviewFlags(cmd, "reason"); err != nil {
		return err
	}
	if len(args) == 0 {
		return graphFailure("invalid_selector", "graph reopen requires at least one Bead ID or beads/PATH", 2)
	}
	reason, _ := cmd.Flags().GetString("reason")
	var hadError bool
	err := withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		results := make([]graphstore.IssueMutationResult, 0, len(args))
		var human strings.Builder
		for i, selector := range args {
			path, err := graphPreviewReopenPath(ctx, store, selector)
			if err != nil {
				if len(args) == 1 {
					return nil, "", err
				}
				_ = graphStorageError(fmt.Errorf("reopening %s: %w", selector, err))
				hadError = true
				continue
			}
			result, err := store.ReopenIssue(ctx, path, reason, getActorWithGit())
			if err != nil {
				if len(args) == 1 {
					return nil, "", err
				}
				_ = graphStorageError(fmt.Errorf("reopening %s: %w", args[i], err))
				hadError = true
				continue
			}
			if !result.Changed {
				if len(args) == 1 {
					return result, fmt.Sprintf("Unchanged %s (status %q)\n", result.Issue.ID, result.Issue.Properties.Status), nil
				}
				fmt.Fprintf(os.Stderr, "%s is not closed (status: %s); nothing to do\n", args[i], result.Issue.Properties.Status)
				continue
			}
			results = append(results, result)
			fmt.Fprintf(&human, "Reopened %s\n", result.Issue.ID)
		}
		if len(args) == 1 {
			return results[0], human.String(), nil
		}
		return results, strings.TrimSuffix(human.String(), "\n"), nil
	})
	if err != nil {
		return err
	}
	if hadError {
		return SilentExit()
	}
	return nil
}

// Resolve each batch member immediately before its checked native writer. A
// missing later public ID must not erase successful earlier reopen results.
func graphPreviewReopenPath(ctx context.Context, store *graphstore.Store, selector string) (string, error) {
	var path string
	if graphPreviewCanonicalResourceAddress(store.ScopeURL(), selector) {
		identity, err := store.ResolveCanonicalPublicResourceIdentity(ctx, selector)
		if err != nil {
			return "", err
		}
		path = identity.PrivatePath
	} else {
		if strings.Contains(selector, "://") {
			return "", fmt.Errorf("%w: reopen requires a canonical URL in this Scope", errGraphListSelector)
		}
		identity, err := store.ResolveAllocatedResourceIdentity(ctx, selector)
		if err == nil {
			path = identity.PrivatePath
		} else {
			if !errors.Is(err, graphstore.ErrNotFound) {
				return "", err
			}
			literal, parseErr := graphPreviewWritableIdentity(selector)
			if parseErr != nil {
				return "", fmt.Errorf("%w: %v", errGraphListSelector, parseErr)
			}
			path, err = store.ResolveExactAllocatedResourcePath(ctx, literal)
			if err != nil {
				return "", err
			}
		}
	}
	if err := graph.ValidateBeadPath(path); err != nil {
		return "", fmt.Errorf("%w: %v", errGraphListSelector, err)
	}
	return path, nil
}
