package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/resourcequery"
)

func TestListFieldFlagsComposeWithoutAST(t *testing.T) {
	previous := graphPreviewActive
	graphPreviewActive = true
	t.Cleanup(func() { graphPreviewActive = previous })
	for _, links := range []bool{false, true} {
		cmd, _, _ := canonicalCollectionTestCommand(t, links, "--property-field", "status=active", "--metadata-field", "gc.team=a=b", "--has-property-key", "body", "--has-metadata-key", "reviewed")
		if !canonicalOutputSelected(cmd) {
			t.Fatal("field flags did not select canonical graph List")
		}
		fields, err := collectionFieldConditions(cmd)
		if err != nil {
			t.Fatal(err)
		}
		records := []json.RawMessage{
			json.RawMessage(`{"properties":{"status":"active","body":"note"},"metadata":{"gc.team":"a=b","reviewed":null}}`),
			json.RawMessage(`{"properties":{"status":"draft","body":"note"},"metadata":{"gc.team":"a=b","reviewed":null}}`),
			json.RawMessage(`{"properties":{"status":"active","body":"note"},"metadata":{"gc":{"team":"a=b"},"reviewed":null}}`),
			json.RawMessage(`{"properties":{"status":"active","body":"note"},"metadata":{"gc.team":"a=b"}}`),
		}
		got, err := fields.Select(context.Background(), records)
		if err != nil || len(got) != 1 || string(got[0]) != string(records[0]) {
			t.Fatalf("literal field AND composition: %s %v", got, err)
		}
	}
}

func TestListFieldFlagsRejectBeforeStore(t *testing.T) {
	for _, flags := range [][]string{
		{"--property-field", ""},
		{"--property-field", "bad"},
		{"--metadata-field", ""},
		{"--metadata-field", "key=value", "--metadata-field", ""},
		{"--metadata-field", "=value"},
		{"--metadata-field", "bad key=value"},
		{"--has-property-key", ""},
		{"--has-property-key", "body", "--has-property-key", ""},
		{"--has-metadata-key", ""},
	} {
		cmd, out, _ := canonicalCollectionTestCommand(t, false, append([]string{"--output-format", "json"}, flags...)...)
		err := runCanonicalCollection(cmd, false)
		var exit *exitError
		if !errors.As(err, &exit) || exit.Code != 2 || strings.Contains(out.String(), "graph_not_initialized") {
			t.Fatalf("field admission reached store for %q: %v %s", flags, err, out.String())
		}
	}
}

func TestListFieldLiteralArguments(t *testing.T) {
	for _, tc := range []struct {
		args   []string
		record string
	}{
		{[]string{"--property-field", "title=a,b"}, `{"properties":{"title":"a,b"}}`},
		{[]string{"--property-field", "title=a,b", "--property-field", "title=a,b"}, `{"properties":{"title":"a,b"}}`},
		{[]string{"--metadata-field", `note=a,"b"`}, `{"metadata":{"note":"a,\"b\""}}`},
		{[]string{"--metadata-field", "token=a=b,c"}, `{"metadata":{"token":"a=b,c"}}`},
	} {
		cmd, _, _ := canonicalCollectionTestCommand(t, false, tc.args...)
		fields, err := collectionFieldConditions(cmd)
		if err != nil {
			t.Fatal(err)
		}
		got, err := fields.Select(context.Background(), []json.RawMessage{json.RawMessage(tc.record)})
		if err != nil || len(got) != 1 {
			t.Fatalf("literal arguments %q changed: %s %v", tc.args, got, err)
		}
	}
}

