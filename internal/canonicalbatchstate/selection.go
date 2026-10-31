package canonicalbatchstate

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strings"

	"github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/canonicalbatch"
)

// These are fail-closed local execution budgets, not semantic result truncation.
const (
	MaxSelectionWork  = 1_000_000
	MaxSelectionBytes = 32 << 20
	MaxSelectedIDs    = 10_000
)

// The receiver does not yet contain the standalone resourcequery module. This
// evaluator deliberately implements only the shared AST's candidate-local List
// subset, plus the selected transaction refs source. It never interprets another
// language or claims full Query, graph, endpoint acquisition or public editions.
type selection struct {
	kind          graphops.ResourceKind
	scan          bool
	references    []string
	sourceBinding string
	bind          string
	predicate     *localNode
	types         []selectionTypeUse
}

type selectionTypeUse struct{ uri, match string }
type localNode struct {
	op       string
	boolean  bool
	operator string
	name     string
	path     []any
	literal  json.RawMessage
	match    string
	typ      string
	children []*localNode
}

type selectionParser struct {
	nodes int
	types []selectionTypeUse
}

func admitSelection(m canonicalbatch.Member) (*selection, error) {
	// Reuse the compiler's retained schema and strict JSON admission. Schema
	// validation includes every branch before our narrower execution admission.
	raw, err := json.Marshal(struct {
		Operation string          `json:"operation"`
		Name      string          `json:"name"`
		Query     json.RawMessage `json:"query"`
	}{"select", m.Name, m.Query})
	if err != nil {
		return nil, fmt.Errorf("%w: selection envelope", ErrInvalid)
	}
	checked, _, err := canonicalbatch.CompileMember(raw, "")
	if err != nil {
		if errors.Is(err, canonicalbatch.ErrUnsupported) {
			return nil, fmt.Errorf("%w: selection grammar: %v", ErrUnsupported, err)
		}
		return nil, fmt.Errorf("%w: selection grammar: %v", ErrInvalid, err)
	}
	if len(checked.Query) > 262144 {
		return nil, fmt.Errorf("%w: selection request bytes", ErrLimit)
	}
	f := nodeFields(checked.Query)
	if text(f, "profile") != "list" {
		return nil, fmt.Errorf("%w: staged receiver supports local List, not full Query", ErrUnsupported)
	}
	s := &selection{}
	p := selectionParser{}
	root := nodeFields(f["selection"])
	if text(root, "op") == "filter" {
		s.bind = text(root, "bind")
		s.predicate, err = p.predicate(root["predicate"], s.bind, 0)
		if err != nil {
			return nil, err
		}
		root = nodeFields(root["input"])
	}
	s.kind = graphops.KindBead
	if text(root, "collection") == "links" {
		s.kind = graphops.KindLink
	}
	switch text(root, "op") {
	case "scan":
		s.scan = true
	case "refs":
		refs := bytes.TrimSpace(root["references"])
		if len(refs) > 0 && refs[0] == '{' {
			s.sourceBinding = text(nodeFields(refs), "binding")
		} else {
			var entries []json.RawMessage
			if json.Unmarshal(refs, &entries) != nil {
				return nil, ErrInvalid
			}
			for _, entry := range entries {
				var uri string
				if json.Unmarshal(entry, &uri) != nil {
					return nil, fmt.Errorf("%w: staged selection has no historical acquisition", ErrUnsupported)
				}
				s.references = append(s.references, uri)
			}
		}
	default:
		return nil, fmt.Errorf("%w: List source requires scan or current refs", ErrUnsupported)
	}
	if m.Kind != "" && m.Kind != s.kind {
		return nil, ErrKind
	}
	s.types = p.types
	return s, nil
}

func nodeFields(raw json.RawMessage) map[string]json.RawMessage {
	var fields map[string]json.RawMessage
	_ = json.Unmarshal(raw, &fields)
	return fields
}

func text(fields map[string]json.RawMessage, key string) string {
	var value string
	_ = json.Unmarshal(fields[key], &value)
	return value
}

func (p *selectionParser) tick(depth int) error {
	p.nodes++
	if p.nodes > 4096 || depth > 64 {
		return fmt.Errorf("%w: selection AST depth/nodes", ErrLimit)
	}
	return nil
}

