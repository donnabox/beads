package graphstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

// ResourceKind names which retained plane holds a subject's version history.
// It is a local read-layer classification of the catalog's immutable backing,
// not a public BDP Type, a Type URL, or a wire enumeration. Two backings
// ("informational" and "dependency") share one kind because both are Links
// whose versions are retained identically; Issues are separated from Memory
// because their ordered history lives on Jim's native issue_versions table
// rather than in graph_preview_versions.
type ResourceKind string

const (
	KindIssue  ResourceKind = "issue"
	KindMemory ResourceKind = "memory"
	KindLink   ResourceKind = "link"
)

// VersionRow is one ordered version of a graph Resource, newest first.
//
// Ordinal is the ordering authority and is local to this store: two clones can
// both hold ordinal 8 for the same subject, each describing a different state
// (see internal/storage/issueops/version_history.go). Version is the opaque
// citable token, the only address a caller may hand back to ReadVersion or
// ReadVersionPair. ChangeAt is for display and time-based selection only; it is
// an observed wall clock, never the ordering key, and nothing here sorts by it.
//
// The JSON member names are fixed. Ordinal is deliberately not spelled
// "revision": graph records already use revision for the opaque token, and a
// native Issue uses it for the row-lock CAS token, so reusing the name here
// would invite a local ordering key to be read as a citable address.
type VersionRow struct {
	Ordinal     int64     `json:"ordinal"`   // ordering authority
	Version     string    `json:"version"`   // opaque citable token (feeds show --version / compare)
	ChangeAt    time.Time `json:"change_at"` // display and --at selection; never the ordering key
	Actor       string    `json:"actor"`
	Attribution string    `json:"attribution"` // native attribution_status for Issues; "" for Memory/Link
}

// PreviewVersionListLimit bounds this disposable non-paginated reader. Refusing
// is deliberate: silently returning a prefix would make an ordered history look
// complete, which is the one thing a History reader must not do.
const PreviewVersionListLimit = 1000

// Versions returns the ordered version list for one Resource path, newest first.
//
// Three outcomes are kept strictly distinct, because an empty list must never
// stand in for a refusal:
//   - no allocation of this path: ErrNotFound;
//   - an allocation whose plane holds no ordering authority:
//     ErrCapabilityUnavailable;
//   - an existing subject with nothing retained: (kind, nil, nil).
//
// Corrupt retained state remains ErrInvalidStore and an oversized history
// remains ErrLimitExceeded; neither is ever reported as an empty list.
//
// It supplies no snapshots, lineage, or diffs; ReadVersion resolves a returned
// token to a record. A deleted subject still has a history, so this reader does
// not refuse one, and it reports exactly what the retained tables hold. One
// listed token is therefore not resolvable as a Resource: a deleted Link's final
// retained version is its private deletion marker, which ReadVersion refuses
// with ErrGone. Whether to withhold that row is an open presentation question
// for the caller, not a correctness question for this reader. A deleted Memory's
// final live head is a real retained Resource (deletedMemoryInTx reads it).
func (s *Store) Versions(ctx context.Context, path string) (ResourceKind, []VersionRow, error) {
	if err := validateResourcePath(path); err != nil {
		return "", nil, err
	}
	var kind ResourceKind
	var rows []VersionRow
	err := s.withTx(ctx, false, func(tx *sql.Tx) error {
		if err := checkBinding(ctx, tx, s.options); err != nil {
			return err
		}
		var err error
		kind, rows, err = s.versionsInTx(ctx, tx, path)
		return err
	})
	if err != nil {
		return "", nil, err
	}
	return kind, rows, nil
}

