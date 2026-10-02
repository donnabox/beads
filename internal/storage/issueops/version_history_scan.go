package issueops

import "github.com/steveyegge/beads/internal/types"

// UnversionableIssue names one issue whose metadata RecordVersionInTx would
// refuse to version, and the refusal it would return.
type UnversionableIssue struct {
	ID  string
	Err error
}

// FindUnversionableMetadata runs CheckMetadataVersionable over every issue in
// issues that the mint would version and returns the ones it refuses, in the
// order they were given. It is the core of the check made when versioned
// history is switched on: with history on, a write that introduces such a value
// aborts in its own transaction, but a row that already holds one -- written
// while history was off, or by a path that does not mint -- would fail every
// later write to it.
//
// It skips the rows the mint never versions by IsWisp: ephemeral and no-history
// rows are never versioned, so a value they hold can never abort a write. It
// cannot skip legacy rows, because participation is not on types.Issue, so on a
// store with the write fence (design §16.2b) it is a deliberate
// over-approximation: the mint skips a legacy row whatever it holds, and this
// still refuses over it, which is the safe direction because it can only refuse
// more. It does no I/O; the caller loads the rows, which is what lets one
// function serve the direct store and the proxied route.
func FindUnversionableMetadata(issues []*types.Issue) []UnversionableIssue {
	var found []UnversionableIssue
	for _, issue := range issues {
		if issue == nil || IsWisp(issue) {
			continue
		}
		if err := CheckMetadataVersionable(issue.Metadata); err != nil {
			found = append(found, UnversionableIssue{ID: issue.ID, Err: err})
		}
	}
	return found
}
