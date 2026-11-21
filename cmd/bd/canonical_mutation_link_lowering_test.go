package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/canonicalbatch"
	"github.com/steveyegge/beads/internal/canonicalbatchstate"
)

const canonicalLinkLoweringScope = "https://example.invalid/link-edits/"

// Exercise the existing complete member compiler, not a second test encoder.
// Only its registered command pointer is substituted; no command tree is copied.
func canonicalLinkUpdateCompilerFixture(t *testing.T, argv ...string) ([]byte, error) {
	t.Helper()
	stdioMutex.Lock()
	defer stdioMutex.Unlock()
	old := linkUpdateCmd
	defer func() { linkUpdateCmd = old }()
	cmd := &cobra.Command{Use: "update"}
	for _, name := range []string{"from", "to", "type", "if-revision", "dry-run-mode", "scope", "properties", "metadata", "replace-properties", "replace-metadata", "patch-properties", "patch-metadata", "revision-provenance"} {
		cmd.Flags().String(name, "", "")
	}
	for _, name := range []string{"set-property", "unset-property", "set-metadata", "unset-metadata"} {
		cmd.Flags().StringArray(name, nil, "")
	}
	cmd.SetIn(bytes.NewBufferString("must not be read for an invalid endpoint"))
	if err := cmd.ParseFlags(argv); err != nil { return nil, err }
	linkUpdateCmd = cmd
	return canonicalCompileCRUDMember(cmd, cmd.Flags().Args(), canonicalLinkLoweringScope)
}

func TestCanonicalLinkUpdateEndpointEmission(t *testing.T) {
	raw, err := canonicalLinkUpdateCompilerFixture(t, "edge", "--from=next", "--to=[https://example.invalid/link-edits/target]@=target-r8", "--type=related", "--if-revision==edge-r1", "--set-property=title=Retargeted")
	if err != nil { t.Fatal(err) }
	var got map[string]json.RawMessage
	if err := json.Unmarshal(raw, &got); err != nil { t.Fatal(err) }
	for name, want := range map[string]string{
		"operation": `"updateLink"`,
		"link": `"https://example.invalid/link-edits/edge"`,
		"source": `"https://example.invalid/link-edits/next"`,
		"target": `{"revision":"target-r8","uri":"https://example.invalid/link-edits/target"}`,
		"type": `"https://example.invalid/link-edits/related"`,
		"expectedRevision": `"edge-r1"`,
	} {
		if string(got[name]) != want { t.Fatalf("%s changed: %s, want %s", name, got[name], want) }
	}
	if !bytes.Contains(got["propertiesChange"], []byte(`"value":"Retargeted"`)) { t.Fatalf("combined property edit lost: %s", raw) }
}

func TestCanonicalLinkUpdateEndpointOmissionAndUnpin(t *testing.T) {
	raw, err := canonicalLinkUpdateCompilerFixture(t, "edge", "--to=target")
	if err != nil { t.Fatal(err) }
	var got map[string]json.RawMessage
	if err := json.Unmarshal(raw, &got); err != nil { t.Fatal(err) }
	if _, exists := got["source"]; exists { t.Fatalf("omitted source became an edit: %s", raw) }
	if string(got["target"]) != `"https://example.invalid/link-edits/target"` { t.Fatalf("bare target did not lower to a complete unpinned replacement: %s", raw) }
}

func TestCanonicalLinkUpdateEndpointBindings(t *testing.T) {
	raw, err := canonicalLinkUpdateCompilerFixture(t, "batch-ref::edge", "--from=batch-ref::source", "--to=batch-ref::target")
	if err != nil { t.Fatal(err) }
	var got map[string]json.RawMessage
	if err := json.Unmarshal(raw, &got); err != nil { t.Fatal(err) }
	for field, label := range map[string]string{"link":"edge", "source":"source", "target":"target"} {
		if string(got[field]) != `{"binding":"`+label+`"}` { t.Fatalf("%s binding was not structural: %s", field, got[field]) }
	}
}

func TestCanonicalLinkUpdateEndpointRefusals(t *testing.T) {
	for _, value := range []string{"", "next@", "[https://example.invalid/link-edits/next]", "batch-ref::", "next@=", "[bad]@=r"} {
		t.Run(value, func(t *testing.T) {
			if raw, err := canonicalLinkUpdateCompilerFixture(t, "edge", "--to="+value, "--properties=@must-not-read"); err == nil || raw != nil || !strings.Contains(err.Error(), "--to:") { t.Fatalf("invalid endpoint consumed other input or produced a member: %s %v", raw, err) }
		})
	}
	for _, value := range []string{"next@3", "next@current"} {
		if raw, err := canonicalLinkUpdateCompilerFixture(t, "edge", "--to="+value); !errors.Is(err, canonicalbatch.ErrUnsupported) || raw != nil { t.Fatalf("online revision was discarded or invented: %s %v", raw, err) }
	}
}

