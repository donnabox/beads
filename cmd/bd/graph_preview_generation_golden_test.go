//go:build cgo

package main

import (
	"slices"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// graphPreviewGoldenTypes pins the Type names a fresh init installs, in order.
var graphPreviewGoldenTypes = map[string][]string{
	"link-preview-v9\n":  {"memory", "issue", "dependency", "related", "example-follows", "example-cites"},
	"link-preview-v10\n": {"memory", "evaluation", "constraint", "policy", "decision", "procedure", "requirement", "scenario", "specification", "references", "superseded-by", "requires", "defines", "subject-to", "conformance", "addresses", "satisfies", "violates", "issue", "dependency"},
}

// An older bd checks only the marker (and a schema version that names the
// storage layout, which adding a Type does not change), so the marker is the one
// signal that a workspace holds a Type set the reader does not know. Changing
// what a fresh init installs without changing the marker lets an older bd open
// the workspace and fail later, at its first list or inventory read.
func TestGraphPreviewGenerationPinsFreshInstallTypes(t *testing.T) {
	golden, ok := graphPreviewGoldenTypes[graphPreviewGeneration]
	if !ok {
		t.Fatalf("graphPreviewGeneration is %q but graphPreviewGoldenTypes has no entry for it. "+
			"A new generation needs a marker, supported-set entry, and a golden Type list", graphPreviewGeneration)
	}
	if got := graphstore.FreshTypeNames(); !slices.Equal(got, golden) {
		t.Fatalf("a fresh init now installs %v, but generation %q pins %v. "+
			"Adding, removing or reordering an installed Type is a workspace format change: "+
			"bump graphPreviewGeneration, update the fresh-only accepted set, and add a golden entry", got, graphPreviewGeneration, golden)
	}
	if !graphPreviewGenerationSupported([]byte(graphPreviewGeneration)) {
		t.Fatalf("the generation a fresh init writes (%q) is not in the accepted set, so every new "+
			"workspace would be refused at its first command: add it to graphPreviewSupportedGenerations",
			graphPreviewGeneration)
	}
}
