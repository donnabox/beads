package compare

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"unicode/utf8"

	"github.com/steveyegge/beads/internal/replay/oracle"
)

// ColumnSkew lists the columns that exist on one side of a comparison only,
// because the two databases were written under different schemas.
type ColumnSkew struct {
	OracleOnly    []string `json:"oracle_only,omitempty"`
	CandidateOnly []string `json:"candidate_only,omitempty"`
}

// Skew is the schema skew a comparison saw, by table. A table whose schemas
// agree is absent. It is recorded once per run and is never a mismatch.
type Skew map[string]ColumnSkew

// Payload renders a view as the JSON object Compare hashes: each compared column
// of the issue as null (SQL NULL) or its exact text, the JSON-typed columns as
// the documents they hold, and the dependencies section as an array of edges
// built the same way. Stamp columns are left out. A nil view is the empty object.
//
// A JSON column that holds SQL NULL is the one exception to "null for SQL NULL":
// its member is left out. The same column can also hold the document null, a
// different stored value that is written as null, and leaving the member out keeps
// the two apart without a marker value a document could also contain.
//
// Text is never passed through a number, so integers, decimals and timestamps
// keep their exact spelling. Invalid UTF-8 is an error here rather than a quiet
// U+FFFD, because two different byte strings must not become equal.
func Payload(v *oracle.View) ([]byte, error) {
	return payload(v, nil, nil)
}

// CompareViews compares two views as Compare compares two payloads, on the
// columns that both sides' schemas have and that are not stamps. A column on one
// side only is returned in the skew and never decides the result. A view that is
// nil stands for an issue with no row at that ref: two missing issues match, and
// one missing and one present do not. That mismatch is named for the side that
// lacks the row, CategoryIssueMissing when the candidate does and
// CategoryIssueExtra when the oracle does. Whether a row exists does not depend
// on any number in it, so that category outranks the number gate: the pair is a
// proven mismatch and never uncomparable.
func CompareViews(oracleView, candidateView *oracle.View) (Result, Skew, error) {
	skew := Skew{}
	var issueShared, edgeShared map[string]bool
	if oracleView != nil && candidateView != nil {
		var s ColumnSkew
		issueShared, s = intersect(oracleView.Issue.Columns, candidateView.Issue.Columns)
		if len(s.OracleOnly)+len(s.CandidateOnly) > 0 {
			skew[tableIssues] = s
		}
		// The columns of a section are known only from its rows, so a section
		// that is empty on either side has nothing to intersect.
		if len(oracleView.Dependencies) > 0 && len(candidateView.Dependencies) > 0 {
			edgeShared, s = intersect(oracleView.Dependencies[0].Columns, candidateView.Dependencies[0].Columns)
			if len(s.OracleOnly)+len(s.CandidateOnly) > 0 {
				skew[tableDependencies] = s
			}
		}
	}
	oracleJSON, err := payload(oracleView, issueShared, edgeShared)
	if err != nil {
		return Result{}, nil, fmt.Errorf("oracle view: %w", err)
	}
	candidateJSON, err := payload(candidateView, issueShared, edgeShared)
	if err != nil {
		return Result{}, nil, fmt.Errorf("candidate view: %w", err)
	}
	result, err := Compare(oracleJSON, candidateJSON)
	if err != nil {
		return Result{}, nil, err
	}
	// A missing row renders as the empty object, which the classifier cannot tell
	// from any other difference, so the category is decided here, from which view
	// is nil. It replaces whatever Compare chose, the number gate's included.
	if result.Mismatch != nil && (oracleView == nil) != (candidateView == nil) {
		if candidateView == nil {
			result.Mismatch.Category = CategoryIssueMissing
		} else {
			result.Mismatch.Category = CategoryIssueExtra
		}
	}
	return result, skew, nil
}

// intersect returns the columns both lists have and, in sorted order, the ones
// only one of them has. Empty results are nil.
func intersect(oracleCols, candidateCols []string) (map[string]bool, ColumnSkew) {
	inOracle, inCandidate := map[string]bool{}, map[string]bool{}
	for _, c := range oracleCols {
		inOracle[c] = true
	}
	for _, c := range candidateCols {
		inCandidate[c] = true
	}
	shared := map[string]bool{}
	var skew ColumnSkew
	for _, c := range oracleCols {
		if inCandidate[c] {
			shared[c] = true
		} else {
			skew.OracleOnly = append(skew.OracleOnly, c)
		}
	}
	for _, c := range candidateCols {
		if !inOracle[c] {
			skew.CandidateOnly = append(skew.CandidateOnly, c)
		}
	}
	sort.Strings(skew.OracleOnly)
	sort.Strings(skew.CandidateOnly)
	return shared, skew
}

// payload renders v keeping only the columns shared allows (nil allows all).
func payload(v *oracle.View, issueShared, edgeShared map[string]bool) ([]byte, error) {
	if v == nil {
		return []byte("{}"), nil
	}
	var buf bytes.Buffer
	buf.WriteByte('{')
	n, err := writeMembers(&buf, v.Issue, tableIssues, issueShared)
	if err != nil {
		return nil, err
	}
	if n > 0 {
		buf.WriteByte(',')
	}
	buf.WriteString(`"dependencies":[`)
	for i, edge := range v.Dependencies {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.WriteByte('{')
		if _, err := writeMembers(&buf, edge, tableDependencies, edgeShared); err != nil {
			return nil, err
		}
		buf.WriteByte('}')
	}
	buf.WriteString("]}")
	return buf.Bytes(), nil
}

// writeMembers writes the members of one row's JSON object, in column order,
// and returns how many it wrote.
//
// A SQL NULL is written as null, except in a JSON-typed column, where the member
// is left out. That column can hold SQL NULL or the document null, they are
// different stored values, and writing both as null would call them equal.
// Leaving the member out is the one rendering no document can also produce, so it
// needs no marker value. The count is of members written, so the separators stay
// right wherever in the row the omitted column sits.
func writeMembers(buf *bytes.Buffer, row oracle.Row, table string, shared map[string]bool) (int, error) {
	if len(row.Cells) < len(row.Columns) {
		return 0, fmt.Errorf("%s row has %d cells for %d columns", table, len(row.Cells), len(row.Columns))
	}
	n := 0
	for i, col := range row.Columns {
		if isStamp(table, col) || (shared != nil && !shared[col]) {
			continue
		}
		cell := row.Cells[i]
		if cell.Null && isJSONColumn(table, col) {
			continue
		}
		if n > 0 {
			buf.WriteByte(',')
		}
		n++
		buf.Write(jsonString(col))
		buf.WriteByte(':')
		switch {
		case cell.Null:
			buf.WriteString("null")
		case !utf8.ValidString(cell.Text):
			return 0, fmt.Errorf("%s.%s is not valid UTF-8", table, col)
		case isJSONColumn(table, col):
			if !json.Valid([]byte(cell.Text)) {
				return 0, fmt.Errorf("%s.%s is not valid JSON: %q", table, col, cell.Text)
			}
			buf.WriteString(cell.Text)
		default:
			buf.Write(jsonString(cell.Text))
		}
	}
	return n, nil
}

// jsonString encodes s as a JSON string without the HTML escaping encoding/json
// applies by default, so the payload reads as the text it carries. s must be
// valid UTF-8, which every caller has checked.
func jsonString(s string) []byte {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(s) // a valid string always encodes
	return bytes.TrimSuffix(b.Bytes(), []byte("\n"))
}
