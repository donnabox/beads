//go:build cgo

package main

import (
	"reflect"
	"testing"
	"time"

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
				graphPatchProcess(t, bd, work, home, nil, tc.code, 90*time.Second, append(tc.args, "--json")...)
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
