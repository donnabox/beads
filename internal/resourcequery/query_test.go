package resourcequery

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	graph "github.com/steveyegge/beads/graphops"
)

type ast = map[string]any

const base = "https://query.example/"

func scan(collection string) ast { return ast{"op": "scan", "collection": collection} }
func refset(collection string, refs ...any) ast {
	if refs == nil {
		refs = []any{}
	}
	return ast{"op": "refs", "collection": collection, "references": refs}
}
func binding(name string) ast        { return ast{"op": "binding", "name": name} }
func literal(v any) ast              { return ast{"op": "literal", "value": v} }
func field(of ast, parts ...any) ast { return ast{"op": "field", "of": of, "path": parts} }
func compare(op string, a, b ast) ast {
	return ast{"op": "compare", "operator": op, "left": a, "right": b}
}
func filter(input ast, name string, predicate any) ast {
	return ast{"op": "filter", "input": input, "bind": name, "predicate": predicate}
}
func endpoint(of ast, side string) ast { return ast{"op": "endpoint", "of": of, "side": side} }
func walk(roots ast, direction string, depth any) ast {
	return ast{"op": "walk", "roots": roots, "direction": direction, "depth": depth}
}
func request(t *testing.T, profile, form string, selection ast) []byte {
	t.Helper()
	raw, err := json.Marshal(ast{"ir": IR, "profile": profile, "selection": selection, "result": ast{"form": form}})
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
func resource(t *testing.T, kind graph.ResourceKind, id, typ, revision string, properties ast, endpoints ...any) Resource {
	t.Helper()
	body := ast{"id": base + id, "revision": revision, "type": base + typ, "properties": properties, "metadata": ast{"owner": "Donna"}}
	if len(endpoints) > 0 {
		body["source"] = endpoints[0]
		body["target"] = endpoints[1]
	}
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	r, err := NewResource(kind, raw)
	if err != nil {
		t.Fatal(err)
	}
	return r
}
func bead(t *testing.T, id string, props ast) Resource {
	return resource(t, graph.KindBead, id, "Note", "r1", props)
}
func link(t *testing.T, id, typ string, source, target any) Resource {
	return resource(t, graph.KindLink, id, typ, "l1", ast{}, source, target)
}
func execute(t *testing.T, raw []byte, d Dataset) Response {
	t.Helper()
	response, err := Execute(context.Background(), raw, d, Limits{})
	if err != nil {
		t.Fatal(err)
	}
	if !response.Evaluation.Complete || response.Evaluation.Scope != "response" {
		t.Fatalf("invalid evaluation: %+v", response.Evaluation)
	}
	return response
}
func ids(t *testing.T, items []json.RawMessage) []string {
	t.Helper()
	out := make([]string, 0, len(items))
	for _, raw := range items {
		var body struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Fatal(err)
		}
		out = append(out, strings.TrimPrefix(body.ID, base))
	}
	return out
}
func assertIDs(t *testing.T, response Response, want ...string) {
	t.Helper()
	result, ok := response.Result.(ResourcesResult)
	if !ok {
		t.Fatalf("expected resources, got %T", response.Result)
	}
	if want == nil {
		want = []string{}
	}
	if got := ids(t, result.Items); !reflect.DeepEqual(got, want) {
		t.Fatalf("IDs = %v, want %v", got, want)
	}
}
func assertFailure(t *testing.T, raw []byte, d Dataset, want error) {
	t.Helper()
	response, err := Execute(context.Background(), raw, d, Limits{})
	if !errors.Is(err, want) {
		t.Fatalf("error = %v, want %v", err, want)
	}
	if response.Result != nil || response.Evaluation.Complete {
		t.Fatalf("failure returned partial success: %+v", response)
	}
}

