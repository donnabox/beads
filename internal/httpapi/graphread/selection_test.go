package graphread

import (
	"encoding/json"
	"errors"
	"net/url"
	"reflect"
	"testing"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/httpapi/bdpwire"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/resourceget"
)

const selectionScope = "https://example.test/read/"

var selectionLimits = SelectorLimits{Bytes: 16384, Depth: 256, Nodes: 2048}

func selectionFixture() Inventory {
	bead := func(id, typ, status string) bdpwire.BeadRecord {
		return bdpwire.BeadRecord{ID: selectionScope + "beads/" + id, Type: selectionScope + "types/" + typ, Revision: "old",
			Properties: bdpwire.Properties{"status": json.RawMessage(`"` + status + `"`)}}
	}
	link := func(id, source, target, pin string) bdpwire.LinkRecord {
		return bdpwire.LinkRecord{ID: selectionScope + "links/" + id, Type: selectionScope + "types/link", Revision: "l1",
			Source: bdpwire.Reference{URI: source, Revision: pin}, Target: bdpwire.Reference{URI: target}, Properties: bdpwire.Properties{}}
	}
	return Inventory{
		Beads: []bdpwire.BeadRecord{bead("z", "child", "open"), bead("a", "parent", "closed"), bead("m", "child", "closed")},
		Links: []bdpwire.LinkRecord{
			link("z", selectionScope+"beads/z", "urn:partner:UPPER", "old"),
			link("a", selectionScope+"beads/a", selectionScope+"beads/z", ""),
			link("self", selectionScope+"beads/z", selectionScope+"beads/z", ""),
		},
		Types: []bdpwire.TypeDescriptor{
			{ID: selectionScope + "types/child", Name: "Child", Describes: bdpwire.DescribesBead, ConformsTo: bdpwire.TypeIDs{selectionScope + "types/parent"}},
			{ID: selectionScope + "types/parent", Name: "Parent", Describes: bdpwire.DescribesBead},
			{ID: selectionScope + "types/link", Name: "Relationship", Describes: bdpwire.DescribesLink},
		},
	}
}

func selectedIDs(t *testing.T, rows []json.RawMessage) []string {
	t.Helper()
	ids := []string{}
	for _, row := range rows {
		var value struct{ ID string }
		if err := json.Unmarshal(row, &value); err != nil {
			t.Fatal(err)
		}
		ids = append(ids, value.ID)
	}
	return ids
}

func TestCollectionPredicates(t *testing.T) {
	for _, tc := range []struct {
		name, collection string
		query            url.Values
		ids              []string
	}{
		{"all", "beads", nil, []string{"beads/a", "beads/m", "beads/z"}},
		{"exact", "beads", url.Values{"type": {selectionScope + "types/child"}}, []string{"beads/m", "beads/z"}},
		{"self and inherited", "beads", url.Values{"conformsTo": {selectionScope + "types/parent"}}, []string{"beads/a", "beads/m", "beads/z"}},
		{"AND", "beads", url.Values{"type": {selectionScope + "types/child"}, "conformsTo": {selectionScope + "types/parent"}, "selector": {`$[?@.properties.status == "closed"]`}, "limit": {"1"}}, []string{"beads/m"}},
		{"unknown Type", "beads", url.Values{"type": {"https://unknown.test/type"}}, []string{}},
		{"local pin transparent", "links", url.Values{"source": {"beads/z"}, "target": {"urn:partner:UPPER"}}, []string{"links/z"}},
		{"external exact", "links", url.Values{"target": {"urn:partner:upper"}}, []string{}},
		{"literal prefix source", "links", url.Values{"source": {"links/a"}}, []string{}},
		{"literal prefix target", "links", url.Values{"target": {"alias/a"}}, []string{}},
		{"endpoint OR", "links", url.Values{"endpoint": {selectionScope + "beads/z"}}, []string{"links/a", "links/self", "links/z"}},
		{"selector stored pin", "links", url.Values{"source": {"beads/z"}, "selector": {`$[?@.source == "https://example.test/read/beads/z"]`}}, []string{"links/self"}},
		{"selector pin URI", "links", url.Values{"selector": {`$[?@.source.uri == "https://example.test/read/beads/z"]`}}, []string{"links/z"}},
		{"selector literal not normalized", "links", url.Values{"selector": {`$[?@.target == "beads/z"]`}}, []string{}},
		{"limit does not truncate selection", "beads", url.Values{"limit": {"1"}}, []string{"beads/a", "beads/m", "beads/z"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			query, err := CompileCollection(selectionScope, tc.collection, tc.query, selectionLimits)
			if err != nil {
				t.Fatal(err)
			}
			inventory := selectionFixture()
			before, _ := json.Marshal(inventory)
			rows, err := query.Select(inventory)
			if err != nil {
				t.Fatal(err)
			}
			want := make([]string, 0, len(tc.ids))
			for _, id := range tc.ids {
				want = append(want, selectionScope+id)
			}
			if got := selectedIDs(t, rows); !reflect.DeepEqual(got, want) {
				t.Fatalf("got %v, want %v", got, want)
			}
			after, _ := json.Marshal(inventory)
			if string(before) != string(after) {
				t.Fatal("selection mutated inventory")
			}
		})
	}
}

