package main

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"sort"
	"testing"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/httpapi/bdpwire"
	"github.com/steveyegge/beads/internal/resourcequery"
)

const canonicalGraphFixtureScope = "https://graph.example/"

func canonicalGraphOptionsForTest(roots ...string) canonicalGraphRequestOptions {
	options := canonicalGraphRequestOptions{Direction: "outgoing", Depth: json.RawMessage(`"unbounded"`)}
	for _, root := range roots {
		options.Roots = append(options.Roots, bdpwire.Reference{URI: canonicalGraphFixtureScope + root})
	}
	return options
}
func canonicalGraphRequestForTest(t *testing.T, options canonicalGraphRequestOptions) *resourcequery.Request {
	t.Helper()
	request, err := buildCanonicalGraphRequest(options)
	if err != nil {
		t.Fatal(err)
	}
	return request
}
func canonicalGraphFixtureForTest(t *testing.T) resourcequery.Dataset {
	t.Helper()
	data := resourcequery.Dataset{TypeKinds: map[string]graph.ResourceKind{}, TypeFamilies: map[string][]string{}}
	for _, typ := range []string{"Requires", "Defines", "Reference"} {
		uri := canonicalGraphFixtureScope + typ
		data.TypeKinds[uri], data.TypeFamilies[uri] = graph.KindLink, []string{uri}
	}
	data.TypeFamilies[canonicalGraphFixtureScope+"Defines"] = []string{canonicalGraphFixtureScope + "Defines", canonicalGraphFixtureScope + "Requires"}
	appendResource := func(kind graph.ResourceKind, body map[string]any) {
		t.Helper()
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		resource, err := resourcequery.NewResource(kind, raw)
		if err != nil {
			t.Fatal(err)
		}
		data.Resources = append(data.Resources, resource)
	}
	for _, id := range []string{"A", "B", "C", "X", "D"} {
		appendResource(graph.KindBead, map[string]any{"id": canonicalGraphFixtureScope + id, "revision": "r1", "type": canonicalGraphFixtureScope + "Note", "properties": map[string]any{"title": id}})
	}
	for _, link := range []struct{ id, typ, source, target string }{
		{"ab", "Defines", "A", "B"}, {"ab-parallel", "Requires", "A", "B"},
		{"bc", "Requires", "B", "C"}, {"ca", "Requires", "C", "A"},
		{"ax-reference", "Reference", "A", "X"}, {"xd-hidden", "Requires", "X", "D"},
	} {
		appendResource(graph.KindLink, map[string]any{"id": canonicalGraphFixtureScope + link.id, "revision": "l1", "type": canonicalGraphFixtureScope + link.typ, "source": canonicalGraphFixtureScope + link.source, "target": canonicalGraphFixtureScope + link.target, "properties": map[string]any{}})
	}
	return data
}
func canonicalGraphExecuteForTest(t *testing.T, options canonicalGraphRequestOptions, data resourcequery.Dataset) resourcequery.GraphResult {
	t.Helper()
	request := canonicalGraphRequestForTest(t, options)
	response, err := resourcequery.Execute(context.Background(), request.Bytes(), data, resourcequery.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	if !response.Evaluation.Complete || response.Evaluation.Scope != "response" {
		t.Fatalf("incomplete graph: %+v", response.Evaluation)
	}
	result, ok := response.Result.(resourcequery.GraphResult)
	if !ok || result.Form != "graph" {
		t.Fatalf("changed result shape: %#v", response.Result)
	}
	return result
}
func canonicalGraphIDsForTest(t *testing.T, records []json.RawMessage) []string {
	t.Helper()
	ids := make([]string, 0, len(records))
	for _, record := range records {
		var body struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(record, &body); err != nil {
			t.Fatal(err)
		}
		ids = append(ids, body.ID)
	}
	sort.Strings(ids)
	return ids
}
func canonicalGraphWantIDsForTest(t *testing.T, records []json.RawMessage, want ...string) {
	t.Helper()
	expected := make([]string, 0, len(want))
	for _, id := range want {
		expected = append(expected, canonicalGraphFixtureScope+id)
	}
	sort.Strings(expected)
	if got := canonicalGraphIDsForTest(t, records); !reflect.DeepEqual(got, expected) {
		t.Fatalf("IDs %v, want %v", got, expected)
	}
}

func TestCanonicalGraphRequestUsesExistingProfile(t *testing.T) {
	options := canonicalGraphOptionsForTest("A")
	options.LinkTypeURL, options.TypeMatch = canonicalGraphFixtureScope+"Requires", "family"
	request := canonicalGraphRequestForTest(t, options)
	if request.Profile() != "graph" || request.ResultForm() != "graph" || request.ResourceKind() != graph.KindLink {
		t.Fatalf("dispatch metadata: %s %s %s", request.Profile(), request.ResultForm(), request.ResourceKind())
	}
	var actual map[string]any
	if err := json.Unmarshal(request.Bytes(), &actual); err != nil {
		t.Fatal(err)
	}
	selection := actual["selection"].(map[string]any)
	if actual["ir"] != resourcequery.IR || selection["op"] != "walk" || selection["bind"] != "link" {
		t.Fatalf("not a direct received graph profile: %#v", actual)
	}
	if _, outerFilter := selection["input"]; outerFilter {
		t.Fatal("through lowered to outer selection")
	}
	// The Request owns serialized operands; later caller mutation changes none.
	before := request.Bytes()
	options.Roots[0].URI = "https://different.example/A"
	options.Depth[1] = 'X'
	exposed := request.Bytes()
	exposed[0] = 'x'
	if !reflect.DeepEqual(request.Bytes(), before) {
		t.Fatal("request aliases caller/exposed bytes")
	}
}

func TestCanonicalGraphRequestPreservesReferencesAndAdmission(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*canonicalGraphRequestOptions)
		want   error
	}{
		{"relative-root", func(o *canonicalGraphRequestOptions) { o.Roots[0].URI = "A" }, resourcequery.ErrInvalid},
		{"duplicate-root", func(o *canonicalGraphRequestOptions) { o.Roots = append(o.Roots, o.Roots[0]) }, resourcequery.ErrInvalid},
		{"static-current-looking-pin", func(o *canonicalGraphRequestOptions) { o.Roots[0].Revision = "r1" }, resourcequery.ErrUnsupported},
		{"static-opaque-pin", func(o *canonicalGraphRequestOptions) { o.Roots[0].Revision = "+ /@=" }, resourcequery.ErrUnsupported},
		{"direction", func(o *canonicalGraphRequestOptions) { o.Direction = "both" }, resourcequery.ErrInvalid},
		{"missing-depth", func(o *canonicalGraphRequestOptions) { o.Depth = nil }, resourcequery.ErrInvalid},
		{"negative-depth", func(o *canonicalGraphRequestOptions) { o.Depth = json.RawMessage(`-1`) }, resourcequery.ErrInvalid},
		{"fractional-depth", func(o *canonicalGraphRequestOptions) { o.Depth = json.RawMessage(`1.5`) }, resourcequery.ErrInvalid},
		{"malformed-depth", func(o *canonicalGraphRequestOptions) { o.Depth = json.RawMessage(`]`) }, resourcequery.ErrInvalid},
		{"missing-match", func(o *canonicalGraphRequestOptions) { o.LinkTypeURL = canonicalGraphFixtureScope + "Requires" }, resourcequery.ErrInvalid},
		{"missing-Type", func(o *canonicalGraphRequestOptions) { o.TypeMatch = "family" }, resourcequery.ErrInvalid},
		{"local-Type-name", func(o *canonicalGraphRequestOptions) { o.LinkTypeURL, o.TypeMatch = "Requires", "family" }, resourcequery.ErrInvalid},
		{"unknown-match", func(o *canonicalGraphRequestOptions) {
			o.LinkTypeURL, o.TypeMatch = canonicalGraphFixtureScope+"Requires", "refines"
		}, resourcequery.ErrInvalid},
		{"invalid-UTF8-identity", func(o *canonicalGraphRequestOptions) { o.Roots[0].URI += string([]byte{0xff}) }, resourcequery.ErrInvalid},
	} {
		t.Run(tc.name, func(t *testing.T) {
			options := canonicalGraphOptionsForTest("A")
			tc.change(&options)
			request, err := buildCanonicalGraphRequest(options)
			if !errors.Is(err, tc.want) || request != nil {
				t.Fatalf("request=%v error=%v, want %v", request, err, tc.want)
			}
		})
	}
	options := canonicalGraphOptionsForTest("A@=literal")
	request := canonicalGraphRequestForTest(t, options)
	var raw map[string]any
	if err := json.Unmarshal(request.Bytes(), &raw); err != nil {
		t.Fatal(err)
	}
	refs := raw["selection"].(map[string]any)["roots"].(map[string]any)["references"].([]any)
	if refs[0] != canonicalGraphFixtureScope+"A@=literal" {
		t.Fatalf("whole URI parsed as a CLI suffix: %#v", refs)
	}
}

