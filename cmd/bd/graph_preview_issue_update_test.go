package main

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
)

func issueTextCommand(t *testing.T, args ...string) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{}
	cmd.Flags().IntP("estimate", "e", 0, "")
	for _, name := range append(append([]string{}, graphPreviewIssueEditFlags...), "properties", "design-file", "session", "if-assignee", "if-status", "if-revision", "if-source-revision") {
		if cmd.Flags().Lookup(name) == nil {
			if name == "clear-notes" || name == "stdin" {
				cmd.Flags().Bool(name, false, "")
			} else {
				cmd.Flags().String(name, "", "")
			}
		}
	}
	cmd.Flags().Bool("allow-empty-description", false, "")
	cmd.Flags().Bool("claim", false, "")
	cmd.Flags().Bool("force", false, "")
	if err := cmd.ParseFlags(args); err != nil {
		t.Fatal(err)
	}
	return cmd
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
	request, err = graphPreviewIssueEditRequest(issueTextCommand(t, "--design="), "beads/task")
	if err != nil || request.Design == nil || *request.Design != "" || request.Title != nil || request.Description != nil || request.AcceptanceCriteria != nil || !request.Unconditional {
		t.Fatalf("omitted fields did not remain absent: %+v %v", request, err)
	}
}

func TestGraphPreviewIssueStatusIntent(t *testing.T) {
	for _, tc := range []struct {
		args        []string
		wantStatus  types.Status
		wantForce   bool
		wantSession string
	}{
		{[]string{"--status=in_progress"}, types.StatusInProgress, false, ""},
		{[]string{"--status=closed", "--force", "--session=review"}, types.StatusClosed, true, "review"},
		{[]string{"--status=closed", "--force=false"}, types.StatusClosed, false, ""},
	} {
		cmd := issueTextCommand(t, tc.args...)
		request, err := graphPreviewIssueEditRequest(cmd, "beads/work")
		if err != nil || request.Status == nil || *request.Status != tc.wantStatus || request.ForceClosePolicy != tc.wantForce {
			t.Fatalf("status intent lost: %+v %v", request, err)
		}
		if tc.wantSession != "" && (request.ClosedBySession == nil || *request.ClosedBySession != tc.wantSession) {
			t.Fatalf("session attribution lost: %+v", request)
		}
	}
}

func TestGraphPreviewIssuePropertyMerge(t *testing.T) {
	for _, tc := range []struct {
		name      string
		args      []string
		wantError bool
	}{
		{"empty", []string{"--properties={}", "--if-revision=seen"}, false},
		{"typed", []string{"--properties={\"title\":\"New\",\"priority\":2,\"estimated_minutes\":7}", "--if-revision=seen"}, false},
		{"clear-estimate", []string{"--properties={\"estimated_minutes\":null}", "--if-revision=seen"}, false},
		{"duplicate", []string{"--title=New", "--properties={\"title\":\"New\"}", "--if-revision=seen"}, true},
		{"duplicate-alias", []string{"--body=New", "--properties={\"description\":\"New\"}", "--if-revision=seen"}, true},
		{"read-only", []string{"--properties={\"status\":\"closed\"}", "--if-revision=seen"}, true},
		{"wrong-type", []string{"--properties={\"priority\":\"2\"}", "--if-revision=seen"}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := issueTextCommand(t, tc.args...)
			request, err := graphPreviewIssueEditRequest(cmd, "beads/task")
			if err == nil {
				err = graphPreviewIssuePropertyMerge(cmd, &request)
			}
			if (err != nil) != tc.wantError {
				t.Fatalf("request=%+v error=%v", request, err)
			}
			if err == nil && !request.PropertiesProvided {
				t.Fatal("property presence was lost")
			}
			if tc.name == "typed" && (request.Title == nil || *request.Title != "New" || request.Priority == nil || *request.Priority != 2 || request.EstimatedMinutes == nil || *request.EstimatedMinutes != 7) {
				t.Fatalf("typed values lost: %+v", request)
			}
			if tc.name == "clear-estimate" && !request.ClearEstimatedMinutes {
				t.Fatal("nullable estimate clear lost")
			}
		})
	}
}

