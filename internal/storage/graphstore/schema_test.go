package graphstore

import (
	"slices"
	"testing"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/types"
)

func TestInformationalDescriptorEndpointSemantics(t *testing.T) {
	const scope = "https://example.test/descriptor/"
	descriptor, err := relatedDescriptor(scope)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := graph.ParseTypeDescriptor(descriptor.CanonicalJSON())
	if err != nil {
		t.Fatal(err)
	}
	if parsed.ID() != scope+"types/references" {
		t.Fatal("legacy Related spelling must resolve to References")
	}
	source, hasSource := parsed.Source()
	target, hasTarget := parsed.Target()
	if !hasSource || !hasTarget {
		t.Fatal("Link descriptor must expose endpoint constraints")
	}
	for _, endpoint := range []graph.EndpointConstraint{source, target} {
		if required := endpoint.ConformsTo(); len(required) != 0 {
			t.Fatalf("conformsTo requires ALL listed Types, so a list of alternatives rejects mixed endpoints: %v", required)
		}
		if external := endpoint.EffectiveExternal(); external != graph.ExternalOpaque {
			t.Fatal("partial inventory must not pretend to encode correlated endpoint rules")
		}
	}
	// The blocking adapter remains specifically constrained to Issue endpoints.
	blocking, err := dependencyDescriptor(scope)
	if err != nil {
		t.Fatal(err)
	}
	source, _ = blocking.Source()
	target, _ = blocking.Target()
	for _, endpoint := range []graph.EndpointConstraint{source, target} {
		required := endpoint.ConformsTo()
		if len(required) != 1 || required[0] != IssueTypeURL(scope) {
			t.Fatalf("blocking Dependency must still require Issue: %v", required)
		}
	}
}

func TestDependencyDescriptorUsesOneNativeType(t *testing.T) {
	const scope = "https://example.test/relationships/"
	if IssueTypeURL(scope) != scope+"types/issue" ||
		MemoryTypeURL(scope) != scope+"types/memory" ||
		RelatedTypeURL(scope) != scope+"types/references" ||
		DependencyTypeURL(scope) != scope+"types/dependency" {
		t.Fatal("a built-in Type has a duplicate or versioned identity")
	}
	want := []string{"memory", "evaluation", "constraint", "policy", "decision", "procedure", "requirement", "scenario", "specification", "references", "superseded-by", "requires", "defines", "subject-to", "conformance", "addresses", "satisfies", "violates", "issue", "dependency"}
	if got := FreshTypeNames(); !slices.Equal(got, want) {
		t.Fatalf("fresh Type set = %v, want %v", got, want)
	}
	issue, err := issueDescriptor(scope)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := issue.Owns(DependencyTypeURL(scope)); !ok {
		t.Fatal("Issue must own the one native Dependency Link Type")
	}
	for _, native := range []types.DependencyType{types.DepBlocks, types.DepParentChild, types.DepDuplicates, types.DepSupersedes, "tracks"} {
		id, ok := relationshipDependencyTypeURL(scope, native)
		if !ok || id != DependencyTypeURL(scope) {
			t.Fatalf("native %q must use the one Dependency Type: %q, %t", native, id, ok)
		}
		link, err := relationshipDependencyDescriptor(scope, native)
		if err != nil {
			t.Fatal(err)
		}
		if link.ID() != id {
			t.Fatalf("descriptor ID %q, want %q", link.ID(), id)
		}
		for _, endpoint := range []func() (graph.EndpointConstraint, bool){link.Source, link.Target} {
			constraint, present := endpoint()
			if !present || len(constraint.ConformsTo()) != 1 || constraint.ConformsTo()[0] != IssueTypeURL(scope) || constraint.EffectiveExternal() != graph.ExternalNone {
				t.Fatalf("%q must require one live local Issue endpoint", id)
			}
		}
	}
	for _, invalid := range []types.DependencyType{"", "123456789012345678901234567890123"} {
		if _, ok := relationshipDependencyTypeURL(scope, invalid); ok {
			t.Fatalf("invalid kind %q acquired native Type", invalid)
		}
	}
	if _, ok := nativeRelationshipDependencyType(scope, scope+"types/preview-related-v2"); ok {
		t.Fatal("informational Link mapped to a native Dependency")
	}
}

func TestInstalledIssueDependencyProjectionMapping(t *testing.T) {
	const scope = "https://example.test/relationships/"
	for _, native := range []types.DependencyType{types.DepBlocks, types.DepParentChild, types.DepDuplicates, types.DepSupersedes, "arbitrary"} {
		if got, ok := nativeDependencyTypeURLForIssueType(scope, IssueTypeURL(scope), native); !ok || got != DependencyTypeURL(scope) {
			t.Fatalf("native %q Type = %q, %t", native, got, ok)
		}
	}
	if _, ok := nativeDependencyTypeURLForIssueType(scope, scope+"types/uninstalled", types.DepBlocks); ok {
		t.Fatal("uninstalled Issue Type admitted native Dependency")
	}
	if isNativeDependencyLinkType(scope, IssueTypeURL(scope), scope+"types/preview-parent-child-v1") {
		t.Fatal("distinct nominal kind Type remains admitted")
	}
}

func TestPreviewDependencyKindValidation(t *testing.T) {
	for _, tc := range []struct {
		input, want types.DependencyType
	}{
		{"", types.DepBlocks},
		{types.DepParentChild, types.DepParentChild},
		{"my-own-kind", "my-own-kind"},
	} {
		got, err := previewDependencyType(tc.input)
		if err != nil || got != tc.want {
			t.Fatalf("kind %q = %q, %v; want %q", tc.input, got, err, tc.want)
		}
	}
	for _, input := range []types.DependencyType{
		"123456789012345678901234567890123", "\xff",
	} {
		if _, err := previewDependencyType(input); err == nil {
			t.Fatalf("invalid Dependency kind %q was accepted", input)
		}
	}
}