func TestStrictDecodeAndPreAcquisitionAdmission(t *testing.T) {
	valid := request(t, "list", "resources", scan("beads"))
	r, err := Decode(valid)
	if err != nil {
		t.Fatal(err)
	}
	if r.Profile() != "list" || r.ResultForm() != "resources" || r.ResourceKind() != graph.KindBead {
		t.Fatalf("wrong dispatch: %+v", r)
	}
	copyRaw := r.Bytes()
	copyRaw[0] = 'x'
	if string(r.Bytes()) != string(valid) {
		t.Fatal("Request leaked mutable bytes")
	}
	for _, raw := range []string{
		`{"ir":"bdp-query-design-1","ir":"bdp-query-design-1","profile":"list","selection":{"op":"scan","collection":"beads"},"result":{"form":"resources"}}`,
		string(valid) + ` {}`,
		strings.Replace(string(valid), `"collection":"beads"`, `"collection":"beads","extra":true`, 1),
		strings.Replace(string(valid), `"collection":"beads"`, `"collection":"types"`, 1),
		strings.Replace(string(valid), `"op":"scan"`, `"op":"future"`, 1),
		`{"ir":"bdp-query-design-1","profile":"query","selection":{"op":"refs","collection":"beads","references":["relative"]},"result":{"form":"resources"}}`,
	} {
		if _, err := Decode([]byte(raw)); !errors.Is(err, ErrInvalid) {
			t.Fatalf("accepted %s: %v", raw, err)
		}
	}
	badLexical := filter(scan("beads"), "b", ast{"op": "or", "operands": []any{true, compare("eq", field(binding("unknown"), "x"), literal(nil))}})
	if _, err := Decode(request(t, "query", "resources", badLexical)); !errors.Is(err, ErrInvalid) {
		t.Fatalf("skipped branch bypassed lexical admission: %v", err)
	}
	badInput := filter(ast{"op": "endpoints", "input": walk(binding("b"), "outgoing", 1), "side": "target"}, "b", true)
	if _, err := Decode(request(t, "query", "resources", badInput)); !errors.Is(err, ErrInvalid) {
		t.Fatalf("binder leaked into own input: %v", err)
	}
	for _, selection := range []ast{
		filter(scan("beads"), "b", ast{"op": "any", "input": scan("links")}),
		filter(scan("beads"), "b", compare("eq", field(binding("b"), "x"), field(binding("b"), "y"))),
	} {
		if _, err := Decode(request(t, "list", "resources", selection)); !errors.Is(err, ErrInvalid) {
			t.Fatalf("List accepted nonlocal predicate: %v", err)
		}
	}
	wrongKind := filter(scan("beads"), "b", ast{"op": "present", "value": field(endpoint(binding("b"), "target"), "properties", "x")})
	if _, err := Decode(request(t, "query", "resources", wrongKind)); !errors.Is(err, ErrInvalid) {
		t.Fatalf("accepted Bead endpoint: %v", err)
	}
	mask := strings.TrimSuffix(string(valid), "}") + `,"projection":{"omit":["/properties"]}}`
	if _, err := Decode([]byte(mask)); !errors.Is(err, ErrUnsupported) {
		t.Fatalf("mask silently accepted: %v", err)
	}
	batch := refset("beads")
	batch["references"] = ast{"binding": "prior"}
	if _, err := Decode(request(t, "query", "resources", batch)); !errors.Is(err, ErrUnsupported) {
		t.Fatalf("Batch input silently accepted: %v", err)
	}
	l := DefaultLimits()
	l.Nodes = 1
	if err := Admit(r, l); err != nil {
		t.Fatalf("single scan should fit: %v", err)
	}
	l.RequestBytes = 1
	if err := Admit(r, l); !errors.Is(err, ErrLimit) {
		t.Fatalf("request byte bound ignored: %v", err)
	}
}

