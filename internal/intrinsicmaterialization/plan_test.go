package intrinsicmaterialization

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/intrinsictypecatalog"
)

func selected(t *testing.T) *intrinsictypecatalog.Catalog {
	t.Helper()
	c, err := intrinsictypecatalog.LoadSelected()
	if err != nil {
		t.Fatal(err)
	}
	return c
}
func candidates() (Scope, []Allocation) {
	scope := Scope{"https://example.org/local/", "0123456789abcdef0123456789abcdef", "context-one"}
	out := make([]Allocation, 0, 6)
	for i, m := range members() {
		// Deliberately literal, segmented public IDs independent of private kind.
		id := "types/" + string(m.role)
		if m.role == Refines {
			id = "links/definition/refines"
		}
		path := "beads/" + string(m.role)
		if m.kind == graph.KindLink {
			path = "links/" + string(m.role)
		}
		out = append(out, Allocation{m.role, id, path, m.kind, Reference{scope.URI + id, fmt.Sprintf("edition-%d", i)}, scope.AuthorityID})
	}
	return scope, out
}
func assembled(t *testing.T) (*intrinsictypecatalog.Catalog, *Plan) {
	t.Helper()
	c := selected(t)
	s, a := candidates()
	p, err := Build(c, s, a)
	if err != nil {
		t.Fatal(err)
	}
	return c, p
}
func object(t *testing.T, raw []byte) map[string]json.RawMessage {
	t.Helper()
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		t.Fatal(err)
	}
	return obj
}
func equalJSON(t *testing.T, a, b []byte) {
	t.Helper()
	equal, err := graph.JSONEqual(a, b)
	if err != nil || !equal {
		t.Fatalf("different JSON values: %v", err)
	}
}
func requireRefusal(t *testing.T, err error, reason string) {
	t.Helper()
	if !errors.Is(err, ErrPlan) || !strings.Contains(err.Error(), reason) {
		t.Fatalf("want ErrPlan %q, got %v", reason, err)
	}
}

func TestAssemble(t *testing.T) {
	c, p := assembled(t)
	bindings := p.Bindings()
	if len(bindings) != 6 {
		t.Fatal("member census")
	}
	for i, b := range bindings {
		if b.Allocation.Role != members()[i].role {
			t.Fatal("fixed member order")
		}
		if got, ok := p.LookupLocal(b.Allocation.Local); !ok || got.Global != b.Global {
			t.Fatal("exact local lookup")
		}
		if got, ok := p.LocalFor(b.Global); !ok || got != b.Allocation.Local {
			t.Fatal("reverse exact lookup")
		}
		original := object(t, b.Original)
		local := object(t, b.Record)
		var id, revision string
		_ = json.Unmarshal(local["id"], &id)
		_ = json.Unmarshal(local["revision"], &revision)
		if id != b.Allocation.Local.URI || revision != b.Allocation.Local.Revision {
			t.Fatal("local identity")
		}
		equalJSON(t, original["properties"], local["properties"])
		equalJSON(t, original["type"], local["type"])
		if i < 4 {
			if b.Allocation.Kind != graph.KindBead {
				t.Fatal("all definitions are Beads")
			}
			source, ok := c.Definition(b.Global)
			if !ok || !bytes.Equal(source, b.Original) {
				t.Fatal("full original retained")
			}
			var owned map[string][]json.RawMessage
			if err := json.Unmarshal(local["ownedLinks"], &owned); err != nil {
				t.Fatal(err)
			}
			switch b.Allocation.Role {
			case Type, Refines:
				if len(owned) != 0 {
					t.Fatal("empty owned state")
				}
			case BeadType, LinkType:
				index := 4
				if b.Allocation.Role == LinkType {
					index = 5
				}
				bucket := owned[namespace+string(Refines)]
				if len(owned) != 1 || len(bucket) != 1 {
					t.Fatal("complete exact Refines bucket")
				}
				equalJSON(t, bucket[0], bindings[index].Record)
			}
		} else {
			equalJSON(t, b.Original, c.Declarations()[i-4])
			var source string
			var target Reference
			_ = json.Unmarshal(local["source"], &source)
			_ = json.Unmarshal(local["target"], &target)
			owner := bindings[i-3].Allocation.Local // indexes 1 and 2
			if source != owner.URI || target != bindings[0].Allocation.Local {
				t.Fatal("actual local source/exact target")
			}
			if bytes.Equal(b.Original, b.Record) {
				t.Fatal("local Link is not relabeled global capture")
			}
		}
	}
	if kind, ok := c.Category(bindings[3].Global); !ok || kind != graph.KindLink || bindings[3].Allocation.Kind != graph.KindBead {
		t.Fatal("instance kind is not definition kind")
	}
	// Equal opaque revision tokens across distinct Resources are legal.
	s, a := candidates()
	for i := range a {
		a[i].Local.Revision = "same-token"
	}
	if _, err := Build(c, s, a); err != nil {
		t.Fatal("revision tokens need not be globally unique:", err)
	}
}

