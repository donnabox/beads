package main

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/clioutput"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func TestGraphRetainedVersionProjectionUsesActualSummary(t *testing.T) {
	rows := graphVersionRowsJSON([]graphstore.VersionRow{{
		Ordinal: 9007199254740993, Version: "opaque:exact-revision", Actor: "agent", Attribution: "claimed",
		ChangeAt: time.Date(2026, 10, 20, 6, 0, 0, 123000, time.UTC),
	}})
	raw, err := json.Marshal(rows[0])
	if err != nil {
		t.Fatal(err)
	}
	renderer, err := clioutput.New(graphVersionItemDescriptor("fixture-contract"), clioutput.Options{
		Formats: []string{"json"}, Include: []string{"/local_revision"}, Omit: []string{"envelope"},
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := renderer.Render(raw)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"local_revision":9007199254740993,"version":"opaque:exact-revision"}` + "\n"
	if string(got) != want {
		t.Fatalf("summary changed its identity, exact ordinal or shape: %s", got)
	}
	if _, err := clioutput.New(graphVersionItemDescriptor("fixture-contract"), clioutput.Options{Omit: []string{"/version"}}); err == nil {
		t.Fatal("allowed omission of the only citable version token")
	}
}

func TestGraphRetainedJSONLItemsKeepSubjectWithoutEnvelope(t *testing.T) {
	for _, raw := range []bool{false, true} {
		options := clioutput.Options{Formats: []string{"jsonl"}, Include: []string{"/local_revision"}}
		if raw {
			options.Omit = []string{"envelope"}
		}
		renderer, err := newGraphVersionOutput("fixture-contract", options)
		if err != nil {
			t.Fatal(err)
		}
		var out bytes.Buffer
		stream, err := renderer.NewJSONL(&out)
		if err != nil {
			t.Fatal(err)
		}
		item, err := graphVersionStreamItem("beads/source", "issue", map[string]any{"version": "opaque:one", "local_revision": int64(9007199254740993)})
		if err != nil {
			t.Fatal(err)
		}
		if err := stream.Item(item); err != nil {
			t.Fatal(err)
		}
		if err := stream.Finish(true, "exhausted", nil); err != nil {
			t.Fatal(err)
		}
		lines := bytes.Split(bytes.TrimSpace(out.Bytes()), []byte{'\n'})
		wantLines := 3
		if raw {
			wantLines = 1
		}
		if len(lines) != wantLines {
			t.Fatalf("raw=%v got %d JSONL records: %s", raw, len(lines), out.String())
		}
		line := lines[0]
		if !raw {
			line = lines[1]
		}
		var got struct {
			Data       json.RawMessage `json:"data"`
			Resource   string          `json:"resource"`
			Kind       string          `json:"kind"`
			Version    string          `json:"version"`
			Local      json.Number     `json:"local_revision"`
			RecordType string          `json:"record_type"`
		}
		if !raw {
			var wrapper struct {
				RecordType string          `json:"record_type"`
				Data       json.RawMessage `json:"data"`
			}
			if err := json.Unmarshal(line, &wrapper); err != nil || wrapper.RecordType != "item" {
				t.Fatalf("framed item: %v %s", err, line)
			}
			line = wrapper.Data
		}
		if err := json.Unmarshal(line, &got); err != nil || got.Resource != "beads/source" || got.Kind != "issue" || got.Version != "opaque:one" || got.Local.String() != "9007199254740993" {
			t.Fatalf("raw=%v lost version context/precision: %v %+v", raw, err, got)
		}
	}
}

func TestGraphRetainedJSONLRefusesUncarriedHistoryContext(t *testing.T) {
	for _, extra := range []string{"currentState", "terminal_versions", "future_history_context"} {
		result := map[string]any{"resource": "beads/source", "kind": "memory", "versions": []any{}, extra: map[string]any{"revision": "terminal"}}
		if err := graphVersionJSONLContextSupported(result); err == nil {
			t.Fatalf("silently discarded native %s context", extra)
		}
	}
}

func TestGraphRetainedJSONLContextRefusalHasFailureTerminal(t *testing.T) {
	result := map[string]any{"resource": "beads/source", "kind": "memory", "versions": []any{}, "terminal_versions": []any{"deleted"}}
	contextErr := graphVersionJSONLContextSupported(result)
	if contextErr == nil {
		t.Fatal("native terminal context was accepted")
	}
	renderer, err := newGraphVersionOutput("fixture-contract", clioutput.Options{Formats: []string{"jsonl"}})
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	stream, err := renderer.NewJSONL(cmd.OutOrStdout())
	if err != nil {
		t.Fatal(err)
	}
	if failure := canonicalFailure(cmd, graphVersionJSONLFailure{stream}, "capability_unavailable", contextErr.Error(), 5); failure == nil {
		t.Fatal("context refusal returned success")
	}
	lines := bytes.Split(bytes.TrimSpace(stdout.Bytes()), []byte{'\n'})
	if len(lines) != 2 {
		t.Fatalf("expected header and failure terminal, got %s", stdout.String())
	}
	var terminal struct {
		RecordType string `json:"record_type"`
		Outcome    string `json:"outcome"`
		Error      struct {
			Problem struct {
				Code string `json:"code"`
			} `json:"problem"`
		} `json:"error"`
	}
	if err := json.Unmarshal(lines[1], &terminal); err != nil || terminal.RecordType != "terminal" || terminal.Outcome != "failure" || terminal.Error.Problem.Code != "capability_unavailable" || stderr.Len() == 0 {
		t.Fatalf("context refusal lost failure terminal: %v %s", err, stdout.String())
	}
}

func TestGraphRetainedComparisonPreservesActualProviderValues(t *testing.T) {
	// This exercises the real provider carrier and CLI carried-attribution
	// projection. It does not invent a BDP History comparison DTO or fetch
	// snapshots to fill fields absent from this operation.
	original := graphstore.VersionComparison{
		Resource: graphstore.VersionComparisonResource{ID: "https://example.test/links/a", Type: "https://example.test/types/related", Source: "https://example.test/beads/a", Target: "https://example.test/beads/b"},
		From: graphstore.VersionComparisonEndpoint{Version: "z-requested-first",
			Attribution: graphstore.Attribution{Actor: "agent", Status: "claimed", RecordedAt: "2026-10-20T06:00:00Z"}},
		To:       graphstore.VersionComparisonEndpoint{Version: "a-requested-second"},
		Compared: []string{"properties", "metadata"}, Unsupported: []string{},
		Changes: []graphstore.VersionChange{
			{Area: "properties", Member: "escaped/key~", From: graphstore.VersionValue{Present: true, Value: json.RawMessage(`null`)}, To: graphstore.VersionValue{Present: true, Value: json.RawMessage(`9007199254740993`)}},
			{Area: "metadata", Member: "missing", From: graphstore.VersionValue{Present: false}, To: graphstore.VersionValue{Present: true, Value: json.RawMessage(`null`)}},
		},
	}
	raw, err := graphProjectCompleteRecords(original)
	if err != nil {
		t.Fatal(err)
	}
	options, err := graphComparisonOptions(clioutput.Options{Formats: []string{"json"}, Omit: []string{"envelope"}})
	if err != nil {
		t.Fatal(err)
	}
	renderer, err := clioutput.New(graphComparisonDescriptor("fixture-contract"), options)
	if err != nil {
		t.Fatal(err)
	}
	got, err := renderer.Render(raw)
	if err != nil {
		t.Fatal(err)
	}
	var wantValue, gotValue any
	wantDecoder := json.NewDecoder(bytes.NewReader(raw))
	wantDecoder.UseNumber()
	gotDecoder := json.NewDecoder(bytes.NewReader(got))
	gotDecoder.UseNumber()
	if err := wantDecoder.Decode(&wantValue); err != nil {
		t.Fatal(err)
	}
	if err := gotDecoder.Decode(&gotValue); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(wantValue, gotValue) {
		t.Fatalf("comparison operands, paths, presence or exact values changed: %s", got)
	}
}

func TestGraphRetainedComparisonRefusesUnsupportedProjection(t *testing.T) {
	for _, options := range []clioutput.Options{
		{Include: []string{"/properties"}},
		{Include: []string{""}},
		{Omit: []string{"/changes"}},
		{Omit: []string{"/from"}},
		{Omit: []string{"/properties/large"}},
	} {
		if _, err := graphComparisonOptions(options); err == nil {
			t.Fatalf("unsupported comparison mask accepted: %+v", options)
		}
	}
}
