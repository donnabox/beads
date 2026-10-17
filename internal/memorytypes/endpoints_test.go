package memorytypes

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCandidateEndpointPairs(t *testing.T) {
	// The draft descriptors are an independent, test-only oracle. No descriptor
	// parsing or illustrative identities are part of the production package.
	var fixture struct {
		Definitions []struct {
			ID         string `json:"id"`
			Properties struct {
				LinkConstraints struct {
					Pairs []struct {
						Source candidateEndpoint `json:"source"`
						Target candidateEndpoint `json:"target"`
					} `json:"typeConstraints"`
				} `json:"linkConstraints"`
			} `json:"properties"`
		} `json:"definitions"`
	}
	readJSON(t, "testdata/candidate/definitions.json", &fixture)
	links := map[string]LinkKind{"association": Association, "memory-guidance": Guidance, "memory-applies-to": AppliesTo, "memory-supports": Supports, "memory-challenges": Challenges, "memory-supersedes": MemorySupersedes}
	variants := []EndpointFacts{{Class: LocalBead, Kind: Memory}, {Class: LocalBead, Kind: Observation}, {Class: LocalBead, Kind: Decision}, {Class: LocalBead, Kind: Procedure}, {Class: LocalBead, Kind: Policy}, {Class: LocalBead, Kind: Lesson}, {Class: LocalBead, Kind: Issue}, {Class: LocalBead, Kind: OtherBead}, {Class: CrossScopeBead}, {Class: ExternalResource}}
	count := 0
	for _, definition := range fixture.Definitions {
		name := definition.ID[strings.LastIndex(definition.ID, "/")+1:]
		kind, ok := links[name]
		if !ok {
			continue
		}
		count++
		t.Run(name, func(t *testing.T) {
			for _, source := range variants {
				for _, target := range variants {
					source.CanonicalIdentity = "source"
					target.CanonicalIdentity = "target"
					target.Pinned = true
					want := false
					for _, pair := range definition.Properties.LinkConstraints.Pairs {
						if pair.Source.matches(t, source) && pair.Target.matches(t, target) {
							want = true
						}
					}
					err := ValidateEndpoints(kind, source, target)
					if (err == nil) != want {
						t.Errorf("source=%+v target=%+v: %v, expected acceptance=%v", source, target, err, want)
					}
				}
			}
		})
	}
	if count != 6 {
		t.Fatalf("checked %d candidate relationships, expected 6", count)
	}
}

// Independent test-side family inventory; never use production IsMemory here.
var fixtureMemoryKinds = map[BeadKind]struct{}{
	Memory: {}, Observation: {}, Decision: {}, Procedure: {}, Policy: {}, Lesson: {},
}

type candidateEndpoint struct {
	ResourceKind string            `json:"resourceKind"`
	Scope        string            `json:"scope"`
	Types        []json.RawMessage `json:"types"`
}

func (c candidateEndpoint) matches(t *testing.T, e EndpointFacts) bool {
	t.Helper()
	if c.ResourceKind == "resource" {
		return e.Class == ExternalResource
	}
	if c.ResourceKind != "bead" {
		t.Fatal("unsupported fixture resource", c.ResourceKind)
	}
	if c.Scope == "cross" {
		return e.Class == CrossScopeBead
	}
	if c.Scope != "same" {
		t.Fatal("unsupported fixture scope", c.Scope)
	}
	if e.Class != LocalBead {
		return false
	}
	for _, raw := range c.Types {
		if string(raw) == `"*"` {
			return true
		}
		var ref struct {
			URI      string `json:"uri"`
			Revision string `json:"revision"`
		}
		if err := json.Unmarshal(raw, &ref); err != nil {
			t.Fatal(err)
		}
		switch ref.URI {
		case "https://example.org/memory-types/memory":
			if ref.Revision != "memory-contract-1" {
				t.Fatal("changed exact contract", ref)
			}
			if _, known := fixtureMemoryKinds[e.Kind]; known {
				return true
			}
		case "https://example.org/types/beads/issue":
			if ref.Revision != "issue-1" {
				t.Fatal("changed exact contract", ref)
			}
			if e.Kind == Issue {
				return true
			}
		default:
			t.Fatal("unrecognized fixture contract", ref)
		}
	}
	return false
}

