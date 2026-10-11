package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
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
		for _, uncertain := range []bool{false, true} {
			name := fmt.Sprintf("%v/uncertain=%t", cancellation, uncertain)
			t.Run(name, func(t *testing.T) {
				stderr, err := os.CreateTemp(t.TempDir(), "stderr")
				if err != nil {
					t.Fatal(err)
				}
				oldStderr, oldJSON := os.Stderr, jsonOutput
				os.Stderr, jsonOutput = stderr, true
				defer func() { os.Stderr, jsonOutput = oldStderr, oldJSON; _ = stderr.Close() }()
				input := cancellation
				code, exit := "capability_unavailable", 5
				if uncertain {
					input = errors.Join(graphstore.ErrOutcomeUnknown, cancellation)
					code, exit = "outcome_unknown", 6
				}
				result := graphStorageError(input)
				var failure *exitError
				if !errors.As(result, &failure) || failure.Code != exit {
					t.Fatalf("error=%v want exit=%d", result, exit)
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
				if envelope.Code != code || envelope.Retryable {
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
