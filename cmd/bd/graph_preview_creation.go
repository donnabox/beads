package main

import (
	"github.com/steveyegge/beads/internal/memoryapi"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// This is a deterministic display summary, not an inferred semantic title.
// The complete body remains unchanged; an empty summary gets "Untitled memory".
func graphPreviewMemoryTitle(body string) string {
	return memoryapi.TitleFromBody(body)
}

// Human summaries use exact public identity rather than a private locator.
func graphPreviewMemoryDisplayID(memory graphstore.Record) string {
	if memory.PublicID != "" {
		return memory.PublicID
	}
	return memory.ID
}
