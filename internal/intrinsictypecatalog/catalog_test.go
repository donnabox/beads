package intrinsictypecatalog

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	graph "github.com/steveyegge/beads/graphops"
)

func mustSelected(t *testing.T) *Catalog {
	t.Helper()
	c, err := LoadSelected()
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestSelectedPublication(t *testing.T) {
	c := mustSelected(t)
	if s := SelectedSource(); s.Commit != "ba8d18c6355cbd6dd3499befe03641874bb01c73" || s.SHA256 != digest(c.Bytes()) || s.Bytes != len(c.Bytes()) {
		t.Fatal("source binding")
	}
	if !bytes.Equal(c.Bytes(), selectedBytes) || len(c.Definitions()) != 4 || len(c.Declarations()) != 2 {
		t.Fatal("complete original publication")
	}
	for _, name := range []string{"type", "bead-type", "link-type", "refines"} {
		ref := pin(name)
		raw, ok := c.Definition(ref)
		if !ok {
			t.Fatal(name)
		}
		var d definition
		if err := decode(raw, &d); err != nil {
			t.Fatal(err)
		}
		want := pin("bead-type")
		kind := graph.KindBead
		if name == "refines" {
			want = pin("link-type")
			kind = graph.KindLink
		}
		if d.Type != want {
			t.Fatalf("%s affiliation", name)
		}
		if got, ok := c.Category(ref); !ok || got != kind {
			t.Fatalf("%s category", name)
		}
		parents, ok := c.Ancestors(ref)
		if !ok {
			t.Fatal("missing closure")
		}
		if name == "bead-type" || name == "link-type" {
			if !reflect.DeepEqual(parents, []Reference{pin("type")}) {
				t.Fatal("recorded parents")
			}
		} else if len(parents) != 0 {
			t.Fatal("invented parent")
		}
	}
	if _, ok := c.Definition(Reference{pin("type").URI, "newer"}); ok {
		t.Fatal("revision fallback")
	}
	if _, ok := c.Definition(Reference{pin("type").URI, ""}); ok {
		t.Fatal("default fallback")
	}
	if _, ok := c.Category(Reference{}); ok {
		t.Fatal("unknown category")
	}
	if _, ok := c.Ancestors(Reference{}); ok {
		t.Fatal("unknown closure")
	}
	// Full property/schema/owned-state records must survive; never a legacy
	// projected descriptor that discards its own affiliation or declarations.
	if !bytes.Contains(c.Definitions()[1], []byte(`"schemaResources"`)) || !bytes.Contains(c.Definitions()[1], []byte(`"ownedLinks"`)) {
		t.Fatal("lossy record projection")
	}
}

func TestIndependentOutputs(t *testing.T) {
	c := mustSelected(t)
	original := c.Bytes()
	b := c.Bytes()
	b[0] = '!'
	defs := c.Definitions()
	defs[0][0] = '!'
	defs[1] = nil
	decls := c.Declarations()
	decls[0][0] = '!'
	raw, _ := c.Definition(pin("type"))
	raw[0] = '!'
	ancestors, _ := c.Ancestors(pin("bead-type"))
	ancestors[0] = Reference{}
	if !bytes.Equal(c.Bytes(), original) || c.Definitions()[0][0] == '!' || c.Declarations()[0][0] == '!' {
		t.Fatal("accessor alias")
	}
	fresh := mustSelected(t)
	if !bytes.Equal(fresh.Bytes(), original) {
		t.Fatal("cross-load alias")
	}
	parents, _ := c.Ancestors(pin("bead-type"))
	if parents[0] != pin("type") {
		t.Fatal("ancestor alias")
	}
	// The private validator also takes ownership rather than retaining input.
	input := bytes.Clone(selectedBytes)
	checked, err := validatePublication(input)
	if err != nil {
		t.Fatal(err)
	}
	input[0] = '!'
	if !bytes.Equal(checked.Bytes(), original) {
		t.Fatal("input alias")
	}
}

func mutation(t *testing.T, change func(map[string]any)) []byte {
	t.Helper()
	var x map[string]any
	if err := json.Unmarshal(selectedBytes, &x); err != nil {
		t.Fatal(err)
	}
	change(x)
	raw, err := json.Marshal(x)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
func defAt(x map[string]any, i int) map[string]any {
	return x["definitions"].([]any)[i].(map[string]any)
}
func props(x map[string]any, i int) map[string]any { return defAt(x, i)["properties"].(map[string]any) }
func captured(x map[string]any, i int) map[string]any {
	return defAt(x, i)["ownedLinks"].(map[string]any)[pin("refines").URI].([]any)[0].(map[string]any)
}

func TestPublicationRefusals(t *testing.T) {
	tests := []struct {
		name, want string
		change     func(map[string]any)
	}{
		{"extra-root", "unknown field", func(x map[string]any) { x["originScope"] = "https://invented.invalid/" }},
		{"missing-definition", "four definitions", func(x map[string]any) { x["definitions"] = x["definitions"].([]any)[:3] }},
		{"extra-definition", "four definitions", func(x map[string]any) { x["definitions"] = append(x["definitions"].([]any), defAt(x, 0)) }},
		{"duplicate-definition", "duplicate definition", func(x map[string]any) { x["definitions"].([]any)[1] = defAt(x, 0) }},
		{"unknown-edition", "missing exact intrinsic", func(x map[string]any) { defAt(x, 0)["revision"] = "type-2" }},
		{"unknown-literal-id", "missing exact intrinsic", func(x map[string]any) { defAt(x, 0)["id"] = namespace + "Type" }},
		{"wrong-category", "incorrect intrinsic affiliation", func(x map[string]any) { defAt(x, 3)["type"] = pin("bead-type") }},
		{"unversioned-affiliation", "missing exact reference", func(x map[string]any) { defAt(x, 0)["type"] = map[string]any{"uri": pin("bead-type").URI} }},
		{"literal-field-case", "missing literal carrier field", func(x map[string]any) { d := defAt(x, 0); d["ID"] = d["id"]; delete(d, "id") }},
		{"literal-reference-case", "missing exact reference", func(x map[string]any) {
			defAt(x, 0)["type"] = map[string]any{"URI": pin("bead-type").URI, "revision": "bead-type-1"}
		}},
		{"captured-cycle", "exact Refines cycle", func(x map[string]any) {
			captured(x, 1)["target"] = pin("link-type")
			captured(x, 2)["target"] = pin("bead-type")
			x["declarations"].([]any)[0] = captured(x, 1)
			x["declarations"].([]any)[1] = captured(x, 2)
		}},
		{"missing-abstract", "incorrect abstract", func(x map[string]any) { delete(props(x, 0), "abstract") }},
		{"abstract-concrete", "incorrect abstract", func(x map[string]any) { props(x, 1)["abstract"] = true }},
		{"missing-ownership", "Type must own Refines", func(x map[string]any) { delete(props(x, 0), "ownsOutgoing") }},
		{"cross-category-pairs", "category-preserving", func(x map[string]any) { props(x, 3)["linkConstraints"] = map[string]any{} }},
		{"missing-declaration", "two declarations", func(x map[string]any) { x["declarations"] = x["declarations"].([]any)[:1] }},
		{"duplicate-declaration", "duplicate declaration", func(x map[string]any) { x["declarations"].([]any)[1] = x["declarations"].([]any)[0] }},
		{"capture-copy", "captured declaration mismatch", func(x map[string]any) { captured(x, 1)["revision"] = "changed" }},
		{"capture-source", "captured source differs", func(x map[string]any) {
			captured(x, 1)["source"] = pin("type").URI
			x["declarations"].([]any)[0] = captured(x, 1)
		}},
		{"empty-capture", "empty captured group", func(x map[string]any) { defAt(x, 1)["ownedLinks"].(map[string]any)[pin("refines").URI] = []any{} }},
		{"wrong-nominal", "invalid nominal", func(x map[string]any) {
			owned := defAt(x, 1)["ownedLinks"].(map[string]any)
			owned["wrong"] = owned[pin("refines").URI]
			delete(owned, pin("refines").URI)
		}},
		{"missing-exact-target", "missing exact Refines target", func(x map[string]any) {
			captured(x, 1)["target"] = Reference{pin("type").URI, "type-2"}
			x["declarations"].([]any)[0] = captured(x, 1)
		}},
		{"wrong-kind-edge", "cross-category Refines", func(x map[string]any) {
			captured(x, 1)["target"] = pin("refines")
			x["declarations"].([]any)[0] = captured(x, 1)
		}},
		{"missing-standard", "missing retained standard", func(x map[string]any) { delete(props(x, 0)["schemaResources"].(map[string]any), dialect) }},
		{"missing-parent-schema", "missing retained Type schema", func(x map[string]any) {
			delete(props(x, 1)["schemaResources"].(map[string]any), schemaNamespace+"type-definition.properties.schema.json")
		}},
		{"altered-standard", "retained schema source drift", func(x map[string]any) {
			props(x, 0)["schemaResources"].(map[string]any)[dialect].(map[string]any)["title"] = "forged"
		}},
		{"altered-root", "retained schema source drift", func(x map[string]any) { props(x, 0)["propertiesSchema"].(map[string]any)["title"] = "forged" }},
		{"conflicting-copy", "conflicting retained intrinsic", func(x map[string]any) {
			props(x, 1)["schemaResources"].(map[string]any)[schemaNamespace+"type-definition.properties.schema.json"].(map[string]any)["title"] = "forged"
		}},
		{"invented-origin", "unknown field", func(x map[string]any) { defAt(x, 0)["originScope"] = "https://invented.invalid/" }},
		{"invalid-properties", "metatype properties", func(x map[string]any) { props(x, 1)["propertyAccess"] = map[string]any{"not-a-pointer": "writable"} }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := validatePublication(mutation(t, tt.change))
			if c != nil || !errors.Is(err, graph.ErrValidation) || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("got catalog=%v err=%v; need %s", c, err, tt.want)
			}
		})
	}
	for _, raw := range [][]byte{[]byte(`{"definitions":[],"definitions":[],"declarations":[]}`), append(bytes.Clone(selectedBytes), []byte(`{}`)...), {0xff}} {
		if c, err := validatePublication(raw); c != nil || err == nil {
			t.Fatal("ambiguous/malformed JSON accepted")
		}
	}
}

