package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/memoryquery"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
)

type memoryPresentationReader struct {
	nodes map[string]memoryquery.CurrentNode
	links []graph.Link
}

func (r memoryPresentationReader) Current(_ context.Context, path string) (memoryquery.CurrentNode, error) {
	n, ok := r.nodes[path]
	if !ok {
		return n, errors.New("missing fixture node")
	}
	return n, nil
}
func (r memoryPresentationReader) Incident(_ context.Context, path string, _ int) ([]graph.Link, error) {
	var out []graph.Link
	for _, l := range r.links {
		if l.Source().Path() == path || l.Target().Path() == path {
			out = append(out, l)
		}
	}
	return out, nil
}

// Exercise the real pure scoped reader, then bind its selected identities to
// canonical records. No whole-graph projector or Owned-copy re-traversal.
func memoryPresentationFixture(t *testing.T, roots []memoryquery.CurrentRoot, direction graph.Direction, edges ...genericTestEdge) graphstore.CurrentFamilyRead {
	t.Helper()
	r := memoryPresentationReader{nodes: map[string]memoryquery.CurrentNode{}}
	for _, name := range []string{"a", "b", "c", "d", "e"} {
		r.nodes["beads/"+name] = memoryquery.CurrentNode{Path: "beads/" + name, TypeURL: genericTestScope + "types/policy", Version: "current"}
	}
	for _, e := range edges {
		source, err := graph.NewInScopeRef("beads/"+e.source, "current")
		if err != nil {
			t.Fatal(err)
		}
		target, err := graph.NewInScopeRef("beads/"+e.target, "")
		if err != nil {
			t.Fatal(err)
		}
		revision, err := graph.NewRevision("revision-" + e.id)
		if err != nil {
			t.Fatal(err)
		}
		l, err := graph.NewLink(graph.LinkSpec{Path: "links/" + e.id, TypeURL: genericTestScope + "types/references", Revision: revision, Source: source, Target: target})
		if err != nil {
			t.Fatal(err)
		}
		r.links = append(r.links, l)
	}
	g, err := memoryquery.CurrentFamily(context.Background(), r, func(context.Context, string) (bool, error) { return true, nil }, roots, direction, memoryquery.Limits{MaxNodes: 20, MaxLinks: 20, MaxInspected: 1000})
	if err != nil {
		t.Fatal(err)
	}
	out := graphstore.CurrentFamilyRead{Graph: g}
	for _, n := range g.Nodes {
		out.Records = append(out.Records, graphstore.Record{ID: graph.CanonicalURL(genericTestScope, n.Path), Type: n.TypeURL, Version: n.Version, Properties: graphstore.Properties{Title: strings.ToUpper(strings.TrimPrefix(n.Path, "beads/"))}, Owned: []json.RawMessage{json.RawMessage(`{"id":"ignored-owned-copy"}`)}})
	}
	for _, l := range g.Links {
		out.Records = append(out.Records, graphstore.LinkRecord{ID: l.URL(genericTestScope), Type: l.TypeURL(), Version: l.Revision().String(), Source: l.Source().URL(genericTestScope), Target: l.Target().URL(genericTestScope), SourceRevision: l.Source().Pin(), TargetRevision: l.Target().Pin()})
	}
	return out
}

