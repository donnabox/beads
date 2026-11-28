package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

func TestCanonicalMutationInvocationOwnsArgs(t *testing.T) {
	cmd := &cobra.Command{Use: "update"}
	checked := []string{"update", "one", "--if-revision==original"}
	restore, err := bindCanonicalMutationInvocationArgs(cmd, context.Background(), checked)
	if err != nil {
		t.Fatal(err)
	}
	defer restore()
	checked[2] = "--if-revision==substituted"
	process := []string{"--unknown-prefix", "--if-revision==other", "--force=false", "--force=true"}
	got, err := canonicalMutationInvocationArgsFrom(cmd, process)
	if err != nil || !slices.Equal(got, []string{"update", "one", "--if-revision==original"}) {
		t.Fatalf("ambient or borrowed argv replaced checked invocation: %v %v", got, err)
	}
	got[2] = "--if-revision==modified-copy"
	again, err := canonicalMutationInvocationArgsFrom(cmd, process)
	if err != nil || again[2] != "--if-revision==original" {
		t.Fatalf("accessor exposed shared argument storage: %v %v", again, err)
	}
}

func TestCanonicalMutationInvocationEmptyAndCompatFallback(t *testing.T) {
	cmd := &cobra.Command{Use: "batch"}
	process := []string{"--unknown-prefix", "legacy"}
	got, err := canonicalMutationInvocationArgsFrom(cmd, process)
	if err != nil || !slices.Equal(got, process) {
		t.Fatalf("unbound compatibility argv changed: %v %v", got, err)
	}
	got[0] = "changed"
	if process[0] != "--unknown-prefix" {
		t.Fatal("compatibility accessor borrowed process argument storage")
	}
	restore, err := bindCanonicalMutationInvocationArgs(cmd, context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	got, err = canonicalMutationInvocationArgsFrom(cmd, process)
	if err != nil || len(got) != 0 {
		t.Fatalf("bound empty argv fell back to process input: %v %v", got, err)
	}
	restore()
	restore()
	got, err = canonicalMutationInvocationArgsFrom(cmd, process)
	if err != nil || !slices.Equal(got, process) || cmd.Context() != nil {
		t.Fatalf("restore changed compatibility behavior or original nil context: %v %v", got, err)
	}
}

func TestCanonicalMutationInvocationPreservesRootContext(t *testing.T) {
	root, cmd := &cobra.Command{Use: "bd"}, &cobra.Command{Use: "create"}
	root.AddCommand(cmd)
	execution, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute))
	defer cancel()
	root.SetContext(execution)
	if cmd.Context() != nil {
		t.Fatal("fixture did not exercise Cobra's pre-Execute nil leaf context")
	}
	restore, err := bindCanonicalMutationInvocationArgs(cmd, root.Context(), []string{"create", "--type=memory"})
	if err != nil {
		t.Fatal(err)
	}
	wantDeadline, _ := execution.Deadline()
	if deadline, ok := cmd.Context().Deadline(); !ok || !deadline.Equal(wantDeadline) {
		t.Fatal("binding detached the root deadline")
	}
	cancel()
	if !errors.Is(cmd.Context().Err(), context.Canceled) {
		t.Fatal("binding detached root cancellation")
	}
	restore()
	if cmd.Context() != nil || root.Context() != execution {
		t.Fatal("restoration changed original leaf/root context")
	}
	if _, err := bindCanonicalMutationInvocationArgs(cmd, nil, []string{"create"}); err == nil || cmd.Context() != nil {
		t.Fatal("binding manufactured Background before actual execution context was available")
	}
}

func TestCanonicalMutationInvocationPreservesExplicitLeafContext(t *testing.T) {
	type contextMarker struct{}
	cmd := &cobra.Command{Use: "create"}
	leaf := context.WithValue(context.Background(), contextMarker{}, "leaf")
	cmd.SetContext(leaf)
	root, cancel := context.WithCancel(context.Background())
	cancel()
	restore, err := bindCanonicalMutationInvocationArgs(cmd, root, []string{"create"})
	if err != nil {
		t.Fatal(err)
	}
	if cmd.Context().Value(contextMarker{}) != "leaf" || cmd.Context().Err() != nil {
		t.Fatal("existing leaf context was replaced by root context")
	}
	restore()
	if cmd.Context() != leaf {
		t.Fatal("explicit original leaf context was not restored exactly")
	}
}