func TestExactRefinesDAG(t *testing.T) {
	c := mustSelected(t)
	// A metatype self-affiliation is accepted above; an actual Refines cycle
	// (including a self-edge) must be refused by the distinct graph traversal.
	for _, edges := range []map[Reference][]Reference{
		{pin("type"): {pin("type")}},
		{pin("type"): {pin("bead-type")}, pin("bead-type"): {pin("type")}},
		{pin("type"): {Reference{pin("type").URI, "unavailable"}}},
	} {
		if result, err := refinesClosure(c.definitions, edges); result != nil || err == nil {
			t.Fatal("bad exact graph accepted")
		}
	}
}

func TestSchemaClosureIsOffline(t *testing.T) {
	t.Run("selected-standard-builtins", func(t *testing.T) {
		c := mustSelected(t)
		set, err := schemas(c.definitions[pin("type")])
		if err != nil {
			t.Fatal(err)
		}
		// This physically retained environment compiles all eight standard IDs,
		// while selecting the standard dialect itself as the returned validator.
		meta, err := compileSchemas(schemaSet{dialect, set.resources})
		if err != nil {
			t.Fatal(err)
		}
		if err := meta.Validate(map[string]any{"type": "object"}); err != nil {
			t.Fatal(err)
		}
		if err := meta.Validate(map[string]any{"type": false}); err == nil {
			t.Fatal("built-in meta-schema did not validate")
		}
	})
	t.Run("forged-selected-standards", func(t *testing.T) {
		c := mustSelected(t)
		for uri := range selectedSchemaDigests() {
			if _, standard := selectedStandardDigest(uri); !standard {
				continue
			}
			t.Run(strings.TrimPrefix(uri, "https://json-schema.org/draft/2020-12/"), func(t *testing.T) {
				set, err := schemas(c.definitions[pin("type")])
				if err != nil {
					t.Fatal(err)
				}
				var forged map[string]any
				if err := json.Unmarshal(set.resources[uri], &forged); err != nil {
					t.Fatal(err)
				}
				forged["x-forged"] = true
				raw, err := json.Marshal(forged)
				if err != nil {
					t.Fatal(err)
				}
				set.resources[uri] = raw
				compiled, err := compileSchemas(set)
				if compiled != nil || !errors.Is(err, graph.ErrValidation) || !strings.Contains(err.Error(), "retained standard schema source drift") {
					t.Fatalf("forged standard not classified/refused: %v", err)
				}
			})
		}
	})
	for _, uri := range []string{dialect + "#suffix", "http://json-schema.org/draft/2020-12/schema", "https://json-schema.org/draft/2019-09/schema", "https://json-schema.org/draft/2020-12/meta/not-selected"} {
		if _, standard := selectedStandardDigest(uri); standard {
			t.Fatalf("broad standard identity match: %s", uri)
		}
	}
	// Actual compiler, not a parallel semantic model; no network resolver is
	// configured. Use a private schema helper to exercise closure independently
	// of fixed source hashes, which already refuse any altered public document.
	cases := []struct{ name, ref string }{
		{"unavailable", "https://unavailable.invalid/schema"},
		{"missing-anchor", "https://schemas.invalid/root#not-there"},
		{"missing-pointer", "https://schemas.invalid/root#/$defs/missing"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			raw, err := json.Marshal(map[string]any{"$schema": dialect, "$id": "https://schemas.invalid/root", "$ref": tt.ref})
			if err != nil {
				t.Fatal(err)
			}
			_, err = compileSchemas(schemaSet{"https://schemas.invalid/root", map[string]json.RawMessage{"https://schemas.invalid/root": raw}})
			if !errors.Is(err, graph.ErrValidation) {
				t.Fatalf("expected actual offline closure refusal: %v", err)
			}
		})
	}
	if _, err := (offlineLoader{}).Load(dialect); !errors.Is(err, graph.ErrValidation) {
		t.Fatal("loader fallback")
	}
}

