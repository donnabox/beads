package intrinsictypecatalog

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/santhosh-tekuri/jsonschema/v6"
	graph "github.com/steveyegge/beads/graphops"
)

// Reference is a literal exact identity/revision pair. No URI rewriting or
// latest/default lookup occurs in this package.
type Reference struct {
	URI      string `json:"uri"`
	Revision string `json:"revision"`
}

type definition struct {
	ID         string                       `json:"id"`
	Revision   string                       `json:"revision"`
	Type       Reference                    `json:"type"`
	Properties map[string]json.RawMessage   `json:"properties"`
	Owned      map[string][]json.RawMessage `json:"ownedLinks"`
}

type declaration struct {
	ID         string                     `json:"id"`
	Revision   string                     `json:"revision"`
	Type       Reference                  `json:"type"`
	Source     string                     `json:"source"`
	Target     Reference                  `json:"target"`
	Properties map[string]json.RawMessage `json:"properties"`
}

type publication struct {
	Definitions  []json.RawMessage `json:"definitions"`
	Declarations []json.RawMessage `json:"declarations"`
}

// Catalog is an immutable checked fixed publication, not an installed catalog.
// All accessors return independent copies, including complete original records.
type Catalog struct {
	source      Source
	edition     string
	raw         []byte
	publication publication
	definitions map[Reference]definition
	records     map[Reference]json.RawMessage
	ancestors   map[Reference][]Reference
}

// LoadSelected validates only the embedded, source-selected publication. There
// is deliberately no caller-selected digest, URI, arbitrary input or trust flag.
func LoadSelected() (*Catalog, error) { return loadSelected(selectedBytes) }

func loadSelected(raw []byte) (*Catalog, error) { return loadSourceBytes(raw, SelectedSource()) }

// LoadCurrent loads only the embedded successor for new fresh contexts.
func LoadCurrent() (*Catalog, error) { return LoadSource(CurrentSource()) }

// LoadSource accepts exactly one complete trusted source tuple. A caller's
// digest cannot authenticate arbitrary bytes or select an unknown publication.
func LoadSource(source Source) (*Catalog, error) {
	switch source {
	case SelectedSource():
		return loadSourceBytes(selectedBytes, source)
	case CurrentSource():
		return loadSourceBytes(currentBytes, source)
	default:
		return nil, invalid("unknown fixed publication source")
	}
}

func loadSourceBytes(raw []byte, source Source) (*Catalog, error) {
	var edition string
	switch source {
	case SelectedSource():
		edition = "1"
	case CurrentSource():
		edition = "2"
	default:
		return nil, invalid("unknown fixed publication source")
	}
	if len(raw) != source.Bytes || digest(raw) != source.SHA256 {
		return nil, invalid("selected publication bytes differ from qualified source")
	}
	return validatePublicationFor(raw, source, edition)
}

// Source returns the exact immutable artifact behind this checked catalog.
func (c *Catalog) Source() Source {
	if c == nil {
		return Source{}
	}
	return c.source
}

// Pin selects a named intrinsic only inside this checked fixed publication.
func (c *Catalog) Pin(name string) (Reference, bool) {
	if c == nil {
		return Reference{}, false
	}
	ref := c.pin(name)
	_, ok := c.definitions[ref]
	return ref, ok
}

func (c *Catalog) pin(name string) Reference {
	return Reference{namespace + name, name + "-" + c.edition}
}
func (c *Catalog) schemaNamespace() string {
	return "https://github.com/gastownhall/bdp/schemas/types/intrinsic-" + c.edition + "/"
}

// Bytes returns the complete unchanged publication serialization.
func (c *Catalog) Bytes() []byte { return bytes.Clone(c.raw) }

// Definitions returns complete definition records in publication order.
func (c *Catalog) Definitions() []json.RawMessage { return cloneRecords(c.publication.Definitions) }

// Declarations returns complete retained declaration records in publication order.
func (c *Catalog) Declarations() []json.RawMessage { return cloneRecords(c.publication.Declarations) }

// Definition returns only a literal exact match, never another revision.
func (c *Catalog) Definition(ref Reference) (json.RawMessage, bool) {
	raw, ok := c.records[ref]
	return bytes.Clone(raw), ok
}

