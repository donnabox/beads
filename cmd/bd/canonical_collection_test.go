package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
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
		for _, flag := range []string{"from", "to", "incident-to"} {
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
		{"source-and-target", true, []string{"--from", "beads/a", "--to", "beads/b"}, 1},
		{"no-kind-prefix-insertion", true, []string{"--from", "a", "--to", "beads/b"}, 0},
		{"incident-in", true, []string{"--incident-to", "beads/b", "--direction", "in"}, 1},
		{"incident-out-empty", true, []string{"--incident-to", "beads/b", "--direction", "out"}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd, _, _ := canonicalCollectionTestCommand(t, tc.links, append([]string{"--output-format", "json", "--unpaged"}, tc.args...)...)
			var plan canonicalCollectionPlan
			if tc.links {
				// The BDP selector is pure. Product Link flags now resolve opaque
				// public IDs through a real store, which the installed workflow
				// below covers; keep this inventory test focused on BDP filtering.
				parameters := url.Values{}
				for _, binding := range []struct{ flag, parameter string }{{"from", "source"}, {"to", "target"}, {"incident-to", "endpoint"}} {
					if cmd.Flags().Changed(binding.flag) {
						value, _ := cmd.Flags().GetString(binding.flag)
						parameters.Set(binding.parameter, value)
					}
				}
				direction, _ := cmd.Flags().GetString("direction")
				if direction != "both" {
					key := "source"
					if direction == "in" {
						key = "target"
					}
					parameters.Set(key, parameters.Get("endpoint"))
					parameters.Del("endpoint")
				}
				var err error
				plan.selection, err = graphread.CompileCollection(scope, "links", parameters, graphread.SelectorLimits{})
				if err != nil {
					t.Fatal(err)
				}
			} else {
				plan, _ = canonicalCollectionTestPlan(t, cmd, false)
			}
			records, err := plan.selection.Select(inventory)
			if err != nil || len(records) != tc.want {
				t.Fatalf("selection=%s err=%v want=%d", records, err, tc.want)
			}
		})
	}
}

