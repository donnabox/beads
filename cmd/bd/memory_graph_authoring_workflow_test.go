//go:build cgo

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/httpapi/bdpwire"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

type memoryAuthoredLink struct {
	Link           bdpwire.LinkRecord `json:"link"`
	Source         bdpwire.BeadRecord `json:"source"`
	Changed        bool               `json:"changed"`
	ReplacedSource json.RawMessage    `json:"replacedSource"`
}

// Every invocation is an installed process; normal stores/allocators/Type
// validators own all writes. No SQL seeding or command-global test injection.
func TestMemoryLinkAuthoringInstalledWorkflow(t *testing.T) {
	isolateBeadsDirForTest(t)
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			const scope = "https://example.invalid/memory-authoring/"
			work, home := graphListWorkspace(t, bd, engine, scope)
			call := func(args ...string) string { t.Helper(); return graphPolicyCLI(t, bd, work, home, nil, "", args...) }
			old := func(args ...string) string { t.Helper(); return call(append(args, "--json")...) }
			snapshot := func(id string) string { t.Helper(); return old("show", scope+id) }
			for _, id := range []string{"a", "b", "c", "d"} {
				old("create", id, "--id", id, "--type", "constraint")
			}
			native := graphMixedResult[graphstore.IssueRecord](t, old("create", "native", "--graph-path", "beads/i", "--type", "task"))
			decode := func(raw string) memoryAuthoredLink {
				t.Helper()
				var frame struct {
					Schema  int                `json:"schema_version"`
					Pin     string             `json:"bdp_spec"`
					Kind    string             `json:"result_kind"`
					Outcome string             `json:"outcome"`
					Data    memoryAuthoredLink `json:"data"`
				}
				if err := json.Unmarshal([]byte(raw), &frame); err != nil || frame.Schema != 2 || frame.Pin != bdpwire.Pin || frame.Kind != "mutation" || frame.Outcome != "success" || !frame.Data.Changed || frame.Data.Link.ID == "" || frame.Data.Link.Revision == "" || frame.Data.Source.Revision == "" {
					t.Fatalf("wrong mutation contract: %s %v", raw, err)
				}
				if strings.Contains(raw, `"version":`) || strings.Contains(raw, `"actor":`) {
					t.Fatalf("storage DTO leaked: %s", raw)
				}
				return frame.Data
			}
			create := func(verb, src, target string, extra ...string) memoryAuthoredLink {
				t.Helper()
				args := append([]string{"memory", verb, src, "--to", target, "--json"}, extra...)
				raw := call(args...)
				if verb == "supersede" {
					t.Logf("ACTUAL_EXAMPLE provider=%s args=%q output=%s", engine, args, raw)
				}
				return decode(raw)
			}
			refuse := func(code string, exit int, args ...string) {
				t.Helper()
				ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
				defer cancel()
				command := graphMemoryReadCommand(ctx, bd, work, home, append(args, "--json")...)
				var out, stderr bytes.Buffer
				command.Stdout, command.Stderr = &out, &stderr
				err := command.Run()
				var process *exec.ExitError
				if !errors.As(err, &process) || process.ExitCode() != exit || !strings.Contains(stderr.String(), code) {
					t.Fatalf("wrong refusal %v: %s %s %v", args, &out, &stderr, err)
				}
				var frame struct {
					Outcome string          `json:"outcome"`
					Data    json.RawMessage `json:"data"`
				}
				if err := json.Unmarshal(out.Bytes(), &frame); err != nil || frame.Outcome != "failure" || len(frame.Data) != 0 {
					t.Fatalf("refusal not canonical %s %s %v", &out, &stderr, err)
				}
			}
			// Configured-unused native declarations must not break Memory Link
			// projection; retained custom bindings remain usable after config edits.
			graphListConfig(t, work, "types.custom: \"review-class\"\n")
			create("reference", "a", "b", "--id=configured-native-context")
			custom := graphMixedResult[graphstore.IssueRecord](t, old("create", "Custom native", "--type", "review-class"))
			graphListConfig(t, work, "types.custom: \"\"\n")
			create("reference", custom.ID, "b", "--id=retained-native-context")
			beforeA, beforeB := snapshot("a"), snapshot("b")
			beforeLinks := old("link", "list", "--incident-to", "a")
			for _, op := range memoryLinkOperations {
				for _, flags := range [][]string{
					{"--from=a", "--from=a", "--to=b"},
					{"--from=a", "--from=c", "--to=b"},
					{"--from=a", "--to=b", "--to=b"},
					{"--from=a", "--to=b", "--to=c"},
				} {
					refuse("invalid_selector", 2, append([]string{"memory", op.name}, flags...)...)
				}
			}
			if snapshot("a") != beforeA || snapshot("b") != beforeB || old("link", "list", "--incident-to", "a") != beforeLinks {
				t.Fatal("duplicate endpoint flags changed graph state")
			}
			first := create("reference", "a", "b", "--title=Context", "--body=Why", "--status=reviewed", "--properties", `{"counter":9007199254740994,"nested":{"x":[true,null]}}`, "--metadata", `{"private":{"tag":"retained"}}`, "--actor=memory-author")
			if first.Link.Type != scope+"references" || first.Link.Source.URI != scope+"a" || first.Link.Target.URI != scope+"b" || string(first.Link.Properties["counter"]) != "9007199254740994" || string(first.Link.Properties["body"]) != `"Why"` || first.Link.Attribution == nil || first.Link.Attribution.Principal != "memory-author" || string(first.Link.Metadata["private"]) != `{"tag":"retained"}` {
				t.Fatalf("authored fields lost: %+v", first)
			}
			parallel := create("reference", "a", "b")
			if first.Link.ID == parallel.Link.ID {
				t.Fatal("parallel references deduplicated")
			}
			retired := create("supersede", "a", "b", "--id=retirement")
			if retired.Link.Source.URI != scope+"a" || retired.Link.Target.URI != scope+"b" || retired.Link.Type != scope+"superseded-by" {
				t.Fatal(retired)
			}
			before := snapshot("a")
			refuse("invalid_properties", 2, "memory", "supersede", "a", "--to=b", "--id=another-retirement")
			if snapshot("a") != before {
				t.Fatal("repeated supersede revised source")
			}
			parent := create("defines", "b", "c", "--id=definition")
			if parent.Link.Type != scope+"defines" || parent.Source.ID != scope+"b" || len(parent.ReplacedSource) == 0 {
				t.Fatal(parent)
			}
			beforeC := snapshot("c")
			beforeD := snapshot("d")
			refuse("invalid_properties", 2, "memory", "defines", "d", "--to=c", "--id=second-parent")
			refuse("invalid_properties", 2, "memory", "defines", "c", "--to=b", "--id=defines-cycle")
			if snapshot("c") != beforeC || snapshot("d") != beforeD {
				t.Fatal("ownership/cycle refusal changed sources")
			}
			// A cycle with a Requires edge is legal; only the Defines subgraph is a tree.
			required := create("requires", "c", "b", "--id=mixed-requirement")
			if required.Link.Type != scope+"requires" {
				t.Fatal(required)
			}
			governed := create("subject-to", native.ID, "b", "--id=issue-rule")
			nativeTaskType, err := graphstore.NativeIssueTypeURL(scope, "task")
			if err != nil {
				t.Fatal(err)
			}
			if governed.Link.Type != scope+"subject-to" || governed.Source.Type != nativeTaskType || governed.Link.Source.URI != native.ID {
				t.Fatal(governed)
			}
			// The stale guard is enforced atomically by the same writer, before either
			// a Link identity or source revision is published.
			advanced := create("requires", "b", "d", "--id=advance-owner")
			if advanced.Source.Revision == parent.Source.Revision {
				t.Fatal("owned fixture did not advance source")
			}
			before = snapshot("b")
			refuse("revision_conflict", 4, "memory", "reference", "b", "--to=d", "--id=guarded", "--if-source-revision="+parent.Source.Revision)
			if snapshot("b") != before {
				t.Fatal("stale source guard changed source")
			}
			// Invalid properties/selectors are rejected before the writer. Read-only
			// refusal also preserves the exact source postimage.
			before = snapshot("d")
			for _, guard := range []string{string([]byte{0xff}), strings.Repeat("x", graphstore.PreviewVersionTokenLimit+1)} {
				refuse("invalid_selector", 2, "memory", "reference", "d", "--to=b", "--if-source-revision="+guard)
			}
			for _, flag := range []string{"title", "body", "status"} {
				refuse("invalid_properties", 2, "memory", "reference", "d", "--to=b", "--"+flag+"="+string([]byte{0xff}))
			}
			refuse("invalid_properties", 2, "memory", "reference", "d", "--to=b", "--title=flag", "--properties", `{"title":"duplicate"}`)
			refuse("invalid_properties", 2, "memory", "reference", "d", "--to=b", "--properties", `{"n":9007199254740993}`)
			refuse("invalid_selector", 2, "memory", "reference", "d", "--to=b", "--output-format=tree")
			refuse("invalid_selector", 2, "memory", "reference", "d")
			refuse("invalid_selector", 2, "memory", "supersede", "d")
			refuse("invalid_selector", 2, "memory", "reference", "d", "other", "--to=b")
			refuse("invalid_selector", 2, "memory", "reference", "d", "--to=https://foreign.invalid/beads/b")
			refuse("invalid_selector", 2, "memory", "reference", "d", "--to=b@old", "--target-revision=other")
			refuse("invalid_selector", 2, "memory", "reference", "d", "--to=b", "--to=c")
			refuse("invalid_selector", 2, "memory", "reference", "d", "--to=b", "--include=/link/id")
			refuse("permission_denied", 5, "memory", "reference", "d", "--to=b", "--readonly")
			if snapshot("d") != before {
				t.Fatal("pre-write refusal changed source")
			}
			// Human output must identify the actual created Link; raw JSON is the
			// operation result, and quiet success never suppresses explicit JSON.
			human := call("memory", "reference", "d", "--to=b", "--id=human-link")
			t.Logf("ACTUAL_EXAMPLE provider=%s args=%q output=%s", engine, []string{"memory", "reference", "d", "--to=b", "--id=human-link"}, human)
			if strings.TrimSpace(human) != "human-link [references] d → b" {
				t.Fatalf("human confirmation lost Link identity: %s", human)
			}
			var raw memoryAuthoredLink
			output := call("memory", "reference", "d", "--to=b", "--id=raw-link", "--json", "--omit=envelope", "--quiet")
			if err := json.Unmarshal([]byte(output), &raw); err != nil || !raw.Changed || raw.Link.ID != scope+"raw-link" {
				t.Fatalf("raw mutation: %s %v", output, err)
			}
		})
	}
}
