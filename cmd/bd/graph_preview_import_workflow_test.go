//go:build cgo

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
)

// Installed process admission, input selection and persisted graph projections
// are distinct from the lower transaction fault-injection tests.
func TestGraphPreviewLegacyImportInstalledWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	fixture, err := os.ReadFile(filepath.Join("testdata", "legacy-import", "ordinary.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			isolateBeadsDirForTest(t)
			work, home := t.TempDir(), t.TempDir()
			const scope = "https://example.invalid/legacy-import/"
			call := func(input string, code string, args ...string) string {
				t.Helper()
				if args[0] != "list" {
					args = append(args, "--json")
				}
				return graphPatchProcess(t, bd, work, home, strings.NewReader(input), code, 90*time.Second, args...)
			}
			init := []string{"init", "--graph-mode", "link", "--scope-url", scope, "--skip-hooks", "--skip-agents", "--non-interactive"}
			if engine == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for required shared-server import proof")
				}
				init = append(init, "--server", "--external", "--server-host", "127.0.0.1", "--server-port", port, "--server-user", "root")
			}
			call("", "", init...)
			before := call("", "", "list", "--all", "--format", "records-json")
			for _, input := range []string{
				string(fixture) + "\n{broken}",
				"",
				`{"_schema":"beads-jsonl/1"}`,
				`{"id":"one-a","title":"a"}` + "\n" + `{"id":"two-b","title":"b"}`,
				strings.Replace(string(fixture), `"depends_on_id":"legacy-a"`, `"depends_on_id":"absent"`, 1),
				strings.Replace(string(fixture), `"type":"blocks"`, `"type":"`+strings.Repeat("x", types.MaxDependencyTypeLen+1)+`"`, 1),
				string(fixture) + `{"_type":"issue","id":"bad","title":"bad","history":[]}`,
				`{"id":"old-a","title":"a","dependencies":[{"depends_on_id":"old-b","type":"blocks"}]}
{"id":"old-b","title":"b","dependencies":[{"depends_on_id":"old-a","type":"blocks"}]}`,
				`{"id":"old-a","title":"Ancestor"}
{"id":"old-b","title":"Child","dependencies":[{"depends_on_id":"old-a","type":"parent-child"}]}
{"id":"old-c","title":"Grandchild","dependencies":[{"depends_on_id":"old-b","type":"parent-child"},{"depends_on_id":"old-a","type":"blocks"}]}`,
			} {
				call(input, "invalid_properties", "import", "-")
				if got := call("", "", "list", "--all", "--format", "records-json"); got != before {
					t.Fatal("refused import changed inventory")
				}
			}
			call(string(fixture), "capability_unavailable", "import", "-", "--allow-stale")
			call(string(fixture), "capability_unavailable", "import", "-", "--dedup")
			call(string(fixture), "capability_unavailable", "import", "-", "--global")
			call(string(fixture), "invalid_properties", "import", "-", "--source-prefix", "")
			call(string(fixture), "invalid_properties", "import", "-", "--source-prefix", "wrong", "--dry-run")
			call("", "invalid_selector", "import", "one", "two")
			call("", "invalid_properties", "import", "one", "--input", "two")
			call(strings.Repeat(" ", 16*1024*1024+1), "capability_unavailable", "import", "-")
			call(string(fixture), "permission_denied", "import", "-", "--readonly")
			call(string(fixture), "invalid_properties", "import")
			preview := graphMixedResult[graphstore.LegacyImportResult](t, call(string(fixture), "", "import", "-", "--source-prefix", "legacy", "--dry-run"))
			if !preview.DryRun || preview.IssuePrefix != "legacy" || preview.Issues != 2 || preview.Dependencies != 1 || preview.Comments != 1 || preview.Memories != 1 {
				t.Fatalf("preview=%+v", preview)
			}
			if call("", "", "list", "--all", "--format", "records-json") != before {
				t.Fatal("dry run reserved data")
			}
			file := filepath.Join(work, "ordinary.jsonl")
			if err := os.WriteFile(file, fixture, 0600); err != nil {
				t.Fatal(err)
			}
			result := graphMixedResult[graphstore.LegacyImportResult](t, call("", "", "import", "--input", file))
			if result.DryRun || result.IDs["legacy-b"] != scope+"beads/legacy-b" || result.MemoryIDs["fixture-key"] != scope+"beads/fixture-key" {
				t.Fatalf("result=%+v", result)
			}
			issue := graphMixedResult[graphstore.IssueRecord](t, call("", "", "show", "legacy-b", "--format", "graph-json"))
			if len(issue.Owned) != 1 || issue.Properties.Notes != "A note" || issue.Properties.ID != "legacy-b" {
				t.Fatalf("issue=%+v", issue)
			}
			memory := graphMixedResult[graphstore.Record](t, call("", "", "show", "fixture-key"))
			if memory.Properties.Title != "fixture-key" || memory.Properties.Body != "Remembered context" {
				t.Fatalf("memory=%+v", memory)
			}
			comments := call("", "", "comments", "legacy-b")
			if !strings.Contains(comments, "fixture-commenter") || !strings.Contains(comments, "A preserved comment") {
				t.Fatalf("comments=%s", comments)
			}
			retained := call("", "", "show", issue.ID, "--version", issue.Revision)
			if !strings.Contains(retained, "A note") {
				t.Fatal("missing retained initial Issue")
			}
			inventory := call("", "", "list", "--all", "--format", "records-json")
			versions := call("", "", "versions", issue.ID)
			for _, args := range [][]string{{"import", file}, {"import", file, "--dry-run"}} {
				call("", "invalid_properties", args...)
				if call("", "", "list", "--all", "--format", "records-json") != inventory || call("", "", "versions", issue.ID) != versions {
					t.Fatal("repeat import changed data or History")
				}
			}
			fresh := graphMixedResult[graphstore.IssueRecord](t, call("", "", "create", "After import"))
			if !strings.HasPrefix(fresh.Properties.ID, "legacy-") {
				t.Fatalf("new Issue prefix: %s", fresh.Properties.ID)
			}
			call("", "", "dep", "add", fresh.ID, "legacy-a")
			call("", "", "dep", "add", "legacy-b", fresh.ID)
		})
	}
}

