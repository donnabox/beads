package typecatalog

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
)

func definitionCaptureExamples(t *testing.T) []json.RawMessage {
	t.Helper()
	raw, err := os.ReadFile("testdata/definition-records.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Examples []struct {
			Name     string          `json:"name"`
			Response json.RawMessage `json:"response"`
		} `json:"examples"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	want := []string{"task", "type", "bead-type", "link-type", "refines"}
	if len(fixture.Examples) != len(want) {
		t.Fatal("incomplete reviewed example set")
	}
	result := make([]json.RawMessage, len(want))
	for i, example := range fixture.Examples {
		if example.Name != want[i] {
			t.Fatalf("example %d: got %s", i, example.Name)
		}
		result[i] = example.Response
	}
	return result
}

func TestDefinitionCaptureReviewedRecords(t *testing.T) {
	for _, raw := range definitionCaptureExamples(t) {
		capture, err := CaptureDefinition(raw)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(capture.Bytes(), raw) {
			t.Fatal("original wrapper bytes changed")
		}
		var wrapper map[string]json.RawMessage
		if err := json.Unmarshal(raw, &wrapper); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(capture.DefinitionBytes(), wrapper["definition"]) {
			t.Fatal("original inner record bytes changed")
		}
		var record map[string]json.RawMessage
		if err := json.Unmarshal(wrapper["definition"], &record); err != nil {
			t.Fatal(err)
		}
		var id, revision string
		if err := json.Unmarshal(record["id"], &id); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(record["revision"], &revision); err != nil {
			t.Fatal(err)
		}
		if capture.Reference() != (Reference{URI: id, Revision: revision}) {
			t.Fatal("exact definition pair changed")
		}
		if _, present := record["metadata"]; present {
			t.Fatal("fixture unexpectedly supplies metadata")
		}
		if _, present := record["revisionProvenance"]; present {
			t.Fatal("fixture unexpectedly supplies provenance")
		}
		if origin, present := capture.OriginScope(); !present || origin == "" {
			t.Fatal("explicit fixture context was lost")
		}
	}
}

const definitionCaptureLiteral = ` { "definition": {
 "id":"https://example.test/type", "revision":"r+1",
 "type":{"uri":"https://example.test/bead-type","revision":"meta-9"},
 "properties":{"propertiesSchema":true,"schemaResources":{"https://example.test/s":false},
   "propertyAccess":{"/a/b":"init-only"},"unknown":[9007199254740993,1e400,1e-400,-0.00,"\u0061"]},
 "ownedLinks":{"https://example.test/refines":[{
   "id":"https://example.test/decl", "revision":"link-7",
   "type":{"uri":"https://example.test/refines","revision":"relationship-3"},
   "source":"https://example.test/type",
   "target":{"uri":"https://example.test/type","revision":"r-old"},"properties":{}
 }]},"extension":{"nested":[null,true,{"arbitrary":"kept"}]}
} } `

func TestDefinitionCaptureLosslessAndIndependent(t *testing.T) {
	input := []byte(definitionCaptureLiteral)
	capture, err := CaptureDefinition(input)
	if err != nil {
		t.Fatal(err)
	}
	original := bytes.Clone(input)
	wantDefinition := capture.DefinitionBytes()
	for i := range input {
		input[i] = 'x'
	}
	for _, output := range [][]byte{capture.Bytes(), capture.DefinitionBytes()} {
		for i := range output {
			output[i] = 'x'
		}
	}
	if !bytes.Equal(capture.Bytes(), original) || !bytes.Equal(capture.DefinitionBytes(), wantDefinition) {
		t.Fatal("mutable input/output changed retained lexemes")
	}
	if _, present := capture.OriginScope(); present {
		t.Fatal("unknown origin fabricated")
	}
	if capture.Reference() != (Reference{URI: "https://example.test/type", Revision: "r+1"}) {
		t.Fatal("identity/revision changed")
	}
	// A separate newer snapshot is just another value, never a lookup/default.
	newer := bytes.ReplaceAll(original, []byte(`"r+1"`), []byte(`"r-new"`))
	newer = bytes.ReplaceAll(newer, []byte(`"link-7"`), []byte(`"link-new"`))
	newCapture, err := CaptureDefinition(newer)
	if err != nil {
		t.Fatal(err)
	}
	if newCapture.Reference() == capture.Reference() || !bytes.Equal(capture.Bytes(), original) {
		t.Fatal("new capture moved the old exact state")
	}
	// Present optional values remain present, including null inside opaque data.
	present := bytes.Replace(original, []byte(`"extension":`), []byte(`"metadata":{},"revisionProvenance":{"opaque":null},"extension":`), 1)
	withFields, err := CaptureDefinition(present)
	if err != nil {
		t.Fatalf("opaque supplied fields must be retained, not interpreted: %v", err)
	}
	if !bytes.Equal(withFields.Bytes(), present) || !bytes.Contains(withFields.DefinitionBytes(), []byte(`"metadata":{},"revisionProvenance":{"opaque":null}`)) {
		t.Fatal("supplied optional fields lost")
	}
}

func TestDefinitionCaptureDoesNotApplyLegacyURLAdmission(t *testing.T) {
	// These structural carrier spellings are not assertions of valid canonical
	// publication identities. The capture must neither reject nor normalize them
	// by applying the older fixed-catalog admission policy.
	for _, uri := range []string{"https://EXAMPLE.test/a/../b#fragment", "https://user:pass@example.test/type", "https://%65xample.org/"} {
		raw := bytes.Replace([]byte(definitionCaptureLiteral), []byte("https://example.test/type"), []byte(uri), 1)
		capture, err := CaptureDefinition(raw)
		if err != nil {
			t.Fatal(err)
		}
		if capture.Reference().URI != uri || !bytes.Equal(capture.Bytes(), raw) {
			t.Fatal("capture applied identity normalization")
		}
	}
}

func TestDefinitionCaptureRejectsAmbiguousOrMissingCarriers(t *testing.T) {
	cases := []struct {
		name string
		raw  []byte
	}{
		{"trailing", []byte(definitionCaptureLiteral + `{}`)},
		{"not-object", []byte(`[]`)},
		{"missing-definition", []byte(`{}`)},
		{"null-definition", []byte(`{"definition":null}`)},
		{"extra-wrapper", bytes.Replace([]byte(definitionCaptureLiteral), []byte(`{ "definition"`), []byte(`{ "registered":true,"definition"`), 1)},
		{"malformed-authority", bytes.Replace([]byte(definitionCaptureLiteral), []byte("https://example.test/type"), []byte("https://["), 1)},
		{"malformed-endpoint-authority", bytes.Replace([]byte(definitionCaptureLiteral), []byte(`"target":{"uri":"https://example.test/type"`), []byte(`"target":{"uri":"https://["`), 1)},
		{"bad-owned-key", bytes.Replace([]byte(definitionCaptureLiteral), []byte(`"ownedLinks":{"https://example.test/refines"`), []byte(`"ownedLinks":{"*"`), 1)},
		{"links-plane", bytes.Replace([]byte(definitionCaptureLiteral), []byte(`"ownedLinks":`), []byte(`"links":{},"ownedLinks":`), 1)},
		{"floating-owned-affiliation", bytes.Replace([]byte(definitionCaptureLiteral), []byte(`{"uri":"https://example.test/refines","revision":"relationship-3"}`), []byte(`"https://example.test/refines"`), 1)},
		{"missing-source", bytes.Replace([]byte(definitionCaptureLiteral), []byte(`"source":`), []byte(`"otherSource":`), 1)},
		{"numeric-target", bytes.Replace([]byte(definitionCaptureLiteral), []byte(`{"uri":"https://example.test/type","revision":"r-old"}`), []byte(`42`), 1)},
		{"missing-link-properties", bytes.Replace([]byte(definitionCaptureLiteral), []byte(`"r-old"},"properties":{}`), []byte(`"r-old"},"otherProperties":{}`), 1)},
		{"empty-revision", bytes.Replace([]byte(definitionCaptureLiteral), []byte(`"r+1"`), []byte(`""`), 1)},
		{"floating-affiliation", bytes.Replace([]byte(definitionCaptureLiteral), []byte(`{"uri":"https://example.test/bead-type","revision":"meta-9"}`), []byte(`"https://example.test/bead-type"`), 1)},
		{"pin-extra-field", bytes.Replace([]byte(definitionCaptureLiteral), []byte(`"meta-9"`), []byte(`"meta-9","extra":true`), 1)},
		{"missing-owned", bytes.Replace([]byte(definitionCaptureLiteral), []byte(`"ownedLinks":`), []byte(`"otherOwned":`), 1)},
		{"null-owned-group", bytes.Replace([]byte(definitionCaptureLiteral), []byte(`[{`), []byte(`null,"other":[{`), 1)},
		{"null-origin", bytes.Replace([]byte(definitionCaptureLiteral), []byte(`{ "definition"`), []byte(`{ "originScope":null,"definition"`), 1)},
		{"relative-origin", bytes.Replace([]byte(definitionCaptureLiteral), []byte(`{ "definition"`), []byte(`{ "originScope":"local/","definition"`), 1)},
		{"duplicate-id", bytes.Replace([]byte(definitionCaptureLiteral), []byte(`"id":"https://example.test/type"`), []byte(`"id":"https://example.test/type","\u0069d":"https://other.test/type"`), 1)},
		{"nested-duplicate", bytes.Replace([]byte(definitionCaptureLiteral), []byte(`"arbitrary":"kept"`), []byte(`"arbitrary":"kept","arbitrary":"other"`), 1)},
		{"surrogate-key", bytes.Replace([]byte(definitionCaptureLiteral), []byte(`"arbitrary"`), []byte(`"\ud800"`), 1)},
		{"surrogate-value", bytes.Replace([]byte(definitionCaptureLiteral), []byte(`"kept"`), []byte(`"\udc00"`), 1)},
		{"invalid-utf8", bytes.Replace([]byte(definitionCaptureLiteral), []byte("kept"), []byte{0xff}, 1)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := CaptureDefinition(tc.raw); err == nil {
				t.Fatal("invalid or ambiguous carrier accepted")
			}
		})
	}
}
