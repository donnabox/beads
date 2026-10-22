package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// A relationship tree presents the selected graph; it does not select another
// graph. Nodes and Links remain authoritative, including induced cross-links.
type graphRelationshipTree struct {
	result           graphGenericResult
	nodes            map[string]graphGenericNode
	adjacent         map[string][]graphGenericLink
	outgoing         map[string][]string
	treeEdges        map[string]string
	expanded         map[string]bool
	emitted          map[string]bool
	w                *graphExportWriter
	limit, remaining int
	err              error
}

type graphTreeLine struct {
	link      graphGenericLink
	next      string
	direction int
	reference string
	children  []graphTreeLine
}

func renderGraphRelationshipTree(w *graphExportWriter, r graphGenericResult, limit int) error {
	t := graphRelationshipTree{result: r, nodes: map[string]graphGenericNode{}, adjacent: map[string][]graphGenericLink{}, outgoing: map[string][]string{}, treeEdges: map[string]string{}, expanded: map[string]bool{}, emitted: map[string]bool{}, w: w, limit: limit, remaining: limit}
	invalid := func() error {
		return fmt.Errorf("%w: incomplete or duplicate selected graph", graphstore.ErrInvalidStore)
	}
	for _, n := range r.Nodes {
		if n.ID == "" || n.Type == "" || t.nodes[n.ID].ID != "" {
			return invalid()
		}
		t.nodes[n.ID] = n
	}
	if (r.Root == "" && !r.All) || (r.Root != "" && t.nodes[r.Root].ID == "") {
		return invalid()
	}
	links := map[string]bool{}
	for _, l := range r.Links {
		if l.ID == "" || l.Type == "" || links[l.ID] || t.nodes[l.Source].ID == "" || t.nodes[l.Target].ID == "" {
			return invalid()
		}
		links[l.ID] = true
		t.outgoing[l.Source] = append(t.outgoing[l.Source], l.Target)
		if r.Direction != "in" {
			t.adjacent[l.Source] = append(t.adjacent[l.Source], l)
		}
		if r.Direction != "out" && (r.Direction == "in" || l.Target != l.Source) {
			t.adjacent[l.Target] = append(t.adjacent[l.Target], l)
		}
	}
	if r.Direction != "in" && r.Direction != "out" && r.Direction != "both" {
		return invalid()
	}
	for _, id := range r.Frontier {
		if t.nodes[id].ID == "" {
			return invalid()
		}
	}
	for id := range t.adjacent {
		sort.Slice(t.adjacent[id], func(i, j int) bool {
			a, b := t.adjacent[id][i], t.adjacent[id][j]
			if a.Type != b.Type {
				return a.Type < b.Type
			}
			x, _ := t.other(a, id)
			y, _ := t.other(b, id)
			if t.nodes[x].Type != t.nodes[y].Type {
				return t.nodes[x].Type < t.nodes[y].Type
			}
			if c := graph.CompareCodeUnits(t.nodes[x].Title, t.nodes[y].Title); c != 0 {
				return c < 0
			}
			if x != y {
				return x < y
			}
			return a.ID < b.ID
		})
	}
	// The requested root comes first. Remaining components (including --all
	// isolates) are a stable forest, not fabricated connections between roots.
	roots := make([]string, 0, len(t.nodes)+1)
	if r.Root != "" {
		roots = append(roots, r.Root)
	}
	ids := make([]string, 0, len(t.nodes))
	for id := range t.nodes {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	roots = append(roots, ids...)
	roots = t.chooseExpansion(roots)
	first := true
	for _, id := range roots {
		if t.stopped() {
			break
		}
		if t.expanded[id] {
			continue
		}
		if !first {
			t.println("")
		}
		first = false
		t.println(t.nodeLabel(id))
		if !t.stopped() {
			t.writeLines(t.expand(id), "")
		}
	}
	if !first {
		t.println("")
	}
	t.println(fmt.Sprintf("%d Beads, %d Links", len(t.nodes), len(links)))
	t.println(fmt.Sprintf("Complete: %t; direction: %s; depth: %d; frontier: %d", r.Complete, r.Direction, r.Depth, len(r.Frontier)))
	for _, id := range r.Frontier {
		if t.stopped() {
			break
		}
		t.println(fmt.Sprintf("More relationships beyond depth %d: %s", r.Depth, t.label(id)))
	}
	if t.err != nil {
		return t.err
	}
	return w.wrapError("graph")
}

func (t *graphRelationshipTree) stopped() bool { return t.err != nil || t.w.err != nil }

// Keep the existing graph writer and its first-error behavior. Enforce the
// result's byte bound before appending each line to the command's private
// buffer; the caller publishes only a complete successful rendering.
func (t *graphRelationshipTree) println(line string) {
	if t.stopped() {
		return
	}
	if len(line)+1 > t.remaining {
		t.err = fmt.Errorf("%w: graph output exceeds %d bytes", graphstore.ErrLimitExceeded, t.limit)
		return
	}
	t.remaining -= len(line) + 1
	t.w.println(line)
}

func (t *graphRelationshipTree) label(value string) string {
	for _, prefix := range []string{"beads/", "links/", "types/"} {
		if strings.HasPrefix(value, t.result.Scope+prefix) {
			value = strings.TrimPrefix(value, t.result.Scope+prefix)
			break
		}
	}
	// Escape data, not tree structure. In particular a title, ID or pin cannot
	// inject terminal controls or an extra relationship line.
	quoted := strconv.QuoteToGraphic(value)
	return quoted[1 : len(quoted)-1]
}

func (t *graphRelationshipTree) nodeLabel(id string) string {
	n := t.nodes[id]
	label := t.label(id) + " [" + t.label(n.Type) + "]"
	if n.Status != "" {
		label += " (" + t.label(n.Status) + ")"
	}
	return label + " " + strconv.QuoteToGraphic(n.Title)
}

func (t *graphRelationshipTree) other(l graphGenericLink, id string) (string, int) {
	if l.Source == id {
		return l.Target, 1
	}
	return l.Source, -1
}

// Choose shortest-hop parents before placing any display-only cross-reference.
// This only arranges the already-selected graph; it acquires or selects nothing.
func (t *graphRelationshipTree) chooseExpansion(candidates []string) []string {
	assigned := map[string]bool{}
	roots := []string{}
	for _, root := range candidates {
		if assigned[root] {
			continue
		}
		roots = append(roots, root)
		assigned[root] = true
		queue := []string{root}
		for cursor := 0; cursor < len(queue); cursor++ {
			id := queue[cursor]
			for _, l := range t.adjacent[id] {
				next, _ := t.other(l, id)
				if assigned[next] {
					continue
				}
				assigned[next] = true
				t.treeEdges[l.ID] = next
				queue = append(queue, next)
			}
		}
	}
	return roots
}

func (t *graphRelationshipTree) expand(id string) []graphTreeLine {
	t.expanded[id] = true
	lines := []graphTreeLine{}
	for _, l := range t.adjacent[id] {
		if t.emitted[l.ID] {
			continue
		}
		t.emitted[l.ID] = true
		next, direction := t.other(l, id)
		line := graphTreeLine{link: l, next: next, direction: direction}
		if t.treeEdges[l.ID] != next {
			line.reference = " (shown elsewhere)"
			if t.closesCycle(l) {
				line.reference = " (cycle)"
			}
		} else {
			line.children = t.expand(next)
		}
		lines = append(lines, line)
	}
	return lines
}

func (t *graphRelationshipTree) writeLines(lines []graphTreeLine, prefix string) {
	for i, line := range lines {
		if t.stopped() {
			return
		}
		branch, continuation := "├── ", "│   "
		if i == len(lines)-1 {
			branch, continuation = "└── ", "    "
		}
		arrow := "→"
		if line.direction < 0 {
			arrow = "←"
		}
		l := line.link
		node := t.label(line.next) + " [" + t.label(t.nodes[line.next].Type) + "]"
		if line.reference == "" {
			node = t.nodeLabel(line.next)
		}
		text := fmt.Sprintf("%s%s%s (%s) %s %s%s", prefix, branch, t.label(l.Type), t.label(l.ID), arrow, node, line.reference)
		if l.SourceRevision != "" || l.TargetRevision != "" {
			text += fmt.Sprintf(" [source revision=%s target revision=%s]", strconv.QuoteToGraphic(l.SourceRevision), strconv.QuoteToGraphic(l.TargetRevision))
		}
		t.println(text)
		t.writeLines(line.children, prefix+continuation)
	}
}

// A tree ancestor (or a future sibling) is not itself proof of a directed
// cycle. A return path through selected authored Links is. This bounded local
// display check neither consults excluded Links nor changes graph membership.
func (t *graphRelationshipTree) closesCycle(link graphGenericLink) bool {
	seen := map[string]bool{link.Target: true}
	queue := []string{link.Target}
	for cursor := 0; cursor < len(queue); cursor++ {
		id := queue[cursor]
		if id == link.Source {
			return true
		}
		for _, next := range t.outgoing[id] {
			if !seen[next] {
				seen[next] = true
				queue = append(queue, next)
			}
		}
	}
	return false
}