func TestCollectionInvalidQueries(t *testing.T) {
	for _, tc := range []struct{ collection, query string }{
		{"beads", "view=properties"}, {"beads", "source=beads/a"}, {"links", "include=links"},
		{"types", "type=https://example.test/types/a"}, {"types", "selector=x"},
		{"beads", "limit=1&limit=1"}, {"links", "cursor=a&cursor=b"}, {"beads", "type="},
		{"beads", "limit="}, {"beads", "limit=0"}, {"beads", "limit=01"}, {"beads", "limit=-1"},
		{"beads", "limit=1.0"}, {"beads", "limit=9007199254740992"}, {"beads", "cursor="},
		{"beads", "type=types/a"}, {"beads", "conformsTo=urn:type:a"},
		{"links", "source=links//a"}, {"links", "target=alias//a"},
		{"links", "endpoint=https://EXAMPLE.test/read/beads/a"},
		{"links", "endpoint=https://example.test/read/beads/%2561"},
		{"types", "cursor=abc&unknown=x"}, {"unknown", ""},
	} {
		t.Run(tc.collection+"/"+tc.query, func(t *testing.T) {
			values, err := url.ParseQuery(tc.query)
			if err != nil {
				t.Fatal(err)
			}
			_, err = CompileCollection(selectionScope, tc.collection, values, selectionLimits)
			var invalid *ParameterError
			if !errors.As(err, &invalid) {
				t.Fatalf("not invalid-parameter: %v", err)
			}
		})
	}
	_, err := CompileCollection(selectionScope, "beads", url.Values{"selector": {"$..*"}}, selectionLimits)
	var selectorErr *SelectorError
	if !errors.As(err, &selectorErr) {
		t.Fatalf("Selector error class lost: %v", err)
	}
}

func TestCollectionContinuationCannotRestart(t *testing.T) {
	values := url.Values{"limit": {"1"}, "type": {selectionScope + "types/child"}}
	first, err := CompileCollection(selectionScope, "beads", values, selectionLimits)
	if err != nil {
		t.Fatal(err)
	}
	values.Set("cursor", "opaque")
	next, err := CompileCollection(selectionScope, "beads", values, selectionLimits)
	if err != nil {
		t.Fatal(err)
	}
	if next.Limit() != 1 || next.Cursor() != "opaque" || next.Projection() != first.Projection() || next.ContinuationURL() != first.ContinuationURL() {
		t.Fatal("continuation context changed")
	}
	if rows, err := next.Select(selectionFixture()); err == nil || rows != nil {
		t.Fatal("cursor restarted against current inventory")
	}
	values.Set("type", "https://changed.test/type")
	if next.ContinuationURL() != first.ContinuationURL() {
		t.Fatal("compiled request changed through caller's parameters")
	}
}

func TestCollectionTypeSummariesAndIntegrity(t *testing.T) {
	q, err := CompileCollection(selectionScope, "types", nil, selectionLimits)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := q.Select(selectionFixture())
	if err != nil || len(rows) != 3 {
		t.Fatalf("Type summaries: %d %v", len(rows), err)
	}
	for _, row := range rows {
		var summary bdpwire.TypeSummary
		if err := bdpwire.Unmarshal(row, &summary); err != nil {
			t.Fatal(err)
		}
	}
	broken := selectionFixture()
	broken.Types = broken.Types[:1]
	if rows, err := q.Select(broken); !errors.Is(err, graphstore.ErrInvalidStore) || rows != nil {
		t.Fatal("incomplete ancestry produced a partial answer")
	}
}

