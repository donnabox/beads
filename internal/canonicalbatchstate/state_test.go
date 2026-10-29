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

func TestMutationPolicyObservesCanceledIntentWithDetachedInputs(t *testing.T) {
	base := existing("one")
	members := []canonicalbatch.Member{
		{Operation: "updateBead", Bead: operand("one"), PropertiesChange: raw(`[{"op":"replace","path":"/title","value":"temporary"}]`)},
		{Operation: "updateBead", Bead: operand("one"), PropertiesChange: raw(`[{"op":"replace","path":"/title","value":"initial"}]`)},
	}
	wantRaw := bytes.Clone(members[0].PropertiesChange)
	observed := 0
	input := Input{Scope: testScope, Current: map[string]Resource{base.ID: base}, Members: members}
	input.AdmitMutation = func(index int, member canonicalbatch.Member, before *Resource) error {
		if index != observed || before == nil || before.ID != base.ID {
			t.Fatalf("policy did not receive resolved predecessor: %d %+v", index, before)
		}
		if index == 1 && string(before.Properties) != `{"title":"temporary"}` {
			t.Fatalf("policy observed baseline instead of preceding staged state: %s", before.Properties)
		}
		observed++
		before.Properties[0], member.PropertiesChange[0] = '!', '!'
		return nil
	}
	plan, err := Stage(input)
	if err != nil || observed != 2 || len(plan.Outcomes) != 1 || plan.Outcomes[0].Changed || !bytes.Equal(plan.Final[base.ID].Envelope, base.Envelope) || !bytes.Equal(members[0].PropertiesChange, wantRaw) {
		t.Fatalf("policy arguments mutated input/state or lost canceled intent: %+v %v", plan, err)
	}
	denied := errors.New("domain policy refused authored intent")
	input.AdmitMutation = func(_ int, _ canonicalbatch.Member, _ *Resource) error { return denied }
	if plan, err := Stage(input); !errors.Is(err, denied) || !reflect.DeepEqual(plan, Plan{}) {
		t.Fatalf("policy failure returned partial success: %+v %v", plan, err)
	}
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

func TestResolvedMutationAdmissionSeesResolvedCreates(t *testing.T) {
	base := existing("baseline")
	created := createBead("ignored", "new-bead")
	created.ID = ""
	created.Metadata = raw(`{"initialized":true}`)
	linked := createLink("new-link", "ignored", "baseline")
	linked.Source = canonicalbatch.Operand{Binding: "new-bead"}
	linked.Target.Revision = base.Revision
	linked.Properties = raw(`{"note":"authored"}`)
	linked.Metadata = raw(`{"link":true}`)
	var order []int
	resolved := 0
	plan, err := Stage(Input{
		Scope: testScope, Current: map[string]Resource{base.ID: base},
		Members: []canonicalbatch.Member{created, linked},
		Allocate: func(index int, kind graphops.ResourceKind) (string, error) {
			if index != 0 || kind != graphops.KindBead {
				t.Fatal("unexpected allocation")
			}
			return testScope + "allocated", nil
		},
		ResolveType: func(ref canonicalbatch.Reference) (canonicalbatch.Reference, error) {
			ref.Revision = "resolved-type-revision"
			return ref, nil
		},
		AdmitMutation: func(index int, _ canonicalbatch.Member, _ *Resource) error {
			order = append(order, index*2)
			return nil
		},
		AdmitResolvedMutation: func(index int, member canonicalbatch.Member, before, after *Resource) error {
			order = append(order, index*2+1)
			resolved++
			if before != nil || after == nil || after.Revision != "" || after.Type.Revision != "resolved-type-revision" {
				t.Fatalf("unresolved create: before=%+v after=%+v", before, after)
			}
			if index == 0 {
				if after.ID != testScope+"allocated" || after.Kind != graphops.KindBead || string(after.Properties) != `{"title":"new"}` || string(after.Metadata) != `{"initialized":true}` {
					t.Fatalf("lost resolved Bead: %+v", after)
				}
			} else if index == 1 {
				if member.Source.Binding != "new-bead" || after.ID != testScope+"new-link" || after.Kind != graphops.KindLink || after.Source != reference("allocated") || after.Target != (canonicalbatch.Reference{URI: base.ID, Revision: base.Revision}) || string(after.Properties) != `{"note":"authored"}` || string(after.Metadata) != `{"link":true}` {
					t.Fatalf("lost authored/resolved Link distinction: member=%+v after=%+v", member, after)
				}
			} else {
				t.Fatalf("unexpected resolved index %d", index)
			}
			return nil
		},
	})
	if err != nil || resolved != 2 || !reflect.DeepEqual(order, []int{0, 1, 2, 3}) || len(plan.Outcomes) != 2 {
		t.Fatalf("resolved create admission/order: %+v %v order=%v", plan, err, order)
	}
}

func TestResolvedMutationAdmissionDetachesAllArguments(t *testing.T) {
	base := existing("one")
	base.Metadata = raw(`{"keep":true}`)
	members := []canonicalbatch.Member{
		{Operation: "updateBead", Bead: operand("one"), Properties: raw(`{"title":"first"}`), Metadata: raw(`{"keep":true,"phase":1}`), RevisionProvenance: raw(`{"source":"caller"}`)},
		{Operation: "updateBead", Bead: operand("one"), PropertiesChange: raw(`[{"op":"replace","path":"/title","value":"final"}]`), MetadataChange: raw(`[{"op":"replace","path":"/phase","value":2}]`)},
	}
	memberBytes, err := json.Marshal(members)
	if err != nil {
		t.Fatal(err)
	}
	baselineBytes, err := json.Marshal(base)
	if err != nil {
		t.Fatal(err)
	}
	observed := 0
	plan, err := Stage(Input{
		Scope: testScope, Current: map[string]Resource{base.ID: base}, Members: members,
		AdmitResolvedMutation: func(index int, member canonicalbatch.Member, before, after *Resource) error {
			if index != observed || before == nil || after == nil || before.ID != base.ID || after.ID != base.ID {
				t.Fatalf("missing resolved update: %d %+v %+v", index, before, after)
			}
			if index == 1 && (string(before.Properties) != `{"title":"first"}` || string(before.Metadata) != `{"keep":true,"phase":1}` || string(after.Properties) != `{"title":"final"}` || string(after.Metadata) != `{"keep":true,"phase":2}`) {
				t.Fatalf("resolved callback lost prior staged content: %+v %+v", before, after)
			}
			observed++
			for _, data := range []json.RawMessage{member.Properties, member.Metadata, member.PropertiesChange, member.MetadataChange, member.RevisionProvenance, member.Query, before.Properties, before.Metadata, before.Envelope, after.Properties, after.Metadata, after.Envelope} {
				if len(data) != 0 {
					data[0] = '!'
				}
			}
			member.Bead.URI, member.Type.URI = "changed member target", "changed member Type"
			before.ID, after.ID, after.Type.URI = "changed predecessor", "changed postimage", "changed postimage Type"
			return nil
		},
	})
	if err != nil || observed != 2 {
		t.Fatalf("detached update admission: %+v %v", plan, err)
	}
	final := plan.Final[base.ID]
	if final.ID != base.ID || final.Type != base.Type || string(final.Properties) != `{"title":"final"}` || string(final.Metadata) != `{"keep":true,"phase":2}` || !bytes.Equal(final.Envelope, base.Envelope) || string(plan.Provenance[base.ID][0].Value) != `{"source":"caller"}` {
		t.Fatalf("callback mutated staged state: %+v", plan)
	}
	afterMemberBytes, err := json.Marshal(members)
	if err != nil {
		t.Fatal(err)
	}
	afterBaselineBytes, err := json.Marshal(base)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(memberBytes, afterMemberBytes) || !bytes.Equal(baselineBytes, afterBaselineBytes) {
		t.Fatal("resolved callback mutated caller input")
	}
}

func TestResolvedMutationAdmissionRejectsCreateBeforeCancellation(t *testing.T) {
	denied := errors.New("resolved creation policy refused")
	members := []canonicalbatch.Member{
		createBead("prefix", ""), createBead("denied", ""),
		{Operation: "deleteBead", Bead: operand("denied")},
	}
	observed := 0
	plan, err := Stage(Input{
		Scope: testScope, Members: members,
		AdmitResolvedMutation: func(index int, _ canonicalbatch.Member, before, after *Resource) error {
			observed++
			if before != nil || after == nil || after.ID == "" || string(after.Properties) != `{"title":"new"}` {
				t.Fatalf("missing resolved creation before refusal: %+v %+v", before, after)
			}
			if index == 1 {
				return denied
			}
			return nil
		},
	})
	var memberErr *MemberError
	if !errors.Is(err, denied) || !errors.As(err, &memberErr) || memberErr.OperationIndex != 1 || observed != 2 || !reflect.DeepEqual(plan, Plan{}) {
		t.Fatalf("rejected canceled creation returned partial plan: %+v %v observed=%d", plan, err, observed)
	}
}

func TestResolvedMutationAdmissionSeesCanceledCreateAndDeletion(t *testing.T) {
	members := []canonicalbatch.Member{
		createBead("one", "new"),
		{Operation: "updateBead", Bead: canonicalbatch.Operand{Binding: "new"}, PropertiesChange: raw(`[{"op":"replace","path":"/title","value":"edited"}]`)},
		{Operation: "deleteBead", Bead: canonicalbatch.Operand{Binding: "new"}},
	}
	var observed []int
	admit := func(index int, _ canonicalbatch.Member, before, after *Resource) error {
		observed = append(observed, index)
		if index == 0 {
			if before != nil || after == nil || after.ID != testScope+"one" {
				t.Fatalf("missing creation: %+v %+v", before, after)
			}
		} else if index == 1 {
			if before == nil || after == nil || string(before.Properties) != `{"title":"new"}` || string(after.Properties) != `{"title":"edited"}` {
				t.Fatalf("missing staged edit: %+v %+v", before, after)
			}
		} else if index == 2 {
			if before == nil || before.ID != testScope+"one" || string(before.Properties) != `{"title":"edited"}` || after != nil {
				t.Fatalf("deletion did not receive staged predecessor and nil postimage: %+v %+v", before, after)
			}
		}
		return nil
	}
	plan, err := Stage(Input{Scope: testScope, Members: members, AdmitResolvedMutation: admit})
	if err != nil || !reflect.DeepEqual(observed, []int{0, 1, 2}) || len(plan.Final) != 0 || len(plan.Outcomes) != 1 || plan.Outcomes[0].Outcome != "allocated" || !reflect.DeepEqual(plan.ConsumedIDs, []string{testScope + "one"}) {
		t.Fatalf("canceled create skipped resolved admission: %+v %v observed=%v", plan, err, observed)
	}
	denied := errors.New("resolved deletion refused")
	plan, err = Stage(Input{Scope: testScope, Members: members, AdmitResolvedMutation: func(index int, member canonicalbatch.Member, before, after *Resource) error {
		if err := admit(index, member, before, after); err != nil {
			return err
		}
		if after == nil {
			return denied
		}
		return nil
	}})
	var memberErr *MemberError
	if !errors.Is(err, denied) || !errors.As(err, &memberErr) || memberErr.OperationIndex != 2 || !reflect.DeepEqual(plan, Plan{}) {
		t.Fatalf("resolved deletion refusal returned partial plan: %+v %v", plan, err)
	}
}

func TestResolvedMutationAdmissionRunsAfterSuccessfulResolutionOnly(t *testing.T) {
	base := existing("source")
	member := createLink("link", "source", "missing")
	beforeCalled, resolvedCalled := false, false
	plan, err := Stage(Input{
		Scope: testScope, Current: map[string]Resource{base.ID: base}, Members: []canonicalbatch.Member{member},
		AdmitMutation: func(_ int, _ canonicalbatch.Member, before *Resource) error {
			beforeCalled = true
			if before != nil {
				t.Fatal("create received predecessor")
			}
			return nil
		},
		AdmitResolvedMutation: func(_ int, _ canonicalbatch.Member, _, _ *Resource) error {
			resolvedCalled = true
			return nil
		},
	})
	if !errors.Is(err, ErrNotFound) || !beforeCalled || resolvedCalled || !reflect.DeepEqual(plan, Plan{}) {
		t.Fatalf("failed endpoint resolution reached resolved observer: %+v %v before=%t resolved=%t", plan, err, beforeCalled, resolvedCalled)
	}
}

func TestResolvedMutationAdmissionOrderedSetFanout(t *testing.T) {
	one, two := existing("one"), existing("two")
	members := []canonicalbatch.Member{
		selectMember("picked", beadScan, ""),
		{Operation: "updateBead", Bead: canonicalbatch.Operand{Binding: "picked"}, PropertiesChange: raw(`[{"op":"replace","path":"/title","value":"updated"}]`)},
		{Operation: "deleteBead", Bead: canonicalbatch.Operand{Binding: "picked"}},
	}
	var observed []Mapping
	callback := func(index int, member canonicalbatch.Member, before, after *Resource) error {
		if before == nil || member.Bead.URI != before.ID || member.Bead.Binding != "" {
			t.Fatalf("set member did not resolve its scalar target: %+v %+v", member, before)
		}
		observed = append(observed, Mapping{OperationIndex: index, ID: before.ID})
		if index == 1 {
			if after == nil || string(before.Properties) != `{"title":"initial"}` || string(after.Properties) != `{"title":"updated"}` {
				t.Fatalf("set update did not resolve content: %+v %+v", before, after)
			}
		} else if index == 2 {
			if string(before.Properties) != `{"title":"updated"}` || after != nil {
				t.Fatalf("set deletion lost staged update: %+v %+v", before, after)
			}
		} else {
			t.Fatalf("nonmutation reached observer: %d", index)
		}
		return nil
	}
	input := Input{Scope: testScope, Current: map[string]Resource{two.ID: two, one.ID: one}, Members: members, AdmitResolvedMutation: callback}
	plan, err := Stage(input)
	want := []Mapping{{OperationIndex: 1, ID: one.ID}, {OperationIndex: 1, ID: two.ID}, {OperationIndex: 2, ID: one.ID}, {OperationIndex: 2, ID: two.ID}}
	if err != nil || !reflect.DeepEqual(observed, want) || len(plan.Final) != 0 || len(plan.Outcomes) != 2 {
		t.Fatalf("resolved set fanout order: %+v %v observed=%+v", plan, err, observed)
	}
	denied := errors.New("second set target refused")
	input.AdmitResolvedMutation = func(index int, member canonicalbatch.Member, before, after *Resource) error {
		if err := callback(index, member, before, after); err != nil {
			return err
		}
		if index == 1 && before.ID == two.ID {
			return denied
		}
		return nil
	}
	plan, err = Stage(input)
	var memberErr *MemberError
	if !errors.Is(err, denied) || !errors.As(err, &memberErr) || memberErr.OperationIndex != 1 || !reflect.DeepEqual(plan, Plan{}) || !bytes.Equal(one.Properties, raw(`{"title":"initial"}`)) || !bytes.Equal(two.Properties, raw(`{"title":"initial"}`)) {
		t.Fatalf("resolved set refusal leaked partial plan/input edits: %+v %v", plan, err)
	}
}
