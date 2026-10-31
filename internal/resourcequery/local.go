package resourcequery

import "strings"

// predicateKernel operates on admitted nodes and local values. Resource
// acquisition, collection evaluation and edition comparison stay with Query.
type predicateKernel struct {
	tick       func() error
	val        func(*node) (value, error)
	typeIs     func(*node) (bool, error)
	quantified func(*node) (bool, error)
}

func matchesType(typ string, n *node, families map[string]map[string]bool) bool {
	if n.match == "exact" {
		return typ == n.typ
	}
	return families[typ][n.typ]
}

func fieldValue(body map[string]any, path []segment, tick func() error) (value, error) {
	var current any = body
	for _, segment := range path {
		if err := tick(); err != nil {
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
func (k predicateKernel) predicate(n *node) (bool, error) {
	if err := k.tick(); err != nil {
		return false, err
	}
	switch n.op {
	case "bool":
		return n.boolean, nil
	case "and", "or":
		for _, child := range n.operands {
			result, err := k.predicate(child)
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
		result, err := k.predicate(n.predicate)
		return !result, err
	case "typeIs":
		return k.typeIs(n)
	case "present":
		v, err := k.val(n.value)
		return v.present, err
	case "compare":
		a, err := k.val(n.left)
		if err != nil {
			return false, err
		}
		b, err := k.val(n.right)
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
		a, err := k.val(n.text)
		if err != nil {
			return false, err
		}
		b, err := k.val(n.substring)
		if err != nil {
			return false, err
		}
		left, lOK := a.data.(string)
		right, rOK := b.data.(string)
		return a.present && b.present && lOK && rOK && strings.Contains(left, right), nil
	case "member":
		a, err := k.val(n.value)
		if err != nil {
			return false, err
		}
		b, err := k.val(n.array)
		if err != nil {
			return false, err
		}
		array, ok := b.data.([]any)
		if !a.present || !b.present || !ok {
			return false, nil
		}
		for _, item := range array {
			if err = k.tick(); err != nil {
				return false, err
			}
			if equal(a, value{data: item, present: true}) {
				return true, nil
			}
		}
		return false, nil
	case "any", "all":
		if k.quantified != nil {
			return k.quantified(n)
		}

	}
	return false, fail("invalid-predicate", n.at, "unadmitted predicate", ErrInvalid)
}