func TestGraphPreviewIssueNotesInput(t *testing.T) {
	for _, tc := range []struct {
		name       string
		args       []string
		wantNotes  string
		wantForce  bool
		wantAppend bool
	}{
		{"replace", []string{"--notes=New", "--force", "--if-revision=seen"}, "New", true, false},
		{"clear", []string{"--clear-notes", "--if-revision=seen"}, "", false, false},
		{"append", []string{"--append-notes=Next", "--if-revision=seen"}, "", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := issueTextCommand(t, tc.args...)
			if !graphPreviewIssueEditFlagsChanged(cmd) {
				t.Fatal("notes-only edit did not select the Issue route")
			}
			request, err := graphPreviewIssueEditRequest(cmd, "beads/task")
			if err != nil || request.ExpectedRevision != "seen" || request.ForceNotesOverwrite != tc.wantForce || (request.AppendNotes != nil) != tc.wantAppend {
				t.Fatalf("notes mode or guard lost: %+v %v", request, err)
			}
			if tc.wantAppend {
				if request.Notes != nil || *request.AppendNotes != "Next" {
					t.Fatalf("append became replacement: %+v", request)
				}
			} else if request.Notes == nil || *request.Notes != tc.wantNotes || request.AppendNotes != nil {
				t.Fatalf("replacement/clear lost presence: %+v", request)
			}
		})
	}
}

func TestGraphPreviewIssueTextRefusals(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		code int
	}{
		{"missing-fields", []string{}, 2},
		{"empty-guard", []string{"--title=Title", "--if-revision="}, 2},
		{"blank-title", []string{"--title=  "}, 2},
		{"long-title", []string{"--title=" + strings.Repeat("x", 501)}, 2},
		{"invalid-utf8", []string{"--design=\xff"}, 2},
		{"different-aliases", []string{"--description=a", "--body=b"}, 2},
		{"empty-notes", []string{"--notes="}, 2},
		// Append still cannot be combined with replacement notes.
		{"append", []string{"--design=Design", "--append-notes=More", "--notes=Replace"}, 2},
		{"notes-clear", []string{"--notes=Replace", "--clear-notes"}, 2},
		{"append-clear", []string{"--append-notes=More", "--clear-notes"}, 2},
		{"clear-false", []string{"--clear-notes=false"}, 2},
		{"force-without-notes", []string{"--title=Title", "--force"}, 2},
		{"source-guard", []string{"--design=Design", "--if-source-revision=other"}, 5},
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
	request, err := graphPreviewIssueEditRequest(issueTextCommand(t, "--priority=P0", "--title=  Urgent  ", "--description="), "beads/work")
	if err != nil || request.Priority == nil || *request.Priority != 0 || request.Title == nil || *request.Title != "Urgent" || request.Description == nil || *request.Description != "" || !request.Unconditional {
		t.Fatalf("combined edit lost: %+v %v", request, err)
	}
	request, err = graphPreviewIssueEditRequest(issueTextCommand(t, "--title=Text only"), "beads/work")
	if err != nil || request.Priority != nil {
		t.Fatalf("omitted priority became explicit: %+v %v", request, err)
	}
}

func TestGraphPreviewIssuePriorityRefusals(t *testing.T) {
	for _, value := range []string{"", "-1", "5", "P5", "high", "P", "\xff"} {
		t.Run(fmt.Sprintf("value-%q", value), func(t *testing.T) {
			_, err := graphPreviewIssueEditRequest(issueTextCommand(t, "--priority="+value), "beads/work")
			var failure *exitError
			if !errors.As(err, &failure) || failure.Code != 2 {
				t.Fatalf("invalid priority must refuse with exit2: %v", err)
			}
		})
	}
	request, err := graphPreviewIssueEditRequest(issueTextCommand(t, "--priority=0", "--stdin=false"), "beads/work")
	if err != nil || request.Priority == nil || *request.Priority != 0 || request.Description != nil {
		t.Fatalf("false stdin flag changed priority edit: %+v %v", request, err)
	}
	_, err = graphPreviewIssueEditRequest(issueTextCommand(t, "--priority=0", "--body-file=missing"), "beads/work")
	var missing *exitError
	if !errors.As(err, &missing) || missing.Code != 1 {
		t.Fatalf("missing body file must refuse before storage: %v", err)
	}
	cmd := issueTextCommand(t, "--priority=0", "--properties={}")
	request, err = graphPreviewIssueEditRequest(cmd, "beads/work")
	if err != nil || !cmd.Flags().Changed("properties") || request.Priority == nil || *request.Priority != 0 {
		t.Fatalf("mixed priority and empty property merge: %+v %v", request, err)
	}
}

