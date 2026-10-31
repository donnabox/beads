package resourcequery

import (
	"context"
	"errors"
	"testing"

	graph "github.com/steveyegge/beads/graphops"
)

func currentFact(id, revision string, kind graph.ResourceKind) CurrentEditionFact {
	return CurrentEditionFact{Reference: Edition{ID: base + id, Revision: revision}, Kind: kind}
}
func currentEndpointSelection(id string) ast {
	return ast{"op": "endpoints", "input": refset("links", base+id), "side": "target"}
}

// These facts stand for a producer's independently checked current capture.
// They are fixture assertions, not native provenance. Seeds are floating refs;
// only stored endpoint pins reach the opt-in dynamic body-acquisition gate.
func TestCurrentEditionAuthorityOptInAndKindCoverage(t *testing.T) {
	const token = "opaque/@=https://edition.example/a?x=1#bytes"
	d := Dataset{
		Resources: []Resource{
			resource(t, graph.KindBead, "A", "Note", token, ast{}),
			resource(t, graph.KindBead, "TypeDefinition", "BeadType", "type-local-7", ast{}),
			link(t, "to-generic", "Requires", ast{"uri": base + "A", "revision": token}, ast{"uri": base + "A", "revision": token}),
			link(t, "to-type", "Requires", ast{"uri": base + "A", "revision": token}, ast{"uri": base + "TypeDefinition", "revision": "type-local-7"}),
		},
		CurrentEditions: NewCurrentEditionAuthorityFromTrustedCapture([]CurrentEditionFact{
			currentFact("A", token, graph.KindBead),
			currentFact("TypeDefinition", "type-local-7", graph.KindBead),
			currentFact("to-generic", "l1", graph.KindLink), currentFact("to-type", "l1", graph.KindLink),
		}),
	}
	assertIDs(t, execute(t, request(t, "query", "resources", currentEndpointSelection("to-generic")), d), "A")
	assertIDs(t, execute(t, request(t, "query", "resources", currentEndpointSelection("to-type")), d), "TypeDefinition")
	selection := filter(scan("links"), "l", compare("eq", field(endpoint(binding("l"), "target"), "revision"), literal(token)))
	assertIDs(t, execute(t, request(t, "query", "resources", selection), d), "to-generic")
	response := execute(t, request(t, "query", "graph", walk(refset("beads", base+"A"), "outgoing", 1)), d)
	result, ok := response.Result.(GraphResult)
	if !ok || len(result.Endpoints) != 2 || len(result.Beads) != 2 ||
		result.Endpoints[0].Source.Revision != token || result.Endpoints[0].Target.Revision != token ||
		result.Endpoints[1].Target != (Edition{ID: base + "TypeDefinition", Revision: "type-local-7"}) {
		t.Fatalf("dynamic generic/Type body editions lost: %+v", response.Result)
	}
}

func TestCurrentEditionAuthorityCannotComeFromMatchingBodyOrTypeFacts(t *testing.T) {
	for _, token := range []string{"r1", "current"} {
		d := Dataset{
			Resources: []Resource{
				resource(t, graph.KindBead, "A", "Note", token, ast{}),
				link(t, "self", "Requires", base+"A", ast{"uri": base + "A", "revision": token}),
			},
			TypeEditions: []TypeEditionFact{{Reference: Edition{ID: base + "Note", Revision: "type-1"}, Kind: graph.KindBead, Current: true, Ancestors: []Edition{{ID: base + "Note", Revision: "type-1"}}}},
		}
		assertFailure(t, request(t, "query", "resources", currentEndpointSelection("self")), d, ErrUnsupported)
		assertIDs(t, execute(t, request(t, "query", "resources", refset("beads", base+"A")), d), "A")
		stored := filter(scan("links"), "l", compare("eq", field(binding("l"), "target"), literal(ast{"uri": base + "A", "revision": token})))
		assertIDs(t, execute(t, request(t, "query", "resources", stored), d), "self")
	}
}

