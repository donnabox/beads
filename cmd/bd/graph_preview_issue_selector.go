package main

import (
	"context"
	"errors"
	"fmt"
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

// Resolve an exact public ID before passing its immutable allocation path to
// an existing checked writer. A compatibility shorthand can change meaning
// between this read and the writer's transaction, so it must use a canonical
// Scope URL (or a writer that resolves shorthand in its own transaction).
func graphPreviewResolvedWritePath(selector string) (string, error) {
	path, localID, err := graphPreviewResourceReadSelector(selector)
	if err != nil {
		return "", graphFailure("invalid_selector", err.Error(), 2)
	}
	if localID == "" {
		return path, nil
	}
	var selected string
	var record any
	err = withGraphStoreOutput(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		selected, record, err = store.ReadResourceSelected(ctx, path, localID)
		return nil, "", err
	}, func(any, string) error { return nil })
	if err != nil {
		return "", err
	}
	exact := selected == localID // Memory and Link public IDs are their paths.
	if issue, ok := record.(graphstore.IssueRecord); ok {
		exact = issue.Properties.ID == localID
	}
	if !exact {
		return "", graphFailure("invalid_selector", fmt.Sprintf("write selector %q is a compatibility shorthand; use the full public ID or canonical Scope URL", localID), 2)
	}
	return selected, err
}

func graphPreviewResolvedHistoricalPath(selector string) (string, error) {
	path, localID, err := graphPreviewResourceReadSelector(selector)
	if err != nil {
		return "", graphFailure("invalid_selector", err.Error(), 2)
	}
	if localID == "" {
		return path, nil
	}
	var selected string
	err = withGraphStoreOutput(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		selected, err = store.ResolveAllocatedResourcePath(ctx, localID)
		return nil, "", err
	}, func(any, string) error { return nil })
	return selected, err
}

// Current mixed reads accept opaque local public IDs or exact same-Scope URLs.
// The local spelling never chooses Bead versus Link from a prefix.
func graphPreviewResourceReadSelector(selector string) (path, localID string, err error) {
	if path, kind, ok := graph.SplitCanonicalURL(graphPreviewConfig.GraphScopeURL, selector); ok {
		if (kind == graph.KindBead && graph.ValidateBeadPath(path) == nil) || (kind == graph.KindLink && graph.ValidateLinkPath(path) == nil) {
			return path, "", nil
		}
		return "", "", errors.New("expected a local Resource ID or canonical Scope URL")
	}
	if strings.Contains(selector, "://") || graph.ValidateBeadPath("beads/"+selector) != nil || graph.ValidateLinkPath("links/"+selector) != nil {
		return "", "", errors.New("expected a local Resource ID or canonical Scope URL")
	}
	return "beads/" + selector, selector, nil
}
