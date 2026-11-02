package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/spf13/cobra"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/resourceget"
)

func TestNativeConcreteCLICollectionSelection(t *testing.T) {
	const scope = "https://example.test/native/"
	task, _ := graphstore.NativeIssueTypeURL(scope, "task")
	blocks, _ := graphstore.NativeDependencyTypeURL(scope, "blocks")
	dependsOn, _ := graphstore.NativeDependencyTypeURL(scope, "depends-on")
	blockedBy, _ := graphstore.NativeDependencyTypeURL(scope, "blocked-by")
	for _, tc := range []struct {
		name          string
		links         bool
		flags         []string
		exact, family string
		refuses       bool
	}{
		{name: "default-issue-family", family: graphstore.IssueTypeURL(scope)},
		{name: "default-link-all", links: true},
		{name: "issue-family", flags: []string{"--type=issue"}, family: graphstore.IssueTypeURL(scope)},
		{name: "task-shorthand-family", flags: []string{"--type=task"}, family: task},
		{name: "task-exact", flags: []string{"--exact-type=task"}, exact: task},
		{name: "feature-alias", flags: []string{"--type=feat"}, family: scope + "feature"},
		{name: "decision-alias", flags: []string{"--exact-type=adr"}, exact: scope + "decision"},
		{name: "literal-full-alias", flags: []string{"--type=" + scope + "feat"}, family: scope + "feat"},
		{name: "literal-segmented-alias", flags: []string{"--type=types/adr"}, family: scope + "types/adr"},
		{name: "custom-case", flags: []string{"--type=Future-Name"}, family: scope + "Future-Name"},
		{name: "task-full-exact", flags: []string{"--exact-type=" + task}, exact: task},
		{name: "memory-family", flags: []string{"--type=memory"}, family: scope + "memory"},
		{name: "literal-prefixed-memory", flags: []string{"--type=types/memory"}, family: scope + "types/memory"},
		{name: "literal-prefixed-decision", flags: []string{"--type=types/decision"}, family: scope + "types/decision"},
		{name: "native-decision-distinct", flags: []string{"--type=decision"}, family: scope + "decision"},
		{name: "blocks-family", links: true, flags: []string{"--type=blocks"}, family: blocks},
		{name: "literal-depends-on-family", links: true, flags: []string{"--type=depends-on"}, family: dependsOn},
		{name: "literal-blocked-by-exact", links: true, flags: []string{"--exact-type=blocked-by"}, exact: blockedBy},
		{name: "dependency-family", links: true, flags: []string{"--type=dependency"}, family: graphstore.DependencyTypeURL(scope)},
		{name: "any-all", flags: []string{"--type=any"}},
		{name: "compat-family-alias", flags: []string{"--conforms-to=" + task}, family: task},
		{name: "external-exact-uri", flags: []string{"--exact-type=https://external.test/task"}, exact: "https://external.test/task"},
		{name: "issue-family-and-task-exact", flags: []string{"--type=issue", "--exact-type=task"}, exact: task, family: graphstore.IssueTypeURL(scope)},
		{name: "family-alias-and-task-exact", flags: []string{"--conforms-to=" + graphstore.IssueTypeURL(scope), "--exact-type=task"}, exact: task, family: graphstore.IssueTypeURL(scope)},
		{name: "dependency-family-and-blocks-exact", links: true, flags: []string{"--type=dependency", "--exact-type=blocks"}, exact: blocks, family: graphstore.DependencyTypeURL(scope)},
		{name: "duplicate-family-bindings", flags: []string{"--type=issue", "--conforms-to=" + graphstore.IssueTypeURL(scope)}, refuses: true},
		{name: "empty", flags: []string{"--type="}, refuses: true},
		{name: "exact-wildcard-refused", flags: []string{"--exact-type=any"}, refuses: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			for _, flag := range []string{"type", "exact-type", "conforms-to"} {
				cmd.Flags().String(flag, "", "")
			}
			if err := cmd.ParseFlags(tc.flags); err != nil {
				t.Fatal(err)
			}
			exact, family, err := graphPreviewCollectionTypesForLayout(cmd, scope, tc.links, graphstore.SchemaVersion)
			if (err != nil) != tc.refuses || exact != tc.exact || family != tc.family {
				t.Fatalf("exact=%q family=%q err=%v; want %q %q refusal=%t", exact, family, err, tc.exact, tc.family, tc.refuses)
			}
		})
	}
}

