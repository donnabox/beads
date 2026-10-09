package main

import (
	"fmt"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// This is a preview result projection, not a public History or wire contract.
// The store supplies the predecessor captured in the successful transaction.
func graphPreviewReplacementSummary(summary string, replaced *graphstore.ReplacedMemory) string {
	if replaced == nil {
		return summary
	}
	if replaced.Attribution.Actor == "" {
		return fmt.Sprintf("%s\nReplaced Memory %s version %s; no recorded attribution", summary, replaced.ID, replaced.Version)
	}
	return fmt.Sprintf("%s\nReplaced Memory %s version %s; recorded attribution: actor=%q basis=%q recordedAt=%q",
		summary, replaced.ID, replaced.Version, replaced.Attribution.Actor,
		graphPublicAttributionBasis(replaced.Attribution.Status), replaced.Attribution.RecordedAt)
}
