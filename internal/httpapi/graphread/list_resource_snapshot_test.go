package graphread

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/resourcequery"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/resourceget"
)

func resourceListSnapshot(t *testing.T) graphstore.ResourceSnapshot {
	t.Helper()
	const scope = "https://query-mapper.example/"
	first := querySnapshotDocument(t, "First", "old", "ordinary")
	second := querySnapshotDocument(t, "Second", "old", "definition-shaped ordinary data")
	var body map[string]any
	if err := json.Unmarshal(second.NativeJSON(), &body); err != nil {
		t.Fatal(err)
	}
	body["properties"] = map[string]any{"describes": "bead", "propertiesSchema": map[string]any{"type": "object"}}
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	second, err = resourceget.NewResourceDocument(scope, graph.KindBead, raw)
	if err != nil {
		t.Fatal(err)
	}
	link, err := resourceget.NewResourceDocument(scope, graph.KindLink, []byte(`{"id":"https://query-mapper.example/L","revision":"link-r1","type":{"uri":"https://query-mapper.example/R","revision":"r"},"source":{"uri":"https://query-mapper.example/First","revision":"resource-r1"},"target":{"uri":"https://query-mapper.example/Second","revision":"resource-r1"},"properties":{"kept":true},"metadata":{"tag":"full-body"}}`))
	if err != nil {
		t.Fatal(err)
	}
	tRef := resourceget.TypeReference{URI: scope + "T", Revision: "old"}
	rRef := resourceget.TypeReference{URI: scope + "R", Revision: "r"}
	return graphstore.ResourceSnapshot{WriterToken: "same-capture", Documents: []resourceget.ResourceDocument{first, second, link}, Types: []graphstore.ResourceTypeFact{
		{Reference: tRef, Definition: tRef, Describes: graph.KindBead, Current: true, Ancestors: []resourceget.TypeReference{tRef}},
		{Reference: rRef, Definition: rRef, Describes: graph.KindLink, Current: true, Ancestors: []resourceget.TypeReference{rRef}},
	}}
}

func resourceListAST(collection string) []byte {
	return []byte(`{"ir":"bdp-query-design-1","profile":"list","selection":{"op":"scan","collection":"` + collection + `"},"result":{"form":"resources"}}`)
}

func TestResourceListSnapshotKeepsCompleteBodiesAndOtherKind(t *testing.T) {
	for _, tc := range []struct {
		collection string
		kind       graph.ResourceKind
	}{{"beads", graph.KindBead}, {"links", graph.KindLink}} {
		t.Run(tc.collection, func(t *testing.T) {
			snapshot := resourceListSnapshot(t)
			data, err := resourceQueryDataset(context.Background(), snapshot)
			if err != nil {
				t.Fatal(err)
			}
			raw := resourceListAST(tc.collection)
			if err = AdmitResourceList(raw, tc.kind); err != nil {
				t.Fatal(err)
			}
			result, err := matchResourceListSnapshot(context.Background(), raw, tc.kind, snapshot, data)
			if err != nil {
				t.Fatal(err)
			}
			if result.WriterToken != snapshot.WriterToken || len(result.Types) != 2 || len(result.Documents) != 3 {
				t.Fatalf("capture changed: %+v", result)
			}
			for i, original := range snapshot.Documents {
				if !bytes.Equal(result.Documents[i].NativeJSON(), original.NativeJSON()) {
					t.Fatal("List lost complete bytes, pins, endpoint or metadata")
				}
			}
		})
	}
}

func TestResourceListSnapshotRefusesSelectionOutsideCapture(t *testing.T) {
	snapshot := resourceListSnapshot(t)
	other := snapshot
	other.Documents = []resourceget.ResourceDocument{querySnapshotDocument(t, "Absent", "old", "not in original capture")}
	data, err := resourceQueryDataset(context.Background(), other)
	if err != nil {
		t.Fatal(err)
	}
	result, err := matchResourceListSnapshot(context.Background(), resourceListAST("beads"), graph.KindBead, snapshot, data)
	if !errors.Is(err, resourcequery.ErrDataset) || result.Documents != nil || result.Types != nil {
		t.Fatalf("unbound selection escaped: %+v %v", result, err)
	}
}

func TestResourceListSnapshotCancellationReturnsNoPartialCapture(t *testing.T) {
	snapshot := resourceListSnapshot(t)
	data, err := resourceQueryDataset(context.Background(), snapshot)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result, err := matchResourceListSnapshot(ctx, resourceListAST("beads"), graph.KindBead, snapshot, data)
	if !errors.Is(err, context.Canceled) || result.Documents != nil || result.Types != nil {
		t.Fatalf("partial canceled capture escaped: %+v %v", result, err)
	}
}