func TestCoreNumbersMissingNullAndLiteralPaths(t *testing.T) {
	d := Dataset{Resources: []Resource{bead(t, "A", ast{"n": json.Number("1.0"), "nil": nil, "false": false, "owner.team": ast{"names": []any{"Donna"}}, "text": "a%b_.*"}), bead(t, "B", ast{"n": "1"})}}
	b := binding("b")
	cases := []struct {
		name      string
		predicate any
		want      []string
	}{
		{"numeric-equality", compare("eq", field(b, "properties", "n"), literal(json.Number("1e0"))), []string{"A"}},
		{"bool-not-number", compare("eq", field(b, "properties", "false"), literal(0)), nil},
		{"missing-not-null", compare("eq", field(b, "properties", "absent"), literal(nil)), nil},
		{"null-present", ast{"op": "present", "value": field(b, "properties", "nil")}, []string{"A"}},
		{"missing-ne", compare("ne", field(b, "properties", "absent"), literal(nil)), []string{"A", "B"}},
		{"two-missing-equal", compare("eq", field(b, "properties", "x"), field(b, "properties", "y")), []string{"A", "B"}},
		{"literal-key-array-index", compare("eq", field(b, "properties", "owner.team", "names", 0), literal("Donna")), []string{"A"}},
		{"literal-substring", ast{"op": "contains", "text": field(b, "properties", "text"), "substring": literal("%b_")}, []string{"A"}},
		{"no-regex", ast{"op": "contains", "text": field(b, "properties", "text"), "substring": literal("a.*_")}, nil},
		{"membership-numeric", ast{"op": "member", "value": field(b, "properties", "n"), "array": literal([]any{json.Number("1e0")})}, []string{"A"}},
		{"ordered-no-coercion", compare("gt", field(b, "properties", "n"), literal(0)), []string{"A"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertIDs(t, execute(t, request(t, "query", "resources", filter(scan("beads"), "b", tc.predicate)), d), tc.want...)
		})
	}
	unsafe := filter(scan("beads"), "b", compare("eq", field(b, "properties", "n"), literal(json.Number("9007199254740993"))))
	if _, err := Decode(request(t, "query", "resources", unsafe)); !errors.Is(err, ErrInvalid) {
		t.Fatalf("Core unsafe number admitted: %v", err)
	}
}

func TestNestedCorrelationShadowingAnyAll(t *testing.T) {
	d := Dataset{Resources: []Resource{bead(t, "A", ast{}), bead(t, "B", ast{}), link(t, "ab", "Requires", base+"A", base+"B")}}
	b := binding("b")
	l := binding("l")
	correlated := filter(scan("beads"), "b", ast{"op": "any", "input": scan("links"), "bind": "l", "predicate": compare("eq", endpoint(l, "source"), b)})
	assertIDs(t, execute(t, request(t, "query", "resources", correlated), d), "A")
	shadow := filter(scan("beads"), "b", ast{"op": "all", "input": scan("links"), "bind": "b", "predicate": compare("eq", ast{"op": "id", "of": binding("b")}, literal(base+"ab"))})
	assertIDs(t, execute(t, request(t, "query", "resources", shadow), d), "A", "B")
	for _, tc := range []struct {
		op   string
		want []string
	}{{"any", nil}, {"all", []string{"A", "B"}}} {
		pred := ast{"op": tc.op, "input": refset("links"), "bind": "l", "predicate": false}
		assertIDs(t, execute(t, request(t, "query", "resources", filter(scan("beads"), "b", pred)), d), tc.want...)
	}
}

func TestThroughVersusOuterSelection(t *testing.T) {
	d := Dataset{Resources: []Resource{bead(t, "A", ast{"draft": false}), bead(t, "B", ast{"draft": false}), bead(t, "C", ast{"draft": true}), link(t, "ab", "Requires", base+"A", base+"B"), link(t, "bc", "Requires", base+"B", base+"C")}}
	pred := compare("eq", field(endpoint(binding("l"), "target"), "properties", "draft"), literal(true))
	w := walk(refset("beads", base+"A"), "outgoing", "unbounded")
	assertIDs(t, execute(t, request(t, "query", "resources", filter(w, "l", pred)), d), "bc")
	w["bind"] = "l"
	w["through"] = pred
	assertIDs(t, execute(t, request(t, "query", "resources", w), d))
}

