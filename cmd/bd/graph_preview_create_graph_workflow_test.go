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
			links := graphMixedResult[[]graphstore.LinkRecord](t, call("link", "list", "--source", result.IDs["child"]))
			if len(links) != 1 {
				t.Fatalf("symbolic edge did not publish one canonical Link: %+v", links)
			}
			writePlan(`{"nodes":[{"key":"root","title":"Root","type":"epic"},{"key":"child","title":"Child","parent_key":"root"},{"key":"alias","title":"Alias","parent":"root","parent_id":"missing-parent"}]}`)
			var hierarchy GraphApplyResult
			if err := json.Unmarshal([]byte(call("create", "--graph", path)), &hierarchy); err != nil {
				t.Fatal(err)
			}
			root := graphMixedResult[graphstore.IssueRecord](t, call("show", hierarchy.IDs["root"], "--format", "graph-json"))
			for _, key := range []string{"child", "alias"} {
				owned := graphMixedResult[[]graphstore.LinkRecord](t, call("link", "list", "--source", hierarchy.IDs[key]))
				if len(owned) != 1 || owned[0].Target != root.ID || owned[0].Properties["dep_type"] != "parent-child" {
					t.Fatalf("graph parent key %s did not publish parent-child Link: %+v", key, owned)
				}
			}
			writePlan(`{"nodes":[{"key":"external-child","title":"External child","parent_id":"` + hierarchy.IDs["root"] + `"}]}`)
			var externalChild GraphApplyResult
			if err := json.Unmarshal([]byte(call("create", "--graph", path)), &externalChild); err != nil {
				t.Fatal(err)
			}
			owned := graphMixedResult[[]graphstore.LinkRecord](t, call("link", "list", "--source", externalChild.IDs["external-child"]))
			if len(owned) != 1 || owned[0].Target != root.ID || owned[0].Properties["dep_type"] != "parent-child" {
				t.Fatalf("graph parent_id did not publish parent-child Link: %+v", owned)
			}
			writePlan(`{"nodes":[{"key":"inline-source","id":"batch-inline-source","title":"Inline source","deps":[{"target":"inline-target"},{"target":"` + root.Properties.ID + `","type":"related"}]},{"key":"inline-target","id":"batch-inline-target","title":"Inline target"}]}`)
			var inline GraphApplyResult
			if err := json.Unmarshal([]byte(call("create", "--graph", path)), &inline); err != nil {
				t.Fatal(err)
			}
			inlineTarget := graphMixedResult[graphstore.IssueRecord](t, call("show", inline.IDs["inline-target"], "--format", "graph-json"))
			inlineLinks := graphMixedResult[[]graphstore.LinkRecord](t, call("link", "list", "--source", inline.IDs["inline-source"]))
			if len(inlineLinks) != 2 || inlineLinks[0].Target == inlineLinks[1].Target {
				t.Fatalf("inline dependencies did not publish two distinct Links: %+v", inlineLinks)
			}
			foundForward, foundExisting := false, false
			for _, link := range inlineLinks {
				if link.Target == inlineTarget.ID && link.Properties["dep_type"] == "blocks" {
					foundForward = true
				}
				if link.Target == root.ID && link.Properties["dep_type"] == "related" {
					foundExisting = true
				}
			}
			if !foundForward || !foundExisting {
				t.Fatalf("inline dependencies lost their target or type: %+v", inlineLinks)
			}
			writePlan(`{"nodes":[{"key":"deferred-by-date","id":"batch-deferred-by-date","title":"Deferred by date","defer_until":"2099-01-02T03:04:05Z","deps":[{"target":"` + inline.IDs["inline-target"] + `"}]}]}`)
			var deferredByDate GraphApplyResult
			if err := json.Unmarshal([]byte(call("create", "--graph", path)), &deferredByDate); err != nil {
				t.Fatal(err)
			}
			deferredRecord := graphMixedResult[graphstore.IssueRecord](t, call("show", deferredByDate.IDs["deferred-by-date"], "--format", "graph-json"))
			wantDefer, err := time.Parse(time.RFC3339, "2099-01-02T03:04:05Z")
			if err != nil {
				t.Fatal(err)
			}
			if deferredRecord.Properties.Status != "deferred" || deferredRecord.Properties.DeferUntil == nil || !deferredRecord.Properties.DeferUntil.Equal(wantDefer) || len(deferredRecord.Owned) != 1 {
				t.Fatalf("graph plan lost initial deferral or local Link: %+v", deferredRecord)
			}
			_, kind, inlineVersions, _ := graphVersionsListed(t, call("versions", inline.IDs["inline-source"]))
			if kind != "issue" || len(inlineVersions) != 1 {
				t.Fatalf("inline dependencies must be part of one canonical inception version: %+v", inlineVersions)
			}
			writePlan(`{"nodes":[{"key":"external","id":"batch-external-must-not-exist","title":"External","deps":[{"target":"external:other#1"}]}]}`)
			refuse("capability_unavailable", "create", "--graph", path)
			refuse("not_found", "show", "batch-external-must-not-exist")
			writePlan(`{"nodes":[{"key":"missing","id":"batch-inline-missing","title":"Missing","deps":[{"target":"batch-missing"}]}]}`)
			refuse("not_found", "create", "--graph", path)
			refuse("not_found", "show", "batch-inline-missing")
			writePlan(`{"nodes":[{"key":"later","title":"Later","id":"batch-later"}],"edges":[{"from_key":"later","to_id":"` + result.IDs["parent"] + `"}]}`)
			var linked GraphApplyResult
			if err := json.Unmarshal([]byte(call("create", "--graph", path)), &linked); err != nil {
				t.Fatal(err)
			}
			if linked.IDs["later"] != "batch-later" {
				t.Fatalf("existing-target batch lost explicit source ID: %+v", linked)
			}
			if got := graphMixedResult[[]graphstore.LinkRecord](t, call("link", "list", "--source", "batch-later")); len(got) != 1 || got[0].Target != parent.ID {
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
			if got := graphMixedResult[[]graphstore.LinkRecord](t, call("link", "list", "--source", "batch-fields")); len(got) != 1 || got[0].Target != parent.ID {
				t.Fatalf("initial fields broke canonical Link publication: %+v", got)
			}
			writePlan(`{"nodes":[{"key":"deferred","id":"batch-deferred","title":"Deferred","assignee":"agent","assign_after_create":true}],"edges":[{"from_key":"deferred","to_id":"` + parent.Properties.ID + `"}]}`)
			var deferred GraphApplyResult
			if err := json.Unmarshal([]byte(call("create", "--graph", path)), &deferred); err != nil {
				t.Fatal(err)
			}
			assigned := graphMixedResult[graphstore.IssueRecord](t, call("show", deferred.IDs["deferred"], "--format", "graph-json"))
			if assigned.Properties == nil || assigned.Properties.Assignee != "agent" || len(assigned.Owned) != 1 {
				t.Fatalf("deferred assignment lost final Issue or Link: %+v", assigned)
			}
			_, kind, deferredVersions, _ := graphVersionsListed(t, call("versions", deferred.IDs["deferred"]))
			if kind != "issue" || len(deferredVersions) != 1 {
				t.Fatalf("deferred assignment must have one canonical inception version: %+v", deferredVersions)
			}
			writePlan(`{"nodes":[{"key":"referencing","id":"batch-referencing","title":"Referencing","metadata":{"note":"kept"},"metadata_refs":{"peer":"referenced","self":"referencing"},"assignee":"agent","assign_after_create":true},{"key":"referenced","id":"batch-referenced","title":"Referenced"}],"edges":[{"from_key":"referencing","to_key":"referenced"}]}`)
			var referenced GraphApplyResult
			if err := json.Unmarshal([]byte(call("create", "--graph", path)), &referenced); err != nil {
				t.Fatal(err)
			}
			withRefs := graphMixedResult[graphstore.IssueRecord](t, call("show", referenced.IDs["referencing"], "--format", "graph-json"))
			var metadata map[string]any
			if err := json.Unmarshal(withRefs.Metadata, &metadata); err != nil {
				t.Fatal(err)
			}
			if metadata["note"] != "kept" || metadata["peer"] != referenced.IDs["referenced"] || metadata["self"] != referenced.IDs["referencing"] || withRefs.Properties == nil || withRefs.Properties.Assignee != "agent" || len(withRefs.Owned) != 1 {
				t.Fatalf("graph plan lost metadata refs or Link: %+v", withRefs)
			}
			call("create", "Existing source", "--id", "batch-existing-source")
			original := graphMixedResult[graphstore.IssueRecord](t, call("show", "batch-existing-source", "--format", "graph-json"))
			writePlan(`{"nodes":[{"key":"new-target","id":"batch-new-target","title":"New target"}],"edges":[{"from_id":"batch-existing-source","to_key":"new-target"}]}`)
			var existingSource GraphApplyResult
			if err := json.Unmarshal([]byte(call("create", "--graph", path)), &existingSource); err != nil {
				t.Fatal(err)
			}
			if existingSource.IDs["new-target"] != "batch-new-target" {
				t.Fatalf("existing-source plan did not create target: %+v", existingSource)
			}
			updated := graphMixedResult[graphstore.IssueRecord](t, call("show", "batch-existing-source", "--format", "graph-json"))
			if updated.Revision == original.Revision || len(updated.Owned) != 1 {
				t.Fatalf("existing source did not retain a new Link/version: %+v", updated)
			}
			newTarget := graphMixedResult[graphstore.IssueRecord](t, call("show", "batch-new-target", "--format", "graph-json"))
			if got := graphMixedResult[[]graphstore.LinkRecord](t, call("link", "list", "--source", "batch-existing-source")); len(got) != 1 || got[0].Target != newTarget.ID {
				t.Fatalf("existing source did not publish a Link to the new target: %+v", got)
			}
			writePlan(`{"nodes":[{"key":"multi-a","id":"batch-multi-a","title":"A"},{"key":"multi-b","id":"batch-multi-b","title":"B"}],"edges":[{"from_id":"batch-existing-source","to_key":"multi-a"},{"from_id":"beads/batch-existing-source","to_key":"multi-b"}]}`)
			var multi GraphApplyResult
			if err := json.Unmarshal([]byte(call("create", "--graph", path)), &multi); err != nil {
				t.Fatal(err)
			}
			if multi.IDs["multi-a"] != "batch-multi-a" || multi.IDs["multi-b"] != "batch-multi-b" {
				t.Fatalf("multi-Link plan did not bind both targets: %+v", multi)
			}
			updated = graphMixedResult[graphstore.IssueRecord](t, call("show", "batch-existing-source", "--format", "graph-json"))
			if len(updated.Owned) != 3 {
				t.Fatalf("existing source lost a Link after multi-Link plan: %+v", updated)
			}
			if got := graphMixedResult[[]graphstore.LinkRecord](t, call("link", "list", "--source", "batch-existing-source")); len(got) != 3 {
				t.Fatalf("multi-Link plan did not publish all Links: %+v", got)
			}
			_, kind, versions, _ := graphVersionsListed(t, call("versions", "batch-existing-source"))
			if kind != "issue" || len(versions) != 4 || versions[0].Version != updated.Revision || versions[3].Version != original.Revision {
				t.Fatalf("multi-Link plan did not retain every source version: %+v", versions)
			}
			intermediate := graphMixedResult[graphstore.IssueRecord](t, call("show", "batch-existing-source", "--version", versions[1].Version))
			if len(intermediate.Owned) != 2 {
				t.Fatalf("multi-Link plan's intermediate History has wrong owned set: %+v", intermediate)
			}
			writePlan(`{"nodes":[{"key":"existing-pair-node","id":"batch-existing-pair-node","title":"New node"}],"edges":[{"from_id":"batch-existing-source","to_id":"` + result.IDs["parent"] + `"}]}`)
			var existingPair GraphApplyResult
			if err := json.Unmarshal([]byte(call("create", "--graph", path)), &existingPair); err != nil {
				t.Fatal(err)
			}
			if existingPair.IDs["existing-pair-node"] != "batch-existing-pair-node" {
				t.Fatalf("existing-pair plan did not create its node: %+v", existingPair)
			}
			got := graphMixedResult[[]graphstore.LinkRecord](t, call("link", "list", "--source", "batch-existing-source"))
			parentLinks := 0
			for _, link := range got {
				if link.Target == parent.ID {
					parentLinks++
				}
			}
			if len(got) != 4 || parentLinks != 1 {
				t.Fatalf("existing-pair plan did not publish its Link: %+v", got)
			}
			_, kind, versions, _ = graphVersionsListed(t, call("versions", "batch-existing-source"))
			if kind != "issue" || len(versions) != 5 {
				t.Fatalf("existing-pair plan did not retain source History: %+v", versions)
			}
			writePlan(`{"nodes":[{"key":"repeat","id":"batch-repeat","title":"Repeat"}],"edges":[{"from_id":"batch-existing-source","to_id":"` + result.IDs["parent"] + `"}]}`)
			var repeated GraphApplyResult
			if err := json.Unmarshal([]byte(call("create", "--graph", path)), &repeated); err != nil || repeated.IDs["repeat"] != "batch-repeat" {
				t.Fatalf("change-free pair re-add did not create its new node: result=%+v err=%v", repeated, err)
			}
			unchanged := graphMixedResult[graphstore.IssueRecord](t, call("show", "batch-existing-source", "--format", "graph-json"))
			if unchanged.Revision != versions[0].Version || len(unchanged.Owned) != 4 {
				t.Fatalf("change-free pair re-add changed the source: %+v", unchanged)
			}
			_, kind, versions, _ = graphVersionsListed(t, call("versions", "batch-existing-source"))
			if kind != "issue" || len(versions) != 5 {
				t.Fatalf("change-free pair re-add changed History: %+v", versions)
			}
			writePlan(`{"nodes":[{"key":"conflict","id":"batch-conflict-must-not-exist","title":"Must not exist"}],"edges":[{"from_id":"batch-existing-source","to_id":"` + result.IDs["parent"] + `","type":"related"}]}`)
			refuse("capability_unavailable", "create", "--graph", path)
			refuse("not_found", "show", "batch-conflict-must-not-exist")
			writePlan(`{"nodes":[{"key":"missing-source-target","id":"batch-missing-source-target","title":"Must not exist"}],"edges":[{"from_id":"batch-missing-source","to_key":"missing-source-target"}]}`)
			refuse("not_found", "create", "--graph", path)
			refuse("not_found", "show", "batch-missing-source-target")
			writePlan(`{"nodes":[{"key":"dupe","title":"Must not exist","id":"batch-dupe"}],"edges":[{"from_key":"dupe","to_id":"` + result.IDs["parent"] + `"},{"from_key":"dupe","to_id":"beads/` + result.IDs["parent"] + `"}]}`)
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
