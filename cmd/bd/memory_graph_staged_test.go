package main

import (
	"errors"
	"reflect"
	"testing"

	"github.com/steveyegge/beads/internal/httpapi/bdpwire"
	"github.com/steveyegge/beads/internal/memoryquery"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func memoryStagedContextFixture() (memoryGraphResult, graphstore.CurrentConstraintsRead) {
	const scope = "https://example.invalid/staged-context/"
	bead := func(id string) string { return scope + "beads/" + id }
	link := func(id string) string { return scope + "links/" + id }
	result := memoryGraphResult{Scope: scope, Links: []bdpwire.LinkRecord{
		{ID: link("sc"), Source: bdpwire.Reference{URI: bead("s")}, Target: bdpwire.Reference{URI: bead("c")}},
		{ID: link("cs"), Source: bdpwire.Reference{URI: bead("c")}, Target: bdpwire.Reference{URI: bead("s")}},
		{ID: link("sd"), Source: bdpwire.Reference{URI: bead("s")}, Target: bdpwire.Reference{URI: bead("d")}},
	}, Traversal: memoryGraphTraversal{Roots: []memoryGraphRoot{{ID: bead("s"), Revision: "current", Depth: func() *int { n := 2; return &n }(), Beads: []string{bead("c"), bead("d"), bead("s")}, Links: []string{link("cs"), link("sc"), link("sd")}, Frontier: []string{}}}}}
	captured := graphstore.CurrentConstraintsRead{Stages: []memoryquery.ConstraintStage{{Root: memoryquery.ConstraintRoot{Path: "beads/s", RequiresDepth: 2}, SubjectTo: []memoryquery.ConstraintSeed{{Link: "links/sc", Target: "beads/c"}}, Requires: memoryquery.CurrentStage{Seeds: []string{"beads/c"}, Nodes: []string{"beads/c", "beads/d", "beads/s"}, Links: []string{"links/cs", "links/sd"}, Frontier: []string{}}}}}
	return result, captured
}

func TestMemoryConstraintStageContext(t *testing.T) {
	result, read := memoryStagedContextFixture()
	if err := memoryConstraintStages(&result, read); err != nil {
		t.Fatal(err)
	}
	stage := result.Traversal.Roots[0].Stages
	if stage == nil || len(stage.Requires.Beads) != 3 || len(stage.Requires.Links) != 2 || len(stage.SubjectTo) != 1 || stage.SubjectTo[0].Target != result.Scope+"beads/c" {
		t.Fatalf("canonical phase witness lost: %+v", stage)
	}
	for _, tc := range []struct {
		name   string
		change func(*memoryGraphResult, *graphstore.CurrentConstraintsRead)
	}{
		{"missing-stage", func(_ *memoryGraphResult, r *graphstore.CurrentConstraintsRead) { r.Stages = nil }},
		{"wrong-root-slot", func(_ *memoryGraphResult, r *graphstore.CurrentConstraintsRead) { r.Stages[0].Root.Path = "beads/c" }},
		{"wrong-stage-depth", func(_ *memoryGraphResult, r *graphstore.CurrentConstraintsRead) { r.Stages[0].Root.RequiresDepth = 1 }},
		{"wrong-declaration-source", func(_ *memoryGraphResult, r *graphstore.CurrentConstraintsRead) {
			r.Stages[0].SubjectTo[0].Link = "links/cs"
		}},
		{"duplicate-declaration", func(_ *memoryGraphResult, r *graphstore.CurrentConstraintsRead) {
			r.Stages[0].SubjectTo = append(r.Stages[0].SubjectTo, r.Stages[0].SubjectTo[0])
		}},
		{"wrong-seed-set", func(_ *memoryGraphResult, r *graphstore.CurrentConstraintsRead) {
			r.Stages[0].Requires.Seeds = []string{"beads/s"}
		}},
		{"missing-required-node", func(_ *memoryGraphResult, r *graphstore.CurrentConstraintsRead) {
			r.Stages[0].Requires.Nodes = []string{"beads/c", "beads/d"}
		}},
		{"duplicate-required-link", func(_ *memoryGraphResult, r *graphstore.CurrentConstraintsRead) {
			r.Stages[0].Requires.Links = append(r.Stages[0].Requires.Links, "links/sd")
		}},
		{"missing-required-link", func(_ *memoryGraphResult, r *graphstore.CurrentConstraintsRead) {
			r.Stages[0].Requires.Links = []string{"links/cs"}
		}},
		{"unanchored-phase", func(result *memoryGraphResult, r *graphstore.CurrentConstraintsRead) {
			r.Stages[0].SubjectTo = nil
			r.Stages[0].Requires.Seeds = nil
			result.Links = append(result.Links[:0], result.Links[1:]...)
			result.Traversal.Roots[0].Links = []string{result.Scope + "links/cs", result.Scope + "links/sd"}
		}},
		{"phase-edge-overlap", func(_ *memoryGraphResult, r *graphstore.CurrentConstraintsRead) {
			r.Stages[0].Requires.Links = append(r.Stages[0].Requires.Links, "links/sc")
		}},
		{"wrong-frontier", func(_ *memoryGraphResult, r *graphstore.CurrentConstraintsRead) {
			r.Stages[0].Requires.Frontier = []string{"beads/c"}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result, read := memoryStagedContextFixture()
			tc.change(&result, &read)
			if err := memoryConstraintStages(&result, read); !errors.Is(err, graphstore.ErrInvalidStore) {
				t.Fatalf("accepted bad stage context: %v", err)
			}
			if result.Traversal.Roots[0].Stages != nil {
				t.Fatal("failed validation published partial root context")
			}
		})
	}
	// Even a later-root failure must not mutate an earlier root's context.
	result, read = memoryStagedContextFixture()
	result.Traversal.Roots = append(result.Traversal.Roots, result.Traversal.Roots[0])
	read.Stages = append(read.Stages, read.Stages[0])
	read.Stages[1].Root.RequiresDepth++
	before := append([]memoryGraphRoot{}, result.Traversal.Roots...)
	if err := memoryConstraintStages(&result, read); err == nil || !reflect.DeepEqual(before, result.Traversal.Roots) {
		t.Fatalf("later-root failure changed context: %v", err)
	}
}
