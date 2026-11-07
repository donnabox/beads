package main

import (
	"context"
	"fmt"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// Familiar Issue commands keep the ordinary top-level JSON contract. Generic
// graph commands continue to use graphPrint's preview envelope.
func withGraphFamiliarStore(fn func(context.Context, *graphstore.Store) (any, string, error)) error {
	return withGraphStoreOutput(fn, graphPrintFamiliar)
}

// Claim already resolves its audit identity this way. Coordination commands
// must use the same default, or an unconfigured --actor cannot heartbeat the
// claim that this CLI just created.
func graphPreviewCoordinationActor() string { return getActorWithGit() }

func graphPrintFamiliar(result any, human string) error {
	if jsonOutput {
		return outputJSON(result)
	}
	if !quietFlag {
		_, err := fmt.Println(human) //nolint:forbidigo // Familiar commands use the existing graph output seam, which has no Cobra writer.
		return err
	}
	return nil
}
