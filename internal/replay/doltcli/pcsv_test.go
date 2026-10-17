package doltcli_test

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/replay/doltcli"
	"github.com/steveyegge/beads/internal/replay/replaytest"
)

// sqlString renders s as a MySQL string literal.
func sqlString(s string) string {
	esc := strings.NewReplacer(`\`, `\\`, `'`, `\'`, "\n", `\n`, "\r", `\r`, "\t", `\t`, "\x00", `\0`)
	return "'" + esc.Replace(s) + "'"
}

// sqlNullable renders a nullable string as a literal or NULL.
func sqlNullable(s *string) string {
	if s == nil {
		return "NULL"
	}
	return sqlString(*s)
}

func strp(s string) *string { return &s }

// TestB2PCSV pins the one undocumented thing the reader depends on: how
// `dolt sql -r csv` tells SQL NULL from the empty string. NULL is a bare empty
// field and ” is a quoted one, so a reader that keeps the quoting can tell
// them apart and one that does not (encoding/csv) cannot. Nothing in Dolt's
// documentation promises that, so a corpus of awkward values goes through real
// dolt and every cell is checked against what the SQL engine itself says: a
// Dolt release that changes the convention fails here, loudly, instead of
// quietly turning NULL back into ”.
//
// It logs the Dolt version it ran on; the lanes run it once on the local Dolt
// and once on the version the server lane pins.
func TestB2PCSV(t *testing.T) {
	ctx := context.Background()
	dir := replaytest.NewDoltDB(t, "pcsv")
	version, _, _ := strings.Cut(strings.TrimSpace(replaytest.RunDolt(t, dir, "version")), "\n")
	t.Logf("B2.PCSV dolt version: %s", version)

	replaytest.RunDolt(t, dir, "sql", "-q",
		"CREATE TABLE pcsv (id INT PRIMARY KEY, txt TEXT, vc VARCHAR(255), n BIGINT, dt DATETIME, js JSON)")

	textCorpus := []struct {
		name  string
		value *string
	}{
		{"SQL NULL", nil},
		{"empty string", strp("")},
		{"one space", strp(" ")},
		{"padded", strp("  padded  ")},
		{"the word NULL", strp("NULL")},
		{"the word null", strp("null")},
		{"two quote characters", strp(`""`)},
		{"one quote character", strp(`"`)},
		{"comma", strp("a,b")},
		{"lone comma", strp(",")},
		{"leading comma", strp(",lead")},
		{"quote and comma", strp(`say "hi", then leave`)},
		{"newline", strp("line1\nline2")},
		{"carriage return and newline", strp("a\r\nb")},
		{"trailing newline", strp("end\n")},
		{"tab", strp("a\tb")},
		{"backslash", strp(`a\b`)},
		{"non-ASCII", strp("héllo ☃ 日本語")},
		{"leading zeros", strp("007")},
		{"text that looks like JSON", strp(`{"a":1}`)},
	}
	wantText := map[int]*string{}
	for i, c := range textCorpus {
		id := i + 1
		wantText[id] = c.value
		replaytest.RunDolt(t, dir, "sql", "-q", fmt.Sprintf(
			"INSERT INTO pcsv (id, txt, vc) VALUES (%d, %s, %s)", id, sqlNullable(c.value), sqlNullable(c.value)))
	}
	wantN := map[int]string{101: "0", 102: "-1", 103: "1727000000000000123"}
	for id, v := range wantN {
		replaytest.RunDolt(t, dir, "sql", "-q", fmt.Sprintf("INSERT INTO pcsv (id, n) VALUES (%d, %s)", id, v))
	}
	const wantDT = "2026-01-02 03:04:05"
	replaytest.RunDolt(t, dir, "sql", "-q", fmt.Sprintf("INSERT INTO pcsv (id, dt) VALUES (201, %s)", sqlString(wantDT)))
	wantJS := map[int]string{
		301: `{}`,
		302: `null`,
		303: `"NULL"`,
		304: `{"k":"x,y","q":"say \"hi\"","n":"a\nb","u":"héllo"}`,
		305: `[1,2]`,
	}
	for id, doc := range wantJS {
		replaytest.RunDolt(t, dir, "sql", "-q", fmt.Sprintf("INSERT INTO pcsv (id, js) VALUES (%d, %s)", id, sqlString(doc)))
	}
	totalRows := len(textCorpus) + len(wantN) + 1 + len(wantJS)

	// The indicator columns are the SQL engine's own answer to "is this NULL",
	// as text that cannot be confused with the value.
	const wide = "SELECT id, " +
		"txt, IF(txt IS NULL, 'y', 'n') AS txt_null, vc, IF(vc IS NULL, 'y', 'n') AS vc_null, " +
		"n, IF(n IS NULL, 'y', 'n') AS n_null, dt, IF(dt IS NULL, 'y', 'n') AS dt_null, " +
		"js, IF(js IS NULL, 'y', 'n') AS js_null FROM pcsv ORDER BY id"
	header, rows, err := doltcli.Query(ctx, dir, wide)
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	wantHeader := []string{"id", "txt", "txt_null", "vc", "vc_null", "n", "n_null", "dt", "dt_null", "js", "js_null"}
	if !reflect.DeepEqual(header, wantHeader) {
		t.Fatalf("header = %v, want %v", header, wantHeader)
	}
	if len(rows) != totalRows {
		t.Fatalf("got %d rows, want %d", len(rows), totalRows)
	}

	type key struct {
		id  int
		col string
	}
	cells := map[key]doltcli.Cell{}
	for _, row := range rows {
		if len(row) != len(header) {
			t.Fatalf("row has %d cells, want %d: %#v", len(row), len(header), row)
		}
		byName := map[string]doltcli.Cell{}
		for i, h := range header {
			byName[h] = row[i]
		}
		var id int
		if _, err := fmt.Sscanf(byName["id"].Text, "%d", &id); err != nil || byName["id"].Null {
			t.Fatalf("unreadable id cell %#v", byName["id"])
		}
		for _, col := range []string{"txt", "vc", "n", "dt", "js"} {
			ind := byName[col+"_null"]
			if ind.Null || (ind.Text != "y" && ind.Text != "n") {
				t.Fatalf("row %d: indicator %s_null = %#v, want y or n", id, col, ind)
			}
			if got, want := byName[col].Null, ind.Text == "y"; got != want {
				t.Errorf("row %d column %s: Cell.Null = %v, the SQL engine says IS NULL = %v (cell %#v)", id, col, got, want, byName[col])
			}
			cells[key{id, col}] = byName[col]
		}
	}

	for id, want := range wantText {
		for _, col := range []string{"txt", "vc"} {
			got := cells[key{id, col}]
			switch {
			case want == nil && !got.Null:
				t.Errorf("row %d (%s) column %s: want NULL, got %#v", id, textCorpus[id-1].name, col, got)
			case want != nil && (got.Null || got.Text != *want):
				t.Errorf("row %d (%s) column %s: Text = %q (Null=%v), want %q", id, textCorpus[id-1].name, col, got.Text, got.Null, *want)
			}
		}
	}
	for id, want := range wantN {
		if got := cells[key{id, "n"}]; got.Null || got.Text != want {
			t.Errorf("row %d column n: %#v, want %q", id, got, want)
		}
	}
	if got := cells[key{201, "dt"}]; got.Null || got.Text != wantDT {
		t.Errorf("row 201 column dt: %#v, want %q", got, wantDT)
	}
	for id, want := range wantJS {
		got := cells[key{id, "js"}]
		if got.Null {
			t.Errorf("row %d column js: NULL, want the document %s", id, want)
			continue
		}
		var gotDoc, wantDoc any
		if err := json.Unmarshal([]byte(got.Text), &gotDoc); err != nil {
			t.Errorf("row %d column js: %q is not JSON: %v", id, got.Text, err)
			continue
		}
		if err := json.Unmarshal([]byte(want), &wantDoc); err != nil {
			t.Fatalf("test data %s: %v", want, err)
		}
		if !reflect.DeepEqual(gotDoc, wantDoc) {
			t.Errorf("row %d column js: %s, want the document %s", id, got.Text, want)
		}
	}

	// One column at a time is the shape where a NULL row is an empty line, and
	// a reader that skips empty lines would lose the row.
	for _, col := range []string{"txt", "vc", "js"} {
		_, narrow, err := doltcli.Query(ctx, dir, "SELECT "+col+" FROM pcsv ORDER BY id")
		if err != nil {
			t.Fatalf("Query one column %s: %v", col, err)
		}
		if len(narrow) != totalRows {
			t.Fatalf("one-column query of %s returned %d rows, want %d: a NULL row was lost", col, len(narrow), totalRows)
		}
		i := 0
		for _, row := range rows {
			var id int
			_, _ = fmt.Sscanf(row[0].Text, "%d", &id)
			if len(narrow[i]) != 1 {
				t.Fatalf("one-column row %d has %d cells", i, len(narrow[i]))
			}
			if want := cells[key{id, col}]; narrow[i][0] != want {
				t.Errorf("one-column %s, row id %d: %#v, want %#v", col, id, narrow[i][0], want)
			}
			i++
		}
	}

	// A result with no rows is still its schema.
	zeroHeader, zeroRows, err := doltcli.Query(ctx, dir, "SELECT id, txt FROM pcsv WHERE 1 = 0")
	if err != nil {
		t.Fatalf("Query with no rows: %v", err)
	}
	if want := []string{"id", "txt"}; !reflect.DeepEqual(zeroHeader, want) || len(zeroRows) != 0 {
		t.Errorf("empty result = (%v, %v rows), want (%v, 0 rows)", zeroHeader, len(zeroRows), want)
	}
}

// TestSQLQuoteRoundTrip runs awkward values through real dolt: each one, quoted by
// SQLQuote and stored, is found again by an equality test on the same quoting and
// reads back as exactly the string that went in, and none of them runs a
// statement of its own. dolt reads a backslash in a string literal as an escape,
// so a value ending in a backslash-quote pair ends the literal early unless the
// backslash is doubled along with the quote.
func TestSQLQuoteRoundTrip(t *testing.T) {
	dir := replaytest.NewDoltDB(t, "quote")
	ctx := context.Background()
	replaytest.RunDolt(t, dir, "sql", "-q", "CREATE TABLE q (n INT PRIMARY KEY, v TEXT)")

	values := []string{
		"plain-1",
		`it's`,
		`a\b`,
		`trailing\`,
		`x\'`,
		`x\''`,
		`x\'); CREATE TABLE zz_injected (k INT); --`,
		`\\'; CREATE TABLE zz_injected (k INT); --`,
		"two\nlines",
	}
	for i, v := range values {
		if _, err := doltcli.Run(ctx, dir, "sql", "-q", fmt.Sprintf("INSERT INTO q VALUES (%d, %s)", i, doltcli.SQLQuote(v))); err != nil {
			t.Fatalf("storing %q as %s: %v", v, doltcli.SQLQuote(v), err)
		}
	}

	for i, v := range values {
		_, rows, err := doltcli.Query(ctx, dir, "SELECT n, v FROM q WHERE v = "+doltcli.SQLQuote(v))
		if err != nil {
			t.Fatalf("looking up %q: %v", v, err)
		}
		if len(rows) != 1 || rows[0][0].Text != fmt.Sprint(i) || rows[0][1].Text != v {
			t.Errorf("looking up %q found %v, want the one row %d holding it", v, rows, i)
		}
	}

	_, tables, err := doltcli.Query(ctx, dir, "SHOW TABLES")
	if err != nil {
		t.Fatalf("SHOW TABLES: %v", err)
	}
	if len(tables) != 1 || tables[0][0].Text != "q" {
		t.Errorf("tables = %v, want only q: a quoted value ran a statement of its own", tables)
	}
}
