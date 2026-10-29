package resourcequery

import graph "github.com/steveyegge/beads/graphops"

type expressionType struct {
	category string
	kind     graph.ResourceKind
}
type environment map[string]graph.ResourceKind

func collectionKind(name string) graph.ResourceKind {
	if name == "beads" {
		return graph.KindBead
	}
	return graph.KindLink
}
func bind(env environment, name string, kind graph.ResourceKind) environment {
	copyEnv := environment{}
	for k, v := range env {
		copyEnv[k] = v
	}
	copyEnv[name] = kind
	return copyEnv
}

// Admit checks the entire request, including branches that may short-circuit.
// It does not consult a dataset, interpret Type authority or perform acquisition.
func Admit(r *Request, limits Limits) error {
	if r == nil {
		return fail("invalid-request", "", "Decode must construct the request", ErrInvalid)
	}
	copyRequest := *r
	return admit(&copyRequest, limits)
}
func admit(r *Request, limits Limits) error {
	l, err := normalizeLimits(limits)
	if err != nil {
		return err
	}
	if r == nil || r.root == nil || len(r.raw) == 0 {
		return fail("invalid-request", "", "Decode must construct the request", ErrInvalid)
	}
	if len(r.raw) > l.RequestBytes || r.nodes > l.Nodes || r.nesting > l.Nesting {
		return fail("request-limit", "/query", "request exceeds admitted limits", ErrLimit)
	}
	r.families = nil
	r.references = nil
	t, err := r.infer(r.root, environment{})
	if err != nil {
		return err
	}
	if t.category != "collection" {
		return fail("invalid-selection", r.root.at, "selection must be a Resource collection", ErrInvalid)
	}
	r.kind = t.kind
	if r.profile == "graph" {
		w := r.root
		if r.form != "graph" || w.op != "walk" || w.roots.op != "refs" || w.roots.collection != "beads" {
			return fail("profile-mismatch", w.at, "graph profile requires explicit Bead roots and direct walk", ErrInvalid)
		}
		if w.predicate != nil && !(w.predicate.op == "bool" && w.predicate.boolean) && (w.predicate.op != "typeIs" || w.predicate.of.op != "binding" || w.predicate.of.name != w.bind) {
			return fail("profile-mismatch", w.at, "graph profile permits only one Link Type restriction", ErrInvalid)
		}
	}
	if r.profile == "list" {
		if r.form == "graph" {
			return fail("profile-mismatch", "/query/result", "List cannot return a graph", ErrInvalid)
		}
		n := r.root
		if n.op == "filter" {
			if n.input.op != "scan" {
				return fail("profile-mismatch", n.at, "List filters one current scan", ErrInvalid)
			}
			if err = r.localPredicate(n.predicate, n.bind); err != nil {
				return err
			}
			n = n.input
		}
		if n.op != "scan" {
			return fail("profile-mismatch", n.at, "List requires a current inventory", ErrInvalid)
		}
	}
	if r.form == "graph" && r.root.op != "walk" {
		return fail("graph-shape-unsupported", r.root.at, "graph requires a direct walk", ErrUnsupported)
	}
	return nil
}

