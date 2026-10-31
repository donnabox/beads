package graphstore

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/canonicalbatch"
	"github.com/steveyegge/beads/internal/canonicalbatchstate"
)

func TestAtomicBatchTypeAuthority(t *testing.T) {
	const base, child, leaf, link = "https://example.invalid/base", "https://example.invalid/child", "https://example.invalid/leaf", "https://example.invalid/link"
	descriptor := func(id string, kind graph.ResourceKind, parents ...string) graph.TypeDescriptor {
		t.Helper()
		spec := graph.TypeDescriptorSpec{ID: id, Name: id, Describes: kind, ConformsTo: parents}
		if kind == graph.KindLink {
			source, target := graph.EndpointConstraint{}, graph.EndpointConstraint{}
			spec.Source, spec.Target = &source, &target
		}
		d, err := graph.NewTypeDescriptor(spec)
		if err != nil {
			t.Fatal(err)
		}
		return d
	}
	facts := []graph.TypeDescriptor{descriptor(leaf, graph.KindBead, child), descriptor(base, graph.KindBead), descriptor(child, graph.KindBead, base), descriptor(link, graph.KindLink)}
	kinds, families, err := atomicBatchTypeAuthority(context.Background(), facts)
	if err != nil || kinds[leaf] != graph.KindBead || !reflect.DeepEqual(families[leaf], []string{base, child, leaf}) || !reflect.DeepEqual(families[link], []string{link}) {
		t.Fatalf("invalid checked closure: %v %v %v", kinds, families, err)
	}
	for _, invalid := range [][]graph.TypeDescriptor{
		{facts[0]}, {facts[0], facts[0]}, {descriptor(child, graph.KindBead, link), facts[3]},
	} {
		if _, _, err := atomicBatchTypeAuthority(context.Background(), invalid); !errors.Is(err, ErrInvalidStore) {
			t.Fatalf("incomplete inventory accepted: %v", err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := atomicBatchTypeAuthority(ctx, facts); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled capture accepted: %v", err)
	}
}

func TestAtomicBatchMemberIdentityArrays(t *testing.T) {
	members := []canonicalbatch.Member{{Operation: "createBead", Name: "created"}, {Operation: "select", Name: "chosen"}, {Operation: "updateBead", Bead: canonicalbatch.Operand{Binding: "chosen"}}, {Operation: "select", Name: "empty"}, {Operation: "deleteBead", Bead: canonicalbatch.Operand{Binding: "empty"}}}
	plan := canonicalbatchstate.Plan{
		Mappings:   []canonicalbatchstate.Mapping{{OperationIndex: 0, Operation: "createBead", ID: "https://example.invalid/b"}, {OperationIndex: 2, Operation: "updateBead", ID: "https://example.invalid/b"}, {OperationIndex: 2, Operation: "updateBead", ID: "https://example.invalid/a"}},
		Selections: []canonicalbatchstate.SelectionCapture{{OperationIndex: 1, IDs: []string{"https://example.invalid/b", "https://example.invalid/a"}}, {OperationIndex: 3, IDs: []string{}}},
	}
	rows, err := atomicBatchOperationResults(members, plan)
	if err != nil || len(rows) != len(members) || !reflect.DeepEqual(rows[1].IDs, []string{"https://example.invalid/a", "https://example.invalid/b"}) || !reflect.DeepEqual(rows[1].IDs, rows[2].IDs) || rows[3].IDs == nil || rows[4].IDs == nil {
		t.Fatalf("incomplete mapping: %+v %v", rows, err)
	}
	retained := BatchResult{CapabilityVersion: canonicalbatch.Version, Results: rows, Resources: []BatchResourceOutcome{}}
	members[1].Name = "renamed"
	members[2].Bead.Binding = "renamed"
	replayed, err := projectAtomicBatchLabels(retained, members)
	if err != nil || replayed.Results[1].OperationName != "renamed" || replayed.Changed || replayed.Transaction != "" {
		t.Fatalf("read-only replay failed: %+v %v", replayed, err)
	}
	raw, err := json.Marshal(replayed)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["results"].([]any)[4].(map[string]any)["ids"] == nil {
		t.Fatalf("empty mapping was null: %s", raw)
	}
	if _, err := atomicBatchOperationResults(members, canonicalbatchstate.Plan{}); !errors.Is(err, ErrInvalidStore) {
		t.Fatalf("missing singleton accepted: %v", err)
	}
	rows[1].IDs = append(rows[1].IDs, rows[1].IDs[0])
	if _, err := projectAtomicBatchLabels(retained, members); !errors.Is(err, ErrInvalidStore) {
		t.Fatalf("invalid retained array accepted: %v", err)
	}
}

func TestAtomicBatchRetainedScalarCandidateCompatibility(t *testing.T) {
	const raw = `{"capabilityVersion":"2","changed":false,"results":[{"operationIndex":0,"operation":"updateBead","id":"https://example.invalid/a"}],"resources":[]}`
	var retained BatchResult
	if err := json.Unmarshal([]byte(raw), &retained); err != nil {
		t.Fatal(err)
	}
	replayed, err := projectAtomicBatchLabels(retained, []canonicalbatch.Member{{Operation: "updateBead"}})
	if err != nil {
		t.Fatal(err)
	}
	got, err := canonicalJSON(replayed)
	if err != nil {
		t.Fatal(err)
	}
	want, err := graph.CanonicalizeJSON([]byte(raw))
	if err != nil || string(got) != string(want) {
		t.Fatalf("old retained result was rewritten: %s %v", got, err)
	}
	retained.Results[0].IDs = []string{"https://example.invalid/a"}
	if _, err := projectAtomicBatchLabels(retained, []canonicalbatch.Member{{Operation: "updateBead"}}); !errors.Is(err, ErrInvalidStore) {
		t.Fatalf("mixed retained result encodings accepted: %v", err)
	}
}

func TestAtomicBatchResultMappingPresence(t *testing.T) {
	for _, fields := range []string{
		`"id":"https://example.invalid/a","ids":null`,
		`"id":"","ids":[]`, `"id":null`, `"ids":null`,
		`"ids":[],"operationName":null`, `"ids":[],"operationName":""`, `"ids":[],"future":true`,
		`"ids":[],"ids":[]`,
	} {
		var result BatchOperationResult
		if err := json.Unmarshal([]byte(`{"operationIndex":0,"operation":"updateBead",`+fields+`}`), &result); err == nil {
			t.Fatalf("malformed mapping was normalized: %s", fields)
		}
	}
	var empty BatchOperationResult
	if err := json.Unmarshal([]byte(`{"operationIndex":0,"operation":"updateBead","ids":[]}`), &empty); err != nil || empty.IDs == nil {
		t.Fatalf("valid empty array lost presence: %+v %v", empty, err)
	}
	if _, err := json.Marshal(BatchOperationResult{Operation: "updateBead", ID: "https://example.invalid/a", IDs: []string{}}); err == nil {
		t.Fatal("producer serialized a mixed mapping")
	}
}
