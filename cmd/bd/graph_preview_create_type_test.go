package main

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
)

func graphCreateTypeCommand(t *testing.T, flags []string) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{}
	registerCommonIssueFlags(cmd)
	for _, name := range []string{"id", "title", "properties", "type", "priority", "labels", "label", "spec-id", "due", "file", "graph", "storage-class"} {
		cmd.Flags().String(name, "", "")
	}
	cmd.Flags().Int("estimate", 0, "")
	if err := cmd.ParseFlags(flags); err != nil {
		t.Fatal(err)
	}
	return cmd
}

func TestGraphPreviewIssueCreateProperties(t *testing.T) {
	issue := &types.Issue{}
	cmd := graphCreateTypeCommand(t, nil)
	err := graphPreviewApplyIssueCreateProperties(cmd, issue, map[string]any{
		"design": "Design notes", "labels": []any{"release", "preview"},
		"estimated_minutes": float64(45), "external_ref": "gh-7450",
	})
	if err != nil || issue.Design != "Design notes" || len(issue.Labels) != 2 || issue.EstimatedMinutes == nil || *issue.EstimatedMinutes != 45 || issue.ExternalRef == nil || *issue.ExternalRef != "gh-7450" {
		t.Fatalf("typed Issue property initialization failed: %+v, %v", issue, err)
	}
	for _, tc := range []struct {
		name  string
		flags []string
		props map[string]any
	}{
		{"duplicate-estimate", []string{"--estimate=45"}, map[string]any{"estimated_minutes": float64(45)}},
		{"fractional-estimate", nil, map[string]any{"estimated_minutes": 1.5}},
		{"immutable-status", nil, map[string]any{"status": "closed"}},
		{"wrong-label-type", nil, map[string]any{"labels": []any{"valid", float64(3)}}},
		{"invalid-due", nil, map[string]any{"due_at": "tomorrow"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := graphCreateTypeCommand(t, tc.flags)
			var failure *exitError
			if err := graphPreviewApplyIssueCreateProperties(cmd, &types.Issue{}, tc.props); !errors.As(err, &failure) || failure.Code != 2 {
				t.Fatalf("expected typed property refusal, got %v", err)
			}
		})
	}
}

func TestGraphPreviewCreateBeadTypeSelection(t *testing.T) {
	const scope = "https://example.invalid/project/"
	previous := graphPreviewTypeGeneration
	graphPreviewTypeGeneration = 0
	t.Cleanup(func() { graphPreviewTypeGeneration = previous })
	for _, tc := range []struct {
		name, want string
		flags      []string
		code       int
	}{
		{name: "default-issue", want: scope + "types/native-issue/7461736b"},
		{name: "classification-is-concrete-type", flags: []string{"--type=bug"}, want: scope + "types/native-issue/627567"},
		{name: "prospective-native-classification", flags: []string{"--type=preview-task"}, want: scope + "types/native-issue/707265766965772d7461736b"},
		{name: "prospective-example-classification", flags: []string{"--type=example-task"}, want: scope + "types/native-issue/6578616d706c652d7461736b"},
		{name: "prospective-custom-issue", flags: []string{"--type=preview-issue"}, want: scope + "types/native-issue/707265766965772d6973737565"},
		{name: "relative-issue", flags: []string{"--type=types/issue"}, code: 2},
		{name: "canonical-issue", flags: []string{"--type=" + graphstore.IssueTypeURL(scope)}, code: 2},
		{name: "prospective-custom-memory-label", flags: []string{"--type=preview-memory"}, want: scope + "types/native-issue/707265766965772d6d656d6f7279"},
		{name: "prospective-custom-label", flags: []string{"--type=memory-preview"}, want: scope + "types/native-issue/6d656d6f72792d70726576696577"},
		{name: "relative-memory", flags: []string{"--type=types/memory"}, want: graphstore.MemoryTypeURL(scope)},
		{name: "canonical-memory", flags: []string{"--type=" + graphstore.MemoryTypeURL(scope)}, want: graphstore.MemoryTypeURL(scope)},
		{name: "empty", flags: []string{"--type="}, code: 2},
		{name: "foreign", flags: []string{"--type=https://foreign.invalid/types/preview-memory"}, code: 2},
		{name: "uninstalled", flags: []string{"--type=types/custom"}, code: 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := graphPreviewCreateBeadType(graphCreateTypeCommand(t, tc.flags), scope)
			if tc.code == 0 {
				if err != nil || got != tc.want {
					t.Fatalf("type=%q want=%q error=%v", got, tc.want, err)
				}
			} else {
				var failure *exitError
				if got != "" || !errors.As(err, &failure) || failure.Code != tc.code {
					t.Fatalf("refusal type=%q error=%v want exit %d", got, err, tc.code)
				}
			}
		})
	}
}

