package oracle

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/steveyegge/beads/internal/replay/doltcli"
	"github.com/steveyegge/beads/internal/replay/replaytest"
)

// b2Fixture is a throwaway Dolt database shaped like the part of a bd store
// the view reads, with one commit per phase so a read can be taken before the
// issues table exists, with issues but no dependencies table, and with both.
type b2Fixture struct {
	dir        string
	root       string // the commit dolt init made: no tables yet
	issuesOnly string // issues and their rows exist, the dependencies table does not
	full       string // both tables, with rows
}

func b2Commit(t *testing.T, dir, msg string) string {
	t.Helper()
	replaytest.RunDolt(t, dir, "add", "-A")
	replaytest.RunDolt(t, dir, "commit", "-m", msg)
	return replaytest.HeadCommit(t, dir)
}

func newB2Fixture(t *testing.T) b2Fixture {
	t.Helper()
	dir := replaytest.NewDoltDB(t, "viewdb")
	fx := b2Fixture{dir: dir, root: replaytest.HeadCommit(t, dir)}

	replaytest.RunDolt(t, dir, "sql", "-q",
		"CREATE TABLE issues (id VARCHAR(255) PRIMARY KEY, title VARCHAR(500) NOT NULL, notes LONGTEXT, metadata JSON)")
	replaytest.RunDolt(t, dir, "sql", "-q", `INSERT INTO issues (id, title, notes, metadata) VALUES
		('x-1', 'One', NULL, '{}'),
		('x-2', 'Two', '', NULL),
		('x-3', 'Three', 'note', '{"a":1}')`)
	fx.issuesOnly = b2Commit(t, dir, "issues and rows")

	replaytest.RunDolt(t, dir, "sql", "-q", `CREATE TABLE dependencies (
		id CHAR(36) PRIMARY KEY, issue_id VARCHAR(255) NOT NULL, type VARCHAR(32) NOT NULL, metadata JSON,
		depends_on_issue_id VARCHAR(255), depends_on_wisp_id VARCHAR(255), depends_on_external VARCHAR(255))`)
	// x-1 owns four edges, one to an external target; x-2 owns one edge that
	// points at x-1, which x-1's view must not carry.
	replaytest.RunDolt(t, dir, "sql", "-q", `INSERT INTO dependencies
		(id, issue_id, type, metadata, depends_on_issue_id, depends_on_wisp_id, depends_on_external) VALUES
		('d1', 'x-1', 'blocks',       NULL, 'x-2', NULL, NULL),
		('d2', 'x-1', 'parent-child', '{}', 'x-3', NULL, NULL),
		('d3', 'x-1', 'blocks',       NULL, 'x-0', NULL, NULL),
		('d4', 'x-2', 'blocks',       NULL, 'x-1', NULL, NULL),
		('d5', 'x-1', 'related',      NULL, NULL,  NULL, 'ext:thing')`)
	fx.full = b2Commit(t, dir, "dependencies and rows")
	return fx
}

// b2FlattenedRef returns a database and a commit hash that was real once and
// has since been garbage collected, as history flattening leaves it.
func b2FlattenedRef(t *testing.T) (dir, ref string) {
	t.Helper()
	dir = replaytest.NewDoltDB(t, "flatdb")
	root := replaytest.HeadCommit(t, dir)
	replaytest.RunDolt(t, dir, "sql", "-q", "CREATE TABLE issues (id VARCHAR(255) PRIMARY KEY, title VARCHAR(500) NOT NULL)")
	ref = b2Commit(t, dir, "a commit that will not survive")
	replaytest.RunDolt(t, dir, "reset", "--hard", root)
	replaytest.RunDolt(t, dir, "gc")
	return dir, ref
}

func b2IDs(t *testing.T, rows []Row) []string {
	t.Helper()
	var ids []string
	for _, r := range rows {
		c, ok := r.Cell("id")
		if !ok || c.Null {
			t.Fatalf("a row without an id: %#v", r)
		}
		ids = append(ids, c.Text)
	}
	return ids
}

