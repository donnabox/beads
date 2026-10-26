package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestGraphLegacyUnknownFieldsWarning(t *testing.T) {
	var warnings bytes.Buffer
	batch, err := parseGraphLegacyInput(strings.NewReader(`{"id":"private-id","title":"private-title","future\n\u001b[31m":{"secret":"private-value"}}`), true, &warnings)
	if err != nil || len(batch.Issues) != 1 {
		t.Fatalf("batch=%+v err=%v", batch, err)
	}
	text := warnings.String()
	if !strings.Contains(text, `storage admission has not occurred`) || !strings.Contains(text, `["issue","future\n\u001b[31m"]: 1 occurrence`) || strings.ContainsAny(text, "\x1b") || strings.Count(text, "\n") != 2 {
		t.Fatalf("unsafe or misleading diagnostic: %q", text)
	}
	for _, private := range []string{"private-id", "private-title", "private-value", "secret"} {
		if strings.Contains(text, private) {
			t.Fatalf("diagnostic exposed source context: %q", text)
		}
	}
	var stdout bytes.Buffer
	if err := graphPrintTo(&stdout, map[string]int{"issues": len(batch.Issues)}, "parsed", false, true); err != nil || !json.Valid(stdout.Bytes()) || strings.Contains(stdout.String(), "Warning") {
		t.Fatalf("structured stdout contaminated: %q (%v)", stdout.String(), err)
	}
	for _, input := range []string{`{"title":"plain"}`, `{"title":"bad","priority":"x","future":true}`} {
		warnings.Reset()
		_, _ = parseGraphLegacyInput(strings.NewReader(input), true, &warnings)
		if warnings.Len() != 0 {
			t.Fatalf("warning before complete parsing or without loss: %q", warnings.String())
		}
	}
	warnings.Reset()
	if _, err := parseGraphLegacyInput(strings.NewReader(`{"title":"x","future":1}`), false, &warnings); err == nil || warnings.Len() != 0 {
		t.Fatalf("false option was not strict: %v %q", err, warnings.String())
	}
	failure := errors.New("warning sink unavailable")
	batch, err = parseGraphLegacyInput(strings.NewReader(`{"title":"x","future":1}`), true, &graphFailingWriter{err: failure})
	var warningError *graphLegacyImportWarningError
	if !errors.Is(err, failure) || !errors.As(err, &warningError) || len(batch.Issues) != 0 {
		t.Fatalf("warning failure lost classification/cause or did not stop before admission: %+v (%v)", batch, err)
	}
}