// versionsInTx uses the caller's authority-checked transaction and validated
// path. Allocation, ordering and retained rows are all read in that single
// snapshot, so a concurrent writer cannot interleave a new head between the
// kind resolution and the list.
func (s *Store) versionsInTx(ctx context.Context, tx *sql.Tx, path string) (ResourceKind, []VersionRow, error) {
	var kind, backing, state, head string
	var key sql.NullString
	err := tx.QueryRowContext(ctx, `SELECT resource_kind,backing,allocation_state,revision,backing_key
 FROM graph_preview_catalog WHERE path=?`, path).Scan(&kind, &backing, &state, &head, &key)
	if errors.Is(err, sql.ErrNoRows) {
		// Retained rows without an allocation are corruption, not an absence.
		// Same refusal readVersionInTx makes: a History reader must not report
		// "never allocated" for a subject whose versions are still on disk.
		var retained int
		if err := tx.QueryRowContext(ctx, `SELECT
 (SELECT COUNT(*) FROM graph_preview_versions WHERE path=?) +
 (SELECT COUNT(*) FROM graph_preview_issue_versions WHERE path=?)`, path, path).Scan(&retained); err != nil {
			return "", nil, err
		}
		if retained != 0 {
			return "", nil, fmt.Errorf("%w: retained subject lacks allocation", ErrInvalidStore)
		}
		return "", nil, fmt.Errorf("%w: graph path has no allocation", ErrNotFound)
	}
	if err != nil {
		return "", nil, err
	}
	if !authorityID.MatchString(head) || (state != "live" && state != "deleted") ||
		(kind != "bead" && kind != "link") || (kind == "bead") != strings.HasPrefix(path, "beads/") {
		return "", nil, fmt.Errorf("%w: invalid retained subject allocation", ErrInvalidStore)
	}
	switch {
	case backing == "issue" && kind == "bead":
		// The Issue plane's ordinal is the native revision, so the mapping must
		// name the same Issue the catalog is bound to before it is trusted.
		if !key.Valid || key.String == "" {
			return "", nil, fmt.Errorf("%w: invalid Issue allocation", ErrInvalidStore)
		}
		rows, err := issueVersionsInTx(ctx, tx, path, key.String)
		return KindIssue, rows, err
	case backing == "generic" && kind == "bead":
		rows, err := previewVersionsInTx(ctx, tx, path)
		return KindMemory, rows, err
	case (backing == "informational" || backing == "dependency") && kind == "link":
		rows, err := previewVersionsInTx(ctx, tx, path)
		return KindLink, rows, err
	}
	return "", nil, fmt.Errorf("%w: unsupported retained allocation", ErrInvalidStore)
}