func TestListCountOnlyBeforePageAndProjection(t *testing.T) {
	for _, links := range []bool{false, true} {
		cmd, out, _ := canonicalCollectionTestCommand(t, links, "--count-only", "--output-format", "json", "--page-size", "1", "--limit", "1", "--include", "/properties/title", "--omit", "envelope")
		plan, failure := canonicalCollectionTestPlan(t, cmd, links)
		if err := renderCanonicalCollection(cmd, plan, []json.RawMessage{canonicalCollectionRecord("a"), canonicalCollectionRecord("b")}, failure); err != nil {
			t.Fatal(err)
		}
		var result struct {
			Count int `json:"count"`
		}
		if err := json.Unmarshal(out.Bytes(), &result); err != nil || result.Count != 2 {
			t.Fatalf("count followed page/mask: %s %v", out.String(), err)
		}
	}
}

func TestListNewFlagsPreserveDependencyMetadataRoute(t *testing.T) {
	previous := graphPreviewActive
	graphPreviewActive = false
	t.Cleanup(func() { graphPreviewActive = previous })
	cmd, _, _ := canonicalCollectionTestCommand(t, false, "--metadata-field", "n=1", "--has-metadata-key", "n")
	if canonicalOutputSelected(cmd) || collectionFieldFlagsSelected(cmd) {
		t.Fatal("legacy metadata flags changed dependency route")
	}
	for _, flags := range [][]string{{"--property-field", "status=active"}, {"--has-property-key", "body"}, {"--count-only"}} {
		cmd, out, _ := canonicalCollectionTestCommand(t, false, append([]string{"--output-format", "json"}, flags...)...)
		err := runListCore(cmd, nil)
		var exit *exitError
		if !errors.As(err, &exit) || exit.Code != 5 || !strings.Contains(out.String(), `"code":"capability_unavailable"`) {
			t.Fatalf("new graph flags fell through dependency List: %v %s", err, out.String())
		}
	}
}

func TestListFieldBudgetRefusesBeforeStore(t *testing.T) {
	args := []string{"--output-format", "json"}
	for i := 0; i < 257; i++ {
		args = append(args, "--has-property-key", "body")
	}
	cmd, out, _ := canonicalCollectionTestCommand(t, false, args...)
	err := runCanonicalCollection(cmd, false)
	var exit *exitError
	if !errors.As(err, &exit) || exit.Code != 5 || !strings.Contains(out.String(), `"code":"capability_unavailable"`) || strings.Contains(out.String(), "graph_not_initialized") {
		t.Fatalf("budget error misclassified or reached store: %v %s", err, out.String())
	}
}

func TestListFieldOperationalErrorClassification(t *testing.T) {
	for _, tc := range []struct {
		cause  error
		code   string
		status int
	}{
		{resourcequery.ErrLimit, "capability_unavailable", 5},
		{context.Canceled, "capability_unavailable", 5},
		{context.DeadlineExceeded, "capability_unavailable", 5},
		{resourcequery.ErrDataset, "invalid_store", 5},
		{resourcequery.ErrInvalid, "invalid_selector", 2},
	} {
		cmd, out, _ := canonicalCollectionTestCommand(t, false, "--output-format", "json")
		_, failure := canonicalCollectionTestPlan(t, cmd, false)
		err := canonicalCollectionRequestFailure(cmd, failure, tc.cause)
		var exit *exitError
		if !errors.As(err, &exit) || exit.Code != tc.status || !strings.Contains(out.String(), `"code":"`+tc.code+`"`) {
			t.Fatalf("%v classification: %v %s", tc.cause, err, out.String())
		}
	}
}

const listASTScanFixture = `{"ir":"bdp-query-design-1","profile":"list","selection":{"op":"scan","collection":"beads"},"result":{"form":"resources"}}`

