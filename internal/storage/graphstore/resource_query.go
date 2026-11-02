package graphstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	graph "github.com/steveyegge/beads/graphops"
)

// resourceQueryAllocationLimit limits ALL allocations, including tombstones, before any
// selected Type or native Resource payload is acquired. Ordinary inventory keeps
// PreviewSnapshotLimit; neither bound is a request-controlled page size.
const resourceQueryAllocationLimit = 10000

// ReadResourceQuerySnapshot captures complete Resources and their exact checked
// Type facts in ONE read transaction. It does not interpret query expressions.
// Every failure, including cancellation or read-transaction rollback/connection
// failure, returns the zero snapshot rather than a usable partial inventory.
func (s *Store) ReadResourceQuerySnapshot(ctx context.Context) (ResourceSnapshot, error) {
	if ctx == nil {
		return ResourceSnapshot{}, errors.New("Resource query requires a context")
	}
	if err := ctx.Err(); err != nil {
		return ResourceSnapshot{}, err
	}
	if s.LayoutVersion() != TypesFreshSchemaVersion {
		return ResourceSnapshot{}, ErrCapabilityUnavailable
	}
	var result ResourceSnapshot
	err := s.withDefinitionReadTx(ctx, txResourceQuery, func(tx *sql.Tx, budget *intrinsicMetadataBudget) error {
		var err error
		result, err = s.readResourceQuerySnapshotInTx(ctx, tx, budget)
		return err
	})
	err = errors.Join(err, ctx.Err())
	if err != nil {
		return ResourceSnapshot{}, err
	}
	return result, nil
}

func (s *Store) readResourceQuerySnapshotInTx(ctx context.Context, tx *sql.Tx, budget *intrinsicMetadataBudget) (ResourceSnapshot, error) {
	// withDefinitionReadTx has already checked actual same-Tx binding/schema/
	// policy and admitted this exact read purpose, before intrinsic body reads.
	if err := checkResourceQueryPreflightInTx(ctx, tx); err != nil {
		return ResourceSnapshot{}, err
	}
	result, err := s.currentResourceDocumentsAtLimitInTx(ctx, tx, budget, resourceQueryAllocationLimit)
	if err != nil {
		return ResourceSnapshot{}, err
	}
	if err = s.afterStage("resource-query-inventory"); err != nil {
		return ResourceSnapshot{}, err
	}
	if err = ctx.Err(); err != nil {
		return ResourceSnapshot{}, err
	}
	return result, nil
}

func checkResourceQueryPreflightInTx(ctx context.Context, tx *sql.Tx) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	var count uint64
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM graph_preview_catalog`).Scan(&count); err != nil {
		return err
	}
	if count > resourceQueryAllocationLimit {
		return fmt.Errorf("%w: Resource query allocation count exceeds %d", ErrLimitExceeded, resourceQueryAllocationLimit)
	}
	return checkResourceQueryReadBytesInTx(ctx, tx)
}

// Existing current-read coverage remains intact (including deleted/native
// Issue state, dependencies, labels, comments, and orphan informational rows).
// Only the recorded CURRENT intrinsic context/source and its six member rows
// are added; selected retained native captures are still separately measured
// and debited before exact acquisition by the common production projector.
func resourceQueryReadSizeQueries() []currentReadSizeQuery {
	const selectedContext = `context_id IN (SELECT m.context_id FROM graph_preview_intrinsic_members m JOIN graph_preview_catalog c ON c.path=m.path AND c.revision=m.revision WHERE m.role='type' AND c.backing='intrinsic' AND c.allocation_state='live')`
	return append(currentReadSizeQueries(),
		currentReadSizeQuery{"graph_preview_intrinsic_contexts WHERE " + selectedContext, "context_id,scope_url,authority_id,source_repository,source_commit,source_path,source_sha256,source_bytes,semantic_source_commit,interpretation,context_sha256", 1},
		currentReadSizeQuery{"graph_preview_intrinsic_members WHERE " + selectedContext, "path,revision,allocation_key,context_id,role,public_id,resource_uri,resource_kind,global_uri,global_revision,record,record_sha256", 1},
	)
}

func checkResourceQueryReadBytesInTx(ctx context.Context, tx *sql.Tx) error {
	var used uint64
	for _, part := range resourceQueryReadSizeQueries() {
		if err := ctx.Err(); err != nil {
			return err
		}
		fields := []string{"256"}
		for _, column := range strings.Split(part.columns, ",") {
			fields = append(fields, "COALESCE(OCTET_LENGTH("+strings.TrimSpace(column)+"),0)")
		}
		query := "SELECT CAST(LEAST(COALESCE(SUM((" + strings.Join(fields, "+") + ")*?),0),?) AS UNSIGNED) FROM " + part.from
		var size uint64
		if err := tx.QueryRowContext(ctx, query, part.copies, PreviewCurrentReadByteLimit+1).Scan(&size); err != nil {
			return err
		}
		if size > PreviewCurrentReadByteLimit-used {
			return fmt.Errorf("%w: complete Resource query acquisition exceeds %d bytes", ErrLimitExceeded, PreviewCurrentReadByteLimit)
		}
		used += size
	}
	// Preserve the fixed361 pre-hydration alternative-authority refusal. An
	// active same-ID wisp must never redirect Issue hydration outside the sum.
	var wisps int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM wisps w JOIN graph_preview_catalog c ON c.backing='issue' AND c.backing_key=w.id WHERE c.allocation_state='live'`).Scan(&wisps); err != nil {
		return err
	}
	if wisps != 0 {
		return fmt.Errorf("%w: live graph Issue also names a wisp", ErrInvalidStore)
	}
	return ctx.Err()
}

// Header mapping is shared with ordinary inventory, with its original cap.
// The Query caller has additionally bounded ALL allocations before entering
// the common capture. No body is returned or acquired by this header reader.
func readCommonLiveIdentitiesInTx(ctx context.Context, tx *sql.Tx, limit int) ([]AllocatedResourceIdentity, error) {
	if limit != PreviewSnapshotLimit && limit != resourceQueryAllocationLimit {
		return nil, ErrInvalidStore
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT `+allocatedIdentityColumns+` FROM graph_preview_catalog WHERE allocation_state='live' ORDER BY path LIMIT ?`, limit+1)
	if err != nil {
		return nil, err
	}
	identities := []AllocatedResourceIdentity{}
	seen := map[string]bool{}
	for rows.Next() {
		if err = ctx.Err(); err != nil {
			break
		}
		var identity AllocatedResourceIdentity
		var backing string
		if err = rows.Scan(&identity.PrivatePath, &identity.Kind, &identity.State, &backing, &identity.PublicID); err != nil {
			break
		}
		if validateResourcePath(identity.PrivatePath) != nil || validatePublicID(identity.PublicID) != nil || identity.State != "live" || identity.Kind != graph.KindBead && identity.Kind != graph.KindLink {
			err = ErrInvalidStore
			break
		}
		if seen[identity.PublicID] {
			err = fmt.Errorf("%w: duplicate common Resource public identity", ErrInvalidStore)
			break
		}
		seen[identity.PublicID] = true
		if len(identities) == limit {
			err = ErrLimitExceeded
			break
		}
		identities = append(identities, identity)
	}
	if err == nil && rows.NextResultSet() {
		err = ErrInvalidStore
	}
	err = errors.Join(err, ctx.Err(), rows.Err(), rows.Close())
	if err != nil {
		return nil, err
	}
	return identities, nil
}
