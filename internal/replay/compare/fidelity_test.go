package compare

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/replay/doltcli"
	"github.com/steveyegge/beads/internal/replay/oracle"
	"github.com/steveyegge/beads/internal/replay/replaytest"
	"github.com/steveyegge/beads/internal/storage/issueops"
)

func TestMain(m *testing.M) {
	os.Exit(replaytest.Main(m))
}

// b2Row builds a row from alternating column names and values: a string is
// text, nil is SQL NULL.
func b2Row(kv ...any) oracle.Row {
	var r oracle.Row
	for i := 0; i < len(kv); i += 2 {
		r.Columns = append(r.Columns, kv[i].(string))
		switch v := kv[i+1].(type) {
		case nil:
			r.Cells = append(r.Cells, doltcli.Cell{Null: true})
		case string:
			r.Cells = append(r.Cells, doltcli.Cell{Text: v})
		default:
			panic("b2Row: value must be a string or nil")
		}
	}
	return r
}

func b2View(issue oracle.Row, deps ...oracle.Row) *oracle.View {
	return &oracle.View{Issue: issue, Dependencies: deps}
}

// b2Edge is a dependency row carrying every column of the table.
func b2Edge(id, issueID, typ, target string) oracle.Row {
	return b2Row("id", id, "issue_id", issueID, "type", typ, "created_at", "2026-01-01 00:00:00", "created_by", "someone",
		"metadata", "{}", "thread_id", nil, "depends_on_issue_id", target, "depends_on_wisp_id", nil, "depends_on_external", nil)
}

// b2EdgeMetadata is b2Edge with its metadata cell replaced: a string is the
// stored document, nil is SQL NULL.
func b2EdgeMetadata(meta any) oracle.Row {
	row := b2Edge("d1", "x-1", "blocks", "x-2")
	for i, col := range row.Columns {
		if col != "metadata" {
			continue
		}
		switch v := meta.(type) {
		case nil:
			row.Cells[i] = doltcli.Cell{Null: true}
		case string:
			row.Cells[i] = doltcli.Cell{Text: v}
		default:
			panic("b2EdgeMetadata: value must be a string or nil")
		}
	}
	return row
}

// b2RowWithMetadataAt is b2Row of kv with a metadata cell inserted before the
// pair at index pos, so a test can put the column first, in the middle or last.
func b2RowWithMetadataAt(pos int, meta any, kv ...any) oracle.Row {
	args := append([]any{}, kv[:2*pos]...)
	args = append(args, "metadata", meta)
	args = append(args, kv[2*pos:]...)
	return b2Row(args...)
}

// b2Issue is an issues row carrying a few compared columns and a few stamps.
func b2Issue(id string, extra ...any) oracle.Row {
	kv := []any{"id", id, "title", "A title", "status", "open", "notes", nil,
		"created_at", "2026-01-01 00:00:00", "updated_at", "2026-01-01 00:00:00", "metadata", "{}"}
	// A later pair for a column replaces the default; a new column is appended.
	for i := 0; i < len(extra); i += 2 {
		replaced := false
		for j := 0; j < len(kv); j += 2 {
			if kv[j] == extra[i] {
				kv[j+1] = extra[i+1]
				replaced = true
			}
		}
		if !replaced {
			kv = append(kv, extra[i], extra[i+1])
		}
	}
	return b2Row(kv...)
}

