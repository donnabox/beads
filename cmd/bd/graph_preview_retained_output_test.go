package main

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
	"time"

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
