package resourcequery

import (
	"encoding/json"
	"sort"

	graph "github.com/steveyegge/beads/graphops"
)

func (e *evaluation) collection(n *node, env bindings) ([]*record, error) {
	out := make([]*record, 0)
	_, err := e.stream(n, env, func(r *record) (bool, error) { out = append(out, r); return true, nil })
	return out, err
}

// A false visitor continuation is a definitive proof, never an evaluator cutoff.
type visitor func(*record) (bool, error)

func (e *evaluation) stream(n *node, env bindings, yield visitor) (bool, error) {
	if err := e.tick(); err != nil {
		return false, err
	}
	seen := map[Edition]bool{}
	emit := func(r *record) (bool, error) {
		if err := e.tick(); err != nil {
			return false, err
		}
		if seen[r.edition()] {
			return true, nil
		}
		seen[r.edition()] = true
		if err := e.bounded(len(seen)); err != nil {
			return false, err
		}
		return yield(r)
	}
	switch n.op {
	case "scan":
		input := e.beads
		if n.collection == "links" {
			input = e.links
		}
		for _, r := range input {
			more, err := emit(r)
			if err != nil || !more {
				return more, err
			}
		}
	case "refs":
		for _, ref := range n.refs {
			r, err := e.resolve(ref, collectionKind(n.collection))
			if err != nil {
				return false, err
			}
			more, err := emit(r)
			if err != nil || !more {
				return more, err
			}
		}
	case "filter":
		return e.stream(n.input, env, func(item *record) (bool, error) {
			yes, err := e.predicate(n.predicate, withBinding(env, n.bind, item))
			if err != nil {
				return false, err
			}
			if yes {
				return emit(item)
			}
			return true, nil
		})
	case "endpoints":
		return e.stream(n.input, env, func(item *record) (bool, error) {
			ref := item.source
			if n.side == "target" {
				ref = item.target
			}
			r, err := e.resolve(ref, graph.KindBead)
			if err != nil {
				return false, err
			}
			return emit(r)
		})
	case "walk":
		return e.walk(n, env, emit)
	default:
		return false, fail("invalid-collection", n.at, "unadmitted collection", ErrInvalid)
	}
	return true, nil
}
func rawRecords(records []*record) []json.RawMessage {
	out := make([]json.RawMessage, 0, len(records))
	for _, r := range records {
		out = append(out, append(json.RawMessage(nil), r.raw...))
	}
	return out
}
func (e *evaluation) execute() (Response, error) {
	items, err := e.collection(e.request.root, bindings{})
	if err != nil {
		return Response{}, err
	}
	sortRecords(items)
	var result any
	switch e.request.form {
	case "count":
		result = CountResult{Form: "count", ResourceKind: e.request.kind, Count: len(items)}
	case "resources":
		result = ResourcesResult{Form: "resources", ResourceKind: e.request.kind, Items: rawRecords(items), Next: nil}
	case "graph":
		g, err := e.graph(items)
		if err != nil {
			return Response{}, err
		}
		result = g
	}
	traces := make([]Traversal, 0, len(e.traces))
	for _, trace := range e.traces {
		traces = append(traces, trace)
	}
	sort.Slice(traces, func(i, j int) bool { return traces[i].At < traces[j].At })
	if err = e.ctx.Err(); err != nil {
		return Response{}, err
	}
	return Response{Result: result, Evaluation: Evaluation{Scope: "response", Complete: true, Traversals: traces}}, nil
}
func (e *evaluation) graph(links []*record) (GraphResult, error) {
	roots := append([]*record(nil), e.walkRoots[e.request.root]...)
	sortRecords(roots)
	beads := append([]*record(nil), roots...)
	g := GraphResult{Form: "graph", Roots: make([]Edition, 0, len(roots)), Links: rawRecords(links), Endpoints: make([]EndpointMapping, 0, len(links))}
	for _, root := range roots {
		g.Roots = append(g.Roots, root.edition())
	}
	for _, link := range links {
		source, err := e.resolve(link.source, graph.KindBead)
		if err != nil {
			return GraphResult{}, err
		}
		target, err := e.resolve(link.target, graph.KindBead)
		if err != nil {
			return GraphResult{}, err
		}
		beads = append(beads, source, target)
		g.Endpoints = append(g.Endpoints, EndpointMapping{Link: link.edition(), Source: source.edition(), Target: target.edition()})
	}
	beads = distinct(beads)
	sortRecords(beads)
	if err := e.bounded(len(beads) + len(links)); err != nil {
		return GraphResult{}, err
	}
	g.Beads = rawRecords(beads)
	return g, nil
}
