package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func canonicalMutationLeafFixture(t *testing.T, path []string) (*cobra.Command, *cobra.Command) {
	t.Helper()
	d, ok := lookupCanonicalMutationLeaf(path)
	if !ok { t.Fatal("fixture requires a covered operation") }
	root := &cobra.Command{Use: "bd", SilenceErrors: true, SilenceUsage: true}
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	parent := root
	for _, part := range path[:len(path)-1] {
		group := &cobra.Command{Use: part}
		parent.AddCommand(group)
		parent = group
	}
	cmd := &cobra.Command{Use: path[len(path)-1], RunE: func(*cobra.Command, []string) error { return nil }}
	parent.AddCommand(cmd)
	for _, name := range d.Flags {
		switch name {
		case "force", "json", "readonly", "quiet", "no-color": cmd.Flags().Bool(name, false, "")
		case "set-property", "unset-property", "set-metadata", "unset-metadata", "batch-arg", "output-format", "include", "omit": cmd.Flags().StringArray(name, nil, "")
		default: cmd.Flags().String(name, "", "")
		}
	}
	cmd.Args = func(cmd *cobra.Command, args []string) error { return admitCanonicalMutationLeaf(cmd, path, args) }
	return root, cmd
}

func TestCanonicalMutationLeafAdmissionBeforeHooks(t *testing.T) {
	for _, tc := range []struct { name string; path, args []string }{
		{"missing create Type", []string{"create"}, nil},
		{"empty create Type", []string{"create"}, []string{"--type="}},
		{"compat positional title", []string{"create"}, []string{"title", "--type=memory"}},
		{"implicit last touched", []string{"update"}, []string{"--properties={}"}},
		{"compat multi update", []string{"update"}, []string{"first", "second", "--properties={}"}},
		{"pinned write target", []string{"update"}, []string{"first@=r7", "--properties={}"}},
		{"current write suffix", []string{"delete"}, []string{"first@current", "--force"}},
		{"multiple deletes", []string{"delete"}, []string{"first", "second", "--force"}},
		{"batch binding outside emission", []string{"create"}, []string{"--type=memory", "--batch-bind=x"}},
		{"singleton batch Reference", []string{"update"}, []string{"batch-ref::x", "--properties={}"}},
		{"both allocation modes", []string{"create"}, []string{"--type=memory", "--id=x", "--id-root=prefix"}},
		{"duplicated endpoint", []string{"link", "create"}, []string{"a", "b", "--from=a", "--type=references"}},
		{"missing target endpoint", []string{"link", "create"}, []string{"a", "--type=references"}},
		{"batch endpoint outside emission", []string{"link", "create"}, []string{"batch-ref::a", "b", "--type=references"}},
		{"implicit current endpoint pin", []string{"link", "create"}, []string{"a@current", "b", "--type=references"}},
		{"conflicting edit modes before input", []string{"update"}, []string{"a", "--properties=@must-not-read", "--patch-properties=@-"}},
		{"invalid preview mode", []string{"create"}, []string{"--type=memory", "--dry-run-mode="}},
		{"missing batch file", []string{"batch"}, nil},
		{"invalid batch format before input", []string{"batch"}, []string{"--file=must-not-read", "--input-format=command-text"}},
		{"empty retry filename", []string{"batch"}, []string{"--retry-file="}},
		{"retry semantic override", []string{"batch"}, []string{"--retry-file=must-not-read", "--file=also-must-not-read"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, cmd := canonicalMutationLeafFixture(t, tc.path)
			acquired, ran := false, false
			root.PersistentPreRunE = func(*cobra.Command, []string) error { acquired = true; return nil }
			cmd.RunE = func(*cobra.Command, []string) error { ran = true; return nil }
			root.SetArgs(append(append([]string(nil), tc.path...), tc.args...))
			_, err := root.ExecuteC()
			var refusal *canonicalMutationLeafRefusal
			if !errors.As(err, &refusal) || refusal.Code != "invalid_selector" || refusal.ExitCode != 2 {
				t.Fatalf("expected structured invalid input refusal, got %v", err)
			}
			if acquired || ran { t.Fatal("invalid canonical invocation reached an acquisition hook or executor") }
		})
	}
}

func TestCanonicalMutationLeafPreservesReferencesAndInputs(t *testing.T) {
	root, cmd := canonicalMutationLeafFixture(t, []string{"link", "create"})
	from, to := "first@=r3", "[https://example.invalid/s/second]@=r8"
	var received []string
	cmd.RunE = func(_ *cobra.Command, args []string) error { received = append([]string(nil), args...); return nil }
	root.SetArgs([]string{"link", "create", from, "--to", to, "--type=references@=type-r4", "--properties=@must-not-read"})
	if _, err := root.ExecuteC(); err != nil { t.Fatal(err) }
	if len(received) != 1 || received[0] != from { t.Fatalf("source pin changed: %v", received) }
	if got, _ := cmd.Flags().GetString("to"); got != to { t.Fatalf("target pin changed: %q", got) }
	if got, _ := cmd.Flags().GetString("type"); got != "references@=type-r4" { t.Fatalf("Type pin changed: %q", got) }
	if got, _ := cmd.Flags().GetString("properties"); got != "@must-not-read" { t.Fatalf("input read or rewritten: %q", got) }
}

func TestCanonicalMutationLeafRejectsExplicitCompatibilityFlags(t *testing.T) {
	root, cmd := canonicalMutationLeafFixture(t, []string{"delete"})
	cmd.Flags().Bool("cascade", false, "")
	root.SetArgs([]string{"delete", "one", "--cascade=false"})
	_, err := root.ExecuteC()
	var refusal *canonicalMutationLeafRefusal
	if !errors.As(err, &refusal) || !strings.Contains(refusal.Message, "--cascade") { t.Fatalf("compatibility flag admitted: %v", err) }
}