// B2.RefContract: what the read side promises about a ref. One that predates
// the table reads as "no such row", nil and no error; one that cannot be
// resolved at all is a typed error a caller can test for, so a scenario that
// expects exactly that failure cannot pass on an unrelated one.
func TestB2RefContract(t *testing.T) {
	ctx := context.Background()
	fx := newB2Fixture(t)
	flatDir, flatRef := b2FlattenedRef(t)

	t.Run("a ref older than the issues table is nil and not an error", func(t *testing.T) {
		row, err := QueryAsOf(ctx, fx.dir, fx.root, "x-1")
		if err != nil || row != nil {
			t.Errorf("QueryAsOf at a ref before the table = (%v, %v), want (nil, nil)", row, err)
		}
		view, err := ReadView(ctx, fx.dir, fx.root, "x-1")
		if err != nil || view != nil {
			t.Errorf("ReadView at a ref before the table = (%v, %v), want (nil, nil)", view, err)
		}
		cols, err := Columns(ctx, fx.dir, fx.root, "issues")
		if err != nil || cols != nil {
			t.Errorf("Columns at a ref before the table = (%v, %v), want (nil, nil)", cols, err)
		}
	})

	t.Run("a ref that cannot be resolved satisfies ErrRefUnresolvable", func(t *testing.T) {
		for _, c := range []struct{ name, dir, ref string }{
			{"an unknown branch", fx.dir, "no-such-branch"},
			{"a well-formed hash that is not in the database", fx.dir, "0123456789abcdefghijklmnopqrstuv"},
			{"a commit that history flattening removed", flatDir, flatRef},
		} {
			t.Run(c.name, func(t *testing.T) {
				if _, err := QueryAsOf(ctx, c.dir, c.ref, "x-1"); !errors.Is(err, ErrRefUnresolvable) {
					t.Errorf("QueryAsOf error = %v, want errors.Is ErrRefUnresolvable", err)
				}
				if _, err := ReadView(ctx, c.dir, c.ref, "x-1"); !errors.Is(err, ErrRefUnresolvable) {
					t.Errorf("ReadView error = %v, want errors.Is ErrRefUnresolvable", err)
				}
				if _, err := Columns(ctx, c.dir, c.ref, "issues"); !errors.Is(err, ErrRefUnresolvable) {
					t.Errorf("Columns error = %v, want errors.Is ErrRefUnresolvable", err)
				}
			})
		}
	})

	t.Run("an unrelated failure is neither nil nor ErrRefUnresolvable", func(t *testing.T) {
		notADatabase := t.TempDir()
		if view, err := ReadView(ctx, notADatabase, "main", "x-1"); err == nil || errors.Is(err, ErrRefUnresolvable) || view != nil {
			t.Errorf("ReadView in a directory that is not a dolt database = (%v, %v), want a plain error", view, err)
		}
		if row, err := QueryAsOf(ctx, notADatabase, "main", "x-1"); err == nil || errors.Is(err, ErrRefUnresolvable) || row != nil {
			t.Errorf("QueryAsOf in a directory that is not a dolt database = (%v, %v), want a plain error", row, err)
		}
		if _, err := ReadView(ctx, fx.dir, "not a ref; DROP TABLE issues", "x-1"); err == nil || errors.Is(err, ErrRefUnresolvable) {
			t.Errorf("ReadView with an unsafe ref: err = %v, want a validation error that is not ErrRefUnresolvable", err)
		}
		if _, err := Columns(ctx, fx.dir, fx.full, "issues; DROP TABLE issues"); err == nil || errors.Is(err, ErrRefUnresolvable) {
			t.Errorf("Columns with an unsafe table name: err = %v, want a validation error", err)
		}
	})

	t.Run("a missing issue at a good ref is nil and not an error", func(t *testing.T) {
		view, err := ReadView(ctx, fx.dir, fx.full, "x-nope")
		if err != nil || view != nil {
			t.Errorf("ReadView of an issue that does not exist = (%v, %v), want (nil, nil)", view, err)
		}
	})
}

