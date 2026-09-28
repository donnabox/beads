package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"
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
		name     string
		args     []string
		required bool
		wantErr  bool
	}{
		{"optional absent", nil, false, false},
		{"required absent", nil, true, true},
		{"observed", []string{"--if-revision=observed"}, true, false},
		{"empty", []string{"--if-revision="}, true, true},
		{"unconditional", []string{"--unconditional"}, true, false},
		{"false is not permission", []string{"--unconditional=false"}, false, true},
		{"both", []string{"--if-revision=observed", "--unconditional"}, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().String("if-revision", "", "")
			cmd.Flags().Bool("unconditional", false, "")
			if err := cmd.ParseFlags(tc.args); err != nil {
				t.Fatal(err)
			}
			_, _, err := graphPreviewRevisionGuard(cmd, false, tc.required)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err=%v, wantErr=%v", err, tc.wantErr)
			}
		})
	}
}