func TestCanonicalMutationLeafRejectsCompatibilityAliases(t *testing.T) {
	root, cmd := canonicalMutationLeafFixture(t, []string{"create"})
	cmd.Aliases = []string{"new"}
	root.SetArgs([]string{"new", "--type=memory"})
	_, err := root.ExecuteC()
	var refusal *canonicalMutationLeafRefusal
	if !errors.As(err, &refusal) || !strings.Contains(refusal.Message, "alias") { t.Fatalf("compatibility alias admitted: %v", err) }
}

func TestCanonicalMutationLeafGraphRouteRefusesFallback(t *testing.T) {
	_, cmd := canonicalMutationLeafFixture(t, []string{"batch"})
	if err := requireCanonicalMutationGraphRoute(cmd, []string{"batch"}, false); err == nil { t.Fatal("legacy batch fallback admitted") }
	if err := requireCanonicalMutationGraphRoute(cmd, []string{"batch"}, true); err != nil { t.Fatal(err) }
	if err := cmd.Flags().Set("dry-run-mode", "batch"); err != nil { t.Fatal(err) }
	if err := requireCanonicalMutationGraphRoute(cmd, []string{"batch"}, false); err != nil { t.Fatalf("offline authoring requires no provider: %v", err) }
}

func TestCanonicalMutationLeafExecutorCapturesHandler(t *testing.T) {
	stdioMutex.Lock()
	defer stdioMutex.Unlock()
	oldCreate, oldActive := createCmd, graphPreviewActive
	_, createCmdFixture := canonicalMutationLeafFixture(t, []string{"create"})
	createCmd = createCmdFixture
	defer func() { createCmd, graphPreviewActive = oldCreate, oldActive }()
	if err := createCmd.Flags().Set("type", "memory"); err != nil { t.Fatal(err) }
	called := 0
	createCmd.RunE = func(*cobra.Command, []string) error { called++; return nil }
	run, err := canonicalMutationLeafExecutor(createCmd, []string{"create"})
	if err != nil { t.Fatal(err) }
	createCmd.RunE = run
	graphPreviewActive = true
	if err := run(createCmd, nil); err != nil { t.Fatal(err) }
	if called != 1 { t.Fatalf("existing handler not called exactly once: %d", called) }
	graphPreviewActive = false
	if err := run(createCmd, nil); err == nil { t.Fatal("legacy mutation fallback admitted") }
	if called != 1 { t.Fatal("legacy handler executed after missing graph admission") }
}

func TestCanonicalMutationLeafExecutorRejectsWrongCommand(t *testing.T) {
	if _, err := canonicalMutationLeafExecutor(createCmd, []string{"link", "create"}); err == nil { t.Fatal("Bead command admitted as Link creation") }
}

func TestCanonicalMutationLeafUnsupportedEncoderBeforeHooks(t *testing.T) {
	root, _ := canonicalMutationLeafFixture(t, []string{"batch"})
	acquired := false
	root.PersistentPreRunE = func(*cobra.Command, []string) error { acquired = true; return nil }
	root.SetArgs([]string{"batch", "--file=must-not-read", "--batch-version=999"})
	_, err := root.ExecuteC()
	var refusal *canonicalMutationLeafRefusal
	if !errors.As(err, &refusal) || refusal.Code != "capability_unavailable" || refusal.ExitCode != 5 || acquired { t.Fatalf("unsupported encoder acquired a provider or changed diagnosis: %v", err) }
}

func TestCanonicalMutationLeafOccurrencesBeforeHooks(t *testing.T) {
	for _, tc := range []struct { name string; path, argv []string }{
		{"changed Type", []string{"create"}, []string{"create", "--type=memory", "--type=constraint"}},
		{"changed guard", []string{"update"}, []string{"update", "one", "--properties={}", "--if-revision==r1", "--if-revision==r2"}},
		{"discarded JSON input", []string{"update"}, []string{"update", "one", "--properties=@must-not-read", "--properties={}"}},
		{"changed delete apply intent", []string{"delete"}, []string{"delete", "one", "--force=false", "--force=true"}},
		{"equal duplicated endpoint", []string{"link", "create"}, []string{"link", "create", "--from=a", "--from=a", "--to=b", "--type=references"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, cmd := canonicalMutationLeafFixture(t, tc.path)
			acquired := false
			root.PersistentPreRunE = func(*cobra.Command, []string) error { acquired = true; return nil }
			cmd.Args = canonicalMutationLeafArgs(cmd, tc.path, tc.argv)
			root.SetArgs(tc.argv)
			_, err := root.ExecuteC()
			var refusal *canonicalMutationLeafRefusal
			if !errors.As(err, &refusal) || refusal.ExitCode != 2 || acquired { t.Fatalf("conflicting occurrence reached a provider or changed diagnosis: %v", err) }
		})
	}
}

func TestCanonicalMutationLeafArgsOwnsCheckedArgv(t *testing.T) {
	root, cmd := canonicalMutationLeafFixture(t, []string{"update"})
	path := []string{"update"}
	argv := []string{"update", "one", "--set-property=a=1", "--set-property=b=2"}
	cmd.Args = canonicalMutationLeafArgs(cmd, path, argv)
	actualArgv := append([]string(nil), argv...)
	// Without detached copies, these changes would select create's Args or
	// report conflicting properties instead of accepting the original edits.
	path[0] = "create"
	argv[2], argv[3] = "--properties=@must-not-read", "--properties={}"
	root.SetArgs(actualArgv)
	if _, err := root.ExecuteC(); err != nil { t.Fatalf("repeatable edits or captured argv changed: %v", err) }
}
