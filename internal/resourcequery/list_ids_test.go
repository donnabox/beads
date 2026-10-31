package resourcequery

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	graph "github.com/steveyegge/beads/graphops"
)

func selectionBytes(t *testing.T, selection ast) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(selection)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
func listSelector(t *testing.T, selection ast, limits Limits) *ListIDSelector {
	t.Helper()
	s, err := CompileListIDs(selectionBytes(t, selection), limits)
	if err != nil {
		t.Fatal(err)
	}
	return s
}
func stagedCandidate(t *testing.T, id string, props ast) ListCandidate {
	t.Helper()
	return ListCandidate{Data: selectionBytes(t, ast{"id": base + id, "type": base + "Note", "properties": props})}
}
func listInventory(candidates ...ListCandidate) ListInventory {
	if candidates == nil {
		candidates = []ListCandidate{}
	}
	return ListInventory{Kind: graph.KindBead, Candidates: candidates,
		TypeKinds:    map[string]graph.ResourceKind{base + "Note": graph.KindBead, base + "Root": graph.KindBead},
		TypeFamilies: map[string][]string{base + "Note": {base + "Note", base + "Root"}, base + "Root": {base + "Root"}}}
}
func listIDs(t *testing.T, s *ListIDSelector, input ListInventory, want ...string) {
	t.Helper()
	got, err := s.Select(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if want == nil {
		want = []string{}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("IDs = %v, want %v", got, want)
	}
}
func listFailure(t *testing.T, s *ListIDSelector, input ListInventory, want error) {
	t.Helper()
	got, err := s.Select(context.Background(), input)
	if !errors.Is(err, want) || got != nil {
		t.Fatalf("IDs = %v, error = %v, want nil IDs and %v", got, err, want)
	}
}

func TestListIDsRevisionedQueryParity(t *testing.T) {
	// These are honest complete records accepted by Query, with actual revisions.
	// Only their detached staged views omit those revisions; no Resource is faked.
	records := []Resource{
		resource(t, graph.KindBead, "c", "Note", "real-c", ast{"n": json.Number("2.0"), "nil": nil, "title": "café猫", "tags": []any{"x", nil, json.Number("2")}, "revision": "property-r", "object": ast{"x": 1}}),
		resource(t, graph.KindBead, "a", "Note", "real-a", ast{"n": 1, "title": "CAFÉ猫", "tags": []any{}, "revision": "other"}),
		resource(t, graph.KindBead, "b", "Note", "real-b", ast{"n": "2", "title": nil}),
	}
	inventory := listInventory()
	// Preserve the supported pinned Type-reference representation in both paths.
	var pinned ast
	if err := json.Unmarshal(records[0].Data, &pinned); err != nil {
		t.Fatal(err)
	}
	pinned["type"] = ast{"uri": base + "Note", "revision": "actual-type-r"}
	records[0].Data = selectionBytes(t, pinned)
	for _, r := range records {
		var body ast
		if err := json.Unmarshal(r.Data, &body); err != nil {
			t.Fatal(err)
		}
		delete(body, "revision")
		inventory.Candidates = append(inventory.Candidates, ListCandidate{Data: selectionBytes(t, body)})
	}
	b := binding("x")
	f := func(parts ...any) ast { return field(b, append([]any{"properties"}, parts...)...) }
	cases := []struct {
		name      string
		predicate any
		want      []string
	}{
		{"true", true, []string{"a", "b", "c"}},
		{"false", false, nil},
		{"numeric equality", compare("eq", f("n"), literal(json.Number("2e0"))), []string{"c"}},
		{"numeric order", compare("gt", f("n"), literal(1)), []string{"c"}},
		{"string and number differ", compare("eq", f("n"), literal("2")), []string{"b"}},
		{"different ordered types", compare("lt", f("n"), literal("2")), nil},
		{"missing is not null", compare("eq", f("missing"), literal(nil)), nil},
		{"explicit null", compare("eq", f("nil"), literal(nil)), []string{"c"}},
		{"missing ne null", compare("ne", f("nil"), literal(nil)), []string{"a", "b"}},
		{"present null", ast{"op": "present", "value": f("nil")}, []string{"c"}},
		{"missing absent", ast{"op": "present", "value": f("missing")}, nil},
		{"indexed null", compare("eq", f("tags", 1), literal(nil)), []string{"c"}},
		{"out of range", ast{"op": "present", "value": f("tags", 9)}, nil},
		{"wrong container", ast{"op": "present", "value": f("n", "x")}, nil},
		{"case sensitive contains", ast{"op": "contains", "text": f("title"), "substring": literal("fé猫")}, []string{"c"}},
		{"nonstring contains", ast{"op": "contains", "text": f("n"), "substring": literal("2")}, []string{"b"}},
		{"numeric membership", ast{"op": "member", "value": literal(json.Number("2.0")), "array": f("tags")}, []string{"c"}},
		{"null membership", ast{"op": "member", "value": f("nil"), "array": literal([]any{nil})}, []string{"c"}},
		{"object membership", ast{"op": "member", "value": f("object"), "array": literal([]any{ast{"x": json.Number("1.0")}})}, []string{"c"}},
		{"id set", ast{"op": "member", "value": ast{"op": "id", "of": b}, "array": literal([]any{base + "c", base + "a", base + "a"})}, []string{"a", "c"}},
		{"metadata", compare("eq", field(b, "metadata", "owner"), literal("Donna")), []string{"a", "b", "c"}},
		{"ordinary property revision", compare("eq", f("revision"), literal("property-r")), []string{"c"}},
		{"raw type pin", compare("eq", field(b, "type", "revision"), literal("actual-type-r")), []string{"c"}},
		{"exact type", ast{"op": "typeIs", "of": b, "type": base + "Note", "match": "exact"}, []string{"a", "b", "c"}},
		{"not and or", ast{"op": "and", "operands": []any{ast{"op": "not", "predicate": false}, ast{"op": "or", "operands": []any{false, compare("eq", f("n"), literal(1))}}}}, []string{"a"}},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			sel := filter(scan("beads"), "x", tt.predicate)
			old := execute(t, request(t, "list", "resources", sel), Dataset{Resources: records, TypeKinds: inventory.TypeKinds, TypeFamilies: inventory.TypeFamilies})
			assertIDs(t, old, tt.want...)
			s := listSelector(t, sel, Limits{})
			got, err := s.Select(context.Background(), inventory)
			if err != nil {
				t.Fatal(err)
			}
			var oldIDs []string
			for _, raw := range old.Result.(ResourcesResult).Items {
				var x struct{ ID string }
				if err := json.Unmarshal(raw, &x); err != nil {
					t.Fatal(err)
				}
				oldIDs = append(oldIDs, x.ID)
			}
			if oldIDs == nil {
				oldIDs = []string{}
			}
			if !reflect.DeepEqual(got, oldIDs) {
				t.Fatalf("selector %v != Query %v", got, oldIDs)
			}
		})
	}
}

