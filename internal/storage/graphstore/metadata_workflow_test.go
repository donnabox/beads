//go:build cgo

package graphstore

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/storage"
	publicops "github.com/steveyegge/beads/issueops"
)

func TestCommonMetadataStoredAndRetainedAcrossResourceKinds(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, o := issueExperimentOptions(t, backend)
			s, err := OpenExisting(ctx, o)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = s.Close() })

			memory, err := s.Create(ctx, CreateRequest{Path: "beads/plan", Title: "Plan", Body: "Body", Metadata: json.RawMessage(`{"team":"docs"}`)})
			if err != nil || !bytes.Equal(memory.Metadata, []byte(`{"team":"docs"}`)) {
				t.Fatalf("Memory create metadata = %s, %v", memory.Metadata, err)
			}
			issueRequest := plainIssue("Work")
			issueRequest.Issue.Metadata = json.RawMessage(`{"team":"issues"}`)
			issue, err := s.CreateIssue(ctx, "beads/work", issueRequest)
			if err != nil || !bytes.Equal(issue.Metadata, []byte(`{"team":"issues"}`)) || len(issue.Properties.Metadata) != 0 {
				t.Fatalf("Issue create metadata = %s, properties=%+v, %v", issue.Metadata, issue.Properties, err)
			}
			link, err := s.AddInformationalLink(ctx, LinkCreateRequest{Path: "links/context", SourcePath: "beads/plan", TargetPath: "beads/work", ExpectedSourceRevision: memory.Revision, Properties: map[string]any{"note": "original"}, Metadata: json.RawMessage(`{"source":"manual"}`)})
			if err != nil || !bytes.Equal(link.Link.Metadata, []byte(`{"source":"manual"}`)) {
				t.Fatalf("Link create metadata = %s, %v", link.Link.Metadata, err)
			}
			owned := link.Source.(Record)
			merge := publicops.MetadataPatch{Merge: publicops.Field[json.RawMessage]{Set: true, Value: json.RawMessage(`{"team":"release"}`)}}
			memoryChanged, err := s.PatchMemory(ctx, MemoryPatchRequest{Path: "beads/plan", ExpectedRevision: owned.Revision, Metadata: merge})
			if err != nil || !memoryChanged.Changed || memoryChanged.Memory.Properties != memory.Properties || !bytes.Equal(memoryChanged.Memory.Metadata, []byte(`{"team":"release"}`)) {
				t.Fatalf("Memory metadata-only edit = %+v, %v", memoryChanged, err)
			}
			if _, err := s.PatchMemory(ctx, MemoryPatchRequest{Path: "beads/plan", ExpectedRevision: owned.Revision, Metadata: merge}); !errors.Is(err, ErrConflict) {
				t.Fatalf("stale Memory metadata no-op guard = %v", err)
			}
			memoryNoop, err := s.PatchMemory(ctx, MemoryPatchRequest{Path: "beads/plan", ExpectedRevision: memoryChanged.Memory.Revision, Metadata: merge})
			if err != nil || memoryNoop.Changed || memoryNoop.Memory.Revision != memoryChanged.Memory.Revision {
				t.Fatalf("Memory metadata no-op = %+v, %v", memoryNoop, err)
			}
			linkChanged, err := s.UpdateLink(ctx, LinkUpdateRequest{Path: "links/context", ExpectedRevision: link.Link.Revision, ExpectedSourceRevision: memoryChanged.Memory.Revision, MetadataOnly: true, Metadata: publicops.MetadataPatch{Set: map[string]json.RawMessage{"source": json.RawMessage(`"reviewed"`)}}})
			if err != nil || !linkChanged.Changed || linkChanged.Link.Properties["note"] != "original" || !bytes.Equal(linkChanged.Link.Metadata, []byte(`{"source":"reviewed"}`)) {
				t.Fatalf("Link metadata-only edit = %+v, %v", linkChanged, err)
			}
			if _, err := s.UpdateLink(ctx, LinkUpdateRequest{Path: "links/context", ExpectedRevision: linkChanged.Link.Revision, ExpectedSourceRevision: memoryChanged.Memory.Revision, MetadataOnly: true, Metadata: merge}); !errors.Is(err, ErrConflict) {
				t.Fatalf("stale owning-source guard = %v", err)
			}
			issueChanged, err := s.UpdateIssue(ctx, UpdateIssueRequest{Path: "beads/work", Actor: "editor", ExpectedRevision: issue.Revision, Metadata: merge})
			if err != nil || !issueChanged.Changed || !bytes.Equal(issueChanged.Issue.Metadata, []byte(`{"team":"release"}`)) || len(issueChanged.Issue.Properties.Metadata) != 0 {
				t.Fatalf("Issue metadata-only edit = %+v, %v", issueChanged, err)
			}
			// The ordinary metadata merge unmarshals an object into a map. Graph
			// admission must reject duplicate raw members before that loses them.
			beforeDuplicate := workflowState(t, ctx, s)
			duplicate := publicops.MetadataPatch{Merge: publicops.Field[json.RawMessage]{Set: true, Value: json.RawMessage(`{"team":"one","team":"two"}`)}}
			if got, err := s.UpdateIssue(ctx, UpdateIssueRequest{Path: "beads/work", Actor: "editor", ExpectedRevision: issueChanged.Issue.Revision, Metadata: duplicate}); !errors.Is(err, storage.ErrValidation) || !reflect.DeepEqual(got, IssueMutationResult{}) {
				t.Fatalf("duplicate metadata merge must refuse with zero Issue result: %+v, %v", got, err)
			}
			if !reflect.DeepEqual(beforeDuplicate, workflowState(t, ctx, s)) {
				t.Fatal("duplicate metadata merge changed native Issue or History state")
			}
			// The ordinary Issue writer accepts broader JSON than graph I-JSON.
			// A refused graph update must roll back its native row and History too.
			invalid := publicops.MetadataPatch{Merge: publicops.Field[json.RawMessage]{Set: true, Value: json.RawMessage(`{"nanoseconds":1727000000000000000}`)}}
			if _, err := s.UpdateIssue(ctx, UpdateIssueRequest{Path: "beads/work", Actor: "editor", ExpectedRevision: issueChanged.Issue.Revision, Metadata: invalid}); err == nil {
				t.Fatal("Issue accepted metadata outside graph I-JSON")
			}
			unchanged, err := s.Read(ctx, "beads/work")
			if err != nil {
				t.Fatal(err)
			}
			if got := unchanged.(IssueRecord); got.Revision != issueChanged.Issue.Revision || !bytes.Equal(got.Metadata, issueChanged.Issue.Metadata) {
				t.Fatalf("invalid metadata changed Issue: %+v", got)
			}
			for _, want := range []struct {
				path, version string
				metadata      json.RawMessage
			}{
				{"beads/plan", memory.Version, memory.Metadata},
				{"beads/plan", memoryChanged.Memory.Version, memoryChanged.Memory.Metadata},
				{"beads/work", issue.Version, issue.Metadata},
				{"beads/work", issueChanged.Issue.Version, issueChanged.Issue.Metadata},
				{"links/context", link.Link.Version, link.Link.Metadata},
				{"links/context", linkChanged.Link.Version, linkChanged.Link.Metadata},
			} {
				retained, err := s.ReadVersion(ctx, want.path, want.version)
				if err != nil {
					t.Fatalf("retained %s %s: %v", want.path, want.version, err)
				}
				var got json.RawMessage
				switch value := retained.(type) {
				case Record:
					got = value.Metadata
				case IssueRecord:
					got = value.Metadata
				case LinkRecord:
					got = value.Metadata
				}
				if !bytes.Equal(got, want.metadata) {
					t.Fatalf("retained %s %s metadata = %s, want %s", want.path, want.version, got, want.metadata)
				}
			}
		})
	}
}

