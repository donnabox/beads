//go:build cgo

package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// These are actual command front doors, deliberately run without any workspace.
// A provider-dependent command cannot accidentally pass this workflow.
func TestCanonicalMemberEmissionWorkflow(t *testing.T) {
	isolateBeadsDirForTest(t)
	bd := buildBDUnderTest(t)
	work, home := t.TempDir(), t.TempDir()
	const scope = "https://example.invalid/canonical-batch/"
	emit := func(args ...string) json.RawMessage {
		t.Helper()
		out := graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--scope", scope, "--dry-run-mode", "batch")...)
		var member map[string]json.RawMessage
		if err := json.Unmarshal([]byte(out), &member); err != nil {
			t.Fatal(err)
		}
		if len(member["operation"]) == 0 {
			t.Fatalf("missing member discriminator: %s", out)
		}
		return json.RawMessage(out)
	}
	a := emit("create", "--type", "memory", "--id", "note-a", "--batch-bind", "first", "--properties", `{"title":"A","body":"Alpha"}`)
	b := emit("create", "--type", "memory", "--id", "note-b", "--batch-bind", "second", "--properties", `{"title":"B","body":"Beta"}`)
	rel := emit("link", "create", "batch-ref::first", "batch-ref::second", "--type", "references", "--id", "relation")
	edit := emit("update", "note-a", "--set-property", "a/b~c=42", "--unset-property", "missing", "--patch-metadata", "[]", "--if-revision", "=baseline")
	var edited struct {
		Properties []struct{ Path string } `json:"propertiesChange"`
		Expected   string                  `json:"expectedRevision"`
		Metadata   json.RawMessage         `json:"metadataChange"`
	}
	if err := json.Unmarshal(edit, &edited); err != nil {
		t.Fatal(err)
	}
	if edited.Properties[0].Path != "/a~1b~0c" || edited.Expected != "baseline" || string(edited.Metadata) != "[]" {
		t.Fatalf("lost authored intent: %s", edit)
	}
	emit("delete", "note-a", "--force")
	emit("link", "update", "relation", "--replace-properties", `{}`, "--set-metadata", "reviewed=true")
	emit("link", "delete", "relation", "--force")
	file := filepath.Join(work, "members.jsonl")
	if err := os.WriteFile(file, append(append(a, b...), rel...), 0o600); err != nil {
		t.Fatal(err)
	}
	out := graphPolicyCLI(t, bd, work, home, nil, "", "batch", "--file", file, "--scope", scope, "--dry-run-mode", "batch")
	var request struct {
		Version    string            `json:"capabilityVersion"`
		Operations []json.RawMessage `json:"operations"`
	}
	if err := json.Unmarshal([]byte(out), &request); err != nil {
		t.Fatal(err)
	}
	if request.Version != "2" || len(request.Operations) != 3 {
		t.Fatalf("bad compiled request: %s", out)
	}
	for _, args := range [][]string{
		{"delete", "note-a", "--scope", scope, "--dry-run-mode", "batch"},
		{"create", "--type", "memory", "--force", "--scope", scope, "--dry-run-mode", "batch"},
		{"link", "create", "--from", "note-a", "--from", "note-b", "--to", "note-b", "--type", "references", "--scope", scope, "--dry-run-mode", "batch"},
		{"batch", "--file", file, "--dry-run", "--scope", scope, "--dry-run-mode", "batch"},
		{"batch", "--file", file, "--scope", scope, "--dry-run-mode", "batch", "--omit", "envelope"},
		{"batch", "--file", file, "--scope", scope, "--scope", "https://other.invalid/", "--dry-run-mode", "batch"},
		{"create", "--type", "memory", "--properties", `{"title":"First"}`, "--properties", `{"title":"Discarded first"}`, "--scope", scope, "--dry-run-mode", "batch"},
		{"update", "note-a", "--patch-metadata", "[]", "--patch-metadata", `[{"op":"add","path":"/x","value":1}]`, "--scope", scope, "--dry-run-mode", "batch"},
	} {
		graphPolicyCLI(t, bd, work, home, nil, "stderr:invalid_selector:", args...)
	}
	// Real root admission must reject conflicting execution actors before
	// workspace/provider discovery, and reject ignored emission/retry inputs.
	graphPolicyCLI(t, bd, work, home, nil, "stderr:conflicting --actor values", "batch", "--file", file, "--scope", scope, "--actor", "first", "--actor", "second")
	graphPolicyCLI(t, bd, work, home, nil, "stderr:--actor is not supported in canonical authoring", "batch", "--file", file, "--scope", scope, "--dry-run-mode", "batch", "--actor", "ignored")
	graphPolicyCLI(t, bd, work, home, nil, "stderr:--actor is not supported in canonical authoring", "batch", "--retry-file", filepath.Join(work, "must-not-read.json"), "--actor", "override")
	if _, err := os.Stat(filepath.Join(work, ".beads")); !os.IsNotExist(err) {
		t.Fatal("offline authoring initialized a workspace")
	}
}

// Execution consumes the actor as frozen invocation input. Emission has no
// execution identity, and retry must use its already-frozen actor.
func TestCanonicalBatchActorFlagAdmission(t *testing.T) {
	saved, savedOutput, savedReporter := canonicalBatchRetry, graphFailureOutput, graphFailureReporter
	canonicalBatchRetry, graphFailureReporter = nil, nil
	var diagnostics bytes.Buffer
	graphFailureOutput = &diagnostics
	t.Cleanup(func() {
		canonicalBatchRetry, graphFailureOutput, graphFailureReporter = saved, savedOutput, savedReporter
	})
	for _, tc := range []struct{ mode, want string }{
		{"", "canonical batch requires --file"},
		{"batch", "--actor is not supported in canonical authoring"},
	} {
		cmd := &cobra.Command{Use: "batch"}
		cmd.Flags().String("actor", "", "")
		cmd.Flags().String("dry-run-mode", "", "")
		cmd.Flags().String("file", "", "")
		args := []string{"--actor", "test-author"}
		if tc.mode != "" {
			args = append(args, "--dry-run-mode", tc.mode)
		}
		if err := cmd.ParseFlags(args); err != nil {
			t.Fatal(err)
		}
		diagnostics.Reset()
		err := runCanonicalBatch(cmd)
		status, ok := err.(*exitError)
		if !ok || status.Code != 2 || !strings.Contains(diagnostics.String(), tc.want) {
			t.Fatalf("mode %q: wanted %q and status2, got %v: %s", tc.mode, tc.want, err, diagnostics.String())
		}
	}
}
