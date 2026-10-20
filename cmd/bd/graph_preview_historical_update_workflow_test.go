//go:build cgo

package main

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func TestGraphPreviewHistoricalUpdateWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := graphListWorkspace(t, bd, engine, "https://example.invalid/historical-update/")
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			original := graphMixedResult[graphstore.Record](t, call("remember", "first", "--id", "beads/subject", "--metadata", `{"old":null}`))
			current := graphMixedResult[graphstore.MemoryMutationResult](t, call("update", "subject", "--properties", `{"body":"second"}`, "--replace-metadata", `{"new":true}`))
			applied := graphMixedResult[graphstore.MemoryMutationResult](t, call("update", "subject", "--from-revision", original.Revision, "--if-revision", current.Memory.Revision))
			if !applied.Changed || applied.Memory.Properties != original.Properties || string(applied.Memory.Metadata) != string(original.Metadata) || applied.Memory.Attribution.HistoricalUpdate == nil || applied.Memory.Attribution.HistoricalUpdate.PreviousRevision != current.Memory.Revision {
				t.Fatalf("wrong historical result: %+v", applied)
			}
			noop := graphMixedResult[graphstore.MemoryMutationResult](t, call("update", "subject", "--from_revision", original.Revision))
			if noop.Changed || !reflect.DeepEqual(noop.Memory, applied.Memory) {
				t.Fatalf("alias/noop rewrote envelope: %+v", noop)
			}
			for _, tc := range []struct {
				code string
				args []string
			}{
				{"revision_conflict", []string{"update", "subject", "--from-revision", original.Revision, "--if-revision", current.Memory.Revision}},
				{"revision_unknown", []string{"update", "subject", "--from-revision", "unknown"}},
				{"invalid_selector", []string{"update", "subject", "--from-revision", ""}},
				{"invalid_selector", []string{"update", "subject@current", "--from-revision", original.Revision}},
				{"capability_unavailable", []string{"update", "links/not-created", "--from-revision", original.Revision}},
			} {
				graphPolicyCLI(t, bd, work, home, nil, tc.code, append(tc.args, "--json")...)
			}
			if got := graphMixedResult[graphstore.Record](t, call("show", "subject")); !reflect.DeepEqual(got, applied.Memory) {
				t.Fatal("refusal changed target")
			}
			if got := graphMixedResult[graphstore.Record](t, call("show", "subject", "--version", original.Revision)); !reflect.DeepEqual(got, original) {
				t.Fatal("original history changed")
			}
		})
	}
}

func TestGraphPreviewHistoricalRecoveryWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := graphListWorkspace(t, bd, engine, "https://example.invalid/historical-recovery/")
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			first := graphMixedResult[graphstore.Record](t, call("remember", "first", "--id", "beads/recover"))
			last := graphMixedResult[graphstore.MemoryMutationResult](t, call("update", "recover", "--properties", `{"body":"last"}`))
			deleted := graphMixedResult[graphstore.MemoryDeleteResult](t, call("delete", "recover", "--force", "--if-revision", last.Memory.Revision))
			if deleted.Terminal == nil || deleted.Terminal.Revision == last.Memory.Revision {
				t.Fatalf("missing actual D: %+v", deleted)
			}
			type historyResult struct {
				CurrentState *graphstore.MemoryTerminalState `json:"currentState"`
				Terminal     []map[string]any                `json:"terminal_versions"`
				Versions     []map[string]any                `json:"versions"`
			}
			history := graphMixedResult[historyResult](t, call("versions", "recover"))
			if !reflect.DeepEqual(history.CurrentState, deleted.Terminal) || len(history.Terminal) != 1 || history.Terminal[0]["version"] != deleted.Terminal.Revision {
				t.Fatalf("terminal carrier: %+v", history)
			}
			graphPolicyCLI(t, bd, work, home, nil, "gone", "show", "recover", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "gone", "show", "recover", "--version", deleted.Terminal.Revision, "--json")
			graphPolicyCLI(t, bd, work, home, nil, "revision_conflict", "update", "recover", "--from-revision", first.Revision, "--if-revision", last.Memory.Revision, "--json")
			recovered := graphMixedResult[graphstore.MemoryMutationResult](t, call("update", "recover", "--from-revision", first.Revision, "--if-revision", deleted.Terminal.Revision, "--properties", `{"body":"explicit after source"}`))
			if !recovered.Changed || recovered.Memory.ID != first.ID || recovered.Memory.Properties.Body != "explicit after source" || recovered.Memory.Attribution.Inception != "revert" || recovered.Memory.Attribution.HistoricalUpdate == nil || recovered.Memory.Attribution.HistoricalUpdate.PreviousRevision != deleted.Terminal.Revision {
				t.Fatalf("wrong recovery: %+v", recovered)
			}
			if got := graphMixedResult[graphstore.Record](t, call("show", "recover", "--version", first.Revision)); !reflect.DeepEqual(got, first) {
				t.Fatal("recovery changed retained source")
			}
			if got := graphMixedResult[graphstore.Record](t, call("show", "recover")); !reflect.DeepEqual(got, recovered.Memory) {
				t.Fatal("recovery did not install result")
			}
			graphPolicyCLI(t, bd, work, home, nil, "gone", "show", "recover", "--version", deleted.Terminal.Revision, "--json")
		})
	}
}

