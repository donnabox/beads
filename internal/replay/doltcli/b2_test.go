package doltcli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// B2.Tokenizer: the tokenizer behind Query turns dolt's csv text into cells,
// and the one signal it must keep is whether a field was quoted: a bare empty
// field is SQL NULL, a quoted empty field is the empty string.
func TestB2TokenizerCells(t *testing.T) {
	null := Cell{Null: true}
	text := func(s string) Cell { return Cell{Text: s} }
	cases := []struct {
		name       string
		in         string
		wantHeader []string
		wantRows   [][]Cell
	}{
		{
			name:       "a bare empty field is NULL and a quoted empty field is the empty string",
			in:         "a,b\n,\"\"\n",
			wantHeader: []string{"a", "b"},
			wantRows:   [][]Cell{{null, text("")}},
		},
		{
			name:       "a single-column NULL row is an empty line, not a missing row",
			in:         "a\n\n\"\"\nx\n",
			wantHeader: []string{"a"},
			wantRows:   [][]Cell{{null}, {text("")}, {text("x")}},
		},
		{
			name:       "the words NULL and null are text",
			in:         "a,b\nNULL,null\n",
			wantHeader: []string{"a", "b"},
			wantRows:   [][]Cell{{text("NULL"), text("null")}},
		},
		{
			name:       "quoted fields carry commas, doubled quotes and newlines",
			in:         "a,b\n\"x,y\",\"say \"\"hi\"\"\"\n\"l1\nl2\",z\n",
			wantHeader: []string{"a", "b"},
			wantRows:   [][]Cell{{text("x,y"), text("say \"hi\"")}, {text("l1\nl2"), text("z")}},
		},
		{
			name:       "CRLF record terminators are accepted and a CR inside quotes is data",
			in:         "a,b\r\n1,2\r\n,\"\"\r\n\"x\r\ny\",3\r\n",
			wantHeader: []string{"a", "b"},
			wantRows:   [][]Cell{{text("1"), text("2")}, {null, text("")}, {text("x\r\ny"), text("3")}},
		},
		{
			name:       "the last record needs no terminator",
			in:         "a,b\n1,2",
			wantHeader: []string{"a", "b"},
			wantRows:   [][]Cell{{text("1"), text("2")}},
		},
		{
			name:       "a header with no rows is a schema",
			in:         "a,b\n",
			wantHeader: []string{"a", "b"},
			wantRows:   nil,
		},
		{
			name:       "no output at all is no header and no rows",
			in:         "",
			wantHeader: nil,
			wantRows:   nil,
		},
		{
			name:       "quoted header names are unquoted",
			in:         "\"a b\",c\n1,2\n",
			wantHeader: []string{"a b", "c"},
			wantRows:   [][]Cell{{text("1"), text("2")}},
		},
		{
			name:       "NULL at every position of a row",
			in:         "a,b,c\n,2,\n1,,3\n,,\n",
			wantHeader: []string{"a", "b", "c"},
			wantRows:   [][]Cell{{null, text("2"), null}, {text("1"), null, text("3")}, {null, null, null}},
		},
		{
			name:       "a space is data and a quoted space keeps it",
			in:         "a\n\" \"\n",
			wantHeader: []string{"a"},
			wantRows:   [][]Cell{{text(" ")}},
		},
		{
			name:       "non-ASCII text is kept byte for byte",
			in:         "a\nh\u00e9llo \u2603\n",
			wantHeader: []string{"a"},
			wantRows:   [][]Cell{{text("h\u00e9llo \u2603")}},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			header, rows, err := parseCSV([]byte(c.in))
			if err != nil {
				t.Fatalf("parseCSV(%q): %v", c.in, err)
			}
			if !reflect.DeepEqual(header, c.wantHeader) {
				t.Errorf("header = %#v, want %#v", header, c.wantHeader)
			}
			if len(rows) != len(c.wantRows) || (len(rows) > 0 && !reflect.DeepEqual(rows, c.wantRows)) {
				t.Errorf("rows = %#v, want %#v", rows, c.wantRows)
			}
		})
	}

	for name, in := range map[string]string{
		"an unterminated quote":                   "a,b\n1,\"2\n",
		"a bare quote inside an unquoted field":   "a,b\n1,x\"y\n",
		"text after a closing quote":              "a,b\n1,\"x\"y\n",
		"too few fields":                          "a,b\n1\n",
		"too many fields":                         "a,b\n1,2,3\n",
		"an empty line under a two-column header": "a,b\n\n",
	} {
		t.Run("malformed: "+name, func(t *testing.T) {
			if _, _, err := parseCSV([]byte(in)); err == nil {
				t.Fatalf("parseCSV(%q) succeeded, want an error", in)
			}
		})
	}
}

