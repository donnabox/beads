package main

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func graphTreeTestResult(root, direction string, names []string, edges ...genericTestEdge) graphGenericResult {
	r := graphGenericResult{Scope: genericTestScope, Direction: direction, Depth: 1000, Complete: true}
	if root != "" {
		r.Root = genericTestScope + "beads/" + root
	} else {
		r.All = true
	}
	for _, name := range names {
		r.Nodes = append(r.Nodes, graphGenericNode{ID: genericTestScope + "beads/" + name, Type: genericTestScope + "types/memory", Title: strings.ToUpper(name)})
	}
	for _, edge := range edges {
		r.Links = append(r.Links, graphGenericLink{ID: genericTestScope + "links/" + edge.id, Type: genericTestScope + "types/references", Source: genericTestScope + "beads/" + edge.source, Target: genericTestScope + "beads/" + edge.target})
	}
	return r
}

func graphTreeTestText(t *testing.T, r graphGenericResult) string {
	t.Helper()
	var output bytes.Buffer
	w := &graphExportWriter{out: &output}
	if err := renderGraphRelationshipTree(w, r, graphstore.PreviewCurrentReadByteLimit); err != nil {
		t.Fatal(err)
	}
	for _, link := range r.Links {
		id := strings.TrimPrefix(link.ID, r.Scope)
		if strings.Count(output.String(), "("+id+")") != 1 {
			t.Fatalf("actual selected Link %s must appear exactly once:\n%s", id, output.String())
		}
	}
	return output.String()
}

func TestGraphRelationshipTreeReferencesAndCycles(t *testing.T) {
	for _, tc := range []struct {
		name, root, direction string
		edges                 []genericTestEdge
		cycles, references    int
	}{
		{"parallel-is-not-cycle", "a", "both", []genericTestEdge{{"ab", "a", "b"}, {"ab2", "a", "b"}}, 0, 1},
		{"mixed-direction-dag-is-not-cycle", "a", "both", []genericTestEdge{{"ab", "a", "b"}, {"cb", "c", "b"}, {"ca", "c", "a"}}, 0, 1},
		{"outgoing-cycle", "a", "out", []genericTestEdge{{"ab", "a", "b"}, {"bc", "b", "c"}, {"ca", "c", "a"}}, 1, 1},
		{"both-cycle", "a", "both", []genericTestEdge{{"ab", "a", "b"}, {"bc", "b", "c"}, {"ca", "c", "a"}}, 1, 1},
		{"incoming-cycle", "a", "in", []genericTestEdge{{"ab", "a", "b"}, {"bc", "b", "c"}, {"ca", "c", "a"}}, 1, 1},
		{"diamond", "a", "out", []genericTestEdge{{"ab", "a", "b"}, {"ac", "a", "c"}, {"bd", "b", "d"}, {"cd", "c", "d"}}, 0, 1},
		{"diamond-both", "a", "both", []genericTestEdge{{"ab", "a", "b"}, {"ac", "a", "c"}, {"bd", "b", "d"}, {"cd", "c", "d"}}, 0, 1},
		{"self-cycle", "a", "both", []genericTestEdge{{"aa", "a", "a"}}, 1, 1},
		{"incoming-self-cycle", "a", "in", []genericTestEdge{{"aa", "a", "a"}}, 1, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Authoritative selection comes from the real projection, so the
			// tree cannot add the fixture's unrelated isolated Beads or Links.
			snapshot := genericTestSnapshot(t, tc.edges...)
			r, err := projectGraphGeneric(snapshot, genericTestScope, genericTestInput(tc.root, tc.direction, 10))
			if err != nil {
				t.Fatal(err)
			}
			text := graphTreeTestText(t, r)
			if strings.Count(text, "(cycle)") != tc.cycles || strings.Count(text, "(shown elsewhere)") != tc.references {
				t.Fatalf("wrong directed-cycle/shared reference:\n%s", text)
			}
			if len(strings.Split(text, "\n")) > len(r.Links)+len(r.Nodes)+5 {
				t.Fatalf("repeated expansion:\n%s", text)
			}
			if tc.direction == "in" {
				for _, edge := range tc.edges {
					if edge.source != edge.target && !strings.Contains(text, "(links/"+edge.id+") ← ") {
						t.Fatalf("incoming indentation reversed authored arrow for %s:\n%s", edge.id, text)
					}
				}
			}
		})
	}
}

