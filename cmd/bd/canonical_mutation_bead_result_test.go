package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/resourceget"
)

// All raw/document fixtures in this file prove REPRESENTATION ONLY. They do not
// set graphstore's private checked14 facet, authenticate a writer or prove commit.
// Actual producer/Close/output integration remains a separately owned gate.
func beadMutationRawFixture(t *testing.T) json.RawMessage {
	t.Helper()
	raw := json.RawMessage("{\"id\":\"https://example.test/work/a\",\"type\":{\"uri\":\"https://types.example.test/Task\",\"revision\":\"type <>& /@= exact\"},\"revision\":\"live <>& /@= R\",\"properties\":{\"n\":1.00,\"exponent\":1e+02,\"body\":\"original <>&\",\"literal\":{\"id\":null,\"next\":[],\"type\":false}}}")
	// Constructor validation is shape validation, NEVER producer authority.
	if _, err := resourceget.NewResourceDocument("https://example.test/work/", graph.KindBead, raw); err != nil {
		t.Fatal(err)
	}
	return raw
}

func beadMutationFields(t *testing.T, raw json.RawMessage, names ...string) map[string]json.RawMessage {
	t.Helper()
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	if len(fields) != len(names) {
		t.Fatalf("unexpected field set: %s", raw)
	}
	for _, name := range names {
		if _, present := fields[name]; !present {
			t.Fatalf("missing %s: %s", name, raw)
		}
	}
	return fields
}

type beadMutationCallerBuiltFixture struct {
	document resourceget.ResourceDocument
}

func (f beadMutationCallerBuiltFixture) ResourceDocument() (resourceget.ResourceDocument, bool) {
	return f.document, true
}

func TestCanonicalBeadMutationMapperPreparesExplicitOperationAndDescriptor(t *testing.T) {
	for _, operation := range []string{"", "Create", "createBead", "link create", "batch"} {
		if mapper, err := newCanonicalBeadMutationMapper(operation); err == nil || mapper != nil {
			t.Fatalf("operation %q admitted", operation)
		}
	}
	for _, operation := range []string{"create", "update", "delete"} {
		mapper, err := newCanonicalBeadMutationMapper(operation)
		if err != nil {
			t.Fatal(err)
		}
		descriptor := mapper.Descriptor()
		if descriptor.BDPSpec != "https://github.com/gastownhall/bdp/schemas/bdp-v0.1.schema.json" || len(descriptor.Required) != 3 ||
			descriptor.Required[0] != "outcome" || descriptor.Required[1] != "changed" {
			t.Fatalf("wrong selected descriptor: %#v", descriptor)
		}
		last := "resource"
		if operation == "delete" {
			last = "deleted"
		}
		if descriptor.Required[2] != last {
			t.Fatalf("wrong operation arm: %#v", descriptor)
		}
		descriptor.Required[0] = "caller-change"
		if mapper.Descriptor().Required[0] != "outcome" {
			t.Fatal("descriptor slices alias mapper state")
		}
	}
}

func TestCanonicalBeadMutationSelectionUsesActualUpdateAndNoopFacts(t *testing.T) {
	memory := graphstore.Record{ID: "https://example.test/work/a", Type: "https://types.example.test/Task", Revision: "same"}
	issue := graphstore.IssueRecord{ID: "https://example.test/work/b", Type: "https://types.example.test/Task", Revision: "issue-R"}
	for _, value := range []any{memory, issue} {
		selected, err := canonicalBeadMutationSelect("create", value)
		if err != nil || !selected.changed {
			t.Fatalf("create selection: %#v %v", selected, err)
		}
	}
	for _, changed := range []bool{false, true} {
		selected, err := canonicalBeadMutationSelect("update", graphstore.MemoryMutationResult{Memory: memory, Changed: changed})
		if err != nil || selected.changed != changed || selected.record.(graphstore.Record).Revision != "same" {
			t.Fatalf("Memory change/no-op truth changed: %#v %v", selected, err)
		}
		selected, err = canonicalBeadMutationSelect("update", graphstore.IssueMutationResult{Issue: issue, Changed: changed, OpenChildren: 7})
		if err != nil || selected.changed != changed || selected.record.(graphstore.IssueRecord).Revision != "issue-R" {
			t.Fatalf("Issue change/no-op truth changed: %#v %v", selected, err)
		}
	}
	if _, err := canonicalBeadMutationSelect("update", graphstore.MemoryMutationResult{Memory: memory, Replaced: &graphstore.ReplacedMemory{}}); !errors.Is(err, errCanonicalBeadMutationResult) {
		t.Fatalf("unsupported replacement admitted: %v", err)
	}
}

