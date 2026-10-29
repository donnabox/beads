package graphstore

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"unicode/utf8"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/intrinsicmaterialization"
	"github.com/steveyegge/beads/resourceget"
)

const customDefinitionEncoding = "type-definition-context/1"
const generatedDefinitionProperties = `{"propertiesSchema":{"$schema":"https://json-schema.org/draft/2020-12/schema","type":"object"}}`

type customDefinitionContext struct {
	Encoding  string                             `json:"encoding"`
	ContextID string                             `json:"contextID"`
	Type      intrinsicmaterialization.Reference `json:"type"`
}

// CreateBeadType is the first small ordinary definition writer, not general
// schema admission or registration. Only the fixed generated object schema is
// supported. The six intrinsic members and their captured records never change.
func (s *Store) CreateBeadType(ctx context.Context, id, actor string) (resourceget.DefinitionDocument, error) {
	if s.options.Binding.SchemaVersion != TypesFreshSchemaVersion {
		return resourceget.DefinitionDocument{}, ErrCapabilityUnavailable
	}
	if err := validatePublicID(id); err != nil {
		return resourceget.DefinitionDocument{}, err
	}
	if IsReservedTypeName(id) {
		return resourceget.DefinitionDocument{}, fmt.Errorf("%w: reserved intrinsic Type name", graph.ErrValidation)
	}
	if !utf8.ValidString(actor) {
		return resourceget.DefinitionDocument{}, fmt.Errorf("%w: invalid actor encoding", graph.ErrValidation)
	}
	budget, err := newIntrinsicMetadataBudget(PreviewCurrentReadByteLimit)
	if err != nil {
		return resourceget.DefinitionDocument{}, err
	}
	var result resourceget.DefinitionDocument
	err = s.withTxPolicy(ctx, txPolicy{purpose: txDefinitionCreate, allowance: budget, mayWrite: true, commit: true}, func(tx *sql.Tx) error {
		contextID, e := currentIntrinsicContextInTx(ctx, tx, budget)
		if e != nil {
			return e
		}
		plan, e := loadIntrinsicContextInTx(ctx, tx, s.options, contextID, budget)
		if e != nil {
			return e
		}
		ownType, e := customDefinitionAffiliation(plan)
		if e != nil {
			return e
		}
		if e = budget.chargeBody(8*(len(s.ScopeURL())+len(id)) + 4*len(actor) + 16384); e != nil {
			return e
		}
		if e = s.afterStage("definition-validated"); e != nil {
			return e
		}
		if e = s.touchCoordination(ctx, tx); e != nil {
			return e
		}
		exists, e := publicResourceIDExistsInTx(ctx, tx, id)
		if e != nil {
			return e
		}
		if exists {
			return publicResourceIDReservedErrorInTx(ctx, tx, id)
		}
		key, e := freshToken()
		if e != nil {
			return e
		}
		revision, e := freshToken()
		if e != nil {
			return e
		}
		path := "beads/" + key
		doc, e := generatedBeadTypeDocument(s.ScopeURL(), id, revision, ownType)
		if e != nil {
			return e
		}
		codec, e := encodeCustomDefinitionContext(contextID, ownType)
		if e != nil {
			return e
		}
		body := doc.NativeJSON()
		if _, e = tx.ExecContext(ctx, `INSERT INTO graph_preview_catalog (path,resource_kind,type_url,revision,allocation_state,backing,backing_key,public_id) VALUES (?,?,?,?,'live','type-definition',?,?)`, path, string(graph.KindBead), ownType.URI, revision, key, id); e != nil {
			return e
		}
		if e = s.afterStage("definition-catalog"); e != nil {
			return e
		}
		if _, e = tx.ExecContext(ctx, `INSERT INTO graph_preview_payloads (path,properties,metadata) VALUES (?,?,?)`, path, body, codec); e != nil {
			return e
		}
		if e = s.afterStage("definition-payload"); e != nil {
			return e
		}
		if e = insertPreviewVersionInTx(ctx, tx, path, revision, body, actor); e != nil {
			return e
		}
		if e = s.afterStage("definition-version"); e != nil {
			return e
		}
		checked, e := readCustomDefinitionInTx(ctx, tx, s.options, path, revision, plan, budget)
		if e != nil {
			return e
		}
		if !bytes.Equal(checked.NativeJSON(), body) {
			return fmt.Errorf("%w: new Type readback differs", ErrInvalidStore)
		}
		if e = s.afterStage("definition-readback"); e != nil {
			return e
		}
		result = checked
		return nil
	})
	if err != nil {
		return resourceget.DefinitionDocument{}, err
	}
	return result, nil
}

