package driver

import (
	"context"
	"fmt"
	"strings"

	"github.com/steveyegge/beads/internal/replay/doltcli"
)

// Commit is one commit of the oracle's history, as dolt_log lists it.
type Commit struct {
	Hash string
	// Parents are the commit's parents, the mainline first. A root has none.
	Parents []string
	Message string
	// Date is the commit time as dolt prints it.
	Date string
}

// Step is the unit of replay: the net change from one commit of the first-parent
// chain to a later one. One step is replayed as one plan over every issue it
// touches, and each (step, issue) is compared and recorded once.
type Step struct {
	// Index is the step's position in the run, from 0.
	Index    int
	From, To Commit
	// Merge is set when To is a merge commit. The step is then the net diff from
	// the mainline parent to the merge, so the side branch's commits are covered by
	// it and are not steps of their own.
	Merge bool
	// Net is set when the step spans more than one commit of the chain, which only
	// sampled mode does.
	Net bool
}

// ReadLog reads dolt_log once and returns every commit reachable from HEAD, in
// no particular order, and the hash of HEAD. dolt_log promises no order, and
// gives sibling commits the same commit_order, so nothing here relies on either:
// the history is followed by parent pointers alone.
func ReadLog(ctx context.Context, dataDir string) (commits []Commit, head string, err error) {
	_, headRows, err := doltcli.Query(ctx, dataDir, "SELECT hashof('HEAD')")
	if err != nil {
		return nil, "", fmt.Errorf("read log: head of %s: %w", dataDir, err)
	}
	if len(headRows) != 1 || len(headRows[0]) != 1 || headRows[0][0].Text == "" {
		return nil, "", fmt.Errorf("read log: %s has no head commit", dataDir)
	}
	header, rows, err := doltcli.Query(ctx, dataDir, "SELECT commit_hash, parents, message, date FROM dolt_log")
	if err != nil {
		return nil, "", fmt.Errorf("read log: %w", err)
	}
	commits = make([]Commit, len(rows))
	for i, r := range rows {
		row := doltcli.RowMap(header, r)
		commits[i] = Commit{
			Hash:    row["commit_hash"],
			Parents: splitParents(row["parents"]),
			Message: row["message"],
			Date:    row["date"],
		}
	}
	return commits, headRows[0][0].Text, nil
}

// splitParents splits dolt_log's parents column, a comma and a space between
// hashes, mainline first, and empty for a root.
func splitParents(s string) []string {
	var parents []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			parents = append(parents, p)
		}
	}
	return parents
}

// FirstParentChain returns the commits on the first-parent chain from head back
// to the root, oldest first. It reads only parent pointers, so the order of
// commits does not matter. A head or a parent that is not among the commits, and
// a cycle, are errors: a log that cannot be followed is not replayed.
func FirstParentChain(commits []Commit, head string) ([]Commit, error) {
	byHash := make(map[string]Commit, len(commits))
	for _, c := range commits {
		byHash[c.Hash] = c
	}
	var newestFirst []Commit
	seen := make(map[string]bool, len(commits))
	for cur := head; ; {
		c, ok := byHash[cur]
		if !ok {
			return nil, fmt.Errorf("first-parent chain: commit %s is not in the log", cur)
		}
		if seen[cur] {
			return nil, fmt.Errorf("first-parent chain: cycle at commit %s", cur)
		}
		seen[cur] = true
		newestFirst = append(newestFirst, c)
		if len(c.Parents) == 0 {
			break
		}
		cur = c.Parents[0]
	}
	chain := make([]Commit, len(newestFirst))
	for i, c := range newestFirst {
		chain[len(newestFirst)-1-i] = c
	}
	return chain, nil
}

// Walk is the part of the oracle's history a run replays: the first-parent chain
// from the base to HEAD.
type Walk struct {
	chain []Commit
	// Total is every commit reachable from HEAD, side branches and the commits
	// before the base included, so a summary can say how much of the history the
	// steps stand for.
	Total int
}

// newWalk builds a Walk from a whole first-parent chain, root first, and the
// number of commits in the log. bd's one-time schema-bootstrap commits are
// collapsed to the last one, which is the base: the first content commit's own
// diff needs an earlier commit to start from, and history before the oracle's
// latest schema commit cannot be diffed (migration 0043 changed the primary key
// of dependencies), so it is not replayed.
func newWalk(chain []Commit, total int) *Walk {
	last := -1
	for i, c := range chain {
		if isBootstrapCommit(c.Message) {
			last = i
		}
	}
	if last > 0 {
		chain = chain[last:]
	}
	return &Walk{chain: chain, Total: total}
}

// ReadWalk reads the oracle's history and follows it from HEAD along first
// parents. The result is a coverage limit a run must state: a merge is one step,
// and what happened on its side branch is covered by that step's net diff.
func ReadWalk(ctx context.Context, dataDir string) (*Walk, error) {
	commits, head, err := ReadLog(ctx, dataDir)
	if err != nil {
		return nil, err
	}
	chain, err := FirstParentChain(commits, head)
	if err != nil {
		return nil, fmt.Errorf("walk %s: %w", dataDir, err)
	}
	return newWalk(chain, len(commits)), nil
}

// Base is the commit every step's state is a change from: the last bootstrap
// commit on the chain.
func (w *Walk) Base() Commit { return w.chain[0] }

// Head is the commit the last step ends at.
func (w *Walk) Head() Commit { return w.chain[len(w.chain)-1] }

// Steps returns the steps of the walk, oldest first. A sampleSize of zero or
// less is every commit of the chain, one step each; a positive sampleSize picks
// that many evenly spaced commits and replays the net change between neighbors.
func (w *Walk) Steps(sampleSize int) []Step {
	positions := make([]int, len(w.chain))
	for i := range positions {
		positions[i] = i
	}
	positions = selectSample(positions, sampleSize)
	var steps []Step
	for i := 0; i+1 < len(positions); i++ {
		from, to := w.chain[positions[i]], w.chain[positions[i+1]]
		steps = append(steps, Step{
			Index: i,
			From:  from,
			To:    to,
			Merge: len(to.Parents) > 1,
			Net:   positions[i+1]-positions[i] > 1,
		})
	}
	return steps
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
