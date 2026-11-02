package graphstore

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"slices"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/intrinsicmaterialization"
	"github.com/steveyegge/beads/internal/intrinsictypecatalog"
	"github.com/steveyegge/beads/resourceget"
)

type checkedResourceCapture struct {
	plan      *intrinsicmaterialization.Plan
	snapshot  ResourceSnapshot
	installed checkedInstalledTypes
}

// captureTypeRecordsInTx builds neither a second registry nor a Type-only
// reader island. These are the actual complete stored definition Resources and
// their exact checked index facts, reused within THIS transaction only.
func (s *Store) captureTypeRecordsInTx(ctx context.Context, tx *sql.Tx, budget *intrinsicMetadataBudget) (checkedResourceCapture, error) {
	id, err := currentIntrinsicContextInTx(ctx, tx, budget)
	if err != nil {
		return checkedResourceCapture{}, err
	}
	plan, err := loadIntrinsicContextInTx(ctx, tx, s.options, id, budget)
	if err != nil {
		return checkedResourceCapture{}, err
	}
	// The six checked members are the complete intrinsic allocation family.
	// An unrecorded extra allocation must not hide behind mapping allowances.
	var intrinsicBeads, intrinsicLinks int
	if err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM graph_preview_catalog WHERE backing='intrinsic'`).Scan(&intrinsicBeads); err != nil {
		return checkedResourceCapture{}, err
	}
	if err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM graph_preview_catalog WHERE backing='intrinsic-link'`).Scan(&intrinsicLinks); err != nil {
		return checkedResourceCapture{}, err
	}
	if intrinsicBeads != 4 || intrinsicLinks != 2 {
		return checkedResourceCapture{}, ErrInvalidStore
	}
	installed, err := loadInstalledTypesInTx(ctx, tx, s.options, plan, budget)
	if err != nil {
		return checkedResourceCapture{}, err
	}
	if err = budget.chargeBody(4*plan.Source().Bytes + 2*intrinsicmaterialization.MaxContextBytes); err != nil {
		return checkedResourceCapture{}, err
	}
	global, err := intrinsictypecatalog.LoadSource(plan.Source())
	if err != nil {
		return checkedResourceCapture{}, err
	}
	result := checkedResourceCapture{plan: plan, installed: installed, snapshot: ResourceSnapshot{Documents: []resourceget.ResourceDocument{}, Types: []ResourceTypeFact{}}}
	result.snapshot.Documents = append(result.snapshot.Documents, installed.documents...)
	result.snapshot.Documents = append(result.snapshot.Documents, installed.links...)
	result.snapshot.Types = append(result.snapshot.Types, installed.facts...)
	for _, binding := range plan.Bindings() {
		allocation := binding.Allocation
		if err = budget.chargeBody(8*len(binding.Record) + 8192); err != nil {
			return checkedResourceCapture{}, err
		}
		if allocation.Kind == graph.KindLink {
			document, err := resourceget.NewResourceDocument(s.ScopeURL(), graph.KindLink, binding.Record)
			if err != nil {
				return checkedResourceCapture{}, err
			}
			result.snapshot.Documents = append(result.snapshot.Documents, document)
			continue
		}
		native := IntrinsicDocument{PublicID: allocation.PublicID, PrivatePath: allocation.PrivatePath, Kind: allocation.Kind, Revision: allocation.Local.Revision, ContextID: id, Global: intrinsictypecatalog.Reference{URI: binding.Global.URI, Revision: binding.Global.Revision}, Body: binding.Record}
		definition, err := definitionDocumentForNative(s.ScopeURL(), native)
		if err != nil {
			return checkedResourceCapture{}, err
		}
		document, err := definition.Resource()
		if err != nil {
			return checkedResourceCapture{}, err
		}
		result.snapshot.Documents = append(result.snapshot.Documents, document)
		semantic := intrinsictypecatalog.Reference{URI: binding.Global.URI, Revision: binding.Global.Revision}
		kind, known := global.Category(semantic)
		if !known {
			return checkedResourceCapture{}, ErrInvalidStore
		}
		parents, known := global.Ancestors(semantic)
		if !known {
			return checkedResourceCapture{}, ErrInvalidStore
		}
		ref := resourceget.TypeReference{URI: semantic.URI, Revision: semantic.Revision}
		ancestors := []resourceget.TypeReference{ref}
		for _, parent := range parents {
			ancestors = append(ancestors, resourceget.TypeReference{URI: parent.URI, Revision: parent.Revision})
		}
		var original struct {
			Properties struct {
				Abstract bool `json:"abstract"`
			} `json:"properties"`
		}
		if err = json.Unmarshal(binding.Original, &original); err != nil {
			return checkedResourceCapture{}, err
		}
		result.snapshot.Types = append(result.snapshot.Types, ResourceTypeFact{Reference: ref, Definition: resourceget.TypeReference{URI: allocation.Local.URI, Revision: allocation.Local.Revision}, Describes: kind, Abstract: original.Properties.Abstract, Current: true, Ancestors: ancestors})
	}
	custom, err := s.listCustomDefinitionsInTx(ctx, tx, plan, budget)
	if err != nil {
		return checkedResourceCapture{}, err
	}
	for _, definition := range custom {
		// The existing native constructor has already checked the exact generated
		// BeadType body/omission rule. This is not general authored-Type admission.
		if err = budget.chargeBody(8 * (len(definition.ID()) + len(definition.Revision()) + len(s.ScopeURL()) + 4096)); err != nil {
			return checkedResourceCapture{}, err
		}
		document, err := definition.Resource()
		if err != nil {
			return checkedResourceCapture{}, err
		}
		result.snapshot.Documents = append(result.snapshot.Documents, document)
		ref := resourceget.TypeReference{URI: definition.ID(), Revision: definition.Revision()}
		result.snapshot.Types = append(result.snapshot.Types, ResourceTypeFact{Reference: ref, Definition: ref, Describes: graph.KindBead, Current: true, Ancestors: []resourceget.TypeReference{ref}})
	}
	return result, nil
}

