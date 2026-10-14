package graphstore

import (
	"slices"
	"strings"
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
	if parsed.ID() != scope+"types/preview-related" {
		t.Fatal("corrected descriptor must have a fresh experimental identity")
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
		if external, present := endpoint.External(); !present || external != graph.ExternalNone {
			t.Fatal("preview must refuse external endpoints")
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

func TestFreshPreviewTypeIdentitiesAndOwnership(t *testing.T) {
	const scope = "https://example.test/relationships/"
	issue, err := issueDescriptor(scope)
	if err != nil {
		t.Fatal(err)
	}
	if issue.ID() != scope+"types/preview-issue" || len(issue.OwnsOutgoing()) != 4 {
		t.Fatalf("single Issue Type identity or ownership differs: %s, %v", issue.ID(), issue.OwnsOutgoing())
	}
	seen := map[string]bool{}
	for _, native := range []types.DependencyType{types.DepBlocks, types.DepParentChild, types.DepDuplicates, types.DepSupersedes} {
		id, ok := relationshipDependencyTypeURL(scope, native)
		if !ok || seen[id] || strings.Contains(id, "-v1") || strings.Contains(id, "-v2") || strings.Contains(id, "-v3") {
			t.Fatalf("native %q lacks one unsuffixed Type: %q, %t", native, id, ok)
		}
		seen[id] = true
		if _, owns := issue.Owns(id); !owns {
			t.Fatalf("Issue does not own %q", id)
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
			if !present || len(constraint.ConformsTo()) != 1 || constraint.ConformsTo()[0] != issue.ID() || constraint.EffectiveExternal() != graph.ExternalNone {
				t.Fatalf("%q must require one live local Issue endpoint", id)
			}
		}
	}
	if DependencyTypeURL(scope) != scope+"types/preview-blocks" ||
		MemoryTypeURL(scope) != scope+"types/preview-memory" ||
		RelatedTypeURL(scope) != scope+"types/preview-related" {
		t.Fatal("a built-in Type retained a versioned identity")
	}
	want := []string{"memory", "issue", "dependency", "related", "example-follows", "example-cites", "parent-child", "duplicates", "supersedes"}
	fresh := FreshTypeNames()
	if !slices.Equal(fresh, want) {
		t.Fatalf("fresh Type set = %v, want %v", fresh, want)
	}
}

func TestInstalledIssueDependencyProjectionMapping(t *testing.T) {
	const scope = "https://example.test/relationships/"
	issue := IssueTypeURL(scope)
	for _, native := range []types.DependencyType{types.DepBlocks, types.DepParentChild, types.DepDuplicates, types.DepSupersedes} {
		want, _ := relationshipDependencyTypeURL(scope, native)
		if got, ok := nativeDependencyTypeURLForIssueType(scope, issue, native); !ok || got != want {
			t.Fatalf("%s Type = %q, %t; want %q", native, got, ok, want)
		}
		if got, ok := nativeDependencyTypeForIssueType(scope, issue, want); !ok || got != native {
			t.Fatalf("%s reverse mapping = %q, %t", want, got, ok)
		}
	}
	if got, ok := nativeDependencyTypeURLForIssueType(scope, issue, types.DepRelated); ok || got != "" {
		t.Fatalf("informational kind mapped as Dependency: %q, %t", got, ok)
	}
	if got, ok := nativeDependencyTypeForIssueType(scope, issue, RelatedTypeURL(scope)); ok || got != "" {
		t.Fatalf("informational Type mapped as Dependency: %q, %t", got, ok)
	}
}