func TestTypesCollectionDefaultIncludesAllOrdinaryBeads(t *testing.T) {
	const scope = "https://example.test/ordinary/"
	for _, tc := range []struct {
		name          string
		layout        int
		links         bool
		args          []string
		exact, family string
	}{
		{name: "legacy-beads", layout: graphstore.SchemaVersion, family: graphstore.IssueTypeURL(scope)},
		{name: "types-beads", layout: graphstore.TypesFreshSchemaVersion},
		{name: "types-links", layout: graphstore.TypesFreshSchemaVersion, links: true},
		{name: "explicit-native-family", layout: graphstore.TypesFreshSchemaVersion, args: []string{"--type=issue"}, family: graphstore.IssueTypeURL(scope)},
		{name: "explicit-definition-family", layout: graphstore.TypesFreshSchemaVersion, args: []string{"--type=bead-type"}, family: scope + "bead-type"},
		{name: "exact-whole-definition", layout: graphstore.TypesFreshSchemaVersion, args: []string{"--exact-type=types/bead-type"}, exact: scope + "types/bead-type"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			for _, flag := range []string{"type", "exact-type", "conforms-to"} {
				cmd.Flags().String(flag, "", "")
			}
			if err := cmd.ParseFlags(tc.args); err != nil {
				t.Fatal(err)
			}
			exact, family, err := graphPreviewCollectionTypesForLayout(cmd, scope, tc.links, tc.layout)
			if err != nil || exact != tc.exact || family != tc.family {
				t.Fatalf("selection %q/%q: %v; want %q/%q", exact, family, err, tc.exact, tc.family)
			}
		})
	}
}

func capturedCollectionFixture(t *testing.T) graphstore.ResourceSnapshot {
	t.Helper()
	const scope = "https://example.test/captured/"
	base := resourceget.TypeReference{URI: "https://types.example/type", Revision: "meaning2"}
	child := resourceget.TypeReference{URI: "https://types.example/bead-type", Revision: "meaning2"}
	oldChild := resourceget.TypeReference{URI: child.URI, Revision: "meaning1"}
	other := resourceget.TypeReference{URI: scope + "task", Revision: "task1"}
	snapshot := graphstore.ResourceSnapshot{WriterToken: "captured-token", Types: []graphstore.ResourceTypeFact{
		{Reference: base, Definition: resourceget.TypeReference{URI: scope + "type", Revision: "local1"}, Current: true, Describes: graph.KindBead, Ancestors: []resourceget.TypeReference{base}},
		{Reference: child, Definition: resourceget.TypeReference{URI: scope + "bead-type", Revision: "local2"}, Current: true, Describes: graph.KindBead, Ancestors: []resourceget.TypeReference{child, base}},
		{Reference: oldChild, Definition: resourceget.TypeReference{URI: scope + "bead-type", Revision: "local-old"}, Describes: graph.KindBead, Ancestors: []resourceget.TypeReference{oldChild, base}},
		{Reference: other, Definition: other, Current: true, Describes: graph.KindBead, Ancestors: []resourceget.TypeReference{other}},
	}}
	for i, ref := range []resourceget.TypeReference{child, oldChild, other} {
		raw, err := json.Marshal(map[string]any{"id": scope + []string{"current", "retained", "ordinary"}[i], "revision": "record1", "type": ref, "properties": map[string]any{"title": "exact stored body"}})
		if err != nil {
			t.Fatal(err)
		}
		doc, err := resourceget.NewResourceDocument(scope, graph.KindBead, raw)
		if err != nil {
			t.Fatal(err)
		}
		snapshot.Documents = append(snapshot.Documents, doc)
	}
	return snapshot
}

