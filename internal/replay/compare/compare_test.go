// The tests cover four things:
//
//   - canonicalization and hashing: object keys are sorted recursively, numbers
//     take ECMA-262 form, and the digest is lowercase hex SHA-256. The number
//     case is the behavioral proof that Canonicalize delegates to the JCS
//     library rather than re-implementing RFC 8785 by hand.
//   - matching: identical payloads, and payloads that differ only in key order,
//     compare equal on their canonical form.
//   - the mismatch record: both payloads are carried verbatim, and every
//     classification rule, including the priority order between them, is
//     exercised by TestCompare_ClassifiesMismatchCategory.
//   - the package has no database, clone or network dependency; its only
//     non-standard-library import is the JCS library.
package compare

import (
	"encoding/hex"
	"strings"
	"testing"
)

func TestCanonicalize_SortsObjectKeysRecursively(t *testing.T) {
	in := []byte(`{"b":1,"a":{"d":2,"c":3}}`)
	got, err := Canonicalize(in)
	if err != nil {
		t.Fatalf("Canonicalize(%s): unexpected error: %v", in, err)
	}
	want := `{"a":{"c":3,"d":2},"b":1}`
	if string(got) != want {
		t.Fatalf("Canonicalize(%s) = %s, want %s", in, got, want)
	}
}

func TestCanonicalize_NumbersUseECMA262Formatting(t *testing.T) {
	// RFC 8785 mandates ECMA-262 number formatting (no trailing ".0" for a
	// whole-valued float). A hand-rolled canonicalizer built on
	// json.Marshal would not give this for free -- this is the behavioral
	// proof that Canonicalize really delegates to jcs.Transform
	// rather than reimplementing RFC 8785 by hand.
	in := []byte(`{"n":1.0}`)
	got, err := Canonicalize(in)
	if err != nil {
		t.Fatalf("Canonicalize(%s): unexpected error: %v", in, err)
	}
	want := `{"n":1}`
	if string(got) != want {
		t.Fatalf("Canonicalize(%s) = %s, want %s", in, got, want)
	}
}

func TestCanonicalize_InvalidJSON_ReturnsError(t *testing.T) {
	if _, err := Canonicalize([]byte(`{not valid json`)); err == nil {
		t.Fatal("Canonicalize: want error for invalid JSON, got nil")
	}
}

func TestHash_ReturnsLowercaseHexSHA256(t *testing.T) {
	got := Hash([]byte("hello"))

	// Verified ground truth: printf '%s' hello | sha256sum
	const want = "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	if got != want {
		t.Fatalf("Hash(%q) = %s, want %s", "hello", got, want)
	}
	if len(got) != 64 {
		t.Fatalf("Hash(%q): got length %d, want 64 (32-byte digest, hex-encoded)", "hello", len(got))
	}
	if got != strings.ToLower(got) {
		t.Fatalf("Hash(%q) = %s, want lowercase hex", "hello", got)
	}
	if _, err := hex.DecodeString(got); err != nil {
		t.Fatalf("Hash(%q) = %s: not valid hex: %v", "hello", got, err)
	}
	if got2 := Hash([]byte("hello")); got != got2 {
		t.Fatalf("Hash(%q) not deterministic: %s vs %s", "hello", got, got2)
	}
	if got3 := Hash([]byte("world")); got == got3 {
		t.Fatalf("Hash: distinct inputs %q and %q produced the same digest %s", "hello", "world", got)
	}
}

func TestCompare_IdenticalPayloads_Match(t *testing.T) {
	payload := []byte(`{"id":"x-1","title":"Example","content_hash":"aaa111"}`)

	result, err := Compare(payload, payload)
	if err != nil {
		t.Fatalf("Compare: unexpected error: %v", err)
	}
	if !result.Matched {
		t.Fatalf("Compare(identical payloads): Matched = false, want true")
	}
	if result.OracleHash == "" || result.OracleHash != result.CandidateHash {
		t.Fatalf("Compare(identical payloads): OracleHash=%q CandidateHash=%q, want equal and non-empty", result.OracleHash, result.CandidateHash)
	}
	if result.Mismatch != nil {
		t.Fatalf("Compare(identical payloads): Mismatch = %+v, want nil", result.Mismatch)
	}
}

func TestCompare_KeyOrderOnlyDifference_StillMatches(t *testing.T) {
	oracle := []byte(`{"id":"x-1","title":"Example","meta":{"a":1,"b":2},"dependencies":["x-2","x-3"]}`)
	candidate := []byte(`{"dependencies":["x-2","x-3"],"meta":{"b":2,"a":1},"title":"Example","id":"x-1"}`)

	result, err := Compare(oracle, candidate)
	if err != nil {
		t.Fatalf("Compare: unexpected error: %v", err)
	}
	if !result.Matched {
		t.Fatalf("Compare(key-order-only difference): Matched = false, want true (comparison is on canonical form)")
	}
	if result.Mismatch != nil {
		t.Fatalf("Compare(key-order-only difference): Mismatch = %+v, want nil", result.Mismatch)
	}
}