func TestCanonicalGraphRequestThroughControlsExpansion(t *testing.T) {
	data := canonicalGraphFixtureForTest(t)
	options := canonicalGraphOptionsForTest("A")
	options.LinkTypeURL, options.TypeMatch = canonicalGraphFixtureScope+"Requires", "family"
	result := canonicalGraphExecuteForTest(t, options, data)
	canonicalGraphWantIDsForTest(t, result.Links, "ab", "ab-parallel", "bc", "ca")
	canonicalGraphWantIDsForTest(t, result.Beads, "A", "B", "C")
	// X's Requires link must remain unreachable behind A's rejected Reference.
	// A cycle/parallel links stay finite and retain distinct Link identities.
	if len(result.Roots) != 1 || result.Roots[0] != (resourcequery.Edition{ID: canonicalGraphFixtureScope + "A", Revision: "r1"}) || len(result.Endpoints) != 4 {
		t.Fatalf("lost graph context: %+v", result)
	}
	for _, mapping := range result.Endpoints {
		if mapping.Link.Revision != "l1" || mapping.Source.Revision != "r1" || mapping.Target.Revision != "r1" {
			t.Fatalf("edition mapping changed: %+v", mapping)
		}
	}
	options.TypeMatch = "exact"
	result = canonicalGraphExecuteForTest(t, options, data)
	canonicalGraphWantIDsForTest(t, result.Links, "ab-parallel", "bc", "ca")
	options.LinkTypeURL = canonicalGraphFixtureScope + "Defines"
	result = canonicalGraphExecuteForTest(t, options, data)
	canonicalGraphWantIDsForTest(t, result.Links, "ab")
	canonicalGraphWantIDsForTest(t, result.Beads, "A", "B")
}