func TestPinsSelfAndCorrelatedPairs(t *testing.T) {
	memory := EndpointFacts{Class: LocalBead, Kind: Procedure, CanonicalIdentity: "memory"}
	other := EndpointFacts{Class: LocalBead, Kind: Decision, CanonicalIdentity: "other"}
	for kind := Association; kind <= MemorySupersedes; kind++ {
		self := memory
		self.Pinned = true
		if err := ValidateEndpoints(kind, memory, self); err == nil {
			t.Errorf("kind %d accepted self across pins", kind)
		}
	}
	for _, kind := range []LinkKind{Supports, Challenges, MemorySupersedes} {
		if err := ValidateEndpoints(kind, memory, other); err == nil {
			t.Errorf("kind %d accepted unpinned citation", kind)
		}
		other.Pinned = true
		if err := ValidateEndpoints(kind, memory, other); err != nil {
			t.Fatal(err)
		}
		other.Pinned = false
	}
	issueSource := EndpointFacts{Class: LocalBead, Kind: Issue, CanonicalIdentity: "issue-one"}
	issueTarget := issueSource
	issueTarget.CanonicalIdentity = "issue-two"
	// Each Issue independently matches one side of Association's union, but no
	// single arm accepts Issue->Issue. Flattening endpoint unions would admit it.
	if err := ValidateEndpoints(Association, issueSource, issueTarget); err == nil {
		t.Fatal("flattened correlated pair arms")
	}
	if err := ValidateEndpoints(Guidance, issueSource, memory); err != nil {
		t.Fatal(err)
	}
	if err := ValidateEndpoints(Guidance, memory, issueSource); err == nil {
		t.Fatal("reversed Guidance")
	}
	// No graph-state parameter: valid parallel relationships and longer cycles
	// are not accidentally prohibited by the self-link check.
	for range 2 {
		if err := ValidateEndpoints(Association, memory, other); err != nil {
			t.Fatal(err)
		}
	}
	if err := ValidateEndpoints(Association, other, memory); err != nil {
		t.Fatal(err)
	}
}

func TestUnresolvedAndExternalFacts(t *testing.T) {
	memory := EndpointFacts{Class: LocalBead, Kind: Memory, CanonicalIdentity: "memory"}
	for _, endpoint := range []EndpointFacts{
		{CanonicalIdentity: "unknown"},
		{Class: LocalBead, CanonicalIdentity: "unresolved-type"},
		{Class: LocalBead, Kind: Memory},
		{Class: CrossScopeBead, Kind: Memory, CanonicalIdentity: "claimed-remote-memory"},
		{Class: ExternalResource, Kind: Issue, CanonicalIdentity: "claimed-resource-issue"},
		{Class: 255, CanonicalIdentity: "unknown-class"},
	} {
		if err := ValidateEndpoints(Association, memory, endpoint); err == nil {
			t.Errorf("accepted unproven facts %+v", endpoint)
		}
	}
	remote := EndpointFacts{Class: CrossScopeBead, CanonicalIdentity: "opaque-remote", Pinned: true}
	if err := ValidateEndpoints(Supports, memory, remote); err != nil {
		t.Fatal(err)
	}
	resource := EndpointFacts{Class: ExternalResource, CanonicalIdentity: "opaque-resource", Pinned: true}
	if err := ValidateEndpoints(Supports, memory, resource); err == nil {
		t.Fatal("resource mistaken for cross-scope Bead")
	}
	if err := ValidateEndpoints(UnknownLink, memory, resource); err == nil {
		t.Fatal("unknown relationship admitted")
	}
}

func TestCandidateMembership(t *testing.T) {
	var fixture struct {
		Declarations []struct {
			Source string `json:"source"`
			Target struct {
				URI      string `json:"uri"`
				Revision string `json:"revision"`
			} `json:"target"`
		} `json:"declarations"`
	}
	readJSON(t, "testdata/candidate/definitions.json", &fixture)
	if len(fixture.Declarations) != 5 {
		t.Fatal("candidate changed conformance declarations")
	}
	seen := map[BeadKind]bool{}
	for _, declaration := range fixture.Declarations {
		name := strings.TrimPrefix(declaration.Source, "https://example.org/memory-types/")
		kind, ok := fixtureKinds[name]
		if !ok || kind == Memory || seen[kind] {
			t.Fatalf("unexpected child declaration %s", declaration.Source)
		}
		seen[kind] = true
		if declaration.Target.URI != "https://example.org/memory-types/memory" || declaration.Target.Revision != "memory-contract-1" {
			t.Fatal("candidate exact parent changed")
		}
		if !kind.ConformsTo(Memory) {
			t.Fatal("missing candidate membership", kind)
		}
	}
}

func TestNoEntirelyExternalPair(t *testing.T) {
	for _, kind := range []LinkKind{Association, Guidance, AppliesTo, Supports, Challenges, MemorySupersedes} {
		for _, sourceClass := range []EndpointClass{CrossScopeBead, ExternalResource} {
			for _, targetClass := range []EndpointClass{CrossScopeBead, ExternalResource} {
				source := EndpointFacts{Class: sourceClass, CanonicalIdentity: "source", Pinned: true}
				target := EndpointFacts{Class: targetClass, CanonicalIdentity: "target", Pinned: true}
				if err := ValidateEndpoints(kind, source, target); err == nil {
					t.Errorf("kind %d admits all-external pair: %v -> %v", kind, sourceClass, targetClass)
				}
			}
		}
	}
}

func TestClassificationDoesNotChangeIdentity(t *testing.T) {
	memory := EndpointFacts{Class: LocalBead, Kind: Memory, CanonicalIdentity: "same"}
	for _, class := range []EndpointClass{CrossScopeBead, ExternalResource} {
		other := EndpointFacts{Class: class, CanonicalIdentity: "same", Pinned: true}
		if err := ValidateEndpoints(Association, memory, other); err == nil {
			t.Errorf("class %d created a new identity namespace", class)
		}
	}
}
