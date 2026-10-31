package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/storage/issueops"
)

func canonicalCRUDTestCommand(t *testing.T, flags ...string) *cobra.Command {
	t.Helper()
	cmd := memoryPropertiesPatchCommand(t)
	canonicalCRUDRegisterFlags(cmd)
	cmd.Flags().StringArray("set-metadata", nil, "")
	cmd.Flags().StringArray("unset-metadata", nil, "")
	if err := cmd.ParseFlags(flags); err != nil {
		t.Fatal(err)
	}
	return cmd
}

func TestCanonicalCRUDIndependentObjectModes(t *testing.T) {
	cmd := canonicalCRUDTestCommand(t, "--set-property=title=Reviewed", `--patch-metadata=[{"op":"add","path":"/reviewed","value":true}]`)
	if err := canonicalCRUDNormalizeUpdate(cmd); err != nil {
		t.Fatal(err)
	}
	patch, err := graphPreviewPropertyOperations(cmd)
	if err != nil || string(patch) != `[{"op":"add","path":"/title","value":"Reviewed"}]` {
		t.Fatalf("property patch=%s %v", patch, err)
	}
	metadata, err := graphPreviewMetadataPatch(cmd)
	if err != nil || !metadata.Operations.Set {
		t.Fatalf("metadata patch=%+v %v", metadata, err)
	}
	after, changed, err := issueops.ApplyMetadataPatch(json.RawMessage(`{"keep":1}`), metadata)
	if err != nil || !changed || string(after) != `{"keep":1,"reviewed":true}` {
		t.Fatalf("metadata=%s %t %v", after, changed, err)
	}
}

func TestCanonicalCRUDModeConflictsAndStdinOwnership(t *testing.T) {
	for _, flags := range [][]string{
		{"--set-property=x=1", "--properties={}"},
		{"--unset-property=x", "--replace-properties={}"},
		{"--patch-properties=[]", "--patch=[]"},
		{"--patch-properties=[]", "--title=conflict"},
		{"--patch-properties=@-", "--metadata=@-"},
		{"--metadata=@-", "--stdin"},
	} {
		cmd := canonicalCRUDTestCommand(t, flags...)
		reader := bytes.NewBufferString("unconsumed")
		cmd.SetIn(reader)
		if err := canonicalCRUDNormalizeUpdate(cmd); err == nil {
			t.Fatalf("accepted %v", flags)
		}
		if reader.String() != "unconsumed" {
			t.Fatalf("read input before refusing %v", flags)
		}
	}
	cmd := canonicalCRUDTestCommand(t, "--patch-metadata=[]", "--metadata={}")
	if _, err := graphPreviewMetadataPatch(cmd); err == nil {
		t.Fatal("accepted conflicting metadata modes")
	}
}

func TestCanonicalCRUDPatchInputRemainsStrictAndEmptyIsValid(t *testing.T) {
	cmd := canonicalCRUDTestCommand(t, "--patch-properties=@-")
	cmd.SetIn(strings.NewReader("[]"))
	if err := canonicalCRUDNormalizeUpdate(cmd); err != nil {
		t.Fatal(err)
	}
	raw, err := graphPreviewPropertyOperations(cmd)
	if err != nil || string(raw) != "[]" {
		t.Fatalf("empty patch=%s %v", raw, err)
	}
	cmd = canonicalCRUDTestCommand(t, `--patch-properties=[{"op":"add","path":"/x","value":1,"value":2}]`)
	if err := canonicalCRUDNormalizeUpdate(cmd); err == nil {
		t.Fatal("accepted duplicate JSON members")
	}
}

func TestCanonicalCRUDMetadataStdinAndExactScalar(t *testing.T) {
	cmd := canonicalCRUDTestCommand(t, "--metadata=@-")
	cmd.SetIn(strings.NewReader(`{"source":"stdin"}`))
	patch, err := graphPreviewMetadataPatch(cmd)
	if err != nil || !patch.Merge.Set || string(patch.Merge.Value) != `{"source":"stdin"}` {
		t.Fatalf("metadata stdin=%+v %v", patch, err)
	}
	cmd = canonicalCRUDTestCommand(t, "--set-metadata=number=1e999")
	if _, err := graphPreviewMetadataPatch(cmd); err == nil {
		t.Fatal("rounded or stringified inadmissible numeric value")
	}
}

func TestCanonicalCRUDLoweredFlagsPassProviderGate(t *testing.T) {
	for _, flags := range [][]string{
		{"--set-property=title=Reviewed"},
		{"--unset-property=absent"},
		{"--patch-properties=[]"},
	} {
		cmd := canonicalCRUDTestCommand(t, flags...)
		if err := canonicalCRUDNormalizeUpdate(cmd); err != nil {
			t.Fatal(err)
		}
		if !cmd.Flags().Changed("patch") {
			t.Fatal("lowering omitted native property operations")
		}
		if err := graphPreviewFlags(cmd, "patch"); err != nil {
			t.Fatalf("normalized %v refused by provider gate: %v", flags, err)
		}
	}
	cmd := canonicalCRUDTestCommand(t, "--set-property=title=Reviewed")
	cmd.Flags().String("unsupported", "", "")
	if err := cmd.Flags().Set("unsupported", "explicit"); err != nil {
		t.Fatal(err)
	}
	if err := canonicalCRUDNormalizeUpdate(cmd); err != nil {
		t.Fatal(err)
	}
	// Clearing an unrelated flag cannot broaden the command's admission.
	cmd.Flags().Lookup("unsupported").Changed = false
	if err := graphPreviewFlags(cmd, "patch"); err == nil {
		t.Fatal("consumed property alias weakened unrelated flag refusal")
	}
}
