//go:build cgo

package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// The nested commands must exercise the same installed CLI route as their
// compatibility spellings, including the checked Memory writer and raw body
// output. The server subtest uses the ordinary shared test server when present.
func TestGraphPreviewMemoryConvenienceWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			const scope = "https://example.invalid/memory-convenience/"
			initArgs := []string{"init", "--graph-mode", "link", "--prefix", "mem", "--scope-url", scope, "--skip-hooks", "--skip-agents", "--non-interactive", "--json"}
			if engine == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for the ordinary server workflow")
				}
				initArgs = append(initArgs, "--server", "--external", "--server-host", "127.0.0.1", "--server-port", port, "--server-user", "root")
			}
			graphPolicyCLI(t, bd, work, home, nil, "", initArgs...)
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			refuse := func(code string, args ...string) {
				t.Helper()
				graphPolicyCLI(t, bd, work, home, nil, code, append(args, "--json")...)
			}
			created := graphMixedResult[graphstore.Record](t, call("memory", "remember", "--id", "policy", "--type", "policy", "--create-only", "--body", "First body", "--title", "Policy title", "--status", "reviewed"))
			if created.ID != scope+"policy" || created.Type != scope+"policy" || created.Properties.Body != "First body" {
				t.Fatalf("nested typed creation lost identity, Type or body: %+v", created)
			}
			assertStatus := func(record graphstore.Record, want string) {
				t.Helper()
				raw, err := json.Marshal(record.Properties)
				if err != nil {
					t.Fatal(err)
				}
				var props map[string]any
				if err := json.Unmarshal(raw, &props); err != nil || props["status"] != want {
					t.Fatalf("status mismatch: want %q, properties=%s, error=%v", want, raw, err)
				}
			}
			assertStatus(created, "reviewed")
			plain := graphMixedResult[graphstore.Record](t, call("memory", "remember", "--id", "plain", "--create-only", "--body", "Plain body", "--status", "open"))
			if plain.Type != scope+"memory" {
				t.Fatalf("omitted Type did not create plain Memory: %+v", plain)
			}
			assertStatus(plain, "open")
			fromProperties := graphMixedResult[graphstore.Record](t, call("remember", "--id", "policy-props", "--type", "policy", "--create-only", "--properties", `{"title":"Properties title","body":"Property body","status":"draft"}`))
			if fromProperties.Type != created.Type || fromProperties.Properties.Title != "Properties title" || fromProperties.Properties.Body != "Property body" {
				t.Fatalf("top-level typed properties creation diverged: %+v", fromProperties)
			}
			assertStatus(fromProperties, "draft")
			for _, field := range []string{"title", "body"} {
				id := "null-" + field
				properties := `{"title":"Valid title","body":null}`
				if field == "title" {
					properties = `{"title":null,"body":"Valid body"}`
				}
				refuse("invalid_properties", "memory", "remember", "--id", id, "--create-only", "--properties", properties)
				createdAfterRefusal := graphMixedResult[graphstore.Record](t, call("memory", "remember", "--id", id, "--create-only", "--title", "Recovered", "--body", "Valid body"))
				if createdAfterRefusal.ID != scope+id || createdAfterRefusal.Properties.Body != "Valid body" {
					t.Fatalf("invalid %s consumed identity %q: %+v", field, id, createdAfterRefusal)
				}
			}
			emptyPropertyBody := graphMixedResult[graphstore.Record](t, call("memory", "remember", "--id", "empty-property-body", "--create-only", "--properties", `{"title":"Empty body","body":""}`))
			if emptyPropertyBody.Properties.Title != "Empty body" || emptyPropertyBody.Properties.Body != "" {
				t.Fatalf("valid empty body string was changed: %+v", emptyPropertyBody)
			}
			refuse("invalid_properties", "memory", "remember", "--id", "conflict", "--create-only", "--body", "One", "--properties", `{"body":"Two"}`)
			issue := graphMixedResult[graphstore.IssueRecord](t, call("create", "Issue target", "--id", "mem-work", "--type", "task"))
			refuse("capability_unavailable", "memory", "forget", issue.Properties.ID)
			call("link", "create", "--from", issue.ID, "--to", fromProperties.ID, "--id", "links/protect", "--type", "references")
			refuse("constraint_violation", "memory", "forget", "policy-props")
			graphPolicyCLI(t, bd, work, home, nil, "stderr:revision_unknown", "memory", "recall", "policy", "--revision", "=no-such-token")
			if got := graphPolicyCLI(t, bd, work, home, nil, "", "memory", "recall", "policy"); got != "First body\n" {
				t.Fatalf("nested recall changed body output: %q", got)
			}
			if top := graphPolicyCLI(t, bd, work, home, nil, "", "recall", "policy"); top != "First body\n" {
				t.Fatalf("top-level recall diverged: %q", top)
			}
			updated := graphMixedResult[graphstore.MemoryMutationResult](t, call("remember", "--update", "policy", "--body=", "--status", "approved", "--if-revision", "="+created.Revision))
			if !updated.Changed || updated.Memory.Properties.Title != "Policy title" || updated.Memory.Properties.Body != "" || updated.Memory.Type != created.Type {
				t.Fatalf("body/status shorthand did not preserve omitted fields and Type: %+v", updated)
			}
			assertStatus(updated.Memory, "approved")
			if got := graphPolicyCLI(t, bd, work, home, nil, "", "memory", "recall", "policy", "--quiet"); got != "\n" {
				t.Fatalf("empty quiet recall must emit one newline, got %q", got)
			}
			unchanged := graphMixedResult[graphstore.MemoryMutationResult](t, call("memory", "remember", "--update", "policy", "--status", "approved", "--if-revision", "="+updated.Memory.Revision))
			if unchanged.Changed || unchanged.Memory.Revision != updated.Memory.Revision {
				t.Fatalf("status no-op changed revision: %+v", unchanged)
			}
			refuse("revision_conflict", "memory", "remember", "--update", "policy", "--status", "stale", "--if-revision", "="+created.Revision)
			refuse("invalid_selector", "memory", "remember", "--id", "other", "--type", "policy", "--body", "Other")
			refuse("invalid_properties", "memory", "remember", "--id", "wrong-kind", "--type", "types/task", "--create-only", "--body", "Wrong")
			refuse("identity_reserved", "memory", "remember", "--id", "policy", "--create-only", "--body", "Duplicate")
			if nested, top := graphPolicyCLI(t, bd, work, home, nil, "", "memory", "memories", "Policy", "--format", "records-json"), graphPolicyCLI(t, bd, work, home, nil, "", "memories", "Policy", "--format", "records-json"); nested != top {
				t.Fatalf("nested memories changed summary output: nested=%s top=%s", nested, top)
			}
			if got := graphPolicyCLI(t, bd, work, home, nil, "", "memory", "forget", "policy", "--if-revision", "="+updated.Memory.Revision, "--json"); !strings.Contains(got, "policy") {
				t.Fatalf("nested forget lacked deletion receipt: %s", got)
			}
			graphPolicyCLI(t, bd, work, home, nil, "stderr:gone", "memory", "recall", "policy")
		})
	}
}