// B2.NullVsEmpty: SQL NULL and the empty string are different values, so they
// give different payloads and do not match. Both used to read as "" and did.
func TestB2NullVsEmpty(t *testing.T) {
	t.Run("payloads", func(t *testing.T) {
		withNull := b2View(b2Issue("x-1", "notes", nil))
		withEmpty := b2View(b2Issue("x-1", "notes", ""))
		pn, err := Payload(withNull)
		if err != nil {
			t.Fatalf("Payload(NULL): %v", err)
		}
		pe, err := Payload(withEmpty)
		if err != nil {
			t.Fatalf("Payload(''): %v", err)
		}
		if bytes.Equal(pn, pe) {
			t.Fatalf("NULL and '' give the same payload: %s", pn)
		}
		if !strings.Contains(string(pn), `"notes":null`) {
			t.Errorf("NULL payload = %s, want notes as a JSON null", pn)
		}
		if !strings.Contains(string(pe), `"notes":""`) {
			t.Errorf("empty-string payload = %s, want notes as an empty JSON string", pe)
		}
		res, _, err := CompareViews(withNull, withEmpty)
		if err != nil {
			t.Fatalf("CompareViews: %v", err)
		}
		if res.Matched {
			t.Error("a NULL column matched an empty-string column")
		}
		if res.Mismatch == nil || !bytes.Equal(res.Mismatch.ExpectedJSON, pn) || !bytes.Equal(res.Mismatch.ActualJSON, pe) {
			t.Errorf("mismatch record = %+v, want both payloads verbatim", res.Mismatch)
		}
	})

	t.Run("read from real dolt databases", func(t *testing.T) {
		replaytest.Require(t, replaytest.NeedDolt)
		read := func(name, notes string) *oracle.View {
			dir := replaytest.NewDoltDB(t, name)
			replaytest.RunDolt(t, dir, "sql", "-q", "CREATE TABLE issues (id VARCHAR(255) PRIMARY KEY, title VARCHAR(500) NOT NULL, notes LONGTEXT)")
			replaytest.RunDolt(t, dir, "sql", "-q", "INSERT INTO issues (id, title, notes) VALUES ('x-1', 'T', "+notes+")")
			replaytest.RunDolt(t, dir, "add", "-A")
			replaytest.RunDolt(t, dir, "commit", "-m", "one row")
			v, err := oracle.ReadView(t.Context(), dir, replaytest.HeadCommit(t, dir), "x-1")
			if err != nil || v == nil {
				t.Fatalf("ReadView %s = (%v, %v)", name, v, err)
			}
			return v
		}
		null, empty := read("nulldb", "NULL"), read("emptydb", "''")
		if res, _, err := CompareViews(null, empty); err != nil || res.Matched {
			t.Errorf("NULL row vs '' row: (matched=%v, err=%v), want not matched", res.Matched, err)
		}
		if res, _, err := CompareViews(null, read("nulldb2", "NULL")); err != nil || !res.Matched {
			t.Errorf("NULL row vs NULL row: (matched=%v, err=%v), want matched", res.Matched, err)
		}
	})
}

// B2.IntGate: a number outside the exact range is not comparable evidence.
// Canonicalizing as a double turns ...993 into ...992, so a comparison that
// only hashed the canonical form would call two different values equal.
func TestB2IntGate(t *testing.T) {
	assertUncomparable := func(t *testing.T, res Result, err error, oracleJSON, candidateJSON string) {
		t.Helper()
		if err != nil {
			t.Fatalf("Compare: %v, want an uncomparable result and no error (the run continues)", err)
		}
		if res.Matched {
			t.Fatal("Matched = true for a pair holding a number outside the exact range")
		}
		if res.Mismatch == nil || res.Mismatch.Category != "number-fidelity" {
			t.Fatalf("Mismatch = %+v, want category number-fidelity", res.Mismatch)
		}
		if string(res.Mismatch.ExpectedJSON) != oracleJSON || string(res.Mismatch.ActualJSON) != candidateJSON {
			t.Errorf("payloads = (%s, %s), want both verbatim (%s, %s)", res.Mismatch.ExpectedJSON, res.Mismatch.ActualJSON, oracleJSON, candidateJSON)
		}
		if !res.Uncomparable() {
			t.Error("Uncomparable() = false, want true so the run can count it separately")
		}
	}

	t.Run("the public compare refuses ...993 against ...992", func(t *testing.T) {
		a, b := `{"id":9007199254740993}`, `{"id":9007199254740992}`
		res, err := Compare([]byte(a), []byte(b))
		assertUncomparable(t, res, err, a, b)
	})
	t.Run("the same oversized literal on both sides is still not a match", func(t *testing.T) {
		a := `{"id":9007199254740993}`
		res, err := Compare([]byte(a), []byte(a))
		assertUncomparable(t, res, err, a, a)
	})
	t.Run("a number nested in a document is refused too", func(t *testing.T) {
		a, b := `{"metadata":{"ts":1727000000000000123}}`, `{"metadata":{"ts":1}}`
		res, err := Compare([]byte(a), []byte(b))
		assertUncomparable(t, res, err, a, b)
	})
	t.Run("numbers inside the exact range still compare, whatever their spelling", func(t *testing.T) {
		res, err := Compare([]byte(`{"n":1}`), []byte(`{"n":1.0}`))
		if err != nil || !res.Matched || res.Uncomparable() {
			t.Errorf("1 vs 1.0: (matched=%v, uncomparable=%v, err=%v), want matched", res.Matched, res.Uncomparable(), err)
		}
		res, err = Compare([]byte(`{"n":9007199254740991}`), []byte(`{"n":9007199254740990}`))
		if err != nil || res.Matched || res.Uncomparable() || res.Mismatch == nil {
			t.Errorf("two exact integers that differ: (matched=%v, uncomparable=%v, err=%v), want a plain mismatch", res.Matched, res.Uncomparable(), err)
		}
	})
	t.Run("through views, a JSON column holding such a number is uncomparable", func(t *testing.T) {
		res, _, err := CompareViews(
			b2View(b2Issue("x-1", "metadata", `{"n":9007199254740993}`)),
			b2View(b2Issue("x-1", "metadata", `{"n":9007199254740992}`)))
		if err != nil {
			t.Fatalf("CompareViews: %v", err)
		}
		if res.Matched || !res.Uncomparable() {
			t.Errorf("(matched=%v, uncomparable=%v), want uncomparable", res.Matched, res.Uncomparable())
		}
	})
	t.Run("a number that looks like one but sits in a text column is just text", func(t *testing.T) {
		res, _, err := CompareViews(
			b2View(b2Issue("x-1", "title", "9007199254740993")),
			b2View(b2Issue("x-1", "title", "9007199254740992")))
		if err != nil || res.Matched || res.Uncomparable() {
			t.Errorf("(matched=%v, uncomparable=%v, err=%v), want a plain mismatch: only JSON documents are gated", res.Matched, res.Uncomparable(), err)
		}
	})
}

