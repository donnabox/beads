package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func graphPreviewDependencySnapshot(ctx context.Context, store *graphstore.Store) (graphDependencyProjection, error) {
	snapshot, err := store.CurrentSnapshot(ctx)
	if err != nil {
		return graphDependencyProjection{}, err
	}
	return projectGraphDependencies(snapshot, store.ScopeURL())
}

// The blocking view reads a complete snapshot before projecting a component
// or tree. A positive row cap cannot bound that read, so refuse it up front.
func graphPreviewRejectDependencyRowCap(cmd *cobra.Command, subject ...string) error {
	label := "blocking dependency snapshot"
	if len(subject) != 0 {
		label = subject[0]
	}
	var maxRows int
	var source string
	if cmd.Flags().Changed(maxRowsFlagName) {
		value, err := cmd.Flags().GetInt(maxRowsFlagName)
		if err != nil || value < 0 {
			return graphFailure("invalid_selector", "--max-rows must be a non-negative integer", 2)
		}
		maxRows, source = value, "--"+maxRowsFlagName
	} else {
		value, envSource, _, ok := parseMaxRowsEnv()
		if !ok {
			return graphFailure("invalid_properties", "BEADS_MAX_ROWS must be a non-negative integer", 2)
		}
		maxRows, source = value, envSource
	}
	if maxRows > 0 {
		return graphFailure("capability_unavailable", fmt.Sprintf("%s does not enforce %s=%d; unset the cap or use --max-rows=0", label, source, maxRows), 5)
	}
	return nil
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
		cycles, err := g.nativeCycleIDs()
		if err != nil {
			return nil, "", err
		}
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
		if jsonOutput {
			if err := outputJSON(result); err != nil {
				return err
			}
			if !result.(graphDependencyCheck).Clean {
				return SilentExit()
			}
			return nil
		}
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
	if err := graphPreviewFlags(cmd, "direction", "reverse", "max-depth", "show-all-paths", "max-rows"); err != nil {
		return err
	}
	if err := graphPreviewRejectDependencyRowCap(cmd); err != nil {
		return err
	}
	if jsonOutput {
		return graphFailure("capability_unavailable", "graph dep tree --json cannot preserve native flattened TreeNode Issue fields; use the human blocking tree", 5)
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
		identity, err := store.AllocatedResourceIdentityForPath(ctx, path)
		if err != nil {
			return nil, "", err
		}
		root := store.ScopeURL() + identity.PublicID
		if !g.hasIssue(root) {
			return nil, "", graphstore.ErrNotFound
		}
		var nodes, dependencies, dependents []graphDependencyTreeNode
		if direction == "both" {
			dependencies = g.tree(root, "down", maxDepth)
			dependents = g.tree(root, "up", maxDepth)
			nodes = append(nodes, dependencies...)
			if len(dependents) > 1 {
				nodes = append(nodes, dependents[1:]...)
			}
		} else {
			nodes = g.tree(root, direction, maxDepth)
		}
		if len(nodes) == 1 {
			nodes = nil // An isolated Issue has no dependency tree.
		}
		if len(nodes) == 0 {
			switch direction {
			case "up":
				return nodes, fmt.Sprintf("%s has no dependents", path), nil
			case "both":
				return nodes, fmt.Sprintf("%s has no dependencies or dependents", path), nil
			default:
				return nodes, fmt.Sprintf("%s has no dependencies", path), nil
			}
		}
		var human strings.Builder
		fmt.Fprintf(&human, "Blocking dependency tree for %s (%s; graph preview)\n", path, direction)
		writeNode := func(node graphDependencyTreeNode) {
			fmt.Fprintf(&human, "%s%s  %s\n", strings.Repeat("  ", node.Depth), graphPreviewDisplayLocalURL(node.Issue.ID), node.Issue.Properties.Title)
		}
		if direction == "both" {
			writeNode(nodes[0])
			fmt.Fprintf(&human, "Dependencies (what %s depends on):\n", path)
			if len(dependencies) == 1 {
				human.WriteString("  (none)\n")
			}
			for _, node := range dependencies[1:] {
				writeNode(node)
			}
			fmt.Fprintf(&human, "Dependents (what depends on %s):\n", path)
			if len(dependents) == 1 {
				human.WriteString("  (none)\n")
			}
			for _, node := range dependents[1:] {
				writeNode(node)
			}
		} else {
			for _, node := range nodes {
				writeNode(node)
			}
		}
		return nodes, strings.TrimSuffix(human.String(), "\n"), nil
	})
}

func runGraphPreviewDepCycles(cmd *cobra.Command, args []string) error {
	if err := graphPreviewFlags(cmd, "limit"); err != nil {
		return err
	}
	if jsonOutput {
		return graphFailure("capability_unavailable", "graph dep cycles --json cannot preserve native Cycle member Issue records; use the human blocking cycle report", 5)
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
