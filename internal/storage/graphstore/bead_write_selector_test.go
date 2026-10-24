//go:build cgo

package graphstore

import "testing"

func TestInformationalPairWritesSelectOpaquePublicBeadID(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, options := issueExperimentOptions(t, backend)
			store, err := OpenExisting(ctx, options)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = store.Close() })
			owner := plainIssue("Path owner")
			owner.Issue.ID = "exp-x"
			pathOwner, err := store.CreateIssue(ctx, "beads/exp-x", owner)
			if err != nil {
				t.Fatal(err)
			}
			literal := plainIssue("Literal public ID")
			literal.Issue.ID = "beads/exp-x"
			literal.ForceIDPrefix = true
			selected, err := store.CreateIssue(ctx, "", literal)
			if err != nil {
				t.Fatal(err)
			}
			memory, err := store.Create(ctx, CreateRequest{Path: "beads/context", Title: "Context"})
			if err != nil {
				t.Fatal(err)
			}
			for _, reverse := range []bool{false, true} {
				request := LinkCreateRequest{TypeURL: store.ScopeURL() + "types/references", Path: "links/out", SourcePath: "beads/beads/exp-x", SourceLocalID: "beads/exp-x", TargetPath: "beads/context", TargetLocalID: "context"}
				if reverse {
					request.Path = "links/in"
					request.SourcePath, request.TargetPath = request.TargetPath, request.SourcePath
					request.SourceLocalID, request.TargetLocalID = request.TargetLocalID, request.SourceLocalID
				}
				created, err := store.AddInformationalLink(ctx, request)
				if err != nil {
					t.Fatal(err)
				}
				wantSource, wantTarget := selected.ID, memory.ID
				if reverse {
					wantSource, wantTarget = memory.ID, selected.ID
				}
				if created.Link.Source != wantSource || created.Link.Target != wantTarget || created.Link.Source == pathOwner.ID || created.Link.Target == pathOwner.ID {
					t.Fatalf("informational Link selected path owner: %+v", created.Link)
				}
				removed, err := store.Unlink(ctx, LinkDeleteRequest{
					SourcePath: request.SourcePath, SourceLocalID: request.SourceLocalID,
					TargetPath: request.TargetPath, TargetLocalID: request.TargetLocalID,
					TypeURL: request.TypeURL, Unconditional: true, DefaultInformationalSource: true,
				})
				if err != nil || !removed.Changed || removed.Link.ID != created.Link.ID {
					t.Fatalf("pair unlink selected path owner or wrong Link: %+v, %v", removed, err)
				}
			}
		})
	}
}