// B2.IntGateOneSided: the number gate reads both payloads, so a number outside
// the exact range makes a pair uncomparable whichever side holds it. The first
// two pairs are ones double rounding would call equal if the gate looked at one
// side only (the oversized number canonicalizes onto the exact one); the third
// shows the gate also reaches a number nested in a document.
func TestB2IntGateOneSided(t *testing.T) {
	uncomparable := func(t *testing.T, oracleJSON, candidateJSON string) {
		t.Helper()
		res, err := Compare([]byte(oracleJSON), []byte(candidateJSON))
		if err != nil {
			t.Fatalf("Compare: %v, want an uncomparable result and no error (the run continues)", err)
		}
		if res.Matched {
			t.Fatal("Matched = true for a pair holding a number outside the exact range")
		}
		if res.Mismatch == nil || res.Mismatch.Category != "number-fidelity" {
			t.Fatalf("Mismatch = %+v, want category number-fidelity", res.Mismatch)
		}
		if string(res.Mismatch.ExpectedJSON) != oracleJSON || string(res.Mismatch.ActualJSON) != candidateJSON {
			t.Errorf("payloads = (%s, %s), want both verbatim (%s, %s)", res.Mismatch.ExpectedJSON, res.Mismatch.ActualJSON, oracleJSON, candidateJSON)
		}
		if !res.Uncomparable() {
			t.Error("Uncomparable() = false, want true so the run can count it separately")
		}
	}
	for _, p := range []struct{ name, exact, oversized string }{
		{"an integer past 2^53", `{"n":9007199254740992}`, `{"n":9007199254740993}`},
		{"a decimal that rounds onto an exact integer", `{"n":9007199254740994}`, `{"n":9007199254740993.5}`},
		{"a number nested in a document", `{"metadata":{"ts":1}}`, `{"metadata":{"ts":1727000000000000123}}`},
	} {
		t.Run(p.name+", held by the oracle side", func(t *testing.T) { uncomparable(t, p.oversized, p.exact) })
		t.Run(p.name+", held by the candidate side", func(t *testing.T) { uncomparable(t, p.exact, p.oversized) })
	}
	t.Run("through views, a JSON column holding such a number is uncomparable on either side", func(t *testing.T) {
		exact, oversized := `{"n":9007199254740992}`, `{"n":9007199254740993}`
		for _, c := range []struct{ name, oracleMeta, candidateMeta string }{
			{"oracle side", oversized, exact},
			{"candidate side", exact, oversized},
		} {
			res, _, err := CompareViews(
				b2View(b2Issue("x-1", "metadata", c.oracleMeta)),
				b2View(b2Issue("x-1", "metadata", c.candidateMeta)))
			if err != nil {
				t.Fatalf("%s: CompareViews: %v", c.name, err)
			}
			if res.Matched || !res.Uncomparable() {
				t.Errorf("%s: (matched=%v, uncomparable=%v), want uncomparable", c.name, res.Matched, res.Uncomparable())
			}
		}
	})
}

