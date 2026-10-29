package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/canonicalbatch"
	"github.com/steveyegge/beads/internal/configfile"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// These tests exercise the real journal and authoring/admission functions without
// an engine. They do not establish a committed decision or post-commit recovery.
func canonicalRecoveryTestIsolation(t *testing.T) {
	t.Helper()
	stdioMutex.Lock()
	t.Cleanup(stdioMutex.Unlock)
	isolateBeadsDirForTest(t)
	oldCreate, oldBatch, oldArgs := createCmd, batchCmd, os.Args
	oldRetry := canonicalBatchRetry
	oldActive, oldConfig, oldDir := graphPreviewActive, graphPreviewConfig, graphPreviewDir
	oldOutput, oldReporter := graphFailureOutput, graphFailureReporter
	oldJSON, oldStructured := jsonOutput, graphPreviewStructuredErrors
	t.Cleanup(func() {
		createCmd, batchCmd, os.Args = oldCreate, oldBatch, oldArgs
		canonicalBatchRetry = oldRetry
		graphPreviewActive, graphPreviewConfig, graphPreviewDir = oldActive, oldConfig, oldDir
		graphFailureOutput, graphFailureReporter = oldOutput, oldReporter
		jsonOutput, graphPreviewStructuredErrors = oldJSON, oldStructured
	})
	canonicalBatchRetry = nil
	graphFailureOutput, graphFailureReporter = &bytes.Buffer{}, nil
	jsonOutput, graphPreviewStructuredErrors = false, false
}

func canonicalRecoveryTestCommand(t *testing.T, name string, args ...string) *cobra.Command {
	t.Helper()
	root := &cobra.Command{Use: "bd"}
	root.PersistentFlags().String("db", "", "")
	cmd := &cobra.Command{Use: name}
	root.AddCommand(cmd)
	for _, flag := range []string{"file", "input-format", "batch-version", "scope", "dry-run-mode", "idempotency-key", "retry-file", "actor", "type", "id", "batch-bind", "properties", "metadata", "revision-provenance"} {
		cmd.Flags().String(flag, "", "")
	}
	cmd.Flags().StringArray("batch-arg", nil, "")
	cmd.Flags().Bool("json", false, "")
	if err := cmd.ParseFlags(args); err != nil {
		t.Fatal(err)
	}
	os.Args = append([]string{"bd", name}, args...)
	return cmd
}

// A well-formed persisted graph binding with an explicitly unavailable route.
// No source is initialized: reaching provider acquisition would fail this test's
// intended offline boundary, rather than being hidden by corrupt metadata.
func canonicalRecoveryConfiguredWorkspace(t *testing.T, scope string) *configfile.Config {
	t.Helper()
	work, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(work, ".beads")
	if err := os.Mkdir(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	cfg := &configfile.Config{Backend: configfile.BackendDolt, Database: "beads.db", GraphMode: "link", GraphScopeURL: scope, GraphAuthorityID: "0123456789abcdef0123456789abcdef", GraphWorkspace: dir, GraphSchemaVersion: graphstore.AtomicBatchSchemaVersion, GraphReady: true, DoltMode: configfile.DoltModeServer, DoltServerHost: "127.0.0.1", DoltServerPort: 3307, DoltServerUser: "root", DoltServerSocket: filepath.Join(work, "unavailable.sock"), DoltDatabase: "canonical_recovery_test"}
	if err := cfg.Save(dir); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, graphPreviewMarker), []byte(graphPreviewGeneration), 0o600); err != nil {
		t.Fatal(err)
	}
	loaded, err := configfile.LoadForDiscovery(dir)
	if err != nil || !reflect.DeepEqual(loaded, cfg) {
		t.Fatalf("fixture lost its configured binding: %+v %v", loaded, err)
	}
	return cfg
}

