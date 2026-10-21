package main

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// Changing the selected catalog changes the CLI workspace marker, not the
// storage schema. A fresh CLI must use the same generation its provider admits.
func TestGraphPreviewSelectedTypeGenerationMatchesBackend(t *testing.T) {
	want := fmt.Sprintf("%s%d\n", graphPreviewGenerationPrefix, graphstore.TypeGeneration)
	if graphPreviewGeneration != want {
		t.Fatalf("fresh CLI marker %q differs from provider generation %q", graphPreviewGeneration, want)
	}
	if len(graphPreviewSupportedGenerations) != 1 || graphPreviewSupportedGenerations[0] != want {
		t.Fatalf("selected catalog must admit only its fresh generation: %q", graphPreviewSupportedGenerations)
	}
	if !graphPreviewGenerationSupported([]byte(want)) || graphPreviewGenerationNewer([]byte(want)) {
		t.Fatalf("fresh provider generation is not admitted as current: %q", want)
	}
	prior := fmt.Sprintf("%s%d\n", graphPreviewGenerationPrefix, graphstore.TypeGeneration-1)
	if graphPreviewGenerationSupported([]byte(prior)) || graphPreviewGenerationNewer([]byte(prior)) {
		t.Fatalf("prior catalog must refuse without being mislabeled as newer: %q", prior)
	}
}

// Native relationship commands remain available in the current selected Type
// generation, while old/incomplete/newer installations do not bypass their gate.
func TestGraphPreviewRelationshipGenerationUsesSelectedCatalog(t *testing.T) {
	previous, previousOutput := graphPreviewTypeGeneration, graphFailureOutput
	var diagnostics bytes.Buffer
	graphFailureOutput = &diagnostics
	t.Cleanup(func() {
		graphPreviewTypeGeneration, graphFailureOutput = previous, previousOutput
	})
	for _, generation := range []int{0, graphstore.TypeGeneration - 1, graphstore.TypeGeneration, graphstore.TypeGeneration + 1} {
		graphPreviewTypeGeneration = generation
		diagnostics.Reset()
		err := requireGraphPreviewRelationshipGeneration()
		if generation == graphstore.TypeGeneration {
			if err != nil || diagnostics.Len() != 0 {
				t.Fatalf("current relationship generation refused: %v %s", err, diagnostics.String())
			}
			continue
		}
		var exit *exitError
		if !errors.As(err, &exit) || exit.Code != 5 || diagnostics.Len() == 0 {
			t.Fatalf("generation %d did not retain capability refusal: %v %s", generation, err, diagnostics.String())
		}
	}
}