// B2.GateAgrees: the harness and the store's mint ask the same question about a
// number, because the harness calls the mint's own gate. The literals are the
// ones the number-fidelity ruling lists; the test does not hard-code which are
// refused, so it stays right whichever version of the rule is at its base.
func TestB2GateAgrees(t *testing.T) {
	corpus := []string{
		"9007199254740993.5", "9007199254740994.5", "9007199254740991.5", "9007199254740990.5",
		"0.1", "3.14159265358979323846", "1e-400",
		"1e100000000", "-1e400", "1e-100000000", "1e1000000",
		"1727000000000000000", "1e300", "9007199254740993", "-9007199254740993", "9007199254740991",
		"9007199254740992", "9.007199254740993e15", "9007199254740993.0",
		"0", "-0", "1.0", "1E30", "42",
	}
	for _, lit := range corpus {
		t.Run(lit, func(t *testing.T) {
			doc := []byte(`{"n":` + lit + `}`)
			res, cmpErr := Compare(doc, doc)
			_, mintErr := issueops.ReplayCanonicalDurableState(doc)
			gateErr := issueops.ReplayRefuseUnrepresentableIntegers(doc)

			if gateErr != nil {
				// The gate refuses it: the compare must say so without failing the
				// run, and the mint must refuse it.
				if cmpErr != nil || res.Matched || !res.Uncomparable() {
					t.Errorf("gate refuses %s but Compare = (matched=%v, uncomparable=%v, err=%v)", lit, res.Matched, res.Uncomparable(), cmpErr)
				}
				if mintErr == nil {
					t.Errorf("gate refuses %s but the mint entry admitted it", lit)
				}
				if errors.Is(gateErr, issueops.ErrIntegerNotRepresentable) != errors.Is(mintErr, issueops.ErrIntegerNotRepresentable) {
					t.Errorf("gate and mint disagree about the error class for %s: %v vs %v", lit, gateErr, mintErr)
				}
				return
			}
			// The gate admits it: the rest of the pipeline must treat it the same
			// way on both sides, and a document compared with itself must match.
			if (cmpErr != nil) != (mintErr != nil) {
				t.Errorf("for %s Compare err = %v but the mint entry err = %v", lit, cmpErr, mintErr)
			}
			if cmpErr == nil && (res.Uncomparable() || !res.Matched) {
				t.Errorf("gate admits %s but Compare of the document with itself = (matched=%v, uncomparable=%v)", lit, res.Matched, res.Uncomparable())
			}
		})
	}
}

// B2.UTF8: a payload is checked with utf8.Valid and invalid bytes are an
// error. They are never replaced with U+FFFD, which would let two different
// byte strings compare equal.
func TestB2UTF8(t *testing.T) {
	bad := []byte("{\"a\":\"\xff\"}")
	alsoBad := []byte("{\"a\":\"\xfe\"}")
	good := []byte(`{"a":"x"}`)
	wantUTF8Error := func(t *testing.T, what string, err error) {
		t.Helper()
		if err == nil || !strings.Contains(strings.ToLower(err.Error()), "utf-8") {
			t.Errorf("%s: err = %v, want an error naming invalid UTF-8", what, err)
		}
	}
	_, err := Compare(bad, good)
	wantUTF8Error(t, "invalid oracle payload", err)
	_, err = Compare(good, bad)
	wantUTF8Error(t, "invalid candidate payload", err)
	res, err := Compare(bad, alsoBad)
	wantUTF8Error(t, "two different invalid payloads", err)
	if res.Matched {
		t.Error("two different invalid payloads matched")
	}

	_, err = Payload(b2View(b2Issue("x-1", "title", "bad \xff text")))
	wantUTF8Error(t, "Payload of a cell with invalid bytes", err)
	_, err = Payload(b2View(b2Issue("x-1", "metadata", "{\"k\":\"\xff\"}")))
	wantUTF8Error(t, "Payload of a JSON cell with invalid bytes", err)
	_, _, err = CompareViews(b2View(b2Issue("x-1", "title", "bad \xff text")), b2View(b2Issue("x-1")))
	wantUTF8Error(t, "CompareViews with an invalid cell", err)

	// Valid non-ASCII is fine, and a mismatch keeps both payloads verbatim.
	a, b := `{"a":"héllo ☃"}`, `{"a":"hello"}`
	res, err = Compare([]byte(a), []byte(b))
	if err != nil || res.Matched || res.Mismatch == nil {
		t.Fatalf("valid non-ASCII vs ASCII: (%+v, %v), want a mismatch", res, err)
	}
	if string(res.Mismatch.ExpectedJSON) != a || string(res.Mismatch.ActualJSON) != b {
		t.Errorf("mismatch payloads = (%s, %s), want both verbatim", res.Mismatch.ExpectedJSON, res.Mismatch.ActualJSON)
	}
}

// B2.NonObjectPayload: a payload is one JSON object. Anything else is an error,
// not a pair the classifier quietly files as "unclassified".
func TestB2NonObjectPayload(t *testing.T) {
	obj := `{"a":1}`
	for _, nonObject := range []string{`[]`, `[1,2]`, `3`, `"x"`, `true`, `null`} {
		t.Run(nonObject, func(t *testing.T) {
			for _, c := range []struct{ name, o, c string }{
				{"oracle side", nonObject, obj},
				{"candidate side", obj, nonObject},
				{"both sides", nonObject, nonObject},
			} {
				res, err := Compare([]byte(c.o), []byte(c.c))
				if !errors.Is(err, ErrNotObject) {
					t.Errorf("%s: err = %v, want errors.Is ErrNotObject", c.name, err)
				}
				if res.Matched {
					t.Errorf("%s: a non-object payload matched", c.name)
				}
			}
		})
	}
	if _, err := Compare([]byte(`{`), []byte(obj)); err == nil {
		t.Error("truncated JSON: want an error")
	}
	if _, err := Compare([]byte(`{"a":1} {"b":2}`), []byte(obj)); err == nil {
		t.Error("two JSON values in one payload: want an error")
	}
}