func TestGraphPreviewCreateBeadTypeUsesConcreteNativeTypes(t *testing.T) {
	const scope = "https://example.invalid/project/"
	previous := graphPreviewTypeGeneration
	graphPreviewTypeGeneration = graphstore.TypeGeneration
	t.Cleanup(func() { graphPreviewTypeGeneration = previous })
	task, _ := graphstore.NativeIssueTypeURL(scope, "task")
	for _, flags := range [][]string{nil, {"--type=" + task}} {
		if got, err := graphPreviewCreateBeadType(graphCreateTypeCommand(t, flags), scope); err != nil || got != task {
			t.Fatalf("flags %v: Issue Type %q, err %v", flags, got, err)
		}
	}
	for _, old := range []string{graphstore.IssueTypeURL(scope), "types/preview-issue-v3"} {
		if got, err := graphPreviewCreateBeadType(graphCreateTypeCommand(t, []string{"--type=" + old}), scope); got != "" || err == nil {
			t.Fatalf("admitted abstract or obsolete Type: %q, %v", got, err)
		}
	}
}

func TestGraphPreviewNativeDependencySelectionUsesConcreteTypes(t *testing.T) {
	const scope = "https://example.invalid/project/"
	for _, native := range []types.DependencyType{types.DepBlocks, types.DepParentChild, types.DepDuplicates, types.DepSupersedes, "my-own-kind"} {
		url, admitted := graphPreviewDependencyTypeURL(scope, native)
		want, _ := graphstore.NativeDependencyTypeURL(scope, string(native))
		decoded, ok := graphPreviewNativeDependencyType(scope, url)
		if !admitted || url != want || !ok || decoded != native {
			t.Fatalf("native %q maps to %q decoded %q; want %q", native, url, decoded, want)
		}
	}
	if _, ok := graphPreviewDependencyTypeURL(scope, "123456789012345678901234567890123"); ok {
		t.Fatal("overlong native kind admitted")
	}
	if _, ok := graphPreviewNativeDependencyType(scope, graphstore.DependencyTypeURL(scope)); ok {
		t.Fatal("abstract Dependency admitted as concrete")
	}
}