func canonicalRecoveryTestInvocation(t *testing.T, workspace, scope string) canonicalBatchInvocation {
	t.Helper()
	request, err := canonicalbatch.Compile([]byte(`{"operation":"createBead","id":"original","type":"memory","properties":{"title":"Original","body":"kept"},"metadata":{"reviewed":true}}`), canonicalbatch.Options{Scope: scope})
	if err != nil {
		t.Fatal(err)
	}
	return canonicalBatchInvocation{Format: "canonical-batch-invocation-1", Workspace: workspace, Scope: scope, Authority: "0123456789abcdef0123456789abcdef", Database: "canonical_recovery_test", Schema: graphstore.AtomicBatchSchemaVersion, Key: "original-key", Actor: "original-author", Request: request.NativeJSON()}
}

func TestCanonicalBatchRecoveryJournalRoundTrip(t *testing.T) {
	workspace, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	want := canonicalRecoveryTestInvocation(t, workspace, "https://example.invalid/original/")
	file, err := saveCanonicalBatchInvocation(want)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Dir(file) != filepath.Join(workspace, "batch-invocations") {
		t.Fatalf("journal escaped selected workspace: %s", file)
	}
	got, err := readCanonicalBatchInvocation(file)
	if err != nil || !reflect.DeepEqual(*got, want) {
		t.Fatalf("frozen binding/actor/key/request changed: %+v %v", got, err)
	}
	// Saving identical input must validate and reuse the same durable file.
	again, err := saveCanonicalBatchInvocation(want)
	if err != nil || again != file {
		t.Fatalf("same invocation created a second journal: %s %v", again, err)
	}
	entries, err := os.ReadDir(filepath.Dir(file))
	if err != nil || len(entries) != 1 {
		t.Fatalf("journal entries: %v %v", entries, err)
	}
	got.Request[0] = '!'
	reread, err := readCanonicalBatchInvocation(file)
	if err != nil || !bytes.Equal(reread.Request, want.Request) {
		t.Fatalf("reader exposed mutable retained request: %+v %v", reread, err)
	}
	// The digest path must not silently overwrite a different retained body.
	if err := os.WriteFile(file, []byte(`{"damaged":true}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := saveCanonicalBatchInvocation(want); err == nil {
		t.Fatal("save overwrote a different retained invocation")
	}
	if _, err := readCanonicalBatchInvocation(file); err == nil {
		t.Fatal("damaged invocation admitted for recovery")
	}
}

func TestCanonicalBatchRecoveryAdmissionUsesFrozenInputs(t *testing.T) {
	canonicalRecoveryTestIsolation(t)
	original := canonicalRecoveryConfiguredWorkspace(t, "https://example.invalid/original/")
	foreign := canonicalRecoveryConfiguredWorkspace(t, "https://example.invalid/foreign/")
	want := canonicalRecoveryTestInvocation(t, original.GraphWorkspace, original.GraphScopeURL)
	file, err := saveCanonicalBatchInvocation(want)
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(filepath.Dir(foreign.GraphWorkspace))
	t.Setenv("BEADS_DIR", "")
	t.Setenv("BD_SCOPE", foreign.GraphScopeURL)
	t.Setenv("BEADS_ACTOR", "today-author")
	t.Setenv("BD_ACTOR", "another-author")
	cmd := canonicalRecoveryTestCommand(t, "batch", "--retry-file", file)
	batchCmd = cmd
	handled, err := admitCanonicalAuthoring(cmd)
	if handled || err != nil || canonicalBatchRetry == nil {
		t.Fatalf("retry admission: handled=%t, saved=%+v, error=%v", handled, canonicalBatchRetry, err)
	}
	if !reflect.DeepEqual(*canonicalBatchRetry, want) {
		t.Fatalf("today's environment changed frozen inputs: %+v", canonicalBatchRetry)
	}
	selected, err := graphCandidateDir(cmd)
	if err != nil || selected != original.GraphWorkspace {
		t.Fatalf("foreign cwd retargeted retry: %s %v", selected, err)
	}
	compiled, err := canonicalbatch.Compile(canonicalBatchRetry.Request, canonicalbatch.Options{Format: "json", Scope: canonicalBatchRetry.Scope})
	if err != nil || !bytes.Equal(compiled.NativeJSON(), want.Request) {
		t.Fatalf("saved native request changed on retry admission: %v", err)
	}
	if _, err := os.Stat(filepath.Join(foreign.GraphWorkspace, "batch-invocations")); !os.IsNotExist(err) {
		t.Fatal("retry admission wrote an invocation in the foreign workspace")
	}
	// Each override must refuse before trying to read an absent recovery file.
	missing := filepath.Join(t.TempDir(), "must-not-read.json")
	for _, override := range [][]string{{"--actor", "replacement"}, {"--scope", foreign.GraphScopeURL}, {"--file", "replacement.jsonl"}, {"--idempotency-key", "replacement"}} {
		cmd := canonicalRecoveryTestCommand(t, "batch", append([]string{"--retry-file", missing}, override...)...)
		batchCmd = cmd
		diagnostics := graphFailureOutput.(*bytes.Buffer)
		diagnostics.Reset()
		handled, err := admitCanonicalAuthoring(cmd)
		if status, ok := err.(*exitError); !handled || !ok || status.Code != 2 || canonicalBatchRetry != nil {
			t.Fatalf("retry override admitted: %v: handled=%t error=%v", override, handled, err)
		}
		if text := diagnostics.String(); !strings.Contains(text, override[0]) || !strings.Contains(text, "not supported") || strings.Contains(text, missing) {
			t.Fatalf("override did not refuse before recovery input access: %s", text)
		}
	}
}

func TestCanonicalBatchAuthoringWithUnreachableConfiguredSource(t *testing.T) {
	canonicalRecoveryTestIsolation(t)
	cfg := canonicalRecoveryConfiguredWorkspace(t, "https://example.invalid/offline/")
	t.Chdir(filepath.Dir(cfg.GraphWorkspace))
	t.Setenv("BEADS_DIR", cfg.GraphWorkspace)
	t.Setenv("BD_SCOPE", "")
	metadataPath := configfile.ConfigPath(cfg.GraphWorkspace)
	metadataBefore, err := os.ReadFile(metadataPath)
	if err != nil {
		t.Fatal(err)
	}
	flags := []string{"--type", "memory", "--id", "offline", "--properties", `{"title":"Offline","body":"kept"}`, "--metadata", `{"reviewed":true}`, "--dry-run-mode", "batch"}
	cmd := canonicalRecoveryTestCommand(t, "create", flags...)
	createCmd = cmd
	var member bytes.Buffer
	cmd.SetOut(&member)
	graphPreviewActive, graphPreviewConfig, graphPreviewDir = true, cfg, cfg.GraphWorkspace
	if handled, err := admitCanonicalAuthoring(cmd); !handled || err != nil || graphPreviewActive || graphPreviewConfig != nil || graphPreviewDir != "" {
		t.Fatalf("member did not take offline admission: handled=%t error=%v", handled, err)
	}
	if err := runCanonicalMemberEmission(cmd, nil); err != nil {
		t.Fatal(err)
	}
	parsed, _, err := canonicalbatch.CompileMember(member.Bytes(), cfg.GraphScopeURL)
	if err != nil || parsed.Type.URI != cfg.GraphScopeURL+"memory" || string(parsed.Metadata) != `{"reviewed":true}` {
		t.Fatalf("offline member lost configured Scope or metadata: %+v %v", parsed, err)
	}
	input := filepath.Join(filepath.Dir(cfg.GraphWorkspace), "members.jsonl")
	if err := os.WriteFile(input, member.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	cmd = canonicalRecoveryTestCommand(t, "batch", "--file", input, "--dry-run-mode", "batch")
	batchCmd = cmd
	var output bytes.Buffer
	cmd.SetOut(&output)
	if handled, err := admitCanonicalAuthoring(cmd); !handled || err != nil {
		t.Fatalf("wrapper did not take offline admission: handled=%t error=%v", handled, err)
	}
	if err := runCanonicalBatch(cmd); err != nil {
		t.Fatal(err)
	}
	request, err := canonicalbatch.Compile(output.Bytes(), canonicalbatch.Options{Format: "json", Scope: cfg.GraphScopeURL})
	if err != nil || len(request.Members()) != 1 || request.Members()[0].Type.URI != parsed.Type.URI {
		t.Fatalf("wrapper changed emitted member: %v", err)
	}
	metadataAfter, err := os.ReadFile(metadataPath)
	if err != nil || !bytes.Equal(metadataBefore, metadataAfter) {
		t.Fatal("offline authoring modified configured workspace metadata")
	}
	entries, err := os.ReadDir(cfg.GraphWorkspace)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.Name() != configfile.ConfigFileName && entry.Name() != graphPreviewMarker {
			t.Fatalf("offline authoring created provider/journal state: %s", entry.Name())
		}
	}
}

// An accepted request must remain recoverable when an older invocation writer
// HTML-escaped its RawMessage, multiplying physical bytes without changing data.
func TestCanonicalBatchRecoveryHTMLExpansionAndLogicalLimit(t *testing.T) {
	workspace, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	const scope = "https://example.invalid/html-recovery/"
	// About 1.2 MiB of accepted content expands beyond the old 4 MiB journal
	// read bound. Keep the request well below the compiler's own input limit.
	body := strings.Repeat("<>&", 400<<10)
	authored := []byte(`{"operation":"createBead","id":"html","type":"memory","properties":{"title":"HTML","body":"` + body + `"}}`)
	batch, err := canonicalbatch.Compile(authored, canonicalbatch.Options{Scope: scope})
	if err != nil {
		t.Fatalf("HTML-rich request must be admitted before testing recovery: %v", err)
	}
	want := canonicalRecoveryTestInvocation(t, workspace, scope)
	want.Request = batch.NativeJSON()
	const canonicalLimit = canonicalbatch.MaxInputBytes + 32768
	if len(want.Request) >= canonicalbatch.MaxInputBytes {
		t.Fatal("regression accidentally exceeds the accepted request budget")
	}
	file, err := saveCanonicalBatchInvocation(want)
	if err != nil {
		t.Fatal(err)
	}
	got, err := readCanonicalBatchInvocation(file)
	if err != nil || !reflect.DeepEqual(*got, want) {
		t.Fatalf("accepted HTML-rich invocation is not recoverable unchanged: %v", err)
	}
	physical, err := os.ReadFile(file)
	if err != nil || len(physical) > canonicalLimit {
		t.Fatalf("new journal unnecessarily expanded admitted input: %d bytes, %v", len(physical), err)
	}
	legacy, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	if len(legacy) <= canonicalLimit {
		t.Fatalf("legacy fixture did not exercise HTML expansion: %d bytes", len(legacy))
	}
	legacyFile := filepath.Join(workspace, "legacy-escaped-invocation.json")
	if err := os.WriteFile(legacyFile, legacy, 0o600); err != nil {
		t.Fatal(err)
	}
	got, err = readCanonicalBatchInvocation(legacyFile)
	if err != nil || !reflect.DeepEqual(*got, want) {
		t.Fatalf("retained escaped invocation changed or was refused: %v", err)
	}
	// A larger physical allowance must not remove the logical envelope bound.
	// The request itself is already admitted; the excess is envelope data, so
	// request compilation alone would not catch a missing logical-size check.
	oversized := canonicalRecoveryTestInvocation(t, workspace, scope)
	oversized.Actor = strings.Repeat("x", canonicalLimit)
	raw, err := json.Marshal(oversized)
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) <= canonicalLimit || len(raw) > 6*canonicalLimit {
		t.Fatal("oversize control is outside the intended physical/logical window")
	}
	if _, err := saveCanonicalBatchInvocation(oversized); err == nil || !strings.Contains(err.Error(), "exceeds input limit") {
		t.Fatalf("save accepted a logically oversized invocation: %v", err)
	}
	oversizedFile := filepath.Join(workspace, "logically-oversized-invocation.json")
	if err := os.WriteFile(oversizedFile, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := readCanonicalBatchInvocation(oversizedFile); err == nil || !strings.Contains(err.Error(), "exceeds input limit") {
		t.Fatalf("logically oversized invocation was not bounded: %v", err)
	}
}
