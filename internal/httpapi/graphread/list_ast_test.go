package graphread

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/httpapi/bdpwire"
	"github.com/steveyegge/beads/internal/resourcequery"
)

func publicListAST(collection, predicate string) []byte {
	selection := fmt.Sprintf(`{"op":"scan","collection":%q}`, collection)
	if predicate != "" {
		selection = fmt.Sprintf(`{"op":"filter","input":%s,"bind":"item","predicate":%s}`, selection, predicate)
	}
	return []byte(fmt.Sprintf(`{"ir":"bdp-query-design-1","profile":"list","selection":%s,"result":{"form":"resources"}}`, selection))
}

func TestPublicListAdmissionBeforeAcquisition(t *testing.T) {
	var reader *Reader
	for _, tc := range []struct {
		name string
		raw  []byte
		kind graph.ResourceKind
		want error
	}{
		{"syntax", []byte(`{`), graph.KindBead, resourcequery.ErrInvalid},
		{"wrong kind", publicListAST("links", ""), graph.KindBead, resourcequery.ErrInvalid},
		{"query profile", []byte(`{"ir":"bdp-query-design-1","profile":"query","selection":{"op":"scan","collection":"beads"},"result":{"form":"resources"}}`), graph.KindBead, resourcequery.ErrInvalid},
		{"aggregate", []byte(`{"ir":"bdp-query-design-1","profile":"list","selection":{"op":"scan","collection":"beads"},"result":{"form":"count"}}`), graph.KindBead, resourcequery.ErrUnsupported},
		{"traversing predicate", publicListAST("links", `{"op":"present","value":{"op":"field","of":{"op":"endpoint","of":{"op":"binding","name":"item"},"side":"source"},"path":["properties","status"]}}`), graph.KindLink, resourcequery.ErrInvalid},
	} {
		t.Run(tc.name, func(t *testing.T) {
			inventory, err := reader.InventoryMatchingList(context.Background(), tc.raw, tc.kind)
			if !errors.Is(err, tc.want) || inventory.Beads != nil || inventory.Links != nil {
				t.Fatalf("request reached storage or returned partial inventory: %+v %v", inventory, err)
			}
		})
	}
}

func TestPublicListRepresentationBoundaryBeforeAcquisition(t *testing.T) {
	var reader *Reader
	present := func(root string) string {
		return fmt.Sprintf(`{"op":"present","value":{"op":"field","of":{"op":"binding","name":"item"},"path":[%q]}}`, root)
	}
	for _, root := range []string{"revisionProvenance", "attribution", "changeContext", "ownedLinks"} {
		for _, collection := range []string{"beads", "links"} {
			kind := graph.KindBead
			if collection == "links" {
				kind = graph.KindLink
			}
			t.Run(collection+"/"+root, func(t *testing.T) {
				// A true first operand must not hide an unsupported field access.
				predicate := `{"op":"or","operands":[true,` + present(root) + `]}`
				got, err := reader.InventoryMatchingList(context.Background(), publicListAST(collection, predicate), kind)
				if !errors.Is(err, resourcequery.ErrUnsupported) || got.Beads != nil || got.Links != nil {
					t.Fatalf("representation mismatch reached acquisition: %+v %v", got, err)
				}
			})
		}
	}
	for _, path := range []string{`["properties","ownedLinks"]`, `["metadata","revisionProvenance"]`, `["unknownRoot"]`} {
		predicate := `{"op":"present","value":{"op":"field","of":{"op":"binding","name":"item"},"path":` + path + `}}`
		if err := AdmitResourceList(publicListAST("beads", predicate), graph.KindBead); err != nil {
			t.Fatalf("ordinary literal property or missing root refused: %s %v", path, err)
		}
	}
}

func publicListInventory(t *testing.T) (Inventory, resourcequery.Dataset) {
	t.Helper()
	a, b := queryRepresentationBead(), queryRepresentationBead()
	a.Metadata["marker"] = json.RawMessage(`null`)
	b.ID, b.Revision, b.Metadata = queryRepresentationBase+"B", "b7", bdpwire.Metadata{}
	b.Properties = bdpwire.Properties{"body": json.RawMessage(`"other"`)}
	link := queryRepresentationLink()
	inventory := Inventory{WriterToken: "captured", Beads: []bdpwire.BeadRecord{a, b}, Links: []bdpwire.LinkRecord{link}}
	data := resourcequery.Dataset{
		TypeKinds:    map[string]graph.ResourceKind{a.Type: graph.KindBead, link.Type: graph.KindLink},
		TypeFamilies: map[string][]string{a.Type: {a.Type}, link.Type: {link.Type}},
	}
	remaining := QueryResourceByteLimit
	for _, value := range []struct {
		kind   graph.ResourceKind
		record any
	}{{graph.KindBead, a}, {graph.KindBead, b}, {graph.KindLink, link}} {
		r, err := queryResourceWithinBudget(value.kind, value.record, &remaining)
		if err != nil {
			t.Fatal(err)
		}
		data.Resources = append(data.Resources, r)
	}
	return inventory, data
}

