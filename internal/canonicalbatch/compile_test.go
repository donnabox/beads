package canonicalbatch_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/canonicalbatch"
)

const scope = "https://work.example/acme/"
const create = `{"operation":"createBead","name":"a","id":"a","type":"note","properties":{"message":"batch-ref::literal"}}`

func TestJSONLCompositionAndDetachedInput(t *testing.T) {
	input := []byte("\n" + create + "\n\n" + `{"operation":"createLink","type":"references","source":{"binding":"a"},"target":{"binding":"guide"}}` + "\n")
	b, err := canonicalbatch.Compile(input, canonicalbatch.Options{Scope: scope, Args: []string{"guide=b"}})
	if err != nil {
		t.Fatal(err)
	}
	for i := range input {
		input[i] = 'x'
	}
	members := b.Members()
	if len(members) != 2 || members[0].Type.URI != scope+"note" || members[1].Source.Binding != "a" || members[1].Target.URI != scope+"b" {
		t.Fatalf("members %#v", members)
	}
	if !bytes.Contains(members[0].Properties, []byte("batch-ref::literal")) {
		t.Fatal("property literal was rewritten")
	}
	before := b.NativeJSON()
	members[0].Properties[0] = 'x'
	members[0].Name = "corrupted"
	output := b.NativeJSON()
	output[0] = 'x'
	if !bytes.Equal(before, b.NativeJSON()) || b.Members()[0].Name != "a" || b.Members()[0].Properties[0] != '{' {
		t.Fatal("batch aliases caller memory")
	}
	var native struct {
		Version    string            `json:"capabilityVersion"`
		Operations []json.RawMessage `json:"operations"`
	}
	if err = json.Unmarshal(before, &native); err != nil || native.Version != "2" || len(native.Operations) != 2 {
		t.Fatalf("native %s %v", before, err)
	}
}

func TestArgumentAndBindingFailures(t *testing.T) {
	for _, tc := range []struct {
		name, input string
		args        []string
	}{
		{"missing", `{"operation":"deleteBead","bead":{"binding":"external"}}`, nil},
		{"unused", create, []string{"unused=a"}},
		{"duplicate argument", `{"operation":"deleteBead","bead":{"binding":"x"}}`, []string{"x=a", "x=b"}},
		{"override", create, []string{"a=b"}},
		{"forward", `{"operation":"deleteBead","bead":{"binding":"a"}}` + "\n" + create, []string{"a=b"}},
		{"duplicate definition", create + "\n" + create, nil},
		{"wrong kind", `{"operation":"createLink","name":"edge","type":"references","source":"a","target":"b"}` + "\n" + `{"operation":"deleteBead","bead":{"binding":"edge"}}`, nil},
		{"pinned mutation argument", `{"operation":"deleteBead","bead":{"binding":"x"}}`, []string{"x=a@=r"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b, err := canonicalbatch.Compile([]byte(tc.input), canonicalbatch.Options{Scope: scope, Args: tc.args})
			if b != nil || !errors.Is(err, canonicalbatch.ErrInvalid) {
				t.Fatalf("batch=%v err=%v", b, err)
			}
		})
	}
}

func TestEntireInputRefused(t *testing.T) {
	cases := map[string][]byte{
		"prefix then malformed": []byte(create + "\n" + `{"operation":`),
		"array":                 []byte("[" + create + "]"), "scalar": []byte(`"command"`), "comment": []byte("# comment\n" + create), "multiline": []byte("{\n\"operation\":\"deleteBead\",\"bead\":\"a\"}\n"),
		"duplicate":              []byte(`{"operation":"deleteBead","bead":"a","bead":"b"}`),
		"nested duplicate":       []byte(`{"operation":"createBead","type":"note","properties":{"a":1,"a":2}}`),
		"lone surrogate":         []byte(`{"operation":"createBead","type":"note","properties":{"a":"\ud800"}}`),
		"rounded number":         []byte(`{"operation":"createBead","type":"note","properties":{"a":9007199254740993}}`),
		"infinite":               []byte(`{"operation":"createBead","type":"note","properties":{"a":1e999}}`),
		"invalid utf8":           append([]byte(`{"operation":"createBead","type":"note","properties":{"a":"`), append([]byte{0xff}, []byte(`"}}`)...)...),
		"unknown field":          []byte(`{"operation":"deleteBead","bead":"a","force":true}`),
		"unknown operation":      []byte(`{"operation":"sequence"}`),
		"empty":                  []byte("\n \t\n"),
		"replacement with patch": []byte(`{"operation":"updateBead","bead":"a","properties":{},"propertiesChange":[]}`),
		"guard only":             []byte(`{"operation":"updateBead","bead":"a","expectedRevision":"r"}`),
		"bad patch":              []byte(`{"operation":"updateBead","bead":"a","propertiesChange":[{"op":"copy","path":"/x","from":"/y"}]}`),
		"pinned target":          []byte(`{"operation":"deleteBead","bead":{"uri":"https://work.example/acme/a","revision":"r"}}`),
		"binding as type":        []byte(`{"operation":"createBead","type":{"binding":"a"}}`),
		"new at ID":              []byte(`{"operation":"createBead","type":"note","id":"x@y"}`),
		"delete bead provenance": []byte(`{"operation":"deleteBead","bead":"a","revisionProvenance":{}}`),
	}
	for name, input := range cases {
		t.Run(name, func(t *testing.T) {
			b, err := canonicalbatch.Compile(input, canonicalbatch.Options{Scope: scope})
			if b != nil || !errors.Is(err, canonicalbatch.ErrInvalid) {
				t.Fatalf("batch=%v err=%v", b, err)
			}
		})
	}
}

