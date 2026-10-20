package clioutput_test

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/clioutput"
)

const bead = `{"id":"b","type":"custom","revision":"r1","properties":{"body":"complete body","large":9007199254740993,"nullable":null,"id":"ordinary property"},"metadata":null}`

func descriptor() clioutput.Descriptor {
	return clioutput.Descriptor{
		BDPSpec:    "fixture-contract-not-a-release-pin",
		Expandable: []string{"/properties", "/metadata", "/ownedLinks"},
		Protection: &clioutput.Protection{
			Required: []string{"id", "type", "revision"},
			Members: map[string]*clioutput.Protection{
				"ownedLinks": {Items: &clioutput.Protection{
					Required: []string{"id", "type", "revision", "source", "target"},
				}},
			},
		},
	}
}

func renderer(t *testing.T, d clioutput.Descriptor, opts clioutput.Options) *clioutput.Renderer {
	t.Helper()
	r, err := clioutput.New(d, opts)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func decoded(t *testing.T, raw []byte) any {
	t.Helper()
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		t.Fatalf("invalid JSON %s: %v", raw, err)
	}
	return value
}

func equalJSON(t *testing.T, actual []byte, expected string) {
	t.Helper()
	if !reflect.DeepEqual(decoded(t, actual), decoded(t, []byte(expected))) {
		t.Fatalf("got %s\nwant %s", actual, expected)
	}
}

func TestCompleteResourceKeepsPayloadMeaning(t *testing.T) {
	r := renderer(t, descriptor(), clioutput.Options{JSON: true})
	out, err := r.Render(json.RawMessage(bead))
	if err != nil {
		t.Fatal(err)
	}
	var framed struct {
		Version    int             `json:"schema_version"`
		Contract   string          `json:"bdp_spec"`
		Kind       string          `json:"result_kind"`
		Outcome    string          `json:"outcome"`
		Data       json.RawMessage `json:"data"`
		Projection json.RawMessage `json:"projection"`
	}
	if err := json.Unmarshal(out, &framed); err != nil {
		t.Fatal(err)
	}
	if framed.Version != 2 || framed.Contract != descriptor().BDPSpec || framed.Kind != "resource" || framed.Outcome != "success" || framed.Projection != nil {
		t.Fatalf("unexpected envelope: %s", out)
	}
	equalJSON(t, framed.Data, bead)
	if !bytes.Contains(framed.Data, []byte("9007199254740993")) {
		t.Fatalf("large integer changed: %s", framed.Data)
	}
}

func TestProjectionSelectsFieldsWithoutInventingValues(t *testing.T) {
	cases := []struct {
		name    string
		include []string
		omit    []string
		want    string
	}{
		{"explicit", []string{"/properties/body"}, nil, `{"id":"b","type":"custom","revision":"r1","properties":{"body":"complete body"}}`},
		{"null", []string{"/properties/nullable", "/metadata"}, nil, `{"id":"b","type":"custom","revision":"r1","properties":{"nullable":null},"metadata":null}`},
		{"missing", []string{"/missing", "/metadata/wrong-shape"}, nil, `{"id":"b","type":"custom","revision":"r1"}`},
		{"omit wins", []string{"/properties"}, []string{"/properties/body", "/properties/id"}, `{"id":"b","type":"custom","revision":"r1","properties":{"large":9007199254740993,"nullable":null}}`},
		{"required type property may disappear", nil, []string{"/properties", "/metadata"}, `{"id":"b","type":"custom","revision":"r1"}`},
		{"root include", []string{""}, nil, bead},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			opts := clioutput.Options{JSON: true, Include: tc.include, Omit: append([]string{"envelope"}, tc.omit...)}
			r := renderer(t, descriptor(), opts)
			out, err := r.Render(json.RawMessage(bead))
			if err != nil {
				t.Fatal(err)
			}
			equalJSON(t, out, tc.want)
		})
	}
}