func TestCurrentEditionAuthorityUnsupportedOutsideCurrentCapture(t *testing.T) {
	d := Dataset{
		Resources:       []Resource{bead(t, "A", ast{}), link(t, "self", "Requires", base+"A", ast{"uri": base + "A", "revision": "r1"})},
		CurrentEditions: NewCurrentEditionAuthorityFromTrustedCapture([]CurrentEditionFact{currentFact("A", "r1", graph.KindBead), currentFact("self", "l1", graph.KindLink)}),
	}
	// Explicit pinned refs are still refused by Decode BEFORE any Dataset or
	// authority is inspected. This applies with nil capability and with opt-in.
	for _, authority := range []*CurrentEditionAuthority{nil, d.CurrentEditions} {
		d.CurrentEditions = authority
		for _, tc := range []struct{ collection, uri, revision string }{
			{"beads", base + "A", "r1"}, {"links", base + "self", "l1"},
			{"beads", base + "A", "old"}, {"beads", base + "A", "current"},
			{"beads", base + "missing", "r1"}, {"beads", "https://foreign.example/A", "r1"},
			{"beads", base + "self", "l1"}, {"links", base + "A", "r1"},
		} {
			raw := request(t, "query", "resources", refset(tc.collection, ast{"uri": tc.uri, "revision": tc.revision}))
			if _, err := Decode(raw); !errors.Is(err, ErrUnsupported) {
				t.Fatalf("static pin bypassed pre-data admission: %v", err)
			}
			assertFailure(t, raw, d, ErrUnsupported)
		}
		condition := ast{"op": "or", "operands": []any{true, ast{"op": "any", "input": refset("beads", ast{"uri": base + "A", "revision": "r1"})}}}
		raw := request(t, "query", "resources", filter(scan("beads"), "b", condition))
		if _, err := Decode(raw); !errors.Is(err, ErrUnsupported) {
			t.Fatalf("skipped static pin bypassed admission: %v", err)
		}
		assertFailure(t, raw, d, ErrUnsupported)
	}
	// Admitted floating Link seeds can reach dynamic pins, which still refuse
	// retained, synthetic-current, missing, foreign or wrong-kind body editions.
	for _, tc := range []struct{ uri, revision string }{
		{base + "A", "old"}, {base + "A", "current"}, {base + "missing", "r1"},
		{"https://foreign.example/A", "r1"}, {base + "edge", "l1"},
	} {
		dynamic := Dataset{
			Resources:       []Resource{bead(t, "A", ast{}), link(t, "edge", "Requires", base+"A", ast{"uri": tc.uri, "revision": tc.revision})},
			CurrentEditions: NewCurrentEditionAuthorityFromTrustedCapture([]CurrentEditionFact{currentFact("A", "r1", graph.KindBead), currentFact("edge", "l1", graph.KindLink)}),
		}
		raw := request(t, "query", "resources", currentEndpointSelection("edge"))
		if _, err := Decode(raw); err != nil {
			t.Fatalf("floating seed must pass admission: %v", err)
		}
		assertFailure(t, raw, dynamic, ErrUnsupported)
	}
	assertFailure(t, request(t, "query", "resources", refset("beads", base+"missing")), d, ErrUnavailable)
}

func TestCurrentEditionAuthorityRejectsIncompleteOrInconsistentCapture(t *testing.T) {
	full := []CurrentEditionFact{currentFact("A", "r1", graph.KindBead), currentFact("self", "l1", graph.KindLink)}
	for _, tc := range []struct {
		name  string
		facts []CurrentEditionFact
	}{
		{"empty authority for nonempty capture", nil}, {"Link coverage absent", full[:1]}, {"Bead coverage absent", full[1:]},
		{"duplicate identity", []CurrentEditionFact{full[0], full[0], full[1]}},
		{"extra identity", append(append([]CurrentEditionFact{}, full...), currentFact("missing", "r1", graph.KindBead))},
		{"retained edition differs", []CurrentEditionFact{currentFact("A", "old", graph.KindBead), full[1]}},
		{"empty edition", []CurrentEditionFact{currentFact("A", "", graph.KindBead), full[1]}},
		{"wrong kind", []CurrentEditionFact{currentFact("A", "r1", graph.KindLink), full[1]}},
		{"unknown kind", []CurrentEditionFact{currentFact("A", "r1", graph.ResourceKind("other")), full[1]}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := Dataset{Resources: []Resource{bead(t, "A", ast{}), link(t, "self", "Requires", base+"A", base+"A")}, CurrentEditions: NewCurrentEditionAuthorityFromTrustedCapture(tc.facts)}
			assertFailure(t, request(t, "query", "resources", scan("beads")), d, ErrDataset)
		})
	}
	d := Dataset{CurrentEditions: &CurrentEditionAuthority{}}
	assertFailure(t, request(t, "query", "resources", scan("beads")), d, ErrDataset)
	d.CurrentEditions = NewCurrentEditionAuthorityFromTrustedCapture(nil)
	assertIDs(t, execute(t, request(t, "query", "resources", scan("beads")), d))
	// Static pins fail admission before malformed data; admitted dynamic body
	// acquisition instead validates the FULL dataset before resolving endpoints.
	d = Dataset{Resources: []Resource{
		bead(t, "A", ast{}), link(t, "self", "Requires", base+"A", ast{"uri": base + "A", "revision": "r1"}),
		{Kind: graph.KindBead, Data: []byte(`{}`)},
	}, CurrentEditions: NewCurrentEditionAuthorityFromTrustedCapture(full)}
	static := request(t, "query", "resources", refset("beads", ast{"uri": base + "A", "revision": "r1"}))
	for _, authority := range []*CurrentEditionAuthority{nil, d.CurrentEditions} {
		d.CurrentEditions = authority
		assertFailure(t, static, d, ErrUnsupported)
		assertFailure(t, request(t, "query", "resources", currentEndpointSelection("self")), d, ErrDataset)
	}
}