// Category describes instances of the selected definition. Every definition
// itself is a Bead; its own affiliation is retained in its complete record.
func (c *Catalog) Category(ref Reference) (graph.ResourceKind, bool) {
	d, ok := c.definitions[ref]
	if !ok {
		return "", false
	}
	if d.Type == c.pin("link-type") {
		return graph.KindLink, true
	}
	return graph.KindBead, true
}

// Ancestors returns exact recorded Refines ancestors, excluding self. Metatype
// affiliations (including BeadType's self-affiliation) are not Refines edges.
func (c *Catalog) Ancestors(ref Reference) ([]Reference, bool) {
	a, ok := c.ancestors[ref]
	return append([]Reference(nil), a...), ok
}

func cloneRecords(records []json.RawMessage) []json.RawMessage {
	result := make([]json.RawMessage, len(records))
	for i, raw := range records {
		result[i] = bytes.Clone(raw)
	}
	return result
}

func identity(d definition) Reference { return Reference{d.ID, d.Revision} }
func digest(b []byte) string          { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func invalid(format string, args ...any) error {
	return fmt.Errorf("%w: fixed intrinsic publication: %s", graph.ErrValidation, fmt.Sprintf(format, args...))
}

// Private semantic entry point supports the two checked loaders and finite mutations.
func validatePublicationFor(raw []byte, source Source, edition string) (*Catalog, error) {
	if len(raw) > 262144 || !utf8.Valid(raw) {
		return nil, invalid("publication size/UTF-8")
	}
	// Canonicalization refuses duplicate names, invalid string encodings and
	// ambiguous JSON before encoding/json can discard information. Raw is retained.
	if _, err := graph.CanonicalizeJSON(raw); err != nil {
		return nil, invalid("JSON: %v", err)
	}
	var p publication
	if err := decode(raw, &p); err != nil {
		return nil, err
	}
	if len(p.Definitions) != 4 || len(p.Declarations) != 2 {
		return nil, invalid("expected four definitions and two declarations")
	}
	c := &Catalog{source: source, edition: edition, raw: bytes.Clone(raw), publication: p, definitions: map[Reference]definition{}, records: map[Reference]json.RawMessage{}}
	for _, record := range p.Definitions {
		var d definition
		if err := decode(record, &d); err != nil {
			return nil, err
		}
		ref := identity(d)
		if _, exists := c.definitions[ref]; exists {
			return nil, invalid("duplicate definition")
		}
		if d.Properties == nil || d.Owned == nil {
			return nil, invalid("complete properties/ownedLinks required")
		}
		c.definitions[ref] = d
		c.records[ref] = bytes.Clone(record)
	}
	for _, name := range []string{"type", "bead-type", "link-type", "refines"} {
		ref := c.pin(name)
		d, ok := c.definitions[ref]
		if !ok {
			return nil, invalid("missing exact intrinsic %s", name)
		}
		affiliation := c.pin("bead-type")
		if name == "refines" {
			affiliation = c.pin("link-type")
		}
		if d.Type != affiliation {
			return nil, invalid("incorrect intrinsic affiliation %s", name)
		}
		var abstract bool
		if value, ok := d.Properties["abstract"]; ok {
			if err := decode(value, &abstract); err != nil {
				return nil, err
			}
		}
		if abstract != (name == "type") {
			return nil, invalid("incorrect abstract fact %s", name)
		}
	}
	if !equalJSON(c.definitions[c.pin("type")].Properties["ownsOutgoing"], []byte(`[{"anyOf":[{"uri":"`+namespace+`refines","revision":"refines-`+c.edition+`"}]}]`)) {
		return nil, invalid("Type must own Refines")
	}
	pairs := []any{}
	for _, name := range []string{"bead-type", "link-type"} {
		endpoint := map[string]any{"resourceKind": "bead", "scope": "either", "types": map[string]any{"anyOf": []Reference{c.pin(name)}}}
		pairs = append(pairs, map[string]any{"source": endpoint, "target": endpoint})
	}
	expected, err := json.Marshal(map[string]any{"typeConstraints": pairs})
	if err != nil {
		return nil, err
	}
	if !equalJSON(c.definitions[c.pin("refines")].Properties["linkConstraints"], expected) {
		return nil, invalid("Refines category-preserving pairs")
	}
	if err := c.checkDeclarations(); err != nil {
		return nil, err
	}
	if err := c.checkSchemas(); err != nil {
		return nil, err
	}
	return c, nil
}

func decode(raw []byte, dst any) error {
	// encoding/json matches struct fields case-insensitively. The fixed wire
	// publication does not: check literal complete carrier members first.
	var fields []string
	var refs []string
	switch dst.(type) {
	case *publication:
		fields = []string{"definitions", "declarations"}
	case *definition:
		fields = []string{"id", "revision", "type", "properties", "ownedLinks"}
		refs = []string{"type"}
	case *declaration:
		fields = []string{"id", "revision", "type", "source", "target", "properties"}
		refs = []string{"type", "target"}
	}
	if fields != nil {
		var members map[string]json.RawMessage
		if err := json.Unmarshal(raw, &members); err != nil {
			return invalid("carrier shape: %v", err)
		}
		if len(members) != len(fields) {
			return invalid("unknown field or incomplete carrier")
		}
		for _, field := range fields {
			if _, ok := members[field]; !ok {
				return invalid("missing literal carrier field %s", field)
			}
		}
		for _, field := range refs {
			var ref map[string]json.RawMessage
			if err := json.Unmarshal(members[field], &ref); err != nil {
				return invalid("exact reference object: %v", err)
			}
			if len(ref) != 2 || ref["uri"] == nil || ref["revision"] == nil {
				return invalid("missing exact reference fields")
			}
		}
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	dec.UseNumber()
	if err := dec.Decode(dst); err != nil {
		return invalid("decode: %v", err)
	}
	if err := dec.Decode(new(any)); err != io.EOF {
		return invalid("extra JSON value")
	}
	return nil
}

func equalJSON(a, b []byte) bool {
	x, err := graph.CanonicalizeJSON(a)
	if err != nil {
		return false
	}
	y, err := graph.CanonicalizeJSON(b)
	return err == nil && bytes.Equal(x, y)
}

func (c *Catalog) checkDeclarations() error {
	indexed := map[Reference]json.RawMessage{}
	for _, raw := range c.publication.Declarations {
		var d declaration
		if err := decode(raw, &d); err != nil {
			return err
		}
		key := Reference{d.ID, d.Revision}
		if _, exists := indexed[key]; exists {
			return invalid("duplicate declaration")
		}
		if d.Properties == nil || d.Revision != "declaration-"+c.edition || d.Type != c.pin("refines") {
			return invalid("declaration exact pins/properties")
		}
		indexed[key] = raw
	}
	seen := map[Reference]bool{}
	edges := map[Reference][]Reference{}
	for ref, owner := range c.definitions {
		for nominal, group := range owner.Owned {
			if nominal != c.pin("refines").URI || len(group) == 0 {
				return invalid("invalid nominal/empty captured group")
			}
			for _, raw := range group {
				var d declaration
				if err := decode(raw, &d); err != nil {
					return err
				}
				key := Reference{d.ID, d.Revision}
				if seen[key] || !equalJSON(raw, indexed[key]) {
					return invalid("captured declaration mismatch/duplicate")
				}
				if d.Source != owner.ID {
					return invalid("captured source differs from owner")
				}
				if d.Type != c.pin("refines") {
					return invalid("wrong captured affiliation")
				}
				if _, ok := c.definitions[d.Target]; !ok {
					return invalid("missing exact Refines target")
				}
				// Category belongs to the definition's own metatype, not its instances.
				if owner.Type != c.definitions[d.Target].Type {
					return invalid("cross-category Refines")
				}
				for _, target := range edges[ref] {
					if target == d.Target {
						return invalid("duplicate exact Refines target")
					}
				}
				edges[ref] = append(edges[ref], d.Target)
				seen[key] = true
			}
		}
	}
	if len(seen) != len(indexed) {
		return invalid("uncaptured declaration")
	}
	var err error
	c.ancestors, err = refinesClosure(c.definitions, edges)
	if err != nil {
		return err
	}
	// Fixed selected topology; validation above separately checks actual captured
	// records, exact targets and DAG rather than relying on these expected names.
	for _, name := range []string{"bead-type", "link-type"} {
		key := Reference{namespace + name + "-refines-type", "declaration-" + c.edition}
		var d declaration
		if err := decode(indexed[key], &d); err != nil {
			return invalid("missing assigned declaration %s", name)
		}
		if d.Source != c.pin(name).URI || d.Target != c.pin("type") {
			return invalid("assigned declaration endpoints")
		}
	}
	for start := range c.definitions {
		visited := map[Reference]bool{}
		pending := []Reference{start}
		for len(pending) > 0 {
			ref := pending[len(pending)-1]
			pending = pending[:len(pending)-1]
			if visited[ref] {
				continue
			}
			d, ok := c.definitions[ref]
			if !ok {
				return invalid("missing exact definition dependency")
			}
			visited[ref] = true
			pending = append(pending, d.Type)
			for _, target := range edges[ref] {
				pending = append(pending, c.pin("refines"), target)
			}
		}
		if len(visited) != 4 {
			return invalid("incomplete exact intrinsic closure")
		}
	}
	return nil
}

func refinesClosure(defs map[Reference]definition, edges map[Reference][]Reference) (map[Reference][]Reference, error) {
	result := map[Reference][]Reference{}
	active := map[Reference]bool{}
	var visit func(Reference) error
	visit = func(ref Reference) error {
		if active[ref] {
			return invalid("exact Refines cycle")
		}
		if _, ok := result[ref]; ok {
			return nil
		}
		if _, ok := defs[ref]; !ok {
			return invalid("missing exact Refines target")
		}
		active[ref] = true
		set := map[Reference]bool{}
		for _, target := range edges[ref] {
			if err := visit(target); err != nil {
				return err
			}
			set[target] = true
			for _, parent := range result[target] {
				set[parent] = true
			}
		}
		ancestors := []Reference{}
		for parent := range set {
			ancestors = append(ancestors, parent)
		}
		sort.Slice(ancestors, func(i, j int) bool {
			if ancestors[i].URI == ancestors[j].URI {
				return ancestors[i].Revision < ancestors[j].Revision
			}
			return ancestors[i].URI < ancestors[j].URI
		})
		result[ref] = ancestors
		delete(active, ref)
		return nil
	}
	for ref := range defs {
		if err := visit(ref); err != nil {
			return nil, err
		}
	}
	return result, nil
}

type offlineLoader struct{}

func (offlineLoader) Load(uri string) (any, error) {
	return nil, invalid("schema resource unavailable offline: %s", uri)
}

type schemaSet struct {
	root      string
	resources map[string]json.RawMessage
}

func schemas(d definition) (schemaSet, error) {
	var root map[string]json.RawMessage
	if err := decode(d.Properties["propertiesSchema"], &root); err != nil {
		return schemaSet{}, err
	}
	var uri string
	if err := decode(root["$id"], &uri); err != nil {
		return schemaSet{}, err
	}
	resources := map[string]json.RawMessage{}
	if err := decode(d.Properties["schemaResources"], &resources); err != nil {
		return schemaSet{}, err
	}
	if resources == nil {
		return schemaSet{}, invalid("schemaResources object required")
	}
	if _, ok := resources[uri]; ok {
		return schemaSet{}, invalid("root resource replaced")
	}
	resources[uri] = d.Properties["propertiesSchema"]
	return schemaSet{uri, resources}, nil
}

// selectedStandardDigest recognizes only the eight exact standard resources
// in the qualified publication. Pinned jsonschema/v6.0.2 embeds JSON-value-
// identical standards and rejects AddResource for their reserved identities.
// Retained presence/content is checked separately; built-ins confer no trust.
func selectedStandardDigest(uri string) (string, bool) {
	switch uri {
	case dialect,
		"https://json-schema.org/draft/2020-12/meta/applicator",
		"https://json-schema.org/draft/2020-12/meta/content",
		"https://json-schema.org/draft/2020-12/meta/core",
		"https://json-schema.org/draft/2020-12/meta/format-annotation",
		"https://json-schema.org/draft/2020-12/meta/meta-data",
		"https://json-schema.org/draft/2020-12/meta/unevaluated",
		"https://json-schema.org/draft/2020-12/meta/validation":
		return selectedSchemaDigests()[uri], true
	default:
		return "", false
	}
}

func compileSchemas(set schemaSet) (*jsonschema.Schema, error) {
	compiler := jsonschema.NewCompiler()
	compiler.DefaultDraft(jsonschema.Draft2020)
	compiler.AssertFormat()
	compiler.UseLoader(offlineLoader{})
	keys := []string{}
	for key := range set.resources {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, uri := range keys {
		var obj map[string]json.RawMessage
		if err := decode(set.resources[uri], &obj); err != nil {
			return nil, err
		}
		var id, draft string
		if err := decode(obj["$id"], &id); err != nil {
			return nil, err
		}
		if err := decode(obj["$schema"], &draft); err != nil {
			return nil, err
		}
		if id != uri || draft != dialect {
			return nil, invalid("schema resource identity/dialect")
		}
		if want, standard := selectedStandardDigest(uri); standard {
			canonical, err := graph.CanonicalizeJSON(set.resources[uri])
			if err != nil || digest(canonical) != want {
				return nil, invalid("retained standard schema source drift: %s", uri)
			}
			// Do not re-add a reserved built-in or ignore ResourceExistsError.
			// The Compile loop below still compiles this exact standard ID.
			continue
		}
		value, err := jsonschema.UnmarshalJSON(bytes.NewReader(set.resources[uri]))
		if err != nil {
			return nil, invalid("schema JSON: %v", err)
		}
		if err := compiler.AddResource(uri, value); err != nil {
			return nil, invalid("schema resource: %v", err)
		}
	}
	for _, uri := range keys {
		if _, err := compiler.Compile(uri); err != nil {
			return nil, invalid("offline schema closure: %v", err)
		}
	}
	compiled, err := compiler.Compile(set.root)
	if err != nil {
		return nil, invalid("root schema: %v", err)
	}
	return compiled, nil
}

func (c *Catalog) checkSchemas() error {
	roots := map[string]json.RawMessage{}
	sets := map[Reference]schemaSet{}
	names := map[string]string{"type": "type-definition", "bead-type": "bead-type-definition", "link-type": "link-type-definition", "refines": "refines"}
	for name, stem := range names {
		ref := c.pin(name)
		set, err := schemas(c.definitions[ref])
		if err != nil {
			return err
		}
		if set.root != c.schemaNamespace()+stem+".properties.schema.json" {
			return invalid("unexpected intrinsic schema identity")
		}
		roots[set.root] = set.resources[set.root]
		sets[ref] = set
	}
	compiled := map[Reference]*jsonschema.Schema{}
	expected := selectedSchemaDigests()
	if c.edition == "2" {
		expected = currentSchemaDigests()
	}
	for _, name := range []string{"type", "bead-type", "link-type", "refines"} {
		ref := c.pin(name)
		set := sets[ref]
		for uri, raw := range set.resources {
			if strings.HasPrefix(uri, c.schemaNamespace()) && !equalJSON(raw, roots[uri]) {
				return invalid("conflicting retained intrinsic schema")
			}
			// Check every retained resource, even unused ones and compiler-reserved
			// standard URIs. Digests are fixed source facts, not caller authentication.
			canonical, err := graph.CanonicalizeJSON(raw)
			if err != nil {
				return err
			}
			if want, ok := expected[uri]; !ok || digest(canonical) != want {
				return invalid("retained schema source drift: %s", uri)
			}
		}
		// A compiler may have built-in standard documents. Require the selected
		// standard closure to be physically retained rather than borrow that fallback.
		for uri := range expected {
			if strings.HasPrefix(uri, "https://json-schema.org/") {
				if _, ok := set.resources[uri]; !ok {
					return invalid("missing retained standard resource")
				}
			}
		}
		if ref == c.pin("bead-type") || ref == c.pin("link-type") {
			if _, ok := set.resources[sets[c.pin("type")].root]; !ok {
				return invalid("missing retained Type schema")
			}
		}
		schema, err := compileSchemas(set)
		if err != nil {
			return err
		}
		compiled[ref] = schema
	}
	for _, d := range c.definitions {
		raw, err := json.Marshal(d.Properties)
		if err != nil {
			return err
		}
		value, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
		if err != nil {
			return err
		}
		if err := compiled[d.Type].Validate(value); err != nil {
			return invalid("metatype properties: %v", err)
		}
	}
	for _, raw := range c.publication.Declarations {
		var d declaration
		if err := decode(raw, &d); err != nil {
			return err
		}
		properties, err := json.Marshal(d.Properties)
		if err != nil {
			return err
		}
		value, err := jsonschema.UnmarshalJSON(bytes.NewReader(properties))
		if err != nil {
			return err
		}
		if err := compiled[c.pin("refines")].Validate(value); err != nil {
			return invalid("Refines properties: %v", err)
		}
	}
	return nil
}
