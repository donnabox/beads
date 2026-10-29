package resourcequery

import (
	"bytes"
	"encoding/json"
	"math/big"
	"regexp"
	"strconv"
	"strings"

	graph "github.com/steveyegge/beads/graphops"
)

type reference struct{ uri, pin string }
type segment struct {
	key     string
	index   int
	indexed bool
}
type node struct {
	op, at, collection, bind, name, side, direction, operator, typ, match   string
	input, roots, of, predicate, left, right, text, substring, value, array *node
	operands                                                                []*node
	refs                                                                    []reference
	path                                                                    []segment
	literal                                                                 any
	depth                                                                   int
	boolean                                                                 bool
}
type familyUse struct {
	uri  string
	kind graph.ResourceKind
	at   string
}

var bindingName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func jsonValue(raw []byte) (any, error) {
	var v any
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	if err := d.Decode(&v); err != nil {
		return nil, err
	}
	return v, nil
}
func absoluteURI(s string) bool {
	return graph.ValidateReferenceURI(s) == nil
}
func object(v any, at string, required, optional []string) (map[string]any, error) {
	m, ok := v.(map[string]any)
	if !ok {
		return nil, fail("invalid-shape", at, "expected an object", ErrInvalid)
	}
	allowed := map[string]bool{}
	for _, k := range required {
		allowed[k] = true
		if _, ok := m[k]; !ok {
			return nil, fail("invalid-shape", at, "missing "+k, ErrInvalid)
		}
	}
	for _, k := range optional {
		allowed[k] = true
	}
	for k := range m {
		if !allowed[k] {
			return nil, fail("invalid-shape", at, "unknown member "+k, ErrInvalid)
		}
	}
	return m, nil
}
func stringMember(m map[string]any, key, at string) (string, error) {
	s, ok := m[key].(string)
	if !ok || s == "" {
		return "", fail("invalid-shape", at, "expected nonempty "+key, ErrInvalid)
	}
	return s, nil
}
func integer(v any) (int, bool) {
	n, ok := v.(json.Number)
	if !ok {
		return 0, false
	}
	r, ok := new(big.Rat).SetString(string(n))
	if !ok || !r.IsInt() || !r.Num().IsInt64() {
		return 0, false
	}
	i := r.Num().Int64()
	vint := int(i)
	return vint, int64(vint) == i
}
func parseReference(v any, at string) (reference, error) {
	if s, ok := v.(string); ok {
		if absoluteURI(s) {
			return reference{uri: s}, nil
		}
		return reference{}, fail("invalid-reference", at, "expected absolute Reference", ErrInvalid)
	}
	m, err := object(v, at, []string{"uri", "revision"}, nil)
	if err != nil {
		return reference{}, err
	}
	u, err := stringMember(m, "uri", at)
	if err != nil {
		return reference{}, err
	}
	pin, err := stringMember(m, "revision", at)
	if err != nil {
		return reference{}, err
	}
	if !absoluteURI(u) {
		return reference{}, fail("invalid-reference", at, "expected absolute Reference", ErrInvalid)
	}
	return reference{u, pin}, nil
}