func TestLegacyMemoryConvenienceAliases(t *testing.T) {
	bd := buildBDUnderTest(t)
	work, home := t.TempDir(), t.TempDir()
	graphPolicyCLI(t, bd, work, home, nil, "", "init", "--prefix", "legacy", "--skip-hooks", "--skip-agents", "--non-interactive")
	graphPolicyCLI(t, bd, work, home, nil, "", "memory", "remember", "Legacy body", "--key", "alias-note")
	if nested, top := graphPolicyCLI(t, bd, work, home, nil, "", "memory", "recall", "alias-note"), graphPolicyCLI(t, bd, work, home, nil, "", "recall", "alias-note"); nested != top || !strings.Contains(nested, "Legacy body") {
		t.Fatalf("legacy recall aliases diverged: nested=%q top=%q", nested, top)
	}
	if nested, top := graphPolicyCLI(t, bd, work, home, nil, "", "memory", "memories", "alias-note"), graphPolicyCLI(t, bd, work, home, nil, "", "memories", "alias-note"); nested != top {
		t.Fatalf("legacy summary aliases diverged: nested=%q top=%q", nested, top)
	}
	graphPolicyCLI(t, bd, work, home, nil, "", "memory", "forget", "alias-note")
	if nested, top := graphPolicyCLI(t, bd, work, home, nil, "", "memory", "memories", "alias-note"), graphPolicyCLI(t, bd, work, home, nil, "", "memories", "alias-note"); nested != top {
		t.Fatalf("legacy post-forget summaries diverged: nested=%q top=%q", nested, top)
	}
}