func TestAllocationRefusals(t *testing.T) {
	c := selected(t)
	cases := []struct {
		name, reason string
		change       func(*Scope, *[]Allocation)
	}{
		{"missing", "exactly six", func(_ *Scope, a *[]Allocation) { *a = (*a)[:5] }},
		{"extra", "exactly six", func(_ *Scope, a *[]Allocation) { *a = append(*a, (*a)[0]) }},
		{"duplicate-role", "duplicate role", func(_ *Scope, a *[]Allocation) { (*a)[1].Role = Type }},
		{"unknown-role", "unknown", func(_ *Scope, a *[]Allocation) { (*a)[0].Role = "other" }},
		{"wrong-kind", "wrong Resource kind", func(_ *Scope, a *[]Allocation) { (*a)[3].Kind = graph.KindLink }},
		{"wrong-authority", "inconsistent authority", func(_ *Scope, a *[]Allocation) { (*a)[0].AuthorityID = "fedcba9876543210fedcba9876543210" }},
		{"bad-authority", "32 lowercase", func(s *Scope, _ *[]Allocation) { s.AuthorityID = "A" }},
		{"wrong-scope", "identity does not match", func(_ *Scope, a *[]Allocation) { (*a)[0].Local.URI = "https://elsewhere.org/types/type" }},
		{"wrong-public-id", "identity does not match", func(_ *Scope, a *[]Allocation) { (*a)[0].PublicID = "other" }},
		{"invalid-public-id", "identity does not match", func(_ *Scope, a *[]Allocation) { (*a)[0].PublicID = "../type" }},
		{"wrong-locator-kind", "private locator", func(_ *Scope, a *[]Allocation) { (*a)[0].PrivatePath = "links/type" }},
		{"duplicate-identity", "duplicate candidate", func(_ *Scope, a *[]Allocation) {
			(*a)[1].PublicID = (*a)[0].PublicID
			(*a)[1].Local.URI = (*a)[0].Local.URI
		}},
		{"duplicate-locator", "duplicate candidate", func(_ *Scope, a *[]Allocation) { (*a)[1].PrivatePath = (*a)[0].PrivatePath }},
		{"empty-revision", "revision", func(_ *Scope, a *[]Allocation) { (*a)[0].Local.Revision = "" }},
		{"invalid-utf8-revision", "revision", func(_ *Scope, a *[]Allocation) { (*a)[0].Local.Revision = string([]byte{0xff}) }},
		{"empty-context", "context", func(s *Scope, _ *[]Allocation) { s.ContextID = "" }},
		{"bad-scope", "Scope", func(s *Scope, _ *[]Allocation) { s.URI = "https://example.org/local" }},
		{"shared-identity", "rebind", func(s *Scope, a *[]Allocation) {
			s.URI = namespace
			for i := range *a {
				(*a)[i].PublicID = string((*a)[i].Role)
				(*a)[i].Local.URI = namespace + string((*a)[i].Role)
			}
		}},
		{"budget", "planning budget", func(_ *Scope, a *[]Allocation) { (*a)[0].Local.Revision = strings.Repeat("x", MaxContextBytes) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s, a := candidates()
			tc.change(&s, &a)
			_, err := Build(c, s, a)
			requireRefusal(t, err, tc.reason)
		})
	}
}

