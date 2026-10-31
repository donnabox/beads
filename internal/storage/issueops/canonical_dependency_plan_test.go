package issueops

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/storage/depid"
	"github.com/steveyegge/beads/internal/types"
	publicops "github.com/steveyegge/beads/issueops"
)

func canonicalPlanExisting(source, target string, kind types.DependencyType) CanonicalExistingDependency {
	return CanonicalExistingDependency{Pair: CanonicalDependencyPair{SourceID: source, TargetID: target}, NativeID: depid.New(source, target), Type: kind}
}

func canonicalPlanMixedRequest() CanonicalDependencyFinalRequest {
	return CanonicalDependencyFinalRequest{
		Actor:      "actor",
		Population: []string{"d", "b", "a", "c"},
		Metadata:   []CanonicalDependencyMetadataFinal{{Existing: canonicalPlanExisting("c", "d", types.DepWaitsFor), Metadata: json.RawMessage(`{"z":2,"a":1}`)}},
		Insert:     []CanonicalDependencyInsertFinal{{Pair: CanonicalDependencyPair{SourceID: "a", TargetID: "b"}, Type: types.DepBlocks}},
		Delete:     []CanonicalExistingDependency{canonicalPlanExisting("b", "c", types.DepConditionalBlocks)},
		Reassert:   []CanonicalDependencyReassertFinal{{Existing: canonicalPlanExisting("d", "a", types.DepRelated)}},
	}
}

func TestCanonicalDependencyPlanMixedTransitionsAndDetachment(t *testing.T) {
	request := canonicalPlanMixedRequest()
	before, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := NormalizeCanonicalDependencyFinalRequest(request)
	if err != nil {
		t.Fatal(err)
	}
	after, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("normalization mutated its request")
	}
	if !reflect.DeepEqual(plan.Population, []string{"a", "b", "c", "d"}) {
		t.Fatalf("population order: %v", plan.Population)
	}
	want := []CanonicalDependencyTransition{CanonicalDependencyInsertTransition, CanonicalDependencyDeleteTransition, CanonicalDependencyMetadataTransition, CanonicalDependencyReassertTransition}
	for index, transition := range want {
		change := plan.Changes[index]
		if change.Transition != transition {
			t.Fatalf("change %d: %s", index, change.Transition)
		}
		if (change.Existing == nil) != (transition == CanonicalDependencyInsertTransition) {
			t.Fatalf("wrong baseline presence: %+v", change)
		}
		if (change.Metadata == nil) != (transition == CanonicalDependencyDeleteTransition) {
			t.Fatalf("wrong metadata presence: %+v", change)
		}
		if change.Existing != nil && (change.Pair != change.Existing.Pair || change.Type != change.Existing.Type) {
			t.Fatalf("changed identity/classification: %+v", change)
		}
	}
	if string(plan.Changes[0].Metadata) != `{}` || string(plan.Changes[2].Metadata) != `{"a":1,"z":2}` || string(plan.Changes[3].Metadata) != `{}` {
		t.Fatalf("metadata normalization: %+v", plan.Changes)
	}
	request.Population[0] = "caller-mutated"
	request.Metadata[0].Metadata[2] = 'x'
	request.Metadata[0].Existing.NativeID = "caller-mutated"
	if plan.Population[3] != "d" || string(plan.Changes[2].Metadata) != `{"a":1,"z":2}` || plan.Changes[2].Existing.NativeID != depid.New("c", "d") {
		t.Fatal("request still aliases plan")
	}
	request = canonicalPlanMixedRequest()
	plan, err = NormalizeCanonicalDependencyFinalRequest(request)
	if err != nil {
		t.Fatal(err)
	}
	plan.Population[0] = "result-mutated"
	plan.Changes[2].Existing.NativeID = "result-mutated"
	plan.Changes[2].Metadata[2] = 'x'
	if request.Population[0] != "d" || request.Metadata[0].Existing.NativeID != depid.New("c", "d") || string(request.Metadata[0].Metadata) != `{"z":2,"a":1}` {
		t.Fatal("plan still aliases caller data")
	}
}

