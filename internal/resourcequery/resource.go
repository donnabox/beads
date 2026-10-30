package resourcequery

import (
	"context"
	"encoding/json"
	"sort"
	"strings"

	graph "github.com/steveyegge/beads/graphops"
)

type record struct {
	kind              graph.ResourceKind
	id, revision, typ string
	typePin           string
	family            map[string]bool
	body              map[string]any
	raw               json.RawMessage
	source, target    reference
}

func (r *record) edition() Edition { return Edition{r.id, r.revision} }

// NewResource normalizes JSON with Core's numeric admission and equality law.
// It validates the evaluator's record boundary, not an installed Type schema.
// The complete input record is preserved; unrecognized full-record fields are
// not silently stripped. Endpoint acquisition is deliberately deferred.
func NewResource(kind graph.ResourceKind, raw []byte) (Resource, error) {
	r, err := readRecord(kind, raw)
	if err != nil {
		return Resource{}, err
	}
	return Resource{Kind: kind, Data: append(json.RawMessage(nil), r.raw...)}, nil
}
func readRecord(kind graph.ResourceKind, raw []byte) (*record, error) {
	bad := func() (*record, error) {
		return nil, fail("invalid-resource", "", "invalid complete current Resource record", ErrDataset)
	}
	if !kind.Valid() {
		return bad()
	}
	canonical, err := graph.CanonicalizeJSON(raw)
	if err != nil {
		return bad()
	}
	v, err := jsonValue(canonical)
	if err != nil {
		return bad()
	}
	m, ok := v.(map[string]any)
	if !ok {
		return bad()
	}
	id, ok := m["id"].(string)
	if !ok || graph.ValidateTypeURL(id) != nil || graph.ValidateReferenceURI(id) != nil || strings.Contains(id, "?") {
		return bad()
	}
	rev, ok := m["revision"].(string)
	if !ok || rev == "" {
		return bad()
	}
	typ, err := parseReference(m["type"], "")
	if err != nil {
		return bad()
	}
	if graph.ValidateTypeURL(typ.uri) != nil {
		return bad()
	}
	if _, ok = m["properties"].(map[string]any); !ok {
		return bad()
	}
	if meta, present := m["metadata"]; present {
		if _, ok = meta.(map[string]any); !ok {
			return bad()
		}
	}
	r := &record{kind: kind, id: id, revision: rev, typ: typ.uri, typePin: typ.pin, body: m, raw: canonical}
	if kind == graph.KindLink {
		r.source, err = parseReference(m["source"], "")
		if err != nil {
			return bad()
		}
		r.target, err = parseReference(m["target"], "")
		if err != nil {
			return bad()
		}
	} else {
		if _, present := m["source"]; present {
			return bad()
		}
		if _, present := m["target"]; present {
			return bad()
		}
	}
	return r, nil
}

type evaluation struct {
	ctx                           context.Context
	request                       *Request
	limits                        Limits
	work                          int
	byID                          map[string]*record
	beads, links                  []*record
	incoming, outgoing, connected map[string][]*record
	families                      map[string]map[string]bool
	traces                        map[string]Traversal
	walkRoots                     map[*node][]*record
}

func newEvaluation(ctx context.Context, r *Request, data Dataset, l Limits) (*evaluation, error) {
	e := &evaluation{ctx: ctx, request: r, limits: l, byID: map[string]*record{}, incoming: map[string][]*record{}, outgoing: map[string][]*record{}, connected: map[string][]*record{}, families: map[string]map[string]bool{}, traces: map[string]Traversal{}, walkRoots: map[*node][]*record{}}
	for _, input := range data.Resources {
		if err := e.tick(); err != nil {
			return nil, err
		}
		rec, err := readRecord(input.Kind, input.Data)
		if err != nil {
			return nil, err
		}
		if _, duplicate := e.byID[rec.id]; duplicate {
			return nil, fail("duplicate-current-resource", "", "dataset contains duplicate current identities", ErrDataset)
		}
		e.byID[rec.id] = rec
		if rec.kind == graph.KindBead {
			e.beads = append(e.beads, rec)
		} else {
			e.links = append(e.links, rec)
		}
	}
	sortRecords(e.beads)
	sortRecords(e.links)
	// Index stored identities once, before applying any query-specific Type,
	// through predicate or revision applicability. Appending sorted Links keeps
	// exactly the former global-scan candidate order, including connected walks.
	for _, link := range e.links {
		if err := e.tick(); err != nil {
			return nil, err
		}
		e.outgoing[link.source.uri] = append(e.outgoing[link.source.uri], link)
		e.incoming[link.target.uri] = append(e.incoming[link.target.uri], link)
		e.connected[link.source.uri] = append(e.connected[link.source.uri], link)
		if link.target.uri != link.source.uri {
			e.connected[link.target.uri] = append(e.connected[link.target.uri], link)
		}
	}
	// All family uses are checked before expression evaluation, including empty
	// inventories and unreachable Boolean branches. No nominal-name fallback.
	if data.TypeEditions != nil {
		if err := e.loadEditionFamilies(data.TypeEditions); err != nil {
			return nil, err
		}
	} else if err := e.loadNominalFamilies(data); err != nil {
		return nil, err
	}
	for _, refs := range r.references {
		for _, ref := range refs.refs {
			if _, err := e.resolve(ref, collectionKind(refs.collection)); err != nil {
				return nil, err
			}
		}
	}
	return e, nil
}