func TestSelectedSourceBytesRefuseDrift(t *testing.T) {
	for _, raw := range [][]byte{append(bytes.Clone(selectedBytes), ' '), bytes.Replace(selectedBytes, []byte("type-1"), []byte("type-2"), 1)} {
		c, err := loadSelected(raw)
		if c != nil || !errors.Is(err, graph.ErrValidation) || !strings.Contains(err.Error(), "bytes differ") {
			t.Fatalf("source drift accepted: %v", err)
		}
	}
}

func TestTwoFixedPublicationEditions(t *testing.T) {
	for _, source := range []Source{SelectedSource(), CurrentSource()} {
		t.Run(source.Path, func(t *testing.T) {
			c, err := LoadSource(source)
			if err != nil {
				t.Fatal(err)
			}
			if c.Source() != source || len(c.Bytes()) != source.Bytes || digest(c.Bytes()) != source.SHA256 {
				t.Fatal("source binding")
			}
			if len(c.Definitions()) != 4 || len(c.Declarations()) != 2 {
				t.Fatal("fixed census")
			}
			for _, name := range []string{"type", "bead-type", "link-type", "refines"} {
				ref, ok := c.Pin(name)
				if !ok {
					t.Fatal(name)
				}
				if _, ok := c.Definition(ref); !ok {
					t.Fatal("missing exact definition")
				}
				parents, ok := c.Ancestors(ref)
				if !ok {
					t.Fatal("missing closure")
				}
				if name == "bead-type" || name == "link-type" {
					typ, _ := c.Pin("type")
					if !reflect.DeepEqual(parents, []Reference{typ}) {
						t.Fatal("wrong exact parent")
					}
				}
				wrong := ref
				wrong.Revision = "unknown"
				if _, ok := c.Definition(wrong); ok {
					t.Fatal("revision fallback")
				}
			}
			ref, _ := c.Pin("bead-type")
			set, err := schemas(c.definitions[ref])
			if err != nil {
				t.Fatal(err)
			}
			compiled, err := compileSchemas(set)
			if err != nil {
				t.Fatal(err)
			}
			if err = compiled.Validate(map[string]any{}); (err == nil) != (source == CurrentSource()) {
				t.Fatalf("omission edition meaning: %v", err)
			}
			if err = compiled.Validate(map[string]any{"propertiesSchema": nil}); err == nil {
				t.Fatal("null became omission")
			}
			if err = compiled.Validate(map[string]any{"propertiesSchema": false}); err != nil {
				t.Fatalf("explicit boolean schema: %v", err)
			}
			b := c.Bytes()
			b[0] = '!'
			if digest(c.Bytes()) != source.SHA256 {
				t.Fatal("mutable source")
			}
		})
	}
}