func TestCanonicalMutationInvocationRejectsInvalidBindings(t *testing.T) {
	cmd, other := &cobra.Command{Use: "create"}, &cobra.Command{Use: "delete"}
	for _, value := range []any{"malformed", canonicalMutationInvocation{command: other, args: []string{"delete", "other", "--force"}}} {
		cmd.SetContext(context.WithValue(context.Background(), canonicalMutationInvocationKey{}, value))
		got, err := canonicalMutationInvocationArgsFrom(cmd, []string{"create", "--type=memory"})
		var refusal *canonicalMutationLeafRefusal
		if !errors.As(err, &refusal) || refusal.ExitCode != 5 || got != nil {
			t.Fatalf("malformed/inherited binding used a compatibility fallback: %v %v", got, err)
		}
	}
}

func TestCanonicalMutationInvocationRestoresEveryExecuteExit(t *testing.T) {
	root, cmd := &cobra.Command{Use: "bd", SilenceErrors: true, SilenceUsage: true}, &cobra.Command{Use: "work"}
	root.AddCommand(cmd)
	root.SetContext(context.Background())
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	failure := errors.New("selected execution failure")
	for _, phase := range []string{"parse", "args", "pre-run", "run", "success"} {
		t.Run(phase, func(t *testing.T) {
			checked := []string{"work", phase}
			if phase == "parse" {
				checked = []string{"work", "--unknown"}
			}
			cmd.Args = func(*cobra.Command, []string) error {
				if phase == "args" {
					return failure
				}
				return nil
			}
			root.PersistentPreRunE = func(*cobra.Command, []string) error {
				if phase == "pre-run" {
					return failure
				}
				return nil
			}
			cmd.RunE = func(*cobra.Command, []string) error {
				if phase == "run" {
					return failure
				}
				return nil
			}
			err := func() error {
				restore, err := bindCanonicalMutationInvocationArgs(cmd, root.Context(), checked)
				if err != nil {
					return err
				}
				defer restore()
				root.SetArgs(checked)
				_, err = root.ExecuteC()
				return err
			}()
			if (phase == "success") != (err == nil) {
				t.Fatalf("unexpected execution result: %v", err)
			}
			got, readErr := canonicalMutationInvocationArgsFrom(cmd, []string{"compat-after-" + phase})
			if cmd.Context() != nil || readErr != nil || !slices.Equal(got, []string{"compat-after-" + phase}) {
				t.Fatalf("invocation leaked across an exit/reused command: %v %v", got, readErr)
			}
		})
	}
}

