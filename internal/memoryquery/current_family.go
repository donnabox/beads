package memoryquery

import (
	"context"
	"fmt"
	"slices"

	graph "github.com/steveyegge/beads/graphops"
)

// CurrentNode binds an identity to its checked current Type and revision. It
// deliberately contains no payload: frontier checks need not hydrate neighbors.
// Canonical records are acquired by the storage adapter after node selection.
type CurrentNode struct{ Path, TypeURL, Version string }

// CurrentContext supplies one coherent, authorized current view. Incident is
// complete BOTH-direction incidence, once per Link identity, including unowned
// Links. It refuses rather than truncates above allowance. Current acquires only
// identity/Type/version facts, not historical payload. Storage separately bounds
// all acquisition, including proven nonmembers and canonical owner validation.
// Pins are preserved; CurrentFamily checks selected references against current
// versions and never silently resolves an old pin to a current record.
type CurrentContext interface {
	Current(context.Context, string) (CurrentNode, error)
	Incident(context.Context, string, int) ([]graph.Link, error)
}

// CurrentRoot is one independent traversal slot. Repeated identities with
// different depths remain distinct slots in the answer.
type CurrentRoot struct {
	Path  string
	Depth int
}

// CurrentView is root-local membership, in code-unit order. Depth chooses nodes;
// Links is the induced selected-family graph on those nodes. Frontier contains
// reached nodes with eligible directional neighbors outside this root's set.
type CurrentView struct {
	Root                   CurrentRoot
	Nodes, Links, Frontier []string
}

// CurrentGraph unions completed root views, without re-inducing over their union.
// These are internal operation values, not a wire or CLI result schema.
type CurrentGraph struct {
	Nodes []CurrentNode
	Links []graph.Link
	Views []CurrentView
}

// ErrCurrentContextUnavailable refuses selected pins that this current-only
// operation cannot bind. It is not an empty answer, historical lookup or frontier.
var ErrCurrentContextUnavailable = fmt.Errorf("selected reference is outside the current read context")