func TestWalkCyclesParallelLinksNearestRootsAndBoundary(t *testing.T) {
	d := Dataset{Resources: []Resource{bead(t, "A", ast{}), bead(t, "B", ast{}), bead(t, "C", ast{}), bead(t, "D", ast{}), link(t, "ab1", "Requires", base+"A", base+"B"), link(t, "ab2", "Requires", base+"A", base+"B"), link(t, "bc", "Requires", base+"B", base+"C"), link(t, "ca", "Requires", base+"C", base+"A"), link(t, "self", "Requires", base+"A", base+"A"), link(t, "bd", "Requires", base+"B", base+"D")}}
	for _, tc := range []struct {
		depth any
		want  []string
	}{{0, nil}, {1, []string{"ab1", "ab2", "self"}}, {2, []string{"ab1", "ab2", "bc", "bd", "self"}}, {"unbounded", []string{"ab1", "ab2", "bc", "bd", "ca", "self"}}} {
		assertIDs(t, execute(t, request(t, "query", "resources", walk(refset("beads", base+"A"), "outgoing", tc.depth)), d), tc.want...)
	}
	// B is already a root: its outgoing Links are eligible at depth one even
	// though A also reaches it by a longer route. Boundary C does not expand.
	assertIDs(t, execute(t, request(t, "query", "resources", walk(refset("beads", base+"A", base+"B"), "outgoing", 1)), d), "ab1", "ab2", "bc", "bd", "self")
	assertIDs(t, execute(t, request(t, "query", "resources", walk(refset("beads", base+"A"), "incoming", 1)), d), "ca", "self")
	assertIDs(t, execute(t, request(t, "query", "resources", walk(refset("beads", base+"A"), "connected", 1)), d), "ab1", "ab2", "ca", "self")
	// B and D are boundary endpoints at depth one; bd is not induced merely
	// because both Bead identities occur in a separately supplied root set.
	d.Resources = append(d.Resources, link(t, "ad", "Requires", base+"A", base+"D"))
	assertIDs(t, execute(t, request(t, "query", "resources", walk(refset("beads", base+"A"), "outgoing", 1)), d), "ab1", "ab2", "ad", "self")
}

func TestPinsIncidenceOpaqueBoundaryAndRequiredAcquisition(t *testing.T) {
	pin := ast{"uri": base + "B", "revision": "old"}
	d := Dataset{Resources: []Resource{bead(t, "A", ast{}), bead(t, "B", ast{}), link(t, "old", "Requires", base+"A", pin), link(t, "floating", "Requires", base+"A", base+"B")}}
	assertIDs(t, execute(t, request(t, "query", "resources", walk(refset("beads", base+"B"), "incoming", 1)), d), "floating")
	w := walk(refset("beads", base+"A"), "outgoing", 1)
	assertIDs(t, execute(t, request(t, "query", "resources", w), d), "floating", "old")
	assertFailure(t, request(t, "query", "resources", walk(refset("beads", base+"A"), "outgoing", 2)), d, ErrUnsupported)
	assertFailure(t, request(t, "query", "graph", w), d, ErrUnsupported)
	assertFailure(t, request(t, "query", "resources", ast{"op": "endpoints", "input": refset("links", base+"old"), "side": "target"}), d, ErrUnsupported)
	w = walk(refset("beads", base+"A"), "outgoing", 2)
	w["bind"] = "l"
	w["through"] = false
	assertIDs(t, execute(t, request(t, "query", "resources", w), d))
	assertFailure(t, request(t, "query", "resources", refset("beads", ast{"uri": base + "A", "revision": "r1"})), d, ErrUnsupported)
	assertFailure(t, request(t, "query", "resources", refset("beads", ast{"uri": base + "A", "revision": "current"})), d, ErrUnsupported)
	d.Resources[0] = resource(t, graph.KindBead, "A", "Note", "current", ast{})
	assertFailure(t, request(t, "query", "resources", refset("beads", ast{"uri": base + "A", "revision": "current"})), d, ErrUnsupported)
	assertIDs(t, execute(t, request(t, "query", "resources", refset("beads", base+"A")), d), "A")
	for _, reference := range []any{base + "hidden", base + "missing", "https://foreign.example/beads/A"} {
		assertFailure(t, request(t, "query", "resources", refset("beads", reference)), d, ErrUnavailable)
	}
}

