package main

import (
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/configfile"
	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func selectedRememberCommand(t *testing.T, flags []string) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{}
	for _, name := range []string{"id", "title", "update", "if-revision", "body-file", "key"} {
		cmd.Flags().String(name, "", "")
	}
	cmd.Flags().Bool("stdin", false, "")
	if err := cmd.ParseFlags(flags); err != nil {
		t.Fatal(err)
	}
	return cmd
}

func TestGraphPreviewRememberSelectedProperties(t *testing.T) {
	original := graphstore.Record{ID: "https://example.invalid/beads/plan", Revision: "observed",
		Properties: graphstore.Properties{Title: "  Preserve 雪\n", Body: "old"},
		Owned:      []json.RawMessage{json.RawMessage(`{"id":"https://example.invalid/links/context","version":"link-version"}`)}}
	before := original
	before.Owned = []json.RawMessage{append(json.RawMessage(nil), original.Owned[0]...)}
	for _, tc := range []struct {
		name, title, body string
		explicitTitle     bool
	}{
		{"preserved-title", original.Properties.Title, "---\r\n雪\r\n---\n  new body\t", false},
		{"clear-body", original.Properties.Title, "", false},
		{"explicit-title", "New title", "new", true},
		{"explicit-empty-title", "", "new", true},
		{"both-empty", "", "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var title *string
			if tc.explicitTitle {
				title = &tc.title
			}
			got, err := graphPreviewRememberUpdateProperties(original, tc.body, title, "observed")
			if err != nil || got != (graphstore.Properties{Title: tc.title, Body: tc.body}) {
				t.Fatalf("properties=%+v err=%v", got, err)
			}
			if !reflect.DeepEqual(original, before) {
				t.Fatal("composition changed the observed record")
			}
		})
	}
	for _, tc := range []struct {
		name     string
		current  any
		revision string
		want     error
	}{
		{"stale-even-same-body", original, "stale", graphstore.ErrConflict},
		{"missing-guard", original, "", storage.ErrValidation},
		{"issue", graphstore.IssueRecord{}, "observed", graphstore.ErrCapabilityUnavailable},
		{"link", graphstore.LinkRecord{}, "observed", graphstore.ErrCapabilityUnavailable},
		{"missing-record", nil, "observed", graphstore.ErrCapabilityUnavailable},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := graphPreviewRememberUpdateProperties(tc.current, original.Properties.Body, nil, tc.revision)
			if !errors.Is(err, tc.want) || got != (graphstore.Properties{}) {
				t.Fatalf("properties=%+v err=%v", got, err)
			}
		})
	}
}

// Changed graph flags must reach workspace admission even with no positional
// body. Bare legacy calls retain their original one-argument requirement.
func TestGraphPreviewRememberSelectedArgumentAdmission(t *testing.T) {
	for _, tc := range []struct {
		name  string
		flags []string
		graph bool
	}{
		{"selected", []string{"--update=beads/plan"}, true},
		{"empty-selected", []string{"--update="}, true},
		{"guard-only", []string{"--if-revision=observed"}, true},
		{"empty-guard", []string{"--if-revision="}, true},
		{"existing-file", []string{"--body-file=body.md"}, true},
		{"legacy-key", []string{"--key=plan"}, false},
		{"bare-legacy", nil, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := selectedRememberCommand(t, tc.flags)
			if rememberGraphFlagsChanged(cmd) != tc.graph {
				t.Fatal("wrong graph admission selection")
			}
			if err := rememberArgs(cmd, nil); (err == nil) != tc.graph {
				t.Fatalf("wrong missing-body admission: %v", err)
			}
			if err := rememberArgs(cmd, []string{"content"}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

type selectedRememberInputProbe struct{ reads int }

func (p *selectedRememberInputProbe) Read([]byte) (int, error) {
	p.reads++
	return 0, io.ErrUnexpectedEOF
}

func TestGraphPreviewRememberSelectedRefusesBeforeInput(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("GT_TOWN_ROOT", "")
	t.Setenv("GT_ROOT", "")
	oldConfig, oldReadonly := graphPreviewConfig, readonlyMode
	oldJSON, oldStructured := jsonOutput, graphPreviewStructuredErrors
	t.Cleanup(func() {
		graphPreviewConfig, readonlyMode = oldConfig, oldReadonly
		jsonOutput, graphPreviewStructuredErrors = oldJSON, oldStructured
	})
	graphPreviewConfig = &configfile.Config{GraphScopeURL: "https://example.invalid/"}
	jsonOutput, graphPreviewStructuredErrors = false, false
	for _, tc := range []struct {
		name     string
		flags    []string
		readonly bool
		code     int
	}{
		{"missing-guard", []string{"--update=beads/plan"}, false, 2},
		{"empty-guard", []string{"--update=beads/plan", "--if-revision="}, false, 2},
		{"invalid-guard", []string{"--update=beads/plan", "--if-revision=\xff"}, false, 2},
		{"oversized-guard", []string{"--update=beads/plan", "--if-revision=" + strings.Repeat("x", graphstore.PreviewVersionTokenLimit+1)}, false, 2},
		{"empty-selector", []string{"--update=", "--if-revision=observed"}, false, 2},
		{"foreign-selector", []string{"--update=https://foreign.invalid/beads/plan", "--if-revision=observed"}, false, 2},
		{"link-selector", []string{"--update=links/context", "--if-revision=observed"}, false, 2},
		{"create-and-update", []string{"--id=beads/new", "--update=beads/plan", "--if-revision=observed"}, false, 2},
		{"create-with-guard", []string{"--id=beads/new", "--title=New", "--if-revision=observed"}, false, 5},
		{"guard-without-selection", []string{"--if-revision=observed"}, false, 5},
		{"invalid-title", []string{"--update=beads/plan", "--if-revision=observed", "--title=\xff"}, false, 2},
		{"legacy-key", []string{"--update=beads/plan", "--if-revision=observed", "--key=legacy"}, false, 5},
		{"readonly", []string{"--update=beads/plan", "--if-revision=observed"}, true, 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			readonlyMode = tc.readonly
			cmd := selectedRememberCommand(t, append(append([]string(nil), tc.flags...), "--stdin"))
			input := &selectedRememberInputProbe{}
			cmd.SetIn(input)
			err := runGraphPreviewRemember(cmd, nil)
			var failure *exitError
			if !errors.As(err, &failure) || failure.Code != tc.code || input.reads != 0 {
				t.Fatalf("refusal=%v reads=%d", err, input.reads)
			}
		})
	}
}
