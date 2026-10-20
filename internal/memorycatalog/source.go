// Package memorycatalog projects the fixed, selected Memory design into pure
// declarations. It does not install Types or admit a product operation.
package memorycatalog

// Source identifies the immutable design input, not a runtime qualification.
type Source struct {
	Repository string
	Commit     string
	Path       string
	SHA256     string
	Bytes      int
}

// SelectedSource returns the independently verified publication identity.
func SelectedSource() Source {
	return Source{
		Repository: "gastownhall/bdp",
		Commit:     "bbd97f4b57ddbcea8ee5ca38adc2ec7d0a784f00",
		Path:       "specs/types/memory/memory.md",
		SHA256:     "25905e496f36ab7f95046c5c98ed681aeb9aa0bb1c832d63beb6038120bd0ac5",
		Bytes:      105123,
	}
}

// Symbol is a catalog-local name, not a canonical Type ID or CLI selector.
type Symbol string

const (
	Memory        Symbol = "memory"
	Evaluation    Symbol = "evaluation"
	Constraint    Symbol = "constraint"
	Policy        Symbol = "policy"
	Decision      Symbol = "decision"
	Procedure     Symbol = "procedure"
	Requirement   Symbol = "requirement"
	Scenario      Symbol = "scenario"
	Specification Symbol = "specification"
	References    Symbol = "references"
	SupersededBy  Symbol = "superseded-by"
	Requires      Symbol = "requires"
	Defines       Symbol = "defines"
	SubjectTo     Symbol = "subject-to"
	Conformance   Symbol = "conformance"
	Addresses     Symbol = "addresses"
	Satisfies     Symbol = "satisfies"
	Violates      Symbol = "violates"
)

func beadSymbols() []Symbol {
	return []Symbol{Memory, Evaluation, Constraint, Policy, Decision, Procedure, Requirement, Scenario, Specification}
}

func linkSymbols() []Symbol {
	return []Symbol{References, SupersededBy, Requires, Defines, SubjectTo, Conformance, Addresses, Satisfies, Violates}
}

func symbols() []Symbol { return append(beadSymbols(), linkSymbols()...) }

func parent(symbol Symbol) Symbol {
	switch symbol {
	case Evaluation, Constraint:
		return Memory
	case Policy, Decision, Procedure, Requirement, Scenario, Specification:
		return Constraint
	case Defines:
		return Requires
	case Addresses, Satisfies, Violates:
		return Conformance
	case SupersededBy, Requires, SubjectTo, Conformance:
		return References
	}
	return ""
}