func memoryPresentationText(t *testing.T, read graphstore.CurrentFamilyRead, d graph.Direction) string {
	t.Helper()
	s, err := renderMemoryCurrentGraph(read, genericTestScope, d, graphstore.PreviewCurrentReadByteLimit)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestMemoryCurrentGraphIndependentRoots(t *testing.T) {
	roots := []memoryquery.CurrentRoot{{Path: "beads/a", Depth: 1}, {Path: "beads/b", Depth: 1}}
	r := memoryPresentationFixture(t, roots, graph.DirectionOut, genericTestEdge{"ab", "a", "b"}, genericTestEdge{"bd", "b", "d"})
	before := fmt.Sprintf("%#v", r)
	s := memoryPresentationText(t, r, graph.DirectionOut)
	if strings.Count(s, "3 Beads, 2 Links") != 1 || strings.Count(s, "(ab)") != 1 || strings.Count(s, "(bd)") != 1 || strings.Count(s, "b [policy] \"B\"") != 2 {
		t.Fatalf("lost per-root expansion or aggregate identity counts:\n%s", s)
	}
	if !strings.Contains(s, "Root a: complete: false; direction: out; depth: 1; frontier: 1") || !strings.Contains(s, "Root a: more relationships beyond depth 1: b") || !strings.Contains(s, "Root b: complete: true; direction: out; depth: 1; frontier: 0") {
		t.Fatalf("frontier was flattened across roots:\n%s", s)
	}
	if fmt.Sprintf("%#v", r) != before {
		t.Fatal("presenter mutated captured read")
	}
	// Same root identity at a different depth is still an independent view.
	r = memoryPresentationFixture(t, []memoryquery.CurrentRoot{{Path: "beads/a", Depth: 0}, {Path: "beads/a", Depth: 2}}, graph.DirectionOut, genericTestEdge{"ab", "a", "b"}, genericTestEdge{"bd", "b", "d"})
	s = memoryPresentationText(t, r, graph.DirectionOut)
	if strings.Count(s, "a [policy] \"A\"") != 2 || strings.Count(s, "3 Beads, 2 Links") != 1 || !strings.Contains(s, "Root a: more relationships beyond depth 0: a") {
		t.Fatal(s)
	}
}

func TestMemoryCurrentGraphParentsCyclesAndPins(t *testing.T) {
	for _, tc := range []struct {
		name      string
		direction graph.Direction
		edges     []genericTestEdge
		cycles    int
	}{
		{"diamond", graph.DirectionOut, []genericTestEdge{{"ab", "a", "b"}, {"ac", "a", "c"}, {"bd", "b", "d"}, {"cd", "c", "d"}}, 0},
		{"parallel", graph.DirectionBoth, []genericTestEdge{{"ab", "a", "b"}, {"ab2", "a", "b"}}, 0},
		{"mixed-dag", graph.DirectionBoth, []genericTestEdge{{"ab", "a", "b"}, {"cb", "c", "b"}, {"ca", "c", "a"}}, 0},
		{"directed-cycle", graph.DirectionBoth, []genericTestEdge{{"ab", "a", "b"}, {"bc", "b", "c"}, {"ca", "c", "a"}}, 1},
		{"incoming", graph.DirectionIn, []genericTestEdge{{"ba", "b", "a"}, {"cb", "c", "b"}}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := memoryPresentationFixture(t, []memoryquery.CurrentRoot{{Path: "beads/a", Depth: 4}}, tc.direction, tc.edges...)
			s := memoryPresentationText(t, r, tc.direction)
			if strings.Count(s, "(cycle)") != tc.cycles {
				t.Fatal(s)
			}
			for _, e := range tc.edges {
				if strings.Count(s, "("+e.id+")") != 1 {
					t.Fatalf("Link multiplicity %s:\n%s", e.id, s)
				}
			}
			if !strings.Contains(s, `source revision="current"`) {
				t.Fatal("lost pin", s)
			}
			if tc.name == "diamond" && !strings.Contains(s, "\n└── references (ac) → c [policy] \"C\"") {
				t.Fatal("cross-link displaced root sibling", s)
			}
			if tc.name == "incoming" && !strings.Contains(s, "(ba) ← b [policy]") {
				t.Fatal("lost authored direction", s)
			}
		})
	}
}

func TestMemoryCurrentGraphMixedTypesAndEscapes(t *testing.T) {
	r := memoryPresentationFixture(t, []memoryquery.CurrentRoot{{Path: "beads/a", Depth: 1}}, graph.DirectionOut, genericTestEdge{"ab", "a", "b"})
	for i, raw := range r.Records {
		if n, ok := raw.(graphstore.Record); ok {
			if strings.HasSuffix(n.ID, "/a") {
				n.Type = graphstore.IssueTypeURL(genericTestScope)
				r.Graph.Nodes[0].TypeURL = n.Type
				r.Records[i] = graphstore.IssueRecord{ID: n.ID, Type: n.Type, Version: n.Version, Properties: &types.Issue{Title: "A\nRoot fake", Status: types.StatusOpen}}
			} else {
				n.Properties.Title = "B\x1b[2J\u202e"
				r.Records[i] = n
			}
		}
	}
	s := memoryPresentationText(t, r, graph.DirectionOut)
	if !strings.Contains(s, `a [issue] (open) "A\nRoot fake"`) || strings.Contains(s, "\x1b") || strings.Contains(s, "\u202e") || strings.Contains(s, "b [policy] (open)") {
		t.Fatal(s)
	}
}

func TestMemoryCurrentGraphCumulativeBudget(t *testing.T) {
	r := memoryPresentationFixture(t, []memoryquery.CurrentRoot{{Path: "beads/a", Depth: 1}, {Path: "beads/b", Depth: 1}}, graph.DirectionOut, genericTestEdge{"ab", "a", "b"}, genericTestEdge{"bd", "b", "d"})
	text := memoryPresentationText(t, r, graph.DirectionOut)
	for _, size := range []int{1, len(text) / 2, len(text) - 1} {
		s, err := renderMemoryCurrentGraph(r, genericTestScope, graph.DirectionOut, size)
		if s != "" || !errors.Is(err, graphstore.ErrLimitExceeded) {
			t.Fatalf("cap%d: %q %v", size, s, err)
		}
	}
	got, err := renderMemoryCurrentGraph(r, genericTestScope, graph.DirectionOut, len(text))
	if err != nil || got != text {
		t.Fatalf("exact cap: %v", err)
	}
}

