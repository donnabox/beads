package issueops

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/storage/dberrors"
	"github.com/steveyegge/beads/internal/types"
)

// DeferWakeActor is the actor recorded on the status_changed event when the
// wake sweep returns an expired dated defer to open. A constant rather than
// the invoking session's actor: the wake is the system honoring the defer
// date, not something the reader who happened to trigger it did.
const DeferWakeActor = "bd-defer-wake"

// WakeDefersResult reports what one sweep woke, per table.
type WakeDefersResult struct {
	// Issues are the permanent-table issue ids returned to open; only these
	// affect Dolt-versioned tables, so only these decide whether the caller
	// mints a dolt commit.
	Issues []string
	// Wisps are the woken wisp ids. Wisp tables are dolt_ignored, so a
	// wisp-only wake never needs a version commit.
	Wisps []string
	// Skipped are the expired defers the sweep did not wake because versioned
	// history is on and the mint would refuse the wake: the issue holds a
	// number outside the I-JSON exact-integer range, in its metadata or in a
	// gate's timeout. Each is left exactly as it was, still deferred and
	// expired, so every later sweep meets it again until the value is
	// corrected. Nothing was written for it, so it does not count toward
	// whether the caller has anything to commit.
	Skipped []UnversionableIssue
}

// WakeDefersCommitMessage names a sweep's dolt commit. n is the number of
// permanent issues woken (len(result.Issues)); callers with n == 0 should not
// commit at all.
func WakeDefersCommitMessage(n int) string {
	return fmt.Sprintf("bd: wake %d expired defer(s)", n)
}

// WakeExpiredDefersInTx returns every DATED defer whose date has passed to
// open: status='deferred' AND defer_until <= now flips to status='open',
// defer_until=NULL — byte-identical to what `bd undefer` writes, so a later
// dateless `bd defer` cannot inherit a stale past date and instantly re-wake.
// A DATELESS defer (defer_until IS NULL) is the indefinite icebox and is
// deliberately never touched: `bd undefer` stays its only exit.
//
// This is the lazy half of the defer contract. `bd defer --until` and
// `bd update --defer` promise "hidden until <date>", but nothing ever flipped
// the status back, so an expired defer stayed invisible to the ready front
// forever. Callers run this sweep at the top of ready-work reads and claims;
// the snapshot-then-recheck shape mirrors ReclaimExpiredLeasesInTx, so a bead
// re-deferred or claimed between the snapshot and its UPDATE is simply skipped.
//
// With versioned history on, each wake mints a version of the issue, and the
// mint refuses a row that holds a value it cannot record faithfully. A refusal
// there would abort the whole transaction, so one such row would stop every
// other expired defer from waking, on every read. The sweep therefore asks the
// mint's own question about each row before it writes anything for it: a row
// the mint would refuse is skipped, left deferred and expired, and reported in
// result.Skipped (and once per process on stderr) while the rest of the batch
// wakes. With history off there is no such check and no extra read.
//
// The caller owns Dolt versioning (commit iff len(result.Issues) > 0) and must
// treat sweep failure as advisory — a ready listing never fails because the
// wake could not run. A skipped row is not a failure: the error is nil.
func WakeExpiredDefersInTx(ctx context.Context, tx DBTX) (WakeDefersResult, error) {
	var result WakeDefersResult
	issues, skipped, err := wakeExpiredDefersInTable(ctx, tx, "issues", "events")
	if err != nil {
		return result, err
	}
	result.Issues = issues
	result.Skipped = skipped
	// Wisps carry the same status/defer_until columns and `bd defer` reaches
	// them through the same UpdateIssue routing. The table is tolerated absent
	// for pre-wisp databases, like every other wisp probe.
	wisps, wispSkipped, err := wakeExpiredDefersInTable(ctx, tx, "wisps", "wisp_events")
	if err != nil {
		if dberrors.IsTableNotExist(err) {
			return result, nil
		}
		return result, err
	}
	result.Wisps = wisps
	result.Skipped = append(result.Skipped, wispSkipped...)
	return result, nil
}

// deferWakeSkipWarned holds the ids whose skipped wake this process has already
// reported. A skipped row stays deferred and expired, so every sweep meets it
// again, and naming it on each ready-work read would repeat one fact about the
// data forever.
var deferWakeSkipWarned sync.Map // issue id -> struct{}

// warnDeferWakeSkippedOnce names a skipped row on stderr, once per process per
// id: the issue, the field the refusal was found in and the refusal itself.
func warnDeferWakeSkippedOnce(skipped UnversionableIssue) {
	if _, seen := deferWakeSkipWarned.LoadOrStore(skipped.ID, struct{}{}); seen {
		return
	}
	fmt.Fprintf(os.Stderr, "warning: defer-wake sweep left %s deferred: waking it would be refused by versioned history (%v); it stays deferred until that value is corrected\n",
		skipped.ID, skipped.Err)
}