// Real existing handlers consume the captured arguments. No global os.Args
// mutation, fake provider or counter-only mutation executor is used.
func TestCanonicalMutationBoundHandlersUseCheckedArgs(t *testing.T) {
	stdioMutex.Lock()
	defer stdioMutex.Unlock()
	originalProcess := append([]string{}, os.Args...)
	oldCreate, oldBatch := createCmd, batchCmd
	oldActive, oldConfig, oldDir := graphPreviewActive, graphPreviewConfig, graphPreviewDir
	oldRetry, oldReporter := canonicalBatchRetry, graphFailureReporter
	defer func() {
		createCmd, batchCmd = oldCreate, oldBatch
		graphPreviewActive, graphPreviewConfig, graphPreviewDir = oldActive, oldConfig, oldDir
		canonicalBatchRetry, graphFailureReporter = oldRetry, oldReporter
	}()
	graphFailureReporter = func(code, message string, exit int) error {
		return &canonicalMutationLeafRefusal{Code: code, Message: message, ExitCode: exit}
	}
	for _, operation := range []string{"create", "batch"} {
		t.Run(operation, func(t *testing.T) {
			root, cmd := canonicalMutationLeafFixture(t, []string{operation})
			root.SetContext(context.Background())
			var out bytes.Buffer
			cmd.SetOut(&out)
			checked := []string{operation, "--scope=" + canonicalLinkLoweringScope, "--dry-run-mode=batch"}
			if operation == "create" {
				createCmd = cmd
				checked = append(checked, "--type=memory", "--properties={\"title\":\"Created\",\"body\":\"Body\"}")
				cmd.RunE = runCanonicalMemberEmission
			} else {
				batchCmd = cmd
				checked = append(checked, "--file=-")
				cmd.SetIn(strings.NewReader("{\"operation\":\"createBead\",\"type\":\"memory\",\"properties\":{\"title\":\"Created\"}}\n"))
				cmd.RunE = func(cmd *cobra.Command, _ []string) error { return runCanonicalBatch(cmd) }
			}
			cmd.Args = canonicalMutationLeafArgs(cmd, []string{operation}, checked)
			admitted := false
			root.PersistentPreRunE = func(actual *cobra.Command, _ []string) error {
				var err error
				admitted, err = admitCanonicalAuthoring(actual)
				return err
			}
			restore, err := bindCanonicalMutationInvocationArgs(cmd, root.Context(), checked)
			if err != nil {
				t.Fatal(err)
			}
			defer restore()
			root.SetArgs(checked)
			if _, err := root.ExecuteC(); err != nil {
				t.Fatalf("real %s handler read ambient arguments: %v", operation, err)
			}
			var result map[string]json.RawMessage
			if err := json.Unmarshal(out.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if !admitted || operation == "create" && string(result["operation"]) != `"createBead"` || operation == "batch" && string(result["capabilityVersion"]) != `"2"` {
				t.Fatalf("real early admission/emission path was not reached: %s", out.Bytes())
			}
		})
	}
	_, invalidBatch := canonicalMutationLeafFixture(t, []string{"batch"})
	batchCmd = invalidBatch
	if err := invalidBatch.Flags().Set("input-format", "jsonl"); err != nil {
		t.Fatal(err)
	}
	if err := invalidBatch.Flags().Set("file", "must-not-read"); err != nil {
		t.Fatal(err)
	}
	invalidBatch.SetContext(context.WithValue(context.Background(), canonicalMutationInvocationKey{}, "malformed"))
	handled, err := admitCanonicalAuthoring(invalidBatch)
	var refusal *canonicalMutationLeafRefusal
	if !handled || !errors.As(err, &refusal) || refusal.ExitCode != 5 {
		t.Fatalf("early batch admission lost invalid-binding diagnosis or reached input: %v", err)
	}
	if !slices.Equal(os.Args, originalProcess) {
		t.Fatal("captured-argument implementation rewrote process arguments")
	}
}

func TestCanonicalMutationBoundLinkCreateKeepsExactOccurrenceChecks(t *testing.T) {
	stdioMutex.Lock()
	defer stdioMutex.Unlock()
	actualHandler := linkCreateCmd.RunE
	oldReporter := graphFailureReporter
	defer func() { graphFailureReporter = oldReporter }()
	graphFailureReporter = func(code, message string, exit int) error {
		return &canonicalMutationLeafRefusal{Code: code, Message: message, ExitCode: exit}
	}
	for _, tc := range []struct {
		argv      []string
		diagnosis string
	}{
		{[]string{"link", "create", "--type=references", "--from=source"}, "requires SOURCE TARGET"},
		{[]string{"link", "create", "--type=references", "--from=source", "--from=source", "--to=target"}, "--from supplied more than once"},
	} {
		root, cmd := canonicalMutationLeafFixture(t, []string{"link", "create"})
		root.SetContext(context.Background())
		cmd.Args = nil // Reach the existing handler's own exact/pure endpoint checks.
		cmd.RunE = actualHandler
		restore, err := bindCanonicalMutationInvocationArgs(cmd, root.Context(), tc.argv)
		if err != nil {
			t.Fatal(err)
		}
		root.SetArgs(tc.argv)
		_, err = root.ExecuteC()
		restore()
		var refusal *canonicalMutationLeafRefusal
		if !errors.As(err, &refusal) || refusal.ExitCode != 2 || !strings.Contains(refusal.Message, tc.diagnosis) {
			t.Fatalf("existing Link-create checker lost its checked invocation or duplicate rule: %v", err)
		}
	}
}
