//go:build cgo

package graphstore

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	publicops "github.com/steveyegge/beads/issueops"
)

func TestHistoricalMemoryUpdateComposition(t *testing.T) {
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			ctx, _, s := memoryDeleteStore(t, engine)
			original, err := s.Create(ctx, CreateRequest{Path: "beads/subject", Title: "Original", Body: "R1", Actor: "original", Metadata: json.RawMessage(`{"old":null,"shared":"old"}`)})
			if err != nil {
				t.Fatal(err)
			}
			later, err := s.PatchMemory(ctx, MemoryPatchRequest{Path: "beads/subject", Body: historicalBody("R2"), Unconditional: true, Actor: "later", Metadata: publicops.MetadataPatch{Set: map[string]json.RawMessage{"new": json.RawMessage(`true`), "shared": json.RawMessage(`"new"`)}, Unset: []string{"old"}}})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := s.Create(ctx, CreateRequest{Path: "beads/target", Title: "Target", Body: "target"}); err != nil {
				t.Fatal(err)
			}
			link, err := s.AddInformationalLink(ctx, LinkCreateRequest{Path: "links/owned", SourcePath: "beads/subject", TargetPath: "beads/target", ExpectedSourceRevision: later.Memory.Revision})
			if err != nil {
				t.Fatal(err)
			}
			current := link.Source.(Record)
			result, err := s.PatchMemory(ctx, MemoryPatchRequest{Path: "beads/subject", FromRevision: original.Revision, ExpectedRevision: current.Revision, Actor: "historical-editor", Metadata: publicops.MetadataPatch{Set: map[string]json.RawMessage{"explicit": json.RawMessage(`7`)}}})
			if err != nil {
				t.Fatal(err)
			}
			if !result.Changed || result.Memory.Revision == current.Revision || result.Memory.Properties != original.Properties || !reflect.DeepEqual(result.Memory.Owned, current.Owned) {
				t.Fatalf("wrong composed result: %+v", result)
			}
			if string(result.Memory.Metadata) != `{"explicit":7,"old":null,"shared":"old"}` {
				t.Fatalf("wrong metadata: %s", result.Memory.Metadata)
			}
			h := result.Memory.Attribution.HistoricalUpdate
			if h == nil || h.FromRevision != original.Revision || h.PreviousRevision != current.Revision || result.Memory.Attribution.Actor != "historical-editor" {
				t.Fatalf("wrong context: %+v", result.Memory.Attribution)
			}
			for _, record := range []Record{original, current, result.Memory} {
				got, err := s.ReadVersion(ctx, "beads/subject", record.Revision)
				if err != nil || !reflect.DeepEqual(got, record) {
					t.Fatalf("retained changed: %+v %v", got, err)
				}
			}
			// Source-only and ordinary no-ops retain an already historical envelope.
			before := workflowState(t, ctx, s)
			for _, request := range []MemoryPatchRequest{{Path: "beads/subject", FromRevision: result.Memory.Revision, ExpectedRevision: result.Memory.Revision, Actor: "different"}, {Path: "beads/subject", PropertiesProvided: true, ExpectedRevision: result.Memory.Revision, Actor: "different"}} {
				noop, err := s.PatchMemory(ctx, request)
				if err != nil || noop.Changed || !reflect.DeepEqual(noop.Memory, result.Memory) {
					t.Fatalf("no-op rewrote context: %+v %v", noop, err)
				}
			}
			if !reflect.DeepEqual(before, workflowState(t, ctx, s)) {
				t.Fatal("no-op changed store")
			}
			// Explicit edits apply to source, but equality is against the real current head.
			body := "R1"
			equal, err := s.PatchMemory(ctx, MemoryPatchRequest{Path: "beads/subject", FromRevision: later.Memory.Revision, Body: &body, ExpectedRevision: result.Memory.Revision, Metadata: publicops.MetadataPatch{Unset: []string{"new"}, Set: map[string]json.RawMessage{"old": json.RawMessage(`null`), "shared": json.RawMessage(`"old"`), "explicit": json.RawMessage(`7`)}}})
			if err != nil || equal.Changed || !reflect.DeepEqual(equal.Memory, result.Memory) {
				t.Fatalf("patch-to-current should no-op: %+v %v", equal, err)
			}
		})
	}
}

func historicalBody(value string) *string { return &value }

