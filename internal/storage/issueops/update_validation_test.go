package issueops

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/config"
	"github.com/steveyegge/beads/internal/storage"
	publicops "github.com/steveyegge/beads/issueops"
)

func TestValidateUpdateRequestRejectsInvalidCanonicalFields(t *testing.T) {
	negative := -1
	for _, tc := range []struct {
		name  string
		patch publicops.IssuePatch
	}{
		{"priority above range", publicops.IssuePatch{Priority: publicops.Field[int]{Set: true, Value: 9}}},
		{"empty title", publicops.IssuePatch{Title: publicops.Field[string]{Set: true}}},
		{"title over byte limit", publicops.IssuePatch{Title: publicops.Field[string]{Set: true, Value: strings.Repeat("t", 501)}}},
		{"negative estimate", publicops.IssuePatch{EstimatedMinutes: publicops.Field[*int]{Set: true, Value: &negative}}},
		{"unknown persistence mode", publicops.IssuePatch{Persistence: publicops.Field[publicops.PersistenceMode]{Set: true, Value: "archived"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateUpdateRequest(publicops.UpdateRequest{Actor: "actor", IssueID: "bd-validate", Patch: tc.patch})
			if !errors.Is(err, storage.ErrValidation) {
				t.Fatalf("ValidateUpdateRequest() error = %v, want ErrValidation", err)
			}
		})
	}
}

// TestValidateUpdateRequestRejectsInvalidNotesRequestShapes covers the two
// notes rules ValidateUpdateRequest owns that are properties of the REQUEST
// rather than of the patch alone, so they cannot be expressed in the
// patch-only table above. Both are contract-level: the batch apply leg reaches
// them here and nowhere else, and the single-issue HTTP handler's own
// pre-check returns before storage is reached, so without these cases the
// branches backing the published ApplyUpdateItem promise are unexercised.
func TestValidateUpdateRequestRejectsInvalidNotesRequestShapes(t *testing.T) {
	for _, tc := range []struct {
		name    string
		request publicops.UpdateRequest
	}{
		{
			name: "forced notes overwrite without a notes value",
			request: publicops.UpdateRequest{
				Actor: "actor", IssueID: "bd-validate",
				ForceNotesOverwrite: true,
			},
		},
		{
			name: "notes replacement combined with append",
			request: publicops.UpdateRequest{
				Actor: "actor", IssueID: "bd-validate",
				Patch: publicops.IssuePatch{
					Notes:       publicops.Field[string]{Set: true, Value: "replacement"},
					AppendNotes: publicops.Field[string]{Set: true, Value: "addition"},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateUpdateRequest(tc.request)
			if !errors.Is(err, storage.ErrValidation) {
				t.Fatalf("ValidateUpdateRequest() error = %v, want ErrValidation", err)
			}
		})
	}
}

func TestApplyMetadataPatchRejectsUnsafeKeysAndNullMerge(t *testing.T) {
	for _, tc := range []struct {
		name  string
		patch publicops.MetadataPatch
	}{
		{"set key with unsafe characters", publicops.MetadataPatch{Set: map[string]json.RawMessage{"bad-key": json.RawMessage(`"value"`)}}},
		{"unset key with unsafe characters", publicops.MetadataPatch{Unset: []string{"bad-key"}}},
		{"null merge payload", publicops.MetadataPatch{Merge: publicops.Field[json.RawMessage]{Set: true, Value: json.RawMessage(`null`)}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			next, changed, err := ApplyMetadataPatch(json.RawMessage(`{"keep":"yes"}`), tc.patch)
			if !errors.Is(err, storage.ErrValidation) {
				t.Fatalf("ApplyMetadataPatch() = %s, %t, %v; want ErrValidation", next, changed, err)
			}
		})
	}
}

func TestOrderedMetadataOperations(t *testing.T) {
	for _, tc := range []struct {
		name, before, operations, want string
		changed                       bool
	}{
		{"empty", `{"keep":true}`, `[]`, `{"keep":true}`, false},
		{"empty-native-metadata", ``, `[]`, `{}`, false},
		{"legacy-null", `null`, `[]`, `{}`, false},
		{"nested-and-array", `{"a/b":{"~n":1},"list":[1,3]}`, `[{"op":"replace","path":"/a~1b/~0n","value":2},{"op":"add","path":"/list/1","value":2}]`, `{"a/b":{"~n":2},"list":[1,2,3]}`, true},
		{"null-is-not-remove", `{"x":1}`, `[{"op":"replace","path":"/x","value":null}]`, `{"x":null}`, true},
		{"tolerant-unset-absent", `{"keep":true}`, `[{"op":"add","path":"/missing","value":null},{"op":"remove","path":"/missing"}]`, `{"keep":true}`, false},
		{"tolerant-unset-present", `{"key":1,"keep":true}`, `[{"op":"add","path":"/key","value":null},{"op":"remove","path":"/key"}]`, `{"keep":true}`, true},
		{"final-object-only", `{"keep":true}`, `[{"op":"replace","path":"","value":[]},{"op":"replace","path":"","value":{"keep":true}}]`, `{"keep":true}`, false},
		{"root-replacement", `{"remove":true}`, `[{"op":"replace","path":"","value":{"new":true}}]`, `{"new":true}`, true},
		{"exact-number-change", `{"n":9007199254740994}`, `[{"op":"replace","path":"/n","value":9007199254740996}]`, `{"n":9007199254740996}`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			current := json.RawMessage(tc.before)
			patch := publicops.MetadataPatch{Operations: publicops.Field[json.RawMessage]{Set: true, Value: json.RawMessage(tc.operations)}}
			next, changed, err := ApplyMetadataPatch(current, patch)
			if err != nil || string(next) != tc.want || changed != tc.changed {
				t.Fatalf("ordered metadata = %s, changed=%t, %v; want %s, %t", next, changed, err, tc.want, tc.changed)
			}
			if string(current) != tc.before || string(patch.Operations.Value) != tc.operations {
				t.Fatal("ordered metadata changed caller-owned input")
			}
		})
	}
}

func TestOrderedMetadataRefusals(t *testing.T) {
	for _, raw := range []string{
		``, `null`, `{}`, `[{"op":"move","path":"/x","from":"/keep"}]`,
		`[{"op":"add","path":"/x","value":{"duplicate":1,"duplicate":2}}]`,
		`[{"op":"replace","path":"/absent","value":1}]`,
		`[{"op":"remove","path":"/absent"}]`,
		`[{"op":"replace","path":"","value":null}]`,
		`[{"op":"replace","path":"","value":[]}]`,
		`[{"op":"remove","path":""}]`,
		`[{"op":"add","path":"/changed","value":true},{"op":"remove","path":"/absent"}]`,
	} {
		current := json.RawMessage(`{"keep":true}`)
		next, changed, err := ApplyMetadataPatch(current, publicops.MetadataPatch{Operations: publicops.Field[json.RawMessage]{Set: true, Value: json.RawMessage(raw)}})
		if !errors.Is(err, storage.ErrValidation) || next != nil || changed || string(current) != `{"keep":true}` {
			t.Errorf("invalid metadata operations %s = %s, %t, %v", raw, next, changed, err)
		}
	}
	for _, other := range []publicops.MetadataPatch{
		{Replace: publicops.Field[json.RawMessage]{Set: true, Value: json.RawMessage(`{}`)}},
		{Merge: publicops.Field[json.RawMessage]{Set: true, Value: json.RawMessage(`{}`)}},
		{Set: map[string]json.RawMessage{"x": json.RawMessage(`1`)}},
		{Unset: []string{"x"}},
	} {
		other.Operations = publicops.Field[json.RawMessage]{Set: true, Value: json.RawMessage(`[]`)}
		if err := ValidateMetadataPatch(other); !errors.Is(err, storage.ErrValidation) {
			t.Errorf("mixed metadata operations admitted: %+v, %v", other, err)
		}
		if _, _, err := ApplyMetadataPatch(json.RawMessage(`{}`), other); !errors.Is(err, storage.ErrValidation) {
			t.Errorf("ApplyMetadataPatch admitted mixed modes: %+v, %v", other, err)
		}
	}
}

func TestOrderedMetadataValidatesFinalStateEvenForEmptyUpdate(t *testing.T) {
	t.Setenv("BEADS_TEST_IGNORE_REPO_CONFIG", "1")
	if err := config.Initialize(); err != nil {
		t.Fatal(err)
	}
	defer config.ResetForTesting()
	config.Set("validation.metadata.mode", "error")
	config.Set("validation.metadata.fields", map[string]interface{}{
		"required_field": map[string]interface{}{"type": "string", "required": true},
	})
	patch := publicops.MetadataPatch{Operations: publicops.Field[json.RawMessage]{Set: true, Value: json.RawMessage(`[]`)}}
	if _, _, err := ApplyMetadataPatch(json.RawMessage(`{}`), patch); err == nil {
		t.Fatal("empty update bypassed configured final-state validation")
	}
	patch.Operations.Value = json.RawMessage(`[{"op":"add","path":"/required_field","value":1},{"op":"replace","path":"/required_field","value":"valid"}]`)
	if next, changed, err := ApplyMetadataPatch(json.RawMessage(`{}`), patch); err != nil || !changed || string(next) != `{"required_field":"valid"}` {
		t.Fatalf("temporary schema violation was not repaired atomically: %s, %t, %v", next, changed, err)
	}
}