func TestDefinitiveProofStopsUnrelatedAcquisition(t *testing.T) {
	d := Dataset{Resources: []Resource{bead(t, "A", ast{}), link(t, "opaque", "Requires", base+"A", base+"missing")}}
	w := walk(binding("b"), "outgoing", "unbounded")
	anyWalk := ast{"op": "any", "input": w}
	response := execute(t, request(t, "query", "resources", filter(scan("beads"), "b", anyWalk)), d)
	assertIDs(t, response, "A")
	if len(response.Evaluation.Traversals) != 1 || response.Evaluation.Traversals[0].Completion != "proof" || response.Evaluation.Traversals[0].Closure != "unknown" {
		t.Fatalf("proof trace: %+v", response.Evaluation)
	}
	allWalk := ast{"op": "all", "input": w, "bind": "l", "predicate": false}
	assertIDs(t, execute(t, request(t, "query", "resources", filter(scan("beads"), "b", allWalk)), d))
	allWalk["predicate"] = true
	assertFailure(t, request(t, "query", "resources", filter(scan("beads"), "b", allWalk)), d, ErrUnavailable)
	// Static Reference assertions never disappear behind a short circuit.
	assertion := ast{"op": "or", "operands": []any{true, ast{"op": "any", "input": refset("beads", base+"missing")}}}
	assertFailure(t, request(t, "query", "resources", filter(scan("beads"), "b", assertion)), d, ErrUnavailable)
}

func TestFamilyAuthorityEvenEmptyAndSkippedAndPinnedType(t *testing.T) {
	b := binding("b")
	family := ast{"op": "typeIs", "of": b, "type": base + "Note", "match": "family"}
	raw := request(t, "query", "resources", filter(scan("beads"), "b", ast{"op": "or", "operands": []any{true, family}}))
	assertFailure(t, raw, Dataset{}, ErrUnsupported)
	d := Dataset{TypeKinds: map[string]graph.ResourceKind{base + "Note": graph.KindBead, base + "Special": graph.KindBead}, TypeFamilies: map[string][]string{base + "Note": {base + "Note"}, base + "Special": {base + "Special", base + "Note"}}, Resources: []Resource{resource(t, graph.KindBead, "A", "Special", "r1", ast{})}}
	assertIDs(t, execute(t, request(t, "query", "resources", filter(scan("beads"), "b", family)), d), "A")
	d.TypeKinds[base+"Note"] = graph.KindLink
	assertFailure(t, raw, d, ErrInvalid)
	// Exact nominal matching ignores a Type definition pin, while field keeps
	// the actual union-shaped Reference rather than flattening it.
	pinned, err := NewResource(graph.KindBead, []byte(`{"id":"https://query.example/A","revision":"r1","type":{"uri":"https://query.example/Note","revision":"definition-7"},"properties":{}}`))
	if err != nil {
		t.Fatal(err)
	}
	d = Dataset{Resources: []Resource{pinned}}
	exact := ast{"op": "typeIs", "of": b, "type": base + "Note", "match": "exact"}
	assertIDs(t, execute(t, request(t, "query", "resources", filter(scan("beads"), "b", exact)), d), "A")
	assertIDs(t, execute(t, request(t, "query", "resources", filter(scan("beads"), "b", compare("eq", field(b, "type", "revision"), literal("definition-7")))), d), "A")
	d.TypeKinds = map[string]graph.ResourceKind{base + "Note": graph.KindBead}
	d.TypeFamilies = map[string][]string{base + "Note": {base + "Note"}}
	assertIDs(t, execute(t, request(t, "query", "resources", filter(scan("beads"), "b", family)), d), "A")
}

func TestFamilyMechanismVersusAuthorizedAuthority(t *testing.T) {
	family := ast{"op": "typeIs", "of": binding("b"), "type": base + "Note", "match": "family"}
	raw := request(t, "query", "count", filter(scan("beads"), "b", ast{"op": "or", "operands": []any{true, family}}))
	for _, tc := range []struct {
		name string
		data Dataset
		want error
	}{
		{"absent mechanism", Dataset{}, ErrUnsupported},
		{"supported empty authority", Dataset{TypeKinds: map[string]graph.ResourceKind{}, TypeFamilies: map[string][]string{}}, ErrUnavailable},
		{"known Type missing ancestry", Dataset{TypeKinds: map[string]graph.ResourceKind{base + "Note": graph.KindBead}, TypeFamilies: map[string][]string{}}, ErrDataset},
		{"record Type missing from claimed complete authority", Dataset{TypeKinds: map[string]graph.ResourceKind{base + "Note": graph.KindBead}, TypeFamilies: map[string][]string{base + "Note": {base + "Note"}}, Resources: []Resource{resource(t, graph.KindBead, "A", "Special", "r1", ast{})}}, ErrDataset},
	} {
		t.Run(tc.name, func(t *testing.T) { assertFailure(t, raw, tc.data, tc.want) })
	}
}