// Synthetic regression for Donna's five-Issue/two-Memory legacy export. The
// process boundary must preserve relates-to as a native Dependency and keep a
// colliding Memory key distinct from the Issue with the same spelling.
func TestGraphPreviewLegacyImportMixedKindsInstalledWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	const fixture = `{"_type":"issue","id":"scratch-8b8","title":"Target"}
{"_type":"issue","id":"scratch-35t","title":"Source","dependencies":[{"issue_id":"scratch-35t","depends_on_id":"scratch-8b8","type":"relates-to","created_by":"source-actor","created_at":"2026-10-23T13:07:20Z","metadata":"{}"},{"issue_id":"scratch-35t","depends_on_id":"scratch-9zj","type":"blocks","metadata":"{}"}]}
{"_type":"issue","id":"scratch-bip","title":"Other target"}
{"_type":"issue","id":"scratch-g5r","title":"Unlinked"}
{"_type":"issue","id":"scratch-9zj","title":"Blocker","dependencies":[{"issue_id":"scratch-9zj","depends_on_id":"scratch-bip","type":"blocks","metadata":"{}"}]}
{"_type":"memory","key":"scratch-8b8","value":"Colliding key, distinct Memory"}
{"_type":"memory","key":"hello","value":"Other Memory"}
`
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			isolateBeadsDirForTest(t)
			work, home := t.TempDir(), t.TempDir()
			const scope = "https://example.invalid/import-mixed-kinds/"
			call := func(input string, args ...string) string {
				t.Helper()
				args = append(args, "--json")
				return graphPatchProcess(t, bd, work, home, strings.NewReader(input), "", 90*time.Second, args...)
			}
			init := []string{"init", "--graph-mode", "link", "--scope-url", scope, "--skip-hooks", "--skip-agents", "--non-interactive"}
			if engine == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for required shared-server import proof")
				}
				init = append(init, "--server", "--external", "--server-host", "127.0.0.1", "--server-port", port, "--server-user", "root")
			}
			call("", init...)
			before := call("", "list", "--all", "--format", "records-json")
			preview := graphMixedResult[graphstore.LegacyImportResult](t, call(fixture, "import", "-", "--dry-run"))
			if !preview.DryRun || preview.Issues != 5 || preview.Memories != 2 || preview.Dependencies != 3 {
				t.Fatalf("preview=%+v", preview)
			}
			if got := call("", "list", "--all", "--format", "records-json"); got != before {
				t.Fatal("dry-run reserved or changed data")
			}
			result := graphMixedResult[graphstore.LegacyImportResult](t, call(fixture, "import", "-"))
			if result.DryRun || !reflect.DeepEqual(result.IDs, preview.IDs) || !reflect.DeepEqual(result.MemoryIDs, preview.MemoryIDs) || !reflect.DeepEqual(result.DependencyIDs, preview.DependencyIDs) {
				t.Fatalf("apply identity maps differ from dry-run: %+v / %+v", result, preview)
			}
			memory := graphMixedResult[graphstore.Record](t, call("", "show", result.MemoryIDs["scratch-8b8"]))
			for pair, kind := range map[string]string{
				"scratch-35t -> scratch-8b8": "relates-to",
				"scratch-35t -> scratch-9zj": "blocks",
				"scratch-9zj -> scratch-bip": "blocks",
			} {
				link := graphMixedResult[graphstore.LinkRecord](t, call("", "link", "show", result.DependencyIDs[pair]))
				ends := strings.Split(pair, " -> ")
				if link.Type != graphstore.DependencyTypeURL(scope) || link.Properties["dep_type"] != kind || link.Source != result.IDs[ends[0]] || link.Target != result.IDs[ends[1]] {
					t.Fatalf("dependency type/kind/endpoints changed for %s: %+v", pair, link)
				}
				if kind == "relates-to" && (link.Attribution.Actor != "source-actor" || link.Attribution.RecordedAt != "2026-10-23T13:07:20Z") {
					t.Fatalf("source dependency attribution changed: %+v", link.Attribution)
				}
				if kind == "blocks" {
					recordedAt, err := time.Parse(time.RFC3339Nano, link.Attribution.RecordedAt)
					if err != nil || recordedAt.IsZero() || link.Attribution.Actor == "" || link.Attribution.Actor != memory.Attribution.Actor {
						t.Fatalf("missing generated dependency attribution: %+v", link.Attribution)
					}
				}
				retained := graphMixedResult[graphstore.LinkRecord](t, call("", "link", "show", link.ID, "--version", link.Revision))
				if !reflect.DeepEqual(retained, link) {
					t.Fatalf("initial retained dependency differs: %+v / %+v", retained, link)
				}
			}
			issueID, memoryID := result.IDs["scratch-8b8"], result.MemoryIDs["scratch-8b8"]
			if issueID != scope+"beads/scratch-8b8" || memoryID == "" || memoryID == issueID || result.MemoryIDs["hello"] != scope+"beads/hello" {
				t.Fatalf("collision identity mapping=%+v", result)
			}
			issue := graphMixedResult[graphstore.IssueRecord](t, call("", "show", issueID, "--format", "graph-json"))
			if memory.ID != memoryID {
				t.Fatalf("Memory read returned wrong identity: %+v", memory)
			}
			if issue.Properties.Title != "Target" || memory.Properties.Title != "scratch-8b8" || memory.Properties.Body != "Colliding key, distinct Memory" || memory.Type != graphstore.MemoryTypeURL(scope) {
				t.Fatalf("colliding Issue/Memory data changed: %+v / %+v", issue, memory)
			}
		})
	}
}

