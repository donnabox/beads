package graphstore

import (
	"context"
	"fmt"

	graph "github.com/steveyegge/beads/graphops"
	publicops "github.com/steveyegge/beads/issueops"
	"github.com/steveyegge/beads/resourceget"
)

// ResourceSnapshot is one bounded, checked native transaction's complete
// current Resource inventory. Links remain Resources, never extra Beads.
// WriterToken is a private coordination observation, not a public revision.
type ResourceSnapshot struct {
	WriterToken string
	Documents   []resourceget.ResourceDocument
	Types       []ResourceTypeFact
}

// ResourceTypeFact is captured beside the records in the SAME checked read.
// Reference is the semantic affiliation; Definition is its actual local Bead
// edition. They differ only for an explicit checked intrinsic correspondence.
// Ancestors includes self and retains exact editions. This is not a mutable
// registry, a URI-only descriptor or a source of independent admission authority.
type ResourceTypeFact struct {
	Reference  resourceget.TypeReference
	Definition resourceget.TypeReference
	Describes  graph.ResourceKind
	Abstract   bool
	// Current is established from the actual live definition allocation in
	// this capture. Retained exact affiliation facts never imply Current.
	Current   bool
	Ancestors []resourceget.TypeReference
}

// BeadDocumentPage applies ordinary Bead visibility and Type selection to a
// common inventory. HasMore reports intentional paging, not an incomplete read.
type BeadDocumentPage struct {
	Items   []resourceget.ResourceDocument
	HasMore bool
}

// ResourceDocumentReader is the ordinary complete read boundary for the
// selected Types-aware store. Implementations must bind allocation/backing,
// exact Type edition and full record in the same checked transaction. It does
// not authorize a constructor's syntax-valid document or infer an ID alias.
// Count applies selection before pagination; a nonzero Limit is not a count.
type ResourceDocumentReader interface {
	GetResourceDocument(context.Context, string, string) (resourceget.ResourceDocument, error)
	GetSelectedResourceDocument(context.Context, string, string, string) (string, resourceget.ResourceDocument, error)
	ListBeadDocuments(context.Context, BeadListRequest) (BeadDocumentPage, error)
	CountBeadDocuments(context.Context, BeadListRequest) (int, error)
	CurrentResourceDocuments(context.Context) (ResourceSnapshot, error)
	// The optional parent uses the existing exact Issue selector, resolved in
	// the SAME transaction as native filtering and complete-document capture.
	// IncludeComments is explicitly unsupported in this common record route;
	// callers must not silently discard requested legacy enrichment.
	ListIssueDocuments(context.Context, publicops.ListRequest, *IssueSelector) (BeadDocumentPage, error)
	CountIssueDocuments(context.Context, publicops.ListRequest, *IssueSelector) (int, error)
}

// ResolveType uses whole exact semantic/local-definition identities in this
// capture. It never strips a path component, folds case or asks for latest.
// CLI classification shorthand must be resolved by its existing native adapter.
func (s ResourceSnapshot) ResolveType(selector string) (resourceget.TypeReference, error) {
	var result resourceget.TypeReference
	for _, fact := range s.Types {
		if !fact.Current {
			continue
		}
		if fact.Reference.URI != selector && fact.Definition.URI != selector {
			continue
		}
		if fact.Reference.URI == "" || fact.Reference.Revision == "" || fact.Definition.URI == "" || fact.Definition.Revision == "" {
			return resourceget.TypeReference{}, fmt.Errorf("%w: incomplete captured Type fact", ErrInvalidStore)
		}
		if result.URI != "" && result != fact.Reference {
			return resourceget.TypeReference{}, fmt.Errorf("%w: ambiguous captured Type identity", ErrInvalidStore)
		}
		result = fact.Reference
	}
	if result.URI == "" {
		return resourceget.TypeReference{}, fmt.Errorf("%w: unknown selected Type", ErrCapabilityUnavailable)
	}
	return result, nil
}

// MatchesType evaluates the actual affiliation against captured exact ancestry.
// A missing affiliation fact is corruption, not a nonmatching record to omit.
// selector is an unpinned whole identity; exact selects nominal affiliation
// rather than refinements, not an invented current-revision restriction.
func (s ResourceSnapshot) MatchesType(affiliation resourceget.TypeReference, selector string, exact bool) (bool, error) {
	var selected *ResourceTypeFact
	for _, fact := range s.Types {
		if fact.Reference == affiliation {
			if selected != nil {
				return false, fmt.Errorf("%w: duplicate exact captured Type", ErrInvalidStore)
			}
			copy := fact
			selected = &copy
		}
	}
	if selected == nil || affiliation.URI == "" || affiliation.Revision == "" {
		return false, fmt.Errorf("%w: Resource affiliation absent from captured Types", ErrInvalidStore)
	}
	if selector == "" {
		return true, nil
	}
	// Predicate semantics use literal semantic identities. The explicit product
	// ResolveType adapter must not introduce definition aliases here.
	known := false
	for _, fact := range s.Types {
		if fact.Current && fact.Reference.URI == selector {
			known = true
		}
	}
	if !known {
		return false, fmt.Errorf("%w: unknown semantic Type selector", ErrCapabilityUnavailable)
	}
	if affiliation.URI == selector {
		return true, nil
	}
	if exact {
		return false, nil
	}
	for _, ancestor := range selected.Ancestors {
		if ancestor.URI == selector {
			return true, nil
		}
	}
	return false, nil
}
