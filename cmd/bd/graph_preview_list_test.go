package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/config"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
)

func graphListTestCommand(t *testing.T, args []string) *cobra.Command {
	t.Helper()
	cmd := newListLimitCommand(t)
	for _, name := range []string{"format", "state", "id", "title", "title-contains", "priority", "priority-min", "priority-max", "sort", "due-before", "due-after"} {
		cmd.Flags().String(name, "", "")
	}
	cmd.Flags().StringP("status", "s", "", "")
	cmd.Flags().StringP("type", "t", "", "")
	cmd.Flags().StringP("assignee", "a", "", "")
	cmd.Flags().StringSlice("exclude-type", nil, "")
	cmd.Flags().String("parent", "", "")
	cmd.Flags().String("filter-parent", "", "")
	for _, name := range []string{"flat", "json", "quiet", "pinned", "no-pinned", "no-parent", "reverse", "ready", "tree", "no-assignee", "overdue", "include-comments"} {
		cmd.Flags().Bool(name, false, "")
	}
	for _, name := range []string{"label", "label-any", "exclude-label"} {
		cmd.Flags().StringSlice(name, nil, "")
	}
	if err := cmd.ParseFlags(args); err != nil {
		t.Fatal(err)
	}
	return cmd
}

func TestGraphPreviewIssueListInput(t *testing.T) {
	t.Chdir(t.TempDir())
	config.ResetForTesting()
	t.Cleanup(config.ResetForTesting)
	if err := config.Initialize(); err != nil {
		t.Fatal(err)
	}
	prior := jsonOutput
	t.Cleanup(func() { jsonOutput = prior })
	for _, tc := range []struct {
		name       string
		args       []string
		ambient    bool
		exit       int
		structured bool
		limit      int
	}{
		{"flat", []string{"--flat", "--limit=2"}, false, 0, false, 2},
		{"tree-false", []string{"--tree=false", "--limit=2"}, false, 0, false, 2},
		{"records", []string{"--format=records-json", "--all"}, false, 0, true, 0},
		{"records-with-flat", []string{"--format=records-json", "--flat", "--limit=2"}, false, 0, true, 2},
		{"records-ambient", []string{"--format=records-json", "--limit=3"}, true, 0, true, 3},
		{"explicit-limit-wins-all", []string{"--flat", "--all", "--limit=1"}, false, 0, false, 1},
		{"bare", nil, false, 0, false, 0},
		{"legacy-json", []string{"--json"}, false, 0, false, 0},
		{"legacy-json-comments", []string{"--json", "--include-comments"}, false, 0, false, 0},
		{"comments-need-json", []string{"--include-comments"}, false, 5, false, 0},
		{"false-json", []string{"--flat", "--json=false"}, false, 5, false, 0},
		{"format-json", []string{"--format=JSON"}, false, 0, false, 0},
		{"format-json-comments", []string{"--format=json", "--include-comments"}, false, 0, false, 0},
		{"ambient-flat", []string{"--flat"}, true, 0, false, 0},
		{"false-flat", []string{"--flat=false"}, false, 5, false, 0},
		{"false-ready", []string{"--flat", "--ready=false"}, false, 0, false, 0},
		{"flat-and-tree-false", []string{"--flat", "--tree=false"}, false, 0, false, 0},
		{"tree-true", []string{"--tree=true"}, false, 5, false, 0},
		{"negative-limit", []string{"--flat", "--limit=-1"}, false, 2, false, 0},
		{"bad-priority", []string{"--flat", "--priority=bad"}, false, 2, false, 0},
		{"sort-id", []string{"--flat", "--sort=id"}, false, 0, false, 0},
		{"id-set", []string{"--id=one", "--id=two,three", "--sort=id"}, false, 0, false, 0},
		{"sort-assignee", []string{"--flat", "--sort=assignee"}, false, 0, false, 0},
		{"sort-closed", []string{"--flat", "--sort=closed", "--all"}, false, 0, false, 0},
		{"both-pin", []string{"--flat", "--pinned", "--no-pinned"}, false, 2, false, 0},
		{"no-parent", []string{"--flat", "--no-parent"}, false, 0, false, 0},
		{"parent-flat", []string{"--flat", "--parent=beads/work"}, false, 0, false, 0},
		{"parent-tree-false", []string{"--tree=false", "--parent=beads/work"}, false, 0, false, 0},
		{"parent-json", []string{"--json", "--parent=beads/work"}, false, 0, false, 0},
		{"parent-records", []string{"--format=records-json", "--parent=beads/work"}, false, 0, true, 0},
		{"parent-default-tree", []string{"--parent=beads/work"}, false, 0, false, 0},
		{"parent-empty", []string{"--flat", "--parent="}, false, 2, false, 0},
		{"both-status-aliases", []string{"--flat", "--status=open", "--state=closed"}, false, 2, false, 0},
		{"repeat-status", []string{"--flat", "--status=open", "-sclosed"}, false, 5, false, 0},
		{"repeat-type", []string{"--flat", "--type=task", "-tbug"}, false, 5, false, 0},
		{"oversize-label", []string{"--flat", "--label=" + strings.Repeat("x", 4097)}, false, 2, false, 0},
		{"invalid-label-utf8", []string{"--flat", "--label=" + string([]byte{255})}, false, 2, false, 0},
		{"explicit-empty-exclude-type", []string{"--flat", "--exclude-type="}, false, 5, false, 0},
		{"empty-label", []string{"--flat", "--label="}, false, 2, false, 0},
		{"blank-labels", []string{"--flat", "--label= , "}, false, 2, false, 0},
		{"mixed-labels", []string{"--flat", "--label= ,a", "--limit=2"}, false, 0, false, 2},
		{"title-not-flag", []string{"--flat", "--title", "--status=closed", "--status=open", "--limit=2"}, false, 0, false, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			jsonOutput = tc.ambient
			cmd := graphListTestCommand(t, tc.args)
			in, structured, err := graphIssueListInput(cmd, append([]string{"list"}, tc.args...))
			if (err != nil) != (tc.exit != 0) {
				t.Fatalf("in=%+v structured=%v error=%v", in, structured, err)
			}
			if tc.exit != 0 {
				var got *exitError
				if !errors.As(err, &got) || got.Code != tc.exit {
					t.Fatalf("expected exit %d, got %v", tc.exit, err)
				}
			}
			if tc.exit == 0 && in.ExcludeTypes != nil {
				t.Fatalf("absent unsupported slice must be canonical zero: %#v", in.ExcludeTypes)
			}
			if tc.exit == 0 && (structured != tc.structured || in.Limit == nil || *in.Limit != tc.limit) {
				t.Fatalf("wrong effective policy: %+v structured=%v", in, structured)
			}
		})
	}
}

