package main

import (
	"bytes"
	"errors"
	"fmt"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/memoryquery"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

type memoryConstraintCapture struct {
	scope string
	nodes map[string]graphGenericNode
	links map[string]graphGenericLink
}

func invalidMemoryConstraintCapture() error {
	return fmt.Errorf("%w: inconsistent captured constraint graph", graphstore.ErrInvalidStore)
}

// Bind only captured canonical records; neither Owned copies nor Type metadata
// supply graph membership. This does not acquire or classify another resource.
func bindMemoryConstraintCapture(read graphstore.CurrentFamilyRead, scope string) (*memoryConstraintCapture, error) {
	c := &memoryConstraintCapture{scope: scope, nodes: map[string]graphGenericNode{}, links: map[string]graphGenericLink{}}
	for _, raw := range read.Records {
		var n graphGenericNode
		switch r := raw.(type) {
		case graphstore.Record:
			n = graphGenericNode{ID: r.ID, Type: r.Type, Version: r.Version, Title: r.Properties.Title}
		case graphstore.IssueRecord:
			if r.Properties == nil {
				return nil, invalidMemoryConstraintCapture()
			}
			n = graphGenericNode{ID: r.ID, Type: r.Type, Version: r.Version, Title: r.Properties.Title, Status: string(r.Properties.Status)}
		case graphstore.LinkRecord:
			path, kind, ok := graph.SplitCanonicalURL(scope, r.ID)
			if !ok || kind != graph.KindLink || c.links[path].ID != "" || r.Type == "" || r.Version == "" {
				return nil, invalidMemoryConstraintCapture()
			}
			c.links[path] = graphGenericLink{ID: r.ID, Type: r.Type, Version: r.Version, Source: r.Source, Target: r.Target, SourceRevision: r.SourceRevision, TargetRevision: r.TargetRevision}
			continue
		default:
			return nil, invalidMemoryConstraintCapture()
		}
		path, kind, ok := graph.SplitCanonicalURL(scope, n.ID)
		if !ok || kind != graph.KindBead || c.nodes[path].ID != "" || n.Type == "" || n.Version == "" {
			return nil, invalidMemoryConstraintCapture()
		}
		c.nodes[path] = n
	}
	if len(c.nodes) != len(read.Graph.Nodes) || len(c.links) != len(read.Graph.Links) {
		return nil, invalidMemoryConstraintCapture()
	}
	nodes, links := map[string]bool{}, map[string]bool{}
	for _, n := range read.Graph.Nodes {
		r, ok := c.nodes[n.Path]
		if !ok || nodes[n.Path] || r.Type != n.TypeURL || r.Version != n.Version {
			return nil, invalidMemoryConstraintCapture()
		}
		nodes[n.Path] = true
	}
	for _, l := range read.Graph.Links {
		r, ok := c.links[l.Path()]
		if !ok || links[l.Path()] || !l.Source().InScope() || !l.Target().InScope() || !nodes[l.Source().Path()] || !nodes[l.Target().Path()] || r.Type != l.TypeURL() || r.Version != l.Revision().String() || r.Source != l.Source().URL(scope) || r.Target != l.Target().URL(scope) || r.SourceRevision != l.Source().Pin() || r.TargetRevision != l.Target().Pin() {
			return nil, invalidMemoryConstraintCapture()
		}
		links[l.Path()] = true
	}
	return c, nil
}

func memoryConstraintSet(paths []string) (map[string]bool, error) {
	s := map[string]bool{}
	for _, p := range paths {
		if s[p] {
			return nil, invalidMemoryConstraintCapture()
		}
		s[p] = true
	}
	return s, nil
}
func sameMemoryConstraintSet(paths []string, want map[string]bool) bool {
	got, err := memoryConstraintSet(paths)
	if err != nil || len(got) != len(want) {
		return false
	}
	for p := range got {
		if !want[p] {
			return false
		}
	}
	return true
}

func (c *memoryConstraintCapture) view(v memoryquery.CurrentView) (graphGenericResult, error) {
	r := graphGenericResult{Scope: c.scope, Root: graph.CanonicalURL(c.scope, v.Root.Path), Depth: v.Root.Depth, Complete: len(v.Frontier) == 0}
	nodes, err := memoryConstraintSet(v.Nodes)
	if err != nil {
		return r, err
	}
	if !nodes[v.Root.Path] || graph.ValidateBeadPath(v.Root.Path) != nil || v.Root.Depth < 0 || v.Root.Depth > graphstore.PreviewSnapshotLimit {
		return r, invalidMemoryConstraintCapture()
	}
	for _, p := range v.Nodes {
		n, ok := c.nodes[p]
		if !ok {
			return r, invalidMemoryConstraintCapture()
		}
		r.Nodes = append(r.Nodes, n)
	}
	if _, err := memoryConstraintSet(v.Links); err != nil {
		return r, err
	}
	for _, p := range v.Links {
		l, ok := c.links[p]
		source, _, src := graph.SplitCanonicalURL(c.scope, l.Source)
		target, _, dst := graph.SplitCanonicalURL(c.scope, l.Target)
		if !ok || !src || !dst || !nodes[source] || !nodes[target] {
			return r, invalidMemoryConstraintCapture()
		}
		r.Links = append(r.Links, l)
	}
	if _, err := memoryConstraintSet(v.Frontier); err != nil {
		return r, err
	}
	for _, p := range v.Frontier {
		if !nodes[p] {
			return r, invalidMemoryConstraintCapture()
		}
		r.Frontier = append(r.Frontier, graph.CanonicalURL(c.scope, p))
	}
	return r, nil
}

type memoryConstraintRootRenderer func(*memoryConstraintCapture, graphGenericResult, memoryquery.CurrentView, int, *graphExportWriter, int) error

// Root renderers share the existing graph writer and one private atomic buffer.
// A budget refusal names the configured total cap, not a root's remainder.
func renderMemoryConstraintCapture(read graphstore.CurrentFamilyRead, scope string, maxBytes int, root memoryConstraintRootRenderer) (string, error) {
	if graph.ValidateScopeURL(scope) != nil || maxBytes < 1 {
		return "", fmt.Errorf("%w: invalid constraint presentation request", graph.ErrValidation)
	}
	if len(read.Graph.Views) == 0 {
		return "", invalidMemoryConstraintCapture()
	}
	c, err := bindMemoryConstraintCapture(read, scope)
	if err != nil {
		return "", err
	}
	var out bytes.Buffer
	w := &graphExportWriter{out: &out}
	limit := func() error {
		return fmt.Errorf("%w: Memory output exceeds %d bytes", graphstore.ErrLimitExceeded, maxBytes)
	}
	line := func(s string) error {
		if len(s)+1 > maxBytes-out.Len() {
			return limit()
		}
		w.println(s)
		return w.wrapError("memory")
	}
	nodes, links := map[string]bool{}, map[string]bool{}
	for i, v := range read.Graph.Views {
		r, err := c.view(v)
		if err != nil {
			return "", err
		}
		for _, p := range v.Nodes {
			nodes[p] = true
		}
		for _, p := range v.Links {
			links[p] = true
		}
		if i > 0 {
			if err := line(""); err != nil {
				return "", err
			}
		}
		if err := root(c, r, v, i, w, maxBytes-out.Len()); err != nil {
			if errors.Is(err, graphstore.ErrLimitExceeded) {
				return "", limit()
			}
			return "", err
		}
	}
	if len(nodes) != len(c.nodes) || len(links) != len(c.links) {
		return "", invalidMemoryConstraintCapture()
	}
	if err := line(""); err != nil {
		return "", err
	}
	if err := line(fmt.Sprintf("Aggregate: %d unique Beads, %d unique Links across %d root views", len(nodes), len(links), len(read.Graph.Views))); err != nil {
		return "", err
	}
	return out.String(), nil
}

// Direct subjects is fixed incoming Subject To, not a tunable traversal. Every
// captured root remains visible even when there are no direct declarations.
// The caller supplies the capture from ReadCurrentSubjects, not another family.
func renderMemorySubjectsGraph(read graphstore.CurrentFamilyRead, scope string, maxBytes int) (string, error) {
	return renderMemoryConstraintCapture(read, scope, maxBytes, func(c *memoryConstraintCapture, r graphGenericResult, v memoryquery.CurrentView, _ int, w *graphExportWriter, remaining int) error {
		if v.Root.Depth != 1 || len(v.Frontier) != 0 {
			return invalidMemoryConstraintCapture()
		}
		r.Direction = "in"
		t, err := newGraphRelationshipTree(w, r, remaining)
		if err != nil {
			return err
		}
		members := map[string]bool{v.Root.Path: true}
		for _, l := range r.Links {
			if l.Target != r.Root {
				return invalidMemoryConstraintCapture()
			}
			p, _, _ := graph.SplitCanonicalURL(scope, l.Source)
			members[p] = true
		}
		if !sameMemoryConstraintSet(v.Nodes, members) {
			return invalidMemoryConstraintCapture()
		}
		t.println(t.nodeLabel(r.Root))
		seen := map[string]bool{r.Root: true}
		rows := []graphTreeLine{}
		for _, l := range t.adjacent[r.Root] {
			row := graphTreeLine{link: l, next: l.Source, direction: -1}
			if seen[l.Source] {
				row.reference = " (shown elsewhere)"
				if t.closesCycle(l) {
					row.reference += " (cycle)"
				}
			}
			seen[l.Source] = true
			rows = append(rows, row)
		}
		t.writeLines(rows, "")
		t.println(fmt.Sprintf("Root %s: fixed direct incoming Subject To scope; complete: true", t.label(r.Root)))
		if t.err != nil {
			return t.err
		}
		return w.wrapError("memory")
	})
}

// Phase1 is one original-subject occurrence; only phase2 expands Requires.
// Separate trees per slot/phase implement (slot,phase,identity), without making
// synthetic resource IDs or reconstructing either phase from the union graph.
func renderMemoryConstraintsGraph(read graphstore.CurrentConstraintsRead, scope string, maxBytes int) (string, error) {
	return renderMemoryConstraintCapture(read.CurrentFamilyRead, scope, maxBytes, func(c *memoryConstraintCapture, r graphGenericResult, v memoryquery.CurrentView, slot int, w *graphExportWriter, remaining int) error {
		if len(read.Stages) != len(read.Graph.Views) {
			return invalidMemoryConstraintCapture()
		}
		stage := read.Stages[slot]
		if stage.Root.Path != v.Root.Path || stage.Root.RequiresDepth != v.Root.Depth {
			return invalidMemoryConstraintCapture()
		}
		seeds, err := memoryConstraintSet(stage.Requires.Seeds)
		if err != nil {
			return err
		}
		phaseNodes, err := memoryConstraintSet(stage.Requires.Nodes)
		if err != nil {
			return err
		}
		phaseLinks, err := memoryConstraintSet(stage.Requires.Links)
		if err != nil {
			return err
		}
		phaseFrontier, err := memoryConstraintSet(stage.Requires.Frontier)
		if err != nil {
			return err
		}
		unionNodes := map[string]bool{v.Root.Path: true}
		unionLinks := map[string]bool{}
		for p := range phaseNodes {
			unionNodes[p] = true
		}
		for p := range phaseLinks {
			unionLinks[p] = true
		}
		phase1 := r
		phase1.Direction = "out"
		phase1.Links = nil
		targets := map[string]bool{}
		for _, seed := range stage.SubjectTo {
			l, ok := c.links[seed.Link]
			if !ok || unionLinks[seed.Link] || l.Source != r.Root || l.Target != graph.CanonicalURL(scope, seed.Target) || !seeds[seed.Target] {
				return invalidMemoryConstraintCapture()
			}
			unionLinks[seed.Link] = true
			targets[seed.Target] = true
			phase1.Links = append(phase1.Links, l)
		}
		if !sameMemoryConstraintSet(stage.Requires.Seeds, targets) || !sameMemoryConstraintSet(v.Nodes, unionNodes) || !sameMemoryConstraintSet(v.Links, unionLinks) || !sameMemoryConstraintSet(v.Frontier, phaseFrontier) {
			return invalidMemoryConstraintCapture()
		}
		for seed := range seeds {
			if !phaseNodes[seed] {
				return invalidMemoryConstraintCapture()
			}
		}
		for p := range phaseFrontier {
			if !phaseNodes[p] {
				return invalidMemoryConstraintCapture()
			}
		}
		first, err := newGraphRelationshipTree(w, phase1, remaining)
		if err != nil {
			return err
		}
		// Stage membership is already admitted. It is not induced from phase1.
		requires := graphGenericResult{Scope: scope, All: true, Direction: "out"}
		for _, p := range stage.Requires.Nodes {
			n, ok := c.nodes[p]
			if !ok {
				return invalidMemoryConstraintCapture()
			}
			requires.Nodes = append(requires.Nodes, n)
		}
		for _, p := range stage.Requires.Links {
			l, ok := c.links[p]
			source, _, src := graph.SplitCanonicalURL(scope, l.Source)
			target, _, dst := graph.SplitCanonicalURL(scope, l.Target)
			if !ok || !src || !dst || !phaseNodes[source] || !phaseNodes[target] {
				return invalidMemoryConstraintCapture()
			}
			requires.Links = append(requires.Links, l)
		}
		t, err := newGraphRelationshipTree(w, requires, remaining)
		if err != nil {
			return err
		}
		anchors := []string{}
		entered := map[string]bool{}
		for _, l := range first.adjacent[r.Root] {
			if !entered[l.Target] {
				anchors = append(anchors, l.Target)
				entered[l.Target] = true
			}
		}
		if err := t.chooseAnchoredExpansion(anchors); err != nil {
			return err
		}
		t.placeLinks(anchors)
		// When absent from phase2 this is display-only phase1 data. Parent
		// selection is already settled; an existing phase2 occurrence can expand.
		t.nodes[r.Root] = c.nodes[v.Root.Path]
		t.println(first.nodeLabel(r.Root))
		rows := []graphTreeLine{}
		entered = map[string]bool{}
		for _, l := range first.adjacent[r.Root] {
			row := graphTreeLine{link: l, next: l.Target, direction: 1}
			if entered[l.Target] {
				row.reference = " (shown elsewhere)"
			} else {
				entered[l.Target] = true
				row.children = t.expand(l.Target)
			}
			rows = append(rows, row)
		}
		if len(t.expanded) != len(phaseNodes) || len(t.emitted) != len(phaseLinks) {
			return invalidMemoryConstraintCapture()
		}
		t.writeLines(rows, "")
		t.println(fmt.Sprintf("Root %s: fixed Subject To then Requires scope; Requires complete: %t; bound: %d; frontier: %d", t.label(r.Root), len(phaseFrontier) == 0, stage.Root.RequiresDepth, len(phaseFrontier)))
		for _, p := range stage.Requires.Frontier {
			t.println(fmt.Sprintf("Root %s: more Requires relationships beyond bound %d: %s", t.label(r.Root), stage.Root.RequiresDepth, t.label(graph.CanonicalURL(scope, p))))
		}
		if t.err != nil {
			return t.err
		}
		return w.wrapError("memory")
	})
}
