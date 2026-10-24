package main

import (
	"fmt"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// This is a preview result projection, not a public History or wire contract.
// The store supplies the predecessor captured in the successful transaction.
func graphPreviewReplacementResult(result any, summary string, replaced *graphstore.ReplacedMemory, err error) (any, string, error) {
	if err != nil {
		return result, "", err
	}
	human, err := graphPreviewReplacementSummary(graphPreviewMutationSummary(summary, result), replaced)
	return result, human, err
}

func graphPreviewReplacementSummary(summary string, replaced *graphstore.ReplacedMemory) (string, error) {
	if replaced == nil {
		return summary, nil
	}
	if replaced.Attribution.Actor == "" {
		if replaced.Attribution.Status != "" && replaced.Attribution.Status != "unknown" {
			return "", fmt.Errorf("%w: stored attribution has no actor", graphstore.ErrInvalidStore)
		}
		return fmt.Sprintf("%s\nReplaced Memory %s version %s; no recorded attribution", summary, replaced.ID, replaced.Version), nil
	}
	basis, err := graphPublicAttributionBasis(replaced.Attribution.Status)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s\nReplaced Memory %s version %s; recorded attribution: actor=%q basis=%q recordedAt=%q",
		summary, replaced.ID, replaced.Version, replaced.Attribution.Actor,
		basis, replaced.Attribution.RecordedAt), nil
}

// Current references come from the writer's result, never from its replaced
// predecessor. Brackets keep canonical URLs unambiguous CLI reference inputs.
func graphPreviewMutationSummary(summary string, result any) string {
	current := func(kind, id, revision string) {
		if id != "" && revision != "" {
			summary += fmt.Sprintf("\nCurrent %s revision: [%s]@=%s", kind, id, revision)
		}
	}
	source := func(value any) {
		switch record := value.(type) {
		case graphstore.Record:
			current("source", record.ID, record.Revision)
		case graphstore.IssueRecord:
			current("source", record.ID, record.Revision)
		}
	}
	switch value := result.(type) {
	case graphstore.MemoryMutationResult:
		current("Memory", value.Memory.ID, value.Memory.Revision)
	case graphstore.IssueMutationResult:
		current("Issue", value.Issue.ID, value.Issue.Revision)
	case graphstore.LinkMutationResult:
		current("Link", value.Link.ID, value.Link.Revision)
		source(value.Source)
	case graphstore.DependencyResult:
		current("Link", value.Link.ID, value.Link.Revision)
		source(value.Source)
	case graphstore.LinkDeleteResult:
		// Removed Links have no new current state; only their owner may do so.
		source(value.Source)
	}
	return summary
}
