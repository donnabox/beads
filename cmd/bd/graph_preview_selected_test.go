package main

import (
	"bytes"
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
	if !implicit.ImplicitDependency {
		t.Fatal("omitted selector lost implicit-default evidence")
	}
	implicit.ImplicitDependency = false // population is equal; diagnostic intent differs
	for _, flags := range [][]string{{"--link-type=Dependency"}, {"--link-type=dependency"}, {"--link-type=types/dependency"}, {"--link-type=Dependency", "--link-type=dependency"}} {
		explicit, err := graphPreviewGenericInput(genericPreviewCommand(t, flags...), []string{"a"}, genericTestScope)
		if err != nil || !reflect.DeepEqual(implicit, explicit) {
			t.Fatalf("omitted selector differs from explicit: %+v %+v %v", implicit, explicit, err)
		}
	}
	for _, flags := range [][]string{{"--all"}, {"--link-type="}, {"--exact-link-type="}} {
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

func TestGraphPreviewSelectedExactAndWildcardInput(t *testing.T) {
	t.Setenv(maxRowsEnvVar, "")
	for _, tc := range []struct {
		name          string
		flags         []string
		family, exact []string
		all           bool
	}{
		{"exact", []string{"--exact-link-type=Requires"}, nil, []string{genericTestScope + "types/requires"}, false},
		{"mixed-deduplicated", []string{"--link-type=requires", "--exact-link-type=Defines", "--exact-link-type=defines", "--link-type=References"}, []string{genericTestScope + "types/references", genericTestScope + "types/requires"}, []string{genericTestScope + "types/defines"}, false},
		{"wildcard", []string{"--link-type=any"}, nil, nil, true},
		{"wildcard-union", []string{"--link-type=any", "--link-type=requires", "--exact-link-type=defines"}, []string{genericTestScope + "types/requires"}, []string{genericTestScope + "types/defines"}, true},
		{"identity-escape", []string{"--exact-link-type=" + genericTestScope + "types/any"}, nil, []string{genericTestScope + "types/any"}, false},
		{"family-identity-escape", []string{"--link-type=" + genericTestScope + "types/any"}, []string{genericTestScope + "types/any"}, nil, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in, err := graphPreviewGenericInput(genericPreviewCommand(t, tc.flags...), []string{"a"}, genericTestScope)
			if err != nil || !reflect.DeepEqual(in.LinkTypes, tc.family) || !reflect.DeepEqual(in.ExactLinkTypes, tc.exact) || in.AllLinkTypes != tc.all || in.ImplicitDependency {
				t.Fatalf("selector admission: %+v %v", in, err)
			}
		})
	}
	if graphCmd.Flags().Lookup("all-link-types") != nil {
		t.Fatal("removed preview-only wildcard flag remains registered")
	}
}

func TestGraphPreviewSelectedExactFamilyUnion(t *testing.T) {
	snapshot := genericTestSnapshot(t, genericTestEdge{"ab", "a", "b"}, genericTestEdge{"bc", "b", "c"}, genericTestEdge{"cd", "c", "d"})
	snapshot.Types = selectedTypeTestDescriptors(t, genericTestScope)
	for i, raw := range snapshot.Records {
		if link, ok := raw.(graphstore.LinkRecord); ok {
			link.Type = genericTestScope + "types/requires"
			if strings.HasSuffix(link.ID, "/bc") {
				link.Type = genericTestScope + "types/defines"
			}
			if strings.HasSuffix(link.ID, "/cd") {
				link.Type = genericTestScope + "types/references"
			}
			link.SourceRevision, link.TargetRevision = "source-pin", "target-pin"
			snapshot.Records[i] = link
		}
	}
	in := genericTestInput("a", "out", 10)
	in.ExactLinkTypes = []string{genericTestScope + "types/requires"}
	got, err := projectGraphGeneric(snapshot, genericTestScope, in)
	if err != nil {
		t.Fatal(err)
	}
	assertGenericSelection(t, got, []string{"a", "b"}, []string{"ab"}, []string{})
	in.ExactLinkTypes = append(in.ExactLinkTypes, genericTestScope+"types/defines")
	got, err = projectGraphGeneric(snapshot, genericTestScope, in)
	if err != nil {
		t.Fatal(err)
	}
	assertGenericSelection(t, got, []string{"a", "b", "c"}, []string{"ab", "bc"}, []string{})
	in.LinkTypes = []string{genericTestScope + "types/references"}
	got, err = projectGraphGeneric(snapshot, genericTestScope, in)
	if err != nil {
		t.Fatal(err)
	}
	assertGenericSelection(t, got, []string{"a", "b", "c", "d"}, []string{"ab", "bc", "cd"}, []string{})
	for _, link := range got.Links {
		if link.SourceRevision != "source-pin" || link.TargetRevision != "target-pin" {
			t.Fatal("selector union changed endpoint pins")
		}
	}
}

func TestGraphPreviewSelectedOmittedIncidence(t *testing.T) {
	snapshot := genericTestSnapshot(t, genericTestEdge{"ab", "a", "b"}, genericTestEdge{"bc", "b", "c"}, genericTestEdge{"dd", "d", "d"})
	snapshot.Types = selectedTypeTestDescriptors(t, genericTestScope)
	for i, raw := range snapshot.Records {
		if link, ok := raw.(graphstore.LinkRecord); ok {
			link.Type = graphstore.DependencyTypeURL(genericTestScope)
			if !strings.HasSuffix(link.ID, "/ab") {
				link.Type = genericTestScope + "types/references"
			}
			snapshot.Records[i] = link
		}
	}
	in := genericTestInput("a", "out", 0)
	in.LinkTypes = []string{graphstore.DependencyTypeURL(genericTestScope)}
	in.ImplicitDependency = true
	got, err := projectGraphGeneric(snapshot, genericTestScope, in)
	if err != nil || len(got.OmittedLinkTypes) != 0 {
		t.Fatalf("unreached incidence disclosed: %+v %v", got, err)
	}
	in.Depth = 1
	got, err = projectGraphGeneric(snapshot, genericTestScope, in)
	if err != nil || !reflect.DeepEqual(got.OmittedLinkTypes, []string{genericTestScope + "types/references"}) {
		t.Fatalf("reached incidence missing: %+v %v", got, err)
	}
	assertGenericSelection(t, got, []string{"a", "b"}, []string{"ab"}, []string{})
	encoded, err := json.Marshal(got)
	if err != nil || strings.Contains(string(encoded), "omitted") {
		t.Fatalf("private notice evidence leaked to JSON: %s %v", encoded, err)
	}
	in.ImplicitDependency = false
	got, err = projectGraphGeneric(snapshot, genericTestScope, in)
	if err != nil || len(got.OmittedLinkTypes) != 0 {
		t.Fatalf("explicit selection disclosed default notice: %+v %v", got, err)
	}
	in.ImplicitDependency = true
	in.MaxNodes = 1
	got, err = projectGraphGeneric(snapshot, genericTestScope, in)
	if err == nil || !reflect.DeepEqual(got, graphGenericResult{}) {
		t.Fatalf("refusal exposed partial notice/result: %+v %v", got, err)
	}
}

func TestGraphPreviewSelectedOpenIssuePopulationRetainsOmittedEvidence(t *testing.T) {
	snapshot := graphstore.Snapshot{Types: selectedTypeTestDescriptors(t, genericTestScope)}
	for _, status := range []types.Status{types.StatusOpen, types.StatusClosed} {
		id := string(status)
		snapshot.Records = append(snapshot.Records, graphstore.IssueRecord{ID: genericTestScope + "beads/" + id, Type: graphstore.IssueTypeURL(genericTestScope), Version: "v1", Properties: &types.Issue{Title: id, Status: status}})
	}
	snapshot.Records = append(snapshot.Records,
		graphstore.Record{ID: genericTestScope + "beads/m", Type: graphstore.MemoryTypeURL(genericTestScope), Version: "v1"},
		graphstore.LinkRecord{ID: genericTestScope + "links/dep", Type: graphstore.DependencyTypeURL(genericTestScope), Version: "v1", Source: genericTestScope + "beads/open", Target: genericTestScope + "beads/closed"},
		graphstore.LinkRecord{ID: genericTestScope + "links/ref", Type: genericTestScope + "types/references", Version: "v1", Source: genericTestScope + "beads/open", Target: genericTestScope + "beads/m"},
	)
	in := genericTestInput("", "both", 1000)
	in.All, in.OpenIssuesOnly, in.ImplicitDependency = true, true, true
	in.LinkTypes = []string{graphstore.DependencyTypeURL(genericTestScope)}
	got, err := projectGraphGeneric(snapshot, genericTestScope, in)
	if err != nil {
		t.Fatal(err)
	}
	assertGenericSelection(t, got, []string{"open"}, []string{}, []string{})
	if !reflect.DeepEqual(got.OmittedLinkTypes, []string{genericTestScope + "types/references"}) {
		t.Fatalf("open-Issue filtering erased authorized incidence: %+v", got)
	}
}

func TestGraphPreviewSelectedAllDependencyPopulationIsSelectorSpecific(t *testing.T) {
	snapshot := graphstore.Snapshot{Types: selectedTypeTestDescriptors(t, genericTestScope)}
	for _, id := range []string{"open", "closed", "isolated"} {
		status := types.StatusOpen
		if id == "closed" {
			status = types.StatusClosed
		}
		snapshot.Records = append(snapshot.Records, graphstore.IssueRecord{ID: genericTestScope + "beads/" + id, Type: graphstore.IssueTypeURL(genericTestScope), Version: "v1", Properties: &types.Issue{Title: id, Status: status}})
	}
	snapshot.Records = append(snapshot.Records,
		graphstore.Record{ID: genericTestScope + "beads/m", Type: graphstore.MemoryTypeURL(genericTestScope), Version: "v1"},
		graphstore.LinkRecord{ID: genericTestScope + "links/dep", Type: graphstore.DependencyTypeURL(genericTestScope), Version: "v1", Source: genericTestScope + "beads/open", Target: genericTestScope + "beads/closed"},
		graphstore.LinkRecord{ID: genericTestScope + "links/ref", Type: genericTestScope + "types/references", Version: "v1", Source: genericTestScope + "beads/open", Target: genericTestScope + "beads/m"},
	)
	family := genericTestInput("", "both", 1000)
	family.All, family.OpenIssuesOnly = true, true
	family.LinkTypes = []string{graphstore.DependencyTypeURL(genericTestScope)}
	got, err := projectGraphGeneric(snapshot, genericTestScope, family)
	if err != nil {
		t.Fatal(err)
	}
	// Sole family Dependency is the familiar open-Issue population, including
	// isolates. Exact Dependency instead uses the general matching endpoints.
	assertGenericSelection(t, got, []string{"isolated", "open"}, []string{}, []string{})
	exact := family
	exact.OpenIssuesOnly = false
	exact.LinkTypes = nil
	exact.ExactLinkTypes = []string{graphstore.DependencyTypeURL(genericTestScope)}
	got, err = projectGraphGeneric(snapshot, genericTestScope, exact)
	if err != nil {
		t.Fatal(err)
	}
	assertGenericSelection(t, got, []string{"closed", "open"}, []string{"dep"}, []string{})
	wildcard := exact
	wildcard.ExactLinkTypes = nil
	wildcard.AllLinkTypes = true
	got, err = projectGraphGeneric(snapshot, genericTestScope, wildcard)
	if err != nil {
		t.Fatal(err)
	}
	assertGenericSelection(t, got, []string{"closed", "m", "open"}, []string{"dep", "ref"}, []string{})
}

func TestGraphPreviewSelectedExportLabelsAndDepthBoundary(t *testing.T) {
	for _, format := range []string{"box", "dot", "html"} {
		for _, truncated := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/truncated=%t", format, truncated), func(t *testing.T) {
				r := graphTreeTestResult("a", "out", []string{"a", "b"}, genericTestEdge{"ab", "a", "b"})
				r.Depth = 3
				if truncated {
					r.Complete, r.Frontier = false, []string{r.Nodes[1].ID}
				}
				var out bytes.Buffer
				writer := &graphExportWriter{out: &out}
				switch format {
				case "box":
					renderGraphSelectedBox(writer, r)
				case "dot":
					renderGraphSelectedDOT(writer, r)
				case "html":
					if err := renderGraphSelectedHTML(writer, r); err != nil {
						t.Fatal(err)
					}
				}
				if err := writer.wrapError("graph"); err != nil {
					t.Fatal(err)
				}
				text := out.String()
				if !strings.Contains(text, "beads/a [types/memory]") {
					t.Fatalf("%s lost node identity/Type label: %s", format, text)
				}
				if format == "html" {
					// HTML keeps exact Link Types in data; the tooltip explicitly
					// brackets that Type rather than changing its wire identity.
					if !strings.Contains(text, `"type":"`+genericTestScope+`types/references"`) || !strings.Contains(text, `"\n[" + d.type + "]\nsource revision: "`) {
						t.Fatal("HTML lost typed Link data/tooltip", text)
					}
				} else if !strings.Contains(text, "[types/references]") {
					t.Fatalf("%s lost Link Type label: %s", format, text)
				}
				boundary := "Depth limit 3 reached"
				if format == "html" {
					boundary = "depth limit 3 reached"
				}
				if strings.Contains(text, boundary) != truncated {
					t.Fatalf("%s depth warning did not match checked frontier: %s", format, text)
				}
				for _, bookkeeping := range []string{"complete:", "Complete:", "complete true", "complete false", "frontier:", "frontier 0", "frontier 1", "direction out"} {
					if strings.Contains(text, bookkeeping) {
						t.Fatalf("%s emitted routine bookkeeping %q", format, bookkeeping)
					}
				}
			})
		}
	}
	var empty bytes.Buffer
	renderGraphSelectedBox(&graphExportWriter{out: &empty}, graphGenericResult{})
	if empty.String() != "No matching Beads.\n" {
		t.Fatalf("empty Box: %q", empty.String())
	}
}

func TestGraphPreviewSelectedBoxEscapesStatus(t *testing.T) {
	r := graphTreeTestResult("a", "out", []string{"a"})
	r.Nodes[0].Status = "open\n└── forged\x1b[2J"
	before := r.Nodes[0].Status
	var out bytes.Buffer
	renderGraphSelectedBox(&graphExportWriter{out: &out}, r)
	if strings.Contains(out.String(), "\x1b") || strings.Contains(out.String(), "\n└── forged") || !strings.Contains(out.String(), `open\n└── forged\x1b[2J`) || r.Nodes[0].Status != before {
		t.Fatalf("Box injected controls or changed captured Status: %q", out.String())
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
				want = "(" + id + ")"
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
	in, err := graphPreviewGenericInput(genericPreviewCommand(t), []string{"beads/n0"}, genericTestScope)
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
