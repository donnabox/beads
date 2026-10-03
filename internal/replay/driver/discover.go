package driver

import (
	"context"
	"fmt"
	"strings"

	"github.com/steveyegge/beads/internal/replay/doltcli"
	"github.com/steveyegge/beads/internal/storage/issueops"
)

// CommitInfo is one dolt_log row surfaced by ListCommits: enough to drive
// UC2's replay loop (Hash) and to tag a mutation_kind of "merge" regardless
// of row-diff kind (IsMerge), per MutationKindFor.
type CommitInfo struct {
	Hash    string
	IsMerge bool
}

// TouchedIssue is one issue (or one of its dependency edges) changed by a
// single commit, as discovered by DiscoverTouchedIssues.
type TouchedIssue struct {
	IssueID string
	Diff    RowDiffKind
}

// ListCommits returns every commit reachable from dataDir's currently
// checked-out branch, oldest first, each tagged with whether it has 2+
// parents (a real merge commit, as opposed to a fast-forward, which creates
// no commit at all). bd's own one-time schema-bootstrap commits are
// collapsed down to just the last one -- see isBootstrapCommit.
func ListCommits(ctx context.Context, dataDir string) ([]CommitInfo, error) {
	header, rows, err := doltcli.Query(ctx, dataDir, "SELECT commit_hash, parents, message FROM dolt_log ORDER BY commit_order ASC")
	if err != nil {
		return nil, fmt.Errorf("list commits: %w", err)
	}
	all := make([]CommitInfo, len(rows))
	lastBootstrap := -1
	for i, r := range rows {
		row := doltcli.RowMap(header, r)
		all[i] = CommitInfo{
			Hash:    row["commit_hash"],
			IsMerge: strings.Contains(row["parents"], ","),
		}
		if isBootstrapCommit(row["message"]) {
			lastBootstrap = i
		}
	}
	// Drop every bootstrap commit except the last: the first real content
	// commit's own diff needs an earlier "from" reference to be discovered at
	// all, and dropping the last bootstrap commit too would leave none.
	// Keeping it is safe -- by the time bd init's final bootstrap commit
	// lands, migration 0043's PK-shape change is long settled, so this
	// boundary commit can never trigger the PK-mismatch this filter exists to
	// avoid (confirmed empirically).
	if lastBootstrap < 0 {
		return all, nil
	}
	return all[lastBootstrap:], nil
}

// isBootstrapCommit reports whether message is one of bd's own one-time
// schema-bootstrap commits rather than a real content mutation (always a
// "bd: ..." message). Covers dolt's default first commit, bd init's own
// commit, the migration runner's "schema: "-prefixed wrapper and inline
// commits, and two migrations that commit internally under other literals:
// 0040's four "create nonlocal table <name>" commits and 0041's "disable
// nonlocal tables for fk migrations" (plus migration_repairs.go's
// "repair: "-prefixed partial-migration recovery commits, same family).
// Verified exhaustively by grepping every CALL DOLT_COMMIT literal under
// internal/storage/schema, not just what one fixture's
// history happened to exercise. Matters because these commits can straddle
// the dependencies table's PK shape change in migration 0043, which
// dolt_commit_diff_dependencies cannot diff across ("could not map primary
// key column issue_id").
func isBootstrapCommit(message string) bool {
	switch message {
	case "Initialize data repository", "bd init", "disable nonlocal tables for fk migrations":
		return true
	}
	for _, prefix := range []string{"schema: ", "create nonlocal table ", "repair: "} {
		if strings.HasPrefix(message, prefix) {
			return true
		}
	}
	return false
}

// DiscoverTouchedIssues returns every issue touched between from (exclusive)
// and to (inclusive) in dataDir: a row-level add/remove/modify on the issues
// table, or a dependency-edge add/remove.
func DiscoverTouchedIssues(ctx context.Context, dataDir, from, to string) ([]TouchedIssue, error) {
	for _, ref := range []string{from, to} {
		if err := issueops.ValidateRef(ref); err != nil {
			return nil, fmt.Errorf("discover touched issues: %w", err)
		}
	}

	touched, err := discoverIssueRows(ctx, dataDir, from, to)
	if err != nil {
		return nil, err
	}
	depTouched, err := discoverDependencyRows(ctx, dataDir, from, to)
	if err != nil {
		return nil, err
	}
	return append(touched, depTouched...), nil
}

func discoverIssueRows(ctx context.Context, dataDir, from, to string) ([]TouchedIssue, error) {
	query := fmt.Sprintf(
		"SELECT * FROM dolt_commit_diff_issues WHERE to_commit=%s AND from_commit=%s",
		doltcli.SQLQuote(to), doltcli.SQLQuote(from))
	header, rows, err := doltcli.Query(ctx, dataDir, query)
	if err != nil {
		return nil, fmt.Errorf("discover issue rows: %w", err)
	}

	touched := make([]TouchedIssue, 0, len(rows))
	for _, r := range rows {
		row := doltcli.RowMap(header, r)
		id := row["to_id"]
		if id == "" {
			id = row["from_id"]
		}
		var kind RowDiffKind
		switch row["diff_type"] {
		case "added":
			kind = RowDiffAdded
		case "removed":
			kind = RowDiffRemoved
		case "modified":
			kind = RowDiffModified
		default:
			return nil, fmt.Errorf("discover issue rows for %s: unrecognized diff_type %q", id, row["diff_type"])
		}
		touched = append(touched, TouchedIssue{IssueID: id, Diff: kind})
	}
	return touched, nil
}

func discoverDependencyRows(ctx context.Context, dataDir, from, to string) ([]TouchedIssue, error) {
	query := fmt.Sprintf(
		"SELECT * FROM dolt_commit_diff_dependencies WHERE to_commit=%s AND from_commit=%s",
		doltcli.SQLQuote(to), doltcli.SQLQuote(from))
	header, rows, err := doltcli.Query(ctx, dataDir, query)
	if err != nil {
		return nil, fmt.Errorf("discover dependency rows: %w", err)
	}

	touched := make([]TouchedIssue, 0, len(rows))
	for _, r := range rows {
		row := doltcli.RowMap(header, r)
		switch row["diff_type"] {
		case "added":
			touched = append(touched, TouchedIssue{IssueID: row["to_issue_id"], Diff: RowDiffDepAdded})
		case "removed":
			touched = append(touched, TouchedIssue{IssueID: row["from_issue_id"], Diff: RowDiffDepRemoved})
		default:
			return nil, fmt.Errorf("discover dependency rows: unrecognized diff_type %q", row["diff_type"])
		}
	}
	return touched, nil
}