func TestGraphRelationshipTreeLayoutAndDirections(t *testing.T) {
	r := graphTreeTestResult("a", "out", []string{"a", "b", "c", "d"}, genericTestEdge{"ab", "a", "b"}, genericTestEdge{"ac", "a", "c"}, genericTestEdge{"bd", "b", "d"}, genericTestEdge{"cd", "c", "d"})
	want := "beads/a [types/memory] \"A\"\n├── [types/references] (links/ab) → beads/b [types/memory] \"B\"\n│   └── [types/references] (links/bd) → beads/d [types/memory] \"D\"\n└── [types/references] (links/ac) → beads/c [types/memory] \"C\"\n    └── [types/references] (links/cd) → beads/d [types/memory] (shown elsewhere)\n\n4 Beads, 4 Links\n"
	if got := graphTreeTestText(t, r); got != want {
		t.Fatalf("relationship layout:\n%s\nwant:\n%s", got, want)
	}
	before := graphTreeTestText(t, r)
	for i, j := 0, len(r.Nodes)-1; i < j; i, j = i+1, j-1 {
		r.Nodes[i], r.Nodes[j] = r.Nodes[j], r.Nodes[i]
	}
	for i, j := 0, len(r.Links)-1; i < j; i, j = i+1, j-1 {
		r.Links[i], r.Links[j] = r.Links[j], r.Links[i]
	}
	if got := graphTreeTestText(t, r); got != before {
		t.Fatalf("input permutation changed output:\n%s", got)
	}
	r = graphTreeTestResult("c", "in", []string{"a", "b", "c", "d"}, genericTestEdge{"bc", "b", "c"}, genericTestEdge{"ab", "a", "b"}, genericTestEdge{"db", "d", "b"})
	for i := range r.Links {
		r.Links[i].Type = genericTestScope + "types/superseded-by"
	}
	text := graphTreeTestText(t, r)
	for _, fragment := range []string{"beads/c [types/memory]", "[types/superseded-by] (links/bc) ← beads/b", "[types/superseded-by] (links/ab) ← beads/a", "[types/superseded-by] (links/db) ← beads/d", "4 Beads, 3 Links"} {
		if !strings.Contains(text, fragment) {
			t.Fatalf("reverse ancestry lost %q:\n%s", fragment, text)
		}
	}
}

func TestGraphRelationshipTreeZeroHopFrontierAndForest(t *testing.T) {
	r := graphTreeTestResult("a", "out", []string{"a"})
	r.Depth, r.Complete, r.Frontier = 0, false, []string{r.Root}
	text := graphTreeTestText(t, r)
	if !strings.HasPrefix(text, "beads/a [types/memory] \"A\"\n") || !strings.Contains(text, "1 Beads, 0 Links") || !strings.Contains(text, "Depth limit 0 reached at beads/a; further relationships are outside this view.") || strings.Contains(text, "└──") {
		t.Fatalf("zero-hop root or honest boundary lost:\n%s", text)
	}
	r = graphTreeTestResult("", "out", []string{"a", "b", "c"}, genericTestEdge{"ab", "a", "b"})
	text = graphTreeTestText(t, r)
	if !strings.Contains(text, "\nbeads/c [types/memory] \"C\"\n") || !strings.Contains(text, "3 Beads, 1 Links") {
		t.Fatalf("selected isolate missing from --all forest:\n%s", text)
	}
	r = graphTreeTestResult("", "out", nil)
	if text := graphTreeTestText(t, r); text != "No matching Beads.\n" {
		t.Fatalf("empty selection:\n%s", text)
	}
}

