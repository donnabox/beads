//go:build cgo

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func TestGraphPreviewCreateGraphAtomicWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			args := []string{"init", "--graph-mode", "link", "--prefix", "batch", "--scope-url", "https://example.invalid/batch/", "--skip-hooks", "--skip-agents", "--non-interactive"}
			if engine == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for shared-server qualification")
				}
				args = append(args, "--server", "--external", "--server-host", "127.0.0.1", "--server-port", port, "--server-user", "root")
			}
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			refuse := func(code string, args ...string) {
				t.Helper()
				graphPolicyCLI(t, bd, work, home, nil, code, append(args, "--json")...)
			}
			call(args...)
			path := filepath.Join(work, "graph.json")
			writePlan := func(raw string) {
				t.Helper()
				if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			writePlan(`{"nodes":[{"key":"child","title":"Child"},{"key":"parent","title":"Parent"}],"edges":[{"from_key":"child","to_key":"parent","type":"blocks"}]}`)
			var result GraphApplyResult
			if err := json.Unmarshal([]byte(call("create", "--graph", path)), &result); err != nil {
				t.Fatal(err)
			}
			if len(result.IDs) != 2 || result.IDs["child"] == "" || result.IDs["parent"] == "" {
				t.Fatalf("symbolic keys did not bind to two Issues: %+v", result)
			}
			child := graphMixedResult[graphstore.IssueRecord](t, call("show", result.IDs["child"], "--format", "graph-json"))
			parent := graphMixedResult[graphstore.IssueRecord](t, call("show", result.IDs["parent"], "--format", "graph-json"))
			if child.Properties.ID != result.IDs["child"] || parent.Properties.ID != result.IDs["parent"] {
				t.Fatalf("created IDs did not resolve: child=%+v parent=%+v", child, parent)
			}
			links := graphMixedResult[[]graphstore.LinkRecord](t, call("link", "list", "--from", result.IDs["child"]))
			if len(links) != 1 {
				t.Fatalf("symbolic edge did not publish one canonical Link: %+v", links)
			}
			writePlan(`{"nodes":[{"key":"later","title":"Later","id":"batch-later"}],"edges":[{"from_key":"later","to_id":"` + result.IDs["parent"] + `"}]}`)
			var linked GraphApplyResult
			if err := json.Unmarshal([]byte(call("create", "--graph", path)), &linked); err != nil {
				t.Fatal(err)
			}
			if linked.IDs["later"] != "batch-later" {
				t.Fatalf("existing-target batch lost explicit source ID: %+v", linked)
			}
			if got := graphMixedResult[[]graphstore.LinkRecord](t, call("link", "list", "--from", "batch-later")); len(got) != 1 || got[0].Target != parent.ID {
				t.Fatalf("existing-target batch did not publish its Link: %+v", got)
			}
			writePlan(`{"nodes":[{"key":"fields","id":"batch-fields","title":"Fields","description":"first line","design":"  design — 雪  ","acceptance_criteria":"accept café","notes":"initial note","spec_id":"spec-7","external_ref":"external-7","assignee":"agent","owner":"team","estimated_minutes":45,"due_at":"2030-01-02T03:04:05Z","labels":["urgent","review"]}],"edges":[{"from_key":"fields","to_id":"` + parent.Properties.ID + `"}]}`)
			var fields GraphApplyResult
			if err := json.Unmarshal([]byte(call("create", "--graph", path)), &fields); err != nil {
				t.Fatal(err)
			}
			created := graphMixedResult[graphstore.IssueRecord](t, call("show", fields.IDs["fields"], "--format", "graph-json"))
			wantDue, err := time.Parse(time.RFC3339, "2030-01-02T03:04:05Z")
			if err != nil {
				t.Fatal(err)
			}
			p := created.Properties
			if p == nil || p.ID != "batch-fields" || p.Description != "first line" || p.Design != "  design — 雪  " ||
				p.AcceptanceCriteria != "accept café" || p.Notes != "initial note" || p.SpecID != "spec-7" ||
				p.ExternalRef == nil || *p.ExternalRef != "external-7" || p.Assignee != "agent" || p.Owner != "team" ||
				p.EstimatedMinutes == nil || *p.EstimatedMinutes != 45 || p.DueAt == nil || !p.DueAt.Equal(wantDue) ||
				!reflect.DeepEqual(p.Labels, []string{"review", "urgent"}) {
				t.Fatalf("atomic graph create lost initial Issue fields: %+v", p)
			}
			if got := graphMixedResult[[]graphstore.LinkRecord](t, call("link", "list", "--from", "batch-fields")); len(got) != 1 || got[0].Target != parent.ID {
				t.Fatalf("initial fields broke canonical Link publication: %+v", got)
			}
			writePlan(`{"nodes":[{"key":"dupe","title":"Must not exist","id":"batch-dupe"}],"edges":[{"from_key":"dupe","to_id":"` + result.IDs["parent"] + `"},{"from_key":"dupe","to_id":"https://example.invalid/batch/` + result.IDs["parent"] + `"}]}`)
			refuse("invalid_properties", "create", "--graph", path)
			refuse("not_found", "show", "batch-dupe")
			writePlan(`{"nodes":[{"key":"orphan","title":"Must not exist","id":"batch-orphan"}],"edges":[{"from_key":"orphan","to_id":"batch-missing"}]}`)
			refuse("not_found", "create", "--graph", path)
			refuse("not_found", "show", "batch-orphan")
			writePlan(`{"nodes":[{"key":"closed","title":"Must not exist","id":"batch-closed","status":"closed"}]}`)
			refuse("capability_unavailable", "create", "--graph", path)
			refuse("not_found", "show", "batch-closed")
		})
	}
}