func (e *evaluation) loadNominalFamilies(data Dataset) error {
	if len(e.request.families) != 0 && (data.TypeKinds == nil || data.TypeFamilies == nil) {
		return fail("family-mechanism-unavailable", e.request.families[0].at, "authoritative Type family interpretation is unsupported", ErrUnsupported)
	}
	familyKinds := map[graph.ResourceKind]bool{}
	for _, use := range e.request.families {
		if err := e.tick(); err != nil {
			return err
		}
		kind, ok := data.TypeKinds[use.uri]
		if !ok {
			return fail("family-authority-unavailable", use.at, "queried Type has no authoritative kind", ErrUnavailable)
		}
		if kind != use.kind {
			return fail("invalid-type-kind", use.at, "queried Type describes another Resource kind", ErrInvalid)
		}
		if err := e.loadFamily(use.uri, kind, data); err != nil {
			return err
		}
		familyKinds[kind] = true
	}
	// An actual Resource's ancestry depends on its kind, not on how many
	// syntactic family tests mention that kind. Preload it once per record.
	for _, rec := range e.byID {
		if familyKinds[rec.kind] {
			if rec.typePin != "" {
				return fail("exact-family-authority-unsupported", "", "pinned Type affiliation requires exact captured ancestry", ErrUnsupported)
			}
			if err := e.loadFamily(rec.typ, rec.kind, data); err != nil {
				return err
			}
			rec.family = e.families[rec.typ]
		}
	}
	return nil
}