func TestOriginalArrayIndexesAndEscapedPointers(t *testing.T) {
	resource := `{"id":"b","type":"custom","revision":"r1","properties":{"a/b":{"~name":["zero","one","two","three"]},"":null}}`
	cases := []struct {
		name    string
		include []string
		omit    []string
		want    string
	}{
		{"omit original indices", nil, []string{"/properties/a~1b/~0name/0", "/properties/a~1b/~0name/2", "/properties/a~1b/~0name/0"}, `{"id":"b","type":"custom","revision":"r1","properties":{"a/b":{"~name":["one","three"]},"":null}}`},
		{"include source order", []string{"/properties/a~1b/~0name/3", "/properties/a~1b/~0name/1"}, nil, `{"id":"b","type":"custom","revision":"r1","properties":{"a/b":{"~name":["one","three"]}}}`},
		{"include omit same original", []string{"/properties/a~1b/~0name/1", "/properties/a~1b/~0name/3"}, []string{"/properties/a~1b/~0name/1"}, `{"id":"b","type":"custom","revision":"r1","properties":{"a/b":{"~name":["three"]}}}`},
		{"empty key", []string{"/properties/"}, nil, `{"id":"b","type":"custom","revision":"r1","properties":{"":null}}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := renderer(t, descriptor(), clioutput.Options{JSON: true, Include: tc.include, Omit: append([]string{"envelope"}, tc.omit...)})
			out, err := r.Render(json.RawMessage(resource))
			if err != nil {
				t.Fatal(err)
			}
			equalJSON(t, out, tc.want)
		})
	}
}

func TestEmbeddedResourceContextIsExplicitAndRecursive(t *testing.T) {
	resource := `{"id":"b","type":"custom","revision":"r1","properties":{"id":"ordinary"},"ownedLinks":[{"id":"l","type":"relation","revision":"l1","source":{"id":"b","revision":"r1"},"target":{"id":"c","revision":"r2"},"properties":{"body":"text"}}]}`
	r := renderer(t, descriptor(), clioutput.Options{JSON: true, Include: []string{"/ownedLinks/0/properties/body"}, Omit: []string{"envelope"}})
	out, err := r.Render(json.RawMessage(resource))
	if err != nil {
		t.Fatal(err)
	}
	equalJSON(t, out, `{"id":"b","type":"custom","revision":"r1","ownedLinks":[{"id":"l","type":"relation","revision":"l1","source":{"id":"b","revision":"r1"},"target":{"id":"c","revision":"r2"},"properties":{"body":"text"}}]}`)

	for _, path := range []string{"/ownedLinks/0/id", "/ownedLinks/0/source/revision", "/ownedLinks/0/target"} {
		r = renderer(t, descriptor(), clioutput.Options{JSON: true, Omit: []string{path}})
		if output, err := r.Render(json.RawMessage(resource)); err == nil || len(output) != 0 {
			t.Fatalf("mandatory omission %s produced %s, %v", path, output, err)
		}
	}
	r = renderer(t, descriptor(), clioutput.Options{JSON: true, Omit: []string{"envelope", "/ownedLinks", "/properties/id"}})
	out, err = r.Render(json.RawMessage(resource))
	if err != nil {
		t.Fatal(err)
	}
	equalJSON(t, out, `{"id":"b","type":"custom","revision":"r1","properties":{}}`)
}

func TestCollectiveContextSurvivesRawProjection(t *testing.T) {
	d := descriptor()
	d.Protection.Required = append(d.Protection.Required, "subjects")
	d.Expandable = append(d.Expandable, "/subjects")
	resource := `{"id":"evaluation","type":"custom","revision":"r1","subjects":[{"id":"a","revision":"a1"},{"id":"b","revision":"b2"}],"properties":{"claim":"collective"}}`
	r := renderer(t, d, clioutput.Options{JSON: true, Include: []string{"/properties/claim"}, Omit: []string{"envelope"}})
	out, err := r.Render(json.RawMessage(resource))
	if err != nil {
		t.Fatal(err)
	}
	equalJSON(t, out, resource)
	r = renderer(t, d, clioutput.Options{JSON: true, Omit: []string{"/subjects/1"}})
	if out, err = r.Render(json.RawMessage(resource)); err == nil || len(out) != 0 {
		t.Fatalf("collective context removal returned %s, %v", out, err)
	}
	// The same key in an ordinary Resource is not a Memory-specific special case.
	r = renderer(t, descriptor(), clioutput.Options{JSON: true, Omit: []string{"/subjects", "envelope"}})
	if _, err = r.Render(json.RawMessage(resource)); err != nil {
		t.Fatal(err)
	}
}

func TestFailurePreservesProblemAndNeverMasqueradesAsData(t *testing.T) {
	problem := json.RawMessage(`{"code":"fixture_problem","status":409,"retry":false,"extension":{"number":9007199254740993}}`)
	r := renderer(t, descriptor(), clioutput.Options{JSON: true})
	out, err := r.RenderFailure("bdp", problem)
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]json.RawMessage
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatal(err)
	}
	if _, exists := result["data"]; exists || string(result["outcome"]) != `"failure"` {
		t.Fatalf("failure looks like data/success: %s", out)
	}
	var wrapped struct {
		Origin  string          `json:"origin"`
		Problem json.RawMessage `json:"problem"`
	}
	if err := json.Unmarshal(result["error"], &wrapped); err != nil {
		t.Fatal(err)
	}
	if wrapped.Origin != "bdp" {
		t.Fatal("lost BDP origin")
	}
	equalJSON(t, wrapped.Problem, string(problem))
	for _, options := range []clioutput.Options{{}, {JSON: true, Omit: []string{"envelope"}}} {
		r = renderer(t, descriptor(), options)
		if out, err = r.RenderFailure("bdp", problem); err != nil || len(out) != 0 {
			t.Fatalf("raw/human error leaked into stdout: %s, %v", out, err)
		}
	}
}

func TestOptionValidationAndDescriptorSnapshot(t *testing.T) {
	for _, options := range []clioutput.Options{
		{JSON: true, Formats: []string{"detail"}},
		{Formats: []string{"json", "detail"}},
		{Formats: []string{"jsonl"}},
		{Formats: []string{"table"}},
		{Formats: []string{""}},
		{Include: []string{"properties/body"}},
		{Include: []string{"#/properties"}},
		{Omit: []string{"/properties/~"}},
		{Omit: []string{"/properties/~2"}},
		{Omit: []string{"/id"}},
		{Include: []string{"/attribution/actor"}},
		{Include: []string{"/properties/\xff"}},
		{Omit: []string{""}},
		{Omit: []string{"envelope"}},
	} {
		if _, err := clioutput.New(descriptor(), options); err == nil {
			t.Fatalf("accepted invalid options: %+v", options)
		}
	}
	d := descriptor()
	opts := clioutput.Options{JSON: true, Formats: []string{"json", "json"}, Include: []string{"/properties/body"}}
	r := renderer(t, d, opts)
	d.Protection.Required[0] = "invented"
	opts.Include[0] = "/properties/large"
	out, err := r.Render(json.RawMessage(bead))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(out, []byte("complete body")) || bytes.Contains(out, []byte("invented")) || bytes.Contains(out, []byte("9007199254740993")) {
		t.Fatalf("caller mutation changed renderer: %s", out)
	}
	cycle := &clioutput.Protection{Required: []string{"id"}}
	cycle.Members = map[string]*clioutput.Protection{"cycle": cycle}
	if _, err = clioutput.New(clioutput.Descriptor{BDPSpec: "fixture", Protection: cycle}, clioutput.Options{}); err == nil {
		t.Fatal("cyclic descriptor accepted")
	}
}

func TestInvalidPayloadReturnsNoSuccessPrefix(t *testing.T) {
	r := renderer(t, descriptor(), clioutput.Options{JSON: true})
	for _, resource := range []string{"", "null", "[]", "{}", bead + bead, `{"id":`} {
		if out, err := r.Render(json.RawMessage(resource)); err == nil || len(out) != 0 {
			t.Fatalf("invalid input produced %s, %v", out, err)
		}
	}
}

func TestDetailDoesNotTruncateOrInterpretControlCharacters(t *testing.T) {
	body := strings.Repeat("long body ", 2000) + "\x1b[31m\nend"
	source := map[string]any{"id": "b", "type": "custom", "revision": "r1", "properties": map[string]any{"body": body}}
	raw, err := json.Marshal(source)
	if err != nil {
		t.Fatal(err)
	}
	r := renderer(t, descriptor(), clioutput.Options{})
	out, err := r.Render(raw)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Count(out, []byte("long body ")) != 2000 || bytes.Contains(out, []byte{0x1b}) || !bytes.Contains(out, []byte(`\u001b[31m\nend`)) {
		t.Fatal("detail truncated a value or emitted terminal control characters")
	}
}

// These are synthetic foundation fixtures, not a qualified BDP payload corpus.
// The real show/link-show adapters must add public exact-pin corpus blobs.
func TestLinkGoldenProjectionAndNativeProvenanceNames(t *testing.T) {
	input, err := os.ReadFile("testdata/link-input.json")
	if err != nil {
		t.Fatal(err)
	}
	d := descriptor()
	d.Protection.Required = append(d.Protection.Required, "source", "target")
	for _, tc := range []struct {
		golden string
		omit   []string
	}{
		{"testdata/link-projected-raw.golden.json", []string{"envelope"}},
		{"testdata/link-projected-envelope.golden.json", nil},
	} {
		r := renderer(t, d, clioutput.Options{JSON: true, Include: []string{"/properties/body"}, Omit: tc.omit})
		out, err := r.Render(input)
		if err != nil {
			t.Fatal(err)
		}
		want, err := os.ReadFile(tc.golden)
		if err != nil {
			t.Fatal(err)
		}
		equalJSON(t, out, string(want))
	}
	r := renderer(t, d, clioutput.Options{JSON: true, Omit: []string{"envelope"}})
	out, err := r.Render(input)
	if err != nil {
		t.Fatal(err)
	}
	equalJSON(t, out, string(input))
	if bytes.Contains(out, []byte("revisionProvenance")) {
		t.Fatal("renderer renamed wire provenance")
	}
}

func TestOptionalEmbeddedAbsenceNullAndEmptyStayDistinct(t *testing.T) {
	r := renderer(t, descriptor(), clioutput.Options{JSON: true, Omit: []string{"envelope"}})
	for _, source := range []string{
		`{"id":"b","type":"custom","revision":"r1"}`,
		`{"id":"b","type":"custom","revision":"r1","ownedLinks":null}`,
		`{"id":"b","type":"custom","revision":"r1","ownedLinks":[]}`,
	} {
		out, err := r.Render(json.RawMessage(source))
		if err != nil {
			t.Fatal(err)
		}
		equalJSON(t, out, source)
	}
}
