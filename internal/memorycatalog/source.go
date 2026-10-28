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
		Commit:     "94f699b947382e3a52385b1895a941575fcc08cb",
		Path:       "specs/types/memory/memory.md",
		SHA256:     "60c319aac0e42ea9825d6b383a6d09b7d789023fded675e647fd657f37961cb0",
		Bytes:      103192,
	}
}

// Symbol is a catalog-local name, not a canonical Type ID or CLI selector.
type Symbol string

const (
	Memory        Symbol = "memory"
	Evaluation    Symbol = "evaluation"
	Constraint    Symbol = "constraint"
	Policy        Symbol = "policy"
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
	return []Symbol{Memory, Evaluation, Constraint, Policy, Requirement, Scenario, Specification}
}

func linkSymbols() []Symbol {
	return []Symbol{References, SupersededBy, Requires, Defines, SubjectTo, Conformance, Addresses, Satisfies, Violates}
}

func symbols() []Symbol { return append(beadSymbols(), linkSymbols()...) }

func parent(symbol Symbol) Symbol {
	switch symbol {
	case Evaluation, Constraint:
		return Memory
	case Policy, Requirement, Scenario, Specification:
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
