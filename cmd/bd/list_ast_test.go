package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