// B2.StderrNoise: a warning dolt prints on stderr, even one shaped like a csv
// row, never becomes a row and never disturbs the ones on stdout.
func TestB2StderrNoise(t *testing.T) {
	t.Run("a csv-shaped warning next to real rows", func(t *testing.T) {
		putDoltFirst(t, "echo 'warning: x, y, z' >&2\nprintf 'id,title\\nx-1,\"Widget, deluxe\"\\n,\"\"\\n'")
		header, rows, err := Query(context.Background(), t.TempDir(), "SELECT 1")
		if err != nil {
			t.Fatalf("Query: %v", err)
		}
		if want := []string{"id", "title"}; !reflect.DeepEqual(header, want) {
			t.Errorf("header = %v, want %v", header, want)
		}
		want := [][]Cell{{{Text: "x-1"}, {Text: "Widget, deluxe"}}, {{Null: true}, {Text: ""}}}
		if !reflect.DeepEqual(rows, want) {
			t.Errorf("rows = %#v, want %#v", rows, want)
		}
	})
	t.Run("a csv-shaped warning next to a header-only result", func(t *testing.T) {
		putDoltFirst(t, "echo 'warning: x, y, z' >&2\nprintf 'id,title\\n'")
		header, rows, err := Query(context.Background(), t.TempDir(), "SELECT 1")
		if err != nil {
			t.Fatalf("Query: %v", err)
		}
		if want := []string{"id", "title"}; !reflect.DeepEqual(header, want) {
			t.Errorf("header = %v, want %v", header, want)
		}
		if len(rows) != 0 {
			t.Errorf("rows = %#v, want none: stderr must never become a row", rows)
		}
	})
}

// deadPID returns the pid of a process that has already exited.
func deadPID(t *testing.T) int {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=^$")
	if err := cmd.Run(); err != nil {
		t.Fatalf("running a short-lived process: %v", err)
	}
	return cmd.Process.Pid
}

// writeServerInfo records info in <dir>/.dolt/sql-server.info, the file a dolt
// sql-server writes at startup (pid, port and a server id, colon separated).
func writeServerInfo(t *testing.T, dir, info string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, ".dolt"), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".dolt", "sql-server.info"), []byte(info), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

// B2.StaleServerInfo: a directory counts as served only while the pid its
// sql-server.info names is alive. A record left behind by a server that has
// exited is stale and must not refuse every later query in that directory.
func TestB2StaleServerInfo(t *testing.T) {
	const serverID = "0b9f1c0e-0000-4000-8000-000000000000"
	root := t.TempDir()
	child := filepath.Join(root, "clone")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	writeServerInfo(t, root, fmt.Sprintf("%d:3306:%s", deadPID(t), serverID))
	if got := servedDataDir(child); got != "" {
		t.Errorf("servedDataDir with a dead pid recorded = %q, want empty (stale record)", got)
	}

	writeServerInfo(t, root, fmt.Sprintf("%d:3306:%s", os.Getpid(), serverID))
	if got := servedDataDir(child); got != root {
		t.Errorf("servedDataDir with a live pid recorded = %q, want %q", got, root)
	}

	// A record that cannot be read as a pid cannot be proven stale, so the
	// directory stays refused.
	writeServerInfo(t, root, "{}")
	if got := servedDataDir(child); got != root {
		t.Errorf("servedDataDir with an unreadable record = %q, want %q", got, root)
	}

	// A stale record nearer the query directory must not hide a live server
	// higher up.
	inner := filepath.Join(root, "inner")
	writeServerInfo(t, inner, fmt.Sprintf("%d:3306:%s", deadPID(t), serverID))
	writeServerInfo(t, root, fmt.Sprintf("%d:3306:%s", os.Getpid(), serverID))
	if got := servedDataDir(filepath.Join(inner, "clone")); got != root {
		t.Errorf("servedDataDir with a stale record below a live one = %q, want %q", got, root)
	}

	// Query follows the same rule.
	writeServerInfo(t, root, fmt.Sprintf("%d:3306:%s", deadPID(t), serverID))
	putDoltFirst(t, "printf 'a\\n1\\n'")
	if _, _, err := Query(context.Background(), child, "SELECT 1"); err != nil {
		t.Errorf("Query under a stale sql-server.info: %v, want it to run", err)
	}
	writeServerInfo(t, root, fmt.Sprintf("%d:3306:%s", os.Getpid(), serverID))
	if _, _, err := Query(context.Background(), child, "SELECT 1"); err == nil || !strings.Contains(err.Error(), "sql-server is serving") {
		t.Errorf("Query under a live sql-server.info: err = %v, want a refusal naming the server", err)
	}
}

// B2.RowMapText: RowMap is the text-only view of a row, kept for callers that
// do not distinguish NULL from the empty string.
func TestB2RowMapText(t *testing.T) {
	got := RowMap([]string{"id", "note", "other"}, []Cell{{Text: "x-1"}, {Null: true}, {Text: ""}})
	want := map[string]string{"id": "x-1", "note": "", "other": ""}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("RowMap = %v, want %v", got, want)
	}
}
