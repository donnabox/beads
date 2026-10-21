//go:build cgo

package graphstore

import (
	"errors"
	"slices"
	"testing"

	"github.com/steveyegge/beads/internal/types"
	publicops "github.com/steveyegge/beads/issueops"
)

func TestGraphIssueTreeSelectedIsCheckedAndUnpaged(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, _, store := issueListFixture(t, backend)
			for _, path := range []string{"beads/root", "beads/child", "beads/grand", "beads/root.dotted", "beads/root.linked-elsewhere", "beads/unrelated"} {
				if _, err := store.CreateIssue(ctx, path, plainIssue(path)); err != nil {
					t.Fatal(err)
				}
			}
			parentType, ok := RelationshipDependencyTypeURL(store.ScopeURL(), types.DepParentChild)
			if !ok {
				t.Fatal("parent-child Type unavailable")
			}
			for _, edge := range []struct{ path, child, parent string }{
				{"links/root-child", "beads/child", "beads/root"},
				{"links/child-grand", "beads/grand", "beads/child"},
				{"links/elsewhere", "beads/root.linked-elsewhere", "beads/unrelated"},
			} {
				if _, err := store.AddDependency(ctx, DependencyRequest{Path: edge.path, Type: types.DepParentChild, TypeURL: parentType, SourcePath: edge.child, TargetPath: edge.parent, Actor: "author"}); err != nil {
					t.Fatal(err)
				}
			}
			selector := IssueSelector{Path: "beads/root"}
			page, err := store.ListIssueTreeSelected(ctx, publicops.ListRequest{Limit: issueListInt(1)}, selector)
			if err != nil {
				t.Fatal(err)
			}
			if page.Root == nil || page.Root.ID != "beads/root" || len(page.Items) != 3 {
				t.Fatalf("tree must be complete despite page limit: %+v", page)
			}
			ids := make([]string, 0, len(page.Items))
			for _, item := range page.Items {
				ids = append(ids, item.ID)
			}
			slices.Sort(ids)
			if want := []string{"beads/child", "beads/grand", "beads/root.dotted"}; !slices.Equal(ids, want) {
				t.Fatalf("selected descendants = %v, want %v", ids, want)
			}
			if !slices.Equal(page.Children["beads/child"], []string{"beads/grand"}) || !slices.Equal(page.Children["beads/root"], []string{"beads/child", "beads/root.dotted"}) {
				t.Fatalf("wrong hierarchy: %+v", page.Children)
			}
			limited, err := store.ListIssueTreeSelected(ctx, publicops.ListRequest{MaxRows: 2, MaxRowsSource: "--max-rows"}, selector)
			if !errors.Is(err, ErrLimitExceeded) || limited.Root != nil || len(limited.Items) != 0 {
				t.Fatalf("whole-tree cap must refuse without a partial result: %+v %v", limited, err)
			}
		})
	}
}

func TestGraphIssueTreeSelectedExcludesDottedLinkedElsewhere(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, _, store := issueListFixture(t, backend)
			for _, path := range []string{"beads/root", "beads/root.a", "beads/other"} {
				if _, err := store.CreateIssue(ctx, path, plainIssue(path)); err != nil {
					t.Fatal(err)
				}
			}
			parentType, ok := RelationshipDependencyTypeURL(store.ScopeURL(), types.DepParentChild)
			if !ok {
				t.Fatal("parent-child Type unavailable")
			}
			if _, err := store.AddDependency(ctx, DependencyRequest{Path: "links/elsewhere", Type: types.DepParentChild, TypeURL: parentType, SourcePath: "beads/root.a", TargetPath: "beads/other", Actor: "author"}); err != nil {
				t.Fatal(err)
			}
			tree, err := store.ListIssueTreeSelected(ctx, publicops.ListRequest{}, IssueSelector{Path: "beads/root"})
			if err != nil {
				t.Fatal(err)
			}
			if len(tree.Items) != 0 {
				t.Fatalf("a dotted Issue linked to another parent is not an orphan descendant: %+v", tree)
			}
		})
	}
}