func TestCompare_MismatchRecord_CarriesFullPayloadsVerbatim(t *testing.T) {
	oracle := []byte(`{"id":"x-1","title":"Old title"}`)
	candidate := []byte(`{"id":"x-1","title":"New title"}`)

	result, err := Compare(oracle, candidate)
	if err != nil {
		t.Fatalf("Compare: unexpected error: %v", err)
	}
	if result.Matched {
		t.Fatalf("Compare(differing title): Matched = true, want false")
	}
	if result.Mismatch == nil {
		t.Fatal("Compare(differing title): Mismatch = nil, want non-nil")
	}
	if string(result.Mismatch.ExpectedJSON) != string(oracle) {
		t.Fatalf("Mismatch.ExpectedJSON = %s, want %s (verbatim oracle payload)", result.Mismatch.ExpectedJSON, oracle)
	}
	if string(result.Mismatch.ActualJSON) != string(candidate) {
		t.Fatalf("Mismatch.ActualJSON = %s, want %s (verbatim candidate payload)", result.Mismatch.ActualJSON, candidate)
	}
	if result.Mismatch.Category == "" {
		t.Fatal("Mismatch.Category = \"\", want a non-empty category")
	}
}

func TestCompare_ClassifiesMismatchCategory(t *testing.T) {
	cases := []struct {
		name         string
		oracle       string
		candidate    string
		wantCategory string
	}{
		{
			name:         "dependency edge differs",
			oracle:       `{"id":"x-1","title":"Example","dependencies":["x-2"]}`,
			candidate:    `{"id":"x-1","title":"Example","dependencies":["x-2","x-3"]}`,
			wantCategory: "dep-edge",
		},
		{
			name:         "depends_on synonym field differs",
			oracle:       `{"id":"x-1","depends_on":["x-2"]}`,
			candidate:    `{"id":"x-1","depends_on":[]}`,
			wantCategory: "dep-edge",
		},
		{
			name:         "content hash differs",
			oracle:       `{"id":"x-1","title":"Example","content_hash":"aaa111"}`,
			candidate:    `{"id":"x-1","title":"Example","content_hash":"bbb222"}`,
			wantCategory: "version-count",
		},
		{
			name:         "revision synonym field differs",
			oracle:       `{"issue_id":"x-1","revision":3}`,
			candidate:    `{"issue_id":"x-1","revision":4}`,
			wantCategory: "version-count",
		},
		{
			name:         "attribution field (created_by) differs",
			oracle:       `{"id":"x-1","title":"Example","created_by":"alice"}`,
			candidate:    `{"id":"x-1","title":"Example","created_by":"bob"}`,
			wantCategory: "attribution",
		},
		{
			name:         "attribution field (assignee) differs",
			oracle:       `{"id":"x-1","assignee":"alice"}`,
			candidate:    `{"id":"x-1","assignee":"bob"}`,
			wantCategory: "attribution",
		},
		{
			name:         "attribution field (change_actor, issue_versions shape) differs",
			oracle:       `{"issue_id":"x-1","revision":3,"change_actor":"alice"}`,
			candidate:    `{"issue_id":"x-1","revision":3,"change_actor":"bob"}`,
			wantCategory: "attribution",
		},
		{
			name:         "epoch differs",
			oracle:       `{"issue_id":"x-1","revision":3,"epoch":1}`,
			candidate:    `{"issue_id":"x-1","revision":3,"epoch":2}`,
			wantCategory: "epoch",
		},
		{
			name:         "only timestamp fields differ",
			oracle:       `{"id":"x-1","title":"Example","created_at":"2026-09-01T00:00:00Z","updated_at":"2026-09-01T00:00:00Z"}`,
			candidate:    `{"id":"x-1","title":"Example","created_at":"2026-09-01T00:00:00Z","updated_at":"2026-09-02T00:00:00Z"}`,
			wantCategory: "as-of-mismatch",
		},
		{
			name:         "generic field value differs",
			oracle:       `{"id":"x-1","title":"Old title"}`,
			candidate:    `{"id":"x-1","title":"New title"}`,
			wantCategory: "unclassified",
		},
		{
			name:         "dependency edge takes priority over an unrelated field also differing",
			oracle:       `{"id":"x-1","title":"Old title","dependencies":["x-2"]}`,
			candidate:    `{"id":"x-1","title":"New title","dependencies":["x-2","x-3"]}`,
			wantCategory: "dep-edge",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result, err := Compare([]byte(c.oracle), []byte(c.candidate))
			if err != nil {
				t.Fatalf("Compare: unexpected error: %v", err)
			}
			if result.Matched {
				t.Fatalf("Compare: Matched = true, want false (fixtures differ)")
			}
			if result.Mismatch == nil {
				t.Fatal("Compare: Mismatch = nil, want non-nil")
			}
			if result.Mismatch.Category != c.wantCategory {
				t.Fatalf("Compare: Mismatch.Category = %q, want %q", result.Mismatch.Category, c.wantCategory)
			}
		})
	}
}

func TestCompare_InvalidJSON_ReturnsError(t *testing.T) {
	if _, err := Compare([]byte(`{not valid`), []byte(`{}`)); err == nil {
		t.Fatal("Compare: want error when oracle payload is invalid JSON, got nil")
	}
	if _, err := Compare([]byte(`{}`), []byte(`{not valid`)); err == nil {
		t.Fatal("Compare: want error when candidate payload is invalid JSON, got nil")
	}
}
