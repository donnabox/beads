package driver

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/steveyegge/beads/internal/replay/doltcli"
	"github.com/steveyegge/beads/internal/replay/oracle"
)

// ErrBaseNotSeeded reports an oracle whose base already holds issues, replayed
// into a work clone that was not seeded from it. Every step would be replayed
// onto an empty clone, so the first touch of an issue that existed at the base
// would fail partway through a long run with a message about the wrong thing.
var ErrBaseNotSeeded = errors.New("the oracle's base holds issues and the work clone was not seeded from it")

// checkSeedGuard refuses a base that holds issues when there is no seed. A seed is
// the record of the seeding that gave the work clone the base state; nothing
// produces one yet, so until it does a run replays only histories that begin
// empty. It runs before anything is written, so a refusal leaves no trace.
func checkSeedGuard(ctx context.Context, oracleDataDir string, w *Walk, seed *SeedRecord) error {
	if seed != nil {
		return nil
	}
	base := w.Base().Hash
	n, err := issuesAt(ctx, oracleDataDir, base)
	if err != nil {
		return fmt.Errorf("seed guard: %w", err)
	}
	if n > 0 {
		return fmt.Errorf("%w: %d issues at base commit %s", ErrBaseNotSeeded, n, base)
	}
	return nil
}

// issuesAt counts the rows of the issues table as of ref; a ref that predates the
// table holds none.
func issuesAt(ctx context.Context, dir, ref string) (int, error) {
	columns, err := oracle.Columns(ctx, dir, ref, "issues")
	if err != nil {
		return 0, err
	}
	if columns == nil {
		return 0, nil
	}
	_, rows, err := doltcli.Query(ctx, dir, "SELECT COUNT(*) FROM issues AS OF "+doltcli.SQLQuote(ref))
	if err != nil {
		return 0, fmt.Errorf("count the issues as of %s: %w", ref, err)
	}
	if len(rows) != 1 || len(rows[0]) != 1 {
		return 0, fmt.Errorf("count the issues as of %s: unexpected result %v", ref, rows)
	}
	n, err := strconv.Atoi(rows[0][0].Text)
	if err != nil {
		return 0, fmt.Errorf("count the issues as of %s: %q is not a count", ref, rows[0][0].Text)
	}
	return n, nil
}
