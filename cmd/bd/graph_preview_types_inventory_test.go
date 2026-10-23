package main

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
)

const typesInventoryScope = "https://example.invalid/inventory/"

func inventoryDescriptor(t *testing.T, id, name string, kind graph.ResourceKind, parents ...string) graph.TypeDescriptor {
	t.Helper()
	spec := graph.TypeDescriptorSpec{ID: id, Name: name, Describes: kind, ConformsTo: parents}
	if kind == graph.KindLink {
		endpoint := graph.EndpointConstraint{}
		spec.Source, spec.Target = &endpoint, &endpoint
	}
	descriptor, err := graph.NewTypeDescriptor(spec)
	if err != nil {
		t.Fatal(err)
	}
	return descriptor
}

func inventoryIDs(t *testing.T, items []json.RawMessage) []string {
	t.Helper()
	ids := []string{}
	for _, item := range items {
		var value struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(item, &value); err != nil {
			t.Fatal(err)
		}
		ids = append(ids, value.ID)
	}
	return ids
}

func TestGraphTypesInventoryUsesSelectedCatalog(t *testing.T) {
	parent := typesInventoryScope + "types/neutral"
	child := typesInventoryScope + "types/specialized"
	selected := []graph.TypeDescriptor{
		inventoryDescriptor(t, child, "Specialized", graph.KindLink, parent),
		inventoryDescriptor(t, typesInventoryScope+"types/custom-memory", "Custom Memory", graph.KindBead),
		inventoryDescriptor(t, parent, "Neutral", graph.KindLink),
	}
	original := []string{selected[0].ID(), selected[1].ID(), selected[2].ID()}
	for _, details := range []bool{false, true} {
		inventory, text, err := graphPreviewTypesInventory(typesInventoryScope, selected, nil, details, true)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(inventoryIDs(t, inventory.LinkTypes), []string{parent, child}) || len(inventory.BeadTypes) != 1 {
			t.Fatalf("did not enumerate provider data: %+v", inventory)
		}
		for _, want := range []string{"custom-memory", "neutral", "specialized", "Bead Types (1)", "Link Types (2)"} {
			if !strings.Contains(text, want) {
				t.Fatalf("missing %q: %s", want, text)
			}
		}
		for _, absent := range []string{"preview-memory", "example-cites", "example-follows", "Blocking Dependency", "Task Issue"} {
			if strings.Contains(text, absent) {
				t.Fatalf("invented absent Type %q: %s", absent, text)
			}
		}
		if strings.Index(text, "neutral") > strings.Index(text, "specialized") {
			t.Fatal("unstable identity order")
		}
		if details && !strings.Contains(text, `"conformsTo"`) {
			t.Fatal("stored parent metadata missing from details")
		}
	}
	if got := []string{selected[0].ID(), selected[1].ID(), selected[2].ID()}; !reflect.DeepEqual(got, original) {
		t.Fatal("sorted caller slice in place")
	}
	// A different membership and parent graph requires no renderer changes.
	selected = []graph.TypeDescriptor{inventoryDescriptor(t, child, "Independent", graph.KindLink)}
	inventory, text, err := graphPreviewTypesInventory(typesInventoryScope, selected, nil, false, true)
	if err != nil || len(inventory.BeadTypes) != 0 || len(inventory.LinkTypes) != 1 || strings.Contains(text, "Neutral") {
		t.Fatalf("catalog replacement failed: %s %v", text, err)
	}
}

func TestGraphTypesInventoryHidesIntrinsicMetatypes(t *testing.T) {
	selected := []graph.TypeDescriptor{
		inventoryDescriptor(t, typesInventoryScope+"types/conformance", "Conformance", graph.KindLink),
	}
	for _, name := range []string{"Type", "BeadType", "LinkType", "ConformsTo", "Refines"} {
		selected = append(selected, inventoryDescriptor(t, typesInventoryScope+"types/"+name, name, graph.KindBead))
	}
	// Identity aliases cannot expose the reserved relationship by renaming it.
	selected = append(selected, inventoryDescriptor(t, typesInventoryScope+"types/conforms-to", "An alias", graph.KindLink))
	selected = append(selected, inventoryDescriptor(t, typesInventoryScope+"types/refines", "Another alias", graph.KindLink))
	for _, details := range []bool{false, true} {
		inventory, text, err := graphPreviewTypesInventory(typesInventoryScope, selected, nil, details, true)
		if err != nil {
			t.Fatal(err)
		}
		if len(inventory.BeadTypes) != 0 || len(inventory.LinkTypes) != 1 || !strings.Contains(text, "Conformance") {
			t.Fatalf("wrong visibility: %s", text)
		}
		encoded, err := json.Marshal(inventory)
		if err != nil {
			t.Fatal(err)
		}
		for _, hidden := range []string{"BeadType", "LinkType", "ConformsTo", "Refines", "An alias", "Another alias"} {
			if strings.Contains(text, hidden) || strings.Contains(string(encoded), hidden) {
				t.Fatalf("leaked %q", hidden)
			}
		}
	}
}