func TestGraphRelationshipTreePreservesLiteralSegmentsAndForeignURLs(t *testing.T) {
	r := graphTreeTestResult("a", "out", []string{"a", "b"}, genericTestEdge{"ab", "a", "b"})
	foreign := "https://elsewhere.invalid/graph/beads/b"
	r.Nodes[1].ID = foreign
	r.Nodes[1].Type = "https://elsewhere.invalid/types/policy"
	r.Links[0].Target = foreign
	text := graphTreeTestText(t, r)
	for _, want := range []string{"beads/a [types/memory]", "[types/references] (links/ab)", foreign + " [https://elsewhere.invalid/types/policy]"} {
		if !strings.Contains(text, want) {
			t.Fatalf("literal identity changed; want %q:\n%s", want, text)
		}
	}
	for _, unwanted := range []string{genericTestScope, "Complete:", "frontier:", "direction:", "Depth limit"} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("complete view emitted %q:\n%s", unwanted, text)
		}
	}
}

func TestGraphRelationshipTreeEscapesAndPreservesData(t *testing.T) {
	r := graphTreeTestResult("a", "out", []string{"a", "b"}, genericTestEdge{"ab", "a", "b"})
	r.Nodes[0].Type = genericTestScope + "types/issue"
	r.Nodes[0].Status = "open"
	r.Nodes[1].Type = genericTestScope + "types/requirement"
	r.Nodes[1].Title = "Line\nTab\tESC\x1b\" 雪\u2028end"
	r.Links[0].Type = genericTestScope + "types/defines"
	r.Links[0].SourceRevision, r.Links[0].TargetRevision = "pin\n\x1b", "v2"
	before := r.Nodes[1].Title
	text := graphTreeTestText(t, r)
	for _, want := range []string{"beads/a [types/issue] (open)", "[types/defines] (links/ab) → beads/b [types/requirement]", `Line\nTab\tESC\x1b\" 雪\u2028end`, `source revision="pin\n\x1b" target revision="v2"`} {
		if !strings.Contains(text, want) {
			t.Fatalf("data lost or controls injected; need %q:\n%s", want, text)
		}
	}
	if strings.ContainsAny(text, "\x1b\t\u2028") || strings.Contains(text, genericTestScope) || strings.Contains(text, "[types/requirement] (open)") || r.Nodes[1].Title != before {
		t.Fatalf("Scope elision/escaping changed authoritative data or invented status:\n%s", text)
	}
}

func TestGraphRelationshipTreeRejectsInvalidSelectedGraph(t *testing.T) {
	for _, tc := range []struct {
		name string
		edit func(*graphGenericResult)
	}{
		{"missing-root", func(r *graphGenericResult) { r.Nodes = r.Nodes[1:] }},
		{"absent-root", func(r *graphGenericResult) { r.Root = "" }},
		{"missing-target", func(r *graphGenericResult) { r.Nodes = r.Nodes[:1] }},
		{"duplicate-node", func(r *graphGenericResult) { r.Nodes = append(r.Nodes, r.Nodes[0]) }},
		{"duplicate-link", func(r *graphGenericResult) { r.Links = append(r.Links, r.Links[0]) }},
		{"missing-type", func(r *graphGenericResult) { r.Nodes[0].Type = "" }},
		{"missing-frontier", func(r *graphGenericResult) { r.Frontier = []string{"absent"} }},
		{"invalid-direction", func(r *graphGenericResult) { r.Direction = "sideways" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := graphTreeTestResult("a", "out", []string{"a", "b"}, genericTestEdge{"ab", "a", "b"})
			tc.edit(&r)
			var output bytes.Buffer
			err := renderGraphRelationshipTree(&graphExportWriter{out: &output}, r, graphstore.PreviewCurrentReadByteLimit)
			if !errors.Is(err, graphstore.ErrInvalidStore) || output.Len() != 0 {
				t.Fatalf("invalid selection fabricated or leaked output: %v %s", err, output.String())
			}
		})
	}
}

