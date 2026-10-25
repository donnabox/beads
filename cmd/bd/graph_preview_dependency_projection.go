package main

import (
	"fmt"
	"sort"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// graphDependencyProjection is a single checked current snapshot, narrowed to
// native Issues and their mapped blocking Dependencies. Its IDs remain canonical
// graph URLs; the familiar graph check renders native backing IDs only at its
// output boundary.
type graphDependencyProjection struct {
	issues map[string]graphstore.IssueRecord
	links  []graphstore.LinkRecord
	down   map[string][]string
	up     map[string][]string
}

func projectGraphDependencies(snapshot graphstore.Snapshot, scope string) (graphDependencyProjection, error) {
	blockingType, err := graphstore.NativeDependencyTypeURL(scope, "blocks")
	if err != nil {
		return graphDependencyProjection{}, err
	}
	g := graphDependencyProjection{
		issues: map[string]graphstore.IssueRecord{}, links: []graphstore.LinkRecord{},
		down: map[string][]string{}, up: map[string][]string{},
	}
	for _, raw := range snapshot.Records {
		if issue, ok := raw.(graphstore.IssueRecord); ok {
			g.issues[issue.ID] = issue
		}
	}
	for _, raw := range snapshot.Records {
		link, ok := raw.(graphstore.LinkRecord)
		if !ok || link.Type != blockingType || link.Properties["dep_type"] != "blocks" {
			continue
		}
		if _, ok := g.issues[link.Source]; !ok {
			return graphDependencyProjection{}, fmt.Errorf("%w: blocking source has no mapped Issue", graphstore.ErrInvalidStore)
		}
		if _, ok := g.issues[link.Target]; !ok {
			return graphDependencyProjection{}, fmt.Errorf("%w: blocking target has no mapped Issue", graphstore.ErrInvalidStore)
		}
		g.links = append(g.links, link)
		g.down[link.Source] = append(g.down[link.Source], link.Target)
		g.up[link.Target] = append(g.up[link.Target], link.Source)
	}
	sort.Slice(g.links, func(i, j int) bool { return g.links[i].ID < g.links[j].ID })
	for _, edges := range []map[string][]string{g.down, g.up} {
		for id := range edges {
			sort.Strings(edges[id])
		}
	}
	return g, nil
}

func (g graphDependencyProjection) hasIssue(id string) bool {
	_, ok := g.issues[id]
	return ok
}

func (g graphDependencyProjection) component(root string) ([]graphstore.IssueRecord, []graphstore.LinkRecord) {
	seen := map[string]bool{root: true}
	queue := []string{root}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		for _, neighbor := range append(append([]string{}, g.down[id]...), g.up[id]...) {
			if !seen[neighbor] {
				seen[neighbor] = true
				queue = append(queue, neighbor)
			}
		}
	}
	issues := make([]graphstore.IssueRecord, 0, len(seen))
	for id := range seen {
		issues = append(issues, g.issues[id])
	}
	sort.Slice(issues, func(i, j int) bool { return issues[i].ID < issues[j].ID })
	links := make([]graphstore.LinkRecord, 0)
	for _, link := range g.links {
		if seen[link.Source] && seen[link.Target] {
			links = append(links, link)
		}
	}
	return issues, links
}

type graphDependencyTreeNode struct {
	Issue    graphstore.IssueRecord `json:"issue"`
	Depth    int                    `json:"depth"`
	ParentID string                 `json:"parent_id"`
	EdgeType string                 `json:"edge_type,omitempty"`
}

func (g graphDependencyProjection) tree(root, direction string, maxDepth int) []graphDependencyTreeNode {
	adjacent := g.down
	if direction == "up" {
		adjacent = g.up
	}
	nodes := make([]graphDependencyTreeNode, 0)
	visited := map[string]bool{}
	var walk func(string, string, int)
	walk = func(id, parent string, depth int) {
		if depth >= maxDepth || visited[id] {
			return
		}
		visited[id] = true
		node := graphDependencyTreeNode{Issue: g.issues[id], Depth: depth, ParentID: parent}
		if parent != "" {
			node.EdgeType = "blocks"
		}
		nodes = append(nodes, node)
		for _, next := range adjacent[id] {
			walk(next, id, depth+1)
		}
	}
	walk(root, "", 0)
	return nodes
}

// cycles follows the native blocking check's DFS/back-edge rule. Sorting only
// stabilizes traversal; this is not the wider tracks or scheduling-edge check.
func (g graphDependencyProjection) cycles() [][]string {
	ids := make([]string, 0, len(g.issues))
	for id := range g.issues {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	visited, inPath := map[string]bool{}, map[string]int{}
	path := []string{}
	cycles := make([][]string, 0)
	var walk func(string)
	walk = func(id string) {
		visited[id] = true
		inPath[id] = len(path)
		path = append(path, id)
		for _, next := range g.down[id] {
			if !visited[next] {
				walk(next)
			} else if start, found := inPath[next]; found {
				cycles = append(cycles, append([]string{}, path[start:]...))
			}
		}
		path = path[:len(path)-1]
		delete(inPath, id)
	}
	for _, id := range ids {
		if !visited[id] {
			walk(id)
		}
	}
	return cycles
}

// graph check is an ordinary Issue command: its cycle members are native Issue
// IDs. Keep cycles() canonical for graph-specific projections and translate only
// at the familiar command boundary.
func (g graphDependencyProjection) nativeCycleIDs() ([][]string, error) {
	cycles := g.cycles()
	for i := range cycles {
		for j, url := range cycles[i] {
			issue, ok := g.issues[url]
			if !ok || issue.Properties == nil || issue.Properties.ID == "" {
				return nil, fmt.Errorf("%w: blocking cycle member has no native Issue ID", graphstore.ErrInvalidStore)
			}
			cycles[i][j] = issue.Properties.ID
		}
	}
	return cycles, nil
}