func TestCanonicalDependencyPlanEquivalentOrder(t *testing.T) {
	first := canonicalPlanMixedRequest()
	first.Insert = append(first.Insert, CanonicalDependencyInsertFinal{Pair: CanonicalDependencyPair{SourceID: "b", TargetID: "d"}, Type: types.DepParentChild, Metadata: json.RawMessage(`{"second":2,"first":1}`)})
	second := canonicalPlanMixedRequest()
	second.Population = []string{"c", "a", "b", "d"}
	second.Metadata[0].Metadata = json.RawMessage(`{"a":1,"z":2}`)
	second.Insert = []CanonicalDependencyInsertFinal{
		{Pair: CanonicalDependencyPair{SourceID: "b", TargetID: "d"}, Type: types.DepParentChild, Metadata: json.RawMessage(`{"first":1,"second":2}`)},
		{Pair: CanonicalDependencyPair{SourceID: "a", TargetID: "b"}, Type: types.DepBlocks, Metadata: json.RawMessage(`{}`)},
	}
	left, err := NormalizeCanonicalDependencyFinalRequest(first)
	if err != nil {
		t.Fatal(err)
	}
	right, err := NormalizeCanonicalDependencyFinalRequest(second)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(left, right) {
		t.Fatalf("equivalent final intents differ: %#v / %#v", left, right)
	}
}

func TestCanonicalDependencyPlanInvalidInputReturnsZero(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*CanonicalDependencyFinalRequest)
	}{
		{"missing actor", func(r *CanonicalDependencyFinalRequest) { r.Actor = "" }},
		{"invalid actor", func(r *CanonicalDependencyFinalRequest) { r.Actor = string([]byte{255}) }},
		{"duplicate population", func(r *CanonicalDependencyFinalRequest) { r.Population = append(r.Population, "a") }},
		{"empty population ID", func(r *CanonicalDependencyFinalRequest) { r.Population[0] = "" }},
		{"invalid population UTF-8", func(r *CanonicalDependencyFinalRequest) { r.Population[0] = string([]byte{255}) }},
		{"native separator", func(r *CanonicalDependencyFinalRequest) { r.Population[0] = "d\x1fhidden" }},
		{"outside source", func(r *CanonicalDependencyFinalRequest) { r.Insert[0].Pair.SourceID = "outside" }},
		{"outside target", func(r *CanonicalDependencyFinalRequest) { r.Insert[0].Pair.TargetID = "outside" }},
		{"missing target", func(r *CanonicalDependencyFinalRequest) { r.Insert[0].Pair.TargetID = "" }},
		{"unknown insert kind", func(r *CanonicalDependencyFinalRequest) { r.Insert[0].Type = "custom" }},
		{"missing insert kind", func(r *CanonicalDependencyFinalRequest) { r.Insert[0].Type = "" }},
		{"unknown baseline kind", func(r *CanonicalDependencyFinalRequest) { r.Delete[0].Type = "custom" }},
		{"missing native ID", func(r *CanonicalDependencyFinalRequest) { r.Delete[0].NativeID = "" }},
		{"different pair native ID", func(r *CanonicalDependencyFinalRequest) { r.Delete[0].NativeID = depid.New("c", "b") }},
		{"uppercase native ID", func(r *CanonicalDependencyFinalRequest) { r.Delete[0].NativeID = strings.ToUpper(r.Delete[0].NativeID) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := canonicalPlanMixedRequest()
			tc.change(&request)
			plan, err := NormalizeCanonicalDependencyFinalRequest(request)
			if !errors.Is(err, storage.ErrValidation) || !reflect.ValueOf(plan).IsZero() {
				t.Fatalf("expected zero invalid plan, got %+v / %v", plan, err)
			}
		})
	}
}

func TestCanonicalDependencyPlanMetadataObjects(t *testing.T) {
	for _, raw := range []json.RawMessage{nil, {}, json.RawMessage(`null`), json.RawMessage(`[]`), json.RawMessage(`1`), json.RawMessage(`"text"`), json.RawMessage(`{"duplicate":1,"duplicate":2}`), json.RawMessage{'{', '"', 'x', '"', ':', '"', 255, '"', '}'}} {
		request := canonicalPlanMixedRequest()
		request.Metadata[0].Metadata = raw
		plan, err := NormalizeCanonicalDependencyFinalRequest(request)
		if !errors.Is(err, storage.ErrValidation) || !reflect.ValueOf(plan).IsZero() {
			t.Fatalf("invalid replacement %q accepted: %+v / %v", raw, plan, err)
		}
	}
	for _, kind := range []string{"insert", "reassert"} {
		t.Run(kind, func(t *testing.T) {
			for _, raw := range []json.RawMessage{json.RawMessage(`null`), json.RawMessage(`[]`)} {
				request := canonicalPlanMixedRequest()
				if kind == "insert" {
					request.Insert[0].Metadata = raw
				} else {
					request.Reassert[0].Metadata = raw
				}
				plan, err := NormalizeCanonicalDependencyFinalRequest(request)
				if !errors.Is(err, storage.ErrValidation) || !reflect.ValueOf(plan).IsZero() {
					t.Fatalf("invalid creation metadata accepted: %+v / %v", plan, err)
				}
			}
		})
	}
}

