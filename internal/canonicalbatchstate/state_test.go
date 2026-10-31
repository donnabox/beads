package canonicalbatchstate

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/canonicalbatch"
)

const testScope = "https://work.example/acme/"

func reference(id string) canonicalbatch.Reference {
	return canonicalbatch.Reference{URI: testScope + id}
}
func operand(id string) canonicalbatch.Operand { return canonicalbatch.Operand{URI: testScope + id} }
func raw(value string) json.RawMessage         { return json.RawMessage(value) }
func existing(id string) Resource {
	return Resource{ID: testScope + id, Kind: graphops.KindBead, Revision: "baseline-" + id, Type: reference("memory"), Properties: raw(`{"title":"initial"}`), Metadata: raw(`{}`), Envelope: raw(`{"original":"retained exactly"}`)}
}
func createBead(id, name string) canonicalbatch.Member {
	return canonicalbatch.Member{Operation: "createBead", Kind: graphops.KindBead, ID: testScope + id, Name: name, Type: reference("memory"), Properties: raw(`{"title":"new"}`)}
}
func createLink(id, source, target string) canonicalbatch.Member {
	return canonicalbatch.Member{Operation: "createLink", Kind: graphops.KindLink, ID: testScope + id, Type: reference("references"), Source: operand(source), Target: operand(target), Properties: raw(`{}`)}
}

func TestCreateTwoBeadsAndLink(t *testing.T) {
	plan, err := Stage(Input{Scope: testScope, Members: []canonicalbatch.Member{createBead("one", ""), createBead("two", ""), createLink("citation", "one", "two")}})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Mappings) != 3 || len(plan.Outcomes) != 3 || len(plan.Final) != 3 || len(plan.ConsumedIDs) != 3 {
		t.Fatalf("incomplete graph: %+v", plan)
	}
	link := plan.Final[testScope+"citation"]
	if link.Source != reference("one") || link.Target != reference("two") {
		t.Fatalf("incorrect Link: %+v", link)
	}
	for _, o := range plan.Outcomes {
		if o.Outcome != "created" || !o.Changed || o.Before != nil || o.After == nil || o.After.Revision != "" {
			t.Fatalf("creation published an invented version: %+v", o)
		}
	}
}

func TestCallerAssignedLocalIDWithColon(t *testing.T) {
	member := createBead("ignored", "")
	member.ID = "team:note"
	plan, err := Stage(Input{Scope: testScope, Members: []canonicalbatch.Member{member}})
	if err != nil {
		t.Fatal(err)
	}
	if plan.Mappings[0].ID != testScope+"team:note" {
		t.Fatalf("local ID changed: %+v", plan.Mappings)
	}
}

func TestBindingsCaptureGeneratedIdentity(t *testing.T) {
	one, two := createBead("ignored", "one"), createBead("ignored", "two")
	one.ID = ""
	two.ID = ""
	link := createLink("citation", "ignored", "ignored")
	link.Source = canonicalbatch.Operand{Binding: "one"}
	link.Target = canonicalbatch.Operand{Binding: "two"}
	allocate := func(index int, kind graphops.ResourceKind) (string, error) {
		if kind != graphops.KindBead {
			t.Fatal("unexpected allocation kind")
		}
		return []string{testScope + "first", testScope + "second"}[index], nil
	}
	plan, err := Stage(Input{Scope: testScope, Members: []canonicalbatch.Member{one, two, link}, Allocate: allocate})
	if err != nil {
		t.Fatal(err)
	}
	if plan.Final[testScope+"citation"].Source.URI != testScope+"first" || plan.Final[testScope+"citation"].Target.URI != testScope+"second" {
		t.Fatalf("binding is not identity capture: %+v", plan.Final)
	}
	if plan.Mappings[0].OperationName != "one" || plan.Mappings[0].ID != testScope+"first" {
		t.Fatalf("missing ordered mapping: %+v", plan.Mappings)
	}
}

