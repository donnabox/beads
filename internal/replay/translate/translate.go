// Package translate turns a historical dolt_log commit's row-diff for one
// issue into the equivalent bd CLI invocation(s) and executes them against a
// working clone -- never a hand-written row insert.
package translate

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/steveyegge/beads/internal/replay/doltcli"
	"github.com/steveyegge/beads/internal/storage/issueops"
)

// ActionKind identifies which bd CLI operation an Action represents.
type ActionKind int

const (
	KindCreate ActionKind = iota
	KindUpdate
	KindDelete
	KindDepAdd
	KindDepRemove
	KindClose
	KindNoop
)

// Action is one bd CLI invocation that replays a single historical mutation
// against a working clone. KindNoop carries no invocation at all.
type Action struct {
	Kind ActionKind
	Argv []string
}

// fieldFlag pairs an issues-table column with the bd flag that sets it on
// create/update. Order is fixed so generated argv is deterministic.
type fieldFlag struct {
	column string
	flag   string
}

// settableFields lists every issues-table column this translator can express
// as a bd create/update flag. A changed column not in this list is a hard
// error (never silently skip a real row change), not a dropped mutation.
var settableFields = []fieldFlag{
	{"title", "--title"},
	{"description", "--description"},
	{"design", "--design"},
	{"acceptance_criteria", "--acceptance"},
	{"notes", "--notes"},
	{"priority", "--priority"},
	{"issue_type", "--type"},
	{"assignee", "--assignee"},
	{"external_ref", "--external-ref"},
	{"spec_id", "--spec-id"},
	{"estimated_minutes", "--estimate"},
	{"status", "--status"},
}

// ignoredColumns change as a side effect of any write (content_hash,
// updated_at, row_lock), as a side effect of a dependency-edge change
// specifically (is_blocked), or are dolt_commit_diff_issues' own range
// metadata rather than real issues-table columns (commit, commit_date) --
// confirmed empirically rather than assumed from the schema alone.
var ignoredColumns = map[string]bool{
	"content_hash": true,
	"updated_at":   true,
	"row_lock":     true,
	"is_blocked":   true,
	"commit":       true,
	"commit_date":  true,
}

// closingColumns are excluded from a modified-row diff once a close
// transition has already been captured as its own KindClose action --
// otherwise close_reason/closed_at/closed_by_session would spuriously
// surface as unsupported field changes on every close.
var closingColumns = map[string]bool{
	"status":            true,
	"close_reason":      true,
	"closed_at":         true,
	"closed_by_session": true,
}

// Classify reads the dolt_log row-diff for issueID between from (exclusive)
// and to (inclusive) in the dolt data directory dataDir, and returns the bd
// CLI invocation(s) that replay the equivalent mutation(s) against a working
// clone. Falls back to a single KindNoop when nothing material changed.
func Classify(ctx context.Context, dataDir, from, to, issueID string) ([]Action, error) {
	for _, ref := range []string{from, to, issueID} {
		if err := issueops.ValidateRef(ref); err != nil {
			return nil, fmt.Errorf("classify: %w", err)
		}
	}

	issueActions, err := classifyIssueRow(ctx, dataDir, from, to, issueID)
	if err != nil {
		return nil, err
	}
	depActions, err := classifyDependencyRows(ctx, dataDir, from, to, issueID)
	if err != nil {
		return nil, err
	}

	actions := append(issueActions, depActions...)
	if len(actions) == 0 {
		return []Action{{Kind: KindNoop}}, nil
	}
	return actions, nil
}

func classifyIssueRow(ctx context.Context, dataDir, from, to, issueID string) ([]Action, error) {
	query := fmt.Sprintf(
		"SELECT * FROM dolt_commit_diff_issues WHERE to_commit=%s AND from_commit=%s AND (to_id=%s OR from_id=%s)",
		doltcli.SQLQuote(to), doltcli.SQLQuote(from), doltcli.SQLQuote(issueID), doltcli.SQLQuote(issueID))
	header, rows, err := doltcli.Query(ctx, dataDir, query)
	if err != nil {
		return nil, fmt.Errorf("classify issues row for %s: %w", issueID, err)
	}
	if len(rows) == 0 {
		return nil, nil
	}
	if len(rows) > 1 {
		return nil, fmt.Errorf("classify issues row for %s: expected at most one row, got %d", issueID, len(rows))
	}
	row := doltcli.RowMap(header, rows[0])

	switch row["diff_type"] {
	case "added":
		return []Action{buildCreateAction(row, issueID)}, nil
	case "removed":
		return []Action{{Kind: KindDelete, Argv: []string{"delete", issueID, "--force"}}}, nil
	case "modified":
		return buildModifiedActions(row, issueID)
	default:
		return nil, fmt.Errorf("classify issues row for %s: unrecognized diff_type %q", issueID, row["diff_type"])
	}
}

