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
	if strings.HasPrefix(selector, graphPreviewConfig.GraphScopeURL) {
		id, err := graph.ScopeLocalPublicResourceID(graphPreviewConfig.GraphScopeURL, selector)
		if err != nil {
			return graphstore.IssueSelector{}, err
		}
		return graphstore.IssueSelector{Path: "beads/" + id, NativeID: id, Exact: true}, nil
	}
	if err := graph.ValidateCanonicalResourceID(selector); err != nil {
		return graphstore.IssueSelector{}, errors.New("expected a local Issue ID or canonical Scope URL")
	}
	return graphstore.IssueSelector{Path: "beads/" + selector, NativeID: selector}, nil
}

// An allocated literal ID wins before write-reference syntax is interpreted.
// An unallocated ID@revision cannot silently become a historical write target.
func graphPreviewWritableResourceArgument(selector string) (string, error) {
	resolved, err := graphPreviewResolveMemoryArgument(selector)
	if err != nil {
		return "", err
	}
	identity, err := graphPreviewWritableIdentity(resolved)
	if err != nil {
		return "", graphFailure("invalid_selector", err.Error(), 2)
	}
	return identity, nil
}

// Resolve an exact public ID before passing its immutable allocation path to
// an existing checked writer. A compatibility shorthand can change meaning
// between this read and the writer's transaction, so it must use a canonical
// Scope URL (or a writer that resolves shorthand in its own transaction).
func graphPreviewResolvedWritePath(selector string) (string, error) {
	selector, err := graphPreviewWritableResourceArgument(selector)
	if err != nil {
		return "", err
	}
	path, localID, err := graphPreviewResourceReadSelector(selector)
	if err != nil {
		return "", graphFailure("invalid_selector", err.Error(), 2)
	}
	if localID == "" {
		return graphPreviewCanonicalAllocationPath(path)
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
	exact := selected == localID // Older Memory and Link allocations publish paths.
	if memory, ok := record.(graphstore.Record); ok && memory.PublicID != "" {
		exact = memory.PublicID == localID
	}
	if issue, ok := record.(graphstore.IssueRecord); ok {
		exact = issue.Properties.ID == localID
	}
	if !exact {
		return "", graphFailure("invalid_selector", fmt.Sprintf("write selector %q is a compatibility shorthand; use the full public ID or canonical Scope URL", localID), 2)
	}
	return selected, nil
}

func graphPreviewResolvedHistoricalPath(selector string) (string, error) {
	selector, err := graphPreviewWritableResourceArgument(selector)
	if err != nil {
		return "", err
	}
	path, localID, err := graphPreviewResourceReadSelector(selector)
	if err != nil {
		return "", graphFailure("invalid_selector", err.Error(), 2)
	}
	if localID == "" {
		return graphPreviewCanonicalAllocationPath(path)
	}
	var selected string
	err = withGraphStoreOutput(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		selected, err = store.ResolveExactAllocatedResourcePath(ctx, localID)
		return nil, "", err
	}, func(any, string) error { return nil })
	return selected, err
}

// Current mixed reads accept opaque local public IDs or exact same-Scope URLs.
// The local spelling never chooses Bead versus Link from a prefix.
func graphPreviewResourceReadSelector(selector string) (path, localID string, err error) {
	return graphPreviewResourceReadSelectorIn(graphPreviewConfig.GraphScopeURL, selector)
}

func graphPreviewResourceReadSelectorIn(scope, selector string) (path, localID string, err error) {
	if strings.HasPrefix(selector, scope) {
		if _, err := graph.ScopeLocalPublicResourceID(scope, selector); err != nil {
			return "", "", err
		}
		return selector, "", nil
	}
	if err := graph.ValidateCanonicalResourceID(selector); err != nil {
		return "", "", errors.New("expected a local Resource ID or canonical Scope URL")
	}
	return "beads/" + selector, selector, nil
}

// Resolve an exact allocated public ID before interpreting optional revision
// syntax. Full Scope URLs and literal IDs are disjoint under the common grammar;
// no kind or alternative allocation is inferred from a rooted spelling.
func graphPreviewResolveMemoryArgument(selector string) (string, error) {
	if strings.HasPrefix(selector, graphPreviewConfig.GraphScopeURL) {
		return selector, nil
	}
	// URL reference syntax is disjoint from admitted local IDs. Leave foreign
	// URLs to the command's reference/Scope validation, never a key codec.
	if strings.Contains(selector, "://") {
		return selector, nil
	}
	// Non-ID reference syntax cannot name an allocated literal ID. Let the
	// reference parser report selector errors without a speculative store read.
	if graph.ValidateCanonicalResourceID(selector) != nil {
		return selector, nil
	}
	resolved := selector
	err := withGraphStoreOutput(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		identity, err := store.ResolveAllocatedResourceIdentity(ctx, selector)
		if errors.Is(err, graphstore.ErrNotFound) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}
		resolved, err = graph.CanonicalPublicResourceURL(store.ScopeURL(), identity.PublicID)
		return nil, "", err
	}, func(any, string) error { return nil })
	return resolved, err
}

// --id/--key forms select literal identity data. They never interpret the data
// as URL or revision syntax, even when another Resource owns that locator.
func graphPreviewResolveLiteralArgument(selector string) (string, error) {
	var resolved string
	err := withGraphStoreOutput(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		identity, err := store.ResolveAllocatedResourceIdentity(ctx, selector)
		if err != nil {
			return nil, "", err
		}
		resolved, err = graph.CanonicalPublicResourceURL(store.ScopeURL(), identity.PublicID)
		return nil, "", err
	}, func(any, string) error { return nil })
	return resolved, err
}

func graphPreviewCanonicalAllocationPath(uri string) (string, error) {
	var path string
	err := withGraphStoreOutput(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		identity, err := store.ResolveCanonicalPublicResourceIdentity(ctx, uri)
		if err == nil {
			path = identity.PrivatePath
		}
		return nil, "", err
	}, func(any, string) error { return nil })
	return path, err
}

func graphPreviewCanonicalResourceAddress(scope, address string) bool {
	_, err := graph.ScopeLocalPublicResourceID(scope, address)
	return err == nil
}

// The projection boundary validates URI spelling only. Selected record cases
// and the checked allocation catalog establish Bead versus Link kind.
func graphPreviewPublicResourceID(scope, uri string) (string, bool) {
	id, err := graph.ScopeLocalPublicResourceID(scope, uri)
	return id, err == nil
}
