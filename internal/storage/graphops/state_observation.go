package graphops

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
)

// The eight positions are the replicated tables in storage spec B4 order.
// These are observed working-set facts, not a witness or an authority verdict.
type stateObservation struct {
	database string
	branch   string
	hashes   [8]string
	version  string
}

func (s stateObservation) descriptorHash() string { return s.hashes[2] }

// Database and branch names have a private 1024-byte observation budget. The
// SQL prefix is character-based (at most 4100 UTF-8 bytes); the byte check below
// refuses oversized values rather than accepting a truncated identity.
// Hash functions return fixed-width labels on the pinned engine. No caller
// supplies identifiers or SQL, and all eight expressions share one statement.
const stateObservationQuery = `SELECT
 SUBSTRING(DATABASE(), 1, 1025), SUBSTRING(ACTIVE_BRANCH(), 1, 1025),
 DOLT_HASHOF_TABLE('graph_scope'),
 DOLT_HASHOF_TABLE('graph_scope_history'),
 DOLT_HASHOF_TABLE('graph_type_descriptors'),
 DOLT_HASHOF_TABLE('graph_beads'),
 DOLT_HASHOF_TABLE('graph_links'),
 DOLT_HASHOF_TABLE('graph_ledger_seq'),
 DOLT_HASHOF_TABLE('graph_ledger_events'),
 DOLT_HASHOF_TABLE('graph_allocations')`

var errObservationDeadline = errors.New("graph observation requires a context deadline")

// observeStateInTx uses the caller's transaction and deadline. It neither opens
// a transaction nor validates ownership/default-branch authority. Engine snapshot
// behavior requires separate qualification; a hash is not complete schema proof.
func observeStateInTx(ctx context.Context, tx queryer) (stateObservation, error) {
	if _, ok := ctx.Deadline(); !ok {
		return stateObservation{}, errObservationDeadline
	}
	if err := ctx.Err(); err != nil {
		return stateObservation{}, err
	}
	items, err := readRows(ctx, tx, stateObservationQuery, nil, 1, func(rows *sql.Rows) (stateObservation, error) {
		var result stateObservation
		var database, branch sql.NullString
		var hashes [8]sql.NullString
		if err := rows.Scan(&database, &branch,
			&hashes[0], &hashes[1], &hashes[2], &hashes[3],
			&hashes[4], &hashes[5], &hashes[6], &hashes[7]); err != nil {
			return stateObservation{}, err
		}
		if !database.Valid || !branch.Valid {
			return stateObservation{}, corrupt(errors.New("NULL database or branch identity"))
		}
		result.database, result.branch = database.String, branch.String
		for i, value := range hashes {
			if !value.Valid {
				return stateObservation{}, corrupt(fmt.Errorf("NULL table hash %d", i))
			}
			result.hashes[i] = value.String
		}
		if result.database == "" || result.branch == "" {
			return stateObservation{}, corrupt(errors.New("missing database or branch identity"))
		}
		if len(result.database) > 1024 || len(result.branch) > 1024 {
			return stateObservation{}, errBudget
		}
		version, err := composeStateVersion(result.hashes)
		if err != nil {
			return stateObservation{}, err
		}
		result.version = version
		return result, nil
	})
	if errors.Is(err, errRowOverflow) {
		err = errors.Join(err, corrupt(errors.New("duplicate state observation")))
	}
	if err != nil {
		return stateObservation{}, err
	}
	if len(items) != 1 {
		return stateObservation{}, corrupt(errors.New("missing state observation"))
	}
	if err := ctx.Err(); err != nil {
		return stateObservation{}, err
	}
	return items[0], nil
}

// B4 preimage: exactly 256 ASCII bytes, the eight 32-character lowercase
// base32 labels in fixed order, without a separator or domain tag. Do not decode
// labels to bytes, normalize them, or substitute DOLT_HASHOF_DB.
func composeStateVersion(hashes [8]string) (string, error) {
	var preimage [256]byte
	for i, label := range hashes {
		if len(label) != 32 {
			return "", corrupt(fmt.Errorf("table hash %d has invalid width", i))
		}
		for j := range label {
			c := label[j]
			if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'v') {
				return "", corrupt(fmt.Errorf("table hash %d has invalid base32 label", i))
			}
		}
		copy(preimage[i*32:], label)
	}
	sum := sha256.Sum256(preimage[:])
	return hex.EncodeToString(sum[:]), nil
}
