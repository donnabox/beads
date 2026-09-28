package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
)

// Returned canonical IDs can be fed back to the CLI. Only the exact local
// spelling is accepted; this preview never resolves aliases or remote stores.
func graphPreviewResourcePath(scope, selector string) (string, error) {
	if graph.ValidateBeadPath(selector) == nil || graph.ValidateLinkPath(selector) == nil {
		return selector, nil
	}
	if path, kind, ok := graph.SplitCanonicalURL(scope, selector); ok && (kind == graph.KindBead || kind == graph.KindLink) {
		return path, nil
	}
	return "", fmt.Errorf("expected a canonical local beads/PATH or links/PATH, or its exact Scope URL")
}

// Both familiar spellings and the experimental generic Type selector reach
// one domain operation. No alternate Link writer can bypass Issue policy.
func runGraphPreviewAddDependency(cmd *cobra.Command, args []string) error {
	if err := graphPreviewWritePolicy(); err != nil {
		return err
	}
	if err := graphPreviewFlags(cmd, "type", "resource-type", "if-source-revision", "unconditional-source"); err != nil {
		return err
	}
	if len(args) != 2 {
		return graphFailure("invalid_selector", "graph Dependency creation requires two canonical beads/PATH arguments", 2)
	}
	paths := make([]string, len(args))
	for i, selector := range args {
		path, err := graphPreviewResourcePath(graphPreviewConfig.GraphScopeURL, selector)
		if err != nil {
			return graphFailure("invalid_selector", err.Error(), 2)
		}
		if err := graph.ValidateBeadPath(path); err != nil {
			return graphFailure("invalid_selector", err.Error(), 2)
		}
		paths[i] = path
	}
	var sourceRevision string
	if cmd.Flags().Changed("resource-type") {
		if cmd.Flags().Changed("type") {
			return graphFailure("invalid_selector", "select either --type or --resource-type", 2)
		}
		sourceRevision, _ = cmd.Flags().GetString("if-source-revision")
		unconditional, _ := cmd.Flags().GetBool("unconditional-source")
		if cmd.Flags().Changed("if-source-revision") == cmd.Flags().Changed("unconditional-source") || (sourceRevision == "" && !unconditional) {
			return graphFailure("invalid_selector", "generic owned Link creation requires either --if-source-revision REVISION or --unconditional-source", 2)
		}
		resourceType, _ := cmd.Flags().GetString("resource-type")
		if resourceType != graphstore.DependencyTypeURL(graphPreviewConfig.GraphScopeURL) {
			return graphFailure("capability_unavailable", "this preview accepts only the installed Type "+graphstore.DependencyTypeURL(graphPreviewConfig.GraphScopeURL), 5)
		}
	} else {
		if cmd.Flags().Changed("if-source-revision") || cmd.Flags().Changed("unconditional-source") {
			return graphFailure("invalid_selector", "source guard options require --resource-type", 2)
		}
		raw, _ := cmd.Flags().GetString("type")
		typ := canonicalDependencyType(types.DependencyType(raw))
		if err := validateDependencyType(typ); err != nil {
			return graphFailure("invalid_properties", err.Error(), 2)
		}
		if typ != types.DepBlocks {
			return graphFailure("capability_unavailable", "this preview supports only local blocking Dependencies", 5)
		}
	}
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		result, err := store.AddDependency(ctx, graphstore.DependencyRequest{SourcePath: paths[0], TargetPath: paths[1], Actor: getActorWithGit(), ExpectedSourceRevision: sourceRevision})
		if err != nil {
			return nil, "", err
		}
		verb := "Created"
		if !result.Changed {
			verb = "Unchanged"
		}
		return result, fmt.Sprintf("%s %s: %s depends on %s\n", verb, result.Link.ID, args[0], args[1]), nil
	})
}

func runGraphPreviewClose(cmd *cobra.Command, args []string) error {
	if err := graphPreviewWritePolicy(); err != nil {
		return err
	}
	if err := graphPreviewFlags(cmd, "reason", "resolution", "message", "comment"); err != nil {
		return err
	}
	if len(args) != 1 {
		return graphFailure("invalid_selector", "graph close requires one canonical beads/PATH", 2)
	}
	path, err := graphPreviewResourcePath(graphPreviewConfig.GraphScopeURL, args[0])
	if err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	if err := graph.ValidateBeadPath(path); err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	reasons, _, err := resolveCloseReasons(cmd, args)
	if err != nil {
		return graphFailure("invalid_properties", err.Error(), 2)
	}
	if err := validateCloseReasons(reasons); err != nil {
		return graphFailure("invalid_properties", err.Error(), 2)
	}
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		result, err := store.CloseIssue(ctx, path, reasons[0], getActorWithGit())
		if err != nil {
			return nil, "", err
		}
		verb := "Closed"
		if !result.Changed {
			verb = "Already closed"
		}
		return result, fmt.Sprintf("%s %s\n", verb, args[0]), nil
	})
}

func runGraphPreviewReady(cmd *cobra.Command, args []string) error {
	if err := graphPreviewFlags(cmd); err != nil {
		return err
	}
	if len(args) != 0 {
		return graphFailure("invalid_selector", "graph ready takes no positional arguments", 2)
	}
	// The legacy resolver warns and ignores malformed environment values.
	// This deliberately bounded preview refuses unsupported policy instead.
	if raw := os.Getenv(maxRowsEnvVar); raw != "" {
		maxRows, err := strconv.Atoi(raw)
		if err != nil || maxRows < 0 {
			return graphFailure("invalid_properties", "BEADS_MAX_ROWS must be a non-negative integer", 2)
		}
		if maxRows > 0 {
			return graphFailure("capability_unavailable", "graph ready preview does not implement a configured row cap; unset BEADS_MAX_ROWS to request the unfiltered view", 5)
		}
	}
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		issues, err := store.ReadyIssues(ctx)
		if err != nil {
			return nil, "", err
		}
		var human strings.Builder
		for _, issue := range issues {
			fmt.Fprintf(&human, "%s  %s\n", issue.ID, issue.Properties.Title)
		}
		if len(issues) == 0 {
			human.WriteString("No ready Issues.\n")
		}
		return issues, strings.TrimSuffix(human.String(), "\n"), nil
	})
}
