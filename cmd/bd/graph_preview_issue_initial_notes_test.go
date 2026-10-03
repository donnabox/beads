//go:build cgo

package main

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/types"
)

func TestGraphPreviewIssueInitialNotesInput(t *testing.T) {
	for _, tc := range []struct {
		name, value string
		supplied    bool
		invalid     bool
	}{
		{name: "omitted"},
		{name: "empty", supplied: true},
		{name: "literal", value: "  雪\r\ncafé e\u0301\n  ", supplied: true},
		{name: "hyphen", value: "-", supplied: true},
		{name: "invalid-utf8", value: "\xff", supplied: true, invalid: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := issueCreateFieldsCommand(t)
			if tc.supplied {
				if err := cmd.ParseFlags([]string{"--notes=" + tc.value}); err != nil {
					t.Fatal(err)
				}
			}
			i := types.Issue{Title: "Preserve", Status: types.StatusOpen, Owner: "supplied-owner", CreatedBy: "supplied-creator"}
			var err error
			diagnostic := captureStderr(t, func() { err = graphPreviewIssueCreateFields(cmd, &i) })
			if tc.invalid {
				var failure *exitError
				if !errors.As(err, &failure) || failure.Code != 2 || !strings.Contains(diagnostic, "invalid_properties") {
					t.Fatalf("invalid notes not refused before storage: %v %q", err, diagnostic)
				}
				return
			}
			if err != nil || i.Notes != tc.value || i.Title != "Preserve" || i.Status != types.StatusOpen || i.Owner != "supplied-owner" || i.CreatedBy != "supplied-creator" {
				t.Fatalf("literal input/default boundary changed: %+v %v", i, err)
			}
		})
	}
}

func TestGraphPreviewIssueCreateAuthorshipDispatch(t *testing.T) {
	bd := buildBDUnderTest(t)
	work, home := t.TempDir(), t.TempDir()
	gitconfig := filepath.Join(home, "missing-gitconfig") // graphPolicyCLI's isolated global config.
	graphPolicyCLI(t, bd, work, home, nil, "", "init", "--graph-mode", "link", "--prefix", "auth", "--scope-url", "https://example.invalid/initial-notes/", "--skip-hooks", "--skip-agents", "--non-interactive", "--json")
	for _, tc := range []struct {
		name, flagActor, beadsActor, oldActor, ownerEnv, gitEmail, wantActor, wantOwner string
	}{
		{"explicit", "explicit-creator", "beads-creator", "old-creator", "  env-owner@example.invalid  ", "git-owner@example.invalid", "explicit-creator", "  env-owner@example.invalid  "},
		{"primary-env", "", "beads-creator", "old-creator", "", "git-owner@example.invalid", "beads-creator", "git-owner@example.invalid"},
		{"deprecated-env", "", "", "old-creator", "", "git-owner@example.invalid", "old-creator", "git-owner@example.invalid"},
		{"git-fallback", "", "", "", "", "git-owner@example.invalid", "Git Creator", "git-owner@example.invalid"},
		{"missing-owner", "", "", "", "", "", "Git Creator", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			gitText := "[user]\n\tname = Git Creator\n"
			if tc.gitEmail != "" {
				gitText += "\temail = " + tc.gitEmail + "\n"
			}
			writeFile(t, gitconfig, []byte(gitText))
			env := []string{"BEADS_ACTOR=" + tc.beadsActor, "BD_ACTOR=" + tc.oldActor, "GIT_AUTHOR_EMAIL=" + tc.ownerEnv}
			path := "beads/" + tc.name
			args := []string{"create", "Authored Issue", "--id", path, "--notes", "  Initial 雪\r\n ", "--assignee", "assigned-worker", "--json"}
			if tc.flagActor != "" {
				args = append(args, "--actor", tc.flagActor)
			}
			created := graphPolicyCLI(t, bd, work, home, env, "", args...)
			var envelope struct {
				Result struct {
					Version     string
					Properties  types.Issue
					Attribution struct{ Actor string }
				}
			}
			if err := json.Unmarshal([]byte(created), &envelope); err != nil {
				t.Fatal(err)
			}
			i := envelope.Result.Properties
			if envelope.Result.Version == "" || i.CreatedBy != tc.wantActor || i.Owner != tc.wantOwner || envelope.Result.Attribution.Actor != tc.wantActor || i.Assignee != "assigned-worker" || i.Status != types.StatusOpen || i.Notes != "  Initial 雪\r\n " || i.LeaseExpiresAt != nil || i.HeartbeatAt != nil {
				t.Fatalf("authorship, initial notes or no-claim boundary lost: %+v", envelope.Result)
			}
			for _, flags := range [][]string{nil, {"--version", envelope.Result.Version}} {
				if got := graphPolicyCLI(t, bd, work, home, nil, "", append([]string{"show", path, "--json"}, flags...)...); got != created {
					t.Fatal("fresh current/exact read changed complete initial authored Issue")
				}
			}
			graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "update", path, "--notes=", "--if-revision", envelope.Result.Version, "--json")
			if got := graphPolicyCLI(t, bd, work, home, nil, "", "show", path, "--json"); got != created {
				t.Fatal("held notes clear changed initial Issue")
			}
		})
	}
	for _, key := range []string{"BEADS_ACTOR", "GIT_AUTHOR_EMAIL"} {
		t.Run("overlong-"+key, func(t *testing.T) {
			env := []string{key + "=" + strings.Repeat("雪", 256)}
			graphPolicyCLI(t, bd, work, home, env, "invalid_properties", "create", "Refused", "--id", "beads/refused", "--notes=No allocation", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "not_found", "show", "beads/refused", "--json")
		})
	}
	graphPolicyCLI(t, bd, work, home, nil, "permission_denied", "create", "Read-only", "--id", "beads/refused", "--notes", "must not allocate", "--readonly", "--json")
	graphPolicyCLI(t, bd, work, home, nil, "not_found", "show", "beads/refused", "--json")
}
