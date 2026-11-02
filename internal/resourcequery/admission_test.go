package resourcequery

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	graph "github.com/steveyegge/beads/graphops"
)

func TestReferenceInputContainerAdmission(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value any
		want  error
	}{
		{"number", 7, ErrInvalid},
		{"null", nil, ErrInvalid},
		{"boolean", true, ErrInvalid},
		{"string", "prior", ErrInvalid},
		{"empty object", ast{}, ErrInvalid},
		{"wrong binding kind", ast{"binding": 7}, ErrInvalid},
		{"empty binding", ast{"binding": ""}, ErrInvalid},
		{"extra member", ast{"binding": "prior", "extra": true}, ErrInvalid},
		{"batch binding", ast{"binding": "prior"}, ErrUnsupported},
	} {
		t.Run(tc.name, func(t *testing.T) {
			selection := ast{"op": "refs", "collection": "beads", "references": tc.value}
			if _, err := Decode(request(t, "query", "resources", selection)); !errors.Is(err, tc.want) {
				t.Fatalf("references=%v: want %v, got %v", tc.value, tc.want, err)
			}
		})
	}
}

func TestWholeASTShapeAdmission(t *testing.T) {
	b := binding("b")
	for _, bad := range []ast{
		{"op": "present", "value": literal(nil)},
		{"op": "compare", "operator": "lt", "left": b, "right": b},
		{"op": "compare", "operator": "eq", "left": b, "right": literal(base + "A")},
		{"op": "and", "operands": []any{}},
		{"op": "any", "input": scan("links"), "bind": "l"},
		{"op": "all", "input": scan("links"), "predicate": true},
		{"op": "contains", "text": b, "substring": literal("x")},
		{"op": "field", "of": b, "path": []any{-1}},
		{"op": "typeIs", "of": b, "type": "relative", "match": "family"},
	} {
		if _, err := Decode(request(t, "query", "resources", filter(scan("beads"), "b", bad))); !errors.Is(err, ErrInvalid) {
			t.Fatalf("admitted %v: %v", bad, err)
		}
	}
	for _, roots := range []ast{scan("links"), literal(nil), scan("beads")} {
		w := walk(roots, "outgoing", 1)
		profile := "query"
		if roots["collection"] == "beads" {
			profile = "graph"
		}
		if _, err := Decode(request(t, profile, "graph", w)); !errors.Is(err, ErrInvalid) {
			t.Fatalf("bad root/profile admitted %v: %v", roots, err)
		}
	}
	w := walk(refset("beads", base+"A"), "outgoing", 1)
	w["bind"] = "l"
	w["through"] = true
	if _, err := Decode(request(t, "graph", "graph", w)); err != nil {
		t.Fatalf("Graph true restriction refused: %v", err)
	}
	w["through"] = false
	if _, err := Decode(request(t, "graph", "graph", w)); !errors.Is(err, ErrInvalid) {
		t.Fatalf("Graph arbitrary restriction admitted: %v", err)
	}
	// List accepts either literal side, but no field-to-field comparison.
	leftLiteral := filter(scan("beads"), "b", compare("eq", literal("draft"), field(b, "properties", "status")))
	if _, err := Decode(request(t, "list", "resources", leftLiteral)); err != nil {
		t.Fatal(err)
	}
	for _, selection := range []ast{scan("beads"), walk(refset("beads"), "outgoing", 1)} {
		if _, err := Decode(request(t, "list", "graph", selection)); !errors.Is(err, ErrInvalid) {
			t.Fatalf("List graph result must be a profile error, not an implementation capability: %v", err)
		}
	}
}