// Nominal family parity requires unpinned affiliations. A pinned affiliation
// needs exact captured ancestry, which ListInventory deliberately cannot carry.
func TestListIDsNominalFamilyQueryParity(t *testing.T) {
	records := []Resource{
		resource(t, graph.KindBead, "a", "Note", "real-a", ast{}),
		resource(t, graph.KindBead, "b", "Root", "real-b", ast{}),
	}
	input := listInventory()
	for _, r := range records {
		var body ast
		if err := json.Unmarshal(r.Data, &body); err != nil {
			t.Fatal(err)
		}
		delete(body, "revision")
		input.Candidates = append(input.Candidates, ListCandidate{Data: selectionBytes(t, body)})
	}
	selection := filter(scan("beads"), "x", ast{"op": "typeIs", "of": binding("x"), "type": base + "Root", "match": "family"})
	assertIDs(t, execute(t, request(t, "list", "resources", selection), Dataset{Resources: records, TypeKinds: input.TypeKinds, TypeFamilies: input.TypeFamilies}), "a", "b")
	listIDs(t, listSelector(t, selection, Limits{}), input, base+"a", base+"b")
}

func TestListIDsPinnedFamilyNeverBorrowsNominalAuthority(t *testing.T) {
	body := ast{"id": base + "a", "type": ast{"uri": base + "Note", "revision": "retained-type"}, "properties": ast{}}
	input := listInventory(ListCandidate{Data: selectionBytes(t, body)})
	body["revision"] = "real-a"
	r, err := NewResource(graph.KindBead, selectionBytes(t, body))
	if err != nil {
		t.Fatal(err)
	}
	data := Dataset{Resources: []Resource{r}, TypeKinds: input.TypeKinds, TypeFamilies: input.TypeFamilies}
	family := ast{"op": "typeIs", "of": binding("x"), "type": base + "Root", "match": "family"}
	for _, tc := range []struct {
		name      string
		predicate any
	}{
		{"direct", family},
		{"or short circuit", ast{"op": "or", "operands": []any{true, family}}},
		{"and short circuit", ast{"op": "and", "operands": []any{false, family}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			selection := filter(scan("beads"), "x", tc.predicate)
			assertFailure(t, request(t, "list", "resources", selection), data, ErrUnsupported)
			listFailure(t, listSelector(t, selection, Limits{}), input, ErrUnsupported)
		})
	}
	// Exact identity and raw pin inspection do not require family evidence.
	for _, predicate := range []any{
		ast{"op": "typeIs", "of": binding("x"), "type": base + "Note", "match": "exact"},
		compare("eq", field(binding("x"), "type", "revision"), literal("retained-type")),
	} {
		selection := filter(scan("beads"), "x", predicate)
		assertIDs(t, execute(t, request(t, "list", "resources", selection), data), "a")
		listIDs(t, listSelector(t, selection, Limits{}), input, base+"a")
	}
}

