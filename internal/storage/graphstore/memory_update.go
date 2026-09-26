package graphstore

import (
	"context"
	"database/sql"
	"fmt"
	"unicode/utf8"

	"github.com/steveyegge/beads/internal/storage"
)

// MemoryUpdateRequest replaces the complete experimental title/body payload.
// Empty strings are values; the caller must distinguish omitted JSON members.
type MemoryUpdateRequest struct {
	Path             string
	Properties       Properties
	Actor            string
	ExpectedRevision string
	Unconditional    bool
}

type MemoryMutationResult struct {
	Memory  Record `json:"memory"`
	Changed bool   `json:"changed"`
}

// UpdateMemory retains the complete accepted Memory, including unchanged owned
// Links, in the same transaction as its properties and revision. This preview
// operation establishes no public History ordering or change-context contract.
func (s *Store) UpdateMemory(ctx context.Context, request MemoryUpdateRequest) (MemoryMutationResult, error) {
	if err := validatePath(request.Path); err != nil {
		return MemoryMutationResult{}, fmt.Errorf("%w: %v", storage.ErrValidation, err)
	}
	if !utf8.ValidString(request.Properties.Title) || !utf8.ValidString(request.Properties.Body) || !utf8.ValidString(request.Actor) {
		return MemoryMutationResult{}, fmt.Errorf("%w: Memory title, body and actor must be UTF-8", storage.ErrValidation)
	}
	var result MemoryMutationResult
	err := s.withTx(ctx, true, func(tx *sql.Tx) error {
		if err := checkBinding(ctx, tx, s.options); err != nil {
			return err
		}
		current, revision, _, err := s.beadEndpointInTx(ctx, tx, request.Path)
		if err != nil {
			return err
		}
		memory, ok := current.(Record)
		if !ok {
			return fmt.Errorf("%w: property replacement supports the experimental Memory Type only", ErrCapabilityUnavailable)
		}
		if err := checkRevisionGuard(request.ExpectedRevision, request.Unconditional, revision, true, "Memory"); err != nil {
			return err
		}
		if memory.Properties == request.Properties {
			result = MemoryMutationResult{Memory: memory}
			return nil
		}
		properties, err := canonicalJSON(request.Properties)
		if err != nil {
			return err
		}
		if err := s.touchCoordination(ctx, tx); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE graph_preview_payloads SET properties=? WHERE path=?`, properties, request.Path); err != nil {
			return err
		}
		if err := s.afterStage("memory-payload"); err != nil {
			return err
		}
		memory.Properties = request.Properties
		accepted, err := s.recordOwnedMemoryInTx(ctx, tx, request.Path, request.Actor, memory)
		if err != nil {
			return err
		}
		result = MemoryMutationResult{Memory: accepted.(Record), Changed: true}
		return nil
	})
	if err != nil {
		return MemoryMutationResult{}, err
	}
	return result, nil
}
