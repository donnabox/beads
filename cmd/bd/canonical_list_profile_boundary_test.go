package main

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/resourcequery"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// Decorate a detached, runnable leaf using the actual semantic-profile adapter.
// The callback is not executed. No production command or provider is installed.
func canonicalListBoundaryProfile(t *testing.T, cmd *cobra.Command, links bool) {
	t.Helper()
	cmd.RunE = func(*cobra.Command, []string) error { return nil }
	root := &cobra.Command{Use: "bd"}
	path := []string{"list"}
	if links {
		group := &cobra.Command{Use: "link"}
		root.AddCommand(group)
		group.AddCommand(cmd)
		path = []string{"link", "list"}
	} else {
		root.AddCommand(cmd)
	}
	decoration, err := prepareCanonicalReadLeaf(root, cliSurfaceCanonical, path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(decoration.Restore)
}

func TestCanonicalListFixedEqualityRefusesBeforeAcquisition(t *testing.T) {
	for _, links := range []bool{false, true} {
		kind := "beads"
		if links {
			kind = "links"
		}
		for _, flag := range []string{"property-field", "metadata-field"} {
			for _, variant := range []struct {
				name string
				args []string
			}{
				{name: "plain"},
				{name: "count", args: []string{"--count-only"}},
				// If the field boundary moved after AST input, this deliberately
				// invalid file spelling would produce an input error instead.
				{name: "ast", args: []string{"--ast", "@"}},
				{name: "ast-count", args: []string{"--ast", "@", "--count-only"}},
			} {
				t.Run(kind+"/"+flag+"/"+variant.name, func(t *testing.T) {
					args := append([]string{"--output-format=json", "--unpaged", "--" + flag, "n=1"}, variant.args...)
					cmd, out, _ := canonicalCollectionTestCommand(t, links, args...)
					canonicalListBoundaryProfile(t, cmd, links)
					// Exercise pure admission as well as the real collection entry.
					options, _, err := canonicalOutputOptions(cmd)
					if err != nil {
						t.Fatal(err)
					}
					if err := canonicalCollectionPreflight(cmd, links, options); !errors.Is(err, graphstore.ErrCapabilityUnavailable) || !strings.Contains(err.Error(), "compat SQL-text equality") {
						t.Fatalf("pure pre-acquisition boundary: %v", err)
					}
					err = runCanonicalCollection(cmd, links)
					var exit *exitError
					if !errors.As(err, &exit) || exit.Code != 5 || !strings.Contains(out.String(), `"code":"capability_unavailable"`) || !strings.Contains(out.String(), "compat SQL-text equality") || strings.Contains(out.String(), "graph_not_initialized") {
						t.Fatalf("field boundary reached AST/store path: err=%v out=%s", err, out.String())
					}
				})
			}
		}
	}
}

func TestCompatListFixedEqualityPreservesCanonicalPresentation(t *testing.T) {
	for _, links := range []bool{false, true} {
		for _, namespace := range []string{"properties", "metadata"} {
			flag := "property-field"
			if namespace == "metadata" {
				flag = "metadata-field"
			}
			for _, outputDefault := range []bool{false, true} {
				cmd, _, _ := canonicalCollectionTestCommand(t, links, "--output-format=json", "--"+flag, "n=1")
				if outputDefault {
					cmd.Annotations = map[string]string{canonicalDefaultAnnotation: "1"}
				}
				if canonicalCLIProfileSelected(cmd) || !canonicalOutputSelected(cmd) {
					t.Fatal("presentation selected a semantic profile")
				}
				fields, err := collectionFieldConditions(cmd)
				if err != nil {
					t.Fatal(err)
				}
				// Exact direct-List compatibility: number and string both
				// compare as SQL text. This must not become typed AST equality.
				records := []json.RawMessage{
					json.RawMessage(`{"` + namespace + `":{"n":1}}`),
					json.RawMessage(`{"` + namespace + `":{"n":"1"}}`),
					json.RawMessage(`{"` + namespace + `":{"n":2}}`),
				}
				got, err := fields.Select(context.Background(), records)
				if err != nil || len(got) != 2 || string(got[0]) != string(records[0]) || string(got[1]) != string(records[1]) {
					t.Fatalf("compat SQL-text behavior changed: got=%s err=%v", got, err)
				}
			}
		}
	}
}

func TestCanonicalListPresenceBoundaryPreservesLiteralKeys(t *testing.T) {
	for _, links := range []bool{false, true} {
		cmd, _, _ := canonicalCollectionTestCommand(t, links, "--output-format=json", "--has-property-key", "a.b", "--has-property-key", "a.b", "--has-metadata-key", "reviewed")
		canonicalListBoundaryProfile(t, cmd, links)
		fields, err := collectionFieldConditions(cmd)
		if err != nil {
			t.Fatal(err)
		}
		records := []json.RawMessage{
			json.RawMessage(`{"properties":{"a.b":null},"metadata":{"reviewed":null}}`),
			json.RawMessage(`{"properties":{"a":{"b":null}},"metadata":{"reviewed":null}}`),
			json.RawMessage(`{"properties":{"a.b":null},"metadata":{}}`),
		}
		got, err := fields.Select(context.Background(), records)
		if err != nil || len(got) != 1 || string(got[0]) != string(records[0]) {
			t.Fatalf("presence changed null/literal-key/AND semantics: %s %v", got, err)
		}
	}
}

func TestCanonicalListFixedBoundaryDoesNotCoerceTypedAST(t *testing.T) {
	raw := []byte(`{"ir":"bdp-query-design-1","profile":"list","selection":{"op":"filter","input":{"op":"scan","collection":"beads"},"bind":"candidate","predicate":{"op":"compare","operator":"eq","left":{"op":"field","of":{"op":"binding","name":"candidate"},"path":["metadata","n"]},"right":{"op":"literal","value":1}}},"result":{"form":"resources"}}`)
	data := resourcequery.Dataset{}
	for _, body := range []string{
		`{"id":"https://example.test/number","type":"https://example.test/Note","revision":"r1","properties":{},"metadata":{"n":1}}`,
		`{"id":"https://example.test/string","type":"https://example.test/Note","revision":"r1","properties":{},"metadata":{"n":"1"}}`,
	} {
		record, err := resourcequery.NewResource(graph.KindBead, []byte(body))
		if err != nil {
			t.Fatal(err)
		}
		data.Resources = append(data.Resources, record)
	}
	response, err := resourcequery.Execute(context.Background(), raw, data, resourcequery.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	result, ok := response.Result.(resourcequery.ResourcesResult)
	if !ok || !response.Evaluation.Complete || len(result.Items) != 1 {
		t.Fatalf("typed AST did not remain distinct: %+v", response)
	}
	var item struct{ ID string }
	if err := json.Unmarshal(result.Items[0], &item); err != nil || item.ID != "https://example.test/number" {
		t.Fatalf("typed AST selected the wrong candidate: %s %v", result.Items[0], err)
	}
}