func TestCommonMetadataPatchValidatesEachRawInput(t *testing.T) {
	for _, patch := range []publicops.MetadataPatch{
		{Merge: publicops.Field[json.RawMessage]{Set: true, Value: json.RawMessage(`{"x":1,"x":2}`)}},
		{Replace: publicops.Field[json.RawMessage]{Set: true, Value: json.RawMessage(`{"x":1,"x":2}`)}},
		{Set: map[string]json.RawMessage{"x": json.RawMessage(`{"a":1,"a":2}`)}, Unset: []string{"x"}},
		{Operations: publicops.Field[json.RawMessage]{Set: true, Value: json.RawMessage(`[{"op":"add","path":"/x","value":{"a":1,"a":2}},{"op":"remove","path":"/x"}]`)}},
	} {
		if err := validateCommonMetadataPatch(patch); !errors.Is(err, storage.ErrValidation) {
			t.Fatalf("raw duplicate metadata patch admitted: %+v, %v", patch, err)
		}
	}
}

func TestOrderedMetadataGraphMutationAtomicity(t *testing.T) {
	operations := func(raw string) publicops.MetadataPatch {
		return publicops.MetadataPatch{Operations: publicops.Field[json.RawMessage]{Set: true, Value: json.RawMessage(raw)}}
	}
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, o := issueExperimentOptions(t, backend)
			s, err := OpenExisting(ctx, o)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = s.Close() })
			memory, err := s.Create(ctx, CreateRequest{Path: "beads/plan", Title: "Plan", Body: "Body"})
			if err != nil {
				t.Fatal(err)
			}
			issue, err := s.CreateIssue(ctx, "beads/work", plainIssue("Work"))
			if err != nil {
				t.Fatal(err)
			}
			link, err := s.AddInformationalLink(ctx, LinkCreateRequest{Path: "links/context", SourcePath: "beads/plan", TargetPath: "beads/work", ExpectedSourceRevision: memory.Revision, Properties: map[string]any{"note": "original"}})
			if err != nil {
				t.Fatal(err)
			}
			owned := link.Source.(Record)
			before := workflowState(t, ctx, s)
			for _, raw := range []string{`[]`, `[{"op":"add","path":"/missing","value":null},{"op":"remove","path":"/missing"}]`} {
				patch := operations(raw)
				gotMemory, err := s.PatchMemory(ctx, MemoryPatchRequest{Path: "beads/plan", Actor: "editor", ExpectedRevision: owned.Revision, Metadata: patch})
				if err != nil || gotMemory.Changed || !reflect.DeepEqual(gotMemory.Memory, owned) {
					t.Fatalf("Memory metadata no-op = %+v, %v", gotMemory, err)
				}
				gotIssue, err := s.UpdateIssue(ctx, UpdateIssueRequest{Path: "beads/work", Actor: "editor", ExpectedRevision: issue.Revision, Metadata: patch})
				if err != nil || gotIssue.Changed || !reflect.DeepEqual(gotIssue.Issue, issue) {
					t.Fatalf("Issue metadata no-op = %+v, %v", gotIssue, err)
				}
				gotLink, err := s.UpdateLink(ctx, LinkUpdateRequest{Path: "links/context", Actor: "editor", ExpectedRevision: link.Link.Revision, ExpectedSourceRevision: owned.Revision, MetadataOnly: true, Metadata: patch})
				if err != nil || gotLink.Changed || !reflect.DeepEqual(gotLink.Link, link.Link) {
					t.Fatalf("Link metadata no-op = %+v, %v", gotLink, err)
				}
			}
			empty := operations(`[]`)
			if _, err := s.PatchMemory(ctx, MemoryPatchRequest{Path: "beads/plan", ExpectedRevision: memory.Revision, Metadata: empty}); !errors.Is(err, ErrConflict) {
				t.Fatalf("empty Memory update skipped stale guard: %v", err)
			}
			if _, err := s.UpdateIssue(ctx, UpdateIssueRequest{Path: "beads/work", Actor: "editor", ExpectedRevision: "stale", Metadata: empty}); !errors.Is(err, ErrConflict) {
				t.Fatalf("empty Issue update skipped stale guard: %v", err)
			}
			for _, guards := range [][2]string{{"stale", owned.Revision}, {link.Link.Revision, memory.Revision}} {
				if _, err := s.UpdateLink(ctx, LinkUpdateRequest{Path: "links/context", ExpectedRevision: guards[0], ExpectedSourceRevision: guards[1], MetadataOnly: true, Metadata: empty}); !errors.Is(err, ErrConflict) {
					t.Fatalf("empty Link update skipped guard %v: %v", guards, err)
				}
			}
			if _, err := s.PatchMemory(ctx, MemoryPatchRequest{Path: "beads/absent", Unconditional: true, Metadata: empty}); !errors.Is(err, ErrNotFound) {
				t.Fatalf("empty update skipped existence: %v", err)
			}
			if _, err := s.PatchMemory(ctx, MemoryPatchRequest{Path: "beads/plan", ExpectedRevision: owned.Revision}); !errors.Is(err, storage.ErrValidation) {
				t.Fatalf("omitted fields became an empty update: %v", err)
			}
			if !reflect.DeepEqual(before, workflowState(t, ctx, s)) {
				t.Fatal("no-op or refusal changed payload, revision, attribution, owned state or events")
			}

			// Properties and ordered metadata share one transaction. Either side's
			// invalid final state must leave both the Resource and its owner alone.
			for _, pair := range [][2]string{
				{`[{"op":"replace","path":"/body","value":"changed"}]`, `[{"op":"add","path":"/x","value":1},{"op":"remove","path":"/absent"}]`},
				{`[{"op":"remove","path":"/body"}]`, `[{"op":"add","path":"/x","value":1}]`},
			} {
				if _, err := s.PatchMemoryProperties(ctx, MemoryPropertiesPatchRequest{Path: "beads/plan", ExpectedRevision: owned.Revision, Patch: []byte(pair[0]), Metadata: operations(pair[1])}); !errors.Is(err, storage.ErrValidation) {
					t.Fatalf("invalid combined update = %v", err)
				}
			}
			title := "must roll back"
			if _, err := s.UpdateIssue(ctx, UpdateIssueRequest{Path: "beads/work", Actor: "editor", ExpectedRevision: issue.Revision, Title: &title, Metadata: operations(`[{"op":"replace","path":"","value":null}]`)}); !errors.Is(err, storage.ErrValidation) {
				t.Fatalf("invalid Issue metadata did not refuse sibling title: %v", err)
			}
			if _, err := s.PatchLinkProperties(ctx, LinkPropertiesPatchRequest{Path: "links/context", ExpectedRevision: link.Link.Revision, ExpectedSourceRevision: owned.Revision, Patch: []byte(`[{"op":"replace","path":"/note","value":"changed"}]`), Metadata: operations(`[{"op":"remove","path":"/absent"}]`)}); !errors.Is(err, storage.ErrValidation) {
				t.Fatalf("invalid Link metadata did not refuse sibling note: %v", err)
			}
			if !reflect.DeepEqual(before, workflowState(t, ctx, s)) {
				t.Fatal("failed combined update leaked payload, revision or owner change")
			}

			updated, err := s.PatchMemoryProperties(ctx, MemoryPropertiesPatchRequest{Path: "beads/plan", Actor: "editor", ExpectedRevision: owned.Revision, Patch: []byte(`[{"op":"replace","path":"/body","value":"changed"}]`), Metadata: empty})
			if err != nil || !updated.Changed || updated.Memory.Properties.Body != "changed" || string(updated.Memory.Metadata) != `{}` {
				t.Fatalf("effectful properties plus empty metadata = %+v, %v", updated, err)
			}
			updatedLink, err := s.UpdateLink(ctx, LinkUpdateRequest{Path: "links/context", Actor: "editor", ExpectedRevision: link.Link.Revision, ExpectedSourceRevision: updated.Memory.Revision, MetadataOnly: true, Metadata: operations(`[{"op":"add","path":"/nested","value":{"list":[1]}},{"op":"add","path":"/nested/list/-","value":2}]`)})
			if err != nil || !updatedLink.Changed || string(updatedLink.Link.Metadata) != `{"nested":{"list":[1,2]}}` || updatedLink.Source.(Record).Revision == updated.Memory.Revision {
				t.Fatalf("ordered Link metadata failed owner transition = %+v, %v", updatedLink, err)
			}
			updatedIssue, err := s.UpdateIssue(ctx, UpdateIssueRequest{Path: "beads/work", Actor: "editor", ExpectedRevision: issue.Revision, Metadata: operations(`[{"op":"add","path":"/reviewed","value":true}]`)})
			if err != nil || !updatedIssue.Changed || string(updatedIssue.Issue.Metadata) != `{"reviewed":true}` || updatedIssue.Issue.Revision == issue.Revision {
				t.Fatalf("ordered Issue metadata did not reach native writer = %+v, %v", updatedIssue, err)
			}

			deleted, err := s.Create(ctx, CreateRequest{Path: "beads/deleted", Title: "Gone", Body: "Gone"})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := s.DeleteMemory(ctx, MemoryDeleteRequest{Path: "beads/deleted", ExpectedRevision: deleted.Revision}); err != nil {
				t.Fatal(err)
			}
			before = workflowState(t, ctx, s)
			if _, err := s.PatchMemory(ctx, MemoryPatchRequest{Path: "beads/deleted", Unconditional: true, Metadata: empty}); !errors.Is(err, ErrNotFound) {
				t.Fatalf("empty update skipped liveness: %v", err)
			}
			if !reflect.DeepEqual(before, workflowState(t, ctx, s)) {
				t.Fatal("empty update changed deleted identity")
			}
		})
	}
}