// loadEditionFamilies never interprets a retained affiliation using the current
// edition's ancestry. Current evidence only resolves unpinned selection; the
// actual record's exact affiliation chooses its membership set.
func (e *evaluation) loadEditionFamilies(facts []TypeEditionFact) error {
	if len(e.request.families) == 0 {
		return nil
	}
	bad := func() error {
		return fail("invalid-family-authority", "", "invalid exact captured Type authority", ErrDataset)
	}
	// Grow only after charging work; a caller-supplied inventory length must
	// not trigger a large allocation before its operational budget is checked.
	byEdition := map[Edition]TypeEditionFact{}
	current := map[string]Edition{}
	for _, fact := range facts {
		if err := e.tick(); err != nil {
			return err
		}
		if graph.ValidateTypeURL(fact.Reference.ID) != nil || graph.ValidateReferenceURI(fact.Reference.ID) != nil || strings.Contains(fact.Reference.ID, "?") || fact.Reference.Revision == "" || !fact.Kind.Valid() {
			return bad()
		}
		if _, duplicate := byEdition[fact.Reference]; duplicate {
			return bad()
		}
		byEdition[fact.Reference] = fact
		if fact.Current {
			if _, duplicate := current[fact.Reference.ID]; duplicate {
				return bad()
			}
			current[fact.Reference.ID] = fact.Reference
		}
	}
	sets := map[Edition]map[string]bool{}
	exactSets := map[Edition]map[Edition]bool{}
	for _, fact := range facts {
		seen := map[Edition]bool{}
		set := map[string]bool{}
		for _, ancestor := range fact.Ancestors {
			if err := e.tick(); err != nil {
				return err
			}
			parent, known := byEdition[ancestor]
			if !known || parent.Kind != fact.Kind || seen[ancestor] {
				return bad()
			}
			seen[ancestor] = true
			set[ancestor.ID] = true
		}
		if !seen[fact.Reference] {
			return bad()
		}
		sets[fact.Reference] = set
		exactSets[fact.Reference] = seen
	}
	// A claimed complete closure must include the captured ancestors of each
	// ancestor. Charge this consistency check rather than trusting a truncated
	// set that could silently turn a family match into a negative result.
	for _, fact := range facts {
		for _, ancestor := range fact.Ancestors {
			for _, inherited := range byEdition[ancestor].Ancestors {
				if err := e.tick(); err != nil {
					return err
				}
				if !exactSets[fact.Reference][inherited] {
					return bad()
				}
			}
		}
	}
	kinds := map[graph.ResourceKind]bool{}
	for _, use := range e.request.families {
		if err := e.tick(); err != nil {
			return err
		}
		selected, known := current[use.uri]
		if !known {
			return fail("family-authority-unavailable", use.at, "queried Type has no captured current authority", ErrUnavailable)
		}
		if byEdition[selected].Kind != use.kind {
			return fail("invalid-type-kind", use.at, "queried Type describes another Resource kind", ErrInvalid)
		}
		kinds[use.kind] = true
	}
	for _, rec := range e.byID {
		if err := e.tick(); err != nil {
			return err
		}
		if !kinds[rec.kind] {
			continue
		}
		affiliation := Edition{ID: rec.typ, Revision: rec.typePin}
		if affiliation.Revision == "" {
			var known bool
			affiliation, known = current[rec.typ]
			if !known {
				return bad()
			}
		}
		fact, known := byEdition[affiliation]
		if !known || fact.Kind != rec.kind {
			return bad()
		}
		rec.family = sets[affiliation]
	}
	return nil
}
func (e *evaluation) loadFamily(uri string, kind graph.ResourceKind, data Dataset) error {
	if err := e.tick(); err != nil {
		return err
	}
	if _, loaded := e.families[uri]; loaded {
		return nil
	}
	k, known := data.TypeKinds[uri]
	ancestors, present := data.TypeFamilies[uri]
	if !known || k != kind || !present {
		return fail("invalid-family-authority", "", "complete Type authority lacks required ancestry", ErrDataset)
	}
	set := map[string]bool{}
	for _, ancestor := range ancestors {
		if err := e.tick(); err != nil {
			return err
		}
		if graph.ValidateTypeURL(ancestor) != nil || data.TypeKinds[ancestor] != kind || set[ancestor] {
			return fail("invalid-family-authority", "", "invalid Type ancestry authority", ErrDataset)
		}
		set[ancestor] = true
	}
	if !set[uri] {
		return fail("invalid-family-authority", "", "Type ancestry must be reflexive", ErrDataset)
	}
	e.families[uri] = set
	return nil
}
func (e *evaluation) tick() error {
	if err := e.ctx.Err(); err != nil {
		return err
	}
	if e.work >= e.limits.MaxWork {
		return fail("work-limit", "", "query work bound exhausted", ErrLimit)
	}
	e.work++
	return nil
}
func (e *evaluation) bounded(n int) error {
	if n > e.limits.MaxResults {
		return fail("result-limit", "", "query collection bound exhausted", ErrLimit)
	}
	return nil
}
func (e *evaluation) resolve(ref reference, kind graph.ResourceKind) (*record, error) {
	// Pin syntax and a locally matching current token do not establish an
	// exact-acquisition/History capability. Refuse before identity lookup.
	if ref.pin != "" {
		return nil, fail("exact-acquisition-unsupported", "", "exact Resource acquisition is not supported", ErrUnsupported)
	}
	if err := e.tick(); err != nil {
		return nil, err
	}
	r := e.byID[ref.uri]
	if r == nil || r.kind != kind {
		return nil, fail("required-resource-unavailable", "", "required Resource edition is unavailable in this current dataset", ErrUnavailable)
	}
	return r, nil
}
func sortRecords(rs []*record) {
	sort.Slice(rs, func(i, j int) bool {
		if rs[i].id != rs[j].id {
			return rs[i].id < rs[j].id
		}
		return rs[i].revision < rs[j].revision
	})
}
func distinct(rs []*record) []*record {
	seen := map[Edition]bool{}
	out := make([]*record, 0, len(rs))
	for _, r := range rs {
		if !seen[r.edition()] {
			seen[r.edition()] = true
			out = append(out, r)
		}
	}
	return out
}
