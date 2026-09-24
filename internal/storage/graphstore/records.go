package graphstore

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"unicode/utf8"

	graph "github.com/steveyegge/beads/graphops"
)

func validatePath(path string) error {
	if err := graph.ValidateBeadPath(path); err != nil {
		return err
	}
	if len(path) > 1024 {
		return errors.New("preview Bead path exceeds 1024 bytes")
	}
	return nil
}

// Create atomically reserves a never-reused canonical path, writes its current
// payload and records the complete accepted snapshot. There is no keyed upsert
// or optional History switch. Commit uncertainty returns no success record.
func (s *Store) Create(ctx context.Context, req CreateRequest) (Record, error) {
	if err := validatePath(req.Path); err != nil {
		return Record{}, err
	}
	if !utf8.ValidString(req.Title) || !utf8.ValidString(req.Body) || !utf8.ValidString(req.Actor) {
		return Record{}, errors.New("Memory title, body and actor must be valid UTF-8")
	}
	revision, err := freshToken()
	if err != nil {
		return Record{}, err
	}
	r := Record{ID: graph.CanonicalURL(s.options.Binding.ScopeURL, req.Path),
		Type: MemoryTypeURL(s.options.Binding.ScopeURL), Revision: revision, Version: revision,
		Properties: Properties{Title: req.Title, Body: req.Body}, Owned: []json.RawMessage{}}
	r.Attribution = Attribution{Actor: req.Actor, Status: "unknown", RecordedAt: time.Now().UTC().Format(time.RFC3339Nano)}
	if req.Actor != "" {
		r.Attribution.Status = "claimed"
	}
	properties, err := canonicalJSON(r.Properties)
	if err != nil {
		return Record{}, err
	}
	snapshot, err := canonicalJSON(r)
	if err != nil {
		return Record{}, err
	}
	err = s.withTx(ctx, true, func(tx *sql.Tx) error {
		if err := checkBinding(ctx, tx, s.options); err != nil {
			return err
		}
		return s.createInTx(ctx, tx, req, r, properties, snapshot)
	})
	if err != nil {
		return Record{}, err
	}
	return r, nil
}

func (s *Store) createInTx(ctx context.Context, tx *sql.Tx, req CreateRequest, r Record, properties, snapshot []byte) error {
	// Each controlled writer changes this same cell to a fresh random value.
	// Concurrent snapshots cannot silently merge disjoint successful mutations.
	token, err := freshToken()
	if err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, `UPDATE graph_preview_scope SET writer_token = ? WHERE singleton = 1`, token)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return fmt.Errorf("%w: missing writer coordination row", ErrInvalidStore)
	}
	if err := s.afterStage("coordination"); err != nil {
		return err
	}
	var exists int
	err = tx.QueryRowContext(ctx, `SELECT 1 FROM graph_preview_catalog WHERE path = ?`, req.Path).Scan(&exists)
	if err == nil {
		return ErrAlreadyExists
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO graph_preview_catalog
        (path, resource_kind, type_url, revision, allocation_state, backing)
        VALUES (?, 'bead', ?, ?, 'live', 'generic')`, req.Path, r.Type, r.Revision); err != nil {
		return err
	}
	if err := s.afterStage("allocation"); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO graph_preview_payloads (path, properties) VALUES (?, ?)`, req.Path, properties); err != nil {
		return err
	}
	if err := s.afterStage("payload"); err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO graph_preview_versions (path, version, snapshot, actor) VALUES (?, ?, ?, ?)`, req.Path, r.Version, snapshot, req.Actor)
	if err != nil {
		return err
	}
	return s.afterStage("retained")
}

func (s *Store) afterStage(stage string) error {
	if s.afterWrite != nil {
		return s.afterWrite(stage)
	}
	return nil
}

// Show reads one canonical Memory and verifies its required retained snapshot
// in the same read transaction. Corrupt/incomplete state is never an absence.
func (s *Store) Show(ctx context.Context, path string) (Record, error) {
	if err := validatePath(path); err != nil {
		return Record{}, err
	}
	var record Record
	err := s.withTx(ctx, false, func(tx *sql.Tx) error {
		if err := checkBinding(ctx, tx, s.options); err != nil {
			return err
		}
		var kind, typ, revision, state, backing string
		err := tx.QueryRowContext(ctx, `SELECT resource_kind, type_url, revision, allocation_state, backing
            FROM graph_preview_catalog WHERE path = ?`, path).Scan(&kind, &typ, &revision, &state, &backing)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return err
		}
		if kind != "bead" || typ != MemoryTypeURL(s.options.Binding.ScopeURL) || !authorityID.MatchString(revision) || state != "live" || backing != "generic" {
			return fmt.Errorf("%w: unsupported or corrupt allocation", ErrInvalidStore)
		}
		var payload, snapshot []byte
		var actor string
		if err := tx.QueryRowContext(ctx, `SELECT properties FROM graph_preview_payloads WHERE path = ?`, path).Scan(&payload); err != nil {
			return fmt.Errorf("%w: missing current payload: %v", ErrInvalidStore, err)
		}
		if err := tx.QueryRowContext(ctx, `SELECT snapshot, actor FROM graph_preview_versions WHERE path = ? AND version = ?`, path, revision).Scan(&snapshot, &actor); err != nil {
			return fmt.Errorf("%w: missing retained state: %v", ErrInvalidStore, err)
		}
		record = Record{ID: graph.CanonicalURL(s.options.Binding.ScopeURL, path), Type: typ,
			Revision: revision, Version: revision, Owned: []json.RawMessage{}}
		var retained Record
		if err := json.Unmarshal(snapshot, &retained); err != nil {
			return fmt.Errorf("%w: malformed retained state", ErrInvalidStore)
		}
		record.Attribution = retained.Attribution
		at, err := time.Parse(time.RFC3339Nano, record.Attribution.RecordedAt)
		if err != nil || at.UTC().Format(time.RFC3339Nano) != record.Attribution.RecordedAt || actor != record.Attribution.Actor ||
			(record.Attribution.Actor == "" && record.Attribution.Status != "unknown") ||
			(record.Attribution.Actor != "" && record.Attribution.Status != "claimed") {
			return fmt.Errorf("%w: invalid retained attribution", ErrInvalidStore)
		}
		if err := json.Unmarshal(payload, &record.Properties); err != nil {
			return fmt.Errorf("%w: malformed payload", ErrInvalidStore)
		}
		canonicalPayload, err := canonicalJSON(record.Properties)
		if err != nil {
			return err
		}
		canonicalSnapshot, err := canonicalJSON(record)
		if err != nil {
			return err
		}
		if !bytes.Equal(payload, canonicalPayload) || !bytes.Equal(snapshot, canonicalSnapshot) {
			return fmt.Errorf("%w: current and retained state differ", ErrInvalidStore)
		}
		return nil
	})
	if err != nil {
		return Record{}, err
	}
	return record, nil
}
