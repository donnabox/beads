package typeref

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"

	"github.com/steveyegge/beads/internal/httpapi/spec"
)

// Direction is which way a document is rewritten.
type Direction uint8

const (
	// ToCallable reads a v0 document and writes the Callable's: a name becomes a
	// Type reference and `issue_type` becomes `type`.
	ToCallable Direction = iota
	// ToV0 reads a Callable's document and writes v0's.
	ToV0
)

// Shape is where the rule reaches inside one JSON value, read from openapi.v0.yaml.
// A nil Shape reaches nothing, and Apply returns its input.
type Shape struct {
	members map[string]*slot
	items   *Shape
	values  *Shape
}

// slot is one member of an object: written under out, with its value converted
// when kind is set, or rewritten by shape when it holds members that are.
type slot struct {
	kind  Kind
	out   string
	shape *Shape
	// refuse marks a member the reading side does not have: v0's spelling of a
	// renamed member, inside a Callable's document. Passing it through would let
	// the old spelling reach v0 around the rule.
	refuse bool
}

// Plans are the shapes of every operation, both ways.
type Plans struct {
	arguments map[Direction]map[string]*Shape
	results   map[Direction]map[string]*Shape
	problem   map[Direction]*Shape
}

// Arguments is where the rule reaches in an operation's flattened `arguments`:
// its query parameters, its body members and its adapter arguments.
func (p *Plans) Arguments(op string, dir Direction) *Shape { return p.arguments[dir][op] }

// Result is where it reaches in an operation's 200 body.
func (p *Plans) Result(op string, dir Direction) *Shape { return p.results[dir][op] }

// Problem is where it reaches in v0's Problem body.
func (p *Plans) Problem(dir Direction) *Shape { return p.problem[dir] }

// Load reads the plans from the embedded openapi.v0.yaml, once.
var Load = sync.OnceValues(func() (*Plans, error) {
	var doc map[string]any
	if err := yaml.Unmarshal(spec.OpenAPIV0(), &doc); err != nil {
		return nil, fmt.Errorf("parse openapi.v0.yaml: %w", err)
	}
	return build(doc)
})

func build(doc map[string]any) (*Plans, error) {
	plans := &Plans{
		arguments: map[Direction]map[string]*Shape{},
		results:   map[Direction]map[string]*Shape{},
		problem:   map[Direction]*Shape{},
	}
	schemas := member(member(doc, "components"), "schemas")
	for component, reach := range members {
		properties := member(member(schemas, component), "properties")
		for name := range reach {
			if _, ok := properties[name]; !ok {
				return nil, fmt.Errorf("the Types rule names %s.%s, which openapi.v0.yaml does not declare", component, name)
			}
		}
	}
	for _, dir := range []Direction{ToCallable, ToV0} {
		b := &builder{doc: doc, dir: dir, memo: map[string]*Shape{}}
		plans.arguments[dir] = map[string]*Shape{}
		plans.results[dir] = map[string]*Shape{}
		for op, operation := range operations(doc) {
			plans.arguments[dir][op] = b.arguments(op, operation)
			response := b.deref(member(member(operation, "responses"), "200"))
			schema := member(member(member(response, "content"), "application/json"), "schema")
			plans.results[dir][op] = b.shape(schema)
		}
		plans.problem[dir] = b.component(problemComponent)
	}
	for op := range parameters {
		if _, ok := plans.arguments[ToV0][op]; !ok {
			return nil, fmt.Errorf("the Types rule names operation %s, which openapi.v0.yaml does not declare", op)
		}
	}
	return plans, nil
}

type builder struct {
	doc  map[string]any
	dir  Direction
	memo map[string]*Shape
}

// names is a member's name on the side read and on the side written.
func (b *builder) names(v0 string) (in, out string) {
	if b.dir == ToCallable {
		return v0, CallableName(v0)
	}
	return CallableName(v0), v0
}