func customDefinitionAffiliation(plan *intrinsicmaterialization.Plan) (intrinsicmaterialization.Reference, error) {
	if plan == nil {
		return intrinsicmaterialization.Reference{}, ErrInvalidStore
	}
	// This exact published pair must have a witness in the checked mapping;
	// LocalFor returns only the reference and avoids cloning all six bodies.
	ref := intrinsicmaterialization.Reference{URI: "https://github.com/gastownhall/bdp/types/type-system/bead-type", Revision: "bead-type-1"}
	if _, ok := plan.LocalFor(ref); !ok {
		return intrinsicmaterialization.Reference{}, ErrInvalidStore
	}
	return ref, nil
}

func generatedBeadTypeDocument(scope, id, revision string, ownType intrinsicmaterialization.Reference) (resourceget.DefinitionDocument, error) {
	uri, err := graph.CanonicalPublicResourceURL(scope, id)
	if err != nil {
		return resourceget.DefinitionDocument{}, err
	}
	body := struct {
		ID         string                             `json:"id"`
		Revision   string                             `json:"revision"`
		Type       intrinsicmaterialization.Reference `json:"type"`
		Properties json.RawMessage                    `json:"properties"`
		OwnedLinks map[string]json.RawMessage         `json:"ownedLinks"`
	}{uri, revision, ownType, json.RawMessage(generatedDefinitionProperties), map[string]json.RawMessage{}}
	raw, err := json.Marshal(body)
	if err != nil {
		return resourceget.DefinitionDocument{}, err
	}
	return resourceget.NewDefinitionDocument(scope, resourceget.DefinitionBead, raw)
}

func encodeCustomDefinitionContext(contextID string, ownType intrinsicmaterialization.Reference) ([]byte, error) {
	if !authorityID.MatchString(contextID) || ownType.URI == "" || ownType.Revision == "" {
		return nil, ErrInvalidStore
	}
	raw, err := json.Marshal(customDefinitionContext{customDefinitionEncoding, contextID, ownType})
	if err != nil {
		return nil, err
	}
	return graph.CanonicalizeJSON(raw)
}

func decodeCustomDefinitionContext(raw []byte) (customDefinitionContext, error) {
	if len(raw) > 512 {
		return customDefinitionContext{}, fmt.Errorf("%w: custom definition context bytes", ErrLimitExceeded)
	}
	canonical, err := graph.CanonicalizeJSON(raw)
	if err != nil {
		return customDefinitionContext{}, fmt.Errorf("%w: custom definition context JSON: %v", ErrInvalidStore, err)
	}
	var result customDefinitionContext
	dec := json.NewDecoder(bytes.NewReader(canonical))
	dec.DisallowUnknownFields()
	if err = dec.Decode(&result); err != nil {
		return customDefinitionContext{}, fmt.Errorf("%w: custom definition context: %v", ErrInvalidStore, err)
	}
	var extra any
	if err = dec.Decode(&extra); err != io.EOF {
		return customDefinitionContext{}, ErrInvalidStore
	}
	expected, err := encodeCustomDefinitionContext(result.ContextID, result.Type)
	if err != nil || result.Encoding != customDefinitionEncoding || !bytes.Equal(canonical, expected) {
		return customDefinitionContext{}, fmt.Errorf("%w: custom definition context contract", ErrInvalidStore)
	}
	return result, nil
}

