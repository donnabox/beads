// Package memorytypes implements candidate, fixed Memory contract checks.
// It has no runtime callers and does not install or resolve type definitions.
package memorytypes

// BeadKind is an internal contract selection, not a public ID or CLI selector.
// Callers must establish the exact applicable contract from resolved affiliation;
// matching properties alone never establishes membership.
type BeadKind uint8

const (
	UnknownBead BeadKind = iota
	Memory
	Observation
	Decision
	Procedure
	Policy
	Lesson
	Issue // The selected Issue contract, including its known conformers.
	OtherBead
)

// IsMemory reports membership in the fixed Memory family.
func (k BeadKind) IsMemory() bool {
	switch k {
	case Memory, Observation, Decision, Procedure, Policy, Lesson:
		return true
	default:
		return false
	}
}

// ConformsTo reports fixed nominal or inherited membership. Unknown values and the OtherBead classification never match, including
// themselves. Issue and other Beads do not conform to Memory.
func (k BeadKind) ConformsTo(contract BeadKind) bool {
	if k < Memory || k > Issue || contract < Memory || contract > Issue {
		return false
	}
	return k == contract || (contract == Memory && k.IsMemory())
}

// LinkKind selects an informational Memory relationship. These values do not
// represent Dependency types and confer no Issue workflow behavior.
type LinkKind uint8

const (
	UnknownLink LinkKind = iota
	Association          // Proposed new name; does not rebind Dependency related/relates-to.
	Guidance
	AppliesTo
	Supports
	Challenges
	MemorySupersedes // Does not invoke Issue-closing supersede behavior.
)

func (k LinkKind) valid() bool { return k >= Association && k <= MemorySupersedes }

// ConformsTo reports nominal or Association-family membership in this fixed
// release catalog. A future dynamic catalog must use actual conformance rather
// than treat these six supplied kinds as the permanently closed family.
func (k LinkKind) ConformsTo(contract LinkKind) bool {
	return k.valid() && contract.valid() && (k == contract || contract == Association)
}
