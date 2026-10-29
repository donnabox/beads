//go:build cgo

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// This workflow exercises separate CLI processes and a fresh embedded workspace.
// The saved invocation must remain useful after its input file disappears and a
// subsequent write changes the live graph.
func TestCanonicalBatchCLIWorkflow(t *testing.T) {
	isolateBeadsDirForTest(t)
	bd := buildBDUnderTest(t)
	work, home := t.TempDir(), t.TempDir()
	const scope = "https://example.invalid/batch-workflow/"
	graphPolicyCLI(t, bd, work, home, nil, "", "init", "--graph-mode", "link", "--atomic-batch", "--scope-url", scope, "--skip-hooks", "--skip-agents", "--non-interactive", "--json")
	emit := func(args ...string) string {
		t.Helper()
		return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--dry-run-mode", "batch")...)
	}
	file := filepath.Join(work, "changes.jsonl")
	write := func(body string) {
		t.Helper()
		if err := os.WriteFile(file, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write(emit("create", "--id", "first", "--type", "memory", "--batch-bind", "a", "--properties", `{"title":"First","body":"Original"}`) +
		emit("create", "--id", "second", "--type", "memory", "--batch-bind", "b", "--properties", `{"title":"Second","body":"Target"}`) +
		emit("link", "create", "batch-ref::a", "batch-ref::b", "--id", "relation", "--type", "references"))
	run := func(failure string, args ...string) graphstore.BatchResult {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, bd, append(args, "--json")...)
		cmd.Dir = work
		cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + home, "TMPDIR=" + os.TempDir(), "XDG_CONFIG_HOME=" + filepath.Join(home, ".config"), "XDG_CACHE_HOME=" + filepath.Join(home, ".cache"), "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=" + filepath.Join(home, "missing-gitconfig"), "BD_DISABLE_METRICS=1", "BD_DISABLE_EVENT_FLUSH=1", "BD_NON_INTERACTIVE=1", "BEADS_DOLT_AUTO_START=0", "DOLT_METRICS_DISABLED=1", "NO_COLOR=1"}
		var out, diagnostics bytes.Buffer
		cmd.Stdout, cmd.Stderr = &out, &diagnostics
		err := cmd.Run()
		if ctx.Err() != nil {
			t.Fatalf("batch CLI exceeded bound: %v\n%s", ctx.Err(), diagnostics.String())
		}
		var envelope struct {
			Outcome string                 `json:"outcome"`
			Data    graphstore.BatchResult `json:"data"`
			Error   struct {
				Problem struct{ Code string } `json:"problem"`
			} `json:"error"`
		}
		if decodeErr := json.Unmarshal(out.Bytes(), &envelope); decodeErr != nil {
			t.Fatalf("invalid batch output: %v\n%s\n%s", decodeErr, out.String(), diagnostics.String())
		}
		if failure != "" {
			if err == nil || envelope.Outcome != "failure" || envelope.Error.Problem.Code != failure {
				t.Fatalf("expected %s: err=%v stdout=%s stderr=%s", failure, err, out.String(), diagnostics.String())
			}
			return graphstore.BatchResult{}
		}
		if err != nil || envelope.Outcome != "success" {
			t.Fatalf("batch failed: %v\n%s\n%s", err, out.String(), diagnostics.String())
		}
		return envelope.Data
	}
	seed := run("", "batch", "--file", file, "--idempotency-key", "original")
	if !seed.Changed || len(seed.Results) != 3 || len(seed.Resources) != 3 {
		t.Fatalf("incomplete graph publication: %+v", seed)
	}
	journals, err := filepath.Glob(filepath.Join(work, ".beads", "batch-invocations", "*.json"))
	if err != nil || len(journals) != 1 {
		t.Fatalf("missing saved invocation: %v %v", journals, err)
	}
	var original graphstore.Record
	for _, row := range seed.Resources {
		if row.ID == scope+"first" {
			if err := json.Unmarshal(row.Resource, &original); err != nil {
				t.Fatal(err)
			}
		}
	}
	if original.Revision == "" {
		t.Fatal("missing real published revision")
	}
	write(emit("update", "first", "--set-property", "title=Changed"))
	run("", "batch", "--file", file, "--idempotency-key", "change")
	write(emit("create", "--id", "must-rollback", "--type", "memory", "--properties", `{"title":"Rollback","body":""}`) + emit("update", "first", "--if-revision", "="+original.Revision, "--set-property", "title=Stale"))
	run("revision_mismatch", "batch", "--file", file, "--idempotency-key", "stale")
	graphPolicyCLI(t, bd, work, home, nil, "not_found", "show", "must-rollback", "--json")
	write(emit("update", "first", "--patch-properties", "[]", "--patch-metadata", "[]"))
	noop := run("", "batch", "--file", file, "--idempotency-key", "noop")
	if noop.Changed || noop.Transaction != "" {
		t.Fatalf("no-op minted publication: %+v", noop)
	}
	if err := os.Remove(file); err != nil {
		t.Fatal(err)
	}
	replay := run("", "batch", "--retry-file", journals[0])
	if !reflect.DeepEqual(seed, replay) {
		t.Fatalf("recovery changed the original result: %+v %+v", seed, replay)
	}
}