func TestGraphRelationshipTreeUsesOnlySelectedLinks(t *testing.T) {
	snapshot := genericTestSnapshot(t, genericTestEdge{"selected", "a", "b"}, genericTestEdge{"excluded", "a", "c"})
	snapshot.Types = selectedTypeTestDescriptors(t, genericTestScope)
	link := snapshot.Records[len(snapshot.Records)-1].(graphstore.LinkRecord)
	link.Type = graphstore.DependencyTypeURL(genericTestScope)
	snapshot.Records[len(snapshot.Records)-1] = link
	in := genericTestInput("a", "out", 1)
	in.LinkTypes = []string{graphstore.RelatedTypeURL(genericTestScope)}
	r, err := projectGraphGeneric(snapshot, genericTestScope, in)
	if err != nil {
		t.Fatal(err)
	}
	text := graphTreeTestText(t, r)
	if strings.Contains(text, "excluded") || strings.Contains(text, "beads/c [types/memory]") || len(r.Links) != 1 {
		t.Fatalf("presenter crossed excluded Link:\n%s", text)
	}
	// Rendering mutates neither the selected graph nor the complete JSON path.
	copyOfNodes, copyOfLinks := append([]graphGenericNode(nil), r.Nodes...), append([]graphGenericLink(nil), r.Links...)
	_ = graphTreeTestText(t, r)
	if !reflect.DeepEqual(r.Nodes, copyOfNodes) || !reflect.DeepEqual(r.Links, copyOfLinks) {
		t.Fatal("presenter mutated the selected graph")
	}
}

func TestGraphRelationshipTreeLargeTitlesAndOutputBudget(t *testing.T) {
	r := graphTreeTestResult("a", "both", []string{"a", "b"})
	title := strings.Repeat("x", 1<<20)
	r.Nodes[1].Title = title
	for i := 0; i < 300; i++ {
		r.Links = append(r.Links, graphGenericLink{ID: genericTestScope + fmt.Sprintf("links/parallel-%03d", i), Type: genericTestScope + "types/references", Source: r.Nodes[0].ID, Target: r.Nodes[1].ID})
	}
	text := graphTreeTestText(t, r)
	if count := strings.Count(text, title); count != 1 {
		t.Fatalf("target title appeared %d times; want 1", count)
	}
	if len(text) > 2<<20 {
		t.Fatalf("parallel title output is %d bytes; want at most 2 MiB", len(text))
	}
	if count := strings.Count(text, "(shown elsewhere)"); count != 299 {
		t.Fatalf("shared references: got %d; want 299", count)
	}
	// A directly supplied summary can expand beyond the text byte bound when
	// controls are escaped. Enforce the output bound independently of input
	// acquisition. Refuse before growing
	// the private buffer beyond that bound and never return publishable text.
	r = graphTreeTestResult("a", "out", []string{"a", "b"}, genericTestEdge{"ab", "a", "b"})
	for i := range r.Nodes {
		r.Nodes[i].Title = strings.Repeat("\x1b", 3<<20)
	}
	var private bytes.Buffer
	err := renderGraphRelationshipTree(&graphExportWriter{out: &private}, r, graphstore.PreviewCurrentReadByteLimit)
	if !errors.Is(err, graphstore.ErrLimitExceeded) || private.Len() > graphstore.PreviewCurrentReadByteLimit {
		t.Fatalf("unbounded private output: %d bytes, %v", private.Len(), err)
	}
	oldJSON, oldQuiet := jsonOutput, quietFlag
	t.Cleanup(func() { jsonOutput, quietFlag = oldJSON, oldQuiet })
	jsonOutput, quietFlag = false, false
	output, err := renderGraphSelected(genericPreviewCommand(t), r, graphstore.Snapshot{})
	if !errors.Is(err, graphstore.ErrLimitExceeded) || output != "" {
		t.Fatalf("oversized graph leaked partial output: %d bytes, %v", len(output), err)
	}
}

func TestGraphRelationshipTreeCallerBudgetAndUnicodeOrder(t *testing.T) {
	r := graphTreeTestResult("a", "out", []string{"a", "b", "c"}, genericTestEdge{"ab", "a", "b"}, genericTestEdge{"ac", "a", "c"})
	r.Nodes[1].Title, r.Nodes[2].Title = "\ue000", "\U00010000"
	text := graphTreeTestText(t, r)
	if strings.Index(text, "(links/ac)") > strings.Index(text, "(links/ab)") {
		t.Fatalf("titles must follow the graph's UTF-16 code-unit order:\n%s", text)
	}
	var output bytes.Buffer
	w := &graphExportWriter{out: &output}
	err := renderGraphRelationshipTree(w, r, 32)
	if !errors.Is(err, graphstore.ErrLimitExceeded) || output.Len() > 32 || w.err != nil || strings.Contains(err.Error(), "writing") {
		t.Fatalf("caller budget must refuse independently of IO errors: %d bytes, writer %v, error %v", output.Len(), w.err, err)
	}
	output.Reset()
	broken := errors.New("output failure")
	w = &graphExportWriter{out: &output, err: broken}
	err = renderGraphRelationshipTree(w, r, graphstore.PreviewCurrentReadByteLimit)
	if !errors.Is(err, broken) || output.Len() != 0 {
		t.Fatalf("first writer failure lost: %v, %d bytes", err, output.Len())
	}
}