func TestGraphPreviewIssueAssigneePresence(t *testing.T) {
	for _, value := range []string{"", "alice", "  雪 / agent  ", strings.Repeat("雪", types.MaxFieldLen)} {
		t.Run(fmt.Sprintf("value-%q", value), func(t *testing.T) {
			cmd := issueTextCommand(t, "--assignee="+value, "--if-revision=observed")
			request, err := graphPreviewIssueEditRequest(cmd, "beads/work")
			if err != nil || !graphPreviewIssueEditFlagsChanged(cmd) || request.Assignee == nil || *request.Assignee != value || request.ExpectedRevision != "observed" || request.Unconditional {
				t.Fatalf("assignment/presence/guard lost: %+v %v", request, err)
			}
		})
	}
	request, err := graphPreviewIssueEditRequest(issueTextCommand(t, "--title=Assigned", "--priority=P0", "--assignee=alice"), "beads/work")
	if err != nil || request.Assignee == nil || *request.Assignee != "alice" || request.Title == nil || *request.Title != "Assigned" || request.Priority == nil || *request.Priority != 0 || !request.Unconditional {
		t.Fatalf("mixed fields lost: %+v %v", request, err)
	}
	request, err = graphPreviewIssueEditRequest(issueTextCommand(t, "--priority=2"), "beads/work")
	if err != nil || request.Assignee != nil {
		t.Fatalf("omission became assignment: %+v %v", request, err)
	}
	for _, tc := range []struct {
		args          []string
		forceAssignee bool
		forceNotes    bool
	}{
		{[]string{"--assignee=bob", "--force"}, true, false},
		{[]string{"--assignee=bob", "--force=false"}, false, false},
		{[]string{"--assignee=bob", "--notes=Revised", "--force"}, true, true},
	} {
		request, err := graphPreviewIssueEditRequest(issueTextCommand(t, tc.args...), "beads/work")
		if err != nil || request.ForceAssigneeTransfer != tc.forceAssignee || request.ForceNotesOverwrite != tc.forceNotes {
			t.Fatalf("force intent %v: %+v %v", tc.args, request, err)
		}
	}
	cmd := issueTextCommand(t, `--properties={"assignee":"bob"}`, "--force")
	request, err = graphPreviewIssueEditRequest(cmd, "beads/work")
	if err == nil {
		err = graphPreviewIssuePropertyMerge(cmd, &request)
	}
	if err != nil || request.Assignee == nil || *request.Assignee != "bob" || !request.ForceAssigneeTransfer || request.ForceNotesOverwrite {
		t.Fatalf("property assignee force intent: %+v %v", request, err)
	}
}

func TestGraphPreviewIssueAssigneeRefusals(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		code int
	}{
		{"utf8", []string{"--assignee=\xff"}, 2},
		{"length", []string{"--assignee=" + strings.Repeat("雪", types.MaxFieldLen+1)}, 2},
		{"empty-guard", []string{"--assignee=alice", "--if-revision="}, 2},
		{"claim", []string{"--assignee=alice", "--claim"}, 5},
		{"false-claim", []string{"--assignee=alice", "--claim=false"}, 5},
		{"assignee-precondition", []string{"--assignee=alice", "--if-assignee="}, 5},
		{"status-precondition", []string{"--assignee=alice", "--if-status=open"}, 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := graphPreviewIssueEditRequest(issueTextCommand(t, tc.args...), "beads/work")
			var failure *exitError
			if !errors.As(err, &failure) || failure.Code != tc.code {
				t.Fatalf("expected exit%d, got %v", tc.code, err)
			}
		})
	}
}

func TestGraphPreviewHistoricalPropertiesForceDefersToComposedNotes(t *testing.T) {
	for _, historical := range []bool{false, true} {
		t.Run(fmt.Sprint(historical), func(t *testing.T) {
			cmd := issueTextCommand(t, "--force")
			request := graphstore.UpdateIssueRequest{Path: "beads/work"}
			if historical {
				request.FromRevision = "opaque-source"
			}
			err := graphPreviewApplyIssueProperties(cmd, &request, map[string]any{"title": "Explicit"})
			if historical && err != nil {
				t.Fatalf("historical notes setter is not available until storage composition: %v", err)
			}
			if !historical && err == nil {
				t.Fatal("ordinary non-notes force guard was weakened")
			}
		})
	}
}
