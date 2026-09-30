// Package oracle is the replay harness's read side: it returns the row an
// issue had at a given commit of a corpus clone, straight from dolt, as a plain
// field-keyed map ready for canonicalization without translation.
//
// It issues only a fixed SELECT against the local clone directory, through
// `dolt sql -q ... AS OF <ref> ...` (never `bd sql`, which refuses outright in
// embedded mode), and never a write or a connection to a shared server
// (NFR1/NFR2).
package oracle

import (
	"context"
	"fmt"

	"github.com/steveyegge/beads/internal/replay/doltcli"
	"github.com/steveyegge/beads/internal/storage/issueops"
)

// QueryAsOf returns the row for issueID in the issues table of the Dolt
// database at dir, as it stood at commit ref, or nil if no such row existed at
// that commit (not yet created, or since deleted). dir must be a local,
// unserved clone directory; doltcli refuses one that a dolt sql-server is
// serving.
//
// ref and issueID are spliced directly into the `dolt sql -q` argument (dolt's
// AS OF clause and CLI have no bind-parameter API), so both are
// charset-validated first via issueops.ValidateRef.
func QueryAsOf(ctx context.Context, dir, ref, issueID string) (map[string]string, error) {
	if err := issueops.ValidateRef(ref); err != nil {
		return nil, fmt.Errorf("oracle: invalid ref %q: %w", ref, err)
	}
	if err := issueops.ValidateRef(issueID); err != nil {
		return nil, fmt.Errorf("oracle: invalid issue id %q: %w", issueID, err)
	}

	query := fmt.Sprintf("SELECT * FROM issues AS OF '%s' WHERE id = '%s'", ref, issueID) //nolint:gosec // G201 -- dolt sql -q has no bind-parameter API; ref/issueID are charset-validated above via issueops.ValidateRef
	header, rows, err := doltcli.Query(ctx, dir, query)
	if err != nil {
		return nil, fmt.Errorf("oracle: querying %s AS OF %s: %w", issueID, ref, err)
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return doltcli.RowMap(header, rows[0]), nil
}
