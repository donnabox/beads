package graphops

import (
	"database/sql"
	"errors"
	"testing"

	graph "github.com/steveyegge/beads/graphops"
)

const fixtureScope = "https://graph.example/demo/"
const memoryType = "https://graph.example/types/memory"
const relationType = "https://graph.example/types/explains"

var fixtureLimits = readLimits{rows: 16, bytes: 64 << 10, valueBytes: 8 << 10}

func present(s string) sql.NullString { return sql.NullString{String: s, Valid: true} }
func validRow() resourceRow {
	return resourceRow{path: "beads/plan", typeURL: memoryType, revision: "unchanged opaque revision", properties: []byte(`{"text":"launch Monday"}`)}
}
func validLinkRow() linkRow {
	row := validRow()
	row.path = "links/plan-decision"
	row.typeURL = relationType
	return linkRow{resourceRow: row, source: endpointRow{kind: "in", path: present("beads/plan")}, target: endpointRow{kind: "in", path: present("beads/decision")}}
}
func memoryDescriptor(t *testing.T, owns ...graph.OwnedLinkDecl) graph.TypeDescriptor {
	t.Helper()
	d, e := graph.NewTypeDescriptor(graph.TypeDescriptorSpec{ID: memoryType, Name: "Memory", Describes: graph.KindBead, OwnsOutgoing: owns})
	if e != nil {
		t.Fatal(e)
	}
	return d
}
func relationDescriptor(t *testing.T) graph.TypeDescriptor {
	t.Helper()
	endpoint, e := graph.NewEndpointConstraint(nil, graph.ExternalOpaque)
	if e != nil {
		t.Fatal(e)
	}
	d, e := graph.NewTypeDescriptor(graph.TypeDescriptorSpec{ID: relationType, Name: "Explains", Describes: graph.KindLink, Source: &endpoint, Target: &endpoint})
	if e != nil {
		t.Fatal(e)
	}
	return d
}

func TestPersistedBeadRefusesRepair(t *testing.T) {
	cases := map[string]func(*resourceRow){
		"noncanonical JSON":     func(r *resourceRow) { r.properties = []byte(`{ "text": "launch Monday" }`) },
		"invalid JSON":          func(r *resourceRow) { r.properties = []byte(`{`) },
		"invalid UTF8":          func(r *resourceRow) { r.properties = []byte{'{', '"', 'x', '"', ':', '"', 0xff, '"', '}'} },
		"noncanonical path":     func(r *resourceRow) { r.path = "beads/../plan" },
		"empty revision":        func(r *resourceRow) { r.revision = "" },
		"invalid revision UTF8": func(r *resourceRow) { r.revision = "\xff" },
		"partial attribution":   func(r *resourceRow) { r.principal = present("author") },
		"unknown attribution":   func(r *resourceRow) { r.principal = present("author"); r.attribution = present("made-up") },
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			r := validRow()
			change(&r)
			_, err := decodeBead(r)
			if !errors.Is(err, errCorrupt) || errors.Is(err, graph.ErrValidation) || errors.Is(err, graph.ErrNotFound) {
				t.Fatalf("error=%v", err)
			}
		})
	}
	bead, err := decodeBead(validRow())
	if err != nil {
		t.Fatal(err)
	}
	if bead.Revision().String() != validRow().revision || bead.Properties().String() != string(validRow().properties) {
		t.Fatal("persisted values changed")
	}
}

func TestPersistedEndpointClassificationAndPins(t *testing.T) {
	row := validLinkRow()
	row.target = endpointRow{kind: "ext", url: present("urn:report:external"), pin: present("  exact pin — α/β\t")}
	link, err := decodeLink(fixtureScope, row)
	if err != nil {
		t.Fatal(err)
	}
	if link.Target().URI() != row.target.url.String || link.Target().Pin() != row.target.pin.String {
		t.Fatal("opaque endpoint changed")
	}
	for name, endpoint := range map[string]endpointRow{
		"external local canonical": {kind: "ext", url: present(fixtureScope + "beads/decision")},
		"external local alias":     {kind: "ext", url: present("https://GRAPH.example:443/demo/beads/decision")},
		"both path and URI":        {kind: "ext", path: present("beads/decision"), url: present("urn:x")},
		"missing path":             {kind: "in"},
		"empty pin":                {kind: "in", path: present("beads/decision"), pin: present("")},
		"invalid pin":              {kind: "in", path: present("beads/decision"), pin: present("\xff")},
		"unknown kind":             {kind: "other", url: present("urn:x")},
	} {
		t.Run(name, func(t *testing.T) {
			r := validLinkRow()
			r.target = endpoint
			_, err := decodeLink(fixtureScope, r)
			if !errors.Is(err, errCorrupt) || errors.Is(err, graph.ErrValidation) || errors.Is(err, graph.ErrNotFound) {
				t.Fatalf("error=%v", err)
			}
		})
	}
	row.source = endpointRow{kind: "ext", url: present("urn:source")}
	if _, err := decodeLink(fixtureScope, row); !errors.Is(err, errCorrupt) {
		t.Fatalf("two external endpoints: %v", err)
	}
}

func TestPersistedDescriptorRefusesRepairOrMismatch(t *testing.T) {
	d := memoryDescriptor(t)
	if _, err := decodeDescriptor(d.ID(), d.CanonicalJSON(), d.Fingerprint(), graph.KindBead); err != nil {
		t.Fatal(err)
	}
	for name, change := range map[string]func() (string, []byte, string, graph.ResourceKind){
		"noncanonical": func() (string, []byte, string, graph.ResourceKind) {
			return d.ID(), append(d.CanonicalJSON(), ' '), d.Fingerprint(), graph.KindBead
		},
		"identity": func() (string, []byte, string, graph.ResourceKind) {
			return relationType, d.CanonicalJSON(), d.Fingerprint(), graph.KindBead
		},
		"fingerprint": func() (string, []byte, string, graph.ResourceKind) {
			return d.ID(), d.CanonicalJSON(), "wrong", graph.KindBead
		},
		"kind": func() (string, []byte, string, graph.ResourceKind) {
			return d.ID(), d.CanonicalJSON(), d.Fingerprint(), graph.KindLink
		},
	} {
		t.Run(name, func(t *testing.T) {
			id, raw, fp, kind := change()
			_, err := decodeDescriptor(id, raw, fp, kind)
			if !errors.Is(err, errCorrupt) || errors.Is(err, graph.ErrValidation) || errors.Is(err, graph.ErrNotFound) {
				t.Fatalf("error=%v", err)
			}
		})
	}
}
