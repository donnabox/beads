package doctor

import (
	"context"
	"fmt"
	"strings"

	storageissueops "github.com/steveyegge/beads/internal/storage/issueops"
	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/issueops"
)

const versionableIssuesName = "Versionable Issues"

// versionableIssuesListLimit is how many offending issues the report names. The count
// it leads with is always the whole number.
const versionableIssuesListLimit = 20

// CheckVersionableIssuesWithStore reports issues that recording a version would
// refuse: a number outside the exact-integer range (magnitude above 2^53-1) in an
// issue's metadata, or a gate whose timeout, held as nanoseconds, is past that bound
// (about 104.25 days). With versioned history on, a write that introduces such a
// value fails in its own transaction, but a row that already holds one, written while
// history was off or by a writer that does not record, would fail every later write
// to it once it participates in history.
//
// It is the check `bd config set versioned-history.enabled true` makes once, kept as a
// standing control. The environment and config.yaml planes can also turn history on,
// and they do not pass through that command, so a store switched on by one of them
// never gets its refusal; this is what finds the rows beforehand, and it can be run
// against any store at any time.
//
// It runs the one function recording runs, over whole issues, through
// issueops.FindUnversionable: the check and the mint cannot disagree because there is
// no second copy of the rule. It reads what the switch reads, every row of both
// planes, and the scan skips the rows recording never versions (ephemeral and
// no-history rows). It cannot see participation, so a legacy row holding such a value
// is named too, which can only make it refuse more.
//
// A store that cannot be read does not pass: a check that could not run has not
// passed, so the answer is a warning that says so. It is deliberately not
// auto-fixable: replacing a value or removing a gate is the owner's decision.
func CheckVersionableIssuesWithStore(ss *SharedStore) DoctorCheck {
	store := ss.Store()
	if store == nil {
		return DoctorCheck{
			Name:    versionableIssuesName,
			Status:  StatusOK,
			Message: "No database yet",
		}
	}
	reader, err := store.IssueReader()
	if err != nil {
		return versionableIssuesNotRun(err)
	}
	return checkVersionableIssues(context.Background(), reader)
}

// checkVersionableIssues is the check over whatever reader answers, so a test can
// hand it a store of its own. It asks for everything in one unbounded read, as the
// switch does: every status, pinned rows, every type (a default listing hides gates,
// and a gate's timeout is the field that case is about) and both planes, in the brief
// projection with no labels and no counts. That is sound because none of what the
// projection leaves out can hold a number. A page is not a store, so there is no
// limit and no row cap.
func checkVersionableIssues(ctx context.Context, reader issueops.Reader) DoctorCheck {
	unlimited := 0
	page, err := reader.List(ctx, issueops.ListRequest{
		AllFlag:         true,
		IncludeAllTypes: true,
		Limit:           &unlimited,
		Brief:           true,
		SkipLabels:      true,
		SkipCounts:      true,
	})
	if err != nil {
		return versionableIssuesNotRun(err)
	}
	rows := make([]*types.Issue, 0, len(page.Items))
	for _, item := range page.Items {
		rows = append(rows, item.Issue)
	}

	found := storageissueops.FindUnversionable(rows)
	if len(found) == 0 {
		return DoctorCheck{
			Name:    versionableIssuesName,
			Status:  StatusOK,
			Message: "No issue holds a value a version could not record",
			Detail:  fmt.Sprintf("%s read; ephemeral and no-history rows are never versioned and are not checked", countedIssues(len(rows))),
		}
	}

	shown := found
	if len(shown) > versionableIssuesListLimit {
		shown = shown[:versionableIssuesListLimit]
	}
	entries := make([]string, 0, len(shown))
	for _, f := range shown {
		entries = append(entries, fmt.Sprintf("%s: %v", f.ID, f.Err))
	}
	detail := strings.Join(entries, "; ")
	if more := len(found) - len(shown); more > 0 {
		detail += fmt.Sprintf(" (+%d more)", more)
	}

	noun, verb := "issues", "hold"
	if len(found) == 1 {
		noun, verb = "issue", "holds"
	}
	return DoctorCheck{
		Name:    versionableIssuesName,
		Status:  StatusWarning,
		Message: fmt.Sprintf("%d %s %s a value a version could not record", len(found), noun, verb),
		Detail:  detail,
		Fix:     versionableIssuesRemedy(found),
	}
}

func countedIssues(n int) string {
	if n == 1 {
		return "1 issue"
	}
	return fmt.Sprintf("%d issues", n)
}

// versionableIssuesNotRun is the answer when the issues could not be read. It is a
// warning, not a pass.
func versionableIssuesNotRun(err error) DoctorCheck {
	return DoctorCheck{
		Name:    versionableIssuesName,
		Status:  StatusWarning,
		Message: "Unable to read issues to check them",
		Detail:  err.Error(),
		Fix:     "A check that could not run has not passed. Resolve the error above and run bd doctor again before turning versioned history on.",
	}
}

// versionableIssuesRemedy is the fix for each kind of field that is wrong, and only
// the kinds that are. The fix depends on the field. A number in metadata is replaced
// or removed with bd update, which works with history on or off, because a version is
// recorded from the state after the write. bd update cannot change a gate's timeout,
// so the gate is removed or its column is set with bd sql, while history is off; with
// history on, every write to that row, a close included, is refused if the row
// participates in history. bd sql needs a server-backed store, which the text says.
func versionableIssuesRemedy(found []storageissueops.UnversionableIssue) string {
	var inMetadata, inTimeout bool
	for _, f := range found {
		switch f.Field {
		case "metadata", "":
			inMetadata = true
		case "timeout":
			inTimeout = true
		}
	}

	lines := []string{
		"Fix each issue named above and run bd doctor again; there is no override. Once versioned history is on, a write to one of them that leaves the value in place would be refused.",
	}
	if inMetadata {
		lines = append(lines,
			"A number in metadata, one command per issue (this works with history on or off):",
			`  bd update <id> --metadata '{"<key>": "<value as a string>"}'   replace the value; a string keeps it exact`,
			"  bd update <id> --unset-metadata <key>                              or remove the key",
		)
	}
	if inTimeout {
		lines = append(lines,
			"A gate timeout is held as nanoseconds, and 9007199254740991 (about 104.25 days) is the most a version can record. bd update cannot change a gate's timeout. Remove the gate, or set its timeout within that limit while history is off:",
			"  bd delete <id> --force   remove the gate; without --force bd only previews, and removing a gate also unblocks anything it was blocking",
			`  bd sql "UPDATE issues SET timeout_ns = <nanoseconds> WHERE id = '<id>'"   set the timeout; bd sql needs a server-backed store, not an embedded one`,
			"With history on, every write to such a gate, a close included, is refused if the row participates in history.",
		)
	}
	return strings.Join(lines, "\n")
}
