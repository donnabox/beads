package issueops

import (
	"database/sql"
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/steveyegge/beads/internal/types"
	publicops "github.com/steveyegge/beads/issueops"
)

// The current record deliberately disagrees with the mocked label rows. This
// exercises ApplyLabelPatch's rows-affected result after its set-difference
// fast path, without claiming that public Update uses a stale hydrated record.
func TestApplyLabelPatchReportsActualRowsAffected(t *testing.T) {
	const id = "bd-label-patch-result"
	tests := []struct {
		name       string
		removeRows int64 // -1 omits the removal from the patch
		addRows    int64 // -1 omits the addition from the patch
	}{
		{name: "stale add affects zero", removeRows: -1, addRows: 0},
		{name: "stale remove affects zero", removeRows: 0, addRows: -1},
		{name: "both affect zero", removeRows: 0, addRows: 0},
		{name: "changed removal survives no-op addition", removeRows: 1, addRows: 0},
		{name: "changed addition after no-op removal", removeRows: 0, addRows: 1},
		{name: "both change", removeRows: 1, addRows: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = db.Close() })
			mock.ExpectBegin()
			tx, err := db.Begin()
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = tx.Rollback() })

			current := &types.Issue{ID: id}
			patch := publicops.LabelPatch{}
			wantChanged := tt.removeRows > 0 || tt.addRows > 0
			expectMutation := func(label string, remove bool, affected int64) {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM wisps WHERE id = ? LIMIT 1")).
					WithArgs(id).WillReturnError(sql.ErrNoRows)
				query := "INSERT IGNORE INTO labels (issue_id, label) VALUES (?, ?)"
				if remove {
					query = "DELETE FROM labels WHERE issue_id = ? AND label = ?"
				}
				mock.ExpectExec(regexp.QuoteMeta(query)).WithArgs(id, label).
					WillReturnResult(sqlmock.NewResult(0, affected))
				if affected == 0 {
					if !remove {
						mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM issues WHERE id = ?")).
							WithArgs(id).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
					}
					return
				}
				mock.ExpectQuery(`(?s)SELECT id FROM events\s+WHERE issue_id =`).
					WillReturnRows(sqlmock.NewRows([]string{"id"}))
				mock.ExpectExec(regexp.QuoteMeta("INSERT INTO events (id, issue_id, event_type, actor, old_value, new_value, comment, created_at)")).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec(regexp.QuoteMeta("UPDATE issues SET updated_at = ?, row_lock = ? WHERE id = ?")).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), id).
					WillReturnResult(sqlmock.NewResult(0, 1))
			}
			if tt.removeRows >= 0 {
				current.Labels = []string{"old"}
				patch.Remove = []string{"old"}
				expectMutation("old", true, tt.removeRows)
			}
			if tt.addRows >= 0 {
				patch.Add = []string{"new"}
				expectMutation("new", false, tt.addRows)
			}
			mock.ExpectRollback()

			// Enable journaling for the all-no-op cases: any attempted snapshot,
			// event, or Issue touch would be unexpected SQL and fail the call.
			// Changed cases isolate result accumulation; real-engine tests own
			// journal and committed-state qualification.
			ctx := WithEventsJournal(t.Context(), !wantChanged)
			changed, err := ApplyLabelPatch(ctx, tx, current, patch, "tester")
			if err != nil {
				t.Fatalf("ApplyLabelPatch: %v", err)
			}
			if changed != wantChanged {
				t.Fatalf("changed = %v, want %v from actual affected rows", changed, wantChanged)
			}
			if err := tx.Rollback(); err != nil {
				t.Fatalf("rollback: %v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("SQL expectations: %v", err)
			}
		})
	}
}
