package graphstore

import (
	"bytes"
	"encoding/json"
	"testing"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/resourceget"
)

// This controls immutable applied-result representation, not native transaction
// admission. Only the actual checked writer may populate the private facet.
func TestAppliedResourceFacetKeepsLegacyJSONAndExactAppliedEdition(t *testing.T) {
	raw := []byte(`{"id":"https://scope.example/333","ownedLinks":{},"properties":{},"revision":"applied","type":{"revision":"definition-inception","uri":"https://scope.example/native-issue/task"}}`)
	document, err := resourceget.NewResourceDocument("https://scope.example/", graph.KindBead, raw)
	if err != nil {
		t.Fatal(err)
	}
	record := IssueRecord{ID: document.ID(), Type: document.Type().URI, Revision: document.Revision()}
	before, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	if _, present := record.ResourceDocument(); present {
		t.Fatal("legacy result fabricated exact affiliation evidence")
	}
	record.resourceDocument = &document
	selected, present := record.ResourceDocument()
	after, err := json.Marshal(record)
	if err != nil || !bytes.Equal(before, after) || !present || selected.Type().Revision != "definition-inception" || !bytes.Equal(selected.NativeJSON(), raw) {
		t.Fatalf("applied facet changed legacy JSON or exact edition: %s, %v", after, err)
	}
	returned := selected.NativeJSON()
	returned[0] = '!'
	selected, present = record.ResourceDocument()
	if !present || !bytes.Equal(selected.NativeJSON(), raw) {
		t.Fatal("caller changed applied result evidence")
	}
	record.Revision = "later-head"
	if _, present := record.ResourceDocument(); present {
		t.Fatal("applied facet silently selected a later revision")
	}
	memory := Record{ID: document.ID(), Type: document.Type().URI, Revision: document.Revision(), resourceDocument: &document}
	if _, present := memory.ResourceDocument(); !present {
		t.Fatal("Memory result lost same immutable facet")
	}
	link := LinkRecord{ID: document.ID(), Type: document.Type().URI, Revision: document.Revision(), resourceDocument: &document}
	if _, present := link.ResourceDocument(); present {
		t.Fatal("Bead facet accepted as Link result")
	}
}
