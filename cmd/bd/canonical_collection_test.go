package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/clioutput"
	"github.com/steveyegge/beads/internal/configfile"
	"github.com/steveyegge/beads/internal/httpapi/bdpwire"
	"github.com/steveyegge/beads/internal/httpapi/graphread"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

const canonicalCollectionScope = "https://example.test/collection/"

func canonicalCollectionTestCommand(t *testing.T, links bool, args ...string) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	isolateBeadsDirForTest(t)
	old := graphPreviewConfig
	graphPreviewConfig = &configfile.Config{GraphScopeURL: canonicalCollectionScope}
	t.Cleanup(func() { graphPreviewConfig = old })
	cmd := &cobra.Command{Use: "list"}
	registerCanonicalCollectionFlags(cmd)
	cmd.Flags().Int("limit", 50, "")
	cmd.Flags().String("type", "", "")
	cmd.Flags().Bool("json", false, "")
	cmd.Flags().String("format", "", "")
	cmd.Flags().Bool("all", false, "")
	if links {
		cmd.Flags().Lookup("limit").Annotations = map[string][]string{"canonical-output-only": {"true"}}
		for _, flag := range []string{"source", "target", "incident-to"} {
			cmd.Flags().String(flag, "", "")
		}
		cmd.Flags().String("direction", "both", "")
	}
	out, diagnostics := &bytes.Buffer{}, &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(diagnostics)
	if err := cmd.ParseFlags(args); err != nil {
		t.Fatal(err)
	}
	return cmd, out, diagnostics
}

func canonicalCollectionTestPlan(t *testing.T, cmd *cobra.Command, links bool) (canonicalCollectionPlan, canonicalCollectionFailure) {
	t.Helper()
	options, _, err := canonicalOutputOptions(cmd)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := canonicalCollectionRequest(cmd, links, options)
	if err != nil {
		t.Fatal(err)
	}
	descriptor, err := canonicalCollectionDescriptor(links, nil)
	if err != nil {
		t.Fatal(err)
	}
	renderer, err := clioutput.NewCollection(descriptor, options)
	if err != nil {
		t.Fatal(err)
	}
	return plan, canonicalCollectionFailure{renderer: renderer, jsonl: plan.jsonl}
}

func canonicalCollectionRecord(id string) json.RawMessage {
	return json.RawMessage(fmt.Sprintf(`{"id":%q,"type":%q,"revision":"r1","properties":{"body":"large","title":"title"},"metadata":{}}`, canonicalCollectionScope+"beads/"+id, canonicalCollectionScope+"types/note"))
}

func TestCanonicalCollectionExistingBDPSelection(t *testing.T) {
	scope := canonicalCollectionScope
	note, related := scope+"types/note", scope+"types/related"
	inventory := graphread.Inventory{
		Beads: []bdpwire.BeadRecord{{ID: scope + "beads/a", Type: note, Revision: "r", Properties: bdpwire.Properties{}, Metadata: bdpwire.Metadata{}}, {ID: scope + "beads/b", Type: note, Revision: "r", Properties: bdpwire.Properties{}, Metadata: bdpwire.Metadata{}}},
		Types: []bdpwire.TypeDescriptor{{ID: note, Name: "Note", Describes: bdpwire.DescribesBead}, {ID: related, Name: "Related", Describes: bdpwire.DescribesLink}},
		Links: []bdpwire.LinkRecord{{ID: scope + "links/l", Type: related, Revision: "r", Source: bdpwire.Reference{URI: scope + "beads/a"}, Target: bdpwire.Reference{URI: scope + "beads/b"}, Properties: bdpwire.Properties{}, Metadata: bdpwire.Metadata{}}},
	}
	for _, tc := range []struct {
		name  string
		links bool
		args  []string
		want  int
	}{
		{"beads", false, []string{"--type", note}, 2},
		{"type-family", false, []string{"--conforms-to", note}, 2},
		{"unknown-type", false, []string{"--type", scope + "types/missing"}, 0},
		{"source-and-target", true, []string{"--source", "a", "--target", "beads/b"}, 1},
		{"incident-in", true, []string{"--incident-to", "b", "--direction", "in"}, 1},
		{"incident-out-empty", true, []string{"--incident-to", "b", "--direction", "out"}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd, _, _ := canonicalCollectionTestCommand(t, tc.links, append([]string{"--output-format", "json", "--unpaged"}, tc.args...)...)
			plan, _ := canonicalCollectionTestPlan(t, cmd, tc.links)
			records, err := plan.selection.Select(inventory)
			if err != nil || len(records) != tc.want {
				t.Fatalf("selection=%s err=%v want=%d", records, err, tc.want)
			}
		})
	}
}

