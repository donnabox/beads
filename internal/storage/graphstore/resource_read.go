package graphstore

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	graph "github.com/steveyegge/beads/graphops"
	publicops "github.com/steveyegge/beads/issueops"
	"github.com/steveyegge/beads/resourceget"
)

func (s *Store) GetResourceDocument(ctx context.Context, address, revision string) (resourceget.ResourceDocument, error) {
	if strings.HasPrefix(address, s.ScopeURL()) {
		_, document, err := s.GetSelectedResourceDocument(ctx, address, "", revision)
		return document, err
	}
	_, document, err := s.GetSelectedResourceDocument(ctx, "", address, revision)
	return document, err
}

// GetSelectedResourceDocument resolves the whole public ID or explicit private
// locator once. It does not manufacture beads/<publicID>, guess a kind from ID
// syntax, or open a second transaction between routing and representation.
func (s *Store) GetSelectedResourceDocument(ctx context.Context, path, localID, revision string) (string, resourceget.ResourceDocument, error) {
	if s.LayoutVersion() != TypesFreshSchemaVersion {
		return "", resourceget.ResourceDocument{}, ErrCapabilityUnavailable
	}
	if revision != "" {
		if err := validateVersionToken(revision); err != nil {
			return "", resourceget.ResourceDocument{}, err
		}
	}
	var document resourceget.ResourceDocument
	var selected string
	err := s.withDefinitionReadTx(ctx, txDefinitionRead, func(tx *sql.Tx, budget *intrinsicMetadataBudget) error {
		identity, e := s.resolveCommonResourceSelectorInTx(ctx, tx, path, localID, budget)
		if e != nil {
			return e
		}
		selected = identity.PrivatePath
		if identity.State == "deleted" && revision == "" {
			return ErrGone
		}
		document, e = s.readCommonResourceInTx(ctx, tx, identity, revision, budget)
		return e
	})
	if err != nil {
		return "", resourceget.ResourceDocument{}, err
	}
	return selected, document, nil
}