func TestResourceBoundaryCopiesAndRefusesMalformedData(t *testing.T) {
	raw := []byte(`{"id":"https://query.example/A","revision":"r1","type":"https://query.example/Note","properties":{"n":1e0}}`)
	r, err := NewResource(graph.KindBead, raw)
	if err != nil {
		t.Fatal(err)
	}
	raw[0] = 'x'
	if r.Data[0] != '{' {
		t.Fatal("NewResource retained caller bytes")
	}
	if !strings.Contains(string(r.Data), `"n":1`) {
		t.Fatalf("record not Core normalized: %s", r.Data)
	}
	for _, body := range []string{
		`{"id":"https://query.example/A","revision":"","type":"https://query.example/Note","properties":{}}`,
		`{"id":"relative","revision":"r1","type":"https://query.example/Note","properties":{}}`,
		`{"id":"https://QUERY.example/A","revision":"r1","type":"https://query.example/Note","properties":{}}`,
		`{"id":"https://query.example/A|B","revision":"r1","type":"https://query.example/Note","properties":{}}`,
		`{"id":"https://query.example/A","revision":"r1","type":"https://query.example/Note","properties":null}`,
		`{"id":"https://query.example/A","revision":"r1","type":"https://query.example/Note","properties":{"n":9007199254740993}}`,
		`{"id":"https://query.example/A","revision":"r1","revision":"r2","type":"https://query.example/Note","properties":{}}`,
	} {
		if _, err := NewResource(graph.KindBead, []byte(body)); !errors.Is(err, ErrDataset) {
			t.Fatalf("admitted invalid Resource %s: %v", body, err)
		}
	}
	assertFailure(t, request(t, "list", "resources", scan("beads")), Dataset{Resources: []Resource{{Kind: graph.KindBead, Data: []byte(`null`)}}}, ErrDataset)
	_, err = NewResource(graph.KindLink, r.Data)
	if !errors.Is(err, ErrDataset) {
		t.Fatalf("Link without endpoints admitted: %v", err)
	}
}

func TestLiteralReferenceUnionAndDistinctEndpoints(t *testing.T) {
	d := Dataset{Resources: []Resource{bead(t, "A", ast{}), bead(t, "B", ast{}), link(t, "ab1", "Requires", ast{"uri": base + "A", "revision": "r1"}, base+"B"), link(t, "ab2", "Requires", base+"A", base+"B")}}
	selection := ast{"op": "endpoints", "input": scan("links"), "side": "target"}
	assertIDs(t, execute(t, request(t, "query", "resources", selection), d), "B")
	rawUnion := filter(scan("links"), "l", compare("eq", field(binding("l"), "source"), literal(ast{"uri": base + "A", "revision": "r1"})))
	assertIDs(t, execute(t, request(t, "query", "resources", rawUnion), d), "ab1")
	// A URI-shaped identity is literal. No public AST/backend normalization can
	// turn it into a different local key or split it at a revision marker.
	uriBead := bead(t, "https%3A%2F%2Fliteral", ast{})
	d.Resources = append(d.Resources, uriBead)
	assertIDs(t, execute(t, request(t, "query", "resources", refset("beads", base+"https%3A%2F%2Fliteral")), d), "https%3A%2F%2Fliteral")
}

func TestDefaultsAndUnavailableDisclosure(t *testing.T) {
	d := Dataset{Resources: []Resource{bead(t, "A", ast{}), link(t, "opaque", "Requires", base+"A", "urn:external:opaque#fragment")}}
	w := ast{"op": "walk", "roots": refset("beads", base+"A"), "direction": "outgoing"}
	response := execute(t, request(t, "query", "resources", w), d)
	assertIDs(t, response, "opaque")
	if response.Evaluation.Traversals[0].Depth != 1 {
		t.Fatalf("missing depth did not mean one: %+v", response.Evaluation)
	}
	w["depth"] = 2
	_, err := Execute(context.Background(), request(t, "query", "resources", w), d, Limits{})
	var failure *Failure
	if !errors.As(err, &failure) || !errors.Is(err, ErrUnavailable) || failure.At != "" || strings.Contains(err.Error(), "external") || strings.Contains(err.Error(), "opaque#") {
		t.Fatalf("unavailable identity leaked or misclassified: %v", err)
	}
	// Normalization is defensive: mutating an exported Resource never bypasses
	// admission merely because NewResource previously accepted it.
	d.Resources[0].Data = json.RawMessage(`{"properties":{}}`)
	assertFailure(t, request(t, "query", "resources", scan("beads")), d, ErrDataset)
}

