package main

import (
	"testing"

	"github.com/spf13/cobra"
)

// TestValidateCommentsAddArgs is a pure unit test for "comments add"'s Args
// guard — it calls the validator directly with no store, no Dolt, and no
// cobra dispatch, so it runs regardless of cgo/Docker availability. Mirrors
// TestValidateCommentArgs (the singular "comment" sibling in comment_test.go)
// for the plural form: PR #5393 review item M1 established that "comments
// add" must reject a reserved-word id exactly like "comment" does, since the
// two are documented as guarded twins (CHANGELOG.md, hash-ids.md).
func TestValidateCommentsAddArgs(t *testing.T) {
	origJSONOutput := jsonOutput
	jsonOutput = false
	t.Cleanup(func() { jsonOutput = origJSONOutput })

	cases := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{name: "bare list is rejected", args: []string{"list", "some text"}, wantErr: true},
		{name: "bare add is rejected", args: []string{"add", "some text"}, wantErr: true},
		{name: "bare rm is rejected", args: []string{"rm", "some text"}, wantErr: true},
		{name: "bare delete is rejected", args: []string{"delete", "some text"}, wantErr: true},
		{name: "real id with text starting with the word list is fine", args: []string{"test-abc123", "list", "of", "things", "to", "do"}, wantErr: false},
		{name: "real id with text starting with the word add is fine", args: []string{"test-abc123", "add", "one", "more", "item"}, wantErr: false},
		{name: "real id alone (text comes from --stdin/--file) is fine", args: []string{"test-abc123"}, wantErr: false},
		{name: "no args at all is rejected by the base MinimumNArgs check", args: []string{}, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateCommentsAddArgs(commentsAddCmd, tc.args)
			if tc.wantErr && err == nil {
				t.Fatalf("validateCommentsAddArgs(%q): expected an error, got nil", tc.args)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("validateCommentsAddArgs(%q): expected no error, got %v", tc.args, err)
			}
		})
	}
}

// TestCheckCommentAuthorAlias pins --author as an alias of --actor: two different
// explicit values are refused, and nothing else is. The flags are declared on a
// tree of its own, so nothing here depends on the real root's parsed state.
func TestCheckCommentAuthorAlias(t *testing.T) {
	origJSONOutput := jsonOutput
	jsonOutput = false
	t.Cleanup(func() { jsonOutput = origJSONOutput })

	cases := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{name: "two different values", args: []string{"--actor", "sam", "add", "--author", "kim"}, wantErr: true},
		{name: "two different values, the short flag", args: []string{"--actor", "sam", "add", "-a", "kim"}, wantErr: true},
		{name: "--actor after the command", args: []string{"add", "--actor", "sam", "--author", "kim"}, wantErr: true},
		{name: "the same value twice", args: []string{"--actor", "sam", "add", "--author", "sam"}},
		{name: "only --actor", args: []string{"--actor", "sam", "add"}},
		{name: "only --author", args: []string{"add", "--author", "kim"}},
		{name: "a blank --author names nobody", args: []string{"--actor", "sam", "add", "--author", ""}},
		{name: "a blank --actor names nobody", args: []string{"--actor", "", "add", "--author", "kim"}},
		{name: "neither", args: []string{"add"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := &cobra.Command{Use: "bd"}
			root.PersistentFlags().String("actor", "", "")
			var got error
			add := &cobra.Command{Use: "add", Run: func(cmd *cobra.Command, _ []string) { got = checkCommentAuthorAlias(cmd) }}
			add.Flags().StringP("author", "a", "", "")
			root.AddCommand(add)
			root.SetArgs(tc.args)
			if err := root.Execute(); err != nil {
				t.Fatalf("parsing %q: %v", tc.args, err)
			}
			if (got != nil) != tc.wantErr {
				t.Fatalf("%q: error = %v, want error %v", tc.args, got, tc.wantErr)
			}
		})
	}
}
