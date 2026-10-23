package main

import (
	"fmt"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/memoryquery"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// Validate shared capture context before choosing an output format. This checks
// topology and request binding without paying for a BDP JSON projection that a
// human tree never uses. Owned payloads are not selected edges.
func validateMemoryCurrentCapture(read graphstore.CurrentFamilyRead, scope string, requested []memoryquery.CurrentRoot) error {
	invalid := func() error { return fmt.Errorf("%w: inconsistent captured Memory graph", graphstore.ErrInvalidStore) }
	if graph.ValidateScopeURL(scope) != nil || len(read.Graph.Views) != len(requested) || len(requested) == 0 {
		return invalid()
	}
	type nodeRecord struct{ typ, revision string }
	nodes := map[string]nodeRecord{}
	links := map[string]graphstore.LinkRecord{}
	for _, value := range read.Records {
		var id, typ, revision, version string
		switch r := value.(type) {
		case graphstore.Record:
			id, typ, revision, version = r.ID, r.Type, r.Revision, r.Version
		case graphstore.IssueRecord:
			if r.Properties == nil {
				return invalid()
			}
			id, typ, revision, version = r.ID, r.Type, r.Revision, r.Version
		case graphstore.LinkRecord:
			path, kind, ok := graph.SplitCanonicalURL(scope, r.ID)
			if !ok || kind != graph.KindLink || links[path].ID != "" || r.Type == "" || r.Revision == "" || r.Version != r.Revision {
				return invalid()
			}
			links[path] = r
			continue
		default:
			return invalid()
		}
		path, kind, ok := graph.SplitCanonicalURL(scope, id)
		if !ok || kind != graph.KindBead || nodes[path].revision != "" || typ == "" || revision == "" || version != revision {
			return invalid()
		}
		nodes[path] = nodeRecord{typ, revision}
	}
	if len(nodes) != len(read.Graph.Nodes) || len(links) != len(read.Graph.Links) {
		return invalid()
	}
	summaryNodes, summaryLinks := map[string]bool{}, map[string]bool{}
	for _, n := range read.Graph.Nodes {
		r, ok := nodes[n.Path]
		if !ok || summaryNodes[n.Path] || r.typ != n.TypeURL || r.revision != n.Version {
			return invalid()
		}
		summaryNodes[n.Path] = true
	}
	for _, l := range read.Graph.Links {
		r, ok := links[l.Path()]
		if !ok || summaryLinks[l.Path()] || !l.Source().InScope() || !l.Target().InScope() || r.Type != l.TypeURL() || r.Revision != l.Revision().String() || r.Source != l.Source().URL(scope) || r.Target != l.Target().URL(scope) || r.SourceRevision != l.Source().Pin() || r.TargetRevision != l.Target().Pin() || !summaryNodes[l.Source().Path()] || !summaryNodes[l.Target().Path()] {
			return invalid()
		}
		summaryLinks[l.Path()] = true
	}
	seenNodes, seenLinks := map[string]bool{}, map[string]bool{}
	for i, view := range read.Graph.Views {
		if view.Root != requested[i] {
			return invalid()
		}
		members, edges, frontier := map[string]bool{}, map[string]bool{}, map[string]bool{}
		for _, path := range view.Nodes {
			if !summaryNodes[path] || members[path] {
				return invalid()
			}
			members[path] = true
			seenNodes[path] = true
		}
		if !members[view.Root.Path] {
			return invalid()
		}
		for _, path := range view.Links {
			l, ok := links[path]
			source, sk, sourceOK := graph.SplitCanonicalURL(scope, l.Source)
			target, tk, targetOK := graph.SplitCanonicalURL(scope, l.Target)
			if !ok || edges[path] || !sourceOK || !targetOK || sk != graph.KindBead || tk != graph.KindBead || !members[source] || !members[target] {
				return invalid()
			}
			edges[path] = true
			seenLinks[path] = true
		}
		for _, path := range view.Frontier {
			if !members[path] || frontier[path] {
				return invalid()
			}
			frontier[path] = true
		}
	}
	if len(seenNodes) != len(nodes) || len(seenLinks) != len(links) {
		return invalid()
	}
	return nil
}