// Decode enforces the frozen schema's closed structural shapes and conservative
// hard parsing bounds. It does not acquire Resources or validate runtime authority.
func Decode(raw []byte) (*Request, error) {
	if len(raw) > 1048576 {
		return nil, fail("request-limit", "", "request exceeds hard bound", ErrLimit)
	}
	canonical, err := graph.CanonicalizeJSON(raw)
	if err != nil {
		return nil, fail("invalid-json", "", "invalid JSON or numeric admission", ErrInvalid)
	}
	v, err := jsonValue(canonical)
	if err != nil {
		return nil, err
	}
	m, err := object(v, "/query", []string{"ir", "profile", "selection", "result"}, []string{"projection"})
	if err != nil {
		return nil, err
	}
	ir, ok := m["ir"].(string)
	if !ok {
		return nil, fail("invalid-ir", "/query/ir", "IR must be a string", ErrInvalid)
	}
	if ir != IR {
		return nil, fail("unsupported-ir", "/query/ir", "unsupported IR", ErrUnsupported)
	}
	profile, ok := m["profile"].(string)
	if !ok || profile != "list" && profile != "graph" && profile != "query" {
		return nil, fail("invalid-profile", "/query/profile", "unknown profile", ErrInvalid)
	}
	rm, err := object(m["result"], "/query/result", []string{"form"}, nil)
	if err != nil {
		return nil, err
	}
	form, ok := rm["form"].(string)
	if !ok || form != "resources" && form != "count" && form != "graph" {
		return nil, fail("invalid-result", "/query/result", "unknown result form", ErrInvalid)
	}
	r := &Request{raw: append([]byte(nil), raw...), profile: profile, form: form}
	r.root, err = r.parse(m["selection"], "/query/selection", 1)
	if err != nil {
		return nil, err
	}
	if p, ok := m["projection"]; ok {
		pm, e := object(p, "/query/projection", nil, []string{"include", "omit"})
		if e != nil {
			return nil, e
		}
		if len(pm) == 0 {
			return nil, fail("invalid-projection", "/query/projection", "empty projection", ErrInvalid)
		}
		for key, val := range pm {
			a, ok := val.([]any)
			if !ok || len(a) == 0 {
				return nil, fail("invalid-projection", "/query/projection", "invalid mask", ErrInvalid)
			}
			for _, x := range a {
				s, ok := x.(string)
				if !ok || key == "omit" && s == "" || !pointer(s) {
					return nil, fail("invalid-projection", "/query/projection", "invalid JSON pointer", ErrInvalid)
				}
			}
		}
		return nil, fail("projection-unsupported", "/query/projection", "Resource masks are not implemented by this evaluator", ErrUnsupported)
	}
	parseLimits := DefaultLimits()
	parseLimits.RequestBytes = 1048576
	parseLimits.Nodes = 16384
	parseLimits.Nesting = 128
	if err = admit(r, parseLimits); err != nil {
		return nil, err
	}
	return r, nil
}
func pointer(s string) bool {
	if s != "" && !strings.HasPrefix(s, "/") {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] == '~' {
			i++
			if i >= len(s) || s[i] != '0' && s[i] != '1' {
				return false
			}
		}
	}
	return true
}

