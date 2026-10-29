package graphread

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/httpapi/bdpwire"
	"github.com/steveyegge/beads/internal/resourcequery"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

const queryRepresentationBase = "https://query.example/"

func TestQueryResourceBudgetBeforeRetention(t *testing.T) {
	bead := queryRepresentationBead()
	raw, err := queryResourceJSON(bead)
	if err != nil {
		t.Fatal(err)
	}
	remaining := len(raw) - 1
	if record, err := queryResourceWithinBudget(graph.KindBead, bead, &remaining); !errors.Is(err, resourcequery.ErrLimit) || record.Data != nil || remaining != len(raw)-1 {
		t.Fatalf("over-budget Resource retained or budget changed: %+v %v remaining=%d", record, err, remaining)
	}
	remaining = len(raw)
	if record, err := queryResourceWithinBudget(graph.KindBead, bead, &remaining); err != nil || len(record.Data) == 0 || remaining != 0 {
		t.Fatalf("exact-bound Resource refused: %+v %v remaining=%d", record, err, remaining)
	}
	if record, err := queryResourceWithinBudget(graph.KindBead, bead, &remaining); !errors.Is(err, resourcequery.ErrLimit) || record.Data != nil || remaining != 0 {
		t.Fatalf("cumulative Resource budget reset: %+v %v remaining=%d", record, err, remaining)
	}
}

func queryRepresentationLink() bdpwire.LinkRecord {
	return bdpwire.LinkRecord{ID: queryRepresentationBase + "L", Type: queryRepresentationBase + "Reference", Revision: "l1",
		Source: bdpwire.Reference{URI: queryRepresentationBase + "A"}, Target: bdpwire.Reference{URI: queryRepresentationBase + "B", Revision: "b7"},
		Properties: bdpwire.Properties{"body": json.RawMessage(`"evidence"`), "attribution": json.RawMessage(`{"domain":"preserve me"}`)},
		Metadata:   bdpwire.Metadata{"source": json.RawMessage(`"import"`)}, Attribution: &bdpwire.Attribution{Principal: "agent:Melinda", Basis: bdpwire.AttributionWriterSupplied}}
}
func queryRepresentationBead() bdpwire.BeadRecord {
	return bdpwire.BeadRecord{ID: queryRepresentationBase + "A", Type: queryRepresentationBase + "Memory", Revision: "a3",
		Properties: bdpwire.Properties{"body": json.RawMessage(`"rails"`), "custom": json.RawMessage(`{"values":[null,1,true]}`)},
		Metadata:   bdpwire.Metadata{"tag": json.RawMessage(`"retained"`)}, Attribution: &bdpwire.Attribution{Principal: "human:Donna", Basis: bdpwire.AttributionUnknown}}
}
func queryRepresentationObject(t *testing.T, raw []byte) map[string]any {
	t.Helper()
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatal(err)
	}
	return result
}
func queryRepresentationEqual(t *testing.T, got []byte, want any) {
	t.Helper()
	expected, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	equal, err := graph.JSONEqual(got, expected)
	if err != nil || !equal {
		t.Fatalf(`representation mismatch:
got %s
want %s
error %v`, got, expected, err)
	}
}

func TestQueryRepresentationFullBeadAndOwnedLink(t *testing.T) {
	link := queryRepresentationLink()
	bead := queryRepresentationBead()
	bead.OwnedLinks = bdpwire.OwnedLinks{link.Type: {link}, queryRepresentationBase + "Empty": nil}
	raw, err := queryResourceJSON(bead)
	if err != nil {
		t.Fatal(err)
	}
	expectedLink := map[string]any{"id": link.ID, "type": link.Type, "revision": link.Revision, "source": link.Source, "target": link.Target, "properties": link.Properties, "metadata": link.Metadata,
		"revisionProvenance": map[string]any{"actor": map[string]any{"id": "agent:Melinda", "basis": "writer-supplied"}}}
	queryRepresentationEqual(t, raw, map[string]any{"id": bead.ID, "type": bead.Type, "revision": bead.Revision, "properties": bead.Properties, "metadata": bead.Metadata,
		"ownedLinks":         map[string]any{link.Type: []any{expectedLink}, queryRepresentationBase + "Empty": []any{}},
		"revisionProvenance": map[string]any{"actor": map[string]any{"id": "human:Donna", "basis": "unknown"}}})
	raw, err = queryResourceJSON(link)
	if err != nil {
		t.Fatal(err)
	}
	queryRepresentationEqual(t, raw, expectedLink)
}

func TestQueryRepresentationContextPresenceStates(t *testing.T) {
	at, agent, empty := "2026-10-28T01:00:00Z", "agent:Evaluator", ""
	cases := []bdpwire.ChangeContext{
		{CommittedAt: bdpwire.ContextTime{State: bdpwire.ContextTimePresent, Value: &at}, Agent: bdpwire.ContextString{State: bdpwire.ContextPresent, Value: &agent}, Message: bdpwire.ContextMessage{State: bdpwire.ContextPresent, Value: &empty}},
		{CommittedAt: bdpwire.ContextTime{State: bdpwire.ContextTimeUndetermined}, Agent: bdpwire.ContextString{State: bdpwire.ContextAbsent}, Message: bdpwire.ContextMessage{State: bdpwire.ContextAbsent}},
		{CommittedAt: bdpwire.ContextTime{State: bdpwire.ContextTimeUndetermined}, Agent: bdpwire.ContextString{State: bdpwire.ContextUndetermined}, Message: bdpwire.ContextMessage{State: bdpwire.ContextUndetermined}},
	}
	for _, change := range cases {
		bead := queryRepresentationBead()
		bead.Attribution = nil
		bead.ChangeContext = &change
		raw, err := queryResourceJSON(bead)
		if err != nil {
			t.Fatal(err)
		}
		body := queryRepresentationObject(t, raw)
		provenance := body["revisionProvenance"].(map[string]any)
		encoded, err := json.Marshal(provenance)
		if err != nil {
			t.Fatal(err)
		}
		queryRepresentationEqual(t, encoded, map[string]any{"committedAt": change.CommittedAt, "agent": change.Agent, "message": change.Message})
		if _, present := provenance["actor"]; present {
			t.Fatal("invented actor")
		}
		if body["changeContext"] != nil || body["attribution"] != nil {
			t.Fatal("legacy root fields leaked")
		}
		// Link and owned-Link mapping must use the identical context conversion.
		link := queryRepresentationLink()
		link.ChangeContext = &change
		raw, err = queryResourceJSON(link)
		if err != nil {
			t.Fatal(err)
		}
		p := queryRepresentationObject(t, raw)["revisionProvenance"].(map[string]any)
		if len(p) != 4 {
			t.Fatalf("Link context members: %v", p)
		}
	}
}