func TestListIDsTrulyRevisionlessAndRecaptured(t *testing.T) {
	a := stagedCandidate(t, "a", ast{"state": "old"})
	b := stagedCandidate(t, "b", ast{"state": "ready", "revision": "user-value"})
	if _, err := NewResource(graph.KindBead, b.Data); !errors.Is(err, ErrDataset) {
		t.Fatalf("complete Resource accepted staged candidate: %v", err)
	}
	sel := filter(scan("beads"), "x", compare("eq", field(binding("x"), "properties", "state"), literal("ready")))
	s := listSelector(t, sel, Limits{})
	if s.ResourceKind() != graph.KindBead {
		t.Fatal("wrong kind")
	}
	listIDs(t, s, listInventory(a))
	listIDs(t, s, listInventory(a, b), base+"b") // earlier create
	a = stagedCandidate(t, "a", ast{"state": "ready"})
	listIDs(t, s, listInventory(b, a), base+"a", base+"b") // earlier update; deterministic order
	listIDs(t, s, listInventory(a), base+"a")              // earlier delete
	listIDs(t, s, listInventory())                         // explicitly complete empty capture
}

func TestListIDsRejectsEntireUnsupportedTree(t *testing.T) {
	b := binding("x")
	cases := []struct {
		name      string
		selection ast
	}{
		{"revision operator", filter(scan("beads"), "x", compare("eq", ast{"op": "rev", "of": b}, literal("r")))},
		{"root revision", filter(scan("beads"), "x", ast{"op": "or", "operands": []any{true, compare("eq", field(b, "revision"), literal(nil))}})},
		{"provenance", filter(scan("beads"), "x", ast{"op": "and", "operands": []any{false, ast{"op": "present", "value": field(b, "provenance")}}})},
		{"source path", filter(scan("links"), "x", ast{"op": "present", "value": field(b, "source")})},
		{"target path", filter(scan("links"), "x", compare("eq", field(b, "target"), literal(nil)))},
		{"unknown root", filter(scan("beads"), "x", ast{"op": "present", "value": field(b, "future")})},
		{"numeric root", filter(scan("beads"), "x", ast{"op": "present", "value": field(b, 0)})},
		{"whole resource field", filter(scan("beads"), "x", ast{"op": "present", "value": ast{"op": "field", "of": b, "path": []any{}}})},
		{"whole resource comparison", filter(scan("beads"), "x", compare("eq", b, b))},
		{"endpoint expression", filter(scan("links"), "x", compare("eq", ast{"op": "id", "of": endpoint(b, "source")}, literal(base+"a")))},
		{"quantifier", filter(scan("beads"), "x", ast{"op": "or", "operands": []any{true, ast{"op": "any", "input": scan("beads")}}})},
		{"walk", walk(refset("beads", base+"a"), "outgoing", 0)},
		{"acquisition", refset("beads", base+"a")},
		{"endpoints", ast{"op": "endpoints", "input": scan("links"), "side": "source"}},
		{"nested filter", filter(filter(scan("beads"), "y", true), "x", true)},
		{"full Query", ast{"ir": IR, "profile": "list", "selection": scan("beads"), "result": ast{"form": "resources"}}},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			s, err := CompileListIDs(selectionBytes(t, tt.selection), Limits{})
			if err == nil || s != nil {
				t.Fatalf("compiled unsupported tree: %v, %v", s, err)
			}
		})
	}
}