func (s *Store) currentResourceDocumentsInTx(ctx context.Context, tx *sql.Tx, budget *intrinsicMetadataBudget) (ResourceSnapshot, error) {
	return s.currentResourceDocumentsAtLimitInTx(ctx, tx, budget, PreviewSnapshotLimit)
}

// Both caps are package-owned. Query separately bounds ALL allocations and
// persisted bytes before entering this same complete capture.
func (s *Store) currentResourceDocumentsAtLimitInTx(ctx context.Context, tx *sql.Tx, budget *intrinsicMetadataBudget, limit int) (ResourceSnapshot, error) {
	if limit != PreviewSnapshotLimit && limit != resourceQueryAllocationLimit {
		return ResourceSnapshot{}, ErrInvalidStore
	}
	if err := ctx.Err(); err != nil {
		return ResourceSnapshot{}, err
	}
	capture, err := s.captureTypeRecordsInTx(ctx, tx, budget)
	if err != nil {
		return ResourceSnapshot{}, err
	}
	ctx = s.contextForResourceCapture(ctx, tx, capture, budget)
	if err = s.checkCollectionMappingsInTx(ctx, tx); err != nil {
		return ResourceSnapshot{}, err
	}
	if err = chargeCommonNativeRowsInTx(ctx, tx, budget); err != nil {
		return ResourceSnapshot{}, err
	}
	if err = checkAllocatedIdentityBytesInBudget(ctx, tx, budget, "allocation_state='live'"); err != nil {
		return ResourceSnapshot{}, err
	}
	result := capture.snapshot
	if err = tx.QueryRowContext(ctx, `SELECT writer_token FROM graph_preview_scope WHERE singleton=1`).Scan(&result.WriterToken); err != nil {
		return ResourceSnapshot{}, err
	}
	if !authorityID.MatchString(result.WriterToken) {
		return ResourceSnapshot{}, ErrInvalidStore
	}
	var invalid int
	if err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM graph_preview_catalog WHERE allocation_state NOT IN ('live','deleted')`).Scan(&invalid); err != nil {
		return ResourceSnapshot{}, err
	}
	if invalid != 0 {
		return ResourceSnapshot{}, ErrInvalidStore
	}
	identities, err := readCommonLiveIdentitiesInTx(ctx, tx, limit)
	if err != nil {
		return ResourceSnapshot{}, err
	}
	byID := map[string]resourceget.ResourceDocument{}
	for _, document := range result.Documents {
		if _, exists := byID[document.ID()]; exists {
			return ResourceSnapshot{}, fmt.Errorf("%w: duplicate checked definition Resource", ErrInvalidStore)
		}
		byID[document.ID()] = document
	}
	seen := map[string]bool{}
	for _, identity := range identities {
		if err := ctx.Err(); err != nil {
			return ResourceSnapshot{}, err
		}
		uri, err := graph.CanonicalPublicResourceURL(s.ScopeURL(), identity.PublicID)
		if err != nil {
			return ResourceSnapshot{}, err
		}
		if seen[uri] {
			return ResourceSnapshot{}, fmt.Errorf("%w: duplicate common Resource public identity", ErrInvalidStore)
		}
		seen[uri] = true
		if document, known := byID[uri]; known {
			if document.Kind() != identity.Kind {
				return ResourceSnapshot{}, ErrInvalidStore
			}
			continue
		}
		record, err := s.readCurrentVerifiedInTx(ctx, tx, identity.PrivatePath)
		if err != nil {
			return ResourceSnapshot{}, err
		}
		var revision string
		switch value := record.(type) {
		case Record:
			revision = value.Revision
		case IssueRecord:
			revision = value.Revision
		case LinkRecord:
			revision = value.Revision
		default:
			return ResourceSnapshot{}, ErrInvalidStore
		}
		document, err := s.projectCheckedNativeInTx(ctx, tx, identity.PrivatePath, revision, record, capture.installed, budget)
		if err != nil {
			return ResourceSnapshot{}, err
		}
		if document.ID() != uri || document.Kind() != identity.Kind {
			return ResourceSnapshot{}, ErrInvalidStore
		}
		result.Documents = append(result.Documents, document)
	}
	for id := range byID {
		if !seen[id] {
			return ResourceSnapshot{}, fmt.Errorf("%w: checked definition missing live allocation", ErrInvalidStore)
		}
	}
	slices.SortFunc(result.Documents, func(a, b resourceget.ResourceDocument) int { return graph.CompareCodeUnits(a.ID(), b.ID()) })
	if err := ctx.Err(); err != nil {
		return ResourceSnapshot{}, err
	}
	return result, nil
}

// Reserve native body/copy work before query hydration/projection, independently
// of existing native-domain validators' own acquisition limits.
func chargeCommonNativeRowsInTx(ctx context.Context, tx *sql.Tx, budget *intrinsicMetadataBudget) error {
	var size uint64
	err := tx.QueryRowContext(ctx, `SELECT CAST(LEAST(COALESCE(SUM(n),0),?) AS UNSIGNED) FROM (
 SELECT 8192+12*(OCTET_LENGTH(v.snapshot)+OCTET_LENGTH(v.actor)) AS n FROM graph_preview_catalog c JOIN graph_preview_versions v ON v.path=c.path AND v.version=c.revision WHERE c.backing IN ('generic','dependency','informational') AND c.allocation_state='live'
 UNION ALL
 SELECT 8192+12*(OCTET_LENGTH(v.durable_state)+OCTET_LENGTH(m.owned)+COALESCE(OCTET_LENGTH(m.revision_context),0)) AS n FROM graph_preview_catalog c JOIN graph_preview_issue_versions m ON m.path=c.path AND m.version=c.revision JOIN issue_versions v ON v.issue_id=m.issue_id AND v.revision=m.issue_revision WHERE c.backing='issue' AND c.allocation_state='live'
 ) bounded_native_resources`, PreviewCurrentReadByteLimit+1).Scan(&size)
	if err != nil {
		return err
	}
	if size > PreviewCurrentReadByteLimit {
		return ErrLimitExceeded
	}
	return budget.chargeBody(int(size))
}

func (s *Store) projectCheckedNativeInTx(ctx context.Context, tx *sql.Tx, path, revision string, record any, installed checkedInstalledTypes, budget *intrinsicMetadataBudget) (resourceget.ResourceDocument, error) {
	var backing string
	if err := tx.QueryRowContext(ctx, `SELECT backing FROM graph_preview_catalog WHERE path=?`, path).Scan(&backing); err != nil {
		return resourceget.ResourceDocument{}, err
	}
	var size uint64
	var raw []byte
	var retainedDomain []byte
	var pin resourceget.TypeReference
	if backing == "issue" {
		if err := tx.QueryRowContext(ctx, `SELECT OCTET_LENGTH(revision_context) FROM graph_preview_issue_versions WHERE path=? AND version=?`, path, revision).Scan(&size); err != nil {
			return resourceget.ResourceDocument{}, fmt.Errorf("%w: Issue exact affiliation evidence: %v", ErrInvalidStore, err)
		}
		if size > PreviewCurrentReadByteLimit {
			return resourceget.ResourceDocument{}, ErrLimitExceeded
		}
		if err := budget.chargeBody(4*int(size) + 8192); err != nil {
			return resourceget.ResourceDocument{}, err
		}
		if err := tx.QueryRowContext(ctx, `SELECT revision_context FROM graph_preview_issue_versions WHERE path=? AND version=? AND OCTET_LENGTH(revision_context)=?`, path, revision, size).Scan(&raw); err != nil {
			return resourceget.ResourceDocument{}, err
		}
		_, selected, err := decodeNativeIssueTypeEdition(raw)
		if err != nil {
			return resourceget.ResourceDocument{}, err
		}
		pin = selected
	} else if backing == "generic" || backing == "dependency" || backing == "informational" {
		if err := tx.QueryRowContext(ctx, `SELECT OCTET_LENGTH(snapshot) FROM graph_preview_versions WHERE path=? AND version=?`, path, revision).Scan(&size); err != nil {
			return resourceget.ResourceDocument{}, err
		}
		if size > PreviewCurrentReadByteLimit {
			return resourceget.ResourceDocument{}, ErrLimitExceeded
		}
		if err := budget.chargeBody(4*int(size) + 8192); err != nil {
			return resourceget.ResourceDocument{}, err
		}
		if err := tx.QueryRowContext(ctx, `SELECT snapshot FROM graph_preview_versions WHERE path=? AND version=? AND OCTET_LENGTH(snapshot)=?`, path, revision, size).Scan(&raw); err != nil {
			return resourceget.ResourceDocument{}, err
		}
		domain, selected, err := decodeNativeSnapshotEdition(raw)
		if err != nil {
			return resourceget.ResourceDocument{}, err
		}
		pin = selected
		retainedDomain = domain
	} else {
		return resourceget.ResourceDocument{}, ErrInvalidStore
	}
	// Earlier acquisition debits include this complete domain serialization.
	domain, err := canonicalJSON(record)
	if err != nil {
		return resourceget.ResourceDocument{}, err
	}
	if backing != "issue" && !bytes.Equal(domain, retainedDomain) {
		return resourceget.ResourceDocument{}, fmt.Errorf("%w: applied Resource differs from retained edition", ErrInvalidStore)
	}
	kind := graph.KindBead
	if _, link := record.(LinkRecord); link {
		kind = graph.KindLink
	}
	return projectNativeResourceInTx(ctx, tx, s.ScopeURL(), kind, domain, pin, installed, budget)
}