func TestFixedPublicationSourceMixRefuses(t *testing.T) {
	mixed := CurrentSource()
	mixed.Commit = SelectedSource().Commit
	for _, source := range []Source{{}, mixed} {
		if c, err := LoadSource(source); c != nil || !errors.Is(err, graph.ErrValidation) {
			t.Fatalf("unknown/mixed source: %v", err)
		}
	}
	if c, err := loadSourceBytes(selectedBytes, CurrentSource()); c != nil || !errors.Is(err, graph.ErrValidation) {
		t.Fatalf("cross-edition bytes: %v", err)
	}
	bad := bytes.Replace(currentBytes, []byte(`"bead-type-2"`), []byte(`"bead-type-1"`), 1)
	if c, err := validatePublicationFor(bad, CurrentSource(), "2"); c != nil || !errors.Is(err, graph.ErrValidation) {
		t.Fatalf("mixed metatype pin: %v", err)
	}
	if c, err := loadSourceBytes(append(bytes.Clone(currentBytes), ' '), CurrentSource()); c != nil || !errors.Is(err, graph.ErrValidation) {
		t.Fatalf("source drift: %v", err)
	}
}

// Legacy finite mutation helpers are test-only, never alternative trust APIs.
func pin(name string) Reference { return Reference{namespace + name, name + "-1"} }
func validatePublication(raw []byte) (*Catalog, error) {
	return validatePublicationFor(raw, SelectedSource(), "1")
}