func TestFamilyPreloadDoesNotMultiplyRecordsByUses(t *testing.T) {
	d := Dataset{TypeKinds: map[string]graph.ResourceKind{base + "Note": graph.KindBead}, TypeFamilies: map[string][]string{base + "Note": {base + "Note"}}}
	for i := 0; i < 200; i++ {
		d.Resources = append(d.Resources, resource(t, graph.KindBead, "N"+strconv.Itoa(i), "Note", "r1", ast{}))
	}
	operands := []any{true}
	for i := 0; i < 100; i++ {
		operands = append(operands, ast{"op": "typeIs", "of": binding("b"), "type": base + "Note", "match": "family"})
	}
	raw := request(t, "query", "count", filter(scan("beads"), "b", ast{"op": "or", "operands": operands}))
	limits := DefaultLimits()
	limits.MaxWork = 5000
	response, err := Execute(context.Background(), raw, d, limits)
	if err != nil {
		t.Fatal(err)
	}
	if got := response.Result.(CountResult).Count; got != 200 {
		t.Fatalf("count = %d", got)
	}
}

func TestFullResourcesCountGraphAndLimits(t *testing.T) {
	d := Dataset{Resources: []Resource{bead(t, "A", ast{"kept": true}), bead(t, "B", ast{}), link(t, "ab", "Requires", base+"A", base+"B")}}
	resources := execute(t, request(t, "list", "resources", scan("beads")), d).Result.(ResourcesResult)
	if resources.Next != nil || !strings.Contains(string(resources.Items[0]), `"owner":"Donna"`) || !strings.Contains(string(resources.Items[0]), `"kept":true`) {
		t.Fatal("full Resource was projected or paged")
	}
	countData := d
	countData.Resources = append(append([]Resource(nil), d.Resources...), link(t, "ab-parallel", "Requires", base+"A", base+"B"))
	count := execute(t, request(t, "query", "count", ast{"op": "endpoints", "input": scan("links"), "side": "target"}), countData).Result.(CountResult)
	if count.Count != 1 || count.ResourceKind != graph.KindBead {
		t.Fatalf("count = %+v", count)
	}
	for _, depth := range []int{0, 1} {
		response := execute(t, request(t, "graph", "graph", walk(refset("beads", base+"A"), "outgoing", depth)), d)
		g := response.Result.(GraphResult)
		if len(g.Roots) != 1 || g.Roots[0] != (Edition{base + "A", "r1"}) {
			t.Fatalf("lost explicit root: %+v", g)
		}
		if depth == 0 {
			if len(g.Beads) != 1 || len(g.Links) != 0 || len(g.Endpoints) != 0 {
				t.Fatalf("zero depth graph: %+v", g)
			}
		} else {
			if len(g.Beads) != 2 || len(g.Links) != 1 || g.Endpoints[0].Target != (Edition{base + "B", "r1"}) {
				t.Fatalf("graph mapping: %+v", g)
			}
		}
		serialized, err := json.Marshal(response)
		if err != nil || strings.Contains(string(serialized), `"traversals":null`) || strings.Contains(string(serialized), `"links":null`) {
			t.Fatalf("invalid carrier: %s %v", serialized, err)
		}
	}
	raw := request(t, "list", "count", scan("beads"))
	l := DefaultLimits()
	l.MaxResults = 1
	response, err := Execute(context.Background(), raw, d, l)
	if !errors.Is(err, ErrLimit) || response.Result != nil {
		t.Fatalf("partial count under result bound: %+v %v", response, err)
	}
	l = DefaultLimits()
	l.MaxWork = 1
	response, err = Execute(context.Background(), raw, d, l)
	if !errors.Is(err, ErrLimit) || response.Result != nil {
		t.Fatalf("work bound: %+v %v", response, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	response, err = Execute(ctx, raw, d, Limits{})
	if !errors.Is(err, context.Canceled) || response.Result != nil {
		t.Fatalf("cancellation: %+v %v", response, err)
	}
	d.Resources = append(d.Resources, d.Resources[0])
	assertFailure(t, raw, d, ErrDataset)
}

func TestTraceAggregatesCorrelatedProofOverDepth(t *testing.T) {
	d := Dataset{Resources: []Resource{bead(t, "A", ast{}), bead(t, "B", ast{}), bead(t, "X", ast{}), link(t, "bad", "Requires", base+"A", base+"X"), link(t, "good", "Requires", base+"B", base+"X")}}
	w := walk(binding("b"), "outgoing", 1)
	pred := ast{"op": "any", "input": w, "bind": "l", "predicate": compare("eq", ast{"op": "id", "of": binding("l")}, literal(base+"good"))}
	response := execute(t, request(t, "query", "resources", filter(scan("beads"), "b", pred)), d)
	assertIDs(t, response, "B")
	traces := response.Evaluation.Traversals
	if len(traces) != 1 || traces[0].Completion != "proof" || traces[0].Closure != "unknown" || traces[0].At != "/query/selection/predicate/input" {
		t.Fatalf("wrong correlated aggregate: %+v", traces)
	}
}

func TestThousandNodeRingClosureWithinDefaultBudgets(t *testing.T) {
	const size = 1000
	d := Dataset{Resources: make([]Resource, 0, size*2)}
	want := make([]string, 0, size)
	for i := 0; i < size; i++ {
		id := "node-" + strconv.Itoa(i)
		d.Resources = append(d.Resources, bead(t, id, ast{}))
		linkID := "edge-" + strconv.Itoa(i)
		want = append(want, linkID)
		d.Resources = append(d.Resources, link(t, linkID, "Requires", base+id, base+"node-"+strconv.Itoa((i+1)%size)))
	}
	// Check exact membership and closure under the ordinary default budget.
	// There is no wall-clock assertion or native-provider performance claim.
	sort.Strings(want)
	for _, direction := range []string{"outgoing", "incoming", "connected"} {
		response := execute(t, request(t, "query", "resources", walk(refset("beads", base+"node-0"), direction, "unbounded")), d)
		assertIDs(t, response, want...)
		trace := response.Evaluation.Traversals[0]
		if trace.Completion != "exhausted" || trace.Closure != "complete" {
			t.Fatalf("ring did not establish closure: %+v", trace)
		}
	}
}

func TestConnectedIndexPreservesProofOrderAndSelfPinApplicability(t *testing.T) {
	d := Dataset{Resources: []Resource{bead(t, "A", ast{}), bead(t, "B", ast{}), link(t, "a-proof", "Requires", base+"B", base+"A"), link(t, "b-self", "Requires", ast{"uri": base + "A", "revision": "old"}, ast{"uri": base + "A", "revision": "r1"}), link(t, "z-unavailable", "Requires", base+"A", base+"missing")}}
	pred := ast{"op": "any", "input": walk(binding("b"), "connected", "unbounded"), "bind": "l", "predicate": compare("eq", ast{"op": "id", "of": binding("l")}, literal(base+"a-proof"))}
	// The sorted incoming witness must precede outgoing unavailability. Merely
	// concatenating outgoing and incoming buckets would change this outcome.
	assertIDs(t, execute(t, request(t, "query", "resources", filter(refset("beads", base+"A"), "b", pred)), d), "A")
	// A same-identity self-Link indexed once can match its target even though
	// its authored source pin is inapplicable to the selected current edition.
	assertIDs(t, execute(t, request(t, "query", "resources", walk(refset("beads", base+"A"), "connected", 1)), d), "a-proof", "b-self", "z-unavailable")
	assertIDs(t, execute(t, request(t, "query", "resources", walk(refset("beads", base+"A"), "outgoing", 1)), d), "z-unavailable")
}
