package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/memoryquery"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func constraintPresentationFixture(t *testing.T, edges ...[4]string) memoryPresentationReader {
	t.Helper()
	r := memoryPresentationReader{nodes: map[string]memoryquery.CurrentNode{}}
	for _, id := range []string{"a", "b", "c", "d", "e", "f"} {
		r.nodes["beads/"+id] = memoryquery.CurrentNode{Path: "beads/" + id, TypeURL: genericTestScope + "types/constraint", Version: "current"}
	}
	for _, e := range edges {
		src, err := graph.NewInScopeRef("beads/"+e[2], "current")
		if err != nil {
			t.Fatal(err)
		}
		dst, err := graph.NewInScopeRef("beads/"+e[3], "")
		if err != nil {
			t.Fatal(err)
		}
		rev, err := graph.NewRevision("revision-" + e[0])
		if err != nil {
			t.Fatal(err)
		}
		l, err := graph.NewLink(graph.LinkSpec{Path: "links/" + e[0], TypeURL: genericTestScope + "types/" + e[1], Source: src, Target: dst, Revision: rev})
		if err != nil {
			t.Fatal(err)
		}
		r.links = append(r.links, l)
	}
	return r
}
func constraintPresentationRecords(g memoryquery.CurrentGraph) graphstore.CurrentFamilyRead {
	r := graphstore.CurrentFamilyRead{Graph: g}
	for _, n := range g.Nodes {
		r.Records = append(r.Records, graphstore.Record{ID: graph.CanonicalURL(genericTestScope, n.Path), Type: n.TypeURL, Version: n.Version, Properties: graphstore.Properties{Title: strings.ToUpper(strings.TrimPrefix(n.Path, "beads/"))}})
	}
	for _, l := range g.Links {
		r.Records = append(r.Records, graphstore.LinkRecord{ID: l.URL(genericTestScope), Type: l.TypeURL(), Version: l.Revision().String(), Source: l.Source().URL(genericTestScope), Target: l.Target().URL(genericTestScope), SourceRevision: l.Source().Pin(), TargetRevision: l.Target().Pin()})
	}
	return r
}
func constraintPresentationFamily(_ context.Context, typ string) (bool, error) {
	return typ == genericTestScope+"types/subject-to", nil
}
func constraintPresentationRequires(_ context.Context, typ string) (bool, error) {
	return typ == genericTestScope+"types/requires" || typ == genericTestScope+"types/defines", nil
}
func constraintPresentationRead(t *testing.T, roots []memoryquery.ConstraintRoot, edges ...[4]string) graphstore.CurrentConstraintsRead {
	t.Helper()
	g, err := memoryquery.CurrentConstraints(context.Background(), constraintPresentationFixture(t, edges...), constraintPresentationFamily, constraintPresentationRequires, roots, memoryquery.Limits{MaxNodes: 30, MaxLinks: 30, MaxInspected: 10000})
	if err != nil {
		t.Fatal(err)
	}
	return graphstore.CurrentConstraintsRead{CurrentFamilyRead: constraintPresentationRecords(g.Graph), Stages: g.Stages}
}
func constraintPresentationText(t *testing.T, r graphstore.CurrentConstraintsRead) string {
	t.Helper()
	s, err := renderMemoryConstraintsGraph(r, genericTestScope, 100000)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestMemoryConstraintsPhaseOccurrences(t *testing.T) {
	edges := [][4]string{{"sab", "subject-to", "a", "b"}, {"ba", "requires", "b", "a"}, {"ad", "requires", "a", "d"}, {"sde", "subject-to", "d", "e"}}
	r := constraintPresentationRead(t, []memoryquery.ConstraintRoot{{Path: "beads/a", RequiresDepth: 4}}, edges...)
	before := fmt.Sprintf("%#v", r)
	s := constraintPresentationText(t, r)
	if strings.Count(s, `a [constraint] "A"`) != 2 || !strings.Contains(s, "\n    └── requires (ba) → a [constraint] \"A\"") || !strings.Contains(s, "\n        └── requires (ad) → d [constraint]") || strings.Contains(s, "(sde)") || strings.Contains(s, "(cycle)") {
		t.Fatalf("phase1 root stole phase2 expansion or union invented a cycle:\n%s", s)
	}
	if !strings.Contains(s, "Aggregate: 3 unique Beads, 3 unique Links across 1 root views") || before != fmt.Sprintf("%#v", r) {
		t.Fatal("mutated capture or wrong aggregate", s)
	}
	// RequiresDepth0 must not import a Requires link from phase1 A to its seed.
	r = constraintPresentationRead(t, []memoryquery.ConstraintRoot{{Path: "beads/a", RequiresDepth: 0}}, edges...)
	s = constraintPresentationText(t, r)
	if strings.Contains(s, "(ba)") || strings.Contains(s, "(ad)") || !strings.Contains(s, "Root a: more Requires relationships beyond bound 0: b") {
		t.Fatal(s)
	}
	// At depth1 A is reached in phase2: its frontier survives phase1 presence.
	r = constraintPresentationRead(t, []memoryquery.ConstraintRoot{{Path: "beads/a", RequiresDepth: 1}}, edges...)
	s = constraintPresentationText(t, r)
	if !strings.Contains(s, "Root a: more Requires relationships beyond bound 1: a") || strings.Contains(s, "(ad)") {
		t.Fatal(s)
	}
}

func TestMemoryConstraintsSeedAnchorsAndCycles(t *testing.T) {
	for _, tc := range []struct {
		name   string
		edges  [][4]string
		cycles int
	}{
		{"multiple-seeds", [][4]string{{"sab", "subject-to", "a", "b"}, {"parallel", "subject-to", "a", "b"}, {"sac", "subject-to", "a", "c"}, {"bc", "requires", "b", "c"}, {"cd", "defines", "c", "d"}}, 0},
		{"requires-cycle", [][4]string{{"sab", "subject-to", "a", "b"}, {"bc", "requires", "b", "c"}, {"cb", "requires", "c", "b"}}, 1},
		{"self-subject-entry", [][4]string{{"self", "subject-to", "a", "a"}, {"ab", "requires", "a", "b"}}, 0},
		{"requires-self-cycle", [][4]string{{"sab", "subject-to", "a", "b"}, {"self", "requires", "b", "b"}}, 1},
		{"diamond", [][4]string{{"sab", "subject-to", "a", "b"}, {"bc", "requires", "b", "c"}, {"bd", "requires", "b", "d"}, {"ce", "requires", "c", "e"}, {"de", "requires", "d", "e"}}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := constraintPresentationRead(t, []memoryquery.ConstraintRoot{{Path: "beads/a", RequiresDepth: 5}}, tc.edges...)
			s := constraintPresentationText(t, r)
			for _, l := range r.Graph.Links {
				if strings.Count(s, "("+strings.TrimPrefix(l.Path(), "links/")+")") != 1 {
					t.Fatal("Link not once", s)
				}
			}
			if strings.Count(s, "(cycle)") != tc.cycles || !strings.Contains(s, `source revision="current"`) {
				t.Fatal(s)
			}
			if tc.name == "multiple-seeds" && (!strings.Contains(s, "\n└── subject-to (sac) → c [constraint] \"C\"") || !strings.Contains(s, "requires (bc) → c [constraint] (shown elsewhere)") || !strings.Contains(s, "\n    └── defines (cd) → d [constraint]")) {
				t.Fatalf("cross-seed edge displaced anchor:\n%s", s)
			}
			if tc.name == "self-subject-entry" && strings.Count(s, `a [constraint] "A"`) != 2 {
				t.Fatal("phase2 self entry collapsed", s)
			}
			for i, j := 0, len(tc.edges)-1; i < j; i, j = i+1, j-1 {
				tc.edges[i], tc.edges[j] = tc.edges[j], tc.edges[i]
			}
			again := constraintPresentationRead(t, []memoryquery.ConstraintRoot{{Path: "beads/a", RequiresDepth: 5}}, tc.edges...)
			if constraintPresentationText(t, again) != s {
				t.Fatal("unstable input permutation")
			}
		})
	}
}

