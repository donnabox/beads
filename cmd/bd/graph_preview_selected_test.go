package main

import (
	"encoding/json"
	"fmt"
	"github.com/steveyegge/beads/internal/types"
	"reflect"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func TestGraphPreviewSelectedInputEquivalence(t *testing.T) {
	t.Setenv(maxRowsEnvVar, "")
	implicit, err := graphPreviewGenericInput(genericPreviewCommand(t), []string{"a"}, genericTestScope)
	if err != nil {
		t.Fatal(err)
	}
	for _, flags := range [][]string{{"--link-type=Dependency"}, {"--link-type=dependency"}, {"--link-type=types/dependency"}, {"--link-type=Dependency", "--link-type=dependency"}} {
		explicit, err := graphPreviewGenericInput(genericPreviewCommand(t, flags...), []string{"a"}, genericTestScope)
		if err != nil || !reflect.DeepEqual(implicit, explicit) {
			t.Fatalf("omitted selector differs from explicit: %+v %+v %v", implicit, explicit, err)
		}
	}
	for _, flags := range [][]string{{"--all-link-types", "--link-type=references"}, {"--all"}, {"--link-type="}} {
		if _, err := graphPreviewGenericInput(genericPreviewCommand(t, flags...), []string{"a"}, genericTestScope); err == nil {
			t.Fatalf("ambiguous input accepted: %v", flags)
		}
	}
	for _, flag := range []string{"--direction=both", "--depth=1"} {
		if _, err := graphPreviewGenericInput(genericPreviewCommand(t, "--all", flag), nil, genericTestScope); err == nil {
			t.Fatalf("all silently ignored %s", flag)
		}
	}
	all, err := graphPreviewGenericInput(genericPreviewCommand(t, "--all", "--link-type=references"), nil, genericTestScope)
	if err != nil || !all.All || all.Root != "" {
		t.Fatalf("all admission: %+v %v", all, err)
	}
}

func TestGraphPreviewSelectedUnionInducedAndNoHiddenBridge(t *testing.T) {
	snapshot := genericTestSnapshot(t, genericTestEdge{"ab", "a", "b"}, genericTestEdge{"ac", "a", "c"}, genericTestEdge{"bc", "b", "c"}, genericTestEdge{"bc2", "b", "c"}, genericTestEdge{"cd", "c", "d"}, genericTestEdge{"ba", "b", "a"})
	snapshot.Types = selectedTypeTestDescriptors(t, genericTestScope)
	for i, raw := range snapshot.Records {
		if l, ok := raw.(graphstore.LinkRecord); ok {
			l.Type = genericTestScope + "types/requires"
			if strings.HasSuffix(l.ID, "/bc2") {
				l.Type = genericTestScope + "types/defines"
			}
			if strings.HasSuffix(l.ID, "/cd") {
				l.Type = genericTestScope + "types/references"
			}
			l.TargetRevision = "pinned-target"
			snapshot.Records[i] = l
		}
	}
	in := genericTestInput("a", "out", 1)
	in.LinkTypes = []string{genericTestScope + "types/requires"}
	got, err := projectGraphGeneric(snapshot, genericTestScope, in)
	if err != nil {
		t.Fatal(err)
	}
	assertGenericSelection(t, got, []string{"a", "b", "c"}, []string{"ab", "ac", "ba", "bc", "bc2"}, []string{})
	for _, l := range got.Links {
		if l.TargetRevision != "pinned-target" {
			t.Fatal("authored pin erased")
		}
	}
	in.Depth = 10
	got, err = projectGraphGeneric(snapshot, genericTestScope, in)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Nodes) != 3 {
		t.Fatal("excluded References became hidden bridge")
	}
	in.LinkTypes = append(in.LinkTypes, genericTestScope+"types/references")
	got, err = projectGraphGeneric(snapshot, genericTestScope, in)
	if err != nil || len(got.Nodes) != 4 || len(got.Links) != 6 {
		t.Fatalf("union: %+v %v", got, err)
	}
	in.LinkTypes = []string{genericTestScope + "types/references"}
	got, err = projectGraphGeneric(snapshot, genericTestScope, in)
	if err != nil || len(got.Links) != 6 {
		t.Fatalf("refinements not included: %+v %v", got, err)
	}
	in.All = true
	in.Root = ""
	in.LinkTypes = []string{genericTestScope + "types/requires"}
	got, err = projectGraphGeneric(snapshot, genericTestScope, in)
	if err != nil || len(got.Nodes) != 3 {
		t.Fatalf("nondependency all admitted unrelated isolated bead: %+v %v", got, err)
	}
	in.All = false
	in.Root = genericTestScope + "beads/d"
	in.Direction = "out"
	got, err = projectGraphGeneric(snapshot, genericTestScope, in)
	if err != nil || len(got.Nodes) != 1 || len(got.Links) != 0 {
		t.Fatalf("isolated explicit root lost: %+v %v", got, err)
	}
}