func (p *selectionParser) predicate(raw json.RawMessage, bind string, depth int) (*localNode, error) {
	if err := p.tick(depth); err != nil {
		return nil, err
	}
	if bytes.Equal(raw, []byte("true")) || bytes.Equal(raw, []byte("false")) {
		return &localNode{op: "bool", boolean: bytes.Equal(raw, []byte("true"))}, nil
	}
	f := nodeFields(raw)
	op := text(f, "op")
	n := &localNode{op: op, operator: text(f, "operator")}
	child := func(raw json.RawMessage, value bool) (*localNode, error) {
		if value {
			return p.value(raw, bind, depth+1)
		}
		return p.predicate(raw, bind, depth+1)
	}
	appendChild := func(raw json.RawMessage, value bool) error {
		c, err := child(raw, value)
		if err != nil {
			return err
		}
		n.children = append(n.children, c)
		return nil
	}
	switch op {
	case "and", "or":
		var entries []json.RawMessage
		if json.Unmarshal(f["operands"], &entries) != nil {
			return nil, ErrInvalid
		}
		for _, entry := range entries {
			if err := appendChild(entry, false); err != nil {
				return nil, err
			}
		}
	case "not":
		if err := appendChild(f["predicate"], false); err != nil {
			return nil, err
		}
	case "compare":
		if err := appendChild(f["left"], true); err != nil {
			return nil, err
		}
		if err := appendChild(f["right"], true); err != nil {
			return nil, err
		}
		if n.children[0].op != "literal" && n.children[1].op != "literal" {
			return nil, fmt.Errorf("%w: List field-to-field comparison", ErrUnsupported)
		}
	case "present":
		if err := appendChild(f["value"], true); err != nil {
			return nil, err
		}
	case "contains":
		if err := appendChild(f["text"], true); err != nil {
			return nil, err
		}
		if err := appendChild(f["substring"], true); err != nil {
			return nil, err
		}
		var substring string
		if n.children[1].op != "literal" || json.Unmarshal(n.children[1].literal, &substring) != nil {
			return nil, fmt.Errorf("%w: List contains needs a literal substring", ErrUnsupported)
		}
	case "member":
		if err := appendChild(f["value"], true); err != nil {
			return nil, err
		}
		if err := appendChild(f["array"], true); err != nil {
			return nil, err
		}
		if n.children[0].op != "literal" && n.children[1].op != "literal" {
			return nil, fmt.Errorf("%w: List field-to-field membership", ErrUnsupported)
		}
	case "typeIs":
		if err := localBinding(f["of"], bind); err != nil {
			return nil, err
		}
		n.typ = text(f, "type")
		n.match = text(f, "match")
		p.types = append(p.types, selectionTypeUse{n.typ, n.match})
	default:
		return nil, fmt.Errorf("%w: List predicate %s", ErrUnsupported, op)
	}
	return n, nil
}

func (p *selectionParser) value(raw json.RawMessage, bind string, depth int) (*localNode, error) {
	if err := p.tick(depth); err != nil {
		return nil, err
	}
	f := nodeFields(raw)
	op := text(f, "op")
	n := &localNode{op: op}
	if op == "literal" {
		n.literal = bytes.Clone(f["value"])
		return n, nil
	}
	if op != "field" && op != "id" {
		return nil, fmt.Errorf("%w: unpublished revision or nonlocal Resource expression", ErrUnsupported)
	}
	if err := localBinding(f["of"], bind); err != nil {
		return nil, err
	}
	if op == "field" {
		decoder := json.NewDecoder(bytes.NewReader(f["path"]))
		decoder.UseNumber()
		if err := decoder.Decode(&n.path); err != nil {
			return nil, ErrInvalid
		}
		first, _ := n.path[0].(string)
		switch first {
		case "id", "type", "source", "target", "properties", "metadata":
		default:
			return nil, fmt.Errorf("%w: field %q has no faithful staged representation", ErrUnsupported, first)
		}
	}
	return n, nil
}

func localBinding(raw json.RawMessage, bind string) error {
	f := nodeFields(raw)
	if text(f, "op") != "binding" {
		return fmt.Errorf("%w: List cannot acquire endpoints", ErrUnsupported)
	}
	if text(f, "name") != bind {
		return fmt.Errorf("%w: unbound local query alias", ErrInvalid)
	}
	return nil
}

func (s *selection) checkBindings(names map[string]graphops.ResourceKind) error {
	if s.sourceBinding == "" {
		return nil
	}
	kind, ok := names[s.sourceBinding]
	if !ok {
		return fmt.Errorf("%w: unknown or forward query input binding %q", ErrInvalid, s.sourceBinding)
	}
	if kind != s.kind {
		return fmt.Errorf("%w: query input binding %q", ErrKind, s.sourceBinding)
	}
	return nil
}