// typed is a slot whose value is a Type, and, inside a Callable's document, the
// refusal of v0's spelling where the rule renames it.
func (s *Shape) typed(b *builder, v0 string, kind Kind) {
	in, out := b.names(v0)
	s.members[in] = &slot{kind: kind, out: out}
	if b.dir == ToV0 && in != out {
		s.members[out] = &slot{refuse: true, out: out}
	}
}

func (b *builder) arguments(op string, operation map[string]any) *Shape {
	s := &Shape{members: map[string]*slot{}}
	list, _ := operation["parameters"].([]any)
	for _, raw := range list {
		param := b.deref(asMap(raw))
		name, _ := param["name"].(string)
		if kind := ParameterKind(op, name); kind != None {
			s.typed(b, name, kind)
		}
	}
	for name, kind := range adapterParameters[op] {
		s.typed(b, name, kind)
	}
	body := b.deref(member(operation, "requestBody"))
	schema := member(member(member(body, "content"), "application/json"), "schema")
	if inner := b.shape(schema); inner != nil {
		for name, member := range inner.members {
			if _, taken := s.members[name]; !taken {
				s.members[name] = member
			}
		}
	}
	return s.orNil()
}

func (b *builder) component(name string) *Shape {
	if s, ok := b.memo[name]; ok {
		return s
	}
	// A placeholder first, so a component that reaches itself ends.
	s := &Shape{members: map[string]*slot{}}
	b.memo[name] = s
	built := b.object(member(member(member(b.doc, "components"), "schemas"), name), name)
	if built == nil {
		b.memo[name] = nil
		return nil
	}
	*s = *built
	return s
}

// shape is the Shape of one schema.
func (b *builder) shape(schema map[string]any) *Shape {
	if ref, ok := schema["$ref"].(string); ok {
		return b.component(ref[strings.LastIndex(ref, "/")+1:])
	}
	return b.object(schema, "")
}

func (b *builder) object(schema map[string]any, component string) *Shape {
	s := &Shape{members: map[string]*slot{}}
	for name, raw := range member(schema, "properties") {
		sub := asMap(raw)
		if kind := MemberKind(component, name); kind != None {
			s.typed(b, name, kind)
			continue
		}
		var inner *Shape
		switch {
		case component == eventRecord && name == eventIssue:
			inner = b.component(issueComponent)
		case component == eventRecord && name == eventDep:
			inner = &Shape{members: map[string]*slot{}}
			in, out := eventDepKind, callableTypeMember
			if b.dir == ToV0 {
				in, out = out, in
				inner.members[eventDepKind] = &slot{refuse: true, out: eventDepKind}
			}
			inner.members[in] = &slot{kind: Link, out: out}
		default:
			inner = b.shape(sub)
		}
		if inner != nil {
			s.members[name] = &slot{out: name, shape: inner}
		}
	}
	if items, ok := schema["items"].(map[string]any); ok {
		s.items = b.shape(items)
	}
	if values, ok := schema["additionalProperties"].(map[string]any); ok {
		s.values = b.shape(values)
	}
	for _, group := range []string{"allOf", "oneOf", "anyOf"} {
		list, _ := schema[group].([]any)
		for _, raw := range list {
			if inner := b.shape(asMap(raw)); inner != nil {
				for name, member := range inner.members {
					s.members[name] = member
				}
				if s.items == nil {
					s.items = inner.items
				}
				if s.values == nil {
					s.values = inner.values
				}
			}
		}
	}
	return s.orNil()
}

func (s *Shape) orNil() *Shape {
	if s == nil || (len(s.members) == 0 && s.items == nil && s.values == nil) {
		return nil
	}
	return s
}

func (b *builder) deref(node map[string]any) map[string]any {
	for {
		ref, ok := node["$ref"].(string)
		if !ok {
			return node
		}
		cur := b.doc
		for _, part := range strings.Split(strings.TrimPrefix(ref, "#/"), "/") {
			cur = member(cur, part)
		}
		node = cur
	}
}

