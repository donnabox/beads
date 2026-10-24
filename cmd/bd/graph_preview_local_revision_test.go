package main

import "testing"

func TestGraphPreviewLocalOrdinalAndExactTokenSyntax(t *testing.T) {
	for _, tc := range []struct {
		input   string
		ordinal int64
		token   string
		literal bool
	}{
		{"m1@4", 4, "", false},
		{"m1@=123", 0, "123", true},
		{"m1@=current", 0, "current", true},
		{"m1@=unpinned", 0, "unpinned", true},
		{"m1@current", 0, "current", false},
	} {
		ref, err := graphPreviewParseReference(tc.input)
		if err != nil || ref.Identity != "m1" || !ref.Explicit || ref.Ordinal != tc.ordinal || ref.Revision != tc.token || ref.LiteralRevision != tc.literal {
			t.Fatalf("%q: %+v %v", tc.input, ref, err)
		}
		if tc.literal && ref.currentSelection() {
			t.Fatalf("exact token %q became current/floating", tc.token)
		}
	}
	for _, input := range []string{"m1@0", "m1@9223372036854775808", "m1@="} {
		if _, err := graphPreviewParseReference(input); err == nil {
			t.Fatalf("accepted %q", input)
		}
	}
}