func TestMemoryConstraintsIndependentSlotsAndBudget(t *testing.T) {
	r := constraintPresentationRead(t, []memoryquery.ConstraintRoot{{Path: "beads/a", RequiresDepth: 0}, {Path: "beads/a", RequiresDepth: 2}, {Path: "beads/e", RequiresDepth: 1}}, [4]string{"sab", "subject-to", "a", "b"}, [4]string{"ba", "requires", "b", "a"}, [4]string{"ad", "requires", "a", "d"})
	s := constraintPresentationText(t, r)
	if strings.Count(s, "(sab)") != 2 || strings.Count(s, "(ba)") != 1 || strings.Count(s, "(ad)") != 1 || !strings.Contains(s, "Aggregate: 4 unique Beads, 3 unique Links across 3 root views") || !strings.Contains(s, `e [constraint] "E"`) {
		t.Fatal(s)
	}
	for _, cap := range []int{1, len(s) / 2, len(s) - 1} {
		got, err := renderMemoryConstraintsGraph(r, genericTestScope, cap)
		if got != "" || !errors.Is(err, graphstore.ErrLimitExceeded) {
			t.Fatalf("cap%d returned %q %v", cap, got, err)
		}
	}
	if got, err := renderMemoryConstraintsGraph(r, genericTestScope, len(s)); err != nil || got != s {
		t.Fatal("exact cap", err)
	}
}