func TestGraphPreviewIssueListReplayPreservesFlags(t *testing.T) {
	args := []string{"--flat", "--status=open,closed", "--title", "--status=deferred", "--label=a", "--label=b"}
	cmd := graphListTestCommand(t, args)
	before := cmd.Flags().Lookup("status").Value.String()
	if err := graphIssueListRepeatedFilters(cmd, append([]string{"list"}, args...)); err != nil {
		t.Fatal(err)
	}
	after := cmd.Flags().Lookup("status").Value.String()
	if before != after || !cmd.Flags().Changed("status") {
		t.Fatal("admission changed original parser values")
	}
	labels, _ := cmd.Flags().GetStringSlice("label")
	if strings.Join(labels, ",") != "a,b" {
		t.Fatalf("labels changed: %v", labels)
	}
}

func TestGraphPreviewIssueListOutput(t *testing.T) {
	page := graphstore.IssueListPage{Items: []graphstore.IssueRecord{{ID: "https://example.invalid/a/beads/work", Properties: &types.Issue{Title: "quote\"\n雪", Status: types.StatusOpen, Priority: 2}}}, HasMore: true}
	human, err := renderGraphIssueList(page, "", false, false)
	if err != nil || human != "Issues (1; more: true; graph preview)\n\"https://example.invalid/a/beads/work\" \"open\" P2 \"quote\\\"\\n雪\"\nMore matching Issues exist; increase --limit or use --all within preview bounds.\n" {
		t.Fatalf("human=%q err=%v", human, err)
	}
	quiet, err := renderGraphIssueList(page, "", false, true)
	if err != nil || quiet != "" {
		t.Fatalf("quiet=%q %v", quiet, err)
	}
	output, err := renderGraphIssueList(page, "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	var envelope struct {
		Preview bool                     `json:"preview"`
		Result  graphstore.IssueListPage `json:"result"`
	}
	if err := json.Unmarshal([]byte(output), &envelope); err != nil || !envelope.Preview || !envelope.Result.HasMore || len(envelope.Result.Items) != 1 || envelope.Result.Items[0].Properties.Title != page.Items[0].Properties.Title {
		t.Fatalf("JSON=%s %v", output, err)
	}
	empty, err := renderGraphIssueList(graphstore.IssueListPage{Items: []graphstore.IssueRecord{}}, "", false, false)
	if err != nil || empty != "Issues (0; more: false; graph preview)\n" {
		t.Fatalf("empty=%q %v", empty, err)
	}
	page.Items[0].Properties.Title = strings.Repeat("x", graphIssueListOutputLimit)
	if output, err := renderGraphIssueList(page, "", true, false); err == nil || output != "" {
		t.Fatal("oversized page emitted partial output")
	}
}

