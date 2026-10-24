//go:build cgo

package main

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/memoryquery"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
	publicops "github.com/steveyegge/beads/issueops"
)

func TestMemoryConstraintsAndSubjectsEmbeddedCapture(t *testing.T) {
	isolateBeadsDirForTest(t)
	workspace, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	o := graphstore.Options{Backend: "embedded", DataDir: filepath.Join(workspace, "dolt"), Database: "constraint_presenter", Branch: "main", IssuePrefix: "con", Binding: graphstore.Binding{WorkspaceID: workspace, ScopeURL: genericTestScope, AuthorityID: "0123456789abcdef0123456789abcdef", SchemaVersion: graphstore.SchemaVersion}}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	if err := graphstore.Init(ctx, o); err != nil {
		t.Fatal(err)
	}
	s, err := graphstore.OpenExisting(ctx, o)
	if err != nil {
		t.Fatal(err)
	}
	closed := false
	t.Cleanup(func() {
		if !closed {
			if err := s.Close(); err != nil {
				t.Error(err)
			}
		}
	})
	if _, err := s.CreateIssue(ctx, "beads/work", publicops.CreateRequest{Actor: "presenter", ForceIDPrefix: true, Issue: &publicops.Issue{ID: "beads/work", Title: "Work", Status: types.StatusOpen, IssueType: types.TypeTask, Priority: 2}}); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"a", "b", "c", "d"} {
		if _, err := s.Create(ctx, graphstore.CreateRequest{Path: "beads/" + id, TypeURL: genericTestScope + "types/policy", Title: strings.ToUpper(id), Actor: "presenter"}); err != nil {
			t.Fatal(err)
		}
	}
	for _, e := range [][4]string{{"ba", "requires", "b", "a"}, {"ad", "defines", "a", "d"}, {"not-replayed", "subject-to", "a", "c"}} {
		if _, err := s.AddInformationalLink(ctx, graphstore.LinkCreateRequest{Path: "links/" + e[0], TypeURL: genericTestScope + "types/" + e[1], SourcePath: "beads/" + e[2], TargetPath: "beads/" + e[3], UnconditionalSource: true, Actor: "presenter"}); err != nil {
			t.Fatal(err)
		}
	}
	b, err := s.Show(ctx, "beads/b")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.AddInformationalLink(ctx, graphstore.LinkCreateRequest{Path: "links/subject", TypeURL: genericTestScope + "types/subject-to", SourcePath: "beads/work", TargetPath: "beads/b", TargetPin: graphstore.EndpointPin{Revision: b.Revision}, UnconditionalSource: true, Actor: "presenter"}); err != nil {
		t.Fatal(err)
	}
	limits := memoryquery.Limits{MaxNodes: 20, MaxLinks: 20, MaxInspected: 10000}
	constraints, err := s.ReadCurrentConstraints(ctx, graphstore.CurrentConstraintsRequest{Roots: []graphstore.CurrentConstraintRoot{{Path: "beads/work", RequiresDepth: 3}}, Limits: limits, MaxBytes: graphstore.PreviewCurrentReadByteLimit})
	if err != nil {
		t.Fatal(err)
	}
	subjects, err := s.ReadCurrentSubjects(ctx, graphstore.CurrentSubjectsRequest{Roots: []string{"beads/b"}, Limits: limits, MaxBytes: graphstore.PreviewCurrentReadByteLimit})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	closed = true
	text, err := renderMemoryConstraintsGraph(constraints, genericTestScope, 100000)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, `beads/work [types/issue] (open) "Work"`) || !strings.Contains(text, `target revision="`+b.Revision+`"`) || !strings.Contains(text, "[types/defines] (links/ad) → beads/d [types/policy]") || strings.Contains(text, "not-replayed") || !strings.Contains(text, "4 Beads, 3 Links across 1 views") {
		t.Fatal(text)
	}
	t.Logf("actual staged capture after Close:\n%s", text)
	text, err = renderMemorySubjectsGraph(subjects, genericTestScope, 100000)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "[types/subject-to] (links/subject) ← beads/work [types/issue]") || strings.Contains(text, "complete:") || strings.Contains(text, "frontier:") || strings.Contains(text, "requires") || strings.Contains(text, "depth") {
		t.Fatal(text)
	}
	t.Logf("actual direct capture after Close:\n%s", text)
}