func TestGraphLegacyUnknownFieldsAdmissionBeforeProvider(t *testing.T) {
	isolateBeadsDirForTest(t)
	t.Chdir(t.TempDir())
	for _, name := range []string{"BEADS_DB", "BD_DB", "BD_GRAPH_MODE"} {
		t.Setenv(name, "")
	}
	dir := t.TempDir()
	t.Setenv("BEADS_DIR", dir)
	oldCmd, oldActive, oldConfig, oldDir := importCmd, graphPreviewActive, graphPreviewConfig, graphPreviewDir
	oldGeneration, oldJSON, oldStructured, oldOutput := graphPreviewTypeGeneration, jsonOutput, graphPreviewStructuredErrors, graphFailureOutput
	t.Cleanup(func() {
		importCmd, graphPreviewActive, graphPreviewConfig, graphPreviewDir = oldCmd, oldActive, oldConfig, oldDir
		graphPreviewTypeGeneration, jsonOutput, graphPreviewStructuredErrors, graphFailureOutput = oldGeneration, oldJSON, oldStructured, oldOutput
	})
	jsonOutput = true
	for _, damaged := range []bool{false, true} {
		if damaged {
			if err := os.WriteFile(filepath.Join(dir, "metadata.json"), []byte("{broken"), 0600); err != nil {
				t.Fatal(err)
			}
		}
		for _, value := range []string{"omitted", "true", "false"} {
			cmd := &cobra.Command{Use: "import"}
			parent := &cobra.Command{Use: "bd"}
			parent.AddCommand(cmd)
			cmd.Flags().Bool("ignore-unknown-fields", false, "")
			if value != "omitted" {
				if err := cmd.Flags().Set("ignore-unknown-fields", value); err != nil {
					t.Fatal(err)
				}
			}
			importCmd = cmd
			var diagnostic bytes.Buffer
			graphFailureOutput = &diagnostic
			handled, err := admitGraphPreview(cmd)
			if value == "omitted" {
				if handled || err != nil || diagnostic.Len() != 0 {
					t.Fatalf("ordinary omitted option changed: damaged=%t handled=%t err=%v diagnostic=%q", damaged, handled, err, diagnostic.String())
				}
				continue
			}
			var failure *exitError
			var envelope struct {
				Code string `json:"code"`
			}
			want := "capability_unavailable"
			if damaged {
				want = "graph_not_initialized"
			}
			if !handled || !errors.As(err, &failure) || failure.Code != 5 || json.Unmarshal(diagnostic.Bytes(), &envelope) != nil || envelope.Code != want {
				t.Fatalf("presence fell through: damaged=%t value=%s handled=%t err=%v diagnostic=%q", damaged, value, handled, err, diagnostic.String())
			}
			if graphPreviewActive || graphPreviewConfig != nil || graphPreviewDir != "" || graphPreviewTypeGeneration != 0 {
				t.Fatal("legacy gate activated graph storage")
			}
			if err := runImport(cmd, nil); err == nil || !strings.Contains(err.Error(), "graph workspace") {
				t.Fatalf("ordinary handler ignored explicit presence: %v", err)
			}
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 1 || entries[0].Name() != "metadata.json" {
		t.Fatalf("provider gate wrote workspace state: %v (%v)", entries, err)
	}
	contents, err := os.ReadFile(filepath.Join(dir, "metadata.json"))
	if err != nil || string(contents) != "{broken" {
		t.Fatal("damaged metadata was changed")
	}
}

// The actual flag binding and route allowlist must both reach opt-in parsing.
// A real warning-write failure stops the valid parse before any provider opens;
// strict false instead stops on the unknown field. No test-only API is needed.
func TestGraphLegacyUnknownFieldsRoutePlumbing(t *testing.T) {
	stdioMutex.Lock()
	defer stdioMutex.Unlock()
	isolateBeadsDirForTest(t)
	work := t.TempDir()
	t.Chdir(work)
	t.Setenv("BD_MIGRATION_FREEZE_FILE", "")
	t.Setenv("BEADS_DIR", filepath.Join(work, ".beads"))
	resetCommandFlags(rootCmd)
	defer resetCommandFlags(rootCmd)
	oldActive, oldDir := graphPreviewActive, graphPreviewDir
	oldStderr, oldOutput := os.Stderr, graphFailureOutput
	t.Cleanup(func() {
		graphPreviewActive, graphPreviewDir = oldActive, oldDir
		os.Stderr, graphFailureOutput = oldStderr, oldOutput
	})
	graphPreviewActive, graphPreviewDir, jsonOutput = true, work, true
	closed, err := os.CreateTemp(t.TempDir(), "closed-warning-sink")
	if err != nil {
		t.Fatal(err)
	}
	if err := closed.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stderr = closed
	source := filepath.Join(work, "source.jsonl")
	if err := os.WriteFile(source, []byte(`{"id":"old-a","title":"private-source-title","future":"private-source-value"}`), 0600); err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{"omitted", "true", "false"} {
		if value != "omitted" {
			if err := importCmd.Flags().Set("ignore-unknown-fields", value); err != nil {
				t.Fatal(err)
			}
		}
		var diagnostic bytes.Buffer
		graphFailureOutput = &diagnostic
		err := runImport(importCmd, []string{source})
		var failure *exitError
		var envelope struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}
		wantCode, wantExit := "invalid_properties", 2
		want := "unknown legacy field"
		if value == "true" {
			wantCode, wantExit, want = "output_error", 5, os.ErrClosed.Error()
		}
		if !errors.As(err, &failure) || failure.Code != wantExit || json.Unmarshal(diagnostic.Bytes(), &envelope) != nil || envelope.Code != wantCode {
			t.Fatalf("flag/error did not reach its contract: value=%s err=%v diagnostic=%q", value, err, diagnostic.String())
		}
		if !strings.Contains(envelope.Message, want) {
			t.Fatalf("wrong option reached parser: value=%s diagnostic=%q want %q", value, diagnostic.String(), want)
		}
		if value == "true" && !strings.Contains(envelope.Message, "storage admission has not occurred") {
			t.Fatalf("warning failure omitted its no-effect disposition: %q", diagnostic.String())
		}
		for _, sourceValue := range []string{"old-a", "private-source-title", "private-source-value"} {
			if strings.Contains(envelope.Message, sourceValue) {
				t.Fatalf("error envelope exposed source value %q: %q", sourceValue, diagnostic.String())
			}
		}
	}
	if _, err := os.Stat(filepath.Join(work, ".beads")); !os.IsNotExist(err) {
		t.Fatalf("route opened or created workspace: %v", err)
	}
}

func TestGraphLegacyUnknownFieldsTerminalWarnings(t *testing.T) {
	for _, name := range []string{"c1\u009b2K\u009dtitle", "del\x7f", "bidi\u202eoverride", "é😀", `literal\u009b`} {
		raw, _ := json.Marshal(map[string]any{"title": "private-title", name: "private-value"})
		var warnings bytes.Buffer
		if _, err := parseGraphLegacyInput(strings.NewReader(string(raw)), true, &warnings); err != nil {
			t.Fatal(err)
		}
		for _, r := range warnings.String() {
			if (r < 0x20 && r != '\n') || r >= 0x7f {
				t.Fatalf("terminal control/non-ASCII rune %U in warning %q", r, warnings.String())
			}
		}
		if !strings.Contains(warnings.String(), "1 unknown field across 1 distinct path;") || !strings.Contains(warnings.String(), ": 1 occurrence\n") || strings.Contains(warnings.String(), "private-") {
			t.Fatalf("warning lost grammar or disclosed values: %q", warnings.String())
		}
	}
	prefix := strings.Repeat("x", 64)
	raw, _ := json.Marshal(map[string]any{"title": "x", prefix + "one": 1, prefix + "two": 2})
	var warnings bytes.Buffer
	if _, err := parseGraphLegacyInput(strings.NewReader(string(raw)), true, &warnings); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(warnings.String()), "\n")
	if len(lines) != 3 || strings.TrimSpace(lines[1]) == strings.TrimSpace(lines[2]) || !strings.Contains(lines[0], "2 unknown fields across 2 distinct paths;") || strings.Count(warnings.String(), "full-path digest") != 2 {
		t.Fatalf("truncated warnings indistinguishable: %q", warnings.String())
	}
}