func (r *Request) infer(n *node, env environment) (expressionType, error) {
	invalid := func() (expressionType, error) {
		return expressionType{}, fail("invalid-kind", n.at, "invalid operand category or Resource kind", ErrInvalid)
	}
	need := func(child *node, category string) (expressionType, error) {
		if child == nil {
			return expressionType{}, fail("invalid-node", n.at, "missing AST operand", ErrInvalid)
		}
		t, e := r.infer(child, env)
		if e == nil && t.category != category {
			return expressionType{}, fail("invalid-kind", child.at, "invalid operand category", ErrInvalid)
		}
		return t, e
	}
	switch n.op {
	case "scan", "refs":
		if n.op == "refs" {
			for _, ref := range n.refs {
				if ref.pin != "" {
					return expressionType{}, fail("exact-acquisition-unsupported", n.at, "exact Resource acquisition is not supported", ErrUnsupported)
				}
			}
			r.references = append(r.references, n)
		}
		return expressionType{"collection", collectionKind(n.collection)}, nil
	case "binding":
		kind, ok := env[n.name]
		if !ok {
			return expressionType{}, fail("unbound-name", n.at, "unbound lexical name", ErrInvalid)
		}
		return expressionType{"resource", kind}, nil
	case "endpoint":
		t, e := need(n.of, "resource")
		if e != nil {
			return t, e
		}
		if t.kind != graph.KindLink {
			return invalid()
		}
		return expressionType{"resource", graph.KindBead}, nil
	case "endpoints":
		t, e := need(n.input, "collection")
		if e != nil {
			return t, e
		}
		if t.kind != graph.KindLink {
			return invalid()
		}
		return expressionType{"collection", graph.KindBead}, nil
	case "filter":
		t, e := need(n.input, "collection")
		if e != nil {
			return t, e
		}
		p, e := r.infer(n.predicate, bind(env, n.bind, t.kind))
		if e != nil {
			return p, e
		}
		if p.category != "predicate" {
			return invalid()
		}
		return t, nil
	case "walk":
		t, e := r.infer(n.roots, env)
		if e != nil {
			return t, e
		}
		if (t.category != "collection" && t.category != "resource") || t.kind != graph.KindBead {
			return invalid()
		}
		if n.predicate != nil {
			p, e := r.infer(n.predicate, bind(env, n.bind, graph.KindLink))
			if e != nil {
				return p, e
			}
			if p.category != "predicate" {
				return invalid()
			}
		}
		return expressionType{"collection", graph.KindLink}, nil
	case "literal":
		return expressionType{category: "value"}, nil
	case "field", "id", "rev":
		_, e := need(n.of, "resource")
		return expressionType{category: "value"}, e
	case "bool":
		return expressionType{category: "predicate"}, nil
	case "and", "or":
		for _, p := range n.operands {
			if _, e := need(p, "predicate"); e != nil {
				return expressionType{}, e
			}
		}
	case "not":
		if _, e := need(n.predicate, "predicate"); e != nil {
			return expressionType{}, e
		}
	case "present":
		if n.value.op != "field" {
			return invalid()
		}
		if _, e := need(n.value, "value"); e != nil {
			return expressionType{}, e
		}
	case "contains":
		if _, e := need(n.text, "value"); e != nil {
			return expressionType{}, e
		}
		if _, e := need(n.substring, "value"); e != nil {
			return expressionType{}, e
		}
	case "member":
		if _, e := need(n.value, "value"); e != nil {
			return expressionType{}, e
		}
		if _, e := need(n.array, "value"); e != nil {
			return expressionType{}, e
		}
	case "typeIs":
		t, e := need(n.of, "resource")
		if e != nil {
			return t, e
		}
		if n.match == "family" {
			r.families = append(r.families, familyUse{n.typ, t.kind, n.at})
		}
	case "compare":
		a, e := r.infer(n.left, env)
		if e != nil {
			return a, e
		}
		b, e := r.infer(n.right, env)
		if e != nil {
			return b, e
		}
		if a.category != b.category || a.category != "resource" && a.category != "value" {
			return invalid()
		}
		if a.category == "resource" && n.operator != "eq" && n.operator != "ne" {
			return invalid()
		}
	case "any", "all":
		t, e := need(n.input, "collection")
		if e != nil {
			return t, e
		}
		if n.predicate != nil {
			p, e := r.infer(n.predicate, bind(env, n.bind, t.kind))
			if e != nil {
				return p, e
			}
			if p.category != "predicate" {
				return invalid()
			}
		}
	default:
		return invalid()
	}
	return expressionType{category: "predicate"}, nil
}

func localValue(n *node, name string) bool {
	return n.op == "literal" || (n.op == "field" || n.op == "id" || n.op == "rev") && n.of.op == "binding" && n.of.name == name
}
func (r *Request) localPredicate(n *node, name string) error {
	bad := func() error {
		return fail("profile-mismatch", n.at, "predicate is outside candidate-local List", ErrInvalid)
	}
	switch n.op {
	case "bool":
		return nil
	case "and", "or":
		for _, c := range n.operands {
			if e := r.localPredicate(c, name); e != nil {
				return e
			}
		}
	case "not":
		return r.localPredicate(n.predicate, name)
	case "typeIs":
		if n.of.op != "binding" || n.of.name != name {
			return bad()
		}
	case "present":
		if !localValue(n.value, name) {
			return bad()
		}
	case "compare":
		if !localValue(n.left, name) || !localValue(n.right, name) || n.left.op != "literal" && n.right.op != "literal" {
			return bad()
		}
	case "contains":
		if !localValue(n.text, name) || n.substring.op != "literal" {
			return bad()
		}
		if _, ok := n.substring.literal.(string); !ok {
			return bad()
		}
	case "member":
		if !localValue(n.value, name) || !localValue(n.array, name) || n.value.op != "literal" && n.array.op != "literal" {
			return bad()
		}
	default:
		return bad()
	}
	return nil
}