func TestCapturedCollectionResolvesDefinitionOnSameFacts(t *testing.T) {
	const scope = "https://example.test/captured/"
	for _, tc := range []struct {
		name    string
		request graphstore.BeadListRequest
		ids     []string
		more    bool
	}{
		{name: "local-type-family", request: graphstore.BeadListRequest{ConformsTo: scope + "type"}, ids: []string{"current", "retained"}},
		{name: "local-exact-family-intersection", request: graphstore.BeadListRequest{TypeURL: scope + "bead-type", ConformsTo: scope + "type"}, ids: []string{"current", "retained"}},
		{name: "literal-semantic-uris", request: graphstore.BeadListRequest{TypeURL: "https://types.example/bead-type", ConformsTo: "https://types.example/type"}, ids: []string{"current", "retained"}},
		{name: "whole-identity-empty-intersection", request: graphstore.BeadListRequest{TypeURL: scope + "task", ConformsTo: scope + "type"}, ids: []string{}},
		{name: "limit-preserved", request: graphstore.BeadListRequest{ConformsTo: scope + "type", Limit: 1, MaxRows: 2, All: true}, ids: []string{"current"}, more: true},
		{name: "unfiltered", request: graphstore.BeadListRequest{All: true}, ids: []string{"current", "retained", "ordinary"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			snapshot := capturedCollectionFixture(t)
			before, err := json.Marshal(snapshot)
			if err != nil {
				t.Fatal(err)
			}
			request := tc.request
			page, err := graphPreviewSelectCapturedBeads(snapshot, request)
			if err != nil || page.HasMore != tc.more || len(page.Items) != len(tc.ids) {
				t.Fatalf("selection=%+v err=%v", page, err)
			}
			for i, id := range tc.ids {
				if page.Items[i].ID() != scope+id || !bytes.Equal(page.Items[i].NativeJSON(), snapshot.Documents[i].NativeJSON()) {
					t.Fatalf("exact stored document changed: %s", page.Items[i].NativeJSON())
				}
			}
			after, err := json.Marshal(snapshot)
			if err != nil || !bytes.Equal(before, after) || request != tc.request {
				t.Fatal("selector request/captured records or facts mutated")
			}
		})
	}
}

func TestCapturedCollectionRefusesUnresolvedAndAmbiguousSelectors(t *testing.T) {
	const scope = "https://example.test/captured/"
	for _, tc := range []struct {
		name      string
		request   graphstore.BeadListRequest
		ambiguous bool
		want      error
	}{
		{name: "unknown-exact", request: graphstore.BeadListRequest{TypeURL: scope + "absent"}, want: graphstore.ErrCapabilityUnavailable},
		{name: "unknown-family", request: graphstore.BeadListRequest{ConformsTo: scope + "absent"}, want: graphstore.ErrCapabilityUnavailable},
		{name: "bare-not-guessed", request: graphstore.BeadListRequest{ConformsTo: "type"}, want: graphstore.ErrCapabilityUnavailable},
		{name: "ambiguous-current-definition", request: graphstore.BeadListRequest{ConformsTo: scope + "type"}, ambiguous: true, want: graphstore.ErrInvalidStore},
		{name: "maxrows-before-paging", request: graphstore.BeadListRequest{ConformsTo: scope + "type", MaxRows: 1, Limit: 1}, want: graphstore.ErrLimitExceeded},
	} {
		t.Run(tc.name, func(t *testing.T) {
			snapshot := capturedCollectionFixture(t)
			if tc.ambiguous {
				fact := snapshot.Types[0]
				fact.Reference = resourceget.TypeReference{URI: "https://types.example/other", Revision: "other1"}
				snapshot.Types = append(snapshot.Types, fact)
			}
			before, err := json.Marshal(snapshot)
			if err != nil {
				t.Fatal(err)
			}
			request := tc.request
			page, err := graphPreviewSelectCapturedBeads(snapshot, request)
			if !errors.Is(err, tc.want) || !reflect.ValueOf(page).IsZero() {
				t.Fatalf("refusal exposed page=%+v err=%v", page, err)
			}
			after, err := json.Marshal(snapshot)
			if err != nil || !bytes.Equal(before, after) || request != tc.request {
				t.Fatal("refusal mutated request/capture")
			}
		})
	}
}