func TestCanonicalLinkCollectionUsesCapturedEndpointRevisions(t *testing.T) {
	cmd, _, _ := canonicalCollectionTestCommand(t, true, "--output-format", "json", "--unpaged")
	_ = cmd
	scope := canonicalCollectionScope
	bead, linkType := scope+"types/note", scope+"types/related"
	inventory := graphread.Inventory{
		Beads: []bdpwire.BeadRecord{
			{ID: scope + "a", Type: bead, Revision: "r2", Properties: bdpwire.Properties{}, Metadata: bdpwire.Metadata{}},
			{ID: scope + "b", Type: bead, Revision: "r1", Properties: bdpwire.Properties{}, Metadata: bdpwire.Metadata{}},
		},
		Types: []bdpwire.TypeDescriptor{{ID: bead, Name: "Note", Describes: bdpwire.DescribesBead}, {ID: linkType, Name: "Related", Describes: bdpwire.DescribesLink}},
	}
	for _, link := range []struct{ id, source, target, sourcePin string }{
		{"floating", "a", "b", ""}, {"old", "a", "b", "r1"}, {"current", "a", "b", "r2"},
		{"reverse", "b", "a", ""}, {"self", "a", "a", ""},
	} {
		inventory.Links = append(inventory.Links, bdpwire.LinkRecord{ID: scope + "links/" + link.id, Type: linkType, Revision: "l1",
			Source: bdpwire.Reference{URI: scope + link.source, Revision: link.sourcePin},
			Target: bdpwire.Reference{URI: scope + link.target}, Properties: bdpwire.Properties{}, Metadata: bdpwire.Metadata{}})
	}
	selection, err := graphread.CompileCollection(scope, "links", url.Values{}, graphread.SelectorLimits{})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		req  graphstore.LinkCollectionRequest
		want int
	}{
		{"bare incidence", graphstore.LinkCollectionRequest{IncidentPath: "beads/a", IncidentURL: scope + "a"}, 5},
		{"historical pin", graphstore.LinkCollectionRequest{IncidentPath: "beads/a", IncidentURL: scope + "a", IncidentPin: graphstore.LinkPinSelection{Revision: "r1", Explicit: true}}, 1},
		{"current pin", graphstore.LinkCollectionRequest{IncidentPath: "beads/a", IncidentURL: scope + "a", IncidentPin: graphstore.LinkPinSelection{Current: true, Explicit: true}}, 4},
		{"all revisions override", graphstore.LinkCollectionRequest{IncidentPath: "beads/a", IncidentURL: scope + "a", IncidentPin: graphstore.LinkPinSelection{Revision: "r1", Explicit: true}, AllRevisions: true}, 5},
		{"out pair", graphstore.LinkCollectionRequest{SourcePath: "beads/a", SourceURL: scope + "a", TargetPath: "beads/b", TargetURL: scope + "b", PairDirection: "out"}, 3},
		{"in pair", graphstore.LinkCollectionRequest{SourcePath: "beads/a", SourceURL: scope + "a", TargetPath: "beads/b", TargetURL: scope + "b", PairDirection: "in"}, 1},
		{"both pair", graphstore.LinkCollectionRequest{SourcePath: "beads/a", SourceURL: scope + "a", TargetPath: "beads/b", TargetURL: scope + "b", PairDirection: "both"}, 4},
		{"self once", graphstore.LinkCollectionRequest{SourcePath: "beads/a", SourceURL: scope + "a", TargetPath: "beads/a", TargetURL: scope + "a", PairDirection: "both"}, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := selectCanonicalCollection(canonicalCollectionPlan{selection: selection, links: true, linkRequest: tc.req}, inventory)
			if err != nil || len(got) != tc.want {
				t.Fatalf("got %d want %d: %v", len(got), tc.want, err)
			}
		})
	}
	missing := inventory
	missing.Beads = inventory.Beads[1:]
	if _, err := selectCanonicalCollection(canonicalCollectionPlan{selection: selection, links: true,
		linkRequest: graphstore.LinkCollectionRequest{IncidentPath: "beads/a", IncidentURL: scope + "a"}}, missing); !errors.Is(err, graphstore.ErrNotFound) {
		t.Fatalf("noncurrent selected Bead returned an empty collection: %v", err)
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
		{"--output-format", "json", "--type", "task", "--conforms-to", canonicalCollectionScope + "types/task"},
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
		{[]string{"--limit=010", "--limit=10"}, true},
		{[]string{"--limit=0x10", "--limit=16"}, false},
		{[]string{"--page-size=010", "--page-size=10"}, true},
		{[]string{"--page-size=0x10", "--page-size=16"}, false},
		{[]string{"--paged=true", "--paged=false"}, true},
		{[]string{"--type", "a", "--type", "b"}, true},
		{[]string{"--output-format", "json", "--omit", "/metadata", "--omit", "/properties"}, false},
	} {
		cmd, _, _ := canonicalCollectionTestCommand(t, false, tc.args...)
		if err := canonicalScalarOccurrences(cmd, tc.args, "type", "limit", "page-size", "paged"); (err != nil) != tc.bad {
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

func TestCanonicalCollectionPreservesLegacyArgs(t *testing.T) {
	legacyError := errors.New("legacy positional diagnostic")
	for _, canonical := range []bool{false, true} {
		cmd := &cobra.Command{Use: "list", Args: func(_ *cobra.Command, _ []string) error { return legacyError }}
		registerCanonicalCollectionFlags(cmd)
		args := []string{"unexpected"}
		if canonical {
			args = append(args, "--output-format=json")
		}
		if err := cmd.ParseFlags(args); err != nil {
			t.Fatal(err)
		}
		err := cmd.ValidateArgs(cmd.Flags().Args())
		if canonical && err != nil || !canonical && !errors.Is(err, legacyError) {
			t.Fatalf("canonical=%v err=%v", canonical, err)
		}
	}
}

func TestCanonicalCollectionDefaultPageBoundary(t *testing.T) {
	records := make([]json.RawMessage, 51)
	for i := range records {
		records[i] = canonicalCollectionRecord(fmt.Sprintf("b%02d", i))
	}
	for _, format := range []string{"default", "table", "detail", "json"} {
		t.Run(format, func(t *testing.T) {
			cmd, out, diagnostics := canonicalCollectionTestCommand(t, false, "--output-format", format)
			plan, failure := canonicalCollectionTestPlan(t, cmd, false)
			var exit *exitError
			if err := renderCanonicalCollection(cmd, plan, records, failure); !errors.As(err, &exit) || exit.Code != 5 {
				t.Fatalf("default bound not refused: %v", err)
			}
			if !strings.Contains(diagnostics.String(), "requested page boundary of 50") || !strings.Contains(diagnostics.String(), "no items were emitted") {
				t.Fatalf("unclear bound diagnostic: %s", diagnostics)
			}
			if bytes.Contains(out.Bytes(), []byte(`"next":null`)) || bytes.Contains(out.Bytes(), []byte(`"outcome":"success"`)) {
				t.Fatalf("false complete collection: %s", out)
			}
		})
	}
}

func TestCanonicalCollectionRejectsIssueAllSelection(t *testing.T) {
	for _, flag := range []string{"--all", "--all=false"} {
		cmd, out, diagnostics := canonicalCollectionTestCommand(t, false, "--output-format=json", flag)
		var exit *exitError
		if err := runCanonicalCollection(cmd, false); !errors.As(err, &exit) || exit.Code != 5 {
			t.Fatalf("Issue selection flag silently admitted: %v", err)
		}
		if !strings.Contains(diagnostics.String(), "does not implement --all") || !bytes.Contains(out.Bytes(), []byte(`"outcome":"failure"`)) {
			t.Fatalf("missing explicit refusal: %s %s", out, diagnostics)
		}
	}
}

func TestCanonicalCollectionTypeGrammarDiagnostics(t *testing.T) {
	for _, links := range []bool{false, true} {
		for _, flag := range []string{"type", "exact-type", "conforms-to"} {
			for _, value := range []string{canonicalCollectionScope + "beads/note", "https://"} {
				cmd, _, _ := canonicalCollectionTestCommand(t, links, "--output-format=json", "--"+flag, value)
				options, _, _ := canonicalOutputOptions(cmd)
				if _, err := canonicalCollectionRequest(cmd, links, options); err == nil {
					t.Fatalf("invalid Type %q admitted by %s", value, flag)
				}
			}
			for _, value := range []string{canonicalCollectionScope + "types/note", "https://types.example.test/external-note"} {
				cmd, _, _ := canonicalCollectionTestCommand(t, links, "--output-format=json", "--"+flag, value)
				options, _, _ := canonicalOutputOptions(cmd)
				if _, err := canonicalCollectionRequest(cmd, links, options); err != nil {
					t.Fatalf("canonical Type %s refused: %v", value, err)
				}
			}
		}
		for _, value := range []string{"bug", "types/note"} {
			cmd, _, _ := canonicalCollectionTestCommand(t, links, "--output-format=json", "--conforms-to", value)
			options, _, _ := canonicalOutputOptions(cmd)
			if _, err := canonicalCollectionRequest(cmd, links, options); err == nil || !strings.Contains(err.Error(), "--conforms-to requires a canonical absolute HTTP(S) Type URI") {
				t.Fatalf("alias spelling guard: %v", err)
			}
		}
	}
}