func TestListASTInputAndOutputOptIn(t *testing.T) {
	path := filepath.Join(t.TempDir(), "list.json")
	if err := os.WriteFile(path, []byte(listASTScanFixture), 0600); err != nil {
		t.Fatal(err)
	}
	for _, input := range []string{listASTScanFixture, "@" + path, "-"} {
		t.Run(input[:1], func(t *testing.T) {
			cmd, _, _ := canonicalCollectionTestCommand(t, false, "--ast", input)
			cmd.SetIn(strings.NewReader(listASTScanFixture))
			if !canonicalOutputSelected(cmd) {
				t.Fatal("AST did not opt into canonical List")
			}
			options, selected, err := canonicalOutputOptions(cmd)
			if err != nil || !selected || len(options.Formats) != 0 {
				t.Fatalf("List output contract: %+v %t %v", options, selected, err)
			}
			raw, err := readCollectionAST(cmd, false)
			if err != nil || string(raw) != listASTScanFixture {
				t.Fatalf("List input: %s %v", raw, err)
			}
		})
	}
	cmd, _, _ := canonicalCollectionTestCommand(t, false)
	if canonicalOutputSelected(cmd) {
		t.Fatal("plain legacy List changed output contract")
	}
	if raw, err := readCollectionAST(cmd, false); raw != nil || err != nil {
		t.Fatalf("no-AST path read input: %s %v", raw, err)
	}
	cmd, _, _ = canonicalCollectionTestCommand(t, false, "--ast", listASTScanFixture, "--json")
	if _, _, err := canonicalOutputOptions(cmd); err == nil {
		t.Fatal("legacy JSON silently mixed into canonical AST output")
	}
}

func TestListASTRefusalBeforeOpeningStore(t *testing.T) {
	for _, tc := range []struct {
		name, raw string
		links     bool
		status    int
	}{
		{"malformed", `{`, false, 2},
		{"mismatched population", listASTScanFixture, true, 2},
		{"query profile", strings.Replace(listASTScanFixture, `"profile":"list"`, `"profile":"query"`, 1), false, 2},
		{"count control", strings.Replace(listASTScanFixture, `"form":"resources"`, `"form":"count"`, 1), false, 5},
		{"unknown sort control", strings.TrimSuffix(listASTScanFixture, "}") + `,"sort":["id"]}`, false, 2},
		{"AST projection control", strings.TrimSuffix(listASTScanFixture, "}") + `,"projection":{"include":["/properties/title"]}}`, false, 5},
		{"representation mismatch", `{"ir":"bdp-query-design-1","profile":"list","selection":{"op":"filter","input":{"op":"scan","collection":"beads"},"bind":"item","predicate":{"op":"present","value":{"op":"field","of":{"op":"binding","name":"item"},"path":["revisionProvenance"]}}},"result":{"form":"resources"}}`, false, 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd, out, _ := canonicalCollectionTestCommand(t, tc.links, "--output-format", "json", "--ast", tc.raw)
			err := runCanonicalCollection(cmd, tc.links)
			var exit *exitError
			if !errors.As(err, &exit) || exit.Code != tc.status || strings.Contains(out.String(), "graph_not_initialized") || strings.Contains(out.String(), `"items"`) {
				t.Fatalf("List refusal reached store/returned membership: %v %s", err, out.String())
			}
		})
	}
}

func TestListASTDependencyModeRefusesBeforeLegacyList(t *testing.T) {
	previous := graphPreviewActive
	graphPreviewActive = false
	t.Cleanup(func() { graphPreviewActive = previous })
	cmd, out, _ := canonicalCollectionTestCommand(t, false, "--output-format", "json", "--ast", listASTScanFixture)
	err := runListCore(cmd, nil)
	var exit *exitError
	if !errors.As(err, &exit) || exit.Code != 5 || !strings.Contains(out.String(), `"code":"capability_unavailable"`) {
		t.Fatalf("dependency-mode AST fell through: %v %s", err, out.String())
	}
	cmd, out, _ = canonicalCollectionTestCommand(t, true, "--output-format", "json", "--ast", strings.Replace(listASTScanFixture, `"collection":"beads"`, `"collection":"links"`, 1))
	err = runGraphPreviewLinkCollection(cmd, nil)
	if !errors.As(err, &exit) || exit.Code != 5 || !strings.Contains(out.String(), `"code":"capability_unavailable"`) {
		t.Fatalf("dependency-mode Link AST fell through: %v %s", err, out.String())
	}
}
