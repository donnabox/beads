package main

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/types"
)

func issueTextCommand(t *testing.T, args ...string) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{}
	for _, name := range append(append([]string{}, graphPreviewIssueEditFlags...), "properties", "notes", "body-file", "design-file", "append-notes", "status", "if-revision", "if-source-revision") {
		if name == "set-labels" {
			cmd.Flags().StringSlice(name, nil, "")
			continue
		}
		cmd.Flags().String(name, "", "")
	}
	cmd.Flags().Bool("unconditional", false, "")
	cmd.Flags().Bool("stdin", false, "")
	if err := cmd.ParseFlags(args); err != nil {
		t.Fatal(err)
	}
	return cmd
}

func TestGraphPreviewIssuePriorityInput(t *testing.T) {
	for _, tc := range []struct {
		input string
		want  int
	}{
		{"0", 0}, {"1", 1}, {"2", 2}, {"3", 3}, {"4", 4},
		{"P0", 0}, {"p2", 2}, {" P3 ", 3},
		// Reuse ordinary update's existing permissive parser; no new grammar.
		{"2suffix", 2},
	} {
		t.Run(tc.input, func(t *testing.T) {
			cmd := issueTextCommand(t, "--priority="+tc.input, "--if-revision=observed")
			if !graphPreviewIssueEditFlagsChanged(cmd) {
				t.Fatal("priority-only update did not select Issue route")
			}
			request, err := graphPreviewIssueEditRequest(cmd, "beads/work")
			if err != nil || request.Priority == nil || *request.Priority != tc.want || request.Title != nil || request.ExpectedRevision != "observed" {
				t.Fatalf("priority/presence/guard lost: %+v %v", request, err)
			}
		})
	}
	request, err := graphPreviewIssueEditRequest(issueTextCommand(t, "--priority=P0", "--title=  Urgent  ", "--description=", "--unconditional"), "beads/work")
	if err != nil || request.Priority == nil || *request.Priority != 0 || request.Title == nil || *request.Title != "Urgent" || request.Description == nil || *request.Description != "" || !request.Unconditional {
		t.Fatalf("combined edit lost: %+v %v", request, err)
	}
	request, err = graphPreviewIssueEditRequest(issueTextCommand(t, "--title=Text only", "--unconditional"), "beads/work")
	if err != nil || request.Priority != nil {
		t.Fatalf("omitted priority became explicit: %+v %v", request, err)
	}
}

func TestGraphPreviewIssuePriorityRefusals(t *testing.T) {
	for _, value := range []string{"", "-1", "5", "P5", "high", "P", "\xff"} {
		t.Run(fmt.Sprintf("value-%q", value), func(t *testing.T) {
			_, err := graphPreviewIssueEditRequest(issueTextCommand(t, "--priority="+value, "--unconditional"), "beads/work")
			var failure *exitError
			if !errors.As(err, &failure) || failure.Code != 2 {
				t.Fatalf("invalid priority must refuse with exit2: %v", err)
			}
		})
	}
	for _, args := range [][]string{
		{"--priority=0", "--unconditional", "--properties={}"},
		{"--priority=0", "--unconditional", "--notes="},
		{"--priority=0", "--unconditional", "--status=open"},
		{"--priority=0", "--unconditional", "--body-file=missing"},
		{"--priority=0", "--unconditional", "--stdin=false"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			cmd := issueTextCommand(t, args...)
			cmd.SetIn(failingMemoryBodyReader{})
			_, err := graphPreviewIssueEditRequest(cmd, "beads/work")
			var failure *exitError
			if !errors.As(err, &failure) || failure.Code != 5 {
				t.Fatalf("unsupported explicit option must refuse with exit5 before input: %v", err)
			}
		})
	}
}

func TestGraphPreviewIssueTextPresenceAndAliases(t *testing.T) {
	cmd := issueTextCommand(t, "--title=  Revised 雪  ", "--description=", "--body=", "--message=", "--design=-", "--acceptance=Done", "--if-revision=observed")
	request, err := graphPreviewIssueEditRequest(cmd, "beads/task")
	if err != nil {
		t.Fatal(err)
	}
	if *request.Title != "Revised 雪" || *request.Description != "" || *request.Design != "-" || *request.AcceptanceCriteria != "Done" || request.ExpectedRevision != "observed" || request.Unconditional {
		t.Fatalf("text or presence changed: %+v", request)
	}
	request, err = graphPreviewIssueEditRequest(issueTextCommand(t, "--design=", "--unconditional"), "beads/task")
	if err != nil || request.Design == nil || *request.Design != "" || request.Title != nil || request.Description != nil || request.AcceptanceCriteria != nil {
		t.Fatalf("omitted fields did not remain absent: %+v %v", request, err)
	}
}