func TestQueryRepresentationNoInventedProvenanceOrOwnedState(t *testing.T) {
	bead := queryRepresentationBead()
	bead.Attribution = nil
	raw, err := queryResourceJSON(bead)
	if err != nil {
		t.Fatal(err)
	}
	body := queryRepresentationObject(t, raw)
	for _, key := range []string{"attribution", "changeContext", "revisionProvenance", "ownedLinks", "links"} {
		if _, present := body[key]; present {
			t.Fatalf("invented %s", key)
		}
	}
	bead.OwnedLinks = bdpwire.OwnedLinks{}
	raw, err = queryResourceJSON(bead)
	if err != nil {
		t.Fatal(err)
	}
	owned, ok := queryRepresentationObject(t, raw)["ownedLinks"].(map[string]any)
	if !ok || len(owned) != 0 {
		t.Fatal("lost explicitly empty ownership")
	}
	link := queryRepresentationLink()
	link.Attribution = nil
	raw, err = queryResourceJSON(link)
	if err != nil {
		t.Fatal(err)
	}
	if _, present := queryRepresentationObject(t, raw)["revisionProvenance"]; present {
		t.Fatal("invented Link provenance")
	}
}

func TestQueryRepresentationRejectsInvalidStateAndAggregate(t *testing.T) {
	actor := queryRepresentationBead()
	actor.Attribution = &bdpwire.Attribution{Principal: "", Basis: bdpwire.AttributionUnknown}
	basis := queryRepresentationBead()
	basis.Attribution = &bdpwire.Attribution{Principal: "human:Donna", Basis: "authenticated"}
	contextRecord := queryRepresentationBead()
	contextRecord.ChangeContext = &bdpwire.ChangeContext{}
	invalidValue := "forbidden"
	stateRecord := queryRepresentationBead()
	stateRecord.ChangeContext = &bdpwire.ChangeContext{CommittedAt: bdpwire.ContextTime{State: bdpwire.ContextTimeUndetermined}, Agent: bdpwire.ContextString{State: bdpwire.ContextAbsent, Value: &invalidValue}, Message: bdpwire.ContextMessage{State: bdpwire.ContextAbsent}}
	aggregate := queryRepresentationBead()
	aggregate.Links = &bdpwire.LinkCollection{}
	missingMetadata := queryRepresentationBead()
	missingMetadata.Metadata = nil
	badOwned := queryRepresentationBead()
	link := queryRepresentationLink()
	link.Source.URI = queryRepresentationBase + "Different"
	badOwned.OwnedLinks = bdpwire.OwnedLinks{link.Type: {link}}
	missingLinkMetadata := queryRepresentationLink()
	missingLinkMetadata.Metadata = nil
	for i, value := range []any{actor, basis, contextRecord, stateRecord, aggregate, missingMetadata, badOwned, missingLinkMetadata, struct{}{}} {
		if _, err := queryResourceJSON(value); !errors.Is(err, graphstore.ErrInvalidStore) {
			t.Fatalf("case %d: %v", i, err)
		}
	}
}

func TestQueryPredicateAndOutputShareSelectedRevisionProvenance(t *testing.T) {
	bead := queryRepresentationBead()
	raw, err := queryResourceJSON(bead)
	if err != nil {
		t.Fatal(err)
	}
	record, err := resourcequery.NewResource(graph.KindBead, raw)
	if err != nil {
		t.Fatal(err)
	}
	ast := []byte(`{"ir":"bdp-query-design-1","profile":"list","selection":{"op":"filter","input":{"op":"scan","collection":"beads"},"bind":"b","predicate":{"op":"compare","operator":"eq","left":{"op":"field","of":{"op":"binding","name":"b"},"path":["revisionProvenance","actor","id"]},"right":{"op":"literal","value":"human:Donna"}}},"result":{"form":"resources"}}`)
	response, err := resourcequery.Execute(context.Background(), ast, resourcequery.Dataset{Resources: []resourcequery.Resource{record}}, resourcequery.Limits{})
	if err != nil {
		t.Fatal(err)
	}
	result, ok := response.Result.(resourcequery.ResourcesResult)
	if !ok || len(result.Items) != 1 {
		t.Fatalf("predicate failed: %+v", response)
	}
	body := queryRepresentationObject(t, result.Items[0])
	if body["attribution"] != nil || body["changeContext"] != nil || body["revisionProvenance"] == nil {
		t.Fatalf("wrong selected shape: %s", result.Items[0])
	}
	// These are protocol envelope names only: custom nested properties survive.
	if !strings.Contains(string(result.Items[0]), `"custom"`) {
		t.Fatal("lost custom property")
	}
}