func TestCanonicalDependencyPlanDuplicatePairTransitions(t *testing.T) {
	appendTransition := func(request *CanonicalDependencyFinalRequest, transition int) {
		existing := canonicalPlanExisting("a", "b", types.DepBlocks)
		switch transition {
		case 0:
			request.Metadata = append(request.Metadata, CanonicalDependencyMetadataFinal{Existing: existing, Metadata: json.RawMessage(`{}`)})
		case 1:
			request.Insert = append(request.Insert, CanonicalDependencyInsertFinal{Pair: existing.Pair, Type: existing.Type})
		case 2:
			request.Delete = append(request.Delete, existing)
		case 3:
			request.Reassert = append(request.Reassert, CanonicalDependencyReassertFinal{Existing: existing})
		}
	}
	for left := 0; left < 4; left++ {
		for right := left; right < 4; right++ {
			t.Run(fmt.Sprintf("%d-%d", left, right), func(t *testing.T) {
				request := CanonicalDependencyFinalRequest{Actor: "actor", Population: []string{"a", "b"}}
				appendTransition(&request, left)
				appendTransition(&request, right)
				plan, err := NormalizeCanonicalDependencyFinalRequest(request)
				if !errors.Is(err, storage.ErrValidation) || !reflect.ValueOf(plan).IsZero() {
					t.Fatalf("duplicate pair accepted: %+v / %v", plan, err)
				}
			})
		}
	}
}

func TestCanonicalDependencyPlanOrdinaryDottedAdmission(t *testing.T) {
	existing := canonicalPlanExisting("bd-parent.1", "bd-parent", types.DepBlocks)
	for _, transition := range []CanonicalDependencyTransition{CanonicalDependencyMetadataTransition, CanonicalDependencyInsertTransition, CanonicalDependencyReassertTransition} {
		request := CanonicalDependencyFinalRequest{Actor: "actor", Population: []string{"bd-parent", "bd-parent.1"}}
		switch transition {
		case CanonicalDependencyMetadataTransition:
			request.Metadata = []CanonicalDependencyMetadataFinal{{Existing: existing, Metadata: json.RawMessage(`{}`)}}
		case CanonicalDependencyInsertTransition:
			request.Insert = []CanonicalDependencyInsertFinal{{Pair: existing.Pair, Type: existing.Type}}
		case CanonicalDependencyReassertTransition:
			request.Reassert = []CanonicalDependencyReassertFinal{{Existing: existing}}
		}
		plan, err := NormalizeCanonicalDependencyFinalRequest(request)
		var dotted *publicops.DottedChildDependencyError
		if !errors.As(err, &dotted) || !reflect.ValueOf(plan).IsZero() {
			t.Fatalf("%s lost ordinary dotted-child refusal: %+v / %v", transition, plan, err)
		}
	}
	// Removal must still permit cleanup of a preexisting forbidden assertion.
	plan, err := NormalizeCanonicalDependencyFinalRequest(CanonicalDependencyFinalRequest{Actor: "actor", Population: []string{"bd-parent", "bd-parent.1"}, Delete: []CanonicalExistingDependency{existing}})
	if err != nil || len(plan.Changes) != 1 || plan.Changes[0].Transition != CanonicalDependencyDeleteTransition {
		t.Fatalf("ordinary deletion newly restricted: %+v / %v", plan, err)
	}
	existing.Type = types.DepParentChild
	if _, err := NormalizeCanonicalDependencyFinalRequest(CanonicalDependencyFinalRequest{Actor: "actor", Population: []string{"bd-parent", "bd-parent.1"}, Insert: []CanonicalDependencyInsertFinal{{Pair: existing.Pair, Type: existing.Type}}}); err != nil {
		t.Fatalf("ordinary immediate hierarchy edge refused: %v", err)
	}
}