func TestGraphTypesInventoryPreservesIssueAndDependencyClassifications(t *testing.T) {
	selected := []graph.TypeDescriptor{
		inventoryDescriptor(t, graphstore.IssueTypeURL(typesInventoryScope), "Issue", graph.KindBead),
		inventoryDescriptor(t, graphstore.DependencyTypeURL(typesInventoryScope), "Dependency", graph.KindLink),
		inventoryDescriptor(t, graphstore.RelatedTypeURL(typesInventoryScope), "Related Memory", graph.KindLink),
	}
	inventory, text, err := graphPreviewTypesInventory(typesInventoryScope, selected, []string{"team-review"}, false, true)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range coreWorkTypes {
		if !strings.Contains(text, string(entry.Type)) {
			t.Fatalf("missing Issue classification %s", entry.Type)
		}
	}
	for _, entry := range systemWorkTypes() {
		if !strings.Contains(text, entry.Name) {
			t.Fatalf("missing system classification %s", entry.Name)
		}
	}
	for _, kind := range types.WellKnownDependencyTypes() {
		if !strings.Contains(text, string(kind)) {
			t.Fatalf("missing Dependency kind %s", kind)
		}
	}
	for _, want := range []string{"team-review", "Team Review Issue", "Blocking Dependency", "Parent/Child Dependency", "references"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing preserved selector %q: %s", want, text)
		}
	}
	if len(inventory.BeadTypes) != 1 || len(inventory.LinkTypes) != 2 {
		t.Fatal("classification labels became invented nominal definitions")
	}
	details, detailText, err := graphPreviewTypesInventory(typesInventoryScope, selected, []string{"team-review"}, true, true)
	if err != nil || len(details.BeadTypes) != 1 || strings.Contains(detailText, "team-review") {
		t.Fatalf("details should describe nominal definitions: %s %v", detailText, err)
	}
}

func TestGraphTypesInventoryEmptyCatalogIsEmpty(t *testing.T) {
	for _, details := range []bool{false, true} {
		inventory, text, err := graphPreviewTypesInventory(typesInventoryScope, nil, []string{"configured-but-no-issue-definition"}, details, true)
		if err != nil || inventory.BeadTypes == nil || inventory.LinkTypes == nil || len(inventory.BeadTypes) != 0 || len(inventory.LinkTypes) != 0 {
			t.Fatalf("empty inventory: %+v %v", inventory, err)
		}
		if strings.Contains(text, "configured-but") || !strings.Contains(text, "Bead Types (0)") || !strings.Contains(text, "Link Types (0)") {
			t.Fatalf("invented catalog: %s", text)
		}
	}
}

// Nominal Type selectors keep their namespace so Memory Decision cannot be
// confused with the existing native Issue decision classification.
func TestGraphTypesInventoryKeepsNominalSelectorsAndAbstractDefinitions(t *testing.T) {
	issue := inventoryDescriptor(t, graphstore.IssueTypeURL(typesInventoryScope), "Issue", graph.KindBead)
	decision := inventoryDescriptor(t, typesInventoryScope+"types/decision", "Decision", graph.KindBead)
	endpoint := graph.EndpointConstraint{}
	conformance, err := graph.NewTypeDescriptor(graph.TypeDescriptorSpec{
		ID: typesInventoryScope + "types/conformance", Name: "Conformance", Describes: graph.KindLink,
		Description: "Abstract: cannot be used as a declared Type.", Source: &endpoint, Target: &endpoint,
	})
	if err != nil {
		t.Fatal(err)
	}
	selected := []graph.TypeDescriptor{decision, conformance, issue}
	var baseline []byte
	for _, details := range []bool{false, true} {
		inventory, text, err := graphPreviewTypesInventory(typesInventoryScope, selected, nil, details, true)
		if err != nil {
			t.Fatal(err)
		}
		for _, selector := range []string{"types/decision", "types/conformance"} {
			if !strings.Contains(text, selector) {
				t.Fatalf("missing disambiguated nominal selector %q: %s", selector, text)
			}
		}
		if details && !strings.Contains(text, "Abstract: cannot be used as a declared Type.") {
			t.Fatal("abstract definition disappeared from details")
		}
		encoded, err := json.Marshal(inventory)
		if err != nil {
			t.Fatal(err)
		}
		if baseline == nil {
			baseline = encoded
		} else if !reflect.DeepEqual(baseline, encoded) {
			t.Fatal("details changed the installed JSON inventory")
		}
		if !reflect.DeepEqual(inventoryIDs(t, inventory.LinkTypes), []string{conformance.ID()}) {
			t.Fatal("abstract public definition was hidden")
		}
	}
}

func TestGraphTypesAssessmentAvailability(t *testing.T) {
	var selected []graph.TypeDescriptor
	beadNames := map[string]bool{"memory": true, "evaluation": true, "constraint": true, "policy": true, "decision": true, "procedure": true, "requirement": true, "scenario": true, "specification": true, "issue": true}
	for _, name := range graphstore.FreshTypeNames() {
		kind := graph.KindLink
		if beadNames[name] {
			kind = graph.KindBead
		}
		selected = append(selected, inventoryDescriptor(t, typesInventoryScope+"types/"+name, name, kind))
	}
	for _, enabled := range []bool{false, true} {
		for _, details := range []bool{false, true} {
			inventory, text, err := graphPreviewTypesInventory(typesInventoryScope, selected, nil, details, enabled)
			if err != nil {
				t.Fatal(err)
			}
			want := 15
			if enabled {
				want = 20
			}
			if len(inventory.BeadTypes)+len(inventory.LinkTypes) != want {
				t.Fatal("wrong available Type census")
			}
			raw, err := json.Marshal(inventory)
			if err != nil {
				t.Fatal(err)
			}
			for _, name := range []string{"evaluation", "conformance", "addresses", "satisfies", "violates"} {
				id := typesInventoryScope + "types/" + name
				if strings.Contains(string(raw), id) != enabled || strings.Contains(text, "types/"+name) != enabled {
					t.Fatalf("inconsistent text/JSON availability %s", name)
				}
			}
		}
	}
	if len(selected) != 20 {
		t.Fatal("public filtering mutated installed inventory")
	}
}
