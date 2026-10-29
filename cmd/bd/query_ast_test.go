package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/steveyegge/beads/internal/resourcequery"
)

const queryASTCountFixture = `{"ir":"bdp-query-design-1","profile":"list","selection":{"op":"scan","collection":"beads"},"result":{"form":"count"}}`

func queryASTTestCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "query"}
	cmd.Flags().StringArray("ast", nil, "")
	cmd.Flags().Bool("json", false, "")
	registerCanonicalOutputFlags(cmd)
	return cmd
}

func TestQueryASTInputsAreOneExactDocument(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "query.json")
	if err := os.WriteFile(path, []byte(queryASTCountFixture), 0600); err != nil {
		t.Fatal(err)
	}
	for _, input := range []string{queryASTCountFixture, "@" + path, "-"} {
		cmd := queryASTTestCommand()
		cmd.SetIn(strings.NewReader(queryASTCountFixture))
		if err := cmd.Flags().Set("ast", input); err != nil {
			t.Fatal(err)
		}
		raw, err := readQueryAST(cmd, nil)
		if err != nil || string(raw) != queryASTCountFixture {
			t.Fatalf("input %q: %s %v", input, raw, err)
		}
		request, err := resourcequery.Decode(raw)
		if err != nil || request.ResultForm() != "count" {
			t.Fatalf("AST differs by input transport: %v", err)
		}
	}
}

func TestQueryASTInputRefusesAmbiguityAndOversize(t *testing.T) {
	for _, tc := range []struct {
		name   string
		inputs []string
		args   []string
	}{
		{"missing", nil, nil},
		{"repeated", []string{queryASTCountFixture, queryASTCountFixture}, nil},
		{"expression", []string{queryASTCountFixture}, []string{"status=open"}},
		{"empty filename", []string{"@"}, nil},
		{"oversize", []string{strings.Repeat(" ", resourcequery.DefaultLimits().RequestBytes+1)}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := queryASTTestCommand()
			for _, input := range tc.inputs {
				if err := cmd.Flags().Set("ast", input); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := readQueryAST(cmd, tc.args); err == nil {
				t.Fatal("expected refusal")
			}
		})
	}
}

func TestQueryASTUsesExistingEnvelopeAndRawDelivery(t *testing.T) {
	for _, raw := range []bool{false, true} {
		cmd := queryASTTestCommand()
		if raw {
			if err := cmd.Flags().Set("omit", "envelope"); err != nil {
				t.Fatal(err)
			}
		}
		renderer, err := queryASTRenderer(cmd, "count")
		if err != nil {
			t.Fatal(err)
		}
		response := resourcequery.Response{Result: resourcequery.CountResult{Form: "count", ResourceKind: "bead", Count: 2},
			Evaluation: resourcequery.Evaluation{Scope: "response", Complete: true, Traversals: []resourcequery.Traversal{}}}
		var out bytes.Buffer
		cmd.SetOut(&out)
		if err := renderCanonicalResult(cmd, renderer, response); err != nil {
			t.Fatal(err)
		}
		var value map[string]json.RawMessage
		if err := json.Unmarshal(out.Bytes(), &value); err != nil {
			t.Fatal(err)
		}
		if raw {
			if value["result"] == nil || value["evaluation"] == nil || value["schema_version"] != nil {
				t.Fatal(out.String())
			}
		} else {
			if value["data"] == nil || string(value["result_kind"]) != `"aggregate"` || value["schema_version"] == nil {
				t.Fatal(out.String())
			}
		}
	}
}