func buildCreateAction(row map[string]string, issueID string) Action {
	// --force: a work clone's bd project almost always has a different ID
	// prefix than the source it's replaying from, which bd's create
	// otherwise refuses as a prefix mismatch (confirmed empirically).
	argv := []string{"create", row["to_title"], "--id", issueID, "--force"}
	for _, ff := range settableFields {
		if ff.column == "title" {
			continue // positional above, not a flag on create
		}
		if v := row["to_"+ff.column]; v != "" {
			argv = append(argv, ff.flag, v)
		}
	}
	return Action{Kind: KindCreate, Argv: argv}
}

func buildModifiedActions(row map[string]string, issueID string) ([]Action, error) {
	changed := make(map[string]bool)
	for k, v := range row {
		if !strings.HasPrefix(k, "to_") {
			continue
		}
		base := k[len("to_"):]
		fromKey := "from_" + base
		fromVal, ok := row[fromKey]
		if !ok || ignoredColumns[base] {
			continue
		}
		if v != fromVal {
			changed[base] = true
		}
	}

	closing := row["from_status"] != "closed" && row["to_status"] == "closed"
	if closing {
		for c := range closingColumns {
			delete(changed, c)
		}
	}

	var actions []Action
	if len(changed) > 0 {
		var updateArgv []string
		for _, ff := range settableFields {
			if !changed[ff.column] {
				continue
			}
			updateArgv = append(updateArgv, ff.flag, row["to_"+ff.column])
			delete(changed, ff.column)
		}
		if len(changed) > 0 {
			unmapped := make([]string, 0, len(changed))
			for c := range changed {
				unmapped = append(unmapped, c)
			}
			sort.Strings(unmapped)
			return nil, fmt.Errorf("classify issues row for %s: unsupported field change(s): %s", issueID, strings.Join(unmapped, ", "))
		}
		actions = append(actions, Action{Kind: KindUpdate, Argv: append([]string{"update", issueID}, updateArgv...)})
	}

	if closing {
		argv := []string{"close", issueID}
		if reason := row["to_close_reason"]; reason != "" {
			argv = append(argv, "--reason", reason)
		}
		actions = append(actions, Action{Kind: KindClose, Argv: argv})
	}

	return actions, nil
}

func classifyDependencyRows(ctx context.Context, dataDir, from, to, issueID string) ([]Action, error) {
	query := fmt.Sprintf(
		"SELECT * FROM dolt_commit_diff_dependencies WHERE to_commit=%s AND from_commit=%s AND (to_issue_id=%s OR from_issue_id=%s)",
		doltcli.SQLQuote(to), doltcli.SQLQuote(from), doltcli.SQLQuote(issueID), doltcli.SQLQuote(issueID))
	header, rows, err := doltcli.Query(ctx, dataDir, query)
	if err != nil {
		return nil, fmt.Errorf("classify dependency rows for %s: %w", issueID, err)
	}

	var actions []Action
	for _, r := range rows {
		row := doltcli.RowMap(header, r)
		switch row["diff_type"] {
		case "added":
			actions = append(actions, Action{
				Kind: KindDepAdd,
				Argv: []string{"dep", "add", row["to_issue_id"], row["to_depends_on_issue_id"], "--type", row["to_type"]},
			})
		case "removed":
			actions = append(actions, Action{
				Kind: KindDepRemove,
				Argv: []string{"dep", "remove", row["from_issue_id"], row["from_depends_on_issue_id"]},
			})
		default:
			return nil, fmt.Errorf("classify dependency row for %s: unsupported diff_type %q", issueID, row["diff_type"])
		}
	}
	return actions, nil
}

// Execute runs one Action's bd CLI invocation against workDir, resolving bd on
// PATH. KindNoop returns immediately without spawning a process.
func Execute(ctx context.Context, workDir string, action Action) error {
	if action.Kind == KindNoop {
		return nil
	}
	bdPath, err := exec.LookPath("bd")
	if err != nil {
		return fmt.Errorf("execute: %w", err)
	}
	return ExecuteWith(ctx, bdPath, workDir, action)
}

// ExecuteWith is Execute with an explicit bd binary, for callers that must not
// depend on whatever bd PATH resolves to. A relative bdPath is taken relative to
// the caller's directory: bd runs with workDir as its working directory, which
// would otherwise change what the path means.
func ExecuteWith(ctx context.Context, bdPath, workDir string, action Action) error {
	if action.Kind == KindNoop {
		return nil
	}
	absBd, err := filepath.Abs(bdPath)
	if err != nil {
		return fmt.Errorf("execute: resolving %s: %w", bdPath, err)
	}
	cmd := exec.CommandContext(ctx, absBd, action.Argv...)
	cmd.Dir = workDir
	cmd.Env = doltcli.SanitizedEnv(os.Environ())
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("bd %s: %w\n%s", strings.Join(action.Argv, " "), err, out)
	}
	return nil
}
