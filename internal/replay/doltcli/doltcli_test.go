package doltcli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func hasVar(env []string, name string) bool {
	prefix := name + "="
	for _, kv := range env {
		if strings.HasPrefix(kv, prefix) {
			return true
		}
	}
	return false
}

func TestSanitizedEnv_StripsAmbientVars(t *testing.T) {
	base := []string{
		"BEADS_DOLT_SERVER_PORT=12345",
		"BEADS_DOLT_PORT=12345",
		"BEADS_ACTOR=someone",
		"BD_ACTOR=someone",
		"GT_ROOT=/some/path",
		"BEADS_DIR=/some/beads/dir",
		"BEADS_HOLDER_TOKEN=secret",
		"GC_BEADS_SCOPE_ROOT=/scope",
		"BEADS_DOLT_AUTO_START=1",
		"BEADS_DOLT_SYNC_CLI_REMOTES=1",
		"BEADS_BACKUP_ENABLED=1",
		"PATH=/usr/bin",
		"HOME=/nonexistent-home",
	}
	got := SanitizedEnv(base)

	for _, forbidden := range []string{
		"BEADS_DOLT_SERVER_PORT", "BEADS_DOLT_PORT", "BEADS_ACTOR", "BD_ACTOR", "GT_ROOT",
		"BEADS_DIR", "BEADS_HOLDER_TOKEN", "GC_BEADS_SCOPE_ROOT", "BEADS_DOLT_AUTO_START",
		"BEADS_DOLT_SYNC_CLI_REMOTES", "BEADS_BACKUP_ENABLED",
	} {
		if hasVar(got, forbidden) {
			t.Errorf("SanitizedEnv leaked ambient var %s", forbidden)
		}
	}
	for _, kept := range []string{"PATH", "HOME"} {
		if !hasVar(got, kept) {
			t.Errorf("SanitizedEnv dropped non-ambient var %s", kept)
		}
	}
}

func TestSanitizedEnv_EmptyInput(t *testing.T) {
	if got := SanitizedEnv(nil); len(got) != 0 {
		t.Errorf("SanitizedEnv(nil) = %v, want empty", got)
	}
}

func TestServedDataDir_DetectsLiveServer(t *testing.T) {
	root := t.TempDir()
	child := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if got := servedDataDir(child); got != "" {
		t.Errorf("servedDataDir(%s) = %q, want empty (no live server)", child, got)
	}
	served := filepath.Join(root, "a")
	if err := os.MkdirAll(filepath.Join(served, ".dolt"), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(served, ".dolt", "sql-server.info"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if got := servedDataDir(child); got != served {
		t.Errorf("servedDataDir(%s) = %q, want %q", child, got, served)
	}
}

// putDoltFirst installs a stand-in dolt with the given script body ahead of any
// real one on PATH.
func putDoltFirst(t *testing.T, body string) {
	t.Helper()
	dir := t.TempDir()
	script := "#!/bin/sh\n" + body + "\n"
	if err := os.WriteFile(filepath.Join(dir, "dolt"), []byte(script), 0o755); err != nil {
		t.Fatalf("writing the dolt stand-in: %v", err)
	}
	t.Setenv("PATH", dir+":/usr/bin:/bin")
}

func TestQuery_RefusesServedDir(t *testing.T) {
	// No dolt is needed: the refusal happens before anything is started.
	t.Setenv("PATH", t.TempDir())
	served := t.TempDir()
	if err := os.MkdirAll(filepath.Join(served, ".dolt"), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(served, ".dolt", "sql-server.info"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	_, _, err := Query(context.Background(), filepath.Join(served, "clone"), "SELECT 1")
	if err == nil || !strings.Contains(err.Error(), "sql-server is serving") {
		t.Fatalf("Query on a served directory: err = %v, want a refusal naming the server", err)
	}
}

func TestQuery_ParsesStdoutOnly(t *testing.T) {
	putDoltFirst(t, "echo 'warning: something noisy' >&2\nprintf 'id,title\\nx-1,\"Widget, deluxe\"\\n'")
	header, rows, err := Query(context.Background(), t.TempDir(), "SELECT 1")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if want := []string{"id", "title"}; !reflect.DeepEqual(header, want) {
		t.Errorf("header = %v, want %v", header, want)
	}
	if want := [][]Cell{{{Text: "x-1"}, {Text: "Widget, deluxe"}}}; !reflect.DeepEqual(rows, want) {
		t.Errorf("rows = %v, want %v", rows, want)
	}
}

func TestQuery_NoOutputMeansNoRows(t *testing.T) {
	putDoltFirst(t, "true")
	header, rows, err := Query(context.Background(), t.TempDir(), "SELECT 1")
	if err != nil || header != nil || rows != nil {
		t.Fatalf("Query with no output = (%v, %v, %v), want (nil, nil, nil)", header, rows, err)
	}
}

func TestQuery_ReportsStderrOnFailure(t *testing.T) {
	putDoltFirst(t, "echo 'table not found: nope' >&2\nexit 1")
	_, _, err := Query(context.Background(), t.TempDir(), "SELECT * FROM nope")
	if err == nil || !strings.Contains(err.Error(), "table not found: nope") {
		t.Fatalf("Query error = %v, want it to carry dolt's stderr", err)
	}
}

func TestQuery_MissingDolt(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	if _, _, err := Query(context.Background(), t.TempDir(), "SELECT 1"); err == nil {
		t.Fatal("Query with no dolt on PATH: want an error")
	}
}

func TestRowMap(t *testing.T) {
	got := RowMap([]string{"id", "title", "status"}, []Cell{{Text: "x-1"}, {Text: "Widget"}})
	want := map[string]string{"id": "x-1", "title": "Widget"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("RowMap with a short row = %v, want %v", got, want)
	}
}

func TestSQLQuote(t *testing.T) {
	for in, want := range map[string]string{
		"x-1":   "'x-1'",
		"it's":  "'it''s'",
		"":      "''",
		"a''b":  "'a''''b'",
		"a b c": "'a b c'",
		// dolt reads a backslash in a string literal as an escape, so each one is
		// doubled as well: a lone backslash before a doubled quote would otherwise
		// escape the first of the pair and end the literal on the second.
		`a\b`:  `'a\\b'`,
		`\`:    `'\\'`,
		`a\\b`: `'a\\\\b'`,
		`x\'`:  `'x\\'''`,
	} {
		if got := SQLQuote(in); got != want {
			t.Errorf("SQLQuote(%q) = %s, want %s", in, got, fmt.Sprint(want))
		}
	}
}
