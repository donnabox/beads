package main

import (
	"errors"
	"strings"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// Only bare CLI spelling may resolve through the native Issue ID catalog.
// An explicit beads/PATH or canonical URL always denotes that exact path.
func graphPreviewIssueSelector(selector string) (graphstore.IssueSelector, error) {
	path, err := graphPreviewResourcePath(graphPreviewConfig.GraphScopeURL, selector)
	if err != nil || graph.ValidateBeadPath(path) != nil {
		return graphstore.IssueSelector{}, errors.New("expected a local Issue ID or beads/PATH")
	}
	selected := graphstore.IssueSelector{Path: path}
	if selector == strings.TrimPrefix(path, "beads/") {
		selected.NativeID = selector
	}
	return selected, nil
}