func TestCanonicalCollectionCompleteNativePayload(t *testing.T) {
	cmd, out, _ := canonicalCollectionTestCommand(t, false, "--output-format", "json", "--unpaged", "--omit", "/properties/body")
	plan, failure := canonicalCollectionTestPlan(t, cmd, false)
	if err := renderCanonicalCollection(cmd, plan, []json.RawMessage{canonicalCollectionRecord("a")}, failure); err != nil {
		t.Fatal(err)
	}
	var envelope struct {
		SchemaVersion int                  `json:"schema_version"`
		BDPSpec       string               `json:"bdp_spec"`
		Data          json.RawMessage      `json:"data"`
		Completion    clioutput.Completion `json:"completion"`
	}
	if err := json.Unmarshal(out.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.SchemaVersion != 2 || envelope.BDPSpec != bdpwire.Pin || envelope.Completion.Returned != 1 || envelope.Completion.Exhausted == nil || !*envelope.Completion.Exhausted {
		t.Fatalf("wrong envelope: %s", out.Bytes())
	}
	var page bdpwire.BeadCollection
	if err := bdpwire.Unmarshal(envelope.Data, &page); err != nil {
		t.Fatal(err)
	}
	if page.Next != nil || len(page.Items) != 1 || page.Items[0].Properties["body"] != nil {
		t.Fatalf("bad native page: %s", envelope.Data)
	}
}

func TestCanonicalCollectionRefusesFalseTerminalPage(t *testing.T) {
	records := []json.RawMessage{canonicalCollectionRecord("a"), canonicalCollectionRecord("b")}
	for _, args := range [][]string{
		{"--output-format", "json", "--page-size", "1"},
		{"--output-format", "json", "--unpaged", "--limit", "1"},
		{"--output-format", "jsonl", "--limit", "1"},
	} {
		cmd, out, diagnostics := canonicalCollectionTestCommand(t, false, args...)
		plan, failure := canonicalCollectionTestPlan(t, cmd, false)
		err := renderCanonicalCollection(cmd, plan, records, failure)
		var exit *exitError
		if !errors.As(err, &exit) || exit.Code != 5 {
			t.Fatalf("truncated selection accepted: %v", err)
		}
		if bytes.Contains(out.Bytes(), []byte(`"outcome":"success"`)) || bytes.Contains(out.Bytes(), []byte(`"record_type":"item"`)) || bytes.Contains(out.Bytes(), []byte(`"next":null`)) {
			t.Fatalf("false terminal page: %s", out.Bytes())
		}
		if !strings.Contains(diagnostics.String(), "--unpaged") {
			t.Fatalf("missing guidance: %s", diagnostics.String())
		}
	}
}

func TestCanonicalCollectionJSONLDefaultCompleteAndRaw(t *testing.T) {
	for _, raw := range []bool{false, true} {
		args := []string{"--output-format", "jsonl"}
		if raw {
			args = append(args, "--omit", "envelope")
		}
		cmd, out, _ := canonicalCollectionTestCommand(t, false, args...)
		plan, failure := canonicalCollectionTestPlan(t, cmd, false)
		records := make([]json.RawMessage, 51)
		for i := range records {
			records[i] = canonicalCollectionRecord(fmt.Sprintf("n%02d", i))
		}
		if err := renderCanonicalCollection(cmd, plan, records, failure); err != nil {
			t.Fatal(err)
		}
		lines := bytes.Split(bytes.TrimSpace(out.Bytes()), []byte("\n"))
		want := 53
		if raw {
			want = 51
		}
		if len(lines) != want {
			t.Fatalf("got %d records want %d", len(lines), want)
		}
		if raw && bytes.Contains(out.Bytes(), []byte(`"record_type"`)) {
			t.Fatal("raw stream contains framing")
		}
	}
}

func TestCanonicalCollectionAdmissionBeforeRead(t *testing.T) {
	for _, args := range [][]string{
		{"--output-format", "json", "--page-size", "0"},
		{"--output-format", "json", "--limit", "-1"},
		{"--output-format", "json", "--paged", "--unpaged"},
		{"--output-format", "json", "--type", "task"},
		{"--output-format", "json", "--continue", "opaque"},
	} {
		cmd, _, _ := canonicalCollectionTestCommand(t, false, args...)
		options, _, err := canonicalOutputOptions(cmd)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := canonicalCollectionRequest(cmd, false, options); err == nil {
			t.Fatalf("invalid args accepted: %v", args)
		}
	}
	cmd, _, _ := canonicalCollectionTestCommand(t, true, "--output-format", "json", "--direction", "out")
	options, _, _ := canonicalOutputOptions(cmd)
	if _, err := canonicalCollectionRequest(cmd, true, options); err == nil {
		t.Fatal("direction without incident accepted")
	}
	cmd, _, _ = canonicalCollectionTestCommand(t, false, "--output-format", "json", "--continue", "opaque")
	options, _, _ = canonicalOutputOptions(cmd)
	if _, err := canonicalCollectionRequest(cmd, false, options); !errors.Is(err, graphstore.ErrCapabilityUnavailable) {
		t.Fatal("continuation not classified unsupported")
	}
}

func TestCanonicalCollectionScalarConflicts(t *testing.T) {
	for _, tc := range []struct {
		args []string
		bad  bool
	}{
		{[]string{"--limit", "1", "--limit", "2"}, true},
		{[]string{"--limit", "1", "--limit", "01"}, false},
		{[]string{"--paged=true", "--paged=false"}, true},
		{[]string{"--type", "a", "--type", "b"}, true},
		{[]string{"--output-format", "json", "--omit", "/metadata", "--omit", "/properties"}, false},
	} {
		cmd, _, _ := canonicalCollectionTestCommand(t, false, tc.args...)
		if err := canonicalCollectionScalarOccurrences(cmd, tc.args); (err != nil) != tc.bad {
			t.Fatalf("args=%v err=%v bad=%t", tc.args, err, tc.bad)
		}
	}
}

func TestCanonicalCollectionPublicGoldenCorpus(t *testing.T) {
	for _, tc := range []struct {
		name   string
		links  bool
		format string
		omit   []string
		golden string
	}{
		{"beads", false, "json", []string{"/properties/body"}, "beads.projected.json"},
		{"links", true, "jsonl", nil, "links.jsonl"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			args := []string{"--output-format", tc.format, "--unpaged"}
			for _, path := range tc.omit {
				args = append(args, "--omit", path)
			}
			cmd, out, _ := canonicalCollectionTestCommand(t, tc.links, args...)
			plan, failure := canonicalCollectionTestPlan(t, cmd, tc.links)
			input, err := os.ReadFile(filepath.Join("testdata", "canonical-collections", tc.name+".input.json"))
			if err != nil {
				t.Fatal(err)
			}
			var page graphread.Page
			if err := json.Unmarshal(input, &page); err != nil {
				t.Fatal(err)
			}
			if err := renderCanonicalCollection(cmd, plan, page.Items, failure); err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile(filepath.Join("testdata", "canonical-collections", tc.golden))
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(out.Bytes(), want) {
				t.Fatalf("public corpus changed\ngot %s\nwant %s", out.Bytes(), want)
			}
		})
	}
}

func TestCanonicalCollectionFlagsRequireOptIn(t *testing.T) {
	for _, flag := range []string{"--conforms-to=types/example", "--paged", "--unpaged", "--page-size=20", "--continue=token", "--limit=10"} {
		t.Run(flag, func(t *testing.T) {
			cmd, _, _ := canonicalCollectionTestCommand(t, true, flag)
			if _, selected, err := canonicalOutputOptions(cmd); err == nil || selected || !strings.Contains(err.Error(), "requires --output-format") {
				t.Fatalf("selected=%v err=%v", selected, err)
			}
		})
	}
	cmd, _, _ := canonicalCollectionTestCommand(t, false, "--limit=10")
	if _, selected, err := canonicalOutputOptions(cmd); err != nil || selected {
		t.Fatalf("shipping list limit changed: selected=%v err=%v", selected, err)
	}
}