// issueVersionsInTx orders a graph Issue's history by the native revision the
// preview mapping points at. Those native rows exist in every graph workspace
// even though config versioned-history.enabled defaults to false: each graph
// Issue write path scopes its own transaction with
// issueops.ScopeVersionedHistoryTransaction(tx, true), which forces recording on
// for that transaction alone.
//
// Mappings with no native rows at all are still refused with
// ErrCapabilityUnavailable rather than answered with a short list: whatever the
// cause, this plane then holds no ordering authority, and an ordered history is
// the one thing this reader must not fake. A partial join is narrower than that
// and is corruption, since each mapping is written in the same transaction as
// the native row it names.
//
// The ORDER BY is load-bearing, not cosmetic: an unordered select over
// graph_preview_issue_versions returns rows in hash order, measured on a real
// graph workspace.
func issueVersionsInTx(ctx context.Context, tx *sql.Tx, path, backingKey string) ([]VersionRow, error) {
	var mapped int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM graph_preview_issue_versions WHERE path=?`, path).Scan(&mapped); err != nil {
		return nil, err
	}
	if mapped > PreviewVersionListLimit {
		return nil, fmt.Errorf("%w: at most %d retained versions can be returned; preview pagination is unavailable", ErrLimitExceeded, PreviewVersionListLimit)
	}
	if mapped == 0 {
		return nil, nil
	}
	// change_actor and attribution_status are VARCHAR on the native plane, so
	// the row count above already bounds this acquisition; no blob is selected.
	rows, err := tx.QueryContext(ctx, `SELECT m.issue_id,m.issue_revision,m.version,v.change_at,v.change_actor,v.attribution_status
 FROM graph_preview_issue_versions m JOIN issue_versions v ON v.issue_id=m.issue_id AND v.revision=m.issue_revision
 WHERE m.path=? ORDER BY v.revision DESC`, path)
	if err != nil {
		return nil, err
	}
	result := []VersionRow{}
	for rows.Next() {
		var issueID string
		var row VersionRow
		var actor, status sql.NullString
		if err := rows.Scan(&issueID, &row.Ordinal, &row.Version, &row.ChangeAt, &actor, &status); err != nil {
			return nil, errors.Join(err, rows.Close())
		}
		if issueID != backingKey {
			return nil, errors.Join(fmt.Errorf("%w: invalid Issue retained mapping", ErrInvalidStore), rows.Close())
		}
		// change_actor is nullable on the native plane; an unrecorded actor is
		// an empty Actor with attribution_status explaining it, never an error.
		row.Actor, row.Attribution = actor.String, status.String
		row.ChangeAt = row.ChangeAt.UTC()
		result = append(result, row)
	}
	if err := errors.Join(rows.Err(), rows.Close()); err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("%w: the Issue plane holds no retained versions to order", ErrCapabilityUnavailable)
	}
	if len(result) != mapped {
		return nil, fmt.Errorf("%w: mapped Issue retained body is missing", ErrInvalidStore)
	}
	return checkedVersionOrder(result)
}

// previewVersionsInTx orders Memory and Link history by graph_preview_versions'
// own ordinal, which the preview writer allocates (insertPreviewVersionInTx)
// rather than the Issue domain. Every retained row is reported, including a
// deleted Link's deletion marker; see Versions.
//
// This plane cannot lose its ordering authority the way the Issue plane can:
// ordinal is NOT NULL in the same row as the snapshot it orders, so there is no
// second table to fall out of step with. ErrCapabilityUnavailable is therefore
// reachable from the Issue plane only.
func previewVersionsInTx(ctx context.Context, tx *sql.Tx, path string) ([]VersionRow, error) {
	var retained int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM graph_preview_versions WHERE path=?`, path).Scan(&retained); err != nil {
		return nil, err
	}
	if retained > PreviewVersionListLimit {
		return nil, fmt.Errorf("%w: at most %d retained versions can be returned; preview pagination is unavailable", ErrLimitExceeded, PreviewVersionListLimit)
	}
	if retained == 0 {
		return nil, nil
	}
	// actor is LONGBLOB here, so corrupt or oversized rows must be charged in
	// SQL before any of them is acquired into Go. snapshot is never selected:
	// an ordered list is metadata, so a large history costs only its actors.
	// SUM may otherwise surface as floating point; clamp above the only
	// relevant threshold and cast, exactly as checkCurrentReadBytes does.
	var size uint64
	if err := tx.QueryRowContext(ctx, `SELECT CAST(LEAST(COALESCE(SUM(OCTET_LENGTH(actor)+256),0),?) AS UNSIGNED)
 FROM graph_preview_versions WHERE path=?`, PreviewCurrentReadByteLimit+1, path).Scan(&size); err != nil {
		return nil, err
	}
	if size > PreviewCurrentReadByteLimit {
		return nil, fmt.Errorf("%w: retained version list exceeds the %d-byte acquisition budget", ErrLimitExceeded, PreviewCurrentReadByteLimit)
	}
	rows, err := tx.QueryContext(ctx, `SELECT ordinal,version,change_at,actor FROM graph_preview_versions
 WHERE path=? ORDER BY ordinal DESC`, path)
	if err != nil {
		return nil, err
	}
	result := []VersionRow{}
	for rows.Next() {
		var row VersionRow
		if err := rows.Scan(&row.Ordinal, &row.Version, &row.ChangeAt, &row.Actor); err != nil {
			return nil, errors.Join(err, rows.Close())
		}
		row.ChangeAt = row.ChangeAt.UTC()
		// Attribution stays empty: this plane records an actor and derives
		// claimed/unknown from it at read time (validVersionAttribution). There
		// is no stored status column to report, and synthesizing one here would
		// claim a native field this plane does not have.
		result = append(result, row)
	}
	if err := errors.Join(rows.Err(), rows.Close()); err != nil {
		return nil, err
	}
	return checkedVersionOrder(result)
}

// checkedVersionOrder refuses a list the caller could not reason about: every
// token must be a well-formed address, every actor valid UTF-8, and the
// ordinals must descend strictly. Both planes allocate ordinals as MAX+1 under
// a unique key, so a repeated or non-positive ordinal is corruption rather than
// a tie this reader may break by some other column.
func checkedVersionOrder(rows []VersionRow) ([]VersionRow, error) {
	previous := int64(0)
	for i, row := range rows {
		if !authorityID.MatchString(row.Version) || !utf8.ValidString(row.Actor) ||
			!utf8.ValidString(row.Attribution) || row.Ordinal < 1 {
			return nil, fmt.Errorf("%w: invalid retained version row", ErrInvalidStore)
		}
		if i > 0 && row.Ordinal >= previous {
			return nil, fmt.Errorf("%w: retained versions are not strictly ordered", ErrInvalidStore)
		}
		previous = row.Ordinal
	}
	return rows, nil
}
