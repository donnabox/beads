//go:build cgo

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/configfile"
	"github.com/steveyegge/beads/internal/httpapi/bdpwire"
	"github.com/steveyegge/beads/internal/httpapi/graphread"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

var updateCanonicalShowCorpus = flag.Bool("canonical-show-update", false, "regenerate canonical show corpus from real CLI output")

func TestCanonicalShowWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			const scope = "https://example.invalid/canonical-show/"
			initArgs := []string{"init", "--graph-mode", "link", "--scope-url", scope, "--skip-hooks", "--skip-agents", "--non-interactive"}
			if engine == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for ordinary shared-server canonical output qualification")
				}
				initArgs = append(initArgs, "--server", "--external", "--server-host", "127.0.0.1", "--server-port", port, "--server-user", "root")
			}
			call := func(args ...string) string { t.Helper(); return graphPolicyCLI(t, bd, work, home, nil, "", args...) }
			legacy := func(args ...string) json.RawMessage {
				t.Helper()
				return graphMixedResult[json.RawMessage](t, call(append(args, "--json")...))
			}
			legacy(initArgs...)
			var initial graphstore.Record
			if err := json.Unmarshal(legacy("remember", "Big body & <literal>", "--id", "beads/plan", "--title", "Plan"), &initial); err != nil {
				t.Fatal(err)
			}
			legacy("create", "Target", "--id", "beads/work", "--type", "types/constraint")
			var receipt struct {
				Link graphstore.LinkRecord `json:"link"`
			}
			if err := json.Unmarshal(legacy("link", "beads/plan", "beads/work", "--type", "types/subject-to", "--id", "links/note", "--properties", `{"note":"unchanged"}`), &receipt); err != nil {
				t.Fatal(err)
			}
			linkPath := strings.TrimPrefix(receipt.Link.ID, scope)
			var linked graphstore.Record
			if err := json.Unmarshal(legacy("show", scope+"beads/plan"), &linked); err != nil || len(linked.Owned) != 1 {
				t.Fatalf("legacy owned result changed: %+v %v", linked, err)
			}
			decode := func(out string) json.RawMessage {
				t.Helper()
				var env struct {
					Schema  int             `json:"schema_version"`
					BDP     string          `json:"bdp_spec"`
					Kind    string          `json:"result_kind"`
					Outcome string          `json:"outcome"`
					Data    json.RawMessage `json:"data"`
				}
				if err := json.Unmarshal([]byte(out), &env); err != nil {
					t.Fatalf("bad envelope: %s: %v", out, err)
				}
				if env.Schema != 2 || env.BDP != bdpwire.Pin || env.Kind != "resource" || env.Outcome != "success" {
					t.Fatalf("wrong contract: %s", out)
				}
				return env.Data
			}
			equalJSON := func(a, b []byte) {
				t.Helper()
				var x, y any
				if json.Unmarshal(a, &x) != nil || json.Unmarshal(b, &y) != nil || !reflect.DeepEqual(x, y) {
					t.Fatalf("JSON mismatch:\n%s\n%s", a, b)
				}
			}
			beadOutput := call("show", scope+"beads/plan", "--output-format=json")
			canonicalShowGolden(t, "bead-owned", beadOutput)
			bead := decode(beadOutput)
			linkOutput := call("link", "show", receipt.Link.ID, "--output-format=json")
			canonicalShowGolden(t, "link", linkOutput)
			link := decode(linkOutput)
			equalJSON(link, decode(call("show", receipt.Link.ID, "--output-format=json")))
			equalJSON(link, decode(call("show", "note", "--output-format=json")))
			equalJSON(link, decode(call("link", "show", "note", "--output-format=json")))
			equalJSON(legacy("show", receipt.Link.ID), legacy("link", "show", "note"))
			// Compare public CLI payloads to the real protocol projector, with store
			// lifetime entirely outside child processes.
			cfg, err := configfile.Load(filepath.Join(work, ".beads"))
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			options := graphOptions(cfg)
			options.DataDir = filepath.Join(work, ".beads", "embeddeddolt")
			store, err := graphstore.OpenExisting(ctx, options)
			if err != nil {
				t.Fatal(err)
			}
			reader := graphread.New(store)
			for _, v := range []struct {
				path string
				got  []byte
			}{{"beads/plan", bead}, {linkPath, link}} {
				expected, err := reader.Resource(ctx, v.path)
				if err != nil {
					t.Fatal(err)
				}
				raw, err := json.Marshal(expected)
				if err != nil {
					t.Fatal(err)
				}
				equalJSON(v.got, raw)
			}
			if err := store.Close(); err != nil {
				t.Fatal(err)
			}
			var resource bdpwire.BeadRecord
			if err := json.Unmarshal(bead, &resource); err != nil {
				t.Fatal(err)
			}
			if len(resource.OwnedLinks[receipt.Link.Type]) != 1 {
				t.Fatalf("complete inline owned Links lost: %s", bead)
			}
			equalJSON(bead, []byte(call("show", scope+"beads/plan", "--output-format=json", "--omit=envelope")))
			projected := decode(call("show", scope+"beads/plan", "--output-format=json", "--include=/properties/title", "--omit=/metadata"))
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(projected, &fields); err != nil {
				t.Fatal(err)
			}
			if len(fields) != 4 || fields["id"] == nil || fields["type"] == nil || fields["revision"] == nil || string(fields["properties"]) != `{"title":"Plan"}` {
				t.Fatalf("projection: %s", projected)
			}
			// Historical owner output must retain its old owned state, never today's.
			old := decode(call("show", "["+scope+"beads/plan]@="+initial.Revision, "--output-format=json"))
			var oldBead bdpwire.BeadRecord
			if err := json.Unmarshal(old, &oldBead); err != nil || oldBead.Revision != initial.Revision || len(oldBead.OwnedLinks) != len(resource.OwnedLinks) {
				t.Fatalf("historical state changed: %s %v", old, err)
			}
			for typ := range resource.OwnedLinks {
				group, present := oldBead.OwnedLinks[typ]
				if !present || len(group) != 0 {
					t.Fatalf("retained owner must preserve selected empty groups without current Links: %s", old)
				}
			}
			equalJSON(old, decode(call("show", scope+"beads/plan", "--revision", initial.Revision, "--output-format=json")))
			detail := call("show", scope+"beads/plan", "--output-format=detail", "--omit=/ownedLinks")
			if !strings.Contains(detail, "Big body & <literal>") {
				t.Fatalf("human detail escaped or lost body: %s", detail)
			}
			// Canonical failures have separate stdout frames and stderr diagnostics.
			for _, failure := range []struct {
				args      []string
				code      string
				outside   bool
				ambiguous bool
			}{
				{[]string{"show", "beads/missing"}, "not_found", false, false},
				{[]string{"show", scope + "beads/plan", "--json"}, "invalid_selector", false, true},
				{[]string{"show", scope + "beads/plan", "--output-format=json", "--json=false"}, "invalid_selector", false, true},
				{[]string{"show", scope + "beads/plan", "--graph-mode=invalid"}, "invalid_selector", false, false},
				{[]string{"show", scope + "beads/plan", "--omit=/revision"}, "invalid_selector", false, false},
				{[]string{"show", scope + "beads/plan", "--revision=first", "--revision=second"}, "invalid_selector", false, false},
				{[]string{"link", "show", receipt.Link.ID, "--revision=first", "--revision=second"}, "invalid_selector", false, false},
				{[]string{"link", "show"}, "invalid_selector", false, false},
				{[]string{"show", scope + "beads/plan"}, "capability_unavailable", true, false},
			} {
				for _, raw := range []bool{false, true} {
					args := append(append([]string{}, failure.args...), "--output-format=json")
					if raw {
						args = append(args, "--omit=envelope")
					}
					cmd := graphMemoryReadCommand(ctx, bd, work, home, args...)
					cmd.Dir = work
					if failure.outside {
						cmd.Dir = t.TempDir()
					}
					var stdout, stderr bytes.Buffer
					cmd.Stdout = &stdout
					cmd.Stderr = &stderr
					if err := cmd.Run(); err == nil {
						t.Fatal("missing Resource succeeded")
					}
					if !strings.Contains(stderr.String(), failure.code) {
						t.Fatalf("missing %s diagnostic for %v: %s", failure.code, args, &stderr)
					}
					if raw || failure.ambiguous {
						if stdout.Len() != 0 {
							t.Fatalf("raw failure content: %s", &stdout)
						}
					} else {
						var frame map[string]json.RawMessage
						if err := json.Unmarshal(stdout.Bytes(), &frame); err != nil || string(frame["outcome"]) != `"failure"` || frame["error"] == nil {
							t.Fatalf("missing failure envelope: %s %v", &stdout, err)
						}
					}
				}
			}
		})
	}
}

// Only protocol revision tokens are normalized. Properties, metadata, identities,
// References and the actual BDP pin are kept exactly as emitted by the process.
func canonicalShowGolden(t *testing.T, name, output string) {
	t.Helper()
	decoder := json.NewDecoder(strings.NewReader(output))
	decoder.UseNumber()
	var frame map[string]any
	if err := decoder.Decode(&frame); err != nil {
		t.Fatal(err)
	}
	data := frame["data"].(map[string]any)
	data["revision"] = "REVISION"
	if groups, ok := data["ownedLinks"].(map[string]any); ok {
		for _, group := range groups {
			for _, item := range group.([]any) {
				item.(map[string]any)["revision"] = "REVISION"
			}
		}
	}
	raw, err := json.MarshalIndent(frame, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	raw = append(raw, '\n')
	path := filepath.Join("testdata", "canonical-show", name+".json")
	if *updateCanonicalShowCorpus {
		if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(raw, want) {
		t.Fatalf("canonical public corpus changed: %s\n%s", path, raw)
	}
}
