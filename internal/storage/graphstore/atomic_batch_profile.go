package graphstore

import (
	"context"
	"database/sql"
	"fmt"
)

// AtomicBatchSchemaVersion selects a fresh, opt-in local batch candidate. It
// does not migrate ordinary layout 11 or advertise a BDP capability/Core profile.
const AtomicBatchSchemaVersion = 15

var atomicBatchDDL = []string{
	`CREATE TABLE graph_preview_batch_decisions (
 decision_key VARBINARY(256) PRIMARY KEY, context LONGBLOB NOT NULL,
 request LONGBLOB NOT NULL, actor LONGBLOB NOT NULL, result LONGBLOB NOT NULL)`,
	`CREATE TABLE graph_preview_batch_allocations (
 public_id VARBINARY(1024) PRIMARY KEY, resource_kind VARBINARY(16) NOT NULL)`,
}

// InitAtomicBatch provisions only a fresh explicitly selected candidate store.
func InitAtomicBatch(ctx context.Context, o Options) error {
	return initAtVersion(ctx, o, AtomicBatchSchemaVersion, atomicBatchDDL)
}

// OpenAtomicBatch never provisions or upgrades either old or candidate stores.
func OpenAtomicBatch(ctx context.Context, o Options) (*Store, error) {
	return openAtVersion(ctx, o, AtomicBatchSchemaVersion)
}

func checkAtomicBatchSchemaInTx(ctx context.Context, tx *sql.Tx) error {
	for _, query := range []string{
		`SELECT decision_key, context, request, actor, result FROM graph_preview_batch_decisions LIMIT 0`,
		`SELECT public_id, resource_kind FROM graph_preview_batch_allocations LIMIT 0`,
	} {
		rows, err := tx.QueryContext(ctx, query)
		if err != nil {
			return fmt.Errorf("%w: incomplete local batch schema: %v", ErrInvalidStore, err)
		}
		if err := rows.Close(); err != nil {
			return err
		}
	}
	return nil
}