func operations(doc map[string]any) map[string]map[string]any {
	ops := map[string]map[string]any{}
	for _, item := range member(doc, "paths") {
		for _, raw := range asMap(item) {
			if op, ok := raw.(map[string]any); ok {
				if id, ok := op["operationId"].(string); ok {
					ops[id] = op
				}
			}
		}
	}
	return ops
}

func member(node map[string]any, key string) map[string]any { return asMap(node[key]) }

func asMap(v any) map[string]any {
	m, _ := v.(map[string]any)
	return m
}

// Converter spells one Type value the other way. at is the JSON Pointer of the
// value inside the document Apply was given.
type Converter func(kind Kind, value, at string) (string, error)

// MemberError is a member the reading side does not have.
type MemberError struct {
	At, Member, Spelled string
}

func (e *MemberError) Error() string {
	return fmt.Sprintf("`%s` is not a member here; a Type is spelled `%s` (callables.md §2)", e.Member, e.Spelled)
}

// ValueError is a value the rule reaches that is not a string.
type ValueError struct{ At string }

func (e *ValueError) Error() string { return "a Type is a string" }

// Apply rewrites raw under the shape: renamed members, and every value the rule
// reaches spelled by convert. Members keep their order, and a value the rule does
// not reach keeps its bytes. changed says the output differs from raw.
func (s *Shape) Apply(raw json.RawMessage, convert Converter) (json.RawMessage, bool, error) {
	return s.apply(raw, convert, "")
}

func (s *Shape) apply(raw json.RawMessage, convert Converter, at string) (json.RawMessage, bool, error) {
	trimmed := bytes.TrimSpace(raw)
	if s == nil || len(trimmed) == 0 {
		return raw, false, nil
	}
	switch trimmed[0] {
	case '{':
		if len(s.members) == 0 && s.values == nil {
			return raw, false, nil
		}
		return s.applyObject(trimmed, convert, at)
	case '[':
		if s.items == nil {
			return raw, false, nil
		}
		return applyArray(trimmed, at, func(item json.RawMessage, at string) (json.RawMessage, bool, error) {
			return s.items.apply(item, convert, at)
		})
	}
	return raw, false, nil
}

func (s *Shape) applyObject(raw json.RawMessage, convert Converter, at string) (json.RawMessage, bool, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	if _, err := dec.Token(); err != nil {
		return nil, false, err
	}
	out := bytes.NewBuffer(make([]byte, 0, len(raw)+64))
	out.WriteByte('{')
	changed, first := false, true
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return nil, false, err
		}
		key, _ := tok.(string)
		var value json.RawMessage
		if err := dec.Decode(&value); err != nil {
			return nil, false, err
		}
		where := at + "/" + pointerEscape(key)
		name := key
		sl := s.members[key]
		if sl == nil && s.values != nil {
			sl = &slot{out: key, shape: s.values}
		}
		if sl != nil {
			if sl.refuse {
				return nil, false, &MemberError{At: where, Member: key, Spelled: spelledFor(s, key)}
			}
			name = sl.out
			var next json.RawMessage
			var moved bool
			if sl.kind != None {
				next, moved, err = convertValue(sl.kind, value, convert, where)
			} else {
				next, moved, err = sl.shape.apply(value, convert, where)
			}
			if err != nil {
				return nil, false, err
			}
			if moved || name != key {
				changed = true
			}
			value = next
		}
		if !first {
			out.WriteByte(',')
		}
		first = false
		encoded, err := marshalString(name)
		if err != nil {
			return nil, false, err
		}
		out.Write(encoded)
		out.WriteByte(':')
		out.Write(value)
	}
	if !changed {
		return raw, false, nil
	}
	out.WriteByte('}')
	return out.Bytes(), true, nil
}

