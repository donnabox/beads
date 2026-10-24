//go:build cgo

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/httpapi/bdpwire"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// These are separate installed processes over persisted state, not formatter
// examples. Both engines use disposable workspaces; the server is opt-in.
func TestMemoryGraphInstalledWorkflow(t *testing.T) {
	isolateBeadsDirForTest(t)
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			const scope = "https://example.invalid/memory-command/"
			work, home := graphListWorkspace(t, bd, engine, scope)
			call := func(args ...string) string { t.Helper(); return graphPolicyCLI(t, bd, work, home, nil, "", args...) }
			legacy := func(args ...string) string { t.Helper(); return call(append(args, "--json")...) }
			for _, id := range []string{"a", "c"} {
				legacy("create", id, "--id", id, "--type", "task")
			}
			for _, id := range []string{"b", "d", "isolated"} {
				legacy("create", id, "--id", id, "--type", "types/memory", "--properties", `{"body":"Content","extra":{"keep":true}}`)
			}
			for _, e := range [][3]string{{"a", "b", "ab"}, {"b", "c", "bc"}, {"c", "d", "cd"}, {"a", "b", "parallel"}, {"b", "a", "cycle"}} {
				legacy("link", "create", e[0], e[1], "--id", e[2], "--type", "types/references", "--properties", `{"title":"Authored relationship","status":"active"}`)
			}
			decode := func(raw string) memoryGraphResult {
				t.Helper()
				var env struct {
					Schema  int               `json:"schema_version"`
					BDP     string            `json:"bdp_spec"`
					Kind    string            `json:"result_kind"`
					Outcome string            `json:"outcome"`
					Data    memoryGraphResult `json:"data"`
				}
				if err := json.Unmarshal([]byte(raw), &env); err != nil || env.Schema != 2 || env.BDP != bdpwire.Pin || env.Kind != "graph" || env.Outcome != "success" {
					t.Fatalf("wrong graph contract: %s (%v)", raw, err)
				}
				return env.Data
			}
			read := func(verb string, args ...string) memoryGraphResult {
				t.Helper()
				return decode(call(append(append([]string{"memory", verb}, args...), "--json")...))
			}
			mixed := read("references", "a", "--depth=3")
			if len(mixed.Beads) != 4 || len(mixed.Links) != 5 || !mixed.Traversal.Complete || len(mixed.Traversal.Roots) != 1 {
				t.Fatalf("mixed Issue/Memory traversal: %+v", mixed)
			}
			for _, l := range mixed.Links {
				if l.Source.URI == "" || l.Target.URI == "" || l.Revision == "" || l.Type != scope+"types/references" {
					t.Fatalf("lost canonical Link: %+v", l)
				}
			}
			for _, b := range mixed.Beads {
				if b.Type == scope+"types/memory" && !bytes.Contains(b.Properties["extra"], []byte(`"keep":true`)) {
					t.Fatalf("lost complete properties: %+v", b)
				}
			}
			shallow := read("references", "a")
			if len(shallow.Beads) != 2 || len(shallow.Links) != 3 || shallow.Traversal.Complete || !slices.Equal(shallow.Traversal.Roots[0].Frontier, []string{scope + "beads/b"}) {
				t.Fatalf("induced reverse/parallel/frontier: %+v", shallow)
			}
			inverse := read("referenced-by", "b")
			if len(inverse.Beads) != 2 || len(inverse.Links) != 3 || inverse.Traversal.Direction != "in" {
				t.Fatalf("inverse context: beads=%d links=%d traversal=%+v", len(inverse.Beads), len(inverse.Links), inverse.Traversal)
			}
			multi := read("references", "a", "b", "--depth=1")
			if len(multi.Traversal.Roots) != 2 || len(multi.Beads) != 3 || len(multi.Links) != 4 || len(multi.Traversal.Roots[0].Beads) != 2 || len(multi.Traversal.Roots[1].Beads) != 3 {
				t.Fatalf("independent multi-root depth: %+v", multi)
			}
			zero := read("references", "a", "--depth=0")
			if len(zero.Beads) != 1 || len(zero.Links) != 0 || zero.Traversal.Roots[0].Revision == "" {
				t.Fatalf("zero-hop root lost: %+v", zero)
			}
			isolated := read("references", "isolated")
			if len(isolated.Beads) != 1 || len(isolated.Links) != 0 || !isolated.Traversal.Complete {
				t.Fatalf("isolated root: %+v", isolated)
			}
			var raw memoryGraphResult
			if err := json.Unmarshal([]byte(call("memory", "references", "a", "--json", "--output-format=json", "--omit=envelope")), &raw); err != nil || raw.Scope != scope || len(raw.Links) != 3 {
				t.Fatalf("JSON aliases/raw projection: %+v %v", raw, err)
			}
			human := call("memory", "references", "a", "b")
			for _, want := range []string{"└──", "references", "(links/ab)", "(links/parallel)", "(cycle)", "3 Beads, 4 Links across 2 views"} {
				if !strings.Contains(human, want) {
					t.Fatalf("human forest lost %q:\n%s", want, human)
				}
			}
			if explicit := call("memory", "references", "a", "b", "--output-format=tree"); explicit != human {
				t.Fatalf("explicit tree differs:\n%s\n%s", explicit, human)
			}
			detail := call("memory", "references", "a", "--output-format=detail")
			for _, want := range []string{"scope", "traversal", "memory-command"} {
				if !strings.Contains(detail, want) {
					t.Fatalf("detail lost %s: %s", want, detail)
				}
			}
			for _, format := range []string{"tree", "detail"} {
				if output := call("memory", "references", "a", "--output-format="+format, "--quiet"); output != "" {
					t.Fatalf("quiet %s wrote output: %s", format, output)
				}
			}
			if !strings.Contains(human, "[types/references]") || !strings.Contains(human, "beads/a") || !strings.Contains(human, "links/ab") || strings.Contains(human, scope) {
				t.Fatalf("literal kind segments lost or Scope prefix retained:\n%s", human)
			}
			legacy("link", "create", "b", "d", "--type", "types/superseded-by", "--id", "retirement")
			successor := read("successor", "b")
			if len(successor.Beads) != 2 || len(successor.Links) != 1 || (successor.Traversal.Roots[0].Depth == nil || *successor.Traversal.Roots[0].Depth != 1000) {
				t.Fatalf("replacement chain: %+v", successor)
			}
			supersedes := read("supersedes", "d")
			if len(supersedes.Beads) != 2 || len(supersedes.Links) != 1 || (supersedes.Traversal.Roots[0].Depth == nil || *supersedes.Traversal.Roots[0].Depth != 1) {
				t.Fatalf("replacement inverse: %+v", supersedes)
			}
			current := graphMixedResult[graphstore.Record](t, legacy("show", "d"))
			legacy("link", "create", "isolated", "d", "--type", "types/references", "--id", "pin", "--target-revision", current.Revision)
			pinned := read("references", "isolated")
			if len(pinned.Links) != 1 || pinned.Links[0].Target.Revision != current.Revision {
				t.Fatalf("current pin erased: %+v", pinned)
			}
			legacy("update", "d", "--properties", `{"body":"Next revision"}`)
			for _, tc := range []struct {
				args []string
				code string
			}{
				{[]string{"references", "missing"}, "not_found"},
				{[]string{"references", "isolated"}, "capability_unavailable"},
				{[]string{"references", "a", "--max-nodes=1"}, "capability_unavailable"},
				{[]string{"references", "a", "--include=/beads"}, "capability_unavailable"},
				{[]string{"references", "a", "--include=/properties/title"}, "invalid_selector"},
				{[]string{"successor", "a"}, "invalid_properties"},
				{[]string{"references", "b@old"}, "capability_unavailable"},
			} {
				for _, raw := range []bool{false, true} {
					args := append(append([]string{"memory"}, tc.args...), "--json")
					if raw {
						args = append(args, "--omit=envelope")
					}
					ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
					command := graphMemoryReadCommand(ctx, bd, work, home, args...)
					var stdout, stderr bytes.Buffer
					command.Stdout, command.Stderr = &stdout, &stderr
					err := command.Run()
					cancel()
					if err == nil || !strings.Contains(stderr.String(), tc.code) {
						t.Fatalf("missing refusal %v: err=%v stdout=%s stderr=%s", args, err, &stdout, &stderr)
					}
					if raw {
						if stdout.Len() != 0 {
							t.Fatalf("raw failure emitted output: %s", &stdout)
						}
						continue
					}
					var frame struct {
						Outcome string `json:"outcome"`
						Error   struct {
							Problem struct {
								Code string `json:"code"`
							} `json:"problem"`
						} `json:"error"`
						Data json.RawMessage `json:"data"`
					}
					if err := json.Unmarshal(stdout.Bytes(), &frame); err != nil || frame.Outcome != "failure" || frame.Error.Problem.Code != tc.code || (len(frame.Data) > 0 && string(frame.Data) != "null") {
						t.Fatalf("bad atomic failure: %s (%v)", &stdout, err)
					}
				}
			}
		})
	}
}
