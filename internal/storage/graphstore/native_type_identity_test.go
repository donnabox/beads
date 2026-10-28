package graphstore

import (
	"strings"
	"testing"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/types"
	publicops "github.com/steveyegge/beads/issueops"
)

// Generated identities must expose the selected name, rather than an adapter
// implementation namespace. Literal caller-authored IDs are a separate path.
func TestGeneratedTypeIdentitiesUseMeaningfulScopeLocalNames(t *testing.T) {
	const scope = "https://example.test/work/"
	for _, tc := range []struct {
		name  string
		build func(string, string) (string, error)
	}{
		{"task", NativeIssueTypeURL},
		{"decision", NativeIssueTypeURL},
		{"blocks", NativeDependencyTypeURL},
		{"parent-child", NativeDependencyTypeURL},
	} {
		got, err := tc.build(scope, tc.name)
		if err != nil || got != scope+tc.name {
			t.Errorf("generated %s identity = %q, %v; want %q", tc.name, got, err, scope+tc.name)
		}
	}
	if got := MemoryTypeURL(scope); got != scope+"memory" {
		t.Errorf("Memory identity = %q; want %q", got, scope+"memory")
	}
}

func TestNativeConcreteTypeIdentity(t *testing.T) {
	const scope = "https://example.test/work/"
	for _, role := range []graph.ResourceKind{graph.KindBead, graph.KindLink} {
		values := []string{"custom/task", "a b", "é", "e\u0301", "UPPER", "upper", "percent%value"}
		if role == graph.KindBead {
			for _, v := range types.AllIssueTypes {
				values = append(values, string(v))
			}
		} else {
			for _, v := range types.WellKnownDependencyTypes() {
				values = append(values, string(v))
			}
		}
		for _, value := range values {
			build := NativeIssueTypeURL
			wantBase := IssueTypeURL(scope)
			if role == graph.KindLink {
				build = NativeDependencyTypeURL
				wantBase = DependencyTypeURL(scope)
			}
			id, err := build(scope, value)
			if err != nil {
				t.Fatal(err)
			}
			base, got, ok := NativeTypeClassification(scope, id, role)
			if !ok || got != value || base != wantBase {
				t.Fatalf("exact role roundtrip %q: %q %q %t", value, base, got, ok)
			}
			d, err := nativeConcreteDescriptor(scope, id, role)
			if err != nil {
				t.Fatal(err)
			}
			if d.ID() != id || d.Name() != value || d.Describes() != role || len(d.ConformsTo()) != 1 || d.ConformsTo()[0] != wantBase {
				t.Fatalf("native descriptor lost role/identity: %s", d.CanonicalJSON())
			}
			parsed, err := graph.ParseTypeDescriptor(d.CanonicalJSON())
			if err != nil || parsed.Fingerprint() != d.Fingerprint() {
				t.Fatalf("descriptor roundtrip: %v", err)
			}
		}
	}
}

func TestNativeConcreteTypeIdentityRefusals(t *testing.T) {
	const scope = "https://example.test/work/"
	for _, value := range []string{"", string([]byte{0xff})} {
		if _, err := NativeIssueTypeURL(scope, value); err == nil {
			t.Fatalf("mapped invalid classification %q", value)
		}
		if _, err := NativeDependencyTypeURL(scope, value); err == nil {
			t.Fatalf("mapped invalid kind %q", value)
		}
	}
	for _, id := range []string{IssueTypeURL(scope), DependencyTypeURL(scope), scope + "memory", scope + "bad//path", scope + "bad/%61", "https://other.test/task"} {
		if _, _, ok := NativeTypeClassification(scope, id, graph.KindBead); ok {
			t.Fatalf("decoded noncanonical/foreign identity %q", id)
		}
	}
	if id, err := NativeIssueTypeURL(scope, strings.Repeat("é", 32)); err != nil || len(id) > len(scope)+512 {
		t.Fatalf("native Unicode column bound: %s %v", id, err)
	}
	if _, err := NativeIssueTypeURL(scope, strings.Repeat("x", 33)); err == nil {
		t.Fatal("unrepresentable Issue class accepted")
	}
	if _, err := NativeDependencyTypeURL(scope, strings.Repeat("é", 17)); err == nil {
		t.Fatal("Dependency byte bound relaxed")
	}
	if _, err := NativeIssueTypeURL("not-a-scope", "task"); err == nil {
		t.Fatal("invalid Scope accepted")
	}
}

