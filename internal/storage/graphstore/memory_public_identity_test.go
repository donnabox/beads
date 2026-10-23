//go:build cgo

package graphstore

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/storage"
)

func TestMemoryPublicIdentityExactLifecycle(t *testing.T) {
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			ctx, _, s := memoryDeleteStore(t, engine)
			keys := []string{"plain", " plain ", "PLAIN", "雪/%@?#", "é", "é", "beads/plain", "beads/free", strings.Repeat("long", 600)}
			for _, key := range keys {
				body := "  Exact body\n雪\n"
				created, err := s.Create(ctx, CreateRequest{PublicID: key, Title: "Independent title", Body: body})
				if err != nil {
					t.Fatal(err)
				}
				path := strings.TrimPrefix(created.ID, s.ScopeURL())
				if created.PublicID != key || created.Properties.Body != body {
					t.Fatalf("identity/content changed: %+v", created)
				}
				selected, err := s.ResolveExactAllocatedResourcePath(ctx, key)
				if err != nil || selected != path {
					t.Fatalf("resolve %q: %q %v", key, selected, err)
				}
				_, read, err := s.ReadResourceSelected(ctx, "beads/"+key, key)
				if err != nil || !reflect.DeepEqual(read, created) {
					t.Fatalf("read %q: %+v %v", key, read, err)
				}
				retained, err := s.ReadVersion(ctx, path, created.Revision)
				if err != nil || !reflect.DeepEqual(retained, created) {
					t.Fatalf("retained %q: %+v %v", key, retained, err)
				}
				next := "Replacement body"
				edited, err := s.PatchMemory(ctx, MemoryPatchRequest{Path: path, Body: &next, ExpectedRevision: created.Revision})
				if err != nil || edited.Memory.PublicID != key || edited.Memory.Properties.Title != created.Properties.Title {
					t.Fatalf("overwrite: %+v %v", edited, err)
				}
				deleted, err := s.DeleteMemory(ctx, MemoryDeleteRequest{Path: path, ExpectedRevision: edited.Memory.Revision})
				if err != nil || deleted.Terminal == nil {
					t.Fatalf("delete: %+v %v", deleted, err)
				}
				before := workflowState(t, ctx, s)
				if _, err := s.Create(ctx, CreateRequest{PublicID: key, Title: "Forbidden reuse"}); !errors.Is(err, ErrAlreadyExists) {
					t.Fatalf("reuse: %v", err)
				}
				if !reflect.DeepEqual(before, workflowState(t, ctx, s)) {
					t.Fatal("reuse changed state")
				}
				restored, err := s.PatchMemory(ctx, MemoryPatchRequest{Path: path, FromRevision: created.Revision, ExpectedRevision: deleted.Terminal.Revision})
				if err != nil || restored.Memory.PublicID != key || restored.Memory.ID != created.ID || restored.Memory.Revision == created.Revision {
					t.Fatalf("recovery: %+v %v", restored, err)
				}
			}
		})
	}
}

func TestMemoryPublicIdentityCrossKindAndAmbiguity(t *testing.T) {
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			ctx, _, s := memoryDeleteStore(t, engine)
			issueReq := plainIssue("Issue")
			issueReq.Issue.ID = "exp-overlap"
			if _, err := s.CreateIssue(ctx, "beads/issue", issueReq); err != nil {
				t.Fatal(err)
			}
			before := workflowState(t, ctx, s)
			if _, err := s.Create(ctx, CreateRequest{PublicID: "exp-overlap", Title: "Collision"}); !errors.Is(err, ErrAlreadyExists) {
				t.Fatalf("collision: %v", err)
			}
			if !reflect.DeepEqual(before, workflowState(t, ctx, s)) {
				t.Fatal("collision changed state")
			}
			one, err := s.Create(ctx, CreateRequest{PublicID: "one", Title: "One"})
			if err != nil {
				t.Fatal(err)
			}
			two, err := s.Create(ctx, CreateRequest{PublicID: one.ID, Title: "URL-shaped key"})
			if err != nil {
				t.Fatal(err)
			}
			before = workflowState(t, ctx, s)
			if _, err := s.ResolveMemorySelector(ctx, one.ID); !errors.Is(err, storage.ErrValidation) {
				t.Fatalf("ambiguous selector chose target: %v", err)
			}
			selected, err := s.ResolveExactAllocatedResourcePath(ctx, one.ID)
			if err != nil || s.ScopeURL()+selected != two.ID {
				t.Fatalf("explicit literal selection: %s %v", selected, err)
			}
			if !reflect.DeepEqual(before, workflowState(t, ctx, s)) {
				t.Fatal("ambiguity changed state")
			}
		})
	}
}

func TestMemoryPublicIdentityIsNotBDPWire(t *testing.T) {
	record := Record{ID: "https://example.invalid/beads/locator", Type: "https://example.invalid/types/memory", Revision: "r", Version: "r", Properties: Properties{Title: "Title", Body: "Body"}}
	before, err := canonicalJSON(record)
	if err != nil {
		t.Fatal(err)
	}
	record.PublicID = " exact 雪 @key "
	after, err := canonicalJSON(record)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("local public ID altered pinned canonical record JSON")
	}
}
