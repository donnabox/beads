package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

type graphDependencyView struct {
	Root         string                   `json:"root"`
	Issues       []graphstore.IssueRecord `json:"issues"`
	Dependencies []graphstore.LinkRecord  `json:"dependencies"`
}

func graphPreviewDependencySnapshot(ctx context.Context, store *graphstore.Store) (graphDependencyProjection, error) {
	snapshot, err := store.CurrentSnapshot(ctx)
	if err != nil {
		return graphDependencyProjection{}, err
	}
	return projectGraphDependencies(snapshot, store.ScopeURL())
}

func runGraphPreviewDependencyGraph(cmd *cobra.Command, args []string) error {
	if err := graphPreviewFlags(cmd); err != nil {
		return err
	}
	if len(args) != 1 {
		return graphFailure("invalid_selector", "graph requires one local Issue ID; --view generic retains generic Link traversal", 2)
	}
	path, err := graphPreviewBeadSelector(args[0])
	if err != nil {
		return err
	}
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		g, err := graphPreviewDependencySnapshot(ctx, store)
		if err != nil {
			return nil, "", err
		}
		root := store.ScopeURL() + path
		if !g.hasIssue(root) {
			return nil, "", graphstore.ErrNotFound
		}
		issues, links := g.component(root)
		result := graphDependencyView{Root: root, Issues: issues, Dependencies: links}
		var human strings.Builder
		fmt.Fprintf(&human, "Blocking dependency graph for %s (%d Issues, %d Dependencies; graph preview)\n", path, len(issues), len(links))
		for _, issue := range issues {
			fmt.Fprintf(&human, "%s  %s  %s\n", graphPreviewDisplayLocalURL(issue.ID), issue.Properties.Status, issue.Properties.Title)
		}
		for _, link := range links {
			fmt.Fprintf(&human, "%s depends on %s\n", graphPreviewDisplayLocalURL(link.Source), graphPreviewDisplayLocalURL(link.Target))
		}
		return result, strings.TrimSuffix(human.String(), "\n"), nil
	})
}

type graphDependencyCheck struct {
	Clean   bool       `json:"clean"`
	Cycles  [][]string `json:"cycles"`
	Summary struct {
		CycleCount int `json:"cycle_count"`
	} `json:"summary"`
}

func runGraphPreviewCheck(cmd *cobra.Command, args []string) error {
	if err := graphPreviewFlags(cmd); err != nil {
		return err
	}
	if len(args) != 0 {
		return graphFailure("invalid_selector", "graph check takes no Issue ID", 2)
	}
	return withGraphStoreOutput(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		g, err := graphPreviewDependencySnapshot(ctx, store)
		if err != nil {
			return nil, "", err
		}
		cycles := g.cycles()
		result := graphDependencyCheck{Clean: len(cycles) == 0, Cycles: cycles}
		result.Summary.CycleCount = len(cycles)
		if result.Clean {
			return result, "Blocking dependency graph is acyclic (graph preview)", nil
		}
		var human strings.Builder
		fmt.Fprintf(&human, "%d blocking dependency cycles (graph preview)\n", len(cycles))
		for _, cycle := range cycles {
			fmt.Fprintln(&human, strings.Join(cycle, " → "))
		}
		return result, strings.TrimSuffix(human.String(), "\n"), nil
	}, func(result any, human string) error {
		if err := graphPrintTo(cmd.OutOrStdout(), result, human, quietFlag, jsonOutput); err != nil {
			return err
		}
		if !result.(graphDependencyCheck).Clean {
			return SilentExit()
		}
		return nil
	})
}

func runGraphPreviewDepTree(cmd *cobra.Command, args []string) error {
	if err := graphPreviewFlags(cmd, "direction", "reverse", "max-depth", "show-all-paths"); err != nil {
		return err
	}
	if len(args) != 1 {
		return graphFailure("invalid_selector", "dep tree requires one local Issue ID", 2)
	}
	path, err := graphPreviewBeadSelector(args[0])
	if err != nil {
		return err
	}
	direction, _ := cmd.Flags().GetString("direction")
	reverse, _ := cmd.Flags().GetBool("reverse")
	if direction == "" {
		direction = "down"
		if reverse {
			direction = "up"
		}
	}
	if direction != "down" && direction != "up" && direction != "both" {
		return graphFailure("invalid_selector", "dep tree direction must be down, up, or both", 2)
	}
	maxDepth, _ := cmd.Flags().GetInt("max-depth")
	if maxDepth < 1 {
		return graphFailure("invalid_selector", "dep tree --max-depth must be at least 1", 2)
	}
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		g, err := graphPreviewDependencySnapshot(ctx, store)
		if err != nil {
			return nil, "", err
		}
		root := store.ScopeURL() + path
		if !g.hasIssue(root) {
			return nil, "", graphstore.ErrNotFound
		}
		var nodes []graphDependencyTreeNode
		if direction == "both" {
			up := g.tree(root, "up", maxDepth)
			if len(up) > 1 {
				nodes = append(nodes, up[1:]...)
			}
			nodes = append(nodes, g.tree(root, "down", maxDepth)...)
		} else {
			nodes = g.tree(root, direction, maxDepth)
		}
		var human strings.Builder
		fmt.Fprintf(&human, "Blocking dependency tree for %s (%s; graph preview)\n", path, direction)
		for _, node := range nodes {
			fmt.Fprintf(&human, "%s%s  %s\n", strings.Repeat("  ", node.Depth), graphPreviewDisplayLocalURL(node.Issue.ID), node.Issue.Properties.Title)
		}
		return nodes, strings.TrimSuffix(human.String(), "\n"), nil
	})
}

func runGraphPreviewDepCycles(cmd *cobra.Command, args []string) error {
	if err := graphPreviewFlags(cmd, "limit"); err != nil {
		return err
	}
	if len(args) != 0 {
		return graphFailure("invalid_selector", "dep cycles takes no Issue ID", 2)
	}
	limit, _ := cmd.Flags().GetInt("limit")
	if limit < 0 {
		return graphFailure("invalid_selector", "dep cycles --limit must be non-negative", 2)
	}
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		g, err := graphPreviewDependencySnapshot(ctx, store)
		if err != nil {
			return nil, "", err
		}
		cycles := g.cycles()
		if len(cycles) == 0 {
			return cycles, "No blocking dependency cycles (graph preview)", nil
		}
		shown := cycles
		if limit > 0 && len(shown) > limit {
			shown = shown[:limit]
		}
		var human strings.Builder
		fmt.Fprintf(&human, "%d blocking dependency cycles (graph preview)\n", len(cycles))
		for _, cycle := range shown {
			fmt.Fprintln(&human, strings.Join(cycle, " → "))
		}
		return cycles, strings.TrimSuffix(human.String(), "\n"), nil
	})
}