func TestCatalogRefusal(t *testing.T) {
	s, a := candidates()
	for _, c := range []*intrinsictypecatalog.Catalog{nil, new(intrinsictypecatalog.Catalog)} {
		_, err := Build(c, s, a)
		requireRefusal(t, err, "catalog")
	}
	var zero Plan
	_, err := zero.EncodeContext()
	requireRefusal(t, err, "uninitialized")
	var nilPlan *Plan
	_, err = nilPlan.EncodeContext()
	requireRefusal(t, err, "uninitialized")
	if _, ok := zero.LookupLocal(Reference{}); ok {
		t.Fatal("zero lookup")
	}
	if _, ok := nilPlan.LocalFor(Reference{}); ok {
		t.Fatal("nil reverse lookup")
	}
}

func TestContextRoundTrip(t *testing.T) {
	c, p := assembled(t)
	raw, err := p.EncodeContext()
	if err != nil {
		t.Fatal(err)
	}
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, raw, "", "  "); err != nil {
		t.Fatal(err)
	}
	got, err := DecodeContext(c, pretty.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	again, err := got.EncodeContext()
	if err != nil || !bytes.Equal(raw, again) {
		t.Fatal("deterministic complete round trip", err)
	}
	_, err = DecodeContext(nil, raw)
	requireRefusal(t, err, "catalog")
}

func changedContext(t *testing.T, p *Plan, change func(*contextPayload)) []byte {
	t.Helper()
	raw, err := p.EncodeContext()
	if err != nil {
		t.Fatal(err)
	}
	var env envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatal(err)
	}
	change(&env.Payload)
	payload, err := canonicalPayload(env.Payload)
	if err != nil {
		t.Fatal(err)
	}
	env.SHA256 = digest(payload)
	raw, err = json.Marshal(env)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestContextRefusals(t *testing.T) {
	c, p := assembled(t)
	raw, err := p.EncodeContext()
	if err != nil {
		t.Fatal(err)
	}
	withoutChecksum, err := json.Marshal(map[string]any{"payload": p.payload})
	if err != nil {
		t.Fatal(err)
	}
	shapes := []struct {
		name, reason string
		raw          []byte
	}{
		{"empty", "context JSON", nil},
		{"oversize", "planning budget", bytes.Repeat([]byte(" "), MaxContextBytes+1)},
		{"duplicate", "context JSON", []byte(`{"payload":{},"payload":{},"sha256":"x"}`)},
		{"case-confusion", "context differs", bytes.Replace(raw, []byte(`"payload":`), []byte(`"Payload":`), 1)},
		{"unknown", "context shape", bytes.Replace(raw, []byte(`"payload":{`), []byte(`"payload":{"unknown":1,`), 1)},
		{"invalid-utf8", "context JSON", []byte{'"', 0xff, '"'}},
		{"surrogate", "context JSON", []byte(`{"payload":"\ud800"}`)},
		{"trailing", "context JSON", append(append([]byte(nil), raw...), []byte(` {}`)...)},
		{"missing-payload", "context checksum", []byte(`{"sha256":"x"}`)},
		{"missing-checksum", "context checksum", withoutChecksum},
		{"checksum-corruption", "context checksum", bytes.Replace(raw, []byte(`"sha256":"`), []byte(`"sha256":"changed`), 1)},
	}
	for _, tc := range shapes {
		t.Run(tc.name, func(t *testing.T) { _, err := DecodeContext(c, tc.raw); requireRefusal(t, err, tc.reason) })
	}
	mutations := []struct {
		name, reason string
		change       func(*contextPayload)
	}{
		{"encoding", "context differs", func(x *contextPayload) { x.Encoding = "other" }},
		{"semantic-source", "context differs", func(x *contextPayload) { x.SemanticSource = "other" }},
		{"artifact-source", "context differs", func(x *contextPayload) { x.ArtifactSource.Commit = "other" }},
		{"binding-global", "context differs", func(x *contextPayload) { x.Bindings[0].Global.Revision = "other" }},
		{"original", "context differs", func(x *contextPayload) { x.Bindings[0].Original = json.RawMessage(`{}`) }},
		{"wrong-local-affiliation", "context differs", func(x *contextPayload) {
			f := object(t, x.Bindings[3].Record)
			put(f, "type", x.Bindings[0].Global)
			x.Bindings[3].Record, _ = json.Marshal(f)
		}},
		{"floating-target", "context differs", func(x *contextPayload) {
			f := object(t, x.Bindings[4].Record)
			put(f, "target", x.Bindings[0].Allocation.Local.URI)
			x.Bindings[4].Record, _ = json.Marshal(f)
		}},
		{"global-source", "context differs", func(x *contextPayload) {
			f := object(t, x.Bindings[4].Record)
			put(f, "source", x.Bindings[1].Global.URI)
			x.Bindings[4].Record, _ = json.Marshal(f)
		}},
		{"rewritten-selector", "context differs", func(x *contextPayload) {
			x.Bindings[0].Record = bytes.Replace(x.Bindings[0].Record, []byte("https://github.com/gastownhall/bdp/schemas/types/intrinsic-1/"), []byte("https://example.org/local/schemas/"), 1)
		}},

		{"local-body", "context differs", func(x *contextPayload) { x.Bindings[0].Record = json.RawMessage(`{}`) }},
		{"capture", "context differs", func(x *contextPayload) {
			f := object(t, x.Bindings[1].Record)
			put(f, "ownedLinks", map[string]any{})
			x.Bindings[1].Record, _ = json.Marshal(f)
		}},
		{"missing-role", "exactly six", func(x *contextPayload) { x.Bindings = x.Bindings[:5] }},
		{"reordered", "context differs", func(x *contextPayload) { x.Bindings[0], x.Bindings[1] = x.Bindings[1], x.Bindings[0] }},
		{"empty-revision", "revision", func(x *contextPayload) { x.Bindings[0].Allocation.Local.Revision = "" }},
	}
	for _, tc := range mutations {
		t.Run(tc.name, func(t *testing.T) {
			altered := changedContext(t, p, tc.change)
			_, err := DecodeContext(c, altered)
			requireRefusal(t, err, tc.reason)
		})
	}
}

