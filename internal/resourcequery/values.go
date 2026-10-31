package resourcequery

import (
	"encoding/json"
	"math/big"
	"strings"

	graph "github.com/steveyegge/beads/graphops"
)

type bindings map[string]*record

func withBinding(env bindings, name string, r *record) bindings {
	out := bindings{}
	for k, v := range env {
		out[k] = v
	}
	out[name] = r
	return out
}

type value struct {
	data     any
	present  bool
	resource *record
}

func (e *evaluation) resource(n *node, env bindings) (*record, error) {
	if err := e.tick(); err != nil {
		return nil, err
	}
	if n.op == "binding" {
		return env[n.name], nil
	}
	r, err := e.resource(n.of, env)
	if err != nil {
		return nil, err
	}
	ref := r.source
	if n.side == "target" {
		ref = r.target
	}
	return e.resolve(ref, graph.KindBead)
}
func (e *evaluation) val(n *node, env bindings) (value, error) {
	if err := e.tick(); err != nil {
		return value{}, err
	}
	if n.op == "literal" {
		return value{data: n.literal, present: true}, nil
	}
	if n.op == "binding" || n.op == "endpoint" {
		r, err := e.resource(n, env)
		return value{resource: r, present: err == nil}, err
	}
	r, err := e.resource(n.of, env)
	if err != nil {
		return value{}, err
	}
	if n.op == "id" {
		return value{data: r.id, present: true}, nil
	}
	if n.op == "rev" {
		return value{data: r.revision, present: true}, nil
	}
	return fieldValue(r.body, n.path, e.tick)
}

func equal(a, b value) bool {
	if a.present != b.present {
		return false
	}
	if !a.present {
		return true
	}
	if a.resource != nil || b.resource != nil {
		return a.resource != nil && b.resource != nil && a.resource.edition() == b.resource.edition()
	}
	aJSON, _ := json.Marshal(a.data)
	bJSON, _ := json.Marshal(b.data)
	ok, _ := graph.JSONEqual(aJSON, bJSON)
	return ok
}
func order(a, b value) (int, bool) {
	if !a.present || !b.present {
		return 0, false
	}
	if left, ok := a.data.(string); ok {
		right, ok := b.data.(string)
		if !ok {
			return 0, false
		}
		return strings.Compare(left, right), true
	}
	left, ok := a.data.(json.Number)
	if !ok {
		return 0, false
	}
	right, ok := b.data.(json.Number)
	if !ok {
		return 0, false
	}
	l, lOK := new(big.Rat).SetString(string(left))
	r, rOK := new(big.Rat).SetString(string(right))
	if !lOK || !rOK {
		return 0, false
	}
	return l.Cmp(r), true
}

// Query retains Resource acquisition and collection quantifiers; all local
// predicates run through the same kernel as the revisionless List selector.
func (e *evaluation) predicate(n *node, env bindings) (bool, error) {
	k := predicateKernel{
		tick: e.tick,
		val:  func(n *node) (value, error) { return e.val(n, env) },
		typeIs: func(n *node) (bool, error) {
			r, err := e.resource(n.of, env)
			if err != nil {
				return false, err
			}
			if n.match == "exact" {
				return r.typ == n.typ, nil
			}
			return r.family[n.typ], nil
		},
		quantified: func(n *node) (bool, error) { return e.quantified(n, env) },
	}
	return k.predicate(n)
}
func (e *evaluation) quantified(n *node, env bindings) (bool, error) {
	answer := n.op == "all"
	_, err := e.stream(n.input, env, func(item *record) (bool, error) {
		result := true
		if n.predicate != nil {
			var err error
			result, err = e.predicate(n.predicate, withBinding(env, n.bind, item))
			if err != nil {
				return false, err
			}
		}
		if n.op == "any" && result {
			answer = true
			return false, nil
		}
		if n.op == "all" && !result {
			answer = false
			return false, nil
		}
		return true, nil
	})
	return answer, err
}