func (s *Store) resolveCommonResourceSelectorInTx(ctx context.Context, tx *sql.Tx, path, localID string, budget *intrinsicMetadataBudget) (AllocatedResourceIdentity, error) {
	if localID != "" {
		if err := validatePublicID(localID); err != nil {
			return AllocatedResourceIdentity{}, err
		}
		const predicate = `(backing='issue' AND backing_key=?) OR (backing IN ('generic','dependency','informational','intrinsic','intrinsic-link','type-definition','installed-type','installed-type-link') AND COALESCE(public_id,path)=?)`
		if err := checkAllocatedIdentityBytesInBudget(ctx, tx, budget, predicate, localID, localID); err != nil {
			return AllocatedResourceIdentity{}, err
		}
		rows, err := tx.QueryContext(ctx, `SELECT `+allocatedIdentityColumns+` FROM graph_preview_catalog WHERE `+predicate+` LIMIT 2`, localID, localID)
		if err != nil {
			return AllocatedResourceIdentity{}, err
		}
		identity, err := readAllocatedIdentityRowsWithInstalled(rows, true)
		if err != nil {
			return AllocatedResourceIdentity{}, err
		}
		if identity.PublicID != localID {
			return AllocatedResourceIdentity{}, ErrInvalidStore
		}
		if path != "" {
			uri, err := graph.CanonicalPublicResourceURL(s.ScopeURL(), localID)
			if err != nil || path != identity.PrivatePath && path != uri {
				return AllocatedResourceIdentity{}, fmt.Errorf("%w: inconsistent whole Resource selectors", graph.ErrValidation)
			}
		}
		return identity, nil
	}
	if strings.HasPrefix(path, s.ScopeURL()) {
		id, err := graph.ScopeLocalPublicResourceID(s.ScopeURL(), path)
		if err != nil {
			return AllocatedResourceIdentity{}, err
		}
		return s.resolveCommonResourceSelectorInTx(ctx, tx, "", id, budget)
	}
	if err := validateResourcePath(path); err != nil {
		return AllocatedResourceIdentity{}, err
	}
	if err := checkAllocatedIdentityBytesInBudget(ctx, tx, budget, "path=?", path); err != nil {
		return AllocatedResourceIdentity{}, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT `+allocatedIdentityColumns+` FROM graph_preview_catalog WHERE path=? LIMIT 2`, path)
	if err != nil {
		return AllocatedResourceIdentity{}, err
	}
	identity, err := readAllocatedIdentityRowsWithInstalled(rows, true)
	if err != nil {
		return AllocatedResourceIdentity{}, err
	}
	if identity.PrivatePath != path {
		return AllocatedResourceIdentity{}, ErrInvalidStore
	}
	return identity, nil
}

func (s *Store) readCommonResourceInTx(ctx context.Context, tx *sql.Tx, identity AllocatedResourceIdentity, revision string, budget *intrinsicMetadataBudget) (resourceget.ResourceDocument, error) {
	var backing, current string
	if err := tx.QueryRowContext(ctx, `SELECT backing,revision FROM graph_preview_catalog WHERE path=?`, identity.PrivatePath).Scan(&backing, &current); err != nil {
		return resourceget.ResourceDocument{}, err
	}
	switch backing {
	case "intrinsic", "intrinsic-link", "type-definition":
		var native IntrinsicDocument
		if err := s.readSelectedDefinitionInTx(ctx, tx, identity.PrivatePath, revision, &native, budget); err != nil {
			return resourceget.ResourceDocument{}, err
		}
		if native.Kind == graph.KindLink {
			return resourceget.NewResourceDocument(s.ScopeURL(), native.Kind, native.Body)
		}
		document, err := definitionDocumentForNative(s.ScopeURL(), native)
		if err != nil {
			return resourceget.ResourceDocument{}, err
		}
		return document.Resource()
	}
	capture, err := s.captureTypeRecordsInTx(ctx, tx, budget)
	if err != nil {
		return resourceget.ResourceDocument{}, err
	}
	ctx = s.contextForResourceCapture(ctx, tx, capture, budget)
	if backing == "installed-type" || backing == "installed-type-link" {
		uri, err := graph.CanonicalPublicResourceURL(s.ScopeURL(), identity.PublicID)
		if err != nil {
			return resourceget.ResourceDocument{}, err
		}
		for _, document := range capture.snapshot.Documents {
			if document.ID() == uri {
				if revision == "" || revision == document.Revision() {
					return document, nil
				}
				var retained int
				if err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM graph_preview_versions WHERE path=? AND version=?`, identity.PrivatePath, revision).Scan(&retained); err != nil {
					return resourceget.ResourceDocument{}, err
				}
				if retained != 0 {
					return resourceget.ResourceDocument{}, fmt.Errorf("%w: installed Type retained edition needs its recorded contract", ErrInvalidStore)
				}
				return resourceget.ResourceDocument{}, ErrVersionUnknown
			}
		}
		return resourceget.ResourceDocument{}, fmt.Errorf("%w: installed allocation absent from checked inventory", ErrInvalidStore)
	}
	if revision == "" {
		revision = current
	}
	var record any
	if revision == current && identity.State == "live" {
		record, err = s.readCurrentVerifiedInTx(ctx, tx, identity.PrivatePath)
	} else {
		record, err = s.readVersionInTx(ctx, tx, identity.PrivatePath, revision)
	}
	if err != nil {
		return resourceget.ResourceDocument{}, err
	}
	return s.projectCheckedNativeInTx(ctx, tx, identity.PrivatePath, revision, record, capture.installed, budget)
}

func (s *Store) CurrentResourceDocuments(ctx context.Context) (ResourceSnapshot, error) {
	if s.LayoutVersion() != TypesFreshSchemaVersion {
		return ResourceSnapshot{}, ErrCapabilityUnavailable
	}
	var result ResourceSnapshot
	err := s.withDefinitionReadTx(ctx, txDefinitionList, func(tx *sql.Tx, budget *intrinsicMetadataBudget) error {
		var err error
		result, err = s.currentResourceDocumentsInTx(ctx, tx, budget)
		return err
	})
	if err != nil {
		return ResourceSnapshot{}, err
	}
	return result, nil
}

// SelectBeadDocuments selects from this already captured Resource inventory.
// Selectors are literal semantic Type URIs, as in the raw paging predicate.
// A product caller resolving a local definition identity must call ResolveType
// on THIS same capture first and pass its checked semantic URI. This method
// neither resolves aliases nor reacquires storage or substitutes current pins.
func (s ResourceSnapshot) SelectBeadDocuments(request BeadListRequest) (BeadDocumentPage, error) {
	return pageResourceBeads(s, request)
}