func TestMemoryConstraintsMalformedStages(t *testing.T) {
	for _, tc := range []struct {
		name string
		edit func(*graphstore.CurrentConstraintsRead)
	}{
		{"missing-stage", func(r *graphstore.CurrentConstraintsRead) { r.Stages = nil }},
		{"missing-record", func(r *graphstore.CurrentConstraintsRead) { r.Records = r.Records[1:] }},
		{"wrong-type", func(r *graphstore.CurrentConstraintsRead) { r.Graph.Nodes[0].TypeURL += "other" }},
		{"wrong-pin", func(r *graphstore.CurrentConstraintsRead) {
			for i, raw := range r.Records {
				if l, ok := raw.(graphstore.LinkRecord); ok {
					l.SourceRevision = "other"
					r.Records[i] = l
					break
				}
			}
		}},
		{"wrong-slot", func(r *graphstore.CurrentConstraintsRead) { r.Stages[0].Root.Path = "beads/e" }},
		{"wrong-depth", func(r *graphstore.CurrentConstraintsRead) { r.Stages[0].Root.RequiresDepth++ }},
		{"duplicate-declaration", func(r *graphstore.CurrentConstraintsRead) {
			r.Stages[0].SubjectTo = append(r.Stages[0].SubjectTo, r.Stages[0].SubjectTo[0])
		}},
		{"wrong-target-witness", func(r *graphstore.CurrentConstraintsRead) { r.Stages[0].SubjectTo[0].Target = "beads/c" }},
		{"missing-seed", func(r *graphstore.CurrentConstraintsRead) { r.Stages[0].Requires.Seeds = nil }},
		{"foreign-phase-node", func(r *graphstore.CurrentConstraintsRead) {
			r.Stages[0].Requires.Nodes = append(r.Stages[0].Requires.Nodes, "beads/e")
		}},
		{"missing-phase-link", func(r *graphstore.CurrentConstraintsRead) { r.Stages[0].Requires.Links = nil }},
		{"cross-phase-frontier", func(r *graphstore.CurrentConstraintsRead) {
			r.Stages[0].Requires.Frontier = []string{"beads/a"}
			r.Graph.Views[0].Frontier = []string{"beads/a"}
		}},
		{"late-stage-invalid", func(r *graphstore.CurrentConstraintsRead) {
			r.Stages = append(r.Stages, r.Stages[0])
			r.Graph.Views = append(r.Graph.Views, r.Graph.Views[0])
			r.Stages[1].Root.Path = "beads/e"
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := constraintPresentationRead(t, []memoryquery.ConstraintRoot{{Path: "beads/a", RequiresDepth: 1}}, [4]string{"sab", "subject-to", "a", "b"}, [4]string{"bc", "requires", "b", "c"})
			tc.edit(&r)
			s, err := renderMemoryConstraintsGraph(r, genericTestScope, 100000)
			if s != "" || !errors.Is(err, graphstore.ErrInvalidStore) {
				t.Fatalf("accepted malformed stage %q %v", s, err)
			}
		})
	}
}

