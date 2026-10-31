package graphstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/steveyegge/beads/resourceget"
)

// nativePredecessorTypeInTx uses only the latest retained predecessor. A
// terminal names its immediately preceding last-live edition; the independent
// selected historical content source and current Type head supply no pin.
func (s *Store) nativePredecessorTypeInTx(ctx context.Context, tx *sql.Tx, path, nominal, version, actor string, ordinal int64, raw []byte, budget *intrinsicMetadataBudget) (resourceget.TypeReference, error) {
	var marker struct {
		State string `json:"state"`
	}
	if json.Unmarshal(raw, &marker) != nil {
		return resourceget.TypeReference{}, ErrInvalidStore
	}
	if marker.State != "deleted" {
		_, pin, err := decodeNativeSnapshotEdition(raw)
		return pin, err
	}
	var terminal ResourceTerminalState
	if validatePath(path) == nil {
		checked, err := s.decodeMemoryTerminal(ctx, tx, path, version, raw, actor)
		if err != nil {
			return resourceget.TypeReference{}, err
		}
		terminal = checked
	} else if validateLinkPath(path) == nil {
		checked, err := s.decodeLinkTerminal(ctx, tx, path, nominal, version, raw, actor)
		if err != nil {
			return resourceget.TypeReference{}, err
		}
		terminal = ResourceTerminalState{ID: checked.ID, Revision: checked.Revision, Version: checked.Version, State: checked.State, PreviousVersion: checked.PreviousVersion, Attribution: checked.Attribution}
	} else {
		return resourceget.TypeReference{}, ErrInvalidStore
	}
	if ordinal <= 1 {
		return resourceget.TypeReference{}, fmt.Errorf("%w: terminal has no last-live predecessor", ErrInvalidStore)
	}
	var size uint64
	if err := tx.QueryRowContext(ctx, `SELECT OCTET_LENGTH(snapshot)+OCTET_LENGTH(actor) FROM graph_preview_versions WHERE path=? AND version=? AND ordinal=?`, path, terminal.PreviousVersion, ordinal-1).Scan(&size); err != nil {
		return resourceget.TypeReference{}, fmt.Errorf("%w: last-live predecessor: %w", ErrInvalidStore, err)
	}
	if size == 0 || size > PreviewCurrentReadByteLimit {
		return resourceget.TypeReference{}, ErrLimitExceeded
	}
	if err := budget.chargeBody(4*int(size) + 8192); err != nil {
		return resourceget.TypeReference{}, err
	}
	var retained []byte
	if err := tx.QueryRowContext(ctx, `SELECT snapshot FROM graph_preview_versions WHERE path=? AND version=? AND ordinal=? AND OCTET_LENGTH(snapshot)+OCTET_LENGTH(actor)=?`, path, terminal.PreviousVersion, ordinal-1, size).Scan(&retained); err != nil {
		return resourceget.TypeReference{}, err
	}
	domain, pin, err := decodeNativeSnapshotEdition(retained)
	if err != nil {
		return resourceget.TypeReference{}, err
	}
	var live struct {
		ID          string      `json:"id"`
		Revision    string      `json:"revision"`
		Version     string      `json:"version"`
		State       string      `json:"state"`
		Attribution Attribution `json:"attribution"`
	}
	if json.Unmarshal(domain, &live) != nil || live.ID != terminal.ID || live.Revision != terminal.PreviousVersion || live.Version != terminal.PreviousVersion || live.State != "" || live.Attribution.Inception != terminal.Attribution.Inception || pin.URI != nominal {
		return resourceget.TypeReference{}, fmt.Errorf("%w: last-live Type evidence differs from terminal baseline", ErrInvalidStore)
	}
	return pin, nil
}
