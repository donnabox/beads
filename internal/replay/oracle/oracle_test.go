// The tests cover what the oracle read side must guarantee:
//
//   - a row that exists at the queried commit comes back as a plain
//     field-keyed map, ready for canonicalization without translation:
//     TestQueryAsOf_RowExists
//   - a row that did not exist yet, or was since deleted, comes back as nil:
//     TestQueryAsOf_PreCreation, TestQueryAsOf_PostDelete
//   - dolt sql -q has no bind-parameter API, so ref and issue id are
//     charset-validated before they reach the SQL string:
//     TestQueryAsOf_RejectsInvalidInput
//
// Every fixture is a throwaway database in a temporary directory, queried
// directly by the dolt CLI; nothing here touches a shared server. The
// environment deny-list and the served-directory guard are tested in doltcli.
package oracle

import (
	"context"
	"testing"

	"github.com/steveyegge/beads/internal/replay/replaytest"
)

// queryFixture holds the three commit hashes of a throwaway issues-table
// fixture: before the row is inserted, right after insert, and after delete.
type queryFixture struct {
	dir          string
	beforeInsert string
	afterInsert  string
	afterDelete  string
}

// newQueryFixture creates a throwaway Dolt database seeded with one issues
// row, then deletes it, committing after each phase so AS OF queries against
// the three returned hashes can observe the row not-yet-existing, existing,
// and deleted.
func newQueryFixture(t *testing.T) queryFixture {
	t.Helper()
	dir := replaytest.NewDoltDB(t, "fixturedb")
	replaytest.RunDolt(t, dir, "sql", "-q", "CREATE TABLE issues (id VARCHAR(64) PRIMARY KEY, title VARCHAR(255), status VARCHAR(32));")
	replaytest.RunDolt(t, dir, "add", "-A")
	replaytest.RunDolt(t, dir, "commit", "-m", "create issues table")
	beforeInsert := replaytest.HeadCommit(t, dir)

	replaytest.RunDolt(t, dir, "sql", "-q", "INSERT INTO issues VALUES ('bd-1', 'Test issue', 'open');")
	replaytest.RunDolt(t, dir, "add", "-A")
	replaytest.RunDolt(t, dir, "commit", "-m", "insert bd-1")
	afterInsert := replaytest.HeadCommit(t, dir)

	replaytest.RunDolt(t, dir, "sql", "-q", "DELETE FROM issues WHERE id = 'bd-1';")
	replaytest.RunDolt(t, dir, "add", "-A")
	replaytest.RunDolt(t, dir, "commit", "-m", "delete bd-1")
	afterDelete := replaytest.HeadCommit(t, dir)

	return queryFixture{dir: dir, beforeInsert: beforeInsert, afterInsert: afterInsert, afterDelete: afterDelete}
}

func TestQueryAsOf_RowExists(t *testing.T) {
	fx := newQueryFixture(t)
	got, err := QueryAsOf(context.Background(), fx.dir, fx.afterInsert, "bd-1")
	if err != nil {
		t.Fatalf("QueryAsOf: %v", err)
	}
	if got == nil {
		t.Fatalf("QueryAsOf returned nil row, want a populated map")
	}
	want := map[string]string{"id": "bd-1", "title": "Test issue", "status": "open"}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("field %q = %q, want %q (full row: %v)", k, got[k], v, got)
		}
	}
}

func TestQueryAsOf_PreCreation(t *testing.T) {
	fx := newQueryFixture(t)
	got, err := QueryAsOf(context.Background(), fx.dir, fx.beforeInsert, "bd-1")
	if err != nil {
		t.Fatalf("QueryAsOf: %v", err)
	}
	if got != nil {
		t.Fatalf("QueryAsOf returned %v, want nil (row not created yet)", got)
	}
}

func TestQueryAsOf_PostDelete(t *testing.T) {
	fx := newQueryFixture(t)
	got, err := QueryAsOf(context.Background(), fx.dir, fx.afterDelete, "bd-1")
	if err != nil {
		t.Fatalf("QueryAsOf: %v", err)
	}
	if got != nil {
		t.Fatalf("QueryAsOf returned %v, want nil (row deleted)", got)
	}
}

func TestQueryAsOf_RejectsInvalidInput(t *testing.T) {
	fx := newQueryFixture(t)
	if _, err := QueryAsOf(context.Background(), fx.dir, "not a ref; DROP TABLE issues", "bd-1"); err == nil {
		t.Fatal("QueryAsOf accepted an unsafe ref, want an error")
	}
	if _, err := QueryAsOf(context.Background(), fx.dir, fx.afterInsert, "bd-1'; DROP TABLE issues; --"); err == nil {
		t.Fatal("QueryAsOf accepted an unsafe issue id, want an error")
	}
}