func TestRenderGraphIssueTree(t *testing.T) {
	root := &types.Issue{ID: "beads/root", Title: "Root", Status: types.StatusOpen, Priority: 1}
	child := &types.Issue{ID: "beads/child", Title: "Child", Status: types.StatusOpen, Priority: 2}
	grand := &types.Issue{ID: "beads/grand", Title: "Grand", Status: types.StatusInProgress, Priority: 2}
	tree := graphstore.IssueTreeSelected{
		Root: root, Items: []*types.Issue{grand, child},
		Children: map[string][]string{"beads/root": {"beads/child"}, "beads/child": {"beads/grand"}},
	}
	output, err := renderGraphIssueTree(tree, "", "priority", false)
	if err != nil || !strings.Contains(output, "Root") || !strings.Contains(output, "├") && !strings.Contains(output, "└") || strings.Index(output, "Child") > strings.Index(output, "Grand") || !strings.Contains(output, "Total: 3 issues") {
		t.Fatalf("tree output: %q %v", output, err)
	}
	noChildren, err := renderGraphIssueTree(graphstore.IssueTreeSelected{Root: root}, "", "", false)
	if err != nil || noChildren != "Issue 'beads/root' has no children\n" {
		t.Fatalf("empty tree output: %q %v", noChildren, err)
	}
	bad := graphstore.IssueTreeSelected{Root: root, Items: []*types.Issue{child}, Children: map[string][]string{"beads/root": {"beads/missing"}}}
	if output, err := renderGraphIssueTree(bad, "", "", false); err == nil || output != "" {
		t.Fatalf("invalid tree emitted partial output: %q %v", output, err)
	}
	tree.Children["beads/grand"] = []string{"beads/child"}
	cyclic, err := renderGraphIssueTree(tree, "", "", false)
	if err != nil || strings.Count(cyclic, treeCycleMarker) != 1 {
		t.Fatalf("stored parent-child cycle must terminate with one marker: %q %v", cyclic, err)
	}
}

