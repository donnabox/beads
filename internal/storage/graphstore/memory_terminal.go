package graphstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

// MemoryTerminalState is retained deletion evidence, not a live Memory snapshot.
// It deliberately omits the Type, properties, metadata and owned Links. The
// previous live snapshot supplies the separately acquired protected baseline.
type MemoryTerminalState struct {
	ID              string      `json:"id"`
	Revision        string      `json:"revision"`
	Version         string      `json:"version"`
	State           string      `json:"state"`
	PreviousVersion string      `json:"previousVersion"`
	Attribution     Attribution `json:"attribution"`
}

func (s *Store) decodeMemoryTerminal(path, version string, raw []byte, actor string) (MemoryTerminalState, error) {
	var terminal MemoryTerminalState
	if json.Unmarshal(raw, &terminal) != nil || !sameVersionJSON(raw, terminal) ||
		terminal.ID != s.ScopeURL()+path || terminal.Revision != version || terminal.Version != version ||
		terminal.State != "deleted" || !authorityID.MatchString(version) ||
		!authorityID.MatchString(terminal.PreviousVersion) || terminal.PreviousVersion == version ||
		terminal.Attribution.HistoricalUpdate != nil || !validVersionAttribution(terminal.Attribution, actor) {
		return MemoryTerminalState{}, fmt.Errorf("%w: invalid retained Memory deletion marker", ErrInvalidStore)
	}
	return terminal, nil
}

// deletedMemoryStateInTx acquires both the terminal head and its last-current
// baseline. Legacy preview deletions whose head is still live remain readable,
// but return no terminal: that old live token can never stand in for deletion D.
func (s *Store) deletedMemoryStateInTx(ctx context.Context, tx *sql.Tx, path string) (Record, *MemoryTerminalState, error) {
	var kind, typ, revision, state, backing string
	var key sql.NullString
	if err := tx.QueryRowContext(ctx, `SELECT resource_kind,type_url,revision,allocation_state,backing,backing_key FROM graph_preview_catalog WHERE path=?`, path).Scan(&kind, &typ, &revision, &state, &backing, &key); err != nil {
		return Record{}, nil, fmt.Errorf("%w: missing deleted Memory allocation: %v", ErrInvalidStore, err)
	}
	if err := s.validDeletedMemoryAllocationInTx(ctx, tx, path, kind, typ, revision, state, backing, key); err != nil {
		return Record{}, nil, err
	}
	raw, actor, err := readVersionBytes(ctx, tx, path, revision, revision)
	if err != nil {
		return Record{}, nil, err
	}
	var marker struct {
		State string `json:"state"`
	}
	if err := json.Unmarshal(raw, &marker); err != nil {
		return Record{}, nil, fmt.Errorf("%w: malformed deleted Memory head", ErrInvalidStore)
	}
	var terminal *MemoryTerminalState
	baselineRevision := revision
	if marker.State != "" {
		decoded, err := s.decodeMemoryTerminal(path, revision, raw, actor)
		if err != nil {
			return Record{}, nil, err
		}
		terminal = &decoded
		baselineRevision = decoded.PreviousVersion
		raw, actor, err = readVersionBytes(ctx, tx, path, baselineRevision, baselineRevision)
		if err != nil {
			return Record{}, nil, err
		}
	}
	memory, err := s.decodeMemoryVersion(ctx, tx, path, typ, baselineRevision, raw, actor)
	if err != nil {
		return Record{}, nil, err
	}
	if len(memory.Owned) != 0 || (terminal != nil && terminal.Attribution.Inception != memory.Attribution.Inception) {
		return Record{}, nil, fmt.Errorf("%w: deleted Memory baseline differs from protected terminal facts", ErrInvalidStore)
	}
	return memory, terminal, nil
}

// VersionHistoryResult is the native authorized History carrier for current terminal
// evidence. Rows and current state are acquired in one checked snapshot. It is
// separate from ordinary Read, which continues refusing a deleted subject.
type VersionHistoryResult struct {
	Kind         ResourceKind         `json:"kind"`
	Versions     []VersionRow         `json:"versions"`
	CurrentState *MemoryTerminalState `json:"currentState,omitempty"`
}

func (s *Store) History(ctx context.Context, path string) (VersionHistoryResult, error) {
	if err := validateResourcePath(path); err != nil {
		return VersionHistoryResult{}, err
	}
	var result VersionHistoryResult
	err := s.withTx(ctx, false, func(tx *sql.Tx) error {
		if err := checkBinding(ctx, tx, s.options); err != nil {
			return err
		}
		var err error
		result.Kind, result.Versions, err = s.versionsInTx(ctx, tx, path, shippedVersionBounds())
		if err != nil {
			return err
		}
		if result.Kind == KindMemory {
			var state string
			if err := tx.QueryRowContext(ctx, `SELECT allocation_state FROM graph_preview_catalog WHERE path=?`, path).Scan(&state); err != nil {
				return err
			}
			if state == "deleted" {
				_, result.CurrentState, err = s.deletedMemoryStateInTx(ctx, tx, path)
				return err
			}
		}
		return nil
	})
	if err != nil {
		return VersionHistoryResult{}, err
	}
	return result, nil
}