// B2.SchemaSkew: a column present on one side only is compared on the
// intersection of the two schemas and recorded as skew, never as a mismatch.
func TestB2SchemaSkew(t *testing.T) {
	base := b2Issue("x-1")
	t.Run("the candidate has a column the oracle lacks", func(t *testing.T) {
		res, skew, err := CompareViews(b2View(base), b2View(b2Issue("x-1", "spec_id", "S-1")))
		if err != nil || !res.Matched {
			t.Fatalf("(matched=%v, err=%v), want matched on the intersection", res.Matched, err)
		}
		want := Skew{"issues": {CandidateOnly: []string{"spec_id"}}}
		if !reflect.DeepEqual(skew, want) {
			t.Errorf("skew = %#v, want %#v", skew, want)
		}
	})
	t.Run("the oracle has a column the candidate lacks", func(t *testing.T) {
		res, skew, err := CompareViews(b2View(b2Issue("x-1", "spec_id", "S-1")), b2View(base))
		if err != nil || !res.Matched {
			t.Fatalf("(matched=%v, err=%v), want matched on the intersection", res.Matched, err)
		}
		want := Skew{"issues": {OracleOnly: []string{"spec_id"}}}
		if !reflect.DeepEqual(skew, want) {
			t.Errorf("skew = %#v, want %#v", skew, want)
		}
	})
	t.Run("skew does not hide a real difference on the intersection", func(t *testing.T) {
		res, skew, err := CompareViews(b2View(base), b2View(b2Issue("x-1", "spec_id", "S-1", "title", "Another")))
		if err != nil || res.Matched {
			t.Fatalf("(matched=%v, err=%v), want a mismatch on title", res.Matched, err)
		}
		if len(skew["issues"].CandidateOnly) != 1 {
			t.Errorf("skew = %#v, want the one-sided column still recorded", skew)
		}
	})
	t.Run("the dependencies table can skew too", func(t *testing.T) {
		edge := b2Edge("d1", "x-1", "blocks", "x-2")
		newer := b2Row("id", "d1", "issue_id", "x-1", "type", "blocks", "created_at", "2026-01-01 00:00:00", "created_by", "someone",
			"metadata", "{}", "thread_id", nil, "depends_on_issue_id", "x-2", "depends_on_wisp_id", nil, "depends_on_external", nil, "weight", "3")
		res, skew, err := CompareViews(b2View(base, edge), b2View(base, newer))
		if err != nil || !res.Matched {
			t.Fatalf("(matched=%v, err=%v), want matched on the intersection", res.Matched, err)
		}
		want := Skew{"dependencies": {CandidateOnly: []string{"weight"}}}
		if !reflect.DeepEqual(skew, want) {
			t.Errorf("skew = %#v, want %#v", skew, want)
		}
	})
	t.Run("no skew when the schemas agree", func(t *testing.T) {
		if _, skew, err := CompareViews(b2View(base), b2View(b2Issue("x-1"))); err != nil || len(skew) != 0 {
			t.Errorf("skew = %#v (err %v), want none", skew, err)
		}
	})
}