// This separate producer-backed corpus proves that ordinary keys and a
// hyphenated source prefix survive the installed process and engine boundaries.
// Generate its fixture with legacy-import-fixture.py --profile completeness
// --prefix web-app; never substitute hand-edited output for the producer receipt.
func TestGraphPreviewLegacyImportCompletenessInstalledWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	fixture, err := os.ReadFile(filepath.Join("testdata", "legacy-import-complete", "ordinary.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	var issues []*types.Issue
	memories := map[string]string{}
	for _, line := range bytes.Split(fixture, []byte{'\n'}) {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var record struct {
			Kind  string `json:"_type"`
			Key   string `json:"key"`
			Value string `json:"value"`
		}
		if err := json.Unmarshal(line, &record); err != nil {
			t.Fatal(err)
		}
		switch record.Kind {
		case "issue":
			var issue types.Issue
			if err := json.Unmarshal(line, &issue); err != nil {
				t.Fatal(err)
			}
			issues = append(issues, &issue)
		case "memory":
			if _, exists := memories[record.Key]; exists {
				t.Fatalf("producer repeated memory key %q", record.Key)
			}
			memories[record.Key] = record.Value
		default:
			t.Fatalf("unexpected producer record %q", record.Kind)
		}
	}
	if len(issues) != 4 || len(memories) != 16 {
		t.Fatalf("completeness producer corpus: %d Issues, %d memories", len(issues), len(memories))
	}
	for _, key := range []string{"fixture-key", "has space", "has%20space", " leading", "trailing ", "記憶", "%E8%A8%98%E6%86%B6", "100%", ".", "..", "path/part", "double//separator", "back\\slash", "tab\tline\nbreak", "web-app-a1b", "empty-value"} {
		if _, exists := memories[key]; !exists {
			t.Fatalf("producer omitted memory key %q", key)
		}
	}
	if memories["empty-value"] != "" {
		t.Fatal("producer empty value changed")
	}
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			isolateBeadsDirForTest(t)
			work, home := t.TempDir(), t.TempDir()
			const scope = "https://example.invalid/legacy-import-complete/"
			call := func(input, code string, args ...string) string {
				t.Helper()
				if args[0] != "list" {
					args = append(args, "--json")
				}
				return graphPatchProcess(t, bd, work, home, strings.NewReader(input), code, 90*time.Second, args...)
			}
			refuse := func(input, message string, args ...string) {
				t.Helper()
				graphLegacyImportRefusal(t, bd, work, home, input, message, args...)
			}
			init := []string{"init", "--graph-mode", "link", "--scope-url", scope, "--skip-hooks", "--skip-agents", "--non-interactive"}
			if engine == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for required shared-server completeness proof")
				}
				init = append(init, "--server", "--external", "--server-host", "127.0.0.1", "--server-port", port, "--server-user", "root")
			}
			call("", "", init...)
			before := call("", "", "list", "--all", "--format", "records-json")
			for _, invalid := range []struct{ record, message string }{
				{`{"_type":"memory","key":"has space","value":"duplicate"}`, "duplicate memory key"},
				{`{"_type":"memory","key":"new-key","value":null}`, "null is unsupported for legacy string"},
				{`{"_type":"issue","id":"web-api-123","title":"Different source prefix"}`, "requires one Issue ID prefix"},
			} {
				input := strings.TrimRight(string(fixture), "\n") + "\n" + invalid.record + "\n"
				refuse(input, invalid.message, "import", "-")
				if got := call("", "", "list", "--all", "--format", "records-json"); got != before {
					t.Fatal("refused completeness import changed Bead inventory")
				}
			}
			var ambiguous []byte
			for _, source := range issues {
				if source.ID == "web-app-abcdef" {
					var err error
					ambiguous, err = json.Marshal(source)
					if err != nil {
						t.Fatal(err)
					}
				}
			}
			if len(ambiguous) == 0 {
				t.Fatal("producer lacks all-letter hash suffix")
			}
			refuse(string(ambiguous), "source Issue prefix is ambiguous", "import", "-", "--dry-run")
			explicit := graphMixedResult[graphstore.LegacyImportResult](t, call(string(ambiguous), "", "import", "-", "--source-prefix", "web-app", "--dry-run"))
			if explicit.IssuePrefix != "web-app" || explicit.IDs["web-app-abcdef"] != scope+"beads/web-app-abcdef" {
				t.Fatalf("explicit prefix remapped original ID: %+v", explicit)
			}
			refuse(string(ambiguous), "requires one Issue ID prefix", "import", "-", "--source-prefix", "other", "--dry-run")
			preview := graphMixedResult[graphstore.LegacyImportResult](t, call(string(fixture), "", "import", "-", "--dry-run"))
			configuredPreview := graphMixedResult[graphstore.LegacyImportResult](t, call(string(fixture), "", "import", "-", "--source-prefix", "web-app", "--dry-run"))
			if !reflect.DeepEqual(preview, configuredPreview) {
				t.Fatalf("inferred and explicit prefix dry-runs differ: %+v %+v", preview, configuredPreview)
			}
			// Exported counts are projections, not authoritative imported state.
			// Deliberately mutate copies; the producer fixture stays byte-identical.
			projected := bytes.Replace(fixture, []byte(`"dependency_count":1`), []byte(`"dependency_count":71`), 1)
			if bytes.Equal(projected, fixture) {
				t.Fatal("producer corpus lacks expected dependency-count projection")
			}
			wrongCounts := bytes.Replace(projected, []byte(`"comment_count":1`), []byte(`"comment_count":73`), 1)
			if bytes.Equal(wrongCounts, projected) {
				t.Fatal("producer corpus lacks expected comment-count projection")
			}
			countPreview := graphMixedResult[graphstore.LegacyImportResult](t, call(string(wrongCounts), "", "import", "-", "--source-prefix", "web-app", "--dry-run"))
			if !reflect.DeepEqual(preview, countPreview) {
				t.Fatalf("exported projections changed import result: %+v %+v", preview, countPreview)
			}
			lines := bytes.Split(bytes.TrimSpace(fixture), []byte{'\n'})
			for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
				lines[i], lines[j] = lines[j], lines[i]
			}
			reversed := string(bytes.Join(lines, []byte{'\n'})) + "\n"
			reordered := graphMixedResult[graphstore.LegacyImportResult](t, call(reversed, "", "import", "-", "--dry-run"))
			if !preview.DryRun || preview.IssuePrefix != "web-app" || preview.Issues != len(issues) || preview.Memories != len(memories) || preview.Dependencies != 1 || preview.Comments != 1 || !reflect.DeepEqual(preview, reordered) {
				t.Fatalf("dry-run counts or order-dependent remapping: %+v %+v", preview, reordered)
			}
			if got := call("", "", "list", "--all", "--format", "records-json"); got != before {
				t.Fatal("completeness dry-run reserved data")
			}
			result := graphMixedResult[graphstore.LegacyImportResult](t, call(reversed, "", "import", "-", "--source-prefix", "web-app"))
			if result.DryRun || len(result.DependencyIDs) != 1 || result.IssuePrefix != preview.IssuePrefix || !reflect.DeepEqual(result.IDs, preview.IDs) || !reflect.DeepEqual(result.MemoryIDs, preview.MemoryIDs) || !reflect.DeepEqual(result.DependencyIDs, preview.DependencyIDs) {
				t.Fatalf("apply did not preserve dry-run identity maps: %+v", result)
			}
			allocated := map[string]bool{}
			for _, source := range issues {
				id := result.IDs[source.ID]
				if id != scope+"beads/"+source.ID {
					t.Fatalf("Issue identity changed: %q -> %q", source.ID, id)
				}
				allocated[id] = true
				got := graphMixedResult[graphstore.IssueRecord](t, call("", "", "show", id, "--format", "graph-json"))
				want := *source
				want.Metadata, want.Dependencies, want.Comments = nil, nil, nil
				want.IsBlocked = false
				if got.Properties == nil {
					t.Fatalf("Issue %s has no properties in graph-json output", source.ID)
				}
				actual := *got.Properties
				actual.IsBlocked = false
				want.Labels = append([]string(nil), want.Labels...)
				actual.Labels = append([]string(nil), actual.Labels...)
				sort.Strings(want.Labels)
				sort.Strings(actual.Labels)
				wantJSON, err := json.Marshal(want)
				if err != nil {
					t.Fatal(err)
				}
				gotJSON, err := json.Marshal(actual)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(wantJSON, gotJSON) {
					t.Fatalf("Issue %s lost exported fields:\nwant %s\ngot %s", source.ID, wantJSON, gotJSON)
				}
				metadata := source.Metadata
				if len(metadata) == 0 {
					metadata = []byte(`{}`)
				}
				canonical, err := graph.CanonicalizeJSON(metadata)
				if err != nil || got.Revision == "" || !bytes.Equal(canonical, got.Metadata) {
					t.Fatalf("Issue %s metadata changed: %s (%v)", source.ID, got.Metadata, err)
				}
				wantOwned := 0
				if source.ID == "web-app-b2c" {
					wantOwned = 1
				}
				if len(got.Owned) != wantOwned {
					t.Fatalf("Issue %s dependency count: got %d want %d", source.ID, len(got.Owned), wantOwned)
				}
				if wantOwned == 1 {
					var link graphstore.LinkRecord
					if err := json.Unmarshal(got.Owned[0], &link); err != nil {
						t.Fatal(err)
					}
					if link.ID == "" || result.DependencyIDs["web-app-b2c -> web-app-a1b"] != link.ID || link.Type != graphstore.DependencyTypeURL(scope) || link.Source != id || link.Target != result.IDs["web-app-a1b"] || link.Properties["dep_type"] != "blocks" {
						t.Fatalf("Issue %s dependency lost type or endpoints: %+v", source.ID, link)
					}
				}
				retained := graphMixedResult[graphstore.IssueRecord](t, call("", "", "show", id, "--version", got.Revision))
				if !reflect.DeepEqual(got.Properties, retained.Properties) || !bytes.Equal(got.Metadata, retained.Metadata) || !reflect.DeepEqual(got.Owned, retained.Owned) {
					t.Fatalf("Issue %s initial retained state differs", source.ID)
				}
				var comments []*types.Comment
				if err := json.Unmarshal([]byte(call("", "", "comments", id)), &comments); err != nil {
					t.Fatal(err)
				}
				// The imported native comment feed keeps the source Issue's
				// public ID even when its graph allocation path differs.
				wantComments := make([]*types.Comment, len(source.Comments))
				for i, original := range source.Comments {
					comment := *original
					comment.IssueID = source.ID
					wantComments[i] = &comment
				}
				if !reflect.DeepEqual(wantComments, comments) && !(len(wantComments) == 0 && len(comments) == 0) {
					t.Fatalf("Issue %s comment feed changed: %+v", source.ID, comments)
				}
			}
			for key, value := range memories {
				id := result.MemoryIDs[key]
				if !strings.HasPrefix(id, scope) || allocated[id] {
					t.Fatalf("memory %q has absent or colliding graph identity %q", key, id)
				}
				if err := graph.ValidateBeadPath(strings.TrimPrefix(id, scope)); err != nil {
					t.Fatalf("memory %q has invalid graph path: %v", key, err)
				}
				allocated[id] = true
				got := graphMixedResult[graphstore.Record](t, call("", "", "show", id))
				retained := graphMixedResult[graphstore.Record](t, call("", "", "show", id, "--version", got.Revision))
				if got.ID != id || got.Revision == "" || got.Properties.Title != key || got.Properties.Body != value || !reflect.DeepEqual(got.Properties, retained.Properties) {
					t.Fatalf("memory %q lost original key/value or retained state: %+v %+v", key, got, retained)
				}
			}
			for _, key := range []string{"fixture-key", "has%20space", "%E8%A8%98%E6%86%B6", "path/part", "empty-value"} {
				if result.MemoryIDs[key] != scope+"beads/"+key {
					t.Fatalf("safe free memory path changed for %q", key)
				}
			}
			inventory := call("", "", "list", "--all", "--format", "records-json")
			listed, more := graphListDecode(t, inventory)
			if more || len(listed) != len(allocated) {
				t.Fatalf("completeness Bead inventory: got %d, more=%t, want %d", len(listed), more, len(allocated))
			}
			seen := map[string]bool{}
			for _, bead := range listed {
				if !allocated[bead.ID] || seen[bead.ID] {
					t.Fatalf("unexpected or repeated imported Bead %q", bead.ID)
				}
				seen[bead.ID] = true
			}
			// Inventory is a Bead projection, not a database-wide snapshot.
			// Full catalog/config/history rollback belongs to lower store tests;
			// sample all three Resource families across the installed boundary.
			histories := map[string]string{}
			for _, id := range []string{result.IDs["web-app-b2c"], result.MemoryIDs["has space"], result.DependencyIDs["web-app-b2c -> web-app-a1b"]} {
				histories[id] = call("", "", "versions", id)
			}
			for _, dry := range []bool{false, true} {
				args := []string{"import", "-", "--source-prefix", "web-app"}
				if dry {
					args = append(args, "--dry-run")
				}
				refuse(string(fixture), "requires an empty fresh graph workspace", args...)
				if call("", "", "list", "--all", "--format", "records-json") != inventory {
					t.Fatal("repeat completeness import changed Bead inventory")
				}
				for id, versions := range histories {
					if call("", "", "versions", id) != versions {
						t.Fatalf("repeat completeness import changed Resource history: %s", id)
					}
				}
			}
			fresh := graphMixedResult[graphstore.IssueRecord](t, call("", "", "create", "After hyphenated prefix import"))
			if !strings.HasPrefix(fresh.Properties.ID, "web-app-") {
				t.Fatalf("hyphenated source prefix was not adopted: %s", fresh.Properties.ID)
			}
			call("", "", "dep", "add", fresh.ID, "web-app-a1b")
			call("", "", "dep", "add", "web-app-b2c", fresh.ID)
		})
	}
}

// Pin the failed admission stage as well as the standard graph refusal shape.
// Kept import-local so a stronger oracle does not alter other workflow helpers.
func graphLegacyImportRefusal(t *testing.T, bd, work, home, input, message string, args ...string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
	defer cancel()
	cmd := graphMemoryReadCommand(ctx, bd, work, home, append(args, "--json")...)
	cmd.Stdin = strings.NewReader(input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	var exit *exec.ExitError
	var diagnostic struct {
		Code      string `json:"code"`
		Message   string `json:"message"`
		Retryable *bool  `json:"retryable"`
	}
	if ctx.Err() != nil || !errors.As(err, &exit) || exit.ExitCode() != 2 || stdout.Len() != 0 || json.Unmarshal(stderr.Bytes(), &diagnostic) != nil || diagnostic.Code != "invalid_properties" || diagnostic.Retryable == nil || *diagnostic.Retryable || !strings.Contains(diagnostic.Message, message) {
		t.Fatalf("import refusal %v want %q: err=%v stdout=%s stderr=%s", args, message, err, stdout.String(), stderr.String())
	}
}