func TestPinnedRefsRefuseBeforeDatasetAndSkippedBranches(t *testing.T) {
	for _, collection := range []string{"beads", "links"} {
		for _, id := range []string{base + "A", base + "unknown"} {
			for _, pin := range []string{"r1", "retained-old", "current"} {
				raw := request(t, "query", "resources", refset(collection, ast{"uri": id, "revision": pin}))
				if _, err := Decode(raw); !errors.Is(err, ErrUnsupported) {
					t.Fatalf("pinned refs admitted before acquisition: %v", err)
				}
				// Invalid data must never be inspected to classify an unsupported
				// explicit exact read, regardless of current/unknown token spelling.
				assertFailure(t, raw, Dataset{Resources: []Resource{{Kind: graph.KindBead, Data: json.RawMessage(`null`)}}}, ErrUnsupported)
			}
		}
	}
	pinned := refset("beads", ast{"uri": base + "A", "revision": "r1"})
	skipped := ast{"op": "or", "operands": []any{true, ast{"op": "any", "input": pinned}}}
	raw := request(t, "query", "resources", filter(scan("beads"), "b", skipped))
	if _, err := Decode(raw); !errors.Is(err, ErrUnsupported) {
		t.Fatalf("skipped branch bypassed exact-acquisition admission: %v", err)
	}
	// Depth zero does not bypass the pinned-root capability boundary either.
	if _, err := Decode(request(t, "graph", "graph", walk(pinned, "outgoing", 0))); !errors.Is(err, ErrUnsupported) {
		t.Fatalf("pinned graph root admitted: %v", err)
	}
}

func TestPinnedEndpointsUniformlyUnsupportedButStoredPinsStayData(t *testing.T) {
	for _, id := range []string{base + "B", base + "unknown", "https://foreign.example/B"} {
		for _, pin := range []string{"r1", "retained-old", "current"} {
			stored := ast{"uri": id, "revision": pin}
			d := Dataset{Resources: []Resource{bead(t, "A", ast{}), bead(t, "B", ast{}), link(t, "pinned", "Requires", base+"A", stored)}}
			w := walk(refset("beads", base+"A"), "outgoing", 1)
			assertIDs(t, execute(t, request(t, "query", "resources", w), d), "pinned")
			rawTest := filter(scan("links"), "l", compare("eq", field(binding("l"), "target"), literal(stored)))
			assertIDs(t, execute(t, request(t, "query", "resources", rawTest), d), "pinned")
			for _, selection := range []ast{
				{"op": "endpoints", "input": refset("links", base+"pinned"), "side": "target"},
				filter(scan("links"), "l", compare("eq", ast{"op": "id", "of": endpoint(binding("l"), "target")}, literal(base+"B"))),
				walk(refset("beads", base+"A"), "outgoing", 2),
			} {
				raw := request(t, "query", "resources", selection)
				assertFailure(t, raw, d, ErrUnsupported)
				_, err := Execute(context.Background(), raw, d, Limits{})
				var failure *Failure
				if !errors.As(err, &failure) || failure.At != "" || errors.Is(err, ErrUnavailable) {
					t.Fatalf("exact acquisition leaked existence or became unavailable: %v", err)
				}
			}
			assertFailure(t, request(t, "query", "graph", w), d, ErrUnsupported)
		}
	}
	// Even a selected token literally spelled current does not grant exact
	// acquisition; a floating Reference still selects that actual current body.
	d := Dataset{Resources: []Resource{resource(t, graph.KindBead, "A", "Note", "current", ast{}), link(t, "self", "Requires", base+"A", ast{"uri": base + "A", "revision": "current"})}}
	assertIDs(t, execute(t, request(t, "query", "resources", walk(refset("beads", base+"A"), "incoming", 1)), d), "self")
	assertFailure(t, request(t, "query", "resources", ast{"op": "endpoints", "input": scan("links"), "side": "target"}), d, ErrUnsupported)
}

// SUP-IR-003: wrong JSON kind is malformed input; an unknown string is a
// well-shaped identifier that this implementation does not support.
func TestIRKindAndUnsupportedIdentifier(t *testing.T) {
	for _, tc := range []struct {
		name string
		ir   any
		want error
	}{
		{"numeric kind", 17, ErrInvalid},
		{"null kind", nil, ErrInvalid},
		{"unknown string", IR + "-future", ErrUnsupported},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := json.Marshal(ast{"ir": tc.ir, "profile": "query", "selection": scan("beads"), "result": ast{"form": "count"}})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := Decode(raw); !errors.Is(err, tc.want) {
				t.Fatalf("IR classification = %v, want %v", err, tc.want)
			}
		})
	}
}