func TestGraphPreviewBeadListOutput(t *testing.T) {
	const scope = "https://example.invalid/a/"
	page := graphstore.BeadListPage{Items: []any{
		graphstore.IssueRecord{ID: scope + "beads/work", Properties: &types.Issue{Title: "Ship the release", Status: types.StatusOpen, Priority: 1}},
		graphstore.Record{ID: scope + "beads/plan", Type: scope + "types/memory", Properties: graphstore.Properties{Title: "Release plan"}},
		graphstore.IssueRecord{ID: scope + "beads/odd", Properties: &types.Issue{Title: "quote\"\n雪", Status: "on-ice", Priority: 4}},
	}}

	// An Issue row adds its status and priority between the kind and the title;
	// the Memory row and the header keep their layout.
	human, err := renderGraphBeadList(page, scope, false, "", false, false)
	want := "Beads (3; more: false; graph preview)\n" +
		"  beads/work  Issue   open  P1  Ship the release\n" +
		"  beads/plan  types/memory  Release plan\n" +
		"  beads/odd  Issue   on-ice  P4  quote\"\\n雪\n"
	if err != nil || human != want {
		t.Fatalf("human=%q err=%v\nwant=%q", human, err, want)
	}

	// The hint offers --all only while it would still help.
	page.HasMore = true
	for _, tc := range []struct {
		all  bool
		hint string
	}{
		{false, "More Beads exist; increase --limit or use --all within preview bounds.\n"},
		{true, "More Beads exist; increase --limit within preview bounds.\n"},
	} {
		got, err := renderGraphBeadList(page, scope, tc.all, "", false, false)
		if err != nil || !strings.HasPrefix(got, "Beads (3; more: true; graph preview)\n") || !strings.HasSuffix(got, "P4  quote\"\\n雪\n"+tc.hint) {
			t.Fatalf("all=%t: %q %v", tc.all, got, err)
		}
	}

	quiet, err := renderGraphBeadList(page, scope, false, "", false, true)
	if err != nil || quiet != "" {
		t.Fatalf("quiet=%q %v", quiet, err)
	}

	// Structured output is the page itself, unchanged: same members, same order.
	structured, err := renderGraphBeadList(page, scope, false, "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	var previous bytes.Buffer
	if err := graphPrintTo(&previous, struct {
		Items   []any `json:"items"`
		HasMore bool  `json:"hasMore"`
	}{page.Items, page.HasMore}, "", false, true); err != nil {
		t.Fatal(err)
	}
	if structured != previous.String() {
		t.Fatalf("structured output changed:\n got=%s\nwant=%s", structured, previous.String())
	}

	empty, err := renderGraphBeadList(graphstore.BeadListPage{Items: []any{}}, scope, false, "", false, false)
	if err != nil || empty != "Beads (0; more: false; graph preview)\n" {
		t.Fatalf("empty=%q %v", empty, err)
	}
}

// Every branch of the routing: which query a list runs and what it says about
// it. A typed Issue option with a non-Issue Bead Type is still refused; a
// configured directory label never is, because it cannot apply to a Memory.
func TestGraphPreviewListPlan(t *testing.T) {
	const (
		issueType  = "https://example.invalid/a/types/preview-issue"
		memoryType = "https://example.invalid/a/types/preview-memory"
	)
	for _, tc := range []struct {
		name         string
		typed        []string
		label        string
		selectedType string
		want         graphListPlan
	}{
		{"nothing selects the Issue query", nil, "", "", graphListPlan{}},
		{"a Memory Type alone", nil, "", memoryType, graphListPlan{}},
		{"a typed option selects the Issue query and says so", []string{"status"}, "", "", graphListPlan{issueQuery: true, notice: "Memories are not listed (Issue query selected by: --status)."}},
		{"typed options are named in the order given", []string{"status", "sort"}, "", "", graphListPlan{issueQuery: true, notice: "Memories are not listed (Issue query selected by: --status, --sort)."}},
		{"the Issue Type already narrowed the list to Issues", []string{"status"}, "", issueType, graphListPlan{issueQuery: true}},
		{"a typed option with a Memory Type is refused", []string{"status"}, "", memoryType, graphListPlan{refuse: true}},
		{"a typed option and the configured label are both named", []string{"status"}, "frontend", "", graphListPlan{issueQuery: true, notice: "Memories are not listed (Issue query selected by: --status, directory.labels \"frontend\")."}},
		{"the configured label alone selects the Issue query and says so", nil, "frontend", "", graphListPlan{issueQuery: true, notice: "Memories are not listed (Issue query selected by: directory.labels \"frontend\")."}},
		{"the configured label with the Issue Type", nil, "frontend", issueType, graphListPlan{issueQuery: true}},
		{"the configured label never refuses a Memory Type", nil, "frontend", memoryType, graphListPlan{notice: "directory.labels \"frontend\" is not applied to this Bead Type."}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := planGraphList(tc.typed, tc.label, tc.selectedType, issueType); got != tc.want {
				t.Fatalf("plan %+v, want %+v", got, tc.want)
			}
		})
	}
}

