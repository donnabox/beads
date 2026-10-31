package graphstore

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/intrinsicmaterialization"
	"github.com/steveyegge/beads/internal/intrinsictypecatalog"
)

var intrinsicDDL = []string{
	`CREATE TABLE graph_preview_intrinsic_contexts (
 context_id VARBINARY(32) PRIMARY KEY, scope_url LONGBLOB NOT NULL,
 authority_id VARBINARY(32) NOT NULL, source_repository LONGBLOB NOT NULL,
 source_commit VARBINARY(40) NOT NULL, source_path LONGBLOB NOT NULL,
 source_sha256 VARBINARY(64) NOT NULL, source_bytes LONGBLOB NOT NULL,
 semantic_source_commit VARBINARY(40) NOT NULL, interpretation VARBINARY(64) NOT NULL,
 context_sha256 VARBINARY(64) NOT NULL)`,
	`CREATE TABLE graph_preview_intrinsic_members (
 path VARBINARY(1024) NOT NULL, revision VARBINARY(32) NOT NULL,
 allocation_key VARBINARY(32) NOT NULL, context_id VARBINARY(32) NOT NULL,
 role VARBINARY(32) NOT NULL, public_id LONGBLOB NOT NULL,
 resource_uri LONGBLOB NOT NULL, resource_kind VARBINARY(16) NOT NULL,
 global_uri LONGBLOB NOT NULL, global_revision LONGBLOB NOT NULL,
 record LONGBLOB NOT NULL, record_sha256 VARBINARY(64) NOT NULL,
 PRIMARY KEY (path,revision), UNIQUE KEY one_context_role (context_id,role),
 KEY member_context (context_id))`,
}

func intrinsicDigest(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

// installIntrinsicMembers runs only inside the fresh initializer's transaction.
// Allocations, immutable mapping, final owner captures and retained snapshots
// commit together. No caller-supplied publication or identity mapping is admitted.
func installIntrinsicMembers(ctx context.Context, tx *sql.Tx, o Options, budget *intrinsicMetadataBudget) error {
	if err := (&Store{}).touchCoordination(ctx, tx); err != nil {
		return err
	}
	if err := budget.chargeBody(4*intrinsictypecatalog.CurrentSource().Bytes + 2*intrinsicmaterialization.MaxContextBytes); err != nil {
		return err
	}
	catalog, err := intrinsictypecatalog.LoadCurrent()
	if err != nil {
		return err
	}
	contextID, err := freshToken()
	if err != nil {
		return err
	}
	roles := []intrinsicmaterialization.Role{intrinsicmaterialization.Type, intrinsicmaterialization.BeadType, intrinsicmaterialization.LinkType, intrinsicmaterialization.Refines, intrinsicmaterialization.BeadTypeRefinesType, intrinsicmaterialization.LinkTypeRefinesType}
	allocations := make([]intrinsicmaterialization.Allocation, 0, 6)
	keys := make(map[string]string, 6)
	for i, role := range roles {
		// These are fresh ordinary Resource IDs, not aliases for a hidden
		// type-for-* namespace. Retained contexts keep their recorded IDs.
		id := string(role)
		kind := graph.KindBead
		path := "beads/" + id
		if i >= 4 {
			id = string(role)
			kind = graph.KindLink
			path = "links/" + id
		}
		exists, err := publicResourceIDExistsInTx(ctx, tx, id)
		if err != nil {
			return err
		}
		if exists {
			return publicResourceIDReservedErrorInTx(ctx, tx, id)
		}
		revision, err := freshToken()
		if err != nil {
			return err
		}
		key, err := freshToken()
		if err != nil {
			return err
		}
		keys[path] = key
		uri, err := graph.CanonicalPublicResourceURL(o.Binding.ScopeURL, id)
		if err != nil {
			return err
		}
		allocations = append(allocations, intrinsicmaterialization.Allocation{Role: role, PublicID: id, PrivatePath: path, Kind: kind, Local: intrinsicmaterialization.Reference{URI: uri, Revision: revision}, AuthorityID: o.Binding.AuthorityID})
	}
	plan, err := intrinsicmaterialization.Build(catalog, intrinsicmaterialization.Scope{URI: o.Binding.ScopeURL, AuthorityID: o.Binding.AuthorityID, ContextID: contextID}, allocations)
	if err != nil {
		return err
	}
	encoded, err := plan.EncodeContext()
	if err != nil {
		return err
	}
	source := intrinsictypecatalog.CurrentSource()
	if _, err = tx.ExecContext(ctx, `INSERT INTO graph_preview_intrinsic_contexts (context_id,scope_url,authority_id,source_repository,source_commit,source_path,source_sha256,source_bytes,semantic_source_commit,interpretation,context_sha256) VALUES (?,?,?,?,?,?,?,?,?,?,?)`, contextID, o.Binding.ScopeURL, o.Binding.AuthorityID, source.Repository, source.Commit, source.Path, source.SHA256, catalog.Bytes(), source.SemanticSource(), source.Interpretation(), intrinsicDigest(encoded)); err != nil {
		return err
	}
	for _, member := range plan.Bindings() {
		a := member.Allocation
		var header struct {
			Type intrinsicmaterialization.Reference `json:"type"`
		}
		if err = json.Unmarshal(member.Record, &header); err != nil {
			return err
		}
		backing := "intrinsic"
		if a.Kind == graph.KindLink {
			backing = "intrinsic-link"
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO graph_preview_catalog (path,resource_kind,type_url,revision,allocation_state,backing,backing_key,public_id) VALUES (?,?,?,?,'live',?,?,?)`, a.PrivatePath, string(a.Kind), header.Type.URI, a.Local.Revision, backing, keys[a.PrivatePath], a.PublicID); err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO graph_preview_intrinsic_members (path,revision,allocation_key,context_id,role,public_id,resource_uri,resource_kind,global_uri,global_revision,record,record_sha256) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`, a.PrivatePath, a.Local.Revision, keys[a.PrivatePath], contextID, string(a.Role), a.PublicID, a.Local.URI, string(a.Kind), member.Global.URI, member.Global.Revision, []byte(member.Record), intrinsicDigest(member.Record)); err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO graph_preview_versions (path,version,snapshot,actor,ordinal,change_at) VALUES (?,?,?,'intrinsic-bootstrap',1,UTC_TIMESTAMP(6))`, a.PrivatePath, a.Local.Revision, []byte(member.Record)); err != nil {
			return err
		}
	}
	// Definitions are ordinary allocated Beads, not an alternate descriptor-only
	// inventory. Their source-bound native contracts remain checked against the
	// SAME records and exact editions in subsequent read/write transactions.
	if err = installOrdinaryTypeDefinitionsInTx(ctx, tx, o, plan, budget); err != nil {
		return err
	}
	if _, err = loadIntrinsicContextInTx(ctx, tx, o, contextID, budget); err != nil {
		return fmt.Errorf("%w: final intrinsic census: %v", ErrInvalidStore, err)
	}
	return nil
}