// The caller holds actual same-Tx profile admission. A nonnil known plan MUST
// come from loadIntrinsicContextInTx in this transaction; the stored exact
// context must agree. A retained read loads its recorded context when nil.
func readCustomDefinitionInTx(ctx context.Context, tx *sql.Tx, o Options, path, requested string, known *intrinsicmaterialization.Plan, budget *intrinsicMetadataBudget) (resourceget.DefinitionDocument, error) {
	if o.Binding.SchemaVersion != TypesFreshSchemaVersion {
		return resourceget.DefinitionDocument{}, ErrCapabilityUnavailable
	}
	if err := budget.chargeBody(4096); err != nil {
		return resourceget.DefinitionDocument{}, err
	}
	var count, payloadCount, total uint64
	err := tx.QueryRowContext(ctx, `SELECT COUNT(*),COUNT(p.path),CAST(LEAST(COALESCE(SUM(512+COALESCE(OCTET_LENGTH(c.public_id),0)+COALESCE(OCTET_LENGTH(c.type_url),0)+COALESCE(OCTET_LENGTH(p.properties),0)+COALESCE(OCTET_LENGTH(p.metadata),0)),0),?) AS UNSIGNED) FROM graph_preview_catalog c LEFT JOIN graph_preview_payloads p ON p.path=c.path WHERE c.path=? AND c.backing='type-definition'`, budget.remaining+1, path).Scan(&count, &payloadCount, &total)
	if err != nil {
		return resourceget.DefinitionDocument{}, err
	}
	if count != 1 || payloadCount != 1 {
		return resourceget.DefinitionDocument{}, fmt.Errorf("%w: custom definition allocation/payload missing", ErrInvalidStore)
	}
	if total > uint64(budget.remaining/8) {
		return resourceget.DefinitionDocument{}, ErrLimitExceeded
	}
	if err = budget.chargeBody(8 * int(total)); err != nil {
		return resourceget.DefinitionDocument{}, err
	}
	var id, key, revision, typ, state string
	var metadata []byte
	var bodyLength uint64
	// Acquire only the bounded context/header before validating its recorded
	// intrinsic meaning. The full custom body follows that correspondence.
	err = tx.QueryRowContext(ctx, `SELECT c.public_id,c.backing_key,c.revision,c.type_url,c.allocation_state,p.metadata,OCTET_LENGTH(p.properties) FROM graph_preview_catalog c JOIN graph_preview_payloads p ON p.path=c.path WHERE c.path=? AND c.backing='type-definition' AND c.resource_kind='bead' AND OCTET_LENGTH(p.metadata)<=512`, path).Scan(&id, &key, &revision, &typ, &state, &metadata, &bodyLength)
	if err != nil {
		return resourceget.DefinitionDocument{}, fmt.Errorf("%w: custom definition header: %w", ErrInvalidStore, err)
	}
	if state != "live" || !authorityID.MatchString(key) || !authorityID.MatchString(revision) || path != intrinsicBeadBackingPath(key) || validatePublicID(id) != nil {
		return resourceget.DefinitionDocument{}, fmt.Errorf("%w: custom definition allocation identity", ErrInvalidStore)
	}
	if requested != "" && requested != revision {
		var evidence int
		if err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM graph_preview_versions WHERE path=? AND version=?`, path, requested).Scan(&evidence); err != nil {
			return resourceget.DefinitionDocument{}, err
		}
		if evidence == 0 {
			return resourceget.DefinitionDocument{}, ErrVersionUnknown
		}
		return resourceget.DefinitionDocument{}, fmt.Errorf("%w: unsupported custom definition edition", ErrInvalidStore)
	}
	codec, err := decodeCustomDefinitionContext(metadata)
	if err != nil {
		return resourceget.DefinitionDocument{}, err
	}
	if known == nil {
		known, err = loadIntrinsicContextInTx(ctx, tx, o, codec.ContextID, budget)
		if err != nil {
			return resourceget.DefinitionDocument{}, err
		}
	}
	scope := known.Scope()
	if scope.ContextID != codec.ContextID || scope.URI != o.Binding.ScopeURL || scope.AuthorityID != o.Binding.AuthorityID {
		return resourceget.DefinitionDocument{}, fmt.Errorf("%w: custom definition recorded context differs", ErrInvalidStore)
	}
	ownType, err := customDefinitionAffiliation(known)
	if err != nil {
		return resourceget.DefinitionDocument{}, err
	}
	if codec.Type != ownType || typ != ownType.URI {
		return resourceget.DefinitionDocument{}, fmt.Errorf("%w: custom definition affiliation differs", ErrInvalidStore)
	}
	expected, err := generatedBeadTypeDocument(o.Binding.ScopeURL, id, revision, ownType)
	if err != nil {
		return resourceget.DefinitionDocument{}, fmt.Errorf("%w: generated definition: %v", ErrInvalidStore, err)
	}
	expectedBytes := expected.NativeJSON()
	if bodyLength != uint64(len(expectedBytes)) {
		return resourceget.DefinitionDocument{}, fmt.Errorf("%w: custom definition body length differs", ErrInvalidStore)
	}
	var raw []byte
	if err = tx.QueryRowContext(ctx, `SELECT properties FROM graph_preview_payloads WHERE path=? AND OCTET_LENGTH(properties)=?`, path, len(expectedBytes)).Scan(&raw); err != nil {
		return resourceget.DefinitionDocument{}, fmt.Errorf("%w: custom definition body: %w", ErrInvalidStore, err)
	}
	if !bytes.Equal(raw, expectedBytes) {
		return resourceget.DefinitionDocument{}, fmt.Errorf("%w: custom definition body/schema differs", ErrInvalidStore)
	}
	var versions int
	if err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM graph_preview_versions WHERE path=?`, path).Scan(&versions); err != nil {
		return resourceget.DefinitionDocument{}, err
	}
	if versions != 1 {
		return resourceget.DefinitionDocument{}, fmt.Errorf("%w: unsupported custom definition history", ErrInvalidStore)
	}
	if err = budget.chargeBody(2 * len(raw)); err != nil {
		return resourceget.DefinitionDocument{}, err
	}
	var snapshot []byte
	err = tx.QueryRowContext(ctx, `SELECT snapshot FROM graph_preview_versions WHERE path=? AND version=? AND ordinal=1 AND OCTET_LENGTH(snapshot)=?`, path, revision, len(raw)).Scan(&snapshot)
	if err != nil {
		return resourceget.DefinitionDocument{}, fmt.Errorf("%w: custom definition retained snapshot: %w", ErrInvalidStore, err)
	}
	if !bytes.Equal(snapshot, raw) {
		return resourceget.DefinitionDocument{}, fmt.Errorf("%w: custom definition retained snapshot differs", ErrInvalidStore)
	}
	return expected, nil
}