func TestIncidentSelection(t *testing.T) {
	for _, tc := range []struct {
		direction string
		count     int
	}{{"both", 3}, {"inbound", 2}, {"outbound", 2}} {
		q, err := CompileIncident(selectionScope, "beads/z", url.Values{"direction": {tc.direction}}, selectionLimits)
		if err != nil {
			t.Fatal(err)
		}
		rows, err := q.Select(selectionFixture())
		if err != nil || len(rows) != tc.count {
			t.Fatalf("%s: %d %v", tc.direction, len(rows), err)
		}
	}
	missing, err := CompileIncident(selectionScope, "beads/missing", nil, selectionLimits)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := missing.Select(selectionFixture()); !errors.Is(err, graphstore.ErrNotFound) {
		t.Fatalf("missing Bead: %v", err)
	}
	for _, values := range []url.Values{{"direction": {"out"}}, {"type": {selectionScope + "types/link"}}, {"view": {"properties"}}} {
		if _, err := CompileIncident(selectionScope, "beads/z", values, selectionLimits); err == nil {
			t.Fatal("invalid incident parameters admitted")
		}
	}
}

// These are selection fixtures, not evidence of native installation/admission.
// Current child v2 has different ancestry from the exact v1 on the Task.
func resourceSelectionFixture(t *testing.T) graphstore.ResourceSnapshot {
	t.Helper()
	ref := func(name, revision string) resourceget.TypeReference {
		return resourceget.TypeReference{URI: selectionScope + name, Revision: revision}
	}
	meta, parent, oldChild, child, relation := ref("bead-type", "m1"), ref("types/base", "p1"), ref("types/task", "c1"), ref("types/task", "c2"), ref("types/refines", "r1")
	fact := func(reference resourceget.TypeReference, kind graph.ResourceKind, current bool, ancestors ...resourceget.TypeReference) graphstore.ResourceTypeFact {
		return graphstore.ResourceTypeFact{Reference: reference, Definition: reference, Describes: kind, Current: current, Ancestors: append([]resourceget.TypeReference{reference}, ancestors...)}
	}
	result := graphstore.ResourceSnapshot{Types: []graphstore.ResourceTypeFact{
		fact(meta, graph.KindBead, true), fact(parent, graph.KindBead, true),
		fact(oldChild, graph.KindBead, false, parent), fact(child, graph.KindBead, true), fact(relation, graph.KindLink, true),
	}}
	for _, row := range []struct {
		id         string
		kind       graph.ResourceKind
		typ        resourceget.TypeReference
		properties map[string]any
	}{
		{"type", graph.KindBead, meta, map[string]any{"name": "type"}},
		{"bead-type", graph.KindBead, meta, map[string]any{"name": "bead-type"}},
		{"333", graph.KindBead, oldChild, map[string]any{"title": "Hello world", "status": "open", "metadata": map[string]any{"type": "unchanged"}}},
		{"edge", graph.KindLink, relation, map[string]any{}},
	} {
		value := map[string]any{"id": selectionScope + row.id, "revision": "record-1", "type": row.typ, "properties": row.properties, "metadata": map[string]any{"retained": true}}
		if row.kind == graph.KindBead {
			value["ownedLinks"] = map[string]any{}
		} else {
			value["source"] = resourceget.TypeReference{URI: selectionScope + "333", Revision: "task-old"}
			value["target"] = "urn:partner:UPPER"
		}
		raw, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		document, err := resourceget.NewResourceDocument(selectionScope, row.kind, raw)
		if err != nil {
			t.Fatal(err)
		}
		result.Documents = append(result.Documents, document)
	}
	return result
}

func TestResourceCollectionUsesOrdinaryBeadsAndExactCapturedAncestry(t *testing.T) {
	inventory := resourceSelectionFixture(t)
	for _, tc := range []struct {
		name, collection string
		query            url.Values
		ids              []string
	}{
		{"ordinary definitions and instance", "beads", nil, []string{"333", "bead-type", "type"}},
		{"metatype selector", "beads", url.Values{"conformsTo": {selectionScope + "bead-type"}}, []string{"bead-type", "type"}},
		{"retained parent", "beads", url.Values{"conformsTo": {selectionScope + "types/base"}}, []string{"333"}},
		{"nominal across editions", "beads", url.Values{"type": {selectionScope + "types/task"}}, []string{"333"}},
		{"unknown nominal URI is nonmatch", "beads", url.Values{"type": {"https://unknown.test/type"}}, []string{}},
		{"query exact stored pin", "beads", url.Values{"selector": {`$[?@.type.revision == "c1" && @.properties.status == "open"]`}}, []string{"333"}},
		{"endpoint and stored pin", "links", url.Values{"source": {selectionScope + "333"}, "target": {"urn:partner:UPPER"}, "selector": {`$[?@.source.revision == "task-old"]`}}, []string{"edge"}},
		{"case-sensitive foreign endpoint", "links", url.Values{"target": {"urn:partner:upper"}}, []string{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			query, err := CompileCollection(selectionScope, tc.collection, tc.query, selectionLimits)
			if err != nil {
				t.Fatal(err)
			}
			rows, err := query.SelectResources(inventory)
			if err != nil {
				t.Fatal(err)
			}
			want := make([]string, 0, len(tc.ids))
			for _, id := range tc.ids {
				want = append(want, selectionScope+id)
			}
			if got := selectedIDs(t, rows); !reflect.DeepEqual(got, want) {
				t.Fatalf("got %v want %v", got, want)
			}
			for _, raw := range rows {
				matched := false
				for _, document := range inventory.Documents {
					if string(raw) == string(document.NativeJSON()) {
						matched = true
					}
				}
				if !matched {
					t.Fatal("selection rewrote or reconstructed a complete record")
				}
			}
		})
	}
}