func TestMemoryCurrentGraphRefusesInconsistentCapture(t *testing.T) {
	for _, tc := range []struct {
		name string
		edit func(*graphstore.CurrentFamilyRead)
	}{
		{"missing-record", func(r *graphstore.CurrentFamilyRead) { r.Records = r.Records[1:] }},
		{"duplicate-record", func(r *graphstore.CurrentFamilyRead) { r.Records = append(r.Records, r.Records[0]) }},
		{"wrong-version", func(r *graphstore.CurrentFamilyRead) { r.Graph.Nodes[0].Version = "other" }},
		{"wrong-type", func(r *graphstore.CurrentFamilyRead) { r.Graph.Nodes[0].TypeURL += "other" }},
		{"wrong-pin", func(r *graphstore.CurrentFamilyRead) {
			for i, raw := range r.Records {
				if l, ok := raw.(graphstore.LinkRecord); ok {
					l.SourceRevision = "other"
					r.Records[i] = l
				}
			}
		}},
		{"invalid-second-root", func(r *graphstore.CurrentFamilyRead) {
			r.Graph.Views = append(r.Graph.Views, r.Graph.Views[0])
			r.Graph.Views[1].Root.Path = "beads/e"
		}},
		{"missing-root", func(r *graphstore.CurrentFamilyRead) { r.Graph.Views[0].Root.Path = "beads/e" }},
		{"duplicate-node", func(r *graphstore.CurrentFamilyRead) {
			r.Graph.Views[0].Nodes = append(r.Graph.Views[0].Nodes, r.Graph.Views[0].Nodes[0])
		}},
		{"foreign-link", func(r *graphstore.CurrentFamilyRead) {
			r.Graph.Views[0].Links = append(r.Graph.Views[0].Links, "links/missing")
		}},
		{"missing-endpoint", func(r *graphstore.CurrentFamilyRead) { r.Graph.Views[0].Nodes = r.Graph.Views[0].Nodes[:1] }},
		{"frontier-not-reached", func(r *graphstore.CurrentFamilyRead) { r.Graph.Views[0].Frontier = []string{"beads/e"} }},
		{"unrepresented-union", func(r *graphstore.CurrentFamilyRead) { r.Graph.Views[0].Links = nil }},
		{"nil-issue", func(r *graphstore.CurrentFamilyRead) {
			n := r.Records[0].(graphstore.Record)
			r.Records[0] = graphstore.IssueRecord{ID: n.ID, Type: n.Type, Version: n.Version}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := memoryPresentationFixture(t, []memoryquery.CurrentRoot{{Path: "beads/a", Depth: 1}}, graph.DirectionOut, genericTestEdge{"ab", "a", "b"})
			tc.edit(&r)
			s, err := renderMemoryCurrentGraph(r, genericTestScope, graph.DirectionOut, 10000)
			if s != "" || !errors.Is(err, graphstore.ErrInvalidStore) {
				t.Fatalf("published inconsistent capture %q %v", s, err)
			}
		})
	}
}

func TestMemoryCurrentGraphActualFamilyLabels(t *testing.T) {
	for _, typ := range []string{"superseded-by", "requires", "defines"} {
		t.Run(typ, func(t *testing.T) {
			r := memoryPresentationFixture(t, []memoryquery.CurrentRoot{{Path: "beads/a", Depth: 1}}, graph.DirectionOut, genericTestEdge{"ab", "a", "b"})
			l := r.Graph.Links[0]
			typed, err := graph.NewLink(graph.LinkSpec{Path: l.Path(), TypeURL: genericTestScope + "types/" + typ, Revision: l.Revision(), Source: l.Source(), Target: l.Target()})
			if err != nil {
				t.Fatal(err)
			}
			r.Graph.Links[0] = typed
			for i, raw := range r.Records {
				if record, ok := raw.(graphstore.LinkRecord); ok {
					record.Type = typed.TypeURL()
					r.Records[i] = record
				}
			}
			s := memoryPresentationText(t, r, graph.DirectionOut)
			if !strings.Contains(s, typ+" (ab) → b [policy]") || strings.Count(s, "2 Beads, 1 Links") != 1 {
				t.Fatal(s)
			}
		})
	}
}

func TestMemoryCurrentGraphKeepsRootLocalEdges(t *testing.T) {
	edges := []genericTestEdge{{"ab", "a", "b"}}
	roots := []memoryquery.CurrentRoot{{Path: "beads/a", Depth: 0}, {Path: "beads/b", Depth: 0}}
	r := memoryPresentationFixture(t, roots, graph.DirectionBoth, edges...)
	s := memoryPresentationText(t, r, graph.DirectionBoth)
	if len(r.Graph.Links) != 0 || strings.Contains(s, "(ab)") || !strings.Contains(s, "2 Beads, 0 Links") || strings.Count(s, "frontier: 1") != 2 {
		t.Fatalf("union re-induced an unselected Link:\n%s", s)
	}
	roots[0].Depth, roots[1].Depth = 1, 1
	r = memoryPresentationFixture(t, roots, graph.DirectionBoth, edges...)
	s = memoryPresentationText(t, r, graph.DirectionBoth)
	if strings.Count(s, "(ab)") != 2 || strings.Count(s, "2 Beads, 1 Links") != 1 || !strings.Contains(s, "(ab) → b [policy]") || !strings.Contains(s, "(ab) ← a [policy]") {
		t.Fatalf("shared Link lost per-root placement or unique count:\n%s", s)
	}
}