func TestMemorySubjectsFixedDirectScope(t *testing.T) {
	r := constraintPresentationFixture(t, [4]string{"ab", "subject-to", "a", "b"}, [4]string{"parallel", "subject-to", "a", "b"}, [4]string{"cb", "subject-to", "c", "b"}, [4]string{"ca", "subject-to", "c", "a"}, [4]string{"bd", "subject-to", "b", "d"}, [4]string{"ac", "requires", "a", "c"})
	g, err := memoryquery.CurrentSubjects(context.Background(), r, constraintPresentationFamily, []string{"beads/b", "beads/d", "beads/e"}, memoryquery.Limits{MaxNodes: 20, MaxLinks: 20, MaxInspected: 10000})
	if err != nil {
		t.Fatal(err)
	}
	read := constraintPresentationRecords(g)
	s, err := renderMemorySubjectsGraph(read, genericTestScope, 100000)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(s, "subject-to (ab) ← a [constraint]") || !strings.Contains(s, "fixed direct incoming Subject To scope; complete: true") || strings.Contains(s, "depth") || strings.Contains(s, "bound") || strings.Contains(s, "(ca)") || strings.Contains(s, "(ac)") || !strings.Contains(s, "Aggregate: 5 unique Beads, 4 unique Links across 3 root views") {
		t.Fatal(s)
	}
	if strings.Count(s, "(parallel)") != 1 || !strings.Contains(s, `e [constraint] "E"`) {
		t.Fatal(s)
	}
	if got, err := renderMemorySubjectsGraph(read, genericTestScope, len(s)); err != nil || got != s {
		t.Fatal("exact direct cap", err)
	}
	if got, err := renderMemorySubjectsGraph(read, genericTestScope, len(s)-1); got != "" || !errors.Is(err, graphstore.ErrLimitExceeded) {
		t.Fatal("non-atomic direct cap", err)
	}
	read.Graph.Views[0].Root.Depth = 0
	if got, err := renderMemorySubjectsGraph(read, genericTestScope, 100000); got != "" || !errors.Is(err, graphstore.ErrInvalidStore) {
		t.Fatal("accepted invented direct depth0", err)
	}
}

func TestMemoryConstraintsEscapingAndAggregateCoverage(t *testing.T) {
	r := constraintPresentationRead(t, []memoryquery.ConstraintRoot{{Path: "beads/a", RequiresDepth: 2}}, [4]string{"sab", "subject-to", "a", "b"}, [4]string{"bc", "requires", "b", "c"}, [4]string{"parallel", "requires", "b", "c"})
	for i, raw := range r.Records {
		if n, ok := raw.(graphstore.Record); ok {
			n.Properties.Title += "\nRoot fake\x1b[2J\u202e"
			r.Records[i] = n
		}
	}
	s := constraintPresentationText(t, r)
	if strings.ContainsAny(s, "\x1b\u202e") || !strings.Contains(s, `\nRoot fake\x1b`) {
		t.Fatal("data controls changed framing", s)
	}
	// Drop only one parallel witness from both stage and root view. Connectivity
	// survives; the union capture still contains the Link, so coverage must refuse.
	drop := func(paths []string) []string {
		out := []string{}
		for _, p := range paths {
			if p != "links/parallel" {
				out = append(out, p)
			}
		}
		return out
	}
	r.Stages[0].Requires.Links = drop(r.Stages[0].Requires.Links)
	r.Graph.Views[0].Links = drop(r.Graph.Views[0].Links)
	if got, err := renderMemoryConstraintsGraph(r, genericTestScope, 100000); got != "" || !errors.Is(err, graphstore.ErrInvalidStore) {
		t.Fatalf("dropped selected union edge: %q %v", got, err)
	}
}

func TestMemorySubjectsRejectsSecondaryOrIncompleteScope(t *testing.T) {
	fixture := constraintPresentationFixture(t, [4]string{"ab", "subject-to", "a", "b"}, [4]string{"cb", "subject-to", "c", "b"}, [4]string{"ca", "subject-to", "c", "a"})
	g, err := memoryquery.CurrentSubjects(context.Background(), fixture, constraintPresentationFamily, []string{"beads/b"}, memoryquery.Limits{MaxNodes: 20, MaxLinks: 20, MaxInspected: 1000})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		edit func(*memoryquery.CurrentGraph)
	}{
		{"secondary-link", func(g *memoryquery.CurrentGraph) {
			for _, l := range fixture.links {
				if l.Path() == "links/ca" {
					g.Links = append(g.Links, l)
					g.Views[0].Links = append(g.Views[0].Links, l.Path())
				}
			}
		}},
		{"frontier", func(g *memoryquery.CurrentGraph) { g.Views[0].Frontier = []string{"beads/a"} }},
		{"late-invalid", func(g *memoryquery.CurrentGraph) { g.Views = append(g.Views, g.Views[0]); g.Views[1].Root.Depth = 0 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			copyGraph := g
			copyGraph.Links = append([]graph.Link(nil), g.Links...)
			copyGraph.Views = append([]memoryquery.CurrentView(nil), g.Views...)
			tc.edit(&copyGraph)
			got, err := renderMemorySubjectsGraph(constraintPresentationRecords(copyGraph), genericTestScope, 100000)
			if got != "" || !errors.Is(err, graphstore.ErrInvalidStore) {
				t.Fatalf("accepted non-direct capture: %q %v", got, err)
			}
		})
	}
}