func TestResourceCollectionRejectsIncompleteCaptureWithoutPartialResults(t *testing.T) {
	query, err := CompileCollection(selectionScope, "beads", nil, selectionLimits)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		mutate func(*graphstore.ResourceSnapshot)
	}{
		{"duplicate", func(s *graphstore.ResourceSnapshot) { s.Documents = append(s.Documents, s.Documents[0]) }},
		{"missing exact affiliation", func(s *graphstore.ResourceSnapshot) { s.Types = append(s.Types[:2:2], s.Types[3:]...) }},
		{"wrong kind", func(s *graphstore.ResourceSnapshot) { s.Types[0].Describes = graph.KindLink }},
		{"zero document", func(s *graphstore.ResourceSnapshot) {
			s.Documents = append(s.Documents, resourceget.ResourceDocument{})
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			snapshot := resourceSelectionFixture(t)
			tc.mutate(&snapshot)
			rows, err := query.SelectResources(snapshot)
			if !errors.Is(err, graphstore.ErrInvalidStore) || rows != nil {
				t.Fatalf("partial result %s, error %v", rows, err)
			}
		})
	}
}

func TestResourceNominalSelectionSeparatesRawURIFromProductLookup(t *testing.T) {
	snapshot := resourceSelectionFixture(t)
	local := snapshot.Types[0].Reference
	semantic := resourceget.TypeReference{URI: "https://protocol.example/bead-type", Revision: "global-2"}
	snapshot.Types[0].Reference = semantic
	snapshot.Types[0].Ancestors = []resourceget.TypeReference{semantic}
	// Definition remains the independently checked local record edition. Only
	// this fixture's correspondence evidence bridges its two distinct IDs.
	for i, document := range snapshot.Documents {
		if document.Type() != local {
			continue
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(document.NativeJSON(), &fields); err != nil {
			t.Fatal(err)
		}
		rawType, err := json.Marshal(semantic)
		if err != nil {
			t.Fatal(err)
		}
		fields["type"] = rawType
		raw, err := json.Marshal(fields)
		if err != nil {
			t.Fatal(err)
		}
		snapshot.Documents[i], err = resourceget.NewResourceDocument(selectionScope, document.Kind(), raw)
		if err != nil {
			t.Fatal(err)
		}
	}
	q, err := CompileCollection(selectionScope, "beads", url.Values{"type": {local.URI}}, selectionLimits)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := q.SelectResources(snapshot)
	if err != nil || len(rows) != 0 {
		t.Fatalf("raw nominal URI was rewritten: %s %v", rows, err)
	}
	resolved, err := q.ResolveInstalledTypeSelectors(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	rows, err = resolved.SelectResources(snapshot)
	if err != nil || len(rows) != 2 {
		t.Fatalf("explicit checked product lookup lost definitions: %s %v", rows, err)
	}
	again, err := q.SelectResources(snapshot)
	if err != nil || len(again) != 0 {
		t.Fatal("product lookup mutated raw selection")
	}
	for _, tc := range []struct {
		name, selector string
		product        bool
		want           int
	}{
		{"raw local family is not an alias", local.URI, false, 0},
		{"raw semantic family", semantic.URI, false, 2},
		{"explicit product family lookup", local.URI, true, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			family, err := CompileCollection(selectionScope, "beads", url.Values{"conformsTo": {tc.selector}}, selectionLimits)
			if err != nil {
				t.Fatal(err)
			}
			selected := family
			if tc.product {
				selected, err = family.ResolveInstalledTypeSelectors(snapshot)
				if err != nil {
					t.Fatal(err)
				}
			}
			rows, err := selected.SelectResources(snapshot)
			if err != nil || len(rows) != tc.want {
				t.Fatalf("family selection: got %d rows, want %d, error %v", len(rows), tc.want, err)
			}
			if tc.product {
				rows, err = family.SelectResources(snapshot)
				if err != nil || len(rows) != 0 {
					t.Fatal("product family lookup mutated raw selection")
				}
			}
		})
	}
}