func TestNativeVersionDispatchAndLiteralRules(t *testing.T) {
	for _, raw := range []string{`{"capabilityVersion":"1","future":true}`, `{"capabilityVersion":"future","operations":false}`} {
		if _, err := canonicalbatch.Compile([]byte(raw), canonicalbatch.Options{Format: "json"}); !errors.Is(err, canonicalbatch.ErrUnsupported) {
			t.Fatalf("%s: %v", raw, err)
		}
	}
	cases := []struct {
		raw  string
		opts canonicalbatch.Options
	}{
		{`{"operations":[]}`, canonicalbatch.Options{Format: "json"}},
		{`{"capabilityVersion":"2","operations":[]}`, canonicalbatch.Options{Format: "json"}},
		{`{"capabilityVersion":"2","operations":[{"operation":"deleteBead","bead":"a"}]}`, canonicalbatch.Options{Format: "json", Scope: scope}},
		{`{"capabilityVersion":"2","operations":[{"operation":"deleteBead","bead":{"binding":"external"}}]}`, canonicalbatch.Options{Format: "json"}},
		{`{"capabilityVersion":"2","operations":[{"operation":"deleteBead","bead":"https://work.example/acme/a"}]}`, canonicalbatch.Options{Format: "json", Version: "2"}},
		{`{"capabilityVersion":"2","operations":[{"operation":"deleteBead","bead":"https://work.example/acme/a"}]}`, canonicalbatch.Options{Format: "json", Args: []string{"x=a"}}},
	}
	for _, tc := range cases {
		if b, err := canonicalbatch.Compile([]byte(tc.raw), tc.opts); b != nil || !errors.Is(err, canonicalbatch.ErrInvalid) {
			t.Fatalf("%s: %v %v", tc.raw, b, err)
		}
	}
}

func TestPairedSharedFixtures(t *testing.T) {
	for _, name := range []string{"basics", "bindings", "review", "cleanup"} {
		t.Run(name, func(t *testing.T) {
			input, err := os.ReadFile("testdata/" + name + ".json")
			if err != nil {
				t.Fatal(err)
			}
			b, err := canonicalbatch.Compile(input, canonicalbatch.Options{Format: "json", Scope: scope})
			if err != nil {
				t.Fatal(err)
			}
			want, err := graphops.CanonicalizeJSON(input)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(want, b.NativeJSON()) {
				t.Fatalf("fixture changed:\n%s\n%s", want, b.NativeJSON())
			}
		})
	}
}

