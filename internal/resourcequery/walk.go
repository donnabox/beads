package resourcequery

import graph "github.com/steveyegge/beads/graphops"

func applicable(ref reference, r *record) bool {
	return ref.uri == r.id && (ref.pin == "" || ref.pin == r.revision)
}

type frontier struct {
	bead     *record
	distance int
}

func (e *evaluation) candidates(id, direction string) []*record {
	if direction == "outgoing" {
		return e.outgoing[id]
	}
	if direction == "incoming" {
		return e.incoming[id]
	}
	return e.connected[id]
}

func (e *evaluation) walk(n *node, env bindings, emit visitor) (bool, error) {
	var roots []*record
	if n.roots.op == "binding" || n.roots.op == "endpoint" {
		r, err := e.resource(n.roots, env)
		if err != nil {
			return false, err
		}
		roots = []*record{r}
	} else {
		var err error
		roots, err = e.collection(n.roots, env)
		if err != nil {
			return false, err
		}
	}
	roots = distinct(roots)
	e.walkRoots[n] = roots
	queue := make([]frontier, 0, len(roots))
	distance := map[Edition]int{}
	for _, root := range roots {
		distance[root.edition()] = 0
		queue = append(queue, frontier{root, 0})
	}
	collected := map[Edition]bool{}
	boundary := n.depth == 0 && len(roots) > 0
	for head := 0; head < len(queue); head++ {
		if err := e.tick(); err != nil {
			return false, err
		}
		current := queue[head]
		if n.depth >= 0 && current.distance >= n.depth {
			continue
		}
		for _, link := range e.candidates(current.bead.id, n.direction) {
			if err := e.tick(); err != nil {
				return false, err
			}
			// Match stored References to this selected current edition. A pin for
			// another edition is inapplicable, not an invitation to substitute it.
			var opposite reference
			matched := false
			if n.direction != "incoming" && applicable(link.source, current.bead) {
				opposite = link.target
				matched = true
			}
			if !matched && n.direction != "outgoing" && applicable(link.target, current.bead) {
				opposite = link.source
				matched = true
			}
			if !matched {
				continue
			}
			if n.predicate != nil {
				yes, err := e.predicate(n.predicate, withBinding(env, n.bind, link))
				if err != nil {
					return false, err
				}
				if !yes {
					continue
				}
			}
			if !collected[link.edition()] {
				collected[link.edition()] = true
				more, err := emit(link)
				if err != nil {
					return false, err
				}
				if !more {
					e.trace(n, "proof", "unknown")
					return false, nil
				}
			}
			nextDistance := current.distance + 1
			if n.depth >= 0 && nextDistance >= n.depth {
				boundary = true
				continue
			}
			// A Link-only boundary needs no endpoint body. Actual expansion does;
			// a missing/retained/foreign body refuses the whole result.
			neighbor, err := e.resolve(opposite, graph.KindBead)
			if err != nil {
				return false, err
			}
			old, seen := distance[neighbor.edition()]
			if !seen || nextDistance < old {
				distance[neighbor.edition()] = nextDistance
				queue = append(queue, frontier{neighbor, nextDistance})
				if err = e.bounded(len(distance)); err != nil {
					return false, err
				}
			}
		}
	}
	// Boundary inspection is deliberately not performed for an incidental
	// diagnostic: unknown does not claim a truncated result or global closure.
	if boundary {
		e.trace(n, "depth", "unknown")
	} else {
		e.trace(n, "exhausted", "complete")
	}
	return true, nil
}
func (e *evaluation) trace(n *node, completion, closure string) {
	depth := any(n.depth)
	if n.depth < 0 {
		depth = "unbounded"
	}
	trace := Traversal{At: n.at, Depth: depth, Completion: completion, Closure: closure}
	if prior, exists := e.traces[n.at]; exists {
		if prior.Completion == "proof" || prior.Completion == "depth" && trace.Completion == "exhausted" {
			trace.Completion = prior.Completion
		}
		if prior.Closure == "not-complete" || trace.Closure == "not-complete" {
			trace.Closure = "not-complete"
		} else if prior.Closure != "complete" || trace.Closure != "complete" {
			trace.Closure = "unknown"
		}
	}
	e.traces[n.at] = trace
}