func TestGraphRelationshipTreeCanonicalExpansion(t *testing.T) {
	t.Run("both-parallel-keeps-first-endpoint-owner", func(t *testing.T) {
		r := graphTreeTestResult("a", "both", []string{"a", "b"}, genericTestEdge{"ab", "a", "b"}, genericTestEdge{"ab2", "a", "b"})
		text := graphTreeTestText(t, r)
		if !strings.Contains(text, "\n└── [types/references] (links/ab2) → beads/b [types/memory] (shown elsewhere)") || strings.Contains(text, "[types/references] (links/ab2) ← beads/a") {
			t.Fatalf("child stole a parallel Link from its earlier canonical endpoint:\n%s", text)
		}
	})
	t.Run("both-diamond-keeps-sibling", func(t *testing.T) {
		r := graphTreeTestResult("a", "both", []string{"a", "b", "c", "d"}, genericTestEdge{"ab", "a", "b"}, genericTestEdge{"ac", "a", "c"}, genericTestEdge{"bd", "b", "d"}, genericTestEdge{"cd", "c", "d"})
		text := graphTreeTestText(t, r)
		if !strings.Contains(text, "\n└── [types/references] (links/ac) → beads/c [types/memory] \"C\"") || !strings.Contains(text, "[types/references] (links/cd) ← beads/c [types/memory] (shown elsewhere)") {
			t.Fatalf("induced reverse edge changed canonical tree parent:\n%s", text)
		}
	})
	t.Run("depth-one-cross-link-keeps-sibling", func(t *testing.T) {
		snapshot := genericTestSnapshot(t, genericTestEdge{"ab", "a", "b"}, genericTestEdge{"ac", "a", "c"}, genericTestEdge{"bc", "b", "c"}, genericTestEdge{"cb", "c", "b"})
		r, err := projectGraphGeneric(snapshot, genericTestScope, genericTestInput("a", "out", 1))
		if err != nil {
			t.Fatal(err)
		}
		text := graphTreeTestText(t, r)
		if !strings.Contains(text, "\n└── [types/references] (links/ac) → beads/c [types/memory] \"Title c") || strings.Contains(text, "│   └── [types/references] (links/bc) → beads/c [types/memory] \"Title c") {
			t.Fatalf("induced cross-Link deepened a depth-one expansion:\n%s", text)
		}
		if !strings.Contains(text, "[types/references] (links/cb) → beads/b [types/memory]") {
			t.Fatalf("outgoing non-tree Link lost its authored source placement:\n%s", text)
		}
		if !strings.Contains(text, "[types/references] (links/bc) → beads/c [types/memory] (shown elsewhere) (cycle)") {
			t.Fatalf("future cyclic sibling lost its reference marker:\n%s", text)
		}
	})
	t.Run("outgoing-does-not-cross-incoming-bridge", func(t *testing.T) {
		snapshot := genericTestSnapshot(t, genericTestEdge{"ab", "a", "b"}, genericTestEdge{"cb", "c", "b"})
		r, err := projectGraphGeneric(snapshot, genericTestScope, genericTestInput("a", "out", 10))
		if err != nil {
			t.Fatal(err)
		}
		text := graphTreeTestText(t, r)
		if strings.Contains(text, "(links/cb)") || strings.Contains(text, "beads/c [types/memory]") || !strings.Contains(text, "2 Beads, 1 Links") {
			t.Fatalf("display crossed an incoming unselected bridge:\n%s", text)
		}
	})
}
