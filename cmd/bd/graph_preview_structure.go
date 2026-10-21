package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
)

func runGraphPreviewDepList(cmd *cobra.Command, selectors []string) error {
	if err := graphPreviewFlags(cmd, "direction", "type"); err != nil {
		return err
	}
	direction, _ := cmd.Flags().GetString("direction")
	if direction != "down" && direction != "up" {
		return graphFailure("invalid_selector", "dependency direction must be down or up", 2)
	}
	typ, _ := cmd.Flags().GetString("type")
	if typ != "" && !types.DependencyType(typ).IsValid() {
		return graphFailure("invalid_properties", "dependency type must be non-empty and fit the native type column", 2)
	}
	paths := make([]string, 0, len(selectors))
	seen := make(map[string]bool, len(selectors))
	for _, selector := range selectors {
		path, err := graphPreviewBeadSelector(selector)
		if err != nil {
			return err
		}
		if !seen[path] {
			paths = append(paths, path)
			seen[path] = true
		}
	}
	if jsonOutput {
		return graphFailure("capability_unavailable", "graph dep list --json cannot preserve native single-Issue versus batch Dependency record shapes; use the human blocking view", 5)
	}
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		links := make([]graphstore.LinkRecord, 0)
		for _, path := range paths {
			linkDirection := "out"
			if direction == "up" {
				linkDirection = "in"
			}
			found, err := store.ListIssueDependencies(ctx, path, linkDirection, typ)
			if err != nil {
				return nil, "", err
			}
			links = append(links, found...)
		}
		var human strings.Builder
		for _, link := range links {
			fmt.Fprintf(&human, "%s  %s depends on %s via %s\n", graphPreviewDisplayLocalURL(link.ID), graphPreviewDisplayLocalURL(link.Source), graphPreviewDisplayLocalURL(link.Target), link.Properties["dep_type"])
		}
		if len(links) == 0 {
			human.WriteString("No Dependencies.\n")
		}
		return links, strings.TrimSuffix(human.String(), "\n"), nil
	})
}