func TestRestoringBaselineIsNetNoop(t *testing.T) {
	base := existing("one")
	update := canonicalbatch.Member{Operation: "updateBead", Bead: operand("one"), ExpectedRevision: base.Revision, Properties: raw(`{"title":"temporary"}`)}
	restore := update
	restore.Properties = bytes.Clone(base.Properties)
	plan, err := Stage(Input{Scope: testScope, Current: map[string]Resource{base.ID: base}, Members: []canonicalbatch.Member{update, restore}})
	if err != nil {
		t.Fatal(err)
	}
	o := plan.Outcomes[0]
	if o.Outcome != "unchanged" || o.Changed || !reflect.DeepEqual(*o.After, base) || !reflect.DeepEqual(plan.Final[base.ID], base) {
		t.Fatalf("no-op lost baseline envelope: %+v", o)
	}
	if len(plan.Mappings) != 2 {
		t.Fatalf("operation mappings were collapsed: %+v", plan.Mappings)
	}
}

func TestGuardsAlwaysCompareCommittedBaseline(t *testing.T) {
	base := existing("one")
	first := canonicalbatch.Member{Operation: "updateBead", Bead: operand("one"), ExpectedRevision: base.Revision, PropertiesChange: raw(`[{"op":"replace","path":"/title","value":"first"}]`)}
	second := first
	second.PropertiesChange = raw(`[{"op":"replace","path":"/title","value":"second"}]`)
	plan, err := Stage(Input{Scope: testScope, Current: map[string]Resource{base.ID: base}, Members: []canonicalbatch.Member{first, second}})
	if err != nil {
		t.Fatal(err)
	}
	if plan.Outcomes[0].Outcome != "updated" || string(plan.Outcomes[0].After.Properties) != `{"title":"second"}` || plan.Outcomes[0].After.Revision != base.Revision {
		t.Fatalf("bad final state: %+v", plan.Outcomes[0])
	}
}

func TestStaleGuardReturnsNoPlanAndPreservesInput(t *testing.T) {
	base := existing("one")
	baseline := map[string]Resource{base.ID: base}
	first := canonicalbatch.Member{Operation: "updateBead", Bead: operand("one"), Properties: raw(`{"title":"temporary"}`)}
	stale := first
	stale.ExpectedRevision = "stale"
	plan, err := Stage(Input{Scope: testScope, Current: baseline, Members: []canonicalbatch.Member{first, stale}})
	var memberErr *MemberError
	if !errors.Is(err, ErrRevision) || !errors.As(err, &memberErr) || memberErr.OperationIndex != 1 {
		t.Fatalf("wrong refusal: %v", err)
	}
	if !reflect.DeepEqual(plan, Plan{}) || !reflect.DeepEqual(baseline[base.ID], base) {
		t.Fatalf("partial output or caller mutation: %+v %+v", plan, baseline)
	}
}

func TestCreateDeleteConsumesIdentityWithoutVersion(t *testing.T) {
	create := createBead("one", "one")
	remove := canonicalbatch.Member{Operation: "deleteBead", Bead: canonicalbatch.Operand{Binding: "one"}}
	plan, err := Stage(Input{Scope: testScope, Members: []canonicalbatch.Member{create, remove}})
	if err != nil {
		t.Fatal(err)
	}
	o := plan.Outcomes[0]
	if o.Outcome != "allocated" || !o.Changed || o.Before != nil || o.After != nil || len(plan.Final) != 0 || len(plan.ConsumedIDs) != 1 {
		t.Fatalf("invented deleted Resource: %+v", plan)
	}
	_, err = Stage(Input{Scope: testScope, Members: []canonicalbatch.Member{create, remove, createBead("one", "")}})
	if !errors.Is(err, ErrAllocated) {
		t.Fatalf("consumed ID reused: %v", err)
	}
}

func TestDeleteReportsBaselineNotStagedVersion(t *testing.T) {
	base := existing("one")
	update := canonicalbatch.Member{Operation: "updateBead", Bead: operand("one"), Properties: raw(`{"title":"temporary"}`)}
	remove := canonicalbatch.Member{Operation: "deleteBead", Bead: operand("one"), ExpectedRevision: base.Revision}
	plan, err := Stage(Input{Scope: testScope, Current: map[string]Resource{base.ID: base}, Members: []canonicalbatch.Member{update, remove}})
	if err != nil {
		t.Fatal(err)
	}
	o := plan.Outcomes[0]
	if o.Outcome != "deleted" || o.After != nil || !reflect.DeepEqual(*o.Before, base) {
		t.Fatalf("deletion leaked staged state: %+v", o)
	}
}

