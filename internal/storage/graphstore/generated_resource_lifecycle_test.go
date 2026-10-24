//go:build cgo

package graphstore

import (
	"errors"
	"strings"
	"testing"

	graph "github.com/steveyegge/beads/graphops"
)

func TestGeneratedResourceIDsAreCanonicalShortAndRooted(t *testing.T) {
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			ctx, _, s := memoryDeleteStore(t, engine)
			one, err := s.Create(ctx, CreateRequest{Title: "One"})
			if err != nil {
				t.Fatal(err)
			}
			two, err := s.Create(ctx, CreateRequest{IDRoot: "notes/team", Title: "Two"})
			if err != nil {
				t.Fatal(err)
			}
			if len(one.PublicID) != 6 || !strings.HasPrefix(two.PublicID, "notes/team/") || len(strings.TrimPrefix(two.PublicID, "notes/team/")) != 6 {
				t.Fatalf("generated IDs: %q, %q", one.PublicID, two.PublicID)
			}
			onePath := fixtureResourcePath(t, ctx, s, one.ID)
			twoPath := fixtureResourcePath(t, ctx, s, two.ID)
			linked, err := s.AddInformationalLink(ctx, LinkCreateRequest{IDRoot: "notes/refs", SourcePath: onePath, TargetPath: twoPath, TypeURL: s.ScopeURL() + "types/references", UnconditionalSource: true})
			if err != nil {
				t.Fatal(err)
			}
			linkID := strings.TrimPrefix(linked.Link.ID, s.ScopeURL())
			if !strings.HasPrefix(linkID, "notes/refs/") || len(strings.TrimPrefix(linkID, "notes/refs/")) != 6 {
				t.Fatalf("Link ID: %q", linked.Link.ID)
			}
			retained, err := s.ReadVersion(ctx, linked.Link.ID, linked.Link.Revision)
			if err != nil || retained.(LinkRecord).ID != linked.Link.ID {
				t.Fatalf("retained identity: %+v %v", retained, err)
			}
			issue, err := s.CreateIssueWithIDRoot(ctx, "work", plainIssue("Rooted issue"))
			if err != nil {
				t.Fatal(err)
			}
			if !strings.HasPrefix(issue.Properties.ID, "work/") || len(strings.TrimPrefix(issue.Properties.ID, "work/")) != 6 || issue.ID != s.ScopeURL()+issue.Properties.ID {
				t.Fatalf("Issue ID: %+v", issue)
			}
			if _, err := s.Create(ctx, CreateRequest{PublicID: "explicit", IDRoot: "root", Title: "Bad"}); !errors.Is(err, graph.ErrValidation) {
				t.Fatalf("ID+root: %v", err)
			}
			if _, err := s.AddInformationalLink(ctx, LinkCreateRequest{PublicID: "explicit", IDRoot: "root"}); err == nil {
				t.Fatal("Link ID+root accepted")
			}
		})
	}
}
