package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/configfile"
)

func TestGraphPreviewPropertiesAdmission(t *testing.T) {
	const valid = `{"note":"context — 雪"}`
	file := filepath.Join(t.TempDir(), "properties.json")
	if err := os.WriteFile(file, []byte(valid), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, input := range []string{valid, "@" + file, "@-"} {
		properties, err := graphPreviewProperties(input, strings.NewReader(valid))
		if err != nil || !reflect.DeepEqual(properties, map[string]any{"note": "context — 雪"}) {
			t.Fatalf("%q: properties=%v error=%v", input, properties, err)
		}
	}
	for name, input := range map[string]string{
		"null": "null", "array": "[]", "scalar": `"note"`, "trailing": `{} {}`,
		"duplicate":         `{"note":"one","note":"two"}`,
		"escaped duplicate": `{"note":"one","\u006eote":"two"}`,
		"lone surrogate":    `{"note":"\ud800"}`,
		"invalid UTF8":      "{\"note\":\"\xff\"}",
		"oversized":         strings.Repeat(" ", graphPreviewPropertiesLimit) + "{}",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := graphPreviewProperties(input, strings.NewReader("")); err == nil {
				t.Fatal("invalid input admitted")
			}
		})
	}
}

func TestGraphPreviewRevisionGuardIsExplicit(t *testing.T) {
	for _, tc := range []struct {
		name, wantRevision   string
		args                 []string
		wantCurrent, wantErr bool
	}{
		{"absent accepts current", "", nil, true, false},
		{"observed", "observed", []string{"--if-revision==observed"}, false, false},
		{"empty", "", []string{"--if-revision="}, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().String("if-revision", "", "")
			if err := cmd.ParseFlags(tc.args); err != nil {
				t.Fatal(err)
			}
			revision, current, err := graphPreviewRevisionGuard(cmd, false, "plan", false)
			if revision != tc.wantRevision || current != tc.wantCurrent || (err != nil) != tc.wantErr {
				t.Fatalf("revision=%q current=%t err=%v", revision, current, err)
			}
		})
	}
}

func TestGraphPreviewEditRevisionGuardDefaultsToCurrent(t *testing.T) {
	for _, command := range []*cobra.Command{rememberCmd, updateCmd} {
		if command.Flags().Lookup("unconditional") != nil {
			t.Fatalf("%s still advertises the removed edit flag", command.CommandPath())
		}
	}
	for _, tc := range []struct {
		args         []string
		wantRevision string
		wantCurrent  bool
		wantError    bool
	}{
		{nil, "", true, false},
		{[]string{"--if-revision==observed"}, "observed", false, false},
		{[]string{"--if-revision="}, "", false, true},
	} {
		cmd := &cobra.Command{}
		cmd.Flags().String("if-revision", "", "")
		if err := cmd.ParseFlags(tc.args); err != nil {
			t.Fatal(err)
		}
		revision, current, err := graphPreviewEditRevisionGuard(cmd, "plan", false)
		if revision != tc.wantRevision || current != tc.wantCurrent || (err != nil) != tc.wantError {
			t.Fatalf("args=%v revision=%q current=%t err=%v", tc.args, revision, current, err)
		}
	}
}

func TestGraphPreviewUniformRevisionFlagSyntax(t *testing.T) {
	for _, raw := range []string{"1", "3", "9223372036854775807", "=3", "=saved-r1"} {
		if _, err := graphPreviewParseRevisionFlag("revision", raw); err != nil {
			t.Errorf("valid --revision %q: %v", raw, err)
		}
	}
	for _, raw := range []string{"", "0", "-1", "current", "saved-r1", "=", "9223372036854775808"} {
		if _, err := graphPreviewParseRevisionFlag("revision", raw); err == nil {
			t.Errorf("invalid --revision %q accepted", raw)
		}
	}
	for _, tc := range []struct{ raw, token string }{{"=3", "3"}, {"=saved-r1", "saved-r1"}} {
		got, err := graphPreviewRevisionValue("revision", tc.raw, "", false)
		if err != nil || got != tc.token {
			t.Errorf("exact %q = %q, %v; want %q", tc.raw, got, err, tc.token)
		}
	}
}

