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
			graphPolicyCLI(t, bd, work, home, nil, "not_found", "update", "subject", "--from-revision", original.Revision, "--json")
			if got := graphMixedResult[graphstore.Record](t, call("show", original.ID)); !reflect.DeepEqual(got, original) {
				t.Fatal("historical shorthand refusal changed Memory")
			}
			current := graphMixedResult[graphstore.MemoryMutationResult](t, call("update", "beads/subject", "--properties", `{"body":"second"}`, "--replace-metadata", `{"new":true}`))
			applied := graphMixedResult[graphstore.MemoryMutationResult](t, call("update", "beads/subject", "--from-revision", original.Revision, "--if-revision", current.Memory.Revision))
			if !applied.Changed || applied.Memory.Properties != original.Properties || string(applied.Memory.Metadata) != string(original.Metadata) || applied.Memory.Attribution.HistoricalUpdate == nil || applied.Memory.Attribution.HistoricalUpdate.PreviousRevision != current.Memory.Revision {
				t.Fatalf("wrong historical result: %+v", applied)
			}
			noop := graphMixedResult[graphstore.MemoryMutationResult](t, call("update", "beads/subject", "--from_revision", original.Revision))
			if noop.Changed || !reflect.DeepEqual(noop.Memory, applied.Memory) {
				t.Fatalf("alias/noop rewrote envelope: %+v", noop)
			}
			for _, tc := range []struct {
				code string
				args []string
			}{
				{"revision_conflict", []string{"update", "beads/subject", "--from-revision", original.Revision, "--if-revision", current.Memory.Revision}},
				{"revision_unknown", []string{"update", "beads/subject", "--from-revision", "unknown"}},
				{"invalid_selector", []string{"update", "beads/subject", "--from-revision", ""}},
				{"invalid_selector", []string{"update", "subject@current", "--from-revision", original.Revision}},
				{"capability_unavailable", []string{"update", "https://example.invalid/historical-update/links/not-created", "--from-revision", original.Revision}},
			} {
				graphPolicyCLI(t, bd, work, home, nil, tc.code, append(tc.args, "--json")...)
			}
			if got := graphMixedResult[graphstore.Record](t, call("show", original.ID)); !reflect.DeepEqual(got, applied.Memory) {
				t.Fatal("refusal changed target")
			}
			if got := graphMixedResult[graphstore.Record](t, call("show", original.ID, "--version", original.Revision)); !reflect.DeepEqual(got, original) {
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
			last := graphMixedResult[graphstore.MemoryMutationResult](t, call("update", "beads/recover", "--properties", `{"body":"last"}`))
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
			graphPolicyCLI(t, bd, work, home, nil, "gone", "show", first.ID, "--json")
			graphPolicyCLI(t, bd, work, home, nil, "gone", "show", first.ID, "--version", deleted.Terminal.Revision, "--json")
			graphPolicyCLI(t, bd, work, home, nil, "revision_conflict", "update", "beads/recover", "--from-revision", first.Revision, "--if-revision", last.Memory.Revision, "--json")
			recovered := graphMixedResult[graphstore.MemoryMutationResult](t, call("update", "beads/recover", "--from-revision", first.Revision, "--if-revision", deleted.Terminal.Revision, "--properties", `{"body":"explicit after source"}`))
			if !recovered.Changed || recovered.Memory.ID != first.ID || recovered.Memory.Properties.Body != "explicit after source" || recovered.Memory.Attribution.Inception != "revert" || recovered.Memory.Attribution.HistoricalUpdate == nil || recovered.Memory.Attribution.HistoricalUpdate.PreviousRevision != deleted.Terminal.Revision {
				t.Fatalf("wrong recovery: %+v", recovered)
			}
			if got := graphMixedResult[graphstore.Record](t, call("show", first.ID, "--version", first.Revision)); !reflect.DeepEqual(got, first) {
				t.Fatal("recovery changed retained source")
			}
			if got := graphMixedResult[graphstore.Record](t, call("show", first.ID)); !reflect.DeepEqual(got, recovered.Memory) {
				t.Fatal("recovery did not install result")
			}
			graphPolicyCLI(t, bd, work, home, nil, "gone", "show", first.ID, "--version", deleted.Terminal.Revision, "--json")
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
			current := graphMixedResult[graphstore.MemoryMutationResult](t, call("update", "beads/base", "--properties", `{"body":"current"}`, "--replace-metadata", `{"current":true}`))
			operations, err := json.Marshal([]map[string]any{{"op": "replace", "path": "/title", "value": "Intermediate"}, {"op": "replace", "path": "/title", "value": "Patched"}})
			if err != nil {
				t.Fatal(err)
			}
			patched := graphMixedResult[graphstore.MemoryMutationResult](t, call("update", "beads/base", "--from-revision", original.Revision, "--if-revision", current.Memory.Revision, "--patch", string(operations)))
			if !patched.Changed || patched.Memory.Properties.Body != original.Properties.Body || patched.Memory.Properties.Title != "Patched" || string(patched.Memory.Metadata) != string(original.Metadata) {
				t.Fatalf("ordered patch ignored selected base: %+v", patched)
			}
			replaced := graphMixedResult[graphstore.MemoryMutationResult](t, call("update", "beads/base", "--from-revision", original.Revision, "--replace-properties", `{"title":"Replacement","body":"full"}`))
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
			original := graphMixedResult[graphstore.IssueRecord](t, call("create", "Original", "--id", "subject", "--force", "--description", "original body", "--notes", "historical notes", "--metadata", `{"old":null}`))
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
			} {
				graphPolicyCLI(t, bd, work, home, nil, tc.code, append(tc.args, "--json")...)
			}
			got := graphMixedResult[graphstore.IssueRecord](t, call("show", original.ID, "--format", "graph-json"))
			if !reflect.DeepEqual(got, applied.Issue) {
				t.Fatal("refusal changed native destination")
			}
			exact := graphMixedResult[graphstore.IssueRecord](t, call("show", original.ID, "--version", applied.Issue.Revision))
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