func TestHistoricalMemoryUpdateRefusalsAndRollback(t *testing.T) {
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			ctx, _, s := memoryDeleteStore(t, engine)
			first, err := s.Create(ctx, CreateRequest{Path: "beads/subject", Title: "Subject", Body: "first"})
			if err != nil {
				t.Fatal(err)
			}
			current, err := s.PatchMemory(ctx, MemoryPatchRequest{Path: "beads/subject", Body: historicalBody("second"), Unconditional: true})
			if err != nil {
				t.Fatal(err)
			}
			other, err := s.Create(ctx, CreateRequest{Path: "beads/other", Title: "Other", Body: "other"})
			if err != nil {
				t.Fatal(err)
			}
			before := workflowState(t, ctx, s)
			for _, tc := range []struct {
				name, source, guard string
				want                error
			}{
				{"foreign revision", other.Revision, current.Memory.Revision, ErrVersionUnknown},
				{"unknown source", "unavailable-source", current.Memory.Revision, ErrVersionUnknown},
				{"stale guard", first.Revision, first.Revision, ErrConflict},
			} {
				t.Run(tc.name, func(t *testing.T) {
					_, err := s.PatchMemory(ctx, MemoryPatchRequest{Path: "beads/subject", FromRevision: tc.source, ExpectedRevision: tc.guard})
					if !errors.Is(err, tc.want) {
						t.Fatalf("want %v got %v", tc.want, err)
					}
					if !reflect.DeepEqual(before, workflowState(t, ctx, s)) {
						t.Fatal("refusal wrote")
					}
				})
			}
			injected := errors.New("historical rollback")
			for _, stage := range []string{"coordination", "memory-payload", "source-catalog", "source-retained"} {
				t.Run(stage, func(t *testing.T) {
					s.afterWrite = func(at string) error {
						if at == stage {
							return injected
						}
						return nil
					}
					_, err := s.PatchMemory(ctx, MemoryPatchRequest{Path: "beads/subject", FromRevision: first.Revision, Unconditional: true})
					s.afterWrite = nil
					if !errors.Is(err, injected) {
						t.Fatalf("missing failure: %v", err)
					}
					if !reflect.DeepEqual(before, workflowState(t, ctx, s)) {
						t.Fatal("partial rollback")
					}
				})
			}
		})
	}
}

func TestHistoricalMemoryUpdateCurrentTypeAndProtectedEvaluation(t *testing.T) {
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			ctx, _, s := memoryDeleteStore(t, engine)
			original, err := s.Create(ctx, CreateRequest{Path: "beads/evaluation", Title: "Before Evaluation", Body: "historical content"})
			if err != nil {
				t.Fatal(err)
			}
			subject, err := s.Create(ctx, CreateRequest{Path: "beads/subject", Title: "Assessed subject"})
			if err != nil {
				t.Fatal(err)
			}
			subjects, err := json.Marshal(map[string]any{"subjects": []any{map[string]string{"uri": subject.ID, "revision": subject.Revision}}, "body": "current content"})
			if err != nil {
				t.Fatal(err)
			}
			current, err := s.PatchMemory(ctx, MemoryPatchRequest{Path: "beads/evaluation", TypeURL: s.ScopeURL() + "types/evaluation", PropertiesMerge: subjects, Unconditional: true})
			if err != nil {
				t.Fatal(err)
			}
			result, err := s.PatchMemory(ctx, MemoryPatchRequest{Path: "beads/evaluation", FromRevision: original.Revision, PropertiesMerge: json.RawMessage(`{"title":"Edited historical title"}`), ExpectedRevision: current.Memory.Revision})
			if err != nil {
				t.Fatal(err)
			}
			if !result.Changed || result.Memory.Type != current.Memory.Type || result.Memory.Properties.Title != "Edited historical title" || result.Memory.Properties.Body != original.Properties.Body {
				t.Fatalf("wrong final Type/content: %+v", result)
			}
			if err := preserveEvaluationSubjectGroup(current.Memory.Properties, result.Memory.Properties); err != nil {
				t.Fatalf("lost protected group: %v", err)
			}
			before := workflowState(t, ctx, s)
			_, err = s.PatchMemory(ctx, MemoryPatchRequest{Path: "beads/evaluation", FromRevision: original.Revision, PropertiesMerge: json.RawMessage(`{"subjects":[]}`), Unconditional: true})
			if err == nil || !reflect.DeepEqual(before, workflowState(t, ctx, s)) {
				t.Fatalf("protected group edit accepted or wrote: %v", err)
			}
			_, err = s.PatchMemory(ctx, MemoryPatchRequest{Path: "beads/evaluation", FromRevision: original.Revision, TypeURL: original.Type, Unconditional: true})
			if err == nil || !reflect.DeepEqual(before, workflowState(t, ctx, s)) {
				t.Fatalf("historical retype accepted or wrote: %v", err)
			}
		})
	}
}
