package issueops

import (
	"encoding/json"
	"testing"

	publicops "github.com/steveyegge/beads/issueops"
)

func TestOrderedMetadataPatchUsesActualPredecessor(t *testing.T) {
	patch := publicops.MetadataPatch{Operations: publicops.Field[json.RawMessage]{Set: true, Value: json.RawMessage(`[{"op":"replace","path":"/nested/x","value":2},{"op":"add","path":"/list/-","value":"new"}]`)}}
	next, changed, err := ApplyMetadataPatch(json.RawMessage(`{"nested":{"x":1,"keep":true},"list":["old"]}`), patch)
	if err != nil || !changed || string(next) != `{"list":["old","new"],"nested":{"keep":true,"x":2}}` {
		t.Fatalf("next=%s changed=%t err=%v", next, changed, err)
	}
	if _, _, err := ApplyMetadataPatch(json.RawMessage(`{}`), patch); err == nil {
		t.Fatal("strict replace was weakened to merge")
	}
}

func TestOrderedMetadataPatchEmptyAndInvalid(t *testing.T) {
	patch := publicops.MetadataPatch{Operations: publicops.Field[json.RawMessage]{Set: true, Value: json.RawMessage(`[]`)}}
	next, changed, err := ApplyMetadataPatch(nil, patch)
	if err != nil || changed || string(next) != `{}` {
		t.Fatalf("empty patch=%s %t %v", next, changed, err)
	}
	conflict := patch
	conflict.Merge = publicops.Field[json.RawMessage]{Set: true, Value: json.RawMessage(`{}`)}
	if _, _, err := ApplyMetadataPatch(nil, conflict); err == nil {
		t.Fatal("combined patch and merge")
	}
	for _, raw := range []string{`[{"op":"add","path":"/x","value":1,"value":2}]`, `[{"op":"replace","path":"","value":null}]`, `[{"op":"remove","path":"/absent"}]`} {
		patch.Operations.Value = json.RawMessage(raw)
		if _, _, err := ApplyMetadataPatch(nil, patch); err == nil {
			t.Fatalf("accepted %s", raw)
		}
	}
}

func TestOrderedMetadataPatchSnapshotIsIndependent(t *testing.T) {
	raw := json.RawMessage(`[]`)
	input := publicops.UpdateRequest{Patch: publicops.IssuePatch{Metadata: publicops.MetadataPatch{Operations: publicops.Field[json.RawMessage]{Set: true, Value: raw}}}}
	clone := CloneUpdateRequest(input)
	raw[0] = 'x'
	if string(clone.Patch.Metadata.Operations.Value) != "[]" {
		t.Fatal("snapshot retained caller operation bytes")
	}
}