func TestCanonicalBeadMutationDeleteSelectionUsesFinalLiveNotTerminal(t *testing.T) {
	memory := graphstore.Record{ID: "https://example.test/work/a", Type: "https://types.example.test/Task", Revision: "final-live"}
	issue := graphstore.IssueRecord{ID: "https://example.test/work/b", Type: "https://types.example.test/Task", Revision: "issue-final-live"}
	selected, err := canonicalBeadMutationSelect("delete", graphstore.MemoryDeleteResult{
		Memory: memory, Deleted: true, Terminal: &graphstore.MemoryTerminalState{ID: memory.ID, Revision: "terminal-not-live"},
	})
	if err != nil || !selected.changed || selected.record.(graphstore.Record).Revision != "final-live" {
		t.Fatalf("Memory terminal replaced final-live selection: %#v %v", selected, err)
	}
	for _, incidents := range [][]string{nil, {}} {
		selected, err = canonicalBeadMutationSelect("delete", graphstore.IssueDeleteResult{
			Issue: issue, Deleted: true, Terminal: &graphstore.BeadTerminalState{ID: issue.ID, Revision: "issue-terminal-not-live"}, IncidentLinkIDs: incidents,
		})
		if err != nil || !selected.changed || selected.record.(graphstore.IssueRecord).Revision != "issue-final-live" {
			t.Fatalf("Issue terminal replaced final-live selection: %#v %v", selected, err)
		}
	}
	// Selection does not manufacture or require a terminal as identity authority.
	if _, err := canonicalBeadMutationSelect("delete", graphstore.MemoryDeleteResult{Memory: memory, Deleted: true}); err != nil {
		t.Fatalf("final-live selection fell back to terminal: %v", err)
	}
}

func TestCanonicalBeadMutationDeleteRefusesPreviewUnappliedAndCascade(t *testing.T) {
	for _, value := range []any{
		graphstore.MemoryDeleteResult{Preview: true},
		graphstore.MemoryDeleteResult{Preview: true, Deleted: true},
		graphstore.MemoryDeleteResult{},
		graphstore.IssueDeleteResult{Preview: true, Deleted: true},
		graphstore.IssueDeleteResult{},
		graphstore.IssueDeleteResult{Deleted: true, IncidentLinkIDs: []string{"https://example.test/work/link"}},
	} {
		if _, err := canonicalBeadMutationSelect("delete", value); !errors.Is(err, errCanonicalBeadMutationResult) {
			t.Fatalf("unsupported delete facts admitted: %T %#v %v", value, value, err)
		}
	}
}

func TestCanonicalBeadMutationProjectRefusesUncapturedAndWrongResultForms(t *testing.T) {
	raw := beadMutationRawFixture(t)
	document, err := resourceget.NewResourceDocument("https://example.test/work/", graph.KindBead, raw)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		operation string
		value     any
	}{
		{"create", document},
		{"create", beadMutationCallerBuiltFixture{document}},
		{"create", graphstore.LinkRecord{}},
		{"create", graphstore.MemoryMutationResult{}},
		{"update", graphstore.Record{}},
		{"delete", graphstore.Record{}},
		{"delete", graphstore.MemoryDeleteResult{Preview: true, Deleted: true}},
		{"delete", graphstore.LinkDeleteResult{}},
		{"create", (*graphstore.Record)(nil)},
	}
	for _, test := range cases {
		mapper, err := newCanonicalBeadMutationMapper(test.operation)
		if err != nil {
			t.Fatal(err)
		}
		projected := mapper.Project(test.value, "original", nil, nil)
		if projected.PayloadErr == nil || len(projected.Payload) != 0 || projected.Disposition != canonicalMutationOutputReturned {
			t.Fatalf("invalid form promoted or known return made unknown: %#v", projected)
		}
		if original, preview := test.value.(graphstore.MemoryDeleteResult); preview {
			kept, ok := projected.Value.(graphstore.MemoryDeleteResult)
			if !ok || kept.Preview != original.Preview || kept.Deleted != original.Deleted {
				t.Fatal("preview output refusal erased original value")
			}
		}
	}
	mapper, _ := newCanonicalBeadMutationMapper("create")
	legacy := graphstore.Record{ID: document.ID(), Type: document.Type().URI, Revision: document.Revision()}
	projected := mapper.Project(legacy, "known original", nil, nil)
	if !errors.Is(projected.PayloadErr, errCanonicalMutationOutputCapture) || len(projected.Payload) != 0 ||
		projected.Disposition != canonicalMutationOutputReturned || projected.Human != "known original" {
		t.Fatalf("absent private checked facet synthesized: %#v", projected)
	}
}

