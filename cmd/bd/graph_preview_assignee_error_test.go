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
	"time"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/configfile"
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
			parent := &cobra.Command{Use: "bd"}
			parent.AddCommand(cmd)
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
	parent := &cobra.Command{Use: "bd"}
	parent.AddCommand(cmd)
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
			{"uncertain-limit", errors.Join(graphstore.ErrOutcomeUnknown, graphstore.ErrLimitExceeded), "outcome_unknown", 6},
			{"uncertain-capability", errors.Join(graphstore.ErrOutcomeUnknown, graphstore.ErrCapabilityUnavailable), "outcome_unknown", 6},
		} {
			name := fmt.Sprintf("%v/%s", cancellation, test.name)
			t.Run(name, func(t *testing.T) {
				stderr, err := os.CreateTemp(t.TempDir(), "stderr")
				if err != nil {
					t.Fatal(err)
				}
				oldStderr, oldJSON, oldOutput := os.Stderr, jsonOutput, graphFailureOutput
				os.Stderr, jsonOutput, graphFailureOutput = stderr, true, nil
				defer func() { os.Stderr, jsonOutput, graphFailureOutput = oldStderr, oldJSON, oldOutput; _ = stderr.Close() }()
				input := cancellation
				if test.wrap != nil {
					input = errors.Join(test.wrap, cancellation)
				}
				result := graphLegacyImportError(input)
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

func TestGraphPreviewImportStageErrorCategories(t *testing.T) {
	bareSQL := errors.New("catalog insert failed")
	for _, test := range []struct {
		name   string
		mapErr func(error) error
		input  error
		code   string
		exit   int
	}{
		{"open-bare", graphLegacyImportOpenError, bareSQL, "graph_not_initialized", 5},
		{"operation-bare", graphLegacyImportError, bareSQL, "route_unavailable", 5},
		{"operation-invalid-store", graphLegacyImportError, errors.Join(bareSQL, graphstore.ErrInvalidStore), "invalid_store", 5},
		{"operation-uncertain", graphLegacyImportError, errors.Join(bareSQL, graphstore.ErrOutcomeUnknown), "outcome_unknown", 6},
	} {
		t.Run(test.name, func(t *testing.T) {
			oldJSON, oldOutput := jsonOutput, graphFailureOutput
			var diagnostic bytes.Buffer
			jsonOutput, graphFailureOutput = true, &diagnostic
			defer func() { jsonOutput, graphFailureOutput = oldJSON, oldOutput }()
			result := test.mapErr(test.input)
			var failure *exitError
			if !errors.As(result, &failure) || failure.Code != test.exit {
				t.Fatalf("error=%v, want exit=%d", result, test.exit)
			}
			var envelope struct {
				Code      string `json:"code"`
				Retryable bool   `json:"retryable"`
			}
			if err := json.Unmarshal(diagnostic.Bytes(), &envelope); err != nil || envelope.Code != test.code || envelope.Retryable {
				t.Fatalf("diagnostic=%q (%v), want code=%s", diagnostic.String(), err, test.code)
			}
		})
	}
}

func TestGraphPreviewOutcomeUnknownPrecedence(t *testing.T) {
	candidateIDs := []string{"https://example.invalid/work/links/a", "https://example.invalid/work/links/b"}
	reported := &exitError{Code: 2}
	for _, test := range []struct {
		name  string
		cause error
	}{
		{"limit", graphstore.ErrLimitExceeded},
		{"capability", graphstore.ErrCapabilityUnavailable},
		{"ambiguous", &graphstore.ErrAmbiguousLink{CandidateIDs: candidateIDs}},
		// Defensive composition: current transaction constructors do not return
		// exitError, but the shared mapper must not erase a supplied uncertainty.
		{"reported-defensive", reported},
	} {
		t.Run(test.name, func(t *testing.T) {
			stderr, err := os.CreateTemp(t.TempDir(), "stderr")
			if err != nil {
				t.Fatal(err)
			}
			oldStderr, oldJSON, oldOutput := os.Stderr, jsonOutput, graphFailureOutput
			os.Stderr, jsonOutput, graphFailureOutput = stderr, true, nil
			defer func() { os.Stderr, jsonOutput, graphFailureOutput = oldStderr, oldJSON, oldOutput; _ = stderr.Close() }()
			result := graphStorageError(errors.Join(test.cause, graphstore.ErrOutcomeUnknown))
			var failure *exitError
			if !errors.As(result, &failure) || failure.Code != 6 {
				t.Fatalf("uncertainty was hidden: %v", result)
			}
			raw, err := os.ReadFile(stderr.Name())
			if err != nil {
				t.Fatal(err)
			}
			var envelope struct {
				Code      string `json:"code"`
				Message   string `json:"message"`
				Retryable bool   `json:"retryable"`
			}
			if err := json.Unmarshal(raw, &envelope); err != nil || envelope.Code != "outcome_unknown" || envelope.Retryable || !strings.Contains(envelope.Message, "do not automatically replay") || !strings.Contains(envelope.Message, "inspect the canonical ID") {
				t.Fatalf("missing single uncertainty diagnostic or recovery guidance: %q (%v)", raw, err)
			}
		})
	}
}

func TestGraphPreviewDefiniteErrorShortcuts(t *testing.T) {
	candidateIDs := []string{"https://example.invalid/work/links/a", "https://example.invalid/work/links/b"}
	reported := &exitError{Code: 2}
	for _, mapper := range []struct {
		name  string
		apply func(error) error
	}{
		{"ordinary", graphStorageError}, {"import", graphLegacyImportError},
	} {
		for _, test := range []struct {
			name  string
			input error
			code  string
			exit  int
		}{
			{"reported", reported, "", 2},
			{"ambiguous", &graphstore.ErrAmbiguousLink{CandidateIDs: candidateIDs}, "ambiguous_link", 4},
			{"validation", storage.ErrValidation, "invalid_properties", 2},
			{"invalid-store", graphstore.ErrInvalidStore, "invalid_store", 5},
		} {
			t.Run(mapper.name+"/"+test.name, func(t *testing.T) {
				stderr, err := os.CreateTemp(t.TempDir(), "stderr")
				if err != nil {
					t.Fatal(err)
				}
				oldStderr, oldJSON, oldOutput := os.Stderr, jsonOutput, graphFailureOutput
				os.Stderr, jsonOutput, graphFailureOutput = stderr, true, nil
				defer func() { os.Stderr, jsonOutput, graphFailureOutput = oldStderr, oldJSON, oldOutput; _ = stderr.Close() }()
				result := mapper.apply(test.input)
				raw, err := os.ReadFile(stderr.Name())
				if err != nil {
					t.Fatal(err)
				}
				if test.name == "reported" {
					if result != reported || len(raw) != 0 {
						t.Fatalf("reported error changed or emitted a second diagnostic: result=%v raw=%q", result, raw)
					}
					return
				}
				var failure *exitError
				if !errors.As(result, &failure) || failure.Code != test.exit {
					t.Fatalf("definite refusal exit changed: %v want%d", result, test.exit)
				}
				var envelope struct {
					Code         string   `json:"code"`
					Retryable    bool     `json:"retryable"`
					CandidateIDs []string `json:"candidateIDs"`
				}
				if err := json.Unmarshal(raw, &envelope); err != nil || envelope.Code != test.code || envelope.Retryable {
					t.Fatalf("definite refusal diagnostic changed: %q (%v)", raw, err)
				}
				if test.name == "ambiguous" {
					if len(envelope.CandidateIDs) != len(candidateIDs) {
						t.Fatalf("ambiguity candidates changed: %v", envelope.CandidateIDs)
					}
					for i, want := range candidateIDs {
						if envelope.CandidateIDs[i] != want {
							t.Fatalf("ambiguity candidates changed: %v", envelope.CandidateIDs)
						}
					}
				}
			})
		}
	}
}

// Preserve the inherited non-import default category pending a separate CLI
// category/retry decision; it is not a new description of cancellation semantics.
func TestGraphPreviewNonImportCancellationDisposition(t *testing.T) {
	for _, cancellation := range []error{context.DeadlineExceeded, context.Canceled} {
		for _, test := range []struct {
			name string
			wrap error
			code string
			exit int
		}{
			{"plain", nil, "graph_not_initialized", 5},
			{"invalid-store", graphstore.ErrInvalidStore, "invalid_store", 5},
			{"uncertain", graphstore.ErrOutcomeUnknown, "outcome_unknown", 6},
		} {
			t.Run(fmt.Sprintf("%v/%s", cancellation, test.name), func(t *testing.T) {
				stderr, err := os.CreateTemp(t.TempDir(), "stderr")
				if err != nil {
					t.Fatal(err)
				}
				oldStderr, oldJSON, oldOutput := os.Stderr, jsonOutput, graphFailureOutput
				os.Stderr, jsonOutput, graphFailureOutput = stderr, true, nil
				defer func() { os.Stderr, jsonOutput, graphFailureOutput = oldStderr, oldJSON, oldOutput; _ = stderr.Close() }()
				input := cancellation
				if test.wrap != nil {
					input = errors.Join(test.wrap, cancellation)
				}
				result := graphStorageError(input)
				var failure *exitError
				if !errors.As(result, &failure) || failure.Code != test.exit {
					t.Fatalf("non-import error changed: %v, want exit%d", result, test.exit)
				}
				raw, err := os.ReadFile(stderr.Name())
				if err != nil {
					t.Fatal(err)
				}
				var envelope struct {
					Code      string `json:"code"`
					Retryable bool   `json:"retryable"`
				}
				if err := json.Unmarshal(raw, &envelope); err != nil || envelope.Code != test.code || envelope.Retryable {
					t.Fatalf("non-import diagnostic changed: %q (%v)", raw, err)
				}
			})
		}
	}
}

func TestGraphPreviewImportPreservesReportedError(t *testing.T) {
	oldJSON, oldOutput := jsonOutput, graphFailureOutput
	var diagnostic bytes.Buffer
	jsonOutput, graphFailureOutput = true, &diagnostic
	defer func() { jsonOutput, graphFailureOutput = oldJSON, oldOutput }()
	reported := &exitError{Code: 2}
	// Defensive wrapping control: a previously reported refusal must not emit
	// another diagnostic merely because cleanup also supplies cancellation.
	result := graphLegacyImportError(errors.Join(reported, context.Canceled))
	if result != reported || diagnostic.Len() != 0 {
		t.Fatalf("reported import refusal changed or was printed twice: result=%v diagnostic=%q", result, diagnostic.String())
	}
}

// Exercise the actual reader route and raw open-error boundary without a live
// engine. database/sql rejects the completed context before dialing the owned,
// nonexistent socket; no production endpoint or initialized store is involved.
func TestGraphPreviewImportCanceledOpenDisposition(t *testing.T) {
	for _, expired := range []bool{false, true} {
		name := "canceled"
		if expired {
			name = "deadline"
		}
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			if expired {
				ctx, cancel = context.WithDeadline(context.Background(), time.Unix(1, 0))
				defer cancel()
			}
			work, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			oldRoot, oldCommand := rootCtx, cmdCtx
			oldConfig, oldDir, oldGeneration := graphPreviewConfig, graphPreviewDir, graphPreviewTypeGeneration
			oldJSON, oldOutput := jsonOutput, graphFailureOutput
			t.Cleanup(func() {
				rootCtx, cmdCtx = oldRoot, oldCommand
				graphPreviewConfig, graphPreviewDir, graphPreviewTypeGeneration = oldConfig, oldDir, oldGeneration
				jsonOutput, graphFailureOutput = oldJSON, oldOutput
			})
			t.Setenv("BEADS_DOLT_PASSWORD", "disposable-test-password")
			rootCtx, cmdCtx = ctx, nil
			graphPreviewDir, graphPreviewTypeGeneration = work, 0
			graphPreviewConfig = &configfile.Config{
				DoltMode: configfile.DoltModeServer, DoltDatabase: "iris_canceled_open",
				DoltServerSocket: filepath.Join(work, "absent.sock"), DoltServerUser: "root",
				GraphWorkspace: work, GraphScopeURL: "https://example.invalid/import-cancellation/",
				GraphAuthorityID: strings.Repeat("0", 32), GraphSchemaVersion: graphstore.SchemaVersion,
			}
			var diagnostic bytes.Buffer
			jsonOutput, graphFailureOutput = true, &diagnostic
			result := runGraphPreviewImportReader(strings.NewReader(`{"id":"legacy-a","title":"disposable"}`), "stdin")
			var failure *exitError
			if !errors.As(result, &failure) || failure.Code != 5 {
				t.Fatalf("import interruption=%v want exit5", result)
			}
			var envelope struct {
				Code      string `json:"code"`
				Message   string `json:"message"`
				Retryable bool   `json:"retryable"`
			}
			if err := json.Unmarshal(diagnostic.Bytes(), &envelope); err != nil || envelope.Code != "capability_unavailable" || envelope.Retryable || !strings.Contains(envelope.Message, ctx.Err().Error()) {
				t.Fatalf("import route did not classify one raw cancellation: %q (%v)", diagnostic.String(), err)
			}
		})
	}
}

func TestGraphPreviewStoreOpenUsesRawErrorMapper(t *testing.T) {
	oldConfig, oldJSON, oldOutput := graphPreviewConfig, jsonOutput, graphFailureOutput
	t.Cleanup(func() { graphPreviewConfig, jsonOutput, graphFailureOutput = oldConfig, oldJSON, oldOutput })
	graphPreviewConfig = &configfile.Config{} // Refused by option validation, before any connector.
	var diagnostic bytes.Buffer
	jsonOutput, graphFailureOutput = true, &diagnostic
	calls := 0
	mapped := errors.New("mapper result")
	result := withGraphStoreErrorMapper(time.Second, func(context.Context, *graphstore.Store) (any, string, error) {
		t.Fatal("operation ran after failed open")
		return nil, "", nil
	}, func(any, string) error {
		t.Fatal("success output ran after failed open")
		return nil
	}, func(raw error) error {
		calls++
		if !errors.Is(raw, graphstore.ErrInvalidStore) || diagnostic.Len() != 0 {
			t.Fatalf("mapper received reported/replaced error: %v diagnostic=%q", raw, diagnostic.String())
		}
		return mapped
	})
	if result != mapped || calls != 1 || diagnostic.Len() != 0 {
		t.Fatalf("mapper boundary: result=%v calls=%d diagnostic=%q", result, calls, diagnostic.String())
	}
}

func TestGraphPreviewStoreOpenDispatchesOnlyOpenMapper(t *testing.T) {
	oldConfig, oldJSON, oldOutput := graphPreviewConfig, jsonOutput, graphFailureOutput
	t.Cleanup(func() { graphPreviewConfig, jsonOutput, graphFailureOutput = oldConfig, oldJSON, oldOutput })
	graphPreviewConfig = &configfile.Config{} // Refused before dialing any connector.
	var diagnostic bytes.Buffer
	jsonOutput, graphFailureOutput = true, &diagnostic
	openCalls, operationCalls := 0, 0
	mapped := errors.New("open mapper result")
	result := withGraphStoreStageErrorMappers(time.Second, func(context.Context, *graphstore.Store) (any, string, error) {
		t.Fatal("operation ran after failed open")
		return nil, "", nil
	}, func(any, string) error {
		t.Fatal("success output ran after failed open")
		return nil
	}, func(raw error) error {
		openCalls++
		if !errors.Is(raw, graphstore.ErrInvalidStore) || diagnostic.Len() != 0 {
			t.Fatalf("open mapper received reported/replaced error: %v diagnostic=%q", raw, diagnostic.String())
		}
		return mapped
	}, func(error) error {
		operationCalls++
		return errors.New("operation mapper called for failed open")
	})
	if result != mapped || openCalls != 1 || operationCalls != 0 || diagnostic.Len() != 0 {
		t.Fatalf("stage dispatch: result=%v open=%d operation=%d diagnostic=%q", result, openCalls, operationCalls, diagnostic.String())
	}
}

func TestGraphPreviewLegacyImportSummaryPrefix(t *testing.T) {
	withPrefix := graphPreviewLegacyImportSummary("Imported", "source.jsonl", graphstore.LegacyImportResult{IssuePrefix: "web-app", SourceIssuePrefix: "web-app", Issues: 1})
	if !strings.Contains(withPrefix, "Workspace Issue prefix: web-app.") {
		t.Fatalf("adopted prefix not disclosed: %s", withPrefix)
	}
	preview := graphPreviewLegacyImportSummary("Would import", "stdin", graphstore.LegacyImportResult{IssuePrefix: "web-app", SourceIssuePrefix: "web-app", DryRun: true})
	if !strings.Contains(preview, "Would adopt Issue prefix: web-app.") || strings.Contains(preview, "Workspace Issue prefix:") {
		t.Fatalf("dry-run prefix adoption claimed: %s", preview)
	}
	rooted := graphPreviewLegacyImportSummary("Imported", "source.jsonl", graphstore.LegacyImportResult{IssuePrefix: "destination", SourceIssuePrefix: "web-app", IDRoot: "legacy/trial", Issues: 1})
	for _, want := range []string{"ID root: legacy/trial.", "Workspace Issue prefix remains destination.", "Source Issue prefix: web-app."} {
		if !strings.Contains(rooted, want) {
			t.Fatalf("rooted import lost source/destination placement %q: %s", want, rooted)
		}
	}
	memories := graphPreviewLegacyImportSummary("Would import", "stdin", graphstore.LegacyImportResult{Memories: 1})
	if strings.Contains(memories, "Issue prefix:") {
		t.Fatalf("memory-only prefix invented: %s", memories)
	}
}
