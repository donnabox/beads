package graphstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"

	graph "github.com/steveyegge/beads/graphops"
)

// PreviewSnapshotLimit bounds this internal, complete current-state read. It is
// an operational preview limit, not BDP pagination or a claimed Read profile.
const PreviewSnapshotLimit = 1000

// Snapshot contains authoritative live records and installed descriptors from
// one transaction. WriterToken identifies controlled-writer state; it is not a
// public revision, historical selector, or proof against out-of-band SQL writes.
type Snapshot struct {
	WriterToken string
	Records     []any
	Types       []graph.TypeDescriptor
}

// CurrentSnapshot returns a complete bounded current inventory, or zero state
// and an error. It does not retain a database transaction between caller requests.
func (s *Store) CurrentSnapshot(ctx context.Context) (Snapshot, error) {
	var result Snapshot
	err := s.withTx(ctx, false, func(tx *sql.Tx) error {
		var err error
		result, err = s.currentSnapshotInTx(ctx, tx)
		return err
	})
	if err != nil {
		return Snapshot{}, err
	}
	return result, nil
}

func (s *Store) currentSnapshotInTx(ctx context.Context, tx *sql.Tx) (Snapshot, error) {
	if err := checkBinding(ctx, tx, s.options); err != nil {
		return Snapshot{}, err
	}
	if err := s.checkCollectionMappingsInTx(ctx, tx); err != nil {
		return Snapshot{}, err
	}
	result := Snapshot{Records: []any{}, Types: []graph.TypeDescriptor{}}
	if err := tx.QueryRowContext(ctx, `SELECT writer_token FROM graph_preview_scope WHERE singleton=1`).Scan(&result.WriterToken); err != nil {
		return Snapshot{}, err
	}
	// The preview admits exactly four immutable descriptors. Do not silently
	// advertise only a subset if an unsupported installation appears.
	var typeCount int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM graph_preview_types`).Scan(&typeCount); err != nil {
		return Snapshot{}, err
	}
	if typeCount != 4 {
		return Snapshot{}, fmt.Errorf("%w: unsupported Type installation", ErrInvalidStore)
	}
	for _, id := range []string{MemoryTypeURL(s.ScopeURL()), IssueTypeURL(s.ScopeURL()), DependencyTypeURL(s.ScopeURL()), RelatedTypeURL(s.ScopeURL())} {
		descriptor, err := s.readTypeInTx(ctx, tx, id)
		if err != nil {
			return Snapshot{}, err
		}
		result.Types = append(result.Types, descriptor)
	}
	sort.Slice(result.Types, func(i, j int) bool { return graph.CompareCodeUnits(result.Types[i].ID(), result.Types[j].ID()) < 0 })
	var invalid int
	err := tx.QueryRowContext(ctx, `SELECT 1 FROM graph_preview_catalog WHERE allocation_state NOT IN ('live','deleted') LIMIT 1`).Scan(&invalid)
	if err == nil {
		return Snapshot{}, fmt.Errorf("%w: unknown allocation state", ErrInvalidStore)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return Snapshot{}, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT path,backing,resource_kind FROM graph_preview_catalog WHERE allocation_state='live' LIMIT ?`, PreviewSnapshotLimit+1)
	if err != nil {
		return Snapshot{}, err
	}
	type allocation struct{ path, backing, kind string }
	allocations := []allocation{}
	for rows.Next() {
		var a allocation
		if err := rows.Scan(&a.path, &a.backing, &a.kind); err != nil {
			return Snapshot{}, errors.Join(err, rows.Close())
		}
		allocations = append(allocations, a)
	}
	if err := errors.Join(rows.Err(), rows.Close()); err != nil {
		return Snapshot{}, err
	}
	if len(allocations) > PreviewSnapshotLimit {
		return Snapshot{}, fmt.Errorf("%w: complete current inventory exceeds %d live Resources", ErrLimitExceeded, PreviewSnapshotLimit)
	}
	sort.Slice(allocations, func(i, j int) bool { return graph.CompareCodeUnits(allocations[i].path, allocations[j].path) < 0 })
	for _, a := range allocations {
		if err := validateResourcePath(a.path); err != nil {
			return Snapshot{}, fmt.Errorf("%w: malformed catalog path: %v", ErrInvalidStore, err)
		}
		if (a.kind != "bead" && a.kind != "link") || (a.kind == "link") != strings.HasPrefix(a.path, "links/") {
			return Snapshot{}, fmt.Errorf("%w: catalog kind/path mismatch", ErrInvalidStore)
		}
		var record any
		var err error
		switch a.backing {
		case "generic":
			record, err = s.showMemoryInTx(ctx, tx, a.path)
		case "issue":
			record, err = s.showIssueInTx(ctx, tx, a.path)
		case "dependency", "informational":
			record, err = s.showLinkInTx(ctx, tx, a.path)
		default:
			err = fmt.Errorf("%w: unsupported backing", ErrInvalidStore)
		}
		if err != nil {
			return Snapshot{}, err
		}
		result.Records = append(result.Records, record)
	}
	return result, nil
}

// A collection must not hide current authoritative rows through missing catalog
// joins. This preview is a newly initialized, controlled-writer workspace, not
// an adopted legacy database with separately managed Issues.
func (s *Store) checkCollectionMappingsInTx(ctx context.Context, tx *sql.Tx) error {
	if err := s.checkLinkMappingsInTx(ctx, tx); err != nil {
		return err
	}
	for _, query := range []string{
		`SELECT COUNT(*) FROM graph_preview_payloads p LEFT JOIN graph_preview_catalog c ON c.path=p.path WHERE c.path IS NULL OR c.resource_kind<>'bead' OR c.backing<>'generic' OR c.allocation_state<>'live'`,
		`SELECT COUNT(*) FROM issues i LEFT JOIN graph_preview_catalog c ON c.backing='issue' AND c.backing_key=i.id WHERE c.path IS NULL OR c.resource_kind<>'bead' OR c.allocation_state<>'live'`,
		`SELECT COUNT(*) FROM graph_preview_catalog WHERE allocation_state='deleted' AND (resource_kind<>'link' OR backing<>'informational')`,
	} {
		var invalid int
		if err := tx.QueryRowContext(ctx, query).Scan(&invalid); err != nil {
			return err
		}
		if invalid != 0 {
			return fmt.Errorf("%w: incomplete current authority mapping", ErrInvalidStore)
		}
	}
	return nil
}
