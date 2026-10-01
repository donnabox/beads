package graphstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"unicode/utf8"

	"github.com/steveyegge/beads/internal/storage"
)

// ErrDeletionPolicyUnresolved refuses incident-Link cases while their deletion
// policy remains under review. Explicit prior unlink is not an atomic cascade.
var ErrDeletionPolicyUnresolved = errors.New("deletion_policy_unresolved: Memory deletion with live incident Links is not supported")

// MemoryDeleteRequest is an internal disposable-preview operation. Preview is
// read-only and may omit a guard; apply requires an observed revision or explicit
// unconditional intent. Neither mode deletes Links or erases retained versions.
type MemoryDeleteRequest struct {
	Path, Actor, ExpectedRevision string
	Unconditional, Preview        bool
}

// MemoryDeleteResult identifies the actual final live state. Deleted does not
// mint a successor revision, a timestamp, or a public History event.
type MemoryDeleteResult struct {
	Memory  Record `json:"memory"`
	Preview bool   `json:"preview"`
	Deleted bool   `json:"deleted"`
}

func (s *Store) DeleteMemory(ctx context.Context, request MemoryDeleteRequest) (MemoryDeleteResult, error) {
	if err := validatePath(request.Path); err != nil {
		return MemoryDeleteResult{}, fmt.Errorf("%w: %v", storage.ErrValidation, err)
	}
	if !utf8.ValidString(request.Actor) {
		return MemoryDeleteResult{}, fmt.Errorf("%w: actor must be UTF-8", storage.ErrValidation)
	}
	if request.ExpectedRevision != "" {
		if err := validateVersionToken(request.ExpectedRevision); err != nil {
			return MemoryDeleteResult{}, err
		}
	}
	var result MemoryDeleteResult
	err := s.withTx(ctx, !request.Preview, func(tx *sql.Tx) error {
		if err := checkBinding(ctx, tx, s.options); err != nil {
			return err
		}
		current, revision, _, err := s.beadEndpointInTx(ctx, tx, request.Path)
		if err != nil {
			return err
		}
		memory, ok := current.(Record)
		if !ok {
			return fmt.Errorf("%w: deletion supports the experimental Memory Type only", ErrCapabilityUnavailable)
		}
		if err := checkRevisionGuard(request.ExpectedRevision, request.Unconditional, revision, !request.Preview, "Memory"); err != nil {
			return err
		}
		if err := s.checkLinkMappingsInTx(ctx, tx); err != nil {
			return err
		}
		// Only informational Links can have a Memory endpoint in this preview.
		// Complete mapping validation above prevents orphan rows from disappearing
		// in this existence check. Count does not load or truncate incident records.
		var incident int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM graph_preview_links WHERE source_path=? OR target_path=?`, request.Path, request.Path).Scan(&incident); err != nil {
			return err
		}
		if incident != 0 {
			return ErrDeletionPolicyUnresolved
		}
		result = MemoryDeleteResult{Memory: memory, Preview: request.Preview}
		if request.Preview {
			return nil
		}
		if err := s.touchCoordination(ctx, tx); err != nil {
			return err
		}
		changed, err := tx.ExecContext(ctx, `UPDATE graph_preview_catalog SET allocation_state='deleted' WHERE path=? AND allocation_state='live' AND revision=?`, request.Path, revision)
		if err != nil {
			return err
		}
		if n, err := changed.RowsAffected(); err != nil {
			return err
		} else if n != 1 {
			return fmt.Errorf("%w: Memory deletion lost its allocation", ErrInvalidStore)
		}
		if err := s.afterStage("memory-delete-allocation"); err != nil {
			return err
		}
		removed, err := tx.ExecContext(ctx, `DELETE FROM graph_preview_payloads WHERE path=?`, request.Path)
		if err != nil {
			return err
		}
		if n, err := removed.RowsAffected(); err != nil {
			return err
		} else if n != 1 {
			return fmt.Errorf("%w: Memory deletion lost its payload", ErrInvalidStore)
		}
		if err := s.afterStage("memory-delete-payload"); err != nil {
			return err
		}
		if _, err := s.deletedMemoryInTx(ctx, tx, request.Path); err != nil {
			return err
		}
		result.Deleted = true
		return nil
	})
	if err != nil {
		return MemoryDeleteResult{}, err
	}
	return result, nil
}

// deletedMemoryInTx validates a reserved absent Memory against its unchanged
// final live snapshot. It never joins live endpoints into historical owned state.
func (s *Store) deletedMemoryInTx(ctx context.Context, tx *sql.Tx, path string) (Record, error) {
	var kind, typ, revision, state, backing string
	var key sql.NullString
	if err := tx.QueryRowContext(ctx, `SELECT resource_kind,type_url,revision,allocation_state,backing,backing_key FROM graph_preview_catalog WHERE path=?`, path).Scan(&kind, &typ, &revision, &state, &backing, &key); err != nil {
		return Record{}, fmt.Errorf("%w: missing deleted Memory allocation: %v", ErrInvalidStore, err)
	}
	if validatePath(path) != nil || kind != "bead" || typ != MemoryTypeURL(s.ScopeURL()) || !authorityID.MatchString(revision) || state != "deleted" || backing != "generic" || key.Valid {
		return Record{}, fmt.Errorf("%w: invalid deleted Memory allocation", ErrInvalidStore)
	}
	var current int
	if err := tx.QueryRowContext(ctx, `SELECT (SELECT COUNT(*) FROM graph_preview_payloads WHERE path=?) + (SELECT COUNT(*) FROM graph_preview_links WHERE path=? OR source_path=? OR target_path=?) + (SELECT COUNT(*) FROM graph_preview_issue_versions WHERE path=?)`, path, path, path, path, path).Scan(&current); err != nil {
		return Record{}, err
	}
	if current != 0 {
		return Record{}, fmt.Errorf("%w: deleted Memory has current payload, incident Links or Issue mapping", ErrInvalidStore)
	}
	raw, actor, err := readVersionBytes(ctx, tx, path, revision, revision)
	if err != nil {
		return Record{}, err
	}
	memory, err := s.decodeMemoryVersion(ctx, tx, path, typ, revision, raw, actor)
	if err != nil {
		return Record{}, err
	}
	if len(memory.Owned) != 0 {
		return Record{}, fmt.Errorf("%w: deleted Memory final state still owns Links", ErrInvalidStore)
	}
	return memory, nil
}
