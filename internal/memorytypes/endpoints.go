package memorytypes

// EndpointClass distinguishes local contract evidence from unsupported external references.
// Zero is deliberately unresolved, never a wildcard local Bead. The caller must
// prove Bead categories actually identify Beads, never Links masquerading as Beads.
type EndpointClass uint8

const (
	Unresolved EndpointClass = iota
	LocalBead
	CrossScopeBead
	ExternalResource
)

// EndpointFacts contains only facts needed by the fixed relationship predicates.
// CanonicalIdentity must be normalized by the caller, excluding the revision.
// It uses one identity namespace across all classes: changing classification or
// locality must not change an identity or allow a self-link.
// For a local endpoint, Kind must describe the selected live or retained state,
// not necessarily its current state. The caller must separately prove existence,
// visibility, retention, contract revision, and non-workflow participation.
// Opaque external categories carry no verified Kind and are refused in this release.
type EndpointFacts struct {
	Class             EndpointClass
	Kind              BeadKind
	CanonicalIdentity string
	Pinned            bool // True only for an explicitly supplied nonempty revision.
}

// ValidateEndpoints checks correlated pairs, identity irreflexivity and explicit
// citation target pins. Success is conditional on the caller's supplied facts;
// it is not local-state admission or external target verification.
func ValidateEndpoints(kind LinkKind, source, target EndpointFacts) error {
	if !kind.valid() {
		return invalid("/type", "a fixed Memory relationship")
	}
	if err := endpointFacts(source, "/source"); err != nil {
		return err
	}
	if err := endpointFacts(target, "/target"); err != nil {
		return err
	}
	if source.CanonicalIdentity == target.CanonicalIdentity {
		return invalid("/target", "an identity different from source, regardless of revision pins")
	}
	if (kind == Supports || kind == Challenges) && !target.Pinned {
		return invalid("/target/revision", "an explicitly pinned citation target")
	}
	for _, pair := range pairs(kind) {
		if pair.source.matches(source) && pair.target.matches(target) {
			return nil
		}
	}
	return invalid("/endpoints", "source and target matching the same allowed pair")
}

func endpointFacts(e EndpointFacts, path string) error {
	if e.CanonicalIdentity == "" {
		return invalid(path, "a caller-normalized canonical identity")
	}
	switch e.Class {
	case LocalBead:
		if e.Kind < Memory || e.Kind > Issue {
			return invalid(path, "a resolved local Issue or Memory family contract")
		}
	case CrossScopeBead, ExternalResource:
		return invalid(path, "a same-scope Issue or Memory; external endpoints are outside this release")
	default:
		return invalid(path, "a resolved local Issue or Memory classification")
	}
	return nil
}

type endpointMatch uint8

const (
	memory endpointMatch = iota
	issue
)

func (m endpointMatch) matches(e EndpointFacts) bool {
	switch m {
	case memory:
		return e.Class == LocalBead && e.Kind.IsMemory()
	case issue:
		return e.Class == LocalBead && e.Kind == Issue
	}
	return false
}

type endpointPair struct{ source, target endpointMatch }

func pairs(k LinkKind) []endpointPair {
	switch k {
	case Association:
		return []endpointPair{{memory, issue}, {issue, memory}, {memory, memory}}
	case Guidance:
		return []endpointPair{{issue, memory}}
	case AppliesTo:
		return []endpointPair{{memory, issue}, {memory, memory}}
	case Supports, Challenges, MemorySupersedes:
		return []endpointPair{{memory, memory}}
	}
	return nil
}