// The notice names the options the caller typed, in canonical order whatever
// order they were typed in, and names the configured label only while no typed
// label has replaced it.
func TestGraphPreviewIssueListSelection(t *testing.T) {
	typed := graphIssueListTypedOptions(graphListTestCommand(t, []string{"--sort", "title", "--status", "open", "--pinned", "--limit", "5", "--all"}))
	if want := []string{"status", "pinned", "sort"}; !slices.Equal(typed, want) {
		t.Fatalf("typed options %v, want %v: only Issue options count, in canonical order", typed, want)
	}
	if typed := graphIssueListTypedOptions(graphListTestCommand(t, []string{"--assignee="})); !slices.Equal(typed, []string{"assignee"}) {
		t.Fatalf("a supplied empty value must still count, got %v", typed)
	}
	if typed := graphIssueListTypedOptions(graphListTestCommand(t, []string{"--limit", "5", "--all", "--flat"})); len(typed) != 0 {
		t.Fatalf("options that are not Issue options selected the Issue query: %v", typed)
	}

	configured := listInput{}
	configured.LabelsAny = []string{"frontend"}
	for _, tc := range []struct {
		name string
		args []string
		in   listInput
		want string
	}{
		{"configured and untyped", nil, configured, "frontend"},
		{"a typed --label replaces it", []string{"--label", "x"}, configured, ""},
		{"a typed --label-any replaces it", []string{"--label-any", "x"}, configured, ""},
		{"nothing configured", nil, listInput{}, ""},
	} {
		if got := graphIssueListConfigLabel(graphListTestCommand(t, tc.args), tc.in); got != tc.want {
			t.Fatalf("%s: configured label %q, want %q", tc.name, got, tc.want)
		}
	}
}

// The notice is one line directly under the header, in human output only.
func TestGraphPreviewListNoticeRendering(t *testing.T) {
	const scope = "https://example.invalid/a/"
	const notice = "Memories are not listed (Issue query selected by: --status)."
	issues := graphstore.IssueListPage{Items: []graphstore.IssueRecord{{ID: scope + "beads/work", Properties: &types.Issue{Title: "Work", Status: types.StatusOpen, Priority: 1}}}}
	human, err := renderGraphIssueList(issues, notice, false, false)
	if want := "Issues (1; more: false; graph preview)\n" + notice + "\n\"" + scope + "beads/work\" \"open\" P1 \"Work\"\n"; err != nil || human != want {
		t.Fatalf("human=%q err=%v\nwant=%q", human, err, want)
	}
	if quiet, err := renderGraphIssueList(issues, notice, false, true); err != nil || quiet != "" {
		t.Fatalf("quiet=%q %v", quiet, err)
	}
	withNotice, err := renderGraphIssueList(issues, notice, true, false)
	if err != nil {
		t.Fatal(err)
	}
	if withoutNotice, err := renderGraphIssueList(issues, "", true, false); err != nil || withNotice != withoutNotice {
		t.Fatalf("structured output must not carry the notice: %q vs %q (%v)", withNotice, withoutNotice, err)
	}

	const labelNotice = "directory.labels \"frontend\" is not applied to this Bead Type."
	beads := graphstore.BeadListPage{Items: []any{graphstore.Record{ID: scope + "beads/plan", Type: scope + "types/memory", Properties: graphstore.Properties{Title: "Plan"}}}}
	human, err = renderGraphBeadList(beads, scope, false, labelNotice, false, false)
	if want := "Beads (1; more: false; graph preview)\n" + labelNotice + "\n  beads/plan  types/memory  Plan\n"; err != nil || human != want {
		t.Fatalf("human=%q err=%v\nwant=%q", human, err, want)
	}
	if quiet, err := renderGraphBeadList(beads, scope, false, labelNotice, false, true); err != nil || quiet != "" {
		t.Fatalf("quiet=%q %v", quiet, err)
	}
	structured, err := renderGraphBeadList(beads, scope, false, labelNotice, true, false)
	if plain, perr := renderGraphBeadList(beads, scope, false, "", true, false); err != nil || perr != nil || structured != plain {
		t.Fatalf("structured output must not carry the notice: %q vs %q", structured, plain)
	}
}