func TestListIDsClosedCandidateValidation(t *testing.T) {
	valid := string(stagedCandidate(t, "a", ast{}).Data)
	cases := []string{
		`null`, `[]`, `{}`, valid + ` {}`,
		strings.Replace(valid, `"properties":{}`, `"properties":null`, 1),
		strings.Replace(valid, `"properties":{}`, `"properties":[]`, 1),
		strings.Replace(valid, `"properties":{}`, `"properties":{},"metadata":null`, 1),
		strings.Replace(valid, `"properties":{}`, `"properties":{},"revision":""`, 1),
		strings.Replace(valid, `"properties":{}`, `"properties":{},"revision":"real-but-forbidden"`, 1),
		strings.Replace(valid, `"properties":{}`, `"properties":{},"source":"https://query.example/a"`, 1),
		strings.Replace(valid, `"properties":{}`, `"properties":{},"target":null`, 1),
		strings.Replace(valid, `"properties":{}`, `"properties":{},"provenance":{}`, 1),
		strings.Replace(valid, `"properties":{}`, `"properties":{},"future":true`, 1),
		strings.Replace(valid, `"properties":{}`, `"properties":{"n":9007199254740993}`, 1),
		strings.Replace(valid, `"properties":{}`, `"properties":{"x":1,"x":2}`, 1),
		strings.Replace(valid, `"id":"https://query.example/a"`, `"id":"relative"`, 1),
		strings.Replace(valid, `"id":"https://query.example/a"`, `"id":"https://query.example/a?revision=r"`, 1),
		strings.Replace(valid, `"type":"https://query.example/Note"`, `"type":{"uri":"https://query.example/Note"}`, 1),
		strings.Replace(valid, `"type":"https://query.example/Note"`, `"type":{"uri":"https://query.example/Note","revision":""}`, 1),
		strings.Replace(valid, `"type":"https://query.example/Note"`, `"type":{"uri":"https://query.example/Note","revision":"r","future":true}`, 1),
	}
	s := listSelector(t, filter(scan("beads"), "x", false), Limits{})
	for i, raw := range cases {
		t.Run(strings.Join([]string{"candidate", jsonNumber(i)}, "-"), func(t *testing.T) {
			listFailure(t, s, listInventory(ListCandidate{Data: json.RawMessage(raw)}), ErrDataset)
		})
	}
	listFailure(t, s, listInventory(stagedCandidate(t, "a", ast{}), stagedCandidate(t, "a", ast{})), ErrDataset)
}

