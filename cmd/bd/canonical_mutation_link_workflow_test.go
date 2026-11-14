//go:build cgo

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// Real CLI emission and whole-batch compilation, deliberately without a store.
// This is not a database publication, public Write or profile-routing proof.
func TestCanonicalLinkUpdateMemberEmissionWorkflow(t *testing.T) {
	isolateBeadsDirForTest(t)
	bd := buildBDUnderTest(t)
	work, home := t.TempDir(), t.TempDir()
	emit := func(args ...string) []byte {
		t.Helper()
		return []byte(graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--scope", canonicalLinkLoweringScope, "--dry-run-mode", "batch")...))
	}
	raw := emit("link", "update", "edge", "--from=next", "--to=target@=target-r8", "--if-revision==edge-r1")
	var edited map[string]json.RawMessage
	if err := json.Unmarshal(raw, &edited); err != nil { t.Fatal(err) }
	if string(edited["source"]) != `"https://example.invalid/link-edits/next"` || string(edited["target"]) != `{"revision":"target-r8","uri":"https://example.invalid/link-edits/target"}` || string(edited["expectedRevision"]) != `"edge-r1"` { t.Fatalf("CLI endpoint/guard intent lost: %s", raw) }

	var members []byte
	for _, args := range [][]string{
		{"create", "--id=first", "--batch-bind=first", "--type=memory", "--properties={\"title\":\"First\",\"body\":\"A\"}"},
		{"create", "--id=second", "--batch-bind=second", "--type=memory", "--properties={\"title\":\"Second\",\"body\":\"B\"}"},
		{"link", "create", "batch-ref::first", "batch-ref::second", "--id=edge", "--batch-bind=edge", "--type=references"},
		{"link", "update", "batch-ref::edge", "--from=batch-ref::second", "--to=batch-ref::first"},
	} {
		members = append(members, emit(args...)...)
	}
	file := filepath.Join(work, "edits.jsonl")
	if err := os.WriteFile(file, members, 0600); err != nil { t.Fatal(err) }
	compiled := graphPolicyCLI(t, bd, work, home, nil, "", "batch", "--file", file, "--scope", canonicalLinkLoweringScope, "--dry-run-mode=batch")
	var request struct { Operations []map[string]json.RawMessage `json:"operations"` }
	if err := json.Unmarshal([]byte(compiled), &request); err != nil { t.Fatal(err) }
	if len(request.Operations) != 4 { t.Fatalf("batch lost a member: %s", compiled) }
	last := request.Operations[3]
	if string(last["link"]) != `{"binding":"edge"}` || string(last["source"]) != `{"binding":"second"}` || string(last["target"]) != `{"binding":"first"}` { t.Fatalf("same-batch References changed: %s", compiled) }
	graphPolicyCLI(t, bd, work, home, nil, "stderr:invalid_selector:", "link", "update", "edge", "--from=next", "--from=next", "--scope", canonicalLinkLoweringScope, "--dry-run-mode=batch")
	if _, err := os.Stat(filepath.Join(work, ".beads")); !os.IsNotExist(err) { t.Fatal("offline endpoint lowering initialized or touched a workspace") }
}
