package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func TestLegacyImportSourcePrefixIsGraphOnly(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("source-prefix", "", "")
	if err := cmd.Flags().Set("source-prefix", "web-app"); err != nil {
		t.Fatal(err)
	}
	oldGraph := graphPreviewActive
	graphPreviewActive = false
	defer func() { graphPreviewActive = oldGraph }()
	if err := runImport(cmd, nil); err == nil || !strings.Contains(err.Error(), "graph workspace") {
		t.Fatalf("ordinary import admitted graph-only source configuration: %v", err)
	}
}

// Admission must stop graph-only options before root pre-run can open legacy
// storage. Replacing command globals makes this test intentionally sequential.
func TestLegacyImportSourcePrefixAdmissionBeforeLegacyOpen(t *testing.T) {
	isolateBeadsDirForTest(t)
	t.Chdir(t.TempDir())
	for _, name := range []string{"BEADS_DB", "BD_DB", "BD_GRAPH_MODE"} {
		t.Setenv(name, "")
	}
	directory := t.TempDir()
	t.Setenv("BEADS_DIR", directory)
	oldImport, oldActive, oldConfig, oldDir := importCmd, graphPreviewActive, graphPreviewConfig, graphPreviewDir
	oldGeneration := graphPreviewTypeGeneration
	oldJSON, oldStructured, oldOutput := jsonOutput, graphPreviewStructuredErrors, graphFailureOutput
	t.Cleanup(func() {
		importCmd, graphPreviewActive, graphPreviewConfig, graphPreviewDir = oldImport, oldActive, oldConfig, oldDir
		graphPreviewTypeGeneration = oldGeneration
		jsonOutput, graphPreviewStructuredErrors, graphFailureOutput = oldJSON, oldStructured, oldOutput
	})
	jsonOutput = true
	for _, test := range []struct {
		name, value string
		supplied    bool
	}{
		{"omitted", "", false},
		{"prefix", "web-app", true},
		{"explicit-empty", "", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "import"}
			cmd.Flags().String("source-prefix", "", "")
			if test.supplied {
				if err := cmd.Flags().Set("source-prefix", test.value); err != nil {
					t.Fatal(err)
				}
			}
			importCmd = cmd
			var diagnostic bytes.Buffer
			graphFailureOutput = &diagnostic
			handled, err := admitGraphPreview(cmd)
			if !test.supplied {
				if handled || err != nil || diagnostic.Len() != 0 {
					t.Fatalf("ordinary import was hijacked: handled=%t err=%v diagnostic=%q", handled, err, diagnostic.String())
				}
			} else {
				var failure *exitError
				var envelope struct {
					Code      string `json:"code"`
					Retryable bool   `json:"retryable"`
				}
				if !handled || !errors.As(err, &failure) || failure.Code != 5 {
					t.Fatalf("graph-only prefix fell through to legacy opening: handled=%t err=%v", handled, err)
				}
				if err := json.Unmarshal(diagnostic.Bytes(), &envelope); err != nil || envelope.Code != "capability_unavailable" || envelope.Retryable {
					t.Fatalf("unexpected admission diagnostic: %q (%v)", diagnostic.String(), err)
				}
			}
			if graphPreviewActive || graphPreviewConfig != nil || graphPreviewDir != "" || graphPreviewTypeGeneration != 0 {
				t.Fatal("ordinary admission activated graph storage")
			}
		})
	}
	entries, err := os.ReadDir(directory)
	if err != nil || len(entries) != 0 {
		t.Fatalf("admission created workspace state: %v (%v)", entries, err)
	}
	// Corrupt ordinary metadata must not take the legacy discovery fallback
	// before the unsupported graph-only option can be refused.
	if err := os.WriteFile(filepath.Join(directory, "metadata.json"), []byte("{broken"), 0600); err != nil {
		t.Fatal(err)
	}
	cmd := &cobra.Command{Use: "import"}
	cmd.Flags().String("source-prefix", "", "")
	if err := cmd.Flags().Set("source-prefix", ""); err != nil {
		t.Fatal(err)
	}
	importCmd = cmd
	var diagnostic bytes.Buffer
	graphFailureOutput = &diagnostic
	handled, err := admitGraphPreview(cmd)
	var failure *exitError
	if !handled || !errors.As(err, &failure) || failure.Code != 5 || !strings.Contains(diagnostic.String(), "graph_not_initialized") || !strings.Contains(diagnostic.String(), "invalid character") {
		t.Fatalf("corrupt metadata allowed legacy fallback or lost its cause: handled=%t err=%v diagnostic=%q", handled, err, diagnostic.String())
	}
}