func TestCanonicalBeadMutationProjectionPreservesErrorsAndNeverPromotesErrorValue(t *testing.T) {
	document, err := resourceget.NewResourceDocument("https://example.test/work/", graph.KindBead, beadMutationRawFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	mapper, _ := newCanonicalBeadMutationMapper("create")
	opErr, closeErr := errors.New("operation failed"), errors.New("Close IO failure")
	for _, operationErr := range []error{opErr, errors.Join(opErr, graphstore.ErrOutcomeUnknown)} {
		// A valid representation remains unconfirmed: Project must not even
		// attempt capture/serialization after an operation error.
		projected := mapper.Project(document, "unconfirmed", operationErr, closeErr)
		expected := canonicalMutationOutputFailed
		if errors.Is(operationErr, graphstore.ErrOutcomeUnknown) {
			expected = canonicalMutationOutputUnknown
		}
		if projected.Disposition != expected || projected.OperationErr != operationErr || projected.CloseErr != closeErr ||
			projected.PayloadErr != nil || len(projected.Payload) != 0 || !errors.Is(projected.Err(), opErr) || !errors.Is(projected.Err(), closeErr) {
			t.Fatalf("error value promoted or causes lost: %#v", projected)
		}
		kept, ok := projected.Value.(resourceget.ResourceDocument)
		if !ok || !bytes.Equal(kept.NativeJSON(), document.NativeJSON()) {
			t.Fatal("original unconfirmed value discarded")
		}
	}
}

func TestCanonicalBeadMutationKnownReturnCloseErrorRemainsKnownOutputGap(t *testing.T) {
	mapper, _ := newCanonicalBeadMutationMapper("update")
	closeErr := errors.New("Close IO failure")
	value := graphstore.MemoryMutationResult{Memory: graphstore.Record{
		ID: "https://example.test/work/a", Type: "https://types.example.test/Task", Revision: "unchanged-R",
	}, Changed: false}
	projected := mapper.Project(value, "unchanged original", nil, closeErr)
	kept, ok := projected.Value.(graphstore.MemoryMutationResult)
	if !ok || kept.Changed || kept.Memory.Revision != "unchanged-R" || projected.Human != "unchanged original" ||
		projected.OperationErr != nil || projected.CloseErr != closeErr || !errors.Is(projected.Err(), closeErr) ||
		!errors.Is(projected.PayloadErr, errCanonicalMutationOutputCapture) || projected.Disposition != canonicalMutationOutputReturned {
		t.Fatalf("known no-op return/Close cause changed: %#v", projected)
	}
	// Err() is diagnostic aggregation, not an outcome classifier: a Close
	// error can match the unknown sentinel without making the WRITE unknown.
	closeUnknown := errors.Join(errors.New("physical Close error"), graphstore.ErrOutcomeUnknown)
	closeOnly := mapper.Project(value, "known original", nil, closeUnknown)
	if closeOnly.Disposition != canonicalMutationOutputReturned || closeOnly.OperationErr != nil ||
		closeOnly.CloseErr != closeUnknown || !errors.Is(closeOnly.Err(), graphstore.ErrOutcomeUnknown) ||
		!errors.Is(closeOnly.PayloadErr, errCanonicalMutationOutputCapture) {
		t.Fatalf("joined Close cause changed operation outcome: %#v", closeOnly)
	}
	// This legacy DTO deliberately has no real private facet. It proves
	// preservation/refusal only; root must qualify a real checked14 producer.
}

func TestCanonicalBeadMutationWireCreateAndUpdatePreserveFullNativeBytes(t *testing.T) {
	captured := beadMutationRawFixture(t)
	for _, test := range []struct {
		operation, outcome string
		changed            bool
	}{
		{"create", "created", true},
		{"update", "updated", true},
		{"update", "updated", false},
	} {
		payload, err := canonicalBeadMutationWire(test.operation, test.changed, graph.KindBead, captured)
		if err != nil {
			t.Fatal(err)
		}
		fields := beadMutationFields(t, payload, "outcome", "changed", "resource")
		var outcome string
		var changed bool
		if err := json.Unmarshal(fields["outcome"], &outcome); err != nil || outcome != test.outcome {
			t.Fatalf("wrong native outcome: %s %v", payload, err)
		}
		if err := json.Unmarshal(fields["changed"], &changed); err != nil || changed != test.changed {
			t.Fatalf("wrong actual changed/no-op: %s %v", payload, err)
		}
		if !bytes.Equal(fields["resource"], captured) {
			t.Fatalf("full native Resource changed: %s", payload)
		}
		var resource map[string]json.RawMessage
		if err := json.Unmarshal(fields["resource"], &resource); err != nil {
			t.Fatal(err)
		}
		if _, present := resource["metadata"]; present {
			t.Fatal("optional absent metadata was invented")
		}
		affiliation := beadMutationFields(t, resource["type"], "uri", "revision")
		var pin string
		if err := json.Unmarshal(affiliation["revision"], &pin); err != nil || pin != "type <>& /@= exact" {
			t.Fatalf("exact affiliation lost: %s %v", payload, err)
		}
		if !bytes.Contains(payload, []byte("\"n\":1.00")) || !bytes.Contains(payload, []byte("\"exponent\":1e+02")) ||
			!bytes.Contains(payload, []byte("original <>&")) {
			t.Fatalf("numeric spelling or HTML bytes changed: %s", payload)
		}
		// The no-op arm keeps the original captured revision. There is no
		// current-state/read callback in the mapper that could choose R+1.
		var revision string
		if err := json.Unmarshal(resource["revision"], &revision); err != nil || revision != "live <>& /@= R" {
			t.Fatalf("original revision changed: %s %v", payload, err)
		}
	}
}

func TestCanonicalBeadMutationWireDeleteUsesOnlyFinalLiveIdentity(t *testing.T) {
	captured := beadMutationRawFixture(t)
	payload, err := canonicalBeadMutationWire("delete", true, graph.KindBead, captured)
	if err != nil {
		t.Fatal(err)
	}
	fields := beadMutationFields(t, payload, "outcome", "changed", "deleted")
	deleted := beadMutationFields(t, fields["deleted"], "resourceKind", "resource")
	identity := beadMutationFields(t, deleted["resource"], "id", "type", "revision")
	var capturedFields map[string]json.RawMessage
	if err := json.Unmarshal(captured, &capturedFields); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"id", "type", "revision"} {
		if !bytes.Equal(identity[name], capturedFields[name]) {
			t.Fatalf("final-live %s changed: %s", name, payload)
		}
	}
	if string(fields["outcome"]) != "\"deleted\"" || string(fields["changed"]) != "true" ||
		string(deleted["resourceKind"]) != "\"bead\"" {
		t.Fatalf("wrong applied delete arm: %s", payload)
	}
	// No resource postimage, properties, metadata, terminal revision, local
	// wrapper or ownerRevisions can enter these closed typed wire structures.
	for _, name := range []string{"resource", "ownerRevisions", "terminal", "memory", "issue"} {
		if _, present := fields[name]; present {
			t.Fatalf("extra public Bead member %s: %s", name, payload)
		}
	}
}

func TestCanonicalBeadMutationWireRejectsWrongKindOperationAndUnappliedArms(t *testing.T) {
	captured := beadMutationRawFixture(t)
	for _, test := range []struct {
		operation string
		changed   bool
		kind      graph.ResourceKind
		raw       json.RawMessage
	}{
		{"create", true, graph.KindLink, captured},
		{"create", false, graph.KindBead, captured},
		{"delete", false, graph.KindBead, captured},
		{"link create", true, graph.KindBead, captured},
		{"create", true, graph.KindBead, json.RawMessage("null")},
		{"create", true, graph.KindBead, json.RawMessage("[]")},
		{"create", true, graph.KindBead, json.RawMessage("{}")},
	} {
		if _, err := canonicalBeadMutationWire(test.operation, test.changed, test.kind, test.raw); !errors.Is(err, errCanonicalBeadMutationResult) {
			t.Fatalf("unsupported representation facts admitted: %#v %v", test, err)
		}
	}
}
