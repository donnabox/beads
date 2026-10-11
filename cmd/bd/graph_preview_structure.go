package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func runGraphPreviewChildren(cmd *cobra.Command, selector string, pretty bool) error {
	if err := graphPreviewFlags(cmd, "pretty"); err != nil {
		return err
	}
	if pretty {
		return graphFailure("capability_unavailable", "graph children tree rendering is unavailable", 5)
	}
	path, err := graphPreviewBeadSelector(selector)
	if err != nil {
		return err
	}
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		children, err := store.ListChildren(ctx, path)
		if err != nil {
			return nil, "", err
		}
		var human strings.Builder
		fmt.Fprintf(&human, "Children of %s (%d; graph preview)\n", graphPreviewDisplayLocalURL(store.ScopeURL()+path), len(children))
		for _, child := range children {
			fmt.Fprintf(&human, "%s  %s  %s\n", graphPreviewDisplayLocalURL(child.ID), child.Properties.Status, child.Properties.Title)
		}
		return children, strings.TrimSuffix(human.String(), "\n"), nil
	})
}

func runGraphPreviewDepList(cmd *cobra.Command, selectors []string) error {
	if err := graphPreviewFlags(cmd, "direction", "type"); err != nil {
		return err
	}
	direction, _ := cmd.Flags().GetString("direction")
	if direction != "down" && direction != "up" {
		return graphFailure("invalid_selector", "dependency direction must be down or up", 2)
	}
	typ, _ := cmd.Flags().GetString("type")
	if typ != "" && typ != "blocks" {
		return graphFailure("capability_unavailable", "graph dependency listing currently supports only blocking Dependencies", 5)
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
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		links := make([]graphstore.LinkRecord, 0)
		for _, path := range paths {
			linkDirection := "out"
			if direction == "up" {
				linkDirection = "in"
			}
			found, err := store.ListIssueDependencies(ctx, path, linkDirection)
			if err != nil {
				return nil, "", err
			}
			links = append(links, found...)
		}
		var human strings.Builder
		for _, link := range links {
			fmt.Fprintf(&human, "%s  %s depends on %s\n", graphPreviewDisplayLocalURL(link.ID), graphPreviewDisplayLocalURL(link.Source), graphPreviewDisplayLocalURL(link.Target))
		}
		if len(links) == 0 {
			human.WriteString("No blocking Dependencies.\n")
		}
		return links, strings.TrimSuffix(human.String(), "\n"), nil
	})
}
