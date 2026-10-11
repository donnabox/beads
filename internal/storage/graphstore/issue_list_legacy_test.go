//go:build cgo

package graphstore

import (
	"testing"

	publicops "github.com/steveyegge/beads/issueops"
)

func TestGraphIssueLegacyListProjection(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, _, store := issueListFixture(t, backend)
			if _, err := store.Create(ctx, CreateRequest{Path: "beads/context", Body: "not an Issue"}); err != nil {
				t.Fatal(err)
			}
			for _, path := range []string{"beads/work", "beads/prereq"} {
				if _, err := store.CreateIssue(ctx, path, plainIssue(path)); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := store.AddDependency(ctx, DependencyRequest{Path: "links/block", SourcePath: "beads/work", TargetPath: "beads/prereq", Actor: "author"}); err != nil {
				t.Fatal(err)
			}
			if err := store.AddIssueComment(ctx, "beads/work", "author", "why this exists"); err != nil {
				t.Fatal(err)
			}
			page, err := store.ListIssuesLegacy(ctx, publicops.ListRequest{AllFlag: true, Limit: issueListInt(0)})
			if err != nil || page.HasMore || len(page.Items) != 2 {
				t.Fatalf("counted Issue page: %+v %v", page, err)
			}
			byID := make(map[string]bool, len(page.Items))
			for _, item := range page.Items {
				byID[item.ID] = true
				if item.ID == "beads/work" && (item.DependencyCount != 1 || item.CommentCount != 1 || item.CommentsOmitted == nil || !*item.CommentsOmitted || len(item.Comments) != 0) {
					t.Fatalf("work counts and omitted marker: %+v", item)
				}
			}
			if !byID["beads/work"] || !byID["beads/prereq"] || byID["beads/context"] {
				t.Fatalf("wrong canonical Issue IDs: %+v", page.Items)
			}
			included, err := store.ListIssuesLegacy(ctx, publicops.ListRequest{AllFlag: true, Limit: issueListInt(0), IncludeComments: true})
			if err != nil {
				t.Fatal(err)
			}
			for _, item := range included.Items {
				if item.ID == "beads/work" && (item.CommentsOmitted != nil || len(item.Comments) != 1 || item.Comments[0].IssueID != "beads/work") {
					t.Fatalf("included canonical comment: %+v", item)
				}
			}
		})
	}
}