// Metadata-only Issue edits must obey the same post-write acquisition budget
// as notes edits and Memory/Link metadata patches. The original Issue row,
// native History, graph mapping and sibling title all roll back together.
func TestIssueMetadataPatchPreservesCurrentReadBudget(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, o := issueExperimentOptions(t, backend)
			s, err := OpenExisting(ctx, o)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = s.Close() })
			original, err := s.CreateIssue(ctx, "beads/work", plainIssue("Original"))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := s.Create(ctx, CreateRequest{Path: "beads/context", Body: strings.Repeat("x", 7<<20)}); err != nil {
				t.Fatal(err)
			}
			before := workflowState(t, ctx, s)
			title := "Must roll back"
			metadata := json.RawMessage(`{"note":"` + strings.Repeat("m", 1536<<10) + `"}`)
			got, err := s.UpdateIssue(ctx, UpdateIssueRequest{
				Path: "beads/work", Actor: "editor", ExpectedRevision: original.Revision,
				Title: &title, Metadata: publicops.MetadataPatch{Merge: publicops.Field[json.RawMessage]{Set: true, Value: metadata}},
			})
			if !errors.Is(err, ErrLimitExceeded) || !reflect.DeepEqual(got, IssueMutationResult{}) {
				t.Fatalf("over-budget Issue metadata edit must refuse with zero result: %+v, %v", got, err)
			}
			if !reflect.DeepEqual(before, workflowState(t, ctx, s)) {
				t.Fatal("over-budget metadata edit leaked Issue row or History")
			}
			if current, err := s.ShowIssue(ctx, "beads/work"); err != nil || !reflect.DeepEqual(current, original) {
				t.Fatalf("Issue changed after metadata refusal: %+v, %v", current, err)
			}
		})
	}
}