// Adapter inspection is syntax-only: all branches are visited, but literals
// containing an "op" member do not become executable field accesses.
func TestRequestFieldRootsAreCompleteAndDetached(t *testing.T) {
	raw := []byte(`{"ir":"bdp-query-design-1","profile":"list","selection":{"op":"filter","input":{"op":"scan","collection":"beads"},"bind":"b","predicate":{"op":"or","operands":[true,{"op":"present","value":{"op":"field","of":{"op":"binding","name":"b"},"path":["revisionProvenance"]}},{"op":"compare","operator":"eq","left":{"op":"field","of":{"op":"binding","name":"b"},"path":["properties","ownedLinks"]},"right":{"op":"literal","value":{"op":"field","path":["fakeRoot"]}}}]}},"result":{"form":"resources"}}`)
	request, err := Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	roots := request.FieldRoots()
	if len(roots) != 2 || roots[0] != "revisionProvenance" || roots[1] != "properties" {
		t.Fatalf("field roots: %v", roots)
	}
	roots[0] = "mutated"
	if request.FieldRoots()[0] != "revisionProvenance" {
		t.Fatal("inspection mutated admitted request")
	}
	var absent *Request
	if absent.FieldRoots() != nil {
		t.Fatal("nil request invented field roots")
	}
}

// A List AST is a bounded candidate predicate, even when its syntax is shared
// with Query. Prove each refused request is valid Query syntax first, so an
// unrelated parser error cannot stand in for enforcement of the List profile.
func TestListProfileRejectsGraphOperationsRecursively(t *testing.T) {
	l := binding("l")
	bead := binding("b")
	dereference := compare("eq", field(endpoint(l, "target"), "properties", "status"), literal("active"))
	for _, tc := range []struct {
		name      string
		selection ast
	}{
		{"walk root", walk(refset("beads", base+"A"), "outgoing", 1)},
		{"endpoint collection", ast{"op": "endpoints", "input": scan("links"), "side": "target"}},
		{"explicit reference set", refset("beads", base+"A")},
		{"nested selection", filter(filter(scan("beads"), "a", true), "b", true)},
		{"global any", filter(scan("beads"), "b", ast{"op": "any", "input": scan("links"), "bind": "l", "predicate": true})},
		{"global all", filter(scan("beads"), "b", ast{"op": "all", "input": scan("links"), "bind": "l", "predicate": true})},
		{"endpoint property", filter(scan("links"), "l", dereference)},
		{"endpoint type", filter(scan("links"), "l", ast{"op": "typeIs", "of": endpoint(l, "target"), "type": base + "Memory", "match": "family"})},
		{"endpoint identity", filter(scan("links"), "l", compare("eq", ast{"op": "id", "of": endpoint(l, "source")}, literal(base+"A")))},
		{"skipped disjunction", filter(scan("links"), "l", ast{"op": "or", "operands": []any{true, dereference}})},
		{"skipped conjunction", filter(scan("links"), "l", ast{"op": "and", "operands": []any{false, ast{"op": "not", "predicate": dereference}}})},
		{"nested skipped any", filter(scan("beads"), "b", ast{"op": "or", "operands": []any{true, ast{"op": "not", "predicate": ast{"op": "any", "input": scan("links"), "bind": "l", "predicate": true}}}})},
		{"field to field", filter(scan("beads"), "b", compare("eq", field(bead, "properties", "title"), field(bead, "properties", "body")))},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Decode(request(t, "query", "resources", tc.selection)); err != nil {
				t.Fatalf("fixture is not an admitted Query request: %v", err)
			}
			if _, err := Decode(request(t, "list", "resources", tc.selection)); !errors.Is(err, ErrInvalid) {
				t.Fatalf("List admitted a nonlocal shape or returned the wrong failure: %v", err)
			}
		})
	}
}

func TestListProfileKeepsStoredReferencesAndLiteralDataLocal(t *testing.T) {
	l := binding("l")
	for _, predicate := range []ast{
		compare("eq", field(l, "source"), literal(base+"A")),
		compare("eq", field(l, "target", "revision"), literal("r1")),
		compare("eq", field(l, "metadata", "payload"), literal(ast{"op": "endpoint", "side": "target"})),
		{"op": "typeIs", "of": l, "type": base + "References", "match": "family"},
	} {
		if _, err := Decode(request(t, "list", "resources", filter(scan("links"), "l", predicate))); err != nil {
			t.Fatalf("candidate data or literal syntax was mistaken for graph evaluation: %v", err)
		}
	}
}