func TestNativeIssuePropertyReplacementPreservesAffiliation(t *testing.T) {
	source := &types.Issue{Title: "Original", Priority: 2, IssueType: types.TypeTask}
	for _, tc := range []struct {
		name, raw string
		want      types.IssueType
		refuses   bool
	}{
		{name: "replacement-omission", raw: `[{"op":"replace","path":"","value":{"title":"Replacement","priority":1}}]`, want: types.TypeTask},
		{name: "setter-then-property-replacement", raw: `[{"op":"replace","path":"/issue_type","value":"bug"},{"op":"replace","path":"","value":{"title":"Replacement","priority":1}}]`, want: types.TypeBug},
		{name: "property-replacement-then-setter", raw: `[{"op":"replace","path":"","value":{"title":"Replacement","priority":1}},{"op":"replace","path":"/issue_type","value":"bug"}]`, want: types.TypeBug},
		{name: "explicit-remove", raw: `[{"op":"remove","path":"/issue_type"}]`, refuses: true},
		{name: "removed-affiliation-stays-invalid", raw: `[{"op":"remove","path":"/issue_type"},{"op":"replace","path":"","value":{"title":"Replacement","priority":1}}]`, refuses: true},
		{name: "required-title-still-required", raw: `[{"op":"replace","path":"","value":{"priority":1}}]`, refuses: true},
		{name: "required-priority-still-required", raw: `[{"op":"replace","path":"","value":{"title":"Replacement"}}]`, refuses: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			patch, err := historicalIssuePropertiesPatch(source, []byte(tc.raw), publicops.MetadataPatch{})
			if (err != nil) != tc.refuses {
				t.Fatalf("refusal=%t got %v", tc.refuses, err)
			}
			if err == nil && (!patch.IssueType.Set || patch.IssueType.Value != tc.want) {
				t.Fatalf("affiliation changed/dropped: %+v", patch.IssueType)
			}
			if source.IssueType != types.TypeTask || source.Title != "Original" {
				t.Fatal("patch parsing changed source")
			}
		})
	}
}

func TestNativeFamilyRequiresExactDescriptorFacts(t *testing.T) {
	const scope = "https://example.test/work/"
	for _, id := range []string{scope + "future-memory", scope + "neutral", scope + "types/refines", "https://foreign.test/task"} {
		d, err := graph.NewTypeDescriptor(graph.TypeDescriptorSpec{ID: id, Name: "Neutral", Describes: graph.KindBead})
		if err != nil {
			t.Fatal(err)
		}
		if got := nativeFamilyReference(scope, id, graph.KindBead, d); got != id || isConcreteNativeIssue(scope, id, d) {
			t.Fatalf("unrelated descriptor promoted to native: %s", id)
		}
	}
	id, err := NativeIssueTypeURL(scope, "custom/task")
	if err != nil || id != scope+"custom/task" {
		t.Fatalf("segmented name: %s %v", id, err)
	}
	d, err := nativeConcreteDescriptor(scope, id, graph.KindBead)
	if err != nil {
		t.Fatal(err)
	}
	if nativeFamilyReference(scope, id, graph.KindBead) != id {
		t.Fatal("uncaptured custom declaration inferred from spelling")
	}
	if nativeFamilyReference(scope, id, graph.KindBead, d) != IssueTypeURL(scope) {
		t.Fatal("checked custom descriptor lost native ancestry")
	}
	if isConcreteNativeDependency(scope, id, d) {
		t.Fatal("Issue descriptor became Dependency")
	}
}

func TestNativeTypeCanonicalClassificationSpelling(t *testing.T) {
	const scope = "https://example.test/work/"
	for raw, want := range map[string]string{"a b": "a%20b", "percent%value": "percent%25value", "é": "%C3%A9", "e\u0301": "e%CC%81", "custom/task": "custom/task", "UPPER": "UPPER"} {
		id, err := NativeIssueTypeURL(scope, raw)
		if err != nil || id != scope+want {
			t.Errorf("%q -> %q %v want %q", raw, id, err, scope+want)
		}
		_, got, ok := NativeTypeClassification(scope, id, graph.KindBead)
		if !ok || got != raw {
			t.Errorf("roundtrip %q -> %q %t", raw, got, ok)
		}
	}
	for _, raw := range []string{".", "..", "a//b", "a/../b", "a/", string([]byte{0}), "a\\b"} {
		if _, err := NativeIssueTypeURL(scope, raw); err == nil {
			t.Errorf("unrepresentable native label admitted %q", raw)
		}
	}
	for _, raw := range []string{"memory", "references", "blocks", "issue", "dependency"} {
		if _, err := NativeIssueTypeURL(scope, raw); err == nil {
			t.Errorf("selected Type collision admitted %q", raw)
		}
	}
	for _, raw := range []string{"memory", "references", "task", "decision", "issue", "dependency"} {
		if _, err := NativeDependencyTypeURL(scope, raw); err == nil {
			t.Errorf("selected Type collision admitted %q", raw)
		}
	}
}