func TestGraphPreviewHistoricalNativeOrderedPatch(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := graphListWorkspace(t, bd, engine, "https://example.invalid/native-ordered/")
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			original := graphMixedResult[graphstore.IssueRecord](t, call("create", "Original", "--id", "subject", "--force", "--description", "historical body", "--metadata", `{"source":true}`))
			current := graphMixedResult[graphstore.IssueMutationResult](t, call("update", "subject", "--properties", `{"description":"current body"}`, "--replace-metadata", `{"current":true}`))
			patched := graphMixedResult[graphstore.IssueMutationResult](t, call("update", "subject", "--from-revision", original.Revision, "--if-revision", current.Issue.Revision, "--patch", `[{"op":"replace","path":"/design","value":"intermediate"},{"op":"replace","path":"/design","value":"explicit"}]`))
			if !patched.Changed || patched.Issue.Properties.Design != "explicit" || patched.Issue.Properties.Description != "historical body" || string(patched.Issue.Metadata) != string(original.Metadata) {
				t.Fatalf("native ordered CLI ignored source: %+v", patched)
			}
			replaced := graphMixedResult[graphstore.IssueMutationResult](t, call("update", "subject", "--from-revision", original.Revision, "--replace-properties", `{"title":"Replacement","priority":0}`))
			if !replaced.Changed || replaced.Issue.Properties.Title != "Replacement" || replaced.Issue.Properties.Priority != 0 || replaced.Issue.Properties.Description != "" || string(replaced.Issue.Metadata) != string(original.Metadata) {
				t.Fatalf("native replacement wrong: %+v", replaced)
			}
			for _, raw := range []string{`[{"op":"remove","path":"/title"}]`, `[{"op":"add","path":"/labels","value":["old"]}]`, `[{"op":"replace","path":"/priority","value":null}]`} {
				graphPolicyCLI(t, bd, work, home, nil, "invalid_properties", "update", "subject", "--from-revision", original.Revision, "--patch", raw, "--json")
			}
			got := graphMixedResult[graphstore.IssueRecord](t, call("show", original.ID, "--format", "graph-json"))
			if !reflect.DeepEqual(got, replaced.Issue) {
				t.Fatal("native ordered refusal changed state")
			}
		})
	}
}

