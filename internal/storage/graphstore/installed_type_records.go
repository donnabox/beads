package graphstore

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/intrinsicmaterialization"
	"github.com/steveyegge/beads/internal/memorycatalog"
	"github.com/steveyegge/beads/internal/typecatalog"
	"github.com/steveyegge/beads/resourceget"
)

// installedTypeRecordContext is private backing data for an ordinary Type
// Bead. Descriptor is a checked index of THIS record, not another public
// allocation. The complete compiler contract and source-only domain obligations
// stay with the immutable edition; none may be filled from a newer catalog.
type installedTypeRecordContext struct {
	Codec            string                    `json:"codec"`
	IntrinsicContext string                    `json:"intrinsicContext"`
	Descriptor       json.RawMessage           `json:"descriptor"`
	Fingerprint      string                    `json:"fingerprint"`
	Contract         typecatalog.Definition    `json:"contract"`
	Memory           *memorycatalog.Definition `json:"memory,omitempty"`
	RecordSHA256     string                    `json:"recordSHA256"`
}

type installedTypeCandidate struct {
	source      installedTypeSource
	publicID    string
	path        string
	key         string
	reference   resourceget.TypeReference
	affiliation resourceget.TypeReference
	contract    typecatalog.Definition
	document    resourceget.DefinitionDocument
	links       []installedTypeLinkCandidate
}

type installedTypeLinkCandidate struct {
	publicID  string
	path      string
	reference resourceget.TypeReference
	record    json.RawMessage
}

func exactResourceReference(ref intrinsicmaterialization.Reference) resourceget.TypeReference {
	return resourceget.TypeReference{URI: ref.URI, Revision: ref.Revision}
}

// prepareInstalledTypeCandidates runs only after the initializer has checked
// its complete installed descriptor set. It chooses all final Type editions
// before forming ancestry Links and owner captures. It performs no writes.
func prepareInstalledTypeCandidates(ctx context.Context, tx *sql.Tx, o Options, plan *intrinsicmaterialization.Plan, budget *intrinsicMetadataBudget) ([]installedTypeCandidate, error) {
	if plan == nil || plan.Scope().URI != o.Binding.ScopeURL || plan.Scope().AuthorityID != o.Binding.AuthorityID {
		return nil, ErrInvalidStore
	}
	sources, err := selectedInstalledTypeSources(o.Binding.ScopeURL)
	if err != nil {
		return nil, err
	}
	return prepareInstalledTypeSources(ctx, tx, o, plan, budget, sources, nil)
}

