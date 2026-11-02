package graphstore

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/intrinsicmaterialization"
	"github.com/steveyegge/beads/internal/memorycatalog"
	"github.com/steveyegge/beads/internal/typecatalog"
	"github.com/steveyegge/beads/resourceget"
)

// checkedInstalledTypes belongs to one read transaction. Its executable
// definitions are decoded from the SAME stored editions as Documents, after
// their source obligations, descriptor index and complete bodies agree.
// Initializer source is a corruption check, never a moving runtime replacement.
type checkedInstalledTypes struct {
	documents []resourceget.ResourceDocument
	links     []resourceget.ResourceDocument
	facts     []ResourceTypeFact
	catalog   *typecatalog.Catalog
	memories  map[string]memorycatalog.Definition
}

type installedTypeAcquisition struct {
	path, publicID, revision, typeURI, key string
	bodyBytes, metadataBytes               uint64
	body, metadata                         []byte
	context                                installedTypeRecordContext
	descriptor                             graph.TypeDescriptor
	document                               resourceget.DefinitionDocument
}

// loadInstalledTypesInTx does not open a transaction, admit a writer or change
// the six-member intrinsic context. All acquisition and copies debit its
// caller's operation allowance; row/cell limits are not driver heap ceilings.
func loadInstalledTypesInTx(ctx context.Context, tx *sql.Tx, o Options, plan *intrinsicmaterialization.Plan, budget *intrinsicMetadataBudget) (checkedInstalledTypes, error) {
	if plan == nil || plan.Scope().URI != o.Binding.ScopeURL || plan.Scope().AuthorityID != o.Binding.AuthorityID || o.Binding.SchemaVersion != TypesFreshSchemaVersion {
		return checkedInstalledTypes{}, ErrInvalidStore
	}
	if err := budget.chargeBody(4 * intrinsicmaterialization.MaxContextBytes); err != nil {
		return checkedInstalledTypes{}, err
	}
	sources, err := selectedInstalledTypeSourcesInTx(ctx, tx, o.Binding.ScopeURL)
	if err != nil {
		return checkedInstalledTypes{}, err
	}
	var count, total uint64
	err = tx.QueryRowContext(ctx, `SELECT COUNT(*),CAST(LEAST(COALESCE(SUM(OCTET_LENGTH(p.properties)+OCTET_LENGTH(p.metadata)+OCTET_LENGTH(c.public_id)+OCTET_LENGTH(c.type_url)),0),?) AS UNSIGNED) FROM graph_preview_catalog c JOIN graph_preview_payloads p ON p.path=c.path WHERE c.backing='installed-type'`, intrinsicmaterialization.MaxContextBytes+1).Scan(&count, &total)
	if err != nil {
		return checkedInstalledTypes{}, err
	}
	if count != uint64(len(sources)) || total > intrinsicmaterialization.MaxContextBytes {
		return checkedInstalledTypes{}, fmt.Errorf("%w: installed Type acquisition bounds", ErrInvalidStore)
	}
	var allocationCount uint64
	if err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM graph_preview_catalog WHERE backing='installed-type'`).Scan(&allocationCount); err != nil {
		return checkedInstalledTypes{}, err
	}
	if allocationCount != count {
		return checkedInstalledTypes{}, fmt.Errorf("%w: installed Type missing payload", ErrInvalidStore)
	}
	// Retained raw bodies, metadata decode, source comparison, compiler inputs,
	// snapshots and immutable document constructors are charged BEFORE fetch.
	if err = budget.chargeBody(12*int(total) + len(sources)*8192); err != nil {
		return checkedInstalledTypes{}, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT c.path,c.public_id,c.revision,c.type_url,c.backing_key,c.resource_kind,c.allocation_state,OCTET_LENGTH(p.properties),OCTET_LENGTH(p.metadata) FROM graph_preview_catalog c JOIN graph_preview_payloads p ON p.path=c.path WHERE c.backing='installed-type' ORDER BY c.path LIMIT ?`, len(sources)+1)
	if err != nil {
		return checkedInstalledTypes{}, err
	}
	acquired := make([]installedTypeAcquisition, 0, len(sources))
	for rows.Next() {
		var row installedTypeAcquisition
		var kind, state string
		if err = rows.Scan(&row.path, &row.publicID, &row.revision, &row.typeURI, &row.key, &kind, &state, &row.bodyBytes, &row.metadataBytes); err != nil {
			break
		}
		if kind != string(graph.KindBead) || state != "live" || !authorityID.MatchString(row.key) || row.path != "beads/"+row.key || !authorityID.MatchString(row.revision) || row.bodyBytes == 0 || row.metadataBytes == 0 || row.bodyBytes > intrinsicmaterialization.MaxContextBytes || row.metadataBytes > intrinsicmaterialization.MaxContextBytes {
			err = fmt.Errorf("%w: installed Type allocation", ErrInvalidStore)
			break
		}
		acquired = append(acquired, row)
	}
	if err == nil && rows.NextResultSet() {
		err = fmt.Errorf("%w: extra installed Type acquisition result", ErrInvalidStore)
	}
	err = errors.Join(err, rows.Err(), rows.Close())
	if err != nil {
		return checkedInstalledTypes{}, err
	}
	if len(acquired) != len(sources) {
		return checkedInstalledTypes{}, ErrInvalidStore
	}
	refs := make(map[string]resourceget.TypeReference, len(acquired))
	sourceByID := make(map[string]installedTypeSource, len(sources))
	for _, source := range sources {
		sourceByID[source.Descriptor.ID()] = source
	}
	for i := range acquired {
		row := &acquired[i]
		err = tx.QueryRowContext(ctx, `SELECT properties,metadata FROM graph_preview_payloads WHERE path=? AND OCTET_LENGTH(properties)=? AND OCTET_LENGTH(metadata)=?`, row.path, row.bodyBytes, row.metadataBytes).Scan(&row.body, &row.metadata)
		if err != nil {
			return checkedInstalledTypes{}, err
		}
		decoder := json.NewDecoder(bytes.NewReader(row.metadata))
		decoder.DisallowUnknownFields()
		if err = decoder.Decode(&row.context); err != nil {
			return checkedInstalledTypes{}, fmt.Errorf("%w: installed Type metadata: %w", ErrInvalidStore, err)
		}
		var extra any
		if err = decoder.Decode(&extra); err != io.EOF {
			return checkedInstalledTypes{}, fmt.Errorf("%w: installed Type metadata trailing input", ErrInvalidStore)
		}
		encoded, err := canonicalJSON(row.context)
		if err != nil || !bytes.Equal(encoded, row.metadata) {
			return checkedInstalledTypes{}, fmt.Errorf("%w: installed Type metadata encoding", ErrInvalidStore)
		}
		if row.context.Codec != "installed-type-v1" || row.context.IntrinsicContext != plan.Scope().ContextID || row.context.RecordSHA256 != intrinsicDigest(row.body) {
			return checkedInstalledTypes{}, fmt.Errorf("%w: installed Type recorded source", ErrInvalidStore)
		}
		row.descriptor, err = graph.ParseTypeDescriptor(row.context.Descriptor)
		if err != nil {
			return checkedInstalledTypes{}, fmt.Errorf("%w: installed Type index: %w", ErrInvalidStore, err)
		}
		source, ok := sourceByID[row.descriptor.ID()]
		if !ok || row.context.Fingerprint != row.descriptor.Fingerprint() || !bytes.Equal(row.descriptor.CanonicalJSON(), source.Descriptor.CanonicalJSON()) {
			return checkedInstalledTypes{}, fmt.Errorf("%w: installed Type source/index correspondence", ErrInvalidStore)
		}
		var descriptor []byte
		var fingerprint string
		if err = tx.QueryRowContext(ctx, `SELECT descriptor,fingerprint FROM graph_preview_types WHERE name=? AND OCTET_LENGTH(descriptor)=?`, source.Name, len(row.context.Descriptor)).Scan(&descriptor, &fingerprint); err != nil || fingerprint != row.context.Fingerprint || !bytes.Equal(descriptor, row.context.Descriptor) {
			return checkedInstalledTypes{}, fmt.Errorf("%w: installed Type descriptor index: %w", ErrInvalidStore, err)
		}
		uri, err := graph.CanonicalPublicResourceURL(o.Binding.ScopeURL, row.publicID)
		if err != nil || uri != row.descriptor.ID() || row.context.Contract.Identity.URI != uri || row.context.Contract.Identity.Revision != row.revision {
			return checkedInstalledTypes{}, fmt.Errorf("%w: installed Type identity correspondence", ErrInvalidStore)
		}
		if _, duplicate := refs[uri]; duplicate {
			return checkedInstalledTypes{}, fmt.Errorf("%w: duplicate installed Type identity", ErrInvalidStore)
		}
		refs[uri] = resourceget.TypeReference{URI: uri, Revision: row.revision}
		var identities int
		if err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM graph_preview_catalog WHERE public_id=? AND OCTET_LENGTH(public_id)=?`, row.publicID, len(row.publicID)).Scan(&identities); err != nil {
			return checkedInstalledTypes{}, err
		}
		if identities != 1 {
			return checkedInstalledTypes{}, fmt.Errorf("%w: ambiguous installed Type public ID", ErrInvalidStore)
		}
	}
	return finishInstalledTypesInTx(ctx, tx, o, plan, budget, acquired, sourceByID, refs)
}

func finishInstalledTypesInTx(ctx context.Context, tx *sql.Tx, o Options, plan *intrinsicmaterialization.Plan, budget *intrinsicMetadataBudget, acquired []installedTypeAcquisition, sources map[string]installedTypeSource, refs map[string]resourceget.TypeReference) (checkedInstalledTypes, error) {
	result := checkedInstalledTypes{memories: map[string]memorycatalog.Definition{}}
	definitions := make([]typecatalog.Definition, 0, len(acquired))
	for i := range acquired {
		row := &acquired[i]
		source := sources[row.descriptor.ID()]
		expected, err := bindInstalledTypeContract(source.Contract, refs)
		if err != nil {
			return checkedInstalledTypes{}, err
		}
		properties, err := installedTypeProperties(source, expected)
		if err != nil {
			return checkedInstalledTypes{}, err
		}
		var propertyFields map[string]json.RawMessage
		if err = json.Unmarshal(properties, &propertyFields); err != nil {
			return checkedInstalledTypes{}, err
		}
		if schema, present := propertyFields["propertiesSchema"]; present {
			expected.PropertiesSchema = schema
		}
		wantContract, err := canonicalJSON(expected)
		if err != nil {
			return checkedInstalledTypes{}, err
		}
		gotContract, err := canonicalJSON(row.context.Contract)
		if err != nil || !bytes.Equal(wantContract, gotContract) {
			return checkedInstalledTypes{}, fmt.Errorf("%w: installed Type executable contract correspondence", ErrInvalidStore)
		}
		wantMemory, err := canonicalJSON(source.Memory)
		if err != nil {
			return checkedInstalledTypes{}, err
		}
		gotMemory, err := canonicalJSON(row.context.Memory)
		if err != nil || !bytes.Equal(wantMemory, gotMemory) {
			return checkedInstalledTypes{}, fmt.Errorf("%w: installed Type native obligations", ErrInvalidStore)
		}
		var record struct {
			ID         string                       `json:"id"`
			Revision   string                       `json:"revision"`
			Type       resourceget.TypeReference    `json:"type"`
			Properties json.RawMessage              `json:"properties"`
			OwnedLinks map[string][]json.RawMessage `json:"ownedLinks"`
		}
		if err = json.Unmarshal(row.body, &record); err != nil {
			return checkedInstalledTypes{}, fmt.Errorf("%w: installed Type body: %w", ErrInvalidStore, err)
		}
		role := intrinsicmaterialization.BeadType
		if expected.Kind == graph.KindLink {
			role = intrinsicmaterialization.LinkType
		}
		global, ok := plan.GlobalForRole(role)
		if !ok || record.ID != row.descriptor.ID() || record.Revision != row.revision || record.Type != exactResourceReference(global) || row.typeURI != record.Type.URI || !bytes.Equal(record.Properties, properties) {
			return checkedInstalledTypes{}, fmt.Errorf("%w: installed Type body/index correspondence", ErrInvalidStore)
		}
		row.document, err = resourceget.NewDefinitionDocument(o.Binding.ScopeURL, resourceget.DefinitionBead, row.body)
		if err != nil {
			return checkedInstalledTypes{}, fmt.Errorf("%w: installed Type complete body: %w", ErrInvalidStore, err)
		}
		var snapshot []byte
		if err = tx.QueryRowContext(ctx, `SELECT snapshot FROM graph_preview_versions WHERE path=? AND version=? AND OCTET_LENGTH(snapshot)=?`, row.path, row.revision, row.bodyBytes).Scan(&snapshot); err != nil || !bytes.Equal(snapshot, row.body) {
			return checkedInstalledTypes{}, fmt.Errorf("%w: installed Type retained edition: %w", ErrInvalidStore, err)
		}
		document, err := row.document.Resource()
		if err != nil {
			return checkedInstalledTypes{}, err
		}
		result.documents = append(result.documents, document)
		definitions = append(definitions, row.context.Contract)
		if row.context.Memory != nil {
			result.memories[row.descriptor.ID()] = *row.context.Memory
		}
	}
	// The compiler consumes the checked STORED contracts, never expected/source
	// replacements. It closes exact parents, selectors, schemas and domain kinds.
	catalog, err := typecatalog.New(definitions)
	if err != nil {
		return checkedInstalledTypes{}, fmt.Errorf("%w: installed Type compiler: %w", ErrInvalidStore, err)
	}
	result.catalog = catalog
	for _, definition := range definitions {
		closure, err := catalog.Closure(definition.Identity)
		if err != nil {
			return checkedInstalledTypes{}, err
		}
		ancestors := make([]resourceget.TypeReference, 0, len(closure))
		for _, ref := range closure {
			ancestors = append(ancestors, resourceget.TypeReference{URI: ref.URI, Revision: ref.Revision})
		}
		ref := resourceget.TypeReference{URI: definition.Identity.URI, Revision: definition.Identity.Revision}
		result.facts = append(result.facts, ResourceTypeFact{Reference: ref, Definition: ref, Describes: definition.Kind, Abstract: definition.Abstract, Current: true, Ancestors: ancestors})
	}
	// Real Refines captures and their allocations/snapshots must be checked by
	// the same transaction before this capture is exposed or a writer admitted.
	if err = verifyInstalledTypeLinksInTx(ctx, tx, o, plan, budget, acquired, &result); err != nil {
		return checkedInstalledTypes{}, err
	}
	return result, nil
}

func verifyInstalledTypeLinksInTx(ctx context.Context, tx *sql.Tx, o Options, plan *intrinsicmaterialization.Plan, budget *intrinsicMetadataBudget, acquired []installedTypeAcquisition, result *checkedInstalledTypes) error {
	global, ok := plan.GlobalForRole(intrinsicmaterialization.Refines)
	if !ok {
		return ErrInvalidStore
	}
	seen := map[string]bool{}
	for _, row := range acquired {
		var body struct {
			OwnedLinks map[string][]json.RawMessage `json:"ownedLinks"`
		}
		if err := json.Unmarshal(row.body, &body); err != nil {
			return fmt.Errorf("%w: installed Type captures: %w", ErrInvalidStore, err)
		}
		parents := row.context.Contract.Parents
		if len(parents) == 0 {
			if len(body.OwnedLinks) != 0 {
				return fmt.Errorf("%w: unexpected installed Type declaration", ErrInvalidStore)
			}
			continue
		}
		links, present := body.OwnedLinks[global.URI]
		if !present || len(body.OwnedLinks) != 1 || len(links) != len(parents) {
			return fmt.Errorf("%w: incomplete installed Type declarations", ErrInvalidStore)
		}
		matched := map[typecatalog.Reference]bool{}
		for _, raw := range links {
			if err := budget.chargeBody(8*len(raw) + 8192); err != nil {
				return err
			}
			var link struct {
				ID       string                    `json:"id"`
				Revision string                    `json:"revision"`
				Type     resourceget.TypeReference `json:"type"`
				Source   string                    `json:"source"`
				Target   resourceget.TypeReference `json:"target"`
			}
			if err := json.Unmarshal(raw, &link); err != nil {
				return fmt.Errorf("%w: installed Type declaration body: %w", ErrInvalidStore, err)
			}
			target := typecatalog.Reference{URI: link.Target.URI, Revision: link.Target.Revision}
			if link.Type != exactResourceReference(global) || link.Source != row.descriptor.ID() || !authorityID.MatchString(link.Revision) || matched[target] || seen[link.ID] {
				return fmt.Errorf("%w: installed Type declaration identity/capture", ErrInvalidStore)
			}
			found := false
			for _, parent := range parents {
				if parent == target {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("%w: installed Type undeclared parent", ErrInvalidStore)
			}
			matched[target] = true
			seen[link.ID] = true
			expected, err := canonicalJSON(map[string]any{"id": link.ID, "revision": link.Revision, "type": link.Type, "source": link.Source, "target": link.Target, "properties": map[string]any{}})
			if err != nil || !bytes.Equal(expected, raw) {
				return fmt.Errorf("%w: installed Type full declaration correspondence", ErrInvalidStore)
			}
			publicID, err := graph.ScopeLocalPublicResourceID(o.Binding.ScopeURL, link.ID)
			if err != nil {
				return fmt.Errorf("%w: installed Type declaration scope: %w", ErrInvalidStore, err)
			}
			var path, kind, typeURI, revision, state, backing string
			var key sql.NullString
			var allocationCount int
			if err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM graph_preview_catalog WHERE public_id=? AND OCTET_LENGTH(public_id)=?`, publicID, len(publicID)).Scan(&allocationCount); err != nil {
				return err
			}
			if allocationCount != 1 {
				return fmt.Errorf("%w: ambiguous installed declaration public ID", ErrInvalidStore)
			}
			err = tx.QueryRowContext(ctx, `SELECT path,resource_kind,type_url,revision,allocation_state,backing,backing_key FROM graph_preview_catalog WHERE public_id=? AND OCTET_LENGTH(public_id)=? AND OCTET_LENGTH(type_url)=?`, publicID, len(publicID), len(global.URI)).Scan(&path, &kind, &typeURI, &revision, &state, &backing, &key)
			if err != nil || kind != string(graph.KindLink) || typeURI != global.URI || revision != link.Revision || state != "live" || backing != "installed-type-link" || key.Valid || validateLinkPath(path) != nil {
				return fmt.Errorf("%w: installed declaration allocation: %w", ErrInvalidStore, err)
			}
			var payloadCount int
			if err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM graph_preview_payloads WHERE path=?`, path).Scan(&payloadCount); err != nil {
				return err
			}
			if payloadCount != 0 {
				return fmt.Errorf("%w: installed declaration has competing payload", ErrInvalidStore)
			}
			var snapshot []byte
			if err = tx.QueryRowContext(ctx, `SELECT snapshot FROM graph_preview_versions WHERE path=? AND version=? AND OCTET_LENGTH(snapshot)=?`, path, link.Revision, len(raw)).Scan(&snapshot); err != nil || !bytes.Equal(snapshot, raw) {
				return fmt.Errorf("%w: installed declaration retained capture: %w", ErrInvalidStore, err)
			}
			resource, err := resourceget.NewResourceDocument(o.Binding.ScopeURL, graph.KindLink, raw)
			if err != nil {
				return err
			}
			result.links = append(result.links, resource)
		}
	}
	var count int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM graph_preview_catalog WHERE backing='installed-type-link'`).Scan(&count); err != nil {
		return err
	}
	if count != len(seen) {
		return fmt.Errorf("%w: orphan installed declaration", ErrInvalidStore)
	}
	return nil
}