// Exercise the registered command and the gates called by root preRun, rather
// than substituting a detached renderer helper for command admission. These
// tests never initialize a graph or open legacy storage.
func TestQueryASTRegisteredPreRunAndModeAdmission(t *testing.T) {
	stdioMutex.Lock()
	t.Cleanup(stdioMutex.Unlock)
	isolateBeadsDirForTest(t)
	saveAndRestoreGlobals(t)
	t.Setenv("BEADS_DIR", t.TempDir())
	t.Setenv("BD_GRAPH_MODE", "")
	oldActive, oldConfig, oldDir, oldGeneration := graphPreviewActive, graphPreviewConfig, graphPreviewDir, graphPreviewTypeGeneration
	t.Cleanup(func() {
		graphPreviewActive, graphPreviewConfig, graphPreviewDir, graphPreviewTypeGeneration = oldActive, oldConfig, oldDir, oldGeneration
	})
	registered, _, err := rootCmd.Find([]string{"query"})
	if err != nil || registered != queryCmd {
		t.Fatalf("Query is not registered: %v", err)
	}
	pinned := `{"ir":"bdp-query-design-1","profile":"query","selection":{"op":"refs","collection":"beads","references":[{"uri":"https://query.example/A","revision":"r1"}]},"result":{"form":"resources"}}`
	malformedRefs := `{"ir":"bdp-query-design-1","profile":"query","selection":{"op":"refs","collection":"beads","references":7},"result":{"form":"resources"}}`
	batchRefs := strings.Replace(malformedRefs, `"references":7`, `"references":{"binding":"prior"}`, 1)
	projection := strings.TrimSuffix(queryASTCountFixture, "}") + `,"projection":{"omit":["/properties"]}}`
	// This valid flat AST fits hard decoding bounds but exceeds the service's
	// default node budget, independently of data acquisition or runtime work.
	bounded := `{"ir":"bdp-query-design-1","profile":"list","selection":{"op":"filter","input":{"op":"scan","collection":"beads"},"bind":"b","predicate":{"op":"and","operands":[` + strings.Repeat("true,", resourcequery.DefaultLimits().Nodes-1) + `true]}},"result":{"form":"count"}}`
	for _, tc := range []struct {
		name                                 string
		ast, omit, format                    string
		invalidFormat, malformed, capability bool
		bound                                bool
	}{
		{name: "raw admission", ast: queryASTCountFixture, omit: "envelope"},
		{name: "wrapped admission", ast: queryASTCountFixture},
		{name: "invalid format before workspace", ast: queryASTCountFixture, format: "bogus", invalidFormat: true},
		{name: "malformed AST before acquisition", ast: `{"ir":"bdp-query-design-1"}`, malformed: true},
		{name: "malformed refs before acquisition", ast: malformedRefs, malformed: true},
		{name: "batch refs before acquisition", ast: batchRefs, capability: true},
		{name: "pinned refs before acquisition", ast: pinned, capability: true},
		{name: "pinned refs raw before acquisition", ast: pinned, omit: "envelope", capability: true},
		{name: "projection before acquisition", ast: projection, capability: true},
		{name: "projection raw before acquisition", ast: projection, omit: "envelope", capability: true},
		{name: "node bound before acquisition", ast: bounded, capability: true, bound: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, name := range []string{"ast", "omit", "output-format", "include"} {
				flag := queryCmd.Flags().Lookup(name)
				value := flag.Value.(pflag.SliceValue)
				saved, changed := append([]string(nil), value.GetSlice()...), flag.Changed
				t.Cleanup(func() { _ = value.Replace(saved); flag.Changed = changed })
				if err := value.Replace(nil); err != nil {
					t.Fatal(err)
				}
				flag.Changed = false
			}
			oldOut, oldErr := queryCmd.OutOrStdout(), queryCmd.ErrOrStderr()
			t.Cleanup(func() { queryCmd.SetOut(oldOut); queryCmd.SetErr(oldErr) })
			var out, diagnostic bytes.Buffer
			queryCmd.SetOut(&out)
			queryCmd.SetErr(&diagnostic)
			if err := queryCmd.Flags().Set("ast", tc.ast); err != nil {
				t.Fatal(err)
			}
			if tc.omit != "" {
				if err := queryCmd.Flags().Set("omit", tc.omit); err != nil {
					t.Fatal(err)
				}
			}
			if tc.format != "" {
				if err := queryCmd.Flags().Set("output-format", tc.format); err != nil {
					t.Fatal(err)
				}
			}
			if !canonicalOutputSelected(queryCmd) {
				t.Fatal("AST did not select canonical command admission")
			}
			if tc.invalidFormat {
				err := rootCmd.PersistentPreRunE(queryCmd, nil)
				var exit *exitError
				if !errors.As(err, &exit) || exit.Code != 2 || out.Len() != 0 ||
					!strings.Contains(diagnostic.String(), "invalid_selector") || !strings.Contains(diagnostic.String(), "format") {
					t.Fatalf("invalid format reached workspace admission or wrong failure: %v %s / %s", err, out.String(), diagnostic.String())
				}
				return
			}
			if err := validateCanonicalOutputInvocation(queryCmd); err != nil {
				t.Fatalf("root output gate rejected AST: %v", err)
			}
			if tc.malformed || tc.capability {
				if tc.bound {
					request, err := resourcequery.Decode([]byte(tc.ast))
					if err != nil {
						t.Fatalf("bound case must be structurally valid: %v", err)
					}
					if err := resourcequery.Admit(request, resourcequery.DefaultLimits()); !errors.Is(err, resourcequery.ErrLimit) {
						t.Fatalf("bound case must exceed the service admission budget: %v", err)
					}
				}
				graphPreviewActive = true
				graphPreviewConfig = nil
				err = queryCmd.RunE(queryCmd, nil)
				wantCode, wantExit := "invalid_selector", 2
				if tc.capability {
					wantCode, wantExit = "capability_unavailable", 5
				}
				var exit *exitError
				if !errors.As(err, &exit) || exit.Code != wantExit || !strings.Contains(diagnostic.String(), wantCode) {
					t.Fatalf("AST reached storage or wrong refusal: %v %s / %s", err, out.String(), diagnostic.String())
				}
				if tc.omit == "envelope" {
					if out.Len() != 0 {
						t.Fatalf("raw refusal emitted stdout: %s", out.String())
					}
				} else {
					var failure struct {
						Outcome string          `json:"outcome"`
						Data    json.RawMessage `json:"data"`
						Error   struct {
							Origin  string `json:"origin"`
							Problem struct {
								Code string `json:"code"`
							} `json:"problem"`
						} `json:"error"`
					}
					if err := json.Unmarshal(out.Bytes(), &failure); err != nil {
						t.Fatal(err)
					}
					if failure.Outcome != "failure" || failure.Data != nil || failure.Error.Origin != "cli" || failure.Error.Problem.Code != wantCode {
						t.Fatalf("wrong canonical refusal: %s", out.String())
					}
				}
				return
			}
			handled, err := admitCanonicalGraphPreview(queryCmd)
			var exit *exitError
			if !handled || !errors.As(err, &exit) || exit.Code != 5 || graphPreviewActive {
				t.Fatalf("AST would fall through to legacy opening: handled=%t error=%v", handled, err)
			}
			if !strings.Contains(diagnostic.String(), "capability_unavailable") {
				t.Fatalf("missing admission diagnostic: %s", diagnostic.String())
			}
			if tc.omit == "envelope" {
				if out.Len() != 0 {
					t.Fatalf("raw admission emitted stdout: %s", out.String())
				}
			} else {
				var wrapped struct {
					Outcome string `json:"outcome"`
				}
				if err := json.Unmarshal(out.Bytes(), &wrapped); err != nil || wrapped.Outcome != "failure" || !strings.Contains(out.String(), "capability_unavailable") {
					t.Fatalf("wrapped admission omitted failure envelope: %s (%v)", out.String(), err)
				}
			}
		})
	}
}