func TestLegacyListFieldsPreserveTextAndLiteralKeys(t *testing.T) {
	records := []json.RawMessage{
		json.RawMessage(`{"properties":{"status":"active"},"metadata":{"n":1,"marker":null,"gc.routed_to":"local","payload":{"a":1}}}`),
		json.RawMessage(`{"properties":{"status":"active"},"metadata":{"n":"1","marker":"","gc":{"routed_to":"local"},"payload":{"a":1}}}`),
		json.RawMessage(`{"properties":{"status":"draft"},"metadata":{"n":1.0,"marker":false,"gc.routed_to":"other"}}`),
		json.RawMessage(`{"properties":{},"metadata":{}}`),
	}
	for _, tc := range []struct {
		name   string
		fields []LegacyListField
		want   []int
	}{
		{"string and number", []LegacyListField{{Namespace: "metadata", Key: "n", Value: "1"}}, []int{0, 1, 2}},
		{"numeric normalization", []LegacyListField{{Namespace: "metadata", Key: "n", Value: "1.0"}}, nil},
		{"explicit null is present", []LegacyListField{{Namespace: "metadata", Key: "marker", Presence: true}}, []int{0, 1, 2}},
		{"null is not empty string", []LegacyListField{{Namespace: "metadata", Key: "marker", Value: ""}}, []int{1}},
		{"null text", []LegacyListField{{Namespace: "metadata", Key: "marker", Value: "null"}}, []int{0}},
		{"boolean", []LegacyListField{{Namespace: "metadata", Key: "marker", Value: "false"}}, []int{2}},
		{"literal dotted key", []LegacyListField{{Namespace: "metadata", Key: "gc.routed_to", Value: "local"}}, []int{0}},
		{"namespace separation", []LegacyListField{{Namespace: "metadata", Key: "status", Value: "active"}}, nil},
		{"AND across namespaces", []LegacyListField{{Namespace: "properties", Key: "status", Value: "active"}, {Namespace: "metadata", Key: "n", Value: "1"}}, []int{0, 1}},
		{"conflicting repeat", []LegacyListField{{Namespace: "properties", Key: "status", Value: "active"}, {Namespace: "properties", Key: "status", Value: "draft"}}, nil},
		{"SQL container spelling", []LegacyListField{{Namespace: "metadata", Key: "payload", Value: `{"a": 1}`}}, []int{0, 1}},
		{"compact container spelling is different", []LegacyListField{{Namespace: "metadata", Key: "payload", Value: `{"a":1}`}}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			filter, err := CompileLegacyListFields(tc.fields)
			if err != nil {
				t.Fatal(err)
			}
			got, err := filter.Select(context.Background(), records)
			if err != nil || len(got) != len(tc.want) {
				t.Fatalf("selection = %s, %v; want indexes %v", got, err, tc.want)
			}
			for i, index := range tc.want {
				if string(got[i]) != string(records[index]) {
					t.Fatalf("selected record %d: %s", i, got[i])
				}
			}
		})
	}
}

