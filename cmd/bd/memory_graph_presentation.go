package main

import (
	"bytes"
	"fmt"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// renderMemoryCurrentGraph presents only captured ordinary-family membership.
// Each root gets an independent tree; canonical Owned copies are never edges.
// It neither selects another graph nor publishes output. The caller owns the
// command writer and may publish only the complete successful return value.
func renderMemoryCurrentGraph(read graphstore.CurrentFamilyRead, scope string, direction graph.Direction, maxBytes int) (string, error) {
	invalid := func() error { return fmt.Errorf("%w: inconsistent captured Memory graph", graphstore.ErrInvalidStore) }
	if graph.ValidateScopeURL(scope) != nil || !direction.Valid() || maxBytes < 1 || len(read.Graph.Views) == 0 {
		return "", invalid()
	}
	nodes := map[string]graphGenericNode{}
	links := map[string]graphGenericLink{}
	for _, raw := range read.Records {
		var n graphGenericNode
		switch r := raw.(type) {
		case graphstore.Record:
			n = graphGenericNode{ID: r.ID, Type: r.Type, Version: r.Version, Title: r.Properties.Title}
		case graphstore.IssueRecord:
			if r.Properties == nil {
				return "", invalid()
			}
			n = graphGenericNode{ID: r.ID, Type: r.Type, Version: r.Version, Title: r.Properties.Title, Status: string(r.Properties.Status)}
		case graphstore.LinkRecord:
			path, kind, ok := graph.SplitCanonicalURL(scope, r.ID)
			if !ok || kind != graph.KindLink || links[path].ID != "" || r.Type == "" || r.Version == "" {
				return "", invalid()
			}
			links[path] = graphGenericLink{ID: r.ID, Type: r.Type, Version: r.Version, Source: r.Source, Target: r.Target, SourceRevision: r.SourceRevision, TargetRevision: r.TargetRevision}
			continue
		default:
			return "", invalid()
		}
		path, kind, ok := graph.SplitCanonicalURL(scope, n.ID)
		if !ok || kind != graph.KindBead || nodes[path].ID != "" || n.Type == "" || n.Version == "" {
			return "", invalid()
		}
		nodes[path] = n
	}
	// Bind summaries to the same captured identity/Type/version and endpoint pins.
	// Never use record extras or union membership to re-induce a root's Links.
	if len(nodes) != len(read.Graph.Nodes) || len(links) != len(read.Graph.Links) {
		return "", invalid()
	}
	graphNodes, graphLinks := map[string]bool{}, map[string]bool{}
	for _, n := range read.Graph.Nodes {
		r, ok := nodes[n.Path]
		if !ok || graphNodes[n.Path] || r.Type != n.TypeURL || r.Version != n.Version {
			return "", invalid()
		}
		graphNodes[n.Path] = true
	}
	for _, l := range read.Graph.Links {
		r, ok := links[l.Path()]
		if !ok || graphLinks[l.Path()] || !l.Source().InScope() || !l.Target().InScope() || r.Type != l.TypeURL() || r.Version != l.Revision().String() || r.Source != l.Source().URL(scope) || r.Target != l.Target().URL(scope) || r.SourceRevision != l.Source().Pin() || r.TargetRevision != l.Target().Pin() || !graphNodes[l.Source().Path()] || !graphNodes[l.Target().Path()] {
			return "", invalid()
		}
		graphLinks[l.Path()] = true
	}
	var output bytes.Buffer
	w := &graphExportWriter{out: &output}
	line := func(s string) error {
		if len(s)+1 > maxBytes-output.Len() {
			return fmt.Errorf("%w: Memory output exceeds %d bytes", graphstore.ErrLimitExceeded, maxBytes)
		}
		w.println(s)
		return w.wrapError("memory")
	}
	seenNodes, seenLinks := map[string]bool{}, map[string]bool{}
	for i, view := range read.Graph.Views {
		if graph.ValidateBeadPath(view.Root.Path) != nil || view.Root.Depth < 0 || view.Root.Depth > graphGenericBound {
			return "", invalid()
		}
		r := graphGenericResult{Scope: scope, Root: graph.CanonicalURL(scope, view.Root.Path), Direction: direction.String(), Depth: view.Root.Depth, Complete: len(view.Frontier) == 0}
		members, edges, frontier := map[string]bool{}, map[string]bool{}, map[string]bool{}
		for _, path := range view.Nodes {
			if !graphNodes[path] || members[path] {
				return "", invalid()
			}
			members[path], seenNodes[path] = true, true
			r.Nodes = append(r.Nodes, nodes[path])
		}
		if !members[view.Root.Path] {
			return "", invalid()
		}
		for _, path := range view.Links {
			l, ok := links[path]
			source, _, sourceOK := graph.SplitCanonicalURL(scope, l.Source)
			target, _, targetOK := graph.SplitCanonicalURL(scope, l.Target)
			if !ok || edges[path] || !sourceOK || !targetOK || !members[source] || !members[target] {
				return "", invalid()
			}
			edges[path], seenLinks[path] = true, true
			r.Links = append(r.Links, l)
		}
		for _, path := range view.Frontier {
			if !members[path] || frontier[path] {
				return "", invalid()
			}
			frontier[path] = true
			r.Frontier = append(r.Frontier, graph.CanonicalURL(scope, path))
		}
		if i > 0 {
			if err := line(""); err != nil {
				return "", err
			}
		}
		if err := renderGraphRelationshipTreePart(w, r, maxBytes-output.Len(), false); err != nil {
			return "", err
		}
		// Quote display identifiers before framing; arbitrary data cannot inject
		// another root or a false completion/frontier line.
		label := (&graphRelationshipTree{result: r}).label
		if err := line(fmt.Sprintf("Root %s: complete: %t; direction: %s; depth: %d; frontier: %d", label(r.Root), r.Complete, r.Direction, r.Depth, len(r.Frontier))); err != nil {
			return "", err
		}
		for _, id := range r.Frontier {
			if err := line(fmt.Sprintf("Root %s: more relationships beyond depth %d: %s", label(r.Root), r.Depth, label(id))); err != nil {
				return "", err
			}
		}
	}
	if len(seenNodes) != len(nodes) || len(seenLinks) != len(links) {
		return "", invalid()
	}
	if err := line(""); err != nil {
		return "", err
	}
	if err := line(fmt.Sprintf("%d Beads, %d Links", len(seenNodes), len(seenLinks))); err != nil {
		return "", err
	}
	return output.String(), nil
}
