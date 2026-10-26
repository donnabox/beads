package main

import (
	"strconv"
	"testing"
)

// The workspace marker is the first thing admission reads, so its
// classification is exact: only the current fresh generation is accepted,
// only a well-formed higher generation is called newer, and every other byte
// sequence keeps the generic refusal rather than earning a more specific claim.
func TestGraphPreviewGenerationMarkerClassification(t *testing.T) {
	cases := []struct {
		name      string
		marker    string
		supported bool
		newer     bool
	}{
		{"legacy generation", "link-preview-v5\n", false, false},
		{"prior generation", "link-preview-v7\n", false, false},
		{"older supported-era generation", "link-preview-v9\n", false, false},
		{"previous generation", "link-preview-v10\n", false, false},
		{"prior concrete-label generation", "link-preview-v11\n", false, false},
		{"prior flat-root generation", "link-preview-v12\n", false, false},
		{"current generation", "link-preview-v13\n", true, false},
		{"next generation", "link-preview-v14\n", false, true},
		{"numeric not lexical order", "link-preview-v100\n", false, true},
		{"two digits above current", "link-preview-v99\n", false, true},
		{"largest uint64", "link-preview-v18446744073709551615\n", false, true},
		{"overflows uint64", "link-preview-v18446744073709551616\n", false, true},
		{"beyond any integer type", "link-preview-v133456789012345678901234567890\n", false, true},
		{"older generation", "link-preview-v4\n", false, false},
		{"first generation", "link-preview-v1\n", false, false},
		{"zero", "link-preview-v0\n", false, false},
		{"empty", "", false, false},
		{"newline only", "\n", false, false},
		{"no number", "link-preview-v\n", false, false},
		{"no number and no newline", "link-preview-v", false, false},
		{"current without newline", "link-preview-v13", false, false},
		{"newer without newline", "link-preview-v14", false, false},
		{"newer with a second newline", "link-preview-v14\n\n", false, false},
		{"newer with carriage return", "link-preview-v14\r\n", false, false},
		{"leading zero", "link-preview-v06\n", false, false},
		{"leading zeros on a newer number", "link-preview-v007\n", false, false},
		{"trailing text on current", "link-preview-v13x\n", false, false},
		{"trailing text on newer", "link-preview-v14x\n", false, false},
		{"negative number", "link-preview-v-7\n", false, false},
		{"explicit plus sign", "link-preview-v+7\n", false, false},
		{"leading space", " link-preview-v9\n", false, false},
		{"space before number", "link-preview-v 7\n", false, false},
		{"different case", "Link-Preview-V7\n", false, false},
		{"fullwidth digit", "link-preview-v７\n", false, false},
		{"arabic-indic digit", "link-preview-v٧\n", false, false},
		{"embedded NUL", "link-preview-v9\x00\n", false, false},
		{"different prefix", "graph-preview-v7\n", false, false},
		{"two lines", "link-preview-v9\nlink-preview-v10\n", false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			marker := []byte(tc.marker)
			if got := graphPreviewGenerationSupported(marker); got != tc.supported {
				t.Errorf("supported(%q) = %t, want %t", tc.marker, got, tc.supported)
			}
			if got := graphPreviewGenerationNewer(marker); got != tc.newer {
				t.Errorf("newer(%q) = %t, want %t", tc.marker, got, tc.newer)
			}
		})
	}

	// Exactly generation 13 is accepted and exactly those above 13 are
	// newer, however the numbers are written in the ordinary decimal form.
	for generation := range 300 {
		marker := []byte("link-preview-v" + strconv.Itoa(generation) + "\n")
		wantSupported := generation == 13
		if got := graphPreviewGenerationSupported(marker); got != wantSupported {
			t.Errorf("supported(generation %d) = %t, want %t", generation, got, wantSupported)
		}
		if got := graphPreviewGenerationNewer(marker); got != (generation > 13) {
			t.Errorf("newer(generation %d) = %t, want %t", generation, got, generation > 13)
		}
	}

	// What a fresh workspace is written with must always be admitted, and is by
	// definition never newer than the binary that wrote it.
	if !graphPreviewGenerationSupported([]byte(graphPreviewGeneration)) || graphPreviewGenerationNewer([]byte(graphPreviewGeneration)) {
		t.Errorf("the generation a fresh workspace is written with (%q) is not admitted as current", graphPreviewGeneration)
	}
}