func TestGraphPreviewIssueTextRefusals(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		code int
	}{
		{"missing-fields", []string{"--unconditional"}, 2},
		{"missing-guard", []string{"--title=Title"}, 2},
		{"false-unconditional", []string{"--title=Title", "--unconditional=false"}, 2},
		{"mixed-guards", []string{"--title=Title", "--unconditional", "--if-revision=old"}, 2},
		{"blank-title", []string{"--title=  ", "--unconditional"}, 2},
		{"long-title", []string{"--title=" + strings.Repeat("x", 501), "--unconditional"}, 2},
		{"invalid-utf8", []string{"--design=\xff", "--unconditional"}, 2},
		{"different-aliases", []string{"--description=a", "--body=b", "--unconditional"}, 2},
		{"generic-properties", []string{"--title=Title", "--properties={}", "--unconditional"}, 5},
		{"workflow", []string{"--design=Design", "--status=closed", "--unconditional"}, 5},
		{"file", []string{"--design=Design", "--body-file=missing", "--unconditional"}, 5},
		{"stdin-flag", []string{"--design=Design", "--stdin", "--unconditional"}, 5},
		{"stdin-description", []string{"--description=-", "--unconditional"}, 5},
		{"stdin-alias", []string{"--body=-", "--unconditional"}, 5},
		{"notes", []string{"--notes=Notes", "--unconditional"}, 5},
		{"append", []string{"--design=Design", "--append-notes=More", "--unconditional"}, 5},
		{"source-guard", []string{"--design=Design", "--if-source-revision=other", "--unconditional"}, 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := issueTextCommand(t, tc.args...)
			cmd.SetIn(failingMemoryBodyReader{})
			_, err := graphPreviewIssueEditRequest(cmd, "beads/task")
			var failure *exitError
			if !errors.As(err, &failure) || failure.Code != tc.code {
				t.Fatalf("expected refusal %d, got %v", tc.code, err)
			}
		})
	}
}

func TestGraphPreviewIssueLabelReplacement(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		want []string
	}{
		{"normalized", []string{"--set-labels= alpha,alpha,Alpha,café,cafe,雪 ", "--set-labels=beta"}, []string{"alpha", "Alpha", "café", "cafe", "雪", "beta"}},
		{"clear", []string{"--set-labels="}, []string{}},
		{"empty-entries", []string{"--set-labels= , ,"}, []string{}},
		{"comma-label", []string{`--set-labels="one,two",three`}, []string{"one,two", "three"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := issueTextCommand(t, append(tc.args, "--if-revision=observed")...)
			if !graphPreviewIssueEditFlagsChanged(cmd) {
				t.Fatal("labels-only update missed Issue route")
			}
			r, err := graphPreviewIssueEditRequest(cmd, "beads/work")
			if err != nil || r.Labels == nil || !reflect.DeepEqual(*r.Labels, tc.want) || r.ExpectedRevision != "observed" {
				t.Fatalf("label request: %+v %v", r, err)
			}
		})
	}
	r, err := graphPreviewIssueEditRequest(issueTextCommand(t, "--title=Text only", "--unconditional"), "beads/work")
	if err != nil || r.Labels != nil {
		t.Fatalf("omitted labels became a clear: %+v %v", r, err)
	}
	r, err = graphPreviewIssueEditRequest(issueTextCommand(t, "--title=Together", "--priority=P0", "--set-labels=ready", "--unconditional"), "beads/work")
	if err != nil || r.Labels == nil || !reflect.DeepEqual(*r.Labels, []string{"ready"}) || r.Priority == nil || *r.Priority != 0 || r.Title == nil || *r.Title != "Together" {
		t.Fatalf("mixed patch: %+v %v", r, err)
	}
}

func TestGraphPreviewIssueLabelRefusals(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		code int
	}{
		{"guard", []string{"--set-labels=a"}, 2},
		{"false-unconditional", []string{"--set-labels=a", "--unconditional=false"}, 2},
		{"both-guards", []string{"--set-labels=a", "--unconditional", "--if-revision=old"}, 2},
		{"utf8", []string{"--set-labels=\xff", "--unconditional"}, 2},
		{"overlength", []string{"--set-labels=" + strings.Repeat("x", types.MaxFieldLen+1), "--unconditional"}, 2},
		{"file", []string{"--set-labels=a", "--body-file=missing", "--unconditional"}, 5},
		{"properties", []string{"--set-labels=a", "--properties={}", "--unconditional"}, 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := issueTextCommand(t, tc.args...)
			cmd.SetIn(failingMemoryBodyReader{})
			_, err := graphPreviewIssueEditRequest(cmd, "beads/work")
			var failure *exitError
			if !errors.As(err, &failure) || failure.Code != tc.code {
				t.Fatalf("wanted %d: %v", tc.code, err)
			}
		})
	}
}