func (r *Request) parse(v any, at string, depth int) (*node, error) {
	r.nodes++
	if depth > r.nesting {
		r.nesting = depth
	}
	if r.nodes > 16384 || depth > 128 {
		return nil, fail("ast-limit", at, "AST exceeds hard parsing bound", ErrLimit)
	}
	if b, ok := v.(bool); ok {
		return &node{op: "bool", at: at, boolean: b}, nil
	}
	m, ok := v.(map[string]any)
	if !ok {
		return nil, fail("invalid-node", at, "expected tagged AST node", ErrInvalid)
	}
	op, err := stringMember(m, "op", at)
	if err != nil {
		return nil, err
	}
	n := &node{op: op, at: at, depth: 1}
	var required, optional []string
	switch op {
	case "scan":
		required = []string{"op", "collection"}
	case "refs":
		required = []string{"op", "collection", "references"}
	case "filter":
		required = []string{"op", "input", "bind", "predicate"}
	case "walk":
		required = []string{"op", "roots", "direction"}
		optional = []string{"depth", "bind", "through"}
	case "endpoints":
		required = []string{"op", "input", "side"}
	case "binding":
		required = []string{"op", "name"}
	case "endpoint":
		required = []string{"op", "of", "side"}
	case "literal":
		required = []string{"op", "value"}
	case "field":
		required = []string{"op", "of", "path"}
	case "id", "rev":
		required = []string{"op", "of"}
	case "and", "or":
		required = []string{"op", "operands"}
	case "not":
		required = []string{"op", "predicate"}
	case "compare":
		required = []string{"op", "operator", "left", "right"}
	case "present":
		required = []string{"op", "value"}
	case "contains":
		required = []string{"op", "text", "substring"}
	case "member":
		required = []string{"op", "value", "array"}
	case "typeIs":
		required = []string{"op", "of", "type", "match"}
	case "any":
		required = []string{"op", "input"}
		optional = []string{"bind", "predicate"}
	case "all":
		required = []string{"op", "input", "bind", "predicate"}
	default:
		return nil, fail("unknown-node", at, "unknown AST operator", ErrInvalid)
	}
	if _, err = object(v, at, required, optional); err != nil {
		return nil, err
	}
	for _, field := range []struct {
		key  string
		dest *string
	}{{"collection", &n.collection}, {"bind", &n.bind}, {"name", &n.name}, {"side", &n.side}, {"direction", &n.direction}, {"operator", &n.operator}, {"type", &n.typ}, {"match", &n.match}} {
		if _, exists := m[field.key]; exists {
			*field.dest, err = stringMember(m, field.key, at)
			if err != nil {
				return nil, err
			}
		}
	}
	if n.bind != "" && !bindingName.MatchString(n.bind) || n.name != "" && !bindingName.MatchString(n.name) {
		return nil, fail("invalid-binding", at, "invalid binding name", ErrInvalid)
	}
	if n.collection != "" && n.collection != "beads" && n.collection != "links" {
		return nil, fail("invalid-collection", at, "unknown collection", ErrInvalid)
	}
	if n.side != "" && n.side != "source" && n.side != "target" {
		return nil, fail("invalid-side", at, "unknown endpoint side", ErrInvalid)
	}
	if n.direction != "" && n.direction != "incoming" && n.direction != "outgoing" && n.direction != "connected" {
		return nil, fail("invalid-direction", at, "unknown direction", ErrInvalid)
	}
	if op == "compare" && n.operator != "eq" && n.operator != "ne" && n.operator != "lt" && n.operator != "le" && n.operator != "gt" && n.operator != "ge" {
		return nil, fail("invalid-compare", at, "unknown comparison", ErrInvalid)
	}
	if op == "typeIs" && (graph.ValidateTypeURL(n.typ) != nil || n.match != "exact" && n.match != "family") {
		return nil, fail("invalid-type", at, "invalid nominal Type test", ErrInvalid)
	}
	if op == "literal" {
		n.literal = m["value"]
		return n, nil
	}
	if op == "refs" {
		a, ok := m["references"].([]any)
		if !ok {
			return nil, fail("transaction-input-unsupported", at, "standalone reads require a Reference array", ErrUnsupported)
		}
		seen := map[string]bool{}
		for _, x := range a {
			ref, e := parseReference(x, at+"/references")
			if e != nil {
				return nil, e
			}
			b, _ := json.Marshal(x)
			if seen[string(b)] {
				return nil, fail("duplicate-reference", at, "duplicate Reference operand", ErrInvalid)
			}
			seen[string(b)] = true
			n.refs = append(n.refs, ref)
		}
	}
	if op == "walk" {
		if d, exists := m["depth"]; exists {
			if d == "unbounded" {
				n.depth = -1
			} else {
				value, ok := integer(d)
				if !ok || value < 0 {
					return nil, fail("invalid-depth", at, "depth must be a nonnegative integer or unbounded", ErrInvalid)
				}
				n.depth = value
			}
		}
		_, through := m["through"]
		if through != (n.bind != "") {
			return nil, fail("invalid-binding", at, "walk bind and through must occur together", ErrInvalid)
		}
	}
	if op == "any" {
		_, test := m["predicate"]
		if test != (n.bind != "") {
			return nil, fail("invalid-binding", at, "any bind and predicate must occur together", ErrInvalid)
		}
	}
	if op == "field" {
		a, ok := m["path"].([]any)
		if !ok || len(a) == 0 {
			return nil, fail("invalid-path", at, "field path must be nonempty", ErrInvalid)
		}
		for _, x := range a {
			if s, ok := x.(string); ok {
				n.path = append(n.path, segment{key: s})
			} else {
				i, ok := integer(x)
				if !ok || i < 0 {
					return nil, fail("invalid-path", at, "invalid path segment", ErrInvalid)
				}
				n.path = append(n.path, segment{index: i, indexed: true})
			}
		}
	}
	for _, child := range []struct {
		key  string
		dest **node
	}{{"input", &n.input}, {"roots", &n.roots}, {"of", &n.of}, {"predicate", &n.predicate}, {"left", &n.left}, {"right", &n.right}, {"text", &n.text}, {"substring", &n.substring}, {"value", &n.value}, {"array", &n.array}, {"through", &n.predicate}} {
		if x, exists := m[child.key]; exists {
			*child.dest, err = r.parse(x, at+"/"+child.key, depth+1)
			if err != nil {
				return nil, err
			}
		}
	}
	if op == "and" || op == "or" {
		a, ok := m["operands"].([]any)
		if !ok || len(a) == 0 {
			return nil, fail("invalid-operands", at, "Boolean group must be nonempty", ErrInvalid)
		}
		for i, x := range a {
			child, e := r.parse(x, at+"/operands/"+strconv.Itoa(i), depth+1)
			if e != nil {
				return nil, e
			}
			n.operands = append(n.operands, child)
		}
	}
	return n, nil
}
