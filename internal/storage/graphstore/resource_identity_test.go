//go:build cgo

package graphstore

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	graph "github.com/steveyegge/beads/graphops"
)

func TestAllocatedResourceIdentityUsesCatalogKind(t *testing.T) {
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			ctx, _, store := memoryDeleteStore(t, engine)
			memory, err := store.Create(ctx, CreateRequest{PublicID: "links/a-memory", Title: "Literal prefix"})
			if err != nil {
				t.Fatal(err)
			}
			path := strings.TrimPrefix(memory.ID, store.ScopeURL())
			got, err := store.ResolveAllocatedResourceIdentity(ctx, "links/a-memory")
			want := AllocatedResourceIdentity{PublicID: "links/a-memory", Path: path, Kind: graph.KindBead, State: "live"}
			if err != nil || got != want {
				t.Fatalf("prefix selected kind: %+v %v", got, err)
			}
			byURI, err := store.ResolveCanonicalResourceIdentity(ctx, store.ScopeURL()+"links/a-memory")
			if err != nil || byURI != want {
				t.Fatalf("canonical public identity did not select allocation: %+v %v", byURI, err)
			}
			issueRequest := plainIssue("Catalog Issue kind")
			issueRequest.Issue.ID = "types/an-issue"
			issueRequest.ForceIDPrefix = true
			if _, err := store.CreateIssue(ctx, "beads/private-issue", issueRequest); err != nil {
				t.Fatal(err)
			}
			issueIdentity, err := store.ResolveAllocatedResourceIdentity(ctx, "types/an-issue")
			if err != nil || issueIdentity.Kind != graph.KindBead || issueIdentity.Path != "beads/private-issue" || issueIdentity.PublicID != "types/an-issue" {
				t.Fatalf("native Issue public identity: %+v %v", issueIdentity, err)
			}
			link, err := store.AddInformationalLink(ctx, LinkCreateRequest{Path: "links/reference", TypeURL: store.ScopeURL() + "types/references", SourcePath: path, TargetPath: "beads/private-issue"})
			if err != nil {
				t.Fatal(err)
			}
			linkIdentity, err := store.ResolveAllocatedResourceIdentity(ctx, "links/reference")
			if err != nil || linkIdentity.Kind != graph.KindLink || store.ScopeURL()+linkIdentity.Path != link.Link.ID {
				t.Fatalf("Link public identity: %+v %v", linkIdentity, err)
			}
			if _, err := store.Unlink(ctx, LinkDeleteRequest{Path: linkIdentity.Path, ExpectedRevision: link.Link.Revision}); err != nil {
				t.Fatal(err)
			}
			encoded, err := json.Marshal(got)
			if err != nil || string(encoded) != "{}" {
				t.Fatal("allocation helper leaked an assumed wire projection", string(encoded), err)
			}
			if _, err := store.DeleteMemory(ctx, MemoryDeleteRequest{Path: path, ExpectedRevision: memory.Revision}); err != nil {
				t.Fatal(err)
			}
			want.State = "deleted"
			got, err = store.ResolveAllocatedResourceIdentity(ctx, want.PublicID)
			if err != nil || got != want {
				t.Fatalf("deleted identity was lost: %+v %v", got, err)
			}
			before := workflowState(t, ctx, store)
			if _, err := store.ResolveCanonicalResourceIdentity(ctx, store.ScopeURL()+"a//b"); err == nil {
				t.Fatal("canonical helper accepted unselected legacy-key mapping")
			}
			if !reflect.DeepEqual(before, workflowState(t, ctx, store)) {
				t.Fatal("read helper changed state")
			}
			if _, err := store.ResolveAllocatedResourceIdentity(ctx, "missing"); !errors.Is(err, ErrNotFound) {
				t.Fatalf("missing allocation: %v", err)
			}
			// Corrupt a disposable catalog to prove a duplicate cannot silently
			// choose one kind. This is not an identity migration or authoring API.
			if _, err := store.db.ExecContext(ctx, `UPDATE graph_preview_catalog SET backing_key=? WHERE path='beads/private-issue'`, want.PublicID); err != nil {
				t.Fatal(err)
			}
			before = workflowState(t, ctx, store)
			if _, err := store.ResolveAllocatedResourceIdentity(ctx, want.PublicID); !errors.Is(err, ErrInvalidStore) {
				t.Fatalf("duplicate public ID selected a Resource: %v", err)
			}
			if !reflect.DeepEqual(before, workflowState(t, ctx, store)) {
				t.Fatal("duplicate refusal changed state")
			}
		})
	}
}
