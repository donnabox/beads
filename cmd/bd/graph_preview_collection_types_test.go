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
	dependsOn, _ := graphstore.NativeDependencyTypeURL(scope, "depends-on")
	blockedBy, _ := graphstore.NativeDependencyTypeURL(scope, "blocked-by")
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
		{name: "feature-alias", flags: []string{"--type=feat"}, family: scope + "feature"},
		{name: "decision-alias", flags: []string{"--exact-type=adr"}, exact: scope + "decision"},
		{name: "literal-full-alias", flags: []string{"--type=" + scope + "feat"}, family: scope + "feat"},
		{name: "literal-segmented-alias", flags: []string{"--type=types/adr"}, family: scope + "types/adr"},
		{name: "custom-case", flags: []string{"--type=Future-Name"}, family: scope + "Future-Name"},
		{name: "task-full-exact", flags: []string{"--exact-type=" + task}, exact: task},
		{name: "memory-family", flags: []string{"--type=memory"}, family: scope + "memory"},
		{name: "literal-prefixed-memory", flags: []string{"--type=types/memory"}, family: scope + "types/memory"},
		{name: "literal-prefixed-decision", flags: []string{"--type=types/decision"}, family: scope + "types/decision"},
		{name: "native-decision-distinct", flags: []string{"--type=decision"}, family: scope + "decision"},
		{name: "blocks-family", links: true, flags: []string{"--type=blocks"}, family: blocks},
		{name: "literal-depends-on-family", links: true, flags: []string{"--type=depends-on"}, family: dependsOn},
		{name: "literal-blocked-by-exact", links: true, flags: []string{"--exact-type=blocked-by"}, exact: blockedBy},
		{name: "dependency-family", links: true, flags: []string{"--type=dependency"}, family: graphstore.DependencyTypeURL(scope)},
		{name: "any-all", flags: []string{"--type=any"}},
		{name: "compat-family-alias", flags: []string{"--conforms-to=" + task}, family: task},
		{name: "external-exact-uri", flags: []string{"--exact-type=https://external.test/task"}, exact: "https://external.test/task"},
		{name: "issue-family-and-task-exact", flags: []string{"--type=issue", "--exact-type=task"}, exact: task, family: graphstore.IssueTypeURL(scope)},
		{name: "family-alias-and-task-exact", flags: []string{"--conforms-to=" + graphstore.IssueTypeURL(scope), "--exact-type=task"}, exact: task, family: graphstore.IssueTypeURL(scope)},
		{name: "dependency-family-and-blocks-exact", links: true, flags: []string{"--type=dependency", "--exact-type=blocks"}, exact: blocks, family: graphstore.DependencyTypeURL(scope)},
		{name: "duplicate-family-bindings", flags: []string{"--type=issue", "--conforms-to=" + graphstore.IssueTypeURL(scope)}, refuses: true},
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
