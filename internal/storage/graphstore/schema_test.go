package graphstore

import (
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
	if parsed.ID() != scope+"types/preview-related-v2" {
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

func TestRelationshipDescriptorsKeepLegacyIssueTypeImmutable(t *testing.T) {
	const scope = "https://example.test/relationships/"
	legacy, err := issueDescriptor(scope)
	if err != nil {
		t.Fatal(err)
	}
	if legacy.ID() != IssueTypeURL(scope) || len(legacy.OwnsOutgoing()) != 1 {
		t.Fatalf("legacy Issue descriptor changed: %s, %v", legacy.ID(), legacy.OwnsOutgoing())
	}
	if _, ok := legacy.Owns(DependencyTypeURL(scope)); !ok {
		t.Fatal("legacy Issue lost its blocking Link ownership")
	}
	newIssue, err := relationshipIssueDescriptor(scope)
	if err != nil {
		t.Fatal(err)
	}
	if newIssue.ID() == legacy.ID() || len(newIssue.OwnsOutgoing()) != len(relationshipDependencyTypes) {
		t.Fatalf("relationship Issue needs a distinct identity and all native Links: %s, %v", newIssue.ID(), newIssue.OwnsOutgoing())
	}
	seen := make(map[string]types.DependencyType)
	for _, native := range []types.DependencyType{types.DepBlocks, types.DepParentChild, types.DepDuplicates, types.DepSupersedes} {
		id, ok := relationshipDependencyTypeURL(scope, native)
		if !ok || id == DependencyTypeURL(scope) {
			t.Fatalf("missing distinct Link Type for %q: %q", native, id)
		}
		if got, ok := nativeRelationshipDependencyType(scope, id); !ok || got != native {
			t.Fatalf("Link Type %q maps back to %q, %t; want %q", id, got, ok, native)
		}
		if previous, duplicate := seen[id]; duplicate {
			t.Fatalf("native %q and %q share Link Type %q", previous, native, id)
		}
		seen[id] = native
		if _, owns := newIssue.Owns(id); !owns {
			t.Fatalf("Issue v3 does not own %q", id)
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
			if !present || len(constraint.ConformsTo()) != 1 || constraint.ConformsTo()[0] != newIssue.ID() || constraint.EffectiveExternal() != graph.ExternalNone {
				t.Fatalf("%q must require one live local Issue v3 endpoint", id)
			}
		}
	}
	if _, ok := relationshipDependencyTypeURL(scope, types.DepRelated); ok {
		t.Fatal("informational native relationship acquired an Issue-owned Link Type")
	}
	if _, ok := nativeRelationshipDependencyType(scope, scope+"types/preview-related-v2"); ok {
		t.Fatal("informational Link mapped to a native Dependency")
	}
	want := []string{"memory", "issue", "dependency", "related", "example-follows", "example-cites", "issue-v3", "blocks-v2", "parent-child", "duplicates", "supersedes"}
	fresh := FreshTypeNames()
	if len(fresh) != len(want) {
		t.Fatalf("fresh installation has %d Types, want %d", len(fresh), len(want))
	}
	for index, name := range want {
		if fresh[index] != name {
			t.Fatalf("Type %d is %q, want %q", index, fresh[index], name)
		}
	}
}

func TestInstalledIssueDependencyProjectionMapping(t *testing.T) {
	const scope = "https://example.test/relationships/"
	legacy := IssueTypeURL(scope)
	newIssue := RelationshipIssueTypeURL(scope)
	if got, ok := nativeDependencyTypeURLForIssueType(scope, legacy, types.DepBlocks); !ok || got != DependencyTypeURL(scope) {
		t.Fatalf("legacy blocking Type = %q, %t", got, ok)
	}
	if got, ok := nativeDependencyTypeForIssueType(scope, legacy, DependencyTypeURL(scope)); !ok || got != types.DepBlocks {
		t.Fatalf("legacy blocking reverse mapping = %q, %t", got, ok)
	}
	for _, native := range []types.DependencyType{types.DepBlocks, types.DepParentChild, types.DepDuplicates, types.DepSupersedes} {
		want, _ := relationshipDependencyTypeURL(scope, native)
		if got, ok := nativeDependencyTypeURLForIssueType(scope, newIssue, native); !ok || got != want {
			t.Fatalf("v7 %s Type = %q, %t; want %q", native, got, ok, want)
		}
		if got, ok := nativeDependencyTypeForIssueType(scope, newIssue, want); !ok || got != native {
			t.Fatalf("v7 %s reverse mapping = %q, %t", want, got, ok)
		}
	}
	for _, tc := range []struct {
		issueType string
		native    types.DependencyType
	}{
		{legacy, types.DepParentChild},
		{legacy, types.DepDuplicates},
		{legacy, types.DepSupersedes},
		{newIssue, types.DepRelated},
		{scope + "types/uninstalled", types.DepBlocks},
	} {
		if got, ok := nativeDependencyTypeURLForIssueType(scope, tc.issueType, tc.native); ok || got != "" {
			t.Fatalf("unsupported %s/%s mapped to %q, %t", tc.issueType, tc.native, got, ok)
		}
	}
	if got, ok := nativeDependencyTypeForIssueType(scope, legacy, scope+"types/preview-parent-child-v1"); ok || got != "" {
		t.Fatalf("legacy store admitted v7 Link Type %q, %t", got, ok)
	}
	if got, ok := nativeDependencyTypeForIssueType(scope, newIssue, DependencyTypeURL(scope)); ok || got != "" {
		t.Fatalf("v7 store admitted legacy Link Type %q, %t", got, ok)
	}
}
