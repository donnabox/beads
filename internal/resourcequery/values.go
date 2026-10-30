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
	var current any = r.body
	for _, segment := range n.path {
		if err = e.tick(); err != nil {
			return value{}, err
		}
		if segment.indexed {
			array, ok := current.([]any)
			if !ok || segment.index >= len(array) {
				return value{}, nil
			}
			current = array[segment.index]
		} else {
			m, ok := current.(map[string]any)
			if !ok {
				return value{}, nil
			}
			var exists bool
			current, exists = m[segment.key]
			if !exists {
				return value{}, nil
			}
		}
	}
	return value{data: current, present: true}, nil
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
func (e *evaluation) predicate(n *node, env bindings) (bool, error) {
	if err := e.tick(); err != nil {
		return false, err
	}
	switch n.op {
	case "bool":
		return n.boolean, nil
	case "and", "or":
		for _, child := range n.operands {
			result, err := e.predicate(child, env)
			if err != nil {
				return false, err
			}
			if n.op == "and" && !result {
				return false, nil
			}
			if n.op == "or" && result {
				return true, nil
			}
		}
		return n.op == "and", nil
	case "not":
		result, err := e.predicate(n.predicate, env)
		return !result, err
	case "typeIs":
		r, err := e.resource(n.of, env)
		if err != nil {
			return false, err
		}
		if n.match == "exact" {
			return r.typ == n.typ, nil
		}
		return r.family[n.typ], nil
	case "present":
		v, err := e.val(n.value, env)
		return v.present, err
	case "compare":
		a, err := e.val(n.left, env)
		if err != nil {
			return false, err
		}
		b, err := e.val(n.right, env)
		if err != nil {
			return false, err
		}
		if n.operator == "eq" {
			return equal(a, b), nil
		}
		if n.operator == "ne" {
			return !equal(a, b), nil
		}
		c, ok := order(a, b)
		if !ok {
			return false, nil
		}
		switch n.operator {
		case "lt":
			return c < 0, nil
		case "le":
			return c <= 0, nil
		case "gt":
			return c > 0, nil
		case "ge":
			return c >= 0, nil
		}
	case "contains":
		a, err := e.val(n.text, env)
		if err != nil {
			return false, err
		}
		b, err := e.val(n.substring, env)
		if err != nil {
			return false, err
		}
		left, lOK := a.data.(string)
		right, rOK := b.data.(string)
		return a.present && b.present && lOK && rOK && strings.Contains(left, right), nil
	case "member":
		a, err := e.val(n.value, env)
		if err != nil {
			return false, err
		}
		b, err := e.val(n.array, env)
		if err != nil {
			return false, err
		}
		array, ok := b.data.([]any)
		if !a.present || !b.present || !ok {
			return false, nil
		}
		for _, item := range array {
			if err = e.tick(); err != nil {
				return false, err
			}
			if equal(a, value{data: item, present: true}) {
				return true, nil
			}
		}
		return false, nil
	case "any", "all":
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
	return false, fail("invalid-predicate", n.at, "unadmitted predicate", ErrInvalid)
}