func TestCurrentEditionAuthorityPreservesIncidenceAndBoundary(t *testing.T) {
	d := Dataset{
		Resources: []Resource{
			bead(t, "A", ast{}), bead(t, "B", ast{}),
			link(t, "current", "Requires", base+"A", ast{"uri": base + "B", "revision": "r1"}),
			link(t, "old", "Requires", base+"A", ast{"uri": base + "B", "revision": "old"}),
		},
		CurrentEditions: NewCurrentEditionAuthorityFromTrustedCapture([]CurrentEditionFact{
			currentFact("A", "r1", graph.KindBead), currentFact("B", "r1", graph.KindBead),
			currentFact("current", "l1", graph.KindLink), currentFact("old", "l1", graph.KindLink),
		}),
	}
	assertIDs(t, execute(t, request(t, "query", "resources", walk(refset("beads", base+"B"), "incoming", 1)), d), "current")
	assertIDs(t, execute(t, request(t, "query", "resources", walk(refset("beads", base+"A"), "outgoing", 1)), d), "current", "old")
	assertFailure(t, request(t, "query", "resources", walk(refset("beads", base+"A"), "outgoing", 2)), d, ErrUnsupported)
	assertFailure(t, request(t, "query", "graph", walk(refset("beads", base+"A"), "outgoing", 1)), d, ErrUnsupported)
	assertFailure(t, request(t, "query", "resources", currentEndpointSelection("old")), d, ErrUnsupported)
	w := walk(refset("beads", base+"A"), "outgoing", 2)
	w["bind"] = "l"
	w["through"] = false
	assertIDs(t, execute(t, request(t, "query", "resources", w), d))
	w["through"] = compare("eq", field(binding("l"), "id"), literal(base+"current"))
	assertIDs(t, execute(t, request(t, "query", "resources", w), d), "current")
	response := execute(t, request(t, "query", "graph", w), d)
	result, ok := response.Result.(GraphResult)
	if !ok || len(result.Beads) != 2 || len(result.Links) != 1 || len(result.Endpoints) != 1 || result.Endpoints[0].Target != (Edition{ID: base + "B", Revision: "r1"}) {
		t.Fatalf("current graph acquisition changed: %+v", response.Result)
	}
}

func TestCurrentEditionAuthorityDetachesFactsAndRechecksCorrespondence(t *testing.T) {
	facts := []CurrentEditionFact{currentFact("A", "r1", graph.KindBead), currentFact("self", "l1", graph.KindLink)}
	d := Dataset{
		Resources:       []Resource{bead(t, "A", ast{}), link(t, "self", "Requires", base+"A", ast{"uri": base + "A", "revision": "r1"})},
		CurrentEditions: NewCurrentEditionAuthorityFromTrustedCapture(facts),
	}
	facts[0] = currentFact("A", "forged", graph.KindLink)
	assertIDs(t, execute(t, request(t, "query", "resources", currentEndpointSelection("self")), d), "A")
	d.Resources[0] = resource(t, graph.KindBead, "A", "Note", "r2", ast{})
	assertFailure(t, request(t, "query", "resources", scan("beads")), d, ErrDataset)
	d.Resources[0] = bead(t, "A", ast{})
	d.Resources = append(d.Resources, bead(t, "unrelated", ast{}))
	assertFailure(t, request(t, "query", "resources", scan("beads")), d, ErrDataset)
}

func TestCurrentEditionAuthorityKeepsCancellationAndWorkBounds(t *testing.T) {
	d := Dataset{Resources: []Resource{bead(t, "A", ast{})}, CurrentEditions: NewCurrentEditionAuthorityFromTrustedCapture([]CurrentEditionFact{currentFact("A", "r1", graph.KindBead)})}
	raw := request(t, "query", "resources", scan("beads"))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	response, err := Execute(ctx, raw, d, Limits{})
	if !errors.Is(err, context.Canceled) || response.Result != nil || response.Evaluation.Complete {
		t.Fatalf("cancellation returned usable data: %+v %v", response, err)
	}
	limits := DefaultLimits()
	limits.MaxWork = 1
	response, err = Execute(context.Background(), raw, d, limits)
	if !errors.Is(err, ErrLimit) || response.Result != nil || response.Evaluation.Complete {
		t.Fatalf("authority work was unbounded: %+v %v", response, err)
	}
}