func TestIndependentContexts(t *testing.T) {
	c, p := assembled(t)
	raw, err := p.EncodeContext()
	if err != nil {
		t.Fatal(err)
	}
	bs := p.Bindings()
	old := bs[0].Allocation.Local
	global := bs[0].Global
	bs[0].Record[0] = '!'
	bs[0].Original[0] = '!'
	bs[0].Allocation.Local.Revision = "changed"
	found, ok := p.LookupLocal(old)
	if !ok {
		t.Fatal("lost old binding")
	}
	found.Record[0] = '!'
	again, err := p.EncodeContext()
	if err != nil || !bytes.Equal(raw, again) {
		t.Fatal("access leaked mutable state")
	}
	s, a := candidates()
	s.ContextID = "context-two"
	for i := range a {
		a[i].Local.Revision = "new-edition"
	}
	next, err := Build(c, s, a)
	if err != nil {
		t.Fatal(err)
	}
	newer, ok := next.LocalFor(global)
	if !ok || newer == old {
		t.Fatal("new independent context")
	}
	if got, ok := p.LocalFor(global); !ok || got != old {
		t.Fatal("old context changed")
	}
	if _, ok := p.LookupLocal(newer); ok {
		t.Fatal("current fallback")
	}
	if _, ok := p.LocalFor(Reference{URI: global.URI}); ok {
		t.Fatal("unpinned fallback")
	}
	a[0].Local.Revision = "caller-mutated"
	if got, _ := next.LocalFor(global); got != newer {
		t.Fatal("input slice retained")
	}
	decoded, err := DecodeContext(c, raw)
	if err != nil {
		t.Fatal(err)
	}
	raw[0] = '!'
	if got, _ := decoded.LocalFor(global); got != old {
		t.Fatal("decoder retained caller bytes")
	}
}