// spelledFor names the member a refused one is spelled as.
func spelledFor(s *Shape, refused string) string {
	for in, sl := range s.members {
		if sl.out == refused && !sl.refuse {
			return in
		}
	}
	return refused
}

func applyArray(raw json.RawMessage, at string, each func(json.RawMessage, string) (json.RawMessage, bool, error)) (json.RawMessage, bool, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	if _, err := dec.Token(); err != nil {
		return nil, false, err
	}
	out := bytes.NewBuffer(make([]byte, 0, len(raw)+64))
	out.WriteByte('[')
	changed := false
	for i := 0; dec.More(); i++ {
		var item json.RawMessage
		if err := dec.Decode(&item); err != nil {
			return nil, false, err
		}
		next, moved, err := each(item, at+"/"+strconv.Itoa(i))
		if err != nil {
			return nil, false, err
		}
		changed = changed || moved
		if i > 0 {
			out.WriteByte(',')
		}
		out.Write(next)
	}
	if !changed {
		return raw, false, nil
	}
	out.WriteByte(']')
	return out.Bytes(), true, nil
}

// convertValue spells a Type value, or each of an array of them. null stays null.
func convertValue(kind Kind, raw json.RawMessage, convert Converter, at string) (json.RawMessage, bool, error) {
	trimmed := bytes.TrimSpace(raw)
	switch {
	case bytes.Equal(trimmed, []byte("null")):
		return raw, false, nil
	case len(trimmed) > 0 && trimmed[0] == '[':
		return applyArray(trimmed, at, func(item json.RawMessage, at string) (json.RawMessage, bool, error) {
			if bytes.Equal(bytes.TrimSpace(item), []byte("null")) {
				return nil, false, &ValueError{At: at}
			}
			return convertValue(kind, item, convert, at)
		})
	}
	var value string
	if err := json.Unmarshal(trimmed, &value); err != nil {
		return nil, false, &ValueError{At: at}
	}
	next, err := convert(kind, value, at)
	if err != nil {
		return nil, false, err
	}
	if next == value {
		return raw, false, nil
	}
	encoded, err := marshalString(next)
	return encoded, true, err
}

func marshalString(s string) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(s); err != nil {
		return nil, err
	}
	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}

// pointerEscape is RFC 6901's escaping of one reference token.
func pointerEscape(token string) string {
	return strings.ReplaceAll(strings.ReplaceAll(token, "~", "~0"), "/", "~1")
}

// ApplyMembers is Apply on an object held as its members, as the Callables mount
// holds `arguments`. It rewrites members in place and answers the names that
// changed, under their new names.
func (s *Shape) ApplyMembers(members map[string]json.RawMessage, convert Converter) ([]string, error) {
	if s == nil {
		return nil, nil
	}
	keys := make([]string, 0, len(members))
	for key := range members {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var changed []string
	for _, key := range keys {
		sl := s.members[key]
		if sl == nil {
			continue
		}
		where := "/" + pointerEscape(key)
		if sl.refuse {
			return nil, &MemberError{At: where, Member: key, Spelled: spelledFor(s, key)}
		}
		var next json.RawMessage
		var moved bool
		var err error
		if sl.kind != None {
			next, moved, err = convertValue(sl.kind, members[key], convert, where)
		} else {
			next, moved, err = sl.shape.apply(members[key], convert, where)
		}
		if err != nil {
			return nil, err
		}
		if !moved && sl.out == key {
			continue
		}
		delete(members, key)
		if _, taken := members[sl.out]; taken {
			return nil, &MemberError{At: where, Member: key, Spelled: sl.out}
		}
		members[sl.out] = next
		changed = append(changed, sl.out)
	}
	return changed, nil
}

// Rename is the name a top-level member is written under.
func (s *Shape) Rename(name string) string {
	if s != nil {
		if sl := s.members[name]; sl != nil && !sl.refuse {
			return sl.out
		}
	}
	return name
}
