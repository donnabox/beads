package graphread

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/resourcequery"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/resourceget"
)

func querySnapshotDocument(t *testing.T, id, typePin, text string) resourceget.ResourceDocument {
	t.Helper()
	raw, err := json.Marshal(map[string]any{
		"id": "https://query-mapper.example/" + id, "revision": "resource-r1",
		"type":       map[string]any{"uri": "https://query-mapper.example/T", "revision": typePin},
		"properties": map[string]any{"text": text}, "metadata": map[string]any{"kept": true},
	})
	if err != nil {
		t.Fatal(err)
	}
	document, err := resourceget.NewResourceDocument("https://query-mapper.example/", graph.KindBead, raw)
	if err != nil {
		t.Fatal(err)
	}
	return document
}

func TestResourceQueryMapperRetainsCapturedExactAuthority(t *testing.T) {
	const base = "https://query-mapper.example/"
	ref := func(name, pin string) resourceget.TypeReference {
		return resourceget.TypeReference{URI: base + name, Revision: pin}
	}
	oldT, newT, oldA, newA, b := ref("T", "old"), ref("T", "new"), ref("A", "old-a"), ref("A", "new-a"), ref("B", "b")
	snapshot := graphstore.ResourceSnapshot{Documents: []resourceget.ResourceDocument{querySnapshotDocument(t, "Old", "old", "kept"), querySnapshotDocument(t, "New", "new", "kept")}, Types: []graphstore.ResourceTypeFact{
		{Reference: oldT, Definition: ref("local-T", "local-old"), Describes: graph.KindBead, Ancestors: []resourceget.TypeReference{oldT, oldA}},
		{Reference: newT, Definition: ref("local-T", "local-new"), Describes: graph.KindBead, Current: true, Ancestors: []resourceget.TypeReference{newT, b}},
		{Reference: oldA, Definition: oldA, Describes: graph.KindBead, Ancestors: []resourceget.TypeReference{oldA}},
		{Reference: newA, Definition: newA, Describes: graph.KindBead, Current: true, Ancestors: []resourceget.TypeReference{newA}},
		{Reference: b, Definition: b, Describes: graph.KindBead, Current: true, Ancestors: []resourceget.TypeReference{b}},
	}}
	data, err := resourceQueryDataset(context.Background(), snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if data.TypeEditions == nil || data.TypeFamilies != nil || data.TypeKinds != nil || len(data.Resources) != 2 {
		t.Fatalf("wrong authority boundary: %+v", data)
	}
	for i, document := range snapshot.Documents {
		if !bytes.Equal(document.NativeJSON(), data.Resources[i].Data) {
			t.Fatal("mapped record lost captured full bytes")
		}
	}
	if data.TypeEditions[0].Reference.ID != oldT.URI || data.TypeEditions[0].Reference.Revision != oldT.Revision || data.TypeEditions[0].Current || data.TypeEditions[0].Ancestors[1].Revision != oldA.Revision {
		t.Fatal("retained Type facts changed during mapping")
	}
	// Mutating exported native input facts after mapping cannot retarget the
	// evaluator's already detached context to a current or local definition.
	snapshot.Types[0].Ancestors[1] = newA
	snapshot.Types[0].Reference = ref("local-T", "local-old")
	raw := []byte(`{"ir":"bdp-query-design-1","profile":"query","selection":{"op":"filter","input":{"op":"scan","collection":"beads"},"bind":"b","predicate":{"op":"typeIs","of":{"op":"binding","name":"b"},"type":"https://query-mapper.example/A","match":"family"}},"result":{"form":"resources"}}`)
	response, err := resourcequery.Execute(context.Background(), raw, data, resourcequery.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	result, ok := response.Result.(resourcequery.ResourcesResult)
	if !ok || !response.Evaluation.Complete || len(result.Items) != 1 || !bytes.Equal(result.Items[0], data.Resources[0].Data) {
		t.Fatalf("captured old ancestry replaced: %+v", response)
	}
}

func TestResourceQueryMapperRefusesWithoutPartialInput(t *testing.T) {
	empty := graphstore.ResourceSnapshot{Types: []graphstore.ResourceTypeFact{}}
	data, err := resourceQueryDataset(context.Background(), empty)
	if err != nil || data.TypeEditions == nil || len(data.TypeEditions) != 0 {
		t.Fatalf("empty exact authority lost: %+v %v", data, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for _, tc := range []struct {
		name     string
		ctx      context.Context
		snapshot graphstore.ResourceSnapshot
		want     error
	}{
		{"canceled", ctx, empty, context.Canceled},
		{"missing-authority", context.Background(), graphstore.ResourceSnapshot{}, resourcequery.ErrDataset},
		{"invalid-second-document", context.Background(), graphstore.ResourceSnapshot{Types: []graphstore.ResourceTypeFact{}, Documents: []resourceget.ResourceDocument{querySnapshotDocument(t, "First", "old", "one"), {}}}, resourcequery.ErrDataset},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, err := resourceQueryDataset(tc.ctx, tc.snapshot)
			if !errors.Is(err, tc.want) || data.Resources != nil || data.TypeEditions != nil {
				t.Fatalf("partial evaluator input escaped: %+v %v", data, err)
			}
		})
	}
}

func TestResourceQueryMapperEnforcesActualMappedBytes(t *testing.T) {
	// Each complete record is below the aggregate limit. Together they exceed
	// it; a fixed count or per-record allowance must not authorize the pair.
	text := strings.Repeat("x", QueryResourceByteLimit/2)
	snapshot := graphstore.ResourceSnapshot{Types: []graphstore.ResourceTypeFact{}, Documents: []resourceget.ResourceDocument{
		querySnapshotDocument(t, "One", "old", text), querySnapshotDocument(t, "Two", "old", text),
	}}
	data, err := resourceQueryDataset(context.Background(), snapshot)
	if !errors.Is(err, resourcequery.ErrLimit) || data.Resources != nil || data.TypeEditions != nil {
		t.Fatalf("aggregate mapped bytes escaped: %+v %v", data, err)
	}
}

// A representation-only mapper accepts test/user-constructible immutable input,
// so it must never manufacture exact CURRENT acquisition authority. Only the
// owning Reader bridge may assert it after a completed native checked read.
func TestResourceQueryMapperCannotAssertCurrentAcquisition(t *testing.T) {
	const base = "https://query-mapper.example/"
	linkRaw := []byte(`{"id":"https://query-mapper.example/L","revision":"link-r1","type":{"uri":"https://query-mapper.example/LT","revision":"lt-1"},"source":{"uri":"https://query-mapper.example/A","revision":"resource-r1"},"target":{"uri":"https://query-mapper.example/A","revision":"resource-r1"},"properties":{"note":"keep"},"metadata":{"kept":true}}`)
	link, err := resourceget.NewResourceDocument(base, graph.KindLink, linkRaw)
	if err != nil {
		t.Fatal(err)
	}
	snapshot := graphstore.ResourceSnapshot{WriterToken: "resource-r1", Types: []graphstore.ResourceTypeFact{}, Documents: []resourceget.ResourceDocument{querySnapshotDocument(t, "A", "old", "body"), link}}
	data, err := resourceQueryDataset(context.Background(), snapshot)
	if err != nil || data.CurrentEditions != nil {
		t.Fatalf("representation inferred provenance: %+v %v", data, err)
	}
	// The request has NO static pin. Actual acquisition of the stored target
	// pin reaches the shared dynamic resolver; matching body/WriterToken does
	// not confer authority. A static pinned request would refuse earlier in
	// Decode, irrespective of Dataset, and would not test this boundary.
	raw := []byte(`{"ir":"bdp-query-design-1","profile":"query","selection":{"op":"endpoints","input":{"op":"refs","collection":"links","references":["https://query-mapper.example/L"]},"side":"target"},"result":{"form":"resources"}}`)
	response, err := resourcequery.Execute(context.Background(), raw, data, resourcequery.DefaultLimits())
	if !errors.Is(err, resourcequery.ErrUnsupported) || response.Result != nil {
		t.Fatalf("matching document/WriterToken authorized dynamic exact acquisition: %+v %v", response, err)
	}
	var reader *Reader
	captured, escaped, err := reader.resourceQueryCapture(context.Background())
	if !errors.Is(err, resourcequery.ErrDataset) || captured.Documents != nil || escaped.Resources != nil || escaped.TypeEditions != nil || escaped.CurrentEditions != nil {
		t.Fatalf("absent native reader returned partial authority: %+v %+v %v", captured, escaped, err)
	}
}
