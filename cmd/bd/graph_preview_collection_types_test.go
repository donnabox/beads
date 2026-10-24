package main

import (
	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"testing"
)

func TestNativeConcreteCLICollectionSelection(t *testing.T) {
	const scope = "https://example.test/native/"
	task, _ := graphstore.NativeIssueTypeURL(scope, "task")
	blocks, _ := graphstore.NativeDependencyTypeURL(scope, "blocks")
	for _, tc := range []struct {
		name          string
		links         bool
		flags         []string
		exact, family string
		refuses       bool
	}{
		{name: "default-issue-family", family: graphstore.IssueTypeURL(scope)},
		{name: "default-link-all", links: true},
		{name: "issue-family", flags: []string{"--type=issue"}, family: graphstore.IssueTypeURL(scope)},
		{name: "task-shorthand-family", flags: []string{"--type=task"}, family: task},
		{name: "task-exact", flags: []string{"--exact-type=task"}, exact: task},
		{name: "task-full-exact", flags: []string{"--exact-type=" + task}, exact: task},
		{name: "memory-family", flags: []string{"--type=types/memory"}, family: scope + "types/memory"},
		{name: "memory-decision", flags: []string{"--type=types/decision"}, family: scope + "types/decision"},
		{name: "native-decision-distinct", flags: []string{"--type=decision"}, family: scope + "types/native-issue/6465636973696f6e"},
		{name: "blocks-family", links: true, flags: []string{"--type=blocks"}, family: blocks},
		{name: "dependency-family", links: true, flags: []string{"--type=dependency"}, family: graphstore.DependencyTypeURL(scope)},
		{name: "any-all", flags: []string{"--type=any"}},
		{name: "compat-family-alias", flags: []string{"--conforms-to=" + task}, family: task},
		{name: "external-exact-uri", flags: []string{"--exact-type=https://external.test/task"}, exact: "https://external.test/task"},
		{name: "conflicting-selectors", flags: []string{"--type=task", "--exact-type=task"}, refuses: true},
		{name: "empty", flags: []string{"--type="}, refuses: true},
		{name: "exact-wildcard-refused", flags: []string{"--exact-type=any"}, refuses: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			for _, flag := range []string{"type", "exact-type", "conforms-to"} {
				cmd.Flags().String(flag, "", "")
			}
			if err := cmd.ParseFlags(tc.flags); err != nil {
				t.Fatal(err)
			}
			exact, family, err := graphPreviewCollectionTypes(cmd, scope, tc.links)
			if (err != nil) != tc.refuses || exact != tc.exact || family != tc.family {
				t.Fatalf("exact=%q family=%q err=%v; want %q %q refusal=%t", exact, family, err, tc.exact, tc.family, tc.refuses)
			}
		})
	}
}
