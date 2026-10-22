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
		want := []memoryquery.CurrentRoot{{Path: "beads/a", Depth: op.depth}, {Path: "beads/b", Depth: op.depth}, {Path: "beads/a", Depth: op.depth}, {Path: "beads/a", Depth: op.depth}}
		if !reflect.DeepEqual(got.Roots, want) || got.Family.URI != scope+"types/"+op.family || got.Direction != op.direction {
			t.Fatalf("%s: %+v", op.name, got)
		}
	}
	op := memoryReadOperations[0]
	for _, tc := range []struct {
		name        string
		flags, args []string
		capability  bool
	}{
		{name: "missing"}, {name: "link", args: []string{"links/a"}},
		{name: "foreign", args: []string{"https://elsewhere.invalid/beads/a"}},
		{name: "old-pin", args: []string{"a@old"}, capability: true},
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
	if !canonicalOutputDefault(memoryGraphCmd) {
		t.Fatal("new group is not canonical by default")
	}
}