func prepareInstalledTypeSources(ctx context.Context, tx *sql.Tx, o Options, plan *intrinsicmaterialization.Plan, budget *intrinsicMetadataBudget, sources []installedTypeSource, existing map[string]resourceget.TypeReference) ([]installedTypeCandidate, error) {
	if plan == nil || plan.Scope().URI != o.Binding.ScopeURL || plan.Scope().AuthorityID != o.Binding.AuthorityID {
		return nil, ErrInvalidStore
	}
	// Fixed source planning is charged before allocations/copies. Dynamic
	// first-use retains the same operation allowance and never raises its cap.
	if err := budget.chargeBody(4 * intrinsicmaterialization.MaxContextBytes); err != nil {
		return nil, err
	}
	if len(sources)+len(existing) > PreviewSnapshotLimit {
		return nil, ErrLimitExceeded
	}
	refs := make(map[string]resourceget.TypeReference, len(sources)+len(existing))
	for uri, ref := range existing {
		refs[uri] = ref
	}
	candidates := make([]installedTypeCandidate, 0, len(sources))
	var err error
	reserved := map[string]bool{}
	for _, source := range sources {
		id, err := graph.ScopeLocalPublicResourceID(o.Binding.ScopeURL, source.Descriptor.ID())
		if err != nil {
			return nil, err
		}
		// The already-checked descriptor owns this exact identity. Only its
		// matching definition allocation may share it; existing Resources and
		// deleted allocations are still lifetime conflicts.
		owner, err := allocatedResourceIDOwnerInTx(ctx, tx, id)
		if err != nil {
			return nil, err
		}
		if owner != "" {
			return nil, fmt.Errorf("%w: installed definition identity is already allocated", ErrAlreadyExists)
		}
		key, err := freshToken()
		if err != nil {
			return nil, err
		}
		revision, err := freshToken()
		if err != nil {
			return nil, err
		}
		ref := resourceget.TypeReference{URI: source.Descriptor.ID(), Revision: revision}
		if _, exists := refs[ref.URI]; exists {
			return nil, ErrInvalidStore
		}
		refs[ref.URI] = ref
		reserved[id] = true
		candidates = append(candidates, installedTypeCandidate{source: source, publicID: id, path: "beads/" + key, key: key, reference: ref})
	}
	beadType, ok := plan.GlobalForRole(intrinsicmaterialization.BeadType)
	if !ok {
		return nil, ErrInvalidStore
	}
	linkType, ok := plan.GlobalForRole(intrinsicmaterialization.LinkType)
	if !ok {
		return nil, ErrInvalidStore
	}
	refines, ok := plan.GlobalForRole(intrinsicmaterialization.Refines)
	if !ok {
		return nil, ErrInvalidStore
	}
	for i := range candidates {
		candidate := &candidates[i]
		candidate.contract, err = bindInstalledTypeContract(candidate.source.Contract, refs)
		if err != nil {
			return nil, err
		}
		properties, err := installedTypeProperties(candidate.source, candidate.contract)
		if err != nil {
			return nil, err
		}
		// Compile the actual stored schema bytes, including their selected
		// base/dialect annotations, rather than an independently rebuilt root.
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(properties, &fields); err != nil {
			return nil, err
		}
		if schema, present := fields["propertiesSchema"]; present {
			candidate.contract.PropertiesSchema = schema
		}
		owned := map[string][]json.RawMessage{}
		for _, parent := range candidate.contract.Parents {
			key, err := freshToken()
			if err != nil {
				return nil, err
			}
			id := key
			if reserved[id] {
				return nil, ErrAlreadyExists
			}
			reserved[id] = true
			uri, err := graph.CanonicalPublicResourceURL(o.Binding.ScopeURL, id)
			if err != nil {
				return nil, err
			}
			if exists, err := publicResourceIDExistsInTx(ctx, tx, id); err != nil {
				return nil, err
			} else if exists {
				return nil, publicResourceIDReservedErrorInTx(ctx, tx, id)
			}
			revision, err := freshToken()
			if err != nil {
				return nil, err
			}
			record, err := canonicalJSON(map[string]any{"id": uri, "revision": revision, "type": exactResourceReference(refines), "source": candidate.reference.URI, "target": resourceget.TypeReference{URI: parent.URI, Revision: parent.Revision}, "properties": map[string]any{}})
			if err != nil {
				return nil, err
			}
			if _, err := resourceget.NewDefinitionDocument(o.Binding.ScopeURL, resourceget.DefinitionLink, record); err != nil {
				return nil, err
			}
			candidate.links = append(candidate.links, installedTypeLinkCandidate{publicID: id, path: "links/" + key, reference: resourceget.TypeReference{URI: uri, Revision: revision}, record: record})
			owned[refines.URI] = append(owned[refines.URI], record)
		}
		ownType := beadType
		if candidate.contract.Kind == graph.KindLink {
			ownType = linkType
		}
		candidate.affiliation = exactResourceReference(ownType)
		if err := budget.chargeBody(8 * (len(properties) + len(o.Binding.ScopeURL) + len(candidate.publicID) + 4096)); err != nil {
			return nil, err
		}
		body, err := canonicalJSON(map[string]any{"id": candidate.reference.URI, "revision": candidate.reference.Revision, "type": exactResourceReference(ownType), "properties": properties, "ownedLinks": owned})
		if err != nil {
			return nil, err
		}
		candidate.document, err = resourceget.NewDefinitionDocument(o.Binding.ScopeURL, resourceget.DefinitionBead, body)
		if err != nil {
			return nil, err
		}
	}
	return candidates, nil
}

// installOrdinaryTypeDefinitionsInTx joins fresh bootstrap's transaction. It
// creates no schema objects and never adds roles to the six-member context.
// The caller owns the same writer fence as intrinsic bootstrap and all rows
// commit/roll back together with final definition editions and owner captures.
func installOrdinaryTypeDefinitionsInTx(ctx context.Context, tx *sql.Tx, o Options, plan *intrinsicmaterialization.Plan, budget *intrinsicMetadataBudget) error {
	candidates, err := prepareInstalledTypeCandidates(ctx, tx, o, plan, budget)
	if err != nil {
		return err
	}
	return installTypeCandidatesInTx(ctx, tx, plan, budget, candidates)
}