// wakeRefusal reports whether the mint would refuse the wake of id, which the
// sweep asks before it writes anything for the row; nil means the mint accepts
// the row or skips it. It loads the row as the mint does (GetIssueInTx routes a
// wisp to the wisps table) and decides in the mint's order: mintSkipsRow, which
// is the mint's own list of rows it never versions, then the admission gate over
// the whole issue. A row the mint skips is never refused, so a no-history row, a
// wisp and a legacy record holding the same value all wake.
//
// The wake changes only status, defer_until and updated_at, so the row the mint
// sees after it refuses exactly when this one does. The mint also loads the
// issue's dependencies before its gate, which this does not: a dependency's
// metadata is a string, so it adds no number to the issue the gate reads.
//
// An error is a failed read, or storage.ErrNotFound when the row is gone; it is
// never a refusal.
func wakeRefusal(ctx context.Context, tx DBTX, id string) (*UnversionableIssue, error) {
	issue, err := GetIssueInTx(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	if skips, err := mintSkipsRow(ctx, tx, id, issue, mintUpdate); err != nil || skips {
		return nil, err
	}
	if refused := FindUnversionable([]*types.Issue{issue}); len(refused) > 0 {
		return &refused[0], nil
	}
	return nil, nil
}

func wakeExpiredDefersInTable(ctx context.Context, tx DBTX, table, eventsTable string) ([]string, []UnversionableIssue, error) {
	// Snapshot first so each genuinely-woken row gets its own event. The
	// UPDATE below repeats the whole predicate, so a row rescued between the
	// SELECT and its UPDATE (re-deferred further out, claimed, closed) matches
	// nothing and is skipped rather than clobbered.
	//nolint:gosec // G201: table is a hardcoded constant from the caller above.
	rows, err := tx.QueryContext(ctx, fmt.Sprintf(`
		SELECT id FROM %s
		WHERE status = 'deferred' AND defer_until IS NOT NULL
		  AND defer_until <= UTC_TIMESTAMP()
	`, table))
	if err != nil {
		return nil, nil, fmt.Errorf("wake expired defers: scan %s: %w", table, err)
	}
	var expired []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			_ = rows.Close()
			return nil, nil, fmt.Errorf("wake expired defers: scan %s row: %w", table, err)
		}
		expired = append(expired, id)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, nil, fmt.Errorf("wake expired defers: iterate %s: %w", table, err)
	}
	if err := rows.Close(); err != nil {
		return nil, nil, fmt.Errorf("wake expired defers: close %s rows: %w", table, err)
	}
	if len(expired) == 0 {
		return nil, nil, nil
	}

	// The mint at the end of each wake can refuse the row, and its error would
	// abort the transaction. So with history on, each row is checked against the
	// mint before anything is written for it (see wakeRefusal). The flag is read
	// once for the pass, from the transaction the mint reads it from; with it
	// off there is no check and no extra read. There is no per-table branch:
	// the mint's own rules already exempt a wisp.
	checkMint := versionedHistoryEnabled(tx)

	var woken []string
	var skipped []UnversionableIssue
	now := time.Now().UTC()
	for _, id := range expired {
		if checkMint {
			refused, err := wakeRefusal(ctx, tx, id)
			if errors.Is(err, storage.ErrNotFound) {
				continue // rescued concurrently — leave it be
			}
			if err != nil {
				return woken, skipped, fmt.Errorf("wake expired defer %s: %w", id, err)
			}
			if refused != nil {
				skipped = append(skipped, *refused)
				warnDeferWakeSkippedOnce(*refused)
				continue
			}
		}
		// row_lock is rewritten so a concurrent claim/update conflicts at
		// commit time instead of cell-merging with this write — the same
		// invariant the lease scheme depends on.
		//nolint:gosec // G201: table is a hardcoded constant from the caller above.
		res, err := tx.ExecContext(ctx, fmt.Sprintf(`
			UPDATE %s
			SET status = 'open', defer_until = NULL, updated_at = ?, row_lock = ?
			WHERE id = ? AND status = 'deferred' AND defer_until IS NOT NULL
			  AND defer_until <= UTC_TIMESTAMP()
		`, table), now, freshRowLock(), id)
		if err != nil {
			return woken, skipped, fmt.Errorf("wake expired defer %s: %w", id, err)
		}
		n, err := res.RowsAffected()
		if err != nil {
			return woken, skipped, fmt.Errorf("wake expired defer %s rows affected: %w", id, err)
		}
		if n == 0 {
			continue // rescued concurrently — leave it be
		}
		if err := RecordFullEventInTable(ctx, tx, eventsTable, id, types.EventStatusChanged,
			DeferWakeActor, string(types.StatusDeferred), string(types.StatusOpen)); err != nil {
			return woken, skipped, fmt.Errorf("record wake event for %s: %w", id, err)
		}
		// A wake is a status change, so it journals as an update. Emitted past
		// the rows-affected re-check, so a concurrently-rescued bead records
		// nothing.
		if err := RecordEventInTx(ctx, tx, EventUpdate, id, DeferWakeActor); err != nil {
			return woken, skipped, err
		}
		// The status flip is durable state, so the woken bead is versioned
		// under the same system actor the event carries.
		if err := RecordVersionInTx(ctx, tx, id, DeferWakeActor); err != nil {
			return woken, skipped, err
		}
		woken = append(woken, id)
	}
	return woken, skipped, nil
}
