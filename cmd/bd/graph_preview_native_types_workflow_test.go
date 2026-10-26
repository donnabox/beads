//go:build cgo

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/httpapi/bdpwire"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func TestGraphPreviewNativeConcreteTypeWorkflow(t *testing.T) {
	isolateBeadsDirForTest(t)
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			const scope = "https://example.invalid/native-cli/"
			work, home := graphListWorkspace(t, bd, engine, scope)
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			first := graphMixedResult[graphstore.IssueRecord](t, call("create", "Default task", "--graph-path", "beads/first"))
			bug := graphMixedResult[graphstore.IssueRecord](t, call("create", "Bug", "--graph-path", "beads/bug", "--type", "bug"))
			taskType, _ := graphstore.NativeIssueTypeURL(scope, "task")
			bugType, _ := graphstore.NativeIssueTypeURL(scope, "bug")
			if first.Type != taskType || bug.Type != bugType {
				t.Fatalf("native declarations: %+v %+v", first, bug)
			}
			graphListConfig(t, work, "types.custom: \"review-class\"\n")
			custom := graphMixedResult[graphstore.IssueRecord](t, call("create", "Custom", "--graph-path", "beads/custom", "--type", "review-class"))
			customType, _ := graphstore.NativeIssueTypeURL(scope, "review-class")
			if custom.Type != customType {
				t.Fatalf("custom declaration lost: %+v", custom)
			}
			decision := graphMixedResult[graphstore.IssueRecord](t, call("create", "Native decision", "--graph-path", "beads/native-decision", "--type", "decision"))
			memory := graphMixedResult[graphstore.Record](t, call("create", "Memory scenario", "--id", "beads/memory-scenario", "--type", "scenario"))
			if decision.Type == memory.Type || memory.Type != scope+"scenario" {
				t.Fatal("native/Memory names collided")
			}
			for _, tc := range []struct {
				flags []string
				want  int
			}{{nil, 4}, {[]string{"--type", "issue"}, 4}, {[]string{"--type", "any"}, 5}, {[]string{"--exact-type", "task"}, 1}, {[]string{"--exact-type", taskType}, 1}, {[]string{"--type", "issue", "--exact-type", "task"}, 1}, {[]string{"--type", "bug", "--exact-type", "task"}, 0}, {[]string{"--exact-type", "issue"}, 0}, {[]string{"--type", "memory"}, 1}, {[]string{"--type", "decision"}, 1}, {[]string{"--type", "scenario"}, 1}} {
				paths, more := graphListBeads(t, bd, work, home, scope, tc.flags...)
				if more || len(paths) != tc.want {
					t.Fatalf("list %v: count=%d more=%t want=%d", tc.flags, len(paths), more, tc.want)
				}
			}
			canonicalBeadCount := func(flags ...string) int {
				t.Helper()
				out := graphPolicyCLI(t, bd, work, home, nil, "", append([]string{"list", "--output-format", "json", "--unpaged"}, flags...)...)
				var frame struct {
					Data json.RawMessage `json:"data"`
				}
				if err := json.Unmarshal([]byte(out), &frame); err != nil {
					t.Fatal(err)
				}
				var page bdpwire.BeadCollection
				if err := bdpwire.Unmarshal(frame.Data, &page); err != nil {
					t.Fatal(err)
				}
				return len(page.Items)
			}
			if got := canonicalBeadCount("--type", "issue", "--exact-type", "task"); got != 1 {
				t.Fatalf("canonical matching Type intersection returned %d Beads, want 1", got)
			}
			if got := canonicalBeadCount("--conforms-to", graphstore.IssueTypeURL(scope), "--exact-type", "task"); got != 1 {
				t.Fatalf("canonical family alias intersection returned %d Beads, want 1", got)
			}
			if got := canonicalBeadCount("--type", "bug", "--exact-type", "task"); got != 0 {
				t.Fatalf("canonical disjoint Type intersection returned %d Beads, want 0", got)
			}
			graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "list", "--type", "issue", "--exact-type", "task", "--status", "open", "--format", "records-json")
			for _, tc := range []struct {
				flags []string
				want  int
			}{{nil, 4}, {[]string{"--type", "any"}, 5}, {[]string{"--type", "memory"}, 1}, {[]string{"--exact-type", "task"}, 1}, {[]string{"--type", "issue", "--exact-type", "task"}, 1}, {[]string{"--type", "bug", "--exact-type", "task"}, 0}, {[]string{"--exact-type", "issue"}, 0}} {
				var result struct {
					Count int `json:"count"`
				}
				out := call(append([]string{"count"}, tc.flags...)...)
				if err := json.Unmarshal([]byte(out), &result); err != nil || result.Count != tc.want {
					t.Fatalf("count %v: %s %v", tc.flags, out, err)
				}
			}
			linked := graphMixedResult[graphstore.DependencyResult](t, call("link", "create", first.ID, bug.ID, "--type", "parent-child", "--id", "links/child"))
			parentType, _ := graphstore.NativeDependencyTypeURL(scope, "parent-child")
			if linked.Link.Type != parentType {
				t.Fatalf("native Link declaration: %+v", linked)
			}
			for _, tc := range []struct {
				flags []string
				want  int
			}{{[]string{"--type", "dependency"}, 1}, {[]string{"--exact-type", "parent-child"}, 1}, {[]string{"--type", "dependency", "--exact-type", "parent-child"}, 1}, {[]string{"--type", "blocks", "--exact-type", "parent-child"}, 0}, {[]string{"--exact-type", "dependency"}, 0}} {
				listed := graphMixedResult[[]graphstore.LinkRecord](t, call(append([]string{"link", "list"}, tc.flags...)...))
				counted := graphMixedResult[map[string]int](t, call(append([]string{"link", "count"}, tc.flags...)...))
				if len(listed) != tc.want || counted["count"] != tc.want {
					t.Fatalf("Link list/count %v disagree: %+v %+v", tc.flags, listed, counted)
				}
			}
			canonicalLinkCount := func(flags ...string) int {
				t.Helper()
				out := graphPolicyCLI(t, bd, work, home, nil, "", append([]string{"link", "list", "--output-format", "json", "--unpaged"}, flags...)...)
				var frame struct {
					Data json.RawMessage `json:"data"`
				}
				if err := json.Unmarshal([]byte(out), &frame); err != nil {
					t.Fatal(err)
				}
				var page bdpwire.LinkCollection
				if err := bdpwire.Unmarshal(frame.Data, &page); err != nil {
					t.Fatal(err)
				}
				return len(page.Items)
			}
			if got := canonicalLinkCount("--type", "dependency", "--exact-type", "parent-child"); got != 1 {
				t.Fatalf("canonical matching Link Type intersection returned %d, want 1", got)
			}
			if got := canonicalLinkCount("--type", "blocks", "--exact-type", "parent-child"); got != 0 {
				t.Fatalf("canonical disjoint Link Type intersection returned %d, want 0", got)
			}
			unknown, _ := graphstore.NativeDependencyTypeURL(scope, "unknown-kind")
			for _, kind := range []string{"unknown-kind", unknown} {
				graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "link", "create", first.ID, custom.ID, "--type", kind, "--json")
			}
			changed := graphMixedResult[graphstore.IssueMutationResult](t, call("update", first.ID, "--type", "feature", "--title", "Renamed"))
			featureType, _ := graphstore.NativeIssueTypeURL(scope, "feature")
			if changed.Issue.Type != featureType || changed.Issue.Properties.Title != "Renamed" {
				t.Fatalf("native setter ignored: %+v", changed)
			}
			retained := graphMixedResult[graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "show", first.ID, "--revision", "="+first.Revision, "--format", "graph-json"))
			if retained.Type != taskType {
				t.Fatalf("retained affiliation changed: %+v", retained)
			}
			before := graphMixedResult[graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "show", first.ID, "--format", "graph-json"))
			graphPolicyCLI(t, bd, work, home, nil, "invalid_selector", "update", first.ID, "--type", "issue", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "invalid_properties", "update", first.ID, "--type", bugType, "--properties", `{"issue_type":"feature"}`, "--json")
			after := graphMixedResult[graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "show", first.ID, "--format", "graph-json"))
			if !reflect.DeepEqual(before, after) {
				t.Fatal("refused native setter changed state")
			}
			patched := graphMixedResult[graphstore.IssueMutationResult](t, call("update", first.ID, "--patch", `[{"op":"replace","path":"/issue_type","value":"bug"}]`))
			if patched.Issue.Type != bugType {
				t.Fatalf("native property setter diverged: %+v", patched)
			}
			// Native notes-clearing intent remains guarded. The affiliation test uses
			// a nonempty notes value rather than relaxing that independent safeguard.
			beforeClear := graphMixedResult[graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "show", first.ID, "--format", "graph-json"))
			graphPolicyCLI(t, bd, work, home, nil, "invalid_properties", "update", first.ID, "--replace-properties", `{"title":"Replacement","priority":1}`, "--json")
			afterClear := graphMixedResult[graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "show", first.ID, "--format", "graph-json"))
			if !reflect.DeepEqual(beforeClear, afterClear) {
				t.Fatal("refused native notes clear changed state")
			}
			replaced := graphMixedResult[graphstore.IssueMutationResult](t, call("update", first.ID, "--replace-properties", `{"title":"Replacement","priority":1,"notes":"Replacement notes"}`))
			if replaced.Issue.Type != bugType || replaced.Issue.Properties.Title != "Replacement" {
				t.Fatalf("replacement lost affiliation: %+v", replaced)
			}
			restored := graphMixedResult[graphstore.IssueMutationResult](t, call("update", first.ID, "--from-revision", "="+first.Revision))
			if restored.Issue.Type != taskType {
				t.Fatalf("historical restore lost affiliation: %+v", restored)
			}
			raw := call("types")
			inventory := graphMixedResult[graphTypeInventory](t, raw)
			found := map[string]bool{}
			for _, entry := range append(inventory.BeadTypes, inventory.LinkTypes...) {
				var descriptor bdpwire.TypeDescriptor
				if err := bdpwire.Unmarshal(entry, &descriptor); err != nil {
					t.Fatal(err)
				}
				found[descriptor.ID] = true
				if descriptor.ID == graphstore.IssueTypeURL(scope) || descriptor.ID == graphstore.DependencyTypeURL(scope) {
					if !strings.Contains(descriptor.Description, "Abstract native") || !strings.Contains(descriptor.Description, "cannot be declared directly") {
						t.Fatalf("JSON inventory hides abstract admission contract: %s", entry)
					}
				}
			}
			for _, id := range []string{taskType, bugType, customType, parentType, graphstore.IssueTypeURL(scope), graphstore.DependencyTypeURL(scope)} {
				if !found[id] {
					t.Fatalf("inventory omitted %s: %s", id, raw)
				}
			}
			human := graphPolicyCLI(t, bd, work, home, nil, "", "types")
			if !strings.Contains(human, "Issue (abstract)") || !strings.Contains(human, "Dependency (abstract)") {
				t.Fatalf("abstract status hidden: %s", human)
			}
			if err := os.WriteFile(filepath.Join(work, ".beads", "config.yaml"), []byte("types.custom: \"\"\n"), 0600); err != nil {
				t.Fatal(err)
			}
			// Used declarations remain discoverable after authorization is removed.
			raw = call("types")
			if !strings.Contains(raw, customType) {
				t.Fatal("retained custom Type disappeared")
			}
			graphPolicyCLI(t, bd, work, home, nil, "invalid_properties", "create", "Unauthorized", "--type", "review-class", "--json")
		})
	}
}
