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
	"strings"
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
	command := func(ctx context.Context, args ...string) *exec.Cmd {
		cmd := exec.CommandContext(ctx, bd, append(args, "--json")...)
		cmd.Dir = work
		cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + home, "TMPDIR=" + os.TempDir(), "XDG_CONFIG_HOME=" + filepath.Join(home, ".config"), "XDG_CACHE_HOME=" + filepath.Join(home, ".cache"), "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=" + filepath.Join(home, "missing-gitconfig"), "BD_DISABLE_METRICS=1", "BD_DISABLE_EVENT_FLUSH=1", "BD_NON_INTERACTIVE=1", "BEADS_DOLT_AUTO_START=0", "DOLT_METRICS_DISABLED=1", "NO_COLOR=1"}
		return cmd
	}
	run := func(failure string, args ...string) graphstore.BatchResult {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()
		cmd := command(ctx, args...)
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

	// Selection and set mutation cross the real CLI/compiler/provider path in
	// one invocation; the CLI does not pre-list or expand targets itself.
	write(`{"operation":"select","name":"chosen","query":{"ir":"bdp-query-design-1","profile":"list","selection":{"op":"filter","input":{"op":"scan","collection":"beads"},"bind":"b","predicate":{"op":"compare","operator":"eq","left":{"op":"field","of":{"op":"binding","name":"b"},"path":["properties","title"]},"right":{"op":"literal","value":"Changed"}}}}}` + "\n" +
		emit("update", "batch-ref::chosen", "--set-property", "body=Selected atomically"))
	selected := run("", "batch", "--file", file, "--idempotency-key", "selection")
	if !selected.Changed || len(selected.Results) != 2 || !reflect.DeepEqual(selected.Results[0].IDs, []string{scope + "first"}) || !reflect.DeepEqual(selected.Results[0].IDs, selected.Results[1].IDs) {
		t.Fatalf("selection was not executed with the writes: %+v", selected)
	}
	write(`{"operation":"select","name":"empty","query":{"ir":"bdp-query-design-1","profile":"list","selection":{"op":"filter","input":{"op":"scan","collection":"beads"},"bind":"b","predicate":false}}}` + "\n" +
		emit("delete", "batch-ref::empty", "--force"))
	empty := run("", "batch", "--file", file, "--idempotency-key", "empty-selection")
	if empty.Changed || empty.Transaction != "" || len(empty.Resources) != 0 || len(empty.Results) != 2 || empty.Results[0].IDs == nil || empty.Results[1].IDs == nil || len(empty.Results[1].IDs) != 0 {
		t.Fatalf("empty set did not retain complete member results: %+v", empty)
	}
	emptyReplay := run("", "batch", "--file", file, "--idempotency-key", "empty-selection")
	if !reflect.DeepEqual(empty, emptyReplay) {
		t.Fatalf("empty set decision was reexecuted or lost: %+v %+v", empty, emptyReplay)
	}

	// The native Issue stays in its authoritative row/version plane. Two batch
	// edits publish its final state once rather than creating a generic stand-in.
	native := graphMixedResult[graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "create", "Native", "--type", "task", "--id", "native", "--force", "--json"))
	write(emit("update", "native", "--set-property", "title=Temporary") +
		emit("update", "native", "--set-property", "title=Published", "--set-metadata", "checked=true"))
	nativeResult := run("", "batch", "--file", file, "--idempotency-key", "native-content", "--actor", "test-author")
	if !nativeResult.Changed || len(nativeResult.Resources) != 1 || len(nativeResult.Results) != 2 {
		t.Fatalf("native batch returned incomplete results: %+v", nativeResult)
	}
	var published graphstore.IssueRecord
	if err := json.Unmarshal(nativeResult.Resources[0].Resource, &published); err != nil {
		t.Fatal(err)
	}
	shown := graphMixedResult[graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "show", native.ID, "--format", "graph-json", "--json"))
	// graphProjectCompleteRecords intentionally removes the legacy version
	// alias from CLI output. The experimental batch retains the native record;
	// compare every field after removing only that checked redundant alias.
	if shown.Version != "" || published.Version != published.Revision {
		t.Fatalf("unexpected native/CLI revision aliases: %+v %+v", published, shown)
	}
	publishedCLI := published
	publishedCLI.Version = ""
	if published.Type != native.Type || published.Properties.Title != "Published" || published.Attribution.Actor != "test-author" || published.Attribution.Status != "claimed" || !reflect.DeepEqual(shown, publishedCLI) {
		t.Fatalf("batch result diverged from native ordinary show: %+v %+v", published, shown)
	}
	// Historical native content uses the existing adapter and the same new
	// canonical object flags, independently of batch's selected member grammar.
	copied := graphMixedResult[graphstore.IssueMutationResult](t, graphPolicyCLI(t, bd, work, home, nil, "", "update", native.ID, "--actor", "test-author", "--from-revision", "="+native.Revision,
		"--set-property", "title=Historical edited", "--patch-metadata", `[{"op":"add","path":"/copied","value":true}]`, "--if-revision", "="+published.Revision, "--json"))
	if !copied.Changed || copied.Issue.Type != native.Type || copied.Issue.Properties.Title != "Historical edited" || string(copied.Issue.Metadata) != `{"copied":true}` {
		t.Fatalf("native copy lost canonical edit intent: %+v", copied)
	}
	// Native Link CRUD uses the same emitted members and real batch route.
	// The concrete Type and native dep_type agree; no generic Link is substituted.
	nativeTarget := graphMixedResult[graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "create", "Native target", "--type", "task", "--id", "native-target", "--force", "--json"))
	write(emit("link", "create", native.ID, nativeTarget.ID, "--type", "blocks", "--id", "native-batch-link", "--properties", `{"dep_type":"blocks"}`))
	createdNativeLink := run("", "batch", "--file", file, "--idempotency-key", "native-link-create", "--actor", "test-author")
	var batchLink graphstore.LinkRecord
	for _, row := range createdNativeLink.Resources {
		if row.ID == scope+"native-batch-link" {
			if err := json.Unmarshal(row.Resource, &batchLink); err != nil {
				t.Fatal(err)
			}
		}
	}
	shownLink := graphMixedResult[graphstore.LinkRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "link", "show", scope+"native-batch-link", "--json"))
	if batchLink.Revision == "" || batchLink.Version != batchLink.Revision || shownLink.Version != "" {
		t.Fatalf("unexpected native Link revision aliases: %+v %+v", batchLink, shownLink)
	}
	batchLink.Version = ""
	if !reflect.DeepEqual(shownLink, batchLink) || batchLink.Properties["dep_type"] != "blocks" {
		t.Fatalf("native batch Link differs from ordinary show: %+v %+v", batchLink, shownLink)
	}
	write(emit("link", "update", scope+"native-batch-link", "--set-metadata", "review=final"))
	updatedNativeLink := run("", "batch", "--file", file, "--idempotency-key", "native-link-metadata", "--actor", "test-author")
	if !updatedNativeLink.Changed {
		t.Fatal("native Link metadata command did not apply")
	}
	shownLink = graphMixedResult[graphstore.LinkRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "link", "show", scope+"native-batch-link", "--json"))
	if string(shownLink.Metadata) != `{"review":"final"}` || shownLink.Revision == batchLink.Revision {
		t.Fatalf("native metadata not final-published: %+v", shownLink)
	}
	write(emit("link", "delete", scope+"native-batch-link", "--force"))
	deletedNativeLink := run("", "batch", "--file", file, "--idempotency-key", "native-link-delete", "--actor", "test-author")
	var sawDeleted bool
	for _, row := range deletedNativeLink.Resources {
		if row.ID == scope+"native-batch-link" {
			sawDeleted = row.Outcome == "deleted" && row.Deleted != nil
		}
	}
	if !deletedNativeLink.Changed || !sawDeleted {
		t.Fatalf("native Link delete result incomplete: %+v", deletedNativeLink)
	}

	// A read-only stdout descriptor forces a real write error after confirmed
	// commit, without SIGPIPE or a production test hook. Recovery must return
	// that committed decision rather than execute the create a second time.
	write(emit("create", "--id", "lost-output", "--type", "memory", "--properties", `{"title":"Committed without output","body":""}`))
	stdout, err := os.Open(file)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	lost := command(ctx, "batch", "--file", file, "--idempotency-key", "lost-output")
	var diagnostics bytes.Buffer
	lost.Stdout, lost.Stderr = stdout, &diagnostics
	err = lost.Run()
	_ = stdout.Close()
	if ctx.Err() != nil || err == nil || !strings.Contains(diagnostics.String(), "Batch committed; recover its retained result") {
		t.Fatalf("did not observe confirmed commit followed by failed output: %v context=%v stderr=%s", err, ctx.Err(), diagnostics.String())
	}
	journals, err = filepath.Glob(filepath.Join(work, ".beads", "batch-invocations", "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	var recovery string
	for _, journal := range journals {
		invocation, readErr := readCanonicalBatchInvocation(journal)
		if readErr != nil {
			t.Fatal(readErr)
		}
		if invocation.Key == "lost-output" {
			recovery = journal
		}
	}
	if recovery == "" {
		t.Fatal("confirmed commit has no saved recovery invocation")
	}
	if err := os.Remove(file); err != nil {
		t.Fatal(err)
	}
	recovered := run("", "batch", "--retry-file", recovery)
	if !recovered.Changed || len(recovered.Resources) != 1 || recovered.Resources[0].ID != scope+"lost-output" || recovered.Resources[0].Outcome != "created" {
		t.Fatalf("lost output did not recover committed result: %+v", recovered)
	}
	if again := run("", "batch", "--retry-file", recovery); !reflect.DeepEqual(recovered, again) {
		t.Fatalf("recovery changed final revision or result: %+v %+v", recovered, again)
	}

}
