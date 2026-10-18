package graphstore

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage/issueops"
	"github.com/steveyegge/beads/internal/types"
)

// These tables are deliberately preview-specific. They share the standard
// Beads database; Issue and Dependency payloads are not copied into them.
// The catalog alone owns allocation and current revision. A later specialized
// Issue provider must participate in this same transaction/identity model.
var previewDDL = []string{
	`CREATE TABLE graph_preview_links (
        path VARBINARY(1024) PRIMARY KEY, source_path VARBINARY(1024) NOT NULL,
        target_path VARBINARY(1024) NOT NULL, properties LONGBLOB NOT NULL,
        metadata LONGBLOB NULL,
        attribution LONGBLOB NOT NULL, INDEX source_links (source_path))`,
	`CREATE TABLE graph_preview_scope (
        singleton INT PRIMARY KEY, workspace LONGBLOB NOT NULL,
        scope_url LONGBLOB NOT NULL, authority_id VARBINARY(32) NOT NULL,
        schema_version INT NOT NULL, writer_token VARBINARY(32) NOT NULL)`,
	`CREATE TABLE graph_preview_types (
        name VARBINARY(64) PRIMARY KEY, descriptor LONGBLOB NOT NULL,
        fingerprint VARBINARY(64) NOT NULL)`,
	`CREATE TABLE graph_preview_catalog (
        path VARBINARY(1024) PRIMARY KEY, resource_kind VARBINARY(16) NOT NULL,
        type_url LONGBLOB NOT NULL, revision VARBINARY(32) NOT NULL,
        allocation_state VARBINARY(16) NOT NULL, backing VARBINARY(32) NOT NULL,
        backing_key VARBINARY(255) NULL, UNIQUE KEY one_backing (backing, backing_key))`,
	`CREATE TABLE graph_preview_payloads (
        path VARBINARY(1024) PRIMARY KEY, properties LONGBLOB NOT NULL,
        metadata LONGBLOB NULL)`,
	`CREATE TABLE graph_preview_versions (
        path VARBINARY(1024) NOT NULL, version VARBINARY(32) NOT NULL,
        snapshot LONGBLOB NOT NULL, actor LONGBLOB NOT NULL,
        ordinal BIGINT NOT NULL, change_at DATETIME(6) NOT NULL,
        PRIMARY KEY (path, version), UNIQUE KEY one_path_ordinal (path, ordinal))`,
	`CREATE TABLE graph_preview_issue_versions (
        path VARBINARY(1024) NOT NULL, version VARBINARY(32) NOT NULL,
        issue_id VARBINARY(255) NOT NULL, issue_revision BIGINT NOT NULL, owned LONGBLOB NOT NULL,
        PRIMARY KEY (path, version), UNIQUE KEY one_issue_version (issue_id, issue_revision))`,
}

// MemoryTypeURL is a Scope-local preview descriptor. Its payload is
// title/body and it owns outgoing Links.
func MemoryTypeURL(scope string) string { return scope + "types/preview-memory" }

func memoryDescriptor(scope string) (graph.TypeDescriptor, error) {
	owns, err := graph.NewWildcardOwnedLinkDecl(PreviewOwnedLinkLimit)
	if err != nil {
		return graph.TypeDescriptor{}, err
	}
	return graph.NewTypeDescriptor(graph.TypeDescriptorSpec{
		ID: MemoryTypeURL(scope), Name: "Preview Memory", Describes: graph.KindBead,
		Description:  "Disposable C0 preview. The installed writer accepts exactly UTF-8 title and body strings, owns up to 1000 outgoing Links, and always records a complete retained snapshot. This descriptor does not claim the complete Memory or History contract.",
		OwnsOutgoing: []graph.OwnedLinkDecl{owns},
	})
}

// IssueTypeURL deliberately avoids settling the nominal Task/Bug discussion.
func IssueTypeURL(scope string) string { return scope + "types/preview-issue" }

// DependencyTypeURL is the one native Dependency Link Type. The kind is stored
// in the Link's dep_type property, not its nominal Type URL.
func DependencyTypeURL(scope string) string { return scope + "types/preview-dependency" }

