package graphstore

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	graph "github.com/steveyegge/beads/graphops"
)

// These tables are deliberately preview-specific. They share the standard
// Beads database; Issue and Dependency payloads are not copied into them.
// The catalog alone owns allocation and current revision. A later specialized
// Issue provider must participate in this same transaction/identity model.
var previewDDL = []string{
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
        allocation_state VARBINARY(16) NOT NULL, backing VARBINARY(32) NOT NULL)`,
	`CREATE TABLE graph_preview_payloads (
        path VARBINARY(1024) PRIMARY KEY, properties LONGBLOB NOT NULL)`,
	`CREATE TABLE graph_preview_versions (
        path VARBINARY(1024) NOT NULL, version VARBINARY(32) NOT NULL,
        snapshot LONGBLOB NOT NULL, actor LONGBLOB NOT NULL,
        PRIMARY KEY (path, version))`,
}

// MemoryTypeURL is a Scope-local experimental descriptor, not the eventual
// production Memory Type URI. Its sole payload is title/body; owned is empty.
func MemoryTypeURL(scope string) string { return scope + "types/preview-memory-v1" }

func memoryDescriptor(scope string) (graph.TypeDescriptor, error) {
	return graph.NewTypeDescriptor(graph.TypeDescriptorSpec{
		ID: MemoryTypeURL(scope), Name: "Experimental Memory v1", Describes: graph.KindBead,
		Description: "Disposable C0 preview. The installed writer accepts exactly UTF-8 title and body strings, with no owned Links, and always records a retained snapshot. This descriptor does not claim the complete Memory or History contract.",
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
	for _, ddl := range previewDDL {
		if _, err = conn.ExecContext(ctx, ddl); err != nil {
			return err
		}
	}
	descriptor, err := memoryDescriptor(o.Binding.ScopeURL)
	if err != nil {
		return err
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
	if _, err = tx.ExecContext(ctx, `INSERT INTO graph_preview_types (name, descriptor, fingerprint) VALUES ('memory', ?, ?)`, descriptor.CanonicalJSON(), descriptor.Fingerprint()); err != nil {
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
        schema_version, writer_token FROM graph_preview_scope WHERE singleton = 1`).Scan(
		&b.WorkspaceID, &b.ScopeURL, &b.AuthorityID, &b.SchemaVersion, &token)
	if err != nil {
		return fmt.Errorf("%w: read binding: %v", ErrInvalidStore, err)
	}
	if b != o.Binding || !authorityID.MatchString(token) {
		return fmt.Errorf("%w: persisted binding differs from workspace", ErrInvalidStore)
	}
	var descriptor []byte
	var fingerprint string
	if err := tx.QueryRowContext(ctx, `SELECT descriptor, fingerprint FROM graph_preview_types WHERE name = 'memory'`).Scan(&descriptor, &fingerprint); err != nil {
		return fmt.Errorf("%w: read Memory descriptor: %v", ErrInvalidStore, err)
	}
	expected, err := memoryDescriptor(b.ScopeURL)
	if err != nil {
		return err
	}
	installed, err := graph.ParseTypeDescriptor(descriptor)
	if err != nil {
		return fmt.Errorf("%w: invalid Memory descriptor: %v", ErrInvalidStore, err)
	}
	if !bytes.Equal(descriptor, expected.CanonicalJSON()) || fingerprint != expected.Fingerprint() || fingerprint != installed.Fingerprint() {
		return fmt.Errorf("%w: Memory descriptor differs", ErrInvalidStore)
	}
	// Reject incomplete bootstrap on open, even before a caller selects a row.
	for _, query := range []string{
		`SELECT path, resource_kind, type_url, revision, allocation_state, backing FROM graph_preview_catalog LIMIT 0`,
		`SELECT path, properties FROM graph_preview_payloads LIMIT 0`,
		`SELECT path, version, snapshot, actor FROM graph_preview_versions LIMIT 0`,
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
