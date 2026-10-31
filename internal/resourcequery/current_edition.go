package resourcequery

import graph "github.com/steveyegge/beads/graphops"

// CurrentEditionFact is a producer's checked CURRENT selection for one Resource.
// Kind covers both Beads (including Type definition Resources) and Links. A
// retained Type affiliation or an owned retained Link is not current evidence.
type CurrentEditionFact struct {
	Reference Edition
	Kind      graph.ResourceKind
}

// CurrentEditionAuthority is an explicit trusted producer assertion, not a
// storage reader or proof inferred from a Resource body. Its zero value is not
// an asserted capture. Execute checks full correspondence with Dataset.Resources.
// The carrier exposes no mutable facts and provides no historical acquisition.
// It applies only to dynamic body acquisition through stored endpoint pins.
// Decode/Admit still reject explicit pinned refs before Dataset inspection;
// this carrier does not make static pinned refs, even skipped ones, admissible.
type CurrentEditionAuthority struct {
	complete bool
	facts    []CurrentEditionFact
}

// NewCurrentEditionAuthorityFromTrustedCapture declares that facts enumerate
// EVERY current Resource in the SAME complete checked acquisition as Resources,
// under the SAME authorization and visibility. The producer must establish each
// ID, kind and actual revision from authoritative live allocation/record evidence
// in that capture, including Type definition Beads and Links. It must return no
// usable Dataset when acquisition, validation or transaction completion fails.
//
// Calling this constructor is an authority assertion. It cannot verify native
// provenance: NewResource, a matching body token, a private writer token, nominal
// Type metadata, retained-only evidence or a legacy synthetic "current" marker
// cannot justify calling it. Only the producer that owns the checked capture may
// opt in; consumers must never synthesize facts from arbitrary Resource JSON.
//
// The facts are copied. A nil/empty slice explicitly asserts a complete EMPTY
// capture. Execute validates all records and correspondence before evaluating
// any expression; malformed or incomplete assertions fail with ErrDataset.
func NewCurrentEditionAuthorityFromTrustedCapture(facts []CurrentEditionFact) *CurrentEditionAuthority {
	return &CurrentEditionAuthority{complete: true, facts: append([]CurrentEditionFact(nil), facts...)}
}

func (e *evaluation) loadCurrentEditions(authority *CurrentEditionAuthority) error {
	if authority == nil {
		return nil
	}
	bad := func() error {
		return fail("invalid-current-edition-authority", "", "invalid complete current Resource authority", ErrDataset)
	}
	if !authority.complete {
		return bad()
	}
	// Charge each fact before map growth. Check the whole capture, even when a
	// query is unpinned or cannot reach a record. No selection-shaped authority.
	current := map[Edition]graph.ResourceKind{}
	seen := map[string]bool{}
	for _, fact := range authority.facts {
		if err := e.tick(); err != nil {
			return err
		}
		r := e.byID[fact.Reference.ID]
		if r == nil || seen[fact.Reference.ID] || fact.Reference.Revision == "" ||
			r.kind != fact.Kind || r.revision != fact.Reference.Revision {
			return bad()
		}
		seen[fact.Reference.ID] = true
		current[fact.Reference] = fact.Kind
	}
	if len(current) != len(e.byID) {
		return bad()
	}
	e.currentEditions = current
	return nil
}