// CurrentFamily mechanically adapts PR176 8dbb393c's separate BFS and induced
// edge/frontier passes (cmd/bd/graph_preview_generic_projection.go). It preserves
// authored direction, parallel edges and cycle closure. Each root has its own
// distance map, while output and candidate-work limits accumulate across roots.
// MaxInspected counts each candidate processing visit, including nonmembers and
// repeated visits in different root views, not merely distinct output Links.
// Every error returns zero state. It performs no satisfaction inference.
func CurrentFamily(ctx context.Context, reader CurrentContext, family RequiresFamily, roots []CurrentRoot, direction graph.Direction, limits Limits) (CurrentGraph, error) {
	if ctx == nil {
		return CurrentGraph{}, fmt.Errorf("%w: missing context", graph.ErrValidation)
	}
	if err := ctx.Err(); err != nil {
		return CurrentGraph{}, err
	}
	if reader == nil || family == nil || !direction.Valid() || len(roots) == 0 || len(roots) > 1000 || limits.MaxNodes < 1 || limits.MaxLinks < 0 || limits.MaxInspected < 0 {
		return CurrentGraph{}, fmt.Errorf("%w: invalid current family request", graph.ErrValidation)
	}
	for _, root := range roots {
		if graph.ValidateBeadPath(root.Path) != nil || root.Depth < 0 || root.Depth > 1000 {
			return CurrentGraph{}, fmt.Errorf("%w: invalid current root", graph.ErrValidation)
		}
	}
	nodes := map[string]CurrentNode{}
	links := map[string]graph.Link{}
	feeds := map[string][]graph.Link{}
	selected := map[string]bool{}
	classified := map[string]bool{}
	facts := map[string]CurrentNode{}
	work := 0
	current := func(path string) (CurrentNode, error) {
		if n, ok := facts[path]; ok {
			return n, nil
		}
		n, err := reader.Current(ctx, path)
		if err != nil {
			return CurrentNode{}, err
		}
		if n.Path != path || n.TypeURL == "" || n.Version == "" {
			return CurrentNode{}, fmt.Errorf("%w: invalid current identity", ErrReadContext)
		}
		facts[path] = n
		return n, nil
	}
	admit := func(path string) error {
		if _, ok := nodes[path]; ok {
			return nil
		}
		if len(nodes) >= limits.MaxNodes {
			return fmt.Errorf("%w: current node union", ErrLimitExceeded)
		}
		n, err := current(path)
		if err != nil {
			return err
		}
		nodes[path] = n
		return nil
	}
	incident := func(path string) ([]graph.Link, error) {
		if v, ok := feeds[path]; ok {
			return v, nil
		}
		v, err := reader.Incident(ctx, path, limits.MaxInspected-work)
		if err != nil {
			return nil, err
		}
		if len(v) > limits.MaxInspected-work {
			return nil, fmt.Errorf("%w: incidence acquisition", ErrLimitExceeded)
		}
		seen := map[string]bool{}
		for _, l := range v {
			if l.IsZero() || !l.Source().InScope() || !l.Target().InScope() || (l.Source().Path() != path && l.Target().Path() != path) || seen[l.Path()] {
				return nil, fmt.Errorf("%w: invalid/duplicate current incidence", ErrReadContext)
			}
			seen[l.Path()] = true
		}
		feeds[path] = v
		return v, nil
	}
	match := func(l graph.Link) (bool, error) {
		if classified[l.TypeURL()] {
			return selected[l.TypeURL()], nil
		}
		yes, err := family(ctx, l.TypeURL())
		if err == nil {
			classified[l.TypeURL()] = true
			selected[l.TypeURL()] = yes
		}
		return yes, err
	}
	visit := func() error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if work >= limits.MaxInspected {
			return fmt.Errorf("%w: current traversal candidate work", ErrLimitExceeded)
		}
		work++
		return nil
	}
	bind := func(l graph.Link) error {
		for _, ref := range []graph.Ref{l.Source(), l.Target()} {
			if !ref.Pinned() {
				continue
			}
			n, err := current(ref.Path())
			if err != nil {
				return err
			}
			if ref.Pin() != n.Version {
				return fmt.Errorf("%w: %s", ErrCurrentContextUnavailable, ref.Path())
			}
		}
		return nil
	}
	next := func(id string, l graph.Link) (string, bool) {
		switch direction {
		case graph.DirectionOut:
			return l.Target().Path(), l.Source().Path() == id
		case graph.DirectionIn:
			return l.Source().Path(), l.Target().Path() == id
		default:
			if l.Source().Path() == id {
				return l.Target().Path(), true
			}
			return l.Source().Path(), true
		}
	}
	result := CurrentGraph{Views: make([]CurrentView, 0, len(roots)), Nodes: []CurrentNode{}, Links: []graph.Link{}}
	for _, root := range roots {
		if err := admit(root.Path); err != nil {
			return CurrentGraph{}, err
		}
		distance := map[string]int{root.Path: 0}
		queue := []string{root.Path}
		for cursor := 0; cursor < len(queue); cursor++ {
			if err := ctx.Err(); err != nil {
				return CurrentGraph{}, err
			}
			id := queue[cursor]
			if distance[id] >= root.Depth {
				continue
			}
			feed, err := incident(id)
			if err != nil {
				return CurrentGraph{}, err
			}
			for _, l := range feed {
				if err := visit(); err != nil {
					return CurrentGraph{}, err
				}
				yes, err := match(l)
				if err != nil {
					return CurrentGraph{}, err
				}
				neighbor, eligible := next(id, l)
				if !yes || !eligible {
					continue
				}
				if err := bind(l); err != nil {
					return CurrentGraph{}, err
				}
				if _, seen := distance[neighbor]; !seen {
					if err := admit(neighbor); err != nil {
						return CurrentGraph{}, err
					}
					distance[neighbor] = distance[id] + 1
					queue = append(queue, neighbor)
				}
			}
		}
		view := CurrentView{Root: root, Nodes: []string{}, Links: []string{}, Frontier: []string{}}
		emitted := map[string]graph.Link{}
		for _, id := range queue {
			view.Nodes = append(view.Nodes, id)
			feed, err := incident(id)
			if err != nil {
				return CurrentGraph{}, err
			}
			frontier := false
			for _, l := range feed {
				if err := visit(); err != nil {
					return CurrentGraph{}, err
				}
				yes, err := match(l)
				if err != nil {
					return CurrentGraph{}, err
				}
				if !yes {
					continue
				}
				_, src := distance[l.Source().Path()]
				_, dst := distance[l.Target().Path()]
				neighbor, eligible := next(id, l)
				_, reached := distance[neighbor]
				if (src && dst) || (eligible && !reached) {
					if err := bind(l); err != nil {
						return CurrentGraph{}, err
					}
				}
				if src && dst {
					emitted[l.Path()] = l
				}
				if eligible && !reached {
					frontier = true
				}
			}
			if frontier {
				view.Frontier = append(view.Frontier, id)
			}
		}
		for path, l := range emitted {
			if _, ok := links[path]; !ok && len(links) >= limits.MaxLinks {
				return CurrentGraph{}, fmt.Errorf("%w: current Link union", ErrLimitExceeded)
			}
			links[path] = l
			view.Links = append(view.Links, path)
		}
		sortCurrentPaths(view.Nodes)
		sortCurrentPaths(view.Links)
		sortCurrentPaths(view.Frontier)
		result.Views = append(result.Views, view)
	}
	for _, n := range nodes {
		result.Nodes = append(result.Nodes, n)
	}
	for _, l := range links {
		result.Links = append(result.Links, l)
	}
	slices.SortFunc(result.Nodes, func(a, b CurrentNode) int { return graph.CompareCodeUnits(a.Path, b.Path) })
	slices.SortFunc(result.Links, func(a, b graph.Link) int { return graph.CompareCodeUnits(a.Path(), b.Path()) })
	if err := ctx.Err(); err != nil {
		return CurrentGraph{}, err
	}
	return result, nil
}

func sortCurrentPaths(v []string) { slices.SortFunc(v, graph.CompareCodeUnits) }