func TestNewCreationCannotSatisfyGuardOrPin(t *testing.T) {
	create := createBead("one", "one")
	guard := canonicalbatch.Member{Operation: "updateBead", Bead: canonicalbatch.Operand{Binding: "one"}, ExpectedRevision: "guessed", Properties: raw(`{}`)}
	if _, err := Stage(Input{Scope: testScope, Members: []canonicalbatch.Member{create, guard}}); !errors.Is(err, ErrRevision) {
		t.Fatalf("creation satisfied guard: %v", err)
	}
	link := createLink("citation", "one", "one")
	link.Source.Revision = "guessed"
	if _, err := Stage(Input{Scope: testScope, Members: []canonicalbatch.Member{create, link}}); !errors.Is(err, ErrRevision) {
		t.Fatalf("creation satisfied exact pin: %v", err)
	}
}

func TestTypeAndEndpointEditsPreserveOmissionsAndClearPins(t *testing.T) {
	one, two := existing("one"), existing("two")
	link := Resource{ID: testScope + "citation", Kind: graphops.KindLink, Revision: "link-baseline", Type: reference("references"), Properties: raw(`{}`), Metadata: raw(`{}`), Source: canonicalbatch.Reference{URI: one.ID, Revision: one.Revision}, Target: canonicalbatch.Reference{URI: two.ID, Revision: two.Revision}}
	change := canonicalbatch.Member{Operation: "updateLink", Link: operand("citation"), Source: operand("one"), ExpectedRevision: link.Revision}
	plan, err := Stage(Input{Scope: testScope, Current: map[string]Resource{one.ID: one, two.ID: two, link.ID: link}, Members: []canonicalbatch.Member{change}})
	if err != nil {
		t.Fatal(err)
	}
	after := plan.Final[link.ID]
	if after.Source.Revision != "" || after.Target != link.Target || after.Type != link.Type {
		t.Fatalf("whole Reference replacement or omission broken: %+v", after)
	}
	change.Type = canonicalbatch.Reference{URI: testScope + "new-type", Revision: "type-pin"}
	plan, err = Stage(Input{Scope: testScope, Current: map[string]Resource{one.ID: one, two.ID: two, link.ID: link}, Members: []canonicalbatch.Member{change}})
	if err != nil || plan.Final[link.ID].Type != change.Type {
		t.Fatalf("whole Type affiliation lost: %+v %v", plan, err)
	}
}

func TestAllSyntaxAdmittedBeforeAllocator(t *testing.T) {
	create := createBead("ignored", "")
	create.ID = ""
	bad := canonicalbatch.Member{Operation: "updateBead", Bead: operand("one"), PropertiesChange: raw(`[{"op":"replace","path":"bad","value":1}]`)}
	called := false
	_, err := Stage(Input{Scope: testScope, Members: []canonicalbatch.Member{create, bad}, Allocate: func(int, graphops.ResourceKind) (string, error) { called = true; return testScope + "one", nil }})
	if err == nil || called {
		t.Fatalf("malformed later input allowed allocation: called=%v err=%v", called, err)
	}
}

func TestInputAndResultNeverShareMutableBytes(t *testing.T) {
	base := existing("one")
	member := canonicalbatch.Member{Operation: "updateBead", Bead: operand("one"), Properties: raw(`{"title":"new"}`)}
	plan, err := Stage(Input{Scope: testScope, Current: map[string]Resource{base.ID: base}, Members: []canonicalbatch.Member{member}})
	if err != nil {
		t.Fatal(err)
	}
	plan.Final[base.ID].Properties[2] = 'X'
	plan.Outcomes[0].Before.Envelope[2] = 'X'
	if string(base.Envelope) != `{"original":"retained exactly"}` || string(member.Properties) != `{"title":"new"}` || string(plan.Outcomes[0].After.Properties) != `{"title":"new"}` {
		t.Fatalf("mutable bytes shared across boundaries")
	}
}

func TestProvenanceConflictOnlyWhenVersionWillPublish(t *testing.T) {
	base := existing("one")
	first := canonicalbatch.Member{Operation: "updateBead", Bead: operand("one"), Properties: raw(`{"title":"temporary"}`), RevisionProvenance: raw(`{"message":"first"}`)}
	second := first
	second.RevisionProvenance = raw(`{"message":"second"}`)
	_, err := Stage(Input{Scope: testScope, Current: map[string]Resource{base.ID: base}, Members: []canonicalbatch.Member{first, second}})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("conflicting provenance published: %v", err)
	}
	second.Properties = bytes.Clone(base.Properties)
	plan, err := Stage(Input{Scope: testScope, Current: map[string]Resource{base.ID: base}, Members: []canonicalbatch.Member{first, second}})
	if err != nil || plan.Outcomes[0].Changed {
		t.Fatalf("provenance alone changed net no-op: %+v %v", plan, err)
	}
}