func TestLegacyListFieldsAdmissionAndCancellation(t *testing.T) {
	for _, fields := range [][]LegacyListField{
		{{Namespace: "source", Key: "status", Value: "active"}},
		{{Namespace: "metadata", Key: ""}},
		make([]LegacyListField, 257),
		{{Namespace: "metadata", Key: "key", Value: strings.Repeat("x", 16385)}},
	} {
		if _, err := CompileLegacyListFields(fields); err == nil {
			t.Fatal("invalid or excessive field conditions admitted")
		}
	}
	fields := []LegacyListField{{Namespace: "metadata", Key: "marker", Presence: true}}
	filter, err := CompileLegacyListFields(fields)
	if err != nil {
		t.Fatal(err)
	}
	fields[0].Namespace = "source"
	records := []json.RawMessage{json.RawMessage(`{"metadata":{"marker":null}}`)}
	if got, err := filter.Select(context.Background(), records); err != nil || len(got) != 1 {
		t.Fatalf("caller mutated compiled fields: %s %v", got, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if got, err := filter.Select(ctx, records); !errors.Is(err, context.Canceled) || got != nil {
		t.Fatalf("canceled selection returned data: %s %v", got, err)
	}
	if got, err := filter.Select(context.Background(), []json.RawMessage{json.RawMessage(`{`)}); !errors.Is(err, resourcequery.ErrDataset) || got != nil {
		t.Fatalf("corrupt candidate returned data: %s %v", got, err)
	}
}

func TestLegacyListSQLTextRendering(t *testing.T) {
	for _, tc := range []struct{ raw, want string }{
		{`null`, "null"},
		{`"null"`, "null"},
		{`""`, ""},
		{`true`, "true"},
		{`1.0`, "1"},
		{`1e21`, "1000000000000000000000"},
		{`1e-7`, "0.0000001"},
		{`[1,2]`, "[1, 2]"},
		{`{"aa":1,"b":2}`, `{"b": 2, "aa": 1}`},
		{`{"key":["a\nb","a\\b","a\"b",null]}`, "{\"key\": [\"a\nb\", \"a\\b\", \"a\"b\", null]}"},
		{`["<>&","\u0001","雪"]`, "[\"<>&\", \"\x01\", \"雪\"]"},
	} {
		budget := 65536
		got, err := legacyListSQLText(context.Background(), json.RawMessage(tc.raw), &budget)
		if err != nil || got != tc.want {
			t.Fatalf("SQL text for %s = %q,%v; want %q", tc.raw, got, err, tc.want)
		}
	}
	budget := 2
	if _, err := legacyListSQLText(context.Background(), json.RawMessage(`"abc"`), &budget); !errors.Is(err, resourcequery.ErrLimit) {
		t.Fatalf("comparison byte budget: %v", err)
	}
}

func TestPublicListMatchesFullEditionsAndRetainsEndpoints(t *testing.T) {
	inventory, data := publicListInventory(t)
	field := `{"op":"field","of":{"op":"binding","name":"item"},"path":["properties","body"]}`
	for _, tc := range []struct {
		name, collection, predicate string
		kind                        graph.ResourceKind
		beads, links                int
	}{
		{"bead local substring", "beads", `{"op":"contains","text":` + field + `,"substring":{"op":"literal","value":"rails"}}`, graph.KindBead, 1, 1},
		{"link local substring retains both endpoints", "links", `{"op":"contains","text":` + field + `,"substring":{"op":"literal","value":"evidence"}}`, graph.KindLink, 2, 1},
		{"empty links retains endpoints", "links", `false`, graph.KindLink, 2, 0},
		{"present distinguishes absent from null", "beads", `{"op":"present","value":{"op":"field","of":{"op":"binding","name":"item"},"path":["metadata","marker"]}}`, graph.KindBead, 1, 1},
		{"family authority", "beads", `{"op":"typeIs","of":{"op":"binding","name":"item"},"type":"https://query.example/Memory","match":"family"}`, graph.KindBead, 2, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw := publicListAST(tc.collection, tc.predicate)
			if err := AdmitResourceList(raw, tc.kind); err != nil {
				t.Fatal(err)
			}
			got, err := matchListInventory(context.Background(), raw, tc.kind, inventory, data)
			if err != nil || len(got.Beads) != tc.beads || len(got.Links) != tc.links || got.WriterToken != inventory.WriterToken {
				t.Fatalf("captured List selection: beads=%d links=%d err=%v", len(got.Beads), len(got.Links), err)
			}
			if len(inventory.Beads) != 2 || len(inventory.Links) != 1 {
				t.Fatal("input inventory mutated")
			}
			if len(got.Beads) > 0 && (got.Beads[0].Revision != "a3" || got.Beads[0].Attribution == nil) {
				t.Fatal("public edition/provenance lost")
			}
		})
	}
	t.Run("edition mismatch refuses all", func(t *testing.T) {
		bad := inventory
		bad.Beads = append([]bdpwire.BeadRecord(nil), inventory.Beads...)
		bad.Beads[0].Revision = "different"
		got, err := matchListInventory(context.Background(), publicListAST("beads", ""), graph.KindBead, bad, data)
		if !errors.Is(err, resourcequery.ErrDataset) || got.Beads != nil {
			t.Fatalf("mismatched edition accepted: %+v %v", got, err)
		}
	})
	t.Run("cancel returns no partial inventory", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		got, err := matchListInventory(ctx, publicListAST("beads", ""), graph.KindBead, inventory, data)
		if !errors.Is(err, context.Canceled) || got.Beads != nil {
			t.Fatalf("cancelled List accepted: %+v %v", got, err)
		}
	})
}
