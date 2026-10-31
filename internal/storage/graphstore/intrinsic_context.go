package graphstore

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/intrinsicmaterialization"
	"github.com/steveyegge/beads/internal/intrinsictypecatalog"
)

// loadIntrinsicContextInTx validates stored mapping against the immutable source
// and all six real allocations/snapshots in this transaction. Digests are
// corruption controls, never caller authority. No current/global fallback.
func loadIntrinsicContextInTx(ctx context.Context, tx *sql.Tx, o Options, id string, budget *intrinsicMetadataBudget) (*intrinsicmaterialization.Plan, error) {
	if o.Binding.SchemaVersion != TypesFreshSchemaVersion || !authorityID.MatchString(id) {
		return nil, ErrInvalidStore
	}
	// Acquire a bounded header first. Neither current source nor caller-provided
	// digest may reinterpret a recorded context or authorize its body bytes.
	if err := budget.chargeBody(4*len(o.Binding.ScopeURL) + 16384); err != nil {
		return nil, err
	}
	var scope, authority, repo, commit, path, digest, semantic, interpretation, contextDigest string
	var sourceBytes uint64
	err := tx.QueryRowContext(ctx, `SELECT scope_url,authority_id,source_repository,source_commit,source_path,source_sha256,OCTET_LENGTH(source_bytes),semantic_source_commit,interpretation,context_sha256 FROM graph_preview_intrinsic_contexts WHERE context_id=? AND OCTET_LENGTH(scope_url)=? AND OCTET_LENGTH(source_repository)<=256 AND OCTET_LENGTH(source_path)<=1024`, id, len(o.Binding.ScopeURL)).Scan(&scope, &authority, &repo, &commit, &path, &digest, &sourceBytes, &semantic, &interpretation, &contextDigest)
	if err != nil {
		return nil, fmt.Errorf("%w: intrinsic context header: %v", ErrInvalidStore, err)
	}
	if scope != o.Binding.ScopeURL || authority != o.Binding.AuthorityID {
		return nil, fmt.Errorf("%w: intrinsic context authority", ErrInvalidStore)
	}
	source, err := intrinsicContextSource(repo, commit, path, digest, sourceBytes, semantic, interpretation)
	if err != nil {
		return nil, err
	}
	// Four source/header copies and two context buffers remain charged before
	// loading, fetching or reconstructing their complete bytes.
	if err = budget.chargeBody(4*source.Bytes + 2*intrinsicmaterialization.MaxContextBytes); err != nil {
		return nil, err
	}
	catalog, err := intrinsictypecatalog.LoadSource(source)
	if err != nil {
		return nil, err
	}
	var raw []byte
	err = tx.QueryRowContext(ctx, `SELECT source_bytes FROM graph_preview_intrinsic_contexts WHERE context_id=? AND OCTET_LENGTH(source_bytes)=?`, id, source.Bytes).Scan(&raw)
	if err != nil || !bytes.Equal(raw, catalog.Bytes()) {
		return nil, fmt.Errorf("%w: intrinsic source bytes: %v", ErrInvalidStore, err)
	}
	var count, total uint64
	if err = tx.QueryRowContext(ctx, `SELECT COUNT(*),CAST(LEAST(COALESCE(SUM(OCTET_LENGTH(record)+OCTET_LENGTH(public_id)+OCTET_LENGTH(resource_uri)+OCTET_LENGTH(global_uri)+OCTET_LENGTH(global_revision)),0),?) AS UNSIGNED) FROM graph_preview_intrinsic_members WHERE context_id=?`, intrinsicmaterialization.MaxContextBytes+1, id).Scan(&count, &total); err != nil {
		return nil, err
	}
	if count != 6 || total > intrinsicmaterialization.MaxContextBytes {
		return nil, fmt.Errorf("%w: intrinsic member acquisition", ErrInvalidStore)
	}
	// Eight retained member/assembler copies; independent1MiB structural cap
	// above remains, and cannot substitute for this shared operation debit.
	if err := budget.chargeBody(8*int(total) + 6*4096); err != nil {
		return nil, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT path,revision,allocation_key,role,public_id,resource_uri,resource_kind,global_uri,global_revision,record,record_sha256 FROM graph_preview_intrinsic_members WHERE context_id=? LIMIT 7`, id)
	if err != nil {
		return nil, err
	}
	type persisted struct {
		allocation intrinsicmaterialization.Allocation
		key        string
		global     intrinsicmaterialization.Reference
		record     []byte
		digest     string
	}
	members := make([]persisted, 0, 6)
	allocations := make([]intrinsicmaterialization.Allocation, 0, 6)
	for rows.Next() {
		var m persisted
		m.allocation.AuthorityID = authority
		if err = rows.Scan(&m.allocation.PrivatePath, &m.allocation.Local.Revision, &m.key, &m.allocation.Role, &m.allocation.PublicID, &m.allocation.Local.URI, &m.allocation.Kind, &m.global.URI, &m.global.Revision, &m.record, &m.digest); err != nil {
			break
		}
		if !authorityID.MatchString(m.key) || m.digest != intrinsicDigest(m.record) {
			err = ErrInvalidStore
			break
		}
		members = append(members, m)
		allocations = append(allocations, m.allocation)
	}
	err = errors.Join(err, rows.Err(), rows.Close())
	if err != nil {
		return nil, err
	}
	if len(members) != 6 {
		return nil, ErrInvalidStore
	}
	plan, err := intrinsicmaterialization.Build(catalog, intrinsicmaterialization.Scope{URI: scope, AuthorityID: authority, ContextID: id}, allocations)
	if err != nil {
		return nil, fmt.Errorf("%w: stored intrinsic mapping: %v", ErrInvalidStore, err)
	}
	encoded, err := plan.EncodeContext()
	if err != nil {
		return nil, err
	}
	if intrinsicDigest(encoded) != contextDigest {
		return nil, fmt.Errorf("%w: intrinsic context digest", ErrInvalidStore)
	}
	for _, m := range members {
		expected, ok := plan.LookupLocal(m.allocation.Local)
		if !ok || expected.Global != m.global || !bytes.Equal(expected.Record, m.record) {
			return nil, fmt.Errorf("%w: intrinsic body/source mismatch", ErrInvalidStore)
		}
		var header struct {
			Type intrinsicmaterialization.Reference `json:"type"`
		}
		if err = json.Unmarshal(expected.Record, &header); err != nil {
			return nil, err
		}
		var kind, backing, key, publicID, current, state, currentType string
		err = tx.QueryRowContext(ctx, `SELECT resource_kind,backing,backing_key,public_id,revision,allocation_state,type_url FROM graph_preview_catalog WHERE path=? AND OCTET_LENGTH(public_id)=? AND OCTET_LENGTH(type_url)=?`, m.allocation.PrivatePath, len(m.allocation.PublicID), len(header.Type.URI)).Scan(&kind, &backing, &key, &publicID, &current, &state, &currentType)
		if err != nil {
			return nil, fmt.Errorf("%w: intrinsic allocation: %v", ErrInvalidStore, err)
		}
		wantBacking := "intrinsic"
		if m.allocation.Kind == graph.KindLink {
			wantBacking = "intrinsic-link"
		}
		if currentType != header.Type.URI {
			return nil, fmt.Errorf("%w: intrinsic catalog affiliation differs", ErrInvalidStore)
		}
		if kind != string(m.allocation.Kind) || backing != wantBacking || key != m.key || publicID != m.allocation.PublicID || state != "live" || current != m.allocation.Local.Revision {
			return nil, fmt.Errorf("%w: intrinsic allocation correspondence", ErrInvalidStore)
		}
		if err = budget.chargeBody(2*len(m.record) + 4096); err != nil {
			return nil, err
		}
		var snapshot []byte
		err = tx.QueryRowContext(ctx, `SELECT snapshot FROM graph_preview_versions WHERE path=? AND version=? AND OCTET_LENGTH(snapshot)=?`, m.allocation.PrivatePath, m.allocation.Local.Revision, len(m.record)).Scan(&snapshot)
		if err != nil || !bytes.Equal(snapshot, m.record) {
			return nil, fmt.Errorf("%w: intrinsic retained snapshot: %v", ErrInvalidStore, err)
		}
	}
	return plan, nil
}

// A complete source/semantic/encoding tuple selects one of two retained fixed
// meanings. Unknown or individually mixed fields never fall back to current.
func intrinsicContextSource(repo, commit, path, sha string, size uint64, semantic, interpretation string) (intrinsictypecatalog.Source, error) {
	for _, source := range []intrinsictypecatalog.Source{intrinsictypecatalog.SelectedSource(), intrinsictypecatalog.CurrentSource()} {
		if repo == source.Repository && commit == source.Commit && path == source.Path && sha == source.SHA256 && size == uint64(source.Bytes) && semantic == source.SemanticSource() && interpretation == source.Interpretation() {
			return source, nil
		}
	}
	return intrinsictypecatalog.Source{}, fmt.Errorf("%w: unknown or mixed intrinsic source/context correspondence", ErrInvalidStore)
}