// B2.PayloadShape: a view becomes one JSON object: every compared column as
// null or its exact text, JSON columns as documents, stamps left out, and the
// dependencies section alongside the issue's own columns.
func TestB2PayloadShape(t *testing.T) {
	view := b2View(b2Issue("x-1", "metadata", `{"b":2,"a":[1,2]}`), b2Edge("d1", "x-1", "blocks", "x-2"))
	raw, err := Payload(view)
	if err != nil {
		t.Fatalf("Payload: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("payload is not a JSON object: %v\n%s", err, raw)
	}
	for _, stamp := range []string{"created_at", "updated_at"} {
		if _, ok := doc[stamp]; ok {
			t.Errorf("payload carries the stamp column %s", stamp)
		}
	}
	for col, want := range map[string]any{"id": "x-1", "title": "A title", "status": "open", "notes": nil} {
		got, ok := doc[col]
		if !ok || got != want {
			t.Errorf("payload %s = (%v, present=%v), want %v", col, got, ok, want)
		}
	}
	meta, ok := doc["metadata"].(map[string]any)
	if !ok || !reflect.DeepEqual(meta["a"], []any{1.0, 2.0}) || meta["b"] != 2.0 {
		t.Errorf("payload metadata = %#v, want the parsed document, not a string", doc["metadata"])
	}
	deps, ok := doc["dependencies"].([]any)
	if !ok || len(deps) != 1 {
		t.Fatalf("payload dependencies = %#v, want one edge", doc["dependencies"])
	}
	edge := deps[0].(map[string]any)
	for _, stamp := range []string{"id", "created_at", "created_by"} {
		if _, ok := edge[stamp]; ok {
			t.Errorf("edge carries the stamp column %s", stamp)
		}
	}
	if edge["type"] != "blocks" || edge["depends_on_issue_id"] != "x-2" || edge["thread_id"] != nil {
		t.Errorf("edge = %#v", edge)
	}
	if _, present := edge["thread_id"]; !present {
		t.Error("a NULL edge column must appear as null, not be dropped")
	}

	none, err := Payload(nil)
	if err != nil || string(none) != `{}` {
		t.Errorf("Payload(nil) = (%s, %v), want an empty object", none, err)
	}
	empty, err := Payload(b2View(b2Issue("x-1")))
	if err != nil {
		t.Fatalf("Payload: %v", err)
	}
	if !strings.Contains(string(empty), `"dependencies":[]`) {
		t.Errorf("payload of an issue with no edges = %s, want an empty dependencies section", empty)
	}
}

// B2.JSONColumnNull: a JSON column can hold SQL NULL or the document null. They
// are different stored values, and rendering both as null called them equal. A
// SQL NULL in a JSON column is left out of the payload and the document null is
// the member with the value null; every other column still renders SQL NULL as
// null, so SQL NULL and the text null stay apart there too.
func TestB2JSONColumnNull(t *testing.T) {
	// A carrier puts one metadata cell (nil is SQL NULL, a string is the stored
	// document) into a JSON column of one table. section picks the JSON object
	// that holds the member out of a decoded payload. wantCategory is what a
	// mismatch in that table is called: a difference in the dependencies section is
	// a dep-edge, and the pair needs no category of its own either way.
	carriers := []struct {
		name         string
		wantCategory string
		kv           []any // the other columns the position test puts around metadata
		view         func(meta any) *oracle.View
		viewOfRow    func(row oracle.Row) *oracle.View
		section      func(doc map[string]any) map[string]any
	}{
		{
			name:         "issues.metadata",
			wantCategory: "unclassified",
			kv:           []any{"id", "x-1", "title", "T"},
			view:         func(meta any) *oracle.View { return b2View(b2Issue("x-1", "metadata", meta)) },
			viewOfRow:    func(row oracle.Row) *oracle.View { return b2View(row) },
			section:      func(doc map[string]any) map[string]any { return doc },
		},
		{
			name:         "dependencies.metadata",
			wantCategory: "dep-edge",
			kv:           []any{"issue_id", "x-1", "type", "blocks"},
			view:         func(meta any) *oracle.View { return b2View(b2Issue("x-1"), b2EdgeMetadata(meta)) },
			viewOfRow:    func(row oracle.Row) *oracle.View { return b2View(b2Issue("x-1"), row) },
			section: func(doc map[string]any) map[string]any {
				edges, _ := doc["dependencies"].([]any)
				if len(edges) != 1 {
					return nil
				}
				edge, _ := edges[0].(map[string]any)
				return edge
			},
		},
	}
	decode := func(t *testing.T, v *oracle.View) (raw []byte, doc map[string]any) {
		t.Helper()
		raw, err := Payload(v)
		if err != nil {
			t.Fatalf("Payload: %v", err)
		}
		if err := json.Unmarshal(raw, &doc); err != nil {
			t.Fatalf("payload is not a JSON object: %v\n%s", err, raw)
		}
		return raw, doc
	}

	for _, c := range carriers {
		t.Run(c.name, func(t *testing.T) {
			t.Run("comparing", func(t *testing.T) {
				for _, p := range []struct {
					name         string
					oracle, cand any
					wantMatched  bool
				}{
					{"SQL NULL against the document null", nil, "null", false},
					{"the document null against SQL NULL", "null", nil, false},
					{"SQL NULL against SQL NULL", nil, nil, true},
					{"the document null against the document null", "null", "null", true},
					{"SQL NULL against an empty object", nil, "{}", false},
					{"an empty object against SQL NULL", "{}", nil, false},
					{"the document null against an empty object", "null", "{}", false},
					{"an empty object against the document null", "{}", "null", false},
				} {
					t.Run(p.name, func(t *testing.T) {
						res, _, err := CompareViews(c.view(p.oracle), c.view(p.cand))
						if err != nil {
							t.Fatalf("CompareViews: %v", err)
						}
						if res.Matched != p.wantMatched {
							t.Fatalf("Matched = %v, want %v", res.Matched, p.wantMatched)
						}
						if !p.wantMatched && (res.Mismatch == nil || res.Mismatch.Category != c.wantCategory) {
							t.Errorf("Mismatch = %+v, want category %q", res.Mismatch, c.wantCategory)
						}
					})
				}
			})

			t.Run("payload shape", func(t *testing.T) {
				member := func(meta any) (value any, present bool) {
					_, doc := decode(t, c.view(meta))
					section := c.section(doc)
					if section == nil {
						t.Fatalf("the payload has no such section: %v", doc)
					}
					value, present = section["metadata"]
					return value, present
				}
				if v, present := member(nil); present {
					t.Errorf("SQL NULL rendered as metadata = %v, want the member left out", v)
				}
				if v, present := member("null"); !present || v != nil {
					t.Errorf("the document null rendered as (%v, present=%v), want the member present with the value null", v, present)
				}
				if v, present := member("{}"); !present || !reflect.DeepEqual(v, map[string]any{}) {
					t.Errorf("an empty object rendered as (%v, present=%v), want the member present as an empty object", v, present)
				}
			})

			// b2RowWithMetadataAt orders the columns as given, so the omitted member can
			// be the first one written, a middle one, the last, or the only one.
			t.Run("the omitted member can be first, in the middle, last or alone", func(t *testing.T) {
				for _, p := range []struct {
					name string
					pos  int
					kv   []any
				}{
					{"first", 0, c.kv},
					{"in the middle", 1, c.kv},
					{"last", 2, c.kv},
					{"alone", 0, nil},
				} {
					t.Run(p.name, func(t *testing.T) {
						row := func(meta any) oracle.Row { return b2RowWithMetadataAt(p.pos, meta, p.kv...) }
						raw, doc := decode(t, c.viewOfRow(row(nil)))
						if !json.Valid(raw) {
							t.Fatalf("payload is not valid JSON: %s", raw)
						}
						section := c.section(doc)
						if section == nil {
							t.Fatalf("the payload has no such section: %s", raw)
						}
						if _, present := section["metadata"]; present {
							t.Errorf("payload %s carries a metadata member for SQL NULL", raw)
						}
						for i := 0; i < len(p.kv); i += 2 {
							if got := section[p.kv[i].(string)]; got != p.kv[i+1] {
								t.Errorf("payload %s: %s = %v, want %v", raw, p.kv[i], got, p.kv[i+1])
							}
						}
						res, _, err := CompareViews(c.viewOfRow(row(nil)), c.viewOfRow(row(nil)))
						if err != nil || !res.Matched {
							t.Errorf("SQL NULL against SQL NULL: (matched=%v, err=%v), want matched", res.Matched, err)
						}
						res, _, err = CompareViews(c.viewOfRow(row(nil)), c.viewOfRow(row("null")))
						if err != nil || res.Matched {
							t.Errorf("SQL NULL against the document null: (matched=%v, err=%v), want a mismatch", res.Matched, err)
						}
					})
				}
			})
		})
	}

	t.Run("a TEXT column keeps SQL NULL and the text null apart, as before", func(t *testing.T) {
		notes := func(v any) (value any, present bool) {
			_, doc := decode(t, b2View(b2Issue("x-1", "notes", v)))
			value, present = doc["notes"]
			return value, present
		}
		if v, present := notes(nil); !present || v != nil {
			t.Errorf("SQL NULL notes rendered as (%v, present=%v), want notes present with the value null", v, present)
		}
		if v, present := notes("null"); !present || v != "null" {
			t.Errorf("the text null in notes rendered as (%v, present=%v), want notes as the string \"null\"", v, present)
		}
	})
}

// B2.StampsAreNotCompared: columns bd fills from the wall clock, derives or
// takes from the writing environment never decide a match; every other column
// does, including one the static lists have never heard of.
func TestB2StampsAreNotCompared(t *testing.T) {
	base := b2Issue("x-1")
	same := func(t *testing.T, what string, a, b *oracle.View) {
		t.Helper()
		res, _, err := CompareViews(a, b)
		if err != nil || !res.Matched {
			t.Errorf("%s: (matched=%v, err=%v), want matched", what, res.Matched, err)
		}
	}
	differ := func(t *testing.T, what string, a, b *oracle.View) {
		t.Helper()
		res, _, err := CompareViews(a, b)
		if err != nil || res.Matched {
			t.Errorf("%s: (matched=%v, err=%v), want a mismatch", what, res.Matched, err)
		}
	}
	same(t, "updated_at", b2View(base), b2View(b2Issue("x-1", "updated_at", "2031-05-05 05:05:05")))
	same(t, "created_at", b2View(base), b2View(b2Issue("x-1", "created_at", "2031-05-05 05:05:05")))
	same(t, "a dependency's created_at, created_by and surrogate id",
		b2View(base, b2Edge("d1", "x-1", "blocks", "x-2")),
		b2View(base, b2Row("id", "other-id", "issue_id", "x-1", "type", "blocks", "created_at", "2031-01-01 00:00:00", "created_by", "someone else",
			"metadata", "{}", "thread_id", nil, "depends_on_issue_id", "x-2", "depends_on_wisp_id", nil, "depends_on_external", nil)))
	differ(t, "title", b2View(base), b2View(b2Issue("x-1", "title", "Different")))
	differ(t, "due_at is user input and is compared", b2View(b2Issue("x-1", "due_at", "2026-02-01 00:00:00")), b2View(b2Issue("x-1", "due_at", "2026-02-02 00:00:00")))
	differ(t, "defer_until is user input and is compared", b2View(b2Issue("x-1", "defer_until", nil)), b2View(b2Issue("x-1", "defer_until", "2026-02-02 00:00:00")))
	differ(t, "a column no list knows about", b2View(b2Issue("x-1", "brand_new_column", "a")), b2View(b2Issue("x-1", "brand_new_column", "b")))
}

// B2.SectionCategories: a difference is categorised by the section it is in,
// so a dependency difference is a dep-edge whatever else differs.
func TestB2SectionCategories(t *testing.T) {
	base := b2Issue("x-1")
	edgeA := b2Edge("d1", "x-1", "blocks", "x-2")
	edgeB := b2Edge("d1", "x-1", "blocks", "x-3")
	cases := []struct {
		name     string
		oracle   *oracle.View
		cand     *oracle.View
		wantCat  string
		wantDiff bool
	}{
		{"an extra edge on the candidate", b2View(base), b2View(base, edgeA), "dep-edge", true},
		{"an edge the candidate lacks, both issues present", b2View(base, edgeA), b2View(base), "dep-edge", true},
		{"an edge pointing elsewhere", b2View(base, edgeA), b2View(base, edgeB), "dep-edge", true},
		{"a different edge type", b2View(base, edgeA), b2View(base, b2Edge("d1", "x-1", "related", "x-2")), "dep-edge", true},
		{"edge and assignee both differ: the edge is the more actionable signal", b2View(base, edgeA), b2View(b2Issue("x-1", "assignee", "bob"), edgeB), "dep-edge", true},
		{"only the assignee differs", b2View(b2Issue("x-1", "assignee", "amy")), b2View(b2Issue("x-1", "assignee", "bob")), "attribution", true},
		{"only the title differs", b2View(base), b2View(b2Issue("x-1", "title", "Other")), "unclassified", true},
		{"only a user-set timestamp differs", b2View(b2Issue("x-1", "due_at", "2026-02-01 00:00:00")), b2View(b2Issue("x-1", "due_at", "2026-02-02 00:00:00")), "as-of-mismatch", true},
		{"identical views", b2View(base, edgeA), b2View(base, edgeA), "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res, _, err := CompareViews(c.oracle, c.cand)
			if err != nil {
				t.Fatalf("CompareViews: %v", err)
			}
			if res.Matched == c.wantDiff {
				t.Fatalf("Matched = %v, want %v", res.Matched, !c.wantDiff)
			}
			if c.wantDiff && (res.Mismatch == nil || res.Mismatch.Category != c.wantCat) {
				t.Errorf("Mismatch = %+v, want category %q", res.Mismatch, c.wantCat)
			}
		})
	}
}

