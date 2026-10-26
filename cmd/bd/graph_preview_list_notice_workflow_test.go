//go:build cgo

package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// Typed Issue filters and configured labels select the right population without
// preview bookkeeping in human output. A configured Issue label never refuses
// a Memory Type. The label comes from a workspace config.yaml.
func TestGraphPreviewListNoticeWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			const scope = "https://example.invalid/notice/"
			work, home := graphListWorkspace(t, bd, engine, scope)
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			human := func(extra ...string) []string {
				t.Helper()
				out := graphPolicyCLI(t, bd, work, home, nil, "", append([]string{"list"}, extra...)...)
				return strings.Split(strings.TrimSuffix(out, "\n"), "\n")
			}
			structured := func(extra ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append([]string{"list", "--format", "records-json"}, extra...)...)
			}
			expect := func(what string, got []string, want ...string) {
				t.Helper()
				if strings.Join(got, "\n") != strings.Join(want, "\n") {
					t.Fatalf("%s:\n got=%q\nwant=%q", what, got, want)
				}
			}
			call("remember", "plan body", "--id", "beads/plan", "--title", "Plan")
			workIssue := graphMixedResult[graphstore.IssueRecord](t, call("create", "Frontend work", "--graph-path", "beads/work", "--priority", "1", "--labels", "frontend"))
			otherIssue := graphMixedResult[graphstore.IssueRecord](t, call("create", "Other work", "--graph-path", "beads/other", "--priority", "2"))
			workRow := workIssue.Properties.ID + ` [task] open P1 "Frontend work"`
			otherRow := otherIssue.Properties.ID + ` [task] open P2 "Other work"`

			// A typed Issue option selects the Issue query without a preview-only
			// bookkeeping notice in the default human output.
			lines := human("--status", "open")
			if len(lines) != 3 || lines[0] != "Issues (2; more: false; graph preview)" || lines[1] != workRow || lines[2] != otherRow {
				t.Fatalf("--status open: %q", lines)
			}
			for _, args := range [][]string{{"--status", "open", "--sort", "priority"}, {"--sort", "priority", "--status", "open"}} {
				if got := human(args...); len(got) != 3 || got[0] != lines[0] || got[1] != workRow || got[2] != otherRow {
					t.Fatalf("%v changed selected Issues: %q", args, got)
				}
			}
			// Structured output carries no notice: it is byte-identical to the
			// Issue Type route, which never prints one, and --quiet prints nothing.
			typed := structured("--status", "open")
			if narrowed := structured("--type", "issue", "--status", "open"); typed != narrowed || strings.Contains(typed, "not listed") {
				t.Fatalf("structured output changed:\n got=%s\nwant=%s", typed, narrowed)
			}
			if quiet := graphPolicyCLI(t, bd, work, home, nil, "", "list", "--status", "open", "--quiet"); quiet != "" {
				t.Fatalf("--quiet printed %q", quiet)
			}
			// Narrowing to the Issue Type already told the caller it is Issues only.
			expect("Issue Type with a typed option", human("--type", "issue", "--status", "open"),
				"Issues (2; more: false; graph preview)", workRow, otherRow)

			// A configured directory label for this directory selects the Issue
			// query too. config.yaml is read from the workspace, and the label
			// pattern is matched against the working directory.
			graphListConfig(t, work, "directory:\n  labels:\n    \""+filepath.Base(work)+"\": frontend\n")
			expect("configured label alone", human(), "Issues (1; more: false; graph preview)", workRow)
			expect("configured label with the Issue Type", human("--type", "issue"), "Issues (1; more: false; graph preview)", workRow)
			// A Memory has no labels, so the label is not applied to one and nothing is refused.
			memoryRows := human("--type", "memory")
			if len(memoryRows) != 3 || memoryRows[0] != "Beads (1; more: false; graph preview)" ||
				memoryRows[1] != `directory.labels "frontend" is not applied to this Bead Type.` ||
				memoryRows[2] != `beads/plan [types/memory] "Plan"` {
				t.Fatalf("configured Issue label changed Memory selection: %q", memoryRows)
			}
			// A typed label replaces the configured one.
			expect("typed label", human("--label", "frontend"), "Issues (1; more: false; graph preview)", workRow)
			expect("typed option and configured label", human("--status", "open"), "Issues (1; more: false; graph preview)", workRow)
			if out := structured(); strings.Contains(out, "not listed") || strings.Contains(out, "not applied") || !strings.Contains(out, workIssue.ID) || strings.Contains(out, scope+"beads/plan") {
				t.Fatalf("structured list with the configured label: %s", out)
			}
			// A typed Issue option with a Memory Type is still refused.
			graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "list", "--type", "memory", "--status", "open", "--format", "records-json")
		})
	}
}