func installTypeCandidatesInTx(ctx context.Context, tx *sql.Tx, plan *intrinsicmaterialization.Plan, budget *intrinsicMetadataBudget, candidates []installedTypeCandidate) error {
	for _, candidate := range candidates {
		body := candidate.document.NativeJSON()
		codec := installedTypeRecordContext{Codec: "installed-type-v1", IntrinsicContext: plan.Scope().ContextID, Descriptor: candidate.source.Descriptor.CanonicalJSON(), Fingerprint: candidate.source.Descriptor.Fingerprint(), Contract: candidate.contract, Memory: candidate.source.Memory, RecordSHA256: intrinsicDigest(body)}
		metadata, err := canonicalJSON(codec)
		if err != nil {
			return err
		}
		if err := budget.chargeBody(4 * len(metadata)); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO graph_preview_catalog (path,resource_kind,type_url,revision,allocation_state,backing,backing_key,public_id) VALUES (?,'bead',?,?,'live','installed-type',?,?)`, candidate.path, candidate.affiliation.URI, candidate.reference.Revision, candidate.key, candidate.publicID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO graph_preview_payloads (path,properties,metadata) VALUES (?,?,?)`, candidate.path, body, metadata); err != nil {
			return err
		}
		if err := insertPreviewVersionInTx(ctx, tx, candidate.path, candidate.reference.Revision, body, "type-bootstrap"); err != nil {
			return err
		}
		if err := budget.chargeBody(4*len(body) + 2*len(metadata) + 4096); err != nil {
			return err
		}
		var storedBody, storedMetadata, snapshot []byte
		if err := tx.QueryRowContext(ctx, `SELECT p.properties,p.metadata,v.snapshot FROM graph_preview_catalog c JOIN graph_preview_payloads p ON p.path=c.path JOIN graph_preview_versions v ON v.path=c.path AND v.version=c.revision WHERE c.path=? AND c.resource_kind='bead' AND c.backing='installed-type' AND c.backing_key=? AND c.public_id=? AND OCTET_LENGTH(c.public_id)=? AND c.type_url=? AND OCTET_LENGTH(c.type_url)=? AND c.revision=? AND c.allocation_state='live' AND OCTET_LENGTH(p.properties)=? AND OCTET_LENGTH(p.metadata)=? AND OCTET_LENGTH(v.snapshot)=?`, candidate.path, candidate.key, candidate.publicID, len(candidate.publicID), candidate.affiliation.URI, len(candidate.affiliation.URI), candidate.reference.Revision, len(body), len(metadata), len(body)).Scan(&storedBody, &storedMetadata, &snapshot); err != nil {
			return fmt.Errorf("%w: installed Type readback: %w", ErrInvalidStore, err)
		}
		if !bytes.Equal(storedBody, body) || !bytes.Equal(storedMetadata, metadata) || !bytes.Equal(snapshot, body) {
			return fmt.Errorf("%w: installed Type readback differs", ErrInvalidStore)
		}
		for _, link := range candidate.links {
			refines, ok := plan.GlobalForRole(intrinsicmaterialization.Refines)
			if !ok {
				return ErrInvalidStore
			}
			if _, err := tx.ExecContext(ctx, `INSERT INTO graph_preview_catalog (path,resource_kind,type_url,revision,allocation_state,backing,public_id) VALUES (?,'link',?,?,'live','installed-type-link',?)`, link.path, refines.URI, link.reference.Revision, link.publicID); err != nil {
				return err
			}
			if err := insertPreviewVersionInTx(ctx, tx, link.path, link.reference.Revision, link.record, "type-bootstrap"); err != nil {
				return err
			}
			if err := budget.chargeBody(2*len(link.record) + 4096); err != nil {
				return err
			}
			var stored []byte
			if err := tx.QueryRowContext(ctx, `SELECT v.snapshot FROM graph_preview_catalog c JOIN graph_preview_versions v ON v.path=c.path AND v.version=c.revision WHERE c.path=? AND c.resource_kind='link' AND c.backing='installed-type-link' AND c.public_id=? AND OCTET_LENGTH(c.public_id)=? AND c.type_url=? AND OCTET_LENGTH(c.type_url)=? AND c.revision=? AND c.allocation_state='live' AND OCTET_LENGTH(v.snapshot)=?`, link.path, link.publicID, len(link.publicID), refines.URI, len(refines.URI), link.reference.Revision, len(link.record)).Scan(&stored); err != nil {
				return fmt.Errorf("%w: installed declaration readback: %w", ErrInvalidStore, err)
			}
			if !bytes.Equal(stored, link.record) {
				return fmt.Errorf("%w: installed declaration readback differs", ErrInvalidStore)
			}
		}
	}
	return nil
}
