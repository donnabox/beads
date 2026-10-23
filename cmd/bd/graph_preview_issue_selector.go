package main

import (
	"errors"
	"strings"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// A local Issue ID is opaque: even beads/PATH spelling first names that whole
// native ID. Only a canonical Scope URL selects an exact graph path.
func graphPreviewIssueSelector(selector string) (graphstore.IssueSelector, error) {
	if path, kind, ok := graph.SplitCanonicalURL(graphPreviewConfig.GraphScopeURL, selector); ok {
		if kind != graph.KindBead || graph.ValidateBeadPath(path) != nil {
			return graphstore.IssueSelector{}, errors.New("expected a local Issue ID or canonical Bead URL")
		}
		return graphstore.IssueSelector{Path: path}, nil
	}
	if strings.Contains(selector, "://") {
		return graphstore.IssueSelector{}, errors.New("expected a local Issue ID or canonical Bead URL")
	}
	path := "beads/" + selector
	if graph.ValidateBeadPath(path) != nil {
		return graphstore.IssueSelector{}, errors.New("expected a local Issue ID or canonical Bead URL")
	}
	return graphstore.IssueSelector{Path: path, NativeID: selector}, nil
}