func TestGraphPreviewLocalTypeSelectors(t *testing.T) {
	const scope = "https://example.invalid/demo/"
	for _, selector := range []string{"references", "references", scope + "references", "subject-to", "subject-to", "beads/foo", "https://other.invalid/types/example-follows"} {
		want := selector
		if !strings.Contains(selector, "://") {
			want = scope + selector
		}
		got, err := graphPreviewTypeURL(scope, selector)
		if err != nil || got != want {
			t.Errorf("%q = %q %v want %q", selector, got, err, want)
		}
	}
	for _, selector := range []string{"", "types/", "types/../x", "types/example%2Dfollows", "types/example?x=1", "types//x", scope + "types/example-follows#x", "bad?type", "bad%2Ftype"} {
		if _, err := graphPreviewTypeURL(scope, selector); err == nil {
			t.Errorf("accepted malformed canonical ID %q", selector)
		}
	}
}

func TestGraphPreviewTypeSelectionPreservesCompleteLocalIDs(t *testing.T) {
	const scope = "https://example.invalid/demo/"
	for _, selector := range []string{"memory", scope + "memory", "references", scope + "references", "types/literal", scope + "types/literal"} {
		want := selector
		if !strings.Contains(selector, "://") {
			want = scope + selector
		}
		got, err := graphPreviewTypeURL(scope, selector)
		if err != nil || got != want {
			t.Errorf("selector %q = %q, %v; want complete identity %q", selector, got, err, want)
		}
	}
}

func TestGraphPreviewLinkDisplayKeepsForeignURLs(t *testing.T) {
	old := graphPreviewConfig
	graphPreviewConfig = &configfile.Config{GraphScopeURL: "https://example.invalid/demo/"}
	t.Cleanup(func() { graphPreviewConfig = old })
	for input, want := range map[string]string{
		"https://example.invalid/demo/references": "references",
		"https://example.invalid/demo/links/edge": "links/edge",
		"https://example.invalid/demo/beads/work": "beads/work",
		"https://remote.invalid/beads/work":       "https://remote.invalid/beads/work",
	} {
		if got := graphPreviewDisplayLocalURL(input); got != want {
			t.Errorf("%q displayed as %q, want %q", input, got, want)
		}
	}
}

func TestGraphPreviewLinkTypeFlags(t *testing.T) {
	old := graphPreviewConfig
	graphPreviewConfig = &configfile.Config{GraphScopeURL: "https://example.invalid/demo/"}
	defer func() { graphPreviewConfig = old }()
	for _, tc := range []struct {
		args []string
		want string
		fail bool
	}{
		{nil, "", false},
		{[]string{"--type", "references"}, "https://example.invalid/demo/references", false},
		{[]string{"--type", "types/preview-related"}, "https://example.invalid/demo/types/preview-related", false},
		{[]string{"--type", "blocks"}, "https://example.invalid/demo/blocks", false},
		{[]string{"--type", "parent-child"}, "https://example.invalid/demo/parent-child", false},
		{[]string{"--type", "random"}, "", true},
		{[]string{"--type="}, "", true},
	} {
		cmd := &cobra.Command{}
		registerGraphLinkTypeFlag(cmd)
		if err := cmd.ParseFlags(tc.args); err != nil {
			t.Fatal(err)
		}
		got, err := graphPreviewLinkType(cmd)
		if (err != nil) != tc.fail || got != tc.want {
			t.Fatalf("%v: got=%q err=%v", tc.args, got, err)
		}
		if cmd.Flags().Lookup("type") == nil || cmd.Flags().Lookup("link-type") != nil || cmd.Flags().Lookup("resource-type") != nil {
			t.Fatal("only the unified type flag should be registered")
		}
	}
}