func TestRefuseUnsupportedSelectionBeforeAllocator(t *testing.T) {
	create := createBead("ignored", "")
	create.ID = ""
	called := false
	_, err := Stage(Input{Scope: testScope, Members: []canonicalbatch.Member{create, {Operation: "select", Name: "unsupported", Query: raw(`{"ir":"bdp-query-design-1","profile":"query","selection":{"op":"scan","collection":"beads"}}`)}}, Allocate: func(int, graphops.ResourceKind) (string, error) { called = true; return testScope + "one", nil }})
	if !errors.Is(err, ErrUnsupported) || called {
		t.Fatalf("unsupported selection was partially staged: %v", err)
	}
}

func TestAbsentMetadataPatchesPreserveNoopEnvelope(t *testing.T) {
	for _, tc := range []struct {
		name    string
		changes []string
	}{
		{"empty patch", []string{`[]`}},
		{"single patch cancellation", []string{`[{"op":"add","path":"/temporary","value":true},{"op":"remove","path":"/temporary"}]`}},
		{"separate member cancellation", []string{`[{"op":"add","path":"/temporary","value":true}]`, `[{"op":"remove","path":"/temporary"}]`}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			base := existing("one")
			base.Metadata = nil
			members := make([]canonicalbatch.Member, len(tc.changes))
			for i, change := range tc.changes {
				members[i] = canonicalbatch.Member{Operation: "updateBead", Bead: operand("one"), ExpectedRevision: base.Revision, MetadataChange: raw(change)}
			}
			plan, err := Stage(Input{Scope: testScope, Current: map[string]Resource{base.ID: base}, Members: members})
			if err != nil {
				t.Fatal(err)
			}
			outcome := plan.Outcomes[0]
			if outcome.Changed || outcome.Outcome != "unchanged" || !reflect.DeepEqual(*outcome.After, base) || !reflect.DeepEqual(plan.Final[base.ID], base) {
				t.Fatalf("metadata patch materialized absent object or changed recorded envelope: %+v", outcome)
			}
			if outcome.After.Metadata != nil || outcome.After.Revision != base.Revision || !bytes.Equal(outcome.After.Envelope, base.Envelope) {
				t.Fatal("absent metadata, original revision or recorded envelope was lost")
			}
		})
	}
}

func TestAbsentMetadataRealEditsAndExplicitReplacementMaterialize(t *testing.T) {
	for _, tc := range []struct {
		name    string
		members []canonicalbatch.Member
		want    string
	}{
		{"real patch", []canonicalbatch.Member{{Operation: "updateBead", Bead: operand("one"), MetadataChange: raw(`[{"op":"add","path":"/reviewed","value":true}]`)}}, `{"reviewed":true}`},
		{"explicit replacement", []canonicalbatch.Member{{Operation: "updateBead", Bead: operand("one"), Metadata: raw(`{}`)}}, `{}`},
		{"explicit replacement survives cancelling patches", []canonicalbatch.Member{
			{Operation: "updateBead", Bead: operand("one"), Metadata: raw(`{}`)},
			{Operation: "updateBead", Bead: operand("one"), MetadataChange: raw(`[{"op":"add","path":"/temporary","value":true}]`)},
			{Operation: "updateBead", Bead: operand("one"), MetadataChange: raw(`[{"op":"remove","path":"/temporary"}]`)},
		}, `{}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			base := existing("one")
			base.Metadata = nil
			plan, err := Stage(Input{Scope: testScope, Current: map[string]Resource{base.ID: base}, Members: tc.members})
			if err != nil {
				t.Fatal(err)
			}
			outcome := plan.Outcomes[0]
			if !outcome.Changed || outcome.Outcome != "updated" || string(outcome.After.Metadata) != tc.want {
				t.Fatalf("metadata edit lost intentional object presence: %+v", outcome)
			}
			if base.Metadata != nil {
				t.Fatal("caller baseline metadata was modified")
			}
		})
	}
}