func TestCanonicalLinkUpdateEmissionStagesWholeReplacement(t *testing.T) {
	current := map[string]canonicalbatchstate.Resource{}
	for _, id := range []string{"source", "target", "next"} {
		uri := canonicalLinkLoweringScope+id
		current[uri] = canonicalbatchstate.Resource{ID: uri, Kind: graph.KindBead, Revision: id+"-r1", Type: canonicalbatch.Reference{URI: canonicalLinkLoweringScope+"memory"}, Properties: json.RawMessage(`{"title":"existing"}`), Metadata: json.RawMessage(`{}`)}
	}
	edge := canonicalLinkLoweringScope+"edge"
	current[edge] = canonicalbatchstate.Resource{ID: edge, Kind: graph.KindLink, Revision: "edge-r1", Type: canonicalbatch.Reference{URI: canonicalLinkLoweringScope+"references"}, Properties: json.RawMessage(`{}`), Metadata: json.RawMessage(`{}`), Source: canonicalbatch.Reference{URI: canonicalLinkLoweringScope+"source", Revision: "source-r1"}, Target: canonicalbatch.Reference{URI: canonicalLinkLoweringScope+"target", Revision: "target-r1"}}
	for _, tc := range []struct { name, from, id, revision string; changed bool }{
		{"replace identity and unpin", "next", "next", "", true},
		{"same identity and unpin", "source", "source", "", true},
		{"same complete Reference is no-op", "source@=source-r1", "source", "source-r1", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := canonicalLinkUpdateCompilerFixture(t, "edge", "--from="+tc.from, "--if-revision==edge-r1")
			if err != nil { t.Fatal(err) }
			member, _, err := canonicalbatch.CompileMember(raw, canonicalLinkLoweringScope)
			if err != nil { t.Fatal(err) }
			plan, err := canonicalbatchstate.Stage(canonicalbatchstate.Input{Scope: canonicalLinkLoweringScope, Current: current, Members: []canonicalbatch.Member{member}})
			if err != nil { t.Fatal(err) }
			got := plan.Final[edge]
			if got.Source != (canonicalbatch.Reference{URI: canonicalLinkLoweringScope+tc.id, Revision: tc.revision}) || got.Target != current[edge].Target || got.Revision != "edge-r1" || len(plan.Outcomes) != 1 || plan.Outcomes[0].Changed != tc.changed { t.Fatalf("whole replacement changed the wrong state or minted a public revision: %+v", got) }
			if current[edge].Source.URI != canonicalLinkLoweringScope+"source" || current[edge].Source.Revision != "source-r1" { t.Fatal("lowering/staging mutated the caller's baseline") }
		})
	}
}

func TestCanonicalLinkUpdateEndpointAdmissionBeforeHooks(t *testing.T) {
	for _, args := range [][]string{
		{"edge", "--to=", "--dry-run-mode=batch"},
		{"edge", "--from=next@current", "--dry-run-mode=batch"},
		{"edge", "--to=batch-ref::next"},
		{"edge", "--to=next", "--to=next", "--dry-run-mode=batch"},
	} {
		root, cmd := canonicalMutationLeafFixture(t, []string{"link", "update"})
		argv := append([]string{"link", "update"}, args...)
		cmd.Args = canonicalMutationLeafArgs(cmd, []string{"link", "update"}, argv)
		acquired := false
		root.PersistentPreRunE = func(*cobra.Command, []string) error { acquired = true; return nil }
		root.SetArgs(argv)
		_, err := root.ExecuteC()
		var refusal *canonicalMutationLeafRefusal
		if !errors.As(err, &refusal) || refusal.ExitCode != 2 || acquired { t.Fatalf("invalid Link update reached provider hooks: %v", err) }
	}
}

func TestCanonicalLinkUpdateEmittedBindingsStageEarlierCreates(t *testing.T) {
	raw, err := canonicalLinkUpdateCompilerFixture(t, "batch-ref::edge", "--from=batch-ref::second", "--to=batch-ref::first")
	if err != nil { t.Fatal(err) }
	prefix := []byte(`{"operation":"createBead","id":"first","name":"first","type":"memory","properties":{"title":"First"}}
{"operation":"createBead","id":"second","name":"second","type":"memory","properties":{"title":"Second"}}
{"operation":"createLink","id":"edge","name":"edge","type":"references","source":{"binding":"first"},"target":{"binding":"second"}}
`)
	batch, err := canonicalbatch.Compile(append(prefix, raw...), canonicalbatch.Options{Format: "jsonl", Scope: canonicalLinkLoweringScope})
	if err != nil { t.Fatal(err) }
	plan, err := canonicalbatchstate.Stage(canonicalbatchstate.Input{Scope: canonicalLinkLoweringScope, Members: batch.Members()})
	if err != nil { t.Fatal(err) }
	link := plan.Final[canonicalLinkLoweringScope+"edge"]
	if link.Source != (canonicalbatch.Reference{URI: canonicalLinkLoweringScope+"second"}) || link.Target != (canonicalbatch.Reference{URI: canonicalLinkLoweringScope+"first"}) || link.Revision != "" || len(plan.ConsumedIDs) != 3 || len(plan.Outcomes) != 3 { t.Fatalf("earlier creates/captures did not compose into one final graph: %+v", plan) }
}