func TestQueryASTOutputGateValidatesFormats(t *testing.T) {
	stdioMutex.Lock()
	t.Cleanup(stdioMutex.Unlock)
	for _, tc := range []struct {
		name    string
		formats []string
		raw     bool
		invalid bool
		legacy  bool
		include string
	}{
		{name: "unsupported", formats: []string{"bogus"}, invalid: true},
		{name: "conflicting", formats: []string{"json", "detail"}, invalid: true},
		{name: "detail raw", formats: []string{"detail"}, raw: true, invalid: true},
		{name: "default JSON"},
		{name: "explicit JSON", formats: []string{"json"}},
		{name: "raw JSON", formats: []string{"json"}, raw: true},
		{name: "default raw JSON", raw: true},
		{name: "legacy query unchanged", legacy: true},
		{name: "legacy format requires AST", legacy: true, formats: []string{"json"}, invalid: true},
		{name: "legacy raw requires AST", legacy: true, raw: true, invalid: true},
		{name: "legacy include requires AST", legacy: true, include: "/metadata", invalid: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, name := range []string{"ast", "omit", "output-format", "include"} {
				flag := queryCmd.Flags().Lookup(name)
				value := flag.Value.(pflag.SliceValue)
				saved, changed := append([]string(nil), value.GetSlice()...), flag.Changed
				t.Cleanup(func() { _ = value.Replace(saved); flag.Changed = changed })
				if err := value.Replace(nil); err != nil {
					t.Fatal(err)
				}
				flag.Changed = false
			}
			oldOut, oldErr := queryCmd.OutOrStdout(), queryCmd.ErrOrStderr()
			t.Cleanup(func() { queryCmd.SetOut(oldOut); queryCmd.SetErr(oldErr) })
			var out, diagnostic bytes.Buffer
			queryCmd.SetOut(&out)
			queryCmd.SetErr(&diagnostic)
			if !tc.legacy {
				if err := queryCmd.Flags().Set("ast", queryASTCountFixture); err != nil {
					t.Fatal(err)
				}
			}
			if tc.include != "" {
				if err := queryCmd.Flags().Set("include", tc.include); err != nil {
					t.Fatal(err)
				}
			}
			for _, format := range tc.formats {
				if err := queryCmd.Flags().Set("output-format", format); err != nil {
					t.Fatal(err)
				}
			}
			if tc.raw {
				if err := queryCmd.Flags().Set("omit", "envelope"); err != nil {
					t.Fatal(err)
				}
			}
			err := validateCanonicalOutputInvocation(queryCmd)
			if !tc.invalid {
				if err != nil || out.Len() != 0 || diagnostic.Len() != 0 {
					t.Fatalf("valid JSON control refused: %v %s / %s", err, out.String(), diagnostic.String())
				}
				return
			}
			var exit *exitError
			if !errors.As(err, &exit) || exit.Code != 2 || out.Len() != 0 || !strings.Contains(diagnostic.String(), "invalid_selector") {
				t.Fatalf("invalid output control admitted or wrong failure: %v %s / %s", err, out.String(), diagnostic.String())
			}
		})
	}
}
