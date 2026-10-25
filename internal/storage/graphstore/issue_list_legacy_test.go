//go:build cgo

package graphstore

import (
	"testing"

	"github.com/steveyegge/beads/internal/types"
	publicops "github.com/steveyegge/beads/issueops"
)

func TestGraphIssueLegacyListProjection(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, _, store := issueListFixture(t, backend)
			if _, err := store.Create(ctx, CreateRequest{Title: "Context", Path: "beads/context", Body: "not an Issue"}); err != nil {
				t.Fatal(err)
			}
			ids := make(map[string]string)
			for _, path := range []string{"beads/work", "beads/prereq", "beads/parent", "beads/parent.lookalike"} {
				created, err := store.CreateIssue(ctx, path, plainIssue(path))
				if err != nil {
					t.Fatal(err)
				}
				ids[path] = created.Properties.ID
			}
			if _, err := store.AddDependency(ctx, DependencyRequest{Path: "links/block", SourcePath: "beads/work", TargetPath: "beads/prereq", Actor: "author"}); err != nil {
				t.Fatal(err)
			}
			parentType, ok := RelationshipDependencyTypeURL(store.ScopeURL(), types.DepParentChild)
			if !ok {
				t.Fatal("parent-child Type unavailable")
			}
			if _, err := store.AddDependency(ctx, DependencyRequest{Path: "links/parent", Type: types.DepParentChild, TypeURL: parentType, SourcePath: "beads/work", TargetPath: "beads/parent", Actor: "author"}); err != nil {
				t.Fatal(err)
			}
			if err := store.AddIssueComment(ctx, "beads/work", "author", "why this exists"); err != nil {
				t.Fatal(err)
			}
			page, err := store.ListIssuesLegacy(ctx, publicops.ListRequest{AllFlag: true, Limit: issueListInt(0)})
			if err != nil || page.HasMore || len(page.Items) != 4 {
				t.Fatalf("counted Issue page: %+v %v", page, err)
			}
			byID := make(map[string]bool, len(page.Items))
			for _, item := range page.Items {
				byID[item.ID] = true
				if item.ID == ids["beads/work"] && (item.DependencyCount != 1 || item.CommentCount != 1 || item.CommentsOmitted == nil || !*item.CommentsOmitted || len(item.Comments) != 0 || item.Parent == nil || *item.Parent != ids["beads/parent"]) {
					t.Fatalf("work counts and omitted marker: %+v", item)
				}
			}
			if !byID[ids["beads/work"]] || !byID[ids["beads/prereq"]] || !byID[ids["beads/parent"]] || byID["beads/context"] {
				t.Fatalf("wrong canonical Issue IDs: %+v", page.Items)
			}
			included, err := store.ListIssuesLegacy(ctx, publicops.ListRequest{AllFlag: true, Limit: issueListInt(0), IncludeComments: true})
			if err != nil {
				t.Fatal(err)
			}
			for _, item := range included.Items {
				if item.ID == ids["beads/work"] && (item.CommentsOmitted != nil || len(item.Comments) != 1 || item.Comments[0].IssueID != ids["beads/work"]) {
					t.Fatalf("included canonical comment: %+v", item)
				}
			}
			children, err := store.ListChildrenLegacy(ctx, "beads/parent")
			if err != nil || len(children) != 1 {
				t.Fatalf("native children: %+v %v", children, err)
			}
			if got := children[0]; got.ID != ids["beads/work"] || got.Parent == nil || *got.Parent != ids["beads/parent"] || got.DependencyCount != 1 || got.CommentCount != 1 || got.CommentsOmitted == nil || !*got.CommentsOmitted {
				t.Fatalf("children legacy row lost ordinary shape or native parent: %+v", got)
			}
			selector := IssueSelector{Path: "beads/parent"}
			selected, err := store.ListIssuesSelected(ctx, publicops.ListRequest{SortBy: "id", Limit: issueListInt(0)}, selector)
			if err != nil || selected.HasMore || len(selected.Items) != 2 || selected.Items[0].ID != store.ScopeURL()+"beads/parent.lookalike" || selected.Items[1].ID != store.ScopeURL()+"beads/work" {
				t.Fatalf("selected graph parent page must include edge and graph dotted child: %+v %v", selected, err)
			}
			selectedLegacy, err := store.ListIssuesLegacySelected(ctx, publicops.ListRequest{SortBy: "id", Limit: issueListInt(0)}, selector)
			if err != nil || selectedLegacy.HasMore || len(selectedLegacy.Items) != 2 || selectedLegacy.Items[0].ID != "beads/parent.lookalike" || selectedLegacy.Items[1].ID != "beads/work" || selectedLegacy.Items[1].Parent == nil || *selectedLegacy.Items[1].Parent != "beads/parent" {
				t.Fatalf("selected counted parent page disagrees with checked graph result: %+v %v", selectedLegacy, err)
			}
			empty, err := store.ListChildrenLegacy(ctx, "beads/prereq")
			if err != nil || empty == nil || len(empty) != 0 {
				t.Fatalf("empty native children: %+v %v", empty, err)
			}
		})
	}
}