// A typed Issue option selects the Issue query without a preview-only notice,
// and with a Memory Type it is refused. Empty --type is an invalid Type family.
func TestGraphPreviewListTypedOptionsSelectIssueQuery(t *testing.T) {
	bd := buildBDUnderTest(t)
	work, home := graphListWorkspace(t, bd, "embedded", "https://example.invalid/typed/")
	for _, args := range [][]string{{"remember", "plan body", "--id", "beads/plan", "--title", "Plan"}, {"create", "Frontend work", "--graph-path", "beads/work"}} {
		graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
	}
	for _, tc := range []struct{ option, value string }{
		{"status", ""}, {"state", "open"}, {"state", ""}, {"type", "issue"}, {"assignee", ""},
	} {
		arg := "--" + tc.option + "=" + tc.value
		out := graphPolicyCLI(t, bd, work, home, nil, "", "list", arg)
		lines := strings.Split(strings.TrimSuffix(out, "\n"), "\n")
		if len(lines) != 2 || !strings.HasPrefix(lines[0], "Issues (") || !strings.Contains(lines[1], "Frontend work") || strings.Contains(out, "Memories are not listed") {
			t.Errorf("%s did not select the Issue query: %q", arg, lines)
		}
	}
	graphPolicyCLI(t, bd, work, home, nil, "invalid_selector", "list", "--type=", "--format", "records-json")
	for _, option := range []string{"status", "state", "assignee"} {
		graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "list", "--type", "memory", "--"+option+"=", "--format", "records-json")
	}
}

func TestGraphPreviewCustomIssueTypePrefixWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := graphListWorkspace(t, bd, engine, "https://example.invalid/custom-prefix/")
			graphListConfig(t, work, "types.custom: \"preview-task,example-task\"\n")
			listedTypes := graphPolicyCLI(t, bd, work, home, nil, "", "types")
			if !strings.Contains(listedTypes, "preview-task") || !strings.Contains(listedTypes, "example-task") ||
				!strings.Contains(listedTypes, "types/native-issue/707265766965772d7461736b") ||
				!strings.Contains(listedTypes, "types/native-issue/6578616d706c652d7461736b") {
				t.Fatalf("registered Issue types missing from human inventory: %s", listedTypes)
			}
			selected := make(map[string]graphstore.IssueRecord)
			for _, tc := range []struct{ kind, id string }{
				{"preview-task", "beads/preview"},
				{"example-task", "beads/example"},
			} {
				selected[tc.kind] = graphMixedResult[graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "create", tc.kind, "--graph-path", tc.id, "--type", tc.kind, "--json"))
			}
			for _, tc := range []struct{ kind, otherKind string }{
				{"preview-task", "example-task"},
				{"example-task", "preview-task"},
			} {
				exactType, err := graphstore.NativeIssueTypeURL("https://example.invalid/custom-prefix/", tc.kind)
				if err != nil {
					t.Fatal(err)
				}
				listed := graphPolicyCLI(t, bd, work, home, nil, "", "list", "--type", "issue", "--exact-type", exactType, "--format", "records-json")
				if !strings.Contains(listed, selected[tc.kind].ID) || strings.Contains(listed, selected[tc.otherKind].ID) {
					t.Fatalf("registered Issue type %q did not select only its allocated ID: %s", tc.kind, listed)
				}
				shorthand := graphPolicyCLI(t, bd, work, home, nil, "", "list", "--type", tc.kind, "--format", "records-json")
				if !strings.Contains(shorthand, selected[tc.kind].ID) || strings.Contains(shorthand, selected[tc.otherKind].ID) {
					t.Fatalf("registered Issue shorthand %q did not select only its allocated ID: %s", tc.kind, shorthand)
				}
			}
		})
	}
}
