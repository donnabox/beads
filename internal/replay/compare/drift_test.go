package compare

import (
	"context"
	"sort"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/replay/doltcli"
	"github.com/steveyegge/beads/internal/replay/replaytest"
)

// b2SchemaColumn is one column of a replayed table as a freshly initialized
// store has it.
type b2SchemaColumn struct {
	Name     string
	DataType string
}

// b2StoreSchema initializes a bd store from this tree and reads the columns of
// the replayed tables back from information_schema, in schema order. The store
// is the ground truth: a migration that adds a column shows up here, and the
// drift tests below turn it into a failure until someone decides how the
// column is compared.
func b2StoreSchema(t *testing.T) map[string][]b2SchemaColumn {
	t.Helper()
	replaytest.Require(t, replaytest.NeedDolt|replaytest.NeedBd)
	replaytest.Isolate(t)
	project := replaytest.InitBdProject(t, "schema", replaytest.BdBin(t))
	data := replaytest.DataDir(t, project)

	_, rows, err := doltcli.Query(context.Background(), data,
		"SELECT table_name, column_name, data_type FROM information_schema.columns "+
			"WHERE table_schema = DATABASE() AND table_name IN ('issues', 'dependencies') "+
			"ORDER BY table_name, ordinal_position")
	if err != nil {
		t.Fatalf("reading the store schema: %v", err)
	}
	out := map[string][]b2SchemaColumn{}
	for _, r := range rows {
		table := r[0].Text
		out[table] = append(out[table], b2SchemaColumn{Name: r[1].Text, DataType: strings.ToLower(r[2].Text)})
	}
	for _, table := range []string{"issues", "dependencies"} {
		if len(out[table]) == 0 {
			t.Fatalf("the fresh store has no columns for %s: %v", table, out)
		}
	}
	return out
}

func b2Set(names []string) map[string]bool {
	s := map[string]bool{}
	for _, n := range names {
		s[n] = true
	}
	return s
}

func b2Sorted(s map[string]bool) []string {
	out := make([]string, 0, len(s))
	for n := range s {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

// B2.JSONColumnDrift: the columns the comparator parses as JSON documents are
// exactly the JSON-typed columns of the replayed tables. A JSON column missing
// from the static set would be compared as opaque text, sensitive to how each
// Dolt release spells its numbers; one listed but not JSON would be parsed as
// something it is not.
func TestB2JSONColumnDrift(t *testing.T) {
	schema := b2StoreSchema(t)
	for _, table := range []string{"issues", "dependencies"} {
		fromSchema := map[string]bool{}
		for _, c := range schema[table] {
			if c.DataType == "json" {
				fromSchema[c.Name] = true
			}
		}
		static := b2Set(jsonColumns[table])
		for _, name := range b2Sorted(fromSchema) {
			if !static[name] {
				t.Errorf("%s.%s is a JSON column in the store but is not in the comparator's static JSON set", table, name)
			}
		}
		for _, name := range b2Sorted(static) {
			if !fromSchema[name] {
				t.Errorf("%s.%s is in the comparator's static JSON set but is not a JSON column in the store", table, name)
			}
		}
	}
}

// B2.ColumnDrift: every column of every replayed table is either compared or a
// stamp, and never both. A migration that adds a column fails here until
// someone decides which, so a column can never be silently left out of the
// comparison.
func TestB2ColumnDrift(t *testing.T) {
	schema := b2StoreSchema(t)
	for _, table := range []string{"issues", "dependencies"} {
		compared := b2Set(comparedColumns[table])
		stamps := stampColumns[table]
		inSchema := map[string]bool{}
		for _, c := range schema[table] {
			inSchema[c.Name] = true
			_, isStamp := stamps[c.Name]
			switch {
			case compared[c.Name] && isStamp:
				t.Errorf("%s.%s is listed as both compared and a stamp", table, c.Name)
			case !compared[c.Name] && !isStamp:
				t.Errorf("%s.%s is in the store but is neither compared nor a stamp: decide which and add it to the lists", table, c.Name)
			}
		}
		for _, name := range b2Sorted(compared) {
			if !inSchema[name] {
				t.Errorf("%s.%s is listed as compared but is not a column of the store", table, name)
			}
		}
		for name, reason := range stamps {
			if !inSchema[name] {
				t.Errorf("%s.%s is listed as a stamp but is not a column of the store", table, name)
			}
			if strings.TrimSpace(reason) == "" {
				t.Errorf("stamp %s.%s has no reason", table, name)
			}
		}
		if len(stamps) == 0 {
			t.Errorf("%s has no stamp columns: every replayed table carries at least a write time", table)
		}
	}
}
