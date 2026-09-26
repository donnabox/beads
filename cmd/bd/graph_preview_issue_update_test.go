package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func issueTextCommand(t *testing.T, args ...string) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{}
	for _, name := range append(append([]string{}, graphPreviewIssueTextFlags...), "properties", "notes", "body-file", "design-file", "append-notes", "status", "if-revision", "if-source-revision") {
		cmd.Flags().String(name, "", "")
	}
	cmd.Flags().Bool("unconditional", false, "")
	cmd.Flags().Bool("stdin", false, "")
	if err := cmd.ParseFlags(args); err != nil {
		t.Fatal(err)
	}
	return cmd
}

func TestGraphPreviewIssueTextPresenceAndAliases(t *testing.T) {
	cmd := issueTextCommand(t, "--title=  Revised 雪  ", "--description=", "--body=", "--message=", "--design=-", "--acceptance=Done", "--if-revision=observed")
	request, err := graphPreviewIssueTextRequest(cmd, "beads/task")
	if err != nil {
		t.Fatal(err)
	}
	if *request.Title != "Revised 雪" || *request.Description != "" || *request.Design != "-" || *request.AcceptanceCriteria != "Done" || request.ExpectedRevision != "observed" || request.Unconditional {
		t.Fatalf("text or presence changed: %+v", request)
	}
	request, err = graphPreviewIssueTextRequest(issueTextCommand(t, "--design=", "--unconditional"), "beads/task")
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
			_, err := graphPreviewIssueTextRequest(cmd, "beads/task")
			var failure *exitError
			if !errors.As(err, &failure) || failure.Code != tc.code {
				t.Fatalf("expected refusal %d, got %v", tc.code, err)
			}
		})
	}
}