// B2.ViewOwnedEdges: the view carries the issue row and the edges the issue
// owns (its issue_id end), in a fixed order; an edge that merely points at the
// issue belongs to its owner's view.
func TestB2ViewOwnedEdges(t *testing.T) {
	ctx := context.Background()
	fx := newB2Fixture(t)

	view, err := ReadView(ctx, fx.dir, fx.full, "x-1")
	if err != nil || view == nil {
		t.Fatalf("ReadView x-1 = (%v, %v)", view, err)
	}
	if want := []string{"id", "title", "notes", "metadata"}; !reflect.DeepEqual(view.Issue.Columns, want) {
		t.Errorf("Issue.Columns = %v, want the schema order %v", view.Issue.Columns, want)
	}
	if c, _ := view.Issue.Cell("id"); c.Text != "x-1" {
		t.Errorf("Issue id = %#v, want x-1", c)
	}
	// Ordered by target issue then type, with the edge that has no issue target
	// (NULL sorts first) ahead of the rest.
	if got, want := b2IDs(t, view.Dependencies), []string{"d5", "d3", "d1", "d2"}; !reflect.DeepEqual(got, want) {
		t.Errorf("x-1 dependency ids = %v, want %v (owned edges only, sorted)", got, want)
	}
	for _, edge := range view.Dependencies {
		if want := []string{"id", "issue_id", "type", "metadata", "depends_on_issue_id", "depends_on_wisp_id", "depends_on_external"}; !reflect.DeepEqual(edge.Columns, want) {
			t.Errorf("edge Columns = %v, want %v", edge.Columns, want)
		}
	}

	other, err := ReadView(ctx, fx.dir, fx.full, "x-2")
	if err != nil || other == nil {
		t.Fatalf("ReadView x-2 = (%v, %v)", other, err)
	}
	if got, want := b2IDs(t, other.Dependencies), []string{"d4"}; !reflect.DeepEqual(got, want) {
		t.Errorf("x-2 dependency ids = %v, want %v", got, want)
	}

	leaf, err := ReadView(ctx, fx.dir, fx.full, "x-3")
	if err != nil || leaf == nil {
		t.Fatalf("ReadView x-3 = (%v, %v)", leaf, err)
	}
	if len(leaf.Dependencies) != 0 {
		t.Errorf("x-3 owns no edges, got %v", b2IDs(t, leaf.Dependencies))
	}
}

// B2.RowKeepsNullVsEmpty: a row keeps the difference between SQL NULL and the
// empty string, for text and JSON columns alike, and says which columns it was
// read under.
func TestB2RowKeepsNullVsEmpty(t *testing.T) {
	ctx := context.Background()
	fx := newB2Fixture(t)
	get := func(id string) Row {
		t.Helper()
		v, err := ReadView(ctx, fx.dir, fx.full, id)
		if err != nil || v == nil {
			t.Fatalf("ReadView %s = (%v, %v)", id, v, err)
		}
		return v.Issue
	}
	cases := []struct {
		id, col string
		want    doltcli.Cell
	}{
		{"x-1", "notes", doltcli.Cell{Null: true}},
		{"x-2", "notes", doltcli.Cell{Text: ""}},
		{"x-3", "notes", doltcli.Cell{Text: "note"}},
		{"x-1", "metadata", doltcli.Cell{Text: "{}"}},
		{"x-2", "metadata", doltcli.Cell{Null: true}},
	}
	for _, c := range cases {
		got, ok := get(c.id).Cell(c.col)
		if !ok || got != c.want {
			t.Errorf("%s.%s = (%#v, %v), want %#v", c.id, c.col, got, ok, c.want)
		}
	}
	if _, ok := get("x-1").Cell("no_such_column"); ok {
		t.Error("Cell reported a column the row was not read with")
	}
}

// B2.ViewTableAbsent: a table the ref does not have yet contributes an empty
// section, not an error; only the issue row decides whether the view exists.
func TestB2ViewTableAbsent(t *testing.T) {
	ctx := context.Background()
	fx := newB2Fixture(t)
	view, err := ReadView(ctx, fx.dir, fx.issuesOnly, "x-1")
	if err != nil || view == nil {
		t.Fatalf("ReadView before the dependencies table existed = (%v, %v), want the issue", view, err)
	}
	if len(view.Dependencies) != 0 {
		t.Errorf("Dependencies = %v, want none while the table does not exist", b2IDs(t, view.Dependencies))
	}
}

// B2.ColumnsAtRef: the schema at a ref is the column list of the table as it
// stood there, in schema order.
func TestB2ColumnsAtRef(t *testing.T) {
	ctx := context.Background()
	fx := newB2Fixture(t)
	cols, err := Columns(ctx, fx.dir, fx.full, "issues")
	if err != nil {
		t.Fatalf("Columns issues: %v", err)
	}
	if want := []string{"id", "title", "notes", "metadata"}; !reflect.DeepEqual(cols, want) {
		t.Errorf("issues columns = %v, want %v", cols, want)
	}
	cols, err = Columns(ctx, fx.dir, fx.full, "dependencies")
	if err != nil {
		t.Fatalf("Columns dependencies: %v", err)
	}
	if want := []string{"id", "issue_id", "type", "metadata", "depends_on_issue_id", "depends_on_wisp_id", "depends_on_external"}; !reflect.DeepEqual(cols, want) {
		t.Errorf("dependencies columns = %v, want %v", cols, want)
	}
	if cols, err := Columns(ctx, fx.dir, fx.issuesOnly, "dependencies"); err != nil || cols != nil {
		t.Errorf("dependencies columns before the table existed = (%v, %v), want (nil, nil)", cols, err)
	}
}