// B2.ExistenceVerdicts: an issue that does not exist on either side matches
// (a delete was replayed faithfully); existing on one side only does not, and the
// category says which side lacks it. The category is the only place the case is
// recorded, so it is asserted, not only Matched. Whether a row exists does not
// depend on any number in it, so a number outside the exact range in the present
// view does not make the pair uncomparable.
func TestB2ExistenceVerdicts(t *testing.T) {
	if res, _, err := CompareViews(nil, nil); err != nil || !res.Matched || res.Mismatch != nil {
		t.Errorf("CompareViews(nil, nil) = (matched=%v, mismatch=%+v, err=%v), want matched with no mismatch", res.Matched, res.Mismatch, err)
	}
	oneSided := func(t *testing.T, res Result, err error, wantCategory string) {
		t.Helper()
		if err != nil {
			t.Fatalf("CompareViews: %v", err)
		}
		if res.Matched {
			t.Fatal("Matched = true for a pair where exactly one issue exists")
		}
		if res.Mismatch == nil || res.Mismatch.Category != wantCategory {
			t.Errorf("Mismatch = %+v, want category %q", res.Mismatch, wantCategory)
		}
		if res.Uncomparable() {
			t.Error("Uncomparable() = true, want false: a missing row is a proven mismatch")
		}
	}
	for _, c := range []struct {
		name string
		view *oracle.View
	}{
		// A payload has a dependencies member whether or not the issue has edges.
		{"an issue with no edges", b2View(b2Issue("x-1"))},
		{"an issue with an edge", b2View(b2Issue("x-1"), b2Edge("d1", "x-1", "blocks", "x-2"))},
		{"an issue whose metadata holds a number past 2^53", b2View(b2Issue("x-1", "metadata", `{"n":9007199254740993}`))},
	} {
		t.Run(c.name+", missing from the candidate", func(t *testing.T) {
			res, _, err := CompareViews(c.view, nil)
			oneSided(t, res, err, "issue-missing")
		})
		t.Run(c.name+", extra on the candidate", func(t *testing.T) {
			res, _, err := CompareViews(nil, c.view)
			oneSided(t, res, err, "issue-extra")
		})
	}
}
