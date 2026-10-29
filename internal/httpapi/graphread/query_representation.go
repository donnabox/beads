package graphread

// These closed DTOs select the v0.1 full current Resource representation for
// Query only. Existing Read projection and its older pinned codec are unchanged.
// Mapping happens before evaluation: predicates and returned records therefore
// see the same revisionProvenance fields, never a post-evaluation rename.

import (
	"encoding/json"
	"fmt"
	"unicode/utf8"

	"github.com/steveyegge/beads/internal/httpapi/bdpwire"
	"github.com/steveyegge/beads/internal/resourcequery"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

type queryRevisionActor struct {
	ID    string                   `json:"id"`
	Basis bdpwire.AttributionBasis `json:"basis"`
}

type queryRevisionProvenance struct {
	Actor       *queryRevisionActor     `json:"actor,omitempty"`
	CommittedAt *bdpwire.ContextTime    `json:"committedAt,omitempty"`
	Agent       *bdpwire.ContextString  `json:"agent,omitempty"`
	Message     *bdpwire.ContextMessage `json:"message,omitempty"`
}

type queryBeadRecord struct {
	ID                 string                       `json:"id"`
	Type               string                       `json:"type"`
	Revision           string                       `json:"revision"`
	Properties         bdpwire.Properties           `json:"properties"`
	Metadata           bdpwire.Metadata             `json:"metadata"`
	OwnedLinks         map[string][]queryLinkRecord `json:"ownedLinks,omitzero"`
	RevisionProvenance *queryRevisionProvenance     `json:"revisionProvenance,omitempty"`
}

type queryLinkRecord struct {
	ID                 string                   `json:"id"`
	Type               string                   `json:"type"`
	Revision           string                   `json:"revision"`
	Source             bdpwire.Reference        `json:"source"`
	Target             bdpwire.Reference        `json:"target"`
	Properties         bdpwire.Properties       `json:"properties"`
	Metadata           bdpwire.Metadata         `json:"metadata"`
	RevisionProvenance *queryRevisionProvenance `json:"revisionProvenance,omitempty"`
}

func queryResourceJSON(value any) (json.RawMessage, error) {
	var mapped any
	switch record := value.(type) {
	case bdpwire.BeadRecord:
		bead, err := queryBead(record)
		if err != nil {
			return nil, err
		}
		mapped = bead
	case bdpwire.LinkRecord:
		link, err := queryLink(record)
		if err != nil {
			return nil, err
		}
		mapped = link
	default:
		return nil, fmt.Errorf("%w: unexpected Query projection record", graphstore.ErrInvalidStore)
	}
	raw, err := json.Marshal(mapped)
	if err != nil {
		return nil, fmt.Errorf("%w: Query representation encoding", graphstore.ErrInvalidStore)
	}
	return raw, nil
}

func queryProvenance(attribution *bdpwire.Attribution, change *bdpwire.ChangeContext) (*queryRevisionProvenance, error) {
	if attribution == nil && change == nil {
		return nil, nil
	}
	result := &queryRevisionProvenance{}
	if attribution != nil {
		if attribution.Principal == "" || !utf8.ValidString(attribution.Principal) || !attribution.Basis.Valid() {
			return nil, fmt.Errorf("%w: invalid Query revision actor", graphstore.ErrInvalidStore)
		}
		result.Actor = &queryRevisionActor{ID: attribution.Principal, Basis: attribution.Basis}
	}
	if change != nil {
		if err := change.Validate(); err != nil {
			return nil, fmt.Errorf("%w: invalid Query revision context", graphstore.ErrInvalidStore)
		}
		if change.CommittedAt.State == bdpwire.ContextTimePresent {
			value := *change.CommittedAt.Value
			if !queryRepresentationInputDateTime(value) {
				return nil, fmt.Errorf("%w: invalid Query revision timestamp", graphstore.ErrInvalidStore)
			}
			// Core requires uppercase T/Z in output. This mapper cannot rewrite
			// a recorded spelling or silently omit its provenance to satisfy it.
			if value[10] != 'T' || value[len(value)-1] == 'z' {
				return nil, fmt.Errorf("%w: recorded Query revision timestamp requires representation normalization", resourcequery.ErrUnsupported)
			}
		}
		for _, value := range []*string{change.CommittedAt.Value, change.Agent.Value, change.Message.Value} {
			if value != nil && !utf8.ValidString(*value) {
				return nil, fmt.Errorf("%w: invalid Query revision context text", graphstore.ErrInvalidStore)
			}
		}
		// All three recorded state objects travel together. Missing context does not
		// manufacture absent/undetermined state objects or any authentication claim.
		result.CommittedAt = &change.CommittedAt
		result.Agent = &change.Agent
		result.Message = &change.Message
	}
	return result, nil
}

func queryBead(record bdpwire.BeadRecord) (queryBeadRecord, error) {
	if record.Links != nil {
		return queryBeadRecord{}, fmt.Errorf("%w: incident aggregate is not a Query inventory record", graphstore.ErrInvalidStore)
	}
	if record.Metadata == nil {
		return queryBeadRecord{}, fmt.Errorf("%w: current Query Bead metadata absent", graphstore.ErrInvalidStore)
	}
	provenance, err := queryProvenance(record.Attribution, record.ChangeContext)
	if err != nil {
		return queryBeadRecord{}, err
	}
	result := queryBeadRecord{ID: record.ID, Type: record.Type, Revision: record.Revision, Properties: record.Properties, Metadata: record.Metadata, RevisionProvenance: provenance}
	if record.OwnedLinks != nil {
		if err := record.OwnedLinks.Validate(); err != nil {
			return queryBeadRecord{}, fmt.Errorf("%w: invalid Query owned groups", graphstore.ErrInvalidStore)
		}
		result.OwnedLinks = make(map[string][]queryLinkRecord, len(record.OwnedLinks))
		for typ, group := range record.OwnedLinks {
			mapped := make([]queryLinkRecord, 0, len(group))
			for _, owned := range group {
				if owned.Type != typ || owned.Source.URI != record.ID {
					return queryBeadRecord{}, fmt.Errorf("%w: inconsistent Query owned Link", graphstore.ErrInvalidStore)
				}
				link, err := queryLink(owned)
				if err != nil {
					return queryBeadRecord{}, err
				}
				mapped = append(mapped, link)
			}
			result.OwnedLinks[typ] = mapped
		}
	}
	return result, nil
}

func queryLink(record bdpwire.LinkRecord) (queryLinkRecord, error) {
	if record.Metadata == nil {
		return queryLinkRecord{}, fmt.Errorf("%w: current Query Link metadata absent", graphstore.ErrInvalidStore)
	}
	provenance, err := queryProvenance(record.Attribution, record.ChangeContext)
	if err != nil {
		return queryLinkRecord{}, err
	}
	return queryLinkRecord{ID: record.ID, Type: record.Type, Revision: record.Revision, Source: record.Source, Target: record.Target, Properties: record.Properties, Metadata: record.Metadata, RevisionProvenance: provenance}, nil
}