func jsonNumber(i int) string { raw, _ := json.Marshal(i); return string(raw) }

func TestListIDsAuthorityAndCaptureFailures(t *testing.T) {
	s := listSelector(t, scan("beads"), Limits{})
	cases := []struct {
		name   string
		change func(*ListInventory)
		want   error
	}{
		{"missing capture", func(x *ListInventory) { x.Candidates = nil }, ErrUnavailable},
		{"wrong kind", func(x *ListInventory) { x.Kind = graph.KindLink }, ErrDataset},
		{"invalid kind", func(x *ListInventory) { x.Kind = "future" }, ErrDataset},
		{"unavailable kinds", func(x *ListInventory) { x.TypeKinds = nil }, ErrDataset},
		{"unavailable families", func(x *ListInventory) { x.TypeFamilies = nil }, ErrDataset},
		{"missing candidate type", func(x *ListInventory) { delete(x.TypeKinds, base+"Note"); delete(x.TypeFamilies, base+"Note") }, ErrDataset},
		{"wrong candidate type kind", func(x *ListInventory) {
			x.TypeKinds[base+"Note"] = graph.KindLink
			x.TypeFamilies[base+"Note"] = []string{base + "Note"}
		}, ErrDataset},
		{"missing family", func(x *ListInventory) { delete(x.TypeFamilies, base+"Note") }, ErrDataset},
		{"nonreflexive", func(x *ListInventory) { x.TypeFamilies[base+"Note"] = []string{base + "Root"} }, ErrDataset},
		{"duplicate ancestor", func(x *ListInventory) { x.TypeFamilies[base+"Note"] = []string{base + "Note", base + "Note"} }, ErrDataset},
		{"unknown ancestor", func(x *ListInventory) { x.TypeFamilies[base+"Note"] = []string{base + "Note", base + "Missing"} }, ErrDataset},
		{"wrong ancestor kind", func(x *ListInventory) { x.TypeKinds[base+"Root"] = graph.KindLink }, ErrDataset},
		{"orphan family", func(x *ListInventory) { x.TypeFamilies[base+"Missing"] = []string{base + "Missing"} }, ErrDataset},
		{"incomplete transitive ancestry", func(x *ListInventory) {
			x.TypeKinds[base+"Top"] = graph.KindBead
			x.TypeFamilies[base+"Top"] = []string{base + "Top"}
			x.TypeFamilies[base+"Root"] = []string{base + "Root", base + "Top"}
		}, ErrDataset},
		{"invalid authority URI", func(x *ListInventory) {
			x.TypeKinds["relative"] = graph.KindBead
			x.TypeFamilies["relative"] = []string{"relative"}
		}, ErrDataset},
		{"invalid authority kind", func(x *ListInventory) { x.TypeKinds[base+"Root"] = "future" }, ErrDataset},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			input := listInventory(stagedCandidate(t, "a", ast{}))
			tt.change(&input)
			listFailure(t, s, input, tt.want)
		})
	}
	// Missing families refuse even if no candidate would evaluate the test.
	for _, match := range []string{"family"} {
		t.Run("unreachable "+match, func(t *testing.T) {
			p := ast{"op": "or", "operands": []any{true, ast{"op": "typeIs", "of": binding("x"), "type": base + "Missing", "match": match}}}
			listFailure(t, listSelector(t, filter(scan("beads"), "x", p), Limits{}), listInventory(), ErrUnavailable)
			p = ast{"op": "and", "operands": []any{false, ast{"op": "typeIs", "of": binding("x"), "type": base + "Root", "match": match}}}
			input := listInventory()
			input.TypeKinds[base+"Root"] = graph.KindLink
			input.TypeFamilies[base+"Root"] = []string{base + "Root"}
			input.TypeFamilies[base+"Note"] = []string{base + "Note"}
			listFailure(t, listSelector(t, filter(scan("beads"), "x", p), Limits{}), input, ErrInvalid)
		})
	}
}