func TestGraphPreviewHistoricalNativeRecovery(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := graphListWorkspace(t, bd, engine, "https://example.invalid/native-recovery/")
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			refuse := func(code string, args ...string) {
				t.Helper()
				graphPolicyCLI(t, bd, work, home, nil, code, append(args, "--json")...)
			}
			source := graphMixedResult[graphstore.IssueRecord](t, call("create", "Original", "--id", "subject", "--force", "--description", "historical body", "--labels", "old", "--metadata", `{"source":true}`))
			current := graphMixedResult[graphstore.IssueMutationResult](t, call("update", "subject", "--properties", `{"description":"last current"}`, "--replace-metadata", `{"current":true}`))
			deleted := graphMixedResult[graphstore.IssueDeleteResult](t, call("delete", "subject", "--force", "--if-revision", current.Issue.Revision))
			if deleted.Terminal == nil || deleted.Terminal.Revision == current.Issue.Revision {
				t.Fatal("native delete omitted actual D")
			}
			type historyResult struct {
				CurrentState *graphstore.BeadTerminalState `json:"currentState"`
				Terminal     []map[string]any              `json:"terminal_versions"`
				Versions     []map[string]any              `json:"versions"`
			}
			assertCanonicalHistory := func(want historyResult) {
				t.Helper()
				raw := graphPolicyCLI(t, bd, work, home, nil, "", "versions", "subject", "--output-format", "json")
				var envelope struct {
					SchemaVersion int `json:"schema_version"`
					Data          struct {
						CurrentState *struct{ ID, Revision, PreviousVersion, State string } `json:"currentState"`
						Terminal     []map[string]any                                       `json:"terminal_versions"`
						Versions     []map[string]any                                       `json:"versions"`
					} `json:"data"`
				}
				if err := json.Unmarshal([]byte(raw), &envelope); err != nil || envelope.SchemaVersion != 2 || !reflect.DeepEqual(envelope.Data.Terminal, want.Terminal) || !reflect.DeepEqual(envelope.Data.Versions, want.Versions) {
					t.Fatalf("canonical native history lost terminal/ordinal context: decode=%v got=%s want=%+v", err, raw, want)
				}
				state := envelope.Data.CurrentState
				if (state == nil) != (want.CurrentState == nil) || (state != nil && (state.ID != want.CurrentState.ID || state.Revision != want.CurrentState.Revision || state.PreviousVersion != want.CurrentState.PreviousVersion || state.State != "deleted")) {
					t.Fatalf("canonical history lost current deletion identity: %s", raw)
				}
			}
			history := graphMixedResult[historyResult](t, call("versions", "subject"))
			if !reflect.DeepEqual(history.CurrentState, deleted.Terminal) || len(history.Terminal) != 1 || len(history.Versions) != 2 {
				t.Fatalf("native history carrier: %+v", history)
			}
			assertCanonicalHistory(history)
			refuse("gone", "show", source.ID)
			refuse("gone", "show", source.ID, "--version", deleted.Terminal.Revision)
			refuse("revision_conflict", "update", "subject", "--from-revision", source.Revision, "--if-revision", current.Issue.Revision)
			refuse("gone", "update", "subject", "--from-revision", deleted.Terminal.Revision, "--if-revision", deleted.Terminal.Revision)
			recovered := graphMixedResult[graphstore.IssueMutationResult](t, call("update", "subject", "--from-revision", source.Revision, "--if-revision", deleted.Terminal.Revision, "--patch", `[{"op":"replace","path":"/title","value":"Recovered"}]`, "--set-metadata", "explicit=7", "--actor", "restorer"))
			h := recovered.Issue.Attribution.HistoricalUpdate
			if !recovered.Changed || recovered.Issue.ID != source.ID || recovered.Issue.Type != source.Type || recovered.Issue.Properties.ID != source.Properties.ID || recovered.Issue.Properties.Title != "Recovered" || recovered.Issue.Properties.Description != "historical body" || string(recovered.Issue.Metadata) != `{"explicit":7,"source":true}` || h == nil || h.FromRevision != source.Revision || h.PreviousRevision != deleted.Terminal.Revision || recovered.Issue.Attribution.Actor != "restorer" || recovered.Issue.Attribution.Inception != "revert" || len(recovered.Issue.Properties.Labels) != 0 || !recovered.Issue.Properties.CreatedAt.Equal(current.Issue.Properties.CreatedAt) {
				t.Fatalf("native CLI recovered wrong state: %+v", recovered)
			}
			got := graphMixedResult[graphstore.IssueRecord](t, call("show", source.ID, "--format", "graph-json"))
			if !reflect.DeepEqual(got, recovered.Issue) {
				t.Fatal("native recovery not visible to fresh process")
			}
			for _, selection := range [][]string{nil, {"--version", recovered.Issue.Revision}} {
				args := append([]string{"show", source.ID, "--output-format", "json"}, selection...)
				raw := graphPolicyCLI(t, bd, work, home, nil, "", args...)
				var envelope struct {
					Data struct {
						ID, Revision string
						Metadata     json.RawMessage
					} `json:"data"`
				}
				if err := json.Unmarshal([]byte(raw), &envelope); err != nil || envelope.Data.ID != recovered.Issue.ID || envelope.Data.Revision != recovered.Issue.Revision || string(envelope.Data.Metadata) != string(recovered.Issue.Metadata) {
					t.Fatalf("canonical native recovered metadata changed: decode=%v got=%s", err, raw)
				}
			}
			before := call("show", source.ID, "--format", "graph-json")
			refuse("revision_conflict", "update", "subject", "--from-revision", source.Revision, "--if-revision", deleted.Terminal.Revision)
			if after := call("show", source.ID, "--format", "graph-json"); after != before {
				t.Fatal("stale native recovery replay changed state")
			}
			history = graphMixedResult[historyResult](t, call("versions", "subject"))
			if history.CurrentState != nil || len(history.Terminal) != 1 || len(history.Versions) != 3 {
				t.Fatalf("native recovered history lost D: %+v", history)
			}
			assertCanonicalHistory(history)
			secondDelete := graphMixedResult[graphstore.IssueDeleteResult](t, call("delete", "subject", "--force"))
			equal := graphMixedResult[graphstore.IssueMutationResult](t, call("update", "subject", "--from_revision", recovered.Issue.Revision))
			if !equal.Changed || equal.Issue.Revision == recovered.Issue.Revision || equal.Issue.Attribution.HistoricalUpdate.PreviousRevision != secondDelete.Terminal.Revision {
				t.Fatal("same-content native recovery was no-op or alias lost D")
			}
			history = graphMixedResult[historyResult](t, call("versions", "subject"))
			if history.CurrentState != nil || len(history.Terminal) != 2 || len(history.Versions) != 4 {
				t.Fatalf("repeated native recovery lost terminal or citable rows: %+v", history)
			}
			assertCanonicalHistory(history)
			refuse("gone", "show", source.ID, "--version", secondDelete.Terminal.Revision)
		})
	}
}