func TestCanonicalDependencyPlanSelfPairs(t *testing.T) {
	existing := canonicalPlanExisting("a", "a", types.DepRelated)
	for _, request := range []CanonicalDependencyFinalRequest{
		{Actor: "actor", Population: []string{"a"}, Metadata: []CanonicalDependencyMetadataFinal{{Existing: existing, Metadata: json.RawMessage(`{}`)}}},
		{Actor: "actor", Population: []string{"a"}, Insert: []CanonicalDependencyInsertFinal{{Pair: existing.Pair, Type: existing.Type}}},
		{Actor: "actor", Population: []string{"a"}, Delete: []CanonicalExistingDependency{existing}},
		{Actor: "actor", Population: []string{"a"}, Reassert: []CanonicalDependencyReassertFinal{{Existing: existing}}},
	} {
		plan, err := NormalizeCanonicalDependencyFinalRequest(request)
		if !errors.Is(err, publicops.ErrSelfDependency) || !reflect.ValueOf(plan).IsZero() {
			t.Fatalf("self pair accepted or reclassified: %+v / %v", plan, err)
		}
	}
}

func TestCanonicalDependencyPlanBoundsAndNativeKinds(t *testing.T) {
	request := CanonicalDependencyFinalRequest{Actor: "actor", Population: make([]string, CanonicalDependencyPlanLimit)}
	for index := range request.Population {
		request.Population[index] = fmt.Sprintf("native-%d", index)
		if index > 0 {
			request.Insert = append(request.Insert, CanonicalDependencyInsertFinal{Pair: CanonicalDependencyPair{SourceID: "native-0", TargetID: request.Population[index]}, Type: types.DepRelated})
		}
	}
	request.Delete = []CanonicalExistingDependency{canonicalPlanExisting("native-1", "native-0", types.DepRelated)}
	plan, err := NormalizeCanonicalDependencyFinalRequest(request)
	if err != nil || len(plan.Population) != CanonicalDependencyPlanLimit || len(plan.Changes) != CanonicalDependencyPlanLimit {
		t.Fatalf("exact boundary failed: %d / %d / %v", len(plan.Population), len(plan.Changes), err)
	}
	tooMany := request
	tooMany.Population = append(append([]string(nil), request.Population...), "one-more")
	if plan, err := NormalizeCanonicalDependencyFinalRequest(tooMany); !errors.Is(err, storage.ErrValidation) || !reflect.ValueOf(plan).IsZero() {
		t.Fatalf("population overflow accepted: %+v / %v", plan, err)
	}
	tooMany = request
	tooMany.Reassert = []CanonicalDependencyReassertFinal{{Existing: canonicalPlanExisting("native-2", "native-0", types.DepRelated)}}
	if plan, err := NormalizeCanonicalDependencyFinalRequest(tooMany); !errors.Is(err, storage.ErrValidation) || !reflect.ValueOf(plan).IsZero() {
		t.Fatalf("combined transition overflow accepted: %+v / %v", plan, err)
	}
	for _, kind := range types.WellKnownDependencyTypes() {
		plan, err := NormalizeCanonicalDependencyFinalRequest(CanonicalDependencyFinalRequest{Actor: "actor", Population: []string{"a", "b"}, Insert: []CanonicalDependencyInsertFinal{{Pair: CanonicalDependencyPair{SourceID: "a", TargetID: "b"}, Type: kind}}})
		if err != nil || len(plan.Changes) != 1 || plan.Changes[0].Type != kind {
			t.Fatalf("recognized native kind %q refused: %+v / %v", kind, plan, err)
		}
	}
}

func TestCanonicalDependencyPlanOpaqueIDsAndCodeUnitOrder(t *testing.T) {
	request := CanonicalDependencyFinalRequest{Actor: " actor ", Population: []string{"\uffff", "\U0001f600", "team:note"}, Insert: []CanonicalDependencyInsertFinal{
		{Pair: CanonicalDependencyPair{SourceID: "\uffff", TargetID: "team:note"}, Type: types.DepRelated},
		{Pair: CanonicalDependencyPair{SourceID: "\U0001f600", TargetID: "team:note"}, Type: types.DepRelated},
	}}
	plan, err := NormalizeCanonicalDependencyFinalRequest(request)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Actor != request.Actor || !reflect.DeepEqual(plan.Population, []string{"team:note", "\U0001f600", "\uffff"}) || plan.Changes[0].Pair.SourceID != "\U0001f600" {
		t.Fatalf("normalized literal data or wrong ordering: %+v", plan)
	}
	empty, err := NormalizeCanonicalDependencyFinalRequest(CanonicalDependencyFinalRequest{Actor: "actor"})
	if err != nil || len(empty.Changes) != 0 || len(empty.Population) != 0 {
		t.Fatalf("empty final native intent refused: %+v / %v", empty, err)
	}
}
