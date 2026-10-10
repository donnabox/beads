package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

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