func TestGraphPreviewHistoricalOrderedPatchBase(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := graphListWorkspace(t, bd, engine, "https://example.invalid/historical-patch/")
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			original := graphMixedResult[graphstore.Record](t, call("remember", "Original", "--id", "beads/base", "--metadata", `{"source":true}`))
			current := graphMixedResult[graphstore.MemoryMutationResult](t, call("update", "base", "--properties", `{"body":"current"}`, "--replace-metadata", `{"current":true}`))
			operations, err := json.Marshal([]map[string]any{{"op": "replace", "path": "/title", "value": "Intermediate"}, {"op": "replace", "path": "/title", "value": "Patched"}})
			if err != nil {
				t.Fatal(err)
			}
			patched := graphMixedResult[graphstore.MemoryMutationResult](t, call("update", "base", "--from-revision", original.Revision, "--if-revision", current.Memory.Revision, "--patch", string(operations)))
			if !patched.Changed || patched.Memory.Properties.Body != original.Properties.Body || patched.Memory.Properties.Title != "Patched" || string(patched.Memory.Metadata) != string(original.Metadata) {
				t.Fatalf("ordered patch ignored selected base: %+v", patched)
			}
			replaced := graphMixedResult[graphstore.MemoryMutationResult](t, call("update", "base", "--from-revision", original.Revision, "--replace-properties", `{"title":"Replacement","body":"full"}`))
			if !replaced.Changed || replaced.Memory.Properties.Title != "Replacement" || replaced.Memory.Properties.Body != "full" || string(replaced.Memory.Metadata) != string(original.Metadata) {
				t.Fatalf("replacement ignored source metadata/base ordering: %+v", replaced)
			}
		})
	}
}

func TestGraphPreviewHistoricalNativeIssueWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := graphListWorkspace(t, bd, engine, "https://example.invalid/native-history/")
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			original := graphMixedResult[graphstore.IssueRecord](t, call("create", "Original", "--id", "beads/subject", "--description", "original body", "--notes", "historical notes", "--metadata", `{"old":null}`))
			current := graphMixedResult[graphstore.IssueMutationResult](t, call("update", "subject", "--properties", `{"title":"Current","description":"current body"}`, "--replace-metadata", `{"current":true}`, "--notes", "current notes", "--force"))
			applied := graphMixedResult[graphstore.IssueMutationResult](t, call("update", "subject", "--from-revision", original.Revision, "--if-revision", current.Issue.Revision, "--properties", `{"title":"Explicit"}`, "--set-metadata", "explicit=7", "--force"))
			if !applied.Changed || applied.Issue.Properties.Title != "Explicit" || applied.Issue.Properties.Description != original.Properties.Description || applied.Issue.Properties.Notes != original.Properties.Notes || string(applied.Issue.Metadata) != `{"explicit":7,"old":null}` || applied.Issue.Attribution.HistoricalUpdate == nil || applied.Issue.Attribution.HistoricalUpdate.PreviousRevision != current.Issue.Revision || applied.Issue.Attribution.Inception != original.Attribution.Inception {
				t.Fatalf("wrong historical native result: %+v", applied)
			}
			noop := graphMixedResult[graphstore.IssueMutationResult](t, call("update", "subject", "--from_revision", applied.Issue.Revision))
			if noop.Changed || !reflect.DeepEqual(noop.Issue, applied.Issue) {
				t.Fatal("native historical no-op rewrote envelope")
			}
			for _, tc := range []struct {
				code string
				args []string
			}{
				{"revision_conflict", []string{"update", "subject", "--from-revision", original.Revision, "--if-revision", current.Issue.Revision}},
				{"revision_unknown", []string{"update", "subject", "--from-revision", "unknown"}},
				{"capability_unavailable", []string{"update", "subject", "--from-revision", original.Revision, "--patch", `[{"op":"replace","path":"/title","value":"Patch"}]`}},
			} {
				graphPolicyCLI(t, bd, work, home, nil, tc.code, append(tc.args, "--json")...)
			}
			got := graphMixedResult[graphstore.IssueRecord](t, call("show", "subject", "--format", "graph-json"))
			if !reflect.DeepEqual(got, applied.Issue) {
				t.Fatal("refusal changed native destination")
			}
			exact := graphMixedResult[graphstore.IssueRecord](t, call("show", "subject", "--version", applied.Issue.Revision))
			want := applied.Issue
			properties := *want.Properties
			properties.ContentHash = ""
			properties.RowVersion = 0
			want.Properties = &properties
			if !reflect.DeepEqual(exact, want) {
				t.Fatal("native historical context not retained")
			}
		})
	}
}
