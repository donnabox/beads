package graphread

import (
	"bytes"
	"context"
	"testing"

	graph "github.com/steveyegge/beads/graphops"
)

func TestResourceListSnapshotPredicateKeepsOnlyMatchingCompleteCandidate(t *testing.T) {
	snapshot := resourceListSnapshot(t)
	data, err := resourceQueryDataset(context.Background(), snapshot)
	if err != nil {
		t.Fatal(err)
	}
	// The selected record merely has descriptor-shaped properties. The same
	// generic predicate and result path must apply to every ordinary Bead.
	raw := []byte(`{"ir":"bdp-query-design-1","profile":"list","selection":{"op":"filter","input":{"op":"scan","collection":"beads"},"bind":"b","predicate":{"op":"compare","operator":"eq","left":{"op":"field","of":{"op":"binding","name":"b"},"path":["properties","describes"]},"right":{"op":"literal","value":"bead"}}},"result":{"form":"resources"}}`)
	if err := AdmitResourceList(raw, graph.KindBead); err != nil {
		t.Fatal(err)
	}
	selected, err := matchResourceListSnapshot(context.Background(), raw, graph.KindBead, snapshot, data)
	if err != nil {
		t.Fatal(err)
	}
	if len(selected.Documents) != 2 {
		t.Fatalf("wrong selected population: %d", len(selected.Documents))
	}
	for i, original := range snapshot.Documents[1:] {
		if !bytes.Equal(selected.Documents[i].NativeJSON(), original.NativeJSON()) {
			t.Fatal("predicate changed complete candidate or retained other-kind endpoint context")
		}
	}
	if len(snapshot.Documents) != 3 {
		t.Fatal("selection mutated the original capture")
	}
}
