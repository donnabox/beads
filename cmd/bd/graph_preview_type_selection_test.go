package main

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/configfile"
	"github.com/steveyegge/beads/internal/memorycatalog"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/typecatalog"
)

func TestGraphPreviewSelectedTypeKindAndRefinement(t *testing.T) {
	const scope = typesInventoryScope
	memory := inventoryDescriptor(t, graphstore.MemoryTypeURL(scope), "Memory", graph.KindBead)
	parent := inventoryDescriptor(t, scope+"types/future-parent", "Future parent", graph.KindBead, memory.ID())
	child := inventoryDescriptor(t, scope+"types/future-child", "Future child", graph.KindBead, parent.ID())
	link := inventoryDescriptor(t, scope+"types/future-link", "Future link", graph.KindLink)
	installed := []graph.TypeDescriptor{link, child, memory, parent}
	if err := graphPreviewValidateSelectedType(scope, installed, child.ID(), graph.KindBead); err != nil {
		t.Fatal(err)
	}
	if !graphPreviewTypeRefines(installed, child.ID(), memory.ID()) || !graphPreviewTypeRefines(installed, memory.ID(), memory.ID()) {
		t.Fatal("transitive refinement depends on a hardcoded Type list")
	}
	for _, test := range []struct {
		id   string
		kind graph.ResourceKind
	}{
		{child.ID(), graph.KindLink}, {link.ID(), graph.KindBead},
		{scope + "types/missing", graph.KindBead}, {"https://foreign.invalid/types/future-child", graph.KindBead},
	} {
		if err := graphPreviewValidateSelectedType(scope, installed, test.id, test.kind); err == nil {
			t.Fatalf("admitted wrong/uninstalled kind: %+v", test)
		}
	}
	if graphPreviewTypeRefines(installed, link.ID(), memory.ID()) || graphPreviewTypeRefines(installed, "missing", "missing") {
		t.Fatal("invented refinement from an unavailable definition")
	}
	cycle := []graph.TypeDescriptor{
		inventoryDescriptor(t, scope+"types/a", "A", graph.KindBead, scope+"types/b"),
		inventoryDescriptor(t, scope+"types/b", "B", graph.KindBead, scope+"types/a"),
	}
	if graphPreviewTypeRefines(cycle, cycle[0].ID(), memory.ID()) {
		t.Fatal("cycle invented a missing family")
	}
}

func TestGraphPreviewSelectedTypeRefusesIntrinsicsAndKeepsConformance(t *testing.T) {
	for _, name := range []string{"Type", "BeadType", "LinkType", "ConformsTo", "Refines", "refines", "conforms-to"} {
		d := inventoryDescriptor(t, typesInventoryScope+name, name, graph.KindBead)
		if err := graphPreviewValidateSelectedType(typesInventoryScope, []graph.TypeDescriptor{d}, d.ID(), graph.KindBead); err == nil {
			t.Errorf("intrinsic identity admitted %s", d.ID())
		}
		for _, id := range []string{typesInventoryScope + "types/" + name, "https://foreign.test/" + name} {
			unrelated := inventoryDescriptor(t, id, name, graph.KindBead)
			if graphPreviewIntrinsicType(typesInventoryScope, unrelated) {
				t.Errorf("unrelated whole identity classified by basename: %s", id)
			}
		}
	}
	conformance := inventoryDescriptor(t, typesInventoryScope+"conformance", "Conformance", graph.KindLink)
	if err := graphPreviewValidateSelectedType(typesInventoryScope, []graph.TypeDescriptor{conformance}, conformance.ID(), graph.KindLink); err != nil {
		t.Fatal(err)
	}
	renamed := inventoryDescriptor(t, typesInventoryScope+"types/not-the-name", "BeadType", graph.KindBead)
	if err := graphPreviewValidateSelectedType(typesInventoryScope, []graph.TypeDescriptor{renamed}, renamed.ID(), graph.KindBead); err != nil {
		t.Fatal("unrelated whole-ID was hidden by name", err)
	}
}

func TestGraphPreviewSelectedTypeKeepsNativeDecisionClassification(t *testing.T) {
	nativeDecision, err := graphstore.NativeIssueTypeURL(typesInventoryScope, "decision")
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct{ value, want string }{
		{"decision", nativeDecision},
		{"types/decision", typesInventoryScope + "types/decision"},
	} {
		got, err := graphPreviewCreateBeadType(graphCreateTypeCommand(t, []string{"--type=" + test.value}), typesInventoryScope)
		if err != nil || got != test.want {
			t.Fatalf("selector %q: %q, %v", test.value, got, err)
		}
	}
}