// Dispatch uses the actual stored backing, never the caller ID's spelling.
func (s *Store) readSelectedDefinitionInTx(ctx context.Context, tx *sql.Tx, path, revision string, out *IntrinsicDocument, budget *intrinsicMetadataBudget) error {
	if err := budget.chargeBody(1024); err != nil {
		return err
	}
	var backing string
	err := tx.QueryRowContext(ctx, `SELECT backing FROM graph_preview_catalog WHERE path=?`, path).Scan(&backing)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if backing == "intrinsic" || backing == "intrinsic-link" {
		return s.readIntrinsicDocumentInTx(ctx, tx, path, revision, out, budget)
	}
	if backing != "type-definition" {
		return fmt.Errorf("%w: unsupported Types definition backing", ErrInvalidStore)
	}
	doc, err := readCustomDefinitionInTx(ctx, tx, s.options, path, revision, nil, budget)
	if err != nil {
		return err
	}
	id, err := graph.ScopeLocalPublicResourceID(s.ScopeURL(), doc.ID())
	if err != nil {
		return ErrInvalidStore
	}
	*out = IntrinsicDocument{PublicID: id, PrivatePath: path, Kind: graph.KindBead, Revision: doc.Revision(), Body: doc.NativeJSON()}
	return nil
}

// Acquire every custom allocation, including unsupported terminal rows (which
// must refuse), before selection. LEFT JOIN/count agreement catches missing
// payloads instead of silently omitting them. Cardinality follows the shared
// byte allowance and per-row bookkeeping; no new ID grammar/count policy.
func (s *Store) listCustomDefinitionsInTx(ctx context.Context, tx *sql.Tx, plan *intrinsicmaterialization.Plan, budget *intrinsicMetadataBudget) ([]resourceget.DefinitionDocument, error) {
	if err := budget.chargeBody(4096); err != nil {
		return nil, err
	}
	var count, payloadCount, total uint64
	err := tx.QueryRowContext(ctx, `SELECT COUNT(*),COUNT(p.path),CAST(LEAST(COALESCE(SUM(512+OCTET_LENGTH(c.path)+COALESCE(OCTET_LENGTH(c.public_id),0)+COALESCE(OCTET_LENGTH(p.properties),0)+COALESCE(OCTET_LENGTH(p.metadata),0)),0),?) AS UNSIGNED) FROM graph_preview_catalog c LEFT JOIN graph_preview_payloads p ON p.path=c.path WHERE c.backing='type-definition'`, budget.remaining+1).Scan(&count, &payloadCount, &total)
	if err != nil {
		return nil, err
	}
	if count != payloadCount {
		return nil, fmt.Errorf("%w: custom definition inventory/payload differs", ErrInvalidStore)
	}
	if count > uint64(budget.remaining/512) || total > uint64(budget.remaining/8) {
		return nil, ErrLimitExceeded
	}
	if err = budget.chargeBody(8 * int(total)); err != nil {
		return nil, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT path FROM graph_preview_catalog WHERE backing='type-definition' ORDER BY BINARY path LIMIT ?`, count+1)
	if err != nil {
		return nil, err
	}
	paths := make([]string, 0, int(count))
	for rows.Next() {
		if uint64(len(paths)) >= count {
			err = fmt.Errorf("%w: custom definition inventory changed", ErrInvalidStore)
			break
		}
		var path string
		if err = rows.Scan(&path); err != nil {
			break
		}
		if validatePath(path) != nil {
			err = ErrInvalidStore
			break
		}
		paths = append(paths, path)
	}
	if err == nil && rows.NextResultSet() {
		err = fmt.Errorf("%w: extra custom inventory result set", ErrInvalidStore)
	}
	err = errors.Join(err, rows.Err(), rows.Close())
	if err != nil {
		return nil, err
	}
	if uint64(len(paths)) != count {
		return nil, fmt.Errorf("%w: custom inventory census differs", ErrInvalidStore)
	}
	result := make([]resourceget.DefinitionDocument, 0, len(paths))
	for _, path := range paths {
		doc, e := readCustomDefinitionInTx(ctx, tx, s.options, path, "", plan, budget)
		if e != nil {
			return nil, e
		}
		result = append(result, doc)
	}
	return result, nil
}
