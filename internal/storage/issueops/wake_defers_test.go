package issueops

import (
	"slices"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// With versioned history off the mint is a no-op, so nothing can refuse a row and the sweep has no
// reason to look at one before it wakes it. When history is on, the sweep checks each expired row
// against the mint before it writes anything (see the embedded-store tests for what that decides);
// that check must cost nothing when history is off.
//
// The mock scripts exactly the statements a one-row sweep issues: the expired-row SELECT, the
// UPDATE, the event's same-content SELECT and its INSERT, and the empty wisps pass. Nothing else is
// expected, so a sweep that loaded the row (or read its participation) before waking it would
// arrive at the mock as an unexpected statement and fail the sweep instead of passing it.
func TestWakeExpiredDefersWithHistoryOffLoadsNothingBeforeWaking(t *testing.T) {
	_, mock, tx := beginMockTx(t)

	mock.ExpectQuery(`SELECT id FROM issues WHERE status = 'deferred'`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("bd-1"))
	mock.ExpectExec(`UPDATE issues SET status = 'open', defer_until = NULL`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "bd-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT id FROM events WHERE issue_id`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectExec(`INSERT INTO events`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT id FROM wisps WHERE status = 'deferred'`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	// tx is not scoped for versioned history, which is how every transaction reads while it is off.
	got, err := WakeExpiredDefersInTx(t.Context(), tx)
	if err != nil {
		t.Fatalf("WakeExpiredDefersInTx = %v, want success with only the scripted statements", err)
	}
	if !slices.Equal(got.Issues, []string{"bd-1"}) || len(got.Wisps) != 0 {
		t.Errorf("woken = issues %v, wisps %v, want [bd-1] and none", got.Issues, got.Wisps)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the sweep did not issue the scripted statements: %v", err)
	}
}