func (s *selection) checkTypes(input Input) error {
	for _, use := range s.types {
		if input.TypeKinds == nil {
			return fmt.Errorf("%w: Type authority is unavailable", ErrUnsupported)
		}
		kind, ok := input.TypeKinds[use.uri]
		if !ok {
			return fmt.Errorf("%w: nominal query Type", ErrNotFound)
		}
		if kind != s.kind {
			return fmt.Errorf("%w: nominal query Type", ErrKind)
		}
		if use.match == "family" {
			if input.TypeFamilies == nil {
				return fmt.Errorf("%w: Type family authority is unavailable", ErrUnsupported)
			}
			for uri, k := range input.TypeKinds {
				if k != s.kind {
					continue
				}
				family, present := input.TypeFamilies[uri]
				reflexive := false
				for _, ancestor := range family {
					if ancestor == uri {
						reflexive = true
					}
				}
				if !present || !reflexive {
					return fmt.Errorf("%w: incomplete asserted Type ancestry", ErrInvalid)
				}
			}
		}
	}
	for _, uri := range s.references {
		if _, err := graphops.ScopeLocalPublicResourceID(input.Scope, uri); err != nil {
			return fmt.Errorf("%w: cross-Scope staged selection", ErrUnsupported)
		}
	}
	return nil
}

func (s *staging) selectionTick() error {
	s.selectionWork++
	if s.selectionWork > MaxSelectionWork {
		return fmt.Errorf("%w: cumulative work", ErrLimit)
	}
	return nil
}

func (s *staging) selectIDs(index int, a admitted) error {
	query := a.selection
	ids := []string{}
	if query.scan {
		for id, r := range s.live {
			if err := s.selectionTick(); err != nil {
				return err
			}
			if r.Kind == query.kind {
				ids = append(ids, id)
			}
		}
	} else if query.sourceBinding != "" {
		capture, ok := s.bindings[query.sourceBinding]
		if !ok {
			return ErrInvalid
		}
		ids = append(ids, capture.ids...)
	} else {
		ids = append(ids, query.references...)
	}
	unique := map[string]bool{}
	candidates := make([]string, 0, len(ids))
	for _, id := range ids {
		if !unique[id] {
			unique[id] = true
			candidates = append(candidates, id)
		}
	}
	sort.Slice(candidates, func(i, j int) bool { return graphops.CompareCodeUnits(candidates[i], candidates[j]) < 0 })
	selected := []string{}
	for _, id := range candidates {
		if err := s.selectionTick(); err != nil {
			return err
		}
		r, ok := s.live[id]
		if !ok {
			return fmt.Errorf("%w: captured/query Resource", ErrNotFound)
		}
		if r.Kind != query.kind {
			return ErrKind
		}
		keep := true
		if query.predicate != nil {
			body, err := s.selectionBody(r)
			if err != nil {
				return err
			}
			keep, err = s.testLocal(query.predicate, r, body)
			if err != nil {
				return err
			}
		}
		if keep {
			selected = append(selected, id)
			if len(selected) > MaxSelectedIDs {
				return fmt.Errorf("%w: complete captured set", ErrLimit)
			}
		}
	}
	s.bindings[a.member.Name] = capturedBinding{kind: query.kind, ids: append([]string{}, selected...), set: true}
	s.selections = append(s.selections, SelectionCapture{OperationIndex: index, Name: a.member.Name, Kind: query.kind, IDs: append([]string{}, selected...)})
	return nil
}

func (s *staging) selectionBody(r Resource) (map[string]any, error) {
	if err := s.chargeSelectionResource(r); err != nil {
		return nil, err
	}
	body := map[string]any{"id": r.ID}
	body["type"] = referenceValue(r.Type)
	for _, field := range []struct {
		name string
		raw  json.RawMessage
	}{{"properties", r.Properties}, {"metadata", r.Metadata}} {
		if field.raw == nil {
			continue
		}
		value, err := decodeValue(field.raw)
		if err != nil {
			return nil, err
		}
		body[field.name] = value
	}
	if r.Kind == graphops.KindLink {
		body["source"] = referenceValue(r.Source)
		body["target"] = referenceValue(r.Target)
	}
	return body, nil
}

func (s *staging) chargeSelectionResource(r Resource) error {
	s.selectionBytes += len(r.Properties) + len(r.Metadata)
	if s.selectionBytes > MaxSelectionBytes {
		return fmt.Errorf("%w: cumulative Resource value work", ErrLimit)
	}
	return nil
}

func referenceValue(ref canonicalbatch.Reference) any {
	if ref.Revision == "" {
		return ref.URI
	}
	return map[string]any{"uri": ref.URI, "revision": ref.Revision}
}

