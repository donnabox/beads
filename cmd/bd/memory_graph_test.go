package main

import (
	"errors"
	"reflect"
	"testing"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/memoryquery"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func memoryTestCommand(t *testing.T, op memoryReadOperation, flags ...string) *cobra.Command {
	t.Helper()
	c := newMemoryReadCommand(op)
	root := &cobra.Command{Use: "bd"}
	root.PersistentFlags().Bool("json", false, "")
	root.PersistentFlags().String("format", "", "")
	parent := &cobra.Command{Use: "memory", Annotations: memoryGraphCmd.Annotations}
	root.AddCommand(parent)
	parent.AddCommand(c)
	t.Cleanup(func() { delete(canonicalFailureFactories, c) })
	if err := c.ParseFlags(flags); err != nil {
		t.Fatal(err)
	}
	return c
}

func TestMemoryCurrentInput(t *testing.T) {
	t.Setenv(maxRowsEnvVar, "")
	const scope = "https://example.invalid/memory-cli/"
	for _, op := range memoryReadOperations {
		c := memoryTestCommand(t, op)
		got, err := memoryCurrentRequest(c, []string{"a", scope + "beads/b", "a@current", "a"}, scope, op)
		if err != nil {
			t.Fatal(err)
		}
		want := []memoryquery.CurrentRoot{{Path: "a", Depth: op.depth}, {Path: "beads/b", Depth: op.depth}, {Path: "a", Depth: op.depth}, {Path: "a", Depth: op.depth}}
		wantFamily, wantDirection := scope+"types/"+op.family, op.direction
		if op.name == "constraints" || op.name == "subjects" {
			// These fixed discovery operations reuse only the admitted roots
			// and bounds, then dispatch to their dedicated store APIs.
			wantFamily = ""
			wantDirection = (graphstore.CurrentFamilyRequest{}).Direction
			if c.Flags().Lookup("depth") != nil {
				t.Fatalf("%s exposed an ordinary-family depth flag", op.name)
			}
		}
		if !reflect.DeepEqual(got.Roots, want) || got.Family.URI != wantFamily || got.Family.Revision != "" || got.Direction != wantDirection {
			t.Fatalf("%s: %+v", op.name, got)
		}
	}
	op := memoryReadOperations[0]
	// Kind comes from the catalog, never a literal local-ID prefix.
	for _, id := range []string{"beads/a", "links/a", "types/a"} {
		got, err := memoryCurrentRequest(memoryTestCommand(t, op), []string{id}, scope, op)
		if err != nil || len(got.Roots) != 1 || got.Roots[0].Path != id {
			t.Fatalf("literal identity changed: %s %+v %v", id, got, err)
		}
	}
	for _, tc := range []struct {
		name        string
		flags, args []string
		capability  bool
	}{
		{name: "missing"}, {name: "malformed", args: []string{"a//b"}},
		{name: "foreign", args: []string{"https://elsewhere.invalid/beads/a"}},
		{name: "old-pin", args: []string{"a@=old"}, capability: true},
		{name: "legacy-token", args: []string{"a@old"}},
		{name: "negative-depth", flags: []string{"--depth=-1"}, args: []string{"a"}},
		{name: "too-deep", flags: []string{"--depth=1001"}, args: []string{"a"}},
		{name: "empty-node-budget", flags: []string{"--max-nodes=0"}, args: []string{"a"}},
		{name: "large-link-budget", flags: []string{"--max-links=1001"}, args: []string{"a"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := memoryTestCommand(t, op, tc.flags...)
			got, err := memoryCurrentRequest(c, tc.args, scope, op)
			if err == nil || !reflect.DeepEqual(got, graphstore.CurrentFamilyRequest{}) {
				t.Fatalf("admitted invalid input: %+v %v", got, err)
			}
			if tc.capability && !errors.Is(err, graphstore.ErrCapabilityUnavailable) {
				t.Fatalf("historical selector silently reinterpreted: %v", err)
			}
		})
	}
	c := memoryTestCommand(t, op, "--depth=0", "--max-nodes=1", "--max-links=1")
	got, err := memoryCurrentRequest(c, []string{"a"}, scope, op)
	if err != nil || got.Roots[0].Depth != 0 {
		t.Fatalf("zero-hop input: %+v %v", got, err)
	}
	t.Setenv(maxRowsEnvVar, "1")
	if _, err := memoryCurrentRequest(c, []string{"a"}, scope, op); !errors.Is(err, graphstore.ErrCapabilityUnavailable) {
		t.Fatalf("silently ignored row bound: %v", err)
	}
}

func TestMemoryCanonicalBoundary(t *testing.T) {
	for _, tc := range []struct {
		flags  []string
		refuse bool
	}{
		{}, {flags: []string{"--json"}}, {flags: []string{"--output-format=json"}},
		{flags: []string{"--json", "--output-format=json"}},
		{flags: []string{"--output-format=tree"}},
		{flags: []string{"--output-format=default", "--output-format=tree"}},
		{flags: []string{"--json", "--output-format=tree"}, refuse: true},
		{flags: []string{"--json", "--output-format=detail"}, refuse: true},
		{flags: []string{"--output-format=json", "--output-format=detail"}, refuse: true},
		{flags: []string{"--format=json"}, refuse: true},
	} {
		c := memoryTestCommand(t, memoryReadOperations[0], tc.flags...)
		opts, selected, err := canonicalOutputOptions(c)
		if err == nil {
			_, _, err = newMemoryGraphRenderer(opts)
		}
		if !selected || (err != nil) != tc.refuse {
			t.Fatalf("%v: selected=%v err=%v", tc.flags, selected, err)
		}
	}
	if canonicalOutputDefault(rememberCmd) || canonicalOutputDefault(graphCmd) {
		t.Fatal("new namespace changed an existing command contract")
	}
	for _, op := range memoryReadOperations {
		registered, _, err := rootCmd.Find([]string{"memory", op.name})
		if err != nil || registered == nil || registered.Name() != op.name || !canonicalOutputDefault(registered) || registered.InheritedFlags().Lookup("json") == nil {
			t.Fatalf("registered command lost inherited contract: %s %v", op.name, err)
		}
	}
	if canonicalOutputDefault(memoryGraphCmd) || canonicalOutputDefault(memoryRememberAliasCmd) || canonicalOutputDefault(memoryRecallAliasCmd) || canonicalOutputDefault(memoryMemoriesAliasCmd) || canonicalOutputDefault(memoryForgetAliasCmd) {
		t.Fatal("new group changed a convenience alias output contract")
	}
}
