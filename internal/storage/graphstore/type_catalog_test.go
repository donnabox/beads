package graphstore

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/memorycatalog"
	"github.com/steveyegge/beads/internal/storage"
)

func TestSelectedTypeInventoryMatchesCompleteCatalog(t *testing.T) {
	const scope = "https://example.test/selected/"
	catalog, err := newSelectedTypeCatalog(scope)
	if err != nil {
		t.Fatal(err)
	}
	byID := map[string]graph.TypeDescriptor{}
	for _, definition := range previewTypeDefinitions() {
		descriptor, err := definition.build(scope)
		if err != nil {
			t.Fatalf("%s: %v", definition.name, err)
		}
		if _, duplicate := byID[descriptor.ID()]; duplicate {
			t.Fatalf("duplicate installed identity %s", descriptor.ID())
		}
		if strings.HasPrefix(definition.name, "example-") || strings.HasPrefix(definition.name, "preview-") {
			t.Fatalf("obsolete selected Type: %s", definition.name)
		}
		if definition.name == "related" {
			t.Fatal("legacy Related spelling installed a separate Type")
		}
		byID[descriptor.ID()] = descriptor
	}
	if len(byID) != 20 {
		t.Fatalf("installed %d Types, want 18 selected Memory Types plus native Issue/Dependency", len(byID))
	}
	for _, declaration := range catalog.projection.Definitions() {
		descriptor, present := byID[declaration.Binding.Type.URI]
		if !present || descriptor.Describes() != declaration.Kind {
			t.Fatalf("missing or wrong category for %s", declaration.Symbol)
		}
		parents := descriptor.ConformsTo()
		if len(parents) != len(declaration.Parents) {
			t.Fatalf("%s parent count differs", declaration.Symbol)
		}
		for _, parent := range declaration.Parents {
			found := false
			for _, id := range parents {
				found = found || id == parent.URI
			}
			if !found {
				t.Fatalf("%s missing parent %s", declaration.Symbol, parent.URI)
			}
		}
		if declaration.Kind == graph.KindLink && !IsInformationalTypeURL(scope, descriptor.ID()) {
			t.Fatalf("selected Link %s is unreachable by generic Link adapter", declaration.Symbol)
		}
		if declaration.Kind == graph.KindBead && len(descriptor.OwnsOutgoing()) != 0 {
			t.Fatalf("%s inventory invented finite ownership limits", declaration.Symbol)
		}
		if declaration.Abstract {
			description, _ := descriptor.Description()
			if !strings.Contains(description, "Abstract") {
				t.Fatalf("abstract %s is not identified in partial details", declaration.Symbol)
			}
		}
	}
	conformance, ok := catalog.projection.Definition(memorycatalog.Conformance)
	if !ok || !conformance.Abstract {
		t.Fatal("Conformance must be present and abstract")
	}
	if _, present := byID[conformance.Binding.Type.URI]; !present {
		t.Fatal("abstract Conformance must remain visible in inventory")
	}
	for _, name := range []string{"type", "bead-type", "link-type", "conforms-to", "refines", "Type", "BeadType", "LinkType", "ConformsTo", "Refines", "example-follows", "example-cites", "preview-memory", "preview-issue", "preview-dependency", "preview-related"} {
		if _, present := byID[scope+"types/"+name]; present {
			t.Fatalf("intrinsic/obsolete %s leaked into external catalog", name)
		}
	}
	if RelatedTypeURL(scope) != scope+"types/references" || IsInformationalTypeURL(scope, DependencyTypeURL(scope)) {
		t.Fatal("legacy Related alias or native Dependency separation differs")
	}
}

func TestSelectedReferencePropertiesUseSharedURISyntax(t *testing.T) {
	const scope = "https://example.test/selected/"
	catalog, err := newSelectedTypeCatalog(scope)
	if err != nil {
		t.Fatal(err)
	}
	for _, authority := range []any{"urn:example:☃", "https://example.test/a|b", map[string]string{"uri": "urn:example:☃", "revision": "r1"}} {
		raw, err := json.Marshal(map[string]any{"title": "Policy", "body": "", "authority": authority})
		if err != nil {
			t.Fatal(err)
		}
		if err := catalog.validateProperties(scope+"types/policy", raw, graph.KindBead); !errors.Is(err, storage.ErrValidation) {
			t.Fatalf("invalid authority admitted: %v", err)
		}
	}
}