func TestGraphPreviewCreateMemoryInput(t *testing.T) {
	for _, tc := range []struct {
		name        string
		flags, args []string
		title, body string
		code        int
	}{
		{name: "positional-title", args: []string{"Code flow policy"}, title: "Code flow policy"},
		{name: "explicit-title", flags: []string{"--title=Explicit title", "--description=Body text"}, title: "Explicit title", body: "Body text"},
		{name: "matching-titles", args: []string{"Same title"}, flags: []string{"--title=Same title"}, title: "Same title"},
		{name: "body-summary", flags: []string{"--body=  Code  flow\tpolicy\nKeep all body text."}, title: "Code flow policy", body: "  Code  flow\tpolicy\nKeep all body text."},
		{name: "message-summary", flags: []string{"--message=Memory — 雪"}, title: "Memory — 雪", body: "Memory — 雪"},
		{name: "matching-body-aliases", flags: []string{"--description=Same body", "--body=Same body", "--message=Same body"}, title: "Same body", body: "Same body"},
		{name: "exact-properties", flags: []string{`--properties={"title":"Property title","body":"Property body"}`}, title: "Property title", body: "Property body"},
		{name: "properties-body-summary", flags: []string{`--properties={"body":"Property body"}`}, title: "Property body", body: "Property body"},
		{name: "properties-disjoint-title", args: []string{"Positional title"}, flags: []string{`--properties={"body":"Property body"}`}, title: "Positional title", body: "Property body"},
		{name: "explicit-empty-body", flags: []string{"--body="}, title: "Untitled memory"},
		{name: "no-content", code: 2},
		{name: "conflicting-titles", args: []string{"Other"}, flags: []string{"--title=Title"}, code: 2},
		{name: "empty-title", flags: []string{"--title=", "--body=Body"}, code: 2},
		{name: "blank-title", flags: []string{"--title= \t", "--body=Body"}, code: 2},
		{name: "blank-positional-title", args: []string{" "}, flags: []string{"--body=Body"}, code: 2},
		{name: "multiple-titles", args: []string{"One", "Two"}, code: 2},
		{name: "conflicting-body-aliases", flags: []string{"--description=One", "--body=Two"}, code: 2},
		{name: "empty-description-conflict", flags: []string{"--description=", "--message=Body"}, code: 2},
		{name: "empty-message-conflict", flags: []string{"--description=Body", "--message="}, code: 2},
		{name: "properties-title-duplicate", args: []string{"Positional title"}, flags: []string{`--properties={"title":"Property title"}`}, code: 2},
		{name: "properties-body-duplicate", flags: []string{"--body=Body", `--properties={"body":"Body"}`}, code: 2},
		{name: "properties-unknown", flags: []string{`--properties={"content":"Body"}`}, code: 2},
		{name: "properties-wrong-type", flags: []string{`--properties={"body":3}`}, code: 2},
		{name: "invalid-title", flags: []string{"--title=\xff"}, code: 2},
		{name: "invalid-body", flags: []string{"--body=\xff"}, code: 2},
		{name: "description-stdin", flags: []string{"--description=-"}, code: 5},
		{name: "body-stdin", flags: []string{"--body=-"}, code: 5},
		{name: "message-stdin", flags: []string{"--message=-"}, code: 5},
		{name: "body-file", flags: []string{"--body-file=/does/not/exist"}, code: 5},
		{name: "stdin", flags: []string{"--stdin"}, code: 5},
		{name: "issue-priority", flags: []string{"--priority=2"}, code: 5},
		{name: "issue-labels", flags: []string{"--labels=memory"}, code: 5},
		{name: "issue-assignee", flags: []string{"--assignee=actor"}, code: 5},
		{name: "issue-design", flags: []string{"--design=Design"}, code: 5},
		{name: "issue-notes", flags: []string{"--notes=Notes"}, code: 5},
		{name: "issue-due", flags: []string{"--due=tomorrow"}, code: 5},
		{name: "issue-estimate", flags: []string{"--estimate=0"}, code: 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := graphCreateTypeCommand(t, append([]string{"--type=types/memory"}, tc.flags...))
			probe := &selectedRememberInputProbe{}
			cmd.SetIn(probe)
			title, body, err := graphPreviewCreateMemoryInput(cmd, tc.args)
			if tc.code == 0 {
				if err != nil || title != tc.title || body != tc.body {
					t.Fatalf("title=%q body=%q error=%v; want %q %q", title, body, err, tc.title, tc.body)
				}
			} else {
				var failure *exitError
				if !errors.As(err, &failure) || failure.Code != tc.code || title != "" || body != "" {
					t.Fatalf("refusal title=%q body=%q error=%v want exit %d", title, body, err, tc.code)
				}
			}
			if probe.reads != 0 {
				t.Fatal("inline Memory create read stdin")
			}
		})
	}
}
