//go:build cgo

package main

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/memoryquery"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/typecatalog"
	"github.com/steveyegge/beads/internal/types"
	publicops "github.com/steveyegge/beads/issueops"
)

// Real storage hydration -> captured read -> human tree. This pins the producer
// concrete record types and verifies the bridge after the store is closed.
func TestMemoryCurrentGraphEmbeddedCapture(t *testing.T) {
	isolateBeadsDirForTest(t)
	workspace, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	o := graphstore.Options{Backend: "embedded", DataDir: filepath.Join(workspace, "dolt"), Database: "memory_bridge", Branch: "main", IssuePrefix: "mem", Binding: graphstore.Binding{WorkspaceID: workspace, ScopeURL: genericTestScope, AuthorityID: "0123456789abcdef0123456789abcdef", SchemaVersion: graphstore.SchemaVersion}}
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
	if _, err := s.CreateIssue(ctx, "beads/work", publicops.CreateRequest{Actor: "bridge-author", Issue: &publicops.Issue{Title: "Native work", Status: types.StatusOpen, IssueType: types.TypeTask, Priority: 2}}); err != nil {
		t.Fatal(err)
	}
	policy, err := s.Create(ctx, graphstore.CreateRequest{Path: "beads/rail", TypeURL: genericTestScope + "types/policy", Title: "Memory rail", Actor: "bridge-author"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.AddInformationalLink(ctx, graphstore.LinkCreateRequest{Path: "links/rail", TypeURL: genericTestScope + "types/references", SourcePath: "beads/work", TargetPath: "beads/rail", TargetPin: graphstore.EndpointPin{Revision: policy.Revision}, UnconditionalSource: true, Actor: "bridge-author"}); err != nil {
		t.Fatal(err)
	}
	read, err := s.ReadCurrentFamily(ctx, graphstore.CurrentFamilyRequest{Roots: []memoryquery.CurrentRoot{{Path: "beads/work", Depth: 1}, {Path: "beads/rail", Depth: 0}}, Family: typecatalog.Reference{URI: genericTestScope + "types/references"}, Direction: graph.DirectionOut, Limits: memoryquery.Limits{MaxNodes: 10, MaxLinks: 10, MaxInspected: 1000}, MaxBytes: graphstore.PreviewCurrentReadByteLimit})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	closed = true
	text, err := renderMemoryCurrentGraph(read, genericTestScope, graph.DirectionOut, graphstore.PreviewCurrentReadByteLimit)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, `beads/work [types/issue] (open) "Native work"`) || !strings.Contains(text, `beads/rail [types/policy] "Memory rail"`) || !strings.Contains(text, `target revision="`+policy.Revision+`"`) || strings.Count(text, "2 Beads, 1 Links across 2 views") != 1 {
		t.Fatal(text)
	}
	t.Logf("actual captured read after store close:\n%s", text)
}