func decodeValue(raw json.RawMessage) (any, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	err := decoder.Decode(&value)
	return value, err
}

type selectionValue struct {
	present bool
	data    any
}

func (s *staging) localValue(n *localNode, r Resource, body map[string]any) (selectionValue, error) {
	if err := s.selectionTick(); err != nil {
		return selectionValue{}, err
	}
	if n.op == "literal" {
		value, err := decodeValue(n.literal)
		return selectionValue{true, value}, err
	}
	if n.op == "id" {
		return selectionValue{true, r.ID}, nil
	}
	var current any = body
	for _, part := range n.path {
		if err := s.selectionTick(); err != nil {
			return selectionValue{}, err
		}
		switch key := part.(type) {
		case string:
			object, ok := current.(map[string]any)
			if !ok {
				return selectionValue{}, nil
			}
			current, ok = object[key]
			if !ok {
				return selectionValue{}, nil
			}
		case json.Number:
			array, ok := current.([]any)
			if !ok {
				return selectionValue{}, nil
			}
			index, err := key.Int64()
			if err != nil || index < 0 || index >= int64(len(array)) {
				return selectionValue{}, nil
			}
			current = array[index]
		default:
			return selectionValue{}, ErrInvalid
		}
	}
	return selectionValue{true, current}, nil
}

func (s *staging) testLocal(n *localNode, r Resource, body map[string]any) (bool, error) {
	if err := s.selectionTick(); err != nil {
		return false, err
	}
	switch n.op {
	case "bool":
		return n.boolean, nil
	case "and", "or":
		for _, child := range n.children {
			ok, err := s.testLocal(child, r, body)
			if err != nil {
				return false, err
			}
			if n.op == "and" && !ok {
				return false, nil
			}
			if n.op == "or" && ok {
				return true, nil
			}
		}
		return n.op == "and", nil
	case "not":
		ok, err := s.testLocal(n.children[0], r, body)
		return !ok, err
	case "typeIs":
		if n.match == "exact" {
			return r.Type.URI == n.typ, nil
		}
		family, ok := s.input.TypeFamilies[r.Type.URI]
		if !ok {
			return false, fmt.Errorf("%w: missing actual Type ancestry", ErrInvalid)
		}
		for _, ancestor := range family {
			if ancestor == n.typ {
				return true, nil
			}
		}
		return false, nil
	case "present":
		value, err := s.localValue(n.children[0], r, body)
		return value.present, err
	}
	left, err := s.localValue(n.children[0], r, body)
	if err != nil {
		return false, err
	}
	right, err := s.localValue(n.children[1], r, body)
	if err != nil {
		return false, err
	}
	switch n.op {
	case "compare":
		if n.operator == "eq" || n.operator == "ne" {
			equal, err := equalSelectionValues(left, right)
			if n.operator == "ne" {
				equal = !equal
			}
			return equal, err
		}
		comparison, comparable := orderSelectionValues(left, right)
		if !comparable {
			return false, nil
		}
		switch n.operator {
		case "lt":
			return comparison < 0, nil
		case "le":
			return comparison <= 0, nil
		case "gt":
			return comparison > 0, nil
		case "ge":
			return comparison >= 0, nil
		}
	case "contains":
		text, lok := left.data.(string)
		substring, rok := right.data.(string)
		return left.present && right.present && lok && rok && strings.Contains(text, substring), nil
	case "member":
		array, ok := right.data.([]any)
		if !left.present || !right.present || !ok {
			return false, nil
		}
		for _, item := range array {
			if err := s.selectionTick(); err != nil {
				return false, err
			}
			equal, err := equalSelectionValues(left, selectionValue{true, item})
			if err != nil {
				return false, err
			}
			if equal {
				return true, nil
			}
		}
		return false, nil
	}
	return false, ErrUnsupported
}

func equalSelectionValues(a, b selectionValue) (bool, error) {
	if a.present != b.present {
		return false, nil
	}
	if !a.present {
		return true, nil
	}
	left, err := json.Marshal(a.data)
	if err != nil {
		return false, err
	}
	right, err := json.Marshal(b.data)
	if err != nil {
		return false, err
	}
	return graphops.JSONEqual(left, right)
}

func orderSelectionValues(a, b selectionValue) (int, bool) {
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
	l, lok := new(big.Rat).SetString(string(left))
	r, rok := new(big.Rat).SetString(string(right))
	if !lok || !rok {
		return 0, false
	}
	return l.Cmp(r), true
}
