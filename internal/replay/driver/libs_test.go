package driver_test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/steveyegge/beads/internal/replay/compare"
	"github.com/steveyegge/beads/internal/replay/driver"
	"github.com/steveyegge/beads/internal/replay/oracle"
	"github.com/steveyegge/beads/internal/replay/replaytest"
)

// seededDB returns a dolt database whose issues table holds one row with the
// given title, and the commit that holds it.
func seededDB(t *testing.T, name, title string) (dir, head string) {
	t.Helper()
	dir = replaytest.NewDoltDB(t, name)
	replaytest.RunDolt(t, dir, "sql", "-q", "CREATE TABLE issues (id VARCHAR(64) PRIMARY KEY, title VARCHAR(255), status VARCHAR(32))")
	replaytest.RunDolt(t, dir, "sql", "-q", "INSERT INTO issues VALUES ('x-1', '"+title+"', 'open')")
	replaytest.RunDolt(t, dir, "add", "-A")
	replaytest.RunDolt(t, dir, "commit", "-m", "seed")
	return dir, replaytest.HeadCommit(t, dir)
}

// B1.DriverUsesLibs: the driver reads rows through oracle.QueryAsOf and
// compares them with compare.Compare, in process. The result it returns is the
// library's own type, not a renamed copy, and equals what a direct call to the
// library produces for the same two rows.
func TestB1DriverUsesLibs(t *testing.T) {
	replaytest.Require(t, replaytest.NeedDolt)
	ctx := context.Background()

	oracleDir, oracleHead := seededDB(t, "oracle", "Widget")
	sameDir, sameHead := seededDB(t, "same", "Widget")
	otherDir, otherHead := seededDB(t, "other", "Gadget")

	read := func(dir, head string) func(context.Context) (map[string]string, error) {
		return func(ctx context.Context) (map[string]string, error) {
			return oracle.QueryAsOf(ctx, dir, head, "x-1")
		}
	}

	cases := []struct {
		name      string
		dir, head string
		wantMatch bool
	}{
		{"identical rows match", sameDir, sameHead, true},
		{"different title mismatches", otherDir, otherHead, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := driver.ReplayAndCompare(ctx, read(oracleDir, oracleHead), read(tc.dir, tc.head))
			if err != nil {
				t.Fatalf("ReplayAndCompare: %v", err)
			}
			var _ compare.Result = got // the driver returns the library's type

			oracleRow, err := oracle.QueryAsOf(ctx, oracleDir, oracleHead, "x-1")
			if err != nil {
				t.Fatalf("oracle.QueryAsOf (oracle): %v", err)
			}
			otherRow, err := oracle.QueryAsOf(ctx, tc.dir, tc.head, "x-1")
			if err != nil {
				t.Fatalf("oracle.QueryAsOf (candidate): %v", err)
			}
			// The rows hold only id, title and status, none of which the driver
			// treats as incidental, so they go to the comparator unchanged.
			oracleJSON, _ := json.Marshal(oracleRow)
			otherJSON, _ := json.Marshal(otherRow)
			want, err := compare.Compare(oracleJSON, otherJSON)
			if err != nil {
				t.Fatalf("compare.Compare: %v", err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("driver result differs from the library's:\n got  %+v\n want %+v", got, want)
			}
			if got.Matched != tc.wantMatch {
				t.Errorf("Matched = %v, want %v", got.Matched, tc.wantMatch)
			}
		})
	}
}