func pageResourceBeads(snapshot ResourceSnapshot, request BeadListRequest) (BeadDocumentPage, error) {
	if request.Limit < 0 || request.MaxRows < 0 {
		return BeadDocumentPage{}, fmt.Errorf("%w: negative Bead bounds", graph.ErrValidation)
	}
	for _, selector := range []string{request.TypeURL, request.ConformsTo} {
		if selector == "" {
			continue
		}
		known := false
		for _, fact := range snapshot.Types {
			if fact.Current && fact.Reference.URI == selector {
				known = true
			}
		}
		if !known {
			return BeadDocumentPage{}, fmt.Errorf("%w: unknown semantic Type selector", ErrCapabilityUnavailable)
		}
	}
	result := BeadDocumentPage{Items: []resourceget.ResourceDocument{}}
	for _, document := range snapshot.Documents {
		if document.Kind() != graph.KindBead {
			continue
		}
		nominal, err := snapshot.MatchesType(document.Type(), request.TypeURL, true)
		if err != nil {
			return BeadDocumentPage{}, err
		}
		family, err := snapshot.MatchesType(document.Type(), request.ConformsTo, false)
		if err != nil {
			return BeadDocumentPage{}, err
		}
		if nominal && family {
			result.Items = append(result.Items, document)
		}
	}
	if request.MaxRows > 0 && len(result.Items) > request.MaxRows {
		return BeadDocumentPage{}, ErrLimitExceeded
	}
	if request.Limit > 0 && len(result.Items) > request.Limit {
		result.Items = result.Items[:request.Limit]
		result.HasMore = true
	}
	return result, nil
}

func (s *Store) ListBeadDocuments(ctx context.Context, request BeadListRequest) (BeadDocumentPage, error) {
	snapshot, err := s.CurrentResourceDocuments(ctx)
	if err != nil {
		return BeadDocumentPage{}, err
	}
	// Under14 this is ordinary Bead inventory; old Issue hiding policy is only
	// applied by the explicit native-Issue filter seam. All does not omit Types.
	return snapshot.SelectBeadDocuments(request)
}

func (s *Store) CountBeadDocuments(ctx context.Context, request BeadListRequest) (int, error) {
	request.Limit = 0
	page, err := s.ListBeadDocuments(ctx, request)
	if err != nil {
		return 0, err
	}
	return len(page.Items), nil
}

func (s *Store) ListIssueDocuments(ctx context.Context, request publicops.ListRequest, parent *IssueSelector) (BeadDocumentPage, error) {
	if s.LayoutVersion() != TypesFreshSchemaVersion {
		return BeadDocumentPage{}, ErrCapabilityUnavailable
	}
	if request.IncludeComments {
		return BeadDocumentPage{}, fmt.Errorf("%w: common Resource projection does not include unversioned Issue comment enrichment", ErrCapabilityUnavailable)
	}
	prepared, err := prepareIssueListRequest(request)
	if err != nil {
		return BeadDocumentPage{}, err
	}
	if parent != nil {
		if err = validateIssueSelector(*parent); err != nil {
			return BeadDocumentPage{}, err
		}
	}
	var result BeadDocumentPage
	err = s.withDefinitionReadTx(ctx, txDefinitionList, func(tx *sql.Tx, budget *intrinsicMetadataBudget) error {
		capture, e := s.captureTypeRecordsInTx(ctx, tx, budget)
		if e != nil {
			return e
		}
		ctx = s.contextForResourceCapture(ctx, tx, capture, budget)
		if e = chargeCommonNativeRowsInTx(ctx, tx, budget); e != nil {
			return e
		}
		selected, e := s.listIssuesWithParentInTx(ctx, tx, &prepared, parent)
		if e != nil {
			return e
		}
		result = BeadDocumentPage{Items: []resourceget.ResourceDocument{}, HasMore: selected.HasMore}
		for _, record := range selected.Items {
			identity, e := s.resolveCommonResourceSelectorInTx(ctx, tx, record.ID, "", budget)
			if e != nil {
				return e
			}
			document, e := s.projectCheckedNativeInTx(ctx, tx, identity.PrivatePath, record.Revision, record, capture.installed, budget)
			if e != nil {
				return e
			}
			result.Items = append(result.Items, document)
		}
		return nil
	})
	if err != nil {
		return BeadDocumentPage{}, err
	}
	return result, nil
}

func (s *Store) CountIssueDocuments(ctx context.Context, request publicops.ListRequest, parent *IssueSelector) (int, error) {
	zero := 0
	request.Limit = &zero
	request.Offset = 0
	page, err := s.ListIssueDocuments(ctx, request, parent)
	if err != nil {
		return 0, err
	}
	if page.HasMore {
		return 0, ErrLimitExceeded
	}
	return len(page.Items), nil
}

// Compile-time coupling of the concrete Store and the frozen receiving seam.
var _ ResourceDocumentReader = (*Store)(nil)