func TestSelectionsResolveStructuralArgumentsOnly(t *testing.T) {
	first := `{"operation":"select","name":"picked","query":{"ir":"bdp-query-design-1","profile":"query","selection":{"op":"refs","collection":"beads","references":{"binding":"root"}}}}`
	second := `{"operation":"select","name":"filtered","query":{"ir":"bdp-query-design-1","profile":"list","selection":{"op":"filter","input":{"op":"refs","collection":"beads","references":{"binding":"picked"}},"bind":"b","predicate":{"op":"compare","operator":"eq","left":{"op":"field","of":{"op":"binding","name":"b"},"path":["properties","x"]},"right":{"op":"literal","value":{"binding":"not-an-argument","op":"refs"}}}}}}`
	b, err := canonicalbatch.Compile([]byte(first+"\n"+second+"\n"+`{"operation":"deleteBead","bead":{"binding":"filtered"}}`), canonicalbatch.Options{Scope: scope, Args: []string{"root=a"}})
	if err != nil {
		t.Fatal(err)
	}
	members := b.Members()
	if members[0].Kind != graphops.KindBead || !members[0].Set || !bytes.Contains(members[0].Query, []byte(`"references":["https://work.example/acme/a"]`)) {
		t.Fatalf("first selector %s", members[0].Query)
	}
	if !bytes.Contains(members[1].Query, []byte(`"binding":"not-an-argument"`)) {
		t.Fatal("literal object was interpreted")
	}
	bad := strings.Replace(first, `"references":{"binding":"root"}`, `"references":[{"uri":"https://work.example/acme/a","revision":"r"}]`, 1)
	if _, err = canonicalbatch.Compile([]byte(bad), canonicalbatch.Options{}); !errors.Is(err, canonicalbatch.ErrUnsupported) {
		t.Fatalf("historical selection: %v", err)
	}
}

func TestMemberEmitterAndReferenceSyntax(t *testing.T) {
	ref, err := canonicalbatch.ParseReference("note:v1@=r@7", scope)
	if err != nil || ref.URI != scope+"note:v1" || ref.Revision != "r@7" {
		t.Fatalf("%#v %v", ref, err)
	}
	for _, tc := range []struct{ input, uri, revision string }{
		{"[https://[::1]/note]@=123", "https://[::1]/note", "123"},
		{"https://work.example/acme/a@=literal", "https://work.example/acme/a@=literal", ""},
	} {
		ref, err = canonicalbatch.ParseReference(tc.input, scope)
		if err != nil || ref.URI != tc.uri || ref.Revision != tc.revision {
			t.Fatalf("%s: %#v %v", tc.input, ref, err)
		}
	}
	for _, input := range []string{"a@3", "a@current"} {
		if _, err = canonicalbatch.ParseReference(input, scope); !errors.Is(err, canonicalbatch.ErrUnsupported) {
			t.Fatalf("%s: %v", input, err)
		}
	}
	for _, input := range []string{"a@opaque", "[https://example.test/a]", "batch-ref::a"} {
		if _, err = canonicalbatch.ParseReference(input, scope); !errors.Is(err, canonicalbatch.ErrInvalid) {
			t.Fatalf("%s: %v", input, err)
		}
	}
	operand, err := canonicalbatch.ParseOperand("batch-ref::future", scope)
	if err != nil || operand.Binding != "future" {
		t.Fatalf("%#v %v", operand, err)
	}
	m, emitted, err := canonicalbatch.CompileMember([]byte(`{"operation":"createLink","type":"references","source":{"binding":"a"},"target":{"binding":"b"}}`), scope)
	if err != nil || m.Source.Binding != "a" || emitted[len(emitted)-1] != '\n' {
		t.Fatalf("%#v %s %v", m, emitted, err)
	}
}

func TestPresenceScopeAndProvenance(t *testing.T) {
	b, err := canonicalbatch.Compile([]byte(`{"operation":"updateBead","bead":"a","metadataChange":[],"properties":{},"expectedRevision":"=literal-token","revisionProvenance":{"actor":{"id":"agent:a","basis":"writer-supplied"},"agent":null,"message":""}}`), canonicalbatch.Options{Scope: scope})
	if err != nil {
		t.Fatal(err)
	}
	m := b.Members()[0]
	if string(m.MetadataChange) != "[]" || string(m.Properties) != "{}" || m.Metadata != nil || m.ExpectedRevision != "=literal-token" {
		t.Fatalf("presence lost: %#v", m)
	}
	for _, raw := range []string{
		`{"operation":"deleteBead","bead":"https://else.example/acme/a"}`,
		`{"operation":"createBead","type":"note","id":"https://else.example/acme/a"}`,
	} {
		if _, err = canonicalbatch.Compile([]byte(raw), canonicalbatch.Options{Scope: scope}); !errors.Is(err, canonicalbatch.ErrUnsupported) {
			t.Fatalf("cross scope: %v", err)
		}
	}
	if _, err = canonicalbatch.Compile([]byte(`{"operation":"createLink","type":"references","source":"a","target":"https://else.example/acme/a"}`), canonicalbatch.Options{Scope: scope}); err != nil {
		t.Fatalf("opaque external endpoint refused: %v", err)
	}
}