func TestGraphPreviewAssigneeOwnershipError(t *testing.T) {
	stderr, err := os.CreateTemp(t.TempDir(), "stderr")
	if err != nil {
		t.Fatal(err)
	}
	oldStderr, oldJSON := os.Stderr, jsonOutput
	os.Stderr, jsonOutput = stderr, true
	t.Cleanup(func() {
		os.Stderr, jsonOutput = oldStderr, oldJSON
		_ = stderr.Close()
	})
	result := graphStorageError(fmt.Errorf("active assignment: %w", storage.ErrAlreadyClaimed))
	var failure *exitError
	if !errors.As(result, &failure) || failure.Code != 4 {
		t.Fatalf("ownership refusal must exit4: %v", result)
	}
	if _, err := stderr.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	var envelope struct {
		Code      string `json:"code"`
		Message   string `json:"message"`
		Retryable bool   `json:"retryable"`
	}
	if err := json.NewDecoder(stderr).Decode(&envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Code != "constraint_violation" || envelope.Retryable || envelope.Message != "active assignment: "+storage.ErrAlreadyClaimed.Error() {
		t.Fatalf("ownership refusal misclassified: %+v", envelope)
	}
}

func TestGraphPreviewImportDeadlineDisposition(t *testing.T) {
	for _, cancellation := range []error{context.DeadlineExceeded, context.Canceled} {
		for _, test := range []struct {
			name string
			wrap error
			code string
			exit int
		}{
			{"plain", nil, "capability_unavailable", 5},
			{"invalid-store", graphstore.ErrInvalidStore, "capability_unavailable", 5},
			{"uncertain", graphstore.ErrOutcomeUnknown, "outcome_unknown", 6},
			{"uncertain-invalid-store", errors.Join(graphstore.ErrOutcomeUnknown, graphstore.ErrInvalidStore), "outcome_unknown", 6},
		} {
			name := fmt.Sprintf("%v/%s", cancellation, test.name)
			t.Run(name, func(t *testing.T) {
				stderr, err := os.CreateTemp(t.TempDir(), "stderr")
				if err != nil {
					t.Fatal(err)
				}
				oldStderr, oldJSON := os.Stderr, jsonOutput
				os.Stderr, jsonOutput = stderr, true
				defer func() { os.Stderr, jsonOutput = oldStderr, oldJSON; _ = stderr.Close() }()
				input := cancellation
				if test.wrap != nil {
					input = errors.Join(test.wrap, cancellation)
				}
				result := graphStorageError(input)
				var failure *exitError
				if !errors.As(result, &failure) || failure.Code != test.exit {
					t.Fatalf("error=%v want exit=%d", result, test.exit)
				}
				if _, err := stderr.Seek(0, 0); err != nil {
					t.Fatal(err)
				}
				var envelope struct {
					Code      string `json:"code"`
					Retryable bool   `json:"retryable"`
				}
				if err := json.NewDecoder(stderr).Decode(&envelope); err != nil {
					t.Fatal(err)
				}
				if envelope.Code != test.code || envelope.Retryable {
					t.Fatalf("envelope=%+v", envelope)
				}
			})
		}
	}
}

func TestGraphPreviewLegacyImportSummaryPrefix(t *testing.T) {
	withPrefix := graphPreviewLegacyImportSummary("Imported", "source.jsonl", graphstore.LegacyImportResult{IssuePrefix: "web-app", Issues: 1})
	if !strings.Contains(withPrefix, "Workspace Issue prefix: web-app.") {
		t.Fatalf("adopted prefix not disclosed: %s", withPrefix)
	}
	preview := graphPreviewLegacyImportSummary("Would import", "stdin", graphstore.LegacyImportResult{IssuePrefix: "web-app", DryRun: true})
	if !strings.Contains(preview, "Would adopt Issue prefix: web-app.") || strings.Contains(preview, "Workspace Issue prefix:") {
		t.Fatalf("dry-run prefix adoption claimed: %s", preview)
	}
	memories := graphPreviewLegacyImportSummary("Would import", "stdin", graphstore.LegacyImportResult{Memories: 1})
	if strings.Contains(memories, "Issue prefix:") {
		t.Fatalf("memory-only prefix invented: %s", memories)
	}
}
