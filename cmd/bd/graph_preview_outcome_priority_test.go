package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func TestGraphPreviewOutcomeUnknownPrecedesDefiniteErrors(t *testing.T) {
	for _, cause := range []error{
		graphstore.ErrAlreadyExists,
		graphstore.ErrNotFound,
		graphstore.ErrGone,
		graphstore.ErrCapabilityUnavailable,
		graphstore.ErrLimitExceeded,
		&graphstore.ErrAmbiguousLink{},
		&exitError{Code: 2},
	} {
		t.Run(cause.Error(), func(t *testing.T) {
			var diagnostic bytes.Buffer
			oldOutput, oldJSON := graphFailureOutput, jsonOutput
			graphFailureOutput, jsonOutput = &diagnostic, true
			t.Cleanup(func() { graphFailureOutput, jsonOutput = oldOutput, oldJSON })

			result := graphStorageError(errors.Join(cause, graphstore.ErrOutcomeUnknown))
			var failure *exitError
			if !errors.As(result, &failure) || failure.Code != 6 {
				t.Fatalf("uncertain write classified as definite: %v", result)
			}
			var envelope struct {
				Code      string `json:"code"`
				Message   string `json:"message"`
				Retryable bool   `json:"retryable"`
			}
			if err := json.Unmarshal(diagnostic.Bytes(), &envelope); err != nil {
				t.Fatalf("invalid diagnostic %q: %v", diagnostic.String(), err)
			}
			if envelope.Code != "outcome_unknown" || envelope.Retryable ||
				!strings.Contains(envelope.Message, "do not automatically replay") ||
				!strings.Contains(envelope.Message, "inspect the canonical ID") {
				t.Fatalf("uncertain write lost recovery guidance: %+v", envelope)
			}
		})
	}
}