func dependencyDescriptor(scope string) (graph.TypeDescriptor, error) {
	endpoint, err := graph.NewEndpointConstraint([]string{IssueTypeURL(scope)}, graph.ExternalNone)
	if err != nil {
		return graph.TypeDescriptor{}, err
	}
	return graph.NewTypeDescriptor(graph.TypeDescriptorSpec{
		ID: DependencyTypeURL(scope), Name: "Preview Dependency", Describes: graph.KindLink,
		Description: "Issue-owned native Dependency with live unpinned local Issue endpoints. The required dep_type string property mirrors the native Dependency kind; no other properties or metadata edits are supported. One ordered Issue pair has at most one Dependency.",
		Source:      &endpoint, Target: &endpoint,
	})
}

// The one installed Issue Type owns the native Dependency Links exposed by
// this fresh workspace. Issue classification remains an unconstrained property.
func issueDescriptor(scope string) (graph.TypeDescriptor, error) {
	owned := make([]graph.OwnedLinkDecl, 0, len(relationshipDependencyTypes))
	for _, relation := range relationshipDependencyTypes {
		link, err := graph.NewOwnedLinkDecl(scope+relation.path, "", PreviewOwnedLinkLimit)
		if err != nil {
			return graph.TypeDescriptor{}, err
		}
		owned = append(owned, link)
	}
	return graph.NewTypeDescriptor(graph.TypeDescriptorSpec{
		ID: IssueTypeURL(scope), Name: "Preview Issue", Describes: graph.KindBead,
		Description:  "Disposable specialized Issue adapter with source-owned Dependency Links. Issue classification remains a property. Native Issue writers own lifecycle and retained History; these Link Types do not define a close reason or authorize an edge-plus-close operation.",
		OwnsOutgoing: owned,
	})
}

type relationshipDependencyType struct {
	native      types.DependencyType
	installName string
	path        string
	name        string
}

// Kept temporarily for the old installer seam while the fresh generation's
// Type catalog is consolidated. Every valid kind uses one nominal Type.
var relationshipDependencyTypes = [...]relationshipDependencyType{
	{types.DepBlocks, "dependency", "types/preview-dependency", "Preview Dependency"},
}

func relationshipDependencyTypeURL(scope string, native types.DependencyType) (string, bool) {
	if native.IsValid() {
		return DependencyTypeURL(scope), true
	}
	return "", false
}

// RelationshipDependencyTypeURL names a native Dependency Link Type. This
// is an identity lookup, not permission to create the relationship.
func RelationshipDependencyTypeURL(scope string, native types.DependencyType) (string, bool) {
	return relationshipDependencyTypeURL(scope, native)
}

// RelationshipBlockingTypeURL is the installed blocking Link identity.
func RelationshipBlockingTypeURL(scope string) string {
	id, _ := relationshipDependencyTypeURL(scope, types.DepBlocks)
	return id
}

func nativeRelationshipDependencyType(scope, id string) (types.DependencyType, bool) {
	if id == DependencyTypeURL(scope) {
		// The Type identifies a Dependency, while dependencies.type supplies
		// its kind. Callers must not infer a kind from this return value.
		return types.DepBlocks, true
	}
	return "", false
}

func relationshipDependencyInstallName(scope, id string) (string, bool) {
	for _, relation := range relationshipDependencyTypes {
		if id == scope+relation.path {
			return relation.installName, true
		}
	}
	return "", false
}

func relationshipDependencyDescriptor(scope string, native types.DependencyType) (graph.TypeDescriptor, error) {
	if native.IsValid() {
		return dependencyDescriptor(scope)
	}
	return graph.TypeDescriptor{}, fmt.Errorf("unsupported native relationship Dependency %q", native)
}

// RelatedTypeURL names the private experimental informational Link descriptor.
func RelatedTypeURL(scope string) string { return scope + "types/preview-related" }

func relatedDescriptor(scope string) (graph.TypeDescriptor, error) {
	// Endpoint conformsTo is a conjunction, not an alternative-Type list.
	// The preview admits only Issue and Memory Beads; an empty requirement set
	// truthfully accepts either without claiming it conforms to both Types.
	endpoint, err := graph.NewEndpointConstraint(nil, graph.ExternalNone)
	if err != nil {
		return graph.TypeDescriptor{}, err
	}
	return graph.NewTypeDescriptor(graph.TypeDescriptorSpec{
		ID: RelatedTypeURL(scope), Name: "Preview Related", Describes: graph.KindLink,
		Description: "Disposable informational Link with live unpinned local Issue or Memory endpoints and independent multiedge identity. Closed properties record admits only optional UTF-8 string note. Memory sources own outgoing Links; Issue sources do not own this Type. No scheduling effect or production Type contract.",
		Source:      &endpoint, Target: &endpoint,
	})
}