func TestListIDsLinkViewsNeedNoEndpoints(t *testing.T) {
	candidate := stagedCandidate(t, "link", ast{"role": "author"})
	var body ast
	if err := json.Unmarshal(candidate.Data, &body); err != nil {
		t.Fatal(err)
	}
	body["type"] = base + "Authorship"
	input := ListInventory{Kind: graph.KindLink, Candidates: []ListCandidate{{Data: selectionBytes(t, body)}}, TypeKinds: map[string]graph.ResourceKind{base + "Authorship": graph.KindLink}, TypeFamilies: map[string][]string{base + "Authorship": {base + "Authorship"}}}
	s := listSelector(t, filter(scan("links"), "x", compare("eq", field(binding("x"), "properties", "role"), literal("author"))), Limits{})
	listIDs(t, s, input, base+"link")
}

func TestListIDsLimitsAndNoPartialResults(t *testing.T) {
	input := listInventory(stagedCandidate(t, "b", ast{}), stagedCandidate(t, "a", ast{}))
	l := DefaultLimits()
	l.MaxResults = 1
	listFailure(t, listSelector(t, scan("beads"), l), input, ErrLimit)
	l = DefaultLimits()
	l.MaxWork = 1
	listFailure(t, listSelector(t, scan("beads"), l), input, ErrLimit)
	for _, mutate := range []func(*Limits){
		func(l *Limits) { l.RequestBytes = 1 }, func(l *Limits) { l.Nodes = 1 }, func(l *Limits) { l.Nesting = 1 }, func(l *Limits) { l.MaxWork = 0 },
	} {
		l := DefaultLimits()
		mutate(&l)
		s, err := CompileListIDs(selectionBytes(t, filter(scan("beads"), "x", true)), l)
		if err == nil || s != nil {
			t.Fatalf("limits accepted: %+v", l)
		}
	}
	s := listSelector(t, scan("beads"), Limits{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if got, err := s.Select(ctx, input); !errors.Is(err, context.Canceled) || got != nil {
		t.Fatalf("canceled: %v %v", got, err)
	}
	if got, err := s.Select(nil, input); !errors.Is(err, ErrInvalid) || got != nil {
		t.Fatalf("nil context: %v %v", got, err)
	}
	var zero ListIDSelector
	listFailure(t, &zero, input, ErrInvalid)
	var absent *ListIDSelector
	if absent.ResourceKind() != "" || zero.ResourceKind() != "" {
		t.Fatal("uncompiled kind")
	}
	listFailure(t, absent, input, ErrInvalid)
}

func TestListIDsCompileOwnsSelectionAndSelectOwnsCandidates(t *testing.T) {
	raw := selectionBytes(t, scan("beads"))
	s, err := CompileListIDs(raw, Limits{})
	if err != nil {
		t.Fatal(err)
	}
	raw[0] = 'x'
	input := listInventory(stagedCandidate(t, "a", ast{"x": 1}))
	before := append([]byte(nil), input.Candidates[0].Data...)
	listIDs(t, s, input, base+"a")
	if string(before) != string(input.Candidates[0].Data) {
		t.Fatal("Select mutated caller candidate")
	}
	listIDs(t, s, input, base+"a")
}

func TestListIDsCoreOnlyAuthorityParity(t *testing.T) {
	candidate := stagedCandidate(t, "a", ast{"revision": "ordinary", "source": "ordinary", "provenance": nil})
	input := ListInventory{Kind: graph.KindBead, Candidates: []ListCandidate{candidate}}
	baseline := Dataset{Resources: []Resource{resource(t, graph.KindBead, "a", "Note", "real-r", ast{"revision": "ordinary", "source": "ordinary", "provenance": nil})}}
	b := binding("x")
	for _, selection := range []ast{
		scan("beads"),
		filter(scan("beads"), "x", compare("eq", field(b, "properties", "revision"), literal("ordinary"))),
		filter(scan("beads"), "x", ast{"op": "present", "value": field(b, "properties", "provenance")}),
		filter(scan("beads"), "x", ast{"op": "typeIs", "of": b, "type": base + "Note", "match": "exact"}),
	} {
		assertIDs(t, execute(t, request(t, "list", "resources", selection), baseline), "a")
		listIDs(t, listSelector(t, selection, Limits{}), input, base+"a")
	}
	// An exact test on an unknown Type is false, as in Query, without requiring
	// Types authority or substituting a weaker interpretation for family tests.
	selection := filter(scan("beads"), "x", ast{"op": "typeIs", "of": b, "type": base + "Unknown", "match": "exact"})
	assertIDs(t, execute(t, request(t, "list", "resources", selection), baseline))
	listIDs(t, listSelector(t, selection, Limits{}), input)
	listIDs(t, listSelector(t, selection, Limits{}), listInventory(candidate))
	for _, predicate := range []any{
		ast{"op": "typeIs", "of": b, "type": base + "Note", "match": "family"},
		ast{"op": "or", "operands": []any{true, ast{"op": "typeIs", "of": b, "type": base + "Note", "match": "family"}}},
	} {
		s := listSelector(t, filter(scan("beads"), "x", predicate), Limits{})
		listFailure(t, s, input, ErrUnsupported)
		listFailure(t, s, ListInventory{Kind: graph.KindBead, Candidates: []ListCandidate{}}, ErrUnsupported)
	}
	listIDs(t, listSelector(t, scan("beads"), Limits{}), ListInventory{Kind: graph.KindBead, Candidates: []ListCandidate{}})
}

// cancelAtContext cancels at a deterministic context observation, allowing a
// check after the first identity has matched without timer or scheduler races.
type cancelAtContext struct {
	context.Context
	calls, cancelAt int
}

func (c *cancelAtContext) Err() error {
	c.calls++
	if c.calls >= c.cancelAt {
		return context.Canceled
	}
	return nil
}
func TestListIDsCancellationAndWorkAfterMatchDiscardIDs(t *testing.T) {
	input := ListInventory{Kind: graph.KindBead, Candidates: []ListCandidate{stagedCandidate(t, "a", ast{}), stagedCandidate(t, "b", ast{})}}
	s := listSelector(t, scan("beads"), Limits{})
	ctx := &cancelAtContext{Context: context.Background(), cancelAt: 5}
	got, err := s.Select(ctx, input)
	if !errors.Is(err, context.Canceled) || got != nil || ctx.calls != 5 {
		t.Fatalf("cancellation after match: IDs %v, err %v, calls %d", got, err, ctx.calls)
	}
	l := DefaultLimits()
	l.MaxWork = 3 // two parsed views, one emitted ID, then fail
	listFailure(t, listSelector(t, scan("beads"), l), input, ErrLimit)
}

func TestListIDsRejectsMalformedExpandedSelections(t *testing.T) {
	for _, raw := range []json.RawMessage{
		json.RawMessage(`{"op":"scan","collection":"beads","collection":"links"}`),
		json.RawMessage(`{"op":"scan","collection":"beads"} {}`),
		selectionBytes(t, filter(scan("beads"), "x", compare("eq", field(binding("x"), "properties", "n"), literal(json.Number("9007199254740993"))))),
		selectionBytes(t, filter(scan("beads"), "x", ast{"op": "externalScalar", "step": "earlier"})),
	} {
		s, err := CompileListIDs(raw, Limits{})
		if err == nil || s != nil {
			t.Fatalf("malformed expanded selection compiled: %s", raw)
		}
	}
}