func TestGraphPreviewSelectedCanonicalAndFormats(t *testing.T) {
	snapshot := genericTestSnapshot(t, genericTestEdge{"ab", "a", "b"}, genericTestEdge{"ba", "b", "a"}, genericTestEdge{"ab2", "a", "b"})
	got, err := projectGraphGeneric(snapshot, genericTestScope, genericTestInput("a", "out", 1))
	if err != nil {
		t.Fatal(err)
	}
	oldJSON, oldQuiet := jsonOutput, quietFlag
	t.Cleanup(func() { jsonOutput, quietFlag = oldJSON, oldQuiet })
	jsonOutput = true
	quietFlag = true
	output, err := renderGraphSelected(genericPreviewCommand(t), got, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(output, `"version"`) || !strings.Contains(output, `"revision"`) || !strings.Contains(output, "SECRET_BODY") {
		t.Fatalf("complete record projection bypassed: %s", output)
	}
	var envelope struct {
		Result struct {
			Beads     []json.RawMessage
			Links     []graphstore.LinkRecord
			Traversal graphSelectedTraversal
		}
	}
	if err := json.Unmarshal([]byte(output), &envelope); err != nil || len(envelope.Result.Beads) != 2 || len(envelope.Result.Links) != 3 {
		t.Fatalf("canonical graph lost records: %v %s", err, output)
	}
	jsonOutput = false
	quietFlag = false
	for _, format := range []string{"", "--compact", "--box", "--dot", "--html"} {
		flags := []string{}
		if format != "" {
			flags = append(flags, format)
		}
		text, err := renderGraphSelected(genericPreviewCommand(t, flags...), got, snapshot)
		if err != nil {
			t.Fatal(err)
		}
		for _, id := range []string{"links/ab", "links/ab2", "links/ba"} {
			want := id
			if format == "" || format == "--compact" {
				want = "(" + strings.TrimPrefix(id, "links/") + ")"
			}
			if !strings.Contains(text, want) {
				t.Fatalf("%s lost actual link %s: %s", format, id, text)
			}
		}
		if format == "--html" {
			for _, fragment := range []string{`.link { stroke: #888;`, `join("path")`, `link.append("title")`, `function graphLinkPath`, `<h3>Typed Bead graph</h3>`} {
				if !strings.Contains(text, fragment) {
					t.Fatalf("HTML missing visible typed-link contract %q", fragment)
				}
			}
			if strings.Contains(text, `<h3>Status</h3>`) {
				t.Fatal("native workflow legend leaked into generic HTML")
			}
		}
		if strings.Contains(text, "(ready)") || strings.Contains(text, "Priority: P") {
			t.Fatal("generic graph invented workflow semantics")
		}
	}
}

// Retain the predecessor journey's graph-selection assertions while decoding
// the new complete canonical CLI graph rather than its old summary-only JSON.
func graphSelectedTestSummary(t *testing.T, output string) graphGenericResult {
	t.Helper()
	var envelope struct {
		Result struct {
			Scope string
			Beads []struct {
				ID, Type, Revision string
				Properties         struct{ Title string }
			}
			Links     []graphstore.LinkRecord
			Traversal graphSelectedTraversal
		}
	}
	if err := json.Unmarshal([]byte(output), &envelope); err != nil {
		t.Fatal(err)
	}
	r := envelope.Result
	result := graphGenericResult{Scope: r.Scope, Root: r.Traversal.Root, Direction: r.Traversal.Direction, Depth: r.Traversal.Depth, Complete: r.Traversal.Complete, Frontier: r.Traversal.Frontier, Nodes: []graphGenericNode{}, Links: []graphGenericLink{}}
	for _, n := range r.Beads {
		result.Nodes = append(result.Nodes, graphGenericNode{ID: n.ID, Type: n.Type, Title: n.Properties.Title, Version: n.Revision})
	}
	for _, l := range r.Links {
		result.Links = append(result.Links, graphGenericLink{ID: l.ID, Type: l.Type, Source: l.Source, Target: l.Target, Version: l.Revision, SourceRevision: l.SourceRevision, TargetRevision: l.TargetRevision})
	}
	return result
}

func TestGraphPreviewSelectedInducedLinkLimitAndMissingRecords(t *testing.T) {
	snapshot := genericTestSnapshot(t, genericTestEdge{"ab", "a", "b"}, genericTestEdge{"ac", "a", "c"}, genericTestEdge{"bc", "b", "c"})
	in := genericTestInput("a", "out", 1)
	in.MaxLinks = 2
	if result, err := projectGraphGeneric(snapshot, genericTestScope, in); err == nil || !reflect.DeepEqual(result, graphGenericResult{}) {
		t.Fatalf("induced third cross-link bypassed bound: %+v %v", result, err)
	}
	in.MaxLinks = 3
	result, err := projectGraphGeneric(snapshot, genericTestScope, in)
	if err != nil {
		t.Fatal(err)
	}
	for _, missing := range []string{"beads/a", "links/ab"} {
		broken := snapshot
		broken.Records = nil
		for _, raw := range snapshot.Records {
			switch r := raw.(type) {
			case graphstore.Record:
				if r.ID == genericTestScope+missing {
					continue
				}
			case graphstore.LinkRecord:
				if r.ID == genericTestScope+missing {
					continue
				}
			}
			broken.Records = append(broken.Records, raw)
		}
		if out, err := graphSelectedRecords(result, broken); err == nil || !reflect.DeepEqual(out, graphSelectedOutput{}) {
			t.Fatalf("missing canonical record leaked/panicked: %+v %v", out, err)
		}
	}
}

func TestGraphPreviewSelectedOrdinaryDefaultKeepsLargeDependencyComponent(t *testing.T) {
	t.Setenv(maxRowsEnvVar, "")
	in, err := graphPreviewGenericInput(genericPreviewCommand(t), []string{"n0"}, genericTestScope)
	if err != nil {
		t.Fatal(err)
	}
	snapshot := graphstore.Snapshot{Types: selectedTypeTestDescriptors(t, genericTestScope)}
	for i := 0; i < 151; i++ {
		id := fmt.Sprintf("n%d", i)
		snapshot.Records = append(snapshot.Records, graphstore.IssueRecord{ID: genericTestScope + "beads/" + id, Type: graphstore.IssueTypeURL(genericTestScope), Version: "v-" + id, Revision: "v-" + id, Properties: &types.Issue{ID: id, Title: id, Status: types.StatusOpen}})
		if i > 0 {
			snapshot.Records = append(snapshot.Records, graphstore.LinkRecord{ID: fmt.Sprintf("%slinks/e%d", genericTestScope, i), Type: graphstore.DependencyTypeURL(genericTestScope), Source: fmt.Sprintf("%sbeads/n%d", genericTestScope, i-1), Target: genericTestScope + "beads/" + id, Version: "v-" + id, Revision: "v-" + id})
		}
	}
	result, err := projectGraphGeneric(snapshot, genericTestScope, in)
	if err != nil || len(result.Nodes) != 151 || len(result.Links) != 150 || !result.Complete {
		t.Fatalf("ordinary default shrank prior supported component: %+v %v", result, err)
	}
}