func canonicalJSON(value any) ([]byte, error) {
	b, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return graph.CanonicalizeJSON(b)
}

func freshToken() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

func installPreview(ctx context.Context, conn *sql.Conn, o Options) (err error) {
	prefix := o.IssuePrefix
	if prefix == "" {
		prefix = "bd"
	}
	if !utf8.ValidString(prefix) || strings.TrimSpace(prefix) != prefix {
		return fmt.Errorf("%w: invalid Issue prefix", ErrInvalidStore)
	}
	for _, ddl := range previewDDL {
		if _, err = conn.ExecContext(ctx, ddl); err != nil {
			return err
		}
	}
	token, err := freshToken()
	if err != nil {
		return err
	}
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	finished := false
	defer func() {
		if !finished {
			err = errors.Join(err, tx.Rollback())
		}
	}()
	b := o.Binding
	if _, err = tx.ExecContext(ctx, `INSERT INTO graph_preview_scope
        (singleton, workspace, scope_url, authority_id, schema_version, writer_token)
        VALUES (1, ?, ?, ?, ?, ?)`, b.WorkspaceID, b.ScopeURL, b.AuthorityID, b.SchemaVersion, token); err != nil {
		return err
	}
	for _, definition := range previewTypeDefinitions() {
		descriptor, buildErr := definition.build(b.ScopeURL)
		if buildErr != nil {
			return buildErr
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO graph_preview_types (name, descriptor, fingerprint) VALUES (?, ?, ?)`, definition.name, descriptor.CanonicalJSON(), descriptor.Fingerprint()); err != nil {
			return err
		}
	}
	if _, err = tx.ExecContext(ctx, "INSERT INTO config (`key`,value) VALUES ('issue_prefix', ?)", prefix); err != nil {
		return err
	}
	if _, err = issueops.ReadConfigPrefix(ctx, tx); err != nil {
		return err
	}
	if err = checkBinding(ctx, tx, o); err != nil {
		return err
	}
	err = tx.Commit()
	finished = true
	if err != nil {
		return errors.Join(ErrOutcomeUnknown, err)
	}
	return nil
}

func checkBinding(ctx context.Context, tx *sql.Tx, o Options) error {
	var b Binding
	var token string
	err := tx.QueryRowContext(ctx, `SELECT workspace, scope_url, authority_id,
        schema_version, writer_token FROM graph_preview_scope WHERE singleton = 1
 AND OCTET_LENGTH(workspace)=? AND OCTET_LENGTH(scope_url)=?`, len(o.Binding.WorkspaceID), len(o.Binding.ScopeURL)).Scan(
		&b.WorkspaceID, &b.ScopeURL, &b.AuthorityID, &b.SchemaVersion, &token)
	if err != nil {
		return fmt.Errorf("%w: read binding: %v", ErrInvalidStore, err)
	}
	if b != o.Binding || !authorityID.MatchString(token) {
		return fmt.Errorf("%w: persisted binding differs from workspace", ErrInvalidStore)
	}
	installed, err := installedPreviewTypes(ctx, tx, b.ScopeURL)
	if err != nil {
		return err
	}
	count := len(installed)
	if count != len(previewTypeDefinitions()) {
		return fmt.Errorf("%w: workspace marker and installed Type generation differ", ErrInvalidStore)
	}
	if _, err := issueops.ReadConfigPrefix(ctx, tx); err != nil {
		return fmt.Errorf("%w: Issue configuration: %v", ErrInvalidStore, err)
	}
	// Reject incomplete bootstrap on open, even before a caller selects a row.
	for _, query := range []string{
		`SELECT path, resource_kind, type_url, revision, allocation_state, backing, backing_key FROM graph_preview_catalog LIMIT 0`,
		`SELECT path, properties, metadata FROM graph_preview_payloads LIMIT 0`,
		`SELECT path, source_path, target_path, properties, metadata, attribution FROM graph_preview_links LIMIT 0`,
		`SELECT path, version, snapshot, actor, ordinal, change_at FROM graph_preview_versions LIMIT 0`,
		`SELECT path, version, issue_id, issue_revision, owned FROM graph_preview_issue_versions LIMIT 0`,
		`SELECT issue_id, revision, durable_state, attribution_status FROM issue_versions LIMIT 0`,
	} {
		rows, err := tx.QueryContext(ctx, query)
		if err != nil {
			return fmt.Errorf("%w: incomplete schema: %v", ErrInvalidStore, err)
		}
		if err := rows.Close(); err != nil {
			return err
		}
	}
	return nil
}