// Defer filters remain outside this graph adapter even when supplied empty
// or false. Admitted due filters must not bypass that refusal.
func TestGraphPreviewIssueDeferredFilters(t *testing.T) {
	old := jsonOutput
	jsonOutput = false
	t.Cleanup(func() { jsonOutput = old })
	for _, flags := range [][]string{
		{"--defer-before=2030-01-01"}, {"--defer-before="},
		{"--defer-after=2030-01-01"}, {"--deferred"}, {"--deferred=false", "--overdue"},
	} {
		args := append([]string{"--format=records-json"}, flags...)
		cmd := graphListTestCommand(t, nil)
		cmd.Flags().String("defer-before", "", "")
		cmd.Flags().String("defer-after", "", "")
		cmd.Flags().Bool("deferred", false, "")
		if err := cmd.ParseFlags(args); err != nil {
			t.Fatal(err)
		}
		_, _, err := graphIssueListInput(cmd, append([]string{"list"}, args...))
		var failure *exitError
		if !errors.As(err, &failure) || failure.Code != 5 {
			t.Fatalf("deferred filter %v reached query: %v", flags, err)
		}
	}
}

func TestGraphPreviewIssueAssigneeListInput(t *testing.T) {
	t.Chdir(t.TempDir())
	config.ResetForTesting()
	t.Cleanup(config.ResetForTesting)
	if err := config.Initialize(); err != nil {
		t.Fatal(err)
	}
	old := jsonOutput
	jsonOutput = false
	t.Cleanup(func() { jsonOutput = old })
	for _, tc := range []struct {
		name       string
		args       []string
		assignee   string
		unassigned bool
		exit       int
	}{
		{"literal", []string{"--assignee=  雪 / agent  "}, "  雪 / agent  ", false, 0},
		{"short", []string{"-aalice"}, "alice", false, 0},
		{"empty", []string{"--assignee="}, "", false, 0},
		{"unassigned", []string{"--no-assignee"}, "", true, 0},
		{"false-unassigned", []string{"--no-assignee=false"}, "", false, 0},
		{"intersection", []string{"--assignee=alice", "--no-assignee"}, "alice", true, 0},
		{"empty-intersection", []string{"--assignee=", "--no-assignee"}, "", true, 0},
		{"value-not-flag", []string{"--title", "--assignee=bob", "--assignee=alice"}, "alice", false, 0},
		{"lookup-beyond-field-bound", []string{"--assignee=" + strings.Repeat("a", 256)}, strings.Repeat("a", 256), false, 0},
		{"repeated-long", []string{"--assignee=alice", "--assignee=alice"}, "", false, 5},
		{"repeated-short", []string{"-aalice", "-a", "bob"}, "", false, 5},
		{"repeated-mixed-empty", []string{"--assignee=", "-abob"}, "", false, 5},
		{"repeated-empty", []string{"--assignee=", "--assignee="}, "", false, 5},
		{"utf8", []string{"--assignee=\xff"}, "", false, 2},
		{"request-limit", []string{"--assignee=" + strings.Repeat("a", 4097)}, "", false, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			args := append([]string{"--format=records-json", "--all"}, tc.args...)
			cmd := graphListTestCommand(t, args)
			in, structured, err := graphIssueListInput(cmd, append([]string{"list"}, args...))
			if tc.exit != 0 {
				var failure *exitError
				if !errors.As(err, &failure) || failure.Code != tc.exit {
					t.Fatalf("wantexit%d got %v", tc.exit, err)
				}
				return
			}
			if err != nil || !structured || in.Assignee != tc.assignee || in.NoAssignee != tc.unassigned {
				t.Fatalf("filter changed: %+v %v", in, err)
			}
			value, _ := cmd.Flags().GetString("assignee")
			if value != tc.assignee {
				t.Fatal("occurrence replay changed original assignee flag")
			}
		})
	}
}
