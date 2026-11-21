package issueops

import "testing"

// These tests pin the stored derivation rule for
// issue_versions.attribution_status, the NOT NULL column migration 0068 step 6
// adds. Its internal tokens are "claimed" and "unknown". Public BDP carried
// attribution instead uses basis "writer-supplied" or "unknown" when a
// principal exists; graphread.projectAttribution performs that projection.
// "imported" is provenance, not an attribution basis.
//
// RecordVersionInTx receives only a plain actor string from every call site
// (none passes any additional attribution context), so the only signal
// available to derive attribution_status from is whether actor is empty: a
// non-empty actor is "claimed", an empty one is "unknown".
func TestAttributionStatusForActor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		actor string
		want  string
	}{
		{name: "non-empty actor is claimed", actor: "alice", want: attributionStatusClaimed},
		{name: "empty actor is unknown", actor: "", want: attributionStatusUnknown},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := attributionStatusForActor(tc.actor); got != tc.want {
				t.Errorf("attributionStatusForActor(%q) = %q, want %q", tc.actor, got, tc.want)
			}
		})
	}
}

// TestAttributionStatusValuesAreTheStoredEncoding pins the native column's
// closed vocabulary and checks that actor-based derivation stays within it.
// It does not assert public BDP vocabulary; graphread tests that projection.
func TestAttributionStatusValuesAreTheStoredEncoding(t *testing.T) {
	t.Parallel()

	values := map[string]string{
		"claimed": attributionStatusClaimed,
		"unknown": attributionStatusUnknown,
	}
	for want, got := range values {
		if got != want {
			t.Errorf("attribution status constant = %q, want %q", got, want)
		}
	}

	legal := map[string]bool{"claimed": true, "unknown": true}
	for _, actor := range []string{"", "alice", "agent:builder"} {
		if got := attributionStatusForActor(actor); !legal[got] {
			t.Errorf("attributionStatusForActor(%q) = %q, want one of claimed|unknown", actor, got)
		}
	}
}