func TestGraphPreviewSelectedCreatePreservesCompleteProperties(t *testing.T) {
	input := `{"title":"Exact subjects","body":"Original","subjects":[{"uri":"https://example.invalid/beads/a","revision":"one"}],"extra":{"null":null,"items":[1,true,"雪"]}}`
	cmd := graphCreateTypeCommand(t, []string{"--type=evaluation", "--properties=" + input})
	raw, err := graphPreviewCreateMemoryPropertiesJSON(cmd, nil)
	if err != nil {
		t.Fatal(err)
	}
	var want, got map[string]json.RawMessage
	if err := json.Unmarshal([]byte(input), &want); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"subjects", "extra"} {
		canonicalWant, err := graph.CanonicalizeJSON(want[name])
		if err != nil {
			t.Fatal(err)
		}
		canonicalGot, err := graph.CanonicalizeJSON(got[name])
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(canonicalGot, canonicalWant) {
			t.Fatalf("complete property %s lost: %s", name, raw)
		}
	}
	for _, invalid := range []string{`{"title":"Good","body":null}`, `{"title":"Good","body":3}`, `{"title":"One","title":"Two"}`} {
		cmd := graphCreateTypeCommand(t, []string{"--type=evaluation", "--properties=" + invalid})
		if _, err := graphPreviewCreateMemoryPropertiesJSON(cmd, nil); err == nil {
			t.Fatalf("invalid complete input admitted: %s", invalid)
		}
	}
}

func TestGraphPreviewSelectedTypeFlowsIntoOrderedPatch(t *testing.T) {
	previous := graphPreviewConfig
	graphPreviewConfig = &configfile.Config{GraphScopeURL: typesInventoryScope}
	t.Cleanup(func() { graphPreviewConfig = previous })
	cmd := memoryPropertiesPatchCommand(t)
	cmd.Flags().String("type", "", "")
	if err := cmd.ParseFlags([]string{"--type=policy", `--patch=[{"op":"add","path":"/extra","value":true}]`}); err != nil {
		t.Fatal(err)
	}
	request, err := graphPreviewMemoryPropertiesPatchRequest(cmd, "beads/a")
	if err != nil || request.TypeURL != typesInventoryScope+"policy" || !strings.Contains(string(request.Patch), "extra") {
		t.Fatalf("retype/patch request lost intent: %+v %v", request, err)
	}
}

// Build valid partial inventory descriptors from the selected declaration API,
// rather than copying an obsolete six-Type fixture or inventing permissive
// property schemas. Production validation stays in the checked provider.
func selectedTypeTestDescriptors(t *testing.T, scope string) []graph.TypeDescriptor {
	t.Helper()
	input := memorycatalog.Inputs{
		Bindings: map[memorycatalog.Symbol]memorycatalog.Binding{},
		Issue:    typecatalog.Reference{URI: graphstore.IssueTypeURL(scope)},
	}
	// This fixture models the complete selected Memory declaration, independently
	// of the workspace's fresh installation and native adapter inventories.
	for _, symbol := range []memorycatalog.Symbol{
		memorycatalog.Memory, memorycatalog.Evaluation, memorycatalog.Constraint,
		memorycatalog.Policy,
		memorycatalog.Requirement, memorycatalog.Scenario, memorycatalog.Specification,
		memorycatalog.References, memorycatalog.SupersededBy, memorycatalog.Requires,
		memorycatalog.Defines, memorycatalog.SubjectTo, memorycatalog.Conformance,
		memorycatalog.Addresses, memorycatalog.Satisfies, memorycatalog.Violates,
	} {
		name := string(symbol)
		input.Bindings[symbol] = memorycatalog.Binding{
			Type: typecatalog.Reference{URI: scope + name}, SchemaBase: scope + "schemas/" + name,
		}
	}
	projection, err := memorycatalog.New(input)
	if err != nil {
		t.Fatal(err)
	}
	var descriptors []graph.TypeDescriptor
	for _, definition := range projection.Definitions() {
		var parents []string
		for _, parent := range definition.Parents {
			parents = append(parents, parent.URI)
		}
		descriptors = append(descriptors, inventoryDescriptor(t, definition.Binding.Type.URI, string(definition.Symbol), definition.Kind, parents...))
	}
	descriptors = append(descriptors, inventoryDescriptor(t, graphstore.IssueTypeURL(scope), "Issue", graph.KindBead), inventoryDescriptor(t, graphstore.DependencyTypeURL(scope), "Dependency", graph.KindLink))
	for _, name := range graphstore.FreshTypeNames() {
		id := scope + name
		for _, role := range []graph.ResourceKind{graph.KindBead, graph.KindLink} {
			base, classification, native := graphstore.NativeTypeClassification(scope, id, role)
			if native {
				descriptors = append(descriptors, inventoryDescriptor(t, id, classification, role, base))
				break
			}
		}
	}

	return descriptors
}

func TestGraphPreviewSelectedGenericProjectionUsesInstalledRefinements(t *testing.T) {
	snapshot := genericTestSnapshot(t, genericTestEdge{"ab", "a", "b"})
	snapshot.Types = selectedTypeTestDescriptors(t, genericTestScope)
	record := snapshot.Records[0].(graphstore.Record)
	record.Type = genericTestScope + "policy"
	snapshot.Records[0] = record
	got, err := projectGraphGeneric(snapshot, genericTestScope, genericTestInput("a", "out", 1))
	if err != nil || len(got.Nodes) != 2 || got.Nodes[0].Type != record.Type {
		t.Fatalf("refined installed Bead rejected or reshaped: %+v %v", got, err)
	}
}