func TestCanonicalGraphRequestDirectionsAndMultipleRoots(t *testing.T) {
	data := canonicalGraphFixtureForTest(t)
	for _, tc := range []struct {
		direction string
		links     []string
	}{
		{"outgoing", []string{"ab", "ab-parallel", "ax-reference"}},
		{"incoming", []string{"ca"}},
		{"connected", []string{"ab", "ab-parallel", "ax-reference", "ca"}},
	} {
		t.Run(tc.direction, func(t *testing.T) {
			options := canonicalGraphOptionsForTest("A")
			options.Direction, options.Depth = tc.direction, json.RawMessage(`1`)
			result := canonicalGraphExecuteForTest(t, options, data)
			canonicalGraphWantIDsForTest(t, result.Links, tc.links...)
		})
	}
	options := canonicalGraphOptionsForTest("A", "X")
	options.Depth = json.RawMessage(`1`)
	result := canonicalGraphExecuteForTest(t, options, data)
	canonicalGraphWantIDsForTest(t, result.Links, "ab", "ab-parallel", "ax-reference", "xd-hidden")
	canonicalGraphWantIDsForTest(t, result.Beads, "A", "B", "X", "D")
	if len(result.Roots) != 2 {
		t.Fatalf("lost requested roots: %+v", result.Roots)
	}
}

func TestCanonicalGraphRequestZeroDepthAndEmptyRoots(t *testing.T) {
	data := canonicalGraphFixtureForTest(t)
	options := canonicalGraphOptionsForTest("A")
	options.Depth = json.RawMessage(`0`)
	result := canonicalGraphExecuteForTest(t, options, data)
	canonicalGraphWantIDsForTest(t, result.Beads, "A")
	canonicalGraphWantIDsForTest(t, result.Links)
	if len(result.Roots) != 1 || len(result.Endpoints) != 0 {
		t.Fatalf("zero depth changed roots: %+v", result)
	}
	// The existing profile accepts an explicit empty root set; do not invent
	// a missing-ID CLI policy here or expand it to an implicit scan/--all.
	options.Roots = nil
	result = canonicalGraphExecuteForTest(t, options, data)
	canonicalGraphWantIDsForTest(t, result.Beads)
	canonicalGraphWantIDsForTest(t, result.Links)
	if len(result.Roots) != 0 {
		t.Fatalf("empty roots became a scan: %+v", result)
	}
}

func TestCanonicalGraphRequestNoPartialOrAuthorityFallback(t *testing.T) {
	data := canonicalGraphFixtureForTest(t)
	options := canonicalGraphOptionsForTest("A")
	options.LinkTypeURL, options.TypeMatch = canonicalGraphFixtureScope+"Requires", "family"
	request := canonicalGraphRequestForTest(t, options)
	noAuthority := data
	noAuthority.TypeKinds, noAuthority.TypeFamilies = nil, nil
	response, err := resourcequery.Execute(context.Background(), request.Bytes(), noAuthority, resourcequery.DefaultLimits())
	if !errors.Is(err, resourcequery.ErrUnsupported) || response.Result != nil {
		t.Fatalf("family silently became exact: %+v %v", response, err)
	}
	limits := resourcequery.DefaultLimits()
	limits.MaxResults = 1
	response, err = resourcequery.Execute(context.Background(), request.Bytes(), data, limits)
	if !errors.Is(err, resourcequery.ErrLimit) || response.Result != nil {
		t.Fatalf("limited graph returned partial success: %+v %v", response, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	response, err = resourcequery.Execute(ctx, request.Bytes(), data, resourcequery.DefaultLimits())
	if !errors.Is(err, context.Canceled) || response.Result != nil {
		t.Fatalf("canceled graph returned success: %+v %v", response, err)
	}
}
